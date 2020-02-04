package http

import (
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/api"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/api/mock_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/msgraph"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
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
		setup                func(dependencies *api.Dependencies)
		expectedHTTPResponse string
		expectedHTTPCode     int
	}{
		{
			name:                 "unauthorized user",
			expectedHTTPResponse: "Not authorized\n",
			expectedHTTPCode:     http.StatusUnauthorized,
		},
		{
			name:             "MM user already connected",
			mattermostUserID: "fake@mattermost.com",
			setup: func(d *api.Dependencies) {
				su := &store.User{Remote: &remote.User{Mail: "remote_email@example.com"}}
				us := d.UserStore.(*mock_store.MockUserStore)
				us.EXPECT().LoadUser("fake@mattermost.com").Return(su, nil).Times(1)
			},
			expectedHTTPResponse: "User is already connected to remote_email@example.com\n",
			expectedHTTPCode:     http.StatusInternalServerError,
		},
		{
			name:             "unable to store user state",
			mattermostUserID: "fake@mattermost.com",
			setup: func(d *api.Dependencies) {
				us := d.UserStore.(*mock_store.MockUserStore)
				us.EXPECT().LoadUser("fake@mattermost.com").Return(nil, errors.New("Remote user not found")).Times(1)

				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().StoreOAuth2State(gomock.Any()).Return(errors.New("unable to store state")).Times(1)
			},
			expectedHTTPCode:     http.StatusInternalServerError,
			expectedHTTPResponse: "unable to store state\n",
		},
		{
			name:             "successful redirect",
			mattermostUserID: "fake@mattermost.com",
			setup: func(d *api.Dependencies) {
				us := d.UserStore.(*mock_store.MockUserStore)
				us.EXPECT().LoadUser("fake@mattermost.com").Return(nil, errors.New("Remote user not found")).Times(1)

				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().StoreOAuth2State(gomock.Any()).Return(nil).Times(1)
			},
			expectedHTTPCode:     http.StatusFound,
			expectedHTTPResponse: "",
		},
		{
			name:             "Connecting bot, user is not admin",
			mattermostUserID: "fake@mattermost.com",
			queryStr:         "bot=true",
			setup: func(d *api.Dependencies) {
				d.IsAuthorizedAdmin = func(userID string) (bool, error) {
					return false, nil
				}
			},
			expectedHTTPResponse: "Not authorized\n",
			expectedHTTPCode:     http.StatusUnauthorized,
		},
		{
			name:             "Connecting bot, user is admin",
			mattermostUserID: "fake@mattermost.com",
			queryStr:         "bot=true",
			setup: func(d *api.Dependencies) {
				d.IsAuthorizedAdmin = func(userID string) (bool, error) {
					return true, nil
				}

				us := d.UserStore.(*mock_store.MockUserStore)
				us.EXPECT().LoadUser("bot_user_id").Return(nil, errors.New("Remote user not found")).Times(1)

				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().StoreOAuth2State(gomock.Any()).Return(nil).Times(1)
			},
			expectedHTTPResponse: "",
			expectedHTTPCode:     http.StatusFound,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			conf := &config.Config{
				StoredConfig: config.StoredConfig{
					OAuth2Authority:    "common",
					OAuth2ClientID:     "fakeclientid",
					OAuth2ClientSecret: "fakeclientsecret",
				},
				PluginURL: "http://localhost",
				BotUserID: "bot_user_id",
			}

			handler := NewHandler(conf)

			dependencies := mock_api.NewMockDependencies(ctrl)
			dependencies.Remote = remote.Makers[msgraph.Kind](conf, &bot.NilLogger{})
			if tc.setup != nil {
				tc.setup(dependencies)
			}

			apiconf := api.Config{
				Config:       conf,
				Dependencies: dependencies,
			}
			r := newHTTPRequest(apiconf, tc.mattermostUserID, tc.queryStr)
			w := defaultMockResponseWriter()

			handler.oauth2Connect(w, r)

			assert.Equal(t, tc.expectedHTTPCode, w.StatusCode)
			assert.Equal(t, tc.expectedHTTPResponse, string(w.Bytes))
		})
	}
}
