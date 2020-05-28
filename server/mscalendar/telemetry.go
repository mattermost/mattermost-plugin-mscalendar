package mscalendar

const (
	welcomeFlowCompletionEvent = "welcomeFlowCompletion"
)

func (m *mscalendar) TrackWelcomeFlowCompletion() {
	m.Tracker.Track(welcomeFlowCompletionEvent, map[string]interface{}{})
}
