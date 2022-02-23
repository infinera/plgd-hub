package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/plgd-dev/go-coap/v2/message"
	coapCodes "github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
	"github.com/plgd-dev/hub/v2/coap-gateway/coapconv"
	"github.com/plgd-dev/hub/v2/coap-gateway/service/observation"
	grpcgwClient "github.com/plgd-dev/hub/v2/grpc-gateway/client"
	"github.com/plgd-dev/hub/v2/identity-store/events"
	kitNetGrpc "github.com/plgd-dev/hub/v2/pkg/net/grpc"
	"github.com/plgd-dev/hub/v2/pkg/security/jwt"
	"github.com/plgd-dev/hub/v2/pkg/strings"
	"github.com/plgd-dev/hub/v2/pkg/sync/task/future"
	"github.com/plgd-dev/hub/v2/resource-aggregate/commands"
	"github.com/plgd-dev/kit/v2/codec/cbor"
)

type CoapSignInReq struct {
	DeviceID    string `json:"di"`
	UserID      string `json:"uid"`
	AccessToken string `json:"accesstoken"`
	Login       bool   `json:"login"`
}

type CoapSignInResp struct {
	ExpiresIn int64 `json:"expiresin"`
}

/// Check that all required request fields are set
func (request CoapSignInReq) checkOAuthRequest() error {
	if request.DeviceID == "" {
		return fmt.Errorf("invalid device id")
	}
	if request.UserID == "" {
		return fmt.Errorf("invalid user id")
	}
	if request.AccessToken == "" {
		return fmt.Errorf("invalid access token")
	}
	return nil
}

/// Update empty values
func (request CoapSignInReq) updateOAUthRequestIfEmpty(deviceID, userID, accessToken string) CoapSignInReq {
	if request.DeviceID == "" {
		request.DeviceID = deviceID
	}
	if request.UserID == "" {
		request.UserID = userID
	}
	if request.AccessToken == "" {
		request.AccessToken = accessToken
	}
	return request
}

/// Get data for sign in response
func getSignInContent(expiresIn int64, options message.Options) (message.MediaType, []byte, error) {
	coapResp := CoapSignInResp{
		ExpiresIn: expiresIn,
	}

	accept := coapconv.GetAccept(options)
	encode, err := coapconv.GetEncoder(accept)
	if err != nil {
		return 0, nil, err
	}
	out, err := encode(coapResp)
	if err != nil {
		return 0, nil, err

	}
	return accept, out, nil
}

func setNewDeviceSubscriber(client *Client, owner, deviceID string) error {
	getContext := func() (context.Context, context.CancelFunc) {
		return client.GetContext(), func() {
			// no-op
		}
	}

	deviceSubscriber, err := grpcgwClient.NewDeviceSubscriber(getContext, owner, deviceID,
		func() func() (when time.Time, err error) {
			var count uint64
			maxRand := client.server.config.APIs.COAP.KeepAlive.Timeout / 2
			if maxRand <= 0 {
				maxRand = time.Second * 10
			}
			return func() (when time.Time, err error) {
				count++
				r := rand.Int63n(int64(maxRand) / 2)
				next := time.Now().Add(client.server.config.APIs.COAP.KeepAlive.Timeout + time.Duration(r))
				client.Debugf("next iteration %v of retrying reconnect to grpc-client will be at %v", count, next)
				return next, nil
			}
		}, client.server.rdClient, client.server.resourceSubscriber)
	if err != nil {
		return fmt.Errorf("cannot create device subscription for device %v: %w", deviceID, err)
	}
	oldDeviceSubscriber := client.replaceDeviceSubscriber(deviceSubscriber)
	if oldDeviceSubscriber != nil {
		if err = oldDeviceSubscriber.Close(); err != nil {
			client.Errorf("failed to close replaced device subscriber: %v", err)
		}
	}
	h := grpcgwClient.NewDeviceSubscriptionHandlers(client)
	deviceSubscriber.SubscribeToPendingCommands(h)
	return nil
}

type updateType int

const (
	updateTypeNone    updateType = 0
	updateTypeNew     updateType = 1
	updateTypeChanged updateType = 2
)

func (client *Client) updateAuthorizationContext(deviceID, userID, accessToken string, validUntil time.Time, jwtClaims jwt.Claims) updateType {
	authCtx := authorizationContext{
		DeviceID:    deviceID,
		UserID:      userID,
		AccessToken: accessToken,
		Expire:      validUntil,
		JWTClaims:   jwtClaims,
	}
	oldAuthCtx := client.SetAuthorizationContext(&authCtx)

	if oldAuthCtx.GetDeviceID() == "" {
		return updateTypeNew
	}
	if oldAuthCtx.GetDeviceID() != deviceID || oldAuthCtx.GetUserID() != userID {
		return updateTypeChanged
	}
	return updateTypeNone
}

func (client *Client) updateBySignInData(ctx context.Context, upd updateType, deviceId, owner string) error {
	if upd == updateTypeChanged {
		client.cancelResourceSubscriptions(true)
		if err := client.closeDeviceSubscriber(); err != nil {
			client.Errorf("failed to close previous device subscription: %w", err)
		}
		if err := client.closeDeviceObserver(client.Context()); err != nil {
			client.Errorf("failed to close previous device observer: %w", err)
		}
		client.unsubscribeFromDeviceEvents()
	}

	if upd != updateTypeNone {
		if err := setNewDeviceSubscriber(client, owner, deviceId); err != nil {
			return fmt.Errorf("cannot set device subscriber: %w", err)
		}
	}

	if err := client.server.devicesStatusUpdater.Add(client); err != nil {
		return fmt.Errorf("cannot update cloud device status: %w", err)
	}

	return nil
}

func subscribeToDeviceEvents(ctx context.Context, client *Client, owner, deviceID string) error {
	if err := client.subscribeToDeviceEvents(owner, func(e *events.Event) {
		evt := e.GetDevicesUnregistered()
		if evt == nil {
			return
		}
		if evt.Owner != owner {
			return
		}
		if !strings.Contains(evt.DeviceIds, deviceID) {
			return
		}
		if err := client.Close(); err != nil {
			client.Errorf("sign in error: cannot close client: %w", err)
		}
	}); err != nil {
		return fmt.Errorf("cannot subscribe to device events: %w", err)
	}
	return nil
}

func subscribeAndValidateDeviceAccess(ctx context.Context, client *Client, owner, deviceID string, subscribe bool) (bool, error) {
	// subscribe to updates before checking cache, so when the device gets removed during sign in
	// the client will always be closed
	if subscribe {
		if err := subscribeToDeviceEvents(ctx, client, owner, deviceID); err != nil {
			return false, err
		}
	}

	return client.server.ownerCache.OwnsDevice(ctx, deviceID)
}

func signInError(err error) error {
	return fmt.Errorf("sign in error: %w", err)
}

func setNewDeviceObserver(ctx context.Context, client *Client, deviceID string, resetObservationType bool) {
	newDeviceObserverFut, setDeviceObserver := future.New()
	oldDeviceObserverFut := client.replaceDeviceObserver(newDeviceObserverFut)

	createDeviceObserver := func() {
		observationType := observation.ObservationType_Detect
		oldDeviceObserver, err := toDeviceObserver(ctx, oldDeviceObserverFut)
		if err != nil {
			client.Errorf("failed to get replaced device observer: %w", err)
		}
		if err == nil && oldDeviceObserver != nil {
			// if the device didn't change we can skip detection and force the previous observation type
			if !resetObservationType {
				observationType = oldDeviceObserver.GetObservationType()
			}
			oldDeviceObserver.Clean(ctx)
		}

		deviceObserver, err := observation.NewDeviceObserver(client.Context(), deviceID, client, client,
			observation.MakeResourcesObserverCallbacks(client.onObserveResource, client.onGetResourceContent),
			observation.WithObservationType(observationType),
			observation.WithLogger(client.getLogger()))
		if err != nil {
			client.Errorf("%w", signInError(fmt.Errorf("cannot create observer for device %v: %w", deviceID, err)))
			setDeviceObserver(nil, err)
			return
		}
		setDeviceObserver(deviceObserver, nil)
	}

	if err := client.server.taskQueue.Submit(createDeviceObserver); err != nil {
		client.Errorf("%w", signInError(fmt.Errorf("failed to register resource observations for device %v: %w", deviceID, err)))
		setDeviceObserver(nil, err)
	}
}

// https://github.com/openconnectivityfoundation/security/blob/master/swagger2.0/oic.sec.session.swagger.json
func signInPostHandler(req *mux.Message, client *Client, signIn CoapSignInReq) {
	logErrorAndCloseClient := func(err error, code coapCodes.Code) {
		client.logAndWriteErrorResponse(req, fmt.Errorf("cannot handle sign in: %w", err), code, req.Token)
		if err := client.Close(); err != nil {
			client.Errorf("%w", signInError(err))
		}
	}

	if err := signIn.checkOAuthRequest(); err != nil {
		logErrorAndCloseClient(err, coapCodes.BadRequest)
		return
	}

	jwtClaims, err := client.ValidateToken(req.Context, signIn.AccessToken)
	if err != nil {
		logErrorAndCloseClient(err, coapCodes.InternalServerError)
		return
	}

	err = client.server.VerifyDeviceID(client.tlsDeviceID, jwtClaims)
	if err != nil {
		logErrorAndCloseClient(err, coapCodes.Unauthorized)
		return
	}

	if err := jwtClaims.ValidateOwnerClaim(client.server.config.APIs.COAP.Authorization.OwnerClaim, signIn.UserID); err != nil {
		logErrorAndCloseClient(err, coapCodes.InternalServerError)
		return
	}

	validUntil, err := jwtClaims.ExpiresAt()
	if err != nil {
		logErrorAndCloseClient(err, coapCodes.InternalServerError)
		return
	}
	deviceID := client.ResolveDeviceID(jwtClaims, signIn.DeviceID)

	upd := client.updateAuthorizationContext(deviceID, signIn.UserID, signIn.AccessToken, validUntil, jwtClaims)

	ctx := kitNetGrpc.CtxWithToken(kitNetGrpc.CtxWithIncomingToken(req.Context, signIn.AccessToken), signIn.AccessToken)
	valid, err := subscribeAndValidateDeviceAccess(ctx, client, signIn.UserID, deviceID, upd != updateTypeNone)
	if err != nil {
		logErrorAndCloseClient(err, coapCodes.InternalServerError)
		return
	}
	if !valid {
		logErrorAndCloseClient(fmt.Errorf("access to device('%s') denied", deviceID), coapCodes.Unauthorized)
		return
	}

	expiresIn := validUntilToExpiresIn(validUntil)
	accept, out, err := getSignInContent(expiresIn, req.Options)
	if err != nil {
		logErrorAndCloseClient(err, coapCodes.InternalServerError)
		return
	}

	if err := client.updateBySignInData(ctx, upd, deviceID, signIn.UserID); err != nil {
		logErrorAndCloseClient(err, coapCodes.InternalServerError)
		return
	}

	if validUntil.IsZero() {
		client.server.expirationClientCache.Delete(deviceID)
	} else {
		setExpirationClientCache(client.server.expirationClientCache, deviceID, client, time.Now().Add(time.Second*time.Duration(expiresIn)))
	}

	client.exchangeCache.Clear()
	client.refreshCache.Clear()

	client.sendResponse(req, coapCodes.Changed, req.Token, accept, out)

	// try to register observations to the device at the cloud.
	setNewDeviceObserver(ctx, client, deviceID, upd == updateTypeChanged)
}

func updateDeviceMetadata(req *mux.Message, client *Client) error {
	oldAuthCtx := client.CleanUp(true)
	if oldAuthCtx.GetDeviceID() != "" {
		ctx := kitNetGrpc.CtxWithToken(req.Context, oldAuthCtx.GetAccessToken())
		client.server.expirationClientCache.Delete(oldAuthCtx.GetDeviceID())

		_, err := client.server.raClient.UpdateDeviceMetadata(ctx, &commands.UpdateDeviceMetadataRequest{
			DeviceId: oldAuthCtx.GetDeviceID(),
			Update: &commands.UpdateDeviceMetadataRequest_Status{
				Status: &commands.ConnectionStatus{
					Value: commands.ConnectionStatus_OFFLINE,
				},
			},
			CommandMetadata: &commands.CommandMetadata{
				Sequence:     client.coapConn.Sequence(),
				ConnectionId: client.remoteAddrString(),
			},
		})
		if err != nil {
			// Device will be still reported as online and it can fix his state by next calls online, offline commands.
			return fmt.Errorf("cannot update cloud device status: %w", err)
		}
	}
	return nil
}

func signOutPostHandler(req *mux.Message, client *Client, signOut CoapSignInReq) {
	logErrorAndCloseClient := func(err error, code coapCodes.Code) {
		client.logAndWriteErrorResponse(req, fmt.Errorf("cannot handle sign out: %w", err), code, req.Token)
		if err := client.Close(); err != nil {
			client.Errorf("sign out error: %w", err)
		}
	}

	if signOut.DeviceID == "" || signOut.UserID == "" || signOut.AccessToken == "" {
		authCurrentCtx, err := client.GetAuthorizationContext()
		if err != nil {
			logErrorAndCloseClient(err, coapCodes.InternalServerError)
			return
		}
		signOut = signOut.updateOAUthRequestIfEmpty(authCurrentCtx.DeviceID, authCurrentCtx.UserID, authCurrentCtx.AccessToken)
	}

	if err := signOut.checkOAuthRequest(); err != nil {
		logErrorAndCloseClient(err, coapCodes.BadRequest)
		return
	}

	jwtClaims, err := client.ValidateToken(req.Context, signOut.AccessToken)
	if err != nil {
		logErrorAndCloseClient(err, coapCodes.InternalServerError)
		return
	}

	err = client.server.VerifyDeviceID(client.tlsDeviceID, jwtClaims)
	if err != nil {
		logErrorAndCloseClient(err, coapCodes.Unauthorized)
		return
	}

	if err := jwtClaims.ValidateOwnerClaim(client.server.config.APIs.COAP.Authorization.OwnerClaim, signOut.UserID); err != nil {
		logErrorAndCloseClient(err, coapCodes.InternalServerError)
		return
	}

	if err := updateDeviceMetadata(req, client); err != nil {
		logErrorAndCloseClient(err, coapconv.GrpcErr2CoapCode(err, coapconv.Update))
		return
	}

	client.sendResponse(req, coapCodes.Changed, req.Token, message.AppOcfCbor, []byte{0xA0}) // empty object
}

// Sign-in
// https://github.com/openconnectivityfoundation/security/blob/master/swagger2.0/oic.sec.session.swagger.json
func signInHandler(req *mux.Message, client *Client) {
	switch req.Code {
	case coapCodes.POST:
		var signIn CoapSignInReq
		err := cbor.ReadFrom(req.Body, &signIn)
		if err != nil {
			client.logAndWriteErrorResponse(req, fmt.Errorf("cannot handle sign in: %w", err), coapCodes.BadRequest, req.Token)
			return
		}
		switch signIn.Login {
		case true:
			signInPostHandler(req, client, signIn)
		default:
			signOutPostHandler(req, client, signIn)
		}
	default:
		client.logAndWriteErrorResponse(req, fmt.Errorf("forbidden request from %v", client.remoteAddrString()), coapCodes.Forbidden, req.Token)
	}
}
