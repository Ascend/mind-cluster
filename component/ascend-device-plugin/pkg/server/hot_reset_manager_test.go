/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package server

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"Ascend-device-plugin/pkg/plugin"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	npuCommon "ascend-common/devmanager/common"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		return
	}
}

func TestNewUnifiedHotResetManager(t *testing.T) {
	convey.Convey("test NewUnifiedHotResetManager", t, func() {
		convey.Convey("should return manager with initialized fields", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			convey.So(mgr, convey.ShouldNotBeNil)
			convey.So(mgr.tokenBucket, convey.ShouldNotBeNil)
			convey.So(mgr.idleTimeMgr, convey.ShouldNotBeNil)
		})
		convey.Convey("nil dmgr should still return manager", func() {
			mgr := NewUnifiedHotResetManager(nil, &device.HwAscend910Manager{}, nil)
			convey.So(mgr, convey.ShouldNotBeNil)
		})
	})
}

func TestUnifiedHotResetManagerNilReceiver(t *testing.T) {
	convey.Convey("test nil receiver", t, func() {
		convey.Convey("UnifiedHotReset on nil should not panic", func() {
			var mgr *UnifiedHotResetManager
			mgr.UnifiedHotReset(nil)
		})
	})
}

func TestUnifiedHotResetHotResetClose(t *testing.T) {
	convey.Convey("test hot reset close", t, func() {
		convey.Convey("HotReset close should return early", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			oldHotReset := common.ParamOption.HotReset
			common.ParamOption.HotReset = common.HotResetClose
			defer func() { common.ParamOption.HotReset = oldHotReset }()
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": {{LogicID: 0, Health: v1beta1.Unhealthy}},
			}
			mgr.UnifiedHotReset(groupDevice)
		})
	})
}

func TestFilterFaultDevices(t *testing.T) {
	convey.Convey("test filterFaultDevices", t, func() {
		convey.Convey("should exclude NotHandleFault and PreSeparateNPU", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			patch := gomonkey.ApplyFuncReturn(common.GetFaultType, common.NotHandleFault)
			defer patch.Reset()
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": {{LogicID: 0, Health: v1beta1.Unhealthy, FaultCodes: []int64{1}}},
			}
			result := mgr.filterFaultDevices(groupDevice)
			convey.So(len(result), convey.ShouldEqual, 0)
		})
		convey.Convey("should include RestartNPU fault devices", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			patch := gomonkey.ApplyFuncReturn(common.GetFaultType, common.RestartNPU)
			defer patch.Reset()
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": {{LogicID: 0, Health: v1beta1.Unhealthy, FaultCodes: []int64{1}}},
			}
			result := mgr.filterFaultDevices(groupDevice)
			convey.So(len(result), convey.ShouldEqual, 1)
		})
	})
}

func TestFilterFaultDevicesSeparateNPU(t *testing.T) {
	convey.Convey("test filterFaultDevices with SeparateNPU", t, func() {
		convey.Convey("SeparateNPU should mark need external ops", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			patch := gomonkey.ApplyFuncReturn(common.GetFaultType, common.SeparateNPU).
				ApplyFunc(device.WriteResetInfo,
					func(_ device.ResetInfo, _ device.WriteMode, _ bool) {})
			defer patch.Reset()
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": {{LogicID: 0, Health: v1beta1.Unhealthy, FaultCodes: []int64{1}}},
			}
			result := mgr.filterFaultDevices(groupDevice)
			convey.So(len(result), convey.ShouldEqual, 0)
		})
		convey.Convey("virtual device should be skipped", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910-2c-100-0": {{LogicID: 0, Health: v1beta1.Unhealthy}},
			}
			result := mgr.filterFaultDevices(groupDevice)
			convey.So(len(result), convey.ShouldEqual, 0)
		})
	})
}

func TestGetRingSize(t *testing.T) {
	convey.Convey("test getRingSize", t, func() {
		mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
			&device.HwAscend910Manager{}, nil)
		convey.Convey("Ascend910A should return 4", func() {
			oldCardType := common.ParamOption.RealCardType
			common.ParamOption.RealCardType = api.Ascend910A
			defer func() { common.ParamOption.RealCardType = oldCardType }()
			convey.So(mgr.getRingSize(&common.NpuDevice{}, 0, 0), convey.ShouldEqual, common.Ascend910RingsNum)
		})
		convey.Convey("910B infer board should return 1", func() {
			oldCardType := common.ParamOption.RealCardType
			common.ParamOption.RealCardType = api.Ascend910B
			defer func() { common.ParamOption.RealCardType = oldCardType }()
			convey.So(mgr.getRingSize(&common.NpuDevice{}, common.A300IA2BoardId, 4), convey.ShouldEqual, 1)
		})
		convey.Convey("910B large device count should return 16", func() {
			oldCardType := common.ParamOption.RealCardType
			common.ParamOption.RealCardType = api.Ascend910B
			defer func() { common.ParamOption.RealCardType = oldCardType }()
			convey.So(mgr.getRingSize(&common.NpuDevice{}, 0, 16), convey.ShouldEqual, common.A200TA2RingsNum)
		})
		convey.Convey("910B train should return 8", func() {
			oldCardType := common.ParamOption.RealCardType
			common.ParamOption.RealCardType = api.Ascend910B
			defer func() { common.ParamOption.RealCardType = oldCardType }()
			convey.So(mgr.getRingSize(&common.NpuDevice{}, 0, 8), convey.ShouldEqual, common.Ascend910BRingsNumTrain)
		})
		convey.Convey("unknown type should return 1", func() {
			oldCardType := common.ParamOption.RealCardType
			common.ParamOption.RealCardType = "unknown"
			defer func() { common.ParamOption.RealCardType = oldCardType }()
			convey.So(mgr.getRingSize(&common.NpuDevice{}, 0, 0), convey.ShouldEqual, 1)
		})
	})
}

func TestGetHccsRingDevices(t *testing.T) {
	convey.Convey("test getHccsRingDevices", t, func() {
		mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
			&device.HwAscend910Manager{}, nil)
		convey.Convey("should return devices in same ring", func() {
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": {
					{LogicID: 0}, {LogicID: 1}, {LogicID: 2}, {LogicID: 3},
					{LogicID: 4},
				},
			}
			ringDevs, faultDev := mgr.getHccsRingDevices(&common.NpuDevice{LogicID: 1}, 4, groupDevice)
			convey.So(len(ringDevs), convey.ShouldEqual, 4)
			convey.So(faultDev.LogicID, convey.ShouldEqual, 1)
		})
		convey.Convey("ring size 1 should return single device", func() {
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": {{LogicID: 5}},
			}
			ringDevs, faultDev := mgr.getHccsRingDevices(&common.NpuDevice{LogicID: 5}, 1, groupDevice)
			convey.So(len(ringDevs), convey.ShouldEqual, 1)
			convey.So(faultDev.LogicID, convey.ShouldEqual, 5)
		})
	})
}

func TestCheckDeviceRecovered(t *testing.T) {
	convey.Convey("test checkDeviceRecovered", t, func() {
		convey.Convey("healthy device should clear idle time", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			mgr.idleTimeMgr.RecordIdleTime(0)
			patch := gomonkey.ApplyFuncReturn(device.ReadResetInfo, device.ResetInfo{}).
				ApplyFunc(device.WriteResetInfo,
					func(_ device.ResetInfo, _ device.WriteMode, _ bool) {})
			defer patch.Reset()
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": {{LogicID: 0, Health: v1beta1.Healthy}},
			}
			mgr.checkDeviceRecovered(groupDevice)
			_, ok := mgr.idleTimeMgr.GetIdleTime(0)
			convey.So(ok, convey.ShouldBeFalse)
		})
		convey.Convey("unhealthy device should keep idle time", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			mgr.idleTimeMgr.RecordIdleTime(0)
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": {{LogicID: 0, Health: v1beta1.Unhealthy}},
			}
			mgr.checkDeviceRecovered(groupDevice)
			_, ok := mgr.idleTimeMgr.GetIdleTime(0)
			convey.So(ok, convey.ShouldBeTrue)
		})
	})
}

func TestConfirmAndPrepareReset(t *testing.T) {
	convey.Convey("test confirmAndPrepareReset", t, func() {
		convey.Convey("device in resetting should return false", func() {
			dmgr := &devmanager.DeviceManagerMock{}
			mgr := NewUnifiedHotResetManager(dmgr, &device.HwAscend910Manager{}, nil)
			patch := gomonkey.ApplyMethodReturn(&device.HwAscend910Manager{}, "GetIfCardsInResetting", true)
			defer patch.Reset()
			ringDevs := []*common.NpuDevice{{LogicID: 0}}
			result := mgr.confirmAndPrepareReset(ringDevs, ringDevs[0])
			convey.So(result, convey.ShouldBeFalse)
		})
		convey.Convey("healthy device should return false", func() {
			dmgr := &devmanager.DeviceManagerMock{}
			mgr := NewUnifiedHotResetManager(dmgr, &device.HwAscend910Manager{}, nil)
			patch := gomonkey.ApplyMethodReturn(&device.HwAscend910Manager{}, "GetIfCardsInResetting", false).
				ApplyMethodReturn(dmgr, "GetDeviceHealth", uint32(0), nil)
			defer patch.Reset()
			ringDevs := []*common.NpuDevice{{LogicID: 0}}
			result := mgr.confirmAndPrepareReset(ringDevs, ringDevs[0])
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestConfirmAndPrepareResetSuccess(t *testing.T) {
	convey.Convey("test confirmAndPrepareReset success", t, func() {
		convey.Convey("unhealthy device should consume token and set resetting", func() {
			dmgr := &devmanager.DeviceManagerMock{}
			devMgr := &device.HwAscend910Manager{}
			mgr := NewUnifiedHotResetManager(dmgr, devMgr, nil)
			patch := gomonkey.ApplyMethodReturn(devMgr, "GetIfCardsInResetting", false).
				ApplyMethodReturn(dmgr, "GetDeviceHealth", uint32(1), nil).
				ApplyMethod(devMgr, "SetCardsInResetting",
					func(_ *device.HwAscend910Manager, _ int32, _ bool) {})
			defer patch.Reset()
			ringDevs := []*common.NpuDevice{{LogicID: 0}}
			faultDev := ringDevs[0]
			result := mgr.confirmAndPrepareReset(ringDevs, faultDev)
			convey.So(result, convey.ShouldBeTrue)
			convey.So(mgr.tokenBucket.GetTokens(0), convey.ShouldEqual, tokenMaxCount-1)
		})
	})
}

func TestConvertToResetDevices(t *testing.T) {
	convey.Convey("test convertToResetDevices", t, func() {
		convey.Convey("should convert NpuDevice to ResetDevice", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			devs := []*common.NpuDevice{
				{LogicID: 0, CardID: 1, DeviceID: 2, PhyID: 3},
			}
			result := mgr.convertToResetDevices(devs)
			convey.So(len(result), convey.ShouldEqual, 1)
			convey.So(result[0].LogicID, convey.ShouldEqual, 0)
			convey.So(result[0].CardID, convey.ShouldEqual, 1)
		})
		convey.Convey("empty input should return empty slice", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			result := mgr.convertToResetDevices(nil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})
	})
}

func TestExecDriverReset(t *testing.T) {
	convey.Convey("test execDriverReset", t, func() {
		convey.Convey("reset success should return nil", func() {
			dmgr := &devmanager.DeviceManagerMock{}
			mgr := NewUnifiedHotResetManager(dmgr, &device.HwAscend910Manager{}, nil)
			patch := gomonkey.ApplyMethodReturn(dmgr, "SetDeviceReset", nil).
				ApplyMethodReturn(dmgr, "GetDeviceBootStatus", common.BootStartFinish, nil)
			defer patch.Reset()
			ringDevs := []*common.NpuDevice{{LogicID: 0}}
			err := mgr.execDriverReset(0, ringDevs)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("reset failed should return error", func() {
			dmgr := &devmanager.DeviceManagerMock{}
			mgr := NewUnifiedHotResetManager(dmgr, &device.HwAscend910Manager{}, nil)
			patch := gomonkey.ApplyMethodReturn(dmgr, "SetDeviceReset", errors.New("reset failed"))
			defer patch.Reset()
			ringDevs := []*common.NpuDevice{{LogicID: 0}}
			err := mgr.execDriverReset(0, ringDevs)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestSetPluginManager(t *testing.T) {
	convey.Convey("test SetPluginManager", t, func() {
		convey.Convey("should set plugin manager", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			pm := plugin.NewPluginManager()
			mgr.SetPluginManager(pm)
			convey.So(mgr.pluginMgr, convey.ShouldEqual, pm)
		})
	})
}

func TestGetA3AssociatedDevices(t *testing.T) {
	convey.Convey("test getA3AssociatedDevices", t, func() {
		convey.Convey("GetCardIDDeviceID failed should return nil", func() {
			dmgr := &devmanager.DeviceManagerMock{}
			mgr := NewUnifiedHotResetManager(dmgr, &device.HwAscend910Manager{}, nil)
			patch := gomonkey.ApplyMethodReturn(dmgr, "GetCardIDDeviceID",
				int32(0), int32(0), errors.New("error"))
			defer patch.Reset()
			ringDevs, _ := mgr.getA3AssociatedDevices(&common.NpuDevice{LogicID: 0}, nil)
			convey.So(ringDevs, convey.ShouldBeNil)
		})
		convey.Convey("GetAssociatedLogicIDs failed should return nil", func() {
			dmgr := &devmanager.DeviceManagerMock{}
			devMgr := &device.HwAscend910Manager{}
			mgr := NewUnifiedHotResetManager(dmgr, devMgr, nil)
			patch := gomonkey.ApplyMethodReturn(dmgr, "GetCardIDDeviceID", int32(0), int32(0), nil).
				ApplyMethodReturn(devMgr, "GetAssociatedLogicIDs", []int32{}, errors.New("error"))
			defer patch.Reset()
			ringDevs, _ := mgr.getA3AssociatedDevices(&common.NpuDevice{LogicID: 0}, nil)
			convey.So(ringDevs, convey.ShouldBeNil)
		})
	})
}

func TestIsRingFreeWithPod(t *testing.T) {
	convey.Convey("test isRingFree with pod", t, func() {
		convey.Convey("device with pod should return false", func() {
			dmgr := &devmanager.DeviceManagerMock{}
			mgr := NewUnifiedHotResetManager(dmgr, &device.HwAscend910Manager{}, nil)
			ringDevs := []*common.NpuDevice{{LogicID: 0, PodUsed: true}}
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": ringDevs,
			}
			result := mgr.isRingFree(ringDevs, &PodResource{}, groupDevice)
			convey.So(result, convey.ShouldBeFalse)
		})
		convey.Convey("device with process should return false", func() {
			dmgr := &devmanager.DeviceManagerMock{}
			mgr := NewUnifiedHotResetManager(dmgr, &device.HwAscend910Manager{}, nil)
			patch := gomonkey.ApplyMethodReturn(dmgr, "GetDevProcessInfo",
				&npuCommon.DevProcessInfo{ProcNum: 1}, nil)
			defer patch.Reset()
			ringDevs := []*common.NpuDevice{{LogicID: 0, PodUsed: false}}
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": ringDevs,
			}
			result := mgr.isRingFree(ringDevs, &PodResource{}, groupDevice)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestIsRingFreeNoProcess(t *testing.T) {
	convey.Convey("test isRingFree no process", t, func() {
		convey.Convey("device without pod and process should return true", func() {
			dmgr := &devmanager.DeviceManagerMock{}
			mgr := NewUnifiedHotResetManager(dmgr, &device.HwAscend910Manager{}, nil)
			patch := gomonkey.ApplyMethodReturn(dmgr, "GetDevProcessInfo",
				&npuCommon.DevProcessInfo{ProcNum: 0}, nil)
			defer patch.Reset()
			ringDevs := []*common.NpuDevice{{LogicID: 0, PodUsed: false}}
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": ringDevs,
			}
			result := mgr.isRingFree(ringDevs, &PodResource{}, groupDevice)
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

func TestGetDuoCardDevices(t *testing.T) {
	convey.Convey("test getDuoCardDevices", t, func() {
		mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
			&device.HwAscend910Manager{}, nil)
		convey.Convey("should return devices with same CardID", func() {
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend310P": {
					{LogicID: 0, CardID: 1}, {LogicID: 1, CardID: 1}, {LogicID: 2, CardID: 2},
				},
			}
			ringDevs, faultDev := mgr.getDuoCardDevices(
				&common.NpuDevice{LogicID: 0, CardID: 1}, groupDevice)
			convey.So(len(ringDevs), convey.ShouldEqual, 2)
			convey.So(faultDev.LogicID, convey.ShouldEqual, 0)
		})
	})
}

func TestExecuteResetSuccess(t *testing.T) {
	convey.Convey("test executeReset success", t, func() {
		convey.Convey("successful reset should clear idle time and set device init", func() {
			dmgr := &devmanager.DeviceManagerMock{}
			devMgr := &device.HwAscend910Manager{}
			mgr := NewUnifiedHotResetManager(dmgr, devMgr, nil)
			mgr.idleTimeMgr.RecordIdleTime(0)
			resettingSetFalse := false
			deviceInitCalled := false
			patch := gomonkey.ApplyMethodReturn(dmgr, "SetDeviceReset", nil).
				ApplyMethodReturn(dmgr, "GetDeviceBootStatus", common.BootStartFinish, nil).
				ApplyMethod(devMgr, "SetCardsInResetting",
					func(_ *device.HwAscend910Manager, id int32, v bool) {
						if !v && id == 0 {
							resettingSetFalse = true
						}
					}).
				ApplyFunc(common.SetDeviceInit, func(id int32) {
					if id == 0 {
						deviceInitCalled = true
					}
				})
			defer patch.Reset()
			ringDevs := []*common.NpuDevice{{LogicID: 0, CardID: 0, DeviceID: 0, PhyID: 0}}
			faultDev := ringDevs[0]
			mgr.executeReset(ringDevs, faultDev)
			_, ok := mgr.idleTimeMgr.GetIdleTime(0)
			convey.So(ok, convey.ShouldBeFalse)
			convey.So(resettingSetFalse, convey.ShouldBeTrue)
			convey.So(deviceInitCalled, convey.ShouldBeTrue)
		})
	})
}

func TestExecuteResetWithPluginManager(t *testing.T) {
	convey.Convey("test executeReset with plugin manager", t, func() {
		convey.Convey("plugin manager should execute hooks and reset state", func() {
			dmgr := &devmanager.DeviceManagerMock{}
			devMgr := &device.HwAscend910Manager{}
			mgr := NewUnifiedHotResetManager(dmgr, devMgr, nil)
			pm := plugin.NewPluginManager()
			mgr.SetPluginManager(pm)
			resettingSetFalse := false
			patch := gomonkey.ApplyMethodReturn(dmgr, "SetDeviceReset", nil).
				ApplyMethodReturn(dmgr, "GetDeviceBootStatus", common.BootStartFinish, nil).
				ApplyMethod(devMgr, "SetCardsInResetting",
					func(_ *device.HwAscend910Manager, _ int32, v bool) {
						if !v {
							resettingSetFalse = true
						}
					}).
				ApplyFunc(common.SetDeviceInit, func(_ int32) {})
			defer patch.Reset()
			ringDevs := []*common.NpuDevice{{LogicID: 0, CardID: 0, DeviceID: 0, PhyID: 0}}
			faultDev := ringDevs[0]
			mgr.executeReset(ringDevs, faultDev)
			convey.So(resettingSetFalse, convey.ShouldBeTrue)
		})
	})
}

func TestFindFaultDevType(t *testing.T) {
	convey.Convey("test findFaultDevType", t, func() {
		mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
			&device.HwAscend910Manager{}, nil)
		convey.Convey("should find dev type for fault device", func() {
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": {{LogicID: 0, CardID: 1}, {LogicID: 1, CardID: 2}},
				"Ascend310": {{LogicID: 2, CardID: 3}},
			}
			devType := mgr.findFaultDevType(&common.NpuDevice{LogicID: 1}, groupDevice)
			convey.So(devType, convey.ShouldEqual, "Ascend910")
		})
		convey.Convey("should return empty for unknown device", func() {
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": {{LogicID: 0}},
			}
			devType := mgr.findFaultDevType(&common.NpuDevice{LogicID: 99}, groupDevice)
			convey.So(devType, convey.ShouldEqual, "")
		})
	})
}

func TestGetResetRingDevices(t *testing.T) {
	convey.Convey("test getResetRingDevices", t, func() {
		convey.Convey("ringSize 1 should return single device", func() {
			devMgr := &device.HwAscend910Manager{}
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{}, devMgr, nil)
			oldCardType := common.ParamOption.RealCardType
			common.ParamOption.RealCardType = "unknown"
			defer func() { common.ParamOption.RealCardType = oldCardType }()
			patch := gomonkey.ApplyMethodReturn(devMgr, "GetServerBoardId", uint32(0), nil)
			defer patch.Reset()
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": {{LogicID: 0}},
			}
			faultDev := &common.NpuDevice{LogicID: 0}
			ringDevs, _ := mgr.getResetRingDevices(faultDev, groupDevice)
			convey.So(len(ringDevs), convey.ShouldEqual, 1)
		})
	})
}

func TestConfirmAndPrepareReset_DeviceNotReady(t *testing.T) {
	convey.Convey("test confirmAndPrepareReset device not ready", t, func() {
		convey.Convey("device not ready error should continue reset", func() {
			dmgr := &devmanager.DeviceManagerMock{}
			devMgr := &device.HwAscend910Manager{}
			mgr := NewUnifiedHotResetManager(dmgr, devMgr, nil)
			patch := gomonkey.ApplyMethodReturn(devMgr, "GetIfCardsInResetting", false).
				ApplyMethodReturn(dmgr, "GetDeviceHealth", uint32(0), errors.New("error code -8012")).
				ApplyMethod(devMgr, "SetCardsInResetting",
					func(_ *device.HwAscend910Manager, _ int32, _ bool) {})
			defer patch.Reset()
			ringDevs := []*common.NpuDevice{{LogicID: 0}}
			result := mgr.confirmAndPrepareReset(ringDevs, ringDevs[0])
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

func TestFilterOpsDevices(t *testing.T) {
	convey.Convey("test filterOpsDevices", t, func() {
		convey.Convey("should filter only devices present in resetInfo", func() {
			resetInfo := device.ResetInfo{
				ManualResetDevs: []device.ResetDevice{
					{LogicID: 0}, {LogicID: 1},
				},
			}
			needClear := filterOpsDevices(resetInfo, []int32{0, 1, 2, 3})
			convey.So(len(needClear), convey.ShouldEqual, 2)
			convey.So(needClear[0], convey.ShouldEqual, 0)
			convey.So(needClear[1], convey.ShouldEqual, 1)
		})
		convey.Convey("should return empty when no match", func() {
			resetInfo := device.ResetInfo{}
			needClear := filterOpsDevices(resetInfo, []int32{0, 1})
			convey.So(len(needClear), convey.ShouldEqual, 0)
		})
	})
}

func TestMarkNeedExternalOps(t *testing.T) {
	convey.Convey("test markNeedExternalOps", t, func() {
		convey.Convey("should write reset info with WMAppend", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			dev := &common.NpuDevice{
				LogicID:  7,
				CardID:   0,
				DeviceID: 1,
				PhyID:    10,
			}
			var capturedInfo device.ResetInfo
			var capturedMode device.WriteMode
			patch := gomonkey.ApplyFunc(device.WriteResetInfo,
				func(info device.ResetInfo, mode device.WriteMode, updateNode bool) {
					capturedInfo = info
					capturedMode = mode
				})
			defer patch.Reset()
			mgr.markNeedExternalOps(dev)
			convey.So(capturedMode, convey.ShouldEqual, device.WMAppend)
			convey.So(len(capturedInfo.ManualResetDevs), convey.ShouldEqual, 1)
			convey.So(capturedInfo.ManualResetDevs[0].LogicID, convey.ShouldEqual, 7)
			convey.So(capturedInfo.ManualResetDevs[0].CardId, convey.ShouldEqual, 0)
			convey.So(capturedInfo.ManualResetDevs[0].DeviceId, convey.ShouldEqual, 1)
			convey.So(capturedInfo.ManualResetDevs[0].PhyID, convey.ShouldEqual, 10)
			convey.So(len(capturedInfo.ThirdPartyResetDevs), convey.ShouldEqual, 0)
		})
		convey.Convey("should write device with default CardID", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			dev := &common.NpuDevice{
				LogicID:  3,
				CardID:   -1,
				DeviceID: -1,
				PhyID:    5,
			}
			var capturedInfo device.ResetInfo
			patch := gomonkey.ApplyFunc(device.WriteResetInfo,
				func(info device.ResetInfo, mode device.WriteMode, updateNode bool) {
					capturedInfo = info
				})
			defer patch.Reset()
			mgr.markNeedExternalOps(dev)
			convey.So(capturedInfo.ManualResetDevs[0].CardId, convey.ShouldEqual, -1)
			convey.So(capturedInfo.ManualResetDevs[0].DeviceId, convey.ShouldEqual, -1)
			convey.So(capturedInfo.ManualResetDevs[0].PhyID, convey.ShouldEqual, 5)
		})
	})
}

func TestClearNeedExternalOps1(t *testing.T) {
	convey.Convey("test clearNeedExternalOps", t, func() {
		convey.Convey("no devices in reset info should return early", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			devs := []*common.NpuDevice{
				{LogicID: 0, CardID: 0, DeviceID: 0, PhyID: 0},
			}
			writeCalled := false
			patch := gomonkey.ApplyFuncReturn(device.ReadResetInfo, device.ResetInfo{}).
				ApplyFunc(device.WriteResetInfo,
					func(_ device.ResetInfo, _ device.WriteMode, _ bool) {
						writeCalled = true
					})
			defer patch.Reset()
			mgr.clearNeedExternalOps(devs)
			convey.So(writeCalled, convey.ShouldBeFalse)
		})
		convey.Convey("device in ManualResetDevs should be cleared", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			devs := []*common.NpuDevice{
				{LogicID: 7, CardID: 0, DeviceID: 1, PhyID: 10},
			}
			resetInfo := device.ResetInfo{
				ManualResetDevs: []device.ResetDevice{
					{LogicID: 7, CardId: 0, DeviceId: 1, PhyID: 10},
				},
			}
			var capturedDelInfo device.ResetInfo
			var capturedMode device.WriteMode
			patch := gomonkey.ApplyFuncReturn(device.ReadResetInfo, resetInfo).
				ApplyFunc(device.WriteResetInfo,
					func(info device.ResetInfo, mode device.WriteMode, updateNode bool) {
						capturedDelInfo = info
						capturedMode = mode
					})
			defer patch.Reset()
			mgr.clearNeedExternalOps(devs)
			convey.So(capturedMode, convey.ShouldEqual, device.WMDelete)
			convey.So(len(capturedDelInfo.ManualResetDevs), convey.ShouldEqual, 1)
			convey.So(capturedDelInfo.ManualResetDevs[0].LogicID, convey.ShouldEqual, 7)
			convey.So(len(capturedDelInfo.ThirdPartyResetDevs), convey.ShouldEqual, 1)
			convey.So(capturedDelInfo.ThirdPartyResetDevs[0].LogicID, convey.ShouldEqual, 7)
		})
	})
}

func TestClearNeedExternalOps2(t *testing.T) {
	convey.Convey("test clearNeedExternalOps", t, func() {
		convey.Convey("device in ThirdPartyResetDevs should be cleared", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			devs := []*common.NpuDevice{
				{LogicID: 2, CardID: 1, DeviceID: 0, PhyID: 3},
			}
			resetInfo := device.ResetInfo{
				ThirdPartyResetDevs: []device.ResetDevice{
					{LogicID: 2, CardId: 1, DeviceId: 0, PhyID: 3},
				},
			}
			var capturedDelInfo device.ResetInfo
			patch := gomonkey.ApplyFuncReturn(device.ReadResetInfo, resetInfo).
				ApplyFunc(device.WriteResetInfo,
					func(info device.ResetInfo, mode device.WriteMode, updateNode bool) {
						capturedDelInfo = info
					})
			defer patch.Reset()
			mgr.clearNeedExternalOps(devs)
			convey.So(len(capturedDelInfo.ThirdPartyResetDevs), convey.ShouldEqual, 1)
			convey.So(capturedDelInfo.ThirdPartyResetDevs[0].LogicID, convey.ShouldEqual, 2)
			convey.So(len(capturedDelInfo.ManualResetDevs), convey.ShouldEqual, 1)
		})
		convey.Convey("device not in reset info should not trigger write", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			devs := []*common.NpuDevice{
				{LogicID: 5, CardID: 0, DeviceID: 0, PhyID: 5},
			}
			resetInfo := device.ResetInfo{
				ManualResetDevs: []device.ResetDevice{
					{LogicID: 3, CardId: 0, DeviceId: 0, PhyID: 3},
				},
			}
			writeCalled := false
			patch := gomonkey.ApplyFuncReturn(device.ReadResetInfo, resetInfo).
				ApplyFunc(device.WriteResetInfo,
					func(_ device.ResetInfo, _ device.WriteMode, _ bool) {
						writeCalled = true
					})
			defer patch.Reset()
			mgr.clearNeedExternalOps(devs)
			convey.So(writeCalled, convey.ShouldBeFalse)
		})
	})
}

func TestClearNeedExternalOps3(t *testing.T) {
	convey.Convey("test clearNeedExternalOps", t, func() {
		convey.Convey("multiple devices partial match should only clear matched ones", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			devs := []*common.NpuDevice{
				{LogicID: 0, CardID: 0, DeviceID: 0, PhyID: 0},
				{LogicID: 1, CardID: 0, DeviceID: 1, PhyID: 1},
			}
			resetInfo := device.ResetInfo{
				ManualResetDevs: []device.ResetDevice{
					{LogicID: 0, CardId: 0, DeviceId: 0, PhyID: 0},
					{LogicID: 1, CardId: 0, DeviceId: 1, PhyID: 1},
				},
				ThirdPartyResetDevs: []device.ResetDevice{
					{LogicID: 0, CardId: 0, DeviceId: 0, PhyID: 0},
				},
			}
			var capturedDelInfo device.ResetInfo
			patch := gomonkey.ApplyFuncReturn(device.ReadResetInfo, resetInfo).
				ApplyFunc(device.WriteResetInfo,
					func(info device.ResetInfo, mode device.WriteMode, updateNode bool) {
						capturedDelInfo = info
					})
			defer patch.Reset()
			mgr.clearNeedExternalOps(devs)
			convey.So(len(capturedDelInfo.ManualResetDevs), convey.ShouldEqual, 2)
			convey.So(len(capturedDelInfo.ThirdPartyResetDevs), convey.ShouldEqual, 2)
		})
		convey.Convey("empty devs slice should return early", func() {
			mgr := NewUnifiedHotResetManager(&devmanager.DeviceManagerMock{},
				&device.HwAscend910Manager{}, nil)
			writeCalled := false
			patch := gomonkey.ApplyFuncReturn(device.ReadResetInfo, device.ResetInfo{}).
				ApplyFunc(device.WriteResetInfo,
					func(_ device.ResetInfo, _ device.WriteMode, _ bool) {
						writeCalled = true
					})
			defer patch.Reset()
			mgr.clearNeedExternalOps(nil)
			convey.So(writeCalled, convey.ShouldBeFalse)
		})
	})
}
