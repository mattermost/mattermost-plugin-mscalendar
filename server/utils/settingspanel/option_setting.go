package settingspanel

import (
	"errors"
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
)

type optionSetting struct {
	title       string
	description string
	id          string
	dependsOn   string
	options     []string
	store       SettingStore
}

func NewOptionSetting(id string, title string, description string, dependsOn string, options []string, store SettingStore) Setting {
	return &optionSetting{
		title:       title,
		description: description,
		id:          id,
		dependsOn:   dependsOn,
		options:     options,
		store:       store,
	}
}

func (s *optionSetting) Set(userID string, value string) error {
	err := s.store.SetSetting(userID, s.id, value)
	if err != nil {
		return err
	}

	return nil
}

func (s *optionSetting) Get(userID string) (string, error) {
	value, err := s.store.GetSetting(userID, s.id)
	if err != nil {
		return "", err
	}
	valueString, ok := value.(string)
	if !ok {
		return "", errors.New("current value is not a string")
	}

	return valueString, nil
}

func (s *optionSetting) GetID() string {
	return s.id
}

func (s *optionSetting) GetTitle() string {
	return s.title
}

func (s *optionSetting) GetDescription() string {
	return s.description
}

func (s *optionSetting) GetDependency() string {
	return s.dependsOn
}

func (s *optionSetting) ToPost(userID, settingHandler string, disabled bool) (*model.Post, error) {
	sa, err := s.GetSlackAttachments(userID, settingHandler, disabled)
	if err != nil {
		return nil, err
	}

	post := model.Post{}
	model.ParseSlackAttachment(&post, []*model.SlackAttachment{sa})

	return &post, nil
}

func (s *optionSetting) GetSlackAttachments(userID, settingHandler string, disabled bool) (*model.SlackAttachment, error) {
	title := fmt.Sprintf("Setting: %s", s.title)
	currentValueMessage := "Disabled"

	actions := []*model.PostAction{}
	if !disabled {
		currentTextValue, err := s.Get(userID)
		if err != nil {
			return nil, err
		}
		currentValueMessage = fmt.Sprintf("Current value: %s", currentTextValue)

		actionOptions := model.PostAction{
			Name: "Select an option:",
			Integration: &model.PostActionIntegration{
				URL: settingHandler + "?" + s.id + "=true",
			},
			Type:    "select",
			Options: stringsToOptions(s.options),
		}

		actions = []*model.PostAction{&actionOptions}
	}

	text := fmt.Sprintf("%s\n%s", s.description, currentValueMessage)
	sa := model.SlackAttachment{
		Title:   title,
		Text:    text,
		Actions: actions,
	}
	return &sa, nil
}

func (s *optionSetting) IsDisabled(foreignValue string) bool {
	return foreignValue == "false"
}
