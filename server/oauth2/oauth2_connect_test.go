package oauth2

import (
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/msgraph"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store/mock_store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
)

func TestOAuth2Connect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tcs := []struct {
		name                 string
		mattermostUserID     string
		queryStr             string
		setup                func(dependencies *mscalendar.Dependencies)
		expectedHTTPResponse string
		expectedHTTPCode     int
	}{
		{
			name:                 "unauthorized user",
			expectedHTTPResponse: "Not authorized\n",
			expectedHTTPCode:     http.StatusUnauthorized,
		},
		{
			name:             "unable to store user state",
			mattermostUserID: "fake@mattermost.com",
			setup: func(d *mscalendar.Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().StoreOAuth2State(gomock.Any()).Return(errors.New("unable to store state")).Times(1)
			},
			expectedHTTPCode:     http.StatusInternalServerError,
			expectedHTTPResponse: "unable to store state\n",
		},
		{
			name:             "successful redirect",
			mattermostUserID: "fake@mattermost.com",
			setup: func(d *mscalendar.Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().StoreOAuth2State(gomock.Any()).Return(nil).Times(1)
			},
			expectedHTTPCode:     http.StatusFound,
			expectedHTTPResponse: "",
		},
	}

	router := &mux.Router{}
	RegisterHTTP(router)

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			conf := &config.Config{
				StoredConfig: config.StoredConfig{
					OAuth2Authority:    "common",
					OAuth2ClientID:     "fakeclientid",
					OAuth2ClientSecret: "fakeclientsecret",
				},
				PluginURL: "http://localhost",
			}

			dependencies := NewMockDependencies(ctrl)
			dependencies.Remote = remote.Makers[msgraph.Kind](conf, &bot.NilLogger{})
			if tc.setup != nil {
				tc.setup(dependencies)
			}

			apiconf := mscalendar.Config{
				Config:       conf,
				Dependencies: dependencies,
			}
			r := newHTTPRequest(apiconf, tc.mattermostUserID, tc.queryStr)
			w := defaultMockResponseWriter()

			oauth2Connect(w, r)

			assert.Equal(t, tc.expectedHTTPCode, w.StatusCode)
			assert.Equal(t, tc.expectedHTTPResponse, string(w.Bytes))
		})
	}
}
