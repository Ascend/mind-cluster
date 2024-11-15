package fault

import (
	"fmt"
	"sync"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/device"
)

// deviceFaultProcessCenter
type deviceFaultProcessCenter struct {
	baseFaultCenter
	mutex   sync.RWMutex
	infoMap map[string]*constant.DeviceInfo
}

func newDeviceFaultProcessCenter() *deviceFaultProcessCenter {
	deviceCenter := &deviceFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(),
		mutex:           sync.RWMutex{},
		infoMap:         make(map[string]*constant.DeviceInfo),
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

// AdvanceDeviceCm more structure device info
type AdvanceDeviceCm struct {
	ServerType       string
	CmName           string
	SuperPodID       int32
	ServerIndex      int32
	DeviceList       map[string][]constant.DeviceFault
	CarUnHealthy     []string
	NetworkUnhealthy []string
	UpdateTime       int64
}

func (deviceCenter *deviceFaultProcessCenter) getInfoMap() map[string]*constant.DeviceInfo {
	deviceCenter.mutex.RLock()
	defer deviceCenter.mutex.RUnlock()
	return device.DeepCopyInfos(deviceCenter.infoMap)
}

func (deviceCenter *deviceFaultProcessCenter) setInfoMap(infos map[string]*constant.DeviceInfo) {
	deviceCenter.mutex.Lock()
	defer deviceCenter.mutex.Unlock()
	deviceCenter.infoMap = device.DeepCopyInfos(infos)
}

func (deviceCenter *deviceFaultProcessCenter) updateInfoFromCm(newInfo *constant.DeviceInfo) {
	deviceCenter.mutex.Lock()
	defer deviceCenter.mutex.Unlock()
	length := len(deviceCenter.infoMap)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("DeviceInfo length=%d > %d, SwitchInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
	}
	deviceCenter.infoMap[newInfo.CmName] = newInfo
}

func (deviceCenter *deviceFaultProcessCenter) delInfoFromCm(newInfo *constant.DeviceInfo) {
	deviceCenter.mutex.Lock()
	defer deviceCenter.mutex.Unlock()
	delete(deviceCenter.infoMap, newInfo.CmName)
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
