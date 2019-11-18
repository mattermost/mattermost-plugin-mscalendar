// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

type webhook struct {
	ChangeType                     string `json:"changeType"`
	ClientState                    string `json:"clientState,omitempty"`
	Resource                       string `json:"resource,omitempty"`
	SubscriptionExpirationDateTime string `json:"subscriptionExpirationDateTime,omitempty"`
	SubscriptionID                 string `json:"subscriptionId"`
	ResourceData                   struct {
		DataType string `json:"@odata.type"`
	} `json:"resourceData"`
}

func (r *impl) HandleEventNotification(w http.ResponseWriter, req *http.Request, loadf remote.LoadSubscriptionCreatorF) []*remote.EventNotification {

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

	notifications := []*remote.EventNotification{}
	for _, wh := range v.Value {
		creator, token, creatorMattermostID, sub, err := loadf(wh.SubscriptionID)
		if err != nil {
			// w.WriteHeader(http.StatusNotFound)
			w.WriteHeader(http.StatusAccepted)
			r.logger.LogInfo("Failed to process webhook",
				"error", err.Error())
			return nil
		}

		if sub.ClientState != "" && sub.ClientState != wh.ClientState {
			w.WriteHeader(http.StatusUnauthorized)
			r.logger.LogInfo("Unauthorized webhook")
			return nil
		}

		n := &remote.EventNotification{
			SubscriptionID:                      wh.SubscriptionID,
			Subscription:                        sub,
			SubscriptionCreator:                 creator,
			SubscriptionCreatorMattermostUserID: creatorMattermostID,
		}

		client := r.NewClient(context.Background(), token)
		switch wh.ResourceData.DataType {
		case "#Microsoft.Graph.Message":
			em := remote.EventMessage{}
			path := wh.Resource + "?$expand=microsoft.graph.eventMessage/Event"
			err = client.Call(http.MethodGet, path, nil, &em)
			if err != nil {
				r.logger.LogInfo("Error fetching resource",
					"error", err.Error(),
					"subscriptionID", wh.SubscriptionID,
					"creatorID", creator.ID)
				w.WriteHeader(http.StatusAccepted)
				// w.WriteHeader(http.StatusInternalServerError)
				return nil
			}
			n.EventMessage = &em
			n.Event = em.Event

			switch em.MeetingMessageType {
			case "meetingRequest":
				n.Change = remote.ChangeInvitedMe
			case "meetingCancelled":
				n.Change = remote.ChangeMeetingCancelled
			case "meetingAccepted":
				n.Change = remote.ChangeAccepted
			case "meetingTenativelyAccepted":
				n.Change = remote.ChangeTentativelyAccepted
			case "meetingDeclined":
				n.Change = remote.ChangeDeclined
			default:
				r.logger.LogInfo("Non-calendar message",
					"subscriptionID", wh.SubscriptionID,
					"creatorID", creator.ID,
					"subject", em.Subject)
				w.WriteHeader(http.StatusAccepted)
				return nil
			}

		case "#Microsoft.Graph.Event":
			event := remote.Event{}
			err = client.Call(http.MethodGet, wh.Resource, nil, &event)
			if err != nil {
				r.logger.LogInfo("Error fetching resource",
					"error", err.Error(),
					"subscriptionID", wh.SubscriptionID,
					"creatorID", creator.ID)
				w.WriteHeader(http.StatusAccepted)
				return nil
			}
			n.Event = &event

			switch wh.ChangeType {
			case "created":
				n.Change = remote.ChangeEventCreated
			case "updated":
				n.Change = remote.ChangeEventUpdated
			case "deleted":
				n.Change = remote.ChangeEventDeleted
			default:
				r.logger.LogInfo("Unrecognized Event change type: "+wh.ChangeType,
					"subscriptionID", wh.SubscriptionID,
					"creatorID", creator.ID)
				w.WriteHeader(http.StatusAccepted)
				return nil
			}

		default:
			r.logger.LogInfo("Unknown resource type: "+wh.ResourceData.DataType,
				"subscriptionID", wh.SubscriptionID,
				"creatorID", creator.ID)
			w.WriteHeader(http.StatusAccepted)
			return nil
		}

		notifications = append(notifications, n)
	}

	return notifications
}
