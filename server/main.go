package main

import (
	mattermostplugin "github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-mscalendar/mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/engine"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/plugin"
)

var BuildHash string
var BuildHashShort string
var BuildDate string
var CalendarProvider string

func main() {
	config.Provider = mscalendar.GetMSCalendarProviderConfig()

	mattermostplugin.ClientMain(
		plugin.NewWithEnv(
			engine.Env{
				Config: &config.Config{
					PluginID:       manifest.ID,
					PluginVersion:  manifest.Version,
					BuildHash:      BuildHash,
					BuildHashShort: BuildHashShort,
					BuildDate:      BuildDate,
					Provider:       config.Provider,
				},
				Dependencies: &engine.Dependencies{},
			}))
}
