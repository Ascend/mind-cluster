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

// Package ipmimonitor for the ipmi monitor manager test
package ipmimonitor

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/u-root/u-root/pkg/ipmi"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	return initLog()
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return errors.New("init hwlog failed")
	}
	return nil
}

const (
	testDeviceType            = "CPU"
	faultCode1                = "00000001"
	faultCode2                = "00000002"
	waitGoroutineFinishedTime = 1500 * time.Millisecond
)

var (
	testErr         = errors.New("test error")
	testFaultEvents = []*common.FaultEvent{
		{
			ErrorCode:  faultCode1,
			Severity:   0,
			DeviceType: testDeviceType,
			DeviceId:   0,
		},
		{
			ErrorCode:  faultCode2,
			Severity:   1,
			DeviceType: testDeviceType,
			DeviceId:   1,
		},
	}
	alarmResp       = []byte{00, 07, 00, 01, 00, 01, 00, 01, 00, 00, 28, 01, 00, 00, 28, 02, 01, 00, 00, 00, 00, 00, 00}
	alarmResp2      = []byte{00, 07, 00, 01, 00, 01, 128, 01}
	currentAlarmReq = []byte{0x30, 0x94, 0xDB, 0x07, 0x00, 0x40, 0x00, 0x00, 0x00, 0x0E, 0xFF}
)

func TestIpmiEventMonitor(t *testing.T) {
	var patches = gomonkey.ApplyFuncReturn(ipmi.Open, &ipmi.IPMI{}, nil).
		ApplyMethodReturn(&ipmi.IPMI{}, "RawCmd", currentAlarmReq, nil).
		ApplyMethodReturn(&ipmi.IPMI{}, "Close", nil).
		ApplyGlobalVar(&common.ParamOption, common.Option{MonitorPeriod: 1})
	defer patches.Reset()

	convey.Convey("test IpmiEventMonitor method 'Init'", t, testIpmiMonitorInit)
	convey.Convey("test IpmiEventMonitor method 'Monitoring'", t, testIpmiMonitorMonitoring)
	convey.Convey("test IpmiEventMonitor method 'Stop'", t, testIpmiMonitorStop)
	convey.Convey("test IpmiEventMonitor method 'UpdateFaultDevList'", t, testUpdateFaultDevList)
	convey.Convey("test IpmiEventMonitor method 'GetCurrentAlarmFaultEvents'", t, testGetCurrentAlarmFaultEvents)
	convey.Convey("test IpmiEventMonitor method 'GetMonitorData' and 'Name'", t, testGetMonitorDataAndName)
	convey.Convey("test IpmiEventMonitor pure functions", t, testPureFunctions)
}

func testIpmiMonitorInit() {
	convey.Convey("test method Init success", func() {
		monitor := NewIpmiEventMonitor()
		err := monitor.Init()
		convey.So(err, convey.ShouldBeNil)
	})
}

func testIpmiMonitorMonitoring() {
	convey.Convey("test method Monitoring", func() {
		monitor := NewIpmiEventMonitor()
		var p1 = gomonkey.ApplyMethodReturn(&IpmiEventMonitor{}, "GetCurrentAlarmFaultEvents", testFaultEvents, nil)
		defer p1.Reset()
		go func() {
			monitor.Monitoring()
		}()
		time.Sleep(waitGoroutineFinishedTime)
		faultEvent1 := monitor.faultManager.GetFaultDevList()
		faultEvent2 := GetFaultDevList(testFaultEvents)
		convey.So(len(faultEvent1), convey.ShouldEqual, len(faultEvent2))
		for _, fault1 := range faultEvent1 {
			if fault1 == nil {
				continue
			}
			for _, fault2 := range faultEvent2 {
				if fault2 == nil {
					continue
				}
				if fault1.DeviceType == fault2.DeviceType && fault1.DeviceId == fault2.DeviceId {
					convey.So(fault1.FaultCode, convey.ShouldResemble, fault2.FaultCode)
				}
			}
		}
		monitor.Stop()
	})
}

func testIpmiMonitorStop() {
	convey.Convey("test method Stop when ipmiTool is nil", func() {
		monitor := NewIpmiEventMonitor()
		monitor.Stop()
		convey.So(<-monitor.stopChan, convey.ShouldResemble, struct{}{})
	})

	convey.Convey("test method Stop when ipmiTool is not nil and close success", func() {
		monitor := NewIpmiEventMonitor()
		monitor.ipmiTool = &ipmi.IPMI{}
		monitor.Stop()
		convey.So(<-monitor.stopChan, convey.ShouldResemble, struct{}{})
	})

	convey.Convey("test method Stop when close failed", func() {
		monitor := NewIpmiEventMonitor()
		monitor.ipmiTool = &ipmi.IPMI{}
		var p = gomonkey.ApplyMethodReturn(&ipmi.IPMI{}, "Close", testErr)
		defer p.Reset()
		monitor.Stop()
		convey.So(<-monitor.stopChan, convey.ShouldResemble, struct{}{})
	})
}

func testUpdateFaultDevList() {
	convey.Convey("test UpdateFaultDevList open ipmi failed", func() {
		monitor := NewIpmiEventMonitor()
		var p = gomonkey.ApplyFuncReturn(ipmi.Open, &ipmi.IPMI{}, testErr)
		defer p.Reset()
		err := monitor.UpdateFaultDevList()
		convey.So(err, convey.ShouldResemble, testErr)
	})

	convey.Convey("test UpdateFaultDevList open success and update success", func() {
		monitor := NewIpmiEventMonitor()
		var p = gomonkey.ApplyMethodReturn(&IpmiEventMonitor{}, "GetCurrentAlarmFaultEvents", testFaultEvents, nil)
		defer p.Reset()
		err := monitor.UpdateFaultDevList()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test UpdateFaultDevList with opened tool and update success", func() {
		monitor := NewIpmiEventMonitor()
		monitor.ipmiTool = &ipmi.IPMI{}
		var p = gomonkey.ApplyMethodReturn(&IpmiEventMonitor{}, "GetCurrentAlarmFaultEvents", testFaultEvents, nil)
		defer p.Reset()
		err := monitor.UpdateFaultDevList()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test UpdateFaultDevList update failed and close success", func() {
		monitor := NewIpmiEventMonitor()
		var p = gomonkey.ApplyMethodReturn(&IpmiEventMonitor{}, "GetCurrentAlarmFaultEvents", nil, testErr)
		defer p.Reset()
		err := monitor.UpdateFaultDevList()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(monitor.ipmiTool, convey.ShouldBeNil)
	})

	convey.Convey("test UpdateFaultDevList update failed and close failed", func() {
		monitor := NewIpmiEventMonitor()
		var p1 = gomonkey.ApplyMethodReturn(&IpmiEventMonitor{}, "GetCurrentAlarmFaultEvents", nil, testErr)
		defer p1.Reset()
		var p2 = gomonkey.ApplyMethodReturn(&ipmi.IPMI{}, "Close", testErr)
		defer p2.Reset()
		err := monitor.UpdateFaultDevList()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(monitor.ipmiTool, convey.ShouldBeNil)
	})
}

func testGetCurrentAlarmFaultEvents() {
	convey.Convey("test method GetCurrentAlarmFaultEvents success", func() {
		monitor := NewIpmiEventMonitor()
		monitor.ipmiTool = &ipmi.IPMI{}
		var p1 = gomonkey.ApplyMethodSeq(&ipmi.IPMI{}, "RawCmd", []gomonkey.OutputCell{
			{Values: gomonkey.Params{alarmResp, nil}},
			{Values: gomonkey.Params{alarmResp2, nil}, Times: 2},
		})
		defer p1.Reset()
		_, err := monitor.GetCurrentAlarmFaultEvents()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test method GetCurrentAlarmFaultEvents failed, RawCmd error", func() {
		monitor := NewIpmiEventMonitor()
		monitor.ipmiTool = &ipmi.IPMI{}
		var p2 = gomonkey.ApplyMethodSeq(&ipmi.IPMI{}, "RawCmd", []gomonkey.OutputCell{
			{Values: gomonkey.Params{alarmResp, nil}},
			{Values: gomonkey.Params{nil, testErr}},
		})
		defer p2.Reset()
		_, err := monitor.GetCurrentAlarmFaultEvents()
		expErr := errors.New("get another alarm msg from ipmi failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func testGetMonitorDataAndName() {
	convey.Convey("test method GetMonitorData", func() {
		monitor := NewIpmiEventMonitor()
		data := monitor.GetMonitorData()
		convey.So(data, convey.ShouldNotBeNil)
	})

	convey.Convey("test method Name", func() {
		monitor := NewIpmiEventMonitor()
		convey.So(monitor.Name(), convey.ShouldEqual, common.PluginMonitorIpmi)
	})
}

func testPureFunctions() {
	convey.Convey("test GetCurrentAlarmReq", func() {
		req := GetCurrentAlarmReq(0)
		convey.So(req, convey.ShouldNotBeNil)
		convey.So(len(req), convey.ShouldBeGreaterThan, 0)
	})

	convey.Convey("test GetTotalEventsNum success", func() {
		num := GetTotalEventsNum([]byte{0x28, 0x01})
		convey.So(num, convey.ShouldEqual, int64(296))
	})

	convey.Convey("test GetTotalEventsNum invalid length", func() {
		num := GetTotalEventsNum([]byte{0x28})
		convey.So(num, convey.ShouldEqual, int64(0))
	})

	convey.Convey("test GetFaultEvents invalid length", func() {
		events := GetFaultEvents([]byte{0x00, 0x01})
		convey.So(events, convey.ShouldBeNil)
	})

	convey.Convey("test GetFaultEvent invalid length", func() {
		event := GetFaultEvent([]byte{0x00, 0x01})
		convey.So(event, convey.ShouldBeNil)
	})

	convey.Convey("test GetFaultDevList", func() {
		list := GetFaultDevList(testFaultEvents)
		convey.So(len(list), convey.ShouldEqual, 2)
	})
}
