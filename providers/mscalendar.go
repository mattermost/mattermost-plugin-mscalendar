package providers

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/msgraph"
)

const (
	ProviderMSCalendar            = msgraph.Kind
	ProviderMSCalendarDisplayName = "Microsoft Calendar"
	ProviderMSCalendarRepository  = "mattermost-plugin-mscalendar"
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

		EncryptedStore: false,
	}
}
