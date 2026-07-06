// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package device a series of device function
package device

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const (
	// maxCmDataSize is the max data size for a single ConfigMap (~1MB limit, using 800KB for safety margin)
	maxCmDataSize = 800 * 1024
)

// ParseDeviceInfoCM get device info from configmap obj
func ParseDeviceInfoCM(deviceCm *v1.ConfigMap) (*constant.DeviceInfo, error) {
	devInfoCM := constant.DeviceInfoCM{}
	data, ok := deviceCm.Data[api.DeviceInfoCMDataKey]
	if !ok {
		return &constant.DeviceInfo{}, fmt.Errorf("configmap %s has no %s", deviceCm.Name, api.DeviceInfoCMDataKey)
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
	device.RackID = devInfoCM.RackID
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

// GetSafeData splits deviceInfos into chunks that fit within K8s ConfigMap size limit (~1MB).
// Each chunk is as close to maxCmDataSize (800KB) as possible.
func GetSafeData(deviceInfos map[string]*constant.DeviceInfo) []string {
	return util.SplitMapToSafeChunks(deviceInfos, maxCmDataSize,
		func(m map[string]*constant.DeviceInfo) string {
			return util.ObjToString(m)
		})
}
