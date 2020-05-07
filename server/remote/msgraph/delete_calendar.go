// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import "github.com/larkox/mattermost-plugin-utils/bot/logger"

func (c *client) DeleteCalendar(remoteUserID string, calID string) error {
	err := c.rbuilder.Users().ID(remoteUserID).Calendars().ID(calID).Request().Delete(c.ctx)
	if err != nil {
		return err
	}
	c.Logger.With(logger.LogContext{}).Infof("msgraph: DeleteCalendar deleted calendar `%v`.", calID)
	return nil
}
