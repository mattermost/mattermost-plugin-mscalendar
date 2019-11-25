// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
//  See License for license information.

package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/store"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/fields"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/kvstore"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const maxQueueSize = 1024

const (
	FieldSubject        = "Subject"
	FieldBodyPreview    = "BodyPreview"
	FieldImportance     = "Importance"
	FieldDuration       = "Duration"
	FieldWhen           = "When"
	FieldLocation       = "Location"
	FieldAttendees      = "Attendees"
	FieldOrganizer      = "Organizer"
	FieldResponseStatus = "ResponseStatus"
)

type NotificationHandler interface {
	http.Handler
	Configure(apiConfig Config)
	Quit()
}

type notificationHandler struct {
	Config
	incoming   chan *remote.Notification
	queue      chan *remote.Notification
	queueSize  int
	configChan chan Config
	quit       chan bool
}

func NewNotificationHandler(apiConfig Config) NotificationHandler {
	h := &notificationHandler{
		Config:     apiConfig,
		incoming:   make(chan (*remote.Notification)),
		queue:      make(chan (*remote.Notification), maxQueueSize),
		configChan: make(chan (Config)),
		quit:       make(chan (bool)),
	}
	go h.work()
	return h
}

func (h *notificationHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	notifications := h.Remote.HandleNotification(w, req, h.loadUserSubscription)
	for _, n := range notifications {
		h.incoming <- n
	}
}

func (h *notificationHandler) Configure(apiConfig Config) {
	h.configChan <- apiConfig
}

func (h *notificationHandler) Quit() {
	h.quit <- true
}

func (h *notificationHandler) work() {
	for {
		select {
		case n := <-h.incoming:
			if h.queueSize >= maxQueueSize {
				h.Logger.LogError(
					fmt.Sprintf("Notification queue full (%v), dropped notification", h.queueSize))
				continue
			}
			h.queueSize++
			h.queue <- n

		case n := <-h.queue:
			h.queueSize--
			h.processNotification(n)

		case apiConfig := <-h.configChan:
			h.Config = apiConfig

		case <-h.quit:
			return
		}
	}
}

func (h *notificationHandler) processNotification(n *remote.Notification) {
	message := ""
	changed := false
	prior, err := h.EventStore.LoadUserEvent(n.Subscription.CreatorID, n.Event.ID)
	switch err {
	case kvstore.ErrNotFound:
		changed = true
		message = h.formatNewEventNotification(n)
		prior = &store.Event{}
	case nil:
		changed, message = h.formatUpdatedEventNotification(n, prior)
	default:
		return
	}
	if !changed {
		return
	}

	err = h.Poster.PostDirect(n.SubscriptionCreatorMattermostUserID, message, "")
	if err != nil {
		h.Logger.LogInfo("Failed to post notification message: " + err.Error())
	}

	prior.Remote = n.Event
	err = h.EventStore.StoreUserEvent(n.Subscription.CreatorID, prior)
	if err != nil {
		return
	}

	h.Logger.LogDebug("Processed event notification",
		"MattermostUserID", n.SubscriptionCreatorMattermostUserID,
		"SubsriptionID", n.SubscriptionID,
		"Message", message)
	return
}

func (h *notificationHandler) formatUpdatedEventNotification(n *remote.Notification, prior *store.Event) (bool, string) {
	priorFields := eventFields(prior.Remote)
	newFields := eventFields(n.Event)
	changed, added, updated, deleted := fields.Diff(priorFields, newFields)
	if !changed {
		return false, ""
	}

	message := fmt.Sprintf("Updated event: [%s](%s)\n", n.Event.Subject, n.Event.Weblink)
	for _, k := range added {
		message += fmt.Sprintf("- %s: %s\n", k, newFields[k].Strings())
	}
	for _, k := range updated {
		message += fmt.Sprintf("- %s: ~~%s~~ \u2192 %s\n",
			k, priorFields[k].Strings(), newFields[k].Strings())
	}
	for _, k := range deleted {
		message += fmt.Sprintf("- %s: ~~%s~~\n", k, newFields[k].Strings())
	}

	h.Logger.LogDebug("Processed event notification",
		"MattermostUserID", n.SubscriptionCreatorMattermostUserID,
		"SubsriptionID", n.SubscriptionID,
		"Message", message)
	return true, message
}

func (h *notificationHandler) formatNewEventNotification(n *remote.Notification) string {
	message := ""
	if n.ChangeType == "crated" {
		message = fmt.Sprintf("New event: [%s](%s)\n", n.Event.Subject, n.Event.Weblink)
	} else {
		message = fmt.Sprintf("Previously unseen event: [%s](%s)\n", n.Event.Subject, n.Event.Weblink)
	}
	for k, v := range eventFields(n.Event) {
		message += fmt.Sprintf("- %s: %s\n", k, v.Strings())
	}
	return message
}

func eventFields(e *remote.Event) fields.Fields {
	date := func(dt *remote.DateTime) (time.Time, string) {
		if dt == nil {
			return time.Time{}, "n/a"
		}
		t := dt.Time()
		format := "Monday, January 02"
		if t.Year() != time.Now().Year() {
			format = "Monday, January 02, 2006"
		}
		format += " at " + time.Kitchen
		return t, t.Format(format)
	}

	start, startDate := date(e.Start)
	end, _ := date(e.End)

	minutes := int(end.Sub(start).Round(time.Minute).Minutes())
	hours := int(end.Sub(start).Hours())
	minutes -= int(hours * 60)
	days := int(end.Sub(start).Hours()) / 24
	hours -= days * 24

	dur := ""
	switch {
	case days > 0:
		dur = fmt.Sprintf("%v days", days)

	case e.IsAllDay:
		dur = "all-day"

	default:
		switch hours {
		case 0:
			// ignore
		case 1:
			dur = "one hour"
		default:
			dur = fmt.Sprintf("%v hours", hours)
			if minutes > 0 {
				if dur != "" {
					dur += ", "
				}
				dur += fmt.Sprintf("%v minutes", minutes)
			}
		}
	}

	attendees := []fields.Value{}
	for _, a := range e.Attendees {
		attendees = append(attendees, fields.NewStringValue(
			fmt.Sprintf("[%s](mailto:%s) (%s)",
				a.EmailAddress.Name, a.EmailAddress.Address, a.Status.Response)))
	}

	ff := fields.Fields{
		FieldSubject:     fields.NewStringValue(e.Subject),
		FieldBodyPreview: fields.NewStringValue(e.BodyPreview),
		FieldImportance:  fields.NewStringValue(e.Importance),
		// TODO "if recurring"
		FieldWhen: fields.NewStringValue(startDate),
		FieldOrganizer: fields.NewStringValue(
			fmt.Sprintf("[%s](mailto:%s)",
				e.Organizer.EmailAddress.Name, e.Organizer.EmailAddress.Address)),
		FieldLocation:       fields.NewStringValue(e.Location.DisplayName),
		FieldResponseStatus: fields.NewStringValue(e.ResponseStatus.Response),
		FieldAttendees:      fields.NewMultiValue(attendees...),
	}

	return ff
}

func (h *notificationHandler) loadUserSubscription(subscriptionID string) (*remote.User, *oauth2.Token, string, *remote.Subscription, error) {
	sub, err := h.SubscriptionStore.LoadSubscription(subscriptionID)
	if err != nil {
		return nil, nil, "", nil, err
	}
	creator, err := h.UserStore.LoadUser(sub.MattermostCreatorID)
	if err != nil {
		return nil, nil, "", nil, err
	}
	if sub.Remote.ID != creator.Settings.EventSubscriptionID {
		return nil, nil, "", nil, errors.New("Subscription is orphaned")
	}
	return creator.Remote, creator.OAuth2Token, creator.MattermostUserID, sub.Remote, nil
}
