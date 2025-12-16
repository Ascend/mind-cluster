// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package pod a series of pod util function
package pod

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/node"
)

const (
	torIpTag     = "sharedTorIp"
	podLabelKey  = "app"
	podGroupKey  = "scheduling.k8s.io/group-name"
	acJobNameKey = "job-name"
	vcJobNameKey = "volcano.sh/job-name"
)

// GetJobKeyByPod get job unique key by pod
func GetJobKeyByPod(info *v1.Pod) string {
	if info == nil {
		hwlog.RunLog.Errorf("serious error, get unique key failed, pod is nil")
		return ""
	}
	for _, owner := range info.GetOwnerReferences() {
		if owner.Controller != nil && *owner.Controller {
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

// GetEnvByPod get pod env
func GetEnvByPod(pods map[string]v1.Pod, envName string) string {
	for _, po := range pods {
		find, env := getEnvFromPod(po, envName)
		if find {
			return env
		}
	}
	return ""
}

func getEnvFromPod(pod v1.Pod, envName string) (bool, string) {
	for _, container := range pod.Spec.Containers {
		find, env := getEnvFromContainer(container, envName)
		if find {
			return find, env
		}
	}
	return false, ""
}

func getEnvFromContainer(container v1.Container, envName string) (bool, string) {
	for _, env := range container.Env {
		if env.Name == envName {
			return true, env.Value
		}
	}
	return false, ""
}

// ConstructRankTableByPod construct rank table by pod
func ConstructRankTableByPod(
	podGroup v1beta1.PodGroup, podsInJob map[string]v1.Pod, replicas int) (constant.RankTable, int) {
	var rankTable constant.RankTable
	if replicas <= 0 {
		hwlog.RunLog.Error("illegal param replicas")
		return rankTable, 0
	}
	completedPodNum := 0
	rankTable.ServerList = make([]constant.ServerHccl, 0, replicas)
	ipSnMap := node.GetNodeIPAndSNMap()
	for _, pod := range podsInJob {
		podRank := getPodRank(pod)
		if podRank == -1 || podRank >= replicas {
			hwlog.RunLog.Warnf("illegal job information, replicas is %d, but podRank is %d", replicas, podRank)
			continue
		}
		podDev, isShouldAllocated := getPodDevice(pod)
		if len(podDev.Devices) > 0 || !isShouldAllocated {
			completedPodNum++
		}
		if len(podDev.Devices) == 0 {
			continue
		}
		scaleOutType, err := getScaleOutType(podGroup, podsInJob)
		if err != nil {
			hwlog.RunLog.Errorf("getScaleOutType failed: %v", err)
		}

		serverInfo := getServerInfo(scaleOutType, podDev, pod, ipSnMap, podRank)
		rankTable.ServerList = append(rankTable.ServerList, serverInfo)
	}
	sort.Slice(rankTable.ServerList, func(i, j int) bool {
		iRankID, iErr := strconv.Atoi(rankTable.ServerList[i].DeviceList[0].RankID)
		jRankID, jErr := strconv.Atoi(rankTable.ServerList[j].DeviceList[0].RankID)
		if iErr != nil || jErr != nil {
			return false
		}
		return iRankID < jRankID
	})
	rankTable.ServerCount = strconv.Itoa(len(rankTable.ServerList))
	return rankTable, completedPodNum
}

func getServerInfo(scaleOutType string, podDev constant.PodDevice,
	pod v1.Pod, ipSnMap map[string]string, podRank int) constant.ServerHccl {
	server := constant.ServerHccl{
		ServerID:     podDev.ServerID,
		PodID:        podDev.PodName,
		PodNameSpace: pod.Namespace,
		ServerName:   pod.Spec.NodeName,
		SuperPodId:   podDev.SuperPodId,
		ServerSN:     getSN(pod.Spec.NodeName, podDev.ServerID, ipSnMap),
	}

	podDevNum := len(podDev.Devices)
	for idx, dev := range podDev.Devices {
		serverDev := constant.Device{
			DeviceID:      dev.DeviceID,
			DeviceIP:      dev.DeviceIP,
			RankID:        strconv.Itoa(podRank*podDevNum + idx),
			SuperDeviceID: dev.SuperDeviceID,
		}
		setScaleOutNetwork(dev, scaleOutType, &serverDev)
		server.DeviceList = append(server.DeviceList, serverDev)
	}
	return server
}

func getSN(serverName, serverID string, ipSnMap map[string]string) string {
	if serverName != "" {
		return node.GetNodeSNByName(serverName)
	}
	if serverID != "" {
		return ipSnMap[serverID]
	}
	hwlog.RunLog.Warn("server name and server id are both empty")
	return ""
}

func getPodDevice(pod v1.Pod) (constant.PodDevice, bool) {
	devInfo, exist := pod.Annotations[api.Pod910DeviceAnno]
	if !exist {
		return constant.PodDevice{}, shouldAllocated(pod.Spec.Containers)
	}
	var podDev constant.PodDevice
	if err := json.Unmarshal([]byte(devInfo), &podDev); err != nil {
		hwlog.RunLog.Errorf("parse annotation of pod %s/%s error: %v", pod.Namespace, pod.Name, err)
	}
	return podDev, true
}

func getPodRank(pod v1.Pod) int {
	rankIndexStr, ok := pod.Annotations[api.PodRankIndexAnno]
	if !ok {
		hwlog.RunLog.Errorf("get pod annotation %s failed, pod: %s", api.PodRankIndexAnno, pod.Name)
		return -1
	}
	rankIndexInt, err := strconv.Atoi(rankIndexStr)
	if err != nil {
		hwlog.RunLog.Errorf("string to int failed, pod: %s, err: %v", pod.Name, err)
		return -1
	}
	return rankIndexInt
}

// GetPodDeviceNumByJobId get pod device num by job Key
func GetPodDeviceNumByJobId(jobKey string) int {
	podManager.podMapMutex.RLock()
	defer podManager.podMapMutex.RUnlock()
	podsInJob := podManager.jobPodMap[jobKey]

	// is inference task?
	var isInference bool

	for _, pod := range podsInJob {
		podDev, _ := getPodDevice(pod)
		if pod.Labels != nil && pod.Labels[constant.MindIeJobIdLabelKey] != "" &&
			pod.Labels[constant.MindIeAppTypeLabelKey] != "" {
			isInference = true
		}
		if len(podDev.Devices) != 0 {
			return len(podDev.Devices)
		}
	}
	// inference task, return 0
	if isInference || len(podsInJob) == 0 {
		return 0
	}
	hwlog.RunLog.Warnf("failed get pod device num, job key: %s, len(podsInJob): %d", jobKey, len(podsInJob))
	return 0
}

// GetPodByRankIndex get pod by rank index
func GetPodByRankIndex(jobKey, podRank string) v1.Pod {
	podsInJob := GetPodByJobId(jobKey)
	for _, pod := range podsInJob {
		if pod.Annotations[api.PodRankIndexAnno] == podRank {
			return pod
		}
	}
	return v1.Pod{}
}

// GetModelFramework get model framework
func GetModelFramework(podsInJob map[string]v1.Pod) string {
	for _, pod := range podsInJob {
		modelFramework := pod.GetLabels()
		framework, ok := modelFramework[podLabelKey]
		if ok {
			return framework
		}
	}
	hwlog.RunLog.Debug("get framework from pod failed")
	return ""
}

// GetMinMember get min member
func GetMinMember(podsInJob map[string]v1.Pod) int {
	for _, pod := range podsInJob {
		annos := pod.GetAnnotations()
		minMemberStr, ok := annos[api.MinAvailableKey]
		if !ok {
			continue
		}
		minMember, err := strconv.Atoi(minMemberStr)
		if err != nil || minMember <= 0 {
			hwlog.RunLog.Errorf("minMemberStr is invalid, pod: %s, err: %v, minMember:%s",
				pod.Name, err, minMemberStr)
			return 0
		}
		return minMember
	}
	return 0
}

// DeviceAllocateIsCompleted pod need to be allocated and have already been allocated
func DeviceAllocateIsCompleted(p v1.Pod) bool {
	// pod need to be allocated
	containers := p.Spec.Containers
	if len(containers) == 0 {
		return false
	}
	if !shouldAllocated(containers) {
		return true
	}
	// pod already been allocated
	_, exist := p.Annotations[api.Pod910DeviceAnno]
	return exist
}

func shouldAllocated(containers []v1.Container) bool {
	for _, container := range containers {
		resourceLimits := container.Resources.Limits
		if len(resourceLimits) == 0 {
			continue
		}
		for resourceName := range resourceLimits {
			if strings.Contains(resourceName.String(), api.ResourceNamePrefix) {
				return true
			}
		}
	}
	return false
}

// GetPGByPod get PodGroup by pod info
func GetPGByPod(jobKey string) (jobName, pgName, namespace string) {
	podsInJob := GetPodByJobId(jobKey)
	for _, po := range podsInJob {
		jobName, pgName, namespace = GetPGInfo(&po)
		if jobName != "" && pgName != "" && namespace != "" {
			return jobName, pgName, namespace
		}
	}

	hwlog.RunLog.Errorf("job(uid=%s) relative pods is empty, get pgName, jobName failed", jobKey)
	return
}

// ConstructServersByJobKey Construct server hccl list by job key
func ConstructServersByJobKey(jobKey string) map[string]constant.ServerHccl {
	podMap := GetPodByJobId(jobKey)
	servers := make(map[string]constant.ServerHccl, len(podMap))
	for _, pod := range podMap {
		if pod.Spec.NodeName == "" {
			hwlog.RunLog.Warnf("job: %s server pod %s is not scheduled", jobKey, pod.Name)
			return nil
		}
		nodeName := pod.Spec.NodeName
		servers[nodeName] = constant.ServerHccl{
			ServerID:     node.GetNodeIpByName(nodeName),
			PodID:        pod.Name,
			PodNameSpace: pod.Namespace,
			ServerName:   nodeName,
			ServerSN:     node.GetNodeSNByName(nodeName),
		}
	}
	return servers
}

// GetUsedDevicesByNodeName get used devices by node name
func GetUsedDevicesByNodeName(nodeName string) sets.String {
	podManager.podMapMutex.RLock()
	defer podManager.podMapMutex.RUnlock()

	usedDevice := sets.String{}
	if podManager.nodePodMap == nil {
		return usedDevice
	}
	if _, exist := podManager.nodePodMap[nodeName]; !exist {
		return usedDevice
	}
	for _, pod := range podManager.nodePodMap[nodeName] {
		if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded {
			continue
		}
		ascendReal, exist := pod.Annotations[api.PodAnnotationAscendReal]
		if !exist || ascendReal == "" {
			continue
		}
		usedDevice = usedDevice.Insert(strings.Split(ascendReal, constant.Comma)...)
	}
	return usedDevice
}
