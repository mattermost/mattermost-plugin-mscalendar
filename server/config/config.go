package config

import "github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"

// StoredConfig represents the data stored in and managed with the Mattermost
// config.
type StoredConfig struct {
	OAuth2Authority    string
	OAuth2ClientID     string
	OAuth2ClientSecret string

	EnableStatusSync   bool
	EnableDailySummary bool

	bot.Config
}

// Config represents the the metadata handed to all request runners (command,
// http).
type Config struct {
	StoredConfig

	BuildDate              string
	BuildHash              string
	BuildHashShort         string
	MattermostSiteHostname string
	MattermostSiteURL      string
	PluginID               string
	PluginURL              string
	PluginURLPath          string
	PluginVersion          string
}
