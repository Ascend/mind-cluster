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

// Package mount — file_test.go
//
// Tests for FileProvider.GetMounts and readListFile.
package mount

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cdispec "tags.cncf.io/container-device-interface/specs-go"

	"ascend-common/common-utils/hwlog"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

// =============================================================================
// FileProvider.GetMounts — happy paths
// =============================================================================

func TestFileProvider_ValidFiles(t *testing.T) {
	tmpDir := t.TempDir()

	paths := []string{filepath.Join(tmpDir, "f1"), filepath.Join(tmpDir, "f2"), filepath.Join(tmpDir, "f3")}
	for _, p := range paths {
		if err := os.WriteFile(p, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}

	listContent := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n",
		paths[0], paths[1], filepath.Join(tmpDir, "no_such"), paths[2], filepath.Join(tmpDir, "missing"))
	if err := os.WriteFile(filepath.Join(tmpDir, "m.list"), []byte(listContent), 0644); err != nil {
		t.Fatal(err)
	}

	mounts, err := (&FileProvider{Dir: tmpDir, MountNames: "m"}).GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 3 {
		t.Fatalf("expected 3 mounts, got %d", len(mounts))
	}

	expect := map[string]bool{paths[0]: true, paths[1]: true, paths[2]: true}
	for _, m := range mounts {
		assertBindMount(t, m)
		if !expect[m.HostPath] {
			t.Errorf("unexpected mount path: %s", m.HostPath)
		}
	}
}

func assertBindMount(t *testing.T, m *cdispec.Mount) {
	t.Helper()
	if m.Type != "bind" {
		t.Errorf("mount %s: Type = %q, want bind", m.HostPath, m.Type)
	}
	if m.HostPath != m.ContainerPath {
		t.Errorf("mount %s: ContainerPath = %q, want %q", m.HostPath, m.ContainerPath, m.HostPath)
	}
	if len(m.Options) != 3 || m.Options[0] != "rbind" || m.Options[1] != "rprivate" || m.Options[2] != "ro" {
		t.Errorf("mount %s: Options = %v, want [rbind rprivate ro]", m.HostPath, m.Options)
	}
}

func TestFileProvider_EmptyDir(t *testing.T) {
	provider := &FileProvider{Dir: t.TempDir()}
	mounts, err := provider.GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 0 {
		t.Fatalf("expected 0 mounts for empty dir, got %d", len(mounts))
	}
}

func TestFileProvider_NonExistentDir(t *testing.T) {
	provider := &FileProvider{Dir: "/nonexistent/directory/for/test"}
	mounts, err := provider.GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error for non-existent dir: %v", err)
	}
	if len(mounts) != 0 {
		t.Fatalf("expected 0 mounts for non-existent dir, got %d", len(mounts))
	}
}

func TestFileProvider_SkipComments(t *testing.T) {
	tmpDir := t.TempDir()
	validPath := filepath.Join(tmpDir, "valid_file")
	if err := os.WriteFile(validPath, nil, 0644); err != nil {
		t.Fatal(err)
	}

	listContent := "# This is a comment\n\n" + validPath + "\n# Another comment\n\n   # indented comment\n"
	listFile := filepath.Join(tmpDir, "mounts.list")
	if err := os.WriteFile(listFile, []byte(listContent), 0644); err != nil {
		t.Fatal(err)
	}

	provider := &FileProvider{Dir: tmpDir, MountNames: "mounts"}
	mounts, err := provider.GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 1 {
		t.Fatalf("expected 1 mount (only valid path), got %d", len(mounts))
	}
	if mounts[0].HostPath != validPath {
		t.Errorf("expected mount path %q, got %q", validPath, mounts[0].HostPath)
	}
}

func TestFileProvider_NoListFiles(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), nil, 0644); err != nil {
		t.Fatal(err)
	}

	provider := &FileProvider{Dir: tmpDir}
	mounts, err := provider.GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 0 {
		t.Fatalf("expected 0 mounts when no .list files present, got %d", len(mounts))
	}
}

func TestFileProvider_MultipleListFiles(t *testing.T) {
	tmpDir := t.TempDir()

	valid1 := filepath.Join(tmpDir, "lib1.so")
	valid2 := filepath.Join(tmpDir, "lib2.so")
	for _, p := range []string{valid1, valid2} {
		if err := os.WriteFile(p, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}

	list1 := filepath.Join(tmpDir, "base.list")
	list2 := filepath.Join(tmpDir, "extra.list")
	if err := os.WriteFile(list1, []byte(valid1+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(list2, []byte(valid2+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	mounts, err := (&FileProvider{Dir: tmpDir, MountNames: "base,extra"}).GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 2 {
		t.Fatalf("expected 2 mounts from 2 list files, got %d", len(mounts))
	}
}

// =============================================================================
// FileProvider.GetMounts — MountNames behavior
// =============================================================================

func TestFileProvider_MountNamesEmptyDefaultsToBase(t *testing.T) {
	tmpDir := t.TempDir()

	validPath := filepath.Join(tmpDir, "lib.so")
	if err := os.WriteFile(validPath, nil, 0644); err != nil {
		t.Fatal(err)
	}

	// base.list should be read by default
	if err := os.WriteFile(filepath.Join(tmpDir, "base.list"), []byte(validPath+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// extra.list should be ignored when MountNames is empty
	if err := os.WriteFile(filepath.Join(tmpDir, "extra.list"), []byte("/should/not/read\n"), 0644); err != nil {
		t.Fatal(err)
	}

	mounts, err := (&FileProvider{Dir: tmpDir}).GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 1 {
		t.Fatalf("expected 1 mount (only base.list), got %d", len(mounts))
	}
	if mounts[0].HostPath != validPath {
		t.Errorf("expected mount path %q, got %q", validPath, mounts[0].HostPath)
	}
}

func TestFileProvider_MountNamesSingleCustom(t *testing.T) {
	tmpDir := t.TempDir()

	validPath := filepath.Join(tmpDir, "custom_lib.so")
	if err := os.WriteFile(validPath, nil, 0644); err != nil {
		t.Fatal(err)
	}

	// custom.list should be read
	if err := os.WriteFile(filepath.Join(tmpDir, "custom.list"), []byte(validPath+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// base.list should be ignored
	if err := os.WriteFile(filepath.Join(tmpDir, "base.list"), []byte("/should/not/read\n"), 0644); err != nil {
		t.Fatal(err)
	}

	mounts, err := (&FileProvider{Dir: tmpDir, MountNames: "custom"}).GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 1 {
		t.Fatalf("expected 1 mount (only custom.list), got %d", len(mounts))
	}
	if mounts[0].HostPath != validPath {
		t.Errorf("expected mount path %q, got %q", validPath, mounts[0].HostPath)
	}
}

func TestFileProvider_MountNamesMultipleCommaSeparated(t *testing.T) {
	tmpDir := t.TempDir()

	aPath := filepath.Join(tmpDir, "a_lib.so")
	bPath := filepath.Join(tmpDir, "b_lib.so")
	for _, p := range []string{aPath, bPath} {
		if err := os.WriteFile(p, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "a.list"), []byte(aPath+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.list"), []byte(bPath+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// c.list should be ignored
	if err := os.WriteFile(filepath.Join(tmpDir, "c.list"), []byte("/should/not/read\n"), 0644); err != nil {
		t.Fatal(err)
	}

	mounts, err := (&FileProvider{Dir: tmpDir, MountNames: "a,b"}).GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 2 {
		t.Fatalf("expected 2 mounts from a.list and b.list, got %d", len(mounts))
	}
}

func TestFileProvider_MountNamesSpacesInCommaList(t *testing.T) {
	tmpDir := t.TempDir()

	validPath := filepath.Join(tmpDir, "spaced.so")
	if err := os.WriteFile(validPath, nil, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "first.list"), []byte(validPath+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "second.list"), []byte(validPath+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	mounts, err := (&FileProvider{Dir: tmpDir, MountNames: " first , second "}).GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 2 {
		t.Fatalf("expected 2 mounts with spaced MountNames, got %d", len(mounts))
	}
}

func TestFileProvider_MountNamesMissingFileSkipped(t *testing.T) {
	tmpDir := t.TempDir()

	validPath := filepath.Join(tmpDir, "exists.so")
	if err := os.WriteFile(validPath, nil, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "exists.list"), []byte(validPath+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// "missing" has no corresponding .list file
	mounts, err := (&FileProvider{Dir: tmpDir, MountNames: "exists,missing"}).GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 1 {
		t.Fatalf("expected 1 mount (missing skipped), got %d", len(mounts))
	}
}

// =============================================================================
// parseMountNames unit tests
// =============================================================================

func TestParseMountNames_Empty(t *testing.T) {
	names := parseMountNames("")
	if len(names) != 1 || names[0] != "base" {
		t.Fatalf("expected [base] for empty input, got %v", names)
	}
}

func TestParseMountNames_Single(t *testing.T) {
	names := parseMountNames("custom")
	if len(names) != 1 || names[0] != "custom" {
		t.Fatalf("expected [custom], got %v", names)
	}
}

func TestParseMountNames_Multiple(t *testing.T) {
	names := parseMountNames("a,b,c")
	if len(names) != 3 || names[0] != "a" || names[1] != "b" || names[2] != "c" {
		t.Fatalf("expected [a b c], got %v", names)
	}
}

func TestParseMountNames_WithSpaces(t *testing.T) {
	names := parseMountNames(" x , y ")
	if len(names) != 2 || names[0] != "x" || names[1] != "y" {
		t.Fatalf("expected [x y] with trimmed spaces, got %v", names)
	}
}

func TestParseMountNames_EmptyEntries(t *testing.T) {
	names := parseMountNames("a,,b")
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Fatalf("expected [a b] with empty entry skipped, got %v", names)
	}
}

// =============================================================================
// FileProvider.GetMounts — error paths
// =============================================================================

func TestFileProvider_FileNotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "not-a-dir")
	if err := os.WriteFile(tmpFile, nil, 0644); err != nil {
		t.Fatal(err)
	}

	provider := &FileProvider{Dir: tmpFile}
	mounts, err := provider.GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 0 {
		t.Fatalf("expected 0 mounts for file-not-directory, got %d", len(mounts))
	}
}

// =============================================================================
// readListFile — error paths
// =============================================================================

func TestReadListFile_OpenError(t *testing.T) {
	provider := &FileProvider{Dir: t.TempDir()}
	_, err := provider.readListFile("/nonexistent/path/foo.list")
	if err == nil {
		t.Fatal("expected error from opening non-existent file")
	}
}

func TestReadListFile_ScanError(t *testing.T) {
	tmpDir := t.TempDir()
	listFile := filepath.Join(tmpDir, "big.list")

	bigLine := make([]byte, 65537)
	for i := range bigLine {
		bigLine[i] = 'a'
	}
	bigLine[65536] = '\n'
	if err := os.WriteFile(listFile, bigLine, 0644); err != nil {
		t.Fatal(err)
	}

	provider := &FileProvider{Dir: tmpDir}
	_, err := provider.readListFile(listFile)
	if err == nil {
		t.Fatal("expected scanner error for oversized line")
	}
}

func TestReadListFile_TooManyEntries(t *testing.T) {
	tmpDir := t.TempDir()
	listFile := filepath.Join(tmpDir, "many.list")

	for i := 0; i < 130; i++ {
		entry := filepath.Join(tmpDir, "entry"+string(rune('0'+i%10)))
		os.WriteFile(entry, nil, 0644)
	}

	var content string
	for i := 0; i < 130; i++ {
		content += filepath.Join(tmpDir, "entry"+string(rune('0'+i%10))) + "\n"
	}
	if err := os.WriteFile(listFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	provider := &FileProvider{Dir: tmpDir}
	_, err := provider.readListFile(listFile)
	if err == nil {
		t.Fatal("expected error for too many entries")
	}
}

// =============================================================================
// FileProvider.GetMounts — HostRoot prefixing
// =============================================================================

func TestFileProvider_HostRootPrefixesNonDevPath(t *testing.T) {
	tmpDir := t.TempDir()
	hostRoot := t.TempDir()

	// Simulate the host filesystem mounted under HostRoot: host /usr/local/lib
	// appears at <hostRoot>/usr/local/lib inside the container.
	hostPath := "/usr/local/lib"
	if err := os.MkdirAll(filepath.Join(hostRoot, "usr", "local", "lib"), 0755); err != nil {
		t.Fatal(err)
	}

	listFile := filepath.Join(tmpDir, "base.list")
	if err := os.WriteFile(listFile, []byte(hostPath+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := &FileProvider{Dir: tmpDir, MountNames: "base", HostRoot: hostRoot}
	mounts, err := provider.GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 1 {
		t.Fatalf("expected 1 mount, got %d", len(mounts))
	}
	// The emitted HostPath must be the original host path, without the prefix.
	if mounts[0].HostPath != hostPath {
		t.Errorf("HostPath = %q, want %q (no host root prefix)", mounts[0].HostPath, hostPath)
	}
	if mounts[0].ContainerPath != hostPath {
		t.Errorf("ContainerPath = %q, want %q", mounts[0].ContainerPath, hostPath)
	}
}

func TestFileProvider_HostRootDevPathNoPrefix(t *testing.T) {
	if _, err := os.Stat("/dev/null"); err != nil {
		t.Skip("skipping: /dev/null not available")
	}
	tmpDir := t.TempDir()
	hostRoot := t.TempDir()

	// /dev/null is visible directly in the (privileged) container, so it must
	// be stat'd without the HostRoot prefix. If it were prefixed, the stat
	// would fail (<hostRoot>/dev/null does not exist) and the entry would be
	// skipped.
	devPath := "/dev/null"
	listFile := filepath.Join(tmpDir, "base.list")
	if err := os.WriteFile(listFile, []byte(devPath+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := &FileProvider{Dir: tmpDir, MountNames: "base", HostRoot: hostRoot}
	mounts, err := provider.GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 1 {
		t.Fatalf("expected 1 mount for /dev path, got %d", len(mounts))
	}
	if mounts[0].HostPath != devPath {
		t.Errorf("HostPath = %q, want %q", mounts[0].HostPath, devPath)
	}
}

func TestFileProvider_HostRootMissingPathSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	hostRoot := t.TempDir()

	// The path does not exist under HostRoot, so it should be skipped.
	missingPath := "/usr/local/missing"
	listFile := filepath.Join(tmpDir, "base.list")
	if err := os.WriteFile(listFile, []byte(missingPath+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := &FileProvider{Dir: tmpDir, MountNames: "base", HostRoot: hostRoot}
	mounts, err := provider.GetMounts()
	if err != nil {
		t.Fatalf("GetMounts returned error: %v", err)
	}
	if len(mounts) != 0 {
		t.Fatalf("expected 0 mounts for missing path under HostRoot, got %d", len(mounts))
	}
}
