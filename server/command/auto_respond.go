package command

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
)

func (c *Command) autoRespond(parameters ...string) (string, bool, error) {
	if len(parameters) == 0 {
		request := model.OpenDialogRequest{
			TriggerId: c.Args.TriggerId,
			URL:       c.Config.PluginURL + config.PathDialog + config.PathAutoRespond,
			Dialog: model.Dialog{
				CallbackId: "",
				Title:      "Auto Respond Message",
				Elements: []model.DialogElement{
					{
						DisplayName: "Autorespond Message",
						Name:        "auto_respond",
						Type:        "text",
						Placeholder: "Enter an autorespond message.",
					},
				},
				SubmitLabel:    "Submit",
				NotifyOnCancel: false,
				State:          "",
			},
		}

		c.MSCalendar.OpenAutoRespondDialog(request)
		return "", false, nil
	}

	autoRespondMessage := strings.Join(parameters, " ")
	err := c.MSCalendar.SetUserAutoRespondMessage(c.Args.UserId, autoRespondMessage)
	if err != nil {
		return "Error setting autorespond message. Your user may not be connected. Connect it with the `/mscalendar connect` command.", false, nil
	}

	return fmt.Sprintf("Autorespond message changed to: '%s'", autoRespondMessage), false, nil
}
