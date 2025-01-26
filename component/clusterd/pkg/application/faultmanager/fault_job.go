// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"encoding/json"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/interface/kube"
)

var relationFaultStrategies = make([]RelationFaultStrategy, 0)
var faultDurationStrategies = make([]FaultDuration, 0)
var relationFaultTypeMap = make(sets.String)
var triggerFaultMap = make(sets.String)
var faultCodeTimeOutMap = make(map[string]int64)

func init() {
	if fileBytes := LoadConfigFromFile(faultCustomizationPath); fileBytes != nil {
		initRelationFaultStrategies(fileBytes)
		initRelationFaultCodesMap()
	}
	if fileBytes := LoadConfigFromFile(faultDuration); fileBytes != nil {
		initFaultDuration(fileBytes)
		initFaultCodeTimeOutMap()
	}
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
		hwlog.RunLog.Errorf("unmarshal fault code byte failed: %v", err)
		return
	}
}

func initFaultDuration(fileBytes []byte) {
	var tmpFaultDurationStrategies []FaultDuration
	if err := json.Unmarshal(fileBytes, &tmpFaultDurationStrategies); err != nil {
		hwlog.RunLog.Errorf("unmarshal fault code byte failed: %v", err)
		return
	}
	if len(tmpFaultDurationStrategies) == 0 {
		hwlog.RunLog.Error("fault duration fault config is invalid")
		return
	}
	for _, faultConfig := range tmpFaultDurationStrategies {
		if !validateFaultDurationConfig(faultConfig) {
			continue
		}
		faultDurationStrategies = append(faultDurationStrategies, faultConfig)
	}
}

func validateFaultDurationConfig(faultConfig FaultDuration) bool {
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
	fJob.FaultStrategy = FaultStrategy{}
	fJob.TriggerFault = nil
	fJob.AllFaultCode = make(sets.String)
	fJob.SeparateNodes = make(sets.String)
	if fJob.PodNames == nil {
		fJob.PodNames = make(map[string]string)
	}
	if fJob.ProcessingFaultCode == nil {
		fJob.ProcessingFaultCode = make(sets.String)
	}
	if fJob.PodStrategiesMaps == nil {
		fJob.PodStrategiesMaps = make(map[string]string)
	}
}

func (fJob *FaultJob) Process() {
	fJob.preStartProcess()
	fJob.processNetworkFault()
	fJob.preStopProcess()
}

func (fJob *FaultJob) preStartProcess() {
	var networkFaultInfo []*faultInfo
	for _, fault := range fJob.RelationFaults {
		if fJob.AllFaultCode.Has(fault.FaultUid) {
			networkFaultInfo = append(networkFaultInfo, fault)
			continue
		}
		hwlog.RunLog.Infof("fault code is not exist %v, delete it in ProcessingFaultCode", fault.FaultUid)
		fJob.ProcessingFaultCode.Delete(fault.FaultUid)
	}
	fJob.RelationFaults = networkFaultInfo
	hwlog.RunLog.Debugf("after perstart precess, relation faults is %v", util.ObjToString(fJob.RelationFaults))
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
		if err := kube.RetryPatchPodLabels(podName, fJob.NameSpace, patchPodTimes,
			map[string]string{taskFaultKey: strategy}); err != nil {
			hwlog.RunLog.Errorf("patch pod label failed: %v", err)
		}
	}
	fJob.PodStrategiesMaps = podStrategiesMaps
}

func (fJob *FaultJob) clearProcessedAndTimeOutFault() {
	var networkFaultInfo []*faultInfo
	preStopTime := time.Now().UnixMilli()
	for _, fault := range fJob.RelationFaults {
		if fault.ExecutedStrategy == constant.SeparateFaultStrategy {
			fJob.ProcessingFaultCode.Delete(fault.FaultUid)
			continue
		}
		if preStopTime-fault.FaultTime >= fault.DealMaxTime*kilo {
			hwlog.RunLog.Infof("fault code %s is time out, process as default strategy", fault.FaultUid)
			fJob.addFaultStrategyForTimeOutCode(fault)
			continue
		}
		networkFaultInfo = append(networkFaultInfo, fault)
	}
	fJob.RelationFaults = networkFaultInfo
}

func (fJob *FaultJob) addFaultStrategyForTimeOutCode(fault *faultInfo) {
	if fault.ExecutedStrategy != "" {
		return
	}
	if fault.FaultType == switchFaultType {
		fJob.FaultStrategy.NodeLvList[fault.NodeName] = constant.SubHealthFaultStrategy
	}
}

func (fJob *FaultJob) initFaultJobByDeviceFault(nodeFaultInfo constant.AdvanceDeviceFaultCm, serverList constant.ServerHccl) {
	if fJob.SeparateNodes.Has(serverList.ServerName) {
		return
	}
	for _, deviceInfo := range serverList.DeviceList {
		deviceName := nodeFaultInfo.ServerType + "-" + deviceInfo.DeviceID
		fault, ok := nodeFaultInfo.FaultDeviceList[deviceName]
		if !ok {
			continue
		}
		fJob.initFaultInfoByDeviceFault(fault, serverList.ServerName, deviceInfo.RankID,
			util.IsSliceContain(deviceName, nodeFaultInfo.CardUnHealthy))
	}
}

func (fJob *FaultJob) initFaultInfoByDeviceFault(faultList []constant.DeviceFault, nodeName, rankId string, isCardUnhealthy bool) {
	for _, fault := range faultList {
		for faultCode, faultTimeAndLevel := range fault.FaultTimeAndLevelMap {
			if isAssociateFault(faultCode) && !isCardUnhealthy {
				tmpFaultInfo := faultInfo{
					NodeName:    nodeName,
					FaultType:   deviceFaultType,
					NPUName:     fault.NPUName,
					FaultCode:   faultCode,
					FaultLevel:  faultTimeAndLevel.FaultLevel,
					FaultTime:   time.Now().UnixMilli(),
					DealMaxTime: getFaultCodeDelMaxTime(faultCode),
					FaultUid:    nodeName + "-" + fault.NPUName + "-" + faultCode,
				}
				fJob.AllFaultCode.Insert(tmpFaultInfo.FaultUid)
				fJob.addFaultInfoByCodeType(&tmpFaultInfo)
			}
		}
	}
}

func (fJob *FaultJob) addFaultInfoByCodeType(faultInfo *faultInfo) {
	if relationFaultTypeMap.Has(faultInfo.FaultCode) {
		if fJob.ProcessingFaultCode.Has(faultInfo.FaultUid) {
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
			return
		}
		fJob.TriggerFault = append(fJob.TriggerFault, *faultInfo)
	}
}

func isAssociateFault(faultCode string) bool {
	return relationFaultTypeMap.Has(faultCode) || triggerFaultMap.Has(faultCode)
}

func (fJob *FaultJob) initFaultJobBySwitchFault(switchInfo *constant.SwitchInfo, serverList constant.ServerHccl) {
	if switchInfo == nil {
		return
	}
	if switchInfo.NodeStatus == nodeUnhealthy {
		fJob.SeparateNodes.Insert(serverList.ServerName)
		return
	}
	for _, fCode := range switchInfo.FaultCode {
		var tmpSwitchFaultInfo simpleSwitchFaultInfo
		if err := json.Unmarshal([]byte(fCode), &tmpSwitchFaultInfo); err != nil {
			hwlog.RunLog.Errorf("unmarshal switch faultinfo failed:%v", err)
			continue
		}
		if isAssociateFault(tmpSwitchFaultInfo.AssembledFaultCode) {
			tmpFaultInfo := faultInfo{
				NodeName:    serverList.ServerName,
				NPUName:     allCardId,
				FaultType:   switchFaultType,
				FaultCode:   tmpSwitchFaultInfo.AssembledFaultCode,
				FaultTime:   time.Now().UnixMilli(),
				DealMaxTime: getFaultCodeDelMaxTime(tmpSwitchFaultInfo.AssembledFaultCode),
				FaultLevel:  switchInfo.FaultLevel,
				FaultUid:    serverList.ServerName + "-" + allCardId + "-" + tmpSwitchFaultInfo.AssembledFaultCode,
			}
			fJob.AllFaultCode.Insert(tmpFaultInfo.FaultUid)
			fJob.addFaultInfoByCodeType(&tmpFaultInfo)
		}
	}
}
