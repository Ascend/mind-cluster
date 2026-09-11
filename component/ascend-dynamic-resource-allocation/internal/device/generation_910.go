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
	devcommon "ascend-common/devmanager/common"
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

// ListNpuDevices publishes physical devices or their pre-created static vNPUs.
func (g *Ascend910Generation) ListNpuDevices() ([]NpuDevice, error) {
	devNum, devList, err := g.dmgr.GetDeviceList()
	if err != nil {
		return nil, err
	}
	devs := make([]NpuDevice, 0, devNum)
	for i := int32(0); i < devNum; i++ {
		physical, err := g.buildNpuDevice(devList[i])
		if err != nil {
			return nil, err
		}
		vDevInfo, err := g.dmgr.GetVirtualDeviceInfo(devList[i])
		if err != nil {
			hwlog.RunLog.Warnf("The virtual device is considered not exist, please check the error: %v", err)
			devs = append(devs, physical)
			continue
		}
		if vDevInfo.TotalResource.VDevNum == 0 {
			devs = append(devs, physical)
			continue
		}
		virtualDevices := g.buildStaticVNPUDevices(physical, vDevInfo)
		devs = append(devs, virtualDevices...)
	}
	hwlog.RunLog.Infof("Ascend910 enumerated %d devices", len(devs))
	return devs, nil
}

func (g *Ascend910Generation) buildStaticVNPUDevices(
	physical NpuDevice, info devcommon.VirtualDevInfo) []NpuDevice {
	if int(info.TotalResource.VDevNum) != len(info.VDevInfo) {
		hwlog.RunLog.Warnf("logicID %d reports %d static vNPUs but returns %d details",
			physical.LogicID, info.TotalResource.VDevNum, len(info.VDevInfo))
	}
	devices := make([]NpuDevice, 0, len(info.VDevInfo))
	for _, vDev := range info.VDevInfo {
		if !devcommon.IsValidVDevID(vDev.VDevID) {
			hwlog.RunLog.Warnf("skip static vNPU on physical device %d with invalid vDevID %d",
				physical.PhyID, vDev.VDevID)
			continue
		}
		vnpuType, err := devcommon.GetVNPUTypeByTemplate(physical.DevType, vDev.QueryInfo.Name)
		if err != nil {
			hwlog.RunLog.Warnf("skip static vNPU %d on physical device %d: resolve type failed: %v",
				vDev.VDevID, physical.PhyID, err)
			continue
		}
		if vDev.QueryInfo.Computing.Aic <= 0 {
			hwlog.RunLog.Warnf("skip static vNPU %d on physical device %d with invalid AI core count %v",
				vDev.VDevID, physical.PhyID, vDev.QueryInfo.Computing.Aic)
			continue
		}
		device := physical
		device.Kind = StaticVNPUDevice
		device.VDevID = vDev.VDevID
		device.TemplateName = vDev.QueryInfo.Name
		device.VNPUType = vnpuType
		device.AICore = int64(vDev.QueryInfo.Computing.Aic)
		device.DeviceName = fmt.Sprintf("%s%s%d%s%d", StaticVNPUDevice, devcommon.Minus,
			vDev.VDevID, devcommon.Minus, physical.PhyID)
		devices = append(devices, device)
	}
	return devices
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
		DevType: g.dmgr.GetDevType(),
		// DeviceName follows the unified "<NPUNamePrefix>-<phyID>" convention across all generations.
		DeviceName: fmt.Sprintf("%s-%d", consts.NPUNamePrefix, phyID),
		IP:         ip,
		LogicID:    logicID,
		PhyID:      phyID,
		CardID:     cardID,
		DeviceID:   deviceID,
		Kind:       PhysicalDevice,
	}, nil
}

// DeviceAttributes publishes physical identity and static-vNPU metadata.
func (g *Ascend910Generation) DeviceAttributes(dev NpuDevice) map[resourceapi.QualifiedName]resourceapi.DeviceAttribute {
	attributes := map[resourceapi.QualifiedName]resourceapi.DeviceAttribute{
		attrKeyType:       {StringValue: ptr.To(consts.NPUNamePrefix)},
		attrKeyPhysicID:   {IntValue: ptr.To(int64(dev.PhyID))},
		attrKeyChipName:   {StringValue: ptr.To(g.getChipName(dev.LogicID))},
		attrKeyDeviceKind: {StringValue: ptr.To(string(dev.Kind))},
	}
	if dev.Kind == StaticVNPUDevice {
		attributes[attrKeyVDevID] = resourceapi.DeviceAttribute{IntValue: ptr.To(int64(dev.VDevID))}
		attributes[attrKeyTemplate] = resourceapi.DeviceAttribute{StringValue: ptr.To(dev.TemplateName)}
		attributes[attrKeyVNPUType] = resourceapi.DeviceAttribute{StringValue: ptr.To(dev.VNPUType)}
		attributes[attrKeyAICore] = resourceapi.DeviceAttribute{IntValue: ptr.To(dev.AICore)}
	}
	return attributes
}

// PhyIDToMountID is a no-op on 910 generations where devices mount by phyID,
// so the input is returned unchanged as the mount ID.
func (g *Ascend910Generation) PhyIDToMountID(phyID int32) (int32, error) {
	return phyID, nil
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
