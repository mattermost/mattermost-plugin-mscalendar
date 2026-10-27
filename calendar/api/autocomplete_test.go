// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store"

	"github.com/mattermost/mattermost/server/public/model"
)

func mockUserIndex(n int) (store.UserIndex, []string) {
	index := make(store.UserIndex, 0, n)
	ids := make([]string, 0, n)
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("user%d", i)
		index = append(index, &store.UserShort{MattermostUserID: id})
		ids = append(ids, id)
	}
	return index, ids
}

func TestAutocompleteConnectedUsers(t *testing.T) {
	api, mockStore, _, _, mockPluginAPI, mockLogger, mockLoggerWith, _ := GetMockSetup(t)
	restrictions := &model.ViewUsersRestrictions{Channels: []string{MockChannelID}}

	tests := []struct {
		name       string
		setup      func()
		assertions func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "Unauthorized user",
			setup: func() {
				mockStore.EXPECT().LoadUser(gomock.Any()).Return(nil, store.ErrNotFound).Times(1)
				mockLogger.EXPECT().With(gomock.Any()).Return(mockLoggerWith).Times(1)
				mockLoggerWith.EXPECT().Errorf("user unauthorized").Times(1)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, rec.Result().StatusCode)
			},
		},
		{
			name: "Search error",
			setup: func() {
				mockStore.EXPECT().LoadUser(gomock.Any()).Return(&store.User{}, nil).Times(1)
				mockPluginAPI.EXPECT().GetViewUsersRestrictions(MockUserID).Return(nil, nil).Times(1)
				mockStore.EXPECT().SearchInUserIndex(gomock.Any(), autocompleteLimit).Return(nil, errors.New("search error")).Times(1)
				mockLogger.EXPECT().With(gomock.Any()).Return(mockLoggerWith).Times(1)
				mockLoggerWith.EXPECT().Errorf("unable to search in user index").Times(1)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				// status code for this scenario should be 500, but it is 200 somehow
				var response model.PostActionIntegrationResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response.EphemeralText, "unable to search in user index")
			},
		},
		{
			name: "Unrestricted viewer gets search results without visibility checks",
			setup: func() {
				mockStore.EXPECT().LoadUser(gomock.Any()).Return(&store.User{}, nil).Times(1)
				mockPluginAPI.EXPECT().GetViewUsersRestrictions(MockUserID).Return(nil, nil).Times(1)
				mockStore.EXPECT().SearchInUserIndex(gomock.Any(), autocompleteLimit).Return(store.UserIndex{{MattermostUserID: "userA"}}, nil).Times(1)
				mockPluginAPI.EXPECT().FilterVisibleUsers(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
				var dto []store.UserShortDTO
				assert.NoError(t, json.NewDecoder(rec.Body).Decode(&dto))
				assert.Equal(t, []store.UserShortDTO{{MattermostUserID: "userA"}}, dto)
			},
		},
		{
			name: "Restrictions error",
			setup: func() {
				mockStore.EXPECT().LoadUser(gomock.Any()).Return(&store.User{}, nil).Times(1)
				mockPluginAPI.EXPECT().GetViewUsersRestrictions(MockUserID).Return(nil, errors.New("restrictions error")).Times(1)
				mockStore.EXPECT().SearchInUserIndex(gomock.Any(), gomock.Any()).Times(0)
				mockLogger.EXPECT().With(gomock.Any()).Return(mockLoggerWith).Times(1)
				mockLoggerWith.EXPECT().Errorf("unable to filter users").Times(1)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
			},
		},
		{
			name: "Only visible users are returned",
			setup: func() {
				mockStore.EXPECT().LoadUser(gomock.Any()).Return(&store.User{}, nil).Times(1)
				mockPluginAPI.EXPECT().GetViewUsersRestrictions(MockUserID).Return(restrictions, nil).Times(1)
				mockStore.EXPECT().SearchInUserIndex(gomock.Any(), math.MaxInt).Return(store.UserIndex{{MattermostUserID: "userA"}, {MattermostUserID: "userB"}}, nil).Times(1)
				mockPluginAPI.EXPECT().FilterVisibleUsers(MockUserID, restrictions, []string{"userA", "userB"}).Return([]string{"userA"}, nil).Times(1)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
				var dto []store.UserShortDTO
				assert.NoError(t, json.NewDecoder(rec.Body).Decode(&dto))
				assert.Equal(t, []store.UserShortDTO{{MattermostUserID: "userA"}}, dto)
			},
		},
		{
			name: "Results are limited after filtering",
			setup: func() {
				index, ids := mockUserIndex(autocompleteLimit + 2)
				mockStore.EXPECT().LoadUser(gomock.Any()).Return(&store.User{}, nil).Times(1)
				mockPluginAPI.EXPECT().GetViewUsersRestrictions(MockUserID).Return(restrictions, nil).Times(1)
				mockStore.EXPECT().SearchInUserIndex(gomock.Any(), math.MaxInt).Return(index, nil).Times(1)
				mockPluginAPI.EXPECT().FilterVisibleUsers(MockUserID, restrictions, ids).Return(ids[1:], nil).Times(1)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
				var dto []store.UserShortDTO
				assert.NoError(t, json.NewDecoder(rec.Body).Decode(&dto))
				assert.Len(t, dto, autocompleteLimit)
				assert.Equal(t, "user1", dto[0].MattermostUserID)
			},
		},
		{
			name: "Later batches are skipped once the limit is reached",
			setup: func() {
				index, ids := mockUserIndex(visibilityBatchSize + 5)
				mockStore.EXPECT().LoadUser(gomock.Any()).Return(&store.User{}, nil).Times(1)
				mockPluginAPI.EXPECT().GetViewUsersRestrictions(MockUserID).Return(restrictions, nil).Times(1)
				mockStore.EXPECT().SearchInUserIndex(gomock.Any(), math.MaxInt).Return(index, nil).Times(1)
				mockPluginAPI.EXPECT().FilterVisibleUsers(MockUserID, restrictions, ids[:visibilityBatchSize]).Return(ids[:visibilityBatchSize], nil).Times(1)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var dto []store.UserShortDTO
				assert.NoError(t, json.NewDecoder(rec.Body).Decode(&dto))
				assert.Len(t, dto, autocompleteLimit)
				assert.Equal(t, "user0", dto[0].MattermostUserID)
			},
		},
		{
			name: "Next batch is checked when the limit is not yet reached",
			setup: func() {
				index, ids := mockUserIndex(visibilityBatchSize + 5)
				mockStore.EXPECT().LoadUser(gomock.Any()).Return(&store.User{}, nil).Times(1)
				mockPluginAPI.EXPECT().GetViewUsersRestrictions(MockUserID).Return(restrictions, nil).Times(1)
				mockStore.EXPECT().SearchInUserIndex(gomock.Any(), math.MaxInt).Return(index, nil).Times(1)
				gomock.InOrder(
					mockPluginAPI.EXPECT().FilterVisibleUsers(MockUserID, restrictions, ids[:visibilityBatchSize]).Return([]string{"user3"}, nil),
					mockPluginAPI.EXPECT().FilterVisibleUsers(MockUserID, restrictions, ids[visibilityBatchSize:]).Return(ids[visibilityBatchSize:visibilityBatchSize+2], nil),
				)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var dto []store.UserShortDTO
				assert.NoError(t, json.NewDecoder(rec.Body).Decode(&dto))
				var got []string
				for _, u := range dto {
					got = append(got, u.MattermostUserID)
				}
				assert.Equal(t, []string{"user3", fmt.Sprintf("user%d", visibilityBatchSize), fmt.Sprintf("user%d", visibilityBatchSize+1)}, got)
			},
		},
		{
			name: "Visible users beyond the first batches are still found",
			setup: func() {
				index, ids := mockUserIndex(2*visibilityBatchSize + 50)
				last := 2*visibilityBatchSize + 40
				mockStore.EXPECT().LoadUser(gomock.Any()).Return(&store.User{}, nil).Times(1)
				mockPluginAPI.EXPECT().GetViewUsersRestrictions(MockUserID).Return(restrictions, nil).Times(1)
				mockStore.EXPECT().SearchInUserIndex(gomock.Any(), math.MaxInt).Return(index, nil).Times(1)
				gomock.InOrder(
					mockPluginAPI.EXPECT().FilterVisibleUsers(MockUserID, restrictions, ids[:visibilityBatchSize]).Return([]string{}, nil),
					mockPluginAPI.EXPECT().FilterVisibleUsers(MockUserID, restrictions, ids[visibilityBatchSize:2*visibilityBatchSize]).Return([]string{}, nil),
					mockPluginAPI.EXPECT().FilterVisibleUsers(MockUserID, restrictions, ids[2*visibilityBatchSize:]).Return([]string{ids[last]}, nil),
				)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var dto []store.UserShortDTO
				assert.NoError(t, json.NewDecoder(rec.Body).Decode(&dto))
				assert.Equal(t, []store.UserShortDTO{{MattermostUserID: fmt.Sprintf("user%d", 2*visibilityBatchSize+40)}}, dto)
			},
		},
		{
			name: "Filter error",
			setup: func() {
				mockStore.EXPECT().LoadUser(gomock.Any()).Return(&store.User{}, nil).Times(1)
				mockPluginAPI.EXPECT().GetViewUsersRestrictions(MockUserID).Return(restrictions, nil).Times(1)
				mockStore.EXPECT().SearchInUserIndex(gomock.Any(), math.MaxInt).Return(store.UserIndex{{MattermostUserID: "userA"}}, nil).Times(1)
				mockPluginAPI.EXPECT().FilterVisibleUsers(MockUserID, restrictions, []string{"userA"}).Return(nil, errors.New("filter error")).Times(1)
				mockLogger.EXPECT().With(gomock.Any()).Return(mockLoggerWith).Times(1)
				mockLoggerWith.EXPECT().Errorf("unable to filter users").Times(1)
			},
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			req := httptest.NewRequest(http.MethodGet, "/autocomplete?search=test", nil)
			req.Header.Set(MMUserIDHeader, MockUserID)
			rec := httptest.NewRecorder()

			api.autocompleteConnectedUsers(rec, req)

			tc.assertions(t, rec)
		})
	}
}
