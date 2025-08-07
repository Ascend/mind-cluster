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
)

// DeleteDbData 删除数据库数据
func DeleteDbData(dbCtx *db.SnpDbContext) error {
	tableNames := []string{"CANN_API", "COMMUNICATION_OP", "MSTX_EVENTS", "STEP_TIME", "TASK"}
	for _, name := range tableNames {
		if err := deleteTable(dbCtx, name); err != nil {
			return err
		}
	}
	return nil
}

// deleteTable 删除数据库表所有数据
func deleteTable(dbCtx *db.SnpDbContext, tableName string) error {
	return db.Delete(dbCtx, fmt.Sprintf("DELETE FROM %s;", tableName), []any{})
}

// DeleteTableDataBeforeStep 删除当前step前的所有数据
func DeleteTableDataBeforeStep(dbCtx *db.SnpDbContext, stepId int64) error {
	sql := `DELETE FROM %s WHERE startNs <= (SELECT startNs FROM STEP_TIME WHERE id = ?)`
	var opTables = []string{"CANN_API", "COMMUNICATION_OP", "MSTX_EVENTS", "TASK"}
	for _, table := range opTables {
		err := db.Delete(dbCtx, fmt.Sprintf(sql, table), []any{stepId})
		if err != nil {
			return err
		}
	}
	sql = `DELETE FROM STEP_TIME WHERE id <= ?`
	err := db.Delete(dbCtx, sql, []any{stepId})
	return err
}
