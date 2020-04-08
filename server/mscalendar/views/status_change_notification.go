package views

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-server/v5/model"
)

func RenderStatusChangeNotificationView(events []*remote.ScheduleItem, status, url string) *model.SlackAttachment {
	for _, e := range events {
		if e.Start.Time().After(time.Now()) {
			return statusChangeAttachments(e, status, url)
		}
	}
	return statusChangeAttachments(nil, status, url)
}

func renderScheduleItem(si *remote.ScheduleItem, status string) string {
	if si == nil {
		return fmt.Sprintf("You have no upcoming events.\n Shall I change your status to %s?", status)
	}

	resp := fmt.Sprintf("Your event with subject `%s` will start soon.", si.Subject)
	if si.Subject == "" {
		resp = "An event with no subject will start soon."
	}

	resp += "\nShall I change your status to %s?"
	return resp
}

func statusChangeAttachments(event *remote.ScheduleItem, status, url string) *model.SlackAttachment {
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
