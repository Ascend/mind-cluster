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

/*
Package chip8node8ra64sp is using for HuaWei ascend 800I A5 rack super pod affinity schedule.
*/
package chip8node8ra64sp

import (
	"encoding/json"
	"math/bits"

	"volcano.sh/volcano/pkg/scheduler/api"
)

// ra64SpBlockConfig is the storable form of the downgraded constraint: the npu numbers
// of one sp-block and one ra-block, the units the job configuration counts them in.
type ra64SpBlockConfig struct {
	SpBlockNPUNum int `json:"sp-block"` // npu num of one super pod block
	RaBlockNPUNum int `json:"ra-block"` // npu num of one rack block
}

// DowngradeConstraint plugs this policy into the common downgrade entry right after
// the job validation: the rack selection and the super pod selection derive from the
// block pair it settles.
func (tp *chip8node8ra64sp) DowngradeConstraint(jobID api.JobID) {
	if tp == nil {
		return
	}
	tp.DowngradeConstraintIfTimeout(jobID, tp.ra64SpBlockOfWaitedWindows, tp.applyRa64SpBlockConfig)
}

// ra64SpBlockOfWaitedWindows is the ladder of this policy: the ra-block halves first
// down to one whole node, then the sp-block halves, both once per window
// (util.DefaultDowngradedFactor), ra-block never growing past sp-block. The path keeps
// spBlock%raBlock==0 at every level: a halved raBlock still divides spBlock and the
// second stage runs with a one-node raBlock. The level reported back paces the effect
// time of the pod markers.
func (tp *chip8node8ra64sp) ra64SpBlockOfWaitedWindows(windows int) (string, int, bool) {
	if tp.configuredSpBlock <= 0 || tp.configuredRaBlock <= 0 || tp.MaxNodeNPUNum <= 0 || windows <= 0 {
		return "", 0, false
	}
	// the ladder settles before the wait deepens: the constraint and the level of every
	// deeper wait are the ones of its bottom
	raSteps := ra64BottomLevel(tp.configuredRaBlock)
	spSteps := ra64BottomLevel(tp.configuredSpBlock)
	if windows > raSteps+spSteps {
		windows = raSteps + spSteps
	}
	spBlockNodeNum, raBlockNodeNum := tp.configuredSpBlock, tp.configuredRaBlock
	if windows <= raSteps {
		raBlockNodeNum >>= uint(windows)
	} else {
		raBlockNodeNum = nodeNum1
		spBlockNodeNum = tp.halvedSpBlockNodeNum(windows - raSteps)
	}
	if spBlockNodeNum >= tp.configuredSpBlock && raBlockNodeNum >= tp.configuredRaBlock {
		// the configuration holds no lower level than itself
		return "", 0, false
	}
	data, err := json.Marshal(ra64SpBlockConfig{
		SpBlockNPUNum: spBlockNodeNum * tp.MaxNodeNPUNum,
		RaBlockNPUNum: tp.raBlockNPUOfNodeCount(raBlockNodeNum),
	})
	if err != nil {
		return "", 0, false
	}
	return string(data), windows, true
}

// halvedSpBlockNodeNum halves the npu number of the configured sp-block windows
// times and turns it back to a node count: a half which does not fill whole
// nodes falls to the bottom line of one node, the same guard the node-group
// policies apply. A whole-node half keeps the task divisibility: the node count
// halves, so the task count per node doubles.
func (tp *chip8node8ra64sp) halvedSpBlockNodeNum(windows int) int {
	npuNum := (tp.configuredSpBlock * tp.MaxNodeNPUNum) >> uint(windows)
	if npuNum < tp.MaxNodeNPUNum || npuNum%tp.MaxNodeNPUNum != 0 {
		return nodeNum1
	}
	return npuNum / tp.MaxNodeNPUNum
}

// raBlockNPUOfNodeCount turns a halved raBlock node count back to its ra-block npu
// number: a count below the configured one fills whole nodes, the configured count
// itself keeps the configured npu number, which may be below one whole node.
func (tp *chip8node8ra64sp) raBlockNPUOfNodeCount(raBlockNodeNum int) int {
	if raBlockNodeNum < tp.configuredRaBlock {
		return raBlockNodeNum * tp.MaxNodeNPUNum
	}
	return tp.TpBlockNPUNum
}

// ra64BottomLevel is the halvings of a node count down to one node: the trailing
// zeros of a power of two, one more when it is not.
func ra64BottomLevel(nodeCount int) int {
	level := bits.TrailingZeros(uint(nodeCount))
	if nodeCount&(nodeCount-1) != 0 {
		level++
	}
	return level
}

// applyRa64SpBlockConfig applies the constraint to the session, refusing an unreadable
// one, one which breaks the block relation or the task divisibility the job validation
// established and one which does not lower the session.
func (tp *chip8node8ra64sp) applyRa64SpBlockConfig(config string) bool {
	var cfg ra64SpBlockConfig
	if json.Unmarshal([]byte(config), &cfg) != nil || cfg.SpBlockNPUNum <= 0 || cfg.RaBlockNPUNum <= 0 ||
		tp.MaxNodeNPUNum <= 0 || cfg.SpBlockNPUNum%tp.MaxNodeNPUNum != 0 {
		return false
	}
	spBlockNodeNum := cfg.SpBlockNPUNum / tp.MaxNodeNPUNum
	raBlockNodeNum := (cfg.RaBlockNPUNum + plusTpBlockNum) / tp.MaxNodeNPUNum
	blockRelationBroken := spBlockNodeNum <= 0 || raBlockNodeNum <= 0 ||
		raBlockNodeNum > spBlockNodeNum || spBlockNodeNum%raBlockNodeNum != 0
	taskDivisibilityBroken := tp.NPUTaskNum > 0 &&
		(tp.NPUTaskNum%spBlockNodeNum != 0 || tp.NPUTaskNum%raBlockNodeNum != 0)
	noLowerLevel := spBlockNodeNum >= tp.spBlock && raBlockNodeNum >= tp.tpBlock
	if blockRelationBroken || taskDivisibilityBroken || noLowerLevel {
		return false
	}
	tp.spBlock = spBlockNodeNum
	tp.tpBlock = raBlockNodeNum
	tp.SpBlockNPUNum = cfg.SpBlockNPUNum
	tp.TpBlockNPUNum = cfg.RaBlockNPUNum
	return true
}
