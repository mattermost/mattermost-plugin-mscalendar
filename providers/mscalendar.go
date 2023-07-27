package providers

import "github.com/mattermost/mattermost-plugin-mscalendar/server/config"

const (
	ProviderMSCalendar            = "mscalendar"
	ProviderMSCalendarDisplayName = "Microsoft Calendar"
	ProviderMSCalendarRepository  = ""
)

func GetMSCalendarProviderConfig() *config.ProviderConfig {
	return &config.ProviderConfig{
		Name:        ProviderMSCalendar,
		DisplayName: ProviderMSCalendarDisplayName,
		Repository:  ProviderMSCalendarRepository,

		CommandTrigger: ProviderMSCalendar,

		TelemetryShortName: ProviderMSCalendar,

		BotUsername:    ProviderMSCalendar,
		BotDisplayName: ProviderMSCalendarDisplayName,
	}
}
