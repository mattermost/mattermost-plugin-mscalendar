package testhttp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	shttp "github.com/mattermost/mattermost-plugin-msoffice/server/http"
	"github.com/mattermost/mattermost-plugin-msoffice/server/kvstore/mock_kvstore"
	"github.com/mattermost/mattermost-plugin-msoffice/server/user"
	"github.com/mattermost/mattermost-plugin-msoffice/server/user/mock_user"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/mock_utils"
	"github.com/mattermost/mattermost-server/app"
	"github.com/stretchr/testify/assert"
)

func makeUserRequest(userID, rawQuery string) *http.Request {
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api := &app.PluginAPI{}

	config := &config.Config{}

	config.OAuth2Authority = "common"
	config.OAuth2ClientId = "fakeclientid"
	config.OAuth2ClientSecret = "fakeclientsecret"
	config.PluginURL = "http://localhost"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	tcs := []struct {
		name                  string
		r                     *http.Request
		setupMocks            func(*mock_kvstore.MockKVStore, *mock_user.MockOAuth2StateStore, *mock_utils.MockBotPoster)
		registerResponderFunc func()
		expectedHTTPResponse  string
		expectedHTTPCode      int
	}{
		{
			name:       "unauthorized user",
			setupMocks: func(kv *mock_kvstore.MockKVStore, ss *mock_user.MockOAuth2StateStore, bp *mock_utils.MockBotPoster) {},
			r:          &http.Request{},
			registerResponderFunc: func() {},
			expectedHTTPResponse:  "Not authorized\n",
			expectedHTTPCode:      http.StatusUnauthorized,
		},
		{
			name:       "missing authorization code",
			setupMocks: func(kv *mock_kvstore.MockKVStore, ss *mock_user.MockOAuth2StateStore, bp *mock_utils.MockBotPoster) {},
			r:          makeUserRequest("fake@mattermost.com", "code="),
			registerResponderFunc: func() {},
			expectedHTTPResponse:  "missing authorization code\n",
			expectedHTTPCode:      http.StatusBadRequest,
		},
		{
			name: "missing state",
			setupMocks: func(kv *mock_kvstore.MockKVStore, ss *mock_user.MockOAuth2StateStore, bp *mock_utils.MockBotPoster) {
				ss.EXPECT().Verify(gomock.Eq("")).Return(errors.New("unable to verify state")).Times(1)
			},
			r: makeUserRequest("fake@mattermost.com", "code=fakecode&state="),
			registerResponderFunc: func() {},
			expectedHTTPResponse:  "missing stored state: unable to verify state\n",
			expectedHTTPCode:      http.StatusBadRequest,
		},
		{
			name: "user state not authorized",
			setupMocks: func(kv *mock_kvstore.MockKVStore, ss *mock_user.MockOAuth2StateStore, bp *mock_utils.MockBotPoster) {
				ss.EXPECT().Verify(gomock.Eq("user_nomatch@mattermost.com")).Return(nil).Times(1)
			},
			r: makeUserRequest("fake@mattermost.com", "code=fakecode&state=user_nomatch@mattermost.com"),
			registerResponderFunc: func() {},
			expectedHTTPResponse:  "Not authorized, user ID mismatch.\n",
			expectedHTTPCode:      http.StatusUnauthorized,
		},
		{
			name: "unable to exchange auth code for token",
			setupMocks: func(kv *mock_kvstore.MockKVStore, ss *mock_user.MockOAuth2StateStore, bp *mock_utils.MockBotPoster) {
				ss.EXPECT().Verify(gomock.Eq("user_fake@mattermost.com")).Return(nil).Times(1)
			},
			r: makeUserRequest("fake@mattermost.com", "code=fakecode&state=user_fake@mattermost.com"),
			registerResponderFunc: badTokenExchangeResponderFunc,
			expectedHTTPResponse:  "oauth2: cannot fetch token: 400\nResponse: {\"error\":\"invalid request\"}\n",
			expectedHTTPCode:      http.StatusInternalServerError,
		},
		{
			name: "microsoft graph api client unable to get user info",
			setupMocks: func(kv *mock_kvstore.MockKVStore, ss *mock_user.MockOAuth2StateStore, bp *mock_utils.MockBotPoster) {
				ss.EXPECT().Verify(gomock.Eq("user_fake@mattermost.com")).Return(nil).Times(1)
			},
			r: makeUserRequest("fake@mattermost.com", "code=fakecode&state=user_fake@mattermost.com"),
			registerResponderFunc: unauthorizedTokenGraphAPIResponderFunc,
			expectedHTTPResponse: `Request to url 'https://graph.microsoft.com/v1.0/me' returned error.
    Code: InvalidAuthenticationToken
    Message: Access token is empty.
`,
			expectedHTTPCode: http.StatusInternalServerError,
		},
		{
			name: "UserStore unable to store user info",
			setupMocks: func(kv *mock_kvstore.MockKVStore, ss *mock_user.MockOAuth2StateStore, bp *mock_utils.MockBotPoster) {
				kv.EXPECT().Store(gomock.Any(), gomock.Any()).Return(errors.New("forced kvstore error")).Times(1)
				ss.EXPECT().Verify(gomock.Eq("user_fake@mattermost.com")).Return(nil).Times(1)
			},
			r: makeUserRequest("fake@mattermost.com", "code=fakecode&state=user_fake@mattermost.com"),
			registerResponderFunc: statusOKGraphAPIResponderFunc,
			expectedHTTPResponse:  "Unable to connect: forced kvstore error\n",
			expectedHTTPCode:      http.StatusInternalServerError,
		},
		{
			name: "successfully completed oauth2 login",
			setupMocks: func(kv *mock_kvstore.MockKVStore, ss *mock_user.MockOAuth2StateStore, bp *mock_utils.MockBotPoster) {
				kv.
					EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(2)

				ss.
					EXPECT().
					Verify(gomock.Eq("user_fake@mattermost.com")).
					Return(nil).
					Times(1)

				bp.
					EXPECT().
					PostDirect(gomock.Eq("fake@mattermost.com"), gomock.Eq(getBotPosterMessage("displayName-value")), gomock.Eq("custom_TODO")).
					Return(nil).
					Times(1)
			},
			r: makeUserRequest("fake@mattermost.com", "code=fakecode&state=user_fake@mattermost.com"),
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

			mockKVStore := mock_kvstore.NewMockKVStore(ctrl)
			mockOAuth2StateStore := mock_user.NewMockOAuth2StateStore(ctrl)
			mockBotPoster := mock_utils.NewMockBotPoster(ctrl)

			tc.setupMocks(mockKVStore, mockOAuth2StateStore, mockBotPoster)

			handler := shttp.Handler{
				Config: config,
				API:    api,
			}

			handler.UserStore = user.NewStore(mockKVStore)
			handler.OAuth2StateStore = mockOAuth2StateStore
			handler.BotPoster = mockBotPoster

			w := defaultMockResponseWriter()

			handler.OAuth2Complete(w, tc.r)

			assert.Equal(t, tc.expectedHTTPCode, w.StatusCode)
			assert.Equal(t, tc.expectedHTTPResponse, string(w.Bytes))
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

	tokenResponse := `{
    "token_type": "Bearer",
    "scope": "user.read%20Fmail.read",
    "expires_in": 3600,
    "access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5HVEZ2ZEstZnl0aEV1Q...",
    "refresh_token": "AwABAAAAvPM1KaPlrEqdFSBzjqfTGAMxZGUTdM0t4B4..."
}`

	tokenResponder := httpmock.NewStringResponder(http.StatusOK, tokenResponse)

	httpmock.RegisterResponder("POST", tokenURL, tokenResponder)

	meRequestURL := "https://graph.microsoft.com/v1.0/me"

	meResponse := `{
    "error": {
        "code": "InvalidAuthenticationToken",
        "message": "Access token is empty.",
        "innerError": {
            "request-id": "d1a6e016-c7c4-4caf-9a7f-2d7079dc05d2",
            "date": "2019-11-12T00:49:46"
        }
    }
}`

	meResponder := httpmock.NewStringResponder(http.StatusUnauthorized, meResponse)

	httpmock.RegisterResponder("GET", meRequestURL, meResponder)
}

func statusOKGraphAPIResponderFunc() {
	tokenURL := "https://login.microsoftonline.com/common/oauth2/v2.0/token"

	tokenResponse := `{
    "token_type": "Bearer",
    "scope": "user.read%20Fmail.read",
    "expires_in": 3600,
    "access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5HVEZ2ZEstZnl0aEV1Q...",
    "refresh_token": "AwABAAAAvPM1KaPlrEqdFSBzjqfTGAMxZGUTdM0t4B4..."
}`

	tokenResponder := httpmock.NewStringResponder(http.StatusOK, tokenResponse)

	httpmock.RegisterResponder("POST", tokenURL, tokenResponder)

	meRequestURL := "https://graph.microsoft.com/v1.0/me"

	meResponse := `{
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
}`

	meResponder := httpmock.NewStringResponder(http.StatusOK, meResponse)

	httpmock.RegisterResponder("GET", meRequestURL, meResponder)
}

func getBotPosterMessage(displayName string) string {
	return fmt.Sprintf("### Welcome to the Microsoft Office plugin!\n"+
		"Here is some info to prove we got you logged in\n"+
		"Name: %s \n", displayName)
}
