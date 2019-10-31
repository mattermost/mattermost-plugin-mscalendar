// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package plugin

import (
	"fmt"
	gohttp "net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"

	"github.com/mattermost/mattermost-plugin-msoffice/server/command"
	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/http"
	"github.com/mattermost/mattermost-plugin-msoffice/server/kvstore"
)

type Plugin struct {
	plugin.MattermostPlugin
	configLock *sync.RWMutex
	config     *config.Config

	httpProtoHandler *http.Handler

	Templates map[string]*template.Template
}

func NewWithConfig(conf *config.Config) *Plugin {
	return &Plugin{
		configLock:       &sync.RWMutex{},
		config:           conf,
		httpProtoHandler: http.NewProtoHandler(),
	}
}

func (p *Plugin) OnActivate() error {
	// if p.Templates == nil {
	// 	templatesPath := filepath.Join(*(p.API.GetConfig().PluginSettings.Directory),
	// 		p.config.PluginId, "server", "dist", "templates")
	// 	templates, err := loadTemplates(templatesPath)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	p.Templates = templates
	// }

	p.config.ImportedAPI = config.ImportedAPI{
		Helpers:           p.Helpers,
		PAPI:              p.API,
		KVStore:           kvstore.NewPluginStore(p.API),
		IsAuthorizedAdmin: p.IsAuthorizedAdmin,
	}

	command.Init(p.API.RegisterCommand)

	p.API.LogInfo(p.config.PluginId + " activated")
	return nil
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	conf := p.getConfig()
	oldStored := conf.StoredConfig
	newStored := config.StoredConfig{}
	err := p.API.LoadPluginConfiguration(&newStored)
	if err != nil {
		return errors.WithMessage(err, "failed to load plugin configuration")
	}

	botUserId := conf.BotUserId
	if newStored.BotUserName != oldStored.BotUserName {
		user, appErr := p.API.GetUserByUsername(newStored.BotUserName)
		if appErr != nil {
			return errors.WithMessage(appErr, fmt.Sprintf("unable to load user %s", newStored.BotUserName))
		}
		botUserId = user.Id
	}

	mattermostSiteURL := *p.API.GetConfig().ServiceSettings.SiteURL
	pluginURLPath := "/plugins/" + conf.PluginId
	pluginURL := strings.TrimRight(mattermostSiteURL, "/") + pluginURLPath

	p.updateConfig(func(c *config.Config) {
		// TODO Update c.BotIconURL = ""
		c.BotUserId = botUserId

		c.MattermostSiteURL = mattermostSiteURL
		c.PluginURL = pluginURL
		c.PluginURLPath = pluginURLPath
	})

	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	conf := p.getConfig()

	h := command.Handler{
		Config:    conf,
		Context:   c,
		Args:      args,
		ChannelId: args.ChannelId,
	}
	out, err := h.Handle()
	if err != nil {
		p.API.LogError(err.Error())
		return nil, model.NewAppError("msofficeplugin.ExecuteCommand", "Unable to execute command.", nil, err.Error(), gohttp.StatusInternalServerError)
	}

	post := &model.Post{
		UserId:    conf.BotUserId,
		ChannelId: args.ChannelId,
		Message:   out,
	}
	_ = conf.PAPI.SendEphemeralPost(args.UserId, post)
	return &model.CommandResponse{}, nil
}

func (p *Plugin) ServeHTTP(pc *plugin.Context, w gohttp.ResponseWriter, r *gohttp.Request) {
	conf := p.getConfig()
	p.httpProtoHandler.CloneWithConfig(conf).ServeHTTP(w, r)
}

func (p *Plugin) getConfig() *config.Config {
	p.configLock.RLock()
	defer p.configLock.RUnlock()
	return &(*p.config)
}

func (p *Plugin) updateConfig(f func(*config.Config)) config.Config {
	p.configLock.Lock()
	defer p.configLock.Unlock()

	f(p.config)
	return *p.config
}

func loadTemplates(dir string) (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		template, err := template.ParseFiles(path)
		if err != nil {
			return nil
		}
		key := path[len(dir):]
		templates[key] = template
		return nil
	})
	if err != nil {
		return nil, errors.WithMessage(err, "OnActivate: failed to load templates")
	}
	return templates, nil
}
