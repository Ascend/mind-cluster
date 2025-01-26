// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package device a series of device function
package device

import (
	"encoding/json"
	"fmt"

	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const safeDeviceSize = 1000

// ParseDeviceInfoCM get device info from configmap obj
func ParseDeviceInfoCM(obj interface{}) (*constant.DeviceInfo, error) {
	deviceCm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return &constant.DeviceInfo{}, fmt.Errorf("not device configmap")
	}
	devInfoCM := constant.DeviceInfoCM{}
	data, ok := deviceCm.Data[constant.DevInfoCMKey]
	if !ok {
		return &constant.DeviceInfo{}, fmt.Errorf("configmap %s has no %s", deviceCm.Name, constant.DevInfoCMKey)
	}

	if unmarshalErr := json.Unmarshal([]byte(data), &devInfoCM); unmarshalErr != nil {
		return &constant.DeviceInfo{}, fmt.Errorf("unmarshal failed: %v, configmap name: %s", unmarshalErr, deviceCm.Name)
	}

	if !util.EqualDataHash(devInfoCM.CheckCode, devInfoCM.DeviceInfo) {
		return &constant.DeviceInfo{}, fmt.Errorf("device info configmap %s is not valid", deviceCm.Name)
	}
	var device constant.DeviceInfo
	device.DeviceList = devInfoCM.DeviceInfo.DeviceList
	device.UpdateTime = devInfoCM.DeviceInfo.UpdateTime
	device.ServerIndex = devInfoCM.ServerIndex
	device.SuperPodID = devInfoCM.SuperPodID
	device.CmName = deviceCm.Name
	return &device, nil
}

// DeepCopy deep copy deviceInfo
func DeepCopy(info *constant.DeviceInfo) *constant.DeviceInfo {
	if info == nil {
		return nil
	}
	data, err := json.Marshal(info)
	if err != nil {
		hwlog.RunLog.Errorf("marshal device failed , err is %v", err)
		return nil
	}
	newDeviceInfo := &constant.DeviceInfo{}
	if err := json.Unmarshal(data, newDeviceInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal device failed , err is %v", err)
		return nil
	}
	return newDeviceInfo
}

// DeepCopyInfos deep copy deviceInfos
func DeepCopyInfos(infos map[string]*constant.DeviceInfo) map[string]*constant.DeviceInfo {
	res := make(map[string]*constant.DeviceInfo)
	for key, val := range infos {
		res[key] = DeepCopy(val)
	}
	return res
}

// GetSafeData get data every 1000 DeviceInfo
func GetSafeData(deviceInfos map[string]*constant.DeviceInfo) []string {
	if len(deviceInfos) == 0 {
		return []string{}
	}
	if len(deviceInfos) <= safeDeviceSize {
		return []string{util.ObjToString(deviceInfos)}
	}
	deviceSlice := make([]string, 0, len(deviceInfos)/safeDeviceSize+1)
	childDeviceInfos := make(map[string]*constant.DeviceInfo, safeDeviceSize)
	for cmName, deviceInfo := range deviceInfos {
		childDeviceInfos[cmName] = deviceInfo
		if len(childDeviceInfos)%safeDeviceSize == 0 {
			deviceSlice = append(deviceSlice, util.ObjToString(childDeviceInfos))
			childDeviceInfos = make(map[string]*constant.DeviceInfo, safeDeviceSize)
		}
	}
	if len(childDeviceInfos) != 0 {
		deviceSlice = append(deviceSlice, util.ObjToString(childDeviceInfos))
	}
	return deviceSlice
}
