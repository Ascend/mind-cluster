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

// Package reporter for the report manager test
package reporter

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
	"nodeD/pkg/reporter/cmreporter"
)

var reportManager *ReportManager

func TestReportManager(t *testing.T) {
	reportManager = NewReporterManager(testK8sClient)
	convey.Convey("test ReportManager method 'SetNextFaultProcessor'", t, testReportMgrSetNextFaultProcessor)
	convey.Convey("test ReportManager method 'Execute'", t, testReportMgrExecute)
}

func testReportMgrSetNextFaultProcessor() {
	if reportManager == nil {
		panic("reportManager is nil")
	}
	reportManager.SetNextFaultProcessor(reportManager)
	convey.So(reportManager.nextFaultProcessor, convey.ShouldResemble, reportManager)
}

func testReportMgrExecute() {
	if reportManager == nil {
		panic("reportManager is nil")
	}
	setReporters()
	convey.Convey("test method Execute success", testExc)
	convey.Convey("test method Execute failed, ConfigMapReporter unmarshal error", testExcErrUnmarshal)
	convey.Convey("test method Execute failed, ConfigMapReporter update cm error", testExcErrUpdate)
}

func testExc() {
	go func() {
		reportManager.Execute(testFCDevInfo, "")
	}()
	time.Sleep(waitGoroutineFinishedTime)
	_, err := testK8sClient.GetConfigMap(testNodeInfoName, api.DLNamespace)
	convey.So(err, convey.ShouldNotBeNil)
}

func testExcErrUnmarshal() {
	err := deleteCM(testNodeInfoName, api.DLNamespace)
	convey.So(err, convey.ShouldBeNil)
	var p1 = gomonkey.ApplyFuncReturn(json.Marshal, nil, testErr)
	defer p1.Reset()
	go func() {
		reportManager.Execute(testFCDevInfo, "")
	}()
	time.Sleep(waitGoroutineFinishedTime)
	_, err = testK8sClient.GetConfigMap(testNodeInfoName, api.DLNamespace)
	convey.So(errors.IsNotFound(err), convey.ShouldBeTrue)
}

func testExcErrUpdate() {
	const retryInterval = 4 * time.Second
	err := deleteCM(testNodeInfoName, api.DLNamespace)
	convey.So(err, convey.ShouldBeNil)
	var p2 = gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "UpdateConfigMap", nil, testErr)
	defer p2.Reset()
	go func() {
		reportManager.Execute(testFCDevInfo, "")
	}()
	time.Sleep(retryInterval)
	_, err = testK8sClient.GetConfigMap(testNodeInfoName, api.DLNamespace)
	convey.So(errors.IsNotFound(err), convey.ShouldBeTrue)
}

func setReporters() {
	if reportManager == nil {
		panic("reportManager is nil")
	}
	cmReporter := cmreporter.NewConfigMapReporter(testK8sClient)
	reportManager.reporters = append(reportManager.reporters, cmReporter)
}

func parseNodeInfoCMData(data string) common.NodeInfoCM {
	var nodeInfo common.NodeInfoCM
	err := json.Unmarshal([]byte(data), &nodeInfo)
	convey.So(err, convey.ShouldBeNil)
	return nodeInfo
}
