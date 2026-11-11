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

package rdma

import (
	"ascend-common/common-utils/hwlog"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"ascend-common/common-utils/utils"
)

func TestDiscoverRdmaCollections(t *testing.T) {
	dir := secureRdmaTestDir(t)
	mustWriteFile(t, filepath.Join(dir, rdmaDeviceTableName),
		"cc08 * 0102 fallback\ncc08 8200 0102 driver\ncc08 8201 0102 other\n"+
			"0xCC08 0x8200 0x0102 extra\ncc08 8200 0102 DRIVER\n")
	sysfs := t.TempDir()
	replaceRdmaSysfs(t, sysfs)
	first := fakeRdmaHCA(t, t.TempDir(), "real", "0xCC08", "0x8200", "0X0102")
	requireRdmaTestSuccess(t, os.Symlink(first, filepath.Join(sysfs, "hca0")))
	fakeRdmaHCA(t, sysfs, "hca1", "cc08", "8201", "0102")
	fakeRdmaHCA(t, sysfs, "hca2", "cc08", "8200", "0102")
	broken := fakeRdmaHCA(t, sysfs, "hca3", "cc08", "8200", "0102")
	requireRdmaTestSuccess(t, os.Remove(filepath.Join(broken, "device", "class_code")))
	fakeRdmaHCA(t, sysfs, "hca4", "ffff", "8200", "0102")
	for _, tc := range []struct {
		name           string
		hcaNames, want []string
		errText        string
	}{
		{"stable_dedup", []string{"hca0", "hca1", "hca2"}, []string{"driver", "extra", "other"}, ""},
		{"unselected_bad_cards", []string{"hca0"}, []string{"driver", "extra"}, ""},
		{"selected_broken", []string{"hca0", "hca3"}, nil, "hca3"},
		{"selected_unmatched", []string{"hca0", "hca4"}, nil, "hca4"},
		{"selected_missing", []string{"hca0", "missing"}, nil, "missing"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DiscoverRdmaCollections(dir, tc.hcaNames)
			if tc.errText != "" {
				if err == nil || got != nil || !strings.Contains(err.Error(), tc.errText) {
					t.Fatalf("collections=%v, err=%v, want error containing %q", got, err, tc.errText)
				}
				return
			}
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("collections=%v, err=%v, want %v", got, err, tc.want)
			}
		})
	}
	if got, err := DiscoverRdmaCollections("/missing", nil); err != nil || len(got) != 0 {
		t.Fatal("empty request must not read table")
	}
}

func TestDiscoverRdmaCollectionsFailsWithoutAnyMatch(t *testing.T) {
	dir := secureRdmaTestDir(t)
	mustWriteFile(t, filepath.Join(dir, rdmaDeviceTableName), "cc08 8200 0102 driver\n")
	replaceRdmaSysfs(t, filepath.Join(t.TempDir(), "missing"))
	if _, err := DiscoverRdmaCollections(dir, []string{"unmatched"}); err == nil {
		t.Fatal("missing sysfs must fail")
	}

	sysfs := t.TempDir()
	rdmaSysfsDir = sysfs
	fakeRdmaHCA(t, sysfs, "unmatched", "ffff", "8200", "0102")
	if _, err := DiscoverRdmaCollections(dir, []string{"unmatched"}); err == nil || !strings.Contains(err.Error(), "no matching") {
		t.Fatalf("all unmatched must fail: %v", err)
	}

	mustWriteFile(t, filepath.Join(dir, rdmaDeviceTableName), "bad row\n")
	if _, err := DiscoverRdmaCollections(dir, []string{"unmatched"}); err == nil || !strings.Contains(err.Error(), "line 1") {
		t.Fatalf("invalid explicit-request table must fail: %v", err)
	}
}

func TestParseRdmaDeviceTableAndMatch(t *testing.T) {
	table := "# comment\n 0xCC08\t*\t0X0102 FALLBACK # inline\n" +
		"cc08 8200 0102 Exact\ncc08 8200 0102 Later\n"
	rules, err := parseRdmaDeviceTable(strings.NewReader(table), "table.conf")
	requireRdmaTestSuccess(t, err)
	if got := matchRdmaCollections(rules, rdmaHardwareIdentity{"cc08", "8200", "0102"}); !reflect.DeepEqual(got, []string{"exact", "later"}) {
		t.Fatalf("all collections for the exact tuple must beat the fallback: %v", got)
	}
	if got := matchRdmaCollections(rules, rdmaHardwareIdentity{"cc08", "8201", "0102"}); !reflect.DeepEqual(got, []string{"fallback"}) {
		t.Fatalf("fallback collections = %v", got)
	}
	if got := matchRdmaCollections(rules, rdmaHardwareIdentity{"ffff", "8200", "0102"}); len(got) != 0 {
		t.Fatalf("unexpected collections for unmatched identity: %v", got)
	}
}

func TestParseRdmaDeviceTableRejectsInvalidOrPartialInput(t *testing.T) {
	for _, table := range []string{
		"", "# only comments\n", "cc08 8200 0102", "cc08 8200 0102 driver extra",
		"cc08 8200 0102 ../driver", "cc08 8200 0102 /driver", "cc08 8200 0102 driver.list",
		"cc08 8200 0102 1driver", "cc08 8200 0102 " + strings.Repeat("a", 65),
		"cc08 8200 0102 driver\n" + strings.Repeat("x", 70*1024),
	} {
		assertRdmaTableFailure(t, strings.NewReader(table), "bad.conf")
	}

	for _, name := range []string{"a", strings.Repeat("A", 64), "Driver_1825-v2"} {
		if _, err := normalizeRdmaCollection(name); err != nil {
			t.Errorf("valid collection %q: %v", name, err)
		}
	}
}

func TestParseRdmaDeviceTableLineLimit(t *testing.T) {
	prefix := "cc08 8200 0102 driver #"
	for _, lineSize := range []int{rdmaDeviceTableMaxLineSize - 1, rdmaDeviceTableMaxLineSize} {
		line := prefix + strings.Repeat("x", lineSize-len(prefix))
		for _, ending := range []string{"", "\n", "\r\n"} {
			rules, err := parseRdmaDeviceTable(strings.NewReader(line+ending), "limit.conf")
			if err != nil || len(rules) != 1 || rules[0].collection != "driver" {
				t.Fatalf("line size %d ending %q: rules=%v, err=%v", lineSize, ending, rules, err)
			}
		}
	}
}

func TestParseRdmaDeviceTableRejectsOversizeLinesWithoutPartialRules(t *testing.T) {
	for _, lineSize := range []int{rdmaDeviceTableMaxLineSize + 1, rdmaDeviceTableMaxLineSize * 2} {
		table := "cc08 8200 0102 driver\n#" + strings.Repeat("x", lineSize-1) + "\n"
		assertRdmaTableFailure(t, strings.NewReader(table), "oversize.conf", "oversize.conf", "line 2", "65536 bytes")
	}
}

func TestParseRdmaDeviceTableInterruptedScanHasContext(t *testing.T) {
	for _, path := range []string{"partial.conf", "interrupted.conf"} {
		reader := io.MultiReader(strings.NewReader("cc08 8200 0102 driver\n"), rdmaFailReader{})
		assertRdmaTableFailure(t, reader, path, path, "line 2", "read interrupted")
	}
}

func assertRdmaTableFailure(t *testing.T, reader io.Reader, path string, details ...string) {
	t.Helper()
	rules, err := parseRdmaDeviceTable(reader, path)
	if err == nil || rules != nil {
		t.Fatalf("invalid table returned rules=%v, err=%v", rules, err)
	}
	for _, detail := range details {
		if !strings.Contains(err.Error(), detail) {
			t.Errorf("error %q must contain %q", err, detail)
		}
	}
}

func TestMatchRdmaCollections(t *testing.T) {
	for name, tc := range map[string]struct {
		rules []rdmaDeviceRule
		want  []string
	}{
		"without_rules": {nil, nil},
		"all_wildcard_mappings": {[]rdmaDeviceRule{
			{identity: rdmaHardwareIdentity{"*", "*", "*"}, collection: "base"},
			{identity: rdmaHardwareIdentity{"*", "*", "*"}, collection: "extra"},
		}, []string{"base", "extra"}},
		"exact_mappings_replace_fallback": {[]rdmaDeviceRule{
			{identity: rdmaHardwareIdentity{"cc08", "*", "0102"}, collection: "fallback", rank: 2},
			{identity: rdmaHardwareIdentity{"cc08", "*", "0102"}, collection: "fallback_extra", rank: 2},
			{identity: rdmaHardwareIdentity{"cc08", "8200", "0102"}, collection: "driver", rank: 3},
			{identity: rdmaHardwareIdentity{"cc08", "*", "0102"}, collection: "ignored", rank: 2},
			{identity: rdmaHardwareIdentity{"cc08", "8200", "0102"}, collection: "extra", rank: 3},
		}, []string{"driver", "extra"}},
		"same_rank_retains_first_tuple": {[]rdmaDeviceRule{
			{identity: rdmaHardwareIdentity{"cc08", "*", "0102"}, collection: "driver", rank: 2},
			{identity: rdmaHardwareIdentity{"*", "8200", "0102"}, collection: "other_tuple", rank: 2},
			{identity: rdmaHardwareIdentity{"cc08", "*", "0102"}, collection: "extra", rank: 2},
			{identity: rdmaHardwareIdentity{"ffff", "8200", "0102"}, collection: "unmatched", rank: 3},
		}, []string{"driver", "extra"}},
		"unmatched_identity": {[]rdmaDeviceRule{
			{identity: rdmaHardwareIdentity{"ffff", "8200", "0102"}, collection: "unmatched", rank: 3},
		}, nil},
	} {
		t.Run(name, func(t *testing.T) {
			got := matchRdmaCollections(tc.rules, rdmaHardwareIdentity{"cc08", "8200", "0102"})
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("collections=%v, want %v", got, tc.want)
			}
		})
	}
}

func requireRdmaTestSuccess(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// Configuration checks include parent permissions; because /tmp is writable by other users,
// these integration tests use a private directory under the user's home.
func secureRdmaTestDir(t *testing.T) string {
	t.Helper()
	home, err := os.UserHomeDir()
	requireRdmaTestSuccess(t, err)
	dir, err := os.MkdirTemp(home, "ascend-rdma-test-")
	requireRdmaTestSuccess(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if _, err := utils.RealDirChecker(dir, true, false); err != nil {
		t.Skipf("test requires a home with safe parent permissions: %v", err)
	}
	return dir
}

func fakeRdmaHCA(t *testing.T, root, name, vendorID, deviceID, classCode string) string {
	t.Helper()
	hca := filepath.Join(root, name)
	requireRdmaTestSuccess(t, os.MkdirAll(filepath.Join(hca, "device"), 0755))
	for field, value := range map[string]string{"vendor": vendorID, "device": deviceID, "class_code": classCode} {
		mustWriteFile(t, filepath.Join(hca, "device", field), value+"\n")
	}
	return hca
}

type rdmaFailReader struct{}

func (rdmaFailReader) Read([]byte) (int, error) {
	return 0, errors.New("read interrupted")
}

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestDeviceTableClassWildcardKeepsVendorAndDeviceExact(t *testing.T) {
	rules, err := parseRdmaDeviceTable(strings.NewReader(
		"0xcc08 0x8200 0x0102 rdma_driver\n0xcc08 0x8200 * rdma_driver\n"), "rdma_devices.conf")
	requireRdmaTestSuccess(t, err)
	for _, test := range []struct {
		identity rdmaHardwareIdentity
		matches  bool
	}{
		{rdmaHardwareIdentity{"cc08", "8200", "0102"}, true},
		{rdmaHardwareIdentity{"cc08", "8200", "0200"}, true},
		{rdmaHardwareIdentity{"cc08", "8201", "0102"}, false},
		{rdmaHardwareIdentity{"ffff", "8200", "0102"}, false},
	} {
		got := matchRdmaCollections(rules, test.identity)
		if (len(got) != 0) != test.matches {
			t.Fatalf("identity=%+v collections=%v", test.identity, got)
		}
	}
}

func TestReadRdmaIdentityNormalizesValues(t *testing.T) {
	hcaSysfsPath := fakeRdmaHCA(t, t.TempDir(), "hca0", " 0xCC08 ", "0X8200", " 0x0102 ")
	identity, err := readRdmaIdentity(hcaSysfsPath)
	requireRdmaTestSuccess(t, err)
	want := rdmaHardwareIdentity{vendorID: "cc08", deviceID: "8200", classCode: "0102"}
	if identity != want {
		t.Fatalf("identity=%+v, want %+v", identity, want)
	}
}

func TestReadRdmaIdentityErrorsPreserveEarlierValues(t *testing.T) {
	for _, test := range []struct {
		fieldName string
		want      rdmaHardwareIdentity
	}{
		{"vendor", rdmaHardwareIdentity{}},
		{"device", rdmaHardwareIdentity{vendorID: "cc08"}},
		{"class_code", rdmaHardwareIdentity{vendorID: "cc08", deviceID: "8200"}},
	} {
		for _, state := range []string{"missing", "empty"} {
			t.Run(test.fieldName+"/"+state, func(t *testing.T) {
				hcaSysfsPath := fakeRdmaHCA(t, t.TempDir(), "hca0", "0xCC08", "0X8200", "0x0102")
				fieldPath := filepath.Join(hcaSysfsPath, hcaDeviceDir, test.fieldName)
				if state == "missing" {
					requireRdmaTestSuccess(t, os.Remove(fieldPath))
				} else {
					mustWriteFile(t, fieldPath, " \n")
				}
				identity, err := readRdmaIdentity(hcaSysfsPath)
				if err == nil || !strings.Contains(err.Error(), "read field "+test.fieldName+":") {
					t.Fatalf("expected field error, got %v", err)
				}
				if identity != test.want {
					t.Fatalf("identity=%+v, want %+v", identity, test.want)
				}
				if state == "missing" && !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("missing-file cause was lost: %v", err)
				}
				if state == "empty" && err.Error() != "read field "+test.fieldName+": empty identity" {
					t.Fatalf("unexpected empty-field error: %v", err)
				}
			})
		}
	}
}
