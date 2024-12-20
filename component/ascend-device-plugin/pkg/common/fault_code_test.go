/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/utils"
	"ascend-common/devmanager/common"
)

// TestLoadFaultCodeFromFile for test LoadFaultCodeFromFile
func TestLoadFaultCodeFromFile(t *testing.T) {
	convey.Convey("test LoadFaultCodeFromFile", t, func() {
		convey.Convey("utils.LoadFile err", func() {
			mockLoadFile := gomonkey.ApplyFuncReturn(utils.LoadFile, nil, errors.New("failed"))
			defer mockLoadFile.Reset()
			convey.So(LoadFaultCodeFromFile(), convey.ShouldNotBeNil)
		})
		convey.Convey("test LoadFaultCodeFromFile", func() {
			mockLoadFile := gomonkey.ApplyFuncReturn(utils.LoadFile, nil, nil)
			defer mockLoadFile.Reset()
			mockUnmarshal := gomonkey.ApplyFuncReturn(json.Unmarshal, errors.New("failed"))
			defer mockUnmarshal.Reset()
			convey.So(LoadFaultCodeFromFile(), convey.ShouldNotBeNil)
		})
	})
}

// TestGetFaultTypeByCode for test GetFaultTypeByCode
func TestGetFaultTypeByCode(t *testing.T) {
	convey.Convey("test GetFaultTypeByCode", t, func() {
		faultCodes := []int64{1}
		convey.Convey("fault type NormalNPU", func() {
			convey.So(GetFaultTypeByCode(nil), convey.ShouldEqual, NormalNPU)
		})
		convey.Convey("fault type NotHandleFault", func() {
			faultTypeCode = FaultTypeCode{NotHandleFaultCodes: faultCodes}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, NotHandleFault)
		})
		convey.Convey("fault type SeparateNPU", func() {
			faultTypeCode = FaultTypeCode{SeparateNPUCodes: faultCodes}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, SeparateNPU)
			faultTypeCode = FaultTypeCode{}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, SeparateNPU)
		})
		convey.Convey("fault type RestartNPU", func() {
			faultTypeCode = FaultTypeCode{RestartNPUCodes: faultCodes}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, RestartNPU)
		})
		convey.Convey("fault type FreeRestartNPU", func() {
			faultTypeCode = FaultTypeCode{FreeRestartNPUCodes: faultCodes}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, FreeRestartNPU)
		})
		convey.Convey("fault type RestartBusiness", func() {
			faultTypeCode = FaultTypeCode{RestartBusinessCodes: faultCodes}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, RestartBusiness)
		})
		convey.Convey("fault type RestartRequestCodes", func() {
			faultTypeCode = FaultTypeCode{RestartRequestCodes: faultCodes}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, RestartRequest)
		})
	})
}

// TestSetDeviceInit for test SetDeviceInit
func TestSetDeviceInit(t *testing.T) {
	initId := int32(0)
	initLogicIDsLen := 1
	convey.Convey("test SetDeviceInit", t, func() {
		convey.Convey("SetDeviceInit success", func() {
			initLogicIDs = nil
			SetDeviceInit(initId)
			convey.So(len(initLogicIDs), convey.ShouldEqual, initLogicIDsLen)
		})
	})
}

// TestGetAndCleanLogicID for test GetAndCleanLogicID
func TestGetAndCleanLogicID(t *testing.T) {
	convey.Convey("test GetAndCleanLogicID", t, func() {
		convey.Convey("initLogicIDs is empty", func() {
			initLogicIDs = nil
			convey.So(GetAndCleanLogicID(), convey.ShouldBeNil)
		})
		convey.Convey("initLogicIDs is not empty", func() {
			testIDs := []int32{1}
			initLogicIDs = testIDs
			convey.So(GetAndCleanLogicID(), convey.ShouldResemble, testIDs)
		})
	})
}

// TestSetNewFaultAndCacheOnceRecoverFault for test SetNewFaultAndCacheOnceRecoverFault
func TestSetNewFaultAndCacheOnceRecoverFault(t *testing.T) {
	convey.Convey("test SetNewFaultAndCacheOnceRecoverFault", t, func() {
		convey.Convey("SetNewFaultAndCacheOnceRecoverFault success", func() {
			recoverFaultMap = make(map[int32][]int64, GeneralMapSize)
			logicID := int32(0)
			faultInfos := []common.DevFaultInfo{
				{Assertion: common.FaultRecover},
				{Assertion: common.FaultRecover, EventID: 1},
				{Assertion: common.FaultOnce, EventID: 0},
				{Assertion: common.FaultOccur, EventID: LinkDownFaultCode},
				{Assertion: common.FaultRecover, EventID: LinkDownFaultCode},
			}
			device := &NpuDevice{FaultCodes: []int64{1}}
			expectedFaultCodes, expectedFaultMapLen := []int64{0}, 2
			NetworkFaultCodes = sets.NewInt64()
			NetworkFaultCodes.Insert(LinkDownFaultCode)
			SetNewFaultAndCacheOnceRecoverFault(logicID, faultInfos, device)
			convey.So(device.FaultCodes, convey.ShouldResemble, expectedFaultCodes)
			convey.So(len(recoverFaultMap[logicID]), convey.ShouldEqual, expectedFaultMapLen)
		})
	})
}

// TestSetNetworkNewFaultAndCacheOnceRecoverFault for test SetNetworkNewFaultAndCacheOnceRecoverFault
func TestSetNetworkNewFaultAndCacheOnceRecoverFault(t *testing.T) {
	convey.Convey("test SetNetworkNewFaultAndCacheOnceRecoverFault", t, func() {
		convey.Convey("SetNetworkNewFaultAndCacheOnceRecoverFault success", func() {
			recoverNetworkFaultMap = make(map[int32][]int64, GeneralMapSize)
			logicID := int32(0)
			eventId0 := int64(0)
			eventId1 := int64(1)
			faultInfos := []common.DevFaultInfo{
				{Assertion: common.FaultRecover},
				{Assertion: common.FaultRecover, EventID: eventId1},
				{Assertion: common.FaultOnce, EventID: eventId0},
				{Assertion: common.FaultOccur, EventID: LinkDownFaultCode},
				{Assertion: common.FaultRecover, EventID: LinkDownFaultCode},
				{Assertion: common.FaultOnce, EventID: LinkDownFaultCode},
			}
			device := &NpuDevice{NetworkFaultCodes: []int64{LinkDownFaultCode}}
			expectedNetworkFaultCodes := []int64{LinkDownFaultCode, LinkDownFaultCode}
			expectedRecoverNetworkFaultMapLen := 1
			NetworkFaultCodes = sets.NewInt64()
			NetworkFaultCodes.Insert(LinkDownFaultCode)
			SetNetworkNewFaultAndCacheOnceRecoverFault(logicID, faultInfos, device)
			convey.So(device.NetworkFaultCodes, convey.ShouldResemble, expectedNetworkFaultCodes)
			convey.So(len(recoverNetworkFaultMap[logicID]), convey.ShouldEqual, expectedRecoverNetworkFaultMapLen)
		})
	})
}

// TestDelOnceRecoverFault for test DelOnceRecoverFault
func TestDelOnceRecoverFault(t *testing.T) {
	convey.Convey("test DelOnceRecoverFault", t, func() {
		convey.Convey("DelOnceRecoverFault success", func() {
			faultCodes := []int64{1}
			networkFaultCodes := []int64{LinkDownFaultCode}
			logicId := int32(0)
			expectedNum := 0
			device := &NpuDevice{LogicID: logicId, FaultCodes: faultCodes, NetworkFaultCodes: networkFaultCodes}
			recoverFaultMap = map[int32][]int64{
				logicId: faultCodes,
			}
			recoverNetworkFaultMap = map[int32][]int64{
				logicId: networkFaultCodes,
			}
			groupDevice := map[string][]*NpuDevice{
				"test": {device},
			}
			DelOnceRecoverFault(groupDevice)
			convey.So(len(device.FaultCodes), convey.ShouldEqual, expectedNum)
			convey.So(len(recoverFaultMap), convey.ShouldEqual, expectedNum)
			convey.So(len(device.NetworkFaultCodes), convey.ShouldEqual, expectedNum)
			convey.So(len(recoverNetworkFaultMap), convey.ShouldEqual, expectedNum)
		})
	})
}

// TestDelOnceFrequencyFault for test TestDelOnceFrequencyFault
func TestDelOnceFrequencyFault(t *testing.T) {
	convey.Convey("test DelOnceFrequencyFault", t, func() {
		convey.Convey("DelOnceFrequencyFault success", func() {
			logicId := int32(0)
			frequencyValue := []int64{3, 5, 6}
			eventId := "80C98008"
			result := 0
			faultFrequencyCache := &FaultFrequencyCache{
				Frequency: map[int32][]int64{
					logicId: frequencyValue,
				},
			}
			faultFrequencyMap = map[string]*FaultFrequencyCache{
				eventId: faultFrequencyCache,
			}
			recoverFaultFrequencyMap = map[int32]string{
				logicId: eventId,
			}
			DelOnceFrequencyFault()
			convey.So(len(faultFrequencyMap[eventId].Frequency[logicId]), convey.ShouldEqual, result)
			convey.So(len(recoverFaultFrequencyMap), convey.ShouldEqual, result)
		})
	})
}

// TestSaveDevFaultInfo for test SaveDevFaultInfo
func TestSaveDevFaultInfo(t *testing.T) {
	convey.Convey("test SaveDevFaultInfo", t, func() {
		convey.Convey("SaveDevFaultInfo success", func() {
			expectedNum0 := 0
			expectedNum1 := 1
			eventId := int64(1)
			devFaultInfoMap = make(map[int32][]common.DevFaultInfo, GeneralMapSize)
			SaveDevFaultInfo(common.DevFaultInfo{})
			convey.So(len(devFaultInfoMap), convey.ShouldEqual, expectedNum0)
			SaveDevFaultInfo(common.DevFaultInfo{EventID: eventId})
			convey.So(len(devFaultInfoMap), convey.ShouldEqual, expectedNum1)
		})
	})
}

// TestTakeOutDevFaultInfo for test TakeOutDevFaultInfo
func TestTakeOutDevFaultInfo(t *testing.T) {
	convey.Convey("test TakeOutDevFaultInfo", t, func() {
		convey.Convey("TakeOutDevFaultInfo success", func() {
			expectedNum := 0
			eventId := int64(1)
			logicId := int32(0)
			devFaultInfoMap = make(map[int32][]common.DevFaultInfo, GeneralMapSize)
			convey.So(len(GetAndCleanFaultInfo()), convey.ShouldEqual, expectedNum)
			testInfo := []common.DevFaultInfo{{EventID: eventId}}
			devFaultInfoMap[logicId] = testInfo
			convey.So(GetAndCleanFaultInfo()[logicId], convey.ShouldResemble, testInfo)
			convey.So(len(devFaultInfoMap[logicId]), convey.ShouldEqual, expectedNum)
		})
	})
}

// TestGetNetworkFaultTypeByCode for test GetNetworkFaultTypeByCode
func TestGetNetworkFaultTypeByCode(t *testing.T) {
	convey.Convey("test GetNetworkFaultTypeByCode", t, func() {
		faultCodes := []int64{LinkDownFaultCode}
		testFaultCode := int64(1)
		convey.Convey("fault type NormalNetwork", func() {
			convey.So(GetNetworkFaultTypeByCode(nil), convey.ShouldEqual, NormalNetwork)
		})
		convey.Convey("fault type NotHandleFault", func() {
			faultTypeCode = FaultTypeCode{
				NotHandleFaultNetworkCodes: faultCodes,
				NotHandleFaultCodes:        []int64{testFaultCode},
			}
			convey.So(GetNetworkFaultTypeByCode(faultCodes), convey.ShouldEqual, NotHandleFault)
		})
		convey.Convey("fault type SeparateNPU", func() {
			faultTypeCode = FaultTypeCode{
				SeparateNPUNetworkCodes: faultCodes,
				NotHandleFaultCodes:     []int64{testFaultCode},
			}
			convey.So(GetNetworkFaultTypeByCode(faultCodes), convey.ShouldEqual, SeparateNPU)
		})
		convey.Convey("fault type PreSeparateNPU", func() {
			faultTypeCode = FaultTypeCode{
				PreSeparateNPUNetworkCodes: faultCodes,
				NotHandleFaultCodes:        []int64{testFaultCode},
			}
			convey.So(GetNetworkFaultTypeByCode(faultCodes), convey.ShouldEqual, PreSeparateNPU)
			faultTypeCode = FaultTypeCode{}
			convey.So(GetNetworkFaultTypeByCode(faultCodes), convey.ShouldEqual, PreSeparateNPU)
		})
		convey.Convey("read json failed", func() {
			faultTypeCode = FaultTypeCode{}
			mockLoadFile := gomonkey.ApplyFuncReturn(utils.LoadFile, nil, errors.New("failed"))
			defer mockLoadFile.Reset()
			convey.So(GetNetworkFaultTypeByCode(faultCodes), convey.ShouldEqual, PreSeparateNPU)
		})
	})
}

// TestDevFaultInfoBasedTimeAscendLen for test DevFaultInfoBasedTimeAscend.Len
func TestDevFaultInfoBasedTimeAscendLen(t *testing.T) {
	convey.Convey("test DevFaultInfoBasedTimeAscend.Len success", t, func() {
		devFault := []common.DevFaultInfo{{}}
		convey.So(DevFaultInfoBasedTimeAscend(devFault).Len(), convey.ShouldEqual, len(devFault))
	})
}

// TestDevFaultInfoBasedTimeAscendSwap for test DevFaultInfoBasedTimeAscend.Swap
func TestDevFaultInfoBasedTimeAscendSwap(t *testing.T) {
	convey.Convey("test DevFaultInfoBasedTimeAscend.Swap success", t, func() {
		devFault := DevFaultInfoBasedTimeAscend([]common.DevFaultInfo{{EventID: 0}, {EventID: 1}})
		iKey, jKey := 0, 1
		if len(devFault) > iKey && len(devFault) > jKey {
			expectIVal, expectJVal := devFault[jKey], devFault[iKey]
			devFault.Swap(iKey, jKey)
			convey.So(devFault[iKey], convey.ShouldResemble, expectIVal)
			convey.So(devFault[jKey], convey.ShouldResemble, expectJVal)
		}
	})
}

// TestDevFaultInfoBasedTimeAscendLess for test DevFaultInfoBasedTimeAscend.Less
func TestDevFaultInfoBasedTimeAscendLess(t *testing.T) {
	convey.Convey("test DevFaultInfoBasedTimeAscend.Less success", t, func() {
		devFault := DevFaultInfoBasedTimeAscend([]common.DevFaultInfo{{AlarmRaisedTime: 0}, {AlarmRaisedTime: 1}})
		iKey, jKey := 0, 1
		convey.So(devFault.Less(iKey, jKey), convey.ShouldBeTrue)
	})
}

// TestQueryManuallyFaultInfoByLogicID for test QueryManuallyFaultInfoByLogicID
func TestQueryManuallyFaultInfoByLogicID(t *testing.T) {
	convey.Convey("test QueryManuallyFaultInfoByLogicID", t, func() {
		convey.Convey("test valid logicID", func() {
			logicID := int32(10)
			_, ok := manuallySeparateNpuMap[logicID]
			convey.So(QueryManuallyFaultInfoByLogicID(logicID), convey.ShouldEqual, ok)
		})
		convey.Convey("test invalid logicID", func() {
			logicID := int32(20)
			convey.So(QueryManuallyFaultInfoByLogicID(logicID), convey.ShouldBeFalse)
		})
	})
}

// TestSetManuallyFaultNPUHandled for test SetManuallyFaultNPUHandled
func TestSetManuallyFaultNPUHandled(t *testing.T) {
	convey.Convey("test SetManuallyFaultNPUHandled success", t, func() {
		logicId := int32(0)
		manuallySeparateNpuMap = map[int32]ManuallyFaultInfo{logicId: {FirstHandle: true}}
		expectVal := map[int32]ManuallyFaultInfo{logicId: {FirstHandle: false}}
		SetManuallyFaultNPUHandled()
		convey.So(manuallySeparateNpuMap, convey.ShouldResemble, expectVal)
	})
}

// TestCollectEachFaultEvent for test collectEachFaultEvent
func TestCollectEachFaultEvent(t *testing.T) {
	convey.Convey("test collectEachFaultEvent success", t, func() {
		logicID := int32(0)
		linkDownFaultTimeout := int64(30)
		linkDownRecoverTimeout := int64(60)
		linkDownFaultCodeStr := strings.ToLower(strconv.FormatInt(LinkDownFaultCode, Hex))
		faultDurationMap = map[string]*FaultDurationCache{
			linkDownFaultCodeStr: {
				FaultDuration: FaultDuration{
					FaultTimeout:   linkDownFaultTimeout,
					RecoverTimeout: linkDownRecoverTimeout,
					FaultHandling:  PreSeparateNPU,
				},
			},
		}
		faultInfos := []common.DevFaultInfo{
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur},
			{EventID: CardDropFaultCode, Assertion: common.FaultOccur},
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover},
		}
		collectEachFaultEvent(logicID, faultInfos)
		expected1 := 1
		expected2 := 2
		convey.So(len(faultDurationMap), convey.ShouldEqual, expected1)
		convey.So(len(faultDurationMap[linkDownFaultCodeStr].Duration[logicID].FaultEventQueue),
			convey.ShouldEqual, expected2)
	})
}

// TestSortFaultEventsInAscendingOrder for test sortFaultEventsInAscendingOrder
func TestSortFaultEventsInAscendingOrder(t *testing.T) {
	convey.Convey("test sortFaultEventsInAscendingOrder success", t, func() {
		logicID := int32(0)
		linkDownFaultCodeStr := strings.ToLower(strconv.FormatInt(LinkDownFaultCode, Hex))
		CardDropFaultCodeStr := strings.ToLower(strconv.FormatInt(CardDropFaultCode, Hex))
		faultDurationMap = map[string]*FaultDurationCache{
			linkDownFaultCodeStr: {
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU,
				},
			},
			CardDropFaultCodeStr: {
				FaultDuration: FaultDuration{
					FaultTimeout:   120,
					RecoverTimeout: 0,
					FaultHandling:  SeparateNPU,
				},
			},
		}
		faultInfos := []common.DevFaultInfo{
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur, AlarmRaisedTime: 4},
			{EventID: CardDropFaultCode, Assertion: common.FaultRecover, AlarmRaisedTime: 3},
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover, AlarmRaisedTime: 1},
			{EventID: CardDropFaultCode, Assertion: common.FaultOccur, AlarmRaisedTime: 2},
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover, AlarmRaisedTime: 9},
		}

		linkDownFaultExpectVal := []common.DevFaultInfo{{EventID: LinkDownFaultCode, Assertion: common.FaultRecover,
			AlarmRaisedTime: 1}, {EventID: LinkDownFaultCode, Assertion: common.FaultOccur, AlarmRaisedTime: 4}, {
			EventID: LinkDownFaultCode, Assertion: common.FaultRecover, AlarmRaisedTime: 9}}

		cardDropFaultExpectVal := []common.DevFaultInfo{{EventID: CardDropFaultCode, Assertion: common.FaultOccur,
			AlarmRaisedTime: 2}, {EventID: CardDropFaultCode, Assertion: common.FaultRecover, AlarmRaisedTime: 3}}

		collectEachFaultEvent(logicID, faultInfos)
		sortFaultEventsInAscendingOrder(logicID, linkDownFaultCodeStr)
		sortFaultEventsInAscendingOrder(logicID, CardDropFaultCodeStr)
		convey.So(faultDurationMap[linkDownFaultCodeStr].Duration[logicID].FaultEventQueue, convey.ShouldResemble,
			linkDownFaultExpectVal)
		convey.So(faultDurationMap[CardDropFaultCodeStr].Duration[logicID].FaultEventQueue, convey.ShouldResemble,
			cardDropFaultExpectVal)
	})
}

// TestMergeContinuousElementBasedAssertion for test mergeContinuousElementBasedAssertion
func TestMergeContinuousElementBasedAssertion(t *testing.T) {
	convey.Convey("test merge fault occur continuous assertion success", t, func() {
		devFaultInfo := []common.DevFaultInfo{{Assertion: common.FaultOccur}, {Assertion: common.FaultOccur}}
		expectVal := []common.DevFaultInfo{{Assertion: common.FaultOccur}}
		mergeContinuousElementBasedAssertion(&devFaultInfo)
		convey.So(devFaultInfo, convey.ShouldResemble, expectVal)
	})

	convey.Convey("test merge fault recover continuous assertion success", t, func() {
		devFaultInfo := []common.DevFaultInfo{{Assertion: common.FaultRecover}, {Assertion: common.FaultRecover},
			{Assertion: common.FaultRecover}}
		expectVal := []common.DevFaultInfo{{Assertion: common.FaultRecover}}
		mergeContinuousElementBasedAssertion(&devFaultInfo)
		convey.So(devFaultInfo, convey.ShouldResemble, expectVal)
	})

	convey.Convey("test merge mix fault continuous assertion success", t, func() {
		devFaultInfo := []common.DevFaultInfo{{Assertion: common.FaultRecover}, {Assertion: common.FaultRecover},
			{Assertion: common.FaultOccur}, {Assertion: common.FaultOccur}}
		expectVal := []common.DevFaultInfo{{Assertion: common.FaultRecover}, {Assertion: common.FaultOccur}}
		mergeContinuousElementBasedAssertion(&devFaultInfo)
		convey.So(devFaultInfo, convey.ShouldResemble, expectVal)
	})
}

// TestClearFirstEventBasedOnFaultStatus for test clearFirstEventBasedOnFaultStatus
func TestClearFirstEventBasedOnFaultStatus(t *testing.T) {
	convey.Convey("test clearFirstEventBasedOnFaultStatus timeout success", t, func() {
		faultDurationData := FaultDurationData{
			TimeoutStatus: false,
			FaultEventQueue: []common.DevFaultInfo{{
				EventID:         LinkDownFaultCode,
				Assertion:       common.FaultRecover,
				AlarmRaisedTime: 2}, {
				EventID:         LinkDownFaultCode,
				Assertion:       common.FaultOccur,
				AlarmRaisedTime: 4,
			}}}
		expectVal := []common.DevFaultInfo{{
			EventID:         LinkDownFaultCode,
			Assertion:       common.FaultOccur,
			AlarmRaisedTime: 4}}
		clearFirstEventBasedOnFaultStatus(&faultDurationData)
		convey.So(faultDurationData.FaultEventQueue, convey.ShouldResemble, expectVal)
	})

	convey.Convey("test clearFirstEventBasedOnFaultStatus recover success", t, func() {
		faultDurationData := FaultDurationData{
			TimeoutStatus: true,
			FaultEventQueue: []common.DevFaultInfo{{
				EventID:         LinkDownFaultCode,
				Assertion:       common.FaultOccur,
				AlarmRaisedTime: 2}, {
				EventID:         LinkDownFaultCode,
				Assertion:       common.FaultRecover,
				AlarmRaisedTime: 4,
			}}}
		expectVal := []common.DevFaultInfo{{
			EventID:         LinkDownFaultCode,
			Assertion:       common.FaultRecover,
			AlarmRaisedTime: 4}}
		clearFirstEventBasedOnFaultStatus(&faultDurationData)
		convey.So(faultDurationData.FaultEventQueue, convey.ShouldResemble, expectVal)
	})
}

// TestCleanFaultQueue for test cleanFaultQueue when fault time status is false
func TestCleanFaultQueueWhenFaultTimeStatusFalse(t *testing.T) {
	convey.Convey("test CleanFaultQueue when fault time status is false", t, func() {
		logicID := int32(0)
		linkDownFaultCodeStr := strings.ToLower(strconv.FormatInt(LinkDownFaultCode, Hex))
		CardDropFaultCodeStr := strings.ToLower(strconv.FormatInt(CardDropFaultCode, Hex))
		faultDurationMap = map[string]*FaultDurationCache{
			linkDownFaultCodeStr: {
				Duration: map[int32]FaultDurationData{logicID: {TimeoutStatus: false}},
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU,
				},
			},
			CardDropFaultCodeStr: {
				FaultDuration: FaultDuration{
					FaultTimeout:   120,
					RecoverTimeout: 0,
					FaultHandling:  SeparateNPU,
				},
			},
		}

		faultInfos := []common.DevFaultInfo{
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur, AlarmRaisedTime: 165},
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover, AlarmRaisedTime: 100},
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur, AlarmRaisedTime: 150},
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur, AlarmRaisedTime: 150},
		}
		expectVal := []common.DevFaultInfo{{
			EventID:         LinkDownFaultCode,
			Assertion:       common.FaultOccur,
			AlarmRaisedTime: 150}}

		collectEachFaultEvent(logicID, faultInfos)
		sortFaultEventsInAscendingOrder(logicID, linkDownFaultCodeStr)
		cleanFaultQueue(logicID, linkDownFaultCodeStr)
		convey.So(faultDurationMap[linkDownFaultCodeStr].Duration[logicID].FaultEventQueue,
			convey.ShouldResemble, expectVal)
	})
}

// TestCleanFaultQueue for test cleanFaultQueue when fault time status is true
func TestCleanFaultQueueWhenFaultTimeStatusTrue(t *testing.T) {
	convey.Convey("test CleanFaultQueue when fault time status is true", t, func() {
		logicID := int32(0)
		linkDownFaultCodeStr := strings.ToLower(strconv.FormatInt(LinkDownFaultCode, Hex))
		faultDurationMap = map[string]*FaultDurationCache{
			linkDownFaultCodeStr: {
				Duration: map[int32]FaultDurationData{logicID: {TimeoutStatus: true,
					FaultEventQueue: []common.DevFaultInfo{}}},
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU,
				},
			},
		}

		faultInfos := []common.DevFaultInfo{
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur, AlarmRaisedTime: 165},
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover, AlarmRaisedTime: 100},
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur, AlarmRaisedTime: 150},
		}
		expectVal := []common.DevFaultInfo{{
			EventID:         LinkDownFaultCode,
			Assertion:       common.FaultRecover,
			AlarmRaisedTime: 100}, {
			EventID:         LinkDownFaultCode,
			Assertion:       common.FaultOccur,
			AlarmRaisedTime: 150,
		}}

		collectEachFaultEvent(logicID, faultInfos)
		sortFaultEventsInAscendingOrder(logicID, linkDownFaultCodeStr)
		cleanFaultQueue(logicID, linkDownFaultCodeStr)
		convey.So(faultDurationMap[linkDownFaultCodeStr].Duration[logicID].FaultEventQueue,
			convey.ShouldResemble, expectVal)
	})
}

// TestHandleFaultQueueCase01 for test handleFaultQueue case 01
func TestHandleFaultQueueCase01(t *testing.T) {
	convey.Convey("test handleFaultQueue case 01", t, func() {
		logicID := int32(0)
		alarmRaisedTime50, alarmRaisedTime81, alarmRaisedTime82 := int64(50), int64(81), int64(82)
		linkDownFaultCodeStr := strings.ToLower(strconv.FormatInt(LinkDownFaultCode, Hex))
		faultDurationMap = map[string]*FaultDurationCache{
			linkDownFaultCodeStr: {
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU,
				},
			},
		}

		faultInfos := []common.DevFaultInfo{
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur,
				AlarmRaisedTime: alarmRaisedTime50 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover,
				AlarmRaisedTime: alarmRaisedTime81 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur,
				AlarmRaisedTime: alarmRaisedTime82 * SecondMagnification},
		}
		expectVal := []common.DevFaultInfo{
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur,
				AlarmRaisedTime: alarmRaisedTime50 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover,
				AlarmRaisedTime: alarmRaisedTime81 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur,
				AlarmRaisedTime: alarmRaisedTime82 * SecondMagnification},
		}

		collectEachFaultEvent(logicID, faultInfos)
		sortFaultEventsInAscendingOrder(logicID, linkDownFaultCodeStr)
		cleanFaultQueue(logicID, linkDownFaultCodeStr)
		handleFaultQueue(logicID, linkDownFaultCodeStr)

		faultDurationData := faultDurationMap[linkDownFaultCodeStr].Duration[logicID]
		faultDurationTime := int64(31)
		convey.So(faultDurationData.TimeoutStatus, convey.ShouldEqual, true)
		convey.So(faultDurationData.FaultEventQueue, convey.ShouldResemble, expectVal)
		convey.So(faultDurationData.FaultDurationTime, convey.ShouldEqual, faultDurationTime*SecondMagnification)
		convey.So(faultDurationData.FaultRecoverDurationTime, convey.ShouldEqual, 0)
	})
}

// TestHandleFaultQueueCase02 for test handleFaultQueue case 02
func TestHandleFaultQueueCase02(t *testing.T) {
	convey.Convey("test handleFaultQueue case 02", t, func() {
		logicID := int32(0)
		alarmRaisedTime30, alarmRaisedTime50, alarmRaisedTime80 := int64(30), int64(50), int64(80)
		linkDownFaultCodeStr := strings.ToLower(strconv.FormatInt(LinkDownFaultCode, Hex))
		faultDurationMap = map[string]*FaultDurationCache{
			linkDownFaultCodeStr: {
				Duration: map[int32]FaultDurationData{logicID: {FaultEventQueue: []common.DevFaultInfo{}}},
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU,
				},
			},
		}

		faultInfos := []common.DevFaultInfo{
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur,
				AlarmRaisedTime: alarmRaisedTime50 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover,
				AlarmRaisedTime: alarmRaisedTime80 * SecondMagnification},
		}
		expectVal := make([]common.DevFaultInfo, 0)

		collectEachFaultEvent(logicID, faultInfos)
		sortFaultEventsInAscendingOrder(logicID, linkDownFaultCodeStr)
		cleanFaultQueue(logicID, linkDownFaultCodeStr)
		handleFaultQueue(logicID, linkDownFaultCodeStr)

		faultDurationData := faultDurationMap[linkDownFaultCodeStr].Duration[logicID]
		convey.So(faultDurationData.TimeoutStatus, convey.ShouldEqual, false)
		convey.So(faultDurationData.FaultEventQueue, convey.ShouldResemble, expectVal)
		convey.So(faultDurationData.FaultDurationTime, convey.ShouldEqual, alarmRaisedTime30*SecondMagnification)
		convey.So(faultDurationData.FaultRecoverDurationTime, convey.ShouldEqual, 0)
	})
}

// TestHandleFaultQueueCase03 for test handleFaultQueue case 03
func TestHandleFaultQueueCase03(t *testing.T) {
	convey.Convey("test handleFaultQueue case 03", t, func() {
		logicID := int32(0)
		alarmRaisedTime50, alarmRaisedTime80, alarmRaisedTime82, alarmRaisedTime112 :=
			int64(50), int64(80), int64(82), int64(112)
		linkDownFaultCodeStr := strings.ToLower(strconv.FormatInt(LinkDownFaultCode, Hex))
		faultDurationMap = map[string]*FaultDurationCache{
			linkDownFaultCodeStr: {
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU,
				},
			},
		}

		faultInfos := []common.DevFaultInfo{
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur,
				AlarmRaisedTime: alarmRaisedTime50 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover,
				AlarmRaisedTime: alarmRaisedTime80 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur,
				AlarmRaisedTime: alarmRaisedTime82 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover,
				AlarmRaisedTime: alarmRaisedTime112 * SecondMagnification},
		}
		expectVal := make([]common.DevFaultInfo, 0)

		collectEachFaultEvent(logicID, faultInfos)
		sortFaultEventsInAscendingOrder(logicID, linkDownFaultCodeStr)
		cleanFaultQueue(logicID, linkDownFaultCodeStr)
		handleFaultQueue(logicID, linkDownFaultCodeStr)

		alarmRaisedTime30 := int64(30)
		faultDurationData := faultDurationMap[linkDownFaultCodeStr].Duration[logicID]
		convey.So(faultDurationData.TimeoutStatus, convey.ShouldEqual, false)
		convey.So(faultDurationData.FaultEventQueue, convey.ShouldResemble, expectVal)
		convey.So(faultDurationData.FaultDurationTime, convey.ShouldEqual, alarmRaisedTime30*SecondMagnification)
		convey.So(faultDurationData.FaultRecoverDurationTime, convey.ShouldEqual, 0)
	})
}

// TestHandleFaultQueueCase04 for test handleFaultQueue case 04
func TestHandleFaultQueueCase04(t *testing.T) {
	convey.Convey("test handleFaultQueue case 04", t, func() {
		logicID := int32(0)
		alarmRaisedTime50, AlarmRaisedTime110 := int64(50), int64(110)
		linkDownFaultCodeStr := strings.ToLower(strconv.FormatInt(LinkDownFaultCode, Hex))
		faultDurationMap = map[string]*FaultDurationCache{
			linkDownFaultCodeStr: {
				Duration: map[int32]FaultDurationData{logicID: {TimeoutStatus: true}},
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU,
				},
			},
		}

		faultInfos := []common.DevFaultInfo{
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover,
				AlarmRaisedTime: alarmRaisedTime50 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur,
				AlarmRaisedTime: AlarmRaisedTime110 * SecondMagnification},
		}
		expectVal := make([]common.DevFaultInfo, 0)

		collectEachFaultEvent(logicID, faultInfos)
		sortFaultEventsInAscendingOrder(logicID, linkDownFaultCodeStr)
		cleanFaultQueue(logicID, linkDownFaultCodeStr)
		handleFaultQueue(logicID, linkDownFaultCodeStr)

		alarmRaisedTime60 := int64(60)
		faultDurationData := faultDurationMap[linkDownFaultCodeStr].Duration[logicID]
		convey.So(faultDurationData.TimeoutStatus, convey.ShouldEqual, true)
		convey.So(faultDurationData.FaultEventQueue, convey.ShouldResemble, expectVal)
		convey.So(faultDurationData.FaultDurationTime, convey.ShouldEqual, 0)
		convey.So(faultDurationData.FaultRecoverDurationTime, convey.ShouldEqual, alarmRaisedTime60*SecondMagnification)
	})
}

// TestHandleFaultQueueCase05 for test handleFaultQueue case 05
func TestHandleFaultQueueCase05(t *testing.T) {
	convey.Convey("test handleFaultQueue case 05", t, func() {
		logicID := int32(0)
		linkDownFaultCodeStr := strings.ToLower(strconv.FormatInt(LinkDownFaultCode, Hex))
		AlarmRaisedTime50, AlarmRaisedTime111, AlarmRaisedTime112 := int64(50), int64(111), int64(112)
		faultDurationMap = map[string]*FaultDurationCache{
			linkDownFaultCodeStr: {
				Duration: map[int32]FaultDurationData{logicID: {TimeoutStatus: true}},
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU,
				},
			},
		}

		faultInfos := []common.DevFaultInfo{
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover,
				AlarmRaisedTime: AlarmRaisedTime50 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur,
				AlarmRaisedTime: AlarmRaisedTime111 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover,
				AlarmRaisedTime: AlarmRaisedTime112 * SecondMagnification},
		}
		expectVal := make([]common.DevFaultInfo, 0)

		collectEachFaultEvent(logicID, faultInfos)
		sortFaultEventsInAscendingOrder(logicID, linkDownFaultCodeStr)
		cleanFaultQueue(logicID, linkDownFaultCodeStr)
		handleFaultQueue(logicID, linkDownFaultCodeStr)

		AlarmRaisedTime1, AlarmRaisedTime61 := int64(1), int64(61)
		faultDurationData := faultDurationMap[linkDownFaultCodeStr].Duration[logicID]
		convey.So(faultDurationData.TimeoutStatus, convey.ShouldEqual, false)
		convey.So(faultDurationData.FaultEventQueue, convey.ShouldResemble, expectVal)
		convey.So(faultDurationData.FaultDurationTime, convey.ShouldEqual,
			AlarmRaisedTime1*SecondMagnification)
		convey.So(faultDurationData.FaultRecoverDurationTime, convey.ShouldEqual,
			AlarmRaisedTime61*SecondMagnification)
	})
}

// TestHandleFaultQueueCase06 for test handleFaultQueue case 06
func TestHandleFaultQueueCase06(t *testing.T) {
	convey.Convey("test handleFaultQueue case 06", t, func() {
		logicID := int32(0)
		linkDownFaultCodeStr := strings.ToLower(strconv.FormatInt(LinkDownFaultCode, Hex))
		faultDurationMap = map[string]*FaultDurationCache{
			linkDownFaultCodeStr: {
				Duration: map[int32]FaultDurationData{logicID: {TimeoutStatus: true,
					FaultEventQueue: []common.DevFaultInfo{}}},
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU,
				},
			},
		}
		AlarmRaisedTime50, AlarmRaisedTime111, AlarmRaisedTime142 := int64(50), int64(111), int64(142)
		faultInfos := []common.DevFaultInfo{
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover,
				AlarmRaisedTime: AlarmRaisedTime50 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur,
				AlarmRaisedTime: AlarmRaisedTime111 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover,
				AlarmRaisedTime: AlarmRaisedTime142 * SecondMagnification},
		}
		expectVal := []common.DevFaultInfo{
			{EventID: LinkDownFaultCode, Assertion: common.FaultOccur,
				AlarmRaisedTime: AlarmRaisedTime111 * SecondMagnification},
			{EventID: LinkDownFaultCode, Assertion: common.FaultRecover,
				AlarmRaisedTime: AlarmRaisedTime142 * SecondMagnification},
		}

		collectEachFaultEvent(logicID, faultInfos)
		sortFaultEventsInAscendingOrder(logicID, linkDownFaultCodeStr)
		cleanFaultQueue(logicID, linkDownFaultCodeStr)
		handleFaultQueue(logicID, linkDownFaultCodeStr)

		faultDurationData := faultDurationMap[linkDownFaultCodeStr].Duration[logicID]
		AlarmRaisedTime31, AlarmRaisedTime61 := int64(31), int64(61)
		convey.So(faultDurationData.TimeoutStatus, convey.ShouldEqual, true)
		convey.So(faultDurationData.FaultEventQueue, convey.ShouldResemble, expectVal)
		convey.So(faultDurationData.FaultDurationTime, convey.ShouldEqual,
			AlarmRaisedTime31*SecondMagnification)
		convey.So(faultDurationData.FaultRecoverDurationTime, convey.ShouldEqual,
			AlarmRaisedTime61*SecondMagnification)
	})
}

// TestResetFaultCustomizationCache for test ResetFaultCustomizationCache
func TestResetFaultCustomizationCache(t *testing.T) {
	convey.Convey("test ResetFaultCustomizationCache success", t, func() {
		faultFrequencyMap = map[string]*FaultFrequencyCache{
			"80E18005": {
				Frequency: make(map[int32][]int64, GeneralMapSize),
				FaultFrequency: FaultFrequency{
					TimeWindow:    86400,
					Times:         2,
					FaultHandling: ManuallySeparateNPU,
				},
			},
		}
		faultDurationMap = map[string]*FaultDurationCache{
			"81078603": {
				Duration: make(map[int32]FaultDurationData, GeneralMapSize),
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU,
				},
			},
		}

		expectVal := 0
		ResetFaultCustomizationCache()
		convey.So(len(faultFrequencyMap), convey.ShouldEqual, expectVal)
		convey.So(len(faultDurationMap), convey.ShouldEqual, expectVal)
	})
}

// TestSaveManuallyFaultInfo for test SaveManuallyFaultInfo
func TestSaveManuallyFaultInfo(t *testing.T) {
	convey.Convey("test SaveManuallyFaultInfo", t, func() {
		convey.Convey("test valid logicID", func() {
			manuallySeparateNpuMap = make(map[int32]ManuallyFaultInfo, GeneralMapSize)
			logicID, expectVal := int32(10), 1
			SaveManuallyFaultInfo(logicID)
			convey.So(len(manuallySeparateNpuMap), convey.ShouldEqual, expectVal)
		})
		convey.Convey("test invalid logicID", func() {
			manuallySeparateNpuMap = make(map[int32]ManuallyFaultInfo, GeneralMapSize)
			logicID, expectVal := int32(20), 0
			SaveManuallyFaultInfo(logicID)
			convey.So(len(manuallySeparateNpuMap), convey.ShouldEqual, expectVal)
		})
	})
}

// TestDeleteManuallyFaultInfo for test DeleteManuallyFaultInfo
func TestDeleteManuallyFaultInfo(t *testing.T) {
	convey.Convey("test DeleteManuallyFaultInfo", t, func() {
		convey.Convey("test valid logicID", func() {
			manuallySeparateNpuMap = make(map[int32]ManuallyFaultInfo, GeneralMapSize)
			logicID, expectVal := int32(10), 1
			SaveManuallyFaultInfo(logicID)
			convey.So(len(manuallySeparateNpuMap), convey.ShouldEqual, expectVal)
		})
		convey.Convey("test invalid logicID", func() {
			manuallySeparateNpuMap = make(map[int32]ManuallyFaultInfo, GeneralMapSize)
			logicID, expectVal := int32(20), 0
			SaveManuallyFaultInfo(logicID)
			convey.So(len(manuallySeparateNpuMap), convey.ShouldEqual, expectVal)
		})
	})
}

// TestLoadGraceToleranceCustomization for test loadGraceToleranceCustomization
func TestLoadGraceToleranceCustomization(t *testing.T) {
	convey.Convey("test loadGraceToleranceCustomization success", t, func() {
		graceToleranceCustomization := GraceToleranceCustomization{
			WaitDeviceResetTime:      150,
			WaitProcessReadCMTime:    30,
			WaitFaultSelfHealingTime: 15,
		}
		WaitDeviceResetTime = time.Duration(0)
		WaitProcessReadCMTime = time.Duration(0)
		WaitFaultSelfHealingTime = time.Duration(0)
		loadGraceToleranceCustomization(graceToleranceCustomization)
		expectResetTime, expectReadCMTime, expectSelfHealingTime := 150, 30, 15
		convey.So(WaitDeviceResetTime, convey.ShouldEqual, expectResetTime)
		convey.So(WaitProcessReadCMTime, convey.ShouldEqual, expectReadCMTime)
		convey.So(WaitFaultSelfHealingTime, convey.ShouldEqual, expectSelfHealingTime)
	})

	convey.Convey("test loadGraceToleranceCustomization abnormal condition success", t, func() {
		graceToleranceCustomization := GraceToleranceCustomization{
			WaitDeviceResetTime:      59,
			WaitProcessReadCMTime:    91,
			WaitFaultSelfHealingTime: 0,
		}
		WaitDeviceResetTime = time.Duration(0)
		WaitProcessReadCMTime = time.Duration(0)
		WaitFaultSelfHealingTime = time.Duration(0)
		loadGraceToleranceCustomization(graceToleranceCustomization)
		expectResetTime, expectReadCMTime, expectSelfHealingTime := 150, 30, 15
		convey.So(WaitDeviceResetTime, convey.ShouldEqual, expectResetTime)
		convey.So(WaitProcessReadCMTime, convey.ShouldEqual, expectReadCMTime)
		convey.So(WaitFaultSelfHealingTime, convey.ShouldEqual, expectSelfHealingTime)
	})
}

// TestValidateFaultFrequencyCustomizationPart1 for test validateFaultFrequencyCustomization
func TestValidateFaultFrequencyCustomizationPart1(t *testing.T) {
	convey.Convey("test validateFaultFrequencyCustomization success", t, func() {
		faultFrequencyCustomization := FaultFrequencyCustomization{
			EventId: []string{"80C98000", "80B78000"},
			FaultFrequency: FaultFrequency{
				TimeWindow:    86400,
				Times:         3,
				FaultHandling: ManuallySeparateNPU,
			},
		}
		result := validateFaultFrequencyCustomization(faultFrequencyCustomization)
		convey.So(result, convey.ShouldEqual, true)
	})
}

// TestValidateFaultFrequencyCustomizationPart2 for test validateFaultFrequencyCustomization
func TestValidateFaultFrequencyCustomizationPart2(t *testing.T) {
	convey.Convey("test validateFaultFrequencyCustomization failed case1", t, func() {
		faultFrequencyCustomization := FaultFrequencyCustomization{
			EventId: []string{},
			FaultFrequency: FaultFrequency{
				TimeWindow:    86400,
				Times:         3,
				FaultHandling: ManuallySeparateNPU,
			},
		}
		result := validateFaultFrequencyCustomization(faultFrequencyCustomization)
		convey.So(result, convey.ShouldEqual, false)
	})
	convey.Convey("test validateFaultFrequencyCustomization failed case2", t, func() {
		faultFrequencyCustomization := FaultFrequencyCustomization{
			EventId: []string{"80C98000", "80B78000"},
			FaultFrequency: FaultFrequency{
				TimeWindow:    59,
				Times:         3,
				FaultHandling: ManuallySeparateNPU,
			},
		}
		result := validateFaultFrequencyCustomization(faultFrequencyCustomization)
		convey.So(result, convey.ShouldEqual, false)
	})
	convey.Convey("test validateFaultFrequencyCustomization failed case3", t, func() {
		faultFrequencyCustomization := FaultFrequencyCustomization{
			EventId: []string{"80C98000", "80B78000"},
			FaultFrequency: FaultFrequency{
				TimeWindow:    60,
				Times:         0,
				FaultHandling: ManuallySeparateNPU,
			},
		}
		result := validateFaultFrequencyCustomization(faultFrequencyCustomization)
		convey.So(result, convey.ShouldEqual, false)
	})
	convey.Convey("test validateFaultFrequencyCustomization failed case4", t, func() {
		faultFrequencyCustomization := FaultFrequencyCustomization{
			EventId: []string{"80C98000", "80B78000"},
			FaultFrequency: FaultFrequency{
				TimeWindow:    60,
				Times:         2,
				FaultHandling: "separatesNPU",
			},
		}
		result := validateFaultFrequencyCustomization(faultFrequencyCustomization)
		convey.So(result, convey.ShouldEqual, false)
	})
}

// TestLoadFaultFrequencyCustomizationCase1 for test loadFaultFrequencyCustomization
func TestLoadFaultFrequencyCustomizationCase1(t *testing.T) {
	convey.Convey("test loadFaultFrequencyCustomization success case1", t, func() {
		faultCode1 := "80C98000"
		faultCode2 := "80E18005"
		faultFrequencyCustomizations := []FaultFrequencyCustomization{{
			EventId: []string{faultCode1}, FaultFrequency: FaultFrequency{
				TimeWindow:    86400,
				Times:         2,
				FaultHandling: ManuallySeparateNPU,
			}}, {
			EventId: []string{faultCode2, faultCode1}, FaultFrequency: FaultFrequency{
				TimeWindow:    86400,
				Times:         3,
				FaultHandling: ManuallySeparateNPU}}}
		expectVal := map[string]*FaultFrequencyCache{
			strings.ToLower(faultCode1): {Frequency: make(map[int32][]int64, common.MaxErrorCodeCount),
				FaultFrequency: FaultFrequency{
					TimeWindow:    86400,
					Times:         2,
					FaultHandling: ManuallySeparateNPU}},
			strings.ToLower(faultCode2): {Frequency: make(map[int32][]int64, common.MaxErrorCodeCount),
				FaultFrequency: FaultFrequency{
					TimeWindow:    86400,
					Times:         3,
					FaultHandling: ManuallySeparateNPU}}}
		faultFrequencyMap = make(map[string]*FaultFrequencyCache, common.MaxErrorCodeCount)
		loadFaultFrequencyCustomization(faultFrequencyCustomizations)
		convey.So(faultFrequencyMap, convey.ShouldResemble, expectVal)
	})
}

// TestLoadFaultFrequencyCustomizationCase2 for test loadFaultFrequencyCustomization
func TestLoadFaultFrequencyCustomizationCase2(t *testing.T) {
	convey.Convey("test loadFaultFrequencyCustomization success case2", t, func() {
		faultCode1 := "80C98000"
		faultCode2 := "80E18005"
		faultFrequencyCustomizations := []FaultFrequencyCustomization{{
			EventId: []string{faultCode1}, FaultFrequency: FaultFrequency{
				TimeWindow:    86400,
				Times:         0,
				FaultHandling: ManuallySeparateNPU,
			}}, {
			EventId: []string{faultCode2}, FaultFrequency: FaultFrequency{
				TimeWindow:    86400,
				Times:         3,
				FaultHandling: ManuallySeparateNPU}}}
		expectVal := map[string]*FaultFrequencyCache{
			strings.ToLower(faultCode2): {Frequency: make(map[int32][]int64, common.MaxErrorCodeCount),
				FaultFrequency: FaultFrequency{
					TimeWindow:    86400,
					Times:         3,
					FaultHandling: ManuallySeparateNPU}}}
		faultFrequencyMap = map[string]*FaultFrequencyCache{
			strings.ToLower(faultCode1): {Frequency: make(map[int32][]int64, common.MaxErrorCodeCount),
				FaultFrequency: FaultFrequency{
					TimeWindow:    86400,
					Times:         2,
					FaultHandling: ManuallySeparateNPU}},
			strings.ToLower(faultCode2): {Frequency: make(map[int32][]int64, common.MaxErrorCodeCount),
				FaultFrequency: FaultFrequency{
					TimeWindow:    86400,
					Times:         3,
					FaultHandling: ManuallySeparateNPU}}}
		loadFaultFrequencyCustomization(faultFrequencyCustomizations)
		convey.So(faultFrequencyMap, convey.ShouldResemble, expectVal)
	})
}

// TestValidateFaultDurationCustomizationPart1 for test validateFaultDurationCustomization
func TestValidateFaultDurationCustomizationPart1(t *testing.T) {
	convey.Convey("test validateFaultFrequencyCustomization success", t, func() {
		faultFrequencyCustomization := FaultDurationCustomization{
			EventId: []string{"81078603"},
			FaultDuration: FaultDuration{
				FaultTimeout:   30,
				RecoverTimeout: 60,
				FaultHandling:  PreSeparateNPU,
			},
		}
		result := validateFaultDurationCustomization(faultFrequencyCustomization)
		convey.So(result, convey.ShouldEqual, true)
	})
}

// TestValidateFaultDurationCustomization for test validateFaultDurationCustomization
func TestValidateFaultDurationCustomizationPart2(t *testing.T) {
	convey.Convey("test validateFaultFrequencyCustomization failed case1", t, func() {
		faultFrequencyCustomization := FaultDurationCustomization{
			EventId: []string{},
			FaultDuration: FaultDuration{
				FaultTimeout:   30,
				RecoverTimeout: 60,
				FaultHandling:  PreSeparateNPU,
			},
		}
		result := validateFaultDurationCustomization(faultFrequencyCustomization)
		convey.So(result, convey.ShouldEqual, false)
	})
	convey.Convey("test validateFaultFrequencyCustomization failed case2", t, func() {
		faultFrequencyCustomization := FaultDurationCustomization{
			EventId: []string{"81078603"},
			FaultDuration: FaultDuration{
				FaultTimeout:   -1,
				RecoverTimeout: 60,
				FaultHandling:  PreSeparateNPU,
			},
		}
		result := validateFaultDurationCustomization(faultFrequencyCustomization)
		convey.So(result, convey.ShouldEqual, false)
	})
	convey.Convey("test validateFaultFrequencyCustomization failed case3", t, func() {
		faultFrequencyCustomization := FaultDurationCustomization{
			EventId: []string{"81078603"},
			FaultDuration: FaultDuration{
				FaultTimeout:   30,
				RecoverTimeout: -1,
				FaultHandling:  PreSeparateNPU,
			},
		}
		result := validateFaultDurationCustomization(faultFrequencyCustomization)
		convey.So(result, convey.ShouldEqual, false)
	})
	convey.Convey("test validateFaultFrequencyCustomization failed case4", t, func() {
		faultFrequencyCustomization := FaultDurationCustomization{
			EventId: []string{"81078603"},
			FaultDuration: FaultDuration{
				FaultTimeout:   30,
				RecoverTimeout: 60,
				FaultHandling:  ManuallySeparateNPU,
			},
		}
		result := validateFaultDurationCustomization(faultFrequencyCustomization)
		convey.So(result, convey.ShouldEqual, false)
	})
}

// TestLoadFaultDurationCustomizationCase1 for test loadFaultDurationCustomizationCase1
func TestLoadFaultDurationCustomizationCase1(t *testing.T) {
	convey.Convey("test loadFaultDurationCustomizationCase1 success case1", t, func() {
		faultCode1 := "81078603"
		faultCode2 := "80E0180F"
		faultDurationCustomization := []FaultDurationCustomization{{
			EventId: []string{faultCode1}, FaultDuration: FaultDuration{
				FaultTimeout:   30,
				RecoverTimeout: 60,
				FaultHandling:  PreSeparateNPU,
			}}, {
			EventId: []string{faultCode2, faultCode1}, FaultDuration: FaultDuration{
				FaultTimeout:   120,
				RecoverTimeout: 0,
				FaultHandling:  RestartBusiness}}}
		expectVal := map[string]*FaultDurationCache{
			strings.ToLower(faultCode1): {Duration: make(map[int32]FaultDurationData, common.MaxErrorCodeCount),
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU}},
			strings.ToLower(faultCode2): {Duration: make(map[int32]FaultDurationData, common.MaxErrorCodeCount),
				FaultDuration: FaultDuration{
					FaultTimeout:   120,
					RecoverTimeout: 0,
					FaultHandling:  RestartBusiness}}}
		faultDurationMap = make(map[string]*FaultDurationCache, common.MaxErrorCodeCount)
		loadFaultDurationCustomization(faultDurationCustomization)
		convey.So(faultDurationMap, convey.ShouldResemble, expectVal)
	})
}

// TestLoadFaultDurationCustomizationCase2 for test loadFaultDurationCustomizationCase1
func TestLoadFaultDurationCustomizationCase2(t *testing.T) {
	convey.Convey("test loadFaultDurationCustomizationCase1 success case2", t, func() {
		faultCode1 := "81078603"
		faultCode2 := "80E0180F"
		faultDurationCustomization := []FaultDurationCustomization{{
			EventId: []string{faultCode1}, FaultDuration: FaultDuration{
				FaultTimeout:   -1,
				RecoverTimeout: 60,
				FaultHandling:  PreSeparateNPU,
			}}, {
			EventId: []string{faultCode2}, FaultDuration: FaultDuration{
				FaultTimeout:   120,
				RecoverTimeout: 0,
				FaultHandling:  RestartBusiness}}}
		expectVal := map[string]*FaultDurationCache{
			strings.ToLower(faultCode2): {Duration: make(map[int32]FaultDurationData, common.MaxErrorCodeCount),
				FaultDuration: FaultDuration{
					FaultTimeout:   120,
					RecoverTimeout: 0,
					FaultHandling:  RestartBusiness}}}
		faultDurationMap = map[string]*FaultDurationCache{
			strings.ToLower(faultCode1): {Duration: make(map[int32]FaultDurationData, common.MaxErrorCodeCount),
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU}},
			strings.ToLower(faultCode2): {Duration: make(map[int32]FaultDurationData, common.MaxErrorCodeCount),
				FaultDuration: FaultDuration{
					FaultTimeout:   120,
					RecoverTimeout: 0,
					FaultHandling:  RestartBusiness}}}
		loadFaultDurationCustomization(faultDurationCustomization)
		convey.So(faultDurationMap, convey.ShouldResemble, expectVal)
	})
}

// TestGetMostSeriousFaultType for test getMostSeriousFaultType
func TestGetMostSeriousFaultType(t *testing.T) {
	convey.Convey("test getMostSeriousFaultType success case1", t, func() {
		fautTypes := []string{NotHandleFault, ManuallySeparateNPU}
		convey.So(getMostSeriousFaultType(fautTypes), convey.ShouldEqual, ManuallySeparateNPU)
	})

	convey.Convey("test getMostSeriousFaultType success case2", t, func() {
		fautTypes := []string{NotHandleFault, SeparateNPU}
		convey.So(getMostSeriousFaultType(fautTypes), convey.ShouldEqual, SeparateNPU)
	})

	convey.Convey("test getMostSeriousFaultType success case3", t, func() {
		fautTypes := []string{NotHandleFault, PreSeparateNPU}
		convey.So(getMostSeriousFaultType(fautTypes), convey.ShouldEqual, PreSeparateNPU)
	})

	convey.Convey("test getMostSeriousFaultType success case4", t, func() {
		fautTypes := []string{NotHandleFault, RestartNPU}
		convey.So(getMostSeriousFaultType(fautTypes), convey.ShouldEqual, RestartNPU)
	})

	convey.Convey("test getMostSeriousFaultType success case5", t, func() {
		fautTypes := []string{NotHandleFault, FreeRestartNPU}
		convey.So(getMostSeriousFaultType(fautTypes), convey.ShouldEqual, FreeRestartNPU)
	})

	convey.Convey("test getMostSeriousFaultType success case6", t, func() {
		fautTypes := []string{NotHandleFault, RestartBusiness}
		convey.So(getMostSeriousFaultType(fautTypes), convey.ShouldEqual, RestartBusiness)
	})

	convey.Convey("test getMostSeriousFaultType success case7", t, func() {
		fautTypes := []string{NotHandleFault, NotHandleFault}
		convey.So(getMostSeriousFaultType(fautTypes), convey.ShouldEqual, NotHandleFault)
	})

	convey.Convey("test getMostSeriousFaultType success case8", t, func() {
		fautTypes := make([]string, 0)
		convey.So(getMostSeriousFaultType(fautTypes), convey.ShouldEqual, NormalNPU)
	})
}

// TestGetFaultTypeFromFaultFrequency for test GetFaultTypeFromFaultFrequency
func TestGetFaultTypeFromFaultFrequency(t *testing.T) {
	convey.Convey("test GetFaultTypeFromFaultFrequency success case1", t, func() {
		logicId := int32(0)
		faultCode := "80E18005"
		timeDiff := int64(2)
		faultFrequencyMap = map[string]*FaultFrequencyCache{
			strings.ToLower(faultCode): {
				Frequency: map[int32][]int64{
					logicId: {time.Now().Unix() - timeDiff},
				},
				FaultFrequency: FaultFrequency{
					TimeWindow:    86400,
					Times:         3,
					FaultHandling: ManuallySeparateNPU,
				},
			},
		}
		convey.So(GetFaultTypeFromFaultFrequency(logicId, NetworkFaultMode), convey.ShouldEqual, NormalNPU)
	})

	convey.Convey("test GetFaultTypeFromFaultFrequency success case2", t, func() {
		logicId := int32(0)
		faultCode := "80E18005"
		firstTime, secondTime, thirdTime := int64(10), int64(8), int64(2)
		faultFrequencyMap = map[string]*FaultFrequencyCache{
			strings.ToLower(faultCode): {
				Frequency: map[int32][]int64{
					logicId: {time.Now().Unix() - firstTime, time.Now().Unix() - secondTime,
						time.Now().Unix() - thirdTime},
				},
				FaultFrequency: FaultFrequency{
					TimeWindow:    86400,
					Times:         3,
					FaultHandling: ManuallySeparateNPU,
				},
			},
		}
		manuallySeparateNpuMap = make(map[int32]ManuallyFaultInfo, GeneralMapSize)
		recoverFaultFrequencyMap = make(map[int32]string, GeneralMapSize)
		convey.So(GetFaultTypeFromFaultFrequency(logicId, ChipFaultMode), convey.ShouldEqual, ManuallySeparateNPU)
		convey.So(manuallySeparateNpuMap[logicId].FirstHandle, convey.ShouldEqual, true)
		convey.So(recoverFaultFrequencyMap[logicId], convey.ShouldEqual, strings.ToLower("80E18005"))
	})
}

// TestGetFaultTypeFromFaultDurationPart1 for test GetFaultTypeFromFaultDuration
func TestGetFaultTypeFromFaultDurationPart1(t *testing.T) {
	convey.Convey("test GetFaultTypeFromFaultDuration success part1", t, func() {
		logicId := int32(0)
		faultCode := "80E0180F"
		faultDurationTime := int64(121)
		faultDurationMap = map[string]*FaultDurationCache{
			strings.ToLower(faultCode): {
				Duration: map[int32]FaultDurationData{
					logicId: {
						TimeoutStatus:            true,
						FaultEventQueue:          []common.DevFaultInfo{},
						FaultDurationTime:        faultDurationTime * SecondMagnification,
						FaultRecoverDurationTime: 0,
					},
				},
				FaultDuration: FaultDuration{
					FaultTimeout:   120,
					RecoverTimeout: 0,
					FaultHandling:  RestartBusiness,
				},
			},
		}
		convey.So(GetFaultTypeFromFaultDuration(logicId, ChipFaultMode), convey.ShouldEqual, RestartBusiness)
	})
}

// TestGetFaultTypeFromFaultDurationPart2 for test GetFaultTypeFromFaultDuration
func TestGetFaultTypeFromFaultDurationPart2(t *testing.T) {
	convey.Convey("test GetFaultTypeFromFaultDuration success part2", t, func() {
		logicId := int32(0)
		faultCode := "81078603"
		faultDurationTime := int64(25)
		faultDurationMap = map[string]*FaultDurationCache{
			strings.ToLower(faultCode): {
				Duration: map[int32]FaultDurationData{
					logicId: {
						TimeoutStatus:            false,
						FaultEventQueue:          []common.DevFaultInfo{},
						FaultDurationTime:        faultDurationTime * SecondMagnification,
						FaultRecoverDurationTime: 0,
					},
				},
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU,
				},
			},
		}
		convey.So(GetFaultTypeFromFaultDuration(logicId, NetworkFaultMode), convey.ShouldEqual, NormalNPU)
	})
}

// TestGetFaultType for test GetFaultType
func TestGetFaultType(t *testing.T) {
	convey.Convey("test GetFaultType success", t, func() {
		logicId := int32(0)
		faultCode := "80E0180F"
		faultCodes := []int64{0x80E0180F, 0x80C98002}
		firstTime, secondTime, thirdTime := int64(10), int64(8), int64(2)
		faultFrequencyMap = map[string]*FaultFrequencyCache{
			strings.ToLower(faultCode): {
				Frequency: map[int32][]int64{
					logicId: {time.Now().Unix() - firstTime, time.Now().Unix() - secondTime,
						time.Now().Unix() - thirdTime},
				},
				FaultFrequency: FaultFrequency{
					TimeWindow:    86400,
					Times:         3,
					FaultHandling: ManuallySeparateNPU,
				},
			},
		}
		faultDurationTime := int64(121)
		faultDurationMap = map[string]*FaultDurationCache{
			strings.ToLower(faultCode): {
				Duration: map[int32]FaultDurationData{
					logicId: {
						TimeoutStatus:            true,
						FaultEventQueue:          []common.DevFaultInfo{},
						FaultDurationTime:        faultDurationTime * SecondMagnification,
						FaultRecoverDurationTime: 0,
					},
				},
				FaultDuration: FaultDuration{
					FaultTimeout:   120,
					RecoverTimeout: 0,
					FaultHandling:  RestartBusiness,
				},
			},
		}
		manuallySeparateNpuMap = make(map[int32]ManuallyFaultInfo, GeneralMapSize)
		recoverFaultFrequencyMap = make(map[int32]string, GeneralMapSize)
		convey.So(GetFaultType(faultCodes, logicId), convey.ShouldEqual, ManuallySeparateNPU)
		convey.So(manuallySeparateNpuMap[logicId].FirstHandle, convey.ShouldEqual, true)
		convey.So(recoverFaultFrequencyMap[logicId], convey.ShouldEqual, strings.ToLower(faultCode))
	})
}

// TestGetNetworkFaultType for test GetNetworkFaultType
func TestGetNetworkFaultType(t *testing.T) {
	convey.Convey("test GetNetworkFaultType success", t, func() {
		logicId := int32(0)
		faultCode := "81078603"
		faultCodes := []int64{0x81078603}
		faultDurationTime := int64(31)
		faultDurationMap = map[string]*FaultDurationCache{
			strings.ToLower(faultCode): {
				Duration: map[int32]FaultDurationData{
					logicId: {
						TimeoutStatus:            true,
						FaultEventQueue:          []common.DevFaultInfo{},
						FaultDurationTime:        faultDurationTime * SecondMagnification,
						FaultRecoverDurationTime: 0,
					},
				},
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU,
				},
			},
		}
		convey.So(GetNetworkFaultType(faultCodes, logicId), convey.ShouldEqual, PreSeparateNPU)
	})
}

// TestSetAlarmRaisedTime for test setAlarmRaisedTime
func TestSetAlarmRaisedTime(t *testing.T) {
	alarmRaisedTime := int64(123)
	convey.Convey("test setAlarmRaisedTime success case1", t, func() {
		device := NpuDevice{
			FaultCodes:      []int64{0x80C98002, 0x80C98003},
			AlarmRaisedTime: alarmRaisedTime,
		}
		setAlarmRaisedTime(&device)
		convey.So(device.AlarmRaisedTime, convey.ShouldEqual, alarmRaisedTime)
	})

	convey.Convey("test setAlarmRaisedTime success case2", t, func() {
		device := NpuDevice{
			FaultCodes:      []int64{},
			AlarmRaisedTime: alarmRaisedTime,
		}
		setAlarmRaisedTime(&device)
		convey.So(device.AlarmRaisedTime, convey.ShouldEqual, 0)
	})

	convey.Convey("test setAlarmRaisedTime success case3", t, func() {
		device := NpuDevice{
			FaultCodes:      []int64{0x80C98002},
			AlarmRaisedTime: 0,
		}
		setAlarmRaisedTime(&device)
		convey.So(device.AlarmRaisedTime, convey.ShouldBeGreaterThan, 0)
	})
}

// TestSetNetworkAlarmRaisedTime for test setNetworkAlarmRaisedTime
func TestSetNetworkAlarmRaisedTime(t *testing.T) {
	networkAlarmRaisedTime := int64(125)
	convey.Convey("test setNetworkAlarmRaisedTime success case1", t, func() {
		device := NpuDevice{
			NetworkFaultCodes:      []int64{0x81078603},
			NetworkAlarmRaisedTime: networkAlarmRaisedTime,
		}
		setNetworkAlarmRaisedTime(&device)
		convey.So(device.NetworkAlarmRaisedTime, convey.ShouldEqual, networkAlarmRaisedTime)
	})

	convey.Convey("test setNetworkAlarmRaisedTime success case2", t, func() {
		device := NpuDevice{
			NetworkFaultCodes:      []int64{},
			NetworkAlarmRaisedTime: networkAlarmRaisedTime,
		}
		setNetworkAlarmRaisedTime(&device)
		convey.So(device.NetworkAlarmRaisedTime, convey.ShouldEqual, 0)
	})

	convey.Convey("test setNetworkAlarmRaisedTime success case3", t, func() {
		device := NpuDevice{
			NetworkFaultCodes:      []int64{0x81078603},
			NetworkAlarmRaisedTime: 0,
		}
		setNetworkAlarmRaisedTime(&device)
		convey.So(device.NetworkAlarmRaisedTime, convey.ShouldBeGreaterThan, 0)
	})
}

// TestGetChangedDevFaultInfo for test GetChangedDevFaultInfo
func TestGetChangedDevFaultInfo(t *testing.T) {
	convey.Convey("test GetChangedDevFaultInfo success case1", t, func() {
		device := NpuDevice{LogicID: 0}
		oldErrCodes := make([]int64, 0)
		newErrCodes := []int64{0x81078603}
		devFaultInfos := GetChangedDevFaultInfo(&device, oldErrCodes, newErrCodes)
		convey.So(len(devFaultInfos), convey.ShouldEqual, 1)
		convey.So(devFaultInfos[0].Assertion, convey.ShouldEqual, common.FaultOccur)
		convey.So(devFaultInfos[0].EventID, convey.ShouldEqual, 0x81078603)
	})

	convey.Convey("test GetChangedDevFaultInfo success case2", t, func() {
		device := NpuDevice{LogicID: 0}
		oldErrCodes := []int64{0x81078603}
		newErrCodes := make([]int64, 0)
		devFaultInfos := GetChangedDevFaultInfo(&device, oldErrCodes, newErrCodes)
		convey.So(len(devFaultInfos), convey.ShouldEqual, 1)
		convey.So(devFaultInfos[0].Assertion, convey.ShouldEqual, common.FaultRecover)
		convey.So(devFaultInfos[0].EventID, convey.ShouldEqual, 0x81078603)
	})

	convey.Convey("test GetChangedDevFaultInfo success case3", t, func() {
		device := NpuDevice{LogicID: 0}
		oldErrCodes := []int64{0x81078603}
		newErrCodes := []int64{0x81078603}
		devFaultInfos := GetChangedDevFaultInfo(&device, oldErrCodes, newErrCodes)
		convey.So(len(devFaultInfos), convey.ShouldEqual, 0)
	})
}

// TestGetTimeoutFaultCodes for test GetTimeoutFaultLevelAndCodes
func TestGetTimeoutFaultCodes(t *testing.T) {
	convey.Convey("test GetTimeoutFaultCodes success", t, func() {
		logicID := int32(0)
		linkDownFaultCodeStr := strings.ToLower(strconv.FormatInt(LinkDownFaultCode, Hex))
		CardDropFaultCodeStr := strings.ToLower(strconv.FormatInt(CardDropFaultCode, Hex))
		ResetFinishFaultCodeStr := strings.ToLower(strconv.FormatInt(ResetFinishFaultCode, Hex))
		NetworkFaultCodes = sets.NewInt64(LinkDownFaultCode)
		faultDurationMap = map[string]*FaultDurationCache{
			CardDropFaultCodeStr: {
				Duration: map[int32]FaultDurationData{logicID: {TimeoutStatus: true}},
				FaultDuration: FaultDuration{
					FaultTimeout:   120,
					RecoverTimeout: 0,
					FaultHandling:  SeparateNPU,
				},
			},
			linkDownFaultCodeStr: {
				Duration: map[int32]FaultDurationData{logicID: {TimeoutStatus: false}},
				FaultDuration: FaultDuration{
					FaultTimeout:   30,
					RecoverTimeout: 60,
					FaultHandling:  PreSeparateNPU,
				},
			},
			ResetFinishFaultCodeStr: {
				Duration: map[int32]FaultDurationData{logicID: {TimeoutStatus: true}},
				FaultDuration: FaultDuration{
					FaultTimeout:   240,
					RecoverTimeout: 0,
					FaultHandling:  SeparateNPU,
				},
			},
		}
		expectedChipFaultCodesLen := 2
		expectedNetworkFaultCodes := make(map[int64]FaultTimeAndLevel)
		convey.So(len(GetTimeoutFaultLevelAndCodes(ChipFaultMode, logicID)), convey.ShouldResemble, expectedChipFaultCodesLen)
		convey.So(GetTimeoutFaultLevelAndCodes(NetworkFaultMode, logicID), convey.ShouldResemble, expectedNetworkFaultCodes)
	})
}

// TestGetFrequencyFaultCodes for test GetFrequencyFaultLevelAndCodes
func TestGetFrequencyFaultCodes(t *testing.T) {
	convey.Convey("test GetTimeoutFaultCodes success", t, func() {
		logicID := int32(0)
		faultCode1 := "80C98000"
		faultCode2 := "80E18005"
		firstTime, secondTime, thirdTime := int64(10), int64(8), int64(2)
		NetworkFaultCodes = sets.NewInt64(LinkDownFaultCode)
		faultFrequencyMap = map[string]*FaultFrequencyCache{
			strings.ToLower(faultCode1): {
				Frequency: map[int32][]int64{
					logicID: {time.Now().Unix() - firstTime},
				},
				FaultFrequency: FaultFrequency{
					TimeWindow:    86400,
					Times:         2,
					FaultHandling: ManuallySeparateNPU,
				},
			},
			strings.ToLower(faultCode2): {
				Frequency: map[int32][]int64{
					logicID: {time.Now().Unix() - firstTime, time.Now().Unix() - secondTime,
						time.Now().Unix() - thirdTime},
				},
				FaultFrequency: FaultFrequency{
					TimeWindow:    86400,
					Times:         3,
					FaultHandling: ManuallySeparateNPU,
				},
			},
			strings.ToLower(faultCode2): {
				Frequency: map[int32][]int64{
					logicID: {time.Now().Unix() - firstTime},
				},
				FaultFrequency: FaultFrequency{
					TimeWindow:    86400,
					Times:         0,
					FaultHandling: ManuallySeparateNPU,
				},
			},
		}
		expectedChipFaultCodesLen := 2
		expectedNetworkFaultCodes := make(map[int64]FaultTimeAndLevel)
		convey.So(len(GetFrequencyFaultLevelAndCodes(ChipFaultMode, logicID)), convey.ShouldResemble, expectedChipFaultCodesLen)
		convey.So(GetFrequencyFaultLevelAndCodes(NetworkFaultMode, logicID), convey.ShouldResemble, expectedNetworkFaultCodes)
	})
}

// TestLoadSwitchFaultCode  Test LoadSwitchFaultCode
func TestLoadSwitchFaultCode(t *testing.T) {
	convey.Convey("test LoadSwitchFaultCode", t, func() {
		switchFileInfo := SwitchFaultFileInfo{
			NotHandleFaultCodes: []string{generalFaultCode},
		}
		bytes, err := json.Marshal(switchFileInfo)
		convey.So(err, convey.ShouldBeNil)
		err = LoadSwitchFaultCode(bytes)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(NotHandleFaultCodes) > 0, convey.ShouldBeTrue)
		convey.So(NotHandleFaultCodes[firstFaultIdx] == generalFaultCode, convey.ShouldBeTrue)
	})
}

// TestHbmFaultManager  Test HbmFaultManager
func TestHbmFaultManager(t *testing.T) {
	var fakeAlarmRaisedTime int64 = 123456789
	convey.Convey("test new fault manager", t, func() {
		hbmFaultManager := NewHbmFaultManager()
		convey.So(hbmFaultManager, convey.ShouldNotBeNil)
	})
	convey.Convey("test update hbm occur time ", t, func() {
		hbmFaultManager := NewHbmFaultManager()
		faultInfo := common.DevFaultInfo{
			LogicID:         1,
			EventID:         HbmDoubleBitFaultCode,
			AlarmRaisedTime: fakeAlarmRaisedTime,
		}
		hbmFaultManager.updateHbmOccurTime(faultInfo)
		hbmOccurTime, ok := hbmFaultManager.HbmOccurTimeCache[1]
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(hbmOccurTime, convey.ShouldEqual, fakeAlarmRaisedTime)
	})
	convey.Convey("test fault event in que ", t, func() {
		hbmFaultManager := NewHbmFaultManager()
		faultInfo := common.DevFaultInfo{
			LogicID:         1,
			EventID:         AivBusFaultCode,
			AlarmRaisedTime: fakeAlarmRaisedTime,
		}
		hbmFaultManager.aicFaultEventInQue(faultInfo)
		faultInfoList, ok := hbmFaultManager.AicFaultEventQue[1]
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(faultInfoList), convey.ShouldEqual, 1)
		convey.So(faultInfoList[0].AlarmRaisedTime, convey.ShouldEqual, fakeAlarmRaisedTime)
	})
	convey.Convey("test fault event in que ", t, func() {
		hbmFaultManager := NewHbmFaultManager()
		faultInfo := common.DevFaultInfo{
			LogicID:         1,
			EventID:         AicBusFaultCode,
			AlarmRaisedTime: 100000000,
		}
		hbmFaultManager.aicFaultEventInQue(faultInfo)
		faultInfoList := hbmFaultManager.aicFaultEventOutQue(1)
		convey.So(len(faultInfoList), convey.ShouldEqual, 1)
		convey.So(faultInfoList[0].AlarmRaisedTime, convey.ShouldEqual, 100000000)
		convey.So(faultInfoList[0].EventID, convey.ShouldEqual, AicBusFaultCode)
	})
}

func TestUpdateDeviceFaultTimeMap(t *testing.T) {
	t.Run("Test_updateDeviceFaultTimeMap", func(t *testing.T) {
		npuDevice := &NpuDevice{
			FaultTimeMap: make(map[int64]int64),
		}
		eventId := int64(4326743278)
		faultInfo := common.DevFaultInfo{
			EventID:         eventId,
			AlarmRaisedTime: time.Now().UnixMilli(),
		}
		isAdd := true
		updateDeviceFaultTimeMap(npuDevice, faultInfo, isAdd)
		faultTime, found := npuDevice.FaultTimeMap[eventId]
		if !found {
			t.Errorf("cannot found fault time")
			return
		}
		if faultTime != faultInfo.AlarmRaisedTime {
			t.Errorf("cannot found fault error")
			return
		}

		isAdd = false
		updateDeviceFaultTimeMap(npuDevice, faultInfo, isAdd)
		_, found = npuDevice.FaultTimeMap[eventId]
		if found {
			t.Errorf("found fault time")
			return
		}
	})
}
