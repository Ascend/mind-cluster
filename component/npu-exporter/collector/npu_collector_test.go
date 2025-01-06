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

// Package collector for Prometheus
package collector

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/cache"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/collector/container/isula"
	"huawei.com/npu-exporter/v6/collector/container/v1"
)

const (
	cacheTime = 60 * time.Second
	timestamp = 1606402
	waitTime  = 2 * time.Second
	npuCount  = 8
	time5s    = 5 * time.Second
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

// TestGetChipInfo test  method getChipInfo
func TestGetChipInfo(t *testing.T) {
	tests := []testCase{
		newTestCase("should return chip info successfully when dsmi works normally", false,
			&devmanager.DeviceManagerMock{}),
		newTestCase("should return nil when dsmi works abnormally", true, &devmanager.DeviceManagerMockErr{}),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chipInfo := packChipInfo(0, tt.mockPart.(devmanager.DeviceInterface))
			t.Logf("%#v", chipInfo)
			assert.NotNil(t, chipInfo)
			if tt.wantErr {
				assert.Nil(t, chipInfo.ChipIfo)
			} else {
				assert.NotNil(t, chipInfo.ChipIfo)
			}
		})
	}
}

// TestStartToGetNetInfo test  method startToGetNetInfo
func TestStartToGetNetInfo(t *testing.T) {
	tests := []testCase{
		newTestCase("should return chip info successfully when dsmi works normally", false,
			&devmanager.DeviceManagerMock{}),
	}
	mk := gomonkey.ApplyFunc(hccn.GetNPULinkStatus, func(_ int32) (string, error) {
		return "UP", nil
	}).ApplyFunc(hccn.GetNPUInterfaceTraffic, func(_ int32) (float64, float64, error) {
		return 1, 1, nil
	}).ApplyFunc(hccn.GetNPULinkUpNum, func(_ int32) (int, error) {
		return 1, nil
	}).ApplyFunc(hccn.GetNPULinkSpeed, func(_ int32) (int, error) {
		return 1, nil
	})
	mk.ApplyFunc(hccn.GetNPUOpticalInfo, mockGetNPUOpticalInfo)
	mk.ApplyFunc(hccn.GetNPUStatInfo, mockGetNPUStatInfo)
	defer mk.Reset()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancelFunc := context.WithCancel(context.Background())
			startToGetNetInfo(ctx, tt.mockPart.(devmanager.DeviceInterface), 1*time.Second)
			<-time.After(1 * time.Second)
			cancelFunc()
		})
	}
}

func mockGetNPUOpticalInfo(_ int32) (map[string]string, error) {
	return map[string]string{
		"Tx_Power0":   "1 mW",
		"Tx_Power1":   "1 mW",
		"Tx_Power2":   "1 mW",
		"Tx_Power3":   "1 mW",
		"Rx_Power0":   "1 mW",
		"Rx_Power1":   "1 mW",
		"Rx_Power2":   "1 mW",
		"Rx_Power3":   "1 mW",
		"Vcc":         "1 mV",
		"temperature": "50 C",
		"present":     "1.0",
	}, nil
}

func mockGetNPUStatInfo(_ int32) (map[string]int, error) {

	res := make(map[string]int)
	res["mac_rx_mac_pause_num"] = 0
	res["mac_tx_mac_pause_num"] = 0
	res["mac_rx_pfc_pkt_num"] = 0
	res["mac_tx_pfc_pkt_num"] = 0
	res["mac_rx_bad_pkt_num"] = 0
	res["mac_tx_bad_pkt_num"] = 0
	res["roce_rx_all_pkt_num"] = 0
	res["roce_tx_all_pkt_num"] = 0
	res["roce_rx_err_pkt_num"] = 0
	res["roce_tx_err_pkt_num"] = 0
	res["roce_rx_cnp_pkt_num"] = 0
	res["roce_tx_cnp_pkt_num"] = 0
	res["mac_rx_bad_oct_num"] = 0
	res["mac_tx_bad_oct_num"] = 0
	res["roce_unexpected_ack_num"] = 0
	res["roce_out_of_order_num"] = 0
	res["roce_verification_err_num"] = 0
	res["roce_qp_status_err_num"] = 0
	res["roce_new_pkt_rty_num"] = 0
	res["roce_ecn_db_num"] = 0
	res["mac_rx_fcs_err_pkt_num"] = 0
	return res, nil
}

// TestGetHealthCode test getHealthCode
func TestGetHealthCode(t *testing.T) {
	tests := []struct {
		name   string
		health string
		want   int
	}{
		{
			name:   "should return 1 when given Healthy",
			health: Healthy,
			want:   1,
		},
		{
			name:   "should return 0 when given UnHealthy",
			health: UnHealthy,
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHealthCode(tt.health); got != tt.want {
				t.Errorf("getHealthCode() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

// TestGetNPUInfo test method of getNPUInfo
func TestGetNPUInfo(t *testing.T) {
	tests := []struct {
		name string
		args devmanager.DeviceInterface
		want []HuaWeiNPUCard
	}{
		{
			name: "should return at lease one NPUInfo",
			args: &devmanager.DeviceManagerMock{},
			want: []HuaWeiNPUCard{{
				DeviceList: nil,
				Timestamp:  time.Time{},
				CardID:     0,
			}},
		},
		{
			name: "should return zero NPU",
			args: &devmanager.DeviceManagerMockErr{},
			want: []HuaWeiNPUCard{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNPUInfo(tt.args); len(got) != len(tt.want) {
				t.Errorf("getNPUInfo() = %#v,want %#v", got, tt.want)
			}
		})
	}
}

// TestGetNPUInfoFor910A3 test method of getNPUInfo for 910A3
func TestGetNPUInfoFor910A3(t *testing.T) {
	tests := []struct {
		name string
		args devmanager.DeviceInterface
		want []HuaWeiNPUCard
	}{
		{
			name: "should return at lease one NPUInfo",
			args: &devmanager.DeviceManager910A3Mock{},
			want: []HuaWeiNPUCard{{
				DeviceList: nil,
				Timestamp:  time.Time{},
				CardID:     0,
			}},
		},
		{
			name: "should return zero NPU",
			args: &devmanager.DeviceManager910A3MockErr{},
			want: []HuaWeiNPUCard{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNPUInfo(tt.args); len(got) != len(tt.want) {
				t.Errorf("getNPUInfo() = %#v,want %#v", got, tt.want)
			}
		})
	}
}

type newNpuCollectorTestCase struct {
	cacheTime    time.Duration
	updateTime   time.Duration
	deviceParser *container.DevicesParser
}

// TestNewNpuCollector test method of NewNpuCollector
func TestNewNpuCollector(t *testing.T) {
	const defaultUpdateTime = 5 * time.Second
	tc := newNpuCollectorTestCase{
		cacheTime:    cacheTime,
		updateTime:   defaultUpdateTime,
		deviceParser: &container.DevicesParser{},
	}

	c := NewNpuCollector(tc.cacheTime, tc.updateTime, tc.deviceParser)
	assert.Equal(t, defaultUpdateTime, c.updateTime)
	assert.Equal(t, cacheTime, c.cacheTime)
	assert.Equal(t, tc.deviceParser, c.devicesParser)
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

func mockGetNPUInfo(dmgr devmanager.DeviceInterface) []HuaWeiNPUCard {
	npuList := make([]HuaWeiNPUCard, 0)
	for devicePhysicID := int32(0); devicePhysicID < npuCount; devicePhysicID++ {
		chipInfo := &HuaWeiAIChip{
			HealthStatus:      Healthy,
			ErrorCodes:        []int64{},
			Utilization:       0,
			Temperature:       0,
			Power:             0,
			Voltage:           0,
			AICoreCurrentFreq: 0,
			Meminf: &common.MemoryInfo{
				MemorySize:  0,
				Frequency:   0,
				Utilization: 0,
			},
			ChipIfo: &common.ChipInfo{
				Type:    "Ascend",
				Name:    "910Awn",
				Version: "V1",
			},
			HbmInfo: &common.HbmAggregateInfo{
				HbmInfo: &common.HbmInfo{
					MemorySize:        0,
					Frequency:         0,
					Usage:             0,
					Temp:              0,
					BandWidthUtilRate: 0,
				},
			},
			DevProcessInfo:  &common.DevProcessInfo{},
			LinkStatus:      LinkDown,
			NetHealthStatus: UnHealthy,
		}
		chipInfo.DeviceID = int(devicePhysicID)
		npuCard := HuaWeiNPUCard{
			CardID:     int(devicePhysicID),
			DeviceList: []*HuaWeiAIChip{chipInfo},
			Timestamp:  time.Unix(timestamp, 0),
		}
		npuList = append(npuList, npuCard)
	}
	return npuList
}

// TestStart test start method
func TestStart(t *testing.T) {
	ch := make(chan os.Signal)
	tests := []struct {
		collector *npuCollector
		name      string
	}{
		{
			name: "should set cache successfully",
			collector: &npuCollector{
				cache:         cache.New(cacheSize),
				cacheTime:     cacheTime,
				updateTime:    time5s,
				devicesParser: makeMockDevicesParser(),
			},
		},
	}
	mk := gomonkey.ApplyFunc(getNPUInfo, mockGetNPUInfo)
	defer mk.Reset()
	patch := gomonkey.ApplyFunc(devmanager.AutoInit, func(_ string) (*devmanager.DeviceManager, error) {
		return &devmanager.DeviceManager{DcMgr: &devmanager.A910Manager{}}, nil
	})
	defer patch.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go Start(ctx, cancel, tt.collector)
			time.Sleep(waitTime)
			objm, err := tt.collector.cache.Get(npuListCacheKey)
			assert.NotNil(t, objm)
			assert.Nil(t, err)

			ch1 := make(chan *prometheus.Desc, 1000)
			tt.collector.Describe(ch1)
			t.Logf("Describe len(ch):%v", len(ch1))
			assert.NotEmpty(t, ch1, "Expected ch1 to be not empty")

			ch2 := make(chan prometheus.Metric, 1000)
			tt.collector.Collect(ch2)
			t.Logf("Collect len(ch):%v", len(ch2))
			assert.NotEmpty(t, ch1, "Expected ch1 to be not empty")

			go func() {
				ch <- os.Interrupt
				close(ch)
			}()
		})
	}
}

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&config, nil)
}
