package config

// StoredConfig represents the data stored in and managed with the Mattermost
// config.
type StoredConfig struct {
	OAuth2Authority    string
	OAuth2ClientId     string
	OAuth2ClientSecret string

	// Bot username
	BotUserName string `json:"username"`

	// AdminUserIDs contains a comma-separated list of user IDs that are allowed
	// to administer plugin functions, even if not Mattermost sysadmins.
	AdminUserIDs string
}

// Config represents the the metadata handed to all request runners (command,
// http).
type Config struct {
	StoredConfig

	BuildHash      string
	BuildHashShort string
	BuildDate      string

	MattermostSiteHostname string
	MattermostSiteURL      string
	PluginId               string
	PluginURL              string
	PluginURLPath          string
	PluginVersion          string

	BotIconURL string
	BotUserId  string
}
