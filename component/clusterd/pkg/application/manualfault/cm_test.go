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

// Package manualfault test for processing manual separate npu info
package manualfault

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/application/publicfault"
	"clusterd/pkg/application/silentfault"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/manualfault"
)

const (
	node1 = "node1"
	node2 = "node2"

	dev1 = "dev1"
	dev2 = "dev2"
	dev3 = "dev3"

	code1 = "code1"

	len0 = 0
	len1 = 1
	len2 = 2

	receiveTime1 = 1771059600000 // 2026-02-14 09:00:00
	receiveTime3 = 1771059620000 // 2026-02-14 09:00:20
	receiveTime4 = 1771059630000 // 2026-02-14 09:00:30
)

func getDemoNodeInfo() map[string]manualfault.NodeCmInfo {
	return map[string]manualfault.NodeCmInfo{
		node1: {
			Total: []string{dev1},
			Detail: map[string][]manualfault.DevCmInfo{
				dev1: {
					{
						FaultCode:        code1,
						FaultLevel:       constant.ManuallySeparateNPU,
						LastSeparateTime: receiveTime1,
					},
				},
			},
		},
	}
}

func getDemoCm() *v1.ConfigMap {
	info := getDemoNodeInfo()
	data := manualfault.ConvertNodeInfoToCmData(info)
	cm := &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Data:       data,
	}
	return cm
}

func TestCheckManualDiffAndDelete(t *testing.T) {
	convey.Convey("test func checkManualDiffAndDelete success", t, testCheckManualDiffAndDelete)
	convey.Convey("test func checkManualDiffAndDelete failed, manual cm is nil", t, testErrGetCm)
}

func testCheckManualDiffAndDelete() {
	prepareData()
	faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	manualfault.LastCmInfo = faultCmInfo
	checkManualDiffAndDelete(getDemoCm())
	faultCmInfo2, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo2), convey.ShouldEqual, len1)
	info, ok := faultCmInfo2[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1})
	info, ok = faultCmInfo2[node2]
	convey.So(ok, convey.ShouldBeFalse)
}

func testErrGetCm() {
	manualfault.InitFaultCmInfo()
	checkManualDiffAndDelete(nil)
	faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo), convey.ShouldEqual, len0)
}

const (
	defaultFaultWindowHours = 24
	defaultFaultThreshold   = 3
	defaultFaultFreeHours   = 48
)

var validPolicy = conf.ManuallySeparatePolicy{
	Enabled: true,
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

func TestReleaseManualFault(t *testing.T) {
	convey.Convey("test func releaseManualFault success", t, testReleaseManualFault)
	convey.Convey("test func releaseManualFault failed, deep copy error", t, testErrDeepCp)
	convey.Convey("test func releaseManualFault failed, close release switch", t, testCloseRelease)
}

func testReleaseManualFault() {
	testTime := time.Date(2026, 02, 16, 9, 0, 5, 0, time.UTC)
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&time.Time{}), "UnixMilli", func(_ *time.Time) int64 {
		return testTime.UnixMilli()
	})
	defer patch.Reset()
	conf.SetManualSeparatePolicy(validPolicy)
	prepareData()

	faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo), convey.ShouldEqual, len2)
	info, ok := faultCmInfo[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1, dev2})
	info, ok = faultCmInfo[node2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev3})

	releaseManualFault()
	faultCmInfo2, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo2), convey.ShouldEqual, len2)
	info, ok = faultCmInfo2[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev2})
	info, ok = faultCmInfo2[node2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev3})
}

func testErrDeepCp() {
	conf.SetManualSeparatePolicy(validPolicy)
	prepareData()

	faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo), convey.ShouldEqual, len2)
	info, ok := faultCmInfo[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1, dev2})
	info, ok = faultCmInfo[node2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev3})
	p1 := gomonkey.ApplyFuncReturn(util.DeepCopy, testErr)

	releaseManualFault()
	p1.Reset()
	faultCmInfo2, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo2), convey.ShouldEqual, len2)
	info, ok = faultCmInfo2[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1, dev2})
	info, ok = faultCmInfo2[node2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev3})
}

func testCloseRelease() {
	closeRelease := validPolicy
	closeRelease.Enabled = false
	conf.SetManualSeparatePolicy(closeRelease)
	prepareData()

	faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo), convey.ShouldEqual, len2)
	info, ok := faultCmInfo[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1, dev2})
	info, ok = faultCmInfo[node2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev3})
	p1 := gomonkey.ApplyFuncReturn(util.DeepCopy, testErr)

	releaseManualFault()
	p1.Reset()
	faultCmInfo2, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo2), convey.ShouldEqual, len2)
	info, ok = faultCmInfo2[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1, dev2})
	info, ok = faultCmInfo2[node2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev3})
}

func prepareData() {
	manualfault.InitFaultCmInfo()
	fault1 := manualfault.FaultInfo{
		NodeName:    node1,
		DevName:     dev1,
		FaultCode:   code1,
		ReceiveTime: time.Now().Add(-(defaultFaultFreeHours + 1) * time.Hour).UnixMilli(), // before 49h
	}
	fault2 := manualfault.FaultInfo{
		NodeName:    node1,
		DevName:     dev2,
		FaultCode:   code1,
		ReceiveTime: time.Now().Add(-(defaultFaultFreeHours - 1) * time.Hour).UnixMilli(), // before 47h
	}
	fault3 := manualfault.FaultInfo{
		NodeName:    node2,
		DevName:     dev3,
		FaultCode:   code1,
		ReceiveTime: time.Now().Add(-(defaultFaultFreeHours - 1) * time.Hour).UnixMilli(), // before 47h
	}
	manualfault.FaultCmInfo.AddSeparateDev(fault1)
	manualfault.FaultCmInfo.AddSeparateDev(fault2)
	manualfault.FaultCmInfo.AddSeparateDev(fault3)
}

func TestUpdateManualCm(t *testing.T) {
	convey.Convey("test func updateManualCm merges manual and silent entries", t, func() {
		manualInfo := map[string]manualfault.NodeCmInfo{
			node1: {Total: []string{dev1}, Detail: map[string][]manualfault.DevCmInfo{
				dev1: {{FaultCode: code1, FaultLevel: constant.ManuallySeparateNPU}},
			}},
		}
		p1 := gomonkey.ApplyMethodReturn(&manualfault.FaultCmInfo, "DeepCopy", manualInfo, nil)
		defer p1.Reset()
		p2 := gomonkey.ApplyFuncReturn(silentfault.GetSilentFaultCmInfoForMerge, map[string]manualfault.NodeCmInfo{
			node1: {Total: []string{dev2}, Detail: map[string][]manualfault.DevCmInfo{
				dev2: {{FaultCode: code1, FaultLevel: constant.SilentFault}},
			}},
		})
		defer p2.Reset()

		var written map[string]manualfault.NodeCmInfo
		var writtenRV string
		p3 := gomonkey.ApplyFunc(manualfault.UpdateOrCreateManualCm, func(cmInfo map[string]manualfault.NodeCmInfo, rv string) {
			written = cmInfo
			writtenRV = rv
		})
		defer p3.Reset()

		updateManualCm(&v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "12345"}})

		convey.So(written, convey.ShouldNotBeNil)
		convey.So(writtenRV, convey.ShouldEqual, "12345")
		merged := written[node1]
		convey.So(len(merged.Total), convey.ShouldEqual, len2)
		convey.So(len(merged.Detail), convey.ShouldEqual, len2)
	})

	convey.Convey("test func updateManualCm nil cm uses empty resourceVersion", t, func() {
		p1 := gomonkey.ApplyMethodReturn(&manualfault.FaultCmInfo, "DeepCopy", map[string]manualfault.NodeCmInfo{}, nil)
		defer p1.Reset()
		p2 := gomonkey.ApplyFuncReturn(silentfault.GetSilentFaultCmInfoForMerge, map[string]manualfault.NodeCmInfo{})
		defer p2.Reset()

		var writtenRV string
		p3 := gomonkey.ApplyFunc(manualfault.UpdateOrCreateManualCm, func(_ map[string]manualfault.NodeCmInfo, rv string) {
			writtenRV = rv
		})
		defer p3.Reset()

		updateManualCm(nil)
		convey.So(writtenRV, convey.ShouldEqual, "")
	})

	convey.Convey("test func updateManualCm deep copy error", t, func() {
		p1 := gomonkey.ApplyMethodReturn(&manualfault.FaultCmInfo, "DeepCopy", nil, testErr)
		defer p1.Reset()

		var called bool
		p2 := gomonkey.ApplyFunc(manualfault.UpdateOrCreateManualCm, func(map[string]manualfault.NodeCmInfo, string) {
			called = true
		})
		defer p2.Reset()

		updateManualCm(nil)
		convey.So(called, convey.ShouldBeFalse)
	})
}

func TestProcess(t *testing.T) {
	const processInterval = 500 * time.Millisecond
	convey.Convey("test func ProcessManuSep success", t, func() {
		var hasExecuted bool
		var p1 = gomonkey.ApplyFunc(manualfault.UpdateOrCreateManualCm, func(map[string]manualfault.NodeCmInfo, string) {
			hasExecuted = true
			return
		})
		defer p1.Reset()
		ctx, cancel := context.WithCancel(context.TODO())
		go ProcessManuSep(ctx)
		time.Sleep(processInterval)
		cancel()
		convey.So(hasExecuted, convey.ShouldBeFalse)
	})
}

func TestLoadManualCmInfo(t *testing.T) {
	convey.Convey("test func LoadManualCmInfo success", t, func() {
		manualfault.InitFaultCmInfo()
		p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, getDemoCm(), nil)
		defer p1.Reset()
		LoadManualCmInfo()
		faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
		convey.So(err, convey.ShouldBeNil)
		convey.So(faultCmInfo, convey.ShouldResemble, getDemoNodeInfo())
	})
	convey.Convey("test func LoadManualCmInfo success, cm is nil", t, func() {
		manualfault.InitFaultCmInfo()
		p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, nil, nil)
		defer p1.Reset()
		LoadManualCmInfo()
		faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(faultCmInfo), convey.ShouldEqual, 0)
	})
	convey.Convey("test func LoadManualCmInfo failed, get manual cm failed", t, func() {
		manualfault.InitFaultCmInfo()
		p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, nil, testErr)
		defer p1.Reset()
		LoadManualCmInfo()
		faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(faultCmInfo), convey.ShouldEqual, len0)
	})
	convey.Convey("test func LoadManualCmInfo failed, parse cm info failed", t, func() {
		manualfault.InitFaultCmInfo()
		p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, getDemoCm(), nil).
			ApplyFuncReturn(manualfault.ParseManualCm, nil, testErr)
		defer p1.Reset()
		LoadManualCmInfo()
		faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(faultCmInfo), convey.ShouldEqual, len0)
	})
}

func TestFilterSilentFaultNodes(t *testing.T) {
	convey.Convey("test func filterSilentFaultNodes", t, func() {
		cmInfo := map[string]manualfault.NodeCmInfo{
			node1: {
				Total: []string{dev1, dev2},
				Detail: map[string][]manualfault.DevCmInfo{
					dev1: {{FaultCode: code1, FaultLevel: constant.ManuallySeparateNPU}},
					dev2: {{FaultCode: code1, FaultLevel: constant.SilentFault}},
				},
			},
			node2: {
				Total: []string{dev1},
				Detail: map[string][]manualfault.DevCmInfo{
					dev1: {{FaultCode: code1, FaultLevel: constant.SilentFault}},
				},
			},
		}
		filtered := filterSilentFaultNodes(cmInfo)
		convey.So(len(filtered), convey.ShouldEqual, len1)
		convey.So(len(filtered[node1].Detail), convey.ShouldEqual, len1)
		convey.So(filtered[node1].Detail[dev1][0].FaultLevel, convey.ShouldEqual, constant.ManuallySeparateNPU)
	})
}

func TestHandleSilentFaultEvent(t *testing.T) {
	convey.Convey("test func handleSilentFaultEvent", t, func() {
		conf.SetSilentFaultPolicy(conf.SilentFaultPolicy{Enabled: true})
		silentfault.ResetCache()
		p := gomonkey.ApplyFuncReturn(getDeviceType, "Ascend910")
		defer p.Reset()

		handleSilentFaultEvent(publicfault.SilentFaultEvent{
			Assertion: constant.AssertionOccur,
			NodeName:  node1,
			Resource:  constant.SilentFaultResource,
			FaultId:   "silent-fault-node1",
			DevIds:    []int32{0, 1},
		})
		convey.So(silentfault.SilentFaultCmInfo.Len(), convey.ShouldEqual, len1)
		info, _ := silentfault.SilentFaultCmInfo.Get(node1)
		convey.So(info.DevList[0], convey.ShouldEqual, "Ascend910-0")
		convey.So(info.DevList[1], convey.ShouldEqual, "Ascend910-1")

		handleSilentFaultEvent(publicfault.SilentFaultEvent{
			Assertion: constant.AssertionRecover,
			NodeName:  node1,
			Resource:  constant.SilentFaultResource,
			FaultId:   "silent-fault-node1",
		})
		convey.So(silentfault.SilentFaultCmInfo.Len(), convey.ShouldEqual, len0)

		silentfault.ResetCache()
	})
}

func TestBuildSilentFaultRecoverMsg(t *testing.T) {
	convey.Convey("test func buildSilentFaultRecoverMsg", t, func() {
		msg := buildSilentFaultRecoverMsg(constant.SilentFaultResource, "silent-fault-node1", node1, time.Now())
		convey.So(msg.Resource, convey.ShouldEqual, constant.SilentFaultResource)
		convey.So(len(msg.Faults), convey.ShouldEqual, 1)
		convey.So(msg.Faults[0].FaultId, convey.ShouldEqual, "silent-fault-node1")
		convey.So(msg.Faults[0].Assertion, convey.ShouldEqual, constant.AssertionRecover)
		convey.So(msg.Faults[0].FaultCode, convey.ShouldEqual, constant.SilentFaultCode)
	})
}

func TestHasSilentEntry(t *testing.T) {
	convey.Convey("test func hasSilentEntry", t, func() {
		cmInfo := map[string]manualfault.NodeCmInfo{
			node1: {Detail: map[string][]manualfault.DevCmInfo{
				dev1: {{FaultLevel: constant.SilentFault}},
			}},
		}
		convey.So(hasSilentEntry(cmInfo[node1]), convey.ShouldBeTrue)

		cmInfo2 := map[string]manualfault.NodeCmInfo{
			node1: {Detail: map[string][]manualfault.DevCmInfo{
				dev1: {{FaultLevel: constant.ManuallySeparateNPU}},
			}},
		}
		convey.So(hasSilentEntry(cmInfo2[node1]), convey.ShouldBeFalse)
		convey.So(hasSilentEntry(manualfault.NodeCmInfo{}), convey.ShouldBeFalse)
	})
}

func TestDiffManuallyDeletedSilent(t *testing.T) {
	convey.Convey("test func diffManuallyDeletedSilent node removed", t, func() {
		manualfault.LastCmInfo = map[string]manualfault.NodeCmInfo{
			node1: {Total: []string{dev1}, Detail: map[string][]manualfault.DevCmInfo{
				dev1: {{FaultLevel: constant.SilentFault}},
			}},
			node2: {Total: []string{dev1}, Detail: map[string][]manualfault.DevCmInfo{
				dev1: {{FaultLevel: constant.SilentFault}},
			}},
		}

		// only node1's silent device remains in cm; node2 is treated as deleted
		cmInfo := map[string]manualfault.NodeCmInfo{
			node1: {Total: []string{dev1}, Detail: map[string][]manualfault.DevCmInfo{
				dev1: {{FaultLevel: constant.SilentFault}},
			}},
		}
		p := gomonkey.ApplyFuncReturn(manualfault.ParseManualCm, cmInfo, nil)
		defer p.Reset()

		deleted := diffManuallyDeletedSilent(&v1.ConfigMap{})
		convey.So(len(deleted), convey.ShouldEqual, len1)
		convey.So(deleted[0], convey.ShouldEqual, node2)

		manualfault.LastCmInfo = nil
	})

	convey.Convey("test func diffManuallyDeletedSilent all cards deleted triggers release", t, func() {
		manualfault.LastCmInfo = map[string]manualfault.NodeCmInfo{
			node1: {Total: []string{dev1, dev2}, Detail: map[string][]manualfault.DevCmInfo{
				dev1: {{FaultLevel: constant.SilentFault}},
				dev2: {{FaultLevel: constant.SilentFault}},
			}},
		}

		// all cards removed from Total (Detail left untouched)
		cmInfo := map[string]manualfault.NodeCmInfo{
			node1: {Total: []string{}, Detail: map[string][]manualfault.DevCmInfo{
				dev1: {{FaultLevel: constant.SilentFault}},
				dev2: {{FaultLevel: constant.SilentFault}},
			}},
		}
		p := gomonkey.ApplyFuncReturn(manualfault.ParseManualCm, cmInfo, nil)
		defer p.Reset()

		deleted := diffManuallyDeletedSilent(&v1.ConfigMap{})
		convey.So(len(deleted), convey.ShouldEqual, len1)
		convey.So(deleted[0], convey.ShouldEqual, node1)

		manualfault.LastCmInfo = nil
	})

	convey.Convey("test func diffManuallyDeletedSilent partial cards deleted keeps isolation", t, func() {
		manualfault.LastCmInfo = map[string]manualfault.NodeCmInfo{
			node1: {Total: []string{dev1, dev2}, Detail: map[string][]manualfault.DevCmInfo{
				dev1: {{FaultLevel: constant.SilentFault}},
				dev2: {{FaultLevel: constant.SilentFault}},
			}},
		}

		// only dev2 removed from Total; node is still partially silent, so no release
		cmInfo := map[string]manualfault.NodeCmInfo{
			node1: {Total: []string{dev1}, Detail: map[string][]manualfault.DevCmInfo{
				dev1: {{FaultLevel: constant.SilentFault}},
				dev2: {{FaultLevel: constant.SilentFault}},
			}},
		}
		p := gomonkey.ApplyFuncReturn(manualfault.ParseManualCm, cmInfo, nil)
		defer p.Reset()

		deleted := diffManuallyDeletedSilent(&v1.ConfigMap{})
		convey.So(deleted, convey.ShouldBeEmpty)

		manualfault.LastCmInfo = nil
	})

	convey.Convey("test func diffManuallyDeletedSilent not written yet", t, func() {
		manualfault.LastCmInfo = nil
		deleted := diffManuallyDeletedSilent(&v1.ConfigMap{})
		convey.So(len(deleted), convey.ShouldEqual, len0)
	})
}

func TestReleaseSilentFault(t *testing.T) {
	convey.Convey("test func releaseSilentFault switch off", t, func() {
		conf.SetSilentFaultPolicy(conf.SilentFaultPolicy{Enabled: false})
		called := false
		p := gomonkey.ApplyFunc(sendSilentFaultRecover, func(string, string) { called = true })
		defer p.Reset()
		releaseSilentFault(nil)
		convey.So(called, convey.ShouldBeFalse)
	})

	convey.Convey("test func releaseSilentFault auto release", t, func() {
		conf.SetSilentFaultPolicy(conf.SilentFaultPolicy{
			Enabled: true,
			Release: struct {
				FaultFreeSecond int `yaml:"fault_free_seconds"`
			}{FaultFreeSecond: 100},
		})
		silentfault.ResetCache()
		// rebuild with a historical LastSeparateTime so the real Expired hits this node
		silentfault.SilentFaultCmInfo.Restore(node1, "silent-fault-node1", constant.SilentFaultResource, []string{dev1}, 1)

		var released []string
		p1 := gomonkey.ApplyFunc(sendSilentFaultRecover, func(node, reason string) {
			released = append(released, reason)
		})
		defer p1.Reset()
		p2 := gomonkey.ApplyFuncReturn(diffManuallyDeletedSilent, []string{})
		defer p2.Reset()

		releaseSilentFault(nil)
		convey.So(len(released), convey.ShouldEqual, len1)
		convey.So(released[0], convey.ShouldEqual, "auto release")
		silentfault.ResetCache()
	})

	convey.Convey("test func releaseSilentFault manual removal triggers recover", t, func() {
		conf.SetSilentFaultPolicy(conf.SilentFaultPolicy{Enabled: true})
		silentfault.ResetCache()

		var released []string
		p1 := gomonkey.ApplyFunc(sendSilentFaultRecover, func(node, reason string) {
			released = append(released, reason)
		})
		defer p1.Reset()
		p2 := gomonkey.ApplyFunc(diffManuallyDeletedSilent, func(*v1.ConfigMap) []string {
			return []string{node1}
		})
		defer p2.Reset()

		releaseSilentFault(&v1.ConfigMap{})
		convey.So(len(released), convey.ShouldEqual, len1)
		convey.So(released[0], convey.ShouldEqual, "manually deleted")
		silentfault.ResetCache()
	})

	convey.Convey("test func releaseSilentFault nil cm skips manual removal", t, func() {
		conf.SetSilentFaultPolicy(conf.SilentFaultPolicy{Enabled: true})
		silentfault.ResetCache()

		var released []string
		p1 := gomonkey.ApplyFunc(sendSilentFaultRecover, func(node, reason string) {
			released = append(released, reason)
		})
		defer p1.Reset()
		p2 := gomonkey.ApplyFuncReturn(diffManuallyDeletedSilent, []string{})
		defer p2.Reset()

		releaseSilentFault(nil)
		convey.So(len(released), convey.ShouldEqual, len0)
		silentfault.ResetCache()
	})
}

func TestSendSilentFaultRecover(t *testing.T) {
	convey.Convey("test func sendSilentFaultRecover", t, func() {
		silentfault.ResetCache()
		silentfault.SilentFaultCmInfo.Upsert(node1, "silent-fault-node1", constant.SilentFaultResource, []string{dev1})

		var count int
		p := gomonkey.ApplyFunc(publicfault.PubFaultCollector, func(*api.PubFaultInfo) error {
			count++
			return nil
		})
		defer p.Reset()

		sendSilentFaultRecover(node1, "test")
		convey.So(count, convey.ShouldEqual, len1)
		silentfault.ResetCache()
	})

	convey.Convey("test func sendSilentFaultRecover node not exist", t, func() {
		silentfault.ResetCache()
		called := false
		p := gomonkey.ApplyFunc(publicfault.PubFaultCollector, func(*api.PubFaultInfo) error {
			called = true
			return nil
		})
		defer p.Reset()
		sendSilentFaultRecover(node2, "test")
		convey.So(called, convey.ShouldBeFalse)
	})
}

func TestGetDeviceType(t *testing.T) {
	convey.Convey("get device type success", t, func() {
		p := gomonkey.ApplyFuncReturn(faultmanager.QueryDeviceInfoToReport, map[string]*constant.AdvanceDeviceFaultCm{
			node1: {DeviceType: "Ascend910"},
		})
		defer p.Reset()
		convey.So(getDeviceType(node1), convey.ShouldEqual, "Ascend910")
	})
	convey.Convey("get device type node not exist", t, func() {
		p := gomonkey.ApplyFuncReturn(faultmanager.QueryDeviceInfoToReport, map[string]*constant.AdvanceDeviceFaultCm{})
		defer p.Reset()
		convey.So(getDeviceType(node1), convey.ShouldEqual, "")
	})
	convey.Convey("get device type device cm nil", t, func() {
		p := gomonkey.ApplyFuncReturn(faultmanager.QueryDeviceInfoToReport, map[string]*constant.AdvanceDeviceFaultCm{
			node1: nil,
		})
		defer p.Reset()
		convey.So(getDeviceType(node1), convey.ShouldEqual, "")
	})
}

func TestResolveDevNames(t *testing.T) {
	convey.Convey("resolveDevNames builds names from device type", t, func() {
		p := gomonkey.ApplyFuncReturn(getDeviceType, "Ascend910")
		defer p.Reset()

		names := resolveDevNames("n1", []int32{0, 1})
		convey.So(names, convey.ShouldResemble, []string{"Ascend910-0", "Ascend910-1"})
	})

	convey.Convey("resolveDevNames degrades to bare id when device type empty", t, func() {
		p := gomonkey.ApplyFuncReturn(getDeviceType, "")
		defer p.Reset()

		convey.So(resolveDevNames("n1", []int32{0}), convey.ShouldResemble, []string{"0"})
	})
}
