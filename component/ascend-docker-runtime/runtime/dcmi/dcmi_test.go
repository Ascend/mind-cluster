/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

package dcmi

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

const (
	// mockCardId0 represents a mock card ID 0 for testing
	mockCardId0 = 0
	// mockInvalidCardId represents an invalid card ID for testing
	mockInvalidCardId = -1
	// mockDeviceId0 represents a mock device ID 0 for testing
	mockDeviceId0 = 0
	// mockDeviceId1 represents a mock device ID 1 for testing
	mockDeviceId1 = 1
	// mockInvalidDeviceId represents an invalid device ID for testing
	mockInvalidDeviceId = -1
	// mockVDeviceId represents a mock virtual device ID for testing
	mockVDeviceId = 100
	// mockInvalidVDeviceId represents an invalid virtual device ID for testing
	mockInvalidVDeviceId = -1
	// mockUniqueId represents a mock unique ID for testing
	mockUniqueId = 10
	// mockCoreString represents a mock core string for testing
	mockCoreString = "vir04"
	// mockCardListLength represents the length of mock card list
	mockCardListLength = 2
	// mockDeviceNum represents the number of mock devices
	mockDeviceNum = 2
	// mockProductType represents a mock product type string
	mockProductType = "Ascend310"
	// mockA950ProductType represents a mock Ascend950 product type string
	mockA950ProductType = "Ascend950"
	// EmptyDeviceNum represents zero device count
	EmptyDeviceNum = 0
)

// mockCardList is a slice of mock card IDs for testing
var mockCardList = []int32{0, 1}

// mockChipInfo is a mock ChipInfo structure for testing Ascend310
var mockChipInfo = &ChipInfo{
	Name:    "Ascend310",
	Type:    "AI Processor",
	Version: "1.0",
}

// mockA950ChipInfo is a mock ChipInfo structure for testing Ascend950
var mockA950ChipInfo = &ChipInfo{
	Name:    "Ascend950",
	Type:    "AI Processor",
	Version: "2.0",
}

// MockDcV1Driver implements DcDriverV1Interface for testing
type MockDcV1Driver struct {
	// initializeFunc is a mock function for DcInitialize
	initializeFunc func() error
	// shutdownFunc is a mock function for DcShutDown
	shutdownFunc func()
	// getCardListFunc is a mock function for DcGetCardList
	getCardListFunc func() (int32, []int32, error)
	// getDeviceNumFunc is a mock function for DcGetDeviceNumInCard
	getDeviceNumFunc func(cardID int32) (int32, error)
	// getLogicIDFunc is a mock function for DcGetDeviceLogicID
	getLogicIDFunc func(cardID, deviceID int32) (int32, error)
	// getProductTypeFunc is a mock function for DcGetProductType
	getProductTypeFunc func(cardID, deviceID int32) (string, error)
	// createVDeviceFunc is a mock function for DcCreateVDevice
	createVDeviceFunc func(cardID, deviceID int32, coreNum string) (int32, error)
	// destroyVDeviceFunc is a mock function for DcDestroyVDevice
	destroyVDeviceFunc func(cardID, deviceID int32, vDevID int32) error
	// getChipInfoFunc is a mock function for DcGetChipInfo
	getChipInfoFunc func(cardID, deviceID int32) (*ChipInfo, error)
	// getDeviceLogicidFromPhyid is a mock function for DcGetDeviceLogicidFromPhyid
	getDeviceLogicidFromPhyid func(phyID int32) (int32, error)
}

// DcGetDeviceLogicidFromPhyid mocks the convert phy id to logic id function
func (m *MockDcV1Driver) DcGetDeviceLogicidFromPhyid(phyID int32) (int32, error) {
	if m.initializeFunc != nil {
		return m.getDeviceLogicidFromPhyid(phyID)
	}
	return 0, nil
}

// DcInitialize mocks the initialization function
func (m *MockDcV1Driver) DcInitialize() error {
	if m.initializeFunc != nil {
		return m.initializeFunc()
	}
	return nil
}

// DcShutDown mocks the shutdown function
func (m *MockDcV1Driver) DcShutDown() {
	if m.shutdownFunc != nil {
		m.shutdownFunc()
	}
}

// DcGetCardList mocks getting the card list
func (m *MockDcV1Driver) DcGetCardList() (int32, []int32, error) {
	if m.getCardListFunc != nil {
		return m.getCardListFunc()
	}
	return 0, []int32{}, nil
}

// DcGetDeviceNumInCard mocks getting the number of devices in a card
func (m *MockDcV1Driver) DcGetDeviceNumInCard(cardID int32) (int32, error) {
	if m.getDeviceNumFunc != nil {
		return m.getDeviceNumFunc(cardID)
	}
	return 0, nil
}

// DcGetDeviceLogicID mocks getting the logic ID of a device
func (m *MockDcV1Driver) DcGetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	if m.getLogicIDFunc != nil {
		return m.getLogicIDFunc(cardID, deviceID)
	}
	return 0, nil
}

// DcGetProductType mocks getting the product type
func (m *MockDcV1Driver) DcGetProductType(cardID, deviceID int32) (string, error) {
	if m.getProductTypeFunc != nil {
		return m.getProductTypeFunc(cardID, deviceID)
	}
	return "", nil
}

// DcCreateVDevice mocks creating a virtual device
func (m *MockDcV1Driver) DcCreateVDevice(cardID, deviceID int32, coreNum string) (int32, error) {
	if m.createVDeviceFunc != nil {
		return m.createVDeviceFunc(cardID, deviceID, coreNum)
	}
	return 0, nil
}

// DcDestroyVDevice mocks destroying a virtual device
func (m *MockDcV1Driver) DcDestroyVDevice(cardID, deviceID int32, vDevID int32) error {
	if m.destroyVDeviceFunc != nil {
		return m.destroyVDeviceFunc(cardID, deviceID, vDevID)
	}
	return nil
}

// DcGetChipInfo mocks getting chip information
func (m *MockDcV1Driver) DcGetChipInfo(cardID, deviceID int32) (*ChipInfo, error) {
	if m.getChipInfoFunc != nil {
		return m.getChipInfoFunc(cardID, deviceID)
	}
	return nil, nil
}

// MockDcV2Driver implements DcDriverV2Interface for testing
type MockDcV2Driver struct {
	// initializeFunc is a mock function for DcInitialize
	initializeFunc func() error
	// shutdownFunc is a mock function for DcShutDown
	shutdownFunc func()
	// getDeviceListFunc is a mock function for DcGetDeviceList
	getDeviceListFunc func() (int32, []int32, error)
	// createVDeviceFunc is a mock function for DcCreateVDevice
	createVDeviceFunc func(deviceID int32, coreNum string) (int32, error)
	// destroyVDeviceFunc is a mock function for DcDestroyVDevice
	destroyVDeviceFunc func(deviceID int32, vDevID int32) error
	// getChipInfoFunc is a mock function for DcGetChipInfo
	getChipInfoFunc func(deviceID int32) (*ChipInfo, error)
}

// DcInitialize mocks the initialization function for A950 driver
func (m *MockDcV2Driver) DcInitialize() error {
	if m.initializeFunc != nil {
		return m.initializeFunc()
	}
	return nil
}

// DcShutDown mocks the shutdown function for A950 driver
func (m *MockDcV2Driver) DcShutDown() {
	if m.shutdownFunc != nil {
		m.shutdownFunc()
	}
}

// DcGetDeviceList mocks getting the device list for A950
func (m *MockDcV2Driver) DcGetDeviceList() (int32, []int32, error) {
	if m.getDeviceListFunc != nil {
		return m.getDeviceListFunc()
	}
	return 0, []int32{}, nil
}

// DcCreateVDevice mocks creating a virtual device for A950
func (m *MockDcV2Driver) DcCreateVDevice(deviceID int32, coreNum string) (int32, error) {
	if m.createVDeviceFunc != nil {
		return m.createVDeviceFunc(deviceID, coreNum)
	}
	return 0, nil
}

// DcDestroyVDevice mocks destroying a virtual device for A950
func (m *MockDcV2Driver) DcDestroyVDevice(deviceID int32, vDevID int32) error {
	if m.destroyVDeviceFunc != nil {
		return m.destroyVDeviceFunc(deviceID, vDevID)
	}
	return nil
}

// DcGetChipInfo mocks getting chip information for A950
func (m *MockDcV2Driver) DcGetChipInfo(deviceID int32) (*ChipInfo, error) {
	if m.getChipInfoFunc != nil {
		return m.getChipInfoFunc(deviceID)
	}
	return nil, nil
}

// MockNpuV1Worker extends NpuV1Worker for testing with mock driver
type MockNpuV1Worker struct {
	*NpuV1Worker
}

// MockNpuV2Worker extends NpuV2Worker for testing with mock driver
type MockNpuV2Worker struct {
	*NpuV2Worker
}

// Test validation functions
// TestIsValidCardID tests the isValidCardID function
func TestIsValidCardID(t *testing.T) {
	tests := []struct {
		name   string
		cardID int32
		want   bool
	}{
		{"valid card id", 0, true},
		{"valid card id max", math.MaxInt32 - 1, true},
		{"valid card id boundary", math.MaxInt32, false},
		{"invalid negative", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidCardID(tt.cardID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsValidDeviceID tests the isValidDeviceID function
func TestIsValidDeviceID(t *testing.T) {
	tests := []struct {
		name     string
		deviceID int32
		want     bool
	}{
		{"valid device id min", 0, true},
		{"valid device id max", 3, true},
		{"invalid negative", -1, false},
		{"invalid too large", 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidDeviceID(tt.deviceID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsValidCardIDAndDeviceID tests the isValidCardIDAndDeviceID function
func TestIsValidCardIDAndDeviceID(t *testing.T) {
	tests := []struct {
		name     string
		cardID   int32
		deviceID int32
		want     bool
	}{
		{"both valid", 0, 0, true},
		{"card invalid", -1, 0, false},
		{"device invalid", 0, 4, false},
		{"both invalid", -1, 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidCardIDAndDeviceID(tt.cardID, tt.deviceID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsValidA950DeviceID tests the isValidA950DeviceID function
func TestIsValidA950DeviceID(t *testing.T) {
	tests := []struct {
		name     string
		deviceID int32
		want     bool
	}{
		{"valid device id min", 0, true},
		{"valid device id max", math.MaxInt32 - 1, true},
		{"valid boundary", math.MaxInt32, false},
		{"invalid negative", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidA950DeviceID(tt.deviceID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsValidChipInfo tests the isValidChipInfo function
func TestIsValidChipInfo(t *testing.T) {
	tests := []struct {
		name string
		chip *ChipInfo
		want bool
	}{
		{
			name: "all fields empty",
			chip: &ChipInfo{},
			want: false,
		},
		{
			name: "only name",
			chip: &ChipInfo{Name: "Ascend310"},
			want: true,
		},
		{
			name: "only type",
			chip: &ChipInfo{Type: "AI Processor"},
			want: true,
		},
		{
			name: "only version",
			chip: &ChipInfo{Version: "1.0"},
			want: true,
		},
		{
			name: "all fields",
			chip: &ChipInfo{Name: "Ascend310", Type: "AI Processor", Version: "1.0"},
			want: true,
		},
		{
			name: "nil chip",
			chip: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidChipInfo(tt.chip)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Test NpuV1Worker methods with mock
// TestNpuV1WorkerInitialize tests the Initialize method of NpuV1Worker
func TestNpuV1WorkerInitialize(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return nil
		},
	}
	worker := &NpuV1Worker{DcMgr: mockDriver}

	err := worker.Initialize()
	assert.NoError(t, err)

	// Test error case
	mockDriverWithErr := &MockDcV1Driver{
		initializeFunc: func() error {
			return errors.New("init failed")
		},
	}
	workerWithErr := &NpuV1Worker{DcMgr: mockDriverWithErr}
	err = workerWithErr.Initialize()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "init failed")
}

// TestNpuV1WorkerShutDown tests the ShutDown method of NpuV1Worker
func TestNpuV1WorkerShutDown(t *testing.T) {
	shutdownCalled := false
	mockDriver := &MockDcV1Driver{
		shutdownFunc: func() {
			shutdownCalled = true
		},
	}
	worker := &NpuV1Worker{DcMgr: mockDriver}

	worker.ShutDown()
	assert.True(t, shutdownCalled)
}

// TestNpuV1WorkerFindDeviceSuccess tests successful device finding in NpuV1Worker
func TestNpuV1WorkerFindDeviceSuccess(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		createVDeviceFunc: func(cardID, deviceID int32, coreNum string) (int32, error) {
			assert.Equal(t, int32(mockCardId0), cardID)
			assert.Equal(t, int32(mockDeviceId1), deviceID)
			assert.Equal(t, mockCoreString, coreNum)
			return mockVDeviceId, nil
		},
	}

	mockWorker := &MockNpuV1Worker{
		NpuV1Worker: &NpuV1Worker{DcMgr: mockDriver},
	}

	patchFindDevice := gomonkey.ApplyMethod(reflect.TypeOf(&NpuV1Worker{}), "FindDevice", func(f *NpuV1Worker, _ int32) (int32, int32, error) {
		return mockDeviceId1, mockCardId0, nil
	})
	defer patchFindDevice.Reset()

	vdevInfo, err := mockWorker.CreateVDevice(mockUniqueId, mockCoreString)
	assert.NoError(t, err)
	assert.Equal(t, int32(mockCardId0), vdevInfo.CardID)
	assert.Equal(t, int32(mockDeviceId1), vdevInfo.DeviceID)
	assert.Equal(t, int32(mockVDeviceId), vdevInfo.VdeviceID)
}

// TestNpuV1WorkerFindDeviceError tests device finding failure in NpuV1Worker
func TestNpuV1WorkerFindDeviceError(t *testing.T) {
	mockDriver := &MockDcV1Driver{}

	mockWorker := &MockNpuV1Worker{
		NpuV1Worker: &NpuV1Worker{DcMgr: mockDriver},
	}

	patchFindDevice := gomonkey.ApplyMethod(reflect.TypeOf(&NpuV1Worker{}), "FindDevice", func(f *NpuV1Worker, _ int32) (int32, int32, error) {
		return mockDeviceId0, mockCardId0, fmt.Errorf("find device failed")
	})
	defer patchFindDevice.Reset()

	_, err := mockWorker.CreateVDevice(mockUniqueId, mockCoreString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "find device failed")
}

// TestNpuV1WorkerCreateVDeviceSuccess tests successful virtual device creation in NpuV1Worker
func TestNpuV1WorkerCreateVDeviceSuccess(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		createVDeviceFunc: func(cardID, deviceID int32, coreNum string) (int32, error) {
			assert.Equal(t, int32(mockCardId0), cardID)
			assert.Equal(t, int32(mockDeviceId1), deviceID)
			assert.Equal(t, mockCoreString, coreNum)
			return mockVDeviceId, nil
		},
	}

	mockWorker := &MockNpuV1Worker{
		NpuV1Worker: &NpuV1Worker{DcMgr: mockDriver},
	}

	patchFindDevice := gomonkey.ApplyMethod(reflect.TypeOf(&NpuV1Worker{}), "FindDevice", func(f *NpuV1Worker, _ int32) (int32, int32, error) {
		return mockDeviceId1, mockCardId0, nil
	})
	defer patchFindDevice.Reset()

	vdevInfo, err := mockWorker.CreateVDevice(mockUniqueId, "vir04")
	assert.NoError(t, err)
	assert.Equal(t, int32(mockCardId0), vdevInfo.CardID)
	assert.Equal(t, int32(mockDeviceId1), vdevInfo.DeviceID)
	assert.Equal(t, int32(mockVDeviceId), vdevInfo.VdeviceID)
}

// TestNpuV1WorkerCreateVDeviceError tests virtual device creation failure in NpuV1Worker
func TestNpuV1WorkerCreateVDeviceError(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		createVDeviceFunc: func(cardID, deviceID int32, coreNum string) (int32, error) {
			return mockInvalidVDeviceId, errors.New("create failed")
		},
	}

	mockWorker := &MockNpuV1Worker{
		NpuV1Worker: &NpuV1Worker{DcMgr: mockDriver},
	}

	patchFindDevice := gomonkey.ApplyMethod(reflect.TypeOf(&NpuV1Worker{}), "FindDevice", func(f *NpuV1Worker, _ int32) (int32, int32, error) {
		return mockDeviceId1, mockCardId0, nil
	})
	defer patchFindDevice.Reset()

	vdevInfo, err := mockWorker.CreateVDevice(mockUniqueId, mockCoreString)
	assert.Error(t, err)
	assert.Equal(t, int32(mockInvalidCardId), vdevInfo.CardID)
	assert.Equal(t, int32(mockInvalidDeviceId), vdevInfo.DeviceID)
	assert.Equal(t, int32(math.MaxInt32), vdevInfo.VdeviceID)
}

// TestNpuV1WorkerDestroyVDeviceSuccess tests successful virtual device destruction in NpuV1Worker
func TestNpuV1WorkerDestroyVDeviceSuccess(t *testing.T) {
	destroyCalled := false
	mockDriver := &MockDcV1Driver{
		destroyVDeviceFunc: func(cardID, deviceID int32, vDevID int32) error {
			destroyCalled = true
			assert.Equal(t, int32(mockCardId0), cardID)
			assert.Equal(t, int32(mockDeviceId1), deviceID)
			assert.Equal(t, int32(mockVDeviceId), vDevID)
			return nil
		},
	}

	mockWorker := &MockNpuV1Worker{
		NpuV1Worker: &NpuV1Worker{DcMgr: mockDriver},
	}

	patchFindDevice := gomonkey.ApplyMethod(reflect.TypeOf(&NpuV1Worker{}), "FindDevice", func(f *NpuV1Worker, _ int32) (int32, int32, error) {
		return mockDeviceId1, mockCardId0, nil
	})
	defer patchFindDevice.Reset()

	err := mockWorker.DestroyVDevice(mockUniqueId, mockVDeviceId)
	assert.NoError(t, err)
	assert.True(t, destroyCalled)
}

// TestNpuV1WorkerDestroyVDeviceInvalidVDevID tests virtual device destruction with invalid VDevID
func TestNpuV1WorkerDestroyVDeviceInvalidVDevID(t *testing.T) {
	mockDriver := &MockDcV1Driver{}

	mockWorker := &MockNpuV1Worker{
		NpuV1Worker: &NpuV1Worker{DcMgr: mockDriver},
	}

	patchFindDevice := gomonkey.ApplyMethod(reflect.TypeOf(&NpuV1Worker{}), "FindDevice", func(f *NpuV1Worker, _ int32) (int32, int32, error) {
		return mockDeviceId1, mockCardId0, nil
	})
	defer patchFindDevice.Reset()

	err := mockWorker.DestroyVDevice(mockUniqueId, -1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "param error on vDevID")
}

// TestNpuV1WorkerDestroyVDeviceFindDeviceError tests virtual device destruction when device finding fails
func TestNpuV1WorkerDestroyVDeviceFindDeviceError(t *testing.T) {
	mockDriver := &MockDcV1Driver{}

	mockWorker := &MockNpuV1Worker{
		NpuV1Worker: &NpuV1Worker{DcMgr: mockDriver},
	}

	patchFindDevice := gomonkey.ApplyMethod(reflect.TypeOf(&NpuV1Worker{}), "FindDevice", func(f *NpuV1Worker, _ int32) (int32, int32, error) {
		return mockDeviceId0, mockCardId0, errors.New("find device failed")
	})
	defer patchFindDevice.Reset()

	err := mockWorker.DestroyVDevice(mockUniqueId, mockVDeviceId)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "find device failed")
}

// TestNpuV1WorkerGetProductTypeSuccess tests successful product type retrieval
func TestNpuV1WorkerGetProductTypeSuccess(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
		getCardListFunc: func() (int32, []int32, error) {
			return mockDeviceNum, mockCardList, nil
		},
		getDeviceNumFunc: func(cardID int32) (int32, error) {
			return mockDeviceNum, nil
		},
		getProductTypeFunc: func(cardID, deviceID int32) (string, error) {
			if cardID == 0 && deviceID == 0 {
				return "Ascend310", nil
			}
			return "", errors.New("not supported")
		},
	}

	worker := &NpuV1Worker{DcMgr: mockDriver}
	productType, err := worker.GetProductType()

	assert.NoError(t, err)
	assert.Equal(t, "Ascend310", productType)
}

// TestNpuV1WorkerGetProductTypeInitError tests product type retrieval with initialization failure
func TestNpuV1WorkerGetProductTypeInitError(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return errors.New("init failed")
		},
	}

	worker := &NpuV1Worker{DcMgr: mockDriver}
	_, err := worker.GetProductType()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot init dcmi")
}

// TestNpuV1WorkerGetProductTypeEmptyCardList tests product type retrieval with empty card list
func TestNpuV1WorkerGetProductTypeEmptyCardList(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
		getCardListFunc: func() (int32, []int32, error) {
			return 0, []int32{}, errors.New("get card list failed")
		},
	}

	worker := &NpuV1Worker{DcMgr: mockDriver}
	productType, err := worker.GetProductType()
	assert.Error(t, err)
	assert.Equal(t, "", productType)
}

// TestNpuV1WorkerGetProductTypeGetCardListError tests product type retrieval with card list error
func TestNpuV1WorkerGetProductTypeGetCardListError(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
		getCardListFunc: func() (int32, []int32, error) {
			return 0, nil, errors.New("get card list failed")
		},
	}

	worker := &NpuV1Worker{DcMgr: mockDriver}
	_, err := worker.GetProductType()
	assert.Error(t, err)
}

// TestNpuV1WorkerGetProductType tests the GetProductType method
func TestNpuV1WorkerGetProductType(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
		getCardListFunc: func() (int32, []int32, error) {
			return mockCardListLength, mockCardList, nil
		},
		getDeviceNumFunc: func(cardID int32) (int32, error) {
			return mockDeviceNum, nil
		},
		getProductTypeFunc: func(cardID, deviceID int32) (string, error) {
			return mockProductType, nil
		},
	}

	worker := &NpuV1Worker{DcMgr: mockDriver}
	productType, err := worker.GetProductType()
	assert.NoError(t, err)
	assert.Equal(t, mockProductType, productType)
}

// TestNpuV1WorkerGetChipNameSuccess tests successful chip name retrieval
func TestNpuV1WorkerGetChipNameSuccess(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
		getCardListFunc: func() (int32, []int32, error) {
			return mockCardListLength, mockCardList, nil
		},
		getDeviceNumFunc: func(cardID int32) (int32, error) {
			return mockDeviceNum, nil
		},
		getChipInfoFunc: func(cardID, deviceID int32) (*ChipInfo, error) {
			return mockChipInfo, nil
		},
	}

	worker := &NpuV1Worker{DcMgr: mockDriver}
	chipName, err := worker.GetChipName()

	assert.NoError(t, err)
	assert.Equal(t, mockProductType, chipName)
}

// TestNpuV1WorkerGetChipNameInitError tests chip name retrieval with initialization failure
func TestNpuV1WorkerGetChipNameInitError(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return errors.New("init failed")
		},
	}

	worker := &NpuV1Worker{DcMgr: mockDriver}
	_, err := worker.GetChipName()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot init dcmi")
}

// TestNpuV1WorkerGetChipNameGetCardListError tests chip name retrieval with card list error
func TestNpuV1WorkerGetChipNameGetCardListError(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
		getCardListFunc: func() (int32, []int32, error) {
			return EmptyDeviceNum, nil, errors.New("get card list failed")
		},
	}

	worker := &NpuV1Worker{DcMgr: mockDriver}
	_, err := worker.GetChipName()
	assert.Error(t, err)
}

// TestNpuV1WorkerGetChipNameNoCardFound tests chip name retrieval with no cards found
func TestNpuV1WorkerGetChipNameNoCardFound(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
		getCardListFunc: func() (int32, []int32, error) {
			return EmptyDeviceNum, []int32{}, nil
		},
	}

	worker := &NpuV1Worker{DcMgr: mockDriver}
	_, err := worker.GetChipName()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no card found")
}

// TestNpuV1WorkerGetChipNameSkipInvalidDevices tests chip name retrieval skipping invalid devices
func TestNpuV1WorkerGetChipNameSkipInvalidDevices(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
		getCardListFunc: func() (int32, []int32, error) {
			return mockCardListLength, mockCardList, nil
		},
		getDeviceNumFunc: func(cardID int32) (int32, error) {
			return mockDeviceNum, nil
		},
		getChipInfoFunc: func(cardID, deviceID int32) (*ChipInfo, error) {
			return mockChipInfo, nil
		},
	}

	worker := &NpuV1Worker{DcMgr: mockDriver}
	chipName, err := worker.GetChipName()
	assert.NoError(t, err)
	assert.Equal(t, mockProductType, chipName)
}

// TestNpuV1WorkerGetChipNameInvalidChipInfo tests chip name retrieval with invalid chip info
func TestNpuV1WorkerGetChipNameInvalidChipInfo(t *testing.T) {
	mockDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
		getCardListFunc: func() (int32, []int32, error) {
			return mockCardListLength, mockCardList, nil
		},
		getDeviceNumFunc: func(cardID int32) (int32, error) {
			return mockDeviceNum, nil
		},
		getChipInfoFunc: func(cardID, deviceID int32) (*ChipInfo, error) {
			return &ChipInfo{}, nil
		},
	}

	worker := &NpuV1Worker{DcMgr: mockDriver}
	_, err := worker.GetChipName()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot get valid chip info")
}

// Test NpuV2Worker methods
// TestNpuV2WorkerInitialize tests the Initialize method of NpuV2Worker
func TestNpuV2WorkerInitialize(t *testing.T) {
	mockDriver := &MockDcV2Driver{
		initializeFunc: func() error {
			return nil
		},
	}
	worker := &NpuV2Worker{DcMgr: mockDriver}

	err := worker.Initialize()
	assert.NoError(t, err)

	// Test error case
	mockDriverWithErr := &MockDcV2Driver{
		initializeFunc: func() error {
			return errors.New("init failed")
		},
	}
	workerWithErr := &NpuV2Worker{DcMgr: mockDriverWithErr}
	err = workerWithErr.Initialize()
	assert.Error(t, err)
}

// TestNpuV2WorkerShutDown tests the ShutDown method of NpuV2Worker
func TestNpuV2WorkerShutDown(t *testing.T) {
	shutdownCalled := false
	mockDriver := &MockDcV2Driver{
		shutdownFunc: func() {
			shutdownCalled = true
		},
	}
	worker := &NpuV2Worker{DcMgr: mockDriver}

	worker.ShutDown()
	assert.True(t, shutdownCalled)
}

// TestNpuV2WorkerCreateVDeviceError tests virtual device creation failure in NpuV2Worker
func TestNpuV2WorkerCreateVDeviceError(t *testing.T) {
	mockDriver := &MockDcV2Driver{
		createVDeviceFunc: func(deviceID int32, coreNum string) (int32, error) {
			return mockInvalidVDeviceId, errors.New("create failed")
		},
	}

	worker := &NpuV2Worker{DcMgr: mockDriver}
	vdevInfo, err := worker.CreateVDevice(mockUniqueId, "vir04")

	assert.Error(t, err)
	assert.Equal(t, int32(mockInvalidCardId), vdevInfo.VdeviceID)
	assert.Equal(t, int32(mockInvalidDeviceId), vdevInfo.CardID)
	assert.Equal(t, int32(mockInvalidVDeviceId), vdevInfo.DeviceID)
}

// TestNpuV2WorkerDestroyVDeviceError tests virtual device destruction failure in NpuV2Worker
func TestNpuV2WorkerDestroyVDeviceError(t *testing.T) {
	mockDriver := &MockDcV2Driver{
		destroyVDeviceFunc: func(deviceID int32, vDevID int32) error {
			return errors.New("destroy failed")
		},
	}

	worker := &NpuV2Worker{DcMgr: mockDriver}
	err := worker.DestroyVDevice(mockUniqueId, mockVDeviceId)
	assert.Error(t, err)
}

// TestNpuV2WorkerGetProductType tests the GetProductType method of NpuV2Worker
func TestNpuV2WorkerGetProductType(t *testing.T) {
	worker := &NpuV2Worker{}
	productType, err := worker.GetProductType()

	assert.NoError(t, err)
	assert.Equal(t, "[not support]", productType)
}

// TestNpuV2WorkerGetChipNameSuccess tests successful chip name retrieval in NpuV2Worker
func TestNpuV2WorkerGetChipNameSuccess(t *testing.T) {
	mockDriver := &MockDcV2Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
		getDeviceListFunc: func() (int32, []int32, error) {
			return mockDeviceNum, mockCardList, nil
		},
		getChipInfoFunc: func(deviceID int32) (*ChipInfo, error) {
			if deviceID == 0 {
				return mockA950ChipInfo, nil
			}
			return nil, errors.New("not supported")
		},
	}

	worker := &NpuV2Worker{DcMgr: mockDriver}
	chipName, err := worker.GetChipName()

	assert.NoError(t, err)
	assert.Equal(t, mockA950ProductType, chipName)
}

// TestNpuV2WorkerGetChipNameInitError tests chip name retrieval with initialization failure in NpuV2Worker
func TestNpuV2WorkerGetChipNameInitError(t *testing.T) {
	mockDriver := &MockDcV2Driver{
		initializeFunc: func() error {
			return errors.New("init failed")
		},
	}

	worker := &NpuV2Worker{DcMgr: mockDriver}
	_, err := worker.GetChipName()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot init dcmi")
}

// TestNpuV2WorkerGetChipNameGetDeviceListError tests chip name retrieval with device list error in NpuV2Worker
func TestNpuV2WorkerGetChipNameGetDeviceListError(t *testing.T) {
	mockDriver := &MockDcV2Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
		getDeviceListFunc: func() (int32, []int32, error) {
			return EmptyDeviceNum, nil, errors.New("get device list failed")
		},
	}

	worker := &NpuV2Worker{DcMgr: mockDriver}
	_, err := worker.GetChipName()
	assert.Error(t, err)
}

// TestNpuV2WorkerGetChipNameNoDeviceFound tests chip name retrieval with no devices found in NpuV2Worker
func TestNpuV2WorkerGetChipNameNoDeviceFound(t *testing.T) {
	mockDriver := &MockDcV2Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
		getDeviceListFunc: func() (int32, []int32, error) {
			return EmptyDeviceNum, []int32{}, nil
		},
	}

	worker := &NpuV2Worker{DcMgr: mockDriver}
	_, err := worker.GetChipName()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no card found")
}

// TestNpuV2WorkerGetChipNameSkipInvalidDevices tests chip name retrieval skipping invalid devices in NpuV2Worker
func TestNpuV2WorkerGetChipNameSkipInvalidDevices(t *testing.T) {
	mockDriver := &MockDcV2Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
		getDeviceListFunc: func() (int32, []int32, error) {
			return mockDeviceNum, mockCardList, nil
		},
		getChipInfoFunc: func(deviceID int32) (*ChipInfo, error) {
			return mockA950ChipInfo, nil
		},
	}

	worker := &NpuV2Worker{DcMgr: mockDriver}
	chipName, err := worker.GetChipName()
	assert.NoError(t, err)
	assert.Equal(t, mockA950ProductType, chipName)
}

// TestGetMatchingNpuV1Worker tests the GetMatchingNpuV1Worker function
func TestGetMatchingNpuV1Worker(t *testing.T) {
	// Save original managerList
	originalList := managerList
	defer func() { managerList = originalList }()

	// Test with valid manager
	mockDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
	}

	mockWorker := &NpuV1Worker{DcMgr: mockDriver}
	managerList = []WorkerInterface{mockWorker}

	worker, err := GetMatchingNpuWorker()
	assert.NoError(t, err)
	assert.NotNil(t, worker)

	// Test with no valid manager
	mockDriverWithErr := &MockDcV1Driver{
		initializeFunc: func() error {
			return errors.New("init failed")
		},
	}

	mockWorkerWithErr := &NpuV1Worker{DcMgr: mockDriverWithErr}
	managerList = []WorkerInterface{mockWorkerWithErr}

	_, err = GetMatchingNpuWorker()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find a valid manager")

	// Test with multiple managers, first one fails
	failDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return errors.New("first failed")
		},
	}

	successDriver := &MockDcV1Driver{
		initializeFunc: func() error {
			return nil
		},
		shutdownFunc: func() {},
	}

	managerList = []WorkerInterface{
		&NpuV1Worker{DcMgr: failDriver},
		&NpuV1Worker{DcMgr: successDriver},
	}

	worker, err = GetMatchingNpuWorker()
	assert.NoError(t, err)
	assert.NotNil(t, worker)
}
