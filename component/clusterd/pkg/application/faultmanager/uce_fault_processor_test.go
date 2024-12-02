// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"

	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/kube"
)

const UceFaultTime = int64(100 * time.Second)

func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// =============Test canFilterUceDeviceFaultInfo===========
// Test current time not exceed (UceFaultTime + JobReportRecoverTimeout), should filter uce fault
func TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation1(t *testing.T) {
	want := true
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
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
			uceDevice := uceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  constant.JobNotRecover,
				CompleteTime: constant.JobNotRecoverComplete,
			}
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
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
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	// don't receive job recover info
	t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation2-1", func(t *testing.T) {
		currentTime := UceFaultTime + constant.JobReportRecoverTimeout + 1 // exceed 1ns
		uceDevice := uceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  constant.JobNotRecover,
			CompleteTime: constant.JobNotRecoverComplete,
		}
		if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})

	// receive recover info, but exceed (UceFaultTime + JobReportRecoverTimeout)
	t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation2-2", func(t *testing.T) {
		currentTime := UceFaultTime + constant.JobReportRecoverTimeout + 1
		uceDevice := uceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  UceFaultTime + constant.JobReportRecoverTimeout + 1, // exceed 1ns
			CompleteTime: constant.JobNotRecoverComplete,
		}
		if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// but not exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, should filter uce fault
func TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation3(t *testing.T) {
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	want := true
	for i := 0; i < 10; i++ {
		t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation3-"+string(rune(i)), func(t *testing.T) {
			uceDevice := uceDeviceInfo{
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
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, but don't receive job recover complete info, should not filter uce fault
func TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation4(t *testing.T) {
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	want := false
	if err != nil {
		t.Errorf("%v", err)
	}
	// don't receive recover complete info
	for i := 0; i < 10; i++ {
		t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation4-1-"+string(rune(i)), func(t *testing.T) {
			uceDevice := uceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
				CompleteTime: constant.JobNotRecoverComplete,
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}

	// receive recover complete info, but exceed (RecoverTime + JobReportCompleteTimeout)
	for i := 0; i < 10; i++ {
		t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation4-2-"+string(rune(i)), func(t *testing.T) {
			uceDevice := uceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
				CompleteTime: constant.JobReportRecoverTimeout + 1, // exceed 1s
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, and receive job recover complete info, should filter uce fault
func TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation5(t *testing.T) {
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	want := true
	if err != nil {
		t.Errorf("%v", err)
	}
	for i := 0; i < 10; i++ {
		t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation5-"+string(rune(i)), func(t *testing.T) {
			uceDevice := uceDeviceInfo{
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
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, but receive invalid job recover complete info, should not filter uce fault
func TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation6(t *testing.T) {
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	want := false
	if err != nil {
		t.Errorf("%v", err)
	}
	// CompleteTime smaller than FaultTime
	t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation6-1", func(t *testing.T) {
		uceDevice := uceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
			CompleteTime: UceFaultTime - 1,
		}

		currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
		if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})

	// CompleteTime smaller than RecoverTime
	t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation6-2", func(t *testing.T) {
		uceDevice := uceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
			CompleteTime: constant.JobNotRecoverComplete,
		}
		uceDevice.CompleteTime = uceDevice.RecoverTime - 1

		currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
		if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, and receive job recover complete info, and complete time valid, should filter
func TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation7(t *testing.T) {
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	want := true
	if err != nil {
		t.Errorf("%v", err)
	}
	for i := 0; i < 10; i++ {
		t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation7-"+string(rune(i)), func(t *testing.T) {
			uceDevice := uceDeviceInfo{
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
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
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
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	want := false
	if err != nil {
		t.Errorf("%v", err)
	}
	for i := 0; i < 10; i++ {
		t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation8-"+string(rune(i)), func(t *testing.T) {
			uceDevice := uceDeviceInfo{
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
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// ========= Test mindio callback ===========
func TestUceFaultProcessorCallbackForReportUceInfo(t *testing.T) {
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("TestUceFaultProcessorCallbackForReportUceInfo", func(t *testing.T) {
		cmDeviceInfos, _, _, jobsPodWorkers, _, testFileErr := readObjectFromUceProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		if err != nil {
			t.Errorf("%v", err)
		}
		type testsuit struct {
			jobId       string
			rankId      string
			recoverTime int64
		}
		ts1Success := testsuit{
			jobId:       "job1",
			rankId:      "1",
			recoverTime: 10,
		}

		ts2Fault := testsuit{
			jobId:       "job2",
			rankId:      "9",
			recoverTime: 10,
		}
		defer func() {
			processor.reportInfo.InfoMap = make(map[string]map[string]map[string]reportInfo)
		}()
		processor.jobServerInfoMap = kube.JobMgr.GetJobServerInfoMap()
		processor.nodeDeviceCmMap = getAdvanceDeviceCmForNodeMap(cmDeviceInfos)
		err = deviceFaultProcessCenter.callbackForReportUceInfo(ts1Success.jobId, ts1Success.rankId, ts1Success.recoverTime)
		nodeName, deviceId, _ := getNodeAndDeviceFromJobIdAndRankId(ts1Success.jobId, ts1Success.rankId, processor.jobServerInfoMap)
		deviceName := "Ascend910-" + deviceId
		info, ok := processor.reportInfo.InfoMap[ts1Success.jobId][nodeName][deviceName]
		if err != nil || !ok || info.RecoverTime != ts1Success.recoverTime {
			t.Errorf("test CallbackForReportUceInfo success failed.")
		}
		err = deviceFaultProcessCenter.callbackForReportUceInfo(ts2Fault.jobId, ts2Fault.rankId, ts2Fault.recoverTime)
		if err == nil {
			t.Errorf("test CallbackForReportUceInfo fault failed.")
		}
	})
}

// =============Test scenario===========
func TestUceFaultProcessorGetUceDeviceOfNodes(t *testing.T) {
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("TestUceFaultProcessorGetUceDeviceOfNodes", func(t *testing.T) {
		cmDeviceInfos, _, uceNodesInfos, _, _, testFileErr := readObjectFromUceProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}

		processor.nodeDeviceCmMap = getAdvanceDeviceCmForNodeMap(cmDeviceInfos)
		deviceOfNodes := processor.getUceDeviceOfNodes()
		if !reflect.DeepEqual(deviceOfNodes, uceNodesInfos) {
			t.Errorf("getUceDeviceOfNodes() = %v, want %v",
				util.ObjToString(deviceOfNodes), util.ObjToString(uceNodesInfos))
		}
	})
}

func TestUceFaultProcessorGetUceDevicesForUceTolerateJobs(t *testing.T) {
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("TestUceFaultProcessorGetUceDevicesForUceTolerateJobs", func(t *testing.T) {
		cmDeviceInfos, _, _, jobsPodWorkers, expectUceJobsInfo, testFileErr := readObjectFromUceProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}

		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		processor.jobServerInfoMap = kube.JobMgr.GetJobServerInfoMap()
		processor.nodeDeviceCmMap = getAdvanceDeviceCmForNodeMap(cmDeviceInfos)
		processor.uceDeviceOfNode = processor.getUceDeviceOfNodes()
		processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs()
		if !reflect.DeepEqual(processor.uceDevicesOfUceJob, expectUceJobsInfo) {
			t.Errorf("getUceDevicesForUceTolerateJobs() = %v, want %v",
				util.ObjToString(processor.uceDevicesOfUceJob), util.ObjToString(expectUceJobsInfo))
		}
	})
}

func TestUceFaultProcessorProcessUceFaultInfo(t *testing.T) {
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("TestUceFaultProcessorProcessUceFaultInfo", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, _, jobsPodWorkers, _, testFileErr := readObjectFromUceProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}

		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		processor.jobServerInfoMap = kube.JobMgr.GetJobServerInfoMap()
		processor.nodeDeviceCmMap = getAdvanceDeviceCmForNodeMap(cmDeviceInfos)
		processor.uceDeviceOfNode = processor.getUceDeviceOfNodes()
		processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs()
		currentTime := 109 * time.Second.Milliseconds()
		processor.processUceFaultInfo(currentTime)
		advanceDeviceCmForNodeMapToString(processor.nodeDeviceCmMap, cmDeviceInfos)
		result := getAdvanceDeviceCmForNodeMap(cmDeviceInfos)
		want := getAdvanceDeviceCmForNodeMap(expectProcessedDeviceInfos)
		if !reflect.DeepEqual(result, want) {
			t.Errorf("processUceFaultInfo() = %v, want %v",
				util.ObjToString(result), util.ObjToString(want))
		}
	})
}

func TestUceFaultProcessorScenario1(t *testing.T) {
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("TestUceFaultProcessorScenario1", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, jobsPodWorkers, reportInfos, testFileErr :=
			readObjectFromUceScenarioTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}

		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		processor.reportInfo = reportInfos
		processor.jobServerInfoMap = kube.JobMgr.GetJobServerInfoMap()
		processor.nodeDeviceCmMap = getAdvanceDeviceCmForNodeMap(cmDeviceInfos)
		processor.uceDeviceOfNode = processor.getUceDeviceOfNodes()
		processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs()
		currentTime := 100 * time.Second.Milliseconds()
		processor.processUceFaultInfo(currentTime)
		advanceDeviceCmForNodeMapToString(processor.nodeDeviceCmMap, cmDeviceInfos)
		result := getAdvanceDeviceCmForNodeMap(cmDeviceInfos)
		want := getAdvanceDeviceCmForNodeMap(expectProcessedDeviceInfos)
		if !reflect.DeepEqual(result, want) {
			t.Errorf("processUceFaultInfo() = %v, want %v",
				util.ObjToString(result), util.ObjToString(want))
		}
	})
}
