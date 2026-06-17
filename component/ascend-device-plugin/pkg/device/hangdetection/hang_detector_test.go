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

package hangdetection

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager"
	npuCommon "ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"
)

const (
	mockLogicID = int32(0)
	aicoreUsage = 80
	memUsage    = 60
	pktNum      = 100
	cpuTime     = 100
	clkTck1     = 100
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		panic(err)
	}
}

func resetHangState() {
	hangStateMapMu.Lock()
	hangStateMap = make(map[int32]*HangState)
	hangStateMapMu.Unlock()

	npuFaultCacheMu.Lock()
	npuFaultCache = make([]*npuCommon.DevFaultInfo, 0)
	npuFaultCacheMu.Unlock()
}

func newTestHangDetector() *HangDetector {
	return &HangDetector{
		dmgr:            &devmanager.DeviceManager{},
		npuDevPortInfos: make(map[int][]int),
	}
}

// TestDetectNPU for test detectNPU
func TestDetectNPU(t *testing.T) {
	convey.Convey("test detectNPU", t, func() {
		resetHangState()
		hd := newTestHangDetector()
		convey.Convey("when get process info failed, should call npuHangEventDisappear", func() {
			patches := gomonkey.ApplyMethod(&devmanager.DeviceManager{}, "GetDevProcessInfo",
				func(_ *devmanager.DeviceManager, _ int32) (*npuCommon.DevProcessInfo, error) {
					return nil, fmt.Errorf("get process info failed")
				})
			defer patches.Reset()
			hd.detectNPU(mockLogicID)
			state := hd.getOrCreateHangState(mockLogicID)
			convey.So(state.HangCount, convey.ShouldEqual, 0)
		})
		convey.Convey("when procNum is 0, should call npuHangEventDisappear", func() {
			patches := gomonkey.ApplyMethod(&devmanager.DeviceManager{}, "GetDevProcessInfo",
				func(_ *devmanager.DeviceManager, _ int32) (*npuCommon.DevProcessInfo, error) {
					return &npuCommon.DevProcessInfo{ProcNum: 0}, nil
				})
			defer patches.Reset()
			hd.detectNPU(mockLogicID)
			state := hd.getOrCreateHangState(mockLogicID)
			convey.So(state.HangCount, convey.ShouldEqual, 0)
		})
		convey.Convey("when procNum > 0 and hang condition met after two rounds, should call npuHangEventOccur", func() {
			patches := gomonkey.ApplyPrivateMethod(hd, "collectMetrics",
				func(_ int32, _ *npuCommon.DevProcessInfo, metrics *HangMetrics) {}).
				ApplyPrivateMethod(hd, "isHangConditionMet",
					func(_ int32, _ *HangMetrics) bool { return true }).
				ApplyFunc(common.DoSaveDevFaultInfo, func(_ npuCommon.DevFaultInfo, _ bool) {}).
				ApplyMethod(&devmanager.DeviceManager{}, "GetDevProcessInfo",
					func(_ *devmanager.DeviceManager, _ int32) (*npuCommon.DevProcessInfo, error) {
						return &npuCommon.DevProcessInfo{ProcNum: 1}, nil
					})
			defer patches.Reset()
			hd.detectNPU(mockLogicID)
			hd.detectNPU(mockLogicID)
			state := hd.getOrCreateHangState(mockLogicID)
			convey.So(state.HangCount, convey.ShouldBeGreaterThanOrEqualTo, 1)
		})
	})
}

// TestCollectMetrics for test collectMetrics
func TestCollectMetrics(t *testing.T) {
	convey.Convey("test collectMetrics", t, func() {
		hd := newTestHangDetector()

		convey.Convey("02-when metrics is not nil, should fill all fields", func() {
			calledTimes := 0
			patches := gomonkey.ApplyPrivateMethod(hd, "collectUtilization",
				func(int32, *HangMetrics) { calledTimes++ }).
				ApplyPrivateMethod(hd, "collectMemoryUsage",
					func(int32, *HangMetrics) { calledTimes++ }).
				ApplyPrivateMethod(hd, "collectTraffic",
					func(int32, *HangMetrics) { calledTimes++ }).
				ApplyPrivateMethod(hd, "collectCPUTime",
					func(*npuCommon.DevProcessInfo, *HangMetrics) { calledTimes++ })
			defer patches.Reset()
			hd.collectMetrics(mockLogicID, &npuCommon.DevProcessInfo{ProcNum: 1}, &HangMetrics{})
			const expectedCalledTimes = 4
			convey.So(calledTimes, convey.ShouldEqual, expectedCalledTimes)
		})
	})
}

// TestCollectUtilization for test collectUtilization
func TestCollectUtilization(t *testing.T) {
	convey.Convey("test collectUtilization", t, func() {
		hd := newTestHangDetector()

		convey.Convey("01-when get utilization success, should set AICoreUsage", func() {
			patches := gomonkey.ApplyMethod(&devmanager.DeviceManager{}, "GetDeviceUtilizationRate",
				func(_ *devmanager.DeviceManager, _ int32, _ npuCommon.DeviceType) (uint32, error) {
					return aicoreUsage, nil
				})
			defer patches.Reset()

			metrics := &HangMetrics{}
			hd.collectUtilization(mockLogicID, metrics)
			convey.So(metrics.AICoreUsage, convey.ShouldEqual, aicoreUsage)
		})
	})
}

// TestCollectMemoryUsage for test collectMemoryUsage
func TestCollectMemoryUsage(t *testing.T) {
	convey.Convey("test collectMemoryUsage", t, func() {
		hd := newTestHangDetector()

		convey.Convey("when get hbm info success, should set MemoryUsage", func() {
			patches := gomonkey.ApplyMethod(&devmanager.DeviceManager{}, "GetDeviceUtilizationRate",
				func(_ *devmanager.DeviceManager, _ int32, _ npuCommon.DeviceType) (uint32, error) {
					return memUsage, nil
				})
			defer patches.Reset()

			metrics := &HangMetrics{}
			hd.collectMemoryUsage(mockLogicID, metrics)
			convey.So(metrics.MemoryUsage, convey.ShouldEqual, memUsage)
		})
	})
}

// TestCollectTraffic for test collectTraffic
func TestCollectTraffic(t *testing.T) {
	convey.Convey("test collectTraffic", t, func() {
		hd := newTestHangDetector()
		origCardType := common.ParamOption.RealCardType
		defer func() {
			common.ParamOption.RealCardType = origCardType
		}()

		convey.Convey("when card type is Ascend910A5, should collect UB traffic", func() {
			common.ParamOption.RealCardType = api.Ascend910A5
			called := false
			patches := gomonkey.ApplyPrivateMethod(hd, "collectUBTraffic",
				func(int32, *HangMetrics) { called = true })
			defer patches.Reset()
			hd.collectTraffic(mockLogicID, nil)
			convey.So(called, convey.ShouldBeTrue)
		})

		convey.Convey("when card type is not Ascend910A5, should collect RoCE traffic", func() {
			common.ParamOption.RealCardType = api.Ascend910A3
			called := false
			patches := gomonkey.ApplyPrivateMethod(hd, "collectRoCETraffic",
				func(int32, *HangMetrics) { called = true })
			defer patches.Reset()
			hd.collectTraffic(mockLogicID, nil)
			convey.So(called, convey.ShouldBeTrue)
		})
	})
}

// TestCollectRoCETraffic for test collectRoCETraffic
func TestCollectRoCETraffic(t *testing.T) {
	convey.Convey("test collectRoCETraffic", t, func() {
		hd := newTestHangDetector()
		convey.Convey("when get roce stat success, should set RoCE metrics", func() {
			patches := gomonkey.ApplyMethod(hd.dmgr, "GetPhysicIDFromLogicID",
				func(_ *devmanager.DeviceManager, _ int32) (int32, error) {
					return 0, nil
				}).ApplyFuncReturn(hccn.GetNPUStatInfo, map[string]int{
				roceTxAllPktNum: pktNum,
				roceRxAllPktNum: pktNum,
			}, nil)
			defer patches.Reset()

			metrics := &HangMetrics{}
			hd.collectRoCETraffic(mockLogicID, metrics)
			convey.So(metrics.RoceTxPkts, convey.ShouldEqual, pktNum)
			convey.So(metrics.RoceRxPkts, convey.ShouldEqual, pktNum)
		})
	})
}

// TestCollectUBTraffic for test collectUBTraffic
func TestCollectUBTraffic(t *testing.T) {
	convey.Convey("test collectUBTraffic", t, func() {
		convey.Convey("when port info is empty, UB metrics should remain 0", func() {
			hd := newTestHangDetector()
			hd.npuDevPortInfos = map[int][]int{}

			metrics := &HangMetrics{}
			hd.collectUBTraffic(mockLogicID, metrics)
			convey.So(metrics.UBTxFlits, convey.ShouldEqual, 0)
			convey.So(metrics.UBRxFlits, convey.ShouldEqual, 0)
		})

		convey.Convey("when get UB stat success, should accumulate tx and rx", func() {
			hd := newTestHangDetector()
			hd.npuDevPortInfos = map[int][]int{0: {1}}
			patches := gomonkey.ApplyFuncReturn(hccn.GetNPUUbStatInfo, map[string]string{
				txBusiFlitNum: fmt.Sprintf("%d", pktNum),
				rxBusiFlitNum: fmt.Sprintf("%d", pktNum),
			}, nil)
			defer patches.Reset()

			metrics := &HangMetrics{}
			hd.collectUBTraffic(mockLogicID, metrics)
			convey.So(metrics.UBTxFlits, convey.ShouldEqual, pktNum)
			convey.So(metrics.UBRxFlits, convey.ShouldEqual, pktNum)
		})
	})
}

// TestCollectCPUTime for test collectCPUTime
func TestCollectCPUTime(t *testing.T) {
	convey.Convey("test collectCPUTime", t, func() {
		hd := newTestHangDetector()

		convey.Convey("01-when get CPU time success, should accumulate total", func() {
			patches := gomonkey.ApplyFunc(getProcessCPUTime, func(pid int32) (int64, error) {
				return cpuTime, nil
			})
			defer patches.Reset()

			procInfo := &npuCommon.DevProcessInfo{
				ProcNum:      1,
				DevProcArray: []npuCommon.DevProcInfo{{Pid: 1}},
			}
			metrics := &HangMetrics{}
			hd.collectCPUTime(procInfo, metrics)
			convey.So(metrics.CPUTime, convey.ShouldEqual, cpuTime)
		})
	})
}

// TestIsHangConditionMet for test isHangConditionMet
func TestIsHangConditionMet(t *testing.T) {
	convey.Convey("test isHangConditionMet", t, func() {
		logicID := int32(0)
		hd := newTestHangDetector()
		mockCardType := gomonkey.ApplyGlobalVar(&common.ParamOption.RealCardType, api.Ascend910A3)
		defer mockCardType.Reset()

		convey.Convey("01-when no last metrics, should return false", func() {
			resetHangState()
			curMetrics := &HangMetrics{ProcessNum: 1}
			result := hd.isHangConditionMet(logicID, curMetrics)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("02-when all hang conditions met, should return true", func() {
			resetHangState()

			state := hd.getOrCreateHangState(logicID)
			state.Metrics = &HangMetrics{AICoreUsage: 0, MemoryUsage: 5, CPUTime: cpuTime, RoceTxPkts: 50, RoceRxPkts: 50, ProcessNum: 1}

			curMetrics := &HangMetrics{AICoreUsage: 4, MemoryUsage: 5, CPUTime: cpuTime, RoceTxPkts: 60, RoceRxPkts: 60, ProcessNum: 1}
			result := hd.isHangConditionMet(logicID, curMetrics)
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

// TestNpuHangEventOccur for test npuHangEventOccur
func TestNpuHangEventOccur(t *testing.T) {
	convey.Convey("test npuHangEventOccur", t, func() {
		hd := newTestHangDetector()

		convey.Convey("when HangCount reaches DetectDuration, should report hang fault", func() {
			resetHangState()
			patches := gomonkey.ApplyFuncReturn(common.GetHangDetectionThreshold, common.HangThreshold{
				DetectDuration: 3,
			}).ApplyFunc(common.DoSaveDevFaultInfo, func(_ npuCommon.DevFaultInfo, _ bool) {})
			defer patches.Reset()
			var hangCount = 3
			for i := 0; i < hangCount; i++ {
				hd.npuHangEventOccur(mockLogicID, &HangMetrics{})
			}
			state := hd.getOrCreateHangState(mockLogicID)
			convey.So(state.HangCount, convey.ShouldEqual, hangCount)
			convey.So(state.IsFault, convey.ShouldBeTrue)
		})
	})
}

// TestNpuHangEventDisappear for test npuHangEventDisappear
func TestNpuHangEventDisappear(t *testing.T) {
	convey.Convey("test npuHangEventDisappear", t, func() {
		hd := newTestHangDetector()

		convey.Convey("when previously reported, should report recover and reset", func() {
			resetHangState()
			patches := gomonkey.ApplyFuncReturn(common.GetHangDetectionThreshold, common.HangThreshold{
				DetectDuration: 1,
			}).ApplyFunc(common.DoSaveDevFaultInfo, func(_ npuCommon.DevFaultInfo, _ bool) {})
			defer patches.Reset()

			hd.npuHangEventOccur(mockLogicID, &HangMetrics{})
			state := hd.getOrCreateHangState(mockLogicID)
			convey.So(state.IsFault, convey.ShouldBeTrue)

			hd.npuHangEventDisappear(mockLogicID, &HangMetrics{})
			state = hd.getOrCreateHangState(mockLogicID)
			convey.So(state.HangCount, convey.ShouldEqual, 0)
			convey.So(state.IsFault, convey.ShouldBeFalse)
		})
	})
}

// TestGetOrCreateHangState for test getOrCreateHangState
func TestGetOrCreateHangState(t *testing.T) {
	convey.Convey("test getOrCreateHangState", t, func() {
		hd := newTestHangDetector()

		convey.Convey("when state does not exist, should create new state", func() {
			resetHangState()
			state := hd.getOrCreateHangState(mockLogicID)
			convey.So(state, convey.ShouldNotBeNil)
			convey.So(state.LogicID, convey.ShouldEqual, mockLogicID)
			convey.So(state.HangCount, convey.ShouldEqual, 0)
		})
	})
}

// TestGetProcessCPUTime for test getProcessCPUTime
func TestGetProcessCPUTime(t *testing.T) {
	convey.Convey("test getProcessCPUTime", t, func() {
		convey.Convey("01-when get current process CPU time, should succeed", func() {
			utime, stime := int64(11), int64(12)
			patches := gomonkey.ApplyFuncReturn(utils.IsExist, true).
				ApplyFunc(utils.LoadFile, func(_ string) ([]byte, error) {
					return []byte("1234 (test) S 1 2 3 4 5 6 7 8 9 10 11 12 100 200 13 14 15"), nil
				})
			defer patches.Reset()

			cpuTime, err := getProcessCPUTime(int32(os.Getpid()))
			convey.So(err, convey.ShouldBeNil)
			convey.So(cpuTime, convey.ShouldEqual, (utime+stime)/clkTck)
		})

		convey.Convey("02-when pid does not exist, should return error", func() {
			_, err := getProcessCPUTime(0)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestGetNpuDevNetPortInfos for test getNpuDevNetPortInfos
func TestGetNpuDevNetPortInfos(t *testing.T) {
	convey.Convey("test getNpuDevNetPortInfos", t, func() {
		hd := newTestHangDetector()
		origCardType := common.ParamOption.RealCardType
		defer func() {
			common.ParamOption.RealCardType = origCardType
		}()

		convey.Convey("when card type is not Ascend910A5, should return nil", func() {
			common.ParamOption.RealCardType = api.Ascend910A3
			result, err := hd.getNpuDevNetPortInfos()
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("when card type is Ascend910A5 and GetNpuDevNetPortInfo success, should return port info", func() {
			common.ParamOption.RealCardType = api.Ascend910A5
			patches := gomonkey.ApplyMethod(&devmanager.DeviceManager{}, "GetDeviceList",
				func(_ *devmanager.DeviceManager) (int32, []int32, error) {
					return 1, []int32{0}, nil
				}).ApplyFuncReturn(hccn.GetNpuDevNetPortInfo, map[int][]int{0: {4, 5}}, nil)
			defer patches.Reset()

			result, err := hd.getNpuDevNetPortInfos()
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

// TestReportHangFault for test reportHangFault
func TestReportHangFault(t *testing.T) {
	convey.Convey("test reportHangFault", t, func() {
		hd := newTestHangDetector()

		convey.Convey("should append fault with FaultOccur to cache", func() {
			resetHangState()
			hd.reportHangFault(mockLogicID)
			faultInfos := GetAndCleanAllHangFaultCache()
			convey.So(len(faultInfos), convey.ShouldEqual, 1)
			convey.So(faultInfos[0].LogicID, convey.ShouldEqual, mockLogicID)
			convey.So(faultInfos[0].Assertion, convey.ShouldEqual, npuCommon.FaultOccur)
			convey.So(faultInfos[0].EventID, convey.ShouldEqual, npuCommon.HangFaultCode)
		})
	})
}

// TestReportHangRecover for test reportHangRecover
func TestReportHangRecover(t *testing.T) {
	convey.Convey("test reportHangRecover", t, func() {
		hd := newTestHangDetector()

		convey.Convey("should append fault with FaultRecover to cache", func() {
			resetHangState()
			hd.reportHangRecover(mockLogicID)
			faultInfos := GetAndCleanAllHangFaultCache()
			convey.So(len(faultInfos), convey.ShouldEqual, 1)
			convey.So(faultInfos[0].LogicID, convey.ShouldEqual, mockLogicID)
			convey.So(faultInfos[0].Assertion, convey.ShouldEqual, npuCommon.FaultRecover)
			convey.So(faultInfos[0].EventID, convey.ShouldEqual, npuCommon.HangFaultCode)
		})
	})
}

func buildAuxvDataWithClkTck(value uint64) []byte {
	buf := new(bytes.Buffer)
	// AT_CLKTCK entry: type=17, value=<clkTck>
	binary.Write(buf, binary.NativeEndian, uint64(17))
	binary.Write(buf, binary.NativeEndian, value)
	// AT_NULL terminator: type=0, value=0
	binary.Write(buf, binary.NativeEndian, uint64(0))
	binary.Write(buf, binary.NativeEndian, uint64(0))
	return buf.Bytes()
}

func buildAuxvDataWithOutClkTck() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.NativeEndian, uint64(0))
	binary.Write(buf, binary.NativeEndian, uint64(0))
	return buf.Bytes()
}

// TestSetClkTck for test setClkTck
func TestSetClkTckAndGetSysClockTicks(t *testing.T) {
	convey.Convey("test getSysClockTicks", t, func() {
		convey.Convey("when auxv contains AT_CLKTCK entry, should return its value", func() {
			patches := gomonkey.ApplyFunc(utils.LoadFile, func(_ string) ([]byte, error) {
				return buildAuxvDataWithClkTck(uint64(clkTck1)), nil
			})
			defer patches.Reset()
			clkTck = 0
			setClkTck()
			convey.So(clkTck, convey.ShouldEqual, int64(clkTck1))
		})

		convey.Convey("when auxv does not contain AT_CLKTCK entry, should return error", func() {
			patches := gomonkey.ApplyFunc(utils.LoadFile, func(_ string) ([]byte, error) {
				return buildAuxvDataWithOutClkTck(), nil
			})
			defer patches.Reset()
			clkTck = clkTck1
			setClkTck()
			convey.So(clkTck, convey.ShouldEqual, 0)
		})
	})
}

// TestGetAndCleanAllHangFaultCache for test GetAndCleanAllHangFaultCache
func TestGetAndCleanAllHangFaultCache(t *testing.T) {
	convey.Convey("test GetAndCleanAllHangFaultCache", t, func() {
		convey.Convey("when cache is empty, should return empty slice", func() {
			resetHangState()
			faultInfos := GetAndCleanAllHangFaultCache()
			convey.So(len(faultInfos), convey.ShouldEqual, 0)
		})

		convey.Convey("when cache has entries, should return all and clear cache", func() {
			resetHangState()
			hd := newTestHangDetector()
			hd.reportHangFault(mockLogicID)
			hd.reportHangRecover(mockLogicID)

			faultInfos := GetAndCleanAllHangFaultCache()
			convey.So(len(faultInfos), convey.ShouldEqual, 2)

			faultInfos2 := GetAndCleanAllHangFaultCache()
			convey.So(len(faultInfos2), convey.ShouldEqual, 0)
		})
	})
}

// TestRegisterLogicIDForProducer for test RegisterLogicIDForProducer
func TestRegisterLogicIDForProducer(t *testing.T) {
	convey.Convey("test RegisterLogicIDForProducer", t, func() {
		convey.Convey("should register logicID and snapshot returns it", func() {
			logicIdMap = sync.Map{}
			RegisterLogicIDForProducer(mockLogicID)
			logicIDs := snapshotRegisteredLogicIDs()
			convey.So(len(logicIDs), convey.ShouldEqual, 1)
			convey.So(logicIDs[0], convey.ShouldEqual, mockLogicID)
		})

		convey.Convey("should not duplicate logicID", func() {
			logicIdMap = sync.Map{}
			RegisterLogicIDForProducer(mockLogicID)
			RegisterLogicIDForProducer(mockLogicID)
			logicIDs := snapshotRegisteredLogicIDs()
			convey.So(len(logicIDs), convey.ShouldEqual, 1)
		})
	})
}

// TestStartHangDetectionProducer for test StartHangDetectionProducer
func TestStartHangDetectionProducer(t *testing.T) {
	convey.Convey("test StartHangDetectionProducer", t, func() {
		convey.Convey("when dmgr is nil, should return immediately", func() {
			called := false
			patches := gomonkey.ApplyFunc(runHangDetectionProducer,
				func(ctx context.Context) { called = true })
			defer patches.Reset()
			StartHangDetectionProducer(context.Background(), nil)
			convey.So(called, convey.ShouldBeFalse)
		})

		convey.Convey("when dmgr is not nil, should start producer", func() {
			hangDetector = &HangDetector{npuDevPortInfos: make(map[int][]int)}
			called := false
			patches := gomonkey.ApplyPrivateMethod(hangDetector, "getNpuDevNetPortInfos",
				func(*HangDetector) (map[int][]int, error) { return nil, fmt.Errorf("mock error") }).
				ApplyFuncReturn(common.LoadHangDetectionConfigFromFile).
				ApplyFunc(runHangDetectionProducer, func(ctx context.Context) {
					called = true
				})
			defer patches.Reset()
			StartHangDetectionProducer(context.Background(), &devmanager.DeviceManager{})
			convey.So(called, convey.ShouldBeTrue)
		})
	})
}

func TestRunHangDetectionProducer(t *testing.T) {
	convey.Convey("test runHangDetectionProducer", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		called := false
		secondCh := make(chan time.Time, 1)
		patches := gomonkey.ApplyPrivateMethod(hangDetector, "detectNPU",
			func(*HangDetector, int32) { called = true }).
			ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
				return &time.Ticker{C: secondCh}
			})
		defer patches.Reset()
		RegisterLogicIDForProducer(mockLogicID)
		const sleepTime = 100 * time.Millisecond
		go func() {
			secondCh <- time.Now()
			time.Sleep(sleepTime)
			cancel()
		}()
		runHangDetectionProducer(ctx)
		convey.So(called, convey.ShouldBeTrue)
	})
}
