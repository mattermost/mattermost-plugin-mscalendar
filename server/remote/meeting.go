// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import "time"

type FindMeetingTimesParameters struct {
	Attendees                 []Attendee          `json:"attendees,omitempty"`
	LocationConstraint        *LocationConstraint `json:"locationConstraint,omitempty"`
	TimeConstraint            *TimeConstraint     `json:"timeConstraint,omitempty"`
	MeetingDuration           *time.Duration      `json:"meetingDuration,omitempty"`
	MaxCandidates             *int                `json:"maxCandidates,omitempty"`
	IsOrganizerOptional       *bool               `json:"isOrganizerOptional,omitempty"`
	ReturnSuggestionReasons   *bool               `json:"returnSuggestionReasons,omitempty"`
	MinimumAttendeePercentage *float64            `json:"minimumAttendeePercentage,omitempty"`
}

type TimeConstraint struct {
	ActivityDomain string     `json:"activityDomain,omitempty"`
	TimeSlots      []TimeSlot `json:"timeSlots,omitempty"`
}
type MeetingTimeSuggestion struct {
	AttendeeAvailability  []*AttendeeAvailability
	Confidence            float32 `json:"confidence"`
	Locations             []*Location
	MeetingTimeSlot       *TimeSlot
	Order                 int32  `json:"order"`
	OrganizerAvailability string `json:"organizerAvailability"`
	SuggestionReason      string `json:"suggestionReason"`
}

type AttendeeAvailability struct {
	Attendee     *Attendee
	Availability string `json:"availability"`
}

type MeetingTimeSuggestionResults struct {
	MeetingTimeSuggestions []*MeetingTimeSuggestion `json:"meetingTimeSuggestions"`
	EmptySuggestionReason  string                   `json:"emptySuggestionReason"`
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
