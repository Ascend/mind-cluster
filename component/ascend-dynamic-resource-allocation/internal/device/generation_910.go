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
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	"ascend-common/api"
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

// DCMI reports capacities as uint64, but ResourceSlice attributes are int64.
// Values above maxInt64ResourceValue are sentinels or corrupt data, not capacity.
const (
	// maxInt64ResourceValue is the largest uint64 that still fits in an int64
	// attribute. It is inclusive: this exact value is publishable.
	maxInt64ResourceValue uint64 = 1<<63 - 1
	// maxAICoreFloatBound is the exclusive float64 bound for DCMI AICore totals.
	// It is 2^63 rather than 2^63-1 because float64 cannot represent 2^63-1,
	// so the inclusive int64 limit would round up to this same value.
	maxAICoreFloatBound float64 = 1 << 63
)

func supportsDynamicVNPU(devType string) bool {
	return devType == api.Ascend910B || devType == api.Ascend910A3
}

// Ascend910Generation embeds AscendCommonGeneration for the shared dmgr field
// and SetDmgr; only 910-specific logic lives here.
type Ascend910Generation struct {
	AscendCommonGeneration
	PresetVirtualDevice bool
}

// NewAscend910Generation creates an Ascend910 generation instance.
func NewAscend910Generation() *Ascend910Generation {
	return &Ascend910Generation{PresetVirtualDevice: true}
}

// ListNpuDevices publishes physical devices or their pre-created static vNPUs.
func (g *Ascend910Generation) ListNpuDevices() ([]NpuDevice, error) {
	devNum, devList, err := g.dmgr.GetDeviceList()
	if err != nil {
		return nil, err
	}
	devs := make([]NpuDevice, 0, devNum)
	skippableInvalidDeviceCount := 0
	for i := int32(0); i < devNum; i++ {
		validDevices, foundSkippableInvalidDevice, err := g.collectNpuDevicesByLogicID(devList[i])
		if err != nil {
			return nil, err
		}
		if foundSkippableInvalidDevice {
			skippableInvalidDeviceCount++
			continue
		}
		devs = append(devs, validDevices...)
	}
	if skippableInvalidDeviceCount > 0 {
		hwlog.RunLog.Errorf("Ascend910 left %d physical devices out of the ResourceSlice", skippableInvalidDeviceCount)
	}
	hwlog.RunLog.Infof("Ascend910 enumerated %d devices", len(devs))
	return devs, nil
}

// getChipHBM reads the HBM capacity of one chip. Each chip is queried on its
// own so a mismatched reading is reported against the device it came from.
func (g *Ascend910Generation) getChipHBM(phyID int32, logicID int32) (uint64, error) {
	hbm, err := g.dmgr.GetDeviceHbmInfo(logicID)
	if err != nil {
		return 0, fmt.Errorf("query HBM on physical device %d: %w", phyID, err)
	}
	if hbm == nil {
		return 0, fmt.Errorf("query HBM returned empty memory info on physical device %d", phyID)
	}
	return hbm.MemorySize, nil
}

func (g *Ascend910Generation) collectNpuDevicesByLogicID(logicID int32) ([]NpuDevice, bool, error) {
	npuDev, err := g.buildNpuDevice(logicID)
	if err != nil {
		return nil, false, err
	}
	vDevInfo, queryErr := g.dmgr.GetVirtualDeviceInfo(logicID)
	if g.PresetVirtualDevice {
		// An unreadable inventory is logged, and this card is skipped.
		if queryErr != nil {
			hwlog.RunLog.Errorf("physical device %d is absent from the ResourceSlice: vNPU inventory is unknown: %v",
				npuDev.PhyID, queryErr)
			return nil, true, nil
		}
		// An invalid inventory is logged, and this card is skipped.
		if err := validateVNPUInventory(npuDev.LogicID, vDevInfo); err != nil {
			hwlog.RunLog.Errorf("physical device %d is absent from the ResourceSlice: %v", npuDev.PhyID, err)
			return nil, true, nil
		}
		// Preset mode with no vNPU: publish the unsliced card without resource totals.
		if vDevInfo.TotalResource.VDevNum == 0 {
			return g.collectPhyNpu(npuDev, vDevInfo)
		}
		// Preset mode with existing vNPUs: publish static slices, each with its own quota.
		return g.collectStaticVNPU(npuDev, vDevInfo)
	}
	// An unreadable or invalid inventory is returned, and enumeration stops.
	if queryErr != nil {
		return nil, false, fmt.Errorf("query checkpoint-owned vNPUs on logic device %d: %w", npuDev.LogicID, queryErr)
	}
	if err := validateVNPUInventory(npuDev.LogicID, vDevInfo); err != nil {
		return nil, false, err
	}
	// Dynamic mode: AICore comes from the pool. Memory is this chip's HBM capacity.
	return g.collectDynamicVNPU(npuDev, vDevInfo)
}

// collectPhyNpu publishes one unsliced card. It carries no allocation quota,
// so DeviceAttributes publishes no AICore, AICpu, or memory for it.
func (g *Ascend910Generation) collectPhyNpu(
	npuDev NpuDevice, _ devcommon.VirtualDevInfo) ([]NpuDevice, bool, error) {
	return []NpuDevice{npuDev}, false, nil
}

// collectStaticVNPU publishes pre-created slices. Each slice keeps its own quota.
func (g *Ascend910Generation) collectStaticVNPU(
	npuDev NpuDevice, info devcommon.VirtualDevInfo) ([]NpuDevice, bool, error) {
	staticVNPUDevs := g.buildStaticVNPUDevices(npuDev, info)
	if len(staticVNPUDevs) == 0 {
		hwlog.RunLog.Errorf("physical device %d is absent from the ResourceSlice: no static vNPU could be recorded",
			npuDev.PhyID)
		return nil, true, nil
	}
	return staticVNPUDevs, false, nil
}

// collectDynamicVNPU publishes an unsliced card. Each chip is queried for its own HBM capacity.
func (g *Ascend910Generation) collectDynamicVNPU(
	npuDev NpuDevice, info devcommon.VirtualDevInfo) ([]NpuDevice, bool, error) {
	if info.TotalResource.VDevNum > 0 {
		return nil, false, fmt.Errorf("unowned pre-created vNPUs on logic device %d in dynamic mode", npuDev.LogicID)
	}
	hbmMB, err := g.getChipHBM(npuDev.PhyID, npuDev.LogicID)
	if err != nil {
		return nil, false, err
	}
	if err := npuDev.setDynamicVNPUResources(info.TotalResource.Computing, hbmMB); err != nil {
		return nil, false, err
	}
	if supportsDynamicVNPU(npuDev.DevType) && npuDev.AICore <= 0 {
		return nil, false, fmt.Errorf("physical device %d reports invalid AI core capacity", npuDev.PhyID)
	}
	return []NpuDevice{npuDev}, false, nil
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
		if err := device.setStaticVNPUResources(vDev.QueryInfo.Computing); err != nil {
			hwlog.RunLog.Errorf("static vNPU %d on physical device %d is absent from the ResourceSlice: %v",
				vDev.VDevID, physical.PhyID, err)
			continue
		}
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
	if g.publishesCardResources(dev) {
		attributes[consts.AICoreNum] = resourceapi.DeviceAttribute{IntValue: ptr.To(dev.AICore)}
		attributes[consts.AICpuNum] = resourceapi.DeviceAttribute{IntValue: ptr.To(dev.AICpu)}
	}
	if g.publishesMemory() {
		attributes[consts.MemorySize] = resourceapi.DeviceAttribute{IntValue: ptr.To(dev.MemorySize)}
	}
	if dev.Kind == StaticVNPUDevice {
		attributes[attrKeyVDevID] = resourceapi.DeviceAttribute{IntValue: ptr.To(int64(dev.VDevID))}
		attributes[attrKeyTemplate] = resourceapi.DeviceAttribute{StringValue: ptr.To(dev.TemplateName)}
		attributes[attrKeyVNPUType] = resourceapi.DeviceAttribute{StringValue: ptr.To(dev.VNPUType)}
	}
	return attributes
}

// publishesCardResources reports whether AICore and AICpu are published.
// A preset unsliced card has no allocation quota, so it publishes neither.
// Static vNPUs and dynamic-mode cards always carry a quota.
func (g *Ascend910Generation) publishesCardResources(dev NpuDevice) bool {
	return !g.PresetVirtualDevice || dev.Kind != PhysicalDevice
}

// publishesMemory reports whether the memory attribute is published. Only
// dynamic-mode cards report HBM capacity. A preset whole card carries no quota,
// and a static vNPU does not publish memory at all.
func (g *Ascend910Generation) publishesMemory() bool {
	return !g.PresetVirtualDevice
}

// setStaticVNPUResources records the AICore and AICpu quota of one static slice.
// Memory is not recorded: a static vNPU never publishes memory.
func (dev *NpuDevice) setStaticVNPUResources(r devcommon.CgoComputingResource) error {
	if err := validateAICore(dev.PhyID, r.Aic); err != nil {
		return err
	}
	dev.AICore = int64(r.Aic)
	dev.AICpu = int64(r.DeviceAicpu)
	return nil
}

// setDynamicVNPUResources records a dynamic-mode card.
// AICore and AICpu come from the pool. memoryMB is the HBM capacity in MB.
// The pool memory in r is not published.
func (dev *NpuDevice) setDynamicVNPUResources(r devcommon.CgoComputingResource, memoryMB uint64) error {
	if err := validateAICore(dev.PhyID, r.Aic); err != nil {
		return err
	}
	if memoryMB > maxInt64ResourceValue {
		return fmt.Errorf("invalid HBM memory size %d on physical device %d", memoryMB, dev.PhyID)
	}
	dev.AICore = int64(r.Aic)
	dev.AICpu = int64(r.DeviceAicpu)
	dev.MemorySize = int64(memoryMB)
	return nil
}

func validateAICore(phyID int32, aic float32) error {
	if aic < 0 || float64(aic) >= maxAICoreFloatBound || float64(int64(aic)) != float64(aic) {
		return fmt.Errorf("invalid DCMI AICore totals on physical device %d", phyID)
	}
	return nil
}

// DeviceCapacity publishes capacity when DCMI reports it for an unsliced
// physical device. DCMI decides which templates the hardware accepts.
func (g *Ascend910Generation) DeviceCapacity(
	dev NpuDevice) map[resourceapi.QualifiedName]resourceapi.DeviceCapacity {
	if dev.Kind != PhysicalDevice || dev.AICore <= 0 {
		return nil
	}
	return map[resourceapi.QualifiedName]resourceapi.DeviceCapacity{
		consts.AICoreNum: {Value: *resource.NewQuantity(dev.AICore, resource.DecimalSI)},
	}
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
