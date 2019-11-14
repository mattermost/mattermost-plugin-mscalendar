// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

type EventNotification struct {
	SubscriptionID          string
	ChangeType              string
	Event                   *Event
	Subscription            *Subscription
	Creator                 *User
	CreatorMattermostUserID string
}
