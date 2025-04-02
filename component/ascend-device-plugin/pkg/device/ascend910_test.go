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
	"ascend-common/api"
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

var ascend910testErr = errors.New("test")

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
		mockHandleResetProcess.ApplyPrivateMethod(manager, "canBeReset",
			func(dev *common.DevFaultInfo) (bool, error) {
				return true, nil
			})
		defer mockHandleResetProcess.Reset()
		mockHandleResetProcess.ApplyPrivateMethod(manager, "hotResetTryOutBand",
			func(_ *HwAscend910Manager, devs []*common.NpuDevice) {
				return
			})
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

// TestHotResetTryOutBand test the function hotResetTryOutBand
func TestHotResetTryOutBand(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test hotResetTryOutBand", t, func() {
		patch := gomonkey.ApplyPrivateMethod(manager, "updateResetInfo",
			func(failDevs, sucDevs []ResetDevice) {
				return
			})
		defer patch.Reset()
		devs := []*common.NpuDevice{
			{
				Health: v1beta1.Healthy,
			},
			{
				Health: v1beta1.Unhealthy,
			},
		}
		flag := false
		patch.ApplyPrivateMethod(manager, "execOutBandReset", func(devs, sucDevs []ResetDevice) error {
			flag = true
			return nil
		})
		convey.Convey("01-not A3, flag should be false", func() {
			common.ParamOption.RealCardType = common.Ascend910B
			manager.hotResetTryOutBand(devs)
			convey.So(flag, convey.ShouldBeFalse)
		})
		convey.Convey("02-A3, flag should be true", func() {
			common.ParamOption.RealCardType = common.Ascend910A3
			manager.hotResetTryOutBand(devs)
			convey.So(flag, convey.ShouldBeTrue)
		})
	})
}

// TestNpuDevToResetDev test the function npuDevToResetDev
func TestNpuDevToResetDev(t *testing.T) {
	dev := common.NpuDevice{
		CardID:   id1,
		DeviceID: id1,
		LogicID:  id1,
	}
	rest := ResetDevice{
		CardId:   id1,
		DeviceId: id1,
		LogicID:  id1,
	}
	convey.Convey("test npuDevToResetDev", t, func() {
		ret := npuDevToResetDev(dev)
		convey.So(ret, convey.ShouldResemble, rest)
	})
}

// TestCanBeReset an ut for function canBeReset
func TestCanBeReset(t *testing.T) {
	manager := createFake910Manager()
	manager.hotResetManager = newTestHotResetManager(common.Ascend910, common.Train)
	convey.Convey("exec ut function canBeReset", t, func() {
		convey.Convey("A3 device can reset, should return true", func() {
			common.ParamOption.RealCardType = common.Ascend910A3
			patch1 := gomonkey.ApplyPrivateMethod(manager, "canA3BeReset",
				func(dev *common.DevFaultInfo) bool {
					return true
				})
			patch1.ApplyPrivateMethod(manager, "canResetDeviceByLogicID", func(logicID int32) bool {
				return true
			})
			defer patch1.Reset()
			_, err := manager.canBeReset(mockSingleDevFaultInfo())
			convey.So(err, convey.ShouldBeNil)
		})
		common.ParamOption.RealCardType = common.Ascend910B
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

// TestCanA3BeReset test the function canA3BeReset
func TestCanA3BeReset(t *testing.T) {
	manager := createFake910Manager()
	dev := &common.DevFaultInfo{LogicId: int32(id1)}
	convey.Convey("test canA3BeReset", t, func() {
		convey.Convey("01-get cardID failed, should return false", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
				int32(id1), int32(id1), testErr)
			defer patch1.Reset()
			ret := manager.canA3BeReset(dev)
			convey.So(ret, convey.ShouldBeFalse)
		})
		patch := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
			int32(id1), int32(id1), nil)
		defer patch.Reset()
		convey.Convey("02-get associated card error, should return false", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "getAssociatedLogicIDs",
				func(logicID, cardID, deviceID int32) ([]int32, error) {
					return nil, testErr
				})
			defer patch1.Reset()
			ret := manager.canA3BeReset(dev)
			convey.So(ret, convey.ShouldBeFalse)
		})
		patch.ApplyPrivateMethod(manager, "getAssociatedLogicIDs",
			func(logicID, cardID, deviceID int32) ([]int32, error) {
				return []int32{id1}, nil
			})
		convey.Convey("03-get pod list failed, should return false", func() {
			patch1 := gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "GetAllPodList",
				nil, testErr)
			defer patch1.Reset()
			ret := manager.canA3BeReset(dev)
			convey.So(ret, convey.ShouldBeFalse)
		})
		patch.ApplyMethodReturn(&kubeclient.ClientK8s{}, "GetAllPodList", nil, nil)
		patch.ApplyPrivateMethod(manager, "getBusyChipListFromPod",
			func(podList *v1.PodList) []string {
				return []string{}
			})
		convey.Convey("04-get chip active error, should return false", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "isChipActive",
				func(logicID int32, busyChipList []string) (bool, error) {
					return false, testErr
				})
			defer patch1.Reset()
			ret := manager.canA3BeReset(dev)
			convey.So(ret, convey.ShouldBeFalse)
		})
	})
}

// TestCanA3BeResetPatch1 test the function canA3BeReset patch1
func TestCanA3BeResetPatch1(t *testing.T) {
	manager := createFake910Manager()
	dev := &common.DevFaultInfo{LogicId: int32(id1)}
	convey.Convey("test canA3BeReset patch1", t, func() {
		patch := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
			int32(id1), int32(id1), nil)
		defer patch.Reset()
		patch.ApplyPrivateMethod(manager, "getAssociatedLogicIDs",
			func(logicID, cardID, deviceID int32) ([]int32, error) {
				return []int32{id1}, nil
			})
		patch.ApplyMethodReturn(&kubeclient.ClientK8s{}, "GetAllPodList", nil, nil)
		patch.ApplyPrivateMethod(manager, "getBusyChipListFromPod",
			func(podList *v1.PodList) []string {
				return []string{}
			})
		convey.Convey("05-chip not active, should return false", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "isChipActive",
				func(logicID int32, busyChipList []string) (bool, error) {
					return false, nil
				})
			defer patch1.Reset()
			ret := manager.canA3BeReset(dev)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("06-success, should return true", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "isChipActive",
				func(logicID int32, busyChipList []string) (bool, error) {
					return true, nil
				})
			defer patch1.Reset()
			ret := manager.canA3BeReset(dev)
			convey.So(ret, convey.ShouldBeTrue)
		})
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
		convey.Convey("A3 device success, should return nil", func() {
			common.ParamOption.RealCardType = common.Ascend910A3
			patch1 := gomonkey.ApplyPrivateMethod(manager, "setUnhealthyForA3",
				func(devStatusList []*common.NpuDevice) error {
					return nil
				})
			defer patch1.Reset()
			isHotResetOn = false
			err := manager.setAllDevUnhealthyOnRing(devList)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestSetUnhealthyForA3 test the function setUnhealthyForA3
func TestSetUnhealthyForA3(t *testing.T) {
	devs := []*common.NpuDevice{
		{LogicID: id1},
		{LogicID: id2},
	}
	manager := createFake910Manager()
	convey.Convey("test setUnhealthyForA3", t, func() {
		inResetDev = int32(id1)
		convey.Convey("02-get associated card error, should return error", func() {
			patch1 := gomonkey.ApplyFuncReturn(IsDevBusy, true)
			patch1.ApplyPrivateMethod(manager, "getAssociatedLogicIDs",
				func(logicID, cardID, deviceID int32) ([]int32, error) {
					return nil, testErr
				})
			defer patch1.Reset()
			err := manager.setUnhealthyForA3(devs)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-success, should return nil", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "getAssociatedLogicIDs",
				func(logicID, cardID, deviceID int32) ([]int32, error) {
					return []int32{logicID}, nil
				})
			defer patch1.Reset()
			err := manager.setUnhealthyForA3(devs)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetAssociatedLogicIDs test the function getAssociatedLogicIDs
func TestGetAssociatedLogicIDs(t *testing.T) {
	manager := createFake910Manager()
	const id1int32 = int32(id1)
	convey.Convey("test getAssociatedLogicIDs", t, func() {
		convey.Convey("01-get brother card error, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBrotherCardID",
				id1int32, testErr)
			defer patch1.Reset()
			_, err := manager.getAssociatedLogicIDs(id1int32, id1int32, id1int32)
			convey.So(err, convey.ShouldBeError)
		})
		patch := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBrotherCardID",
			id1int32, nil)
		defer patch.Reset()
		convey.Convey("02-get logic id failed, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDeviceLogicID",
				id1int32, testErr)
			defer patch1.Reset()
			_, err := manager.getAssociatedLogicIDs(id1int32, id1int32, id1int32)
			convey.So(err, convey.ShouldBeError)
		})
		patch.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDeviceLogicID",
			id1int32, nil)
		convey.Convey("03-success, should return nil", func() {
			_, err := manager.getAssociatedLogicIDs(id1int32, id1int32, id1int32)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestTryResetDevice an ut for function tryResetDevice
func TestTryResetDevice(t *testing.T) {
	manager := createFake910Manager()
	patch := gomonkey.ApplyFunc(AddBusyDev, func(cardID, deviceID int32) {
		return
	})
	patch.ApplyFunc(AddResetCnt, func(cardID, deviceID int32) {
		return
	})
	defer patch.Reset()
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
		mockHotResetComplete.ApplyPrivateMethod(manager, "getResetIndexForA3",
			func(logicID int32) (int32, error) {
				return int32(id1), nil
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
		api.PodRankIndexAnno:    "1",
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

func getSinglePodWithMoreInfo(podName string, annotation map[string]string, labels map[string]string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        podName,
			Annotations: annotation,
			Labels:      labels,
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
	mockGetNeedResetDevMapPatch.ApplyPrivateMethod(&HwAscend910Manager{}, "getA3LogicMapByAssociation",
		func(devFaultInfoList []*common.TaskDevInfo) (map[int32]int32, error) {
			return nil, testErr
		})
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

// TestGetA3LogicMapByAssociation test the function getA3LogicMapByAssociation
func TestGetA3LogicMapByAssociation(t *testing.T) {
	convey.Convey("test getNeedResetDeviceLogicIdMap", t, func() {
		manager := createFake910Manager()
		devs := []*common.TaskDevInfo{
			{DevFaultInfo: common.DevFaultInfo{LogicId: int32(id1)}},
		}
		convey.Convey("01-not A3, should return error", func() {
			common.ParamOption.RealCardType = common.Ascend910B
			_, err := manager.getA3LogicMapByAssociation(devs)
			convey.So(err, convey.ShouldBeError)
		})
		common.ParamOption.RealCardType = common.Ascend910A3
		convey.Convey("02-get cardId failed, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
				int32(id1), int32(id1), testErr)
			defer patch1.Reset()
			_, err := manager.getA3LogicMapByAssociation(devs)
			convey.So(err, convey.ShouldBeError)
		})
		patch := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
			int32(id1), int32(id1), nil)
		defer patch.Reset()
		convey.Convey("03-get associated card failed, should return error", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "getAssociatedLogicIDs",
				func(logicID, cardID, deviceID int32) ([]int32, error) {
					return nil, testErr
				})
			defer patch1.Reset()
			_, err := manager.getA3LogicMapByAssociation(devs)
			convey.So(err, convey.ShouldBeError)
		})
		patch.ApplyPrivateMethod(manager, "getAssociatedLogicIDs",
			func(logicID, cardID, deviceID int32) ([]int32, error) {
				return []int32{int32(id1)}, nil
			})
		convey.Convey("04-success, should return nil", func() {
			_, err := manager.getA3LogicMapByAssociation(devs)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestExecRescan test the function execRescan
func TestExecRescan(t *testing.T) {
	manager := createFake910Manager()
	devs := []ResetDevice{
		{CardId: int32(id1), DeviceId: int32(id2)},
	}
	patch := gomonkey.ApplyFunc(WriteResetInfo, func(resetInfo ResetInfo, writeMode WriteMode, update bool) {
		return
	})
	flag := false
	patch.ApplyFunc(FreeBusyDev, func(cardID, deviceID int32) {
		flag = true
	})
	defer patch.Reset()
	convey.Convey("test execRescan", t, func() {
		convey.Convey("01-rescan error, flag should be false", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "RescanSoc", testErr)
			defer patch1.Reset()
			manager.execRescan(devs)
			convey.So(flag, convey.ShouldBeFalse)
		})
		convey.Convey("02-rescan success, flag should be true", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "RescanSoc", nil)
			defer patch1.Reset()
			manager.execRescan(devs)
			convey.So(flag, convey.ShouldBeTrue)
		})
	})
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
				patch1 := mockGetNeedResetDevMap().
					ApplyFunc(time.After, func(d time.Duration) <-chan time.Time {
						ch := make(chan time.Time, 1)
						ch <- time.Now()
						return ch
					})
				defer patch1.Reset()
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

func mockGetDeviceNetWorkHealth(code uint32, err error) *gomonkey.Patches {
	return gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
		"GetDeviceNetWorkHealth", code, err)
}

// TestHandleResetProcess for test handleResetProcess
func TestHandleResetProcess(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test handleResetProcess", t, func() {
		manager.hotResetManager = &HotResetTools{
			resetDevNumOnce: common.Ascend910RingsNum,
			globalDevFaultInfo: map[int32]*common.DevFaultInfo{
				chipPhyID0: {Policy: common.EmptyError},
			},
		}
		patch := gomonkey.ApplyFuncReturn((*HwAscend910Manager).execHotReset, nil).
			ApplyMethodReturn(&devmanager.DeviceManagerMock{},
				"GetDeviceAllErrorCode", int32(0), []int64{}, errors.New("get error code failed")).
			ApplyFunc(common.SetDeviceInit, func(int32) {}).
			ApplyGlobalVar(&isolateDevList, []int32{})
		defer patch.Reset()
		classifyDevs := map[string][]*common.NpuDevice{
			common.Ascend910: {{LogicID: chipPhyID0, Health: ""}},
		}
		npuDev := &common.NpuDevice{LogicID: chipPhyID0}
		devInfo := &common.DevFaultInfo{LogicId: chipPhyID0}
		manager.handleResetProcess(classifyDevs, devInfo, npuDev)
		convey.So(classifyDevs[common.Ascend910][0].Health, convey.ShouldEqual, v1beta1.Healthy)
		convey.So(isolateDevList, convey.ShouldResemble, []int32{chipPhyID0})
	})
}

// TestHandleL2L3FaultRestart for test handleL2L3FaultRestart
func TestHandleL2L3FaultRestart(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test handleL2L3FaultRestart", t, func() {
		patch := gomonkey.ApplyGlobalVar(&isL3FaultExistMap, map[int32]bool{id2: true, id3: false})
		defer patch.Reset()
		convey.Convey("01-free reset error,should return false", func() {
			devFaultInfo := &common.DevFaultInfo{LogicId: id1, Policy: common.FreeResetError}
			l3RestartFlag := manager.handleL2L3FaultRestart(devFaultInfo)
			convey.So(l3RestartFlag, convey.ShouldBeFalse)
			convey.So(isL3FaultExistMap[id1], convey.ShouldBeFalse)
		})

		convey.Convey("02-restart error and value is true in isL3FaultExistMap, should return true", func() {
			devFaultInfo := &common.DevFaultInfo{LogicId: id2, Policy: common.RestartError}
			l3RestartFlag := manager.handleL2L3FaultRestart(devFaultInfo)
			convey.So(l3RestartFlag, convey.ShouldBeTrue)
			convey.So(isL3FaultExistMap[id2], convey.ShouldBeFalse)
		})
		convey.Convey("03-restart error and value is false in isL3FaultExistMap, should return false", func() {
			devFaultInfo := &common.DevFaultInfo{LogicId: id3, Policy: common.RestartError}
			l3RestartFlag := manager.handleL2L3FaultRestart(devFaultInfo)
			convey.So(l3RestartFlag, convey.ShouldBeFalse)
			convey.So(isL3FaultExistMap[id3], convey.ShouldBeTrue)
		})
	})
}

// TestGetResetIndexForA3 for test getResetIndexForA3
func TestGetResetIndexForA3(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test getResetIndexForA3", t, func() {
		convey.Convey("01-get card id device id failed, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
				int32(id1), int32(id1), errors.New("get card id device id failed"))
			defer patch1.Reset()
			retId, err := manager.getResetIndexForA3(chipPhyID0)
			convey.So(err.Error(), convey.ShouldEqual, "get card id device id failed")
			convey.So(retId, convey.ShouldEqual, errorId)
		})
		patch := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
			int32(id1), int32(id1), nil)
		defer patch.Reset()
		convey.Convey("02-get card id device id failed, should return error", func() {
			patch1 := gomonkey.ApplyFuncReturn((*HwAscend910Manager).getAssociatedLogicIDs, []int32{}, nil)
			defer patch1.Reset()
			retId, err := manager.getResetIndexForA3(chipPhyID0)
			convey.So(err.Error(), convey.ShouldEqual, "sort logic ids failed, logic ids [], sorted ids []")
			convey.So(retId, convey.ShouldEqual, errorId)
		})

		convey.Convey("03-get reset index success, should return nil", func() {
			patch1 := gomonkey.ApplyFuncReturn((*HwAscend910Manager).getAssociatedLogicIDs, []int32{id1, id2, id3}, nil)
			defer patch1.Reset()
			retId, err := manager.getResetIndexForA3(chipPhyID0)
			convey.So(err, convey.ShouldBeNil)
			convey.So(retId, convey.ShouldEqual, int32(id1))
		})
	})
}

// TestUpgradeHotResetError for test upgradeHotResetError
func TestUpgradeHotResetError(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test upgradeHotResetError", t, func() {
		npuDev := &common.NpuDevice{LogicID: chipPhyID0}
		patch := gomonkey.ApplyGlobalVar(&isolateDevList, []int32{})
		defer patch.Reset()
		convey.Convey("01-Ascend910 dev is empty, should not upgrade", func() {
			manager.upgradeHotResetError(map[string][]*common.NpuDevice{}, npuDev)
			convey.So(isolateDevList, convey.ShouldResemble, []int32{chipPhyID0})
		})
		classifyDevs := map[string][]*common.NpuDevice{
			common.Ascend910: {
				{LogicID: chipPhyID0, Health: ""},
				{LogicID: chipPhyID1, Health: ""},
				{LogicID: chipPhyID2, Health: ""},
				{LogicID: chipPhyID3, Health: v1beta1.Unhealthy},
				{LogicID: chipPhyID4, Health: v1beta1.Unhealthy},
			},
		}
		convey.Convey("02-resetDevNumOnce is 0, should not upgrade", func() {
			manager.hotResetManager = &HotResetTools{}
			manager.upgradeHotResetError(classifyDevs, npuDev)
			convey.So(isolateDevList, convey.ShouldResemble, []int32{chipPhyID0})
			convey.So(classifyDevs[common.Ascend910][0].Health, convey.ShouldBeEmpty)
			convey.So(classifyDevs[common.Ascend910][3].Health, convey.ShouldEqual, v1beta1.Unhealthy)
		})
		convey.Convey("03-upgrade success", func() {
			manager.hotResetManager = &HotResetTools{
				resetDevNumOnce: common.Ascend910RingsNum,
				globalDevFaultInfo: map[int32]*common.DevFaultInfo{
					chipPhyID1: {Policy: common.RestartError},
					chipPhyID2: {Policy: common.RestartError},
					chipPhyID3: {Policy: common.EmptyError},
					chipPhyID4: {Policy: common.IgnoreError},
				},
			}
			manager.upgradeHotResetError(classifyDevs, npuDev)
			convey.So(isolateDevList, convey.ShouldResemble, []int32{chipPhyID0})
			convey.So(classifyDevs[common.Ascend910][0].Health, convey.ShouldBeEmpty)
			convey.So(classifyDevs[common.Ascend910][1].Health, convey.ShouldBeEmpty)
			convey.So(classifyDevs[common.Ascend910][2].Health, convey.ShouldBeEmpty)
			convey.So(classifyDevs[common.Ascend910][3].Health, convey.ShouldEqual, v1beta1.Healthy)
			convey.So(classifyDevs[common.Ascend910][4].Health, convey.ShouldEqual, v1beta1.Unhealthy)
		})
	})
}

// TestRefreshDevFaultInfoForResetProcess for test refreshDevFaultInfoForResetProcess
func TestRefreshDevFaultInfoForResetProcess(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test refreshDevFaultInfoForResetProcess", t, func() {
		manager.hotResetManager = &HotResetTools{}
		devInfo := &common.DevFaultInfo{LogicId: chipPhyID0}
		convey.Convey("01-error code is empty, should return false and nil", func() {
			isShouldUpgrade, err := manager.refreshDevFaultInfoForResetProcess(devInfo)
			convey.So(isShouldUpgrade, convey.ShouldBeFalse)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-error code is empty, should return true and error", func() {
			patch := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
				"GetDeviceAllErrorCode", int32(0), []int64{}, errors.New("get error code failed"))
			defer patch.Reset()
			isShouldUpgrade, err := manager.refreshDevFaultInfoForResetProcess(devInfo)
			convey.So(isShouldUpgrade, convey.ShouldBeTrue)
			convey.So(err.Error(), convey.ShouldEqual, "get error code failed")
		})
		convey.Convey("03-refresh success, should return true and nil", func() {
			patch := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
				"GetDeviceAllErrorCode", int32(0), []int64{1}, nil).
				ApplyFuncReturn(common.GetFaultType, common.RestartRequest)
			defer patch.Reset()
			isShouldUpgrade, err := manager.refreshDevFaultInfoForResetProcess(devInfo)
			convey.So(isShouldUpgrade, convey.ShouldBeTrue)
			convey.So(err, convey.ShouldBeNil)
			convey.So(devInfo.Policy, convey.ShouldEqual, common.RestartRequestError)
		})
	})
}

// TestGetDeviceNetworkState for test getDeviceNetworkState
func TestGetDeviceNetworkState(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test getDeviceNetworkState", t, func() {
		convey.Convey("01-classifyDevs is empty, should return error", func() {
			mockHealth := mockGetDeviceNetWorkHealth(uint32(0), errors.New("get network health failed"))
			defer mockHealth.Reset()
			_, err := manager.getDeviceNetworkState(chipPhyID0, v1beta1.Healthy)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-health code is networkDetectOK, should return healthy and nil", func() {
			mockHealth := mockGetDeviceNetWorkHealth(networkDetectOK, nil)
			defer mockHealth.Reset()
			code, err := manager.getDeviceNetworkState(chipPhyID0, v1beta1.Unhealthy)
			convey.So(err, convey.ShouldBeNil)
			convey.So(code, convey.ShouldEqual, v1beta1.Healthy)
		})
		convey.Convey("03-health code is unexpected value, should return unhealthy and nil", func() {
			mockHealth := mockGetDeviceNetWorkHealth(uint32(1), nil)
			defer mockHealth.Reset()
			code, err := manager.getDeviceNetworkState(chipPhyID0, v1beta1.Healthy)
			convey.So(err, convey.ShouldBeNil)
			convey.So(code, convey.ShouldEqual, v1beta1.Unhealthy)
		})
	})
}

// TestUpdateHotResetCache for test updateHotResetCache
func TestUpdateHotResetCache(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test updateHotResetCache", t, func() {
		manager.hotResetManager = &HotResetTools{}
		patch := gomonkey.ApplyGlobalVar(&isolateDevList, []int32{})
		defer patch.Reset()
		convey.Convey("01-classifyDevs is empty, should return error", func() {
			err := manager.updateHotResetCache(nil)
			convey.So(err.Error(), convey.ShouldEqual, "ascend 910 device list not found")
		})
		convey.Convey("02-Ascend910 dev is empty, should return error", func() {
			classifyDevs := map[string][]*common.NpuDevice{common.Ascend910: {}}
			err := manager.updateHotResetCache(classifyDevs)
			convey.So(err.Error(), convey.ShouldEqual, "npu device list is nil")
		})
		convey.Convey("03-set task dev info cache failed, should return error", func() {
			patch1 := gomonkey.ApplyGlobalVar(&isolateDevList, []int32{chipPhyID1, chipPhyID2}).
				ApplyPrivateMethod(&HwAscend910Manager{}, "setTaskDevInfoCache",
					func(*HwAscend910Manager) error { return errors.New("set task dev info cache error") }).
				ApplyFuncReturn(common.GetFaultType, common.NotHandleFault)
			defer patch1.Reset()
			classifyDevs := map[string][]*common.NpuDevice{
				common.Ascend910: {
					{LogicID: chipPhyID1, Health: v1beta1.Unhealthy},
					{LogicID: chipPhyID2, Health: v1beta1.Healthy}},
			}
			err := manager.updateHotResetCache(classifyDevs)
			convey.So(err.Error(), convey.ShouldEqual, "set task dev info cache error")
			convey.So(isolateDevList, convey.ShouldResemble, []int32{chipPhyID1})
			devFaultInfo, err := manager.hotResetManager.GetGlobalDevFaultInfo(chipPhyID1)
			convey.So(err, convey.ShouldBeNil)
			convey.So(devFaultInfo, convey.ShouldNotBeNil)
			convey.So(devFaultInfo.Policy, convey.ShouldEqual, common.IsolateError)
		})
	})
}

// TestUpdateUpgradeErrorInfo for test updateUpgradeErrorInfo
func TestUpdateUpgradeErrorInfo(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test  updateUpgradeErrorInfo", t, func() {
		patch := gomonkey.ApplyGlobalVar(&isolateDevList, []int32{})
		defer patch.Reset()
		convey.Convey("01-isolateDevList is empty, should return nil", func() {
			err := manager.updateUpgradeErrorInfo(nil)
			convey.So(err, convey.ShouldBeNil)
		})
		patch.ApplyGlobalVar(&isolateDevList, []int32{chipPhyID2, chipPhyID4})
		convey.Convey("02-classifyDevs is empty, should return error", func() {
			err := manager.updateUpgradeErrorInfo(map[string][]*common.NpuDevice{})
			convey.So(err.Error(), convey.ShouldEqual, "no Ascend 910 device found in cache")
		})
		convey.Convey("03-update success, should return nil", func() {
			classifyDevs := map[string][]*common.NpuDevice{
				common.Ascend910: {
					{
						LogicID: chipPhyID1,
						Health:  v1beta1.Healthy,
					},
					{
						LogicID: chipPhyID2,
						Health:  v1beta1.Unhealthy,
					},
					{
						LogicID: chipPhyID4,
						Health:  v1beta1.Healthy,
					},
					{
						LogicID: chipPhyID5,
						Health:  v1beta1.Healthy,
					},
				},
			}
			err := manager.updateUpgradeErrorInfo(classifyDevs)
			convey.So(err, convey.ShouldBeNil)
			convey.So(isolateDevList, convey.ShouldResemble, []int32{chipPhyID2})
		})
	})
}

func mockSetTaskDevInfoCacheFuncData1() []v1.Pod {
	annotationDevError := map[string]string{
		common.HuaweiAscend910: "Ascend910#4,Ascend910#5,Ascend910#6",
	}
	annotationSuccess := map[string]string{
		common.HuaweiAscend910: "Ascend910-4,Ascend910-5,Ascend910-6",
	}
	annotationSuccess1 := map[string]string{
		common.HuaweiAscend910: "Ascend910-4,Ascend910-5,Ascend910-6",
		api.PodRankIndexAnno:   "2",
	}
	labelSuccess := map[string]string{
		common.ResetTaskNameKey: "task1",
	}
	podList := []v1.Pod{
		getSinglePodWithMoreInfo("test-pod1", map[string]string{}, labelSuccess),
		getSinglePodWithMoreInfo("test-pod2", annotationDevError, labelSuccess),
		getSinglePodWithMoreInfo("test-pod3", annotationSuccess, labelSuccess),
		getSinglePodWithMoreInfo("test-pod4", annotationSuccess1, labelSuccess),
	}
	return podList
}

func mockSetTaskDevInfoCacheFuncData2() []v1.Pod {
	annotationError := map[string]string{
		common.HuaweiAscend910: "Ascend910-4,Ascend910-5,Ascend910-6",
		api.PodRankIndexAnno:   "1.2",
	}
	labelSuccess := map[string]string{
		common.ResetTaskNameKey: "task1",
	}
	podList := []v1.Pod{
		getSinglePodWithMoreInfo("test-pod1", annotationError, labelSuccess),
	}
	return podList
}

// TestSetTaskDevInfoCache for test setTaskDevInfoCache
func TestSetTaskDevInfoCache(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test setTaskDevInfoCache", t, func() {
		manager.hotResetManager = &HotResetTools{
			resetDevNumOnce: common.Ascend910RingsNum,
			globalDevFaultInfo: map[int32]*common.DevFaultInfo{
				chipPhyID4: {}, chipPhyID5: {}, chipPhyID6: {},
			},
		}
		patch := gomonkey.ApplyMethod(&devmanager.DeviceManagerMock{}, "GetLogicIDFromPhysicID",
			func(dm *devmanager.DeviceManagerMock, physicID int32) (int32, error) {
				return physicID, nil
			})
		defer patch.Reset()
		convey.Convey("01-set task dev info success, should return nil", func() {
			podList := mockSetTaskDevInfoCacheFuncData1()
			patch1 := gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "GetActivePodListCache", podList)
			defer patch1.Reset()
			err := manager.setTaskDevInfoCache()
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test setTaskDevInfoCache", func() {
			podList := mockSetTaskDevInfoCacheFuncData2()
			patch1 := gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "GetActivePodListCache", podList)
			defer patch1.Reset()
			err := manager.setTaskDevInfoCache()
			convey.So(err.Error(), convey.ShouldEqual, `strconv.Atoi: parsing "1.2": invalid syntax`)
		})
	})
}

// TestHandleUpdateCaches for test handleUpdateCaches
func TestHandleUpdateCaches(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test handleUpdateCaches", t, func() {
		taskName := "taskName"
		usedDevice := map[string]struct{}{taskName: {}}
		taskDevList := map[string][]int32{taskName: {1, 2, 3}}
		taskDevFaultInfo := map[string][]*common.TaskDevInfo{}
		taskPod := map[string]v1.Pod{}
		manager.hotResetManager = &HotResetTools{
			allTaskDevList: taskDevList,
			resetTask:      map[string]struct{}{taskName: {}},
		}
		err := manager.handleUpdateCaches(usedDevice, nil, taskDevFaultInfo, taskPod)
		convey.So(err.Error(), convey.ShouldEqual, "task device list is nil")
		err = manager.handleUpdateCaches(usedDevice, taskDevList, nil, taskPod)
		convey.So(err.Error(), convey.ShouldEqual, "taskDevFaultInfo is nil")
		err = manager.handleUpdateCaches(usedDevice, taskDevList, taskDevFaultInfo, nil)
		convey.So(err.Error(), convey.ShouldEqual, "taskPod is nil")
		err = manager.handleUpdateCaches(usedDevice, taskDevList, taskDevFaultInfo, taskPod)
		convey.So(err, convey.ShouldBeNil)
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
		resetInfo := &common.TaskResetInfo{RankList: make([]*common.TaskDevInfo, 0)}
		convey.Convey("01-policy level is 2, call restartRequestProcess, should return nil", func() {
			manager := createFake910Manager()
			manager.hotResetManager = &HotResetTools{}
			err := manager.runProcessTask(taskName, common.RestartRequestErrorLevel, resetInfo, nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-policy level is 3, call restartProcess, should return nil", func() {
			manager := createFake910Manager()
			manager.hotResetManager = &HotResetTools{}
			err := manager.runProcessTask(taskName, common.RestartErrorLevel, resetInfo, nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("03-policy level is 5, call resetProcess, should return nil", func() {
			manager := createFake910Manager()
			manager.hotResetManager = &HotResetTools{}
			err := manager.runProcessTask(taskName, common.ResetErrorLevel, resetInfo, nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("04-policy level is 6, call resetProcess, should return nil", func() {
			manager := createFake910Manager()
			manager.hotResetManager = &HotResetTools{}
			err := manager.runProcessTask(taskName, common.IsolateErrorLevel, resetInfo, nil)
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
			patch1 := gomonkey.ApplyPrivateMethod(manager, "canResetDevice", func(_, _, _ int32) bool {
				return false
			})
			patch1.ApplyPrivateMethod(manager, "updateResetInfo", func(_, _ []ResetDevice) {
				return
			})
			defer patch1.Reset()
			manager.hotResetManager = &HotResetTools{
				faultDev2PodMap: map[int32]v1.Pod{},
				resetDevNumOnce: 1,
			}
			devMap := map[int32]int32{chipPhyID0: chipPhyID0}
			convey.So(manager.execResetDevice(devMap), convey.ShouldBeNil)
		})
	})
}

// TestGetNeedResetDevMapForA3 test the function getNeedResetDevMapForA3
func TestGetNeedResetDevMapForA3(t *testing.T) {
	manager := createFake910Manager()
	devs := []*common.TaskDevInfo{
		{DevFaultInfo: common.DevFaultInfo{
			LogicId: int32(id1),
			Policy:  common.RestartError,
		}},
	}
	convey.Convey("test getNeedResetDevMapForA3", t, func() {
		convey.Convey("01-not A3, should return error", func() {
			common.ParamOption.RealCardType = common.Ascend910B
			_, err := manager.getNeedResetDevMapForA3(devs)
			convey.So(err, convey.ShouldBeError)
		})
		common.ParamOption.RealCardType = common.Ascend910A3
		convey.Convey("02-get index error, should return error", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "getResetIndexForA3",
				func(_ *HwAscend910Manager, logicID int32) (int32, error) {
					return errorId, testErr
				})
			defer patch1.Reset()
			_, err := manager.getNeedResetDevMapForA3(devs)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-success, should return nil", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "getResetIndexForA3",
				func(logicID int32) (int32, error) {
					return errorId, nil
				})
			defer patch1.Reset()
			_, err := manager.getNeedResetDevMapForA3(devs)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestCanResetDevice test the function canResetDevice
func TestCanResetDevice(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test canResetDevice", t, func() {
		convey.Convey("01-dev busy, should return false", func() {
			patch1 := gomonkey.ApplyFuncReturn(IsDevBusy, true)
			defer patch1.Reset()
			convey.So(manager.canResetDevice(id1, id1), convey.ShouldBeFalse)
		})
		patch := gomonkey.ApplyFuncReturn(IsDevBusy, false)
		defer patch.Reset()
		convey.Convey("02-reset cnt over, should return false", func() {
			patch1 := gomonkey.ApplyFuncReturn(GetResetCnt, common.MaxResetTimes+id1)
			defer patch1.Reset()
			convey.So(manager.canResetDevice(id1, id1), convey.ShouldBeFalse)
		})
		patch.ApplyFuncReturn(GetResetCnt, common.MaxResetTimes-id1)
		convey.Convey("03-success, should return true", func() {
			convey.So(manager.canResetDevice(id1, id1), convey.ShouldBeTrue)
		})
	})
}

// TestExecOutBandReset test the function execOutBandReset
func TestExecOutBandReset(t *testing.T) {
	manager := createFake910Manager()
	const testCardID, testDeviceID = 0, 0
	convey.Convey("test execOutBandReset", t, func() {
		patch := gomonkey.ApplyPrivateMethod(manager, "updateResetInfo", func(failDevs, sucDevs []ResetDevice) {
			return
		})
		patch.ApplyPrivateMethod(manager, "scanDeviceForThirdParty",
			func(failDevs []ResetDevice) {
				return
			})
		patch.ApplyPrivateMethod(manager, "fillResetDevs", func(devs []ResetDevice) ([]ResetDevice, error) {
			return devs, nil
		})
		defer patch.Reset()
		common.ParamOption.RealCardType = common.Ascend910A3
		convey.Convey("01-reset error, should return error", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "resetDeviceOutBand",
				func(cardId, deviceId int32) error {
					return ascend910testErr
				})
			defer patch1.Reset()
			err := manager.execOutBandReset([]ResetDevice{
				{CardId: testCardID, DeviceId: testDeviceID},
			}, []ResetDevice{})
			convey.So(err, convey.ShouldBeError)
		})
		patch.ApplyPrivateMethod(manager, "resetDeviceOutBand",
			func(cardId, deviceId int32) error {
				return nil
			})
		convey.Convey("02-success, should return nil", func() {
			err := manager.execOutBandReset([]ResetDevice{
				{CardId: testCardID, DeviceId: testDeviceID},
			}, []ResetDevice{})
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestUpdateResetInfo test the function updateResetInfo
func TestUpdateResetInfo(t *testing.T) {
	manager := createFake910Manager()
	const id = 0
	var testDevs = []ResetDevice{
		{CardId: id},
	}
	convey.Convey("test updateResetInfo", t, func() {
		patch := gomonkey.ApplyFuncReturn(GetResetInfoMgr, ResetInfoMgr{})
		patch.ApplyFuncReturn(ReadResetInfo, ResetInfo{})
		ri := ResetInfo{}
		patch.ApplyFunc(WriteResetInfo, func(resetInfo ResetInfo, writeMode WriteMode, update bool) {
			ri = resetInfo
			return
		})
		patch.ApplyPrivateMethod(manager, "fillResetDevs",
			func(_ *HwAscend910Manager, devs []ResetDevice) ([]ResetDevice, error) {
				return devs, nil
			})
		defer patch.Reset()
		convey.Convey("01-A3 device, should append to third party", func() {
			common.ParamOption.RealCardType = common.Ascend910A3
			manager.updateResetInfo(testDevs, []ResetDevice{})
			convey.So(ri.ThirdPartyResetDevs, convey.ShouldResemble, testDevs)
		})
		convey.Convey("02-not A3, should append to manual devices", func() {
			common.ParamOption.RealCardType = common.Ascend910B
			manager.updateResetInfo(testDevs, []ResetDevice{})
			convey.So(ri.ManualResetDevs, convey.ShouldResemble, testDevs)
		})
	})
}

// TestResetDeviceOutBand test the function resetDeviceOutBand
func TestResetDeviceOutBand(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test resetDeviceOutBand", t, func() {
		const testCardID, testDeviceID = 0, 0
		convey.Convey("01-out band channel error, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
				"GetOutBandChannelState", ascend910testErr)
			defer patch1.Reset()
			err := manager.resetDeviceOutBand(testCardID, testDeviceID)
			convey.So(err, convey.ShouldBeError)
		})
		patch := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetOutBandChannelState", nil)
		defer patch.Reset()
		convey.Convey("02-pre reset error, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
				"PreResetSoc", ascend910testErr)
			defer patch1.Reset()
			err := manager.resetDeviceOutBand(testCardID, testDeviceID)
			convey.So(err, convey.ShouldBeError)
		})
		patch.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"PreResetSoc", nil)
		convey.Convey("03-reset out band error, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
				"SetDeviceResetOutBand", ascend910testErr)
			defer patch1.Reset()
			err := manager.resetDeviceOutBand(testCardID, testDeviceID)
			convey.So(err, convey.ShouldBeError)
		})
		patch.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"SetDeviceResetOutBand", nil)
		convey.Convey("04-rescan error, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
				"RescanSoc", ascend910testErr)
			defer patch1.Reset()
			err := manager.resetDeviceOutBand(testCardID, testDeviceID)
			convey.So(err, convey.ShouldBeError)
		})
		patch.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"RescanSoc", nil)
		convey.Convey("05-success, should return nil", func() {
			err := manager.resetDeviceOutBand(testCardID, testDeviceID)
			convey.So(err, convey.ShouldBeNil)
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
			mockHealth := mockGetDeviceNetWorkHealth(uint32(0), errors.New("get network health failed"))
			defer mockHealth.Reset()
			convey.So(manager.isNetResetCompleted(logicID), convey.ShouldBeFalse)
		})
		convey.Convey("02-network status is unhealthy, should return false", func() {
			mockHealth := mockGetDeviceNetWorkHealth(uint32(1), nil)
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

func mockFilterCheck() *gomonkey.Patches {
	patch := gomonkey.ApplyMethodReturn(&HotResetTools{}, "GetDevListInReset",
		map[int32]struct{}{
			int32(id1): {},
		})
	patch.ApplyPrivateMethod(NewHwAscend910Manager(), "isDevShouldBeIsolate",
		func(faultyDevLogicId int32) bool {
			return false
		})
	return patch
}

// TestFilterDevStatusForA3 test the function filterDevStatusForA3
func TestFilterDevStatusForA3(t *testing.T) {
	manager := createFake910Manager()
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
	devs := []*common.NpuDevice{
		{LogicID: int32(id1)},
	}
	patch := mockFilterCheck()
	defer patch.Reset()
	convey.Convey("test TestFilterDevStatusForA3", t, func() {
		convey.Convey("01-get associated card error, should return error", func() {
			patch1 := gomonkey.ApplyFuncReturn(IsDevBusy, false)
			patch1.ApplyPrivateMethod(manager, "getAssociatedLogicIDs",
				func(logicID, cardID, deviceID int32) ([]int32, error) {
					return []int32{id1}, testErr
				})
			defer patch1.Reset()
			err := manager.filterDevStatusForA3(devs)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-success, should return nil", func() {
			patch1 := gomonkey.ApplyFuncReturn(IsDevBusy, false)
			patch1.ApplyPrivateMethod(manager, "getAssociatedLogicIDs",
				func(logicID, cardID, deviceID int32) ([]int32, error) {
					return []int32{id1}, nil
				})
			defer patch1.Reset()
			err := manager.filterDevStatusForA3(devs)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestFillResetDevs test the function fillResetDevs
func TestFillResetDevs(t *testing.T) {
	manager := createFake910Manager()
	devs := []ResetDevice{
		{LogicID: id1, PhyID: id1},
	}
	common.ParamOption.RealCardType = common.Ascend910A3
	convey.Convey("test fillResetDevs", t, func() {
		convey.Convey("01-get npu error, should return err", func() {
			patch1 := gomonkey.ApplyMethodReturn(manager, "GetNPUs",
				common.NpuAllInfo{}, testErr)
			defer patch1.Reset()
			_, err := manager.fillResetDevs(devs)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-logic id not exist, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(manager, "GetNPUs",
				common.NpuAllInfo{AllDevs: []common.NpuDevice{
					{LogicID: id2},
				}}, nil)
			defer patch1.Reset()
			_, err := manager.fillResetDevs(devs)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-success, should return nil", func() {
			patch1 := gomonkey.ApplyMethodReturn(manager, "GetNPUs",
				common.NpuAllInfo{AllDevs: []common.NpuDevice{
					{LogicID: id1},
				}}, nil)
			patch1.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBrotherCardID",
				int32(id1), nil)
			defer patch1.Reset()
			_, err := manager.fillResetDevs(devs)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCanResetDeviceByLogicID(t *testing.T) {
	convey.Convey("test canResetDeviceByLogicID", t, func() {
		manager := createFake910Manager()
		patch := gomonkey.ApplyPrivateMethod(manager, "canResetDevice", func(cardID, deviceID int32) bool {
			return true
		})
		defer patch.Reset()
		convey.Convey("01-get card id error, should return false", func() {
			patch1 := gomonkey.ApplyMethodReturn(manager.GetDmgr(), "GetCardIDDeviceID", int32(id1),
				int32(id1), testErr)
			defer patch1.Reset()
			convey.So(manager.canResetDeviceByLogicID(int32(id1)), convey.ShouldBeFalse)
		})
		convey.Convey("02-success, should return true", func() {
			patch1 := gomonkey.ApplyMethodReturn(manager.GetDmgr(), "GetCardIDDeviceID", int32(id1),
				int32(id1), nil)
			defer patch1.Reset()
			convey.So(manager.canResetDeviceByLogicID(int32(id1)), convey.ShouldBeTrue)
		})
	})
}

func TestGetResetIndex(t *testing.T) {
	convey.Convey("test getResetIndex", t, func() {
		manager := createFake910Manager()
		manager.hotResetManager = newTestHotResetManager(common.Ascend910, common.Train)
		dev := &common.NpuDevice{}
		convey.Convey("01-A3, get idx success, should return nil", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "getResetIndexForA3",
				func(logicID int32) (int32, error) {
					return int32(id1), nil
				})
			defer patch1.Reset()
			_, err := manager.getResetIndex(dev)
			convey.So(err, convey.ShouldBeNil)
		})
		patch := gomonkey.ApplyPrivateMethod(manager, "getResetIndexForA3",
			func(logicID int32) (int32, error) {
				return int32(id1), testErr
			})
		defer patch.Reset()
		convey.Convey("02-get dev num once err, should return err", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager.hotResetManager, "GetResetDevNumOnce",
				func() (int, error) {
					return id1, testErr
				})
			defer patch1.Reset()
			_, err := manager.getResetIndex(dev)
			convey.So(err, convey.ShouldBeError)
		})
		patch.ApplyPrivateMethod(manager.hotResetManager, "GetResetDevNumOnce",
			func() (int, error) {
				return id1, nil
			})
		convey.Convey("03-success, should return nil", func() {
			_, err := manager.getResetIndex(dev)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func getDevFaultInfoTestCases() []struct {
	name           string
	logicID        int32
	mockFaultInfo  *common.DevFaultInfo
	mockErr        error
	expectedResult *common.DevFaultInfo
} {
	return []struct {
		name           string
		logicID        int32
		mockFaultInfo  *common.DevFaultInfo
		mockErr        error
		expectedResult *common.DevFaultInfo
	}{
		{
			name:    "get fault info success, should return fault",
			logicID: 1,
			mockFaultInfo: &common.DevFaultInfo{
				LogicId: 1,
				Policy:  "test_policy",
			},
			mockErr: nil,
			expectedResult: &common.DevFaultInfo{
				LogicId: 1,
				Policy:  "test_policy",
			},
		},
		{
			name:           "get fault error, should return nil",
			logicID:        2,
			mockFaultInfo:  nil,
			mockErr:        testErr,
			expectedResult: nil,
		},
		{
			name:    "no fault, should return nil",
			logicID: 3,
			mockFaultInfo: &common.DevFaultInfo{
				LogicId: 3,
				Policy:  common.EmptyError,
			},
			mockErr:        nil,
			expectedResult: nil,
		},
	}
}

func TestGetDevFaultInfo(t *testing.T) {
	tests := getDevFaultInfoTestCases()
	for _, tt := range tests {
		convey.Convey("test getDevFaultInfo", t, func() {
			manager := createFake910Manager()
			manager.hotResetManager = newTestHotResetManager(common.Ascend910, common.Train)
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyMethod(manager.hotResetManager, "GetGlobalDevFaultInfo",
				func(_ *HotResetTools, logicID int32) (*common.DevFaultInfo, error) {
					return tt.mockFaultInfo, tt.mockErr
				})
			result := manager.getDevFaultInfo(tt.logicID)
			convey.So(result, convey.ShouldResemble, tt.expectedResult)
		})
	}
}
