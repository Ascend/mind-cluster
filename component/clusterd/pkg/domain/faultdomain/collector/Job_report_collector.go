// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package collector collect information to process
package collector

import (
	"fmt"
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/job"
)

var ReportInfoCollector *JobReportInfoCollector

// JobReportInfoCollector job report info collector
type JobReportInfoCollector struct {
	// JobId->node->device->report_info
	RetryMap map[string]map[string]map[string]constant.ReportInfo
	// JobId->reportFaultTime
	NoRetryMap map[string]int64
	RwMutex    sync.RWMutex
}

func init() {
	ReportInfoCollector = &JobReportInfoCollector{
		RetryMap:   make(map[string]map[string]map[string]constant.ReportInfo),
		NoRetryMap: make(map[string]int64),
		RwMutex:    sync.RWMutex{},
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
	if info, ok := reportInfos.RetryMap[jobId][nodeName][deviceName]; ok {
		return info
	}
	return noReport
}

// GetNoRetryReportTime get no retry report time
func (reportInfos *JobReportInfoCollector) GetNoRetryReportTime(jobId string) int64 {
	reportTime := constant.JobShouldReportFault
	reportInfos.RwMutex.RLock()
	defer reportInfos.RwMutex.RUnlock()
	if time, ok := reportInfos.NoRetryMap[jobId]; ok {
		return time
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
func (reportInfos *JobReportInfoCollector) ReportNoRetryInfo(jobId string, reportFaultTime int64) {
	reportInfos.RwMutex.Lock()
	defer reportInfos.RwMutex.Unlock()
	noRetryMap := reportInfos.NoRetryMap
	if noRetryMap == nil {
		noRetryMap = make(map[string]int64)
	}
	noRetryMap[jobId] = reportFaultTime
	reportInfos.NoRetryMap = noRetryMap
	hwlog.RunLog.Infof("callbackForReportNoRetryInfo receive report info(%s, %d)", jobId, reportFaultTime)
	hwlog.RunLog.Debugf("Current no retry reportInfo is %s", util.ObjToString(reportInfos.NoRetryMap))
}
