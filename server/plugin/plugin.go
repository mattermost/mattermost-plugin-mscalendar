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

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/api"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/command"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/msgraph"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/oauth2connect"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/pluginapi"
)

type Plugin struct {
	plugin.MattermostPlugin

	envLock *sync.RWMutex
	bot     bot.Bot
	mscalendar.Env
	statusSyncJob         *mscalendar.StatusSyncJob
	notificationProcessor mscalendar.NotificationProcessor
	httpHandler           *httputils.Handler

	Templates map[string]*template.Template
}

func NewWithEnv(env mscalendar.Env) *Plugin {
	return &Plugin{
		envLock: &sync.RWMutex{},
		Env:     env,
	}
}

func (p *Plugin) OnActivate() error {
	p.Env.Dependencies.PluginAPI = pluginapi.New(p.API)
	p.bot = bot.New(p.API, p.Helpers)
	err := p.bot.Ensure(
		&model.Bot{
			Username:    config.BotUserName,
			DisplayName: config.BotDisplayName,
			Description: config.BotDescription,
		},
		"assets/profile.png")
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}

	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "couldn't get bundle path")
	}
	err = p.loadTemplates(bundlePath)
	if err != nil {
		return err
	}

	command.Register(p.API.RegisterCommand)
	return nil
}

func (p *Plugin) OnConfigurationChange() error {
	env := p.getEnv()
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
	pluginURLPath := "/plugins/" + env.Config.PluginID
	pluginURL := strings.TrimRight(*mattermostSiteURL, "/") + pluginURLPath

	p.updateEnv(func(env *mscalendar.Env) {
		env.StoredConfig = stored
		env.Config.MattermostSiteURL = *mattermostSiteURL
		env.Config.MattermostSiteHostname = mattermostURL.Hostname()
		env.Config.PluginURL = pluginURL
		env.Config.PluginURLPath = pluginURLPath
		env.Dependencies.Remote = remote.Makers[msgraph.Kind](env.Config, env.Logger)

		p.bot = p.bot.WithConfig(stored.BotConfig)
		p.Env.Config.BotUserID = p.bot.MattermostUserID()
		p.Env.Dependencies.Logger = p.bot
		p.Env.Dependencies.Poster = p.bot
		p.Env.Dependencies.Store = store.NewPluginStore(p.API, p.bot)
		if p.notificationProcessor == nil {
			p.notificationProcessor = mscalendar.NewNotificationProcessor(*env)
		} else {
			p.notificationProcessor.Configure(*env)
		}

		p.httpHandler = httputils.NewHandler()
		oauth2connect.Init(p.httpHandler, mscalendar.NewOAuth2App(*env))
		api.Init(p.httpHandler, *env, p.notificationProcessor)
	})

	p.POC_initUserStatusSyncJob()

	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	env := p.getEnv()
	mscalendar := mscalendar.New(env, args.UserId)

	command := command.Command{
		Context:    c,
		Args:       args,
		ChannelID:  args.ChannelId,
		Config:     env.Config,
		MSCalendar: mscalendar,
	}
	out, err := command.Handle()
	if err != nil {
		p.API.LogError(err.Error())
		return nil, model.NewAppError("mscalendarplugin.ExecuteCommand", "Unable to execute command.", nil, err.Error(), gohttp.StatusInternalServerError)
	}

	env.Poster.Ephemeral(args.UserId, args.ChannelId, out)
	return &model.CommandResponse{}, nil
}

func (p *Plugin) ServeHTTP(pc *plugin.Context, w gohttp.ResponseWriter, req *gohttp.Request) {
	p.httpHandler.ServeHTTP(w, req)
}

func (p *Plugin) getEnv() mscalendar.Env {
	p.envLock.RLock()
	defer p.envLock.RUnlock()
	return p.Env
}

func (p *Plugin) updateEnv(f func(*mscalendar.Env)) {
	p.envLock.Lock()
	defer p.envLock.Unlock()

	f(&p.Env)
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

// POC_initUserStatusSyncJob begins a job that runs every 5 minutes to update the MM user's status based on their status in their Microsoft calendar
// This needs to be improved to run on a single node in the HA environment. Hence why the name is currently prefixed with POC
func (p *Plugin) POC_initUserStatusSyncJob() {
	env := p.getEnv()
	enable := env.Config.EnableStatusSync
	logger := env.Logger

	// Config is set to enable. No job exists, start a new job.
	if enable && p.statusSyncJob == nil {
		logger.Debugf("Enabling user status sync job")

		job := mscalendar.NewStatusSyncJob(mscalendar.New(env, env.Config.BotUserID))
		p.statusSyncJob = job
		go job.Start()
	}

	// Config is set to disable. Job exists, kill existing job.
	if !enable && p.statusSyncJob != nil {
		logger.Debugf("Disabling user status sync job")

		p.statusSyncJob.Cancel()
		p.statusSyncJob = nil
	}
}
