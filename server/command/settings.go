// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

func (c *Command) settings(parameters ...string) (string, error) {
	c.MSCalendar.PrintSettings(c.Args.UserId)
	out := "The bot will show you the settings."
	return out, nil
}
