package mscalendar

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/flow"
)

type WelcomeFlow struct {
	controller bot.FlowController
	onFlowDone func(userID string)
	url        string
	steps      []flow.Step
}

func NewWelcomeFlow(bot bot.FlowController, welcomer Welcomer) *WelcomeFlow {
	wf := WelcomeFlow{
		url:        "/welcome",
		controller: bot,
		onFlowDone: welcomer.WelcomeFlowEnd,
	}
	wf.makeSteps()
	return &wf
}

func (wf *WelcomeFlow) Step(i int) flow.Step {
	if i < 0 {
		return nil
	}
	if i >= len(wf.steps) {
		return nil
	}
	return wf.steps[i]
}

func (wf *WelcomeFlow) URL() string {
	return wf.url
}

func (wf *WelcomeFlow) Length() int {
	return len(wf.steps)
}

func (wf *WelcomeFlow) StepDone(userID string, step int, value bool) {
	wf.controller.NextStep(userID, step, value)
}

func (wf *WelcomeFlow) FlowDone(userID string) {
	wf.onFlowDone(userID)
}

func (wf *WelcomeFlow) makeSteps() {
	steps := []flow.Step{}
	steps = append(steps, &flow.EmptyStep{
		Title:   "Update Status",
		Message: "You can update your status to Away or Do not disturb when you are in a meeting by typing `/mscalendar settings`.",
	}, &flow.SimpleStep{
		Title:                "Confirm status change",
		Message:              "Do you want to receive confirmations before we update your status for each event?",
		PropertyName:         store.GetConfirmationPropertyName,
		TrueButtonMessage:    "Yes - I would like to get confirmations",
		FalseButtonMessage:   "No - Update my status automatically",
		TrueResponseMessage:  "Cool, we'll also send you confirmations before updating your status.",
		FalseResponseMessage: "Cool, we'll update your status automatically with no confirmation.",
	}, &flow.SimpleStep{
		Title:                "Set Custom Status",
		Message:              "Do you want to set custom status automatically on Mattermost when you are in a meeting?",
		PropertyName:         store.SetCustomStatusPropertyName,
		TrueButtonMessage:    "Yes - I would like to set custom status automatically",
		FalseButtonMessage:   "No - Don't set custom status",
		TrueResponseMessage:  "Cool, we'll set custom status automatically when you are in a meeting.",
		FalseResponseMessage: "Cool, we'll not set custom status when you are in a meeting.",
	}, &flow.SimpleStep{
		Title:                "Subscribe to events",
		Message:              "Do you want to receive notifications when you are invited to an event?",
		PropertyName:         store.SubscribePropertyName,
		TrueButtonMessage:    "Yes - I would like to receive notifications for new events",
		FalseButtonMessage:   "No - Do not notify me of new events",
		TrueResponseMessage:  "Great, you will receive a message any time you receive a new event.",
		FalseResponseMessage: "Great, you will not receive any notification on new events.",
	}, &flow.SimpleStep{
		Title:                "Receive reminder",
		Message:              "Do you want to receive a reminder for upcoming events?",
		PropertyName:         store.ReceiveUpcomingEventReminderName,
		TrueButtonMessage:    "Yes - I would like to receive reminders for upcoming events",
		FalseButtonMessage:   "No - Do not notify me of upcoming events",
		TrueResponseMessage:  "Great, you will receive a message before your meetings.",
		FalseResponseMessage: "Great, you will not receive any notification for upcoming events.",
	}, &flow.EmptyStep{
		Title:   "Daily Summary",
		Message: "Remember that you can set-up a daily summary by typing `/mscalendar summary time 8:00AM`.",
	})

	wf.steps = steps
}
