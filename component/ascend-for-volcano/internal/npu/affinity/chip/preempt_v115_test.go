//go:build volcano_v115

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

package chip

import (
	"fmt"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/affinity/chip/topo"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func newTwoGroupTree(owners map[string][]int) *topo.ChipNode {
	root := topo.ParseTopology("[[0,1,2,3],[4,5,6,7]]")
	if root == nil {
		panic("parse [[0,1,2,3],[4,5,6,7]] failed")
	}
	root.Init(nil, nil, owners)
	return root
}

func newTestNode(root *topo.ChipNode) *plugin.NPUNode {
	return &plugin.NPUNode{CommonNode: plugin.CommonNode{Name: "node-1", ChipTopo: root}}
}

func newPeTask(id int, chips ...int) *api.TaskInfo {
	names := make([]string, 0, len(chips))
	for _, c := range chips {
		names = append(names, fmt.Sprintf("%s%d", util.NPU910CardNamePre, c))
	}
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{
		UID:         types.UID(fmt.Sprintf("uid-%d", id)),
		Annotations: map[string]string{util.NPU910CardName: strings.Join(names, ",")},
	}}
	return &api.TaskInfo{
		UID:  api.TaskID(fmt.Sprintf("uid-%d", id)),
		Name: fmt.Sprintf("pod-%d", id),
		Pod:  pod,
	}
}

func newPreemptorTask(req int, mode string) *api.TaskInfo {
	anno := map[string]string{}
	if mode != "" {
		anno[util.ScheduleModeAnnoKey] = mode
	}
	return &api.TaskInfo{
		UID:  "uid-preemptor",
		Name: "preemptor",
		Pod:  &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "preemptor", UID: "uid-preemptor", Annotations: anno}},
		Resreq: &api.Resource{ScalarResources: map[v1.ResourceName]float64{
			v1.ResourceName(util.HwPreName + "Ascend910"): float64(req * util.NPUHexKilo),
		}},
	}
}

func newNames(tasks []*api.TaskInfo) []string {
	var out []string
	for _, t := range tasks {
		if t == nil {
			continue
		}
		out = append(out, t.Name)
	}
	return out
}

func TestPreemptOrReclaimNilGuards(t *testing.T) {
	tp := &chipHandler{}
	pre := newPreemptorTask(2, "")
	cands := []*api.TaskInfo{newPeTask(0, 0)}
	node := newTestNode(newTwoGroupTree(nil))
	for name, call := range map[string]func() ([]*api.TaskInfo, bool){
		"nil tp":        func() ([]*api.TaskInfo, bool) { var h *chipHandler; return h.Preemptable(pre, cands, node) },
		"nil preemptor": func() ([]*api.TaskInfo, bool) { return tp.Preemptable(nil, cands, node) },
		"nil node":      func() ([]*api.TaskInfo, bool) { return tp.Preemptable(pre, cands, nil) },
		"no req":        func() ([]*api.TaskInfo, bool) { return tp.Preemptable(newPreemptorTask(0, ""), cands, node) },
		"no topo":       func() ([]*api.TaskInfo, bool) { return tp.Preemptable(pre, cands, newTestNode(nil)) },
	} {
		sel, ok := call()
		if ok || sel != nil {
			t.Fatalf("%s: want (nil,false), got selected=%v ok=%v", name, newNames(sel), ok)
		}
	}
}

func enableTopoAware(t *testing.T) {
	t.Helper()
	TopologyAwarePreemptActive = true
	t.Cleanup(func() { TopologyAwarePreemptActive = false })
}

func TestPreemptOrReclaimSoftZeroEviction(t *testing.T) {
	enableTopoAware(t)
	tp := newTestHandler()
	root := newTwoGroupTree(nil)
	sel, ok := tp.Preemptable(newPreemptorTask(2, ""), nil, newTestNode(root))
	if !ok || sel != nil {
		t.Fatalf("want (nil,true), got selected=%v ok=%v", newNames(sel), ok)
	}
}

func TestPreemptOrReclaimHardZeroEviction(t *testing.T) {
	enableTopoAware(t)
	tp := newTestHandler()
	tp.NPUJob.ScheduleMode = util.HardScheduleMode
	owners := map[string][]int{"p4": {4}, "p5": {5}, "p6": {6}, "p7": {7}}
	root := newTwoGroupTree(owners)
	sel, ok := tp.Preemptable(newPreemptorTask(4, "hard"), nil, newTestNode(root))
	if !ok || sel != nil {
		t.Fatalf("want (nil,true), got selected=%v ok=%v", newNames(sel), ok)
	}
}

func TestPreemptOrReclaimReturnsAllSoft(t *testing.T) {
	enableTopoAware(t)
	tp := newTestHandler()
	owners := map[string][]int{"p0": {0}, "p1": {1}, "p2": {2}, "p3": {3}}
	root := newTwoGroupTree(owners)
	preemptees := []*api.TaskInfo{newPeTask(0, 0), newPeTask(1, 1), newPeTask(2, 2), newPeTask(3, 3)}
	sel, ok := tp.Preemptable(newPreemptorTask(5, ""), preemptees, newTestNode(root))
	if !ok {
		t.Fatalf("want ok=true")
	}
	if len(sel) != len(preemptees) {
		t.Fatalf("want all candidates (%d), got %v", len(preemptees), newNames(sel))
	}
}

func TestPreemptOrReclaimReturnsAllHard(t *testing.T) {
	enableTopoAware(t)
	tp := newTestHandler()
	tp.NPUJob.ScheduleMode = util.HardScheduleMode
	owners := map[string][]int{"p0": {0}, "p1": {1}, "p2": {2}, "p3": {3}, "p4": {4}, "p5": {5}, "p6": {6}, "p7": {7}}
	root := newTwoGroupTree(owners)
	preemptees := []*api.TaskInfo{
		newPeTask(0, 0), newPeTask(1, 1), newPeTask(2, 2), newPeTask(3, 3),
		newPeTask(4, 4), newPeTask(5, 5), newPeTask(6, 6), newPeTask(7, 7),
	}
	sel, ok := tp.Preemptable(newPreemptorTask(4, "hard"), preemptees, newTestNode(root))
	if !ok {
		t.Fatalf("want ok=true")
	}
	if len(sel) != len(preemptees) {
		t.Fatalf("want all candidates (%d), got %v", len(preemptees), newNames(sel))
	}
}

func TestPreemptOrReclaimNotSatisfiable(t *testing.T) {
	enableTopoAware(t)
	tp := newTestHandler()
	root := topo.ParseTopology("[[0,1,2,3],[4,5,6,7]]")
	owners := map[string][]int{"p2": {2}, "p3": {3}, "p4": {4}, "p5": {5}, "p6": {6}, "p7": {7}}
	root.Init(map[int]struct{}{0: {}, 1: {}}, nil, owners)
	preemptees := []*api.TaskInfo{
		newPeTask(2, 2), newPeTask(3, 3), newPeTask(4, 4),
		newPeTask(5, 5), newPeTask(6, 6), newPeTask(7, 7),
	}
	sel, ok := tp.Preemptable(newPreemptorTask(8, ""), preemptees, newTestNode(root))
	if ok || sel != nil {
		t.Fatalf("want (nil,false), got selected=%v ok=%v", newNames(sel), ok)
	}
}

func TestPreemptOrReclaimReclaimSetOnlyPreemptees(t *testing.T) {
	enableTopoAware(t)
	tp := newTestHandler()
	owners := map[string][]int{"p0": {0}, "p1": {1}, "p2": {2}, "p3": {3}, "p4": {4}, "p5": {5}}
	root := newTwoGroupTree(owners)
	preemptees := []*api.TaskInfo{newPeTask(0, 0), newPeTask(1, 1), newPeTask(2, 2), newPeTask(3, 3)}
	sel, ok := tp.Preemptable(newPreemptorTask(7, ""), preemptees, newTestNode(root))
	if ok || sel != nil {
		t.Fatalf("want (nil,false), got selected=%v ok=%v", newNames(sel), ok)
	}
}

func TestPreemptOrReclaimReclaimableIDs(t *testing.T) {
	root := newTwoGroupTree(nil)
	node := newTestNode(root)
	preemptees := []*api.TaskInfo{newPeTask(0, 0), newPeTask(1, 1, 2)}
	got := reclaimableIDsOf(preemptees, node)
	want := map[int]struct{}{0: {}, 1: {}, 2: {}}
	if len(got) != len(want) {
		t.Fatalf("want %d ids, got %d", len(want), len(got))
	}
	for id := range want {
		if _, ok := got[id]; !ok {
			t.Fatalf("missing recovered id %d", id)
		}
	}
}

func TestRouteEvictionParamRouting(t *testing.T) {
	orig := TopologyAwarePreemptActive
	defer func() { TopologyAwarePreemptActive = orig }()

	tp := newTestHandler()
	tp.NPUJob.ScheduleMode = util.HardScheduleMode
	owners := map[string][]int{"p0": {0}, "p1": {1}, "p2": {2}, "p3": {3}, "p4": {4}, "p5": {5}, "p6": {6}, "p7": {7}}
	root := newTwoGroupTree(owners)
	preemptees := []*api.TaskInfo{
		newPeTask(0, 0), newPeTask(1, 1), newPeTask(2, 2), newPeTask(3, 3),
		newPeTask(4, 4), newPeTask(5, 5), newPeTask(6, 6), newPeTask(7, 7),
	}

	TopologyAwarePreemptActive = true
	sel, ok := tp.Preemptable(newPreemptorTask(4, "hard"), preemptees, newTestNode(root))
	if !ok {
		t.Fatalf("[on] want ok=true")
	}
	if len(sel) != len(preemptees) {
		t.Fatalf("[on] want all %d candidates (framework convergence), got %v", len(preemptees), newNames(sel))
	}

	TopologyAwarePreemptActive = false
	sel, ok = tp.Preemptable(newPreemptorTask(4, "hard"), preemptees, newTestNode(root))
	if !ok {
		t.Fatalf("[off] want ok=true")
	}
	if len(sel) != 4 {
		t.Fatalf("[off] want 4 victims (select one group), got %v", newNames(sel))
	}
}

// TestReclaimableAlwaysSelective pins reclaim to selective eviction no matter
// TopologyAwarePreemptActive: on the v1.15 path only Preemptable is allowed to
// use the topology-aware full-candidate set (used by TestRouteEvictionParamRouting
// to prove the routing differs), reclaim must always pick the minimal group.
func TestReclaimableAlwaysSelective(t *testing.T) {
	for _, aware := range []bool{true, false} {
		TopologyAwarePreemptActive = aware
		tp := newTestHandler()
		tp.NPUJob.ScheduleMode = util.HardScheduleMode
		owners := map[string][]int{"p0": {0}, "p1": {1}, "p2": {2}, "p3": {3}, "p4": {4}, "p5": {5}, "p6": {6}, "p7": {7}}
		root := newTwoGroupTree(owners)
		reclaimees := []*api.TaskInfo{
			newPeTask(0, 0), newPeTask(1, 1), newPeTask(2, 2), newPeTask(3, 3),
			newPeTask(4, 4), newPeTask(5, 5), newPeTask(6, 6), newPeTask(7, 7),
		}
		sel, ok := tp.Reclaimable(newPreemptorTask(4, "hard"), reclaimees, newTestNode(root))
		if !ok {
			t.Fatalf("aware=%v: want ok=true", aware)
		}
		if len(sel) != 4 {
			t.Fatalf("aware=%v: want 4 victims from one group (selective), got %v", aware, newNames(sel))
		}
	}
	TopologyAwarePreemptActive = false
}
