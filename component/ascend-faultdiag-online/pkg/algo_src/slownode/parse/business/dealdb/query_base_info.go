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

// QueryMataData 查询元数据表
func QueryMataData(dbCtx *db.SnpDbContext, name string) (*model.ValueView, error) {
	querySql := `select value from META_DATA where name = ?`
	return db.QuerySingleLine(dbCtx, querySql, []any{name}, model.ValueViewMapping)
}

// QueryStringId 查询字符串id表
func QueryStringId(dbCtx *db.SnpDbContext, value string) (*model.IdView, error) {
	querySql := `SELECT id from "STRING_IDS" where value = ?`
	return db.QuerySingleLine(dbCtx, querySql, []any{value}, model.IdViewMapping)
}

// QueryAllStringIds 查找STING_IDS所有数据
func QueryAllStringIds(dbCtx *db.SnpDbContext) ([]*model.StringIdsView, error) {
	querySql := `SELECT id, value FROM STRING_IDS;`
	return db.Query(dbCtx, querySql, []any{}, model.StringIdsMapping)
}

// QueryAllStepTime 查询STEP_TIME表中所有的信息
func QueryAllStepTime(dbCtx *db.SnpDbContext) ([]*model.StepStartEndNs, error) {
	querySql := `select id, startNs, endNs from "STEP_TIME";`
	return db.Query(dbCtx, querySql, []any{}, model.StepStartEndNsMapping)
}
