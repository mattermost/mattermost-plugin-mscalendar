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

func withUser(mscalendar *mscalendar) error {
	if mscalendar.user != nil {
		return nil
	}

	user, err := mscalendar.UserStore.LoadUser(mscalendar.mattermostUserID)
	if err != nil {
		return err
	}

	mscalendar.user = user
	return nil
}
