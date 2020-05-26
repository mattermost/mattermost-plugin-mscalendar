// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package jobs

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
)

type JobManager struct {
	registeredJobs sync.Map
	activeJobs     sync.Map
	env            mscalendar.Env
	papi           cluster.JobPluginAPI
	mux            sync.Mutex
}

type RegisteredJob struct {
	id       string
	interval time.Duration
	work     func(env mscalendar.Env)
}

var scheduleFunc = func(api cluster.JobPluginAPI, id string, wait cluster.NextWaitInterval, cb func()) (io.Closer, error) {
	return cluster.Schedule(api, id, wait, cb)
}

type activeJob struct {
	RegisteredJob
	ScheduledJob io.Closer
	Context      context.Context
}

func newActiveJob(rj RegisteredJob, sched io.Closer, ctx context.Context) *activeJob {
	return &activeJob{
		RegisteredJob: rj,
		ScheduledJob:  sched,
		Context:       ctx,
	}
}

// NewJobManager creates a JobManager for to let plugin.go coordinate with the scheduled jobs.
func NewJobManager(papi cluster.JobPluginAPI, env mscalendar.Env) *JobManager {
	return &JobManager{
		papi: papi,
		env:  env,
	}
}

// AddJob accepts a RegisteredJob, stores it, and activates it if enabled.
func (jm *JobManager) AddJob(job RegisteredJob) {
	jm.registeredJobs.Store(job.id, job)
	jm.activateJob(job)
}

// Close deactivates all active jobs. It is called in the plugin hook OnDeactivate.
func (jm *JobManager) Close() error {
	jm.activeJobs.Range(func(k interface{}, v interface{}) bool {
		job := v.(*activeJob)
		err := jm.deactivateJob(job.RegisteredJob)
		if err != nil {
			jm.env.Logger.Debugf("Failed to deactivate job: %v", err)
		}

		return true
	})
	return nil
}

// activateJob creates an ActiveJob, starts it, and stores it in the job manager.
func (jm *JobManager) activateJob(job RegisteredJob) error {
	scheduled, err := scheduleFunc(jm.papi, job.id, cluster.MakeWaitForRoundedInterval(job.interval), func() { job.work(jm.getEnv()) })
	if err != nil {
		return err
	}

	actJob := newActiveJob(job, scheduled, context.Background())

	jm.activeJobs.Store(job.id, actJob)
	return nil
}

// deactivateJob closes the job, releasing the cluster mutex, then remoes the job from the job manager.
func (jm *JobManager) deactivateJob(job RegisteredJob) error {
	v, ok := jm.activeJobs.Load(job.id)
	if !ok {
		return fmt.Errorf("Attempted to deactivate a non-active job %s", job.id)
	}

	scheduledJob := v.(*activeJob)
	err := scheduledJob.ScheduledJob.Close()
	if err != nil {
		return err
	}
	jm.activeJobs.Delete(job.id)

	return nil
}

// getEnv returns the mscalendar.Env stored on the job manager
func (jm *JobManager) getEnv() mscalendar.Env {
	return jm.env
}
