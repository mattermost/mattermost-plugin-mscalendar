// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"sync"
	"time"
)

const refreshOAuth2TokenJobInterval = 24 * time.Hour

type RefreshOAuth2TokenJob struct {
	Env

	cancel     chan struct{}
	cancelled  chan struct{}
	cancelOnce sync.Once
}

func NewRefreshOAuth2TokenJob(env Env) *RefreshOAuth2TokenJob {
	return &RefreshOAuth2TokenJob{
		cancel:    make(chan struct{}),
		cancelled: make(chan struct{}),
		Env:       env,
	}
}

func (job *RefreshOAuth2TokenJob) Start() {
	go func() {
		defer close(job.cancelled)

		ticker := time.NewTicker(refreshOAuth2TokenJobInterval)
		defer func() {
			ticker.Stop()
		}()

		for {
			select {
			case <-ticker.C:
				job.work()
			case <-job.cancel:
				return
			}
		}
	}()
}

func (job *RefreshOAuth2TokenJob) Cancel() {
	job.cancelOnce.Do(func() {
		close(job.cancel)
	})
	<-job.cancelled
}

func (j *RefreshOAuth2TokenJob) work() {
	j.Logger.Debugf("Refresh OAuth2 token job beginning")

	err := New(j.Env, "").RefreshAllOAuth2Tokens()
	if err != nil {
		j.Logger.Errorf("Error during refresh OAuth2 token job", "error", err.Error())
	}

	j.Logger.Debugf("Refresh OAuth2 token job finished")
}
