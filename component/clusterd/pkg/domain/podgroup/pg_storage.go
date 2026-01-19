// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package podgroup a series of pg storage function
package podgroup

import (
	"sync"

	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
)

const (
	maxJobNum    = 100000
	initJobNum   = 100
	frameWorkKey = "framework"
)

var pgManager Manager

// Manager use for podGroup data manager
type Manager struct {
	pgMap      map[string]v1beta1.PodGroup
	pgMapMutex sync.RWMutex
}

func init() {
	pgManager.pgMap = make(map[string]v1beta1.PodGroup, initJobNum)
	pgManager.pgMapMutex = sync.RWMutex{}
}

// SavePodGroup save podGroup with lock, Please do not add time-consuming code
func SavePodGroup(pgInfo *v1beta1.PodGroup) {
	if pgInfo == nil {
		hwlog.RunLog.Error("pgInfo is nil")
		return
	}
	pgManager.pgMapMutex.Lock()
	defer pgManager.pgMapMutex.Unlock()
	if len(pgManager.pgMap) > maxJobNum {
		hwlog.RunLog.Errorf("pgMap length will exceed %d, podGroup namespace=%s, name=%s save failed",
			maxJobNum, pgInfo.Namespace, pgInfo.Name)
		return
	}
	pgManager.pgMap[GetJobKeyByPG(pgInfo)] = *pgInfo
}

// DeletePodGroup delete podGroup with lock, Please do not add time-consuming code
func DeletePodGroup(pgInfo *v1beta1.PodGroup) {
	pgManager.pgMapMutex.Lock()
	delete(pgManager.pgMap, GetJobKeyByPG(pgInfo))
	pgManager.pgMapMutex.Unlock()
}

// GetPodGroup get podGroup with lock, Please do not add time-consuming code
func GetPodGroup(jobKey string) v1beta1.PodGroup {
	pgManager.pgMapMutex.RLock()
	defer pgManager.pgMapMutex.RUnlock()
	return pgManager.pgMap[jobKey]
}

// CheckPodGroupExist check podGroup with lock, Please do not add time-consuming code
func CheckPodGroupExist(jobKey string) bool {
	pgManager.pgMapMutex.RLock()
	_, exist := pgManager.pgMap[jobKey]
	pgManager.pgMapMutex.RUnlock()
	return exist
}
