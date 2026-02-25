/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package chip8node8ra64sp for node test
package chip8node8ra64sp

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	itest "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

// for test cases use
const (
	rackOsNum      = 8
	npuNum8        = 8
	nodeInfoIdx4   = 4
	nodeInfoIdx9   = 9
	nodeInfoIdx185 = 185
)

var (
	npuList2        = []int{0, 1}
	npuList8        = []int{0, 1, 2, 3, 4, 5, 6, 7}
	testRackNPUTop  = rackNpuTopType{}
	testRackNPUTop1 = rackNpuTopType{
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
	}
	testRackNPUTop2 = rackNpuTopType{
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
	}
	testAnnoName         = "test"
	invalidTestAnnoValue = "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4," +
		"Ascend910-5,Ascend910-6,Ascend910-7,Ascend910-8"
)

// checkNodeNPUByTaskTestCase CheckNodeNPUByTask test case
type checkNodeNPUByTaskTestCase struct {
	Task          *api.TaskInfo
	Name          string
	Attr          util.SchedulerJobAttr
	Node          plugin.NPUNode
	WantErr       error
	TpBlockNPUNum int
	MaxNodeNPUNum int
	TaskNodeNPU   string
}

func buildCardAnnotationStr(npuList []int) string {
	var str string
	for index, npu := range npuList {
		if index == 0 {
			str = fmt.Sprintf("%s%d", util.NPU910CardNamePre, npu)
		} else {
			str = fmt.Sprintf("%s,%s%d", str, util.NPU910CardNamePre, npu)
		}
	}
	return str
}

// TestCheckNodeNPUByTask
func TestCheckNodeNPUByTask(t *testing.T) {
	npu := New(SuperPodx8SchedulerName)
	testCases := buildcheckNodeNPUByTaskTestCases01()
	testCases = append(testCases, buildcheckNodeNPUByTaskTestCases02()...)
	testCases = append(testCases, buildcheckNodeNPUByTaskTestCases03()...)
	testCases = append(testCases, buildcheckNodeNPUByTaskTestCases04()...)
	testCases = append(testCases, buildcheckNodeNPUByTaskTestCases05()...)
	testCases = append(testCases, buildcheckNodeNPUByTaskTestCases06()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			job := test.FakeNormalTestJob("job", 1)
			test.SetFakeJobResRequest(job, util.NPU910CardName, tt.TaskNodeNPU)
			attr := itest.FakeSchedulerJobAttrByJob(job)
			sJob := plugin.SchedulerJob{}
			sJob.SchedulerJobAttr = attr
			env := plugin.ScheduleEnv{
				ClusterCache: plugin.ClusterCache{
					Jobs: map[api.JobID]plugin.SchedulerJob{job.UID: sJob},
				},
			}
			npu.SetSchedulerAttr(attr)
			npu.SetSchedulerEnv(env)
			npu.TpBlockNPUNum = tt.TpBlockNPUNum

			if err := npu.CheckNodeNPUByTask(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

func buildcheckNodeNPUByTaskTestCases01() []checkNodeNPUByTaskTestCase {
	return []checkNodeNPUByTaskTestCase{
		{
			Name:          "01-CheckNodeNPUByTask return nil when node npu meet task req",
			Task:          test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex8),
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPU910CardName: buildCardAnnotationStr(npuList8),
						networkUnhealthyNPU: "",
					},
				},
			},
			WantErr: nil,
		},
		{
			Name:          "02-CheckNodeNPUByTask return err when task is not npu task",
			Task:          test.FakeTaskWithResReq("pod1", util.NPU910CardName, util.NPUIndex8),
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPU910CardName: buildCardAnnotationStr(npuList8),
						networkUnhealthyNPU: "",
					},
				},
			},
			WantErr: errors.New("task<pod1> is not npu task"),
		},
	}
}

func buildcheckNodeNPUByTaskTestCases02() []checkNodeNPUByTaskTestCase {
	return []checkNodeNPUByTaskTestCase{
		{
			Name:          "04-CheckNodeNPUByTask return err when node has no req npu",
			Task:          test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex8),
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPU910CardName: buildCardAnnotationStr(npuList2),
						networkUnhealthyNPU: "",
					},
				},
			},
			WantErr: errors.New("checkNodeNPUByTask the npus on this node don't satisfy the schedulable topology err: node don't have enough npu resource, req<8>, idle<2>"),
		},
		{
			Name:          "05-CheckNodeNPUByTask return err when node has no req npu",
			Task:          test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex8),
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-4",
					},
				},
			},
			WantErr: errors.New("node<node1> don't have resource<huawei.com/Ascend910-NetworkUnhealthy>"),
		},
	}
}

func buildcheckNodeNPUByTaskTestCases03() []checkNodeNPUByTaskTestCase {
	return []checkNodeNPUByTaskTestCase{
		{
			Name:          "07-CheckNodeNPUByTask return err when task is nil",
			Task:          nil,
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
				},
			},
			WantErr: errors.New(util.ArgumentError),
		},
		{
			Name:          "08-CheckNodeNPUByTask return err when node annotation is nil",
			Task:          test.FakeTaskWithResReq("pod1", util.NPU310PCardName, util.NPUIndex2),
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: nil,
				},
			},
			WantErr: errors.New(util.ArgumentError),
		},
	}
}

func buildcheckNodeNPUByTaskTestCases04() []checkNodeNPUByTaskTestCase {
	task := test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex8)
	return []checkNodeNPUByTaskTestCase{
		{
			Name:          "9-CheckNodeNPUByTask return nil when tp-block is valid",
			Task:          task,
			TpBlockNPUNum: 8,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPU910CardName: buildCardAnnotationStr(npuList8),
						networkUnhealthyNPU: "",
					},
				},
			},
			WantErr: nil,
		},
	}
}

func buildcheckNodeNPUByTaskTestCases05() []checkNodeNPUByTaskTestCase {
	task := test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex8)
	return []checkNodeNPUByTaskTestCase{
		{
			Name:          "10-CheckNodeNPUByTask return err when node has no req npu",
			Task:          task,
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5," +
							"Ascend910-6,Ascend910-7",
						networkUnhealthyNPU: "Ascend910-5",
					},
				},
			},
			WantErr: errors.New("checkNodeNPUByTask the npus on this node don't satisfy the schedulable topology err: node don't have enough npu resource, req<8>, idle<7>"),
		},
	}
}

func buildcheckNodeNPUByTaskTestCases06() []checkNodeNPUByTaskTestCase {
	task := test.FakeTaskWithResReq("pod0", util.NPU910CardName, npuNumber8)
	return []checkNodeNPUByTaskTestCase{
		{
			Name:          "11-CheckNodeNPUByTask return err when node RackID invalid",
			Task:          task,
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5," +
							"Ascend910-6,Ascend910-7",
					},
					RackID: -1,
				},
			},
			WantErr: errors.New("node rack-id is invalid for node=node1, id=-1"),
		},
		{
			Name:          "16-CheckNodeNPUByTask return err when node SuperPodID invalid",
			Task:          task,
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5," +
							"Ascend910-6,Ascend910-7",
					},
					SuperPodID: -1,
				},
			},
			WantErr: errors.New("the super-pod-id of node is invalid for node=node1, id=-1"),
		},
	}
}

func buildNodeBaseInfoArr(count int) []nodeBaseInfo {
	var nodes []nodeBaseInfo
	for i := 0; i < count; i++ {
		nodes = append(nodes, nodeBaseInfo{
			name:       fmt.Sprintf("node%d", i),
			superPodID: 1,
			rackID:     1,
		})
	}
	return nodes
}

type getUsableTopFromNode struct {
	Name    string
	Node    plugin.NPUNode
	WantErr error
}

func buildGetUsableTopFromNodeTest1() []getUsableTopFromNode {
	return []getUsableTopFromNode{
		{
			Name:    "Case1-annotation is empty",
			WantErr: errors.New(util.ArgumentError),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{},
				},
			},
		},
		{
			Name:    "Case2-annotation without the key of huawei.com/Ascend910",
			WantErr: fmt.Errorf("getUsableTopFromNode node1 don't have %s", util.NPU910CardName),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						testAnnoName: "Ascend910-0,Ascend910-1",
					},
				},
			},
		},
		{
			Name:    "Case3-node npu num more than 8",
			WantErr: fmt.Errorf("node npu num is invalid, and the npus index: [0 1 2 3 4 5 6 7 8]"),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPU910CardName: invalidTestAnnoValue,
					},
				},
			},
		},
	}
}

func buildGetUsableTopFromNodeTest2() []getUsableTopFromNode {
	return []getUsableTopFromNode{
		{
			Name:    "Case4-node networkUnHealthy key is empty",
			WantErr: fmt.Errorf("node<node1> don't have resource<%s>", networkUnhealthyNPU),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPU910CardName: "Ascend910-0,Ascend910-1",
					},
				},
			},
		},
		{
			Name:    "Case5-node networkUnHealthy npu num is more than 8",
			WantErr: fmt.Errorf("node<node1> npu networkUnhealthy top<[0 1 2 3 4 5 6 7 8]> is invalid"),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPU910CardName: "Ascend910-0,Ascend910-1",
						networkUnhealthyNPU: invalidTestAnnoValue,
					},
				},
			},
		},
	}
}

func TestGetUsableTopFromNode(t *testing.T) {
	npu := New(SuperPodx8SchedulerName)
	// Set up NPUJob with ReqNPUName to avoid nil pointer
	npu.SetSchedulerAttr(util.SchedulerJobAttr{
		NPUJob: &util.NPUJob{
			ReqNPUName: util.NPU910CardName,
		},
	})
	testCases := buildGetUsableTopFromNodeTest1()
	testCases = append(testCases, buildGetUsableTopFromNodeTest2()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if _, err := npu.getUsableTopFromNode(tt.Node); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("ValidNPUJob() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}
