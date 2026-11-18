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
Package chip8node8sp is using for HuaWei ascend 800I A5 SuperPod affinity schedule.
*/
package chip8node8sp

import (
	"encoding/json"
	"math/bits"

	"volcano.sh/volcano/pkg/scheduler/api"
)

// spBlockConfig is the storable form of the downgraded constraint: the npu number of
// one sp-block, the unit the sp-block of the job configuration is counted in.
type spBlockConfig struct {
	SpBlockNPUNum int `json:"sp-block"` // npu num of one super pod block
}

// DowngradeConstraint plugs this policy into the common downgrade entry right after
// the job validation: every node selection derives from the sp-block it settles.
func (tp *chip8node8sp) DowngradeConstraint(jobID api.JobID) {
	if tp == nil {
		return
	}
	tp.DowngradeConstraintIfTimeout(jobID, tp.spBlockOfWaitedWindows, tp.applySpBlockConfig)
}

// spBlockOfWaitedWindows is the ladder of this policy: the constraint of the level the
// waited windows buy, the sp-block of the job configuration halved once per window
// (util.DefaultDowngradedFactor). A level which does not fill whole nodes falls to the
// bottom line of one node directly, every deeper level settling there as well.
func (tp *chip8node8sp) spBlockOfWaitedWindows(windows int) (string, int, bool) {
	if tp.configuredSpBlock <= 0 || tp.MaxNodeNPUNum <= 0 || windows <= 0 {
		return "", 0, false
	}
	// the ladder settles before the wait deepens: the constraint and the level of every
	// deeper wait are the ones of its bottom
	if bottomLevel := tp.spBlockBottomLevel(); windows > bottomLevel {
		windows = bottomLevel
	}
	configuredNPUNum := tp.configuredSpBlock * tp.MaxNodeNPUNum
	// the windows are the halvings of the configuration, one shift settles the level
	// instead of walking the ladder step by step
	npuNum := configuredNPUNum >> uint(windows)
	if npuNum < tp.MaxNodeNPUNum || npuNum%tp.MaxNodeNPUNum != 0 {
		npuNum = tp.MaxNodeNPUNum
	}
	if npuNum >= configuredNPUNum {
		// the configuration holds no lower level than itself
		return "", 0, false
	}
	data, err := json.Marshal(spBlockConfig{SpBlockNPUNum: npuNum})
	if err != nil {
		return "", 0, false
	}
	return string(data), windows, true
}

// spBlockBottomLevel is the level the ladder of the configuration settles at, every
// deeper wait holding it as well: the halvings which keep whole nodes are the trailing
// zeros of the sp-block node count, a node count which is not a power of two spends the
// halving past them on the bottom line.
func (tp *chip8node8sp) spBlockBottomLevel() int {
	level := bits.TrailingZeros(uint(tp.configuredSpBlock))
	if tp.configuredSpBlock&(tp.configuredSpBlock-1) != 0 {
		level++
	}
	return level
}

// applySpBlockConfig applies the constraint to the session, refusing an unreadable
// one, one which does not fill whole nodes and one which does not lower the session.
func (tp *chip8node8sp) applySpBlockConfig(config string) bool {
	var cfg spBlockConfig
	if json.Unmarshal([]byte(config), &cfg) != nil || cfg.SpBlockNPUNum <= 0 || tp.MaxNodeNPUNum <= 0 ||
		cfg.SpBlockNPUNum%tp.MaxNodeNPUNum != 0 || cfg.SpBlockNPUNum >= tp.spBlock*tp.MaxNodeNPUNum {
		return false
	}
	tp.spBlock = cfg.SpBlockNPUNum / tp.MaxNodeNPUNum
	tp.SpBlockNPUNum = cfg.SpBlockNPUNum
	return true
}
