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
	"sort"

	"k8s.io/klog/v2"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// SelectEvictSoft returns the smallest prefix of preemptees (in framework order)
// whose eviction leaves the tree with at least req free chips — zero eviction
// when it already has req usable chips, abstain (nil, false) when even evicting
// every preemptee cannot free req chips.
func (n *ChipNode) SelectEvictSoft(req int, allow bool, preemptees []*api.TaskInfo) ([]*api.TaskInfo, bool) {
	if n == nil {
		return nil, false
	}
	if n.CountUsable(allow) >= req {
		klog.V(util.LogInfoLev).Infof("SelectEvictSoft: usable=%d>=req=%d, zero eviction",
			n.CountUsable(allow), req)
		return nil, true
	}
	reclaim := make(map[int]struct{})
	selected := make([]*api.TaskInfo, 0, len(preemptees))
	for _, pe := range preemptees {
		if pe == nil || pe.Pod == nil {
			continue
		}
		chips := util.GetAllocatedChipIDsFromPod(pe.Pod)
		if len(chips) == 0 {
			continue
		}
		for _, id := range chips {
			reclaim[id] = struct{}{}
		}
		selected = append(selected, pe)
		if n.CountEvictSet(allow, reclaim) >= req {
			klog.V(util.LogInfoLev).Infof("SelectEvictSoft: req<%d> satisfied with %d victims",
				req, len(selected))
			return selected, true
		}
	}
	klog.V(util.LogInfoLev).Infof("SelectEvictSoft: req<%d> not satisfiable with %d victims, abstain",
		req, len(selected))
	return nil, false
}

// SelectEvictHard keeps its eviction feasibility aligned with canFitHardCheck:
// the request is investigated across the same placement classes canFitHardCheck
// accepts — one leaf group dense-fit, a whole subtree of size req fully emptied
// (whole-subtree fill), or k sibling subtrees of size req/k fully emptied
// (whole-group composition). Zero eviction uses canFitHardCheck == FitNormal
// directly, so every placement that already fits (including k whole groups or a
// whole free subtree) passes without victims. Victims are always chosen from the
// preemptees, only pods holding chips inside the target scope are evicted, and
// the target needing the fewest victims is preferred.
func (n *ChipNode) SelectEvictHard(req int, allow bool, preemptees []*api.TaskInfo) ([]*api.TaskInfo, bool) {
	if n == nil {
		return nil, false
	}
	if n.CanFitHardFor(req, allow, nil) == FitNormal {
		klog.V(util.LogInfoLev).Infof("SelectEvictHard: req<%d> already hard-fittable, zero eviction", req)
		return nil, true
	}
	var nodes []*ChipNode
	n.CollectNodes(&nodes)
	var best []*api.TaskInfo
	bestCount := -1
	consider := func(units []*ChipNode, need int, full bool) {
		sel, ok := evictVictimsFor(units, need, full, allow, preemptees)
		if ok && (bestCount == -1 || len(sel) < bestCount) {
			bestCount = len(sel)
			best = sel
		}
	}
	// (a) single leaf-group dense fit.
	for _, g := range n.LeafGroups(allow) {
		if g.Total() >= req {
			consider([]*ChipNode{g}, req, false)
		}
	}
	// (b) whole-subtree fill: some intermediate node has exactly req chips.
	for _, nd := range nodes {
		children := nd.Children()
		if len(children) == 0 || len(children[0].Children()) == 0 {
			continue // leaf / leaf group, covered by (a)
		}
		if nd.Total() == req {
			consider([]*ChipNode{nd}, req, true)
		}
	}
	// (c) whole-group composition: req == k*size of sibling subtrees.
	for _, nd := range nodes {
		children := nd.Children()
		if len(children) < util.NPUIndex2 {
			continue
		}
		if len(children[0].Children()) == 0 {
			continue // leaf group, covered by (a)
		}
		g := children[0].Total()
		if g == 0 || req%g != 0 {
			continue
		}
		k := req / g
		if k < util.NPUIndex2 || k >= len(children) {
			continue
		}
		if units := wholeGroupTarget(children, k, allow, preemptees); units != nil {
			consider(units, req, true)
		}
	}
	if best != nil {
		klog.V(util.LogInfoLev).Infof("SelectEvictHard: req<%d> satisfied after %d victims", req, bestCount)
		return best, true
	}
	klog.V(util.LogInfoLev).Infof("SelectEvictHard: req<%d> not satisfiable after full eviction, abstain", req)
	return nil, false
}

// evictVictimsFor returns the smallest prefix of preemptees (in framework
// order) whose eviction frees the target: `need` chips across the units, or the
// whole units when full. Only pods holding at least one chip inside the target
// scope are considered; chips outside the scope are ignored.
func evictVictimsFor(units []*ChipNode, need int, full, allow bool,
	preemptees []*api.TaskInfo) ([]*api.TaskInfo, bool) {
	scope := make(map[int]struct{})
	for _, u := range units {
		for id := range u.ChipIDs() {
			scope[id] = struct{}{}
		}
	}
	reclaim := make(map[int]struct{})
	selected := make([]*api.TaskInfo, 0, len(preemptees))
	for _, pe := range preemptees {
		if pe == nil || pe.Pod == nil {
			continue
		}
		touched := false
		for _, id := range util.GetAllocatedChipIDsFromPod(pe.Pod) {
			if _, ok := scope[id]; ok {
				reclaim[id] = struct{}{}
				touched = true
			}
		}
		if !touched {
			continue
		}
		selected = append(selected, pe)
		freed := 0
		for _, u := range units {
			freed += u.CountEvictSet(allow, reclaim)
		}
		if full {
			if freed == need { // disjoint units: freed == need means all free
				return selected, true
			}
		} else if freed >= need {
			return selected, true
		}
	}
	return nil, false
}

// wholeGroupTarget picks the k sibling subtrees that need the fewest victims to
// fully empty, for the whole-group composition placement (canFitHardCheck class
// (c)). Returns nil when fewer than k siblings can be fully emptied.
func wholeGroupTarget(children []*ChipNode, k int, allow bool,
	preemptees []*api.TaskInfo) []*ChipNode {
	type cand struct {
		node *ChipNode
		cost int
	}
	cands := make([]cand, 0, len(children))
	for _, c := range children {
		if sel, ok := evictVictimsFor([]*ChipNode{c}, c.Total(), true, allow, preemptees); ok {
			cands = append(cands, cand{c, len(sel)})
		}
	}
	if len(cands) < k {
		return nil
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].cost < cands[j].cost })
	units := make([]*ChipNode, 0, k)
	for i := 0; i < k; i++ {
		units = append(units, cands[i].node)
	}
	return units
}
