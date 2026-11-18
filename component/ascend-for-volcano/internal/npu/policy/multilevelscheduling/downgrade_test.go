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

package multilevelscheduling

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

const (
	mlsLevel1Nodes = 1
	mlsLevel2Nodes = 2
	mlsLevel4Nodes = 4
	mlsLevel8Nodes = 8
	mlsTaskNum8    = 8
	mlsLevel1Key   = "level1"
	mlsLevel2Key   = "level2"
)

func buildMLSAffinityBlocks(level1, level2 int) map[string]int {
	return map[string]int{
		mlsLevel1Key: level1,
		mlsLevel2Key: level2,
	}
}

type affinityOfWaitedWindowsCase struct {
	name          string
	configured    map[string]int
	taskNum       int
	windows       int
	wantOk        bool
	wantLevel1    int
	wantLevel2    int
	wantLevelBack int
}

func buildAffinityOfWaitedWindowsCases() []affinityOfWaitedWindowsCase {
	return []affinityOfWaitedWindowsCase{
		{name: "should return false when the configured blocks are empty",
			taskNum: mlsTaskNum8, windows: 1},
		{name: "should return false when windows is not positive",
			configured: buildMLSAffinityBlocks(mlsLevel1Nodes, mlsLevel2Nodes),
			taskNum:    mlsTaskNum8, windows: 0},
		{name: "should halve the finest level first when one window passed",
			configured: buildMLSAffinityBlocks(mlsLevel2Nodes, mlsLevel4Nodes),
			taskNum:    mlsTaskNum8, windows: 1,
			wantOk: true, wantLevel1: mlsLevel1Nodes, wantLevel2: mlsLevel4Nodes,
			wantLevelBack: 1},
		{name: "should halve every level in order when windows pass the first level",
			configured: buildMLSAffinityBlocks(mlsLevel2Nodes, mlsLevel4Nodes),
			taskNum:    mlsTaskNum8, windows: 2,
			wantOk: true, wantLevel1: mlsLevel1Nodes, wantLevel2: mlsLevel2Nodes,
			wantLevelBack: 2},
		{name: "should settle every level at one when windows exceed the ladder",
			configured: buildMLSAffinityBlocks(mlsLevel2Nodes, mlsLevel4Nodes),
			taskNum:    mlsTaskNum8, windows: 10,
			wantOk: true, wantLevel1: mlsLevel1Nodes, wantLevel2: mlsLevel1Nodes,
			wantLevelBack: 3},
		{name: "should return false when the configuration holds no lower level",
			configured: buildMLSAffinityBlocks(mlsLevel1Nodes, mlsLevel1Nodes),
			taskNum:    mlsTaskNum8, windows: 10},
	}
}

func TestAffinityOfWaitedWindows(t *testing.T) {
	for _, tt := range buildAffinityOfWaitedWindowsCases() {
		convey.Convey(tt.name, t, func() {
			mh := newTestHandler()
			mh.configuredAffinityBlocks = tt.configured
			mh.NPUJob = &util.NPUJob{NPUTaskNum: tt.taskNum}
			config, level, ok := mh.affinityOfWaitedWindows(tt.windows)
			convey.So(ok, convey.ShouldEqual, tt.wantOk)
			if !tt.wantOk {
				convey.So(config, convey.ShouldBeEmpty)
				convey.So(level, convey.ShouldEqual, 0)
				return
			}
			var cfg affinityBlocksConfig
			convey.So(json.Unmarshal([]byte(config), &cfg), convey.ShouldBeNil)
			convey.So(cfg[mlsLevel1Key], convey.ShouldEqual, tt.wantLevel1)
			convey.So(cfg[mlsLevel2Key], convey.ShouldEqual, tt.wantLevel2)
			convey.So(level, convey.ShouldEqual, tt.wantLevelBack)
		})
	}
}

type divisibleBlocksCase struct {
	name    string
	blocks  map[string]int
	taskNum int
	want    bool
}

func buildDivisibleBlocksCases() []divisibleBlocksCase {
	return []divisibleBlocksCase{
		{name: "should return true when the whole level chain divides",
			blocks:  buildMLSAffinityBlocks(mlsLevel2Nodes, mlsLevel4Nodes),
			taskNum: mlsTaskNum8, want: true},
		{name: "should return false when the task count does not divide the coarsest level",
			blocks:  buildMLSAffinityBlocks(mlsLevel2Nodes, mlsLevel4Nodes),
			taskNum: 6, want: false},
		{name: "should return false when a level is missing",
			blocks:  map[string]int{mlsLevel2Key: mlsLevel4Nodes},
			taskNum: mlsTaskNum8, want: false},
		{name: "should return false when a level count is not positive",
			blocks:  buildMLSAffinityBlocks(0, mlsLevel4Nodes),
			taskNum: mlsTaskNum8, want: false},
	}
}

func TestDivisibleBlocks(t *testing.T) {
	for _, tt := range buildDivisibleBlocksCases() {
		convey.Convey(tt.name, t, func() {
			mh := newTestHandler()
			mh.NPUJob = &util.NPUJob{NPUTaskNum: tt.taskNum}
			convey.So(mh.divisibleBlocks(tt.blocks), convey.ShouldEqual, tt.want)
		})
	}
}

type applyAffinityConfigCase struct {
	name       string
	config     string
	configured map[string]int
	taskNum    int
	wantApply  bool
}

func buildApplyAffinityConfigCases() []applyAffinityConfigCase {
	// a valid downgraded config: level1 halved, level2 keeps
	downgraded, _ := json.Marshal(affinityBlocksConfig(
		buildMLSAffinityBlocks(mlsLevel1Nodes, mlsLevel4Nodes)))
	// no lower level: the config equals the configured one
	noLower, _ := json.Marshal(affinityBlocksConfig(
		buildMLSAffinityBlocks(mlsLevel2Nodes, mlsLevel4Nodes)))
	// an unknown level key
	unknownLevel, _ := json.Marshal(affinityBlocksConfig(
		map[string]int{"level3": mlsLevel1Nodes, mlsLevel2Key: mlsLevel4Nodes}))
	return []applyAffinityConfigCase{
		{name: "should return false when the config is unreadable", config: "not-json",
			configured: buildMLSAffinityBlocks(mlsLevel2Nodes, mlsLevel4Nodes),
			taskNum:    mlsTaskNum8},
		{name: "should return false when the config does not lower the session",
			config: string(noLower), configured: buildMLSAffinityBlocks(mlsLevel2Nodes, mlsLevel4Nodes),
			taskNum: mlsTaskNum8},
		{name: "should return false when the config holds an unknown level",
			config: string(unknownLevel), configured: buildMLSAffinityBlocks(mlsLevel2Nodes, mlsLevel4Nodes),
			taskNum: mlsTaskNum8},
		{name: "should apply the downgraded constraint to the session",
			config: string(downgraded), configured: buildMLSAffinityBlocks(mlsLevel2Nodes, mlsLevel4Nodes),
			taskNum: mlsTaskNum8, wantApply: true},
	}
}

func TestApplyAffinityConfig(t *testing.T) {
	for _, tt := range buildApplyAffinityConfigCases() {
		convey.Convey(tt.name, t, func() {
			mh := newTestHandler()
			mh.configuredAffinityBlocks = tt.configured
			mh.NPUJob = &util.NPUJob{NPUTaskNum: tt.taskNum,
				Tasks: map[api.TaskID]util.NPUTask{}}
			got := mh.applyAffinityConfig(tt.config)
			convey.So(got, convey.ShouldEqual, tt.wantApply)
			if !tt.wantApply {
				convey.So(mh.NPUJob.AffinityBlocks, convey.ShouldBeEmpty)
				return
			}
			convey.So(mh.NPUJob.AffinityBlocks[mlsLevel1Key], convey.ShouldEqual, mlsLevel1Nodes)
			convey.So(mh.NPUJob.AffinityBlocks[mlsLevel2Key], convey.ShouldEqual, mlsLevel4Nodes)
			convey.So(len(mh.taskLevels), convey.ShouldBeGreaterThan, 0)
		})
	}
}

func TestDowngradeConstraintNilHandler(t *testing.T) {
	convey.Convey("should return directly when handler is nil", t, func() {
		var mh *MultilevelHandler
		mh.DowngradeConstraint(api.JobID("nil-job"))
	})
}
