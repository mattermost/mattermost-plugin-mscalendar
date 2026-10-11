// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package engine

import (
	"fmt"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/config"
)

func TestNewConnectAttachment(t *testing.T) {
	env := Env{
		Config: &config.Config{
			Provider: config.ProviderConfig{
				DisplayName:    "Microsoft Calendar",
				CommandTrigger: "mscalendar",
			},
		},
		Dependencies: &Dependencies{},
	}
	pluginURL := "http://localhost:8065/plugins/com.mattermost.mscalendar"

	b := &mscBot{
		Env:       env,
		pluginURL: pluginURL,
	}

	attachment := b.newConnectAttachment()

	require.NotNil(t, attachment)
	require.Equal(t, "Connect your Microsoft Calendar account", attachment.Title)
	require.Equal(t, ConnectMessageBody, attachment.Text)
	require.Len(t, attachment.Actions, 1)

	action := attachment.Actions[0]
	require.Equal(t, ConnectActionID, action.Id)
	require.Equal(t, "Connect account", action.Name)
	require.Equal(t, model.PostActionTypeButton, action.Type)
	require.NotNil(t, action.Integration)
	require.Equal(t, pluginURL+"/oauth2/connect", action.Integration.URL)

	expectedFallback := fmt.Sprintf("Connect your Microsoft Calendar account\n%s\n[Connect account](%s/oauth2/connect)", ConnectMessageBody, pluginURL)
	require.Equal(t, expectedFallback, attachment.Fallback)
}
