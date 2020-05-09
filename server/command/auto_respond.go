package command

import (
	"fmt"
	"strings"
)

const autoRespondHelp = "### Autorespond commands:\n" +
	"`/mscalendar autorespond <message>` - Set the autorespond message\n"

func (c *Command) autoRespond(parameters ...string) (string, bool, error) {

	if len(parameters) == 0 {
		return autoRespondHelp, false, nil
	}

	autoRespondMessage := strings.Join(parameters[0:], " ")

	err := c.MSCalendar.SetUserAutoRespondMessage(c.Args.UserId, autoRespondMessage)
	if err != nil {
		return "Error setting autorespond message. Your user may not be connected. Connect it with the `/mscalendar connect` command.", false, nil
	}

	return fmt.Sprintf("Autorespond message changed to: '%s'", autoRespondMessage), false, nil
}
