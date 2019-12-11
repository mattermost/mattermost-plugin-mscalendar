// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

func (c *client) DeleteCalendar(calID string) error {
	err := c.rbuilder.Me().Calendars().ID(calID).Request().Delete(c.ctx)
	if err != nil {
		return err
	}
	return nil
}
