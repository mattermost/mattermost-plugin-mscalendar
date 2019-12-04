// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

type Subscription struct {
	ID                 string `json:"id"`
	Resource           string `json:"resource,omitempty"`
	ApplicationID      string `json:"applicationId,omitempty"`
	ChangeType         string `json:"changeType,omitempty"`
	ClientState        string `json:"clientState,omitempty"`
	NotificationURL    string `json:"notificationUrl,omitempty"`
	ExpirationDateTime string `json:"expirationDateTime,omitempty"`
	CreatorID          string `json:"creatorId,omitempty"`
}
