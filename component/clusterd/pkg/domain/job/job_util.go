// Copyright (c) Huawei Technologies Co., Ltd. 2024-2026. All rights reserved.

// Package job a series of job function
package job

import (
	"encoding/json"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/custom"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
)

const (
	cmDataInitLength = 16
	maxHcclSliceSize = 800 * 1024
	vcJobKind        = "Job"
	masterAddr       = "MASTER_ADDR"
)

const (
	// StatusJobRunning is the running job status
	StatusJobRunning = "running"
	// StatusJobPending is the pending job status
	StatusJobPending = "pending"
	// StatusJobFail is the failed job status
	StatusJobFail = "failed"
	// StatusJobCompleted is the complete job status
	StatusJobCompleted = "complete"
)

const (
	// StatusRankTableInit is the init rankTable status
	StatusRankTableInit = "initializing"
	// StatusRankTableComplete is the complete rankTable status
	StatusRankTableComplete = "complete"
	// CustomJobID  custom id key
	CustomJobID = "custom-job-id"
)

// PreDeleteCmAndCache set job status
func PreDeleteCmAndCache(jobKey string) {
	jobInfo, ok := GetJobCache(jobKey)
	if !ok {
		return
	}
	if jobInfo.AddTime == 0 {
		jobInfo.AddTime = time.Now().Unix()
	}
	jobInfo.IsPreDelete = true
	// when a job is deleted, if it is not in a successful state, it must be in a failed state
	if jobInfo.Status != StatusJobCompleted {
		jobInfo.Status = StatusJobFail
	}
	jobInfo.DeleteTime = time.Now().Unix()
	jobInfo.LastUpdatedCmTime = time.Now().Unix()
	hccls := getHcclSlice(jobInfo.JobRankTable)
	if preDeleteCM(jobInfo, hccls) {
		hwlog.RunLog.Debugf("pre delete job:%s success", jobInfo.Name)
		SaveJobCache(jobKey, jobInfo)
	}
}

// DeleteCmAndCache delete job cm and cache info
func DeleteCmAndCache(jobKey string) {
	jobInfo, ok := GetJobCache(jobKey)
	if !ok {
		return
	}
	jobInfos := GetJobByNameSpaceAndNameAndPreDelete(jobInfo.Name, jobInfo.NameSpace, false)
	if len(jobInfos) > 0 {
		hwlog.RunLog.Infof("job(%s) with same name, only delete local cache", jobInfo.Name)
		DeleteJobCache(jobKey)
	} else if deleteCm(jobInfo) {
		hwlog.RunLog.Debugf("delete job:%s success", jobInfo.Name)
		DeleteJobCache(jobKey)
	}
}

// InitCmAndCache init cm and cache
func InitCmAndCache(podGroup v1beta1.PodGroup, podsInJob map[string]v1.Pod) {
	if len(podGroup.Name) == 0 || len(podGroup.GetOwnerReferences()) == 0 {
		hwlog.RunLog.Error("podGroup is nil, init configmap failed")
		return
	}
	// 1.init job basic info
	jobInfo := getJobBasicInfoByPG(podGroup, podsInJob)
	// 2.set job status info
	jobInfo.Status = StatusJobPending
	jobInfo.IsPreDelete = false
	jobInfo.JobRankTable = constant.RankTable{}
	jobInfo.LastUpdatedCmTime = time.Now().Unix()
	if initCM(jobInfo) {
		hwlog.RunLog.Debugf("init job:%s success", jobInfo.Name)
		SaveJobCache(jobInfo.Key, jobInfo)
	}

}

func getSidFromLabels(labels map[string]string) string {
	if labels == nil {
		return ""
	}

	customJobKeyVal, hasCustomJobKey := labels[constant.CustomJobKeyLabel]
	if hasCustomJobKey && strings.TrimSpace(customJobKeyVal) != "" {
		secondLevelVal, hasSecondLevelVal := labels[customJobKeyVal]
		if hasSecondLevelVal && strings.TrimSpace(secondLevelVal) != "" {
			return strings.TrimSpace(secondLevelVal)
		}
	}

	customJobIdVal, hasCustomJobId := labels[constant.CustomJobIdLabel]
	if hasCustomJobId && strings.TrimSpace(customJobIdVal) != "" {
		return strings.TrimSpace(customJobIdVal)
	}

	return ""
}

func getSidForJobInfo(pgInfo v1beta1.PodGroup, podsInJob map[string]v1.Pod) string {
	if sid := getSidFromLabels(pgInfo.Labels); sid != "" {
		return sid
	}
	for _, p := range podsInJob {
		if sid := getSidFromLabels(p.Labels); sid != "" {
			return sid
		}
	}
	sid := podgroup.GetJobKeyByPG(&pgInfo)
	hwlog.RunLog.Debugf("no sid found in labels, use jobId:%s", sid)
	return sid
}

func getJobBasicInfoByPG(pgInfo v1beta1.PodGroup, podsInJob map[string]v1.Pod) constant.JobInfo {
	var jobInfo constant.JobInfo
	key, name := podgroup.GetJobKeyAndNameByPG(&pgInfo)
	jobInfo.Key = key
	jobInfo.Name = name
	jobInfo.PgName = pgInfo.Name
	jobInfo.Replicas = max(int(pgInfo.Spec.MinMember), pod.GetMinMember(podsInJob))
	jobInfo.TotalCmNum = 1
	jobInfo.JobType = podgroup.GetJobTypeByPG(&pgInfo)
	jobInfo.NameSpace = pgInfo.Namespace
	jobInfo.Framework = podgroup.GetModelFramework(&pgInfo)
	jobInfo.ResourceType = podgroup.GetResourceType(&pgInfo)
	jobInfo.CustomJobID = pgInfo.Annotations[CustomJobID]
	jobInfo.MultiInstanceJobId = pgInfo.Labels[constant.MindIeJobIdLabelKey]
	jobInfo.AppType = pgInfo.Labels[constant.MindIeAppTypeLabelKey]
	jobInfo.AddTime = time.Now().Unix()
	jobInfo.Sid = getSidForJobInfo(pgInfo, podsInJob)
	return jobInfo
}

// UpdateCmAndCache update cm and cache
func UpdateCmAndCache(status string, jobKey string, podGroup v1beta1.PodGroup,
	podsInJob map[string]v1.Pod) {
	jobInfo, ok := GetJobCacheDeepCopy(jobKey)
	if !ok || jobInfo.Name == "" {
		jobInfo = getJobBasicInfoByPG(podGroup, podsInJob)
	}
	if jobInfo.AddTime == 0 {
		jobInfo.AddTime = time.Now().Unix()
	}
	jobInfo.Status = status
	jobInfo.IsPreDelete = false
	var completedPodNum int
	jobInfo.JobRankTable, completedPodNum = pod.ConstructRankTableByPod(podsInJob, jobInfo.Replicas)
	if jobInfo.ResourceType == api.NPULowerCase {
		customScaleOutType := strings.ToUpper(strings.TrimSpace(podGroup.Labels[constant.ScaleOutTypeLabel]))
		pod.ConstructRankListV2(&jobInfo.JobRankTable, podsInJob, jobInfo.Replicas, customScaleOutType)
	}
	if jobInfo.Framework == "" {
		// vcjob framework in pod label, it is empty when init jobInfo with podGroup
		jobInfo.Framework = pod.GetModelFramework(podsInJob)
	}
	jobInfo.LastUpdatedCmTime = time.Now().Unix()
	if completedPodNum == jobInfo.Replicas {
		jobInfo.JobRankTable.Status = StatusRankTableComplete
		jobInfo.PreServerList = jobInfo.JobRankTable.ServerList
		updateUseNodeNames(&jobInfo, podsInJob)
		initJobShareTorInfo(&jobInfo, podsInJob)
	} else {
		jobInfo.JobRankTable.Status = StatusRankTableInit
	}
	hccls := getHcclSlice(jobInfo.JobRankTable)
	jobInfo.TotalCmNum = len(hccls)
	if jobInfo.TotalCmNum == 0 {
		jobInfo.TotalCmNum = 1
	}
	jobInfo.JobRankTable.Total = jobInfo.TotalCmNum
	result := true
	for i := 0; i < jobInfo.TotalCmNum; i++ {
		hccl := ""
		if i < len(hccls) {
			hccl = hccls[i]
		}
		result = updateCM(jobInfo, i, hccl) && result
	}
	if result {
		hwlog.RunLog.Debugf("update job:%s success", jobInfo.Name)
		SaveJobCache(jobInfo.Key, jobInfo)
	}
}

// jobInfo.NodeNames store history node names used by pod
func updateUseNodeNames(jobInfo *constant.JobInfo, podsInJob map[string]v1.Pod) {
	if jobInfo.NodeNames == nil {
		jobInfo.NodeNames = make(map[string]string)
	}
	newNodeNames := make(map[string]string, len(jobInfo.NodeNames))
	for podUid, nodeName := range jobInfo.NodeNames {
		newNodeNames[podUid] = nodeName
	}
	for _, podTemp := range podsInJob {
		newNodeNames[string(podTemp.UID)] = podTemp.Spec.NodeName
	}
	jobInfo.NodeNames = newNodeNames
}

func initJobShareTorInfo(jobInfo *constant.JobInfo, podsInJob map[string]v1.Pod) {
	if jobInfo.Framework != ptFramework {
		return
	}
	if jobInfo.MasterAddr != "" || jobInfo.SharedTorIp != "" {
		return
	}
	jobInfo.SharedTorIp = pod.GetSharedTorIpByPod(podsInJob)
	if jobInfo.JobType == vcJobKind {
		if len(jobInfo.JobRankTable.ServerList) > 0 {
			jobInfo.MasterAddr = jobInfo.JobRankTable.ServerList[0].HostIp
		}
	} else {
		jobInfo.MasterAddr = pod.GetEnvByPod(podsInJob, masterAddr)
	}
}

func getHcclSlice(table constant.RankTable) []string {
	return getHcclSliceBySize(table, maxHcclSliceSize)
}

// computeRankEnd returns the RankList end index aligned to ServerList[serverBegin:serverEnd].
func computeRankEnd(table constant.RankTable, serverBegin, serverEnd, rankBegin int) int {
	rankEnd := rankBegin
	for i := serverBegin; i < serverEnd; i++ {
		rankEnd += len(table.ServerList[i].DeviceList)
	}
	if rankEnd > len(table.RankList) {
		rankEnd = len(table.RankList)
	}
	return rankEnd
}

// marshalHcclPart marshals the sub-table for ServerList[serverBegin:serverEnd] with aligned RankList.
func marshalHcclPart(table constant.RankTable, serverBegin, serverEnd, rankBegin int) (string, error) {
	part := table
	part.ServerList = table.ServerList[serverBegin:serverEnd]
	if len(table.RankList) > 0 {
		part.RankList = table.RankList[rankBegin:computeRankEnd(table, serverBegin, serverEnd, rankBegin)]
	}
	b, err := json.Marshal(part)
	return string(b), err
}

// getHcclSliceBySize slices ServerList by byte threshold (maxSize), keeping RankList aligned.
// Each slice carries the full Total (= number of slices) and ServerCount, matching prior semantics.
func getHcclSliceBySize(table constant.RankTable, maxSize int) []string {
	if len(table.ServerList) == 0 {
		return nil
	}
	// Phase 1: measure split boundaries by binary search over server count.
	type boundary struct{ serverEnd, rankEnd int }
	bounds := make([]boundary, 0)
	serverBegin, rankBegin := 0, 0
	for serverBegin < len(table.ServerList) {
		remaining := len(table.ServerList) - serverBegin
		serverCount := util.BinarySearchMaxFit(remaining, maxSize, func(k int) int {
			s, _ := marshalHcclPart(table, serverBegin, serverBegin+k, rankBegin)
			return len(s)
		})
		if serverCount < 1 {
			serverCount = 1 // single oversized server: emit as-is (fallback)
			hwlog.RunLog.Warnf("single server exceeds hccl slice size limit, serverBegin=%d", serverBegin)
		}
		serverEnd := serverBegin + serverCount
		rankEnd := computeRankEnd(table, serverBegin, serverEnd, rankBegin)
		bounds = append(bounds, boundary{serverEnd, rankEnd})
		serverBegin, rankBegin = serverEnd, rankEnd
	}
	// Phase 2: emit slices with Total = number of slices.
	total := len(bounds)
	table.Total = total
	hcclJsons := make([]string, 0, total)
	serverBegin, rankBegin = 0, 0
	for _, b := range bounds {
		str, err := marshalHcclPart(table, serverBegin, b.serverEnd, rankBegin)
		if err != nil {
			hwlog.RunLog.Errorf("Marshal hccl json part %v error, error is %v", len(hcclJsons), err)
			serverBegin, rankBegin = b.serverEnd, b.rankEnd
			continue
		}
		hcclJsons = append(hcclJsons, str)
		serverBegin, rankBegin = b.serverEnd, b.rankEnd
	}
	return hcclJsons
}

// GetJobServerInfoMap could get all job info in once query
func GetJobServerInfoMap() constant.JobServerInfoMap {
	allJobServerMap := make(map[string]map[string]constant.ServerHccl)
	allRetryJobFlag := make(map[string]bool)
	resourceType := make(map[string]string)
	for jobKey, jobInfo := range GetAllJobCache() {
		jobServerMap := buildJobServerInfoMap(jobInfo)
		allJobServerMap[jobKey] = jobServerMap
		allRetryJobFlag[jobKey] = podgroup.JudgeRetryByJobKey(jobKey)
		resourceType[jobKey] = jobInfo.ResourceType
	}

	return constant.JobServerInfoMap{InfoMap: allJobServerMap,
		RetryTolerate: allRetryJobFlag, ResourceType: resourceType}
}

func buildJobServerInfoMap(jobInfo constant.JobInfo) map[string]constant.ServerHccl {
	jobServerMap := make(map[string]constant.ServerHccl)
	for _, server := range jobInfo.PreServerList {
		copyServerHccl := constant.ServerHccl{
			DeviceList:   make([]constant.Device, 0),
			ServerID:     server.ServerID,
			HostIp:       server.HostIp,
			PodID:        server.PodID,
			PodNameSpace: server.PodNameSpace,
			ServerName:   server.ServerName,
			ServerSN:     server.ServerSN,
			PodName:      server.PodName,
			ContainerIds: server.ContainerIds,
		}
		for _, dev := range server.DeviceList {
			copyDev := constant.Device{
				DeviceID: dev.DeviceID,
				DeviceIP: dev.DeviceIP,
				RankID:   dev.RankID,
			}
			copyServerHccl.DeviceList = append(copyServerHccl.DeviceList, copyDev)
		}
		jobServerMap[server.ServerName] = copyServerHccl
	}
	return jobServerMap
}

// GetJobIsRunning get job is running
func GetJobIsRunning(jobKey string) bool {
	jobCache, _ := GetJobCache(jobKey)
	return jobCache.Status == StatusJobRunning
}

// GetJobIsExists get job is exists
func GetJobIsExists(jobKey string) bool {
	_, ok := GetJobCache(jobKey)
	return ok
}

// FlushLastUpdateTime flush lastUpdateTime
func FlushLastUpdateTime(jobKey string) {
	jobInfo, ok := GetJobCache(jobKey)
	if !ok {
		return
	}
	jobInfo.LastUpdatedCmTime = time.Now().Unix()
	SaveJobCache(jobKey, jobInfo)
}

// IsMindIeServerPod check pod is mindie server pod
func IsMindIeServerPod(podInfo v1.Pod) bool {
	return podInfo.Labels != nil && podInfo.Labels[constant.MindIeJobIdLabelKey] != "" &&
		podInfo.Labels[constant.MindIeAppTypeLabelKey] == constant.ServerAppType
}

// IsMindIeServerJob check job is mindie server job
func IsMindIeServerJob(jobInfo *constant.JobInfo) bool {
	return jobInfo != nil && jobInfo.MultiInstanceJobId != "" && jobInfo.AppType == constant.ServerAppType
}

// GetCustomFilterFaultJobAndUsedDeviceInfoMap get custom filter fault job info map and job used device info map
func GetCustomFilterFaultJobAndUsedDeviceInfoMap() (map[string]map[string]constant.JobInfo,
	map[string]map[string]sets.String) {
	jobInfoMap := make(map[string]map[string]constant.JobInfo)
	deviceInfoMap := make(map[string]map[string]sets.String)
	allJob := GetAllJobCache()
	for jobKey, jobInfo := range allJob {
		podsInJob := pod.GetPodByJobId(jobKey)
		if len(podsInJob) == 0 {
			continue
		}
		if !custom.JudgeFilterFaultAnnosByJobKey(jobKey) {
			continue
		}
		jobUsedDevices := sets.String{}
		for _, podInfo := range podsInJob {
			nodeName := podInfo.Spec.NodeName
			if _, exists := jobInfoMap[nodeName]; !exists {
				jobInfoMap[nodeName] = make(map[string]constant.JobInfo)
				deviceInfoMap[nodeName] = make(map[string]sets.String)
			}
			if _, exists := jobInfoMap[nodeName][jobKey]; !exists {
				jobInfoMap[nodeName][jobKey] = jobInfo
			}
			if realDevice, exist := podInfo.Annotations[api.PodAnnotationAscendReal]; exist && realDevice != "" {
				jobUsedDevices = jobUsedDevices.Insert(strings.Split(realDevice, constant.Comma)...)
			}
			deviceInfoMap[nodeName][jobKey] = jobUsedDevices
		}
	}
	return jobInfoMap, deviceInfoMap
}

// DeepCopyServerHcclSlice deep copy ServerHccl Slice
func DeepCopyServerHcclSlice(serverList []constant.ServerHccl) []constant.ServerHccl {
	if len(serverList) == 0 {
		return []constant.ServerHccl{}
	}
	serverListCopy := make([]constant.ServerHccl, 0, len(serverList))
	for _, server := range serverList {
		serverCopy := server
		if server.DeviceList != nil {
			serverCopy.DeviceList = make([]constant.Device, len(server.DeviceList))
			// if Device has reference type field, should not use copy
			copy(serverCopy.DeviceList, server.DeviceList)
		}
		serverListCopy = append(serverListCopy, serverCopy)
	}
	return serverListCopy
}

// DeepCopyJobInfo deep copy jobInfo
func DeepCopyJobInfo(job *constant.JobInfo) *constant.JobInfo {
	// shallow copy
	copyJob := *job
	// deep copy
	if job.NodeNames != nil {
		copyJob.NodeNames = make(map[string]string, len(job.NodeNames))
		for k, v := range job.NodeNames {
			copyJob.NodeNames[k] = v
		}
	}
	copyJob.PreServerList = DeepCopyServerHcclSlice(job.PreServerList)
	copyJob.JobRankTable.ServerList = DeepCopyServerHcclSlice(job.JobRankTable.ServerList)
	return &copyJob
}
