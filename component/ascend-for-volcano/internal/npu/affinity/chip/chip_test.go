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
	"strconv"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/affinity/chip/topo"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// mockChipTask builds a task requesting reqNum NPUs on a huawei.com 910 resource.
func mockChipTask(reqNum int) *api.TaskInfo {
	return &api.TaskInfo{
		Name: "task1",
		Pod:  &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: "uid-1"}},
		Resreq: &api.Resource{
			ScalarResources: map[v1.ResourceName]float64{
				v1.ResourceName(util.HwPreName + util.Ascend910): float64(reqNum * util.NPUHexKilo),
			},
		},
	}
}

func mockChipNode(name string, root *topo.ChipNode) plugin.NPUNode {
	n := plugin.NPUNode{CommonNode: plugin.CommonNode{Name: name, Annotation: map[string]string{}}}
	n.ChipTopo = root
	return n
}

// flatNode builds a node with a real flat topology [0..n-1] and no faults/owners.
func flatNode(name string, n int) plugin.NPUNode {
	root := topo.ParseTopology(topo.BuildFlatTopology(n))
	root.Init(nil, nil, nil)
	return mockChipNode(name, root)
}

// fullCardAnno returns the allocation annotation for all n chips, e.g.
// "Ascend910-0,Ascend910-1,...": UpdateNodeInfo reads it to compute the leftover list.
func fullCardAnno(n int) string {
	parts := make([]string, 0, n)
	for i := 0; i < n; i++ {
		parts = append(parts, util.NPU910CardNamePre+strconv.Itoa(i))
	}
	return strings.Join(parts, ",")
}

// newTestHandler returns a handler whose embedded *NPUJob is populated.
// ScheduleMode / ParameterPlaneUnhealthyTolerance live on NPUJob (an embedded
// pointer), and MaxNodeNPUNum gates UpdateNodeInfo; a bare &chipHandler{} would
// nil-deref / wrongly reject when reading them.
func newTestHandler() *chipHandler {
	h := &chipHandler{}
	h.NPUJob = &util.NPUJob{}
	h.MaxNodeNPUNum = maxNodeNPUNum
	return h
}

func TestNew(t *testing.T) {
	h := New().(*chipHandler)
	if h.GetPluginName() != PolicyName {
		t.Errorf("GetPluginName() = %q, want %q", h.GetPluginName(), PolicyName)
	}
	if h.MaxNodeNPUNum != maxNodeNPUNum {
		t.Errorf("MaxNodeNPUNum = %d, want %d", h.MaxNodeNPUNum, maxNodeNPUNum)
	}
}

func TestShouldUseAffinity(t *testing.T) {
	tests := []struct {
		name string
		ann  map[string]string
		sel  map[string]string
		want bool
	}{
		{"affinity by default", nil, nil, true},
		{"empty maps", map[string]string{}, map[string]string{}, true},
		{"schedule policy opts out", map[string]string{util.SchedulePolicyAnnoKey: "xxx"}, nil, false},
		{"deprecated accelerator selector opts out", nil, map[string]string{util.AcceleratorTypeKeyDeprecated: "A5000"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldUseAffinity(tt.ann, tt.sel); got != tt.want {
				t.Errorf("ShouldUseAffinity(%v, %v) = %v, want %v", tt.ann, tt.sel, got, tt.want)
			}
		})
	}
}

func TestCheckNodeNPUByTask(t *testing.T) {
	good := mockChipTask(4)

	var nilHandler *chipHandler
	if err := nilHandler.CheckNodeNPUByTask(good, mockChipNode("n1", nil)); err == nil {
		t.Error("nil handler should error")
	} else if err.Error() != util.ArgumentError {
		t.Errorf("nil handler err = %v, want %s", err, util.ArgumentError)
	}

	tp := newTestHandler()
	if err := tp.CheckNodeNPUByTask(nil, mockChipNode("n1", nil)); err == nil {
		t.Error("nil task should error")
	}
	if err := tp.CheckNodeNPUByTask(mockChipTask(0), mockChipNode("n1", nil)); err == nil {
		t.Error("req 0 should error")
	}

	// missing topology tree
	if err := tp.CheckNodeNPUByTask(good, mockChipNode("n2", nil)); err == nil {
		t.Error("nil ChipTopo should error")
	} else if !strings.Contains(err.Error(), util.NPUResourceUnavailableError) {
		t.Errorf("nil ChipTopo err = %v, want it to contain %s", err, util.NPUResourceUnavailableError)
	}

	// real 8-chip flat topology fits a 4-chip task
	if err := tp.CheckNodeNPUByTask(good, flatNode("n3", 8)); err != nil {
		t.Errorf("flat8 req4 should pass, got %v", err)
	}

	// too small: Fit returns FitFailed -> NPUResourceUnavailableError
	if err := tp.CheckNodeNPUByTask(good, flatNode("n4", 2)); err == nil {
		t.Error("flat2 req4 should error")
	} else if !strings.Contains(err.Error(), util.NPUResourceUnavailableError) {
		t.Errorf("flat2 err = %v, want it to contain %s", err, util.NPUResourceUnavailableError)
	}

	// fully busy: only eviction can satisfy the request -> NPUResourceShortageError
	busy := topo.ParseTopology(topo.BuildFlatTopology(8))
	busy.Init(nil, nil, map[string][]int{"o": {0, 1, 2, 3, 4, 5, 6, 7}})
	if err := tp.CheckNodeNPUByTask(good, mockChipNode("n5", busy)); err == nil {
		t.Error("fully-busy flat8 req4 soft should report shortage")
	} else if !strings.Contains(err.Error(), util.NPUResourceShortageError) {
		t.Errorf("busy err = %v, want it to contain %s", err, util.NPUResourceShortageError)
	}
}

func TestScoreBestNPUNodesErrors(t *testing.T) {
	tp := newTestHandler()
	good := mockChipTask(2)
	nodes := []*api.NodeInfo{{Name: "n1"}}
	scoreMap := map[string]float64{"n1": 0}

	if err := tp.ScoreBestNPUNodes(nil, nodes, scoreMap); err == nil {
		t.Error("nil task should error")
	}
	if err := tp.ScoreBestNPUNodes(good, nil, scoreMap); err == nil {
		t.Error("nil nodes should error")
	}
	if err := tp.ScoreBestNPUNodes(good, nodes, map[string]float64{}); err == nil {
		t.Error("empty scoreMap should error")
	}
	if err := tp.ScoreBestNPUNodes(mockChipTask(0), nodes, scoreMap); err == nil {
		t.Error("req 0 should error")
	}
}

func TestUseAnnotation(t *testing.T) {
	tp := newTestHandler()
	tp.ReqNPUName = util.NPU910CardName

	// no topology tree -> nil
	if got := tp.UseAnnotation(mockChipTask(4), mockChipNode("n1", nil)); got != nil {
		t.Errorf("no-topo UseAnnotation = %v, want nil", got)
	}

	// topology too small to satisfy the request -> SelectChips returns nil
	if got := tp.UseAnnotation(mockChipTask(4), flatNode("n2", 2)); got != nil {
		t.Errorf("too-small topo UseAnnotation = %v, want nil", got)
	}

	// topology fits: real SelectChips picks [0,1,2,3], TryAllocate registers them
	node := flatNode("n3", 8)
	node.Annotation[util.NPU910CardName] = fullCardAnno(8)
	got := tp.UseAnnotation(mockChipTask(4), node)
	if got == nil {
		t.Fatal("fits UseAnnotation = nil, want *NPUNode")
	}
	// leftover annotation must only keep the unallocated chips
	left := node.Annotation[util.NPU910CardName]
	if strings.Contains(left, "Ascend910-0") {
		t.Errorf("leftover annotation still contains selected chip 0: %q", left)
	}
	if !strings.Contains(left, "Ascend910-7") {
		t.Errorf("leftover annotation lost unselected chip 7: %q", left)
	}
	// the same pod cannot be allocated twice -> rejected
	if got2 := tp.UseAnnotation(mockChipTask(4), node); got2 != nil {
		t.Errorf("duplicate UseAnnotation should be rejected, got %v", got2)
	}
}

func TestReleaseAnnotation(t *testing.T) {
	tp := newTestHandler()
	tp.ReqNPUName = util.NPU910CardName

	// node without a topology still returns the node
	if got := tp.ReleaseAnnotation(mockChipTask(4), mockChipNode("n1", nil)); got == nil {
		t.Error("ReleaseAnnotation on no-topo node = nil, want *NPUNode")
	}

	// unknown task: rollback error is tolerated, node still returned
	if got := tp.ReleaseAnnotation(mockChipTask(4), flatNode("n2", 8)); got == nil {
		t.Error("ReleaseAnnotation on unknown task = nil, want *NPUNode")
	}

	// a task allocated through UseAnnotation is released -> chips become usable
	node := flatNode("n3", 8)
	node.Annotation[util.NPU910CardName] = fullCardAnno(8)
	if got := tp.UseAnnotation(mockChipTask(4), node); got == nil {
		t.Fatal("UseAnnotation should succeed before release")
	}
	if got := tp.ReleaseAnnotation(mockChipTask(4), node); got == nil {
		t.Fatal("ReleaseAnnotation after allocation = nil, want *NPUNode")
	}
	if fit := node.ChipTopo.Fit(&util.Request{ReqNPUNum: 8, Mode: util.SoftScheduleMode}); fit != topo.FitNormal {
		t.Errorf("after release all 8 chips should be usable, Fit = %d, want FitNormal", fit)
	}
}
