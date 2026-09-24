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

package silentfault

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
)

func TestFaultTimes(t *testing.T) {
	convey.Convey("fault times", t, func() {
		f := constant.DeviceFault{FaultCode: "c", FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			"c": {FaultTime: 100000},
		}}
		convey.So(faultTimes(f, 999), convey.ShouldResemble, []int64{100000})

		f2 := constant.DeviceFault{FaultCode: "c", FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			"c": {FaultReceivedTime: 200000},
		}}
		convey.So(faultTimes(f2, 999), convey.ShouldResemble, []int64{200000})

		f3 := constant.DeviceFault{FaultCode: "c"}
		convey.So(faultTimes(f3, 999), convey.ShouldResemble, []int64{999})
	})

	convey.Convey("public fault second converted to millisecond", t, func() {
		f := constant.DeviceFault{FaultType: constant.PublicFaultType, FaultCode: "c",
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{"c": {FaultTime: 100}}}
		convey.So(faultTimes(f, 0), convey.ShouldResemble, []int64{100000})
	})

	convey.Convey("empty fault code key skipped", t, func() {
		f := constant.DeviceFault{FaultCode: "", FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			"":  {FaultTime: 100},
			"c": {FaultTime: 200000},
		}}
		convey.So(faultTimes(f, 999), convey.ShouldResemble, []int64{200000})
	})

	convey.Convey("is silent self fault", t, func() {
		f := constant.DeviceFault{FaultType: constant.PublicFaultType, FaultLevel: constant.SilentFault}
		convey.So(isSilentSelfFault(f), convey.ShouldBeTrue)
		f.FaultLevel = constant.SeparateNPU
		convey.So(isSilentSelfFault(f), convey.ShouldBeFalse)
	})
}

func TestFaultTimesMapFallback(t *testing.T) {
	convey.Convey("return all map entry times when fault code is missing", t, func() {
		f := constant.DeviceFault{FaultCode: "missing", FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			"other": {FaultTime: 300000},
		}}
		convey.So(faultTimes(f, 0), convey.ShouldResemble, []int64{300000})
	})
}

func TestRecordNodeFaults(t *testing.T) {
	convey.Convey("record node faults", t, func() {
		faultHardwareLog.Reset()
		FaultRecorderProcessor.recordNodeFaults(map[string]*constant.NodeInfo{
			"n1": {NodeInfoNoName: constant.NodeInfoNoName{FaultDevList: []*constant.FaultDev{{}}}, UpdateTime: 100000},
			"n2": {UpdateTime: 200000},
		})
		convey.So(faultHardwareLog.HasFaultIn("n1", 100000, 100000), convey.ShouldBeTrue)
		convey.So(faultHardwareLog.HasFaultIn("n2", 200000, 200000), convey.ShouldBeFalse)
	})
}

func TestRecordSwitchFaults(t *testing.T) {
	convey.Convey("record switch faults", t, func() {
		faultHardwareLog.Reset()
		FaultRecorderProcessor.recordSwitchFaults(map[string]*constant.SwitchInfo{
			constant.SwitchInfoPrefix + "n1": {SwitchFaultInfo: constant.SwitchFaultInfo{
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"k": {FaultTime: 300000},
				},
			}},
		})
		convey.So(faultHardwareLog.HasFaultIn("n1", 300000, 300000), convey.ShouldBeTrue)
	})
}

func TestRecordSwitchFaultsFallback(t *testing.T) {
	convey.Convey("record switch faults fallback to received/update time", t, func() {
		faultHardwareLog.Reset()
		FaultRecorderProcessor.recordSwitchFaults(map[string]*constant.SwitchInfo{
			constant.SwitchInfoPrefix + "n1": {SwitchFaultInfo: constant.SwitchFaultInfo{
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"a": {FaultReceivedTime: 200000},
					"b": {},
				},
				UpdateTime: 400000,
			}},
		})
		convey.So(faultHardwareLog.HasFaultIn("n1", 200000, 200000), convey.ShouldBeTrue)
		convey.So(faultHardwareLog.HasFaultIn("n1", 400000, 400000), convey.ShouldBeTrue)
	})
}

func TestRecordDeviceFaults(t *testing.T) {
	convey.Convey("record device faults increment", t, func() {
		faultHardwareLog.Reset()
		first := &constant.AdvanceDeviceFaultCm{UpdateTime: 1000, FaultDeviceList: map[string][]constant.DeviceFault{
			"d0": {{FaultCode: "c", FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{"c": {FaultTime: 100000}}}},
		}}
		FaultRecorderProcessor.recordDeviceFaults(map[string]*constant.AdvanceDeviceFaultCm{"n1": first})
		convey.So(faultHardwareLog.HasFaultIn("n1", 100000, 100000), convey.ShouldBeTrue)

		// same cm: no change
		FaultRecorderProcessor.recordDeviceFaults(map[string]*constant.AdvanceDeviceFaultCm{"n1": first})

		// new fault time: append
		second := &constant.AdvanceDeviceFaultCm{UpdateTime: 1000, FaultDeviceList: map[string][]constant.DeviceFault{
			"d0": {{FaultCode: "c", FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{"c": {FaultTime: 200000}}}},
		}}
		FaultRecorderProcessor.recordDeviceFaults(map[string]*constant.AdvanceDeviceFaultCm{"n1": second})
		convey.So(faultHardwareLog.HasFaultIn("n1", 200000, 200000), convey.ShouldBeTrue)
	})
}

func TestRecordAll(t *testing.T) {
	convey.Convey("record all device faults first round", t, func() {
		faultHardwareLog.Reset()
		FaultRecorderProcessor.recordAll("n1", map[string][]constant.DeviceFault{
			"d0": {{FaultCode: "c", FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{"c": {FaultTime: 100000}}}},
			"d1": {{FaultCode: "s", FaultType: constant.PublicFaultType, FaultLevel: constant.SilentFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{"s": {FaultTime: 999000}}}},
		}, 50)
		convey.So(faultHardwareLog.HasFaultIn("n1", 100000, 100000), convey.ShouldBeTrue)
		convey.So(faultHardwareLog.HasFaultIn("n1", 999000, 999000), convey.ShouldBeFalse)
	})
}

func TestFaultRecorderProcessor(t *testing.T) {
	convey.Convey("fault recorder process dispatch by type", t, func() {
		enableSilentFault()
		faultHardwareLog.Reset()

		FaultRecorderProcessor.Process(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
			AllConfigmap: map[string]*constant.AdvanceDeviceFaultCm{
				"n1": {UpdateTime: 1000, FaultDeviceList: map[string][]constant.DeviceFault{
					"d0": {{FaultCode: "c", FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{"c": {FaultTime: 100000}}}},
				}},
			},
		})
		convey.So(faultHardwareLog.HasFaultIn("n1", 100000, 100000), convey.ShouldBeTrue)

		faultHardwareLog.Reset()
		dev := &constant.AdvanceDeviceFaultCm{UpdateTime: 2000, FaultDeviceList: map[string][]constant.DeviceFault{
			"d0": {{FaultCode: "c", FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{"c": {FaultTime: 200000}}}},
		}}
		out := FaultRecorderProcessor.Process(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
			AllConfigmap: map[string]*constant.AdvanceDeviceFaultCm{"n1": dev},
		})
		convey.So(out, convey.ShouldNotBeNil)
	})
}

func TestProcessDispatchNodeSwitch(t *testing.T) {
	convey.Convey("process dispatch node and switch payload", t, func() {
		enableSilentFault()
		faultHardwareLog.Reset()
		FaultRecorderProcessor.Process(constant.OneConfigmapContent[*constant.NodeInfo]{
			AllConfigmap: map[string]*constant.NodeInfo{
				"n1": {NodeInfoNoName: constant.NodeInfoNoName{FaultDevList: []*constant.FaultDev{{}}}, UpdateTime: 100000},
			},
		})
		convey.So(faultHardwareLog.HasFaultIn("n1", 100000, 100000), convey.ShouldBeTrue)

		faultHardwareLog.Reset()
		FaultRecorderProcessor.Process(constant.OneConfigmapContent[*constant.SwitchInfo]{
			AllConfigmap: map[string]*constant.SwitchInfo{
				constant.SwitchInfoPrefix + "n1": {SwitchFaultInfo: constant.SwitchFaultInfo{
					FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{"k": {FaultTime: 300000}},
				}},
			},
		})
		convey.So(faultHardwareLog.HasFaultIn("n1", 300000, 300000), convey.ShouldBeTrue)
	})
}
