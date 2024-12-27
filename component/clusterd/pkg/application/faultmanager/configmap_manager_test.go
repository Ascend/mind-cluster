// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"reflect"
	"sync"
	"testing"

	"clusterd/pkg/common/constant"
)

func TestFaultCenterCmManagerSetAndGetConfigmap(t *testing.T) {
	t.Run("", func(t *testing.T) {
		deviceCM := configMap[*constant.DeviceInfo]{configmap: make(map[string]*constant.DeviceInfo)}
		cm1 := &constant.DeviceInfo{
			DeviceInfoNoName: constant.DeviceInfoNoName{},
			CmName:           "cm1",
		}
		deviceCM.updateCmInfo(cm1, true)
		cm2 := &constant.DeviceInfo{
			DeviceInfoNoName: constant.DeviceInfoNoName{},
			CmName:           "cm2",
		}
		deviceCM.updateCmInfo(cm2, true)
		if !reflect.DeepEqual(deviceCM.configmap["cm1"], cm1) ||
			!reflect.DeepEqual(deviceCM.configmap["cm2"], cm2) {
			t.Errorf("TestFaultCenterCmManagerSetAndGetDeviceInfoCm failed")
		}
		faultManager := faultCenterCmManager[*constant.DeviceInfo]{
			mutex:        sync.RWMutex{},
			originalCm:   configMap[*constant.DeviceInfo]{},
			processingCm: configMap[*constant.DeviceInfo]{},
			processedCm:  configMap[*constant.DeviceInfo]{},
		}
		faultManager.updateOriginalCm(cm1, true)
		faultManager.updateOriginalCm(cm2, true)
		if !reflect.DeepEqual(deviceCM, faultManager.getOriginalCm()) {
			t.Errorf("TestFaultCenterCmManagerSetAndGetDeviceInfoCm failed")
		}
		faultManager.setProcessingCm(faultManager.getOriginalCm())
		if !reflect.DeepEqual(deviceCM, faultManager.getProcessingCm()) {
			t.Errorf("TestFaultCenterCmManagerSetAndGetDeviceInfoCm failed")
		}
		faultManager.setProcessedCm(faultManager.getProcessingCm())
		if !reflect.DeepEqual(deviceCM, faultManager.getProcessedCm()) {
			t.Errorf("TestFaultCenterCmManagerSetAndGetDeviceInfoCm failed")
		}
		faultManager.updateOriginalCm(cm1, false)
		faultManager.updateOriginalCm(cm2, false)
		if len(faultManager.getOriginalCm().configmap) != 0 {
			t.Errorf("TestFaultCenterCmManagerSetAndGetDeviceInfoCm failed")
		}
	})
}
