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

// chipnode_stub.go is a placeholder re-implementation of the ChipNode type and
// its exported API. It exists so that the rest of the tree (chip.go, node.go,
// ...) keeps compiling while the real implementation in chipnode.go is removed
// or under development. Every body is empty and returns the zero value of its
// result type; it deliberately matches the exact exported surface of
// chipnode.go so the two files are drop-in interchangeable for dependent code.
//
// Build rule: put this file in the package only when chipnode.go is absent —
// both files must never be present at the same time (duplicate definitions).

import "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"

const maxTopoDepth = 64

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
	return nil
}

func BuildFlatTopology(count int) string {
	return ""
}

func (t *ChipNode) Init(faulty, netUnh map[int]struct{}, owners map[string][]int) {}

func (t *ChipNode) TryAllocate(taskKey string, chips []int) error {
	return nil
}

func (t *ChipNode) Rollback(taskKey string) error {
	return nil
}

func (t *ChipNode) Clone() *ChipNode {
	return nil
}

func (n *ChipNode) MaxChipID() int {
	return 0
}

type FitClass int

const (
	FitNormal FitClass = iota
	FitEvict
	FitFailed
)

// Fit mirrors topo.ChipNode.Fit. Zero value of FitClass is FitNormal.
func (t *ChipNode) Fit(req *util.Request) FitClass {
	return 0
}

// SelectChips mirrors topo.ChipNode.SelectChips.
func (n *ChipNode) SelectChips(req *util.Request) []int {
	return nil
}

// Score mirrors topo.ChipNode.Score.
func (n *ChipNode) Score(req int, allowNetUnh bool) float64 {
	return 0
}
