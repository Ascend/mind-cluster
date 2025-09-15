// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package statistics test for statistic funcs about fault
package statistics

import (
	"context"

	"volcano.sh/apis/pkg/apis/batch/v1alpha1"

	"ascend-common/api"
	"ascend-common/api/ascend-operator/apis/batch/v1"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/statistics"
)

const (
	jobNotifyChanLen = 1000
)

// JobCollectorMgr used to manage job statistic data.
type JobCollectorMgr struct {
	JobNotify chan constant.JobNotifyMsg
}

var (
	// GlobalJobCollectMgr is a global instance of StatisticInfo used for statistic data.
	GlobalJobCollectMgr *JobCollectorMgr
)

func init() {
	GlobalJobCollectMgr = &JobCollectorMgr{
		JobNotify: make(chan constant.JobNotifyMsg, jobNotifyChanLen),
	}
}

// JobCollector get the updated status of the job and update the statistics
func (j *JobCollectorMgr) JobCollector(ctx context.Context) {

	// load configMap into Cache if configMap is existed
	statistics.JobStcMgrInst.LoadConfigMapToCache(api.DLNamespace, statistics.JobStcCMName)
	go statistics.JobStcMgrInst.CheckJobScheduleTimeout(ctx)

	var handlers = map[string]func(string){
		constant.PGAdd:         statistics.JobStcMgrInst.UpdateStcByPGCreate,
		constant.PGUpdate:      statistics.JobStcMgrInst.UpdateStcByPGUpdate,
		constant.PGDelete:      statistics.JobStcMgrInst.PreDeleteJobStatistic,
		constant.JobInfoDelete: statistics.JobStcMgrInst.DeleteJobStatistic,
		constant.ACJobCreate:   statistics.JobStcMgrInst.JobStcByACJobCreate,
		constant.ACJobUpdate:   statistics.JobStcMgrInst.JobStcByACJobUpdate,
		constant.ACJobDelete:   statistics.JobStcMgrInst.JobStcByJobDelete,
		constant.VCJobCreate:   statistics.JobStcMgrInst.JobStcByVCJobCreate,
		constant.VCJobDelete:   statistics.JobStcMgrInst.JobStcByJobDelete,
	}
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("job Collector stop work")
			return
		case notifyMsg := <-j.JobNotify:
			handler, ok := handlers[notifyMsg.Operator]
			if !ok {
				hwlog.RunLog.Warnf("unexpected operator, JobKey: %s, Operator: %s",
					notifyMsg.JobKey, notifyMsg.Operator)
				continue
			}
			handler(notifyMsg.JobKey)
		}
	}
}

// ACJobInfoCollector collector acJob info
func ACJobInfoCollector(oldInfo, newInfo *v1.AscendJob, operator string) {
	if oldInfo == nil || newInfo == nil {
		hwlog.RunLog.Error("oldInfo or newInfo is nil")
		return
	}
	switch operator {
	case constant.AddOperator, constant.UpdateOperator:
		statistics.SaveJob(newInfo)
	case constant.DeleteOperator:
		statistics.DeleteJob(newInfo)
	default:
		hwlog.RunLog.Errorf("error operator: %s", operator)
		return
	}
	acJobMessage(oldInfo, newInfo, operator)
}

// acJobMessage set job operator with ascendJob
func acJobMessage(oldJobInfo, newJobInfo *v1.AscendJob, operator string) {
	jobKey := string(newJobInfo.UID)
	switch operator {
	case constant.AddOperator:
		GlobalJobCollectMgr.JobNotify <- constant.JobNotifyMsg{Operator: constant.ACJobCreate, JobKey: jobKey}
	case constant.UpdateOperator:
		GlobalJobCollectMgr.JobNotify <- constant.JobNotifyMsg{Operator: constant.ACJobUpdate, JobKey: jobKey}
	case constant.DeleteOperator:
		GlobalJobCollectMgr.JobNotify <- constant.JobNotifyMsg{Operator: constant.ACJobDelete, JobKey: jobKey}
	default:
		hwlog.RunLog.Errorf("abnormal informer operator: %s", operator)
	}
}

// VCJobInfoCollector collector acJob info
func VCJobInfoCollector(oldInfo, newInfo *v1alpha1.Job, operator string) {
	if oldInfo == nil || newInfo == nil {
		hwlog.RunLog.Error("oldInfo or newInfo is nil")
		return
	}
	switch operator {
	case constant.AddOperator, constant.UpdateOperator:
		statistics.SaveJob(newInfo)
	case constant.DeleteOperator:
		statistics.DeleteJob(newInfo)
	default:
		hwlog.RunLog.Errorf("error operator: %s", operator)
		return
	}
	vcJobMessage(oldInfo, newInfo, operator)
}

// vcJobMessage set job operator with ascendJob
func vcJobMessage(oldJobInfo, newJobInfo *v1alpha1.Job, operator string) {
	jobKey := string(newJobInfo.UID)
	switch operator {
	case constant.AddOperator:
		GlobalJobCollectMgr.JobNotify <- constant.JobNotifyMsg{Operator: constant.VCJobCreate, JobKey: jobKey}
	case constant.UpdateOperator: // nothing to do currently
	case constant.DeleteOperator:
		GlobalJobCollectMgr.JobNotify <- constant.JobNotifyMsg{Operator: constant.VCJobDelete, JobKey: jobKey}
	default:
		hwlog.RunLog.Errorf("abnormal informer operator: %s", operator)
	}
}
