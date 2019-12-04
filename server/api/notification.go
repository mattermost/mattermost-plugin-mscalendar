// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
//  See License for license information.

package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/store"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/fields"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/kvstore"
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
	notifications := h.Remote.HandleWebhook(w, req)
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
			err := h.processNotification(n)
			if err != nil {
				h.Logger.LogInfo("Failed to process notification: "+err.Error(),
					"SubsriptionID", n.SubscriptionID)
			}

		case apiConfig := <-h.configChan:
			h.Config = apiConfig

		case <-h.quit:
			return
		}
	}
}

func (h *notificationHandler) processNotification(n *remote.Notification) error {
	sub, err := h.SubscriptionStore.LoadSubscription(n.SubscriptionID)
	if err != nil {
		return err
	}
	creator, err := h.UserStore.LoadUser(sub.MattermostCreatorID)
	if err != nil {
		return err
	}
	if sub.Remote.ID != creator.Settings.EventSubscriptionID {
		return errors.New("Subscription is orphaned")
	}
	if sub.Remote.ClientState != "" && sub.Remote.ClientState != n.ClientState {
		return errors.New("Unauthorized webhook")
	}

	n.Subscription = sub.Remote
	n.SubscriptionCreator = creator.Remote

	var client remote.Client
	if !n.RecommendRenew || n.IsBare {
		client = h.Remote.NewClient(context.Background(), creator.OAuth2Token)
	}

	if n.RecommendRenew {
		var renewed *remote.Subscription
		renewed, err = client.RenewSubscription(n.SubscriptionID)
		if err != nil {
			return err
		}

		storedSub := &store.Subscription{
			Remote:              renewed,
			MattermostCreatorID: creator.MattermostUserID,
			PluginVersion:       h.Config.PluginVersion,
		}
		err = h.SubscriptionStore.StoreUserSubscription(creator, storedSub)
		if err != nil {
			return err
		}
		h.Logger.LogDebug("Renewed user subscription",
			"MattermostUserID", creator.MattermostUserID,
			"SubsriptionID", n.SubscriptionID)
	}

	if n.IsBare {
		n, err = client.GetNotificationData(n)
		if err != nil {
			return err
		}
	}

	var sa *model.SlackAttachment
	prior, err := h.EventStore.LoadUserEvent(creator.MattermostUserID, n.Event.ID)
	if err != nil && err != kvstore.ErrNotFound {
		return err
	}
	if prior != nil {
		var changed bool
		changed, sa = updatedEventSlackAttachment(n, prior.Remote)
		if !changed {
			h.Logger.LogDebug("No changes detected in event",
				"MattermostUserID", creator.MattermostUserID,
				"SubsriptionID", n.SubscriptionID)
			return nil
		}
	} else {
		sa = newEventSlackAttachment(n)
		prior = &store.Event{}
	}

	err = h.Poster.PostDirectAttachments(creator.MattermostUserID, sa)
	if err != nil {
		return err
	}

	prior.Remote = n.Event
	err = h.EventStore.StoreUserEvent(creator.MattermostUserID, prior)
	if err != nil {
		return err
	}

	h.Logger.LogDebug("Processed notification: "+sa.Title,
		"MattermostUserID", creator.MattermostUserID,
		"SubsriptionID", n.SubscriptionID)
	return nil
}

func newEventSlackAttachment(n *remote.Notification) *model.SlackAttachment {
	sa := &model.SlackAttachment{
		AuthorName: n.Event.Organizer.EmailAddress.Name,
		AuthorLink: "mailto:" + n.Event.Organizer.EmailAddress.Address,
		Title:      "New event: " + n.Event.Subject,
		TitleLink:  n.Event.Weblink,
		Text:       n.Event.BodyPreview,
	}

	for n, v := range eventToFields(n.Event) {
		// skip some fields
		switch n {
		case FieldBodyPreview, FieldSubject, FieldOrganizer, FieldResponseStatus:
			continue
		}

		sa.Fields = append(sa.Fields, &model.SlackAttachmentField{
			Title: n,
			Value: fmt.Sprintf("%s", v.Strings()),
			Short: true,
		})
	}

	return sa
}

func updatedEventSlackAttachment(n *remote.Notification, prior *remote.Event) (bool, *model.SlackAttachment) {
	sa := &model.SlackAttachment{
		AuthorName: n.Event.Organizer.EmailAddress.Name,
		AuthorLink: "mailto:" + n.Event.Organizer.EmailAddress.Address,
		Title:      "Updated: " + n.Event.Subject,
		TitleLink:  n.Event.Weblink,
		Text:       n.Event.BodyPreview,
	}

	newFields := eventToFields(n.Event)
	priorFields := eventToFields(prior)
	changed, added, updated, deleted := fields.Diff(priorFields, newFields)
	if !changed {
		return false, nil
	}

	for _, k := range added {
		sa.Fields = append(sa.Fields, &model.SlackAttachmentField{
			Title: k,
			Value: newFields[k].Strings(),
			Short: true,
		})
	}
	for _, k := range updated {
		sa.Fields = append(sa.Fields, &model.SlackAttachmentField{
			Title: k,
			Value: fmt.Sprintf("~~%s~~ \u2192 %s", priorFields[k].Strings(), newFields[k].Strings()),
			Short: true,
		})
	}
	for _, k := range deleted {
		sa.Fields = append(sa.Fields, &model.SlackAttachmentField{
			Title: k,
			Value: fmt.Sprintf("~~%s~~", priorFields[k].Strings()),
			Short: true,
		})
	}

	return true, sa
}

func eventToFields(e *remote.Event) fields.Fields {
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
		}
		if minutes > 0 {
			if dur != "" {
				dur += ", "
			}
			dur += fmt.Sprintf("%v minutes", minutes)
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
		FieldWhen:        fields.NewStringValue(startDate),
		FieldDuration:    fields.NewStringValue(dur),
		FieldOrganizer: fields.NewStringValue(
			fmt.Sprintf("[%s](mailto:%s)",
				e.Organizer.EmailAddress.Name, e.Organizer.EmailAddress.Address)),
		FieldLocation:       fields.NewStringValue(e.Location.DisplayName),
		FieldResponseStatus: fields.NewStringValue(e.ResponseStatus.Response),
		FieldAttendees:      fields.NewMultiValue(attendees...),
	}

	return ff
}
