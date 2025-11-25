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
	"errors"
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
			}).ApplyFuncReturn(containerdomain.GetDevCache, &containerdomain.DevCache{})
			defer patches.Reset()

			result := isNpuHoldByContainer([]int32{1, 2})
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("When npu is not held by container", func() {
			patches := gomonkey.ApplyMethodFunc(containerdomain.GetDevCache(), "GetDevsRelatedCtrs", func(phyId int32) []string {
				return []string{}
			}).ApplyFuncReturn(containerdomain.GetDevCache, &containerdomain.DevCache{})
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
				}).
			ApplyFunc(faultdomain.GetFaultLevelByCode, func(codes []int64) string {
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
				}).
			ApplyFunc(faultdomain.GetFaultLevelByCode, func(codes []int64) string {
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

// TestResetMgr_ProcessResetWork_AllowFalse tests scenario where allowToResetNpu returns false
func TestResetMgr_ProcessResetWork_AllowFalse(t *testing.T) {
	convey.Convey("When allowToResetNpu returns false", t, func() {
		r := &ResetMgr{
			resetCache: domain.NewNpuInResetCache(),
			countCache: domain.NewFailedResetCountCache(),
		}
		patches := gomonkey.ApplyPrivateMethod(r, "allowToResetNpu", func(*ResetMgr) bool {
			return false
		})
		defer patches.Reset()

		r.processResetWork()
		convey.So(r.countCache.GetAllFailedResetCountNpuId(), convey.ShouldBeEmpty)
	})
}

// TestResetMgr_ProcessResetWork_GetFaultCacheFails tests scenario where getFaultCache fails
func TestResetMgr_ProcessResetWork_GetFaultCacheFails(t *testing.T) {
	convey.Convey("When getFaultCache fails", t, func() {
		r := &ResetMgr{
			resetCache: domain.NewNpuInResetCache(),
			countCache: domain.NewFailedResetCountCache(),
		}
		patches := gomonkey.ApplyPrivateMethod(r, "allowToResetNpu", func() bool {
			return true
		}).
			ApplyFunc(getFaultCache, func() (map[int32]map[int64]map[string]*common.DevFaultInfo, error) {
				return nil, errors.New("mock error")
			})
		defer patches.Reset()

		r.processResetWork()
		convey.So(r.countCache.GetAllFailedResetCountNpuId(), convey.ShouldBeEmpty)
	})
}

// TestResetMgr_ProcessResetWork_NoFaults tests scenario with no faults to reset
func TestResetMgr_ProcessResetWork_NoFaults(t *testing.T) {
	convey.Convey("When there are no faults to reset", t, func() {
		r := &ResetMgr{
			resetCache: domain.NewNpuInResetCache(),
			countCache: domain.NewFailedResetCountCache(),
		}
		patches := gomonkey.ApplyPrivateMethod(r, "allowToResetNpu", func() bool {
			return true
		}).ApplyFunc(getFaultCache, func() (map[int32]map[int64]map[string]*common.DevFaultInfo, error) {
			return map[int32]map[int64]map[string]*common.DevFaultInfo{}, nil
		}).ApplyFunc(getNeedToHandleFaults,
			func(faults map[int32]map[int64]map[string]*common.DevFaultInfo) []int32 {
				return []int32{}
			})
		defer patches.Reset()

		r.processResetWork()
		convey.So(r.countCache.GetAllFailedResetCountNpuId(), convey.ShouldBeEmpty)
	})
}

// TestResetMgr_ProcessResetWork_FilteredByCount tests scenario filtered by count limit
func TestResetMgr_ProcessResetWork_FilteredByCount(t *testing.T) {
	convey.Convey("When there are faults to reset but filtered out by count limit", t, func() {
		r := &ResetMgr{
			resetCache: domain.NewNpuInResetCache(),
			countCache: domain.NewFailedResetCountCache(),
		}
		patches := gomonkey.ApplyPrivateMethod(r, "allowToResetNpu", func() bool {
			return true
		}).ApplyFunc(getFaultCache, func() (map[int32]map[int64]map[string]*common.DevFaultInfo, error) {
			return map[int32]map[int64]map[string]*common.DevFaultInfo{
				1: {1001: {"module1": &common.DevFaultInfo{}}},
			}, nil
		}).ApplyFunc(getNeedToHandleFaults, func(faults map[int32]map[int64]map[string]*common.DevFaultInfo) []int32 {
			return []int32{1}
		}).ApplyPrivateMethod(r, "filterCountLimit", func(faultNpus []int32) []int32 {
			return []int32{}
		})
		defer patches.Reset()

		r.processResetWork()
		convey.So(r.countCache.GetAllFailedResetCountNpuId(), convey.ShouldBeEmpty)
	})
}

// TestResetMgr_ProcessResetWork_AllChecksPass tests scenario where all checks pass
func TestResetMgr_ProcessResetWork_AllChecksPass(t *testing.T) {
	convey.Convey("When there are faults to reset and all checks pass", t, func() {
		r := &ResetMgr{
			resetCache: domain.NewNpuInResetCache(),
			countCache: domain.NewFailedResetCountCache(),
		}
		patches := gomonkey.ApplyPrivateMethod(r, "allowToResetNpu", func() bool { return true }).
			ApplyFunc(getFaultCache, func() (map[int32]map[int64]map[string]*common.DevFaultInfo, error) {
				return map[int32]map[int64]map[string]*common.DevFaultInfo{
					1: {1001: {"module1": &common.DevFaultInfo{ReceiveTime: time.Now().Unix() - 70}}},
				}, nil
			}).
			ApplyFunc(getNeedToHandleFaults, func(faults map[int32]map[int64]map[string]*common.DevFaultInfo) []int32 {
				return []int32{1}
			}).
			ApplyPrivateMethod(r, "filterCountLimit", func(faultNpus []int32) []int32 { return faultNpus }).
			ApplyFunc(getAllRelatedNpus, func(faultNpus []int32) []domain.ResetNpuInfos {
				return []domain.ResetNpuInfos{{FaultId: 1, RelatedIds: []int32{1, 2}}}
			}).
			ApplyFunc(isNpuHoldByContainer, func(phyIds []int32) bool { return false }).
			ApplyFunc(isNpuHoldByProcess, func(phyIds []int32) (bool, error) { return false, nil }).
			ApplyFunc(isFaultExist, func(relatedIds []int32) bool { return true }).
			ApplyPrivateMethod(r, "hotReset", func(info domain.ResetNpuInfos) {})

		defer patches.Reset()

		r.processResetWork()
		convey.So(r.countCache.GetFailedResetCount(1), convey.ShouldEqual, 0)
	})
}

// TestResetMgr_ProcessResetWork_HeldByContainer tests scenario where npu is held by container
func TestResetMgr_ProcessResetWork_HeldByContainer(t *testing.T) {
	convey.Convey("When npu is held by container", t, func() {
		r := &ResetMgr{
			resetCache: domain.NewNpuInResetCache(),
			countCache: domain.NewFailedResetCountCache(),
		}
		patches := gomonkey.ApplyPrivateMethod(r, "allowToResetNpu", func() bool {
			return true
		}).
			ApplyFunc(getFaultCache, func() (map[int32]map[int64]map[string]*common.DevFaultInfo, error) {
				return map[int32]map[int64]map[string]*common.DevFaultInfo{
					1: {
						1001: {
							"module1": &common.DevFaultInfo{ReceiveTime: time.Now().Unix() - 70},
						},
					},
				}, nil
			}).
			ApplyFunc(getNeedToHandleFaults, func(faults map[int32]map[int64]map[string]*common.DevFaultInfo) []int32 {
				return []int32{1}
			}).
			ApplyPrivateMethod(r, "filterCountLimit", func(faultNpus []int32) []int32 {
				return faultNpus
			}).
			ApplyFunc(getAllRelatedNpus, func(faultNpus []int32) []domain.ResetNpuInfos {
				return []domain.ResetNpuInfos{
					{FaultId: 1, RelatedIds: []int32{1, 2}},
				}
			}).
			ApplyFunc(isNpuHoldByContainer, func(phyIds []int32) bool {
				return true
			})
		defer patches.Reset()

		r.processResetWork()
		convey.So(r.countCache.GetFailedResetCount(1), convey.ShouldEqual, 0)
	})
}

// TestResetMgr_ProcessResetWork_HeldByProcess tests scenario where npu is held by process
func TestResetMgr_ProcessResetWork_HeldByProcess(t *testing.T) {
	convey.Convey("When npu is held by process", t, func() {
		r := &ResetMgr{
			resetCache: domain.NewNpuInResetCache(),
			countCache: domain.NewFailedResetCountCache(),
		}
		patches := gomonkey.ApplyPrivateMethod(r, "allowToResetNpu", func() bool {
			return true
		}).
			ApplyFunc(getFaultCache, func() (map[int32]map[int64]map[string]*common.DevFaultInfo, error) {
				return map[int32]map[int64]map[string]*common.DevFaultInfo{
					1: {1001: {"module1": &common.DevFaultInfo{ReceiveTime: time.Now().Unix() - 70}}},
				}, nil
			}).
			ApplyFunc(getNeedToHandleFaults, func(faults map[int32]map[int64]map[string]*common.DevFaultInfo) []int32 {
				return []int32{1}
			}).
			ApplyPrivateMethod(r, "filterCountLimit", func(faultNpus []int32) []int32 {
				return faultNpus
			}).
			ApplyFunc(getAllRelatedNpus, func(faultNpus []int32) []domain.ResetNpuInfos {
				return []domain.ResetNpuInfos{{FaultId: 1, RelatedIds: []int32{1, 2}}}
			}).
			ApplyFunc(isNpuHoldByContainer, func(phyIds []int32) bool {
				return false
			}).
			ApplyFunc(isNpuHoldByProcess, func(phyIds []int32) (bool, error) {
				return true, nil
			})
		defer patches.Reset()

		r.processResetWork()
		convey.So(r.countCache.GetFailedResetCount(1), convey.ShouldEqual, 0)
	})
}

// TestResetMgr_ProcessResetWork_FaultNotExist tests scenario where fault no longer exists
func TestResetMgr_ProcessResetWork_FaultNotExist(t *testing.T) {
	convey.Convey("When fault no longer exists before reset", t, func() {
		r := &ResetMgr{
			resetCache: domain.NewNpuInResetCache(),
			countCache: domain.NewFailedResetCountCache(),
		}
		patches := gomonkey.ApplyPrivateMethod(r, "allowToResetNpu", func() bool {
			return true
		}).
			ApplyFunc(getFaultCache, func() (map[int32]map[int64]map[string]*common.DevFaultInfo, error) {
				return map[int32]map[int64]map[string]*common.DevFaultInfo{
					1: {1001: {"module1": &common.DevFaultInfo{ReceiveTime: time.Now().Unix() - 70}}},
				}, nil
			}).
			ApplyFunc(getNeedToHandleFaults, func(faults map[int32]map[int64]map[string]*common.DevFaultInfo) []int32 {
				return []int32{1}
			}).
			ApplyPrivateMethod(r, "filterCountLimit", func(faultNpus []int32) []int32 {
				return faultNpus
			}).
			ApplyFunc(getAllRelatedNpus, func(faultNpus []int32) []domain.ResetNpuInfos {
				return []domain.ResetNpuInfos{{FaultId: 1, RelatedIds: []int32{1, 2}}}
			}).
			ApplyFunc(isNpuHoldByContainer, func(phyIds []int32) bool {
				return false
			}).
			ApplyFunc(isNpuHoldByProcess, func(phyIds []int32) (bool, error) {
				return false, nil
			}).
			ApplyFunc(isFaultExist, func(relatedIds []int32) bool {
				return false
			})
		defer patches.Reset()

		r.processResetWork()
		convey.So(r.countCache.GetFailedResetCount(1), convey.ShouldEqual, 0)
	})
}

// TestResetMgr_HotReset_ExecDeviceResetFails tests hotReset when execDeviceReset fails
func TestResetMgr_HotReset_ExecDeviceResetFails(t *testing.T) {
	convey.Convey("When execDeviceReset fails", t, func() {
		r := &ResetMgr{
			resetCache: domain.NewNpuInResetCache(),
			countCache: domain.NewFailedResetCountCache(),
		}

		patches := gomonkey.ApplyFunc(execDeviceReset, func(faultPhyId int32) error {
			return errors.New("mock error")
		})
		defer patches.Reset()

		info := domain.ResetNpuInfos{FaultId: 1, RelatedIds: []int32{1, 2}}
		r.hotReset(info)
		convey.So(r.countCache.GetFailedResetCount(1), convey.ShouldEqual, 1)
		convey.So(r.lastSuccessResetTime, convey.ShouldBeNil)
	})
}

// TestResetMgr_HotReset_GetResetStatusFails tests hotReset when getResetSuccessfulStatus fails
func TestResetMgr_HotReset_GetResetStatusFails(t *testing.T) {
	convey.Convey("When getResetSuccessfulStatus fails", t, func() {
		r := &ResetMgr{
			resetCache: domain.NewNpuInResetCache(),
			countCache: domain.NewFailedResetCountCache(),
		}

		patches := gomonkey.ApplyFunc(execDeviceReset, func(faultPhyId int32) error {
			return nil
		}).
			ApplyFunc(getResetSuccessfulStatus, func(info domain.ResetNpuInfos) error {
				return errors.New("mock error")
			})
		defer patches.Reset()

		info := domain.ResetNpuInfos{FaultId: 1, RelatedIds: []int32{1, 2}}
		r.hotReset(info)
		convey.So(r.countCache.GetFailedResetCount(1), convey.ShouldEqual, 1)
		convey.So(r.lastSuccessResetTime, convey.ShouldBeNil)
	})
}

// TestResetMgr_HotReset_Success tests successful hotReset scenario
func TestResetMgr_HotReset_Success(t *testing.T) {
	convey.Convey("When hotReset success", t, func() {
		r := &ResetMgr{
			resetCache: domain.NewNpuInResetCache(),
			countCache: domain.NewFailedResetCountCache(),
		}

		patches := gomonkey.ApplyFunc(execDeviceReset, func(faultPhyId int32) error {
			return nil
		}).
			ApplyFunc(getResetSuccessfulStatus, func(info domain.ResetNpuInfos) error {
				return nil
			})
		defer patches.Reset()

		info := domain.ResetNpuInfos{FaultId: 1, RelatedIds: []int32{1, 2}}
		r.hotReset(info)
		convey.So(r.countCache.GetFailedResetCount(1), convey.ShouldEqual, 0)
		convey.So(r.lastSuccessResetTime, convey.ShouldNotBeNil)
	})
}

// TestExecDeviceReset_SuccessAfterRetry tests successful device reset after retry
func TestExecDeviceReset_SuccessAfterRetry(t *testing.T) {
	convey.Convey("When execDeviceReset success after retry", t, func() {
		var callCount int
		const testCallTimes = 3
		patches := gomonkey.ApplyMethodFunc(devmgr.DevMgr, "GetLogicIdByPhyId", mockGetLogicIdByPhyId).
			ApplyMethodReturn(devmgr.DevMgr, "GetDmgr", &devmanager.DeviceManager{}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "GetCardIDDeviceID", func(logicId int32) (int32, int32, error) {
				return 1, 1, nil
			}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "SetDeviceReset", func(cardID, deviceID int32) error {
				callCount++
				if callCount < testCallTimes {
					return errors.New("mock error")
				}
				return nil
			})
		defer patches.Reset()

		err := execDeviceReset(1)
		convey.So(err, convey.ShouldBeNil)
		convey.So(callCount, convey.ShouldEqual, testCallTimes)
	})
}

// TestExecDeviceReset_FailsAfterAllRetries tests device reset failure after all retries
func TestExecDeviceReset_FailsAfterAllRetries(t *testing.T) {
	convey.Convey("When execDeviceReset fails after all retries", t, func() {
		patches := gomonkey.ApplyMethodFunc(devmgr.DevMgr, "GetLogicIdByPhyId", mockGetLogicIdByPhyId).
			ApplyMethodReturn(devmgr.DevMgr, "GetDmgr", &devmanager.DeviceManager{}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "GetCardIDDeviceID", func(logicId int32) (int32, int32, error) {
				return 1, 1, nil
			}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "SetDeviceReset", func(cardID, deviceID int32) error {
				return errors.New("mock error")
			})
		defer patches.Reset()

		err := execDeviceReset(1)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestExecDeviceReset_GetCardIDDeviceIDFails tests failure when GetCardIDDeviceID fails
func TestExecDeviceReset_GetCardIDDeviceIDFails(t *testing.T) {
	convey.Convey("When GetCardIDDeviceID fails", t, func() {
		patches := gomonkey.ApplyMethodFunc(devmgr.DevMgr, "GetLogicIdByPhyId", mockGetLogicIdByPhyId).
			ApplyMethodReturn(devmgr.DevMgr, "GetDmgr", &devmanager.DeviceManager{}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "GetCardIDDeviceID", func(logicId int32) (int32, int32, error) {
				return 0, 0, errors.New("mock error")
			})
		defer patches.Reset()

		err := execDeviceReset(1)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestGetResetSuccessfulStatus_AllDevicesBootSuccess tests when all devices boot successfully
func TestGetResetSuccessfulStatus_AllDevicesBootSuccess(t *testing.T) {
	convey.Convey("When all devices boot successfully", t, func() {
		patches := gomonkey.ApplyMethodFunc(devmgr.DevMgr, "GetLogicIdByPhyId", mockGetLogicIdByPhyId).
			ApplyMethodReturn(devmgr.DevMgr, "GetDmgr", &devmanager.DeviceManager{}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "GetDeviceBootStatus", func(logicId int32) (int, error) {
				return devcommon.BootStartFinish, nil
			})
		defer patches.Reset()

		info := domain.ResetNpuInfos{RelatedIds: []int32{1, 2}}
		err := getResetSuccessfulStatus(info)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestGetResetSuccessfulStatus_DeviceBootWithRetry tests when device boot requires retry
func TestGetResetSuccessfulStatus_DeviceBootWithRetry(t *testing.T) {
	convey.Convey("When one device boot fails initially but succeeds after retry", t, func() {
		var callCount int
		const testCallTimes = 3
		patches := gomonkey.ApplyMethodFunc(devmgr.DevMgr, "GetLogicIdByPhyId", mockGetLogicIdByPhyId).
			ApplyMethodReturn(devmgr.DevMgr, "GetDmgr", &devmanager.DeviceManager{}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "GetDeviceBootStatus", func(logicId int32) (int, error) {
				callCount++
				if callCount < testCallTimes {
					return 0, nil
				}
				return devcommon.BootStartFinish, nil
			})
		defer patches.Reset()

		info := domain.ResetNpuInfos{RelatedIds: []int32{1}}
		err := getResetSuccessfulStatus(info)
		convey.So(err, convey.ShouldBeNil)
		convey.So(callCount, convey.ShouldEqual, testCallTimes)
	})
}

// TestGetResetSuccessfulStatus_BootStatusError tests when GetDeviceBootStatus returns error
func TestGetResetSuccessfulStatus_BootStatusError(t *testing.T) {
	convey.Convey("When GetDeviceBootStatus returns error", t, func() {
		patches := gomonkey.ApplyMethodFunc(devmgr.DevMgr, "GetLogicIdByPhyId", mockGetLogicIdByPhyId).
			ApplyMethodReturn(devmgr.DevMgr, "GetDmgr", &devmanager.DeviceManager{}).
			ApplyMethodFunc(devmgr.DevMgr.GetDmgr(), "GetDeviceBootStatus", func(logicId int32) (int, error) {
				return 0, errors.New("mock error")
			})
		defer patches.Reset()

		info := domain.ResetNpuInfos{RelatedIds: []int32{1}}
		err := getResetSuccessfulStatus(info)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestGetFaultCache tests the getFaultCache function
func TestGetFaultCache(t *testing.T) {
	convey.Convey("Test getFaultCache", t, func() {
		convey.Convey("When get fault cache successfully", func() {
			expectedFaults := map[int32]map[int64]map[string]*common.DevFaultInfo{
				1: {
					1001: {
						"module1": &common.DevFaultInfo{ReceiveTime: time.Now().Unix()},
					},
				},
			}

			// Create mock fault cache
			mockFaultCache := &faultdomain.FaultCache{}
			patches := gomonkey.ApplyFunc(faultdomain.GetFaultCache, func() *faultdomain.FaultCache {
				return mockFaultCache
			}).
				ApplyMethodFunc(mockFaultCache, "DeepCopy", func() (map[int32]map[int64]map[string]*common.DevFaultInfo, error) {
					return expectedFaults, nil
				})
			defer patches.Reset()

			result, err := getFaultCache()
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, expectedFaults)
		})

		convey.Convey("When fault cache is empty", func() {
			mockFaultCache := &faultdomain.FaultCache{}
			patches := gomonkey.ApplyFunc(faultdomain.GetFaultCache, func() *faultdomain.FaultCache {
				return mockFaultCache
			}).
				ApplyMethodFunc(mockFaultCache, "DeepCopy", func() (map[int32]map[int64]map[string]*common.DevFaultInfo, error) {
					return map[int32]map[int64]map[string]*common.DevFaultInfo{}, nil
				})
			defer patches.Reset()

			result, err := getFaultCache()
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldBeEmpty)
		})
	})
}

// TestGetNeedToHandleFaults_AllNeedHandling tests when all faults need handling
func TestGetNeedToHandleFaults_AllNeedHandling(t *testing.T) {
	convey.Convey("When there are faults need to handle", t, func() {
		patches := gomonkey.ApplyFunc(isFaultsNeedToHandle, func(faultInfoMap map[int64]map[string]*common.DevFaultInfo) bool {
			return true
		})
		defer patches.Reset()

		faults := map[int32]map[int64]map[string]*common.DevFaultInfo{
			1: {1001: {"module1": &common.DevFaultInfo{}}},
			2: {1002: {"module2": &common.DevFaultInfo{}}},
		}

		result := getNeedToHandleFaults(faults)
		convey.So(len(result), convey.ShouldResemble, len(faults))
	})
}

// TestGetNeedToHandleFaults_NoneNeedHandling tests when no faults need handling
func TestGetNeedToHandleFaults_NoneNeedHandling(t *testing.T) {
	convey.Convey("When no faults need to handle", t, func() {
		patches := gomonkey.ApplyFunc(isFaultsNeedToHandle, func(faultInfoMap map[int64]map[string]*common.DevFaultInfo) bool {
			return false
		})
		defer patches.Reset()

		faults := map[int32]map[int64]map[string]*common.DevFaultInfo{
			1: {1001: {"module1": &common.DevFaultInfo{}}},
		}

		result := getNeedToHandleFaults(faults)
		convey.So(result, convey.ShouldBeEmpty)
	})
}

// TestGetNeedToHandleFaults_MixedNeedHandling tests when some faults need handling and some don't
func TestGetNeedToHandleFaults_MixedNeedHandling(t *testing.T) {
	convey.Convey("When mixed faults - some need handling, some don't", t, func() {
		callCount := 0
		patches := gomonkey.ApplyFunc(isFaultsNeedToHandle, func(faultInfoMap map[int64]map[string]*common.DevFaultInfo) bool {
			callCount++
			// NPU 1 needs handling, NPU 2 doesn't
			return callCount == 1
		})
		defer patches.Reset()

		faults := map[int32]map[int64]map[string]*common.DevFaultInfo{
			1: {1001: {"module1": &common.DevFaultInfo{}}},
			2: {1002: {"module2": &common.DevFaultInfo{}}},
		}

		result := getNeedToHandleFaults(faults)
		convey.So(len(result), convey.ShouldResemble, 1)
	})
}

// TestIsFaultsNeedToHandle_L4L5Fault tests L4/L5 fault scenario
func TestIsFaultsNeedToHandle_L4L5Fault(t *testing.T) {
	convey.Convey("When there is L4/L5 fault that needs immediate handling", t, func() {
		patches := gomonkey.ApplyFunc(faultdomain.GetFaultLevelByCode, func(codes []int64) string {
			return common.RestartNPU // L4/L5 level fault
		})
		defer patches.Reset()

		faultInfoMap := map[int64]map[string]*common.DevFaultInfo{
			1001: {"module1": &common.DevFaultInfo{}},
		}

		result := isFaultsNeedToHandle(faultInfoMap)
		convey.So(result, convey.ShouldBeTrue)
	})
}

// TestIsFaultsNeedToHandle_L2L3FaultLasting tests L2/L3 fault with lasting condition
func TestIsFaultsNeedToHandle_L2L3FaultLasting(t *testing.T) {
	convey.Convey("When there is L2/L3 fault that needs lasting check", t, func() {
		patches := gomonkey.ApplyFunc(faultdomain.GetFaultLevelByCode, func(codes []int64) string {
			return common.RestartRequest // L2/L3 level fault
		}).
			ApplyFunc(checkLastingFaultNeedToReset, func(faultInfo map[string]*common.DevFaultInfo) bool {
				return true // Lasting more than 60 seconds
			})
		defer patches.Reset()

		faultInfoMap := map[int64]map[string]*common.DevFaultInfo{
			1001: {"module1": &common.DevFaultInfo{}},
		}

		result := isFaultsNeedToHandle(faultInfoMap)
		convey.So(result, convey.ShouldBeTrue)
	})
}

// TestIsFaultsNeedToHandle_L2L3FaultNotLasting tests L2/L3 fault without lasting condition
func TestIsFaultsNeedToHandle_L2L3FaultNotLasting(t *testing.T) {
	convey.Convey("When L2/L3 fault but not lasting enough", t, func() {
		patches := gomonkey.ApplyFunc(faultdomain.GetFaultLevelByCode, func(codes []int64) string {
			return common.RestartRequest // L2/L3 level fault
		}).
			ApplyFunc(checkLastingFaultNeedToReset, func(faultInfo map[string]*common.DevFaultInfo) bool {
				return false // Not lasting more than 60 seconds
			})
		defer patches.Reset()

		faultInfoMap := map[int64]map[string]*common.DevFaultInfo{
			1001: {"module1": &common.DevFaultInfo{}},
		}

		result := isFaultsNeedToHandle(faultInfoMap)
		convey.So(result, convey.ShouldBeFalse)
	})
}

// TestIsFaultsNeedToHandle_NoNeedHandling tests non-handling fault levels
func TestIsFaultsNeedToHandle_NoNeedHandling(t *testing.T) {
	convey.Convey("When no faults need handling", t, func() {
		patches := gomonkey.ApplyFunc(faultdomain.GetFaultLevelByCode, func(codes []int64) string {
			return "L1" // Level that doesn't need handling
		})
		defer patches.Reset()

		faultInfoMap := map[int64]map[string]*common.DevFaultInfo{
			1001: {"module1": &common.DevFaultInfo{}},
		}

		result := isFaultsNeedToHandle(faultInfoMap)
		convey.So(result, convey.ShouldBeFalse)
	})
}

// TestIsFaultsNeedToHandle_MultipleFaults tests multiple fault codes with short-circuit logic
func TestIsFaultsNeedToHandle_MultipleFaults(t *testing.T) {
	convey.Convey("When multiple fault codes - first one needs handling", t, func() {
		callCount := 0
		patches := gomonkey.ApplyFunc(faultdomain.GetFaultLevelByCode, func(codes []int64) string {
			callCount++
			if callCount == 1 {
				return common.RestartNPU // First one needs handling
			}
			return "L1" // Others don't need handling
		})
		defer patches.Reset()

		faultInfoMap := map[int64]map[string]*common.DevFaultInfo{
			1001: {"module1": &common.DevFaultInfo{}},
			1002: {"module2": &common.DevFaultInfo{}},
		}

		result := isFaultsNeedToHandle(faultInfoMap)
		convey.So(result, convey.ShouldBeTrue)
		convey.So(callCount, convey.ShouldEqual, 1) // Should short-circuit return, only called once
	})
}

// TestCheckLastingFaultNeedToReset tests the checkLastingFaultNeedToReset function
func TestCheckLastingFaultNeedToReset(t *testing.T) {
	convey.Convey("Test checkLastingFaultNeedToReset", t, func() {
		convey.Convey("When fault lasting more than 60 seconds", func() {
			faultInfo := map[string]*common.DevFaultInfo{
				"module1": {ReceiveTime: time.Now().Unix() - 70}, // 70 seconds ago
			}
			result := checkLastingFaultNeedToReset(faultInfo)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("When fault lasting less than 60 seconds", func() {
			faultInfo := map[string]*common.DevFaultInfo{
				"module1": {ReceiveTime: time.Now().Unix() - 30}, // 30 seconds ago
			}
			result := checkLastingFaultNeedToReset(faultInfo)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When multiple faults - one lasting more than 60 seconds", func() {
			faultInfo := map[string]*common.DevFaultInfo{
				"module1": {ReceiveTime: time.Now().Unix() - 30},  // 30 seconds ago
				"module2": {ReceiveTime: time.Now().Unix() - 70},  // 70 seconds ago
				"module3": {ReceiveTime: time.Now().Unix() - 100}, // 100 seconds ago
			}
			result := checkLastingFaultNeedToReset(faultInfo)
			convey.So(result, convey.ShouldBeTrue) // Returns true if any exceeds 60 seconds
		})

		convey.Convey("When multiple faults - all lasting less than 60 seconds", func() {
			faultInfo := map[string]*common.DevFaultInfo{
				"module1": {ReceiveTime: time.Now().Unix() - 30}, // 30 seconds ago
				"module2": {ReceiveTime: time.Now().Unix() - 50}, // 50 seconds ago
				"module3": {ReceiveTime: time.Now().Unix() - 59}, // 59 seconds ago
			}
			result := checkLastingFaultNeedToReset(faultInfo)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When fault info is empty", func() {
			faultInfo := map[string]*common.DevFaultInfo{}
			result := checkLastingFaultNeedToReset(faultInfo)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}
