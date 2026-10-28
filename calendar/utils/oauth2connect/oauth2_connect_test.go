// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package oauth2connect

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/httputils"
)

type mockOAuth2App struct {
	initOAuth2Fn     func(userID string) (string, error)
	completeOAuth2Fn func(userID, code, state string) error
}

func (m *mockOAuth2App) InitOAuth2(userID string) (string, error) {
	if m.initOAuth2Fn != nil {
		return m.initOAuth2Fn(userID)
	}
	return "https://login.microsoftonline.com/oauth2/authorize?client_id=test", nil
}

func (m *mockOAuth2App) CompleteOAuth2(userID, code, state string) error {
	if m.completeOAuth2Fn != nil {
		return m.completeOAuth2Fn(userID, code, state)
	}
	return nil
}

func setupOAuth2Router(app App) http.Handler {
	handler := httputils.NewHandler()
	providerConfig := config.ProviderConfig{
		DisplayName:    "Microsoft Calendar",
		CommandTrigger: "mscalendar",
	}
	Init(handler, app, providerConfig)
	return handler
}

func TestOAuth2Connect_Unauthorized(t *testing.T) {
	app := &mockOAuth2App{}
	router := setupOAuth2Router(app)

	t.Run("GET without user id returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oauth2/connect", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("POST without user id returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/oauth2/connect", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestOAuth2Connect_InitError(t *testing.T) {
	app := &mockOAuth2App{
		initOAuth2Fn: func(userID string) (string, error) {
			return "", errors.New("failed to init oauth2")
		},
	}
	router := setupOAuth2Router(app)

	t.Run("GET with init failure returns 500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oauth2/connect", nil)
		req.Header.Set("Mattermost-User-ID", "user_123")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("POST with init failure returns 500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/oauth2/connect", nil)
		req.Header.Set("Mattermost-User-ID", "user_123")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestOAuth2Connect_Success(t *testing.T) {
	const expectedRedirectURL = "https://login.microsoftonline.com/oauth2/authorize?client_id=test"
	app := &mockOAuth2App{
		initOAuth2Fn: func(userID string) (string, error) {
			require.Equal(t, "user_123", userID)
			return expectedRedirectURL, nil
		},
	}
	router := setupOAuth2Router(app)

	t.Run("GET redirects browser with 302 Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oauth2/connect", nil)
		req.Header.Set("Mattermost-User-ID", "user_123")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusFound, rec.Code)
		require.Equal(t, expectedRedirectURL, rec.Header().Get("Location"))
	})

	t.Run("POST returns PostActionIntegrationResponse with GotoLocation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/oauth2/connect", nil)
		req.Header.Set("Mattermost-User-ID", "user_123")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var resp model.PostActionIntegrationResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		require.Equal(t, expectedRedirectURL, resp.GotoLocation)
	})
}
