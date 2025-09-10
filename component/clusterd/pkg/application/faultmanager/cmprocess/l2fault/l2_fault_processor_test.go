// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package l2fault test for l2 fault processor
package l2fault

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/sets"

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
	jobId2      = "jobId2"
	deviceName1 = "Ascend910-0"
	deviceName2 = "Ascend910-1"
	mindIeJobId = "mindie-ms"
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
	convey.Convey("test getDeletedDeviceL2Fault", t, func() {
		deviceFaults := mockDeviceFaults()
		jobInfoMap := map[string]constant.JobInfo{jobId1: {}}
		mindIeJobInfoMap := map[string]map[string]constant.JobInfo{nodeName1: {jobId1: {}}}
		mindIeDeviceInfoMap := map[string]map[string]sets.String{nodeName1: {jobId1: {}}}
		patch := gomonkey.ApplyFuncReturn(job.GetMindIeServerJobAndUsedDeviceInfoMap,
			mindIeJobInfoMap, mindIeDeviceInfoMap)
		defer patch.Reset()
		convey.Convey("if job not use any device, remove fault from delete list", func() {
			jobUsedDeviceInfoMap := map[string]sets.String{jobId2: sets.NewString(deviceName1)}
			res := getDeletedDeviceL2Fault(deviceFaults, deviceName1, jobInfoMap, jobUsedDeviceInfoMap)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if job not use fault device, remove fault from delete list", func() {
			jobUsedDeviceInfoMap := map[string]sets.String{jobId1: sets.NewString(deviceName2)}
			res := getDeletedDeviceL2Fault(deviceFaults, deviceName1, jobInfoMap, jobUsedDeviceInfoMap)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		jobUsedDeviceInfoMap := map[string]sets.String{jobId1: sets.NewString(deviceName1)}
		convey.Convey("if not found device fault level, remove fault from delete list", func() {
			deviceFaults[0].FaultTimeAndLevelMap = map[string]constant.FaultTimeAndLevel{}
			res := getDeletedDeviceL2Fault(deviceFaults, deviceName1, jobInfoMap, jobUsedDeviceInfoMap)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if fault should be report, remove fault from delete list", func() {
			patch1 := gomonkey.ApplyFuncReturn(shouldReportFault, true)
			defer patch1.Reset()
			res := getDeletedDeviceL2Fault(deviceFaults, deviceName1, jobInfoMap, jobUsedDeviceInfoMap)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if fault should not be report, add fault to delete list", func() {
			patch2 := gomonkey.ApplyFuncReturn(shouldReportFault, false)
			defer patch2.Reset()
			res := getDeletedDeviceL2Fault(deviceFaults, deviceName1, jobInfoMap, jobUsedDeviceInfoMap)
			convey.So(len(res) == 1, convey.ShouldBeTrue)
		})
	})
}

func TestDealL2SwitchFault(t *testing.T) {
	convey.Convey("test getDeletedSwitchL2Fault", t, func() {
		switchInfoMap := mockSwitchInfo()
		jobInfoMap := map[string]constant.JobInfo{jobId1: {}}
		mindIeJobInfoMap := map[string]map[string]constant.JobInfo{nodeName1: {jobId1: {}}}
		mindIeDeviceInfoMap := map[string]map[string]sets.String{nodeName1: {jobId1: {}}}
		patch := gomonkey.ApplyFuncReturn(job.GetMindIeServerJobAndUsedDeviceInfoMap,
			mindIeJobInfoMap, mindIeDeviceInfoMap)
		defer patch.Reset()
		convey.Convey("if not found switch fault level, remove fault from delete list", func() {
			switchInfoMap.FaultTimeAndLevelMap = map[string]constant.FaultTimeAndLevel{}
			res := getDeletedSwitchL2Fault(switchInfoMap, jobInfoMap)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if fault should be report, remove fault from delete list", func() {
			patch1 := gomonkey.ApplyFuncReturn(shouldReportFault, true)
			defer patch1.Reset()
			res := getDeletedSwitchL2Fault(switchInfoMap, jobInfoMap)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if fault should not be report, add fault to delete list", func() {
			patch2 := gomonkey.ApplyFuncReturn(shouldReportFault, false)
			defer patch2.Reset()
			res := getDeletedSwitchL2Fault(switchInfoMap, jobInfoMap)
			convey.So(len(res) == 1, convey.ShouldBeTrue)
		})
	})
}

type mockFaultPublisher struct {
	isSubscribed bool
}

func (m *mockFaultPublisher) IsSubscribed(topic, subscriber string) bool {
	return m.isSubscribed
}

type testShouldReportFaultCases struct {
	jobId1      string
	deviceName1 string
	timeout     time.Duration
}

func TestShouldReportFault(t *testing.T) {
	convey.Convey("Test shouldReportFault behavior under different conditions", t, func() {
		mockPubSubscribed := &mockFaultPublisher{isSubscribed: true}
		mockPubNotSubscribed := &mockFaultPublisher{isSubscribed: false}

		patchNow := func(ts int64) *gomonkey.Patches {
			return gomonkey.ApplyFunc(time.Now, func() time.Time {
				return time.UnixMilli(ts)
			})
		}

		baseFaultTimeAndLevel := constant.FaultTimeAndLevel{
			FaultTime: time.Now().UnixMilli(),
		}

		testCases := testShouldReportFaultCases{
			jobId1:      "test-job-1",
			deviceName1: "npu-0",
			timeout:     selfrecoverFaultTimeout,
		}

		patchWithOffset := func(offset time.Duration) *gomonkey.Patches {
			ts := time.Now().Add(offset).UnixMilli()
			return patchNow(ts)
		}

		convey.Convey("When fault level is not L2, should report fault", func() {
			fault := baseFaultTimeAndLevel
			fault.FaultLevel = constant.NotHandleFaultLevelStr

			res := shouldReportFault(fault, constant.JobInfo{}, "", "")
			convey.So(res, convey.ShouldBeTrue)
		})

		l2Fault := baseFaultTimeAndLevel
		l2Fault.FaultLevel = constant.RestartRequest
		testL2LevelFaultScenarios(l2Fault, testCases, patchWithOffset, mockPubSubscribed, mockPubNotSubscribed)
	})
}

func testL2LevelFaultScenarios(l2Fault constant.FaultTimeAndLevel, testCases testShouldReportFaultCases,
	patchWithOffset func(time.Duration) *gomonkey.Patches,
	mockPubSubscribed, mockPubNotSubscribed *mockFaultPublisher) {
	convey.Convey("For L2 level faults (RestartRequest)", func() {
		convey.Convey("When fault duration exceeds 10s, should report", func() {
			patch := patchWithOffset(testCases.timeout + time.Second)
			defer patch.Reset()

			res := shouldReportFault(l2Fault, constant.JobInfo{}, "", "")
			convey.So(res, convey.ShouldBeTrue)
		})

		convey.Convey("When job is not subscribed, should report", func() {
			common.SetPublisher(mockPubNotSubscribed)
			patch := patchWithOffset(testCases.timeout-time.Second).
				ApplyFuncReturn(common.Publisher.IsSubscribed, false)
			defer patch.Reset()

			res := shouldReportFault(l2Fault, constant.JobInfo{Key: testCases.jobId1},
				testCases.deviceName1, "")
			convey.So(res, convey.ShouldBeTrue)
		})

		convey.Convey("When job is subscribed, should NOT report", func() {
			common.SetPublisher(mockPubSubscribed)
			patch := patchWithOffset(testCases.timeout-time.Second).
				ApplyFuncReturn(common.Publisher.IsSubscribed, true)
			defer patch.Reset()

			res := shouldReportFault(l2Fault, constant.JobInfo{Key: testCases.jobId1},
				testCases.deviceName1, "")
			convey.So(res, convey.ShouldBeFalse)
		})
	})
}
