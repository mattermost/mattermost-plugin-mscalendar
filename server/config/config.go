package config

import "github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"

var Provider ProviderConfig

// StoredConfig represents the data stored in and managed with the Mattermost
// config.
type StoredConfig struct {
	OAuth2Authority    string
	OAuth2ClientID     string
	OAuth2ClientSecret string
	bot.Config
	EnableStatusSync   bool
	EnableDailySummary bool

	GoogleDomainVerifyKey string
}

// ProviderConfig represents the specific configuration that changes when building for different
// calendar providers.
type ProviderConfig struct {
	Name               string
	DisplayName        string
	Repository         string
	CommandTrigger     string
	TelemetryShortName string
	BotUsername        string
	BotDisplayName     string
}

// Config represents the the metadata handed to all request runners (command,
// http).
type Config struct {
	PluginID               string
	BuildDate              string
	BuildHash              string
	BuildHashShort         string
	MattermostSiteHostname string
	MattermostSiteURL      string
	PluginURL              string
	PluginURLPath          string
	PluginVersion          string
	StoredConfig
	Provider ProviderConfig
}

func (c *Config) GetNotificationURL() string {
	return c.PluginURL + FullPathEventNotification
}
