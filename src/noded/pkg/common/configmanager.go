/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common for common function
package common

import (
	"sync"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
)

// ConfigManager manage fault config
type ConfigManager interface {
	GetFaultConfig() *FaultConfig
	GetFaultTypeCode() *FaultTypeCode
	SetFaultConfig(*FaultConfig)
	SetFaultTypeCode(*FaultTypeCode)
}

// NewConfigManager create a config manager
func NewConfigManager() ConfigManager {
	return &ConfigTools{
		config:     &FaultConfig{&FaultTypeCode{}},
		configLock: &sync.Mutex{},
	}
}

// ConfigTools the config tool definition
type ConfigTools struct {
	config     *FaultConfig
	configLock *sync.Mutex
}

// GetFaultConfig return fault config
func (c *ConfigTools) GetFaultConfig() *FaultConfig {
	if c.config == nil {
		hwlog.RunLog.Error("config is nil when get fault config")
		return nil
	}
	c.configLock.Lock()
	defer c.configLock.Unlock()
	return c.config
}

// GetFaultTypeCode return fault type code
func (c *ConfigTools) GetFaultTypeCode() *FaultTypeCode {
	if c.config == nil {
		hwlog.RunLog.Error("config is nil when get fault type code")
		return nil
	}
	c.configLock.Lock()
	defer c.configLock.Unlock()
	return c.config.FaultTypeCode
}

// SetFaultConfig set the fault config
func (c *ConfigTools) SetFaultConfig(faultConfig *FaultConfig) {
	if c.config == nil {
		hwlog.RunLog.Error("config is nil when set fault config")
		return
	}
	c.configLock.Lock()
	defer c.configLock.Unlock()
	DeepCopyFaultConfig(c.config, faultConfig)
}

// SetFaultTypeCode set the fault type code
func (c *ConfigTools) SetFaultTypeCode(faultTypeCode *FaultTypeCode) {
	if c.config == nil {
		hwlog.RunLog.Error("config is nil when set fault type code")
		return
	}
	c.configLock.Lock()
	defer c.configLock.Unlock()
	DeepCopyFaultTypeCode(c.config.FaultTypeCode, faultTypeCode)
}
