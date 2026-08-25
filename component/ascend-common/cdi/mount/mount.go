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

// Package mount — mount.go
//
// Mount config types and the shared entry→mount builder with its filesystem
// helpers (glob expansion, symlink ownership).
package mount

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"ascend-common/common-utils/hwlog"
)

// ubType marks a mount entry group as UB user-space driver files, gated by
// the external DisableUBMounts option.
const ubType = "UB"

// MountEntry is a normalized group of mount entries read from a mount source,
// before glob expansion and existence checks (see buildMounts).
type MountEntry struct {
	// Paths are the host paths in this group, or glob patterns when they
	// contain wildcard characters (*?[).
	Paths []string `json:"path"`

	// Type marks the group; "UB" for UB user-space files (gated by
	// DisableUBMounts), empty otherwise.
	Type string `json:"type,omitempty"`
}

// containsGlob checks if a path contains glob wildcard characters.
func containsGlob(path string) bool {
	return strings.ContainsAny(path, "*?[")
}

// expandGlobPath expands a glob pattern and returns matched paths (files and
// directories) with symlink validation: the symlink target must stay in the
// same directory as the pattern (anti-escape) and both the symlink and its
// target must be owned by root (anti-hijack). Mirrors the legacy
// ascend-docker-runtime hook/process/process.go expandGlobPath semantics.
func expandGlobPath(pattern string, allowLink bool, hostRoot string) []string {
	var paths []string
	matches, err := filepath.Glob(pattern)
	if err != nil {
		hwlog.RunLog.Warnf("failed to glob %s: %v", pattern, err)
		return paths
	}
	expectedDir := filepath.Dir(pattern)
	for _, match := range matches {
		realPath, err := filepath.EvalSymlinks(match)
		if err != nil {
			hwlog.RunLog.Warnf("failed to resolve symlink %s: %v", match, err)
			continue
		}
		if filepath.Dir(realPath) != expectedDir {
			hwlog.RunLog.Warnf("symlink %s points to %s outside expected dir %s", match, realPath, expectedDir)
			continue
		}
		if !checkSymlinkOwner(match, allowLink, hostRoot) {
			continue
		}
		paths = append(paths, match)
	}
	return paths
}

// getFileUID extracts the UID from os.FileInfo. Returns 0 if extraction fails.
func getFileUID(info os.FileInfo) uint32 {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return stat.Uid
	}
	return 0
}

// checkSymlinkOwner verifies that the entry at absPath is mountable:
// non-symlinks pass through; symlinks are rejected when allowLink is
// disabled, otherwise both the symlink and its target must be owned by root
// (UID 0). The hostRoot prefix is applied to the stat target exactly like
// StatHostPath, so non-/dev paths are resolved under the host root.
func checkSymlinkOwner(absPath string, allowLink bool, hostRoot string) bool {
	statPath := absPath
	if hostRoot != "" && !strings.HasPrefix(absPath, "/dev/") {
		statPath = filepath.Join(hostRoot, absPath)
	}
	linkStat, err := os.Lstat(statPath)
	if err != nil {
		hwlog.RunLog.Warnf("failed to lstat %s: %v", statPath, err)
		return false
	}
	if linkStat.Mode()&os.ModeSymlink == 0 {
		return true
	}
	if !allowLink {
		hwlog.RunLog.Warnf("skip symlink %s: allow-link is disabled", absPath)
		return false
	}
	realStat, err := os.Stat(statPath)
	if err != nil {
		hwlog.RunLog.Warnf("%s may not exists, error: %v", statPath, err)
		return false
	}
	if getFileUID(linkStat) != 0 || getFileUID(realStat) != 0 {
		hwlog.RunLog.Warnf("symlink %s or its target not owned by root", absPath)
		return false
	}
	return true
}
