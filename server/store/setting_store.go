package store

import (
	"fmt"
	"strings"
)

const (
	UpdateStatusSettingID               = "update_status"
	GetConfirmationSettingID            = "get_confirmation"
	ReceiveNotificationsDuringMeetingID = "receive_notification"
	ReceiveRemindersSettingID           = "get_reminders"
	DailySummarySettingID               = "summary_setting"
)

func (s *pluginStore) SetSetting(userID, settingID string, value interface{}) error {
	user, err := s.LoadUser(userID)
	if err != nil {
		return err
	}

	switch settingID {
	case UpdateStatusSettingID:
		storableValue, ok := value.(bool)
		if !ok {
			return fmt.Errorf("cannot read value %v for setting %s (expecting bool)", value, settingID)
		}
		user.Settings.UpdateStatus = storableValue
		s.Tracker.TrackAutomaticStatusUpdate(userID, storableValue, "settings")
	case GetConfirmationSettingID:
		storableValue, ok := value.(bool)
		if !ok {
			return fmt.Errorf("cannot read value %v for setting %s (expecting bool)", value, settingID)
		}
		user.Settings.GetConfirmation = storableValue
	case ReceiveNotificationsDuringMeetingID:
		storableValue, ok := value.(bool)
		if !ok {
			return fmt.Errorf("cannot read value %v for setting %s (expecting bool)", value, settingID)
		}
		user.Settings.ReceiveNotificationsDuringMeeting = storableValue
	case ReceiveRemindersSettingID:
		storableValue, ok := value.(bool)
		if !ok {
			return fmt.Errorf("cannot read value %v for setting %s (expecting bool)", value, settingID)
		}
		user.Settings.ReceiveReminders = storableValue
	case DailySummarySettingID:
		s.updateDailySummarySettingForUser(user, value)
	default:
		return fmt.Errorf("setting %s not found", settingID)
	}

	err = s.StoreUser(user)
	if err != nil {
		return err
	}

	return nil
}

func (s *pluginStore) GetSetting(userID, settingID string) (interface{}, error) {
	user, err := s.LoadUser(userID)
	if err != nil {
		return nil, err
	}

	switch settingID {
	case UpdateStatusSettingID:
		return user.Settings.UpdateStatus, nil
	case GetConfirmationSettingID:
		return user.Settings.GetConfirmation, nil
	case ReceiveNotificationsDuringMeetingID:
		return user.Settings.ReceiveNotificationsDuringMeeting, nil
	case ReceiveRemindersSettingID:
		return user.Settings.ReceiveReminders, nil
	case DailySummarySettingID:
		dsum := user.Settings.DailySummary
		return dsum, nil
	default:
		return nil, fmt.Errorf("setting %s not found", settingID)
	}
}

func DefaultDailySummaryUserSettings() *DailySummaryUserSettings {
	return &DailySummaryUserSettings{
		PostTime: "8:00AM",
		Timezone: "Eastern Standard Time",
		Enable:   false,
	}
}
func (s *pluginStore) updateDailySummarySettingForUser(user *User, value interface{}) error {
	if user.Settings.DailySummary == nil {
		user.Settings.DailySummary = DefaultDailySummaryUserSettings()
	}

	dsum := user.Settings.DailySummary

	stringValue := value.(string)
	splittedValue := strings.Split(stringValue, " ")
	timezone := strings.Join(splittedValue[1:], " ")

	switch splittedValue[0] {
	case "true":
		dsum.Enable = true
	case "false":
		dsum.Enable = false
	default:
		dsum.PostTime = splittedValue[0]
		dsum.Timezone = timezone
	}

	return nil
}

func (s *pluginStore) SetPanelPostID(userID string, postID string) error {
	err := s.settingsPanelKV.Store(userID, []byte(postID))
	if err != nil {
		return err
	}
	return nil
}

func (s *pluginStore) GetPanelPostID(userID string) (string, error) {
	postID, err := s.settingsPanelKV.Load(userID)
	if err != nil {
		return "", err
	}

	return string(postID), nil
}

func (s *pluginStore) DeletePanelPostID(userID string) error {
	err := s.settingsPanelKV.Delete(userID)
	if err != nil {
		return err
	}
	return nil
}
