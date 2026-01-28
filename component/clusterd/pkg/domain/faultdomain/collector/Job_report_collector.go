// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package collector collect information to process
package collector

import (
	"fmt"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/job"
)

var ReportInfoCollector *JobReportInfoCollector

// dataExpireTime report time expire time
const dataExpireTime = 60 * 1000

// JobReportInfoCollector job report info collector
type JobReportInfoCollector struct {
	// JobId->node->device->report_info
	RetryMap map[string]map[string]map[string]constant.ReportInfo
	// JobId->podRank->reportFaultTime
	SingleProcessMap map[string]map[string]int64
	RwMutex          sync.RWMutex
}

func init() {
	ReportInfoCollector = &JobReportInfoCollector{
		RetryMap:         make(map[string]map[string]map[string]constant.ReportInfo),
		SingleProcessMap: make(map[string]map[string]int64),
		RwMutex:          sync.RWMutex{},
	}
}

func (reportInfos *JobReportInfoCollector) GetInfo(jobId, nodeName, deviceName string) constant.ReportInfo {
	noReport := constant.ReportInfo{
		RecoverTime:  constant.JobNotRecover,
		CompleteTime: constant.JobNotRecoverComplete,
	}
	if reportInfos == nil {
		return noReport
	}
	reportInfos.RwMutex.RLock()
	defer reportInfos.RwMutex.RUnlock()
	info, ok := reportInfos.RetryMap[jobId][nodeName][deviceName]
	if !ok {
		return noReport
	}
	currentTime := time.Now().UnixMilli()
	if currentTime-info.RecoverTime > dataExpireTime {
		reportInfos.deleteInfo(jobId, nodeName, deviceName)
		hwlog.RunLog.Infof("delete expired report info(%s, %s, %s), info: %v", jobId, nodeName, deviceName, info)
		return noReport
	}
	return info
}

func (reportInfos *JobReportInfoCollector) deleteInfo(jobId string, nodeName string, deviceName string) {
	delete(reportInfos.RetryMap[jobId][nodeName], deviceName)
}

// GetSingleProcessFaultReportTime get single process fault report time
func (reportInfos *JobReportInfoCollector) GetSingleProcessFaultReportTime(jobId, podRankStr string) int64 {
	reportTime := constant.JobShouldReportFault
	reportInfos.RwMutex.RLock()
	defer reportInfos.RwMutex.RUnlock()
	if podRankTime, ok := reportInfos.SingleProcessMap[jobId]; ok {
		if t, ok1 := podRankTime[podRankStr]; ok1 {
			return t
		}
	}
	return reportTime
}

func (reportInfos *JobReportInfoCollector) GetInfoWithoutJobId(nodeName, deviceName string) constant.ReportInfo {
	noReport := constant.ReportInfo{
		RecoverTime:  constant.JobNotRecover,
		CompleteTime: constant.JobNotRecoverComplete,
	}
	if reportInfos == nil {
		return noReport
	}
	reportInfos.RwMutex.RLock()
	defer reportInfos.RwMutex.RUnlock()
	for _, infoMapValue := range reportInfos.RetryMap {
		if infoMapValue == nil {
			continue
		}
		if info, ok := infoMapValue[nodeName][deviceName]; ok {
			return info
		}
	}
	return noReport
}

// ReportRetryInfo report retry info
func (reportInfos *JobReportInfoCollector) ReportRetryInfo(jobId string, rankId string,
	recoverTime int64, faultType string) error {
	jobServerInfoMap := job.GetJobServerInfoMap()
	nodeName, deviceId, err := faultdomain.GetNodeAndDeviceFromJobIdAndRankId(jobId, rankId, jobServerInfoMap)
	if err != nil {
		err = fmt.Errorf("report info failed, exception: %v", err)
		hwlog.RunLog.Error(err)
		return err
	}
	deviceName := jobServerInfoMap.ResourceType[jobId] + "-" + deviceId
	reportInfos.RwMutex.Lock()
	defer reportInfos.RwMutex.Unlock()
	infoMap := reportInfos.RetryMap
	info := constant.ReportInfo{
		RecoverTime:  recoverTime,
		CompleteTime: constant.JobNotRecoverComplete,
		FaultType:    faultType,
	}
	if infoMap == nil {
		infoMap = make(map[string]map[string]map[string]constant.ReportInfo)
	}
	if _, ok := infoMap[jobId]; !ok {
		infoMap[jobId] = make(map[string]map[string]constant.ReportInfo)
		if _, ok := infoMap[jobId][nodeName]; !ok {
			infoMap[jobId][nodeName] = make(map[string]constant.ReportInfo)
		}
		infoMap[jobId][nodeName][deviceName] = info
	} else {
		if _, ok := infoMap[jobId][nodeName]; !ok {
			infoMap[jobId][nodeName] = make(map[string]constant.ReportInfo)
		}
		infoMap[jobId][nodeName][deviceName] = info
	}
	reportInfos.RetryMap = infoMap
	hwlog.RunLog.Infof("callbackForReportRetryInfo receive report info(%s, %s, %d)", jobId, rankId, recoverTime)
	hwlog.RunLog.Debugf("Current retry reportInfo is %s", util.ObjToString(reportInfos.RetryMap))
	return nil
}

// ReportNoRetryInfo report no retry fault info
func (reportInfos *JobReportInfoCollector) ReportNoRetryInfo(jobId string, exitPodRanks []string, reportFaultTime int64) {
	reportInfos.RwMutex.Lock()
	defer reportInfos.RwMutex.Unlock()
	noRetryMap := reportInfos.SingleProcessMap
	if noRetryMap == nil {
		noRetryMap = make(map[string]map[string]int64)
	}
	if _, ok := noRetryMap[jobId]; !ok {
		noRetryMap[jobId] = make(map[string]int64)
	}
	for _, exitPodRank := range exitPodRanks {
		noRetryMap[jobId][exitPodRank] = reportFaultTime
	}
	reportInfos.SingleProcessMap = noRetryMap
	hwlog.RunLog.Infof("callbackForReportNoRetryInfo receive report info(%s, %v, %d)", jobId, exitPodRanks,
		reportFaultTime)
	hwlog.RunLog.Debugf("Current no retry reportInfo is %s", util.ObjToString(reportInfos.SingleProcessMap))
}
