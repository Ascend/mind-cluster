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
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-dynamic-resource-allocation/pkg/consts"
)

// hwLoggingFlagNames lists every flag registered by HWLoggingConfig.RegisterFlags.
var hwLoggingFlagNames = []string{"logFile", "logLevel", "maxAge", "maxBackups"}

// TestHWLoggingConfig_RegisterFlags verifies that HWLoggingConfig.RegisterFlags
// registers all logging flags on flag.CommandLine.
func TestHWLoggingConfig_RegisterFlags(t *testing.T) {
	withFreshFlagSet(t)

	(&HWLoggingConfig{}).RegisterFlags()

	for _, name := range hwLoggingFlagNames {
		if flag.Lookup(name) == nil {
			t.Errorf("flag %q is not registered on flag.CommandLine", name)
		}
	}
}

// TestHWLoggingConfig_RegisterFlagsDefaults verifies that
// HWLoggingConfig.RegisterFlags sets default values for all flags.
func TestHWLoggingConfig_RegisterFlagsDefaults(t *testing.T) {
	withFreshFlagSet(t)

	lg := &HWLoggingConfig{}
	lg.RegisterFlags()

	// Parse with explicit empty arguments so every flag keeps its default
	// value; passing nil would reuse os.Args of the test binary.
	if err := flag.CommandLine.Parse([]string{}); err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	runFlagChecks(t, loggingDefaultChecks(lg))
}

// loggingDefaultChecks builds HWLoggingConfig default-value assertions.
func loggingDefaultChecks(lg *HWLoggingConfig) []flagCheck {
	return []flagCheck{
		{"logFile default", lg.hwLogConfig.LogFileName, api.DefaultLogFile},
		{"logLevel default", lg.hwLogConfig.LogLevel, consts.DefaultLogLevel},
		{"maxAge default", lg.hwLogConfig.MaxAge, consts.DefaultLogMaxAge},
		{"maxBackups default", lg.hwLogConfig.MaxBackups, hwlog.DefaultBackups},
	}
}

// TestHWLoggingConfig_RegisterFlagsParsesOverrides verifies that
// HWLoggingConfig.RegisterFlags parses and applies flag overrides.
func TestHWLoggingConfig_RegisterFlagsParsesOverrides(t *testing.T) {
	withFreshFlagSet(t)

	lg := &HWLoggingConfig{}
	lg.RegisterFlags()

	args := []string{
		"-logFile=/tmp/ut-hwlog.log",
		"-logLevel=3",
		"-maxAge=30",
		"-maxBackups=10",
	}
	if err := flag.CommandLine.Parse(args); err != nil {
		t.Fatalf("Parse(%v) error = %v, want nil", args, err)
	}

	runFlagChecks(t, loggingOverrideChecks(lg))
}

// loggingOverrideChecks builds HWLoggingConfig override assertions.
func loggingOverrideChecks(lg *HWLoggingConfig) []flagCheck {
	return []flagCheck{
		{"logFile", lg.hwLogConfig.LogFileName, "/tmp/ut-hwlog.log"},
		{"logLevel", lg.hwLogConfig.LogLevel, 3},
		{"maxAge", lg.hwLogConfig.MaxAge, 30},
		{"maxBackups", lg.hwLogConfig.MaxBackups, 10},
	}
}

// TestHWLoggingConfig_InitLogModule verifies InitLogModule error handling and
// successful initialization.
func TestHWLoggingConfig_InitLogModule(t *testing.T) {
	// Error cases must run before the success case: hwlog.RunLog is a global
	// singleton, and once initialization succeeds later calls return nil
	// without validating the config anymore.
	runInitLogModuleErrorCases(t)
	runInitLogModuleSuccessCase(t)
}

// runInitLogModuleErrorCases verifies InitLogModule rejects invalid configs.
func runInitLogModuleErrorCases(t *testing.T) {
	t.Helper()
	tests := []struct {
		name        string
		mutate      func(lg *HWLoggingConfig)
		errContains string
	}{
		{
			"log level above maximum",
			func(lg *HWLoggingConfig) { lg.hwLogConfig.LogLevel = 4 },
			"the log level range should be [-1, 3]",
		},
		{
			"max age below minimum",
			func(lg *HWLoggingConfig) { lg.hwLogConfig.MaxAge = 6 },
			"the maxage of backup logs range is [7,700]",
		},
		{
			"zero max backups",
			func(lg *HWLoggingConfig) { lg.hwLogConfig.MaxBackups = 0 },
			"the number of backup log file range is (0, 180]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg := newTestHWLoggingConfig(t)
			tt.mutate(lg)

			err := lg.InitLogModule(context.Background())
			if err == nil {
				t.Fatalf("InitLogModule() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("InitLogModule() error = %v, want containing %q", err, tt.errContains)
			}
		})
	}
}

// runInitLogModuleSuccessCase verifies InitLogModule initializes the run
// logger and creates the log file.
func runInitLogModuleSuccessCase(t *testing.T) {
	t.Helper()
	logFile := filepath.Join(t.TempDir(), "ut-run.log")
	lg := &HWLoggingConfig{
		hwLogConfig: hwlog.LogConfig{
			LogFileName: logFile,
			LogLevel:    consts.DefaultLogLevel,
			MaxAge:      consts.DefaultLogMaxAge,
			MaxBackups:  hwlog.DefaultBackups,
		},
	}

	if err := lg.InitLogModule(context.Background()); err != nil {
		t.Fatalf("InitLogModule() error = %v, want nil", err)
	}
	if _, err := os.Stat(logFile); err != nil {
		t.Errorf("Stat(%q) error = %v, want log file to be created", logFile, err)
	}
}

// newTestHWLoggingConfig builds a valid config pointing at a log file in a
// temp directory; error cases mutate individual fields on top of it.
func newTestHWLoggingConfig(t *testing.T) *HWLoggingConfig {
	t.Helper()
	return &HWLoggingConfig{
		hwLogConfig: hwlog.LogConfig{
			LogFileName: filepath.Join(t.TempDir(), "ut.log"),
			LogLevel:    consts.DefaultLogLevel,
			MaxAge:      consts.DefaultLogMaxAge,
			MaxBackups:  hwlog.DefaultBackups,
		},
	}
}
