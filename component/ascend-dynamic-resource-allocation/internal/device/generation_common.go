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
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
)

const (
	// attrKeyType is the ResourceSlice attribute key for the device type,
	// which is always npu for NPU devices.
	attrKeyType = "type"
	// attrKeyPhysicID is the ResourceSlice attribute key for physical device ID.
	attrKeyPhysicID = "physicId"
	// attrKeyChipName is the ResourceSlice attribute key for the chip name
	// reported by dmgr.GetChipInfo.
	attrKeyChipName = "chipName"
)

// AscendCommonGeneration holds the device manager shared by every generation.
// Each concrete generation embeds this struct to obtain dmgr and SetDmgr
// without redeclaring them; only the genuinely per-generation logic lives in
// the embedding type.
type AscendCommonGeneration struct {
	dmgr devmanager.DeviceInterface
}

// SetDmgr satisfies DraGenerationInterface via embedding. The factory calls
// this once to hand the device manager to the generation.
func (c *AscendCommonGeneration) SetDmgr(d devmanager.DeviceInterface) {
	c.dmgr = d
}

// GetDevType returns the device type reported by the device manager.
func (c *AscendCommonGeneration) GetDevType() string {
	return c.dmgr.GetDevType()
}

// GetProductTypes returns the product type array reported by the device manager.
func (c *AscendCommonGeneration) GetProductTypes() []string {
	return c.dmgr.GetProductTypeArray()
}

// getChipName returns the chip name reported by dmgr.GetChipInfo for the given
// logic ID. On error or nil info it returns the empty string so DeviceAttributes
// always publishes a stable chipName attribute.
func (c *AscendCommonGeneration) getChipName(logicID int32) string {
	info, err := c.dmgr.GetChipInfo(logicID)
	if err != nil || info == nil {
		hwlog.RunLog.Warnf("get chip info failed, logicID=%d, err=%v, infoNil=%v", logicID, err, info == nil)
		return ""
	}
	return info.Name
}
