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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/api"
	"ascend-common/devmanager"
	npuCommon "ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"
)

const (
	phyIDNum    = 1
	logicIDNum  = 2
	vDevIDNum   = 3
	aiCoreNum   = 4
	aiCoreCount = 8
	vDevChipID  = 100
	FaultOnce   = 1
	NoneFault   = 0

	atlas300VPro     = "Atlas 300V Pro"
	ascend910FakeID0 = "Ascend910-0"
	ascend910FakeID1 = "Ascend910-1"
	ascend910FakeID2 = "Ascend910-2"
)

func deepCopyGroupDevice(groupDevice map[string][]*common.NpuDevice) map[string][]*common.NpuDevice {
	newGroupDevice := make(map[string][]*common.NpuDevice, len(groupDevice))
	for deviceType, npuDevices := range groupDevice {
		newNpuDevices := make([]*common.NpuDevice, 0, len(npuDevices))
		for _, npuDevice := range npuDevices {
			newNpuDevice := &common.NpuDevice{
				FaultCodes:             npuDevice.FaultCodes,
				AlarmRaisedTime:        npuDevice.AlarmRaisedTime,
				NetworkFaultCodes:      npuDevice.NetworkFaultCodes,
				NetworkAlarmRaisedTime: npuDevice.NetworkAlarmRaisedTime,
				DevType:                npuDevice.DevType,
				DeviceName:             npuDevice.DeviceName,
				Health:                 npuDevice.Health,
				NetworkHealth:          npuDevice.NetworkHealth,
				IP:                     npuDevice.IP,
				LogicID:                npuDevice.LogicID,
				PhyID:                  npuDevice.PhyID,
				CardID:                 npuDevice.CardID,
			}
			newNpuDevices = append(newNpuDevices, newNpuDevice)
		}
		newGroupDevice[deviceType] = newNpuDevices
	}
	return newGroupDevice
}

// TestGetChipAICore for test GetChipAICore
func TestGetChipAICore(t *testing.T) {
	convey.Convey("test GetChipAICore", t, func() {
		// 01-stub ParamOption, get chip ai core success, should be equal to coreCnt
		coreCnt := int32(8)
		tool := mockAscendTools()
		mockOption := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{AiCoreCount: coreCnt})
		defer mockOption.Reset()
		convey.So(tool.GetChipAICore() == coreCnt, convey.ShouldBeTrue)
	})
}

// TestGetName for test GetName
func TestGetName(t *testing.T) {
	convey.Convey("test GetName", t, func() {
		// 01-get ascend tools name success, should be equal to toolName
		tool := mockAscendTools()
		toolName := "mock tool"
		tool.name = toolName
		convey.So(tool.GetName() == toolName, convey.ShouldBeTrue)
	})
}

// TestConvertLogicIDsToDeviceNames for test convertLogicIDsToDeviceNames
func TestConvertLogicIDsToDeviceNames(t *testing.T) {
	convey.Convey("test convertLogicIDsToDeviceNames", t, func() {
		convey.Convey("01-get device run mode failed, should return empty string", func() {
			tool := mockAscendTools()
			mockOption := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{RealCardType: ""})
			defer mockOption.Reset()
			logicIds := []int32{3}
			convey.So(tool.convertLogicIDsToDeviceNames(logicIds), convey.ShouldEqual, "")
		})
		convey.Convey("02-convert logic ids to device names success, should return Ascend910-1", func() {
			tool := mockAscendTools()
			mockOption := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{RealCardType: api.Ascend910A3})
			defer mockOption.Reset()
			logicIds := []int32{3}
			convey.So(tool.convertLogicIDsToDeviceNames(logicIds), convey.ShouldEqual, api.Ascend910MinuxPrefix+"1")
		})
	})
}

// TestHandleManuallySeparateNPUFaultInfo for test  handleManuallySeparateNPUFaultInfo
func TestHandleManuallySeparateNPUFaultInfo(t *testing.T) {
	convey.Convey("test handleManuallySeparateNPUFaultInfo", t, func() {
		convey.Convey("01-get device run mode fail, should return empty string", func() {
			tool := mockAscendTools()
			convey.So(tool.handleManuallySeparateNPUFaultInfo(), convey.ShouldEqual, "")
		})
		tool := mockAscendTools()
		mockOption := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{RealCardType: api.Ascend910A3})
		defer mockOption.Reset()
		convey.Convey("02-manually fault cache is empty, should return empty string", func() {
			convey.So(tool.handleManuallySeparateNPUFaultInfo(), convey.ShouldEqual, "")
		})
		mockStatus := gomonkey.ApplyFuncReturn(common.QueryManuallyFaultNPULogicIDsByHandleStatus, []int32{3})
		defer mockStatus.Reset()
		mockMethod := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"GetManuallySeparateNPUIDFromDeviceInfo",
			func(_ *kubeclient.ClientK8s, deviceInfoCMName, deviceInfoCMNamespace string) []int32 {
				return []int32{1}
			})
		defer mockMethod.Reset()
		convey.Convey("03-handle fault info success, manually separate npu 1, should return Ascend910-1", func() {
			convey.So(tool.handleManuallySeparateNPUFaultInfo(), convey.ShouldEqual, api.Ascend910MinuxPrefix+"1")
		})
	})
}

// TestIsDeviceStatusChange testIsDeviceStatusChange
func TestIsDeviceStatusChange(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test IsDeviceStatusChange true", t, func() {
		devices := map[string][]*common.NpuDevice{api.Ascend910: {{Health: v1beta1.Healthy}}}
		aiCoreDevice := []*common.NpuDevice{{Health: v1beta1.Healthy}}
		oldDevice := deepCopyGroupDevice(devices)
		tool.UpdateHealth(devices, aiCoreDevice, api.Ascend910)
		res := tool.GetChange(devices, oldDevice)
		convey.So(res, convey.ShouldNotBeNil)
	})
	tool = AscendTools{name: api.Ascend310P, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMockErr{}}
	convey.Convey("test IsDeviceStatusChange which chip is unhealthy ", t, func() {
		devices := map[string][]*common.NpuDevice{api.Ascend310P: {{Health: v1beta1.Unhealthy}}}
		aiCoreDevice := []*common.NpuDevice{{Health: v1beta1.Unhealthy}}
		oldDevice := deepCopyGroupDevice(devices)
		tool.UpdateHealth(devices, aiCoreDevice, api.Ascend310P)
		res := tool.GetChange(devices, oldDevice)
		convey.So(res, convey.ShouldNotBeNil)
	})
}

// TestSetAICoreHealthyIfVNpu for test setAICoreHealthyIfVNpu
func TestSetAICoreHealthyIfVNpu(t *testing.T) {
	convey.Convey("test setAICoreHealthyIfVNpu", t, func() {
		groupDevices := map[string][]*common.NpuDevice{api.Ascend910: {{LogicID: 1, Health: v1beta1.Healthy}}}
		aiCoreDevs := []*common.NpuDevice{{LogicID: 1, Health: v1beta1.Unhealthy}}
		convey.Convey("01-presetVDevice is false, not update aiCoreDevs", func() {
			mockOption := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{PresetVDevice: false})
			defer mockOption.Reset()
			setAICoreHealthyIfVNpu(groupDevices, aiCoreDevs)
			convey.So(aiCoreDevs[0].Health == v1beta1.Healthy, convey.ShouldBeTrue)
		})
		convey.Convey("presetVDevice is true, update aiCoreDevs, aiCoreDevs[0] should be Unhealthy", func() {
			mockOption := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{PresetVDevice: true})
			defer mockOption.Reset()
			setAICoreHealthyIfVNpu(groupDevices, aiCoreDevs)
			convey.So(aiCoreDevs[0].Health == v1beta1.Healthy, convey.ShouldBeFalse)
		})
	})
}

// TestSetHealthyIfDuoCard for test setHealthyIfDuoCard
func TestSetHealthyIfDuoCard(t *testing.T) {
	convey.Convey("test setHealthyIfDuoCard", t, func() {
		convey.Convey("01-not contain atlas 300IDuo, should not update groupDevice", func() {
			mockOption := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{ProductTypes: []string{}})
			defer mockOption.Reset()
			setHealthyIfDuoCard(map[string][]*common.NpuDevice{})
		})
		convey.Convey("02-HotReset is false, should not update groupDevice", func() {
			mockOption := gomonkey.ApplyGlobalVar(&common.ParamOption,
				common.Option{ProductTypes: []string{common.Atlas300IDuo}, HotReset: common.HotResetTrainOnLine})
			defer mockOption.Reset()
			setHealthyIfDuoCard(map[string][]*common.NpuDevice{})
		})
		convey.Convey("03-not found devices, should not update groupDevice", func() {
			mockOption := gomonkey.ApplyGlobalVar(&common.ParamOption,
				common.Option{ProductTypes: []string{common.Atlas300IDuo}, HotReset: common.HotResetInfer})
			defer mockOption.Reset()
			setHealthyIfDuoCard(map[string][]*common.NpuDevice{})
		})
		convey.Convey("04-update unhealthy card status, should update groupDevice", func() {
			mockOption := gomonkey.ApplyGlobalVar(&common.ParamOption,
				common.Option{ProductTypes: []string{common.Atlas300IDuo}, HotReset: common.HotResetInfer})
			defer mockOption.Reset()
			groupDevices := map[string][]*common.NpuDevice{
				api.Ascend310P: {{CardID: 0, Health: v1beta1.Healthy}, {CardID: 0, Health: v1beta1.Unhealthy}},
			}
			setHealthyIfDuoCard(groupDevices)
			convey.So(groupDevices[api.Ascend310P][0].Health == v1beta1.Healthy, convey.ShouldBeFalse)
		})
	})
}

// TestIsHealthy for test isHealthy
func TestIsHealthy(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test isHealthy", t, func() {
		device := &common.NpuDevice{
			FaultCodes: []int64{int64(0x80C98008), int64(0x80CB8008)},
			LogicID:    0,
			DeviceName: "Ascend910-0",
		}
		convey.Convey("01-normal npu is healthy, device should be healthy", func() {
			mockFaultType := gomonkey.ApplyFuncReturn(common.GetFaultType, common.NormalNPU)
			defer mockFaultType.Reset()
			convey.So(tool.isHealthy(device) == v1beta1.Healthy, convey.ShouldBeTrue)
		})
		convey.Convey("02-PreSeparate npu is healthy and npu is used now, device should be healthy", func() {
			mockFaultType := gomonkey.ApplyFuncReturn(common.GetFaultType, common.PreSeparateNPU)
			defer mockFaultType.Reset()
			mockNpuIsUseNow := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(AscendTools)), "npuIsUsedNow",
				func(_ *AscendTools, deviceName string) bool { return true })
			defer mockNpuIsUseNow.Reset()
			convey.So(tool.isHealthy(device) == v1beta1.Healthy, convey.ShouldBeTrue)
		})
		convey.Convey("03-PreSeparate npu is healthy and npu is not used now, device should be unhealthy", func() {
			mockFaultType := gomonkey.ApplyFuncReturn(common.GetFaultType, common.PreSeparateNPU)
			defer mockFaultType.Reset()
			mockNpuIsUseNow := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(AscendTools)), "npuIsUsedNow",
				func(_ *AscendTools, deviceName string) bool { return false })
			defer mockNpuIsUseNow.Reset()
			convey.So(tool.isHealthy(device) == v1beta1.Unhealthy, convey.ShouldBeTrue)
		})
	})
}

// TestIsNetworkHealthy for test isNetworkHealthy
func TestIsNetworkHealthy(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test isNetworkHealthy", t, func() {
		device := &common.NpuDevice{}
		convey.Convey("01-network faultType is NormalNpu, network should be healthy", func() {
			mockFaultType := gomonkey.ApplyFuncReturn(common.GetNetworkFaultType, common.NormalNPU)
			defer mockFaultType.Reset()
			convey.So(tool.isNetworkHealthy(device) == v1beta1.Healthy, convey.ShouldBeTrue)
		})
		convey.Convey("02-network faultType is PreSeparateNpu, network should be unhealthy", func() {
			mockFaultType := gomonkey.ApplyFuncReturn(common.GetNetworkFaultType, common.PreSeparateNPU)
			defer mockFaultType.Reset()
			convey.So(tool.isNetworkHealthy(device) == v1beta1.Unhealthy, convey.ShouldBeTrue)
		})
	})
}

// TestMoreThanFiveMin for test moreThanFiveMin
func TestMoreThanFiveMin(t *testing.T) {
	convey.Convey("test moreThanFiveMin", t, func() {
		// 01-AlarmRaisedTime is 0, should return false
		device := &common.NpuDevice{AlarmRaisedTime: 0}
		convey.So(moreThanFiveMin(device), convey.ShouldBeFalse)
		// 02-more then five minutes, should return true
		device.AlarmRaisedTime = 3000
		convey.So(moreThanFiveMin(device), convey.ShouldBeTrue)
	})
}

// TestNetworkMoreThanFiveMin for test networkMoreThanFiveMin
func TestNetworkMoreThanFiveMin(t *testing.T) {
	convey.Convey("test networkMoreThanFiveMin", t, func() {
		// 01-AlarmRaisedTime is 0, should return false
		device := &common.NpuDevice{NetworkAlarmRaisedTime: 0}
		convey.So(networkMoreThanFiveMin(device), convey.ShouldBeFalse)
		// 02-more then five minutes, should return true
		device.NetworkAlarmRaisedTime = 3000
		convey.So(networkMoreThanFiveMin(device), convey.ShouldBeTrue)
	})
}

// TestLogFaultModeChange for test LogFaultModeChange
func TestLogFaultModeChange(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test LogFaultModeChange", t, func() {
		convey.Convey("01-faultMode is empty, should update fault mode", func() {
			device := &common.NpuDevice{LogicID: 1}
			mockFaultMode := gomonkey.ApplyGlobalVar(&faultMode, map[int32]string{})
			defer mockFaultMode.Reset()
			tool.LogFaultModeChange(device, []int32{0, 1}, common.Polling)
			convey.So(faultMode[1] == common.Polling, convey.ShouldBeTrue)
		})
		convey.Convey("02-old mode is equal to new mode, should not update fault mode", func() {
			device := &common.NpuDevice{LogicID: 1}
			mockFaultMode := gomonkey.ApplyGlobalVar(&faultMode, map[int32]string{1: common.Polling})
			defer mockFaultMode.Reset()
			tool.LogFaultModeChange(device, []int32{0, 1}, common.Polling)
			convey.So(faultMode[1] == common.Polling, convey.ShouldBeTrue)
		})
		convey.Convey("03-old mode is different from new mode, should update fault mode", func() {
			device := &common.NpuDevice{LogicID: 1, Health: v1beta1.Unhealthy, AlarmRaisedTime: 3000}
			mockFaultMode := gomonkey.ApplyGlobalVar(&faultMode, map[int32]string{1: common.Subscribe})
			defer mockFaultMode.Reset()
			tool.LogFaultModeChange(device, []int32{0, 1}, common.Polling)
			convey.So(faultMode[1] == common.Polling, convey.ShouldBeTrue)
		})
	})
}

// TestGetNPUsByShareMode for test getNPUsByShareMode
func TestGetNPUsByShareMode(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test getNPUsByShareMode", t, func() {
		// 01-stub ShareCount, get npu by share mode success, share devices number is equal to ShareCount
		mockOption := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{ShareCount: 2})
		defer mockOption.Reset()
		davinCiDev := common.DavinCiDev{LogicID: 1, PhyID: 1, CardID: 1, IP: "127.0.0.1"}
		numNum := 2
		devs := tool.getNPUsByShareMode(davinCiDev)
		convey.So(len(devs) == numNum, convey.ShouldBeTrue)
		convey.So(devs[0].IP == "127.0.0.1", convey.ShouldBeTrue)
		convey.So(devs[0].DeviceName == api.Ascend910MinuxPrefix+"1-0", convey.ShouldBeTrue)
	})
}

// TestAssembleShareModeDevices for test assembleShareModeDevices
func TestAssembleShareModeDevices(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test assembleShareModeDevices", t, func() {
		// 01-get npu by share mode success, update npuDevs ans devTypes
		mockOption := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{ShareCount: 1})
		defer mockOption.Reset()
		davinCiDev := common.DavinCiDev{LogicID: 1, PhyID: 1, CardID: 1, IP: "127.0.0.1"}
		npuDevs := &[]common.NpuDevice{}
		devTypes := &[]string{}
		tool.assembleShareModeDevices(davinCiDev, npuDevs, devTypes)
		convey.So(len(*npuDevs) == 1, convey.ShouldBeTrue)
		convey.So(len(*devTypes) == 1, convey.ShouldBeTrue)
		convey.So((*npuDevs)[0].IP == "127.0.0.1", convey.ShouldBeTrue)
	})
}

// TestSetDeviceUsage for test SetDeviceUsage
func TestSetDeviceUsage(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test SetDeviceUsage", t, func() {
		devLoginID := int32(2)
		convey.Convey("01-node info is nil, should return error", func() {
			mockGetNodeMethod := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
				"GetNode", func(_ *kubeclient.ClientK8s) (*v1.Node, error) {
					return nil, errors.New("node is nil")
				})
			defer mockGetNodeMethod.Reset()
			convey.So(tool.SetDeviceUsage(devLoginID), convey.ShouldNotBeNil)
		})
		convey.Convey("02-Nodes are used for inference, should return nil", func() {
			node := getMockNode()
			node.Labels[common.ServerUsageLabelKey] = common.Infer
			mockGetNodeMethod := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
				"GetNode", func(_ *kubeclient.ClientK8s) (*v1.Node, error) { return node, nil })
			defer mockGetNodeMethod.Reset()
			convey.So(tool.SetDeviceUsage(devLoginID), convey.ShouldBeNil)
		})
		node := getMockNode()
		node.Labels[common.ServerUsageLabelKey] = common.Train
		mockGetNodeMethod := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"GetNode", func(_ *kubeclient.ClientK8s) (*v1.Node, error) { return node, nil })
		defer mockGetNodeMethod.Reset()
		convey.Convey("03-get board error, should return error", func() {
			mockGetServerBoardIdMethod := gomonkey.ApplyMethod(reflect.TypeOf(new(AscendTools)),
				"GetServerBoardId", func(_ *AscendTools, devLogicID int32) (uint32, error) {
					return 0, errors.New("get board error")
				})
			defer mockGetServerBoardIdMethod.Reset()
			convey.So(tool.SetDeviceUsage(devLoginID), convey.ShouldNotBeNil)
		})
		convey.Convey("get board success", func() {
			boardID := uint32(0x3c)
			patch := gomonkey.ApplyMethod(reflect.TypeOf(new(AscendTools)),
				"GetServerBoardId", func(_ *AscendTools, devLogicID int32) (uint32, error) { return boardID, nil })
			defer patch.Reset()
			convey.Convey("04-devType is not Ascend910B, should return nil", func() {
				convey.So(tool.SetDeviceUsage(devLoginID), convey.ShouldBeNil)
			})
			convey.Convey("05-devType is Ascend910B, should return nil", func() {
				patch.ApplyMethodReturn(tool.dmgr, "GetDevType", api.Ascend910B)
				convey.So(tool.SetDeviceUsage(devLoginID), convey.ShouldBeNil)
			})
		})
	})
}

// TestGetServerBoardId for test GetServerBoardId
func TestGetServerBoardId(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test GetServerBoardId", t, func() {
		devLogicID := int32(2)
		convey.Convey("01-board id is not empty, should return nil", func() {
			tool.boardId = common.A800IA2NoneHccsBoardId
			_, err := tool.GetServerBoardId(devLogicID)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-board id is empty, should return nil", func() {
			tool.boardId = common.EmptyBoardId
			boardId, err := tool.GetServerBoardId(devLogicID)
			convey.So(err, convey.ShouldBeNil)
			convey.So(boardId == 0, convey.ShouldBeTrue)
		})
	})
}

// TestWriteFaultToEvent for test writeFaultToEvent
func TestWriteFaultToEvent(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test writeFaultToEvent", t, func() {
		mockDoWriteFaultToEventMethod := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(AscendTools)),
			"doWriteFaultToEvent", func(_ *AscendTools, faultInfo npuCommon.DevFaultInfo) error {
				return errors.New("write fault to event fail")
			})
		defer mockDoWriteFaultToEventMethod.Reset()
		allFaultInfo <- npuCommon.DevFaultInfo{LogicID: 0, Assertion: FaultOnce, EventID: common.CardDropFaultCode}
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(time.Second)
			cancel()
		}()
		tool.WriteFaultToEvent(ctx)
	})
}

// TestDoWriteFaultToEvent for test doWriteFaultToEvent
func TestDoWriteFaultToEvent(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test doWriteFaultToEvent", t, func() {
		mockGetNodeNameFromEnv := gomonkey.ApplyFuncReturn(kubeclient.GetNodeNameFromEnv, "mock node", nil)
		defer mockGetNodeNameFromEnv.Reset()
		mockGetPodNameFromEnv := gomonkey.ApplyFuncReturn(common.GetPodNameFromEnv, "mock pod", nil)
		defer mockGetPodNameFromEnv.Reset()
		mockCreateEvent := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"CreateEvent", func(_ *kubeclient.ClientK8s, evt *v1.Event) (*v1.Event, error) { return nil, nil })
		defer mockCreateEvent.Reset()
		convey.Convey("01-assertion is invalid, should return error", func() {
			faultInfo := npuCommon.DevFaultInfo{LogicID: 0, Assertion: 3, EventID: common.CardDropFaultCode}
			convey.So(tool.doWriteFaultToEvent(faultInfo), convey.ShouldNotBeNil)
		})
		convey.Convey("02-write fault to event success, should return nil", func() {
			faultInfo := npuCommon.DevFaultInfo{LogicID: 0, Assertion: npuCommon.FaultRecover, EventID: common.LinkDownFaultCode}
			convey.So(tool.doWriteFaultToEvent(faultInfo), convey.ShouldBeNil)
		})
	})
}

// TestHandleDropCardFaultEvents for test HandleDropCardFaultEvents
func TestHandleDropCardFaultEvents(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test HandleDropCardFaultEvents", t, func() {
		mockSaveDevFaultInfo := gomonkey.ApplyFunc(common.DoSaveDevFaultInfo,
			func(devFaultInfo npuCommon.DevFaultInfo, enableDelay bool) {})
		defer mockSaveDevFaultInfo.Reset()
		convey.Convey("01-occur fault event, CardDrop should be true", func() {
			mockCheckCardDropFault := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(AscendTools)),
				"checkCardDropFault", func(_ *AscendTools, logicID int32) bool { return true })
			defer mockCheckCardDropFault.Reset()
			npuDevice := &common.NpuDevice{LogicID: 1, CardDrop: false}
			tool.generateCardDropFaultEvents(npuDevice)
			convey.So(npuDevice.CardDrop, convey.ShouldBeTrue)
		})
		convey.Convey("recover fault event, CardDrop should be false", func() {
			mockCheckCardDropFault := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(AscendTools)),
				"checkCardDropFault", func(_ *AscendTools, logicID int32) bool { return false })
			defer mockCheckCardDropFault.Reset()
			npuDevice := &common.NpuDevice{LogicID: 1, CardDrop: true}
			tool.generateCardDropFaultEvents(npuDevice)
			convey.So(npuDevice.CardDrop, convey.ShouldBeFalse)
		})
	})
}

// TestCheckCardDropFault for test checkCardDropFault
func TestCheckCardDropFault(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test checkCardDropFault", t, func() {
		convey.Convey("01-check card drop fault fail, should return false", func() {
			convey.So(tool.checkCardDropFault(1), convey.ShouldBeFalse)
		})
		convey.Convey("02-check card drop fault success, should return true", func() {
			mockGetDeviceIPAddress := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetDeviceHealth", func(_ *devmanager.DeviceManagerMock, logicID int32) (uint32, error) {
					return 0, errors.New("error is " + npuCommon.DeviceNotReadyErrCodeStr)
				})
			defer mockGetDeviceIPAddress.Reset()
			convey.So(tool.checkCardDropFault(1), convey.ShouldBeTrue)
		})
	})
}

// TestHandleLostChipFaultEvents for test HandleLostChipFaultEvents
func TestHandleLostChipFaultEvents(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test HandleLostChipFaultEvents", t, func() {
		device := &common.NpuDevice{
			FaultCodes: []int64{},
			LogicID:    1,
		}
		mockGlobalVar := gomonkey.ApplyGlobalVar(&isFirstFlushFault, true)
		defer mockGlobalVar.Reset()
		convey.Convey("01-get device all error code fail, devFaultInfoMap should not be updated", func() {
			mockGetDeviceAllErrorCode := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetDeviceAllErrorCode", func(_ *devmanager.DeviceManagerMock, logicID int32) (int32, []int64, error) {
					return 1, nil, errors.New("mock failure message")
				})
			defer mockGetDeviceAllErrorCode.Reset()
			tool.HandleLostChipFaultEvents(device, nil)
			faultInfo := common.GetAndCleanFaultInfo()
			convey.So(len(faultInfo) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("02-handle lost chip fault event, devFaultInfoMap should be updated", func() {
			mockGetDeviceAllErrorCode := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetDeviceAllErrorCode", func(_ *devmanager.DeviceManagerMock, logicID int32) (int32, []int64, error) {
					return 1, []int64{common.LinkDownFaultCode, common.CardDropFaultCode}, nil
				})
			defer mockGetDeviceAllErrorCode.Reset()
			tool.HandleLostChipFaultEvents(device, nil)
			faultInfoMap := common.GetAndCleanFaultInfo()
			convey.So(len(faultInfoMap) == 1, convey.ShouldBeTrue)
			faultInfo, ok := faultInfoMap[1]
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(len(faultInfo) == 1, convey.ShouldBeTrue)
			convey.So(faultInfo[0].EventID == common.CardDropFaultCode, convey.ShouldBeTrue)
		})
	})
}

// TestHandleLostNetworkFaultEvents for test HandleLostNetworkFaultEvents
func TestHandleLostNetworkFaultEvents(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test HandleLostNetworkFaultEvents", t, func() {
		device := &common.NpuDevice{
			FaultCodes: []int64{},
			LogicID:    1,
		}
		mockGlobalVar := gomonkey.ApplyGlobalVar(&isFirstFlushFault, true)
		defer mockGlobalVar.Reset()

		convey.Convey("01-get device all error code fail, devFaultInfoMap should not be updated", func() {
			mockGetDeviceAllErrorCode := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetDeviceAllErrorCode", func(_ *devmanager.DeviceManagerMock, logicID int32) (int32, []int64, error) {
					return 1, nil, errors.New("mock failure message")
				})
			defer mockGetDeviceAllErrorCode.Reset()
			tool.HandleLostNetworkFaultEvents(device, nil)
			faultInfo := common.GetAndCleanFaultInfo()
			convey.So(len(faultInfo) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("02-handle lost network fault event, devFaultInfoMap should be updated", func() {
			mockGetDeviceAllErrorCode := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetDeviceAllErrorCode", func(_ *devmanager.DeviceManagerMock, logicID int32) (int32, []int64, error) {
					return 1, []int64{common.LinkDownFaultCode, common.CardDropFaultCode}, nil
				})
			defer mockGetDeviceAllErrorCode.Reset()
			tool.HandleLostNetworkFaultEvents(device, nil)
			faultInfoMap := common.GetAndCleanFaultInfo()
			convey.So(len(faultInfoMap) == 1, convey.ShouldBeTrue)
			faultInfo, ok := faultInfoMap[1]
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(len(faultInfo) == 1, convey.ShouldBeTrue)
			convey.So(faultInfo[0].EventID == common.LinkDownFaultCode, convey.ShouldBeTrue)
		})
	})
}

// TestAssembleVirtualDevices testAssembleVirtualDevices
func TestAssembleVirtualDevices(t *testing.T) {
	convey.Convey("test assembleVirtualDevices", t, func() {
		tool := AscendTools{name: api.Ascend910, client: &kubeclient.ClientK8s{},
			dmgr: &devmanager.DeviceManagerMock{}}

		var device []common.NpuDevice
		var deivceType []string
		davinCiDev := common.DavinCiDev{
			PhyID:   phyIDNum,
			LogicID: logicIDNum,
		}

		QueryInfo := npuCommon.CgoVDevQueryInfo{
			Computing: npuCommon.CgoComputingResource{Aic: aiCoreNum},
			Name:      "vir16",
		}
		vDevInfos := npuCommon.VirtualDevInfo{
			VDevInfo: []npuCommon.CgoVDevQueryStru{{QueryInfo: QueryInfo, VDevID: vDevIDNum}},
		}
		tool.assembleVirtualDevices(davinCiDev, vDevInfos, &device, &deivceType)
		testRes := common.NpuDevice{
			DevType:       common.Ascend910vir16,
			DeviceName:    fmt.Sprintf("%s-%d-%d", common.Ascend910vir16, vDevIDNum, phyIDNum),
			Health:        v1beta1.Healthy,
			NetworkHealth: v1beta1.Healthy,
			LogicID:       logicIDNum,
			PhyID:         phyIDNum,
		}
		convey.So(device, convey.ShouldContain, testRes)
	})
}

// TestAddPodAnnotation1 for test the interface AddPodAnnotation, part 1
func TestAddPodAnnotation1(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test AddPodAnnotation 1", t, func() {
		convey.Convey("GetDeviceListID failed", func() {
			err := tool.AddPodAnnotation(&common.PodDeviceInfo{
				Pod:        v1.Pod{},
				KltDevice:  nil,
				RealDevice: []string{api.Ascend910},
			}, common.Ascend910vir2, "", nil)
			convey.So(err, convey.ShouldBeNil)
		})
		mockTryUpdatePodAnnotation := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"TryUpdatePodAnnotation", func(_ *kubeclient.ClientK8s, pod *v1.Pod,
				annotation map[string]string) error {
				return nil
			})
		defer mockTryUpdatePodAnnotation.Reset()
		convey.Convey("physical device 310P", func() {
			tool.name = api.Ascend310P
			err := tool.AddPodAnnotation(&common.PodDeviceInfo{
				Pod:        v1.Pod{},
				KltDevice:  nil,
				RealDevice: []string{api.Ascend310P + "-0"},
			}, api.Ascend310P, "", nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("virtual device", func() {
			err := tool.AddPodAnnotation(&common.PodDeviceInfo{
				Pod:        v1.Pod{},
				KltDevice:  nil,
				RealDevice: []string{common.Ascend310Pc2 + "-100-0"},
			}, common.Ascend310Pc2, "", nil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestAddPodAnnotation2 for test the interface AddPodAnnotation, part 2
func TestAddPodAnnotation2(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test AddPodAnnotation 2", t, func() {
		mockTryUpdatePodAnnotation := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"TryUpdatePodAnnotation", func(_ *kubeclient.ClientK8s, pod *v1.Pod,
				annotation map[string]string) error {
				return nil
			})
		defer mockTryUpdatePodAnnotation.Reset()
		mockGetLogicIDFromPhysicID := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetLogicIDFromPhysicID", func(_ *devmanager.DeviceManagerMock, physicID int32) (int32, error) {
				return 0, nil
			})
		defer mockGetLogicIDFromPhysicID.Reset()
		convey.Convey("GetDeviceIPAddress failed", func() {
			mockGetDeviceIPAddress := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetDeviceIPAddress", func(_ *devmanager.DeviceManagerMock, logicID, ipType int32) (
					string, error) {
					return "", fmt.Errorf("error")
				})
			defer mockGetDeviceIPAddress.Reset()
			err := tool.AddPodAnnotation(&common.PodDeviceInfo{
				Pod:        v1.Pod{},
				KltDevice:  nil,
				RealDevice: []string{api.Ascend910 + "-0"},
			}, api.Ascend910, "", nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("GetDeviceIPAddress ok", func() {
			mockGetDeviceIPAddress := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetDeviceIPAddress", func(_ *devmanager.DeviceManagerMock, logicID, ipType int32) (
					string, error) {
					return "", nil
				})
			defer mockGetDeviceIPAddress.Reset()
			err := tool.AddPodAnnotation(&common.PodDeviceInfo{
				Pod:        v1.Pod{},
				KltDevice:  nil,
				RealDevice: []string{api.Ascend910 + "-0"},
			}, api.Ascend910, "", nil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestAddPodAnnotation3 for test the interface AddPodAnnotation, part 3
func TestAddPodAnnotation3(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test AddPodAnnotation 3", t, func() {
		mockTryUpdatePodAnnotation := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"TryUpdatePodAnnotation", func(_ *kubeclient.ClientK8s, pod *v1.Pod,
				annotation map[string]string) error {
				return nil
			})
		defer mockTryUpdatePodAnnotation.Reset()
		convey.Convey("GetLogicIDFromPhysicID failed", func() {
			mockGetLogicIDFromPhysicID := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetLogicIDFromPhysicID", func(_ *devmanager.DeviceManagerMock, physicID int32) (int32, error) {
					return 0, fmt.Errorf("error")
				})
			defer mockGetLogicIDFromPhysicID.Reset()
			err := tool.AddPodAnnotation(&common.PodDeviceInfo{
				Pod:        v1.Pod{},
				KltDevice:  nil,
				RealDevice: []string{api.Ascend910 + "-0"},
			}, api.Ascend910, "", nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("ParseInt failed", func() {
			tool.name = api.Ascend910
			err := tool.AddPodAnnotation(&common.PodDeviceInfo{
				Pod:        v1.Pod{},
				KltDevice:  nil,
				RealDevice: []string{api.Ascend910 + "-a"},
			}, api.Ascend910, "", nil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestCreateVirtualDevice testCreateVirtualDevice
func TestCreateVirtualDevice(t *testing.T) {
	tool := AscendTools{name: api.Ascend310P, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	convey.Convey("test CreateVirtualDevice", t, func() {
		convey.Convey("CreateVirtualDevice success", func() {
			mockGetLogicIDFromPhysicID := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetLogicIDFromPhysicID", func(_ *devmanager.DeviceManagerMock, physicID int32) (int32, error) {
					return 0, nil
				})
			mockCreate := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"CreateVirtualDevice", func(_ *devmanager.DeviceManagerMock, logicID int32,
					vDevInfo npuCommon.CgoCreateVDevRes) (npuCommon.CgoCreateVDevOut, error) {
					return npuCommon.CgoCreateVDevOut{}, nil
				})
			defer mockCreate.Reset()
			defer mockGetLogicIDFromPhysicID.Reset()
			_, err := tool.CreateVirtualDevice(0, "vir01")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestDestroyVirtualDevice testDestroyVirtualDevice
func TestDestroyVirtualDevice(t *testing.T) {
	tool := AscendTools{name: api.Ascend310P, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	convey.Convey("test DestroyVirtualDevice", t, func() {
		convey.Convey("DestroyVirtualDevice success", func() {
			mockGetLogicIDFromPhysicID := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetLogicIDFromPhysicID", func(_ *devmanager.DeviceManagerMock, physicID int32) (int32, error) {
					return 0, nil
				})
			mockDestroy := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"DestroyVirtualDevice", func(_ *devmanager.DeviceManagerMock, _ int32, _ uint32) error {
					return nil
				})
			defer mockDestroy.Reset()
			defer mockGetLogicIDFromPhysicID.Reset()
			err := tool.DestroyVirtualDevice("Ascend310P-1c-100-0")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetChipAiCoreCount testGetChipAiCoreCount
func TestGetChipAiCoreCount(t *testing.T) {
	tool := AscendTools{name: api.Ascend310P, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	res := getVirtualDevInfo(aiCoreNum)
	mockLogicIDs := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
		"GetDeviceList", func(_ *devmanager.DeviceManagerMock) (int32, []int32, error) {
			return 1, []int32{0}, nil
		})
	mockVirtual := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
		"GetVirtualDeviceInfo", func(_ *devmanager.DeviceManagerMock, _ int32) (
			npuCommon.VirtualDevInfo, error) {
			return res, nil
		})
	defer mockVirtual.Reset()
	defer mockLogicIDs.Reset()
	convey.Convey("test GetChipAiCoreCount 1", t, func() {
		convey.Convey("GetChipAiCoreCount failed", func() {
			_, err := tool.GetChipAiCoreCount()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
	res = getVirtualDevInfo(aiCoreCount)
	convey.Convey("test GetChipAiCoreCount 2", t, func() {
		convey.Convey("GetChipAiCoreCount success", func() {
			_, err := tool.GetChipAiCoreCount()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func getVirtualDevInfo(aic float32) npuCommon.VirtualDevInfo {
	return npuCommon.VirtualDevInfo{
		TotalResource: npuCommon.CgoSocTotalResource{
			Computing: npuCommon.CgoComputingResource{
				Aic: aic,
			},
		},
		VDevInfo: []npuCommon.CgoVDevQueryStru{
			{
				VDevID: vDevChipID,
			},
		},
	}
}

// TestAppendVGroupInfo testAppendVGroupInfo
func TestAppendVGroupInfo(t *testing.T) {
	tool := AscendTools{name: api.Ascend310P, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	res := getVirtualDevInfo(aiCoreCount)
	convey.Convey("test AppendVGroupInfo", t, func() {
		convey.Convey("AppendVGroupInfo success", func() {
			mockGetLogicIDFromPhysicID := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetLogicIDFromPhysicID", func(_ *devmanager.DeviceManagerMock, physicID int32) (int32, error) {
					return 0, nil
				})
			mockVirtual := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetVirtualDeviceInfo", func(_ *devmanager.DeviceManagerMock, _ int32) (
					npuCommon.VirtualDevInfo, error) {
					return res, nil
				})
			defer mockVirtual.Reset()
			defer mockGetLogicIDFromPhysicID.Reset()
			allocateDevice := []string{
				"Ascend310P-1c-100-0",
			}
			tool.AppendVGroupInfo(allocateDevice)
			convey.So(len(allocateDevice), convey.ShouldEqual, 1)
		})
	})
}

// TestCheckDeviceTypeLabel testCheckDeviceTypeLabel
func TestCheckDeviceTypeLabel(t *testing.T) {
	tool := AscendTools{name: api.Ascend310P, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	node := getMockNode()
	convey.Convey("test CheckDeviceTypeLabel", t, func() {
		convey.Convey("CheckDeviceTypeLabel get node failed", func() {
			mockNode := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetNode",
				func(_ *kubeclient.ClientK8s) (*v1.Node, error) {
					return nil, fmt.Errorf("failed to get node")
				})
			defer mockNode.Reset()
			err := tool.CheckDeviceTypeLabel()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("CheckDeviceTypeLabel success", func() {
			mockNode := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetNode",
				func(_ *kubeclient.ClientK8s) (*v1.Node, error) {
					return node, nil
				})
			defer mockNode.Reset()
			delete(node.Labels, common.ServerTypeLabelKey)
			err := tool.CheckDeviceTypeLabel()
			convey.So(err, convey.ShouldNotBeNil)
			common.ParamOption.AiCoreCount = aiCoreCount
			node.Labels[common.ServerTypeLabelKey] = api.Ascend310PMinuxPrefix + "8"
			err = tool.CheckDeviceTypeLabel()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func getMockNode() *v1.Node {
	labels := make(map[string]string, 1)
	labels[common.ServerTypeLabelKey] = "Ascend310P-8"
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
	}
}

// TestAssemble310PMixedPhyDevices test assemble310PMixedPhyDevices
func TestAssemble310PMixedPhyDevices(t *testing.T) {
	convey.Convey("test assembleVirtualDevices", t, func() {
		tool := AscendTools{name: api.Ascend310P, client: &kubeclient.ClientK8s{},
			dmgr: &devmanager.DeviceManagerMock{}}
		var device []common.NpuDevice
		var deivceType []string
		davinCiDev := common.DavinCiDev{
			PhyID:   phyIDNum,
			LogicID: logicIDNum,
		}
		mockProductType := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetProductType",
			func(_ *devmanager.DeviceManagerMock, cardID int32, deviceID int32) (string, error) {
				return atlas300VPro, nil
			})
		defer mockProductType.Reset()
		productTypeMap := common.Get310PProductType()
		tool.assemble310PMixedPhyDevices(davinCiDev, &device, &deivceType)
		testRes := common.NpuDevice{
			DevType:       productTypeMap[atlas300VPro],
			DeviceName:    fmt.Sprintf("%s-%d", productTypeMap[atlas300VPro], phyIDNum),
			Health:        v1beta1.Healthy,
			NetworkHealth: v1beta1.Healthy,
			LogicID:       logicIDNum,
			PhyID:         phyIDNum,
		}
		convey.So(device, convey.ShouldContain, testRes)
	})
}

// TestIfCardsInResetting test if card in reset
func TestIfCardsInResetting(t *testing.T) {
	convey.Convey("test if card in reset func", t, func() {
		tool := NewHwAscend910Manager()
		tool.SetCardsInResetting(common.FirstDevice, true)
		convey.So(tool.GetIfCardsInResetting(common.FirstDevice), convey.ShouldEqual, true)
		convey.So(tool.GetIfCardsInResetting(logicIDNum), convey.ShouldBeFalse)
		tool.SetCardsInResetting(common.FirstDevice, false)
		convey.So(tool.GetIfCardsInResetting(common.FirstDevice), convey.ShouldBeFalse)
	})
}

// TestGetResetFailedTimes test get reset failed times
func TestGetResetFailedTimes(t *testing.T) {
	convey.Convey("test set reset failed times", t, func() {
		tool := NewHwAscend910Manager()
		tool.SetResetFailedTimes(common.FirstDevice, FaultOnce)
		convey.So(tool.GetResetFailedTimes(common.FirstDevice), convey.ShouldEqual, FaultOnce)
		convey.So(tool.GetResetFailedTimes(logicIDNum), convey.ShouldEqual, NoneFault)
	})
}

func TestRemoveDuplicateErr(t *testing.T) {
	convey.Convey("test remove duplicate errors", t, func() {
		code98008 := int64(0x80C98008)
		codeB8008 := int64(0x80CB8008)
		code98002 := int64(0x80C98002)
		code98003 := int64(0x80C98003)
		code98009 := int64(0x80C98009)
		codeB8002 := int64(0x80CB8002)
		codeB8009 := int64(0x80CB8009)
		oldErrors := []int64{code98008, code98002, code98003, code98009, codeB8002, codeB8008, codeB8009}
		tool := NewHwAscend910Manager()
		newErrors := tool.removeDuplicateErr(oldErrors)
		convey.So(len(oldErrors), convey.ShouldEqual, len(newErrors))
		baseErrors := []int64{code98008, codeB8008}
		oldErrors = []int64{code98008, code98008, code98008, code98008, code98008, code98008, codeB8008, codeB8008}
		newErrors = tool.removeDuplicateErr(oldErrors)
		convey.So(len(baseErrors), convey.ShouldEqual, len(newErrors))
	})
}

// TestGetResetInfoData for test getResetInfoData
func TestGetResetInfoData(t *testing.T) {
	convey.Convey("test getResetInfoData", t, func() {
		resetInfo := &v1.ConfigMap{}
		convey.Convey("01-reset.json not exist, should return error", func() {
			_, err := getResetInfoData(resetInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-unmarshal fail, should return error", func() {
			mockUnmarshal := gomonkey.ApplyFuncReturn(json.Unmarshal, errors.New("fail"))
			defer mockUnmarshal.Reset()
			resetInfo.Data = map[string]string{common.ResetInfoCMDataKey: "yyy"}
			_, err := getResetInfoData(resetInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-unmarshal success, should return nil", func() {
			taskResetInfo := common.TaskResetInfo{RankList: []*common.TaskDevInfo{{RankId: 0}, {RankId: 1}}, UpdateTime: 20}
			devNumsFromRestJson := 2
			resetByte, err := json.Marshal(taskResetInfo)
			convey.So(err, convey.ShouldBeNil)
			resetInfo.Data = map[string]string{common.ResetInfoCMDataKey: string(resetByte)}
			devInfo, err := getResetInfoData(resetInfo)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(devInfo) == devNumsFromRestJson, convey.ShouldBeTrue)
		})

	})
}

// TestNpuIsUsedNow for test npuIsUsedNow
func TestNpuIsUsedNow(t *testing.T) {
	convey.Convey("test npuIsUsedNow", t, func() {
		tool := mockAscendTools()
		annotationTag := fmt.Sprintf("%s%s", api.ResourceNamePrefix, api.Ascend910)
		pods := []v1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "mock pod1", Annotations: map[string]string{annotationTag: ""}}},
			{ObjectMeta: metav1.ObjectMeta{Name: "mock pod2", Annotations: map[string]string{annotationTag: "device1,device2"}}},
		}
		mockMethod := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetActivePodListCache",
			func(_ *kubeclient.ClientK8s) []v1.Pod { return pods })
		defer mockMethod.Reset()
		convey.Convey("01-device exist, should return true", func() {
			convey.So(tool.npuIsUsedNow("device1"), convey.ShouldBeTrue)
		})
		convey.Convey("02-device not exist, should return false", func() {
			convey.So(tool.npuIsUsedNow("device3"), convey.ShouldBeFalse)
		})
	})
}

// TestGetRealUsedDevices for test getRealUsedDevices
func TestGetRealUsedDevices(t *testing.T) {
	convey.Convey("test getRealUsedDevices", t, func() {
		tool := mockAscendTools()
		annotationTag := api.PodAnnotationAscendReal
		pods := []v1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "mock pod1", Annotations: map[string]string{"test tag": "device3"}}},
			{ObjectMeta: metav1.ObjectMeta{Name: "mock pod2", Annotations: map[string]string{annotationTag: "device1,device2"}}},
		}
		mockMethod := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetActivePodListCache",
			func(_ *kubeclient.ClientK8s) []v1.Pod { return pods })
		defer mockMethod.Reset()
		podUsedDev := tool.getRealUsedDevices()
		// 01-annotationTag has target device, should contain target device
		convey.So(podUsedDev.Has("device1"), convey.ShouldBeTrue)
		// 02-annotationTag has not target device, should not contain target device
		convey.So(podUsedDev.Has("device3"), convey.ShouldBeFalse)
	})
}

func TestGetDevStatesDevSet(t *testing.T) {
	convey.Convey("test getDevStatesDevSet", t, func() {
		tool := mockAscendTools()
		mockGetRealUsedDevices := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&tool), "getRealUsedDevices",
			func(_ *AscendTools) sets.String { return sets.String{} })
		mockGroupDevsByStatus := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&tool), "groupDevsByStatus",
			func(_ *AscendTools, subClassDevices []*common.NpuDevice, runMode string) common.DevStatusSet {
				return common.DevStatusSet{HealthDevices: sets.String{"Ascend910-0": sets.Empty{}}}
			})
		mockGetPodsUsedNpu := mockGetPodsUsedNpuByCommon()
		defer func() {
			mockGetRealUsedDevices.Reset()
			mockGroupDevsByStatus.Reset()
			mockGetPodsUsedNpu.Reset()
		}()
		mockClassifyDevs := map[string][]*common.NpuDevice{api.Ascend910: {{Health: v1beta1.Healthy}}}
		common.ParamOption.PresetVDevice = true
		res := tool.getDevStatesDevSet(mockClassifyDevs)
		convey.So(len(res.FreeHealthyDevice), convey.ShouldEqual, 1)
		convey.So(len(res.UnHealthyDevice), convey.ShouldEqual, 0)
		convey.So(len(res.NetUnHealthyDevice), convey.ShouldEqual, 0)
		convey.So(len(res.RecoveringDevices), convey.ShouldEqual, 0)
		convey.So(len(res.DeviceFault), convey.ShouldEqual, 0)
	})
}

func mockAscendTools() AscendTools {
	return AscendTools{name: api.Ascend910, client: &kubeclient.ClientK8s{}, dmgr: &devmanager.DeviceManagerMock{}}
}

// A device has both network fault and card fault, `getDeviceFaults` should return two `DeviceFault`
func TestAscendToolsGetDeviceFaults(t *testing.T) {
	t.Run("getDeviceFaults", func(t *testing.T) {
		tool := &AscendTools{}
		base16 := 16
		var faultTime1 int64 = 100000
		var faultTime2 int64 = 110000
		var faultTime3 int64 = 120000
		device := &common.NpuDevice{
			FaultCodes:             []int64{int64(0x80C98008), int64(0x80CB8008)},
			AlarmRaisedTime:        100000,
			NetworkFaultCodes:      []int64{int64(common.LinkDownFaultCode)},
			NetworkAlarmRaisedTime: 110000,
			FaultTimeMap: map[int64]int64{
				int64(0x80C98008):               faultTime1,
				int64(0x80CB8008):               faultTime2,
				int64(common.LinkDownFaultCode): faultTime3,
			},
			DeviceName: "Ascend910-0",
		}
		got := tool.getDeviceFaults(device)
		want := []common.DeviceFault{{
			FaultType:            common.CardNetworkUnhealthy,
			NPUName:              "Ascend910-0",
			LargeModelFaultLevel: common.GetNetworkFaultType(device.NetworkFaultCodes, device.LogicID),
			FaultLevel:           common.GetNetworkFaultType(device.NetworkFaultCodes, device.LogicID),
			FaultHandling:        common.GetNetworkFaultType(device.NetworkFaultCodes, device.LogicID),
			FaultCode:            strings.ToUpper(common.Int64Tool.ToHexString(device.NetworkFaultCodes)),
			FaultTimeAndLevelMap: map[string]common.FaultTimeAndLevel{
				strings.ToUpper(strconv.FormatInt(common.LinkDownFaultCode, base16)): {faultTime3,
					common.GetNetworkFaultType(device.NetworkFaultCodes, device.LogicID)},
			},
		}, {
			FaultType:            common.CardUnhealthy,
			NPUName:              "Ascend910-0",
			LargeModelFaultLevel: common.GetFaultType(device.FaultCodes, device.LogicID),
			FaultLevel:           common.GetFaultType(device.FaultCodes, device.LogicID),
			FaultHandling:        common.GetFaultType(device.FaultCodes, device.LogicID),
			FaultCode:            strings.ToUpper(common.Int64Tool.ToHexString(device.FaultCodes)),
			FaultTimeAndLevelMap: map[string]common.FaultTimeAndLevel{
				strings.ToUpper(strconv.FormatInt(int64(0x80C98008), base16)): {faultTime1,
					common.GetFaultType([]int64{0x80C98008}, device.LogicID),
				},
				strings.ToUpper(strconv.FormatInt(int64(0x80CB8008), base16)): {faultTime2,
					common.GetFaultType([]int64{0x80CB8008}, device.LogicID)},
			},
		},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("getDeviceFaults() = %v, want %v", got, want)
		}
	})
}

// initTestObjects initializes test objects
func initTestObjects() (*AscendTools, *common.NpuDevice) {
	tool := &AscendTools{
		dmgr: &devmanager.DeviceManager{},
	}
	device := &common.NpuDevice{
		LogicID:           0,
		DeviceID:          0,
		NetworkHealth:     v1beta1.Healthy,
		NetworkFaultCodes: []int64{},
	}
	return tool, device
}

func resetNetWorkLimiter() {
	netWorkLimiterMap = make(map[int32]*rate.Limiter, common.GeneralMapSize)
}

// testGetDeviceFaultFailed returns closure for get device fault failed scenario
func testGetDeviceFaultFailed(tool *AscendTools, device *common.NpuDevice) func() {
	return func() {
		defer resetNetWorkLimiter()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyMethod(&devmanager.DeviceManager{}, "GetDeviceAllErrorCode",
			func(_ *devmanager.DeviceManager, _ int32) (int32, []int64, error) {
				return -1, nil, errors.New("get device fault failed")
			})
		tool.generateNetworkFaultEventsBasedOnFaultCacheChange(device)
	}
}

// testNetworkFaultCodesNotEmpty returns closure for network fault codes not empty scenario
func testNetworkFaultCodesNotEmpty(tool *AscendTools, device *common.NpuDevice) func() {
	return func() {
		defer resetNetWorkLimiter()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		mockFaultCode := int64(0x81078603)
		patches.ApplyMethod(&devmanager.DeviceManager{}, "GetDeviceAllErrorCode",
			func(_ *devmanager.DeviceManager, _ int32) (int32, []int64, error) {
				return 1, []int64{mockFaultCode}, nil
			})
		oldNetworkFaultCodes := common.NetworkFaultCodes
		common.NetworkFaultCodes = sets.NewInt64()
		common.NetworkFaultCodes.Insert(mockFaultCode)
		defer func() { common.NetworkFaultCodes = oldNetworkFaultCodes }()
		saveCalled := false
		patches.ApplyFunc(common.DoSaveDevFaultInfo, func(_ npuCommon.DevFaultInfo, _ bool) {
			saveCalled = true
		})
		tool.generateNetworkFaultEventsBasedOnFaultCacheChange(device)
		convey.So(saveCalled, convey.ShouldBeTrue)
	}
}

// testNetworkFaultCodesEmptyLinkStatusFailed returns closure for link status failed scenario
func testNetworkFaultCodesEmptyLinkStatusFailed(tool *AscendTools, device *common.NpuDevice) func() {
	return func() {
		defer resetNetWorkLimiter()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyMethod(&devmanager.DeviceManager{}, "GetDeviceAllErrorCode",
			func(_ *devmanager.DeviceManager, _ int32) (int32, []int64, error) {
				return 0, []int64{}, nil
			})
		patches.ApplyFuncReturn(hccn.GetNPULinkStatus, "", errors.New("get link status failed"))
		tool.generateNetworkFaultEventsBasedOnFaultCacheChange(device)
	}
}

// testNetworkFaultCodesEmptyHealthyLinkDown returns closure for healthy but link down scenario
func testNetworkFaultCodesEmptyHealthyLinkDown(tool *AscendTools, device *common.NpuDevice) func() {
	return func() {
		defer resetNetWorkLimiter()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyMethod(&devmanager.DeviceManager{}, "GetDeviceAllErrorCode",
			func(_ *devmanager.DeviceManager, _ int32) (int32, []int64, error) {
				return 0, []int64{}, nil
			})
		patches.ApplyFuncReturn(hccn.GetNPULinkStatus, npuCommon.NPUNetworkLinkDownStatus, nil)
		saveCalled := false
		patches.ApplyFunc(common.DoSaveDevFaultInfo, func(_ npuCommon.DevFaultInfo, _ bool) {
			saveCalled = true
		})
		tool.generateNetworkFaultEventsBasedOnFaultCacheChange(device)
		convey.So(saveCalled, convey.ShouldBeTrue)
	}
}

// testNetworkFaultCodesEmptyUnhealthyLinkUp returns closure for unhealthy but link up scenario
func testNetworkFaultCodesEmptyUnhealthyLinkUp(tool *AscendTools, device *common.NpuDevice) func() {
	return func() {
		defer resetNetWorkLimiter()
		device.NetworkHealth = v1beta1.Unhealthy
		device.NetworkFaultCodes = []int64{common.LinkDownFaultCode}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyMethod(&devmanager.DeviceManager{}, "GetDeviceAllErrorCode",
			func(_ *devmanager.DeviceManager, _ int32) (int32, []int64, error) {
				return 0, []int64{}, nil
			})
		patches.ApplyFuncReturn(hccn.GetNPULinkStatus, npuCommon.NPUNetworkLinkUpStatus, nil)
		saveCalled := false
		patches.ApplyFunc(common.DoSaveDevFaultInfo, func(_ npuCommon.DevFaultInfo, _ bool) {
			saveCalled = true
		})
		tool.generateNetworkFaultEventsBasedOnFaultCacheChange(device)
		convey.So(saveCalled, convey.ShouldBeTrue)
	}
}

// testNetworkFaultCodesEmptyHealthMatchLink returns closure for health status matches link status scenario
func testNetworkFaultCodesEmptyHealthMatchLink(tool *AscendTools, device *common.NpuDevice) func() {
	return func() {
		defer resetNetWorkLimiter()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyMethod(&devmanager.DeviceManager{}, "GetDeviceAllErrorCode",
			func(_ *devmanager.DeviceManager, _ int32) (int32, []int64, error) {
				return 0, []int64{}, nil
			})
		patches.ApplyFuncReturn(hccn.GetNPULinkStatus, npuCommon.NPUNetworkLinkUpStatus, nil)
		saveCalled := false
		patches.ApplyFunc(common.DoSaveDevFaultInfo, func(_ npuCommon.DevFaultInfo, _ bool) {
			saveCalled = true
		})
		tool.generateNetworkFaultEventsBasedOnFaultCacheChange(device)
		convey.So(saveCalled, convey.ShouldBeFalse)
	}
}

// TestGenerateNetworkFaultEventsBasedOnFaultCacheChange main test function
func TestGenerateNetworkFaultEventsBasedOnFaultCacheChange(t *testing.T) {
	convey.Convey("test generateNetworkFaultEventsBasedOnFaultCacheChange", t, func() {
		tool, device := initTestObjects()

		// Call sub-scenarios by passing closure to Convey
		convey.Convey("get device fault failed",
			testGetDeviceFaultFailed(tool, device))
		convey.Convey("network fault codes is not empty",
			testNetworkFaultCodesNotEmpty(tool, device))
		convey.Convey("network fault codes is empty, get link status failed",
			testNetworkFaultCodesEmptyLinkStatusFailed(tool, device))
		convey.Convey("network fault codes is empty, device healthy but link down",
			testNetworkFaultCodesEmptyHealthyLinkDown(tool, device))
		convey.Convey("network fault codes is empty, device unhealthy but link up",
			testNetworkFaultCodesEmptyUnhealthyLinkUp(tool, device))
		convey.Convey("network fault codes is empty, device health status matches link status",
			testNetworkFaultCodesEmptyHealthMatchLink(tool, device))
		convey.Convey("network limiter initialization",
			testNetworkLimiterInitialization(tool, device))
		convey.Convey("network limiter rate limiting",
			testNetworkLimiterRateLimiting(tool, device))
		convey.Convey("network limiter multiple devices",
			testNetworkLimiterMultipleDevices(tool,
				&common.NpuDevice{LogicID: 0, DeviceID: 3,
					NetworkHealth: v1beta1.Healthy, NetworkFaultCodes: []int64{}},
				&common.NpuDevice{LogicID: 1, DeviceID: 4,
					NetworkHealth: v1beta1.Healthy, NetworkFaultCodes: []int64{}}))
	})
}

// testNetworkLimiterInitialization tests network limiter initialization
func testNetworkLimiterInitialization(tool *AscendTools, device *common.NpuDevice) func() {
	return func() {
		defer resetNetWorkLimiter()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyMethod(&devmanager.DeviceManager{}, "GetDeviceAllErrorCode",
			func(_ *devmanager.DeviceManager, _ int32) (int32, []int64, error) {
				return 0, []int64{}, nil
			})
		patches.ApplyFuncReturn(hccn.GetNPULinkStatus, npuCommon.NPUNetworkLinkUpStatus, nil)
		patches.ApplyFunc(common.DoSaveDevFaultInfo, func(_ npuCommon.DevFaultInfo, _ bool) {})

		tool.generateNetworkFaultEventsBasedOnFaultCacheChange(device)
	}
}

// testNetworkLimiterRateLimiting tests network limiter rate limiting functionality
func testNetworkLimiterRateLimiting(tool *AscendTools, device *common.NpuDevice) func() {
	return func() {
		defer resetNetWorkLimiter()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyMethod(&devmanager.DeviceManager{}, "GetDeviceAllErrorCode",
			func(_ *devmanager.DeviceManager, _ int32) (int32, []int64, error) {
				return 0, []int64{}, nil
			})
		patches.ApplyFuncReturn(hccn.GetNPULinkStatus, npuCommon.NPUNetworkLinkUpStatus, nil)
		patches.ApplyFunc(common.DoSaveDevFaultInfo, func(_ npuCommon.DevFaultInfo, _ bool) {})

		tool.generateNetworkFaultEventsBasedOnFaultCacheChange(device)
		tool.generateNetworkFaultEventsBasedOnFaultCacheChange(device)
	}
}

// testNetworkLimiterMultipleDevices tests network limiter with multiple devices
func testNetworkLimiterMultipleDevices(tool *AscendTools, device1, device2 *common.NpuDevice) func() {
	return func() {
		defer resetNetWorkLimiter()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyMethod(&devmanager.DeviceManager{}, "GetDeviceAllErrorCode",
			func(_ *devmanager.DeviceManager, _ int32) (int32, []int64, error) {
				return 0, []int64{}, nil
			})
		patches.ApplyFuncReturn(hccn.GetNPULinkStatus, npuCommon.NPUNetworkLinkUpStatus, nil)
		patches.ApplyFunc(common.DoSaveDevFaultInfo, func(_ npuCommon.DevFaultInfo, _ bool) {})

		tool.generateNetworkFaultEventsBasedOnFaultCacheChange(device1)
		tool.generateNetworkFaultEventsBasedOnFaultCacheChange(device2)
	}
}

// TestCompareDeviceList for test compareDeviceList
func TestCompareDeviceList(t *testing.T) {
	convey.Convey("test compareDeviceList", t, func() {
		convey.Convey("01-deviceList and newDeviceList are both nil, should return true", func() {
			res := compareDeviceList(nil, nil)
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("02-deviceList is nil and newDeviceList is not nil, should return false", func() {
			res := compareDeviceList(nil, map[string]string{})
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("03-deviceList and newDeviceList length are different, should return false", func() {
			deviceList := map[string]string{"key1": "value1"}
			newDeviceList := map[string]string{"key1": "value1", "key2": "value2"}
			res := compareDeviceList(deviceList, newDeviceList)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("04-deviceList and newDeviceList only key are different, should return false", func() {
			deviceList := map[string]string{"key1": "value"}
			newDeviceList := map[string]string{"key2": "value"}
			res := compareDeviceList(deviceList, newDeviceList)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("05-deviceList and newDeviceList only value are different, should return false", func() {
			deviceList := map[string]string{"key": "value1"}
			newDeviceList := map[string]string{"key": "value2"}
			res := compareDeviceList(deviceList, newDeviceList)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("06-deviceList and newDeviceList only fault_time are different, should return true", func() {
			deviceList := map[string]string{`key`: `key:value,"fault_time":11111,key:value`}
			newDeviceList := map[string]string{`key`: `key:value,"fault_time":22222,key:value`}
			res := compareDeviceList(deviceList, newDeviceList)
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("07-deviceList and newDeviceList are same, should return true", func() {
			deviceList := map[string]string{"key": "value"}
			newDeviceList := map[string]string{"key": "value"}
			res := compareDeviceList(deviceList, newDeviceList)
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}

func TestGetUsedDevices(t *testing.T) {
	tool := &AscendTools{}
	tests := []struct {
		name           string
		input          []*common.NpuDevice
		expectedOutput sets.String
	}{
		{
			name:           "01-nil input, should return emtpy sets.String",
			input:          nil,
			expectedOutput: sets.NewString(),
		},
		{
			name: "02-no devices used, should return emtpy sets.String",
			input: []*common.NpuDevice{
				{DeviceName: "device0", PodUsed: false},
				{DeviceName: "device1", PodUsed: false},
			},
			expectedOutput: sets.NewString(),
		},
		{
			name: "03-some devices used, , should return used devices",
			input: []*common.NpuDevice{
				{DeviceName: "device0", PodUsed: true},
				{DeviceName: "device1", PodUsed: false},
				{DeviceName: "device2", PodUsed: true},
			},
			expectedOutput: sets.NewString("device0", "device2"),
		},
	}

	convey.Convey("getUsedDevices tests", t, func() {
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				result := tool.getUsedDevices(tt.input)
				convey.So(result, convey.ShouldEqual, tt.expectedOutput)
			})
		}
	})
}

func TestGetCurDeviceFaultCode(t *testing.T) {
	convey.Convey("test getCurDeviceFaultCode", t, func() {
		tool := mockAscendTools()
		convey.Convey("when devFaultInfo is empty, should return empty result", func() {
			res := tool.getCurDeviceFaultCode(0, []npuCommon.DevFaultInfo{})
			convey.So(res.Len(), convey.ShouldEqual, 0)
		})
		devFaultInfo := []npuCommon.DevFaultInfo{
			{LogicID: 0, Assertion: npuCommon.FaultRecover, EventID: common.LinkDownFaultCode},
		}
		convey.Convey("when dcmi get device fault code failed, should return empty result", func() {
			mockMethod := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
				"GetDeviceAllErrorCode", int32(0), []int64{}, errors.New("failed"))
			defer mockMethod.Reset()
			res := tool.getCurDeviceFaultCode(0, devFaultInfo)
			convey.So(res.Len(), convey.ShouldEqual, 0)
		})
		convey.Convey("when dcmi get device fault code success, should return fault code sets", func() {
			mockMethod := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
				"GetDeviceAllErrorCode", int32(1), []int64{common.LinkDownFaultCode}, nil)
			defer mockMethod.Reset()
			res := tool.getCurDeviceFaultCode(0, devFaultInfo)
			convey.So(res.Len(), convey.ShouldEqual, 1)
		})
	})
}
