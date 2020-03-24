package store

import "fmt"

const (
	UpdateStatusPropertyName    = "update_status"
	GetConfirmationPropertyName = "get_confirmation"
	SubscribePropertyName       = "subscribe"
)

func (s *pluginStore) SetProperty(userID, propertyName string, value bool) error {
	user, err := s.LoadUser(userID)
	if err != nil {
		return err
	}

	switch propertyName {
	case UpdateStatusPropertyName:
		user.Flags.WelcomeUpdateStatus = value
	case GetConfirmationPropertyName:
		user.Flags.WelcomeGetConfirmation = value
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
		user.Flags.WelcomeUpdateStatusPostID = postID
	case GetConfirmationPropertyName:
		user.Flags.WelcomeGetConfirmationPostID = postID
	case SubscribePropertyName:
		user.Flags.WelcomeSubscribePostID = postID
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
		return user.Flags.WelcomeUpdateStatusPostID, nil
	case GetConfirmationPropertyName:
		return user.Flags.WelcomeGetConfirmationPostID, nil
	case SubscribePropertyName:
		return user.Flags.WelcomeSubscribePostID, nil
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
		user.Flags.WelcomeUpdateStatusPostID = ""
	case GetConfirmationPropertyName:
		user.Flags.WelcomeGetConfirmationPostID = ""
	case SubscribePropertyName:
		user.Flags.WelcomeSubscribePostID = ""
	default:
		return fmt.Errorf("property %s not found", propertyName)
	}

	err = s.StoreUser(user)
	if err != nil {
		return err
	}

	return nil
}
