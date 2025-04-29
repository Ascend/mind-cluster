// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package uce contain uce process
package uce

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/yaml"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/faultdomain/collector"
	"clusterd/pkg/domain/job"
)

const UceFaultTime = int64(100 * time.Second)

func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

func TestMain(m *testing.M) {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	m.Run()
}

// =============Test canFilterUceDeviceFaultInfo===========
// Test current time not exceed (UceFaultTime + JobReportRecoverTimeout), should filter uce fault
func TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation1(t *testing.T) {
	want := true
	for i := 0; i < 10; i++ {
		t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation1-"+string(rune(i)), func(t *testing.T) {
			var currentTime int64
			if i == 0 {
				currentTime = UceFaultTime
			} else if i == 1 {
				currentTime = UceFaultTime + constant.JobReportRecoverTimeout
			} else {
				currentTime = rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1)
			}
			uceDevice := constant.UceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  constant.JobNotRecover,
				CompleteTime: constant.JobNotRecoverComplete,
			}
			if got := UceProcessor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v. uceDevice = %v, currentTime = %v",
					got, want, uceDevice, currentTime)
			}
		})
	}
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// but don't receive job recover info, should not filter uce fault
func TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation2(t *testing.T) {
	want := false
	// don't receive job recover info
	t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation2-1", func(t *testing.T) {
		currentTime := UceFaultTime + constant.JobReportRecoverTimeout + 1 // exceed 1ns
		uceDevice := constant.UceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  constant.JobNotRecover,
			CompleteTime: constant.JobNotRecoverComplete,
		}
		if got := UceProcessor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})

	// receive recover info, but exceed (UceFaultTime + JobReportRecoverTimeout)
	t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation2-2", func(t *testing.T) {
		currentTime := UceFaultTime + constant.JobReportRecoverTimeout + 1
		uceDevice := constant.UceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  UceFaultTime + constant.JobReportRecoverTimeout + 1, // exceed 1ns
			CompleteTime: constant.JobNotRecoverComplete,
		}
		if got := UceProcessor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// but not exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, should filter uce fault
func TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation3(t *testing.T) {
	want := true
	for i := 0; i < 10; i++ {
		t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation3-"+string(rune(i)), func(t *testing.T) {
			uceDevice := constant.UceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
				CompleteTime: constant.JobNotRecoverComplete,
			}
			var currentTime int64
			if i == 0 {
				currentTime = uceDevice.RecoverTime
			} else if i == 1 {
				currentTime = uceDevice.RecoverTime + constant.JobReportCompleteTimeout
			} else {
				currentTime = rand.Int63nRange(uceDevice.RecoverTime, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			}
			if got := UceProcessor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, but don't receive job recover complete info, should not filter uce fault
func TestUceFaultUceProcessorCanFilterUceDeviceFaultInfoSituation4(t *testing.T) {
	want := false
	// don't receive recover complete info
	for i := 0; i < 10; i++ {
		t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation4-1-"+string(rune(i)), func(t *testing.T) {
			uceDevice := constant.UceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
				CompleteTime: constant.JobNotRecoverComplete,
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := UceProcessor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}

	// receive recover complete info, but exceed (RecoverTime + JobReportCompleteTimeout)
	for i := 0; i < 10; i++ {
		t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation4-2-"+string(rune(i)), func(t *testing.T) {
			uceDevice := constant.UceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
				CompleteTime: constant.JobReportRecoverTimeout + 1, // exceed 1s
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := UceProcessor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, and receive job recover complete info, should filter uce fault
func TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation5(t *testing.T) {
	want := true
	for i := 0; i < 10; i++ {
		t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation5-"+string(rune(i)), func(t *testing.T) {
			uceDevice := constant.UceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
				CompleteTime: constant.JobNotRecoverComplete,
			}
			if i == 0 {
				uceDevice.CompleteTime = uceDevice.RecoverTime
			} else if i == 1 {
				uceDevice.CompleteTime = uceDevice.RecoverTime + constant.JobReportCompleteTimeout
			} else {
				uceDevice.CompleteTime = rand.Int63nRange(uceDevice.RecoverTime, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := UceProcessor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, but receive invalid job recover complete info, should not filter uce fault
func TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation6(t *testing.T) {
	want := false
	// CompleteTime smaller than FaultTime
	t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation6-1", func(t *testing.T) {
		uceDevice := constant.UceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
			CompleteTime: UceFaultTime - 1,
		}

		currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
		if got := UceProcessor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})

	// CompleteTime smaller than RecoverTime
	t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation6-2", func(t *testing.T) {
		uceDevice := constant.UceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
			CompleteTime: constant.JobNotRecoverComplete,
		}
		uceDevice.CompleteTime = uceDevice.RecoverTime - 1

		currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
		if got := UceProcessor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, and receive job recover complete info, and complete time valid, should filter
func TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation7(t *testing.T) {
	want := true
	for i := 0; i < 10; i++ {
		t.Run("TestUceFaultUceProcessorCanFilterUceDeviceFaultInfoSituation7-"+string(rune(i)), func(t *testing.T) {
			uceDevice := constant.UceDeviceInfo{
				DeviceName:  "test-device",
				FaultTime:   UceFaultTime,
				RecoverTime: UceFaultTime - 1,
			}
			if i == 0 {
				uceDevice.CompleteTime = uceDevice.FaultTime
			} else if i == 1 {
				uceDevice.CompleteTime = uceDevice.RecoverTime + constant.JobReportCompleteTimeout
			} else {
				uceDevice.CompleteTime = rand.Int63nRange(uceDevice.FaultTime, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := UceProcessor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, and receive job recover complete info, and complete time valid,
// but complete time - recover time exceeds  JobReportCompleteTimeout, should not filter
func TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation8(t *testing.T) {
	want := false
	for i := 0; i < 10; i++ {
		t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation8-"+string(rune(i)), func(t *testing.T) {
			uceDevice := constant.UceDeviceInfo{
				DeviceName:  "test-device",
				FaultTime:   UceFaultTime,
				RecoverTime: UceFaultTime - int64(31*time.Second),
			}
			if i == 0 {
				uceDevice.CompleteTime = uceDevice.FaultTime
			} else if i == 1 {
				uceDevice.CompleteTime = uceDevice.RecoverTime + constant.JobReportCompleteTimeout
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := UceProcessor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// =============Test scenario===========
func TestUceFaultProcessorGetUceDeviceOfNodes(t *testing.T) {
	t.Run("TestUceFaultProcessorGetUceDeviceOfNodes", func(t *testing.T) {
		cmDeviceInfos, _, uceNodesInfos, _, _, testFileErr := readObjectFromUceProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}

		UceProcessor.nodeDeviceCmMap = faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos)
		deviceOfNodes := UceProcessor.getUceDeviceOfNodes()
		if !reflect.DeepEqual(deviceOfNodes, uceNodesInfos) {
			t.Errorf("getUceDeviceOfNodes() = %v, want %v",
				util.ObjToString(deviceOfNodes), util.ObjToString(uceNodesInfos))
		}
	})
}

func TestUceFaultProcessorGetUceDevicesForUceTolerateJobs(t *testing.T) {
	t.Run("TestUceFaultProcessorGetUceDevicesForUceTolerateJobs", func(t *testing.T) {
		cmDeviceInfos, _, _, jobServerInfoMap, expectUceJobsInfo, testFileErr := readObjectFromUceProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}

		UceProcessor.jobServerInfoMap = jobServerInfoMap
		UceProcessor.nodeDeviceCmMap = faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos)
		UceProcessor.uceDeviceOfNode = UceProcessor.getUceDeviceOfNodes()
		UceProcessor.uceDevicesOfUceJob = UceProcessor.getUceDevicesForUceTolerateJobs()
		if !reflect.DeepEqual(UceProcessor.uceDevicesOfUceJob, expectUceJobsInfo) {
			t.Errorf("getUceDevicesForUceTolerateJobs() = %v, want %v",
				util.ObjToString(UceProcessor.uceDevicesOfUceJob), util.ObjToString(expectUceJobsInfo))
		}
	})
}

func TestUceFaultProcessorProcessUceFaultInfo(t *testing.T) {
	t.Run("TestUceFaultProcessorProcessUceFaultInfo", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, _, jobServerInfoMap, _, testFileErr := readObjectFromUceProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}

		UceProcessor.jobServerInfoMap = jobServerInfoMap
		UceProcessor.nodeDeviceCmMap = faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos)
		UceProcessor.uceDeviceOfNode = UceProcessor.getUceDeviceOfNodes()
		UceProcessor.uceDevicesOfUceJob = UceProcessor.getUceDevicesForUceTolerateJobs()
		currentTime := 109 * time.Second.Milliseconds()
		UceProcessor.processUceFaultInfo(currentTime)
		result := UceProcessor.nodeDeviceCmMap
		want := faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](expectProcessedDeviceInfos)
		if !reflect.DeepEqual(result, want) {
			t.Errorf("result:\n%v\n\nwant:\n%v",
				util.ObjToString(result), util.ObjToString(want))
		}
	})
}

func TestUceFaultProcessorScenario1(t *testing.T) {
	t.Run("TestUceFaultProcessorScenario1", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, jobServerInfoMap, reportInfos, testFileErr :=
			readObjectFromUceScenarioTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		collector.ReportInfoCollector = reportInfos

		UceProcessor.jobServerInfoMap = jobServerInfoMap
		UceProcessor.nodeDeviceCmMap = faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos)
		UceProcessor.uceDeviceOfNode = UceProcessor.getUceDeviceOfNodes()
		UceProcessor.uceDevicesOfUceJob = UceProcessor.getUceDevicesForUceTolerateJobs()
		currentTime := 100 * time.Second.Milliseconds()
		UceProcessor.processUceFaultInfo(currentTime)
		result := UceProcessor.nodeDeviceCmMap
		want := faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](expectProcessedDeviceInfos)
		if !reflect.DeepEqual(result, want) {
			t.Errorf("processUceFaultInfo() = %v, \n\nwant %v",
				util.ObjToString(result), util.ObjToString(want))
		}
	})
}

func TestUceFaultProcessorScenario2(t *testing.T) {
	t.Run("TestUceFaultProcessorScenario2", func(t *testing.T) {
		cmDeviceInfos, expProcessedDeviceInfos, jobServerInfoMap, reportInfos, testFileErr :=
			readObjectFromUceScenarioTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		content := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
			AllConfigmap: faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos),
			UpdateConfigmap: []constant.InformerCmItem[*constant.AdvanceDeviceFaultCm]{
				{
					IsAdd: false,
					Data:  nil},
			},
		}
		collector.ReportInfoCollector = reportInfos
		mockJob := gomonkey.ApplyFunc(job.GetJobServerInfoMap, func() constant.JobServerInfoMap {
			return jobServerInfoMap
		})
		mockTime := time.Time{}
		mockUnixMilli := gomonkey.ApplyPrivateMethod(mockTime, "UnixMilli", func() int64 {
			return 100 * time.Second.Milliseconds()
		})
		mockNow := gomonkey.ApplyFunc(time.Now, func() time.Time {
			return mockTime
		})
		defer func() {
			mockJob.Reset()
			mockNow.Reset()
			mockUnixMilli.Reset()
		}()

		resultContent := UceProcessor.Process(content).(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
		result := resultContent.AllConfigmap
		want := faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](expProcessedDeviceInfos)
		if !reflect.DeepEqual(result, want) {
			t.Errorf("processUceFaultInfo() = %v, want %v",
				util.ObjToString(result), util.ObjToString(want))
		}
	})
}

func readObjectFromUceProcessorTestYaml() (
	map[string]*constant.DeviceInfo, map[string]*constant.DeviceInfo,
	map[string]constant.UceNodeInfo, constant.JobServerInfoMap, map[string]constant.UceJobInfo, error) {

	var testDataPath = "../../../../../testdata/resource/uce_fault_processor_test.yaml"
	var maxFileSize = 10000
	var cmDeviceInfos = make(map[string]*constant.DeviceInfo)
	var expectProcessedDeviceInfos = make(map[string]*constant.DeviceInfo)
	var uceNodesInfos = make(map[string]constant.UceNodeInfo)
	var expectUceJobsInfo = make(map[string]constant.UceJobInfo)
	var err error
	var fileSize int64
	var decoder *yaml.YAMLOrJSONDecoder
	var jobDevices = make(map[string]map[string]constant.ServerHccl)
	var jobIsUce = make(map[string]bool)
	var jobServerInfo constant.JobServerInfoMap
	var open *os.File

	fileInfo, err := os.Stat(testDataPath)
	if err != nil {
		err = fmt.Errorf("testDataPath invalid")
		return cmDeviceInfos, expectProcessedDeviceInfos, uceNodesInfos, jobServerInfo, expectUceJobsInfo, err
	}
	fileSize = fileInfo.Size()
	if fileSize > int64(maxFileSize) {
		err = fmt.Errorf("testData file size too big")
		return cmDeviceInfos, expectProcessedDeviceInfos, uceNodesInfos, jobServerInfo, expectUceJobsInfo, err
	}

	open, err = os.Open(testDataPath)
	if err != nil {
		err = fmt.Errorf("open testData file failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, uceNodesInfos, jobServerInfo, expectUceJobsInfo, err
	}

	decoder = yaml.NewYAMLOrJSONDecoder(open, maxFileSize)

	return extractContentForUceTest(err, decoder, cmDeviceInfos, expectProcessedDeviceInfos, uceNodesInfos,
		jobServerInfo, expectUceJobsInfo, jobDevices, jobIsUce)
}

func extractContentForUceTest(err error, decoder *yaml.YAMLOrJSONDecoder, cmDeviceInfos map[string]*constant.DeviceInfo,
	expectProcessedDeviceInfos map[string]*constant.DeviceInfo, uceNodesInfos map[string]constant.UceNodeInfo,
	jobServerInfo constant.JobServerInfoMap, expectUceJobsInfo map[string]constant.UceJobInfo,
	jobDevices map[string]map[string]constant.ServerHccl, jobIsUce map[string]bool) (map[string]*constant.DeviceInfo,
	map[string]*constant.DeviceInfo, map[string]constant.UceNodeInfo, constant.JobServerInfoMap,
	map[string]constant.UceJobInfo, error) {
	err = decoder.Decode(&cmDeviceInfos)
	if err != nil {
		err = fmt.Errorf("cmDeviceInfos decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, uceNodesInfos, jobServerInfo, expectUceJobsInfo, err
	}

	err = decoder.Decode(&expectProcessedDeviceInfos)
	if err != nil {
		err = fmt.Errorf("expectProcessedDeviceInfos decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, uceNodesInfos, jobServerInfo, expectUceJobsInfo, err
	}

	err = decoder.Decode(&uceNodesInfos)
	if err != nil {
		err = fmt.Errorf("uceNodesInfos decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, uceNodesInfos, jobServerInfo, expectUceJobsInfo, err
	}

	err = decoder.Decode(&jobDevices)
	if err != nil {
		err = fmt.Errorf("jobs decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, uceNodesInfos, jobServerInfo, expectUceJobsInfo, err
	}
	jobServerInfo.InfoMap = jobDevices

	err = decoder.Decode(&jobIsUce)
	if err != nil {
		err = fmt.Errorf("jobIsUce decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, uceNodesInfos, jobServerInfo, expectUceJobsInfo, err
	}
	jobServerInfo.UceTolerate = jobIsUce

	err = decoder.Decode(&expectUceJobsInfo)
	if err != nil {
		err = fmt.Errorf("expectUceJobsInfo decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, uceNodesInfos, jobServerInfo, expectUceJobsInfo, err
	}
	return cmDeviceInfos, expectProcessedDeviceInfos, uceNodesInfos, jobServerInfo, expectUceJobsInfo, err
}

func readObjectFromUceScenarioTestYaml() (
	map[string]*constant.DeviceInfo, map[string]*constant.DeviceInfo,
	constant.JobServerInfoMap, *collector.JobReportInfoCollector, error) {

	var testDataPath = "../../../../../testdata/resource/uce_scenario_test.yaml"
	var cmDeviceInfos = make(map[string]*constant.DeviceInfo)
	var expectDeviceInfos = make(map[string]*constant.DeviceInfo)
	var err error
	var fileSize int64
	var decoder *yaml.YAMLOrJSONDecoder
	var jobServerInfo constant.JobServerInfoMap
	var open *os.File
	var reportInfos collector.JobReportInfoCollector
	maxFileSize := 10000

	fileInfo, err := os.Stat(testDataPath)
	if err != nil {
		err = fmt.Errorf("testDataPath invalid")
		return cmDeviceInfos, expectDeviceInfos, jobServerInfo, &reportInfos, err
	}
	fileSize = fileInfo.Size()
	if fileSize > int64(maxFileSize) {
		err = fmt.Errorf("testData file size too big")
		return cmDeviceInfos, expectDeviceInfos, jobServerInfo, &reportInfos, err
	}
	open, err = os.Open(testDataPath)
	if err != nil {
		err = fmt.Errorf("open testData file failed")
		return cmDeviceInfos, expectDeviceInfos, jobServerInfo, &reportInfos, err
	}
	decoder = yaml.NewYAMLOrJSONDecoder(open, maxFileSize)
	return extractContent(decoder, cmDeviceInfos, expectDeviceInfos, jobServerInfo, reportInfos)
}

func extractContent(decoder *yaml.YAMLOrJSONDecoder, cmDeviceInfos map[string]*constant.DeviceInfo,
	expectProcessedDeviceInfos map[string]*constant.DeviceInfo, jobServerInfo constant.JobServerInfoMap,
	reportInfos collector.JobReportInfoCollector) (map[string]*constant.DeviceInfo, map[string]*constant.DeviceInfo,
	constant.JobServerInfoMap, *collector.JobReportInfoCollector, error) {
	err := decoder.Decode(&cmDeviceInfos)
	if err != nil {
		err = fmt.Errorf("cmDeviceInfos decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, jobServerInfo, &reportInfos, err
	}

	err = decoder.Decode(&expectProcessedDeviceInfos)
	if err != nil {
		err = fmt.Errorf("expectProcessedDeviceInfos decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, jobServerInfo, &reportInfos, err
	}

	var jobDevices = make(map[string]map[string]constant.ServerHccl)
	err = decoder.Decode(&jobDevices)
	if err != nil {
		err = fmt.Errorf("jobs decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, jobServerInfo, &reportInfos, err
	}
	var jobIsUce = make(map[string]bool)
	err = decoder.Decode(&jobIsUce)
	if err != nil {
		err = fmt.Errorf("josIsUce decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, jobServerInfo, &reportInfos, err
	}
	jobServerInfo.InfoMap = jobDevices
	jobServerInfo.UceTolerate = jobIsUce

	err = decoder.Decode(&reportInfos)
	if err != nil {
		err = fmt.Errorf("reportInfos decode failed")
	}
	return cmDeviceInfos, expectProcessedDeviceInfos, jobServerInfo, &reportInfos, err
}
