// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package fault a series of controller test function
package fault

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/fault"
)

const (
	fakeJobID1 = "fakeJobId1"
	fakeJobID2 = "fakeJobId2"
)

func getMockFaultDeviceListForTest() []constant.FaultDevice {
	return []constant.FaultDevice{
		{ServerName: "node1", ServerId: "1", DeviceId: "0", FaultLevel: constant.SeparateNPU,
			DeviceType: constant.FaultTypeNPU},
		{ServerName: "node1", ServerId: "1", DeviceId: "1", FaultLevel: constant.RestartNPU,
			DeviceType: constant.FaultTypeNPU},
		{ServerName: "node1", ServerId: "1", DeviceId: "2", FaultLevel: constant.SubHealthFault,
			DeviceType: constant.FaultTypeNPU},
		{ServerName: "node1", ServerId: "1", DeviceId: "3", FaultLevel: constant.NotHandleFault,
			DeviceType: constant.FaultTypeNPU},
		{ServerName: "node0", ServerId: "0", DeviceId: "0", FaultLevel: constant.SeparateNPU,
			DeviceType: constant.FaultTypeSwitch},
		{ServerName: "node2", ServerId: "2", DeviceId: "0", FaultLevel: constant.SubHealthFault,
			DeviceType: constant.FaultTypeNPU},
		{ServerName: "node3", ServerId: "3", DeviceId: "0", FaultLevel: constant.NotHandleFault,
			DeviceType: constant.FaultTypeNPU},
	}
}

func getMockFaultMsgForTest() *fault.FaultMsgSignal {
	return &fault.FaultMsgSignal{
		JobId:      fakeJobID1,
		SignalType: constant.SignalTypeFault,
		NodeFaultInfo: []*fault.NodeFaultInfo{
			{
				NodeName:   "node0",
				NodeIP:     "0",
				FaultLevel: constant.UnHealthyState,
				FaultDevice: []*fault.DeviceFaultInfo{
					{DeviceId: "0", DeviceType: constant.FaultTypeSwitch, FaultLevel: constant.UnHealthyState},
				},
			},
			{
				NodeName:   "node1",
				NodeIP:     "1",
				FaultLevel: constant.UnHealthyState,
				FaultDevice: []*fault.DeviceFaultInfo{
					{DeviceId: "0", DeviceType: constant.FaultTypeNPU, FaultLevel: constant.UnHealthyState},
					{DeviceId: "1", DeviceType: constant.FaultTypeNPU, FaultLevel: constant.UnHealthyState},
					{DeviceId: "2", DeviceType: constant.FaultTypeNPU, FaultLevel: constant.SubHealthyState},
					{DeviceId: "3", DeviceType: constant.FaultTypeNPU, FaultLevel: constant.HealthyState},
				},
			},
			{
				NodeName:   "node2",
				NodeIP:     "2",
				FaultLevel: constant.SubHealthyState,
				FaultDevice: []*fault.DeviceFaultInfo{
					{DeviceId: "0", DeviceType: constant.FaultTypeNPU, FaultLevel: constant.SubHealthyState},
				},
			},
			{
				NodeName:   "node3",
				NodeIP:     "3",
				FaultLevel: constant.HealthyState,
				FaultDevice: []*fault.DeviceFaultInfo{
					{DeviceId: "0", DeviceType: constant.FaultTypeNPU, FaultLevel: constant.HealthyState},
				},
			},
		},
	}
}

// TestFaultDeviceToSortedFaultMsgSignal for test faultDeviceToSortedFaultMsgSignal
func TestFaultDeviceToSortedFaultMsgSignal(t *testing.T) {
	normalMsg := &fault.FaultMsgSignal{
		JobId:      fakeJobID1,
		SignalType: constant.SignalTypeNormal,
	}
	normalMsgWithFaultInfo := &fault.FaultMsgSignal{
		JobId:      fakeJobID1,
		SignalType: constant.SignalTypeNormal,
		NodeFaultInfo: []*fault.NodeFaultInfo{
			{
				NodeName:   "node3",
				NodeIP:     "3",
				FaultLevel: constant.HealthyState,
				FaultDevice: []*fault.DeviceFaultInfo{
					{DeviceId: "0", DeviceType: constant.FaultTypeNPU, FaultLevel: constant.HealthyState},
				},
			},
		},
	}
	convey.Convey("faultList is empty, should convert to normal msg", t, func() {
		msg := faultDeviceToSortedFaultMsgSignal(fakeJobID1, nil)
		convey.So(msg, convey.ShouldResemble, normalMsg)
	})
	convey.Convey("faultList includes not only L1 faults , should convert to fault msg", t, func() {
		msg := faultDeviceToSortedFaultMsgSignal(fakeJobID1, getMockFaultDeviceListForTest())
		convey.So(msg, convey.ShouldResemble, getMockFaultMsgForTest())
	})
	convey.Convey("faultList includes only L1 faults, should convert to normal msg", t, func() {
		faultDevice := []constant.FaultDevice{{ServerName: "node3", ServerId: "3", DeviceId: "0",
			FaultLevel: constant.NotHandleFault, DeviceType: constant.FaultTypeNPU}}
		msg := faultDeviceToSortedFaultMsgSignal(fakeJobID1, faultDevice)
		convey.So(msg, convey.ShouldResemble, normalMsgWithFaultInfo)
	})
	convey.Convey("faultList includes only L1 faults, nodeInfo is nil,"+
		"should convert to normal msg", t, func() {
		patch := gomonkey.ApplyFuncReturn(getNodeFaultInfo, nil)
		defer patch.Reset()
		faultDevice := []constant.FaultDevice{{ServerName: "node3", ServerId: "3", DeviceId: "0",
			FaultLevel: constant.NotHandleFault, DeviceType: constant.FaultTypeNPU}}
		msg := faultDeviceToSortedFaultMsgSignal(fakeJobID1, faultDevice)
		convey.So(msg, convey.ShouldResemble, normalMsg)
	})
}

func getMockFaultDeviceListForTest1() []constant.FaultDevice {
	return []constant.FaultDevice{
		{ServerName: "node1", ServerId: "1", DeviceId: "0", FaultLevel: constant.SeparateNPU,
			DeviceType: constant.FaultTypeNPU},
	}
}

func getMockFaultMsgForTest1() *fault.FaultMsgSignal {
	return &fault.FaultMsgSignal{
		JobId:      fakeJobID1,
		SignalType: constant.SignalTypeFault,
		NodeFaultInfo: []*fault.NodeFaultInfo{{
			NodeName:   "node1",
			NodeIP:     "1",
			FaultLevel: constant.UnHealthyState,
			FaultDevice: []*fault.DeviceFaultInfo{
				{DeviceId: "0", DeviceType: constant.FaultTypeNPU, FaultLevel: constant.UnHealthyState},
			}}}}
}

// TestCheckPublishFault for test checkPublishFault
func TestCheckPublishFault(t *testing.T) {
	patch := gomonkey.ApplyFuncReturn(job.GetJobCache, constant.JobInfo{MultiInstanceJobId: fakeJobID1,
		AppType: "controller"}, true)
	defer patch.Reset()
	allJobFaultInfo := map[string]constant.JobFaultInfo{
		fakeJobID2: {FaultDevice: getMockFaultDeviceListForTest1()},
	}
	service := fakeFaultService()
	service.addPublisher(fakeJobID1)
	faultPublisher, ok := service.getPublisher(fakeJobID1)
	if !ok {
		t.Error("get faultPublisher fail")
		return
	}
	faultPublisher.SetSubscribe(true)
	sendChan := faultPublisher.GetSentChan()

	var data *fault.FaultMsgSignal = nil
	convey.Convey("occur fault, should send fault msg", t, func() {
		service.checkPublishFault(allJobFaultInfo)
		convey.So(len(sendChan), convey.ShouldEqual, 1)
		data = <-sendChan
		convey.So(compareFaultMsg(data, getMockFaultMsgForTest1()), convey.ShouldBeTrue)
	})
	convey.Convey("occur not change, should not send fault msg", t, func() {
		faultPublisher.SetSentData(data)
		service.checkPublishFault(allJobFaultInfo)
		convey.So(len(sendChan), convey.ShouldEqual, 0)
	})
	convey.Convey("fault recover, should send fault recover msg", t, func() {
		faultPublisher.SetSentData(data)
		service.checkPublishFault(nil)
		convey.So(len(sendChan), convey.ShouldEqual, 1)
		data = <-sendChan
		convey.So(compareFaultMsg(data, &fault.FaultMsgSignal{
			JobId:      fakeJobID1,
			SignalType: constant.SignalTypeNormal,
		}), convey.ShouldBeTrue)
	})
	convey.Convey("fault recover and already sent to the client, should not send fault recover msg", t, func() {
		faultPublisher.SetSentData(data)
		service.checkPublishFault(nil)
		convey.So(len(sendChan), convey.ShouldEqual, 0)
	})
}

// TestCheckFaultFromFaultCenter for test checkFaultFromFaultCenter
func TestCheckFaultFromFaultCenter(t *testing.T) {
	convey.Convey("occur fault, should send fault msg", t, func() {
		service := fakeFaultService()
		ctx, cancel := context.WithCancel(context.Background())
		service.serviceCtx = ctx
		success := atomic.Bool{}
		go func() {
			service.checkFaultFromFaultCenter()
			success.Store(true)
		}()
		cancel()
		time.Sleep(sleepTime)
		convey.So(success.Load(), convey.ShouldBeTrue)
	})
}

func TestFilterFault(t *testing.T) {
	convey.Convey("filter NotHandle and NotHandFault error", t, func() {
		faultDeviceList := []constant.FaultDevice{
			{ServerName: "node1", DeviceId: "0", FaultLevel: constant.SeparateNPU},
			{ServerName: "node1", DeviceId: "1", FaultLevel: constant.NotHandleFault},
			{ServerName: "node1", DeviceId: "1", FaultLevel: constant.NotHandleFaultLevelStr}}
		filteredList := filterFault(faultDeviceList)
		convey.So(len(filteredList), convey.ShouldEqual, 1)
		convey.So(filteredList, convey.ShouldResemble, []constant.FaultDevice{
			{ServerName: "node1", DeviceId: "0", FaultLevel: constant.SeparateNPU}})
	})
}

// Helper function to create a FaultMsgSignal with given fields
func newFaultMsg(signalType string, jobId string, nodeFaultInfo []*fault.NodeFaultInfo) *fault.FaultMsgSignal {
	return &fault.FaultMsgSignal{
		SignalType:    signalType,
		JobId:         jobId,
		NodeFaultInfo: nodeFaultInfo,
	}
}

// Helper function to create a NodeFault for testing
func newNodeFault(nodeName string) *fault.NodeFaultInfo {
	return &fault.NodeFaultInfo{
		NodeName: nodeName,
	}
}

type FaultMsgTestCase struct {
	name     string
	this     *fault.FaultMsgSignal
	other    *fault.FaultMsgSignal
	expected bool
}

func buildFaultMsgTestCases1() []FaultMsgTestCase {
	return []FaultMsgTestCase{
		{
			name:     "TC01 - Both nil",
			this:     nil,
			other:    nil,
			expected: true,
		},
		{
			name:     "TC02 - This nil, Other is Normal",
			this:     nil,
			other:    newFaultMsg(constant.SignalTypeNormal, "", nil),
			expected: true,
		},
		{
			name:     "TC03 - This is Normal, Other nil",
			this:     newFaultMsg(constant.SignalTypeNormal, "", nil),
			other:    nil,
			expected: true,
		},
		{
			name:     "TC04 - Both are Normal",
			this:     newFaultMsg(constant.SignalTypeNormal, "", nil),
			other:    newFaultMsg(constant.SignalTypeNormal, "", nil),
			expected: true,
		},
	}
}

func buildFaultMsgTestCases2() []FaultMsgTestCase {
	return []FaultMsgTestCase{
		{
			name:     "TC05 - SignalType mismatch",
			this:     newFaultMsg("Fault", "1", nil),
			other:    newFaultMsg(constant.SignalTypeNormal, "1", nil),
			expected: false,
		},
		{
			name:     "TC06 - JobId mismatch",
			this:     newFaultMsg("Fault", "1", nil),
			other:    newFaultMsg("Fault", "2", nil),
			expected: false,
		},
		{
			name: "TC07 - NodeFaultInfo content different",
			this: newFaultMsg("Fault", "1", []*fault.NodeFaultInfo{
				newNodeFault("node1"),
			}),
			other: newFaultMsg("Fault", "1", []*fault.NodeFaultInfo{
				newNodeFault("node2"),
			}),
			expected: false,
		},
		{
			name: "TC08 - All fields match",
			this: newFaultMsg("Fault", "1", []*fault.NodeFaultInfo{
				newNodeFault("node1"),
			}),
			other: newFaultMsg("Fault", "1", []*fault.NodeFaultInfo{
				newNodeFault("node1"),
			}),
			expected: true,
		},
	}
}

func TestCompareFaultMsg(t *testing.T) {
	tests := buildFaultMsgTestCases1()
	tests = append(tests, buildFaultMsgTestCases2()...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareFaultMsg(tt.this, tt.other)
			assert.Equal(t, tt.expected, result)
		})
	}
}
