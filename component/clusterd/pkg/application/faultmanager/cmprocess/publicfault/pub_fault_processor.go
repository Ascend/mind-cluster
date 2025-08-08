// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault public fault processor
package publicfault

import (
	"strconv"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/publicfault"
)

// PubFaultProcessor public fault processor
var PubFaultProcessor *pubFaultProcessor

type pubFaultProcessor struct {
	pubFaultInfo map[string]*constant.PubFaultCache
	devCMInfo    *constant.AdvanceDeviceFaultCm
}

func init() {
	PubFaultProcessor = &pubFaultProcessor{
		pubFaultInfo: make(map[string]*constant.PubFaultCache),
		devCMInfo:    &constant.AdvanceDeviceFaultCm{},
	}
}

// Process public fault process
func (p *pubFaultProcessor) Process(info any) any {
	if publicfault.PubFaultCache == nil || len(publicfault.PubFaultCache.GetPubFault()) == 0 {
		return info
	}
	processContent, ok := info.(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
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
	for nodeName, devCMInfo := range deviceInfos {
		pubFaults, ok := copyFaultCache[nodeName]
		if !ok {
			continue
		}
		p.pubFaultInfo = pubFaults
		p.devCMInfo = devCMInfo
		p.faultJoin()
	}
	processContent.AllConfigmap = deviceInfos
	return processContent
}

func (p *pubFaultProcessor) faultJoin() {
	modified := false
	for _, pubFaultCache := range p.pubFaultInfo {
		// add public fault to fault list
		pubFaultCache.FaultDevNames = convertNPUIdsToName(pubFaultCache.FaultDevIds, p.devCMInfo.DeviceType)
		for _, faultDevName := range pubFaultCache.FaultDevNames {
			fault := constant.DeviceFault{
				FaultType:            constant.PublicFaultType,
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
			}
			p.devCMInfo.AddFaultAndFix(fault)
			modified = true
		}
	}
	if modified {
		p.devCMInfo.UpdateTime = time.Now().Unix()
		faultdomain.SortDataForAdvanceDeviceInfo(p.devCMInfo)
	}
}

func convertNPUIdsToName(phyIds []int32, devType string) []string {
	var npuNames []string
	for _, id := range phyIds {
		idStr := strconv.Itoa(int(id))
		npuNames = append(npuNames, devType+"-"+idStr)
	}
	return npuNames
}
