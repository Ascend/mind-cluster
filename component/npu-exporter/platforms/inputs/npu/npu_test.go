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

// Package npu this for parse and pack
package npu

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/influxdata/telegraf"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/collector/metrics"
	"huawei.com/npu-exporter/v6/utils/logger"
)

const (
	num5              = 5
	testVersionMetric = "npu_exporter_version_info"
	testVersionValue  = "7.0.0"
)

func init() {
	logger.HwLogConfig = &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	logger.InitLogger("Prometheus")
	initChain()
}

func initChain() {
	common.ChainForSingleGoroutine = []common.MetricsCollector{
		&metrics.VersionCollector{},
	}
}

func mockNewNpuCollector() *common.NpuCollector {
	tc := newNpuCollectorTestCase{
		cacheTime:    time.Duration(num5),
		updateTime:   time.Duration(num5),
		deviceParser: &container.DevicesParser{},
		dmgr:         &devmanager.DeviceManager{},
	}
	c := common.NewNpuCollector(tc.cacheTime, tc.updateTime, tc.deviceParser, tc.dmgr)
	return c
}

// TestGather verifies different device type scenarios
func TestGather(t *testing.T) {
	tests := []struct {
		name        string
		deviceType  string
		expectedTag string
	}{
		{name: api.Ascend910A3,
			deviceType:  api.Ascend910A3,
			expectedTag: api.Ascend910,
		},
		{name: api.Ascend310P,
			deviceType:  api.Ascend310P,
			expectedTag: api.Ascend310P,
		},
	}
	npu := &WatchNPU{
		collector: mockNewNpuCollector(),
	}
	acc := &MockAccumulator{}

	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyMethodReturn(npu.collector.Dmgr, "GetDevType", tt.deviceType)
			patches.ApplyFuncReturn(common.GetContainerNPUInfo, nil)
			patches.ApplyFuncReturn(common.GetChipListWithVNPU, nil)
			patches.ApplyMethodFunc(common.ChainForSingleGoroutine[0], "UpdateTelegraf",
				func(ch chan<- common.TelegrafMetric, n *common.NpuCollector,
					containerMap map[int32]container.DevicesInfo, chips []common.HuaWeiAIChip) {
					ch <- common.TelegrafMetric{
						DeviceID: common.NoDeviceID,
						VDevID:   common.NoDeviceID,
						Fields:   map[string]interface{}{testVersionMetric: testVersionValue},
					}
				})

			err := npu.Gather(acc)
			convey.So(err, convey.ShouldBeNil)
			convey.So(acc.fields["ascend,device="+strings.ToLower(tt.expectedTag)], convey.ShouldNotBeEmpty)
		})
	}
}

type newNpuCollectorTestCase struct {
	cacheTime    time.Duration
	updateTime   time.Duration
	deviceParser *container.DevicesParser
	dmgr         *devmanager.DeviceManager
}

// MockAccumulator is a mock implementation of telegraf.Accumulator
type MockAccumulator struct {
	fields map[string]map[string]interface{}
}

func (m *MockAccumulator) AddFields(measurement string, fields map[string]interface{}, tags map[string]string,
	t ...time.Time) {
	if m.fields == nil {
		m.fields = make(map[string]map[string]interface{})
	}
	pairs := make([]string, 0, len(tags))
	for k, v := range tags {
		pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
	}
	sort.Strings(pairs)
	metricKey := measurement + "," + strings.Join(pairs, ",")
	m.fields[metricKey] = fields
}

func (m *MockAccumulator) AddGauge(measurement string, fields map[string]interface{}, tags map[string]string,
	t ...time.Time) {
}

func (m *MockAccumulator) AddCounter(measurement string, fields map[string]interface{}, tags map[string]string,
	t ...time.Time) {
}

func (m *MockAccumulator) AddSummary(measurement string, fields map[string]interface{}, tags map[string]string,
	t ...time.Time) {
}

func (m *MockAccumulator) AddHistogram(measurement string, fields map[string]interface{}, tags map[string]string,
	t ...time.Time) {
}

func (m *MockAccumulator) AddMetric(metric telegraf.Metric) {
}

func (m *MockAccumulator) SetPrecision(precision time.Duration) {
}

func (m *MockAccumulator) AddError(err error) {
}

func (m *MockAccumulator) WithTracking(maxTracked int) telegraf.TrackingAccumulator {
	return nil
}

const (
	testDevTagValue = "ascend910"
	testDeviceID0   = 0
	testDeviceID1   = 1
	testVDevID100   = 100
	mockValue50     = 50
	testFieldTemp   = "temperature"
	testFieldPower  = "power"
	testMeasCustom  = "custom_meas"
	testTagJob      = "job"
	testTagValTrain = "train"
)

// TestBuildTags should return custom labels when Labels is set
func TestBuildTags(t *testing.T) {
	tests := []struct {
		name     string
		data     common.TelegrafMetric
		devTag   string
		expected map[string]string
	}{
		{name: "should return custom labels when Labels is set",
			data:     common.TelegrafMetric{Labels: map[string]string{testTagJob: testTagValTrain}},
			devTag:   testDevTagValue,
			expected: map[string]string{testTagJob: testTagValTrain}},
		{name: "should return node level tag when DeviceID < 0 and Labels nil",
			data:     common.TelegrafMetric{DeviceID: common.NoDeviceID, VDevID: common.NoDeviceID},
			devTag:   testDevTagValue,
			expected: map[string]string{deviceTagKey: testDevTagValue}},
		{name: "should return device tag when DeviceID >= 0 and VDevID < 0",
			data:     common.TelegrafMetric{DeviceID: testDeviceID0, VDevID: common.NoDeviceID},
			devTag:   testDevTagValue,
			expected: map[string]string{deviceTagKey: testDevTagValue + "-0"}},
		{name: "should return device and vdev tags when VDevID >= 0",
			data:     common.TelegrafMetric{DeviceID: testDeviceID1, VDevID: testVDevID100},
			devTag:   testDevTagValue,
			expected: map[string]string{deviceTagKey: testDevTagValue + "-1", vDevTagKey: "100"}},
	}
	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			result := buildTags(tt.data, tt.devTag)
			convey.So(result, convey.ShouldResemble, tt.expected)
		})
	}
}

// TestGroupKey should produce deterministic key for same measurement and tags
func TestGroupKey(t *testing.T) {
	tests := []struct {
		name        string
		measurement string
		tags        map[string]string
		expected    string
	}{
		{name: "should return measurement only when tags empty",
			measurement: devName,
			tags:        map[string]string{},
			expected:    devName},
		{name: "should produce sorted key when tags given",
			measurement: devName,
			tags:        map[string]string{vDevTagKey: "100", deviceTagKey: "npu-0"},
			expected:    devName + "|" + deviceTagKey + "=npu-0|" + vDevTagKey + "=100"},
	}
	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			result := groupKey(tt.measurement, tt.tags)
			convey.So(result, convey.ShouldEqual, tt.expected)
		})
	}
}

// TestMergeFields should keep first when duplicate field name
func TestMergeFields(t *testing.T) {
	tests := []struct {
		name     string
		dst      map[string]interface{}
		src      map[string]interface{}
		expected map[string]interface{}
	}{
		{name: "should merge all fields when no duplicate",
			dst:      map[string]interface{}{testFieldTemp: mockValue50},
			src:      map[string]interface{}{testFieldPower: mockValue50},
			expected: map[string]interface{}{testFieldTemp: mockValue50, testFieldPower: mockValue50}},
		{name: "should keep first when duplicate field name",
			dst:      map[string]interface{}{testFieldTemp: mockValue50},
			src:      map[string]interface{}{testFieldTemp: mockValue50},
			expected: map[string]interface{}{testFieldTemp: mockValue50}},
	}
	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			mergeFields(tt.dst, tt.src)
			convey.So(tt.dst, convey.ShouldResemble, tt.expected)
		})
	}
}

func newDeviceMetric(deviceID int32, fields map[string]interface{}) common.TelegrafMetric {
	return common.TelegrafMetric{
		DeviceID: deviceID, VDevID: common.NoDeviceID, Fields: fields}
}

func newVnpuMetric(deviceID, vdevID int32, fields map[string]interface{}) common.TelegrafMetric {
	return common.TelegrafMetric{DeviceID: deviceID, VDevID: vdevID, Fields: fields}
}

func newLabeledMetric(measurement string, labels map[string]string,
	fields map[string]interface{}) common.TelegrafMetric {
	return common.TelegrafMetric{
		Measurement: measurement, Labels: labels,
		DeviceID: common.NoDeviceID, VDevID: common.NoDeviceID, Fields: fields}
}

func metricKey(measurement string, tagPairs ...string) string {
	return measurement + "," + strings.Join(tagPairs, ",")
}

func consumeAndVerify(t *testing.T, metrics []common.TelegrafMetric, devTagValue string, expectedKeys []string) {
	t.Helper()
	ch := make(chan common.TelegrafMetric, chanCacheSize)
	acc := &MockAccumulator{}
	for _, m := range metrics {
		ch <- m
	}
	close(ch)
	consumeAndReport(acc, ch, devTagValue)
	for _, key := range expectedKeys {
		convey.So(acc.fields[key], convey.ShouldNotBeEmpty)
	}
}

// TestConsumeAndReport should aggregate and report metrics correctly
func TestConsumeAndReport(t *testing.T) {
	devTag := deviceTagKey + "=" + testDevTagValue
	device0Tag := deviceTagKey + "=" + testDevTagValue + "-0"
	singleField := map[string]interface{}{testFieldTemp: mockValue50}
	tests := []struct {
		name         string
		metrics      []common.TelegrafMetric
		expectedKeys []string
	}{
		{name: "should skip metric when Fields empty",
			metrics:      []common.TelegrafMetric{newDeviceMetric(testDeviceID0, map[string]interface{}{})},
			expectedKeys: nil},
		{name: "should aggregate device metrics into one series",
			metrics: []common.TelegrafMetric{
				newDeviceMetric(testDeviceID0, singleField),
				newDeviceMetric(testDeviceID0, map[string]interface{}{testFieldPower: mockValue50}),
			},
			expectedKeys: []string{metricKey(devName, device0Tag)}},
		{name: "should use custom measurement and labels",
			metrics: []common.TelegrafMetric{newLabeledMetric(
				testMeasCustom,
				map[string]string{testTagJob: testTagValTrain},
				map[string]interface{}{"loss": mockValue50})},
			expectedKeys: []string{metricKey(testMeasCustom, testTagJob+"="+testTagValTrain)}},
		{name: "should use default measurement when empty",
			metrics:      []common.TelegrafMetric{newDeviceMetric(common.NoDeviceID, singleField)},
			expectedKeys: []string{metricKey(devName, devTag)}},
		{name: "should separate vnpu from device by vdev_id tag",
			metrics: []common.TelegrafMetric{
				newDeviceMetric(testDeviceID0, singleField),
				newVnpuMetric(testDeviceID0, testVDevID100,
					map[string]interface{}{testFieldPower: mockValue50}),
			},
			expectedKeys: []string{
				metricKey(devName, device0Tag),
				metricKey(devName, device0Tag, vDevTagKey+"=100")}},
	}
	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			consumeAndVerify(t, tt.metrics, testDevTagValue, tt.expectedKeys)
		})
	}
}
