// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

func (c *client) DeleteCalendarByID(calendarID string) error {
	// TODO: Implement

	err := c.rbuilder.Me().Calendars().ID(calendarID).Request().Delete(c.ctx)
	if err != nil {
		return err
	}

	return nil
}
