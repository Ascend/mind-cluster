/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package process

import (
	"fmt"
	"os"
	"testing"

	"github.com/opencontainers/runtime-spec/specs-go"
	cdispec "tags.cncf.io/container-device-interface/specs-go"
)

const (
	devMajorDavinci = 64
	deviceNodePerm  = 0660
)

func cdidDeviceNode(path string, major, minor int64) *cdispec.DeviceNode {
	mode := os.FileMode(deviceNodePerm)
	return &cdispec.DeviceNode{Path: path, Type: "c", Major: major, Minor: minor, FileMode: &mode}
}

func cdidSpec(devices []cdispec.Device, mounts []*cdispec.Mount, env []string) *cdispec.Spec {
	return &cdispec.Spec{
		Version:        "0.8.0",
		Kind:           "ascend.com/npu",
		ContainerEdits: cdispec.ContainerEdits{Mounts: mounts, Env: env},
		Devices:        devices,
	}
}

func TestInjectEdits_NilSpec(t *testing.T) {
	err := InjectEdits(nil, cdidSpec(nil, nil, nil))
	if err == nil {
		t.Fatal("expected error for nil spec")
	}
}

func TestInjectEdits_NilCDISpec(t *testing.T) {
	spec := &specs.Spec{}
	if err := InjectEdits(spec, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInjectEdits_DevicesAndCgroup(t *testing.T) {
	spec := &specs.Spec{}
	cdSpec := cdidSpec(
		[]cdispec.Device{
			{Name: "0", ContainerEdits: cdispec.ContainerEdits{
				DeviceNodes: []*cdispec.DeviceNode{cdidDeviceNode("/dev/davinci0", devMajorDavinci, 0)},
			}},
		},
		nil, nil,
	)

	if err := InjectEdits(spec, cdSpec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.Linux == nil || spec.Linux.Resources == nil {
		t.Fatal("Linux/Resources not created")
	}
	if len(spec.Linux.Devices) != 1 || spec.Linux.Devices[0].Path != "/dev/davinci0" {
		t.Errorf("device not injected correctly")
	}
	if len(spec.Linux.Resources.Devices) != 1 {
		t.Errorf("cgroup device not injected")
	}
}

func TestInjectEdits_Mounts(t *testing.T) {
	spec := &specs.Spec{}
	cdSpec := cdidSpec(nil,
		[]*cdispec.Mount{
			&cdispec.Mount{HostPath: "/usr/local/Ascend", ContainerPath: "/usr/local/Ascend", Type: "bind", Options: []string{"rbind", "rprivate", "ro"}},
			&cdispec.Mount{HostPath: "/host/etc/foo.conf", ContainerPath: "/etc/foo.conf", Type: "bind", Options: []string{"ro"}},
		},
		nil,
	)

	if err := InjectEdits(spec, cdSpec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spec.Mounts) != 2 {
		t.Fatalf("got %d mounts, want 2", len(spec.Mounts))
	}
	// Sorted by path depth: /etc/foo.conf (depth 1) before /usr/local/Ascend (depth 2).
	if spec.Mounts[0].Destination != "/etc/foo.conf" {
		t.Errorf("mounts[0] = %q", spec.Mounts[0].Destination)
	}
	if spec.Mounts[1].Destination != "/usr/local/Ascend" {
		t.Errorf("mounts[1] = %q", spec.Mounts[1].Destination)
	}
}

func TestInjectEdits_MountDeduplication(t *testing.T) {
	spec := &specs.Spec{
		Mounts: []specs.Mount{
			{Destination: "/usr/local/Ascend", Type: "bind", Options: []string{"rbind", "rprivate", "ro"}},
		},
	}
	cdSpec := cdidSpec(nil,
		[]*cdispec.Mount{
			&cdispec.Mount{HostPath: "/usr/local/Ascend", ContainerPath: "/usr/local/Ascend", Type: "bind", Options: []string{"rbind"}},
			&cdispec.Mount{HostPath: "/new/mount", ContainerPath: "/new/mount", Type: "bind", Options: []string{"ro"}},
		},
		nil,
	)

	if err := InjectEdits(spec, cdSpec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spec.Mounts) != 2 {
		t.Fatalf("got %d mounts, want 2", len(spec.Mounts))
	}
	// Sorted: /new/mount (depth 1) before /usr/local/Ascend (depth 2).
	if spec.Mounts[0].Destination != "/new/mount" {
		t.Errorf("mounts[0] = %q", spec.Mounts[0].Destination)
	}
}

func TestInjectEdits_Env(t *testing.T) {
	spec := &specs.Spec{Process: &specs.Process{Env: []string{"FOO=bar"}}}
	cdSpec := cdidSpec(nil, nil, []string{"PATH=/custom/bin", "ASCEND_HOME=/usr/local/Ascend"})

	if err := InjectEdits(spec, cdSpec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Standard CDI Apply appends env edits after existing.
	if fmt.Sprintf("%v", spec.Process.Env) != "[FOO=bar PATH=/custom/bin ASCEND_HOME=/usr/local/Ascend]" {
		t.Errorf("env = %v", spec.Process.Env)
	}
}

func TestInjectEdits_EmptySpec(t *testing.T) {
	spec := &specs.Spec{}
	cdSpec := cdidSpec(
		[]cdispec.Device{
			{Name: "0", ContainerEdits: cdispec.ContainerEdits{
				DeviceNodes: []*cdispec.DeviceNode{cdidDeviceNode("/dev/davinci0", devMajorDavinci, 0)},
			}},
		},
		[]*cdispec.Mount{&cdispec.Mount{HostPath: "/usr/local/Ascend", ContainerPath: "/usr/local/Ascend", Type: "bind"}},
		[]string{"ASCEND_HOME=/usr/local/Ascend"},
	)

	if err := InjectEdits(spec, cdSpec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Linux == nil || len(spec.Linux.Devices) != 1 {
		t.Error("Linux/Devices not created")
	}
	if len(spec.Mounts) != 1 {
		t.Errorf("got %d mounts", len(spec.Mounts))
	}
	if spec.Process == nil || len(spec.Process.Env) != 1 {
		t.Error("Process/Env not created")
	}
}
