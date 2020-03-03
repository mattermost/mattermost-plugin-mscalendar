package mscalendar

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/welcome_flow"
)

type welcomeApp struct {
	Env
}

func NewWelcomeApp(env Env) welcome_flow.App {
	return &welcomeApp{
		Env: env,
	}
}

func (app *welcomeApp) SetUpdateStatus(mattermostUserID string, updateStatus bool) error {
	user, err := app.Store.LoadUser(mattermostUserID)
	if err != nil {
		return err
	}
	user.Settings.UpdateStatus = updateStatus
	err = app.Store.StoreUser(user)
	if err != nil {
		return err
	}
	err = app.Welcomer.AfterUpdateStatus(mattermostUserID, updateStatus)
	return err
}

func (app *welcomeApp) SetGetConfirmation(mattermostUserID string, getConfirmation bool) error {
	user, err := app.Store.LoadUser(mattermostUserID)
	if err != nil {
		return err
	}
	user.Settings.GetConfirmation = getConfirmation
	err = app.Store.StoreUser(user)
	if err != nil {
		return err
	}
	err = app.Welcomer.AfterSetConfirmations(mattermostUserID, getConfirmation)
	return err
}

func (m *mscalendar) Welcome(userID string) error {
	return m.Welcomer.Welcome(userID)
}
