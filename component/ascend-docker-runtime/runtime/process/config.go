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

package process

import (
	"encoding/json"
	"os"
	"sync"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/mindxcheckutils"
)

// Config holds ascend-docker-runtime configuration loaded from config file.
type Config struct {
	InjectionMode string `json:"injectionMode"`
}

const (
	defaultInjectionMode = "legacy"
	cdiInjectionMode     = "cdi"
)

// configFilePath is the path to the runtime configuration JSON file.
// Built from the shared config directory constant in ascend-common/api.
// Exported as var to allow test overrides.
var configFilePath = api.RunTimeDConfigPath + "/config.json"

var (
	config     *Config
	configOnce sync.Once
)

// loadConfig reads and caches the runtime configuration.
// Thread-safe via sync.Once; returns the same *Config on repeated calls.
func loadConfig() *Config {
	configOnce.Do(func() {
		config = loadConfigFromFile()
	})
	return config
}

func loadConfigFromFile() *Config {
	cfg := &Config{InjectionMode: defaultInjectionMode}

	if _, err := mindxcheckutils.RealFileChecker(configFilePath, false, false, mindxcheckutils.DefaultSize); err != nil {
		hwlog.RunLog.Warnf("config file %s failed security check: %v, using default injectionMode=%s",
			configFilePath, err, defaultInjectionMode)
		return cfg
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		hwlog.RunLog.Warnf("config file %s not found, using default injectionMode=%s",
			configFilePath, defaultInjectionMode)
		return cfg
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		hwlog.RunLog.Warnf("failed to parse config file %s: %v, using default injectionMode=%s",
			configFilePath, err, defaultInjectionMode)
		return &Config{InjectionMode: defaultInjectionMode}
	}

	switch cfg.InjectionMode {
	case defaultInjectionMode, cdiInjectionMode:
		// valid
	default:
		hwlog.RunLog.Warnf("unknown injectionMode: %s, falling back to %s",
			cfg.InjectionMode, defaultInjectionMode)
		cfg.InjectionMode = defaultInjectionMode
	}

	return cfg
}
