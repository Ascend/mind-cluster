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
		patch.ApplyPrivateMethod(manager, "canResetDevice", func(cardID, deviceID, logicId int32) bool {
			return true
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
			defer patch1.Reset()
			_, err := manager.canBeReset(nil)
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
		convey.Convey("01-get associated card error, should return error", func() {
			patch1 := gomonkey.ApplyFuncReturn(IsDevBusy, true)
			patch1.ApplyPrivateMethod(manager, "getAssociatedLogicIDs",
				func(logicID, cardID, deviceID int32) ([]int32, error) {
					return nil, testErr
				})
			defer patch1.Reset()
			err := manager.setUnhealthyForA3(devs)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-success, should return nil", func() {
			patch1 := gomonkey.ApplyFuncReturn(IsDevBusy, true)
			patch1.ApplyPrivateMethod(manager, "getAssociatedLogicIDs",
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
		mockGetCM := mockGetCM()
		defer mockGetCM.Reset()
		manager.hotResetManager = &HotResetTools{
			allTaskDevFaultInfo: map[string][]*common.TaskDevInfo{
				"task1": getTaskInfo(),
			},
			taskPod: map[string]v1.Pod{
				"task1": getSinglePod("pod1", map[string]string{}),
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
		defer mockGetCM.Reset()
		defer mockUpdateCM.Reset()
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
		err = manager.filterDevStatus(mockGroupDevice())
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
			Policy:  common.FreeResetError,
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
				func(logicID int32) (int32, error) {
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
		patch.ApplyFunc(WriteResetInfo, func(resetInfo ResetInfo, writeMode WriteMode) {
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
		mockGetCMFromCacheMethod := mockGetCMFromCache()
		defer mockGetCMFromCacheMethod.Reset()
		convey.So(manager.isDevShouldBeIsolate(chipPhyID3), convey.ShouldBeFalse)
	})
}

func mockGetCMFromCache() *gomonkey.Patches {
	nodeDeviceData := common.TaskResetInfo{
		UpdateTime: 11111111,
		RankList:   mockTaskDevInfoList(),
	}
	cm := &v1.ConfigMap{
		Data: map[string]string{common.ResetInfoCMDataKey: string(common.MarshalData(nodeDeviceData))},
	}
	mockFunc := gomonkey.ApplyMethodReturn(&HotResetTools{}, "GetCMFromCache", cm, nil)
	return mockFunc
}

// TestFilterDevStatusForA3 test the function filterDevStatusForA3
func TestFilterDevStatusForA3(t *testing.T) {
	manager := createFake910Manager()
	devs := []*common.NpuDevice{
		{LogicID: id1},
	}
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
		convey.Convey("03-get brother id fail, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBrotherCardID",
				int32(id1), testErr)
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
