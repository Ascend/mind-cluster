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

// Package mount — file.go
//
// FileProvider: reads mount entries from *.list files in a directory.
package mount

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cdispec "tags.cncf.io/container-device-interface/specs-go"

	"ascend-common/common-utils/hwlog"
)

const (
	// maxEntriesPerFile limits the number of mount entries in a single .list file.
	maxEntriesPerFile = 128
)

// FileProvider reads mount entries from *.list files in a directory.
// Each non-comment, non-blank line is treated as a host path and generates
// a bind mount with rbind, rprivate, and ro options.
type FileProvider struct {
	// Dir is the directory containing .list files.
	Dir string
	// MountNames is a comma-separated list of mount configuration names.
	// Each name corresponds to a <name>.list file under Dir.
	// If empty, defaults to "base" (reads only base.list).
	MountNames string
	// HostRoot is the path where the host root filesystem is mounted inside
	// the container (e.g. "/host"). When non-empty, existence checks for
	// non-/dev paths are performed against <HostRoot><path>, since the host
	// filesystem is only visible under that prefix. /dev device nodes are
	// stat'd directly (visible via privileged). The emitted HostPath remains
	// the original host path without the prefix. An empty HostRoot disables
	// prefixing (compatible with the legacy ascend-docker-runtime usage).
	HostRoot string
}

// GetMounts reads specific .list files from the directory based on MountNames,
// parses each line as an absolute host path, and returns corresponding Mount
// entries. Non-existent paths are skipped with a warning. An empty or missing
// directory returns an empty list without error.
func (p *FileProvider) GetMounts() ([]*cdispec.Mount, error) {
	dirInfo, err := os.Stat(p.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			hwlog.RunLog.Warnf("FileMountProvider: directory not found: %s", p.Dir)
			return nil, nil
		}
		return nil, fmt.Errorf("stat directory %s: %w", p.Dir, err)
	}

	if !dirInfo.IsDir() {
		hwlog.RunLog.Warnf("FileMountProvider: path is not a directory: %s", p.Dir)
		return nil, nil
	}

	names := parseMountNames(p.MountNames)
	mounts := make([]*cdispec.Mount, 0)
	for _, name := range names {
		listFile := filepath.Join(p.Dir, name+".list")
		if _, err := os.Stat(listFile); err != nil {
			if os.IsNotExist(err) {
				hwlog.RunLog.Warnf("FileMountProvider: list file not found: %s", listFile)
				continue
			}
			return nil, fmt.Errorf("stat list file %s: %w", listFile, err)
		}
		entries, err := p.readListFile(listFile)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", listFile, err)
		}
		mounts = append(mounts, entries...)
	}

	return mounts, nil
}

// parseMountNames splits a comma-separated MountNames string into a slice of
// individual names. Empty or blank entries are skipped. If the input is empty,
// returns ["base"] to align with legacy parseMounts behavior.
func parseMountNames(mountNames string) []string {
	if mountNames == "" {
		return []string{"base"}
	}
	names := make([]string, 0)
	for _, name := range strings.Split(mountNames, ",") {
		name = strings.TrimSpace(name)
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

// readListFile parses a single .list file and returns Mount entries.
func (p *FileProvider) readListFile(path string) ([]*cdispec.Mount, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	mounts := make([]*cdispec.Mount, 0)
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
			hwlog.RunLog.Warnf("FileMountProvider: skip invalid path %q in %s: %v", line, path, err)
			continue
		}

		if err := StatHostPath(p.HostRoot, absPath); err != nil {
			hwlog.RunLog.Warnf("FileMountProvider: skip non-existent path %q in %s: %v", absPath, path, err)
			continue
		}

		mounts = append(mounts, &cdispec.Mount{
			HostPath:      absPath,
			ContainerPath: absPath,
			Type:          mountTypeBind,
			Options:       []string{mountOptRBind, mountOptRPrivate, mountOptReadOnly},
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
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
