// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type AuthResponse struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (c *client) GetSuperuserToken() (string, error) {
	params := map[string]string{
		"client_id":     c.conf.OAuth2ClientID,
		"scope":         "https://graph.microsoft.com/.default",
		"client_secret": c.conf.OAuth2ClientSecret,
		"grant_type":    "client_credentials",
	}

	u := "https://login.microsoftonline.com/" + c.conf.OAuth2Authority + "/oauth2/v2.0/token"
	res := AuthResponse{}

	data := url.Values{}
	data.Set("client_id", params["client_id"])
	data.Set("scope", params["scope"])
	data.Set("client_secret", params["client_secret"])
	data.Set("grant_type", params["grant_type"])

	_, err := c.CallFormPost(http.MethodPost, u, data, &res)
	if err != nil {
		return "", errors.Wrap(err, "msgraph GetSuperuserToken")
	}

	return res.AccessToken, nil
}
