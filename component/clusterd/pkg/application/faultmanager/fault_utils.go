// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"encoding/json"
	"strings"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"huawei.com/npu-exporter/v6/common-utils/utils"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func getFaultCodeTimeOutMap() map[string]int64 {
	return faultCodeTimeOutMap
}

func setFaultCodeTimeOutMap(faultCode string, delTime int64) {
	faultCodeTimeOutMap[faultCode] = delTime
}

func getFaultCodeDelMaxTime(faultCode string) int64 {
	return getFaultCodeTimeOutMap()[faultCode]
}

func cmNameToNodeName(cmName string) string {
	if !strings.HasPrefix(cmName, constant.DeviceInfoPrefix) {
		hwlog.RunLog.Errorf("CmName %s has not prefix %s", cmName, constant.DeviceInfoPrefix)
		return cmName
	}
	return strings.TrimPrefix(cmName, constant.DeviceInfoPrefix)
}

func getAdvanceDeviceCmForNodeMap(deviceInfoCms map[string]*constant.DeviceInfo) map[string]AdvanceDeviceFaultCm {
	advanceDeviceCmForNodeMap := make(map[string]AdvanceDeviceFaultCm)
	for _, deviceInfo := range deviceInfoCms {
		advanceDeviceCmForNodeMap[cmNameToNodeName(deviceInfo.CmName)] = getAdvanceDeviceCm(deviceInfo)
	}
	return advanceDeviceCmForNodeMap
}

// deviceName->faults
func getAdvanceDeviceCm(devInfo *constant.DeviceInfo) AdvanceDeviceFaultCm {
	advanceDeviceCm := AdvanceDeviceFaultCm{
		CmName:      devInfo.CmName,
		SuperPodID:  devInfo.SuperPodID,
		ServerIndex: devInfo.ServerIndex,
		UpdateTime:  devInfo.UpdateTime,
		ServerType:  getServerType(devInfo),
	}
	if faultList, ok := devInfo.DeviceList[getFaultListKey(devInfo)]; ok {
		var devicesFault []constant.DeviceFault
		err := json.Unmarshal([]byte(faultList), &devicesFault)
		if err != nil {
			hwlog.RunLog.Errorf("get fault list for node %v failed. "+
				"Json unmarshall exception: %v", devInfo.CmName, err)
			return advanceDeviceCm
		}
		deviceFaultMap := make(map[string][]constant.DeviceFault)
		for _, deviceFault := range devicesFault {
			if _, ok := deviceFaultMap[deviceFault.NPUName]; !ok {
				deviceFaultMap[deviceFault.NPUName] = make([]constant.DeviceFault, 0)
			}
			hwlog.RunLog.Debugf("device fault: %s of cm %s, time: %s",
				util.ObjToString(deviceFault), devInfo.CmName, util.ReadableMsTime(devInfo.UpdateTime))
			// device plugin may merge multiple fault codes in one string
			deviceFaults := splitDeviceFault(deviceFault, cmNameToNodeName(devInfo.CmName))
			deviceFaultMap[deviceFault.NPUName] = append(deviceFaultMap[deviceFault.NPUName], deviceFaults...)
		}
		advanceDeviceCm.FaultDeviceList = deviceFaultMap
	} else {
		hwlog.RunLog.Infof("get fault list for node %v failed. fault list does not exist", devInfo.CmName)
	}
	if networkUnhealthyCardList, ok := devInfo.DeviceList[getNetworkUnhealthyKey(devInfo)]; ok {
		cardList := strings.Split(networkUnhealthyCardList, ",")
		advanceDeviceCm.NetworkUnhealthy = cardList
	} else {
		hwlog.RunLog.Infof("get NetworkUnhealthy list for node %v failed. fault list does not exist",
			devInfo.CmName)
	}
	if cardUnhealthyCardList, ok := devInfo.DeviceList[getCardUnhealthyKey(devInfo)]; ok {
		var cardList []string
		if len(cardUnhealthyCardList) == 0 {
			cardList = make([]string, 0)
		} else {
			cardList = strings.Split(cardUnhealthyCardList, ",")
		}
		advanceDeviceCm.CardUnHealthy = cardList
	}
	return advanceDeviceCm
}

func getServerType(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, Ascend910Server) {
			return Ascend910Server
		}
		if strings.Contains(key, Ascend310PServer) {
			return Ascend310PServer
		}
		if strings.Contains(key, Ascend310Server) {
			return Ascend310Server
		}
	}
	hwlog.RunLog.Warn("cannot decide server type")
	return Ascend910Server
}

// device plugin may merge multiple fault codes in one string
func splitDeviceFault(faultInfo constant.DeviceFault, nodeName string) []constant.DeviceFault {
	deviceFaults := make([]constant.DeviceFault, 0)
	faultInfo.FaultCode = strings.Replace(faultInfo.FaultCode, " ", "", -1)
	codes := strings.Split(faultInfo.FaultCode, ",")
	for _, code := range codes {
		newFault := constant.DeviceFault{
			FaultType:            faultInfo.FaultType,
			NPUName:              faultInfo.NPUName,
			LargeModelFaultLevel: faultInfo.FaultLevel,
			FaultLevel:           faultInfo.FaultLevel,
			FaultHandling:        faultInfo.FaultLevel,
			FaultCode:            code,
		}
		deviceFaults = append(deviceFaults, newFault)
	}
	return deviceFaults
}

func getFaultListKey(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, "huawei.com/Ascend") && strings.Contains(key, "-Fault") {
			return key
		}
	}
	return ""
}

func getNetworkUnhealthyKey(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, "huawei.com/Ascend") && strings.Contains(key, "-NetworkUnhealthy") {
			return key
		}
	}
	return ""
}

func getCardUnhealthyKey(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, "huawei.com/Ascend") && strings.Contains(key, "-Unhealthy") {
			return key
		}
	}
	return ""
}

func isCqeFault(faultCode string) bool {
	return strings.Contains(faultCode, constant.DevCqeFaultCode) ||
		strings.Contains(faultCode, constant.HostCqeFaultCode)
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
