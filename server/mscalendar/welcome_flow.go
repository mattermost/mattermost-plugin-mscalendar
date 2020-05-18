package mscalendar

import (
	"github.com/gorilla/mux"
	"github.com/larkox/mattermost-plugin-utils/bot/poster"
	"github.com/larkox/mattermost-plugin-utils/flow"
	"github.com/larkox/mattermost-plugin-utils/flow/steps"
	"github.com/larkox/mattermost-plugin-utils/freetext_fetcher"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

func MakeFlowSteps(ftStore freetext_fetcher.FreetextStore, r *mux.Router, posterBot poster.Poster) []steps.Step {
	stepList := []steps.Step{
		steps.NewSimpleStep(
			"Update Status",
			"Do you wish your Mattermost status to be automatically updated to be *Do not disturb* at the time of your Microsoft Calendar events?",
			store.UpdateStatusPropertyName,
			"Yes - Update my status",
			"No - Don't update my status",
			":thumbsup: Got it! We'll automatically update your status in Mattermost.",
			":thumbsup: Got it! We won't update your status in Mattermost.",
			0,
			1,
		),
		steps.NewSimpleStep(
			"Confirm status change",
			"Do you want to receive confirmations before we update your status for each event?",
			store.GetConfirmationPropertyName,
			"Yes - I would like to get confirmations",
			"No - Update my status automatically",
			"Cool, we'll also send you confirmations before updating your status.",
			"Cool, we'll update your status automatically with no confirmation.",
			0,
			0,
		),
		steps.NewSimpleStep(
			"Subscribe to events",
			"Do you want to receive notifications when you receive a new event?",
			store.SubscribePropertyName,
			"Yes - I would like to receive notifications for new events",
			"No - Do not notify me of new events",
			"Great, you will receive a message any time you receive a new event.",
			"Great, you will not receive any notification on new events.",
			0,
			0,
		),
		steps.NewFreetextStep(
			"Free text test",
			"This is the step description. It can also include information about the validation rules. In this case, the string must be longer than 3 characters.",
			store.TestPropertyName,
			config.PathFreeTextHandler,
			ftStore,
			func(message string) string {
				if len(message) < 3 {
					return "The string must be longer than 3 characters."
				}
				return ""
			},
			r,
			posterBot,
		),
		steps.NewEmptyStep(
			"Daily Summary",
			"Remember that you can set-up a daily summary by typing `/mscalendar summary time 8:00AM`.",
		),
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

func (s *flowStoreWrapper) SetProperty(userID, propertyName string, value interface{}) error {
	if propertyName == store.SubscribePropertyName {
		boolValue := value.(string) == "true"
		if boolValue {
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
