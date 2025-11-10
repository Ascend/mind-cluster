/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package fileutils provides some utils to operate the files
*/
package fileutils

import (
	"errors"
	"path/filepath"
	"strings"

	"ascend-common/common-utils/utils"
)

// CheckPath checks the path for path traversal and returns the absolute path
func CheckPath(path string) (string, error) {
	if containsPathTraversal(path) {
		return "", errors.New("path traversal detected")
	}

	absPath, err := utils.CheckPath(path)
	if err != nil {
		return absPath, err
	}

	return absPath, nil
}

func containsPathTraversal(path string) bool {
	cleanPath := filepath.Clean(path)
	// check the path has path traversal after cleaning
	if strings.Contains(cleanPath, "..") {
		return true
	}
	if strings.Contains(path, "..") || strings.Contains(path, "//") {
		return true
	}

	return false
}

// ReadLimitBytes check the path and read the content by giving limitation
func ReadLimitBytes(path string, limitLength int) ([]byte, error) {
	if containsPathTraversal(path) {
		return nil, errors.New("path traversal detected")
	}
	// utils.ReadLimitBytes will CheckPath again
	return utils.ReadLimitBytes(path, limitLength)
}
