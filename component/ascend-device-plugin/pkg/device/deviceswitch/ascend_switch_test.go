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
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/common-utils/hwlog"
	devmanagercommon "ascend-common/devmanager/common"
)

const (
	linkDownAlarmID    = 0x08520003
	commonFaultAlarmID = 0x00f10509
	laneFaultID        = 132333
	halfLaneFaultID    = 132332
	invalidPortFaultID = 155912
	enginAlarmID       = 0x00f103b6
	enginFaultID       = 155909
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// TestUpdateSwitchFaultLevel for test UpdateSwitchFaultLevel
func TestUpdateSwitchFaultLevel(t *testing.T) {
	convey.Convey("test UpdateSwitchFaultLevel", t, func() {
		// 01-update common.SwitchFaultLevelMap success
		notHandleCode := "[0x00f1ff09,155913,cpu,na]"
		restartRequestCode := "[0x00f103b0,155649,na,NoneExist]"
		preSeparateCode := "[0x00f103b0,155907,na,na]"
		separateCode := "[0x00f103b0,155649,na,na]"
		mockNotHandleCodes := gomonkey.ApplyGlobalVar(&common.NotHandleFaultCodes, []string{notHandleCode})
		defer mockNotHandleCodes.Reset()
		mockRestartRequestCodes := gomonkey.ApplyGlobalVar(&common.RestartRequestFaultCodes, []string{restartRequestCode})
		defer mockRestartRequestCodes.Reset()
		mockPreseparateCodes := gomonkey.ApplyGlobalVar(&common.PreSeparateFaultCodes, []string{preSeparateCode})
		defer mockPreseparateCodes.Reset()
		mockSeparateCodes := gomonkey.ApplyGlobalVar(&common.SeparateFaultCodes, []string{separateCode})
		defer mockSeparateCodes.Reset()
		UpdateSwitchFaultLevel()
		convey.So(common.SwitchFaultLevelMap[notHandleCode], convey.ShouldEqual, common.NotHandleFaultLevel)
		convey.So(common.SwitchFaultLevelMap[restartRequestCode], convey.ShouldEqual, common.RestartRequestFaultLevel)
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
			EventType:      linkDownAlarmID,
			SubType:        invalidNum,
			PeerPortDevice: common.PeerDeviceChipOrCpuPort,
		}
		convey.Convey("01-EventType is SwitchPortFault, SubType is PortDown, "+
			"AssembledFaultCode should be [0x08520003,na,cpu,na]", func() {
			setExtraFaultInfo(event)
			convey.So(event.AssembledFaultCode, convey.ShouldEqual, "[0x08520003,na,cpu,na]")
		})
		convey.Convey("02-EventType is SwitchPortFault, SubType is PortLaneReduceQuarter, "+
			"AssembledFaultCode should be [0x00f10509,132333,npu,na]", func() {
			event.EventType = commonFaultAlarmID
			event.SubType = laneFaultID
			event.PeerPortDevice = common.PeerDeviceNpuPort
			setExtraFaultInfo(event)
			convey.So(event.AssembledFaultCode, convey.ShouldEqual, "[0x00f10509,132333,npu,na]")
		})
		convey.Convey("03-EventType is SwitchPortFault,SubType is PortLaneReduceHalf, "+
			"AssembledFaultCode should be [0x00f10509,132332,L2,na]", func() {
			event.EventType = commonFaultAlarmID
			event.SubType = halfLaneFaultID
			event.PeerPortDevice = common.PeerDeviceL2Port
			setExtraFaultInfo(event)
			convey.So(event.AssembledFaultCode, convey.ShouldEqual, "[0x00f10509,132332,L2,na]")
		})
		convey.Convey("04-EventType is SwitchPortFault,SubType is other type, "+
			"AssembledFaultCode should be [0x00f1ff09,155912,na,na]", func() {
			event.EventType = commonFaultAlarmID
			event.SubType = invalidPortFaultID
			event.PeerPortDevice = common.PeerDeviceL2Port
			event.SwitchPortId = invalidNum
			setExtraFaultInfo(event)
			convey.So(event.AssembledFaultCode, convey.ShouldEqual, "[0x00f10509,155912,na,na]")
		})
		convey.Convey("05-EventType is 6, AssembledFaultCode should be [0x00f103b6,155909,na,na]", func() {
			event.EventType = enginAlarmID
			event.SubType = enginFaultID
			event.SwitchPortId = invalidNum
			setExtraFaultInfo(event)
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

func TestUpdateSwitchFaultCode(t *testing.T) {
	convey.Convey("test updateSwitchFaultCode", t, func() {
		var mockCurrentSwitchFault = make([]common.SwitchFaultEvent, 0, common.GeneralMapSize)
		convey.Convey("01-when switch fault cache is not init and get empty fault, cache is not update", func() {
			patches := gomonkey.ApplyFuncReturn(GetSwitchFaults, []common.SwitchFaultEvent{}, nil).
				ApplyFuncReturn(common.GetSwitchFaultCode, []common.SwitchFaultEvent{}).
				ApplyFunc(common.SetSwitchFaultCode, func(newFaults []common.SwitchFaultEvent) {
					mockCurrentSwitchFault = newFaults
				})
			defer patches.Reset()
			updateSwitchFaultCode(false)
			convey.So(len(mockCurrentSwitchFault), convey.ShouldEqual, 0)
		})
		convey.Convey("02-switch fault cache is init and get faults, cache is update success", func() {
			mockCurrentSwitchFault = []common.SwitchFaultEvent{{}, {}}
			patches := gomonkey.ApplyFuncReturn(GetSwitchFaults, []common.SwitchFaultEvent{{}}, nil).
				ApplyFuncReturn(common.GetSwitchFaultCode, mockCurrentSwitchFault).
				ApplyFunc(common.SetSwitchFaultCode, func(newFaults []common.SwitchFaultEvent) {
					mockCurrentSwitchFault = newFaults
				})
			defer patches.Reset()
			updateSwitchFaultCode(true)
			convey.So(len(mockCurrentSwitchFault), convey.ShouldEqual, 1)
		})
		convey.Convey("03-get switch faults failed, cache is not update", func() {
			mockCurrentSwitchFault = []common.SwitchFaultEvent{{}}
			patches := gomonkey.ApplyFuncReturn(GetSwitchFaults, nil,
				fmt.Errorf("get switch faults failed")).
				ApplyFuncReturn(common.GetSwitchFaultCode, mockCurrentSwitchFault).
				ApplyFunc(common.SetSwitchFaultCode, func(newFaults []common.SwitchFaultEvent) {
					mockCurrentSwitchFault = newFaults
				})
			defer patches.Reset()
			updateSwitchFaultCode(true)
			convey.So(len(mockCurrentSwitchFault), convey.ShouldEqual, 1)
		})
	})
}
