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
	"ascend-common/api"
	"fmt"
	"os"
	"path/filepath"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

// WriteToFile write data to file
func WriteToFile(info, path string) error {
	return WriteToFileWithPerm(info, path, DefaultPerm, DefaultPerm)
}

// WriteToFileWithPerm write data to file with permission
func WriteToFileWithPerm(info, path string, dirPerm, filePerm os.FileMode) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("the path %s is not an absolute path", path)
	}
	dirPath := filepath.Dir(path)
	err := os.MkdirAll(dirPath, dirPerm)
	if err != nil {
		return err
	}
	hwlog.RunLog.Infof("start write info into file: %s", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePerm)
	if err != nil {
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			hwlog.RunLog.Errorf("close file failed, err: %v", err)
		}
	}()
	if _, err := utils.CheckPath(path); err != nil {
		return err
	}
	_, err = f.WriteString(info)
	return err
}

// RemoveResetFileAndDir remove file and dir
func RemoveResetFileAndDir(namespace, name string) error {
	file := GenResetFileName(namespace, name)
	rmErr := os.Remove(file)
	if rmErr != nil && !os.IsNotExist(rmErr) {
		return fmt.Errorf("failed to remove file(%s): %v", file, rmErr)
	}
	typeFile := GenResetTypeFileName(namespace, name)
	rmErr = os.Remove(typeFile)
	if rmErr != nil && !os.IsNotExist(rmErr) {
		return fmt.Errorf("failed to remove file(%s): %v", typeFile, rmErr)
	}
	resetDir := GenResetDirName(namespace, name)
	if rmErr = os.Remove(resetDir); rmErr != nil && !os.IsNotExist(rmErr) {
		return fmt.Errorf("failed to remove dir(%s): %v", typeFile, rmErr)
	}
	hwlog.RunLog.Infof("delete cm(%s) file(%s)", name, file)
	return nil
}

// RemoveDataTraceFileAndDir remove the job related data-trace config dir
func RemoveDataTraceFileAndDir(namespace, jobName string) error {
	dataTraceDirName := fmt.Sprintf("%s/%s", DataTraceConfigDir, namespace+"."+DataTraceCmPrefix+jobName)
	if !filepath.IsAbs(dataTraceDirName) {
		return fmt.Errorf("the path %s is not an absolute path", dataTraceDirName)
	}
	if _, err := utils.CheckPath(dataTraceDirName); err != nil {
		return fmt.Errorf("the path %s is invalid, err: %v", dataTraceDirName, err)
	}
	hwlog.RunLog.Infof("will delete data trace file: %s", dataTraceDirName)
	return os.RemoveAll(dataTraceDirName)
}

// RemoveSoftShareDeviceFileAndDir remove soft share device file and dir
func RemoveSoftShareDeviceFileAndDir(namespace, jobName string) error {
	softShareDeviceDirName := fmt.Sprintf("%s%s", api.SoftShareDeviceConfigDir, namespace+"."+jobName)
	if !filepath.IsAbs(softShareDeviceDirName) {
		return fmt.Errorf("the path %s is not an absolute path", softShareDeviceDirName)
	}
	if _, err := utils.CheckPath(softShareDeviceDirName); err != nil {
		return fmt.Errorf("the path %s is invalid, err: %v", softShareDeviceDirName, err)
	}
	hwlog.RunLog.Infof("will delete share device file: %s", softShareDeviceDirName)
	return os.RemoveAll(softShareDeviceDirName)
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
