// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package settingspanel

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/httputils"
)

const settingsURL = "/settings"

type stubPanel struct {
	setCalled bool
}

func (p *stubPanel) Set(userID, settingID string, value interface{}) error {
	p.setCalled = true
	return nil
}
func (p *stubPanel) Print(userID string)                       {}
func (p *stubPanel) ToPost(userID string) (*model.Post, error) { return &model.Post{}, nil }
func (p *stubPanel) Clear(userID string) error                 { return nil }
func (p *stubPanel) URL() string                               { return settingsURL }
func (p *stubPanel) GetSettingIDs() []string                   { return nil }

func TestHandleActionInvalidSettingID(t *testing.T) {
	panel := &stubPanel{}
	handler := httputils.NewHandler()
	Init(handler, panel)

	requestBody := model.PostActionIntegrationRequest{
		Context: map[string]interface{}{
			ContextIDKey:          123,
			ContextButtonValueKey: "on",
		},
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, settingsURL, io.NopCloser(bytes.NewBuffer(bodyBytes)))
	req.Header.Set("Mattermost-User-Id", "mockUserID")
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		handler.ServeHTTP(rec, req)
	})

	assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	var response model.PostActionIntegrationResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response.EphemeralText, "Error: invalid setting id")
	assert.False(t, panel.setCalled, "panel.Set should not be called for an invalid setting id")
}
