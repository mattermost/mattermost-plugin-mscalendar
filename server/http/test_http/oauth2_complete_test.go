package testhttp

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	shttp "github.com/mattermost/mattermost-plugin-msoffice/server/http"
	"github.com/mattermost/mattermost-plugin-msoffice/server/user"
	"github.com/mattermost/mattermost-server/app"
	"github.com/stretchr/testify/assert"
)

func getUserRequest(userID, rawQuery string) *http.Request {
	r := &http.Request{
		Header: make(http.Header),
	}

	r.URL = &url.URL{
		RawQuery: rawQuery,
	}
	r.Header.Add("Mattermost-User-ID", userID)

	return r
}

func TestOAuth2Complete(t *testing.T) {
	api := &app.PluginAPI{}
	kv := newMockKVStore(nil, nil)

	config := &config.Config{}

	config.OAuth2Authority = "common"
	config.OAuth2ClientId = "fakeclientid"
	config.OAuth2ClientSecret = "fakeclientsecret"
	config.PluginURL = "http://localhost"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	tcs := []struct {
		name                  string
		handler               shttp.Handler
		r                     *http.Request
		w                     *mockResponseWriter
		registerResponderFunc func()
		expectedHTTPResponse  string
		expectedHTTPCode      int
	}{
		{
			name: "unauthorized user",
			handler: shttp.Handler{
				Config:           config,
				UserStore:        user.NewStore(kv),
				OAuth2StateStore: newMockOAuth2StateStore(nil),
				API:              api,
			},
			r: &http.Request{},
			w: defaultMockResponseWriter(),
			registerResponderFunc: func() {},
			expectedHTTPResponse:  "Not authorized\n",
			expectedHTTPCode:      http.StatusUnauthorized,
		},
		{
			name: "missing authorization code",
			handler: shttp.Handler{
				Config:           config,
				UserStore:        user.NewStore(kv),
				OAuth2StateStore: newMockOAuth2StateStore(nil),
				API:              api,
			},
			r: getUserRequest("fake@mattermost.com", "code="),
			w: defaultMockResponseWriter(),
			registerResponderFunc: func() {},
			expectedHTTPResponse:  "missing authorization code\n",
			expectedHTTPCode:      http.StatusBadRequest,
		},
		{
			name: "missing state",
			handler: shttp.Handler{
				Config:           config,
				UserStore:        user.NewStore(kv),
				OAuth2StateStore: newMockOAuth2StateStore(errors.New("unable to verify state")),
				API:              api,
			},
			r: getUserRequest("fake@mattermost.com", "code=fakecode&state="),
			w: defaultMockResponseWriter(),
			registerResponderFunc: func() {},
			expectedHTTPResponse:  "missing stored state: unable to verify state\n",
			expectedHTTPCode:      http.StatusBadRequest,
		},
		{
			name: "user state not authorized",
			handler: shttp.Handler{
				Config:           config,
				UserStore:        user.NewStore(kv),
				OAuth2StateStore: newMockOAuth2StateStore(nil),
				API:              api,
			},
			r: getUserRequest("fake@mattermost.com", "code=fakecode&state=user_nomatch@mattermost.com"),
			w: defaultMockResponseWriter(),
			registerResponderFunc: func() {},
			expectedHTTPResponse:  "Not authorized, user ID mismatch.\n",
			expectedHTTPCode:      http.StatusUnauthorized,
		},
		{
			name: "unable to exchange auth code for token",
			handler: shttp.Handler{
				Config:           config,
				UserStore:        user.NewStore(kv),
				OAuth2StateStore: newMockOAuth2StateStore(nil),
				API:              api,
			},
			r: getUserRequest("fake@mattermost.com", "code=fakecode&state=user_fake@mattermost.com"),
			w: defaultMockResponseWriter(),
			registerResponderFunc: badTokenExchangeResponderFunc,
			expectedHTTPResponse:  "oauth2: cannot fetch token: 400\nResponse: {\"error\":\"invalid request\"}\n",
			expectedHTTPCode:      http.StatusInternalServerError,
		},
		{
			name: "microsoft graph api client unable to get user info",
			handler: shttp.Handler{
				Config:           config,
				UserStore:        user.NewStore(kv),
				OAuth2StateStore: newMockOAuth2StateStore(nil),
				API:              api,
			},
			r: getUserRequest("fake@mattermost.com", "code=fakecode&state=user_fake@mattermost.com"),
			w: defaultMockResponseWriter(),
			registerResponderFunc: unauthorizedTokenGraphAPIResponderFunc,
			expectedHTTPResponse: `Request to url 'https://graph.microsoft.com/v1.0/me' returned error.
    Code: InvalidAuthenticationToken
    Message: Access token is empty.
`,
			expectedHTTPCode: http.StatusInternalServerError,
		},
		{
			name: "UserStore unable to store user info",
			handler: shttp.Handler{
				Config:           config,
				UserStore:        user.NewStore(newMockKVStore(nil, errors.New("forced kvstore error"))),
				OAuth2StateStore: newMockOAuth2StateStore(nil),
				API:              api,
			},
			r: getUserRequest("fake@mattermost.com", "code=fakecode&state=user_fake@mattermost.com"),
			w: defaultMockResponseWriter(),
			registerResponderFunc: statusOKGraphAPIResponderFunc,
			expectedHTTPResponse:  "Unable to connect: forced kvstore error\n",
			expectedHTTPCode:      http.StatusInternalServerError,
		},
		{
			name: "successfully completed oauth2 login",
			handler: shttp.Handler{
				Config:           config,
				UserStore:        user.NewStore(kv),
				OAuth2StateStore: newMockOAuth2StateStore(nil),
				BotPoster:        newMockBotPoster(nil),
				API:              api,
			},
			r: getUserRequest("fake@mattermost.com", "code=fakecode&state=user_fake@mattermost.com"),
			w: defaultMockResponseWriter(),
			registerResponderFunc: statusOKGraphAPIResponderFunc,
			expectedHTTPResponse: `
		<!DOCTYPE html>
		<html>
			<head>
				<script>
					window.close();
				</script>
			</head>
			<body>
				<p>Completed connecting to Microsoft Office. Please close this window.</p>
			</body>
		</html>
		`,
			expectedHTTPCode: http.StatusOK,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.registerResponderFunc()

			tc.handler.OAuth2Complete(tc.w, tc.r)

			assert.Equal(t, tc.expectedHTTPCode, tc.w.StatusCode)
			assert.Equal(t, tc.expectedHTTPResponse, string(tc.w.Bytes))
		})
	}
}

func badTokenExchangeResponderFunc() {
	url := "https://login.microsoftonline.com/common/oauth2/v2.0/token"

	responder := httpmock.NewStringResponder(http.StatusBadRequest, `{"error":"invalid request"}`)

	httpmock.RegisterResponder("POST", url, responder)
}

func unauthorizedTokenGraphAPIResponderFunc() {
	tokenURL := "https://login.microsoftonline.com/common/oauth2/v2.0/token"

	tokenResponder := httpmock.NewStringResponder(http.StatusOK, `{
    "token_type": "Bearer",
    "scope": "user.read%20Fmail.read",
    "expires_in": 3600,
    "access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5HVEZ2ZEstZnl0aEV1Q...",
    "refresh_token": "AwABAAAAvPM1KaPlrEqdFSBzjqfTGAMxZGUTdM0t4B4..."
}`)

	httpmock.RegisterResponder("POST", tokenURL, tokenResponder)

	meRequestURL := "https://graph.microsoft.com/v1.0/me"

	meResponder := httpmock.NewStringResponder(http.StatusUnauthorized, `{
    "error": {
        "code": "InvalidAuthenticationToken",
        "message": "Access token is empty.",
        "innerError": {
            "request-id": "d1a6e016-c7c4-4caf-9a7f-2d7079dc05d2",
            "date": "2019-11-12T00:49:46"
        }
    }
}`)

	httpmock.RegisterResponder("GET", meRequestURL, meResponder)
}

func statusOKGraphAPIResponderFunc() {
	tokenURL := "https://login.microsoftonline.com/common/oauth2/v2.0/token"

	tokenResponder := httpmock.NewStringResponder(http.StatusOK, `{
    "token_type": "Bearer",
    "scope": "user.read%20Fmail.read",
    "expires_in": 3600,
    "access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5HVEZ2ZEstZnl0aEV1Q...",
    "refresh_token": "AwABAAAAvPM1KaPlrEqdFSBzjqfTGAMxZGUTdM0t4B4..."
}`)

	httpmock.RegisterResponder("POST", tokenURL, tokenResponder)

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
