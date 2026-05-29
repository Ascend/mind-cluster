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
	"context"
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
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
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
	ascend910LogicID0  = api.Ascend910MinuxPrefix + "0"
	ascend910LogicID1  = api.Ascend910MinuxPrefix + "1"
	ascend910LogicID2  = api.Ascend910MinuxPrefix + "2"
	ascend910LogicID3  = api.Ascend910MinuxPrefix + "3"
	ascend910LogicID4  = api.Ascend910MinuxPrefix + "4"
	ascend910LogicID5  = api.Ascend910MinuxPrefix + "5"
	ascend910LogicID6  = api.Ascend910MinuxPrefix + "6"
	ascend910LogicID7  = api.Ascend910MinuxPrefix + "7"
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

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestHwAscend910ManagerGetNPUs(t *testing.T) {
	convey.Convey("910 test GetNPUs", t, func() {
		manager := createFake910Manager()
		allInfo, err := manager.GetNPUs()
		convey.So(err, convey.ShouldBeNil)
		convey.So(allInfo.AllDevTypes[0], convey.ShouldEqual, api.Ascend910)
		convey.So(allInfo.AllDevs[0].DeviceName, convey.ShouldEqual,
			fmt.Sprintf("%s-%d", api.Ascend910, allInfo.AllDevs[0].PhyID))
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
		mockGetPodsUsedNpu := mockGetPodsUsedNpuByCommon()
		mockGetConfigMap := mockGetDeviceInfoCMCache(map[string]string{api.Ascend910: ascend910LogicID1})
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
		manager.DoWithVolcanoListAndWatch(groupDevice, 0)
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
			nodeDeviceData *common.NodeDeviceInfoCache, manuallySeparateNPU string,
			_ common.SwitchFaultInfo, dpuInfo common.DpuInfo, cm string) (*common.NodeDeviceInfoCache, error) {
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

func mockGetPodsUsedNpuByCommon() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
		"GetPodsUsedNpuByCommon", func(_ *kubeclient.ClientK8s) sets.String {
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
	common.ParamOption.RealCardType = api.Ascend910A
	convey.Convey("exec ut function GraceTolerance", t, func() {
		mockPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetAllPodList",
			func(_ *kubeclient.ClientK8s) (*v1.PodList, error) {
				return mockGetAllPodList(), nil
			})
		mockGetCM := mockGetCM()
		defer mockGetCM.Reset()
		defer mockPodList.Reset()
		patch := gomonkey.ApplyMethod(new(HotResetTools), "SyncResetCM",
			func(_ *HotResetTools, _ context.Context, _ *kubeclient.ClientK8s) { return })
		defer patch.Reset()
		manager.GraceTolerance(context.TODO(), mockGroupDevice())
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
			common.ParamOption.RealCardType = api.Ascend910B
			manager.hotResetTryOutBand(devs)
			convey.So(flag, convey.ShouldBeFalse)
		})
		convey.Convey("02-A3, flag should be true", func() {
			common.ParamOption.RealCardType = api.Ascend910A3
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
	manager.hotResetManager = newTestHotResetManager(api.Ascend910A, common.Ascend910BRingsNumTrain)
	convey.Convey("exec ut function canBeReset", t, func() {
		convey.Convey("A3 device can reset, should return true", func() {
			common.ParamOption.RealCardType = api.Ascend910A3
			patch1 := gomonkey.ApplyPrivateMethod(manager, "canA3BeReset",
				func(dev *common.DevFaultInfo) bool {
					return true
				})
			patch1.ApplyPrivateMethod(manager, "canResetDeviceByLogicID", func(logicID int32) bool {
				return true
			})
			defer patch1.Reset()
			_, err := manager.canBeReset(mockSingleDevFaultInfo(), mockOneEmptyPodList())
			convey.So(err, convey.ShouldBeNil)
		})
		common.ParamOption.RealCardType = api.Ascend910B
		resultBool, err := manager.canBeReset(mockSingleDevFaultInfo(), mockOneEmptyPodList())
		convey.So(resultBool, convey.ShouldBeTrue)
		convey.So(err, convey.ShouldBeNil)

		resultBool, err = manager.canBeReset(mockSingleDevFaultInfo(), mockGetAllPodList())
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
				int32(id1), int32(id1), ascend910testErr)
			defer patch1.Reset()
			ret := manager.canA3BeReset(dev, mockOneEmptyPodList())
			convey.So(ret, convey.ShouldBeFalse)
		})
		patch := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
			int32(id1), int32(id1), nil)
		defer patch.Reset()
		convey.Convey("02-get associated card error, should return false", func() {
			patch1 := gomonkey.ApplyMethod(manager, "GetAssociatedLogicIDs",
				func(_ *HwAscend910Manager, logicID, cardID, deviceID int32) ([]int32, error) {
					return nil, ascend910testErr
				})
			defer patch1.Reset()
			ret := manager.canA3BeReset(dev, mockOneEmptyPodList())
			convey.So(ret, convey.ShouldBeFalse)
		})
		patch.ApplyMethod(manager, "GetAssociatedLogicIDs",
			func(_ *HwAscend910Manager, logicID, cardID, deviceID int32) ([]int32, error) {
				return []int32{id1}, nil
			})
		patch.ApplyMethodReturn(&kubeclient.ClientK8s{}, "GetAllPodList", nil, nil)
		patch.ApplyPrivateMethod(manager, "getBusyChipListFromPod",
			func(podList *v1.PodList) []string {
				return []string{}
			})
		convey.Convey("03-get chip active error, should return false", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "isChipActive",
				func(logicID int32, busyChipList []string) (bool, error) {
					return false, ascend910testErr
				})
			defer patch1.Reset()
			ret := manager.canA3BeReset(dev, mockOneEmptyPodList())
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
		patch.ApplyPrivateMethod(manager, "GetAssociatedLogicIDs",
			func(_ *HwAscend910Manager, logicID, cardID, deviceID int32) ([]int32, error) {
				return []int32{id1}, nil
			})
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
			ret := manager.canA3BeReset(dev, mockOneEmptyPodList())
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("06-success, should return true", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "isChipActive",
				func(logicID int32, busyChipList []string) (bool, error) {
					return true, nil
				})
			defer patch1.Reset()
			ret := manager.canA3BeReset(dev, mockOneEmptyPodList())
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
	manager.hotResetManager = newTestHotResetManager(api.Ascend910A, common.Ascend910BRingsNumTrain)
	devInfo := mockSingleDevFaultInfo()
	common.ParamOption.RealCardType = api.Ascend910B
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
		err := manager.execHotReset(nil, devInfo)
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
		devStatusList := devList[api.Ascend910]
		manager.hotResetManager = &HotResetTools{
			resetDevNumOnce: 8,
		}
		inResetDev = -1
		common.ParamOption.RealCardType = api.Ascend910B

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
			common.ParamOption.RealCardType = api.Ascend910A3
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
			patch1.ApplyPrivateMethod(manager, "GetAssociatedLogicIDs",
				func(_ *HwAscend910Manager, logicID, cardID, deviceID int32) ([]int32, error) {
					return nil, ascend910testErr
				})
			defer patch1.Reset()
			err := manager.setUnhealthyForA3(devs)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-success, should return nil", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "GetAssociatedLogicIDs",
				func(_ *HwAscend910Manager, logicID, cardID, deviceID int32) ([]int32, error) {
					return []int32{logicID}, nil
				})
			defer patch1.Reset()
			err := manager.setUnhealthyForA3(devs)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetAssociatedLogicIDs test the function GetAssociatedLogicIDs
func TestGetAssociatedLogicIDs(t *testing.T) {
	manager := createFake910Manager()
	const id1int32 = int32(id1)
	convey.Convey("test GetAssociatedLogicIDs", t, func() {
		convey.Convey("01-get brother card error, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBrotherCardID",
				id1int32, ascend910testErr)
			defer patch1.Reset()
			_, err := manager.GetAssociatedLogicIDs(id1int32, id1int32, id1int32)
			convey.So(err, convey.ShouldBeError)
		})
		patch := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBrotherCardID",
			id1int32, nil)
		defer patch.Reset()
		convey.Convey("02-get logic id failed, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDeviceLogicID",
				id1int32, ascend910testErr)
			defer patch1.Reset()
			_, err := manager.GetAssociatedLogicIDs(id1int32, id1int32, id1int32)
			convey.So(err, convey.ShouldBeError)
		})
		patch.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDeviceLogicID",
			id1int32, nil)
		convey.Convey("03-success, should return nil", func() {
			_, err := manager.GetAssociatedLogicIDs(id1int32, id1int32, id1int32)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
func TestIsRingResetComplete(t *testing.T) {
	manager := createFake910Manager()
	manager.hotResetManager = newTestHotResetManager(api.Ascend910A, common.Ascend910BRingsNumTrain)
	common.ParamOption.RealCardType = api.Ascend910B
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
		common.ParamOption.RealCardType = api.Ascend910
		err = manager.isRingResetComplete(logicID, false)
		convey.So(err, convey.ShouldBeNil)
	})
}
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
		devices[api.Ascend910][chipPhyID1].Health = v1beta1.Unhealthy
		devices[api.Ascend910][chipPhyID3].Health = v1beta1.Unhealthy
		devices[api.Ascend910][chipPhyID5].Health = v1beta1.Unhealthy
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
		api.HuaweiAscend910: api.Ascend910 + "-0," + api.Ascend910 + "-1",
	}
	annotationEmpty := map[string]string{
		api.HuaweiAscend910: "",
	}
	annotationErr := map[string]string{}
	annotationErrRank := map[string]string{
		common.ResetTaskNameKey: "task1",
	}
	annotationSuccess := map[string]string{
		common.ResetTaskNameKey: "task1",
		api.PodRankIndexAnno:    "1",
		api.HuaweiAscend910: api.Ascend910 + "-4," + api.Ascend910 +
			"-5," + api.Ascend910 + "-6," + api.Ascend910 + "-7",
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
		api.HuaweiAscend910: "",
	}
	return &v1.PodList{
		Items: []v1.Pod{
			getSinglePod("test-pod2", annotationEmpty),
		},
	}
}

func mockGroupDevice() map[string][]*common.NpuDevice {
	return map[string][]*common.NpuDevice{
		api.Ascend910: mockNpuDevices(),
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
		DevType:       api.Ascend910,
		DeviceName:    fmt.Sprintf("%s-%d", api.Ascend910, autoID),
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
func TestExecRescan(t *testing.T) {
	manager := createFake910Manager()
	devs := []ResetDevice{
		{CardId: int32(id1), DeviceId: int32(id2)},
	}
	patch := gomonkey.ApplyFunc(WriteResetInfo, func(resetInfo ResetInfo, writeMode WriteMode, update bool) {
		return
	})
	flag := false
	patch.ApplyFunc(FreeBusyDev, func(logicID int32) {
		flag = true
	})
	defer patch.Reset()
	convey.Convey("test execRescan", t, func() {
		convey.Convey("01-rescan error, flag should be false", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "RescanSoc", ascend910testErr)
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

// TestIsNeedBlockAllDevice ut for method isNeedBlockAllDevice,using new board id
func TestIsNeedBlockAllDevice(t *testing.T) {

	GetServerBoardIdPatch := mockGetServerBoardId(A800IA2WithHccs)
	defer GetServerBoardIdPatch.Reset()

	doTestIsNeedBlockAllDevice(t)
}

// TestIsNeedBlockAllDevice ut for method isNeedBlockAllDevice,using old board id
func TestIsNeedBlockAllDeviceUsOldBoardId(t *testing.T) {

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
	GetServerBoardIdPatch := mockGetServerBoardId(common.A800IA2NoneHccsBoardId)
	defer GetServerBoardIdPatch.Reset()

	doTestNoNeedToBlock(t)
}

// TestNoNeedToBlock test need block all device with none hccs A800IA2,using old board id
func TestNoNeedToBlockUsingOldId(t *testing.T) {
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
	GetServerBoardIdPatch := mockGetServerBoardId(A800IA2WithHccs)
	defer GetServerBoardIdPatch.Reset()

	doTestNodeNeedToBlockWithNotHandleErr(t)
}

func TestNodeNeedToBlockWithNotHandleErrUsingOldId(t *testing.T) {
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
			api.Ascend910: {{LogicID: chipPhyID0, Health: ""}},
		}
		npuDev := &common.NpuDevice{LogicID: chipPhyID0}
		devInfo := &common.DevFaultInfo{LogicId: chipPhyID0}
		manager.handleResetProcess(classifyDevs, devInfo, npuDev)
		convey.So(classifyDevs[api.Ascend910][0].Health, convey.ShouldEqual, v1beta1.Healthy)
		convey.So(isolateDevList, convey.ShouldResemble, []int32{chipPhyID0})
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
			patch1 := gomonkey.ApplyFuncReturn((*HwAscend910Manager).GetAssociatedLogicIDs, []int32{}, nil)
			defer patch1.Reset()
			retId, err := manager.getResetIndexForA3(chipPhyID0)
			convey.So(err.Error(), convey.ShouldEqual, "sort logic ids failed, logic ids [], sorted ids []")
			convey.So(retId, convey.ShouldEqual, errorId)
		})

		convey.Convey("03-get reset index success, should return nil", func() {
			patch1 := gomonkey.ApplyFuncReturn((*HwAscend910Manager).GetAssociatedLogicIDs, []int32{id1, id2, id3}, nil)
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
			api.Ascend910: {
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
			convey.So(classifyDevs[api.Ascend910][0].Health, convey.ShouldBeEmpty)
			convey.So(classifyDevs[api.Ascend910][3].Health, convey.ShouldEqual, v1beta1.Unhealthy)
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
			convey.So(classifyDevs[api.Ascend910][0].Health, convey.ShouldBeEmpty)
			convey.So(classifyDevs[api.Ascend910][1].Health, convey.ShouldBeEmpty)
			convey.So(classifyDevs[api.Ascend910][2].Health, convey.ShouldBeEmpty)
			convey.So(classifyDevs[api.Ascend910][3].Health, convey.ShouldEqual, v1beta1.Healthy)
			convey.So(classifyDevs[api.Ascend910][4].Health, convey.ShouldEqual, v1beta1.Unhealthy)
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

// TestUpdateHotResetCache for test updateHotResetCache
func TestUpdateHotResetCache(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test updateHotResetCache", t, func() {
		manager.hotResetManager = &HotResetTools{}
		patch := gomonkey.ApplyGlobalVar(&isolateDevList, []int32{})
		defer patch.Reset()
		convey.Convey("01-classifyDevs is empty, should return error", func() {
			err := manager.updateHotResetCache(nil)
			convey.So(err.Error(), convey.ShouldEqual, "ascend npu device list not found")
		})
		convey.Convey("02-Ascend910 dev is empty, should return error", func() {
			classifyDevs := map[string][]*common.NpuDevice{api.Ascend910: {}}
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
				api.Ascend910: {
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
			convey.So(err.Error(), convey.ShouldEqual, "no Ascend npu device found in cache")
		})
		convey.Convey("03-update success, should return nil", func() {
			classifyDevs := map[string][]*common.NpuDevice{
				api.Ascend910: {
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
		api.HuaweiAscend910: "A1N0#4,A1N0#5,A1N0#6",
	}
	annotationSuccess := map[string]string{
		api.HuaweiAscend910: api.Ascend910 + "-4," + api.Ascend910 + "-5," + api.Ascend910 + "-6",
	}
	annotationSuccess1 := map[string]string{
		api.HuaweiAscend910:  api.Ascend910 + "-4," + api.Ascend910 + "-5," + api.Ascend910 + "-6",
		api.PodRankIndexAnno: "2",
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
		api.HuaweiAscend910:  api.Ascend910 + "-4," + api.Ascend910 + "-5," + api.Ascend910 + "-6," + api.Ascend910 + "-7",
		api.PodRankIndexAnno: "1.2",
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
			ret := manager.isReSchedulingScene(4)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("03-is rescheduling scene, should return true", func() {
			ret := manager.isReSchedulingScene(3)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

func TestCanResetDevice(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test canResetDevice", t, func() {
		convey.Convey("01-dev busy, should return false", func() {
			patch1 := gomonkey.ApplyFuncReturn(IsDevBusy, true)
			defer patch1.Reset()
			convey.So(manager.canResetDevice(id1), convey.ShouldBeFalse)
		})
		patch := gomonkey.ApplyFuncReturn(IsDevBusy, false)
		defer patch.Reset()
		convey.Convey("02-reset cnt over, should return false", func() {
			patch1 := gomonkey.ApplyFuncReturn(GetResetCnt, common.MaxResetTimes+id1)
			defer patch1.Reset()
			convey.So(manager.canResetDevice(id1), convey.ShouldBeFalse)
		})
		patch.ApplyFuncReturn(GetResetCnt, common.MaxResetTimes-id1)
		convey.Convey("03-success, should return true", func() {
			convey.So(manager.canResetDevice(id1), convey.ShouldBeTrue)
		})
	})
}

// TestExecOutBandReset test the function execOutBandReset
func TestExecOutBandReset(t *testing.T) {
	manager := createFake910Manager()
	const testCardID, testDeviceID, sleepTime = 0, 0, 50 * time.Millisecond
	mockAddAnnotation := gomonkey.ApplyMethod(
		&kubeclient.ClientK8s{}, "AddAnnotation",
		func(_ *kubeclient.ClientK8s, key, value string) error {
			return nil
		})
	defer mockAddAnnotation.Reset()
	convey.Convey("test execOutBandReset", t, func() {
		patch := gomonkey.ApplyPrivateMethod(manager, "updateResetInfo",
			func(_ *HwAscend910Manager, failDevs, sucDevs []ResetDevice) {
				return
			}).
			ApplyPrivateMethod(manager, "scanDeviceForThirdParty",
				func(_ *HwAscend910Manager, failDevs []ResetDevice) {
					return
				}).
			ApplyPrivateMethod(manager, "fillResetDevs",
				func(_ *HwAscend910Manager, devs []ResetDevice) ([]ResetDevice, error) {
					return devs, nil
				})
		defer patch.Reset()
		common.ParamOption.RealCardType = api.Ascend910A3
		convey.Convey("01-reset error, should return error", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "resetDeviceOutBand",
				func(_ *HwAscend910Manager, cardId, deviceId int32) error {
					return ascend910testErr
				})
			defer patch1.Reset()
			err := manager.execOutBandReset([]ResetDevice{
				{CardId: testCardID, DeviceId: testDeviceID},
			}, []ResetDevice{})
			convey.So(err, convey.ShouldBeError)
		})
		patch.ApplyPrivateMethod(manager, "resetDeviceOutBand",
			func(_ *HwAscend910Manager, cardId, deviceId int32) error {
				return nil
			})
		patch.ApplyPrivateMethod(manager, "isRingResetComplete",
			func(_ *HwAscend910Manager, oriLogicID int32, shouldCheckNet bool) error {
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
			common.ParamOption.RealCardType = api.Ascend910A3
			manager.updateResetInfo(testDevs, []ResetDevice{})
			convey.So(ri.ThirdPartyResetDevs, convey.ShouldResemble, testDevs)
		})
		convey.Convey("02-not A3, should append to manual devices", func() {
			common.ParamOption.RealCardType = api.Ascend910B
			manager.updateResetInfo(testDevs, []ResetDevice{})
			convey.So(ri.ManualResetDevs, convey.ShouldResemble, testDevs)
		})
	})
}

// TestResetDeviceOutBand test the function resetDeviceOutBand
func TestResetDeviceOutBand(t *testing.T) {
	manager := createFake910Manager()
	convey.Convey("test resetDeviceOutBand", t, func() {
		const testLogicID = 0
		convey.Convey("01-out band channel error, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
				"GetOutBandChannelState", ascend910testErr)
			defer patch1.Reset()
			err := manager.resetDeviceOutBand(testLogicID)
			convey.So(err, convey.ShouldBeError)
		})
		patch := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetOutBandChannelState", nil)
		defer patch.Reset()
		convey.Convey("02-pre reset error, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
				"PreResetSoc", ascend910testErr)
			defer patch1.Reset()
			err := manager.resetDeviceOutBand(testLogicID)
			convey.So(err, convey.ShouldBeError)
		})
		patch.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"PreResetSoc", nil)
		convey.Convey("03-reset out band error, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
				"SetDeviceResetOutBand", ascend910testErr)
			defer patch1.Reset()
			err := manager.resetDeviceOutBand(testLogicID)
			convey.So(err, convey.ShouldBeError)
		})
		patch.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"SetDeviceResetOutBand", nil)
		convey.Convey("04-rescan error, should return error", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
				"RescanSoc", ascend910testErr)
			defer patch1.Reset()
			err := manager.resetDeviceOutBand(testLogicID)
			convey.So(err, convey.ShouldBeError)
		})
		patch.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"RescanSoc", nil)
		convey.Convey("05-success, should return nil", func() {
			err := manager.resetDeviceOutBand(testLogicID)
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
			patch1.ApplyPrivateMethod(manager, "GetAssociatedLogicIDs",
				func(_ *HwAscend910Manager, logicID, cardID, deviceID int32) ([]int32, error) {
					return []int32{id1}, ascend910testErr
				})
			defer patch1.Reset()
			err := manager.filterDevStatusForA3(devs)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-success, should return nil", func() {
			patch1 := gomonkey.ApplyFuncReturn(IsDevBusy, false)
			patch1.ApplyPrivateMethod(manager, "GetAssociatedLogicIDs",
				func(_ *HwAscend910Manager, logicID, cardID, deviceID int32) ([]int32, error) {
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
	common.ParamOption.RealCardType = api.Ascend910A3
	convey.Convey("test fillResetDevs", t, func() {
		convey.Convey("01-get npu error, should return err", func() {
			patch1 := gomonkey.ApplyMethodReturn(manager, "GetNPUs",
				common.NpuAllInfo{}, ascend910testErr)
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
		convey.Convey("01-IsDevBusy is true, should return false", func() {
			patch1 := gomonkey.ApplyFuncReturn(IsDevBusy, true)
			defer patch1.Reset()
			convey.So(manager.canResetDeviceByLogicID(int32(id1)), convey.ShouldBeFalse)
		})
		convey.Convey("02-IsDevBusy is false, should return true", func() {
			patch1 := gomonkey.ApplyFuncReturn(IsDevBusy, false)
			defer patch1.Reset()
			convey.So(manager.canResetDeviceByLogicID(int32(id1)), convey.ShouldBeTrue)
		})
	})
}

func TestGetResetIndex(t *testing.T) {
	convey.Convey("test getResetIndex", t, func() {
		manager := createFake910Manager()
		manager.hotResetManager = newTestHotResetManager(api.Ascend910A, common.Ascend910BRingsNumTrain)
		dev := &common.NpuDevice{}
		convey.Convey("01-A3, get idx success, should return nil", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager, "getResetIndexForA3",
				func(logicID int32) (int32, error) {
					return int32(id1), nil
				})
			defer patch1.Reset()
			_, err := manager.getResetIndex(dev.LogicID)
			convey.So(err, convey.ShouldBeNil)
		})
		patch := gomonkey.ApplyPrivateMethod(manager, "getResetIndexForA3",
			func(logicID int32) (int32, error) {
				return int32(id1), ascend910testErr
			})
		defer patch.Reset()
		convey.Convey("02-get dev num once err, should return err", func() {
			patch1 := gomonkey.ApplyPrivateMethod(manager.hotResetManager, "GetResetDevNumOnce",
				func() (int, error) {
					return id1, ascend910testErr
				})
			defer patch1.Reset()
			_, err := manager.getResetIndex(dev.LogicID)
			convey.So(err, convey.ShouldBeError)
		})
		patch.ApplyPrivateMethod(manager.hotResetManager, "GetResetDevNumOnce",
			func() (int, error) {
				return id1, nil
			})
		convey.Convey("03-success, should return nil", func() {
			_, err := manager.getResetIndex(dev.LogicID)
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
			mockErr:        ascend910testErr,
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
			manager.hotResetManager = newTestHotResetManager(api.Ascend910A, common.Ascend910BRingsNumTrain)
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

func TestIsFaultNeedRestart(t *testing.T) {
	convey.Convey("Test isFaultNeedRestart", t, func() {
		hnm := &HwAscend910Manager{}
		common.ParamOption.RealCardType = api.Ascend910A3
		convey.Convey("when policy is RestartError", func() {
			devFaultInfo := &common.DevFaultInfo{
				LogicId: 1,
				Policy:  common.RestartError,
			}

			convey.Convey("should store fault time when first fault", func() {
				resetTimeMap.Delete(devFaultInfo.LogicId)
				result := hnm.isFaultNeedRestart(devFaultInfo)
				convey.So(result, convey.ShouldBeFalse)
				faultTime, exists := resetTimeMap.Load(devFaultInfo.LogicId)
				convey.So(exists, convey.ShouldBeTrue)
				convey.So(faultTime, convey.ShouldNotBeZeroValue)
			})

			convey.Convey("should return true and delete map when timeout", func() {
				oldTime := time.Now().Unix() - common.ResetFaultToleranceTimeInterval - 1
				resetTimeMap.Store(devFaultInfo.LogicId, oldTime)
				result := hnm.isFaultNeedRestart(devFaultInfo)
				convey.So(result, convey.ShouldBeTrue)
				_, exists := resetTimeMap.Load(devFaultInfo.LogicId)
				convey.So(exists, convey.ShouldBeFalse)
			})

			convey.Convey("should return false when within tolerance time", func() {
				recentTime := time.Now().Unix() - common.ResetFaultToleranceTimeInterval + common.BaseDec
				resetTimeMap.Store(devFaultInfo.LogicId, recentTime)
				result := hnm.isFaultNeedRestart(devFaultInfo)
				convey.So(result, convey.ShouldBeFalse)
				faultTime, exists := resetTimeMap.Load(devFaultInfo.LogicId)
				convey.So(exists, convey.ShouldBeTrue)
				convey.So(faultTime, convey.ShouldEqual, recentTime)
			})

		})
	})
}

func TestIsFaultNeedRestart2(t *testing.T) {
	convey.Convey("Test isFaultNeedRestart", t, func() {
		hnm := &HwAscend910Manager{}
		common.ParamOption.RealCardType = api.Ascend910A3
		convey.Convey("when policy is RestartError", func() {
			devFaultInfo := &common.DevFaultInfo{LogicId: 1, Policy: common.RestartRequestError}
			convey.Convey("should return false when within tolerance time and policy is RestartRequestError", func() {
				recentTime := time.Now().Unix() - common.ResetFaultToleranceTimeInterval + common.BaseDec
				resetTimeMap.Store(devFaultInfo.LogicId, recentTime)
				result := hnm.isFaultNeedRestart(devFaultInfo)
				convey.So(result, convey.ShouldBeFalse)
				faultTime, exists := resetTimeMap.Load(devFaultInfo.LogicId)
				convey.So(exists, convey.ShouldBeTrue)
				convey.So(faultTime, convey.ShouldEqual, recentTime)
			})
			devFaultInfo = &common.DevFaultInfo{LogicId: 1, Policy: common.FreeResetError}
			convey.Convey("should return false when within tolerance time and policy is FreeResetError", func() {
				recentTime := time.Now().Unix() - common.ResetFaultToleranceTimeInterval + common.BaseDec
				resetTimeMap.Store(devFaultInfo.LogicId, recentTime)
				result := hnm.isFaultNeedRestart(devFaultInfo)
				convey.So(result, convey.ShouldBeFalse)
				faultTime, exists := resetTimeMap.Load(devFaultInfo.LogicId)
				convey.So(exists, convey.ShouldBeTrue)
				convey.So(faultTime, convey.ShouldEqual, recentTime)
			})
			devFaultInfo = &common.DevFaultInfo{LogicId: 1, Policy: common.ResetError}
			convey.Convey("should return false when within tolerance time and policy is reset", func() {
				recentTime := time.Now().Unix() - common.ResetFaultToleranceTimeInterval + common.BaseDec
				resetTimeMap.Store(devFaultInfo.LogicId, recentTime)
				result := hnm.isFaultNeedRestart(devFaultInfo)
				convey.So(result, convey.ShouldBeFalse)
				faultTime, exists := resetTimeMap.Load(devFaultInfo.LogicId)
				convey.So(exists, convey.ShouldBeTrue)
				convey.So(faultTime, convey.ShouldEqual, recentTime)
			})
			convey.Convey("should return true when device type is A2 and policy is reset", func() {
				recentTime := time.Now().Unix() - common.ResetFaultToleranceTimeInterval + common.BaseDec
				resetTimeMap.Store(devFaultInfo.LogicId, recentTime)
				common.ParamOption.RealCardType = api.Ascend910B
				result := hnm.isFaultNeedRestart(devFaultInfo)
				convey.So(result, convey.ShouldBeTrue)
			})
		})
	})
}

// TestCheckFaultIsExist tests the checkFaultIsExist function
func TestCheckFaultIsExist(t *testing.T) {
	manager := createFake910Manager()
	manager.hotResetManager = &HotResetTools{}

	convey.Convey("Test checkFaultIsExist", t, func() {
		convey.Convey("01-no ascend 910 device, should return true", func() {
			classifyDevs := map[string][]*common.NpuDevice{}
			result := manager.checkFaultIsExist(classifyDevs, 0)
			convey.So(result, convey.ShouldBeTrue)
		})
		convey.Convey("02-get reset index failed, should return true", func() {
			patch := gomonkey.ApplyPrivateMethod(manager, "getResetIndex",
				func(dev *common.NpuDevice) (int32, error) {
					return -1, errors.New("fakeError")
				})
			defer patch.Reset()
			classifyDevs := map[string][]*common.NpuDevice{
				api.Ascend910: {
					{LogicID: 0},
				},
			}
			result := manager.checkFaultIsExist(classifyDevs, 0)
			convey.So(result, convey.ShouldBeTrue)
		})
		convey.Convey("03-fault device exists in the same ring, should return true", func() {
			patch := gomonkey.ApplyPrivateMethod(manager, "getResetIndex",
				func(dev *common.NpuDevice) (int32, error) {
					return 0, nil
				})
			defer patch.Reset()
			patch.ApplyMethodReturn(manager.dmgr, "GetDeviceAllErrorCode", int32(0),
				[]int64{common.CardDropFaultCode}, nil)
			defer patch.Reset()
			patch.ApplyMethodReturn(manager.hotResetManager, "GetDevProcessPolicy", common.RestartRequestError)
			defer patch.Reset()
			classifyDevs := map[string][]*common.NpuDevice{
				api.Ascend910: {
					{LogicID: 0},
					{LogicID: 1},
				},
			}
			result := manager.checkFaultIsExist(classifyDevs, 0)
			convey.So(result, convey.ShouldBeTrue)
		})

	})
}

// TestTryResetDeviceOffline tests the tryResetDeviceOffline function
func TestTryResetDeviceOffline(t *testing.T) {
	convey.Convey("Test tryResetDeviceOffline", t, func() {
		manager := createFake910Manager()
		manager.hotResetManager = &HotResetTools{}
		logicId := int32(0)
		classifyDevs := map[string][]*common.NpuDevice{api.Ascend910: {{LogicID: 0}, {LogicID: 1}}}
		convey.Convey("01-First fault exists, second fault disappears, should return nil", func() {
			callCount := 0
			// Mock checkFaultIsExist to return true on first call and false on second call
			patch1 := gomonkey.ApplyPrivateMethod(manager, "checkFaultIsExist",
				func(_ *HwAscend910Manager, devs map[string][]*common.NpuDevice, logicID int32) bool {
					callCount++
					if callCount == 1 {
						return true
					} // First call - fault exists
					return false // Second call - fault disappears
				})
			defer patch1.Reset()
			// Mock SetDeviceReset to return nil (success) on first call
			patch2 := gomonkey.ApplyMethodReturn(manager, "GetDmgr", manager.dmgr)
			defer patch2.Reset()
			patch3 := gomonkey.ApplyMethod(manager.dmgr, "SetDeviceReset",
				func(_ *devmanager.DeviceManagerMock, logicID int32) error {
					if callCount == 1 {
						return errors.New("fakeError") // First call - fault exists
					}
					return nil // Second call - fault disappears
				})
			defer patch3.Reset()
			patch4 := gomonkey.ApplyFunc(AddResetCnt, func(logicID int32) {})
			defer patch4.Reset()
			patch5 := gomonkey.ApplyFunc(AddBusyDev, func(logicID int32) {})
			defer patch5.Reset()
			err := manager.tryResetDeviceOffline(classifyDevs, logicId)
			convey.So(err, convey.ShouldBeNil)
			convey.So(callCount, convey.ShouldEqual, common.MapSizeTwo)
		})
	})
}

// TestTryResetDeviceOffline tests the tryResetDeviceOffline function
func TestTryResetDeviceOffline2(t *testing.T) {
	convey.Convey("Test tryResetDeviceOffline", t, func() {
		manager := createFake910Manager()
		manager.hotResetManager = &HotResetTools{}
		logicId := int32(0)
		classifyDevs := map[string][]*common.NpuDevice{api.Ascend910: {{LogicID: 0}, {LogicID: 1}}}
		convey.Convey("02-First fault does not exist, should return immediately", func() {
			// Mock checkFaultIsExist to return false on first call
			patch1 := gomonkey.ApplyPrivateMethod(manager, "checkFaultIsExist",
				func(_ *HwAscend910Manager, devs map[string][]*common.NpuDevice, logicID int32) bool {
					return false
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(AddResetCnt, func(logicID int32) {})
			defer patch2.Reset()
			patch3 := gomonkey.ApplyFunc(AddBusyDev, func(logicID int32) {})
			defer patch3.Reset()
			err := manager.tryResetDeviceOffline(classifyDevs, logicId)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetResetTime(t *testing.T) {
	convey.Convey("test getResetTime", t, func() {
		convey.Convey("01-not exist in map, should return 0", func() {
			result := getResetTime(999)
			convey.So(result, convey.ShouldEqual, int64(0))
		})
		convey.Convey("02-exist in map, should return value", func() {
			resetTimeMap.Store(int32(0), int64(12345))
			defer resetTimeMap.Delete(int32(0))
			result := getResetTime(0)
			convey.So(result, convey.ShouldEqual, int64(12345))
		})
	})
}

func TestGetAscend910Name(t *testing.T) {
	convey.Convey("test getAscend910Name", t, func() {
		convey.Convey("01-A5 card type, should return NPULowerCase", func() {
			origType := common.ParamOption.RealCardType
			common.ParamOption.RealCardType = api.Ascend910A5
			defer func() { common.ParamOption.RealCardType = origType }()
			result := getAscend910Name()
			convey.So(result, convey.ShouldEqual, api.NPULowerCase)
		})
		convey.Convey("02-non-A5 card type, should return Ascend910", func() {
			origType := common.ParamOption.RealCardType
			common.ParamOption.RealCardType = api.Ascend910A3
			defer func() { common.ParamOption.RealCardType = origType }()
			result := getAscend910Name()
			convey.So(result, convey.ShouldEqual, api.Ascend910)
		})
	})
}

func TestNewHwAscend910Manager(t *testing.T) {
	convey.Convey("test NewHwAscend910Manager", t, func() {
		convey.Convey("01-should create manager with correct fields", func() {
			mgr := NewHwAscend910Manager()
			convey.So(mgr, convey.ShouldNotBeNil)
			convey.So(mgr.devCount, convey.ShouldEqual, common.MaxDevicesNum)
		})
	})
}

func TestClearDeviceStatus(t *testing.T) {
	convey.Convey("test clearDeviceStatus", t, func() {
		convey.Convey("01-should reset all device status to normal", func() {
			devList := []*common.NpuDevice{
				{LogicID: 0, Status: common.NPUResettingStatus},
				{LogicID: 1, Status: common.NPUResettingStatus},
			}
			clearDeviceStatus(devList)
			for _, dev := range devList {
				convey.So(dev.Status, convey.ShouldEqual, common.NPUNormalStatus)
			}
		})
	})
}

func TestStartUpHotReset(t *testing.T) {
	convey.Convey("test startUpHotReset", t, func() {
		manager := createFake910Manager()
		classifyDevs := map[string][]*common.NpuDevice{api.Ascend910: {{LogicID: 0}}}
		devFaultInfo := &common.DevFaultInfo{LogicId: 0, Policy: common.ResetError}
		npuDev := &common.NpuDevice{LogicID: 0}
		convey.Convey("01-should set inResetDev and call handleResetProcess", func() {
			patch := gomonkey.ApplyPrivateMethod(manager, "handleResetProcess",
				func(_ *HwAscend910Manager, _ map[string][]*common.NpuDevice,
					_ *common.DevFaultInfo, _ *common.NpuDevice) {
				})
			defer patch.Reset()
			err := manager.startUpHotReset(classifyDevs, devFaultInfo, npuDev)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestScanDeviceForThirdParty(t *testing.T) {
	convey.Convey("test scanDeviceForThirdParty", t, func() {
		manager := createFake910Manager()
		convey.Convey("01-empty failDevs, should return immediately", func() {
			manager.scanDeviceForThirdParty(nil)
		})
		convey.Convey("02-with failDevs, should schedule execRescan", func() {
			patch := gomonkey.ApplyFunc(WriteResetInfo,
				func(resetInfo ResetInfo, writeMode WriteMode, update bool) {})
			defer patch.Reset()
			origDelay := common.ParamOption.ThirdPartyScanDelay
			common.ParamOption.ThirdPartyScanDelay = 0
			defer func() { common.ParamOption.ThirdPartyScanDelay = origDelay }()
			failDevs := []ResetDevice{{LogicID: 0}}
			manager.scanDeviceForThirdParty(failDevs)
			time.Sleep(100 * time.Millisecond)
		})
	})
}

func TestIsShouldCheckNet(t *testing.T) {
	convey.Convey("test isShouldCheckNet", t, func() {
		manager := createFake910Manager()
		convey.Convey("01-should return result of isRunningDistributed", func() {
			patch := gomonkey.ApplyPrivateMethod(manager, "isRunningDistributed",
				func(_ *HwAscend910Manager, _ int32) bool { return true })
			defer patch.Reset()
			result := manager.isShouldCheckNet(0)
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

func TestSetDpu(t *testing.T) {
	convey.Convey("test SetDpu", t, func() {
		manager := createFake910Manager()
		convey.Convey("01-should set dpu info correctly", func() {
			dpuList := []common.DpuCMData{{Name: "dpu0"}}
			npuToDpusMap := map[string][]string{"npu0": {"dpu0"}}
			manager.SetDpu("pcie", dpuList, npuToDpusMap)
			convey.So(manager.dpu.BusType, convey.ShouldEqual, "pcie")
			convey.So(manager.dpu.DPUList, convey.ShouldHaveLength, 1)
			convey.So(manager.dpu.NpuToDpusMap, convey.ShouldHaveLength, 1)
			convey.So(manager.dpu.UpdateTime, convey.ShouldBeGreaterThan, 0)
		})
	})
}
