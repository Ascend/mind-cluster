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

// Package monitoring for the monitor manager test
package monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"nodeD/pkg/common"
	"nodeD/pkg/control"
	"nodeD/pkg/monitoring/ipmimonitor"
)

var (
	monitorManager *MonitorManager
)

func TestReportManager(t *testing.T) {
	monitorManager = NewMonitorManager(testK8sClient)
	convey.Convey("test MonitorManager method 'SetNextFaultProcessor'", t, testMonitorMgrSetNextFaultProcessor)
	convey.Convey("test MonitorManager method 'Init'", t, testMonitorMgrInit)
	convey.Convey("test MonitorManager method 'Run'", t, testMonitorMgrRun)
	convey.Convey("test MonitorManager method 'Execute'", t, testMonitorMgrExecute)
}

func testMonitorMgrSetNextFaultProcessor() {
	if monitorManager == nil {
		panic("monitorManager is nil")
	}
	controller := control.NewNodeController(testK8sClient)
	monitorManager.SetNextFaultProcessor(controller)
	convey.So(monitorManager.nextFaultProcessor, convey.ShouldResemble, controller)
}

func testMonitorMgrInit() {
	if monitorManager == nil {
		panic("monitorManager is nil")
	}
	var err error
	go func() {
		err = monitorManager.Init()
	}()
	time.Sleep(waitGoroutineFinishedTime)
	monitorManager.Stop()
	convey.So(err, convey.ShouldBeNil)

	var p1 = gomonkey.ApplyMethodReturn(&ipmimonitor.IpmiEventMonitor{}, "Init", mockErr)
	defer p1.Reset()
	go func() {
		err = monitorManager.Init()
	}()
	time.Sleep(retryInterval)
	monitorManager.Stop()
	convey.So(err, convey.ShouldBeNil)
}

func testMonitorMgrRun() {
	if monitorManager == nil {
		panic("monitorManager is nil")
	}
	ctx, cancel := context.WithCancel(context.Background())
	haveStopped := false
	const defaultReportInterval = 5
	select {
	case <-common.GetUpdateChan():
		fmt.Println("clear update chan")
	default:
		fmt.Println("update chan already clear")
	}
	go func() {
		common.ParamOption.ReportInterval = defaultReportInterval
		monitorManager.Run(ctx)
		haveStopped = true
	}()
	cancel()
	time.Sleep(waitGoroutineFinishedTime)
	convey.So(haveStopped, convey.ShouldBeTrue)
}

func testMonitorMgrExecute() {
	if monitorManager == nil {
		panic("monitorManager is nil")
	}

	testFaultDevInfo := &common.FaultDevInfo{
		FaultDevList: []*common.FaultDev{
			{
				DeviceType: testDeviceType,
				DeviceId:   0,
				FaultCode:  []string{faultCode1, faultCode2},
				FaultLevel: common.PreSeparateFault,
			},
			{
				DeviceType: testDeviceType,
				DeviceId:   1,
				FaultCode:  []string{faultCode1, faultCode2},
				FaultLevel: common.PreSeparateFault,
			},
		},
		NodeStatus: common.PreSeparate,
	}
	var p1 = gomonkey.ApplyMethodReturn(&control.NodeController{}, "Execute")
	defer p1.Reset()
	monitorManager.Execute(testFaultDevInfo)
}

func TestParseTriggers(t *testing.T) {
	deviceInfoHandled := false
	patch := gomonkey.ApplyMethod(&MonitorManager{}, "Execute",
		func(_ *MonitorManager, faultDevInfo *common.FaultDevInfo) {
			deviceInfoHandled = true
			return
		})
	defer patch.Reset()
	convey.Convey("has signal, should update device info", t, func() {
		select {
		case common.GetUpdateChan() <- struct{}{}:
			fmt.Print("send to update chane")
		default:
			fmt.Println("update channel is full")
		}
		if monitorManager == nil {
			t.Error("monitorManager is nil")
		}
		monitorManager.parseTriggers()
		convey.So(deviceInfoHandled, convey.ShouldBeTrue)
	})
	convey.Convey("no signal, should not update device info", t, func() {
		deviceInfoHandled = false
		select {
		case <-common.GetUpdateChan():
			fmt.Print("clear update chane")
		default:
			fmt.Println("update channel is empty")
		}
		if monitorManager == nil {
			t.Error("monitorManager is nil")
		}
		monitorManager.parseTriggers()
		convey.So(deviceInfoHandled, convey.ShouldBeFalse)
	})
}
