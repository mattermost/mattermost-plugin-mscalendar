// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

type Notification struct {
	ChangeType                          string
	Event                               *Event
	SubscriptionID                      string
	Subscription                        *Subscription
	SubscriptionCreator                 *User
	SubscriptionCreatorMattermostUserID string
	EntityRawData                       []byte
}
