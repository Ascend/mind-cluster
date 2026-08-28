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
	"errors"
	"reflect"

	"github.com/agiledragon/gomonkey/v2"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/dynamic-resource-allocation/kubeletplugin"
	"k8s.io/dynamic-resource-allocation/resourceslice"
	"k8s.io/utils/ptr"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"ascend-dynamic-resource-allocation/internal/device"
	draFlags "ascend-dynamic-resource-allocation/internal/flags"
	"ascend-dynamic-resource-allocation/internal/plugin"
)

// initLog initialises hwlog so driver code that logs does not panic on a nil
// logger during tests.
func initLog() {
	hwLogConfig := hwlog.LogConfig{OnlyToStdout: true}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func init() { initLog() }

// errSentinel is a stable sentinel error reused by gomonkey stubs.
var errSentinel = errors.New("test sentinel error")

// fakeGeneration is a DraGenerationInterface stub whose behaviour is fully
// controlled by the test case fields. It lets the driver exercise every
// branch of pullNPUInfo, buildDriverResources and Start without touching
// hardware or the device package internals.
type fakeGeneration struct {
	devices      []device.NpuDevice
	listErr      error
	devType      string
	productTypes []string
	releasedName string
}

func (f *fakeGeneration) SetDmgr(devmanager.DeviceInterface) {}

func (f *fakeGeneration) ListNpuDevices() ([]device.NpuDevice, error) {
	return f.devices, f.listErr
}

func (f *fakeGeneration) DeviceAttributes(dev device.NpuDevice) map[resourceapi.QualifiedName]resourceapi.DeviceAttribute {
	return map[resourceapi.QualifiedName]resourceapi.DeviceAttribute{
		"deviceType": {StringValue: ptr.To(dev.DevType)},
		"physicId":   {IntValue: ptr.To(int64(dev.PhyID))},
	}
}

func (f *fakeGeneration) GetReleasedName() string   { return f.releasedName }
func (f *fakeGeneration) GetDevType() string        { return f.devType }
func (f *fakeGeneration) GetProductTypes() []string { return f.productTypes }

// newDriverWithFake builds an AscendDraDriver wired to a fakeGeneration and a
// minimal, non-nil plugin so promoted methods can dispatch. The returned
// plugin pointer is patched by the caller as needed.
func newDriverWithFake(gen *fakeGeneration) (*AscendDraDriver, *plugin.AscendDraPlugin) {
	cfg := &draFlags.DRAConfig{
		DraOption: &draFlags.DRAOption{NodeName: "test-node"},
	}
	adp := &plugin.AscendDraPlugin{
		ResourcePublisher: &plugin.ResourcePublisher{Helper: &kubeletplugin.Helper{}},
		DraHealthManager:  &plugin.DraHealthManager{},
	}
	return NewAscendDraDriver(cfg, gen, adp), adp
}

// pluginMethodPatches holds the four boundary-method patches the driver tests
// reuse. Each field may be nil when the test does not need to stub that method.
type pluginMethodPatches struct {
	registerService  *gomonkey.Patches
	publishResources *gomonkey.Patches
	startHealthCheck *gomonkey.Patches
	stop             *gomonkey.Patches
}

func (p *pluginMethodPatches) Reset() {
	if p.registerService != nil {
		p.registerService.Reset()
	}
	if p.publishResources != nil {
		p.publishResources.Reset()
	}
	if p.startHealthCheck != nil {
		p.startHealthCheck.Reset()
	}
	if p.stop != nil {
		p.stop.Reset()
	}
}

// patchPluginMethods stubs the four plugin boundary methods. The error
// arguments (which may be nil) are returned by the patched RegisterService,
// PublishResources and StartHealthyCheck; stopCalled records whether Stop was
// invoked. All four methods are always patched.
func patchPluginMethods(p *pluginMethodPatches, registerErr, publishErr, healthErr error, stopCalled *bool) {
	p.registerService = gomonkey.ApplyMethod(
		reflect.TypeOf(&plugin.AscendDraPlugin{}), "RegisterService",
		func(_ *plugin.AscendDraPlugin, _ context.Context, _ *draFlags.DRAConfig) error {
			return registerErr
		})
	p.publishResources = gomonkey.ApplyMethod(
		reflect.TypeOf(&plugin.ResourcePublisher{}), "PublishResources",
		func(_ *plugin.ResourcePublisher, _ context.Context, _ resourceslice.DriverResources) error {
			return publishErr
		})
	p.startHealthCheck = gomonkey.ApplyMethod(
		reflect.TypeOf(&plugin.DraHealthManager{}), "StartHealthyCheck",
		func(_ *plugin.DraHealthManager, _ context.Context) error {
			return healthErr
		})
	p.stop = gomonkey.ApplyMethod(
		reflect.TypeOf(&kubeletplugin.Helper{}), "Stop",
		func(_ *kubeletplugin.Helper) {
			if stopCalled != nil {
				*stopCalled = true
			}
		})
}
