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

// Package profiling_service contains functions that support dynamically collecting profiling data
package profiling_service

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
)

// GlobalRankId is the global rank id pass in by python api
var GlobalRankId int

// diskUsageUpperlimitMB is the  ProfilingBaseDir total upper limit containing all  jobs
var diskUsageUpperLimitMB = constant.DefaultDiskUpperLimitInMB

// SetDiskUsageUpperLimitMB is the ProfilingBaseDir total upper limit containing all jobs
func SetDiskUsageUpperLimitMB(upperLimitInMB int) {
	diskUsageUpperLimitMB = upperLimitInMB
}

// SaveProfilingDataIntoFile save current profiling data to file
func SaveProfilingDataIntoFile(rank int) error {
	if len(ProfilingRecordsMark) == 0 && len(ProfilingRecordsApi) == 0 && len(ProfilingRecordsKernel) == 0 {
		hwlog.RunLog.Debugf("ProfilingRecords is all empty, will do nothing")
		return nil
	}

	savePath, err := getCurrentSavePath(rank)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get profiling saving path, err: %s", err.Error())
		return err
	}
	newestFileName, err := getNewestFileName(savePath)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get newest file to write profiling data,err: %s", err.Error())
		return fmt.Errorf("failed to get newest file to write profiling data,err: %s", err.Error())
	}
	fileName := path.Join(savePath, newestFileName)
	if _, err := os.Stat(fileName); err == nil {
		if err := os.Rename(fileName, fileName+".tmp"); err != nil {
			hwlog.RunLog.Errorf("failed to rename existing file %s, err:%s", fileName, fileName+".tmp")
		}
		hwlog.RunLog.Infof("file %s has been rename to %s", fileName, fileName+".tmp")
	}
	fileNameTmp := path.Join(savePath, newestFileName+".tmp")
	hwlog.RunLog.Debugf("rank:%v,the save fileName is %s", GlobalRankId, fileNameTmp)
	file, err := os.OpenFile(fileNameTmp, os.O_WRONLY|os.O_APPEND|os.O_CREATE, constant.ProfilingFileMode)
	if err != nil {
		hwlog.RunLog.Errorf("failed to open save file, err:%s", err)
		return err
	}
	defer file.Close()
	hwlog.RunLog.Debugf("rank:%v,start to write profiling data", GlobalRankId)
	if err = saveProfileFile(file); err != nil {
		return err
	}
	hwlog.RunLog.Debugf("finished write profiling file for rank:%d", rank)
	return nil
}

func getNewestFileName(filePath string) (string, error) {
	entries, err := os.ReadDir(filePath)
	if err != nil {
		return "", err
	}
	var latestTime int64 = -1
	var latestFileName string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		// convert filename to timestamp
		timeStr := fileName
		timestamp, err := strconv.ParseInt(timeStr, constant.TenBase, constant.BitSize64)
		if err != nil {
			continue
		}
		if timestamp > latestTime {
			latestTime = timestamp
			latestFileName = fileName
		}
	}
	if latestFileName == "" {
		return fmt.Sprintf("%d", time.Now().Unix()), nil
	}
	overSized, err := isFileOver10MB(path.Join(filePath, latestFileName))
	if err != nil {
		return "", err
	}
	if !overSized {
		return latestFileName, nil
	}
	return fmt.Sprintf("%d", time.Now().Unix()), nil
}

func isFileOver10MB(filePath string) (bool, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}
	size := fileInfo.Size()
	return size > constant.SizeLimitPerProfilingFile, nil
}

func saveProfileFile(file *os.File) error {
	// .tmp means that file is not done writing yet
	// 将 profilingRecords 序列化为 JSON
	hwlog.RunLog.Debugf("rank:%v,will stat to save file", GlobalRankId)
	recordsBytes := writeToBytes()
	hwlog.RunLog.Debugf("rank:%v, finished to unmarsh marker to string at:%v", GlobalRankId, time.Now())

	if err := writeLongStringToFileWithBuffer(file, recordsBytes); err != nil {
		hwlog.RunLog.Errorf("Error writing to file: %s", err.Error())
		return err
	}
	hwlog.RunLog.Debugf("Data successfully written to %s, will try to rename file to indicate wrote", file.Name())
	err := os.Rename(file.Name(), strings.TrimRight(file.Name(), ".tmp"))
	if err != nil {
		hwlog.RunLog.Errorf("Error renaming file:%s", err.Error())
		return err
	}
	return nil
}

// writeLongStringToFileWithBuffer writes a string into file with a timeout
func writeLongStringToFileWithBuffer(file *os.File, bytes []byte) error {
	writer := bufio.NewWriter(file)
	count, err := writer.Write(bytes)
	if err != nil {
		hwlog.RunLog.Errorf("Error writing to buffer: %s", err.Error())
		return err
	}

	if count < len(bytes) {
		hwlog.RunLog.Warn("the writer count is less than len bytes")
	}

	if err := writer.Flush(); err != nil {
		hwlog.RunLog.Errorf("Error flushing buffer to file: %s", err.Error())
		return err
	}
	return nil
}

func writeToBytes() []byte {
	var buffer bytes.Buffer
	estimatedSize := (len(ProfilingRecordsMark) + len(ProfilingRecordsApi) +
		len(ProfilingRecordsKernel)) * constant.DefaultRecordLength
	if estimatedSize < constant.LeastBufferSize { // 最小预分配4KB
		estimatedSize = constant.LeastBufferSize
	}
	buffer.Grow(estimatedSize)
	constant.MuMark.Lock()
	defer constant.MuMark.Unlock()
	for _, v := range ProfilingRecordsMark {
		buffer.Write(v.Marshal())
		buffer.WriteByte(constant.LineSeperator)
	}
	ProfilingRecordsMark = make([]MsptiActivityMark, 0)

	constant.MuApi.Lock()
	defer constant.MuApi.Unlock()
	for _, v := range ProfilingRecordsApi {
		buffer.Write(v.Marshal())
		buffer.WriteByte(constant.LineSeperator)
	}
	ProfilingRecordsApi = make([]MsptiActivityApi, 0)

	constant.MuKernal.Lock()
	defer constant.MuKernal.Unlock()
	for _, v := range ProfilingRecordsKernel {
		buffer.Write(v.Marshal())
		buffer.WriteByte(constant.LineSeperator)
	}
	ProfilingRecordsKernel = make([]MsptiActivityKernel, 0)
	return buffer.Bytes()
}

func getCurrentSavePath(rank int) (string, error) {
	// rank should be pass in by user python script, for example dist.get_rank in pytorch
	uid := os.Getenv(constant.TaskUidKey)
	// fault tolerance，if pg id not found, use default_task_id
	if uid == "" {
		uid = "default_task_id_" + strconv.Itoa(int(time.Now().Unix()))
	}
	rankPath := path.Join(constant.ProfilingBaseDir, uid, strconv.Itoa(rank))
	if len(rankPath) > constant.PathLengthLimit {
		return "", errors.New("path is too long, will not create it")
	}
	// non-sensitive data
	if err := os.MkdirAll(rankPath, constant.ProfilingDirMode); err != nil {
		hwlog.RunLog.Error(err)
		return "", err
	}
	return rankPath, nil
}

// getDirSizeInMB will return specific dir size in MB
func getDirSizeInMB(path string) (float64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	// convert bytes to MB
	sizeMB := float64(size) / constant.BytesPerMB
	return sizeMB, nil
}

// getProfileFiles get all profiling files
func getProfileFiles(jobDir string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(jobDir + "/" + strconv.Itoa(GlobalRankId))
	if err != nil {
		return nil, err
	}
	// filter out files
	var profileFile []os.FileInfo
	for _, file := range files {
		if !file.IsDir() {
			profileFile = append(profileFile, file)
		}
	}
	// sort by mod time
	sort.Slice(profileFile, func(i, j int) bool {
		return profileFile[i].ModTime().Before(profileFile[j].ModTime())
	})
	return profileFile, nil
}

// deleteOldestFileForEachRank delete the oldest file of each job by ModTime
func deleteOldestFileForEachRank(jobDir string) error {
	// get all the profiling files of one job, and return its files by modtime
	profileFiles, err := getProfileFiles(jobDir)
	if err != nil {
		return err
	}
	if len(profileFiles) == 0 {
		hwlog.RunLog.Infof("No profiling files found in %s\n", jobDir)
		return nil
	}
	// deleting the oldest profile file
	// each time will delete 20% of the profile files
	for i := 0; i < len(profileFiles)/constant.NumberOfParts && len(profileFiles) > constant.MinProfilingFileNum; i++ {
		oldestFilePath := filepath.Join(jobDir, strconv.Itoa(GlobalRankId),
			profileFiles[i].Name())
		if _, err := os.Stat(oldestFilePath); os.IsNotExist(err) {
			hwlog.RunLog.Errorf("file %s dose not exist", oldestFilePath)
			continue
		}
		err := os.Remove(oldestFilePath)
		if err != nil {
			hwlog.RunLog.Errorf("failed to delete file %s, err: %v", oldestFilePath, err)
		}
	}
	return nil
}

// ManageSaveProfiling is the main function for manage saving profiling files
func ManageSaveProfiling(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			hwlog.RunLog.Errorf("manager of saving all profiling files has paniced, err: %v", r)
			fmt.Printf("[ERROR] %s manager of saving all profiling files has paniced, err: %v\n", time.Now(), r)
		}
	}()
	hwlog.RunLog.Info("start to watch for ManageSaveProfiling")
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warnf("manage profiling disk usage received exit signal")
			return
		default:
			if !needSave() {
				hwlog.RunLog.Debugf("rank:%v, no need to save profiling to disk", GlobalRankId)
				time.Sleep(constant.CheckProfilingCacheInterval)
				continue
			}
			if err := SaveProfilingDataIntoFile(GlobalRankId); err != nil {
				hwlog.RunLog.Errorf("failed to save profiling, error: %v", err)
			}
			time.Sleep(constant.CheckProfilingCacheInterval)
		}
	}
}

func needSave() bool {
	if len(ProfilingRecordsMark)+len(ProfilingRecordsApi)+len(ProfilingRecordsKernel) > constant.MaxCacheRecords {
		return true
	} else {
		hwlog.RunLog.Infof("will flush all profiling records")
		if err := FlushAllActivity(); err != nil {
			hwlog.RunLog.Errorf("failed to flush profiling data,err: %s", err.Error())
			return false
		}
	}
	return true
}

// ManageProfilingDiskUsage when the dir usage is more than guage, will delete the oldest file
func ManageProfilingDiskUsage(baseDir string, ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			hwlog.RunLog.Errorf("manager of disk usage has paniced, err: %v", r)
			fmt.Printf("[ERROR] %s manager of disk usage has paniced, err: %v\n", time.Now(), r)
		}
	}()
	hwlog.RunLog.Infof("start to watch for ManageProfilingDiskUsage")
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warnf("ManageProfilingDiskUsage received exit signal")
			return
		default:
			// "/user/cluster-info/profiling"
			usedSize, err := getDirSizeInMB(baseDir)
			if err != nil {
				hwlog.RunLog.Errorf("failed to get dir[%s] disk size, err:%s", baseDir, err.Error())
				time.Sleep(constant.DiskUsageCheckInterval * time.Second)
				continue
			}
			if int(usedSize) > diskUsageUpperLimitMB {
				dealWithDiskUsage(baseDir, usedSize)
			} else {
				hwlog.RunLog.Debugf("rank:%d, disk usage is under threshold, no deletion necessary", GlobalRankId)
			}
			time.Sleep(constant.DiskUsageCheckInterval * time.Second)
		}
	}
}

func dealWithDiskUsage(baseDir string, usedSize float64) {
	hwlog.RunLog.Infof("path %s has used %d MB", baseDir, int(usedSize))
	// walk all job-uid dir
	jobDirs, err := ioutil.ReadDir(baseDir)
	if err != nil {
		hwlog.RunLog.Errorf("failed to read base directory: %s", err.Error())
		return
	}
	for _, jobDir := range jobDirs {
		if jobDir.IsDir() {
			jobDirPath := filepath.Join(baseDir, jobDir.Name())
			err := deleteOldestFileForEachRank(jobDirPath)
			if err != nil {
				hwlog.RunLog.Errorf("Failed to delete oldest step in %s: %s", jobDirPath, err.Error())
			}
		}
	}
}
