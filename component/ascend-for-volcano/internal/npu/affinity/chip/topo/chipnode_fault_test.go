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
)

func TestChipNodeExclude(t *testing.T) {
	root := ParseTopology(BuildFlatTopology(4))
	if root == nil {
		t.Fatal("ParseTopology returned nil")
	}
	root.Init(&NodeFaultChips{}, nil, nil)

	if got := root.CountUsable(false); got != 4 {
		t.Fatalf("CountUsable before exclude = %d, want 4", got)
	}

	excluded := root.Exclude(map[int]struct{}{0: {}})
	if excluded == root {
		t.Fatal("Exclude should return a clone, not the receiver")
	}
	if got := excluded.CountUsable(false); got != 3 {
		t.Fatalf("CountUsable after exclude = %d, want 3", got)
	}
	// the original tree stays untouched
	if got := root.CountUsable(false); got != 4 {
		t.Fatalf("original CountUsable = %d, want 4", got)
	}

	// empty set returns the receiver unchanged
	if got := root.Exclude(nil); got != root {
		t.Fatal("Exclude(nil) should return the receiver")
	}
	if got := root.Exclude(map[int]struct{}{}); got != root {
		t.Fatal("Exclude(empty) should return the receiver")
	}
	var nilNode *ChipNode
	if got := nilNode.Exclude(map[int]struct{}{0: {}}); got != nil {
		t.Fatal("Exclude on nil receiver should return nil")
	}
}

func TestChipNodeSilentFaultChips(t *testing.T) {
	root := ParseTopology(BuildFlatTopology(4))
	if root == nil {
		t.Fatal("ParseTopology returned nil")
	}
	root.Init(&NodeFaultChips{Silent: map[int]struct{}{0: {}, 2: {}}}, nil, nil)

	want := map[int]struct{}{0: {}, 2: {}}
	if got := root.SilentFaultChips(); !reflect.DeepEqual(got, want) {
		t.Fatalf("SilentFaultChips() = %v, want %v", got, want)
	}

	var nilNode *ChipNode
	if got := nilNode.SilentFaultChips(); len(got) != 0 {
		t.Fatalf("SilentFaultChips on nil = %v, want empty", got)
	}
}

func TestChipNodePodErrorChips(t *testing.T) {
	root := ParseTopology(BuildFlatTopology(4))
	if root == nil {
		t.Fatal("ParseTopology returned nil")
	}
	root.Init(&NodeFaultChips{}, map[int]int64{1: 1000, 3: 2000}, nil)

	want := map[int]int64{1: 1000, 3: 2000}
	if got := root.PodErrorChips(); !reflect.DeepEqual(got, want) {
		t.Fatalf("PodErrorChips() = %v, want %v", got, want)
	}

	var nilNode *ChipNode
	if got := nilNode.PodErrorChips(); len(got) != 0 {
		t.Fatalf("PodErrorChips on nil = %v, want empty", got)
	}
}

func TestChipNodeSetPodErrorTimes(t *testing.T) {
	root := ParseTopology(BuildFlatTopology(4))
	if root == nil {
		t.Fatal("ParseTopology returned nil")
	}
	root.Init(&NodeFaultChips{}, map[int]int64{1: 1000}, nil)

	root.SetPodErrorTimes(map[int]int64{0: 500, 2: 0})
	want := map[int]int64{0: 500}
	if got := root.PodErrorChips(); !reflect.DeepEqual(got, want) {
		t.Fatalf("PodErrorChips after SetPodErrorTimes = %v, want %v", got, want)
	}
}
