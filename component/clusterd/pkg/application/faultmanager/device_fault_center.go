// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"fmt"
	"sync"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/job"
)

func newDeviceFaultProcessCenter() *deviceFaultProcessCenter {
	manager := faultCenterCmManager[*constant.DeviceInfo]{
		mutex:        sync.RWMutex{},
		originalCm:   configMap[*constant.DeviceInfo]{configmap: make(map[string]*constant.DeviceInfo)},
		processingCm: configMap[*constant.DeviceInfo]{configmap: make(map[string]*constant.DeviceInfo)},
		processedCm:  configMap[*constant.DeviceInfo]{configmap: make(map[string]*constant.DeviceInfo)},
	}
	deviceCenter := &deviceFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(&manager),
	}

	var processorForUceAccompanyFault = newUceAccompanyFaultProcessor(deviceCenter)
	var processorForUceFault = newUceFaultProcessor(deviceCenter)
	var processForJobFaultRank = newJobRankFaultInfoProcessor(deviceCenter)
	var processLinkDownCqeFault = newLinkDownCqeFaultProcessor(deviceCenter)

	deviceCenter.addProcessors([]faultProcessor{
		processorForUceAccompanyFault, // this processor filter the uce accompany faults, before processorForUceFault
		processorForUceFault,          // this processor filter the uce faults.
		processLinkDownCqeFault,       // this processor filter the cqe, link down faults.
		processForJobFaultRank,        // this processor need to get filtered faults
	})
	return deviceCenter
}

func (deviceCenter *deviceFaultProcessCenter) getUceFaultProcessor() (*uceFaultProcessor, error) {
	for _, processor := range deviceCenter.processorList {
		if processor, ok := processor.(*uceFaultProcessor); ok {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("can not find uceFaultProcessor in FaultProcessCenter")
}

func (deviceCenter *deviceFaultProcessCenter) getUceAccompanyFaultProcessor() (*uceAccompanyFaultProcessor, error) {
	for _, processor := range deviceCenter.processorList {
		if processor, ok := processor.(*uceAccompanyFaultProcessor); ok {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("can not find uceAccompanyFaultProcessor in FaultProcessCenter")
}

func (deviceCenter *deviceFaultProcessCenter) getJobFaultRankProcessor() (*jobRankFaultInfoProcessor, error) {
	for _, processor := range deviceCenter.processorList {
		if processor, ok := processor.(*jobRankFaultInfoProcessor); ok {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("can not find jobRankFaultInfoProcessor in FaultProcessCenter")
}

func (deviceCenter *deviceFaultProcessCenter) callbackForReportUceInfo(jobId, rankId string, recoverTime int64) error {
	processor, err := deviceCenter.getUceFaultProcessor()
	if err != nil {
		hwlog.RunLog.Error(err)
		return err
	}
	return processor.reportUceInfo(jobId, rankId, recoverTime)
}

func (deviceCenter *deviceFaultProcessCenter) process() {
	currentTime := time.Now().UnixMilli()
	if deviceCenter.isProcessLimited(currentTime) {
		return
	}
	deviceCenter.lastProcessTime = currentTime
	deviceCenter.setProcessingCm(deviceCenter.getOriginalCm())
	deviceCenter.jobServerInfoMap = job.GetJobServerInfoMap()
	hwlog.RunLog.Debugf("job server info map: %v", util.ObjToString(deviceCenter.jobServerInfoMap))
	deviceCenter.baseFaultCenter.process()
	deviceCenter.setProcessedCm(deviceCenter.getProcessingCm())
}
