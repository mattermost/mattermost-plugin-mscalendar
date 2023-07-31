package providers

import "github.com/mattermost/mattermost-plugin-mscalendar/server/config"

// GetProviderConfig returns the appropriate provider configuration based on the defined provider
// name
func GetProviderConfig(providerName string) *config.ProviderConfig {
	if providerName == ProviderGCal {
		return GetGcalProviderConfig()
	}
	return GetMSCalendarProviderConfig()
}
