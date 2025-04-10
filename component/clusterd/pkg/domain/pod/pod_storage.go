// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package pod a series of pod storage function
package pod

import (
	"sync"

	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/util"
)

const (
	maxPodNum  = 1000000
	initPodNum = 1000
	torTag     = "isSharedTor"
	sharedTor  = "1"
	initJobNum = 100
)

var podManager Manager

// Manager use for pod data manager
type Manager struct {
	podMap      map[string]v1.Pod
	jobPodMap   map[string]map[string]v1.Pod
	podMapMutex sync.RWMutex
}

func init() {
	podManager.podMap = make(map[string]v1.Pod, initPodNum)
	podManager.jobPodMap = make(map[string]map[string]v1.Pod, initJobNum)
	podManager.podMapMutex = sync.RWMutex{}
}

// SavePod save pod with lock, Please do not add time-consuming code
func SavePod(podInfo *v1.Pod) {
	podManager.podMapMutex.Lock()
	defer podManager.podMapMutex.Unlock()
	if len(podManager.podMap) > maxPodNum {
		hwlog.RunLog.Errorf("podMap length will exceed %d, pod namespace=%s, name=%s save failed",
			maxPodNum, podInfo.Namespace, podInfo.Name)
		return
	}
	podManager.podMap[GetPodKey(podInfo)] = *podInfo
	jobKey := GetJobKeyByPod(podInfo)
	if podManager.jobPodMap[jobKey] == nil {
		podManager.jobPodMap[jobKey] = map[string]v1.Pod{}
	}
	podManager.jobPodMap[jobKey][GetPodKey(podInfo)] = *podInfo
}

// DeletePod delete pod with lock, Please do not add time-consuming code
func DeletePod(podInfo *v1.Pod) {
	podManager.podMapMutex.Lock()
	delete(podManager.podMap, GetPodKey(podInfo))
	jobKey := GetJobKeyByPod(podInfo)
	if len(podManager.jobPodMap[jobKey]) > 0 {
		delete(podManager.jobPodMap[jobKey], GetPodKey(podInfo))
		if len(podManager.jobPodMap[jobKey]) == 0 {
			delete(podManager.jobPodMap, jobKey)
		}
	}
	podManager.podMapMutex.Unlock()
}

// GetPodByJobId get pod by jobId
func GetPodByJobId(jobKey string) map[string]v1.Pod {
	podManager.podMapMutex.RLock()
	defer podManager.podMapMutex.RUnlock()
	localPodMap := podManager.jobPodMap[jobKey]
	newPodMap := new(map[string]v1.Pod)
	err := util.DeepCopy(newPodMap, localPodMap)
	if err != nil {
		hwlog.RunLog.Errorf("copy podMap failed, err：%v", err)
	}
	return *newPodMap
}
