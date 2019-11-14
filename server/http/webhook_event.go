// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/store"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

func (h *Handler) apiEventWebhook(w http.ResponseWriter, req *http.Request) {
	// Microsoft graph requires webhook endpoint validation, see
	// https://docs.microsoft.com/en-us/graph/webhooks#notification-endpoint-validation
	vtok := req.FormValue("validationToken")
	if vtok != "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(vtok))
		h.Logger.LogDebug("Validated event webhook endpoint")
		return
	}

	rawData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		h.internalServerError(w, err)
		return
	}

	// go func() {
	// Get the list of webhooks
	var v struct {
		Value []*remote.Webhook `json:"value"`
	}
	err = json.Unmarshal(rawData, &v)
	if err != nil {
		h.badRequest(w, err)
		return
	}

	// Load and verify subscriptions
	for _, wh := range v.Value {
		var sub *store.Subscription
		sub, err = h.SubscriptionStore.LoadSubscription(wh.SubscriptionID)
		if err != nil {
			h.Logger.LogInfo("Subscription not found: "+err.Error(),
				"subscriptionID", wh.SubscriptionID)
			// TODO try to delete the subscription? But what credentials to use?
			// there is a userID in the notification, can try using those credentials,
			// but doesn't seem right. Maybe a DM to the owner with a "Delete" button?

			// h.notFound(w, err)
			//<><> TODO uncomment - return success for now
			w.WriteHeader(http.StatusAccepted)
			return
		}

		var creator *store.User
		creator, err = h.UserStore.LoadUser(sub.MattermostCreatorID)
		if err != nil {
			h.Logger.LogInfo("Subscription user not found: "+err.Error(),
				"subscriptionID", wh.SubscriptionID,
				"userID", sub.MattermostCreatorID)
			// h.notFound(w, err)
			//<><> TODO uncomment - return success for now
			w.WriteHeader(http.StatusAccepted)
			return
		}

		if sub.Remote.ID != creator.Settings.EventSubscriptionID {
			h.Logger.LogInfo("Subscription is orphaned",
				"subscriptionID", wh.SubscriptionID,
				"userID", sub.MattermostCreatorID,
				"userSubscriptionID", creator.Settings.EventSubscriptionID)
			// TODO: delete the orphaned subscription?
			// h.notFound(w, errors.New("orphaned subscription "+sub.Remote.ID))
			//<><> TODO uncomment - return success for now
			w.WriteHeader(http.StatusAccepted)
			return
		}

		client := h.Remote.NewClient(context.Background(), h.Config, creator.OAuth2Token, h.Logger)
		var event *remote.Event
		event, err = client.GetUserEvent(creator.Remote.ID, sub.Remote.ID)

		err = h.BotPoster.PostDirect(sub.MattermostCreatorID,
			fmt.Sprintf("%s: %s\n", wh.ChangeType, utils.PrettyJSON(event)), "")
		if err != nil {
			h.internalServerError(w, err)
			return
		}
	}
	// }()

	w.WriteHeader(http.StatusAccepted)
	h.Logger.LogDebug("Webhook received")
	return
}
