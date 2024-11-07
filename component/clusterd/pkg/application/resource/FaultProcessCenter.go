package resource

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/switchinfo"
	"clusterd/pkg/interface/kube"
	"encoding/json"
	"fmt"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"math"
	"sync"
	"time"
)

var faultProcessCenter = &FaultProcessCenter{
	deviceFaultProcessor: []faultProcessor{&uceFaultProcessor{
		JobReportRecoverTimeout:  10,
		JobReportCompleteTimeout: 30,
		mindIoReportInfo: &mindIoReportInfosForAllJobs{
			Infos:   make(map[string]map[string]map[string]mindIoReportInfo),
			RwMutex: sync.RWMutex{},
		},
	}},
}

const JobNotRecover = int64(math.MaxInt64)
const JobNotRecoverComplete = int64(math.MaxInt64)

// The faultProcessor process the fault information.
type faultProcessor interface {
	process()
}

type FaultProcessCenter struct {
	deviceInfos          map[string]*constant.DeviceInfo
	nodeInfos            map[string]*constant.NodeInfo
	switchInfos          map[string]*constant.SwitchInfo
	deviceFaultProcessor []faultProcessor
	nodeFaultProcessor   []faultProcessor
	switchFaultProcessor []faultProcessor
}

/*
The uceFaultProcessor process uce fault reporting information.
If the device fault is UCE fault, then determine whether the job running on the device can tolerate UCE faults.
If they can tolerate it, the reporting of the UCE fault should be delayed by 10 seconds.
*/
type uceFaultProcessor struct {
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
			RecoverTime:  JobNotRecover,
			CompleteTime: JobNotRecoverComplete,
		}
	}
	reportInfos.RwMutex.RLock()
	defer reportInfos.RwMutex.RUnlock()
	if info, ok := reportInfos.Infos[jobId][nodeName][deviceName]; ok {
		return info
	}
	return mindIoReportInfo{
		RecoverTime:  JobNotRecover,
		CompleteTime: JobNotRecoverComplete,
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

func (jobInfo *uceJobInfo) initUceDeviceFromNodeAndMindIo(
	uceNode uceNodeInfo, reportInfos *mindIoReportInfosForAllJobs, serverList []*job.ServerHccl) uceNodeInfo {
	devicesOfJobOnNode := jobInfo.getDevicesNameOfJobOnNode(uceNode.NodeName, serverList)

	jobUceNodeInfo := uceNodeInfo{
		NodeName:   uceNode.NodeName,
		DeviceInfo: make(map[string]uceDeviceInfo),
	}

	for _, deviceOfJob := range devicesOfJobOnNode {
		if uceDevice, ok := uceNode.DeviceInfo[deviceOfJob]; ok {
			reportInfo := reportInfos.getInfo(jobInfo.JobId, uceNode.NodeName, deviceOfJob)
			jobUceNodeInfo.DeviceInfo[uceDevice.DeviceName] = uceDeviceInfo{
				DeviceName:   deviceOfJob,
				FaultTime:    uceDevice.FaultTime,
				RecoverTime:  reportInfo.RecoverTime,
				CompleteTime: reportInfo.CompleteTime,
			}
		}
	}

	return jobUceNodeInfo
}

func (processor *uceFaultProcessor) process() {
	if kube.JobMgr == nil {
		hwlog.RunLog.Infof("jobMgr is nil, cannot Filter uce fault report")
		return
	}
	processor.uceDeviceOfNode = processor.getUceDeviceOfNodes()
	processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs()
	currentTime := time.Now().UnixMilli()
	faultProcessCenter.deviceInfos = processor.processUceFaultInfo(faultProcessCenter.deviceInfos, currentTime)
}

func (processor *uceFaultProcessor) processUceFaultInfo(
	deviceInfos map[string]*constant.DeviceInfo, currentTime int64) map[string]*constant.DeviceInfo {
	for cmName, deviceInfo := range deviceInfos {
		nodeName, err := util.CmNameToNodeName(cmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		faultList := processor.processEachNodeUceFaultInfo(nodeName, deviceInfo, currentTime)
		faultListBytes, err := json.Marshal(faultList)
		if err != nil {
			hwlog.RunLog.Error(fmt.Errorf("marshal fault list for node %s failed. Exception: %v", nodeName, err))
		}

		deviceInfo.DeviceList[device.GetFaultListKey()] = string(faultListBytes)
	}
	return deviceInfos
}

func (processor *uceFaultProcessor) processEachNodeUceFaultInfo(
	nodeName string, orgDeviceInfo *constant.DeviceInfo, currentTime int64) []constant.DeviceFault {
	faultMap := device.GetFaultMap(orgDeviceInfo)
	for _, uceJob := range processor.uceDevicesOfUceJob {
		for deviceName, uceDevice := range uceJob.UceNode[nodeName].DeviceInfo {
			if processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime) {
				faultMap = processor.filterUceDeviceFaultInfo(deviceName, faultMap)
			}
		}
	}
	deviceFaultList := make([]constant.DeviceFault, 0)
	for _, faultDevice := range faultMap {
		deviceFaultList = append(deviceFaultList, faultDevice)
	}
	return deviceFaultList
}

func (processor *uceFaultProcessor) filterUceDeviceFaultInfo(
	deviceName string, faultMap map[string]constant.DeviceFault) map[string]constant.DeviceFault {
	delete(faultMap, deviceName)
	return faultMap
}

func (processor *uceFaultProcessor) canFilterUceDeviceFaultInfo(uceDevice uceDeviceInfo, currentTime int64) bool {
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

func (processor *uceFaultProcessor) currentTimeIsNotExceedMindIoReportRecoverTimeout(
	uceDevice uceDeviceInfo, currentTime int64) bool {
	return uceDevice.FaultTime >= currentTime-processor.JobReportRecoverTimeout
}

func (processor *uceFaultProcessor) mindIoRecoverTimeIsNotExceedRecoverTimeout(
	uceDevice uceDeviceInfo) bool {
	return uceDevice.FaultTime >= uceDevice.RecoverTime-processor.JobReportRecoverTimeout
}

func (processor *uceFaultProcessor) currentTimeIsNotExceedMindIoReportCompleteTimeout(
	uceDevice uceDeviceInfo, currentTime int64) bool {
	return processor.JobReportCompleteTimeout+uceDevice.RecoverTime >= currentTime
}

func (processor *uceFaultProcessor) mindIoReportCompleteTimeIsNotExceedCompleteTimeout(
	uceDevice uceDeviceInfo) bool {
	// invalid complete time
	if uceDevice.CompleteTime < uceDevice.FaultTime || uceDevice.CompleteTime < uceDevice.RecoverTime {
		return false
	}
	return processor.JobReportCompleteTimeout+uceDevice.RecoverTime >= uceDevice.CompleteTime
}

func (processor *uceFaultProcessor) getUceDeviceOfNodes() map[string]uceNodeInfo {
	uceNodes := make(map[string]uceNodeInfo)
	for _, deviceInfo := range faultProcessCenter.deviceInfos {
		nodeName, err := util.CmNameToNodeName(deviceInfo.CmName)
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

func (processor *uceFaultProcessor) getNodesNameFromDeviceInfo() []string {
	nodesName := make([]string, 0)
	for cmName, _ := range faultProcessCenter.deviceInfos {
		nodeName, err := util.CmNameToNodeName(cmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		nodesName = append(nodesName, nodeName)
	}
	return nodesName
}

func (processor *uceFaultProcessor) getUceDevicesForUceTolerateJobs() map[string]uceJobInfo {
	nodesName := processor.getNodesNameFromDeviceInfo()
	uceJobs := make(map[string]uceJobInfo)
	kube.JobMgr.RwMutex.RLock()
	defer kube.JobMgr.RwMutex.RUnlock()
	for jobUid, worker := range kube.JobMgr.BsWorker {
		// If job cannot tolerate uce fault, don't Filter device info
		if !processor.jobTolerateUceFault(jobUid) {
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
			devicesOfJobOnNode := jobInfo.getDevicesNameOfJobOnNode(nodeName, serverList)
			if len(devicesOfJobOnNode) == 0 {
				continue
			}
			jobInfo.UceNode[nodeName] =
				jobInfo.initUceDeviceFromNodeAndMindIo(
					processor.uceDeviceOfNode[nodeName], processor.mindIoReportInfo, serverList)
		}
		if len(jobInfo.UceNode) != 0 {
			uceJobs[jobUid] = jobInfo
		}
	}
	return uceJobs
}

// TODO 如何判断job是uce容忍的
func (processor *uceFaultProcessor) jobTolerateUceFault(jobUid string) bool {
	return true
}

func (processor *uceFaultProcessor) getUceFaultDevices(nodeName string, deviceInfo *constant.DeviceInfo) uceNodeInfo {
	faultList := device.GetFaultMap(deviceInfo)
	nodeInfo := uceNodeInfo{
		NodeName:   nodeName,
		DeviceInfo: make(map[string]uceDeviceInfo),
	}
	for _, faultDevice := range faultList {
		nodeInfo.DeviceInfo[faultDevice.NPUName] = uceDeviceInfo{
			DeviceName:   faultDevice.NPUName,
			FaultTime:    faultDevice.FaultTime,
			RecoverTime:  JobNotRecover,
			CompleteTime: JobNotRecoverComplete,
		}
	}
	return nodeInfo
}

// CallbackForReportUceInfo cluster grpc should call back for report uce fault situation
func (processor *uceFaultProcessor) CallbackForReportUceInfo(jobUid, nodeName, deviceName string, info mindIoReportInfo) {
	processor.mindIoReportInfo.RwMutex.Lock()
	defer processor.mindIoReportInfo.RwMutex.Unlock()
	reportInfo := processor.mindIoReportInfo.Infos
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
}

func (jobInfo *uceJobInfo) getDevicesNameOfJobOnNode(nodeName string, serverList []*job.ServerHccl) []string {
	var devices []*job.Device
	for _, server := range serverList {
		if server.ServerName != nodeName {
			continue
		}
		devices = server.DeviceList
	}
	devicesName := make([]string, len(devices))
	for idx, dev := range devices {
		devicesName[idx] = util.DeviceID2DeviceKey(dev.DeviceID)
	}
	return devicesName
}

func (faultProcessCenter *FaultProcessCenter) QueryDeviceInfoToReport() map[string]*constant.DeviceInfo {
	cmManager.Lock()
	faultProcessCenter.deviceInfos = device.DeepCopyInfos(cmManager.deviceInfoMap)
	cmManager.Unlock()
	for _, processor := range faultProcessCenter.deviceFaultProcessor {
		processor.process()
	}
	return faultProcessCenter.deviceInfos
}

func (faultProcessCenter *FaultProcessCenter) QuerySwitchInfoToReport() map[string]*constant.SwitchInfo {
	cmManager.Lock()
	faultProcessCenter.switchInfos = switchinfo.DeepCopyInfos(cmManager.switchInfoMap)
	cmManager.Unlock()
	for _, processor := range faultProcessCenter.switchFaultProcessor {
		processor.process()
	}
	return faultProcessCenter.switchInfos
}

func (faultProcessCenter *FaultProcessCenter) QueryNodeInfoToReport() map[string]*constant.NodeInfo {
	cmManager.Lock()
	faultProcessCenter.nodeInfos = node.DeepCopyInfos(cmManager.nodeInfoMap)
	cmManager.Unlock()
	for _, processor := range faultProcessCenter.nodeFaultProcessor {
		processor.process()
	}
	return faultProcessCenter.nodeInfos
}
