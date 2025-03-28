// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package statistics a series of statistic function
package statistics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
)

func TestUpdateStatistic(t *testing.T) {
	now := time.Now().Unix()
	baseJobStc := constant.JobStatistic{PodFirstRunningTime: 0, PodLastFaultTime: 0}
	t.Run("Pending status - no change", func(t *testing.T) {
		jobInfo := constant.JobInfo{Status: job.StatusJobPending}
		result := UpdateStatistic(baseJobStc, jobInfo)
		assert.Equal(t, job.StatusJobPending, result.Status)
	})
	t.Run("Running status - first running", func(t *testing.T) {
		jobInfo := constant.JobInfo{
			Status: job.StatusJobRunning,
			PreServerList: []constant.ServerHccl{{DeviceList: []constant.Device{{DeviceID: "1"}, {DeviceID: "2"}}},
				{DeviceList: []constant.Device{{DeviceID: "3"}}}}}
		result := UpdateStatistic(baseJobStc, jobInfo)
		assert.NotZero(t, result.PodFirstRunningTime)
		assert.Equal(t, int64(3), result.CardNums)
	})
	t.Run("Running status - recover from fault", func(t *testing.T) {
		jobStc := baseJobStc
		jobStc.PodLastFaultTime = now - 1
		jobStc.PodFirstRunningTime = now - 1 - 1
		jobInfo := constant.JobInfo{Status: job.StatusJobRunning}
		result := UpdateStatistic(jobStc, jobInfo)
		assert.NotZero(t, result.PodLastRunningTime)
	})
	t.Run("Completed status - set stop time", func(t *testing.T) {
		jobInfo := constant.JobInfo{Status: job.StatusJobCompleted}
		result := UpdateStatistic(baseJobStc, jobInfo)
		assert.NotZero(t, result.StopTime)
	})
	t.Run("Failed status - normal failure", func(t *testing.T) {
		jobInfo := constant.JobInfo{Status: job.StatusJobFail}
		result := UpdateStatistic(baseJobStc, jobInfo)
		assert.NotZero(t, result.PodLastFaultTime)
		assert.Equal(t, int64(1), result.PodFaultTimes)
	})
	t.Run("Failed status with pre-delete", func(t *testing.T) {
		jobInfo := constant.JobInfo{Status: job.StatusJobFail, IsPreDelete: true}
		result := UpdateStatistic(baseJobStc, jobInfo)
		assert.Equal(t, result.PodLastFaultTime, result.StopTime)
	})
	t.Run("Unknown status - no change", func(t *testing.T) {
		jobInfo := constant.JobInfo{Status: "Unknown"}
		result := UpdateStatistic(baseJobStc, jobInfo)
		assert.Equal(t, baseJobStc.PodFirstRunningTime, result.PodFirstRunningTime)
	})
}

func TestInitStatistic(t *testing.T) {
	t.Run("With CustomJobID", func(t *testing.T) {
		jobInfo := constant.JobInfo{
			CustomJobID: "custom-123",
			Name:        "test-job",
			NameSpace:   "default",
			Status:      job.StatusJobPending,
		}

		result := InitStatistic(jobInfo, "k8s-123")
		assert.Equal(t, "custom-123", result.CustomJobID)
		assert.Equal(t, "test-job", result.Name)
		assert.Equal(t, job.StatusJobPending, result.Status)
	})

	t.Run("Without CustomJobID", func(t *testing.T) {
		jobInfo := constant.JobInfo{
			Name:      "k8s-job",
			NameSpace: api.KubeNS,
			Status:    job.StatusJobRunning,
		}

		result := InitStatistic(jobInfo, "k8s-456")
		assert.Equal(t, "k8s-456", result.K8sJobID)
		assert.Equal(t, "k8s-job", result.Name)
		assert.Zero(t, result.CardNums)
	})

	t.Run("Verify default values", func(t *testing.T) {
		jobInfo := constant.JobInfo{Status: job.StatusJobPending}
		result := InitStatistic(jobInfo, "test")
		assert.Zero(t, result.PodFirstRunningTime)
		assert.Zero(t, result.StopTime)
		assert.Zero(t, result.PodFaultTimes)
	})
}
