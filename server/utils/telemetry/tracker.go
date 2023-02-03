package telemetry

import "github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"

type Tracker interface {
	Track(event string, properties map[string]interface{})
	TrackUserEvent(event string, userID string, properties map[string]interface{})
}

type Client interface {
	Enqueue(t Track) error
	Close() error
}

type Track struct {
	Properties map[string]interface{}
	UserID     string
	Event      string
}

type tracker struct {
	client             Client
	logger             bot.Logger
	serverVersion      string
	pluginID           string
	pluginVersion      string
	telemetryShortName string
	diagnosticID       string
	enabled            bool
}

func NewTracker(c Client, diagnosticID, serverVersion, pluginID, pluginVersion, telemetryShortName string, enableDiagnostics bool, logger bot.Logger) Tracker {
	return &tracker{
		telemetryShortName: telemetryShortName,
		client:             c,
		diagnosticID:       diagnosticID,
		serverVersion:      serverVersion,
		pluginID:           pluginID,
		pluginVersion:      pluginVersion,
		enabled:            enableDiagnostics,
		logger:             logger,
	}
}

func (t *tracker) Track(event string, properties map[string]interface{}) {
	if !t.enabled || t.client == nil {
		return
	}

	event = t.telemetryShortName + "_" + event
	properties["PluginID"] = t.pluginID
	properties["PluginVersion"] = t.pluginVersion
	properties["ServerVersion"] = t.serverVersion

	err := t.client.Enqueue(Track{
		UserID:     t.diagnosticID,
		Event:      event,
		Properties: properties,
	})

	if err != nil {
		t.logger.Warnf("cannot enqueue telemetry event, err=%s", err.Error())
	}
}

func (t *tracker) TrackUserEvent(event string, userID string, properties map[string]interface{}) {
	properties["UserActualID"] = userID
	t.Track(event, properties)
}
