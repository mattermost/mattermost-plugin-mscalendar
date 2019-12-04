// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

type User struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName,omitempty"`
	UserPrincipalName string `json:"userPrincipalName,omitempty"`
}
