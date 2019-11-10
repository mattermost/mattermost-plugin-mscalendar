// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package plugin

import (
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
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote/msgraph"
	"github.com/mattermost/mattermost-plugin-msoffice/server/store"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/bot"
)

type Plugin struct {
	plugin.MattermostPlugin
	configLock  *sync.RWMutex
	config      *config.Config
	httpHandler *http.Handler

	// KVStore           kvstore.KVStore
	UserStore         store.UserStore
	OAuth2StateStore  store.OAuth2StateStore
	SubscriptionStore store.SubscriptionStore
	Remote            remote.Remote

	Templates map[string]*template.Template
}

func NewWithConfig(conf *config.Config) *Plugin {
	return &Plugin{
		configLock: &sync.RWMutex{},
		config:     conf,
	}
}

func (p *Plugin) OnActivate() error {
	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "couldn't get bundle path")
	}

	// Set up or update the plugin's bot account
	botUserID, err := bot.EnsureWithProfileImage(
		p.API, p.Helpers,
		config.BotUserName, config.BotDisplayName, config.BotDescription,
		filepath.Join(bundlePath, "assets", "profile.png"))
	if err != nil {
		return err
	}
	p.updateConfig(func(conf *config.Config) {
		conf.BotUserID = botUserID
	})

	// Templates
	err = p.loadTemplates(bundlePath)
	if err != nil {
		return err
	}

	// API dependencies
	store := store.NewPluginStore(p.API)
	p.UserStore = store
	p.OAuth2StateStore = store
	p.SubscriptionStore = store

	// Handlers
	command.Register(p.API.RegisterCommand)

	p.httpHandler = p.newHTTPHandler(&(*p.config))

	p.API.LogInfo(p.config.PluginID + " activated")
	return nil
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	conf := p.getConfig()
	stored := config.StoredConfig{}
	err := p.API.LoadPluginConfiguration(&stored)
	if err != nil {
		return errors.WithMessage(err, "failed to load plugin configuration")
	}

	if stored.OAuth2Authority == "" ||
		stored.OAuth2ClientID == "" ||
		stored.OAuth2ClientSecret == "" {
		return errors.WithMessage(err, "failed to configure: OAuth2 credentials to be set in the config")
	}

	mattermostSiteURL := p.API.GetConfig().ServiceSettings.SiteURL
	if mattermostSiteURL == nil {
		return errors.New("plugin requires Mattermost Site URL to be set")
	}
	mattermostURL, err := url.Parse(*mattermostSiteURL)
	if err != nil {
		return err
	}
	pluginURLPath := "/plugins/" + conf.PluginID
	pluginURL := strings.TrimRight(*mattermostSiteURL, "/") + pluginURLPath

	p.updateConfig(func(c *config.Config) {
		c.StoredConfig = stored

		c.MattermostSiteURL = *mattermostSiteURL
		c.MattermostSiteHostname = mattermostURL.Hostname()
		c.PluginURL = pluginURL
		c.PluginURLPath = pluginURLPath

		clone := &(*c)
		p.httpHandler = p.newHTTPHandler(clone)
		p.Remote = remote.Makers[msgraph.Kind](clone, p.API)
	})

	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	conf := p.getConfig()

	h := command.Handler{
		Config:            conf,
		UserStore:         p.UserStore,
		SubscriptionStore: p.SubscriptionStore,
		Poster:            bot.NewPoster(p.API, conf),
		Logger:            p.API,
		IsAuthorizedAdmin: p.IsAuthorizedAdmin,
		Remote:            p.Remote,
		Context:           c,
		Args:              args,
		ChannelID:         args.ChannelId,
		MattermostUserID:  args.UserId,
	}
	out, err := h.Handle()
	if err != nil {
		p.API.LogError(err.Error())
		return nil, model.NewAppError("msofficeplugin.ExecuteCommand", "Unable to execute command.", nil, err.Error(), gohttp.StatusInternalServerError)
	}
	h.Poster.PostEphemeral(args.UserId, args.ChannelId, out)

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
		OAuth2StateStore:  p.OAuth2StateStore,
		SubscriptionStore: p.SubscriptionStore,
		Poster:            bot.NewPoster(p.API, conf),
		Logger:            p.API,
		IsAuthorizedAdmin: p.IsAuthorizedAdmin,
		Remote:            p.Remote,
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

func (p *Plugin) loadTemplates(bundlePath string) error {
	if p.Templates != nil {
		return nil
	}

	templatesPath := filepath.Join(bundlePath, "assets", "templates")
	templates := make(map[string]*template.Template)
	err := filepath.Walk(templatesPath, func(path string, info os.FileInfo, err error) error {
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
		key := path[len(templatesPath):]
		templates[key] = template
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "OnActivate/loadTemplates failed")
	}
	p.Templates = templates
	return nil
}
