package settingspanel

import (
	"errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-server/v5/model"
)

type Setting interface {
	Set(userID string, value string) error
	Get(userID string) (string, error)
	GetID() string
	GetDependency() string
	IsDisabled(foreignValue string) bool
	GetTitle() string
	GetDescription() string
	ToPost(userID, settingHandler string, disabled bool) (*model.Post, error)
	GetSlackAttachments(userID, settngHandler string, disabled bool) (*model.SlackAttachment, error)
}

type Panel interface {
	Set(userID, settingID string, value string) error
	Print(userID string)
	Update(userID string) error
	Clear(userID string) error
	URL() string
	GetSettingIDs() []string
}

type SettingStore interface {
	SetSetting(userID, settingID string, value interface{}) error
	GetSetting(userID, settingID string) (interface{}, error)
}

type PanelStore interface {
	SetPanelPostIDs(userID string, postIDs map[string]string) error
	GetPanelPostIDs(userID string) (map[string]string, error)
	DeletePanelPostIDs(userID string) error
}

type panel struct {
	settings       map[string]Setting
	settingKeys    []string
	poster         bot.Poster
	logger         bot.Logger
	store          PanelStore
	settingHandler string
	pluginURL      string
}

func NewSettingsPanel(settings []Setting, poster bot.Poster, logger bot.Logger, store PanelStore, settingHandler, pluginURL string) Panel {
	settingsMap := make(map[string]Setting)
	settingKeys := []string{}
	for _, s := range settings {
		settingsMap[s.GetID()] = s
		settingKeys = append(settingKeys, s.GetID())
	}

	return &panel{
		settings:       settingsMap,
		settingKeys:    settingKeys,
		poster:         poster,
		logger:         logger,
		store:          store,
		settingHandler: settingHandler,
		pluginURL:      pluginURL,
	}
}

func (p *panel) Set(userID, settingID string, value string) error {
	s, ok := p.settings[settingID]
	if !ok {
		return errors.New("cannot find setting " + settingID)
	}

	err := s.Set(userID, value)
	if err != nil {
		return err
	}
	return nil
}

func (p *panel) GetSettingIDs() []string {
	return p.settingKeys
}

func (p *panel) URL() string {
	return p.settingHandler
}

func (p *panel) Print(userID string) {
	err := p.cleanPreviousSettingsPosts(userID)
	if err != nil {
		p.logger.Errorf("could not clean previous setting post")
	}

	postIDs := make(map[string]string)
	for _, key := range p.settingKeys {
		s := p.settings[key]
		sa, loopErr := s.GetSlackAttachments(userID, p.pluginURL+p.settingHandler, p.isSettingDisabled(userID, s))
		if loopErr != nil {
			p.logger.Errorf("error creating the slack attachment, err=" + loopErr.Error())
			continue
		}
		postID, loopErr := p.poster.DMWithAttachments(userID, sa)
		if loopErr != nil {
			p.logger.Errorf("error creating the message, err=", loopErr.Error())
			continue
		}
		postIDs[s.GetID()] = postID
	}
	err = p.store.SetPanelPostIDs(userID, postIDs)
	if err != nil {
		p.logger.Errorf("could not set the post IDs, err=", err.Error())
	}
}

func (p *panel) Update(userID string) error {
	postIDs, err := p.store.GetPanelPostIDs(userID)
	if err != nil {
		return err
	}

	for _, s := range p.settings {
		post, err := s.ToPost(userID, p.pluginURL+p.settingHandler, p.isSettingDisabled(userID, s))
		if err != nil {
			p.logger.Errorf("error creating the slack attachment for setting %s, err=%s", s.GetID(), err.Error())
			continue
		}
		post.Id = postIDs[s.GetID()]
		err = p.poster.DMUpdatePost(post)
		if err != nil {
			p.logger.Errorf("error updating the post for setting %s, err=%s", s.GetID(), err.Error())
		}
	}

	return nil
}

func (p *panel) cleanPreviousSettingsPosts(userID string) error {
	postIDs, err := p.store.GetPanelPostIDs(userID)
	if err != nil {
		return err
	}

	for _, v := range postIDs {
		err = p.poster.DeletePost(v)
		if err != nil {
			p.logger.Errorf("could not delete setting post")
		}
	}

	err = p.store.DeletePanelPostIDs(userID)
	if err != nil {
		return err
	}

	return nil
}

func (p *panel) Clear(userID string) error {
	return p.cleanPreviousSettingsPosts(userID)
}

func (p *panel) isSettingDisabled(userID string, s Setting) bool {
	dependencyID := s.GetDependency()
	if dependencyID == "" {
		return false
	}
	dependency, ok := p.settings[dependencyID]
	if !ok {
		p.logger.Errorf("settings dependency %s not found", dependencyID)
		return false
	}

	value, err := dependency.Get(userID)
	if err != nil {
		p.logger.Errorf("cannot get dependency %s value", dependencyID)
		return false
	}
	return s.IsDisabled(value)
}
