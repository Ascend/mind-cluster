/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
 */

// Package faultshoot contain fault proces
package faultshoot

import (
	"reflect"
	"testing"

	"clusterd/pkg/common/constant"
)

func TestSplitDeviceFault(t *testing.T) {
	t.Run("Test_splitDeviceFault", func(t *testing.T) {
		var faultInfo = constant.DeviceFault{
			NPUName:   "Ascend910-0",
			FaultCode: "0x1,0x2",
			FaultTimeMap: map[string]int64{
				"0x1": 1,
				"0x2": 2,
			},
		}

		got := splitDeviceFault(faultInfo)
		want := []constant.DeviceFault{
			{
				NPUName:   "Ascend910-0",
				FaultCode: "0x1",
				FaultTimeMap: map[string]int64{
					"0x1": 1,
					"0x2": 2,
				},
			}, {
				NPUName:   "Ascend910-0",
				FaultCode: "0x2",
				FaultTimeMap: map[string]int64{
					"0x1": 1,
					"0x2": 2,
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
				NPUName:   "Ascend910-0",
				FaultCode: "0x1",
				FaultTimeMap: map[string]int64{
					"0x1": 1,
					"0x2": 2,
				},
			},
			{
				NPUName:   "Ascend910-0",
				FaultCode: "0x2",
				FaultTimeMap: map[string]int64{
					"0x1": 1,
					"0x2": 2,
				},
			},
		}
		want := constant.DeviceFault{
			NPUName:   "Ascend910-0",
			FaultCode: "0x1,0x2",
			FaultTimeMap: map[string]int64{
				"0x1": 1,
				"0x2": 2,
			},
		}
		got, err := mergeDeviceFault(split)
		if err != nil {
			t.Errorf("mergeDeviceFault() error = %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("mergeDeviceFault() got = %v, want %v", got, want)
		}
	})
}

func TestGetAdvanceDeviceCm(t *testing.T) {
	info := &constant.DeviceInfo{
		DeviceInfoNoName: constant.DeviceInfoNoName{
			DeviceList: map[string]string{
				"huawei.com/Ascend910-Fault": `[{"fault_time_map":{"1801": 1234, "1809":5678},"npu_name": "xxx"}]`,
			},
			UpdateTime: 0,
		},
		CmName:      "",
		SuperPodID:  0,
		ServerIndex: 0,
	}
	advanceDeviceCm := getAdvanceDeviceCm(info)
	tim, ok := advanceDeviceCm.DeviceList["xxx"][0].FaultTimeMap["1801"]
	if !ok {
		t.Errorf("TestGetAdvanceDeviceCm failed")
		return
	}
	if tim != 1234 {
		t.Errorf("TestGetAdvanceDeviceCm failed")
	}
}
