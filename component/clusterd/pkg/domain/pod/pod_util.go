// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package pod a series of pod util function
package pod

import (
	"encoding/json"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

var (
	podDeviceKey       = "ascend.kubectl.kubernetes.io/ascend-910-configuration"
	podRankIndexKey    = "hccl/rankIndex"
	torIpTag           = "sharedTorIp"
	podLabelKey        = "app"
	resourceNamePrefix = "huawei.com/"
	podGroupKey        = "scheduling.k8s.io/group-name"
	acJobNameKey       = "job-name"
	vcJobNameKey       = "volcano.sh/job-name"
)

// GetJobKeyByPod get job unique key by pod
func GetJobKeyByPod(info *v1.Pod) string {
	if info == nil {
		hwlog.RunLog.Errorf("serious error, get unique key failed, pod is nil")
		return ""
	}
	for _, owner := range info.GetOwnerReferences() {
		if *owner.Controller {
			return string(owner.UID)
		}
	}
	hwlog.RunLog.Errorf("serious error, get unique key failed, pod don't have controller")
	return ""
}

// GetPodKey get pod unique key
func GetPodKey(info *v1.Pod) string {
	if info == nil {
		hwlog.RunLog.Errorf("serious error, get unique key failed, pod is nil")
		return ""
	}
	return string(info.UID)
}

// GetPGInfo get podGroup name and namespace and job name by pod info
func GetPGInfo(info *v1.Pod) (jobName, pgName, namespace string) {
	if info == nil {
		hwlog.RunLog.Error("serious error, get podgroup name and namespace failed, pod is nil")
		return "", "", ""
	}

	// get pg name
	annotations := info.GetAnnotations()
	pgName, ok := annotations[podGroupKey]
	if !ok {
		hwlog.RunLog.Errorf("serious error, get podgroup name failed, "+
			"pod(ns=%s, name=%s) doesn`t exist %s annotation", info.Namespace, info.Name, podGroupKey)
		return "", "", ""
	}

	// get jobName
	labels := info.GetLabels()
	jobName, ok = labels[vcJobNameKey]
	if !ok {
		jobName, ok = labels[acJobNameKey]
		if !ok {
			hwlog.RunLog.Errorf("serious error, get job name failed, pod(ns=%s, name=%s) "+
				"doesn`t exist %s or %s label", info.Namespace, info.Name, acJobNameKey, vcJobNameKey)
			return "", "", ""
		}
	}

	return jobName, pgName, info.Namespace
}

// GetSharedTorIpByPod get shared tor ip by pod
func GetSharedTorIpByPod(pods map[string]v1.Pod) string {
	sharedTorIps := make([]string, 0)
	sharedTorIpMap := make(map[string]struct{}, 0)
	for _, pod := range pods {
		if pod.Annotations == nil {
			continue
		}
		k, ok := pod.Annotations[torTag]
		if !ok || k != sharedTor {
			continue
		}
		sharedTorIp, ok := pod.Annotations[torIpTag]
		if !ok {
			continue
		}
		if _, ok := sharedTorIpMap[sharedTorIp]; ok {
			continue
		}
		sharedTorIps = append(sharedTorIps, sharedTorIp)
		sharedTorIpMap[sharedTorIp] = struct{}{}
	}
	if len(sharedTorIps) == 0 {
		return ""
	}
	shardTorIpBytes, err := json.Marshal(sharedTorIps)
	if err != nil {
		hwlog.RunLog.Errorf("marshal sharedTorIps failed, err: %v", err)
		return ""
	}
	return string(shardTorIpBytes)
}

// InitRankTableByPod init rank table by pod
func InitRankTableByPod(podJobMap map[string]v1.Pod, replicas int) constant.RankTable {
	var rankTable constant.RankTable
	if replicas <= 0 {
		hwlog.RunLog.Error("illegal param replicas")
		return rankTable
	}
	rankTable.ServerList = make([]constant.ServerHccl, replicas)
	for _, pod := range podJobMap {
		nodeRank := getNodeRank(pod)
		if nodeRank == -1 || nodeRank >= replicas {
			hwlog.RunLog.Warnf("illegal job information, replicas is %d, but nodeRank is %d", replicas, nodeRank)
			continue
		}
		var server constant.ServerHccl
		podDevice := getPodDevice(pod)
		server.ServerID = podDevice.ServerID
		server.PodID = podDevice.PodName
		server.ServerName = pod.Spec.NodeName

		podDeviceNum := len(podDevice.Devices)
		for index, device := range podDevice.Devices {
			var serverDevice constant.Device
			serverDevice.DeviceID = device.DeviceID
			serverDevice.DeviceIP = device.DeviceIP
			serverDevice.RankID = strconv.Itoa(nodeRank*podDeviceNum + index)
			server.DeviceList = append(server.DeviceList, serverDevice)
		}
		rankTable.ServerList[nodeRank] = server
	}
	rankTable.ServerCount = strconv.Itoa(len(rankTable.ServerList))
	return rankTable
}

func getPodDevice(pod v1.Pod) constant.PodDevice {
	deviceInfo, exist := pod.Annotations[podDeviceKey]
	if !exist {
		return constant.PodDevice{}
	}
	var podDevice constant.PodDevice
	if err := json.Unmarshal([]byte(deviceInfo), &podDevice); err != nil {
		hwlog.RunLog.Errorf("parse annotation of pod %s/%s error: %v", pod.Namespace, pod.Name, err)
	}
	return podDevice
}

func getNodeRank(pod v1.Pod) int {
	rankIndexStr, rankExist := pod.Annotations[podRankIndexKey]
	if !rankExist {
		hwlog.RunLog.Errorf("get pod annotation %s failed, pod:%s", podRankIndexKey, pod.Name)
		return -1
	}
	intValue, err := strconv.Atoi(rankIndexStr)
	if err != nil {
		hwlog.RunLog.Errorf("string to int failed, pod:%s, err:%v", pod.Name, err)
		return -1
	}
	return intValue
}

// GetPodDeviceNumByJobId get pod device num by job Key
func GetPodDeviceNumByJobId(jobKey string) int {
	podManager.podMapMutex.RLock()
	defer podManager.podMapMutex.RUnlock()
	podJobMap := podManager.podJobMap[jobKey]
	for _, pod := range podJobMap {
		podDevice := getPodDevice(pod)
		if len(podDevice.Devices) != 0 {
			return len(podDevice.Devices)
		}
	}
	return 0
}

// GetPodByRankIndex get pod by rank index
func GetPodByRankIndex(jobKey, podRank string) v1.Pod {
	podJobMap := GetPodByJobId(jobKey)
	for _, pod := range podJobMap {
		if pod.Annotations[podRankIndexKey] == podRank {
			return pod
		}
	}
	return v1.Pod{}
}

// GetModelFramework get model framework
func GetModelFramework(podJobMap map[string]v1.Pod) string {
	for _, pod := range podJobMap {
		modelFramework := pod.GetLabels()
		framework, ok := modelFramework[podLabelKey]
		if ok {
			return framework
		}
	}
	hwlog.RunLog.Debug("get framework from pod failed")
	return ""
}

// DeviceAllocateIsCompleted pod need to be allocated and have already been allocated
func DeviceAllocateIsCompleted(p v1.Pod) bool {
	// pod need to be allocated
	containers := p.Spec.Containers
	if len(containers) == 0 {
		return true
	}
	shouldAllocated := false
	for _, container := range containers {
		resourceLimits := container.Resources.Limits
		if len(resourceLimits) == 0 {
			continue
		}
		for resourceName, resourceLimit := range resourceLimits {
			if resourceLimit.Value() > 0 && strings.Contains(resourceName.String(), resourceNamePrefix) {
				shouldAllocated = true
				break
			}
		}
		if shouldAllocated {
			break
		}
	}
	if !shouldAllocated {
		return true
	}
	// pod already been allocated
	_, exist := p.Annotations[podDeviceKey]
	return exist
}
