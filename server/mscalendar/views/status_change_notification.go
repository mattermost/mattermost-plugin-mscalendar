package views

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-server/v5/model"
)

var prettyStatuses = map[string]string{
	model.STATUS_ONLINE:  "Online",
	model.STATUS_AWAY:    "Away",
	model.STATUS_DND:     "Do Not Disturb",
	model.STATUS_OFFLINE: "Offline",
}

func RenderStatusChangeNotificationView(events []*remote.Event, status, url string) *model.SlackAttachment {
	for _, e := range events {
		if e.Start.Time().After(time.Now()) {
			return statusChangeAttachments(e, status, url)
		}
	}
	return statusChangeAttachments(nil, status, url)
}

func renderScheduleItem(event *remote.Event, status string) string {
	if event == nil {
		return fmt.Sprintf("You have no upcoming events.\n Shall I change your status to %s?", prettyStatuses[status])
	}

	resp := fmt.Sprintf("Your event with subject `%s` will start soon.", event.Subject)
	if event.Subject == "" {
		resp = "An event with no subject will start soon."
	}

	resp += fmt.Sprintf("\nShall I change your status to %s?", prettyStatuses[status])
	return resp
}

func statusChangeAttachments(event *remote.Event, status, url string) *model.SlackAttachment {
	actionYes := &model.PostAction{
		Name: "Yes",
		Integration: &model.PostActionIntegration{
			URL: url,
			Context: map[string]interface{}{
				"value":     true,
				"change_to": status,
			},
		},
	}

	actionNo := &model.PostAction{
		Name: "No",
		Integration: &model.PostActionIntegration{
			URL: url,
			Context: map[string]interface{}{
				"value": false,
			},
		},
	}

	sa := &model.SlackAttachment{
		Title:   "Status change",
		Text:    renderScheduleItem(event, status),
		Actions: []*model.PostAction{actionYes, actionNo},
	}

	return sa
}
