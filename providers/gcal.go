package providers

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/gcal"
)

const (
	ProviderGCal            = gcal.Kind
	ProviderGCalDisplayName = "Google Calendar"
	ProviderGCalRepository  = ProviderMSCalendarRepository
)

func GetGcalProviderConfig() *config.ProviderConfig {
	return &config.ProviderConfig{
		Name:        ProviderGCal,
		DisplayName: ProviderGCalDisplayName,
		Repository:  ProviderGCalRepository,

		CommandTrigger: ProviderGCal,

		TelemetryShortName: ProviderGCal,

		BotUsername:    ProviderGCal,
		BotDisplayName: ProviderGCalDisplayName,
		Features: config.ProviderFeatures{
			EncryptedStore:     true,
			EventNotifications: false,
		},
	}
}
