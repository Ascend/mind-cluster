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
	"strings"
	"sync"

	resourceapi "k8s.io/api/resource/v1"
	drapbv1 "k8s.io/kubelet/pkg/apis/dra/v1beta1"
	"k8s.io/kubernetes/pkg/kubelet/checkpointmanager"

	"ascend-common/common-utils/hwlog"
	"ascend-dynamic-resource-allocation/internal/device"
	"ascend-dynamic-resource-allocation/internal/flags"
	"ascend-dynamic-resource-allocation/pkg/consts"
)

// PreparedDevices is a slice of prepared devices for a claim.
type PreparedDevices []*PreparedDevice

// PreparedClaims maps claim UID to its prepared devices.
type PreparedClaims map[string]PreparedDevices

// DynamicVNPUInterface is the hardware lifecycle required by DeviceState.
type DynamicVNPUInterface interface {
	Create(phyID int32, template string, requestedAICore int64) (*device.DynamicVNPU, error)
	Destroy(logicID int32, vDevID uint32) error
}

// PreparedDevice wraps the DRA device protocol type.
type PreparedDevice struct {
	drapbv1.Device
	DynamicVNPU   bool   `json:"dynamicVNPU,omitempty"`
	CDIDeviceName string `json:"cdiDeviceName,omitempty"`
	ShareID       string `json:"shareID,omitempty"`
	LogicID       int32  `json:"logicID,omitempty"`
	PhyID         int32  `json:"phyID,omitempty"`
	VDevID        uint32 `json:"vDevID,omitempty"`
	TemplateName  string `json:"templateName,omitempty"`
}

// DeviceNames returns the CDI-facing identity of every prepared device. A
// dynamic vNPU uses its created-device name, which CdiSpecManager later scopes
// to the claim; physical and static devices use the scheduler allocation name.
func (pds PreparedDevices) DeviceNames() []string {
	names := make([]string, 0, len(pds))
	for _, pd := range pds {
		if pd.CDIDeviceName != "" {
			names = append(names, pd.CDIDeviceName)
		} else {
			names = append(names, pd.DeviceName)
		}
	}
	return names
}

// DeviceState owns the checkpoint and CDI spec lifecycle for prepared claims.
type DeviceState struct {
	sync.Mutex
	specs              CdiSpecInterface
	checkpointManager  checkpointmanager.CheckpointManager
	checkpoint         *Checkpoint
	draOption          *flags.DRAOption
	dynamicVNPUManager DynamicVNPUInterface
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

// ValidateMode rejects a checkpoint prepared in a different allocation mode.
func (s *DeviceState) ValidateMode(preset bool) error {
	s.Lock()
	defer s.Unlock()
	if s.checkpoint == nil || s.checkpoint.V1 == nil {
		return errors.New("checkpoint is not loaded")
	}

	for _, devices := range s.checkpoint.V1.PreparedClaims {
		for _, prepared := range devices {
			if !isPreparedDeviceModeCompatible(preset, prepared) {
				return fmt.Errorf("checkpoint device %s conflicts with presetVirtualDevice=%t", prepared.DeviceName, preset)
			}
		}
	}
	return nil
}

// isPreparedDeviceModeCompatible distinguishes preset assignments from dynamically created vNPUs.
func isPreparedDeviceModeCompatible(preset bool, preparedDevice *PreparedDevice) bool {
	return preset != preparedDevice.DynamicVNPU
}

// Prepare creates hardware, writes CDI and checkpoints; failures are cleaned up immediately.
func (s *DeviceState) Prepare(claim *resourceapi.ResourceClaim) (PreparedDevices, error) {
	s.Lock()
	defer s.Unlock()
	if claim == nil || claim.UID == "" {
		return nil, errors.New("claim UID is required")
	}
	if s.checkpoint == nil || s.checkpoint.V1 == nil {
		return nil, errors.New("checkpoint is not loaded")
	}
	claimUID := string(claim.UID)
	preparedClaims := s.checkpoint.V1.PreparedClaims
	if preparedClaims[claimUID] != nil {
		return preparedClaims[claimUID], nil
	}
	preparedDevices, err := s.prepareDevices(claim)
	if err != nil {
		return nil, s.rollbackPrepare(claimUID, preparedDevices, fmt.Errorf("prepare failed: %w", err))
	}
	if len(preparedDevices) == 0 {
		return nil, nil
	}
	cdiDeviceIDs, err := s.specs.WriteClaimSpec(claimUID, preparedDevices.DeviceNames())
	if err == nil && len(cdiDeviceIDs) != len(preparedDevices) {
		err = errors.New("CDI device count differs from prepared devices")
	}
	if err != nil {
		return nil, s.rollbackPrepare(claimUID, preparedDevices, err)
	}
	for i, pd := range preparedDevices {
		pd.CdiDeviceIds = []string{cdiDeviceIDs[i]}
	}
	preparedClaims[claimUID] = preparedDevices
	if err := s.checkpointManager.CreateCheckpoint(consts.DriverPluginCheckpointFile, s.checkpoint); err != nil {
		delete(preparedClaims, claimUID)
		return nil, s.rollbackPrepare(claimUID, preparedDevices, fmt.Errorf("unable to sync to checkpoint: %w", err))
	}
	return preparedDevices, nil
}

// rollbackPrepare reports each cleanup failure and preserves the original error.
func (s *DeviceState) rollbackPrepare(claimUID string, preparedDevices PreparedDevices, prepareErr error) error {
	cdiErr := s.specs.DeleteClaimSpec(claimUID)
	if cdiErr != nil {
		hwlog.RunLog.Errorf("rollback CDI spec for claim %s failed: %v", claimUID, cdiErr)
	}
	cleanupErr := s.unprepareDevices(preparedDevices)
	if cleanupErr != nil {
		hwlog.RunLog.Errorf("rollback devices for claim %s failed: %v", claimUID, cleanupErr)
	}
	return errors.Join(prepareErr, cdiErr, cleanupErr)
}

// Unprepare releases an assignment and retains ownership on any error.
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
	if err := s.unprepareDevices(devices); err != nil {
		return fmt.Errorf("unprepare failed: %w", err)
	}
	if err := s.specs.DeleteClaimSpec(claimUID); err != nil {
		return fmt.Errorf("unable to delete CDI spec file for claim: %w", err)
	}
	delete(preparedClaims, claimUID)
	if err := s.checkpointManager.CreateCheckpoint(consts.DriverPluginCheckpointFile, s.checkpoint); err != nil {
		preparedClaims[claimUID] = devices
		return fmt.Errorf("unable to sync to checkpoint: %w", err)
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
	identities := make(map[string]bool)
	for _, result := range claim.Status.Allocation.Devices.Results {
		if result.Driver != consts.DriverName {
			continue
		}
		preparedDevice := &PreparedDevice{
			Device: drapbv1.Device{
				RequestNames: []string{result.Request},
				PoolName:     result.Pool,
				DeviceName:   result.Device,
			},
		}
		if result.ShareID != nil {
			preparedDevice.ShareID = string(*result.ShareID)
		}
		template, err := dynamicVNPUTemplate(claim, result.Request)
		if err != nil {
			return preparedDevices, err
		}
		phyID, requestedAICore, err := s.validateDynamicAllocation(result, template)
		if err != nil {
			return preparedDevices, err
		}
		if template != "" {
			key := result.Pool + "/" + result.Device + "/" + preparedDevice.ShareID
			if identities[key] {
				return preparedDevices, fmt.Errorf("duplicate allocation %s", key)
			}
			identities[key] = true
			created, createErr := s.dynamicVNPUManager.Create(phyID, template, requestedAICore)
			if created == nil {
				if createErr == nil {
					createErr = errors.New("create returned no device")
				}
				return preparedDevices, createErr
			}
			preparedDevice.DynamicVNPU = true
			preparedDevice.CDIDeviceName = created.CDIDeviceName
			preparedDevice.LogicID = created.LogicID
			preparedDevice.PhyID = phyID
			preparedDevice.VDevID = created.VDevID
			preparedDevice.TemplateName = created.TemplateName
			if createErr != nil {
				return append(preparedDevices, preparedDevice), createErr
			}
		}
		preparedDevices = append(preparedDevices, preparedDevice)
	}
	hwlog.RunLog.Debugf("prepared devices for claim '%v': %+v", claim.UID, preparedDevices)
	return preparedDevices, nil
}

// validateDynamicAllocation checks the request before any hardware operation.
// Static allocations have no template and return without dynamic validation.
func (s *DeviceState) validateDynamicAllocation(
	result resourceapi.DeviceRequestAllocationResult, template string) (int32, int64, error) {
	capacity, hasAICoreCapacity := result.ConsumedCapacity[resourceapi.QualifiedName(consts.AICoreNum)]
	if s.draOption != nil && s.draOption.PresetVirtualDevice && (template != "" || hasAICoreCapacity) {
		return 0, 0, fmt.Errorf("dynamic vNPU allocation requires presetVirtualDevice=false")
	}
	if template == "" && hasAICoreCapacity {
		return 0, 0, fmt.Errorf("request %s consumed %s capacity without dynamic vNPU config",
			result.Request, consts.AICoreNum)
	}
	if template == "" {
		return 0, 0, nil
	}
	if s.dynamicVNPUManager == nil {
		return 0, 0, errors.New("dynamic vNPU manager is unavailable")
	}
	if result.ShareID == nil || *result.ShareID == "" {
		return 0, 0, fmt.Errorf("request %s has no share ID for dynamic vNPU allocation", result.Request)
	}
	requestedAICore, exact := capacity.AsInt64()
	if !hasAICoreCapacity || !exact || requestedAICore <= 0 {
		return 0, 0, fmt.Errorf("request %s has no positive integer consumed %s capacity",
			result.Request, consts.AICoreNum)
	}
	phyID, err := parsePhysicalNPUName(result.Device)
	return int32(phyID), requestedAICore, err
}

func (s *DeviceState) unprepareDevices(devices PreparedDevices) error {
	for _, prepared := range devices {
		if !prepared.DynamicVNPU {
			continue
		}
		if s.dynamicVNPUManager == nil {
			return errors.New("dynamic vNPU manager is unavailable")
		}
		if err := s.dynamicVNPUManager.Destroy(prepared.LogicID, prepared.VDevID); err != nil {
			hwlog.RunLog.Errorf("destroy vNPU logicID=%d vDevID=%d: %v", prepared.LogicID, prepared.VDevID, err)
			return err
		}
	}
	return nil
}

type dynamicVNPUConfig struct {
	Template string `json:"template"`
}

func dynamicVNPUTemplate(claim *resourceapi.ResourceClaim, request string) (string, error) {
	if claim.Status.Allocation == nil {
		return "", nil
	}
	var template string
	for _, config := range claim.Status.Allocation.Devices.Config {
		// The allocator records whether this config came from DeviceClass or ResourceClaim.
		if config.Source != resourceapi.AllocationConfigSourceClass &&
			config.Source != resourceapi.AllocationConfigSourceClaim {
			continue
		}
		if config.DeviceConfiguration.Opaque == nil ||
			config.DeviceConfiguration.Opaque.Driver != consts.DriverName || !matchesRequest(config.Requests, request) {
			continue
		}
		var decoded dynamicVNPUConfig
		if err := json.Unmarshal(config.DeviceConfiguration.Opaque.Parameters.Raw, &decoded); err != nil {
			return "", fmt.Errorf("decode dynamic vNPU config for request %s: %w", request, err)
		}
		decoded.Template = strings.TrimSpace(decoded.Template)
		if decoded.Template == "" {
			return "", fmt.Errorf("invalid dynamic vNPU config for request %s", request)
		}
		if template != "" && decoded.Template != template {
			return "", fmt.Errorf("conflicting dynamic vNPU templates for request %s: %q and %q",
				request, template, decoded.Template)
		}
		template = decoded.Template
	}
	return template, nil
}

func matchesRequest(configRequests []string, allocatedRequest string) bool {
	if len(configRequests) == 0 {
		return true
	}
	for _, configRequest := range configRequests {
		if configRequest == allocatedRequest || strings.HasPrefix(allocatedRequest, configRequest+"/") {
			return true
		}
	}
	return false
}
