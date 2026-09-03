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
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// This file tests the real ChipNode implementation: topology parsing, hard/soft
// fit decisions, chip selection, occupancy tracking and scoring. Every tree used
// here is built through the exported API (ParseTopology/BuildFlatTopology) and
// Init'ed the same way the scheduler does, so the assertions pin real behavior.

const nested16Topo = "[[[0,1],[2,3],[4,5],[6,7]],[[8,9],[10,11],[12,13],[14,15]]]"

// initTopo parses raw and applies Init; tests probe behavior through the tree.
func initTopo(t *testing.T, raw string, faulty, netUnh map[int]struct{}, owners map[string][]int) *ChipNode {
	t.Helper()
	root := ParseTopology(raw)
	if root == nil {
		t.Fatalf("ParseTopology(%q) = nil", raw)
	}
	root.Init(faulty, netUnh, owners)
	return root
}

// flatTopo builds an n-chip flat topology [0..n-1] with no faults or owners.
func flatTopo(t *testing.T, n int) *ChipNode {
	t.Helper()
	return initTopo(t, BuildFlatTopology(n), nil, nil, nil)
}

func TestParseTopology(t *testing.T) {
	tests := []struct {
		raw     string
		wantNil bool
		wantID  int // MaxChipID when valid
	}{
		{"", true, 0},
		{"not-json", true, 0},
		{"{}", true, 0},
		{"[]", true, 0},
		{"[0,0]", true, 0},       // duplicate chip id
		{"[-1,0]", true, 0},      // negative chip id
		{"[0,1.5]", true, 0},     // non-integer id
		{"[0,[1,2]]", true, 0},   // mixed leaf and group
		{"[[0,1],[2]]", true, 0}, // unequal group sizes
		{"0", false, 0},
		{"[0,1,2,3]", false, 3},
		{"[[0,1],[2,3]]", false, 3},
		{nested16Topo, false, 15},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			root := ParseTopology(tt.raw)
			if tt.wantNil {
				if root != nil {
					t.Errorf("ParseTopology(%q) = %v, want nil", tt.raw, root)
				}
				return
			}
			if root == nil {
				t.Fatalf("ParseTopology(%q) = nil", tt.raw)
			}
			if root.Raw != tt.raw {
				t.Errorf("Raw = %q, want %q", root.Raw, tt.raw)
			}
			if got := root.MaxChipID(); got != tt.wantID {
				t.Errorf("MaxChipID(%q) = %d, want %d", tt.raw, got, tt.wantID)
			}
		})
	}
}

func TestBuildFlatTopology(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, ""},
		{-1, ""},
		{1, "[0]"},
		{4, "[0,1,2,3]"},
		{8, "[0,1,2,3,4,5,6,7]"},
	}
	for _, tt := range tests {
		if got := BuildFlatTopology(tt.n); got != tt.want {
			t.Errorf("BuildFlatTopology(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}

	// round-trip: a generated flat topology must parse back into the same tree
	if root := ParseTopology(BuildFlatTopology(8)); root == nil || root.MaxChipID() != 7 {
		t.Errorf("round-trip of BuildFlatTopology(8) failed: %v", root)
	}
}

func TestFitSoftMode(t *testing.T) {
	root := flatTopo(t, 6)
	req := func(num int) *util.Request {
		return &util.Request{ReqNPUNum: num, Mode: util.SoftScheduleMode}
	}
	if got := root.Fit(req(4)); got != FitNormal {
		t.Errorf("6-chip flat req 4 soft = %d, want FitNormal", got)
	}
	if got := root.Fit(req(6)); got != FitNormal {
		t.Errorf("6-chip flat req 6 soft = %d, want FitNormal", got)
	}
	if got := root.Fit(req(7)); got != FitFailed {
		t.Errorf("6-chip flat req 7 soft = %d, want FitFailed", got)
	}

	// occupy 3 chips: usable drops to 3, but all 6 remain evictable
	if err := root.TryAllocate("t", []int{0, 1, 2}); err != nil {
		t.Fatalf("TryAllocate: %v", err)
	}
	if got := root.Fit(req(3)); got != FitNormal {
		t.Errorf("req 3 with 3 busy = %d, want FitNormal (usable)", got)
	}
	if got := root.Fit(req(4)); got != FitEvict {
		t.Errorf("req 4 with 3 busy = %d, want FitEvict", got)
	}
	if got := root.Fit(req(7)); got != FitFailed {
		t.Errorf("req 7 with 3 busy = %d, want FitFailed", got)
	}
}

func TestFitHardMode(t *testing.T) {
	root := initTopo(t, nested16Topo, nil, nil, nil)
	hard := func(num int) *util.Request {
		return &util.Request{ReqNPUNum: num, Mode: util.HardScheduleMode}
	}

	// whole-group combinations: 6 = 3 groups of 2, 8 = one superpod, 16 = whole tree
	for _, num := range []int{1, 2, 4, 6, 8, 16} {
		if got := root.Fit(hard(num)); got != FitNormal {
			t.Errorf("nested16 req %d hard = %d, want FitNormal", num, got)
		}
	}
	// odd requests that no single allocation domain can satisfy
	for _, num := range []int{3, 5, 7, 9, 10, 15} {
		if got := root.Fit(hard(num)); got != FitFailed {
			t.Errorf("nested16 req %d hard = %d, want FitFailed", num, got)
		}
	}

	// a busy group must not block the whole-group request: [0,1] occupied, req 6
	// can still be satisfied by groups [2,3],[4,5],[6,7]
	if err := root.TryAllocate("t", []int{0, 1}); err != nil {
		t.Fatalf("TryAllocate: %v", err)
	}
	if got := root.Fit(hard(6)); got != FitNormal {
		t.Errorf("nested16 req 6 hard with [0,1] busy = %d, want FitNormal", got)
	}
	if got := root.Fit(hard(7)); got != FitFailed {
		t.Errorf("nested16 req 7 hard with [0,1] busy = %d, want FitFailed", got)
	}

	// occupy everything: only eviction can satisfy whole-group requests
	all := make([]int, 0, 16)
	for i := 0; i < 16; i++ {
		all = append(all, i)
	}
	full := initTopo(t, nested16Topo, nil, nil, map[string][]int{"all": all})
	if got := full.Fit(hard(6)); got != FitEvict {
		t.Errorf("nested16 req 6 hard fully busy = %d, want FitEvict", got)
	}
	if got := full.Fit(hard(17)); got != FitFailed {
		t.Errorf("nested16 req 17 hard = %d, want FitFailed", got)
	}

	// flat topology is a single allocation domain
	flat := flatTopo(t, 8)
	if got := flat.Fit(hard(6)); got != FitNormal {
		t.Errorf("8-chip flat req 6 hard = %d, want FitNormal", got)
	}
	if got := flat.Fit(hard(9)); got != FitFailed {
		t.Errorf("8-chip flat req 9 hard = %d, want FitFailed", got)
	}
}

func TestSelectChipsSoftMode(t *testing.T) {
	root := flatTopo(t, 8)
	soft := func(num int) *util.Request {
		return &util.Request{ReqNPUNum: num, Mode: util.SoftScheduleMode}
	}
	if got := root.SelectChips(soft(4)); !reflect.DeepEqual(got, []int{0, 1, 2, 3}) {
		t.Errorf("flat8 soft req 4 = %v, want [0 1 2 3]", got)
	}
	if got := root.SelectChips(soft(8)); !reflect.DeepEqual(got, []int{0, 1, 2, 3, 4, 5, 6, 7}) {
		t.Errorf("flat8 soft req 8 = %v, want full ascending", got)
	}
	if got := root.SelectChips(soft(9)); got != nil {
		t.Errorf("flat8 soft req 9 = %v, want nil", got)
	}
	if got := root.SelectChips(soft(0)); got != nil {
		t.Errorf("flat8 soft req 0 = %v, want nil", got)
	}

	nested := initTopo(t, nested16Topo, nil, nil, nil)
	if got := nested.SelectChips(soft(4)); !reflect.DeepEqual(got, []int{0, 1, 2, 3}) {
		t.Errorf("nested16 soft req 4 = %v, want [0 1 2 3]", got)
	}
	// soft mode spreads across groups and avoids a busy group
	if err := nested.TryAllocate("t", []int{0, 1}); err != nil {
		t.Fatalf("TryAllocate: %v", err)
	}
	if got := nested.SelectChips(soft(2)); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Errorf("nested16 soft req 2 with [0,1] busy = %v, want [2 3]", got)
	}
}

func TestSelectChipsHardMode(t *testing.T) {
	root := initTopo(t, nested16Topo, nil, nil, nil)
	hard := func(num int) *util.Request {
		return &util.Request{ReqNPUNum: num, Mode: util.HardScheduleMode}
	}
	if got := root.SelectChips(hard(6)); !reflect.DeepEqual(got, []int{0, 1, 2, 3, 4, 5}) {
		t.Errorf("nested16 hard req 6 = %v, want [0 1 2 3 4 5]", got)
	}
	if got := root.SelectChips(hard(8)); !reflect.DeepEqual(got, []int{0, 1, 2, 3, 4, 5, 6, 7}) {
		t.Errorf("nested16 hard req 8 = %v, want [0..7]", got)
	}
	if got := root.SelectChips(hard(16)); !reflect.DeepEqual(got, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}) {
		t.Errorf("nested16 hard req 16 = %v, want full ascending", got)
	}
	if got := root.SelectChips(hard(9)); got != nil {
		t.Errorf("nested16 hard req 9 = %v, want nil", got)
	}

	// whole-group selection skips the busy [0,1] group
	if err := root.TryAllocate("t", []int{0, 1}); err != nil {
		t.Fatalf("TryAllocate: %v", err)
	}
	if got := root.SelectChips(hard(6)); !reflect.DeepEqual(got, []int{2, 3, 4, 5, 6, 7}) {
		t.Errorf("nested16 hard req 6 with [0,1] busy = %v, want [2 3 4 5 6 7]", got)
	}
}

func TestSelectChipsSoftPrefersHardPlacement(t *testing.T) {
	soft := func(num int) *util.Request {
		return &util.Request{ReqNPUNum: num, Mode: util.SoftScheduleMode}
	}
	hard := func(num int) *util.Request {
		return &util.Request{ReqNPUNum: num, Mode: util.HardScheduleMode}
	}

	root := initTopo(t, "[[0,1],[2,3],[4,5],[6,7]]", nil, nil, map[string][]int{"t": {1}})
	if got := root.SelectChips(soft(4)); !reflect.DeepEqual(got, []int{2, 3, 4, 5}) {
		t.Errorf("soft req 4 = %v, want whole groups [2 3 4 5]", got)
	}
	// strict placement exists, so soft now converges to the hard result
	if got := root.SelectChips(hard(4)); !reflect.DeepEqual(got, []int{2, 3, 4, 5}) {
		t.Errorf("hard req 4 = %v, want [2 3 4 5]", got)
	}

	// same whole-group preference through a deeper tree (super-pod level)
	nested := initTopo(t, nested16Topo, nil, nil, map[string][]int{"t": {1}})
	if got := nested.SelectChips(soft(4)); !reflect.DeepEqual(got, []int{2, 3, 4, 5}) {
		t.Errorf("nested soft req 4 = %v, want [2 3 4 5]", got)
	}
}

func TestSelectChipsSoftFallsBackToSpread(t *testing.T) {
	// req 5 cannot be placed strictly (no leaf domain holds 5, and 5 is not a
	// whole-group multiple), so soft mode relaxes to cross-group spreading
	// while hard mode refuses to place it at all.
	soft := func(num int) *util.Request {
		return &util.Request{ReqNPUNum: num, Mode: util.SoftScheduleMode}
	}
	root := initTopo(t, "[[0,1],[2,3],[4,5],[6,7]]", nil, nil, map[string][]int{"t": {1}})
	if got := root.SelectChips(soft(5)); !reflect.DeepEqual(got, []int{0, 2, 3, 4, 5}) {
		t.Errorf("soft req 5 = %v, want spread [0 2 3 4 5]", got)
	}
	if got := root.SelectChips(&util.Request{ReqNPUNum: 5, Mode: util.HardScheduleMode}); got != nil {
		t.Errorf("hard req 5 = %v, want nil", got)
	}
}

func TestSelectChipsToleratesNetUnhealthy(t *testing.T) {
	// mark cards 5 and 6 parameter-plane unhealthy
	netUnh := map[int]struct{}{5: {}, 6: {}}
	root := initTopo(t, BuildFlatTopology(8), nil, netUnh, nil)

	// without tolerance the unhealthy cards are unusable -> pick the healthy six
	noTol := root.SelectChips(&util.Request{ReqNPUNum: 6, Mode: util.SoftScheduleMode})
	if !reflect.DeepEqual(noTol, []int{0, 1, 2, 3, 4, 7}) {
		t.Errorf("no-tolerance req 6 = %v, want [0 1 2 3 4 7]", noTol)
	}
	// with tolerance the unhealthy cards are used first (fault concentration)
	withTol := root.SelectChips(&util.Request{ReqNPUNum: 2, Mode: util.SoftScheduleMode, AllowNetUnhealthy: true})
	if !reflect.DeepEqual(withTol, []int{5, 6}) {
		t.Errorf("tolerant req 2 = %v, want [5 6]", withTol)
	}
	if got := root.Fit(&util.Request{ReqNPUNum: 8, Mode: util.SoftScheduleMode, AllowNetUnhealthy: true}); got != FitNormal {
		t.Errorf("tolerant req 8 = %d, want FitNormal", got)
	}
	if got := root.Fit(&util.Request{ReqNPUNum: 8, Mode: util.SoftScheduleMode}); got != FitFailed {
		t.Errorf("no-tolerance req 8 = %d, want FitFailed", got)
	}
}

func TestInitMarksFaulty(t *testing.T) {
	root := initTopo(t, BuildFlatTopology(8), map[int]struct{}{0: {}, 1: {}}, nil, nil)
	if got := root.Fit(&util.Request{ReqNPUNum: 4, Mode: util.SoftScheduleMode}); got != FitNormal {
		t.Errorf("4 req on 6 healthy = %d, want FitNormal", got)
	}
	if got := root.Fit(&util.Request{ReqNPUNum: 7, Mode: util.SoftScheduleMode}); got != FitFailed {
		t.Errorf("7 req on 6 healthy = %d, want FitFailed", got)
	}
	// faulty chips can never be allocated
	if err := root.TryAllocate("x", []int{0}); err == nil {
		t.Error("TryAllocate on faulty chip 0: want error")
	}
	got := root.SelectChips(&util.Request{ReqNPUNum: 1, Mode: util.SoftScheduleMode})
	if !reflect.DeepEqual(got, []int{2}) {
		t.Errorf("first selectable chip = %v, want [2]", got)
	}
}

func TestTryAllocateAndRollback(t *testing.T) {
	root := flatTopo(t, 8)

	if err := root.TryAllocate("t1", []int{0, 1}); err != nil {
		t.Fatalf("TryAllocate t1: %v", err)
	}
	// same task cannot allocate twice
	if err := root.TryAllocate("t1", []int{2}); err == nil {
		t.Error("duplicate task TryAllocate: want error")
	}
	// a busy chip cannot be claimed by another task
	if err := root.TryAllocate("t2", []int{1}); err == nil {
		t.Error("TryAllocate on busy chip 1: want error")
	}
	// duplicate chip ids in one request roll the whole request back
	if err := root.TryAllocate("t3", []int{2, 2}); err == nil {
		t.Error("TryAllocate duplicate id [2,2]: want error")
	}

	if err := root.Rollback("t1"); err != nil {
		t.Fatalf("Rollback t1: %v", err)
	}
	// rolling back twice is an error, unknown tasks too
	if err := root.Rollback("t1"); err == nil {
		t.Error("second Rollback t1: want error")
	}
	if err := root.Rollback("nobody"); err == nil {
		t.Error("Rollback unknown task: want error")
	}
	// the freed chips are reusable
	if err := root.TryAllocate("t4", []int{0, 1}); err != nil {
		t.Errorf("reuse after rollback: %v", err)
	}

	// occupancy affects fit: 6 of 8 busy leaves only 2 usable
	busy := initTopo(t, BuildFlatTopology(8), nil, nil, map[string][]int{"t": {0, 1, 2, 3, 4, 5}})
	if got := busy.Fit(&util.Request{ReqNPUNum: 3, Mode: util.SoftScheduleMode}); got != FitEvict {
		t.Errorf("6 busy req 3 soft = %d, want FitEvict", got)
	}
	if got := busy.Fit(&util.Request{ReqNPUNum: 2, Mode: util.SoftScheduleMode}); got != FitNormal {
		t.Errorf("6 busy req 2 soft = %d, want FitNormal", got)
	}
}

func TestMaxChipID(t *testing.T) {
	tests := []struct {
		raw  string
		want int
	}{
		{"0", 0},
		{"[0,1,2,3]", 3},
		{"[[0,1],[2,3]]", 3},
		{nested16Topo, 15},
	}
	for _, tt := range tests {
		root := ParseTopology(tt.raw)
		if got := root.MaxChipID(); got != tt.want {
			t.Errorf("MaxChipID(%q) = %d, want %d", tt.raw, got, tt.want)
		}
	}
}

func TestScore(t *testing.T) {
	reports := []struct {
		name string
		tree *ChipNode
		req  int
		want float64
	}{
		{"flat fit", flatTopo(t, 8), 4, affinityTopScore},
		{"flat fit all", flatTopo(t, 8), 8, affinityTopScore},
		{"nested whole-group", initTopo(t, nested16Topo, nil, nil, nil), 6, affinityTopScore},
		{"nested whole superpod", initTopo(t, nested16Topo, nil, nil, nil), 8, affinityTopScore},
		{"nested odd req can't fit hard", initTopo(t, nested16Topo, nil, nil, nil), 7, 0},
		{"flat too small", flatTopo(t, 2), 8, 0},
	}
	for _, tt := range reports {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tree.Score(tt.req, false); got != tt.want {
				t.Errorf("Score(%d) = %v, want %v", tt.req, got, tt.want)
			}
		})
	}
	if got := (*ChipNode)(nil).Score(4, false); got != 0 {
		t.Errorf("nil Score = %v, want 0", got)
	}
}

func TestCloneIsIndependent(t *testing.T) {
	root := initTopo(t, nested16Topo, nil, nil, map[string][]int{"t": {0, 1}})
	clone := root.Clone()
	if clone == nil {
		t.Fatal("Clone() = nil")
	}
	if clone.MaxChipID() != root.MaxChipID() || clone.Raw != root.Raw {
		t.Errorf("clone structure diverged: maxID %d/%d raw %q/%q",
			clone.MaxChipID(), root.MaxChipID(), clone.Raw, root.Raw)
	}

	// mutating the clone must not leak into the original's occupancy
	if err := clone.Rollback("t"); err != nil {
		t.Fatalf("clone Rollback: %v", err)
	}
	if err := clone.TryAllocate("c2", []int{0, 1}); err != nil {
		t.Errorf("clone reuse after rollback: %v", err)
	}
	// original still holds [0,1] for "t", so a new owner cannot claim chip 0
	if err := root.TryAllocate("o2", []int{0}); err == nil {
		t.Error("original chip 0 should still be busy after clone mutation")
	}

	if got := (*ChipNode)(nil).Clone(); got != nil {
		t.Errorf("nil Clone = %v, want nil", got)
	}
}

func TestNilReceiverGuards(t *testing.T) {
	var nilNode *ChipNode
	if got := nilNode.Fit(&util.Request{ReqNPUNum: 4, Mode: util.SoftScheduleMode}); got != FitFailed {
		t.Errorf("nil Fit = %d, want FitFailed", got)
	}
	if got := nilNode.SelectChips(&util.Request{ReqNPUNum: 4, Mode: util.SoftScheduleMode}); got != nil {
		t.Errorf("nil SelectChips = %v, want nil", got)
	}
	if got := (*ChipNode)(nil).Score(4, false); got != 0 {
		t.Errorf("nil Score = %v, want 0", got)
	}
}
