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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
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

// newTestHotResetManager new a hot reset manager example
func newTestHotResetManager(deviceType string, model string, deviceNum int) HotResetManager {
	common.ParamOption.RealCardType = deviceType
	return NewHotResetManager(model, deviceNum, common.EmptyBoardId)
}

// TestGetChipCountOnRing for test the default count of ring ond different device
func TestGetChipCountOnRing(t *testing.T) {
	convey.Convey("test GetChipCountOnRing", t, func() {
		convey.Convey("test 910 chip count on ring success", func() {
			ascend910HotResetManager := newTestHotResetManager(api.Ascend910A, common.Train,
				common.Ascend910BRingsNumTrain)
			convey.So(ascend910HotResetManager, convey.ShouldNotBeNil)
			resetDevNumOnce, err := ascend910HotResetManager.GetResetDevNumOnce()
			convey.So(resetDevNumOnce, convey.ShouldEqual, common.Ascend910RingsNum)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test 910B train chip count on ring success", func() {
			ascend910BTrainHotResetManager := newTestHotResetManager(api.Ascend910B, common.Train,
				common.Ascend910BRingsNumTrain)
			convey.So(ascend910BTrainHotResetManager, convey.ShouldNotBeNil)
			resetDevNumOnce, err := ascend910BTrainHotResetManager.GetResetDevNumOnce()
			convey.So(resetDevNumOnce, convey.ShouldEqual, common.Ascend910BRingsNumTrain)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test 910B Infer chip count on ring success", func() {
			ascend910BInferHotResetManager := newTestHotResetManager(api.Ascend910B, common.Infer,
				common.Ascend910BRingsNumTrain)
			convey.So(ascend910BInferHotResetManager, convey.ShouldNotBeNil)
			resetDevNumOnce, err := ascend910BInferHotResetManager.GetResetDevNumOnce()
			convey.So(resetDevNumOnce, convey.ShouldEqual, common.Ascend910BRingsNumTrain)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test 910A3 chip count on ring success", func() {
			ascend910A3HotResetManager := newTestHotResetManager(api.Ascend910A3, common.Train,
				common.Ascend910A3RingsNum)
			convey.So(ascend910A3HotResetManager, convey.ShouldNotBeNil)
			resetDevNumOnce, err := ascend910A3HotResetManager.GetResetDevNumOnce()
			convey.So(resetDevNumOnce, convey.ShouldEqual, common.Ascend910A3RingsNum)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetChipCountOnRing for test the default count of ring ond different device
func TestGetChipCountOnRing2(t *testing.T) {
	convey.Convey("test GetChipCountOnRing", t, func() {
		convey.Convey("test 910 chip count on ring success", func() {
			a200A2HotResetManager := newTestHotResetManager(api.Ascend910B, common.Train,
				common.A200TA2RingsNum)
			convey.So(a200A2HotResetManager, convey.ShouldNotBeNil)
			resetDevNumOnce, err := a200A2HotResetManager.GetResetDevNumOnce()
			convey.So(resetDevNumOnce, convey.ShouldEqual, common.A200TA2RingsNum)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
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
		convey.Convey("01-taskDevFaultInfo is not nil, should return nil", func() {
			tool := &HotResetTools{}
			convey.So(tool.allTaskDevFaultInfo, convey.ShouldBeNil)
			taskDevList := map[string][]*common.TaskDevInfo{"test": {}}
			err := tool.UpdateTaskDevFaultInfoCache(taskDevList)
			convey.So(err, convey.ShouldBeNil)
			convey.So(tool.allTaskDevFaultInfo, convey.ShouldNotBeNil)
		})
		convey.Convey("02-taskDevFaultInfo is nil, should return error", func() {
			tool := &HotResetTools{}
			convey.So(tool.allTaskDevFaultInfo, convey.ShouldBeNil)
			var taskDevList map[string][]*common.TaskDevInfo
			err := tool.UpdateTaskDevFaultInfoCache(taskDevList)
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

// TestHandlePodAddEvent for test handlePodAddEvent
func TestHandlePodAddEvent(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test handlePodAddEvent", t, func() {
		patch := gomonkey.ApplyPrivateMethod(new(HotResetTools), "getPodFromCache", func(_ *HotResetTools,
			_ string) (*v1.Pod, error) {
			return nil, errors.New("pod not found")
		})
		mokeEvent := kubeclient.Event{
			Resource: kubeclient.PodResource,
			Key:      fakePod,
		}
		defer patch.Reset()
		convey.Convey("01-add pod event, should return true", func() {
			mokeEvent.Type = kubeclient.EventTypeAdd
			ascend910HotResetManager.queue.Add(mokeEvent)
			convey.So(ascend910HotResetManager.processNextWorkItem(), convey.ShouldBeTrue)
		})
		convey.Convey("02-delete pod event, should return true", func() {
			mokeEvent.Type = kubeclient.EventTypeDelete
			ascend910HotResetManager.queue.Add(mokeEvent)
			convey.So(ascend910HotResetManager.processNextWorkItem(), convey.ShouldBeTrue)
		})
		convey.Convey("03-default pod event, should return true", func() {
			mokeEvent.Type = kubeclient.EventTypeUpdate
			ascend910HotResetManager.queue.Add(mokeEvent)
			convey.So(ascend910HotResetManager.processNextWorkItem(), convey.ShouldBeTrue)
		})
		convey.Convey("04-update cm event, should return true", func() {
			mokeEvent.Type = kubeclient.EventTypeUpdate
			mokeEvent.Resource = kubeclient.CMResource
			ascend910HotResetManager.queue.Add(mokeEvent)
			convey.So(ascend910HotResetManager.processNextWorkItem(), convey.ShouldBeTrue)
		})
		convey.Convey("04-delete cm event, should return true", func() {
			mokeEvent.Type = kubeclient.EventTypeDelete
			mokeEvent.Resource = kubeclient.CMResource
			ascend910HotResetManager.queue.Add(mokeEvent)
			convey.So(ascend910HotResetManager.processNextWorkItem(), convey.ShouldBeTrue)
		})
		convey.Convey("06-default cm event, should return true", func() {
			mokeEvent.Type = kubeclient.EventTypeAdd
			mokeEvent.Resource = kubeclient.CMResource
			ascend910HotResetManager.queue.Add(mokeEvent)
			convey.So(ascend910HotResetManager.processNextWorkItem(), convey.ShouldBeTrue)
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
		convey.Convey("invalid key slice length", func() {
			event.Key = "default/fake-pod/pod1"
			ascend910HotResetManager.jobs[event.Key] = "fake-job"
			ascend910HotResetManager.handlePodDeleteEvent(event)
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

// TestWriteCMToFileCase1 test write cm to file
func TestWriteCMToFileCase1(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test writeCMToFile case1", t, func() {
		convey.Convey("write cm to file failed when cm has not reset.json", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "reset-config-test1",
				},
				Data: map[string]string{"xxx": "yyy"},
			}
			err := ascend910HotResetManager.writeCMToFile(cm)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("write cm to file failed when dir is not exist", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "reset-config-test2",
				},
				Data: map[string]string{common.ResetInfoCMDataKey: "yyy"},
			}
			err := ascend910HotResetManager.writeCMToFile(cm)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("write cm to file success when dir is exist", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "reset-config-test2",
				},
				Data: map[string]string{common.ResetInfoCMDataKey: "yyy", common.ResetInfoTypeKey: "zzz"},
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

// TestWriteCMToFileCase2 test write cm to file
func TestWriteCMToFileCase2(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test writeCMToFile case2", t, func() {
		convey.Convey("write cm to file failed when cm has not profilingSwitch", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "data-trace-test1",
				},
				Data: map[string]string{"xxx": "yyy"},
			}
			err := ascend910HotResetManager.writeCMToFile(cm)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("write cm to file success when dir is exist", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "data-trace-test1",
				},
				Data: map[string]string{common.DataTraceCmProfilingSwitchKey: "xxx"},
			}
			patch := gomonkey.ApplyFuncReturn(common.WriteToFile, nil)
			defer patch.Reset()
			err := ascend910HotResetManager.writeCMToFile(cm)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestHandleConfigMapEvent for test handleConfigMapEvent
func TestHandleConfigMapEvent(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test handleConfigMapEvent", t, func() {
		convey.Convey("01-default event type, forget event", func() {
			mokeEvent := kubeclient.Event{Type: kubeclient.EventTypeAdd}
			ascend910HotResetManager.handleConfigMapEvent(mokeEvent)
		})
	})
}

// TestHandleCMAddEventCase1 test of handleCMAddEvent
func TestHandleCMAddEventCase1(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test handleCMAddEvent case1", t, func() {
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
					Name:      "reset-config-test",
					Namespace: "default",
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

		convey.Convey("os stat return error", func() {
			ascend910HotResetManager.handleCMUpdateEvent(mokeEvent)
		})
		convey.Convey("cm obj will return false", func() {
			patch3 := gomonkey.ApplyFuncReturn(os.Stat, nil, nil)
			defer patch3.Reset()

			ascend910HotResetManager.handleCMUpdateEvent(mokeEvent)
		})
	})
}

// TestHandleCMAddEventCase2 test of handleCMAddEvent
func TestHandleCMAddEventCase2(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test handleCMAddEvent case2", t, func() {
		mokeEvent := kubeclient.Event{
			Resource: kubeclient.CMResource,
			Key:      "default/data-trace-test",
			Type:     kubeclient.EventTypeAdd,
		}
		ascend910HotResetManager.queue.Add(mokeEvent)
		patch1 := gomonkey.ApplyMethod(new(HotResetTools), "GetCMFromCache", func(_ *HotResetTools,
			_ string) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "data-trace-test",
					Namespace: "default",
				},
				Data: map[string]string{common.DataTraceCmProfilingSwitchKey: "YYY"},
			}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyPrivateMethod(new(HotResetTools), "writeCmToFileSystem", func(_ *HotResetTools,
			_ *v1.ConfigMap, _ string, _ string, _ interface{}) error {
			return nil
		})
		defer patch2.Reset()

		convey.Convey("os stat return error", func() {
			ascend910HotResetManager.handleCMUpdateEvent(mokeEvent)
		})
		convey.Convey("cm obj will return false", func() {
			patch3 := gomonkey.ApplyFuncReturn(os.Stat, nil, nil)
			defer patch3.Reset()

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

// TestWriteCmToFileSystem test write file to system
func TestWriteCmToFileSystem(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	cmName := "default/data-trace-test-default"
	fileContent := "YYY"
	convey.Convey("test handleCMDeleteEvent", t, func() {
		convey.Convey("test handle delete event success", func() {
			mokeEvent := kubeclient.Event{
				Resource: kubeclient.CMResource,
				Key:      "default/" + cmName,
				Type:     kubeclient.EventTypeAdd,
			}
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: "default",
				},
				Data: map[string]string{common.DataTraceCmProfilingSwitchKey: fileContent},
			}
			ascend910HotResetManager.queue.Add(mokeEvent)
			dir := fmt.Sprintf("%s/%s", common.DataTraceConfigDir, cm.Namespace+"."+cm.Name)
			fileFullName := filepath.Join(dir, common.DataTraceCmProfilingSwitchKey)
			err := ascend910HotResetManager.writeCmToFileSystem(cm, common.DataTraceCmProfilingSwitchKey,
				fileFullName, mokeEvent)
			convey.ShouldBeNil(err)
			_, err = os.Stat(fileFullName)
			convey.ShouldBeNil(err)
			content, err := os.ReadFile(fileFullName)
			convey.ShouldBeNil(err)
			convey.ShouldEqual(string(content), fileContent)
			err = os.RemoveAll(fileFullName)
			convey.ShouldBeNil(err)
		})
	})
}

// TestWriteCmToFileSystemWithoutKey test write file to system but cm not key
func TestWriteCmToFileSystemWithoutKey(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	cmName := "default/data-trace-test-default"
	fileContent := "YYY"
	convey.Convey("test handleCMDeleteEvent", t, func() {
		convey.Convey("test handle delete event success", func() {
			mokeEvent := kubeclient.Event{
				Resource: kubeclient.CMResource,
				Key:      "default/" + cmName,
				Type:     kubeclient.EventTypeAdd,
			}
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: "default",
				},
				Data: map[string]string{common.DataTraceCmProfilingSwitchKey + "fake": fileContent},
			}
			ascend910HotResetManager.queue.Add(mokeEvent)
			dir := fmt.Sprintf("%s/%s", common.DataTraceConfigDir, cm.Namespace+"."+cm.Name)
			fileFullName := filepath.Join(dir, common.DataTraceCmProfilingSwitchKey)
			err := ascend910HotResetManager.writeCmToFileSystem(cm, common.DataTraceCmProfilingSwitchKey,
				fileFullName, mokeEvent)
			convey.ShouldNotBeNil(err)
		})
	})
}

func newHotResetTools() *HotResetTools {
	return &HotResetTools{
		resetDevNumOnce:  common.Ascend910RingsNum,
		resetTask:        map[string]struct{}{},
		resetDev:         map[int32]struct{}{},
		faultDev2PodMap:  map[int32]v1.Pod{},
		jobs:             map[string]string{},
		noResetCmPodKeys: map[string]struct{}{},
		queue:            workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		cmIndexer:        cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{}),
		podIndexer:       cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{}),
	}
}

// TestIsTaskDevListChange for test isTaskDevListChange
func TestIsTaskDevListChange(t *testing.T) {
	convey.Convey("test isTaskDevListChange", t, func() {
		tool := &HotResetTools{}
		tool.allTaskDevList = map[string][]int32{"fake task": {1, 2, 3}}
		// 01-task is not in allTaskDevList, should return false
		convey.So(tool.isTaskDevListChange("true task", nil), convey.ShouldBeFalse)
		// 02-task is in allTaskDevList, new task dev list is empty, should return false
		convey.So(tool.isTaskDevListChange("fake task", nil), convey.ShouldBeFalse)
		// 03-task is in allTaskDevList, new task dev list is not empty, should return true
		newTaskDevList := map[string][]int32{"fake task": {1, 2}}
		convey.So(tool.isTaskDevListChange("fake task", newTaskDevList), convey.ShouldBeTrue)
	})
}

func TestGetTaskNameByPod(t *testing.T) {
	convey.Convey("test GetTaskNameByPod", t, func() {
		// 01-get task name by pod success, should return target task name
		pod := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-pod",
				Labels: map[string]string{
					common.ResetTaskNameKey: "task-name",
				},
			},
		}
		tool := &HotResetTools{}
		convey.So(tool.GetTaskNameByPod(pod), convey.ShouldEqual, "task-name")
	})
}

func TestGetResetDevNumOnce(t *testing.T) {
	convey.Convey("test GetResetDevNumOnce", t, func() {
		convey.Convey("01-resetDevNumOnce is zero, should return error", func() {
			tool := &HotResetTools{resetDevNumOnce: 0}
			result, err := tool.GetResetDevNumOnce()
			convey.So(result, convey.ShouldEqual, 0)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-resetDevNumOnce is not zero, should return value", func() {
			tool := &HotResetTools{resetDevNumOnce: 4}
			result, err := tool.GetResetDevNumOnce()
			convey.So(result, convey.ShouldEqual, 4)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetGlobalDevFaultInfo(t *testing.T) {
	convey.Convey("test GetGlobalDevFaultInfo", t, func() {
		convey.Convey("01-logicID not in cache, should return error", func() {
			tool := &HotResetTools{globalDevFaultInfo: map[int32]*common.DevFaultInfo{}}
			result, err := tool.GetGlobalDevFaultInfo(0)
			convey.So(result, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-logicID in cache, should return info", func() {
			info := &common.DevFaultInfo{LogicId: 0, Policy: common.ResetError}
			tool := &HotResetTools{globalDevFaultInfo: map[int32]*common.DevFaultInfo{0: info}}
			result, err := tool.GetGlobalDevFaultInfo(0)
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(result.LogicId, convey.ShouldEqual, int32(0))
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
