// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"testing"

	"clusterd/pkg/application/faultmanager/uce"
	"clusterd/pkg/common/constant"
	"github.com/agiledragon/gomonkey/v2"
)

func TestJobRankFaultInfoProcessorCanDoStepRetry(t *testing.T) {
	t.Run("TestJobRankFaultInfoProcessorCanDoStepRetry", func(t *testing.T) {
		jobFaultRankProcessor, _ := GlobalFaultProcessCenter.getJobFaultRankProcessor()
		uceProcessor := uce.NewUceFaultProcessor()
		patches := gomonkey.ApplyPrivateMethod(uceProcessor, "GetUceDeviceFromJob",
			func(jobId, nodeName, deviceName string) (constant.UceDeviceInfo, bool) {
				return constant.UceDeviceInfo{
					DeviceName:   "test",
					FaultTime:    0,
					RecoverTime:  0,
					CompleteTime: 0,
				}, true
			})
		defer patches.Reset()
		retry := jobFaultRankProcessor.canDoStepRetry("jobId", "nodeName", "deviceName")
		if !retry {
			t.Error("TestJobRankFaultInfoProcessorCanDoStepRetry")
		}
	})
}

func TestUceInBusinessPlane(t *testing.T) {
	t.Run("TestUceInBusinessPlane", func(t *testing.T) {
		jobFaultRankProcessor, _ := GlobalFaultProcessCenter.getJobFaultRankProcessor()
		uceProcessor := uce.NewUceFaultProcessor()
		patches := gomonkey.ApplyPrivateMethod(uceProcessor, "GetUceDeviceFromJob",
			func(jobId, nodeName, deviceName string) (constant.UceDeviceInfo, bool) {
				return constant.UceDeviceInfo{
					DeviceName:   "test",
					FaultTime:    0,
					RecoverTime:  0,
					CompleteTime: 0,
				}, true
			})
		defer patches.Reset()
		isUceInBusinessPlane := jobFaultRankProcessor.uceInBusinessPlane("jobId", "nodeName", "deviceName")
		if isUceInBusinessPlane {
			t.Error("TestUceInBusinessPlane")
		}
	})
}
