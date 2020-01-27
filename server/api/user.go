package api

func (api *api) GetUserTimezone(mattermostUserID string) (string, error) {
	client, err := api.MakeClient()
	if err != nil {
		return "", err
	}

	remoteUserID, err := api.getRemoteUserID(mattermostUserID)
	if err != nil {
		return "", err
	}

	settings, err := client.GetUserMailboxSettings(remoteUserID)
	if err != nil {
		return "", err
	}

	return settings.TimeZone, nil
}

func (api *api) getRemoteUserID(mattermostUserID string) (string, error) {
	if api.user != nil && api.user.MattermostUserID == mattermostUserID {
		return api.user.Remote.ID, nil
	}

	u, storeErr := api.UserStore.LoadUser(mattermostUserID)
	if storeErr != nil {
		return "", storeErr
	}
	return u.Remote.ID, nil
}
