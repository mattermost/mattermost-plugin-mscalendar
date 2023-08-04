package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
)

const (
	createEventDateTimeFormat = "2006-01-02 15:04"
	createEventDateFormat     = "2006-01-02"
)

type createEventPayload struct {
	AllDay    bool     `json:"all_day"`
	Attendees []string `json:"attendees"`
	Date      string   `json:"date"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	// Reminder  bool     `json:"reminder"
	Description string `json:"description,omitempty"`
	Subject     string `json:"subject"`
	Location    string `json:"location,omitempty"`
	ChannelID   string `json:"channel_id"`
}

func (cep createEventPayload) ToRemoteEvent(loc *time.Location) (*remote.Event, error) {
	var evt remote.Event

	evt.IsAllDay = cep.AllDay

	start, err := cep.parseStartTime(loc)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing start time")
	}

	end, err := cep.parseEndTime(loc)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing start time")
	}

	if !cep.AllDay {
		evt.Start = &remote.DateTime{
			DateTime: start.Format(remote.RFC3339NanoNoTimezone),
			TimeZone: loc.String(),
		}
		evt.End = &remote.DateTime{
			DateTime: end.Format(remote.RFC3339NanoNoTimezone),
			TimeZone: loc.String(),
		}
	} else {
		date, err := cep.parseDate(loc)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing date")
		}

		evt.Start = &remote.DateTime{
			DateTime: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc).Format(remote.RFC3339NanoNoTimezone),
			TimeZone: loc.String(),
		}
		evt.End = &remote.DateTime{
			DateTime: time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 99, loc).Format(remote.RFC3339NanoNoTimezone),
			TimeZone: loc.String(),
		}
	}

	if cep.Description != "" {
		evt.Body = &remote.ItemBody{
			Content:     cep.Description,
			ContentType: "text/plain",
		}
	}
	evt.Subject = cep.Subject
	if cep.Location != "" {
		evt.Location = &remote.Location{
			DisplayName: cep.Location,
		}
	}

	return &evt, nil
}

func (cep createEventPayload) parseStartTime(loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(createEventDateTimeFormat, fmt.Sprintf("%s %s", cep.Date, cep.StartTime), loc)
}

func (cep createEventPayload) parseEndTime(loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(createEventDateTimeFormat, fmt.Sprintf("%s %s", cep.Date, cep.EndTime), loc)
}

func (cep createEventPayload) parseDate(loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(createEventDateFormat, cep.Date, loc)
}

func (cep createEventPayload) IsValid(loc *time.Location) error {
	if cep.Subject == "" {
		return fmt.Errorf("subject must not be empty")
	}

	if cep.Date == "" {
		return fmt.Errorf("date must not be empty")
	}

	_, err := cep.parseDate(loc)
	if err != nil {
		return fmt.Errorf("invalid date")
	}

	if cep.StartTime == "" && cep.EndTime == "" && !cep.AllDay {
		return fmt.Errorf("either start time/end time must be set or event should last all day")
	}

	start, err := cep.parseStartTime(loc)
	if err != nil {
		return fmt.Errorf("please use a valid start time")
	}

	end, err := cep.parseEndTime(loc)
	if err != nil {
		return fmt.Errorf("please use a valid end time")
	}

	if start.After(end) {
		return fmt.Errorf("end date should be after start date")
	}

	return nil
}

func (api *api) createEvent(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-Id")
	if mattermostUserID == "" {
		httputils.WriteUnauthorizedError(w, fmt.Errorf("unauthorized"))
		return
	}

	user, errStore := api.Store.LoadUser(mattermostUserID)
	if errStore != nil && !errors.Is(errStore, store.ErrNotFound) {
		httputils.WriteInternalServerError(w, errStore)
		return
	}
	if errors.Is(errStore, store.ErrNotFound) {
		httputils.WriteUnauthorizedError(w, fmt.Errorf("unauthorized"))
		return
	}

	var payload createEventPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httputils.WriteBadRequestError(w, err)
		return
	}

	if !api.PluginAPI.CanLinkEventToChannel(payload.ChannelID, user.MattermostUserID) {
		httputils.WriteUnauthorizedError(w, fmt.Errorf("you don't have permission to link events in this channel"))
		return
	}

	client := api.Remote.MakeClient(context.Background(), user.OAuth2Token)

	mailbox, errMailbox := client.GetMailboxSettings(user.Remote.ID)
	if errMailbox != nil {
		httputils.WriteInternalServerError(w, errMailbox)
		return
	}

	loc, errLocation := time.LoadLocation(mailbox.TimeZone)
	if errLocation != nil {
		httputils.WriteInternalServerError(w, errLocation)
		return
	}

	if err := payload.IsValid(loc); err != nil {
		httputils.WriteBadRequestError(w, err)
		return
	}

	event, errParse := payload.ToRemoteEvent(loc)
	if errParse != nil {
		httputils.WriteBadRequestError(w, errParse)
		return
	}

	for _, pa := range payload.Attendees {
		var attendee remote.Attendee

		if strings.Contains(pa, "@") {
			attendee.EmailAddress = &remote.EmailAddress{
				Address: pa,
			}
		} else {
			attendeeUser, err := api.Store.LoadUser(pa)
			if err != nil {
				api.Logger.With(bot.LogContext{"attendee_mm_id": pa}).Errorf("error loading attendee from mattermost user id")
				continue
			}

			attendee.RemoteID = attendeeUser.Remote.ID
		}

		event.Attendees = append(event.Attendees, &attendee)
	}

	_, err := client.CreateEvent(user.Remote.ID, event)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSONResponse(w, `{"ok": true}`, http.StatusCreated)
}
