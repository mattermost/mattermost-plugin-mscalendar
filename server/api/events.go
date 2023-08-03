package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
)

const createEventTimeFormat = "2006-01-02 15:04"

type createEventLocationPayload struct {
	DisplayName string `json:"display_name"`
	Street      string `json:"street,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
	Country     string `json:"country,omitempty"`
}

type createEventPayload struct {
	AllDay    bool     `json:"all_day"`
	Attendees []string `json:"attendees"`
	Date      string   `json:"date"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	// Reminder  bool     `json:"reminder"
	Description string                      `json:"description,omitempty"`
	Subject     string                      `json:"subject"`
	Location    *createEventLocationPayload `json:"location,omitempty"`
}

func (cep createEventPayload) ToRemoteEvent() *remote.Event {
	var evt remote.Event

	evt.IsAllDay = cep.AllDay

	if cep.Date != "" {
		evt.Start = &remote.DateTime{
			DateTime: cep.Date,
		}
		evt.End = &remote.DateTime{
			DateTime: cep.Date,
		}
	} else {
		evt.Start = &remote.DateTime{DateTime: cep.StartTime}
		evt.End = &remote.DateTime{DateTime: cep.EndTime}
	}
	if cep.Description != "" {
		evt.Body = &remote.ItemBody{
			Content:     cep.Description,
			ContentType: "text/plain",
		}
	}
	evt.Subject = cep.Subject
	if cep.Location != nil {
		evt.Location = &remote.Location{
			DisplayName: evt.Location.DisplayName,
			Address:     evt.Location.Address,
		}
	}

	return &evt
}

func (cep createEventPayload) parseStartTime(loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(createEventTimeFormat, fmt.Sprintf("%s %s", cep.Date, cep.StartTime), loc)
}

func (cep createEventPayload) parseEndTime(loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(createEventTimeFormat, fmt.Sprintf("%s %s", cep.Date, cep.EndTime), loc)
}

func (cep createEventPayload) IsValid(loc *time.Location) error {
	if cep.Subject == "" {
		return fmt.Errorf("subject must not be empty")
	}

	if cep.Date == "" {
		return fmt.Errorf("date must not be empty")
	}

	if cep.StartTime == "" && cep.EndTime == "" && !cep.AllDay {
		return fmt.Errorf("either start time/end time must be set or event should last all day")
	}

	if _, err := cep.parseStartTime(loc); err != nil {
		return fmt.Errorf("please use a valid start time")
	}

	if _, err := cep.parseEndTime(loc); err != nil {
		return fmt.Errorf("please use a valid end time")
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

	event := payload.ToRemoteEvent()
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

	result, err := client.CreateEvent(user.Remote.ID, event)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSONResponse(w, result, http.StatusCreated)
}
