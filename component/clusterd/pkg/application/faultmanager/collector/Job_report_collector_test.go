// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package collector contain information collect to process
package collector

import (
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/job"
)

const (
	JobId      = "Job"
	NodeName   = "Node"
	RankId     = "0"
	DeviceId   = "0"
	DeviceName = constant.Ascend910Server + "-" + DeviceId
)

func TestMain(m *testing.M) {
	hwLogConfig := &hwlog.LogConfig{LogFileName: "../../../testdata/clusterd.log"}
	hwLogConfig.MaxBackups = 30
	hwLogConfig.MaxAge = 7
	hwLogConfig.LogLevel = 0
	if err := hwlog.InitRunLogger(hwLogConfig, nil); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}
	InitReportInfoCollector()
	m.Run()
}

func TestJobReportInfoCollectorReportUceInfoSuccess(t *testing.T) {
	t.Run("report uce info success", func(t *testing.T) {
		mock := gomonkey.ApplyFunc(faultdomain.GetNodeAndDeviceFromJobIdAndRankId,
			func(jobId, rankId string, jobServerInfoMap constant.JobServerInfoMap) (string, string, error) {
				return NodeName, DeviceId, nil
			})
		defer mock.Reset()

		err := ReportInfoCollector.ReportUceInfo(JobId, RankId, time.Now().UnixMilli())
		if err != nil {
			t.Error(err)
		}
	})
}

func TestJobReportInfoCollectorReportUceInfoFail(t *testing.T) {
	t.Run("report uce info fail", func(t *testing.T) {
		mock := gomonkey.ApplyFunc(faultdomain.GetNodeAndDeviceFromJobIdAndRankId,
			func(jobId, rankId string, jobServerInfoMap constant.JobServerInfoMap) (string, string, error) {
				return "", "", fmt.Errorf("not find node and device from jobId %v and rankid %v", jobId, rankId)
			})
		defer mock.Reset()

		err := ReportInfoCollector.ReportUceInfo(JobId, RankId, time.Now().UnixMilli())
		if err == nil {
			t.Error("report uce info should fail")
		}
	})
}

func TestJobReportInfoCollectorGetInfo(t *testing.T) {
	t.Run("get report info success", func(t *testing.T) {
		mockFaultdomain := gomonkey.ApplyFunc(faultdomain.GetNodeAndDeviceFromJobIdAndRankId,
			func(jobId, rankId string, jobServerInfoMap constant.JobServerInfoMap) (string, string, error) {
				return NodeName, DeviceId, nil
			})
		mockJob := gomonkey.ApplyFunc(job.GetJobServerInfoMap,
			func() constant.JobServerInfoMap {
				return constant.JobServerInfoMap{
					ResourceType: map[string]string{
						JobId: constant.Ascend910Server,
					},
				}
			})
		defer func() {
			mockFaultdomain.Reset()
			mockJob.Reset()
		}()
		err := ReportInfoCollector.ReportUceInfo(JobId, RankId, time.Now().UnixMilli())
		if err != nil {
			t.Error(err)
		}
		reportInfo := ReportInfoCollector.GetInfo(JobId, NodeName, DeviceName)
		if reportInfo.RecoverTime == constant.JobNotRecover {
			t.Error("get report info fail")
		}
		reportInfo = ReportInfoCollector.GetInfoWithoutJobId(NodeName, DeviceName)
		if reportInfo.RecoverTime == constant.JobNotRecover {
			t.Error("get report info without jobid fail")
		}
	})
}
