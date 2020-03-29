// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
//  See License for license information.

package mscalendar

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/fields"
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

const (
	OptionYes          = "Yes"
	OptionNotResponded = "Not responded"
	OptionNo           = "No"
	OptionMaybe        = "Maybe"
)

type NotificationProcessor interface {
	Configure(Env)
	Enqueue(notifications ...*remote.Notification)
	Quit()
}

type notificationProcessor struct {
	Env
	envChan chan Env

	incoming  chan *remote.Notification
	queue     chan *remote.Notification
	queueSize int
	quit      chan bool
}

func NewNotificationProcessor(env Env) NotificationProcessor {
	processor := &notificationProcessor{
		Env:      env,
		envChan:  make(chan (Env)),
		incoming: make(chan (*remote.Notification)),
		queue:    make(chan (*remote.Notification), maxQueueSize),
		quit:     make(chan (bool)),
	}
	go processor.work()
	return processor
}

func (processor *notificationProcessor) Enqueue(notifications ...*remote.Notification) {
	for _, n := range notifications {
		processor.incoming <- n
	}
}

func (processor *notificationProcessor) Configure(env Env) {
	processor.envChan <- env
}

func (processor *notificationProcessor) Quit() {
	processor.quit <- true
}

func (processor *notificationProcessor) work() {
	for {
		select {
		case n := <-processor.incoming:
			if processor.queueSize >= maxQueueSize {
				processor.Env.Logger.Errorf("webhook notification: queue full (`%v`), dropped notification", processor.queueSize)
				continue
			}
			processor.queueSize++
			processor.queue <- n

		case n := <-processor.queue:
			processor.queueSize--
			err := processor.processNotification(n)
			if err != nil {
				processor.Logger.With(bot.LogContext{
					"subscriptionID": n.SubscriptionID,
				}).Infof("webhook notification: failed: `%v`.", err)
			}

		case env := <-processor.envChan:
			processor.Env = env

		case <-processor.quit:
			return
		}
	}
}

func (processor *notificationProcessor) processNotification(n *remote.Notification) error {
	sub, err := processor.Store.LoadSubscription(n.SubscriptionID)
	if err != nil {
		return err
	}
	creator, err := processor.Store.LoadUser(sub.MattermostCreatorID)
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
	if n.RecommendRenew || n.IsBare {
		client = processor.Remote.MakeClient(context.Background(), creator.OAuth2Token)
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
			PluginVersion:       processor.Config.PluginVersion,
		}
		err = processor.Store.StoreUserSubscription(creator, storedSub)
		if err != nil {
			return err
		}
		processor.Logger.With(bot.LogContext{
			"MattermostUserID": creator.MattermostUserID,
			"SubscriptionID":   n.SubscriptionID,
		}).Debugf("webhook notification: renewed user subscription.")
	}

	if n.IsBare {
		n, err = client.GetNotificationData(n)
		if err != nil {
			return err
		}
	}

	var sa *model.SlackAttachment
	prior, err := processor.Store.LoadUserEvent(creator.MattermostUserID, n.Event.ID)
	if err != nil && err != store.ErrNotFound {
		return err
	}
	if prior != nil {
		var changed bool
		changed, sa = processor.updatedEventSlackAttachment(n, prior.Remote)
		if !changed {
			processor.Logger.With(bot.LogContext{
				"MattermostUserID": creator.MattermostUserID,
				"SubscriptionID":   n.SubscriptionID,
				"ChangeType":       n.ChangeType,
				"EventID":          n.Event.ID,
			}).Debugf("webhook notification: no changes detected in event.")
			return nil
		}
	} else {
		sa = processor.newEventSlackAttachment(n)
		prior = &store.Event{}
	}

	err = processor.Poster.DMWithAttachments(creator.MattermostUserID, sa)
	if err != nil {
		return err
	}

	prior.Remote = n.Event
	err = processor.Store.StoreUserEvent(creator.MattermostUserID, prior)
	if err != nil {
		return err
	}

	processor.Logger.With(bot.LogContext{
		"MattermostUserID": creator.MattermostUserID,
		"SubscriptionID":   n.SubscriptionID,
	}).Debugf("Notified: %s.", sa.Title)

	return nil
}

func (processor *notificationProcessor) newSlackAttachment(n *remote.Notification) *model.SlackAttachment {
	return &model.SlackAttachment{
		AuthorName: n.Event.Organizer.EmailAddress.Name,
		AuthorLink: "mailto:" + n.Event.Organizer.EmailAddress.Address,
		TitleLink:  n.Event.Weblink,
		Title:      n.Event.Subject,
		Text:       n.Event.BodyPreview,
	}
}

func (processor *notificationProcessor) newEventSlackAttachment(n *remote.Notification) *model.SlackAttachment {
	sa := processor.newSlackAttachment(n)
	sa.Title = "(new) " + sa.Title

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

	processor.addPostActionSelect(sa, n.Event)
	return sa
}

func (processor *notificationProcessor) updatedEventSlackAttachment(n *remote.Notification, prior *remote.Event) (bool, *model.SlackAttachment) {
	sa := processor.newSlackAttachment(n)
	sa.Title = "(updated) " + sa.Title

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

	processor.addPostActionSelect(sa, n.Event)
	return true, sa
}

func (processor *notificationProcessor) actionURL(action string) string {
	return fmt.Sprintf("%s/%s/%s", processor.Config.PluginURLPath, config.PathPostAction, action)
}

func (processor *notificationProcessor) addPostActions(sa *model.SlackAttachment, event *remote.Event) {
	if !event.ResponseRequested {
		return
	}
	context := map[string]interface{}{
		config.EventIDKey: event.ID,
	}
	sa.Actions = []*model.PostAction{
		{
			Name: "Accept",
			Type: model.POST_ACTION_TYPE_BUTTON,
			Integration: &model.PostActionIntegration{
				URL:     processor.actionURL(config.PathAccept),
				Context: context,
			},
		},
		{
			Name: "Tentatively Accept",
			Type: model.POST_ACTION_TYPE_BUTTON,
			Integration: &model.PostActionIntegration{
				URL:     processor.actionURL(config.PathTentative),
				Context: context,
			},
		},
		{
			Name: "Decline",
			Type: model.POST_ACTION_TYPE_BUTTON,
			Integration: &model.PostActionIntegration{
				URL:     processor.actionURL(config.PathDecline),
				Context: context,
			},
		},
	}
}

func (processor *notificationProcessor) addPostActionSelect(sa *model.SlackAttachment, event *remote.Event) {
	if !event.ResponseRequested {
		return
	}

	context := map[string]interface{}{
		config.EventIDKey: event.ID,
	}

	pa := &model.PostAction{
		Name: "Response",
		Type: model.POST_ACTION_TYPE_SELECT,
		Integration: &model.PostActionIntegration{
			URL:     processor.actionURL(config.PathRespond),
			Context: context,
		},
	}

	for _, o := range []string{OptionNotResponded, OptionYes, OptionNo, OptionMaybe} {
		pa.Options = append(pa.Options, &model.PostActionOptions{Text: o, Value: o})
	}
	switch event.ResponseStatus.Response {
	case "notResponded":
		pa.DefaultOption = OptionNotResponded
	case "accepted":
		pa.DefaultOption = OptionYes
	case "declined":
		pa.DefaultOption = OptionNo
	case "tentativelyAccepted":
		pa.DefaultOption = OptionMaybe
	}

	sa.Actions = []*model.PostAction{pa}
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
