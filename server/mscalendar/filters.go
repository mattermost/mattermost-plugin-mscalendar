// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

type filterf func(*mscalendar) error

func (mscalendar *mscalendar) Filter(filters ...filterf) error {
	for _, filter := range filters {
		err := filter(mscalendar)
		if err != nil {
			return err
		}
	}
	return nil
}

func withActingUserExpanded(mscalendar *mscalendar) error {
	return mscalendar.ExpandUser(mscalendar.actingUser)
}

func withUserExpanded(user *User) func(mscalendar *mscalendar) error {
	return func(mscalendar *mscalendar) error {
		return mscalendar.ExpandUser(user)
	}
}

func withClient(mscalendar *mscalendar) error {
	if mscalendar.client != nil {
		return nil
	}
	client, err := mscalendar.MakeClient()
	if err != nil {
		return err
	}
	mscalendar.client = client
	return nil
}
