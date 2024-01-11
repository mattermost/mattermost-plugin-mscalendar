package flow

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v6/model"
)

type EmptyStep struct {
	Title   string
	Message string
}

func (s *EmptyStep) PostSlackAttachment(flowHandler string, i int) *model.SlackAttachment {
	sa := model.SlackAttachment{
		Title:    s.Title,
		Text:     s.Message,
		Fallback: fmt.Sprintf("%s: %s", s.Title, s.Message),
	}

	return &sa
}

func (s *EmptyStep) ResponseSlackAttachment(value bool) *model.SlackAttachment {
	return nil
}

func (s *EmptyStep) GetPropertyName() string {
	return ""
}

func (s *EmptyStep) ShouldSkip(value bool) int {
	return 0
}

func (s *EmptyStep) IsEmpty() bool {
	return true
}
