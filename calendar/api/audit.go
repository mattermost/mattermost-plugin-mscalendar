// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api

// CreateEventAuditParams holds request audit data for the createEvent operation.
type CreateEventAuditParams struct {
	MattermostUserID string `json:"mattermost_user_id"`
	ChannelID        string `json:"channel_id,omitempty"`
}

func (p CreateEventAuditParams) Auditable() map[string]any {
	return map[string]any{
		"mattermost_user_id": p.MattermostUserID,
		"channel_id":         p.ChannelID,
	}
}

// CreateEventAuditResult holds the outcome of the createEvent operation.
type CreateEventAuditResult struct {
	EventID string `json:"event_id"`
	ICalUID string `json:"ical_uid"`
}

func (r CreateEventAuditResult) Auditable() map[string]any {
	return map[string]any{
		"event_id": r.EventID,
		"ical_uid": r.ICalUID,
	}
}
