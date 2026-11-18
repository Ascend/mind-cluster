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

package chip8node8sp

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const (
	layoutSpBlock    = 4
	layoutTaskNum    = 8
	layoutTaskOnNode = "node-placed"
	layoutJobName    = "layout-job"
	spBlockNodeNum2  = 2
	spBlockNodeNum3  = 3
	spBlockNodeNum6  = 6
	levelNum1        = 1
)

var layoutJobUID = api.JobID("job-layout-test")

// layoutTestNodes builds count super nodes with the same name prefix.
func layoutTestNodes(prefix string, count int) []plugin.SuperNode {
	nodes := make([]plugin.SuperNode, 0, count)
	for i := 0; i < count; i++ {
		nodes = append(nodes, plugin.SuperNode{Name: prefix + strconv.Itoa(i)})
	}
	return nodes
}

type layoutMatchCase struct {
	name      string
	fJob      *rescheduling.FaultJob
	spBlock   int
	taskNum   int
	wantMatch bool
}

func buildLayoutMatchCases() []layoutMatchCase {
	return []layoutMatchCase{
		{name: "should match when fault job is nil", fJob: nil, spBlock: layoutSpBlock, wantMatch: true},
		{name: "should match when layout is empty", fJob: &rescheduling.FaultJob{}, spBlock: layoutSpBlock,
			wantMatch: true},
		{name: "should match when sp-block is not positive", spBlock: 0, taskNum: 2, wantMatch: true,
			fJob: &rescheduling.FaultJob{SuperPods: map[string][]plugin.SuperNode{
				"0": layoutTestNodes("n0-", 2)}}},
		{name: "should match when every block is a whole sp-block", spBlock: layoutSpBlock,
			taskNum: layoutTaskNum, wantMatch: true,
			fJob: &rescheduling.FaultJob{SuperPods: map[string][]plugin.SuperNode{
				"0": layoutTestNodes("n0-", 4), "1": layoutTestNodes("n1-", 4)}}},
		{name: "should match when the trailing block is short", spBlock: layoutSpBlock, taskNum: 6,
			wantMatch: true,
			fJob: &rescheduling.FaultJob{SuperPods: map[string][]plugin.SuperNode{
				"0": layoutTestNodes("n0-", 4), "1": layoutTestNodes("n1-", 2)}}},
		{name: "should not match when a non-trailing block is short", spBlock: layoutSpBlock,
			taskNum: 6, wantMatch: false,
			fJob: &rescheduling.FaultJob{SuperPods: map[string][]plugin.SuperNode{
				"0": layoutTestNodes("n0-", 2), "1": layoutTestNodes("n1-", 4)}}},
		{name: "should not match when a block is longer than sp-block", spBlock: layoutSpBlock,
			taskNum: 5, wantMatch: false,
			fJob: &rescheduling.FaultJob{SuperPods: map[string][]plugin.SuperNode{
				"0": layoutTestNodes("n0-", 5)}}},
		{name: "should match when the highest block is the short trailing one", spBlock: layoutSpBlock,
			taskNum: 6, wantMatch: true,
			fJob: &rescheduling.FaultJob{SuperPods: map[string][]plugin.SuperNode{
				"2": layoutTestNodes("n2-", 4), "10": layoutTestNodes("n10-", 2)}}},
		{name: "should not match when a numerically lower block is short", spBlock: layoutSpBlock,
			taskNum: 6, wantMatch: false,
			fJob: &rescheduling.FaultJob{SuperPods: map[string][]plugin.SuperNode{
				"9": layoutTestNodes("n9-", 2), "10": layoutTestNodes("n10-", 4)}}},
		{name: "should not match when the block key is non-numeric", spBlock: layoutSpBlock,
			taskNum: 2, wantMatch: false,
			fJob: &rescheduling.FaultJob{SuperPods: map[string][]plugin.SuperNode{
				"tail": layoutTestNodes("nt-", 2)}}},
		{name: "should match when the layout does not cover the whole job", spBlock: layoutSpBlock,
			taskNum: 3, wantMatch: true,
			fJob: &rescheduling.FaultJob{SuperPods: map[string][]plugin.SuperNode{
				"0": layoutTestNodes("n0-", 2)}}},
	}
}

func TestSuperPodsLayoutMatchesSpBlock(t *testing.T) {
	for _, tt := range buildLayoutMatchCases() {
		convey.Convey(tt.name, t, func() {
			tp := newModuleSuperPod()
			tp.spBlock = tt.spBlock
			tp.NPUJob = &util.NPUJob{NPUTaskNum: tt.taskNum}
			convey.So(tp.superPodsLayoutMatchesSpBlock(tt.fJob), convey.ShouldEqual, tt.wantMatch)
		})
	}
}

func TestSuperPodsLayoutMatchesSpBlockNilHandler(t *testing.T) {
	convey.Convey("should match when handler is nil", t, func() {
		var tp *chip8node8sp
		convey.So(tp.superPodsLayoutMatchesSpBlock(&rescheduling.FaultJob{}), convey.ShouldBeTrue)
	})
}

func TestSuperPodsLayoutMatchesSpBlockNilNPUJob(t *testing.T) {
	convey.Convey("should match when npu job is nil", t, func() {
		tp := newModuleSuperPod()
		tp.spBlock = 2
		fJob := &rescheduling.FaultJob{SuperPods: map[string][]plugin.SuperNode{
			"0": layoutTestNodes("n0-", 4)}}
		convey.So(tp.superPodsLayoutMatchesSpBlock(fJob), convey.ShouldBeTrue)
	})
}

type layoutMismatchCase struct {
	name       string
	placedTask bool
	wantErr    bool
}

func TestHandleLayoutMismatch(t *testing.T) {
	cases := []layoutMismatchCase{
		{name: "should wait with error when a placed task pins the old layout",
			placedTask: true, wantErr: true},
		{name: "should clear layout and cache when every task is pending", placedTask: false},
	}
	for _, tt := range cases {
		convey.Convey(tt.name, t, func() {
			tp := newModuleSuperPod()
			tp.spBlock = layoutSpBlock
			tp.NPUJob = &util.NPUJob{NPUTaskNum: layoutTaskNum,
				Tasks: map[api.TaskID]util.NPUTask{}}
			fJob := &rescheduling.FaultJob{JobUID: layoutJobUID, JobName: layoutJobName,
				SuperPods: map[string][]plugin.SuperNode{"0": layoutTestNodes("n0-", 4)}}
			tp.SuperPodInfo.SuperPodReschdInfo[layoutJobUID] = fJob.SuperPods
			if tt.placedTask {
				tp.NPUJob.Tasks["task-0"] = util.NPUTask{NodeName: layoutTaskOnNode}
			}
			got, err := tp.handleLayoutMismatch(fJob)
			convey.So(err != nil, convey.ShouldEqual, tt.wantErr)
			if tt.wantErr {
				convey.So(fJob.SuperPods, convey.ShouldNotBeNil)
				return
			}
			convey.So(got, convey.ShouldBeNil)
			convey.So(fJob.SuperPods, convey.ShouldBeNil)
			_, cached := tp.SuperPodInfo.SuperPodReschdInfo[layoutJobUID]
			convey.So(cached, convey.ShouldBeFalse)
		})
	}
}

type jobHasPlacedTaskCase struct {
	name   string
	tasks  map[api.TaskID]util.NPUTask
	wantOn bool
}

func TestJobHasPlacedTask(t *testing.T) {
	cases := []jobHasPlacedTaskCase{
		{name: "should return false when no task holds a node", wantOn: false,
			tasks: map[api.TaskID]util.NPUTask{"task-0": {}}},
		{name: "should return true when a task still holds a node", wantOn: true,
			tasks: map[api.TaskID]util.NPUTask{"task-0": {}, "task-1": {NodeName: layoutTaskOnNode}}},
		{name: "should return false when every task is pending", wantOn: false,
			tasks: map[api.TaskID]util.NPUTask{"task-0": {}, "task-1": {}}},
	}
	for _, tt := range cases {
		convey.Convey(tt.name, t, func() {
			tp := newModuleSuperPod()
			tp.NPUJob = &util.NPUJob{Tasks: tt.tasks}
			convey.So(tp.jobHasPlacedTask(), convey.ShouldEqual, tt.wantOn)
		})
	}
}

func TestJobHasPlacedTaskNilHandler(t *testing.T) {
	convey.Convey("should return false when handler is nil", t, func() {
		var tp *chip8node8sp
		convey.So(tp.jobHasPlacedTask(), convey.ShouldBeFalse)
	})
}

type spBlockOfWaitedWindowsCase struct {
	name              string
	configuredSpBlock int
	maxNodeNPUNum     int
	windows           int
	wantOk            bool
	wantSpBlockNPUNum int
	wantLevel         int
}

const (
	ladderSpBlockNPU8  = 8
	ladderSpBlockNPU12 = 12
	ladderSpBlockNPU32 = 32
	ladderNodeNPU8     = 8
)

func buildSpBlockOfWaitedWindowsCases() []spBlockOfWaitedWindowsCase {
	return []spBlockOfWaitedWindowsCase{
		{name: "should return false when configured sp-block is not positive",
			configuredSpBlock: 0, maxNodeNPUNum: ladderNodeNPU8, windows: 1},
		{name: "should return false when windows is not positive",
			configuredSpBlock: layoutSpBlock, maxNodeNPUNum: ladderNodeNPU8, windows: 0},
		{name: "should halve one whole node when one window passed",
			configuredSpBlock: layoutSpBlock, maxNodeNPUNum: ladderNodeNPU8, windows: 1,
			wantOk: true, wantSpBlockNPUNum: ladderSpBlockNPU8 * spBlockNodeNum2,
			wantLevel: levelNum1},
		{name: "should halve twice when two windows passed",
			configuredSpBlock: layoutSpBlock, maxNodeNPUNum: ladderNodeNPU8, windows: 2,
			wantOk: true, wantSpBlockNPUNum: ladderSpBlockNPU8, wantLevel: spBlockNodeNum2},
		{name: "should settle at the bottom level when windows exceed it",
			configuredSpBlock: layoutSpBlock, maxNodeNPUNum: ladderNodeNPU8, windows: 10,
			wantOk: true, wantSpBlockNPUNum: ladderSpBlockNPU8, wantLevel: spBlockNodeNum2},
		{name: "should fall to one node when the half breaks whole nodes",
			configuredSpBlock: spBlockNodeNum3, maxNodeNPUNum: ladderNodeNPU8, windows: 1,
			wantOk: true, wantSpBlockNPUNum: ladderSpBlockNPU8, wantLevel: levelNum1},
		{name: "should return false when the configuration holds no lower level",
			configuredSpBlock: levelNum1, maxNodeNPUNum: ladderNodeNPU8, windows: 10},
	}
}

func TestSpBlockOfWaitedWindows(t *testing.T) {
	for _, tt := range buildSpBlockOfWaitedWindowsCases() {
		convey.Convey(tt.name, t, func() {
			tp := newModuleSuperPod()
			tp.configuredSpBlock = tt.configuredSpBlock
			tp.MaxNodeNPUNum = tt.maxNodeNPUNum
			config, level, ok := tp.spBlockOfWaitedWindows(tt.windows)
			convey.So(ok, convey.ShouldEqual, tt.wantOk)
			if !tt.wantOk {
				convey.So(config, convey.ShouldBeEmpty)
				convey.So(level, convey.ShouldEqual, 0)
				return
			}
			var cfg spBlockConfig
			convey.So(json.Unmarshal([]byte(config), &cfg), convey.ShouldBeNil)
			convey.So(cfg.SpBlockNPUNum, convey.ShouldEqual, tt.wantSpBlockNPUNum)
			convey.So(level, convey.ShouldEqual, tt.wantLevel)
		})
	}
}

type spBlockBottomLevelCase struct {
	name              string
	configuredSpBlock int
	want              int
}

func buildSpBlockBottomLevelCases() []spBlockBottomLevelCase {
	return []spBlockBottomLevelCase{
		{name: "should return the trailing zeros when sp-block is a power of two",
			configuredSpBlock: layoutSpBlock, want: spBlockNodeNum2},
		{name: "should return one more when sp-block is not a power of two",
			configuredSpBlock: spBlockNodeNum6, want: spBlockNodeNum2},
		{name: "should return zero when sp-block is one",
			configuredSpBlock: levelNum1, want: 0},
	}
}

func TestSpBlockBottomLevel(t *testing.T) {
	for _, tt := range buildSpBlockBottomLevelCases() {
		convey.Convey(tt.name, t, func() {
			tp := newModuleSuperPod()
			tp.configuredSpBlock = tt.configuredSpBlock
			convey.So(tp.spBlockBottomLevel(), convey.ShouldEqual, tt.want)
		})
	}
}

type applySpBlockConfigCase struct {
	name          string
	config        string
	spBlock       int
	maxNodeNPUNum int
	spBlockNPUNum int
	wantApply     bool
	wantSpBlock   int
	wantSpNPU     int
}

const applyTestMaxNodeNPU8 = 8

func buildApplySpBlockConfigCases() []applySpBlockConfigCase {
	downgraded, _ := json.Marshal(spBlockConfig{SpBlockNPUNum: ladderSpBlockNPU8})
	equal, _ := json.Marshal(spBlockConfig{SpBlockNPUNum: ladderSpBlockNPU32})
	broken, _ := json.Marshal(spBlockConfig{SpBlockNPUNum: ladderSpBlockNPU12})
	return []applySpBlockConfigCase{
		{name: "should return false when the config is unreadable", config: "not-json",
			spBlock: layoutSpBlock, maxNodeNPUNum: applyTestMaxNodeNPU8},
		{name: "should return false when the config does not fill whole nodes",
			config: string(broken), spBlock: layoutSpBlock, maxNodeNPUNum: applyTestMaxNodeNPU8},
		{name: "should return false when the config does not lower the session",
			config: string(equal), spBlock: layoutSpBlock, maxNodeNPUNum: applyTestMaxNodeNPU8},
		{name: "should apply the downgraded constraint to the session",
			config: string(downgraded), spBlock: layoutSpBlock, maxNodeNPUNum: applyTestMaxNodeNPU8,
			wantApply: true, wantSpBlock: levelNum1, wantSpNPU: ladderSpBlockNPU8},
	}
}

func TestApplySpBlockConfig(t *testing.T) {
	for _, tt := range buildApplySpBlockConfigCases() {
		convey.Convey(tt.name, t, func() {
			tp := newModuleSuperPod()
			tp.NPUJob = &util.NPUJob{}
			tp.spBlock = tt.spBlock
			tp.SpBlockNPUNum = tt.spBlockNPUNum
			tp.MaxNodeNPUNum = tt.maxNodeNPUNum
			got := tp.applySpBlockConfig(tt.config)
			convey.So(got, convey.ShouldEqual, tt.wantApply)
			if !tt.wantApply {
				convey.So(tp.spBlock, convey.ShouldEqual, tt.spBlock)
				convey.So(tp.SpBlockNPUNum, convey.ShouldEqual, tt.spBlockNPUNum)
				return
			}
			convey.So(tp.spBlock, convey.ShouldEqual, tt.wantSpBlock)
			convey.So(tp.SpBlockNPUNum, convey.ShouldEqual, tt.wantSpNPU)
		})
	}
}
