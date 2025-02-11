// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package faultrank contain job fault rank process
package faultrank

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/cmprocess/uce"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/kube"
)

const (
	// jobId job id
	jobId = "Job"
	// nodeName node name
	nodeName = "Node"
)

func TestMain(m *testing.M) {
	hwLogConfig := &hwlog.LogConfig{LogFileName: "../../../../../testdata/clusterd.log"}
	hwLogConfig.MaxBackups = hwlog.DefaultMaxBackups
	hwLogConfig.MaxAge = hwlog.DefaultMinSaveAge
	hwLogConfig.LogLevel = constant.DefaultLogLevel
	if err := hwlog.InitRunLogger(hwLogConfig, nil); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}
	m.Run()
}

func getDemoJobServerMap() constant.JobServerInfoMap {
	return constant.JobServerInfoMap{
		InfoMap: map[string]map[string]constant.ServerHccl{
			jobId: {
				nodeName: constant.ServerHccl{
					DeviceList: []constant.Device{{
						DeviceID: "0",
						RankID:   "0",
					}, {
						DeviceID: "1",
						RankID:   "1",
					}},
				},
			},
		},
	}
}

func TestFaultProcessorImplProcess(t *testing.T) {
	t.Run("test node fail, job fault rank list should correct", func(t *testing.T) {
		jobServerMap := getDemoJobServerMap()
		mockKube := gomonkey.ApplyFunc(kube.GetNode, func(name string) *v1.Node {
			return nil
		})
		mockJob := gomonkey.ApplyFunc(job.GetJobServerInfoMap, func() constant.JobServerInfoMap {
			return jobServerMap
		})
		defer func() {
			mockKube.Reset()
			mockJob.Reset()
		}()
		JobFaultRankProcessor.Process(constant.AllConfigmapContent{})
		faultRankInfos := JobFaultRankProcessor.GetJobFaultRankInfos()
		if len(faultRankInfos[jobId].FaultList) != len(jobServerMap.InfoMap[jobId][nodeName].DeviceList) {
			t.Error("TestFaultProcessorImplProcess fail")
		}
	})
}

func TestJobRankFaultInfoProcessorCanDoStepRetry(t *testing.T) {
	t.Run("TestJobRankFaultInfoProcessorCanDoStepRetry", func(t *testing.T) {
		patches := gomonkey.ApplyPrivateMethod(uce.UceProcessor, "GetUceDeviceFromJob",
			func(jobId, nodeName, deviceName string) (constant.UceDeviceInfo, bool) {
				return constant.UceDeviceInfo{
					DeviceName:   "test",
					FaultTime:    0,
					RecoverTime:  0,
					CompleteTime: 0,
				}, true
			})
		defer patches.Reset()
		retry := JobFaultRankProcessor.canDoStepRetry("jobId", "nodeName", "deviceName")
		if !retry {
			t.Error("TestJobRankFaultInfoProcessorCanDoStepRetry")
		}
	})
}

func TestUceInBusinessPlane(t *testing.T) {
	t.Run("TestUceInBusinessPlane", func(t *testing.T) {
		patches := gomonkey.ApplyPrivateMethod(uce.UceProcessor, "GetUceDeviceFromJob",
			func(jobId, nodeName, deviceName string) (constant.UceDeviceInfo, bool) {
				return constant.UceDeviceInfo{
					DeviceName:   "test",
					FaultTime:    0,
					RecoverTime:  0,
					CompleteTime: 0,
				}, true
			})
		defer patches.Reset()
		isUceInBusinessPlane := JobFaultRankProcessor.uceInBusinessPlane("jobId", "nodeName", "deviceName")
		if isUceInBusinessPlane {
			t.Error("TestUceInBusinessPlane")
		}
	})
}

func TestGetHealthState(t *testing.T) {
	convey.Convey("getHealthState", t, func() {
		convey.Convey("01-FaultLevel has SeparateNPU, should return UnHealthyState", func() {
			faultLilst := []constant.FaultRank{
				{FaultLevel: constant.SeparateNPU},
				{FaultLevel: constant.SubHealthFault}}
			convey.So(getHealthState(faultLilst, []string{}, nil),
				convey.ShouldEqual, constant.UnHealthyState)
		})
		convey.Convey("02-nodeStatusList has unHealthy status, should return UnHealthyState",
			func() {
				nodeStatusList := []string{constant.UnHealthyState, constant.SubHealthyState}
				status := getHealthState(nil, nodeStatusList, nil)
				convey.So(status == constant.UnHealthyState, convey.ShouldBeTrue)
			})
		convey.Convey("03-PodStrategiesMaps has Separate strategy, should return UnHealthyState",
			func() {
				PodStrategiesMaps := map[string]string{
					"pod1": constant.SeparateFaultStrategy,
					"pod2": constant.SubHealthFaultStrategy,
				}
				status := getHealthState(nil, nil, PodStrategiesMaps)
				convey.So(status == constant.UnHealthyState, convey.ShouldBeTrue)
			})
		convey.Convey("04-SubHealthy status, should return SubHealthyState", func() {
			PodStrategiesMaps := map[string]string{
				"pod2": constant.SubHealthFaultStrategy,
			}
			nodeStatusList := []string{constant.SubHealthyState}
			faultLilst := []constant.FaultRank{{FaultLevel: constant.SubHealthFault}}
			status := getHealthState(faultLilst, nodeStatusList, PodStrategiesMaps)
			convey.So(status == constant.SubHealthyState, convey.ShouldBeTrue)
		})
		convey.Convey("05-Healthy status, should return HealthyState", func() {
			status := getHealthState(nil, nil, nil)
			convey.So(status == constant.HealthyState, convey.ShouldBeTrue)
		})
	})
}
