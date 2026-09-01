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
	"errors"
	"fmt"
	"sync"

	resourceapi "k8s.io/api/resource/v1"
	drapbv1 "k8s.io/kubelet/pkg/apis/dra/v1beta1"
	"k8s.io/kubernetes/pkg/kubelet/checkpointmanager"

	"ascend-common/common-utils/hwlog"
	"ascend-dynamic-resource-allocation/internal/flags"
	"ascend-dynamic-resource-allocation/pkg/consts"
)

// PreparedDevices is a slice of prepared devices for a claim.
type PreparedDevices []*PreparedDevice

// PreparedClaims maps claim UID to its prepared devices.
type PreparedClaims map[string]PreparedDevices

// PreparedDevice wraps the DRA device protocol type.
type PreparedDevice struct {
	drapbv1.Device
}

// GetDevices returns the underlying device slice.
func (pds PreparedDevices) GetDevices() []*drapbv1.Device {
	var devices []*drapbv1.Device
	for _, pd := range pds {
		devices = append(devices, &pd.Device)
	}
	return devices
}

// DeviceNames returns the DeviceName field of every prepared device. Used to
// hand the allocated device identity list to CdiSpecManager, which extracts
// the numeric suffix as the device ID for CDI spec generation.
func (pds PreparedDevices) DeviceNames() []string {
	names := make([]string, 0, len(pds))
	for _, pd := range pds {
		names = append(names, pd.DeviceName)
	}
	return names
}

// DeviceState owns the checkpoint and CDI spec lifecycle for prepared claims.
type DeviceState struct {
	sync.Mutex
	specs             CdiSpecInterface
	checkpointManager checkpointmanager.CheckpointManager
	checkpoint        *Checkpoint
	draOption         *flags.DRAOption
}

// NewDeviceState creates or reuses the checkpoint-backed device state.
func NewDeviceState(draOption *flags.DRAOption, specs CdiSpecInterface) (*DeviceState, error) {
	checkpointManager, err := checkpointmanager.NewCheckpointManager(draOption.DriverPluginPath())
	if err != nil {
		return nil, fmt.Errorf("unable to create checkpoint manager: %w", err)
	}

	state := &DeviceState{
		specs:             specs,
		checkpointManager: checkpointManager,
		draOption:         draOption,
	}

	checkpoints, err := state.checkpointManager.ListCheckpoints()
	if err != nil {
		return nil, fmt.Errorf("unable to list checkpoints: %w", err)
	}
	hwlog.RunLog.Debugf("[Checkpoints]: %v", checkpoints)
	checkpointExists := false
	for _, c := range checkpoints {
		if c == consts.DriverPluginCheckpointFile {
			checkpointExists = true
			break
		}
	}

	if !checkpointExists {
		checkpoint := newCheckpoint()
		if err := state.checkpointManager.CreateCheckpoint(consts.DriverPluginCheckpointFile, checkpoint); err != nil {
			return nil, fmt.Errorf("unable to sync to checkpoint: %w", err)
		}
		hwlog.RunLog.Info("checkpoint initialized")
	} else {
		hwlog.RunLog.Info("checkpoint already exists, reusing")
	}

	if err := state.LoadCheckpoint(); err != nil {
		return nil, err
	}
	return state, nil
}

// LoadCheckpoint reads the durable prepared-claim state into memory. It is
// called before the plugin is registered with kubelet so a corrupted
// checkpoint prevents the driver from serving requests.
func (s *DeviceState) LoadCheckpoint() error {
	s.Lock()
	defer s.Unlock()

	checkpoint := newCheckpoint()
	if err := s.checkpointManager.GetCheckpoint(consts.DriverPluginCheckpointFile, checkpoint); err != nil {
		return fmt.Errorf("unable to sync from checkpoint: %w", err)
	}
	if checkpoint.V1 == nil {
		return errors.New("checkpoint v1 payload is missing")
	}
	if checkpoint.V1.PreparedClaims == nil {
		checkpoint.V1.PreparedClaims = make(PreparedClaims)
	}
	s.checkpoint = checkpoint
	hwlog.RunLog.Infof("checkpoint loaded, preparedClaims=%d", len(checkpoint.V1.PreparedClaims))
	return nil
}

// Prepare allocates devices for a claim. Idempotent: a previously prepared
// claim returns the same prepared devices from the checkpoint. After
// allocating, the CDI spec is generated and the assignment is persisted.
func (s *DeviceState) Prepare(claim *resourceapi.ResourceClaim) ([]*drapbv1.Device, error) {
	s.Lock()
	defer s.Unlock()

	claimUID := string(claim.UID)

	if s.checkpoint == nil || s.checkpoint.V1 == nil {
		return nil, errors.New("checkpoint is not loaded")
	}
	preparedClaims := s.checkpoint.V1.PreparedClaims

	if preparedClaims[claimUID] != nil {
		hwlog.RunLog.Debugf("claim %v already prepared, reusing", claimUID)
		return preparedClaims[claimUID].GetDevices(), nil
	}

	preparedDevices, err := s.prepareDevices(claim)
	if err != nil {
		return nil, fmt.Errorf("prepare failed: %w", err)
	}

	cdiDeviceIDs, err := s.specs.WriteClaimSpec(claimUID, preparedDevices.DeviceNames())
	if err != nil {
		return nil, fmt.Errorf("unable to create CDI spec file for claim: %w", err)
	}
	for _, pd := range preparedDevices {
		pd.CdiDeviceIds = cdiDeviceIDs
	}

	preparedClaims[claimUID] = preparedDevices
	if err := s.checkpointManager.CreateCheckpoint(consts.DriverPluginCheckpointFile, s.checkpoint); err != nil {
		delete(preparedClaims, claimUID)
		if cleanupErr := s.specs.DeleteClaimSpec(claimUID); cleanupErr != nil {
			hwlog.RunLog.Warnf("rollback CDI spec for claim %v failed: %v", claimUID, cleanupErr)
		}
		return nil, fmt.Errorf("unable to sync to checkpoint: %w", err)
	}

	hwlog.RunLog.Infof("devices prepared for claim %v, count=%d, cdiIDs=%v",
		claimUID, len(preparedDevices), cdiDeviceIDs)
	return preparedClaims[claimUID].GetDevices(), nil
}

// Unprepare releases devices for a claim. Idempotent: for a claim missing
// from the checkpoint it still removes any leftover CDI spec, healing the
// orphan left behind when a previous Unprepare failed after the checkpoint
// was already persisted.
func (s *DeviceState) Unprepare(claimUID string) error {
	s.Lock()
	defer s.Unlock()

	if s.checkpoint == nil || s.checkpoint.V1 == nil {
		return errors.New("checkpoint is not loaded")
	}
	preparedClaims := s.checkpoint.V1.PreparedClaims

	if preparedClaims[claimUID] == nil {
		hwlog.RunLog.Debugf("claim %v not found in checkpoint, cleaning up leftover CDI spec", claimUID)
		// Deleting an absent spec is a no-op, so unknown claims stay harmless;
		// an orphaned one gets removed on kubelet's retry.
		if err := s.specs.DeleteClaimSpec(claimUID); err != nil {
			return fmt.Errorf("unable to delete CDI spec file for claim: %w", err)
		}
		return nil
	}

	devices := preparedClaims[claimUID]
	if err := s.unprepareDevices(claimUID, devices); err != nil {
		return fmt.Errorf("unprepare failed: %w", err)
	}

	// Persist the checkpoint first: on failure the CDI spec stays untouched,
	// so a retry can replay the full Unprepare flow.
	delete(preparedClaims, claimUID)
	if err := s.checkpointManager.CreateCheckpoint(consts.DriverPluginCheckpointFile, s.checkpoint); err != nil {
		preparedClaims[claimUID] = devices
		return fmt.Errorf("unable to sync to checkpoint: %w", err)
	}

	if err := s.specs.DeleteClaimSpec(claimUID); err != nil {
		return fmt.Errorf("unable to delete CDI spec file for claim: %w", err)
	}

	hwlog.RunLog.Infof("devices unprepared for claim %v", claimUID)
	return nil
}

func (s *DeviceState) prepareDevices(claim *resourceapi.ResourceClaim) (PreparedDevices, error) {
	if claim.Status.Allocation == nil {
		return nil, errors.New("claim not yet allocated")
	}

	// Walk through the device allocation results and construct the list of
	// prepared devices to return. CdiDeviceIds are filled in by the caller
	// after the CDI spec file has been written by CdiSpecManager.
	var preparedDevices PreparedDevices
	for _, result := range claim.Status.Allocation.Devices.Results {
		device := &PreparedDevice{
			Device: drapbv1.Device{
				RequestNames: []string{result.Request},
				PoolName:     result.Pool,
				DeviceName:   result.Device,
			},
		}
		preparedDevices = append(preparedDevices, device)
	}
	hwlog.RunLog.Debugf("prepared devices for claim '%v': %+v", claim.UID, preparedDevices)
	return preparedDevices, nil
}

func (s *DeviceState) unprepareDevices(claimUID string, devices PreparedDevices) error {
	return nil
}
