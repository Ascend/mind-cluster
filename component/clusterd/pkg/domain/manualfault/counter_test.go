/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package manualfault test of counter for hardware frequency fault
package manualfault

import (
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/domain/conf"
)

// node1, dev1, code1, time1
func getDemoFault1() *Fault {
	return &Fault{
		Code:        code1,
		JobId:       job1,
		NodeName:    node1,
		DevName:     dev1,
		ReceiveTime: receiveTime1,
	}
}

// node1, dev1, code1, time2
func getDemoFault2() *Fault {
	return &Fault{
		Code:        code1,
		JobId:       job1,
		NodeName:    node1,
		DevName:     dev1,
		ReceiveTime: receiveTime2,
	}
}

// node1, dev1, code1, time3
func getDemoFault3() *Fault {
	return &Fault{
		Code:        code1,
		JobId:       job1,
		NodeName:    node1,
		DevName:     dev1,
		ReceiveTime: receiveTime3,
	}
}

// node1, dev1, code2, time4
func getDemoFault4() *Fault {
	return &Fault{
		Code:        code2,
		JobId:       job1,
		NodeName:    node1,
		DevName:     dev1,
		ReceiveTime: receiveTime4,
	}
}

// node1, dev2, code1, time5
func getDemoFault5() *Fault {
	return &Fault{
		Code:        code1,
		JobId:       job1,
		NodeName:    node1,
		DevName:     dev2,
		ReceiveTime: receiveTime5,
	}
}

// node2, dev3, code2, time2
func getDemoFault6() *Fault {
	return &Fault{
		Code:        code2,
		JobId:       job1,
		NodeName:    node2,
		DevName:     dev3,
		ReceiveTime: receiveTime2,
	}
}

const (
	defaultFaultWindowHours = 24
	defaultFaultThreshold   = 3
	defaultFaultFreeHours   = 48
)

var (
	validPolicy = conf.ManuallySeparatePolicy{
		Separate: struct {
			FaultWindowHours int `yaml:"fault_window_hours"`
			FaultThreshold   int `yaml:"fault_threshold"`
		}{
			FaultWindowHours: defaultFaultWindowHours,
			FaultThreshold:   defaultFaultThreshold,
		},
		Release: struct {
			FaultFreeHours int `yaml:"fault_free_hours"`
		}{
			FaultFreeHours: defaultFaultFreeHours,
		},
	}
	threshold1 = conf.ManuallySeparatePolicy{
		Separate: struct {
			FaultWindowHours int `yaml:"fault_window_hours"`
			FaultThreshold   int `yaml:"fault_threshold"`
		}{
			FaultWindowHours: defaultFaultWindowHours,
			FaultThreshold:   1,
		},
		Release: struct {
			FaultFreeHours int `yaml:"fault_free_hours"`
		}{
			FaultFreeHours: defaultFaultFreeHours,
		},
	}
)

func TestAddFault(t *testing.T) {
	conf.SetManualSeparatePolicy(validPolicy)
	testTime := time.Date(2026, 02, 14, 10, 0, 0, 0, time.UTC)
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&time.Time{}), "UnixMilli", func(_ *time.Time) int64 {
		return testTime.UnixMilli()
	})
	defer patch.Reset()
	convey.Convey("test func AddFault, input is nil", t, testNilInput)
	convey.Convey("test func AddFault, reach frequency", t, testReachFrequency)
	convey.Convey("test func AddFault, add different code", t, testAddDifferentCode)
	convey.Convey("test func AddFault, add different dev", t, testAddDifferentDev)
	convey.Convey("test func AddFault, add different node", t, testAddDifferentNode)
}

func testNilInput() {
	InitCounter()
	Counter.AddFault(nil)
	convey.So(len(Counter.faults), convey.ShouldEqual, 0)
}

func testReachFrequency() {
	InitCounter()
	// add fault1: node1, dev1, code1, time1
	fault1 := getDemoFault1()
	Counter.AddFault(fault1)
	convey.So(len(Counter.faults), convey.ShouldEqual, len1)
	info, ok := Counter.faults[fault1.NodeName]
	convey.So(ok, convey.ShouldBeTrue)
	f, ok := info[fault1.DevName]
	convey.So(ok, convey.ShouldBeTrue)
	times, ok := f.fault[fault1.Code]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(times, convey.ShouldResemble, []int64{fault1.ReceiveTime})

	// add fault2: node1, dev1, code1, time2
	fault2 := getDemoFault2()
	Counter.AddFault(fault2)
	convey.So(len(Counter.faults), convey.ShouldEqual, len1)
	info, ok = Counter.faults[node1]
	convey.So(ok, convey.ShouldBeTrue)
	f, ok = info[dev1]
	convey.So(ok, convey.ShouldBeTrue)
	times, ok = f.fault[fault2.Code]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(times, convey.ShouldResemble, []int64{fault1.ReceiveTime, fault2.ReceiveTime})

	// add fault3: node1, dev1, code1, time3
	fault3 := getDemoFault3()
	Counter.AddFault(fault3)
	convey.So(len(Counter.faults), convey.ShouldEqual, 0)
}

func testAddDifferentCode() {
	InitCounter()
	// add fault1: node1, dev1, code1, time1
	fault1 := getDemoFault1()
	Counter.AddFault(fault1)
	convey.So(len(Counter.faults), convey.ShouldEqual, len1)
	info, ok := Counter.faults[fault1.NodeName]
	convey.So(ok, convey.ShouldBeTrue)
	f, ok := info[fault1.DevName]
	convey.So(ok, convey.ShouldBeTrue)
	times, ok := f.fault[fault1.Code]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(times, convey.ShouldResemble, []int64{fault1.ReceiveTime})

	// add fault4: node1, dev1, code2, time1
	fault4 := getDemoFault4()
	Counter.AddFault(fault4)
	convey.So(len(Counter.faults), convey.ShouldEqual, len1)
	info, ok = Counter.faults[fault4.NodeName]
	convey.So(ok, convey.ShouldBeTrue)
	f, ok = info[fault4.DevName]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(f.fault), convey.ShouldEqual, len2)
	times, ok = f.fault[fault4.Code]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(times, convey.ShouldResemble, []int64{fault4.ReceiveTime})
}

func testAddDifferentDev() {
	InitCounter()
	// add fault1: node1, dev1, code1, time1
	fault1 := getDemoFault1()
	Counter.AddFault(fault1)
	convey.So(len(Counter.faults), convey.ShouldEqual, len1)
	info, ok := Counter.faults[fault1.NodeName]
	convey.So(ok, convey.ShouldBeTrue)
	f, ok := info[fault1.DevName]
	convey.So(ok, convey.ShouldBeTrue)
	times, ok := f.fault[fault1.Code]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(times, convey.ShouldResemble, []int64{fault1.ReceiveTime})

	// add fault5: node1, dev2, code1, time2
	fault5 := getDemoFault5()
	Counter.AddFault(fault5)
	convey.So(len(Counter.faults), convey.ShouldEqual, len1)
	info, ok = Counter.faults[fault5.NodeName]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(info), convey.ShouldEqual, len2)
	f, ok = info[fault5.DevName]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(f.fault), convey.ShouldEqual, len1)
	times, ok = f.fault[fault5.Code]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(times, convey.ShouldResemble, []int64{fault5.ReceiveTime})
}

func testAddDifferentNode() {
	InitCounter()
	// add fault1: node1, dev1, code1, time1
	fault1 := getDemoFault1()
	Counter.AddFault(fault1)
	convey.So(len(Counter.faults), convey.ShouldEqual, len1)
	info, ok := Counter.faults[fault1.NodeName]
	convey.So(ok, convey.ShouldBeTrue)
	f, ok := info[fault1.DevName]
	convey.So(ok, convey.ShouldBeTrue)
	times, ok := f.fault[fault1.Code]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(times, convey.ShouldResemble, []int64{fault1.ReceiveTime})

	// add fault6: node2, dev3, code2, time2
	fault6 := getDemoFault6()
	Counter.AddFault(fault6)
	convey.So(len(Counter.faults), convey.ShouldEqual, len2)
	info, ok = Counter.faults[fault6.NodeName]
	convey.So(ok, convey.ShouldBeTrue)
	f, ok = info[fault6.DevName]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(f.fault), convey.ShouldEqual, len1)
	times, ok = f.fault[fault6.Code]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(times, convey.ShouldResemble, []int64{fault6.ReceiveTime})
}

func TestThresholdIs1(t *testing.T) {
	convey.Convey("test frequency fault threshold is 1", t, testThreshold1)
	convey.Convey("test frequency fault threshold is 1, fault 2 times -> threshold 1 -> clear", t, testUnReachThreshold)
	convey.Convey("test frequency fault threshold is 1, new code", t, testDifferentCode)
	convey.Convey("test frequency fault threshold is 1, new dev", t, testDifferentDev)
}

func testThreshold1() {
	InitCounter()
	conf.SetManualSeparatePolicy(threshold1)

	// add fault1: node1, dev1, code1, time1
	Counter.AddFault(getDemoFault1())
	convey.So(len(Counter.faults), convey.ShouldEqual, len0)
}

func testUnReachThreshold() {
	InitCounter()
	conf.SetManualSeparatePolicy(validPolicy)
	// add fault1: node1, dev1, code1, time1
	Counter.AddFault(getDemoFault1())
	// add fault2: node1, dev1, code1, time2
	Counter.AddFault(getDemoFault2())
	conf.SetManualSeparatePolicy(threshold1)
	// add fault3: node1, dev1, code1, time3
	Counter.AddFault(getDemoFault3())
	convey.So(len(Counter.faults), convey.ShouldEqual, len0)
}

func testDifferentCode() {
	InitCounter()
	conf.SetManualSeparatePolicy(validPolicy)
	// add fault1: node1, dev1, code1, time1
	Counter.AddFault(getDemoFault1())
	convey.So(len(Counter.faults), convey.ShouldEqual, len1)
	conf.SetManualSeparatePolicy(threshold1)
	// add fault4: node1, dev1, code2, time4
	Counter.AddFault(getDemoFault4())
	convey.So(len(Counter.faults), convey.ShouldEqual, len1)
}

func testDifferentDev() {
	InitCounter()
	conf.SetManualSeparatePolicy(validPolicy)
	// add fault1: node1, dev1, code1, time1
	Counter.AddFault(getDemoFault1())
	convey.So(len(Counter.faults), convey.ShouldEqual, len1)
	conf.SetManualSeparatePolicy(threshold1)
	// add fault5: node1, dev2, code1, time5
	Counter.AddFault(getDemoFault5())
	convey.So(len(Counter.faults), convey.ShouldEqual, len1)
}

func TestClearFault(t *testing.T) {
	conf.SetManualSeparatePolicy(validPolicy)
	convey.Convey("test func ClearDevFault, data prepare", t, testPrepare)
	convey.Convey("test func ClearDevFault, clear code only", t, testClearCode)
	convey.Convey("test func ClearDevFault, clear code -> clear dev", t, testClearDev)
	convey.Convey("test func ClearDevFault, clear not exits item", t, testClearNotExist)
	convey.Convey("test func ClearDevFault, clear dev -> clear node", t, testClearNode)
}

func testPrepare() {
	InitCounter()
	// add fault1: node1, dev1, code1, time1
	Counter.AddFault(getDemoFault1())
	// add fault2: node1, dev1, code1, time2
	Counter.AddFault(getDemoFault2())
	// add fault4: node1, dev1, code2, time4
	Counter.AddFault(getDemoFault4())
	// add fault5: node1, dev2, code1, time5
	Counter.AddFault(getDemoFault5())
	// add fault6: node2, dev3, code2, time2
	Counter.AddFault(getDemoFault6())
	convey.So(len(Counter.faults), convey.ShouldEqual, len2) // 2 node
	info, ok := Counter.faults[node1]                        // node 1
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(info), convey.ShouldEqual, len2) // 2 dev
	f, ok := info[dev1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(f.fault), convey.ShouldEqual, len2) // 2 code on dev1
	times, ok := f.fault[code1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(times), convey.ShouldResemble, len2) // 2 times for code1

	times, ok = f.fault[code2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(times), convey.ShouldResemble, len1) // 1 times for code2

	f, ok = info[dev2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(f.fault), convey.ShouldEqual, len1) // 1 code on dev2
	times, ok = f.fault[code1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(times), convey.ShouldResemble, len1) // 1 times for code1

	info, ok = Counter.faults[node2] // node 2
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(info), convey.ShouldEqual, len1) // 1 dev
	f, ok = info[dev3]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(f.fault), convey.ShouldEqual, len1) // 1 code on dev3
	times, ok = f.fault[code2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(times), convey.ShouldResemble, len1) // 1 times for code2
}

func testClearCode() {
	Counter.ClearDevFault(node1, dev1, code1)
	info, ok := Counter.faults[node1] // node 1
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(info), convey.ShouldEqual, len2) // 2 dev
	f, ok := info[dev1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(f.fault), convey.ShouldEqual, len1) // 2 -> 1 code on dev1
	times, ok := f.fault[code1]
	convey.So(ok, convey.ShouldBeFalse)
	times, ok = f.fault[code2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(times), convey.ShouldResemble, len1) // 1 times for code1
}

func testClearDev() {
	Counter.ClearDevFault(node1, dev1, code2)
	info, ok := Counter.faults[node1] // node 1
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(info), convey.ShouldEqual, len1) // 2 -> 1 dev
	_, ok = info[dev1]
	convey.So(ok, convey.ShouldBeFalse)
}

func testClearNotExist() {
	Counter.ClearDevFault(node3, dev2, code1)
	convey.So(len(Counter.faults), convey.ShouldEqual, len2)

	Counter.ClearDevFault(node1, dev3, code1)
	info, ok := Counter.faults[node1] // node 1
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(info), convey.ShouldEqual, len1)
}

func testClearNode() {
	Counter.ClearDevFault(node1, dev2, code1)
	convey.So(len(Counter.faults), convey.ShouldEqual, len1)
	_, ok := Counter.faults[node1] // node 2 -> 1
	convey.So(ok, convey.ShouldBeFalse)

	Counter.ClearDevFault(node2, dev3, code2) // node 1 -> 0
	convey.So(len(Counter.faults), convey.ShouldEqual, len0)
}

func TestIsReachFrequency(t *testing.T) {
	conf.SetManualSeparatePolicy(validPolicy)
	convey.Convey("test func ClearDevFault, data prepare", t, func() {
		times := []int64{receiveTime1, receiveTime2, receiveTime6}
		res := Counter.isReachFrequency(times)
		convey.So(res, convey.ShouldBeFalse)

		times = []int64{receiveTime0, receiveTime1, receiveTime2}
		res = Counter.isReachFrequency(times)
		convey.So(res, convey.ShouldBeFalse)

		conf.SetManualSeparatePolicy(threshold1)
		times = []int64{receiveTime0}
		res = Counter.isReachFrequency(times)
		convey.So(res, convey.ShouldBeTrue)
	})
}
