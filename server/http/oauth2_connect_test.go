package http

import (
	"errors"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/mocks"
	"github.com/mattermost/mattermost-plugin-msoffice/server/user"
	"github.com/stretchr/testify/assert"
)

func TestOAuth2Connect(t *testing.T) {
	kv := &mocks.MockKVStore{
		KVStore: make(map[string][]byte),
		Err:     nil,
	}

	tcs := []struct {
		name                 string
		handler              Handler
		r                    *http.Request
		w                    *mocks.MockResponseWriter
		expectedHTTPResponse string
		expectedHTTPCode     int
	}{
		{
			name: "unauthorized user",
			r:    &http.Request{},
			w:    mocks.DefaultMockResponseWriter(),
			handler: Handler{
				Config:           &config.Config{},
				UserStore:        user.NewStore(kv),
				OAuth2StateStore: &mocks.MockOAuth2StateStore{},
			},
			expectedHTTPResponse: "Not authorized\n",
			expectedHTTPCode:     http.StatusUnauthorized,
		},
		{
			name: "unable to store user state",
			r: &http.Request{
				Header: http.Header{textproto.CanonicalMIMEHeaderKey("mattermost-user-id"): []string{"fake@mattermost.com"}},
			},
			w: mocks.DefaultMockResponseWriter(),
			handler: Handler{
				Config:           &config.Config{},
				UserStore:        user.NewStore(kv),
				OAuth2StateStore: &mocks.MockOAuth2StateStore{Err: errors.New("unable to store state")},
			},
			expectedHTTPResponse: "{\"error\":\"An internal error has occurred. Check app server logs for details.\",\"details\":\"unable to store state\"}",
			expectedHTTPCode:     http.StatusInternalServerError,
		},
		{
			name: "successful redirect",
			r: &http.Request{
				Header: http.Header{textproto.CanonicalMIMEHeaderKey("mattermost-user-id"): []string{"fake@mattermost.com"}},
			},
			w: mocks.DefaultMockResponseWriter(),
			handler: Handler{
				Config:           &config.Config{},
				UserStore:        user.NewStore(kv),
				OAuth2StateStore: &mocks.MockOAuth2StateStore{},
			},
			expectedHTTPResponse: "",
			expectedHTTPCode:     http.StatusFound,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.handler.oauth2Connect(tc.w, tc.r)

			assert.Equal(t, tc.expectedHTTPCode, tc.w.StatusCode)
			assert.Equal(t, tc.expectedHTTPResponse, string(tc.w.Bytes))
		})
	}
}
