// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
)

// Handler handles commands
type Command struct {
	Context    *plugin.Context
	Args       *model.CommandArgs
	ChannelID  string
	Config     *config.Config
	MSCalendar mscalendar.MSCalendar
}

func getHelp() string {
	help := `
TODO: help text.
`
	return utils.CodeBlock(fmt.Sprintf(
		help,
	))
}

// RegisterFunc is a function that allows the runner to register commands with the mattermost server.
type RegisterFunc func(*model.Command) error

// Register should be called by the plugin to register all necessary commands
func Register(registerFunc RegisterFunc) {
	_ = registerFunc(&model.Command{
		Trigger:          config.CommandTrigger,
		DisplayName:      "Microsoft Calendar",
		Description:      "Interact with your outlook calendar.",
		AutoComplete:     true,
		AutoCompleteDesc: "help, info, connect, disconnect, connect_bot, disconnect_bot, subscribe, showcals, viewcal, createcal, deletecal, createevent, findmeetings, availability, summary",
		AutoCompleteHint: "(subcommand)",
	})
}

// Handle should be called by the plugin when a command invocation is received from the Mattermost server.
func (c *Command) Handle() (string, error) {
	cmd, parameters, err := c.isValid()
	if err != nil {
		return "", err
	}

	handler := c.help
	switch cmd {
	case "info":
		handler = c.info
	case "connect":
		handler = c.connect
	case "connect_bot":
		handler = c.connectBot
	case "disconnect":
		handler = c.disconnect
	case "disconnect_bot":
		handler = c.disconnectBot
	case "summary":
		handler = c.dailySummary
	case "viewcal":
		handler = c.viewCalendar
	case "createcal":
		handler = c.createCalendar
	case "createevent":
		handler = c.createEvent
	case "deletecal":
		handler = c.deleteCalendar
	case "subscribe":
		handler = c.subscribe
	case "findmeetings":
		handler = c.findMeetings
	case "showcals":
		handler = c.showCalendars
	case "availability":
		handler = c.availability
	case "settings":
		handler = c.settings
	}
	out, err := handler(parameters...)
	if err != nil {
		return out, errors.WithMessagef(err, "Command /%s %s failed", config.CommandTrigger, cmd)
	}

	return out, nil
}

func (c *Command) isValid() (subcommand string, parameters []string, err error) {
	if c.Context == nil || c.Args == nil {
		return "", nil, errors.New("Invalid arguments to command.Handler")
	}
	split := strings.Fields(c.Args.Command)
	command := split[0]
	if command != "/"+config.CommandTrigger {
		return "", nil, errors.Errorf("%q is not a supported command. Please contact your system administrator.", command)
	}

	parameters = []string{}
	subcommand = ""
	if len(split) > 1 {
		subcommand = split[1]
	}
	if len(split) > 2 {
		parameters = split[2:]
	}

	return subcommand, parameters, nil
}

func (c *Command) user() *mscalendar.User {
	return mscalendar.NewUser(c.Args.UserId)
}
