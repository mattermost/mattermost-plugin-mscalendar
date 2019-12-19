// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"
	"net/url"
)

type AuthResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

func (c *client) getAppLevelToken() (string, error) {
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

	_, err := c.Call(http.MethodPost, u, data, &res)
	if err != nil {
		return "", err
	}

	return res.AccessToken, nil
}

type SingleRequest struct {
	ID      string            `json:"id"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Body    interface{}       `json:"body"`
	Headers map[string]string `json:"headers"`
}

type SingleResponse struct {
	ID      string            `json:"id"`
	Status  int               `json:"status"`
	Body    interface{}       `json:"body"`
	Headers map[string]string `json:"headers"`
}

type FullBatchResponse struct {
	Responses []*SingleResponse `json:"responses"`
}

type FullBatchRequest struct {
	Requests []*SingleRequest `json:"requests"`
}

func (c *client) batchRequest(requests []*SingleRequest, out interface{}) error {
	batchReq := FullBatchRequest{Requests: requests}
	u := "https://graph.microsoft.com/v1.0/$batch"

	_, err := c.Call(http.MethodPost, u, batchReq, out)
	if err != nil {
		return err
	}

	return nil
}
