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
	"strings"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/utils/ptr"

	"ascend-common/common-utils/hwlog"
	"ascend-dynamic-resource-allocation/pkg/consts"
)

const (
	// ipAddrTypeV4 is the IPv4 address family identifier used by dmgr.
	ipAddrTypeV4 = 0
	// ipAddrTypeV6 is the IPv6 address family identifier used by dmgr.
	ipAddrTypeV6 = 1
	// ipv6LinkTypePrefix is the IPv6 link-local prefix, filtered out as unusable.
	ipv6LinkTypePrefix = "fe80"
)

// Ascend910Generation embeds AscendCommonGeneration for the shared dmgr field
// and SetDmgr; only 910-specific logic lives here.
type Ascend910Generation struct {
	AscendCommonGeneration
}

// NewAscend910Generation creates an Ascend910 generation instance.
func NewAscend910Generation() *Ascend910Generation {
	return &Ascend910Generation{}
}

// GetReleasedName returns the released name for Ascend 910.
func (g *Ascend910Generation) GetReleasedName() string {
	return consts.Ascend910ReleasedName
}

// ListNpuDevices enumerates all 910 devices via dmgr.GetDeviceList and
// assembles each one. The driver sees only the resulting list.
func (g *Ascend910Generation) ListNpuDevices() ([]NpuDevice, error) {
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
	hwlog.RunLog.Infof("Ascend910 enumerated %d devices", len(devs))
	return devs, nil
}

// buildNpuDevice fills the full 910 device shape, including IP, CardID and
// DeviceID which are meaningful for this generation. Private: the driver
// never calls this, only ListNpuDevices does.
func (g *Ascend910Generation) buildNpuDevice(logicID int32) (NpuDevice, error) {
	phyID, err := g.dmgr.GetPhysicIDFromLogicID(logicID)
	if err != nil {
		return NpuDevice{}, err
	}
	cardID, deviceID, err := g.dmgr.GetCardIDDeviceID(logicID)
	if err != nil {
		return NpuDevice{}, err
	}
	ip, err := g.getDeviceIP(logicID)
	if err != nil {
		hwlog.RunLog.Warnf("get device ip failed, err: %v", err)
		ip = ""
	}
	return NpuDevice{
		DevType:    g.dmgr.GetDevType(),
		DeviceName: fmt.Sprintf("%s-%d", g.GetReleasedName(), phyID),
		IP:         ip,
		LogicID:    logicID,
		PhyID:      phyID,
		CardID:     cardID,
		DeviceID:   deviceID,
	}, nil
}

// DeviceAttributes publishes the deviceType and physicId for 910 devices.
func (g *Ascend910Generation) DeviceAttributes(dev NpuDevice) map[resourceapi.QualifiedName]resourceapi.DeviceAttribute {
	return map[resourceapi.QualifiedName]resourceapi.DeviceAttribute{
		attrKeyDeviceType: {StringValue: ptr.To(dev.DevType)},
		attrKeyPhysicID:   {IntValue: ptr.To(int64(dev.PhyID))},
	}
}

// getDeviceIP returns the first non-link-local device IP, preferring IPv4.
func (g *Ascend910Generation) getDeviceIP(logicID int32) (string, error) {
	deviceIp, err := g.dmgr.GetDeviceIPAddress(logicID, ipAddrTypeV4)
	if err == nil {
		return deviceIp, nil
	}
	deviceIp, err = g.dmgr.GetDeviceIPAddress(logicID, ipAddrTypeV6)
	if err != nil {
		return "", err
	}
	if strings.Index(deviceIp, ipv6LinkTypePrefix) == 0 {
		return "", fmt.Errorf("logicID(%d) ip %v is a link type ipv6 address", logicID, deviceIp)
	}
	return deviceIp, nil
}
