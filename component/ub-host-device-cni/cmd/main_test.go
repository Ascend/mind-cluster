// Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/containernetworking/cni/pkg/skel"
)

// newFakeUBBus builds a fake /sys/bus/ub/devices tree: addr -> net/<ifname>.
func newFakeUBBus(t *testing.T, devs map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for addr, ifName := range devs {
		netDir := filepath.Join(root, addr, "net")
		if err := os.MkdirAll(netDir, 0o755); err != nil {
			t.Fatalf("failed to create %s: %v", netDir, err)
		}
		// A fake net entry (regular file is enough; only the name is used).
		if err := os.WriteFile(filepath.Join(netDir, ifName), nil, 0o644); err != nil {
			t.Fatalf("failed to create %s: %v", ifName, err)
		}
	}
	return root
}

func TestGetUBDeviceIDs(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })

	// NAD "device" takes priority over runtimeConfig.deviceID.
	sysBusUb = newFakeUBBus(t, map[string]string{"00015": "eth0"})
	cfg := &NetConf{UBMode: true, Device: "eth0", RuntimeConfig: struct {
		DeviceID string `json:"deviceID,omitempty"`
	}{DeviceID: "ignored"}}
	ids, err := getUBDeviceIDs(&skel.CmdArgs{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 1 || ids[0] != "00015" {
		t.Fatalf("expected [00015], got %v", ids)
	}

	// An unresolvable NAD device must fail with a clear error.
	sysBusUb = t.TempDir()
	cfg = &NetConf{UBMode: true, Device: "does-not-exist0"}
	_, err = getUBDeviceIDs(&skel.CmdArgs{}, cfg)
	if err == nil {
		t.Fatal("expected an error for an unresolvable NAD device")
	}
	if !strings.Contains(err.Error(), "cannot resolve UB device from config device") {
		t.Fatalf("unexpected error: %v", err)
	}

	// Without a NAD device, fall back to runtimeConfig.deviceID.
	sysBusUb = newFakeUBBus(t, map[string]string{"00015": "eth0"})
	cfg = &NetConf{UBMode: true, RuntimeConfig: struct {
		DeviceID string `json:"deviceID,omitempty"`
	}{DeviceID: "000abc"}}
	ids, err = getUBDeviceIDs(&skel.CmdArgs{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 1 || ids[0] != "000abc" {
		t.Fatalf("expected [000abc], got %v", ids)
	}

	// No device at all must fail.
	cfg = &NetConf{UBMode: true}
	_, err = getUBDeviceIDs(&skel.CmdArgs{}, cfg)
	if err == nil {
		t.Fatal("expected an error when no device is configured")
	}
	if !strings.Contains(err.Error(), "no allocated DPU devices found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUbDeviceIDFromNAD(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })

	sysBusUb = newFakeUBBus(t, map[string]string{"00015": "eth0"})

	// Configured device resolves to its UB address.
	addr, ok, err := ubDeviceIDFromNAD(&NetConf{Device: "eth0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || addr != "00015" {
		t.Fatalf("expected (00015, true), got (%q, %v)", addr, ok)
	}

	// No device configured -> ok=false, no error.
	addr, ok, err = ubDeviceIDFromNAD(&NetConf{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok || addr != "" {
		t.Fatalf("expected (\"\", false), got (%q, %v)", addr, ok)
	}
}

func TestUbAddressForInterfaceName(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })

	sysBusUb = newFakeUBBus(t, map[string]string{
		"00015": "eth0",
		"00016": "eth1",
	})

	addr, err := ubAddressForInterfaceName("eth1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr != "00016" {
		t.Fatalf("expected 00016, got %q", addr)
	}

	// Unknown interface name -> error.
	_, err = ubAddressForInterfaceName("eth9")
	if err == nil {
		t.Fatal("expected an error for an unknown interface")
	}
	if !strings.Contains(err.Error(), "no UB device owns interface") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetUbInterfaceName(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })

	sysBusUb = newFakeUBBus(t, map[string]string{"00015": "eth0"})

	name, err := getUbInterfaceName("00015")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "eth0" {
		t.Fatalf("expected eth0, got %q", name)
	}

	// Unknown device -> error.
	if _, err := getUbInterfaceName("00099"); err == nil {
		t.Fatal("expected an error for an unknown device")
	}

	// A device whose net dir is empty -> error.
	sysBusUb = t.TempDir()
	if err := os.MkdirAll(filepath.Join(sysBusUb, "00015", "net"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := getUbInterfaceName("00015"); err == nil {
		t.Fatal("expected an error for an empty net dir")
	}
}

func TestHasDpdkDriver(t *testing.T) {
	orig := sysBusPCI
	t.Cleanup(func() { sysBusPCI = orig })

	addr := "0000:00:1f.6"
	pciRoot := t.TempDir()
	sysBusPCI = pciRoot

	// Build the symlink /sys/bus/pci/devices/<addr>/driver -> drivers/<name>.
	mkDriver := func(name string) error {
		target := filepath.Join(pciRoot, "drivers", name)
		if err := os.MkdirAll(target, 0o755); err != nil {
			return err
		}
		link := filepath.Join(pciRoot, addr, "driver")
		if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
			return err
		}
		return os.Symlink(target, link)
	}

	// A userspace driver enables DPDK mode.
	if err := mkDriver("vfio-pci"); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	ok, err := hasDpdkDriver(addr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected DPDK mode for vfio-pci")
	}

	// A kernel driver disables DPDK mode.
	if err := mkDriver("e1000"); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	ok, err = hasDpdkDriver(addr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected no DPDK mode for e1000")
	}

	// Missing driver symlink -> error.
	_, err = hasDpdkDriver("0000:00:00.0")
	if err == nil {
		t.Fatal("expected an error when no driver symlink exists")
	}
}

func TestLoadConf(t *testing.T) {
	// Valid UB config needs no device.
	cfg, err := loadConf([]byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","ubMode":true}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.UBMode {
		t.Fatal("expected ubMode to be true")
	}

	// runtimeConfig.deviceID is preserved for UB mode.
	cfg, err = loadConf([]byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","ubMode":true,"runtimeConfig":{"deviceID":"000abc"}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RuntimeConfig.DeviceID != "000abc" {
		t.Fatalf("expected deviceID 000abc, got %q", cfg.RuntimeConfig.DeviceID)
	}

	// Non-UB config without any device must fail.
	if _, err := loadConf([]byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device"}`)); err == nil {
		t.Fatal("expected an error for a device-less non-UB config")
	}

	// Invalid JSON must fail.
	if _, err := loadConf([]byte(`{`)); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}
