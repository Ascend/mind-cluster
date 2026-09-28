/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package cache is using for Huawei Ascend affinity schedule cross-session caches.
*/
package cache

import (
	"sync"
	"time"

	"volcano.sh/volcano/pkg/scheduler/api"
)

// timeoutCache is the process-wide schedule timeout state, reached only
// through the package functions. It is pure timing: the constraint downgrade
// owns common/downgrade and derives its time inputs from this store.
var timeoutCache = newScheduleTimeoutCache()

// scheduleTimeoutCache is the state behind the package functions, guarded by
// a mutex: volcano runs the node predicates of one task in parallel, so the
// wait clock of one job may be started by several goroutines at once.
type scheduleTimeoutCache struct {
	mu        sync.Mutex
	waitStart map[api.JobID]int64 // job uid -> start of the current waiting episode, unix second
	// session is the elapsed snapshot of the session start, an entry exists
	// only for the jobs which held a clock at the stamp
	session map[api.JobID]int64
	nowFunc func() time.Time
}

// newScheduleTimeoutCache builds an empty state running on the wall clock.
func newScheduleTimeoutCache() *scheduleTimeoutCache {
	return &scheduleTimeoutCache{
		waitStart: map[api.JobID]int64{},
		nowFunc:   time.Now,
	}
}

// RecordWaitStart starts the wait clock of a job when the scheduler first
// sees it waiting (at the session open), the first record wins and survives
// a job-level reschedule.
func RecordWaitStart(jobID api.JobID) {
	if jobID == "" {
		return
	}
	timeoutCache.mu.Lock()
	defer timeoutCache.mu.Unlock()
	if _, ok := timeoutCache.waitStart[jobID]; ok {
		return
	}
	timeoutCache.waitStart[jobID] = timeoutCache.nowFunc().Unix()
}

// WaitStartTime returns the recorded start of the current waiting episode of
// a job, false means the record does not exist.
func WaitStartTime(jobID api.JobID) (int64, bool) {
	if jobID == "" {
		return 0, false
	}
	timeoutCache.mu.Lock()
	defer timeoutCache.mu.Unlock()
	waitStart, ok := timeoutCache.waitStart[jobID]
	return waitStart, ok
}

// EndRound drops the wait clock of a job whose whole scope is scheduled, the
// downgrade level of common/downgrade is kept for the later rescheduling.
func EndRound(jobID api.JobID) {
	if jobID == "" {
		return
	}
	timeoutCache.mu.Lock()
	defer timeoutCache.mu.Unlock()
	delete(timeoutCache.waitStart, jobID)
}

// StampSession rebuilds the elapsed snapshot once per session, so every task
// of a job decides alike. Only the key set of the given map is read.
func StampSession(allWaitingOf map[api.JobID]bool) {
	timeoutCache.mu.Lock()
	defer timeoutCache.mu.Unlock()
	timeoutCache.session = make(map[api.JobID]int64, len(allWaitingOf))
	now := timeoutCache.nowFunc().Unix()
	for jobID := range allWaitingOf {
		if waitStart, ok := timeoutCache.waitStart[jobID]; ok && now >= waitStart {
			timeoutCache.session[jobID] = now - waitStart
		}
	}
}

// SessionElapsed returns the session-start wait seconds of the job, the
// snapshot value is fixed for the whole session so all the tasks decide alike.
func SessionElapsed(jobID api.JobID) (int64, bool) {
	if jobID == "" {
		return 0, false
	}
	timeoutCache.mu.Lock()
	defer timeoutCache.mu.Unlock()
	elapsed, ok := timeoutCache.session[jobID]
	return elapsed, ok
}

// Sweep removes the wait clocks of the jobs which are no longer alive. The key
// snapshot and the removal hold the cache lock, the alive predicate runs between
// the two windows outside it: it walks the job table of the schedule env and must
// never hold this lock.
func Sweep(alive func(api.JobID) bool) {
	timeoutCache.mu.Lock()
	jobIDs := make([]api.JobID, 0, len(timeoutCache.waitStart))
	for jobID := range timeoutCache.waitStart {
		jobIDs = append(jobIDs, jobID)
	}
	timeoutCache.mu.Unlock()

	var dead []api.JobID
	for _, jobID := range jobIDs {
		if !alive(jobID) {
			dead = append(dead, jobID)
		}
	}
	if len(dead) == 0 {
		return
	}
	timeoutCache.mu.Lock()
	defer timeoutCache.mu.Unlock()
	for _, jobID := range dead {
		delete(timeoutCache.waitStart, jobID)
	}
}

// ResetTimeoutCacheForTest drops every record of the timeout state and
// restores the wall clock, unit tests only.
func ResetTimeoutCacheForTest() {
	timeoutCache = newScheduleTimeoutCache()
}

// SetTimeoutNowFuncForTest overrides the clock the timeout state reads, unit
// tests only.
func SetTimeoutNowFuncForTest(nowFunc func() time.Time) {
	if nowFunc == nil {
		return
	}
	timeoutCache.mu.Lock()
	defer timeoutCache.mu.Unlock()
	timeoutCache.nowFunc = nowFunc
}
