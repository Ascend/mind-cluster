// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault public fault processor
package publicfault

import (
	"encoding/json"
	"strconv"
	"strings"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/publicfault"
)

// PubFaultProcessor public fault processor
var PubFaultProcessor *pubFaultProcessor

type pubFaultProcessor struct {
	pubFaultInfo map[string]*constant.PubFaultCache
	devCMInfo    *constant.DeviceInfo
}

func init() {
	PubFaultProcessor = &pubFaultProcessor{
		pubFaultInfo: make(map[string]*constant.PubFaultCache),
		devCMInfo:    &constant.DeviceInfo{},
	}
}

// Process public fault process
func (p *pubFaultProcessor) Process(info any) any {
	if publicfault.PubFaultCache == nil || len(publicfault.PubFaultCache.GetPubFault()) == 0 {
		return info
	}
	processContent, ok := info.(constant.OneConfigmapContent[*constant.DeviceInfo])
	if !ok {
		hwlog.RunLog.Error("input is not DeviceInfo type", info)
		return info
	}
	deviceInfos := processContent.AllConfigmap

	copyFaultCache, err := publicfault.PubFaultCache.DeepCopy()
	if err != nil {
		hwlog.RunLog.Errorf("public fault processor process failed, error: %v", err)
		return processContent
	}
	for devCMName, devCMInfo := range deviceInfos {
		nodeName := strings.TrimPrefix(devCMName, constant.DeviceInfoPrefix)
		pubFaults, ok := copyFaultCache[nodeName]
		if !ok {
			continue
		}
		p.pubFaultInfo = pubFaults
		p.devCMInfo = devCMInfo
		p.faultJoin(nodeName)
	}
	processContent.AllConfigmap = deviceInfos
	return processContent
}

func (p *pubFaultProcessor) faultJoin(nodeName string) []constant.DeviceFault {
	faultKey, faultList := faultdomain.GetFaultListInfo(p.devCMInfo)
	var dpCMFaults []constant.DeviceFault
	if err := json.Unmarshal([]byte(faultList), &dpCMFaults); err != nil {
		hwlog.RunLog.Errorf("unmarshal fault list for node <%s> failed, error: %v", nodeName, err)
		return nil
	}
	devType := faultdomain.GetDeviceType(p.devCMInfo)

	var newFaultList []constant.DeviceFault
	if err := util.DeepCopy(&newFaultList, &dpCMFaults); err != nil {
		hwlog.RunLog.Errorf("deep copy device cm faults failed, err: %v", err)
		return nil
	}

	dpNPUFaultLevelMap := make(map[string]string)
	for _, dpCMFault := range newFaultList {
		dpNPUFaultLevelMap[dpCMFault.NPUName] = dpCMFault.FaultLevel
	}

	for _, pubFaultCache := range p.pubFaultInfo {
		// add public fault to fault list
		pubFaultCache.FaultDevNames = convertNPUIdsToName(pubFaultCache.FaultDevIds, devType)
		for _, faultDevName := range pubFaultCache.FaultDevNames {
			newFaultList = append(newFaultList, constant.DeviceFault{
				FaultType:            pubFaultCache.FaultType,
				NPUName:              faultDevName,
				LargeModelFaultLevel: pubFaultCache.FaultLevel,
				FaultLevel:           pubFaultCache.FaultLevel,
				FaultHandling:        pubFaultCache.FaultLevel,
				FaultCode:            pubFaultCache.FaultCode,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					pubFaultCache.FaultCode: {
						FaultTime:  pubFaultCache.FaultTime,
						FaultLevel: pubFaultCache.FaultLevel,
					}},
			})
		}

		for _, pubFaultDev := range pubFaultCache.FaultDevNames {
			// public fault id does not exist in dp cm
			faultLevel, ok := dpNPUFaultLevelMap[pubFaultDev]
			if !ok {
				p.updateAvailAndUnhealthy(pubFaultCache.FaultLevel, pubFaultDev)
				continue
			}
			// public fault id existed in dp cm
			seriousLevel := faultdomain.GetMostSeriousFaultLevel([]string{pubFaultCache.FaultLevel, faultLevel})
			p.updateAvailAndUnhealthy(seriousLevel, pubFaultDev)
		}
	}
	p.updateFaultList(newFaultList, faultKey)
	return newFaultList
}

func (p *pubFaultProcessor) updateFaultList(newFaultList []constant.DeviceFault, faultKey string) {
	faultListData, err := json.Marshal(newFaultList)
	if err != nil {
		hwlog.RunLog.Errorf("marshal device fault list failed, error: %v", err)
		return
	}
	p.devCMInfo.DeviceList[faultKey] = string(faultListData)
}

func convertNPUIdsToName(phyIds []int32, devType string) []string {
	var npuNames []string
	for _, id := range phyIds {
		idStr := strconv.Itoa(int(id))
		npuNames = append(npuNames, devType+"-"+idStr)
	}
	return npuNames
}

func (p *pubFaultProcessor) updateAvailAndUnhealthy(faultLevel string, NPUName string) {
	if faultLevel == constant.SeparateNPU {
		faultdomain.DelDevFromAvailList(p.devCMInfo, []string{NPUName})
		faultdomain.AddDevFromUnhealthyList(p.devCMInfo, []string{NPUName})
	}
}
