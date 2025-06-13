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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"nodeD/pkg/control/nodesn"
	"nodeD/pkg/reporter"
)

var controlManager *ControllerManager

func TestNodeController(t *testing.T) {
	controlManager = NewControlManager(testK8sClient)
	convey.Convey("test NodeController method 'SetNextFaultProcessor'", t, testNodeCtlSetNextFaultProcessor)
	convey.Convey("test NodeController method 'Init'", t, testNodeCtlInit)
	convey.Convey("test NodeController method 'initNodeAnnotation'", t, testInitNodeAnnotation)
}

func testNodeCtlSetNextFaultProcessor() {
	if controlManager == nil {
		panic("nodeController is nil")
	}
	nextFaultProcessor := reporter.NewReporterManager(testK8sClient)
	controlManager.SetNextFaultProcessor(nextFaultProcessor)
	convey.So(controlManager.nextFaultProcessor, convey.ShouldResemble, nextFaultProcessor)
}

func testNodeCtlInit() {
	if controlManager == nil {
		panic("nodeController is nil")
	}
	var p1 = gomonkey.ApplyFuncReturn(nodesn.GetNodeSN, "123", nil)
	defer p1.Reset()
	var p2 = gomonkey.ApplyMethodReturn(controlManager.kubeClient, "AddAnnotation", nil)
	defer p2.Reset()
	err := controlManager.Init()
	convey.So(err, convey.ShouldBeNil)
}

func testInitNodeAnnotation() {
	if controlManager == nil {
		panic("nodeController is nil")
	}
	convey.Convey("when GetNodeSN success and AddAnnotation success, should return nil", func() {
		var p1 = gomonkey.ApplyFuncReturn(nodesn.GetNodeSN, "123", nil)
		defer p1.Reset()
		var p2 = gomonkey.ApplyMethodReturn(controlManager.kubeClient, "AddAnnotation", nil)
		defer p2.Reset()
		err := controlManager.InitNodeAnnotation()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("when GetNodeSN failed and AddAnnotation success, should return err", func() {
		var p1 = gomonkey.ApplyFuncReturn(nodesn.GetNodeSN, "", errors.New("test err"))
		defer p1.Reset()
		var p2 = gomonkey.ApplyMethodReturn(controlManager.kubeClient, "AddAnnotation", nil)
		defer p2.Reset()
		err := controlManager.InitNodeAnnotation()
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("when GetNodeSN success and AddAnnotation failed, should return err", func() {
		var p1 = gomonkey.ApplyFuncReturn(nodesn.GetNodeSN, "123", nil)
		defer p1.Reset()
		var p2 = gomonkey.ApplyMethodReturn(controlManager.kubeClient,
			"AddAnnotation", errors.New("test err"))
		defer p2.Reset()
		err := controlManager.InitNodeAnnotation()
		convey.So(err, convey.ShouldNotBeNil)
	})
}
