package oauth

import "golang.org/x/oauth2"

type AuthStyle string

const (
	// AuthStyleAutoDetect means to auto-detect which authentication
	// style the provider wants by trying both ways and caching
	// the successful way for the future.
	AuthStyleAutoDetect AuthStyle = "AutoDetect"

	// AuthStyleInParams sends the "client_id" and "client_secret"
	// in the POST body as application/x-www-form-urlencoded parameters.
	AuthStyleInParams AuthStyle = "InParams"

	// AuthStyleInHeader sends the client_id and client_password
	// using HTTP Basic Authorization. This is an optional style
	// described in the OAuth2 RFC 6749 section 2.3.1.
	AuthStyleInHeader AuthStyle = "InHeader"
)

func (a AuthStyle) ToOAuth2() oauth2.AuthStyle {
	switch a {
	case AuthStyleInHeader:
		return oauth2.AuthStyleInHeader
	case AuthStyleInParams:
		return oauth2.AuthStyleInParams
	}
	return oauth2.AuthStyleAutoDetect
}

type Endpoint struct {
	AuthURL   	string    `yaml:"auth-url" json:"auth-url"`
	TokenURL  	string    `yaml:"token-url" json:"token-url"`
	AuthStyle 	AuthStyle `yaml:"auth-style" json:"auth-style"`
}

type Config struct {
	ClientID     string   `yaml:"client-id" json:"client-id"`
	ClientSecret string   `yaml:"client-secret" json:"client-secret"`
	Scopes       []string `yaml:"scopes" json:"scopes"`
	Endpoint     Endpoint `yaml:"endpoint" json:"endpoint"`
	Audience     string   `yaml:"audience" json:"audience"`
	RedirectURL  string   `yaml:"redirect-url" json:"redirect-url"`
	AccessType   string   `yaml:"access-type" json:"access-type"`
	ResponseType string   `yaml:"response-type" json:"response-type"`
	ResponseMode string   `yaml:"response-mode" json:"response-mode"`
}

func (c Config) ToOAuth2() oauth2.Config {
	return oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURL,
		Scopes:       c.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   c.Endpoint.AuthURL,
			TokenURL:  c.Endpoint.TokenURL,
			AuthStyle: c.Endpoint.AuthStyle.ToOAuth2(),
		},
	}
}

func (c Config) AuthCodeURL(csrfToken string) string {
	aud := c.Audience
	accessType := c.AccessType
	responseType := c.ResponseType
	responseMode := c.ResponseMode

	opts := make([]oauth2.AuthCodeOption, 0, 2)

	if accessType != "" {
		opts = append(opts, oauth2.SetAuthURLParam("access_type", accessType))
	}
	if aud != "" {
		opts = append(opts, oauth2.SetAuthURLParam("audience", aud))
	}
	if responseType != "" {
		opts = append(opts, oauth2.SetAuthURLParam("response_type", responseType))
	}
	if responseMode != "" {
		opts = append(opts, oauth2.SetAuthURLParam("response_mode", responseMode))
	}
	auth := c.ToOAuth2()
	return (&auth).AuthCodeURL(csrfToken, opts...)
}
