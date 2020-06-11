// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package oauth2connect

import (
	"encoding/json"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"golang.org/x/oauth2"
)

const (
	oAuth2StateTimeToLive = 300 // seconds
)

type OAuther interface {
	GetToken(userID string) (*oauth2.Token, error)
	GetURL() string
	Deauth(userID string) error
}

type oAuther struct {
	api             plugin.API
	onConnect       func(userID string, token *oauth2.Token)
	storePrefix     string
	pluginURL       string
	oAuthURL        string
	connectedString string
	config          *oauth2.Config
}

func NewOAuther(h *httputils.Handler, api plugin.API, pluginURL, oAuthURL, storePrefix, connectedString string, onConnect func(userID string, token *oauth2.Token), oAuthConfig *oauth2.Config) OAuther {
	o := &oAuther{
		api:             api,
		onConnect:       onConnect,
		storePrefix:     storePrefix,
		pluginURL:       pluginURL,
		oAuthURL:        oAuthURL,
		config:          oAuthConfig,
		connectedString: connectedString,
	}

	o.config.RedirectURL = pluginURL + oAuthURL + "/complete"

	oauth2Router := h.Router.PathPrefix(oAuthURL).Subrouter()
	oauth2Router.HandleFunc("/connect", o.oauth2Connect).Methods("GET")
	oauth2Router.HandleFunc("/complete", o.oauth2Complete).Methods("GET")

	return o
}

func (o *oAuther) GetURL() string {
	return o.pluginURL + o.oAuthURL + "/connect"
}

func (o *oAuther) GetToken(userID string) (*oauth2.Token, error) {
	rawToken, appErr := o.api.KVGet(o.getTokenKey(userID))
	if appErr != nil {
		return nil, appErr
	}

	var token *oauth2.Token
	err := json.Unmarshal(rawToken, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (o *oAuther) getTokenKey(userID string) string {
	return o.storePrefix + "token_" + userID
}

func (o *oAuther) getStateKey(userID string) string {
	return o.storePrefix + "state_" + userID
}

func (o *oAuther) Deauth(userID string) error {
	err := o.api.KVDelete(o.getTokenKey(userID))
	if err != nil {
		return err
	}

	return nil
}
