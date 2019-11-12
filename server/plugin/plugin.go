// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package plugin

import (
	"fmt"
	gohttp "net/http"
	"net/url"
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
	"github.com/mattermost/mattermost-plugin-msoffice/server/user"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

type Plugin struct {
	plugin.MattermostPlugin
	configLock  *sync.RWMutex
	config      *config.Config
	httpHandler *http.Handler

	KVStore          kvstore.KVStore
	UserStore        user.Store
	OAuth2StateStore user.OAuth2StateStore

	Templates map[string]*template.Template
}

func NewWithConfig(conf *config.Config) *Plugin {
	return &Plugin{
		configLock: &sync.RWMutex{},
		config:     conf,
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

	kv := kvstore.NewPluginStore(p.API)
	p.KVStore = kv
	p.UserStore = user.NewStore(kv)
	p.OAuth2StateStore = user.NewOAuth2StateStore(p.API)

	command.Register(p.API.RegisterCommand)
	p.httpHandler = p.newHTTPHandler(&(*p.config))

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

	if newStored.OAuth2Authority == "" ||
		newStored.OAuth2ClientId == "" ||
		newStored.OAuth2ClientSecret == "" {
		return errors.WithMessage(err, "failed to configure: OAuth2 credentials to be set in the config")
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
	mattermostURL, err := url.Parse(mattermostSiteURL)
	if err != nil {
		return err
	}
	pluginURLPath := "/plugins/" + conf.PluginId
	pluginURL := strings.TrimRight(mattermostSiteURL, "/") + pluginURLPath

	p.updateConfig(func(c *config.Config) {
		c.StoredConfig = newStored

		// TODO Update c.BotIconURL = ""
		c.BotUserId = botUserId
		c.MattermostSiteURL = mattermostSiteURL
		c.MattermostSiteHostname = mattermostURL.Hostname()
		c.PluginURL = pluginURL
		c.PluginURLPath = pluginURLPath

		p.httpHandler = p.newHTTPHandler(&(*c))
	})

	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	conf := p.getConfig()
	poster := utils.NewBotPoster(conf, p.API)

	h := command.Handler{
		Config:            conf,
		UserStore:         p.UserStore,
		API:               p.API,
		BotPoster:         utils.NewBotPoster(conf, p.API),
		IsAuthorizedAdmin: p.IsAuthorizedAdmin,
		Context:           c,
		Args:              args,
		ChannelId:         args.ChannelId,
		MattermostUserId:  args.UserId,
	}
	out, err := h.Handle()
	if err != nil {
		p.API.LogError(err.Error())
		return nil, model.NewAppError("msofficeplugin.ExecuteCommand", "Unable to execute command.", nil, err.Error(), gohttp.StatusInternalServerError)
	}
	poster.PostEphemeral(args.UserId, args.ChannelId, out)

	return &model.CommandResponse{}, nil
}

func (p *Plugin) ServeHTTP(pc *plugin.Context, w gohttp.ResponseWriter, r *gohttp.Request) {
	p.configLock.RLock()
	handler := p.httpHandler
	p.configLock.RUnlock()

	handler.ServeHTTP(w, r)
}

func (p *Plugin) newHTTPHandler(conf *config.Config) *http.Handler {
	h := &http.Handler{
		Config:            conf,
		UserStore:         p.UserStore,
		API:               p.API,
		BotPoster:         utils.NewBotPoster(conf, p.API),
		IsAuthorizedAdmin: p.IsAuthorizedAdmin,
		OAuth2StateStore:  p.OAuth2StateStore,
	}
	h.InitRouter()
	return h
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
