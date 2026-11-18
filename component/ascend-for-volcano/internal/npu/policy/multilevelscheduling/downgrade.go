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
Package multilevelscheduling for scheduling NPU job with general abstract network topology configuration.
*/
package multilevelscheduling

import (
	"encoding/json"
	"strconv"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// affinityBlocksConfig is the storable form of the downgraded constraint: the req node
// count of every affinity level, the keys the affinity-config of the job counts in.
type affinityBlocksConfig map[string]int

// DowngradeConstraint plugs this policy into the common downgrade entry right after
// the job validation: every task tree derives from the level values it settles.
func (mh *MultilevelHandler) DowngradeConstraint(jobID api.JobID) {
	if mh == nil {
		return
	}
	mh.DowngradeConstraintIfTimeout(jobID, mh.affinityOfWaitedWindows, mh.applyAffinityConfig)
}

// affinityOfWaitedWindows is the ladder of this policy: the constraint of the level the
// waited windows buy, the levels halved one per window (util.DefaultDowngradedFactor)
// from the finest level1 up to the coarsest one, a level settling at one once halved
// past its powers of two. A halving which breaks the divisibility of the level chain
// falls to one directly, one dividing every count. The level reported back paces the
// effect time of the pod markers.
func (mh *MultilevelHandler) affinityOfWaitedWindows(windows int) (string, int, bool) {
	if len(mh.configuredAffinityBlocks) == 0 || windows <= 0 {
		return "", 0, false
	}
	blocks := make(affinityBlocksConfig, len(mh.configuredAffinityBlocks))
	for level, reqNode := range mh.configuredAffinityBlocks {
		blocks[level] = reqNode
	}
	level := 0
	// walk the ladder from the finest level1 up, every whole window halving the level
	// it spends on
	for i := util.Level1Number; i <= len(blocks) && windows > 0; i++ {
		key := util.TopoLevelPrefix + strconv.Itoa(i)
		for blocks[key] > 1 && windows > 0 {
			blocks[key] >>= 1
			if !mh.divisibleBlocks(blocks) {
				blocks[key] = 1
			}
			windows--
			level++
		}
	}
	if level == 0 {
		// the configuration holds no lower level than itself
		return "", 0, false
	}
	data, err := json.Marshal(blocks)
	if err != nil {
		return "", 0, false
	}
	return string(data), level, true
}

// divisibleBlocks reports whether the blocks keep the level chain divisible: the task
// count divides the coarsest level and every level divides the one below it.
func (mh *MultilevelHandler) divisibleBlocks(blocks affinityBlocksConfig) bool {
	upper := mh.NPUTaskNum - mh.CountBackupTasks()
	for i := len(blocks); i >= util.Level1Number; i-- {
		key := util.TopoLevelPrefix + strconv.Itoa(i)
		reqNode, ok := blocks[key]
		if !ok || reqNode <= 0 || upper%reqNode != 0 {
			return false
		}
		upper = reqNode
	}
	return true
}

// applyAffinityConfig applies the constraint to the session, refusing an unreadable
// one, one which breaks the level chain and one which does not lower the session.
func (mh *MultilevelHandler) applyAffinityConfig(config string) bool {
	var cfg affinityBlocksConfig
	if json.Unmarshal([]byte(config), &cfg) != nil || len(cfg) == 0 ||
		len(cfg) != len(mh.configuredAffinityBlocks) {
		return false
	}
	downgraded := false
	for level, reqNode := range cfg {
		configured, ok := mh.configuredAffinityBlocks[level]
		if !ok || reqNode <= 0 || reqNode > configured {
			return false
		}
		downgraded = downgraded || reqNode < configured
	}
	if !downgraded || !mh.divisibleBlocks(cfg) {
		return false
	}
	taskLevels, err := util.GetTaskTreeLevels(cfg, mh.NPUTaskNum-mh.CountBackupTasks())
	if err != nil {
		return false
	}
	mh.AffinityBlocks = cfg
	mh.taskLevels = taskLevels
	return true
}
