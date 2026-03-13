/* Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package metrics for general collector
package metrics

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/utils/logger"
)

func init() {
	logger.HwLogConfig = &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	logger.InitLogger("Prometheus")
	initChain()
}

// TestInitDesc tests the initDesc function
func TestInitDesc(t *testing.T) {
	convey.Convey("Test initDesc function", t, func() {
		// Create a channel with sufficient buffer
		ch := make(chan *prometheus.Desc, 10)

		// Create test descriptors
		desc1 := prometheus.NewDesc("test_metric1", "Test metric 1", nil, nil)
		desc2 := prometheus.NewDesc("test_metric2", "Test metric 2", nil, nil)
		desc3 := prometheus.NewDesc("test_metric3", "Test metric 3", nil, nil)
		descs := []*prometheus.Desc{desc1, desc2, desc3}

		// Call the function to test
		initDesc(ch, descs)

		// Close the channel to signal no more data
		close(ch)

		// Collect all descriptors from the channel
		var receivedDescs []*prometheus.Desc
		for desc := range ch {
			receivedDescs = append(receivedDescs, desc)
		}

		// Verify all descriptors were sent
		convey.So(len(receivedDescs), convey.ShouldEqual, len(descs))

		// Verify the descriptors are correct
		convey.So(receivedDescs[0], convey.ShouldEqual, desc1)
		convey.So(receivedDescs[1], convey.ShouldEqual, desc2)
		convey.So(receivedDescs[2], convey.ShouldEqual, desc3)
	})

	convey.Convey("Test initDesc with empty descriptors", t, func() {
		// Create a channel
		ch := make(chan *prometheus.Desc, 10)

		// Call the function with empty descriptors
		initDesc(ch, []*prometheus.Desc{})

		// Close the channel
		close(ch)

		// Verify no descriptors were sent
		var receivedDescs []*prometheus.Desc
		for desc := range ch {
			receivedDescs = append(receivedDescs, desc)
		}

		convey.So(len(receivedDescs), convey.ShouldEqual, 0)
	})
}

// TestDescribeUB test Describe
func TestDescribeUB(t *testing.T) {

	convey.Convey("test prometheus desc ", t, func() {
		ch := make(chan *prometheus.Desc, maxMetricsCount)
		c := UbCollector{}
		c.Describe(ch)
		t.Logf("Describe len(ch):%v", len(ch))
		convey.So(ch, convey.ShouldNotBeEmpty)
	})
}

// TestGetUBStatInfo tests the getUBStatInfo function
func TestGetUBStatInfo(t *testing.T) {
	convey.Convey("Test getUBStatInfo function with success and is UBOE port", t, func() {
		// Create test parameters
		logicID := int32(0)
		dieID := 0
		portID := 0

		// Mock hccn.GetNPUUbStatInfo to return success with isUboePort=1
		mockUbInfo := map[string]string{
			isUboePort: "1", // is UBOE port
		}

		// Track function calls
		getNPUUbStatInfoCalled := false
		convertUboeExtensionsCalled := false
		convertUBCommonStatsCalled := false
		resetErrCntCalled := false

		// Use gomonkey to patch functions
		patches := gomonkey.ApplyFunc(hccn.GetNPUUbStatInfo, func(logicID int32, dieID, portID int32) (map[string]string, error) {
			getNPUUbStatInfoCalled = true
			return mockUbInfo, nil
		}).ApplyFunc(convertUboeExtensions, func(ubInfos *common.UBInfo, ubInfo map[string]string) {
			convertUboeExtensionsCalled = true
		}).ApplyFunc(convertUBCommonStats, func(ubInfos *common.UBInfo, ubInfo map[string]string) {
			convertUBCommonStatsCalled = true
		}).ApplyFunc(hwlog.ResetErrCnt, func(domain string, id interface{}) {
			resetErrCntCalled = true
		})
		defer patches.Reset()

		// Call the function to test
		result := getUBStatInfo(logicID, dieID, portID)

		// Verify all expected functions were called
		convey.So(getNPUUbStatInfoCalled, convey.ShouldBeTrue)
		convey.So(convertUboeExtensionsCalled, convey.ShouldBeTrue)
		convey.So(convertUBCommonStatsCalled, convey.ShouldBeTrue)
		convey.So(resetErrCntCalled, convey.ShouldBeTrue)

		// Verify the result is not nil and has the correct structure
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.UBCommonStats, convey.ShouldNotBeNil)
		convey.So(result.UboeExtensions, convey.ShouldNotBeNil)
	})
	testGetUBStatCase2(t)
	testGetUBStatCase3(t)
}

func testGetUBStatCase2(t *testing.T) {
	convey.Convey("Test getUBStatInfo function with success but not UBOE port", t, func() {
		// Create test parameters
		logicID := int32(0)
		dieID := 0
		portID := 0

		// Mock hccn.GetNPUUbStatInfo to return success with isUboePort=0
		mockUbInfo := map[string]string{
			isUboePort: "0", // not UBOE port
		}

		// Track function calls
		getNPUUbStatInfoCalled := false
		convertUboeExtensionsCalled := false
		convertUBCommonStatsCalled := false
		resetErrCntCalled := false

		// Use gomonkey to patch functions
		patches := gomonkey.ApplyFunc(hccn.GetNPUUbStatInfo, func(logicID int32, dieID, portID int32) (map[string]string, error) {
			getNPUUbStatInfoCalled = true
			return mockUbInfo, nil
		}).ApplyFunc(convertUboeExtensions, func(ubInfos *common.UBInfo, ubInfo map[string]string) {
			convertUboeExtensionsCalled = true
		}).ApplyFunc(convertUBCommonStats, func(ubInfos *common.UBInfo, ubInfo map[string]string) {
			convertUBCommonStatsCalled = true
		}).ApplyFunc(hwlog.ResetErrCnt, func(domain string, id interface{}) {
			resetErrCntCalled = true
		})
		defer patches.Reset()

		// Call the function to test
		result := getUBStatInfo(logicID, dieID, portID)

		// Verify expected functions were called
		convey.So(getNPUUbStatInfoCalled, convey.ShouldBeTrue)
		convey.So(convertUboeExtensionsCalled, convey.ShouldBeFalse) // Should not be called
		convey.So(convertUBCommonStatsCalled, convey.ShouldBeTrue)
		convey.So(resetErrCntCalled, convey.ShouldBeTrue)

		// Verify the result is not nil and has the correct structure
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.UBCommonStats, convey.ShouldNotBeNil)
		convey.So(result.UboeExtensions, convey.ShouldNotBeNil)
	})
}

func testGetUBStatCase3(t *testing.T) {
	convey.Convey("Test getUBStatInfo function with failure", t, func() {
		// Create test parameters
		logicID := int32(0)
		dieID := 0
		portID := 0

		// Create a test error
		testErr := fmt.Errorf("test error")

		// Track function calls
		getNPUUbStatInfoCalled := false
		logWarnMetricsWithLimitCalled := false

		// Use gomonkey to patch functions
		patches := gomonkey.ApplyFunc(hccn.GetNPUUbStatInfo, func(logicID int32, dieID, portID int32) (map[string]string, error) {
			getNPUUbStatInfoCalled = true
			return nil, testErr
		}).ApplyFunc(logWarnMetricsWithLimit, func(domain string, logicID int32, uDie, port int, err error) {
			logWarnMetricsWithLimitCalled = true
			convey.So(err, convey.ShouldEqual, testErr)
		})
		defer patches.Reset()

		// Call the function to test
		result := getUBStatInfo(logicID, dieID, portID)

		// Verify expected functions were called
		convey.So(getNPUUbStatInfoCalled, convey.ShouldBeTrue)
		convey.So(logWarnMetricsWithLimitCalled, convey.ShouldBeTrue)

		// Verify the result is not nil even on error (should return empty UBInfo)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.UBCommonStats, convey.ShouldNotBeNil)
		convey.So(result.UboeExtensions, convey.ShouldNotBeNil)
	})
}

// TestCollectUBInfo tests the collectUBInfo function
func TestCollectUBInfo(t *testing.T) {
	convey.Convey("Test collectUBInfo function", t, func() {
		// Track calls to getUBStatInfo
		callCount := 0
		expectedCalls := common.MaxDieID * common.MaxPortID

		// Mock getUBStatInfo function to match the new signature
		mockGetUBStatInfo := func(logicID int32, dieID, portID int) *common.UBInfo {
			callCount++
			// Return a mock UBInfo
			return &common.UBInfo{
				UBCommonStats:  &common.UBCommonStats{},
				UboeExtensions: &common.UBOEExtensions{},
			}
		}

		// Use gomonkey to patch getUBStatInfo
		patch := gomonkey.ApplyFunc(getUBStatInfo, mockGetUBStatInfo)
		defer patch.Reset()

		// Call the function to test
		logicID := int32(0)
		result := collectUbInfo(logicID)

		// Verify getUBStatInfo was called the expected number of times
		convey.So(callCount, convey.ShouldEqual, expectedCalls)

		// Verify the result contains the expected number of UBInfo objects
		convey.So(len(result), convey.ShouldEqual, expectedCalls)
	})
}

// TestIsSupportedUB tests the IsSupported method of UbCollector
func TestIsSupportedUB(t *testing.T) {
	n := mockNewNpuCollector()
	cases := []testCase{
		buildTestCase("UBCollector: testIsSupported on Ascend910A3", &UbCollector{}, api.Ascend910A3, false),
		buildTestCase("UBCollector: testIsSupported on Ascend910A5", &UbCollector{}, api.Ascend910A5, true),
	}

	for _, c := range cases {
		patches := gomonkey.NewPatches()
		convey.Convey(c.name, t, func() {
			defer patches.Reset()
			patches.ApplyMethodReturn(n.Dmgr, "GetMainBoardId", uint32(api.Atlas9501DMainBoardID))
			patches.ApplyMethodReturn(n.Dmgr, "GetDevType", c.deviceType)
			isSupported := c.collectorType.IsSupported(n)
			convey.So(isSupported, convey.ShouldEqual, c.expectValue)
		})
	}
}

func TestUpdateTelegrafUB(t *testing.T) {
	n := mockNewNpuCollector()
	convey.Convey("TestUpdateTelegrafUB", t, func() {
		containerInfos := mockGetContainerNPUInfo()
		chips := mockGetNPUChipList()
		c := UbCollector{}
		mockUBCache(n, chips, colcommon.GetCacheKey(&c))
		fieldsMap := make(map[string]map[string]interface{})
		result := c.UpdateTelegraf(fieldsMap, n, containerInfos, chips)
		convey.So(result, convey.ShouldNotBeEmpty)
	})
}

func mockUBCache(n *colcommon.NpuCollector, chips []colcommon.HuaWeiAIChip, cacheKey string) {
	localCache := sync.Map{}
	for _, chip := range chips {
		localCache.Store(chip.PhyId, ubCache{chip: chip, timestamp: time.Now(),
			ubInfo: mockUBInfo()})
	}
	colcommon.UpdateCache[ubCache](n, cacheKey, &localCache)
}

func mockUBInfo() []*common.UBInfo {
	var newUbInfos []*common.UBInfo
	for i := 0; i < (common.MaxDieID * common.MaxPortID); i++ {
		ubInfos := common.UBInfo{
			UBCommonStats: initUBCommonStats(),
			UboeExtensions: &common.UBOEExtensions{
				CoreMibRxPausePkts: 1,
				CoreMibTxPausePkts: 1,
				CoreMibRxPfcPkts:   1,
				CoreMibTxPfcPkts:   1,
				CoreMibRxBadPkts:   1,
				CoreMibTxBadPkts:   1,
				CoreMibRxBadOctets: 1,
				CoreMibTxBadOctets: 1,
			},
		}
		newUbInfos = append(newUbInfos, &ubInfos)
	}
	return newUbInfos
}

func initUBCommonStats() *common.UBCommonStats {
	return &common.UBCommonStats{
		UbIpv4PktCntRx:    1,
		UbIpv6PktCntRx:    1,
		UnicIpv4PktCntRx:  1,
		UnicIpv6PktCntRx:  1,
		UbCompactPktCntRx: 1,
		UbUmocCtphCntRx:   1,
		UbUmocNtphCntRx:   1,
		UbMemPktCntRx:     1,
		UnknownPktCntRx:   1,
		DropIndCntRx:      1,
		ErrIndCntRx:       1,
		ToHostPktCntRx:    1,
		ToImpPktCntRx:     1,
		ToMarPktCntRx:     1,
		ToLinkPktCntRx:    1,
		ToNocPktCntRx:     1,
		RouteErrCntRx:     1,
		OutErrCntRx:       1,
		LengthErrCntRx:    1,
		RxBusiFlitNum:     1,
		RxSendAckFlit:     1,
		UbIpv4PktCntTx:    1,
		UbIpv6PktCntTx:    1,
		UnicIpv4PktCntTx:  1,
		UnicIpv6PktCntTx:  1,
		UbCompactPktCntTx: 1,
		UbUmocCtphCntTx:   1,
		UbUmocNtphCntTx:   1,
		UbMemPktCntTx:     1,
		UnknownPktCntTx:   1,
		DropIndCntTx:      1,
		ErrIndCntTx:       1,
		LpbkIndCntTx:      1,
		OutErrCntTx:       1,
		LengthErrCntTx:    1,
		TxBusiFlitNum:     1,
		TxRecvAckFlit:     1,
		RetryReqSum:       1,
		RetryAckSum:       1,
		CrcErrorSum:       1,
	}
}

// TestCollectToCacheUB test CollectToCache
func TestCollectToCacheUB(t *testing.T) {
	n := mockNewNpuCollector()

	convey.Convey("TestCollectToCacheUB", t, func() {

		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFuncReturn(hccn.GetNPUUbStatInfo, mockUBStatInfo(), nil)
		chips := mockGetNPUChipList()
		c := UbCollector{}
		c.PreCollect(n, chips)
		c.CollectToCache(n, chips)
		convey.So(colcommon.GetInfoFromCache[ubCache](n, colcommon.GetCacheKey(&UbCollector{})),
			convey.ShouldNotBeEmpty)
	})
}

func mockUBStatInfo() map[string]string {
	return map[string]string{
		isUboePort:        "0",
		ubIpv4PktCntRx:    "1",
		ubIpv6PktCntRx:    "1",
		unicIpv4PktCntRx:  "1",
		unicIpv6PktCntRx:  "1",
		ubCompactPktCntRx: "1",
		ubUmocCtphCntRx:   "1",
		ubUmocNtphCntRx:   "1",
		ubMemPktCntRx:     "1",
		unknownPktCntRx:   "1",
		dropIndCntRx:      "1",
		errIndCntRx:       "1",
		toHostPktCntRx:    "1",
		toImpPktCntRx:     "1",
		toMarPktCntRx:     "1",
		toLinkPktCntRx:    "1",
		toNocPktCntRx:     "1",
		routeErrCntRx:     "1",
		outErrCntRx:       "1",
		lengthErrCntRx:    "1",
		rxBusiFlitNum:     "1",
		rxSendAckFlit:     "1",
		ubIpv4PktCntTx:    "1",
		ubIpv6PktCntTx:    "1",
		unicIpv4PktCntTx:  "1",
		unicIpv6PktCntTx:  "1",
		ubCompactPktCntTx: "1",
		ubUmocCtphCntTx:   "1",
		ubUmocNtphCntTx:   "1",
		ubMemPktCntTx:     "1",
		unknownPktCntTx:   "1",
		dropIndCntTx:      "1",
		errIndCntTx:       "1",
		lpbkIndCntTx:      "1",
		outErrCntTx:       "1",
		lengthErrCntTx:    "1",
		txBusiFlitNum:     "1",
		txRecvAckFlit:     "1",
		retryReqSum:       "1",
		retryAckSum:       "1",
		crcErrorSum:       "1",
	}
}

// TestUpdatePrometheusUB test UpdatePrometheus
func TestUpdatePrometheusUB(t *testing.T) {
	n := mockNewNpuCollector()
	convey.Convey("TestUpdatePrometheusUB", t, func() {
		containerInfos := mockGetContainerNPUInfo()
		chips := mockGetNPUChipList()
		c := UbCollector{}
		ch := make(chan prometheus.Metric)

		// Mock promUpdateUbInfo to avoid blocking
		promUpdateUbInfoCalled := false
		patches := gomonkey.ApplyFunc(promUpdateUbInfo, func(ch chan<- prometheus.Metric, cache ubCache, timestamp time.Time, cardLabel []string) {
			promUpdateUbInfoCalled = true
		})
		defer patches.Reset()

		mockUBCache(n, chips, colcommon.GetCacheKey(&c))
		c.UpdatePrometheus(ch, n, containerInfos, chips)

		// Verify promUpdateUbInfo was called
		convey.So(promUpdateUbInfoCalled, convey.ShouldBeTrue)
	})
}
