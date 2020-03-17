package flow

import (
	"strconv"

	"github.com/mattermost/mattermost-server/v5/model"
)

type Flow interface {
	Step(i int) Step
	URL() string
	Length() int
	StepDone(userID string, value bool)
	FlowDone(userID string)
}

type FlowStore interface {
	SetProperty(userID, propertyName string, value bool) error
	SetPostID(userID, propertyName, postID string) error
	GetPostID(userID, propertyName string) (string, error)
	RemovePostID(userID, propertyName string) error
}

type Step interface {
	PostSlackAttachment(flowHandler string, i int) *model.SlackAttachment
	ResponseSlackAttachment(value bool) *model.SlackAttachment
	PropertyName() string
	ShouldSkip(value bool) int
}

type step struct {
	title                string
	message              string
	propertyName         string
	trueButtonMessage    string
	falseButtonMessage   string
	trueResponseMessage  string
	falseResponseMessage string
	trueSkip             int
	falseSkip            int
}

func NewStep(
	title,
	message,
	propertyName,
	trueButtonMessage,
	falseButtonMessage,
	trueResponseMessage,
	falseResponseMessage string,
	trueSkip,
	falseSkip int,
) Step {
	return &step{
		title:                title,
		message:              message,
		propertyName:         propertyName,
		trueButtonMessage:    trueButtonMessage,
		falseButtonMessage:   falseButtonMessage,
		trueResponseMessage:  trueResponseMessage,
		falseResponseMessage: falseResponseMessage,
		trueSkip:             trueSkip,
		falseSkip:            falseSkip,
	}
}

func (s *step) PostSlackAttachment(flowHandler string, i int) *model.SlackAttachment {
	actionTrue := model.PostAction{
		Name: s.trueButtonMessage,
		Integration: &model.PostActionIntegration{
			URL: flowHandler + "?" + s.propertyName + "=true&step=" + strconv.Itoa(i),
		},
	}

	actionFalse := model.PostAction{
		Name: s.falseButtonMessage,
		Integration: &model.PostActionIntegration{
			URL: flowHandler + "?" + s.propertyName + "=false&step=" + strconv.Itoa(i),
		},
	}

	sa := model.SlackAttachment{
		Title:   s.title,
		Text:    s.message,
		Actions: []*model.PostAction{&actionTrue, &actionFalse},
	}

	return &sa
}

func (s *step) ResponseSlackAttachment(value bool) *model.SlackAttachment {
	message := s.falseResponseMessage
	if value {
		message = s.trueResponseMessage
	}

	sa := model.SlackAttachment{
		Title:   s.title,
		Text:    message,
		Actions: []*model.PostAction{},
	}

	return &sa
}

func (s *step) PropertyName() string {
	return s.propertyName
}

func (s *step) ShouldSkip(value bool) int {
	if value {
		return s.trueSkip
	}

	return s.falseSkip
}
