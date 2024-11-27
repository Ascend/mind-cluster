// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"clusterd/pkg/common/util"
	"reflect"
	"testing"

	"clusterd/pkg/common/constant"
)

func TestSplitDeviceFault(t *testing.T) {
	t.Run("Test_splitDeviceFault", func(t *testing.T) {
		var faultInfo = constant.DeviceFault{
			NPUName:   "Ascend910-0",
			FaultCode: "0x1,0x2",
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
				"0x1": {FaultLevel: NotHandleFault, FaultTime: 1},
				"0x2": {FaultLevel: SubHealthFault, FaultTime: 1},
			},
		}

		got := splitDeviceFault(faultInfo)
		want := []constant.DeviceFault{
			{
				NPUName:              "Ascend910-0",
				FaultCode:            "0x1",
				FaultLevel:           NotHandleFault,
				LargeModelFaultLevel: NotHandleFault,
				FaultHandling:        NotHandleFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: NotHandleFault, FaultTime: 1},
					"0x2": {FaultLevel: SubHealthFault, FaultTime: 1},
				},
			}, {
				NPUName:              "Ascend910-0",
				FaultCode:            "0x2",
				FaultLevel:           SubHealthFault,
				LargeModelFaultLevel: SubHealthFault,
				FaultHandling:        SubHealthFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: NotHandleFault, FaultTime: 1},
					"0x2": {FaultLevel: SubHealthFault, FaultTime: 1},
				},
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("splitDeviceFault() = %v, want %v", got, want)
		}
	})
}

func TestMergeDeviceFault(t *testing.T) {
	t.Run("Test_mergeDeviceFault", func(t *testing.T) {
		split := []constant.DeviceFault{
			{
				NPUName:    "Ascend910-0",
				FaultCode:  "0x1",
				FaultLevel: NotHandleFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: NotHandleFault, FaultTime: 1},
					"0x2": {FaultLevel: SubHealthFault, FaultTime: 1},
				},
			},
			{
				NPUName:    "Ascend910-0",
				FaultCode:  "0x2",
				FaultLevel: SubHealthFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: NotHandleFault, FaultTime: 1},
					"0x2": {FaultLevel: SubHealthFault, FaultTime: 1},
				},
			},
		}
		want := constant.DeviceFault{
			NPUName:              "Ascend910-0",
			FaultCode:            "0x1,0x2",
			FaultLevel:           SubHealthFault,
			LargeModelFaultLevel: SubHealthFault,
			FaultHandling:        SubHealthFault,
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
				"0x1": {FaultLevel: NotHandleFault, FaultTime: 1},
				"0x2": {FaultLevel: SubHealthFault, FaultTime: 1},
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

func TestGetAdvanceDeviceCm(t *testing.T) {
	info := &constant.DeviceInfo{
		DeviceInfoNoName: constant.DeviceInfoNoName{
			DeviceList: map[string]string{
				"huawei.com/Ascend910-Fault": `
[
  {
    "fault_time_and_level_map":
      {
        "1801": {"fault_time":1234, "fault_level": "RestartBusiness"}, 
        "1809": {"fault_time":5678, "fault_level": "NotHandleFault"}
      },"npu_name": "xxx"
  }
]`,
			},
			UpdateTime: 0,
		},
		CmName:      "",
		SuperPodID:  0,
		ServerIndex: 0,
	}
	advanceDeviceCm := getAdvanceDeviceCm(info)
	faultTimeAndLevel, ok := advanceDeviceCm.FaultDeviceList["xxx"][0].FaultTimeAndLevelMap["1801"]
	if !ok {
		t.Errorf("TestGetAdvanceDeviceCm failed")
		return
	}
	if faultTimeAndLevel.FaultTime != 1234 {
		t.Errorf("TestGetAdvanceDeviceCm failed")
		return
	}
	if faultTimeAndLevel.FaultLevel != "RestartBusiness" {
		t.Errorf("TestGetAdvanceDeviceCm failed")
	}
}
