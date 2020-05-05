// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package jobs

import (
	"io"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/jobs/mock_cluster"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/mock_plugin_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/mock_remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store/mock_store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot/mock_bot"
)

func newTestEnv(ctrl *gomock.Controller) mscalendar.Env {
	s := mock_store.NewMockStore(ctrl)
	poster := mock_bot.NewMockPoster(ctrl)
	mockRemote := mock_remote.NewMockRemote(ctrl)
	mockPluginAPI := mock_plugin_api.NewMockPluginAPI(ctrl)

	logger := &bot.NilLogger{}
	return mscalendar.Env{
		Dependencies: &mscalendar.Dependencies{
			Store:     s,
			Logger:    logger,
			Poster:    poster,
			Remote:    mockRemote,
			PluginAPI: mockPluginAPI,
		},
	}
}

func TestJobManagerOnConfigurationChange(t *testing.T) {
	for name, tc := range map[string]struct {
		enabled        bool
		active         bool
		numCloseCalls  int
		expectedActive bool
	}{
		"Not active, config set to disabled": {
			enabled:        false,
			active:         false,
			numCloseCalls:  0,
			expectedActive: false,
		},
		"Not active, config set to enabled": {
			enabled:        true,
			active:         false,
			numCloseCalls:  0,
			expectedActive: true,
		},
		"Active, config set to disabled": {
			enabled:        false,
			active:         true,
			numCloseCalls:  1,
			expectedActive: false,
		},
		"Active, config set to enabled": {
			enabled:        true,
			active:         true,
			numCloseCalls:  0,
			expectedActive: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockJobsPluginAPI := mock_cluster.NewMockJobPluginAPI(ctrl)

			mc := &mockCloser{numCalls: 0}
			scheduleFunc = func(api cluster.JobPluginAPI, id string, wait cluster.NextWaitInterval, cb func()) (io.Closer, error) {
				cb()
				return mc, nil
			}

			env := newTestEnv(ctrl)

			j := RegisteredJob{
				id:                name,
				interval:          5 * time.Minute,
				work:              func(env mscalendar.Env) {},
				isEnabledByConfig: func(env mscalendar.Env) bool { return tc.enabled },
			}
			jm := NewJobManager(mockJobsPluginAPI, env)
			jm.AddJob(j)
			defer jm.Close()

			if tc.active {
				err := jm.activateJob(j)
				require.Nil(t, err)
			}

			err := jm.OnConfigurationChange(env)
			require.Nil(t, err)
			time.Sleep(5 * time.Millisecond)

			require.Equal(t, tc.expectedActive, jm.isJobActive(j.id))
			require.Equal(t, tc.numCloseCalls, mc.numCalls)
		})
	}
}

type mockCloser struct {
	numCalls int
}

func (mc *mockCloser) Close() error {
	mc.numCalls++
	return nil
}
