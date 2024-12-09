/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/opencontainers/runtime-spec/specs-go"

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
func (w *mockWorker) CreateVDevice(_, _ int32, _ string) (int32, error) {

	return int32(mockDeviceID), nil
}

// DestroyVDevice destroy virtual device
func (w *mockWorker) DestroyVDevice(_, _, _ int32) error {
	return nil
}

// FindDevice find device by phyical id
func (w *mockWorker) FindDevice(_ int32) (int32, int32, error) {
	return 0, 0, nil
}

// GetProductType gets product type
func (w *mockWorker) GetProductType(cardID, deviceID int32) (string, error) {
	return "", nil
}

// GetChipInfo gets chip info
func (w *mockWorker) GetChipInfo(cardID, deviceID int32) (*ChipInfo, error) {
	return &ChipInfo{}, nil
}

// TestCreateVDevice tests the function CreateVDevice
func TestCreateVDevice(t *testing.T) {
	t.Log("TestCreateVDevice start")
	process := specs.Process{}
	spec := specs.Spec{Process: &process}
	spec.Process.Env = []string{}
	backups := 2
	logMaxAge := 365
	fileMaxSize := 2
	runLogConfig := hwlog.LogConfig{
		LogFileName: "./test/run.log",
		LogLevel:    0,
		MaxBackups:  backups,
		FileMaxSize: fileMaxSize,
		MaxAge:      logMaxAge,
	}
	if err := hwlog.InitRunLogger(&runLogConfig, context.Background()); err != nil {
		t.Fatalf("hwlog init failed, error is %v", err)
	}

	// no split, all ok
	deviceIdList := make([]int, 0)
	vdevice, err := CreateVDevice(&mockWorker{}, &spec, deviceIdList)
	if err != nil {
		t.Fatalf("%v %v", vdevice, err)
	}

	// no npu assigin for split
	spec.Process.Env = []string{"ASCEND_VNPU_SPECS=vir04"}
	vdevice, err = CreateVDevice(&mockWorker{}, &spec, deviceIdList)
	if err == nil {
		t.Fatalf("%v %v", vdevice, err)
	}

	// split ok
	deviceIdList = []int{0}
	spec.Process.Env = []string{"ASCEND_VNPU_SPECS=vir04", "ASCEND_VISIBLE_DEVICES=0"}
	vdevice, err = CreateVDevice(&mockWorker{}, &spec, deviceIdList)
	if err != nil {
		t.Fatalf("%v %v", vdevice, err)
	}
	if vdevice.VdeviceID != mockDeviceID {
		t.Fatalf("%v %v", vdevice, err)
	}

}

// TestGetChipName tests the function GetChipName
func TestGetChipName(t *testing.T) {
	patchInitialize := gomonkey.ApplyMethod(reflect.TypeOf(&NpuWorker{}), "Initialize", func(f *NpuWorker) error {
		return nil
	})
	defer patchInitialize.Reset()
	patchShutDown := gomonkey.ApplyMethod(reflect.TypeOf(&NpuWorker{}), "ShutDown", func(f *NpuWorker) {
		return
	})
	defer patchShutDown.Reset()
	patch := gomonkey.ApplyFunc(GetCardList, func() (int32, []int32, error) {
		return 1, []int32{0}, nil
	})
	defer patch.Reset()
	patchGetDeviceNumInCard := gomonkey.ApplyFunc(GetDeviceNumInCard, func(cardID int32) (int32, error) {
		return 1, nil
	})
	defer patchGetDeviceNumInCard.Reset()
	patchGetChipInfo := gomonkey.ApplyMethod(reflect.TypeOf(&NpuWorker{}), "GetChipInfo",
		func(f *NpuWorker, cardID int32, deviceID int32) (*ChipInfo, error) {
			return &ChipInfo{
				Name:    "a",
				Type:    "b",
				Version: "1",
			}, nil
		})
	defer patchGetChipInfo.Reset()
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "GetChipName success 1",
			want:    "a",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetChipName()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChipName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetChipName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetProductType tests the function GetProductType
func TestGetProductType(t *testing.T) {
	patch := gomonkey.ApplyFunc(GetCardList, func() (int32, []int32, error) {
		return 1, []int32{0}, nil
	})
	defer patch.Reset()
	patchGetDeviceNumInCard := gomonkey.ApplyFunc(GetDeviceNumInCard, func(cardID int32) (int32, error) {
		return 1, nil
	})
	defer patchGetDeviceNumInCard.Reset()
	tests := []struct {
		name    string
		w       WorkerInterface
		want    string
		wantErr bool
	}{
		{
			name:    "GetProductType success case 1",
			w:       &mockWorker{},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetProductType(tt.w)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProductType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetProductType() got = %v, want %v", got, tt.want)
			}
		})
	}
}
