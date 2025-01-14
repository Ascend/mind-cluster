// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"reflect"
	"sort"
	"testing"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func TestSplitDeviceFault(t *testing.T) {
	t.Run("TestSplitDeviceFault", func(t *testing.T) {
		npuName := "Ascend910-0"
		var faultInfo = constant.DeviceFault{
			NPUName:   npuName,
			FaultCode: "0x1,0x2",
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
				"0x1": {FaultLevel: NotHandleFault, FaultTime: 1},
				"0x2": {FaultLevel: SubHealthFault, FaultTime: 1},
			},
		}

		got := splitDeviceFault(faultInfo, "node1")
		want := []constant.DeviceFault{
			{
				NPUName:              npuName,
				FaultCode:            "0x1",
				FaultLevel:           NotHandleFault,
				LargeModelFaultLevel: NotHandleFault,
				FaultHandling:        NotHandleFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: NotHandleFault, FaultTime: 1},
				},
			}, {
				NPUName:              npuName,
				FaultCode:            "0x2",
				FaultLevel:           SubHealthFault,
				LargeModelFaultLevel: SubHealthFault,
				FaultHandling:        SubHealthFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x2": {FaultLevel: SubHealthFault, FaultTime: 1},
				},
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("splitDeviceFault() = %v, want %v", got, want)
		}
	})
}

// TestSplitDeviceFaultWithManuallySeparateFaultLevel should split out a DeviceFault as middle data,
// when dp report ManuallySeparateNPU
func TestSplitDeviceFaultWithManuallySeparateFaultLevel(t *testing.T) {
	t.Run("TestSplitDeviceFaultWithManuallySeparateFaultLevel", func(t *testing.T) {
		npuName := "Ascend910-0"
		var faultInfo = constant.DeviceFault{
			NPUName:              npuName,
			FaultCode:            "",
			FaultLevel:           ManuallySeparateNPU,
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{},
		}

		got := splitDeviceFault(faultInfo, "node1")
		want := []constant.DeviceFault{
			{
				NPUName:              npuName,
				FaultCode:            ManuallySeparateNPU,
				FaultLevel:           ManuallySeparateNPU,
				LargeModelFaultLevel: ManuallySeparateNPU,
				FaultHandling:        ManuallySeparateNPU,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					ManuallySeparateNPU: {
						FaultTime:  constant.UnknownFaultTime,
						FaultLevel: ManuallySeparateNPU,
					},
				},
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("splitDeviceFault() = %v, want %v", got, want)
		}
	})
}

// TestMergeSameTypeDeviceFault should be merged, when fault type is same
func TestMergeSameTypeDeviceFault(t *testing.T) {
	t.Run("Test_mergeDeviceFault", func(t *testing.T) {
		npuName := "Ascend910-0"
		split := []constant.DeviceFault{
			{
				FaultType:  CardUnhealthy,
				NPUName:    npuName,
				FaultCode:  "0x1",
				FaultLevel: NotHandleFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: NotHandleFault, FaultTime: 1},
					"0x2": {FaultLevel: SubHealthFault, FaultTime: 1},
				},
			},
			{
				FaultType:  CardUnhealthy,
				NPUName:    npuName,
				FaultCode:  "0x2",
				FaultLevel: SubHealthFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: NotHandleFault, FaultTime: 1},
					"0x2": {FaultLevel: SubHealthFault, FaultTime: 1},
				},
			},
		}
		want := []constant.DeviceFault{
			{
				FaultType:            CardUnhealthy,
				NPUName:              npuName,
				FaultCode:            "0x1,0x2",
				FaultLevel:           SubHealthFault,
				LargeModelFaultLevel: SubHealthFault,
				FaultHandling:        SubHealthFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: NotHandleFault, FaultTime: 1},
					"0x2": {FaultLevel: SubHealthFault, FaultTime: 1},
				},
			},
		}
		got, err := mergeDeviceFault(split)
		if err != nil {
			t.Errorf("mergeDeviceFault() error = %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("mergeDeviceFault() got = %v, want %v", util.ObjToString(got), util.ObjToString(want))
		}
	})
}

// TestMergeDifferentTypeDeviceFault should not be merged, when fault type isn't same
func TestMergeDifferentTypeDeviceFault(t *testing.T) {
	t.Run("Test_mergeDeviceFault", func(t *testing.T) {
		npuName := "Ascend910-0"
		split := []constant.DeviceFault{
			{
				FaultType:            CardUnhealthy,
				NPUName:              npuName,
				FaultCode:            "0x1",
				FaultLevel:           NotHandleFault,
				LargeModelFaultLevel: NotHandleFault,
				FaultHandling:        NotHandleFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: NotHandleFault, FaultTime: 1},
				},
			},
			{
				FaultType:            CardNetworkUnhealthy,
				NPUName:              npuName,
				FaultCode:            "0x2",
				FaultLevel:           SubHealthFault,
				LargeModelFaultLevel: SubHealthFault,
				FaultHandling:        SubHealthFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x2": {FaultLevel: SubHealthFault, FaultTime: 1},
				},
			},
		}
		got, err := mergeDeviceFault(split)
		if err != nil {
			t.Errorf("mergeDeviceFault() error = %v", err)
		}
		sort.Slice(got, func(i, j int) bool {
			return got[i].FaultType > got[j].FaultType
		})
		if !reflect.DeepEqual(got, split) {
			t.Errorf("mergeDeviceFault() got = %v, want %v", util.ObjToString(got), util.ObjToString(split))
		}
	})
}

// TestMergeManuallySeparateNpuTypeDeviceFault should combine other fault info and ManuallySeparateNPU.
func TestMergeManuallySeparateNpuTypeDeviceFault(t *testing.T) {
	t.Run("TestMergeManuallySeparateNpuTypeDeviceFault", func(t *testing.T) {
		npuName := "Ascend910-0"
		split := []constant.DeviceFault{
			{
				FaultType:            CardUnhealthy,
				NPUName:              npuName,
				FaultCode:            ManuallySeparateNPU,
				FaultLevel:           ManuallySeparateNPU,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{},
			},
			{
				FaultType:  CardUnhealthy,
				NPUName:    npuName,
				FaultCode:  constant.UceFaultCode,
				FaultLevel: RestartBusiness,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					constant.UceFaultCode: {
						FaultTime:  constant.UnknownFaultTime,
						FaultLevel: RestartBusiness,
					},
				},
			},
		}
		got, err := mergeDeviceFault(split)
		if err != nil {
			t.Errorf("mergeDeviceFault() error = %v", err)
		}
		want := []constant.DeviceFault{
			{
				FaultType:            CardUnhealthy,
				NPUName:              npuName,
				FaultCode:            constant.UceFaultCode,
				FaultLevel:           ManuallySeparateNPU,
				LargeModelFaultLevel: ManuallySeparateNPU,
				FaultHandling:        ManuallySeparateNPU,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					constant.UceFaultCode: {
						FaultTime:  constant.UnknownFaultTime,
						FaultLevel: RestartBusiness,
					},
				},
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("TestMergeManuallySeparateNpuTypeDeviceFault() got = %v, want %v", util.ObjToString(got), util.ObjToString(want))
		}
	})
}

func TestGetAdvanceDeviceCm(t *testing.T) {
	info := &constant.DeviceInfo{
		DeviceInfoNoName: constant.DeviceInfoNoName{
			DeviceList: map[string]string{
				"huawei.com/Ascend910-Fault": `
[
  {
	"fault_code": "1801   ,  1809  ",
    "fault_time_and_level_map":
      {
        "1801": {"fault_time":1234, "fault_level": "RestartBusiness"}, 
        "1809": {"fault_time":5678, "fault_level": "NotHandleFault"}
      },"npu_name": "xxx"
  }
]`,
			},
		},
	}
	advanceDeviceCm := getAdvanceDeviceCm(info)
	if len(advanceDeviceCm.FaultDeviceList["xxx"]) != 2 {
		t.Errorf("TestGetAdvanceDeviceCm failed")
		return
	}
	faultTimeAndLevel, ok := advanceDeviceCm.FaultDeviceList["xxx"][0].FaultTimeAndLevelMap["1801"]
	if !ok || faultTimeAndLevel.FaultTime != 1234 || faultTimeAndLevel.FaultLevel != "RestartBusiness" {
		t.Errorf("TestGetAdvanceDeviceCm failed")
		return
	}
}
