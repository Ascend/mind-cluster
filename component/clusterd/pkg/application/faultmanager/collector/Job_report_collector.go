// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package collector contain fault process
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

// JobId->node->device->report_info
type JobReportInfoCollector struct {
	InfoMap map[string]map[string]map[string]constant.ReportInfo
	RwMutex sync.RWMutex
}

func InitReportInfoCollector() *JobReportInfoCollector {
	ReportInfoCollector = &JobReportInfoCollector{
		InfoMap: make(map[string]map[string]map[string]constant.ReportInfo),
		RwMutex: sync.RWMutex{},
	}
	return ReportInfoCollector
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
	if info, ok := reportInfos.InfoMap[jobId][nodeName][deviceName]; ok {
		return info
	}
	return noReport
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
	for _, infoMapValue := range reportInfos.InfoMap {
		if infoMapValue == nil {
			continue
		}
		if info, ok := infoMapValue[nodeName][deviceName]; ok {
			return info
		}
	}
	return noReport
}

func (reportInfos *JobReportInfoCollector) ReportUceInfo(jobId string, rankId string, recoverTime int64) error {
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
	infoMap := reportInfos.InfoMap
	info := constant.ReportInfo{
		RecoverTime:  recoverTime,
		CompleteTime: constant.JobNotRecoverComplete,
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
	reportInfos.InfoMap = infoMap
	hwlog.RunLog.Infof("callbackForReportUceInfo receive report info(%s, %s, %d)", jobId, rankId, recoverTime)
	hwlog.RunLog.Debugf("Current reportInfo is %s", util.ObjToString(reportInfos.InfoMap))
	return nil
}
