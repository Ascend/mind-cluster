package fault

import (
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/interface/kube"
	"fmt"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"sync"
)

// DeviceFaultProcessCenter
type DeviceFaultProcessCenter struct {
	BaseFaultCenter
	mutex sync.RWMutex
	infos map[string]*constant.DeviceInfo
}

func NewDeviceFaultProcessCenter() *DeviceFaultProcessCenter {
	deviceCenter := &DeviceFaultProcessCenter{
		mutex:           sync.RWMutex{},
		infos:           make(map[string]*constant.DeviceInfo),
		BaseFaultCenter: newBaseFaultCenter(),
	}

	var processorForUceAccompanyFault = &UceAccompanyFaultProcessor{
		DiagnosisAccompanyTimeout: constant.DiagnosisAccompanyTimeout,
		uceAccompanyFaultQue:      make(map[string]map[string][]constant.DeviceFault),
		uceFaultTime:              make(map[string]map[string]int64),
		deviceCenter:              deviceCenter,
	}
	var processorForUceFault = &UceFaultProcessor{
		JobReportRecoverTimeout:  constant.JobReportRecoverTimeout,
		JobReportCompleteTimeout: constant.JobReportCompleteTimeout,
		mindIoReportInfo: &mindIoReportInfosForAllJobs{
			Infos:   make(map[string]map[string]map[string]mindIoReportInfo),
			RwMutex: sync.RWMutex{},
		},
		deviceCenter: deviceCenter,
	}
	var processForJobFaultRank = &JobRankFaultInfoProcessor{
		jobFaultInfos: make(map[string]FaultInfo),
		deviceCenter:  deviceCenter,
	}

	deviceCenter.addProcessors([]FaultProcessor{
		processorForUceAccompanyFault,
		processorForUceFault,
		processForJobFaultRank})
	return deviceCenter
}

func (deviceCenter *DeviceFaultProcessCenter) GetDeviceInfos() map[string]*constant.DeviceInfo {
	deviceCenter.mutex.RLock()
	defer deviceCenter.mutex.RUnlock()
	return device.DeepCopyInfos(deviceCenter.infos)
}

func (deviceCenter *DeviceFaultProcessCenter) SetDeviceInfos(infos map[string]*constant.DeviceInfo) {
	deviceCenter.mutex.Lock()
	defer deviceCenter.mutex.Unlock()
	deviceCenter.infos = device.DeepCopyInfos(infos)
}

func (deviceCenter *DeviceFaultProcessCenter) InformerAddCallback(oldInfo, newInfo *constant.DeviceInfo) bool {
	deviceCenter.mutex.Lock()
	defer deviceCenter.mutex.Unlock()
	length := len(deviceCenter.infos)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("SwitchInfo length=%d > %d, SwitchInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
		return false
	}
	oldInfo, found := deviceCenter.infos[newInfo.CmName]
	deviceCenter.infos[newInfo.CmName] = newInfo
	return found && device.BusinessDataIsNotEqual(oldInfo, newInfo)
}

func (deviceCenter *DeviceFaultProcessCenter) InformerDelCallback(oldInfo, newInfo *constant.DeviceInfo) bool {
	deviceCenter.mutex.Lock()
	defer deviceCenter.mutex.Unlock()
	oldInfo, found := deviceCenter.infos[newInfo.CmName]
	delete(deviceCenter.infos, newInfo.CmName)
	deviceCenter.infos[newInfo.CmName] = newInfo
	return found
}

func (deviceCenter *DeviceFaultProcessCenter) GetUceFaultProcessor() (*UceFaultProcessor, error) {
	for _, processor := range deviceCenter.processors {
		if processor, ok := processor.(*UceFaultProcessor); ok {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("can not find uceFaultProcessor in FaultProcessCenter")
}

func (deviceCenter *DeviceFaultProcessCenter) GetUceAccompanyFaultProcessor() (*UceAccompanyFaultProcessor, error) {
	for _, processor := range deviceCenter.processors {
		if processor, ok := processor.(*UceAccompanyFaultProcessor); ok {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("can not find uceAccompanyFaultProcessor in FaultProcessCenter")
}

func (deviceCenter *DeviceFaultProcessCenter) GetJobFaultRankProcessor() (*JobRankFaultInfoProcessor, error) {
	for _, processor := range deviceCenter.processors {
		if processor, ok := processor.(*JobRankFaultInfoProcessor); ok {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("can not find jobRankFaultInfoProcessor in FaultProcessCenter")
}

func (deviceCenter *DeviceFaultProcessCenter) CallbackForReportUceInfo(jobUid, rankId string, recoverTime int64) error {
	processor, err := deviceCenter.GetUceFaultProcessor()
	if err != nil {
		hwlog.RunLog.Error(err)
		return err
	}
	nodeName, deviceId, err := kube.JobMgr.GetNodeAndDeviceFromJobIdAndRankId(jobUid, rankId)
	if err != nil {
		err = fmt.Errorf("mindIO report info failed, exception: %v", err)
		hwlog.RunLog.Error(err)
		return err
	}
	deviceName := deviceID2DeviceKey(deviceId)
	processor.mindIoReportInfo.RwMutex.Lock()
	defer processor.mindIoReportInfo.RwMutex.Unlock()
	reportInfo := processor.mindIoReportInfo.Infos
	info := mindIoReportInfo{
		RecoverTime:  recoverTime,
		CompleteTime: constant.JobNotRecoverComplete,
	}
	if reportInfo == nil {
		reportInfo = make(map[string]map[string]map[string]mindIoReportInfo)
	}
	if _, ok := reportInfo[jobUid]; !ok {
		reportInfo[jobUid] = make(map[string]map[string]mindIoReportInfo)
		if _, ok := reportInfo[jobUid][nodeName]; !ok {
			reportInfo[jobUid][nodeName] = make(map[string]mindIoReportInfo)
		}
		reportInfo[jobUid][nodeName][deviceName] = info
	} else {
		if _, ok := reportInfo[jobUid][nodeName]; !ok {
			reportInfo[jobUid][nodeName] = make(map[string]mindIoReportInfo)
		}
		reportInfo[jobUid][nodeName][deviceName] = info
	}
	processor.mindIoReportInfo.Infos = reportInfo
	return nil
}
