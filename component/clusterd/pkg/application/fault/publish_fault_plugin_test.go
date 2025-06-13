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
		convey.So(data, convey.ShouldResemble, getMockFaultMsgForTest1())
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
		convey.So(data, convey.ShouldResemble, &fault.FaultMsgSignal{
			JobId:      fakeJobID1,
			SignalType: constant.SignalTypeNormal,
		})
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
