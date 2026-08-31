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

	"ascend-dynamic-resource-allocation/pkg/consts"
)

// kubeClientFlagNames lists every flag registered by KubeClientConfig.RegisterFlags.
var kubeClientFlagNames = []string{"kubeconfig", "kube-api-qps", "kube-api-burst"}

// TestKubeClientConfig_RegisterFlags verifies that KubeClientConfig.RegisterFlags
// registers all client flags on flag.CommandLine.
func TestKubeClientConfig_RegisterFlags(t *testing.T) {
	withFreshFlagSet(t)

	(&KubeClientConfig{}).RegisterFlags()

	for _, name := range kubeClientFlagNames {
		if flag.Lookup(name) == nil {
			t.Errorf("flag %q is not registered on flag.CommandLine", name)
		}
	}
}

// TestKubeClientConfig_RegisterFlagsDefaults verifies that
// KubeClientConfig.RegisterFlags sets default values for all flags.
func TestKubeClientConfig_RegisterFlagsDefaults(t *testing.T) {
	withFreshFlagSet(t)

	kc := &KubeClientConfig{}
	kc.RegisterFlags()

	// Parse with explicit empty arguments so every flag keeps its default
	// value; passing nil would reuse os.Args of the test binary.
	if err := flag.CommandLine.Parse([]string{}); err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	runFlagChecks(t, clientDefaultChecks(kc))
}

// clientDefaultChecks builds KubeClientConfig default-value assertions.
func clientDefaultChecks(kc *KubeClientConfig) []flagCheck {
	return []flagCheck{
		{"kubeconfig default", kc.KubeConfig, ""},
		{"kube-api-qps default", kc.KubeAPIQPS, float64(consts.DefaultKubeAPIQPS)},
		{"kube-api-burst default", kc.KubeAPIBurst, consts.DefaultKubeAPIBurst},
	}
}

// TestKubeClientConfig_RegisterFlagsParsesOverrides verifies that
// KubeClientConfig.RegisterFlags parses and applies flag overrides.
func TestKubeClientConfig_RegisterFlagsParsesOverrides(t *testing.T) {
	withFreshFlagSet(t)

	kc := &KubeClientConfig{}
	kc.RegisterFlags()

	args := []string{
		"-kubeconfig=/tmp/ut-kubeconfig",
		"-kube-api-qps=20",
		"-kube-api-burst=40",
	}
	if err := flag.CommandLine.Parse(args); err != nil {
		t.Fatalf("Parse(%v) error = %v, want nil", args, err)
	}

	runFlagChecks(t, clientOverrideChecks(kc))
}

// clientOverrideChecks builds KubeClientConfig override assertions.
func clientOverrideChecks(kc *KubeClientConfig) []flagCheck {
	return []flagCheck{
		{"kubeconfig", kc.KubeConfig, "/tmp/ut-kubeconfig"},
		{"kube-api-qps", kc.KubeAPIQPS, float64(20)},
		{"kube-api-burst", kc.KubeAPIBurst, 40},
	}
}

// TestKubeClientConfig_NewClientSetConfig_OutOfCluster verifies that
// NewClientSetConfig builds a rest.Config from a kubeconfig file with the
// configured QPS and burst values.
func TestKubeClientConfig_NewClientSetConfig_OutOfCluster(t *testing.T) {
	kc := &KubeClientConfig{
		KubeConfig:   writeTestKubeconfig(t),
		KubeAPIQPS:   20,
		KubeAPIBurst: 40,
	}

	cfg, err := kc.NewClientSetConfig()
	if err != nil {
		t.Fatalf("NewClientSetConfig() error = %v, want nil", err)
	}
	if cfg.Host != "https://127.0.0.1:6443" {
		t.Errorf("Host = %q, want %q", cfg.Host, "https://127.0.0.1:6443")
	}
	if cfg.QPS != float32(kc.KubeAPIQPS) {
		t.Errorf("QPS = %v, want %v", cfg.QPS, kc.KubeAPIQPS)
	}
	if cfg.Burst != kc.KubeAPIBurst {
		t.Errorf("Burst = %v, want %v", cfg.Burst, kc.KubeAPIBurst)
	}
}

// TestKubeClientConfig_NewClientSetConfig_Errors verifies that
// NewClientSetConfig fails for in-cluster mode without env and for a
// missing kubeconfig file.
func TestKubeClientConfig_NewClientSetConfig_Errors(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) *KubeClientConfig
		errContains string
	}{
		{
			"in-cluster without service env",
			setupInClusterNoEnv,
			"create in-cluster client configuration",
		},
		{
			"missing kubeconfig file",
			func(t *testing.T) *KubeClientConfig {
				return &KubeClientConfig{KubeConfig: filepath.Join(t.TempDir(), "missing.yaml")}
			},
			"create out-of-cluster client configuration",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kc := tt.setup(t)

			cfg, err := kc.NewClientSetConfig()
			if err == nil {
				t.Fatalf("NewClientSetConfig() = %+v, want error containing %q", cfg, tt.errContains)
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("NewClientSetConfig() error = %v, want containing %q", err, tt.errContains)
			}
		})
	}
}

// TestKubeClientConfig_NewClientSets verifies ClientSets creation from a
// kubeconfig file and failure outside a cluster without a kubeconfig.
func TestKubeClientConfig_NewClientSets(t *testing.T) {
	t.Run("success from kubeconfig file", func(t *testing.T) {
		kc := &KubeClientConfig{
			KubeConfig:   writeTestKubeconfig(t),
			KubeAPIQPS:   consts.DefaultKubeAPIQPS,
			KubeAPIBurst: consts.DefaultKubeAPIBurst,
		}

		cs, err := kc.NewClientSets()
		if err != nil {
			t.Fatalf("NewClientSets() error = %v, want nil", err)
		}
		if cs.Core == nil {
			t.Error("NewClientSets().Core is nil, want non-nil core client")
		}
	})

	t.Run("error outside cluster without kubeconfig", func(t *testing.T) {
		kc := setupInClusterNoEnv(t)

		cs, err := kc.NewClientSets()
		if err == nil {
			t.Fatalf("NewClientSets() = %+v, want error", cs)
		}
		if !strings.Contains(err.Error(), "create client configuration") {
			t.Errorf("NewClientSets() error = %v, want containing %q", err, "create client configuration")
		}
	})
}

// setupInClusterNoEnv builds an empty config and clears the in-cluster
// service env vars so rest.InClusterConfig is guaranteed to fail.
func setupInClusterNoEnv(t *testing.T) *KubeClientConfig {
	t.Helper()
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv("KUBERNETES_SERVICE_PORT", "")
	return &KubeClientConfig{}
}

// writeTestKubeconfig writes a minimal valid kubeconfig file and returns
// its path.
func writeTestKubeconfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kubeconfig")
	content := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: ut-cluster
contexts:
- context:
    cluster: ut-cluster
    user: ut-user
  name: ut-context
current-context: ut-context
users:
- name: ut-user
  user:
    token: ut-token
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
	return path
}
