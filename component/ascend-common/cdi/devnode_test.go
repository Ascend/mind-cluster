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

// Package cdi — devnode_test.go
//
// Tests for GeneratePerDeviceNodes.
package cdi

import (
	"testing"
)

// =============================================================================
// GeneratePerDeviceNodes — direct tests
// =============================================================================

func TestGeneratePerDeviceNodes_Empty(t *testing.T) {
	_, err := GeneratePerDeviceNodes([]int{}, false)
	if err == nil {
		t.Fatal("expected error for empty device IDs")
	}
}

func TestGeneratePerDeviceNodes_NegativeID(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	tests := []struct {
		name      string
		deviceIDs []int
	}{
		{name: "single negative", deviceIDs: []int{-1}},
		{name: "all negative", deviceIDs: []int{-1, -2, -3}},
		{name: "mixed valid and negative", deviceIDs: []int{5, 0, -7, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GeneratePerDeviceNodes(tt.deviceIDs, false)
			if err == nil {
				t.Fatal("expected error for negative logic ID")
			}
		})
	}
}

func TestGeneratePerDeviceNodes_Virtual(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	devices, err := GeneratePerDeviceNodes([]int{0, 1}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("got %d devices, want 2", len(devices))
	}
	if devices[0].ContainerEdits.DeviceNodes[0].HostPath != "/dev/vdavinci0" {
		t.Errorf("HostPath = %q, want /dev/vdavinci0", devices[0].ContainerEdits.DeviceNodes[0].HostPath)
	}
	if devices[0].ContainerEdits.DeviceNodes[0].Path != "/dev/davinci0" {
		t.Errorf("Path = %q, want /dev/davinci0", devices[0].ContainerEdits.DeviceNodes[0].Path)
	}
	if devices[1].ContainerEdits.DeviceNodes[0].HostPath != "/dev/vdavinci1" {
		t.Errorf("HostPath = %q, want /dev/vdavinci1", devices[1].ContainerEdits.DeviceNodes[0].HostPath)
	}
}
