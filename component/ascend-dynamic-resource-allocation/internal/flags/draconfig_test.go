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
	"testing"

	"k8s.io/dynamic-resource-allocation/kubeletplugin"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-dynamic-resource-allocation/pkg/consts"
)

// defaultHealthzAddress is the default listen port used by the healthz package.
const defaultHealthzAddress = "11251"

// draConfigFlagNames lists every flag registered by NewDraConfig (healthz
// section) and DRAConfig.RegisterFlags (all other sections).
var draConfigFlagNames = []string{
	// HWLoggingConfig
	"logFile", "logLevel", "maxAge", "maxBackups",
	// DRAOption
	"cdi-root", "kubelet-registrar-directory-path",
	"kubelet-plugins-directory-path", api.DeviceResetTimeout,
	// KubeClientConfig
	"kubeconfig", "kube-api-qps", "kube-api-burst",
	// healthz, registered by NewDraConfig
	"enable-healthz", "healthz-address", "tls-cert-file", "tls-private-key-file",
}

// TestNewDraConfig verifies that NewDraConfig returns a non-nil DRAConfig
// instance with default field values.
func TestNewDraConfig(t *testing.T) {
	withFreshFlagSet(t)

	cfg := NewDraConfig()
	if cfg == nil {
		t.Fatal("NewDraConfig() returned nil, want non-nil")
	}

	sections := map[string]interface{}{
		"HwLogConfig":      cfg.HwLogConfig,
		"DraOption":        cfg.DraOption,
		"KubeClientConfig": cfg.KubeClientConfig,
		"DraHealthzConfig": cfg.DraHealthzConfig,
	}
	for name, section := range sections {
		if section == nil {
			t.Errorf("NewDraConfig().%s is nil, want non-nil default instance", name)
		}
	}
}

// TestNewDraConfigRegistersHealthzFlags verifies that NewDraConfig registers
// the healthz flags on flag.CommandLine.
func TestNewDraConfigRegistersHealthzFlags(t *testing.T) {
	withFreshFlagSet(t)

	_ = NewDraConfig()

	healthzFlags := draConfigFlagNames[len(draConfigFlagNames)-4:]
	for _, name := range healthzFlags {
		if flag.Lookup(name) == nil {
			t.Errorf("flag %q is not registered on flag.CommandLine by NewDraConfig", name)
		}
	}
}

// TestDRAConfig_RegisterFlags verifies that DRAConfig.RegisterFlags registers
// all flags on flag.CommandLine.
func TestDRAConfig_RegisterFlags(t *testing.T) {
	withFreshFlagSet(t)

	cfg := NewDraConfig()
	cfg.RegisterFlags()

	for _, name := range draConfigFlagNames {
		if flag.Lookup(name) == nil {
			t.Errorf("flag %q is not registered on flag.CommandLine", name)
		}
	}
}

// TestDRAConfig_RegisterFlagsDefaults verifies that DRAConfig.RegisterFlags
// sets default values for all flags.
func TestDRAConfig_RegisterFlagsDefaults(t *testing.T) {
	withFreshFlagSet(t)
	nodeName := "ut-default-node"
	t.Setenv(consts.NodeNameEnv, nodeName)

	cfg := NewDraConfig()
	cfg.RegisterFlags()

	// Parse with explicit empty arguments so every flag keeps its default
	// value; passing nil would reuse os.Args of the test binary.
	if err := flag.CommandLine.Parse([]string{}); err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	runFlagChecks(t, allDefaultChecks(cfg, nodeName))
}

// TestDRAConfig_RegisterFlagsParsesOverrides verifies that DRAConfig.RegisterFlags
// parses and applies flag overrides.
func TestDRAConfig_RegisterFlagsParsesOverrides(t *testing.T) {
	withFreshFlagSet(t)

	cfg := NewDraConfig()
	cfg.RegisterFlags()

	args := overrideArgs()
	if err := flag.CommandLine.Parse(args); err != nil {
		t.Fatalf("Parse(%v) error = %v, want nil", args, err)
	}

	runFlagChecks(t, allOverrideChecks(cfg))
}

// allDefaultChecks builds the default-value assertions for every section.
func allDefaultChecks(cfg *DRAConfig, nodeName string) []flagCheck {
	checks := hwLoggingDefaultChecks(cfg)
	checks = append(checks, draOptionDefaultChecks(cfg, nodeName)...)
	checks = append(checks, kubeClientDefaultChecks(cfg)...)
	return append(checks, healthzDefaultChecks(cfg)...)
}

// allOverrideChecks builds the override assertions for every section.
func allOverrideChecks(cfg *DRAConfig) []flagCheck {
	checks := hwLoggingOverrideChecks(cfg)
	checks = append(checks, draOptionOverrideChecks(cfg)...)
	checks = append(checks, kubeClientOverrideChecks(cfg)...)
	return append(checks, healthzOverrideChecks(cfg)...)
}

// overrideArgs returns command-line arguments overriding every registered flag.
func overrideArgs() []string {
	return []string{
		"-logFile=/tmp/ut-ascend-dra.log",
		"-logLevel=3",
		"-maxAge=30",
		"-maxBackups=10",
		"-cdi-root=/tmp/ut-cdi",
		"-kubelet-registrar-directory-path=/tmp/ut-registrar",
		"-kubelet-plugins-directory-path=/tmp/ut-plugins",
		"-" + api.DeviceResetTimeout + "=60",
		"-kubeconfig=/tmp/ut-kubeconfig",
		"-kube-api-qps=20",
		"-kube-api-burst=40",
		"-enable-healthz=true",
		"-healthz-address=12345",
		"-tls-cert-file=/tmp/ut-cert.pem",
		"-tls-private-key-file=/tmp/ut-key.pem",
	}
}

// hwLoggingDefaultChecks builds HWLoggingConfig default-value assertions.
func hwLoggingDefaultChecks(cfg *DRAConfig) []flagCheck {
	return []flagCheck{
		{"logFile default", cfg.HwLogConfig.hwLogConfig.LogFileName, api.DefaultLogFile},
		{"logLevel default", cfg.HwLogConfig.hwLogConfig.LogLevel, consts.DefaultLogLevel},
		{"maxAge default", cfg.HwLogConfig.hwLogConfig.MaxAge, consts.DefaultLogMaxAge},
		{"maxBackups default", cfg.HwLogConfig.hwLogConfig.MaxBackups, hwlog.DefaultBackups},
	}
}

// draOptionDefaultChecks builds DRAOption default-value assertions.
func draOptionDefaultChecks(cfg *DRAConfig, nodeName string) []flagCheck {
	return []flagCheck{
		{"node-name default", cfg.DraOption.NodeName, nodeName},
		{"cdi-root default", cfg.DraOption.CdiRoot, consts.DefaultCDIRoot},
		{
			"kubelet-registrar-directory-path default",
			cfg.DraOption.KubeletRegistrarDirectoryPath, kubeletplugin.KubeletRegistryDir,
		},
		{
			"kubelet-plugins-directory-path default",
			cfg.DraOption.KubeletPluginsDirectoryPath, kubeletplugin.KubeletPluginsDir,
		},
		{"deviceResetTimeout default", cfg.DraOption.DeviceResetTimeout, api.DefaultDeviceResetTimeout},
	}
}

// kubeClientDefaultChecks builds KubeClientConfig default-value assertions.
func kubeClientDefaultChecks(cfg *DRAConfig) []flagCheck {
	return []flagCheck{
		{"kubeconfig default", cfg.KubeClientConfig.KubeConfig, ""},
		{"kube-api-qps default", cfg.KubeClientConfig.KubeAPIQPS, float64(consts.DefaultKubeAPIQPS)},
		{"kube-api-burst default", cfg.KubeClientConfig.KubeAPIBurst, consts.DefaultKubeAPIBurst},
	}
}

// healthzDefaultChecks builds healthz section default-value assertions.
func healthzDefaultChecks(cfg *DRAConfig) []flagCheck {
	return []flagCheck{
		{"enable-healthz default", cfg.DraHealthzConfig.EnableHealthz, false},
		{"healthz-address default", cfg.DraHealthzConfig.HealthzAddress, defaultHealthzAddress},
		{"tls-cert-file default", cfg.DraHealthzConfig.TLSCertFile, ""},
		{"tls-private-key-file default", cfg.DraHealthzConfig.TLSPrivateKeyFile, ""},
	}
}

// hwLoggingOverrideChecks builds HWLoggingConfig override assertions.
func hwLoggingOverrideChecks(cfg *DRAConfig) []flagCheck {
	return []flagCheck{
		{"logFile", cfg.HwLogConfig.hwLogConfig.LogFileName, "/tmp/ut-ascend-dra.log"},
		{"logLevel", cfg.HwLogConfig.hwLogConfig.LogLevel, 3},
		{"maxAge", cfg.HwLogConfig.hwLogConfig.MaxAge, 30},
		{"maxBackups", cfg.HwLogConfig.hwLogConfig.MaxBackups, 10},
	}
}

// draOptionOverrideChecks builds DRAOption override assertions.
func draOptionOverrideChecks(cfg *DRAConfig) []flagCheck {
	return []flagCheck{
		{"cdi-root", cfg.DraOption.CdiRoot, "/tmp/ut-cdi"},
		{
			"kubelet-registrar-directory-path",
			cfg.DraOption.KubeletRegistrarDirectoryPath, "/tmp/ut-registrar",
		},
		{
			"kubelet-plugins-directory-path",
			cfg.DraOption.KubeletPluginsDirectoryPath, "/tmp/ut-plugins",
		},
		{"deviceResetTimeout", cfg.DraOption.DeviceResetTimeout, 60},
	}
}

// kubeClientOverrideChecks builds KubeClientConfig override assertions.
func kubeClientOverrideChecks(cfg *DRAConfig) []flagCheck {
	return []flagCheck{
		{"kubeconfig", cfg.KubeClientConfig.KubeConfig, "/tmp/ut-kubeconfig"},
		{"kube-api-qps", cfg.KubeClientConfig.KubeAPIQPS, float64(20)},
		{"kube-api-burst", cfg.KubeClientConfig.KubeAPIBurst, 40},
	}
}

// healthzOverrideChecks builds healthz section override assertions.
func healthzOverrideChecks(cfg *DRAConfig) []flagCheck {
	return []flagCheck{
		{"enable-healthz", cfg.DraHealthzConfig.EnableHealthz, true},
		{"healthz-address", cfg.DraHealthzConfig.HealthzAddress, "12345"},
		{"tls-cert-file", cfg.DraHealthzConfig.TLSCertFile, "/tmp/ut-cert.pem"},
		{"tls-private-key-file", cfg.DraHealthzConfig.TLSPrivateKeyFile, "/tmp/ut-key.pem"},
	}
}
