package store

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/kvstore"
)

const (
	UpdateStatusSettingID    = "update_status"
	GetConfirmationSettingID = "get_confirmation"
	DailySummarySettingID    = "summary_setting"
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
	case GetConfirmationSettingID:
		storableValue, ok := value.(bool)
		if !ok {
			return fmt.Errorf("cannot read value %v for setting %s (expecting bool)", value, settingID)
		}
		user.Settings.GetConfirmation = storableValue
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
	case DailySummarySettingID:
		return s.getDailySummarySettingForUser(userID)
	default:
		return nil, fmt.Errorf("setting %s not found", settingID)
	}
}

func (s *pluginStore) getDailySummarySettingForUser(userID string) (string, error) {
	dsumIndex, err := s.LoadDailySummaryIndex()
	if err != nil {
		return "", err
	}

	for _, dsum := range dsumIndex {
		if dsum.MattermostUserID == userID {
			if !dsum.Enable {
				return "Daily summary not set", nil
			}
			return fmt.Sprintf("Daily summary set at %s (%s)", dsum.PostTime, dsum.Timezone), nil
		}
	}

	return "Daily summary not set", nil
}

func (s *pluginStore) SetPanelPostID(userID string, postID string) error {
	err := kvstore.StoreJSON(s.settingsPanelKV, userID, postID)
	if err != nil {
		return err
	}
	return nil
}

func (s *pluginStore) GetPanelPostID(userID string) (string, error) {
	var postID string
	err := kvstore.LoadJSON(s.settingsPanelKV, userID, &postID)
	if err != nil {
		return "", err
	}

	return postID, nil
}

func (s *pluginStore) DeletePanelPostID(userID string) error {
	err := s.settingsPanelKV.Delete(userID)
	if err != nil {
		return err
	}
	return nil
}
