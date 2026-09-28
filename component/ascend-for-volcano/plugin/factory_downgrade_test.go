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
Package plugin is using for HuaWei Ascend pin affinity schedule.
*/
package plugin

import (
	"strconv"
	"testing"
	"time"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/cache"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/downgrade"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

type schedulerDowngradeTimeoutTest struct {
	name string
	conf map[string]string
	want int
}

func buildSchedulerDowngradeTimeoutTestCases() []schedulerDowngradeTimeoutTest {
	return []schedulerDowngradeTimeoutTest{
		{name: "01 nil conf falls back to default", conf: nil, want: DefaultSchedulerDowngradeTimeout},
		{name: "02 absent key falls back to default", conf: map[string]string{}, want: DefaultSchedulerDowngradeTimeout},
		{name: "03 invalid value falls back to default",
			conf: map[string]string{schedulerDowngradeTimeoutKey: "abc"}, want: DefaultSchedulerDowngradeTimeout},
		{name: "04 non-positive value falls back to default",
			conf: map[string]string{schedulerDowngradeTimeoutKey: "-5"}, want: DefaultSchedulerDowngradeTimeout},
		{name: "05 zero value falls back to default",
			conf: map[string]string{schedulerDowngradeTimeoutKey: "0"}, want: DefaultSchedulerDowngradeTimeout},
		{name: "06 value over the upper bound falls back to default",
			conf: map[string]string{schedulerDowngradeTimeoutKey: "86401"}, want: DefaultSchedulerDowngradeTimeout},
		{name: "07 lower bound value", conf: map[string]string{schedulerDowngradeTimeoutKey: "1"}, want: 1},
		{name: "08 upper bound value",
			conf: map[string]string{schedulerDowngradeTimeoutKey: "86400"}, want: 86400},
		{name: "09 valid value", conf: map[string]string{schedulerDowngradeTimeoutKey: "30"}, want: 30},
	}
}

func TestGetSchedulerDowngradeTimeout(t *testing.T) {
	for _, tt := range buildSchedulerDowngradeTimeoutTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSchedulerDowngradeTimeout(tt.conf); got != tt.want {
				t.Errorf("getSchedulerDowngradeTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

// newDowngradeJob builds a downgrade-enabled job.
func newDowngradeJob() SchedulerJob {
	return SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{
		ComJob: util.ComJob{Annotation: map[string]string{util.SchedulerDowngradeAnnoKey: "true"}},
		NPUJob: &util.NPUJob{},
	}}
}

// buildRoundJobInfo builds an api job of scopedNum npu tasks: pendingNum of
// them waiting, pipelinedNum holding a reserved node without a committed
// binding and the rest placed with the marker annotation.
func buildRoundJobInfo(jobID api.JobID, pendingNum, pipelinedNum, scopedNum int,
	placedAnno map[string]string) *api.JobInfo {
	tasks := make(map[api.TaskID]*api.TaskInfo, scopedNum)
	for i := 0; i < scopedNum; i++ {
		uid := "task" + strconv.Itoa(i)
		status, anno := api.Bound, placedAnno
		switch {
		case i < pendingNum:
			status, anno = api.Pending, nil
		case i < pendingNum+pipelinedNum:
			status, anno = api.Pipelined, nil
		}
		task := newPlacedNPUTask(uid, status, anno)
		if status == api.Pipelined {
			task.NodeName = "node-" + uid
		}
		tasks[api.TaskID(uid)] = task
	}
	return &api.JobInfo{UID: jobID, Tasks: tasks}
}

func TestReconcileTimingState(t *testing.T) {
	cache.ResetTimeoutCacheForTest()
	downgrade.ResetDowngradeCacheForTest()
	sHandle := &ScheduleHandler{}
	sHandle.Jobs = map[api.JobID]SchedulerJob{"kept-job": {}}

	sHandle.reconcileTimingState(nil)
	cache.RecordWaitStart("kept-job")
	cache.RecordWaitStart("gone-job")
	downgrade.RecordDowngrade("gone-job", configSpBlock2, 100)

	// the next session reconciles the same global state and sweeps the stale record
	sHandle.Jobs = map[api.JobID]SchedulerJob{"kept-job": {}}
	sHandle.reconcileTimingState(nil)
	if _, ok := cache.WaitStartTime("gone-job"); ok {
		t.Errorf("record of the job which has left should be swept")
	}
	if _, ok := cache.WaitStartTime("kept-job"); !ok {
		t.Errorf("record of the job still in schedule env should be kept")
	}
	if _, ok := downgrade.DowngradeConfig("gone-job"); ok {
		t.Errorf("downgrade record of the job which has left should be swept")
	}
}

type schedulerDowngradeSeedTest struct {
	name         string
	placed       int    // the placed npu tasks of the job carrying the markers
	preConfig    string // the cache already holds this snapshot
	preClock     bool   // the process already holds a carried wait clock
	legacyMarker bool   // the pods carry the value of the previous version
	disabled     bool
	wantConfig   string
	wantClockOn  bool // the session open anchors the waiting rest after the seed, the clock is on for every job
}

func buildSchedulerDowngradeSeedTestCases() []schedulerDowngradeSeedTest {
	return []schedulerDowngradeSeedTest{
		{name: "01 the partially placed job is seeded", placed: 3,
			wantConfig: configSpBlock2, wantClockOn: true},
		{name: "02 the fully placed job is seeded too", placed: 4, wantConfig: configSpBlock2},
		{name: "03 an existing state is never overwritten", placed: 3, preConfig: configSpBlock1,
			wantConfig: configSpBlock1, wantClockOn: true},
		{name: "04 the downgrade-disabled job is timed but never seeded", placed: 3, disabled: true,
			wantClockOn: true},
		{name: "05 a marker of the previous version is refused", placed: 3, legacyMarker: true,
			wantClockOn: true},
		{name: "06 a job holding a carried clock is never seeded", placed: 3, preClock: true,
			wantClockOn: true},
	}
}

// TestReconcileTimingStateSeed covers the lazy seed: the state is rebuilt
// from the pod markers only when the process holds nothing about the job, the
// seed itself opens no wait clock, the session open anchors the waiting rest
// right after it. The config assertions of the seeded cases also prove the
// reconciliation order: a wait clock started before the seed would mask the
// restart and block the seed.
func TestReconcileTimingStateSeed(t *testing.T) {
	for _, tt := range buildSchedulerDowngradeSeedTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			cache.ResetTimeoutCacheForTest()
			downgrade.ResetDowngradeCacheForTest()
			sHandle := &ScheduleHandler{}
			jobID := api.JobID("seed-job")
			markerConfig := configSpBlock2
			if tt.legacyMarker {
				markerConfig = legacyConfig
			}
			marker := downgradeMarker("1000", markerConfig)
			ssnJobs := map[api.JobID]*api.JobInfo{
				jobID: buildRoundJobInfo(jobID, 4-tt.placed, 0, 4, marker),
			}
			vcJob := newDowngradeJob()
			if tt.disabled {
				vcJob.SchedulerJobAttr.ComJob.Annotation = nil
			}
			if tt.preConfig != "" {
				downgrade.RecordDowngrade(jobID, tt.preConfig, 500)
			}
			if tt.preClock {
				cache.RecordWaitStart(jobID)
			}
			sHandle.Jobs = map[api.JobID]SchedulerJob{jobID: vcJob}

			sHandle.reconcileTimingState(ssnJobs)

			config, ok := downgrade.DowngradeConfig(jobID)
			if tt.wantConfig == "" {
				if ok {
					t.Errorf("no downgrade state expected, got config %q", config)
				}
			} else if !ok || config != tt.wantConfig {
				t.Errorf("downgrade config = %q, %v, want %q, true", config, ok, tt.wantConfig)
			}
			_, clockOn := cache.WaitStartTime(jobID)
			if clockOn != tt.wantClockOn {
				t.Errorf("wait clock = %v, want %v", clockOn, tt.wantClockOn)
			}
		})
	}
}

type timingStateJobRescheduleTest struct {
	name          string
	pending       int
	scoped        int
	effective     bool // the level of the job is already effective
	disabled      bool
	wantStateKept bool
}

func buildTimingStateJobRescheduleTestCases() []timingStateJobRescheduleTest {
	return []timingStateJobRescheduleTest{
		{name: "01 a job whose effective level is scheduled away resets the level",
			pending: 4, scoped: 4, effective: true},
		{name: "02 a partially scheduled job keeps its level",
			pending: 2, scoped: 4, effective: true, wantStateKept: true},
		{name: "03 a job which is still deepening keeps its level",
			pending: 4, scoped: 4, wantStateKept: true},
		{name: "04 the downgrade-disabled job is untouched",
			pending: 4, scoped: 4, effective: true, disabled: true, wantStateKept: true},
	}
}

func TestReconcileTimingStateJobReschedule(t *testing.T) {
	for _, tt := range buildTimingStateJobRescheduleTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			cache.ResetTimeoutCacheForTest()
			downgrade.ResetDowngradeCacheForTest()
			sHandle := &ScheduleHandler{}
			jobID := api.JobID("reschedule-job")
			// the wait clock of the fault episode, the level reset keeps it so
			// the job degrades on top of the already waited time
			cache.RecordWaitStart(jobID)
			downgrade.RecordDowngrade(jobID, configSpBlock2, 1000)
			if tt.effective {
				downgrade.MarkEffective(jobID)
			}
			ssnJobs := map[api.JobID]*api.JobInfo{
				jobID: buildRoundJobInfo(jobID, tt.pending, 0, tt.scoped, nil),
			}
			vcJob := newDowngradeJob()
			if tt.disabled {
				vcJob.SchedulerJobAttr.ComJob.Annotation = nil
			}
			sHandle.Jobs = map[api.JobID]SchedulerJob{jobID: vcJob}

			sHandle.reconcileTimingState(ssnJobs)

			config, configOK := downgrade.DowngradeConfig(jobID)
			if configOK != tt.wantStateKept {
				t.Errorf("downgrade config kept = %v, want %v (config %q)",
					configOK, tt.wantStateKept, config)
			}
			if tt.wantStateKept && config != configSpBlock2 {
				t.Errorf("the kept config = %q, want %q", config, configSpBlock2)
			}
			// the wait clock survives both the reset and the kept state
			if _, clockOK := cache.WaitStartTime(jobID); !clockOK {
				t.Errorf("the wait clock should survive the session reconciliation")
			}
		})
	}
}

type timingStateEndRoundTest struct {
	name        string
	pending     int
	pipelined   int
	scoped      int
	wantClockOn bool
}

func buildTimingStateEndRoundTestCases() []timingStateEndRoundTest {
	return []timingStateEndRoundTest{
		{name: "01 the fully placed round ends and the clock is cleared", scoped: 4},
		{name: "02 the fully pipelined round keeps the clock, no binding is committed",
			pipelined: 4, scoped: 4, wantClockOn: true},
		{name: "03 the partially placed round keeps the clock for the waiting rest",
			pending: 2, scoped: 4, wantClockOn: true},
	}
}

// TestReconcileTimingStateEndRound covers the round completion: the
// wait clock is dropped only when every in-scope task reached a committed
// state, a pipelined round keeps accumulating its wait time. The round rolled
// back to pending is covered by TestReconcileTimingStateRollbackElapsed.
func TestReconcileTimingStateEndRound(t *testing.T) {
	for _, tt := range buildTimingStateEndRoundTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			cache.ResetTimeoutCacheForTest()
			downgrade.ResetDowngradeCacheForTest()
			sHandle := &ScheduleHandler{}
			jobID := api.JobID("end-round-job")
			cache.RecordWaitStart(jobID)
			ssnJobs := map[api.JobID]*api.JobInfo{
				jobID: buildRoundJobInfo(jobID, tt.pending, tt.pipelined, tt.scoped, nil),
			}
			sHandle.Jobs = map[api.JobID]SchedulerJob{jobID: newDowngradeJob()}

			sHandle.reconcileTimingState(ssnJobs)

			_, clockOn := cache.WaitStartTime(jobID)
			if clockOn != tt.wantClockOn {
				t.Errorf("wait clock = %v, want %v", clockOn, tt.wantClockOn)
			}
		})
	}
}

// TestReconcileTimingStateRollbackElapsed covers the backfill
// amplification: a round which is fully pipelined and rolled back every
// session keeps its wait clock, so the elapsed time grows towards the
// downgrade window instead of restarting at every rollback.
func TestReconcileTimingStateRollbackElapsed(t *testing.T) {
	cache.ResetTimeoutCacheForTest()
	downgrade.ResetDowngradeCacheForTest()
	now := int64(1000)
	cache.SetTimeoutNowFuncForTest(func() time.Time { return time.Unix(now, 0) })
	sHandle := &ScheduleHandler{}
	jobID := api.JobID("rollback-job")
	cache.RecordWaitStart(jobID)
	rolledBack := map[api.JobID]*api.JobInfo{
		jobID: buildRoundJobInfo(jobID, 4, 0, 4, nil),
	}
	sHandle.Jobs = map[api.JobID]SchedulerJob{jobID: newDowngradeJob()}

	// every session the round is fully pipelined and rolled back afterwards,
	// the next session sees the whole scope waiting again
	for _, waited := range []int64{120, 240, 360} {
		now += 120
		sHandle.reconcileTimingState(rolledBack)
		elapsed, ok := cache.SessionElapsed(jobID)
		if !ok || elapsed != waited {
			t.Errorf("session elapsed = %d, %v, want %d, true", elapsed, ok, waited)
		}
	}
	if _, clockOn := cache.WaitStartTime(jobID); !clockOn {
		t.Errorf("the wait clock should survive the repeated rollbacks")
	}
}

type timingStateAnchorTest struct {
	name        string
	pending     int
	scoped      int
	disabled    bool
	wantClockOn bool
}

func buildTimingStateAnchorTestCases() []timingStateAnchorTest {
	return []timingStateAnchorTest{
		{name: "01 the waiting job is anchored at its first session", pending: 4, scoped: 4,
			wantClockOn: true},
		{name: "02 the fully placed job is not anchored", scoped: 4},
		{name: "03 the downgrade-disabled job is anchored too", pending: 4, scoped: 4, disabled: true,
			wantClockOn: true},
	}
}

// TestReconcileTimingStateAnchorsWaitClock covers the session-open anchor: the
// wait clock starts at the first session which sees the job waiting, the
// queue starvation before the first predicate is waited time too. A freshly
// anchored clock freezes with elapsed zero, the waited windows are counted
// from the next session on.
func TestReconcileTimingStateAnchorsWaitClock(t *testing.T) {
	for _, tt := range buildTimingStateAnchorTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			cache.ResetTimeoutCacheForTest()
			downgrade.ResetDowngradeCacheForTest()
			sHandle := &ScheduleHandler{}
			jobID := api.JobID("anchor-job")
			ssnJobs := map[api.JobID]*api.JobInfo{
				jobID: buildRoundJobInfo(jobID, tt.pending, 0, tt.scoped, nil),
			}
			vcJob := newDowngradeJob()
			if tt.disabled {
				vcJob.SchedulerJobAttr.ComJob.Annotation = nil
			}
			sHandle.Jobs = map[api.JobID]SchedulerJob{jobID: vcJob}

			sHandle.reconcileTimingState(ssnJobs)

			_, clockOn := cache.WaitStartTime(jobID)
			if clockOn != tt.wantClockOn {
				t.Errorf("wait clock = %v, want %v", clockOn, tt.wantClockOn)
			}
			if !tt.wantClockOn {
				return
			}
			if elapsed, ok := cache.SessionElapsed(jobID); !ok || elapsed != 0 {
				t.Errorf("session elapsed = %d, %v, want 0, true", elapsed, ok)
			}
		})
	}
}

// TestReconcileTimingStateAnchorKeepsFirstRecord covers the first record
// wins rule of the session-open anchor: a clock carried from the earlier sessions
// is never re-anchored, so the already waited time survives the session
// reconciliation.
func TestReconcileTimingStateAnchorKeepsFirstRecord(t *testing.T) {
	cache.ResetTimeoutCacheForTest()
	downgrade.ResetDowngradeCacheForTest()
	now := int64(1000)
	cache.SetTimeoutNowFuncForTest(func() time.Time { return time.Unix(now, 0) })
	sHandle := &ScheduleHandler{}
	jobID := api.JobID("anchor-job")
	ssnJobs := map[api.JobID]*api.JobInfo{
		jobID: buildRoundJobInfo(jobID, 4, 0, 4, nil),
	}
	sHandle.Jobs = map[api.JobID]SchedulerJob{jobID: newDowngradeJob()}
	cache.RecordWaitStart(jobID)

	now = 1010
	sHandle.reconcileTimingState(ssnJobs)

	waitStart, ok := cache.WaitStartTime(jobID)
	if !ok || waitStart != 1000 {
		t.Errorf("wait start = %d, %v, want 1000, true", waitStart, ok)
	}
	elapsed, elapsedOK := cache.SessionElapsed(jobID)
	if !elapsedOK || elapsed != 10 {
		t.Errorf("session elapsed = %d, %v, want 10, true", elapsed, elapsedOK)
	}
}
