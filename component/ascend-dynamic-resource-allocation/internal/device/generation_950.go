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
	"ascend-dynamic-resource-allocation/pkg/consts"
	"fmt"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/utils/ptr"

	"ascend-common/common-utils/hwlog"
)

// Ascend950Generation embeds AscendCommonGeneration for the shared dmgr field
// and SetDmgr; only 950-specific logic lives here. Only logic/physical id and
// dev type are meaningful; CardID/DeviceID/IP stay zero-valued (no sentinel).
type Ascend950Generation struct {
	AscendCommonGeneration
}

// NewAscend950Generation creates an Ascend950 generation instance.
func NewAscend950Generation() *Ascend950Generation {
	return &Ascend950Generation{}
}

// ListNpuDevices enumerates all 950 devices via dmgr.GetDeviceList and
// assembles each one. The driver sees only the resulting list.
func (g *Ascend950Generation) ListNpuDevices() ([]NpuDevice, error) {
	devNum, devList, err := g.dmgr.GetDeviceList()
	if err != nil {
		return nil, err
	}
	devs := make([]NpuDevice, 0, devNum)
	for i := int32(0); i < devNum; i++ {
		dev, err := g.buildNpuDevice(devList[i])
		if err != nil {
			return nil, err
		}
		devs = append(devs, dev)
	}
	hwlog.RunLog.Infof("Ascend950 enumerated %d devices", len(devs))
	return devs, nil
}

// buildNpuDevice for 950: only logic/physical id and dev type are meaningful.
// Private: the driver never calls this, only ListNpuDevices does.
func (g *Ascend950Generation) buildNpuDevice(logicID int32) (NpuDevice, error) {
	phyID, err := g.dmgr.GetPhysicIDFromLogicID(logicID)
	if err != nil {
		return NpuDevice{}, err
	}
	return NpuDevice{
		DevType: g.dmgr.GetDevType(),
		// DeviceName follows the unified "<NPUNamePrefix>-<phyID>" convention across all generations.
		DeviceName: fmt.Sprintf("%s-%d", consts.NPUNamePrefix, phyID),
		LogicID:    logicID,
		PhyID:      phyID,
	}, nil
}

// DeviceAttributes publishes the type, physicId and chipName for 950 devices.
func (g *Ascend950Generation) DeviceAttributes(dev NpuDevice) map[resourceapi.QualifiedName]resourceapi.DeviceAttribute {
	return map[resourceapi.QualifiedName]resourceapi.DeviceAttribute{
		attrKeyType:     {StringValue: ptr.To(consts.NPUNamePrefix)},
		attrKeyPhysicID: {IntValue: ptr.To(int64(dev.PhyID))},
		attrKeyChipName: {StringValue: ptr.To(g.getChipName(dev.LogicID))},
	}
}

// PhyIDToMountID queries the device manager on every call so a 950 (A5) node
// picks up the latest phyID->logicID mapping even if device state changes after
// driver startup. On A5 the device node is keyed by logicID, so the mount ID is
// the logicID returned here.
func (g *Ascend950Generation) PhyIDToMountID(phyID int32) (int32, error) {
	return g.dmgr.GetLogicIDFromPhysicID(phyID)
}
