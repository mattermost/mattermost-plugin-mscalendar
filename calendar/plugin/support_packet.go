// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package plugin

import (
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

type SupportPacket struct {
	Version string `yaml:"version"`

	ConnectedUserCount uint64 `yaml:"connected_user_count"`
	SubscriptionCount  uint64 `yaml:"subscription_count"`
	IsOAuthConfigured  bool   `yaml:"is_oauth_configured"`
}

func (p *Plugin) GenerateSupportData(_ *plugin.Context) ([]*model.FileData, error) {
	var result *multierror.Error

	connectedUserCount, err := p.env.Dependencies.Store.GetConnectedUserCount()
	if err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to get the number of connected users for Support Packet"))
	}

	subscriptionCount, err := p.env.Dependencies.Store.GetSubscriptionCount()
	if err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to get the number of subscriptions for Support Packet"))
	}

	diagnostics := SupportPacket{
		Version:            p.env.PluginVersion,
		ConnectedUserCount: connectedUserCount,
		SubscriptionCount:  subscriptionCount,
		IsOAuthConfigured:  p.env.Config.IsOAuthConfigured(),
	}
	body, err := yaml.Marshal(diagnostics)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal diagnostics")
	}

	return []*model.FileData{{
		Filename: filepath.Join(p.env.PluginVersion, "diagnostics.yaml"),
		Body:     body,
	}}, result.ErrorOrNil()
}
