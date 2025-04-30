// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package uceaccompany contain aiv/aic fault process
package uceaccompany

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/apimachinery/pkg/util/yaml"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/faultdomain/collector"
)

// CurrentTime current time of test case
var CurrentTime int64

func TestMain(m *testing.M) {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	CurrentTime = 95 * time.Second.Milliseconds()
	m.Run()
}

// ======= Test uceAccompanyFaultProcessor
func TestUceAccompanyFaultProcessorProcess(t *testing.T) {
	t.Run("TestUceAccompanyFaultProcessorProcess", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, testFileErr := readObjectFromUceAccompanyProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		UceAccompanyProcessor.deviceCmForNodeMap = faultdomain.GetAdvanceDeviceCmForNodeMap(cmDeviceInfos)
		UceAccompanyProcessor.uceAccompanyFaultInQue()
		UceAccompanyProcessor.filterFaultInfos(CurrentTime)
		faultdomain.AdvanceDeviceCmForNodeMapToString(UceAccompanyProcessor.deviceCmForNodeMap, cmDeviceInfos)
		if !reflect.DeepEqual(faultdomain.GetAdvanceDeviceCmForNodeMap(cmDeviceInfos),
			faultdomain.GetAdvanceDeviceCmForNodeMap(expectProcessedDeviceInfos)) {
			t.Errorf("result = %v, want %v",
				util.ObjToString(cmDeviceInfos), util.ObjToString(expectProcessedDeviceInfos))
		}

		if len(UceAccompanyProcessor.uceAccompanyFaultQue["node1"]["Ascend910-1"]) != 1 &&
			UceAccompanyProcessor.uceAccompanyFaultQue["node1"]["Ascend910-1"][0].FaultCode == "80C98009" {
			t.Error("uceAccompanyFaultQue() is wrong")
		}
	})
}

// ======= Test TestUceAccompanyFaultProcessorProcessE2E
func TestUceAccompanyFaultProcessorProcessE2E(t *testing.T) {
	t.Run("TestUceAccompanyFaultProcessorProcessE2E", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, testFileErr := readObjectFromUceAccompanyProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		content := constant.OneConfigmapContent[*constant.DeviceInfo]{
			AllConfigmap:    cmDeviceInfos,
			UpdateConfigmap: []constant.InformerCmItem[*constant.DeviceInfo]{{}},
		}
		mockTime := time.Time{}
		mockUnixMilli := gomonkey.ApplyPrivateMethod(mockTime, "UnixMilli", func() int64 {
			return CurrentTime
		})
		mockNow := gomonkey.ApplyFunc(time.Now, func() time.Time {
			return mockTime
		})
		defer func() {
			mockNow.Reset()
			mockUnixMilli.Reset()
		}()
		resultContent := UceAccompanyProcessor.Process(content).(constant.OneConfigmapContent[*constant.DeviceInfo])

		if !reflect.DeepEqual(faultdomain.GetAdvanceDeviceCmForNodeMap(resultContent.AllConfigmap),
			faultdomain.GetAdvanceDeviceCmForNodeMap(expectProcessedDeviceInfos)) {
			t.Errorf("result = %v, want %v",
				util.ObjToString(cmDeviceInfos), util.ObjToString(expectProcessedDeviceInfos))
		}

		if len(UceAccompanyProcessor.uceAccompanyFaultQue["node1"]["Ascend910-1"]) != 1 &&
			UceAccompanyProcessor.uceAccompanyFaultQue["node1"]["Ascend910-1"][0].FaultCode == "80C98009" {
			t.Error("uceAccompanyFaultQue() is wrong")
		}
	})
}

// TestUceAccompanyFaultProcessorProcess when aic/aiv in queue is exceed DiagnosisTime, then should add in fault map
func TestUceAccompanyFaultProcessorProcessForAddFault(t *testing.T) {
	t.Run("TestUceAccompanyFaultProcessorProcessForAddFault", func(t *testing.T) {
		nodeName := "node1"
		deviceName := "Ascend910A-0"
		UceAccompanyProcessor.uceAccompanyFaultQue = map[string]map[string][]constant.DeviceFault{
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
		UceAccompanyProcessor.deviceCmForNodeMap = make(map[string]constant.AdvanceDeviceFaultCm)
		UceAccompanyProcessor.filterFaultInfos(CurrentTime)
		if len(UceAccompanyProcessor.deviceCmForNodeMap[nodeName].FaultDeviceList[deviceName]) != 1 {
			t.Error("TestUceAccompanyFaultProcessorProcessForAddFault fail")
			return
		}
		fault := UceAccompanyProcessor.deviceCmForNodeMap[nodeName].FaultDeviceList[deviceName][0]
		if fault.FaultCode != constant.AicFaultCode {
			t.Error("TestUceAccompanyFaultProcessorProcessForAddFault fail")
			return
		}
	})
}

func TestUceAccompanyFaultProcessorIsBusinessUceFault(t *testing.T) {
	t.Run("TestUceAccompanyFaultProcessorIsBusinessUceFault", func(t *testing.T) {
		reportTime := int64(1000)
		patches := gomonkey.ApplyPrivateMethod(collector.ReportInfoCollector, "GetInfoWithoutJobId",
			func(nodeName, deviceName string) constant.ReportInfo {
				return constant.ReportInfo{
					RecoverTime: reportTime,
				}
			})

		defer patches.Reset()
		flag, info := UceAccompanyProcessor.isBusinessUceFault("nodeName", "deviceName")

		if !flag || info.RecoverTime != reportTime {
			t.Error("TestUceAccompanyFaultProcessorIsBusinessUceFault failed.")
		}
	})
}

func readObjectFromUceAccompanyProcessorTestYaml() (
	map[string]*constant.DeviceInfo, map[string]*constant.DeviceInfo, error) {

	var testDataPath = "../../../../../testdata/resource/uce_accompany_processor_test.yaml"
	var cmDeviceInfos = make(map[string]*constant.DeviceInfo)
	var expectProcessedDeviceInfos = make(map[string]*constant.DeviceInfo)
	var err error
	var fileSize int64
	var decoder *yaml.YAMLOrJSONDecoder
	var open *os.File
	maxFileSize := 10000

	fileInfo, err := os.Stat(testDataPath)
	if err != nil {
		err = fmt.Errorf("testDataPath invalid")
		goto RetureLabel
	}
	fileSize = fileInfo.Size()
	if fileSize > int64(maxFileSize) {
		err = fmt.Errorf("testData file size too big")
		goto RetureLabel
	}

	open, err = os.Open(testDataPath)
	if err != nil {
		err = fmt.Errorf("open testData file failed")
		goto RetureLabel
	}

	decoder = yaml.NewYAMLOrJSONDecoder(open, maxFileSize)

	err = decoder.Decode(&cmDeviceInfos)
	if err != nil {
		err = fmt.Errorf("cmDeviceInfos decode failed")
		goto RetureLabel
	}

	err = decoder.Decode(&expectProcessedDeviceInfos)
	if err != nil {
		err = fmt.Errorf("expectProcessedDeviceInfos decode failed")
		goto RetureLabel
	}

RetureLabel:
	return cmDeviceInfos, expectProcessedDeviceInfos, err
}
