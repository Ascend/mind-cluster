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

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
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
