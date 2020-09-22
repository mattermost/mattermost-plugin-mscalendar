package views

import (
	"fmt"
	"net/url"
	"sort"
	"time"

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
		resp += "\n" + group[0].Start.Time().Format("Monday January 02") + "\n\n"
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

func renderTableHeader() string {
	return `| Time | Subject |
| :--: | :-- |`
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
