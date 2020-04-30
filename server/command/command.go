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
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

// Handler handles commands
type Command struct {
	Context    *plugin.Context
	Args       *model.CommandArgs
	ChannelID  string
	Config     *config.Config
	MSCalendar mscalendar.MSCalendar
}

func getNotConnectedText() string {
	return fmt.Sprintf("It looks like your Mattermost account is not connected to a Microsoft account. Please connect your account using `/%s connect`.", config.CommandTrigger)
}

// RegisterFunc is a function that allows the runner to register commands with the mattermost server.
type RegisterFunc func(*model.Command) error

type handleFunc func(parameters ...string) (string, bool, error)

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
func (c *Command) Handle() (string, bool, error) {
	cmd, parameters, err := c.isValid()
	if err != nil {
		return "", false, err
	}

	handler := c.help
	switch cmd {
	case "info":
		handler = c.info
	case "connect":
		handler = c.connect
	case "disconnect":
		handler = c.requireConnectedUser(c.disconnect)
	case "summary":
		handler = c.requireConnectedUser(c.dailySummary)
	case "viewcal":
		handler = c.requireConnectedUser(c.viewCalendar)
	case "createcal":
		handler = c.requireConnectedUser(c.createCalendar)
	case "createevent":
		handler = c.requireConnectedUser(c.createEvent)
	case "deletecal":
		handler = c.requireConnectedUser(c.deleteCalendar)
	case "subscribe":
		handler = c.requireConnectedUser(c.subscribe)
	case "unsubscribe":
		handler = c.requireConnectedUser(c.unsubscribe)
	case "findmeetings":
		handler = c.requireConnectedUser(c.findMeetings)
	case "showcals":
		handler = c.requireConnectedUser(c.showCalendars)
	case "availability":
		handler = c.requireConnectedUser(c.debugAvailability)
	case "settings":
		handler = c.requireConnectedUser(c.settings)
	}
	out, mustRedirectToDM, err := handler(parameters...)
	if err != nil {
		return out, false, errors.WithMessagef(err, "Command /%s %s failed", config.CommandTrigger, cmd)
	}

	return out, mustRedirectToDM, nil
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

func (c *Command) requireConnectedUser(handle handleFunc) handleFunc {
	return func(parameters ...string) (string, bool, error) {
		connected, err := c.isConnected()
		if err != nil {
			return "", false, err
		}

		if !connected {
			return getNotConnectedText(), false, nil
		}
		return handle(parameters...)
	}
}

func (c *Command) isConnected() (bool, error) {
	_, err := c.MSCalendar.GetRemoteUser(c.Args.UserId)
	if err == store.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
