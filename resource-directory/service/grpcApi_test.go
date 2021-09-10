package service_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	coapgwTest "github.com/plgd-dev/cloud/coap-gateway/test"
	"github.com/plgd-dev/cloud/grpc-gateway/pb"
	grpcgwTest "github.com/plgd-dev/cloud/grpc-gateway/test"
	"github.com/plgd-dev/cloud/pkg/log"
	kitNetGrpc "github.com/plgd-dev/cloud/pkg/net/grpc"
	grpcClient "github.com/plgd-dev/cloud/pkg/net/grpc/client"
	"github.com/plgd-dev/cloud/resource-aggregate/commands"
	"github.com/plgd-dev/cloud/resource-aggregate/events"
	rdTest "github.com/plgd-dev/cloud/resource-directory/test"
	"github.com/plgd-dev/cloud/test"
	"github.com/plgd-dev/cloud/test/config"
	oauthTest "github.com/plgd-dev/cloud/test/oauth-server/test"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/kit/codec/cbor"
)

const TEST_TIMEOUT = time.Second * 30

func TestRequestHandler_SubscribeToEvents(t *testing.T) {
	deviceID := test.MustFindDeviceByName(test.TestDeviceName)
	type args struct {
		sub *pb.SubscribeToEvents
	}
	tests := []struct {
		name string
		args args
		want []*pb.Event
	}{
		{
			name: "invalid - invalid type subscription",
			args: args{
				sub: &pb.SubscribeToEvents{
					CorrelationId: "testToken",
				},
			},

			want: []*pb.Event{
				{
					Type: &pb.Event_OperationProcessed_{
						OperationProcessed: &pb.Event_OperationProcessed{
							ErrorStatus: &pb.Event_OperationProcessed_ErrorStatus{
								Code: pb.Event_OperationProcessed_ErrorStatus_OK,
							},
						},
					},
					CorrelationId: "testToken",
				},
				{
					Type: &pb.Event_SubscriptionCanceled_{
						SubscriptionCanceled: &pb.Event_SubscriptionCanceled{
							Reason: "not supported",
						},
					},
					CorrelationId: "testToken",
				},
			},
		},
		{
			name: "without IncludeCurrentState",
			args: args{
				sub: &pb.SubscribeToEvents{
					CorrelationId: "testToken",
					Action: &pb.SubscribeToEvents_CreateSubscription_{
						CreateSubscription: &pb.SubscribeToEvents_CreateSubscription{},
					},
				},
			},
			want: []*pb.Event{
				{
					Type: &pb.Event_OperationProcessed_{
						OperationProcessed: &pb.Event_OperationProcessed{
							ErrorStatus: &pb.Event_OperationProcessed_ErrorStatus{
								Code: pb.Event_OperationProcessed_ErrorStatus_OK,
							},
						},
					},
					CorrelationId: "testToken",
				},
			},
		},
		{
			name: "devices subscription - registered",
			args: args{
				sub: &pb.SubscribeToEvents{
					CorrelationId: "testToken",
					Action: &pb.SubscribeToEvents_CreateSubscription_{
						CreateSubscription: &pb.SubscribeToEvents_CreateSubscription{
							EventFilter: []pb.SubscribeToEvents_CreateSubscription_Event{
								pb.SubscribeToEvents_CreateSubscription_REGISTERED, pb.SubscribeToEvents_CreateSubscription_UNREGISTERED,
							},
							IncludeCurrentState: true,
						},
					},
				},
			},
			want: []*pb.Event{
				{
					Type: &pb.Event_OperationProcessed_{
						OperationProcessed: &pb.Event_OperationProcessed{
							ErrorStatus: &pb.Event_OperationProcessed_ErrorStatus{
								Code: pb.Event_OperationProcessed_ErrorStatus_OK,
							},
						},
					},
					CorrelationId: "testToken",
				},
				{
					Type: &pb.Event_DeviceRegistered_{
						DeviceRegistered: &pb.Event_DeviceRegistered{
							DeviceIds: []string{deviceID},
						},
					},
					CorrelationId: "testToken",
				},
			},
		},
		{
			name: "devices subscription - device metadata updated",
			args: args{
				sub: &pb.SubscribeToEvents{
					CorrelationId: "testToken",
					Action: &pb.SubscribeToEvents_CreateSubscription_{
						CreateSubscription: &pb.SubscribeToEvents_CreateSubscription{
							EventFilter: []pb.SubscribeToEvents_CreateSubscription_Event{
								pb.SubscribeToEvents_CreateSubscription_DEVICE_METADATA_UPDATED,
							},
							IncludeCurrentState: true,
						},
					},
				},
			},
			want: []*pb.Event{
				{
					Type: &pb.Event_OperationProcessed_{
						OperationProcessed: &pb.Event_OperationProcessed{
							ErrorStatus: &pb.Event_OperationProcessed_ErrorStatus{
								Code: pb.Event_OperationProcessed_ErrorStatus_OK,
							},
						},
					},
					CorrelationId: "testToken",
				},
				{
					Type: &pb.Event_DeviceMetadataUpdated{
						DeviceMetadataUpdated: &events.DeviceMetadataUpdated{
							DeviceId: deviceID,
							Status: &commands.ConnectionStatus{
								Value: commands.ConnectionStatus_ONLINE,
							},
						},
					},
					CorrelationId: "testToken",
				},
			},
		},
		{
			name: "device subscription - published",
			args: args{
				sub: &pb.SubscribeToEvents{
					CorrelationId: "testToken",
					Action: &pb.SubscribeToEvents_CreateSubscription_{
						CreateSubscription: &pb.SubscribeToEvents_CreateSubscription{
							DeviceIdFilter: []string{deviceID},
							EventFilter: []pb.SubscribeToEvents_CreateSubscription_Event{
								pb.SubscribeToEvents_CreateSubscription_RESOURCE_PUBLISHED, pb.SubscribeToEvents_CreateSubscription_RESOURCE_UNPUBLISHED,
							},
							IncludeCurrentState: true,
						},
					},
				},
			},
			want: []*pb.Event{
				{
					Type: &pb.Event_OperationProcessed_{
						OperationProcessed: &pb.Event_OperationProcessed{
							ErrorStatus: &pb.Event_OperationProcessed_ErrorStatus{
								Code: pb.Event_OperationProcessed_ErrorStatus_OK,
							},
						},
					},
					CorrelationId: "testToken",
				},
				test.ResourceLinkToPublishEvent(deviceID, "testToken", test.GetAllBackendResourceLinks()),
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), TEST_TIMEOUT)
	defer cancel()

	tearDown := test.SetUp(ctx, t)
	defer tearDown()
	ctx = kitNetGrpc.CtxWithToken(ctx, oauthTest.GetServiceToken(t))

	rdConn, err := grpcClient.New(config.MakeGrpcClientConfig(config.RESOURCE_DIRECTORY_HOST), log.Get())
	require.NoError(t, err)
	defer func() {
		_ = rdConn.Close()
	}()
	c := pb.NewGrpcGatewayClient(rdConn.GRPC())

	_, shutdownDevSim := test.OnboardDevSim(ctx, t, c, deviceID, config.GW_HOST, test.GetAllBackendResourceLinks())
	defer shutdownDevSim()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := c.SubscribeToEvents(ctx)
			require.NoError(t, err)
			defer func() {
				err := client.CloseSend()
				require.NoError(t, err)
			}()
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for _, w := range tt.want {
					ev, err := client.Recv()
					require.NoError(t, err)
					ev.SubscriptionId = w.SubscriptionId
					if ev.GetResourcePublished() != nil {
						test.CleanUpResourceLinksPublished(ev.GetResourcePublished())
					}
					if w.GetResourcePublished() != nil {
						test.CleanUpResourceLinksPublished(w.GetResourcePublished())
					}
					if ev.GetDeviceMetadataUpdated() != nil {
						ev.GetDeviceMetadataUpdated().EventMetadata = nil
						ev.GetDeviceMetadataUpdated().AuditContext = nil
						if ev.GetDeviceMetadataUpdated().GetStatus() != nil {
							ev.GetDeviceMetadataUpdated().GetStatus().ValidUntil = 0
						}
					}
					test.CheckProtobufs(t, tt.want, ev, test.RequireToCheckFunc(require.Contains))
				}
			}()
			err = client.Send(tt.args.sub)
			require.NoError(t, err)
			wg.Wait()
		})
	}
}

func TestRequestHandler_Issue270(t *testing.T) {
	deviceID := test.MustFindDeviceByName(test.TestDeviceName)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*240)
	defer cancel()

	coapgwCfg := coapgwTest.MakeConfig(t)
	rdCfg := rdTest.MakeConfig(t)
	rdCfg.Clients.AuthServer.PullFrequency = time.Second * 15
	rdCfg.Clients.AuthServer.CacheExpiration = time.Minute

	grpcgwCfg := grpcgwTest.MakeConfig(t)

	tearDown := test.SetUp(ctx, t, test.WithCOAPGWConfig(coapgwCfg), test.WithRDConfig(rdCfg), test.WithGRPCGWConfig(grpcgwCfg))
	defer tearDown()
	ctx = kitNetGrpc.CtxWithToken(ctx, oauthTest.GetServiceToken(t))

	rdConn, err := grpcClient.New(config.MakeGrpcClientConfig(config.RESOURCE_DIRECTORY_HOST), log.Get())
	require.NoError(t, err)
	defer func() {
		_ = rdConn.Close()
	}()
	c := pb.NewGrpcGatewayClient(rdConn.GRPC())

	client, err := c.SubscribeToEvents(ctx)
	require.NoError(t, err)

	err = client.Send(&pb.SubscribeToEvents{
		CorrelationId: "testToken",
		Action: &pb.SubscribeToEvents_CreateSubscription_{
			CreateSubscription: &pb.SubscribeToEvents_CreateSubscription{
				EventFilter: []pb.SubscribeToEvents_CreateSubscription_Event{
					pb.SubscribeToEvents_CreateSubscription_DEVICE_METADATA_UPDATED, pb.SubscribeToEvents_CreateSubscription_REGISTERED, pb.SubscribeToEvents_CreateSubscription_UNREGISTERED,
				},
				IncludeCurrentState: true,
			},
		},
	})
	require.NoError(t, err)

	ev, err := client.Recv()
	require.NoError(t, err)
	expectedEvent := &pb.Event{
		SubscriptionId: ev.SubscriptionId,
		Type: &pb.Event_OperationProcessed_{
			OperationProcessed: &pb.Event_OperationProcessed{
				ErrorStatus: &pb.Event_OperationProcessed_ErrorStatus{
					Code: pb.Event_OperationProcessed_ErrorStatus_OK,
				},
			},
		},
		CorrelationId: "testToken",
	}
	fmt.Printf("SUBSCRIPTION ID: %v\n", ev.SubscriptionId)
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))

	ev, err = client.Recv()
	require.NoError(t, err)
	expectedEvent = &pb.Event{
		SubscriptionId: ev.SubscriptionId,
		Type: &pb.Event_DeviceRegistered_{
			DeviceRegistered: &pb.Event_DeviceRegistered{
				DeviceIds: []string{},
			},
		},
		CorrelationId: "testToken",
	}
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))

	deviceID, shutdownDevSim := test.OnboardDevSim(ctx, t, c, deviceID, config.GW_HOST, test.GetAllBackendResourceLinks())

	time.Sleep(time.Second * 10)

	ev, err = client.Recv()
	require.NoError(t, err)
	expectedEvent = &pb.Event{
		SubscriptionId: ev.SubscriptionId,
		Type: &pb.Event_DeviceRegistered_{
			DeviceRegistered: &pb.Event_DeviceRegistered{
				DeviceIds: []string{deviceID},
			},
		},
		CorrelationId: "testToken",
	}
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))

	ev, err = client.Recv()
	require.NoError(t, err)
	if ev.GetDeviceMetadataUpdated() != nil {
		ev.GetDeviceMetadataUpdated().EventMetadata = nil
		ev.GetDeviceMetadataUpdated().AuditContext = nil
		if ev.GetDeviceMetadataUpdated().GetStatus() != nil {
			ev.GetDeviceMetadataUpdated().GetStatus().ValidUntil = 0
		}
	}
	expectedEvent = &pb.Event{
		SubscriptionId: ev.SubscriptionId,
		Type: &pb.Event_DeviceMetadataUpdated{
			DeviceMetadataUpdated: &events.DeviceMetadataUpdated{
				DeviceId: deviceID,
				Status: &commands.ConnectionStatus{
					Value: commands.ConnectionStatus_ONLINE,
				},
			},
		},
		CorrelationId: "testToken",
	}
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))

	shutdownDevSim()
	run := true
	for run {
		ev, err = client.Recv()
		require.NoError(t, err)

		t.Logf("ev after shutdown: %v\n", ev)

		switch {
		case ev.GetDeviceUnregistered() != nil:
			expectedEvent = &pb.Event{
				SubscriptionId: ev.SubscriptionId,
				Type: &pb.Event_DeviceUnregistered_{
					DeviceUnregistered: &pb.Event_DeviceUnregistered{
						DeviceIds: []string{deviceID},
					},
				},
				CorrelationId: "testToken",
			}
			test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))
			run = false
		}
	}
}

func TestRequestHandler_ValidateEventsFlow(t *testing.T) {
	deviceID := test.MustFindDeviceByName(test.TestDeviceName)
	ctx, cancel := context.WithTimeout(context.Background(), TEST_TIMEOUT)
	defer cancel()

	tearDown := test.SetUp(ctx, t)
	defer tearDown()
	ctx = kitNetGrpc.CtxWithToken(ctx, oauthTest.GetServiceToken(t))

	rdConn, err := grpcClient.New(config.MakeGrpcClientConfig(config.RESOURCE_DIRECTORY_HOST), log.Get())
	require.NoError(t, err)
	defer func() {
		_ = rdConn.Close()
	}()
	c := pb.NewGrpcGatewayClient(rdConn.GRPC())

	grpcConn, err := grpcClient.New(config.MakeGrpcClientConfig(config.GRPC_HOST), log.Get())
	require.NoError(t, err)
	defer func() {
		_ = grpcConn.Close()
	}()
	grpcClient := pb.NewGrpcGatewayClient(grpcConn.GRPC())

	deviceID, shutdownDevSim := test.OnboardDevSim(ctx, t, c, deviceID, config.GW_HOST, test.GetAllBackendResourceLinks())

	client, err := c.SubscribeToEvents(ctx)
	require.NoError(t, err)

	err = client.Send(&pb.SubscribeToEvents{
		CorrelationId: "testToken",
		Action: &pb.SubscribeToEvents_CreateSubscription_{
			CreateSubscription: &pb.SubscribeToEvents_CreateSubscription{
				EventFilter: []pb.SubscribeToEvents_CreateSubscription_Event{
					pb.SubscribeToEvents_CreateSubscription_DEVICE_METADATA_UPDATED, pb.SubscribeToEvents_CreateSubscription_REGISTERED, pb.SubscribeToEvents_CreateSubscription_UNREGISTERED,
				},
				IncludeCurrentState: true,
			},
		},
	})
	require.NoError(t, err)

	ev, err := client.Recv()
	require.NoError(t, err)
	expectedEvent := &pb.Event{
		SubscriptionId: ev.SubscriptionId,
		Type: &pb.Event_OperationProcessed_{
			OperationProcessed: &pb.Event_OperationProcessed{
				ErrorStatus: &pb.Event_OperationProcessed_ErrorStatus{
					Code: pb.Event_OperationProcessed_ErrorStatus_OK,
				},
			},
		},
		CorrelationId: "testToken",
	}
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))

	ev, err = client.Recv()
	require.NoError(t, err)
	expectedEvent = &pb.Event{
		SubscriptionId: ev.SubscriptionId,
		Type: &pb.Event_DeviceRegistered_{
			DeviceRegistered: &pb.Event_DeviceRegistered{
				DeviceIds: []string{deviceID},
			},
		},
		CorrelationId: "testToken",
	}
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))

	for {
		ev, err = client.Recv()
		require.NoError(t, err)
		if ev.GetDeviceMetadataUpdated() != nil && ev.GetDeviceMetadataUpdated().GetDeviceId() == deviceID && ev.GetDeviceMetadataUpdated().GetStatus().GetValue() == commands.ConnectionStatus_ONLINE {
			break
		}
		continue
	}
	if ev.GetDeviceMetadataUpdated() != nil {
		ev.GetDeviceMetadataUpdated().EventMetadata = nil
		ev.GetDeviceMetadataUpdated().AuditContext = nil
		if ev.GetDeviceMetadataUpdated().GetStatus() != nil {
			ev.GetDeviceMetadataUpdated().GetStatus().ValidUntil = 0
		}
	}
	expectedEvent = &pb.Event{
		SubscriptionId: ev.SubscriptionId,
		Type: &pb.Event_DeviceMetadataUpdated{
			DeviceMetadataUpdated: &events.DeviceMetadataUpdated{
				DeviceId: deviceID,
				Status: &commands.ConnectionStatus{
					Value: commands.ConnectionStatus_ONLINE,
				},
			},
		},
		CorrelationId: "testToken",
	}
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))

	err = client.Send(&pb.SubscribeToEvents{
		CorrelationId: "testToken",
		Action: &pb.SubscribeToEvents_CreateSubscription_{
			CreateSubscription: &pb.SubscribeToEvents_CreateSubscription{
				ResourceIdFilter: []string{commands.NewResourceID(deviceID, "/light/2").ToString()},
				EventFilter: []pb.SubscribeToEvents_CreateSubscription_Event{
					pb.SubscribeToEvents_CreateSubscription_RESOURCE_CHANGED,
				},
				IncludeCurrentState: true,
			},
		},
	})
	require.NoError(t, err)

	ev, err = client.Recv()
	require.NoError(t, err)
	expectedEvent = &pb.Event{
		SubscriptionId: ev.SubscriptionId,
		Type: &pb.Event_OperationProcessed_{
			OperationProcessed: &pb.Event_OperationProcessed{
				ErrorStatus: &pb.Event_OperationProcessed_ErrorStatus{
					Code: pb.Event_OperationProcessed_ErrorStatus_OK,
				},
			},
		},
		CorrelationId: "testToken",
	}
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))
	subContentChangedID := ev.SubscriptionId

	ev, err = client.Recv()
	require.NoError(t, err)
	expectedEvent = &pb.Event{
		SubscriptionId: subContentChangedID,
		Type: &pb.Event_ResourceChanged{
			ResourceChanged: &events.ResourceChanged{
				ResourceId: commands.NewResourceID(deviceID, "/light/2"),
				Content: &commands.Content{
					CoapContentFormat: int32(message.AppOcfCbor),
					ContentType:       message.AppOcfCbor.String(),
					Data: func() []byte {
						ret, err := base64.StdEncoding.DecodeString("v2JydJ9qY29yZS5saWdodP9iaWafaW9pYy5pZi5yd29vaWMuaWYuYmFzZWxpbmX/ZXN0YXRl9GVwb3dlcgBkbmFtZWVMaWdodP8=")
						require.NoError(t, err)
						return ret
					}(),
				},
				Status:        commands.Status_OK,
				AuditContext:  ev.GetResourceChanged().GetAuditContext(),
				EventMetadata: ev.GetResourceChanged().GetEventMetadata(),
			},
		},
		CorrelationId: "testToken",
	}
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))

	err = client.Send(&pb.SubscribeToEvents{
		CorrelationId: "updatePending + resourceUpdated",
		Action: &pb.SubscribeToEvents_CreateSubscription_{
			CreateSubscription: &pb.SubscribeToEvents_CreateSubscription{
				DeviceIdFilter: []string{deviceID},
				EventFilter: []pb.SubscribeToEvents_CreateSubscription_Event{
					pb.SubscribeToEvents_CreateSubscription_RESOURCE_UPDATE_PENDING, pb.SubscribeToEvents_CreateSubscription_RESOURCE_UPDATED,
				},
				IncludeCurrentState: true,
			},
		},
	})
	require.NoError(t, err)

	ev, err = client.Recv()
	require.NoError(t, err)
	expectedEvent = &pb.Event{
		SubscriptionId: ev.SubscriptionId,
		Type: &pb.Event_OperationProcessed_{
			OperationProcessed: &pb.Event_OperationProcessed{
				ErrorStatus: &pb.Event_OperationProcessed_ErrorStatus{
					Code: pb.Event_OperationProcessed_ErrorStatus_OK,
				},
			},
		},
		CorrelationId: "updatePending + resourceUpdated",
	}
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))
	subUpdatedID := ev.SubscriptionId

	_, err = grpcClient.UpdateResource(ctx, &pb.UpdateResourceRequest{
		ResourceId: commands.NewResourceID(deviceID, "/light/2"),
		Content: &pb.Content{
			ContentType: message.AppOcfCbor.String(),
			Data: func() []byte {
				v := map[string]interface{}{
					"power": 99,
				}
				d, err := cbor.Encode(v)
				require.NoError(t, err)
				return d
			}(),
		},
	})
	require.NoError(t, err)

	var updCorrelationID string
	for i := 0; i < 3; i++ {
		ev, err = client.Recv()
		require.NoError(t, err)
		switch {
		case ev.GetResourceUpdatePending() != nil:
			expectedEvent = &pb.Event{
				SubscriptionId: subUpdatedID,
				Type: &pb.Event_ResourceUpdatePending{
					ResourceUpdatePending: &events.ResourceUpdatePending{
						ResourceId: commands.NewResourceID(deviceID, "/light/2"),
						Content: &commands.Content{
							ContentType:       message.AppOcfCbor.String(),
							CoapContentFormat: -1,
							Data: func() []byte {
								v := map[string]interface{}{
									"power": 99,
								}
								d, err := cbor.Encode(v)
								require.NoError(t, err)
								return d
							}(),
						},
						AuditContext:  ev.GetResourceUpdatePending().GetAuditContext(),
						EventMetadata: ev.GetResourceUpdatePending().GetEventMetadata(),
					},
				},
				CorrelationId: "updatePending + resourceUpdated",
			}
			test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))
			updCorrelationID = ev.GetResourceUpdatePending().GetAuditContext().GetCorrelationId()
		case ev.GetResourceUpdated() != nil:
			expectedEvent = &pb.Event{
				SubscriptionId: subUpdatedID,
				Type: &pb.Event_ResourceUpdated{
					ResourceUpdated: &events.ResourceUpdated{
						ResourceId:    commands.NewResourceID(deviceID, "/light/2"),
						Status:        commands.Status_OK,
						Content:       ev.GetResourceUpdated().GetContent(),
						AuditContext:  commands.NewAuditContext(ev.GetResourceUpdated().GetAuditContext().GetUserId(), updCorrelationID),
						EventMetadata: ev.GetResourceUpdated().GetEventMetadata(),
					},
				},
				CorrelationId: "updatePending + resourceUpdated",
			}
			test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))
		case ev.GetResourceChanged() != nil:
			expectedEvent = &pb.Event{
				SubscriptionId: subContentChangedID,
				Type: &pb.Event_ResourceChanged{
					ResourceChanged: &events.ResourceChanged{
						ResourceId: commands.NewResourceID(deviceID, "/light/2"),
						Content: &commands.Content{
							CoapContentFormat: int32(message.AppOcfCbor),
							ContentType:       message.AppOcfCbor.String(),
							Data:              []byte("\277estate\364epower\030cdnameeLight\377"),
						},
						Status:        commands.Status_OK,
						AuditContext:  ev.GetResourceChanged().GetAuditContext(),
						EventMetadata: ev.GetResourceChanged().GetEventMetadata(),
					},
				},
				CorrelationId: "testToken",
			}
			test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))
		}
	}
	_, err = grpcClient.UpdateResource(ctx, &pb.UpdateResourceRequest{
		ResourceId: commands.NewResourceID(deviceID, "/light/2"),
		Content: &pb.Content{
			ContentType: message.AppOcfCbor.String(),
			Data: func() []byte {
				v := map[string]interface{}{
					"power": 0,
				}
				d, err := cbor.Encode(v)
				require.NoError(t, err)
				return d
			}(),
		},
	})
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		ev, err = client.Recv()
		require.NoError(t, err)
		switch {
		case ev.GetResourceUpdatePending() != nil:
			expectedEvent = &pb.Event{
				SubscriptionId: subUpdatedID,
				Type: &pb.Event_ResourceUpdatePending{
					ResourceUpdatePending: &events.ResourceUpdatePending{
						ResourceId: commands.NewResourceID(deviceID, "/light/2"),
						Content: &commands.Content{
							ContentType:       message.AppOcfCbor.String(),
							CoapContentFormat: -1,
							Data: func() []byte {
								v := map[string]interface{}{
									"power": 0,
								}
								d, err := cbor.Encode(v)
								require.NoError(t, err)
								return d
							}(),
						},
						AuditContext:  ev.GetResourceUpdatePending().GetAuditContext(),
						EventMetadata: ev.GetResourceUpdatePending().GetEventMetadata(),
					},
				},
				CorrelationId: "updatePending + resourceUpdated",
			}
			test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))
			updCorrelationID = ev.GetResourceUpdatePending().GetAuditContext().GetCorrelationId()
		case ev.GetResourceUpdated() != nil:
			expectedEvent = &pb.Event{
				SubscriptionId: subUpdatedID,
				Type: &pb.Event_ResourceUpdated{
					ResourceUpdated: &events.ResourceUpdated{
						ResourceId:    commands.NewResourceID(deviceID, "/light/2"),
						Status:        commands.Status_OK,
						Content:       ev.GetResourceUpdated().GetContent(),
						AuditContext:  commands.NewAuditContext(ev.GetResourceUpdated().GetAuditContext().GetUserId(), updCorrelationID),
						EventMetadata: ev.GetResourceUpdated().GetEventMetadata(),
					},
				},
				CorrelationId: "updatePending + resourceUpdated",
			}
			test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))
		case ev.GetResourceChanged() != nil:
			expectedEvent = &pb.Event{
				SubscriptionId: subContentChangedID,
				Type: &pb.Event_ResourceChanged{
					ResourceChanged: &events.ResourceChanged{
						ResourceId: commands.NewResourceID(deviceID, "/light/2"),
						Content: &commands.Content{
							CoapContentFormat: int32(message.AppOcfCbor),
							ContentType:       message.AppOcfCbor.String(),
							Data:              []byte("\277estate\364epower\000dnameeLight\377"),
						},
						Status:        commands.Status_OK,
						AuditContext:  ev.GetResourceChanged().GetAuditContext(),
						EventMetadata: ev.GetResourceChanged().GetEventMetadata(),
					},
				},
				CorrelationId: "testToken",
			}
			test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))
		}
	}

	err = client.Send(&pb.SubscribeToEvents{
		CorrelationId: "receivePending + resourceReceived",
		Action: &pb.SubscribeToEvents_CreateSubscription_{
			CreateSubscription: &pb.SubscribeToEvents_CreateSubscription{
				DeviceIdFilter: []string{deviceID},
				EventFilter: []pb.SubscribeToEvents_CreateSubscription_Event{
					pb.SubscribeToEvents_CreateSubscription_RESOURCE_RETRIEVE_PENDING, pb.SubscribeToEvents_CreateSubscription_RESOURCE_RETRIEVED,
				},
				IncludeCurrentState: true,
			},
		},
	})
	require.NoError(t, err)

	ev, err = client.Recv()
	require.NoError(t, err)
	expectedEvent = &pb.Event{
		SubscriptionId: ev.SubscriptionId,
		Type: &pb.Event_OperationProcessed_{
			OperationProcessed: &pb.Event_OperationProcessed{
				ErrorStatus: &pb.Event_OperationProcessed_ErrorStatus{
					Code: pb.Event_OperationProcessed_ErrorStatus_OK,
				},
			},
		},
		CorrelationId: "receivePending + resourceReceived",
	}
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))
	subReceivedID := ev.SubscriptionId

	_, err = grpcClient.GetResourceFromDevice(ctx, &pb.GetResourceFromDeviceRequest{
		ResourceId: commands.NewResourceID(deviceID, "/light/2"),
	})
	require.NoError(t, err)
	ev, err = client.Recv()
	require.NoError(t, err)
	expectedEvent = &pb.Event{
		SubscriptionId: subReceivedID,
		Type: &pb.Event_ResourceRetrievePending{
			ResourceRetrievePending: &events.ResourceRetrievePending{
				ResourceId:    commands.NewResourceID(deviceID, "/light/2"),
				AuditContext:  ev.GetResourceRetrievePending().GetAuditContext(),
				EventMetadata: ev.GetResourceRetrievePending().GetEventMetadata(),
			},
		},
		CorrelationId: "receivePending + resourceReceived",
	}
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))
	recvCorrelationID := ev.GetResourceRetrievePending().GetAuditContext().GetCorrelationId()

	ev, err = client.Recv()
	require.NoError(t, err)
	expectedEvent = &pb.Event{
		SubscriptionId: subReceivedID,
		Type: &pb.Event_ResourceRetrieved{
			ResourceRetrieved: &events.ResourceRetrieved{
				ResourceId: commands.NewResourceID(deviceID, "/light/2"),
				Content: &commands.Content{
					ContentType:       message.AppOcfCbor.String(),
					CoapContentFormat: int32(message.AppOcfCbor),
					Data:              []byte("\277estate\364epower\000dnameeLight\377"),
				},
				Status:        commands.Status_OK,
				AuditContext:  commands.NewAuditContext(ev.GetResourceRetrieved().GetAuditContext().GetUserId(), recvCorrelationID),
				EventMetadata: ev.GetResourceRetrieved().GetEventMetadata(),
			},
		},
		CorrelationId: "receivePending + resourceReceived",
	}
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))

	shutdownDevSim()

	run := true
	for run {
		ev, err = client.Recv()
		require.NoError(t, err)

		t.Logf("ev after shutdown: %v\n", ev)

		switch {
		case ev.GetDeviceUnregistered() != nil:
			expectedEvent = &pb.Event{
				SubscriptionId: ev.SubscriptionId,
				Type: &pb.Event_DeviceUnregistered_{
					DeviceUnregistered: &pb.Event_DeviceUnregistered{
						DeviceIds: []string{deviceID},
					},
				},
				CorrelationId: "testToken",
			}
			test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))
			run = false
		}
	}
}
