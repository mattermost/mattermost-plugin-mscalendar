package providers

import "github.com/mattermost/mattermost-plugin-mscalendar/server/config"

const (
	ProviderGCal            = "gcal"
	ProviderGCalDisplayName = "Google Calendar"
	ProviderGCalRepository  = ""
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

		EncryptedStore: true,
	}
}
