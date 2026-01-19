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
	"fmt"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/db"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

// InsertStringIds STING_IDS表中插入数据，并返回自增主键
func InsertStringIds(dbCtx *db.SnpDbContext, value string) (*model.IdView, error) {
	dbCtx.DbWriteLock.Lock()
	defer dbCtx.DbWriteLock.Unlock()
	insertSQL := `INSERT INTO STRING_IDS (value) VALUES (?);`
	id, err := db.Insert(dbCtx, insertSQL, []any{value})
	if err != nil {
		return nil, fmt.Errorf("insert STRING_IDS error: %v", err)
	}
	return &model.IdView{Id: id}, nil
}

// InsertJsonData 批量插入数据
func InsertJsonData(dbCtx *db.SnpDbContext, insertStrSlice []string) error {
	dbCtx.DbWriteLock.Lock()
	defer dbCtx.DbWriteLock.Unlock()
	for _, sql := range insertStrSlice {
		if _, err := db.Insert(dbCtx, sql, []any{}); err != nil {
			return fmt.Errorf("insert json data error: %v, sql is: %s", err, sql)
		}
	}
	return nil
}
