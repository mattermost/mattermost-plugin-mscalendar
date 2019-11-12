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

func TestGetMe(t *testing.T) {
	tcs := []struct {
		name                  string
		client                Client
		registerResponderFunc func()
		expectedMe            *graph.Me
		expectedErr           error
	}{
		{
			name:                  "successful get calendar api call (no token refresh)",
			client:                NewClient(testConfig(), getToken(time.Now().Add(time.Hour))),
			registerResponderFunc: statusOKGraphAPIMeResponderFunc,
			expectedMe: &graph.Me{
				Id:                "id-value",
				UserPrincipalName: "userPrincipalName-value",
				GivenName:         "givenName-value",
				DisplayName:       "displayName-value",
				Surname:           "surname-value",
			},
			expectedErr: nil,
		},
		{
			name:                  "unsuccessful get calendar api call (token refresh needed)",
			client:                NewClient(testConfig(), getToken(time.Now())),
			registerResponderFunc: statusOKGraphAPIMeResponderFunc,
			expectedMe:            nil,
			expectedErr: &url.Error{
				Op:  "Get",
				URL: "https://graph.microsoft.com/v1.0/me",
				Err: &url.Error{
					Op:  "Post",
					URL: fmt.Sprintf(tokenURLEndpoint, testConfig().OAuth2Authority),
					Err: errors.New("no responder found"),
				},
			},
		},
		{
			name:   "successful get calendar api call (with token refresh)",
			client: NewClient(testConfig(), getToken(time.Now())),
			registerResponderFunc: func() {
				statusOKTokenRefreshResponderFunc()
				statusOKGraphAPIMeResponderFunc()
			},
			expectedMe: &graph.Me{
				Id:                "id-value",
				UserPrincipalName: "userPrincipalName-value",
				GivenName:         "givenName-value",
				DisplayName:       "displayName-value",
				Surname:           "surname-value",
			},
			expectedErr: nil,
		},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.registerResponderFunc()

			me, err := tc.client.GetMe()
			if err != nil {
				t.Log(err.Error())
			}

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedMe, me)
		})
	}
}

func statusOKGraphAPIMeResponderFunc() {
	meRequestURL := "https://graph.microsoft.com/v1.0/me"

	meResponder := httpmock.NewStringResponder(http.StatusOK, `{
    "businessPhones": [
        "businessPhones-value"
    ],
    "displayName": "displayName-value",
    "givenName": "givenName-value",
    "jobTitle": "jobTitle-value",
    "mail": "mail-value",
    "mobilePhone": "mobilePhone-value",
    "officeLocation": "officeLocation-value",
    "preferredLanguage": "preferredLanguage-value",
    "surname": "surname-value",
    "userPrincipalName": "userPrincipalName-value",
    "id": "id-value"
}`)

	httpmock.RegisterResponder("GET", meRequestURL, meResponder)
}

func statusOKTokenRefreshResponderFunc() {
	tokenURL := "https://login.microsoftonline.com/common/oauth2/v2.0/token"

	tokenResponder := httpmock.NewStringResponder(http.StatusOK, `{
    "token_type": "Bearer",
    "scope": "user.read%20Fmail.read",
    "expires_in": 3600,
    "access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5HVEZ2ZEstZnl0aEV1Q...",
    "refresh_token": "AwABAAAAvPM1KaPlrEqdFSBzjqfTGAMxZGUTdM0t4B4..."
}`)

	httpmock.RegisterResponder("POST", tokenURL, tokenResponder)
}
