/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Description: dcmi DT Test
package dcmi

import (
	"context"
	"testing"

	"github.com/opencontainers/runtime-spec/specs-go"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

const mockDeviceID = 100

type mockWorker struct{}

func (w *mockWorker) Initialize() error {
	return nil
}

// ShutDown shutdown mock lib
func (w *mockWorker) ShutDown() {
	return
}

// CreateVDevice create v device
func (w *mockWorker) CreateVDevice(_ int32, _ string) (VDeviceInfo, error) {
	return VDeviceInfo{
		CardID:    0,
		DeviceID:  0,
		VdeviceID: mockDeviceID,
	}, nil
}

// DestroyVDevice destroy virtual device
func (w *mockWorker) DestroyVDevice(_, _ int32) error {
	return nil
}

// FindDevice find device by phyical id
func (w *mockWorker) FindDevice(_ int32) (int32, int32, error) {
	return 0, 0, nil
}

// GetProductType gets product type
func (w *mockWorker) GetProductType() (string, error) {
	return "", nil
}

// DcGetChipInfo gets chip info
func (w *mockWorker) GetChipName() (string, error) {
	return "", nil
}

// TestCreateVDevice tests the function CreateVDevice
func TestCreateVDevice(t *testing.T) {
	t.Log("TestCreateVDevice start")
	process := specs.Process{}
	spec := specs.Spec{Process: &process}
	spec.Process.Env = []string{}
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())

	vdevice, err := CreateVDevice(&mockWorker{}, nil, nil)
	if err == nil {
		t.Fatalf("%v %v", vdevice, err)
	}

	// no split, all ok
	deviceIdList := make([]int, 0)
	vdevice, err = CreateVDevice(&mockWorker{}, &spec, deviceIdList)
	if err != nil {
		t.Fatalf("%v %v", vdevice, err)
	}

	// no npu assigin for split
	spec.Process.Env = []string{api.AscendVnpuSpescEnv + "=vir04"}
	vdevice, err = CreateVDevice(&mockWorker{}, &spec, deviceIdList)
	if err == nil {
		t.Fatalf("%v %v", vdevice, err)
	}

	// split ok
	deviceIdList = []int{0}
	spec.Process.Env = []string{api.AscendVnpuSpescEnv + "=vir04", api.AscendVisibleDevicesEnv + "=0"}
	vdevice, err = CreateVDevice(&mockWorker{}, &spec, deviceIdList)
	if err != nil {
		t.Fatalf("%v %v", vdevice, err)
	}
	if vdevice.VdeviceID != mockDeviceID {
		t.Fatalf("%v %v", vdevice, err)
	}

}
