// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"

	"clusterd/pkg/application/faultmanager/collector"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
)

// ======= Test uceAccompanyFaultProcessor
func TestUceAccompanyFaultProcessorProcess(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceAccompanyFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("TestUceAccompanyFaultProcessorProcess", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, testFileErr := readObjectFromUceAccompanyProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		if err != nil {
			t.Errorf("%v", err)
		}
		processor.deviceCmForNodeMap = faultdomain.GetAdvanceDeviceCmForNodeMap(cmDeviceInfos)
		processor.uceAccompanyFaultInQue()
		currentTime := 95 * time.Second.Milliseconds()
		processor.filterFaultInfos(currentTime)
		faultdomain.AdvanceDeviceCmForNodeMapToString(processor.deviceCmForNodeMap, cmDeviceInfos)
		if !reflect.DeepEqual(faultdomain.GetAdvanceDeviceCmForNodeMap(cmDeviceInfos),
			faultdomain.GetAdvanceDeviceCmForNodeMap(expectProcessedDeviceInfos)) {
			t.Errorf("result = %v, want %v",
				util.ObjToString(cmDeviceInfos), util.ObjToString(expectProcessedDeviceInfos))
		}

		if len(processor.uceAccompanyFaultQue["node1"]["Ascend910-1"]) != 1 &&
			processor.uceAccompanyFaultQue["node1"]["Ascend910-1"][0].FaultCode == "80C98009" {
			t.Error("processor.uceAccompanyFaultQue() is wrong")
		}
	})
}

// TestUceAccompanyFaultProcessorProcess when aic/aiv in queue is exceed DiagnosisTime, then should add in fault map
func TestUceAccompanyFaultProcessorProcessForAddFault(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceAccompanyFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("TestUceAccompanyFaultProcessorProcessForAddFault", func(t *testing.T) {
		currentTime := 95 * time.Second.Milliseconds()
		nodeName := "node1"
		deviceName := "Ascend910A-0"
		processor.uceAccompanyFaultQue = map[string]map[string][]constant.DeviceFault{
			nodeName: {
				deviceName: []constant.DeviceFault{
					{
						NPUName:   deviceName,
						FaultCode: constant.AicFaultCode,
						FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
							constant.AicFaultCode: {
								FaultTime:  89 * time.Second.Milliseconds(),
								FaultLevel: constant.RestartRequest,
							},
						},
					},
				},
			},
		}
		processor.deviceCmForNodeMap = make(map[string]constant.AdvanceDeviceFaultCm)
		processor.filterFaultInfos(currentTime)
		if len(processor.deviceCmForNodeMap[nodeName].FaultDeviceList[deviceName]) != 1 {
			t.Error("TestUceAccompanyFaultProcessorProcessForAddFault fail")
			return
		}
		fault := processor.deviceCmForNodeMap[nodeName].FaultDeviceList[deviceName][0]
		if fault.FaultCode != constant.AicFaultCode {
			t.Error("TestUceAccompanyFaultProcessorProcessForAddFault fail")
			return
		}
	})
}

func TestUceAccompanyFaultProcessorIsBusinessUceFault(t *testing.T) {
	t.Run("TestUceAccompanyFaultProcessorIsBusinessUceFault", func(t *testing.T) {
		deviceProcessCenter := NewDeviceFaultProcessCenter()
		uceAcompanyProcessor, _ := deviceProcessCenter.getUceAccompanyFaultProcessor()
		reportTime := int64(1000)
		patches := gomonkey.ApplyPrivateMethod(collector.ReportInfoCollector, "GetInfoWithoutJobId",
			func(nodeName, deviceName string) constant.ReportInfo {
				return constant.ReportInfo{
					RecoverTime: reportTime,
				}
			})

		defer patches.Reset()
		flag, info := uceAcompanyProcessor.isBusinessUceFault("nodeName", "deviceName")

		if !flag || info.RecoverTime != reportTime {
			t.Error("TestUceAccompanyFaultProcessorIsBusinessUceFault failed.")
		}
	})
}
