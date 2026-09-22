/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package plugin is using for HuaWei Ascend pin affinity schedule frame.
*/
package plugin

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/cache"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/consts"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type npuAllocateFuncArgs struct {
	task *api.TaskInfo
}

type npuAllocateFuncTest struct {
	name   string
	fields fields
	args   npuAllocateFuncArgs
	want   string
}

func buildNPUAllocateFuncTest() []npuAllocateFuncTest {
	task := test.FakeNormalTestTasks(1)[0]
	name, num := getVCTaskReqNPUTypeFromTaskInfo(task)
	tmpJobReadyTag := true
	npuTask := util.NPUTask{
		Name: task.Name, NameSpace: task.Namespace, ReqNPUName: name,
		ReqNPUNum: num,
		Label:     getTaskLabels(task), VTask: &util.VTask{}}
	// DRA-managed task: requests the NPU card but is guarded out by isTaskNeedNPUAllocated
	// before any annotation / node bookkeeping is touched.
	draTask := api.NewTaskInfo(test.BuildPodWithReqResource(util.NPU910CardName, "1"))
	draTask.Pod.Spec.ResourceClaims = []v1.PodResourceClaim{{Name: "claim-1"}}
	// Non-NPU task: only CPU requested, also rejected by the allocate guard.
	plainTask := api.NewTaskInfo(test.BuildPodWithReqResource(v1.ResourceCPU, "1"))
	tests := []npuAllocateFuncTest{
		{
			name:   "01-NPUAllocateFunc task nil test",
			fields: fields{},
			args:   npuAllocateFuncArgs{task: nil},
			want:   "",
		},
		{
			name: "05-NPUAllocateFunc DRA task test.",
			fields: fields{NPUPlugins: make(sets.String),
				ScheduleEnv: ScheduleEnv{ClusterCache: NewClusterCache()}},
			args: npuAllocateFuncArgs{task: draTask},
			want: "",
		},
		{
			name: "06-NPUAllocateFunc non NPU task test.",
			fields: fields{NPUPlugins: make(sets.String),
				ScheduleEnv: ScheduleEnv{ClusterCache: NewClusterCache()}},
			args: npuAllocateFuncArgs{task: plainTask},
			want: "",
		},
		{
			name: "02-NPUAllocateFunc no job test.",
			fields: fields{NPUPlugins: make(sets.String),
				ScheduleEnv: ScheduleEnv{
					ClusterCache: NewClusterCache(),
					FrameAttr:    VolcanoFrame{}}},
			args: npuAllocateFuncArgs{task: task},
			want: "",
		},
		{
			name: "03-NPUAllocateFunc no node test",
			fields: fields{NPUPlugins: make(sets.String),
				ScheduleEnv: ScheduleEnv{
					ClusterCache: ClusterCache{
						Jobs: map[api.JobID]SchedulerJob{task.Job: {
							JobReadyTag: &tmpJobReadyTag,
							SchedulerJobAttr: util.SchedulerJobAttr{
								NPUJob: &util.NPUJob{
									Tasks: map[api.TaskID]util.NPUTask{task.UID: npuTask},
								},
							}}},
						Nodes: map[string]NPUNode{}},
					FrameAttr: VolcanoFrame{}}},
			args: npuAllocateFuncArgs{task: task},
			want: "",
		},
		{
			name: "04-NPUAllocateFunc UseAnnotation failed test.",
			fields: fields{NPUPlugins: make(sets.String),
				ScheduleEnv: newDefaultsHandlerByFakeSsn().ScheduleEnv},
			args: npuAllocateFuncArgs{task: task},
			want: "",
		},
	}
	return tests
}

func TestNPUAllocateFunc(t *testing.T) {
	tests := buildNPUAllocateFuncTest()
	temp := func(task *api.TaskInfo) string {
		if task == nil {
			return ""
		}
		value, _ := task.Pod.Annotations[test.NPU910CardName]
		return value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			sHandle.NPUAllocateFunc(tt.args.task)
			value := temp(tt.args.task)
			if value != tt.want {
				t.Errorf("NPUAllocateFunc() got = %v, want %v", value, tt.want)
			}
		})
	}
}

// TestIsTaskNeedNPUAllocated covers the pure guard that decides whether a task
// needs the NPU allocate operation. It only inspects the task itself (nil / DRA /
// NPU resource request) and performs no side effects; the job and node context
// resolution happens later in NPUAllocateFunc.
func TestIsTaskNeedNPUAllocated(t *testing.T) {
	draTask := api.NewTaskInfo(test.BuildPodWithReqResource(util.NPU910CardName, "1"))
	draTask.Pod.Spec.ResourceClaims = []v1.PodResourceClaim{{Name: "claim-1"}}
	plainTask := api.NewTaskInfo(test.BuildPodWithReqResource(v1.ResourceCPU, "1"))
	npuTask := test.BuildTestTaskWithAnnotation(util.NPU910CardName, "1", "Ascend910-4")

	var sHandle ScheduleHandler
	tests := []struct {
		name string
		task *api.TaskInfo
		want bool
	}{
		{"01-nil task not allocated", nil, false},
		{"02-DRA task bypassed", draTask, false},
		{"03-non NPU task rejected", plainTask, false},
		{"04-NPU task admitted", npuTask, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sHandle.isTaskNeedNPUAllocated(tt.task); got != tt.want {
				t.Errorf("isTaskNeedNPUAllocated() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// recordPolicyHandler records which allocation hook runs. RestoreAnnotation
// mirrors the production policy base semantics — re-claiming the Running pod's
// chips on the node free-top from the pod's own annotation via the real
// NPUNode.GetNewNPUNodeAnnotation primitive, without touching the pod
// annotation. (The actual base.NPUHandler / chipHandler implementations live in
// internal/npu/base and internal/npu/affinity/chip; this type cannot import
// internal/npu/base because that package imports plugin.)
type recordPolicyHandler struct {
	ascendTest
	useCalled     bool
	restoreCalled bool
	backupCalled  bool
}

func (m *recordPolicyHandler) UseAnnotation(*api.TaskInfo, NPUNode) *NPUNode {
	m.useCalled = true
	return nil
}

// OnBackupPodAllocated implements plugin.BackupPodAllocatedHook: the production
// multilevel handler syncs job.SuperPods with the backup pod's node, which must
// also happen on the running-restore (unevict) path.
func (m *recordPolicyHandler) OnBackupPodAllocated(*api.TaskInfo, *SchedulerJob, string) {
	m.backupCalled = true
}

func (m *recordPolicyHandler) RestoreAnnotation(task *api.TaskInfo, node NPUNode) *NPUNode {
	m.restoreCalled = true
	if task == nil || task.Pod == nil {
		return nil
	}
	chipIDs := util.GetAllocatedChipIDsFromPod(task.Pod)
	if len(chipIDs) == 0 {
		return nil
	}
	newAnno, err := node.GetNewNPUNodeAnnotation(chipIDs, util.NPU910CardName, util.NPU910CardNamePre)
	if err != nil {
		return nil
	}
	node.Annotation[util.NPU910CardName] = newAnno
	return &node
}

// TestNPUAllocateFuncRunningRollback covers the only path where a Running task
// reaches NPUAllocateFunc: unevict restoring an evicted preempt/reclaim victim
// after the statement is discarded. The pod is still alive, so the allocation
// must not be re-selected (UseAnnotation skipped, being a no-op even if called)
// and the authoritative pod annotation stays untouched; the node free-top
// annotation — which the paired DeallocateFunc re-freed by prepending the pod's
// chips — is re-claimed from the pod's existing annotation, and chip.PodMap
// bookkeeping is restored.
func TestNPUAllocateFuncRunningRollback(t *testing.T) {
	rt := test.BuildTestTaskWithAnnotation(test.NPU910CardName, "1", "Ascend910-4")
	rt.TransactionContext.Status = api.Running
	rt.NodeName = "node0"
	rt.Pod.Annotations[PodRankIndexKey] = "0"
	tmpJobReadyTag := true
	policy := &recordPolicyHandler{}

	sHandle := ScheduleHandler{
		NPUPlugins: make(sets.String),
		ScheduleEnv: ScheduleEnv{
			ClusterCache: ClusterCache{
				Jobs: map[api.JobID]SchedulerJob{
					rt.Job: {
						JobReadyTag: &tmpJobReadyTag,
						SchedulerJobAttr: util.SchedulerJobAttr{
							NPUJob: &util.NPUJob{
								Tasks: map[api.TaskID]util.NPUTask{
									rt.UID: {ReqNPUName: test.NPU910CardName, ReqNPUNum: 1},
								},
								ReqNPUName: test.NPU910CardName,
								ReqNPUNum:  1,
							},
						},
						Owner:         OwnerInfo{OwnerReference: metav1.OwnerReference{UID: "owner-guid"}},
						policyHandler: policy,
					},
				},
				Nodes: map[string]NPUNode{
					rt.NodeName: {
						CommonNode: CommonNode{
							// DeallocateFunc prepended the evicted pod's chip 4
							// back onto the free top: "Ascend910-4,Ascend910-3".
							Annotation: map[string]string{test.NPU910CardName: "Ascend910-4,Ascend910-3"},
						},
						VNode: VNode{
							Chips: map[int]*VChip{
								4: {PodMap: make(map[string]*v1.Pod)},
							},
						},
					},
				},
			},
		},
		AffinityCache: cache.NewPodNodeAffinityCache(),
	}

	sHandle.NPUAllocateFunc(rt)

	if got := rt.Pod.Annotations[test.NPU910CardName]; got != "Ascend910-4" {
		t.Errorf("Running rollback: pod annotation got %q, want %q", got, "Ascend910-4")
	}
	if policy.useCalled {
		t.Error("Running rollback: UseAnnotation must not be called")
	}
	if !policy.restoreCalled {
		t.Error("Running rollback: RestoreAnnotation must be called")
	}
	node := sHandle.Nodes[rt.NodeName]
	if got := node.Annotation[test.NPU910CardName]; got != "Ascend910-3" {
		t.Errorf("Running rollback: node free-top got %q, want %q (chip 4 re-claimed)", got, "Ascend910-3")
	}
	if _, ok := node.Chips[4].PodMap[string(rt.Pod.UID)]; !ok {
		t.Errorf("Running rollback: chip 4 PodMap does not contain pod %s", rt.Pod.UID)
	}
	// The restored pod is landing back on node0, so the "prefer previous node"
	// cache must be re-affirmed on the running-restore path (RecordAssignment
	// also refreshes the entry TTL).
	if got := sHandle.AffinityCache.GetPreferredNode("owner-guid", "0"); got != rt.NodeName {
		t.Errorf("Running rollback: affinity cache got %q, want %q (RecordAssignment ran)", got, rt.NodeName)
	}
}

// TestNPUAllocateFuncRunningRollbackBackup verifies that a hot-switch backup pod
// restored through the running path still fires BackupPodAllocatedHook, keeping
// SuperPods in sync — while still skipping UseAnnotation / pod annotation
// rewrites.
func TestNPUAllocateFuncRunningRollbackBackup(t *testing.T) {
	rt := test.BuildTestTaskWithAnnotation(test.NPU910CardName, "1", "Ascend910-4")
	rt.TransactionContext.Status = api.Running
	rt.NodeName = "node0"
	rt.Pod.Annotations[consts.BackupSourcePodNameKey] = "fault-pod-uid"
	tmpJobReadyTag := true
	policy := &recordPolicyHandler{}

	sHandle := ScheduleHandler{
		NPUPlugins: make(sets.String),
		ScheduleEnv: ScheduleEnv{
			ClusterCache: ClusterCache{
				Jobs: map[api.JobID]SchedulerJob{
					rt.Job: {
						JobReadyTag: &tmpJobReadyTag,
						SchedulerJobAttr: util.SchedulerJobAttr{
							NPUJob: &util.NPUJob{
								Tasks: map[api.TaskID]util.NPUTask{
									rt.UID: {ReqNPUName: test.NPU910CardName, ReqNPUNum: 1},
								},
								ReqNPUName: test.NPU910CardName,
								ReqNPUNum:  1,
							},
						},
						policyHandler: policy,
					},
				},
				Nodes: map[string]NPUNode{
					rt.NodeName: {
						CommonNode: CommonNode{
							Annotation: map[string]string{test.NPU910CardName: "Ascend910-4,Ascend910-3"},
						},
						VNode: VNode{
							Chips: map[int]*VChip{
								4: {PodMap: make(map[string]*v1.Pod)},
							},
						},
					},
				},
			},
		},
	}

	sHandle.NPUAllocateFunc(rt)

	if policy.useCalled {
		t.Error("Running backup restore: UseAnnotation must not be called")
	}
	if !policy.restoreCalled {
		t.Error("Running backup restore: RestoreAnnotation must be called")
	}
	if !policy.backupCalled {
		t.Error("Running backup restore: OnBackupPodAllocated must be called for a backup pod")
	}
	if got := rt.Pod.Annotations[test.NPU910CardName]; got != "Ascend910-4" {
		t.Errorf("Running backup restore: pod annotation got %q, want %q", got, "Ascend910-4")
	}
}

type npuDeallocateFuncArgs struct {
	task *api.TaskInfo
}

type npuDeallocateFuncTest struct {
	name   string
	fields fields
	args   npuDeallocateFuncArgs
	want   string
}

func makeNPUDeallocateFuncTest01(_ *api.TaskInfo) npuDeallocateFuncTest {
	return npuDeallocateFuncTest{
		name:   "01-NPUDeallocateFunc task nil test",
		fields: fields{}, args: npuDeallocateFuncArgs{task: nil}, want: "",
	}
}

func makeNPUDeallocateFuncTest02(vTask *api.TaskInfo) npuDeallocateFuncTest {
	return npuDeallocateFuncTest{
		name: "02-NPUAllocateFunc no job test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{ClusterCache: NewClusterCache()}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest03(vTask *api.TaskInfo) npuDeallocateFuncTest {
	return npuDeallocateFuncTest{
		name: "03-NPUAllocateFunc no node test",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs:  map[api.JobID]SchedulerJob{vTask.Job: {}},
					Nodes: map[string]NPUNode{}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest04(vTask *api.TaskInfo) npuDeallocateFuncTest {
	return npuDeallocateFuncTest{
		name: "04-NPUAllocateFunc UseAnnotation failed test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{
						vTask.Job: {SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: nil}}}},
					Nodes: map[string]NPUNode{vTask.NodeName: {}}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest05(vTask *api.TaskInfo) npuDeallocateFuncTest {
	return npuDeallocateFuncTest{
		name: "05-NPUAllocateFunc pod no req test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{
						vTask.Job: {
							SchedulerJobAttr: util.SchedulerJobAttr{
								NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{vTask.UID: {ReqNPUName: "haha"}}}},
						},
					},
					Nodes: map[string]NPUNode{vTask.NodeName: {}}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest06(vTask *api.TaskInfo) npuDeallocateFuncTest {
	tmpSchedulerJobAttr := util.SchedulerJobAttr{
		NPUJob: &util.NPUJob{
			Tasks: map[api.TaskID]util.NPUTask{
				vTask.UID: {ReqNPUName: test.NPU910CardName, ReqNPUNum: util.NPUIndex2}},
		},
	}

	return npuDeallocateFuncTest{
		name: "06-NPUAllocateFunc pod req num not meet test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{
						vTask.Job: {
							SchedulerJobAttr: tmpSchedulerJobAttr,
						},
					},
					Nodes: map[string]NPUNode{vTask.NodeName: {}}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest07(vTask *api.TaskInfo) npuDeallocateFuncTest {
	tmpSchedulerJobAttr := util.SchedulerJobAttr{
		NPUJob: &util.NPUJob{
			Tasks: map[api.TaskID]util.NPUTask{
				vTask.UID: {ReqNPUName: test.NPU910CardName, ReqNPUNum: 1}},
		},
	}
	tmpNPUNode := NPUNode{
		CommonNode: CommonNode{Annotation: nil},
	}
	return npuDeallocateFuncTest{
		name: "07-NPUAllocateFunc node no annotation value test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{
						vTask.Job: {SchedulerJobAttr: tmpSchedulerJobAttr},
					},
					Nodes: map[string]NPUNode{vTask.NodeName: tmpNPUNode}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest08(vTask *api.TaskInfo) npuDeallocateFuncTest {
	tmpSchedulerJobAttr := util.SchedulerJobAttr{
		NPUJob: &util.NPUJob{
			Tasks: map[api.TaskID]util.NPUTask{
				vTask.UID: {ReqNPUName: test.NPU910CardName, ReqNPUNum: 1,
					VTask: &util.VTask{}}},
		},
	}
	tmpNPUNode := NPUNode{
		CommonNode: CommonNode{
			Annotation: map[string]string{test.NPU910CardName: ""},
		},
	}
	// The task built by BuildTestTaskWithAnnotation has a non-Pending status
	// (Unknown), so after deallocation the pod annotations are retained to let
	// a released pod be rescheduled back to reuse its chips.
	return npuDeallocateFuncTest{
		name: "08-NPUAllocateFunc node has empty annotation value test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{vTask.Job: {SchedulerJobAttr: tmpSchedulerJobAttr,
						policyHandler: New(testPluginName)}},
					Nodes: map[string]NPUNode{vTask.NodeName: tmpNPUNode}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest09(vTask *api.TaskInfo) npuDeallocateFuncTest {
	tmpSchedulerJobAttr := util.SchedulerJobAttr{
		NPUJob: &util.NPUJob{
			Tasks: map[api.TaskID]util.NPUTask{
				vTask.UID: {ReqNPUName: test.NPU910CardName, ReqNPUNum: 1}},
		},
	}
	tmpNPUNode := NPUNode{
		CommonNode: CommonNode{
			Annotation: map[string]string{test.NPU910CardName: "Ascend910-3"},
		},
	}
	return npuDeallocateFuncTest{
		name: "09-NPUAllocateFunc ok test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{vTask.Job: {SchedulerJobAttr: tmpSchedulerJobAttr,
						policyHandler: New(testPluginName)}},
					Nodes: map[string]NPUNode{vTask.NodeName: tmpNPUNode}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest10(_ *api.TaskInfo) npuDeallocateFuncTest {
	// A Pending task is an allocation rollback / unpipeline: the pod never ran
	// on this node, so releaseAnnotation clears the pod annotations.
	pTask := test.BuildTestTaskWithAnnotation(test.NPU910CardName, "1", "Ascend910-4")
	pTask.TransactionContext.Status = api.Pending
	tmpSchedulerJobAttr := util.SchedulerJobAttr{
		NPUJob: &util.NPUJob{
			Tasks: map[api.TaskID]util.NPUTask{
				pTask.UID: {ReqNPUName: test.NPU910CardName, ReqNPUNum: 1}},
		},
	}
	tmpNPUNode := NPUNode{
		CommonNode: CommonNode{
			Annotation: map[string]string{test.NPU910CardName: "Ascend910-3"},
		},
	}
	return npuDeallocateFuncTest{
		name: "10-NPUDeallocateFunc pending task rollback clears annotations test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{pTask.Job: {SchedulerJobAttr: tmpSchedulerJobAttr,
						policyHandler: New(testPluginName)}},
					Nodes: map[string]NPUNode{pTask.NodeName: tmpNPUNode}}}},
		args: npuDeallocateFuncArgs{task: pTask}, want: "",
	}
}

func buildNPUDeallocateFuncTest() []npuDeallocateFuncTest {
	vTask := test.BuildTestTaskWithAnnotation(test.NPU910CardName, "1", "Ascend910-4")
	tests := []npuDeallocateFuncTest{
		makeNPUDeallocateFuncTest01(vTask),
		makeNPUDeallocateFuncTest02(vTask),
		makeNPUDeallocateFuncTest03(vTask),
		makeNPUDeallocateFuncTest04(vTask),
		makeNPUDeallocateFuncTest05(vTask),
		makeNPUDeallocateFuncTest06(vTask),
		makeNPUDeallocateFuncTest07(vTask),
		makeNPUDeallocateFuncTest08(vTask),
		makeNPUDeallocateFuncTest09(vTask),
		makeNPUDeallocateFuncTest10(vTask),
	}
	return tests
}

// TestGetAllocatedChipIDsFromPod covers nil pod, nil annotations and
// multiple card-name annotations (910 / 310P / npu).
func TestGetAllocatedChipIDsFromPod(t *testing.T) {
	tests := []struct {
		name      string
		podName   string
		npuName   string
		annoVal   string
		wantLen   int
		wantFirst int
	}{
		{"01-nil pod handled by caller", "", "", "", 0, 0},
		{"02-910 single", "p1", util.NPU910CardName, "Ascend910-3", 1, 3},
		{"03-910 multi", "p2", util.NPU910CardName, "Ascend910-0,Ascend910-3", 2, 0},
		{"04-310P", "p3", util.NPU310PCardName, "Ascend310P-1,Ascend310P-2", 2, 1},
		{"05-npu", "p4", util.NPUCardName, "npu-1,npu-2", 2, 1},
		{"06-empty annotation", "p5", util.NPU910CardName, "", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.podName == "" {
				if got := util.GetAllocatedChipIDsFromPod(nil); len(got) != 0 {
					t.Errorf("got=%v want empty", got)
				}
				return
			}
			task := test.BuildTestTaskWithAnnotation(tt.npuName, "1", tt.annoVal)
			got := util.GetAllocatedChipIDsFromPod(task.Pod)
			if len(got) != tt.wantLen {
				t.Errorf("got=%v wantLen=%d", got, tt.wantLen)
				return
			}
			if tt.wantLen > 0 && got[0] != tt.wantFirst {
				t.Errorf("got[0]=%d want=%d", got[0], tt.wantFirst)
			}
		})
	}
}

// TestCalcCardFreeCount covers invalid args and whole-card scenarios:
// Chips map empty, preemptee chip counted, unhealthy excluded,
// non-preemptee occupied excluded, multi-card bucketing,
// and non-contiguous available chip IDs (only chips in availableChipIDs counted).
func TestCalcCardFreeCount(t *testing.T) {
	pe := test.BuildTestTaskWithAnnotation(util.NPU910CardName, "1", "Ascend910-0")
	pe.Pod.UID = "pe-uid" // override default UID "-" to distinguish from other pods
	other := test.BuildTestTaskWithAnnotation(util.NPU910CardName, "1", "Ascend910-2")
	other.Pod.UID = "other-uid"
	tests := []struct {
		name     string
		node     *NPUNode
		maxNum   int
		availIDs []int
		wantCard int
		wantCnt  int
		wantFull map[int]int // full map expectation, optional; if nil only wantCard/wantCnt checked
	}{
		{"01-nil node", nil, 0, nil, 0, 0, nil},
		{"02-preemptee+unhealthy+idle", &NPUNode{
			CommonNode: CommonNode{Name: "n1"},
			VNode:      VNode{UnhealthyChipIds: map[int]struct{}{1: {}}}}, 4, []int{0, 1, 2, 3}, 0, 3, nil},
		{"03-with non-preemptee occupied", &NPUNode{
			CommonNode: CommonNode{Name: "n2", Tasks: map[api.TaskID]*api.TaskInfo{api.TaskID("other-uid"): other}},
			VNode:      VNode{}}, 4, []int{0, 1, 2, 3}, 0, 3, nil},
		{"04-multi-card bucketing", &NPUNode{
			CommonNode: CommonNode{Name: "n3"},
			VNode:      VNode{UnhealthyChipIds: map[int]struct{}{5: {}}}}, 4, []int{0, 1, 2, 3, 4, 5, 6, 7}, 1, 3, nil},
		// Simulates real-world scenario: 8 cards in 2 meshes, but only chips 1,2,3,7 are
		// reported as available (0,4,5,6 are unschedulable and absent from node annotation).
		// preemptee pe holds chip 0, which is NOT in availableChipIDs, so it is ignored.
		// Result: mesh 0 has 3 free chips (1,2,3), mesh 1 has 1 free chip (7).
		{"05-non-contiguous available IDs", &NPUNode{
			CommonNode: CommonNode{Name: "n4"},
			VNode:      VNode{}}, 4, []int{1, 2, 3, 7}, 0, 3, map[int]int{0: 3, 1: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcCardFreeCount(tt.node, []*api.TaskInfo{pe}, tt.maxNum, tt.availIDs)
			if tt.node == nil {
				if got != nil {
					t.Errorf("nil node got=%v want nil", got)
				}
				return
			}
			if tt.wantFull != nil {
				if len(got) != len(tt.wantFull) {
					t.Errorf("got=%v want=%v", got, tt.wantFull)
					return
				}
				for k, v := range tt.wantFull {
					if got[k] != v {
						t.Errorf("card[%d]=%d want=%d, full=%v", k, got[k], v, got)
					}
				}
				return
			}
			if got[tt.wantCard] != tt.wantCnt {
				t.Errorf("card[%d]=%d want=%d, full=%v", tt.wantCard, got[tt.wantCard], tt.wantCnt, got)
			}
		})
	}
}

func TestNPUDeallocateFunc(t *testing.T) {
	tests := buildNPUDeallocateFuncTest()
	temp := func(task *api.TaskInfo) string {
		if task == nil {
			return ""
		}
		value, _ := task.Pod.Annotations[test.NPU910CardName]
		return value
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			sHandle.NPUDeallocateFunc(tt.args.task)
			value := temp(tt.args.task)
			if value != tt.want {
				t.Errorf("NPUDeallocateFunc() got = %v, want %v", value, tt.want)
			}
		})
	}
}
