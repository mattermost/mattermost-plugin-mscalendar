// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"
)

func (c *Command) unsubscribe(parameters ...string) (string, error) {
	_, err := c.MSCalendar.LoadMyEventSubscription()
	if err != nil {
		return "You are not subscribed to events.", nil
	}

	err = c.MSCalendar.DeleteMyEventSubscription()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("You have unsubscribed from events."), nil
}
