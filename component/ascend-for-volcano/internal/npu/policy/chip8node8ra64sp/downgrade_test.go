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

// Package chip8node8ra64sp for downgrade functions ut cases
package chip8node8ra64sp

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

const (
	spHalveNodeNum1 = 1
	spHalveNodeNum2 = 2
	spHalveNodeNum3 = 3
	spHalveNodeNum4 = 4
	spHalveNodeNum6 = 6
	spHalveNPU8     = 8
	spHalveNPU16    = 16
)

type halvedSpBlockCase struct {
	name              string
	configuredSpBlock int
	maxNodeNPUNum     int
	windows           int
	want              int
}

func buildHalvedSpBlockCases() []halvedSpBlockCase {
	return []halvedSpBlockCase{
		{name: "should keep the node count when windows is zero",
			configuredSpBlock: spHalveNodeNum4, maxNodeNPUNum: spHalveNPU8,
			windows: 0, want: spHalveNodeNum4},
		{name: "should halve to whole nodes when one window passed",
			configuredSpBlock: spHalveNodeNum4, maxNodeNPUNum: spHalveNPU8,
			windows: 1, want: spHalveNodeNum2},
		{name: "should reach one node when windows halve to the bottom",
			configuredSpBlock: spHalveNodeNum4, maxNodeNPUNum: spHalveNPU8,
			windows: 2, want: spHalveNodeNum1},
		{name: "should fall to one node when the half is below one node",
			configuredSpBlock: spHalveNodeNum4, maxNodeNPUNum: spHalveNPU8,
			windows: 3, want: spHalveNodeNum1},
		{name: "should fall to one node when the half breaks whole nodes",
			configuredSpBlock: spHalveNodeNum3, maxNodeNPUNum: spHalveNPU8,
			windows: 1, want: spHalveNodeNum1},
		{name: "should halve to whole nodes on bigger nodes",
			configuredSpBlock: spHalveNodeNum6, maxNodeNPUNum: spHalveNPU16,
			windows: 1, want: spHalveNodeNum3},
	}
}

func TestHalvedSpBlockNodeNum(t *testing.T) {
	for _, tt := range buildHalvedSpBlockCases() {
		convey.Convey(tt.name, t, func() {
			tp := &chip8node8ra64sp{}
			tp.configuredSpBlock = tt.configuredSpBlock
			tp.MaxNodeNPUNum = tt.maxNodeNPUNum
			convey.So(tp.halvedSpBlockNodeNum(tt.windows), convey.ShouldEqual, tt.want)
		})
	}
}

type raBlockNPUOfNodeCountCase struct {
	name              string
	raBlockNodeNum    int
	configuredRaBlock int
	maxNodeNPUNum     int
	tpBlockNPUNum     int
	want              int
}

func buildRaBlockNPUOfNodeCountCases() []raBlockNPUOfNodeCountCase {
	return []raBlockNPUOfNodeCountCase{
		{name: "should fill whole nodes when the count is below the configured one",
			raBlockNodeNum: spHalveNodeNum2, configuredRaBlock: spHalveNodeNum4,
			maxNodeNPUNum: spHalveNPU8, want: spHalveNPU16},
		{name: "should keep the configured npu number when the count equals the configured one",
			raBlockNodeNum: spHalveNodeNum4, configuredRaBlock: spHalveNodeNum4,
			maxNodeNPUNum: spHalveNPU8, tpBlockNPUNum: spHalveNPU16, want: spHalveNPU16},
	}
}

func TestRaBlockNPUOfNodeCount(t *testing.T) {
	for _, tt := range buildRaBlockNPUOfNodeCountCases() {
		convey.Convey(tt.name, t, func() {
			tp := &chip8node8ra64sp{}
			tp.NPUJob = &util.NPUJob{}
			tp.configuredRaBlock = tt.configuredRaBlock
			tp.MaxNodeNPUNum = tt.maxNodeNPUNum
			tp.TpBlockNPUNum = tt.tpBlockNPUNum
			convey.So(tp.raBlockNPUOfNodeCount(tt.raBlockNodeNum), convey.ShouldEqual, tt.want)
		})
	}
}

type ra64BottomLevelCase struct {
	name      string
	nodeCount int
	want      int
}

func buildRa64BottomLevelCases() []ra64BottomLevelCase {
	return []ra64BottomLevelCase{
		{name: "should return the trailing zeros when the count is a power of two",
			nodeCount: spHalveNodeNum4, want: spHalveNodeNum2},
		{name: "should return one more when the count is not a power of two",
			nodeCount: spHalveNodeNum6, want: spHalveNodeNum2},
		{name: "should return zero when the count is one", nodeCount: spHalveNodeNum1, want: 0},
	}
}

func TestRa64BottomLevel(t *testing.T) {
	for _, tt := range buildRa64BottomLevelCases() {
		convey.Convey(tt.name, t, func() {
			convey.So(ra64BottomLevel(tt.nodeCount), convey.ShouldEqual, tt.want)
		})
	}
}

const (
	ra64TestNPU8     = 8
	ra64TestNPU16    = 16
	ra64TestNPU32    = 32
	ra64TestNPU24    = 24
	ra64TestTaskNum5 = 5
	ra64TestTask16   = 16
)

type ra64SpBlockOfWaitedWindowsCase struct {
	name              string
	configuredSpBlock int
	configuredRaBlock int
	maxNodeNPUNum     int
	windows           int
	wantOk            bool
	wantSpBlockNPUNum int
	wantRaBlockNPUNum int
	wantLevel         int
}

func buildRa64SpBlockOfWaitedWindowsCases() []ra64SpBlockOfWaitedWindowsCase {
	return []ra64SpBlockOfWaitedWindowsCase{
		{name: "should return false when configured ra-block is not positive",
			configuredSpBlock: spHalveNodeNum4, maxNodeNPUNum: ra64TestNPU8, windows: 1},
		{name: "should return false when windows is not positive",
			configuredSpBlock: spHalveNodeNum4, configuredRaBlock: spHalveNodeNum2,
			maxNodeNPUNum: ra64TestNPU8, windows: 0},
		{name: "should halve the ra-block first when the windows stay in its ladder",
			configuredSpBlock: spHalveNodeNum4, configuredRaBlock: spHalveNodeNum2,
			maxNodeNPUNum: ra64TestNPU8, windows: 1,
			wantOk: true, wantSpBlockNPUNum: ra64TestNPU32, wantRaBlockNPUNum: ra64TestNPU8,
			wantLevel: spHalveNodeNum1},
		{name: "should halve the sp-block after the ra-block ladder bottoms out",
			configuredSpBlock: spHalveNodeNum4, configuredRaBlock: spHalveNodeNum2,
			maxNodeNPUNum: ra64TestNPU8, windows: 2,
			wantOk: true, wantSpBlockNPUNum: ra64TestNPU16, wantRaBlockNPUNum: ra64TestNPU8,
			wantLevel: spHalveNodeNum2},
		{name: "should settle at the bottom when windows exceed the whole ladder",
			configuredSpBlock: spHalveNodeNum4, configuredRaBlock: spHalveNodeNum2,
			maxNodeNPUNum: ra64TestNPU8, windows: 10,
			wantOk: true, wantSpBlockNPUNum: ra64TestNPU8, wantRaBlockNPUNum: ra64TestNPU8,
			wantLevel: spHalveNodeNum3},
		{name: "should return false when the configuration holds no lower level",
			configuredSpBlock: spHalveNodeNum1, configuredRaBlock: spHalveNodeNum1,
			maxNodeNPUNum: ra64TestNPU8, windows: 10},
	}
}

func TestRa64SpBlockOfWaitedWindows(t *testing.T) {
	for _, tt := range buildRa64SpBlockOfWaitedWindowsCases() {
		convey.Convey(tt.name, t, func() {
			tp := &chip8node8ra64sp{}
			tp.NPUJob = &util.NPUJob{}
			tp.configuredSpBlock = tt.configuredSpBlock
			tp.configuredRaBlock = tt.configuredRaBlock
			tp.MaxNodeNPUNum = tt.maxNodeNPUNum
			tp.TpBlockNPUNum = ra64TestNPU16
			config, level, ok := tp.ra64SpBlockOfWaitedWindows(tt.windows)
			convey.So(ok, convey.ShouldEqual, tt.wantOk)
			if !tt.wantOk {
				convey.So(config, convey.ShouldBeEmpty)
				convey.So(level, convey.ShouldEqual, 0)
				return
			}
			var cfg ra64SpBlockConfig
			convey.So(json.Unmarshal([]byte(config), &cfg), convey.ShouldBeNil)
			convey.So(cfg.SpBlockNPUNum, convey.ShouldEqual, tt.wantSpBlockNPUNum)
			convey.So(cfg.RaBlockNPUNum, convey.ShouldEqual, tt.wantRaBlockNPUNum)
			convey.So(level, convey.ShouldEqual, tt.wantLevel)
		})
	}
}

type applyRa64SpBlockConfigCase struct {
	name           string
	config         string
	spBlock        int
	tpBlock        int
	npuTaskNum     int
	maxNodeNPUNum  int
	spBlockNPUNum  int
	tpBlockNPUNum  int
	wantApply      bool
	wantSpBlock    int
	wantTpBlock    int
	wantSpBlockNPU int
	wantTpBlockNPU int
}

func buildApplyRa64SpBlockConfigCases() []applyRa64SpBlockConfigCase {
	// a valid downgraded pair: sp-block 2 nodes (16 npus), ra-block 1 node (8 npus)
	downgraded, _ := json.Marshal(ra64SpBlockConfig{
		SpBlockNPUNum: ra64TestNPU16, RaBlockNPUNum: ra64TestNPU8})
	// unreadable config
	unreadable := "not-json"
	// ra-block breaks the block relation: ra-block over sp-block
	relationBroken, _ := json.Marshal(ra64SpBlockConfig{
		SpBlockNPUNum: ra64TestNPU16, RaBlockNPUNum: ra64TestNPU24})
	// the config does not lower the session: sp-block equals the configured one
	noLower, _ := json.Marshal(ra64SpBlockConfig{
		SpBlockNPUNum: ra64TestNPU32, RaBlockNPUNum: ra64TestNPU16})
	// task divisibility broken: 5 tasks divide neither 2 nodes nor 1 node
	divisibilityBroken, _ := json.Marshal(ra64SpBlockConfig{
		SpBlockNPUNum: ra64TestNPU16, RaBlockNPUNum: ra64TestNPU8})
	return []applyRa64SpBlockConfigCase{
		{name: "should return false when the config is unreadable", config: unreadable,
			spBlock: spHalveNodeNum4, tpBlock: spHalveNodeNum2, maxNodeNPUNum: ra64TestNPU8},
		{name: "should return false when the block relation is broken",
			config: string(relationBroken), spBlock: spHalveNodeNum4, tpBlock: spHalveNodeNum2,
			maxNodeNPUNum: ra64TestNPU8, npuTaskNum: ra64TestTask16},
		{name: "should return false when the task divisibility is broken",
			config: string(divisibilityBroken), spBlock: spHalveNodeNum4, tpBlock: spHalveNodeNum2,
			maxNodeNPUNum: ra64TestNPU8, npuTaskNum: ra64TestTaskNum5},
		{name: "should return false when the config does not lower the session",
			config: string(noLower), spBlock: spHalveNodeNum4, tpBlock: spHalveNodeNum2,
			maxNodeNPUNum: ra64TestNPU8, npuTaskNum: ra64TestTask16},
		{name: "should apply the downgraded block pair to the session",
			config: string(downgraded), spBlock: spHalveNodeNum4, tpBlock: spHalveNodeNum2,
			maxNodeNPUNum: ra64TestNPU8, npuTaskNum: ra64TestTask16,
			wantApply: true, wantSpBlock: spHalveNodeNum2, wantTpBlock: spHalveNodeNum1,
			wantSpBlockNPU: ra64TestNPU16, wantTpBlockNPU: ra64TestNPU8},
	}
}

func TestApplyRa64SpBlockConfig(t *testing.T) {
	for _, tt := range buildApplyRa64SpBlockConfigCases() {
		convey.Convey(tt.name, t, func() {
			tp := &chip8node8ra64sp{}
			tp.NPUJob = &util.NPUJob{}
			tp.spBlock = tt.spBlock
			tp.tpBlock = tt.tpBlock
			tp.NPUTaskNum = tt.npuTaskNum
			tp.MaxNodeNPUNum = tt.maxNodeNPUNum
			tp.SpBlockNPUNum = tt.spBlockNPUNum
			tp.TpBlockNPUNum = tt.tpBlockNPUNum
			got := tp.applyRa64SpBlockConfig(tt.config)
			convey.So(got, convey.ShouldEqual, tt.wantApply)
			if !tt.wantApply {
				convey.So(tp.spBlock, convey.ShouldEqual, tt.spBlock)
				convey.So(tp.tpBlock, convey.ShouldEqual, tt.tpBlock)
				return
			}
			convey.So(tp.spBlock, convey.ShouldEqual, tt.wantSpBlock)
			convey.So(tp.tpBlock, convey.ShouldEqual, tt.wantTpBlock)
			convey.So(tp.SpBlockNPUNum, convey.ShouldEqual, tt.wantSpBlockNPU)
			convey.So(tp.TpBlockNPUNum, convey.ShouldEqual, tt.wantTpBlockNPU)
		})
	}
}
