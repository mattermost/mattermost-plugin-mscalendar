package mscalendar

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/msgraph"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store/mock_store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot/mock_bot"
)

const (
	fakeID   = "fake@mattermost.com"
	fakeCode = "fakecode"
)

func TestCompleteOAuth2Happy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	statusOKGraphAPIResponder()

	mscalendar, env := newOAuth2MSCalendar(ctrl)
	ss := env.Dependencies.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
	us := env.Dependencies.UserStore.(*mock_store.MockUserStore)
	poster := env.Dependencies.Poster.(*mock_bot.MockPoster)

	state := ""
	gomock.InOrder(
		ss.EXPECT().StoreOAuth2State(gomock.Any()).DoAndReturn(func(s string) error {
			if !strings.HasSuffix(s, "_"+fakeID) {
				return errors.New("invalid state " + s)
			}
			state = s
			return nil
		}).Times(1),
	)

	redirectURL, err := mscalendar.InitOAuth2(fakeID)
	require.NoError(t, err)
	require.NotEmpty(t, redirectURL)
	require.NotEmpty(t, state)

	gomock.InOrder(
		ss.EXPECT().VerifyOAuth2State(gomock.Eq(state)).Return(nil).Times(1),
		us.EXPECT().StoreUser(gomock.Any()).Return(nil).Times(1),
		poster.EXPECT().DM(
			gomock.Eq(fakeID),
			gomock.Eq(WelcomeMessage),
			gomock.Eq("displayName-value"),
		).Return(nil).Times(1),
	)

	err = mscalendar.CompleteOAuth2(fakeID, fakeCode, state)
	require.NoError(t, err)
}

func TestInitOAuth2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tcs := []struct {
		name             string
		mattermostUserID string
		queryStr         string
		setup            func(dependencies *Dependencies)
		expectError      bool
		expectURL        string
	}{
		{
			name:             "unable to store user state",
			mattermostUserID: fakeID,
			setup: func(d *Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().StoreOAuth2State(gomock.Any()).Return(errors.New("unable to store state")).Times(1)
			},
			expectError: true,
		},
		{
			name:             "successful redirect",
			mattermostUserID: fakeID,
			setup: func(d *Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().StoreOAuth2State(gomock.Any()).Return(nil).Times(1)
			},
			expectURL: "https://login.microsoftonline.com/common/oauth2/v2.0/authorize?access_type=offline&client_id=fakeclientid&redirect_uri=http%3A%2F%2Flocalhost%2Foauth2%2Fcomplete&response_type=code&scope=offline_access+User.Read+Calendars.ReadWrite+Calendars.ReadWrite.Shared+Mail.Read+Mail.Send&state=kbb9cs43z3fxxpc_fake%40mattermost.com",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			mscalendar, env := newOAuth2MSCalendar(ctrl)
			if tc.setup != nil {
				tc.setup(env.Dependencies)
			}
			gotURL, err := mscalendar.InitOAuth2(tc.mattermostUserID)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, noState(tc.expectURL), noState(gotURL))
		})
	}
}

func TestCompleteOAuth2Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	tcs := []struct {
		name              string
		mattermostUserID  string
		code              string
		state             string
		setup             func(*Dependencies)
		registerResponder func()
		expectError       string
	}{
		{
			name:        "missing user",
			expectError: "missing user, code or state",
		},
		{
			name:             "missing authorization code",
			mattermostUserID: fakeID,
			expectError:      "missing user, code or state",
		},
		{
			name:             "missing state",
			mattermostUserID: fakeID,
			code:             fakeCode,
			expectError:      "missing user, code or state",
		},
		{
			name:             "user state not authorized",
			mattermostUserID: fakeID,
			code:             fakeCode,
			state:            "user_nomatch@mattermost.com",
			setup: func(d *Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().VerifyOAuth2State(gomock.Eq("user_nomatch@mattermost.com")).Return(nil).Times(1)
			},
			expectError: "not authorized, user ID mismatch",
		},
		{
			name:              "unable to exchange auth code for token",
			mattermostUserID:  fakeID,
			code:              fakeCode,
			state:             "user_" + fakeID,
			registerResponder: badTokenExchangeResponder,
			setup: func(d *Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().VerifyOAuth2State(gomock.Eq("user_fake@mattermost.com")).Return(nil).Times(1)
			},
			expectError: "cannot fetch token: 400",
		},
		{
			name:              "microsoft graph mscalendar client unable to get user info",
			mattermostUserID:  fakeID,
			code:              fakeCode,
			state:             "user_fake@mattermost.com",
			registerResponder: unauthorizedTokenGraphAPIResponder,
			setup: func(d *Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				ss.EXPECT().VerifyOAuth2State(gomock.Eq("user_fake@mattermost.com")).Return(nil).Times(1)
			},
			expectError: "Access token is empty",
		},
		{
			name:              "UserStore unable to store user info",
			mattermostUserID:  fakeID,
			code:              fakeCode,
			state:             "user_fake@mattermost.com",
			registerResponder: statusOKGraphAPIResponder,
			setup: func(d *Dependencies) {
				ss := d.OAuth2StateStore.(*mock_store.MockOAuth2StateStore)
				us := d.UserStore.(*mock_store.MockUserStore)
				us.EXPECT().StoreUser(gomock.Any()).Return(errors.New("forced kvstore error")).Times(1)
				ss.EXPECT().VerifyOAuth2State(gomock.Eq("user_fake@mattermost.com")).Return(nil).Times(1)
			},
			expectError: "forced kvstore error",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.registerResponder != nil {
				tc.registerResponder()
			}

			mscalendar, env := newOAuth2MSCalendar(ctrl)
			if tc.setup != nil {
				tc.setup(env.Dependencies)
			}

			err := mscalendar.CompleteOAuth2(tc.mattermostUserID, tc.code, tc.state)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectError)
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

func newHTTPRequest(mattermostUserID, rawQuery string) *http.Request {
	r := &http.Request{
		Header: make(http.Header),
		URL: &url.URL{
			RawQuery: rawQuery,
		},
	}
	r.Header.Add("Mattermost-User-ID", mattermostUserID)
	return r
}

func newOAuth2MSCalendar(ctrl *gomock.Controller) (MSCalendar, Env) {
	conf := &config.Config{
		StoredConfig: config.StoredConfig{
			OAuth2Authority:    "common",
			OAuth2ClientID:     "fakeclientid",
			OAuth2ClientSecret: "fakeclientsecret",
		},
		PluginURL: "http://localhost",
	}

	env := Env{
		Config: conf,
		Dependencies: &Dependencies{
			UserStore:        mock_store.NewMockUserStore(ctrl),
			OAuth2StateStore: mock_store.NewMockOAuth2StateStore(ctrl),
			Logger:           &bot.NilLogger{},
			Poster:           mock_bot.NewMockPoster(ctrl),
			Remote:           remote.Makers[msgraph.Kind](conf, &bot.NilLogger{}),
		},
	}

	return New(env, ""), env
}

var stateRegexp = regexp.MustCompile(`^(?P<before>.*)&+state=\w+(?P<after>.*)$`)

func noState(in string) string {
	return stateRegexp.ReplaceAllString(in, "$before$after")
}
