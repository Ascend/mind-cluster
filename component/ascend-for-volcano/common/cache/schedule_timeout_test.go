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
Package cache is using for HuaWei Ascend affinity schedule cross-session caches.
*/
package cache

import (
	"sync"
	"testing"
	"time"

	"volcano.sh/volcano/pkg/scheduler/api"
)

func TestScheduleTimeoutCacheRecordWaitStart(t *testing.T) {
	ResetTimeoutCacheForTest()
	defer ResetTimeoutCacheForTest()
	base := time.Date(2026, 9, 5, 10, 0, 0, 0, time.Local)
	SetTimeoutNowFuncForTest(func() time.Time { return base })

	jobID := api.JobID("timeout-cache-job")
	RecordWaitStart(jobID)
	// record again later keeps the start of the episode
	SetTimeoutNowFuncForTest(func() time.Time { return base.Add(time.Minute) })
	RecordWaitStart(jobID)
	if waitStart, ok := WaitStartTime(jobID); !ok || waitStart != base.Unix() {
		t.Errorf("wait start time = %d, %v, want %d, true", waitStart, ok, base.Unix())
	}

	// the clock restarts after the round ended
	EndRound(jobID)
	restart := base.Add(time.Hour)
	SetTimeoutNowFuncForTest(func() time.Time { return restart })
	RecordWaitStart(jobID)
	if waitStart, ok := WaitStartTime(jobID); !ok || waitStart != restart.Unix() {
		t.Errorf("wait start time = %d, %v, want %d, true", waitStart, ok, restart.Unix())
	}
}

func TestScheduleTimeoutCacheSweep(t *testing.T) {
	ResetTimeoutCacheForTest()

	leftJob := api.JobID("left-job")
	keptJob := api.JobID("kept-job")
	RecordWaitStart(leftJob)
	RecordWaitStart(keptJob)
	Sweep(func(jobID api.JobID) bool { return jobID == keptJob })

	if _, ok := WaitStartTime(leftJob); ok {
		t.Errorf("sweep should remove the job that no longer exists")
	}
	if _, ok := WaitStartTime(keptJob); !ok {
		t.Errorf("sweep should keep the job that still exists")
	}
}

// TestScheduleTimeoutCacheEmptyJobSafe verifies an empty job id is a no-op on
// every entry of the package api.
func TestScheduleTimeoutCacheEmptyJobSafe(t *testing.T) {
	ResetTimeoutCacheForTest()
	RecordWaitStart("")
	EndRound("")
	StampSession(map[api.JobID]bool{"": true})
	Sweep(func(api.JobID) bool { return true })
	if _, ok := WaitStartTime(""); ok {
		t.Errorf("an empty job should hold no wait clock")
	}
	if _, ok := SessionElapsed(""); ok {
		t.Errorf("an empty job should hold no session snapshot")
	}
}

func TestScheduleTimeoutCacheEndRound(t *testing.T) {
	ResetTimeoutCacheForTest()
	jobID := api.JobID("end-round-job")

	// a job which is still waiting holds its clock until the round ends
	RecordWaitStart(jobID)
	if _, ok := WaitStartTime(jobID); !ok {
		t.Errorf("a waiting job should hold a clock")
	}
	EndRound(jobID)
	if _, ok := WaitStartTime(jobID); ok {
		t.Errorf("end round should drop the wait clock")
	}

	// a job which never held a clock is untouched
	staleJob := api.JobID("end-round-stale-job")
	EndRound(staleJob)
	if _, ok := WaitStartTime(staleJob); ok {
		t.Errorf("end round of a job without a clock should record nothing")
	}
}

func TestScheduleTimeoutCacheStampSession(t *testing.T) {
	ResetTimeoutCacheForTest()
	defer ResetTimeoutCacheForTest()
	base := time.Date(2026, 9, 5, 10, 0, 0, 0, time.Local)
	SetTimeoutNowFuncForTest(func() time.Time { return base })

	waitingJob := api.JobID("snapshot-waiting-job")
	partialJob := api.JobID("snapshot-partial-job")
	freshJob := api.JobID("snapshot-fresh-job")
	RecordWaitStart(waitingJob)
	SetTimeoutNowFuncForTest(func() time.Time { return base.Add(time.Minute * 2) })
	RecordWaitStart(partialJob)

	StampSession(map[api.JobID]bool{waitingJob: true, partialJob: false, freshJob: true})

	if elapsed, ok := SessionElapsed(waitingJob); !ok || elapsed != 120 {
		t.Errorf("session elapsed of the waiting job = %d, %v, want 120, true", elapsed, ok)
	}
	if elapsed, ok := SessionElapsed(partialJob); !ok || elapsed != 0 {
		t.Errorf("session elapsed of the partial job = %d, %v, want 0, true", elapsed, ok)
	}
	if _, ok := SessionElapsed(freshJob); ok {
		t.Errorf("a job without a clock holds no elapsed entry")
	}

	// the snapshot is fixed for the whole session, the live clock moved on
	SetTimeoutNowFuncForTest(func() time.Time { return base.Add(time.Hour) })
	if elapsed, ok := SessionElapsed(waitingJob); !ok || elapsed != 120 {
		t.Errorf("session elapsed should stay fixed within the session, got %d, %v", elapsed, ok)
	}

	// a clock recorded after the stamp holds no entry until the next session
	RecordWaitStart(freshJob)
	if _, ok := SessionElapsed(freshJob); ok {
		t.Errorf("a clock recorded after the stamp holds no entry in this session")
	}

	// the next session rebuilds the snapshot from scratch
	StampSession(map[api.JobID]bool{waitingJob: false})
	if elapsed, ok := SessionElapsed(waitingJob); !ok || elapsed != 3600 {
		t.Errorf("session elapsed of the new session = %d, %v, want 3600, true", elapsed, ok)
	}
	if _, ok := SessionElapsed(partialJob); ok {
		t.Errorf("the previous session snapshot should be dropped")
	}
	if _, ok := SessionElapsed(api.JobID("unknown-job")); ok {
		t.Errorf("an unknown job holds no snapshot entry")
	}
}

// TestScheduleTimeoutCacheConcurrentAccess verifies the state survives the
// predicate fan-out: every predicate worker starts the clock of one job.
func TestScheduleTimeoutCacheConcurrentAccess(t *testing.T) {
	ResetTimeoutCacheForTest()
	jobID := api.JobID("concurrent-job")
	const workerNum = 16
	var waitGroup sync.WaitGroup
	for i := 0; i < workerNum; i++ {
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			RecordWaitStart(jobID)
		}()
		go func() {
			defer waitGroup.Done()
			WaitStartTime(jobID)
		}()
	}
	waitGroup.Wait()
	if waitStart, ok := WaitStartTime(jobID); !ok || waitStart <= 0 {
		t.Errorf("concurrent record should leave one wait clock, got %d, %v", waitStart, ok)
	}
}
