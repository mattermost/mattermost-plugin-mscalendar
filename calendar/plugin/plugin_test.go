// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package plugin

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/httputils"
	"github.com/mattermost/mattermost-plugin-mscalendar/msgraph"
)

// newConfiguredTestPlugin returns a plugin, as main builds it, whose stored configuration is stored.
func newConfiguredTestPlugin(t *testing.T, stored config.StoredConfig) *Plugin {
	t.Helper()
	previousProvider := config.Provider
	config.Provider = msgraph.GetMSCalendarProviderConfig()
	t.Cleanup(func() { config.Provider = previousProvider })

	siteURL := "https://mattermost.example.com"
	mockAPI := &plugintest.API{}
	mockAPI.On("LoadPluginConfiguration", mock.AnythingOfType("*config.StoredConfig")).Run(func(args mock.Arguments) {
		*args.Get(0).(*config.StoredConfig) = stored
	}).Return(nil)
	mockAPI.On("GetConfig").Return(&model.Config{ServiceSettings: model.ServiceSettings{SiteURL: &siteURL}}).Maybe()

	p := NewWithEnv(engine.Env{
		Config:       &config.Config{PluginID: "com.mattermost.mscalendar", Provider: config.Provider},
		Dependencies: &engine.Dependencies{},
	})
	p.SetAPI(mockAPI)
	return p
}

func TestOnConfigurationChangeRejectsInvalidConfiguration(t *testing.T) {
	valid := config.StoredConfig{OAuth2ClientID: "client-id", OAuth2ClientSecret: "client-secret", OAuth2Authority: "tenant-id"}
	for name, tc := range map[string]struct {
		mutate  func(*config.StoredConfig)
		message string
	}{
		"unsupported tenant type": {func(c *config.StoredConfig) { c.OAuth2TenantType = "GCCH" }, "OAuth2TenantType"},
		"missing client secret":   {func(c *config.StoredConfig) { c.OAuth2ClientSecret = "" }, "OAuth2 credentials"},
	} {
		t.Run(name, func(t *testing.T) {
			stored := valid
			tc.mutate(&stored)
			p := newConfiguredTestPlugin(t, stored)

			err := p.OnConfigurationChange()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.message)

			env := p.getEnv()
			assert.Equal(t, err, env.configError, "configError must block command and HTTP handling")
			assert.Nil(t, env.Remote, "an invalid configuration must not be applied")
			assert.Equal(t, config.StoredConfig{}, env.StoredConfig, "an invalid configuration must not be applied")
		})
	}
}

func TestOnActivateAfterInvalidConfigurationReturnsError(t *testing.T) {
	// The server calls OnActivate even when OnConfigurationChange failed, so env.Remote is still unset.
	p := newConfiguredTestPlugin(t, config.StoredConfig{
		OAuth2ClientID: "client-id", OAuth2ClientSecret: "client-secret", OAuth2Authority: "tenant-id", OAuth2TenantType: "GCCH",
	})
	require.Error(t, p.OnConfigurationChange())

	var err error
	require.NotPanics(t, func() { err = p.OnActivate() })
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth2TenantType")
}

func TestServeHTTPRecoversFromPanic(t *testing.T) {
	mockAPI := &plugintest.API{}
	mockAPI.On("LogError", "Recovered from panic while serving HTTP request", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	defer mockAPI.AssertExpectations(t)

	handler := httputils.NewHandler()
	handler.Router.HandleFunc("/panic", func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}).Methods(http.MethodGet)

	p := &Plugin{envLock: &sync.RWMutex{}}
	p.SetAPI(mockAPI)
	p.env = Env{httpHandler: handler}

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		p.ServeHTTP(nil, rec, req)
	})

	assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
}

func TestServeHTTPNoPanicSucceeds(t *testing.T) {
	handler := httputils.NewHandler()
	handler.Router.HandleFunc("/ok", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	p := &Plugin{envLock: &sync.RWMutex{}}
	p.env = Env{httpHandler: handler}

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()

	p.ServeHTTP(nil, rec, req)

	assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
}
