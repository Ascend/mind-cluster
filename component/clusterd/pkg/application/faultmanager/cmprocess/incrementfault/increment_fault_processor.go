/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
// Package incrementfault is used to process increment faults

package incrementfault

import (
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/manualfault"
	"clusterd/pkg/domain/pod"
)

// IncrementFaultProcessor is used to process increment faults
var IncrementFaultProcessor *incrementFaultProcessor

type incrementFaultProcessor struct {
	nodeDeviceCmMap map[string]*constant.AdvanceDeviceFaultCm
}

func init() {
	IncrementFaultProcessor = &incrementFaultProcessor{
		nodeDeviceCmMap: make(map[string]*constant.AdvanceDeviceFaultCm),
	}
}

// Process is used to process manually separate faults
func (p *incrementFaultProcessor) Process(info any) any {
	if !conf.GetManualEnabled() {
		return info
	}

	processContent, ok := info.(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	if !ok {
		hwlog.RunLog.Error("input is not deviceinfo type")
		return info
	}
	devInfos := processContent.AllConfigmap
	for nodeName, devCmInfo := range devInfos {
		// only care about occur, ignore recover
		if len(devCmInfo.FaultDeviceList) == 0 {
			continue
		}
		oldDevInfo, ok := p.nodeDeviceCmMap[nodeName]
		if !ok {
			p.dealIncrementFaultForNode(nodeName, devCmInfo.FaultDeviceList, devCmInfo.DeviceType)
			hwlog.RunLog.Debugf("first process node<%s>, treat current faults as increment faults", nodeName)
			continue
		}
		if oldDevInfo.IsSame(devCmInfo) {
			// fault has no change, ignore
			continue
		}
		incrementFault := p.getIncrementFault(oldDevInfo.FaultDeviceList, devCmInfo.FaultDeviceList)
		if len(incrementFault) == 0 {
			// no increment fault, ignore
			hwlog.RunLog.Debugf("no increment fault for node<%s>, ignore", nodeName)
			continue
		}
		p.dealIncrementFaultForNode(nodeName, incrementFault, devCmInfo.DeviceType)
	}
	p.nodeDeviceCmMap = processContent.AllConfigmap
	return info
}

func (p *incrementFaultProcessor) dealIncrementFaultForNode(nodeName string,
	incrementFault map[string][]constant.DeviceFault, devType string) {
	devNameJobMap := pod.CreateDevNameJobMap(nodeName, devType)
	hwlog.RunLog.Debugf("node<%s> increment fault: %v", nodeName, incrementFault)
	for devName, faults := range incrementFault {
		jobId := pod.GetJobIdByDev(devNameJobMap, devName)
		for _, fault := range faults {
			addForManuallyFault(nodeName, devName, jobId, fault.FaultTimeAndLevelMap)
		}
	}
}

func addForManuallyFault(nodeName, devName, jobId string, faultTimeAndLevelMap map[string]constant.FaultTimeAndLevel) {
	for code, level := range faultTimeAndLevelMap {
		// for dp manually separate faults, clusterd counts normally.
		// 1. dp manually separate fault:
		// if fault does not disappear: code is fault code, level is ManuallySeparateNPU
		// if fault disappear: code and level are both ManuallySeparateNPU
		// 2. clusterd manually separate fault: code and level are both ManuallySeparateNPU
		if level.FaultLevel == constant.NotHandleFault || level.FaultLevel == constant.SubHealthFault ||
			(level.FaultLevel == constant.ManuallySeparateNPU && code == constant.ManuallySeparateNPU) {
			continue
		}
		faultInfo := &manualfault.Fault{
			Code:        code,
			JobId:       jobId,
			NodeName:    nodeName,
			DevName:     devName,
			ReceiveTime: level.FaultReceivedTime,
		}
		manualfault.JobFaultMgr.AddFault(faultInfo)
	}
}

// getIncrementFault newInfo have, but oldInfo not have
func (p *incrementFaultProcessor) getIncrementFault(oldInfo, newInfo map[string][]constant.DeviceFault) map[string][]constant.DeviceFault {
	if len(newInfo) == 0 {
		return make(map[string][]constant.DeviceFault)
	}
	incrementFault := make(map[string][]constant.DeviceFault)
	for devName, faults := range newInfo {
		oldFaults, ok := oldInfo[devName]
		if !ok {
			incrementFault[devName] = append([]constant.DeviceFault(nil), faults...)
			continue
		}
		for _, fault := range faults {
			if isContainFault(fault, oldFaults) {
				// not increment
				continue
			}
			// notice: if the device fault info needs to be modified,
			// remember to check if the structure of the device fault info contains a reference type
			incrementFault[devName] = append(incrementFault[devName], fault)
		}
	}
	return incrementFault
}

// isContainFault whether newFault contains in oldFaults. if contains, not increment; if not contain, increment.
func isContainFault(newFault constant.DeviceFault, oldFaults []constant.DeviceFault) bool {
	var isContain bool
	for _, oldFault := range oldFaults {
		for code, level := range oldFault.FaultTimeAndLevelMap {
			newLevel, ok := newFault.FaultTimeAndLevelMap[code]
			if ok && newLevel.FaultReceivedTime == level.FaultReceivedTime && newLevel.FaultLevel == level.FaultLevel {
				isContain = true
				break
			}
		}
	}
	return isContain
}
