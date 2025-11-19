// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package preseparate test for preseparate fault processor
package preseparate

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/strings/slices"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/pod"
)

const (
	deviceName1 = "Ascend910-1"
	nodeName1   = "node1"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func testProcessAdvanceDeviceFaultCm() {
	processor := &preSeparateFaultProcessor{}
	faultDevice := &constant.AdvanceDeviceFaultCm{
		FaultDeviceList: map[string][]constant.DeviceFault{
			deviceName1: {{FaultLevel: constant.PreSeparateNPU}},
		},
		AvailableDeviceList: []string{deviceName1},
		CardUnHealthy:       []string{},
	}
	deviceContent := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap: map[string]*constant.AdvanceDeviceFaultCm{
			nodeName1: faultDevice,
		},
	}
	patches := gomonkey.ApplyFuncReturn(pod.GetUsedDevicesByNodeName, sets.String{})
	defer patches.Reset()
	result := processor.Process(deviceContent)
	convey.So(result, convey.ShouldResemble, deviceContent)
	convey.So(slices.Contains(faultDevice.AvailableDeviceList, deviceName1), convey.ShouldBeFalse)
	convey.So(slices.Contains(faultDevice.CardUnHealthy, deviceName1), convey.ShouldBeTrue)
}

func testProcessSwitchInfo() {
	processor := &preSeparateFaultProcessor{}
	switchInfo := &constant.SwitchInfo{
		SwitchFaultInfo: constant.SwitchFaultInfo{
			FaultLevel: constant.PreSeparateFaultLevelStr,
			NodeStatus: "",
		},
	}
	switchContent := constant.OneConfigmapContent[*constant.SwitchInfo]{
		AllConfigmap: map[string]*constant.SwitchInfo{
			constant.SwitchInfoPrefix + nodeName1: switchInfo,
		},
	}
	patches := gomonkey.ApplyFuncReturn(pod.GetUsedDevicesByNodeName, sets.String{deviceName1: {}})
	defer patches.Reset()
	result := processor.Process(switchContent)
	convey.So(result, convey.ShouldResemble, switchContent)
	convey.So(switchInfo.NodeStatus, convey.ShouldEqual, constant.HealthyState)
}

func testProcessNodeInfo() {
	processor := &preSeparateFaultProcessor{}
	nodeInfo := &constant.NodeInfo{
		NodeInfoNoName: constant.NodeInfoNoName{
			FaultDevList: []*constant.FaultDev{
				{FaultLevel: constant.PreSeparateFault},
			},
			NodeStatus: "",
		},
	}
	nodeContent := constant.OneConfigmapContent[*constant.NodeInfo]{
		AllConfigmap: map[string]*constant.NodeInfo{
			constant.NodeInfoPrefix + nodeName1: nodeInfo,
		},
	}
	patches := gomonkey.ApplyFuncReturn(pod.GetUsedDevicesByNodeName, sets.String{})
	defer patches.Reset()
	result := processor.Process(nodeContent)
	convey.So(result, convey.ShouldResemble, nodeContent)
	convey.So(nodeInfo.NodeStatus, convey.ShouldEqual, constant.UnHealthyState)
}

func TestProcess(t *testing.T) {
	convey.Convey("Test processing AdvanceDeviceFaultCm content", t, testProcessAdvanceDeviceFaultCm)
	convey.Convey("Test processing SwitchInfo content", t, testProcessSwitchInfo)
	convey.Convey("Test processing NodeInfo content", t, testProcessNodeInfo)
	convey.Convey("Test processing unknown type input", t, func() {
		processor := &preSeparateFaultProcessor{}
		unknownInput := struct{ name string }{name: "test"}
		result := processor.Process(unknownInput)
		convey.So(result, convey.ShouldResemble, unknownInput)
	})
}

func TestProcessDeviceFaultCm(t *testing.T) {
	convey.Convey("Test processDeviceFaultCm function", t, func() {
		processor := &preSeparateFaultProcessor{}
		faultCm := &constant.AdvanceDeviceFaultCm{
			FaultDeviceList: map[string][]constant.DeviceFault{
				deviceName1: {
					{FaultLevel: constant.PreSeparateNPU},
					{FaultLevel: constant.NotHandleFault},
				},
			},
			AvailableDeviceList: []string{deviceName1},
			CardUnHealthy:       []string{},
		}

		patches := gomonkey.ApplyFuncReturn(pod.GetUsedDevicesByNodeName, sets.String{})
		defer patches.Reset()
		processor.processDeviceFaultCm(faultCm, nodeName1)
		convey.So(slices.Contains(faultCm.AvailableDeviceList, deviceName1), convey.ShouldBeFalse)
		convey.So(slices.Contains(faultCm.CardUnHealthy, deviceName1), convey.ShouldBeTrue)
	})
}
func TestUpdateCardUnHealthy(t *testing.T) {
	convey.Convey("Test updateCardUnHealthy function", t, func() {
		processor := &preSeparateFaultProcessor{}
		faultCm := &constant.AdvanceDeviceFaultCm{
			CardUnHealthy: []string{},
		}

		convey.Convey("When device is in use", func() {
			patches := gomonkey.ApplyFuncReturn(pod.GetUsedDevicesByNodeName, sets.String{deviceName1: {}})
			defer patches.Reset()
			processor.updateCardUnHealthy(faultCm, nodeName1, deviceName1)
			convey.So(slices.Contains(faultCm.CardUnHealthy, deviceName1), convey.ShouldBeFalse)
		})

		convey.Convey("When device is not in use", func() {
			patches := gomonkey.ApplyFuncReturn(pod.GetUsedDevicesByNodeName, sets.String{})
			defer patches.Reset()
			processor.updateCardUnHealthy(faultCm, nodeName1, deviceName1)
			convey.So(slices.Contains(faultCm.CardUnHealthy, deviceName1), convey.ShouldBeTrue)
		})
	})
}

func TestUpdateNodeStatusBySwitchInfo(t *testing.T) {
	convey.Convey("Test updateNodeStatusBySwitchInfo function", t, func() {
		processor := &preSeparateFaultProcessor{}
		switchInfo := &constant.SwitchInfo{}
		cmName := constant.SwitchInfoPrefix + nodeName1

		convey.Convey("When fault level does not match", func() {
			switchInfo.FaultLevel = "InvalidLevel"
			processor.updateNodeStatusBySwitchInfo(cmName, switchInfo)
			convey.So(switchInfo.NodeStatus, convey.ShouldBeEmpty)
		})

		convey.Convey("When fault level matches and has used devices", func() {
			switchInfo.FaultLevel = constant.PreSeparateFaultLevelStr
			patches := gomonkey.ApplyFuncReturn(pod.GetUsedDevicesByNodeName, sets.String{deviceName1: {}})
			defer patches.Reset()

			processor.updateNodeStatusBySwitchInfo(cmName, switchInfo)
			convey.So(switchInfo.NodeStatus, convey.ShouldEqual, constant.HealthyState)
		})
	})
}

func TestUpdateNodeStatusByNodeInfo(t *testing.T) {
	convey.Convey("Test updateNodeStatusByNodeInfo function", t, func() {
		processor := &preSeparateFaultProcessor{}
		nodeInfo := &constant.NodeInfo{
			NodeInfoNoName: constant.NodeInfoNoName{
				FaultDevList: []*constant.FaultDev{
					{FaultLevel: constant.PreSeparateFault},
					{FaultLevel: constant.NotHandleFault},
				},
				NodeStatus: "",
			},
		}
		cmName := constant.NodeInfoPrefix + nodeName1
		patches := gomonkey.ApplyFuncReturn(faultdomain.GetNodeMostSeriousFaultLevel, constant.PreSeparateFault)
		defer patches.Reset()

		convey.Convey("When no used devices", func() {
			podPatches := gomonkey.ApplyFuncReturn(pod.GetUsedDevicesByNodeName, sets.String{})
			defer podPatches.Reset()
			processor.updateNodeStatusByNodeInfo(cmName, nodeInfo)
			convey.So(nodeInfo.NodeStatus, convey.ShouldEqual, constant.UnHealthyState)
		})

		convey.Convey("When most serious level does not match", func() {
			levelPatches := gomonkey.ApplyFuncReturn(faultdomain.GetNodeMostSeriousFaultLevel, constant.NotHandleFault)
			defer levelPatches.Reset()
			processor.updateNodeStatusByNodeInfo(cmName, nodeInfo)
			convey.So(nodeInfo.NodeStatus, convey.ShouldBeEmpty)
		})
	})
}
