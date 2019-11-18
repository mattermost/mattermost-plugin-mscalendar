// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package utils

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/pkg/errors"
)

func NormalizeInstallURL(mattermostSiteURL, remoteURL string) (string, error) {
	u, err := url.Parse(remoteURL)
	if err != nil {
		return "", err
	}
	if u.Host == "" {
		ss := strings.Split(u.Path, "/")
		if len(ss) > 0 && ss[0] != "" {
			u.Host = ss[0]
			u.Path = path.Join(ss[1:]...)
		}
		u, err = url.Parse(u.String())
		if err != nil {
			return "", err
		}
	}
	if u.Host == "" {
		return "", errors.Errorf("Invalid URL, no hostname: %q", remoteURL)
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	remoteURL = strings.TrimSuffix(u.String(), "/")
	if remoteURL == strings.TrimSuffix(mattermostSiteURL, "/") {
		return "", errors.Errorf("%s is the Mattermost site URL. Please use the remote application's URL.", remoteURL)
	}

	return remoteURL, nil
}

// Reference: https://gobyexample.com/collection-functions
func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func JSON(ref interface{}) string {
	bb, _ := json.MarshalIndent(ref, "", "  ")
	return string(bb)
}

func CodeBlock(in string) string {
	return fmt.Sprintf("\n```\n%s\n```\n", in)
}

func JSONBlock(ref interface{}) string {
	return fmt.Sprintf("\n```json\n%s\n```\n", JSON(ref))
}
