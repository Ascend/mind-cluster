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
package chip8node8sp

import (
	"errors"
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
			result := (*chip8node8sp)(nil).isDelayingJob(nil, nil)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When fJob is nil", func() {
			tp := &chip8node8sp{}
			result := tp.isDelayingJob(nil, nil)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When FaultTasks is nil", func() {
			tp := &chip8node8sp{}
			fJob := &rescheduling.FaultJob{FaultTasks: nil}
			result := tp.isDelayingJob(fJob, nil)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When waiting time exceeds threshold", func() {
			tp := &chip8node8sp{}
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
			tp := &chip8node8sp{}
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
			tp := &chip8node8sp{}
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

func newModuleSuperPod() *chip8node8sp {
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

	return &chip8node8sp{
		NPUHandler: baseHandler,
		VHandle:    &vnpu.VirtualNPU{},
	}
}

func TestSelectNodeFromOriginVSuperPodCase1(t *testing.T) {
	convey.Convey("Test selectNodeFromOriginVSuperPod with full context", t, func() {
		// Setup complete test environment
		tp := newModuleSuperPod()

		fJob := &rescheduling.FaultJob{JobUID: "job-123", SuperPods: make(map[string][]plugin.SuperNode)}
		sMap := map[string]float64{"healthy-pod": 0.9, "faulty-pod": 0.1}
		selectNodes := make(map[string][]plugin.SuperNode)
		totalNodes := map[int32]superPod{
			1: {"healthy-pod": {}, "faulty-pod": {}},
		}
		vSuperPodID := make(map[string]bool)

		// Mock dependencies
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		convey.Convey("With complete structure initialization", func() {
			convey.Convey("Should handle nil inputs properly", func() {
				_, err := (*chip8node8sp)(nil).selectNodeFromOriginVSuperPod(fJob, sMap, selectNodes, totalNodes, vSuperPodID)
				convey.So(err, convey.ShouldNotBeNil)

				_, err = tp.selectNodeFromOriginVSuperPod(nil, sMap, selectNodes, totalNodes, vSuperPodID)
				convey.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestSelectNodeFromOriginVSuperPodCase2(t *testing.T) {
	convey.Convey("Test selectNodeFromOriginVSuperPod with full context case 1", t, func() {
		// Setup complete test environment
		tp := newModuleSuperPod()

		fJob := &rescheduling.FaultJob{JobUID: "job-123", SuperPods: make(map[string][]plugin.SuperNode)}
		sMap := map[string]float64{"healthy-pod": 0.9, "faulty-pod": 0.1}
		selectNodes := make(map[string][]plugin.SuperNode)
		totalNodes := map[int32]superPod{
			1: {"healthy-pod": {}, "faulty-pod": {}},
		}
		vSuperPodID := make(map[string]bool)

		// Mock dependencies
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		convey.Convey("With complete structure initialization", func() {
			convey.Convey("Should handle faulty super pods correctly", func() {
				superPodID := "faulty-sp"
				faultyPods := []plugin.SuperNode{
					{Name: "faulty-pod", SuperPodID: 1},
				}
				fJob.SuperPods[superPodID] = faultyPods

				patches.ApplyFunc(judgeLasTimeTaskIsHealthy, func(_ *rescheduling.FaultJob, _ string) bool {
					return false
				})

				result, err := tp.selectNodeFromOriginVSuperPod(fJob, sMap, selectNodes, totalNodes, vSuperPodID)
				convey.So(err, convey.ShouldBeNil)
				convey.So(result, convey.ShouldContainKey, superPodID)
				convey.So(selectNodes, convey.ShouldBeEmpty)
			})

			convey.Convey("Should handle mixed healthy/faulty pods in super pod", func() {
				superPodID := "mixed-sp"
				mixedPods := []plugin.SuperNode{
					{Name: "healthy-pod", SuperPodID: 0},
					{Name: "faulty-pod", SuperPodID: 1},
				}
				fJob.SuperPods[superPodID] = mixedPods

				patches.ApplyFunc(judgeLasTimeTaskIsHealthy, func(_ *rescheduling.FaultJob, name string) bool {
					return name == "healthy-pod"
				})

				result, err := tp.selectNodeFromOriginVSuperPod(fJob, sMap, selectNodes, totalNodes, vSuperPodID)
				convey.So(err, convey.ShouldBeNil)
				convey.So(result, convey.ShouldContainKey, superPodID)
				convey.So(selectNodes, convey.ShouldBeEmpty)
			})
		})
	})
}

func TestJudgeLasTimeTaskIsHealthyCase1(t *testing.T) {
	convey.Convey("Test judgeLasTimeTaskIsHealthy case 1", t, func() {
		// Test cases for nil checks
		convey.Convey("When fJob is nil", func() {
			result := judgeLasTimeTaskIsHealthy(nil, "node1")
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When FaultTasks is nil", func() {
			fJob := &rescheduling.FaultJob{FaultTasks: nil}
			result := judgeLasTimeTaskIsHealthy(fJob, "node1")
			convey.So(result, convey.ShouldBeFalse)
		})

		// Test cases with actual FaultTasks
		convey.Convey("With actual FaultTasks", func() {
			fJob := &rescheduling.FaultJob{
				FaultTasks: []rescheduling.FaultTask{
					{NodeName: "node1", IsFaultTask: true},
					{NodeName: "node2", IsFaultTask: false},
					{NodeName: "node3", IsFaultTask: true},
				},
			}

			convey.Convey("When node is fault task", func() {
				result := judgeLasTimeTaskIsHealthy(fJob, "node1")
				convey.So(result, convey.ShouldBeFalse)

				result = judgeLasTimeTaskIsHealthy(fJob, "node3")
				convey.So(result, convey.ShouldBeFalse)
			})

			convey.Convey("When node is not fault task", func() {
				result := judgeLasTimeTaskIsHealthy(fJob, "node2")
				convey.So(result, convey.ShouldBeTrue)
			})

			convey.Convey("When node is not in FaultTasks", func() {
				result := judgeLasTimeTaskIsHealthy(fJob, "node4")
				convey.So(result, convey.ShouldBeTrue)
			})

			convey.Convey("When multiple tasks have same node name", func() {
				// Add another task with same node name but different fault status
				fJob.FaultTasks = append(fJob.FaultTasks, rescheduling.FaultTask{
					NodeName:    "node1",
					IsFaultTask: false,
				})

				// Should return false because first matching task is fault
				result := judgeLasTimeTaskIsHealthy(fJob, "node1")
				convey.So(result, convey.ShouldBeFalse)
			})
		})
	})
}

func TestJudgeLasTimeTaskIsHealthyCase2(t *testing.T) {
	convey.Convey("Test judgeLasTimeTaskIsHealthy case 2", t, func() {
		convey.Convey("Edge cases", func() {
			convey.Convey("Empty FaultTasks list", func() {
				fJob := &rescheduling.FaultJob{
					FaultTasks: []rescheduling.FaultTask{},
				}
				result := judgeLasTimeTaskIsHealthy(fJob, "any-node")
				convey.So(result, convey.ShouldBeTrue)
			})

			convey.Convey("Empty node name", func() {
				fJob := &rescheduling.FaultJob{
					FaultTasks: []rescheduling.FaultTask{
						{NodeName: "", IsFaultTask: true},
					},
				}
				result := judgeLasTimeTaskIsHealthy(fJob, "")
				convey.So(result, convey.ShouldBeFalse)
			})
		})
	})
}

const (
	num1 = 1
	num2 = 2
	num3 = 3
)

func TestSchedulableCase1(t *testing.T) {
	convey.Convey("Test schedulable case 1", t, func() {
		// Setup base test environment
		tp := &chip8node8sp{}
		fJob := &rescheduling.FaultJob{
			JobUID: "job-123",
			SuperPods: map[string][]plugin.SuperNode{
				"sp1": {
					{Name: "task1", SuperPodID: 1},
					{Name: "task2", SuperPodID: 1},
				},
			},
			FaultTasks: []rescheduling.FaultTask{
				{TaskName: "task1", IsFaultTask: true},
				{TaskName: "task2", IsFaultTask: true},
				{TaskName: "task3", IsFaultTask: false},
			},
		}

		totalNodes := map[int32]superPod{
			1: {"node1": {}, "node2": {}},
		}

		// Mock dependencies
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		convey.Convey("When tp or fJob is nil", func() {
			convey.So((*chip8node8sp)(nil).schedulable(fJob, totalNodes), convey.ShouldBeFalse)
			convey.So(tp.schedulable(nil, totalNodes), convey.ShouldBeFalse)
		})

		convey.Convey("Edge cases", func() {
			convey.Convey("Empty SuperPods", func() {
				fJob.SuperPods = make(map[string][]plugin.SuperNode)
				result := tp.schedulable(fJob, totalNodes)
				convey.So(result, convey.ShouldBeFalse)
			})

			convey.Convey("Empty FaultTasks", func() {
				fJob.FaultTasks = []rescheduling.FaultTask{}
				result := tp.schedulable(fJob, totalNodes)
				convey.So(result, convey.ShouldBeTrue)
			})
		})
	})
}

func TestIfScheduleCase1(t *testing.T) {
	convey.Convey("Test ifSchedule case 2", t, func() {
		convey.Convey("When count or totalNodes is empty", func() {
			convey.Convey("Empty count map", func() {
				result := ifSchedule(map[int32]int{}, map[int32]superPod{1: {"node1": {}}})
				convey.So(result, convey.ShouldBeFalse)
			})

			convey.Convey("Empty totalNodes map", func() {
				result := ifSchedule(map[int32]int{1: 1}, map[int32]superPod{})
				convey.So(result, convey.ShouldBeFalse)
			})

			convey.Convey("Both empty", func() {
				result := ifSchedule(map[int32]int{}, map[int32]superPod{})
				convey.So(result, convey.ShouldBeFalse)
			})
		})

		convey.Convey("When resources are sufficient", func() {
			count := map[int32]int{
				num1: num2, // Need 2 nodes
				num2: num1, // Need 1 node
			}
			totalNodes := map[int32]superPod{
				num1: {"node1": {}, "node2": {}, "node3": {}}, // 3 available
				num2: {"node4": {}},                           // 1 available
			}

			result := ifSchedule(count, totalNodes)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("When resources are insufficient", func() {
			convey.Convey("For one super pod", func() {
				count := map[int32]int{
					num1: num3, // Need 3 nodes
				}
				totalNodes := map[int32]superPod{
					num1: {"node1": {}, "node2": {}}, // Only 2 available
				}

				result := ifSchedule(count, totalNodes)
				convey.So(result, convey.ShouldBeFalse)
			})

			convey.Convey("For multiple super pods", func() {
				count := map[int32]int{
					num1: num2, // Need 2 nodes (OK)
					num2: num2, // Need 2 nodes (Only 1 available)
				}
				totalNodes := map[int32]superPod{
					num1: {"node1": {}, "node2": {}},
					num2: {"node3": {}},
				}

				result := ifSchedule(count, totalNodes)
				convey.So(result, convey.ShouldBeFalse)
			})
		})
	})
}

func TestIfScheduleCase2(t *testing.T) {
	convey.Convey("Test ifSchedule", t, func() {
		convey.Convey("When super pod ID not found in totalNodes", func() {
			count := map[int32]int{
				num1: num1, // Need 1 node
				num3: num1, // Super pod 3 doesn't exist
			}
			totalNodes := map[int32]superPod{
				num1: {"node1": {}},
				num2: {"node2": {}},
			}

			result := ifSchedule(count, totalNodes)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Edge cases", func() {
			convey.Convey("Exact number of required nodes", func() {
				count := map[int32]int{
					num1: num2, // Need exactly 2 nodes
				}
				totalNodes := map[int32]superPod{
					num1: {"node1": {}, "node2": {}}, // Exactly 2 available
				}

				result := ifSchedule(count, totalNodes)
				convey.So(result, convey.ShouldBeTrue)
			})

			convey.Convey("Zero required nodes", func() {
				count := map[int32]int{
					num1: 0, // Need 0 nodes
				}
				totalNodes := map[int32]superPod{
					num1: {"node1": {}}, // 1 available (but not needed)
				}

				result := ifSchedule(count, totalNodes)
				convey.So(result, convey.ShouldBeTrue)
			})
		})
	})
}

func TestSelectNodeForPodLevelRescheduling(t *testing.T) {
	convey.Convey("Test selectNodeForPodLevelRescheduling", t, func() {
		tp := &chip8node8sp{
			spBlock: 1,
		}

		convey.Convey("When fJob is nil", func() {
			tp.selectNodeForPodLevelRescheduling(nil, nil, nil, nil, nil)
		})

		convey.Convey("When fJob.SuperPods is nil", func() {
			fJob := &rescheduling.FaultJob{}
			tp.selectNodeForPodLevelRescheduling(fJob, nil, nil, nil, nil)
		})

		convey.Convey("When selectNodes or vSuperPodID is nil", func() {
			fJob := &rescheduling.FaultJob{
				SuperPods: map[string][]plugin.SuperNode{},
			}
			tp.selectNodeForPodLevelRescheduling(fJob, nil, nil, nil, map[string][]plugin.SuperNode{})
			tp.selectNodeForPodLevelRescheduling(fJob, nil, nil, map[string]bool{}, nil)
		})

		convey.Convey("When notReadySuperPod contains an unknown ID", func() {
			fJob := &rescheduling.FaultJob{
				SuperPods: map[string][]plugin.SuperNode{
					"sp1": {},
				},
			}
			notReady := map[string]struct{}{"sp2": {}}
			selectNodes := make(map[string][]plugin.SuperNode)
			vSuperPodID := make(map[string]bool)

			tp.selectNodeForPodLevelRescheduling(fJob, notReady, nil, vSuperPodID, selectNodes)
			convey.So(len(selectNodes), convey.ShouldEqual, 0)
		})

		convey.Convey("When fault task IDs are empty", func() {
			mockSuperPodID := int32(10)
			superNode := plugin.SuperNode{Name: "n1", SuperPodID: mockSuperPodID}
			fJob := &rescheduling.FaultJob{
				SuperPods: map[string][]plugin.SuperNode{
					"sp1": {superNode},
				},
			}
			notReady := map[string]struct{}{"sp1": {}}
			selectNodes := make(map[string][]plugin.SuperNode)
			vSuperPodID := make(map[string]bool)

			tp.selectNodeForPodLevelRescheduling(fJob, notReady, nil, vSuperPodID, selectNodes)

			convey.So(selectNodes["sp1"], convey.ShouldResemble, fJob.SuperPods["sp1"])
			convey.So(vSuperPodID["sp1"], convey.ShouldBeTrue)
		})
	})
}

func TestCheckSpBlockGtZero(t *testing.T) {
	const negative = -1
	tests := []struct {
		name     string
		spBlock  int
		expected bool
	}{
		{
			name:     "spBlock > 0",
			spBlock:  10,
			expected: true,
		},
		{
			name:     "spBlock == 0",
			spBlock:  0,
			expected: false,
		},
		{
			name:     "spBlock < 0",
			spBlock:  negative,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &chip8node8sp{
				spBlock: tt.spBlock,
			}
			actual := tp.checkSpBlockGtZero()
			if actual != tt.expected {
				t.Errorf("checkSpBlockGtZero() = %v; want %v", actual, tt.expected)
			}
		})
	}
}

func TestInitRemainderTop(t *testing.T) {
	tp := &chip8node8sp{
		spBlock: 3,
	}
	tp.FrameAttr.SuperPodSize = 20
	result := tp.initRemainderTop()

	expectedLen := tp.spBlock
	expectedInnerLen := tp.FrameAttr.SuperPodSize/tp.spBlock + 1

	if len(result) != expectedLen {
		t.Errorf("expected outer length %d, got %d", expectedLen, len(result))
	}

	for i := range result {
		if len(result[i]) != expectedInnerLen {
			t.Errorf("expected inner length at index %d: %d, got %d", i, expectedInnerLen, len(result[i]))
		}
	}
}

func TestSelectNodeFromOriginSuperPod(t *testing.T) {
	tp := &chip8node8sp{
		spBlock: 1,
	}

	superPods := map[string][]plugin.SuperNode{
		"vsp1": {
			plugin.SuperNode{Name: "n1", SuperPodID: 100},
		},
		"vsp2": {
			plugin.SuperNode{Name: "n2", SuperPodID: 200},
		},
	}

	totalNodes := map[int32]superPod{
		100: {"n1": plugin.NPUNode{}},
		200: {"n2": plugin.NPUNode{}},
	}

	notReadySuperPod := map[string]struct{}{
		"vsp1": {},
		"vsp2": {},
	}

	vSuperPodID := make(map[string]bool)
	selectNodes := make(map[string][]plugin.SuperNode)

	fJob := &rescheduling.FaultJob{
		SuperPods: superPods,
	}

	tp.selectNodeFromOriginSuperPod(fJob, notReadySuperPod, totalNodes, vSuperPodID, selectNodes)

	if !vSuperPodID["vsp1"] || !vSuperPodID["vsp2"] {
		t.Errorf("Expected vSuperPodID to contain vsp1 and vsp2")
	}
	if len(selectNodes["vsp1"]) != 1 || selectNodes["vsp1"][0].Name != "n1" {
		t.Errorf("vsp1 selectNodes mismatch")
	}
	if len(selectNodes["vsp2"]) != 1 || selectNodes["vsp2"][0].Name != "n2" {
		t.Errorf("vsp2 selectNodes mismatch")
	}
	if _, exists := totalNodes[100]["n1"]; exists {
		t.Errorf("Expected n1 to be deleted from totalNodes[100]")
	}
	if _, exists := totalNodes[200]["n2"]; exists {
		t.Errorf("Expected n2 to be deleted from totalNodes[200]")
	}
}

func TestSelectNodeFromOriginSuperPodNilOrEmptyInputs(t *testing.T) {
	var tpNil *chip8node8sp
	tpNil.selectNodeFromOriginSuperPod(nil, nil, nil, nil, nil)

	tp := &chip8node8sp{}
	tp.selectNodeFromOriginSuperPod(nil, nil, nil, nil, nil)

	fJobNilPods := &rescheduling.FaultJob{}
	tp.selectNodeFromOriginSuperPod(fJobNilPods, nil, nil, nil, nil)

	fJobEmptyPods := &rescheduling.FaultJob{SuperPods: map[string][]plugin.SuperNode{}}
	tp.selectNodeFromOriginSuperPod(fJobEmptyPods, nil, nil, nil, nil)
}

func TestSelectNodeFromOriginSuperPodNilMaps(t *testing.T) {
	tp := &chip8node8sp{}

	fJob := &rescheduling.FaultJob{
		SuperPods: map[string][]plugin.SuperNode{
			"vsp": {plugin.SuperNode{Name: "n1", SuperPodID: 100}},
		},
	}

	notReadySuperPod := map[string]struct{}{"vsp": {}}
	totalNodes := map[int32]superPod{
		100: {"n1": plugin.NPUNode{}},
	}

	tp.selectNodeFromOriginSuperPod(fJob, notReadySuperPod, totalNodes, map[string]bool{}, nil)

	tp.selectNodeFromOriginSuperPod(fJob, notReadySuperPod, totalNodes, nil, map[string][]plugin.SuperNode{})
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
			tp := &chip8node8sp{}
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
			tp := &chip8node8sp{spBlock: spBlockNum2}
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
			tp := &chip8node8sp{
				spBlock:    tt.spBlock,
				nodeVPodId: map[string]string{},
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

type getSelectNodesTest struct {
	name         string
	fNodeNameMap map[string]struct{}
	spNodes      []plugin.SuperNode
	spNodeMaps   map[string]plugin.NPUNode
	want         []plugin.SuperNode
}

func buildGetSelectNodesTest() []getSelectNodesTest {
	return []getSelectNodesTest{
		{
			name: "01 will return nil when spNodeMaps is nil",
			want: nil,
		},
		{
			name:         "02 test node is not exist in fNodeNameMap",
			spNodes:      []plugin.SuperNode{{Name: "node0", SuperPodID: 0}},
			fNodeNameMap: nil,
			spNodeMaps:   map[string]plugin.NPUNode{"node0": {}},
			want:         []plugin.SuperNode{{Name: "node0", SuperPodID: 0}},
		},
		{
			name:         "03 test node is exist in fNodeNameMap",
			spNodes:      []plugin.SuperNode{{Name: "node0", SuperPodID: 0}},
			fNodeNameMap: map[string]struct{}{"node0": {}},
			spNodeMaps:   map[string]plugin.NPUNode{"node0": {}},
			want:         []plugin.SuperNode{{Name: "", SuperPodID: 0}},
		},
	}
}

func TestGetSelectNodes(t *testing.T) {
	for _, tt := range buildGetSelectNodesTest() {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSelectNodes(tt.fNodeNameMap, tt.spNodes, tt.spNodeMaps); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSelectNodes() = %v, want %v", got, tt.want)
			}
		})
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
			tp := &chip8node8sp{}
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
	module   *chip8node8sp
	expected *api.ValidateResult
}

func TestCheckReqNPUEqualNodeNPU(t *testing.T) {
	module1 := chip8node8sp{}
	module1.NPUJob = &util.NPUJob{
		Tasks: map[api.TaskID]util.NPUTask{api.TaskID(strconv.Itoa(1)): {
			ReqNPUNum: 8,
		}}}
	module2 := chip8node8sp{}
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
	module4 := chip8node8sp{}
	module4.NPUJob = &util.NPUJob{
		NPUTaskNum:    2,
		ReqNPUNum:     10,
		SpBlockNPUNum: 8,
	}
	module5 := chip8node8sp{}
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
	module1 := chip8node8sp{}
	module1.NPUJob = &util.NPUJob{
		NPUTaskNum:    1,
		ReqNPUNum:     1,
		SpBlockNPUNum: 3,
	}
	module2 := chip8node8sp{}
	module2.NPUJob = &util.NPUJob{
		NPUTaskNum:    1,
		ReqNPUNum:     1,
		SpBlockNPUNum: 1,
	}
	module3 := chip8node8sp{}
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

func TestUseAnnotation(t *testing.T) {
	convey.Convey("test UseAnnotation case 1 ", t, func() {
		module1 := &chip8node8sp{}
		task := &api.TaskInfo{
			Job: api.JobID(strconv.Itoa(1)),
		}
		patches := gomonkey.ApplyFunc((*chip8node8sp).selectNPUFromNode,
			func(_ *chip8node8sp, _ *api.TaskInfo, _ plugin.NPUNode) ([]int, error) {
				return []int{1}, nil
			})
		defer patches.Reset()
		patches1 := gomonkey.ApplyFunc((*base.NPUHandler).UpdateNodeInfo,
			func(_ *base.NPUHandler, _ plugin.NPUNode, _ []int) *plugin.NPUNode {
				return &plugin.NPUNode{}
			})
		defer patches1.Reset()
		result := module1.UseAnnotation(task, plugin.NPUNode{})
		convey.So(result, convey.ShouldBeNil)
	})

	convey.Convey("test UseAnnotation case 2 ", t, func() {
		module2 := &chip8node8sp{}
		patches := gomonkey.ApplyFunc((*chip8node8sp).selectNPUFromNode,
			func(_ *chip8node8sp, _ *api.TaskInfo, _ plugin.NPUNode) ([]int, error) {
				return []int{1}, errors.New("fake error")
			})

		defer patches.Reset()
		result := module2.UseAnnotation(&api.TaskInfo{}, plugin.NPUNode{})
		convey.So(result, convey.ShouldBeNil)
	})

}

func TestSelectNPUFromNode(t *testing.T) {
	TestSelectNPUFromNodePart1(t)
	TestSelectNPUFromNodePart2(t)
}

func TestSelectNPUFromNodePart1(t *testing.T) {
	convey.Convey("test selectNPUFromNode case 1 ", t, func() {
		module1 := &chip8node8sp{}
		task := &api.TaskInfo{
			Job: api.JobID(strconv.Itoa(1)),
		}
		patches := gomonkey.ApplyFunc((*base.NPUHandler).GetTaskReqNPUNum,
			func(_ *base.NPUHandler, _ *api.TaskInfo) (int, error) {
				return 0, errors.New("fake error")
			})
		defer patches.Reset()
		_, err := module1.selectNPUFromNode(task, plugin.NPUNode{})
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestSelectNPUFromNodePart2(t *testing.T) {
	convey.Convey("test selectNPUFromNode case 2 ", t, func() {
		module1 := &chip8node8sp{
			spBlock: 1,
		}
		module1.NPUJob = &util.NPUJob{
			NPUTaskNum: 2,
		}
		task := &api.TaskInfo{
			Job: api.JobID(strconv.Itoa(1)),
		}
		patches := gomonkey.ApplyFunc((*base.NPUHandler).GetTaskReqNPUNum,
			func(_ *base.NPUHandler, _ *api.TaskInfo) (int, error) {
				return 1, nil
			})
		defer patches.Reset()
		patches1 := gomonkey.ApplyFunc((*base.NPUHandler).GetUsableTopFromNode,
			func(_ *base.NPUHandler, _ plugin.NPUNode, _ bool) ([]int, error) {
				return []int{1, 2, 3, 4, 5, 6, 7, 8}, nil
			})
		defer patches1.Reset()
		_, err := module1.selectNPUFromNode(task, plugin.NPUNode{})
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test selectNPUFromNode case 3 ", t, func() {
		module1 := &chip8node8sp{
			spBlock: 1,
		}
		module1.NPUJob = &util.NPUJob{
			NPUTaskNum: 2,
		}
		task := &api.TaskInfo{
			Job: api.JobID(strconv.Itoa(1)),
		}
		patches := gomonkey.ApplyFunc((*base.NPUHandler).GetTaskReqNPUNum,
			func(_ *base.NPUHandler, _ *api.TaskInfo) (int, error) {
				return 1, nil
			})
		defer patches.Reset()
		patches1 := gomonkey.ApplyFunc((*base.NPUHandler).GetUsableTopFromNode,
			func(_ *base.NPUHandler, _ plugin.NPUNode, _ bool) ([]int, error) {
				return []int{1, 2, 3, 4, 5, 6}, nil
			})
		defer patches1.Reset()
		_, err := module1.selectNPUFromNode(task, plugin.NPUNode{})
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestInSuperPods(t *testing.T) {
	convey.Convey("test inSuperPods case 1 ", t, func() {
		superPodId := "1"
		node := plugin.NPUNode{}
		convey.So(inSuperPods(nil, superPodId, node), convey.ShouldBeFalse)
	})

	convey.Convey("test inSuperPods case 2 ", t, func() {
		fJob := &rescheduling.FaultJob{
			SuperPods: map[string][]plugin.SuperNode{
				"1": make([]plugin.SuperNode, 0),
			},
		}
		superPodId := "1"
		node := plugin.NPUNode{}
		convey.So(inSuperPods(fJob, superPodId, node), convey.ShouldBeFalse)
	})

	convey.Convey("test inSuperPods case 3 ", t, func() {
		name := "name"
		superNodes := []plugin.SuperNode{
			{
				Name: name,
			},
		}
		fJob := &rescheduling.FaultJob{
			SuperPods: map[string][]plugin.SuperNode{
				"1": superNodes,
			},
		}
		superPodId := "1"
		node := plugin.NPUNode{}
		node.CommonNode = plugin.CommonNode{
			Name: name,
		}
		convey.So(inSuperPods(fJob, superPodId, node), convey.ShouldBeTrue)
	})
}

func TestSelectNodesFromSuperPods(t *testing.T) {
	convey.Convey("test selectNodesFromSuperPods case 1 ", t, func() {
		module := &chip8node8sp{}
		map0 := map[string]plugin.NPUNode{"0": {}}
		map1 := map[string]plugin.NPUNode{
			"1": {},
			"2": {},
		}
		superPods := []superPod{map0, map1}
		unReadyID := []string{"0"}
		totalCount := 1
		selectNodes := make(map[string][]plugin.SuperNode)
		patches1 := gomonkey.ApplyFunc((*chip8node8sp).selectNodesFromSuperPod,
			func(_ *chip8node8sp, _ string, _ map[string]plugin.NPUNode,
				_ map[string][]plugin.SuperNode) map[string]plugin.NPUNode {
				return map[string]plugin.NPUNode{"0": {}}
			})
		defer patches1.Reset()
		result := module.selectNodesFromSuperPods(unReadyID, &totalCount, superPods, selectNodes)
		exceptResult := 2
		convey.So(len(result), convey.ShouldEqual, exceptResult)
	})
	convey.Convey("test selectNodesFromSuperPods case 2 ", t, func() {
		module := &chip8node8sp{
			spBlock: 3,
		}
		map0 := map[string]plugin.NPUNode{"0": {}}
		map1 := map[string]plugin.NPUNode{
			"1": {},
			"2": {},
		}
		superPods := []superPod{map0, map1}
		unReadyID := []string{"0"}
		totalCount := 1
		selectNodes := make(map[string][]plugin.SuperNode)
		patches1 := gomonkey.ApplyFunc((*chip8node8sp).selectNodesFromSuperPod,
			func(_ *chip8node8sp, _ string, _ map[string]plugin.NPUNode,
				_ map[string][]plugin.SuperNode) map[string]plugin.NPUNode {
				return map[string]plugin.NPUNode{"0": {}}
			})
		defer patches1.Reset()
		result := module.selectNodesFromSuperPods(unReadyID, &totalCount, superPods, selectNodes)
		exceptResult := 2
		convey.So(len(result), convey.ShouldEqual, exceptResult)
	})
}

func TestSelectNodesFromSuperPodsExceptReserve(t *testing.T) {
	convey.Convey("test selectNodesFromSuperPods case 1 ", t, func() {
		module := &chip8node8sp{}
		map0 := map[string]plugin.NPUNode{"0": {}}
		map1 := map[string]plugin.NPUNode{
			"1": {},
			"2": {},
		}
		superPods := []superPod{map0, map1}
		unReadyID := []string{"0"}
		totalCount := 1
		selectNodes := make(map[string][]plugin.SuperNode)
		patches1 := gomonkey.ApplyFunc((*chip8node8sp).selectNodesFromSuperPod,
			func(_ *chip8node8sp, _ string, _ map[string]plugin.NPUNode,
				_ map[string][]plugin.SuperNode) map[string]plugin.NPUNode {
				return map[string]plugin.NPUNode{
					"0": {}}
			})
		defer patches1.Reset()
		result := module.selectNodesFromSuperPodsExceptReserve(unReadyID, &totalCount, superPods, selectNodes)
		exceptResult := 2
		convey.So(len(result), convey.ShouldEqual, exceptResult)
	})

	convey.Convey("test selectNodesFromSuperPods case 2 ", t, func() {
		module := &chip8node8sp{
			spBlock: 3,
		}
		map0 := map[string]plugin.NPUNode{"0": {}}
		map1 := map[string]plugin.NPUNode{
			"1": {},
			"2": {},
		}
		superPods := []superPod{map0, map1}
		unReadyID := []string{"0"}
		totalCount := 1
		selectNodes := make(map[string][]plugin.SuperNode)
		patches1 := gomonkey.ApplyFunc((*chip8node8sp).selectNodesFromSuperPod,
			func(_ *chip8node8sp, _ string, _ map[string]plugin.NPUNode,
				_ map[string][]plugin.SuperNode) map[string]plugin.NPUNode {
				return map[string]plugin.NPUNode{
					"0": {}}
			})
		defer patches1.Reset()
		result := module.selectNodesFromSuperPodsExceptReserve(unReadyID, &totalCount, superPods, selectNodes)
		exceptResult := 2
		convey.So(len(result), convey.ShouldEqual, exceptResult)
	})
}

func TestClassifySuperPod(t *testing.T) {
	convey.Convey("test classifySuperPod case 1 ", t, func() {
		module := &chip8node8sp{
			spBlock: 1,
		}
		module.FrameAttr.SuperPodSize = 2
		map0 := map[string]plugin.NPUNode{
			"0": {},
			"3": {},
		}
		map1 := map[string]plugin.NPUNode{
			"1": {},
			"2": {},
		}
		totalNodes := map[int32]superPod{
			0: map0,
			1: map1,
		}
		pod, _ := module.classifySuperPod(totalNodes)
		convey.So(len(pod.firstLevel[0][2]), convey.ShouldEqual, 2)
		convey.So(len(pod.firstLevel[0][2][0]), convey.ShouldEqual, 2)
		convey.So(len(pod.firstLevel[0][2][1]), convey.ShouldEqual, 2)
	})
}

func TestSelectNodesForFaultPod(t *testing.T) {
	convey.Convey("test selectNodesForFaultPod case 1 ", t, func() {
		fJob := &rescheduling.FaultJob{}
		superNodes := []plugin.SuperNode{
			{
				SuperPodID: 1,
			},
		}
		fJob.SuperPods = map[string][]plugin.SuperNode{
			"0": superNodes,
		}
		ids := []int{0}
		totalNodes := map[int32]superPod{0: {"node1": {}}}
		spn := plugin.SuperNode{
			SuperPodID: 0,
		}
		superPodId := "0"
		selectNodesForFaultPod(fJob, ids, totalNodes, spn, superPodId)
		convey.So(len(totalNodes), convey.ShouldEqual, 1)
	})

}

const (
	mockJobUID = "vcjob/job0"
)

func TestIfPodLevelRescheduling(t *testing.T) {
	t.Run("01-IfPodLevelRescheduling return true when pod-rescheduling label is on", func(t *testing.T) {
		module := &chip8node8sp{}
		fJob := &rescheduling.FaultJob{JobUID: mockJobUID}
		sJob := plugin.SchedulerJob{}
		sJob.Label = map[string]string{util.SinglePodTag: util.EnableFunc}
		module.Jobs = map[api.JobID]plugin.SchedulerJob{
			mockJobUID: sJob,
		}
		if res := module.ifPodLevelRescheduling(fJob); !res {
			t.Errorf("ifPodLevelRescheduling() res = %v, wantRes is true", res)
		}
	})
	t.Run("02-IfPodLevelRescheduling return true when process-recover-enable label is on", func(t *testing.T) {
		module := &chip8node8sp{}
		fJob := &rescheduling.FaultJob{JobUID: mockJobUID}
		sJob := plugin.SchedulerJob{}
		sJob.Label = map[string]string{util.ProcessRecoverEnable: util.EnableFunc}
		module.Jobs = map[api.JobID]plugin.SchedulerJob{
			mockJobUID: sJob,
		}
		if res := module.ifPodLevelRescheduling(fJob); !res {
			t.Errorf("ifPodLevelRescheduling() res = %v, wantRes is true", res)
		}
	})
	t.Run("03-IfPodLevelRescheduling return false when both labels are not on", func(t *testing.T) {
		module := &chip8node8sp{}
		fJob := &rescheduling.FaultJob{JobUID: mockJobUID}
		sJob := plugin.SchedulerJob{}
		sJob.Label = map[string]string{}
		module.Jobs = map[api.JobID]plugin.SchedulerJob{
			mockJobUID: sJob,
		}
		if res := module.ifPodLevelRescheduling(fJob); res {
			t.Errorf("ifPodLevelRescheduling() res = %v, wantRes is false", res)
		}
	})
}
