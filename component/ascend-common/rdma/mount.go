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
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"ascend-common/cdi/mount"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

const optionalRdmaProviderPath = "/usr/lib64/libibverbs/libhrn5-rdmav34.so"

// ReadMountEntries reads and validates one RDMA list and returns ready-to-mount entries.
func ReadMountEntries(configDir, collectionName string) ([]string, []string, error) {
	collectionName, err := normalizeRdmaCollection(collectionName)
	if err != nil {
		return nil, nil, err
	}
	listPath := filepath.Join(configDir, collectionName+".list")
	if err := checkRdmaConfigFile(listPath); err != nil {
		return nil, nil, fmt.Errorf("RDMA collection %s: %w", collectionName, err)
	}
	listFile, err := os.Open(listPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open RDMA collection %s: %w", collectionName, err)
	}
	defer listFile.Close()
	return readRdmaMountEntries(listFile, listPath, os.Lstat)
}

func readRdmaMountEntries(
	reader io.Reader, listPath string, lstat func(string) (os.FileInfo, error),
) ([]string, []string, error) {
	scanner := bufio.NewScanner(reader)
	var fileEntries, dirEntries []string
	count, lineNo := 0, 0
	for scanner.Scan() {
		lineNo++
		entry := strings.TrimSpace(scanner.Text())
		if entry == "" || strings.HasPrefix(entry, "#") {
			continue
		}
		count++
		if count > mount.MaxEntriesPerFile {
			return nil, nil, fmt.Errorf("mount list too long in %s (max %d entries)", listPath, mount.MaxEntriesPerFile)
		}
		files, dirs, err := readRdmaMountEntry(entry, lstat)
		if err != nil {
			return nil, nil, fmt.Errorf("RDMA list %s line %d source %q: %w", listPath, lineNo, entry, err)
		}
		fileEntries = append(fileEntries, files...)
		dirEntries = append(dirEntries, dirs...)
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("scan RDMA list %s near line %d: %w", listPath, lineNo+1, err)
	}
	if count == 0 {
		return nil, nil, fmt.Errorf("RDMA list %s has no effective entries", listPath)
	}
	return fileEntries, dirEntries, nil
}

func readRdmaMountEntry(entry string, lstat func(string) (os.FileInfo, error)) ([]string, []string, error) {
	sourcePath, targetPath, err := mount.ParseMountEntry(entry)
	if err != nil {
		return nil, nil, err
	}
	if !filepath.IsAbs(sourcePath) || strings.Contains(sourcePath, "..") || len(sourcePath) > mount.MaxContainerPathLen {
		return nil, nil, fmt.Errorf("invalid absolute host path %q", sourcePath)
	}
	if strings.ContainsAny(sourcePath, "*?[") {
		return nil, nil, fmt.Errorf("RDMA mount sources do not support wildcards: %q", sourcePath)
	}
	if sourcePath == optionalRdmaProviderPath && targetPath == "" {
		if _, err := lstat(sourcePath); errors.Is(err, os.ErrNotExist) {
			hwlog.RunLog.Warnf("optional RDMA source %s is absent; skipping", sourcePath)
			return nil, nil, nil
		}
	}
	sourceInfo, err := checkStrictRdmaSource(sourcePath, lstat)
	if err != nil {
		return nil, nil, err
	}
	mountEntry := sourcePath
	if targetPath != "" {
		mountEntry += ":" + targetPath
	}
	if sourceInfo.IsDir() {
		return nil, []string{mountEntry}, nil
	}
	return []string{mountEntry}, nil, nil
}

func checkStrictRdmaSource(sourcePath string, lstat func(string) (os.FileInfo, error)) (os.FileInfo, error) {
	linkInfo, err := lstat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("stat required source %s: %w", sourcePath, err)
	}
	if linkInfo.Mode()&os.ModeSymlink != 0 && !mount.CheckSymlinkOwner(sourcePath, true, "") {
		return nil, fmt.Errorf("source symlink %s must have a root-owned existing target and link", sourcePath)
	}
	resolvedSourcePath, err := filepath.EvalSymlinks(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("resolve source %s: %w", sourcePath, err)
	}
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("stat required source %s: %w", sourcePath, err)
	}
	if err := checkStrictRdmaSourceType(sourcePath, sourceInfo); err != nil {
		return nil, err
	}
	if resolvedSourcePath != sourcePath {
		if err := checkStrictRdmaSourceType(resolvedSourcePath, sourceInfo); err != nil {
			return nil, err
		}
	}
	return sourceInfo, nil
}

func checkStrictRdmaSourceType(sourcePath string, sourceInfo os.FileInfo) error {
	var err error
	if sourceInfo.IsDir() {
		_, err = utils.RealDirChecker(sourcePath, true, true)
	} else {
		_, err = utils.RealFileChecker(sourcePath, true, true, rdmaConfigMaxSizeMB)
	}
	if err != nil {
		return fmt.Errorf("check required source %s: %w", sourcePath, err)
	}
	return nil
}
