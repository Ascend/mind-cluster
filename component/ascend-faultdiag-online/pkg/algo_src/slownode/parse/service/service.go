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

// Package service provides some funcs processing slow node
package service

import (
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/business/dealdb"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/business/writecsv"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context/contextdata"
)

// ParseData 清洗慢节点数据
func ParseData(ctxData *contextdata.SnpRankContextData, perStepId int64) error {
	// 删除历史数据
	if err := dealdb.DeleteTableDataBeforeStep(ctxData.DbCtx, perStepId); err != nil {
		return err
	}

	// 增量读，获取step时间
	startEndNsList, err := dealdb.QueryAllStepTime(ctxData.DbCtx)
	if err != nil {
		return err
	}
	if len(startEndNsList) == 0 {
		return nil
	}
	ctxData.RedStep = startEndNsList[len(startEndNsList)-1].Id

	// 处理通信算子数据 --> comm.csv
	globalRank, err := CollectGlobalRank(ctxData, startEndNsList)
	if err != nil {
		return err
	}
	if err = writecsv.WriteGlobalRank(globalRank, ctxData.CsvCtx.GlobalRankCsvFileHandler); err != nil {
		return err
	}

	// 处理迭代时延 --> steptime.csv
	iterateDelay, err := CollectIterateDelay(startEndNsList)
	if err != nil {
		return err
	}
	if err = writecsv.WriteIterateDelay(iterateDelay, ctxData.CsvCtx.StepTimeCsvFileHandler); err != nil {
		return err
	}
	ctxData.StepCount += int64(len(startEndNsList))
	return nil
}
