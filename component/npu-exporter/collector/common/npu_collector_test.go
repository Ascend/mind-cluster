/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2021-2024. All rights reserved.
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

// Package common for general collector
package common

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/collector/container/isula"
	"huawei.com/npu-exporter/v6/collector/container/v1"
	"huawei.com/npu-exporter/v6/utils/logger"
)

var (
	mockErr   = errors.New("mockErr")
	testError = errors.New(testErrorMsg)
)

const (
	cacheTime         = 60 * time.Second
	npuCount          = 8
	defaultUpdateTime = 10 * time.Millisecond
	num2              = 2
	num100            = 100
	mockKey           = "mockKey"
	mockValue         = "mockValue"

	// Test constants for setElabelInfo
	testCardID           = int32(1)
	testProductName      = "Atlas 900"
	testModel            = "Atlas-900-9000"
	testManufacturer     = "Huawei"
	testManufacturerDate = "2023-01-01"
	testSerialNumber     = "SN123456789"
	testDefaultSerial    = "NA"
	testErrorMsg         = "get elabel info failed"

	testDomainProductType = "productType"
	testCardIDForCache    = int32(2)
	testDeviceIDForCache  = int32(3)
	testProductTypeValue  = "Atlas-800I-A2"
	testDomainOther       = "otherDomain"
	testValueString       = "testValue"
	testValueInt          = 100
)

type mockContainerRuntimeOperator struct{}

// Init implements ContainerRuntimeOperator
func (operator *mockContainerRuntimeOperator) Init() error {
	return nil
}

// Close implements ContainerRuntimeOperator
func (operator *mockContainerRuntimeOperator) Close() error {
	return nil
}

// ContainerIDs implements ContainerRuntimeOperator
func (operator *mockContainerRuntimeOperator) GetContainers(ctx context.Context) ([]*container.CommonContainer, error) {
	return []*container.CommonContainer{}, nil
}

// GetContainerInfoByID implements ContainerRuntimeOperator
func (operator *mockContainerRuntimeOperator) GetContainerInfoByID(ctx context.Context, id string) (v1.Spec, error) {
	return v1.Spec{}, nil
}

// GetIsulaContainerInfoByID implements ContainerRuntimeOperator
func (operator *mockContainerRuntimeOperator) GetIsulaContainerInfoByID(ctx context.Context,
	id string) (isula.ContainerJson, error) {
	return isula.ContainerJson{}, nil
}

// GetContainerType implements ContainerRuntimeOperator
func (operator *mockContainerRuntimeOperator) GetContainerType() string {
	return container.DefaultContainer
}

func mockScan4AscendDevices(_ string) ([]int, bool, error) {
	return []int{1}, true, nil
}

func mockGetCgroupPath(controller, specCgroupsPath string) (string, error) {
	return "", nil
}

func makeMockDevicesParser() *container.DevicesParser {
	return &container.DevicesParser{
		RuntimeOperator: new(mockContainerRuntimeOperator),
	}
}

type newNpuCollectorTestCase struct {
	cacheTime    time.Duration
	updateTime   time.Duration
	deviceParser *container.DevicesParser
	dmgr         *devmanager.DeviceManager
}

// TestNewNpuCollector test method of NewNpuCollector
func TestNewNpuCollector(t *testing.T) {
	tc := newNpuCollectorTestCase{
		cacheTime:    cacheTime,
		updateTime:   defaultUpdateTime,
		deviceParser: &container.DevicesParser{},
		dmgr:         &devmanager.DeviceManager{},
	}

	c := NewNpuCollector(tc.cacheTime, tc.updateTime, tc.deviceParser, tc.dmgr)

	assert.NotNil(t, c)
}

type testCase struct {
	name        string
	wantErr     bool
	mockPart    interface{}
	expectValue interface{}
	expectCount interface{}
}

func newTestCase(name string, wantErr bool, mockPart interface{}) testCase {
	return testCase{
		name:     name,
		wantErr:  wantErr,
		mockPart: mockPart,
	}
}

// TestGetChipInfo test  method getChipInfo
func TestGetChipInfo(t *testing.T) {
	tests := []testCase{
		newTestCase("should return chip info successfully when dsmi works normally", false,
			&devmanager.DeviceManagerMock{}),
		newTestCase("should return nil when dsmi works abnormally", true, &devmanager.DeviceManagerMockErr{}),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chipInfo := getNPUChipList(tt.mockPart.(devmanager.DeviceInterface))
			t.Logf("%#v", chipInfo)
			assert.NotNil(t, chipInfo)
			if tt.wantErr {
				assert.Len(t, chipInfo, 0)
			} else {
				assert.NotNil(t, chipInfo)
			}
		})
	}
}

func init() {
	logger.HwLogConfig = &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	logger.InitLogger("Prometheus")
}

func mockGetNPUChipList() []HuaWeiAIChip {
	chips := make([]HuaWeiAIChip, 0)
	for id := int32(0); id < npuCount; id++ {
		chip := HuaWeiAIChip{
			CardId:   id,
			PhyId:    id,
			DeviceID: id,
			LogicID:  id,
		}

		chips = append(chips, chip)
	}
	return chips
}

// TestInitCardInfo test  method getChipInfo
func TestInitCardInfo(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(getNPUChipList, mockGetNPUChipList())
	defer patches.Reset()
	convey.Convey("test InitCardInfo", t, func() {

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()
		npuCollector := mockNewNpuCollector()

		InitCardInfo(&sync.WaitGroup{}, ctx, npuCollector)
		time.Sleep(time.Millisecond * num100)
		cancelFunc()
		chips := getChipListCache(npuCollector)
		convey.So(len(chips), convey.ShouldEqual, npuCount)
	})
}

// TestGetChipListCache test  method getChipListCache
func TestGetChipListCache(t *testing.T) {
	npuCollector := mockNewNpuCollector()
	tests := []testCase{
		{name: "should return 0 chips when cache is nil", wantErr: false, mockPart: func() {}, expectCount: 0},
		{name: "should return chips : " + strconv.Itoa(npuCount), expectCount: npuCount, wantErr: false,
			mockPart: func() { npuCollector.cache.Set(npuListCacheKey, mockGetNPUChipList(), cacheTime) }},
		{name: "should return 0 chips when cache value is nil", wantErr: false, expectCount: 0,
			mockPart: func() { npuCollector.cache.Set(npuListCacheKey, nil, cacheTime) }},
		{name: "should return 0 chips when value is a incorrect type", expectCount: 0, wantErr: false,
			mockPart: func() { npuCollector.cache.Set(npuListCacheKey, &HuaWeiAIChip{}, cacheTime) }},
		{name: "should return 0 chips when cache is empty", expectCount: 0, wantErr: false,
			mockPart: func() { npuCollector.cache.Set(npuListCacheKey, []HuaWeiAIChip{}, cacheTime) },
		},
	}

	convey.Convey("getChipListCache", t, func() {
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				tt.mockPart.(func())()
				chips := getChipListCache(npuCollector)
				assert.Len(t, chips, tt.expectCount.(int))
				convey.So(len(chips), convey.ShouldEqual, tt.expectCount)
			})
		}
	})
}

func mockNewNpuCollector() *NpuCollector {
	tc := newNpuCollectorTestCase{
		cacheTime:    cacheTime,
		updateTime:   defaultUpdateTime,
		deviceParser: &container.DevicesParser{},
		dmgr:         &devmanager.DeviceManager{},
	}
	c := NewNpuCollector(tc.cacheTime, tc.updateTime, tc.deviceParser, tc.dmgr)
	return c
}

func TestNpuChipInfoInitAtFirstTime(t *testing.T) {
	n := mockNewNpuCollector()
	convey.Convey("TestNpuChipInfoInitAtFirstTime", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFuncReturn(getNPUChipList, []HuaWeiAIChip{{CardId: 0}})
		// do test
		npuChipInfoInitAtFirstTime(n)
		// valid cache
		data, err := n.cache.Get(npuListCacheKey)
		convey.So(err, convey.ShouldBeNil)
		chips, ok := data.([]HuaWeiAIChip)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(chips), convey.ShouldEqual, 1)
	})
}

func patchCollectToCache() *gomonkey.Patches {
	return gomonkey.ApplyMethod(&MetricsCollectorAdapter{}, "CollectToCache",
		func(_ *MetricsCollectorAdapter, n *NpuCollector, chipList []HuaWeiAIChip) {
			n.cache.Set(mockKey, mockValue, n.cacheTime)
		})
}

func TestStartCollectForMultiGoroutine(t *testing.T) {
	n := mockNewNpuCollector()
	wg := sync.WaitGroup{}
	ChainForMultiGoroutine = []MetricsCollector{
		&MetricsCollectorAdapter{},
		&MetricsCollectorAdapter{},
	}
	patches := patchCollectToCache()
	defer patches.Reset()
	patches.ApplyFuncReturn(getChipListCache, []HuaWeiAIChip{createChip()})
	convey.Convey("TestStartCollectForMultiGoroutine", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		startCollectForMultiGoroutine(&wg, ctx, n)
		time.Sleep(n.updateTime)
		cancel()
		data, err := n.cache.Get(mockKey)
		convey.So(err, convey.ShouldBeNil)
		value, ok := data.(string)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(value, convey.ShouldEqual, mockValue)
	})
}

func TestRunChipCollector(t *testing.T) {
	n := mockNewNpuCollector()
	patches := patchCollectToCache()
	defer patches.Reset()
	convey.Convey("TestRunChipCollector", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		tickCh := make(chan time.Time)
		patches.ApplyFuncReturn(time.NewTicker, &time.Ticker{C: tickCh})
		close(tickCh)
		go runChipCollector(ctx, n, createChip())
		time.Sleep(n.updateTime)
		cancel()
		data, err := n.cache.Get(mockKey)
		convey.So(err, convey.ShouldBeNil)
		value, ok := data.(string)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(value, convey.ShouldEqual, mockValue)
	})
}

func TestStartCollectSingleGoroutine(t *testing.T) {
	n := mockNewNpuCollector()
	wg := sync.WaitGroup{}
	ChainForSingleGoroutine = []MetricsCollector{
		&MetricsCollectorAdapter{},
	}
	patches := patchCollectToCache()
	defer patches.Reset()
	convey.Convey("TestStartCollectSingleGoroutine", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		startCollectSingleGoroutine(&wg, ctx, n)
		time.Sleep(n.updateTime)
		cancel()
		data, err := n.cache.Get(mockKey)
		convey.So(err, convey.ShouldBeNil)
		value, ok := data.(string)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(value, convey.ShouldEqual, mockValue)
	})
}

type chipsCase struct {
	name        string
	devType     string
	buildChips  func()
	expectValue int
}

func TestGetChipListWithVNPU(t *testing.T) {
	n := mockNewNpuCollector()
	chip := HuaWeiAIChip{}
	tests := []chipsCase{
		{name: "TestGetChipListWithVNPU_310p_no_vnpu",
			devType: api.Ascend310P,
			buildChips: func() {
				chip = createChip()
			},
			expectValue: 1,
		},
		{name: "TestGetChipListWithVNPU_310p_2_vnpus",
			devType: api.Ascend310P,
			buildChips: func() {
				chip = createValidVnpuChip()
			},
			expectValue: num2,
		},
		{name: "TestGetChipListWithVNPU_910",
			devType: api.Ascend910,
			buildChips: func() {
				chip = createChip()
			},
			expectValue: 1,
		},
	}

	convey.Convey("TestGetChipListWithVNPU", t, func() {
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				tt.buildChips()
				patches := gomonkey.NewPatches()
				defer patches.Reset()
				patches.ApplyMethodReturn(n.Dmgr, "GetDevType", tt.devType)
				patches.ApplyFuncReturn(getChipListCache, []HuaWeiAIChip{chip})

				chips := GetChipListWithVNPU(n)
				convey.So(len(chips), convey.ShouldEqual, tt.expectValue)
			})
		}
	})
}

func createValidVnpuChip() HuaWeiAIChip {
	chip := createChip()
	chip.VDevInfos = &common.VirtualDevInfo{
		VDevActivityInfo: []common.VDevActivityInfo{
			{
				VDevID:       0,
				VDevAiCore:   0,
				VDevTotalMem: 0,
				VDevUsedMem:  0,
				IsVirtualDev: true,
			},
			{
				VDevID:       1,
				VDevAiCore:   1,
				VDevTotalMem: 1,
				VDevUsedMem:  1,
				IsVirtualDev: true,
			},
		},
	}
	return chip
}

func createChip() HuaWeiAIChip {
	return HuaWeiAIChip{
		CardId:   0,
		PhyId:    0,
		DeviceID: 0,
		LogicID:  0,
		ChipInfo: &common.ChipInfo{
			Name:    api.Ascend910,
			Type:    "Ascend",
			Version: "V1",
		},
	}
}

func TestSetPCIeBusInfo(t *testing.T) {
	const mockPcieBus = "0000:01:00.0"
	tests := []struct {
		name         string
		productTypes []string
		err          error
		expectValue  string
	}{{
		name:         "TestSetPCIeBusInfo_910",
		productTypes: []string{api.Ascend910},
		err:          nil,
		expectValue:  mockPcieBus,
	}, {
		name:         "TestSetPCIeBusInfo_910_err",
		productTypes: []string{api.Ascend910},
		err:          mockErr,
		expectValue:  "",
	}, {
		name:         "TestSetPCIeBusInfo_Atlas200ISoc",
		productTypes: []string{common.Atlas200ISoc},
		err:          nil,
		expectValue:  mockPcieBus,
	}, {
		name:         "TestSetPCIeBusInfo_Atlas200ISoc_err",
		productTypes: []string{common.Atlas200ISoc},
		err:          mockErr,
		expectValue:  "",
	}}
	chip := createChip()
	convey.Convey("TestSetPCIeBusInfo", t, func() {
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				dmgr := &devmanager.DeviceManager{ProductTypes: tt.productTypes}
				patches := gomonkey.NewPatches()
				defer patches.Reset()
				patches.ApplyMethodReturn(dmgr, "GetPCIeBusInfo", mockPcieBus, tt.err)

				setPCIeBusInfo(0, dmgr, &chip)
				convey.So(chip.PCIeBusInfo, convey.ShouldEqual, tt.expectValue)
			})
		}
	})
}

type setElabelInfoTestCase struct {
	name                   string
	cardID                 int32
	mockElabelInfo         common.ElabelInfo
	mockError              error
	expectSerial           string
	expectProduct          string
	expectModel            string
	expectManufacturer     string
	expectManufacturerDate string
}

func createSetElabelInfoTestCases() []setElabelInfoTestCase {
	return []setElabelInfoTestCase{
		{
			name:   "should set elabel info successfully when GetCardElabelV2 returns valid data",
			cardID: testCardID,
			mockElabelInfo: common.ElabelInfo{
				ProductName:      testProductName,
				Model:            testModel,
				Manufacturer:     testManufacturer,
				ManufacturerDate: testManufacturerDate,
				SerialNumber:     testSerialNumber,
			},
			mockError:              nil,
			expectSerial:           testSerialNumber,
			expectProduct:          testProductName,
			expectModel:            testModel,
			expectManufacturer:     testManufacturer,
			expectManufacturerDate: testManufacturerDate,
		},
		{
			name:                   "should set default elabel info when GetCardElabelV2 returns error",
			cardID:                 testCardID,
			mockElabelInfo:         common.ElabelInfo{},
			mockError:              testError,
			expectSerial:           testDefaultSerial,
			expectProduct:          "",
			expectModel:            "",
			expectManufacturer:     "",
			expectManufacturerDate: "",
		},
	}
}

func executeSetElabelInfoTest(tc setElabelInfoTestCase) {
	// Create mock device manager
	mockDmgr := &devmanager.DeviceManager{}

	// Create test chip
	chip := &HuaWeiAIChip{}

	// Apply gomonkey patches
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(mockDmgr, "GetCardElabelV2",
		tc.mockElabelInfo, tc.mockError)

	// Execute the function under test
	setElabelInfo(chip, mockDmgr, tc.cardID)

	// Verify results
	convey.So(chip.ElabelInfo, convey.ShouldNotBeNil)
	convey.So(chip.ElabelInfo.SerialNumber, convey.ShouldEqual, tc.expectSerial)
}

// TestSetElabelInfo test setElabelInfo method
func TestSetElabelInfo(t *testing.T) {
	testCases := createSetElabelInfoTestCases()

	convey.Convey("TestSetElabelInfo", t, func() {
		for _, tc := range testCases {
			convey.Convey(tc.name, func() {
				executeSetElabelInfoTest(tc)
			})
		}
	})
}

type saveToCacheTestCase struct {
	name      string
	domain    string
	cardID    int32
	deviceID  int32
	value     interface{}
	expectKey string
}

func createSaveToCacheTestCases() []saveToCacheTestCase {
	return []saveToCacheTestCase{
		{
			name:      "should save value to cache when domain cardID and deviceID are valid",
			domain:    testDomainProductType,
			cardID:    testCardIDForCache,
			deviceID:  testDeviceIDForCache,
			value:     testProductTypeValue,
			expectKey: testDomainProductType + "23",
		},
		{
			name:      "should save int value to cache when value is integer",
			domain:    testDomainOther,
			cardID:    testCardIDForCache,
			deviceID:  testDeviceIDForCache,
			value:     testValueInt,
			expectKey: testDomainOther + "23",
		},
		{
			name:      "should save string value to cache when value is string",
			domain:    testDomainOther,
			cardID:    int32(0),
			deviceID:  int32(1),
			value:     testValueString,
			expectKey: testDomainOther + "01",
		},
	}
}

func TestSaveToCache(t *testing.T) {
	testCases := createSaveToCacheTestCases()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clearCache()
			saveToCache(tc.domain, tc.cardID, tc.deviceID, tc.value)
			actualValue, ok := npuInfoCache.Load(tc.expectKey)
			convey.Convey("should save and retrieve value correctly", t, func() {
				convey.So(ok, convey.ShouldBeTrue)
				convey.So(actualValue, convey.ShouldEqual, tc.value)
			})
		})
	}
}

type getFromCacheTestCase struct {
	name        string
	setupCache  func()
	domain      string
	cardID      int32
	deviceID    int32
	expectValue interface{}
	expectOk    bool
}

func createGetFromCacheTestCases() []getFromCacheTestCase {
	return []getFromCacheTestCase{
		{
			name: "should return cached value when key exists in cache",
			setupCache: func() {
				clearCache()
				saveToCache(testDomainProductType, testCardIDForCache, testDeviceIDForCache,
					testProductTypeValue)
			},
			domain:      testDomainProductType,
			cardID:      testCardIDForCache,
			deviceID:    testDeviceIDForCache,
			expectValue: testProductTypeValue,
			expectOk:    true,
		},
		{
			name: "should return nil when key does not exist in cache",
			setupCache: func() {
				clearCache()
			},
			domain:      testDomainProductType,
			cardID:      testCardIDForCache,
			deviceID:    testDeviceIDForCache,
			expectValue: nil,
			expectOk:    false,
		},
		{
			name: "should return int value when cached value is integer",
			setupCache: func() {
				clearCache()
				saveToCache(testDomainOther, testCardIDForCache, testDeviceIDForCache, testValueInt)
			},
			domain:      testDomainOther,
			cardID:      testCardIDForCache,
			deviceID:    testDeviceIDForCache,
			expectValue: testValueInt,
			expectOk:    true,
		},
	}
}

func TestGetFromCache(t *testing.T) {
	testCases := createGetFromCacheTestCases()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupCache()
			actualValue := getFromCache(tc.domain, tc.cardID, tc.deviceID)
			convey.Convey("should return expected value", t, func() {
				if tc.expectOk {
					convey.So(actualValue, convey.ShouldEqual, tc.expectValue)
				} else {
					convey.So(actualValue, convey.ShouldBeNil)
				}
			})
		})
	}
}

type setProductTypeTestCase struct {
	name              string
	devType           string
	cacheValue        interface{}
	productType       string
	productTypeErr    error
	expectProductType string
}

func createSetProductTypeTestCases() []setProductTypeTestCase {
	return []setProductTypeTestCase{
		{
			name:              "should return early when device type is not Ascend310P",
			devType:           api.Ascend910,
			expectProductType: "",
		},
		{
			name:              "should set product type from cache when cache has value",
			devType:           api.Ascend310P,
			cacheValue:        testProductTypeValue,
			expectProductType: testProductTypeValue,
		},
		{
			name:              "should set product type from dmgr when cache is empty and GetProductType succeeds",
			devType:           api.Ascend310P,
			cacheValue:        nil,
			productType:       testProductTypeValue,
			productTypeErr:    nil,
			expectProductType: testProductTypeValue,
		},
		{
			name:              "should not set product type when GetProductType returns error",
			devType:           api.Ascend310P,
			cacheValue:        nil,
			productType:       "",
			productTypeErr:    mockErr,
			expectProductType: "",
		},
		{
			name:              "should set product type from dmgr when cache value is not string",
			devType:           api.Ascend310P,
			cacheValue:        testValueInt,
			productType:       testProductTypeValue,
			productTypeErr:    nil,
			expectProductType: testProductTypeValue,
		},
	}
}

func TestSetProductType(t *testing.T) {
	testCases := createSetProductTypeTestCases()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clearCache()
			chip := &HuaWeiAIChip{}
			mockDmgr := &devmanager.DeviceManager{}
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyMethodReturn(mockDmgr, "GetDevType", tc.devType)
			patches.ApplyFuncReturn(getFromCache, tc.cacheValue)
			needGetProduct := tc.devType == api.Ascend310P &&
				(tc.cacheValue == nil || !isStringType(tc.cacheValue))
			if needGetProduct {
				patches.ApplyMethodReturn(mockDmgr, "GetProductType", tc.productType, tc.productTypeErr)
			}

			setProductType(chip, mockDmgr, testCardIDForCache, testDeviceIDForCache)
			convey.Convey("should set product type correctly", t, func() {
				convey.So(chip.ProductType, convey.ShouldEqual, tc.expectProductType)
			})
		})
	}
}

func isStringType(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

func clearCache() {
	npuInfoCache.Range(func(key, value interface{}) bool {
		npuInfoCache.Delete(key)
		return true
	})
}
