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

// Package dcmi is used to work with Ascend devices
package dcmi

import (
	"fmt"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

// VDeviceInfo vdevice created info
type VDeviceInfo struct {
	CardID    int32
	DeviceID  int32
	VdeviceID int32
}

// WorkerInterface worker interface
type WorkerInterface interface {
	Initialize() error
	ShutDown()
	CreateVDevice(uniqueID int32, coreNum string) (VDeviceInfo, error)
	DestroyVDevice(uniqueID int32, vDevID int32) error
	GetProductType() (string, error)
	GetChipName() (string, error)
}

// CreateVDevice will create virtual device
func CreateVDevice(w WorkerInterface, spec *specs.Spec, devices []int) (VDeviceInfo, error) {
	invalidVDevice := VDeviceInfo{CardID: -1, DeviceID: -1, VdeviceID: -1}
	if spec == nil {
		return invalidVDevice, fmt.Errorf("spec is nil")
	}
	splitDevice, err := extractVpuParam(spec)
	if err != nil {
		return invalidVDevice, err
	}
	if splitDevice == "" {
		return invalidVDevice, nil
	}
	if len(devices) != 1 || devices[0] < 0 || devices[0] >= hiAIMaxCardNum*hiAIMaxDeviceNum {
		hwlog.RunLog.Errorf("invalid devices: %v", devices)
		return invalidVDevice, fmt.Errorf("invalid devices: %v", devices)
	}

	if err := w.Initialize(); err != nil {
		return invalidVDevice, fmt.Errorf("cannot init dcmi : %v", err)
	}
	defer w.ShutDown()

	vDeviceInfo, err := w.CreateVDevice(int32(devices[0]), splitDevice)
	if err != nil || vDeviceInfo.VdeviceID < 0 {
		hwlog.RunLog.Errorf("cannot create vd or vdevice is wrong: %v %v", vDeviceInfo.VdeviceID, err)
		return invalidVDevice, err
	}
	return vDeviceInfo, nil
}

func extractVpuParam(spec *specs.Spec) (string, error) {
	allowSplit := map[string]string{
		"vir01": "vir01", "vir02": "vir02", "vir04": "vir04", "vir08": "vir08", "vir16": "vir16",
		"vir02_1c": "vir02_1c", "vir03_1c_8g": "vir03_1c_8g", "vir04_3c": "vir04_3c",
		"vir04_4c_dvpp": "vir04_4c_dvpp", "vir04_3c_ndvpp": "vir04_3c_ndvpp",
		"vir05_1c_8g": "vir05_1c_8g", "vir05_1c_16g": "vir05_1c_16g",
		"vir06_1c_16g": "vir06_1c_16g", "vir10_3c_16g": "vir10_3c_16g",
		"vir10_3c_16g_nm": "vir10_3c_16g_nm", "vir10_3c_32g": "vir10_3c_32g",
		"vir10_4c_16g_m": "vir10_4c_16g_m", "vir12_3c_32g": "vir12_3c_32g",
	}

	for _, line := range spec.Process.Env {
		words := strings.Split(line, "=")
		const LENGTH int = 2
		if len(words) != LENGTH {
			continue
		}
		if strings.TrimSpace(words[0]) == api.AscendVnpuSpescEnv {
			if split, ok := allowSplit[words[1]]; ok && split != "" {
				return split, nil
			}
			return "", fmt.Errorf("cannot parse param : %v", words[1])
		}
	}
	return "", nil
}
