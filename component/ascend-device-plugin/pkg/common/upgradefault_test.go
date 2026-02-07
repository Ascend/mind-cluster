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

// Package common a series of common function
package common

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		fmt.Printf("log init failed, error is %v\n", err)
	}
}

const (
	oneSecond                           = 1000
	testLogicID1        LogicId         = 1001
	testLogicID2        LogicId         = 1002
	testLogicID3        LogicId         = 1003
	testFaultTime       int64           = 1700000000000
	testFaultCode                       = "E1001"
	testFaultLevel                      = "critical"
	testFaultLevel2                     = "warning"
	manuallySeparateNPU                 = "ManuallySeparateNPU"
	testUpgradeType     UpgradeTypeEnum = "FaultDuration"
	testPhyID1          PhyId           = 2001
	testPhyID2          PhyId           = 2002
	testDevicePrefix                    = "npu"
	testCMJSON                          = `{"npu-2001":[{"upgrade_time":1700000000000,"fault_code":"E1001","fault_level":"critical","upgrade_type":"FaultDuration"}]}`
	invalidCMJSON                       = `{"npu-2001":[{invalid json}]}`
	invalidDeviceName                   = "npu"
)

func TestSaveUpgradeFaultCache(t *testing.T) {
	convey.Convey("Test SaveUpgradeFaultCache", t, func() {
		convey.Convey("should save cache when given valid cache map", func() {
			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode:   testFaultCode,
				FaultLevel:  testFaultLevel,
				UpgradeType: testUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime:           testFaultTime,
				UpgradeFaultReasonKey: faultReasonKey,
			}
			testCache := UpgradeFaultReasonMap[LogicId]{

				testLogicID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}
			originalCache := upgradeFaultCacheMgr.cache
			defer func() { upgradeFaultCacheMgr.cache = originalCache }()

			SaveUpgradeFaultCache(testCache)

			convey.So(upgradeFaultCacheMgr.cache, convey.ShouldResemble, testCache)
		})

		convey.Convey("should handle empty cache when given empty map", func() {
			testCache := UpgradeFaultReasonMap[LogicId]{}
			originalCache := upgradeFaultCacheMgr.cache
			defer func() { upgradeFaultCacheMgr.cache = originalCache }()

			SaveUpgradeFaultCache(testCache)

			convey.So(len(upgradeFaultCacheMgr.cache), convey.ShouldEqual, 0)
		})
	})
}

func TestInsertUpgradeFaultCache(t *testing.T) {
	convey.Convey("Test InsertUpgradeFaultCache", t, func() {
		convey.Convey("should insert new reason when reason not exist", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			originalCache := upgradeFaultCacheMgr.cache
			defer func() { upgradeFaultCacheMgr.cache = originalCache }()
			upgradeFaultCacheMgr.cache = make(UpgradeFaultReasonMap[LogicId])

			InsertUpgradeFaultCache(testLogicID1, testFaultTime, testFaultCode, testFaultLevel, testUpgradeType)

			convey.So(len(upgradeFaultCacheMgr.cache), convey.ShouldEqual, 1)
		})

		convey.Convey("should update reason when newer fault time", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			originalCache := upgradeFaultCacheMgr.cache
			defer func() { upgradeFaultCacheMgr.cache = originalCache }()
			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode:   testFaultCode,
				FaultLevel:  testFaultLevel,
				UpgradeType: testUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime:           testFaultTime - oneSecond,
				UpgradeFaultReasonKey: faultReasonKey,
			}
			upgradeFaultCacheMgr.cache = UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}

			InsertUpgradeFaultCache(testLogicID1, testFaultTime, testFaultCode, testFaultLevel, testUpgradeType)

			reason := upgradeFaultCacheMgr.cache[testLogicID1][UpgradeFaultReasonKey{
				FaultCode:   testFaultCode,
				FaultLevel:  testFaultLevel,
				UpgradeType: testUpgradeType}]
			convey.So(reason.UpgradeTime, convey.ShouldEqual, testFaultTime)
		})
	})
}

func TestRemoveManuallySeparateReasonCache(t *testing.T) {
	convey.Convey("Test RemoveManuallySeparateReasonCache", t, func() {
		convey.Convey("should remove manually separate reason when logic ids contain target level", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			originalCache := upgradeFaultCacheMgr.cache
			originalRemoved := upgradeFaultCacheMgr.removedEvent
			defer func() {
				upgradeFaultCacheMgr.cache = originalCache
				upgradeFaultCacheMgr.removedEvent = originalRemoved
			}()
			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode: "", FaultLevel: manuallySeparateNPU, UpgradeType: AutofillUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime: testFaultTime, UpgradeFaultReasonKey: faultReasonKey,
			}
			upgradeFaultCacheMgr.cache = UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}
			RemoveManuallySeparateReasonCache([]LogicId{testLogicID1})
			convey.So(len(upgradeFaultCacheMgr.cache), convey.ShouldEqual, 0)
			convey.So(len(upgradeFaultCacheMgr.removedEvent), convey.ShouldBeGreaterThan, 0)
		})
		convey.Convey("should not remove when logic ids not contain target level", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			originalCache := upgradeFaultCacheMgr.cache
			defer func() { upgradeFaultCacheMgr.cache = originalCache }()
			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode: testFaultCode, FaultLevel: testFaultLevel, UpgradeType: testUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime: testFaultTime, UpgradeFaultReasonKey: faultReasonKey,
			}
			upgradeFaultCacheMgr.cache = UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}
			RemoveManuallySeparateReasonCache([]LogicId{testLogicID1})
			convey.So(len(upgradeFaultCacheMgr.cache), convey.ShouldEqual, 1)
		})
	})
}

func TestRemoveTimeoutReasonCache(t *testing.T) {
	convey.Convey("Test RemoveTimeoutReasonCache", t, func() {
		convey.Convey("should remove timeout reason when fault code match", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			originalCache := upgradeFaultCacheMgr.cache
			originalRemoved := upgradeFaultCacheMgr.removedEvent
			defer func() {
				upgradeFaultCacheMgr.cache = originalCache
				upgradeFaultCacheMgr.removedEvent = originalRemoved
			}()

			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode: testFaultCode, FaultLevel: testFaultLevel, UpgradeType: testUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime: testFaultTime, UpgradeFaultReasonKey: faultReasonKey,
			}
			upgradeFaultCacheMgr.cache = UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}
			RemoveTimeoutReasonCache(testLogicID1, testFaultCode)
			convey.So(len(upgradeFaultCacheMgr.cache), convey.ShouldEqual, 0)
			convey.So(len(upgradeFaultCacheMgr.removedEvent), convey.ShouldBeGreaterThan, 0)
		})

		convey.Convey("should not remove when fault code not match", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			originalCache := upgradeFaultCacheMgr.cache
			defer func() { upgradeFaultCacheMgr.cache = originalCache }()
			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode: testFaultCode, FaultLevel: testFaultLevel, UpgradeType: testUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime: testFaultTime, UpgradeFaultReasonKey: faultReasonKey,
			}
			upgradeFaultCacheMgr.cache = UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}
			RemoveTimeoutReasonCache(testLogicID1, "E9999")
			convey.So(len(upgradeFaultCacheMgr.cache), convey.ShouldEqual, 1)
		})
	})
}

func TestGetAndCleanRemovedReasonEvent(t *testing.T) {
	convey.Convey("Test GetAndCleanRemovedReasonEvent", t, func() {
		convey.Convey("should get and clean removed events when events exist", func() {
			originalRemoved := upgradeFaultCacheMgr.removedEvent
			defer func() { upgradeFaultCacheMgr.removedEvent = originalRemoved }()

			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode:   testFaultCode,
				FaultLevel:  testFaultLevel,
				UpgradeType: testUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime:           testFaultTime,
				UpgradeFaultReasonKey: faultReasonKey,
			}
			expectedEvents := UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}
			upgradeFaultCacheMgr.removedEvent = expectedEvents

			result := GetAndCleanRemovedReasonEvent()

			convey.So(result, convey.ShouldResemble, expectedEvents)
			convey.So(len(upgradeFaultCacheMgr.removedEvent), convey.ShouldEqual, 0)
		})
	})
}

func TestCopyUpgradeFaultCache(t *testing.T) {
	convey.Convey("Test CopyUpgradeFaultCache", t, func() {
		convey.Convey("should return copy of cache when cache has data", func() {
			originalCache := upgradeFaultCacheMgr.cache
			defer func() { upgradeFaultCacheMgr.cache = originalCache }()

			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode:   testFaultCode,
				FaultLevel:  testFaultLevel,
				UpgradeType: testUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime:           testFaultTime,
				UpgradeFaultReasonKey: faultReasonKey,
			}
			expectedCache := UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}
			upgradeFaultCacheMgr.cache = expectedCache

			result := CopyUpgradeFaultCache()

			convey.So(result, convey.ShouldResemble, expectedCache)
			convey.So(&result, convey.ShouldNotEqual, &expectedCache)
		})
	})
}

func TestEquals(t *testing.T) {
	convey.Convey("Test UpgradeFaultReasonMap.Equals", t, func() {
		convey.Convey("should return true when two maps are equal", func() {
			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode:   testFaultCode,
				FaultLevel:  testFaultLevel,
				UpgradeType: testUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime:           testFaultTime,
				UpgradeFaultReasonKey: faultReasonKey,
			}
			map1 := UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}
			map2 := UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}
			result := map1.Equals(map2)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("should return false when two maps have different reason sets", func() {
			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode:   testFaultCode,
				FaultLevel:  testFaultLevel,
				UpgradeType: testUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime:           testFaultTime,
				UpgradeFaultReasonKey: faultReasonKey,
			}
			map1 := UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}
			faultReasonKey.FaultLevel = testFaultLevel2
			upgradeFaultReason.FaultLevel = testFaultLevel2
			map2 := UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}
			result := map1.Equals(map2)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestGetKeys(t *testing.T) {
	convey.Convey("Test UpgradeFaultReasonMap.GetKeys", t, func() {
		convey.Convey("should return all keys when map has multiple entries", func() {
			reasonMap := UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{},
				testLogicID2: UpgradeFaultReasonSet{},
				testLogicID3: UpgradeFaultReasonSet{},
			}

			keys := reasonMap.GetKeys()

			convey.So(len(keys), convey.ShouldEqual, len(reasonMap))
		})

		convey.Convey("should return empty slice when map is empty", func() {
			reasonMap := UpgradeFaultReasonMap[LogicId]{}

			keys := reasonMap.GetKeys()

			convey.So(len(keys), convey.ShouldEqual, 0)
		})
	})
}

func TestConvertCacheToCm(t *testing.T) {
	convey.Convey("Test UpgradeFaultReasonMap.ConvertCacheToCm", t, func() {
		convey.Convey("should convert successfully when all logic ids can be converted", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyFuncReturn(logicToPhyConvertFunc, int32(testPhyID1), nil)

			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode:   testFaultCode,
				FaultLevel:  testFaultLevel,
				UpgradeType: testUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime:           testFaultTime,
				UpgradeFaultReasonKey: faultReasonKey,
			}
			reasonMap := UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}

			result, err := reasonMap.ConvertCacheToCm(logicToPhyConvertFunc)

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 1)
			convey.So(result[testPhyID1], convey.ShouldNotBeNil)
		})

		convey.Convey("should return error when conversion function fails", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			expectedError := fmt.Errorf("conversion failed")
			patches.ApplyFuncReturn(logicToPhyConvertFunc, int32(0), expectedError)

			reasonMap := UpgradeFaultReasonMap[LogicId]{
				testLogicID1: UpgradeFaultReasonSet{},
			}

			result, err := reasonMap.ConvertCacheToCm(logicToPhyConvertFunc)

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "convert logicId")
		})
	})
}

func TestConvertCmToCache(t *testing.T) {
	convey.Convey("Test UpgradeFaultReasonMap.ConvertCmToCache", t, func() {
		convey.Convey("should convert successfully when all phy ids can be converted", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyFuncReturn(phyToLogicConvertFunc, int32(testLogicID1), nil)

			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode:   testFaultCode,
				FaultLevel:  testFaultLevel,
				UpgradeType: testUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime:           testFaultTime,
				UpgradeFaultReasonKey: faultReasonKey,
			}
			reasonMap := UpgradeFaultReasonMap[PhyId]{
				testPhyID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}

			result, err := reasonMap.ConvertCmToCache(phyToLogicConvertFunc)

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 1)
			convey.So(result[testLogicID1], convey.ShouldNotBeNil)
		})

		convey.Convey("should return error when conversion function fails", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			expectedError := fmt.Errorf("conversion failed")
			patches.ApplyFuncReturn(phyToLogicConvertFunc, int32(0), expectedError)

			reasonMap := UpgradeFaultReasonMap[PhyId]{
				testPhyID1: UpgradeFaultReasonSet{},
			}

			result, err := reasonMap.ConvertCmToCache(phyToLogicConvertFunc)

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "convert phyId")
		})
	})
}

func TestCmToString(t *testing.T) {
	convey.Convey("Test UpgradeFaultReasonMap.CmToString", t, func() {
		convey.Convey("should return correct JSON string when map has data", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyFuncReturn(ObjToString, testCMJSON)

			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode:   testFaultCode,
				FaultLevel:  testFaultLevel,
				UpgradeType: testUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime:           testFaultTime,
				UpgradeFaultReasonKey: faultReasonKey,
			}
			reasonMap := UpgradeFaultReasonMap[PhyId]{
				testPhyID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}

			result := reasonMap.CmToString(testDevicePrefix)

			convey.So(result, convey.ShouldEqual, testCMJSON)
		})

		convey.Convey("should handle empty map correctly", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyFuncReturn(ObjToString, "{}")

			reasonMap := UpgradeFaultReasonMap[PhyId]{}

			result := reasonMap.CmToString(testDevicePrefix)

			convey.So(result, convey.ShouldEqual, "{}")
		})
	})
}

func TestStringToReasonCm(t *testing.T) {
	convey.Convey("Test StringToReasonCm", t, func() {
		testStringToReasonCmValid(t)

		convey.Convey("should return error when JSON is invalid", func() {
			result, err := StringToReasonCm(invalidCMJSON)

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("should return error when device name format is invalid", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyFunc(json.Unmarshal, func(data []byte, v interface{}) error {
				cmData, ok := v.(*map[string][]UpgradeFaultReason)
				if !ok {
					return nil
				}
				*cmData = map[string][]UpgradeFaultReason{
					invalidDeviceName: {},
				}
				return nil
			})
			result, err := StringToReasonCm(`{"npu":[]}`)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "get phyid from")
		})
	})
}

func testStringToReasonCmValid(t *testing.T) {
	convey.Convey("should parse successfully when JSON is valid", func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		faultReasonKey := UpgradeFaultReasonKey{
			FaultCode:   testFaultCode,
			FaultLevel:  testFaultLevel,
			UpgradeType: testUpgradeType,
		}
		upgradeFaultReason := UpgradeFaultReason{
			UpgradeTime:           testFaultTime,
			UpgradeFaultReasonKey: faultReasonKey,
		}
		patches.ApplyFunc(json.Unmarshal, func(data []byte, v interface{}) error {
			cmData, ok := v.(*map[string][]UpgradeFaultReason)
			if !ok {
				return nil
			}
			*cmData = map[string][]UpgradeFaultReason{
				"npu-2001": {upgradeFaultReason},
			}
			return nil
		})

		result, err := StringToReasonCm(testCMJSON)

		convey.So(err, convey.ShouldBeNil)
		convey.So(len(result), convey.ShouldEqual, 1)
		convey.So(result[testPhyID1], convey.ShouldNotBeNil)
	})
}

func TestFixManuallySeparateReason(t *testing.T) {
	convey.Convey("Test UpgradeFaultReasonMap.FixManuallySeparateReason", t, func() {
		convey.Convey("should add missing manually separate reason when NPU in list but not in map", func() {
			reasonMap := make(UpgradeFaultReasonMap[PhyId])
			manuallySeparateList := []PhyId{testPhyID1, testPhyID2}

			reasonMap.FixManuallySeparateReason(manuallySeparateList)

			convey.So(len(reasonMap), convey.ShouldEqual, len(manuallySeparateList))
			convey.So(reasonMap[testPhyID1], convey.ShouldNotBeNil)
			convey.So(reasonMap[testPhyID2], convey.ShouldNotBeNil)
		})

		convey.Convey("should remove redundant manually separate reason when NPU not in list but in map", func() {
			faultReasonKey := UpgradeFaultReasonKey{
				FaultCode:   "",
				FaultLevel:  manuallySeparateNPU,
				UpgradeType: AutofillUpgradeType,
			}
			upgradeFaultReason := UpgradeFaultReason{
				UpgradeTime:           time.Now().UnixMilli(),
				UpgradeFaultReasonKey: faultReasonKey,
			}
			reasonMap := UpgradeFaultReasonMap[PhyId]{
				testPhyID1: UpgradeFaultReasonSet{
					faultReasonKey: upgradeFaultReason,
				},
			}
			var manuallySeparateList []PhyId

			reasonMap.FixManuallySeparateReason(manuallySeparateList)

			convey.So(len(reasonMap), convey.ShouldEqual, 0)
		})

		testShouldKeepInReason()
	})
}

func testShouldKeepInReason() {
	convey.Convey("should keep existing manually separate reason when NPU in list and map", func() {
		originalTime := time.Now().UnixMilli() - oneSecond
		faultReasonKey := UpgradeFaultReasonKey{
			FaultCode:   "",
			FaultLevel:  manuallySeparateNPU,
			UpgradeType: AutofillUpgradeType,
		}
		upgradeFaultReason := UpgradeFaultReason{
			UpgradeTime:           originalTime,
			UpgradeFaultReasonKey: faultReasonKey,
		}
		reasonMap := UpgradeFaultReasonMap[PhyId]{
			testPhyID1: UpgradeFaultReasonSet{
				faultReasonKey: upgradeFaultReason,
			},
		}
		manuallySeparateList := []PhyId{testPhyID1}

		reasonMap.FixManuallySeparateReason(manuallySeparateList)

		reasonSet := reasonMap[testPhyID1]
		convey.So(len(reasonSet), convey.ShouldEqual, 1)
		for _, reason := range reasonSet {
			convey.So(reason.UpgradeTime, convey.ShouldEqual, originalTime)
		}
	})
}

func logicToPhyConvertFunc(logicID int32) (int32, error) {
	return int32(testPhyID1), nil
}

func phyToLogicConvertFunc(phyID int32) (int32, error) {
	return int32(testLogicID1), nil
}
