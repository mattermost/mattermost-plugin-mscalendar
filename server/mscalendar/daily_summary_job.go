// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"sync"
	"time"
)

const dailySummaryJobInterval = 15 * time.Minute

type DailySummaryJob struct {
	Env

	cancel     chan struct{}
	cancelled  chan struct{}
	cancelOnce sync.Once
}

func NewDailySummaryJob(env Env) *DailySummaryJob {
	return &DailySummaryJob{
		cancel:    make(chan struct{}),
		cancelled: make(chan struct{}),
		Env:       env,
	}
}

func (j *DailySummaryJob) Start() {
	go func() {
		defer close(j.cancelled)
		firstRun := j.timerUntilFirstRun()

		var ticker *time.Ticker
		select {
		case <-firstRun.C:
			ticker = time.NewTicker(dailySummaryJobInterval)
			defer func() {
				ticker.Stop()
			}()
			j.work()
		case <-j.cancel:
			return
		}

		for {
			select {
			case <-ticker.C:
				j.work()
			case <-j.cancel:
				return
			}
		}
	}()
}

func (job *DailySummaryJob) Cancel() {
	job.cancelOnce.Do(func() {
		close(job.cancel)
	})
	<-job.cancelled
}

func (j *DailySummaryJob) work() {
	j.Logger.Debugf("Daily summary job beginning")

	err := New(j.Env, "").ProcessAllDailySummary()
	if err != nil {
		j.Logger.Errorf("Error during daily summary job", "error", err.Error())
	}

	j.Logger.Debugf("Daily summary job finished")
}

// timeUntilFirstRun uses a job's interval to compute the time duration until the initial run.
func (j *DailySummaryJob) timerUntilFirstRun() *time.Timer {
	now := timeNowFunc()
	interval := dailySummaryJobInterval

	leftBound := now.Truncate(interval)
	target := leftBound.Add(interval)

	j.Logger.Debugf("Waiting until %s to run daily summary job", target.Format(time.Kitchen))

	diff := target.Sub(now)
	return time.NewTimer(diff)
}
