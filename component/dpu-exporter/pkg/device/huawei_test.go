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

package device

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"huawei.com/dpu-exporter/utils/logger"
)

func init() {
	// Initialize logger to prevent nil pointer in logger.Infof/Warnf/Errorf during tests
	_ = logger.InitLogger()
}

// fakeInfoOutput mimics the output of `hinicadm5 info`.
const fakeInfoOutput = `Card num:1
Device Information:
     Card         UB Entity
|----hinic0(CAL_2X400G_UB_EXP)
          |--------0000f(NIC:ens1f0)
          |--------00010(NIC:ens1p1)`

// fakeHinicadm5Script returns a script body printing the given output.
func fakeHinicadm5Script(output string) string {
	return "#!/bin/sh\ncat <<'EOF'\n" + output + "\nEOF\n"
}

// newFakeScriptManager writes an executable script acting as hinicadm5 and
// returns a manager bound to it.
func newFakeScriptManager(t *testing.T, body string) *HwDpuManager {
	t.Helper()
	path := filepath.Join(t.TempDir(), "hinicadm5")
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake hinicadm5 failed: %v", err)
	}
	return NewHwDpuManagerWithPath(path, sysfsNetBase)
}

// TestNewHwDpuManager covers both exit paths of the constructor
// (requires -gcflags=all=-l for gomonkey, as build/test.sh sets globally).
func TestNewHwDpuManager(t *testing.T) {
	// validation failure → error
	patches := gomonkey.NewPatches()
	patches.ApplyFunc(exec.LookPath, func(string) (string, error) {
		return "", errors.New("not found")
	})
	patches.ApplyFunc(validateBinaryPath, func(string) error {
		return errors.New("real file check failed")
	})
	if _, err := NewHwDpuManager(); err == nil || !strings.Contains(err.Error(), "invalid hinicadm5 path") {
		t.Fatalf("NewHwDpuManager() error = %v, want invalid hinicadm5 path error", err)
	}
	patches.Reset()

	// success: path found in PATH and validated
	patches = gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(exec.LookPath, func(string) (string, error) {
		return "/fake/hinicadm5", nil
	})
	patches.ApplyFunc(validateBinaryPath, func(string) error { return nil })
	m, err := NewHwDpuManager()
	if err != nil {
		t.Fatalf("NewHwDpuManager() error = %v, want nil", err)
	}
	if m.hinicadm5Path != "/fake/hinicadm5" || m.sysfsBasePath != sysfsNetBase {
		t.Errorf("manager paths = (%s, %s), want (/fake/hinicadm5, %s)",
			m.hinicadm5Path, m.sysfsBasePath, sysfsNetBase)
	}
}

// TestValidateBinaryPath covers the real validator: a regular file owned by
// the current user passes; missing files and symlinks are rejected.
func TestValidateBinaryPath(t *testing.T) {
	dir := t.TempDir()

	// nonexistent path → error
	if err := validateBinaryPath(filepath.Join(dir, "no-such-file")); err == nil {
		t.Error("validateBinaryPath(nonexistent) = nil, want error")
	}

	// regular file created by the test process → valid (owner uid matches euid)
	bin := filepath.Join(dir, "hinicadm5")
	if err := os.WriteFile(bin, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateBinaryPath(bin); err != nil {
		t.Errorf("validateBinaryPath(regular file) = %v, want nil", err)
	}

	// symlink → rejected (allowLink = false)
	link := filepath.Join(dir, "link")
	if err := os.Symlink(bin, link); err != nil {
		t.Fatal(err)
	}
	if err := validateBinaryPath(link); err == nil {
		t.Error("validateBinaryPath(symlink) = nil, want error")
	}
}

// TestHwDpuManagerExecAndAutoInit covers ExecHinicadm5/ExecCommand and the
// AutoInit outcomes (success / exec failure / no devices) using a fake script.
func TestHwDpuManagerExecAndAutoInit(t *testing.T) {
	m := newFakeScriptManager(t, fakeHinicadm5Script(fakeInfoOutput))

	// ExecHinicadm5 runs the command and returns its output
	out, err := m.ExecHinicadm5(infoSubCmd)
	if err != nil {
		t.Fatalf("ExecHinicadm5 error = %v", err)
	}
	if !strings.Contains(out, "|----hinic0(CAL_2X400G_UB_EXP)") {
		t.Errorf("ExecHinicadm5 output = %q, want hinicadm5 info output", out)
	}

	// ExecCommand delegates to ExecHinicadm5
	if got, err := m.ExecCommand(infoSubCmd); err != nil || got != out {
		t.Errorf("ExecCommand = (%q, %v), want (%q, nil)", got, err, out)
	}

	// AutoInit parses the output into the DPU list
	if err := m.AutoInit(); err != nil {
		t.Fatalf("AutoInit error = %v", err)
	}
	list := m.GetDpuList()
	if len(list) != 1 {
		t.Fatalf("GetDpuList len = %d, want 1", len(list))
	}
	d := list[0]
	if d.CardName != "hinic0" || d.CardType != "CAL_2X400G_UB_EXP" || d.HcaName != "hinic0" {
		t.Errorf("DPU = %+v", d)
	}
	if len(d.Interfaces) != 2 ||
		d.Interfaces[0].EthName != "ens1f0" || d.Interfaces[1].EthName != "ens1p1" {
		t.Errorf("interfaces = %+v, want ens1f0/ens1p1", d.Interfaces)
	}
	if d.Interfaces[0].HcaName != "hinic0" || d.Interfaces[0].SysfsPath != "/sys/class/net/ens1f0" {
		t.Errorf("interface 0 = %+v", d.Interfaces[0])
	}

	// exec failure → error
	fail := newFakeScriptManager(t, "#!/bin/sh\nexit 1\n")
	if _, err := fail.ExecHinicadm5(infoSubCmd); err == nil {
		t.Error("ExecHinicadm5(exit 1) = nil error, want error")
	}
	if err := fail.AutoInit(); err == nil {
		t.Error("AutoInit(exit 1) = nil error, want error")
	}

	// no devices detected → error
	empty := newFakeScriptManager(t, fakeHinicadm5Script("Card num:0\nnothing"))
	if err := empty.AutoInit(); err == nil || !strings.Contains(err.Error(), "no DPU devices detected") {
		t.Errorf("AutoInit(no devices) error = %v, want no DPU devices detected", err)
	}
}

// TestHwDpuManagerSysfs covers ReadSysfs, ReadSysfsMetric, ListDir and GetCardType.
func TestHwDpuManagerSysfs(t *testing.T) {
	base := t.TempDir()
	m := NewHwDpuManagerWithPath("/usr/sbin/hinicadm5", base)

	// ReadSysfs returns trimmed file content
	f := filepath.Join(base, "file1")
	if err := os.WriteFile(f, []byte("  42 \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := m.ReadSysfs(f); err != nil || got != "42" {
		t.Errorf("ReadSysfs = (%q, %v), want (\"42\", nil)", got, err)
	}
	// missing file → error
	if _, err := m.ReadSysfs(filepath.Join(base, "nope")); err == nil {
		t.Error("ReadSysfs(missing) = nil error, want error")
	}

	// ReadSysfsMetric joins <base>/<eth>/statistics/<name>
	statDir := filepath.Join(base, "eth0", "statistics")
	if err := os.MkdirAll(statDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(statDir, "rx_packets"), []byte("100\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := m.ReadSysfsMetric("eth0", "rx_packets"); err != nil || got != "100" {
		t.Errorf("ReadSysfsMetric = (%q, %v), want (\"100\", nil)", got, err)
	}
	if _, err := m.ReadSysfsMetric("eth0", "missing"); err == nil {
		t.Error("ReadSysfsMetric(missing) = nil error, want error")
	}

	// ListDir returns only non-directory entries, sorted
	if err := os.WriteFile(filepath.Join(base, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	names, err := m.ListDir(base)
	if err != nil {
		t.Fatalf("ListDir error = %v", err)
	}
	if len(names) != 2 || names[0] != "a.txt" || names[1] != "file1" {
		t.Errorf("ListDir = %v, want [a.txt file1]", names)
	}
	// missing dir → error
	if _, err := m.ListDir(filepath.Join(base, "nodir")); err == nil {
		t.Error("ListDir(missing) = nil error, want error")
	}

	if m.GetCardType() != CardTypeHuawei {
		t.Errorf("GetCardType = %s, want %s", m.GetCardType(), CardTypeHuawei)
	}
}

// TestParseHinicadm5Info covers the info parser branches.
func TestParseHinicadm5Info(t *testing.T) {
	// full output → 1 DPU with 2 interfaces
	list := parseHinicadm5Info(fakeInfoOutput)
	if len(list) != 1 {
		t.Fatalf("len = %d, want 1", len(list))
	}
	if list[0].CardName != "hinic0" || list[0].CardType != "CAL_2X400G_UB_EXP" {
		t.Errorf("card = (%s, %s)", list[0].CardName, list[0].CardType)
	}
	if len(list[0].Interfaces) != 2 {
		t.Errorf("interfaces = %d, want 2", len(list[0].Interfaces))
	}

	// empty output → no devices
	if list := parseHinicadm5Info(""); len(list) != 0 {
		t.Errorf("empty output: len = %d, want 0", len(list))
	}

	// an interface line before any card is treated as a card line by the
	// parser (prefix |---- matches both) and yields a bogus card
	list = parseHinicadm5Info("|--------0000f(NIC:ens1f0)\n|----hinic0(T)")
	if len(list) != 2 || list[1].CardName != "hinic0" {
		t.Errorf("interface-before-card parsed = %+v, want 2 DPUs with hinic0 last", list)
	}

	// card line without parentheses → skipped
	if list := parseHinicadm5Info("|----hinic0"); len(list) != 0 {
		t.Errorf("no-paren card line: len = %d, want 0", len(list))
	}

	// interface line without NIC: → skipped
	list = parseHinicadm5Info("|----hinic0(T)\n|--------0000f(FOO:ens1f0)")
	if len(list) != 1 || len(list[0].Interfaces) != 0 {
		t.Errorf("no-NIC interface line: got %+v, want 1 DPU without interfaces", list)
	}
}

// TestParseCardLine covers the card line parser.
func TestParseCardLine(t *testing.T) {
	if name, typ := parseCardLine("|----hinic0(CAL_2X400G_UB_EXP)"); name != "hinic0" || typ != "CAL_2X400G_UB_EXP" {
		t.Errorf("parseCardLine = (%q, %q)", name, typ)
	}
	if name, typ := parseCardLine("|----hinic0"); name != "" || typ != "" {
		t.Errorf("parseCardLine(no paren) = (%q, %q), want empty", name, typ)
	}
}

// TestParseInterfaceLine covers the interface line parser.
func TestParseInterfaceLine(t *testing.T) {
	if got := parseInterfaceLine("|--------0000f(NIC:ens1f0)"); got != "ens1f0" {
		t.Errorf("parseInterfaceLine = %q, want ens1f0", got)
	}
	if got := parseInterfaceLine("|--------0000f"); got != "" {
		t.Errorf("parseInterfaceLine(no NIC) = %q, want empty", got)
	}
}
