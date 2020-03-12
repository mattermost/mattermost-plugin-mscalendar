package mscalendar

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/flow"
)

type welcomeFlow struct {
	steps      []flow.Step
	url        string
	flower     bot.Flower
	onFlowDone func(userID string)
}

func NewWelcomeFlow(url string, bot bot.Flower, onFlowDone func(userID string)) *welcomeFlow {
	wf := welcomeFlow{
		url:        url,
		flower:     bot,
		onFlowDone: onFlowDone,
	}
	wf.makeSteps()
	return &wf
}

func (wf *welcomeFlow) Step(i int) flow.Step {
	if i < 0 {
		return nil
	}
	if i >= len(wf.steps) {
		return nil
	}
	return wf.steps[i]
}

func (wf *welcomeFlow) URL() string {
	return wf.url
}

func (wf *welcomeFlow) Length() int {
	return len(wf.steps)
}

func (wf *welcomeFlow) StepDone(userID string, value bool) {
	wf.flower.NextStep(userID, value)
}

func (wf *welcomeFlow) FlowDone(userID string) {
	wf.onFlowDone(userID)
}

func (wf *welcomeFlow) makeSteps() {
	steps := []flow.Step{}
	steps = append(steps, flow.NewStep(
		"Update Status",
		"Do you wish your Mattermost status to be automatically updated to be *Do not disturb* at the time of your Microsoft Calendar events?",
		store.UpdateStatusPropertyName,
		"Yes - Update my status",
		"No - Don't update my status",
		":thumbsup: Got it! We'll automatically update your status in Mattermost.",
		":thumbsup: Got it! We won't update your status in Mattermost.",
		0,
		1,
	), flow.NewStep(
		"Confirm status change",
		"Do you want to receive confirmations before we update your status for each event?",
		store.GetConfirmationPropertyName,
		"Yes - I will like to get confirmations",
		"No - Update my status automatically",
		"Cool, we'll also send you confirmations before updating your status.",
		"There has been a problem setting the confirmation configuration",
		0,
		0,
	))

	wf.steps = steps
}
