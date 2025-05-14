// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

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
	u := EntraIDEndpoint(c.conf.OAuth2Authority, c.conf.OAuth2TenantType).TokenURL
	res := AuthResponse{}

	data := url.Values{}
	data.Set("client_id", c.conf.OAuth2ClientID)
	data.Set("scope", MSGraphEndpoint(c.conf.OAuth2TenantType)+"/.default")
	data.Set("client_secret", c.conf.OAuth2ClientSecret)
	data.Set("grant_type", "client_credentials")

	_, err := c.CallFormPost(http.MethodPost, u, data, &res)
	if err != nil {
		return "", errors.Wrap(err, "msgraph GetSuperuserToken")
	}

	return res.AccessToken, nil
}
