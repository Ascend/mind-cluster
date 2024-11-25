/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package device a series of common function
package device

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
)

const (
	logicID0             = 0
	logicID1             = 1
	logicID2             = 2
	logicID3             = 3
	lengthOfDevFaultInfo = 4
	rankID0              = 0
	lengthOfDevIdList2   = 2
	lengthOfDevIdList0   = 0
	fakePod              = "default/fake-pod"
)

// mockNpuDevice create a fake npu device info
func mockNpuDevice(logicId int32, faultCode []int64) common.NpuDevice {
	return common.NpuDevice{
		FaultCodes: faultCode,
		LogicID:    logicId,
	}
}

// mockNpuDeviceList create a fake npu device info
func mockNpuDeviceList() []*common.NpuDevice {
	npuDevice0 := mockNpuDevice(logicID0, []int64{2350927360})
	npuDevice1 := mockNpuDevice(logicID1, []int64{})
	npuDevice2 := mockNpuDevice(logicID2, []int64{})
	npuDevice3 := mockNpuDevice(logicID3, []int64{})
	return []*common.NpuDevice{
		&npuDevice0,
		&npuDevice1,
		&npuDevice2,
		&npuDevice3,
	}
}

// mockResetErrDevFaultInfo create a fake dev fault info with reset error
func mockResetErrDevFaultInfo(logicId int32) common.DevFaultInfo {
	return common.DevFaultInfo{
		LogicId:       logicId,
		Status:        common.UnrecoveredStatus,
		Policy:        common.ResetError,
		InitialPolicy: common.ResetError,
		ErrorCode:     []int64{2350927360},
		ErrorCodeHex:  "0x8C204E00",
	}
}

// mockEmptyErrDevFaultInfo create a fake dev fault info with empty error
func mockEmptyErrDevFaultInfo(logicId int32) common.DevFaultInfo {
	return common.DevFaultInfo{
		LogicId:       logicId,
		Status:        common.UnrecoveredStatus,
		Policy:        common.EmptyError,
		InitialPolicy: common.EmptyError,
		ErrorCode:     []int64{},
		ErrorCodeHex:  "",
	}
}

// mockAbnormalErrDevFaultInfo create a fake dev fault info with an abnormal error
func mockAbnormalErrDevFaultInfo(logicId int32) common.DevFaultInfo {
	return common.DevFaultInfo{
		LogicId:       logicId,
		Status:        common.UnrecoveredStatus,
		Policy:        "wrong",
		InitialPolicy: "wrong",
		ErrorCode:     []int64{218739174},
		ErrorCodeHex:  "0x88888888",
	}
}

// mockTaskDevInfoList create a fake task dev info list for test
func mockTaskDevInfoList() []*common.TaskDevInfo {
	return []*common.TaskDevInfo{
		{
			RankId:       0,
			DevFaultInfo: mockResetErrDevFaultInfo(0),
		},
		{
			RankId:       1,
			DevFaultInfo: mockEmptyErrDevFaultInfo(1),
		},
	}
}

// mockWrongTaskDevInfoList create a wrong task dev info list for test
func mockWrongTaskDevInfoList() []*common.TaskDevInfo {
	return []*common.TaskDevInfo{
		{
			RankId:       0,
			DevFaultInfo: mockAbnormalErrDevFaultInfo(0),
		},
	}
}

// mockProcessPolicyTable create a fake process policy table for test
func mockProcessPolicyTable() map[string]int {
	return map[string]int{
		common.EmptyError:          common.EmptyErrorLevel,
		common.IgnoreError:         common.IgnoreErrorLevel,
		common.RestartRequestError: common.RestartRequestErrorLevel,
		common.RestartError:        common.RestartErrorLevel,
		common.ResetError:          common.ResetErrorLevel,
		common.IsolateError:        common.IsolateErrorLevel,
	}
}

// newTestHotResetManager new a hot reset manager example
func newTestHotResetManager(deviceType string, model string) HotResetManager {
	common.ParamOption.RealCardType = deviceType
	deviceNum := 16
	return NewHotResetManager(model, deviceNum)
}

// TestGetChipCountOnRing for test the default count of ring ond different device
func TestGetChipCountOnRing(t *testing.T) {
	convey.Convey("test GetChipCountOnRing", t, func() {
		convey.Convey("test 910 chip count on ring success", func() {
			ascend910HotResetManager := newTestHotResetManager(common.Ascend910, common.Train)
			convey.So(ascend910HotResetManager, convey.ShouldNotBeNil)
			chipCountOnRing := ascend910HotResetManager.GetRingNum()
			convey.So(chipCountOnRing, convey.ShouldEqual, common.Ascend910RingsNum)
		})
		convey.Convey("test 910B train chip count on ring success", func() {
			ascend910BTrainHotResetManager := newTestHotResetManager(common.Ascend910B, common.Train)
			convey.So(ascend910BTrainHotResetManager, convey.ShouldNotBeNil)
			chipCountOnRing := ascend910BTrainHotResetManager.GetRingNum()
			convey.So(chipCountOnRing, convey.ShouldEqual, common.Ascend910BRingsNumTrain)
		})
		convey.Convey("test 910B Infer chip count on ring success", func() {
			ascend910BInferHotResetManager := newTestHotResetManager(common.Ascend910B, common.Infer)
			convey.So(ascend910BInferHotResetManager, convey.ShouldNotBeNil)
			chipCountOnRing := ascend910BInferHotResetManager.GetRingNum()
			convey.So(chipCountOnRing, convey.ShouldEqual, common.Ascend910BRingsNumInfer)
		})
		convey.Convey("test 910A3 chip count on ring success", func() {
			ascend910A3HotResetManager := newTestHotResetManager(common.Ascend910A3, common.Train)
			convey.So(ascend910A3HotResetManager, convey.ShouldNotBeNil)
			chipCountOnRing := ascend910A3HotResetManager.GetRingNum()
			convey.So(chipCountOnRing, convey.ShouldEqual, common.Ascend910A3RingsNum)
		})
	})
}

// TestGetAllTaskDevFaultInfoList for test get all the dev fault info list
func TestGetAllTaskDevFaultInfoList(t *testing.T) {
	convey.Convey("test GetTaskAllDevFaultInfoList", t, func() {
		convey.Convey("test GetTaskAllDevFaultInfoList success when not nil", func() {
			tool := &HotResetTools{allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{"test": {}}}
			convey.So(tool.GetAllTaskDevFaultInfoList(), convey.ShouldNotBeNil)
		})
		convey.Convey("test GetTaskAllDevFaultInfoList success when nil", func() {
			tool := &HotResetTools{}
			convey.So(tool.GetAllTaskDevFaultInfoList(), convey.ShouldBeNil)
		})
	})
}

// TestGetTaskDevFaultInfoList for test get the dev fault info list by task name
func TestGetTaskDevFaultInfoList(t *testing.T) {
	convey.Convey("test GetTaskDevFaultInfoList", t, func() {
		convey.Convey("test GetTaskDevFaultInfoList success", func() {
			tool := &HotResetTools{allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{"test": {}}}
			devInfoList, ok := tool.GetTaskDevFaultInfoList("test")
			convey.So(devInfoList, convey.ShouldNotBeNil)
			convey.So(ok, convey.ShouldBeNil)
		})
		convey.Convey("test GetTaskDevFaultInfoList failed", func() {
			tool := &HotResetTools{}
			devInfoList, ok := tool.GetTaskDevFaultInfoList("test")
			convey.So(devInfoList, convey.ShouldBeNil)
			convey.So(ok, convey.ShouldNotBeNil)
		})
	})
}

// TestGetTaskPod for test get the pod of a task by task name
func TestGetTaskPod(t *testing.T) {
	convey.Convey("test GetTaskPod", t, func() {
		convey.Convey("test GetTaskPod success", func() {
			tool := &HotResetTools{taskPod: map[string]v1.Pod{"test": {}}}
			_, err := tool.GetTaskPod("test")
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test GetTaskPod failed", func() {
			tool := &HotResetTools{}
			_, err := tool.GetTaskPod("test")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestGetDevListInReset for test get the device list in reset
func TestGetDevListInReset(t *testing.T) {
	convey.Convey("test GetDevListInReset", t, func() {
		convey.Convey("test GetDevListInReset success when reset dev exist", func() {
			tool := &HotResetTools{resetDev: map[int32]struct{}{0: {}}}
			deviceList := tool.GetDevListInReset()
			convey.So(deviceList, convey.ShouldNotBeNil)
		})
		convey.Convey("test GetTaskDevFaultInfoList success  when reset dev not exist", func() {
			tool := &HotResetTools{}
			deviceList := tool.GetDevListInReset()
			convey.So(deviceList, convey.ShouldBeNil)
		})
	})
}

// TestGetDevProcessPolicy for test get the process policy by fault type
func TestGetDevProcessPolicy(t *testing.T) {
	convey.Convey("test get dev process policy", t, func() {
		tool := &HotResetTools{}
		convey.Convey("test train and infer model GetDevProcessPolicy success", func() {
			normalNPUPolicy := tool.GetDevProcessPolicy(common.NormalNPU)
			notHandleFaultNPUPolicy := tool.GetDevProcessPolicy(common.NotHandleFault)
			convey.So(normalNPUPolicy, convey.ShouldEqual, common.EmptyError)
			convey.So(notHandleFaultNPUPolicy, convey.ShouldEqual, common.EmptyError)

			restartBusinessPolicy := tool.GetDevProcessPolicy(common.RestartBusiness)
			convey.So(restartBusinessPolicy, convey.ShouldEqual, common.RestartError)

			freeRestartNPUPolicy := tool.GetDevProcessPolicy(common.FreeRestartNPU)
			restartNPUPolicy := tool.GetDevProcessPolicy(common.RestartNPU)
			convey.So(freeRestartNPUPolicy, convey.ShouldEqual, common.FreeResetError)
			convey.So(restartNPUPolicy, convey.ShouldEqual, common.ResetError)

			separateNPUPolicy := tool.GetDevProcessPolicy(common.SeparateNPU)
			convey.So(separateNPUPolicy, convey.ShouldEqual, common.IsolateError)
		})
		convey.Convey("test infer model GetDevProcessPolicy success", func() {
			restartRequestPolicy := tool.GetDevProcessPolicy(common.RestartRequest)
			convey.So(restartRequestPolicy, convey.ShouldEqual, common.RestartRequestError)
		})
	})
}

// TestGetTaskProcessPolicy for test get a process policy by task name
func TestGetTaskProcessPolicy(t *testing.T) {
	convey.Convey("test GetTaskProcessPolicy", t, func() {
		convey.Convey("test GetTaskProcessPolicy success", func() {
			tool := &HotResetTools{
				allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{"test": mockTaskDevInfoList()},
				processPolicyTable:  mockProcessPolicyTable(),
			}
			processPolicy, processPolicyLevel, err := tool.GetTaskProcessPolicy("test")
			convey.So(processPolicy, convey.ShouldEqual, common.ResetError)
			convey.So(processPolicyLevel, convey.ShouldEqual, common.ResetErrorLevel)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test GetTaskDevFaultInfoList failed  when task dev info not exist", func() {
			tool := &HotResetTools{
				processPolicyTable: mockProcessPolicyTable(),
			}
			processPolicy, processPolicyLevel, err := tool.GetTaskProcessPolicy("test")
			convey.So(processPolicy, convey.ShouldEqual, "")
			convey.So(processPolicyLevel, convey.ShouldEqual, -1)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("test GetTaskDevFaultInfoList failed when invalid policy", func() {
			tool := &HotResetTools{
				allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{"test": mockWrongTaskDevInfoList()},
				processPolicyTable:  mockProcessPolicyTable(),
			}
			processPolicy, processPolicyLevel, err := tool.GetTaskProcessPolicy("test")
			convey.So(processPolicy, convey.ShouldEqual, "")
			convey.So(processPolicyLevel, convey.ShouldEqual, -1)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestGetDevList for test get the device list
func TestGetDevList(t *testing.T) {
	convey.Convey("test GetDevList", t, func() {
		convey.Convey("test GetDevList success", func() {
			tool := &HotResetTools{}
			devStr := "Ascend910-0,Ascend910-1"
			devIdList := tool.GetDevIdList(devStr)
			convey.So(len(devIdList), convey.ShouldEqual, lengthOfDevIdList2)
		})
		convey.Convey("test GetDevList failed", func() {
			tool := &HotResetTools{}
			devStr := "Ascend910.0,Ascend910.1"
			devIdList := tool.GetDevIdList(devStr)
			convey.So(len(devIdList), convey.ShouldEqual, lengthOfDevIdList0)
		})
	})
}

// TestGetDevListByPolicyLevel for test get the device list by policy level
func TestDevListByPolicyLevel(t *testing.T) {
	convey.Convey("test GetDevListByPolicyLevel", t, func() {
		convey.Convey("test GetDevListByPolicyLevel success", func() {
			tool := &HotResetTools{
				allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{"test": mockTaskDevInfoList()},
				processPolicyTable:  mockProcessPolicyTable(),
			}
			devList, err := tool.GetDevListByPolicyLevel(tool.allTaskDevFaultInfo["test"], common.ResetErrorLevel)
			convey.So(devList[0], convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
			devList2, err := tool.GetDevListByPolicyLevel(tool.allTaskDevFaultInfo["test"], common.IsolateErrorLevel)
			convey.So(len(devList2), convey.ShouldEqual, 0)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test GetDevListByPolicyLevel failed", func() {
			tool := &HotResetTools{
				allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{"test": mockWrongTaskDevInfoList()},
				processPolicyTable:  mockProcessPolicyTable(),
			}
			devList, err := tool.GetDevListByPolicyLevel(tool.allTaskDevFaultInfo["test"], common.ResetErrorLevel)
			convey.So(devList, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestGetNeedResetDevList for test get the needed be reseted device list
func TestGetNeedResetDevList(t *testing.T) {
	convey.Convey("test GetNeedResetDevMap", t, func() {
		convey.Convey("test GetNeedResetDevMap success", func() {
			tool := &HotResetTools{
				allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{"test": mockTaskDevInfoList()},
				processPolicyTable:  mockProcessPolicyTable(),
			}
			devFaultInfoList, ok := tool.allTaskDevFaultInfo["test"]
			convey.So(ok, convey.ShouldBeTrue)
			devList, err := tool.GetNeedResetDevMap(devFaultInfoList)
			convey.So(err, convey.ShouldBeNil)
			needResetDev, ok := devList[0]
			convey.So(needResetDev, convey.ShouldNotBeNil)
			convey.So(ok, convey.ShouldBeTrue)
			_, ok = devList[1]
			convey.So(ok, convey.ShouldBeFalse)
		})
		convey.Convey("test GetNeedResetDevMap failed", func() {
			tool := &HotResetTools{
				allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{"test": mockWrongTaskDevInfoList()},
				processPolicyTable:  mockProcessPolicyTable(),
			}
			devFaultInfoList, ok := tool.allTaskDevFaultInfo["test"]
			convey.So(ok, convey.ShouldBeTrue)
			devList, err := tool.GetNeedResetDevMap(devFaultInfoList)
			convey.So(devList, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestGetTaskResetInfo for test get the reset info of task to process
func TestGetTaskResetInfo(t *testing.T) {
	convey.Convey("test GetTaskResetInfo", t, func() {
		convey.Convey("test GetTaskResetInfo success", func() {
			tool := &HotResetTools{
				ringNum:             common.Ascend910BRingsNumTrain,
				allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{"test": mockTaskDevInfoList()},
				processPolicyTable:  mockProcessPolicyTable(),
			}
			devFaultInfoList, ok := tool.allTaskDevFaultInfo["test"]
			convey.So(ok, convey.ShouldBeTrue)
			taskResetInfo, err := tool.GetTaskResetInfo(devFaultInfoList, common.ResetError,
				common.ResetError, common.UnrecoveredStatus)
			convey.So(err, convey.ShouldBeNil)
			convey.So(taskResetInfo.RankList[0].RankId, convey.ShouldEqual, 0)
			convey.So(taskResetInfo.RankList[0].Status, convey.ShouldEqual, common.UnrecoveredStatus)
			convey.So(taskResetInfo.RankList[0].Policy, convey.ShouldEqual, common.ResetError)
			convey.So(taskResetInfo.RankList[0].InitialPolicy, convey.ShouldEqual, common.ResetError)
		})
		convey.Convey("test GetTaskResetInfo failed", func() {
			tool := &HotResetTools{
				ringNum:             common.Ascend910BRingsNumTrain,
				allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{"test": mockWrongTaskDevInfoList()},
				processPolicyTable:  mockProcessPolicyTable(),
			}
			devFaultInfoList, ok := tool.allTaskDevFaultInfo["test"]
			convey.So(ok, convey.ShouldBeTrue)
			taskResetInfo, err := tool.GetTaskResetInfo(devFaultInfoList, common.ResetError,
				common.ResetError, common.UnrecoveredStatus)
			convey.So(taskResetInfo, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestGetTaskFaultRankInfo for test get the fault rank info of task
func TestGetTaskFaultRankInfo(t *testing.T) {
	convey.Convey("test GetTaskFaultRankInfo", t, func() {
		convey.Convey("test GetTaskFaultRankInfo success", func() {
			tool := &HotResetTools{
				ringNum:             common.Ascend910BRingsNumTrain,
				allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{"test": mockTaskDevInfoList()},
				processPolicyTable:  mockProcessPolicyTable(),
			}
			devFaultInfoList, ok := tool.allTaskDevFaultInfo["test"]
			convey.So(ok, convey.ShouldBeTrue)
			faultRankInfo, err := tool.GetTaskFaultRankInfo(devFaultInfoList)
			convey.So(err, convey.ShouldBeNil)
			sliceIntEqual(faultRankInfo.FaultRank, []int{0, 1})
		})
		convey.Convey("test GetTaskFaultRankInfo failed", func() {
			tool := &HotResetTools{
				ringNum:             common.Ascend910BRingsNumTrain,
				allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{"test": mockWrongTaskDevInfoList()},
				processPolicyTable:  mockProcessPolicyTable(),
			}
			devFaultInfoList, ok := tool.allTaskDevFaultInfo["test"]
			convey.So(ok, convey.ShouldBeTrue)
			faultRankInfo, err := tool.GetTaskFaultRankInfo(devFaultInfoList)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(faultRankInfo.FaultRank), convey.ShouldEqual, 0)
		})
	})
}

// TestGetFaultDev2PodMap for test get the fault dev with pod map
func TestGetFaultDev2PodMap(t *testing.T) {
	convey.Convey("test GetFaultDev2PodMap", t, func() {
		convey.Convey("test GetFaultDev2PodMap success", func() {
			tool := &HotResetTools{
				faultDev2PodMap: map[int32]v1.Pod{int32(0): {}},
			}
			devPodMap, err := tool.GetFaultDev2PodMap()
			convey.So(err, convey.ShouldBeNil)
			convey.So(devPodMap, convey.ShouldNotBeNil)
		})
		convey.Convey("test GetFaultDev2PodMap failed", func() {
			tool := &HotResetTools{}
			devPodMap, err := tool.GetFaultDev2PodMap()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(devPodMap, convey.ShouldBeNil)
		})
	})
}

// TestGenerateTaskDevFaultInfoList for test generate the dev fault info list of task
func TestGenerateTaskDevFaultInfoList(t *testing.T) {
	convey.Convey("test GenerateTaskDevFaultInfoList", t, func() {
		convey.Convey("test GenerateTaskDevFaultInfoList success", func() {
			resetErrDevFaultInfo := mockResetErrDevFaultInfo(logicID0)
			emptyErrDevFaultInfo1 := mockEmptyErrDevFaultInfo(logicID1)
			emptyErrDevFaultInfo2 := mockEmptyErrDevFaultInfo(logicID2)
			emptyErrDevFaultInfo3 := mockEmptyErrDevFaultInfo(logicID3)
			tool := &HotResetTools{
				globalDevFaultInfo: map[int32]*common.DevFaultInfo{
					0: &resetErrDevFaultInfo,
					1: &emptyErrDevFaultInfo1,
					2: &emptyErrDevFaultInfo2,
					3: &emptyErrDevFaultInfo3,
				},
			}
			devIDList := []int32{0, 1, 2, 3}
			taskDevInfo, err := tool.GenerateTaskDevFaultInfoList(devIDList, "0")
			convey.So(err, convey.ShouldBeNil)
			convey.So(taskDevInfo, convey.ShouldNotBeNil)
			convey.So(len(taskDevInfo), convey.ShouldEqual, lengthOfDevFaultInfo)
			convey.So(taskDevInfo[0].RankId, convey.ShouldEqual, rankID0)
			convey.So(taskDevInfo[0].Status, convey.ShouldEqual, common.UnrecoveredStatus)
			convey.So(taskDevInfo[0].Policy, convey.ShouldEqual, common.ResetError)
			convey.So(taskDevInfo[0].InitialPolicy, convey.ShouldEqual, common.ResetError)
		})
	})
}

// TestUpdateFaultDev2PodMap for test update the fault dev pod map
func TestUpdateFaultDev2PodMap(t *testing.T) {
	convey.Convey("test UpdateFaultDev2PodMap", t, func() {
		convey.Convey("test UpdateFaultDev2PodMap success", func() {
			// mock device 0 unhealthy
			resetErrDevFaultInfo := mockResetErrDevFaultInfo(logicID0)
			emptyErrDevFaultInfo1 := mockEmptyErrDevFaultInfo(logicID1)
			emptyErrDevFaultInfo2 := mockEmptyErrDevFaultInfo(logicID2)
			emptyErrDevFaultInfo3 := mockEmptyErrDevFaultInfo(logicID3)
			tool := &HotResetTools{
				faultDev2PodMap: map[int32]v1.Pod{},
				globalDevFaultInfo: map[int32]*common.DevFaultInfo{
					0: &resetErrDevFaultInfo,
					1: &emptyErrDevFaultInfo1,
					2: &emptyErrDevFaultInfo2,
					3: &emptyErrDevFaultInfo3,
				},
			}
			devIDList := []int32{0, 1, 2, 3}
			err := tool.UpdateFaultDev2PodMap(devIDList, v1.Pod{})
			convey.So(err, convey.ShouldBeNil)
			_, ok := tool.faultDev2PodMap[0]
			convey.So(ok, convey.ShouldBeTrue)
			emptyErrDevFaultInfo0 := mockEmptyErrDevFaultInfo(0)
			// mock device 0 healthy
			tool.globalDevFaultInfo[0] = &emptyErrDevFaultInfo0
			err = tool.UpdateFaultDev2PodMap(devIDList, v1.Pod{})
			convey.So(err, convey.ShouldBeNil)
			_, ok = tool.faultDev2PodMap[0]
			convey.So(ok, convey.ShouldBeFalse)
		})
	})
}

// TestUpdateGlobalDevFaultInfoCache for test update the global fault info in cache
func TestUpdateGlobalDevFaultInfoCache(t *testing.T) {
	convey.Convey("test UpdateGlobalDevFaultInfoCache", t, func() {
		convey.Convey("test UpdateGlobalDevFaultInfoCache success", func() {
			deviceList := mockNpuDeviceList()
			var empty []int32
			tool := &HotResetTools{
				globalDevFaultInfo: map[int32]*common.DevFaultInfo{},
			}
			err := tool.UpdateGlobalDevFaultInfoCache(deviceList, empty)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(tool.globalDevFaultInfo), convey.ShouldEqual, lengthOfDevFaultInfo)
			sliceInt64Equal(tool.globalDevFaultInfo[0].ErrorCode, []int64{2350927360})
		})
	})
}

// TestUpdateTaskDevListCache for test update the task dev list
func TestUpdateTaskDevListCache(t *testing.T) {
	convey.Convey("test UpdateTaskDevListCache", t, func() {
		convey.Convey("test UpdateTaskDevListCache success", func() {
			tool := &HotResetTools{}
			convey.So(tool.allTaskDevList, convey.ShouldBeNil)
			taskDevList := map[string][]int32{"test": {0}}
			err := tool.UpdateTaskDevListCache(taskDevList)
			convey.So(err, convey.ShouldBeNil)
			convey.So(tool.allTaskDevList, convey.ShouldNotBeNil)
		})
		convey.Convey("test UpdateTaskDevListCache failed", func() {
			tool := &HotResetTools{}
			convey.So(tool.allTaskDevList, convey.ShouldBeNil)
			var taskDevList map[string][]int32
			err := tool.UpdateTaskDevListCache(taskDevList)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestUpdateTaskDevFaultInfoCache for test update the task fault info cache
func TestUpdateTaskDevFaultInfoCache(t *testing.T) {
	convey.Convey("test UpdateTaskDevFaultInfoCache", t, func() {
		convey.Convey("test UpdateTaskDevFaultInfoCache success", func() {
			tool := &HotResetTools{}
			convey.So(tool.allTaskDevList, convey.ShouldBeNil)
			taskDevList := map[string][]int32{"test": {0}}
			err := tool.UpdateTaskDevListCache(taskDevList)
			convey.So(err, convey.ShouldBeNil)
			convey.So(tool.allTaskDevList, convey.ShouldNotBeNil)
		})
		convey.Convey("test UpdateTaskDevFaultInfoCache failed", func() {
			tool := &HotResetTools{}
			convey.So(tool.allTaskDevList, convey.ShouldBeNil)
			var taskDevList map[string][]int32
			err := tool.UpdateTaskDevListCache(taskDevList)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestUpdateTaskPodCache for test update the task pod cache
func TestUpdateTaskPodCache(t *testing.T) {
	convey.Convey("test UpdateTaskPodCache", t, func() {
		convey.Convey("test UpdateTaskPodCache success", func() {
			tool := &HotResetTools{}
			convey.So(tool.taskPod, convey.ShouldBeNil)
			taskPod := map[string]v1.Pod{"test": {}}
			err := tool.UpdateTaskPodCache(taskPod)
			convey.So(err, convey.ShouldBeNil)
			convey.So(tool.taskPod, convey.ShouldNotBeNil)
		})
		convey.Convey("test UpdateTaskPodCache failed", func() {
			tool := &HotResetTools{}
			convey.So(tool.taskPod, convey.ShouldBeNil)
			var taskPod map[string]v1.Pod
			err := tool.UpdateTaskPodCache(taskPod)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestUpdateFreeTask for test delete the free task in cache
func TestUpdateFreeTask(t *testing.T) {
	convey.Convey("test UpdateFreeTask", t, func() {
		convey.Convey("test UpdateFreeTask success", func() {
			tool := &HotResetTools{
				resetTask: map[string]struct{}{"test": {}},
			}
			_, ok := tool.resetTask["test"]
			convey.So(ok, convey.ShouldBeTrue)
			taskListUseDevice := map[string]struct{}{}
			newTaskDevList := map[string][]int32{}
			tool.UpdateFreeTask(taskListUseDevice, newTaskDevList)
			_, ok = tool.resetTask["test"]
			convey.So(ok, convey.ShouldBeFalse)
		})
	})
}

// TestIsCurNodeTaskInReset for test judge whether the current node task is being resetting
func TestIsCurNodeTaskInReset(t *testing.T) {
	convey.Convey("test IsCurNodeTaskInReset", t, func() {
		convey.Convey("test IsCurNodeTaskInReset true", func() {
			tool := &HotResetTools{
				resetTask: map[string]struct{}{"test": {}},
			}
			convey.So(tool.IsCurNodeTaskInReset("test"), convey.ShouldBeTrue)
		})
		convey.Convey("test IsCurNodeTaskInReset false", func() {
			tool := &HotResetTools{
				resetTask: map[string]struct{}{},
			}
			convey.So(tool.IsCurNodeTaskInReset("test"), convey.ShouldBeFalse)
		})
	})
}

// TestIsExistFaultyDevInTask for test judge whether the faulty dev exist in task
func TestIsExistFaultyDevInTask(t *testing.T) {
	convey.Convey("test IsExistFaultyDevInTask", t, func() {
		convey.Convey("test IsExistFaultyDevInTask true", func() {
			tool := &HotResetTools{
				allTaskDevList: map[string][]int32{"test": {}},
				resetTask:      map[string]struct{}{"test": {}},
				faultDev2PodMap: map[int32]v1.Pod{0: {
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{common.ResetTaskNameKey: "test"},
						Labels:      map[string]string{common.ResetTaskNameKeyInLabel: "test"},
					},
				},
				},
			}
			convey.So(tool.IsExistFaultyDevInTask("test"), convey.ShouldBeTrue)
		})
		convey.Convey("test IsExistFaultyDevInTask false by not in cache", func() {
			tool := &HotResetTools{}
			convey.So(tool.IsExistFaultyDevInTask("test"), convey.ShouldBeFalse)
		})
		convey.Convey("test IsExistFaultyDevInTask false by not have annotation and label", func() {
			tool := &HotResetTools{
				allTaskDevList: map[string][]int32{"test": {}},
				resetTask:      map[string]struct{}{"test": {}},
				faultDev2PodMap: map[int32]v1.Pod{0: {
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{},
						Labels:      map[string]string{},
					},
				},
				},
			}
			// test the pod have not reset annotation
			convey.So(tool.IsExistFaultyDevInTask("test"), convey.ShouldBeFalse)
			// test the pod have not reset label
			tool.faultDev2PodMap[0].Annotations[common.ResetTaskNameKey] = "test"
			convey.So(tool.IsExistFaultyDevInTask("test"), convey.ShouldBeTrue)
			// test the pod have not reset annotation
			delete(tool.faultDev2PodMap[0].Annotations, common.ResetTaskNameKey)
			tool.faultDev2PodMap[0].Labels[common.ResetTaskNameKeyInLabel] = "test"
			convey.So(tool.IsExistFaultyDevInTask("test"), convey.ShouldBeTrue)
		})
	})
}

// TestSetTaskInReset for test set task in reset task cache
func TestSetTaskInReset(t *testing.T) {
	convey.Convey("test SetTaskInReset", t, func() {
		convey.Convey("test SetTaskInReset success", func() {
			tool := &HotResetTools{
				resetTask: map[string]struct{}{},
			}
			err := tool.SetTaskInReset("test")
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test SetTaskInReset failed", func() {
			tool := &HotResetTools{
				resetTask: map[string]struct{}{"test": {}},
			}
			err := tool.SetTaskInReset("test")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestSetDevInReset for test set dev in reset dev cache
func TestSetDevInReset(t *testing.T) {
	convey.Convey("test SetDevInReset", t, func() {
		convey.Convey("test SetDevInReset success", func() {
			tool := &HotResetTools{
				resetDev: map[int32]struct{}{},
			}
			err := tool.SetDevInReset(0)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test SetDevInReset failed", func() {
			tool := &HotResetTools{
				resetDev: map[int32]struct{}{0: {}},
			}
			err := tool.SetDevInReset(0)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestSetAllDevInReset for test set all dev in reset dev cache
func TestSetAllDevInReset(t *testing.T) {
	convey.Convey("test SetAllDevInReset", t, func() {
		convey.Convey("test SetAllDevInReset success", func() {
			tool := &HotResetTools{
				resetDev: map[int32]struct{}{},
			}
			resetInfo := &common.TaskResetInfo{
				RankList: mockTaskDevInfoList(),
			}
			err := tool.SetAllDevInReset(resetInfo)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test SetAllDevInReset failed", func() {
			tool := &HotResetTools{
				resetDev: map[int32]struct{}{0: {}},
			}
			resetInfo := &common.TaskResetInfo{
				RankList: mockTaskDevInfoList(),
			}
			err := tool.SetAllDevInReset(resetInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestUnSetDevInReset for test unset dev in reset dev cache
func TestUnSetDevInReset(t *testing.T) {
	convey.Convey("test UnSetDevInReset", t, func() {
		convey.Convey("test UnSetDevInReset success", func() {
			tool := &HotResetTools{
				resetDev: map[int32]struct{}{0: {}},
			}
			err := tool.UnSetDevInReset(0)
			convey.So(err, convey.ShouldBeNil)
			_, ok := tool.resetDev[0]
			convey.So(ok, convey.ShouldBeFalse)
		})
		convey.Convey("test UnSetDevInReset failed", func() {
			tool := &HotResetTools{
				resetDev: map[int32]struct{}{},
			}
			err := tool.UnSetDevInReset(0)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestUnSetAllDevInReset for test unset dev in reset dev cache
func TestUnSetAllDevInReset(t *testing.T) {
	convey.Convey("test UnSetAllDevInReset", t, func() {
		convey.Convey("test UnSetAllDevInReset success", func() {
			tool := &HotResetTools{
				resetDev: map[int32]struct{}{0: {}, 1: {}},
			}
			resetInfo := &common.TaskResetInfo{
				RankList: mockTaskDevInfoList(),
			}
			err := tool.UnSetAllDevInReset(resetInfo)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(tool.resetDev), convey.ShouldEqual, 0)
		})
		convey.Convey("test UnSetAllDevInReset failed", func() {
			tool := &HotResetTools{
				resetDev: map[int32]struct{}{},
			}
			resetInfo := &common.TaskResetInfo{
				RankList: mockTaskDevInfoList(),
			}
			err := tool.UnSetAllDevInReset(resetInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestUnSetTaskInReset for test unset task in reset task cache
func TestUnSetTaskInReset(t *testing.T) {
	convey.Convey("test UnSetTaskInReset", t, func() {
		convey.Convey("test UnSetTaskInReset success", func() {
			tool := &HotResetTools{
				resetTask: map[string]struct{}{"test": {}},
			}
			err := tool.UnSetTaskInReset("test")
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(tool.resetDev), convey.ShouldEqual, 0)
		})
		convey.Convey("test UnSetTaskInReset failed", func() {
			tool := &HotResetTools{
				resetTask: map[string]struct{}{},
			}
			err := tool.UnSetTaskInReset("test")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestDeepCopyFunc for test function of deep copy
func TestDeepCopyFunc(t *testing.T) {
	convey.Convey("test deep copy func of tool", t, func() {
		devInfoList := mockTaskDevInfoList()
		tool := &HotResetTools{}
		convey.Convey("test deep copy task dev info struct true", func() {
			devInfo := devInfoList[0]
			devInfoTest := tool.DeepCopyDevInfo(devInfo)
			deepTestDevInfo(devInfo, devInfoTest)
		})
		convey.Convey("test deep copy task dev info struct list true", func() {
			devInfoListTest := tool.DeepCopyDevFaultInfoList(devInfoList)
			convey.So(devInfoListTest, convey.ShouldNotEqual, devInfoList)
			for i := range devInfoList {
				deepTestDevInfo(devInfoList[i], devInfoListTest[i])
			}
		})
	})
}

func deepTestDevInfo(devInfo, devInfoTest *common.TaskDevInfo) {
	convey.So(devInfoTest, convey.ShouldNotBeNil)
	convey.So(devInfo, convey.ShouldNotBeNil)
	convey.So(devInfoTest, convey.ShouldNotEqual, devInfo)
	convey.So(devInfoTest.RankId, convey.ShouldEqual, devInfo.RankId)
	convey.So(devInfoTest.DevFaultInfo, convey.ShouldNotEqual, devInfo.DevFaultInfo)
	convey.So(devInfoTest.DevFaultInfo.LogicId, convey.ShouldEqual, devInfo.DevFaultInfo.LogicId)
	convey.So(devInfoTest.DevFaultInfo.Policy, convey.ShouldEqual, devInfo.DevFaultInfo.Policy)
	convey.So(devInfoTest.DevFaultInfo.Status, convey.ShouldEqual, devInfo.DevFaultInfo.Status)
	convey.So(devInfoTest.DevFaultInfo.InitialPolicy, convey.ShouldEqual, devInfo.DevFaultInfo.InitialPolicy)
	sliceInt64Equal(devInfoTest.DevFaultInfo.ErrorCode, devInfo.DevFaultInfo.ErrorCode)
	convey.So(devInfoTest.DevFaultInfo.ErrorCodeHex, convey.ShouldEqual, devInfo.DevFaultInfo.ErrorCodeHex)
}

func sliceInt64Equal(slice1, slice2 []int64) {
	convey.So(len(slice1), convey.ShouldEqual, len(slice2))
	if len(slice1) != len(slice2) {
		return
	}
	for i := range slice1 {
		convey.So(slice1[i], convey.ShouldEqual, slice2[i])
	}
}

func sliceIntEqual(slice1, slice2 []int) {
	convey.So(len(slice1), convey.ShouldEqual, len(slice2))
	if len(slice1) != len(slice2) {
		return
	}
	for i := range slice1 {
		convey.So(slice1[i], convey.ShouldEqual, slice2[i])
	}
}

// TestCheckConfigMap test check config map
func TestCheckConfigMap(t *testing.T) {
	convey.Convey("test checkConfigMap", t, func() {
		convey.Convey("not cm obj will return false", func() {
			cm := "fake-cm"
			res := checkConfigMap(cm)
			convey.ShouldEqual(res, false)
		})
		convey.Convey("cm's name without request prefix return false", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-name1",
					Namespace: "fake-namespace",
				},
			}
			res := checkConfigMap(cm)
			convey.ShouldEqual(res, false)
		})
		convey.Convey("cm's name with request prefix return false", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      common.ResetInfoCMNamePrefix + "fake-name2",
					Namespace: "fake-namespace",
				},
			}
			res := checkConfigMap(cm)
			convey.ShouldEqual(res, true)
		})
	})
}

// TestHandlePodAddEventJobNameFailed test handle pod add event
func TestHandlePodAddEventJobNameFailed(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	event := kubeclient.Event{
		Resource: kubeclient.PodResource,
		Key:      fakePod,
		Type:     kubeclient.EventTypeAdd,
	}
	convey.Convey("test TestHandlePodAddEventJobNameFailed", t, func() {
		convey.Convey("will do nothing when get job name failed", func() {
			patch := gomonkey.ApplyPrivateMethod(new(HotResetTools), "getPodFromCache", func(_ *HotResetTools,
				_ string) (*v1.Pod, error) {
				return &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{},
					},
				}, nil
			})
			defer patch.Reset()
			ascend910HotResetManager.handlePodAddEvent(event)
			_, ok := ascend910HotResetManager.jobs[event.Key]
			convey.ShouldEqual(ok, false)
		})
		convey.Convey("will do nothing when get cm failed", func() {
			patch := gomonkey.ApplyPrivateMethod(new(HotResetTools), "getPodFromCache", func(_ *HotResetTools,
				_ string) (*v1.Pod, error) {
				return &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "default2",
						Labels:    map[string]string{common.ResetTaskNameKey: "test-job2"},
					},
				}, nil
			})
			defer patch.Reset()
			patch2 := gomonkey.ApplyMethod(new(HotResetTools), "GetCMFromCache", func(_ *HotResetTools,
				_ string) (*v1.ConfigMap, error) {
				return nil, errors.New("cm not found")
			})
			defer patch2.Reset()
			ascend910HotResetManager.handlePodAddEvent(event)
		})
		convey.Convey("will do nothing when pod has not been cached", func() {
			patch := gomonkey.ApplyPrivateMethod(new(HotResetTools), "getPodFromCache", func(_ *HotResetTools,
				_ string) (*v1.Pod, error) {
				return nil, errors.New("pod not found")
			})
			defer patch.Reset()
			ascend910HotResetManager.handlePodAddEvent(event)
		})
	})
}

// TestHandlePodAddEventJobNameSucceed test handle pod add event
func TestHandlePodAddEventJobNameSucceed(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	event := kubeclient.Event{
		Resource: kubeclient.PodResource,
		Key:      fakePod,
		Type:     kubeclient.EventTypeAdd,
	}
	convey.Convey("test TestHandlePodAddEventJobNameSucceed", t, func() {
		convey.Convey("will cache job when get job name success", func() {
			patch := gomonkey.ApplyPrivateMethod(new(HotResetTools), "getPodFromCache", func(_ *HotResetTools,
				_ string) (*v1.Pod, error) {
				return &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
						Labels:    map[string]string{common.ResetTaskNameKey: "test-job"},
					},
				}, nil
			})
			defer patch.Reset()
			patch2 := gomonkey.ApplyMethod(new(HotResetTools), "GetCMFromCache", func(_ *HotResetTools,
				_ string) (*v1.ConfigMap, error) {
				return nil, errors.New("cm not found")
			})
			defer patch2.Reset()
			ascend910HotResetManager.handlePodAddEvent(event)
			jobName, ok := ascend910HotResetManager.jobs[event.Key]
			convey.ShouldEqual(ok, true)
			convey.ShouldEqual(jobName, "test-job")
		})
	})
}

// TestHandlePodAddEventGetCMSucceed test handle pod add event
func TestHandlePodAddEventGetCMSucceed(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	event := kubeclient.Event{
		Resource: kubeclient.PodResource,
		Key:      fakePod,
		Type:     kubeclient.EventTypeAdd,
	}
	convey.Convey("test TestHandlePodAddEventGetCMSucceed", t, func() {
		convey.Convey("will write to file when get cm success", func() {
			patch := gomonkey.ApplyPrivateMethod(new(HotResetTools), "getPodFromCache", func(_ *HotResetTools,
				_ string) (*v1.Pod, error) {
				return &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: "default3",
						Labels:    map[string]string{common.ResetTaskNameKey: "test-job3"},
					},
				}, nil
			})
			defer patch.Reset()
			patch2 := gomonkey.ApplyMethod(new(HotResetTools), "GetCMFromCache", func(_ *HotResetTools,
				_ string) (*v1.ConfigMap, error) {
				return &v1.ConfigMap{
					Data: map[string]string{},
				}, nil
			})
			defer patch2.Reset()
			patch3 := gomonkey.ApplyPrivateMethod(new(HotResetTools), "writeCMToFile", func(_ *HotResetTools,
				_ *v1.ConfigMap) error {
				return nil
			})
			defer patch3.Reset()
			ascend910HotResetManager.handlePodAddEvent(event)
		})
	})
}

// TestHandlePodDeleteEvent test handle pod delete event
func TestHandlePodDeleteEvent(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	event := kubeclient.Event{
		Resource: kubeclient.PodResource,
		Key:      fakePod,
		Type:     kubeclient.EventTypeDelete,
	}
	convey.Convey("test HandlePodDeleteEvent", t, func() {
		convey.Convey("will do nothing when jobs has not been cached", func() {
			ascend910HotResetManager.handlePodDeleteEvent(event)
		})
		convey.Convey("cached job will be deleted", func() {
			ascend910HotResetManager.jobs[event.Key] = "fake-job"
			ascend910HotResetManager.handlePodDeleteEvent(event)
			_, ok := ascend910HotResetManager.jobs[event.Key]
			convey.ShouldEqual(ok, false)
		})
	})
}

// TestGetPodFromCache test get cm from cache
func TestGetPodFromCache(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test GetCMFromCache", t, func() {
		convey.Convey("get pod from cache failed when pod is not exist", func() {
			cm, err := ascend910HotResetManager.getPodFromCache("fake-name3")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(cm, convey.ShouldBeNil)
		})
		convey.Convey("get pod from cache sucess when item is pod", func() {
			ascend910HotResetManager.podIndexer.Add(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
			})
			cm, err := ascend910HotResetManager.getPodFromCache("default/test-pod")
			convey.So(err, convey.ShouldBeNil)
			convey.So(cm, convey.ShouldNotBeNil)
		})
	})
}

// TestGetCMFromCache test get cm from cache
func TestGetCMFromCache(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test GetCMFromCache", t, func() {
		convey.Convey("get cm from cache failed when cm is not exist", func() {
			cm, err := ascend910HotResetManager.GetCMFromCache("fake-name4")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(cm, convey.ShouldBeNil)
		})
		convey.Convey("get cm from cache sucess when item is cm", func() {
			ascend910HotResetManager.cmIndexer.Add(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: "default",
				},
			})
			cm, err := ascend910HotResetManager.GetCMFromCache("default/test-cm")
			convey.So(err, convey.ShouldBeNil)
			convey.So(cm, convey.ShouldNotBeNil)
		})
	})
}

// TestWriteCMToFile test write cm to file
func TestWriteCMToFile(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test writeCMToFile", t, func() {
		convey.Convey("write cm to file failed when cm has not reset.json", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "reset-config-test1",
				},
				Data: map[string]string{"xxx": "yyy"},
			}
			err := ascend910HotResetManager.writeCMToFile(cm)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("write cm to file failed when dir is not exist", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "reset-config-test2",
				},
				Data: map[string]string{common.ResetInfoCMDataKey: "yyy"},
			}
			err := ascend910HotResetManager.writeCMToFile(cm)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("write cm to file success when dir is exist", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "reset-config-test2",
				},
				Data: map[string]string{common.ResetInfoCMDataKey: "yyy"},
			}
			err := os.MkdirAll(common.ResetInfoDir, os.ModePerm)
			if err != nil {
				hwlog.RunLog.Error("mkdir command failed")
			}
			err = ascend910HotResetManager.writeCMToFile(cm)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestHandleCMAddEvent test of handleCMAddEvent
func TestHandleCMAddEvent(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test handleCMAddEvent", t, func() {
		convey.Convey("cm not found will do nothing", func() {
			mokeEvent := kubeclient.Event{
				Resource: kubeclient.CMResource,
				Key:      "default/reset-config-test",
				Type:     kubeclient.EventTypeAdd,
			}
			ascend910HotResetManager.queue.Add(mokeEvent)
			patch := gomonkey.ApplyMethod(new(HotResetTools), "GetCMFromCache", func(_ *HotResetTools,
				_ string) (*v1.ConfigMap, error) {
				return nil, errors.New("not found")
			})
			defer patch.Reset()
			ascend910HotResetManager.handleCMUpdateEvent(mokeEvent)
		})
		convey.Convey("cm obj will return false", func() {
			mokeEvent := kubeclient.Event{
				Resource: kubeclient.CMResource,
				Key:      "default/reset-config-test",
				Type:     kubeclient.EventTypeAdd,
			}
			ascend910HotResetManager.queue.Add(mokeEvent)
			patch1 := gomonkey.ApplyMethod(new(HotResetTools), "GetCMFromCache", func(_ *HotResetTools,
				_ string) (*v1.ConfigMap, error) {
				return &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "default",
						Namespace: "reset-config-test",
					},
					Data: map[string]string{common.ResetInfoCMDataKey: "YYY"},
				}, nil
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyPrivateMethod(new(HotResetTools), "writeCMToFile", func(_ *HotResetTools,
				_ *v1.ConfigMap) error {
				return nil
			})
			defer patch2.Reset()
			ascend910HotResetManager.handleCMUpdateEvent(mokeEvent)
		})
	})
}

// TestHandleCMDeleteEvent
func TestHandleCMDeleteEvent(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test handleCMDeleteEvent", t, func() {
		convey.Convey("test handle delete event success", func() {
			mokeEvent := kubeclient.Event{
				Resource: "",
				Key:      "fake/event",
				Type:     "",
			}
			ascend910HotResetManager.queue.Add(mokeEvent)
			ascend910HotResetManager.handleCMDeleteEvent(mokeEvent)
		})
	})
}

func newHotResetTools() *HotResetTools {
	return &HotResetTools{
		ringNum:          common.Ascend910RingsNum,
		resetTask:        map[string]struct{}{},
		resetDev:         map[int32]struct{}{},
		faultDev2PodMap:  map[int32]v1.Pod{},
		jobs:             map[string]string{},
		noResetCmPodKeys: map[string]string{},
		processPolicyTable: map[string]int{
			common.EmptyError:          common.EmptyErrorLevel,
			common.IgnoreError:         common.IgnoreErrorLevel,
			common.RestartRequestError: common.RestartRequestErrorLevel,
			common.RestartError:        common.RestartErrorLevel,
			common.FreeResetError:      common.FreeResetErrorLevel,
			common.ResetError:          common.ResetErrorLevel,
			common.IsolateError:        common.IsolateErrorLevel,
		},
		queue:      workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		cmIndexer:  cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{}),
		podIndexer: cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{}),
	}
}
