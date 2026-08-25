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
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	cdispec "tags.cncf.io/container-device-interface/specs-go"

	"ascend-common/api"
	"ascend-common/cdi/mount"
)

// BuildSpec — core tests
func TestBuildSpec_Valid(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	mountDir := t.TempDir()
	libFile := filepath.Join(mountDir, "libfoo.so")
	if err := os.WriteFile(libFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	writeBaseList(t, mountDir, libFile)

	spec, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0}, DevType: "Ascend910"}, MountConfig: mount.MountConfig{Dir: mountDir, IsAscendDockerRuntime: true}})
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

	mountDir := t.TempDir()
	libFile := filepath.Join(mountDir, "libfoo.so")
	barFile := filepath.Join(mountDir, "bar")
	for _, p := range []string{libFile, barFile} {
		if err := os.WriteFile(p, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeBaseList(t, mountDir, libFile, barFile)

	edits, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0, 1}, DevType: api.Ascend910}, MountConfig: mount.MountConfig{Dir: mountDir, IsAscendDockerRuntime: true}})
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
	assertMountContains(t, edits.ContainerEdits.Mounts, libFile)
	assertMountContains(t, edits.ContainerEdits.Mounts, barFile)
	// Mount entries must not add env entries; only LD_LIBRARY_PATH from
	// collectAscendLibPaths may appear (host-dependent, e.g. /usr/lib64).
	for _, e := range edits.ContainerEdits.Env {
		if !strings.HasPrefix(e, ldLibraryPathKey+"=") {
			t.Errorf("Env = %v, want only LD_LIBRARY_PATH entries", edits.ContainerEdits.Env)
		}
	}
}

func TestBuildSpec_ErrorFromNodes(t *testing.T) {
	_, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{}, DevType: api.Ascend910}, MountConfig: testMountSource(t)})
	if err != nil {
		t.Fatalf("empty deviceIDs should succeed: %v", err)
	}
}

func TestBuildSpec_EmptyLogicIDs(t *testing.T) {
	spec, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{}, DevType: "Ascend910"}, MountConfig: testMountSource(t)})
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

	mountDir := t.TempDir()
	libFile := filepath.Join(mountDir, "libfoo.so")
	if err := os.WriteFile(libFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	writeBaseList(t, mountDir, libFile)

	spec, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0, 1}, DevType: "Ascend910"}, MountConfig: mount.MountConfig{Dir: mountDir, IsAscendDockerRuntime: true}})
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
func TestBuildSpec_ErrorFromMount(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	defer stubMountBuildFn(func(cfg mount.MountConfig, devType string) ([]*cdispec.Mount, error) {
		return nil, errors.New("fail")
	})()
	_, err := BuildSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0}, DevType: "Ascend910"}, MountConfig: testMountSource(t)})
	if err == nil {
		t.Fatal("expected mount error")
	}
}

// PrepareMountConfigFile
func TestPrepareMountConfigFile(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	if err := PrepareMountConfigFile(dirA); err != nil {
		t.Fatalf("PrepareMountConfigFile returned error: %v", err)
	}
	if err := mount.WriteMountProfile(dirB, mount.DefaultMountProfile()); err != nil {
		t.Fatalf("WriteMountProfile returned error: %v", err)
	}
	a, err := os.ReadFile(filepath.Join(dirA, "mounts.json"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dirB, "mounts.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Errorf("PrepareMountConfigFile output differs from mount.WriteMountProfile(dir, mount.DefaultMountProfile())")
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
