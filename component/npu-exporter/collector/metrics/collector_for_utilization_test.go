/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package metrics for general collector
package metrics

import (
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
)

func TestBuildDefaultMultiUtilInfo(t *testing.T) {
	convey.Convey("TestBuildDefaultMultiUtilInfo", t, func() {
		chip := &chipUtilizationCache{}
		buildDefaultMultiUtilInfo(chip)
		convey.So(chip.Utilization, convey.ShouldEqual, defaultUtilValue)
		convey.So(chip.OverallUtilization, convey.ShouldEqual, defaultUtilValue)
		convey.So(chip.VectorUtilization, convey.ShouldEqual, defaultUtilValue)
		convey.So(chip.CubeUtilization, convey.ShouldEqual, defaultUtilValue)
	})
}

type collectUtilTestCase struct {
	name          string
	logicID       int32
	setupPatches  func(*UtilizationCollector, *devmanager.DeviceManager) *gomonkey.Patches
	expectUtil    int
	expectOverall int
	expectVector  int
	expectCube    int
}

func buildCollectUtilTestCases() []collectUtilTestCase {
	return []collectUtilTestCase{
		{
			name:    "should call realGetDeviceUtilizationRateInfoFunc when it is not nil",
			logicID: testLogicID0,
			setupPatches: func(c *UtilizationCollector, dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				c.realGetDeviceUtilizationRateInfoFunc = collectUtilCommon
				return gomonkey.ApplyMethodReturn(dmgr, "GetDeviceUtilizationRateCommon",
					common.DcmiMultiUtilizationInfo{
						AicUtil:    testAicUtil,
						AivUtil:    testAivUtil,
						AicoreUtil: testAicoreUtil,
						NpuUtil:    testNpuUtil,
					}, nil)
			},
			expectUtil:    int(testAicoreUtil),
			expectOverall: int(testNpuUtil),
			expectVector:  int(testAivUtil),
			expectCube:    int(testAicUtil),
		},
		{
			name:    "should call buildDefaultMultiUtilInfo when func is nil",
			logicID: testLogicID0,
			setupPatches: func(c *UtilizationCollector, dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				c.realGetDeviceUtilizationRateInfoFunc = nil
				return gomonkey.NewPatches()
			},
			expectUtil:    defaultUtilValue,
			expectOverall: defaultUtilValue,
			expectVector:  defaultUtilValue,
			expectCube:    defaultUtilValue,
		},
	}
}

func TestCollectUtil(t *testing.T) {
	convey.Convey("TestCollectUtil", t, func() {
		for _, tt := range buildCollectUtilTestCases() {
			convey.Convey(tt.name, func() {
				dmgr := &devmanager.DeviceManager{}
				c := &UtilizationCollector{}
				chip := &chipUtilizationCache{}
				var patches *gomonkey.Patches
				if tt.setupPatches != nil {
					patches = tt.setupPatches(c, dmgr)
					defer patches.Reset()
				}
				collectUtil(c, tt.logicID, dmgr, chip)
				convey.So(chip.Utilization, convey.ShouldEqual, tt.expectUtil)
				convey.So(chip.OverallUtilization, convey.ShouldEqual, tt.expectOverall)
				convey.So(chip.VectorUtilization, convey.ShouldEqual, tt.expectVector)
				convey.So(chip.CubeUtilization, convey.ShouldEqual, tt.expectCube)
			})
		}
	})
}

type collectUtilV1TestCase struct {
	name          string
	logicID       int32
	devType       string
	setupPatches  func(*devmanager.DeviceManager) *gomonkey.Patches
	expectUtil    int
	expectOverall int
	expectVector  int
	expectCube    int
}

func buildCollectUtilV1TestCases() []collectUtilV1TestCase {
	return []collectUtilV1TestCase{
		{
			name:    "should collect utilizations when device supports vector and overall",
			logicID: testLogicID0,
			devType: common.Ascend910B,
			setupPatches: func(dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				patches := gomonkey.NewPatches()
				patches.ApplyMethodReturn(dmgr, "GetDevType", common.Ascend910B)
				patches.ApplyMethod(dmgr, "GetDeviceUtilizationRate",
					func(_ *devmanager.DeviceManager, _ int32, devType common.DeviceType) (uint32, error) {
						if devType == common.AICore {
							return testAICoreUtil, nil
						}
						if devType == common.VectorCore {
							return testVectorUtil, nil
						}
						if devType == common.Overall {
							return testOverallUtil, nil
						}
						return uint32(0), nil
					})
				return patches
			},
			expectUtil:    int(testAICoreUtil),
			expectOverall: int(testOverallUtil),
			expectVector:  int(testVectorUtil),
			expectCube:    defaultUtilValue,
		},
		{
			name:    "should not collect vector when device does not support it",
			logicID: testLogicID0,
			devType: common.Ascend910,
			setupPatches: func(dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				patches := gomonkey.NewPatches()
				patches.ApplyMethodReturn(dmgr, "GetDevType", common.Ascend910)
				patches.ApplyMethodReturn(dmgr, "GetDeviceUtilizationRate",
					testAICoreUtil, nil)
				return patches
			},
			expectUtil:    int(testAICoreUtil),
			expectOverall: defaultUtilValue,
			expectVector:  defaultUtilValue,
			expectCube:    defaultUtilValue,
		},
	}
}

func TestCollectUtilV1(t *testing.T) {
	convey.Convey("TestCollectUtilV1", t, func() {
		for _, tt := range buildCollectUtilV1TestCases() {
			convey.Convey(tt.name, func() {
				dmgr := &devmanager.DeviceManager{}
				chip := &chipUtilizationCache{}
				var patches *gomonkey.Patches
				if tt.setupPatches != nil {
					patches = tt.setupPatches(dmgr)
					defer patches.Reset()
				}
				collectUtilV1(tt.logicID, dmgr, chip)
				convey.So(chip.Utilization, convey.ShouldEqual, tt.expectUtil)
				convey.So(chip.OverallUtilization, convey.ShouldEqual, tt.expectOverall)
				convey.So(chip.VectorUtilization, convey.ShouldEqual, tt.expectVector)
				convey.So(chip.CubeUtilization, convey.ShouldEqual, tt.expectCube)
			})
		}
	})
}

type collectUtilCommonTestCase struct {
	name          string
	logicID       int32
	setupPatches  func(*devmanager.DeviceManager) *gomonkey.Patches
	expectUtil    int
	expectOverall int
	expectVector  int
	expectCube    int
	expectError   bool
}

func buildcollectUtilCommonTestCases() []collectUtilCommonTestCase {
	return []collectUtilCommonTestCase{
		{
			name:    "should collect all utilizations successfully when api succeeds",
			logicID: testLogicID0,
			setupPatches: func(dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				return gomonkey.ApplyMethodReturn(dmgr, "GetDeviceUtilizationRateCommon",
					common.DcmiMultiUtilizationInfo{
						AicUtil:    testAicUtil,
						AivUtil:    testAivUtil,
						AicoreUtil: testAicoreUtil,
						NpuUtil:    testNpuUtil,
					}, nil)
			},
			expectUtil:    int(testAicoreUtil),
			expectOverall: int(testNpuUtil),
			expectVector:  int(testAivUtil),
			expectCube:    int(testAicUtil),
			expectError:   false,
		},
		{
			name:    "should set zero values when api fails",
			logicID: testLogicID0,
			setupPatches: func(dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				return gomonkey.ApplyMethodReturn(dmgr, "GetDeviceUtilizationRateCommon",
					common.DcmiMultiUtilizationInfo{}, errors.New(apiCallFailedMsg))
			},
			expectUtil:    0,
			expectOverall: 0,
			expectVector:  0,
			expectCube:    0,
			expectError:   true,
		},
	}
}

func TestCollectUtilCommon(t *testing.T) {
	convey.Convey("TestCollectUtilCommon", t, func() {
		for _, tt := range buildcollectUtilCommonTestCases() {
			convey.Convey(tt.name, func() {
				dmgr := &devmanager.DeviceManager{}
				chip := &chipUtilizationCache{}
				var patches *gomonkey.Patches
				if tt.setupPatches != nil {
					patches = tt.setupPatches(dmgr)
					defer patches.Reset()
				}
				collectUtilCommon(tt.logicID, dmgr, chip)
				convey.So(chip.Utilization, convey.ShouldEqual, tt.expectUtil)
				convey.So(chip.OverallUtilization, convey.ShouldEqual, tt.expectOverall)
				convey.So(chip.VectorUtilization, convey.ShouldEqual, tt.expectVector)
				convey.So(chip.CubeUtilization, convey.ShouldEqual, tt.expectCube)
			})
		}
	})
}

func TestUpdateContainerUtilization(t *testing.T) {
	cardLabel := []string{"0", "npu", "die", "pcie", "ns", "pod", "container"}

	convey.Convey("TestUpdateContainerUtilization", t, func() {
		convey.Convey("should report a single metric when container name splits into 3 parts", func() {
			chip := &chipUtilizationCache{Utilization: 80, timestamp: time.Now()}
			containerInfo := container.DevicesInfo{ID: "c1", Name: "ns1_pod1_container1"}

			metrics := collectContainerUtilizationMetrics(chip, containerInfo, cardLabel, colcommon.HuaWeiAIChip{})

			convey.So(metrics, convey.ShouldHaveLength, 1)
			labels, value := readMetricLabelsAndValue(metrics[0])
			convey.So(labels["namespace"], convey.ShouldEqual, "ns")
			convey.So(labels["pod_name"], convey.ShouldEqual, "pod")
			convey.So(labels["container_name"], convey.ShouldEqual, "container")
			convey.So(value, convey.ShouldEqual, float64(80))
		})

		convey.Convey("should not report when container name does not split into 3 parts", func() {
			chip := &chipUtilizationCache{Utilization: 80, timestamp: time.Now()}
			containerInfo := container.DevicesInfo{ID: "c1", Name: "short"}

			metrics := collectContainerUtilizationMetrics(chip, containerInfo, cardLabel, colcommon.HuaWeiAIChip{})

			convey.So(metrics, convey.ShouldHaveLength, 0)
		})

		convey.Convey("should not report when container name is empty", func() {
			chip := &chipUtilizationCache{Utilization: 80, timestamp: time.Now()}
			containerInfo := container.DevicesInfo{}

			metrics := collectContainerUtilizationMetrics(chip, containerInfo, cardLabel, colcommon.HuaWeiAIChip{})

			convey.So(metrics, convey.ShouldHaveLength, 0)
		})

		convey.Convey("should not report when utilization value is invalid", func() {
			chip := &chipUtilizationCache{Utilization: -1, timestamp: time.Now()}
			containerInfo := container.DevicesInfo{ID: "c1", Name: "ns1_pod1_container1"}

			metrics := collectContainerUtilizationMetrics(chip, containerInfo, cardLabel, colcommon.HuaWeiAIChip{})

			convey.So(metrics, convey.ShouldHaveLength, 0)
		})

		convey.Convey("should not report for vnpu virtual device", func() {
			chip := &chipUtilizationCache{Utilization: 80, timestamp: time.Now()}
			containerInfo := container.DevicesInfo{ID: "c1", Name: "ns1_pod1_container1"}

			metrics := collectContainerUtilizationMetrics(chip, containerInfo, cardLabel, colcommon.HuaWeiAIChip{
				VDevActivityInfo: &common.VDevActivityInfo{VDevID: 100, IsVirtualDev: true},
			})

			convey.So(metrics, convey.ShouldHaveLength, 0)
		})
	})
}

func collectContainerUtilizationMetrics(chip *chipUtilizationCache, containerInfo container.DevicesInfo,
	cardLabel []string, chipWithVnpu colcommon.HuaWeiAIChip) []prometheus.Metric {
	ch := make(chan prometheus.Metric, 1)
	go func() {
		defer close(ch)
		updateContainerUtilization(ch, containerInfo, cardLabel, chip, chipWithVnpu)
	}()
	metrics := make([]prometheus.Metric, 0, 1)
	for m := range ch {
		if m.Desc() == npuCtrUtilization {
			metrics = append(metrics, m)
		}
	}
	return metrics
}
