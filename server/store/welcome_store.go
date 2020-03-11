package store

import "github.com/mattermost/mattermost-plugin-mscalendar/server/utils/kvstore"

type WelcomeStore interface {
	LoadUserWelcomePost(mattermostID string) (string, error)
	StoreUserWelcomePost(mattermostID, postID string) error
	DeleteUserWelcomePost(mattermostID string) (string, error)
}

func (s *pluginStore) LoadUserWelcomePost(mattermostID string) (string, error) {
	var postID string
	err := kvstore.LoadJSON(s.welcomeIndexKV, mattermostID, &postID)
	if err != nil {
		return "", err
	}
	return postID, nil
}

func (s *pluginStore) StoreUserWelcomePost(mattermostID, postID string) error {
	err := kvstore.StoreJSON(s.welcomeIndexKV, mattermostID, postID)
	if err != nil {
		return err
	}
	return nil
}

func (s *pluginStore) DeleteUserWelcomePost(mattermostID string) (string, error) {
	var postID string
	kvstore.LoadJSON(s.welcomeIndexKV, mattermostID, &postID)
	err := s.welcomeIndexKV.Delete(mattermostID)
	if err != nil {
		return "", err
	}
	return postID, nil
}
