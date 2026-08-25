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

// Package mount — mount_test.go
//
// Tests for the glob expansion and symlink-ownership helpers used by the
// build pipeline.
package mount

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// fakeFileInfo is a minimal os.FileInfo whose Sys() returns nil, simulating
// a platform where *syscall.Stat_t is not available.
type fakeFileInfo struct{ name string }

func (f fakeFileInfo) Name() string       { return f.name }
func (f fakeFileInfo) Size() int64        { return 0 }
func (f fakeFileInfo) Mode() os.FileMode  { return 0644 }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool        { return false }
func (f fakeFileInfo) Sys() interface{}   { return nil }

func TestContainsGlob(t *testing.T) {
	for _, p := range []string{"/usr/lib64/libummu*", "/usr/lib64/lib?.so", "/usr/lib64/lib[ab].so", "*.so"} {
		if !containsGlob(p) {
			t.Errorf("containsGlob(%q) = false, want true", p)
		}
	}
	if containsGlob("/usr/lib64/libummu.so") {
		t.Error("containsGlob(plain path) = true, want false")
	}
}

// TestExpandGlobPath covers normal expansion and a no-match pattern.
func TestExpandGlobPath(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"libummu.so.1", "libummu.so.2", "liburma.so"} {
		mustWriteFile(t, filepath.Join(tmpDir, name), "")
	}

	tests := []struct {
		name    string
		pattern string
		want    map[string]bool
	}{
		{
			name:    "normal expansion",
			pattern: filepath.Join(tmpDir, "libummu*"),
			want: map[string]bool{
				filepath.Join(tmpDir, "libummu.so.1"): true,
				filepath.Join(tmpDir, "libummu.so.2"): true,
			},
		},
		{name: "no match", pattern: filepath.Join(t.TempDir(), "no_match*")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := expandGlobPath(tt.pattern, true, "")
			if len(paths) != len(tt.want) {
				t.Fatalf("expected %d matches, got %d: %v", len(tt.want), len(paths), paths)
			}
			for _, p := range paths {
				if !tt.want[p] {
					t.Errorf("unexpected match %q", p)
				}
			}
		})
	}
}

func TestExpandGlobPath_SymlinkEscaping(t *testing.T) {
	tmpDir := t.TempDir()
	outside := t.TempDir()

	// Same-dir symlink: target stays within the pattern directory.
	realLib := filepath.Join(tmpDir, "real_lib.so")
	mustWriteFile(t, realLib, "")
	sameDirLink := filepath.Join(tmpDir, "libin.so")
	if err := os.Symlink(realLib, sameDirLink); err != nil {
		t.Fatal(err)
	}

	// Escaping symlink: target lives in another directory and must be skipped.
	outsideTarget := filepath.Join(outside, "target.so")
	mustWriteFile(t, outsideTarget, "")
	escapeLink := filepath.Join(tmpDir, "libout.so")
	if err := os.Symlink(outsideTarget, escapeLink); err != nil {
		t.Fatal(err)
	}

	paths := expandGlobPath(filepath.Join(tmpDir, "lib*.so"), true, "")
	for _, p := range paths {
		if p == escapeLink {
			t.Errorf("escaping symlink %q must be skipped", p)
		}
	}
	if os.Geteuid() == 0 {
		// Root-owned symlink in the same directory is kept.
		if len(paths) != 1 || paths[0] != sameDirLink {
			t.Errorf("matches = %v, want [%s]", paths, sameDirLink)
		}
	} else {
		// Non-root-owned symlink is skipped by the owner check.
		for _, p := range paths {
			if p == sameDirLink {
				t.Errorf("non-root-owned symlink %q should be skipped", p)
			}
		}
	}
}

func TestGetFileUID(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "f.so")
	mustWriteFile(t, tmpFile, "")
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if got := getFileUID(info); got != uint32(os.Getuid()) {
		t.Errorf("getFileUID(real file) = %d, want %d", got, os.Getuid())
	}
	if got := getFileUID(fakeFileInfo{name: "x"}); got != 0 {
		t.Errorf("getFileUID(non Stat_t) = %d, want 0", got)
	}
}

func TestCheckSymlinkOwner(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target.so")
	mustWriteFile(t, target, "")
	link := filepath.Join(tmpDir, "link.so")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	// Non-symlink files pass through regardless of owner.
	if !checkSymlinkOwner(target, true, "") {
		t.Error("checkSymlinkOwner must pass non-symlink files through")
	}

	if os.Geteuid() == 0 {
		// Root: a non-root-owned symlink is rejected, a root-owned one passes.
		if err := os.Lchown(link, 1000, 1000); err != nil {
			t.Fatal(err)
		}
		if checkSymlinkOwner(link, true, "") {
			t.Error("checkSymlinkOwner must reject non-root-owned symlink")
		}

		rootTarget := filepath.Join(tmpDir, "root_target.so")
		mustWriteFile(t, rootTarget, "")
		rootLink := filepath.Join(tmpDir, "root_link.so")
		if err := os.Symlink(rootTarget, rootLink); err != nil {
			t.Fatal(err)
		}
		if !checkSymlinkOwner(rootLink, true, "") {
			t.Error("checkSymlinkOwner must pass root-owned symlink")
		}
	} else {
		// Non-root: the created symlink is already non-root owned.
		if checkSymlinkOwner(link, true, "") {
			t.Error("checkSymlinkOwner must reject non-root-owned symlink")
		}
	}

	// Lstat failure on a missing path is rejected.
	if checkSymlinkOwner(filepath.Join(tmpDir, "missing.so"), true, "") {
		t.Error("checkSymlinkOwner must reject missing path")
	}
}

func TestCheckSymlinkOwner_AllowLinkDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target.so")
	mustWriteFile(t, target, "")
	link := filepath.Join(tmpDir, "link.so")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	// allowLink=false: symlink entries are skipped regardless of ownership,
	// while non-symlink entries still pass.
	if checkSymlinkOwner(link, false, "") {
		t.Error("checkSymlinkOwner with allowLink=false must reject symlink")
	}
	if !checkSymlinkOwner(target, false, "") {
		t.Error("checkSymlinkOwner with allowLink=false must pass non-symlink files")
	}
}
