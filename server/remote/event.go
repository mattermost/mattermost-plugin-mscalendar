// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import "time"

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
	Location                   *Location            `json:"location,omitempty"`
	ResponseStatus             *EventResponseStatus `json:"responseStatus,omitempty"`
	Attendees                  []*Attendee          `json:"attendees,omitempty"`
	Organizer                  *Attendee            `json:"organizer,omitempty"`
}

type ItemBody struct {
	Content     string `json:"content,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

type EventResponseStatus struct {
	Response string `json:"response,omitempty"`
	Time     string `json:"time,omitempty"`
}

type Location struct {
	DisplayName  string       `json:"displayName,omitempty"`
	Address      *Address     `json:"address"`
	Coordinates  *Coordinates `json:"coordinates"`
	LocationType string       `json:"locationType"`
}

type Address struct {
	Street          string `json:"street,omitempty"`
	City            string `json:"city,omitempty"`
	State           string `json:"state,omitempty"`
	CountryOrRegion string `json:"countryOrRegion,omitempty"`
	PostalCode      string `json:"postalCode,omitempty"`
}

type Coordinates struct {
	Latitude  float32 `json:"latitude,omitempty"`
	Longitude float32 `json:"longitude,omitempty"`
}

type Attendee struct {
	Type         string               `json:"type,omitempty"`
	Status       *EventResponseStatus `json:"status,omitempty"`
	EmailAddress *EmailAddress        `json:"emailAddress,omitempty"`
}

// TODO possibly move some of these, non-event types to remote/common.go

// Type = *AttendeeType = options (Required, Optional, Resource)
type AttendeeBase struct {
	Recipient *EmailAddress
	Type      string `json:"type,omitempty"`
}

type FindMeetingTimesParameters struct {
	Attendees                 []AttendeeBase      `json:"attendees,omitempty"`
	LocationConstraint        *LocationConstraint `json:"locationConstraint,omitempty"`
	TimeConstraint            *TimeConstraint     `json:"timeConstraint,omitempty"`
	MeetingDuration           *time.Duration      `json:"meetingDuration,omitempty"`
	MaxCandidates             *int                `json:"maxCandidates,omitempty"`
	IsOrganizerOptional       *bool               `json:"isOrganizerOptional,omitempty"`
	ReturnSuggestionReasons   *bool               `json:"returnSuggestionReasons,omitempty"`
	MinimumAttendeePercentage *float64            `json:"minimumAttendeePercentage,omitempty"`
}

// *ActivityDomain= options (Unknown, Work, Personal, Unrestricted)
type TimeConstraint struct {
	ActivityDomain string     `json:"activityDomain,omitempty"`
	TimeSlots      []TimeSlot `json:"timeSlots,omitempty"`
}
type MeetingTimeSuggestion struct {
	AttendeeAvailability  string
	confidence            float32
	locations             []*Location
	meetingTimeSlot       *TimeSlot
	order                 int32
	organizerAvailability string
	suggestionReason      string
}

type MeetingTimeSuggestionResults struct {
	MeetingTimeSuggestions []*MeetingTimeSuggestion
	emptySuggestionReason  string
}

type TimeSlot struct {
	Start *DateTime `json:"start,omitempty"`
	End   *DateTime `json:"end,omitempty"`
}

type LocationConstraint struct {
	Locations       []LocationConstraintItem `json:"locations,omitempty"`
	IsRequired      *bool                    `json:"isRequired,omitempty"`
	SuggestLocation *bool                    `json:"suggestLocation,omitempty"`
}

type LocationConstraintItem struct {
	Location            *Location
	ResolveAvailability *bool `json:"resolveAvailability,omitempty"`
}
