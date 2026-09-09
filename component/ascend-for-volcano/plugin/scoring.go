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

package plugin

// scoring.go implements the scoring framework as plugins, serving only policies that
// implement ScoreFrameworkAware (e.g. chip-affinity); other policies keep the legacy
// factory.go path, which is not equivalent.
// Each dimension is an independent ScorePlugin; its registered weight is the dimension's bit
// offset (segment start). batchNodeOrderByFramework walks the registry in order, ORs each
// dimension's segment value << weight into uint16 bits, then ×ScoreWeight (uniform positive
// scaling, preserves order). Bit layout: common/util/score_weights.go.
// Dimensions (high bit wins, strict lexicographic priority):
//   previousNode (bit12-11): P1=2<<11 / P2=1<<11 category tiers, 0 none. Candidates are
//     partitioned into selfNode (this rank's last landing) / peerNodes (other ranks' landings) /
//     otherNodes (rest), and every node in a category carries that category's tier:
//     non-fault-pod selfNode→P1 (back-to-original), otherNodes→P2, peerNodes→0;
//     fault-pod swaps self/other: otherNodes→P1 (leave the fault first), selfNode→P2 fallback,
//     peerNodes→0. No arbitrary best-node pick inside a category — topology tie-breaks among
//     equals. PreferPreviousNode=false collapses to binary: previous-fault task landing
//     → 0, other normal → 1; the landing set is read from the fault snapshot
//     (FaultHandle.ScorePreviousFaultNodes, same source as the legacy
//     ScoreBestNPUNodes), not from PrefNodeMap — the rescheduler-only scenario
//     works without the previous-node cache;
//   topology (bit10): FitNormal=1 / evict-only=0 (ScoreBestNPUNodes values passed through as-is;
//     chip-affinity already emits {1,0});
//   subHealth (bit9-8): binary {1,0} — predate 1 (healthy); FaultHandle.ScoreSubHealthGrade
//     writes 0 in place for any switch/card sub-healthy node (bit position unchanged);
//   chipCount (bit7-0): (req/free)×255 availability; free<req (out of range) contributes 0 to
//     prevent cross-segment pollution.
// prev/subHealth are independent of topo (do not call ScoreBestNPUNodes); the topo frame is
// carried only by the topology dimension, whose top-level error aborts synthesis like any
// ordinary plugin error.

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// formatScoreMap renders a score map as a node-name-sorted string so multi-plugin logs align
// per node (Go map iteration is unordered). No precision conversion; uses %v directly.
func formatScoreMap(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%v ", k, m[k])
	}
	return strings.TrimRight(b.String(), " ")
}

// NodeScorePlugin node score plugin interface
type NodeScorePlugin interface {
	Name() string
	Score(task *api.TaskInfo, nodes []*api.NodeInfo, vcJob SchedulerJob) (map[string]int, error)
}

// ScorePluginWithWeight score plugin with bit weight
type ScorePluginWithWeight struct {
	plugin NodeScorePlugin
	weight float64
}

type scorePluginFunc struct {
	name    string
	scoreFn func(task *api.TaskInfo, nodes []*api.NodeInfo, vcJob SchedulerJob) (map[string]int, error)
}

// Name name of score plugin
func (p *scorePluginFunc) Name() string { return p.name }

// Score implement of NodeScorePlugin
func (p *scorePluginFunc) Score(task *api.TaskInfo, nodes []*api.NodeInfo, vcJob SchedulerJob) (map[string]int, error) {
	return p.scoreFn(task, nodes, vcJob)
}

// ScoreFrameworkAware score-framework interface
type ScoreFrameworkAware interface {
	ScoreFrameworkAware() bool
}

func (sHandle *ScheduleHandler) batchNodeOrderByFramework(task *api.TaskInfo,
	nodes []*api.NodeInfo, vcJob SchedulerJob) (map[string]float64, error) {
	if sHandle == nil || task == nil || len(nodes) == 0 {
		klog.V(util.LogDebugLev).Infof("batchNodeOrderByFramework failed: %s.", util.ArgumentError)
		return nil, errors.New(util.ArgumentError)
	}
	bits := make(map[string]uint16, len(nodes))
	for _, p := range sHandle.scorePlugins {
		dim, err := p.plugin.Score(task, nodes, vcJob)
		if err != nil {
			return nil, err
		}
		klog.V(util.LogDebugLev).Infof("framework batchNodeOrder task:%s plugin[%s] scores=%v",
			task.Name, p.plugin.Name(), formatScoreMap(dim))
		shift := uint16(p.weight)
		for _, n := range nodes {
			bits[n.Name] |= uint16(dim[n.Name]) << shift
		}
	}
	return toFloatScore(bits, sHandle.ScoreWeight), nil
}

// toFloatScore converts the per-node uint16 bitfield to float64 and ×ScoreWeight (uniform
// positive scaling, preserves the relative lexicographic order).
func toFloatScore(bits map[string]uint16, weight float64) map[string]float64 {
	out := make(map[string]float64, len(bits))
	for nodeName, b := range bits {
		out[nodeName] = float64(b) * weight
	}
	return out
}

// uniformScoreMap builds a base frame keyed by the candidate list with every node set to v
// (prev's read-only partition frame / subHealth's predate default frame).
func uniformScoreMap(nodes []*api.NodeInfo, v float64) map[string]float64 {
	out := initScoreMap(nodes)
	for nodeName := range out {
		out[nodeName] = v
	}
	return out
}

// initScoreMapInt zero frame keyed by the candidate list (int dimension output; dimensions
// emit integer segment values — guard paths and no-boost fallbacks return this frame).
func initScoreMapInt(nodes []*api.NodeInfo) map[string]int {
	out := make(map[string]int, len(nodes))
	for _, n := range nodes {
		if n != nil {
			out[n.Name] = 0
		}
	}
	return out
}

// toIntScoreMap converts a float64 dimension frame to the int segment value at the dimension
// boundary. math.Round mirrors the scoreChipCount (req/free)×255 convention so a borderline
// fractional topo/health value rounds deterministically instead of truncating.
func toIntScoreMap(m map[string]float64) map[string]int {
	out := make(map[string]int, len(m))
	for nodeName, v := range m {
		out[nodeName] = int(math.Round(v))
	}
	return out
}

func (sHandle *ScheduleHandler) scoreTopology(task *api.TaskInfo, nodes []*api.NodeInfo,
	vcJob SchedulerJob) (map[string]int, error) {
	out := initScoreMap(nodes)
	if err := vcJob.policyHandler.ScoreBestNPUNodes(task, nodes, out); err != nil {
		return nil, err
	}
	return toIntScoreMap(out), nil
}

func (sHandle *ScheduleHandler) scorePreviousNode(task *api.TaskInfo, nodes []*api.NodeInfo,
	vcJob SchedulerJob) (map[string]int, error) {
	// Every candidate starts at 0; all guard paths below keep that zero frame (no boost).
	zero := initScoreMapInt(nodes)

	if vcJob.IsSuperPodJob() || vcJob.IsMultiLevelJob() || vcJob.IsJobHasTorAffinityLabel() {
		return zero, nil
	}
	if !sHandle.FrameAttr.PreferPreviousNode {
		frame := uniformScoreMap(nodes, 1)
		if task != nil && sHandle.FaultHandle != nil {
			sHandle.FaultHandle.ScorePreviousFaultNodes(task, frame)
		}
		return toIntScoreMap(frame), nil
	}
	if task == nil || len(nodes) == 0 {
		return zero, nil
	}
	rankIndex := sHandle.resolveRankIndex(task, vcJob)
	if rankIndex == "" {
		return zero, nil
	}
	prefMap := vcJob.PrefNodeMap
	if prefMap == nil || len(prefMap) == 0 {
		return zero, nil
	}
	myRank, err := strconv.Atoi(rankIndex)
	if err != nil {
		return zero, nil
	}

	// Partition the candidate list (delta's key set) into selfNode (this rank's last landing;
	// ∈candidates ⟺ key exists) / peerNodes (other ranks' landings) / otherNodes (rest).
	// Every node in a category carries that category's tier — no arbitrary best-node pick
	// (over a uniform frame nodeWithMaxScore would be map-order random); node topo quality is
	// carried by the lower-priority topology bits. The float frame is shared with the legacy
	// path (categorizeNodes reads float64); the int segment is produced at the boundary.
	delta := initScoreMap(nodes)
	cat := categorizeNodes(delta, prefMap, myRank)

	if sHandle.isFaultPod(task, vcJob) {
		// fault-pod: otherNodes → P1(2) (leave the fault first), selfNode → P2(1) fallback,
		// peerNodes → 0.
		for nodeName := range cat.otherNodes {
			delta[nodeName] = 2
		}
		if cat.selfNode != "" {
			if _, exists := delta[cat.selfNode]; exists {
				delta[cat.selfNode] = 1
			}
		}
		return toIntScoreMap(delta), nil
	}
	// non-fault-pod: selfNode → P1(2) (back-to-original), otherNodes → P2(1), peerNodes → 0.
	if cat.selfNode != "" {
		if _, exists := delta[cat.selfNode]; exists {
			delta[cat.selfNode] = 2
		}
	}
	for nodeName := range cat.otherNodes {
		delta[nodeName] = 1
	}
	return toIntScoreMap(delta), nil
}

func (sHandle *ScheduleHandler) scoreSubHealth(task *api.TaskInfo, nodes []*api.NodeInfo,
	vcJob SchedulerJob) (map[string]int, error) {
	// Predate all 1 (healthy); FaultHandle.ScoreSubHealthGrade writes 0 in place for any
	// switch/card sub-healthy node, keeping binary {1,0}: healthy = 1, any sub-health = 0.
	out := uniformScoreMap(nodes, 1)
	if sHandle.FaultHandle != nil {
		sHandle.FaultHandle.ScoreSubHealthGrade(out)
	}
	return toIntScoreMap(out), nil
}

func (sHandle *ScheduleHandler) scoreChipCount(task *api.TaskInfo, nodes []*api.NodeInfo,
	vcJob SchedulerJob) (map[string]int, error) {
	delta := make(map[string]int, len(nodes))
	if task == nil || vcJob.NPUJob == nil {
		return delta, nil
	}
	vcTask, ok := vcJob.NPUJob.Tasks[task.UID]
	if !ok || vcTask.ReqNPUNum <= 0 {
		return delta, nil
	}
	npuResourceName := v1.ResourceName(vcTask.ReqNPUName)
	for _, n := range nodes {
		vcNode, nodeOK := sHandle.Nodes[n.Name]
		if !nodeOK {
			continue
		}
		free, _, _ := vcNode.GetChipCount(npuResourceName)
		if free < vcTask.ReqNPUNum {
			// free<req is out of range (candidates satisfy FitNormal ⇒ free≥req): contribute 0 so
			// (req/free)×255 cannot wrap past the low 8 bits and pollute the subHealth
			// (bit9-8)/topo (bit10) high segments.
			continue
		}
		delta[n.Name] = int(math.Round(float64(vcTask.ReqNPUNum) / float64(free) * 255.0))
	}
	return delta, nil
}

// InitScorePlugins assembles the scoring-framework registry (exported entry, called once by
// HandlerStart; the unexported namesake holds the implementation).
func (sHandle *ScheduleHandler) InitScorePlugins() {
	sHandle.initScorePlugins()
}

// initScorePlugins assembles the scoring-framework registry, called once by HandlerStart.
// weight is the dimension's bit offset (segment start); the synthesizer generically shifts
// segment value << weight. Registration order topology→previousNode→subHealth→chipCount is
// fixed: reordering changes the accumulated bits and breaks per-bit agreement with the
// independent frameworkBitFormula. prev weight=ScoreOriginalShift(11): tier 2 → 2<<11=bit12
// (P1), tier 1 → bit11 (P2); topo=10, subHealth=8 (2 bits), chip=0. See score_weights.go.
func (sHandle *ScheduleHandler) initScorePlugins() {
	sHandle.scorePlugins = []ScorePluginWithWeight{
		{plugin: &scorePluginFunc{name: "topology", scoreFn: sHandle.scoreTopology}, weight: util.ScoreTopoShift},
		{plugin: &scorePluginFunc{name: "previousNode", scoreFn: sHandle.scorePreviousNode}, weight: util.ScoreOriginalShift},
		{plugin: &scorePluginFunc{name: "subHealth", scoreFn: sHandle.scoreSubHealth}, weight: util.ScoreHealthShift},
		{plugin: &scorePluginFunc{name: "chipCount", scoreFn: sHandle.scoreChipCount}, weight: util.ScoreAvailShift},
	}
}
