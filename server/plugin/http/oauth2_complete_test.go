package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-msoffice/server/api"
	"github.com/mattermost/mattermost-plugin-msoffice/server/api/mock_api"
	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote/msgraph"
	"github.com/mattermost/mattermost-plugin-msoffice/server/store/mock_store"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/bot/mock_bot"
)

func TestOAuth2Complete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	tcs := []struct {
		name                 string
		mattermostUserID     string
		queryStr             string
		setup                func(api.Dependencies)
		registerResponder    func()
		expectedHTTPCode     int
		expectedHTTPResponse string
	}{
		{
			name:                 "unauthorized user",
			expectedHTTPCode:     http.StatusUnauthorized,
			expectedHTTPResponse: "Not authorized\n",
		},
		{
			name:                 "missing authorization code",
			mattermostUserID:     "fake@mattermost.com",
			queryStr:             "code=",
			expectedHTTPCode:     http.StatusBadRequest,
			expectedHTTPResponse: "missing authorization code\n",
		},
		{
			name:             "missing state",
			mattermostUserID: "fake@mattermost.com",
			queryStr:         "code=fakecode&state=",
			setup: func(d api.Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().VerifyOAuth2State(gomock.Eq("")).Return(errors.New("unable to verify state")).Times(1)
			},
			expectedHTTPCode:     http.StatusUnauthorized,
			expectedHTTPResponse: "missing stored state: unable to verify state\n",
		},
		{
			name:             "user state not authorized",
			mattermostUserID: "fake@mattermost.com",
			queryStr:         "code=fakecode&state=user_nomatch@mattermost.com",
			setup: func(d api.Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().VerifyOAuth2State(gomock.Eq("user_nomatch@mattermost.com")).Return(nil).Times(1)
			},
			expectedHTTPCode:     http.StatusUnauthorized,
			expectedHTTPResponse: "not authorized, user ID mismatch\n",
		},
		{
			name:              "unable to exchange auth code for token",
			mattermostUserID:  "fake@mattermost.com",
			queryStr:          "code=fakecode&state=user_fake@mattermost.com",
			registerResponder: badTokenExchangeResponder,
			setup: func(d api.Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().VerifyOAuth2State(gomock.Eq("user_fake@mattermost.com")).Return(nil).Times(1)
			},
			expectedHTTPCode:     http.StatusUnauthorized,
			expectedHTTPResponse: "oauth2: cannot fetch token: 400\nResponse: {\"error\":\"invalid request\"}\n",
		},
		{
			name:              "microsoft graph api client unable to get user info",
			mattermostUserID:  "fake@mattermost.com",
			queryStr:          "code=fakecode&state=user_fake@mattermost.com",
			registerResponder: unauthorizedTokenGraphAPIResponder,
			setup: func(d api.Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().VerifyOAuth2State(gomock.Eq("user_fake@mattermost.com")).Return(nil).Times(1)
			},
			expectedHTTPCode:     http.StatusUnauthorized,
			expectedHTTPResponse: `401: {"error":{"code":"InvalidAuthenticationToken","message":"Access token is empty.","innerError":{"date":"2019-11-12T00:49:46","request-id":"d1a6e016-c7c4-4caf-9a7f-2d7079dc05d2"}}}` + "\n",
		},
		{
			name:              "UserStore unable to store user info",
			mattermostUserID:  "fake@mattermost.com",
			queryStr:          "code=fakecode&state=user_fake@mattermost.com",
			registerResponder: statusOKGraphAPIResponder,
			setup: func(d api.Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				us := d.UserStore.(*mock_store.MockUserStore)
				us.EXPECT().StoreUser(gomock.Any()).Return(errors.New("forced kvstore error")).Times(1)
				ss.EXPECT().VerifyOAuth2State(gomock.Eq("user_fake@mattermost.com")).Return(nil).Times(1)
			},
			expectedHTTPCode:     http.StatusUnauthorized,
			expectedHTTPResponse: "forced kvstore error\n",
		},
		{
			name:              "successfully completed oauth2 login",
			mattermostUserID:  "fake@mattermost.com",
			queryStr:          "code=fakecode&state=user_fake@mattermost.com",
			registerResponder: statusOKGraphAPIResponder,
			setup: func(d api.Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				us := d.UserStore.(*mock_store.MockUserStore)
				poster := d.Poster.(*mock_bot.MockPoster)

				us.EXPECT().StoreUser(gomock.Any()).Return(nil).Times(1)
				ss.EXPECT().VerifyOAuth2State(gomock.Eq("user_fake@mattermost.com")).Return(nil).Times(1)
				poster.EXPECT().PostDirect(
					gomock.Eq("fake@mattermost.com"), gomock.Eq(getBotPosterMessage("displayName-value")), gomock.Eq("custom_TODO")).
					Return(nil).Times(1)
			},
			expectedHTTPCode: http.StatusOK,
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
		},
	}

	handler := NewHandler()

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.registerResponder != nil {
				tc.registerResponder()
			}

			conf := &config.Config{
				StoredConfig: config.StoredConfig{
					OAuth2Authority:    "common",
					OAuth2ClientID:     "fakeclientid",
					OAuth2ClientSecret: "fakeclientsecret",
				},
				PluginURL: "http://localhost",
			}

			dependencies := mock_api.NewMockDependencies(ctrl)
			dependencies.Remote = remote.Makers[msgraph.Kind](conf, utils.NilLogger)
			if tc.setup != nil {
				tc.setup(dependencies)
			}

			r := newHTTPRequest(conf, dependencies, tc.mattermostUserID, tc.queryStr)
			w := defaultMockResponseWriter()

			handler.oauth2Complete(w, r)

			assert.Equal(t, tc.expectedHTTPCode, w.StatusCode)
			assert.Equal(t, tc.expectedHTTPResponse, string(w.Bytes))
		})
	}
}

func badTokenExchangeResponder() {
	url := "https://login.microsoftonline.com/common/oauth2/v2.0/token"

	responder := httpmock.NewStringResponder(http.StatusBadRequest, `{"error":"invalid request"}`)

	httpmock.RegisterResponder("POST", url, responder)
}

func unauthorizedTokenGraphAPIResponder() {
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

func statusOKGraphAPIResponder() {
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

func newHTTPRequest(conf *config.Config, dependencies api.Dependencies, mattermostUserID, rawQuery string) *http.Request {
	r := &http.Request{
		Header: make(http.Header),
	}

	ctx := r.Context()
	ctx = api.Context(ctx, api.New(dependencies, conf, mattermostUserID))
	ctx = config.Context(ctx, conf)

	r.URL = &url.URL{
		RawQuery: rawQuery,
	}
	r.Header.Add("Mattermost-User-ID", mattermostUserID)

	return r.WithContext(ctx)
}
