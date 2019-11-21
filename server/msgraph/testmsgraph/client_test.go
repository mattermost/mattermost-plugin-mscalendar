package testmsgraph

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/msgraph"
	"golang.org/x/oauth2"
)

const (
	tokenURLEndpoint = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"
)

func testConfig() *config.Config {
	conf := &config.Config{}

	conf.OAuth2Authority = "common"
	conf.OAuth2ClientId = "fakeclientid"
	conf.OAuth2ClientSecret = "fakeclientsecret"
	conf.PluginURL = "http://localhost"

	return conf
}

func getToken(expiry time.Time) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  "fake_access_token",
		TokenType:    "bearer",
		RefreshToken: "fake_refresh_token",
		Expiry:       expiry,
	}
}

func TestNewClient(t *testing.T) {
	conf := testConfig()

	token := getToken(time.Now().Add(time.Hour))

	client := msgraph.NewClient(conf, token)

	if client == nil {
		t.Errorf("expected client to be non-nil but got: %+v", client)
	}
}
