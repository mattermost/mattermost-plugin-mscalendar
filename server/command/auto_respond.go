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
			URL:       c.Config.PluginURL + config.PathDialogs + config.PathAutoRespondMessage,
			Dialog: model.Dialog{
				CallbackId: "",
				Title:      "Auto-Respond Message",
				Elements: []model.DialogElement{
					{
						DisplayName: "Auto-Respond Message",
						Name:        "auto_respond_message",
						Type:        "text",
						Placeholder: "Enter an auto-respond message.",
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
		return fmt.Sprintf("Failed to set auto-respond message. err=%v", err), false, nil
	}

	return fmt.Sprintf("Auto-respond message changed to: '%s'", autoRespondMessage), false, nil
}
