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

// Package cdi — spec_test.go
// Tests for BuildSpec and CDI spec validation.
package cdi

import (
	"fmt"
	"strconv"
	"testing"

	cdispec "tags.cncf.io/container-device-interface/specs-go"
)

// BuildSpec — core tests

func TestBuildSpec_Valid(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	spec, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0}, DevType: "Ascend910"}, Provider: newMockProvider([]*cdispec.Mount{{
		HostPath: "/usr/lib64/libfoo.so", ContainerPath: "/usr/lib64/libfoo.so", Type: "bind", Options: []string{"ro"}},
	})})
	if err != nil {
		t.Fatalf("BuildSpec error: %v", err)
	}
	if err := cdispec.ValidateVersion(spec); err != nil {
		t.Fatalf("version validation: %v", err)
	}
	if spec.Kind != "ascend.com/npu" {
		t.Errorf("Kind = %q", spec.Kind)
	}
	if len(spec.Devices) != 1 || spec.Devices[0].Name != "0" {
		t.Errorf("devices: want [0]")
	}
	if len(spec.Devices[0].ContainerEdits.DeviceNodes) != 1 ||
		spec.Devices[0].ContainerEdits.DeviceNodes[0].Path != "/dev/davinci0" {
		t.Error("per-device node mismatch")
	}
	sce := spec.ContainerEdits
	assertSpecHasNode(t, sce, "/dev/davinci_manager")
	assertSpecHasNode(t, sce, "/dev/dvpp_cmdlist")
	if len(sce.Mounts) == 0 {
		t.Error("Spec.ContainerEdits should have mounts")
	}
}

func TestBuildSpec_ComposesNodes(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	provider := &mockMountProvider{mounts: []*cdispec.Mount{
		{HostPath: "/usr/lib64/libfoo.so", ContainerPath: "/usr/lib64/libfoo.so", Type: "bind", Options: []string{"ro"}},
		{HostPath: "/var/run/bar", ContainerPath: "/var/run/bar", Type: "bind", Options: []string{"rbind", "rprivate", "ro"}},
	}}
	edits, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0, 1}, DevType: Ascend910}, Provider: provider})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	totalNodes := len(edits.ContainerEdits.DeviceNodes)
	for _, dev := range edits.Devices {
		totalNodes += len(dev.ContainerEdits.DeviceNodes)
	}
	if totalNodes != 6 {
		t.Fatalf("got %d total device nodes, want 6", totalNodes)
	}
	if edits.Devices[0].ContainerEdits.DeviceNodes[0].Path != "/dev/davinci0" {
		t.Errorf("Devices[0] node = %q", edits.Devices[0].ContainerEdits.DeviceNodes[0].Path)
	}
	if edits.Devices[1].ContainerEdits.DeviceNodes[0].Path != "/dev/davinci1" {
		t.Errorf("Devices[1] node = %q", edits.Devices[1].ContainerEdits.DeviceNodes[0].Path)
	}
	if len(edits.ContainerEdits.Mounts) < 2 {
		t.Fatalf("got %d mounts, want at least 2", len(edits.ContainerEdits.Mounts))
	}
	assertMountContains(t, edits.ContainerEdits.Mounts, "/usr/lib64/libfoo.so")
	assertMountContains(t, edits.ContainerEdits.Mounts, "/var/run/bar")
	if len(edits.ContainerEdits.Env) != 0 {
		t.Errorf("Env = %v, want empty", edits.ContainerEdits.Env)
	}
}

func TestBuildSpec_ErrorFromNodes(t *testing.T) {
	provider := &mockMountProvider{mounts: []*cdispec.Mount{}}
	_, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{}, DevType: Ascend910}, Provider: provider})
	if err != nil {
		t.Fatalf("empty deviceIDs should succeed: %v", err)
	}
}

func TestBuildSpec_InvalidDevType(t *testing.T) {
	tests := []struct {
		name      string
		deviceIDs []int
		devType   string
		wantErr   bool
		wantEmpty bool
	}{
		{"unknown with IDs", []int{0}, "UnknownType", true, false},
		{"fake with IDs", []int{0}, "FakeDevice", true, false},
		{"empty with IDs", []int{0}, "", true, false},
		{"fake with empty IDs", []int{}, "FakeDevice", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: tt.deviceIDs, DevType: tt.devType}, Provider: &mockMountProvider{mounts: []*cdispec.Mount{}}})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantEmpty && len(spec.Devices) != 0 {
				t.Fatal("expected zero devices")
			}
		})
	}
}

func TestBuildSpec_EmptyLogicIDs(t *testing.T) {
	spec, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{}, DevType: "Ascend910"}, Provider: &mockMountProvider{mounts: []*cdispec.Mount{}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spec.Devices) != 0 || spec.Version != "0.8.0" {
		t.Errorf("empty spec mismatch")
	}
}

func TestBuildSpec_MultipleDevices(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	spec, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0, 1}, DevType: "Ascend910"}, Provider: newMockProvider([]*cdispec.Mount{{
		HostPath: "/usr/lib64/libfoo.so", ContainerPath: "/usr/lib64/libfoo.so", Type: "bind", Options: []string{"ro"}},
	})})
	if err != nil {
		t.Fatalf("BuildSpec error: %v", err)
	}
	if len(spec.Devices) != 2 {
		t.Fatalf("got %d devices", len(spec.Devices))
	}
	for _, id := range []int{0, 1} {
		dev := spec.Devices[id]
		nodes := dev.ContainerEdits.DeviceNodes
		if len(nodes) != 1 || nodes[0].Path != "/dev/davinci"+strconv.Itoa(id) {
			t.Errorf("device %d: want [/dev/davinci%d]", id, id)
		}
		if len(dev.ContainerEdits.Mounts) != 0 {
			t.Error("device should have no mounts")
		}
	}
	sce := spec.ContainerEdits
	assertSpecHasNode(t, sce, "/dev/davinci_manager")
	assertSpecHasNode(t, sce, "/dev/dvpp_cmdlist")
	if len(sce.Mounts) == 0 {
		t.Error("shared mounts missing")
	}
}

// BuildSpec — error paths

func TestBuildSpec_ErrorFromProvider(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()
	_, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0}, DevType: "Ascend910"}, Provider: newMockProviderWithErr(fmt.Errorf("fail"))})
	if err == nil {
		t.Fatal("expected provider error")
	}
}

// validateSpec

func TestValidateSpec_DuplicateDeviceNames(t *testing.T) {
	spec := &cdispec.Spec{
		Version: "0.8.0",
		Kind:    "ascend.com/npu",
		Devices: []cdispec.Device{
			{Name: "0"},
			{Name: "0"},
		},
	}
	if err := validateSpec(spec); err == nil {
		t.Fatal("expected error for duplicate device names")
	}
}

func TestValidateSpec_EmptyDeviceName(t *testing.T) {
	spec := &cdispec.Spec{
		Version: "0.8.0",
		Kind:    "ascend.com/npu",
		Devices: []cdispec.Device{
			{Name: ""},
		},
	}
	if err := validateSpec(spec); err == nil {
		t.Fatal("expected error for empty device name")
	}
}

func TestValidateSpec_InvalidVersion(t *testing.T) {
	spec := &cdispec.Spec{
		Version: "999.0.0",
		Kind:    "ascend.com/npu",
		Devices: []cdispec.Device{
			{Name: "0"},
		},
	}
	if err := validateSpec(spec); err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestValidateSpec_NoDevices(t *testing.T) {
	spec := &cdispec.Spec{
		Version: "0.8.0",
		Kind:    "ascend.com/npu",
		Devices: nil,
	}
	if err := validateSpec(spec); err == nil {
		t.Fatal("expected error for no devices")
	}
}

// Helpers

func assertSpecHasNode(t *testing.T, ce cdispec.ContainerEdits, path string) {
	t.Helper()
	for _, dn := range ce.DeviceNodes {
		if dn.Path == path {
			return
		}
	}
	t.Errorf("Spec.ContainerEdits missing node %q", path)
}

func assertMountContains(t *testing.T, mounts []*cdispec.Mount, path string) {
	t.Helper()
	for _, m := range mounts {
		if m.ContainerPath == path {
			return
		}
	}
	t.Errorf("mount %q not found", path)
}
