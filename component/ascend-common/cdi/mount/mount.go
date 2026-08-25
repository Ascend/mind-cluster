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

// Package mount provides the unified CDI mount engine: a mode-agnostic
// MountConfig descriptor, the two supported modes (legacy
// ascend-docker-runtime .list files and the JSON mount profile), and the
// shared entry→mount builder.
//
// Files:
//   - mount.go:          core types, constants, the Build entry point, and the
//     shared entry→mount builder with its filesystem helpers.
//   - legacy_config.go:  legacy .list mode reader and HCCL topology injection.
//   - profile.go:        mount profile (mounts.json) reader/writer.
package mount

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	cdispec "tags.cncf.io/container-device-interface/specs-go"

	"ascend-common/common-utils/hwlog"
)

// ---------------------------------------------------------------------------
// Mount type constants
// ---------------------------------------------------------------------------

const (
	mountTypeBind    = "bind"
	mountOptRBind    = "rbind"
	mountOptRPrivate = "rprivate"
	mountOptReadOnly = "ro"
)

// MountConfig holds the mount-generation configuration passed to Build.
type MountConfig struct {
	// DisableMounts skips mount generation when true (NODRV mode).
	DisableMounts bool

	// DisableUBMounts suppresses UB user-space files.
	DisableUBMounts bool

	// AllowLink permits symlink entries.
	AllowLink bool

	// Dir is the directory containing the mount configuration: *.list files
	// (list mode) or mounts.json (JSON mode).
	Dir string

	// Names is only used by list mode: a comma-separated list of collection
	// names, each corresponding to a <name>.list file under Dir. Empty
	// defaults to "base".
	Names string

	// IsAscendDockerRuntime selects the legacy ascend-docker-runtime mode
	// (list mode); false (zero value) selects JSON mode.
	IsAscendDockerRuntime bool

	// HostRoot prefixes stat targets for non-/dev paths when non-empty.
	HostRoot string
}

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

// Build is the single entry point of the mount engine. It reads mount entries
// from cfg (list or JSON mode, selected by IsAscendDockerRuntime), builds
// the container mounts, and — for list mode only — appends the HCCL topology
// mounts (JSON mode carries the topology paths data-driven in mounts.json).
// DisableMounts skips mount generation entirely (NODRV mode, handled here so
// callers can pass the full config unconditionally); DisableUBMounts suppresses
// UB user-space files; AllowLink permits symlink entries; HostRoot, when
// non-empty, prefixes the stat target for non-/dev paths (see StatHostPath).
func Build(cfg MountConfig, devType string) ([]*cdispec.Mount, error) {
	if cfg.DisableMounts {
		return nil, nil
	}
	var (
		entries    []MountEntry
		ubIncluded bool
		err        error
	)
	if cfg.IsAscendDockerRuntime {
		entries, ubIncluded, err = readListEntries(cfg.Dir, cfg.Names, cfg.DisableUBMounts)
	} else {
		entries, ubIncluded, err = readProfileEntries(cfg.Dir, devType, cfg.DisableUBMounts)
	}
	if err != nil {
		return nil, fmt.Errorf("cdi: get mounts: %w", err)
	}
	// UB driver files are symlinks; force allow-link when they are mounted.
	allowLink := cfg.AllowLink
	if ubIncluded {
		allowLink = true
	}
	mounts, err := buildMounts(entries, cfg.HostRoot, allowLink)
	if err != nil {
		return nil, fmt.Errorf("cdi: build mounts: %w", err)
	}
	if cfg.IsAscendDockerRuntime {
		mounts = appendTopology(mounts, cfg.HostRoot)
	}
	return mounts, nil
}

// roBindOptions returns the read-only bind-mount options shared by every
// generated mount.
func roBindOptions() []string {
	return []string{mountOptRBind, mountOptRPrivate, mountOptReadOnly}
}

// buildMounts converts normalized entries into CDI bind mounts. Glob entries
// (paths containing wildcard characters) are expanded with symlink containment
// validation (expandGlobPath); every resulting path is existence-checked (with
// HostRoot prefixing via StatHostPath) and symlink ownership is verified via
// checkSymlinkOwner.
func buildMounts(entries []MountEntry, hostRoot string, allowLink bool) ([]*cdispec.Mount, error) {
	mounts := make([]*cdispec.Mount, 0, len(entries))
	for _, entry := range entries {
		for _, path := range entry.Paths {
			if containsGlob(path) {
				for _, matched := range expandGlobPath(path, allowLink, hostRoot) {
					if err := StatHostPath(hostRoot, matched); err != nil {
						hwlog.RunLog.Warnf("Mount: skip non-existent glob match %q: %v", matched, err)
						continue
					}
					mounts = append(mounts, &cdispec.Mount{
						HostPath:      matched,
						ContainerPath: matched,
						Type:          mountTypeBind,
						Options:       roBindOptions(),
					})
				}
				continue
			}

			if err := StatHostPath(hostRoot, path); err != nil {
				hwlog.RunLog.Warnf("Mount: skip non-existent path %q: %v", path, err)
				continue
			}

			if !checkSymlinkOwner(path, allowLink, hostRoot) {
				continue
			}

			mounts = append(mounts, &cdispec.Mount{
				HostPath:      path,
				ContainerPath: path,
				Type:          mountTypeBind,
				Options:       roBindOptions(),
			})
		}
	}

	return mounts, nil
}

// StatHostPath stats the host path for existence. When hostRoot is set,
// non-/dev paths are prefixed with hostRoot because the host filesystem is
// mounted under that prefix inside the container. /dev device nodes are
// visible directly (via privileged) and are stat'd without the prefix. The
// path is checked as-is when hostRoot is empty (host filesystem visible
// directly, e.g. ascend-docker-runtime usage).
func StatHostPath(hostRoot, absPath string) error {
	if hostRoot == "" || strings.HasPrefix(absPath, "/dev/") {
		_, err := os.Stat(absPath)
		return err
	}
	_, err := os.Stat(filepath.Join(hostRoot, absPath))
	return err
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
