package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/plgd-dev/hub/v2/grpc-gateway/pb"
	"github.com/plgd-dev/hub/v2/resource-aggregate/events"
)

type DevicesObservationEvent_type uint8

const (
	DevicesObservationEvent_ONLINE       DevicesObservationEvent_type = 0
	DevicesObservationEvent_OFFLINE      DevicesObservationEvent_type = 1
	DevicesObservationEvent_REGISTERED   DevicesObservationEvent_type = 2
	DevicesObservationEvent_UNREGISTERED DevicesObservationEvent_type = 3
)

type DevicesObservationEvent struct {
	DeviceIDs []string
	Event     DevicesObservationEvent_type
}

type DevicesObservationHandler = interface {
	Handle(ctx context.Context, event DevicesObservationEvent) error
	OnClose()
	Error(err error)
}

type devicesObservation struct {
	h                  DevicesObservationHandler
	removeSubscription func()
}

func (o *devicesObservation) HandleDeviceMetadataUpdated(ctx context.Context, val *events.DeviceMetadataUpdated) error {
	if val.GetConnection() == nil {
		return nil
	}
	event := DevicesObservationEvent_OFFLINE
	if val.GetConnection().IsOnline() {
		event = DevicesObservationEvent_ONLINE
	}
	return o.h.Handle(ctx, DevicesObservationEvent{
		DeviceIDs: []string{val.GetDeviceId()},
		Event:     event,
	})
}

func (o *devicesObservation) HandleDeviceRegistered(ctx context.Context, val *pb.Event_DeviceRegistered) error {
	return o.h.Handle(ctx, DevicesObservationEvent{
		DeviceIDs: val.GetDeviceIds(),
		Event:     DevicesObservationEvent_REGISTERED,
	})
}

func (o *devicesObservation) HandleDeviceUnregistered(ctx context.Context, val *pb.Event_DeviceUnregistered) error {
	return o.h.Handle(ctx, DevicesObservationEvent{
		DeviceIDs: val.GetDeviceIds(),
		Event:     DevicesObservationEvent_UNREGISTERED,
	})
}

func (o *devicesObservation) OnClose() {
	o.removeSubscription()
	o.h.OnClose()
}

func (o *devicesObservation) Error(err error) {
	o.removeSubscription()
	o.h.Error(err)
}

func (c *Client) ObserveDevices(ctx context.Context, handler DevicesObservationHandler) (string, error) {
	ID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	sub, err := c.NewDevicesSubscription(ctx, &devicesObservation{
		h: handler,
		removeSubscription: func() {
			// we can ignore err during removeSubscription, if err != nil then some other
			// part of code already removed the subscription
			_, _ = c.stopObservingDevices(ID.String())
		},
	})
	if err != nil {
		return "", err
	}
	c.insertSubscription(ID.String(), sub)

	return ID.String(), err
}

func (c *Client) stopObservingDevices(observationID string) (wait func(), err error) {
	s, err := c.popSubscription(observationID)
	if err != nil {
		return nil, err
	}
	return s.Cancel()
}

func (c *Client) StopObservingDevices(ctx context.Context, observationID string) error {
	wait, err := c.stopObservingDevices(observationID)
	if err != nil {
		return err
	}
	wait()
	return nil
}
