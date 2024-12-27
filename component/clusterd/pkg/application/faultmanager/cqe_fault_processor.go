// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"strings"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func newLinkDownCqeFaultProcessor(deviceCenter *deviceFaultProcessCenter) *linkDownCqeFaultProcessCenter {
	return &linkDownCqeFaultProcessCenter{
		linkDownCqeFaults: map[string]map[string]map[string]cqeLinkDownFaultRank{},
		cqeFaultTimeList:  map[string][]int64{},
		deviceCenter:      deviceCenter,
	}
}

func (processor *linkDownCqeFaultProcessCenter) process() {
	deviceInfoCms := processor.deviceCenter.getProcessingCm()
	processor.nodeDeviceFaultInfo = getAdvanceDeviceCmForNodeMap(deviceInfoCms) // node deviceInfo
	processor.handleLinkDownCqeFault(processor.deviceCenter.jobServerInfoMap, deviceInfoCms)
	processor.deviceCenter.setProcessingCm(deviceInfoCms)
}

func (processor *linkDownCqeFaultProcessCenter) updateDeviceCmInfo(faultNodeList []string,
	jobDeviceList map[string]map[string]cqeLinkDownFaultRank, deviceInfoCms map[string]*constant.DeviceInfo) {
	for _, nodeName := range faultNodeList {
		cmName := nodeNameToCmName(nodeName)
		deviceInfo, ok := deviceInfoCms[cmName]
		if !ok {
			return
		}
		deviceFaultInfo, ok := processor.nodeDeviceFaultInfo[nodeName]
		if !ok {
			return
		}
		deviceList := jobDeviceList[nodeName]
		cardUnhealthyKey := getCardUnhealthyKey(deviceInfo)
		for _, device := range deviceList {
			if device.IsCqe {
				processor.deleteCqeDevice(nodeName, device.DeviceName, &deviceFaultInfo.CardUnHealthy)
			}
			if device.IsLinkDown {
				processor.addLinkDownDevice(nodeName, device.DeviceName, &deviceFaultInfo.CardUnHealthy)
			}
		}
		deviceInfoCms[cmName].DeviceList[cardUnhealthyKey] = strings.Join(deviceFaultInfo.CardUnHealthy, ",")
		hwlog.RunLog.Infof("nodeName: %s current unhealthy card: %s", nodeName, deviceInfoCms[cmName].DeviceList[cardUnhealthyKey])
	}
}

func (processor *linkDownCqeFaultProcessCenter) addLinkDownDevice(
	nodeName string, deviceName string, deviceList *[]string) {
	deviceIdx := getContainedElementIdx(deviceName, *deviceList)
	if deviceIdx < 0 {
		hwlog.RunLog.Infof("add node: %s linkdown device: %s", nodeName, deviceName)
		*deviceList = append(*deviceList, deviceName)
	}
}

func (processor *linkDownCqeFaultProcessCenter) deleteCqeDevice(
	nodeName string, deviceName string, deviceList *[]string) {
	deviceIdx := getContainedElementIdx(deviceName, *deviceList)
	if deviceIdx >= 0 {
		hwlog.RunLog.Infof("delete node: %s cqe device: %s", nodeName, deviceName)
		*deviceList = append((*deviceList)[:deviceIdx], (*deviceList)[deviceIdx+1:]...)
	}
}

func (processor *linkDownCqeFaultProcessCenter) handleLinkDownCqeFault(
	jobServerInfoMap constant.JobServerInfoMap, deviceInfoCms map[string]*constant.DeviceInfo) {
	// delete the expired jobID
	for jobID, _ := range processor.linkDownCqeFaults {
		if _, ok := jobServerInfoMap.InfoMap[jobID]; !ok {
			delete(processor.linkDownCqeFaults, jobID)
		}
	}
	for jobId, serverList := range jobServerInfoMap.InfoMap {
		if _, ok := processor.linkDownCqeFaults[jobId]; !ok {
			processor.linkDownCqeFaults[jobId] = make(map[string]map[string]cqeLinkDownFaultRank)
		}
		// delete the expired nodeName
		for nodeName, _ := range processor.linkDownCqeFaults[jobId] {
			if _, ok := serverList[nodeName]; !ok {
				delete(processor.linkDownCqeFaults[jobId], nodeName)
			}
		}
		faultNodeList := make([]string, 0)
		jobHasCqe, jobHasLinkDown := false, false
		for nodeName, deviceInfos := range serverList {
			linkDownFaultList, hasCqe, hasLinkDown := processor.getNodeFaultRank(processor.nodeDeviceFaultInfo[nodeName], deviceInfos)
			processor.linkDownCqeFaults[jobId][nodeName] = linkDownFaultList
			if hasCqe {
				jobHasCqe = true
			}
			if hasLinkDown {
				jobHasLinkDown = true
			}
			if hasCqe || hasLinkDown {
				faultNodeList = append(faultNodeList, nodeName)
			}
			hwlog.RunLog.Debugf("nodeName: %s --linkDownFaultList %s--", nodeName, util.ObjToString(linkDownFaultList))
			hwlog.RunLog.Debugf("hasCqe: %t --hasLinkDown %t--", hasCqe, hasLinkDown)
		}
		hwlog.RunLog.Debugf("jobId: %s --jobHasCqe: %t jobHasLinkDown %t--", jobId, jobHasCqe, jobHasLinkDown)
		if jobHasCqe && jobHasLinkDown {
			processor.updateDeviceCmInfo(faultNodeList, processor.linkDownCqeFaults[jobId], deviceInfoCms)
		}
	}
}

func (processor *linkDownCqeFaultProcessCenter) getNodeFaultRank(deviceFaultInfo AdvanceDeviceFaultCm,
	devices constant.ServerHccl) (map[string]cqeLinkDownFaultRank, bool, bool) {
	linkDownCqeFaultList := make(map[string]cqeLinkDownFaultRank)
	hasCqe, hasLinkDown := false, false
	if len(devices.DeviceList) == 0 {
		return linkDownCqeFaultList, false, false
	}

	for _, deviceInfo := range devices.DeviceList {
		deviceName := deviceFaultInfo.ServerType + "-" + deviceInfo.DeviceID
		faultList, ok := deviceFaultInfo.FaultDeviceList[deviceName]
		if !ok {
			continue
		}
		for _, fault := range faultList {
			cqeLinkDownFault, ok := linkDownCqeFaultList[deviceName]
			if !ok {
				cqeLinkDownFault = cqeLinkDownFaultRank{
					DeviceName: deviceName,
				}
			}
			if isCqeFault(fault) {
				cqeLinkDownFault.IsCqe = true
				hasCqe = true
			}
			if isLinkDownFault(fault) {
				cqeLinkDownFault.IsLinkDown = true
				cqeLinkDownFault.LinkDownFaultTime = fault.FaultTimeAndLevelMap[fault.FaultCode].FaultTime
				hasLinkDown = true
			}
			linkDownCqeFaultList[deviceName] = cqeLinkDownFault
		}
	}
	return linkDownCqeFaultList, hasCqe, hasLinkDown
}
