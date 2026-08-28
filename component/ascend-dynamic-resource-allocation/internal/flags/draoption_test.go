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

package flags

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/dynamic-resource-allocation/kubeletplugin"

	"ascend-common/api"
	"ascend-dynamic-resource-allocation/pkg/consts"
)

// draOptionFlagNames lists every flag registered by DRAOption.RegisterFlags.
var draOptionFlagNames = []string{
	"node-name", "cdi-root", "kubelet-registrar-directory-path",
	"kubelet-plugins-directory-path", api.DeviceResetTimeout,
}

// TestDRAOption_RegisterFlags verifies that DRAOption.RegisterFlags registers
// all DRA option flags on flag.CommandLine.
func TestDRAOption_RegisterFlags(t *testing.T) {
	withFreshFlagSet(t)

	(&DRAOption{}).RegisterFlags()

	for _, name := range draOptionFlagNames {
		if flag.Lookup(name) == nil {
			t.Errorf("flag %q is not registered on flag.CommandLine", name)
		}
	}
}

// TestDRAOption_RegisterFlagsDefaults verifies that DRAOption.RegisterFlags
// sets default values for all flags.
func TestDRAOption_RegisterFlagsDefaults(t *testing.T) {
	withFreshFlagSet(t)
	nodeName := "ut-option-node"
	t.Setenv(consts.NodeNameEnv, nodeName)

	opt := &DRAOption{}
	opt.RegisterFlags()

	// Parse with explicit empty arguments so every flag keeps its default
	// value; passing nil would reuse os.Args of the test binary.
	if err := flag.CommandLine.Parse([]string{}); err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	runFlagChecks(t, optionDefaultChecks(opt, nodeName))
}

// TestDRAOption_RegisterFlagsParsesOverrides verifies that DRAOption.RegisterFlags
// parses and applies flag overrides.
func TestDRAOption_RegisterFlagsParsesOverrides(t *testing.T) {
	withFreshFlagSet(t)

	opt := &DRAOption{}
	opt.RegisterFlags()

	args := []string{
		"-node-name=ut-node",
		"-cdi-root=/tmp/ut-cdi",
		"-kubelet-registrar-directory-path=/tmp/ut-registrar",
		"-kubelet-plugins-directory-path=/tmp/ut-plugins",
		"-" + api.DeviceResetTimeout + "=60",
	}
	if err := flag.CommandLine.Parse(args); err != nil {
		t.Fatalf("Parse(%v) error = %v, want nil", args, err)
	}

	runFlagChecks(t, optionOverrideChecks(opt))
}

// TestDRAOption_NodeNameDefaultFromEnv verifies that the node-name flag
// default value follows the NODE_NAME environment variable.
func TestDRAOption_NodeNameDefaultFromEnv(t *testing.T) {
	withFreshFlagSet(t)
	t.Setenv(consts.NodeNameEnv, "node-from-env")

	opt := &DRAOption{}
	opt.RegisterFlags()
	if err := flag.CommandLine.Parse([]string{}); err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if opt.NodeName != "node-from-env" {
		t.Errorf("NodeName = %q, want %q", opt.NodeName, "node-from-env")
	}

	withFreshFlagSet(t)
	t.Setenv(consts.NodeNameEnv, "")

	opt = &DRAOption{}
	opt.RegisterFlags()
	if err := flag.CommandLine.Parse([]string{}); err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if opt.NodeName != "" {
		t.Errorf("NodeName = %q, want empty string when env unset", opt.NodeName)
	}
}

// optionDefaultChecks builds DRAOption default-value assertions.
func optionDefaultChecks(opt *DRAOption, nodeName string) []flagCheck {
	return []flagCheck{
		{"node-name default", opt.NodeName, nodeName},
		{"cdi-root default", opt.CdiRoot, consts.DefaultCDIRoot},
		{
			"kubelet-registrar-directory-path default",
			opt.KubeletRegistrarDirectoryPath, kubeletplugin.KubeletRegistryDir,
		},
		{
			"kubelet-plugins-directory-path default",
			opt.KubeletPluginsDirectoryPath, kubeletplugin.KubeletPluginsDir,
		},
		{"deviceResetTimeout default", opt.DeviceResetTimeout, api.DefaultDeviceResetTimeout},
	}
}

// optionOverrideChecks builds DRAOption override assertions.
func optionOverrideChecks(opt *DRAOption) []flagCheck {
	return []flagCheck{
		{"node-name", opt.NodeName, "ut-node"},
		{"cdi-root", opt.CdiRoot, "/tmp/ut-cdi"},
		{"kubelet-registrar-directory-path", opt.KubeletRegistrarDirectoryPath, "/tmp/ut-registrar"},
		{"kubelet-plugins-directory-path", opt.KubeletPluginsDirectoryPath, "/tmp/ut-plugins"},
		{"deviceResetTimeout", opt.DeviceResetTimeout, 60},
	}
}

// TestDRAOption_Validate verifies directory validation and creation behavior.
func TestDRAOption_Validate(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) *DRAOption
		wantErr     bool
		errContains string
	}{
		{"creates missing directories", setupMissingDirs, false, ""},
		{"passes with existing directories", setupExistingDirs, false, ""},
		{"cdi-root is a file", setupCdiRootFile, true, "cdi-root path validate failed"},
		{
			"registrar path is a file", setupRegistrarFile,
			true, "kubelet-registrar-directory-path validate failed",
		},
		{
			"plugins path is a file", setupPluginsFile,
			true, "kubelet-plugins-directory-path validate failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := tt.setup(t)
			err := opt.Validate()
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				assertDirExists(t, opt.CdiRoot)
				assertDirExists(t, opt.KubeletRegistrarDirectoryPath)
				assertDirExists(t, opt.KubeletPluginsDirectoryPath)
				return
			}
			if err == nil {
				t.Fatalf("Validate() error = nil, want error containing %q", tt.errContains)
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Validate() error = %v, want containing %q", err, tt.errContains)
			}
		})
	}
}

// TestDRAOption_DriverPluginPath verifies the per-driver plugin data path.
func TestDRAOption_DriverPluginPath(t *testing.T) {
	tests := []struct {
		name string
		base string
		want string
	}{
		{"empty base", "", consts.DriverName},
		{
			"normal base", "/var/lib/kubelet/plugins",
			filepath.Join("/var/lib/kubelet/plugins", consts.DriverName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &DRAOption{KubeletPluginsDirectoryPath: tt.base}
			if got := opt.DriverPluginPath(); got != tt.want {
				t.Errorf("DriverPluginPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestEnsureDir verifies directory existence checks and creation.
func TestEnsureDir(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
	}{
		{"creates missing directory", func(t *testing.T) string {
			return filepath.Join(t.TempDir(), "sub")
		}, false},
		{"creates nested directories", func(t *testing.T) string {
			return filepath.Join(t.TempDir(), "a", "b", "c")
		}, false},
		{"existing directory returns nil", func(t *testing.T) string {
			return t.TempDir()
		}, false},
		{"file path returns error", func(t *testing.T) string {
			path := filepath.Join(t.TempDir(), "file.txt")
			writeTestFile(t, path)
			return path
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			if err := ensureDir(path); (err != nil) != tt.wantErr {
				t.Errorf("ensureDir(%q) error = %v, wantErr %v", path, err, tt.wantErr)
			}
		})
	}
}

// newTestDRAOption builds a DRAOption pointing at the given directories.
func newTestDRAOption(cdi, registrar, plugins string) *DRAOption {
	return &DRAOption{
		NodeName:                      "ut-node",
		CdiRoot:                       cdi,
		KubeletRegistrarDirectoryPath: registrar,
		KubeletPluginsDirectoryPath:   plugins,
		DeviceResetTimeout:            api.DefaultDeviceResetTimeout,
	}
}

// missingDirPaths returns three non-existent directory paths under a temp dir.
func missingDirPaths(t *testing.T) (string, string, string) {
	base := t.TempDir()
	return filepath.Join(base, "cdi"),
		filepath.Join(base, "registrar"),
		filepath.Join(base, "plugins")
}

// setupMissingDirs builds an option whose directories do not exist yet.
func setupMissingDirs(t *testing.T) *DRAOption {
	cdi, registrar, plugins := missingDirPaths(t)
	return newTestDRAOption(cdi, registrar, plugins)
}

// setupExistingDirs builds an option whose directories already exist.
func setupExistingDirs(t *testing.T) *DRAOption {
	cdi, registrar, plugins := missingDirPaths(t)
	for _, dir := range []string{cdi, registrar, plugins} {
		if err := os.MkdirAll(dir, consts.DefaultDirMode); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", dir, err)
		}
	}
	return newTestDRAOption(cdi, registrar, plugins)
}

// setupCdiRootFile builds an option whose cdi-root path is a regular file.
func setupCdiRootFile(t *testing.T) *DRAOption {
	cdi, registrar, plugins := missingDirPaths(t)
	writeTestFile(t, cdi)
	return newTestDRAOption(cdi, registrar, plugins)
}

// setupRegistrarFile builds an option whose registrar path is a regular file.
func setupRegistrarFile(t *testing.T) *DRAOption {
	cdi, registrar, plugins := missingDirPaths(t)
	writeTestFile(t, registrar)
	return newTestDRAOption(cdi, registrar, plugins)
}

// setupPluginsFile builds an option whose plugins path is a regular file.
func setupPluginsFile(t *testing.T) *DRAOption {
	cdi, registrar, plugins := missingDirPaths(t)
	writeTestFile(t, plugins)
	return newTestDRAOption(cdi, registrar, plugins)
}

// writeTestFile creates a regular file at path with dummy content.
func writeTestFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("not a dir"), consts.DefaultDirMode); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

// assertDirExists reports an error when path is not an existing directory.
func assertDirExists(t *testing.T, dir string) {
	t.Helper()
	info, err := os.Stat(dir)
	if err != nil {
		t.Errorf("Stat(%q) error = %v, want directory to exist", dir, err)
		return
	}
	if !info.IsDir() {
		t.Errorf("path %q exists but is not a directory", dir)
	}
}
