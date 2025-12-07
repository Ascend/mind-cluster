// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package relationfault contain relation fault process
package relationfault

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/kube"
)

// RelationProcessor network relation fault process
var RelationProcessor *relationFaultProcessor
var loadConfigTag = sync.Once{}

func init() {
	RelationProcessor = &relationFaultProcessor{}
}

func loadConfig() {
	if fileBytes := LoadConfigFromFile(constant.FaultCustomizationPath); fileBytes != nil {
		initRelationFaultStrategies(fileBytes)
		initRelationFaultCodesMap()
	} else {
		hwlog.RunLog.Errorf("load config from file %s failed", constant.FaultCustomizationPath)
	}
	if fileBytes := LoadConfigFromFile(constant.FaultDurationPath); fileBytes != nil {
		initFaultDuration(fileBytes)
		initFaultCodeTimeOutMap()
	} else {
		hwlog.RunLog.Errorf("load config from file %s failed", constant.FaultDurationPath)
	}
}

type relationFaultProcessor struct {
	faultJobs    map[string]*FaultJob
	deviceInfoCm map[string]*constant.AdvanceDeviceFaultCm
	switchInfoCm map[string]*constant.SwitchInfo
	nodeInfoCm   map[string]*constant.NodeInfo
}

// Process job network relation fault info
func (processor *relationFaultProcessor) Process(info any) any {
	loadConfigTag.Do(func() {
		loadConfig()
	})
	content, ok := info.(constant.AllConfigmapContent)
	if !ok {
		hwlog.RunLog.Errorf("convert info to AllConfigmapContent failed")
		return info
	}
	processor.deviceInfoCm = content.DeviceCm
	processor.switchInfoCm = content.SwitchCm
	processor.nodeInfoCm = content.NodeCm
	processor.InitFaultJobs()
	for _, fJob := range processor.faultJobs {
		fJob.Process()
	}
	return nil
}

func (processor *relationFaultProcessor) InitFaultJobs() {
	faultJobs := make(map[string]*FaultJob)
	jobServerInfoMap := job.GetJobServerInfoMap()
	for jobId, serverLists := range jobServerInfoMap.InfoMap {
		if len(serverLists) == 0 {
			hwlog.RunLog.Debugf("job %s serverList is empty", jobId)
			continue
		}
		tmpFaultJob, ok := processor.faultJobs[jobId]
		if !ok {
			tmpFaultJob = &FaultJob{}
		}
		tmpFaultJob.initFaultJobAttr()
		for nodeName, serverList := range serverLists {
			tmpFaultJob.PodNames[serverList.ServerName] = serverList.PodID
			tmpFaultJob.NameSpace = serverList.PodNameSpace
			switchInfo, ok := processor.switchInfoCm[constant.SwitchInfoPrefix+nodeName]
			if ok {
				tmpFaultJob.initBySwitchFault(switchInfo, serverList)
			}
			deviceInfo, ok := processor.deviceInfoCm[nodeName]
			if ok {
				tmpFaultJob.IsA3Job = deviceInfo.SuperPodID >= 0
				tmpFaultJob.initByDeviceFault(deviceInfo, serverList)
			}
		}
		faultJobs[jobId] = tmpFaultJob
		hwlog.RunLog.Debugf("init fault job %v", util.ObjToString(faultJobs))
	}
	processor.faultJobs = faultJobs
}

// GetPodStrategiesMapsByJobId get PodStrategiesMaps by job id
func (processor *relationFaultProcessor) GetPodStrategiesMapsByJobId(jobId string) map[string]string {
	faultJob := processor.getFaultJobByJobId(jobId)
	if faultJob == nil {
		return nil
	}
	return faultJob.PodStrategiesMaps
}

func (processor *relationFaultProcessor) getFaultJobByJobId(jobId string) *FaultJob {
	if processor.faultJobs == nil {
		return nil
	}
	faultJob, ok := processor.faultJobs[jobId]
	if !ok || faultJob == nil {
		return nil
	}
	return faultJob
}

// GetRelationFaultInfo get relation fault info
func (processor *relationFaultProcessor) GetRelationFaultInfo(jobId, nodeName string) []*constant.FaultInfo {
	faultJob := processor.getFaultJobByJobId(jobId)
	if faultJob == nil || faultJob.NodeFaultInfoMap == nil {
		return nil
	}
	return faultJob.NodeFaultInfoMap[nodeName]
}

var relationFaultStrategies = make([]constant.RelationFaultStrategy, 0)
var faultDurationStrategies = make([]constant.FaultDuration, 0)
var relationFaultTypeMap = make(sets.String)
var triggerFaultMap = make(sets.String)
var faultCodeTimeOutMap = make(map[string]int64)

// FaultJob contain some fault info about a fault job
type FaultJob struct {
	InitTime            int64
	IsA3Job             bool
	NameSpace           string
	PodNames            map[string]string
	RelationFaults      []*constant.FaultInfo
	TriggerFault        []constant.FaultInfo
	processedFaultInfo  []constant.FaultInfo
	TMOutTriggerFault   []constant.FaultInfo
	TMOutRelationFaults []*constant.FaultInfo
	FaultStrategy       constant.FaultStrategy
	SeparateNodes       sets.String
	AllFaultCode        sets.String
	ProcessingFaultCode sets.String
	PodStrategiesMaps   map[string]string
	FindNPUUnderSwitch  bool
	NodeFaultInfoMap    map[string][]*constant.FaultInfo
}

func getFaultCodeTimeOutMap() map[string]int64 {
	return faultCodeTimeOutMap
}

func setFaultCodeTimeOutMap(faultCode string, delTime int64) {
	faultCodeTimeOutMap[faultCode] = delTime
}

func getFaultCodeDelMaxTime(faultCode string) int64 {
	return getFaultCodeTimeOutMap()[faultCode]
}

func initRelationFaultStrategies(fileBytes []byte) {
	if err := json.Unmarshal(fileBytes, &relationFaultStrategies); err != nil {
		fmt.Printf("unmarshal fault code byte failed: %v", err)
		return
	}
}

func initFaultDuration(fileBytes []byte) {
	var tmpFaultDurationStrategies []constant.FaultDuration
	if err := json.Unmarshal(fileBytes, &tmpFaultDurationStrategies); err != nil {
		fmt.Printf("unmarshal fault code byte failed: %v", err)
		return
	}
	if len(tmpFaultDurationStrategies) == 0 {
		fmt.Printf("fault duration fault config is invalid")
		return
	}
	for _, faultConfig := range tmpFaultDurationStrategies {
		if !validateFaultDurationConfig(faultConfig) {
			continue
		}
		faultDurationStrategies = append(faultDurationStrategies, faultConfig)
	}
}

func validateFaultDurationConfig(faultConfig constant.FaultDuration) bool {
	if faultConfig.FaultCode == "" {
		hwlog.RunLog.Error("fault code is empty")
		return false
	}
	if faultConfig.TimeOutInterval < 0 {
		hwlog.RunLog.Error("fault code time interval is invalid",
			faultConfig.TimeOutInterval)
		return false
	}
	return true
}

func initFaultCodeTimeOutMap() {
	for _, strategy := range faultDurationStrategies {
		setFaultCodeTimeOutMap(strategy.FaultCode, strategy.TimeOutInterval)
	}
}

func initRelationFaultCodesMap() {
	for _, strategy := range relationFaultStrategies {
		triggerFaultMap.Insert(strategy.TriggerFault)
		for _, fCode := range strategy.RelationFaults {
			relationFaultTypeMap.Insert(fCode)
		}
	}
}

// LoadConfigFromFile load fault config and fault type from local file
func LoadConfigFromFile(filePath string) []byte {
	fileBytes, err := utils.LoadFile(filePath)
	if err != nil {
		return nil
	}
	return fileBytes
}

func (fJob *FaultJob) initFaultJobAttr() {
	fJob.FaultStrategy = constant.FaultStrategy{}
	fJob.TriggerFault = nil
	fJob.AllFaultCode = make(sets.String)
	fJob.SeparateNodes = make(sets.String)
	fJob.NodeFaultInfoMap = make(map[string][]*constant.FaultInfo)
	if fJob.PodNames == nil {
		fJob.PodNames = make(map[string]string)
	}
	if fJob.ProcessingFaultCode == nil {
		fJob.ProcessingFaultCode = make(sets.String)
	}
	if fJob.PodStrategiesMaps == nil {
		fJob.PodStrategiesMaps = make(map[string]string)
	}
	if fJob.InitTime == 0 {
		fJob.InitTime = time.Now().UnixMilli()
		hwlog.RunLog.Infof("init fault job InitTime %v", fJob.InitTime)
	}
}

// Process fJob network relation fault
func (fJob *FaultJob) Process() {
	fJob.preStartProcess()
	fJob.processNetworkFault()
	fJob.preStopProcess()
}

func (fJob *FaultJob) preStartProcess() {
	var networkFaultInfo []*constant.FaultInfo
	tmpTMOutRelationFaults := make([]*constant.FaultInfo, 0)
	for _, fault := range fJob.RelationFaults {
		if fJob.AllFaultCode.Has(fault.FaultUid) {
			networkFaultInfo = append(networkFaultInfo, fault)
			continue
		}
		tmpTMOutRelationFaults = append(tmpTMOutRelationFaults, fault)
		hwlog.RunLog.Infof("fault code is not exist %v, delete it in ProcessingFaultCode", fault.FaultUid)
		fJob.ProcessingFaultCode.Delete(fault.FaultUid)
	}
	now := time.Now().UnixMilli()
	for _, fault := range fJob.TMOutRelationFaults {
		if fault == nil {
			continue
		}
		if now-fault.FaultTime > fault.DealMaxTime*constant.Kilo {
			continue
		}
		tmpTMOutRelationFaults = append(tmpTMOutRelationFaults, fault)
	}
	fJob.RelationFaults = networkFaultInfo
	fJob.TMOutRelationFaults = tmpTMOutRelationFaults
	hwlog.RunLog.Debugf("after perstart precess, relation faults is %v", util.ObjToString(fJob.RelationFaults))
	hwlog.RunLog.Debugf("after perstart precess, TMOutRelationFaults faults is %v",
		util.ObjToString(fJob.TMOutRelationFaults))
}

func (fJob *FaultJob) preStopProcess() {
	fJob.clearProcessedAndTimeOutFault()
	fJob.processFaultStrategies()
}

func (fJob *FaultJob) processFaultStrategies() {
	for nodeName, devices := range fJob.FaultStrategy.DeviceLvList {
		nodeStrategy := ""
		for _, device := range devices {
			if nodeStrategy == constant.SeparateFaultStrategy {
				continue
			}
			nodeStrategy = device.Strategy
		}
		if nodeStrategy == "" || fJob.FaultStrategy.NodeLvList[nodeName] == constant.SeparateFaultStrategy {
			continue
		}
		fJob.FaultStrategy.NodeLvList[nodeName] = nodeStrategy
	}
	podStrategiesMaps := make(map[string]string, len(fJob.FaultStrategy.NodeLvList))
	for nodeName, strategy := range fJob.FaultStrategy.NodeLvList {
		if podName, ok := fJob.PodNames[nodeName]; ok {
			podStrategiesMaps[podName] = strategy
		}
	}
	newStrategiesMaps := new(map[string]string)
	if err := util.DeepCopy(newStrategiesMaps, podStrategiesMaps); err != nil {
		hwlog.RunLog.Errorf("deep copy map failed: %v", err)
		return
	}
	hwlog.RunLog.Debugf("process strategies is %v ", newStrategiesMaps)
	for podName, strategy := range *newStrategiesMaps {
		if strategy == "" {
			continue
		}
		// if strategy is same as last cycle, skip patch pod label
		if fJob.PodStrategiesMaps[podName] == strategy {
			continue
		}
		// if last strategy is SeparateFaultStrategy, skip patch pod label
		if fJob.PodStrategiesMaps[podName] == constant.SeparateFaultStrategy {
			podStrategiesMaps[podName] = constant.SeparateFaultStrategy
		}
		if err := kube.RetryPatchPodLabels(podName, fJob.NameSpace, constant.PatchPodTimes,
			map[string]string{constant.TaskFaultKey: strategy}); err != nil {
			hwlog.RunLog.Errorf("patch pod label failed: %v", err)
		}
	}
	fJob.PodStrategiesMaps = podStrategiesMaps
}

func (fJob *FaultJob) clearProcessedAndTimeOutFault() {
	networkFaultInfo := make([]*constant.FaultInfo, 0)
	preStopTime := time.Now().UnixMilli()
	for _, fault := range fJob.RelationFaults {
		if fault.ExecutedStrategy == constant.SeparateFaultStrategy {
			fJob.ProcessingFaultCode.Delete(fault.FaultUid)
			continue
		}
		if preStopTime-fault.FaultTime >= fault.DealMaxTime*constant.Kilo {
			hwlog.RunLog.Infof("fault code %s is time out, process as default strategy", fault.FaultUid)
			fJob.addFaultStrategyForTimeOutCode(fault)
			continue
		}
		networkFaultInfo = append(networkFaultInfo, fault)
	}
	fJob.RelationFaults = networkFaultInfo
}

func (fJob *FaultJob) addFaultStrategyForTimeOutCode(fault *constant.FaultInfo) {
	if fault.ExecutedStrategy != "" {
		return
	}
	hwlog.RunLog.Infof("fault code %s time %v, fault job init time %v", fault.FaultUid, fault.FaultTime, fJob.InitTime)
	if fault.FaultTime < fJob.InitTime {
		hwlog.RunLog.Infof("fault code %s is old than job, skip add default strategy", fault.FaultUid)
		return
	}
	if fault.FaultType == constant.SwitchFaultType {
		fJob.FaultStrategy.NodeLvList[fault.NodeName] = constant.SubHealthFaultStrategy
		newFault := *fault
		newFault.ExecutedStrategy = constant.SubHealthFaultStrategy
		fJob.updateNodeFaultInfoMap(&newFault)
	}
	if fault.FaultType == constant.DeviceFaultType {
		fJob.execDeviceFaultTMOut(fault)
	}
}

func (fJob *FaultJob) execDeviceFaultTMOut(fault *constant.FaultInfo) {
	if fJob.isMeetTMOutTriggerFault(fault) {
		fJob.FaultStrategy.NodeLvList[fault.NodeName] = constant.SeparateFaultStrategy
		newFault := *fault
		newFault.ExecutedStrategy = constant.SeparateFaultStrategy
		hwlog.RunLog.Infof("fault <%s> meet tmout trigger fault, will separate fault pod", fault.FaultUid)
		fJob.updateNodeFaultInfoMap(&newFault)
	}
}

func (fJob *FaultJob) isMeetTMOutTriggerFault(fault *constant.FaultInfo) bool {
	var tmpTMOutTriggerFault []constant.FaultInfo
	var isMeet bool
	for _, tmOutTriggerFault := range fJob.TMOutTriggerFault {
		if tmOutTriggerFault.FaultTime >= fault.FaultTime &&
			tmOutTriggerFault.FaultTime-fault.FaultTime <= fault.DealMaxTime*constant.Kilo {
			isMeet = true
			hwlog.RunLog.Infof("fault <%s> meet <%s> when time out exec", fault.FaultUid, tmOutTriggerFault.FaultUid)
			continue
		}
		tmpTMOutTriggerFault = append(tmpTMOutTriggerFault, tmOutTriggerFault)
	}
	fJob.TMOutTriggerFault = tmpTMOutTriggerFault
	return isMeet
}

func (fJob *FaultJob) initByDeviceFault(nodeFaultInfo *constant.AdvanceDeviceFaultCm, serverList constant.ServerHccl) {
	if fJob.SeparateNodes.Has(serverList.ServerName) {
		return
	}
	for _, deviceInfo := range serverList.DeviceList {
		deviceName := nodeFaultInfo.DeviceType + "-" + deviceInfo.DeviceID
		fault, ok := nodeFaultInfo.FaultDeviceList[deviceName]
		if !ok {
			continue
		}
		fJob.initFaultInfoByDeviceFault(fault, serverList.ServerName, deviceInfo.RankID,
			util.IsSliceContain(deviceName, nodeFaultInfo.CardUnHealthy))
	}
}

func (fJob *FaultJob) initFaultInfoByDeviceFault(
	faultList []constant.DeviceFault, nodeName, rankId string, isCardUnhealthy bool) {
	for _, fault := range faultList {
		for faultCode, faultTimeAndLevel := range fault.FaultTimeAndLevelMap {
			if isAssociateFault(faultCode) && !isCardUnhealthy {
				tmpFaultInfo := constant.FaultInfo{
					NodeName:    nodeName,
					FaultType:   constant.DeviceFaultType,
					NPUName:     fault.NPUName,
					FaultCode:   faultCode,
					FaultLevel:  faultTimeAndLevel.FaultLevel,
					FaultTime:   faultTimeAndLevel.FaultReceivedTime,
					DealMaxTime: getFaultCodeDelMaxTime(faultCode),
					FaultUid:    nodeName + "-" + fault.NPUName + "-" + faultCode,
					ForceAdd:    fault.ForceAdd,
				}
				fJob.AllFaultCode.Insert(tmpFaultInfo.FaultUid)
				fJob.addFaultInfoByCodeType(&tmpFaultInfo)
			}
		}
	}
}

func (fJob *FaultJob) addFaultInfoByCodeType(faultInfo *constant.FaultInfo) {
	if relationFaultTypeMap.Has(faultInfo.FaultCode) {
		// if linkdown fault is timeout, existed in ProcessingFaultCode and need force add in RelationFaults,
		// when in retry fault process senior
		if !faultInfo.ForceAdd && fJob.ProcessingFaultCode.Has(faultInfo.FaultUid) {
			hwlog.RunLog.Debugf("addFaultInfoByCodeType failed by code %s "+
				"is existed in ProcessingFaultCode", faultInfo.FaultUid)
			return
		}
		hwlog.RunLog.Infof("addFaultInfoByCodeType  code %s in ProcessingFaultCode", faultInfo.FaultUid)
		fJob.ProcessingFaultCode.Insert(faultInfo.FaultUid)
		fJob.RelationFaults = append(fJob.RelationFaults, faultInfo)
	}
	if triggerFaultMap.Has(faultInfo.FaultCode) {
		if fJob.IsA3Job && faultdomain.IsCqeFault(faultInfo.FaultCode) {
			fJob.TMOutTriggerFault = append(fJob.TMOutTriggerFault, *faultInfo)
			return
		}
		fJob.TriggerFault = append(fJob.TriggerFault, *faultInfo)
	}
}

func isAssociateFault(faultCode string) bool {
	return relationFaultTypeMap.Has(faultCode) || triggerFaultMap.Has(faultCode)
}

func (fJob *FaultJob) initBySwitchFault(switchInfo *constant.SwitchInfo, serverList constant.ServerHccl) {
	if switchInfo == nil {
		return
	}
	if switchInfo.NodeStatus == constant.UnHealthyState {
		fJob.SeparateNodes.Insert(serverList.ServerName)
		return
	}
	var forceAdd bool
	for _, faultInfo := range switchInfo.FaultInfo {
		if faultInfo.ForceAdd {
			forceAdd = true
		}
	}
	for faultCode, faultInfo := range switchInfo.FaultTimeAndLevelMap {
		assembledFaultCode, switchChipId, switchPortId := constant.RestoreFaultTimeAndLevelKey(faultCode)
		if isAssociateFault(assembledFaultCode) {
			tmpFaultInfo := constant.FaultInfo{
				NodeName:     serverList.ServerName,
				NPUName:      constant.AllCardId,
				FaultType:    constant.SwitchFaultType,
				FaultCode:    assembledFaultCode,
				SwitchChipId: switchChipId,
				SwitchPortId: switchPortId,
				FaultTime:    faultInfo.FaultReceivedTime,
				DealMaxTime:  getFaultCodeDelMaxTime(assembledFaultCode),
				FaultLevel:   switchInfo.FaultLevel,
				FaultUid: serverList.ServerName + "-" +
					constant.AllCardId + "-" + faultCode,
				ForceAdd: forceAdd,
			}
			fJob.AllFaultCode.Insert(tmpFaultInfo.FaultUid)
			fJob.addFaultInfoByCodeType(&tmpFaultInfo)
		}
	}
}

func (fJob *FaultJob) handleJobFault(relationFault []*constant.FaultInfo,
	triggerFault []constant.FaultInfo, strategyList []constant.RelationFaultStrategy) (
	map[string]string, map[string][]constant.DeviceStrategy) {

	nodeLvList := make(map[string]string)
	deviceLvList := make(map[string][]constant.DeviceStrategy)

	if len(relationFault) <= 0 || len(triggerFault) <= 0 || len(strategyList) <= 0 {
		return nodeLvList, deviceLvList
	}
	hwlog.RunLog.Infof("----relationFault: %s, triggerFault: %s, strategyList: %s----",
		util.ObjToString(relationFault), util.ObjToString(triggerFault), util.ObjToString(strategyList))

	// get code device map
	relationCodeDeviceMap, nodeDeviceList, triggerCodeDeviceMap := fJob.transferFaultToMap(relationFault, triggerFault)
	// find all matching configuration tables
	curFaultTables := fJob.getCodeMatchedTables(relationCodeDeviceMap, triggerCodeDeviceMap, strategyList)
	if len(curFaultTables) <= 0 {
		return nodeLvList, deviceLvList
	}
	hwlog.RunLog.Infof("----curFaultTables: %s----", util.ObjToString(curFaultTables))

	// get matched devices
	allUnhealthyDevices, allSubHealthDevices := fJob.getFaultDevices(curFaultTables, relationCodeDeviceMap)
	hwlog.RunLog.Infof("----allUnhealthyDevices: %s, allSubHealthDevices: %s----",
		util.ObjToString(allUnhealthyDevices), util.ObjToString(allSubHealthDevices))

	// get the nodes and devices that isolated or sub-health
	nodeLvList, deviceLvList = fJob.getAllNodeDeviceList(allUnhealthyDevices, allSubHealthDevices, nodeDeviceList)
	hwlog.RunLog.Infof("----nodeLvList: %s, deviceLvList: %s----",
		util.ObjToString(nodeLvList), util.ObjToString(deviceLvList))

	return nodeLvList, deviceLvList
}

func (fJob *FaultJob) updateNodeFaultInfoMap(fault *constant.FaultInfo) {
	fJob.NodeFaultInfoMap[fault.NodeName] = append(fJob.NodeFaultInfoMap[fault.NodeName], fault)
}

func (fJob *FaultJob) getNPUUnderSwitch(faultDevices []*constant.FaultInfo) []constant.DeviceStrategy {
	return make([]constant.DeviceStrategy, 0)
}

func (fJob *FaultJob) getAllNodeDeviceList(allUnhealthyDevices, allSubHealthDevices []*constant.FaultInfo,
	nodeDeviceList map[string]string) (map[string]string, map[string][]constant.DeviceStrategy) {
	// both isolation and sub-health are required, choose the highest level
	// node level higher than device
	nodeLvList, deviceLvList := fJob.handleAllSubHealthyDevices(nodeDeviceList, allSubHealthDevices)

	return fJob.handleAllUnHealthyDevices(allUnhealthyDevices, nodeDeviceList, deviceLvList, nodeLvList)
}

func (fJob *FaultJob) handleAllUnHealthyDevices(allUnhealthyDevices []*constant.FaultInfo,
	nodeDeviceList map[string]string, deviceLvList map[string][]constant.DeviceStrategy,
	nodeLvList map[string]string) (map[string]string, map[string][]constant.DeviceStrategy) {
	if nodeLvList == nil {
		nodeLvList = make(map[string]string)
	}
	if deviceLvList == nil {
		deviceLvList = make(map[string][]constant.DeviceStrategy)
	}
	for nodeName := range nodeDeviceList {
		fJob.handleUnhealthyOnNode(allUnhealthyDevices, deviceLvList, nodeName, nodeLvList)
	}
	return nodeLvList, deviceLvList
}

func (fJob *FaultJob) handleUnhealthyOnNode(allUnhealthyDevices []*constant.FaultInfo,
	deviceLvList map[string][]constant.DeviceStrategy, nodeName string, nodeLvList map[string]string) {
	deviceList := deviceLvList[nodeName]
	nodeLevel := false
	for _, device := range allUnhealthyDevices {
		if device.NodeName != nodeName {
			continue
		}
		if device.FaultType == constant.SwitchFault {
			if fJob.FindNPUUnderSwitch {
				deviceList = append(deviceList, fJob.getNPUUnderSwitch(allUnhealthyDevices)...)
			} else {
				nodeLvList[nodeName] = constant.SeparateFaultStrategy
				nodeLevel = true
			}
		} else {
			idx := fJob.getDeviceIdx(device, deviceList)
			if idx < 0 {
				deviceList = append(deviceList,
					constant.DeviceStrategy{NPUName: device.NPUName, Strategy: constant.SeparateFaultStrategy})
			} else {
				deviceList[idx] = constant.DeviceStrategy{NPUName: device.NPUName,
					Strategy: constant.SeparateFaultStrategy}
			}
		}
		fJob.updateNodeFaultInfoMap(device)
	}
	if !nodeLevel {
		deviceLvList[nodeName] = deviceList
	}
}

func (fJob *FaultJob) handleAllSubHealthyDevices(nodeDeviceList map[string]string,
	allSubHealthDevices []*constant.FaultInfo) (map[string]string, map[string][]constant.DeviceStrategy) {
	nodeLvList := make(map[string]string)
	deviceLvList := make(map[string][]constant.DeviceStrategy)
	for nodeName := range nodeDeviceList {
		fJob.handleSubHealthyOnNode(allSubHealthDevices, nodeName, nodeLvList, deviceLvList)
	}
	return nodeLvList, deviceLvList
}

func (fJob *FaultJob) handleSubHealthyOnNode(allSubHealthDevices []*constant.FaultInfo, nodeName string,
	nodeLvList map[string]string, deviceLvList map[string][]constant.DeviceStrategy) {
	deviceList := make([]constant.DeviceStrategy, 0)
	nodeLevel := false
	for _, device := range allSubHealthDevices {
		if device.NodeName != nodeName {
			continue
		}

		if device.FaultType == constant.SwitchFault {
			if fJob.FindNPUUnderSwitch {
				deviceList = append(deviceList, fJob.getNPUUnderSwitch(allSubHealthDevices)...)
			} else {
				nodeLvList[nodeName] = constant.SubHealthFaultStrategy
				nodeLevel = true
			}
		} else {
			if fJob.getDeviceIdx(device, deviceList) < 0 {
				deviceList = append(deviceList,
					constant.DeviceStrategy{NPUName: device.NPUName, Strategy: constant.SubHealthFaultStrategy})
			}
		}
		fJob.updateNodeFaultInfoMap(device)
	}
	if !nodeLevel {
		deviceLvList[nodeName] = deviceList
	}
}

func (fJob *FaultJob) getDeviceIdx(device *constant.FaultInfo, deviceList []constant.DeviceStrategy) int {
	for i, d := range deviceList {
		if d.NPUName == device.NPUName {
			return i
		}
	}
	return -1
}

func (fJob *FaultJob) transferFaultToMap(relationFault []*constant.FaultInfo,
	triggerFault []constant.FaultInfo) (map[string][]*constant.FaultInfo,
	map[string]string, map[string]constant.FaultInfo) {
	relationCodeDeviceMap := make(map[string][]*constant.FaultInfo)
	nodeList := make(map[string]string)
	for _, device := range relationFault {
		if _, ok := nodeList[device.NodeName]; !ok {
			nodeList[device.NodeName] = device.NodeName
		}
		if _, ok := relationCodeDeviceMap[device.FaultCode]; !ok {
			relationCodeDeviceMap[device.FaultCode] = make([]*constant.FaultInfo, 0)
		}

		relationCodeDeviceMap[device.FaultCode] = append(relationCodeDeviceMap[device.FaultCode], device)
	}

	triggerCodeDeviceMap := make(map[string]constant.FaultInfo)
	for _, trigger := range triggerFault {
		triggerCodeDeviceMap[trigger.FaultCode] = trigger
	}

	return relationCodeDeviceMap, nodeList, triggerCodeDeviceMap
}

func (fJob *FaultJob) getCodeMatchedTables(relationCodeDeviceMap map[string][]*constant.FaultInfo,
	triggerCodeDeviceMap map[string]constant.FaultInfo,
	strategyList []constant.RelationFaultStrategy) []constant.RelationFaultStrategy {
	curFaultTables := make([]constant.RelationFaultStrategy, 0)

	for _, trigger := range strategyList {
		if _, ok := triggerCodeDeviceMap[trigger.TriggerFault]; !ok {
			continue
		}
		if !fJob.matchRelationFaults(relationCodeDeviceMap, trigger) {
			continue
		}

		curFaultTables = append(curFaultTables, trigger)
	}
	return curFaultTables
}

func (fJob *FaultJob) matchRelationFaults(
	relationCodeDeviceMap map[string][]*constant.FaultInfo, trigger constant.RelationFaultStrategy) bool {
	for _, fault := range trigger.RelationFaults {
		if _, ok := relationCodeDeviceMap[fault]; !ok {
			return false
		}
	}
	return true
}

func (fJob *FaultJob) getFaultDevices(curFaultTables []constant.RelationFaultStrategy,
	relationCodeDeviceMap map[string][]*constant.FaultInfo) ([]*constant.FaultInfo, []*constant.FaultInfo) {
	allUnhealthyDevices := make([]*constant.FaultInfo, 0)
	allSubHealthDevices := make([]*constant.FaultInfo, 0)

	for _, configTable := range curFaultTables {
		unhealthyDevices, subHealthDevices := fJob.getMatchedDevices(configTable, relationCodeDeviceMap)
		allUnhealthyDevices = append(allUnhealthyDevices, unhealthyDevices...)
		allSubHealthDevices = append(allSubHealthDevices, subHealthDevices...)
	}

	return allUnhealthyDevices, allSubHealthDevices
}

func (fJob *FaultJob) getMatchedDevices(configTable constant.RelationFaultStrategy,
	relationCodeDeviceMap map[string][]*constant.FaultInfo) ([]*constant.FaultInfo, []*constant.FaultInfo) {
	unhealthyDevices := make([]*constant.FaultInfo, 0)
	subHealthDevices := make([]*constant.FaultInfo, 0)
	for _, relationFault := range configTable.RelationFaults {
		for _, device := range relationCodeDeviceMap[relationFault] {
			if configTable.FaultStrategy == constant.SubHealthFaultStrategy {
				device.ExecutedStrategy = constant.SubHealthFaultStrategy
				subHealthDevices = append(subHealthDevices, device)
			} else if configTable.FaultStrategy == constant.SeparateFaultStrategy {
				device.ExecutedStrategy = constant.SeparateFaultStrategy
				unhealthyDevices = append(unhealthyDevices, device)
			}
		}
	}

	return unhealthyDevices, subHealthDevices
}

func (fJob *FaultJob) processNetworkFault() {
	nodeLvList, deviceLvList := fJob.handleJobFault(fJob.RelationFaults, fJob.TriggerFault, relationFaultStrategies)
	// if fault is cqe error, record link down fault in deal max time
	if cqeTriggerFault := fJob.getCQETriggerFault(); len(cqeTriggerFault) != 0 {
		hwlog.RunLog.Info("deal cqe trigger fault by in deal max time link down fault")
		newNodeLvList, newDeviceLvList :=
			fJob.handleJobFault(fJob.TMOutRelationFaults, cqeTriggerFault, relationFaultStrategies)
		util.MergeStringMapList(nodeLvList, newNodeLvList)
		util.MergeStringMapList(deviceLvList, newDeviceLvList)
	}
	fJob.FaultStrategy = constant.FaultStrategy{
		NodeLvList:   nodeLvList,
		DeviceLvList: deviceLvList,
	}
}

func (fJob *FaultJob) getCQETriggerFault() []constant.FaultInfo {
	cqeTriggerFault := make([]constant.FaultInfo, 0)
	for _, fault := range fJob.TriggerFault {
		if faultdomain.IsCqeFault(fault.FaultCode) {
			cqeTriggerFault = append(cqeTriggerFault, fault)
		}
	}
	return cqeTriggerFault
}
