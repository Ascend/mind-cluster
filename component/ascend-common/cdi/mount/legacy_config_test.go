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

// Package mount — legacy_config_test.go
//
// Tests for readListEntries and parseListFile.
package mount

import (
	"path/filepath"
	"testing"
)

// writeListFile writes a .list file under dir with the given content.
func writeListFile(t *testing.T, dir, name, content string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(dir, name+".list"), content)
}

func TestParseListFile_OpenError(t *testing.T) {
	if _, err := parseListFile("/nonexistent/path/foo.list"); err == nil {
		t.Fatal("expected error from opening non-existent file")
	}
}

func TestParseListFile_ScanError(t *testing.T) {
	// One oversized line must trigger a scanner error.
	bigLine := make([]byte, 65537)
	for i := range bigLine {
		bigLine[i] = 'a'
	}
	bigLine[65536] = '\n'
	listFile := filepath.Join(t.TempDir(), "big.list")
	mustWriteFile(t, listFile, string(bigLine))

	if _, err := parseListFile(listFile); err == nil {
		t.Fatal("expected scanner error for oversized line")
	}
}

func TestParseListFile_TooManyEntries(t *testing.T) {
	tmpDir := t.TempDir()

	// More than the max-entries limit must be rejected.
	var content string
	for i := 0; i < 130; i++ {
		content += filepath.Join(tmpDir, "entry"+string(rune('0'+i%10))) + "\n"
	}
	listFile := filepath.Join(tmpDir, "many.list")
	mustWriteFile(t, listFile, content)

	if _, err := parseListFile(listFile); err == nil {
		t.Fatal("expected error for too many entries")
	}
}

func TestReadListEntries_UbIncludedFlag(t *testing.T) {
	tmpDir := t.TempDir()
	ubFile := filepath.Join(tmpDir, "ub.so")
	mustWriteFile(t, ubFile, "")
	writeListFile(t, tmpDir, "ub_driver", ubFile+"\n")

	// disableUBMounts=false + ub_driver.list present → ubIncluded=true.
	entries, ubIncluded, err := readListEntries(tmpDir, "", false)
	if err != nil {
		t.Fatalf("readListEntries returned error: %v", err)
	}
	if !ubIncluded {
		t.Error("ubIncluded = false, want true when ub_driver.list exists")
	}
	if len(entries) != 1 || entries[0].Paths[0] != ubFile {
		t.Fatalf("expected 1 ub entry, got %v", entries)
	}

	// disableUBMounts=true → ub_driver.list ignored; no base.list → 0 entries.
	entries, ubIncluded, err = readListEntries(tmpDir, "", true)
	if err != nil {
		t.Fatalf("readListEntries returned error: %v", err)
	}
	if ubIncluded {
		t.Error("ubIncluded = true, want false when UB mounts disabled")
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestReadListEntries_MissingDir(t *testing.T) {
	entries, _, err := readListEntries("/nonexistent/directory/for/test", "", false)
	if err != nil {
		t.Fatalf("missing dir must not be an error, got: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestReadListEntries_DirIsFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "not-a-dir")
	mustWriteFile(t, file, "")

	entries, _, err := readListEntries(file, "", false)
	if err != nil {
		t.Fatalf("file-as-dir must not be an error, got: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestReadListEntries_BaseList(t *testing.T) {
	dir := t.TempDir()
	lib := filepath.Join(dir, "lib.so")
	mustWriteFile(t, lib, "")
	writeListFile(t, dir, "base", lib+"\n")

	entries, ubIncluded, err := readListEntries(dir, "", false)
	if err != nil {
		t.Fatalf("readListEntries returned error: %v", err)
	}
	if ubIncluded {
		t.Error("ubIncluded = true, want false when ub_driver.list is absent")
	}
	if len(entries) != 1 || entries[0].Paths[0] != lib {
		t.Fatalf("expected 1 base entry (%s), got %v", lib, entries)
	}
}
