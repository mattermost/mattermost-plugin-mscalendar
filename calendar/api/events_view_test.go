// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store/mock_store"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine/mock_plugin_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote/mock_remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot/mock_bot"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestViewEvents(t *testing.T) {
	now := time.Now().UTC()
	validFrom := now.Format(time.RFC3339)
	validTo := now.Add(24 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		name       string
		setup      func(*http.Request, *api, *mock_store.MockStore, *mock_bot.MockPoster, *mock_remote.MockRemote, *mock_plugin_api.MockPluginAPI, *mock_bot.MockLogger, *mock_bot.MockLogger, *mock_remote.MockClient)
		assertions func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Missing Mattermost-User-Id header",
			setup: func(req *http.Request, _ *api, _ *mock_store.MockStore, _ *mock_bot.MockPoster, _ *mock_remote.MockRemote, _ *mock_plugin_api.MockPluginAPI, mockLogger *mock_bot.MockLogger, _ *mock_bot.MockLogger, _ *mock_remote.MockClient) {
				req.Header.Del(MMUserIDHeader)
				mockLogger.EXPECT().Errorf("viewEvents, unauthorized user").Times(1)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, rec.Result().StatusCode)
				body, _ := io.ReadAll(rec.Body)
				assert.Contains(t, string(body), "unauthorized")
			},
		},
		{
			name: "Missing from query parameter",
			setup: func(req *http.Request, _ *api, _ *mock_store.MockStore, _ *mock_bot.MockPoster, _ *mock_remote.MockRemote, _ *mock_plugin_api.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger, _ *mock_remote.MockClient) {
				req.Header.Set(MMUserIDHeader, MockUserID)
				req.URL.RawQuery = "to=" + validTo
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
				body, _ := io.ReadAll(rec.Body)
				assert.Contains(t, string(body), "from and to query parameters are required")
			},
		},
		{
			name: "Missing to query parameter",
			setup: func(req *http.Request, _ *api, _ *mock_store.MockStore, _ *mock_bot.MockPoster, _ *mock_remote.MockRemote, _ *mock_plugin_api.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger, _ *mock_remote.MockClient) {
				req.Header.Set(MMUserIDHeader, MockUserID)
				req.URL.RawQuery = "from=" + validFrom
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
				body, _ := io.ReadAll(rec.Body)
				assert.Contains(t, string(body), "from and to query parameters are required")
			},
		},
		{
			name: "Invalid from date format",
			setup: func(req *http.Request, _ *api, _ *mock_store.MockStore, _ *mock_bot.MockPoster, _ *mock_remote.MockRemote, _ *mock_plugin_api.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger, _ *mock_remote.MockClient) {
				req.Header.Set(MMUserIDHeader, MockUserID)
				req.URL.RawQuery = "from=not-a-date&to=" + validTo
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
				body, _ := io.ReadAll(rec.Body)
				assert.Contains(t, string(body), "invalid from parameter")
			},
		},
		{
			name: "Invalid to date format",
			setup: func(req *http.Request, _ *api, _ *mock_store.MockStore, _ *mock_bot.MockPoster, _ *mock_remote.MockRemote, _ *mock_plugin_api.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger, _ *mock_remote.MockClient) {
				req.Header.Set(MMUserIDHeader, MockUserID)
				req.URL.RawQuery = "from=" + validFrom + "&to=not-a-date"
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
				body, _ := io.ReadAll(rec.Body)
				assert.Contains(t, string(body), "invalid to parameter")
			},
		},
		{
			name: "From after to",
			setup: func(req *http.Request, _ *api, _ *mock_store.MockStore, _ *mock_bot.MockPoster, _ *mock_remote.MockRemote, _ *mock_plugin_api.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger, _ *mock_remote.MockClient) {
				req.Header.Set(MMUserIDHeader, MockUserID)
				req.URL.RawQuery = "from=" + validTo + "&to=" + validFrom
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
				body, _ := io.ReadAll(rec.Body)
				assert.Contains(t, string(body), "from must be before or equal to to")
			},
		},
		{
			name: "Range exceeds 62 days",
			setup: func(req *http.Request, _ *api, _ *mock_store.MockStore, _ *mock_bot.MockPoster, _ *mock_remote.MockRemote, _ *mock_plugin_api.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger, _ *mock_remote.MockClient) {
				req.Header.Set(MMUserIDHeader, MockUserID)
				farFuture := now.Add(63 * 24 * time.Hour).Format(time.RFC3339)
				req.URL.RawQuery = "from=" + validFrom + "&to=" + farFuture
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
				body, _ := io.ReadAll(rec.Body)
				assert.Contains(t, string(body), "date range must not exceed 62 days")
			},
		},
		{
			name: "User not found in store (disconnected)",
			setup: func(req *http.Request, _ *api, mockStore *mock_store.MockStore, _ *mock_bot.MockPoster, _ *mock_remote.MockRemote, _ *mock_plugin_api.MockPluginAPI, mockLogger *mock_bot.MockLogger, mockLoggerWith *mock_bot.MockLogger, _ *mock_remote.MockClient) {
				req.Header.Set(MMUserIDHeader, MockUserID)
				req.URL.RawQuery = "from=" + validFrom + "&to=" + validTo

				mockStore.EXPECT().LoadUser(MockUserID).Return(nil, store.ErrNotFound).Times(1)
				mockLogger.EXPECT().With(gomock.Any()).Return(mockLoggerWith).Times(1)
				mockLoggerWith.EXPECT().Errorf("viewEvents, user not found in store").Times(1)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, rec.Result().StatusCode)
				body, _ := io.ReadAll(rec.Body)
				assert.Contains(t, string(body), "unauthorized")
			},
		},
		{
			name: "Engine error fetching calendar events",
			setup: func(req *http.Request, _ *api, mockStore *mock_store.MockStore, _ *mock_bot.MockPoster, mockRemote *mock_remote.MockRemote, mockPluginAPI *mock_plugin_api.MockPluginAPI, mockLogger *mock_bot.MockLogger, mockLoggerWith *mock_bot.MockLogger, mockRemoteClient *mock_remote.MockClient) {
				req.Header.Set(MMUserIDHeader, MockUserID)
				req.URL.RawQuery = "from=" + validFrom + "&to=" + validTo

				mockOAuthToken := &oauth2.Token{}
				mockStore.EXPECT().LoadUser(MockUserID).Return(&store.User{
					MattermostUserID: MockUserID,
					OAuth2Token:      mockOAuthToken,
					Remote:           &remote.User{ID: MockRemoteUserID},
				}, nil).Times(2)
				mockPluginAPI.EXPECT().GetMattermostUser(MockUserID).Return(&model.User{Id: MockUserID}, nil).Times(2)
				mockRemote.EXPECT().MakeUserClient(gomock.Any(), mockOAuthToken, MockUserID, gomock.Any(), gomock.Any()).Return(mockRemoteClient).Times(1)
				mockRemoteClient.EXPECT().GetDefaultCalendarView(MockRemoteUserID, gomock.Any(), gomock.Any()).Return(nil, assert.AnError).Times(1)
				mockLogger.EXPECT().With(gomock.Any()).Return(mockLoggerWith).Times(1)
				mockLoggerWith.EXPECT().Errorf("viewEvents, error fetching calendar events").Times(1)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
				body, _ := io.ReadAll(rec.Body)
				assert.Contains(t, string(body), "error fetching calendar events")
				assert.NotContains(t, string(body), "assert.AnError")
			},
		},
		{
			name: "Happy path - returns events",
			setup: func(req *http.Request, _ *api, mockStore *mock_store.MockStore, _ *mock_bot.MockPoster, mockRemote *mock_remote.MockRemote, mockPluginAPI *mock_plugin_api.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger, mockRemoteClient *mock_remote.MockClient) {
				req.Header.Set(MMUserIDHeader, MockUserID)
				req.URL.RawQuery = "from=" + validFrom + "&to=" + validTo

				mockOAuthToken := &oauth2.Token{}
				mockStore.EXPECT().LoadUser(MockUserID).Return(&store.User{
					MattermostUserID: MockUserID,
					OAuth2Token:      mockOAuthToken,
					Remote:           &remote.User{ID: MockRemoteUserID},
				}, nil).Times(2)
				mockPluginAPI.EXPECT().GetMattermostUser(MockUserID).Return(&model.User{Id: MockUserID}, nil).Times(2)
				mockRemote.EXPECT().MakeUserClient(gomock.Any(), mockOAuthToken, MockUserID, gomock.Any(), gomock.Any()).Return(mockRemoteClient).Times(1)
				mockRemoteClient.EXPECT().GetDefaultCalendarView(MockRemoteUserID, gomock.Any(), gomock.Any()).Return([]*remote.Event{
					{Subject: "Test Event"},
				}, nil).Times(1)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
				body, _ := io.ReadAll(rec.Body)
				assert.Contains(t, string(body), "Test Event")
			},
		},
		{
			name: "Happy path - nil events returns empty array",
			setup: func(req *http.Request, _ *api, mockStore *mock_store.MockStore, _ *mock_bot.MockPoster, mockRemote *mock_remote.MockRemote, mockPluginAPI *mock_plugin_api.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger, mockRemoteClient *mock_remote.MockClient) {
				req.Header.Set(MMUserIDHeader, MockUserID)
				req.URL.RawQuery = "from=" + validFrom + "&to=" + validTo

				mockOAuthToken := &oauth2.Token{}
				mockStore.EXPECT().LoadUser(MockUserID).Return(&store.User{
					MattermostUserID: MockUserID,
					OAuth2Token:      mockOAuthToken,
					Remote:           &remote.User{ID: MockRemoteUserID},
				}, nil).Times(2)
				mockPluginAPI.EXPECT().GetMattermostUser(MockUserID).Return(&model.User{Id: MockUserID}, nil).Times(2)
				mockRemote.EXPECT().MakeUserClient(gomock.Any(), mockOAuthToken, MockUserID, gomock.Any(), gomock.Any()).Return(mockRemoteClient).Times(1)
				mockRemoteClient.EXPECT().GetDefaultCalendarView(MockRemoteUserID, gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
				body, _ := io.ReadAll(rec.Body)
				assert.Equal(t, "[]", string(body))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a, mockStore, mockPoster, mockRemote, mockPluginAPI, mockLogger, mockLoggerWith, mockRemoteClient := GetMockSetup(t)
			a.Config = &config.Config{
				Provider: config.ProviderConfig{
					DisplayName:    "TestCalendar",
					CommandTrigger: "testcal",
				},
			}

			req := httptest.NewRequest(http.MethodGet, "/view", nil)
			rec := httptest.NewRecorder()

			tc.setup(req, a, mockStore, mockPoster, mockRemote, mockPluginAPI, mockLogger, mockLoggerWith, mockRemoteClient)
			a.viewEvents(rec, req)

			tc.assertions(t, rec)
		})
	}
}
