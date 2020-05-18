package store

type FreetextStore interface {
	StartFetching(userID, fetcherID string, payload interface{}) error
	StopFetching(userID, fetcherID string) error
	ShouldProcessFreetext(userID, fetcherID string) (bool, interface{}, error)
}

func (s *pluginStore) StartFetching(userID, fetcherID string, payload string) error {
	user, err := s.LoadUser(userID)
	if err != nil {
		return err
	}

	user.FreetextFetching.ID = fetcherID
	user.FreetextFetching.Payload = payload

	err = s.StoreUser(user)
	if err != nil {
		return err
	}
	return nil
}
func (s *pluginStore) StopFetching(userID, fetcherID string) error {
	user, err := s.LoadUser(userID)
	if err != nil {
		return err
	}

	user.FreetextFetching.ID = ""
	user.FreetextFetching.Payload = ""

	err = s.StoreUser(user)
	if err != nil {
		return err
	}
	return nil
}

func (s *pluginStore) ShouldProcessFreetext(userID, fetcherID string) (bool, string, error) {
	user, err := s.LoadUser(userID)
	if err != nil {
		return false, "", err
	}

	if fetcherID == user.FreetextFetching.ID {
		return true, user.FreetextFetching.Payload, nil
	}

	return false, "", nil
}
