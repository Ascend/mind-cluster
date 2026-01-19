/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package faultcontrol for ipmi fault handling test
package faultcontrol

import (
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"nodeD/pkg/common"
)

var nodeController *NodeController

func TestNodeController(t *testing.T) {
	nodeController = NewNodeController()
	convey.Convey("test NodeController method 'UpdateConfig'", t, testNodeCtlUpdateConfig)
	convey.Convey("test NodeController method 'updateFaultDevInfo'", t, testUpdateFaultDevInfo)
	convey.Convey("test NodeController method 'filterNotSupportFaultCodes'", t, testFilterNotSupportFaultCodes)
	convey.Convey("test NodeController method 'updateFaultLevelMap'", t, testNodeCtlUpdateFaultLevelMap)
	convey.Convey("test NodeController method 'getFaultLevel'", t, testNodeCtlGetFaultLevel)
	convey.Convey("test NodeController method 'getNodeStatus'", t, testNodeCtlGetNodeStatus)
}

func TestName(t *testing.T) {
	nodeController = NewNodeController()
	convey.Convey("test Name", t, func() {
		convey.So(nodeController.Name(), convey.ShouldEqual, common.PluginControlFault)
	})
}

func TestUpdateConfig(t *testing.T) {
	nodeController = NewNodeController()
	convey.Convey("test UpdateConfig", t, func() {
		convey.So(nodeController.UpdateConfig(nil), convey.ShouldEqual, nil)
	})
}

func TestControl(t *testing.T) {
	nodeController = NewNodeController()
	convey.Convey("test Control", t, func() {
		patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(nodeController), "updateFault",
			func(_ *NodeController, faultDevInfo *common.FaultDevInfo) *common.FaultDevInfo {
				return nil
			}).ApplyMethod(reflect.TypeOf(nodeController), "UpdateConfig",
			func(_ *NodeController, faultConfig *common.FaultConfig) *common.FaultConfig {
				return nil
			})
		defer patch.Reset()
		info := &common.FaultAndConfigInfo{
			FaultDevInfo: &common.FaultDevInfo{},
			FaultConfig:  &common.FaultConfig{},
		}
		nodeController.Control(info)
		convey.So(info.FaultDevInfo, convey.ShouldBeNil)
		convey.So(info.FaultConfig, convey.ShouldBeNil)
	})
}

func testNodeCtlUpdateConfig() {
	if nodeController == nil {
		panic("nodeController is nil")
	}

	testFaultConfig := &common.FaultConfig{
		FaultTypeCode: faultTypeCode,
	}
	newFaultConfig := nodeController.UpdateConfig(testFaultConfig)
	convey.So(newFaultConfig, convey.ShouldEqual, testFaultConfig)

	newFaultConfig = nodeController.UpdateConfig(testWrongFaultConfig)
	convey.So(newFaultConfig, convey.ShouldBeNil)
}

func testUpdateFaultDevInfo() {
	if nodeController == nil {
		panic("nodeController is nil")
	}
	resetFaultDevInfo()
	nodeController.updateFault(nodeController.faultManager.GetFaultDevInfo())
	convey.So(nodeController.faultManager.GetNodeStatus(), convey.ShouldEqual, common.PreSeparate)
}

func testFilterNotSupportFaultCodes() {
	if nodeController == nil {
		panic("nodeController is nil")
	}
	newFaultCodes := nodeController.filterNotSupportFaultCodes([]string{faultCode1, wrongFaultCode})
	convey.So(newFaultCodes, convey.ShouldResemble, []string{faultCode1})
}

func testNodeCtlUpdateFaultLevelMap() {
	if nodeController == nil {
		panic("nodeController is nil")
	}

	defer resetFaultLevelMap()
	convey.Convey("test method updateFaultLevelMap failed, faultTypeCode is nil", func() {
		err := nodeController.updateFaultLevelMap(nil)
		convey.So(err, convey.ShouldResemble, errors.New("fault type code is nil"))
	})
	convey.Convey("test method updateFaultLevelMap success", func() {
		err := nodeController.updateFaultLevelMap(faultTypeCode)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test method updateFaultLevelMap failed, notHandleFaultCodes are conflict", func() {
		faultTypeCode := &common.FaultTypeCode{
			NotHandleFaultCodes:   []string{faultCode1, faultCode1},
			PreSeparateFaultCodes: nil,
			SeparateFaultCodes:    nil,
		}
		err := nodeController.updateFaultLevelMap(faultTypeCode)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("test method updateFaultLevelMap failed, preSeparateFaultCodes are conflict", func() {
		faultTypeCode := &common.FaultTypeCode{
			NotHandleFaultCodes:   nil,
			PreSeparateFaultCodes: []string{faultCode1, faultCode1},
			SeparateFaultCodes:    nil,
		}
		err := nodeController.updateFaultLevelMap(faultTypeCode)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("test method updateFaultLevelMap failed, separateFaultCodes are conflict", func() {
		faultTypeCode := &common.FaultTypeCode{
			NotHandleFaultCodes:   nil,
			PreSeparateFaultCodes: nil,
			SeparateFaultCodes:    []string{faultCode1, faultCode1},
		}
		err := nodeController.updateFaultLevelMap(faultTypeCode)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testNodeCtlGetFaultLevel() {
	if nodeController == nil {
		panic("nodeController is nil")
	}

	defer resetFaultLevelMap()
	convey.Convey("test method getFaultLevel, max level is 'NotHandleFaultLevel'", func() {
		nodeController.faultLevelMap = map[string]int{
			faultCode1: common.NotHandleFaultLevel,
		}
		faultLevelStr, faultLevelInt := nodeController.getFaultLevel([]string{faultCode1})
		convey.So(faultLevelStr, convey.ShouldEqual, common.NotHandleFault)
		convey.So(faultLevelInt, convey.ShouldEqual, common.NotHandleFaultLevel)
	})

	convey.Convey("test method getFaultLevel, max level is 'PreSeparateFaultLevel'", func() {
		nodeController.faultLevelMap = map[string]int{
			faultCode1: common.NotHandleFaultLevel,
			faultCode2: common.PreSeparateFaultLevel,
		}
		faultLevelStr, faultLevelInt := nodeController.getFaultLevel([]string{faultCode1, faultCode2})
		convey.So(faultLevelStr, convey.ShouldEqual, common.PreSeparateFault)
		convey.So(faultLevelInt, convey.ShouldEqual, common.PreSeparateFaultLevel)
	})

	convey.Convey("test method getFaultLevel, max level is 'SeparateFaultLevel'", func() {
		nodeController.faultLevelMap = map[string]int{
			faultCode1: common.NotHandleFaultLevel,
			faultCode2: common.PreSeparateFaultLevel,
			faultCode3: common.SeparateFaultLevel,
		}
		faultLevelStr, faultLevelInt := nodeController.getFaultLevel([]string{faultCode1, faultCode2, faultCode3})
		convey.So(faultLevelStr, convey.ShouldEqual, common.SeparateFault)
		convey.So(faultLevelInt, convey.ShouldEqual, common.SeparateFaultLevel)
	})
}

func testNodeCtlGetNodeStatus() {
	if nodeController == nil {
		panic("nodeController is nil")
	}

	const errNodeLevel = 5
	testCases := []struct {
		input  int64
		expRes string
	}{
		{common.NodeUnHealthyLevel, common.NodeUnHealthy},
		{common.NodeSubHealthyLevel, common.PreSeparate},
		{common.NodeHealthyLevel, common.NodeHealthy},
		{errNodeLevel, ""},
	}
	for _, testCase := range testCases {
		res := nodeController.getNodeStatus(testCase.input)
		convey.So(res, convey.ShouldResemble, testCase.expRes)
	}
}
