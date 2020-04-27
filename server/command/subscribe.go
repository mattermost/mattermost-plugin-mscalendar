// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"
)

func (c *Command) subscribe(parameters ...string) (string, error) {
	_, err := c.MSCalendar.LoadMyEventSubscription()
	if err == nil {
		return "Already subscribed to events.", nil
	}

	storedSub, err := c.MSCalendar.CreateMyEventSubscription()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Subscription %s created.", storedSub.Remote.ID), nil
}
