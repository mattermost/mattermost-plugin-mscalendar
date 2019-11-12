package msgraph

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	graph "github.com/jkrecek/msgraph-go"
	"github.com/stretchr/testify/assert"
)

func TestGetUserCalendar(t *testing.T) {
	tcs := []struct {
		name                  string
		client                Client
		registerResponderFunc func()
		expectedCalendars     []*graph.Calendar
		expectedErr           error
	}{
		{
			name:                  "successful get calendars api call (no token refresh)",
			client:                NewClient(testConfig(), getToken(time.Now().Add(time.Hour))),
			registerResponderFunc: statusOKGraphAPICalendarResponderFunc,
			expectedCalendars: []*graph.Calendar{
				&graph.Calendar{
					Id:        "id-value",
					Name:      "name-value",
					Color:     "color-value",
					ChangeKey: "changeKey-value",
				},
			},
			expectedErr: nil,
		},
		{
			name:                  "unsuccessful get calendars api call (token refresh needed)",
			client:                NewClient(testConfig(), getToken(time.Now())),
			registerResponderFunc: statusOKGraphAPICalendarResponderFunc,
			expectedCalendars:     nil,
			expectedErr: &url.Error{
				Op:  "Get",
				URL: "https://graph.microsoft.com/v1.0/me/calendars",
				Err: &url.Error{
					Op:  "Post",
					URL: fmt.Sprintf(tokenURLEndpoint, testConfig().OAuth2Authority),
					Err: errors.New("no responder found"),
				},
			},
		},
		{
			name:   "successful get calendars api call (with token refresh)",
			client: NewClient(testConfig(), getToken(time.Now())),
			registerResponderFunc: func() {
				statusOKTokenRefreshResponderFunc()
				statusOKGraphAPICalendarResponderFunc()
			},
			expectedCalendars: []*graph.Calendar{
				&graph.Calendar{
					Id:        "id-value",
					Name:      "name-value",
					Color:     "color-value",
					ChangeKey: "changeKey-value",
				},
			},
			expectedErr: nil,
		},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.registerResponderFunc()

			calendars, err := tc.client.GetUserCalendar("")
			if err != nil {
				t.Log(err.Error())
			}

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedCalendars, calendars)
		})
	}
}

func statusOKGraphAPICalendarResponderFunc() {
	meRequestURL := "https://graph.microsoft.com/v1.0/me/calendars"

	meResponder := httpmock.NewStringResponder(http.StatusOK, `{
    "value": [
        {
            "changeKey": "changeKey-value",
            "name": "name-value",
            "color": "color-value",
            "id": "id-value"
        }
    ]
}`)

	httpmock.RegisterResponder("GET", meRequestURL, meResponder)
}
