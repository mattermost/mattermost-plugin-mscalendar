package oauth2

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/mock_remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/msgraph"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store/mock_store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot/mock_bot"
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
		setup                func(*mscalendar.Dependencies)
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
			setup: func(d *mscalendar.Dependencies) {
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
			setup: func(d *mscalendar.Dependencies) {
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
			setup: func(d *mscalendar.Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().VerifyOAuth2State(gomock.Eq("user_fake@mattermost.com")).Return(nil).Times(1)
			},
			expectedHTTPCode:     http.StatusUnauthorized,
			expectedHTTPResponse: "oauth2: cannot fetch token: 400\nResponse: {\"error\":\"invalid request\"}\n",
		},
		{
			name:              "microsoft graph mscalendar client unable to get user info",
			mattermostUserID:  "fake@mattermost.com",
			queryStr:          "code=fakecode&state=user_fake@mattermost.com",
			registerResponder: unauthorizedTokenGraphAPIResponder,
			setup: func(d *mscalendar.Dependencies) {
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
			setup: func(d *mscalendar.Dependencies) {
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
			setup: func(d *mscalendar.Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				us := d.UserStore.(*mock_store.MockUserStore)
				poster := d.Poster.(*mock_bot.MockPoster)

				us.EXPECT().StoreUser(gomock.Any()).Return(nil).Times(1)
				ss.EXPECT().VerifyOAuth2State(gomock.Eq("user_fake@mattermost.com")).Return(nil).Times(1)
				poster.EXPECT().DM(
					gomock.Eq("fake@mattermost.com"),
					gomock.Eq(mscalendar.WelcomeMessage),
					gomock.Eq("displayName-value"),
				).Return(nil).Times(1)
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
				<p>Completed connecting to Microsoft Calendar. Please close this window.</p>
			</body>
		</html>
		`,
		},
	}

	router := mux.NewRouter()
	RegisterHTTP(router)

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

			dependencies := NewMockDependencies(ctrl)
			dependencies.Remote = remote.Makers[msgraph.Kind](conf, &bot.NilLogger{})
			if tc.setup != nil {
				tc.setup(dependencies)
			}

			apiconf := mscalendar.Config{
				Config:       conf,
				Dependencies: dependencies,
			}

			r := newHTTPRequest(apiconf, tc.mattermostUserID, tc.queryStr)
			w := defaultMockResponseWriter()

			oauth2Complete(w, r)

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

func newHTTPRequest(apiconf mscalendar.Config, mattermostUserID, rawQuery string) *http.Request {
	r := &http.Request{
		Header: make(http.Header),
	}

	ctx := r.Context()
	ctx = mscalendar.Context(ctx, mscalendar.New(apiconf, mattermostUserID), nil)
	ctx = config.Context(ctx, apiconf.Config)

	r.URL = &url.URL{
		RawQuery: rawQuery,
	}
	r.Header.Add("Mattermost-User-ID", mattermostUserID)

	return r.WithContext(ctx)
}

func NewMockDependencies(ctrl *gomock.Controller) *mscalendar.Dependencies {
	return &mscalendar.Dependencies{
		UserStore:         mock_store.NewMockUserStore(ctrl),
		OAuth2StateStore:  mock_store.NewMockOAuth2StateStore(ctrl),
		SubscriptionStore: mock_store.NewMockSubscriptionStore(ctrl),
		Logger:            &bot.NilLogger{},
		Poster:            mock_bot.NewMockPoster(ctrl),
		Remote:            mock_remote.NewMockRemote(ctrl),
	}
}
