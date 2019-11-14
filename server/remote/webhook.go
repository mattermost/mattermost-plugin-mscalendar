// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

type ResourceData struct {
	ID string `json:"id"`
}

type Webhook struct {
	SubscriptionID                 string       `json:"subscriptionId"`
	SubscriptionExpirationDateTime string       `json:"subscriptionExpirationDateTime,omitempty"`
	ChangeType                     string       `json:"changeType"`
	ResourcePath                   string       `json:"resource,omitempty"`
	ResourceData                   ResourceData `json:"resourceData,omitempty"`
	ClientState                    string       `json:"clientState,omitempty"`
}
