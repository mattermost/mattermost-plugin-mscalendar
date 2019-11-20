// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"golang.org/x/oauth2"
)

func (api *api) HandleEventNotification(w http.ResponseWriter, req *http.Request) {
	notifications := api.Remote.HandleEventNotification(w, req, api.loadUserSubscription)

	go func() {
		for _, n := range notifications {
			message := api.formatEventNotification(n)
			if message == "" {
				continue
			}
			err := api.Poster.PostDirect(n.SubscriptionCreatorMattermostUserID, message, "")
			if err != nil {
				api.Logger.LogInfo("Failed to post notification message: " + err.Error())
				continue
			}
		}
	}()
}

func (api *api) loadUserSubscription(subscriptionID string) (*remote.User, *oauth2.Token, string, *remote.Subscription, error) {
	sub, err := api.SubscriptionStore.LoadSubscription(subscriptionID)
	if err != nil {
		return nil, nil, "", nil, err
	}
	creator, err := api.UserStore.LoadUser(sub.MattermostCreatorID)
	if err != nil {
		return nil, nil, "", nil, err
	}
	if sub.Remote.ID != creator.Settings.EventSubscriptionID {
		return nil, nil, "", nil, errors.New("Subscription is orphaned")
	}
	return creator.Remote, creator.OAuth2Token, creator.MattermostUserID, sub.Remote, nil
}

func (api *api) formatEventNotification(n *remote.EventNotification) string {
	//TODO: make work with nil Events (deleted)
	// isAttendee := false
	// for _, a := range n.Event.Attendees {
	// 	if a.EmailAddress.Address == n.SubscriptionCreator.UserPrincipalName {
	// 		isAttendee = true
	// 		break
	// 	}
	// }
	// isOrganizer := false
	// if n.Event.Organizer.EmailAddress.Address == n.SubscriptionCreator.UserPrincipalName {
	// 	isOrganizer = true
	// }
	// if !isAttendee && !isOrganizer {
	// 	h.Logger.LogInfo("Notification received for an event where user is not mentioned")
	// 	return ""
	// }

	// TODO translate to MM user handle
	from := ""
	if n.EventMessage != nil {
		from = fmt.Sprintf("[%s](mailto:%s)", n.EventMessage.From.EmailAddress.Name, n.EventMessage.From.EmailAddress.Address)
	} else {
		from = fmt.Sprintf("[%s](mailto:%s)", n.Event.Organizer.EmailAddress.Name, n.Event.Organizer.EmailAddress.Address)
	}

	// TODO use mattermost user's local timezone
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

	start, startDate := date(n.Event.Start)
	end, _ := date(n.Event.End)
	onDate := "on " + startDate

	minutes := int(end.Sub(start).Round(time.Minute).Minutes())
	hours := int(end.Sub(start).Hours())
	minutes -= int(hours * 60)
	days := int(end.Sub(start).Hours()) / 24
	hours -= days * 24

	forDur := ""
	meeting := "meeting"
	if n.Event.IsAllDay {
		meeting = "all-day meeting"
		if days > 0 {
			forDur = fmt.Sprintf("for %v days", days)
		}
	} else {
		switch hours {
		case 0:
			// ignore
		case 1:
			forDur = "for one hour"
		default:
			forDur = fmt.Sprintf("for %v hours", hours)
		}

		if minutes > 0 {
			if forDur != "" {
				forDur += fmt.Sprintf(", %v minutes", minutes)
			} else {
				forDur += fmt.Sprintf("for %v minutes", minutes)
			}
		}
	}

	if n.Event.IsOrganizer {
		meeting += "your "
	}

	act := ""
	switch n.Change {
	case remote.ChangeInvitedMe:
		act = "invited you to a"
	case remote.ChangeAccepted:
		act = "accepted"
	case remote.ChangeTentativelyAccepted:
		act = "tentatively accepted"
	case remote.ChangeDeclined:
		act = "declined"
	case remote.ChangeMeetingCancelled:
		act = "cancelled the"
	case remote.ChangeEventCreated:
		act = "created a"
	case remote.ChangeEventUpdated:
		act = "updated the"
	case remote.ChangeEventDeleted:
		act = "deleted the"
	}

	ss := []string{}
	for _, s := range []string{from, act, meeting, forDur, onDate} {
		if len(s) > 0 {
			ss = append(ss, s)
		}
	}
	headline := strings.Join(ss, " ")

	subject := fmt.Sprintf("[%s](%s)", n.Event.Subject, n.Event.Weblink)
	if !n.Event.IsOrganizer {
		organizer := fmt.Sprintf("[%s](mailto:%s)", n.Event.Organizer.EmailAddress.Name, n.Event.Organizer.EmailAddress.Address)
		subject += fmt.Sprintf(", organized by %s", organizer)
	}

	body := n.Event.BodyPreview
	if n.EventMessage != nil {
		body = n.EventMessage.BodyPreview
	}

	out := fmt.Sprintf("%s\n- Subject: %s\n- Summary: %s\n", headline, subject, body)

	return out
}
