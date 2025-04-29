// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package faultmanager contain fault process
package cmmanager

import (
	"reflect"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain/collector"
)

func TestMain(m *testing.M) {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	m.Run()
}

func TestFaultCenterCmManagerSetAndGetConfigmap(t *testing.T) {
	t.Run("", func(t *testing.T) {
		deviceCM := ConfigMap[*constant.DeviceInfo]{Data: make(map[string]*constant.DeviceInfo)}
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
		if !reflect.DeepEqual(deviceCM.Data["cm1"], cm1) ||
			!reflect.DeepEqual(deviceCM.Data["cm2"], cm2) {
			t.Errorf("TestFaultCenterCmManagerSetAndGetDeviceInfoCm failed")
		}
		faultManager := FaultCenterCmManager[*constant.DeviceInfo]{
			mutex:       sync.RWMutex{},
			originalCm:  ConfigMap[*constant.DeviceInfo]{},
			processedCm: ConfigMap[*constant.DeviceInfo]{},
		}
		faultManager.updateOriginalCm(cm1, true)
		faultManager.updateOriginalCm(cm2, true)
		if !reflect.DeepEqual(deviceCM, faultManager.GetOriginalCm()) {
			t.Errorf("TestFaultCenterCmManagerSetAndGetDeviceInfoCm failed")
		}

		faultManager.updateOriginalCm(cm1, false)
		faultManager.updateOriginalCm(cm2, false)
		if len(faultManager.GetOriginalCm().Data) != 0 {
			t.Errorf("TestFaultCenterCmManagerSetAndGetDeviceInfoCm failed")
		}
	})
}

func TestConfigMapEqual(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	convey.Convey("Test ConfigMap.equal()", t, func() {
		cm1 := ConfigMap[*constant.AdvanceDeviceFaultCm]{
			Data: map[string]*constant.AdvanceDeviceFaultCm{
				"node": {
					FaultDeviceList:     nil,
					AvailableDeviceList: []string{"2", "1"},
					Recovering:          []string{"4", "3"},
					CardUnHealthy:       []string{"6", "5"},
					NetworkUnhealthy:    []string{"6", "5"},
					UpdateTime:          10,
				},
			},
		}
		cm2 := new(ConfigMap[*constant.AdvanceDeviceFaultCm])
		util.DeepCopy(cm2, cm1)

		convey.Convey("should return true for equal config maps", func() {
			convey.So(cm1.equal(*cm2), convey.ShouldBeTrue)
		})
	})
}

func TestUpdateBatchOriginalCm(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	convey.Convey("Test UpdateBatchOriginalCm()", t, func() {
		manager := FaultCenterCmManager[*constant.AdvanceDeviceFaultCm]{
			mutex:       sync.RWMutex{},
			cmBuffer:    &collector.ConfigmapCollectBuffer[*constant.AdvanceDeviceFaultCm]{},
			originalCm:  ConfigMap[*constant.AdvanceDeviceFaultCm]{},
			processedCm: ConfigMap[*constant.AdvanceDeviceFaultCm]{},
		}
		manager.UpdateBatchOriginalCm()
		convey.So(manager.originalCm.Data, convey.ShouldBeEmpty)
	})
}

func TestSetAndGetProcessedCm(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	convey.Convey("Test SetProcessedCm() and GetProcessedCm()", t, func() {
		manager := FaultCenterCmManager[*constant.AdvanceDeviceFaultCm]{
			mutex:       sync.RWMutex{},
			cmBuffer:    &collector.ConfigmapCollectBuffer[*constant.AdvanceDeviceFaultCm]{},
			originalCm:  ConfigMap[*constant.AdvanceDeviceFaultCm]{},
			processedCm: ConfigMap[*constant.AdvanceDeviceFaultCm]{},
		}
		cm := ConfigMap[*constant.AdvanceDeviceFaultCm]{
			Data: map[string]*constant.AdvanceDeviceFaultCm{
				"node": {},
			},
		}
		manager.SetProcessedCm(cm)
		convey.So(manager.processedCm, convey.ShouldResemble, cm)

		processedCm := manager.GetProcessedCm()
		convey.So(processedCm, convey.ShouldResemble, cm)
	})
}
