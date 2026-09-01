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
	"encoding/json"
	"errors"
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
	"k8s.io/kubernetes/pkg/kubelet/checkpointmanager"
	"k8s.io/kubernetes/pkg/kubelet/checkpointmanager/checksum"

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

// fakeCheckpointManager delegates to a real manager but can be told to fail
// CreateCheckpoint to simulate checkpoint write failures.
type fakeCheckpointManager struct {
	checkpointmanager.CheckpointManager
	createErr error
}

func (f *fakeCheckpointManager) CreateCheckpoint(key string, cp checkpointmanager.Checkpoint) error {
	if f.createErr != nil {
		return f.createErr
	}
	return f.CheckpointManager.CreateCheckpoint(key, cp)
}

// newInjectableDeviceState builds a DeviceState whose checkpoint manager can
// fail on demand. The initial checkpoint file is created on disk first so
// LoadCheckpoint succeeds.
func newInjectableDeviceState(t *testing.T, specs CdiSpecInterface) (*DeviceState, *fakeCheckpointManager) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), consts.DriverName)
	realMgr, err := checkpointmanager.NewCheckpointManager(dir)
	if err != nil {
		t.Fatalf("NewCheckpointManager() error = %v", err)
	}
	if err := realMgr.CreateCheckpoint(consts.DriverPluginCheckpointFile, newCheckpoint()); err != nil {
		t.Fatalf("CreateCheckpoint() error = %v", err)
	}
	mgr := &fakeCheckpointManager{CheckpointManager: realMgr}
	state := &DeviceState{specs: specs, checkpointManager: mgr}
	if err := state.LoadCheckpoint(); err != nil {
		t.Fatalf("LoadCheckpoint() error = %v", err)
	}
	return state, mgr
}

// TestDeviceState_LoadCheckpoint_Corrupted verifies a corrupted checkpoint
// fails fast at load time instead of serving requests with bad state.
func TestDeviceState_LoadCheckpoint_Corrupted(t *testing.T) {
	t.Run("garbage bytes", func(t *testing.T) {
		dir := t.TempDir()
		if _, err := NewDeviceState(newDRAOption(dir), &fakeCdiSpec{}); err != nil {
			t.Fatalf("NewDeviceState() error = %v, want nil", err)
		}
		cpFile := filepath.Join(dir, consts.DriverName, consts.DriverPluginCheckpointFile)
		if err := os.WriteFile(cpFile, []byte("not-json"), 0600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		_, err := NewDeviceState(newDRAOption(dir), &fakeCdiSpec{})

		if err == nil {
			t.Fatal("NewDeviceState() = nil error, want load failure")
		}
		if !strings.Contains(err.Error(), "unable to sync from checkpoint") {
			t.Errorf("NewDeviceState() error = %v, want containing %q",
				err, "unable to sync from checkpoint")
		}
	})

	t.Run("missing v1 payload", func(t *testing.T) {
		tmp := t.TempDir()
		// Write a file whose v1 is explicitly null: Unmarshal then clears the
		// pre-initialized V1 of newCheckpoint(), so LoadCheckpoint hits its
		// "v1 payload is missing" branch. The checksum must stay valid for
		// the canonical nil-payload form ({"checksum":0}).
		canon, err := json.Marshal(Checkpoint{})
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		data := fmt.Sprintf(`{"checksum":%d,"v1":null}`, checksum.New(canon))
		dir := filepath.Join(tmp, consts.DriverName)
		if err := os.MkdirAll(dir, 0750); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		cpFile := filepath.Join(dir, consts.DriverPluginCheckpointFile)
		if err := os.WriteFile(cpFile, []byte(data), 0600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		_, err = NewDeviceState(newDRAOption(tmp), &fakeCdiSpec{})

		if err == nil {
			t.Fatal("NewDeviceState() = nil error, want load failure")
		}
		if !strings.Contains(err.Error(), "checkpoint v1 payload is missing") {
			t.Errorf("NewDeviceState() error = %v, want containing %q",
				err, "checkpoint v1 payload is missing")
		}
	})

	t.Run("tampered payload", func(t *testing.T) {
		tmp := t.TempDir()
		// Write a well-formed checkpoint, then mutate one payload byte
		// without refreshing the checksum: the JSON still parses, so only
		// VerifyChecksum (invoked inside GetCheckpoint) can reject it.
		cp := newCheckpoint()
		cp.V1.PreparedClaims["uid-x"] = PreparedDevices{
			{Device: drapbv1.Device{DeviceName: "Ascend910-0"}},
		}
		data, err := cp.MarshalCheckpoint()
		if err != nil {
			t.Fatalf("MarshalCheckpoint() error = %v", err)
		}
		tampered := strings.Replace(string(data), "Ascend910-0", "Ascend910-9", 1)
		if tampered == string(data) {
			t.Fatal("failed to tamper with the checkpoint payload")
		}
		dir := filepath.Join(tmp, consts.DriverName)
		if err := os.MkdirAll(dir, 0750); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		cpFile := filepath.Join(dir, consts.DriverPluginCheckpointFile)
		if err := os.WriteFile(cpFile, []byte(tampered), 0600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		_, err = NewDeviceState(newDRAOption(tmp), &fakeCdiSpec{})

		if err == nil {
			t.Fatal("NewDeviceState() = nil error, want checksum failure")
		}
		if !strings.Contains(err.Error(), "unable to sync from checkpoint") {
			t.Errorf("NewDeviceState() error = %v, want containing %q",
				err, "unable to sync from checkpoint")
		}
	})
}

// TestDeviceState_Prepare_RollbackOnCheckpointWriteFailure verifies a failed
// checkpoint write rolls back the in-memory claim and the CDI spec.
func TestDeviceState_Prepare_RollbackOnCheckpointWriteFailure(t *testing.T) {
	specs := &fakeCdiSpec{cdiIDs: []string{"ascend.com/npu=0"}}
	state, mgr := newInjectableDeviceState(t, specs)
	mgr.createErr = errors.New("disk full")
	claim := allocatedClaim("uid-prb", "Ascend910-0")

	devices, err := state.Prepare(claim)

	if err == nil {
		t.Fatalf("Prepare() = %+v devices, want error", devices)
	}
	if !strings.Contains(err.Error(), "unable to sync to checkpoint") {
		t.Errorf("Prepare() error = %v, want containing %q",
			err, "unable to sync to checkpoint")
	}
	if state.checkpoint.V1.PreparedClaims[string(claim.UID)] != nil {
		t.Error("claim still in preparedClaims after rollback, want removed")
	}
	if want := []string{"uid-prb"}; !reflect.DeepEqual(specs.deletes, want) {
		t.Errorf("DeleteClaimSpec calls = %v, want %v (CDI spec rolled back)",
			specs.deletes, want)
	}
}

// TestDeviceState_Unprepare_RollbackOnCheckpointWriteFailure verifies a
// failed checkpoint write keeps the claim in memory and leaves the CDI spec
// untouched, so a kubelet retry can replay the whole flow.
func TestDeviceState_Unprepare_RollbackOnCheckpointWriteFailure(t *testing.T) {
	specs := &fakeCdiSpec{cdiIDs: []string{"ascend.com/npu=0"}}
	state, mgr := newInjectableDeviceState(t, specs)
	claim := allocatedClaim("uid-urb", "Ascend910-0")
	if _, err := state.Prepare(claim); err != nil {
		t.Fatalf("Prepare() error = %v, want nil", err)
	}
	mgr.createErr = errors.New("disk full")

	err := state.Unprepare(string(claim.UID))

	if err == nil {
		t.Fatal("Unprepare() = nil error, want checkpoint write failure")
	}
	if !strings.Contains(err.Error(), "unable to sync to checkpoint") {
		t.Errorf("Unprepare() error = %v, want containing %q",
			err, "unable to sync to checkpoint")
	}
	if state.checkpoint.V1.PreparedClaims[string(claim.UID)] == nil {
		t.Error("claim removed from preparedClaims, want rolled back")
	}
	if len(specs.deletes) != 0 {
		t.Errorf("DeleteClaimSpec calls = %v, want none (CDI spec untouched)",
			specs.deletes)
	}
}

// TestDeviceState_Unprepare_HealsOrphanSpec verifies that unpreparing an
// unknown claim still deletes a leftover CDI spec: kubelet's retry lands on
// the no-op path and heals the orphan left behind when a previous Unprepare
// failed to delete the spec after the checkpoint was already persisted.
func TestDeviceState_Unprepare_HealsOrphanSpec(t *testing.T) {
	specs := &fakeCdiSpec{cdiIDs: []string{"ascend.com/npu=0"}}
	state, _ := newInjectableDeviceState(t, specs)

	err := state.Unprepare("uid-orphan")

	if err != nil {
		t.Fatalf("Unprepare() error = %v, want nil", err)
	}
	if want := []string{"uid-orphan"}; !reflect.DeepEqual(specs.deletes, want) {
		t.Errorf("DeleteClaimSpec calls = %v, want %v (orphan spec healed)",
			specs.deletes, want)
	}
}
