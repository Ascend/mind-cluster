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
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
)

// resetConfig resets the global config cache and sync.Once for test isolation.
func resetConfig() {
	config = nil
	configOnce = sync.Once{}
}

func init() {
	// Ensure hwlog.RunLog is initialized to avoid nil pointer in tests.
	_ = hwlog.InitRunLogger(&hwlog.LogConfig{
		LogFileName: "/dev/null",
		LogLevel:    0,
		MaxBackups:  0,
		MaxAge:      0,
		OnlyToFile:  true,
		FileMaxSize: 1,
	}, nil)
}

// setInstallInfoPath overrides the install info path resolver for a test.
func setInstallInfoPath(t *testing.T, path string) {
	t.Helper()
	orig := installInfoPath
	installInfoPath = func() (string, error) { return path, nil }
	t.Cleanup(func() { installInfoPath = orig })
}

// writeInstallInfo writes an install.info file with the given content and
// overrides installInfoPath to resolve to it. Returns the written path.
func writeInstallInfo(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), installInfoFileName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write install info file: %v", err)
	}
	setInstallInfoPath(t, path)
	return path
}

func TestLoadConfig_Defaults(t *testing.T) {
	resetConfig()
	// Point the install info path to a non-existent file
	setInstallInfoPath(t, filepath.Join(t.TempDir(), installInfoFileName))

	cfg := loadConfig()

	assert.Equal(t, defaultInjectionMode, cfg.InjectionMode)
}

func TestLoadConfig_CDI(t *testing.T) {
	resetConfig()
	writeInstallInfo(t, "injection-mode=cdi\n")

	cfg := loadConfig()

	assert.Equal(t, cdiInjectionMode, cfg.InjectionMode)
}

func TestLoadConfig_Legacy(t *testing.T) {
	resetConfig()
	writeInstallInfo(t, "injection-mode=legacy\n")

	cfg := loadConfig()

	assert.Equal(t, defaultInjectionMode, cfg.InjectionMode)
}

func TestLoadConfig_UnknownMode(t *testing.T) {
	resetConfig()
	writeInstallInfo(t, "injection-mode=unknown_mode\n")

	cfg := loadConfig()

	assert.Equal(t, defaultInjectionMode, cfg.InjectionMode, "unknown mode should fall back to legacy")
}

func TestLoadConfig_MissingKey(t *testing.T) {
	resetConfig()
	writeInstallInfo(t, "version=v1.0.0\narch=x86_64\n")

	cfg := loadConfig()

	assert.Equal(t, defaultInjectionMode, cfg.InjectionMode, "missing injection-mode key should fall back to legacy")
}

func TestLoadConfig_EmptyValue(t *testing.T) {
	resetConfig()
	writeInstallInfo(t, "injection-mode=\n")

	cfg := loadConfig()

	assert.Equal(t, defaultInjectionMode, cfg.InjectionMode, "empty injection-mode should fall back to legacy")
}

func TestLoadConfig_MultiLine(t *testing.T) {
	resetConfig()
	writeInstallInfo(t, "version=v1.0.0\narch=x86_64\ninjection-mode=cdi\ninstall-scene=docker\n")

	cfg := loadConfig()

	assert.Equal(t, cdiInjectionMode, cfg.InjectionMode, "injection-mode in a middle line should be parsed")
}

func TestLoadConfig_KeyWithSpaces(t *testing.T) {
	resetConfig()
	writeInstallInfo(t, "injection-mode = cdi\n")

	cfg := loadConfig()

	assert.Equal(t, cdiInjectionMode, cfg.InjectionMode, "key with surrounding spaces should be parsed")
}

func TestLoadConfig_Cached(t *testing.T) {
	resetConfig()
	path := writeInstallInfo(t, "injection-mode=cdi\n")

	cfg1 := loadConfig()
	assert.Equal(t, cdiInjectionMode, cfg1.InjectionMode)

	// Delete the install info file — loadConfig should still return cached result
	err := os.Remove(path)
	assert.NoError(t, err)

	// Suppress expected warning via gomonkey since file is now gone,
	// but sync.Once prevents re-reading anyway.
	patches := gomonkey.ApplyFunc(hwlog.RunLog.Warnf, func(format string, args ...interface{}) {})
	defer patches.Reset()

	cfg2 := loadConfig()
	assert.Same(t, cfg1, cfg2, "cached config should return same pointer")
	assert.Equal(t, cdiInjectionMode, cfg2.InjectionMode)
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	resetConfig()
	writeInstallInfo(t, "")

	cfg := loadConfig()

	assert.Equal(t, defaultInjectionMode, cfg.InjectionMode)
}
