// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/jobprocess/faultrank"
	"clusterd/pkg/common/constant"
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

func TestRegister(t *testing.T) {
	t.Run("TestRegister", func(t *testing.T) {
		GlobalFaultProcessCenter.Register(make(chan int, 1), constant.AllProcessType)
		GlobalFaultProcessCenter.Register(make(chan int, 1), constant.DeviceProcessType)
		GlobalFaultProcessCenter.Register(make(chan int, 1), constant.NodeProcessType)
		GlobalFaultProcessCenter.Register(make(chan int, 1), constant.SwitchProcessType)
	})
}

func TestQueryJobsFaultInfo(t *testing.T) {
	t.Run("TestQueryJobsFaultInfo", func(t *testing.T) {
		patches := gomonkey.ApplyPrivateMethod(faultrank.JobFaultRankProcessor, "GetJobFaultRankInfosFilterLevel",
			func(faultLevel string) map[string]constant.JobFaultInfo {
				return map[string]constant.JobFaultInfo{"test": {}}
			})
		defer patches.Reset()
		jobsFaultInfo := QueryJobsFaultInfo([]string{constant.NotHandleFault})
		if len(jobsFaultInfo) != 1 {
			t.Error("TestQueryJobsFaultInfo fail")
		}
	})
}
