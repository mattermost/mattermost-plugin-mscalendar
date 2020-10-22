// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

type filterf func(*mscalendar) error

func (m *mscalendar) Filter(filters ...filterf) error {
	for _, filter := range filters {
		err := filter(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func withActingUserExpanded(m *mscalendar) error {
	return m.ExpandUser(m.actingUser)
}

func withUserExpanded(user *User) func(m *mscalendar) error {
	return func(m *mscalendar) error {
		return m.ExpandUser(user)
	}
}

func withRemoteUser(user *User) func(m *mscalendar) error {
	return func(m *mscalendar) error {
		return m.ExpandRemoteUser(user)
	}
}

func withClient(m *mscalendar) error {
	if m.client != nil {
		return nil
	}

	client, err := m.MakeClient()

	if err != nil {
		return err
	}

	m.client = client
	return nil
}

func withSuperuserClient(m *mscalendar) error {
	if m.client != nil {
		return nil
	}

	client, err := m.MakeSuperuserClient()
	if err != nil {
		return err
	}

	m.client = client
	return nil
}
