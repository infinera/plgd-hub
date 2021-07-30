package service

import (
	"context"
	"errors"

	"github.com/plgd-dev/cloud/grpc-gateway/pb"
	"github.com/plgd-dev/cloud/pkg/log"
	kitNetGrpc "github.com/plgd-dev/cloud/pkg/net/grpc"
	"github.com/plgd-dev/cloud/resource-aggregate/commands"
	"github.com/plgd-dev/cloud/resource-aggregate/cqrs/eventstore"
	"github.com/plgd-dev/cloud/resource-aggregate/events"
	"github.com/plgd-dev/kit/strings"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type resourceEvent struct {
	srv pb.GrpcGateway_GetEventsServer
}

type resourceEventHandler func(eventstore.EventUnmarshaler) *pb.GetEventsResponse

func handleResourceLinksPublished(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.ResourceLinksPublished
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_ResourceLinksPublished{
			ResourceLinksPublished: &e,
		},
	}
}

func handleResourceLinksUnpublished(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.ResourceLinksUnpublished
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_ResourceLinksUnpublished{
			ResourceLinksUnpublished: &e,
		},
	}
}

func handleResourceLinksSnapshotTaken(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.ResourceLinksSnapshotTaken
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_ResourceLinksSnapshotTaken{
			ResourceLinksSnapshotTaken: &e,
		},
	}
}

func handleResourceChanged(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.ResourceChanged
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_ResourceChanged{
			ResourceChanged: &e,
		},
	}
}

func handleResourceUpdatePending(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.ResourceUpdatePending
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_ResourceUpdatePending{
			ResourceUpdatePending: &e,
		},
	}
}

func handleResourceUpdated(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.ResourceUpdated
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_ResourceUpdated{
			ResourceUpdated: &e,
		},
	}
}

func handleResourceRetrievePending(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.ResourceRetrievePending
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_ResourceRetrievePending{
			ResourceRetrievePending: &e,
		},
	}
}

func handleResourceRetrieved(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.ResourceRetrieved
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_ResourceRetrieved{
			ResourceRetrieved: &e,
		},
	}
}

func handleResourceDeletePending(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.ResourceDeletePending
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_ResourceDeletePending{
			ResourceDeletePending: &e,
		},
	}
}

func handleResourceDeleted(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.ResourceDeleted
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_ResourceDeleted{
			ResourceDeleted: &e,
		},
	}
}

func handleResourceCreatePending(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.ResourceCreatePending
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_ResourceCreatePending{
			ResourceCreatePending: &e,
		},
	}
}

func handleResourceCreated(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.ResourceCreated
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_ResourceCreated{
			ResourceCreated: &e,
		},
	}
}

func handleResourceStateSnapshotTaken(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.ResourceStateSnapshotTaken
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_ResourceStateSnapshotTaken{
			ResourceStateSnapshotTaken: &e,
		},
	}
}

func handleDeviceMetadataUpdatePending(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.DeviceMetadataUpdatePending
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_DeviceMetadataUpdatePending{
			DeviceMetadataUpdatePending: &e,
		},
	}
}

func handleDeviceMetadataUpdated(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.DeviceMetadataUpdated
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_DeviceMetadataUpdated{
			DeviceMetadataUpdated: &e,
		},
	}
}

func handleDeviceMetadataSnapshotTaken(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	var e events.DeviceMetadataSnapshotTaken
	if err := eu.Unmarshal(&e); err != nil {
		log.Errorf("failed to unmarshal event %v", eu.EventType())
		return nil
	}
	return &pb.GetEventsResponse{
		Type: &pb.GetEventsResponse_DeviceMetadataSnapshotTaken{
			DeviceMetadataSnapshotTaken: &e,
		},
	}
}

func handleEvent(eu eventstore.EventUnmarshaler) *pb.GetEventsResponse {
	log.Debug("handleEvent: EventType %v", eu)
	var handler resourceEventHandler

	switch eu.EventType() {
	case (&events.ResourceLinksPublished{}).EventType():
		handler = handleResourceLinksPublished
	case (&events.ResourceLinksUnpublished{}).EventType():
		handler = handleResourceLinksUnpublished
	case (&events.ResourceLinksSnapshotTaken{}).EventType():
		handler = handleResourceLinksSnapshotTaken
	case (&events.ResourceChanged{}).EventType():
		handler = handleResourceChanged
	case (&events.ResourceUpdatePending{}).EventType():
		handler = handleResourceUpdatePending
	case (&events.ResourceUpdated{}).EventType():
		handler = handleResourceUpdated
	case (&events.ResourceRetrievePending{}).EventType():
		handler = handleResourceRetrievePending
	case (&events.ResourceRetrieved{}).EventType():
		handler = handleResourceRetrieved
	case (&events.ResourceDeletePending{}).EventType():
		handler = handleResourceDeletePending
	case (&events.ResourceDeleted{}).EventType():
		handler = handleResourceDeleted
	case (&events.ResourceCreatePending{}).EventType():
		handler = handleResourceCreatePending
	case (&events.ResourceCreated{}).EventType():
		handler = handleResourceCreated
	case (&events.ResourceStateSnapshotTaken{}).EventType():
		handler = handleResourceStateSnapshotTaken
	case (&events.DeviceMetadataUpdatePending{}).EventType():
		handler = handleDeviceMetadataUpdatePending
	case (&events.DeviceMetadataUpdated{}).EventType():
		handler = handleDeviceMetadataUpdated
	case (&events.DeviceMetadataSnapshotTaken{}).EventType():
		handler = handleDeviceMetadataSnapshotTaken
	}

	if handler == nil {
		log.Errorf("unhandled event type %v", eu.EventType())
		return nil
	}

	return handler(eu)
}

func (p *resourceEvent) Handle(ctx context.Context, iter eventstore.Iter) error {
	log.Debug("resourceEvent.Handle")

	for {
		eu, ok := iter.Next(ctx)
		if !ok {
			break
		}
		if eu.EventType() == "" {
			return errors.New("cannot determine type of event")
		}
		resp := handleEvent(eu)
		if resp == nil {
			continue
		}
		if err := p.srv.Send(resp); err != nil {
			return err
		}
	}

	return iter.Err()
}

func getDeviceQueries(deviceIdFilter []string, userDeviceIds strings.Set) []eventstore.GetEventsQuery {
	var queries []eventstore.GetEventsQuery
	for _, deviceId := range deviceIdFilter {
		if _, ok := userDeviceIds[deviceId]; !ok {
			log.Debugf("permission denied, device with id %v skipped", deviceId)
			continue
		}
		queries = append(queries, eventstore.GetEventsQuery{
			GroupID: deviceId,
		})
	}
	return queries
}

func getResourceQueries(resourceFilter []string, userDeviceIds strings.Set) []eventstore.GetEventsQuery {
	var queries []eventstore.GetEventsQuery
	for _, resourceId := range resourceFilter {
		res := commands.ResourceIdFromString(resourceId)
		if res == nil {
			log.Errorf("invalid resourceIdFilter value %v", resourceFilter)
			continue
		}
		if !userDeviceIds.HasOneOf(res.GetDeviceId()) {
			log.Debugf("permission denied, resource belonging to device %v skipped", res.GetDeviceId())
			continue
		}
		queries = append(queries, eventstore.GetEventsQuery{
			GroupID:     res.GetDeviceId(),
			AggregateID: res.ToUUID(),
		})
	}
	return queries
}

func getUserDeviceQueries(userDeviceIds strings.Set) []eventstore.GetEventsQuery {
	var queries []eventstore.GetEventsQuery
	for device := range userDeviceIds {
		queries = append(queries, eventstore.GetEventsQuery{
			GroupID: device,
		})
	}
	return queries
}

func (r *RequestHandler) GetEvents(req *pb.GetEventsRequest, srv pb.GrpcGateway_GetEventsServer) error {
	owner, err := kitNetGrpc.OwnerFromMD(srv.Context())
	if err != nil {
		return log.LogAndReturnError(status.Errorf(codes.Unauthenticated, "cannot get owner: %v", err))
	}
	userDeviceIds, err := r.userDevicesManager.GetUserDevices(srv.Context(), owner)
	if err != nil {
		return log.LogAndReturnError(status.Errorf(status.Convert(err).Code(), "cannot get owned devices: %v", err))
	}
	if len(userDeviceIds) == 0 {
		log.Debugf("No devices found for user %v", owner)
		return nil
	}
	mapUserDeviceIds := make(strings.Set)
	for _, userDeviceId := range userDeviceIds {
		mapUserDeviceIds[userDeviceId] = struct{}{}
	}

	var queries []eventstore.GetEventsQuery
	if len(req.DeviceIdFilter) == 0 && len(req.ResourceIdFilter) == 0 {
		queries = getUserDeviceQueries(mapUserDeviceIds)
	} else {
		queries = getDeviceQueries(req.DeviceIdFilter, mapUserDeviceIds)
		queries = append(queries, getResourceQueries(req.ResourceIdFilter, mapUserDeviceIds)...)
	}

	err = r.eventStore.GetEvents(srv.Context(), queries, req.TimestampFilter, &resourceEvent{srv: srv})
	if err != nil {
		return log.LogAndReturnError(status.Errorf(status.Convert(err).Code(), "cannot get events: %v", err))
	}
	return nil
}