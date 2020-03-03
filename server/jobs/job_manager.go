package jobs

import (
	"sync"
	"time"

	"github.com/lieut-data/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
	"github.com/pkg/errors"
)

type JobManager struct {
	registeredJobs sync.Map
	activeJobs     sync.Map
	env            mscalendar.Env
	papi           cluster.JobPluginAPI
}

type RegisteredJob struct {
	id                string
	interval          time.Duration
	work              func(env mscalendar.Env)
	isEnabledByConfig func(env mscalendar.Env) bool
}

type activeJob struct {
	RegisteredJob

	cancel     chan struct{}
	cancelled  chan struct{}
	cancelOnce sync.Once
}

// NewJobManager creates a JobManager for to let plugin.go coordinate with the scheduled jobs.
func NewJobManager(papi cluster.JobPluginAPI, env mscalendar.Env) *JobManager {
	return &JobManager{
		papi: papi,
		env:  env,
	}
}

// AddJob accepts a RegisteredJob, stores it, and activates it if enabled.
func (jm *JobManager) AddJob(job RegisteredJob) error {
	jm.registeredJobs.Store(job.id, job)
	if job.isEnabledByConfig(jm.env) {
		err := jm.activateJob(job)
		if err != nil {
			return err
		}
	}

	return nil
}

// OnConfigurationChange activates/deactivates jobs based on their current state, and the current plugin config.
func (jm *JobManager) OnConfigurationChange(env mscalendar.Env) error {
	jm.env = env
	jm.registeredJobs.Range(func(k interface{}, v interface{}) bool {
		job, ok := v.(RegisteredJob)
		if !ok {
			jm.env.Logger.Errorf("Expected RegisteredJob, got %T", v)
			return true
		}
		enabled := job.isEnabledByConfig(env)
		active := jm.isJobActive(job.id)

		// Config is set to enable. Job does not exist, create new job.
		if enabled && !active {
			err := jm.activateJob(job)
			if err != nil {
				jm.env.Logger.Errorf("Error activating job", "id", job.id, "error", err.Error())
			}
		}

		// Config is set to disable. Job exists, kill existing job.
		if !enabled && active {
			err := jm.deactivateJob(job)
			if err != nil {
				jm.env.Logger.Errorf("Error deactivating job", "id", job.id, "error", err.Error())
			}
		}

		return true
	})
	return nil
}

// Close deactivates all active jobs. It is called in the plugin hook OnDeactivate.
func (jm *JobManager) Close() error {
	jm.activeJobs.Range(func(k interface{}, v interface{}) bool {
		job := v.(*activeJob)
		err := jm.deactivateJob(job.RegisteredJob)
		if err != nil {
			jm.env.Logger.Debugf("Failed to deactivate job", "error", err.Error())
		}

		return true
	})
	return nil
}

// activateJob creates an ActiveJob, starts it, and stores it in the job manager.
func (jm *JobManager) activateJob(job RegisteredJob) error {
	_, ok := jm.activeJobs.Load(job.id)
	if ok {
		return errors.Errorf("Attempted to re-activate an active job %s", job.id)
	}

	actJob := &activeJob{
		RegisteredJob: job,
		cancel:        make(chan struct{}),
		cancelled:     make(chan struct{}),
	}
	actJob.start(jm.getEnv, jm.papi)

	jm.activeJobs.Store(job.id, actJob)
	return nil
}

// deactivateJob closes the job, releasing the cluster mutex, then remoes the job from the job manager.
func (jm *JobManager) deactivateJob(job RegisteredJob) error {
	v, ok := jm.activeJobs.Load(job.id)
	if !ok {
		return errors.Errorf("Attempted to deactivate a non-active job %s", job.id)
	}

	actJob := v.(*activeJob)
	actJob.close()
	jm.activeJobs.Delete(job.id)

	return nil
}

// isJobActive checks if a job is currently active, which includes enabled jobs that are waiting to run for their first time.
func (jm *JobManager) isJobActive(id string) bool {
	_, ok := jm.activeJobs.Load(id)
	return ok
}

// getEnv returns the mscalendar.Env stored on the job manager
func (jm *JobManager) getEnv() mscalendar.Env {
	return jm.env
}
