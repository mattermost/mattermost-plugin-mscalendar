package config

import (
	"github.com/mattermost/mattermost-plugin-msoffice/server/kvstore"
	"github.com/mattermost/mattermost-server/plugin"
)

// StoredConfig represents the data stored in and managed with the Mattermost
// config.
type StoredConfig struct {
	// Bot username
	BotUserName string `json:"username"`
}

// ImportedAPI represents the interfaces that are used at request time,
// configured at OnActivate, or overrideable in the tests.
type ImportedAPI struct {
	KVStore kvstore.KVStore
	Helpers plugin.Helpers
	PAPI    plugin.API

	IsAuthorizedAdmin func(string) (bool, error)
}

// Config represents the the metadata handed to all request runners (command,
// http).
type Config struct {
	StoredConfig
	ImportedAPI

	BuildHash      string
	BuildHashShort string
	BuildDate      string

	MattermostSiteURL string
	PluginId          string
	PluginURL         string
	PluginURLPath     string
	PluginVersion     string

	BotIconURL string
	BotUserId  string
}
