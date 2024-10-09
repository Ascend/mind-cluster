/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common a series of common function
package common

import (
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
)

const (
	defaultPerm = 0666
)

// WriteToFile write data to file
func WriteToFile(info, path string) error {
	dirPath := filepath.Dir(path)
	err := os.MkdirAll(dirPath, defaultPerm)
	if err != nil {
		return err
	}
	hwlog.RunLog.Infof("start write info into file: %s", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, defaultPerm)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(info)
	return err
}

// RemoveFileAndDir remove file and dir
func RemoveFileAndDir(namespace, name string) error {
	file := GenResetFileName(namespace, name)
	rmErr := os.Remove(file)
	if rmErr != nil {
		return fmt.Errorf("failed to remove file(%s): %v", file, rmErr)
	}
	typeFile := GenResetTypeFileName(namespace, name)
	rmErr = os.Remove(typeFile)
	if rmErr != nil {
		return fmt.Errorf("failed to remove file(%s): %v", typeFile, rmErr)
	}
	dir := GenResetDirName(namespace, name)
	err := os.Remove(dir)
	if err != nil {
		return fmt.Errorf("failed to remove dir(%s): %v", dir, err)
	}
	hwlog.RunLog.Infof("delete cm(%s) file(%s)", name, file)
	return nil
}

// GenResetDirName generate reset cm dir name
func GenResetDirName(namespace, name string) string {
	return ResetInfoDir + namespace + "." + name
}

// GenResetFileName generate reset cm file name
func GenResetFileName(namespace, name string) string {
	return GenResetDirName(namespace, name) + "/" + ResetInfoCMDataKey
}

// GenResetTypeFileName generate reset cm file name
func GenResetTypeFileName(namespace, name string) string {
	return GenResetDirName(namespace, name) + "/" + ResetInfoTypeKey
}
