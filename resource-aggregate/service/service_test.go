package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	pbAS "github.com/plgd-dev/cloud/authorization/pb"
	authService "github.com/plgd-dev/cloud/authorization/test"
	"github.com/plgd-dev/cloud/pkg/log"
	kitNetGrpc "github.com/plgd-dev/cloud/pkg/net/grpc"
	"github.com/plgd-dev/cloud/pkg/net/grpc/client"
	"github.com/plgd-dev/cloud/resource-aggregate/service"
	"github.com/plgd-dev/cloud/resource-aggregate/test"
	testCfg "github.com/plgd-dev/cloud/test/config"
	oauthTest "github.com/plgd-dev/cloud/test/oauth-server/test"
)

func TestPublishUnpublish(t *testing.T) {
	config := test.MakeConfig(t)
	config.APIs.GRPC.Addr = "localhost:9888"
	config.Clients.Eventstore.SnapshotThreshold = 1

	oauthShutdown := oauthTest.SetUp(t)
	defer oauthShutdown()

	authShutdown := authService.SetUp(t)
	defer authShutdown()

	raShutdown := test.New(t, config)
	defer raShutdown()

	ctx := kitNetGrpc.CtxWithToken(context.Background(), oauthTest.GetDefaultServiceToken(t))

	authConn, err := client.New(testCfg.MakeGrpcClientConfig(config.Clients.AuthServer.Connection.Addr), log.Get())
	require.NoError(t, err)
	defer func() {
		_ = authConn.Close()
	}()
	authClient := pbAS.NewAuthorizationServiceClient(authConn.GRPC())

	raConn, err := client.New(testCfg.MakeGrpcClientConfig(config.APIs.GRPC.Addr), log.Get())
	require.NoError(t, err)
	defer func() {
		_ = raConn.Close()
	}()
	raClient := service.NewResourceAggregateClient(raConn.GRPC())

	deviceId := "dev0"
	href := "/oic/p"
	_, err = authClient.AddDevice(ctx, &pbAS.AddDeviceRequest{
		DeviceId: deviceId,
	})
	require.NoError(t, err)
	defer func() {
		_, err = authClient.DeleteDevices(ctx, &pbAS.DeleteDevicesRequest{
			DeviceIds: []string{deviceId},
		})
		require.NoError(t, err)
	}()

	pubReq := testMakePublishResourceRequest(deviceId, []string{href})
	_, err = raClient.PublishResourceLinks(ctx, pubReq)
	require.NoError(t, err)

	unpubReq := testMakeUnpublishResourceRequest(deviceId, []string{href})
	_, err = raClient.UnpublishResourceLinks(ctx, unpubReq)
	require.NoError(t, err)
}
