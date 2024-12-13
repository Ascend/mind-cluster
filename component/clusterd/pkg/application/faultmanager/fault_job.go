// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"encoding/json"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
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

func (fJob *FaultJob) initFaultJobAttr() {
	if fJob.PodNames == nil {
		fJob.PodNames = make(map[string]string)
	}
	if fJob.SeparateNodes == nil {
		fJob.SeparateNodes = make(sets.String)
	}
	if fJob.ProcessingFaultCode == nil {
		fJob.ProcessingFaultCode = make(sets.String)
	}
	if fJob.PodStrategiesMaps == nil {
		fJob.PodStrategiesMaps = make(map[string]string)
	}
}

func (fJob *FaultJob) process() {
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
		fJob.ProcessingFaultCode.Delete(fault.FaultUid)
	}
	fJob.RelationFaults = networkFaultInfo
	hwlog.RunLog.Debugf("after perstart precess, relation faults is %#v", fJob.RelationFaults)
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
	hwlog.RunLog.Debugf("process strategies is %#v ", newStrategiesMaps)
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
			continue
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
			hwlog.RunLog.Infof("fault code %s is time out, process as separate fault", fault.FaultUid)
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
	if fault.FaultType == deviceFaultType {
		fJob.FaultStrategy.DeviceLvList[fault.NodeName] = append(fJob.FaultStrategy.DeviceLvList[fault.NodeName],
			DeviceStrategy{NPUName: fault.NPUName, Strategy: constant.SeparateFaultStrategy})
	}
}

func (fJob *FaultJob) initFaultJobByDeviceFault(nodeFaultInfo AdvanceDeviceFaultCm, serverList constant.ServerHccl) {
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
			return
		}
		fJob.ProcessingFaultCode.Insert(faultInfo.FaultUid)
		fJob.RelationFaults = append(fJob.RelationFaults, faultInfo)
	}
	if triggerFaultMap.Has(faultInfo.FaultCode) {
		if fJob.IsA3Job && isCqeFault(faultInfo.FaultCode) {
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
