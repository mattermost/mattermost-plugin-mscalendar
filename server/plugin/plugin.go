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

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-msoffice/server/api"
	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote/msgraph"
	"github.com/mattermost/mattermost-plugin-msoffice/server/store"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/bot"

	"github.com/mattermost/mattermost-plugin-msoffice/server/plugin/command"
	"github.com/mattermost/mattermost-plugin-msoffice/server/plugin/http"
)

type Plugin struct {
	plugin.MattermostPlugin
	configLock *sync.RWMutex
	config     *config.Config

	httpHandler         *http.Handler
	notificationHandler api.NotificationHandler

	Templates map[string]*template.Template
}

func NewWithConfig(conf *config.Config) *Plugin {
	return &Plugin{
		configLock: &sync.RWMutex{},
		config:     conf,
	}
}

func (p *Plugin) OnActivate() error {
	botUserID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    config.BotUserName,
		DisplayName: config.BotDisplayName,
		Description: config.BotDescription,
	}, plugin.ProfileImagePath("assets/profile.png"))
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}
	p.config.BotUserID = botUserID

	// Templates
	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "couldn't get bundle path")
	}
	err = p.loadTemplates(bundlePath)
	if err != nil {
		return err
	}

	p.httpHandler = http.NewHandler()
	p.notificationHandler = api.NewNotificationHandler(p.getAPIConfig())

	command.Register(p.API.RegisterCommand)

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
	})

	if p.notificationHandler != nil {
		p.notificationHandler.Configure(p.getAPIConfig())
	}
	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	apiconf := p.getAPIConfig()
	api := api.New(apiconf, args.UserId)
	command := command.Command{
		Context:   c,
		Args:      args,
		ChannelID: args.ChannelId,
		Config:    apiconf.Config,
		API:       api,
	}

	out, err := command.Handle()
	if err != nil {
		p.API.LogError(err.Error())
		return nil, model.NewAppError("msofficeplugin.ExecuteCommand", "Unable to execute command.", nil, err.Error(), gohttp.StatusInternalServerError)
	}

	apiconf.Poster.PostEphemeral(args.UserId, args.ChannelId, out)
	return &model.CommandResponse{}, nil
}

func (p *Plugin) ServeHTTP(pc *plugin.Context, w gohttp.ResponseWriter, req *gohttp.Request) {
	apiconf := p.getAPIConfig()
	mattermostUserID := req.Header.Get("Mattermost-User-ID")
	ctx := req.Context()
	ctx = api.Context(ctx, api.New(apiconf, mattermostUserID), p.notificationHandler)
	ctx = config.Context(ctx, apiconf.Config)

	fmt.Println("<><> URL: ", req.URL.String())
	p.httpHandler.ServeHTTP(w, req.WithContext(ctx))
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

func (p *Plugin) getAPIConfig() api.Config {
	conf := p.getConfig()
	store := store.NewPluginStore(p.API)

	return api.Config{
		Config: conf,
		Dependencies: &api.Dependencies{
			UserStore:         store,
			OAuth2StateStore:  store,
			SubscriptionStore: store,
			EventStore:        store,
			Logger:            p.API,
			Poster:            bot.NewPoster(p.API, conf),
			Remote:            remote.Makers[msgraph.Kind](conf, p.API),
		},
	}
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
