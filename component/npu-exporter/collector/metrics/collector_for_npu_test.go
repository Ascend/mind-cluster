/* Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
)

const (
	testLogicID0     = int32(0)
	defaultUtilValue = -1
	testAicAvgUtil   = uint32(50)
	testAivAvgUtil   = uint32(60)
	testAicoreUtil   = uint32(70)
	testNpuUtil      = uint32(80)
	testAICoreUtil   = uint32(75)
	testVectorUtil   = uint32(65)
	testOverallUtil  = uint32(85)
	apiCallFailedMsg = "api call failed"
)

func TestIsSupportNetworkHealthDevices(t *testing.T) {
	convey.Convey("TestIsSupportNetworkHealthDevices", t, func() {
		result := isSupportNetworkHealthDevices(api.Ascend910A3, 0)
		convey.So(result, convey.ShouldEqual, true)
		result = isSupportNetworkHealthDevices(api.Ascend910A5, api.Atlas9501DMainBoardID)
		convey.So(result, convey.ShouldEqual, true)
		result = isSupportNetworkHealthDevices(api.Ascend910A5, api.Atlas3504PMainBoardID)
		convey.So(result, convey.ShouldEqual, false)
	})
}

func TestBuildDefaultMultiUtilInfo(t *testing.T) {
	convey.Convey("TestBuildDefaultMultiUtilInfo", t, func() {
		chip := &chipCache{}
		buildDefaultMultiUtilInfo(chip)
		convey.So(chip.Utilization, convey.ShouldEqual, defaultUtilValue)
		convey.So(chip.OverallUtilization, convey.ShouldEqual, defaultUtilValue)
		convey.So(chip.VectorUtilization, convey.ShouldEqual, defaultUtilValue)
		convey.So(chip.CubeUtilization, convey.ShouldEqual, defaultUtilValue)
	})
}

type baseInfoPreCollectTestCase struct {
	name         string
	devType      string
	chipList     []colcommon.HuaWeiAIChip
	setupPatches func(*devmanager.DeviceManager) *gomonkey.Patches
	expectFunc   func(*BaseInfoCollector) bool
}

func buildBaseInfoPreCollectTestCases() []baseInfoPreCollectTestCase {
	cases := buildBaseInfoPreCollectBasicTestCases()
	cases = append(cases, buildBaseInfoPreCollectV2SuccessTestCases()...)
	cases = append(cases, buildBaseInfoPreCollectV1FallbackTestCases()...)
	cases = append(cases, buildBaseInfoPreCollectBothFailTestCases()...)
	return cases
}

func buildBaseInfoPreCollectBasicTestCases() []baseInfoPreCollectTestCase {
	return []baseInfoPreCollectTestCase{
		{
			name:     "should use v1 api when devType is not 910B or 910A3",
			devType:  common.Ascend910,
			chipList: []colcommon.HuaWeiAIChip{{LogicID: testLogicID0}},
			setupPatches: func(dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				return gomonkey.ApplyMethodReturn(dmgr, "GetDevType", common.Ascend910)
			},
			expectFunc: func(c *BaseInfoCollector) bool {
				return c.realGetDeviceUtilizationRateInfoFunc != nil &&
					reflect.ValueOf(c.realGetDeviceUtilizationRateInfoFunc).Pointer() ==
						reflect.ValueOf(collectUtilV1).Pointer()
			},
		},
		{
			name:     "should return early when chipList is empty",
			devType:  common.Ascend910B,
			chipList: []colcommon.HuaWeiAIChip{},
			setupPatches: func(dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				return gomonkey.ApplyMethodReturn(dmgr, "GetDevType", common.Ascend910B)
			},
			expectFunc: func(c *BaseInfoCollector) bool {
				return c.realGetDeviceUtilizationRateInfoFunc != nil &&
					reflect.ValueOf(c.realGetDeviceUtilizationRateInfoFunc).Pointer() ==
						reflect.ValueOf(collectUtilV1).Pointer()
			},
		},
	}
}

func buildBaseInfoPreCollectV2SuccessTestCases() []baseInfoPreCollectTestCase {
	return []baseInfoPreCollectTestCase{
		{
			name:     "should use v2 api when GetDeviceUtilizationRateV2 succeeds",
			devType:  common.Ascend910B,
			chipList: []colcommon.HuaWeiAIChip{{LogicID: testLogicID0}},
			setupPatches: func(dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				patches := gomonkey.NewPatches()
				patches.ApplyMethodReturn(dmgr, "GetDevType", common.Ascend910B)
				patches.ApplyMethodReturn(dmgr, "GetDeviceUtilizationRateV2",
					common.NpuMultiUtilizationInfo{}, nil)
				return patches
			},
			expectFunc: func(c *BaseInfoCollector) bool {
				return c.realGetDeviceUtilizationRateInfoFunc != nil &&
					reflect.ValueOf(c.realGetDeviceUtilizationRateInfoFunc).Pointer() ==
						reflect.ValueOf(collectUtilV2).Pointer()
			},
		},
	}
}

func buildBaseInfoPreCollectV1FallbackTestCases() []baseInfoPreCollectTestCase {
	return []baseInfoPreCollectTestCase{
		{
			name:     "should use v1 api when v2 fails but v1 succeeds",
			devType:  common.Ascend910B,
			chipList: []colcommon.HuaWeiAIChip{{LogicID: testLogicID0}},
			setupPatches: func(dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				patches := gomonkey.NewPatches()
				patches.ApplyMethodReturn(dmgr, "GetDevType", common.Ascend910B)
				patches.ApplyMethodReturn(dmgr, "GetDeviceUtilizationRateV2",
					common.NpuMultiUtilizationInfo{}, errors.New(apiCallFailedMsg))
				patches.ApplyMethodReturn(dmgr, "GetDeviceUtilizationRate", testAICoreUtil, nil)
				return patches
			},
			expectFunc: func(c *BaseInfoCollector) bool {
				return c.realGetDeviceUtilizationRateInfoFunc != nil &&
					reflect.ValueOf(c.realGetDeviceUtilizationRateInfoFunc).Pointer() ==
						reflect.ValueOf(collectUtilV1).Pointer()
			},
		},
	}
}

func buildBaseInfoPreCollectBothFailTestCases() []baseInfoPreCollectTestCase {
	return []baseInfoPreCollectTestCase{
		{
			name:     "should set func to nil when both v2 and v1 fail after retries",
			devType:  common.Ascend910B,
			chipList: []colcommon.HuaWeiAIChip{{LogicID: testLogicID0}},
			setupPatches: func(dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				patches := gomonkey.NewPatches()
				patches.ApplyMethodReturn(dmgr, "GetDevType", common.Ascend910B)
				patches.ApplyMethodReturn(dmgr, "GetDeviceUtilizationRateV2",
					common.NpuMultiUtilizationInfo{}, errors.New(apiCallFailedMsg))
				patches.ApplyMethodReturn(dmgr, "GetDeviceUtilizationRate",
					uint32(0), errors.New(apiCallFailedMsg))
				return patches
			},
			expectFunc: func(c *BaseInfoCollector) bool {
				return c.realGetDeviceUtilizationRateInfoFunc == nil
			},
		},
	}
}

func TestBaseInfoCollectorPreCollect(t *testing.T) {
	convey.Convey("TestBaseInfoCollectorPreCollect", t, func() {
		for _, tt := range buildBaseInfoPreCollectTestCases() {
			convey.Convey(tt.name, func() {
				dmgr := &devmanager.DeviceManager{}
				var patches *gomonkey.Patches
				if tt.setupPatches != nil {
					patches = tt.setupPatches(dmgr)
					defer patches.Reset()
				}
				n := &colcommon.NpuCollector{Dmgr: dmgr}
				c := &BaseInfoCollector{}
				c.PreCollect(n, tt.chipList)
				convey.So(tt.expectFunc(c), convey.ShouldBeTrue)
			})
		}
	})
}

type collectUtilTestCase struct {
	name          string
	logicID       int32
	setupPatches  func(*BaseInfoCollector, *devmanager.DeviceManager) *gomonkey.Patches
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
			setupPatches: func(c *BaseInfoCollector, dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				c.realGetDeviceUtilizationRateInfoFunc = collectUtilV2
				return gomonkey.ApplyMethodReturn(dmgr, "GetDeviceUtilizationRateV2",
					common.NpuMultiUtilizationInfo{
						AicAvgUtil: testAicAvgUtil,
						AivAvgUtil: testAivAvgUtil,
						AicoreUtil: testAicoreUtil,
						NpuUtil:    testNpuUtil,
					}, nil)
			},
			expectUtil:    int(testAicoreUtil),
			expectOverall: int(testNpuUtil),
			expectVector:  int(testAivAvgUtil),
			expectCube:    int(testAicAvgUtil),
		},
		{
			name:    "should call buildDefaultMultiUtilInfo when func is nil",
			logicID: testLogicID0,
			setupPatches: func(c *BaseInfoCollector, dmgr *devmanager.DeviceManager) *gomonkey.Patches {
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
				c := &BaseInfoCollector{}
				chip := &chipCache{}
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
				chip := &chipCache{}
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

type collectUtilV2TestCase struct {
	name          string
	logicID       int32
	setupPatches  func(*devmanager.DeviceManager) *gomonkey.Patches
	expectUtil    int
	expectOverall int
	expectVector  int
	expectCube    int
	expectError   bool
}

func buildCollectUtilV2TestCases() []collectUtilV2TestCase {
	return []collectUtilV2TestCase{
		{
			name:    "should collect all utilizations successfully when api succeeds",
			logicID: testLogicID0,
			setupPatches: func(dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				return gomonkey.ApplyMethodReturn(dmgr, "GetDeviceUtilizationRateV2",
					common.NpuMultiUtilizationInfo{
						AicAvgUtil: testAicAvgUtil,
						AivAvgUtil: testAivAvgUtil,
						AicoreUtil: testAicoreUtil,
						NpuUtil:    testNpuUtil,
					}, nil)
			},
			expectUtil:    int(testAicoreUtil),
			expectOverall: int(testNpuUtil),
			expectVector:  int(testAivAvgUtil),
			expectCube:    int(testAicAvgUtil),
			expectError:   false,
		},
		{
			name:    "should set zero values when api fails",
			logicID: testLogicID0,
			setupPatches: func(dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				return gomonkey.ApplyMethodReturn(dmgr, "GetDeviceUtilizationRateV2",
					common.NpuMultiUtilizationInfo{}, errors.New(apiCallFailedMsg))
			},
			expectUtil:    0,
			expectOverall: 0,
			expectVector:  0,
			expectCube:    0,
			expectError:   true,
		},
	}
}

func TestCollectUtilV2(t *testing.T) {
	convey.Convey("TestCollectUtilV2", t, func() {
		for _, tt := range buildCollectUtilV2TestCases() {
			convey.Convey(tt.name, func() {
				dmgr := &devmanager.DeviceManager{}
				chip := &chipCache{}
				var patches *gomonkey.Patches
				if tt.setupPatches != nil {
					patches = tt.setupPatches(dmgr)
					defer patches.Reset()
				}
				collectUtilV2(tt.logicID, dmgr, chip)
				convey.So(chip.Utilization, convey.ShouldEqual, tt.expectUtil)
				convey.So(chip.OverallUtilization, convey.ShouldEqual, tt.expectOverall)
				convey.So(chip.VectorUtilization, convey.ShouldEqual, tt.expectVector)
				convey.So(chip.CubeUtilization, convey.ShouldEqual, tt.expectCube)
			})
		}
	})
}
