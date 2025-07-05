// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package uce contain uce process
package retry

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

// =============Test canFilterRetryDeviceFaultInfo===========
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
			uceDevice := constant.RetryDeviceInfo{
				DeviceName: "test-device",
				FaultDetail: map[string]constant.DeviceFaultDetail{constant.DeviceRetryFault: {
					FaultTime:    UceFaultTime,
					RecoverTime:  constant.JobNotRecover,
					CompleteTime: constant.JobNotRecoverComplete}},
			}
			if got := RetryProcessor.canFilterRetryDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterRetryDeviceFaultInfo() = %v, want %v. uceDevice = %v, currentTime = %v",
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
		uceDevice := constant.RetryDeviceInfo{
			DeviceName: "test-device",
			FaultDetail: map[string]constant.DeviceFaultDetail{constant.DeviceRetryFault: {
				FaultTime:    UceFaultTime,
				RecoverTime:  constant.JobNotRecover,
				CompleteTime: constant.JobNotRecoverComplete}},
		}
		if got := RetryProcessor.canFilterRetryDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterRetryDeviceFaultInfo() = %v, want %v", got, want)
		}
	})

	// receive recover info, but exceed (UceFaultTime + JobReportRecoverTimeout)
	t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation2-2", func(t *testing.T) {
		currentTime := UceFaultTime + constant.JobReportRecoverTimeout + 1
		uceDevice := constant.RetryDeviceInfo{
			DeviceName: "test-device",
			FaultDetail: map[string]constant.DeviceFaultDetail{constant.DeviceRetryFault: {
				FaultTime:    UceFaultTime,
				RecoverTime:  UceFaultTime + constant.JobReportRecoverTimeout + 1, // exceed 1ns
				CompleteTime: constant.JobNotRecoverComplete}},
		}
		if got := RetryProcessor.canFilterRetryDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterRetryDeviceFaultInfo() = %v, want %v", got, want)
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
			uceDevice := constant.RetryDeviceInfo{
				DeviceName: "test-device",
				FaultDetail: map[string]constant.DeviceFaultDetail{constant.DeviceRetryFault: {
					FaultTime:    UceFaultTime,
					RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
					CompleteTime: constant.JobNotRecoverComplete}},
			}
			var currentTime int64
			if i == 0 {
				currentTime = uceDevice.FaultDetail[constant.DeviceRetryFault].RecoverTime
			} else if i == 1 {
				currentTime = uceDevice.FaultDetail[constant.DeviceRetryFault].RecoverTime + constant.JobReportCompleteTimeout
			} else {
				currentTime = rand.Int63nRange(uceDevice.FaultDetail[constant.DeviceRetryFault].RecoverTime,
					uceDevice.FaultDetail[constant.DeviceRetryFault].RecoverTime+constant.JobReportCompleteTimeout+1)
			}
			if got := RetryProcessor.canFilterRetryDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterRetryDeviceFaultInfo() = %v, want %v", got, want)
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
			uceDevice := constant.RetryDeviceInfo{
				DeviceName: "test-device",
				FaultDetail: map[string]constant.DeviceFaultDetail{constant.DeviceRetryFault: {
					FaultTime:    UceFaultTime,
					RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
					CompleteTime: constant.JobNotRecoverComplete}},
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1,
				uceDevice.FaultDetail[constant.DeviceRetryFault].RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := RetryProcessor.canFilterRetryDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterRetryDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}

	// receive recover complete info, but exceed (RecoverTime + JobReportCompleteTimeout)
	for i := 0; i < 10; i++ {
		t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation4-2-"+string(rune(i)), func(t *testing.T) {
			uceDevice := constant.RetryDeviceInfo{
				DeviceName: "test-device",
				FaultDetail: map[string]constant.DeviceFaultDetail{constant.DeviceRetryFault: {
					FaultTime:    UceFaultTime,
					RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
					CompleteTime: constant.JobReportRecoverTimeout + 1, // exceed 1s
				}},
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1,
				uceDevice.FaultDetail[constant.DeviceRetryFault].RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := RetryProcessor.canFilterRetryDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterRetryDeviceFaultInfo() = %v, want %v", got, want)
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
			uceDevice := constant.RetryDeviceInfo{
				DeviceName: "test-device",
				FaultDetail: map[string]constant.DeviceFaultDetail{constant.DeviceRetryFault: {
					FaultTime:    UceFaultTime,
					RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
					CompleteTime: constant.JobNotRecoverComplete}},
			}
			if i == 0 {
				faultDetail := uceDevice.FaultDetail[constant.DeviceRetryFault]
				faultDetail.CompleteTime = uceDevice.FaultDetail[constant.DeviceRetryFault].RecoverTime
				uceDevice.FaultDetail[constant.DeviceRetryFault] = faultDetail
			} else if i == 1 {
				faultDetail := uceDevice.FaultDetail[constant.DeviceRetryFault]
				faultDetail.CompleteTime = uceDevice.FaultDetail[constant.DeviceRetryFault].RecoverTime +
					constant.JobReportCompleteTimeout
				uceDevice.FaultDetail[constant.DeviceRetryFault] = faultDetail
			} else {
				faultDetail := uceDevice.FaultDetail[constant.DeviceRetryFault]
				faultDetail.CompleteTime = rand.Int63nRange(uceDevice.FaultDetail[constant.DeviceRetryFault].RecoverTime,
					uceDevice.FaultDetail[constant.DeviceRetryFault].RecoverTime+constant.JobReportCompleteTimeout+1)
				uceDevice.FaultDetail[constant.DeviceRetryFault] = faultDetail
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1,
				uceDevice.FaultDetail[constant.DeviceRetryFault].RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := RetryProcessor.canFilterRetryDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterRetryDeviceFaultInfo() = %v, want %v", got, want)
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
		uceDevice := constant.RetryDeviceInfo{
			DeviceName: "test-device",
			FaultDetail: map[string]constant.DeviceFaultDetail{constant.DeviceRetryFault: {
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
				CompleteTime: UceFaultTime - 1}},
		}
		currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1,
			uceDevice.FaultDetail[constant.DeviceRetryFault].RecoverTime+constant.JobReportCompleteTimeout+1)
		if got := RetryProcessor.canFilterRetryDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterRetryDeviceFaultInfo() = %v, want %v", got, want)
		}
	})

	// CompleteTime smaller than RecoverTime
	t.Run("TestUceFaultProcessorCanFilterUceDeviceFaultInfoSituation6-2", func(t *testing.T) {
		uceDevice := constant.RetryDeviceInfo{
			DeviceName: "test-device",
			FaultDetail: map[string]constant.DeviceFaultDetail{constant.DeviceRetryFault: {
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
				CompleteTime: constant.JobNotRecoverComplete}},
		}
		faultDetail := uceDevice.FaultDetail[constant.DeviceRetryFault]
		faultDetail.CompleteTime = faultDetail.RecoverTime - 1

		currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1,
			faultDetail.RecoverTime+constant.JobReportCompleteTimeout+1)
		if got := RetryProcessor.canFilterRetryDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterRetryDeviceFaultInfo() = %v, want %v", got, want)
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
			uceDevice := constant.RetryDeviceInfo{
				DeviceName: "test-device",
				FaultDetail: map[string]constant.DeviceFaultDetail{constant.DeviceRetryFault: {
					FaultTime:   UceFaultTime,
					RecoverTime: UceFaultTime - 1}},
			}
			if i == 0 {
				faultDetail := uceDevice.FaultDetail[constant.DeviceRetryFault]
				faultDetail.CompleteTime = faultDetail.FaultTime
				uceDevice.FaultDetail[constant.DeviceRetryFault] = faultDetail
			} else if i == 1 {
				faultDetail := uceDevice.FaultDetail[constant.DeviceRetryFault]
				faultDetail.CompleteTime = faultDetail.RecoverTime + constant.JobReportCompleteTimeout
				uceDevice.FaultDetail[constant.DeviceRetryFault] = faultDetail
			} else {
				faultDetail := uceDevice.FaultDetail[constant.DeviceRetryFault]
				faultDetail.CompleteTime = rand.Int63nRange(faultDetail.FaultTime,
					faultDetail.RecoverTime+constant.JobReportCompleteTimeout+1)
				uceDevice.FaultDetail[constant.DeviceRetryFault] = faultDetail
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1,
				uceDevice.FaultDetail[constant.DeviceRetryFault].RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := RetryProcessor.canFilterRetryDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterRetryDeviceFaultInfo() = %v, want %v", got, want)
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
			uceDevice := constant.RetryDeviceInfo{
				DeviceName: "test-device",
				FaultDetail: map[string]constant.DeviceFaultDetail{constant.DeviceRetryFault: {
					FaultTime:   UceFaultTime,
					RecoverTime: UceFaultTime - int64(31*time.Second)}},
			}
			faultDetail := uceDevice.FaultDetail[constant.DeviceRetryFault]
			if i == 0 {
				faultDetail.CompleteTime = faultDetail.FaultTime
			} else if i == 1 {
				faultDetail.CompleteTime = faultDetail.RecoverTime + constant.JobReportCompleteTimeout
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1,
				faultDetail.RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := RetryProcessor.canFilterRetryDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterRetryDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

func TestClearFault(t *testing.T) {
	t.Run("TestClearFault success", func(t *testing.T) {
		nodeName := "test-node"
		deviceName := "test-device"
		jobId := "test-job"
		fault := constant.DeviceFault{}
		curTime := time.Now().Unix()
		switchFault := constant.SimpleSwitchFaultInfo{
			AlarmRaisedTime: curTime - 100,
		}
		retryInfo := constant.RetryDeviceInfo{
			DeviceName: deviceName,
			FaultDetail: map[string]constant.DeviceFaultDetail{
				constant.DeviceRetryFault: {
					FaultType: constant.HcclFaultType,
				},
			},
		}
		RetryProcessor.onceRetryDeviceInfo = map[string]map[string]constant.RetryDeviceInfo{
			nodeName: {
				deviceName: retryInfo,
			},
		}
		RetryProcessor.onceFaultMap = map[string]map[string]constant.DeviceFault{
			nodeName: {
				deviceName: fault,
			},
		}
		RetryProcessor.linkdownDeviceFault = map[string]map[string]constant.DeviceFault{
			nodeName: {
				deviceName: fault,
			},
		}
		RetryProcessor.linkdownSwitchFault = map[string]constant.SimpleSwitchFaultInfo{
			constant.SwitchInfoPrefix + nodeName: switchFault,
		}
		RetryProcessor.jobServerInfoMap = constant.JobServerInfoMap{
			InfoMap: map[string]map[string]constant.ServerHccl{
				jobId: {
					nodeName: {},
				},
			},
		}
		RetryProcessor.clearFault(jobId)
		if len(RetryProcessor.onceRetryDeviceInfo) != 0 {
			t.Errorf("clearFault() onceRetryDeviceInfo = %v, want %v", RetryProcessor.onceRetryDeviceInfo, 0)
		}
	})
}

// =============Test scenario===========
func TestUceFaultProcessorGetUceDeviceOfNodes(t *testing.T) {
	t.Run("TestUceFaultProcessorGetUceDeviceOfNodes", func(t *testing.T) {
		cmDeviceInfos, _, uceNodesInfos, _, _, testFileErr := readObjectFromUceProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}

		RetryProcessor.nodeDeviceCmMap = faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos)
		deviceOfNodes := RetryProcessor.handleRetryDeviceOfNodes()
		for nodeName, filterNode := range deviceOfNodes {
			for deviceName, filterDevice := range filterNode.DeviceInfo {
				filterDevice.FaultCodeLevel = nil
				deviceOfNodes[nodeName].DeviceInfo[deviceName] = filterDevice
			}
		}
		if !reflect.DeepEqual(deviceOfNodes, uceNodesInfos) {
			t.Errorf("handleRetryDeviceOfNodes() = %v, want %v",
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

		RetryProcessor.normalFaultDetailOfJob = make(map[string]constant.DeviceFaultDetail)
		RetryProcessor.jobServerInfoMap = jobServerInfoMap
		RetryProcessor.nodeDeviceCmMap = faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos)
		RetryProcessor.retryDeviceOfNode = RetryProcessor.handleRetryDeviceOfNodes()
		RetryProcessor.retryDevicesOfJob = RetryProcessor.getRetryDevicesForTolerateJobs(time.Now().UnixMilli())
		for jobId, filterJob := range RetryProcessor.retryDevicesOfJob {
			for nodeName, filterNode := range filterJob.RetryNode {
				for deviceName, filterDevice := range filterNode.DeviceInfo {
					filterDevice.FaultCodeLevel = nil
					RetryProcessor.retryDevicesOfJob[jobId].RetryNode[nodeName].DeviceInfo[deviceName] = filterDevice
				}
			}
		}
		if !reflect.DeepEqual(RetryProcessor.retryDevicesOfJob, expectUceJobsInfo) {
			t.Errorf("getRetryDevicesForTolerateJobs() = %v, want %v",
				util.ObjToString(RetryProcessor.retryDevicesOfJob), util.ObjToString(expectUceJobsInfo))
		}
	})
}

func TestProcessSwitchLinkDownFaults(t *testing.T) {
	t.Run("TestProcessSwitchLinkDownFaults", func(t *testing.T) {
		nodeName := "test-node"
		deviceName := "test-device"
		fault := constant.DeviceFault{}
		curTime := time.Now().Unix()
		switchFault := constant.SimpleSwitchFaultInfo{
			AlarmRaisedTime: curTime - 100,
		}
		RetryProcessor.onceFaultMap = map[string]map[string]constant.DeviceFault{
			nodeName: {
				deviceName: fault,
			},
		}
		RetryProcessor.linkdownDeviceFault = map[string]map[string]constant.DeviceFault{
			nodeName: {
				deviceName: fault,
			},
		}
		RetryProcessor.nodeSwitchCmMap = make(map[string]*constant.SwitchInfo)
		RetryProcessor.nodeSwitchCmMap[constant.SwitchInfoPrefix+"node1"] = &constant.SwitchInfo{
			CmName: constant.SwitchInfoPrefix + "node1",
			SwitchFaultInfo: constant.SwitchFaultInfo{
				FaultInfo: []constant.SimpleSwitchFaultInfo{
					{
						AssembledFaultCode: constant.SwitchLinkDownFaultCode,
						AlarmRaisedTime:    time.Now().Unix(),
					},
				},
			},
		}
		RetryProcessor.linkdownSwitchFault = map[string]constant.SimpleSwitchFaultInfo{
			constant.SwitchInfoPrefix + nodeName: switchFault,
		}
		RetryProcessor.processSwitchLinkDownFaults(nodeName)
	})
}

func TestUceFaultProcessorProcessUceFaultInfo(t *testing.T) {
	t.Run("TestUceFaultProcessorProcessUceFaultInfo", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, _, jobServerInfoMap, _, testFileErr := readObjectFromUceProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}

		RetryProcessor.normalFaultDetailOfJob = make(map[string]constant.DeviceFaultDetail)
		RetryProcessor.jobServerInfoMap = jobServerInfoMap
		RetryProcessor.nodeDeviceCmMap = faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos)
		RetryProcessor.retryDeviceOfNode = RetryProcessor.handleRetryDeviceOfNodes()
		RetryProcessor.retryDevicesOfJob = RetryProcessor.getRetryDevicesForTolerateJobs(time.Now().UnixMilli())
		currentTime := 109 * time.Second.Milliseconds()
		RetryProcessor.processRetryFaultInfo(currentTime)
		result := RetryProcessor.nodeDeviceCmMap
		want := faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](expectProcessedDeviceInfos)
		if !reflect.DeepEqual(result, want) {
			t.Errorf("orgcm:\n%v\n\nresult:\n%v\n\nwant:\n%v",
				util.ObjToString(faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos)),
				util.ObjToString(result), util.ObjToString(want))
		}
	})
}

func TestGetRetryDevicesForTolerateJobs(t *testing.T) {
	t.Run("TestGetRetryDevicesForTolerateJobs", func(t *testing.T) {
		cmDeviceInfos, _, _, jobServerInfoMap, _, testFileErr := readObjectFromUceProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		RetryProcessor.normalFaultDetailOfJob = make(map[string]constant.DeviceFaultDetail)
		RetryProcessor.jobServerInfoMap = jobServerInfoMap
		RetryProcessor.nodeDeviceCmMap = faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos)
		RetryProcessor.nodeSwitchCmMap = make(map[string]*constant.SwitchInfo)
		RetryProcessor.nodeSwitchCmMap[constant.SwitchInfoPrefix+"node1"] = &constant.SwitchInfo{
			CmName: constant.SwitchInfoPrefix + "node1",
			SwitchFaultInfo: constant.SwitchFaultInfo{
				FaultInfo: []constant.SimpleSwitchFaultInfo{
					{
						AssembledFaultCode: constant.SwitchLinkDownFaultCode,
						AlarmRaisedTime:    time.Now().Unix(),
					},
				},
			},
		}
		RetryProcessor.retryDeviceOfNode = RetryProcessor.handleRetryDeviceOfNodes()
		RetryProcessor.retryDevicesOfJob = RetryProcessor.getRetryDevicesForTolerateJobs(time.Now().UnixMilli())
		if len(RetryProcessor.linkdownSwitchFault) == 0 {
			t.Error("add linkdown fault failed")
		}

	})
}

func TestProcessRetryHCCLFaultInfo(t *testing.T) {
	t.Run("TestProcessRetryHCCLFaultInfo node clear linkdown", func(t *testing.T) {
		cmDeviceInfos, _, _, jobServerInfoMap, _, testFileErr := readObjectFromUceProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		RetryProcessor.normalFaultDetailOfJob = make(map[string]constant.DeviceFaultDetail)
		RetryProcessor.jobServerInfoMap = jobServerInfoMap
		RetryProcessor.nodeDeviceCmMap = faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos)
		RetryProcessor.retryDeviceOfNode = RetryProcessor.handleRetryDeviceOfNodes()
		RetryProcessor.retryDevicesOfJob = RetryProcessor.getRetryDevicesForTolerateJobs(time.Now().UnixMilli())
		currentTime := 109 * time.Second.Milliseconds()
		RetryProcessor.hasLinkDown = true
		RetryProcessor.onceFaultMap = map[string]map[string]constant.DeviceFault{
			"nodeName": {
				"dev": {},
			},
		}
		RetryProcessor.processRetryFaultInfo(currentTime)
		if RetryProcessor.hasLinkDown == false {
			t.Errorf("hasLinkDown is clear")
		}
	})
}

func TestProcessRetryHCCLFaultInfo2(t *testing.T) {
	t.Run("TestProcessRetryHCCLFaultInfo clear linkdown", func(t *testing.T) {
		cmDeviceInfos, _, _, jobServerInfoMap, _, testFileErr := readObjectFromUceProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		RetryProcessor.normalFaultDetailOfJob = make(map[string]constant.DeviceFaultDetail)
		RetryProcessor.jobServerInfoMap = jobServerInfoMap
		RetryProcessor.nodeDeviceCmMap = faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos)
		RetryProcessor.retryDeviceOfNode = RetryProcessor.handleRetryDeviceOfNodes()
		RetryProcessor.retryDevicesOfJob = RetryProcessor.getRetryDevicesForTolerateJobs(time.Now().UnixMilli())
		currentTime := 109 * time.Second.Milliseconds()
		RetryProcessor.hasLinkDown = true
		RetryProcessor.onceFaultMap = map[string]map[string]constant.DeviceFault{}
		RetryProcessor.processRetryFaultInfo(currentTime)
		if RetryProcessor.hasLinkDown == true {
			t.Errorf("hasLinkDown is not clear")
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

		RetryProcessor.normalFaultDetailOfJob = make(map[string]constant.DeviceFaultDetail)
		RetryProcessor.jobServerInfoMap = jobServerInfoMap
		RetryProcessor.nodeDeviceCmMap = faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos)
		RetryProcessor.retryDeviceOfNode = RetryProcessor.handleRetryDeviceOfNodes()
		retryInfo := make(map[string]map[string]map[string]constant.ReportInfo)
		util.DeepCopy(&retryInfo, reportInfos.RetryMap)
		collector.ReportInfoCollector = reportInfos
		currentTime := 100 * time.Second.Milliseconds()
		patch := gomonkey.ApplyMethodFunc(collector.ReportInfoCollector, "GetInfo",
			func(jobId string, nodeName string, deviceName string) constant.ReportInfo {
				return retryInfo[jobId][nodeName][deviceName]
			})
		defer patch.Reset()
		RetryProcessor.retryDevicesOfJob = RetryProcessor.getRetryDevicesForTolerateJobs(currentTime)
		RetryProcessor.processRetryFaultInfo(currentTime)
		result := RetryProcessor.nodeDeviceCmMap
		want := faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](expectProcessedDeviceInfos)
		if !reflect.DeepEqual(result, want) {
			t.Errorf("processRetryFaultInfo() = %v, \n\nwant %v",
				util.ObjToString(result), util.ObjToString(want))
		}
	})
}

func TestAddRetryFault(t *testing.T) {
	t.Run("add retry fault ok", func(t *testing.T) {
		nodeName := "test-node"
		deviceName := "test-device"
		deviceInfo := constant.AdvanceDeviceFaultCm{}
		retryInfo := constant.RetryDeviceInfo{
			DeviceName: deviceName,
			FaultDetail: map[string]constant.DeviceFaultDetail{
				constant.DeviceRetryFault: {
					FaultType: constant.HcclFaultType,
				},
			},
		}
		fault := constant.DeviceFault{}
		switchFault := constant.SimpleSwitchFaultInfo{}
		RetryProcessor.onceRetryDeviceInfo = map[string]map[string]constant.RetryDeviceInfo{
			nodeName: {
				deviceName: retryInfo,
			},
		}
		RetryProcessor.onceFaultMap = map[string]map[string]constant.DeviceFault{
			nodeName: {
				deviceName: fault,
			},
		}
		RetryProcessor.linkdownDeviceFault = map[string]map[string]constant.DeviceFault{
			nodeName: {
				deviceName: fault,
			},
		}
		RetryProcessor.linkdownSwitchFault = map[string]constant.SimpleSwitchFaultInfo{
			constant.SwitchInfoPrefix + nodeName: switchFault,
		}
		length := len(RetryProcessor.onceRetryDeviceInfo)
		RetryProcessor.addRetryFault(nodeName, &deviceInfo, deviceName)
		if len(RetryProcessor.onceRetryDeviceInfo) < length {
			t.Error("addRetryFault() failed")
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
		retryInfo := make(map[string]map[string]map[string]constant.ReportInfo)
		util.DeepCopy(&retryInfo, reportInfos.RetryMap)
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
		patch := gomonkey.ApplyMethodFunc(collector.ReportInfoCollector, "GetInfo",
			func(jobId string, nodeName string, deviceName string) constant.ReportInfo {
				return retryInfo[jobId][nodeName][deviceName]
			})
		defer patch.Reset()
		defer func() {
			mockJob.Reset()
			mockNow.Reset()
			mockUnixMilli.Reset()
		}()

		resultContent := RetryProcessor.Process(content).(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
		result := resultContent.AllConfigmap
		want := faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](expProcessedDeviceInfos)
		if !reflect.DeepEqual(result, want) {
			t.Errorf("processRetryFaultInfo() = %v, want %v",
				util.ObjToString(result), util.ObjToString(want))
		}
	})
}

func TestUceFaultProcessorScenario3(t *testing.T) {
	t.Run("TestUceFaultProcessorScenario3", func(t *testing.T) {
		cmDeviceInfos, _, jobServerInfoMap, reportInfos, testFileErr :=
			readObjectFromUceScenarioTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		content := constant.OneConfigmapContent[*constant.SwitchInfo]{
			AllConfigmap: map[string]*constant.SwitchInfo{},
		}
		devContent := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
			AllConfigmap: faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](cmDeviceInfos),
			UpdateConfigmap: []constant.InformerCmItem[*constant.AdvanceDeviceFaultCm]{
				{
					IsAdd: false,
					Data:  nil},
			},
		}
		retryInfo := make(map[string]map[string]map[string]constant.ReportInfo)
		util.DeepCopy(&retryInfo, reportInfos.RetryMap)
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
		patch := gomonkey.ApplyMethodFunc(collector.ReportInfoCollector, "GetInfo",
			func(jobId string, nodeName string, deviceName string) constant.ReportInfo {
				return retryInfo[jobId][nodeName][deviceName]
			})
		defer patch.Reset()
		defer func() {
			mockJob.Reset()
			mockNow.Reset()
			mockUnixMilli.Reset()
		}()
		RetryProcessor.nodeDeviceCmMap = devContent.AllConfigmap
		resultContent := RetryProcessor.Process(content).(constant.OneConfigmapContent[*constant.SwitchInfo])
		if resultContent.AllConfigmap == nil {
			t.Error("result is nil ")
		}

	})
}

func readObjectFromUceProcessorTestYaml() (
	map[string]*constant.DeviceInfo, map[string]*constant.DeviceInfo,
	map[string]constant.RetryNodeInfo, constant.JobServerInfoMap, map[string]constant.RetryJobInfo, error) {

	var testDataPath = "../../../../../testdata/resource/uce_fault_processor_test.yaml"
	var maxFileSize = 20000
	var cmDeviceInfos = make(map[string]*constant.DeviceInfo)
	var expectProcessedDeviceInfos = make(map[string]*constant.DeviceInfo)
	var uceNodesInfos = make(map[string]constant.RetryNodeInfo)
	var expectUceJobsInfo = make(map[string]constant.RetryJobInfo)
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
	expectProcessedDeviceInfos map[string]*constant.DeviceInfo, uceNodesInfos map[string]constant.RetryNodeInfo,
	jobServerInfo constant.JobServerInfoMap, expectUceJobsInfo map[string]constant.RetryJobInfo,
	jobDevices map[string]map[string]constant.ServerHccl, jobIsUce map[string]bool) (map[string]*constant.DeviceInfo,
	map[string]*constant.DeviceInfo, map[string]constant.RetryNodeInfo, constant.JobServerInfoMap,
	map[string]constant.RetryJobInfo, error) {
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
	jobServerInfo.RetryTolerate = jobIsUce

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
	jobServerInfo.RetryTolerate = jobIsUce

	err = decoder.Decode(&reportInfos)
	if err != nil {
		err = fmt.Errorf("reportInfos decode failed")
	}
	return cmDeviceInfos, expectProcessedDeviceInfos, jobServerInfo, &reportInfos, err
}

func TestUpdateNormalFaultDetailOfJob(t *testing.T) {
	const jobName = "job"
	current := time.Now().UnixMilli()
	RetryProcessor.normalFaultDetailOfJob = make(map[string]constant.DeviceFaultDetail)
	t.Run("TestUpdateNormalFaultDetailOfJob, data not exist, update success", func(t *testing.T) {
		detail := constant.DeviceFaultDetail{
			FaultTime: current, ReportTime: constant.JobShouldReportFault, HasFaultAboveL3: true,
		}
		target := constant.DeviceFaultDetail{
			FaultTime: current, ReportTime: constant.JobShouldReportFault, HasFaultAboveL3: true, HasRank0Fault: false,
		}
		RetryProcessor.updateNormalFaultDetailOfJob(jobName, &detail, 1, constant.JobShouldReportFault)
		result, ok := RetryProcessor.normalFaultDetailOfJob[jobName]
		if !ok {
			t.Error("update map failed")
		}
		if !reflect.DeepEqual(result, target) {
			t.Errorf("updateNormalFaultDetailOfJob() = %v, want %v", result, target)
		}

	})
	t.Run("TestUpdateNormalFaultDetailOfJob, data already exist, update success", func(t *testing.T) {
		detail := constant.DeviceFaultDetail{
			FaultTime: constant.JobShouldReportFault, ReportTime: current, HasFaultAboveL3: false,
		}
		target := constant.DeviceFaultDetail{
			FaultTime: current, ReportTime: current, HasFaultAboveL3: true, HasRank0Fault: true,
		}
		RetryProcessor.updateNormalFaultDetailOfJob(jobName, &detail, 0, current)
		result, ok := RetryProcessor.normalFaultDetailOfJob[jobName]
		if !ok {
			t.Error("update map failed")
		}
		if !reflect.DeepEqual(result, target) {
			t.Errorf("updateNormalFaultDetailOfJob() = %v, want %v", result, target)
		}
	})
}

const (
	job1, job2       = "job1", "job2"
	node1, node2     = "node1", "node2"
	device1, device2 = "device1", "device2"
)

func getMockRetryDeviceOfJobMap() map[string]constant.RetryJobInfo {

	return map[string]constant.RetryJobInfo{
		job1: {RetryNode: map[string]constant.RetryNodeInfo{
			node1: {DeviceInfo: map[string]constant.RetryDeviceInfo{
				device1: {
					FaultCodeLevel: map[string]string{"code1": "level1"},
				}},
			}},
		},
	}
}

func TestGetRetryDeviceFromJob(t *testing.T) {
	RetryProcessor.retryDevicesOfJob = getMockRetryDeviceOfJobMap()
	t.Run("GetRetryDeviceFromJob, do not contain jobKey, should return false", func(t *testing.T) {
		_, found := RetryProcessor.GetRetryDeviceFromJob(job2, node1, device2)
		if found {
			t.Errorf("GetRetryDeviceFromJob() = %v, want %v", found, false)
		}
	})
	t.Run("GetRetryDeviceFromJob, contain jobKey and nodeKey, should return false", func(t *testing.T) {
		_, found := RetryProcessor.GetRetryDeviceFromJob(job1, node1, device2)
		if found {
			t.Errorf("GetRetryDeviceFromJob() = %v, want %v", found, false)
		}
	})
	t.Run("GetRetryDeviceFromJob, get info success, should return true", func(t *testing.T) {
		_, found := RetryProcessor.GetRetryDeviceFromJob(job1, node1, device1)
		if !found {
			t.Errorf("GetRetryDeviceFromJob() = %v, want %v", found, true)
		}
	})
}

func TestCanDoRestartInPlace(t *testing.T) {
	currentTime := time.Now().UnixMilli()
	RetryProcessor.normalFaultDetailOfJob = map[string]constant.DeviceFaultDetail{
		"job1": {HasFaultAboveL3: true},
		"job2": {HasRank0Fault: true},
		"job3": {ReportTime: currentTime, FaultTime: currentTime - 1},
		"job4": {FaultTime: currentTime - 1},
		"job5": {FaultTime: currentTime - constant.JobRestartInPlaceTimeout - 1},
	}
	t.Run("CanDoRestartInPlace, can not do restart in place", func(t *testing.T) {
		canDo := RetryProcessor.CanDoRestartInPlace("job0")
		canDo1 := RetryProcessor.CanDoRestartInPlace("job1")
		canDo2 := RetryProcessor.CanDoRestartInPlace("job2")
		canDo3 := RetryProcessor.CanDoRestartInPlace("job3")
		canDo4 := RetryProcessor.CanDoRestartInPlace("job5")
		if canDo || canDo1 || canDo2 || canDo3 || canDo4 {
			t.Errorf("CanDoRestartInPlace() = %v, want %v", canDo, false)
		}
	})
	t.Run("GetRetryDeviceFromJob, CanDoRestartInPlace, can do restart in place", func(t *testing.T) {
		canDo := RetryProcessor.CanDoRestartInPlace("job4")
		if !canDo {
			t.Errorf("CanDoRestartInPlace() = %v, want %v", canDo, true)
		}
	})
}

func TestGetFilterFaultCodeAndLevel(t *testing.T) {
	RetryProcessor.retryDevicesOfJob = getMockRetryDeviceOfJobMap()
	t.Run("GetFilterFaultCodeAndLevel, get map success", func(t *testing.T) {
		faultLevelMap := RetryProcessor.GetFilterFaultCodeAndLevel(job1, node1, device1)
		if faultLevelMap == nil {
			t.Errorf("GetFilterFaultCodeAndLevel() = %v, want: should not be nil", faultLevelMap)
		}
	})
}

func TestJobHasFault(t *testing.T) {
	RetryProcessor.retryDevicesOfJob = getMockRetryDeviceOfJobMap()
	t.Run("JobHasFault, job has fault, should return true", func(t *testing.T) {
		hasFault := RetryProcessor.JobHasFault(job1)
		if !hasFault {
			t.Errorf("JobHasFault() = %v, want %v", hasFault, true)
		}
	})
	t.Run("JobHasFault, job has no fault, should return false", func(t *testing.T) {
		hasFault := RetryProcessor.JobHasFault(job2)
		if hasFault {
			t.Errorf("JobHasFault() = %v, want %v", hasFault, false)
		}
	})
}
