package testhttp

import (
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	shttp "github.com/mattermost/mattermost-plugin-msoffice/server/http"
	"github.com/mattermost/mattermost-plugin-msoffice/server/user"
	"github.com/mattermost/mattermost-plugin-msoffice/server/user/mock_user"
	"github.com/stretchr/testify/assert"
)

func TestOAuth2Connect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tcs := []struct {
		name                 string
		handler              shttp.Handler
		r                    *http.Request
		w                    *mockResponseWriter
		expectedHTTPResponse string
		expectedHTTPCode     int
	}{
		{
			name: "unauthorized user",
			r:    &http.Request{},
			w:    defaultMockResponseWriter(),
			handler: shttp.Handler{
				Config:           &config.Config{},
				UserStore:        user.NewStore(getMockKVStore(ctrl, &mockKVStoreConfig{})),
				OAuth2StateStore: mock_user.NewMockOAuth2StateStore(ctrl),
			},
			expectedHTTPResponse: "Not authorized\n",
			expectedHTTPCode:     http.StatusUnauthorized,
		},
		{
			name: "unable to store user state",
			r:    getUserRequest("fake@mattermost.com", ""),
			w:    defaultMockResponseWriter(),
			handler: shttp.Handler{
				Config:    &config.Config{},
				UserStore: user.NewStore(getMockKVStore(ctrl, &mockKVStoreConfig{})),
				OAuth2StateStore: getMockOAuth2StateStore(ctrl, &mockOAuth2StateStoreConfig{
					useAnyStoreKey: true,
					storeErr:       errors.New("unable to store state"),
					maxTimesStore:  1,
				}),
			},
			expectedHTTPResponse: "{\"error\":\"An internal error has occurred. Check app server logs for details.\",\"details\":\"unable to store state\"}",
			expectedHTTPCode:     http.StatusInternalServerError,
		},
		{
			name: "successful redirect",
			r:    getUserRequest("fake@mattermost.com", ""),
			w:    defaultMockResponseWriter(),
			handler: shttp.Handler{
				Config:    &config.Config{},
				UserStore: user.NewStore(getMockKVStore(ctrl, &mockKVStoreConfig{})),
				OAuth2StateStore: getMockOAuth2StateStore(ctrl, &mockOAuth2StateStoreConfig{
					useAnyStoreKey: true,
					storeErr:       nil,
					maxTimesStore:  1,
				}),
			},
			expectedHTTPResponse: "",
			expectedHTTPCode:     http.StatusFound,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.handler.OAuth2Connect(tc.w, tc.r)

			assert.Equal(t, tc.expectedHTTPCode, tc.w.StatusCode)
			assert.Equal(t, tc.expectedHTTPResponse, string(tc.w.Bytes))
		})
	}
}
