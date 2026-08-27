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

package main

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"

	"huawei.com/dpu-exporter/pkg/collector/dpucollector"
	"huawei.com/dpu-exporter/pkg/collector/metricscollector"
	"huawei.com/dpu-exporter/pkg/configmanager"
	"huawei.com/dpu-exporter/pkg/device"
	prompkg "huawei.com/dpu-exporter/pkg/platform/prometheus"
	"huawei.com/dpu-exporter/utils/logger"
)

func init() {
	// Initialize logger to prevent nil pointer in logger.Info during tests
	_ = logger.InitLogger()
}

// testConfigJSON contains known interval values to verify config application.
const testConfigJSON = `{"hinicadm5CollectorInterval": 7, "sysfsCollectorInterval": 8,
	"dpuListRefreshInterval": 9, "metricWhiteList": []}`

func loadTestConfig(t *testing.T) {
	t.Helper()
	if err := configmanager.LoadConfigFromBytes([]byte(testConfigJSON)); err != nil {
		t.Fatalf("LoadConfigFromBytes failed: %v", err)
	}
}

// checkChains verifies the expected collector chain layout.
func checkChains(t *testing.T) {
	t.Helper()
	dpu, iface := dpucollector.GetChainsSnapshot()
	if len(dpu) != 1 {
		t.Fatalf("dpu metrics chain len = %d, want 1", len(dpu))
	}
	if _, ok := dpu[0].(*metricscollector.Hinicadm5Collector); !ok {
		t.Errorf("dpu chain[0] = %T, want *metricscollector.Hinicadm5Collector", dpu[0])
	}
	if len(iface) != 1 {
		t.Fatalf("iface metrics chain len = %d, want 1", len(iface))
	}
	if _, ok := iface[0].(*metricscollector.SysfsCollector); !ok {
		t.Errorf("iface chain[0] = %T, want *metricscollector.SysfsCollector", iface[0])
	}
}

// TestInitChains covers chain initialization and interval config application.
func TestInitChains(t *testing.T) {
	defer dpucollector.SetChains(nil, nil)
	loadTestConfig(t)

	initChains()
	checkChains(t)

	// intervals from the config must be applied to the global interval map
	if got := configmanager.GetCollectorInterval(configmanager.CacheKeyHinicadm5, 0); got != 7*time.Second {
		t.Errorf("hinicadm5 interval = %v, want 7s", got)
	}
	if got := configmanager.GetCollectorInterval(configmanager.CacheKeySysfs, 0); got != 8*time.Second {
		t.Errorf("sysfs interval = %v, want 8s", got)
	}
	if got := configmanager.GetDpuListRefreshInterval(0); got != 9*time.Second {
		t.Errorf("dpu list refresh interval = %v, want 9s", got)
	}
}

// TestRebuildChains covers chain rebuilding on config hot-reload.
func TestRebuildChains(t *testing.T) {
	defer dpucollector.SetChains(nil, nil)
	loadTestConfig(t)

	// start from empty chains to prove rebuild replaces them
	dpucollector.SetChains(nil, nil)
	rebuildChains()
	checkChains(t)
}

// fakeDeviceManager implements device.DeviceManager without real hardware.
type fakeDeviceManager struct{}

func (f *fakeDeviceManager) AutoInit() error                            { return nil }
func (f *fakeDeviceManager) GetDpuList() []device.DPU                   { return nil }
func (f *fakeDeviceManager) ExecCommand(args ...string) (string, error) { return "", nil }
func (f *fakeDeviceManager) ReadSysfs(path string) (string, error)      { return "", nil }
func (f *fakeDeviceManager) ListDir(path string) ([]string, error)      { return nil, nil }
func (f *fakeDeviceManager) GetCardType() string                        { return device.CardTypeHuawei }

// TestMainFunction covers the full startup and shutdown flow of main():
// config load, device manager init, chain setup, hot-reload watcher, collect
// goroutines, and the shutdown path when the metrics server fails.
// Hardware-dependent dependencies are patched with gomonkey
// (requires running tests with -gcflags=all=-l, as build/test.sh does).
func TestMainFunction(t *testing.T) {
	defer dpucollector.SetChains(nil, nil)

	autoInitCnt, promCnt, exitCnt := int32(0), int32(0), int32(0)
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(configmanager.LoadConfig, func() error {
		return configmanager.LoadConfigFromBytes([]byte(testConfigJSON))
	})

	patches.ApplyFunc(configmanager.StartHotReload, func(context.Context, *sync.WaitGroup) error {
		return nil
	})
	patches.ApplyFunc(device.AutoInit, func(string) (device.DeviceManager, error) {
		atomic.AddInt32(&autoInitCnt, 1)
		return &fakeDeviceManager{}, nil
	})
	patches.ApplyFunc(prompkg.StartPrometheus, func(string, *prompkg.PrometheusCollector) error {
		atomic.AddInt32(&promCnt, 1)
		return errors.New("prometheus server exit for test")
	})
	// main() calls os.Exit(1) after the metrics server fails; make it a no-op
	// so the test process survives and main() returns normally.
	patches.ApplyFunc(os.Exit, func(int) {
		atomic.AddInt32(&exitCnt, 1)
	})

	main()

	if autoInitCnt != 1 {
		t.Errorf("device.AutoInit called %d times, want 1", autoInitCnt)
	}
	if promCnt != 1 {
		t.Errorf("StartPrometheus called %d times, want 1", promCnt)
	}
	// serverErr path: main exits via the patched os.Exit, then shuts down gracefully
	if exitCnt != 1 {
		t.Errorf("os.Exit called %d times, want 1", exitCnt)
	}
	checkChains(t)
}
