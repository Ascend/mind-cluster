/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package superPod is using for HuaWei Ascend800ia5 superPod pin affinity schedule.
*/
package superpod

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"volcano.sh/volcano/pkg/scheduler/api"
	test2 "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend800ia5"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/vnpu"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func TestIsDelayingJobCase1(t *testing.T) {
	convey.Convey("Test isDelayingJob case1", t, func() {
		// Setup common test data
		now := time.Now().Unix()

		convey.Convey("When tp is nil", func() {
			result := (*module800SuperPod)(nil).isDelayingJob(nil, nil)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When fJob is nil", func() {
			tp := &module800SuperPod{}
			result := tp.isDelayingJob(nil, nil)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When FaultTasks is nil", func() {
			tp := &module800SuperPod{}
			fJob := &rescheduling.FaultJob{FaultTasks: nil}
			result := tp.isDelayingJob(fJob, nil)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When waiting time exceeds threshold", func() {
			tp := &module800SuperPod{}
			fJob := &rescheduling.FaultJob{
				JobName:        "test-job",
				RescheduleTime: now - delayingTime - 1, // Exceeds threshold
				FaultTasks:     nil,
			}

			result := tp.isDelayingJob(fJob, nil)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestIsDelayingJobCase2(t *testing.T) {
	// Setup common test data
	now := time.Now().Unix()
	convey.Convey("Test isDelayingJob case2", t, func() {
		convey.Convey("When all non-fault tasks nodes are released", func() {
			tp := &module800SuperPod{}
			fJob := &rescheduling.FaultJob{
				JobName:        "test-job",
				RescheduleTime: now,
				FaultTasks: []rescheduling.FaultTask{
					{IsFaultTask: false, NodeName: "node1"},
					{IsFaultTask: true, NodeName: "node2"}, // Fault task should be skipped
				},
			}

			nodes := []*api.NodeInfo{
				{Name: "node1"}, // Node is available
			}

			patches := gomonkey.ApplyFunc(util.ChangeNodesToNodeMaps,
				func(nodes []*api.NodeInfo) map[string]*api.NodeInfo {
					return map[string]*api.NodeInfo{"node1": {}}
				})
			defer patches.Reset()

			result := tp.isDelayingJob(fJob, nodes)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("When some non-fault task nodes are not released", func() {
			tp := &module800SuperPod{}
			fJob := &rescheduling.FaultJob{
				JobName:        "test-job",
				RescheduleTime: now,
				FaultTasks: []rescheduling.FaultTask{
					{IsFaultTask: false, NodeName: "node1"},
					{IsFaultTask: false, NodeName: "node3"}, // Not in available nodes
				},
			}

			nodes := []*api.NodeInfo{
				{Name: "node1"},
			}

			patches := gomonkey.ApplyFunc(util.ChangeNodesToNodeMaps,
				func(_ []*api.NodeInfo) map[string]*api.NodeInfo {
					return map[string]*api.NodeInfo{"node1": {}}
				})
			defer patches.Reset()

			result := tp.isDelayingJob(fJob, nodes)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func newModuleSuperPod() *module800SuperPod {
	// Setup complete test environment
	baseHandler := base.NPUHandler{
		ScheduleEnv: plugin.ScheduleEnv{
			ClusterCache: plugin.ClusterCache{
				SuperPodInfo: &plugin.SuperPodInfo{
					SuperPodReschdInfo: make(map[api.JobID]map[string][]plugin.SuperNode),
				},
			},
		},
	}

	return &module800SuperPod{
		Base800ia5: ascend800ia5.Base800ia5{
			NPUHandler: baseHandler,
			VHandle:    &vnpu.VirtualNPU{},
		},
	}
}

func newNPUNodeWithSuperPodID(nodeName string, superPodID int32) plugin.NPUNode {
	return plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       nodeName,
			SuperPodID: superPodID,
		},
	}
}

const (
	spBlockNum2 = 2
)

type validNPUJobTest struct {
	name          string
	spBlockNPUNum int
	superPodSize  int
	reqNPUNum     int
	taskNum       int
	wantPass      bool
}

func buildValidNPUJobTestCases() []validNPUJobTest {
	return []validNPUJobTest{
		{
			name:          "01 will return false when spBlockNPUNum is 0",
			spBlockNPUNum: 0,
			wantPass:      false,
		},
		{
			name:          "02 will return false when spBlockNPUNum is 1 but superPodSize is 0 ",
			spBlockNPUNum: util.NPUIndex1,
			wantPass:      false,
		},
		{
			name:          "03 will return false when spBlockNPUNum is not mutiple of node npu",
			spBlockNPUNum: util.CoreNum25,
			wantPass:      false,
		},
		{
			name:          "04 will return false when require npu num is 1 and sp-block is not 1",
			reqNPUNum:     util.NPUIndex1,
			spBlockNPUNum: util.NPUIndex2,
			superPodSize:  util.NPUIndex1,
			wantPass:      false,
		},
		{
			name:          "05 will return false when 1 task job require npu num is 1 and sp-block is not 1",
			taskNum:       util.NPUIndex1,
			reqNPUNum:     util.NPUIndex1,
			spBlockNPUNum: util.NPUIndex2,
			superPodSize:  util.NPUIndex1,
			wantPass:      false,
		},
		{
			name:          "06 will return false when job require npu num is 3 not mutiple of die npu",
			taskNum:       util.NPUIndex1,
			reqNPUNum:     util.NPUIndex3,
			spBlockNPUNum: util.NPUIndex1,
			superPodSize:  util.NPUIndex1,
			wantPass:      false,
		},
	}
}

func TestValidNPUJob(t *testing.T) {
	for _, tt := range buildValidNPUJobTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			tp := &module800SuperPod{}
			tp.NPUJob = &util.NPUJob{}
			tp.MaxNodeNPUNum = 8
			tp.MaxCardNPUNum = 1
			tp.SpBlockNPUNum = tt.spBlockNPUNum
			tp.ReqNPUNum = tt.reqNPUNum
			tp.NPUTaskNum = tt.taskNum
			tp.FrameAttr.SuperPodSize = tt.superPodSize
			if got := tp.ValidNPUJob(); got != nil && !reflect.DeepEqual(got.Pass, tt.wantPass) {
				t.Errorf("ValidNPUJob() = %v, want %v", got.Pass, tt.wantPass)
			}
		})
	}
}

type CheckNodeNPUByTaskTest struct {
	name    string
	node    plugin.NPUNode
	task    *api.TaskInfo
	setup   func() *gomonkey.Patches
	wantErr bool
}

func buildCheckNodeNPUByTaskTestCases01() CheckNodeNPUByTaskTest {
	return CheckNodeNPUByTaskTest{
		name:    "01 will return err when task is nil",
		node:    plugin.NPUNode{},
		wantErr: true,
	}
}

func buildCheckNodeNPUByTaskTestCases02() CheckNodeNPUByTaskTest {
	node := plugin.NPUNode{}
	node.SuperPodID = util.ErrorInt
	node.Annotation = map[string]string{"test": "test"}
	return CheckNodeNPUByTaskTest{
		name:    "02 will return err when node SuperPodID is -1",
		node:    node,
		task:    &api.TaskInfo{},
		wantErr: true,
	}
}

func buildCheckNodeNPUByTaskTestCases03() CheckNodeNPUByTaskTest {
	node := plugin.NPUNode{}
	node.Annotation = map[string]string{"test": "test"}
	return CheckNodeNPUByTaskTest{
		name:    "03 will return err when GetTaskReqNPUNum return err",
		node:    node,
		task:    &api.TaskInfo{},
		wantErr: true,
	}
}

func buildCheckNodeNPUByTaskTestCases04() CheckNodeNPUByTaskTest {
	node := plugin.NPUNode{}
	node.Annotation = map[string]string{"test": "test"}
	return CheckNodeNPUByTaskTest{
		name: "04 will return err when GetUsableTopFromNode return err",
		node: node,
		task: &api.TaskInfo{},
		setup: func() *gomonkey.Patches {
			return gomonkey.ApplyMethod(reflect.TypeOf(&base.NPUHandler{}), "GetTaskReqNPUNum",
				func(_ *base.NPUHandler, _ *api.TaskInfo) (int, error) { return util.NPUIndex8, nil })
		},
		wantErr: true,
	}
}

func buildCheckNodeNPUByTaskTestCases05() CheckNodeNPUByTaskTest {
	node := plugin.NPUNode{}
	node.Annotation = map[string]string{"test": "test"}
	return CheckNodeNPUByTaskTest{
		name: "05 will return err when JudgeNodeAndTaskNPU return err",
		node: node,
		task: &api.TaskInfo{},
		setup: func() *gomonkey.Patches {
			patches := gomonkey.NewPatches()
			patches.ApplyMethod(reflect.TypeOf(&base.NPUHandler{}), "GetTaskReqNPUNum",
				func(_ *base.NPUHandler, _ *api.TaskInfo) (int, error) { return util.NPUIndex8, nil })
			patches.ApplyMethod(reflect.TypeOf(&base.NPUHandler{}), "GetUsableTopFromNode",
				func(_ *base.NPUHandler, node plugin.NPUNode, disFlag bool) ([]int, error) { return nil, nil })
			return patches
		},
		wantErr: true,
	}
}
func buildCheckNodeNPUByTaskTestCases06() CheckNodeNPUByTaskTest {
	node := plugin.NPUNode{}
	node.Annotation = map[string]string{"test": "test"}
	return CheckNodeNPUByTaskTest{
		name: "06 will return nil when node topo meet job require",
		node: node,
		task: &api.TaskInfo{},
		setup: func() *gomonkey.Patches {
			patches := gomonkey.NewPatches()
			patches.ApplyMethod(reflect.TypeOf(&base.NPUHandler{}), "GetTaskReqNPUNum",
				func(_ *base.NPUHandler, _ *api.TaskInfo) (int, error) { return util.NPUIndex2, nil })
			patches.ApplyMethod(reflect.TypeOf(&base.NPUHandler{}), "GetUsableTopFromNode",
				func(_ *base.NPUHandler, node plugin.NPUNode, disFlag bool) ([]int, error) {
					return []int{util.NPUIndex1, util.NPUIndex0}, nil
				})
			return patches
		},
		wantErr: false,
	}
}

func buildCheckNodeNPUByTaskTestCases() []CheckNodeNPUByTaskTest {
	return []CheckNodeNPUByTaskTest{
		buildCheckNodeNPUByTaskTestCases01(),
		buildCheckNodeNPUByTaskTestCases02(),
		buildCheckNodeNPUByTaskTestCases03(),
		buildCheckNodeNPUByTaskTestCases04(),
		buildCheckNodeNPUByTaskTestCases05(),
		buildCheckNodeNPUByTaskTestCases06(),
	}
}

func TestCheckNodeNPUByTask(t *testing.T) {
	for _, tt := range buildCheckNodeNPUByTaskTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				patches := tt.setup()
				defer patches.Reset()
			}
			tp := &module800SuperPod{spBlock: spBlockNum2}
			tp.NPUJob = &util.NPUJob{}
			tp.SetMaxNodeNPUNum(util.NPUIndex8)
			tp.SetMaxCardNPUNum(util.NPUIndex2)
			if err := tp.CheckNodeNPUByTask(tt.task, tt.node); (err != nil) != tt.wantErr {
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type ScoreBestNPUNodesTest struct {
	name    string
	nodes   []*api.NodeInfo
	spBlock int
	env     plugin.ScheduleEnv
	taskNum int
	wantErr bool
}

func buildScoreBestNPUNodesTest() []ScoreBestNPUNodesTest {
	ssn := test.FakeNormalSSN(nil)
	return []ScoreBestNPUNodesTest{
		{
			name:    "01 will return err when node list is nil",
			nodes:   nil,
			env:     *test2.FakeScheduleEnv(),
			wantErr: true,
		},
		{
			name:    "02 will return nil when job has 3 task and sp block is 1 select node success",
			nodes:   ssn.NodeList,
			env:     *test2.FakeScheduleEnv(),
			spBlock: util.NPUIndex1,
			taskNum: util.NPUIndex3,
			wantErr: false,
		},
		{
			name:    "03 will return nil when job has 3 task and sp block is 16 select node success",
			nodes:   ssn.NodeList,
			env:     *test2.FakeScheduleEnv(),
			spBlock: util.NPUIndex16,
			taskNum: util.NPUIndex3,
			wantErr: false,
		},
		{
			name:    "04 will return nil when job has 1 task and sp block is 16 select node success",
			nodes:   ssn.NodeList,
			env:     *test2.FakeScheduleEnv(),
			spBlock: util.NPUIndex16,
			taskNum: util.NPUIndex1,
			wantErr: false,
		},
	}
}

func TestScoreBestNPUNodes(t *testing.T) {
	tTask := test.FakeNormalTestTasks(1)[0]
	scoreMap := map[string]float64{}
	patch := gomonkey.ApplyFunc(rescheduling.GetReSchedulerCache, func() *rescheduling.DealReSchedulerCache {
		return &rescheduling.DealReSchedulerCache{
			FaultJobs: map[api.JobID]*rescheduling.FaultJob{"vcjob/pg0": {
				FaultTasks: []rescheduling.FaultTask{{NodeName: "node0", IsFaultTask: true}},
				JobUID:     "vcjob/pg0", IsFaultJob: true,
				SuperPods: map[string][]plugin.SuperNode{"0": {{Name: "node0", SuperPodID: 0}, {Name: "node1", SuperPodID: 0}}}}},
		}
	})
	defer patch.Reset()
	for _, tt := range buildScoreBestNPUNodesTest() {
		t.Run(tt.name, func(t *testing.T) {
			initScoreMap(scoreMap, tt.nodes)
			tp := &module800SuperPod{
				spBlock: tt.spBlock,
			}
			if err := tp.InitMyJobPlugin(tt.env.Jobs["vcjob/pg0"].SchedulerJobAttr, tt.env); err != nil {
				return
			}
			tp.Label[util.SinglePodTag] = util.EnableFunc
			tp.NPUTaskNum = tt.taskNum
			if err := tp.ScoreBestNPUNodes(tTask, tt.nodes, scoreMap); (err != nil) != tt.wantErr {
				t.Errorf("ScoreBestNPUNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func initScoreMap(sMap map[string]float64, nodes []*api.NodeInfo) {
	for _, node := range nodes {
		sMap[node.Name] = 0
	}
}

type isDelayingJobTest struct {
	name     string
	fJob     *rescheduling.FaultJob
	nodes    []*api.NodeInfo
	expected bool
}

func TestIsDelayingJob(t *testing.T) {
	now := time.Now().Unix()
	tests := []isDelayingJobTest{
		{
			name: "01 timeout case - should skip wait",
			fJob: &rescheduling.FaultJob{JobName: "test-job", RescheduleTime: now - util.NPUIndex11,
				FaultTasks: []rescheduling.FaultTask{{IsFaultTask: false, NodeName: "node1"}}},
			nodes:    []*api.NodeInfo{{Name: "node1"}},
			expected: true,
		},
		{
			name: "02 normal case - node released",
			fJob: &rescheduling.FaultJob{JobName: "test-job", RescheduleTime: now - util.NPUIndex5,
				FaultTasks: []rescheduling.FaultTask{{IsFaultTask: false, NodeName: "node1"}}},
			nodes:    []*api.NodeInfo{{Name: "node1"}},
			expected: true,
		},
		{
			name: "03 normal case - node not released",
			fJob: &rescheduling.FaultJob{JobName: "test-job", RescheduleTime: now - util.NPUIndex5,
				FaultTasks: []rescheduling.FaultTask{{IsFaultTask: false, NodeName: "node2"}}},
			nodes:    []*api.NodeInfo{{Name: "node1"}},
			expected: false,
		},
		{
			name: "04 edge case - no fault tasks",
			fJob: &rescheduling.FaultJob{JobName: "test-job", RescheduleTime: now - util.NPUIndex5,
				FaultTasks: []rescheduling.FaultTask{}},
			nodes:    []*api.NodeInfo{{Name: "node1"}},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &module800SuperPod{}
			result := tp.isDelayingJob(tt.fJob, tt.nodes)
			if result != tt.expected {
				t.Errorf("isDelayingJob() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	convey.Convey("test New func", t, func() {
		name := "testName"
		ascendHandler := New(name)
		convey.So(ascendHandler.GetPluginName(), convey.ShouldEqual, name)
	})
}

type checkRequireNPUTest struct {
	name     string
	module   *module800SuperPod
	expected *api.ValidateResult
}

func TestCheckReqNPUEqualNodeNPU(t *testing.T) {
	module1 := module800SuperPod{}
	module1.NPUJob = &util.NPUJob{
		Tasks: map[api.TaskID]util.NPUTask{api.TaskID(strconv.Itoa(1)): {
			ReqNPUNum: 8,
		}}}
	module2 := module800SuperPod{}
	module2.NPUJob = &util.NPUJob{
		Tasks: map[api.TaskID]util.NPUTask{api.TaskID(strconv.Itoa(1)): {
			ReqNPUNum: 3,
		}}}
	tests := []checkRequireNPUTest{
		{
			name:     "01 require task is ok",
			module:   &module1,
			expected: nil,
		},
		{
			name:   "02 require task is err",
			module: &module2,
			expected: &api.ValidateResult{
				Pass:    false,
				Reason:  jobCheckFailedReason,
				Message: fmt.Sprintf("distributed super-pod job require npu 8*n, instead of 3"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.module.checkReqNPUEqualNodeNPU()
			if reflect.DeepEqual(result, tt.expected) == false {
				t.Errorf("checkReqNPUEqualNodeNPU() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCheckReqNPU(t *testing.T) {
	tests := getCheckRequireNPUTestParam1()
	tests = append(tests, getCheckRequireNPUTestParam2()...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.module.checkRequireNPU()
			if reflect.DeepEqual(result, tt.expected) == false {
				t.Errorf("checkRequireNPU() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func getCheckRequireNPUTestParam2() []checkRequireNPUTest {
	module4 := module800SuperPod{}
	module4.NPUJob = &util.NPUJob{
		NPUTaskNum:    2,
		ReqNPUNum:     10,
		SpBlockNPUNum: 8,
	}
	module5 := module800SuperPod{}
	module5.NPUJob = &util.NPUJob{
		NPUTaskNum:    2,
		ReqNPUNum:     16,
		SpBlockNPUNum: 8,
		Tasks: map[api.TaskID]util.NPUTask{api.TaskID(strconv.Itoa(1)): {
			ReqNPUNum: 8,
		}},
	}
	tests := []checkRequireNPUTest{
		{
			name:   "04 Distributed job reqNPUNum not be multiple of spBlockNPUNum, should return err",
			module: &module4,
			expected: &api.ValidateResult{
				Pass:    false,
				Reason:  jobCheckFailedReason,
				Message: "distributed super-pod job require npu(10) should be multiple of sp-block",
			},
		},
		{
			name:     "05 RequireNPU is ok, should return nil",
			module:   &module5,
			expected: nil,
		},
	}
	return tests
}

func getCheckRequireNPUTestParam1() []checkRequireNPUTest {
	module1 := module800SuperPod{}
	module1.NPUJob = &util.NPUJob{
		NPUTaskNum:    1,
		ReqNPUNum:     1,
		SpBlockNPUNum: 3,
	}
	module2 := module800SuperPod{}
	module2.NPUJob = &util.NPUJob{
		NPUTaskNum:    1,
		ReqNPUNum:     1,
		SpBlockNPUNum: 1,
	}
	module3 := module800SuperPod{}
	module3.NPUJob = &util.NPUJob{
		NPUTaskNum: 1,
		ReqNPUNum:  9,
	}
	tests := []checkRequireNPUTest{
		{
			name:   "01 ReqNPUNum not equals spBlockNPUNum, should return err",
			module: &module1,
			expected: &api.ValidateResult{
				Pass:    false,
				Reason:  jobCheckFailedReason,
				Message: "single super-pod job sp-block annotation should equal require npu num",
			},
		},
		{
			name:     "02 ReqNPUNum equals spBlockNPUNum, should return nil",
			module:   &module2,
			expected: nil,
		},
		{
			name:   "03 ReqNPUNum gt spBlockNPUNum, should return err",
			module: &module3,
			expected: &api.ValidateResult{
				Pass:    false,
				Reason:  jobCheckFailedReason,
				Message: "single super-pod job require npu [1, 2*n], instead of 9",
			},
		},
	}
	return tests
}
