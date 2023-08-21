// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

func (c *Command) create(parameters ...string) (string, bool, error) {
	if len(parameters) == 0 {
		return getDailySummaryHelp(), false, nil
	}

	if parameters[0] == "event" {
		return "Creating events is only supported on desktop.", false, nil
	}

	return "", false, nil
}
