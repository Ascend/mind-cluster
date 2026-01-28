// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package recoverinplace contain filtering fault handling method for single process fault
package recoverinplace

import (
	"reflect"
	"sort"
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

const (
	podRank0 = constant.PodRankStrZero
	podRank1 = "1"
	podRank2 = "2"
	podRank3 = "3"
)

func TestUpdateNormalFaultDetailOfJob(t *testing.T) {
	const jobName = "job"
	current := time.Now().UnixMilli()
	RecoverInplaceProcessor.faultDetailOfJob = make(map[string]*constant.SingleProcessJobPodInfo)
	t.Run("TestUpdateNormalFaultDetailOfJob, data not exist, update success", func(t *testing.T) {
		detail := constant.DeviceFaultDetail{
			FaultTime: current, ReportTime: constant.JobShouldReportFault,
			FaultCodeLevel: map[string]string{"fakeCode": constant.RestartNPU},
		}
		target := &constant.DeviceFaultDetail{
			FaultTime: current, PodRankStr: podRank0, ReportTime: constant.JobShouldReportFault,
			FaultCodeLevel: map[string]string{"fakeCode": constant.RestartNPU},
		}
		RecoverInplaceProcessor.updateNormalFaultDetailOfJob(jobName, podRank0, &detail)
		result, ok := RecoverInplaceProcessor.faultDetailOfJob[jobName]
		if !ok {
			t.Error("update map failed")
		}
		if !reflect.DeepEqual(result.Pod[podRank0], target) {
			t.Errorf("updateNormalFaultDetailOfJob() = %v, want %v", result.Pod[podRank0], target)
		}
	})
	t.Run("TestUpdateNormalFaultDetailOfJob, data already exist, update success", func(t *testing.T) {
		detail := constant.DeviceFaultDetail{
			FaultTime: constant.JobShouldReportFault, ReportTime: current, FaultCodeLevel: map[string]string{},
		}
		target := &constant.DeviceFaultDetail{
			FaultTime: current, PodRankStr: podRank0, ReportTime: current,
			FaultCodeLevel: map[string]string{"fakeCode": constant.RestartNPU},
		}
		RecoverInplaceProcessor.updateNormalFaultDetailOfJob(jobName, podRank0, &detail)
		result, ok := RecoverInplaceProcessor.faultDetailOfJob[jobName]
		if !ok {
			t.Error("update map failed")
		}
		if !reflect.DeepEqual(result.Pod[podRank0], target) {
			t.Errorf("updateNormalFaultDetailOfJob() = %v, want %v", result.Pod[podRank0], target)
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

func TestGetJobRecoverdPodRanks(t *testing.T) {
	RecoverInplaceProcessor.faultDetailOfJob = getMockFaultDetailOfJob()
	t.Run("ignore subHealthy fault", func(t *testing.T) {
		ret, done := RecoverInplaceProcessor.GetJobUnRecoverdPodRanks(job1, constant.SubHealthyIngore)
		sort.Strings(ret)
		assert.Equal(t, false, done)
		assert.Equal(t, []string{podRank1, podRank2}, ret)
	})
	t.Run("not ignore subHealthy fault", func(t *testing.T) {
		ret, done := RecoverInplaceProcessor.GetJobUnRecoverdPodRanks(job1, constant.SubHealthyGraceExit)
		sort.Strings(ret)
		assert.Equal(t, false, done)
		assert.Equal(t, []string{podRank0, podRank1, podRank2}, ret)
	})
}

const (
	job1, job2       = "job1", "job2"
	node1, node2     = "node1", "node2"
	device1, device2 = "device1", "device2"
)

func getMockRetryDeviceOfJobMap() map[string]*constant.SingleProcessJobInfo {
	return map[string]*constant.SingleProcessJobInfo{
		job1: {Node: map[string]*constant.SingleProcessNodeInfo{
			node1: {DeviceInfo: map[string]*constant.SingleProcessDeviceInfo{
				device1: {
					FaultDetail: &constant.DeviceFaultDetail{
						FaultCodeLevel: map[string]string{"code1": "level1"},
					},
				}},
			}},
		},
	}
}

func getMockFaultDetailOfJob() map[string]*constant.SingleProcessJobPodInfo {
	return map[string]*constant.SingleProcessJobPodInfo{
		job1: {Pod: map[string]*constant.DeviceFaultDetail{
			podRank0: {FaultCodeLevel: map[string]string{"fakeCode": constant.SubHealthFault}},
			podRank1: {FaultCodeLevel: map[string]string{"fakeCode": constant.NotHandleFault, "fakeCode1": constant.RestartBusiness}},
			podRank2: {FaultCodeLevel: map[string]string{"fakeCode": constant.RestartNPU, "fakeCode1": constant.RestartBusiness}},
			podRank3: {FaultCodeLevel: map[string]string{"fakeCode": constant.NotHandleFault}},
		}},
	}
}

func TestCanDoRestartInPlace(t *testing.T) {
	currentTime := time.Now().UnixMilli()
	patch := gomonkey.ApplyFuncReturn(podgroup.GetSubHealthStrategyByJobKey, constant.SubHealthyGraceExit)
	defer patch.Reset()
	RecoverInplaceProcessor.faultDetailOfJob = map[string]*constant.SingleProcessJobPodInfo{
		"job1": {Pod: map[string]*constant.DeviceFaultDetail{
			podRank1: {FaultCodeLevel: map[string]string{"fakeCode": constant.RestartNPU}}}, JobId: "job1"},
		"job2": {Pod: map[string]*constant.DeviceFaultDetail{
			podRank0: {FaultCodeLevel: map[string]string{"fakeCode": constant.RestartNPU}}}, JobId: "job2"},
		"job3": {Pod: map[string]*constant.DeviceFaultDetail{podRank1: {
			FaultCodeLevel: map[string]string{"fakeCode": constant.RestartBusiness},
			ReportTime:     currentTime, FaultTime: currentTime - 1}}, JobId: "job3"},
		"job4": {Pod: map[string]*constant.DeviceFaultDetail{podRank1: {
			FaultCodeLevel: map[string]string{"fakeCode": constant.RestartBusiness},
			FaultTime:      currentTime - 1}}, JobId: "job4"},
		"job5": {Pod: map[string]*constant.DeviceFaultDetail{podRank1: {
			FaultCodeLevel: map[string]string{"fakeCode": constant.RestartBusiness},
			FaultTime:      currentTime - constant.JobRestartInPlaceTimeout - 1}}, JobId: "job5"},
	}
	t.Run("CanDoRestartInPlace, can not do restart in place", func(t *testing.T) {
		canDo := RecoverInplaceProcessor.CanDoRestartInPlace("job0", podRank1)
		canDo1 := RecoverInplaceProcessor.CanDoRestartInPlace("job1", podRank1)
		canDo2 := RecoverInplaceProcessor.CanDoRestartInPlace("job2", podRank1)
		canDo3 := RecoverInplaceProcessor.CanDoRestartInPlace("job3", podRank1)
		canDo4 := RecoverInplaceProcessor.CanDoRestartInPlace("job5", podRank1)
		if canDo || canDo1 || canDo2 || canDo3 || canDo4 {
			t.Errorf("CanDoRestartInPlace() = %v, want %v", canDo, false)
		}
	})
	t.Run("CanDoRestartInPlace, can do restart in place", func(t *testing.T) {
		canDo := RecoverInplaceProcessor.CanDoRestartInPlace("job4", podRank1)
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
		patch.ApplyMethodFunc(collector.ReportInfoCollector, "GetSingleProcessFaultReportTime", func(string, string) int64 {
			return currentTime
		})
		RecoverInplaceProcessor.DeviceOfNode = map[string]*constant.SingleProcessNodeInfo{
			nodeName: {
				NodeName: nodeName,
				DeviceInfo: map[string]*constant.SingleProcessDeviceInfo{
					deviceName: {
						FaultDetail: &constant.DeviceFaultDetail{
							FaultCodeLevel: map[string]string{"code1": "level1"},
						}},
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
		RecoverInplaceProcessor.faultDetailOfJob = make(map[string]*constant.SingleProcessJobPodInfo)
		defer func() {
			RecoverInplaceProcessor.DeviceOfNode = map[string]*constant.SingleProcessNodeInfo{}
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
			RecoverInplaceProcessor.DeviceOfNode = map[string]*constant.SingleProcessNodeInfo{}
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
		patch.ApplyPrivateMethod(RecoverInplaceProcessor, "initDeviceFromNodeAndReportInfo", func(jobId, nodeName string) *constant.SingleProcessNodeInfo {
			return &constant.SingleProcessNodeInfo{DeviceInfo: map[string]*constant.SingleProcessDeviceInfo{"modckDevice": {}}}
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
		RecoverInplaceProcessor.DevicesOfJob = map[string]*constant.SingleProcessJobInfo{
			jobID: {
				Node: map[string]*constant.SingleProcessNodeInfo{
					nodeName: {
						NodeName:   nodeName,
						DeviceInfo: map[string]*constant.SingleProcessDeviceInfo{deviceName: {FaultDetail: &constant.DeviceFaultDetail{}}},
					},
				},
			},
		}
		patch.ApplyFunc(podgroup.JudgeRestartProcessByJobKey, func(jobKey string) bool { return true })
		patch.ApplyFunc(faultdomain.SortDataForAdvanceDeviceInfo, func(deviceInfo *constant.AdvanceDeviceFaultCm) {})
		patch.ApplyPrivateMethod(RecoverInplaceProcessor, "initDeviceFromNodeAndReportInfo", func(jobId, nodeName string) *constant.SingleProcessNodeInfo {
			return &constant.SingleProcessNodeInfo{}
		})
		patch.ApplyPrivateMethod(RecoverInplaceProcessor, "canFilterNormalDeviceFaultInfo", func(string, string,
			int64, string) (bool, string) {
			return true, ""
		})
		called := false
		patch.ApplyPrivateMethod(RecoverInplaceProcessor, "filterNormalDeviceFaultInfo", func(deviceName string, advanceDevInfo *constant.AdvanceDeviceFaultCm) {
			called = true
			return
		})
		defer func() {
			RecoverInplaceProcessor.DevicesOfJob = map[string]*constant.SingleProcessJobInfo{}
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
		faultCode1 := "l5fault"
		deviceInfo := &constant.AdvanceDeviceFaultCm{
			FaultDeviceList: map[string][]constant.DeviceFault{
				deviceName: {
					{
						FaultLevel: constant.RestartRequest, FaultCode: faultCode, NPUName: deviceName, FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
							faultCode: {FaultTime: currentTime, FaultLevel: constant.RestartRequest},
						},
					},
					{
						FaultLevel: constant.SeparateNPU, FaultCode: faultCode1, NPUName: deviceName, FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
							faultCode1: {FaultTime: currentTime, FaultLevel: constant.SeparateNPU},
						},
					},
				},
			},
		}
		res := RecoverInplaceProcessor.getFaultDevices(nodeName, deviceInfo)
		assert.NotEqual(t, len(res.DeviceInfo), 0)
		faultCodeLevel := map[string]string{faultCode: constant.RestartRequest, faultCode1: constant.SeparateNPU}
		assert.Equal(t, faultCodeLevel, res.DeviceInfo[deviceName].FaultDetail.FaultCodeLevel)
	})
}
