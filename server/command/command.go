// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"
	"strings"

	pluginapilicense "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/experimental/command"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

// Handler handles commands
type Command struct {
	MSCalendar mscalendar.MSCalendar
	Context    *plugin.Context
	Args       *model.CommandArgs
	Config     *config.Config
	ChannelID  string
}

func getNotConnectedText() string {
	return fmt.Sprintf("It looks like your Mattermost account is not connected to a Microsoft account. Please connect your account using `/%s connect`.", config.CommandTrigger)
}

type handleFunc func(parameters ...string) (string, bool, error)

var cmds = []*model.AutocompleteData{
	model.NewAutocompleteData("connect", "", "Connect to your Microsoft account"),
	model.NewAutocompleteData("disconnect", "", "Disconnect from your Microsoft Account"),
	model.NewAutocompleteData("summary", "", "View your events for today, or edit the settings for your daily summary."),
	model.NewAutocompleteData("viewcal", "", "View your events for the upcoming week."),
	model.NewAutocompleteData("settings", "", "Edit your user personal settings."),
	model.NewAutocompleteData("subscribe", "", "Enable notifications for event invitations and updates."),
	model.NewAutocompleteData("unsubscribe", "", "Disable notifications for event invitations and updates."),
	model.NewAutocompleteData("info", "", "Read information about this version of the plugin."),
	model.NewAutocompleteData("help", "", "Read help text for the commands"),
}

// Register should be called by the plugin to register all necessary commands
func Register(client *pluginapilicense.Client) error {
	names := []string{}
	for _, subCommand := range cmds {
		names = append(names, subCommand.Trigger)
	}

	hint := "[" + strings.Join(names[:4], "|") + "...]"

	cmd := model.NewAutocompleteData(config.CommandTrigger, hint, "Interact with your Outlook calendar.")
	cmd.SubCommands = cmds

	iconData, err := command.GetIconData(&client.System, "assets/profile.svg")
	if err != nil {
		return errors.Wrap(err, "failed to get icon data")
	}

	return client.SlashCommand.Register(&model.Command{
		Trigger:              config.CommandTrigger,
		DisplayName:          "Microsoft Calendar",
		Description:          "Interact with your Outlook calendar.",
		AutoComplete:         true,
		AutoCompleteDesc:     strings.Join(names, ", "),
		AutoCompleteHint:     "(subcommand)",
		AutocompleteData:     cmd,
		AutocompleteIconData: iconData,
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
		handler = c.requireConnectedUser(c.requireAdminUser(c.debugAvailability))
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
		return "", nil, errors.New("invalid arguments to command.Handler")
	}
	split := strings.Fields(c.Args.Command)
	command := split[0]
	if command != "/"+config.CommandTrigger {
		return "", nil, fmt.Errorf("%q is not a supported command. Please contact your system administrator", command)
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

func (c *Command) requireAdminUser(handle handleFunc) handleFunc {
	return func(parameters ...string) (string, bool, error) {
		authorized, err := c.MSCalendar.IsAuthorizedAdmin(c.Args.UserId)
		if err != nil {
			return "", false, err
		}
		if !authorized {
			return "Not authorized", false, nil
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
