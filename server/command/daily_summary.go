package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
)

func (c *Command) dailySummary(parameters ...string) (string, error) {
	if len(parameters) == 0 {
		dsum, err := c.MSCalendar.GetDailySummarySettingsForUser(c.user())
		if err != nil {
			return fmt.Sprintf("Error: %s", err.Error()), nil
		}

		return utils.JSONBlock(dsum), nil
	}

	switch parameters[0] {
	case "enable":
		dsum, err := c.MSCalendar.SetDailySummaryEnabled(c.user(), true)
		if err != nil {
			return fmt.Sprintf("Failed to enable daily summary. %s", err.Error()), nil
		}

		return c.dailySummarySuccess(dsum), nil
	case "disable":
		dsum, err := c.MSCalendar.SetDailySummaryEnabled(c.user(), false)
		if err != nil {
			return fmt.Sprintf("Failed to disable daily summary. %s", err.Error()), nil
		}
		return c.dailySummarySuccess(dsum), nil
	case "time":
		if len(parameters) != 2 {
			return "Invalid args", nil
		}
		val := parameters[1]

		dsum, err := c.MSCalendar.SetDailySummaryPostTime(c.user(), val)
		if err != nil {
			return "Failed to set daily summary post time: " + err.Error(), nil
		}

		return c.dailySummarySuccess(dsum), nil
	default:
		return "Invalid args", nil
	}
}

func (c *Command) dailySummarySuccess(dsum *store.DailySummarySettings) string {
	if !dsum.Enable {
		return "You will not receive your daily summary."
	}

	if dsum.PostTime == "" {
		return "Please set the daily summary time using `/" + config.CommandTrigger + " summary time 8:00AM` for example."
	}
	return fmt.Sprintf("You will receive your daily summary at %s %s.", dsum.PostTime, dsum.Timezone)
}
