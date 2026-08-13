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

// Package cdi — spec_build_test.go
// Edge-case and scenario tests for BuildSpec.
package cdi

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	cdispec "tags.cncf.io/container-device-interface/specs-go"

	"ascend-common/cdi/mount"
)

// BuildSpec — edge cases

func TestBuildSpec_SingleDeviceEdgeCases(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	tests := []struct {
		name             string
		deviceIDs        []int
		wantPath         string
		checkSharedNodes bool
		wantSharedNodes  int
	}{
		{name: "logic ID 7", deviceIDs: []int{7}, wantPath: "/dev/davinci7", checkSharedNodes: true, wantSharedNodes: 4},
		{name: "logic ID 0", deviceIDs: []int{0}, wantPath: "/dev/davinci0"},
		{name: "very large logic ID", deviceIDs: []int{math.MaxInt64}, wantPath: fmt.Sprintf("/dev/davinci%d", math.MaxInt64)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: tt.deviceIDs, DevType: "Ascend910"}, Provider: &mockMountProvider{mounts: []*cdispec.Mount{}}})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if spec.Devices[0].ContainerEdits.DeviceNodes[0].Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", spec.Devices[0].ContainerEdits.DeviceNodes[0].Path, tt.wantPath)
			}
			if tt.checkSharedNodes && len(spec.ContainerEdits.DeviceNodes) != tt.wantSharedNodes {
				t.Errorf("got %d shared nodes, want %d", len(spec.ContainerEdits.DeviceNodes), tt.wantSharedNodes)
			}
		})
	}
}

func TestBuildSpec_SingleDeviceChecks(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	provider := &mockMountProvider{mounts: []*cdispec.Mount{}}
	edits, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{7}, DevType: Ascend910}, Provider: provider})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	perDev := edits.Devices[0].ContainerEdits.DeviceNodes
	shared := edits.ContainerEdits.DeviceNodes
	if len(perDev)+len(shared) != 5 {
		t.Fatalf("got %d device nodes, want 5", len(perDev)+len(shared))
	}
	if perDev[0].Path != "/dev/davinci7" {
		t.Errorf("per-device node = %q, want /dev/davinci7", perDev[0].Path)
	}
	if perDev[0].Major != devMajorDavinci {
		t.Errorf("Major = %d, want %d", perDev[0].Major, devMajorDavinci)
	}
}

func TestBuildSpec_RespectsMountFormat(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	spec, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0}, DevType: "Ascend910"}, Provider: newMockProvider([]*cdispec.Mount{{
		HostPath: "/host/src", ContainerPath: "/container/dst", Type: "bind", Options: []string{"ro", "nosuid"}},
	})})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spec.ContainerEdits.Mounts) < 1 {
		t.Fatal("expected at least 1 mount")
	}
	m := spec.ContainerEdits.Mounts[0]
	if m.ContainerPath != "/container/dst" {
		t.Errorf("ContainerPath = %q, want /container/dst", m.ContainerPath)
	}
	if m.HostPath != "/host/src" {
		t.Errorf("HostPath = %q, want /host/src", m.HostPath)
	}
	if m.Type != "bind" {
		t.Errorf("Type = %q, want bind", m.Type)
	}
}

func TestBuildSpec_NilProviderPanics(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic with nil provider")
		}
	}()
	BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0}, DevType: "Ascend910"}})
	t.Fatal("unreachable")
}

func TestBuildSpec_MultipleLogicIDs(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()
	spec, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{3, 4, 5}, DevType: "Ascend910A5"}, Provider: &mockMountProvider{mounts: []*cdispec.Mount{}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spec.Devices) != 3 {
		t.Fatalf("got %d devices, want 3", len(spec.Devices))
	}
	for i := 0; i < 3; i++ {
		wantPath := fmt.Sprintf("/dev/davinci%d", 3+i)
		if len(spec.Devices[i].ContainerEdits.DeviceNodes) != 1 ||
			spec.Devices[i].ContainerEdits.DeviceNodes[0].Path != wantPath {
			t.Errorf("device[%d] = %q", i, spec.Devices[i].ContainerEdits.DeviceNodes[0].Path)
		}
	}
}

func TestBuildSpec_NegativeLogicIDs(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	tests := []struct {
		name      string
		deviceIDs []int
	}{
		{name: "all negative", deviceIDs: []int{-1, -2, -3}},
		{name: "mixed valid and negative", deviceIDs: []int{5, 0, -7, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: tt.deviceIDs, DevType: "Ascend910"}, Provider: &mockMountProvider{mounts: []*cdispec.Mount{}}})
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

// BuildSpec — mount provider error

func TestBuildSpec_MountProviderError(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	wantErr := errors.New("mount provider failure")
	provider := &mockMountProvider{err: wantErr}
	_, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0}, DevType: Ascend910}, Provider: provider})
	if err == nil {
		t.Fatal("expected error from provider.GetMounts")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("error chain should contain %v, got %v", wantErr, err)
	}
}

// BuildSpec — HCCL topology

func TestBuildSpec_HCCLTopologyFilesExist(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()
	defer saveRestoreHCCLItems()()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "hccl_rootinfo.json")
	if err := os.WriteFile(tmpFile, []byte(`{"version":"1.0"}`), 0644); err != nil {
		t.Fatal(err)
	}
	mount.TopologyItems = []mount.TopologyItem{
		{HostPath: tmpFile, Options: []string{"rbind", "rprivate", "ro"}},
		{HostPath: filepath.Join(tmpDir, "no_such_dir"), Options: []string{"rbind", "rprivate", "ro"}},
	}
	provider := &mockMountProvider{
		mounts: []*cdispec.Mount{
			{HostPath: "/usr/lib64/libfoo.so", ContainerPath: "/usr/lib64/libfoo.so", Type: "bind", Options: []string{"ro"}},
		},
	}
	edits, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0}, DevType: Ascend910}, Provider: provider})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(edits.ContainerEdits.Mounts) != 2 {
		t.Fatalf("got %d mounts, want 2", len(edits.ContainerEdits.Mounts))
	}
	foundHCCL := false
	for _, m := range edits.ContainerEdits.Mounts {
		if m.ContainerPath == tmpFile {
			foundHCCL = true
			if m.HostPath != tmpFile {
				t.Errorf("HCCL mount HostPath = %q, want %q", m.HostPath, tmpFile)
			}
			if m.Type != "bind" {
				t.Errorf("HCCL mount type = %q, want bind", m.Type)
			}
			break
		}
	}
	if !foundHCCL {
		t.Error("HCCL topology mount not found")
	}
}

func TestBuildSpec_HCCLTopologyFilesMissing(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()
	defer saveRestoreHCCLItems()()

	mount.TopologyItems = []mount.TopologyItem{
		{HostPath: "/nonexistent/hccl_rootinfo.json", Options: []string{"rbind", "rprivate", "ro"}},
		{HostPath: "/nonexistent/topo", Options: []string{"rbind", "rprivate", "ro"}},
	}
	provider := &mockMountProvider{
		mounts: []*cdispec.Mount{
			{HostPath: "/usr/lib64/libfoo.so", ContainerPath: "/usr/lib64/libfoo.so", Type: "bind", Options: []string{"ro"}},
		},
	}
	edits, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0}, DevType: Ascend910}, Provider: provider})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(edits.ContainerEdits.Mounts) != 1 {
		t.Fatalf("got %d mounts, want 1", len(edits.ContainerEdits.Mounts))
	}
}
