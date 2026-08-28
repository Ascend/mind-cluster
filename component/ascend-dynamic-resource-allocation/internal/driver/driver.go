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

package driver

import (
	"context"
	"fmt"
	"strings"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/dynamic-resource-allocation/resourceslice"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"ascend-dynamic-resource-allocation/internal/device"
	draFlags "ascend-dynamic-resource-allocation/internal/flags"
	"ascend-dynamic-resource-allocation/internal/plugin"
)

// DraGenerationInterface converges everything that touches the device manager
// or varies by chip generation. The driver holds no dmgr: all device-level
// work (enumeration, per-device assembly, published attributes) lives here.
// Adding a generation means implementing this interface; driver common code
// never branches on devType.
type DraGenerationInterface interface {
	// SetDmgr hands the device manager to the generation. Called once by the
	// factory; the generation owns dmgr from then on.
	SetDmgr(devmanager.DeviceInterface)
	// ListNpuDevices enumerates and assembles every NPU device on the node.
	// Each generation picks its own enumeration scheme (e.g. GetDeviceList +
	// per-id BuildNpuDevice for 910, a different scheme for a future super-
	// pod generation). The driver only sees the resulting device list.
	ListNpuDevices() ([]device.NpuDevice, error)
	// DeviceAttributes returns the attributes published on the
	// ResourceSlice for a device. Centralising this here keeps publish
	// logic free of generation-specific attribute names.
	DeviceAttributes(dev device.NpuDevice) map[resourceapi.QualifiedName]resourceapi.DeviceAttribute
	// GetReleasedName returns the public released product name for this
	// generation (e.g. Ascend910ReleasedName). Used to build DeviceName and
	// any user-facing identifier; centralising it avoids per-generation
	// string constants leaking into driver common code.
	GetReleasedName() string
	GetDevType() string
	GetProductTypes() []string
}

// DraDriverInterface is the lifecycle surface every concrete driver fills.
// Construction is done through NewAscendDraDriver; the setter-based
// DraDriverSettingInterface is gone because its only reason for existing was
// multi-step construction, which a constructor collapses.
type DraDriverInterface interface {
	Start(ctx context.Context) error
	Stop()
}

// AscendDraDriver is the single shared driver skeleton. It wires a generation
// (device knowledge) to an AscendDraPlugin (Kubernetes protocol adapter) and
// drives the Start sequence. It deliberately holds no dmgr: the generation
// owns device access, the plugin owns k8s access, and the driver only
// orchestrates.
type AscendDraDriver struct {
	generation      DraGenerationInterface
	ascendDraPlugin *plugin.AscendDraPlugin
	draConfig       *draFlags.DRAConfig
	groupDevice     map[string][]*device.NpuDevice
	allInfo         device.NpuAllInfo
}

// NewAscendDraDriver is the only construction path. Replaces the previous
// SetDmgr/SetAscendDraPlugin/SetProductTypes setters: those existed only to
// let autoSetDraDriver build the struct in pieces, which a constructor does
// in one shot.
func NewAscendDraDriver(
	draConfig *draFlags.DRAConfig,
	generation DraGenerationInterface,
	ascendDraPlugin *plugin.AscendDraPlugin,
) *AscendDraDriver {
	return &AscendDraDriver{
		draConfig:       draConfig,
		generation:      generation,
		ascendDraPlugin: ascendDraPlugin,
	}
}

func (d *AscendDraDriver) startService(ctx context.Context) error {
	hwlog.RunLog.Info("registering kubelet plugin service")
	return d.ascendDraPlugin.RegisterService(ctx, d.draConfig)
}

func (d *AscendDraDriver) publishResources(ctx context.Context) error {
	hwlog.RunLog.Info("publishing ResourceSlice")
	resources := d.buildDriverResources()
	if err := d.ascendDraPlugin.PublishResources(ctx, resources); err != nil {
		return err
	}
	hwlog.RunLog.Infof("ResourceSlice published, pool=%s, deviceCount=%d",
		d.draConfig.DraOption.NodeName, len(d.allInfo.AllDevs))
	return nil
}

// buildDriverResources maps the discovered devices into the ResourceSlice
// contents. The per-device attributes come from the generation, so this
// function stays generation-agnostic: a future generation that exposes extra
// attributes (e.g. super-pod topology) only has to return them from
// DeviceAttributes instead of editing this code.
func (d *AscendDraDriver) buildDriverResources() resourceslice.DriverResources {
	devices := make([]resourceapi.Device, 0, len(d.allInfo.AllDevs))
	for _, dev := range d.allInfo.AllDevs {
		devices = append(devices, resourceapi.Device{
			Name:       strings.ToLower(dev.DeviceName),
			Attributes: d.generation.DeviceAttributes(*dev),
		})
	}
	return resourceslice.DriverResources{
		Pools: map[string]resourceslice.Pool{
			d.draConfig.DraOption.NodeName: {
				Slices: []resourceslice.Slice{
					{Devices: devices},
				},
			},
		},
	}
}

// Start runs the driver startup sequence: pull devices, register service,
// publish ResourceSlice, start healthz.
func (d *AscendDraDriver) Start(ctx context.Context) error {
	hwlog.RunLog.Info("starting ascend dra driver")
	// 1. pull npu info from generation
	if err := d.pullNPUInfo(); err != nil {
		return err
	}
	// 2. load ckpt
	// 3. start dra service
	if err := d.startService(ctx); err != nil {
		return err
	}
	// 4. publish resourceslice
	if err := d.publishResources(ctx); err != nil {
		return err
	}
	// 5. start healthz
	if err := d.startHealthCheck(ctx); err != nil {
		return err
	}
	hwlog.RunLog.Info("ascend dra driver started")
	return nil
}

func (d *AscendDraDriver) startHealthCheck(ctx context.Context) error {
	return d.ascendDraPlugin.StartHealthyCheck(ctx)
}

// Stop tears down the driver by stopping the underlying kubelet plugin.
func (d *AscendDraDriver) Stop() {
	hwlog.RunLog.Info("stopping ascend dra driver")
	d.ascendDraPlugin.Stop()
}

// pullNPUInfo is generation-agnostic: ask the generation for the device list,
// then classify and stash. No dmgr call, no per-generation branch.
func (d *AscendDraDriver) pullNPUInfo() error {
	hwlog.RunLog.Info("pulling NPU info from generation")
	devs, err := d.generation.ListNpuDevices()
	if err != nil {
		return err
	}
	allDevices := make([]*device.NpuDevice, 0, len(devs))
	allDeviceTypes := make([]string, 0, len(devs))
	for i := range devs {
		allDevices = append(allDevices, &devs[i])
		allDeviceTypes = append(allDeviceTypes, devs[i].DevType)
	}
	allDeviceTypes = removeDuplicate(&allDeviceTypes)
	d.groupDevice = device.ClassifyDevices(allDevices, allDeviceTypes)
	d.allInfo = device.NpuAllInfo{AllDevs: allDevices, AllDevTypes: allDeviceTypes}
	hwlog.RunLog.Infof("NPU info pulled, deviceCount=%d, devTypes=%v",
		len(d.allInfo.AllDevs), d.allInfo.AllDevTypes)
	if len(d.groupDevice) == 0 {
		return fmt.Errorf("no devices type found")
	}
	if len(d.allInfo.AllDevs) == 0 {
		return fmt.Errorf("no devices found")
	}
	return nil
}

func removeDuplicate(allDeviceTypes *[]string) []string {
	deviceTypesMap := make(map[string]string, len(*allDeviceTypes))
	var rmDupDeviceTypes []string
	for _, deviType := range *allDeviceTypes {
		deviceTypesMap[deviType] = deviType
	}
	for _, deviType := range deviceTypesMap {
		rmDupDeviceTypes = append(rmDupDeviceTypes, deviType)
	}
	return rmDupDeviceTypes
}

// AscendDraManager is the top-level driver manager. It owns a DraDriverInterface
// and wires a concrete generation to it at construction time.
type AscendDraManager struct {
	draDriver DraDriverInterface
}

// NewAscendDraManager creates a manager and auto-selects the driver generation.
func NewAscendDraManager(
	cancel context.CancelCauseFunc,
	dmgr devmanager.DeviceInterface,
	draConfig *draFlags.DRAConfig) (*AscendDraManager, error) {
	var adm AscendDraManager
	if err := adm.autoSetDraDriver(cancel, dmgr, draConfig); err != nil {
		return nil, fmt.Errorf("auto set dra driver failed, err: %v", err)
	}
	return &adm, nil
}

// autoSetDraDriver wires a concrete generation to the shared driver. dmgr is
// handed to the generation and never crosses into the driver.
func (adm *AscendDraManager) autoSetDraDriver(
	cancel context.CancelCauseFunc,
	dmgr devmanager.DeviceInterface,
	draConfig *draFlags.DRAConfig) error {
	devType := dmgr.GetDevType()
	var generation DraGenerationInterface
	switch devType {
	case api.Ascend910A, api.Ascend910B, api.Ascend910A3:
		generation = device.NewAscend910Generation()
	case api.Ascend910A5:
		generation = device.NewAscend950Generation()
	default:
		hwlog.RunLog.Errorf("found an unsupported draGen type: %v", devType)
		return fmt.Errorf("an unsupported draGen type: %v", devType)
	}
	hwlog.RunLog.Infof("dra generation selected, devType=%s", devType)
	generation.SetDmgr(dmgr)

	specMgr := plugin.NewCDISpecManager(
		generation.GetDevType(),
		generation.GetProductTypes(),
		draConfig.DraOption.CdiRoot,
	)

	ascendDraPlugin, err := plugin.NewAscendDraPlugin(draConfig, cancel, specMgr)
	if err != nil {
		return fmt.Errorf("new dra plugin err:%v", err)
	}
	adm.draDriver = NewAscendDraDriver(draConfig, generation, ascendDraPlugin)
	return nil
}

// Start starts the underlying driver.
func (adm *AscendDraManager) Start(ctx context.Context) error {
	return adm.draDriver.Start(ctx)
}

// Stop stops the underlying driver.
func (adm *AscendDraManager) Stop() {
	adm.draDriver.Stop()
}
