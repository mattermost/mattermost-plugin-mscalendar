package mscalendar

import (
	"github.com/larkox/mattermost-plugin-utils/flow"
	"github.com/larkox/mattermost-plugin-utils/flow/steps"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

func MakeFlowSteps() []steps.Step {
	stepList := []steps.Step{
		&steps.SimpleStep{
			Title:                "Update Status",
			Message:              "Do you wish your Mattermost status to be automatically updated to be *Do not disturb* at the time of your Microsoft Calendar events?",
			PropertyName:         store.UpdateStatusPropertyName,
			TrueButtonMessage:    "Yes - Update my status",
			FalseButtonMessage:   "No - Don't update my status",
			TrueResponseMessage:  ":thumbsup: Got it! We'll automatically update your status in Mattermost.",
			FalseResponseMessage: ":thumbsup: Got it! We won't update your status in Mattermost.",
			FalseSkip:            1,
		},
		&steps.SimpleStep{
			Title:                "Confirm status change",
			Message:              "Do you want to receive confirmations before we update your status for each event?",
			PropertyName:         store.GetConfirmationPropertyName,
			TrueButtonMessage:    "Yes - I would like to get confirmations",
			FalseButtonMessage:   "No - Update my status automatically",
			TrueResponseMessage:  "Cool, we'll also send you confirmations before updating your status.",
			FalseResponseMessage: "Cool, we'll update your status automatically with no confirmation.",
		},
		&steps.SimpleStep{
			Title:                "Subscribe to events",
			Message:              "Do you want to receive notifications when you receive a new event?",
			PropertyName:         store.SubscribePropertyName,
			TrueButtonMessage:    "Yes - I would like to receive notifications for new events",
			FalseButtonMessage:   "No - Do not notify me of new events",
			TrueResponseMessage:  "Great, you will receive a message any time you receive a new event.",
			FalseResponseMessage: "Great, you will not receive any notification on new events.",
		},
		&steps.EmptyStep{
			Title:   "Daily Summary",
			Message: "Remember that you can set-up a daily summary by typing `/mscalendar summary time 8:00AM`.",
		},
	}

	return stepList
}

func NewWelcomeStore(s flow.FlowStore, env Env) flow.FlowStore {
	return &flowStoreWrapper{
		FlowStore: s,
		Env:       env,
	}
}

type flowStoreWrapper struct {
	flow.FlowStore
	Env
}

func (s *flowStoreWrapper) SetProperty(userID, propertyName string, value bool) error {
	if propertyName == store.SubscribePropertyName {
		if value {
			m := New(s.Env, userID)
			l, err := m.ListRemoteSubscriptions()
			if err != nil {
				return err
			}
			if len(l) >= 1 {
				return nil
			}

			_, err = m.CreateMyEventSubscription()
			if err != nil {
				return err
			}
		}
		return nil
	}

	return s.FlowStore.SetProperty(userID, propertyName, value)
}
