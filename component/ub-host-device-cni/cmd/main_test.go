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
	"encoding/json"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/netlinksafe"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
)

func newFakeUBBus(t *testing.T, devs map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for addr, ifName := range devs {
		netDir := filepath.Join(root, addr, "net")
		if err := os.MkdirAll(netDir, 0o755); err != nil {
			t.Fatalf("failed to create %s: %v", netDir, err)
		}
		if err := os.WriteFile(filepath.Join(netDir, ifName), nil, 0o644); err != nil {
			t.Fatalf("failed to create %s: %v", ifName, err)
		}
	}
	return root
}

// TestGetUBDeviceIDsNADDevice verifies the NAD device priority over runtimeConfig.deviceID.
func TestGetUBDeviceIDsNADDevice(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })
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
}

// TestGetUBDeviceIDsNADUnresolvable verifies the error for an unresolvable NAD device.
func TestGetUBDeviceIDsNADUnresolvable(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })
	sysBusUb = t.TempDir()
	cfg := &NetConf{UBMode: true, Device: "does-not-exist0"}
	_, err := getUBDeviceIDs(&skel.CmdArgs{}, cfg)
	if err == nil {
		t.Fatal("expected an error for an unresolvable NAD device")
	}
	if !strings.Contains(err.Error(), "cannot resolve UB device from config device") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestGetUBDeviceIDsRuntimeConfigFallback verifies the fallback to runtimeConfig.deviceID.
func TestGetUBDeviceIDsRuntimeConfigFallback(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })
	sysBusUb = newFakeUBBus(t, map[string]string{"00015": "eth0"})
	cfg := &NetConf{UBMode: true, RuntimeConfig: struct {
		DeviceID string `json:"deviceID,omitempty"`
	}{DeviceID: "000abc"}}
	ids, err := getUBDeviceIDs(&skel.CmdArgs{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 1 || ids[0] != "000abc" {
		t.Fatalf("expected [000abc], got %v", ids)
	}
}

// TestGetUBDeviceIDsNoDevice verifies the error when no device is configured.
func TestGetUBDeviceIDsNoDevice(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })
	cfg := &NetConf{UBMode: true}
	_, err := getUBDeviceIDs(&skel.CmdArgs{}, cfg)
	if err == nil {
		t.Fatal("expected an error when no device is configured")
	}
	if !strings.Contains(err.Error(), "no allocated DPU devices found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestUbDeviceIDFromNAD verifies device resolution and the empty-config case.
func TestUbDeviceIDFromNAD(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })
	sysBusUb = newFakeUBBus(t, map[string]string{"00015": "eth0"})
	addr, ok, err := ubDeviceIDFromNAD(&NetConf{Device: "eth0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || addr != "00015" {
		t.Fatalf("expected (00015, true), got (%q, %v)", addr, ok)
	}
	addr, ok, err = ubDeviceIDFromNAD(&NetConf{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok || addr != "" {
		t.Fatalf("expected (\"\", false), got (%q, %v)", addr, ok)
	}
}

// TestUbAddressForInterfaceName verifies the interface-to-UB-address lookup.
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
	_, err = ubAddressForInterfaceName("eth9")
	if err == nil {
		t.Fatal("expected an error for an unknown interface")
	}
	if !strings.Contains(err.Error(), "no UB device owns interface") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestGetUbInterfaceName verifies the UB-address-to-interface lookup and error paths.
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
	if _, err := getUbInterfaceName("00099"); err == nil {
		t.Fatal("expected an error for an unknown device")
	}
	sysBusUb = t.TempDir()
	if err := os.MkdirAll(filepath.Join(sysBusUb, "00015", "net"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := getUbInterfaceName("00015"); err == nil {
		t.Fatal("expected an error for an empty net dir")
	}
}

func mkDriverLink(t *testing.T, pciRoot, addr, name string) {
	t.Helper()
	target := filepath.Join(pciRoot, "drivers", name)
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(pciRoot, addr, "driver")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
}

// TestHasDpdkDriverUserspace verifies DPDK mode for a vfio-pci bound device.
func TestHasDpdkDriverUserspace(t *testing.T) {
	orig := sysBusPCI
	t.Cleanup(func() { sysBusPCI = orig })
	pciRoot := t.TempDir()
	sysBusPCI = pciRoot
	mkDriverLink(t, pciRoot, "0000:00:1f.6", "vfio-pci")
	ok, err := hasDpdkDriver("0000:00:1f.6")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected DPDK mode for vfio-pci")
	}
}

// TestHasDpdkDriverKernel verifies no DPDK mode for a kernel driver bound device.
func TestHasDpdkDriverKernel(t *testing.T) {
	orig := sysBusPCI
	t.Cleanup(func() { sysBusPCI = orig })
	pciRoot := t.TempDir()
	sysBusPCI = pciRoot
	mkDriverLink(t, pciRoot, "0000:00:1f.6", "e1000")
	ok, err := hasDpdkDriver("0000:00:1f.6")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected no DPDK mode for e1000")
	}
}

// TestHasDpdkDriverMissing verifies the error for a missing driver symlink.
func TestHasDpdkDriverMissing(t *testing.T) {
	orig := sysBusPCI
	t.Cleanup(func() { sysBusPCI = orig })
	sysBusPCI = t.TempDir()
	_, err := hasDpdkDriver("0000:00:00.0")
	if err == nil {
		t.Fatal("expected an error when no driver symlink exists")
	}
}

// TestLoadConf verifies valid configs and the device-less and invalid-JSON error paths.
func TestLoadConf(t *testing.T) {
	cfg, err := loadConf([]byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","ubMode":true}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.UBMode {
		t.Fatal("expected ubMode to be true")
	}
	cfg, err = loadConf([]byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","ubMode":true,"runtimeConfig":{"deviceID":"000abc"}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RuntimeConfig.DeviceID != "000abc" {
		t.Fatalf("expected deviceID 000abc, got %q", cfg.RuntimeConfig.DeviceID)
	}
	if _, err := loadConf([]byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device"}`)); err == nil {
		t.Fatal("expected an error for a device-less non-UB config")
	}
	if _, err := loadConf([]byte(`{`)); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

// TestHandleDeviceIDEmpty verifies that an empty deviceID is a no-op.
func TestHandleDeviceIDEmpty(t *testing.T) {
	nc := &NetConf{}
	if err := handleDeviceID(nc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestHandleDeviceIDUBMode verifies that UB mode does not map the deviceID.
func TestHandleDeviceIDUBMode(t *testing.T) {
	nc := &NetConf{UBMode: true, RuntimeConfig: struct {
		DeviceID string `json:"deviceID,omitempty"`
	}{DeviceID: "0000:00:1f.6"}}
	if err := handleDeviceID(nc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nc.PCIAddr != "" || nc.auxDevice != "" {
		t.Fatalf("UB mode must not map deviceID, got PCIAddr=%q auxDevice=%q", nc.PCIAddr, nc.auxDevice)
	}
}

// TestHandleDeviceIDPCI verifies that a PCI deviceID is mapped to PCIAddr.
func TestHandleDeviceIDPCI(t *testing.T) {
	origPCI, origAux := sysBusPCI, sysBusAuxiliary
	t.Cleanup(func() { sysBusPCI, sysBusAuxiliary = origPCI, origAux })
	sysBusPCI = filepath.Join(t.TempDir(), "pci")
	sysBusAuxiliary = filepath.Join(t.TempDir(), "aux")
	if err := os.MkdirAll(filepath.Join(sysBusPCI, "0000:00:1f.6"), 0o755); err != nil {
		t.Fatal(err)
	}
	nc := &NetConf{RuntimeConfig: struct {
		DeviceID string `json:"deviceID,omitempty"`
	}{DeviceID: "0000:00:1f.6"}}
	if err := handleDeviceID(nc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nc.PCIAddr != "0000:00:1f.6" || nc.auxDevice != "" {
		t.Fatalf("expected PCIAddr=0000:00:1f.6, got PCIAddr=%q auxDevice=%q", nc.PCIAddr, nc.auxDevice)
	}
}

// TestHandleDeviceIDAuxiliary verifies that an auxiliary deviceID maps to auxDevice.
func TestHandleDeviceIDAuxiliary(t *testing.T) {
	origPCI, origAux := sysBusPCI, sysBusAuxiliary
	t.Cleanup(func() { sysBusPCI, sysBusAuxiliary = origPCI, origAux })
	sysBusPCI = filepath.Join(t.TempDir(), "pci")
	sysBusAuxiliary = filepath.Join(t.TempDir(), "aux")
	if err := os.MkdirAll(filepath.Join(sysBusAuxiliary, "0000:00:1f.6.0"), 0o755); err != nil {
		t.Fatal(err)
	}
	nc := &NetConf{RuntimeConfig: struct {
		DeviceID string `json:"deviceID,omitempty"`
	}{DeviceID: "0000:00:1f.6.0"}}
	if err := handleDeviceID(nc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nc.auxDevice != "0000:00:1f.6.0" || nc.PCIAddr != "" {
		t.Fatalf("expected auxDevice=0000:00:1f.6.0, got PCIAddr=%q auxDevice=%q", nc.PCIAddr, nc.auxDevice)
	}
}

// TestHandleDeviceIDUnknown verifies the error for an unknown deviceID.
func TestHandleDeviceIDUnknown(t *testing.T) {
	origPCI, origAux := sysBusPCI, sysBusAuxiliary
	t.Cleanup(func() { sysBusPCI, sysBusAuxiliary = origPCI, origAux })
	sysBusPCI = filepath.Join(t.TempDir(), "pci")
	sysBusAuxiliary = filepath.Join(t.TempDir(), "aux")
	nc := &NetConf{RuntimeConfig: struct {
		DeviceID string `json:"deviceID,omitempty"`
	}{DeviceID: "0000:99:99.9"}}
	err := handleDeviceID(nc)
	if err == nil {
		t.Fatal("expected an error for an unknown deviceID")
	}
	if !strings.Contains(err.Error(), "not found or unsupported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestParsePrevResult verifies prevResult parsing and the missing-prevResult error.
func TestParsePrevResult(t *testing.T) {
	cfg := &NetConf{}
	if _, err := parsePrevResult(cfg); err == nil {
		t.Fatal("expected an error when prevResult is missing")
	}
	res := &current.Result{
		CNIVersion: "0.3.1",
		Interfaces: []*current.Interface{{Name: "eth0", Sandbox: "netns1"}},
	}
	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("failed to marshal result: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal raw result: %v", err)
	}
	cfg = &NetConf{}
	cfg.CNIVersion = "0.3.1"
	cfg.RawPrevResult = raw
	parsed, err := parsePrevResult(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed.Interfaces) != 1 || parsed.Interfaces[0].Name != "eth0" {
		t.Fatalf("unexpected parsed result: %+v", parsed.Interfaces)
	}
}

// TestFindContainerInterface verifies interface matching and the no-match error.
func TestFindContainerInterface(t *testing.T) {
	result := &current.Result{
		Interfaces: []*current.Interface{
			{Name: "eth0", Sandbox: "netns1"},
			{Name: "net1", Sandbox: "netns1"},
		},
	}
	contMap, err := findContainerInterface(result, &skel.CmdArgs{IfName: "net1", Netns: "netns1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contMap == nil || contMap.Name != "net1" {
		t.Fatalf("expected net1, got %+v", contMap)
	}
	_, err = findContainerInterface(result, &skel.CmdArgs{IfName: "eth9", Netns: "netns1"})
	if err == nil {
		t.Fatal("expected an error when no interface matches")
	}
	if !strings.Contains(err.Error(), "doesn't match configured netns") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestValidateCniContainerInterface verifies empty-name and non-existent-interface rejection.
func TestValidateCniContainerInterface(t *testing.T) {
	intf := current.Interface{Name: "", Sandbox: "netns1"}
	if err := validateCniContainerInterface(intf); err == nil {
		t.Fatal("expected an error for an empty interface name")
	} else if !strings.Contains(err.Error(), "Container interface name missing") {
		t.Fatalf("unexpected error: %v", err)
	}
	intf = current.Interface{Name: "definitely-not-a-real-iface-xyz", Sandbox: "netns1"}
	if err := validateCniContainerInterface(intf); err == nil {
		t.Fatal("expected an error for a non-existent interface")
	}
}

// TestIsUbNetdev verifies that an unresolvable name is not a UB device.
func TestIsUbNetdev(t *testing.T) {
	if isUbNetdev("") {
		t.Fatal("expected false for an unresolvable interface name")
	}
}

// TestLoadConfDeviceID verifies deviceID handling in UB and non-UB modes.
func TestLoadConfDeviceID(t *testing.T) {
	origPCI := sysBusPCI
	t.Cleanup(func() { sysBusPCI = origPCI })
	sysBusPCI = t.TempDir()
	cfg, err := loadConf([]byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","ubMode":true,"runtimeConfig":{"deviceID":"0000:00:1f.6"}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RuntimeConfig.DeviceID != "0000:00:1f.6" || cfg.PCIAddr != "" {
		t.Fatalf("UB mode must keep deviceID, got deviceID=%q PCIAddr=%q", cfg.RuntimeConfig.DeviceID, cfg.PCIAddr)
	}
	_, err = loadConf([]byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","runtimeConfig":{"deviceID":"0000:99:99.9"}}`))
	if err == nil {
		t.Fatal("expected an error for an unresolvable non-UB deviceID")
	}
}

// TestMainVersion verifies that the -version flag prints version output.
func TestMainVersion(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ub-host-device", "-version"}
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	main()
	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}
	if !strings.Contains(string(out), "version=") {
		t.Fatalf("expected version output, got %q", string(out))
	}
}

// TestCmdAddLoadConfError verifies that invalid stdin fails before netns handling.
func TestCmdAddLoadConfError(t *testing.T) {
	if err := cmdAdd(&skel.CmdArgs{StdinData: []byte(`{`)}); err == nil {
		t.Fatal("expected an error for invalid stdin")
	}
}

// TestCmdAddUBNoDevice verifies that UB dispatch with no device fails.
func TestCmdAddUBNoDevice(t *testing.T) {
	args := &skel.CmdArgs{StdinData: []byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","ubMode":true}`)}
	err := cmdAdd(args)
	if err == nil || !strings.Contains(err.Error(), "no allocated DPU devices") {
		t.Fatalf("expected no-device error, got %v", err)
	}
}

// TestCmdAddOpenNetnsError verifies the failure for an unresolvable netns path.
func TestCmdAddOpenNetnsError(t *testing.T) {
	args := &skel.CmdArgs{
		StdinData: []byte(`{"cniVersion":"0.3.1","name":"nd","type":"ub-host-device","device":"eth0"}`),
		Netns:     "/no/such/netns",
	}
	if err := cmdAdd(args); err == nil {
		t.Fatal("expected an error for a missing netns")
	}
}

// TestPrintNoIPAMResultDPDK verifies DPDK-mode result printing without a live link.
func TestPrintNoIPAMResultDPDK(t *testing.T) {
	p := gomonkey.NewPatches()
	defer p.Reset()
	p.ApplyFunc(types.PrintResult, func(types.Result, string) error { return nil })
	cfg := &NetConf{DPDKMode: true}
	cfg.CNIVersion = "0.3.1"
	result := &current.Result{CNIVersion: "0.3.1"}
	result.Interfaces = []*current.Interface{{Name: "eth0"}}
	if err := printNoIPAMResult(cfg, result, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestApplyIPAMResultDPDK verifies that DPDK mode skips interface configuration.
func TestApplyIPAMResultDPDK(t *testing.T) {
	p := gomonkey.NewPatches()
	defer p.Reset()
	p.ApplyFunc(types.PrintResult, func(types.Result, string) error { return nil })
	cfg := &NetConf{DPDKMode: true}
	cfg.CNIVersion = "0.3.1"
	result := &current.Result{CNIVersion: "0.3.1"}
	if err := applyIPAMResult(cfg, nil, "eth0", result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestMoveHostDeviceInNoDevice verifies the error when no host device is configured.
func TestMoveHostDeviceInNoDevice(t *testing.T) {
	if _, err := moveHostDeviceIn(&NetConf{}, nil, "eth0", &current.Result{}); err == nil {
		t.Fatal("expected an error when no host device is configured")
	}
}

// TestRunIPAMExecError verifies the error for a missing IPAM plugin binary.
func TestRunIPAMExecError(t *testing.T) {
	cfg := &NetConf{}
	cfg.IPAM.Type = "fake-ipam"
	if _, err := runIPAM(cfg, &skel.CmdArgs{StdinData: []byte("{}")}, nil); err == nil {
		t.Fatal("expected an error for a missing IPAM plugin")
	}
}

// TestAllocateUBIPAMExecError verifies the UB-mode error for a missing IPAM plugin.
func TestAllocateUBIPAMExecError(t *testing.T) {
	cfg := &NetConf{}
	cfg.IPAM.Type = "fake-ipam"
	if _, err := allocateUBIPAM(cfg, &skel.CmdArgs{StdinData: []byte("{}")}); err == nil {
		t.Fatal("expected an error for a missing IPAM plugin")
	}
}

// TestCmdDelLoadConfError verifies that invalid stdin fails before the netns check.
func TestCmdDelLoadConfError(t *testing.T) {
	if err := cmdDel(&skel.CmdArgs{StdinData: []byte(`{`)}); err == nil {
		t.Fatal("expected an error for invalid stdin")
	}
}

// TestCmdDelEmptyNetns verifies that a missing netns is a no-op for DEL.
func TestCmdDelEmptyNetns(t *testing.T) {
	args := &skel.CmdArgs{
		StdinData: []byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","ubMode":true}`),
		Netns:     "",
	}
	if err := cmdDel(args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCmdDelUBOpenNetnsError verifies the cmdDelUB failure for a missing netns.
func TestCmdDelUBOpenNetnsError(t *testing.T) {
	args := &skel.CmdArgs{Netns: "/no/such/netns"}
	if err := cmdDelUB(args, &NetConf{}); err == nil {
		t.Fatal("expected an error for a missing netns")
	}
}

// TestSetupUBDeviceRequiresIPAM verifies the rejection without IPAM or host-IP inheritance.
func TestSetupUBDeviceRequiresIPAM(t *testing.T) {
	cfg := &NetConf{UBMode: true}
	err := setupUBDevice(cfg, &skel.CmdArgs{}, nil, "00015", &current.Result{})
	if err == nil || !strings.Contains(err.Error(), "ubMode requires either ipam or inheritHostIP") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestSetupUBDeviceHostNotFound verifies the failure for an unresolvable UB device.
func TestSetupUBDeviceHostNotFound(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })
	sysBusUb = t.TempDir()
	cfg := &NetConf{UBMode: true, InheritHostIP: true}
	err := setupUBDevice(cfg, &skel.CmdArgs{}, nil, "00099", &current.Result{})
	if err == nil || !strings.Contains(err.Error(), "failed to find host device") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestGetHostLinkUnknownDevice verifies the error for an unknown UB device.
func TestGetHostLinkUnknownDevice(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })
	sysBusUb = t.TempDir()
	if _, err := getHostLink("00099"); err == nil {
		t.Fatal("expected an error for an unknown UB device")
	}
}

// TestGetLinkDevnameError verifies the error for an unknown device name.
func TestGetLinkDevnameError(t *testing.T) {
	if _, err := getLink("no-such-dev", "", "", "", ""); err == nil {
		t.Fatal("expected an error for an unknown device name")
	}
}

// TestGetLinkMACParseError verifies the error for a malformed MAC.
func TestGetLinkMACParseError(t *testing.T) {
	_, err := getLink("", "not-a-mac", "", "", "")
	if err == nil || !strings.Contains(err.Error(), "failed to parse MAC") {
		t.Fatalf("expected MAC parse error, got %v", err)
	}
}

// TestGetLinkMACNoMatch verifies the not-found error for an unmatched MAC.
func TestGetLinkMACNoMatch(t *testing.T) {
	_, err := getLink("", "aa:bb:cc:dd:ee:ff", "", "", "")
	if err == nil || !strings.Contains(err.Error(), "failed to find physical interface") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

// TestGetLinkKernelPathError verifies rejection of invalid kernel paths.
func TestGetLinkKernelPathError(t *testing.T) {
	if _, err := getLink("", "", "relative/path", "", ""); err == nil {
		t.Fatal("expected an error for a relative kernel path")
	}
	if _, err := getLink("", "", "/sys/devices/no-such-dev", "", ""); err == nil {
		t.Fatal("expected an error for a missing kernel net dir")
	}
}

// TestGetLinkPCINoNetDir verifies the error when a PCI device has no net dir.
func TestGetLinkPCINoNetDir(t *testing.T) {
	orig := sysBusPCI
	t.Cleanup(func() { sysBusPCI = orig })
	sysBusPCI = t.TempDir()
	_, err := getLink("", "", "", "0000:00:1f.6", "")
	if err == nil || !strings.Contains(err.Error(), "no net directory under pci device") {
		t.Fatalf("expected pci net-dir error, got %v", err)
	}
}

// TestGetLinkPCINetDirNoLink verifies the error when the net dir has no real link.
func TestGetLinkPCINetDirNoLink(t *testing.T) {
	orig := sysBusPCI
	t.Cleanup(func() { sysBusPCI = orig })
	pciRoot := t.TempDir()
	sysBusPCI = pciRoot
	netDir := filepath.Join(pciRoot, "0000:00:1f.6", "net")
	if err := os.MkdirAll(netDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(netDir, "fake0"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := getLink("", "", "", "0000:00:1f.6", ""); err == nil {
		t.Fatal("expected an error when the net dir has no real link")
	}
}

// TestGetLinkAuxError verifies the error for a missing auxiliary net dir.
func TestGetLinkAuxError(t *testing.T) {
	orig := sysBusAuxiliary
	t.Cleanup(func() { sysBusAuxiliary = orig })
	sysBusAuxiliary = t.TempDir()
	if _, err := getLink("", "", "", "", "0000:00:1f.6.0"); err == nil {
		t.Fatal("expected an error for a missing auxiliary net dir")
	}
}

// TestGetLinkEmpty verifies the not-found error when no identifier is given.
func TestGetLinkEmpty(t *testing.T) {
	_, err := getLink("", "", "", "", "")
	if err == nil || !strings.Contains(err.Error(), "failed to find physical interface") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

// TestLinkFromPathMissing verifies the error for a missing directory.
func TestLinkFromPathMissing(t *testing.T) {
	if _, err := linkFromPath(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected an error for a missing directory")
	}
}

// TestLinkFromPathNoRealLink verifies the error when no real link exists.
func TestLinkFromPathNoRealLink(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "fake0"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := linkFromPath(dir); err == nil {
		t.Fatal("expected an error when the net dir has no real link")
	}
}

// TestCmdCheckLoadConfError verifies that invalid stdin fails in loadConf.
func TestCmdCheckLoadConfError(t *testing.T) {
	if err := cmdCheck(&skel.CmdArgs{StdinData: []byte(`{`)}); err == nil {
		t.Fatal("expected an error for invalid stdin")
	}
}

// TestCmdCheckOpenNetnsError verifies the failure for an unresolvable netns path.
func TestCmdCheckOpenNetnsError(t *testing.T) {
	args := &skel.CmdArgs{
		StdinData: []byte(`{"cniVersion":"0.3.1","name":"nd","type":"ub-host-device","device":"eth0"}`),
		Netns:     "/no/such/netns",
	}
	if err := cmdCheck(args); err == nil {
		t.Fatal("expected an error for a missing netns")
	}
}

// TestCmdStatusInvalidJSON verifies that invalid JSON fails cmdStatus.
func TestCmdStatusInvalidJSON(t *testing.T) {
	if err := cmdStatus(&skel.CmdArgs{StdinData: []byte(`{`)}); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

// TestCmdStatusIPAMError verifies that a missing IPAM plugin fails cmdStatus.
func TestCmdStatusIPAMError(t *testing.T) {
	args := &skel.CmdArgs{StdinData: []byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","ipam":{"type":"fake-ipam"}}`)}
	if err := cmdStatus(args); err == nil {
		t.Fatal("expected an error for a missing IPAM plugin")
	}
}

// TestCmdStatusNoIPAM verifies that a config without IPAM passes cmdStatus.
func TestCmdStatusNoIPAM(t *testing.T) {
	args := &skel.CmdArgs{StdinData: []byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device"}`)}
	if err := cmdStatus(args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// fakeNetNS implements ns.NetNS without touching the kernel; a nil doFn runs the real closure.
type fakeNetNS struct {
	doFn func(ns.NetNS) error
}

func (f *fakeNetNS) Fd() uintptr  { return 0 }
func (f *fakeNetNS) Path() string { return "/fake/netns" }
func (f *fakeNetNS) Set() error   { return nil }
func (f *fakeNetNS) Close() error { return nil }
func (f *fakeNetNS) Do(fn func(ns.NetNS) error) error {
	if f.doFn != nil {
		return f.doFn(f)
	}
	return fn(f)
}

// TestConfigureIfaceInNSSuccess verifies interface configuration in a fake namespace.
func TestConfigureIfaceInNSSuccess(t *testing.T) {
	fake := &fakeNetNS{doFn: func(ns.NetNS) error { return nil }}
	if err := configureIfaceInNS(fake, "eth0", &current.Result{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestConfigureIfaceInNSError verifies the failure for a missing interface.
func TestConfigureIfaceInNSError(t *testing.T) {
	fake := &fakeNetNS{}
	if err := configureIfaceInNS(fake, "ub-test-no-such", &current.Result{}); err == nil {
		t.Fatal("expected an error for a missing interface")
	}
}

// TestApplyUBIPAMFillsInterfaces verifies that applyUBIPAM fills missing back-references.
func TestApplyUBIPAMFillsInterfaces(t *testing.T) {
	fake := &fakeNetNS{doFn: func(ns.NetNS) error { return nil }}
	iface := &current.Interface{Name: "eth0", Mac: "00:00:00:00:00:01", Sandbox: "/fake/netns"}
	ipamRes := &current.Result{
		IPs: []*current.IPConfig{{Address: net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)}}},
	}
	result := &current.Result{}
	if err := applyUBIPAM(fake, iface, ipamRes, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ipamRes.Interfaces) != 1 || len(result.IPs) != 1 {
		t.Fatalf("expected filled interfaces and ips, got %d/%d", len(ipamRes.Interfaces), len(result.IPs))
	}
	if ipamRes.IPs[0].Interface == nil || *ipamRes.IPs[0].Interface != 0 {
		t.Fatal("expected ipc.Interface to be set to 0")
	}
}

// TestApplyUBIPAMWithInterfaces verifies pass-through of a complete IPAM result.
func TestApplyUBIPAMWithInterfaces(t *testing.T) {
	fake := &fakeNetNS{doFn: func(ns.NetNS) error { return nil }}
	iface := &current.Interface{Name: "eth0"}
	ipamRes := &current.Result{
		Interfaces: []*current.Interface{{Name: "eth0"}},
		IPs: []*current.IPConfig{{
			Interface: current.Int(0),
			Address:   net.IPNet{IP: net.IPv4(10, 0, 0, 2), Mask: net.CIDRMask(24, 32)},
		}},
	}
	result := &current.Result{}
	if err := applyUBIPAM(fake, iface, ipamRes, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.IPs) != 1 {
		t.Fatalf("expected 1 ip, got %d", len(result.IPs))
	}
}

// TestApplyUBIPAMConfigureError verifies the failure when the link cannot be found.
func TestApplyUBIPAMConfigureError(t *testing.T) {
	fake := &fakeNetNS{}
	ipamRes := &current.Result{Interfaces: []*current.Interface{{Name: "eth0"}}}
	err := applyUBIPAM(fake, &current.Interface{Name: "ub-test-no-such"}, ipamRes, &current.Result{})
	if err == nil {
		t.Fatal("expected an error when ConfigureIface cannot find the link")
	}
}

// TestApplyHostAddrsError verifies the failure for a missing interface.
func TestApplyHostAddrsError(t *testing.T) {
	fake := &fakeNetNS{}
	ha := &hostAddrs{addrs: []netlink.Addr{{IPNet: &net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)}}}}
	if err := applyHostAddrs(fake, "ub-test-no-such", ha); err == nil {
		t.Fatal("expected an error for a missing interface")
	}
}

// TestApplyInheritedAddrsSuccess verifies applying inherited host addresses.
func TestApplyInheritedAddrsSuccess(t *testing.T) {
	fake := &fakeNetNS{doFn: func(ns.NetNS) error { return nil }}
	ha := &hostAddrs{addrs: []netlink.Addr{{IPNet: &net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)}}}}
	iface := &current.Interface{Name: "eth0"}
	result := &current.Result{Interfaces: []*current.Interface{{Name: "eth0"}}}
	if err := applyInheritedAddrs(fake, iface, ha, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.IPs) != 1 {
		t.Fatalf("expected 1 ip, got %d", len(result.IPs))
	}
}

// TestApplyInheritedAddrsError verifies the failure for a missing interface.
func TestApplyInheritedAddrsError(t *testing.T) {
	fake := &fakeNetNS{}
	err := applyInheritedAddrs(fake, &current.Interface{Name: "ub-test-no-such"}, &hostAddrs{}, &current.Result{})
	if err == nil {
		t.Fatal("expected an error from applyHostAddrs")
	}
}

// TestUbInterfaceNamesInNetns verifies listing UB-owned interfaces in a namespace.
func TestUbInterfaceNamesInNetns(t *testing.T) {
	fake := &fakeNetNS{}
	if _, err := ubInterfaceNamesInNetns(fake); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPrintLink verifies that a link is rendered into a result and printed.
func TestPrintLink(t *testing.T) {
	p := gomonkey.NewPatches()
	defer p.Reset()
	p.ApplyFunc(types.PrintResult, func(types.Result, string) error { return nil })
	dev := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{
		Name:         "eth0",
		HardwareAddr: []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x01},
	}}
	if err := printLink(dev, "0.3.1", &fakeNetNS{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestValidateContainerStateEmptyName verifies rejection of an empty interface name.
func TestValidateContainerStateEmptyName(t *testing.T) {
	contMap := &current.Interface{Name: ""}
	err := validateContainerState(contMap, &skel.CmdArgs{IfName: "eth0"}, &current.Result{})
	if err == nil {
		t.Fatal("expected an error for an empty interface name")
	}
}

// TestMoveLinkInTempNSError verifies the failure for a missing device.
func TestMoveLinkInTempNSError(t *testing.T) {
	dev := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "ub-test-no-such"}}
	if _, err := moveLinkIn(dev, &fakeNetNS{}, "eth0"); err == nil {
		t.Fatal("expected an error during link move")
	}
}

// TestMoveLinkOutTempNSError verifies the failure for a missing device.
func TestMoveLinkOutTempNSError(t *testing.T) {
	if err := moveLinkOut(&fakeNetNS{}, "ub-test-no-such"); err == nil {
		t.Fatal("expected an error during link move out")
	}
}

// TestListHostAddrs verifies listing host addresses and routes for a link.
func TestListHostAddrs(t *testing.T) {
	dev := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "ub-test-no-such"}}
	ha, err := listHostAddrs(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ha == nil {
		t.Fatal("expected non-nil hostAddrs")
	}
}

// TestRenameAndMoveToContainerError verifies the abort on an injected namespace error.
func TestRenameAndMoveToContainerError(t *testing.T) {
	fake := &fakeNetNS{doFn: func(ns.NetNS) error { return io.EOF }}
	if _, err := renameAndMoveToContainer(fake, fake, "ub-test-no-such", "eth0"); err == nil {
		t.Fatal("expected an error from the injected namespace")
	}
}

// TestRenameAndMoveInTempNSError verifies the failure for a missing device.
func TestRenameAndMoveInTempNSError(t *testing.T) {
	if _, err := renameAndMoveInTempNS(&fakeNetNS{}, &fakeNetNS{}, &fakeNetNS{}, "ub-test-no-such", "eth0"); err == nil {
		t.Fatal("expected an error for a missing device")
	}
}

// TestBringLinkUpError verifies the failure for a missing interface.
func TestBringLinkUpError(t *testing.T) {
	fake := &fakeNetNS{}
	if _, err := bringLinkUp(fake, &fakeNetNS{}, "ub-test-no-such"); err == nil {
		t.Fatal("expected an error for a missing interface")
	}
}

// TestMoveLinkOutToTempNSError verifies the failure for a missing interface.
func TestMoveLinkOutToTempNSError(t *testing.T) {
	fake := &fakeNetNS{}
	if _, err := moveLinkOutToTempNS(fake, &fakeNetNS{}, "ub-test-no-such"); err == nil {
		t.Fatal("expected an error for a missing interface")
	}
}

// TestMoveLinkOutToHostError verifies the failure for a missing interface.
func TestMoveLinkOutToHostError(t *testing.T) {
	fake := &fakeNetNS{}
	if err := moveLinkOutToHost(fake, &fakeNetNS{}, "ub-test-no-such"); err == nil {
		t.Fatal("expected an error for a missing interface")
	}
}

// TestLoadConfDPDK verifies DPDK mode for a vfio-pci bound device.
func TestLoadConfDPDK(t *testing.T) {
	orig := sysBusPCI
	t.Cleanup(func() { sysBusPCI = orig })
	root := t.TempDir()
	sysBusPCI = root
	pciDir := filepath.Join(root, "0000:00:1f.6")
	if err := os.MkdirAll(filepath.Join(root, "vfio-pci"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(pciDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "vfio-pci"), filepath.Join(pciDir, "driver")); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadConf([]byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","pciBusID":"0000:00:1f.6"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.DPDKMode {
		t.Fatal("expected DPDKMode to be true for a vfio-pci device")
	}
}

// TestLoadConfDPDKDriverError verifies the failure for a dangling driver symlink.
func TestLoadConfDPDKDriverError(t *testing.T) {
	orig := sysBusPCI
	t.Cleanup(func() { sysBusPCI = orig })
	root := t.TempDir()
	sysBusPCI = root
	pciDir := filepath.Join(root, "0000:00:1f.6")
	if err := os.MkdirAll(pciDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/definitely/missing/driver", filepath.Join(pciDir, "driver")); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConf([]byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","pciBusID":"0000:00:1f.6"}`)); err == nil {
		t.Fatal("expected an error from the broken driver symlink")
	}
}

// TestApplyIPAMResultConfigureError verifies the failure when the link is missing.
func TestApplyIPAMResultConfigureError(t *testing.T) {
	res := &current.Result{}
	if err := applyIPAMResult(&NetConf{}, &fakeNetNS{}, "ub-test-no-such", res); err == nil {
		t.Fatal("expected an error configuring the interface")
	}
}

// TestCmdAddUBOpenNetnsError verifies the cmdAddUB failure for a missing netns.
func TestCmdAddUBOpenNetnsError(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })
	sysBusUb = newFakeUBBus(t, map[string]string{"00015": "eth0"})
	cfg := &NetConf{UBMode: true, Device: "eth0"}
	args := &skel.CmdArgs{Netns: "/no/such/netns"}
	if err := cmdAddUB(args, cfg); err == nil {
		t.Fatal("expected an error opening the netns")
	}
}

// TestUbAddressForInterfaceNameReadDirError verifies the failure for a missing sysfs root.
func TestUbAddressForInterfaceNameReadDirError(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })
	sysBusUb = filepath.Join(t.TempDir(), "no-such-dir")
	if _, err := ubAddressForInterfaceName("eth0"); err == nil {
		t.Fatal("expected an error when the UB sysfs is missing")
	}
}

// TestUbAddressForInterfaceNameSkipNonDir verifies skipping non-directory entries.
func TestUbAddressForInterfaceNameSkipNonDir(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })
	root := t.TempDir()
	sysBusUb = root
	if err := os.WriteFile(filepath.Join(root, "00099"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	netDir := filepath.Join(root, "00015", "net")
	if err := os.MkdirAll(netDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(netDir, "eth0"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	addr, err := ubAddressForInterfaceName("eth0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr != "00015" {
		t.Fatalf("expected 00015, got %q", addr)
	}
}

// TestGetLinkPCIVirtioGlob verifies virtio net-dir globbing and fake-link failure.
func TestGetLinkPCIVirtioGlob(t *testing.T) {
	orig := sysBusPCI
	t.Cleanup(func() { sysBusPCI = orig })
	root := t.TempDir()
	sysBusPCI = root
	netDir := filepath.Join(root, "0000:00:1f.6", "virtio0", "net")
	if err := os.MkdirAll(netDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(netDir, "fake0"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := getLink("", "", "", "0000:00:1f.6", ""); err == nil {
		t.Fatal("expected an error resolving the fake virtio link")
	}
}

// TestMoveLinkInUpStateDefer verifies the restore-on-error defer for an up link.
func TestMoveLinkInUpStateDefer(t *testing.T) {
	dev := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "ub-test-no-such", Flags: net.FlagUp}}
	if _, err := moveLinkIn(dev, &fakeNetNS{}, "eth0"); err == nil {
		t.Fatal("expected an error during link move")
	}
}

// stubNetlinkFlow patches the netlink/netns plumbing so device moves succeed.
func stubNetlinkFlow(dummy netlink.Link) *gomonkey.Patches {
	p := gomonkey.NewPatches()
	p.ApplyFunc(netlinksafe.LinkByName, func(_ string) (netlink.Link, error) { return dummy, nil })
	p.ApplyFunc(ns.TempNetNS, func() (ns.NetNS, error) { return &fakeNetNS{}, nil })
	p.ApplyFunc(netlink.LinkSetNsFd, func(netlink.Link, int) error { return nil })
	p.ApplyFunc(netlink.LinkSetName, func(netlink.Link, string) error { return nil })
	p.ApplyFunc(netlink.LinkSetAlias, func(netlink.Link, string) error { return nil })
	p.ApplyFunc(netlink.LinkSetUp, func(netlink.Link) error { return nil })
	p.ApplyFunc(types.PrintResult, func(types.Result, string) error { return nil })
	return p
}

// TestCmdAddNonDPDKNoIPAMSuccess verifies the full non-DPDK cmdAdd flow without IPAM.
func TestCmdAddNonDPDKNoIPAMSuccess(t *testing.T) {
	dummy := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "eth0", HardwareAddr: net.HardwareAddr{0x02, 0, 0, 0, 0, 1}}}
	p := stubNetlinkFlow(dummy)
	defer p.Reset()
	p.ApplyFunc(ns.GetNS, func(_ string) (ns.NetNS, error) { return &fakeNetNS{}, nil })
	args := &skel.CmdArgs{IfName: "eth0", Netns: "/fake/netns", StdinData: []byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","device":"eth0"}`)}
	if err := cmdAdd(args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCmdAddWithIPAMSuccess verifies the full non-DPDK cmdAdd flow with IPAM.
func TestCmdAddWithIPAMSuccess(t *testing.T) {
	dummy := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "eth0", HardwareAddr: net.HardwareAddr{0x02, 0, 0, 0, 0, 1}}}
	p := stubNetlinkFlow(dummy)
	defer p.Reset()
	p.ApplyFunc(ns.GetNS, func(_ string) (ns.NetNS, error) { return &fakeNetNS{}, nil })
	p.ApplyFunc(ipam.ExecAdd, func(_ string, _ []byte) (types.Result, error) {
		return &current.Result{CNIVersion: "1.0.0", IPs: []*current.IPConfig{{
			Interface: current.Int(0),
			Address:   net.IPNet{IP: net.IPv4(10, 0, 0, 2), Mask: net.CIDRMask(24, 32)},
		}}}, nil
	})
	p.ApplyFunc(ipam.ConfigureIface, func(_ string, _ *current.Result) error { return nil })
	args := &skel.CmdArgs{IfName: "eth0", Netns: "/fake/netns", StdinData: []byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","device":"eth0","ipam":{"type":"host-local"}}`)}
	if err := cmdAdd(args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCmdAddUBSuccess verifies the full UB-mode cmdAdd flow with IPAM.
func TestCmdAddUBSuccess(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })
	sysBusUb = newFakeUBBus(t, map[string]string{"00015": "eth0"})
	dummy := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "eth0", HardwareAddr: net.HardwareAddr{0x02, 0, 0, 0, 0, 1}}}
	p := stubNetlinkFlow(dummy)
	defer p.Reset()
	p.ApplyFunc(ns.GetNS, func(_ string) (ns.NetNS, error) { return &fakeNetNS{}, nil })
	p.ApplyFunc(ipam.ExecAdd, func(_ string, _ []byte) (types.Result, error) {
		return &current.Result{CNIVersion: "1.0.0", IPs: []*current.IPConfig{{
			Interface: current.Int(0),
			Address:   net.IPNet{IP: net.IPv4(10, 0, 0, 2), Mask: net.CIDRMask(24, 32)},
		}}}, nil
	})
	p.ApplyFunc(ipam.ConfigureIface, func(_ string, _ *current.Result) error { return nil })
	args := &skel.CmdArgs{IfName: "eth0", Netns: "/fake/netns", StdinData: []byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","ubMode":true,"device":"eth0","ipam":{"type":"host-local"}}`)}
	if err := cmdAdd(args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCmdDelNonUBWithIPAMSuccess verifies the full non-UB cmdDel flow with IPAM.
func TestCmdDelNonUBWithIPAMSuccess(t *testing.T) {
	dummy := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "eth0", Alias: "eth0", HardwareAddr: net.HardwareAddr{0x02, 0, 0, 0, 0, 1}}}
	p := stubNetlinkFlow(dummy)
	defer p.Reset()
	p.ApplyFunc(ns.GetNS, func(_ string) (ns.NetNS, error) { return &fakeNetNS{}, nil })
	p.ApplyFunc(ipam.ExecDel, func(_ string, _ []byte) error { return nil })
	args := &skel.CmdArgs{IfName: "eth0", Netns: "/fake/netns", StdinData: []byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","device":"eth0","ipam":{"type":"host-local"}}`)}
	if err := cmdDel(args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCmdDelUBSuccess verifies the full UB-mode cmdDel flow with IPAM.
func TestCmdDelUBSuccess(t *testing.T) {
	orig := sysBusUb
	t.Cleanup(func() { sysBusUb = orig })
	sysBusUb = newFakeUBBus(t, map[string]string{"00015": "eth0"})
	dummy := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "eth0", Alias: "eth0", HardwareAddr: net.HardwareAddr{0x02, 0, 0, 0, 0, 1}}}
	p := stubNetlinkFlow(dummy)
	defer p.Reset()
	p.ApplyFunc(ns.GetNS, func(_ string) (ns.NetNS, error) { return &fakeNetNS{}, nil })
	p.ApplyFunc(ipam.ExecDel, func(_ string, _ []byte) error { return nil })
	p.ApplyFunc(netlink.LinkList, func() ([]netlink.Link, error) { return []netlink.Link{dummy}, nil })
	p.ApplyFunc(isUbNetdev, func(string) bool { return true })
	args := &skel.CmdArgs{IfName: "eth0", Netns: "/fake/netns", StdinData: []byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","ubMode":true,"device":"eth0","ipam":{"type":"host-local"}}`)}
	if err := cmdDel(args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCmdCheckNonDPDKSuccess verifies the full non-DPDK cmdCheck flow.
func TestCmdCheckNonDPDKSuccess(t *testing.T) {
	dummy := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "eth0", HardwareAddr: net.HardwareAddr{0x02, 0, 0, 0, 0, 1}}}
	p := gomonkey.NewPatches()
	defer p.Reset()
	p.ApplyFunc(ns.GetNS, func(_ string) (ns.NetNS, error) { return &fakeNetNS{}, nil })
	p.ApplyFunc(ipam.ExecCheck, func(_ string, _ []byte) error { return nil })
	p.ApplyFunc(netlinksafe.LinkByName, func(_ string) (netlink.Link, error) { return dummy, nil })
	p.ApplyFunc(ip.ValidateExpectedInterfaceIPs, func(_ string, _ []*current.IPConfig) error { return nil })
	p.ApplyFunc(ip.ValidateExpectedRoute, func(_ []*types.Route) error { return nil })
	prev := `"prevResult":{"cniVersion":"0.3.1","interfaces":[{"name":"eth0","mac":"02:00:00:00:00:01","sandbox":"/fake/netns"}],"ips":[{"version":"4","address":"10.0.0.2/24","interface":0}],"routes":[]}`
	args := &skel.CmdArgs{IfName: "eth0", Netns: "/fake/netns", StdinData: []byte(`{"cniVersion":"0.3.1","name":"ub","type":"ub-host-device","device":"eth0","ipam":{"type":"host-local"},` + prev + `}`)}
	if err := cmdCheck(args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
