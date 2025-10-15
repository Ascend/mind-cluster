// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package l2fault test for l2 fault processor
package l2fault

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/common"
)

const (
	nodeName1   = "nodeName1"
	faultCode1  = "faultCode1"
	faultCode2  = "faultCode2"
	jobId1      = "jobId1"
	jobId2      = "jobId2"
	deviceName1 = "Ascend910-0"
	deviceName2 = "Ascend910-1"
	deviceName3 = "Ascend910-2"
	mindIeJobId = "mindie-ms"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func mockDeviceContent() constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm] {
	return constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap: map[string]*constant.AdvanceDeviceFaultCm{
			nodeName1: mockAdvanceDeviceFaultCm(),
		},
	}
}

func mockSwitchContent() constant.OneConfigmapContent[*constant.SwitchInfo] {
	return constant.OneConfigmapContent[*constant.SwitchInfo]{
		AllConfigmap: map[string]*constant.SwitchInfo{
			constant.SwitchInfoPrefix + nodeName1: mockSwitchInfo(),
		},
	}
}

func mockNodeJobInfoMap() map[string]map[string]constant.JobInfo {
	return map[string]map[string]constant.JobInfo{nodeName1: {jobId1: {}}}
}

func mockNodeJobUsedDeviceMap() map[string]map[string]sets.String {
	return map[string]map[string]sets.String{nodeName1: {jobId1: sets.NewString(deviceName1)}}
}

func mockAdvanceDeviceFaultCm() *constant.AdvanceDeviceFaultCm {
	return &constant.AdvanceDeviceFaultCm{
		FaultDeviceList: map[string][]constant.DeviceFault{deviceName1: {mockDeviceFault()}},
	}
}

func mockSwitchInfo() *constant.SwitchInfo {
	return &constant.SwitchInfo{
		SwitchFaultInfo: constant.SwitchFaultInfo{
			FaultInfo: []constant.SimpleSwitchFaultInfo{
				{AssembledFaultCode: faultCode1},
				{AssembledFaultCode: faultCode2},
			},
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
				faultCode1 + "_0_0": {FaultLevel: constant.NotHandleFault},
				faultCode2 + "_0_0": {FaultLevel: constant.RestartRequest},
			},
			FaultLevel: constant.RestartRequest,
			NodeStatus: constant.UnHealthyState,
		},
		CmName: constant.SwitchInfoPrefix + nodeName1,
	}
}

func mockDeviceFaults() []constant.DeviceFault {
	return []constant.DeviceFault{
		{
			FaultCode: faultCode1,
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
				faultCode1: {FaultLevel: constant.NotHandleFault},
			},
		},
		{
			FaultCode: faultCode2,
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
				faultCode2: {FaultLevel: constant.RestartRequest},
			},
		},
	}
}

func mockJobInfoMap() map[string]constant.JobInfo {
	return map[string]constant.JobInfo{
		jobId1: {MultiInstanceJobId: mindIeJobId},
		jobId2: {MultiInstanceJobId: mindIeJobId},
	}
}

func mockJobUsedDeviceMap() map[string]sets.String {
	return map[string]sets.String{
		jobId1: sets.NewString(deviceName1),
		jobId2: sets.NewString(deviceName2),
	}
}

func mockDeviceFault() constant.DeviceFault {
	return constant.DeviceFault{
		FaultType: constant.CardUnhealthy,
		NPUName:   deviceName1,
		FaultCode: faultCode1,
		FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			faultCode1: {FaultLevel: constant.NotHandleFault}},
	}
}

func TestProcessDeviceFaults(t *testing.T) {
	convey.Convey("Test processDeviceFaults", t, func() {
		processor := &l2FaultProcessor{}
		deviceContent := mockDeviceContent()
		jobInfoMap := mockNodeJobInfoMap()
		jobUsedDeviceMap := mockNodeJobUsedDeviceMap()
		convey.Convey("if node has no mindie server job info, return result is empty", func() {
			res := processor.processDeviceFaults(deviceContent, map[string]map[string]constant.JobInfo{},
				jobUsedDeviceMap)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if copy device fault cm fail, return result is empty", func() {
			patch := gomonkey.ApplyFuncReturn(copyAdvanceDeviceFaultCm, nil, errors.New("test err"))
			defer patch.Reset()
			res := processor.processDeviceFaults(deviceContent, jobInfoMap, jobUsedDeviceMap)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if process device faults success, return result is not empty", func() {
			faultDeviceList := make(map[string][]constant.DeviceFault)
			faultDeviceList[deviceName1] = []constant.DeviceFault{mockDeviceFault()}
			patch := gomonkey.ApplyFuncReturn(copyAdvanceDeviceFaultCm, &constant.AdvanceDeviceFaultCm{
				FaultDeviceList: faultDeviceList,
			}, nil)
			defer patch.Reset()
			res := processor.processDeviceFaults(deviceContent, jobInfoMap, jobUsedDeviceMap)
			convey.So(len(res) == 1, convey.ShouldBeTrue)
		})
	})
}

func TestProcessSwitchFaults(t *testing.T) {
	convey.Convey("Test processSwitchFaults", t, func() {
		processor := &l2FaultProcessor{}
		switchContent := mockSwitchContent()
		jobInfoMap := mockNodeJobInfoMap()
		convey.Convey("if node has no mindie server job info, return result is empty", func() {
			res := processor.processSwitchFaults(switchContent, map[string]map[string]constant.JobInfo{})
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if copy switch info fail, return result is empty", func() {
			patch := gomonkey.ApplyFuncReturn(copySwitchInfo, nil, errors.New("test err"))
			defer patch.Reset()
			res := processor.processSwitchFaults(switchContent, jobInfoMap)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if process switch faults success, return result is not empty", func() {
			patch := gomonkey.ApplyFuncReturn(copySwitchInfo, &constant.SwitchInfo{}, nil).
				ApplyFuncReturn(getDeletedSwitchL2Fault, []constant.SimpleSwitchFaultInfo{{}})
			defer patch.Reset()
			res := processor.processSwitchFaults(switchContent, jobInfoMap)
			convey.So(len(res) == 1, convey.ShouldBeTrue)
		})
	})
}

func TestCopyAdvanceDeviceFaultCm(t *testing.T) {
	convey.Convey("Test copyAdvanceDeviceFaultCm", t, func() {
		deviceFaultCm := mockAdvanceDeviceFaultCm()
		convey.Convey("if deep copy device fault success, return copy result and nil", func() {
			res, err := copyAdvanceDeviceFaultCm(deviceFaultCm)
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldResemble, deviceFaultCm)
		})
		convey.Convey("if deep copy device fault fail, return nil and error", func() {
			patch := gomonkey.ApplyFuncReturn(util.DeepCopy, errors.New("test err"))
			defer patch.Reset()
			res, err := copyAdvanceDeviceFaultCm(deviceFaultCm)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(res, convey.ShouldBeNil)
		})
	})
}

func TestCopySwitchInfo(t *testing.T) {
	convey.Convey("Test copySwitchInfo", t, func() {
		switchInfo := mockSwitchInfo()
		convey.Convey("if deep copy switch info success, return copy result and nil", func() {
			res, err := copySwitchInfo(switchInfo)
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldResemble, switchInfo)
		})
		convey.Convey("if deep copy switch info fail, return nil and error", func() {
			patch := gomonkey.ApplyFuncReturn(util.DeepCopy, errors.New("test err"))
			defer patch.Reset()
			res, err := copySwitchInfo(switchInfo)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(res, convey.ShouldBeNil)
		})
	})
}

func TestCollectAndRemoveDeviceFaults(t *testing.T) {
	convey.Convey("Test collectAndRemoveDeviceFaults", t, func() {
		processor := &l2FaultProcessor{}
		src := mockAdvanceDeviceFaultCm()
		dst := &constant.AdvanceDeviceFaultCm{
			FaultDeviceList: map[string][]constant.DeviceFault{},
		}
		jobInfoMap := map[string]constant.JobInfo{jobId1: {}}
		usedDeviceInfoMap := map[string]sets.String{jobId1: sets.NewString(deviceName1)}
		patch := gomonkey.ApplyFuncReturn(getDeletedDeviceL2Fault, src.FaultDeviceList[deviceName1])
		defer patch.Reset()
		processor.collectAndRemoveDeviceFaults(src, dst, jobInfoMap, usedDeviceInfoMap)
		convey.So(len(src.FaultDeviceList[deviceName1]) == 0, convey.ShouldBeTrue)
		convey.So(len(dst.FaultDeviceList[deviceName1]) == 1, convey.ShouldBeTrue)
	})
}

func TestGetDeletedDeviceL2Fault(t *testing.T) {
	convey.Convey("test getDeletedDeviceL2Fault", t, func() {
		faults := mockDeviceFaults()
		jobInfoMap := mockJobInfoMap()
		jobUsedDeviceMap := mockJobUsedDeviceMap()
		convey.Convey("if job not use any device, remove fault from delete list", func() {
			res := getDeletedDeviceL2Fault(faults, deviceName1, jobInfoMap, map[string]sets.String{})
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if fault has no time and level info, remove fault from delete list", func() {
			faultsWithoutTimeAndLevel := []constant.DeviceFault{{FaultCode: faultCode1}, {FaultCode: faultCode2}}
			res := getDeletedDeviceL2Fault(faultsWithoutTimeAndLevel, deviceName1, jobInfoMap, jobUsedDeviceMap)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if job does not use fault npu, remove fault from delete list", func() {
			res := getDeletedDeviceL2Fault(faults, deviceName3, jobInfoMap, jobUsedDeviceMap)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("if job use fault npu and should not report fault, add fault to delete list", func() {
			patch := gomonkey.ApplyFuncReturn(shouldReportFault, false)
			defer patch.Reset()
			res := getDeletedDeviceL2Fault(faults, deviceName2, jobInfoMap, jobUsedDeviceMap)
			convey.So(len(res) > 0, convey.ShouldBeTrue)
		})
	})
}

func TestGetDeletedSwitchL2Fault(t *testing.T) {
	convey.Convey("test getDeletedSwitchL2Fault", t, func() {
		switchInfo := mockSwitchInfo()
		jobInfoMap := mockJobInfoMap()
		convey.Convey("if switchInfo has no fault time and level for faultcode, remove fault from delete list",
			func() {
				switchWithoutTimeAndLevel := mockSwitchInfo()
				switchWithoutTimeAndLevel.FaultTimeAndLevelMap = map[string]constant.FaultTimeAndLevel{}
				res := getDeletedSwitchL2Fault(switchWithoutTimeAndLevel, jobInfoMap)
				convey.So(len(res) == 0, convey.ShouldBeTrue)
				convey.So(len(switchWithoutTimeAndLevel.FaultInfo) > 0, convey.ShouldBeTrue)
			})
		convey.Convey("if should report fault, remove fault from delete list", func() {
			patch := gomonkey.ApplyFuncReturn(shouldReportFault, true)
			defer patch.Reset()
			res := getDeletedSwitchL2Fault(switchInfo, jobInfoMap)
			convey.So(len(res) == 0, convey.ShouldBeTrue)
			convey.So(len(switchInfo.FaultInfo) > 0, convey.ShouldBeTrue)
		})
		convey.Convey("if should not report fault, add fault to delete list", func() {
			delete(switchInfo.FaultTimeAndLevelMap, faultCode1+"_0_0")
			patch := gomonkey.ApplyFuncReturn(shouldReportFault, false)
			defer patch.Reset()
			res := getDeletedSwitchL2Fault(switchInfo, jobInfoMap)
			convey.So(len(res) == 1, convey.ShouldBeTrue)
			convey.So(len(switchInfo.FaultInfo) > 0, convey.ShouldBeTrue)
			convey.So(switchInfo.FaultLevel == constant.NotHandleFault, convey.ShouldBeTrue)
			convey.So(switchInfo.NodeStatus == constant.HealthyState, convey.ShouldBeTrue)
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
			FaultTime:         time.Now().UnixMilli(),
			FaultReceivedTime: time.Now().UnixMilli(),
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
		convey.Convey("When fault duration exceeds 60s, should report", func() {
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
