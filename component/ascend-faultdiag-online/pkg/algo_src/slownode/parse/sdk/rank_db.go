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
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/enum"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context/contextdata"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/db"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils/csvtool"
)

const dataQueMaxSize = 2000
const sqlQueMaxSize = 1000

func initSnpCtx(rankDir string, traffic int64, jobId string) *context.SnpRankContext {
	parserConfig := &config.SlowNodeParserConfig{
		RankDir:                    rankDir,
		DbFilePath:                 filepath.Join(rankDir, constants.DbFileName),
		GlobalRankCsvFilePath:      filepath.Join(rankDir, constants.GlobalRankCsvFileName),
		StepTimeCsvFilePath:        filepath.Join(rankDir, constants.StepTimeCsvFileName),
		ParGroupJsonInputFilePath:  filepath.Join(rankDir, constants.ParGroupJsonFileName),
		ParGroupJsonOutputFilePath: filepath.Join(rankDir, constants.ParGroupJsonFileName),
		Traffic:                    traffic,
	}
	snpContext := &context.SnpRankContext{
		ContextData: &contextdata.SnpRankContextData{
			DbCtx:  db.NewSqliteDbCtx(parserConfig.DbFilePath),
			Config: parserConfig,
		},
		JsonDataQue:  make(chan *model.JsonData, dataQueMaxSize),
		InsertSqlQue: make(chan string, sqlQueMaxSize),
		JobId:        jobId,
		RankId:       filepath.Base(rankDir),
	}
	return snpContext
}

func initRankDb(snpContext *context.SnpRankContext) error {
	// 首次执行，创建数据库db文件
	if err := snpContext.ContextData.DbCtx.Conn(); err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	if err := dealdb.CreateDBTable(snpContext.ContextData.DbCtx); err != nil {
		return fmt.Errorf("error creating database table: %v", err)
	}
	if err := dealdb.DeleteDbData(snpContext.ContextData.DbCtx); err != nil {
		return fmt.Errorf("error deleting database table: %v", err)
	}
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Init database succeeded. The rank dir is: %s",
		snpContext.ContextData.Config.RankDir)
	return nil
}

func initCsvFile(snpContext *context.SnpRankContext) error {
	globalRankCsvFilePath := snpContext.ContextData.Config.GlobalRankCsvFilePath
	stepTimeCsvFilePath := snpContext.ContextData.Config.StepTimeCsvFilePath
	if err := utils.RemoveAllFile([]string{globalRankCsvFilePath, stepTimeCsvFilePath}); err != nil {
		return err
	}
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Remove csv file succeeded. The rank dir is: %s",
		snpContext.ContextData.Config.RankDir)

	globalRankCsvFileHandler, err := csvtool.NewCSVHandler(globalRankCsvFilePath, enum.AppendMode,
		constants.DefaultFilePermission)
	if err != nil {
		return err
	}
	headers := []string{"step_index", "ZP_device", "ZP_host", "PP_device", "PP_host", "dataloader_host"}
	if err := globalRankCsvFileHandler.WriteRow(headers); err != nil {
		return err
	}
	if err := globalRankCsvFileHandler.Flush(); err != nil {
		return err
	}

	stepTimeCsvFileHandler, err := csvtool.NewCSVHandler(stepTimeCsvFilePath, enum.AppendMode,
		constants.DefaultFilePermission)
	if err != nil {
		return err
	}
	stepHeaders := []string{"step time", "durations"}
	if err := stepTimeCsvFileHandler.WriteRow(stepHeaders); err != nil {
		return err
	}
	if err := stepTimeCsvFileHandler.Flush(); err != nil {
		return err
	}

	snpContext.ContextData.CsvCtx = &contextdata.CsvCtx{
		GlobalRankCsvFileHandler: globalRankCsvFileHandler,
		StepTimeCsvFileHandler:   stepTimeCsvFileHandler,
	}
	return nil
}
