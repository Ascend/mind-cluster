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

/*
Package handlejson is used to read parallel group information from JSON data
*/
package handlejson

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/enum"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils"
	coreModel "ascend-faultdiag-online/pkg/core/model"
	globalConstants "ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/fileutils"
)

var callbackFunc coreModel.CallbackFunc = nil

// MergeParallelGroupInfo 并行域信息汇总
func MergeParallelGroupInfo(cg config.DataParseModel) error {
	mergeInput := initParGroupInfo(cg)
	if err := ProcessMerge(mergeInput); err != nil {
		processParGroupResult(cg.JobName, cg.JobId, false)
		return err
	}
	processParGroupResult(cg.JobName, cg.JobId, true)
	return nil
}

func initParGroupInfo(cg config.DataParseModel) *model.MergeParallelGroupInfoInput {
	filePaths := make([]string, 0)
	jobDir := filepath.Join(cg.FilePath, cg.JobId)
	for _, fileName := range cg.ParallelGroupPath {
		fullFilePath := filepath.Join(jobDir, fileName)
		filePaths = append(filePaths, fullFilePath)
	}
	mergeInput := &model.MergeParallelGroupInfoInput{
		FilePaths:      filePaths,
		FileSavePath:   filepath.Join(jobDir, constants.ClusterParGroupFileName),
		DeleteFileFlag: true,
	}
	return mergeInput
}

// ProcessMerge 并行域合并
func ProcessMerge(mergeInput *model.MergeParallelGroupInfoInput) error {
	if mergeInput == nil {
		return errors.New("invalid nil mergeInput")
	}
	if err := checkParGroupInfo(mergeInput.FilePaths); err != nil {
		return err
	}
	groupInfo, err := processData(mergeInput)
	if err != nil {
		return err
	}

	if len(groupInfo) == 0 {
		return errors.New("no parallel group info")
	}
	// global_ranks按升序排列
	for _, info := range groupInfo {
		info.GlobalRanks = utils.UniqueSlice(info.GlobalRanks)
		sort.Slice(info.GlobalRanks, func(i, j int) bool {
			return info.GlobalRanks[i] < info.GlobalRanks[j]
		})
	}

	// 落盘到文件
	updatedData, err := json.MarshalIndent(groupInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to encode modified JSON: %v", err)
	}
	if err = os.WriteFile(mergeInput.FileSavePath, updatedData, constants.DefaultFilePermission); err != nil {
		return fmt.Errorf("unable to save file: %v", err)
	}
	return nil
}

func checkParGroupInfo(filePaths []string) error {
	if len(filePaths) == 0 {
		return errors.New("no parallel group file found")
	}
	for _, fullFilePath := range filePaths {
		if err := utils.CheckFilePerm(fullFilePath, true, false); err != nil {
			return err
		}
	}
	return nil
}

func processParGroupResult(jobName string, jobId string, IsFinished bool) {
	result := &model.MergeParallelGroupInfoResult{
		JobName:      jobName,
		JobId:        jobId,
		IsFinished:   IsFinished,
		FinishedTime: time.Now().UnixMilli(),
	}
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE PARSE]Parallel group info callback failed: %v", err)
		return
	}
	if callbackFunc != nil {
		go callbackFunc(string(jsonStr))
	}
}

func processData(mergeInput *model.MergeParallelGroupInfoInput) (map[string]*model.OpGroupInfo, error) {
	mergeInput.FileMu.Lock()
	defer mergeInput.FileMu.Unlock()
	const initialCapacity = 10
	groupInfo := make(map[string]*model.OpGroupInfo, initialCapacity)
	for _, fullFilePath := range mergeInput.FilePaths {
		var tempData map[string]*model.OpGroupInfo
		// 读取json数据
		readData, err := fileutils.ReadLimitBytes(fullFilePath, globalConstants.Size100M)
		if err != nil {
			return nil, fmt.Errorf("unable to read file %s: %v", fullFilePath, err)
		}
		// 文件系统的行为 + for-select 的竞态窗口问题，可能读到空内容。
		if len(readData) == 0 {
			return nil, fmt.Errorf("file %s is empty", fullFilePath)
		}
		if err := json.Unmarshal(readData, &tempData); err != nil {
			return nil, fmt.Errorf("the JSON file path is as follows: '%s', error unmarshalling JSON: %v",
				fullFilePath, err)
		}

		// 根据算子域合并信息
		for name, info := range tempData {
			if info.GroupName != string(enum.Pp) && info.GroupName != string(enum.Tp) {
				continue
			}
			data, exists := groupInfo[name]
			if !exists {
				groupInfo[name] = info
				continue
			}
			data.GlobalRanks = append(data.GlobalRanks, info.GlobalRanks...)
		}

		if !mergeInput.DeleteFileFlag {
			continue
		}
		// 读取后删除文件
		if err = os.Remove(fullFilePath); err != nil {
			return nil, fmt.Errorf("unable to delete file %s: %v", fullFilePath, err)
		}
	}

	return groupInfo, nil
}

// RegisterParGroupCallback register a function to get parallel group info
func RegisterParGroupCallback(callback coreModel.CallbackFunc) {
	callbackFunc = callback
}
