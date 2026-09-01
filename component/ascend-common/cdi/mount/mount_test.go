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
// Tests for the Build entry point, HCCL topology injection, and the glob
// expansion and symlink-ownership helpers used by the build pipeline.
package mount

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cdispec "tags.cncf.io/container-device-interface/specs-go"
)

func TestBuild_ReadError(t *testing.T) {
	// A malformed mounts.json must surface as a Build error.
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "mounts.json"), "{not json")
	_, err := Build(MountConfig{Dir: dir}, "")
	if err == nil {
		t.Fatal("expected error from the JSON reader")
	}
	if !strings.Contains(err.Error(), "mounts.json") {
		t.Errorf("error should mention mounts.json, got: %v", err)
	}
}

func TestBuild_HostFsPrefixAppliesTopology(t *testing.T) {
	// Topology injection is list-mode-only (JSON mode is data-driven from
	// mounts.json; see TestBuildJSON_NoTopologyInjection).
	hostFsPrefix := t.TempDir()
	topoFile := filepath.Join(hostFsPrefix, "etc", "hccl_rootinfo.json")
	if err := os.MkdirAll(filepath.Dir(topoFile), 0755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, topoFile, `{"version":"1.0"}`)

	orig := TopologyItems
	defer func() { TopologyItems = orig }()
	TopologyItems = []TopologyItem{{HostPath: "/etc/hccl_rootinfo.json", Options: []string{"rbind", "rprivate", "ro"}}}

	// Empty source (no .list mounts); only the topology item should be emitted.
	mounts, err := Build(MountConfig{Dir: t.TempDir(), IsAscendDockerRuntime: true, HostFsPrefix: hostFsPrefix}, "")
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(mounts) != 1 {
		t.Fatalf("expected 1 topology mount, got %d", len(mounts))
	}
	// HostPath must carry the original host path, without the hostFsPrefix.
	if mounts[0].HostPath != "/etc/hccl_rootinfo.json" {
		t.Errorf("HostPath = %q, want /etc/hccl_rootinfo.json", mounts[0].HostPath)
	}
}

// TestBuild_TopologyListModeOnly proves the IsAscendDockerRuntime dispatch:
// list mode (true) appends the HCCL topology mounts, JSON mode (false) does
// not, even when a topology item exists on the host.
func TestBuild_TopologyListModeOnly(t *testing.T) {
	hostFsPrefix := t.TempDir()
	topoFile := filepath.Join(hostFsPrefix, "etc", "hccl_rootinfo.json")
	if err := os.MkdirAll(filepath.Dir(topoFile), 0755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, topoFile, `{"version":"1.0"}`)

	orig := TopologyItems
	defer func() { TopologyItems = orig }()
	TopologyItems = []TopologyItem{{HostPath: "/etc/hccl_rootinfo.json", Options: []string{"rbind", "rprivate", "ro"}}}

	// List mode: topology item appended after the (empty) entry list.
	listMounts, err := Build(MountConfig{Dir: t.TempDir(), IsAscendDockerRuntime: true, HostFsPrefix: hostFsPrefix}, "")
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(listMounts) != 1 || listMounts[0].HostPath != "/etc/hccl_rootinfo.json" {
		t.Fatalf("list mode: expected 1 topology mount, got %v", listMounts)
	}

	// JSON mode (zero-value IsAscendDockerRuntime): no entries and no topology.
	jsonMounts, err := Build(MountConfig{Dir: t.TempDir(), HostFsPrefix: hostFsPrefix}, "")
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(jsonMounts) != 0 {
		t.Fatalf("json mode: expected 0 mounts (no topology injection), got %v", jsonMounts)
	}
}

func TestBuild_UBDriverForcesAllowLink(t *testing.T) {
	tmpDir := t.TempDir()

	baseFile := filepath.Join(tmpDir, "base_lib.so")
	mustWriteFile(t, baseFile, "")
	writeListFile(t, tmpDir, "base", baseFile+"\n")
	// A plain-file ub_driver.list entry suffices to assert the "force" side
	// effect: ubIncluded must be true whenever ub_driver.list is present,
	// regardless of file ownership.
	ubFile := filepath.Join(tmpDir, "ub_plain.so")
	mustWriteFile(t, ubFile, "")
	writeListFile(t, tmpDir, "ub_driver", ubFile+"\n")

	entries, ubIncluded, err := readListEntries(tmpDir, "", true)
	if err != nil {
		t.Fatalf("readListEntries returned error: %v", err)
	}
	if !ubIncluded {
		t.Error("ubIncluded must be true when ub_driver.list is present")
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (base + ub_driver), got %v", entries)
	}

	if os.Geteuid() != 0 {
		return
	}

	// Root only: end-to-end check that a root-owned symlink entry in
	// ub_driver.list is mounted thanks to the forced allow-link.
	ubTarget := filepath.Join(tmpDir, "ub_lib.so")
	mustWriteFile(t, ubTarget, "")
	ubLink := filepath.Join(tmpDir, "ub_link.so")
	if err := os.Symlink(ubTarget, ubLink); err != nil {
		t.Fatal(err)
	}
	writeListFile(t, tmpDir, "ub_driver", ubLink+"\n")

	linkMounts, err := buildListMounts(t, MountConfig{Dir: tmpDir, IsAscendDockerRuntime: true, MountUBDrv: true})
	if err != nil {
		t.Fatalf("buildListMounts returned error: %v", err)
	}
	if len(linkMounts) != 2 {
		t.Fatalf("expected 2 mounts (base + ub_driver symlink), got %v", linkMounts)
	}
	got := map[string]bool{linkMounts[0].HostPath: true, linkMounts[1].HostPath: true}
	if !got[baseFile] || !got[ubLink] {
		t.Errorf("expected mounts for base and ub_driver symlink, got %v", got)
	}
}

// buildListMounts runs the full Build pipeline for a list-mode config with
// HCCL topology injection disabled, so mount counts match the pre-refactor
// list-mount behavior.
func buildListMounts(t *testing.T, cfg MountConfig) ([]*cdispec.Mount, error) {
	t.Helper()
	orig := TopologyItems
	TopologyItems = nil
	defer func() { TopologyItems = orig }()
	return Build(cfg, "")
}

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

// TestExpandGlobPath_HostFsPrefix verifies glob expansion resolves under hostFsPrefix
// (the host filesystem mount point inside the container) but returns the
// un-prefixed host paths, mirroring the HostPath/ContainerPath emitted by
// buildMounts.
func TestExpandGlobPath_HostFsPrefix(t *testing.T) {
	hostFsPrefix := t.TempDir()
	libDir := filepath.Join(hostFsPrefix, "usr", "lib64")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(libDir, "libummu.so.1"), "")
	mustWriteFile(t, filepath.Join(libDir, "libummu.so.2"), "")

	paths := expandGlobPath("/usr/lib64/libummu*", true, hostFsPrefix)
	want := map[string]bool{
		"/usr/lib64/libummu.so.1": true,
		"/usr/lib64/libummu.so.2": true,
	}
	if len(paths) != len(want) {
		t.Fatalf("expected %d matches, got %d: %v", len(want), len(paths), paths)
	}
	for _, p := range paths {
		if !want[p] {
			t.Errorf("unexpected match %q", p)
		}
	}
}

// TestResolveSymlinkTarget_HostFsPrefix verifies that absolute and relative
// symlink targets resolve in the host filesystem namespace when hostFsPrefix is
// set: both resolve to the host path (not the container's own filesystem path).
func TestResolveSymlinkTarget_HostFsPrefix(t *testing.T) {
	hostFsPrefix := t.TempDir()
	libDir := filepath.Join(hostFsPrefix, "usr", "lib64")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Real file created under hostFsPrefix; its host path is what must be returned.
	mustWriteFile(t, filepath.Join(libDir, "libummu.so.1"), "")
	const hostReal = "/usr/lib64/libummu.so.1"

	// Absolute target: must resolve to the host path without error.
	absLink := filepath.Join(libDir, "libabs.so")
	if err := os.Symlink(hostReal, absLink); err != nil {
		t.Fatal(err)
	}
	got, err := resolveSymlinkTarget(hostFsPrefix, absLink)
	if err != nil {
		t.Fatalf("resolveSymlinkTarget(abs) error: %v", err)
	}
	if got != hostReal {
		t.Errorf("resolveSymlinkTarget(abs) = %q, want %q", got, hostReal)
	}

	// Relative target: must resolve within the host namespace to the host path.
	relLink := filepath.Join(libDir, "librel.so")
	if err := os.Symlink("libummu.so.1", relLink); err != nil {
		t.Fatal(err)
	}
	got, err = resolveSymlinkTarget(hostFsPrefix, relLink)
	if err != nil {
		t.Fatalf("resolveSymlinkTarget(rel) error: %v", err)
	}
	if got != hostReal {
		t.Errorf("resolveSymlinkTarget(rel) = %q, want %q", got, hostReal)
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
