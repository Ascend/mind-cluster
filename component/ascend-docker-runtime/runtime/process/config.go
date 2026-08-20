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
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/mindxcheckutils"
)

// Config holds ascend-docker-runtime configuration loaded from the install
// info file that the installer writes during deployment.
type Config struct {
	InjectionMode string
}

const (
	defaultInjectionMode = "legacy"
	cdiInjectionMode     = "cdi"

	// installInfoFileName is the install record file deployed alongside
	// ascend-docker-runtime under the same install path.
	installInfoFileName = "ascend_docker_runtime_install.info"
	// injectionModeKey is the key in install.info that records the value of
	// the --injection-mode install argument.
	injectionModeKey = "injection-mode"
)

// installInfoPath resolves the path to the install info file that records the
// --injection-mode install argument. It is a function variable to allow test
// overrides. Resolution happens lazily on the first loadConfig() call (guarded
// by configOnce), after the run logger is initialized, so errors can be logged.
var installInfoPath = resolveInstallInfoPath

func resolveInstallInfoPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(execPath), installInfoFileName), nil
}

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

	filePath, pathErr := installInfoPath()
	if filePath == "" {
		hwlog.RunLog.Warnf("failed to resolve install info path: %v, using default injection-mode=%s",
			pathErr, defaultInjectionMode)
		return cfg
	}

	if _, err := mindxcheckutils.RealFileChecker(filePath, false, false, mindxcheckutils.DefaultSize); err != nil {
		hwlog.RunLog.Warnf("install info file %s failed security check: %v, using default injection-mode=%s",
			filePath, err, defaultInjectionMode)
		return cfg
	}

	file, err := os.Open(filePath)
	if err != nil {
		hwlog.RunLog.Warnf("install info file %s not found, using default injection-mode=%s",
			filePath, defaultInjectionMode)
		return cfg
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) != injectionModeKey {
			continue
		}
		mode := strings.TrimSpace(parts[1])
		switch mode {
		case defaultInjectionMode, cdiInjectionMode:
			cfg.InjectionMode = mode
		default:
			hwlog.RunLog.Warnf("unknown injection-mode: %s, falling back to %s",
				mode, defaultInjectionMode)
		}
		break
	}

	return cfg
}
