// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job cache function
package job

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/superpod"
)

const (
	preDeleteToDeleteSecond = 60
	updateSecond            = 3600
)

var jobSummaryMap sync.Map
var existingSids sync.Map

// GetJobCache get job cache info
// Do not modify the reference type field of the return value.
// If you want to modify it, please call GetJobCacheDeepCopy
func GetJobCache(jobKey string) (constant.JobInfo, bool) {
	jobInfo, ok := jobSummaryMap.Load(jobKey)
	if !ok {
		return constant.JobInfo{}, ok
	}
	return jobInfo.(constant.JobInfo), ok
}

// GetJobCacheDeepCopy get deep copy job cache info
func GetJobCacheDeepCopy(jobKey string) (constant.JobInfo, bool) {
	jobInfo, ok := GetJobCache(jobKey)
	if !ok {
		return constant.JobInfo{}, ok
	}
	return *DeepCopyJobInfo(&jobInfo), ok
}

// GetAllJobCache get all job cache info
func GetAllJobCache() map[string]constant.JobInfo {
	allJob := map[string]constant.JobInfo{}
	jobSummaryMap.Range(func(key, value any) bool {
		jobKey, ok := key.(string)
		if !ok {
			return true
		}
		jobInfo, ok := value.(constant.JobInfo)
		if !ok {
			return true
		}
		allJob[jobKey] = jobInfo
		return true
	})
	hwlog.RunLog.Debugf("get all job cache, allJob: %v", allJob)
	return allJob
}

// SaveJobCache save job cache info
func SaveJobCache(jobKey string, jobInfo constant.JobInfo) {
	if jobInfo.Sid != "" {
		existingJobKey, exists := existingSids.Load(jobInfo.Sid)
		if exists && existingJobKey.(string) != jobKey {
			hwlog.RunLog.Warnf("Job sid conflict, jobId:%s, sid:%s", jobKey, jobInfo.Sid)
		}
	}

	jobSummaryMap.Store(jobKey, jobInfo)
	if jobInfo.Sid != "" {
		existingSids.Store(jobInfo.Sid, jobKey)
	}
}

// DeleteJobCache delete job cache info
func DeleteJobCache(jobKey string) {
	if jobInfo, ok := GetJobCache(jobKey); ok && jobInfo.Sid != "" {
		existingSids.Delete(jobInfo.Sid)
	}
	hwlog.RunLog.Infof("delete job cache, jobKey: %v", jobKey)
	jobSummaryMap.Delete(jobKey)
}

func findJobByFilter(filter func(constant.JobInfo) bool) constant.JobInfo {
	ji := constant.JobInfo{}
	jobSummaryMap.Range(func(_, value any) bool {
		jobInfo, ok := value.(constant.JobInfo)
		if !ok {
			return true
		}
		if filter(jobInfo) {
			ji = jobInfo
			return false
		}
		return true
	})
	return ji
}

// GetJobByNameSpaceAndName get job by job name and nameSpace
func GetJobByNameSpaceAndName(name, nameSpace string) constant.JobInfo {
	return findJobByFilter(func(ji constant.JobInfo) bool {
		return ji.Name == name && ji.NameSpace == nameSpace
	})
}

// GetRunningJobByNameSpaceAndName get running job by job name and nameSpace
func GetRunningJobByNameSpaceAndName(name, nameSpace string) constant.JobInfo {
	return findJobByFilter(func(ji constant.JobInfo) bool {
		return ji.Name == name && ji.NameSpace == nameSpace && ji.Status == StatusJobRunning
	})
}

// GetJobByNameSpaceAndNameAndPreDelete get job by job name and nameSpace
func GetJobByNameSpaceAndNameAndPreDelete(name, nameSpace string, isPreDelete bool) []constant.JobInfo {
	jobInfos := make([]constant.JobInfo, 0)
	jobSummaryMap.Range(func(_, value any) bool {
		jobInfo, ok := value.(constant.JobInfo)
		if !ok {
			return true
		}
		if jobInfo.Name == name && jobInfo.NameSpace == nameSpace && jobInfo.IsPreDelete == isPreDelete {
			jobInfos = append(jobInfos, jobInfo)
		}
		return true
	})
	return jobInfos
}

// GetShouldDeleteJobKey get should delete job key
func GetShouldDeleteJobKey() []string {
	allJob := make([]string, 0)
	nowTime := time.Now().Unix()
	jobSummaryMap.Range(func(key, value any) bool {
		jobKey, ok := key.(string)
		if !ok {
			return true
		}
		jobInfo, ok := value.(constant.JobInfo)
		if !ok {
			return true
		}
		if jobInfo.IsPreDelete && nowTime-jobInfo.DeleteTime >= preDeleteToDeleteSecond {
			allJob = append(allJob, jobKey)
		}
		return true
	})
	return allJob
}

// GetShouldUpdateJobKey  get should update job key
func GetShouldUpdateJobKey() []string {
	allJobKey := make([]string, 0)
	nowTime := time.Now().Unix()
	jobSummaryMap.Range(func(key, value any) bool {
		jobKey, ok := key.(string)
		if !ok {
			return true
		}
		jobInfo, ok := value.(constant.JobInfo)
		if !ok {
			return true
		}
		if nowTime-jobInfo.LastUpdatedCmTime >= updateSecond {
			allJobKey = append(allJobKey, jobKey)
		}
		return true
	})
	return allJobKey
}

// GetNamespaceByJobIdAndAppType get namespace by jobId and appType
func GetNamespaceByJobIdAndAppType(jobId, appType string) (string, error) {
	namespace := ""
	jobSummaryMap.Range(func(_, value any) bool {
		jobInfo, ok := value.(constant.JobInfo)
		if !ok {
			return true
		}
		if jobInfo.MultiInstanceJobId == jobId && jobInfo.AppType == appType {
			namespace = jobInfo.NameSpace
			return false
		}
		return true
	})
	if namespace == "" {
		return "", fmt.Errorf("no jobId (%s) server job", jobId)
	}
	return namespace, nil
}

// GetInstanceJobKey retrieve the jobKey containing jobID=jobId and app=appType in the label under nameSpace
func GetInstanceJobKey(jobId, namespace, appType string) (string, error) {
	jobKey := ""
	jobSummaryMap.Range(func(key, value any) bool {
		jobInfo, ok := value.(constant.JobInfo)
		if !ok {
			return true
		}
		if jobInfo.MultiInstanceJobId == jobId && jobInfo.NameSpace == namespace && jobInfo.AppType == appType {
			jobKey, ok = key.(string)
			if !ok {
				return true
			}
			return false
		}
		return true
	})
	if jobKey == "" {
		return "", fmt.Errorf("no %s job found", appType)
	}
	return jobKey, nil
}

// GetJobFaultSdIdAndNodeName get job sdid and node name
func GetJobFaultSdIdAndNodeName(jobId string, faultPods map[string]string) map[int]api.SuperPodFaultInfos {
	jobInfo, ok := GetJobCache(jobId)
	if !ok {
		return nil
	}
	faultNodes := make(sets.String)
	for _, podUid := range faultPods {
		nodeName, ok := jobInfo.NodeNames[podUid]
		if !ok {
			hwlog.RunLog.Warnf("no node name for pod %s", podUid)
			continue
		}
		faultNodes.Insert(nodeName)
	}
	hwlog.RunLog.Infof("force release %v fault nodes: %v", jobId, faultNodes)
	superNodes := make(map[int][]string)
	for _, serverInfo := range jobInfo.PreServerList {
		if serverInfo.SuperPodId < 0 {
			hwlog.RunLog.Debugf("force release %s superPodId %v < 0", serverInfo.ServerName, serverInfo.SuperPodId)
			return nil
		}
		if !faultNodes.Has(serverInfo.ServerName) {
			superNodes[serverInfo.SuperPodId] = append(superNodes[serverInfo.SuperPodId], serverInfo.ServerName)
			hwlog.RunLog.Infof("force release serverName: %v, superPodId: %v", serverInfo.ServerName,
				serverInfo.SuperPodId)
			continue
		}
	}
	faultSuperID := getFaultSuperID(faultNodes)
	if len(superNodes) == 0 || len(faultSuperID) == 0 {
		hwlog.RunLog.Warnf("force release no fault superPodId")
		return nil
	}
	faultInfo := make(map[int]api.SuperPodFaultInfos)
	for spId := range faultSuperID {
		faultInfo[spId] = api.SuperPodFaultInfos{
			NodeNames: superNodes[spId], SdIds: faultSuperID[spId], FaultNodes: faultNodes,
			FaultTimes: time.Now().Unix(), JobId: jobId}
	}
	return faultInfo
}

func getFaultSuperID(faultNodes sets.String) map[int][]string {
	faultSuperID := make(map[int][]string)
	for _, nodes := range superpod.ListClusterDevice() {
		superPodID, err := strconv.Atoi(nodes.SuperPodID)
		if err != nil {
			hwlog.RunLog.Errorf("get superPodID failed, err: %v", err)
			continue
		}
		for _, node := range nodes.NodeDeviceMap {
			if !faultNodes.Has(node.NodeName) {
				continue
			}
			for _, sdid := range node.DeviceMap {
				faultSuperID[superPodID] = append(faultSuperID[superPodID], sdid)
			}
		}
	}
	return faultSuperID
}
