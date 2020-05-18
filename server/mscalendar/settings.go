package mscalendar

import (
	"github.com/gorilla/mux"
	"github.com/larkox/mattermost-plugin-utils/panel"
	"github.com/larkox/mattermost-plugin-utils/panel/settings"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

type Settings interface {
	PrintSettings(userID string)
	ClearSettingsPosts(userID string)
}

func (c *mscalendar) PrintSettings(userID string) {
	c.SettingsPanel.Print(userID)
}

func (c *mscalendar) ClearSettingsPosts(userID string) {
	err := c.SettingsPanel.Clear(userID)
	if err != nil {
		c.Logger.Warnf("Error clearing settings posts. err=%v", err)
	}
}

func NewSettingsPanel(
	bot Bot,
	pluginStore store.Store,
	settingsHandler string,
	pluginURL string,
	getTimezone func(userID string) (string, error),
	r *mux.Router,
) panel.Panel {
	settingList := []settings.Setting{}
	settingList = append(settingList, settings.NewBoolSetting(
		store.UpdateStatusSettingID,
		"Update Status",
		"Do you want to update your status on Mattermost when you are in a meeting?",
		"",
		pluginStore,
	))
	settingList = append(settingList, settings.NewBoolSetting(
		store.GetConfirmationSettingID,
		"Get Confirmation",
		"Do you want to get a confirmation before automatically updating your status?",
		store.UpdateStatusSettingID,
		pluginStore,
	))
	settingList = append(settingList, settings.NewBoolSetting(
		store.ReceiveRemindersSettingID,
		"Receive Reminders",
		"Do you want to receive reminders for upcoming events?",
		"",
		pluginStore,
	))
	settingList = append(settingList, NewDailySummarySetting(
		pluginStore,
		getTimezone,
	))
	settingList = append(settingList, settings.NewFreetextSetting(
		store.TestSettingID,
		"Free text test",
		"This is the setting description. It can also include information about the validation rules. In this case, the string must be longer than 3 characters.",
		"Write what you want to store in this test value.",
		"",
		pluginStore,
		config.PathFreeTextHandler,
		pluginURL,
		pluginStore,
		func(in string) string {
			if len(in) < 3 {
				return "The string must be longer than 3 characters."
			}
			return ""
		},
		r,
		bot,
	))
	return panel.NewSettingsPanel(settingList, bot, bot, pluginStore, settingsHandler, pluginURL)
}
