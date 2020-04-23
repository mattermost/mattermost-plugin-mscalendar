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

func RenderStatusChangeNotificationView(sched *remote.ScheduleInformation, status, url string) *model.SlackAttachment {
	for _, s := range sched.ScheduleItems {
		if s.Start.Time().After(time.Now()) {
			return statusChangeAttachments(s, status, url)
		}
	}
	return statusChangeAttachments(nil, status, url)
}

func renderScheduleItem(s *remote.ScheduleItem, status string) string {
	if s == nil {
		return fmt.Sprintf("You have no upcoming events.\n Shall I change your status to %s?", prettyStatuses[status])
	}

	resp := fmt.Sprintf("Your event with subject `%s` will start soon.", s.Subject)
	if s.Subject == "" {
		resp = "An event with no subject will start soon."
	}

	resp += fmt.Sprintf("\nShall I change your status to %s?", prettyStatuses[status])
	return resp
}

func statusChangeAttachments(s *remote.ScheduleItem, status, url string) *model.SlackAttachment {
	actionYes := &model.PostAction{
		Name: "Yes",
		Integration: &model.PostActionIntegration{
			URL: url,
			Context: map[string]interface{}{
				"value":            true,
				"change_to":        status,
				"pretty_change_to": prettyStatuses[status],
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
		Text:    renderScheduleItem(s, status),
		Actions: []*model.PostAction{actionYes, actionNo},
	}

	return sa
}
