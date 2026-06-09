// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package msgraph

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot"
)

func TestHandleWebhook(t *testing.T) {
	remote := &impl{
		conf:   &config.Config{},
		logger: &bot.NilLogger{},
	}

	validWebhook := `{
		"value": [{
			"changeType": "updated",
			"subscriptionId": "sub-123",
			"subscriptionExpirationDateTime": "2030-01-01T00:00:00Z"
		}]
	}`

	tests := []struct {
		name                string
		body                string
		wantStatus          int
		wantNotificationLen int
		wantSubscriptionID  string
	}{
		{
			name:                "empty value array",
			body:                `{"value":[]}`,
			wantStatus:          http.StatusAccepted,
			wantNotificationLen: 0,
		},
		{
			name:                "null entry in value array",
			body:                `{"value":[null]}`,
			wantStatus:          http.StatusAccepted,
			wantNotificationLen: 0,
		},
		{
			name:                "mixed null and valid entries",
			body:                `{"value":[null,{"changeType":"updated","subscriptionId":"sub-123","subscriptionExpirationDateTime":"2030-01-01T00:00:00Z"}]}`,
			wantStatus:          http.StatusAccepted,
			wantNotificationLen: 1,
			wantSubscriptionID:  "sub-123",
		},
		{
			name:                "valid webhook entry",
			body:                validWebhook,
			wantStatus:          http.StatusAccepted,
			wantNotificationLen: 1,
			wantSubscriptionID:  "sub-123",
		},
		{
			name:                "invalid json",
			body:                `{invalid`,
			wantStatus:          http.StatusBadRequest,
			wantNotificationLen: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/notification/v1/event", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()

			require.NotPanics(t, func() {
				notifications := remote.HandleWebhook(rec, req)
				require.Len(t, notifications, tc.wantNotificationLen)
				if tc.wantSubscriptionID != "" {
					require.Equal(t, tc.wantSubscriptionID, notifications[0].SubscriptionID)
				}
			})

			require.Equal(t, tc.wantStatus, rec.Result().StatusCode)
		})
	}
}

func TestHandleWebhookValidationToken(t *testing.T) {
	remote := &impl{
		conf:   &config.Config{},
		logger: &bot.NilLogger{},
	}

	req := httptest.NewRequest(http.MethodPost, "/notification/v1/event?validationToken=test-token", nil)
	rec := httptest.NewRecorder()

	notifications := remote.HandleWebhook(rec, req)

	require.Nil(t, notifications)
	require.Equal(t, http.StatusOK, rec.Result().StatusCode)
	require.Equal(t, "test-token", rec.Body.String())
}
