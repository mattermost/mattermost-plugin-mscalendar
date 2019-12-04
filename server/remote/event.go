// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

type Event struct {
	ID                         string               `json:"id,omitempty"`
	Subject                    string               `json:"subject,omitempty"`
	BodyPreview                string               `json:"bodyPreview,omitempty"`
	Body                       *ItemBody            `json:"Body,omitempty"`
	Importance                 string               `json:"importance,omitempty"`
	IsAllDay                   bool                 `json:"isAllDay,omitempty"`
	IsCancelled                bool                 `json:"isCancelled,omitempty"`
	IsOrganizer                bool                 `json:"isOrganizer,omitempty"`
	ResponseRequested          bool                 `json:"responseRequested,omitempty"`
	ShowAs                     string               `json:"showAs,omitempty"`
	Weblink                    string               `json:"weblink,omitempty"`
	Start                      *DateTime            `json:"start,omitempty"`
	End                        *DateTime            `json:"end,omitempty"`
	ReminderMinutesBeforeStart int32                `json:"reminderMinutesBeforeStart,omitempty"`
	Location                   *EventLocation       `json:"location,omitempty"`
	ResponseStatus             *EventResponseStatus `json:"responseStatus,omitempty"`
	Attendees                  []*EventAttendee     `json:"attendees,omitempty"`
	Organizer                  *EventAttendee       `json:"organizer,omitempty"`
}

type ItemBody struct {
	Content     string `json:"content,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

type EventResponseStatus struct {
	Response string `json:"response,omitempty"`
	Time     string `json:"time,omitempty"`
}

type EventLocation struct {
	DisplayName  string            `json:"displayName,omitempty"`
	Address      *EventAddress     `json:"address"`
	Coordinates  *EventCoordinates `json:"coordinates"`
	LocationType string            `json:"locationType"`
}

type EventAddress struct {
	Street          string `json:"street,omitempty"`
	City            string `json:"city,omitempty"`
	State           string `json:"state,omitempty"`
	CountryOrRegion string `json:"countryOrRegion,omitempty"`
	PostalCode      string `json:"postalCode,omitempty"`
}

type EventCoordinates struct {
	Latitude  float32 `json:"latitude,omitempty"`
	Longitude float32 `json:"longitude,omitempty"`
}

type EventAttendee struct {
	Type         string               `json:"type,omitempty"`
	Status       *EventResponseStatus `json:"status,omitempty"`
	EmailAddress *EmailAddress        `json:"emailAddress,omitempty"`
}
