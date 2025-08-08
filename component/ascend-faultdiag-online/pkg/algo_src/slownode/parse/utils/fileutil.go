/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package utils provides some common utils
package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
)

// ReadLinesFromOffset 从指定偏移位置读取 NDJSON 文件中新增的行，返回这些行和读取后的新偏移位置
func ReadLinesFromOffset(filePath string, startOffset int64) ([]string, int64, error) {
	if err := FileValidator(filePath); err != nil {
		return nil, startOffset, err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, startOffset, err
	}
	defer file.Close()

	// 定位到上次读取位置
	_, err = file.Seek(startOffset, io.SeekStart)
	if err != nil {
		return nil, startOffset, err
	}
	reader := bufio.NewReader(file)
	lines := make([]string, 0)
	for {
		line, err := reader.ReadString('\n') // 会阻塞读取
		if err == io.EOF {
			// 文件末尾，返回当前已读行和最新偏移量
			break
		}
		if err != nil {
			return lines, startOffset, err
		}

		lines = append(lines, line)
	}

	// 获取当前偏移量，作为下一次读取的起点
	nextOffset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return lines, startOffset, err
	}
	return lines, nextOffset, nil
}

// CheckDBFilePerm 检查 SQLite 文件是否可写
func CheckDBFilePerm(dbFilePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(dbFilePath); os.IsNotExist(err) {
		// 文件不存在，检查目录是否即可读可写
		dirPath := filepath.Dir(dbFilePath)
		return checkDirReadWriteAble(dirPath)
	} else if err != nil {
		return fmt.Errorf("failed to check if file exists: %v", err)
	}
	// 文件存在,检查文件是否即可读可写
	return checkFileRedWriteAble(dbFilePath, true, true)
}

// CheckFilePerm 校验文件可读权限
func CheckFilePerm(inputFilePath string, redAble bool, writeAble bool) error {
	_, readFileErr := os.Stat(inputFilePath)
	if readFileErr != nil {
		return fmt.Errorf("file does not exist: %s", inputFilePath)
	}
	if err := FileValidator(inputFilePath); err != nil {
		return err
	}
	return checkFileRedWriteAble(inputFilePath, redAble, writeAble)
}

// IsSymbolicLink 检查文件是否为软连接
func IsSymbolicLink(path string) (bool, error) {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	// 通过文件模式判断是否为符号链接
	return (fileInfo.Mode() & os.ModeSymlink) != 0, nil
}

func checkDirReadWriteAble(dirPath string) error {
	isSymlink, err := IsSymbolicLink(dirPath)
	if err != nil {
		return fmt.Errorf("failed to check symlink: %v, dir path: %s", err, dirPath)
	}
	if isSymlink {
		return fmt.Errorf("symlink is a symlink, dir path: %s", dirPath)
	}
	if _, err := os.ReadDir(dirPath); err != nil {
		return fmt.Errorf("failed to read directory %s: %v", dirPath, err)
	}
	tmpFile, err := os.CreateTemp(dirPath, "sqlite_check_")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v, dir: %s", err, dirPath)
	}
	defer tmpFile.Close()
	if err := tmpFile.Chmod(constants.DefaultFilePermission); err != nil {
		if err := os.Remove(tmpFile.Name()); err != nil {
			return fmt.Errorf("failed to remove temp file: %v, dir: %s", err, dirPath)
		}
		return fmt.Errorf("failed to chmod temp file: %v, dir: %s", err, dirPath)
	}
	if err := os.Remove(tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to remove temp file: %v, dir: %s", err, dirPath)
	}
	return nil
}

func checkFileRedWriteAble(inputFilePath string, redAble bool, writeAble bool) error {
	if redAble && !isFileAble(inputFilePath, os.O_RDONLY) {
		return fmt.Errorf("the current user does not have the permission to read files: %s", inputFilePath)
	}
	if writeAble && !isFileAble(inputFilePath, os.O_WRONLY|os.O_APPEND) {
		return fmt.Errorf("the current user does not have the permission to write files: %s", inputFilePath)
	}
	isSymlink, err := IsSymbolicLink(inputFilePath)
	if err != nil {
		return fmt.Errorf("failed to check symlink: %v, file path: %s", err, inputFilePath)
	}
	if isSymlink {
		return fmt.Errorf("symlink is a symlink, file path: %s", inputFilePath)
	}
	return nil
}

func isFileAble(path string, value int) bool {
	file, err := os.OpenFile(path, value, 0)
	if err != nil {
		return false
	}
	if err = file.Close(); err != nil {
		return false
	}
	return true
}

// FileValidator the file is validate or not, if not, return an error
func FileValidator(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("filePath: %s is a directory, not a file", filePath)
	}
	if fileInfo.Size() > constants.FileMaxSize {
		return fmt.Errorf("file size:(%d) reached the limitation(%d)", fileInfo.Size(), constants.FileMaxSize)
	}
	return nil
}
