// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package recoverinplace contain filtering fault handling method for single process fault
package recoverinplace

import (
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/faultdomain/collector"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
)

func TestMain(m *testing.M) {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	m.Run()
}

func TestUpdateNormalFaultDetailOfJob(t *testing.T) {
	const jobName = "job"
	current := time.Now().UnixMilli()
	RecoverInplaceProcessor.normalFaultDetailOfJob = make(map[string]constant.DeviceFaultDetail)
	t.Run("TestUpdateNormalFaultDetailOfJob, data not exist, update success", func(t *testing.T) {
		detail := constant.DeviceFaultDetail{
			FaultTime: current, ReportTime: constant.JobShouldReportFault, HasFaultAboveL3: true,
		}
		target := constant.DeviceFaultDetail{
			FaultTime: current, ReportTime: constant.JobShouldReportFault, HasFaultAboveL3: true, HasRank0Fault: false,
		}
		RecoverInplaceProcessor.updateNormalFaultDetailOfJob(jobName, &detail, 1, constant.JobShouldReportFault)
		result, ok := RecoverInplaceProcessor.normalFaultDetailOfJob[jobName]
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
		RecoverInplaceProcessor.updateNormalFaultDetailOfJob(jobName, &detail, 0, current)
		result, ok := RecoverInplaceProcessor.normalFaultDetailOfJob[jobName]
		if !ok {
			t.Error("update map failed")
		}
		if !reflect.DeepEqual(result, target) {
			t.Errorf("updateNormalFaultDetailOfJob() = %v, want %v", result, target)
		}
	})
}

func TestGetFilterFaultCodeAndLevel(t *testing.T) {
	RecoverInplaceProcessor.DevicesOfJob = getMockRetryDeviceOfJobMap()
	t.Run("GetFilterFaultCodeAndLevel, get map success", func(t *testing.T) {
		faultLevelMap := RecoverInplaceProcessor.GetFilterFaultCodeAndLevel(job1, node1, device1)
		if faultLevelMap == nil {
			t.Errorf("GetFilterFaultCodeAndLevel() = %v, want: should not be nil", faultLevelMap)
		}
	})
}

func TestJobHasFault(t *testing.T) {
	RecoverInplaceProcessor.DevicesOfJob = getMockRetryDeviceOfJobMap()
	t.Run("JobHasFault, job has fault, should return true", func(t *testing.T) {
		hasFault := RecoverInplaceProcessor.JobHasFault(job1)
		if !hasFault {
			t.Errorf("JobHasFault() = %v, want %v", hasFault, true)
		}
	})
	t.Run("JobHasFault, job has no fault, should return false", func(t *testing.T) {
		hasFault := RecoverInplaceProcessor.JobHasFault(job2)
		if hasFault {
			t.Errorf("JobHasFault() = %v, want %v", hasFault, false)
		}
	})
	RecoverInplaceProcessor.DevicesOfJob = getMockRetryDeviceOfJobMap1()
	t.Run("JobHasFault, job has fault but all can be ignore, should return false", func(t *testing.T) {
		hasFault := RecoverInplaceProcessor.JobHasFault(job1)
		if hasFault {
			t.Errorf("JobHasFault() = %v, want %v", hasFault, false)
		}
	})
	patch := gomonkey.ApplyFuncReturn(podgroup.GetSubHealthStrategyByJobKey, constant.SubHealthyGraceExit)
	defer patch.Reset()
	t.Run("JobHasFault, job has SubHealth fault and can not ignore, should return true", func(t *testing.T) {
		hasFault := RecoverInplaceProcessor.JobHasFault(job1)
		if !hasFault {
			t.Errorf("JobHasFault() = %v, want %v", hasFault, true)
		}
	})
}

const (
	job1, job2       = "job1", "job2"
	node1, node2     = "node1", "node2"
	device1, device2 = "device1", "device2"
)

func getMockRetryDeviceOfJobMap() map[string]constant.SingleProcessJobInfo {
	return map[string]constant.SingleProcessJobInfo{
		job1: {Node: map[string]constant.SingleProcessNodeInfo{
			node1: {DeviceInfo: map[string]constant.SingleProcessDeviceInfo{
				device1: {
					FaultCodeLevel: map[string]string{"code1": "level1"},
				}},
			}},
		},
	}
}

func getMockRetryDeviceOfJobMap1() map[string]constant.SingleProcessJobInfo {
	return map[string]constant.SingleProcessJobInfo{
		job1: {Node: map[string]constant.SingleProcessNodeInfo{
			node1: {DeviceInfo: map[string]constant.SingleProcessDeviceInfo{
				device1: {
					FaultCodeLevel: map[string]string{
						"code1": constant.NotHandleFault,
						"code2": constant.SubHealthFault,
					},
				}},
			}},
		},
	}
}

func TestCanDoRestartInPlace(t *testing.T) {
	currentTime := time.Now().UnixMilli()
	RecoverInplaceProcessor.normalFaultDetailOfJob = map[string]constant.DeviceFaultDetail{
		"job1": {HasFaultAboveL3: true},
		"job2": {HasRank0Fault: true},
		"job3": {ReportTime: currentTime, FaultTime: currentTime - 1},
		"job4": {FaultTime: currentTime - 1},
		"job5": {FaultTime: currentTime - constant.JobRestartInPlaceTimeout - 1},
	}
	t.Run("CanDoRestartInPlace, can not do restart in place", func(t *testing.T) {
		canDo := RecoverInplaceProcessor.CanDoRestartInPlace("job0")
		canDo1 := RecoverInplaceProcessor.CanDoRestartInPlace("job1")
		canDo2 := RecoverInplaceProcessor.CanDoRestartInPlace("job2")
		canDo3 := RecoverInplaceProcessor.CanDoRestartInPlace("job3")
		canDo4 := RecoverInplaceProcessor.CanDoRestartInPlace("job5")
		if canDo || canDo1 || canDo2 || canDo3 || canDo4 {
			t.Errorf("CanDoRestartInPlace() = %v, want %v", canDo, false)
		}
	})
	t.Run("GetRetryDeviceFromJob, CanDoRestartInPlace, can do restart in place", func(t *testing.T) {
		canDo := RecoverInplaceProcessor.CanDoRestartInPlace("job4")
		if !canDo {
			t.Errorf("CanDoRestartInPlace() = %v, want %v", canDo, true)
		}
	})
}

func TestInitDeviceFromNodeAndReportInfo(t *testing.T) {
	t.Run("initDeviceFromNodeAndReportInfo, ok", func(t *testing.T) {
		jobID := "jobID"
		nodeName := "nodeName"
		deviceName := "Ascend910-0"
		currentTime := time.Now().UnixMilli()
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(pod.GetPodDeviceNumByJobId, func(jobKey string) int {
			return 1
		})
		patch.ApplyMethodFunc(collector.ReportInfoCollector, "GetSingleProcessFaultReportTime", func(jobId string) int64 {
			return currentTime
		})
		RecoverInplaceProcessor.DeviceOfNode = map[string]constant.SingleProcessNodeInfo{
			nodeName: {
				NodeName: nodeName,
				DeviceInfo: map[string]constant.SingleProcessDeviceInfo{
					deviceName: {FaultCodeLevel: map[string]string{"code1": "level1"}},
				}},
		}
		RecoverInplaceProcessor.nodeDeviceCmMap = map[string]*constant.AdvanceDeviceFaultCm{
			nodeName: {DeviceType: "Ascend910"},
		}
		RecoverInplaceProcessor.jobServerInfoMap.InfoMap = map[string]map[string]constant.ServerHccl{
			jobID: {
				nodeName: {DeviceList: []constant.Device{{DeviceID: "0", RankID: "0"}}},
			},
		}
		defer func() {
			RecoverInplaceProcessor.DeviceOfNode = map[string]constant.SingleProcessNodeInfo{}
			RecoverInplaceProcessor.nodeDeviceCmMap = map[string]*constant.AdvanceDeviceFaultCm{}
			RecoverInplaceProcessor.jobServerInfoMap.InfoMap = map[string]map[string]constant.ServerHccl{}
		}()
		res := RecoverInplaceProcessor.initDeviceFromNodeAndReportInfo(jobID, nodeName)
		assert.NotEqual(t, res.DeviceInfo, 0)
	})
}

func TestProcess(t *testing.T) {
	t.Run("Process, data is err case", func(t *testing.T) {
		ori := constant.OneConfigmapContent[*constant.SwitchInfo]{}
		res := RecoverInplaceProcessor.Process(ori)
		assert.NotNil(t, res)
	})
	t.Run("Process, data is ok", func(t *testing.T) {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		oriDevInfo := make(map[string]*constant.DeviceInfo)
		oriDevInfo["nodeName"] = &constant.DeviceInfo{}
		content := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
			AllConfigmap:    faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo),
			UpdateConfigmap: nil,
		}
		defer func() {
			RecoverInplaceProcessor.DeviceOfNode = map[string]constant.SingleProcessNodeInfo{}
			RecoverInplaceProcessor.nodeDeviceCmMap = map[string]*constant.AdvanceDeviceFaultCm{}
			RecoverInplaceProcessor.jobServerInfoMap.InfoMap = map[string]map[string]constant.ServerHccl{}
		}()
		RecoverInplaceProcessor.Process(content)
		assert.NotEqual(t, len(RecoverInplaceProcessor.nodeDeviceCmMap), 0)
	})
}

func TestGetRetryDevicesForTolerateJobs(t *testing.T) {
	t.Run("getDevicesForTolerateJobs, data is ok", func(t *testing.T) {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		jobID := "jobID"
		nodeName := "nodeName"
		RecoverInplaceProcessor.nodeDeviceCmMap = map[string]*constant.AdvanceDeviceFaultCm{
			nodeName: {DeviceType: "Ascend910"},
		}
		RecoverInplaceProcessor.jobServerInfoMap.InfoMap = map[string]map[string]constant.ServerHccl{
			jobID: {
				nodeName: {DeviceList: []constant.Device{{DeviceID: "0", RankID: "0"}}},
			},
		}
		patch.ApplyFunc(podgroup.JudgeRestartProcessByJobKey, func(jobKey string) bool {
			return true
		})
		patch.ApplyPrivateMethod(RecoverInplaceProcessor, "initDeviceFromNodeAndReportInfo", func(jobId, nodeName string) constant.RetryNodeInfo {
			return constant.RetryNodeInfo{}
		})
		defer func() {
			RecoverInplaceProcessor.nodeDeviceCmMap = map[string]*constant.AdvanceDeviceFaultCm{}
			RecoverInplaceProcessor.jobServerInfoMap.InfoMap = map[string]map[string]constant.ServerHccl{}
		}()
		res := RecoverInplaceProcessor.getDevicesForTolerateJobs()
		assert.NotEqual(t, len(res), 0)
	})
}

func TestProcessEachNodeRetryFaultInfo(t *testing.T) {
	t.Run("processEachNodeFaultInfo, data is ok", func(t *testing.T) {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		jobID := "jobID"
		nodeName := "nodeName"
		deviceName := "Ascend910-0"
		deviceInfo := &constant.AdvanceDeviceFaultCm{}
		RecoverInplaceProcessor.DevicesOfJob = map[string]constant.SingleProcessJobInfo{
			jobID: {
				Node: map[string]constant.SingleProcessNodeInfo{
					nodeName: {
						NodeName:   nodeName,
						DeviceInfo: map[string]constant.SingleProcessDeviceInfo{deviceName: {}},
					},
				},
			},
		}
		patch.ApplyFunc(podgroup.JudgeRestartProcessByJobKey, func(jobKey string) bool { return true })
		patch.ApplyFunc(faultdomain.SortDataForAdvanceDeviceInfo, func(deviceInfo *constant.AdvanceDeviceFaultCm) {})
		patch.ApplyPrivateMethod(RecoverInplaceProcessor, "initDeviceFromNodeAndReportInfo", func(jobId, nodeName string) constant.RetryNodeInfo {
			return constant.RetryNodeInfo{}
		})
		patch.ApplyPrivateMethod(RecoverInplaceProcessor, "canFilterNormalDeviceFaultInfo", func(jobId string, retryDevice constant.RetryDeviceInfo,
			currentTime int64) bool {
			return true
		})
		called := false
		patch.ApplyPrivateMethod(RecoverInplaceProcessor, "filterNormalDeviceFaultInfo", func(deviceName string, advanceDevInfo *constant.AdvanceDeviceFaultCm) {
			called = true
			return
		})
		defer func() {
			RecoverInplaceProcessor.DevicesOfJob = map[string]constant.SingleProcessJobInfo{}
		}()
		RecoverInplaceProcessor.processEachNodeFaultInfo(nodeName, deviceInfo, time.Now().Unix())
		assert.Equal(t, called, true)
	})
}

func TestGetFaultDevices(t *testing.T) {
	t.Run("getFaultDevices, data is ok", func(t *testing.T) {
		nodeName := "nodeName"
		deviceName := "Ascend910-0"
		currentTime := time.Now().Unix()
		faultCode := "l3fault"
		deviceInfo := &constant.AdvanceDeviceFaultCm{
			FaultDeviceList: map[string][]constant.DeviceFault{
				deviceName: {
					{
						FaultLevel: constant.RestartRequest, FaultCode: faultCode, NPUName: deviceName, FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
							faultCode: {FaultTime: currentTime, FaultLevel: constant.RestartRequest},
						},
					},
					{
						FaultLevel: constant.RestartRequest, FaultCode: faultCode, NPUName: deviceName, FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
							faultCode: {FaultTime: currentTime, FaultLevel: constant.RestartRequest},
						},
					},
				},
			},
		}
		res := RecoverInplaceProcessor.getFaultDevices(nodeName, deviceInfo)
		assert.NotEqual(t, len(res.DeviceInfo), 0)
	})
}
