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

package topo

import (
	"fmt"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

func twoGroupTree(owners map[string][]int) *ChipNode {
	root := ParseTopology("[[0,1,2,3],[4,5,6,7]]")
	if root == nil {
		panic("parse [[0,1,2,3],[4,5,6,7]] failed")
	}
	root.Init(nil, nil, owners)
	return root
}

func peTask(id int, chips ...int) *api.TaskInfo {
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

func names(tasks []*api.TaskInfo) []string {
	var out []string
	for _, t := range tasks {
		if t == nil {
			continue
		}
		out = append(out, t.Name)
	}
	return out
}

func TestSelectEvictSoftZeroEviction(t *testing.T) {
	selected, ok := twoGroupTree(nil).SelectEvictSoft(2, false, nil)
	if !ok || selected != nil {
		t.Fatalf("want (nil,true), got selected=%v ok=%v", names(selected), ok)
	}
}

func TestSelectEvictSoftMinimal(t *testing.T) {
	owners := map[string][]int{"p0": {0}, "p1": {1}, "p2": {2}, "p3": {3}}
	root := twoGroupTree(owners)
	preemptees := []*api.TaskInfo{peTask(0, 0), peTask(1, 1), peTask(2, 2), peTask(3, 3)}
	selected, ok := root.SelectEvictSoft(5, false, preemptees)
	if !ok {
		t.Fatalf("want ok=true")
	}
	if len(selected) != 1 || selected[0].Name != "pod-0" {
		t.Fatalf("want single victim pod-0, got %v", names(selected))
	}
}

func TestSelectEvictSoftNotSatisfiable(t *testing.T) {
	root := ParseTopology("[[0,1,2,3],[4,5,6,7]]")
	owners := map[string][]int{"p2": {2}, "p3": {3}, "p4": {4}, "p5": {5}, "p6": {6}, "p7": {7}}
	root.Init(map[int]struct{}{0: {}, 1: {}}, nil, owners)
	preemptees := []*api.TaskInfo{peTask(2, 2), peTask(3, 3), peTask(4, 4), peTask(5, 5), peTask(6, 6), peTask(7, 7)}
	selected, ok := root.SelectEvictSoft(7, false, preemptees)
	if ok || selected != nil {
		t.Fatalf("want (nil,false), got selected=%v ok=%v", names(selected), ok)
	}
}

func TestSelectEvictSoftUnhealthyChip(t *testing.T) {
	root := ParseTopology("[[0,1,2,3],[4,5,6,7]]")
	owners := map[string][]int{"p0": {0}, "p1": {1}}
	root.Init(nil, map[int]struct{}{0: {}}, owners)
	preemptees := []*api.TaskInfo{peTask(0, 0), peTask(1, 1)}
	selected, ok := root.SelectEvictSoft(7, false, preemptees)
	if !ok {
		t.Fatalf("want ok=true")
	}
	if len(selected) != 2 {
		t.Fatalf("want 2 victims (p0 unhealthy chip not counted w/o allow), got %v", names(selected))
	}
}

// TestSelectEvictSoftFreeFirstOnlyGap pins the free-chips-first contract: the 3 free
// chips {2,3,7} are counted before any victim, so req 5 only needs the 2 chips held by
// pod-0 and nothing more is evicted — eviction tops up the gap, never over-evicts.
func TestSelectEvictSoftFreeFirstOnlyGap(t *testing.T) {
	owners := map[string][]int{"p0": {0, 1}, "p1": {4, 5}, "p2": {6}}
	root := twoGroupTree(owners)
	preemptees := []*api.TaskInfo{peTask(0, 0, 1), peTask(1, 4, 5), peTask(2, 6)}
	selected, ok := root.SelectEvictSoft(5, false, preemptees)
	if !ok {
		t.Fatalf("want ok=true")
	}
	if len(selected) != 1 || selected[0].Name != "pod-0" {
		t.Fatalf("want exactly 1 victim pod-0 (free 2,3,7 cover 3 of req 5), got %v", names(selected))
	}
}

func TestSelectEvictHardZeroEviction(t *testing.T) {
	owners := map[string][]int{"p4": {4}, "p5": {5}, "p6": {6}, "p7": {7}}
	root := twoGroupTree(owners)
	preemptees := []*api.TaskInfo{peTask(4, 4), peTask(5, 5), peTask(6, 6), peTask(7, 7)}
	selected, ok := root.SelectEvictHard(4, false, preemptees)
	if !ok || selected != nil {
		t.Fatalf("want (nil,true), got selected=%v ok=%v", names(selected), ok)
	}
}

func TestSelectEvictHardGroupLocal(t *testing.T) {
	owners := map[string][]int{
		"p0": {0}, "p1": {1}, "p2": {2}, "p3": {3},
		"p4": {4}, "p5": {5}, "p6": {6}, "p7": {7},
	}
	root := twoGroupTree(owners)
	preemptees := []*api.TaskInfo{
		peTask(0, 0), peTask(1, 1), peTask(2, 2), peTask(3, 3),
		peTask(4, 4), peTask(5, 5), peTask(6, 6), peTask(7, 7),
	}
	selected, ok := root.SelectEvictHard(4, false, preemptees)
	if !ok {
		t.Fatalf("want ok=true")
	}
	if len(selected) != 4 {
		t.Fatalf("want 4 victims (one group only), got %v", names(selected))
	}
	for _, pe := range selected {
		if pe.Pod == nil {
			continue
		}
		if !strings.Contains(pe.Pod.Annotations[util.NPU910CardName], "Ascend910-") {
			t.Fatalf("unexpected victim %s", pe.Name)
		}
	}
}

func TestSelectEvictHardMinimalGroup(t *testing.T) {
	owners := map[string][]int{"p0": {0}, "p1": {1}, "p4": {4}, "p5": {5}, "p6": {6}, "p7": {7}}
	root := twoGroupTree(owners)
	preemptees := []*api.TaskInfo{peTask(0, 0), peTask(1, 1), peTask(4, 4), peTask(5, 5), peTask(6, 6), peTask(7, 7)}
	selected, ok := root.SelectEvictHard(4, false, preemptees)
	if !ok {
		t.Fatalf("want ok=true")
	}
	if len(selected) != 2 || selected[0].Name != "pod-0" || selected[1].Name != "pod-1" {
		t.Fatalf("want group0 minimal [pod-0 pod-1], got %v", names(selected))
	}
}

// TestSelectEvictHardWholeTreeFill req=8 = total capacity: canFitHardCheck accepts the
// whole-subtree fill class (req == n.total), so eviction must match by emptying the
// whole tree. Only the actual holders among the preemptees need to be evicted; chips
// 4-7 are never allocated and need no victims.
func TestSelectEvictHardWholeTreeFill(t *testing.T) {
	owners := map[string][]int{"p0": {0}, "p1": {1}, "p2": {2}, "p3": {3}}
	root := twoGroupTree(owners)
	preemptees := []*api.TaskInfo{peTask(0, 0), peTask(1, 1), peTask(2, 2), peTask(3, 3)}
	selected, ok := root.SelectEvictHard(8, false, preemptees)
	if !ok {
		t.Fatalf("want ok=true (whole-subtree fill), got ok=%v selected=%v", ok, names(selected))
	}
	if len(selected) != 4 {
		t.Fatalf("want 4 victims (only the holders), got %v", names(selected))
	}
}

// TestSelectEvictHardKGroupComposition k groups of 2 chips, hard req=6: canFitHardCheck
// accepts the whole-group composition class (req = 3*2 across 3 of the 4 siblings), so
// eviction must match by fully emptying 3 groups. All 6 chips in those groups are held
// by preemptees → 6 victims.
func TestSelectEvictHardKGroupComposition(t *testing.T) {
	root := ParseTopology("[[0,1],[2,3],[4,5],[6,7]]")
	if root == nil {
		t.Fatal("parse failed")
	}
	owners := map[string][]int{
		"p0": {0}, "p1": {1}, "p2": {2}, "p3": {3},
		"p4": {4}, "p5": {5}, "p6": {6}, "p7": {7},
	}
	root.Init(nil, nil, owners)
	preemptees := []*api.TaskInfo{
		peTask(0, 0), peTask(1, 1), peTask(2, 2), peTask(3, 3),
		peTask(4, 4), peTask(5, 5), peTask(6, 6), peTask(7, 7),
	}
	selected, ok := root.SelectEvictHard(6, false, preemptees)
	if !ok {
		t.Fatalf("want ok=true (whole-group composition), got ok=%v selected=%v", ok, names(selected))
	}
	if len(selected) != 6 {
		t.Fatalf("want 6 victims (3 whole groups), got %v", names(selected))
	}
}

// TestSelectEvictHardZeroEvictionWholeFreeTree: a fully free tree satisfies req == total
// through canFitHardCheck's whole-subtree class and must require zero eviction.
func TestSelectEvictHardZeroEvictionWholeFreeTree(t *testing.T) {
	root := twoGroupTree(nil)
	selected, ok := root.SelectEvictHard(8, false, nil)
	if !ok || selected != nil {
		t.Fatalf("want (nil,true) zero eviction, got selected=%v ok=%v", names(selected), ok)
	}
}

// TestSelectEvictHardZeroEvictionWholeGroupsFree: two of four 2-chip groups are already
// free, req=4 = 2*2, so canFitHardCheck's whole-group composition class is FitNormal —
// zero eviction without touching the occupied groups.
func TestSelectEvictHardZeroEvictionWholeGroupsFree(t *testing.T) {
	root := ParseTopology("[[0,1],[2,3],[4,5],[6,7]]")
	owners := map[string][]int{"p4": {4}, "p5": {5}, "p6": {6}, "p7": {7}}
	root.Init(nil, nil, owners)
	selected, ok := root.SelectEvictHard(4, false, nil)
	if !ok || selected != nil {
		t.Fatalf("want (nil,true) zero eviction, got selected=%v ok=%v", names(selected), ok)
	}
}

// TestSelectEvictHardFewestVictimsInScope: victims are only those holding chips inside
// the chosen target scope, and the target needing fewest victims is preferred — group1
// needs only its 2 holders (pod-2), while pod-0/pod-1 holding group0 would need 2; chips
// outside the chosen scope are never counted as reclaimed.
func TestSelectEvictHardFewestVictimsInScope(t *testing.T) {
	owners := map[string][]int{"p0": {0, 1}, "p1": {2}, "p2": {4, 5}}
	root := twoGroupTree(owners)
	preemptees := []*api.TaskInfo{peTask(0, 0, 1), peTask(1, 2), peTask(2, 4, 5)}
	selected, ok := root.SelectEvictHard(4, false, preemptees)
	if !ok {
		t.Fatalf("want ok=true")
	}
	if len(selected) != 1 || selected[0].Name != "pod-2" {
		t.Fatalf("want single victim pod-2 (group1 needs only its 2 holders), got %v", names(selected))
	}
}

func TestSelectEvictNilReceiver(t *testing.T) {
	var root *ChipNode
	selected, ok := root.SelectEvictSoft(2, false, nil)
	if ok || selected != nil {
		t.Fatalf("SelectEvictSoft(nil): want (nil,false), got selected=%v ok=%v", names(selected), ok)
	}
	selected, ok = root.SelectEvictHard(2, false, nil)
	if ok || selected != nil {
		t.Fatalf("SelectEvictHard(nil): want (nil,false), got selected=%v ok=%v", names(selected), ok)
	}
}
