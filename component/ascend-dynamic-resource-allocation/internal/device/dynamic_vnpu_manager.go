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

package device

import (
	"fmt"

	"ascend-common/devmanager"
	devcommon "ascend-common/devmanager/common"
	"ascend-dynamic-resource-allocation/pkg/consts"
)

// DynamicVNPU records the DCMI identity created for one allocation.
type DynamicVNPU struct {
	CDIDeviceName string
	LogicID       int32
	VDevID        uint32
	TemplateName  string
}

// DynamicVNPUManager creates and destroys hard-sliced vNPUs for DRA claims.
type DynamicVNPUManager struct {
	dmgr devmanager.DeviceInterface
}

// NewDynamicVNPUManager creates a manager backed by the node device manager.
func NewDynamicVNPUManager(dmgr devmanager.DeviceInterface) *DynamicVNPUManager {
	return &DynamicVNPUManager{dmgr: dmgr}
}

// Create creates a vNPU and verifies its actual hardware AI-core allocation.
// A returned device remains caller-owned on verification failure and must be cleaned up.
func (m *DynamicVNPUManager) Create(phyID int32, template string, requestedAICore int64) (*DynamicVNPU, error) {
	if !devcommon.IsValidLogicIDOrPhyID(phyID) {
		return nil, fmt.Errorf("invalid physical device ID %d", phyID)
	}
	if requestedAICore <= 0 {
		return nil, fmt.Errorf("requested AI core capacity must be positive, got %d", requestedAICore)
	}
	logicID, err := m.dmgr.GetLogicIDFromPhysicID(phyID)
	if err != nil {
		return nil, fmt.Errorf("resolve logic ID for physical device %d: %w", phyID, err)
	}
	out, err := m.dmgr.CreateVirtualDevice(logicID, devcommon.CgoCreateVDevRes{
		VDevID: devcommon.InvalidID, VfgID: devcommon.InvalidID, TemplateName: template,
	})
	if err != nil {
		return nil, fmt.Errorf("create dynamic vNPU on physical device %d: %w", phyID, err)
	}
	if !devcommon.IsValidVDevID(out.VDevID) {
		return nil, fmt.Errorf("DCMI returned invalid vDevID %d", out.VDevID)
	}
	created := &DynamicVNPU{
		CDIDeviceName: fmt.Sprintf("%s%s%d%s%d", consts.DynamicVNPUDeviceKind, devcommon.Minus,
			out.VDevID, devcommon.Minus, phyID),
		LogicID: logicID, VDevID: out.VDevID, TemplateName: template,
	}
	info, err := m.inventory(logicID)
	if err != nil {
		return created, fmt.Errorf("verify created vNPU %d: %w", out.VDevID, err)
	}
	for _, current := range info.VDevInfo {
		if current.VDevID != out.VDevID {
			continue
		}
		created.TemplateName = current.QueryInfo.Name
		if vnpuInUse(current.QueryInfo) {
			return created, fmt.Errorf("created vNPU %d is already in use", out.VDevID)
		}
		if float64(current.QueryInfo.Computing.Aic) != float64(requestedAICore) {
			return created, fmt.Errorf("vNPU %d hardware template=%q AI cores=%v; requested template=%q AI cores=%d",
				out.VDevID, current.QueryInfo.Name, current.QueryInfo.Computing.Aic, template, requestedAICore)
		}
		return created, nil
	}
	return created, fmt.Errorf("created vNPU %d is missing from hardware inventory", out.VDevID)
}

// Destroy releases a dynamic vNPU. The caller retains checkpoint state until this succeeds.
func (m *DynamicVNPUManager) Destroy(logicID int32, vDevID uint32) error {
	info, err := m.inventory(logicID)
	if err != nil {
		return fmt.Errorf("verify dynamic vNPU %d before destroy: %w", vDevID, err)
	}
	for _, current := range info.VDevInfo {
		if current.VDevID == vDevID {
			if err := m.dmgr.DestroyVirtualDevice(logicID, vDevID); err != nil {
				return fmt.Errorf("destroy dynamic vNPU %d on logic device %d: %w", vDevID, logicID, err)
			}
			return nil
		}
	}
	return nil
}

// inventory rejects an unreadable or internally inconsistent DCMI response.
func (m *DynamicVNPUManager) inventory(logicID int32) (devcommon.VirtualDevInfo, error) {
	info, err := m.dmgr.GetVirtualDeviceInfo(logicID)
	if err != nil {
		return devcommon.VirtualDevInfo{}, err
	}
	return info, validateVNPUInventory(logicID, info)
}

// validateVNPUInventory rejects incomplete snapshots and invalid or duplicate device identities.
func validateVNPUInventory(logicID int32, info devcommon.VirtualDevInfo) error {
	if int(info.TotalResource.VDevNum) != len(info.VDevInfo) {
		return fmt.Errorf("inconsistent DCMI vNPU inventory on logic device %d", logicID)
	}
	seen := make(map[uint32]bool, len(info.VDevInfo))
	for _, current := range info.VDevInfo {
		if !devcommon.IsValidVDevID(current.VDevID) || seen[current.VDevID] {
			return fmt.Errorf("invalid or duplicate vNPU ID %d on logic device %d", current.VDevID, logicID)
		}
		seen[current.VDevID] = true
	}
	return nil
}
func vnpuInUse(q devcommon.CgoVDevQueryInfo) bool {
	// A2 reports 0xffffffff for a freshly created, unbound vNPU.
	return q.IsContainerUsed != 0 || q.ContainerID != 0 ||
		(q.Status != 0 && q.Status != devcommon.InvalidID)
}
