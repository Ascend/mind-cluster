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

// Package app unit tests for reset manager
package app

import (
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	devcommon "ascend-common/devmanager/common"
	"container-manager/pkg/common"
	containerdomain "container-manager/pkg/container/domain"
	"container-manager/pkg/devmgr"
	faultdomain "container-manager/pkg/fault/domain"
	"container-manager/pkg/reset/domain"
)

var mockGetLogicIdByPhyId = func(phyId int32) int32 { return phyId }

// TestMain init log
func TestMain(m *testing.M) {
	err := hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	if err != nil {
		panic(err)
	}
	m.Run()
}

// TestResetMgr_AllowToResetNpu tests the allowToResetNpu method
func TestResetMgr_AllowToResetNpu(t *testing.T) {
	convey.Convey("Test AllowToResetNpu", t, func() {
		r := &ResetMgr{
			resetCache: domain.NewNpuInResetCache(),
		}

		convey.Convey("When reset cache is not empty", func() {
			r.resetCache.SetNpuInReset(1)
			result := r.allowToResetNpu()
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When in cooldown period", func() {
			const mockPreviousTime = -time.Second * 10
			lastTime := time.Now().Add(mockPreviousTime)
			r.lastSuccessResetTime = &lastTime
			result := r.allowToResetNpu()
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When allowed to reset", func() {
			result := r.allowToResetNpu()
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

// TestResetMgr_ResetResetCountCache tests the resetResetCountCache method
func TestResetMgr_ResetResetCountCache(t *testing.T) {
	convey.Convey("Test ResetResetCountCache", t, func() {
		r := &ResetMgr{
			countCache: domain.NewFailedResetCountCache(),
		}
		testId1, testId1Count := int32(1), 3
		testId2, testId2Count := int32(2), 1

		convey.Convey("When there are npus with no faults", func() {

			testNpus := map[int32]int{testId1: testId1Count, testId2: testId2Count}
			for id, count := range testNpus {
				r.countCache.SetFailedResetCount(id, count)
			}
			faultsMap := map[int32]map[int64]map[string]*common.DevFaultInfo{1: {}}
			r.resetResetCountCache(faultsMap)
			count1 := r.countCache.GetFailedResetCount(testId1)
			convey.So(count1, convey.ShouldEqual, testId1Count)
			count2 := r.countCache.GetFailedResetCount(testId2)
			convey.So(count2, convey.ShouldEqual, 0)
		})

		convey.Convey("When all npus have faults", func() {
			r.countCache.SetFailedResetCount(testId1, testId1Count)
			faultsMap := map[int32]map[int64]map[string]*common.DevFaultInfo{
				1: {},
			}
			r.resetResetCountCache(faultsMap)
			count := r.countCache.GetFailedResetCount(testId1)
			convey.So(count, convey.ShouldEqual, testId1Count)
		})
	})
}

// TestResetMgr_FilterCountLimit tests the filterCountLimit method
func TestResetMgr_FilterCountLimit(t *testing.T) {
	convey.Convey("Test FilterCountLimit", t, func() {
		r := &ResetMgr{
			countCache: domain.NewFailedResetCountCache(),
		}
		testId1 := int32(1)
		testId2 := int32(2)

		convey.Convey("When npu reset count exceeds limit", func() {
			testId1Count := 3
			r.countCache.SetFailedResetCount(testId1, testId1Count)
			faultNpus := []int32{testId1, testId2}
			result := r.filterCountLimit(faultNpus)
			convey.So(result, convey.ShouldResemble, []int32{testId2})
		})

		convey.Convey("When no npu exceeds limit", func() {
			r.countCache.SetFailedResetCount(testId1, 0)
			faultNpus := []int32{testId1, testId2}
			result := r.filterCountLimit(faultNpus)
			convey.So(result, convey.ShouldResemble, []int32{testId1, testId2})
		})
	})
}

// TestGetAllRelatedNpus tests the getAllRelatedNpus function
func TestGetAllRelatedNpus(t *testing.T) {
	convey.Convey("Test getAllRelatedNpus", t, func() {
		mockNpuInfo := map[int32]*common.NPUInfo{
			0: {DevsOnRing: []int32{0, 1, 2, 3}},
			1: {DevsOnRing: []int32{0, 1, 2, 3}},
			4: {DevsOnRing: []int32{4, 5, 6, 7}},
		}
		convey.Convey("When successfully get related npus", func() {
			patches := gomonkey.ApplyMethodReturn(devmgr.DevMgr, "GetNodeNPUInfo", mockNpuInfo)
			defer patches.Reset()
			faultNpus := []int32{0, 4}
			result := getAllRelatedNpus(faultNpus)
			expected := []domain.ResetNpuInfos{
				{FaultId: 0, RelatedIds: []int32{0, 1, 2, 3}},
				{FaultId: 4, RelatedIds: []int32{4, 5, 6, 7}},
			}
			convey.So(result, convey.ShouldResemble, expected)
		})

		convey.Convey("When GetPhyIdOnRing returns error", func() {
			patches := gomonkey.ApplyMethodReturn(devmgr.DevMgr, "GetNodeNPUInfo", map[int32]*common.NPUInfo{})
			defer patches.Reset()
			faultNpus := []int32{1}
			result := getAllRelatedNpus(faultNpus)
			convey.So(result, convey.ShouldBeEmpty)
		})

		convey.Convey("When duplicate npus should be filtered", func() {
			patches := gomonkey.ApplyMethodReturn(devmgr.DevMgr, "GetNodeNPUInfo", mockNpuInfo)
			defer patches.Reset()
			faultNpus := []int32{0, 1}
			result := getAllRelatedNpus(faultNpus)
			expected := []domain.ResetNpuInfos{
				{FaultId: 0, RelatedIds: []int32{0, 1, 2, 3}},
			}
			convey.So(result, convey.ShouldResemble, expected)
		})
	})
}

// TestIsNpuHoldByContainer tests the isNpuHoldByContainer function
func TestIsNpuHoldByContainer(t *testing.T) {
	convey.Convey("Test isNpuHoldByContainer", t, func() {
		convey.Convey("When npu is held by container", func() {
			patches := gomonkey.ApplyMethodFunc(containerdomain.GetDevCache(), "GetDevsRelatedCtrs", func(phyId int32) []string {
				return []string{"container1", "container2"}
			})
			defer patches.Reset()

			result := isNpuHoldByContainer([]int32{1, 2})
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("When npu is not held by container", func() {
			patches := gomonkey.ApplyMethodFunc(containerdomain.GetDevCache(), "GetDevsRelatedCtrs", func(phyId int32) []string {
				return []string{}
			})
			defer patches.Reset()

			result := isNpuHoldByContainer([]int32{1, 2})
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

// TestIsNpuHoldByProcess_Held tests when npu is held by process
func TestIsNpuHoldByProcess_Held(t *testing.T) {
	convey.Convey("When npu is held by process", t, func() {
		mockDevProcessInfo := devcommon.DevProcessInfo{
			ProcNum: 2,
			DevProcArray: []devcommon.DevProcInfo{
				{Pid: 1234},
				{Pid: 5678},
			},
		}
		patches := gomonkey.ApplyMethodFunc(devmgr.DevMgr, "GetLogicIdByPhyId", mockGetLogicIdByPhyId).
			ApplyMethodReturn(devmgr.DevMgr, "GetDmgr", &devmanager.DeviceManager{}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "GetDevProcessInfo",
				func(logicId int32) (*devcommon.DevProcessInfo, error) {
					return &mockDevProcessInfo, nil
				})
		defer patches.Reset()

		result, err := isNpuHoldByProcess([]int32{1})
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldBeTrue)
	})
}

// TestIsNpuHoldByProcess_NotHeld tests when npu is not held by process
func TestIsNpuHoldByProcess_NotHeld(t *testing.T) {
	convey.Convey("When npu is not held by process", t, func() {
		patches := gomonkey.ApplyMethodFunc(devmgr.DevMgr, "GetLogicIdByPhyId", mockGetLogicIdByPhyId).
			ApplyMethodReturn(devmgr.DevMgr, "GetDmgr", &devmanager.DeviceManager{}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "GetDevProcessInfo",
				func(logicId int32) (*devcommon.DevProcessInfo, error) {
					return &devcommon.DevProcessInfo{ProcNum: 0}, nil
				})
		defer patches.Reset()

		result, err := isNpuHoldByProcess([]int32{1})
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldBeFalse)
	})
}

// TestIsNpuHoldByProcess_GetProcessInfoError tests when GetDevProcessInfo returns error
func TestIsNpuHoldByProcess_GetProcessInfoError(t *testing.T) {
	convey.Convey("When GetDevProcessInfo returns error", t, func() {
		patches := gomonkey.ApplyMethodFunc(devmgr.DevMgr, "GetLogicIdByPhyId", mockGetLogicIdByPhyId).
			ApplyMethodReturn(devmgr.DevMgr, "GetDmgr", &devmanager.DeviceManager{}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "GetDevProcessInfo",
				func(logicId int32) (*devcommon.DevProcessInfo, error) {
					return nil, fmt.Errorf("mock error")
				})
		defer patches.Reset()

		result, err := isNpuHoldByProcess([]int32{1})
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(result, convey.ShouldBeFalse)
	})
}

// TestIsFaultExist_ExistsAndNeedsHandling tests scenario where fault exists and needs handling
func TestIsFaultExist_ExistsAndNeedsHandling(t *testing.T) {
	convey.Convey("When fault exists and needs handling", t, func() {
		patches := gomonkey.ApplyMethodFunc(devmgr.DevMgr, "GetLogicIdByPhyId", mockGetLogicIdByPhyId).
			ApplyMethodReturn(devmgr.DevMgr, "GetDmgr", &devmanager.DeviceManager{}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "GetDeviceAllErrorCode",
				func(logicId int32) (int32, []int64, error) {
					return 0, []int64{1001, 1002}, nil
				})
		patches.ApplyFunc(faultdomain.GetFaultLevelByCode, func(codes []int64) string {
			return common.RestartNPU
		})
		defer patches.Reset()

		result := isFaultExist([]int32{1})
		convey.So(result, convey.ShouldBeTrue)
	})
}

// TestIsFaultExist_ExistsButNoNeedHandling tests scenario where fault exists but doesn't need handling
func TestIsFaultExist_ExistsButNoNeedHandling(t *testing.T) {
	convey.Convey("When fault exists but doesn't need handling", t, func() {
		patches := gomonkey.ApplyMethodFunc(devmgr.DevMgr, "GetLogicIdByPhyId", mockGetLogicIdByPhyId).
			ApplyMethodReturn(devmgr.DevMgr, "GetDmgr", &devmanager.DeviceManager{}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "GetDeviceAllErrorCode",
				func(logicId int32) (int32, []int64, error) {
					return 0, []int64{3001}, nil
				})
		patches.ApplyFunc(faultdomain.GetFaultLevelByCode, func(codes []int64) string {
			return "L1"
		})
		defer patches.Reset()

		result := isFaultExist([]int32{1})
		convey.So(result, convey.ShouldBeFalse)
	})
}

// TestIsFaultExist_NoFault tests scenario with no fault
func TestIsFaultExist_NoFault(t *testing.T) {
	convey.Convey("When no fault exists", t, func() {
		patches := gomonkey.ApplyMethodFunc(devmgr.DevMgr, "GetLogicIdByPhyId", mockGetLogicIdByPhyId).
			ApplyMethodReturn(devmgr.DevMgr, "GetDmgr", &devmanager.DeviceManager{}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "GetDeviceAllErrorCode",
				func(logicId int32) (int32, []int64, error) {
					return 0, []int64{}, nil
				})
		defer patches.Reset()

		result := isFaultExist([]int32{1})
		convey.So(result, convey.ShouldBeFalse)
	})
}

// TestIsFaultExist_GetErrorCodeFails tests scenario where GetDeviceAllErrorCode fails
func TestIsFaultExist_GetErrorCodeFails(t *testing.T) {
	convey.Convey("When GetDeviceAllErrorCode returns error", t, func() {
		patches := gomonkey.ApplyMethodFunc(devmgr.DevMgr, "GetLogicIdByPhyId", mockGetLogicIdByPhyId).
			ApplyMethodReturn(devmgr.DevMgr, "GetDmgr", &devmanager.DeviceManager{}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "GetDeviceAllErrorCode",
				func(logicId int32) (int32, []int64, error) {
					return 0, nil, fmt.Errorf("mock error")
				})
		defer patches.Reset()
		result := isFaultExist([]int32{1})
		convey.So(result, convey.ShouldBeFalse)
	})
}
