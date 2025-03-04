// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package jobv2 a series of queue function
package jobv2

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"k8s.io/api/core/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/statistics"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
)

const (
	queueOperatorAdd       = "add"
	queueOperatorUpdate    = "update"
	queueOperatorPreDelete = "preDelete"
	queueOperatorDelete    = "delete"
)

var uniqueQueue sync.Map
var limiter *rate.Limiter

const (
	limit               = 5
	burst               = 20
	messageNumThreshold = 5
)

func init() {
	limiter = rate.NewLimiter(limit, burst)
}

// Checker check if the queue is blocked, if not, set update message to queue for check job cache is right
func Checker(ctx context.Context) {
	hourTimer := time.NewTicker(time.Hour)
	minuteTimer := time.NewTicker(time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-hourTimer.C:
			if !checkQueueBlock() {
				addUpdateMessageIfOutdated()
			}
		case <-minuteTimer.C:
			preDeleteToDelete()
		}
	}
}

// preDeleteToDelete more than one minutes preDelete job should delete
func preDeleteToDelete() {
	deleteKeys := job.GetShouldDeleteJobKey()
	if len(deleteKeys) == 0 {
		return
	}
	for _, jobKey := range deleteKeys {
		uniqueQueue.Store(jobKey, queueOperatorDelete)
	}
}

// addUpdateMessageIfOutdated get all should update job key
func addUpdateMessageIfOutdated() {
	allKeys := job.GetShouldUpdateJobKey()
	if len(allKeys) == 0 {
		return
	}
	for _, jobKey := range allKeys {
		// flush LastUpdateTime whatever the configmap is updated or not
		job.FlushLastUpdateTime(jobKey)
		uniqueQueue.Store(jobKey, queueOperatorUpdate)
	}
}

func checkQueueBlock() bool {
	messageLength := 0
	uniqueQueue.Range(func(key, value interface{}) bool {
		messageLength++
		return true
	})
	if messageLength > messageNumThreshold {
		hwlog.RunLog.Errorf("queue blocking. more than %d pending messages, current: %d",
			messageNumThreshold, messageLength)
		return true
	}
	return false
}

// Handler handle message with limiter
func Handler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			jobUniqueKey := ""
			operator := ""
			uniqueQueue.Range(func(key, value interface{}) bool {
				ok := false
				jobUniqueKey, ok = key.(string)
				if !ok {
					return true
				}
				operator, ok = value.(string)
				if !ok {
					return true
				}
				// return false will stop range,
				// not first-in first-out, more than five job modifications per second may result in errors
				return false
			})
			if operator == "" {
				time.Sleep(time.Second)
				break
			}
			// wait token
			err := limiter.Wait(ctx)
			if err != nil {
				hwlog.RunLog.Errorf("limiter wait failed, err: %v", err)
				return
			}
			uniqueQueue.Delete(jobUniqueKey)
			switch operator {
			case queueOperatorAdd:
				addJob(jobUniqueKey)
				jobStcMessage(jobUniqueKey, constant.JobInfoAdd)
			case queueOperatorUpdate:
				updateJob(jobUniqueKey)
				jobStcMessage(jobUniqueKey, constant.JobInfoUpdate)
			case queueOperatorPreDelete:
				preDeleteJob(jobUniqueKey)
				jobStcMessage(jobUniqueKey, constant.JobInfoPreDelete)
			case queueOperatorDelete:
				jobStcMessage(jobUniqueKey, constant.JobInfoDelete)
				deleteJob(jobUniqueKey)
			default:
				hwlog.RunLog.Errorf("error operator: %s, jobKey: %s", operator, jobUniqueKey)
			}
		}
	}
}

// jobStcMessage notify to job statistic
func jobStcMessage(jobKey string, operator string) {
	notifyMsg := constant.JobNotifyMsg{Operator: operator, JobKey: jobKey}
	statistics.GlobalJobCollectMgr.JobNotifyChan <- notifyMsg
}

// podGroupMessage set job operator with pogGroup
func podGroupMessage(newPGInfo *v1beta1.PodGroup, operator string) {
	switch operator {
	case constant.AddOperator:
		uniqueQueue.Store(podgroup.GetJobKeyByPG(newPGInfo), queueOperatorAdd)
	case constant.DeleteOperator:
		uniqueQueue.Store(podgroup.GetJobKeyByPG(newPGInfo), queueOperatorPreDelete)
	case constant.UpdateOperator:
		uniqueQueue.Store(podgroup.GetJobKeyByPG(newPGInfo), queueOperatorUpdate)
	default:
		hwlog.RunLog.Errorf("abnormal informer operator: %s", operator)
	}
}

// podMessage set job operator with pod
func podMessage(oldPodInfo, newPodInfo *v1.Pod, operator string) {
	uniqueQueue.Store(pod.GetJobKeyByPod(newPodInfo), queueOperatorUpdate)
}
