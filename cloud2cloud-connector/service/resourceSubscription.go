package service

import (
	"context"
	"fmt"

	"github.com/go-ocf/cloud/cloud2cloud-connector/events"
	"github.com/go-ocf/cloud/cloud2cloud-connector/store"
	raCqrs "github.com/go-ocf/cloud/resource-aggregate/cqrs"
	pbCQRS "github.com/go-ocf/cloud/resource-aggregate/pb"
	pbRA "github.com/go-ocf/cloud/resource-aggregate/pb"
	"github.com/go-ocf/go-coap/v2/message"
	kitHttp "github.com/go-ocf/kit/net/http"
)

func (s *SubscriptionManager) subscribeToResource(ctx context.Context, linkedAccount store.LinkedAccount, linkedCloud store.LinkedCloud, correlationID, signingSecret, deviceID, resourceHrefLink string) (string, error) {
	resp, err := subscribe(ctx, "/devices/"+deviceID+"/"+resourceHrefLink+"/subscriptions", correlationID, events.SubscriptionRequest{
		URL:           s.eventsURL,
		EventTypes:    []events.EventType{events.EventType_ResourceChanged},
		SigningSecret: signingSecret,
	}, linkedAccount, linkedCloud)
	if err != nil {
		return "", fmt.Errorf("cannot subscribe to device %v for %v: %v", deviceID, linkedAccount.ID, err)
	}
	return resp.SubscriptionId, nil
}

func cancelResourceSubscription(ctx context.Context, linkedAccount store.LinkedAccount, linkedCloud store.LinkedCloud, deviceID, resourceID, subscriptionID string) error {
	err := cancelSubscription(ctx, "/devices/"+deviceID+"/"+resourceID+"/subscriptions/"+subscriptionID, linkedAccount, linkedCloud)
	if err != nil {
		return fmt.Errorf("cannot cancel resource subscription for %v: %v", linkedAccount.ID, err)
	}
	return nil
}

func (s *SubscriptionManager) HandleResourceChangedEvent(ctx context.Context, subscriptionData subscriptionData, header events.EventHeader, body []byte) error {
	coapContentFormat := int32(-1)
	switch header.ContentType {
	case message.AppCBOR.String():
		coapContentFormat = int32(message.AppCBOR)
	case message.AppOcfCbor.String():
		coapContentFormat = int32(message.AppOcfCbor)
	case message.AppJSON.String():
		coapContentFormat = int32(message.AppJSON)
	}

	_, err := s.raClient.NotifyResourceChanged(ctx, &pbRA.NotifyResourceChangedRequest{
		AuthorizationContext: &pbCQRS.AuthorizationContext{
			DeviceId: subscriptionData.subscription.DeviceID,
		},
		ResourceId: raCqrs.MakeResourceId(subscriptionData.subscription.DeviceID, kitHttp.CanonicalHref(subscriptionData.subscription.Href)),
		CommandMetadata: &pbCQRS.CommandMetadata{
			ConnectionId: Cloud2cloudConnectorConnectionId,
			Sequence:     header.SequenceNumber,
		},
		Content: &pbRA.Content{
			Data:              body,
			ContentType:       header.ContentType,
			CoapContentFormat: coapContentFormat,
		},
	})
	if err != nil {
		return fmt.Errorf("cannot update resource aggregate (%v) resource (%v) content changed: %v", subscriptionData.subscription.DeviceID, subscriptionData.subscription.Href, err)
	}

	return nil

}

func (s *SubscriptionManager) HandleResourceEvent(ctx context.Context, header events.EventHeader, body []byte, subscriptionData subscriptionData) error {
	switch header.EventType {
	case events.EventType_ResourceChanged:
		return s.HandleResourceChangedEvent(ctx, subscriptionData, header, body)
	}
	return fmt.Errorf("cannot handle resource event: unsupported Event-Type %v", header.EventType)
}
