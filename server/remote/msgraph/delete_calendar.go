// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import "github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"

func (c *client) DeleteCalendar(calID string) error {
	err := c.rbuilder.Me().Calendars().ID(calID).Request().Delete(c.ctx)
	if err != nil {
		return err
	}
	c.Logger.With(bot.LogContext{}).Infof("msgraph: DeleteCalendar deleted calendar `%v`.", calID)
	return nil
}
