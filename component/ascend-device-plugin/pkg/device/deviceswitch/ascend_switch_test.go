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

// Package deviceswitch functions of getting switch faults code
package deviceswitch

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/common"
	devmanagercommon "ascend-common/devmanager/common"
)

// TestUpdateSwitchFaultLevel for test UpdateSwitchFaultLevel
func TestUpdateSwitchFaultLevel(t *testing.T) {
	convey.Convey("test UpdateSwitchFaultLevel", t, func() {
		// 01-update common.SwitchFaultLevelMap success
		notHandleCode := "[0x00f1ff09,155913,cpu,na]"
		preSeparateCode := "[0x00f103b0,155907,na,na]"
		separateCode := "[0x00f103b0,155649,na,na]"
		mockSwitchFaultLevelMap := gomonkey.ApplyGlobalVar(&common.SwitchFaultLevelMap, map[string]int{})
		defer mockSwitchFaultLevelMap.Reset()
		mockNotHandleCodes := gomonkey.ApplyGlobalVar(&common.NotHandleFaultCodes, []string{notHandleCode})
		defer mockNotHandleCodes.Reset()
		mockPreseparateCodes := gomonkey.ApplyGlobalVar(&common.PreSeparateFaultCodes, []string{preSeparateCode})
		defer mockPreseparateCodes.Reset()
		mockSeparateCodes := gomonkey.ApplyGlobalVar(&common.SeparateFaultCodes, []string{separateCode})
		defer mockSeparateCodes.Reset()
		UpdateSwitchFaultLevel()
		convey.So(common.SwitchFaultLevelMap[notHandleCode], convey.ShouldEqual, common.NotHandleFaultLevel)
		convey.So(common.SwitchFaultLevelMap[preSeparateCode], convey.ShouldEqual, common.PreSeparateFaultLevel)
		convey.So(common.SwitchFaultLevelMap[separateCode], convey.ShouldEqual, common.SeparateFaultLevel)
	})
}

// TestNewSwitchDevManager for test NewSwitchDevManager
func TestNewSwitchDevManager(t *testing.T) {
	convey.Convey("test NewSwitchDevManager", t, func() {
		// 01-create manager success, should return object
		convey.So(NewSwitchDevManager(), convey.ShouldNotBeNil)
	})
}

// TestSetExtraFaultInfo for test SetExtraFaultInfo
func TestSetExtraFaultInfo(t *testing.T) {
	convey.Convey("test setExtraFaultInfo", t, func() {
		event := &common.SwitchFaultEvent{
			EventType:      common.EventTypeOfSwitchPortFault,
			SubType:        common.SubTypeOfPortDown,
			PeerPortDevice: common.PeerDeviceChipOrCpuPort,
		}
		convey.Convey("01-EventType is SwitchPortFault, SubType is PortDown, " +
			"AssembledFaultCode should be [0x08520003,,cpu,na]", func() {
			convey.So(setExtraFaultInfo(event), convey.ShouldBeNil)
			convey.So(event.AssembledFaultCode, convey.ShouldEqual, "[0x08520003,,cpu,na]")
		})
		convey.Convey("02-EventType is SwitchPortFault, SubType is PortLaneReduceQuarter, " +
			"AssembledFaultCode should be [0x00f10509,132333,npu,na]", func() {
			event.SubType = common.SubTypeOfPortLaneReduceQuarter
			event.PeerPortDevice = common.PeerDeviceNpuPort
			convey.So(setExtraFaultInfo(event), convey.ShouldBeNil)
			convey.So(event.AssembledFaultCode, convey.ShouldEqual, "[0x00f10509,132333,npu,na]")
		})
		convey.Convey("03-EventType is SwitchPortFault,SubType is PortLaneReduceHalf, " +
			"AssembledFaultCode should be [0x00f10509,132332,L2,na]", func() {
			event.SubType = common.SubTypeOfPortLaneReduceHalf
			event.PeerPortDevice = common.PeerDeviceL2Port
			convey.So(setExtraFaultInfo(event), convey.ShouldBeNil)
			convey.So(event.AssembledFaultCode, convey.ShouldEqual, "[0x00f10509,132332,L2,na]")
		})
		convey.Convey("04-EventType is SwitchPortFault,SubType is other type, " +
			"AssembledFaultCode should be [0x00f1ff09,155912,na,na]", func() {
			event.SubType = 0
			event.PeerPortDevice = 999999
			convey.So(setExtraFaultInfo(event), convey.ShouldBeNil)
			convey.So(event.AssembledFaultCode, convey.ShouldEqual, "[0x00f1ff09,155912,na,na]")
		})
		convey.Convey("05-EventType is 6, AssembledFaultCode should be [0x00f103b6,155909,na,na]", func() {
			event.EventType = 6
			convey.So(setExtraFaultInfo(event), convey.ShouldBeNil)
			convey.So(event.AssembledFaultCode, convey.ShouldEqual, "[0x00f103b6,155909,na,na]")
		})
	})
}

// TestIsFaultRecoveredEvent for test isFaultRecoveredEvent
func TestIsFaultRecoveredEvent(t *testing.T) {
	convey.Convey("test isFaultRecoveredEvent", t, func() {
		// 01-have fault occur and recover event, should return true
		faultEvent := common.SwitchFaultEvent{Assertion: uint(devmanagercommon.FaultOccur)}
		recoverFaultEvent := common.SwitchFaultEvent{Assertion: uint(devmanagercommon.FaultRecover)}
		convey.So(isFaultRecoveredEvent(faultEvent, recoverFaultEvent), convey.ShouldBeTrue)
		// 02-not have recover event, should return false
		recoverFaultEvent.Assertion = uint(devmanagercommon.FaultOccur)
		convey.So(isFaultRecoveredEvent(faultEvent, recoverFaultEvent), convey.ShouldBeFalse)
	})
}
