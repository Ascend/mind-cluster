// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package l2fault test for l2 fault processor
package l2fault

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/common"
)

const (
	faultLevel  = "NotHandle"
	nodeName1   = "nodeName1"
	faultCode1  = "code1"
	jobId1      = "jobId1"
	deviceName1 = "Ascend910-0"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func mockSwitchInfo() *constant.SwitchInfo {
	return &constant.SwitchInfo{
		SwitchFaultInfo: constant.SwitchFaultInfo{
			FaultInfo: []constant.SimpleSwitchFaultInfo{
				{
					AssembledFaultCode: faultCode1,
				},
			},
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
				faultCode1: {
					FaultLevel: faultLevel,
				},
			},
		},
	}
}

func mockDeviceFaults() []constant.DeviceFault {
	return []constant.DeviceFault{
		{
			FaultType:            constant.CardUnhealthy,
			NPUName:              deviceName1,
			FaultCode:            faultCode1,
			FaultLevel:           constant.SubHealthFault,
			LargeModelFaultLevel: constant.SubHealthFault,
			FaultHandling:        constant.SubHealthFault,
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
				faultCode1: {FaultLevel: constant.NotHandleFault, FaultTime: 1},
			},
		},
	}
}

func TestDealL2DeviceFault(t *testing.T) {
	convey.Convey("test dealDeviceL2Fault", t, func() {
		deviceFaults := mockDeviceFaults()
		patch := gomonkey.ApplyFuncReturn(job.GetInferenceJobIdByNodeName, jobId1)
		defer patch.Reset()
		convey.Convey("if not found device fault level, remove fault from delete list", func() {
			deviceFaults[0].FaultTimeAndLevelMap = map[string]constant.FaultTimeAndLevel{}
			res := dealDeviceL2Fault(deviceFaults, nodeName1, deviceName1)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if fault should be report, remove fault from delete list", func() {
			patch1 := gomonkey.ApplyFuncReturn(shouldReportFault, true)
			defer patch1.Reset()
			res := dealDeviceL2Fault(deviceFaults, nodeName1, deviceName1)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if fault should not be report, add fault to delete list", func() {
			patch2 := gomonkey.ApplyFuncReturn(shouldReportFault, false)
			defer patch2.Reset()
			res := dealDeviceL2Fault(deviceFaults, nodeName1, deviceName1)
			convey.So(len(res) == 1, convey.ShouldBeTrue)
		})
	})
}

func TestDealL2SwitchFault(t *testing.T) {
	convey.Convey("test dealSwitchL2Fault", t, func() {
		switchInfoMap := mockSwitchInfo()
		patch := gomonkey.ApplyFuncReturn(job.GetInferenceJobIdByNodeName, jobId1)
		defer patch.Reset()
		convey.Convey("if not found switch fault level, add fault to result", func() {
			switchInfoMap.FaultTimeAndLevelMap = map[string]constant.FaultTimeAndLevel{}
			res := dealSwitchL2Fault(switchInfoMap, nodeName1)
			convey.So(len(res) == 1, convey.ShouldBeTrue)
		})
		convey.Convey("if fault should be report, add fault to result", func() {
			patch1 := gomonkey.ApplyFuncReturn(shouldReportFault, true)
			defer patch1.Reset()
			res := dealSwitchL2Fault(switchInfoMap, nodeName1)
			convey.So(len(res) == 1, convey.ShouldBeTrue)
		})
		convey.Convey("if fault should not be report, remove fault from result", func() {
			patch2 := gomonkey.ApplyFuncReturn(shouldReportFault, false)
			defer patch2.Reset()
			res := dealSwitchL2Fault(switchInfoMap, nodeName1)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
	})
}

type mockFaultPublisher struct {
	isSubscribed bool
}

func (m *mockFaultPublisher) IsSubscribed(topic, subscriber string) bool {
	return m.isSubscribed
}

func TestShouldReportFault(t *testing.T) {
	convey.Convey("test shouldReportFault", t, func() {
		mockPubSubscribed := &mockFaultPublisher{isSubscribed: true}
		mockPubNotSubscribed := &mockFaultPublisher{isSubscribed: false}
		patchNow := func(ts int64) *gomonkey.Patches {
			return gomonkey.ApplyFunc(time.Now, func() time.Time {
				return time.UnixMilli(ts)
			})
		}
		convey.Convey("if fault level is not L2, should report fault", func() {
			res := shouldReportFault(constant.NotHandleFaultLevelStr, 0, "", "", "")
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("if fault during more than 10s, should report fault", func() {
			patch := patchNow(time.Now().Add(selfrecoverFaultTimeout + time.Second).UnixMilli())
			defer patch.Reset()
			res := shouldReportFault(constant.RestartRequest, 0, "", "", "")
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("if jobId is empty, should report fault", func() {
			patch := patchNow(time.Now().Add(selfrecoverFaultTimeout).UnixMilli())
			defer patch.Reset()
			res := shouldReportFault(constant.RestartRequest, 0, "", "", "")
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("if deviceName is not empty and job not use this device, should report fault", func() {
			patch := patchNow(time.Now().Add(selfrecoverFaultTimeout).UnixMilli()).
				ApplyFuncReturn(job.IsJobUsedDevice, false)
			defer patch.Reset()
			res := shouldReportFault(constant.RestartRequest, 0, jobId1, deviceName1, "")
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("if job not subscribed, should report fault", func() {
			common.SetPublisher(mockPubNotSubscribed)
			patch := patchNow(time.Now().Add(selfrecoverFaultTimeout).UnixMilli()).
				ApplyFuncReturn(job.IsJobUsedDevice, true)
			defer patch.Reset()
			res := shouldReportFault(constant.RestartRequest, 0, jobId1, deviceName1, "")
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("if job subscribed, should not report fault", func() {
			common.SetPublisher(mockPubSubscribed)
			patch := patchNow(time.Now().Add(selfrecoverFaultTimeout).UnixMilli()).
				ApplyFuncReturn(job.IsJobUsedDevice, true)
			defer patch.Reset()
			res := shouldReportFault(constant.RestartRequest, 0, jobId1, deviceName1, "")
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}
