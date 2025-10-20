/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package faultclusterprocess is used to return cluster faults

package faultclusterprocess

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/cmprocess"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/interface/grpc/fault"
	"clusterd/pkg/interface/kube"
)

const (
	invalidCacheTime = -5
)

func init() {
	err := hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
	convey.ShouldBeNil(err)
}

func TestGetMaxLevel(t *testing.T) {
	convey.Convey("Given a NodeFaultInfo with various fault devices", t, func() {
		healthy := constant.HealthyState
		subHealthy := constant.SubHealthyState
		unHealthy := constant.UnHealthyState

		convey.Convey("When there are no fault devices", func() {
			faultInfo := fault.NodeFaultInfo{
				FaultDevice: []*fault.DeviceFaultInfo{},
			}
			maxLevel := getMaxLevel(faultInfo)
			convey.So(maxLevel, convey.ShouldEqual, healthy)
		})

		convey.Convey("When all devices are healthy", func() {
			faultInfo := fault.NodeFaultInfo{
				FaultDevice: []*fault.DeviceFaultInfo{
					{FaultLevel: healthy},
					{FaultLevel: healthy},
				},
			}
			maxLevel := getMaxLevel(faultInfo)
			convey.So(maxLevel, convey.ShouldEqual, healthy)
		})

		convey.Convey("When there is an unhealthy device", func() {
			faultInfo := fault.NodeFaultInfo{
				FaultDevice: []*fault.DeviceFaultInfo{
					{FaultLevel: healthy},
					{FaultLevel: unHealthy},
					{FaultLevel: subHealthy},
				},
			}
			maxLevel := getMaxLevel(faultInfo)
			convey.So(maxLevel, convey.ShouldEqual, unHealthy)
		})
	})
}

func TestGetNodeFaultInfo(t *testing.T) {
	convey.Convey("test getNodeFaultInfo", t, func() {
		const (
			testDeviceId   = 1
			testDeviceType = "NPU"
			testFaultCode  = "E1001"
		)

		convey.Convey("while nodeInfo is nil", func() {
			result := getNodeFaultInfo(nil)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("while nodeInfo FaultDevList empty", func() {
			nodeInfo := &constant.NodeInfo{
				NodeInfoNoName: constant.NodeInfoNoName{
					FaultDevList: []*constant.FaultDev{},
					NodeStatus:   "",
				},
				CmName: "",
			}
			result := getNodeFaultInfo(nodeInfo)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("NodeStatus is NotHandle", func() {
			nodeInfo := &constant.NodeInfo{
				NodeInfoNoName: constant.NodeInfoNoName{
					FaultDevList: []*constant.FaultDev{
						&constant.FaultDev{
							DeviceId:   testDeviceId,
							DeviceType: testDeviceType,
							FaultCode:  []string{testFaultCode},
						},
					},
					NodeStatus: "NotHandle",
				},
				CmName: "",
			}
			result := getNodeFaultInfo(nodeInfo)
			convey.So(len(result), convey.ShouldEqual, 1)
			convey.So(result[0].DeviceId, convey.ShouldEqual, strconv.Itoa(int(testDeviceId)))
			convey.So(result[0].DeviceType, convey.ShouldEqual, testDeviceType)
			convey.So(result[0].FaultCodes, convey.ShouldResemble, []string{testFaultCode})
			convey.So(result[0].FaultLevel, convey.ShouldEqual, constant.HealthyState)
		})
	})
}

// TestNewClusterFaultProcessor tests the creation of a new ClusterFaultProcessor instance.
func TestNewClusterFaultProcessor(t *testing.T) {
	convey.Convey("Test NewClusterFaultProcessor func to return a none nil object", t, func() {
		processor := newClusterFaultProcessor()
		convey.So(processor, convey.ShouldNotBeNil)
		convey.So(processor.ClusterFaultCache, convey.ShouldNotBeNil)
		convey.So(processor.ClusterFaultCache.NodeFaultInfo, convey.ShouldHaveLength, 0)
	})
}

// TestClusterFaultProcessorGatherClusterFaultInfo tests the GatherClusterFaultInfo method.
func TestClusterFaultProcessorGatherClusterFaultInfo(t *testing.T) {
	convey.Convey("Test get all fault info while cache is available", t, func() {
		processor := newClusterFaultProcessor()
		processor.lastUpdateTime = time.Now()
		result := processor.GatherClusterFaultInfo()
		convey.So(result, convey.ShouldEqual, processor.ClusterFaultCache)
	})

	convey.Convey("Test get all fault info while cache is available", t, func() {
		processor := newClusterFaultProcessor()
		processor.lastUpdateTime = time.Now().Add(invalidCacheTime * time.Second)
		mockDependencies()
		result := processor.GatherClusterFaultInfo()
		convey.So(result.SignalType, convey.ShouldEqual, constant.SignalTypeNormal)
	})
}

// TestGetAllKindsFaults tests the getAllKindsFaults function.
func TestGetAllKindsFaults(t *testing.T) {
	convey.Convey("Test get npu device fault,got mock fault", t, func() {
		mockDependencies()
		content := constant.AllConfigmapContent{
			DeviceCm: map[string]*constant.AdvanceDeviceFaultCm{
				"device-node1": {
					FaultDeviceList: map[string][]constant.DeviceFault{
						"npu-0": {{FaultCode: "NPU001", FaultLevel: constant.SubHealthFault}},
					},
				},
			},
		}
		faults := getAllKindsFaults(content)
		convey.So(len(faults), convey.ShouldEqual, 1)
		convey.So(faults[0].FaultDevice[0].FaultCodes[0], convey.ShouldEqual, "NPU001")
	})

	convey.Convey("Test gets switch fault", t, func() {
		mockDependencies()
		nodeName := "switch-node1"
		content := constant.AllConfigmapContent{
			DeviceCm: map[string]*constant.AdvanceDeviceFaultCm{constant.DeviceInfoPrefix + nodeName: {}},
			SwitchCm: map[string]*constant.SwitchInfo{
				constant.SwitchInfoPrefix + nodeName: {
					SwitchFaultInfo: constant.SwitchFaultInfo{
						NodeStatus: constant.PreSeparateFault,
						FaultInfo:  []constant.SimpleSwitchFaultInfo{{AssembledFaultCode: "SW001"}},
					},
				},
			},
		}
		faults := getAllKindsFaults(content)
		convey.So(len(faults), convey.ShouldEqual, 1)
		convey.So(faults[0].FaultDevice[0].FaultCodes[0], convey.ShouldEqual, "SW001")
	})
}

// TestGetSwitchFaultInfo tests the getSwitchFaultInfo function
func TestGetSwitchFaultInfo(t *testing.T) {
	convey.Convey("Test with nil switch info with nil input to get nil", t, func() {
		faults := getSwitchFaultInfo(nil)
		convey.So(faults, convey.ShouldBeNil)
	})

	convey.Convey("Test with switch fault, with switch fault input, got switch faults", t, func() {
		switchInfo := &constant.SwitchInfo{
			SwitchFaultInfo: constant.SwitchFaultInfo{
				NodeStatus: constant.PreSeparateFault,
				FaultInfo:  []constant.SimpleSwitchFaultInfo{{AssembledFaultCode: "SW001"}},
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"SW001_0_0": {FaultLevel: constant.PreSeparateFault}},
			},
		}
		faults := getSwitchFaultInfo(switchInfo)
		convey.So(faults[0].FaultCodes[0], convey.ShouldEqual, "SW001")
		convey.So(faults[0].SwitchFaultInfos[0].SwitchChipId, convey.ShouldEqual, "0")
		convey.So(faults[0].SwitchFaultInfos[0].SwitchPortId, convey.ShouldEqual, "0")
		convey.So(faults[0].SwitchFaultInfos[0].FaultLevel, convey.ShouldEqual, constant.PreSeparateFault)
	})
}

// TestGetNpuDeviceFaultInfo tests the getNpuDeviceFaultInfo function.
func TestGetNpuDeviceFaultInfo(t *testing.T) {
	convey.Convey("Test with nil device cm to get nil return", t, func() {
		faults := getNpuDeviceFaultInfo(nil)
		convey.So(faults, convey.ShouldBeNil)
	})

	convey.Convey("Test with device fault, got device fault", t, func() {
		deviceCm := &constant.AdvanceDeviceFaultCm{
			FaultDeviceList: map[string][]constant.DeviceFault{
				"npu-0": {
					{FaultCode: "NPU001", FaultLevel: constant.SubHealthFault,
						FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
							"NPU001": {FaultLevel: constant.SubHealthFault}},
					},
					{FaultCode: "NPU002", FaultLevel: constant.NotHandleFault,
						FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
							"NPU002": {FaultLevel: constant.NotHandleFault}},
					},
				},
			},
		}
		faults := getNpuDeviceFaultInfo(deviceCm)
		convey.So(faults[0].DeviceId, convey.ShouldEqual, "0")
		convey.So(faults[0].FaultCodes[0], convey.ShouldEqual, "NPU001")
		convey.So(faults[0].FaultLevels[0], convey.ShouldEqual, constant.SubHealthFault)
		convey.So(faults[0].FaultCodes[1], convey.ShouldEqual, "NPU002")
		convey.So(faults[0].FaultLevels[1], convey.ShouldEqual, constant.NotHandleFault)
	})
}

// TestGetDeviceFaultInfo tests the getDeviceFaultInfo function.
func TestGetDeviceFaultInfo(t *testing.T) {
	convey.Convey("Test with healthy state", t, func() {
		deviceFaults := []constant.DeviceFault{
			{
				FaultCode: "NPU001", FaultLevel: constant.NotHandleFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"NPU001": {FaultLevel: constant.NotHandleFault}},
			},
			{
				FaultCode: "NPU002", FaultLevel: constant.RestartRequest,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"NPU002": {FaultLevel: constant.RestartRequest}},
			},
		}
		codes, levels, level := getDeviceFaultInfo(deviceFaults)
		convey.So(codes[0], convey.ShouldEqual, "NPU001")
		convey.So(codes[1], convey.ShouldEqual, "NPU002")
		convey.So(levels[0], convey.ShouldEqual, constant.NotHandleFault)
		convey.So(levels[1], convey.ShouldEqual, constant.RestartRequest)
		convey.So(level, convey.ShouldEqual, constant.UnHealthyState)
	})

	convey.Convey("Test with sub healthy state", t, func() {
		deviceFaults := []constant.DeviceFault{
			{FaultCode: "NPU001", FaultLevel: constant.SubHealthFault},
		}
		_, _, level := getDeviceFaultInfo(deviceFaults)
		convey.So(level, convey.ShouldEqual, constant.SubHealthyState)
	})
	convey.Convey("Test with unhealthy state", t, func() {
		deviceFaults := []constant.DeviceFault{
			{FaultCode: "NPU001", FaultLevel: "unknowLevel"},
		}
		_, _, level := getDeviceFaultInfo(deviceFaults)
		convey.So(level, convey.ShouldEqual, constant.UnHealthyState)
	})
}

// mockDependencies sets up mock implementations for external dependencies.
func mockDependencies() {
	gomonkey.ApplyFuncReturn(cmprocess.DeviceCenter.GetProcessedCm, nil)
	gomonkey.ApplyFuncReturn(cmprocess.SwitchCenter.GetProcessedCm, nil)
	gomonkey.ApplyFuncReturn(cmprocess.NodeCenter.GetProcessedCm, nil)
	gomonkey.ApplyFunc(kube.GetNode, func(nodeName string) *v1.Node {
		return &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName,
			},
		}
	})
	gomonkey.ApplyFuncReturn(node.GetNodeIpByName, "192.168.1.1")
	gomonkey.ApplyFuncReturn(node.GetNodeSNByName, "SN-1")
	gomonkey.ApplyFuncReturn(faultdomain.IsNodeReady, true)
}
