// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"testing"

	"clusterd/pkg/common/constant"
)

func TestGetJobFaultRankProcessor(t *testing.T) {
	t.Run("TestGetJobFaultRankProcessor", func(t *testing.T) {
		_, err := GlobalFaultProcessCenter.getJobFaultRankProcessor()
		if err != nil {
			t.Error("TestGetJobFaultRankProcessor fail")
		}
	})
}

func TestCallbackForReportUceInfo(t *testing.T) {
	t.Run("TestCallbackForReportUceInfo", func(t *testing.T) {
		infos := make([]ReportRecoverInfo, 0)
		infos = append(infos, ReportRecoverInfo{})
		err := GlobalFaultProcessCenter.CallbackForReportUceInfo(infos)
		if err != nil {
			t.Error("TestCallbackForReportUceInfo fail")
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

func TestQueryJobsFaultInfo(t *testing.T) {
	t.Run("TestQueryJobsFaultInfo", func(t *testing.T) {
		processor, _ := GlobalFaultProcessCenter.getJobFaultRankProcessor()
		processor.setJobFaultRankInfos(map[string]JobFaultInfo{"test": {}})
		jobsFaultInfo := GlobalFaultProcessCenter.QueryJobsFaultInfo(constant.NotHandleFault)
		if len(jobsFaultInfo) != 1 {
			t.Error("TestQueryJobsFaultInfo fail")
		}
	})
}
