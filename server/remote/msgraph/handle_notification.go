// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

const renewSubscriptionBeforeExpiration = 12 * time.Hour

type webhookNotification struct {
	ChangeType                     string `json:"changeType"`
	ClientState                    string `json:"clientState,omitempty"`
	Resource                       string `json:"resource,omitempty"`
	SubscriptionExpirationDateTime string `json:"subscriptionExpirationDateTime,omitempty"`
	SubscriptionID                 string `json:"subscriptionId"`
	ResourceData                   struct {
		DataType string `json:"@odata.type"`
	} `json:"resourceData"`
}

func (r *impl) HandleNotification(w http.ResponseWriter, req *http.Request) []*remote.Notification {
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
		Value []*webhookNotification `json:"value"`
	}
	err = json.Unmarshal(rawData, &v)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.LogInfo("Failed to process webhook",
			"error", err.Error())
		return nil
	}

	notifications := []*remote.Notification{}
	for _, wh := range v.Value {
		n := &remote.Notification{
			SubscriptionID:      wh.SubscriptionID,
			ChangeType:          wh.ChangeType,
			ClientState:         wh.ClientState,
			IsBare:              true,
			WebhookRawData:      rawData,
			WebhookNotification: wh,
		}

		expires, err := time.Parse("2006-01-02T15:04:05Z", wh.SubscriptionExpirationDateTime)
		if err != nil {
			r.logger.LogInfo("Invalid subscription expiration in webhook: "+err.Error(),
				"SubscriptionID", wh.SubscriptionID)
			return nil
		}
		expires = expires.Add(-renewSubscriptionBeforeExpiration)
		if time.Now().After(expires) {
			n.RecommendRenew = true
		}

		notifications = append(notifications, n)
	}

	w.WriteHeader(http.StatusAccepted)
	return notifications
}
