// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"reflect"
	"testing"
	"time"

	"clusterd/pkg/common/util"
)

// ======= Test uceAccompanyFaultProcessor

func TestUceAccompanyFaultProcessor_process(t *testing.T) {
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceAccompanyFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("Test_uceAccompanyFaultProcessor_process", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, testFileErr := readObjectFromUceAccompanyProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		if err != nil {
			t.Errorf("%v", err)
		}
		processor.deviceCmForNodeMap = getAdvanceDeviceCmForNodeMap(cmDeviceInfos)
		processor.uceAccompanyFaultInQue()
		currentTime := 95 * time.Second.Milliseconds()
		processor.filterFaultInfos(currentTime)
		advanceDeviceCmForNodeMapToString(processor.deviceCmForNodeMap, cmDeviceInfos)
		if !reflect.DeepEqual(getAdvanceDeviceCmForNodeMap(cmDeviceInfos),
			getAdvanceDeviceCmForNodeMap(expectProcessedDeviceInfos)) {
			t.Errorf("result = %v, want %v",
				util.ObjToString(cmDeviceInfos), util.ObjToString(expectProcessedDeviceInfos))
		}

		if len(processor.uceAccompanyFaultQue["node1"]["Ascend910-1"]) != 1 &&
			processor.uceAccompanyFaultQue["node1"]["Ascend910-1"][0].FaultCode == "80C98009" {
			t.Errorf("processor.uceAccompanyFaultQue() is wrong")
		}
	})
}
