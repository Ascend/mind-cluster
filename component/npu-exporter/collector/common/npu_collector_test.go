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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/collector/container/isula"
	"huawei.com/npu-exporter/v6/collector/container/v1"
	"huawei.com/npu-exporter/v6/utils/logger"
)

var (
	defaultLogFile = "/var/log/mindx-dl/npu-exporter/npu-exporter.log"
)

const (
	cacheTime         = 60 * time.Second
	timestamp         = 1606402
	waitTime          = 2 * time.Second
	npuCount          = 8
	time5s            = 5 * time.Second
	defaultUpdateTime = 5 * time.Second
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
	logger.HwLogConfig.LogLevel = 0
	logger.HwLogConfig.MaxBackups = hwlog.DefaultMaxBackups
	logger.HwLogConfig.LogFileName = defaultLogFile
	logger.HwLogConfig.MaxAge = hwlog.DefaultMinSaveAge
	logger.InitLogger("Prometheus")
}
