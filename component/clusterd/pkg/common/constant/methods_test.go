// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package constant a series of para
package constant

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/util"
)

const (
	device0 = "device-0"
	device1 = "device-1"
)

func TestMain(m *testing.M) {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	m.Run()
}

func TestAdvanceDeviceFaultCmAddFaultAndFix1(t *testing.T) {
	convey.Convey("Given an AdvanceDeviceFaultCm instance", t, func() {
		cm := &AdvanceDeviceFaultCm{
			FaultDeviceList:     make(map[string][]DeviceFault),
			AvailableDeviceList: []string{"npu0", "npu1"},
			CardUnHealthy:       []string{},
			NetworkUnhealthy:    []string{},
		}

		convey.Convey("When adding a new fault with normal fault level", func() {
			normalFault := DeviceFault{
				NPUName:    "npu0",
				FaultType:  CardUnhealthy,
				FaultLevel: NormalNPU,
			}
			cm.AddFaultAndFix(normalFault)

			convey.Convey("It should add the fault but not modify status lists", func() {
				convey.So(cm.FaultDeviceList["npu0"], convey.ShouldHaveLength, 1)
				convey.So(cm.AvailableDeviceList, convey.ShouldContain, "npu0")
				convey.So(cm.CardUnHealthy, convey.ShouldBeEmpty)
				convey.So(cm.NetworkUnhealthy, convey.ShouldBeEmpty)
			})
		})

		convey.Convey("When adding a duplicate fault", func() {
			existingFault := DeviceFault{
				NPUName:    "npu1",
				FaultType:  CardUnhealthy,
				FaultLevel: "critical",
			}
			cm.FaultDeviceList["npu1"] = []DeviceFault{existingFault}

			cm.AddFaultAndFix(existingFault)

			convey.Convey("It should not modify any lists", func() {
				convey.So(cm.FaultDeviceList["npu1"], convey.ShouldHaveLength, 1)
				convey.So(cm.AvailableDeviceList, convey.ShouldContain, "npu1")
			})
		})
	})
}

func TestAdvanceDeviceFaultCmAddFaultAndFix2(t *testing.T) {
	convey.Convey("Given an AdvanceDeviceFaultCm instance", t, func() {
		cm := &AdvanceDeviceFaultCm{
			FaultDeviceList:     make(map[string][]DeviceFault),
			AvailableDeviceList: []string{"npu0", "npu1"},
			CardUnHealthy:       []string{},
			NetworkUnhealthy:    []string{},
		}

		convey.Convey("When adding a new CardUnhealthy fault with critical level", func() {
			criticalFault := DeviceFault{
				NPUName:    "npu0",
				FaultType:  CardUnhealthy,
				FaultLevel: RestartBusiness,
			}
			cm.AddFaultAndFix(criticalFault)

			convey.Convey("It should add the fault and update status lists", func() {
				convey.So(cm.FaultDeviceList["npu0"], convey.ShouldHaveLength, 1)
				convey.So(cm.AvailableDeviceList, convey.ShouldNotContain, "npu0")
				convey.So(cm.CardUnHealthy, convey.ShouldContain, "npu0")
				convey.So(cm.NetworkUnhealthy, convey.ShouldBeEmpty)
			})
		})

		convey.Convey("When adding a new CardNetworkUnhealthy fault with warning level", func() {
			networkFault := DeviceFault{
				NPUName:    "npu1",
				FaultType:  CardNetworkUnhealthy,
				FaultLevel: "warning",
			}
			cm.AddFaultAndFix(networkFault)

			convey.Convey("It should add the fault and update network status", func() {
				convey.So(cm.FaultDeviceList["npu1"], convey.ShouldHaveLength, 1)
				convey.So(cm.AvailableDeviceList, convey.ShouldNotContain, "npu1")
				convey.So(cm.CardUnHealthy, convey.ShouldBeEmpty)
				convey.So(cm.NetworkUnhealthy, convey.ShouldContain, "npu1")
			})
		})

		convey.Convey("When adding a fault to a new NPU not in available list", func() {
			newFault := DeviceFault{
				NPUName:    "npu3",
				FaultType:  CardUnhealthy,
				FaultLevel: "critical",
			}
			cm.AddFaultAndFix(newFault)

			convey.Convey("It should add the fault and update status", func() {
				convey.So(cm.FaultDeviceList["npu3"], convey.ShouldHaveLength, 1)
				convey.So(cm.AvailableDeviceList, convey.ShouldNotContain, "npu3")
				convey.So(cm.CardUnHealthy, convey.ShouldContain, "npu3")
			})
		})
	})
}

func TestAdvanceDeviceFaultCmDelFaultAndFix(t *testing.T) {
	cm := &AdvanceDeviceFaultCm{
		FaultDeviceList:  make(map[string][]DeviceFault),
		CardUnHealthy:    []string{},
		NetworkUnhealthy: []string{},
	}
	convey.Convey("When deleting a non-existent fault, "+
		"It should not modify the fault list or status lists", t, func() {
		cm.FaultDeviceList["npu0"] = []DeviceFault{{
			NPUName:    "npu0",
			FaultType:  CardNetworkUnhealthy,
			FaultLevel: SeparateNPU,
		}}
		cm.NetworkUnhealthy = []string{"npu0"}
		fault := DeviceFault{
			NPUName:    "npu0",
			FaultType:  CardUnhealthy,
			FaultLevel: SubHealthFault,
		}
		cm.DelFaultAndFix(fault)
		convey.So(cm.FaultDeviceList["npu0"], convey.ShouldHaveLength, 1)
		convey.So(cm.NetworkUnhealthy, convey.ShouldContain, "npu0")
	})

	convey.Convey("When deleting the last fault of a type from an NPU,"+
		"It should remove the fault and update CardUnHealthy", t, func() {
		fault1 := DeviceFault{
			NPUName:    "npu1",
			FaultType:  CardUnhealthy,
			FaultLevel: RestartBusiness,
		}
		fault2 := DeviceFault{
			NPUName:    "npu1",
			FaultType:  CardNetworkUnhealthy,
			FaultLevel: RestartBusiness,
		}
		fault3 := DeviceFault{
			NPUName:    "npu1",
			FaultType:  CardUnhealthy,
			FaultLevel: SubHealthFault,
		}
		cm.FaultDeviceList["npu1"] = []DeviceFault{fault1, fault2, fault3}
		cm.CardUnHealthy = []string{"npu1"}
		cm.NetworkUnhealthy = []string{"npu1"}
		cm.DelFaultAndFix(fault1)
		convey.So(cm.FaultDeviceList["npu1"], convey.ShouldHaveLength, 2)
		convey.So(cm.CardUnHealthy, convey.ShouldNotContain, "npu1")
		convey.So(cm.NetworkUnhealthy, convey.ShouldContain, "npu1")
	})
}

func mockAdvanceDeviceFaultCm() *AdvanceDeviceFaultCm {
	return &AdvanceDeviceFaultCm{
		DeviceType:          api.Ascend910,
		CmName:              "xxx",
		SuperPodID:          0,
		ServerIndex:         0,
		FaultDeviceList:     mockFaultDeviceList(device0),
		AvailableDeviceList: make([]string, 0),
		Recovering:          make([]string, 0),
		CardUnHealthy:       make([]string, 0),
		NetworkUnhealthy:    make([]string, 0),
		UpdateTime:          0,
	}
}

func mockFaultDeviceList(deviceName string) map[string][]DeviceFault {
	return map[string][]DeviceFault{
		deviceName: {
			{
				FaultType:            CardUnhealthy,
				FaultTimeAndLevelMap: mockFaultTimeAndLevelMap(UceFaultCode),
			},
		},
	}
}

func mockFaultTimeAndLevelMap(key string) map[string]FaultTimeAndLevel {
	return map[string]FaultTimeAndLevel{
		key: {
			FaultTime:         0,
			FaultReceivedTime: 0,
			FaultLevel:        SeparateNPU,
		},
	}
}

func TestAdvanceDeviceFaultCmIsSame(t *testing.T) {
	convey.Convey("two cm should be same", t, func() {
		cm1 := mockAdvanceDeviceFaultCm()
		cm2 := new(AdvanceDeviceFaultCm)
		util.DeepCopy(cm2, cm1)
		convey.So(cm2.IsSame(cm1), convey.ShouldBeTrue)
	})
}

func TestAdvanceDeviceFaultCmUpdateFaultReceiveTime(t *testing.T) {
	convey.Convey("Test AdvanceDeviceFaultCm UpdateFaultReceiveTime", t, func() {
		convey.Convey("When cm is nil, it should return directly", func() {
			var cm *AdvanceDeviceFaultCm = nil
			if cm != nil {
				cm.UpdateFaultReceiveTime(nil)
			}
			convey.So(cm, convey.ShouldBeNil)
		})
		convey.Convey("When oldInfo type conversion failed, "+
			"cm should use current time to update fault receive time", func() {
			cm := mockAdvanceDeviceFaultCm()
			oldInfo := new(SwitchInfo)
			cm.UpdateFaultReceiveTime(oldInfo)
			faultReceivedTime := cm.FaultDeviceList[device0][0].FaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime
			convey.So(faultReceivedTime, convey.ShouldBeGreaterThan, 0)
		})
		convey.Convey("When oldInfo is nil, cm should use current time to update fault receive time", func() {
			cm := mockAdvanceDeviceFaultCm()
			var oldInfo *AdvanceDeviceFaultCm = nil
			cm.UpdateFaultReceiveTime(oldInfo)
			faultReceivedTime := cm.FaultDeviceList[device0][0].FaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime
			convey.So(faultReceivedTime, convey.ShouldBeGreaterThan, 0)
		})
		convey.Convey("When oldInfo exists target fault, "+
			"cm should use the data in oldInfo to update fault receive time", func() {
			cm := mockAdvanceDeviceFaultCm()
			cm.FaultDeviceList[device0][0].FaultTimeAndLevelMap[AicFaultCode] = FaultTimeAndLevel{
				FaultTime:         0,
				FaultReceivedTime: 0,
				FaultLevel:        SeparateNPU,
			}
			oldInfo := mockAdvanceDeviceFaultCm()
			cm.UpdateFaultReceiveTime(oldInfo)
			faultReceivedTime1 := cm.FaultDeviceList[device0][0].FaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime
			faultReceivedTime2 := cm.FaultDeviceList[device0][0].FaultTimeAndLevelMap[AicFaultCode].FaultReceivedTime
			oldFaultReceivedTime1 :=
				oldInfo.FaultDeviceList[device0][0].FaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime
			convey.So(faultReceivedTime1, convey.ShouldEqual, oldFaultReceivedTime1)
			convey.So(faultReceivedTime2, convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestUpdateFaultReceiveTimeForDevices(t *testing.T) {
	convey.Convey("Test updateFaultReceiveTimeForDevices when oldFaults is nil", t, func() {
		cmFaults := mockFaultDeviceList(device0)
		oldFaults := make(map[string][]DeviceFault)
		updateFaultReceiveTimeForDevices(cmFaults, oldFaults)
		faultReceivedTime := cmFaults[device0][0].FaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime
		convey.So(faultReceivedTime, convey.ShouldBeGreaterThan, 0)
	})
	convey.Convey("Test updateFaultReceiveTimeForDevices when oldFaults exist fault devices", t, func() {
		cmFaults := mockFaultDeviceList(device0)
		cmFaults[device1] = []DeviceFault{{FaultTimeAndLevelMap: mockFaultTimeAndLevelMap(AicFaultCode)}}
		oldFaults := mockFaultDeviceList(device0)
		updateFaultReceiveTimeForDevices(cmFaults, oldFaults)
		faultReceivedTime1 := cmFaults[device0][0].FaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime
		faultReceivedTime2 := cmFaults[device1][0].FaultTimeAndLevelMap[AicFaultCode].FaultReceivedTime
		oldFaultsFaultReceivedTime := oldFaults[device0][0].FaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime
		convey.So(faultReceivedTime1, convey.ShouldEqual, oldFaultsFaultReceivedTime)
		convey.So(faultReceivedTime2, convey.ShouldBeGreaterThan, 0)
	})
}

func mockSwitchInfo() *SwitchInfo {
	return &SwitchInfo{SwitchFaultInfo: SwitchFaultInfo{FaultTimeAndLevelMap: mockFaultTimeAndLevelMap(UceFaultCode)}}
}

func TestSwitchInfoUpdateFaultReceiveTime(t *testing.T) {
	convey.Convey("Test SwitchInfo UpdateFaultReceiveTime", t, func() {
		convey.Convey("When cm is nil, it should return directly", func() {
			var cm *SwitchInfo = nil
			if cm != nil {
				cm.UpdateFaultReceiveTime(nil)
			}
			convey.So(cm, convey.ShouldBeNil)
		})
		convey.Convey("When oldInfo type conversion failed, "+
			"cm should use current time to update fault receive time", func() {
			cm := mockSwitchInfo()
			oldInfo := new(AdvanceDeviceFaultCm)
			cm.UpdateFaultReceiveTime(oldInfo)
			faultReceivedTime := cm.FaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime
			convey.So(faultReceivedTime, convey.ShouldBeGreaterThan, 0)
		})
		convey.Convey("When oldInfo is nil, cm should use current time to update fault receive time", func() {
			cm := mockSwitchInfo()
			var oldInfo *SwitchInfo = nil
			cm.UpdateFaultReceiveTime(oldInfo)
			faultReceivedTime := cm.FaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime
			convey.So(faultReceivedTime, convey.ShouldBeGreaterThan, 0)
		})
		convey.Convey("When oldInfo exists target fault, "+
			"cm should use the data in oldInfo to update fault receive time", func() {
			cm := mockSwitchInfo()
			cm.FaultTimeAndLevelMap[AicFaultCode] = FaultTimeAndLevel{
				FaultTime:         0,
				FaultReceivedTime: 0,
				FaultLevel:        SeparateNPU,
			}
			oldInfo := mockSwitchInfo()
			cm.UpdateFaultReceiveTime(oldInfo)
			faultReceivedTime1 := cm.FaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime
			faultReceivedTime2 := cm.FaultTimeAndLevelMap[AicFaultCode].FaultReceivedTime
			oldFaultReceivedTime1 := oldInfo.FaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime
			convey.So(faultReceivedTime1, convey.ShouldEqual, oldFaultReceivedTime1)
			convey.So(faultReceivedTime2, convey.ShouldBeGreaterThan, 0)
		})
	})

}

func TestUpdateFaultReceiveTimeForSwitchs(t *testing.T) {
	convey.Convey("Test updateFaultReceiveTimeForSwitchs when oldFaultTimeAndLevelMap is nil", t, func() {
		cmFaultTimeAndLevelMap := mockFaultTimeAndLevelMap(UceFaultCode)
		cmFaultTimeAndLevelMap[AicFaultCode] = FaultTimeAndLevel{FaultReceivedTime: 0}
		var oldFaultTimeAndLevelMap map[string]FaultTimeAndLevel = nil
		updateFaultReceiveTimeForSwitchs(cmFaultTimeAndLevelMap, oldFaultTimeAndLevelMap)
		convey.So(cmFaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime, convey.ShouldBeGreaterThan, 0)
	})
	convey.Convey("Test updateFaultReceiveTimeForSwitchs when oldFaultTimeAndLevelMap exists target key",
		t, func() {
			cmFaultTimeAndLevelMap := mockFaultTimeAndLevelMap(UceFaultCode)
			cmFaultTimeAndLevelMap[AicFaultCode] = FaultTimeAndLevel{FaultReceivedTime: 0}
			oldFaultTimeAndLevelMap := mockFaultTimeAndLevelMap(UceFaultCode)
			updateFaultReceiveTimeForSwitchs(cmFaultTimeAndLevelMap, oldFaultTimeAndLevelMap)
			faultReceivedTime1 := cmFaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime
			faultReceivedTime2 := cmFaultTimeAndLevelMap[AicFaultCode].FaultReceivedTime
			oldFaultReceivedTime1 := cmFaultTimeAndLevelMap[UceFaultCode].FaultReceivedTime
			convey.So(faultReceivedTime1, convey.ShouldEqual, oldFaultReceivedTime1)
			convey.So(faultReceivedTime2, convey.ShouldBeGreaterThan, 0)
		})
}

func TestDeviceInfoBusinessDataIsNotEqual(t *testing.T) {
	convey.Convey("Test DeviceInfoBusinessDataIsNotEqual", t, func() {
		dev1 := &DeviceInfo{DeviceInfoNoName: DeviceInfoNoName{DeviceList: map[string]string{"npu0": "healthy"}}}
		dev2 := &DeviceInfo{DeviceInfoNoName: DeviceInfoNoName{DeviceList: map[string]string{"npu0": "healthy"}}}
		dev3 := &DeviceInfo{DeviceInfoNoName: DeviceInfoNoName{DeviceList: map[string]string{"npu0": "faulty"}}}
		dev4 := &DeviceInfo{DeviceInfoNoName: DeviceInfoNoName{DeviceList: map[string]string{"npu1": "healthy"}}}
		convey.So(DeviceInfoBusinessDataIsNotEqual(nil, nil), convey.ShouldBeFalse)
		convey.So(DeviceInfoBusinessDataIsNotEqual(nil, dev1), convey.ShouldBeTrue)
		convey.So(DeviceInfoBusinessDataIsNotEqual(dev1, dev2), convey.ShouldBeFalse)
		convey.So(DeviceInfoBusinessDataIsNotEqual(dev1, dev3), convey.ShouldBeTrue)
		convey.So(DeviceInfoBusinessDataIsNotEqual(dev1, dev4), convey.ShouldBeTrue)
	})
}

func TestSwitchInfoBusinessDataIsNotEqual(t *testing.T) {
	convey.Convey("Test SwitchInfoBusinessDataIsNotEqual", t, func() {
		sw1 := &SwitchInfo{SwitchFaultInfo: SwitchFaultInfo{FaultLevel: "1", NodeStatus: "OK"}}
		sw2 := &SwitchInfo{SwitchFaultInfo: SwitchFaultInfo{FaultLevel: "1", NodeStatus: "OK"}}
		sw3 := &SwitchInfo{SwitchFaultInfo: SwitchFaultInfo{FaultLevel: "2", NodeStatus: "OK"}}
		sw4 := &SwitchInfo{SwitchFaultInfo: SwitchFaultInfo{FaultLevel: "1", NodeStatus: "Error"}}
		convey.So(SwitchInfoBusinessDataIsNotEqual(nil, nil), convey.ShouldBeFalse)
		convey.So(SwitchInfoBusinessDataIsNotEqual(nil, sw1), convey.ShouldBeTrue)
		convey.So(SwitchInfoBusinessDataIsNotEqual(sw1, sw2), convey.ShouldBeFalse)
		convey.So(SwitchInfoBusinessDataIsNotEqual(sw1, sw3), convey.ShouldBeTrue)
		convey.So(SwitchInfoBusinessDataIsNotEqual(sw1, sw4), convey.ShouldBeTrue)
	})
}

func TestNodeInfoBusinessDataIsNotEqual(t *testing.T) {
	convey.Convey("Test NodeInfoBusinessDataIsNotEqual", t, func() {
		node1 := &NodeInfo{NodeInfoNoName: NodeInfoNoName{NodeStatus: "Ready",
			FaultDevList: []*FaultDev{{DeviceId: 0}}}}
		node2 := &NodeInfo{NodeInfoNoName: NodeInfoNoName{NodeStatus: "Ready",
			FaultDevList: []*FaultDev{{DeviceId: 0}}}}
		node3 := &NodeInfo{NodeInfoNoName: NodeInfoNoName{NodeStatus: "NotReady",
			FaultDevList: []*FaultDev{{DeviceId: 0}}}}
		node4 := &NodeInfo{NodeInfoNoName: NodeInfoNoName{NodeStatus: "Ready",
			FaultDevList: []*FaultDev{{DeviceId: 1}, {DeviceId: 2}}}}
		convey.So(NodeInfoBusinessDataIsNotEqual(nil, nil), convey.ShouldBeFalse)
		convey.So(NodeInfoBusinessDataIsNotEqual(nil, node1), convey.ShouldBeTrue)
		convey.So(NodeInfoBusinessDataIsNotEqual(node1, node2), convey.ShouldBeFalse)
		convey.So(NodeInfoBusinessDataIsNotEqual(node1, node3), convey.ShouldBeTrue)
		convey.So(NodeInfoBusinessDataIsNotEqual(node1, node4), convey.ShouldBeTrue)
	})
}

func TestSwitchAddFaultAndFix1(t *testing.T) {
	convey.Convey("add switch fault ok", t, func() {
		cm := &SwitchInfo{}
		fault := SimpleSwitchFaultInfo{}
		cm.AddFaultAndFix(fault)
		convey.So(len(cm.FaultInfo), convey.ShouldEqual, 1)
	})
}
