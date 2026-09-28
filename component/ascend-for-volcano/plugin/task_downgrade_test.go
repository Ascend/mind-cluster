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
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/downgrade"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

func newPlacedNPUTask(uid string, status api.TaskStatus, annotation map[string]string) *api.TaskInfo {
	task := &api.TaskInfo{
		UID:    api.TaskID(uid),
		Name:   uid,
		Resreq: &api.Resource{ScalarResources: map[v1.ResourceName]float64{"huawei.com/Ascend910": 16}},
	}
	task.Status = status
	if status == api.Bound || status == api.Running || status == api.Succeeded {
		task.NodeName = "node-" + uid
	}
	if annotation != nil {
		task.Pod = &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: uid, Annotations: annotation}}
	}
	return task
}

// barePodNPUTask builds a placed NPU task whose pod carries no annotation map.
func barePodNPUTask(uid string) *api.TaskInfo {
	task := newPlacedNPUTask(uid, api.Bound, nil)
	task.Pod = &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: uid}}
	return task
}

// newTerminatingNPUTask builds a placed NPU task whose pod is being deleted.
func newTerminatingNPUTask(uid string, annotation map[string]string) *api.TaskInfo {
	task := newPlacedNPUTask(uid, api.Bound, annotation)
	now := metav1.Now()
	task.Pod.DeletionTimestamp = &now
	return task
}

// the constraint snapshots of the downgrade fixtures, the JSON object shape the
// policy handlers write to the pod markers, and the scalar of the previous version
const (
	configSpBlock1 = `{"sp-block":1}`
	configSpBlock2 = `{"sp-block":2}`
	legacyConfig   = `"16"`
	// seedTraversalRounds is how often a seed read is repeated, the traversal
	// order of the task map of one job is random
	seedTraversalRounds = 5
)

// downgradeMarker builds a complete downgrade marker annotation pair.
func downgradeMarker(effectTime, config string) map[string]string {
	return map[string]string{
		util.SchedulerDowngradedAnnoKey:      effectTime,
		util.SchedulerDowngradedLevelAnnoKey: config,
	}
}

func TestCountRoundTasks(t *testing.T) {
	completeMarker := downgradeMarker("1000", configSpBlock1)
	vcJob := &api.JobInfo{Tasks: map[api.TaskID]*api.TaskInfo{
		"task0": newPlacedNPUTask("task0", api.Bound, completeMarker),
		"task1": newPlacedNPUTask("task1", api.Bound, completeMarker),
		"task2": newPlacedNPUTask("task2", api.Running, nil),
		"task3": newPlacedNPUTask("task3", api.Succeeded, nil),
		"task4": newPlacedNPUTask("task4", api.Pending, nil),
		"task5": newPlacedNPUTask("task5", api.Pipelined, nil),
		// the terminating task is out of the scheduling scope
		"task6": newTerminatingNPUTask("task6", completeMarker),
	}}
	// the non-NPU task is not counted even when it is placed
	nonNPUTask := &api.TaskInfo{UID: "task7"}
	nonNPUTask.NodeName = "node-task7"
	nonNPUTask.Status = api.Bound
	vcJob.Tasks["task7"] = nonNPUTask
	count := countRoundTasks(vcJob)

	if count.scoped != 6 {
		t.Errorf("scoped task num = %d, want 6 (the in-scope npu tasks)", count.scoped)
	}
	if count.pending != 2 {
		t.Errorf("pending task num = %d, want 2 (the tasks without a node)", count.pending)
	}
	if count.scheduled != 4 {
		t.Errorf("scheduled task num = %d, want 4 (the placed npu tasks)", count.scheduled)
	}
}

type readDowngradeSeedTest struct {
	name       string
	tasks      map[api.TaskID]*api.TaskInfo
	wantConfig string
	wantTime   int64
}

func buildReadDowngradeSeedTestCases() []readDowngradeSeedTest {
	oldMarker := downgradeMarker("1000", configSpBlock2)
	newMarker := downgradeMarker("2000", configSpBlock1)
	return []readDowngradeSeedTest{
		{
			name: "01 the marker of the placed pod is read",
			tasks: map[api.TaskID]*api.TaskInfo{
				"task0": newPlacedNPUTask("task0", api.Bound, oldMarker)},
			wantConfig: configSpBlock2, wantTime: 1000,
		},
		{
			name: "02 the terminating pod marker is never used",
			tasks: map[api.TaskID]*api.TaskInfo{
				"task0": newTerminatingNPUTask("task0", oldMarker)},
		},
		{
			name: "03 a partial marker pair is ignored",
			tasks: map[api.TaskID]*api.TaskInfo{
				"task0": newPlacedNPUTask("task0", api.Bound,
					map[string]string{util.SchedulerDowngradedAnnoKey: "1000"})},
		},
		{
			name: "03b a marker pair missing the effect time is ignored",
			tasks: map[api.TaskID]*api.TaskInfo{
				"task0": newPlacedNPUTask("task0", api.Bound,
					map[string]string{util.SchedulerDowngradedLevelAnnoKey: configSpBlock2})},
		},
		{
			name: "03c a pod without a marker is ignored",
			tasks: map[api.TaskID]*api.TaskInfo{
				"task0": newPlacedNPUTask("task0", api.Bound, map[string]string{"other": "value"}),
				// the pod is absent, the annotation map of the pod is nil
				"task1": newPlacedNPUTask("task1", api.Bound, nil),
				"task2": barePodNPUTask("task2"),
			},
		},
		{
			name: "04 the waiting task marker is never used",
			tasks: map[api.TaskID]*api.TaskInfo{
				"task0": newPlacedNPUTask("task0", api.Pending, oldMarker)},
		},
		{
			name: "05 the newest marker wins",
			tasks: map[api.TaskID]*api.TaskInfo{
				"task0": newPlacedNPUTask("task0", api.Bound, oldMarker),
				"task1": newPlacedNPUTask("task1", api.Bound, newMarker)},
			wantConfig: configSpBlock1, wantTime: 2000,
		},
		{
			name: "06 the smaller config string breaks the tie of one effect time",
			tasks: map[api.TaskID]*api.TaskInfo{
				"task0": newPlacedNPUTask("task0", api.Bound, downgradeMarker("1000", configSpBlock2)),
				"task1": newPlacedNPUTask("task1", api.Bound, downgradeMarker("1000", configSpBlock1))},
			wantConfig: configSpBlock1, wantTime: 1000,
		},
	}
}

func TestReadDowngradeSeed(t *testing.T) {
	for _, tt := range buildReadDowngradeSeedTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			vcJob := &api.JobInfo{UID: api.JobID("seed-job"), Tasks: tt.tasks}
			// the traversal of the task map is random, the seed of one job must
			// settle on the same marker whichever way it is walked
			for i := 0; i < seedTraversalRounds; i++ {
				config, effectTime := readDowngradeSeed(vcJob)
				if config != tt.wantConfig || effectTime != tt.wantTime {
					t.Fatalf("downgrade seed = %q, %d, want %q, %d", config, effectTime,
						tt.wantConfig, tt.wantTime)
				}
			}
		})
	}
}

type allocateEffectiveMarkTest struct {
	name          string
	downgradeOn   bool
	wantEffective bool
}

func buildAllocateEffectiveMarkTestCases() []allocateEffectiveMarkTest {
	return []allocateEffectiveMarkTest{
		{name: "01 the allocation of a downgrade job freezes the level",
			downgradeOn: true, wantEffective: true},
		{name: "02 the allocation of a plain job touches no level"},
	}
}

// TestNPUAllocateFuncMarksLevelEffective covers the eager level freeze of the
// allocation: the mark happens at the pipeline moment and survives a rollback,
// the next session factory reset rebuilds the equivalent state from it.
func TestNPUAllocateFuncMarksLevelEffective(t *testing.T) {
	for _, tt := range buildAllocateEffectiveMarkTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			downgrade.ResetDowngradeCacheForTest()
			task := newPlacedNPUTask("effective-task", api.Pipelined, map[string]string{})
			task.Job = api.JobID("effective-job")
			task.NodeName = "effective-node"
			anno := map[string]string{}
			if tt.downgradeOn {
				anno[util.SchedulerDowngradeAnnoKey] = "true"
			}
			sHandle := &ScheduleHandler{ScheduleEnv: ScheduleEnv{ClusterCache: ClusterCache{
				Jobs: map[api.JobID]SchedulerJob{task.Job: {
					SchedulerJobAttr: util.SchedulerJobAttr{
						ComJob: util.ComJob{Annotation: anno}, NPUJob: &util.NPUJob{}},
					JobReadyTag:   util.PtrInit(true),
					policyHandler: New(testPluginName),
				}},
				Nodes: map[string]NPUNode{"effective-node": {}},
			}}}
			downgrade.RecordDowngrade(task.Job, configSpBlock2, 100)

			sHandle.NPUAllocateFunc(task)

			if downgrade.IsEffective(task.Job) != tt.wantEffective {
				t.Errorf("constraint effective = %v, want %v",
					downgrade.IsEffective(task.Job), tt.wantEffective)
			}
		})
	}
}

type deallocateRollbackMarkerTest struct {
	name       string
	status     api.TaskStatus
	wantMarker bool
}

func buildDeallocateRollbackMarkerTestCases() []deallocateRollbackMarkerTest {
	return []deallocateRollbackMarkerTest{
		{name: "01 a rolled back task drops its stale downgrade markers", status: api.Pending},
		{name: "02 an evicted task keeps its downgrade markers",
			status: api.Releasing, wantMarker: true},
	}
}

// TestNPUDeallocateFuncRollbackMarker covers the marker cleanup of the
// allocation rollback: the markers a pipelined pod carries would suppress the
// refresh of the next successful binding, so the rollback drops them.
func TestNPUDeallocateFuncRollbackMarker(t *testing.T) {
	for _, tt := range buildDeallocateRollbackMarkerTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			task := newPlacedNPUTask("marker-task", tt.status, downgradeMarker("1000", configSpBlock2))
			task.Job = api.JobID("marker-job")
			task.NodeName = "marker-node"
			sHandle := &ScheduleHandler{ScheduleEnv: ScheduleEnv{ClusterCache: ClusterCache{
				Jobs: map[api.JobID]SchedulerJob{task.Job: {
					SchedulerJobAttr: util.SchedulerJobAttr{
						ComJob: util.ComJob{Annotation: map[string]string{
							util.SchedulerDowngradeAnnoKey: "true"}},
						NPUJob: &util.NPUJob{}},
					policyHandler: New(testPluginName),
				}},
				Nodes: map[string]NPUNode{"marker-node": {}},
			}}}

			sHandle.NPUDeallocateFunc(task)

			_, hasMark := task.Pod.Annotations[util.SchedulerDowngradedAnnoKey]
			_, hasLevel := task.Pod.Annotations[util.SchedulerDowngradedLevelAnnoKey]
			if hasMark != tt.wantMarker || hasLevel != tt.wantMarker {
				t.Errorf("downgrade markers = %v/%v, want kept %v", hasMark, hasLevel, tt.wantMarker)
			}
		})
	}
}
