// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
	"golang.org/x/oauth2"
)

type webhook struct {
	ChangeType                     string `json:"changeType"`
	ClientState                    string `json:"clientState,omitempty"`
	Resource                       string `json:"resource,omitempty"`
	SubscriptionExpirationDateTime string `json:"subscriptionExpirationDateTime,omitempty"`
	SubscriptionID                 string `json:"subscriptionId"`
}

func (r *impl) ProcessEventWebhook(w http.ResponseWriter, req *http.Request,
	creator func(subID string) (*remote.User, *oauth2.Token, string, *remote.Subscription, error)) []*remote.EventNotification {

	// Microsoft graph requires webhook endpoint validation, see
	// https://docs.microsoft.com/en-us/graph/webhooks#notification-endpoint-validation
	vtok := req.FormValue("validationToken")
	if vtok != "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(vtok))
		r.logger.LogDebug("Validated event webhook endpoint")
		return nil
	}

	rawData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		r.logger.LogInfo("Failed to process webhook",
			"error", err.Error())
		return nil
	}

	// r.logger.LogInfo("<><>-\n" + string(rawData) + "\n-<><>")

	// Get the list of webhooks
	var v struct {
		Value []*webhook `json:"value"`
	}
	err = json.Unmarshal(rawData, &v)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.LogInfo("Failed to process webhook",
			"error", err.Error())
		return nil
	}
	r.logger.LogInfo(fmt.Sprintf("<><> Unmarshalled: %s", utils.PrettyJSON(v.Value)))

	notifications := []*remote.EventNotification{}
	for _, wh := range v.Value {
		creator, token, creatorMattermostId, sub, err := creator(wh.SubscriptionID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			r.logger.LogInfo("Failed to process webhook",
				"error", err.Error())
			return nil
		}

		if sub.ClientState != "" && sub.ClientState != wh.ClientState {
			w.WriteHeader(http.StatusUnauthorized)
			r.logger.LogInfo("Unauthorized webhook")
			return nil
		}

		client := r.NewClient(context.Background(), token)
		// The webhook contains a convenient Resource path to get the relevant resource.
		// However, the present client does not make it easy to query a predefined path,
		// so for now assume the format of /users/UserID/events/EventID, and then use the
		// existing API to retrieve it.
		parts := strings.Split(wh.Resource, "/")
		if len(parts) != 4 || parts[0] != "Users" || parts[2] != "Events" {
			w.WriteHeader(http.StatusBadRequest)
			r.logger.LogInfo("Failed to process webhook",
				"error", fmt.Sprintf("invalid resource format %q, expected /Users/{id}/Events/{id}", wh.Resource))
			return nil
		}

		event, err := client.GetUserEvent(parts[1], parts[3])
		if err != nil {
			r.logger.LogInfo("Error fetching resource",
				"error", err.Error(),
				"subscriptionID", wh.SubscriptionID,
				"creatorID", creator.ID,
				"eventOwnerID", parts[1],
				"resourceID", parts[3])
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		}

		notifications = append(notifications, &remote.EventNotification{
			SubscriptionID:          wh.SubscriptionID,
			ChangeType:              wh.ChangeType,
			Event:                   event,
			Subscription:            sub,
			Creator:                 creator,
			CreatorMattermostUserID: creatorMattermostId,
		})
	}

	return notifications
}
