// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"strings"
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func newJobRankFaultInfoProcessor(deviceCenter *deviceFaultProcessCenter) *jobRankFaultInfoProcessor {
	return &jobRankFaultInfoProcessor{
		jobFaultInfoMap: make(map[string]JobFaultInfo),
		deviceCenter:    deviceCenter,
		mutex:           sync.RWMutex{},
	}
}

func (processor *jobRankFaultInfoProcessor) getJobFaultRankInfos() map[string]JobFaultInfo {
	processor.mutex.RLock()
	defer processor.mutex.RUnlock()
	result := new(map[string]JobFaultInfo)
	err := util.DeepCopy(result, processor.jobFaultInfoMap)
	if err != nil {
		hwlog.RunLog.Errorf("get job fault rank failed, err: %v", err)
		return nil
	}
	hwlog.RunLog.Debugf("get job fault rank: %v", util.ObjToString(*result))
	return *result
}

func (processor *jobRankFaultInfoProcessor) getJobFaultRankInfosFilterLevel(
	faultLevel string) map[string]JobFaultInfo {
	jobFaultRankInfos := processor.getJobFaultRankInfos()
	if jobFaultRankInfos == nil {
		return nil
	}
	for jobId, jobFaultInfo := range jobFaultRankInfos {
		faultList := make([]FaultRank, 0)
		for _, fault := range jobFaultInfo.FaultList {
			if fault.FaultLevel != faultLevel {
				faultList = append(faultList, fault)
			}
		}
		jobFaultInfo.FaultList = faultList
		jobFaultRankInfos[jobId] = jobFaultInfo
	}
	return jobFaultRankInfos
}

func (processor *jobRankFaultInfoProcessor) setJobFaultRankInfos(faultInfos map[string]JobFaultInfo) {
	processor.mutex.Lock()
	defer processor.mutex.Unlock()
	processor.jobFaultInfoMap = faultInfos
}

func (processor *jobRankFaultInfoProcessor) process() {}

func (processor *jobRankFaultInfoProcessor) findFaultRankForJob(nodeDeviceInfoMap map[string]AdvanceDeviceFaultCm,
	nodeName string, serverList map[string]constant.ServerHccl, jobId string) []FaultRank {
	advanceDeviceInfo := nodeDeviceInfoMap[nodeName]
	devicesOfJobOnNode, ok := serverList[nodeName]
	faultRankList := make([]FaultRank, 0)
	if !ok || len(devicesOfJobOnNode.DeviceList) == 0 {
		return faultRankList
	}

	for _, deviceInfo := range devicesOfJobOnNode.DeviceList {
		deviceName := advanceDeviceInfo.ServerType + "-" + deviceInfo.DeviceID
		faultList, found := advanceDeviceInfo.FaultDeviceList[deviceName]
		uceInManagementPlane := false
		if found {
			// scan management plane fault info. management plane may filter uce fault in uceProcessor
			for _, fault := range faultList {
				faultRank := FaultRank{
					RankId:      deviceInfo.RankID,
					FaultCode:   fault.FaultCode,
					FaultLevel:  fault.FaultLevel,
					DoStepRetry: false,
				}
				if strings.Contains(fault.FaultCode, constant.UceFaultCode) {
					// management plane find uce fault
					uceInManagementPlane = true
					faultRank.DoStepRetry = processor.canDoStepRetry(jobId, nodeName, deviceName)
				}
				faultRankList = append(faultRankList, faultRank)
			}
		}
		if uceInManagementPlane {
			continue
		}
		// business plane find uce fault
		if processor.uceInBusinessPlane(jobId, nodeName, deviceName) {
			faultRankList = append(faultRankList, FaultRank{
				RankId:      deviceInfo.RankID,
				FaultCode:   constant.UceFaultCode,
				FaultLevel:  RestartBusiness,
				DoStepRetry: processor.canDoStepRetry(jobId, nodeName, deviceName),
			})
		}
	}
	return faultRankList
}

func (processor *jobRankFaultInfoProcessor) canDoStepRetry(jobId, nodeName, deviceName string) bool {
	uceProcessor, err := processor.deviceCenter.getUceFaultProcessor()
	if err != nil {
		hwlog.RunLog.Errorf("getUceFaultProcessor exception: %v", err)
		return false
	}
	uceDevice, found := uceProcessor.getUceDeviceFromJob(jobId, nodeName, deviceName)
	if !found {
		hwlog.RunLog.Debugf("job %s's uce fault is not on node %s device %s", jobId, nodeName, deviceName)
		return false
	}
	doStepRetry := canDoStepRetry(&uceDevice)
	hwlog.RunLog.Debugf("uceDevice %s stepretry %v", util.ObjToString(uceDevice), doStepRetry)
	return doStepRetry
}

func (processor *jobRankFaultInfoProcessor) uceInBusinessPlane(jobId, nodeName, deviceName string) bool {
	uceProcessor, err := processor.deviceCenter.getUceFaultProcessor()
	if err != nil {
		hwlog.RunLog.Errorf("getUceFaultProcessor exception: %v", err)
		return false
	}
	uceDevice, found := uceProcessor.getUceDeviceFromJob(jobId, nodeName, deviceName)
	// business plane didn't find uce fault
	if !found {
		hwlog.RunLog.Debugf("business plane didn't find uce fault")
		return false
	}
	// business plane found uce fault
	return validBusinessRecoverTime(uceDevice.RecoverTime)
}
