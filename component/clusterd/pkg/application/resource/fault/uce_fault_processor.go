package fault

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/interface/kube"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"sync"
	"time"
)

/*
The UceFaultProcessor process uce fault reporting information.
If the device fault is UCE fault, then determine whether the job running on the device can tolerate UCE faults.
If they can tolerate it, the reporting of the UCE fault should be delayed by 10 seconds.
*/
type UceFaultProcessor struct {
	deviceCenter             *DeviceFaultProcessCenter
	JobReportRecoverTimeout  int64
	JobReportCompleteTimeout int64

	mindIoReportInfo *mindIoReportInfosForAllJobs
	// uceJob->jobInfo
	uceDevicesOfUceJob map[string]uceJobInfo
	// node->DeviceName->uceDeviceInfo
	uceDeviceOfNode map[string]uceNodeInfo
}

// JobId->node->device->report_info
type mindIoReportInfosForAllJobs struct {
	Infos   map[string]map[string]map[string]mindIoReportInfo
	RwMutex sync.RWMutex
}

func (reportInfos *mindIoReportInfosForAllJobs) getInfo(jobId, nodeName, deviceName string) mindIoReportInfo {
	if reportInfos == nil {
		return mindIoReportInfo{
			RecoverTime:  constant.JobNotRecover,
			CompleteTime: constant.JobNotRecoverComplete,
		}
	}
	reportInfos.RwMutex.RLock()
	defer reportInfos.RwMutex.RUnlock()
	if info, ok := reportInfos.Infos[jobId][nodeName][deviceName]; ok {
		return info
	}
	return mindIoReportInfo{
		RecoverTime:  constant.JobNotRecover,
		CompleteTime: constant.JobNotRecoverComplete,
	}
}

type uceDeviceInfo struct {
	// DeviceName has prefix Ascend910
	DeviceName   string
	FaultTime    int64
	RecoverTime  int64
	CompleteTime int64
}

type uceNodeInfo struct {
	NodeName string
	// DeviceName->DeviceInfo
	DeviceInfo map[string]uceDeviceInfo
}

type uceJobInfo struct {
	// UceNode node->nodeInfo
	JobId   string
	UceNode map[string]uceNodeInfo
}

type mindIoReportInfo struct {
	RecoverTime  int64
	CompleteTime int64
}

func (processor *UceFaultProcessor) initUceDeviceFromNodeAndMindIo(jobId string,
	uceNode uceNodeInfo, serverList []*job.ServerHccl) uceNodeInfo {
	devicesOfJobOnNode := getDevicesNameOfJobOnNode(uceNode.NodeName, serverList, jobId)

	jobUceNodeInfo := uceNodeInfo{
		NodeName:   uceNode.NodeName,
		DeviceInfo: make(map[string]uceDeviceInfo),
	}

	for _, deviceOfJob := range devicesOfJobOnNode {
		deviceName := deviceID2DeviceKey(deviceOfJob.DeviceID)
		if uceDevice, ok := uceNode.DeviceInfo[deviceName]; ok {
			reportInfo := processor.mindIoReportInfo.getInfo(jobId, uceNode.NodeName, deviceName)
			jobUceNodeInfo.DeviceInfo[uceDevice.DeviceName] = uceDeviceInfo{
				DeviceName:   deviceName,
				FaultTime:    uceDevice.FaultTime,
				RecoverTime:  reportInfo.RecoverTime,
				CompleteTime: reportInfo.CompleteTime,
			}
		}
	}

	return jobUceNodeInfo
}

func (processor *UceFaultProcessor) Process() {
	if kube.JobMgr == nil {
		hwlog.RunLog.Infof("jobMgr is nil, cannot Filter uce fault report")
		return
	}
	deviceInfos := processor.deviceCenter.GetDeviceInfos()
	processor.uceDeviceOfNode = processor.getUceDeviceOfNodes(deviceInfos)
	processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs(deviceInfos)
	currentTime := time.Now().UnixMilli()
	processor.deviceCenter.SetDeviceInfos(processor.processUceFaultInfo(deviceInfos, currentTime))
}

func (processor *UceFaultProcessor) processUceFaultInfo(
	deviceInfos map[string]*constant.DeviceInfo, currentTime int64) map[string]*constant.DeviceInfo {
	for cmName, deviceInfo := range deviceInfos {
		nodeName, err := cmNameToNodeName(cmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		faultList := processor.processEachNodeUceFaultInfo(nodeName, deviceInfo, currentTime)
		deviceInfo.DeviceList[device.GetFaultListKey()] = faultList
	}
	return deviceInfos
}

func (processor *UceFaultProcessor) processEachNodeUceFaultInfo(
	nodeName string, orgDeviceInfo *constant.DeviceInfo, currentTime int64) string {
	faultMap := device.GetFaultMap(orgDeviceInfo)
	for _, uceJob := range processor.uceDevicesOfUceJob {
		for deviceName, uceDevice := range uceJob.UceNode[nodeName].DeviceInfo {
			if processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime) {
				faultMap = processor.filterUceDeviceFaultInfo(deviceName, faultMap)
			}
		}
	}
	return device.FaultMapToArrayToString(faultMap)
}

func (processor *UceFaultProcessor) filterUceDeviceFaultInfo(
	deviceName string, faultMap map[string][]constant.DeviceFault) map[string][]constant.DeviceFault {
	for _, fault := range faultMap[deviceName] {
		// filter device's uce fault
		if device.IsUceFault(fault) {
			faultMap = device.DeleteFaultFromFaultMap(faultMap, fault)
		}
	}
	return faultMap
}

func (processor *UceFaultProcessor) canFilterUceDeviceFaultInfo(uceDevice uceDeviceInfo, currentTime int64) bool {
	if processor.currentTimeIsNotExceedMindIoReportRecoverTimeout(uceDevice, currentTime) {
		return true
	}
	if processor.mindIoRecoverTimeIsNotExceedRecoverTimeout(uceDevice) {
		if processor.currentTimeIsNotExceedMindIoReportCompleteTimeout(uceDevice, currentTime) {
			return true
		} else if processor.mindIoReportCompleteTimeIsNotExceedCompleteTimeout(uceDevice) {
			return true
		}
		return false
	}
	return false
}

func (processor *UceFaultProcessor) currentTimeIsNotExceedMindIoReportRecoverTimeout(
	uceDevice uceDeviceInfo, currentTime int64) bool {
	return uceDevice.FaultTime >= currentTime-processor.JobReportRecoverTimeout
}

func (processor *UceFaultProcessor) mindIoRecoverTimeIsNotExceedRecoverTimeout(
	uceDevice uceDeviceInfo) bool {
	return uceDevice.FaultTime >= uceDevice.RecoverTime-processor.JobReportRecoverTimeout
}

func (processor *UceFaultProcessor) currentTimeIsNotExceedMindIoReportCompleteTimeout(
	uceDevice uceDeviceInfo, currentTime int64) bool {
	return processor.JobReportCompleteTimeout+uceDevice.RecoverTime >= currentTime
}

func (processor *UceFaultProcessor) mindIoReportCompleteTimeIsNotExceedCompleteTimeout(
	uceDevice uceDeviceInfo) bool {
	// invalid complete time
	if uceDevice.CompleteTime < uceDevice.FaultTime || uceDevice.CompleteTime < uceDevice.RecoverTime {
		return false
	}
	return processor.JobReportCompleteTimeout+uceDevice.RecoverTime >= uceDevice.CompleteTime
}

func (processor *UceFaultProcessor) getUceDeviceOfNodes(deviceInfos map[string]*constant.DeviceInfo) map[string]uceNodeInfo {
	uceNodes := make(map[string]uceNodeInfo)
	for _, deviceInfo := range deviceInfos {
		nodeName, err := cmNameToNodeName(deviceInfo.CmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		uceFaultDevicesOnNode := processor.getUceFaultDevices(nodeName, deviceInfo)

		if len(uceFaultDevicesOnNode.DeviceInfo) == 0 {
			continue
		}
		uceNodes[nodeName] = uceFaultDevicesOnNode
	}
	return uceNodes
}

func (processor *UceFaultProcessor) getUceDevicesForUceTolerateJobs(deviceInfos map[string]*constant.DeviceInfo) map[string]uceJobInfo {
	nodesName := getNodesNameFromDeviceInfo(deviceInfos)
	uceJobs := make(map[string]uceJobInfo)
	kube.JobMgr.RwMutex.RLock()
	defer kube.JobMgr.RwMutex.RUnlock()
	for jobUid, worker := range kube.JobMgr.BsWorker {
		// If job cannot tolerate uce fault, don't Filter device info
		if !kube.JobMgr.JobTolerateUceFault(jobUid) {
			continue
		}
		workerInfo := worker.GetWorkerInfo()
		serverList := workerInfo.CMData.GetServerList()
		jobInfo := uceJobInfo{
			// node->uceNodeInfo
			UceNode: make(map[string]uceNodeInfo),
			JobId:   jobUid,
		}
		for _, nodeName := range nodesName {
			devicesOfJobOnNode := getDevicesNameOfJobOnNode(nodeName, serverList, jobUid)
			if len(devicesOfJobOnNode) == 0 {
				continue
			}
			jobInfo.UceNode[nodeName] =
				processor.initUceDeviceFromNodeAndMindIo(jobUid,
					processor.uceDeviceOfNode[nodeName], serverList)
		}
		if len(jobInfo.UceNode) != 0 {
			uceJobs[jobUid] = jobInfo
		}
	}
	return uceJobs
}

func (processor *UceFaultProcessor) getUceFaultDevices(nodeName string, deviceInfo *constant.DeviceInfo) uceNodeInfo {
	faultMap := device.GetFaultMap(deviceInfo)
	nodeInfo := uceNodeInfo{
		NodeName:   nodeName,
		DeviceInfo: make(map[string]uceDeviceInfo),
	}
	for _, deviceFaults := range faultMap {
		for _, fault := range deviceFaults {
			if !device.IsUceFault(fault) {
				continue
			}
			nodeInfo.DeviceInfo[fault.NPUName] = uceDeviceInfo{
				DeviceName:   fault.NPUName,
				FaultTime:    fault.FaultTime,
				RecoverTime:  constant.JobNotRecover,
				CompleteTime: constant.JobNotRecoverComplete,
			}
		}
	}
	return nodeInfo
}
