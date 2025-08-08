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
Package contextdata.
*/
package contextdata

import (
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/db"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils/csvtool"
)

// SnpRankContextData 上下文数据
type SnpRankContextData struct {
	// DbCtx 数据库上下文
	DbCtx *db.SnpDbContext
	// Config 文件路径配置
	Config *config.SlowNodeParserConfig
	// CsvCtx csv文件处理上下文
	CsvCtx *CsvCtx
	// StepCount 步数计数
	StepCount int64
	// RedStep 读取步数
	RedStep int64
}

// CsvCtx csv文件内容
type CsvCtx struct {
	// GlobalRankCsvFileHandler comm.csv文件处理
	GlobalRankCsvFileHandler *csvtool.CSVHandler
	// StepTimeCsvFileHandler steptime.csv文件处理
	StepTimeCsvFileHandler *csvtool.CSVHandler
}

// Close 关闭文件handler
func (csvCtx *CsvCtx) Close() []error {
	errorSlice := make([]error, 0, 2)
	if csvCtx.GlobalRankCsvFileHandler != nil {
		if err := csvCtx.GlobalRankCsvFileHandler.Close(); err != nil {
			errorSlice = append(errorSlice, err)
		}
	}
	if csvCtx.StepTimeCsvFileHandler != nil {
		if err := csvCtx.StepTimeCsvFileHandler.Close(); err != nil {
			errorSlice = append(errorSlice, err)
		}
	}
	return errorSlice
}
