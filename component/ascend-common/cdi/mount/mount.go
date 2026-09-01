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
	"errors"
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

	// MountUBDrv mounts UB user-space files when true.
	MountUBDrv bool

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

	// HostFsPrefix prefixes stat targets for non-/dev paths when non-empty.
	HostFsPrefix string
}

// ubType marks a mount entry group as UB user-space driver files, gated by
// the external MountUBDrv option.
const ubType = "UB"

// MountEntry is a normalized group of mount entries read from a mount source,
// before glob expansion and existence checks (see buildMounts).
type MountEntry struct {
	// Paths are the host paths in this group, or glob patterns when they
	// contain wildcard characters (*?[).
	Paths []string `json:"path"`
	// Type marks the group; "UB" for UB user-space files (gated by
	// MountUBDrv), empty otherwise.
	Type string `json:"type,omitempty"`
}

// Build is the single entry point of the mount engine. It reads mount entries
// from cfg (list or JSON mode, selected by IsAscendDockerRuntime), builds
// the container mounts, and — for list mode only — appends the HCCL topology
// mounts (JSON mode carries the topology paths data-driven in mounts.json).
// DisableMounts skips mount generation entirely (NODRV mode, handled here so
// callers can pass the full config unconditionally); MountUBDrv mounts
// UB user-space files; AllowLink permits symlink entries; HostFsPrefix, when
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
		entries, ubIncluded, err = readListEntries(cfg.Dir, cfg.Names, cfg.MountUBDrv)
	} else {
		entries, ubIncluded, err = readProfileEntries(cfg.Dir, devType, cfg.MountUBDrv)
	}
	if err != nil {
		return nil, fmt.Errorf("cdi: get mounts: %w", err)
	}
	// UB driver files are symlinks; force allow-link when they are mounted.
	allowLink := cfg.AllowLink
	if ubIncluded {
		allowLink = true
	}
	mounts, err := buildMounts(entries, cfg.HostFsPrefix, allowLink)
	if err != nil {
		return nil, fmt.Errorf("cdi: build mounts: %w", err)
	}
	if cfg.IsAscendDockerRuntime {
		mounts = appendTopology(mounts, cfg.HostFsPrefix)
	}
	return mounts, nil
}

// roBindOptions returns the read-only bind-mount options shared by every
// generated mount.
func roBindOptions() []string {
	return []string{mountOptRBind, mountOptRPrivate, mountOptReadOnly}
}

// buildMounts converts normalized entries into CDI bind mounts. Glob entries
// (paths containing wildcard characters) are expanded and validated by
// expandGlobPath (symlink containment + ownership); plain entries are
// existence-checked (via StatHostPath) and symlink-validated via
// checkSymlinkOwner.
func buildMounts(entries []MountEntry, hostFsPrefix string, allowLink bool) ([]*cdispec.Mount, error) {
	mounts := make([]*cdispec.Mount, 0, len(entries))
	for _, entry := range entries {
		for _, path := range entry.Paths {
			if containsGlob(path) {
				for _, matched := range expandGlobPath(path, allowLink, hostFsPrefix) {
					mounts = append(mounts, newBindMount(matched))
				}
				continue
			}

			if err := StatHostPath(hostFsPrefix, path); err != nil {
				hwlog.RunLog.Warnf("Mount: skip non-existent path %q: %v", path, err)
				continue
			}
			if !checkSymlinkOwner(path, allowLink, hostFsPrefix) {
				continue
			}
			mounts = append(mounts, newBindMount(path))
		}
	}

	return mounts, nil
}

// newBindMount returns a read-only bind mount whose host and container paths
// are identical, as required by CDI for driver library files.
func newBindMount(hostPath string) *cdispec.Mount {
	return &cdispec.Mount{
		HostPath:      hostPath,
		ContainerPath: hostPath,
		Type:          mountTypeBind,
		Options:       roBindOptions(),
	}
}

// toContainerPath maps a host path to the path accessible inside the
// container: when hostFsPrefix is set, non-/dev paths are prefixed with
// hostFsPrefix because the host filesystem is mounted under that prefix; /dev
// device nodes are visible directly (privileged) and are returned as-is.
// With an empty hostFsPrefix the host filesystem is visible directly and the
// path is unchanged.
//
// This is the driver container's stat/access path, not the CDI
// Mount.ContainerPath (the bind-mount target inside the workload container).
func toContainerPath(hostFsPrefix, hostPath string) string {
	if hostFsPrefix == "" || strings.HasPrefix(hostPath, "/dev/") {
		return hostPath
	}
	return filepath.Join(hostFsPrefix, hostPath)
}

// toHostPath maps a container-accessible path back to the host path by
// stripping the hostFsPrefix added by toContainerPath. It is a no-op when
// hostFsPrefix is empty or the path is not actually prefixed (e.g. a /dev path).
func toHostPath(hostFsPrefix, fsPath string) string {
	if hostFsPrefix == "" {
		return fsPath
	}
	return strings.TrimPrefix(fsPath, filepath.Clean(hostFsPrefix))
}

// StatHostPath stats the host path for existence. Lstat is used rather than
// Stat so a symlink entry is not resolved through its target: an absolute
// target would be resolved against the container's own filesystem when
// hostFsPrefix is set, not the host filesystem.
func StatHostPath(hostFsPrefix, absPath string) error {
	_, err := os.Lstat(toContainerPath(hostFsPrefix, absPath))
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
//
// When hostFsPrefix is non-empty, the glob is expanded under the host root (the
// host filesystem is mounted under that prefix inside the container), but the
// returned paths are the un-prefixed host paths, so callers can emit them as
// HostPath/ContainerPath directly.
func expandGlobPath(pattern string, allowLink bool, hostFsPrefix string) []string {
	fsPattern := toContainerPath(hostFsPrefix, pattern)
	matches, err := filepath.Glob(fsPattern)
	if err != nil {
		hwlog.RunLog.Warnf("failed to glob %s: %v", fsPattern, err)
		return nil
	}
	// expectedDir is the host-path directory a resolved symlink target must
	// stay in (anti-escape).
	expectedDir := filepath.Dir(pattern)

	var paths []string
	for _, match := range matches {
		hostPath := toHostPath(hostFsPrefix, match)

		realHostPath, err := resolveSymlinkTarget(hostFsPrefix, match)
		if err != nil {
			hwlog.RunLog.Warnf("failed to resolve symlink %s: %v", hostPath, err)
			continue
		}
		if filepath.Dir(realHostPath) != expectedDir {
			hwlog.RunLog.Warnf("symlink %s points to %s outside expected dir %s", hostPath, realHostPath, expectedDir)
			continue
		}

		if !checkSymlinkOwner(hostPath, allowLink, hostFsPrefix) {
			continue
		}
		paths = append(paths, hostPath)
	}
	return paths
}

// maxSymlinkDepth bounds symlink-chain resolution to avoid loops. It mirrors
// Go's filepath.EvalSymlinks limit (path/filepath/symlink.go linksWalked).
const maxSymlinkDepth = 255

// resolveSymlinkTarget resolves the symlink chain starting at fsPath (a path
// accessible inside the container) and returns the final target as a host
// path. An absolute symlink target is written in host-path form, so it is
// re-rooted under hostFsPrefix (via toContainerPath) to stay reachable inside the
// container; otherwise resolution would land in the container's own filesystem.
func resolveSymlinkTarget(hostFsPrefix, fsPath string) (string, error) {
	hostPath := toHostPath(hostFsPrefix, fsPath)
	for i := 0; i < maxSymlinkDepth; i++ {
		info, err := os.Lstat(fsPath)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink == 0 {
			return hostPath, nil
		}
		target, err := os.Readlink(fsPath)
		if err != nil {
			return "", err
		}
		if !filepath.IsAbs(target) {
			// Relative target: resolve against the symlink's directory in both
			// namespaces.
			hostPath = filepath.Join(filepath.Dir(hostPath), target)
			fsPath = filepath.Join(filepath.Dir(fsPath), target)
			continue
		}
		// Absolute target is already in host-path form; re-map to the container
		// view so the next iteration can stat it.
		hostPath = target
		fsPath = toContainerPath(hostFsPrefix, target)
	}
	return "", errors.New("too many levels of symbolic links")
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
// (UID 0). The symlink target is resolved in the host filesystem namespace
// (an absolute target is re-rooted under hostFsPrefix) so non-/dev paths are
// checked against the host root rather than the container's own filesystem.
func checkSymlinkOwner(absPath string, allowLink bool, hostFsPrefix string) bool {
	statPath := toContainerPath(hostFsPrefix, absPath)
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
	// Resolve the symlink target in the host namespace, then stat its ownership
	// through the container view.
	targetHostPath, err := resolveSymlinkTarget(hostFsPrefix, statPath)
	if err != nil {
		hwlog.RunLog.Warnf("%s may not exists, error: %v", statPath, err)
		return false
	}
	realStat, err := os.Stat(toContainerPath(hostFsPrefix, targetHostPath))
	if err != nil {
		hwlog.RunLog.Warnf("%s may not exists, error: %v", targetHostPath, err)
		return false
	}
	if getFileUID(linkStat) != 0 || getFileUID(realStat) != 0 {
		hwlog.RunLog.Warnf("symlink %s or its target not owned by root", absPath)
		return false
	}
	return true
}
