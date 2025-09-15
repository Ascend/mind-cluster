/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/api"
	"ascend-common/devmanager/common"
)

const (
	cardNum          = 2
	generalFaultCode = "[0x00f103b0,155649,na,NoneExist]"
	firstFaultIdx    = 0
)

// TestToString for test ToString
func TestToString(t *testing.T) {
	convey.Convey("test ToString", t, func() {
		convey.Convey("ToString success", func() {
			testVal1, testVal2 := "test1", "test2"
			testStr := sets.String{}
			testStr.Insert(testVal1, testVal2)
			convey.So(ToString(testStr, ","), convey.ShouldEqual,
				fmt.Sprintf("%s,%s", testVal1, testVal2))
		})
	})
}

// TestConvertDevListToSets for test ConvertDevListToSets
func TestConvertDevListToSets(t *testing.T) {
	convey.Convey("test ConvertDevListToSets", t, func() {
		convey.Convey("devices is empty", func() {
			ret := ConvertDevListToSets("", "")
			convey.So(ret.Len(), convey.ShouldEqual, 0)
		})
		convey.Convey("length of deviceInfo more then MaxDevicesNum", func() {
			devices := ""
			for i := 0; i <= MaxDevicesNum; i++ {
				devices += strconv.Itoa(i) + "."
			}
			ret := ConvertDevListToSets(devices, "")
			convey.So(ret.Len(), convey.ShouldEqual, 0)
		})
		convey.Convey("sepType is DotSepDev, ParseInt failed", func() {
			devices := "a.b.c"
			ret := ConvertDevListToSets(devices, DotSepDev)
			convey.So(ret.Len(), convey.ShouldEqual, 0)
		})
		convey.Convey("sepType is DotSepDev, ParseInt ok", func() {
			devices := "0.1.2"
			ret := ConvertDevListToSets(devices, DotSepDev)
			convey.So(ret.Len(), convey.ShouldEqual, len(strings.Split(devices, ".")))
		})
		convey.Convey("match Ascend910", func() {
			devices := "Ascend910-0.Ascend910-1.Ascend910-2"
			ret := ConvertDevListToSets(devices, DotSepDev)
			convey.So(ret.Len(), convey.ShouldEqual, 0)
			testDevices := "Ascend910-0,Ascend910-1"
			res := ConvertDevListToSets(testDevices, CommaSepDev)
			convey.So(res.Len(), convey.ShouldEqual, cardNum)
		})
		convey.Convey("not match Ascend910", func() {
			devices := "0.1.2"
			ret := ConvertDevListToSets(devices, "")
			convey.So(ret.Len(), convey.ShouldEqual, 0)
		})
	})
}

// TestIsVirtualDev for test IsVirtualDev
func TestIsVirtualDev(t *testing.T) {
	convey.Convey("test IsVirtualDev", t, func() {
		convey.Convey("virtual device", func() {
			ret := IsVirtualDev("Ascend910")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("physical device", func() {
			ret := IsVirtualDev("Ascend910-2c-100-0")
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

// TestGetVNPUSegmentInfo for testGetVNPUSegmentInfo
func TestGetVNPUSegmentInfo(t *testing.T) {
	deviceInfos := []string{"0", "vir02"}
	convey.Convey("test GetVNPUSegmentInfo", t, func() {
		convey.Convey("GetVNPUSegmentInfo success", func() {
			_, _, err := GetVNPUSegmentInfo(deviceInfos)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("device info is empty", func() {
			_, _, err := GetVNPUSegmentInfo(nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		deviceInfos = []string{"165", "vir02"}
		convey.Convey("GetVNPUSegmentInfo failed with upper limit id", func() {
			_, _, err := GetVNPUSegmentInfo(deviceInfos)
			convey.So(err, convey.ShouldNotBeNil)
		})
		deviceInfos = []string{"x", "vir02"}
		convey.Convey("GetVNPUSegmentInfo failed with invalid id", func() {
			_, _, err := GetVNPUSegmentInfo(deviceInfos)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestIsValidNumber for IsValidNumber
func TestIsValidNumber(t *testing.T) {
	convey.Convey("test IsValidNumber", t, func() {
		convey.Convey("IsValidNumber success", func() {
			testVal := "_"
			ret, err := IsValidNumber(testVal)
			convey.So(ret, convey.ShouldEqual, -1)
			convey.So(err, convey.ShouldBeFalse)
		})
	})
}

// TestGetAICore for GetAICore
func TestGetAICore(t *testing.T) {
	convey.Convey("test GetAICore", t, func() {
		convey.Convey("GetAICore success", func() {
			testVal := "0"
			ret, err := GetAICore(testVal)
			convey.So(ret, convey.ShouldEqual, 0)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetTemplateName2DeviceTypeMap for GetTemplateName2DeviceTypeMap
func TestGetTemplateName2DeviceTypeMap(t *testing.T) {
	convey.Convey("test GetTemplateName2DeviceTypeMap", t, func() {
		convey.Convey("GetTemplateName2DeviceTypeMap success", func() {
			convey.So(GetTemplateName2DeviceTypeMap(), convey.ShouldNotBeNil)
		})
	})
}

// TestFakeAiCoreDevice for testFakeAiCoreDevice
func TestFakeAiCoreDevice(t *testing.T) {
	dev := DavinCiDev{
		LogicID: 0,
		PhyID:   0,
	}
	aiCoreDevices := make([]*NpuDevice, 0)
	ParamOption.AiCoreCount = MinAICoreNum
	convey.Convey("test FakeAiCoreDevice", t, func() {
		convey.Convey("FakeAiCoreDevice success", func() {
			FakeAiCoreDevice(dev, &aiCoreDevices)
			convey.So(len(aiCoreDevices), convey.ShouldEqual, MinAICoreNum)
		})
	})
}

// TestCheckCardUsageMode for test CheckCardUsageMode
func TestCheckCardUsageMode(t *testing.T) {
	convey.Convey("test use 310P Mixed Insert and device is correct", t, func() {
		convey.Convey("virtual device", func() {
			ret := CheckCardUsageMode(true, []string{"Atlas 300V Pro", "Atlas 300V"})
			convey.So(ret, convey.ShouldBeNil)
		})
	})
	convey.Convey("test use 310P Mixed Insert and device is incorrect", t, func() {
		convey.Convey("virtual device", func() {
			ret := CheckCardUsageMode(true, []string{"11", "222"})
			convey.So(ret, convey.ShouldNotBeNil)
		})
	})
	convey.Convey("test not use 310P Mixed Insert and device is correct", t, func() {
		convey.Convey("virtual device", func() {
			ret := CheckCardUsageMode(false, []string{"111"})
			convey.So(ret, convey.ShouldBeNil)
		})
	})
	convey.Convey("test CheckCardUsageMode", t, func() {
		convey.Convey("do not get product type", func() {
			convey.So(CheckCardUsageMode(true, []string{}), convey.ShouldNotBeNil)
		})
	})
}

// TestGetSwitchFaultInfo test for convert fault code into struct
func TestGetSwitchFaultInfo(t *testing.T) {
	convey.Convey("test GetSwitchFaultInfo", t, func() {
		convey.Convey("when card type is not Ascend910A3, return empty result", func() {
			ParamOption.RealCardType = common.Ascend910
			convey.So(GetSwitchFaultInfo(), convey.ShouldResemble, SwitchFaultInfo{})
		})
		convey.Convey("when EnableSwitchFault is false, return empty result", func() {
			ParamOption.EnableSwitchFault = false
			convey.So(GetSwitchFaultInfo(), convey.ShouldResemble, SwitchFaultInfo{})
		})
		ParamOption.RealCardType = api.Ascend910A3
		ParamOption.EnableSwitchFault = true
		currentSwitchFault = []SwitchFaultEvent{}
		SwitchFaultLevelMap = map[string]int{}
		convey.Convey("test empty SwitchFaultLevelMap", func() {
			currentSwitchFault = append(currentSwitchFault, SwitchFaultEvent{AssembledFaultCode: generalFaultCode})
			fault := GetSwitchFaultInfo()
			convey.So(fault.FaultLevel == NotHandleFaultLevelStr, convey.ShouldBeTrue)
		})
		convey.Convey("test actually level", func() {
			currentSwitchFault = []SwitchFaultEvent{}
			fault := GetSwitchFaultInfo()
			convey.So(fault.FaultLevel == "", convey.ShouldBeTrue)

			currentSwitchFault = append(currentSwitchFault, SwitchFaultEvent{AssembledFaultCode: generalFaultCode})
			SwitchFaultLevelMap = map[string]int{generalFaultCode: NotHandleFaultLevel}
			switchFaultCodeLevelToCm = map[string]int{}
			mockFunc := gomonkey.ApplyFuncReturn(getSimpleSwitchFaultStr, "", errors.New("failed"))
			fault = GetSwitchFaultInfo()
			convey.So(fault.FaultLevel == NotHandleFaultLevelStr, convey.ShouldBeTrue)
			convey.So(len(fault.FaultTimeAndLevelMap) == 0, convey.ShouldBeTrue)
			mockFunc.Reset()

			SwitchFaultLevelMap = map[string]int{generalFaultCode: PreSeparateFaultLevel}
			switchFaultCodeLevelToCm = map[string]int{}
			fault = GetSwitchFaultInfo()
			convey.So(fault.FaultLevel == PreSeparateFaultLevelStr, convey.ShouldBeTrue)
			convey.So(len(fault.FaultTimeAndLevelMap) == 1, convey.ShouldBeTrue)

			switchFaultCodeLevelToCm = map[string]int{}
			SwitchFaultLevelMap = map[string]int{generalFaultCode: SeparateFaultLevel}
			fault = GetSwitchFaultInfo()
			convey.So(fault.FaultLevel == SeparateFaultLevelStr, convey.ShouldBeTrue)
			convey.So(len(fault.FaultTimeAndLevelMap) == 1, convey.ShouldBeTrue)

			switchFaultCodeLevelToCm = map[string]int{generalFaultCode: NotHandleFaultLevel}
			fault = GetSwitchFaultInfo()
			convey.So(fault.FaultLevel == NotHandleFaultLevelStr, convey.ShouldBeTrue)
			convey.So(len(fault.FaultTimeAndLevelMap) == 1, convey.ShouldBeTrue)
		})
	})
}

// TestUpdateSwitchFaultInfoAndFaultLevel for test UpdateSwitchFaultInfoAndFaultLevel
func TestUpdateSwitchFaultInfoAndFaultLevel(t *testing.T) {
	convey.Convey("test UpdateSwitchFaultInfoAndFaultLevel", t, func() {
		// 01-update switch fault success, switchFaultCodeLevelToCm and switchFault should be updated
		switchFault := SwitchFaultInfo{
			NodeStatus: nodeSubHealthy,
		}
		mockMap := gomonkey.ApplyGlobalVar(&switchFaultCodeLevelToCm, map[string]int{generalFaultCode: SeparateFaultLevel})
		defer mockMap.Reset()
		UpdateSwitchFaultInfoAndFaultLevel(&switchFault)
		convey.So(switchFault.NodeStatus == nodeHealthy, convey.ShouldBeTrue)
		convey.So(switchFaultCodeLevelToCm[generalFaultCode], convey.ShouldEqual, NotHandleFaultLevel)
	})
}

// TestDeepEqualSwitchFaultInfo for test DeepEqualSwitchFaultInfo
func TestDeepEqualSwitchFaultInfo(t *testing.T) {
	convey.Convey("test DeepEqualSwitchFaultInfo", t, func() {
		convey.Convey("when faultCode length different, result return false", func() {
			res := DeepEqualSwitchFaultInfo(SwitchFaultInfo{FaultCode: []string{"1", "2"}},
				SwitchFaultInfo{FaultCode: []string{"1"}})
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("when faultCode elements mismatch, result return false", func() {
			res := DeepEqualSwitchFaultInfo(SwitchFaultInfo{FaultCode: []string{"1", "2"}},
				SwitchFaultInfo{FaultCode: []string{"1", "3"}})
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("when faultLevel different, result return false", func() {
			res := DeepEqualSwitchFaultInfo(SwitchFaultInfo{FaultLevel: NotHandleFaultLevelStr},
				SwitchFaultInfo{FaultLevel: PreSeparateFaultLevelStr})
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("when nodeStatus different, result return false", func() {
			res := DeepEqualSwitchFaultInfo(SwitchFaultInfo{NodeStatus: nodeHealthy},
				SwitchFaultInfo{NodeStatus: nodeUnHealthy})
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("when all fields equal except updateTime, result return true", func() {
			this := SwitchFaultInfo{
				FaultCode:  []string{"1", "2"},
				FaultLevel: NotHandleFaultLevelStr,
				NodeStatus: nodeHealthy,
				UpdateTime: 0}
			other := SwitchFaultInfo{
				FaultCode:  []string{"1", "2"},
				FaultLevel: NotHandleFaultLevelStr,
				NodeStatus: nodeHealthy,
				UpdateTime: 1}
			res := DeepEqualSwitchFaultInfo(this, other)
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}

// TestSetSwitchFaultCode for test SetSwitchFaultCode
func TestSetSwitchFaultCode(t *testing.T) {
	convey.Convey("test SetSwitchFaultCode", t, func() {
		patch := gomonkey.ApplyGlobalVar(&currentSwitchFault, make([]SwitchFaultEvent, 0, GeneralMapSize))
		defer patch.Reset()
		SetSwitchFaultCode(nil)
		convey.So(GetSwitchFaultCode(), convey.ShouldBeNil)
	})
}

// TestCheckL1toL2PlaneDown for checkL1toL2PlaneDown
func TestCheckL1toL2PlaneDown(t *testing.T) {
	convey.Convey("test checkL1toL2PlaneDown", t, func() {
		patch := gomonkey.ApplyGlobalVar(&currentSwitchFault, make([]SwitchFaultEvent, 0, GeneralMapSize))
		defer patch.Reset()
		event := SwitchFaultEvent{
			PeerPortDevice: PeerDeviceL2Port,
			SwitchChipId:   0,
			SwitchPortId:   0,
			Assertion:      1,
		}
		faults := make([]SwitchFaultEvent, 0, L1ToL2PlanePortNum)
		for i := 0; i < L1ToL2PlanePortNum; i++ {
			event.SwitchPortId = uint(i)
			faults = append(faults, event)
		}
		SetSwitchFaultCode(faults)
		convey.So(checkL1toL2PlaneDown(), convey.ShouldBeTrue)
	})
}

// TestConvertToSwitchLevelStr for test convertToSwitchLevelStr
func TestConvertToSwitchLevelStr(t *testing.T) {
	testCases := map[int]string{
		NotHandleFaultLevel:   NotHandleFaultLevelStr,
		ResetErrorLevel:       RestartRequestFaultLevelStr,
		PreSeparateFaultLevel: PreSeparateFaultLevelStr,
		SeparateFaultLevel:    SeparateFaultLevelStr,
		-1:                    NotHandleFaultLevelStr,
	}

	for input, expected := range testCases {
		t.Run("convertToSwitchLevelStr Level_"+string(rune(input)), func(t *testing.T) {
			result := convertToSwitchLevelStr(input)
			if result != expected {
				t.Errorf("For input %d: got %s, want %s", input, result, expected)
			}
		})
	}
}
