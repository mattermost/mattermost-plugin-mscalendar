// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package msgraph

import (
	"context"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/microsoft"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot"
)

func TestLoginEndpoint(t *testing.T) {
	for _, tc := range []struct {
		name       string
		tenant     string
		tenantType string
		authURL    string
		tokenURL   string
	}{
		{
			name:       "empty tenant type defaults to commercial",
			tenant:     "tenant-id",
			tenantType: "",
			authURL:    "https://login.microsoftonline.com/tenant-id/oauth2/v2.0/authorize",
			tokenURL:   "https://login.microsoftonline.com/tenant-id/oauth2/v2.0/token",
		},
		{
			name:       "commercial",
			tenant:     "tenant-id",
			tenantType: TenantTypeCommercial,
			authURL:    "https://login.microsoftonline.com/tenant-id/oauth2/v2.0/authorize",
			tokenURL:   "https://login.microsoftonline.com/tenant-id/oauth2/v2.0/token",
		},
		{
			name:       "GCC High",
			tenant:     "tenant-id",
			tenantType: TenantTypeGCCHigh,
			authURL:    "https://login.microsoftonline.us/tenant-id/oauth2/v2.0/authorize",
			tokenURL:   "https://login.microsoftonline.us/tenant-id/oauth2/v2.0/token",
		},
		{
			name:       "DoD",
			tenant:     "tenant-id",
			tenantType: TenantTypeDoD,
			authURL:    "https://login.microsoftonline.us/tenant-id/oauth2/v2.0/authorize",
			tokenURL:   "https://login.microsoftonline.us/tenant-id/oauth2/v2.0/token",
		},
		{
			name:       "empty tenant falls back to common",
			tenant:     "",
			tenantType: TenantTypeGCCHigh,
			authURL:    "https://login.microsoftonline.us/common/oauth2/v2.0/authorize",
			tokenURL:   "https://login.microsoftonline.us/common/oauth2/v2.0/token",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			endpoint := loginEndpoint(tc.tenant, tc.tenantType)
			assert.Equal(t, tc.authURL, endpoint.AuthURL)
			assert.Equal(t, tc.tokenURL, endpoint.TokenURL)
		})
	}
}

func TestLoginEndpointCommercialMatchesAzureADEndpoint(t *testing.T) {
	// Guards backwards compatibility: the commercial cloud must behave exactly like the
	// previously used golang.org/x/oauth2/microsoft endpoint.
	for _, tenant := range []string{"", "tenant-id"} {
		assert.Equal(t, microsoft.AzureADEndpoint(tenant), loginEndpoint(tenant, ""))
		assert.Equal(t, microsoft.AzureADEndpoint(tenant), loginEndpoint(tenant, TenantTypeCommercial))
	}
}

func TestGraphEndpoints(t *testing.T) {
	for _, tc := range []struct {
		tenantType string
		baseURL    string
		scope      string
		loginHost  string
	}{
		{"", "https://graph.microsoft.com/v1.0", "https://graph.microsoft.com/.default", "https://login.microsoftonline.com"},
		{TenantTypeCommercial, "https://graph.microsoft.com/v1.0", "https://graph.microsoft.com/.default", "https://login.microsoftonline.com"},
		{TenantTypeGCCHigh, "https://graph.microsoft.us/v1.0", "https://graph.microsoft.us/.default", "https://login.microsoftonline.us"},
		{TenantTypeDoD, "https://dod-graph.microsoft.us/v1.0", "https://dod-graph.microsoft.us/.default", "https://login.microsoftonline.us"},
	} {
		t.Run(tc.tenantType, func(t *testing.T) {
			assert.Equal(t, tc.baseURL, graphBaseURL(tc.tenantType))
			assert.Equal(t, tc.scope, graphDefaultScope(tc.tenantType))
			assert.Equal(t, tc.loginHost, loginHost(tc.tenantType))
		})
	}
}

func TestValidateTenantType(t *testing.T) {
	for _, valid := range []string{"", TenantTypeCommercial, TenantTypeGCCHigh, TenantTypeDoD} {
		assert.NoError(t, validateTenantType(valid), valid)
	}
	assert.Error(t, validateTenantType("china"))
	assert.Error(t, validateTenantType("GCCH"))
}

func TestCheckConfigurationTenantType(t *testing.T) {
	r := NewRemote(&config.Config{}, &bot.NilLogger{})
	stored := config.StoredConfig{OAuth2Authority: "tenant-id", OAuth2ClientID: "client-id", OAuth2ClientSecret: "secret"}

	for _, valid := range []string{"", TenantTypeCommercial, TenantTypeGCCHigh, TenantTypeDoD} {
		stored.OAuth2TenantType = valid
		assert.NoError(t, r.CheckConfiguration(stored), valid)
	}

	stored.OAuth2TenantType = "unknown"
	assert.Error(t, r.CheckConfiguration(stored))
}

func TestSuperuserClientUsesConfiguredCloud(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodPost, "https://login.microsoftonline.us/tenant-id/oauth2/v2.0/token",
		func(req *http.Request) (*http.Response, error) {
			require.NoError(t, req.ParseForm())
			assert.Equal(t, "client_credentials", req.PostForm.Get("grant_type"))
			assert.Equal(t, "https://graph.microsoft.us/.default", req.PostForm.Get("scope"))
			return httpmock.NewJsonResponse(http.StatusOK, map[string]any{
				"token_type":   "Bearer",
				"access_token": "gcch-app-token",
				"expires_in":   3600,
			})
		})
	httpmock.RegisterResponder(http.MethodGet, "https://graph.microsoft.us/v1.0/users/remote-user-id/mailboxSettings",
		func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "Bearer gcch-app-token", req.Header.Get("Authorization"))
			return httpmock.NewJsonResponse(http.StatusOK, map[string]any{"timeZone": "Eastern Standard Time"})
		})

	conf := &config.Config{
		StoredConfig: config.StoredConfig{
			OAuth2TenantType:   TenantTypeGCCHigh,
			OAuth2Authority:    "tenant-id",
			OAuth2ClientID:     "client-id",
			OAuth2ClientSecret: "secret",
		},
		PluginURL: "https://mattermost.example.com/plugins/com.mattermost.mscalendar",
	}

	c, err := NewRemote(conf, &bot.NilLogger{}).MakeSuperuserClient(context.Background())
	require.NoError(t, err)

	settings, err := c.GetMailboxSettings("remote-user-id")
	require.NoError(t, err)
	assert.Equal(t, "Eastern Standard Time", settings.TimeZone)

	info := httpmock.GetCallCountInfo()
	assert.Equal(t, 1, info["POST https://login.microsoftonline.us/tenant-id/oauth2/v2.0/token"])
	assert.Equal(t, 1, info["GET https://graph.microsoft.us/v1.0/users/remote-user-id/mailboxSettings"])
}
