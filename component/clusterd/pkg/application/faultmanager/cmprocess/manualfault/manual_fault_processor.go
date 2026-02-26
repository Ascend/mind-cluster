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

// Package manualfault is used to process manually separate faults

package manualfault

import (
	"fmt"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/manualfault"
	"clusterd/pkg/domain/pod"
)

// ManualFaultProcessor is used to process manually separate faults
var ManualFaultProcessor *manualFaultProcessor

type manualFaultProcessor struct {
	nodeDeviceCmMap map[string]*constant.AdvanceDeviceFaultCm
}

func init() {
	ManualFaultProcessor = &manualFaultProcessor{
		nodeDeviceCmMap: make(map[string]*constant.AdvanceDeviceFaultCm),
	}
}

// Process is used to process manually separate faults
func (p *manualFaultProcessor) Process(info any) any {
	if !conf.GetManualEnabled() {
		manualfault.InitJobFaultManager(constant.DefaultSlidingWindow)
		manualfault.InitCounter()
		manualfault.InitFaultCmInfo()
		return info
	}

	processContent, ok := info.(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	if !ok {
		hwlog.RunLog.Error("input is not deviceinfo type")
		return info
	}
	devInfos := processContent.AllConfigmap
	p.loadManualSepToClusterInfoCm(devInfos)

	for nodeName, devCmInfo := range devInfos {
		// only care about occur, ignore recover
		if len(devCmInfo.FaultDeviceList) == 0 {
			continue
		}
		oldDevInfo, ok := p.nodeDeviceCmMap[nodeName]
		if !ok {
			p.dealIncrementFaultForNode(nodeName, devCmInfo.FaultDeviceList, devCmInfo.DeviceType)
			continue
		}
		if oldDevInfo.IsSame(devCmInfo) {
			// fault have no changed, ignore
			continue
		}
		incrementFault := p.getIncrementFault(oldDevInfo.FaultDeviceList, devCmInfo.FaultDeviceList)
		if len(incrementFault) == 0 {
			// no increment fault, ignore
			continue
		}
		p.dealIncrementFaultForNode(nodeName, incrementFault, devCmInfo.DeviceType)
	}
	p.nodeDeviceCmMap = processContent.AllConfigmap
	return info
}

func (p *manualFaultProcessor) loadManualSepToClusterInfoCm(devInfos map[string]*constant.AdvanceDeviceFaultCm) {
	nodeInfo, err := manualfault.FaultCmInfo.DeepCopy()
	if err != nil {
		hwlog.RunLog.Errorf("deep copy fault cm info failed, error: %v", err)
		return
	}

	if len(nodeInfo) == 0 {
		return
	}
	for nodeName, devFaults := range nodeInfo {
		for _, devName := range devFaults.Total {
			hwlog.RunLog.Debugf("node: %s, dev: %s has already been manually separated", nodeName, devName)
			devCmInfo, ok := devInfos[nodeName]
			if !ok {
				continue
			}
			fault := constant.DeviceFault{
				FaultType:            constant.CardUnhealthy,
				NPUName:              devName,
				LargeModelFaultLevel: constant.ManuallySeparateNPU,
				FaultLevel:           constant.ManuallySeparateNPU,
				FaultHandling:        constant.ManuallySeparateNPU,
			}
			devCmInfo.AddFaultAndFix(fault)
		}
	}
	hwlog.RunLog.Debug("load manual fault cache to cluster info cm success")
}

func (p *manualFaultProcessor) dealIncrementFaultForNode(nodeName string,
	incrementFault map[string][]constant.DeviceFault, devType string) {
	devNameJobMap := createDevNameJobMap(nodeName, devType)
	for devName, faults := range incrementFault {
		jobId := getJobIdByDev(devNameJobMap, devName)
		for _, fault := range faults {
			for code, level := range fault.FaultTimeAndLevelMap {
				// for dp isolated faults, clusterd counts normally.
				// dp isolated fault: code is fault code, level is ManuallySeparateNPU
				// clusterd isolated fault: code and level are both ManuallySeparateNPU
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
	}
}

func createDevNameJobMap(nodeName string, devType string) map[string]string {
	devJobMap := make(map[string]string)
	pods, exist := pod.GetPodsByNodeName(nodeName)
	if !exist {
		return nil
	}
	for _, podInfo := range pods {
		jobId := pod.GetJobKeyByPod(&podInfo)
		if jobId == "" {
			continue
		}
		usedDevs := pod.GetPodUsedDev(podInfo)
		for _, devId := range usedDevs {
			devName := convertDevIdToName(devId, devType)
			devJobMap[devName] = jobId
		}
	}
	return devJobMap
}

func getJobIdByDev(devJobMap map[string]string, devName string) string {
	if len(devJobMap) == 0 {
		return ""
	}
	jobId, ok := devJobMap[devName]
	if !ok {
		return ""
	}
	return jobId
}

func convertDevIdToName(id string, devType string) string {
	return fmt.Sprintf("%s-%s", devType, id)
}

// getIncrementFault newInfo have, but oldInfo not have
func (p *manualFaultProcessor) getIncrementFault(oldInfo, newInfo map[string][]constant.DeviceFault) map[string][]constant.DeviceFault {
	if len(oldInfo) == 0 {
		return newInfo
	}
	if len(newInfo) == 0 {
		return make(map[string][]constant.DeviceFault)
	}
	incrementFault := make(map[string][]constant.DeviceFault)
	for devName, faults := range newInfo {
		oldFaults, ok := oldInfo[devName]
		if !ok {
			incrementFault[devName] = faults
			continue
		}
		for _, fault := range faults {
			if isContainFault(fault, oldFaults) {
				// not increment
				continue
			}
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
			if ok && newLevel.FaultReceivedTime == level.FaultReceivedTime {
				isContain = true
				break
			}
		}
	}
	return isContain
}
