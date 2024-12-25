// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
)

func TestJobRankFaultInfoProcessorCanDoStepRetry(t *testing.T) {
	t.Run("TestJobRankFaultInfoProcessorCanDoStepRetry", func(t *testing.T) {
		jobFaultRankProcessor, _ := GlobalFaultProcessCenter.getJobFaultRankProcessor()
		uceProcessor, _ := GlobalFaultProcessCenter.deviceCenter.getUceFaultProcessor()
		gomonkey.ApplyPrivateMethod(uceProcessor, "getUceDeviceFromJob",
			func(jobId, nodeName, deviceName string) (uceDeviceInfo, bool) {
				return uceDeviceInfo{
					DeviceName:   "test",
					FaultTime:    0,
					RecoverTime:  0,
					CompleteTime: 0,
				}, true
			})
		retry := jobFaultRankProcessor.canDoStepRetry("jobId", "nodeName", "deviceName")
		if !retry {
			t.Error("TestJobRankFaultInfoProcessorCanDoStepRetry")
		}
	})
}

func TestUceInBusinessPlane(t *testing.T) {
	t.Run("TestUceInBusinessPlane", func(t *testing.T) {
		jobFaultRankProcessor, _ := GlobalFaultProcessCenter.getJobFaultRankProcessor()
		uceProcessor, _ := GlobalFaultProcessCenter.deviceCenter.getUceFaultProcessor()
		gomonkey.ApplyPrivateMethod(uceProcessor, "getUceDeviceFromJob",
			func(jobId, nodeName, deviceName string) (uceDeviceInfo, bool) {
				return uceDeviceInfo{
					DeviceName:   "test",
					FaultTime:    0,
					RecoverTime:  0,
					CompleteTime: 0,
				}, true
			})
		isUceInBusinessPlane := jobFaultRankProcessor.uceInBusinessPlane("jobId", "nodeName", "deviceName")
		if isUceInBusinessPlane {
			t.Error("TestUceInBusinessPlane")
		}
	})
}
