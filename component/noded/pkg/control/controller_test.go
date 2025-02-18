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

// Package control for the node controller test
package control

import (
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/tools/clientcmd"

	"nodeD/pkg/common"
	"nodeD/pkg/reporter"
)

var nodeController *NodeController

func TestNodeController(t *testing.T) {
	nodeController = NewNodeController(testK8sClient)
	convey.Convey("test NodeController method 'SetNextFaultProcessor'", t, testNodeCtlSetNextFaultProcessor)
	convey.Convey("test NodeController method 'Init'", t, testNodeCtlInit)
	convey.Convey("test NodeController method 'Execute'", t, testNodeCtlExecute)
	convey.Convey("test NodeController method 'UpdateConfig'", t, testNodeCtlUpdateConfig)

	convey.Convey("test NodeController method 'updateFaultDevInfo'", t, testUpdateFaultDevInfo)
	convey.Convey("test NodeController method 'filterNotSupportFaultCodes'", t, testFilterNotSupportFaultCodes)
	convey.Convey("test NodeController method 'updateFaultLevelMap'", t, testNodeCtlUpdateFaultLevelMap)
	convey.Convey("test NodeController method 'getFaultLevel'", t, testNodeCtlGetFaultLevel)
	convey.Convey("test NodeController method 'getNodeStatus'", t, testNodeCtlGetNodeStatus)

	convey.Convey("test NodeController method 'initNodeAnnotation'", t, testInitNodeAnnotation)
}

func testNodeCtlSetNextFaultProcessor() {
	if nodeController == nil {
		panic("nodeController is nil")
	}
	nextFaultProcessor := reporter.NewReporterManager(testK8sClient)
	nodeController.SetNextFaultProcessor(nextFaultProcessor)
	convey.So(nodeController.nextFaultProcessor, convey.ShouldResemble, nextFaultProcessor)
}

func testNodeCtlInit() {
	if nodeController == nil {
		panic("nodeController is nil")
	}
	var p1 = gomonkey.ApplyFuncReturn(GetNodeSN, "123", nil)
	defer p1.Reset()
	var p2 = gomonkey.ApplyMethodReturn(nodeController.kubeClient, "AddAnnotation", nil)
	defer p2.Reset()
	err := nodeController.Init()
	convey.So(err, convey.ShouldBeNil)
}

func testNodeCtlExecute() {
	if nodeController == nil {
		panic("nodeController is nil")
	}
	var p1 = gomonkey.ApplyFuncReturn(clientcmd.BuildConfigFromFlags, nil, testErr)
	defer p1.Reset()
	nodeController.Execute(testFaultDevInfo)
	convey.So(nodeController.faultManager.GetNodeStatus(), convey.ShouldEqual, common.NodeHealthy)
}

func testNodeCtlUpdateConfig() {
	if nodeController == nil {
		panic("nodeController is nil")
	}

	testFaultConfig := &common.FaultConfig{
		FaultTypeCode: faultTypeCode,
	}
	err := nodeController.UpdateConfig(testFaultConfig)
	convey.So(err, convey.ShouldBeNil)

	err = nodeController.UpdateConfig(testWrongFaultConfig)
	expErr := fmt.Errorf("pre separate code %s is conflict, "+
		"please check whether the code not just configured at pre separate level", faultCode1)
	convey.So(err, convey.ShouldResemble, expErr)
}

func testUpdateFaultDevInfo() {
	if nodeController == nil {
		panic("nodeController is nil")
	}
	resetFaultDevInfo()
	nodeController.updateFaultDevInfo()
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

func testInitNodeAnnotation() {
	if nodeController == nil {
		panic("nodeController is nil")
	}
	convey.Convey("when GetNodeSN success and AddAnnotation success, should return nil", func() {
		var p1 = gomonkey.ApplyFuncReturn(GetNodeSN, "123", nil)
		defer p1.Reset()
		var p2 = gomonkey.ApplyMethodReturn(nodeController.kubeClient, "AddAnnotation", nil)
		defer p2.Reset()
		err := nodeController.initNodeAnnotation()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("when GetNodeSN failed and AddAnnotation success, should return err", func() {
		var p1 = gomonkey.ApplyFuncReturn(GetNodeSN, "", errors.New("test err"))
		defer p1.Reset()
		var p2 = gomonkey.ApplyMethodReturn(nodeController.kubeClient, "AddAnnotation", nil)
		defer p2.Reset()
		err := nodeController.initNodeAnnotation()
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("when GetNodeSN success and AddAnnotation failed, should return err", func() {
		var p1 = gomonkey.ApplyFuncReturn(GetNodeSN, "123", nil)
		defer p1.Reset()
		var p2 = gomonkey.ApplyMethodReturn(nodeController.kubeClient,
			"AddAnnotation", errors.New("test err"))
		defer p2.Reset()
		err := nodeController.initNodeAnnotation()
		convey.So(err, convey.ShouldNotBeNil)
	})
}
