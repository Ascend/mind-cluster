// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func (fJob *FaultJob) handleJobFault(relationFault []*faultInfo,
	triggerFault []faultInfo, strategyList []RelationFaultStrategy) (map[string]string, map[string][]DeviceStrategy) {

	nodeLvList := make(map[string]string)
	deviceLvList := make(map[string][]DeviceStrategy)

	if len(relationFault) <= 0 || len(triggerFault) <= 0 && len(strategyList) <= 0 {
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

func (fJob *FaultJob) getNPUUnderSwitch(faultDevices []*faultInfo) []DeviceStrategy {
	return make([]DeviceStrategy, 0)
}

func (fJob *FaultJob) getAllNodeDeviceList(allUnhealthyDevices, allSubHealthDevices []*faultInfo,
	nodeDeviceList map[string]string) (map[string]string, map[string][]DeviceStrategy) {
	nodeLvList := make(map[string]string)
	deviceLvList := make(map[string][]DeviceStrategy)

	// both isolation and sub-health are required, choose the highest level
	// node level higher than device
	for nodeName := range nodeDeviceList {
		deviceList := make([]DeviceStrategy, 0)
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
					deviceList = append(deviceList, DeviceStrategy{NPUName: device.NPUName, Strategy: constant.SubHealthFaultStrategy})
				}
			}
		}
		if !nodeLevel {
			deviceLvList[nodeName] = deviceList
		}
	}

	for nodeName := range nodeDeviceList {
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
					deviceList = append(deviceList, DeviceStrategy{NPUName: device.NPUName, Strategy: constant.SeparateFaultStrategy})
				} else {
					deviceList[idx] = DeviceStrategy{NPUName: device.NPUName, Strategy: constant.SeparateFaultStrategy}
				}
			}
		}
		if !nodeLevel {
			deviceLvList[nodeName] = deviceList
		}
	}
	return nodeLvList, deviceLvList
}

func (fJob *FaultJob) getDeviceIdx(device *faultInfo, deviceList []DeviceStrategy) int {
	for i, d := range deviceList {
		if d.NPUName == device.NPUName {
			return i
		}
	}
	return -1
}

func (fJob *FaultJob) transferFaultToMap(relationFault []*faultInfo,
	triggerFault []faultInfo) (map[string][]*faultInfo, map[string]string, map[string]faultInfo) {
	relationCodeDeviceMap := make(map[string][]*faultInfo)
	nodeList := make(map[string]string)
	for _, device := range relationFault {
		if _, ok := nodeList[device.NodeName]; !ok {
			nodeList[device.NodeName] = device.NodeName
		}
		if _, ok := relationCodeDeviceMap[device.FaultCode]; !ok {
			relationCodeDeviceMap[device.FaultCode] = make([]*faultInfo, 0)
		}

		relationCodeDeviceMap[device.FaultCode] = append(relationCodeDeviceMap[device.FaultCode], device)
	}

	triggerCodeDeviceMap := make(map[string]faultInfo)
	for _, trigger := range triggerFault {
		triggerCodeDeviceMap[trigger.FaultCode] = trigger
	}

	return relationCodeDeviceMap, nodeList, triggerCodeDeviceMap
}

func (fJob *FaultJob) getCodeMatchedTables(relationCodeDeviceMap map[string][]*faultInfo,
	triggerCodeDeviceMap map[string]faultInfo, strategyList []RelationFaultStrategy) []RelationFaultStrategy {
	curFaultTables := make([]RelationFaultStrategy, 0)

	for _, trigger := range strategyList {
		if _, ok := triggerCodeDeviceMap[trigger.TriggerFault]; !ok {
			continue
		}
		if !fJob.relationFaultsMatched(relationCodeDeviceMap, trigger) {
			continue
		}

		curFaultTables = append(curFaultTables, trigger)
	}
	return curFaultTables
}

func (fJob *FaultJob) relationFaultsMatched(relationCodeDeviceMap map[string][]*faultInfo, trigger RelationFaultStrategy) bool {
	for _, fault := range trigger.RelationFaults {
		if _, ok := relationCodeDeviceMap[fault]; !ok {
			return false
		}
	}
	return true
}

func (fJob *FaultJob) getFaultDevices(curFaultTables []RelationFaultStrategy,
	relationCodeDeviceMap map[string][]*faultInfo) ([]*faultInfo, []*faultInfo) {
	allUnhealthyDevices := make([]*faultInfo, 0)
	allSubHealthDevices := make([]*faultInfo, 0)

	for _, configTable := range curFaultTables {
		unhealthyDevices, subHealthDevices := fJob.getMatchedDevices(configTable, relationCodeDeviceMap)
		allUnhealthyDevices = append(allUnhealthyDevices, unhealthyDevices...)
		allSubHealthDevices = append(allSubHealthDevices, subHealthDevices...)
	}

	return allUnhealthyDevices, allSubHealthDevices
}

func (fJob *FaultJob) getMatchedDevices(configTable RelationFaultStrategy,
	relationCodeDeviceMap map[string][]*faultInfo) ([]*faultInfo, []*faultInfo) {
	unhealthyDevices := make([]*faultInfo, 0)
	subHealthDevices := make([]*faultInfo, 0)
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

	fJob.FaultStrategy = FaultStrategy{
		NodeLvList:   nodeLvList,
		DeviceLvList: deviceLvList,
	}
}
