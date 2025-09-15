// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package statistics a series of pg storage function
package statistics

import (
	"sync"

	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
)

// Manager is statistics manager for job
type Manager struct {
	jobMap      map[string]v1.Object
	jobMapMutex sync.RWMutex
	maxCapacity int
}

const (
	initJobNum = 100
)

var (
	jobManager *Manager
)

func init() {
	jobManager = &Manager{
		jobMap:      make(map[string]v1.Object, initJobNum),
		maxCapacity: maxCMJobStatisticNum,
	}
}

// SaveJob save Job with lock, Please do not add time-consuming code
func SaveJob(job v1.Object) {
	if job == nil {
		hwlog.RunLog.Error("jobInfo is nil")
		return
	}
	jobManager.jobMapMutex.Lock()
	defer jobManager.jobMapMutex.Unlock()
	if len(jobManager.jobMap) >= jobManager.maxCapacity {
		hwlog.RunLog.Warnf(
			"jobMap length will exceed %d, namespace=%s, name=%s save failed",
			jobManager.maxCapacity,
			job.GetNamespace(),
			job.GetName(),
		)
		return
	}
	key := string(job.GetUID())
	jobManager.jobMap[key] = job
}

// DeleteJob delete Job with lock, Please do not add time-consuming code
func DeleteJob(job v1.Object) {
	if job == nil {
		hwlog.RunLog.Error("jobInfo is nil")
		return
	}
	jobManager.jobMapMutex.Lock()
	defer jobManager.jobMapMutex.Unlock()
	key := string(job.GetUID())
	delete(jobManager.jobMap, key)
}

// GetJob get Job with lock, Please do not add time-consuming code
func GetJob(key string) v1.Object {
	jobManager.jobMapMutex.RLock()
	defer jobManager.jobMapMutex.RUnlock()
	return jobManager.jobMap[key]
}
