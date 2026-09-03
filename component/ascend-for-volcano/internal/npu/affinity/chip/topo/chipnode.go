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
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"k8s.io/klog/v2"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

const maxTopoDepth = 64

var errNotInSubtree = fmt.Errorf("chip not in subtree")

// ChipNode node of chips tree
type ChipNode struct {
	children     []*ChipNode
	chipID       int
	total        int
	Raw          string
	faulty       int
	netUnhealthy int
	allocated    int
	ownedBy      []string

	owners    map[string][]int
	leafIndex map[int]*ChipNode
}

func ParseTopology(raw string) *ChipNode {
	var data interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil
	}
	root, ok := buildNode(data, 0, make(map[int]struct{}))
	if !ok {
		return nil
	}
	root.Raw = raw
	return root
}

func BuildFlatTopology(count int) string {
	if count < 1 {
		return ""
	}
	var b strings.Builder
	b.Grow(count*util.NPUIndex2 + 1)
	b.WriteByte('[')
	for i := 0; i < count; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(i))
	}
	b.WriteByte(']')
	return b.String()
}

func buildNode(data interface{}, depth int, seen map[int]struct{}) (*ChipNode, bool) {
	if depth > maxTopoDepth {
		return nil, false
	}
	switch v := data.(type) {
	case float64:
		id := int(v)
		if float64(id) != v || id < 0 {
			return nil, false
		}
		if _, dup := seen[id]; dup {
			return nil, false
		}
		seen[id] = struct{}{}
		return &ChipNode{chipID: id, total: 1}, true
	case []interface{}:
		if len(v) == 0 {
			return nil, false
		}
		allChips, allGroups := true, true
		for _, item := range v {
			switch item.(type) {
			case float64:
				allGroups = false
			case []interface{}:
				allChips = false
			default:
				return nil, false
			}
		}
		if !allChips && !allGroups {
			return nil, false
		}
		if allGroups {
			sz := len(v[0].([]interface{}))
			for i := 1; i < len(v); i++ {
				if len(v[i].([]interface{})) != sz {
					return nil, false
				}
			}
		}
		group := &ChipNode{}
		for _, item := range v {
			child, ok := buildNode(item, depth+1, seen)
			if !ok {
				return nil, false
			}
			group.children = append(group.children, child)
			group.total += child.total
		}
		return group, true
	default:
		return nil, false
	}
}

func (n *ChipNode) collectLeaves(out map[int]*ChipNode) {
	if len(n.children) == 0 {
		out[n.chipID] = n
		return
	}
	for _, c := range n.children {
		c.collectLeaves(out)
	}
}

func (n *ChipNode) Init(faulty, netUnh map[int]struct{}, owners map[string][]int) {
	n.owners = owners
	if n.owners == nil {
		n.owners = make(map[string][]int)
	}
	n.leafIndex = make(map[int]*ChipNode, n.total)
	n.collectLeaves(n.leafIndex)
	for _, leaf := range n.leafIndex {
		leaf.faulty, leaf.netUnhealthy, leaf.allocated, leaf.ownedBy = 0, 0, 0, nil
		id := leaf.chipID
		switch {
		case has(faulty, id):
			leaf.faulty = 1
		case has(netUnh, id):
			leaf.netUnhealthy = 1
		}
	}
	keys := make([]string, 0, len(owners))
	for k := range owners {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, id := range owners[k] {
			if err := n.occupyRec(id, k); err != nil {
				klog.V(util.LogWarningLev).Infof("Init: task %s occupy chip %d: %v", k, id, err)
			}
		}
	}
}

func (n *ChipNode) TryAllocate(taskKey string, chips []int) error {
	if _, dup := n.owners[taskKey]; dup {
		return fmt.Errorf("task %s already allocated", taskKey)
	}
	occupied := make([]int, 0, len(chips))
	for _, id := range chips {
		if err := n.occupyRec(id, taskKey); err != nil {
			for _, oid := range occupied {
				if releaseErr := n.releaseRec(oid, taskKey); releaseErr != nil {
					klog.V(util.LogWarningLev).Infof("release chip %d failed, err: %v", oid, err)
				}
			}
			return err
		}
		occupied = append(occupied, id)
	}
	n.owners[taskKey] = chips
	return nil
}

func (n *ChipNode) Rollback(taskKey string) error {
	_, ok := n.owners[taskKey]
	if !ok {
		return fmt.Errorf("task %s not registered", taskKey)
	}
	n.rollbackAll(taskKey)
	delete(n.owners, taskKey)
	return nil
}

func (n *ChipNode) rollbackAll(taskKey string) {
	chips, ok := n.owners[taskKey]
	if !ok {
		return
	}
	for _, id := range chips {
		if err := n.releaseRec(id, taskKey); err != nil {
			klog.V(util.LogWarningLev).Infof("release chip %d failed, err: %v", id, err)
		}
	}
}

func (n *ChipNode) occupyRec(chipID int, taskKey string) error {
	if len(n.children) == 0 {
		if n.chipID != chipID {
			return errNotInSubtree
		}
		if n.faulty == 1 || len(n.ownedBy) > 0 {
			return fmt.Errorf("chip %d not free (faulty=%d)", chipID, n.faulty)
		}
		n.ownedBy = append(n.ownedBy, taskKey)
		n.allocated = 1
		return nil
	}
	for _, c := range n.children {
		if err := c.occupyRec(chipID, taskKey); err == nil {
			return nil
		} else if err != errNotInSubtree {
			return err
		}
	}
	return errNotInSubtree
}

func (n *ChipNode) releaseRec(chipID int, taskKey string) error {
	if len(n.children) == 0 {
		if n.chipID != chipID {
			return errNotInSubtree
		}
		for i, k := range n.ownedBy {
			if k == taskKey {
				n.ownedBy = append(n.ownedBy[:i], n.ownedBy[i+1:]...)
				break
			}
		}
		n.allocated = 0
		return nil
	}
	for _, c := range n.children {
		if err := c.releaseRec(chipID, taskKey); err == nil {
			return nil
		} else if err != errNotInSubtree {
			return err
		}
	}
	return errNotInSubtree
}

func (n *ChipNode) Clone() *ChipNode {
	if n == nil {
		return nil
	}
	c := n.cloneRec()
	c.owners = make(map[string][]int, len(n.owners))
	for k, v := range n.owners {
		c.owners[k] = append([]int(nil), v...)
	}
	c.leafIndex = make(map[int]*ChipNode, len(n.leafIndex))
	c.collectLeaves(c.leafIndex)
	return c
}

func (n *ChipNode) cloneRec() *ChipNode {
	if n == nil {
		return nil
	}
	cl := &ChipNode{
		chipID: n.chipID, total: n.total, Raw: n.Raw,
		faulty: n.faulty, allocated: n.allocated, netUnhealthy: n.netUnhealthy,
		ownedBy: append([]string(nil), n.ownedBy...),
	}
	for _, c := range n.children {
		cl.children = append(cl.children, c.cloneRec())
	}
	return cl
}

func has(s map[int]struct{}, id int) bool {
	_, ok := s[id]
	return ok
}

type FitClass int

const (
	FitNormal FitClass = iota
	FitEvict
	FitFailed
)

// evictLeaf reports whether a leaf can ever be used (now or after eviction):
// it is not faulty and its parameter plane is healthy unless tolerated.
func (n *ChipNode) evictLeaf(allowNetUnh bool) bool {
	return n.faulty == 0 && (n.netUnhealthy == 0 || allowNetUnh)
}

// usableLeaf reports whether a leaf is free to allocate right now.
func (n *ChipNode) usableLeaf(allowNetUnh bool) bool {
	return n.allocated == 0 && n.evictLeaf(allowNetUnh)
}

func (n *ChipNode) CountUsable(allowNetUnh bool) int {
	if len(n.children) == 0 {
		if n.usableLeaf(allowNetUnh) {
			return 1
		}
		return 0
	}
	s := 0
	for _, c := range n.children {
		s += c.CountUsable(allowNetUnh)
	}
	return s
}

func (n *ChipNode) countEvict(allowNetUnh bool) int {
	if len(n.children) == 0 {
		if n.evictLeaf(allowNetUnh) {
			return 1
		}
		return 0
	}
	s := 0
	for _, c := range n.children {
		s += c.countEvict(allowNetUnh)
	}
	return s
}

func (n *ChipNode) CountEvictSet(allowNetUnh bool, reclaimable map[int]struct{}) int {
	if len(n.children) == 0 {
		if n.faulty == 1 {
			return 0
		}
		if n.allocated == 1 {
			if _, ok := reclaimable[n.chipID]; !ok {
				return 0
			}
		}
		if n.netUnhealthy == 0 || allowNetUnh {
			return 1
		}
		return 0
	}
	s := 0
	for _, c := range n.children {
		s += c.CountEvictSet(allowNetUnh, reclaimable)
	}
	return s
}

func (n *ChipNode) countAllocated() int {
	if len(n.children) == 0 {
		return n.allocated
	}
	s := 0
	for _, c := range n.children {
		s += c.countAllocated()
	}
	return s
}

func (n *ChipNode) MaxChipID() int {
	if len(n.children) == 0 {
		return n.chipID
	}
	max := -1
	for _, c := range n.children {
		if v := c.MaxChipID(); v > max {
			max = v
		}
	}
	return max
}

// Fit reports how well the tree can satisfy the request. Strict topology locality
// (req.Mode == hard) and parameter-plane unhealthy tolerance
// (req.AllowNetUnhealthy) are taken from the request.
func (n *ChipNode) Fit(req *util.Request) FitClass {
	if n == nil || req == nil {
		return FitFailed
	}
	reqNum := req.ReqNPUNum
	allowNetUnh := req.AllowNetUnhealthy
	if req.Mode == util.SoftScheduleMode {
		if n.CountUsable(allowNetUnh) >= reqNum {
			return FitNormal
		}
		if n.countEvict(allowNetUnh) >= reqNum {
			return FitEvict
		}
		return FitFailed
	}
	return n.canFitHardCheck(reqNum, allowNetUnh, nil)
}

func (n *ChipNode) canFitHard(req int, allowNetUnh bool) bool {
	return n.canFitHardCheck(req, allowNetUnh, nil) == FitNormal
}

func (n *ChipNode) canFitHardCheck(req int, allowNetUnh bool, reclaimable map[int]struct{}) FitClass {
	if n == nil || req <= 0 {
		return FitFailed
	}
	evict := func(x *ChipNode) int {
		if reclaimable == nil {
			return x.countEvict(allowNetUnh)
		}
		return x.CountEvictSet(allowNetUnh, reclaimable)
	}
	usable := n.CountUsable(allowNetUnh)
	// A leaf, or a group whose direct children are leaves, is a single
	// allocation domain: decide from its usable / evictable totals.
	if len(n.children) == 0 || len(n.children[0].children) == 0 {
		if usable >= req {
			return FitNormal
		}
		if evict(n) >= req {
			return FitEvict
		}
		return FitFailed
	}
	best := FitFailed
	for _, c := range n.children {
		switch c.canFitHardCheck(req, allowNetUnh, reclaimable) {
		case FitNormal:
			return FitNormal
		case FitEvict:
			best = FitEvict
		default:
		}
	}
	if req == n.total {
		if usable == n.total {
			return FitNormal
		}
		if evict(n) == n.total {
			return FitEvict
		}
	}
	if g := n.children[0].total; req%g == 0 {
		if k := req / g; k >= util.NPUIndex2 && k < len(n.children) {
			usableG, evictG := 0, 0
			for _, c := range n.children {
				if c.CountUsable(allowNetUnh) == g {
					usableG++
				}
				if evict(c) == g {
					evictG++
				}
			}
			if usableG >= k {
				return FitNormal
			}
			if evictG >= k {
				best = FitEvict
			}
		}
	}
	return best
}

// CanFitHardFor reports hard-fit feasibility after evicting the chips held by
// tenants in the reclaimable set
func (n *ChipNode) CanFitHardFor(req int, allowNetUnh bool, reclaimable map[int]struct{}) FitClass {
	return n.canFitHardCheck(req, allowNetUnh, reclaimable)
}

// Total returns the total number of leaf chips in the node (group/tree), i.e.
// its capacity. Eviction selection uses per-group capacity to decide whether a
// single group can host req.
func (n *ChipNode) Total() int {
	if n == nil {
		return 0
	}
	return n.total
}

// LeafGroups
func (n *ChipNode) LeafGroups(allowNetUnh bool) []*ChipNode {
	if n == nil {
		return []*ChipNode{}
	}
	return n.sortedLeafGroups(allowNetUnh)
}

// Children returns the direct child nodes (leaf chips or sub-groups).
func (n *ChipNode) Children() []*ChipNode {
	if n == nil {
		return nil
	}
	return n.children
}

// CollectNodes appends every node of the subtree to out in pre-order. Used by
// eviction selection to walk the intermediate groups that canFitHardCheck
// considers for whole-subtree fill / whole-group composition.
func (n *ChipNode) CollectNodes(out *[]*ChipNode) {
	if n == nil {
		return
	}
	*out = append(*out, n)
	for _, c := range n.children {
		c.CollectNodes(out)
	}
}

// ChipIDs returns all leaf chipIDs covered by the node (subtree). Used to map a
// preemptee's chips back into their leaf groups, so in-group eviction can pass
// the reclaimed chip set as the reclaimable argument to CountEvictSet.
func (n *ChipNode) ChipIDs() map[int]struct{} {
	if n == nil {
		return map[int]struct{}{}
	}
	leaves := make(map[int]*ChipNode, n.total)
	n.collectLeaves(leaves)
	res := make(map[int]struct{}, len(leaves))
	for id := range leaves {
		res[id] = struct{}{}
	}
	return res
}

// SelectChips picks req.ReqNPUNum chips for the request. Strict topology locality
// (req.Mode == hard) and parameter-plane unhealthy tolerance
// (req.AllowNetUnhealthy) are taken from the request. Soft mode still prefers
// strict topology locality first and only falls back to relaxed spreading when
// strict placement is infeasible.
func (n *ChipNode) SelectChips(req *util.Request) []int {
	if n == nil || req == nil || req.ReqNPUNum <= 0 {
		return nil
	}
	num := req.ReqNPUNum
	hard := req.Mode == util.HardScheduleMode
	allowNetUnh := req.AllowNetUnhealthy
	if n.CountUsable(allowNetUnh) < num {
		return nil
	}
	if num == n.total && n.CountUsable(allowNetUnh) == n.total {
		return n.allUsableChips(allowNetUnh)
	}
	if hard {
		if !n.canFitHard(num, allowNetUnh) {
			return nil
		}
		return n.allocHard(num, allowNetUnh)
	}
	return n.allocSoft(num, allowNetUnh)
}

func (n *ChipNode) allUsableChips(allowNetUnh bool) []int {
	res := make([]int, 0, n.CountUsable(allowNetUnh))
	n.collectUsable(&res, allowNetUnh)
	return res
}

func (n *ChipNode) collectUsable(out *[]int, allowNetUnh bool) {
	if len(n.children) == 0 {
		if n.usableLeaf(allowNetUnh) {
			*out = append(*out, n.chipID)
		}
		return
	}
	for _, c := range n.children {
		c.collectUsable(out, allowNetUnh)
	}
}

// allocSoft prefers strict topology locality: allocHard already covers any
// placement that fits densely inside one leaf domain (plus whole-group
// composition across domains), so a soft request only relaxes to cross-group
// spreading when strict placement is infeasible.
func (n *ChipNode) allocSoft(req int, allowNetUnh bool) []int {
	if got := n.allocHard(req, allowNetUnh); got != nil {
		return got
	}
	leaves := n.sortedLeafGroups(allowNetUnh)
	var res []int
	for _, leaf := range leaves {
		if len(res) >= req {
			break
		}
		res = append(res, leaf.usableChipsAsc(allowNetUnh)...)
	}
	if len(res) < req {
		return nil
	}
	return res[:req]
}

func (n *ChipNode) allocHard(req int, allowNetUnh bool) []int {
	if len(n.children[0].children) == 0 {
		us := n.usableChipsAsc(allowNetUnh)
		if len(us) < req {
			return nil
		}
		return us[:req]
	}
	cands := make([]*ChipNode, 0, len(n.children))
	for _, c := range n.children {
		if c.canFitHard(req, allowNetUnh) {
			cands = append(cands, c)
		}
	}
	sort.Slice(cands, func(i, j int) bool {
		return n.groupLess(cands[i], cands[j], allowNetUnh)
	})
	for _, c := range cands {
		if req == c.total && c.CountUsable(allowNetUnh) == c.total {
			return c.allUsableChips(allowNetUnh)
		}
		if got := c.allocHard(req, allowNetUnh); got != nil {
			return got
		}
	}
	if g := n.children[0].total; req%g == 0 {
		if k := req / g; k >= util.NPUIndex2 && k < len(n.children) {
			whole := make([]*ChipNode, 0, len(n.children))
			for _, c := range n.children {
				if c.CountUsable(allowNetUnh) == g {
					whole = append(whole, c)
				}
			}
			sort.Slice(whole, func(i, j int) bool {
				return n.groupLess(whole[i], whole[j], allowNetUnh)
			})
			if len(whole) >= k {
				res := make([]int, 0, req)
				for _, c := range whole[:k] {
					res = append(res, c.allUsableChips(allowNetUnh)...)
				}
				return res
			}
		}
	}
	return nil
}

func (n *ChipNode) sortedLeafGroups(allowNetUnh bool) []*ChipNode {
	var leaves []*ChipNode
	n.collectLeafGroups(&leaves)
	sort.Slice(leaves, func(i, j int) bool {
		return n.groupLess(leaves[i], leaves[j], allowNetUnh)
	})
	return leaves
}

// collectLeafGroups appends the nodes whose direct children are leaves
// (the groups a soft request spreads across). Tolerance does not affect the
// group set, so it is applied only when ranking the groups.
func (n *ChipNode) collectLeafGroups(out *[]*ChipNode) {
	if len(n.children) == 0 {
		return
	}
	if n.children[0].children == nil {
		*out = append(*out, n)
		return
	}
	for _, c := range n.children {
		c.collectLeafGroups(out)
	}
}

func (n *ChipNode) groupLess(a, b *ChipNode, allowNetUnh bool) bool {
	ha, hb := a.CountUsable(false), b.CountUsable(false)
	if ha != hb {
		return ha < hb
	}
	if allowNetUnh {
		na, nb := a.CountUsable(true)-ha, b.CountUsable(true)-hb
		if na != nb {
			return na > nb
		}
	}
	return a.children[0].chipID < b.children[0].chipID
}

func (n *ChipNode) collectUsableN(nh, h *[]int, allowNetUnh bool) {
	if len(n.children) == 0 {
		if !n.usableLeaf(allowNetUnh) {
			return
		}
		if n.netUnhealthy == 1 {
			*nh = append(*nh, n.chipID)
		} else {
			*h = append(*h, n.chipID)
		}
		return
	}
	for _, c := range n.children {
		c.collectUsableN(nh, h, allowNetUnh)
	}
}

func (n *ChipNode) usableChipsAsc(allowNetUnh bool) []int {
	var nh, h []int
	n.collectUsableN(&nh, &h, allowNetUnh)
	sort.Ints(nh)
	sort.Ints(h)
	return append(nh, h...)
}

const affinityTopScore = 1000.0

func (n *ChipNode) Score(req int, allowNetUnh bool) float64 {
	if n == nil || req <= 0 {
		return 0
	}
	if n.Fit(&util.Request{
		ReqNPUNum:         req,
		Mode:              util.HardScheduleMode,
		AllowNetUnhealthy: allowNetUnh,
	}) == FitNormal {
		return affinityTopScore
	}
	return 0
}
