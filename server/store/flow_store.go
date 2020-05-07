package store

import "fmt"

const (
	UpdateStatusPropertyName    = "update_status"
	GetConfirmationPropertyName = "get_confirmation"
	SubscribePropertyName       = "subscribe"
	AutoRespondPropertyName     = "auto_respond"
)

func (s *pluginStore) SetProperty(userID, propertyName string, value bool) error {
	user, err := s.LoadUser(userID)
	if err != nil {
		return err
	}

	switch propertyName {
	case UpdateStatusPropertyName:
		user.Settings.UpdateStatus = value
	case GetConfirmationPropertyName:
		user.Settings.GetConfirmation = value
	case AutoRespondPropertyName:
		user.Settings.AutoRespond = value
	default:
		return fmt.Errorf("property %s not found", propertyName)
	}

	err = s.StoreUser(user)
	if err != nil {
		return err
	}

	return nil
}

func (s *pluginStore) SetPostID(userID, propertyName, postID string) error {
	user, err := s.LoadUser(userID)
	if err != nil {
		return err
	}

	switch propertyName {
	case UpdateStatusPropertyName:
		user.WelcomeFlowStatus.UpdateStatusPostID = postID
	case GetConfirmationPropertyName:
		user.WelcomeFlowStatus.GetConfirmationPostID = postID
	case SubscribePropertyName:
		user.WelcomeFlowStatus.SubscribePostID = postID
	case AutoRespondPropertyName:
		user.WelcomeFlowStatus.AutoRespondPostID = postID
	default:
		return fmt.Errorf("property %s not found", propertyName)
	}

	err = s.StoreUser(user)
	if err != nil {
		return err
	}

	return nil
}

func (s *pluginStore) GetPostID(userID, propertyName string) (string, error) {
	user, err := s.LoadUser(userID)
	if err != nil {
		return "", err
	}

	switch propertyName {
	case UpdateStatusPropertyName:
		return user.WelcomeFlowStatus.UpdateStatusPostID, nil
	case GetConfirmationPropertyName:
		return user.WelcomeFlowStatus.GetConfirmationPostID, nil
	case SubscribePropertyName:
		return user.WelcomeFlowStatus.SubscribePostID, nil
	case AutoRespondPropertyName:
		return user.WelcomeFlowStatus.AutoRespondPostID, nil
	default:
		return "", fmt.Errorf("property %s not found", propertyName)
	}
}

func (s *pluginStore) RemovePostID(userID, propertyName string) error {
	user, err := s.LoadUser(userID)
	if err != nil {
		return err
	}

	switch propertyName {
	case UpdateStatusPropertyName:
		user.WelcomeFlowStatus.UpdateStatusPostID = ""
	case GetConfirmationPropertyName:
		user.WelcomeFlowStatus.GetConfirmationPostID = ""
	case SubscribePropertyName:
		user.WelcomeFlowStatus.SubscribePostID = ""
	case AutoRespondPropertyName:
		user.WelcomeFlowStatus.AutoRespondPostID = ""
	default:
		return fmt.Errorf("property %s not found", propertyName)
	}

	err = s.StoreUser(user)
	if err != nil {
		return err
	}

	return nil
}

func (s *pluginStore) GetCurrentStep(userID string) (int, error) {
	user, err := s.LoadUser(userID)
	if err != nil {
		return 0, err
	}

	return user.WelcomeFlowStatus.Step, nil
}
func (s *pluginStore) SetCurrentStep(userID string, step int) error {
	user, err := s.LoadUser(userID)
	if err != nil {
		return err
	}

	user.WelcomeFlowStatus.Step = step

	err = s.StoreUser(user)
	if err != nil {
		return err
	}

	return nil
}

func (s *pluginStore) DeleteCurrentStep(userID string) error {
	user, err := s.LoadUser(userID)
	if err != nil {
		return err
	}

	user.WelcomeFlowStatus.Step = 0

	err = s.StoreUser(user)
	if err != nil {
		return err
	}

	return nil
}
