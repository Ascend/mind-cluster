/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package device

import (
	"fmt"

	"huawei.com/dpu-exporter/utils/logger"
)

// CardTypeHuawei is the card type identifier for Huawei DPU
const CardTypeHuawei = "huawei"

// DeviceManager is the abstraction layer for DPU hardware operations.
// Different card types implement this interface to provide unified access to DPU metrics.
// Per the class diagram, the interface provides generic ExecCommand and ReadSysfs methods.
type DeviceManager interface {
	// AutoInit discovers DPU devices and populates the device list
	AutoInit() error

	// GetDpuList returns the list of discovered DPU cards
	GetDpuList() []DPU

	// ExecCommand executes a card-type-specific CLI command with given args and returns raw output
	ExecCommand(args ...string) (string, error)

	// ReadSysfs reads a file at the given sysfs path and returns its content
	ReadSysfs(path string) (string, error)

	// ListDir lists file names in the given directory
	ListDir(path string) ([]string, error)

	// GetCardType returns the card type identifier (e.g. "huawei")
	GetCardType() string
}

// AutoInit detects the card type and returns the corresponding DeviceManager implementation.
// Currently only Huawei DPU (hinicadm5) is supported.
// Future card types can be added by extending the switch statement.
func AutoInit(cardType string) (DeviceManager, error) {
	switch cardType {
	case CardTypeHuawei:
		dm, err := NewHwDpuManager()
		if err != nil {
			return nil, fmt.Errorf("init huawei dpu manager failed: %w", err)
		}
		if err := dm.AutoInit(); err != nil {
			return nil, fmt.Errorf("huawei dpu auto init failed: %w", err)
		}
		logger.Infof("huawei dpu manager initialized, dpu count: %d", len(dm.GetDpuList()))
		return dm, nil
	default:
		return nil, fmt.Errorf("unsupported cardType: %s", cardType)
	}
}
