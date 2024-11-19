// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultshoot contain fault process
package faultshoot

import (
	"fmt"
	"sync"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/interface/kube"
)

func newDeviceFaultProcessCenter() *deviceFaultProcessCenter {
	deviceCenter := &deviceFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(),
		mutex:           sync.RWMutex{},
		devicePluginCm:  make(map[string]*constant.DeviceInfo),
		processingCm:    make(map[string]*constant.DeviceInfo),
	}

	var processorForUceAccompanyFault = newUceAccompanyFaultProcessor(deviceCenter)
	var processorForUceFault = newUceFaultProcessor(deviceCenter)
	var processForJobFaultRank = newJobRankFaultInfoProcessor(deviceCenter)

	deviceCenter.addProcessors([]faultProcessor{
		processForJobFaultRank,        // this processor don't need to filter anything, so assign on the first position.
		processorForUceAccompanyFault, // this processor filter the uce accompany faults, should before processorForUceFault
		processorForUceFault,          // this processor filter the uce faults.
	})
	return deviceCenter
}

func (deviceCenter *deviceFaultProcessCenter) getProcessingCm() map[string]*constant.DeviceInfo {
	deviceCenter.mutex.RLock()
	defer deviceCenter.mutex.RUnlock()
	return device.DeepCopyInfos(deviceCenter.processingCm)
}

func (deviceCenter *deviceFaultProcessCenter) setProcessingCm(infos map[string]*constant.DeviceInfo) {
	deviceCenter.mutex.Lock()
	defer deviceCenter.mutex.Unlock()
	deviceCenter.processingCm = device.DeepCopyInfos(infos)
}

func (deviceCenter *deviceFaultProcessCenter) getProcessedCm() map[string]*constant.DeviceInfo {
	deviceCenter.mutex.RLock()
	defer deviceCenter.mutex.RUnlock()
	return device.DeepCopyInfos(deviceCenter.processedCm)
}

func (deviceCenter *deviceFaultProcessCenter) setProcessedCm(infos map[string]*constant.DeviceInfo) {
	deviceCenter.mutex.Lock()
	defer deviceCenter.mutex.Unlock()
	deviceCenter.processedCm = device.DeepCopyInfos(infos)
}

func (deviceCenter *deviceFaultProcessCenter) updateDevicePluginCm(newInfo *constant.DeviceInfo) {
	deviceCenter.mutex.Lock()
	defer deviceCenter.mutex.Unlock()
	length := len(deviceCenter.devicePluginCm)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("DeviceInfo length=%d > %d, SwitchInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
		return
	}
	deviceCenter.devicePluginCm[newInfo.CmName] = newInfo
}

func (deviceCenter *deviceFaultProcessCenter) delDevicePluginCm(newInfo *constant.DeviceInfo) {
	deviceCenter.mutex.Lock()
	defer deviceCenter.mutex.Unlock()
	delete(deviceCenter.devicePluginCm, newInfo.CmName)
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
	deviceCenter.setProcessingCm(deviceCenter.devicePluginCm)
	deviceCenter.jobServerInfoMap = kube.JobMgr.GetJobServerInfoMap()
	hwlog.RunLog.Debugf("job server info map: %v", util.ObjToString(deviceCenter.jobServerInfoMap))
	deviceCenter.baseFaultCenter.process()
	deviceCenter.setProcessedCm(deviceCenter.processingCm)
}
