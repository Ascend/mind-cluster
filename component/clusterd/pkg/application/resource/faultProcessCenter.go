package resource

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/switchinfo"
	"clusterd/pkg/interface/kube"
	"context"
	"fmt"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"sync"
	"time"
)

var faultProcessCenter *FaultProcessCenter

// The faultProcessor process the fault information.
type faultProcessor interface {
	process()
}

// The FaultProcessCenter maintain the fault information.
type FaultProcessCenter struct {
	deviceCenter      *deviceFaultProcessCenter
	nodeCenter        *nodeFaultProcessCenter
	switchCenter      *switchFaultProcessCenter
	notifyProcessChan chan int
}

type baseFaultCenter struct {
	processors        []faultProcessor
	lastProcessTime   int64
	subscribeChannels []chan struct{}
	processPeriod     int64
}

func (baseCenter *baseFaultCenter) isProcessLimited(currentTime int64) bool {
	return baseCenter.lastProcessTime+baseCenter.processPeriod < currentTime
}

func (baseCenter *baseFaultCenter) process() {
	currentTime := time.Now().UnixMilli()
	if baseCenter.isProcessLimited(currentTime) {
		return
	}
	baseCenter.lastProcessTime = currentTime
	for _, processor := range baseCenter.processors {
		processor.process()
	}
	for _, ch := range baseCenter.subscribeChannels {
		ch <- struct{}{}
	}
}

// deviceFaultProcessCenter
type deviceFaultProcessCenter struct {
	baseFaultCenter
	mutex sync.RWMutex
	infos map[string]*constant.DeviceInfo
}

func newDeviceFaultProcessCenter() *deviceFaultProcessCenter {
	var processorForUceAccompanyFault = &uceAccompanyFaultProcessor{
		DiagnosisAccompanyTimeout: constant.DiagnosisAccompanyTimeout,
		uceAccompanyFaultQue:      make(map[string]map[string][]constant.DeviceFault),
		uceFaultTime:              make(map[string]map[string]int64),
	}
	var processorForUceFault = &uceFaultProcessor{
		JobReportRecoverTimeout:  constant.JobReportRecoverTimeout,
		JobReportCompleteTimeout: constant.JobReportCompleteTimeout,
		mindIoReportInfo: &mindIoReportInfosForAllJobs{
			Infos:   make(map[string]map[string]map[string]mindIoReportInfo),
			RwMutex: sync.RWMutex{},
		},
	}
	var processForJobFaultRank = &jobRankFaultInfoProcessor{
		jobFaultInfos: make(map[string]FaultInfo),
	}

	return &deviceFaultProcessCenter{
		mutex: sync.RWMutex{},
		infos: make(map[string]*constant.DeviceInfo),
		baseFaultCenter: baseFaultCenter{
			processors: []faultProcessor{
				processorForUceAccompanyFault,
				processorForUceFault,
				processForJobFaultRank,
			},
			lastProcessTime:   0,
			subscribeChannels: make([]chan struct{}, 0),
			processPeriod:     0,
		},
	}
}

// nodeFaultProcessCenter
type nodeFaultProcessCenter struct {
	baseFaultCenter
	infos map[string]*constant.NodeInfo
	mutex sync.RWMutex
}

func newNodeFaultProcessCenter() *nodeFaultProcessCenter {
	return &nodeFaultProcessCenter{
		infos: make(map[string]*constant.NodeInfo),
		mutex: sync.RWMutex{},
		baseFaultCenter: baseFaultCenter{
			processors:        make([]faultProcessor, 0),
			lastProcessTime:   0,
			subscribeChannels: make([]chan struct{}, 0),
			processPeriod:     0,
		},
	}
}

// switchFaultProcessCenter
type switchFaultProcessCenter struct {
	baseFaultCenter
	infos map[string]*constant.SwitchInfo
	mutex sync.RWMutex
}

func newSwitchFaultProcessCenter() *switchFaultProcessCenter {
	return &switchFaultProcessCenter{
		infos: make(map[string]*constant.SwitchInfo),
		mutex: sync.RWMutex{},
		baseFaultCenter: baseFaultCenter{
			processors:        make([]faultProcessor, 0),
			lastProcessTime:   0,
			subscribeChannels: make([]chan struct{}, 0),
			processPeriod:     0,
		},
	}
}

func (deviceCenter *deviceFaultProcessCenter) getDeviceInfos() map[string]*constant.DeviceInfo {
	deviceCenter.mutex.RLock()
	defer deviceCenter.mutex.RUnlock()
	return device.DeepCopyInfos(deviceCenter.infos)
}

func (switchCenter *switchFaultProcessCenter) getSwitchInfos() map[string]*constant.SwitchInfo {
	switchCenter.mutex.RLock()
	defer switchCenter.mutex.RUnlock()
	return switchinfo.DeepCopyInfos(switchCenter.infos)
}

func (nodeCenter *nodeFaultProcessCenter) getNodeInfos() map[string]*constant.NodeInfo {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	return node.DeepCopyInfos(nodeCenter.infos)
}

func (nodeCenter *nodeFaultProcessCenter) setNodeInfos(infos map[string]*constant.NodeInfo) {
	nodeCenter.mutex.RLock()
	defer nodeCenter.mutex.RUnlock()
	nodeCenter.infos = node.DeepCopyInfos(infos)
}

func (deviceCenter *deviceFaultProcessCenter) setDeviceInfos(infos map[string]*constant.DeviceInfo) {
	deviceCenter.mutex.Lock()
	defer deviceCenter.mutex.Unlock()
	deviceCenter.infos = device.DeepCopyInfos(infos)
}

func (switchCenter *switchFaultProcessCenter) setSwitchInfos(infos map[string]*constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	switchCenter.infos = switchinfo.DeepCopyInfos(infos)
}

func (center *FaultProcessCenter) process() {
	center.deviceCenter.process()
	center.nodeCenter.process()
	center.switchCenter.process()
}

func NewFaultProcessCenter(ctx context.Context) {
	faultProcessCenter = &FaultProcessCenter{
		deviceCenter:      newDeviceFaultProcessCenter(),
		nodeCenter:        newNodeFaultProcessCenter(),
		switchCenter:      newSwitchFaultProcessCenter(),
		notifyProcessChan: make(chan int),
	}
	go faultProcessCenter.work(ctx)
}

func (center *FaultProcessCenter) informSwitchInfoAdd(oldInfo, newInfo *constant.SwitchInfo) {
	center.switchCenter.mutex.Lock()
	defer center.switchCenter.mutex.Unlock()
	length := len(center.switchCenter.infos)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("SwitchInfo length=%d > %d, SwitchInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
		return
	}
	oldInfo, found := center.switchCenter.infos[newInfo.CmName]
	center.switchCenter.infos[newInfo.CmName] = newInfo
	if found && switchinfo.BusinessDataIsNotEqual(oldInfo, newInfo) {
		center.notifyFaultCenterProcess(constant.SWITCH_FAULT)
	}
}

func (center *FaultProcessCenter) informSwitchInfoDel(oldInfo, newInfo *constant.SwitchInfo) {
	center.switchCenter.mutex.Lock()
	defer center.switchCenter.mutex.Unlock()
	oldInfo, found := center.switchCenter.infos[newInfo.CmName]
	delete(center.switchCenter.infos, newInfo.CmName)
	center.switchCenter.infos[newInfo.CmName] = newInfo
	if found {
		center.notifyFaultCenterProcess(constant.SWITCH_FAULT)
	}
}

func (center *FaultProcessCenter) informDeviceInfoAdd(oldInfo, newInfo *constant.DeviceInfo) {
	center.deviceCenter.mutex.Lock()
	defer center.deviceCenter.mutex.Unlock()
	length := len(center.deviceCenter.infos)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("DeviceInfo length=%d > %d, DeviceInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
		return
	}
	oldInfo, found := center.deviceCenter.infos[newInfo.CmName]
	center.deviceCenter.infos[newInfo.CmName] = newInfo
	if found && device.BusinessDataIsNotEqual(oldInfo, newInfo) {
		center.notifyFaultCenterProcess(constant.SWITCH_FAULT)
	}
}

func (center *FaultProcessCenter) informDeviceInfoDel(oldInfo, newInfo *constant.DeviceInfo) {
	center.deviceCenter.mutex.Lock()
	defer center.deviceCenter.mutex.Unlock()
	oldInfo, found := center.deviceCenter.infos[newInfo.CmName]
	delete(center.deviceCenter.infos, newInfo.CmName)
	if found {
		center.notifyFaultCenterProcess(constant.DEVICE_FAULT)
	}
}

func (center *FaultProcessCenter) informNodeInfoAdd(oldInfo, newInfo *constant.NodeInfo) {
	center.nodeCenter.mutex.Lock()
	defer center.nodeCenter.mutex.Unlock()
	length := len(center.nodeCenter.infos)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("NodeInfo length=%d > %d, NodeInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
		return
	}
	oldInfo, found := center.nodeCenter.infos[newInfo.CmName]
	center.nodeCenter.infos[newInfo.CmName] = newInfo
	if found && node.BusinessDataIsNotEqual(oldInfo, newInfo) {
		center.notifyFaultCenterProcess(constant.NODE_FAULT)
	}
}

func (center *FaultProcessCenter) informNodeInfoDel(oldInfo, newInfo *constant.NodeInfo) {
	center.nodeCenter.mutex.Lock()
	defer center.nodeCenter.mutex.Unlock()
	oldInfo, found := center.nodeCenter.infos[newInfo.CmName]
	delete(center.nodeCenter.infos, newInfo.CmName)
	if found {
		center.notifyFaultCenterProcess(constant.NODE_FAULT)
	}
}

func (center *FaultProcessCenter) notifyFaultCenterProcess(whichToProcess int) {
	center.notifyProcessChan <- whichToProcess
}

func (center *FaultProcessCenter) work(ctx context.Context) {
	centerTicker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("FaultProcessCenter stop work.")
			return
		case which := <-center.notifyProcessChan:
			switch which {
			case constant.ALL_FAULT:
				center.process()
			case constant.DEVICE_FAULT:
				center.deviceCenter.process()
			case constant.NODE_FAULT:
				center.nodeCenter.process()
			case constant.SWITCH_FAULT:
				center.switchCenter.process()
			}
		case <-centerTicker.C:
			center.process()
		}
	}
}

func (center *FaultProcessCenter) getNodesNameFromDeviceInfo() []string {
	nodesName := make([]string, 0)
	for cmName, _ := range center.deviceCenter.infos {
		nodeName, err := util.CmNameToNodeName(cmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		nodesName = append(nodesName, nodeName)
	}
	return nodesName
}

func (center *FaultProcessCenter) getUceFaultProcessor() (*uceFaultProcessor, error) {
	for _, processor := range center.deviceCenter.processors {
		if processor, ok := processor.(*uceFaultProcessor); ok {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("can not find uceFaultProcessor in FaultProcessCenter")
}

func (center *FaultProcessCenter) getUceAccompanyFaultProcessor() (*uceAccompanyFaultProcessor, error) {
	for _, processor := range center.deviceCenter.processors {
		if processor, ok := processor.(*uceAccompanyFaultProcessor); ok {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("can not find uceAccompanyFaultProcessor in FaultProcessCenter")
}

func (center *FaultProcessCenter) getJobFaultRankProcessor() (*jobRankFaultInfoProcessor, error) {
	for _, processor := range center.deviceCenter.processors {
		if processor, ok := processor.(*jobRankFaultInfoProcessor); ok {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("can not find jobRankFaultInfoProcessor in FaultProcessCenter")
}

// CallbackForReportUceInfo cluster grpc should call back for report uce fault situation
func (center *FaultProcessCenter) CallbackForReportUceInfo(jobUid, rankId string, recoverTime int64) error {
	processor, err := center.getUceFaultProcessor()
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
	deviceName := util.DeviceID2DeviceKey(deviceId)
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
	center.notifyFaultCenterProcess(constant.DEVICE_FAULT)
	return nil
}

// RegisterSubscriber to notify fault occurrence
func (center *FaultProcessCenter) RegisterSubscriber(ch chan struct{}, which int) {
	switch which {
	case constant.SWITCH_FAULT:
		center.switchCenter.registerSubscriber(ch)
	case constant.NODE_FAULT:
		center.nodeCenter.registerSubscriber(ch)
	case constant.DEVICE_FAULT:
		center.deviceCenter.registerSubscriber(ch)
	case constant.ALL_FAULT:
		center.switchCenter.registerSubscriber(ch)
		center.nodeCenter.registerSubscriber(ch)
		center.deviceCenter.registerSubscriber(ch)
	}
	hwlog.RunLog.Errorf("Wrong number %d, cannot decide which to regiter.", which)
}

func (deviceCenter *deviceFaultProcessCenter) registerSubscriber(ch chan struct{}) {
	if deviceCenter.subscribeChannels == nil {
		deviceCenter.subscribeChannels = make([]chan struct{}, 0)
	}
	length := len(deviceCenter.subscribeChannels)
	if length > constant.MAX_FAULT_CENTER_SUBSCRIBER {
		hwlog.RunLog.Errorf("The deviceCenter number of registrants is %d, cannot add any more.", length)
	}
	deviceCenter.subscribeChannels = append(deviceCenter.subscribeChannels, ch)
	hwlog.RunLog.Infof("The deviceCenter number of registrants is %d.", length+1)
}

func (switchCenter *switchFaultProcessCenter) registerSubscriber(ch chan struct{}) {
	if switchCenter.subscribeChannels == nil {
		switchCenter.subscribeChannels = make([]chan struct{}, 0)
	}
	length := len(switchCenter.subscribeChannels)
	if length > constant.MAX_FAULT_CENTER_SUBSCRIBER {
		hwlog.RunLog.Errorf("The switchCenter number of registrants is %d, cannot add any more.", length)
	}
	switchCenter.subscribeChannels = append(switchCenter.subscribeChannels, ch)
	hwlog.RunLog.Infof("The switchCenter number of registrants is %d.", length+1)
}

func (nodeCenter *nodeFaultProcessCenter) registerSubscriber(ch chan struct{}) {
	if nodeCenter.subscribeChannels == nil {
		nodeCenter.subscribeChannels = make([]chan struct{}, 0)
	}
	length := len(nodeCenter.subscribeChannels)
	if length > constant.MAX_FAULT_CENTER_SUBSCRIBER {
		hwlog.RunLog.Errorf("The nodeCenter number of registrants is %d, cannot add any more.", length)
	}
	nodeCenter.subscribeChannels = append(nodeCenter.subscribeChannels, ch)
	hwlog.RunLog.Infof("The nodeCenter number of registrants is %d.", length+1)
}

// TODO 他们需要什么？该函数有点定制化
func (center *FaultProcessCenter) QueryJobsFaultInfo() map[string]FaultInfo {
	processor, err := center.getJobFaultRankProcessor()
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil
	}
	return processor.getJobFaultRankInfos()
}

func (center *FaultProcessCenter) QueryDeviceInfoToReport() map[string]*constant.DeviceInfo {
	return center.deviceCenter.getDeviceInfos()
}

func (center *FaultProcessCenter) QuerySwitchInfoToReport() map[string]*constant.SwitchInfo {
	return center.switchCenter.getSwitchInfos()
}

func (center *FaultProcessCenter) QueryNodeInfoToReport() map[string]*constant.NodeInfo {
	return center.nodeCenter.getNodeInfos()
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

func (processor *uceFaultProcessor) initUceDeviceFromNodeAndMindIo(jobId string,
	uceNode uceNodeInfo, serverList []*job.ServerHccl) uceNodeInfo {
	devicesOfJobOnNode := getDevicesNameOfJobOnNode(uceNode.NodeName, serverList)

	jobUceNodeInfo := uceNodeInfo{
		NodeName:   uceNode.NodeName,
		DeviceInfo: make(map[string]uceDeviceInfo),
	}

	for _, deviceOfJob := range devicesOfJobOnNode {
		deviceName := util.DeviceID2DeviceKey(deviceOfJob.DeviceID)
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

func (processor *uceFaultProcessor) process() {
	if kube.JobMgr == nil {
		hwlog.RunLog.Infof("jobMgr is nil, cannot Filter uce fault report")
		return
	}
	deviceInfos := faultProcessCenter.deviceCenter.getDeviceInfos()
	processor.uceDeviceOfNode = processor.getUceDeviceOfNodes(deviceInfos)
	processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs()
	currentTime := time.Now().UnixMilli()
	faultProcessCenter.deviceCenter.setDeviceInfos(processor.processUceFaultInfo(deviceInfos, currentTime))
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
		deviceInfo.DeviceList[device.GetFaultListKey()] = faultList
	}
	return deviceInfos
}

func (processor *uceFaultProcessor) processEachNodeUceFaultInfo(
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

func (processor *uceFaultProcessor) filterUceDeviceFaultInfo(
	deviceName string, faultMap map[string][]constant.DeviceFault) map[string][]constant.DeviceFault {
	for _, fault := range faultMap[deviceName] {
		// filter device's uce fault
		if device.IsUceFault(fault) {
			faultMap = device.DeleteFaultFromFaultMap(faultMap, fault)
		}
	}
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

func (processor *uceFaultProcessor) getUceDeviceOfNodes(deviceInfos map[string]*constant.DeviceInfo) map[string]uceNodeInfo {
	uceNodes := make(map[string]uceNodeInfo)
	for _, deviceInfo := range deviceInfos {
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

func (processor *uceFaultProcessor) getUceDevicesForUceTolerateJobs() map[string]uceJobInfo {
	nodesName := faultProcessCenter.getNodesNameFromDeviceInfo()
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
			devicesOfJobOnNode := getDevicesNameOfJobOnNode(nodeName, serverList)
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

func (processor *uceFaultProcessor) getUceFaultDevices(nodeName string, deviceInfo *constant.DeviceInfo) uceNodeInfo {
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

func getDevicesNameOfJobOnNode(nodeName string, serverList []*job.ServerHccl) []*job.Device {
	var devices []*job.Device
	for _, server := range serverList {
		if server.ServerName != nodeName {
			continue
		}
		devices = server.DeviceList
	}
	return devices
}

// uceAccompanyFaultProcessor:
// aic aiv fault can be 1) accompanied by uce fault, also can 2) curr alone.
// if 1) aic aiv fault should be filtered. Once find aic fault, check if there is an uce fault 5s ago
// if 2) aic aiv fault should not be retained.
type uceAccompanyFaultProcessor struct {
	// maintain 5s ago device info
	DiagnosisAccompanyTimeout int64
	// nodeName -> deviceName -> faultQue
	uceAccompanyFaultQue map[string]map[string][]constant.DeviceFault
	// uceFaultTime
	uceFaultTime map[string]map[string]int64
}

func (processor *uceAccompanyFaultProcessor) uceAccompanyFaultInQue(deviceInfos map[string]*constant.DeviceInfo) {
	for _, deviceInfo := range deviceInfos {
		nodeName, err := util.CmNameToNodeName(deviceInfo.CmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		processor.uceAccompanyFaultForNode(nodeName, deviceInfo)
	}
}

func (processor *uceAccompanyFaultProcessor) uceAccompanyFaultForNode(nodeName string, deviceInfo *constant.DeviceInfo) {
	if _, ok := processor.uceAccompanyFaultQue[nodeName]; !ok {
		processor.uceAccompanyFaultQue[nodeName] = make(map[string][]constant.DeviceFault)
		processor.uceFaultTime[nodeName] = make(map[string]int64)
	}
	faultMap := device.GetFaultMap(deviceInfo)
	for deviceName, deviceFaults := range faultMap {
		for _, fault := range deviceFaults {
			if device.IsUceFault(fault) {
				processor.uceFaultTime[nodeName][deviceName] = fault.FaultTime
				continue
			}
			if !device.IsUceAccompanyFault(fault) {
				continue
			}
			if _, ok := processor.uceAccompanyFaultQue[nodeName][deviceName]; !ok {
				processor.uceAccompanyFaultQue[nodeName][deviceName] = make([]constant.DeviceFault, 0)
			}

			// in que
			faultsInfo := processor.uceAccompanyFaultQue[nodeName][deviceName]
			processor.uceAccompanyFaultQue[nodeName][deviceName] = append(faultsInfo, fault)
		}
	}
}

func (processor *uceAccompanyFaultProcessor) filterFaultInfos(currentTime int64,
	deviceInfos map[string]*constant.DeviceInfo) map[string]*constant.DeviceInfo {
	for nodeName, nodeFaults := range processor.uceAccompanyFaultQue {
		faultMap := device.GetFaultMap(deviceInfos[util.NodeNameToCmName(nodeName)])
		for deviceName, deviceFaultQue := range nodeFaults {
			newQue, newFaultMap := processor.filterFaultDevice(faultMap, currentTime, nodeName, deviceName, deviceFaultQue)
			nodeFaults[deviceName] = newQue
			faultMap = newFaultMap
		}
		deviceInfos[util.NodeNameToCmName(nodeName)].DeviceList[device.GetFaultListKey()] = device.FaultMapToArrayToString(faultMap)
	}
	return deviceInfos
}

func (processor *uceAccompanyFaultProcessor) filterFaultDevice(
	faultMap map[string][]constant.DeviceFault, currentTime int64, nodeName, deviceName string,
	deviceFaultQue []constant.DeviceFault) ([]constant.DeviceFault, map[string][]constant.DeviceFault) {
	newDeviceFaultQue := make([]constant.DeviceFault, 0)
	for _, fault := range deviceFaultQue {
		uceFaultTime := processor.getDeviceUceFaultTime(nodeName, deviceName)
		accompanyFaultTime := fault.FaultTime
		// if is accompanied fault, filter
		if processor.isAccompaniedFaultByUce(uceFaultTime, accompanyFaultTime) {
			faultMap = device.DeleteFaultFromFaultMap(faultMap, fault)
			continue
		}
		// if current is not exceed diagnosis time,
		// then cannot decide fault is accompany or not, filter, and in que to decide in next turn.
		if !processor.isCurrentExceedDiagnosisTimeout(currentTime, accompanyFaultTime) {
			faultMap = device.DeleteFaultFromFaultMap(faultMap, fault)
			newDeviceFaultQue = append(newDeviceFaultQue, fault)
		}
	}
	return newDeviceFaultQue, faultMap
}

func (processor *uceAccompanyFaultProcessor) getDeviceUceFaultTime(nodeName, deviceName string) int64 {
	if faultTime, ok := processor.uceFaultTime[nodeName][deviceName]; ok {
		return faultTime
	}
	return constant.DeviceNotFault
}

func (processor *uceAccompanyFaultProcessor) isAccompaniedFaultByUce(
	uceFaultTime, uceAccompanyFaultTime int64) bool {
	return util.Abs(uceFaultTime-uceAccompanyFaultTime) <= processor.DiagnosisAccompanyTimeout
}

func (processor *uceAccompanyFaultProcessor) isCurrentExceedDiagnosisTimeout(
	currentTime, uceAccompanyFaultTime int64) bool {
	return uceAccompanyFaultTime < currentTime-processor.DiagnosisAccompanyTimeout
}

func (processor *uceAccompanyFaultProcessor) process() {
	deviceInfos := faultProcessCenter.deviceCenter.getDeviceInfos()
	processor.uceAccompanyFaultInQue(deviceInfos)
	currentTime := time.Now().UnixMilli()
	filteredFaultInfos := processor.filterFaultInfos(currentTime, deviceInfos)
	faultProcessCenter.deviceCenter.setDeviceInfos(filteredFaultInfos)
}

// JobFaultInfoProcessor
type jobRankFaultInfoProcessor struct {
	jobFaultInfos map[string]FaultInfo
}

func (processor *jobRankFaultInfoProcessor) getJobFaultRankInfos() map[string]FaultInfo {
	return processor.jobFaultInfos
}

func (processor *jobRankFaultInfoProcessor) process() {
	deviceInfos := faultProcessCenter.deviceCenter.getDeviceInfos()
	nodesName := faultProcessCenter.getNodesNameFromDeviceInfo()
	kube.JobMgr.RwMutex.RLock()
	defer kube.JobMgr.RwMutex.RUnlock()
	if processor.jobFaultInfos == nil {
		processor.jobFaultInfos = make(map[string]FaultInfo)
	}
	for jobId, worker := range kube.JobMgr.BsWorker {
		jobFaultInfo := FaultInfo{
			jobId:     jobId,
			faultList: make([]FaultRank, 0),
		}

		workerInfo := worker.GetWorkerInfo()
		serverList := workerInfo.CMData.GetServerList()
		for _, nodeName := range nodesName {
			faultRankList := findFaultOnNodeForJob(deviceInfos, nodeName, serverList)
			jobFaultInfo.faultList = append(jobFaultInfo.faultList, faultRankList...)
		}
		processor.jobFaultInfos[jobId] = jobFaultInfo
	}
}

func findFaultOnNodeForJob(
	deviceInfos map[string]*constant.DeviceInfo, nodeName string, serverList []*job.ServerHccl) []FaultRank {
	faultMap := device.GetFaultMap(deviceInfos[nodeName])
	devicesOfJobOnNode := getDevicesNameOfJobOnNode(nodeName, serverList)
	faultRankList := make([]FaultRank, 0)
	if len(devicesOfJobOnNode) == 0 {
		for _, deviceInfo := range devicesOfJobOnNode {
			deviceName := util.DeviceID2DeviceKey(deviceInfo.DeviceID)
			if faultList, ok := faultMap[deviceName]; ok {
				for _, fault := range faultList {
					faultRankList = append(faultRankList, FaultRank{
						rankId:    deviceInfo.RankID,
						faultCode: fault.FaultCode,
					})
				}
			}
		}
	}
	return faultRankList
}

type FaultRank struct {
	rankId    string
	faultCode string
}

type FaultInfo struct {
	jobId     string
	faultList []FaultRank
}
