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
	"errors"
	"math"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// stubScorePlugin fixed-contribution dimension used by synthesizer tests
type stubScorePlugin struct {
	name  string
	delta map[string]int
}

func (p *stubScorePlugin) Name() string { return p.name }

func (p *stubScorePlugin) Score(task *api.TaskInfo, nodes []*api.NodeInfo, vcJob SchedulerJob) (map[string]int, error) {
	return p.delta, nil
}

// TestBatchNodeOrderByFramework
func TestBatchNodeOrderByFramework(t *testing.T) {
	h := &ScheduleHandler{
		ScoreWeight: 10,
		scorePlugins: []ScorePluginWithWeight{
			{plugin: &stubScorePlugin{name: "topology", delta: map[string]int{"node1": 1, "node2": 0}}, weight: util.ScoreTopoShift}, // segment {1,0} shifted by this dimension's weight
			{plugin: &stubScorePlugin{name: "chipCount", delta: map[string]int{"node1": 255, "node2": 128}}, weight: util.ScoreAvailShift},
		},
	}
	vcJob := SchedulerJob{policyHandler: &mockPolicyHandler{}}
	task := &api.TaskInfo{Name: "t1"}
	score, err := h.batchNodeOrderByFramework(task,
		[]*api.NodeInfo{{Name: "node1"}, {Name: "node2"}}, vcJob)
	if err != nil {
		t.Fatalf("batchNodeOrderByFramework() error = %v", err)
	}
	want1 := float64((1<<util.ScoreTopoShift)|255) * 10 // node1: topo(0x400)+avail full (255) → 12790
	want2 := float64(128) * 10                          // node2: avail only → 1280
	if score["node1"] != want1 || score["node2"] != want2 {
		t.Errorf("scores = %v, want node1=%v node2=%v", score, want1, want2)
	}
}

// TestBatchNodeOrderByFrameworkErrWithRegistry
func TestBatchNodeOrderByFrameworkErrWithRegistry(t *testing.T) {
	h := &ScheduleHandler{ScoreWeight: 10}
	h.initScorePlugins()
	vcJob := SchedulerJob{policyHandler: &errorPolicyHandler{}}
	score, err := h.batchNodeOrderByFramework(&api.TaskInfo{Name: "t1"},
		[]*api.NodeInfo{{Name: "node1"}}, vcJob)
	if err == nil {
		t.Fatal("batchNodeOrderByFramework() err = nil, want err (topology plugin error)")
	}
	if score != nil {
		t.Errorf("batchNodeOrderByFramework() score = %v, want nil (plugin error aborts synthesis)", score)
	}
}

// errorPolicyHandler top-level scoring returns an error (ScoreBestNPUNodes in-place interface)
type errorPolicyHandler struct{ mockPolicyHandler }

func (m *errorPolicyHandler) ScoreBestNPUNodes(_ *api.TaskInfo, _ []*api.NodeInfo, _ map[string]float64) error {
	return errors.New(util.ArgumentError)
}

// errorStubScorePlugin
type errorStubScorePlugin struct{ stubScorePlugin }

func (p *errorStubScorePlugin) Score(task *api.TaskInfo, nodes []*api.NodeInfo, vcJob SchedulerJob) (map[string]int, error) {
	return nil, errors.New(util.ArgumentError)
}

// TestBatchNodeOrderByFrameworkPluginErrAbort
func TestBatchNodeOrderByFrameworkPluginErrAbort(t *testing.T) {
	h := &ScheduleHandler{
		ScoreWeight: 10,
		scorePlugins: []ScorePluginWithWeight{
			{plugin: &stubScorePlugin{name: "topology", delta: map[string]int{"node1": 1}}, weight: util.ScoreTopoShift},
			{plugin: &errorStubScorePlugin{stubScorePlugin: stubScorePlugin{name: "chipCount"}}, weight: util.ScoreAvailShift},
		},
	}
	vcJob := SchedulerJob{policyHandler: &mockPolicyHandler{}}
	score, err := h.batchNodeOrderByFramework(&api.TaskInfo{Name: "t1"},
		[]*api.NodeInfo{{Name: "node1"}}, vcJob)
	if err == nil {
		t.Fatal("batchNodeOrderByFramework() err = nil, want err")
	}
	if score != nil {
		t.Errorf("batchNodeOrderByFramework() score = %v, want nil (plugin err aborts synthesis)", score)
	}
}

// TestScoreTopology
func TestScoreTopology(t *testing.T) {
	h := &ScheduleHandler{}
	nodes := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}}
	// mock top level writes no topo → all-0 frame (all nodes keyed)
	got, err := h.scoreTopology(&api.TaskInfo{Name: "t1"}, nodes,
		SchedulerJob{policyHandler: &mockPolicyHandler{}})
	if err != nil {
		t.Fatalf("scoreTopology() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("scoreTopology() len = %d, want 2 (all-node frame)", len(got))
	}
	for _, v := range got {
		if v != 0 {
			t.Errorf("scoreTopology() (mock writes nothing) = %v, want all 0", got)
		}
	}
	// error top level → plugin-error semantics (nil, err): wired to the synthesizer it aborts
	// (score=nil), leaving no half-composite
	if got2, err := h.scoreTopology(&api.TaskInfo{Name: "t2"}, nodes,
		SchedulerJob{policyHandler: &errorPolicyHandler{}}); err == nil || got2 != nil {
		t.Fatalf("scoreTopology() (error policy) err=%v got=%v, want (nil,err)", err, got2)
	}
}

// TestScorePreviousNodeDisabled
func TestScorePreviousNodeDisabled(t *testing.T) {
	h := &ScheduleHandler{}
	nodeInfos := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}}
	got, err := h.scorePreviousNode(&api.TaskInfo{Name: "t1"}, nodeInfos,
		SchedulerJob{policyHandler: &mockPolicyHandler{}})
	if err != nil {
		t.Fatalf("scorePreviousNode() error = %v", err)
	}
	for _, n := range nodeInfos {
		if got[n.Name] != 1 {
			t.Errorf("scorePreviousNode() disabled: [%s] = %v, want 1 (avoid-fault base value, no deduction)",
				n.Name, got[n.Name])
		}
	}
}

// faultLandingHandler marks the given nodes as previous fault-task landings via
// ScorePreviousFaultNodes (fault snapshot source, no PrefNodeMap involved).
type faultLandingHandler struct {
	mockFaultHandler
	nodes map[string]struct{}
}

func (m *faultLandingHandler) ScorePreviousFaultNodes(_ *api.TaskInfo, scoreMap map[string]float64) {
	for node := range m.nodes {
		if _, ok := scoreMap[node]; ok {
			scoreMap[node] = 0
		}
	}
}

// TestScorePreviousNodeFaultNodeBinary: rescheduler-only (prefer-previous-node
// off) — previous fault-task landings get 0, normal candidates keep 1. The
// landing set is sourced from the fault snapshot (ScorePreviousFaultNodes),
// so it works with an empty PrefNodeMap.
func TestScorePreviousNodeFaultNodeBinary(t *testing.T) {
	h := &ScheduleHandler{
		FaultHandle: &faultLandingHandler{nodes: map[string]struct{}{"node1": {}}},
	}
	nodeInfos := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}, {Name: "node3"}}
	vcJob := SchedulerJob{} // no PrefNodeMap: the fault snapshot is the only source
	task := &api.TaskInfo{Name: "t1", Job: "job1"}
	got, err := h.scorePreviousNode(task, nodeInfos, vcJob)
	if err != nil {
		t.Fatalf("scorePreviousNode() error = %v", err)
	}
	want := map[string]int{"node1": 0, "node2": 1, "node3": 1}
	for _, n := range nodeInfos {
		if got[n.Name] != want[n.Name] {
			t.Errorf("scorePreviousNode() [%s] = %v, want %v (previous-fault landing → 0 / normal → 1)",
				n.Name, got[n.Name], want[n.Name])
		}
	}
}

// TestScorePreviousNodeFaultNodeOutsideCandidates: a fault landing that is not
// in the candidate list must not appear in the result (member check only).
func TestScorePreviousNodeFaultNodeOutsideCandidates(t *testing.T) {
	h := &ScheduleHandler{
		FaultHandle: &faultLandingHandler{nodes: map[string]struct{}{"node9": {}}},
	}
	nodeInfos := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}}
	got, err := h.scorePreviousNode(&api.TaskInfo{Name: "t1", Job: "job1"}, nodeInfos, SchedulerJob{})
	if err != nil {
		t.Fatalf("scorePreviousNode() error = %v", err)
	}
	for _, n := range nodeInfos {
		if got[n.Name] != 1 {
			t.Errorf("scorePreviousNode() [%s] = %v, want 1 (non-candidate fault landing ignored)", n.Name, got[n.Name])
		}
	}
}

func TestScorePreviousNodeBoost(t *testing.T) {
	h := &ScheduleHandler{
		ScheduleEnv: ScheduleEnv{
			FrameAttr: VolcanoFrame{ConfigParameters: ConfigParameters{
				DynamicParameters: DynamicParameters{PreferPreviousNode: true},
			}},
		},
	}
	nodeInfos := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}}
	vcJob := SchedulerJob{
		Owner: OwnerInfo{OwnerReference: metav1.OwnerReference{UID: "owner-uid"}},
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{Label: map[string]string{util.TorAffinityKey: util.NullTag}},
			NPUJob: &util.NPUJob{NPUTaskNum: 2, Tasks: fakeTasksForRank()},
		},
		PrefNodeMap:   map[int]string{0: "node1"},
		policyHandler: &mockPolicyHandler{},
	}
	task := &api.TaskInfo{
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "0"}}},
		Job: "job1",
	}
	got, err := h.scorePreviousNode(task, nodeInfos, vcJob)
	if err != nil {
		t.Fatalf("scorePreviousNode() error = %v", err)
	}
	want := 2 // P1 (selfNode∈candidates → tier 2 → synthesizer sets bit12)
	if got["node1"] != want {
		t.Errorf("scorePreviousNode() boost = %v, want node1=%v", got, want)
	}
	if got["node2"] != 1 {
		t.Errorf("scorePreviousNode() = %v, want node2=1 (otherNodes → P2, category tier)",
			got)
	}
}

type truncatingFaultHandler struct{ mockFaultHandler }

func (m *truncatingFaultHandler) ScoreBestNPUNodes(_ *api.TaskInfo, scoreMap map[string]float64) {
	m.truncate(scoreMap)
}

func (m *truncatingFaultHandler) ScoreLastFaultNode(_ *api.TaskInfo, scoreMap map[string]float64) {
	m.truncate(scoreMap)
}

func (m *truncatingFaultHandler) ScoreSubHealthGrade(scoreMap map[string]float64) {
	for nodeName := range scoreMap {
		scoreMap[nodeName] = 0 // any sub-healthy node → 0, healthy base 1 fully lost
	}
}

func (m *truncatingFaultHandler) truncate(scoreMap map[string]float64) {
	for nodeName, score := range scoreMap {
		scoreMap[nodeName] = math.Max(0, score-5)
	}
}

type gradedFaultHandler struct {
	mockFaultHandler
	grades map[string]float64
}

func (m *gradedFaultHandler) ScoreSubHealthGrade(scoreMap map[string]float64) {
	for nodeName, grade := range m.grades {
		if _, ok := scoreMap[nodeName]; ok {
			scoreMap[nodeName] = grade
		}
	}
}

func TestScoreSubHealth(t *testing.T) {
	h := &ScheduleHandler{FaultHandle: &mockFaultHandler{}}
	nodeInfos := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}}
	got, err := h.scoreSubHealth(&api.TaskInfo{Name: "t1"}, nodeInfos, SchedulerJob{})
	if err != nil {
		t.Fatalf("scoreSubHealth() error = %v", err)
	}
	for _, n := range nodeInfos {
		if got[n.Name] != 1 {
			t.Errorf("scoreSubHealth() [%s] = %v, want 1 (healthy → segment value 1)", n.Name, got[n.Name])
		}
	}
}

// TestScoreSubHealthNilHandle FaultHandle nil: predate defaults everything to 1 (healthy).
func TestScoreSubHealthNilHandle(t *testing.T) {
	h := &ScheduleHandler{}
	nodeInfos := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}}
	got, err := h.scoreSubHealth(&api.TaskInfo{Name: "t1"}, nodeInfos, SchedulerJob{})
	if err != nil {
		t.Fatalf("scoreSubHealth() error = %v", err)
	}
	for _, n := range nodeInfos {
		if got[n.Name] != 1 {
			t.Errorf("scoreSubHealth() nil handle [%s] = %v, want 1 (healthy → segment value 1)", n.Name, got[n.Name])
		}
	}
}

func TestScoreSubHealthGraded(t *testing.T) {
	h := &ScheduleHandler{FaultHandle: &truncatingFaultHandler{}}
	nodeInfos := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}}
	got, err := h.scoreSubHealth(&api.TaskInfo{Name: "t1"}, nodeInfos, SchedulerJob{})
	if err != nil {
		t.Fatalf("scoreSubHealth() error = %v", err)
	}
	if got["node1"] != 0 || got["node2"] != 0 {
		t.Errorf("scoreSubHealth() = %v, want node1=0 node2=0 (sub-health → segment value 0)", got)
	}
}

func TestScoreSubHealthGradePassthrough(t *testing.T) {
	h := &ScheduleHandler{FaultHandle: &gradedFaultHandler{
		grades: map[string]float64{"node1": 0, "node2": 1},
	}}
	nodeInfos := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}, {Name: "node3"}}
	got, err := h.scoreSubHealth(&api.TaskInfo{Name: "t1"}, nodeInfos, SchedulerJob{})
	if err != nil {
		t.Fatalf("scoreSubHealth() error = %v", err)
	}
	want := map[string]int{"node1": 0, "node2": 1, "node3": 1}
	for name, w := range want {
		if got[name] != w {
			t.Errorf("scoreSubHealth() [%s] = %v, want %v (graded pass-through)", name, got[name], w)
		}
	}
}

func TestScorePreviousNodePolicyIndependent(t *testing.T) {
	nodeInfos := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}}
	mk := func(policy SchedulerPluginNeed) (*ScheduleHandler, SchedulerJob) {
		h := &ScheduleHandler{
			ScheduleEnv: ScheduleEnv{FrameAttr: VolcanoFrame{ConfigParameters: ConfigParameters{
				DynamicParameters: DynamicParameters{PreferPreviousNode: true},
			}}},
		}
		vcJob := SchedulerJob{
			Owner: OwnerInfo{OwnerReference: metav1.OwnerReference{UID: "owner-uid"}},
			SchedulerJobAttr: util.SchedulerJobAttr{
				ComJob: util.ComJob{},
				NPUJob: &util.NPUJob{NPUTaskNum: 2, Tasks: fakeTasksForRank()},
			},
			PrefNodeMap:   map[int]string{0: "node1"},
			policyHandler: policy,
		}
		return h, vcJob
	}
	task := &api.TaskInfo{
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "0"}}},
		Job: "job1",
	}
	mockH, mockJob := mk(&mockPolicyHandler{})
	mockGot, err := mockH.scorePreviousNode(task, nodeInfos, mockJob)
	if err != nil {
		t.Fatalf("scorePreviousNode() (mock) error = %v", err)
	}
	errH, errJob := mk(&errorPolicyHandler{})
	errGot, err2 := errH.scorePreviousNode(task, nodeInfos, errJob)
	if err2 != nil {
		t.Fatalf("scorePreviousNode() (errorPolicy) error = %v (dimension must not touch top-level scoring)", err2)
	}
	if len(mockGot) != len(errGot) {
		t.Errorf("scorePreviousNode() key set mismatch: mock(len=%d) errorPolicy(len=%d)",
			len(mockGot), len(errGot))
	}
	for name, v := range mockGot {
		if errGot[name] != v {
			t.Errorf("scorePreviousNode() [%s]: mock=%v errorPolicy=%v, want equal (dimension independent)",
				name, v, errGot[name])
		}
	}
	if mockGot["node1"] != 2 {
		t.Errorf("scorePreviousNode() node1 = %v, want 2 (P1, selfNode∈candidates)", mockGot["node1"])
	}
}

func TestScoreSubHealthPolicyIndependent(t *testing.T) {
	nodeInfos := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}}
	run := func(policy SchedulerPluginNeed) map[string]int {
		h := &ScheduleHandler{FaultHandle: &truncatingFaultHandler{}}
		got, err := h.scoreSubHealth(&api.TaskInfo{Name: "t1"}, nodeInfos,
			SchedulerJob{policyHandler: policy})
		if err != nil {
			t.Fatalf("scoreSubHealth() error = %v", err)
		}
		return got
	}
	mock := run(&mockPolicyHandler{})
	topo := run(&topoWritingPolicy{})
	if len(mock) != len(topo) {
		t.Errorf("scoreSubHealth() key set mismatch: mock(len=%d) topoWriting(len=%d)",
			len(mock), len(topo))
	}
	for _, n := range nodeInfos {
		if mock[n.Name] != topo[n.Name] {
			t.Errorf("scoreSubHealth() [%s]: mock=%v topoWriting=%v, want equal (dimension independent)",
				n.Name, mock[n.Name], topo[n.Name])
		}
		if mock[n.Name] != 0 {
			t.Errorf("scoreSubHealth() [%s] = %v, want 0 (any sub-health → segment value 0)", n.Name, mock[n.Name])
		}
	}
}

func TestScoreChipCount(t *testing.T) {
	cases := []struct {
		name      string
		req       int
		freeChips map[string]int
		want      map[string]int
	}{
		{
			name:      "exact-fit-full",
			req:       4,
			freeChips: map[string]int{"node1": 4}, // 4/4×255=255 full
			want:      map[string]int{"node1": 255},
		},
		{
			name:      "tighter-fit-higher",
			req:       4,
			freeChips: map[string]int{"node1": 4, "node2": 8}, // 4/4×255=255, 4/8×255=127.5≈128
			want:      map[string]int{"node1": 255, "node2": 128},
		},
		{
			name:      "partial-fit",
			req:       2,
			freeChips: map[string]int{"node1": 4, "node2": 8}, // 2/4×255=127.5≈128, 2/8×255=63.75≈64
			want:      map[string]int{"node1": 128, "node2": 64},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nodes := make(map[string]NPUNode, len(tc.freeChips))
			nodeInfos := make([]*api.NodeInfo, 0, len(tc.freeChips))
			for name, free := range tc.freeChips {
				nodes[name] = NPUNode{CommonNode: CommonNode{
					Allocate: map[v1.ResourceName]float64{util.NPU910CardName: 8 * util.NPUHexKilo},
					Idle:     map[v1.ResourceName]float64{util.NPU910CardName: float64(free) * util.NPUHexKilo},
				}}
				nodeInfos = append(nodeInfos, &api.NodeInfo{Name: name})
			}
			h := &ScheduleHandler{ScheduleEnv: ScheduleEnv{ClusterCache: ClusterCache{Nodes: nodes}}}
			vcJob := SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{
				Tasks: map[api.TaskID]util.NPUTask{
					"t1": {ReqNPUNum: tc.req, ReqNPUName: util.NPU910CardName},
				},
			}}}
			got, err := h.scoreChipCount(&api.TaskInfo{Name: "t1", UID: "t1"}, nodeInfos, vcJob)
			if err != nil {
				t.Fatalf("scoreChipCount() error = %v", err)
			}
			for name, want := range tc.want {
				if got[name] != want {
					t.Errorf("scoreChipCount() %s = %v, want %v (got=%v)", name, got[name], want, got)
				}
			}
		})
	}
	t.Run("guards-zero", func(t *testing.T) {
		nodeInfos := []*api.NodeInfo{{Name: "node1"}}
		// NPUJob nil → all 0
		got, err := (&ScheduleHandler{}).scoreChipCount(
			&api.TaskInfo{Name: "t1", UID: "t1"}, nodeInfos, SchedulerJob{})
		if err != nil {
			t.Fatalf("scoreChipCount() (nil NPUJob) error = %v", err)
		}
		if got["node1"] != 0 {
			t.Errorf("scoreChipCount() nil NPUJob = %v, want 0", got)
		}
		// NPUJob.Tasks lacks this task → all 0
		vcJob := SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{
			Tasks: map[api.TaskID]util.NPUTask{"other": {ReqNPUNum: 4, ReqNPUName: util.NPU910CardName}},
		}}}
		got, err = (&ScheduleHandler{}).scoreChipCount(
			&api.TaskInfo{Name: "t1", UID: "t1"}, nodeInfos, vcJob)
		if err != nil {
			t.Fatalf("scoreChipCount() (task missing) error = %v", err)
		}
		if got["node1"] != 0 {
			t.Errorf("scoreChipCount() task missing = %v, want 0", got)
		}
		// node has no idle chips (free=0) while the task needs chips → 0 (guard; out-of-range
		// scenario — candidates satisfy FitNormal ⇒ free≥req)
		zeroFreeNode := &ScheduleHandler{ScheduleEnv: ScheduleEnv{ClusterCache: ClusterCache{Nodes: map[string]NPUNode{
			"node1": {CommonNode: CommonNode{
				Allocate: map[v1.ResourceName]float64{util.NPU910CardName: 8 * util.NPUHexKilo},
				Idle:     map[v1.ResourceName]float64{util.NPU910CardName: 0},
			}},
		}}}}
		vcJob = SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{
			Tasks: map[api.TaskID]util.NPUTask{"t1": {ReqNPUNum: 4, ReqNPUName: util.NPU910CardName}},
		}}}
		got, err = zeroFreeNode.scoreChipCount(
			&api.TaskInfo{Name: "t1", UID: "t1"}, nodeInfos, vcJob)
		if err != nil {
			t.Fatalf("scoreChipCount() (free=0) error = %v", err)
		}
		if got["node1"] != 0 {
			t.Errorf("scoreChipCount() free=0 = %v, want 0", got)
		}
		// over-commit guard: free<req (free=2<req=4) → 0 (out of range; candidates satisfy
		// FitNormal ⇒ free≥req; prevents (req/free)×255>255 truncation or wrap polluting the
		// subHealth (bit9-8)/topo high segments)
		overCommitNode := &ScheduleHandler{ScheduleEnv: ScheduleEnv{ClusterCache: ClusterCache{Nodes: map[string]NPUNode{
			"node1": {CommonNode: CommonNode{
				Allocate: map[v1.ResourceName]float64{util.NPU910CardName: 8 * util.NPUHexKilo},
				Idle:     map[v1.ResourceName]float64{util.NPU910CardName: 2 * util.NPUHexKilo},
			}},
		}}}}
		got, err = overCommitNode.scoreChipCount(
			&api.TaskInfo{Name: "t1", UID: "t1"}, nodeInfos, vcJob)
		if err != nil {
			t.Fatalf("scoreChipCount() (free<req) error = %v", err)
		}
		if got["node1"] != 0 {
			t.Errorf("scoreChipCount() free<req = %v, want 0 (over-commit contributes 0)", got)
		}
	})
}

func TestInitScorePlugins(t *testing.T) {
	h := &ScheduleHandler{}
	h.initScorePlugins()
	if len(h.scorePlugins) != 4 {
		t.Fatalf("initScorePlugins() len = %d, want 4", len(h.scorePlugins))
	}
	wantNames := []string{"topology", "previousNode", "subHealth", "chipCount"}
	wantWeights := []float64{util.ScoreTopoShift, util.ScoreOriginalShift, util.ScoreHealthShift, util.ScoreAvailShift}
	for i, p := range h.scorePlugins {
		if name := p.plugin.Name(); name != wantNames[i] {
			t.Errorf("plugin[%d] name = %s, want %s", i, name, wantNames[i])
		}
		if p.weight != wantWeights[i] {
			t.Errorf("plugin[%d] weight = %v, want %v (shift offset)", i, p.weight, wantWeights[i])
		}
	}
}

func legacyBatchNodeOrder(t *testing.T, h *ScheduleHandler, task *api.TaskInfo,
	nodes []*api.NodeInfo, vcJob SchedulerJob) map[string]float64 {
	t.Helper()
	scoreMap := initScoreMap(nodes)
	errGet := vcJob.policyHandler.ScoreBestNPUNodes(task, nodes, scoreMap)
	if errGet != nil {
		t.Fatalf("legacy topo scoring error = %v", errGet)
	}
	h.addPreferPreviousNodeScore(task, scoreMap, vcJob)
	if h.FaultHandle != nil {
		h.FaultHandle.ScoreBestNPUNodes(task, scoreMap)
	}
	for nodeName := range scoreMap {
		scoreMap[nodeName] *= h.ScoreWeight
	}
	return scoreMap
}

func frameworkBitFormula(t *testing.T, h *ScheduleHandler, task *api.TaskInfo,
	nodes []*api.NodeInfo, vcJob SchedulerJob) map[string]float64 {
	t.Helper()
	topoScore := initScoreMap(nodes)
	errGet := vcJob.policyHandler.ScoreBestNPUNodes(task, nodes, topoScore)
	if errGet != nil {
		t.Fatalf("framework topo scoring error = %v", errGet)
	}
	prevDelta, errPrev := h.scorePreviousNode(task, nodes, vcJob)
	if errPrev != nil {
		t.Fatalf("framework prev scoring error = %v", errPrev)
	}
	subHealthFrame := uniformScoreMap(nodes, 1) // predate defaults to 1 (healthy), topo-independent
	if h.FaultHandle != nil {
		h.FaultHandle.ScoreSubHealthGrade(subHealthFrame)
	}
	chipDelta, errChip := h.scoreChipCount(task, nodes, vcJob)
	if errChip != nil {
		t.Fatalf("framework chip scoring error = %v", errChip)
	}
	out := initScoreMap(nodes)
	for nodeName := range topoScore {
		var b uint16
		switch prevDelta[nodeName] {
		case 2: // tier 2 shifted by ScoreOriginalShift = bit12 P1 (derived tier, no dedicated constant)
			b |= 2 << util.ScoreOriginalShift
		case 1:
			b |= 1 << util.ScoreOriginalShift
		}
		if topoScore[nodeName] > 0 {
			b |= 1 << util.ScoreTopoShift
		}
		b |= uint16(subHealthFrame[nodeName]) << util.ScoreHealthShift // segment {0,1} → bit9-8
		b |= uint16(chipDelta[nodeName])                               // segment integer (in-range via candidate contract, same as synthesizer)
		out[nodeName] = float64(b) * h.ScoreWeight
	}
	return out
}

func chipTopoNodeForScoring(anno string, freeChip int) NPUNode {
	n := NPUNode{CommonNode: CommonNode{
		Annotation: map[string]string{util.TopologyAnnoKey: anno},
		Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: 8 * util.NPUHexKilo},
		Idle:       map[v1.ResourceName]float64{util.NPU910CardName: float64(freeChip) * util.NPUHexKilo},
	}}
	n.ParseChipTopology(&api.NodeInfo{})
	return n
}

// chipTask builds the task needed for chip scoring (soft, req=4; Pod.Annotations non-empty).
func chipTask(req int) *api.TaskInfo {
	return &api.TaskInfo{
		Name: "t1",
		UID:  "t1",
		Resreq: &api.Resource{
			ScalarResources: map[v1.ResourceName]float64{util.NPU910CardName: float64(req) * util.NPUHexKilo},
		},
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default", Annotations: map[string]string{}}},
	}
}

type chipLikePolicy struct {
	mockPolicyHandler
	nodes map[string]NPUNode
	allow bool
}

func (m *chipLikePolicy) ScoreFrameworkAware() bool { return true }

func (m *chipLikePolicy) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	if task == nil || len(nodes) == 0 || len(scoreMap) == 0 {
		return errors.New(util.ArgumentError)
	}
	_, req := util.GetNPURequestFromTask(task)
	if req <= 0 {
		return errors.New(util.ArgumentError)
	}
	allow := m.allow
	for _, node := range nodes {
		if node == nil {
			continue
		}
		nNode, ok := m.nodes[node.Name]
		if !ok {
			continue
		}
		root := nNode.ChipTopo
		if root == nil {
			continue
		}
		scoreMap[node.Name] = root.Score(req, allow)
	}
	return nil
}

// frameworkCase scoring-case tier: the four combinations of prev boost / fault truncate,
// shared by both directions of tests.
type frameworkCase struct {
	name    string
	frame   VolcanoFrame
	fault   FaultHandler
	prefMap map[int]string
	anno    map[string]string
}

func buildFrameworkCases() []frameworkCase {
	return []frameworkCase{
		{name: "no-prev-no-fault"},
		{name: "with-prev", frame: VolcanoFrame{ConfigParameters: ConfigParameters{DynamicParameters: DynamicParameters{PreferPreviousNode: true}}},
			prefMap: map[int]string{0: "node1"}, anno: map[string]string{PodRankIndexKey: "0"}},
		{name: "with-fault", fault: &truncatingFaultHandler{}},
		{name: "fault-prev", frame: VolcanoFrame{ConfigParameters: ConfigParameters{DynamicParameters: DynamicParameters{PreferPreviousNode: true}}},
			fault:   &truncatingFaultHandler{mockFaultHandler{faultByRank: true}},
			prefMap: map[int]string{0: "node1"}, anno: map[string]string{PodRankIndexKey: "0"}},
	}
}

func setupFrameworkCase(t *testing.T, tc frameworkCase, policy SchedulerPluginNeed,
	nodes map[string]NPUNode) (*ScheduleHandler, SchedulerJob, *api.TaskInfo, []*api.NodeInfo) {
	t.Helper()
	h := &ScheduleHandler{
		ScoreWeight: 10,
		FaultHandle: tc.fault,
		ScheduleEnv: ScheduleEnv{
			FrameAttr:    tc.frame,
			ClusterCache: ClusterCache{Nodes: nodes, Jobs: map[api.JobID]SchedulerJob{}},
		},
	}
	h.initScorePlugins()
	vcJob := SchedulerJob{
		Owner: OwnerInfo{OwnerReference: metav1.OwnerReference{UID: "owner-uid"}},
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{},
			NPUJob: &util.NPUJob{ReqNPUName: util.NPU910CardName,
				Tasks: map[api.TaskID]util.NPUTask{"t1": {ReqNPUNum: 4, ReqNPUName: util.NPU910CardName}}},
		},
		policyHandler: policy,
		PrefNodeMap:   tc.prefMap,
	}
	h.ScheduleEnv.ClusterCache.Jobs["job1"] = vcJob
	task := chipTask(4)
	task.Job = "job1"
	if tc.anno != nil {
		task.Pod.Annotations = tc.anno
	}
	return h, vcJob, task, []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}}
}

// assertScoreMapEqual same key set + per-node bit comparison (catches extra/missing nodes).
func assertScoreMapEqual(t *testing.T, got, want map[string]float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("key set mismatch: got %d nodes, want %d", len(got), len(want))
	}
	for nodeName := range want {
		if _, ok := got[nodeName]; !ok {
			t.Errorf("missing node %s (want has it)", nodeName)
		}
		if got[nodeName] != want[nodeName] {
			t.Errorf("node %s: got=%v want=%v, bit mismatch", nodeName, got[nodeName], want[nodeName])
		}
	}
}

type topoWritingPolicy struct {
	mockPolicyHandler
}

func (m *topoWritingPolicy) ScoreBestNPUNodes(_ *api.TaskInfo, _ []*api.NodeInfo,
	scoreMap map[string]float64) error {
	scoreMap["node1"] = 100
	scoreMap["node2"] = 100
	return nil
}

func TestBatchNodeOrderFnLegacyFreeze(t *testing.T) {
	nodes := map[string]NPUNode{
		"node1": chipTopoNodeForScoring("[[0,1,2,3],[4,5,6,7]]", 4),
		"node2": chipTopoNodeForScoring("[[0,1,2,3],[4,5,6,7]]", 6),
	}
	for _, tc := range buildFrameworkCases() {
		t.Run(tc.name, func(t *testing.T) {
			h, vcJob, task, nodeInfos := setupFrameworkCase(t, tc, &topoWritingPolicy{}, nodes)
			score, err := h.BatchNodeOrderFn(task, nodeInfos)
			if err != nil {
				t.Fatalf("BatchNodeOrderFn() error = %v", err)
			}
			legacy := legacyBatchNodeOrder(t, h, task, nodeInfos, vcJob)
			assertScoreMapEqual(t, score, legacy)
		})
	}
}

func TestBatchNodeOrderFnFrameworkBit(t *testing.T) {
	chipTopo := func() map[string]NPUNode {
		return map[string]NPUNode{
			"node1": chipTopoNodeForScoring("[[0,1,2,3],[4,5,6,7]]", 4),
			"node2": chipTopoNodeForScoring("[[0,1,2,3],[4,5,6,7]]", 6),
		}
	}
	for _, tc := range buildFrameworkCases() {
		t.Run(tc.name, func(t *testing.T) {
			nodes := chipTopo()
			h, vcJob, task, nodeInfos := setupFrameworkCase(t, tc, &chipLikePolicy{nodes: nodes}, nodes)
			score, err := h.BatchNodeOrderFn(task, nodeInfos)
			if err != nil {
				t.Fatalf("BatchNodeOrderFn() error = %v", err)
			}
			want := frameworkBitFormula(t, h, task, nodeInfos, vcJob)
			assertScoreMapEqual(t, score, want)
		})
	}
}

func TestBatchNodeOrderByFrameworkEntryGuard(t *testing.T) {
	h := &ScheduleHandler{}
	h.initScorePlugins()
	vcJob := SchedulerJob{policyHandler: &mockPolicyHandler{}}
	if _, err := h.batchNodeOrderByFramework(nil, []*api.NodeInfo{{Name: "node1"}}, vcJob); err == nil {
		t.Error("batchNodeOrderByFramework(nil task) err = nil, want ArgumentError")
	}
	if _, err := h.batchNodeOrderByFramework(&api.TaskInfo{Name: "t1"}, nil, vcJob); err == nil {
		t.Error("batchNodeOrderByFramework(empty nodes) err = nil, want ArgumentError")
	}
}

func TestBatchNodeOrderByFrameworkNoPlugins(t *testing.T) {
	h := &ScheduleHandler{ScoreWeight: 10} // scorePlugins not initialized (nil)
	vcJob := SchedulerJob{policyHandler: &mockPolicyHandler{}}
	score, err := h.batchNodeOrderByFramework(&api.TaskInfo{Name: "t1"},
		[]*api.NodeInfo{{Name: "node1"}}, vcJob)
	if err != nil {
		t.Fatalf("batchNodeOrderByFramework() error = %v", err)
	}
	if score == nil {
		t.Fatal("batchNodeOrderByFramework() score = nil, want non-nil empty map")
	}
	if len(score) != 0 {
		t.Errorf("scores = %v, want empty map (empty registry → no dimension contribution)", score)
	}
}

func TestScorePreviousNodeFaultPodOtherP1(t *testing.T) {
	h := &ScheduleHandler{
		FaultHandle: &mockFaultHandler{faultByRank: true},
		ScheduleEnv: ScheduleEnv{
			FrameAttr: VolcanoFrame{ConfigParameters: ConfigParameters{
				DynamicParameters: DynamicParameters{PreferPreviousNode: true},
			}},
		},
	}
	nodeInfos := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}} // node1=selfNode(fault), node2=sole otherNode
	vcJob := SchedulerJob{
		Owner: OwnerInfo{OwnerReference: metav1.OwnerReference{UID: "owner-uid"}},
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{},
			NPUJob: &util.NPUJob{NPUTaskNum: 2, Tasks: fakeTasksForRank()},
		},
		PrefNodeMap:   map[int]string{0: "node1"},
		policyHandler: &mockPolicyHandler{},
	}
	task := &api.TaskInfo{
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "0"}}},
		Job: "job1",
	}
	got, err := h.scorePreviousNode(task, nodeInfos, vcJob)
	if err != nil {
		t.Fatalf("scorePreviousNode() error = %v", err)
	}
	if got["node1"] != 1 {
		t.Errorf("scorePreviousNode() node1(%v) = %v, want 1 (P2 — selfNode fallback, since fault-pod prefers otherNodes)",
			got, got["node1"])
	}
	if got["node2"] != 2 {
		t.Errorf("scorePreviousNode() node2 = %v, want 2 (P1 — otherNode leaves the fault first)", got["node2"])
	}
}

func TestScorePreviousNodeFaultPodSelfInCandidates(t *testing.T) {
	h := &ScheduleHandler{
		FaultHandle: &mockFaultHandler{faultByRank: true},
		ScheduleEnv: ScheduleEnv{
			FrameAttr: VolcanoFrame{ConfigParameters: ConfigParameters{
				DynamicParameters: DynamicParameters{PreferPreviousNode: true},
			}},
		},
	}
	nodeInfos := []*api.NodeInfo{{Name: "node1"}} // no otherNodes
	vcJob := SchedulerJob{
		Owner: OwnerInfo{OwnerReference: metav1.OwnerReference{UID: "owner-uid"}},
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{},
			NPUJob: &util.NPUJob{NPUTaskNum: 2, Tasks: fakeTasksForRank()},
		},
		PrefNodeMap:   map[int]string{0: "node1"},
		policyHandler: &mockPolicyHandler{},
	}
	task := &api.TaskInfo{
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "0"}}},
		Job: "job1",
	}
	got, err := h.scorePreviousNode(task, nodeInfos, vcJob)
	if err != nil {
		t.Fatalf("scorePreviousNode() error = %v", err)
	}
	if got["node1"] != 1 {
		t.Errorf("scorePreviousNode() fault fallback = %v, want node1=1 (P2 — sole candidate is the fault-pod's selfNode)",
			got)
	}
}

func TestScorePreviousNodeNoSelfPeerFallback(t *testing.T) {
	h := &ScheduleHandler{
		ScheduleEnv: ScheduleEnv{
			FrameAttr: VolcanoFrame{ConfigParameters: ConfigParameters{
				DynamicParameters: DynamicParameters{PreferPreviousNode: true},
			}},
		},
	}
	nodeInfos := []*api.NodeInfo{{Name: "node2"}} // selfNode=node1 absent from candidates (nodes); node2 is a peerNode
	vcJob := SchedulerJob{
		Owner: OwnerInfo{OwnerReference: metav1.OwnerReference{UID: "owner-uid"}},
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{},
			NPUJob: &util.NPUJob{NPUTaskNum: 2, Tasks: fakeTasksForRank()},
		},
		PrefNodeMap:   map[int]string{0: "node1", 1: "node2"},
		policyHandler: &mockPolicyHandler{},
	}
	task := &api.TaskInfo{
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "0"}}},
		Job: "job1",
	}
	got, err := h.scorePreviousNode(task, nodeInfos, vcJob)
	if err != nil {
		t.Fatalf("scorePreviousNode() error = %v", err)
	}
	if got["node2"] != 0 {
		t.Errorf("scorePreviousNode() = %v, want node2=0 (peerNodes never boosted — no peer fallback)", got)
	}
}

// TestScorePreviousNodeCategoryTiers locks the category-wide tier assignment: every node in a
// category carries that category's tier (no arbitrary best-node pick), for both fault and
// normal pods — selfNode, a peerNode, and two otherNodes all in the candidate list.
func TestScorePreviousNodeCategoryTiers(t *testing.T) {
	nodeInfos := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}, {Name: "node3"}, {Name: "node4"}}
	prefMap := map[int]string{0: "node1", 1: "node2"} // selfNode=node1, peerNode=node2, others=node3,node4
	task := &api.TaskInfo{
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "0"}}},
		Job: "job1",
	}
	run := func(fault bool) map[string]int {
		var faultHandle FaultHandler
		if fault {
			faultHandle = &mockFaultHandler{faultByRank: true}
		}
		h := &ScheduleHandler{
			FaultHandle: faultHandle,
			ScheduleEnv: ScheduleEnv{
				FrameAttr: VolcanoFrame{ConfigParameters: ConfigParameters{
					DynamicParameters: DynamicParameters{PreferPreviousNode: true},
				}},
			},
		}
		vcJob := SchedulerJob{
			Owner: OwnerInfo{OwnerReference: metav1.OwnerReference{UID: "owner-uid"}},
			SchedulerJobAttr: util.SchedulerJobAttr{
				ComJob: util.ComJob{},
				NPUJob: &util.NPUJob{NPUTaskNum: 2, Tasks: fakeTasksForRank()},
			},
			PrefNodeMap:   prefMap,
			policyHandler: &mockPolicyHandler{},
		}
		got, err := h.scorePreviousNode(task, nodeInfos, vcJob)
		if err != nil {
			t.Fatalf("scorePreviousNode() error = %v", err)
		}
		return got
	}
	// normal pod: selfNode P1, every otherNode P2, peerNode 0.
	got := run(false)
	want := map[string]int{"node1": 2, "node2": 0, "node3": 1, "node4": 1}
	for _, n := range nodeInfos {
		if got[n.Name] != want[n.Name] {
			t.Errorf("normal scorePreviousNode() [%s] = %v, want %v", n.Name, got[n.Name], want[n.Name])
		}
	}
	// fault pod: every otherNode P1, selfNode P2 fallback, peerNode 0.
	got = run(true)
	want = map[string]int{"node1": 1, "node2": 0, "node3": 2, "node4": 2}
	for _, n := range nodeInfos {
		if got[n.Name] != want[n.Name] {
			t.Errorf("fault scorePreviousNode() [%s] = %v, want %v", n.Name, got[n.Name], want[n.Name])
		}
	}
}

func TestScoreBitFieldStrictPriority(t *testing.T) {
	// largest combination without P1: P2(bit11)|topo(bit10)|subHealth(1<<8 at max)|avail(0xFF full-range literal)
	noP1Max := uint16(1<<util.ScoreOriginalShift | 1<<util.ScoreTopoShift |
		1<<util.ScoreHealthShift | 0xFF)
	p1Min := uint16(2 << util.ScoreOriginalShift) // tier 2 shifted = bit12 (smallest P1 0x1000)
	if p1Min <= noP1Max {
		t.Errorf("P1 seg must dominate all lower combos: p1Min=%d noP1Max=%d", p1Min, noP1Max)
	}
	// largest combination without P2: topo|subHealth(1)|avail
	noP2Max := uint16(1<<util.ScoreTopoShift | 1<<util.ScoreHealthShift | 0xFF)
	p2Min := uint16(1 << util.ScoreOriginalShift)
	if p2Min <= noP2Max {
		t.Errorf("P2 seg must dominate all lower combos: p2Min=%d noP2Max=%d", p2Min, noP2Max)
	}
}
