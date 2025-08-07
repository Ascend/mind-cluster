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
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context"
)

// StartWriteSql 写入数据库表
func StartWriteSql(snpRankCtx *context.SnpRankContext, stopFlag chan struct{}) {
	defer snpRankCtx.ContextData.DbCtx.Close()
	for {
		select {
		case _, ok := <-stopFlag:
			if !ok {
				return
			}
		case sql, ok := <-snpRankCtx.InsertSqlQue:
			if !ok {
				return
			}
			err := InsertJsonData(snpRankCtx.ContextData.DbCtx, []string{sql})
			if err != nil {
				continue
			}
		}
	}
}
