// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package http

import (
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

func (h *Handler) webhookEvent(w http.ResponseWriter, req *http.Request) {
	notifications := h.Remote.HandleEventNotification(w, req, h.loadUserSub)
	for _, n := range notifications {
		isAttendee := false
		for _, a := range n.Event.Attendees {
			if a.EmailAddress.Address == n.SubscriptionCreator.UserPrincipalName {
				isAttendee = true
				break
			}
		}
		isOrganizer := false
		if n.Event.Organizer.EmailAddress.Address == n.SubscriptionCreator.UserPrincipalName {
			isOrganizer = true
		}
		if !isAttendee && !isOrganizer {
			h.Logger.LogInfo("Notification received for an event where user is not mentioned")
			continue
		}

		err := h.Poster.PostDirect(n.SubscriptionCreatorMattermostUserID,
			fmt.Sprintf("%s:%s", n.Change, utils.JSONBlock(n)), "")
		if err != nil {
			h.internalServerError(w, err)
			return
		}
	}
	h.Logger.LogDebug("Webhook received")
	return
}

func (h *Handler) loadUserSub(subID string) (*remote.User, *oauth2.Token, string, *remote.Subscription, error) {
	sub, err := h.SubscriptionStore.LoadSubscription(subID)
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
