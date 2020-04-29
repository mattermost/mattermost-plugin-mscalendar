package mscalendar

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/settingspanel"
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
		c.Logger.Warnf("error clearing settings posts, " + err.Error())
	}
}

func NewSettingsPanel(bot bot.Bot, panelStore settingspanel.PanelStore, settingStore settingspanel.SettingStore, settingsHandler, pluginURL string, getTimezone func(userID string) (string, error)) settingspanel.Panel {
	settings := []settingspanel.Setting{}
	settings = append(settings, settingspanel.NewBoolSetting(
		store.UpdateStatusSettingID,
		"Update Status",
		"Do you want to update your status on Mattermost when you are in a meeting?",
		"",
		settingStore,
	))
	settings = append(settings, settingspanel.NewBoolSetting(
		store.GetConfirmationSettingID,
		"Get Confirmation",
		"Do you want to get a confirmation before automatically updating your status?",
		store.UpdateStatusSettingID,
		settingStore,
	))
	settings = append(settings, settingspanel.NewBoolSetting(
		store.ReceiveRemindersSettingID,
		"Receive Reminders",
		"Do you want to receive reminders for upcoming events?",
		"",
		settingStore,
	))
	settings = append(settings, NewDailySummarySetting(
		settingStore,
		getTimezone,
	))
	return settingspanel.NewSettingsPanel(settings, bot, bot, panelStore, settingsHandler, pluginURL)
}
