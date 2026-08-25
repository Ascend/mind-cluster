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

// Package mount — legacy_config.go
//
// Legacy .list mode reader: reads mount entries from <name>.list files in a
// directory (one host path per line), plus the HCCL topology bind-mount
// injection used by list mode.
package mount

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	cdispec "tags.cncf.io/container-device-interface/specs-go"

	"ascend-common/common-utils/hwlog"
)

const (
	// maxEntriesPerFile limits the number of mount entries in a single .list file.
	maxEntriesPerFile = 128

	// ubDriverConfig is the mount collection name for UB driver user-space files
	// (corresponds to ub_driver.list under Dir).
	ubDriverConfig = "ub_driver"
)

// ---------------------------------------------------------------------------
// HCCL topology injection (list mode only)
// ---------------------------------------------------------------------------

const (
	hcclRootInfoPath = "/etc/hccl_rootinfo.json"
	topoDirPath      = "/usr/local/Ascend/driver/topo"
)

// TopologyItems lists host paths that should be bind-mounted into the
// container when present (list mode only). Exported so tests can substitute
// temporary files.
var TopologyItems = []TopologyItem{
	{hcclRootInfoPath, []string{mountOptRBind, mountOptRPrivate, mountOptReadOnly}},
	{topoDirPath, []string{mountOptRBind, mountOptRPrivate, mountOptReadOnly}},
}

// TopologyItem pairs a host path with its bind-mount options.
type TopologyItem struct {
	HostPath string
	Options  []string
}

// appendTopology appends bind-mounts for HCCL topology artifacts when they
// exist on the host. List mode only.
func appendTopology(mounts []*cdispec.Mount, hostRoot string) []*cdispec.Mount {
	for _, item := range TopologyItems {
		if err := StatHostPath(hostRoot, item.HostPath); err != nil {
			continue
		}
		mounts = append(mounts, &cdispec.Mount{
			ContainerPath: item.HostPath,
			HostPath:      item.HostPath,
			Type:          mountTypeBind,
			Options:       item.Options,
		})
	}
	return mounts
}

// readListEntries reads <name>.list files from dir. Each non-comment,
// non-blank line is treated as a host path. Empty names defaults to ["base"];
// when disableUBMounts is false and ub_driver.list exists, the ub_driver
// collection is appended and ubIncluded is set (the caller forces allow-link
// in that case, since UB driver files are symlinks).
func readListEntries(dir, names string, disableUBMounts bool) ([]MountEntry, bool, error) {
	dirInfo, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			hwlog.RunLog.Warnf("Mount: directory not found: %s", dir)
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("stat directory %s: %w", dir, err)
	}

	if !dirInfo.IsDir() {
		hwlog.RunLog.Warnf("Mount: path is not a directory: %s", dir)
		return nil, false, nil
	}

	nameList := parseMountNames(names)
	ubIncluded := false
	if !disableUBMounts {
		ubListFile := filepath.Join(dir, ubDriverConfig+".list")
		if _, err := os.Stat(ubListFile); err == nil {
			if !slices.Contains(nameList, ubDriverConfig) {
				nameList = append(nameList, ubDriverConfig)
			}
			ubIncluded = true
		} else if !os.IsNotExist(err) {
			hwlog.RunLog.Warnf("Mount: stat %s failed: %v", ubListFile, err)
		}
	}

	entries := make([]MountEntry, 0)
	for _, name := range nameList {
		listFile := filepath.Join(dir, name+".list")
		if _, err := os.Stat(listFile); err != nil {
			if os.IsNotExist(err) {
				hwlog.RunLog.Warnf("Mount: list file not found: %s", listFile)
				continue
			}
			return nil, false, fmt.Errorf("stat list file %s: %w", listFile, err)
		}
		fileEntries, err := parseListFile(listFile)
		if err != nil {
			return nil, false, fmt.Errorf("read %s: %w", listFile, err)
		}
		entries = append(entries, fileEntries...)
	}

	return entries, ubIncluded, nil
}

// parseMountNames splits a comma-separated mountNames string into a slice of
// individual names. Empty or blank entries are skipped, and duplicate names
// are deduplicated (first occurrence wins). If the input is empty, returns
// ["base"] to align with legacy parseMounts behavior.
func parseMountNames(mountNames string) []string {
	if mountNames == "" {
		return []string{"base"}
	}
	var names []string
	for _, name := range strings.Split(mountNames, ",") {
		name = strings.TrimSpace(name)
		if name != "" && !slices.Contains(names, name) {
			names = append(names, name)
		}
	}
	return names
}

// parseListFile parses a single .list file and returns normalized entries.
// Each non-comment, non-blank line becomes one MountEntry. Glob detection and
// existence/symlink checks are deferred to buildMounts.
func parseListFile(path string) ([]MountEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	entries := make([]MountEntry, 0)
	scanner := bufio.NewScanner(f)
	entryCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" || line[0] == '#' {
			continue
		}

		entryCount++
		if entryCount > maxEntriesPerFile {
			return nil, fmt.Errorf("mount list too long in %s (max %d entries)", path, maxEntriesPerFile)
		}

		absPath, err := filepath.Abs(line)
		if err != nil {
			hwlog.RunLog.Warnf("Mount: skip invalid path %q in %s: %v", line, path, err)
			continue
		}

		entries = append(entries, MountEntry{Paths: []string{absPath}})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}

	return entries, nil
}
