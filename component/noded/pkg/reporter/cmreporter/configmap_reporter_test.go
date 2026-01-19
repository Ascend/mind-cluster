//go:build !race

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

// Package cmreporter for the cm report manager test
package cmreporter

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/api/errors"

	"ascend-common/api"
	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
)

var cmReporter *ConfigMapReporter

func TestCMReporterReport(t *testing.T) {
	cmReporter = NewConfigMapReporter(testK8sClient)
	if cmReporter == nil {
		t.Error("cmReporter is nil")
	}
	convey.Convey("test ConfigMapReporter method 'Report', origin: no cm; after: no cm", t, testReportNoToNo)
	convey.Convey("test ConfigMapReporter method 'Report', origin: no cm; after: has cm", t, testReportNoToHas)
	convey.Convey("test ConfigMapReporter method 'Report', origin: has cm; after: no cm", t, testReportHasToNo)
	convey.Convey("test ConfigMapReporter method 'Report', origin: has cm; after: has cm", t, testReportHasToHas)
	convey.Convey("test ConfigMapReporter method 'Report', delete error", t, testReportErrDelete)
	convey.Convey("test ConfigMapReporter method 'Report', origin: has cm; after: has cm", t, testReportErrUnmarshal)
	convey.Convey("test ConfigMapReporter method 'Report', origin: has cm; after: has cm", t, testReportErrUpdate)
}

func testReportNoToNo() {
	// origin: no fault, no cm; after: no fault, no cm
	err := deleteCM(testNodeInfoName, api.DLNamespace)
	convey.So(err, convey.ShouldBeNil)

	if cmReporter == nil {
		panic("cmReporter is nil")
	}
	cmReporter.nodeInfoCache.NodeInfo = *testNormalDevInfo
	go func() {
		cmReporter.Report(fcNormalInfo)
	}()
	time.Sleep(waitGoroutineFinishedTime)

	cm, err := testK8sClient.GetConfigMap(testNodeInfoName, api.DLNamespace)
	convey.So(errors.IsNotFound(err), convey.ShouldBeTrue)
	convey.So(cm, convey.ShouldBeNil)
}

func testReportNoToHas() {
	// origin: no fault, no cm; after: has fault, has cm
	err := deleteCM(testNodeInfoName, api.DLNamespace)
	convey.So(err, convey.ShouldBeNil)

	if cmReporter == nil {
		panic("cmReporter is nil")
	}
	cmReporter.nodeInfoCache.NodeInfo = *testNormalDevInfo
	go func() {
		cmReporter.Report(fcFaultInfo)
	}()
	time.Sleep(waitGoroutineFinishedTime)

	cm, err := testK8sClient.GetConfigMap(testNodeInfoName, api.DLNamespace)
	convey.So(err, convey.ShouldBeNil)
	cmInfo := parseNodeInfoCMData(cm.Data[api.NodeInfoCMDataKey])
	convey.So(len(cmInfo.NodeInfo.FaultDevList), convey.ShouldEqual, len(testFaultDevList))
	convey.So(err, convey.ShouldBeNil)
}

func testReportHasToNo() {
	// origin: has fault, has cm; after: no fault, no cm
	cm, err := testK8sClient.GetConfigMap(testNodeInfoName, api.DLNamespace)
	convey.So(err, convey.ShouldBeNil)
	cmInfo := parseNodeInfoCMData(cm.Data[api.NodeInfoCMDataKey])
	convey.So(len(cmInfo.NodeInfo.FaultDevList), convey.ShouldEqual, len(testFaultDevList))
	convey.So(err, convey.ShouldBeNil)

	if cmReporter == nil {
		panic("cmReporter is nil")
	}
	cmReporter.nodeInfoCache.NodeInfo = *testFaultDevInfo
	go func() {
		cmReporter.Report(fcNormalInfo)
	}()
	time.Sleep(waitGoroutineFinishedTime)

	cm, err = testK8sClient.GetConfigMap(testNodeInfoName, api.DLNamespace)
	convey.So(errors.IsNotFound(err), convey.ShouldBeTrue)
	convey.So(cm, convey.ShouldBeNil)
}

func testReportHasToHas() {
	// origin: has fault, has cm; after: has fault, has cm
	testReportNoToHas()
	if cmReporter == nil {
		panic("cmReporter is nil")
	}
	cmReporter.nodeInfoCache.NodeInfo = *testNormalDevInfo
	go func() {
		cmReporter.Report(fcFaultInfo)
	}()
	time.Sleep(waitGoroutineFinishedTime)

	cm, err := testK8sClient.GetConfigMap(testNodeInfoName, api.DLNamespace)
	convey.So(err, convey.ShouldBeNil)
	cmInfo := parseNodeInfoCMData(cm.Data[api.NodeInfoCMDataKey])
	convey.So(len(cmInfo.NodeInfo.FaultDevList), convey.ShouldEqual, len(testFaultDevList))
	convey.So(err, convey.ShouldBeNil)
}

func testReportErrDelete() {
	var p1 = gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "DeleteConfigMap", testErr)
	defer p1.Reset()

	testReportNoToHas()
	if cmReporter == nil {
		panic("cmReporter is nil")
	}
	cmReporter.nodeInfoCache.NodeInfo = *testFaultDevInfo
	go func() {
		cmReporter.Report(fcNormalInfo)
	}()
	time.Sleep(waitGoroutineFinishedTime)

	cm, err := testK8sClient.GetConfigMap(testNodeInfoName, api.DLNamespace)
	convey.So(err, convey.ShouldBeNil)
	cmInfo := parseNodeInfoCMData(cm.Data[api.NodeInfoCMDataKey])
	convey.So(len(cmInfo.NodeInfo.FaultDevList), convey.ShouldEqual, len(testFaultDevList))
	convey.So(err, convey.ShouldBeNil)
}

func testReportErrUnmarshal() {
	err := deleteCM(testNodeInfoName, api.DLNamespace)
	convey.So(err, convey.ShouldBeNil)
	var p1 = gomonkey.ApplyFuncReturn(json.Marshal, nil, testErr)
	defer p1.Reset()
	if cmReporter == nil {
		panic("cmReporter is nil")
	}
	cmReporter.nodeInfoCache.NodeInfo = *testNormalDevInfo
	go func() {
		cmReporter.Report(fcFaultInfo)
	}()
	time.Sleep(waitGoroutineFinishedTime)

	_, err = testK8sClient.GetConfigMap(testNodeInfoName, api.DLNamespace)
	convey.So(errors.IsNotFound(err), convey.ShouldBeTrue)
}

func testReportErrUpdate() {
	const retryInterval = 4 * time.Second
	err := deleteCM(testNodeInfoName, api.DLNamespace)
	convey.So(err, convey.ShouldBeNil)
	var p2 = gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "UpdateConfigMap", nil, testErr)
	defer p2.Reset()
	if cmReporter == nil {
		panic("cmReporter is nil")
	}
	cmReporter.nodeInfoCache.NodeInfo = *testNormalDevInfo
	go func() {
		cmReporter.Report(fcFaultInfo)
	}()
	time.Sleep(retryInterval)

	_, err = testK8sClient.GetConfigMap(testNodeInfoName, api.DLNamespace)
	convey.So(errors.IsNotFound(err), convey.ShouldBeTrue)
}

func parseNodeInfoCMData(data string) common.NodeInfoCM {
	var nodeInfo common.NodeInfoCM
	err := json.Unmarshal([]byte(data), &nodeInfo)
	convey.So(err, convey.ShouldBeNil)
	return nodeInfo
}

func TestCMReporterInit(t *testing.T) {
	cmReporter = NewConfigMapReporter(testK8sClient)
	if cmReporter == nil {
		t.Error("cmReporter is nil")
	}
	convey.Convey("test ConfigMapReporter method 'Report', origin: no cm; after: no cm", t, func() {
		convey.So(cmReporter.Init(), convey.ShouldBeNil)
	})
}
