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
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/db"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

// QueryStepReduceTypeCnt 一张卡一个step的TP域内对比ReduceScatter与AllReduce的数量多的ID和count
func QueryStepReduceTypeCnt(dbCtx *db.SnpDbContext, groupName int64, startNsMin int64,
	startNsMax int64) (*model.IdView, error) {
	querySql := `SELECT id FROM (
		SELECT
			id,
			COUNT( 0 ) cnt
		FROM
			COMMUNICATION_OP
			LEFT JOIN STRING_IDS ON COMMUNICATION_OP.opType = STRING_IDS.id 
		WHERE
			groupName = ? 
			AND startNs >= ?
			AND startNs <= ? 
			AND ( lower( value ) = 'hcclallreduce' OR lower( value ) = 'hcclreducescatter' ) 
		GROUP BY
			id
		ORDER BY
			cnt DESC 
			LIMIT 1)
		`
	return db.QuerySingleLine(dbCtx, querySql, []any{groupName, startNsMin, startNsMax},
		model.IdViewMapping)
}

// QueryZpDurationWhenTpEnable 当Tp开启时，查找zp对应的耗时
func QueryZpDurationWhenTpEnable(dbCtx *db.SnpDbContext, params []any) (*model.HostDeviceDuration, error) {
	querySql := `
		SELECT
			CAST(ROUND(COALESCE(AVG(endNs - startNs), 0)) AS INTEGER) AS host_duration,
			CAST(ROUND(COALESCE(AVG(device_dur), 0)) AS INTEGER) AS device_dur
		FROM
			CANN_API
		INNER JOIN (
			SELECT
			 	endNs - startNs as device_dur,
				connectionId 
			FROM
				COMMUNICATION_OP
			WHERE
				groupName = ? 
				AND startNs >= ?
				AND startNs <= ? 
				AND opType = ?   ----？用筛选算子的结果里的id信息
				AND count>?
			ORDER BY
				startNs 
				LIMIT 1, 3 
			) B ON CANN_API.connectionId = B.connectionId
		`
	return db.QuerySingleLine(dbCtx, querySql, params, model.HdDurMapping)
}

// QueryZpDurationWhenTpDisable 当Tp关闭时，查找zp对应的耗时
func QueryZpDurationWhenTpDisable(dbCtx *db.SnpDbContext, startNsMin int64,
	startNsMax int64) (*model.HostDeviceDuration, error) {
	querySql := `
		SELECT
			CAST(ROUND(COALESCE(AVG(host_duration), 0)) AS INTEGER) AS host_duration,
			CAST(ROUND(COALESCE(AVG(device_duration), 0)) AS INTEGER) AS device_duration 
		FROM
			(
			SELECT
				MSTX_EVENTS.endNs - MSTX_EVENTS.startNs AS host_duration,
				TASK.endNs - TASK.startNs AS device_duration 
			FROM
				MSTX_EVENTS
				LEFT JOIN STRING_IDS AS MSG_IDS ON MSTX_EVENTS.message = MSG_IDS.id
				LEFT JOIN TASK ON TASK.connectionId = MSTX_EVENTS.connectionId 
			WHERE
				MSG_IDS.value = ? 
				AND TASK.startNs >= ?
				AND TASK.startNs <= ? 
			ORDER BY
				TASK.startNs 
			)
		`
	return db.QuerySingleLine(dbCtx, querySql, []any{constants.ForwardWord, startNsMin, startNsMax},
		model.HdDurMapping)
}
