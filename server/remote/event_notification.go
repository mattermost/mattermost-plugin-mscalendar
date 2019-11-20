// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

const (
	ChangeInvitedMe           = "invited.me"
	ChangeAccepted            = "accepted"
	ChangeTentativelyAccepted = "tentativelyAccepted"
	ChangeMeetingCancelled    = "meetingCancelled"
	ChangeDeclined            = "declined"
	ChangeEventCreated        = "eventCreated"
	ChangeEventUpdated        = "eventUpdated"
	ChangeEventDeleted        = "eventDeleted"
)

type EventNotification struct {
	Change                              string
	Event                               *Event
	EventMessage                        *EventMessage
	SubscriptionID                      string
	Subscription                        *Subscription
	SubscriptionCreator                 *User
	SubscriptionCreatorMattermostUserID string
	EntityRawData                       []byte
}

type EventMessage struct {
	ID                 string `json:"id"`
	SentDateTime       string `json:"sentDateTime,omitrmpty"`
	Subject            string `json:"subject,omitempty"`
	BodyPreview        string `json:"bodyPreview,omitempty"`
	Importance         string `json:"importance,omitempty"`
	Weblink            string `json:"weblink,omitempty"`
	MeetingMessageType string `json:"meetingMessageType,omitempty"`
	Event              *Event `json:"event,omitempty"`

	Sender struct {
		*EmailAddress `json:"emailAddress,omitempty"`
	} `json:"sender,omitempty"`

	From struct {
		*EmailAddress `json:"emailAddress,omitempty"`
	} `json:"from,omitempty"`
}
