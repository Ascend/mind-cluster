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
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/db"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

// QueryPpDevDuration 查找pp device侧耗时
func QueryPpDevDuration(dbCtx *db.SnpDbContext, groupName int64, startNsMin int64,
	startNsMax int64) (*model.Duration, error) {
	querySql := `
		SELECT
			endNs - startNs AS duration 
		FROM
			COMMUNICATION_OP
			LEFT JOIN STRING_IDS ON COMMUNICATION_OP.opType = STRING_IDS.id 
		WHERE
			groupName = ? 
			AND startNs >= ?
			AND startNs <= ? 
			AND lower( value ) = 'hcclsend'
		ORDER BY
			startNs 
			LIMIT 1
		`
	return db.QuerySingleLine(dbCtx, querySql, []any{groupName, startNsMin, startNsMax},
		model.DurationMapping)
}

// QueryPpHostDuration 查找pp host侧耗时
func QueryPpHostDuration(dbCtx *db.SnpDbContext, groupName int64, startNsMin int64,
	startNsMax int64) (*model.Duration, error) {
	querySql := `
		SELECT
			CAST(ROUND(COALESCE(AVG(endNs - startNs), 0)) AS INTEGER) AS duration
		FROM
			CANN_API
		INNER JOIN (
			SELECT
				connectionId 
			FROM
				COMMUNICATION_OP
				LEFT JOIN STRING_IDS ON COMMUNICATION_OP.opType = STRING_IDS.id 
			WHERE
				groupName = ? 
				AND startNs >= ?
				AND startNs <= ? 
				AND lower( value ) = 'hcclsend'
			) B ON CANN_API.connectionId = B.connectionId
		`
	return db.QuerySingleLine(dbCtx, querySql, []any{groupName, startNsMin, startNsMax},
		model.DurationMapping)
}
