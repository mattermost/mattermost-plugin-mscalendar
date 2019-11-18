// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

type Event struct {
	ID                string               `json:"id"`
	Subject           string               `json:"subject,omitempty"`
	BodyPreview       string               `json:"bodyPreview,omitempty"`
	Importance        string               `json:"importance,omitempty"`
	IsAllDay          bool                 `json:"isAllDay,omitempty"`
	IsCancelled       bool                 `json:"isCancelled,omitempty"`
	IsOrganizer       bool                 `json:"isOrganizer,omitempty"`
	ResponseRequested bool                 `json:"responseRequested,omitempty"`
	ShowAs            string               `json:"showAs,omitempty"`
	Weblink           string               `json:"weblink,omitempty"`
	Start             *DateTime            `json:"start,omitempty"`
	End               *DateTime            `json:"end,omitempty"`
	Location          *EventLocation       `json:"location,omitempty"`
	ResponseStatus    *EventResponseStatus `json:"responseStatus,omitempty"`
	Attendees         []*EventAttendee     `json:"attendees,omitempty"`
	Organizer         *EventAttendee       `json:"organizer,omitempty"`
}

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
	GraphChangeType                     string
	Event                               *Event
	EventMessage                        *EventMessage
	SubscriptionID                      string
	Subscription                        *Subscription
	SubscriptionCreator                 *User
	SubscriptionCreatorMattermostUserID string
}
type EventMessage struct {
	ID                 string        `json:"id"`
	SentDateTime       string        `json:"sentDateTime,omitrmpty"`
	Subject            string        `json:"subject,omitempty"`
	BodyPreview        string        `json:"bodyPreview,omitempty"`
	Importance         string        `json:"importance,omitempty"`
	Weblink            string        `json:"weblink,omitempty"`
	MeetingMessageType string        `json:"meetingMessageType,omitempty"`
	Sender             *EmailAddress `json:"sender,omitempty"`
	From               *EmailAddress `json:"from,omitempty"`
	Event              *Event        `json:"event,omitempty"`
}

type EventResponseStatus struct {
	Response string `json:"response,omitempty"`
	Time     string `json:"time,omitempty"`
}

type EventLocation struct {
	DisplayName string `json:"displayName,omitempty"`
}

type EventAttendee struct {
	Type         string               `json:"type,omitempty"`
	Status       *EventResponseStatus `json:"status,omitempty"`
	EmailAddress *EmailAddress        `json:"emailAddress,omitempty"`
}
