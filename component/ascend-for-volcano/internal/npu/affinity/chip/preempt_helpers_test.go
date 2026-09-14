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

// Shared preempt/reclaim test helpers. Placed in a build-tag-free file so both
// the tagged (volcano_v115, topology-aware path) and the untagged (pre-v1.15
// selective path) test builds can use them, and so the v1.15 test file does not
// redeclare them.

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

// newNames returns the non-nil tasks' names in order.
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

func enableTopoAware(t *testing.T) {
	t.Helper()
	TopologyAwarePreemptActive = true
	t.Cleanup(func() { TopologyAwarePreemptActive = false })
}

// allNetUnhealthyNode builds a flat n-chip node whose every chip is
// parameter-plane network unhealthy and free.
func allNetUnhealthyNode(n int) *topo.ChipNode {
	root := topo.ParseTopology(topo.BuildFlatTopology(n))
	if root == nil {
		panic("parse BuildFlatTopology failed")
	}
	netUnh := make(map[int]struct{}, n)
	for i := 0; i < n; i++ {
		netUnh[i] = struct{}{}
	}
	root.Init(nil, netUnh, nil)
	return root
}
