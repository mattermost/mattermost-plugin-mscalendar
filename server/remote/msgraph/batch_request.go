// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"
)

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
