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
)

var (
	monitorManager *MonitorManager
)

func TestReportManager(t *testing.T) {
	monitorManager = NewMonitorManager(testK8sClient)
	convey.Convey("test MonitorManager method 'SetNextFaultProcessor'", t, testMonitorMgrSetNextFaultProcessor)
	convey.Convey("test MonitorManager method 'Run'", t, testMonitorMgrRun)
}

func testMonitorMgrSetNextFaultProcessor() {
	if monitorManager == nil {
		panic("monitorManager is nil")
	}
	controller := control.NewControlManager(testK8sClient)
	monitorManager.SetNextFaultProcessor(controller)
	convey.So(monitorManager.nextFaultProcessor, convey.ShouldResemble, controller)
}

func testMonitorMgrRun() {
	if monitorManager == nil {
		panic("monitorManager is nil")
	}
	ctx, cancel := context.WithCancel(context.Background())
	haveStopped := false
	const defaultReportInterval = 5
	select {
	case <-common.GetTrigger():
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

func TestParseTriggers(t *testing.T) {
	deviceInfoHandled := false
	patch := gomonkey.ApplyMethod(&MonitorManager{}, "Execute",
		func(_ *MonitorManager, _ string) {
			deviceInfoHandled = true
			return
		})
	defer patch.Reset()
	convey.Convey("has signal, should update device info", t, func() {
		select {
		case common.GetTrigger() <- "dpc":
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
		case <-common.GetTrigger():
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

func TestTriggerUpdate(t *testing.T) {
	convey.Convey("trigger update success", t, func() {
		verifyUpdateTrigger(t)
		common.TriggerUpdate("test trigger update")
		convey.So(verifyUpdateTrigger(t), convey.ShouldBeTrue)
	})
	convey.Convey("not trigger update", t, func() {
		verifyUpdateTrigger(t)
		if common.GetTrigger() == nil {
			t.Error("updateTriggerChan is nil")
		}
		common.GetTrigger() <- ""
		common.TriggerUpdate("test trigger update")
		convey.So(verifyUpdateTrigger(t), convey.ShouldBeTrue)
	})
}

func verifyUpdateTrigger(t *testing.T) bool {
	if common.GetTrigger() == nil {
		t.Error("updateTriggerChan is nil")
	}
	select {
	case <-common.GetTrigger():
		return true
	default:
		return false
	}
}
