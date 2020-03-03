package jobs

import (
	"fmt"
	"time"

	"github.com/lieut-data/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
)

// start waits until the job's initial run time, then runs the job every interval.
// A cluster mutex is used to synchronize with other nodes that are running the same plugin.
func (job *activeJob) start(getEnv func() mscalendar.Env, papi cluster.MutexPluginAPI) {
	go func() {
		defer close(job.cancelled)

		// Get mutex to gain access to sensitive section
		lock := cluster.NewMutex(papi, job.id)
		lock.Lock()
		defer lock.Unlock()

		// Get initial run's time
		firstRun := job.timeUntilFirstRun()
		timer := time.NewTimer(firstRun)
		getEnv().Logger.Debugf("%s job will run at at `%s`, in `%s`, then every `%s`.", job.id, time.Now().Add(firstRun).Format("Jan 02, 3:04PM MST"), timeDurationDisplay(firstRun), timeDurationDisplay(job.interval))

		select {
		case <-timer.C:
			// time for the initial run
		case <-job.cancel:
			return
		}

		// Prepare ticker for recurring runs
		ticker := time.NewTicker(job.interval)
		defer func() {
			ticker.Stop()
		}()

		// Do initial run
		go job.work(getEnv())

		// Loop through ticker
		for {
			select {
			case <-ticker.C:
				go job.work(getEnv())
			case <-job.cancel:
				return
			}
		}
	}()
}

// close ends the job's ticker loop, which releases the cluster mutex.
func (job *activeJob) close() {
	job.cancelOnce.Do(func() {
		close(job.cancel)
	})
	<-job.cancelled
}

// timeUntilFirstRun uses a job's interval to compute the time duration until the initial run.
func (job RegisteredJob) timeUntilFirstRun() time.Duration {
	now := time.Now()

	leftBound := now.Truncate(job.interval)
	target := leftBound.Add(job.interval)

	return target.Sub(now)
}

func timeDurationDisplay(d time.Duration) string {
	return fmt.Sprintf("%v hours and %v minutes and %v seconds", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}
