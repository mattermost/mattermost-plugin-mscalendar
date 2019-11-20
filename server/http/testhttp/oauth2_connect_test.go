package testhttp

import (
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	shttp "github.com/mattermost/mattermost-plugin-msoffice/server/http"
	"github.com/mattermost/mattermost-plugin-msoffice/server/kvstore/mock_kvstore"
	"github.com/mattermost/mattermost-plugin-msoffice/server/user"
	"github.com/mattermost/mattermost-plugin-msoffice/server/user/mock_user"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/mock_utils"
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
		setupMocks           func(*mock_kvstore.MockKVStore, *mock_user.MockOAuth2StateStore, *mock_utils.MockBotPoster)
		expectedHTTPResponse string
		expectedHTTPCode     int
	}{
		{
			name: "unauthorized user",
			r:    &http.Request{},
			w:    defaultMockResponseWriter(),
			handler: shttp.Handler{
				Config: &config.Config{},
			},
			setupMocks:           func(kv *mock_kvstore.MockKVStore, ss *mock_user.MockOAuth2StateStore, bp *mock_utils.MockBotPoster) {},
			expectedHTTPResponse: "Not authorized\n",
			expectedHTTPCode:     http.StatusUnauthorized,
		},
		{
			name: "unable to store user state",
			r:    makeUserRequest("fake@mattermost.com", ""),
			w:    defaultMockResponseWriter(),
			handler: shttp.Handler{
				Config: &config.Config{},
			},
			setupMocks: func(kv *mock_kvstore.MockKVStore, ss *mock_user.MockOAuth2StateStore, bp *mock_utils.MockBotPoster) {
				ss.EXPECT().Store(gomock.Any()).Return(errors.New("unable to store state")).Times(1)
			},
			expectedHTTPResponse: "{\"error\":\"An internal error has occurred. Check app server logs for details.\",\"details\":\"unable to store state\"}",
			expectedHTTPCode:     http.StatusInternalServerError,
		},
		{
			name: "successful redirect",
			r:    makeUserRequest("fake@mattermost.com", ""),
			w:    defaultMockResponseWriter(),
			handler: shttp.Handler{
				Config: &config.Config{},
			},
			setupMocks: func(kv *mock_kvstore.MockKVStore, ss *mock_user.MockOAuth2StateStore, bp *mock_utils.MockBotPoster) {
				ss.EXPECT().Store(gomock.Any()).Return(nil).Times(1)
			},
			expectedHTTPResponse: "",
			expectedHTTPCode:     http.StatusFound,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			mockKVStore := mock_kvstore.NewMockKVStore(ctrl)
			mockOAuth2StateStore := mock_user.NewMockOAuth2StateStore(ctrl)
			mockBotPoster := mock_utils.NewMockBotPoster(ctrl)

			tc.setupMocks(mockKVStore, mockOAuth2StateStore, mockBotPoster)

			tc.handler.UserStore = user.NewStore(mockKVStore)
			tc.handler.OAuth2StateStore = mockOAuth2StateStore

			tc.handler.OAuth2Connect(tc.w, tc.r)

			assert.Equal(t, tc.expectedHTTPCode, tc.w.StatusCode)
			assert.Equal(t, tc.expectedHTTPResponse, string(tc.w.Bytes))
		})
	}
}
