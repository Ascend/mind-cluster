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
	"path/filepath"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/business/dealdb"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/business/handlejson"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/service"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils"
)

// StartSnpParse 开始异步清洗
func StartSnpParse(rankDir string, traffic int64, jobId string) (*context.SnpRankContext, error) {
	snpRankCtx := initSnpCtx(rankDir, traffic, jobId)
	if err := initRankDb(snpRankCtx); err != nil {
		return nil, err
	}
	if err := initCsvFile(snpRankCtx); err != nil {
		return nil, err
	}
	parseFileCtx := handlejson.NewParseFileContext()
	jobInfo, err := jobPipeline.GetJobInfo(jobId)
	if err != nil {
		return nil, err
	}
	handlejson.StartReadJson(snpRankCtx, parseFileCtx, jobInfo.StopParseFlag, jobInfo.TimeStamp)
	handlejson.StartParseJsonDataToSql(snpRankCtx, parseFileCtx, jobInfo.StopParseFlag)
	go dealdb.StartWriteSql(snpRankCtx, jobInfo.StopParseFlag)

	jobInfo.JobWg.Add(1)
	go startWriteCsvTicker(snpRankCtx, jobInfo)
	return snpRankCtx, nil
}

func startWriteCsvTicker(snpCtx *context.SnpRankContext, jobInfo *model.ParseJobInfo) {
	defer jobInfo.JobWg.Done()
	writeCsvFunc := func() (bool, error) {
		if err := writeEvent(snpCtx); err != nil {
			hwlog.RunLog.Error("[SLOWNODE PARSE]Failed to writing parse result:", err)
		}
		return false, nil
	}

	err := utils.Poller(writeCsvFunc, constants.PollTime, 0, jobInfo.StopParseFlag)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE PARSE]Failed to write parse result and exited: %v, rank dir is: %s",
			err, snpCtx.ContextData.Config.RankDir)
	} else {
		hwlog.RunLog.Infof("[SLOWNODE PARSE]Succeeded in writing parse result and exited, rank dir is: %s",
			snpCtx.ContextData.Config.RankDir)
	}

	if errSlice := snpCtx.ContextData.CsvCtx.Close(); len(errSlice) > 0 {
		hwlog.RunLog.Errorf("[SLOWNODE PARSE]Failed to close csv file: %v, rank dir is: %s",
			errSlice, snpCtx.ContextData.Config.RankDir)
	}

}

func writeEvent(snpCtx *context.SnpRankContext) error {
	err := utils.CheckFilePerm(snpCtx.ContextData.Config.ParGroupJsonInputFilePath, true, false)
	if err != nil {
		return err
	}

	if err := service.ParseData(snpCtx.ContextData, snpCtx.ContextData.RedStep); err != nil {
		return fmt.Errorf("failed to parse node data: %v", err)
	} else {
		rankNum := filepath.Base(snpCtx.ContextData.Config.RankDir)
		hwlog.RunLog.Infof("[SLOWNODE PARSE]The step for parsing data is: step %d, rank num is: %s",
			snpCtx.ContextData.RedStep, rankNum)
	}
	return nil
}
