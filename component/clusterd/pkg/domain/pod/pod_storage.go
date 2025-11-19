// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package pod a series of pod storage function
package pod

import (
	"sync"

	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const (
	maxPodNum  = 1000000
	initPodNum = 1000
	torTag     = "isSharedTor"
	sharedTor  = "1"
	initJobNum = 100
)

var (
	podManager                      Manager
	runningEventChan                = make(chan *v1.Pod, constant.MaxEventChanLen)
	deletedEventChan                = make(chan *v1.Pod, constant.MaxEventChanLen)
	hotSwitchStatusMap              = &sync.Map{}
	backupPodsAfterSourcePodDeleted = &sync.Map{}
)

// Manager use for pod data manager
type Manager struct {
	podMap      map[string]v1.Pod
	nodePodMap  map[string]map[string]v1.Pod // key: node name, value: pods in node
	jobPodMap   map[string]map[string]v1.Pod
	podMapMutex sync.RWMutex
}

func init() {
	podManager.podMap = make(map[string]v1.Pod, initPodNum)
	podManager.nodePodMap = make(map[string]map[string]v1.Pod)
	podManager.jobPodMap = make(map[string]map[string]v1.Pod, initJobNum)
	podManager.podMapMutex = sync.RWMutex{}
}

// SavePod save pod with lock, Please do not add time-consuming code
func SavePod(podInfo *v1.Pod) {
	if podInfo == nil {
		hwlog.RunLog.Error("podInfo is nil")
		return
	}
	podManager.podMapMutex.Lock()
	if len(podManager.podMap) > maxPodNum {
		hwlog.RunLog.Errorf("podMap length will exceed %d, pod namespace=%s, name=%s save failed",
			maxPodNum, podInfo.Namespace, podInfo.Name)
		return
	}

	podKey := GetPodKey(podInfo)
	jobKey := GetJobKeyByPod(podInfo)
	podManager.podMap[podKey] = *podInfo
	if podManager.nodePodMap[podInfo.Spec.NodeName] == nil {
		podManager.nodePodMap[podInfo.Spec.NodeName] = make(map[string]v1.Pod)
	}
	podManager.nodePodMap[podInfo.Spec.NodeName][podKey] = *podInfo
	if podManager.jobPodMap[jobKey] == nil {
		podManager.jobPodMap[jobKey] = map[string]v1.Pod{}
	}
	podManager.jobPodMap[jobKey][podKey] = *podInfo
	podManager.podMapMutex.Unlock()

	if IsNewPodForHotSwitch(podInfo) {
		if !hasHandled(podKey) && podInfo.Status.Phase == v1.PodRunning {
			hwlog.RunLog.Infof("hotswitch new pod running, jobKey=%s, podName=%s", jobKey, podInfo.Name)
			runningEventChan <- podInfo
			hotSwitchStatusMap.Store(podKey, struct{}{})
		}
	}
}

// DeletePod delete pod with lock, Please do not add time-consuming code
func DeletePod(podInfo *v1.Pod) {
	if podInfo == nil {
		hwlog.RunLog.Error("podInfo is nil")
		return
	}
	podManager.podMapMutex.Lock()
	delete(podManager.podMap, GetPodKey(podInfo))
	if podManager.nodePodMap[podInfo.Spec.NodeName] != nil {
		delete(podManager.nodePodMap[podInfo.Spec.NodeName], GetPodKey(podInfo))
		if len(podManager.nodePodMap[podInfo.Spec.NodeName]) == 0 {
			delete(podManager.nodePodMap, podInfo.Spec.NodeName)
		}
	}
	jobKey := GetJobKeyByPod(podInfo)
	if len(podManager.jobPodMap[jobKey]) > 0 {
		delete(podManager.jobPodMap[jobKey], GetPodKey(podInfo))
		if len(podManager.jobPodMap[jobKey]) == 0 {
			delete(podManager.jobPodMap, jobKey)
		}
	}
	podManager.podMapMutex.Unlock()
	value, exists := podInfo.Annotations[api.InHotSwitchFlowKey]
	if exists && value == api.InHotSwitchFlowValue {
		hwlog.RunLog.Infof("hotswitch pod deleted, jobKey=%s, podName=%s, phase:%s", jobKey, podInfo.Name, podInfo.Status.Phase)
		deletedEventChan <- podInfo
		// fault pod deleted, genn job summary info again
		backupNewPodName, ok := podInfo.Annotations[api.BackupNewPodNameKey]
		if ok {
			backupPodsAfterSourcePodDeleted.Store(backupNewPodName, struct{}{})
		}
	}
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

// GetPodByPodId get pod by podid
func GetPodByPodId(podId string) (v1.Pod, bool) {
	podManager.podMapMutex.RLock()
	defer podManager.podMapMutex.RUnlock()
	pod, exist := podManager.podMap[podId]
	if !exist {
		hwlog.RunLog.Errorf("pod %s is not exist in pod storage", podId)
		return v1.Pod{}, false
	}
	return pod, true
}

// GetPodByJobIdAndPodName get pod by jobId and pod name
func GetPodByJobIdAndPodName(jobKey, podName string) (v1.Pod, bool) {
	podManager.podMapMutex.RLock()
	defer podManager.podMapMutex.RUnlock()
	localPodMap := podManager.jobPodMap[jobKey]

	for _, pod := range localPodMap {
		if pod.Name == podName {
			return pod, true
		}
	}
	hwlog.RunLog.Errorf("pod %s is not exist in pod storage", podName)
	return v1.Pod{}, false
}

// GetSimplePodByJobId get pod by jobId
func GetSimplePodByJobId(jobKey string) map[string]*constant.SimplePodInfo {
	podManager.podMapMutex.RLock()
	defer podManager.podMapMutex.RUnlock()
	localPodMap := podManager.jobPodMap[jobKey]
	result := make(map[string]*constant.SimplePodInfo)
	for uid, pod := range localPodMap {
		// podRank may be empty string
		podRank := pod.Annotations[api.PodRankIndexAnno]
		result[uid] = &constant.SimplePodInfo{
			PodUid:  uid,
			PodRank: podRank,
		}
	}
	return result
}

// GetRunningEventChan get running event chan
func GetRunningEventChan() chan *v1.Pod {
	return runningEventChan
}

// GetDeletedEventChan get deleted event chan
func GetDeletedEventChan() chan *v1.Pod {
	return deletedEventChan
}

func IsNewPodForHotSwitch(pod *v1.Pod) bool {
	value, exists := pod.Annotations[api.PodTypeKey]
	return exists && value == api.PodTypeBackup
}

func hasHandled(podKey string) bool {
	_, exists := hotSwitchStatusMap.Load(podKey)
	return exists
}

// IsBackupPodAfterSourcePodDeleted check if backup pod deleted
func IsBackupPodAfterSourcePodDeleted(name string) bool {
	if _, exist := backupPodsAfterSourcePodDeleted.Load(name); exist {
		return true
	}
	return false
}

// DeleteFromBackupPodsMaps delete backup pod from map
func DeleteFromBackupPodsMaps(name string) {
	backupPodsAfterSourcePodDeleted.Delete(name)
}
