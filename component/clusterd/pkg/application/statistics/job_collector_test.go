// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package statistics test for statistic funcs about fault
package statistics

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"

	"ascend-common/api/ascend-operator/apis/batch/v1"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/statistics"
)

func TestJobStatisticCollector(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	patches.ApplyMethodFunc(statistics.JobStcMgrInst, "LoadConfigMapToCache", func(_, _ string) {})
	patches.ApplyMethodFunc(statistics.JobStcMgrInst, "CheckJobScheduleTimeout", func(_ context.Context) {})
	patches.ApplyMethodFunc(statistics.JobStcMgrInst, "UpdateStcByPGCreate", func(jobKey string) {})
	patches.ApplyMethodFunc(statistics.JobStcMgrInst, "UpdateStcByPGUpdate", func(jobKey string) {})
	patches.ApplyMethodFunc(statistics.JobStcMgrInst, "PreDeleteJobStatistic", func(jobKey string) {})
	patches.ApplyMethodFunc(statistics.JobStcMgrInst, "DeleteJobStatistic", func(jobKey string) {})
	patches.ApplyMethodFunc(statistics.JobStcMgrInst, "JobStcByACJobCreate", func(jobKey string) {})
	patches.ApplyMethodFunc(statistics.JobStcMgrInst, "JobStcByACJobUpdate", func(jobKey string) {})
	patches.ApplyMethodFunc(statistics.JobStcMgrInst, "JobStcByJobDelete", func(jobKey string) {})
	patches.ApplyMethodFunc(statistics.JobStcMgrInst, "JobStcByVCJobCreate", func(jobKey string) {})
	go GlobalJobCollectMgr.JobCollector(ctx)
	testCases := []struct {
		msg constant.JobNotifyMsg
	}{
		{constant.JobNotifyMsg{Operator: constant.PGAdd, JobKey: "job1"}},
		{constant.JobNotifyMsg{Operator: constant.PGUpdate, JobKey: "job2"}},
		{constant.JobNotifyMsg{Operator: constant.PGDelete, JobKey: "job3"}},
		{constant.JobNotifyMsg{Operator: constant.JobInfoDelete, JobKey: "job4"}},
		{constant.JobNotifyMsg{Operator: constant.ACJobCreate, JobKey: "job5"}},
		{constant.JobNotifyMsg{Operator: constant.ACJobUpdate, JobKey: "job6"}},
		{constant.JobNotifyMsg{Operator: constant.ACJobDelete, JobKey: "job7"}},
		{constant.JobNotifyMsg{Operator: constant.VCJobCreate, JobKey: "job8"}},
		{constant.JobNotifyMsg{Operator: constant.VCJobDelete, JobKey: "job8"}},
		{constant.JobNotifyMsg{Operator: "not reach", JobKey: "job9"}},
	}
	for _, tc := range testCases {
		GlobalJobCollectMgr.JobNotify <- tc.msg
		time.Sleep(time.Second)
	}
}

func TestACJobInfoCollector(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	saveCall := false
	patches.ApplyFunc(statistics.SaveJob, func(job metav1.Object) {
		saveCall = true
	})
	deleteCall := false
	patches.ApplyFunc(statistics.DeleteJob, func(job metav1.Object) {
		deleteCall = true
	})
	patches.ApplyFunc(acJobMessage, func(oldJobInfo, newJobInfo *v1.AscendJob, operator string) {})
	t.Run("add acJob, new job info is nil", func(t *testing.T) {
		ACJobInfoCollector(&v1.AscendJob{}, nil, constant.AddOperator)
		assert.False(t, saveCall)
	})
	t.Run("add acJob, save job cache", func(t *testing.T) {
		ACJobInfoCollector(&v1.AscendJob{}, &v1.AscendJob{}, constant.AddOperator)
		assert.True(t, saveCall)
	})
	t.Run("update acJob, save job cache", func(t *testing.T) {
		ACJobInfoCollector(&v1.AscendJob{}, &v1.AscendJob{}, constant.UpdateOperator)
		assert.True(t, saveCall)
	})
	t.Run("delete acJob, delete job cache", func(t *testing.T) {
		ACJobInfoCollector(&v1.AscendJob{}, &v1.AscendJob{}, constant.DeleteOperator)
		assert.True(t, deleteCall)
	})
}

func TestAcJobMessage(t *testing.T) {
	t.Run("add acJob, notify add", func(t *testing.T) {
		acJobMessage(&v1.AscendJob{}, &v1.AscendJob{}, constant.AddOperator)
		notify := <-GlobalJobCollectMgr.JobNotify
		assert.Equal(t, constant.ACJobCreate, notify.Operator)
	})
	t.Run("update acJob, notify update", func(t *testing.T) {
		acJobMessage(&v1.AscendJob{}, &v1.AscendJob{}, constant.UpdateOperator)
		notify := <-GlobalJobCollectMgr.JobNotify
		assert.Equal(t, constant.ACJobUpdate, notify.Operator)
	})
	t.Run("delete acJob, notify delete", func(t *testing.T) {
		acJobMessage(&v1.AscendJob{}, &v1.AscendJob{}, constant.DeleteOperator)
		notify := <-GlobalJobCollectMgr.JobNotify
		assert.Equal(t, constant.ACJobDelete, notify.Operator)
	})
}

func TestVCJobInfoCollector(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	saveCall := false
	patches.ApplyFunc(statistics.SaveJob, func(job metav1.Object) {
		saveCall = true
	})
	deleteCall := false
	patches.ApplyFunc(statistics.DeleteJob, func(job metav1.Object) {
		deleteCall = true
	})
	patches.ApplyFunc(vcJobMessage, func(oldJobInfo, newJobInfo *v1alpha1.Job, operator string) {})
	t.Run("add vcJob, new job info is nil", func(t *testing.T) {
		VCJobInfoCollector(&v1alpha1.Job{}, nil, constant.AddOperator)
		assert.False(t, saveCall)
	})
	t.Run("add vcJob, save job cache", func(t *testing.T) {
		VCJobInfoCollector(&v1alpha1.Job{}, &v1alpha1.Job{}, constant.AddOperator)
		assert.True(t, saveCall)
	})
	t.Run("update vcJob, save job cache", func(t *testing.T) {
		VCJobInfoCollector(&v1alpha1.Job{}, &v1alpha1.Job{}, constant.UpdateOperator)
		assert.True(t, saveCall)
	})
	t.Run("delete vcJob, delete job cache", func(t *testing.T) {
		VCJobInfoCollector(&v1alpha1.Job{}, &v1alpha1.Job{}, constant.DeleteOperator)
		assert.True(t, deleteCall)
	})
}

func TestVcJobMessage(t *testing.T) {
	t.Run("add acJob, notify add", func(t *testing.T) {
		vcJobMessage(&v1alpha1.Job{}, &v1alpha1.Job{}, constant.AddOperator)
		notify := <-GlobalJobCollectMgr.JobNotify
		assert.Equal(t, constant.VCJobCreate, notify.Operator)
	})
	t.Run("delete acJob, notify delete", func(t *testing.T) {
		vcJobMessage(&v1alpha1.Job{}, &v1alpha1.Job{}, constant.DeleteOperator)
		notify := <-GlobalJobCollectMgr.JobNotify
		assert.Equal(t, constant.VCJobDelete, notify.Operator)
	})
}
