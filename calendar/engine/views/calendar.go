// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package views

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
)

const (
	TimeFormat12Hour = time.Kitchen
	TimeFormat24Hour = "15:04"
	TimeFormat       = TimeFormat12Hour
)

type Option interface {
	Apply(remote.Event, *model.MessageAttachment)
}

type showTimezoneOption struct {
	timezone   string
	timeFormat string
}

func (tzOpt showTimezoneOption) Apply(event remote.Event, attachment *model.MessageAttachment) {
	tf := TimeFormat12Hour
	if tzOpt.timeFormat != "" {
		tf = tzOpt.timeFormat
	}

	attachment.Text = fmt.Sprintf(
		"%s - %s (%s)",
		event.Start.In(tzOpt.timezone).Time().Format(tf),
		event.End.In(tzOpt.timezone).Time().Format(tf),
		tzOpt.timezone,
	)
	attachment.Fallback = fmt.Sprintf(
		"%s\n%s - %s (%s)",
		attachment.Title,
		event.Start.In(tzOpt.timezone).Time().Format(tf),
		event.End.In(tzOpt.timezone).Time().Format(tf),
		tzOpt.timezone,
	)
}

func ShowTimezoneOption(timezone string, timeFormat ...string) Option {
	if timezone == "" {
		timezone = "UTC"
	}

	tf := ""
	if len(timeFormat) > 0 {
		tf = timeFormat[0]
	}

	return showTimezoneOption{
		timezone:   timezone,
		timeFormat: tf,
	}
}

type timeFormatOption struct {
	timeFormat string
	timezone   string
}

func (opt timeFormatOption) Apply(event remote.Event, attachment *model.MessageAttachment) {
	if opt.timeFormat == "" {
		return
	}

	start := event.Start
	end := event.End
	if opt.timezone != "" {
		start = start.In(opt.timezone)
		end = end.In(opt.timezone)
	}

	attachment.Text = fmt.Sprintf("%s - %s", start.Time().Format(opt.timeFormat), end.Time().Format(opt.timeFormat))
	attachment.Fallback = fmt.Sprintf("%s\n%s - %s", attachment.Title, start.Time().Format(opt.timeFormat), end.Time().Format(opt.timeFormat))
}

func TimeFormatOption(timeFormat string, timezone ...string) Option {
	tz := ""
	if len(timezone) > 0 {
		tz = timezone[0]
	}

	return timeFormatOption{
		timeFormat: timeFormat,
		timezone:   tz,
	}
}

func RenderCalendarView(events []*remote.Event, timeZone string, timeFormat ...string) (string, error) {
	if len(events) == 0 {
		return "You have no upcoming events.", nil
	}

	tf := TimeFormat12Hour
	if len(timeFormat) > 0 && timeFormat[0] != "" {
		tf = timeFormat[0]
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
			eventString, err := renderEvent(e, true, timeZone, tf)
			if err != nil {
				return "", err
			}
			resp += fmt.Sprintf("\n%s", eventString)
		}
	}

	return resp, nil
}

func RenderDaySummary(events []*remote.Event, timezone string, timeFormat ...string) (string, []*model.MessageAttachment, error) {
	if len(events) == 0 {
		return "You have no events for that day", nil, nil
	}

	tf := TimeFormat12Hour
	if len(timeFormat) > 0 && timeFormat[0] != "" {
		tf = timeFormat[0]
	}

	if timezone != "" {
		for _, e := range events {
			e.Start = e.Start.In(timezone)
			e.End = e.End.In(timezone)
		}
	}

	message := fmt.Sprintf("Agenda for %s.\nTimes are shown in %s", events[0].Start.Time().Format("Monday, 02 January"), events[0].Start.TimeZone)

	var attachments []*model.MessageAttachment
	for _, event := range events {
		var actions []*model.PostAction

		fields := []*model.MessageAttachmentField{}
		if event.Location != nil && event.Location.DisplayName != "" {
			fields = append(fields, &model.MessageAttachmentField{
				Title: "Location",
				Value: MarkdownToHTMLEntities(event.Location.DisplayName),
				Short: true,
			})
		}

		attachments = append(attachments, &model.MessageAttachment{
			Title: event.Subject,
			// Text:    event.BodyPreview,
			Text:    fmt.Sprintf("(%s - %s)", event.Start.In(timezone).Time().Format(tf), event.End.In(timezone).Time().Format(tf)),
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

// MarkdownToHTMLEntities converts reserved Markdown characters to their HTML entity equivalents
func MarkdownToHTMLEntities(input string) string {
	replacements := map[rune]string{
		'!':  "&#33;",  // Exclamation Mark
		'#':  "&#35;",  // Hash
		'(':  "&#40;",  // Left Parenthesis
		')':  "&#41;",  // Right Parenthesis
		'*':  "&#42;",  // Asterisk
		'+':  "&#43;",  // Plus Sign
		'-':  "&#45;",  // Dash
		'.':  "&#46;",  // Period
		'/':  "&#47;",  // Forward slash
		':':  "&#58;",  // Colon
		'<':  "&#60;",  // Less than
		'>':  "&#62;",  // Greater Than
		'[':  "&#91;",  // Left Square Bracket
		'\\': "&#92;",  // Back slash
		']':  "&#93;",  // Right Square Bracket
		'_':  "&#95;",  // Underscore
		'`':  "&#96;",  // Backtick
		'|':  "&#124;", // Vertical Bar
		'~':  "&#126;", // Tilde
	}

	var builder strings.Builder
	for _, char := range input {
		if replacement, exists := replacements[char]; exists {
			builder.WriteString(replacement)
		} else {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func renderEvent(event *remote.Event, asRow bool, timeZone string, timeFormat ...string) (string, error) {
	link, err := url.QueryUnescape(event.Weblink)
	if err != nil {
		return "", err
	}

	tf := TimeFormat12Hour
	if len(timeFormat) > 0 && timeFormat[0] != "" {
		tf = timeFormat[0]
	}

	subject := EnsureSubject(event.Subject)

	if event.IsAllDay {
		format := "(All day event) [%s](%s)"
		if asRow {
			format = "| All day event | [%s](%s) |"
		}

		return fmt.Sprintf(format, MarkdownToHTMLEntities(subject), link), nil
	}

	start := event.Start.In(timeZone).Time().Format(tf)
	end := event.End.In(timeZone).Time().Format(tf)

	format := "(%s - %s) [%s](%s)"
	if asRow {
		format = "| %s - %s | [%s](%s) |"
	}

	return fmt.Sprintf(format, start, end, MarkdownToHTMLEntities(subject), link), nil
}

func RenderEventAsAttachment(event *remote.Event, timezone string, options ...Option) (*model.MessageAttachment, error) {
	var actions []*model.PostAction
	fields := []*model.MessageAttachmentField{}
	var titleLink string

	if event.Location != nil && event.Location.DisplayName != "" {
		fields = append(fields, &model.MessageAttachmentField{
			Title: "Location",
			Value: MarkdownToHTMLEntities(event.Location.DisplayName),
			Short: true,
		})
	}

	if event.Conference != nil {
		// Use conference URL as title link if there's conference data present
		titleLink = event.Conference.URL

		title := "Meeting URL"
		if event.Conference.Application != "" {
			title = event.Conference.Application
		}

		fields = append(fields, &model.MessageAttachmentField{
			Title: title,
			Value: event.Conference.URL,
			Short: true,
		})
	}

	attachment := &model.MessageAttachment{
		Title:     MarkdownToHTMLEntities(event.Subject),
		TitleLink: titleLink,
		Text:      fmt.Sprintf("%s - %s", event.Start.In(timezone).Time().Format(time.Kitchen), event.End.In(timezone).Time().Format(time.Kitchen)),
		Fields:    fields,
		Actions:   actions,
		Fallback:  fmt.Sprintf("%s\n%s - %s", MarkdownToHTMLEntities(event.Subject), event.Start.In(timezone).Time().Format(time.Kitchen), event.End.In(timezone).Time().Format(time.Kitchen)),
	}

	for _, opt := range options {
		opt.Apply(*event, attachment)
	}

	return attachment, nil
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

func RenderUpcomingEvent(event *remote.Event, timeZone string, timeFormat ...string) (string, error) {
	message := "You have an upcoming event:\n"
	eventString, err := renderEvent(event, false, timeZone, timeFormat...)
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

func RenderUpcomingEventAsAttachment(event *remote.Event, timeZone string, options ...Option) (message string, attachment *model.MessageAttachment, err error) {
	message = "Upcoming event:\n"
	attachment, err = RenderEventAsAttachment(event, timeZone, options...)
	return message, attachment, err
}
