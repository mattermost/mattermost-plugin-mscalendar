// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

type Client interface {
	GetMe() (*User, error)
	GetUserCalendars(userId string) ([]*Calendar, error)
}
