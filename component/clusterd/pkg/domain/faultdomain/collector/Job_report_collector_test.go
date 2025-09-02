// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package collector collect information to process
package collector

import (
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"

	"ascend-common/api"
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
	DeviceName = api.Ascend910 + "-" + DeviceId
)

func TestMain(m *testing.M) {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	m.Run()
}

func TestJobReportInfoCollectorReportUceInfoSuccess(t *testing.T) {
	t.Run("report uce info success", func(t *testing.T) {
		mock := gomonkey.ApplyFunc(faultdomain.GetNodeAndDeviceFromJobIdAndRankId,
			func(jobId, rankId string, jobServerInfoMap constant.JobServerInfoMap) (string, string, error) {
				return NodeName, DeviceId, nil
			})
		defer mock.Reset()

		err := ReportInfoCollector.ReportRetryInfo(JobId, RankId, time.Now().UnixMilli(), "0")
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

		err := ReportInfoCollector.ReportRetryInfo(JobId, RankId, time.Now().UnixMilli(), "0")
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
						JobId: api.Ascend910,
					},
				}
			})
		defer func() {
			mockFaultdomain.Reset()
			mockJob.Reset()
		}()
		err := ReportInfoCollector.ReportRetryInfo(JobId, RankId, time.Now().UnixMilli(), "0")
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

func TestGetInfoCollectTimeout(t *testing.T) {
	t.Run("get report timeout info", func(t *testing.T) {
		oldMap := ReportInfoCollector.RetryMap
		defer func() {
			ReportInfoCollector.RetryMap = oldMap
		}()
		ReportInfoCollector.RetryMap = map[string]map[string]map[string]constant.ReportInfo{
			JobId: {
				NodeName: {
					DeviceName: {
						RecoverTime:  time.Now().UnixMilli() - dataExpireTime - 1,
						CompleteTime: constant.JobNotRecoverComplete,
					},
				},
			},
		}
		info := ReportInfoCollector.GetInfo(JobId, NodeName, DeviceName)
		if info.RecoverTime != constant.JobNotRecover {
			t.Error("get report info should be timeout")
		}
	})
}

func TestReportAndGetNoRetryReportTime(t *testing.T) {
	t.Run("get no retry report time success", func(t *testing.T) {
		reportTime := ReportInfoCollector.GetNoRetryReportTime(JobId)
		if reportTime != constant.JobShouldReportFault {
			t.Error("get no retry report time fail")
		}
		ReportInfoCollector.ReportNoRetryInfo(JobId, time.Now().UnixMilli())
		reportTime = ReportInfoCollector.GetNoRetryReportTime(JobId)
		if reportTime == constant.JobShouldReportFault {
			t.Error("get no retry report time fail")
		}
	})
}
