package views

import (
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

func RenderCalendarView(events []*remote.Event, timeZone string) (string, error) {
	if len(events) == 0 {
		return "You have no upcoming events.", nil
	}

	if timeZone != "" {
		for _, e := range events {
			e.Start = e.Start.In(timeZone)
			e.End = e.End.In(timeZone)
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Start.Time().Before(events[j].Start.Time())
	})

	resp := "Times are shown in " + events[0].Start.TimeZone
	for _, group := range groupEventsByDate(events) {
		resp += "\n" + group[0].Start.Time().Format("Monday January 02, 2006") + "\n\n"
		resp += renderTableHeader()
		for _, e := range group {
			eventString, err := renderEvent(e, true, timeZone)
			if err != nil {
				return "", err
			}
			resp += fmt.Sprintf("\n%s", eventString)
		}
	}

	return resp, nil
}

func RenderDaySummary(events []*remote.Event, timezone string) (string, []*model.SlackAttachment, error) {
	if len(events) == 0 {
		return "You have no events for that day", nil, nil
	}

	if timezone != "" {
		for _, e := range events {
			e.Start = e.Start.In(timezone)
			e.End = e.End.In(timezone)
		}
	}

	message := fmt.Sprintf("Agenda for %s.\nTimes are shown in %s", events[0].Start.Time().Format("Monday, 02 January"), events[0].Start.TimeZone)

	var attachments []*model.SlackAttachment
	for _, event := range events {
		var actions []*model.PostAction

		fields := []*model.SlackAttachmentField{}
		if event.Location != nil && event.Location.DisplayName != "" {
			fields = append(fields, &model.SlackAttachmentField{
				Title: "Location",
				Value: event.Location.DisplayName,
				Short: true,
			})
		}

		attachments = append(attachments, &model.SlackAttachment{
			Title: event.Subject,
			// Text:    event.BodyPreview,
			Text:    fmt.Sprintf("(%s - %s)", event.Start.In(timezone).Time().Format(time.Kitchen), event.End.In(timezone).Time().Format(time.Kitchen)),
			Fields:  fields,
			Actions: actions,
		})
	}

	return message, attachments, nil
}

func renderTableHeader() string {
	return `| Time | Subject |
| :-- | :-- |`
}

func renderEvent(event *remote.Event, asRow bool, timeZone string) (string, error) {
	start := event.Start.In(timeZone).Time().Format(time.Kitchen)
	end := event.End.In(timeZone).Time().Format(time.Kitchen)

	format := "(%s - %s) [%s](%s)"
	if asRow {
		format = "| %s - %s | [%s](%s) |"
	}

	link, err := url.QueryUnescape(event.Weblink)
	if err != nil {
		return "", err
	}

	subject := EnsureSubject(event.Subject)

	return fmt.Sprintf(format, start, end, subject, link), nil
}

func renderEventAsAttachment(event *remote.Event, timezone string) (*model.SlackAttachment, error) {
	var actions []*model.PostAction
	fields := []*model.SlackAttachmentField{}

	if event.Location != nil && event.Location.DisplayName != "" {
		fields = append(fields, &model.SlackAttachmentField{
			Title: "Location",
			Value: event.Location.DisplayName,
			Short: true,
		})

		// Add actions for known links
		// Disable join meeting button for now, since we don't have a handler and
		// the location url is shown parsed and clickable anyway.
		// if joinMeetingAction := getActionForLocation(event.Location); joinMeetingAction != nil {
		// 	actions = append(actions, joinMeetingAction)
		// }
	}

	return &model.SlackAttachment{
		Title:   event.Subject,
		Text:    fmt.Sprintf("(%s - %s)", event.Start.In(timezone).Time().Format(time.Kitchen), event.End.In(timezone).Time().Format(time.Kitchen)),
		Fields:  fields,
		Actions: actions,
	}, nil
}

func groupEventsByDate(events []*remote.Event) [][]*remote.Event {
	groups := map[string][]*remote.Event{}

	for _, event := range events {
		date := event.Start.Time().Format("2006-01-02")
		_, ok := groups[date]
		if !ok {
			groups[date] = []*remote.Event{}
		}

		groups[date] = append(groups[date], event)
	}

	days := []string{}
	for k := range groups {
		days = append(days, k)
	}
	sort.Strings(days)

	result := [][]*remote.Event{}
	for _, day := range days {
		group := groups[day]
		result = append(result, group)
	}
	return result
}

func RenderUpcomingEvent(event *remote.Event, timeZone string) (string, error) {
	message := "You have an upcoming event:\n"
	eventString, err := renderEvent(event, false, timeZone)
	if err != nil {
		return "", err
	}

	return message + eventString, nil
}

func EnsureSubject(s string) string {
	if s == "" {
		return "(No subject)"
	}

	return s
}

func RenderUpcomingEventAsAttachment(event *remote.Event, timeZone string) (message string, attachment *model.SlackAttachment, err error) {
	message = "You have an upcoming event:\n"
	attachment, err = renderEventAsAttachment(event, timeZone)
	return message, attachment, err
}
