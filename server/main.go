package main

import (
	mattermostplugin "github.com/mattermost/mattermost-server/v6/plugin"

	_ "time/tzdata" // Import tzdata so we have it available in slim environments where tzdata package may not be present

	"github.com/mattermost/mattermost-plugin-mscalendar/providers"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/plugin"
)

var BuildHash string
var BuildHashShort string
var BuildDate string
var CalendarProvider string

func main() {
	config.Provider = *providers.GetProviderConfig(CalendarProvider)

	mattermostplugin.ClientMain(
		plugin.NewWithEnv(
			mscalendar.Env{
				Config: &config.Config{
					PluginID:       manifest.ID,
					PluginVersion:  manifest.Version,
					BuildHash:      BuildHash,
					BuildHashShort: BuildHashShort,
					BuildDate:      BuildDate,
					Provider:       config.Provider,
				},
				Dependencies: &mscalendar.Dependencies{},
			}))
}
