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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
