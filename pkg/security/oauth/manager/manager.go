package manager

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/plgd-dev/cloud/pkg/log"
	"github.com/plgd-dev/cloud/pkg/net/http/client"
	"golang.org/x/oauth2"
)

// Manager holds certificates from filesystem watched for changes
type Manager struct {
	mutex                       sync.Mutex
	config                      clientcredentials.Config
	requestTimeout              time.Duration
	verifyServiceTokenFrequency time.Duration
	startRefreshToken           time.Time
	token                       *oauth2.Token
	httpClient                  *http.Client
	tokenErr                    error
	doneWg                      sync.WaitGroup
	done                        chan struct{}

	http *client.Client
}

// NewManagerFromConfiguration creates a new oauth manager which refreshing token.
func NewManagerFromConfiguration(config Config, tlsCfg *tls.Config) (*Manager, error) {
	cfg := config.ToClientCrendtials()
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 1
	t.IdleConnTimeout = time.Second * 30
	t.TLSClientConfig = tlsCfg
	m, err := new(cfg, &http.Client{
		Transport: t,
		Timeout:   config.RequestTimeout,
	}, config.RequestTimeout, config.TickFrequency)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func new(cfg clientcredentials.Config, httpClient *http.Client, requestTimeout, verifyServiceTokenFrequency time.Duration) (*Manager, error) {
	token, startRefreshToken, err := getToken(cfg, httpClient, requestTimeout)
	if err != nil {
		return nil, err
	}

	mgr := &Manager{
		config:                      cfg,
		token:                       token,
		startRefreshToken:           startRefreshToken,
		requestTimeout:              requestTimeout,
		httpClient:                  httpClient,
		verifyServiceTokenFrequency: verifyServiceTokenFrequency,

		done: make(chan struct{}),
	}
	mgr.doneWg.Add(1)

	go mgr.watchToken()

	return mgr, nil
}

func New(config ConfigV2, logger *zap.Logger) (*Manager, error) {
	http, err := client.New(config.HTTP, logger)
	if err != nil {
		return nil, fmt.Errorf("cannot create http client: %w", err)
	}
	m, err := new(config.ToClientCrendtials(), http.HTTP(), config.HTTP.Timeout, config.VerifyServiceTokenFrequency)
	if err != nil {
		return nil, err
	}
	m.http = http
	return m, nil
}

// GetToken returns token for clients
func (a *Manager) GetToken(ctx context.Context) (*oauth2.Token, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.token, a.tokenErr
}

// Close ends watching token
func (a *Manager) Close() {
	if a.done != nil {
		close(a.done)
		a.doneWg.Wait()
		if a.http != nil {
			a.http.Close()
		}
	}
}

func (a *Manager) shouldRefresh() bool {
	return time.Now().After(a.startRefreshToken)
}

func getToken(cfg clientcredentials.Config, httpClient *http.Client, requestTimeout time.Duration) (*oauth2.Token, time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	token, err := cfg.Token(ctx)
	var startRefreshToken time.Time
	if err == nil {
		now := time.Now()
		startRefreshToken = now.Add(token.Expiry.Sub(now) * 2 / 3)
	}
	return token, startRefreshToken, err
}

func (a *Manager) refreshToken() {
	token, startRefreshToken, err := getToken(a.config, a.httpClient, a.requestTimeout)
	if err != nil {
		log.Errorf("cannot refresh token: %v", err)
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.token = token
	a.tokenErr = err
	a.startRefreshToken = startRefreshToken
}

func (a *Manager) watchToken() {
	defer a.doneWg.Done()
	t := time.NewTicker(a.verifyServiceTokenFrequency)
	defer t.Stop()

	for {
		select {
		case <-a.done:
			return
		case <-t.C:
			if a.shouldRefresh() {
				a.refreshToken()
			}
		}
	}
}