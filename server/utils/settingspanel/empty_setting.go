package settingspanel

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
)

type emptySetting struct {
	title       string
	description string
	id          string
}

func NewEmptySetting(id, title, description string) Setting {
	return &emptySetting{
		id:          id,
		title:       title,
		description: description,
	}
}

func (s *emptySetting) Set(userID string, value string) error {
	return nil
}
func (s *emptySetting) Get(userID string) (string, error) {
	return "", nil
}
func (s *emptySetting) GetID() string {
	return s.id
}
func (s *emptySetting) GetDependency() string {
	return ""
}
func (s *emptySetting) IsDisabled(foreignValue string) bool {
	return false
}
func (s *emptySetting) GetTitle() string {
	return s.title
}
func (s *emptySetting) GetDescription() string {
	return s.description
}
func (s *emptySetting) GetSlackAttachments(userID, settngHandler string, disabled bool) (*model.SlackAttachment, error) {
	title := fmt.Sprintf("Setting: %s", s.title)
	sa := model.SlackAttachment{
		Title: title,
		Text:  s.description,
	}

	return &sa, nil
}
