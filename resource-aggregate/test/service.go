package test

import (
	"github.com/plgd-dev/kit/config"
	"sync"
	"testing"
	"time"

	"github.com/plgd-dev/cloud/resource-aggregate/refImpl"
	"github.com/plgd-dev/cloud/resource-aggregate/service"
	testCfg "github.com/plgd-dev/cloud/test/config"
	"github.com/stretchr/testify/require"
)

func MakeConfig(t *testing.T) service.Config {
	var raCfg service.Config
	err := config.Load(&raCfg)
	require.NoError(t, err)
	raCfg.Service.RA.GrpcAddr = testCfg.RESOURCE_AGGREGATE_HOST
	raCfg.Clients.AuthServer.AuthServerAddr = testCfg.AUTH_HOST
	raCfg.Clients.OAuthProvider.JwksURL = testCfg.JWKS_URL
	raCfg.Clients.OAuthProvider.OAuthConfig.ClientID = testCfg.OAUTH_MANAGER_CLIENT_ID
	raCfg.Clients.OAuthProvider.OAuthConfig.TokenURL = testCfg.OAUTH_MANAGER_ENDPOINT_TOKENURL
	raCfg.Service.RA.Capabilities.UserDevicesManagerTickFrequency = time.Millisecond * 500
	raCfg.Service.RA.Capabilities.UserDevicesManagerExpiration = time.Millisecond * 500
	return raCfg
}

func SetUp(t *testing.T) (TearDown func()) {
	return New(t, MakeConfig(t))
}

func New(t *testing.T, cfg service.Config) func() {

	s, err := refImpl.Init(cfg)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.Serve()
		require.NoError(t, err)
	}()

	return func() {
		s.Shutdown()
		wg.Wait()
	}
}
