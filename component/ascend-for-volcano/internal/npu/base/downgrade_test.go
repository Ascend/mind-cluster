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
Package base is using for HuaWei Ascend affinity schedule base.
*/
package base

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/cache"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/downgrade"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

var downgradeEnabledAnno = map[string]string{util.SchedulerDowngradeAnnoKey: "true"}

// fixtureWindow is the downgrade window the fixtures configure the handler with,
// the configuration layer defaults the value to DefaultSchedulerDowngradeTimeout.
var fixtureWindow = time.Duration(plugin.DefaultSchedulerDowngradeTimeout) * time.Second

// fixtureEffectTime is the effect time the fixtures settle a pre-recorded snapshot
// with, the one of a settled level which a restore session must keep.
const fixtureEffectTime = int64(1000)

// sizeConfig is the fixture constraint of the node-group policies; the size in
// npu number, stored as a JSON snapshot the policy serializes.
type sizeConfig struct {
	Size int `json:"v"`
}

// sizeConfigString is the snapshot the fixture policy serializes.
func sizeConfigString(size int) string {
	return fmt.Sprintf(`{"v":%d}`, size)
}

// nodeGroupPolicy is a stub node-group policy over the common entry: baseSize is the
// configuration baseline of the ladder (like the configuredSpBlock snapshot of the
// real policies, never changed by a downgrade), curSize is the constraint the session
// runs (like the spBlock of the real policies), seen records the window counts the
// common entry asked the ladder for.
type nodeGroupPolicy struct {
	tp       *NPUHandler
	baseSize int
	curSize  int
	applied  []int // the sizes the policy applied to the session, in order
	seen     []int // the window counts the ladder was asked for, in order
}

// newNodeGroupPolicy builds the stub, the session starting from the baseline.
func newNodeGroupPolicy(tp *NPUHandler, baseSize int) *nodeGroupPolicy {
	return &nodeGroupPolicy{tp: tp, baseSize: baseSize, curSize: baseSize}
}

// next is the ladder of the node-group policies: the constraint of the level the waited
// windows buy, halving the baseline once per window, a half which does not fill whole
// nodes falling to the bottom line directly. It reports the level the constraint settled
// at, false when the session holds no lower constraint.
func (p *nodeGroupPolicy) next(windows int) (string, int, bool) {
	p.seen = append(p.seen, windows)
	npuNum, level := p.baseSize, 0
	for level < windows {
		half := npuNum / util.DefaultDowngradedFactor
		if half < p.tp.MaxNodeNPUNum || half%p.tp.MaxNodeNPUNum != 0 {
			half = p.tp.MaxNodeNPUNum
		}
		if half >= npuNum {
			break
		}
		npuNum, level = half, level+1
	}
	if level == 0 {
		return "", 0, false
	}
	return sizeConfigString(npuNum), level, true
}

// apply keeps the constraint of the session, a size which does not lower it is
// refused, the guard of the real policies and of the common restore.
func (p *nodeGroupPolicy) apply(config string) bool {
	var cfg sizeConfig
	if json.Unmarshal([]byte(config), &cfg) != nil || cfg.Size <= 0 || cfg.Size >= p.curSize {
		return false
	}
	p.curSize = cfg.Size
	p.applied = append(p.applied, cfg.Size)
	return true
}

type isDowngradeEnabledTest struct {
	name       string
	annotation map[string]string
	want       bool
}

func buildIsDowngradeEnabledTestCases() []isDowngradeEnabledTest {
	return []isDowngradeEnabledTest{
		{name: "01 disabled when the annotation is absent"},
		{name: "02 disabled when the annotation is not true",
			annotation: map[string]string{util.SchedulerDowngradeAnnoKey: "yes"}},
		{name: "03 enabled when the annotation is true",
			annotation: map[string]string{util.SchedulerDowngradeAnnoKey: "true"}, want: true},
	}
}

func TestIsDowngradeEnabled(t *testing.T) {
	for _, tt := range buildIsDowngradeEnabledTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			tp := &NPUHandler{}
			tp.SchedulerJobAttr = util.SchedulerJobAttr{
				ComJob: util.ComJob{Annotation: tt.annotation},
			}
			if got := tp.IsDowngradeEnabled(); got != tt.want {
				t.Errorf("IsDowngradeEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

type setDowngradedToPodTest struct {
	name       string
	effectTime int64  // the session downgrade mark, 0 means no downgrade took effect
	config     string // the session constraint snapshot, empty means none
	preMark    string // the existing pod marker value, empty means none
	wantMark   string // the expected marker value, empty means none
	wantConfig string // the expected config annotation, empty means none
}

func buildSetDowngradedToPodTestCases() []setDowngradedToPodTest {
	return []setDowngradedToPodTest{
		{name: "01 downgrade mark set writes the effect time and the config",
			effectTime: 100000900, config: `{"v":2}`, wantMark: "100000900", wantConfig: `{"v":2}`},
		{name: "02 no downgrade mark writes nothing"},
		{name: "03 keeps the existing marker, the config is not written either",
			effectTime: 100000900, config: `{"v":2}`, preMark: "999", wantMark: "999"},
		{name: "04 the config is part of the mark, a missing config writes nothing",
			effectTime: 100000900},
	}
}

func TestSetDowngradedToPod(t *testing.T) {
	jobID := api.JobID("downgrade-annotation-job")
	for _, tt := range buildSetDowngradedToPodTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			tp := &NPUHandler{}
			tp.downgradedEffectTime = tt.effectTime
			tp.downgradedConfig = tt.config

			annotations := map[string]string{"existing": "keep"}
			if tt.preMark != "" {
				annotations[util.SchedulerDowngradedAnnoKey] = tt.preMark
			}
			pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "task0", Namespace: "default",
				Annotations: annotations}}
			task := &api.TaskInfo{Name: pod.Name, Job: jobID, Pod: pod}

			tp.setDowngradedToPod(task)

			got := task.Pod.Annotations[util.SchedulerDowngradedAnnoKey]
			if got != tt.wantMark {
				t.Errorf("downgraded annotation = %q, want %q", got, tt.wantMark)
			}
			gotConfig := task.Pod.Annotations[util.SchedulerDowngradedLevelAnnoKey]
			if gotConfig != tt.wantConfig {
				t.Errorf("downgraded config annotation = %q, want %q", gotConfig, tt.wantConfig)
			}
			if got := task.Pod.Annotations["existing"]; got != "keep" {
				t.Errorf("existing annotation was changed to %q", got)
			}
		})
	}
}

func TestMarkDowngraded(t *testing.T) {
	downgrade.ResetDowngradeCacheForTest()
	jobID := api.JobID("mark-downgrade-job")
	tp := &NPUHandler{}
	downgrade.RecordDowngrade(jobID, sizeConfigString(2), 1000)
	config, effectTime, ok := downgrade.DowngradeState(jobID)
	if !ok {
		t.Fatalf("the settled state should be readable")
	}
	tp.markDowngraded(config, effectTime)
	if tp.downgradedEffectTime != 1000 || tp.downgradedConfig != sizeConfigString(2) {
		t.Errorf("downgrade mark = %d, %q, want 1000, %q",
			tp.downgradedEffectTime, tp.downgradedConfig, sizeConfigString(2))
	}
	// the mark is a copy of the pair the caller settled: dropping the state does
	// not move it, the marker annotations are written at bind time from the mark
	downgrade.ResetConstraint(jobID)
	if tp.downgradedEffectTime != 1000 || tp.downgradedConfig != sizeConfigString(2) {
		t.Errorf("downgrade mark = %d, %q, want 1000, %q after the state is cleared",
			tp.downgradedEffectTime, tp.downgradedConfig, sizeConfigString(2))
	}
}

// newDowngradeTestHandler builds a handler over the package-wide states, holding a
// wait clock and an all-waiting session snapshot of the job.
func newDowngradeTestHandler(jobID api.JobID, annotation map[string]string) *NPUHandler {
	cache.ResetTimeoutCacheForTest()
	downgrade.ResetDowngradeCacheForTest()
	tp := &NPUHandler{}
	tp.SchedulerJobAttr = util.SchedulerJobAttr{ComJob: util.ComJob{Annotation: annotation}}
	tp.MaxNodeNPUNum = 8
	tp.FrameAttr.SchedulerDowngradeTimeout = plugin.DefaultSchedulerDowngradeTimeout
	cache.RecordWaitStart(jobID)
	cache.StampSession(map[api.JobID]bool{jobID: true})
	downgrade.StampSession(map[api.JobID]bool{jobID: true})
	return tp
}

// stampWait moves the fake clock over the given wait and rebuilds the snapshots of
// both packages, the wait is what the session reads from then on.
func stampWait(jobID api.JobID, waited time.Duration, allWaiting bool) {
	cache.SetTimeoutNowFuncForTest(func() time.Time { return time.Now().Add(waited) })
	cache.StampSession(map[api.JobID]bool{jobID: allWaiting})
	downgrade.StampSession(map[api.JobID]bool{jobID: allWaiting})
}

// assertInts verifies a recorded sequence of the stub policy, in order.
func assertInts(t *testing.T, got, want []int, what string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s = %v, want %v", what, got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s = %v, want %v", what, got, want)
			return
		}
	}
}

type downgradeConstraintTest struct {
	name       string
	annotation map[string]string
	waited     time.Duration // the wait of the session snapshot, zero keeps the clock fresh
	partial    bool          // part of the round is scheduled: the settled level is kept
	preConfig  string        // the settled snapshot the store already holds
	baseSize   int           // the configuration baseline of the ladder, in npu number
	wantSeen   []int         // the window counts the ladder was asked for, in order
	wantSizes  []int         // the sizes the policy applied, in order
	wantConfig string        // the settled snapshot afterwards, empty means none
	wantLevel  int           // the level the derived snapshot belongs to
}

func buildDowngradeConstraintTestCases() []downgradeConstraintTest {
	return []downgradeConstraintTest{
		{name: "01 disabled by default, the ladder is not called", baseSize: 32},
		{name: "02 enabled but the wait buys no level", annotation: downgradeEnabledAnno, baseSize: 32},
		{name: "03 one waited window settles the first level", annotation: downgradeEnabledAnno,
			waited: fixtureWindow + time.Second, baseSize: 32,
			wantSeen: []int{1}, wantSizes: []int{16}, wantConfig: sizeConfigString(16), wantLevel: 1},
		{name: "04 three waited windows are asked for in one call",
			annotation: downgradeEnabledAnno, waited: fixtureWindow*3 + time.Second, baseSize: 32,
			wantSeen: []int{3}, wantSizes: []int{8}, wantConfig: sizeConfigString(8), wantLevel: 2},
		{name: "05 a ladder at its bottom settles no deeper level",
			annotation: downgradeEnabledAnno, waited: fixtureWindow*3 + time.Second, baseSize: 8,
			wantSeen: []int{3}},
		{name: "06 an all waiting round derives the level, the settled one is not restored",
			annotation: downgradeEnabledAnno, waited: fixtureWindow + time.Second,
			preConfig: sizeConfigString(8), baseSize: 32,
			wantSeen: []int{1}, wantSizes: []int{16}, wantConfig: sizeConfigString(16), wantLevel: 1},
	}
}

// TestDowngradeConstraintIfTimeout covers the branch of the common entry which
// derives the constraint of the wait: the waited windows are handed to the ladder in
// one call, so a long wait settles several steps at once.
func TestDowngradeConstraintIfTimeout(t *testing.T) {
	jobID := api.JobID("downgrade-constraint-job")
	for _, tt := range buildDowngradeConstraintTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			tp := newDowngradeTestHandler(jobID, tt.annotation)
			if tt.preConfig != "" {
				downgrade.RecordDowngrade(jobID, tt.preConfig, fixtureEffectTime)
			}
			stampWait(jobID, tt.waited, true)
			policy := newNodeGroupPolicy(tp, tt.baseSize)

			tp.DowngradeConstraintIfTimeout(jobID, policy.next, policy.apply)

			assertInts(t, policy.seen, tt.wantSeen, "asked windows")
			assertInts(t, policy.applied, tt.wantSizes, "applied sizes")
			config, ok := downgrade.DowngradeConfig(jobID)
			if ok != (tt.wantConfig != "") || (ok && config != tt.wantConfig) {
				t.Errorf("settled config = %q, %v, want %q", config, ok, tt.wantConfig)
			}
			if tt.wantConfig == "" {
				return
			}
			waitStart, _ := cache.WaitStartTime(jobID)
			wantEffectTime := waitStart + int64(fixtureWindow*time.Duration(tt.wantLevel)/time.Second)
			if effectTime, _ := downgrade.DowngradeEffectTime(jobID); effectTime != wantEffectTime {
				t.Errorf("effect time = %d, want %d", effectTime, wantEffectTime)
			}
		})
	}
}

type downgradeRestoreTest struct {
	name      string
	waited    time.Duration // the wait of the session snapshot
	preConfig string        // the settled snapshot the store already holds
	baseSize  int           // the configuration baseline of the ladder
	wantSizes []int         // the sizes the policy applied, in order
}

func buildDowngradeRestoreTestCases() []downgradeRestoreTest {
	return []downgradeRestoreTest{
		{name: "01 the partially scheduled round keeps its settled level",
			waited: fixtureWindow*3 + time.Second, preConfig: sizeConfigString(16), baseSize: 32,
			wantSizes: []int{16}},
		{name: "02 the settled level is restored before the next window is waited",
			preConfig: sizeConfigString(16), baseSize: 32, wantSizes: []int{16}},
		{name: "03 a restore the policy refuses is ignored",
			waited: fixtureWindow*3 + time.Second, preConfig: sizeConfigString(32), baseSize: 32},
		{name: "04 a round without a settled level restores nothing",
			waited: fixtureWindow*3 + time.Second, baseSize: 32},
	}
}

// TestDowngradeConstraintRestore verifies the settled level is kept while the round
// holds tasks scheduled under it, without deriving the level of the wait: the level
// of a partially scheduled round never advances, the effect time does not move.
func TestDowngradeConstraintRestore(t *testing.T) {
	jobID := api.JobID("downgrade-restore-job")
	for _, tt := range buildDowngradeRestoreTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			tp := newDowngradeTestHandler(jobID, downgradeEnabledAnno)
			if tt.preConfig != "" {
				downgrade.RecordDowngrade(jobID, tt.preConfig, fixtureEffectTime)
			}
			// something of the round is scheduled: the snapshot is not all-waiting
			stampWait(jobID, tt.waited, false)
			policy := newNodeGroupPolicy(tp, tt.baseSize)

			tp.DowngradeConstraintIfTimeout(jobID, policy.next, policy.apply)

			// the ladder keeps the settled constraint of the round, it is never asked
			assertInts(t, policy.seen, nil, "asked windows")
			assertInts(t, policy.applied, tt.wantSizes, "applied sizes")
			if config, ok := downgrade.DowngradeConfig(jobID); ok != (tt.preConfig != "") ||
				(ok && config != tt.preConfig) {
				t.Errorf("settled config = %q, %v, want %q", config, ok, tt.preConfig)
			}
			if tt.preConfig != "" {
				if effectTime, _ := downgrade.DowngradeEffectTime(jobID); effectTime != fixtureEffectTime {
					t.Errorf("effect time = %d, want the settled %d", effectTime, fixtureEffectTime)
				}
			}
		})
	}
}

// TestDowngradeConstraintOncePerSession verifies the hook settles one constraint per
// session: the level is a pure function of the session snapshot, so the repeated
// invocations of the hook (the cycle state init plus every scheduling action)
// derive the same level instead of stepping once per invocation.
func TestDowngradeConstraintOncePerSession(t *testing.T) {
	jobID := api.JobID("downgrade-once-job")
	tp := newDowngradeTestHandler(jobID, downgradeEnabledAnno)
	// the job waited three whole windows, every one of them would allow one more level
	stampWait(jobID, fixtureWindow*3+time.Second, true)
	policy := newNodeGroupPolicy(tp, 32)

	// the hook runs from the cycle state init and from every scheduling action
	for i := 0; i < 3; i++ {
		tp.DowngradeConstraintIfTimeout(jobID, policy.next, policy.apply)
	}

	assertInts(t, policy.seen, []int{3, 3, 3}, "asked windows")
	assertInts(t, policy.applied, []int{8}, "applied sizes")
	if policy.curSize != 8 {
		t.Errorf("the settled constraint = %d, want 8", policy.curSize)
	}
	config, ok := downgrade.DowngradeConfig(jobID)
	if !ok || config != sizeConfigString(8) {
		t.Errorf("settled config = %q, %v, want %q, true", config, ok, sizeConfigString(8))
	}
	waitStart, _ := cache.WaitStartTime(jobID)
	wantEffectTime := waitStart + int64(fixtureWindow*2/time.Second)
	if effectTime, _ := downgrade.DowngradeEffectTime(jobID); effectTime != wantEffectTime {
		t.Errorf("effect time = %d, want the settled %d", effectTime, wantEffectTime)
	}
}

type waitedWindowsTest struct {
	name   string
	waited time.Duration
	want   int
	wantOK bool
}

func buildWaitedWindowsTestCases() []waitedWindowsTest {
	return []waitedWindowsTest{
		{name: "01 a fresh clock buys no window", waited: time.Second},
		{name: "02 a wait below one window buys no window", waited: fixtureWindow - 2*time.Second},
		{name: "03 one whole window buys the first level", waited: fixtureWindow, want: 1, wantOK: true},
		{name: "04 a wait between two windows stays at the first level",
			waited: fixtureWindow + fixtureWindow/2, want: 1, wantOK: true},
		{name: "05 three whole windows buy the third level",
			waited: fixtureWindow * 3, want: 3, wantOK: true},
	}
}

// TestDowngradeConstraintWindowCount covers the windows of the wait: the n-th level
// needs n whole windows, measured on the session-start snapshot, and a wait between two
// windows stays at the level it bought.
func TestDowngradeConstraintWindowCount(t *testing.T) {
	jobID := api.JobID("downgrade-window-job")
	for _, tt := range buildWaitedWindowsTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			tp := newDowngradeTestHandler(jobID, downgradeEnabledAnno)
			stampWait(jobID, tt.waited, true)

			windows, ok := tp.waitedWindows(jobID, fixtureWindow)

			if ok != tt.wantOK || windows != tt.want {
				t.Errorf("waitedWindows() = %d, %v, want %d, %v", windows, ok, tt.want, tt.wantOK)
			}
		})
	}
	// a job without a snapshot and a non-positive window buy nothing
	tp := newDowngradeTestHandler(jobID, downgradeEnabledAnno)
	if windows, ok := tp.waitedWindows(jobID, 0); ok || windows != 0 {
		t.Errorf("a non-positive window should buy no level, got %d, %v", windows, ok)
	}
	if windows, ok := tp.waitedWindows("job-without-clock", fixtureWindow); ok || windows != 0 {
		t.Errorf("a job without a snapshot should buy no level, got %d, %v", windows, ok)
	}
	if windows, ok := tp.waitedWindows("", fixtureWindow); ok || windows != 0 {
		t.Errorf("an empty job should buy no level, got %d, %v", windows, ok)
	}
}

// TestDowngradeEffectTime covers the effect time of a settled level: it is the end
// of the window which bought the level, zero when the job holds no clock.
func TestDowngradeEffectTime(t *testing.T) {
	jobID := api.JobID("downgrade-effect-time-job")
	tp := newDowngradeTestHandler(jobID, downgradeEnabledAnno)
	waitStart, _ := cache.WaitStartTime(jobID)

	if effectTime := tp.downgradeEffectTime(jobID, fixtureWindow, 2); effectTime != waitStart+240 {
		t.Errorf("effect time = %d, want %d", effectTime, waitStart+240)
	}
	// the settled clock of the job is dropped by the round end, the effect time of
	// a level without a clock is not derivable
	cache.EndRound(jobID)
	if effectTime := tp.downgradeEffectTime(jobID, fixtureWindow, 2); effectTime != 0 {
		t.Errorf("effect time without a clock = %d, want 0", effectTime)
	}
}

// TestDowngradeConstraintEmptyInput covers the entry guards of the common entry: a nil
// handler, an empty job, missing callbacks and a disabled switch are no-ops.
func TestDowngradeConstraintEmptyInput(t *testing.T) {
	jobID := api.JobID("downgrade-guard-job")
	tp := newDowngradeTestHandler(jobID, downgradeEnabledAnno)
	policy := newNodeGroupPolicy(tp, 32)
	stampWait(jobID, fixtureWindow*3+time.Second, true)

	tp.DowngradeConstraintIfTimeout("", policy.next, policy.apply)
	tp.DowngradeConstraintIfTimeout(jobID, nil, policy.apply)
	tp.DowngradeConstraintIfTimeout(jobID, policy.next, nil)
	disabled := newDowngradeTestHandler(jobID, nil)
	disabled.DowngradeConstraintIfTimeout(jobID, policy.next, policy.apply)
	var nilHandler *NPUHandler
	nilHandler.DowngradeConstraintIfTimeout(jobID, policy.next, policy.apply)

	assertInts(t, policy.seen, nil, "asked windows")
	assertInts(t, policy.applied, nil, "applied sizes")
}
