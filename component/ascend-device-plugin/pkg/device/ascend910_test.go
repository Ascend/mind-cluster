/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package device a series of device function
package device

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/devmanager"
	devcommon "ascend-common/devmanager/common"
)

const (
	chipPhyID0         = 0
	chipPhyID1         = 1
	chipPhyID2         = 2
	chipPhyID3         = 3
	chipPhyID4         = 4
	chipPhyID5         = 5
	chipPhyID6         = 6
	chipPhyID7         = 7
	ascend910LogicID0  = "Ascend910-0"
	ascend910LogicID1  = "Ascend910-1"
	ascend910LogicID2  = "Ascend910-2"
	ascend910LogicID3  = "Ascend910-3"
	ascend910LogicID4  = "Ascend910-4"
	ascend910LogicID5  = "Ascend910-5"
	ascend910LogicID6  = "Ascend910-6"
	ascend910LogicID7  = "Ascend910-7"
	A800IA2WithHccsOld = 0x34
	A800IA2WithHccs    = 0x3d
)

func createFake910Manager() *HwAscend910Manager {
	manager := NewHwAscend910Manager()
	manager.SetDmgr(&devmanager.DeviceManagerMock{})
	return manager
}

func createFakeDeviceInfo() *common.NodeDeviceInfoCache {
	return &common.NodeDeviceInfoCache{
		DeviceInfo: common.NodeDeviceInfo{
			DeviceList: map[string]string{},
		},
		CheckCode: "",
	}
}

func TestHwAscend910ManagerGetNPUs(t *testing.T) {
	convey.Convey("910 test GetNPUs", t, func() {
		manager := createFake910Manager()
		allInfo, err := manager.GetNPUs()
		convey.So(err, convey.ShouldBeNil)
		convey.So(allInfo.AllDevTypes[0], convey.ShouldEqual, common.Ascend910)
		convey.So(allInfo.AllDevs[0].DeviceName, convey.ShouldEqual,
			fmt.Sprintf("%s-%d", common.Ascend910, allInfo.AllDevs[0].PhyID))
	})
}

func TestDoWithVolcanoListAndWatch910(t *testing.T) {
	convey.Convey("910 test DoWithVolcanoListAndWatch", t, func() {
		manager := createFake910Manager()
		fakeKubeInteractor := &kubeclient.ClientK8s{Clientset: nil, NodeName: "NODE_NAME"}
		manager.SetKubeClient(fakeKubeInteractor)
		allInfo, err := manager.GetNPUs()
		convey.So(err, convey.ShouldBeNil)
		groupDevice := ClassifyDevices(allInfo.AllDevs, allInfo.AllDevTypes)
		mockGetPodsUsedNpu := mockGetPodsUsedNpu()
		mockGetConfigMap := mockGetDeviceInfoCMCache(map[string]string{common.Ascend910: ascend910LogicID1})
		mockPatchNodeState := mockPatchNodeState()
		mockCreateConfigMap := mockWriteDeviceInfoDataIntoCM()
		mockNodeBack := mockGetNode()
		defer func() {
			mockGetPodsUsedNpu.Reset()
			mockGetConfigMap.Reset()
			mockPatchNodeState.Reset()
			mockCreateConfigMap.Reset()
			mockNodeBack.Reset()
		}()
		manager.client.SetNodeDeviceInfoCache(createFakeDeviceInfo())
		manager.DoWithVolcanoListAndWatch(groupDevice)
	})
}

func mockGetNode() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetNode",
		func(_ *kubeclient.ClientK8s) (*v1.Node, error) {
			curNode := &v1.Node{}
			curNode.Labels = make(map[string]string, 1)
			return curNode, nil
		})
}

func mockWriteDeviceInfoDataIntoCM() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
		"WriteDeviceInfoDataIntoCM", func(_ *kubeclient.ClientK8s,
			deviceInfo map[string]string, manuallySeparateNPU string, _ common.SwitchFaultInfo, superPodID,
			serverIndex int32) (*common.NodeDeviceInfoCache, error) {
			return &common.NodeDeviceInfoCache{}, nil
		})
}

func mockPatchNodeState() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
		"PatchNodeState", func(_ *kubeclient.ClientK8s, curNode,
			newNode *v1.Node) (*v1.Node, []byte, error) {
			return &v1.Node{}, nil, nil
		})
}

func mockGetDeviceInfoCMCache(deviceList map[string]string) *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
		"GetDeviceInfoCMCache", func(_ *kubeclient.ClientK8s) *common.NodeDeviceInfoCache {
			nodeDeviceData := common.NodeDeviceInfoCache{DeviceInfo: common.NodeDeviceInfo{
				DeviceList: deviceList,
				UpdateTime: time.Now().Unix()}}
			nodeDeviceData.CheckCode = common.MakeDataHash(nodeDeviceData.DeviceInfo)
			return &nodeDeviceData
		})
}

func mockGetPodsUsedNpu() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
		"GetPodsUsedNpu", func(_ *kubeclient.ClientK8s) sets.String {
			return nil
		})
}

func TestToStandardDeviceFmt(t *testing.T) {
	convey.Convey("910 test toStandardDeviceFmt", t, func() {
		hnm := NewHwAscend910Manager()
		devices := sets.String{}.Insert("test910")
		res := hnm.toStandardDeviceFmt(devices)
		convey.So(len(res), convey.ShouldEqual, 1)
	})
}

func TestGetPatchLabel(t *testing.T) {
	convey.Convey("910 getPatchLabel", t, func() {
		hnm := NewHwAscend910Manager()
		devices := sets.String{}.Insert("100-1")
		devices.Insert("100-2")
		res := hnm.getPatchLabel(devices)
		convey.So(res, convey.ShouldBeIn, []string{"1.2", "2.1"})
	})
}

// TestGraceTolerance an ut for function GraceTolerance
func TestGraceTolerance(t *testing.T) {
	manager := createFake910Manager()
	common.ParamOption.RealCardType = common.Ascend910
	convey.Convey("exec ut function GraceTolerance", t, func() {
		mockPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetAllPodList",
			func(_ *kubeclient.ClientK8s) (*v1.PodList, error) {
				return mockGetAllPodList(), nil
			})
		mockGetCM := mockGetCM()
		defer mockGetCM.Reset()
		defer mockPodList.Reset()
		patch := gomonkey.ApplyMethod(new(HotResetTools), "SyncResetCM",
			func(_ *HotResetTools, _ *kubeclient.ClientK8s) { return })
		defer patch.Reset()
		manager.GraceTolerance(mockGroupDevice())
		convey.So(manager.hotResetManager, convey.ShouldNotBeNil)
	})
}

// TestHotResetHandler an ut for function hotResetHandler
func TestHotResetHandler(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("exec ut function hotResetHandler", t, func() {
		mockHandleResetProcess := gomonkey.ApplyFunc((*HwAscend910Manager).handleResetProcess,
			func(ascend910Manager *HwAscend910Manager, classifyDevs map[string][]*common.NpuDevice,
				devInfo *common.DevFaultInfo, npuDev *common.NpuDevice) {
				return
			}).ApplyMethodReturn(&HotResetTools{}, "GetResetDevNumOnce", common.Ascend910RingsNum, nil)
		defer mockHandleResetProcess.Reset()
		// have L4 error, device busy, reset should be down
		// device busy
		mockPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetAllPodList",
			func(_ *kubeclient.ClientK8s) (*v1.PodList, error) {
				return mockGetAllPodList(), nil
			})
		defer mockPodList.Reset()
		manager.hotResetManager = &HotResetTools{
			globalDevFaultInfo: mockDevFaultInfoL4(),
		}
		isHotResetOn = false
		err := manager.hotResetHandler(mockGroupDevice())
		convey.So(isHotResetOn, convey.ShouldBeFalse)
		convey.So(err, convey.ShouldBeNil)

		// have L5 error, device busy, reset should be down
		manager.hotResetManager = &HotResetTools{
			globalDevFaultInfo: mockDevFaultInfoL5(),
		}
		isHotResetOn = false
		err = manager.hotResetHandler(mockGroupDevice())
		convey.So(isHotResetOn, convey.ShouldBeFalse)
		convey.So(err, convey.ShouldBeNil)

		// no L4 L5 error, device busy, reset should be down
		manager.hotResetManager = &HotResetTools{
			globalDevFaultInfo: mockDevFaultInfoNoL4L5(),
		}
		isHotResetOn = false
		err = manager.hotResetHandler(mockGroupDevice())
		convey.So(isHotResetOn, convey.ShouldBeFalse)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestCanBeReset an ut for function canBeReset
func TestCanBeReset(t *testing.T) {
	manager := createFake910Manager()
	manager.hotResetManager = newTestHotResetManager(common.Ascend910, common.Train)
	common.ParamOption.RealCardType = common.Ascend910B

	convey.Convey("exec ut function canBeReset", t, func() {
		// empty situation
		mockPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetAllPodList",
			func(_ *kubeclient.ClientK8s) (*v1.PodList, error) {
				return mockOneEmptyPodList(), nil
			})
		defer mockPodList.Reset()
		resultBool, err := manager.canBeReset(mockSingleDevFaultInfo())
		convey.So(resultBool, convey.ShouldBeTrue)
		convey.So(err, convey.ShouldBeNil)

		// chip busy situation
		mockPodList = gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetAllPodList",
			func(_ *kubeclient.ClientK8s) (*v1.PodList, error) {
				return mockGetAllPodList(), nil
			})
		defer mockPodList.Reset()
		resultBool, err = manager.canBeReset(mockSingleDevFaultInfo())
		convey.So(resultBool, convey.ShouldBeFalse)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestGetBusyChipListFromPod an ut for function getBusyChipListFromPod
func TestGetBusyChipListFromPod(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("exec ut function getBusyChipListFromPod", t, func() {
		fakePods := mockGetAllPodList()
		emptyPod := mockOneEmptyPodList()
		devList := manager.getBusyChipListFromPod(fakePods)
		emptyDevList := manager.getBusyChipListFromPod(emptyPod)
		resultList := []string{ascend910LogicID0, ascend910LogicID1, "",
			ascend910LogicID4, ascend910LogicID5, ascend910LogicID6, ascend910LogicID7}
		convey.So(devList, convey.ShouldResemble, resultList)
		convey.So(emptyDevList, convey.ShouldResemble, []string{""})
	})
}

// TestIsChipActive an ut for function isChipActive
func TestIsChipActive(t *testing.T) {
	manager := createFake910Manager()
	var logicID int32 = 0
	convey.Convey("exec ut function isChipActive", t, func() {
		// empty list
		var busyChipList []string
		activity, err := manager.isChipActive(logicID, busyChipList)
		convey.So(activity, convey.ShouldBeTrue)
		convey.So(err, convey.ShouldBeNil)
		// busy chip not match
		busyChipList = []string{ascend910LogicID1}
		activity, err = manager.isChipActive(logicID, busyChipList)
		convey.So(activity, convey.ShouldBeTrue)
		convey.So(err, convey.ShouldBeNil)
		// busy chip match current chip
		busyChipList = []string{ascend910LogicID0}
		activity, err = manager.isChipActive(logicID, busyChipList)
		convey.So(activity, convey.ShouldBeFalse)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestExecHotReset an ut for function execHotReset
func TestExecHotReset(t *testing.T) {
	manager := createFake910Manager()
	manager.hotResetManager = newTestHotResetManager(common.Ascend910, common.Train)
	devInfo := mockSingleDevFaultInfo()
	common.ParamOption.RealCardType = common.Ascend910B
	convey.Convey("exec ut function execHotReset", t, func() {
		mockIsShouldCheckNet := gomonkey.ApplyFunc((*HwAscend910Manager).isShouldCheckNet,
			func(_ *HwAscend910Manager, logicID int32) bool {
				return false
			})
		// after change mockBootStartFinish value in npu-exporter we could delete mockHotResetComplete
		mockHotResetComplete := gomonkey.ApplyFunc((*HwAscend910Manager).waitDeviceResetComplete,
			func(_ *HwAscend910Manager, logicId int32, totalTime *int, shouldCheckNet bool) error {
				return nil
			})
		defer mockIsShouldCheckNet.Reset()
		defer mockHotResetComplete.Reset()
		err := manager.execHotReset(devInfo)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestSetAllDevUnhealthyOnRing an ut for function setAllDevUnhealthyOnRing
func TestSetAllDevUnhealthyOnRing(t *testing.T) {
	manager := createFake910Manager()
	patch := gomonkey.ApplyMethodReturn(&HotResetTools{}, "GetResetDevNumOnce", common.Ascend910RingsNum, nil)
	defer patch.Reset()
	convey.Convey("exec ut function setAllDevUnhealthyOnRing", t, func() {
		devList := mockGroupDevice()
		devStatusList := devList[common.Ascend910]
		manager.hotResetManager = &HotResetTools{
			resetDevNumOnce: 8,
		}
		inResetDev = -1
		common.ParamOption.RealCardType = common.Ascend910B

		// no reset device situation
		isHotResetOn = false
		err := manager.setAllDevUnhealthyOnRing(devList)
		for i := 0; i < 8; i++ {
			convey.So(devStatusList[i].Health, convey.ShouldEqual, v1beta1.Healthy)
			convey.So(devStatusList[i].NetworkHealth, convey.ShouldEqual, v1beta1.Unhealthy)
		}
		convey.So(err, convey.ShouldBeNil)

		// is doing hot reset situation
		convey.So(inResetDev, convey.ShouldEqual, -1)
		inResetDev = 0
		isHotResetOn = true
		err = manager.setAllDevUnhealthyOnRing(devList)
		for i := 0; i < 8; i++ {
			convey.So(devStatusList[i].NetworkHealth, convey.ShouldEqual, v1beta1.Unhealthy)
		}
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestTryResetDevice an ut for function tryResetDevice
func TestTryResetDevice(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("exec ut function tryResetDevice", t, func() {
		err := manager.tryResetDevice(0, 0)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestIsRingResetComplete an ut for function isRingResetComplete
func TestIsRingResetComplete(t *testing.T) {
	manager := createFake910Manager()
	manager.hotResetManager = newTestHotResetManager(common.Ascend910, common.Train)
	common.ParamOption.RealCardType = common.Ascend910B
	var logicID int32 = 0
	convey.Convey("exec ut function isRingResetComplete", t, func() {
		// after change mockBootStartFinish value in npu-exporter we could delete mockHotResetComplete
		mockHotResetComplete := gomonkey.ApplyFunc((*HwAscend910Manager).waitDeviceResetComplete,
			func(_ *HwAscend910Manager, logicId int32, totalTime *int, shouldCheckNet bool) error {
				return nil
			})
		defer mockHotResetComplete.Reset()
		err := manager.isRingResetComplete(logicID, false)
		convey.So(err, convey.ShouldBeNil)
		common.ParamOption.RealCardType = common.Ascend910
		err = manager.isRingResetComplete(logicID, false)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestProcessAllTask an ut for function processAllTask
func TestProcessAllTask(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("exec ut function TestProcessAllTask", t, func() {
		convey.Convey("01-all task dev fault info list is empty, should return nil", func() {

		})
		patch := mockGetCM().
			ApplyPrivateMethod(&HwAscend910Manager{}, "isolateSceneHandle",
				func(*HwAscend910Manager, int32) bool { return false }).
			ApplyPrivateMethod(&HwAscend910Manager{}, "preProcess",
				func(*HwAscend910Manager, string, string) (*common.TaskResetInfo, error) {
					return nil, nil
				}).
			ApplyPrivateMethod(&HwAscend910Manager{}, "runProcessTask",
				func(*HwAscend910Manager, string, int, *common.TaskResetInfo, map[string][]*common.NpuDevice) error {
					return nil
				}).
			ApplyGlobalVar(&isHotResetOn, false)
		defer patch.Reset()
		manager.hotResetManager = &HotResetTools{
			allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{
				"task1": getTaskInfo(),
				"task2": getTaskInfo1(),
			},
			taskPod: map[string]v1.Pod{
				"task1": getSinglePod("pod1", map[string]string{}),
				"task2": getSinglePod("pod2", map[string]string{}),
			},
		}
		err := manager.processAllTask(mockGroupDevice())
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestFilterDevStatus an ut for function filterDevStatus
func TestFilterDevStatus(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("exec ut function TestFilterDevStatus", t, func() {
		err := manager.filterDevStatus(map[string][]*common.NpuDevice{})
		convey.So(err, convey.ShouldNotBeNil)
		mockGetCM := mockGetCM()
		mockUpdateCM := mockUpdateCM()
		patch := gomonkey.ApplyPrivateMethod(&HwAscend910Manager{}, "isDevShouldBeIsolate",
			func(*HwAscend910Manager, int32) bool { return false })
		defer mockGetCM.Reset()
		defer mockUpdateCM.Reset()
		defer patch.Reset()
		manager.hotResetManager = &HotResetTools{
			resetDevNumOnce: common.Ascend910RingsNum,
			resetDev: map[int32]struct{}{
				chipPhyID1: {},
				chipPhyID3: {},
				chipPhyID5: {},
			},
			faultDev2PodMap: map[int32]v1.Pod{
				chipPhyID3: getSinglePod("pod1", map[string]string{}),
			},
		}
		devices := mockGroupDevice()
		devices[common.Ascend910][chipPhyID1].Health = v1beta1.Unhealthy
		devices[common.Ascend910][chipPhyID3].Health = v1beta1.Unhealthy
		devices[common.Ascend910][chipPhyID5].Health = v1beta1.Unhealthy
		err = manager.filterDevStatus(devices)
		convey.So(err, convey.ShouldBeNil)
	})
}

func mockSingleDevFaultInfo() *common.DevFaultInfo {
	return &common.DevFaultInfo{LogicId: chipPhyID0}
}

func mockDevFaultInfoL4() map[int32]*common.DevFaultInfo {
	return map[int32]*common.DevFaultInfo{
		chipPhyID0: {
			LogicId: chipPhyID0,
			Policy:  "NotExist",
		},
		chipPhyID1: {
			LogicId: chipPhyID1,
			Policy:  "NotExist",
		},
		chipPhyID2: {
			LogicId: chipPhyID2,
			Policy:  common.FreeResetError,
		},
		chipPhyID3: {
			LogicId: chipPhyID3,
			Policy:  "NotExist",
		},
		chipPhyID4: {
			LogicId: chipPhyID4,
			Policy:  "NotExist",
		},
		chipPhyID5: {
			LogicId: chipPhyID5,
			Policy:  "NotExist",
		},
		chipPhyID6: {
			LogicId: chipPhyID1,
			Policy:  "NotExist",
		},
		chipPhyID7: {
			LogicId: chipPhyID1,
			Policy:  "NotExist",
		},
	}
}

func mockDevFaultInfoL5() map[int32]*common.DevFaultInfo {
	return map[int32]*common.DevFaultInfo{
		chipPhyID0: {
			LogicId: chipPhyID0,
			Policy:  "NotExist",
		},
		chipPhyID1: {
			LogicId: chipPhyID1,
			Policy:  "NotExist",
		},
		chipPhyID2: {
			LogicId: chipPhyID2,
			Policy:  common.ResetError,
		},
		chipPhyID3: {
			LogicId: chipPhyID3,
			Policy:  "NotExist",
		},
		chipPhyID4: {
			LogicId: chipPhyID4,
			Policy:  "NotExist",
		},
		chipPhyID5: {
			LogicId: chipPhyID5,
			Policy:  "NotExist",
		},
		chipPhyID6: {
			LogicId: chipPhyID1,
			Policy:  "NotExist",
		},
		chipPhyID7: {
			LogicId: chipPhyID1,
			Policy:  "NotExist",
		},
	}
}

func mockDevFaultInfoNoL4L5() map[int32]*common.DevFaultInfo {
	return map[int32]*common.DevFaultInfo{
		chipPhyID0: {
			LogicId: chipPhyID0,
			Policy:  "NotExist",
		},
		chipPhyID1: {
			LogicId: chipPhyID1,
			Policy:  common.NotHandleFault,
		},
		chipPhyID2: {
			LogicId: chipPhyID2,
			Policy:  common.RestartRequest,
		},
		chipPhyID3: {
			LogicId: chipPhyID3,
			Policy:  common.SeparateNPU,
		},
		chipPhyID4: {
			LogicId: chipPhyID4,
			Policy:  "NotExist",
		},
		chipPhyID5: {
			LogicId: chipPhyID5,
			Policy:  "NotExist",
		},
		chipPhyID6: {
			LogicId: chipPhyID1,
			Policy:  "NotExist",
		},
		chipPhyID7: {
			LogicId: chipPhyID1,
			Policy:  "NotExist",
		},
	}
}

func mockGetCM() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
		"GetConfigMap", func(_ *kubeclient.ClientK8s, _ string, _ string) (*v1.ConfigMap, error) {
			nodeDeviceData := common.TaskResetInfo{
				UpdateTime: 11111111,
			}
			return &v1.ConfigMap{Data: map[string]string{
				common.ResetInfoCMDataKey:      string(common.MarshalData(nodeDeviceData)),
				common.ResetInfoCMCheckCodeKey: common.MakeDataHash(nodeDeviceData)},
			}, nil
		})
}

func mockUpdateCM() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "UpdateConfigMap",
		func(_ *kubeclient.ClientK8s, _ *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{Data: map[string]string{}}, nil
		})
}

func mockGetAllPodList() *v1.PodList {
	annotationHalfRing := map[string]string{
		common.HuaweiAscend910: "Ascend910-0,Ascend910-1",
	}
	annotationEmpty := map[string]string{
		common.HuaweiAscend910: "",
	}
	annotationErr := map[string]string{}
	annotationErrRank := map[string]string{
		common.ResetTaskNameKey: "task1",
	}
	annotationSuccess := map[string]string{
		common.ResetTaskNameKey: "task1",
		common.RankIndexKey:     "1",
		common.HuaweiAscend910:  "Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7",
	}
	return &v1.PodList{
		Items: []v1.Pod{
			getSinglePod("test-pod1", annotationHalfRing),
			getSinglePod("test-pod2", annotationEmpty),
			getSinglePod("test-pod3", annotationErr),
			getSinglePod("test-pod4", annotationErrRank),
			getSinglePod("test-pod5", annotationSuccess),
		},
	}
}

func mockOneEmptyPodList() *v1.PodList {
	annotationEmpty := map[string]string{
		common.HuaweiAscend910: "",
	}
	return &v1.PodList{
		Items: []v1.Pod{
			getSinglePod("test-pod2", annotationEmpty),
		},
	}
}

func mockGroupDevice() map[string][]*common.NpuDevice {
	return map[string][]*common.NpuDevice{
		common.Ascend910: mockNpuDevices(),
	}
}

func mockNpuDevices() []*common.NpuDevice {
	return []*common.NpuDevice{
		getNPU(chipPhyID0),
		getNPU(chipPhyID1),
		getNPU(chipPhyID2),
		getNPU(chipPhyID3),
		getNPU(chipPhyID4),
		getNPU(chipPhyID5),
		getNPU(chipPhyID6),
		getNPU(chipPhyID7),
	}
}

func getTaskInfo() []*common.TaskDevInfo {
	return []*common.TaskDevInfo{
		{
			DevFaultInfo: common.DevFaultInfo{
				LogicId: chipPhyID0,
				Policy:  "NotExist",
			},
		},
		{
			DevFaultInfo: common.DevFaultInfo{
				LogicId: chipPhyID1,
				Policy:  common.IsolateError,
			},
		},
		{
			DevFaultInfo: common.DevFaultInfo{
				LogicId: chipPhyID2,
				Policy:  common.RestartError,
			},
		},
	}
}

func getTaskInfo1() []*common.TaskDevInfo {
	return []*common.TaskDevInfo{
		{
			DevFaultInfo: common.DevFaultInfo{
				LogicId: chipPhyID0,
				Policy:  common.RestartError},
		},
		{
			DevFaultInfo: common.DevFaultInfo{
				LogicId: chipPhyID1,
				Policy:  common.ResetError},
		},
		{
			DevFaultInfo: common.DevFaultInfo{
				LogicId: chipPhyID2,
				Policy:  common.RestartRequestError},
		},
	}
}

func getNPU(autoID int32) *common.NpuDevice {
	return &common.NpuDevice{
		LogicID:       autoID,
		PhyID:         autoID,
		Health:        v1beta1.Healthy,
		NetworkHealth: v1beta1.Unhealthy,
		DevType:       common.Ascend910,
		DeviceName:    fmt.Sprintf("%s-%d", common.Ascend910, autoID),
	}
}

func getSinglePod(podName string, annotation map[string]string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        podName,
			Annotations: annotation,
		},
	}
}

// TestHwAscend910ManagerGetNeedResetDeviceLogicIdMap a ut for method getNeedResetDeviceLogicIdMap
func TestHwAscend910ManagerGetNeedResetDeviceLogicIdMap(t *testing.T) {
	ascendTools := AscendTools{}
	ascendTools.SetDmgr(&devmanager.DeviceManagerMock{})
	devFaultInfoList := []*common.TaskDevInfo{{RankId: 0, DevFaultInfo: common.DevFaultInfo{LogicId: 1}}}
	common.ParamOption.RealCardType = common.Ascend910
	tests := []struct {
		name    string
		want    map[int32]int32
		wantErr bool
	}{
		{
			name:    "getNeedResetDeviceLogicIdList ut",
			want:    map[int32]int32{0: 1, 1: 1, 2: 1, 3: 1},
			wantErr: false,
		},
	}
	mockGetNeedResetDevMapPatch := mockGetNeedResetDevMap()
	defer mockGetNeedResetDevMapPatch.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hnm := &HwAscend910Manager{
				AscendTools:     ascendTools,
				hotResetManager: newTestHotResetManager(common.Ascend910, common.Train),
			}
			got, err := hnm.getNeedResetDeviceLogicIdMap(devFaultInfoList)
			if (err != nil) != tt.wantErr || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNeedResetDeviceLogicIdList() error = %v, wantErr %v, got = %v, want %v", err, tt.wantErr, got, tt.want)
				return
			}
		})
	}
}

type args struct {
	faultDeviceLogicIdMap map[int32]int32
}

type FakeClient struct {
	name    string
	args    args
	want    bool
	wantErr bool
}

func mockTestFakeProcess() []FakeClient {
	return []FakeClient{
		{
			name: "checkNumberOfAllProcessIsZero ut 1",
			args: args{
				faultDeviceLogicIdMap: map[int32]int32{1: 1},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "checkNumberOfAllProcessIsZero ut 2",
			args: args{
				faultDeviceLogicIdMap: map[int32]int32{1: 1},
			},
			want:    true,
			wantErr: false,
		},
	}
}

// TestHwAscend910ManagerCheckNumberOfAllProcessIsZero a ut for method checkNumberOfAllProcessIsZero
func TestHwAscend910ManagerCheckNumberOfAllProcessIsZero(t *testing.T) {
	ascendTools := AscendTools{}
	ascendTools.SetDmgr(&devmanager.DeviceManagerMock{})
	hotResetManager := &HotResetTools{}
	common.ParamOption.RealCardType = common.Ascend910
	tests := mockTestFakeProcess()
	mockGetDevProcessInfoPatch := gomonkey.ApplyMethodSeq(reflect.TypeOf(&devmanager.DeviceManagerMock{}),
		"GetDevProcessInfo", []gomonkey.OutputCell{
			{Values: gomonkey.Params{&devcommon.DevProcessInfo{ProcNum: 1}, nil}},
			{Values: gomonkey.Params{&devcommon.DevProcessInfo{ProcNum: 0}, nil}},
			{Values: gomonkey.Params{&devcommon.DevProcessInfo{ProcNum: 0}, nil}},
			{Values: gomonkey.Params{&devcommon.DevProcessInfo{ProcNum: 0}, nil}},
			{Values: gomonkey.Params{&devcommon.DevProcessInfo{ProcNum: 0}, nil}},
		})
	defer mockGetDevProcessInfoPatch.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hnm := &HwAscend910Manager{
				AscendTools:     ascendTools,
				hotResetManager: hotResetManager,
			}
			got, err := hnm.checkNumberOfAllProcessIsZero(tt.args.faultDeviceLogicIdMap)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Errorf("checkNumberOfAllProcessIsZero() error = %v, wantErr %v, got = %v, want %v", err, tt.wantErr, got, tt.want)
				return
			}
		})
	}
}

// TestHwAscend910ManagerWaitForAllFaultyDeviceProcessesToZero a ut for method waitForAllFaultyDeviceProcessesToZero
func TestHwAscend910ManagerWaitForAllFaultyDeviceProcessesToZero(t *testing.T) {
	ascendTools := AscendTools{}
	ascendTools.SetDmgr(&devmanager.DeviceManagerMock{})
	hotResetManager := &HotResetTools{}
	hnm := &HwAscend910Manager{
		AscendTools:     ascendTools,
		hotResetManager: hotResetManager,
	}
	waitFlushingCMTime := 3
	common.WaitProcessReadCMTime = time.Duration(waitFlushingCMTime)
	common.ParamOption.RealCardType = common.Ascend910
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "waitForAllFaultyDeviceProcessesToZero ut for timeout",
			wantErr: true,
		},
		{
			name:    "normal ut for waitForAllFaultyDeviceProcessesToZero",
			wantErr: false,
		},
	}
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&devmanager.DeviceManagerMock{}),
		"GetDevProcessInfo", func(_ *devmanager.DeviceManagerMock, logicID int32) (*devcommon.DevProcessInfo, error) {
			return &devcommon.DevProcessInfo{ProcNum: 1}, nil
		}).ApplyMethodReturn(&HotResetTools{}, "GetResetDevNumOnce", common.Ascend910RingsNum, nil)
	defer patch.Reset()
	mockGetTaskProcessPolicyPatch := mockGetTaskProcessPolicy()
	defer mockGetTaskProcessPolicyPatch.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "waitForAllFaultyDeviceProcessesToZero ut for timeout" {
				mockGetNeedResetDevMapPatch := mockGetNeedResetDevMap()
				defer mockGetNeedResetDevMapPatch.Reset()
			}
			if err := hnm.waitForAllFaultyDeviceProcessesToZero("", nil); (err != nil) != tt.wantErr {
				t.Errorf("waitForAllFaultyDeviceProcessesToZero() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func mockGetNeedResetDevMap() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(HotResetTools)),
		"GetNeedResetDevMap", func(_ *HotResetTools, _ []*common.TaskDevInfo) (map[int32]int32, error) {
			return map[int32]int32{0: 1}, nil
		})
}

func mockGetTaskProcessPolicy() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(HotResetTools)),
		"GetTaskProcessPolicy", func(_ *HotResetTools, _ string) (string, int, error) {
			return "reset", common.ResetErrorLevel, nil
		})
}

// TestIsNeedBlockAllDevice ut for method isNeedBlockAllDevice,using new board id
func TestIsNeedBlockAllDevice(t *testing.T) {

	patch := mockGetServerUsageLabelCache()
	defer patch.Reset()
	GetServerBoardIdPatch := mockGetServerBoardId(A800IA2WithHccs)
	defer GetServerBoardIdPatch.Reset()

	doTestIsNeedBlockAllDevice(t)
}

// TestIsNeedBlockAllDevice ut for method isNeedBlockAllDevice,using old board id
func TestIsNeedBlockAllDeviceUsOldBoardId(t *testing.T) {

	patch := mockGetServerUsageLabelCache()
	defer patch.Reset()
	GetServerBoardIdPatch := mockGetServerBoardId(A800IA2WithHccsOld)
	defer GetServerBoardIdPatch.Reset()

	doTestIsNeedBlockAllDevice(t)
}

func doTestIsNeedBlockAllDevice(t *testing.T) {
	convey.Convey("test need block all device", t, func() {
		convey.Convey("test need to block devices", func() {

			hnm := NewHwAscend910Manager()
			hnm.SetKubeClient(&kubeclient.ClientK8s{
				Clientset:      &kubernetes.Clientset{},
				NodeName:       "node",
				DeviceInfoName: common.DeviceInfoCMNamePrefix + "node",
				IsApiErr:       false,
			})
			faultDevice := make([]common.DeviceFault, 0)
			block := hnm.isNeedBlockAllDevice(faultDevice)
			convey.So(block, convey.ShouldBeFalse)
			faultDevice = append(faultDevice, common.DeviceFault{
				FaultType:            "",
				NPUName:              "",
				LargeModelFaultLevel: "",
				FaultLevel:           common.RestartRequest,
				FaultHandling:        "",
				FaultCode:            "",
			})
			block = hnm.isNeedBlockAllDevice(faultDevice)
			convey.So(block, convey.ShouldBeTrue)
		})
	})
}

// TestNoNeedToBlock test need block all device with none hccs A800IA2
func TestNoNeedToBlock(t *testing.T) {
	patch := mockGetServerUsageLabelCache()
	defer patch.Reset()
	GetServerBoardIdPatch := mockGetServerBoardId(common.A800IA2NoneHccsBoardId)
	defer GetServerBoardIdPatch.Reset()

	doTestNoNeedToBlock(t)
}

// TestNoNeedToBlock test need block all device with none hccs A800IA2,using old board id
func TestNoNeedToBlockUsingOldId(t *testing.T) {
	patch := mockGetServerUsageLabelCache()
	defer patch.Reset()
	GetServerBoardIdPatch := mockGetServerBoardId(common.A800IA2NoneHccsBoardIdOld)
	defer GetServerBoardIdPatch.Reset()

	doTestNoNeedToBlock(t)
}
func doTestNoNeedToBlock(t *testing.T) {
	convey.Convey("test no need to block devices", t, func() {
		hnm := NewHwAscend910Manager()
		hnm.SetKubeClient(&kubeclient.ClientK8s{
			Clientset:      &kubernetes.Clientset{},
			NodeName:       "node",
			DeviceInfoName: common.DeviceInfoCMNamePrefix + "node",
			IsApiErr:       false,
		})
		faultDevice := make([]common.DeviceFault, 0)
		faultDevice = append(faultDevice, common.DeviceFault{
			FaultType:            "",
			NPUName:              "",
			LargeModelFaultLevel: "",
			FaultLevel:           common.NotHandleFault,
			FaultHandling:        "",
			FaultCode:            "",
		})
		block := hnm.isNeedBlockAllDevice(faultDevice)
		// it is none hccs A800IA2, will not block all devices
		convey.So(block, convey.ShouldBeFalse)
	})
}

func TestNodeNeedToBlockWithNotHandleErr(t *testing.T) {
	patch := mockGetServerUsageLabelCache()
	defer patch.Reset()
	GetServerBoardIdPatch := mockGetServerBoardId(A800IA2WithHccs)
	defer GetServerBoardIdPatch.Reset()

	doTestNodeNeedToBlockWithNotHandleErr(t)
}

func TestNodeNeedToBlockWithNotHandleErrUsingOldId(t *testing.T) {
	patch := mockGetServerUsageLabelCache()
	defer patch.Reset()
	GetServerBoardIdPatch := mockGetServerBoardId(A800IA2WithHccsOld)
	defer GetServerBoardIdPatch.Reset()

	doTestNodeNeedToBlockWithNotHandleErr(t)
}
func doTestNodeNeedToBlockWithNotHandleErr(t *testing.T) {
	convey.Convey("test no need to block devices", t, func() {

		hnm := NewHwAscend910Manager()
		hnm.SetKubeClient(&kubeclient.ClientK8s{
			Clientset:      &kubernetes.Clientset{},
			NodeName:       "node",
			DeviceInfoName: common.DeviceInfoCMNamePrefix + "node",
			IsApiErr:       false,
		})
		faultDevice := make([]common.DeviceFault, 0)
		faultDevice = append(faultDevice, common.DeviceFault{
			FaultType:            "",
			NPUName:              "",
			LargeModelFaultLevel: "",
			FaultLevel:           common.NotHandleFault,
			FaultHandling:        "",
			FaultCode:            "",
		})
		block := hnm.isNeedBlockAllDevice(faultDevice)
		// it is none hccs A800IA2, will not block all devices
		convey.So(block, convey.ShouldBeFalse)
	})
}

func mockGetServerBoardId(devLogicID int) *gomonkey.Patches {
	return gomonkey.ApplyMethodReturn(&AscendTools{}, "GetServerBoardId", uint32(devLogicID), nil)
}

func mockGetServerUsageLabelCache() *gomonkey.Patches {
	return gomonkey.
		ApplyMethodReturn(&kubeclient.ClientK8s{}, "GetServerUsageLabelCache",
			common.Infer, nil)
}

func mockUpdateResetCMStatus(ret error) *gomonkey.Patches {
	return gomonkey.ApplyPrivateMethod(&HwAscend910Manager{}, "updateResetCMStatus",
		func(*HwAscend910Manager, string, string, string, string, []*common.TaskDevInfo) error { return ret })
}

func mockWaitForAllFaultyDeviceProcessesToZero(ret error) *gomonkey.Patches {
	return gomonkey.ApplyPrivateMethod(&HwAscend910Manager{}, "waitForAllFaultyDeviceProcessesToZero",
		func(*HwAscend910Manager, string, []*common.TaskDevInfo) error { return ret })
}

func mockResetDeviceOnce(ret error) *gomonkey.Patches {
	return gomonkey.ApplyPrivateMethod(&HwAscend910Manager{}, "resetDeviceOnce",
		func(*HwAscend910Manager, []*common.TaskDevInfo, map[string][]*common.NpuDevice) error { return ret })
}

func mockUpgradeResetProcess(ret error) *gomonkey.Patches {
	return gomonkey.ApplyPrivateMethod(&HwAscend910Manager{}, "upgradeResetProcess",
		func(*HwAscend910Manager, string, []*common.TaskDevInfo) error { return ret })
}

func mockUpgradeRestartProcess(policy string, err error) *gomonkey.Patches {
	return gomonkey.ApplyPrivateMethod(&HwAscend910Manager{}, "upgradeRestartProcess",
		func(*HwAscend910Manager, string, []*common.TaskDevInfo, map[string][]*common.NpuDevice) (string, error) {
			return policy, err
		})
}

func mockRefreshDevFaultInfo(ret error) *gomonkey.Patches {
	return gomonkey.ApplyPrivateMethod(&HwAscend910Manager{}, "refreshDevFaultInfo",
		func(*HwAscend910Manager, []*common.TaskDevInfo, map[string][]*common.NpuDevice) error { return ret })
}

func mockCheckDevErrorCode(policy string, needUpgrade bool, err error) *gomonkey.Patches {
	return gomonkey.ApplyPrivateMethod(&HwAscend910Manager{}, "checkDevErrorCode",
		func(*HwAscend910Manager, string, []*common.TaskDevInfo, map[string][]*common.NpuDevice) (string, bool, error) {
			return policy, needUpgrade, err
		})
}

// TestConvertPhysicIdToLogicId for test convertPhysicIdToLogicId
func TestConvertPhysicIdToLogicId(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test convertPhysicIdToLogicId", t, func() {
		convey.Convey("01-empty physic id list, should return error", func() {
			logicIds, err := manager.convertPhysicIdToLogicId(nil)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(logicIds, convey.ShouldBeNil)
		})
		patch := gomonkey.ApplyMethod(&devmanager.DeviceManagerMock{}, "GetLogicIDFromPhysicID",
			func(dm *devmanager.DeviceManagerMock, physicID int32) (int32, error) {
				if physicID == chipPhyID0 {
					return -1, errors.New("test error")
				}
				return physicID, nil
			})
		defer patch.Reset()
		convey.Convey("02-get logic id from physic id failed, should return error", func() {
			phyIds := []int32{chipPhyID0, chipPhyID1, chipPhyID2}
			phyIds, err := manager.convertPhysicIdToLogicId(phyIds)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(phyIds, convey.ShouldBeNil)
		})
		convey.Convey("03-covert success, should return nil", func() {
			phyIds := []int32{chipPhyID1, chipPhyID2}
			phyIds, err := manager.convertPhysicIdToLogicId(phyIds)
			convey.So(err, convey.ShouldBeNil)
			convey.So(phyIds, convey.ShouldResemble, phyIds)
		})
	})
}

// TestConvertLogicIdToPhysicId for test convertLogicIdToPhysicId
func TestConvertLogicIdToPhysicId(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test convertLogicIdToPhysicId", t, func() {
		convey.Convey("01-empty logicId list, should return error", func() {
			phyIds, err := manager.convertLogicIdToPhysicId(nil)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(phyIds, convey.ShouldBeNil)
		})
		patch := gomonkey.ApplyMethod(&devmanager.DeviceManagerMock{}, "GetPhysicIDFromLogicID",
			func(dm *devmanager.DeviceManagerMock, logicID int32) (int32, error) {
				if logicID == chipPhyID0 {
					return -1, errors.New("test error")
				}
				return logicID, nil
			})
		defer patch.Reset()
		convey.Convey("02-get phy id from logic id failed, should return error", func() {
			logicIds := []int32{chipPhyID0, chipPhyID1, chipPhyID2}
			phyIds, err := manager.convertLogicIdToPhysicId(logicIds)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(phyIds, convey.ShouldBeNil)
		})
		convey.Convey("03-covert success, should return nil", func() {
			logicIds := []int32{chipPhyID1, chipPhyID2}
			phyIds, err := manager.convertLogicIdToPhysicId(logicIds)
			convey.So(err, convey.ShouldBeNil)
			convey.So(phyIds, convey.ShouldResemble, logicIds)
		})
	})
}

// TestIsReSchedulingScene for test IsReSchedulingScene
func TestIsReSchedulingScene(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test isReSchedulingScene", t, func() {
		npuCount := 2
		convey.Convey("01-get reset dev num once failed, should return false", func() {
			manager.hotResetManager = &HotResetTools{}
			ret := manager.isReSchedulingScene(npuCount)
			convey.So(ret, convey.ShouldBeFalse)
		})
		manager.hotResetManager = &HotResetTools{
			resetDevNumOnce: common.Ascend910RingsNum,
		}
		convey.Convey("02-device usage is not train, should return false", func() {
			ret := manager.isReSchedulingScene(npuCount)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("03-is rescheduling scene, should return true", func() {
			manager.AscendTools.deviceUsage = common.Train
			ret := manager.isReSchedulingScene(npuCount)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

// TestIsTaskInReset for test isTaskInReset
func TestIsTaskInReset(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test isTaskInReset", t, func() {
		manager.hotResetManager = &HotResetTools{
			resetDevNumOnce: common.Ascend910RingsNum,
			taskPod:         map[string]v1.Pod{"task2": {}, "task3": {}},
			resetTask:       map[string]struct{}{"task2": {}},
		}
		convey.Convey("01-get task pod failed, should return false and error", func() {
			taskName := "task1"
			inReset, err := manager.isTaskInReset(taskName)
			convey.So(inReset, convey.ShouldBeFalse)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-cur node task in reset, should return true and nil", func() {
			taskName := "task2"
			inReset, err := manager.isTaskInReset(taskName)
			convey.So(inReset, convey.ShouldBeTrue)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("03-get cm from cache failed, should return false and error", func() {
			patch := mockGetCMFromCache(nil, errors.New("get cm from cache failed"))
			defer patch.Reset()
			taskName := "task3"
			inReset, err := manager.isTaskInReset(taskName)
			convey.So(inReset, convey.ShouldBeFalse)
			convey.So(err.Error(), convey.ShouldEqual, "get cm from cache failed")
		})
		convey.Convey("04-need tolerance, should return false and nil", func() {
			taskDevList := mockTaskDevInfoList()
			taskDevList[0].Status = common.RecoveredStatus
			patch := mockGetCMFromCache(taskDevList, nil)
			defer patch.Reset()
			taskName := "task3"
			inReset, err := manager.isTaskInReset(taskName)
			convey.So(inReset, convey.ShouldBeFalse)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("05-success, should return true and nil", func() {
			taskDevList := mockTaskDevInfoList()
			patch := mockGetCMFromCache(taskDevList, nil)
			defer patch.Reset()
			taskName := "task3"
			inReset, err := manager.isTaskInReset(taskName)
			convey.So(inReset, convey.ShouldBeTrue)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestPolicyLevelHandle for test policyLevelHandle
func TestPolicyLevelHandle(t *testing.T) {
	taskName := "mockTaskName"
	convey.Convey("test policyLevelHandle", t, func() {
		convey.Convey("01-policy level is 2,should return false", func() {
			handle := policyLevelHandle("", taskName, common.RestartRequestErrorLevel)
			convey.So(handle, convey.ShouldBeFalse)
		})
		convey.Convey("02-policy level is 5,should return false", func() {
			mockVar := gomonkey.ApplyGlobalVar(&isHotResetOn, false)
			defer mockVar.Reset()
			handle := policyLevelHandle("", taskName, common.ResetErrorLevel)
			convey.So(handle, convey.ShouldBeFalse)
		})
		convey.Convey("03-other policy level, should return true", func() {
			handle := policyLevelHandle("", taskName, common.FreeResetErrorLevel)
			convey.So(handle, convey.ShouldBeTrue)
			handle = policyLevelHandle("", taskName, common.IsolateErrorLevel)
			convey.So(handle, convey.ShouldBeTrue)
		})
	})
}

// TestIsolateSceneHandle for test isolateSceneHandle
func TestIsolateSceneHandle(t *testing.T) {
	manager := createFake910Manager()
	taskName := "mockTaskName"
	convey.Convey("test isolateSceneHandle", t, func() {
		convey.Convey("01-task is in reset, should return true", func() {
			manager.hotResetManager = &HotResetTools{
				taskPod:        map[string]v1.Pod{taskName: {}},
				allTaskDevList: map[string][]int32{taskName: {}},
				faultDev2PodMap: map[int32]v1.Pod{
					1: getSinglePod("pod1", map[string]string{common.ResetTaskNameKey: taskName})},
			}
			mockFunc := mockGetCMFromCache(mockTaskDevInfoList(), nil)
			mockFunc.ApplyPrivateMethod(&HwAscend910Manager{}, "tryWriteIsolationInfo",
				func(*HwAscend910Manager, string) {})
			defer mockFunc.Reset()
			isolate := manager.isolateSceneHandle(taskName)
			convey.So(isolate, convey.ShouldBeTrue)
		})
	})
}

// TestRunProcessTask for test runProcessTask
func TestRunProcessTask(t *testing.T) {
	taskName := "mockTaskName"
	mockFunc := gomonkey.ApplyPrivateMethod(&HwAscend910Manager{}, "restartRequestProcess",
		func(*HwAscend910Manager, string, *common.TaskResetInfo, map[string][]*common.NpuDevice) { return }).
		ApplyPrivateMethod(&HwAscend910Manager{}, "restartProcess",
			func(*HwAscend910Manager, string, *common.TaskResetInfo, map[string][]*common.NpuDevice) { return }).
		ApplyPrivateMethod(&HwAscend910Manager{}, "resetProcess",
			func(*HwAscend910Manager, string, *common.TaskResetInfo, map[string][]*common.NpuDevice) { return })
	defer mockFunc.Reset()
	convey.Convey("test runProcessTask", t, func() {
		convey.Convey("01-policy level is 2, call restartRequestProcess, should return nil", func() {
			manager := createFake910Manager()
			err := manager.runProcessTask(taskName, common.RestartRequestErrorLevel, nil, nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-policy level is 3, call restartProcess, should return nil", func() {
			manager := createFake910Manager()
			err := manager.runProcessTask(taskName, common.RestartErrorLevel, nil, nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("03-policy level is 5, call resetProcess, should return nil", func() {
			manager := createFake910Manager()
			err := manager.runProcessTask(taskName, common.ResetErrorLevel, nil, nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("04-policy level is 6, call resetProcess, should return nil", func() {
			manager := createFake910Manager()
			err := manager.runProcessTask(taskName, common.IsolateErrorLevel, nil, nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestRestartRequestProcess for test restartRequestProcess
func TestRestartRequestProcess(t *testing.T) {
	manager := createFake910Manager()
	taskName := "mockTaskName"
	resetInfo := &common.TaskResetInfo{RankList: mockTaskDevInfoList()}
	convey.Convey("test restartRequestProcess", t, func() {
		mockFunc := mockUpdateResetCMStatus(errors.New("mock error")).
			ApplyFuncReturn(time.Sleep).ApplyFunc(common.SetDeviceInit, func(int32) { return })
		defer mockFunc.Reset()
		convey.Convey("01-get task dev fault info list failed", func() {
			manager.hotResetManager = &HotResetTools{}
			manager.restartRequestProcess(taskName, resetInfo, nil)
		})
		convey.Convey("get task fault dev fault info list  success", func() {
			manager.hotResetManager = &HotResetTools{
				resetDevNumOnce:     1,
				allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{taskName: getTaskInfo1()},
				resetDev:            map[int32]struct{}{},
			}
			convey.Convey("02-fault dev list is empty", func() {
				mockFunc1 := mockCheckDevErrorCode("", false, nil)
				defer mockFunc1.Reset()
				manager.restartRequestProcess(taskName, resetInfo, nil)
			})
			convey.Convey("03-fault dev list do not upgrade", func() {
				mockFunc1 := mockCheckDevErrorCode("", false, errors.New("mock error"))
				defer mockFunc1.Reset()
				manager.restartRequestProcess(taskName, resetInfo, nil)
			})
			convey.Convey("04-upgrade restart request process failed ", func() {
				mockFunc1 := mockCheckDevErrorCode("", true, errors.New("mock error"))
				defer mockFunc1.Reset()
				manager.restartRequestProcess(taskName, resetInfo, nil)
			})
		})
	})
}

// TestHandleSucceedRestartRequest for test handleSucceedRestartRequest
func TestHandleSucceedRestartRequest(t *testing.T) {
	manager := createFake910Manager()
	taskName := "mockTask1"
	convey.Convey("test handleSucceedRestartRequest", t, func() {
		mockSleep := gomonkey.ApplyFuncReturn(time.Sleep).ApplyFunc(common.SetDeviceInit, func(int32) { return })
		defer mockSleep.Reset()
		convey.Convey("01-update reset cm status failed, should not set resetTask", func() {
			manager.hotResetManager = &HotResetTools{
				resetTask:       map[string]struct{}{taskName: {}},
				resetDevNumOnce: 1,
			}
			manager.handleSucceedRestartRequest(taskName, common.ResetError, nil, nil)
			convey.So(manager.hotResetManager.IsCurNodeTaskInReset(taskName), convey.ShouldBeTrue)
		})
		manager.hotResetManager = &HotResetTools{
			taskPod:         map[string]v1.Pod{taskName: {}},
			resetTask:       map[string]struct{}{taskName: {}},
			resetDevNumOnce: 1,
		}
		mockFunc1 := gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{},
			"WriteResetInfoDataIntoCM", nil, nil)
		defer mockFunc1.Reset()
		convey.Convey("02-unset task in reset failed, ", func() {
			manager.handleSucceedRestartRequest(taskName, common.ResetError, nil, nil)
			convey.So(manager.hotResetManager.IsCurNodeTaskInReset(taskName), convey.ShouldBeFalse)
		})
	})
}

// TestCheckDevErrorCode for test checkDevErrorCode
func TestCheckDevErrorCode(t *testing.T) {
	manager := createFake910Manager()
	manager.hotResetManager = &HotResetTools{}
	taskName := "mockTaskName"
	convey.Convey("test checkDevErrorCode", t, func() {
		convey.Convey("01-refresh dev fault info failed, should return error", func() {
			mockFunc := mockRefreshDevFaultInfo(errors.New("mock refresh dev fault info error"))
			defer mockFunc.Reset()
			_, needUpgrade, resetErr := manager.checkDevErrorCode(taskName, getTaskInfo1(), mockGroupDevice())
			convey.So(needUpgrade, convey.ShouldBeFalse)
			convey.So(resetErr, convey.ShouldNotBeNil)
		})
		mockFunc := mockRefreshDevFaultInfo(nil)
		defer mockFunc.Reset()
		convey.Convey("02-get dev list by policy level failed, should return error", func() {
			_, needUpgrade, resetErr := manager.checkDevErrorCode(taskName, getTaskInfo(), mockGroupDevice())
			convey.So(needUpgrade, convey.ShouldBeFalse)
			convey.So(resetErr, convey.ShouldNotBeNil)
		})
		convey.Convey("03-get empty dev list by policy level, should return nil", func() {
			policy, needUpgrade, resetErr := manager.checkDevErrorCode(taskName, nil, mockGroupDevice())
			convey.So(policy == common.RestartRequestError, convey.ShouldBeTrue)
			convey.So(needUpgrade, convey.ShouldBeFalse)
			convey.So(resetErr, convey.ShouldBeNil)
		})
	})
}

// TestRestartProcess for test restartProcess
func TestRestartProcess(t *testing.T) {
	manager := createFake910Manager()
	taskName := "mockTaskName"
	resetInfo := &common.TaskResetInfo{RankList: mockTaskDevInfoList()}
	convey.Convey("test restartProcess", t, func() {
		mockFunc := mockUpdateResetCMStatus(errors.New("mock error")).ApplyFuncReturn(time.Sleep)
		defer mockFunc.Reset()
		convey.Convey("01-get task dev fault info list failed", func() {
			manager.hotResetManager = &HotResetTools{}
			manager.restartProcess(taskName, resetInfo, nil)
		})
		manager.hotResetManager = &HotResetTools{
			resetDevNumOnce:     1,
			allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{taskName: getTaskInfo1()},
			resetDev:            map[int32]struct{}{},
		}
		convey.Convey("02-wait for all fault device processes to zero failed", func() {
			mockFunc1 := mockWaitForAllFaultyDeviceProcessesToZero(errors.New("mock error"))
			defer mockFunc1.Reset()
			manager.restartProcess(taskName, resetInfo, nil)
		})
		mockFunc1 := mockWaitForAllFaultyDeviceProcessesToZero(nil)
		defer mockFunc1.Reset()
		convey.Convey("03-reset device once failed", func() {
			mockFunc2 := mockRefreshDevFaultInfo(errors.New("mock error"))
			defer mockFunc2.Reset()
			manager.restartProcess(taskName, resetInfo, nil)
		})
		mockFunc2 := mockRefreshDevFaultInfo(nil)
		defer mockFunc2.Reset()
		convey.Convey("04-upgrade restart process failed", func() {
			mockFunc3 := mockUpgradeRestartProcess(common.ResetError, errors.New("mock error"))
			defer mockFunc3.Reset()
			manager.restartProcess(taskName, resetInfo, nil)
		})
		mockFunc3 := mockUpgradeRestartProcess(common.ResetError, nil)
		defer mockFunc3.Reset()
		convey.Convey("05-set reset cm status failed", func() {
			manager.restartProcess(taskName, resetInfo, nil)
		})
		convey.Convey("06-unset task in reset failed", func() {
			mockFunc4 := mockUpdateResetCMStatus(nil)
			defer mockFunc4.Reset()
			manager.restartProcess(taskName, resetInfo, nil)
		})
	})
}

// TestUpgradeRestartProcess for test upgradeRestartProcess
func TestUpgradeRestartProcess(t *testing.T) {
	manager := createFake910Manager()
	manager.hotResetManager = &HotResetTools{}
	taskName := "mockTask"
	convey.Convey("test upgradeRestartProcess", t, func() {
		convey.Convey("01-get devList by policy level failed, should return error", func() {
			_, err := manager.upgradeRestartProcess(taskName, getTaskInfo(), mockGroupDevice())
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-get empty devList by policy level, should return nil", func() {
			code, err := manager.upgradeRestartProcess(taskName, nil, mockGroupDevice())
			convey.So(err, convey.ShouldBeNil)
			convey.So(code == common.RestartError, convey.ShouldBeTrue)
		})
		convey.Convey("03-update reset cm status without wait failed, should return error", func() {
			_, err := manager.upgradeRestartProcess(taskName, getTaskInfo1(), mockGroupDevice())
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("reset device once and write reset info into cm success", func() {
			manager.hotResetManager = &HotResetTools{
				resetDevNumOnce: 1,
				taskPod:         map[string]v1.Pod{taskName: {}},
			}
			mockFunc := mockResetDeviceOnce(nil).
				ApplyMethodReturn(&kubeclient.ClientK8s{}, "WriteResetInfoDataIntoCM", nil, nil)
			defer mockFunc.Reset()
			convey.Convey("04-update reset cm status failed, should return error", func() {
				mockFunc2 := mockUpdateResetCMStatus(errors.New("mock errors"))
				defer mockFunc2.Reset()
				_, err := manager.upgradeRestartProcess(taskName, getTaskInfo1(), mockGroupDevice())
				convey.So(err, convey.ShouldNotBeNil)
			})
			convey.Convey("05-update reset cm status success, should return error", func() {
				mockFunc2 := mockUpdateResetCMStatus(nil)
				defer mockFunc2.Reset()
				_, err := manager.upgradeRestartProcess(taskName, getTaskInfo1(), mockGroupDevice())
				convey.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestUpgradeRestartRequestProcess for test upgradeRestartRequestProcess
func TestUpgradeRestartRequestProcess(t *testing.T) {
	manager := createFake910Manager()
	taskName := "mockTask"
	convey.Convey("upgradeRestartRequestProcess", t, func() {
		mockFunc1 := gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "WriteResetInfoDataIntoCM",
			nil, nil)
		defer mockFunc1.Reset()
		convey.Convey("01-update reset cm status failed, should return error", func() {
			manager.hotResetManager = &HotResetTools{}
			_, err := manager.upgradeRestartRequestProcess(taskName, getTaskInfo1(), mockGroupDevice())
			convey.So(err, convey.ShouldNotBeNil)
		})
		manager.hotResetManager = &HotResetTools{
			resetDevNumOnce: 2,
			taskPod:         map[string]v1.Pod{taskName: {}},
		}
		convey.Convey("02-invalid policy str, should return error", func() {
			_, err := manager.upgradeRestartRequestProcess(taskName, getTaskInfo(), mockGroupDevice())
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-reset device once failed, should return error", func() {
			mockFunc := mockResetDeviceOnce(errors.New("mock errors"))
			defer mockFunc.Reset()
			_, err := manager.upgradeRestartRequestProcess(taskName, getTaskInfo1(), mockGroupDevice())
			convey.So(err, convey.ShouldNotBeNil)
		})
		mockFunc := mockResetDeviceOnce(nil)
		defer mockFunc.Reset()
		convey.Convey("04-get empty devList by policy level, should return nil", func() {
			code, err := manager.upgradeRestartRequestProcess(taskName, nil, mockGroupDevice())
			convey.So(code == common.ResetError, convey.ShouldBeTrue)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("05-update reset cm status, should return error", func() {
			mockFunc2 := mockUpdateResetCMStatus(errors.New("mock errors"))
			defer mockFunc2.Reset()
			convey.Convey("", func() {
				_, err := manager.upgradeRestartRequestProcess(taskName, getTaskInfo1(), mockGroupDevice())
				convey.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestUpdateResetCMStatus(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test updateResetCMStatus", t, func() {
		convey.Convey("01-task is not in reset, should return error", func() {
			manager.hotResetManager = &HotResetTools{}
			taskName := "taskNameNotInTaskPod"
			err := manager.updateResetCMStatus(taskName, common.ResetError,
				common.ResetError, common.RecoveredStatus, nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		mockSleep := gomonkey.ApplyFuncReturn(time.Sleep)
		defer mockSleep.Reset()
		convey.Convey("02-status is recovered and policy is reset, should return error", func() {
			taskName := "mockTask1"
			manager.hotResetManager = &HotResetTools{
				taskPod:   map[string]v1.Pod{taskName: {}},
				resetTask: map[string]struct{}{taskName: {}},
			}
			err := manager.updateResetCMStatus(taskName, common.ResetError,
				common.ResetError, common.RecoveredStatus, nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("target task has data in hotResetManager", func() {
			taskName := "mockTask1"
			manager.hotResetManager = &HotResetTools{
				taskPod:         map[string]v1.Pod{taskName: {}},
				resetTask:       map[string]struct{}{taskName: {}},
				resetDevNumOnce: 1,
			}
			convey.Convey("03-write reset info data into cm failed, should return error", func() {
				mockFunc1 := gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "WriteResetInfoDataIntoCM",
					nil, errors.New("mock error"))
				defer mockFunc1.Reset()
				err := manager.updateResetCMStatus(taskName, common.ResetError,
					common.ResetError, common.RecoveredStatus, nil)
				convey.So(err, convey.ShouldNotBeNil)
			})
			convey.Convey("04-update reset cm status success, should return error", func() {
				mockFunc1 := gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "WriteResetInfoDataIntoCM",
					nil, nil)
				defer mockFunc1.Reset()
				err := manager.updateResetCMStatus(taskName, common.ResetError,
					common.ResetError, common.RecoveredStatus, nil)
				convey.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestUpdateResetCMStatusWithoutWait for test updateResetCMStatusWithoutWait
func TestUpdateResetCMStatusWithoutWait(t *testing.T) {
	manager := createFake910Manager()
	taskName := "mockTaskName"
	convey.Convey("test update resetCMStatus", t, func() {
		convey.Convey("01-get task reset info failed, should return error", func() {
			manager.hotResetManager = &HotResetTools{}
			err := manager.updateResetCMStatusWithoutWait(taskName, common.ResetError,
				common.ResetError, common.UnrecoveredStatus, nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-get task pod failed, should return error", func() {
			manager.hotResetManager = &HotResetTools{
				resetDevNumOnce: 1,
				taskPod:         map[string]v1.Pod{},
			}
			err := manager.updateResetCMStatusWithoutWait(taskName, common.ResetError,
				common.ResetError, common.UnrecoveredStatus, nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-write reset info data info failed, should return error", func() {
			manager.hotResetManager = &HotResetTools{
				resetDevNumOnce: 1,
				taskPod:         map[string]v1.Pod{taskName: {}},
			}
			mockFunc := gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "WriteResetInfoDataIntoCM",
				nil, errors.New("mock error"))
			defer mockFunc.Reset()
			err := manager.updateResetCMStatusWithoutWait(taskName, common.ResetError,
				common.ResetError, common.UnrecoveredStatus, nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})

}

// TestResetProcess for test resetProcess
func TestResetProcess(t *testing.T) {
	manager := createFake910Manager()
	taskName := "mockTaskName"
	manager.hotResetManager = &HotResetTools{
		resetDevNumOnce:     1,
		allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{taskName: getTaskInfo1()},
		resetDev:            map[int32]struct{}{},
	}
	resetInfo := &common.TaskResetInfo{RankList: mockTaskDevInfoList()}
	convey.Convey("test resetProcess", t, func() {
		mockFunc := mockUpdateResetCMStatus(errors.New("mock error"))
		defer mockFunc.Reset()
		convey.Convey("01-wait for all fault device processes to zero failed", func() {
			mockFunc1 := mockWaitForAllFaultyDeviceProcessesToZero(errors.New("mock error"))
			defer mockFunc1.Reset()
			manager.resetProcess(taskName, resetInfo, nil)
		})
		mockFunc1 := mockWaitForAllFaultyDeviceProcessesToZero(nil)
		defer mockFunc1.Reset()
		convey.Convey("02-reset device once failed", func() {
			mockFunc2 := mockResetDeviceOnce(errors.New("mock error"))
			defer mockFunc2.Reset()
			manager.resetProcess(taskName, resetInfo, nil)
		})
		mockFunc2 := mockResetDeviceOnce(nil)
		defer mockFunc2.Reset()
		convey.Convey("03-upgrade reset process failed", func() {
			mockFunc3 := mockUpgradeResetProcess(errors.New("mock error"))
			defer mockFunc3.Reset()
			manager.resetProcess(taskName, resetInfo, nil)
		})
		mockFunc3 := mockUpgradeResetProcess(nil)
		defer mockFunc3.Reset()
		convey.Convey("04-set reset cm status failed", func() {
			mockFunc4 := mockUpdateResetCMStatus(errors.New("mock error"))
			defer mockFunc4.Reset()
			manager.resetProcess(taskName, resetInfo, nil)
		})
		convey.Convey("05-unset task in reset failed", func() {
			mockFunc4 := mockUpdateResetCMStatus(nil)
			defer mockFunc4.Reset()
			manager.resetProcess(taskName, resetInfo, nil)
		})
	})
}

// TestCanContinueGraceProcess for test canContinueGraceProcess
func TestCanContinueGraceProcess(t *testing.T) {
	manager := createFake910Manager()
	taskName := "mockTaskName"
	convey.Convey("test canContinueGraceProcess", t, func() {
		faultDeviceLogicIdMap := map[int32]int32{chipPhyID1: 1, chipPhyID2: 0}
		convey.Convey("01-number of all process is zero, should return true", func() {
			ret := manager.canContinueGraceProcess(faultDeviceLogicIdMap, taskName, true)
			convey.So(ret, convey.ShouldBeTrue)
		})
		convey.Convey("get device process info fail", func() {
			manager.SetDmgr(&devmanager.DeviceManagerMockErr{})
			convey.Convey("02-isLastQuery is false, should return false", func() {
				ret := manager.canContinueGraceProcess(faultDeviceLogicIdMap, taskName, false)
				convey.So(ret, convey.ShouldBeFalse)
			})
			convey.Convey("03-isLastQuery is true, should return true", func() {
				ret := manager.canContinueGraceProcess(faultDeviceLogicIdMap, taskName, true)
				convey.So(ret, convey.ShouldBeTrue)
			})
		})
	})
}

// TestUpgradeResetProcess for test upgradeResetProcess
func TestUpgradeResetProcess(t *testing.T) {
	manager := createFake910Manager()
	manager.hotResetManager = &HotResetTools{resetDevNumOnce: 1}
	taskName := "mockTask1"
	convey.Convey("test upgradeResetProcess", t, func() {
		convey.Convey("01-get need reset devMap fail, should return err", func() {
			err := manager.upgradeResetProcess(taskName, getTaskInfo())
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-task dev info is empty, should return nil", func() {
			err := manager.upgradeResetProcess(taskName, nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("04-update reset cm status fail, should return err", func() {
			mockFunc := mockUpdateResetCMStatus(errors.New("mock error"))
			defer mockFunc.Reset()
			err := manager.upgradeResetProcess(taskName, getTaskInfo1())
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("05-update reset cm status success, should return err", func() {
			mockFunc := mockUpdateResetCMStatus(nil)
			defer mockFunc.Reset()
			err := manager.upgradeResetProcess(taskName, getTaskInfo1())
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestPreProcess for test preProcess
func TestPreProcess(t *testing.T) {
	manager := createFake910Manager()
	manager.hotResetManager = &HotResetTools{
		resetDevNumOnce: 1,
		resetTask:       map[string]struct{}{},
		resetDev:        map[int32]struct{}{},
	}
	taskName := "mockTaskName"
	convey.Convey("01-get task dev fault info fail, should return err", t, func() {
		_, err := manager.preProcess(taskName, "")
		convey.So(err, convey.ShouldNotBeNil)
	})
	taskDevFaultInfo := map[string][]*common.TaskDevInfo{taskName: getTaskInfo1()}
	err := manager.hotResetManager.UpdateTaskDevFaultInfoCache(taskDevFaultInfo)
	convey.Convey("02-get task pod fail, should return err", t, func() {
		convey.So(err, convey.ShouldBeNil)
		_, err1 := manager.preProcess(taskName, "")
		convey.So(err1, convey.ShouldNotBeNil)
	})
	mockFunc := gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "WriteResetInfoDataIntoCM",
		&v1.ConfigMap{Data: map[string]string{}}, nil).
		ApplyMethodReturn(&kubeclient.ClientK8s{}, "WriteFaultInfoDataIntoCM",
			&v1.ConfigMap{Data: map[string]string{}}, nil)
	defer mockFunc.Reset()
	taskPod := map[string]v1.Pod{taskName: getSinglePod("pod1", map[string]string{})}
	err = manager.hotResetManager.UpdateTaskPodCache(taskPod)
	convey.Convey("03-pre process success, should return nil", t, func() {
		convey.So(err, convey.ShouldBeNil)
		_, err1 := manager.preProcess(taskName, common.ResetError)
		convey.So(err1, convey.ShouldBeNil)
		convey.So(manager.hotResetManager.IsCurNodeTaskInReset(taskName), convey.ShouldBeTrue)
		convey.So(len(manager.hotResetManager.GetDevListInReset()) > 0, convey.ShouldBeTrue)
		convey.So(manager.hotResetManager.GetDevListInReset(), convey.ShouldResemble,
			map[int32]struct{}{chipPhyID0: {}, chipPhyID1: {}, chipPhyID2: {}})
	})
}

// TestPostProcess for test postProcess
func TestPostProcess(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test postProcess", t, func() {

		nodeDeviceData := common.TaskResetInfo{
			RankList: mockTaskDevInfoList(),
		}
		convey.Convey("01-unset all dev in reset fail, should return error ", func() {
			manager.hotResetManager = &HotResetTools{
				resetDev: map[int32]struct{}{0: {}},
			}
			convey.So(manager.postProcess("fake-task", &nodeDeviceData), convey.ShouldNotBeNil)
		})
		convey.Convey("02-unset all dev in reset success, should return nil", func() {
			manager.hotResetManager = &HotResetTools{
				resetDev: map[int32]struct{}{0: {}, 1: {}},
			}
			convey.So(manager.postProcess("fake-task", &nodeDeviceData), convey.ShouldBeNil)
		})
	})
}

// TestRefreshDevFaultInfo for test refreshDevFaultInfo
func TestRefreshDevFaultInfo(t *testing.T) {
	manager := createFake910Manager()
	manager.hotResetManager = &HotResetTools{
		resetDevNumOnce: 1,
	}
	convey.Convey("test refreshDevFaultInfo", t, func() {
		convey.Convey("01-910 npu device not exist, should return error", func() {
			err := manager.refreshDevFaultInfo(getTaskInfo1(), nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestResetDeviceOnce for test resetDeviceOnce
func TestResetDeviceOnce(t *testing.T) {
	manager := createFake910Manager()
	manager.hotResetManager = &HotResetTools{
		resetDevNumOnce: 1,
	}
	convey.Convey("01-get need reset devMap fail because of invalid policy, should return err", t, func() {
		err := manager.resetDeviceOnce(getTaskInfo(), mockGroupDevice())
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("02-exec reset device fail, should return err", t, func() {
		mockFunc := gomonkey.ApplyPrivateMethod(&HwAscend910Manager{}, "execResetDevice",
			func(*HwAscend910Manager, map[int32]int32) error { return errors.New("reset device error") })
		defer mockFunc.Reset()
		err := manager.resetDeviceOnce(getTaskInfo1(), mockGroupDevice())
		convey.So(err, convey.ShouldNotBeNil)
	})
	mockFunc := gomonkey.ApplyPrivateMethod(&HwAscend910Manager{}, "execResetDevice",
		func(*HwAscend910Manager, map[int32]int32) error { return nil }).
		ApplyGlobalVar(&resetGoroutine, &sync.Map{}).
		ApplyFuncReturn(common.SetDeviceInit)
	defer mockFunc.Reset()

	convey.Convey("03-refresh dev Fault fail, should return err", t, func() {
		mockFunc1 := mockRefreshDevFaultInfo(errors.New("refresh dev fault error"))
		defer mockFunc1.Reset()
		err := manager.resetDeviceOnce(getTaskInfo1(), mockGroupDevice())
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("04-reset device success, should return nil", t, func() {
		mockFunc1 := mockRefreshDevFaultInfo(nil)
		defer mockFunc1.Reset()
		err := manager.resetDeviceOnce(getTaskInfo1(), mockGroupDevice())
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestExecResetDevice for test execResetDevice
func TestExecResetDevice(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test execResetDevice", t, func() {
		convey.Convey("01-hot reset success, should return nil", func() {
			manager.hotResetManager = &HotResetTools{
				faultDev2PodMap: map[int32]v1.Pod{},
				resetDevNumOnce: 1,
			}
			devMap := map[int32]int32{chipPhyID0: chipPhyID0}
			convey.So(manager.execResetDevice(devMap), convey.ShouldBeNil)
		})
	})
}

// TestIsNetResetCompleted for test isNetResetCompleted
func TestIsNetResetCompleted(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test isNetResetCompleted", t, func() {
		logicID := int32(2)
		convey.So(manager.isNetResetCompleted(logicID), convey.ShouldBeTrue)
		convey.Convey("01-get network health failed, should return false", func() {
			mockHealth := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDeviceNetWorkHealth",
				uint32(0), errors.New("get network health failed"))
			defer mockHealth.Reset()
			convey.So(manager.isNetResetCompleted(logicID), convey.ShouldBeFalse)
		})
		convey.Convey("02-network status is unhealthy, should return false", func() {
			mockHealth := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDeviceNetWorkHealth",
				uint32(1), nil)
			defer mockHealth.Reset()
			convey.So(manager.isNetResetCompleted(logicID), convey.ShouldBeFalse)
		})
	})
}

// TestWaitDeviceResetComplete for test waitDeviceResetComplete
func TestWaitDeviceResetComplete(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test WaitDeviceResetComplete", t, func() {
		convey.Convey("01-wait device reset recover timeout, should return error", func() {
			logicID := int32(2)
			totalTime := common.MaxResetWaitRecoverTime + 1
			err := manager.waitDeviceResetComplete(logicID, &totalTime, false)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-boot start device success, should return nil", func() {
			totalTime := 0
			logicID := int32(2)
			err := manager.waitDeviceResetComplete(logicID, &totalTime, false)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestIsRunningDistributed for test isRunningDistributed
func TestIsRunningDistributed(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test isRunningDistributed", t, func() {
		manager.hotResetManager = &HotResetTools{}
		// 01-faultDev2PodMap is nil, should return false
		logicId := int32(3)
		convey.So(manager.isRunningDistributed(logicId), convey.ShouldBeFalse)
		manager.hotResetManager = &HotResetTools{
			faultDev2PodMap: map[int32]v1.Pod{
				chipPhyID3: getSinglePod("pod1", map[string]string{
					common.DistributedJob: "true",
				}),
			},
		}
		// 02-cant find logic id in faultDev2PodMap, should return false
		logicId = int32(2)
		convey.So(manager.isRunningDistributed(logicId), convey.ShouldBeFalse)
		// 03-is distributed job, should return true
		logicId = int32(3)
		convey.So(manager.isRunningDistributed(logicId), convey.ShouldBeTrue)
	})
}

// TestTryWriteIsolationInfo for test tryWriteIsolationInfo
func TestTryWriteIsolationInfo(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test tryWriteIsolationInfo", t, func() {
		allTaskDevFaultInfo := []*common.TaskDevInfo{
			{
				DevFaultInfo: common.DevFaultInfo{
					LogicId: chipPhyID1,
					Policy:  common.IsolateError,
				},
			},
		}
		manager.hotResetManager = &HotResetTools{
			allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{
				"task1": allTaskDevFaultInfo,
			},
		}
		manager.tryWriteIsolationInfo("task2")
		manager.tryWriteIsolationInfo("task1")
	})
}

// TestIsDevShouldBeIsolate for test isDevShouldBeIsolate
func TestIsDevShouldBeIsolate(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test isDevShouldBeIsolate", t, func() {
		// 01-faultDev2PodMap is nil, should return false
		manager.hotResetManager = &HotResetTools{}
		convey.So(manager.isDevShouldBeIsolate(chipPhyID3), convey.ShouldBeFalse)
		manager.hotResetManager = &HotResetTools{
			faultDev2PodMap: map[int32]v1.Pod{
				chipPhyID3: getSinglePod("pod1", map[string]string{}),
			},
		}
		// 02-faultDev2PodMap is not nil, but target dev is not in map, should return false
		convey.So(manager.isDevShouldBeIsolate(chipPhyID4), convey.ShouldBeFalse)
		// 03-faultDev2PodMap is not nil and target dev is not in map, should return true
		convey.So(manager.isDevShouldBeIsolate(chipPhyID3), convey.ShouldBeTrue)
		manager.hotResetManager = &HotResetTools{
			faultDev2PodMap: map[int32]v1.Pod{
				chipPhyID3: getSinglePod("pod1", map[string]string{
					common.ResetTaskNameKey: "mock-task",
				}),
			},
			cmIndexer: cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{}),
		}
		// 04-get cm from cache failed, should return true
		convey.So(manager.isDevShouldBeIsolate(chipPhyID3), convey.ShouldBeTrue)
		// 05-stub GetCMFromCache,  should return false
		mockGetCMFromCacheMethod := mockGetCMFromCache(mockTaskDevInfoList(), nil)
		defer mockGetCMFromCacheMethod.Reset()
		convey.So(manager.isDevShouldBeIsolate(chipPhyID3), convey.ShouldBeFalse)
	})
}

func mockGetCMFromCache(taskDev []*common.TaskDevInfo, err error) *gomonkey.Patches {
	nodeDeviceData := common.TaskResetInfo{
		UpdateTime: 11111111,
		RankList:   taskDev,
	}
	cm := &v1.ConfigMap{
		Data: map[string]string{common.ResetInfoCMDataKey: string(common.MarshalData(nodeDeviceData))},
	}
	mockFunc := gomonkey.ApplyMethodReturn(&HotResetTools{}, "GetCMFromCache", cm, err)
	return mockFunc
}
