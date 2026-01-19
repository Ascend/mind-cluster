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
Package dealdb.
*/

package dealdb

import (
	"strconv"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/db"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

// QueryHostStartEndTime 筛出该step的所有dataloader的host侧平均耗时。不做出现次数限制。若没有dataloader则就为0
func QueryHostStartEndTime(dbCtx *db.SnpDbContext, stepId int64) (*model.StartEndNs, error) {
	querySql := `
		SELECT
			startNs,
			endNs 
		FROM
			MSTX_EVENTS
			LEFT JOIN STRING_IDS AS MSG_IDS ON MSTX_EVENTS.message == MSG_IDS.id 
		WHERE
			lower( MSG_IDS.value ) like ('%step ' || ? || '%')
		`
	return db.QuerySingleLine(dbCtx, querySql, []any{strconv.FormatInt(stepId, constants.DecimalMark)},
		model.StartEndNsMapping)
}

// QueryDataLoaderHost 筛出该step的所有dataloader的host侧平均耗时。不做出现次数限制。若没有dataloader则就为0
func QueryDataLoaderHost(dbCtx *db.SnpDbContext, startNsMin int64,
	startNsMax int64) (*model.Duration, error) {
	querySql := `
		SELECT
			CAST(ROUND(COALESCE(AVG(endNs - startNs), 0)) AS INTEGER) AS duration 
		FROM
			MSTX_EVENTS
			LEFT JOIN STRING_IDS AS MSG_IDS ON MSTX_EVENTS.message == MSG_IDS.id 
		WHERE
			MSG_IDS.value = ? 
			AND startNs >= ?
			AND startNs <= ?
		`
	return db.QuerySingleLine(dbCtx, querySql, []any{constants.DataLoaderWord, startNsMin, startNsMax},
		model.DurationMapping)
}
