package mscalendarTracker

import "github.com/mattermost/mattermost-plugin-mscalendar/server/utils/telemetry"

const (
	welcomeFlowCompletionEvent    = "welcomeFlowCompletion"
	userAuthenticatedEvent        = "userAuthenticated"
	userDeauthenticatedEvent      = "userDeauthenticated"
	automaticStatusUpdateOnEvent  = "automaticStatusUpdateOn"
	automaticStatusUpdateOffEvent = "automaticStatusUpdateOff"
	dailySummarySentEvent         = "dailySummarySent"
)

type Tracker interface {
	TrackWelcomeFlowCompletion(userID string)
	TrackUserAuthenticated(userID string)
	TrackUserDeauthenticated(userID string)
	TrackDailySummarySent(userID string)
	TrackAutomaticStatusUpdateOn(userID string)
	TrackAutomaticStatusUpdateOff(userID string)
}

func New(t telemetry.Tracker) Tracker {
	return &tracker{
		tracker: t,
	}
}

type tracker struct {
	tracker telemetry.Tracker
}

func (t *tracker) TrackWelcomeFlowCompletion(userID string) {
	t.tracker.TrackUserEvent(welcomeFlowCompletionEvent, userID, map[string]interface{}{})
}

func (t *tracker) TrackUserAuthenticated(userID string) {
	t.tracker.TrackUserEvent(userAuthenticatedEvent, userID, map[string]interface{}{})
}

func (t *tracker) TrackUserDeauthenticated(userID string) {
	t.tracker.TrackUserEvent(userDeauthenticatedEvent, userID, map[string]interface{}{})
}

func (t *tracker) TrackDailySummarySent(userID string) {
	t.tracker.TrackUserEvent(dailySummarySentEvent, userID, map[string]interface{}{})
}

func (t *tracker) TrackAutomaticStatusUpdateOn(userID string) {
	t.tracker.TrackUserEvent(automaticStatusUpdateOnEvent, userID, map[string]interface{}{})
}

func (t *tracker) TrackAutomaticStatusUpdateOff(userID string) {
	t.tracker.TrackUserEvent(automaticStatusUpdateOffEvent, userID, map[string]interface{}{})
}
