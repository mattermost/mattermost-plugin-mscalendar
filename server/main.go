package main

import (
	mattermost "github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	msoffice "github.com/mattermost/mattermost-plugin-msoffice/server/plugin"
)

var BuildHash string
var BuildHashShort string
var BuildDate string

func main() {
	mattermost.ClientMain(
		msoffice.NewWithConfig(
			&config.Config{
				PluginID:       manifest.ID,
				PluginVersion:  manifest.Version,
				BuildHash:      BuildHash,
				BuildHashShort: BuildHashShort,
				BuildDate:      BuildHash,
			}))
}
