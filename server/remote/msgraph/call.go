// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	msgraph "github.com/yaegashi/msgraph.go/v1.0"
)

func (c *client) Call(method, path string, in, out interface{}) error {
	// parts := strings.Split(path, "/")
	// if len(parts) != 4 || !strings.EqualFold(k1, parts[0]) || !strings.EqualFold(k2, parts[2]) {
	// 	return errors.Errorf("invalid resource format %q, expected /%s/{id}/%s/{id}", path, k1, k2)
	// }
	// id1, id2 := parts[1], parts[3]

	errContext := fmt.Sprintf("msgraph: Call failed: method:%s, path:%s", method, path)
	baseURL, err := url.Parse(c.rbuilder.URL())
	if err != nil {
		return errors.WithMessage(err, errContext)
	}
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	path = baseURL.String() + path

	var inBody io.Reader
	if in != nil {
		buf := &bytes.Buffer{}
		err = json.NewEncoder(buf).Encode(in)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, path, inBody)
	if err != nil {
		return err
	}
	if inBody != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	if c.ctx != nil {
		req = req.WithContext(c.ctx)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		if out != nil {
			err := json.NewDecoder(resp.Body).Decode(out)
			if err != nil {
				return err
			}
		}
		return nil

	case http.StatusNoContent:
		return nil

	}

	errResp := msgraph.ErrorResponse{Response: resp}
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	if err != nil {
		return errors.WithMessagef(err, "status: %s", resp.Status)
	}
	if err != nil {
		return err
	}
	return &errResp
}
