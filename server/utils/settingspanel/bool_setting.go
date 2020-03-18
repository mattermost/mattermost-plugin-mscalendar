package settingspanel

import (
	"errors"
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
)

type boolSetting struct {
	title       string
	description string
	id          string
	dependsOn   string
	store       SettingStore
}

func NewBoolSetting(id string, title string, description string, dependsOn string, store SettingStore) Setting {
	return &boolSetting{
		title:       title,
		description: description,
		id:          id,
		dependsOn:   dependsOn,
		store:       store,
	}
}

func (s *boolSetting) Set(userID string, value string) error {
	boolValue := false
	if value == "true" {
		boolValue = true
	}

	err := s.store.SetSetting(userID, s.id, boolValue)
	if err != nil {
		return err
	}

	return nil
}

func (s *boolSetting) Get(userID string) (string, error) {
	value, err := s.store.GetSetting(userID, s.id)
	if err != nil {
		return "", err
	}
	boolValue, ok := value.(bool)
	if !ok {
		return "", errors.New("current value is not a bool")
	}

	stringValue := "false"
	if boolValue {
		stringValue = "true"
	}

	return stringValue, nil
}

func (s *boolSetting) GetID() string {
	return s.id
}

func (s *boolSetting) GetTitle() string {
	return s.title
}

func (s *boolSetting) GetDescription() string {
	return s.description
}

func (s *boolSetting) GetDependency() string {
	return s.dependsOn
}

func (s *boolSetting) ToPost(userID string, settingHandler string, disabled bool) (*model.Post, error) {
	sa, err := s.GetSlackAttachments(userID, settingHandler, disabled)
	if err != nil {
		return nil, err
	}

	post := model.Post{}
	model.ParseSlackAttachment(&post, []*model.SlackAttachment{sa})

	return &post, nil
}

func (s *boolSetting) GetSlackAttachments(userID, settingHandler string, disabled bool) (*model.SlackAttachment, error) {
	title := fmt.Sprintf("Setting: %s", s.title)
	currentValueMessage := "Disabled"

	actions := []*model.PostAction{}
	if !disabled {
		currentValue, err := s.Get(userID)
		if err != nil {
			return nil, err
		}

		currentTextValue := "No"
		if currentValue == "true" {
			currentTextValue = "Yes"
		}
		currentValueMessage = fmt.Sprintf("Current value: %s", currentTextValue)

		actionTrue := model.PostAction{
			Name: "Yes",
			Integration: &model.PostActionIntegration{
				URL: settingHandler + "?" + s.id + "=true",
			},
		}

		actionFalse := model.PostAction{
			Name: "No",
			Integration: &model.PostActionIntegration{
				URL: settingHandler + "?" + s.id + "=false",
			},
		}
		actions = []*model.PostAction{&actionTrue, &actionFalse}

	}

	text := fmt.Sprintf("%s\n%s", s.description, currentValueMessage)
	sa := model.SlackAttachment{
		Title:   title,
		Text:    text,
		Actions: actions,
	}

	return &sa, nil
}

func (s *boolSetting) IsDisabled(foreignValue string) bool {
	return foreignValue == "false"
}
