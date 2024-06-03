package service_test

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/plgd-dev/device/v2/schema/device"
	"github.com/plgd-dev/device/v2/schema/platform"
	"github.com/plgd-dev/go-coap/v3/message"
	caService "github.com/plgd-dev/hub/v2/certificate-authority/test"
	coapgwTest "github.com/plgd-dev/hub/v2/coap-gateway/test"
	"github.com/plgd-dev/hub/v2/grpc-gateway/pb"
	grpcgwService "github.com/plgd-dev/hub/v2/grpc-gateway/test"
	httpgwTest "github.com/plgd-dev/hub/v2/http-gateway/test"
	"github.com/plgd-dev/hub/v2/http-gateway/uri"
	idService "github.com/plgd-dev/hub/v2/identity-store/test"
	pkgGrpc "github.com/plgd-dev/hub/v2/pkg/net/grpc"
	pkgHttp "github.com/plgd-dev/hub/v2/pkg/net/http"
	"github.com/plgd-dev/hub/v2/resource-aggregate/commands"
	"github.com/plgd-dev/hub/v2/resource-aggregate/events"
	raService "github.com/plgd-dev/hub/v2/resource-aggregate/test"
	rdService "github.com/plgd-dev/hub/v2/resource-directory/test"
	"github.com/plgd-dev/hub/v2/test"
	"github.com/plgd-dev/hub/v2/test/config"
	httpTest "github.com/plgd-dev/hub/v2/test/http"
	"github.com/plgd-dev/hub/v2/test/oauth-server/service"
	oauthTest "github.com/plgd-dev/hub/v2/test/oauth-server/test"
	pbTest "github.com/plgd-dev/hub/v2/test/pb"
	testService "github.com/plgd-dev/hub/v2/test/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestRequestHandlerGetResourcePendingCommands(t *testing.T) {
	deviceID := test.MustFindDeviceByName(test.TestDeviceName)
	type args struct {
		deviceID      string
		href          string
		commandFilter []pb.GetPendingCommandsRequest_Command
		accept        string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []*pb.PendingCommand
	}{
		{
			name: "retrieve " + deviceID + device.ResourceURI,
			args: args{
				deviceID: deviceID,
				href:     device.ResourceURI,
				accept:   pkgHttp.ApplicationProtoJsonContentType,
			},
			want: []*pb.PendingCommand{
				{
					Command: &pb.PendingCommand_ResourceCreatePending{
						ResourceCreatePending: pbTest.MakeResourceCreatePending(t, deviceID, device.ResourceURI, test.TestResourceDeviceResourceTypes, "",
							map[string]interface{}{
								"power": 1,
							}),
					},
				},
				{
					Command: &pb.PendingCommand_ResourceDeletePending{
						ResourceDeletePending: &events.ResourceDeletePending{
							ResourceId: &commands.ResourceId{
								DeviceId: deviceID,
								Href:     device.ResourceURI,
							},
							AuditContext:  commands.NewAuditContext(service.DeviceUserID, "", service.DeviceUserID),
							ResourceTypes: test.TestResourceDeviceResourceTypes,
						},
					},
				},
			},
		},
		{
			name: "filter create commands",
			args: args{
				deviceID:      deviceID,
				href:          device.ResourceURI,
				commandFilter: []pb.GetPendingCommandsRequest_Command{pb.GetPendingCommandsRequest_RESOURCE_CREATE},
				accept:        pkgHttp.ApplicationProtoJsonContentType,
			},
			want: []*pb.PendingCommand{
				{
					Command: &pb.PendingCommand_ResourceCreatePending{
						ResourceCreatePending: pbTest.MakeResourceCreatePending(t, deviceID, device.ResourceURI, test.TestResourceDeviceResourceTypes, "",
							map[string]interface{}{
								"power": 1,
							}),
					},
				},
			},
		},
		{
			name: "filter delete commands",
			args: args{
				deviceID:      deviceID,
				href:          device.ResourceURI,
				commandFilter: []pb.GetPendingCommandsRequest_Command{pb.GetPendingCommandsRequest_RESOURCE_DELETE},
				accept:        pkgHttp.ApplicationProtoJsonContentType,
			},
			want: []*pb.PendingCommand{
				{
					Command: &pb.PendingCommand_ResourceDeletePending{
						ResourceDeletePending: &events.ResourceDeletePending{
							ResourceId: &commands.ResourceId{
								DeviceId: deviceID,
								Href:     device.ResourceURI,
							},
							AuditContext:  commands.NewAuditContext(service.DeviceUserID, "", service.DeviceUserID),
							ResourceTypes: test.TestResourceDeviceResourceTypes,
						},
					},
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	testService.ClearDB(ctx, t)
	oauthShutdown := oauthTest.SetUp(t)
	idShutdown := idService.SetUp(t)
	raShutdown := raService.SetUp(t)
	rdShutdown := rdService.SetUp(t)
	grpcShutdown := grpcgwService.SetUp(t)
	caShutdown := caService.SetUp(t)
	secureGWShutdown := coapgwTest.SetUp(t)

	defer caShutdown()
	defer grpcShutdown()
	defer rdShutdown()
	defer raShutdown()
	defer idShutdown()
	defer oauthShutdown()

	shutdownHttp := httpgwTest.SetUp(t)
	defer shutdownHttp()

	token := oauthTest.GetDefaultAccessToken(t)
	ctx = pkgGrpc.CtxWithToken(ctx, token)

	conn, err := grpc.NewClient(config.GRPC_GW_HOST, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		RootCAs: test.GetRootCertificatePool(t),
	})))
	require.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	c := pb.NewGrpcGatewayClient(conn)

	deviceID, shutdownDevSim := test.OnboardDevSim(ctx, t, c, deviceID, config.ACTIVE_COAP_SCHEME+"://"+config.COAP_GW_HOST, test.GetAllBackendResourceLinks())
	defer shutdownDevSim()

	secureGWShutdown()

	createFn := func() {
		createCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		_, errC := c.CreateResource(createCtx, &pb.CreateResourceRequest{
			ResourceId: commands.NewResourceID(deviceID, device.ResourceURI),
			Content: &pb.Content{
				ContentType: message.AppOcfCbor.String(),
				Data: test.EncodeToCbor(t, map[string]interface{}{
					"power": 1,
				}),
			},
		})
		require.Error(t, errC)
	}
	createFn()
	retrieveFn := func() {
		retrieveCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		_, errG := c.GetResourceFromDevice(retrieveCtx, &pb.GetResourceFromDeviceRequest{
			ResourceId: commands.NewResourceID(deviceID, platform.ResourceURI),
		})
		require.Error(t, errG)
	}
	retrieveFn()
	updateFn := func() {
		updateCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		_, errU := c.UpdateResource(updateCtx, &pb.UpdateResourceRequest{
			ResourceId: commands.NewResourceID(deviceID, test.TestResourceLightInstanceHref("1")),
			Content: &pb.Content{
				ContentType: message.AppOcfCbor.String(),
				Data: test.EncodeToCbor(t, map[string]interface{}{
					"power": 1,
				}),
			},
		})
		require.Error(t, errU)
	}
	updateFn()
	deleteFn := func() {
		deleteCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		_, errD := c.DeleteResource(deleteCtx, &pb.DeleteResourceRequest{
			ResourceId: commands.NewResourceID(deviceID, device.ResourceURI),
		})
		require.Error(t, errD)
	}
	deleteFn()
	updateDeviceMetadataFn := func() {
		deleteCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		_, errU := c.UpdateDeviceMetadata(deleteCtx, &pb.UpdateDeviceMetadataRequest{
			DeviceId:    deviceID,
			TwinEnabled: false,
		})
		require.Error(t, errU)
	}
	updateDeviceMetadataFn()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := httpgwTest.NewRequest(http.MethodGet, uri.AliasResourcePendingCommands, nil).AuthToken(token).Accept(tt.args.accept)
			rb.DeviceId(tt.args.deviceID).ResourceHref(tt.args.href).AddCommandsFilter(httpgwTest.ToCommandsFilter(tt.args.commandFilter))
			resp := httpgwTest.HTTPDo(t, rb.Build())
			defer func() {
				_ = resp.Body.Close()
			}()

			var values []*pb.PendingCommand
			for {
				var v pb.PendingCommand
				err = httpTest.Unmarshal(resp.StatusCode, resp.Body, &v)
				if errors.Is(err, io.EOF) {
					break
				}
				require.NoError(t, err)
				values = append(values, &v)
			}
			pbTest.CmpPendingCmds(t, tt.want, values)
		})
	}
}
