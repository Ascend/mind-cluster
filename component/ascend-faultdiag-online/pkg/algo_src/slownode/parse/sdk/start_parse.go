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
	"path/filepath"
	"sync"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/business/handlejson"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils"
)

var jobPipeline = &JobPipeline{
	jobMap: sync.Map{},
}

// StartParse 开始清洗
func StartParse(cg config.DataParseModel) {
	jobPipeline.StartJob(cg)
}

// StopParse 停止清洗
func StopParse(cg config.DataParseModel) {
	jobPipeline.StopJob(cg)
	hwlog.RunLog.Info("[SLOWNODE PARSE]Succeeded in stopping parsed")
}

// ReloadParse 重启清洗
func ReloadParse(cg config.DataParseModel) {
	jobPipeline.RestartJob(cg)
}

// DealParseJob 处理清洗Job
func DealParseJob(cg config.DataParseModel, parseJobInfo *model.ParseJobInfo) {
	if parseJobInfo == nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]invalid nil parseJobInfo")
		return
	}
	defer parseJobInfo.StopWg.Done()
	snpCtxSlice, err := listenAndParse(cg)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE PARSE]Failed to listen and parse data: %v", err)
		closeDbConnect(snpCtxSlice)
		clearJobData(snpCtxSlice, cg)
		return
	}

	DealParseCallback(snpCtxSlice, parseJobInfo)
	DealRankParGroup(snpCtxSlice, parseJobInfo)

	// 阻塞等待job任务中所有协程执行完毕
	parseJobInfo.JobWg.Wait()
	// 清除任务数据
	clearJobData(snpCtxSlice, cg)
}

func clearJobData(snpCtxSlice []*context.SnpRankContext, cg config.DataParseModel) {
	removeParseFile(snpCtxSlice, filepath.Join(cg.FilePath, cg.JobId))
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Succeeded in removing parse file, job id is: %s", cg.JobId)
	jobPipeline.jobMap.Delete(cg.JobId)
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Succeeded in parsing job and exited, job name is: %s, job id is: %s",
		cg.JobName, cg.JobId)
}

func removeParseFile(snpCtxSlice []*context.SnpRankContext, jobDir string) {
	for _, snpCtx := range snpCtxSlice {
		dbFilePath := snpCtx.ContextData.Config.DbFilePath
		dbShmFile := filepath.Join(snpCtx.ContextData.Config.RankDir, "database.db-shm")
		dbWalFile := filepath.Join(snpCtx.ContextData.Config.RankDir, "database.db-wal")
		stepTimeCsvFilePath := snpCtx.ContextData.Config.StepTimeCsvFilePath
		commCsvFilePath := snpCtx.ContextData.Config.GlobalRankCsvFilePath
		removeFiles := []string{dbFilePath, dbShmFile, dbWalFile, stepTimeCsvFilePath, commCsvFilePath}
		if err := utils.RemoveAllFile(removeFiles); err != nil {
			hwlog.RunLog.Warnf("[SLOWNODE PARSE]Failed to remove file: %v", err)
		}
	}

	parGroupJson := filepath.Join(jobDir, constants.ParGroupJsonFileName)
	if err := utils.RemoveAllFile([]string{parGroupJson}); err != nil {
		hwlog.RunLog.Warnf("[SLOWNODE PARSE]Failed to remove file: %v", err)
	}
}

func closeDbConnect(snpCtxSlice []*context.SnpRankContext) {
	for _, snpCtx := range snpCtxSlice {
		if snpCtx.ContextData.DbCtx == nil {
			continue
		}
		if err := snpCtx.ContextData.DbCtx.Close(); err != nil {
			hwlog.RunLog.Errorf("[SLOWNODE PARSE]Failed to close database connection: %v", err)
		}
	}
}

// DealRankParGroup 处理节点所有卡并行域信息
func DealRankParGroup(snpCtxSlice []*context.SnpRankContext, parseJobInfo *model.ParseJobInfo) {
	if parseJobInfo == nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]invalid nil parseJobInfo")
		return
	}
	mergeRankParGroupFunc := func() (bool, error) {
		// 回调前合并卡级别并行域
		return mergeRankParGroup(snpCtxSlice), nil
	}
	parseJobInfo.JobWg.Add(1)
	go func(jobId string) {
		defer parseJobInfo.JobWg.Done()
		err := utils.Poller(mergeRankParGroupFunc, constants.ParGroupTime, 0, parseJobInfo.StopParseFlag)
		if err != nil {
			hwlog.RunLog.Errorf("[SLOWNODE PARSE]Failed to merge parallel group info: %v", err)
		} else {
			hwlog.RunLog.Info("[SLOWNODE PARSE]Succeeded in merging parallel group info and exited")
		}
	}(parseJobInfo.JobId)
}

func mergeRankParGroup(snpCtxSlice []*context.SnpRankContext) bool {
	var parallelGroupFiles []string
	var jobDirPath string
	for _, snpCtx := range snpCtxSlice {
		if snpCtx == nil || snpCtx.ContextData == nil || snpCtx.ContextData.Config == nil {
			continue
		}
		rankDir := snpCtx.ContextData.Config.RankDir
		jobDirPath = filepath.Dir(rankDir)
		parallelGroupFiles = append(parallelGroupFiles, filepath.Join(rankDir, constants.ParGroupJsonFileName))
	}
	mergeInput := &model.MergeParallelGroupInfoInput{
		FilePaths:      parallelGroupFiles,
		FileSavePath:   filepath.Join(jobDirPath, constants.ParGroupJsonFileName),
		DeleteFileFlag: false,
	}
	if err := handlejson.ProcessMerge(mergeInput); err != nil {
		hwlog.RunLog.Warnf("[SLOWNODE PARSE]Failed to merge info within the node: %v", err)
		return false
	}

	return true
}
