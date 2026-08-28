/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	resourceapi "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	drapbv1 "k8s.io/kubelet/pkg/apis/dra/v1beta1"

	draFlags "ascend-dynamic-resource-allocation/internal/flags"
	"ascend-dynamic-resource-allocation/pkg/consts"
)

// newDRAOption builds a DRAOption whose plugin data directory lives under dir.
func newDRAOption(dir string) *draFlags.DRAOption {
	return &draFlags.DRAOption{
		NodeName:                    "ut-node",
		KubeletPluginsDirectoryPath: dir,
	}
}

// multiDeviceClaim builds a claim with one allocation result per device.
func multiDeviceClaim(uid string, devices ...string) *resourceapi.ResourceClaim {
	results := make([]resourceapi.DeviceRequestAllocationResult, 0, len(devices))
	for i, device := range devices {
		results = append(results, resourceapi.DeviceRequestAllocationResult{
			Request: fmt.Sprintf("req-%d", i), Pool: "pool-a", Device: device,
		})
	}
	return &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{UID: types.UID(uid)},
		Status: resourceapi.ResourceClaimStatus{
			Allocation: &resourceapi.AllocationResult{
				Devices: resourceapi.DeviceAllocationResult{Results: results},
			},
		},
	}
}

// preparedDeviceSummaries extracts comparable "name:cdiIDs" pairs, since
// drapbv1.Device contains slices and cannot be compared with ==.
func preparedDeviceSummaries(devs []*drapbv1.Device) []string {
	summaries := make([]string, 0, len(devs))
	for _, d := range devs {
		summaries = append(summaries, fmt.Sprintf("%s:%v", d.DeviceName, d.CdiDeviceIds))
	}
	return summaries
}

// TestPreparedDevices_Helpers verifies GetDevices and DeviceNames on empty
// and populated slices.
func TestPreparedDevices_Helpers(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var pds PreparedDevices

		if got := pds.GetDevices(); len(got) != 0 {
			t.Errorf("GetDevices() len = %d, want 0", len(got))
		}
		if got := pds.DeviceNames(); len(got) != 0 {
			t.Errorf("DeviceNames() len = %d, want 0", len(got))
		}
	})

	t.Run("populated", func(t *testing.T) {
		pds := PreparedDevices{
			{Device: drapbv1.Device{DeviceName: "Ascend910-0", PoolName: "pool-a"}},
			{Device: drapbv1.Device{DeviceName: "Ascend910-1", PoolName: "pool-a"}},
		}

		devices := pds.GetDevices()
		if len(devices) != 2 {
			t.Fatalf("GetDevices() len = %d, want 2", len(devices))
		}
		wantNames := []string{"Ascend910-0", "Ascend910-1"}
		for i, dev := range devices {
			if dev.DeviceName != wantNames[i] {
				t.Errorf("GetDevices()[%d].DeviceName = %q, want %q", i, dev.DeviceName, wantNames[i])
			}
		}
		if got := pds.DeviceNames(); !reflect.DeepEqual(got, wantNames) {
			t.Errorf("DeviceNames() = %v, want %v", got, wantNames)
		}
	})
}

// TestNewDeviceState verifies state creation initializes the checkpoint file,
// a second construction on the same directory reuses the checkpoint, and
// construction fails when the plugin path collides with a regular file.
func TestNewDeviceState(t *testing.T) {
	t.Run("initializes checkpoint", func(t *testing.T) {
		dir := t.TempDir()

		state, err := NewDeviceState(newDRAOption(dir), &fakeCdiSpec{})

		if err != nil {
			t.Fatalf("NewDeviceState() error = %v, want nil", err)
		}
		if state == nil {
			t.Fatal("NewDeviceState() = nil, want state")
		}
		cpFile := filepath.Join(dir, consts.DriverName, consts.DriverPluginCheckpointFile)
		if _, err := os.Stat(cpFile); err != nil {
			t.Errorf("checkpoint file not created: %v", err)
		}
	})

	t.Run("reuses existing checkpoint", func(t *testing.T) {
		dir := t.TempDir()
		draOption := newDRAOption(dir)
		if _, err := NewDeviceState(draOption, &fakeCdiSpec{}); err != nil {
			t.Fatalf("first NewDeviceState() error = %v, want nil", err)
		}

		state, err := NewDeviceState(draOption, &fakeCdiSpec{})

		if err != nil {
			t.Fatalf("second NewDeviceState() error = %v, want nil", err)
		}
		if state == nil {
			t.Fatal("second NewDeviceState() = nil, want state")
		}
	})

	t.Run("plugin path is a file", func(t *testing.T) {
		dir := t.TempDir()
		notDir := filepath.Join(dir, consts.DriverName)
		if err := os.WriteFile(notDir, []byte("x"), 0600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		state, err := NewDeviceState(newDRAOption(dir), &fakeCdiSpec{})

		if err == nil {
			t.Fatalf("NewDeviceState() = %+v, want error", state)
		}
		if !strings.Contains(err.Error(), "unable to create checkpoint manager") {
			t.Errorf("NewDeviceState() error = %v, want containing %q",
				err, "unable to create checkpoint manager")
		}
	})
}

// TestDeviceState_Prepare_Idempotent verifies a repeated Prepare for the
// same claim returns the same devices and writes the CDI spec only once.
func TestDeviceState_Prepare_Idempotent(t *testing.T) {
	specs := &fakeCdiSpec{cdiIDs: []string{"ascend.com/npu=0"}}
	state := newTestDeviceState(t, specs)
	claim := allocatedClaim("uid-idem", "Ascend910-0")

	first, err := state.Prepare(claim)
	if err != nil {
		t.Fatalf("first Prepare() error = %v, want nil", err)
	}
	second, err := state.Prepare(claim)
	if err != nil {
		t.Fatalf("second Prepare() error = %v, want nil", err)
	}

	if len(specs.writes) != 1 {
		t.Errorf("WriteClaimSpec calls = %d, want 1 (idempotent)", len(specs.writes))
	}
	if !reflect.DeepEqual(preparedDeviceSummaries(first), preparedDeviceSummaries(second)) {
		t.Errorf("second Prepare() = %v, want same devices as first",
			preparedDeviceSummaries(second))
	}
}

// TestDeviceState_Prepare_PersistsAcrossRestart verifies a prepared claim
// survives a state reload: a second NewDeviceState on the same directory
// serves the claim from the checkpoint without rewriting the CDI spec. The
// reloaded state uses a different cdiIDs fake to prove the returned IDs come
// from the checkpoint, not a fresh write.
func TestDeviceState_Prepare_PersistsAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	draOption := newDRAOption(dir)
	specs := &fakeCdiSpec{cdiIDs: []string{"ascend.com/npu=0"}}
	state1, err := NewDeviceState(draOption, specs)
	if err != nil {
		t.Fatalf("first NewDeviceState() error = %v, want nil", err)
	}
	claim := allocatedClaim("uid-persist", "Ascend910-3")
	first, err := state1.Prepare(claim)
	if err != nil {
		t.Fatalf("Prepare() error = %v, want nil", err)
	}

	specs2 := &fakeCdiSpec{cdiIDs: []string{"ascend.com/npu=9"}}
	state2, err := NewDeviceState(draOption, specs2)
	if err != nil {
		t.Fatalf("second NewDeviceState() error = %v, want nil", err)
	}

	second, err := state2.Prepare(claim)
	if err != nil {
		t.Fatalf("Prepare() after restart error = %v, want nil", err)
	}
	if len(specs2.writes) != 0 {
		t.Errorf("WriteClaimSpec after restart = %v, want no writes", specs2.writes)
	}
	if !reflect.DeepEqual(preparedDeviceSummaries(first), preparedDeviceSummaries(second)) {
		t.Errorf("Prepare() after restart = %v, want checkpointed devices",
			preparedDeviceSummaries(second))
	}
}

// TestDeviceState_Prepare_MultipleDevices verifies a claim with several
// allocation results prepares one device per result, in order, all sharing
// the CDI IDs of the single claim spec.
func TestDeviceState_Prepare_MultipleDevices(t *testing.T) {
	specs := &fakeCdiSpec{cdiIDs: []string{"ascend.com/npu=0", "ascend.com/npu=1"}}
	state := newTestDeviceState(t, specs)
	claim := multiDeviceClaim("uid-multi", "Ascend910-0", "Ascend910-1", "Ascend910-2")

	devices, err := state.Prepare(claim)

	if err != nil {
		t.Fatalf("Prepare() error = %v, want nil", err)
	}
	if len(devices) != 3 {
		t.Fatalf("len(devices) = %d, want 3", len(devices))
	}
	wantNames := []string{"Ascend910-0", "Ascend910-1", "Ascend910-2"}
	for i, dev := range devices {
		if dev.DeviceName != wantNames[i] {
			t.Errorf("devices[%d].DeviceName = %q, want %q", i, dev.DeviceName, wantNames[i])
		}
		if len(dev.CdiDeviceIds) != 2 {
			t.Errorf("devices[%d].CdiDeviceIds = %v, want both CDI ids", i, dev.CdiDeviceIds)
		}
	}
	if got := specs.writes; len(got) != 1 || got[0] != "uid-multi" {
		t.Errorf("WriteClaimSpec calls = %v, want [uid-multi]", got)
	}
}

// TestDeviceState_Unprepare_Reprepare verifies an unprepared claim can be
// prepared again and gets a fresh CDI spec.
func TestDeviceState_Unprepare_Reprepare(t *testing.T) {
	specs := &fakeCdiSpec{cdiIDs: []string{"ascend.com/npu=0"}}
	state := newTestDeviceState(t, specs)
	claim := allocatedClaim("uid-rep", "Ascend910-0")

	if _, err := state.Prepare(claim); err != nil {
		t.Fatalf("first Prepare() error = %v, want nil", err)
	}
	if err := state.Unprepare("uid-rep"); err != nil {
		t.Fatalf("Unprepare() error = %v, want nil", err)
	}

	devices, err := state.Prepare(claim)

	if err != nil {
		t.Fatalf("second Prepare() error = %v, want nil", err)
	}
	if len(devices) != 1 || devices[0].DeviceName != "Ascend910-0" {
		t.Errorf("second Prepare() devices = %+v, want Ascend910-0", devices)
	}
	if want := []string{"uid-rep", "uid-rep"}; !reflect.DeepEqual(specs.writes, want) {
		t.Errorf("WriteClaimSpec calls = %v, want %v (fresh spec after unprepare)",
			specs.writes, want)
	}
	if want := []string{"uid-rep"}; !reflect.DeepEqual(specs.deletes, want) {
		t.Errorf("DeleteClaimSpec calls = %v, want %v", specs.deletes, want)
	}
}
