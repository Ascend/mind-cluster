// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"testing"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/cmprocess/stresstest"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain/collector"
)

func TestMain(m *testing.M) {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	m.Run()
}

func TestCallbackForReportUceInfo(t *testing.T) {
	t.Run("TestCallbackForReportUceInfo", func(t *testing.T) {
		infos := make([]constant.ReportRecoverInfo, 0)
		infos = append(infos, constant.ReportRecoverInfo{})
		CallbackForReportRetryInfo(infos)
	})
}

func TestCallbackForReportNoRetryInfo(t *testing.T) {
	t.Run("CallbackForReportNoRetryInfo", func(t *testing.T) {
		currentTime := time.Now().UnixMilli()
		CallbackForReportNoRetryInfo("job1", []string{"pod1"}, currentTime)
		reportTime := collector.ReportInfoCollector.GetSingleProcessFaultReportTime("job1", "pod1")
		if reportTime != currentTime {
			t.Error("report no retry info failed")
		}
	})
}

func TestRegister(t *testing.T) {
	t.Run("TestRegister", func(t *testing.T) {
		GlobalFaultProcessCenter.Register(make(chan int, 1), constant.AllProcessType)
		GlobalFaultProcessCenter.Register(make(chan int, 1), constant.DeviceProcessType)
		GlobalFaultProcessCenter.Register(make(chan int, 1), constant.NodeProcessType)
		GlobalFaultProcessCenter.Register(make(chan int, 1), constant.SwitchProcessType)
	})
}

func TestFilterStressTestFault(t *testing.T) {
	t.Run("TestFilterStressTestFault, set filter ok", func(t *testing.T) {
		FilterStressTestFault("job", []string{"node"}, true)
		defer FilterStressTestFault("job", []string{"node"}, false)
		if stresstest.StressTestProcessor == nil {
			t.Error("TestFilterStressTestFault fail")
		}
	})
}
