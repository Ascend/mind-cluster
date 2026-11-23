/* Copyright 2025 Huawei Technologies Co., Ltd.

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

// Package writecsv provides some funcs relevant to csv
package writecsv

import (
	"errors"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils/csvtool"
)

// WriteGlobalRank 通信算子信息写入csv
func WriteGlobalRank(globalRanks []*model.StepGlobalRank, fileHandler *csvtool.CSVHandler) error {
	if fileHandler == nil {
		return errors.New("invalid nil fileHandler")
	}
	var rows [][]string
	for _, rank := range globalRanks {
		if rank == nil {
			continue
		}
		srcRow := []any{rank.StepIndex, rank.ZPDevice, rank.ZPHost, rank.PPDevice, rank.PPHost,
			rank.DataLoaderHost}
		rows = append(rows, utils.ToStringList(srcRow))
	}

	if err := fileHandler.WriteAll(rows); err != nil {
		return err
	}

	return fileHandler.Flush()
}

// WriteIterateDelay 迭代时延信息写入csv
func WriteIterateDelay(iterateDelay []*model.StepIterateDelay, fileHandler *csvtool.CSVHandler) error {
	if fileHandler == nil {
		return errors.New("invalid nil fileHandler")
	}
	var rows [][]string
	for _, rank := range iterateDelay {
		if rank == nil {
			continue
		}
		srcRow := []any{rank.StepTime, rank.Durations}
		rows = append(rows, utils.ToStringList(srcRow))
	}

	if err := fileHandler.WriteAll(rows); err != nil {
		return err
	}

	return fileHandler.Flush()
}
