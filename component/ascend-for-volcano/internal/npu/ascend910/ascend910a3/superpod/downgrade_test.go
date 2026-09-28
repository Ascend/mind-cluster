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
Package superpod is using for HuaWei Atlas 900 A3 SuperPod affinity schedule.
*/
package superpod

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/cache"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/downgrade"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910a3"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

type selectSuperPodDowngradeTest struct {
	name           string
	annotation     map[string]string
	preConfig      string // the downgrade config the store already holds
	expired        bool   // record an expired fail time before scheduling
	wantErr        bool
	wantRecord     bool // timeout record exists after scheduling
	wantDowngraded bool // the settled sp-block after scheduling is the downgraded one
	wantNodeNum    int
}

func buildSelectSuperPodDowngradeTestCases() []selectSuperPodDowngradeTest {
	return []selectSuperPodDowngradeTest{
		{
			name:        "01 downgrade enabled and not expired, return error without record",
			annotation:  downgradeEnabledAnno,
			wantErr:     true,
			wantRecord:  false,
			wantNodeNum: 0,
		},
		{
			name:           "02 downgrade enabled and expired, downgrade sp-block and schedule all tasks",
			annotation:     downgradeEnabledAnno,
			expired:        true,
			wantErr:        false,
			wantRecord:     true,
			wantDowngraded: true,
			wantNodeNum:    npuTaskNum4,
		},
		{
			name:        "03 downgrade disabled by default, keep the original constraint",
			expired:     true,
			wantErr:     true,
			wantRecord:  true,
			wantNodeNum: 0,
		},
		{
			name:           "04 a config of another shape falls back to the configured sp-block",
			annotation:     downgradeEnabledAnno,
			preConfig:      `{"strategy":"Rack"}`,
			expired:        true,
			wantRecord:     true,
			wantDowngraded: true,
			wantNodeNum:    npuTaskNum4,
		},
	}
}

// buildDowngradeFallbackNodes builds 4 nodes: 3 in super pod 0 and 1 in super
// pod 1, fitting the downgraded spBlock 1 but not the original spBlock 2.
func buildDowngradeFallbackNodes() ([]*api.NodeInfo, map[string]plugin.NPUNode) {
	nodeInfos := make([]*api.NodeInfo, 0, npuTaskNum4)
	npuNodes := make(map[string]plugin.NPUNode, npuTaskNum4)
	for i := 0; i < 3; i++ {
		npuNode := newNPUNodeWithSuperPodID("node0-"+strconv.Itoa(i), int32(0))
		npuNodes[npuNode.Name] = npuNode
		nodeInfos = append(nodeInfos, &api.NodeInfo{Name: npuNode.Name})
	}
	npuNode := newNPUNodeWithSuperPodID("node1-0", 1)
	npuNodes[npuNode.Name] = npuNode
	nodeInfos = append(nodeInfos, &api.NodeInfo{Name: npuNode.Name})
	return nodeInfos, npuNodes
}

var downgradeEnabledAnno = map[string]string{util.SchedulerDowngradeAnnoKey: "true"}

// fixtureWindow is the downgrade window the fixtures configure the handler with,
// the configuration layer defaults the value to DefaultSchedulerDowngradeTimeout.
var fixtureWindow = time.Duration(plugin.DefaultSchedulerDowngradeTimeout) * time.Second

// spBlockConfigString is the snapshot of a settled sp-block, in the npu number the
// sp-block of the job configuration is counted in, not the node number.
func spBlockConfigString(spBlock, maxNodeNPUNum int) string {
	return fmt.Sprintf(`{"sp-block":%d}`, spBlock*maxNodeNPUNum)
}

// newDowngradeSession builds a handler over the fallback nodes, the timeout
// state is the package global of common/cache shared across the sessions.
func newDowngradeSession(jobID api.JobID, annotation map[string]string) (*module910SuperPod,
	[]*api.NodeInfo) {
	plg := newDowngradeTestPlugin(jobID)
	plg.ComJob.Annotation = annotation
	nodeInfos, npuNodes := buildDowngradeFallbackNodes()
	plg.Nodes = npuNodes
	return plg, nodeInfos
}

// expireSession stamps an expired session snapshot of the job.
func expireSession(jobID api.JobID, over time.Duration, allWaiting bool) {
	cache.SetTimeoutNowFuncForTest(func() time.Time { return time.Now().Add(over) })
	cache.StampSession(map[api.JobID]bool{jobID: allWaiting})
	downgrade.StampSession(map[api.JobID]bool{jobID: allWaiting})
}

func TestSelectSuperPodForJobDowngrade(t *testing.T) {
	jobID := api.JobID("job-downgrade")
	for _, tt := range buildSelectSuperPodDowngradeTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			cache.ResetTimeoutCacheForTest()
			downgrade.ResetDowngradeCacheForTest()
			plg, nodeInfos := newDowngradeSession(jobID, tt.annotation)
			task := &api.TaskInfo{UID: "0", Job: jobID, Name: "task0"}

			if tt.expired {
				cache.RecordWaitStart(task.Job)
			}
			if tt.preConfig != "" {
				downgrade.RecordDowngrade(jobID, tt.preConfig, 500)
			}
			if tt.expired {
				// the wait covers the windows of a pre-recorded step too
				expireSession(task.Job, fixtureWindow*3+time.Minute, true)
			}
			// the downgrade settles at the post-validation hook, the node
			// selection below runs under the settled constraint
			plg.DowngradeConstraint(jobID)

			selectedNodes, err := plg.selectSuperPodForJob(task, nodeInfos, make(map[string]float64))
			if (err != nil) != tt.wantErr {
				t.Errorf("selectSuperPodForJob() error = %v, wantErr %v", err, tt.wantErr)
			}
			if nodeNum := countSuperPodNodes(selectedNodes); nodeNum != tt.wantNodeNum {
				t.Errorf("selectSuperPodForJob() selected %d nodes, want %d", nodeNum, tt.wantNodeNum)
			}
			_, recorded := cache.WaitStartTime(task.Job)
			if recorded != tt.wantRecord {
				t.Errorf("schedule timeout record is %v, want %v", recorded, tt.wantRecord)
			}
			// the settled constraint of the session and the snapshot the pods are
			// marked with, the pod mark itself is asserted by the base tests
			wantSpBlock := spBlockNum2
			wantConfig := ""
			if tt.wantDowngraded {
				wantSpBlock = spBlockNum1
				wantConfig = spBlockConfigString(wantSpBlock, plg.MaxNodeNPUNum)
			}
			if plg.spBlock != wantSpBlock {
				t.Errorf("sp-block = %d, want %d", plg.spBlock, wantSpBlock)
			}
			if config, ok := downgrade.DowngradeConfig(jobID); ok != tt.wantDowngraded ||
				(ok && config != wantConfig) {
				t.Errorf("downgrade config = %q, %v, want %q, %v",
					config, ok, wantConfig, tt.wantDowngraded)
			}
		})
	}
}

// TestSelectSuperPodForJobRestoreDowngrade verifies the effective level
// survives the session rebuild and is restored without deepening.
func TestSelectSuperPodForJobRestoreDowngrade(t *testing.T) {
	cache.ResetTimeoutCacheForTest()
	downgrade.ResetDowngradeCacheForTest()
	jobID := api.JobID("job-restore-downgrade")
	task := &api.TaskInfo{UID: "0", Job: jobID, Name: "task0"}

	// the downgrade session: the sp-block downgrades from 2 to 1 at the
	// post-validation hook, then the first task allocation marks the level
	// effective
	plg, nodeInfos := newDowngradeSession(jobID, downgradeEnabledAnno)
	cache.RecordWaitStart(jobID)
	expireSession(jobID, fixtureWindow+time.Minute, true)
	plg.DowngradeConstraint(jobID)
	selectedNodes, err := plg.selectSuperPodForJob(task, nodeInfos, make(map[string]float64))
	if err != nil {
		t.Fatalf("the downgrade session should schedule the tasks: %v", err)
	}
	if nodeNum := countSuperPodNodes(selectedNodes); nodeNum != npuTaskNum4 {
		t.Fatalf("the downgrade session selected %d nodes, want %d", nodeNum, npuTaskNum4)
	}
	wantEffectTime, ok := downgrade.DowngradeEffectTime(jobID)
	if !ok || wantEffectTime <= 0 {
		t.Fatalf("the downgrade session should record the effect time, got %d, %v", wantEffectTime, ok)
	}
	downgrade.MarkEffective(jobID)

	// the next session: the handler is rebuilt with the original sp-block, the
	// effective level is restored at the hook and never deepens
	nextPlg, nodeInfos := newDowngradeSession(jobID, downgradeEnabledAnno)
	if nextPlg.spBlock != spBlockNum2 {
		t.Fatalf("the rebuilt handler should start from the original sp-block %d", spBlockNum2)
	}
	expireSession(jobID, fixtureWindow*3+time.Minute, false)
	nextPlg.DowngradeConstraint(jobID)
	selectedNodes, err = nextPlg.selectSuperPodForJob(task, nodeInfos, make(map[string]float64))
	if err != nil {
		t.Fatalf("the restored session should schedule the tasks: %v", err)
	}
	if nodeNum := countSuperPodNodes(selectedNodes); nodeNum != npuTaskNum4 {
		t.Errorf("the restored session selected %d nodes, want %d", nodeNum, npuTaskNum4)
	}
	if nextPlg.spBlock != spBlockNum1 {
		t.Errorf("the restored sp-block = %d, want %d", nextPlg.spBlock, spBlockNum1)
	}
	// the restore session annotates the pods with what the downgrade session
	// wrote, so the effect time must not move
	if restored, _ := downgrade.DowngradeEffectTime(jobID); restored != wantEffectTime {
		t.Errorf("the restored session effect time = %d, want %d", restored, wantEffectTime)
	}
	config, _ := downgrade.DowngradeConfig(jobID)
	// the store snapshot keeps the settled sp-block, one node per block here
	wantConfig := spBlockConfigString(spBlockNum1, nextPlg.MaxNodeNPUNum)
	if config != wantConfig {
		t.Errorf("the effective constraint should not deepen, got config %q, want %q", config, wantConfig)
	}
}

// newDowngradeTestPlugin builds a fresh handler for every test case so that
// all session state (sp-block, timeout cache and downgrade mark) starts clean.
func newDowngradeTestPlugin(jobID api.JobID) *module910SuperPod {
	plg, _ := New(A3x16SchedulerName, ascend910a3.NodeNPUNumber16).(*module910SuperPod)
	plg.SchedulerJobAttr = util.SchedulerJobAttr{
		ComJob: util.ComJob{Name: jobID},
		NPUJob: &util.NPUJob{},
	}
	plg.ScheduleEnv = plugin.ScheduleEnv{}
	plg.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{jobID: {
		SchedulerJobAttr: util.SchedulerJobAttr{
			NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{
				"0": {ReqNPUName: "huawei.com/Ascend910", ReqNPUNum: 2},
			}},
		}}}
	// the fixture derives the sp-block without checkSpBlock, so it sets both
	// the session constraint and the configured snapshot it was derived from
	plg.spBlock = spBlockNum2
	plg.configuredSpBlock = spBlockNum2
	plg.MaxNodeNPUNum = 2
	plg.FrameAttr = plugin.VolcanoFrame{
		ConfigParameters: plugin.ConfigParameters{DynamicParameters: plugin.DynamicParameters{
			SuperPodSize:              superPodSize10,
			ReservePodSize:            reservePodSize2,
			SchedulerDowngradeTimeout: plugin.DefaultSchedulerDowngradeTimeout,
		}}}
	plg.NPUTaskNum = npuTaskNum4
	plg.Tasks = newNPUTasks(npuTaskNum4)
	return plg
}

type applySpBlockConfigTest struct {
	name        string
	config      string
	curSpBlock  int // the sp-block the session runs, in node number
	maxNodeNPU  int
	wantOK      bool
	wantSpBlock int // the settled sp-block, the current one when refused
}

func buildApplySpBlockConfigTestCases() []applySpBlockConfigTest {
	return []applySpBlockConfigTest{
		{name: "01 the npu number of one node is applied",
			config: `{"sp-block":2}`, curSpBlock: 2, maxNodeNPU: 2, wantOK: true, wantSpBlock: 1},
		{name: "02 the npu number of the whole block is refused",
			config: `{"sp-block":4}`, curSpBlock: 2, maxNodeNPU: 2, wantSpBlock: 2},
		{name: "03 a non-aligned npu number is refused instead of truncated",
			config: `{"sp-block":5}`, curSpBlock: 4, maxNodeNPU: 2, wantSpBlock: 4},
		{name: "04 a config of another shape is refused",
			config: `{"strategy":"Rack"}`, curSpBlock: 4, maxNodeNPU: 2, wantSpBlock: 4},
		{name: "05 a handler without a node npu is refused",
			config: `{"sp-block":2}`, curSpBlock: 2, wantSpBlock: 2},
		{name: "06 a non-positive npu number is refused",
			config: `{"sp-block":0}`, curSpBlock: 2, maxNodeNPU: 2, wantSpBlock: 2},
	}
}

// TestApplySpBlockConfig covers the apply guard: every value which does not lower
// the sp-block of the session to a whole node num is refused, not truncated.
func TestApplySpBlockConfig(t *testing.T) {
	for _, tt := range buildApplySpBlockConfigTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			plg, _ := New(A3x16SchedulerName, ascend910a3.NodeNPUNumber16).(*module910SuperPod)
			plg.spBlock = tt.curSpBlock
			plg.MaxNodeNPUNum = tt.maxNodeNPU

			if got := plg.applySpBlockConfig(tt.config); got != tt.wantOK {
				t.Errorf("applySpBlockConfig(%q) = %v, want %v", tt.config, got, tt.wantOK)
			}
			if plg.spBlock != tt.wantSpBlock {
				t.Errorf("sp-block = %d, want %d", plg.spBlock, tt.wantSpBlock)
			}
		})
	}
}

// TestDowngradeConstraintNilHandler covers the nil guard of the hook.
func TestDowngradeConstraintNilHandler(t *testing.T) {
	var plg *module910SuperPod

	plg.DowngradeConstraint(api.JobID("job-nil-handler"))
}

// TestDowngradeConstraintWithoutConfiguredSpBlock covers the handler which never ran
// checkSpBlock: the ladder holds no baseline, so the session keeps its constraint.
func TestDowngradeConstraintWithoutConfiguredSpBlock(t *testing.T) {
	cache.ResetTimeoutCacheForTest()
	downgrade.ResetDowngradeCacheForTest()
	jobID := api.JobID("job-no-configured-sp-block")
	plg, _ := newDowngradeSession(jobID, downgradeEnabledAnno)
	plg.configuredSpBlock = 0
	cache.RecordWaitStart(jobID)
	expireSession(jobID, fixtureWindow+time.Minute, true)

	plg.DowngradeConstraint(jobID)

	if plg.spBlock != spBlockNum2 {
		t.Errorf("sp-block = %d, want the kept %d", plg.spBlock, spBlockNum2)
	}
	if config, ok := downgrade.DowngradeConfig(jobID); ok {
		t.Errorf("no downgrade should be recorded without a configured sp-block, got %q", config)
	}
}

type spBlockOfWaitedWindowsTest struct {
	name        string
	configured  int // the configured sp-block in node number
	nodeNPUNum  int
	windows     int
	wantSpBlock int // the settled sp-block in npu number, zero means no ladder
	wantLevel   int
	wantOK      bool
}

func buildSpBlockOfWaitedWindowsTestCases() []spBlockOfWaitedWindowsTest {
	return []spBlockOfWaitedWindowsTest{
		{name: "01 no waited window holds the configured sp-block", configured: 2, nodeNPUNum: 2},
		{name: "02 the first window halves the sp-block", configured: 4, nodeNPUNum: 2, windows: 1,
			wantSpBlock: 4, wantLevel: 1, wantOK: true},
		{name: "03 a long wait crosses to the bottom line at once", configured: 4, nodeNPUNum: 2,
			windows: 3, wantSpBlock: 2, wantLevel: 2, wantOK: true},
		{name: "04 a wait deeper than the ladder settles at its bottom level",
			configured: 4, nodeNPUNum: 2, windows: 9, wantSpBlock: 2, wantLevel: 2, wantOK: true},
		{name: "05 the half which does not fill whole nodes falls to the bottom line",
			configured: 3, nodeNPUNum: 2, windows: 1, wantSpBlock: 2, wantLevel: 1, wantOK: true},
		{name: "06 the sp-block of one node holds no ladder", configured: 1, nodeNPUNum: 2, windows: 3},
		{name: "07 a handler without a configured sp-block holds no ladder",
			configured: 0, nodeNPUNum: 2, windows: 3},
		{name: "08 a handler without a node npu holds no ladder",
			configured: 2, nodeNPUNum: 0, windows: 3},
	}
}

// TestSpBlockOfWaitedWindows covers the one shot ladder: the settled sp-block halves the
// configuration once per waited window, a level which does not fill whole nodes falls to
// the bottom line directly and every deeper wait settles there too.
func TestSpBlockOfWaitedWindows(t *testing.T) {
	for _, tt := range buildSpBlockOfWaitedWindowsTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			plg := &module910SuperPod{}
			plg.configuredSpBlock = tt.configured
			plg.MaxNodeNPUNum = tt.nodeNPUNum

			config, level, ok := plg.spBlockOfWaitedWindows(tt.windows)

			if ok != tt.wantOK || level != tt.wantLevel || spBlockNPUNum(t, config) != tt.wantSpBlock {
				t.Errorf("spBlockOfWaitedWindows(%d) = %q, %d, %v, want %d npu, %d, %v",
					tt.windows, config, level, ok, tt.wantSpBlock, tt.wantLevel, tt.wantOK)
			}
		})
	}
}

// spBlockStepByStep is the ladder as the common entry walked it before: one halving per
// level from the configured sp-block, the level the ladder settled at reported back.
func spBlockStepByStep(configured, nodeNPUNum, windows int) (int, int) {
	npuNum, level := configured*nodeNPUNum, 0
	for level < windows {
		half := npuNum / util.DefaultDowngradedFactor
		if half < nodeNPUNum || half%nodeNPUNum != 0 {
			half = nodeNPUNum
		}
		if half >= npuNum {
			break
		}
		npuNum, level = half, level+1
	}
	return npuNum, level
}

// TestSpBlockOfWaitedWindowsMatchesTheLadder pins the one shot computation to the step
// by step walk it replaces: the same settled sp-block and the same level for every
// configured sp-block and every wait.
func TestSpBlockOfWaitedWindowsMatchesTheLadder(t *testing.T) {
	for _, nodeNPUNum := range []int{2, 8} {
		for configured := 1; configured <= 16; configured++ {
			for windows := 1; windows <= 8; windows++ {
				plg := &module910SuperPod{}
				plg.configuredSpBlock = configured
				plg.MaxNodeNPUNum = nodeNPUNum
				config, level, ok := plg.spBlockOfWaitedWindows(windows)
				wantNPUNum, wantLevel := spBlockStepByStep(configured, nodeNPUNum, windows)

				if ok != (wantLevel > 0) || (ok && (level != wantLevel ||
					spBlockNPUNum(t, config) != wantNPUNum)) {
					t.Errorf("configured %d nodes, node npu %d, %d windows: got %q, %d, %v, "+
						"want %d npu, %d", configured, nodeNPUNum, windows, config, level, ok,
						wantNPUNum, wantLevel)
				}
			}
		}
	}
}

// spBlockNPUNum reads the npu number of the sp-block of a settled snapshot.
func spBlockNPUNum(t *testing.T, config string) int {
	t.Helper()
	if config == "" {
		return 0
	}
	var cfg spBlockConfig
	if err := json.Unmarshal([]byte(config), &cfg); err != nil {
		t.Errorf("the settled config %q is not a readable sp-block snapshot: %v", config, err)
		return 0
	}
	return cfg.SpBlock
}

func countSuperPodNodes(selectedNodes map[string][]plugin.SuperNode) int {
	count := 0
	for _, sp := range selectedNodes {
		count += len(sp)
	}
	return count
}

type specialJobDowngradeTest struct {
	name     string
	jobID    api.JobID
	label    map[string]string
	kindName string
	isKind   func(*module910SuperPod) bool
}

func buildSpecialJobDowngradeTestCases() []specialJobDowngradeTest {
	return []specialJobDowngradeTest{
		{
			name:  "01 the mindie job downgrades its sp-block",
			jobID: "job-mindie-downgrade",
			label: map[string]string{mindIEJobAppLabelKey: "mindie", mindIEJobIDLabelKey: "0"},
			// the MindIE layout derives from the sp-block, the hook settles the
			// constraint before the whole session, the MindIE split included
			kindName: "mindie", isKind: (*module910SuperPod).isMindIEJob,
		},
		{
			name:     "02 the infer service job downgrades its sp-block",
			jobID:    "job-infer-service-downgrade",
			label:    map[string]string{"inferserviceid": "infer-service-1"},
			kindName: "infer service", isKind: (*module910SuperPod).isInferServiceJobCheck,
		},
	}
}

// TestSpecialJobDowngradesSpBlock verifies the special job kinds downgrade like the
// other jobs: the MindIE layout and the infer service node selection both derive
// from the sp-block, the post-validation hook settles the downgraded constraint for
// them too.
func TestSpecialJobDowngradesSpBlock(t *testing.T) {
	for _, tt := range buildSpecialJobDowngradeTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			cache.ResetTimeoutCacheForTest()
			downgrade.ResetDowngradeCacheForTest()
			plg, _ := newDowngradeSession(tt.jobID, downgradeEnabledAnno)
			plg.Label = tt.label
			cache.RecordWaitStart(tt.jobID)
			expireSession(tt.jobID, fixtureWindow+time.Minute, true)

			if !tt.isKind(plg) {
				t.Fatalf("the test fixture should be a %s job", tt.kindName)
			}
			plg.DowngradeConstraint(tt.jobID)

			if plg.spBlock != spBlockNum1 {
				t.Errorf("%s job sp-block = %d, want the downgraded %d",
					tt.kindName, plg.spBlock, spBlockNum1)
			}
			// the downgrade is recorded too, not only applied in the session
			config, ok := downgrade.DowngradeConfig(tt.jobID)
			wantConfig := spBlockConfigString(spBlockNum1, plg.MaxNodeNPUNum)
			if !ok || config != wantConfig {
				t.Errorf("%s job downgrade config = %q, %v, want %q", tt.kindName, config, ok, wantConfig)
			}
		})
	}
}
