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

// Package sdk provides node parse
package sdk

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils"
)

func listenAndParse(cg config.DataParseModel) ([]*context.SnpRankContext, error) {
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Input rank ids are: %v", cg.RankIds)
	pollTicker := time.NewTicker(constants.PollTime)
	defer pollTicker.Stop()
	timeoutChan := time.After(constants.TimeoutFindFile)
	var tempStopChan = make(chan struct{})
	if stopChan := jobPipeline.GetStopChan(cg.JobId); stopChan != nil {
		tempStopChan = stopChan
	}
	snpCtxSlice := make([]*context.SnpRankContext, 0)
	checkRanks := cg.RankIds
	count := 0
	for {
		select {
		case _, ok := <-tempStopChan:
			if !ok {
				return snpCtxSlice, nil
			}
		case <-pollTicker.C:
			// 满足条件的rank，直接清洗
			meetRanks := checkRankPath(filepath.Join(cg.FilePath, cg.JobId), checkRanks)
			if len(meetRanks) > 0 {
				hwlog.RunLog.Infof("[SLOWNODE PARSE]Matched rank dir: %v", meetRanks)
			}
			eachRankCtxSlice, err := snpEachRank(cg, meetRanks)
			snpCtxSlice = append(snpCtxSlice, eachRankCtxSlice...)
			if err != nil {
				return snpCtxSlice, err
			}
			// 不满足条件的rank，继续轮训校验
			checkRanks = utils.SubtractAndDedupe(checkRanks, meetRanks)
			if len(checkRanks) == 0 {
				hwlog.RunLog.Infof("[SLOWNODE PARSE]Succeeded in checking rank dir: %v", cg.RankIds)
				return snpCtxSlice, nil
			}
			if len(checkRanks) < len(cg.RankIds) {
				count++
			}
			// 部分rank没有数据，检测10分钟后超时退出
			if count >= constants.LoopCount {
				hwlog.RunLog.Warnf("[SLOWNODE PARSE]Some rank dir do not contain original files: %v, timeout: %v",
					checkRanks, constants.PollTime*constants.LoopCount)
				return snpCtxSlice, nil
			}
		case <-timeoutChan:
			return snpCtxSlice, fmt.Errorf("searching for the rank dirs timed out: %v", constants.TimeoutFindFile)
		}
	}
}

func snpEachRank(cg config.DataParseModel, rankIds []string) ([]*context.SnpRankContext, error) {
	snpCtxSlice := make([]*context.SnpRankContext, 0)
	jobDir := filepath.Join(cg.FilePath, cg.JobId)
	for _, rankId := range rankIds {
		rankDirPath := filepath.Join(jobDir, rankId)
		hwlog.RunLog.Infof("[SLOWNODE PARSE]Start parsing data asynchronously: %s", rankDirPath)
		snpCtx, err := StartSnpParse(rankDirPath, cg.Traffic, cg.JobId)
		if err != nil {
			return snpCtxSlice, fmt.Errorf("parse rank %s data failed: %v", rankId, err)
		}
		snpCtxSlice = append(snpCtxSlice, snpCtx)
	}
	return snpCtxSlice, nil
}

func checkRankPath(jobDir string, rankIds []string) []string {
	meetRanks := make([]string, 0)
	// 检查job目录是否存在
	jobDirInfo, err := os.Stat(jobDir)
	if err != nil || !jobDirInfo.IsDir() {
		hwlog.RunLog.Warnf("[SLOWNODE PARSE]The job file does not exist or an error occurs: %v, job path is %s",
			err, jobDir)
		return meetRanks
	}
	// 检查rank目录是否存在
	if len(rankIds) == 0 || len(rankIds) > constants.MaxRankNum {
		hwlog.RunLog.Warnf("[SLOWNODE PARSE]Input rank ids is empty or more than %d", constants.MaxRankNum)
		return meetRanks
	}
	for _, rankId := range rankIds {
		rankDirPath := filepath.Join(jobDir, rankId)
		rankDirInfo, err := os.Stat(rankDirPath)
		if err != nil || !rankDirInfo.IsDir() {
			hwlog.RunLog.Warnf("[SLOWNODE PARSE]The rank file does not exist or an error occurs: %v, "+
				"rank path is: %s", err, rankDirPath)
			continue
		}
		// 检查是否存在目标文件
		if !checkParseFile(rankDirPath) {
			hwlog.RunLog.Warnf("[SLOWNODE PARSE]No parsing files in: %v", rankDirPath)
			continue
		}
		meetRanks = append(meetRanks, rankId)
	}

	return meetRanks
}

func checkParseFile(rankDirPath string) bool {
	entries, err := os.ReadDir(rankDirPath)
	if err != nil {
		hwlog.RunLog.Warnf("[SLOWNODE PARSE]Failed to read the file in checking: %v", rankDirPath)
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == constants.ParGroupJsonFileName {
			return true
		}
	}
	return false
}
