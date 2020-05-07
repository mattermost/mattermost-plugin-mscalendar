package mscalendar

import (
	"github.com/larkox/mattermost-plugin-utils/panel"
	"github.com/larkox/mattermost-plugin-utils/panel/settings"
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
		c.Logger.Warnf("error clearing settings posts, " + err.Error())
	}
}

func NewSettingsPanel(bot Bot, panelStore panel.PanelStore, settingStore settings.SettingStore, settingsHandler, pluginURL string, getTimezone func(userID string) (string, error)) panel.Panel {
	settingList := []settings.Setting{}
	settingList = append(settingList, settings.NewBoolSetting(
		store.UpdateStatusSettingID,
		"Update Status",
		"Do you want to update your status on Mattermost when you are in a meeting?",
		"",
		settingStore,
	))
	settingList = append(settingList, settings.NewBoolSetting(
		store.GetConfirmationSettingID,
		"Get Confirmation",
		"Do you want to get a confirmation before automatically updating your status?",
		store.UpdateStatusSettingID,
		settingStore,
	))
	settingList = append(settingList, settings.NewBoolSetting(
		store.ReceiveRemindersSettingID,
		"Receive Reminders",
		"Do you want to receive reminders for upcoming events?",
		"",
		settingStore,
	))
	settingList = append(settingList, NewDailySummarySetting(
		settingStore,
		getTimezone,
	))
	return panel.NewSettingsPanel(settingList, bot, bot, panelStore, settingsHandler, pluginURL)
}
