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
Package handlejson generates insert SQL statements
*/
package handlejson

import (
	"fmt"
	"strings"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils"
)

var (
	cAnnApiColumns = []string{"startNs", "endNs", "type", "globalTid", "connectionId", "name"}
	commOpColumns  = []string{
		"opName", "startNs", "endNs", "connectionId", "groupName", "opId", "relay", "retry", "dataType", "algType",
		"count", "opType",
	}
	mSTXEventsColumns = []string{
		"startNs", "endNs", "eventType", "rangeId", "category", "message", "globalTid", "endGlobalTid", "domainId",
		"connectionId",
	}
	stepTimeColumns = []string{"id", "startNs", "endNs"}
	taskColumns     = []string{
		"startNs", "endNs", "deviceId", "connectionId", "globalTaskId", "globalPid", "taskType", "contextId",
		"streamId", "taskId", "modelId",
	}
)

// ProcessInsertSQL 生成数据库插入语句
func ProcessInsertSQL(rowData []*CacheData) ([]string, error) {
	var cAnnApiValues []string
	var commOpValues []string
	mSTXEventsValues := make([]string, 0)
	stepTimeValues := make([]string, 0)
	taskValues := make([]string, 0)
	for _, row := range rowData {
		switch row.dbName {
		case constants.DbCAnnApi:
			cAnnApiValues = getCAnnApiValues(row.cAnnApi, cAnnApiValues)
		case constants.DbCommOp:
			commOpValues = getCommOpValues(row.commOp, commOpValues)
		case constants.DbMSTXEvents:
			mSTXEventsValues = getMSTXEventsValues(row.mSTXEvents, mSTXEventsValues)
		case constants.DbStepTime:
			stepTimeValues = getStepTimeValues(row.stepTime, stepTimeValues)
		case constants.DbTask:
			taskValues = getTaskValues(row.task, taskValues)
		default:
		}
	}

	var insertSql []string
	if len(cAnnApiValues) != 0 {
		insertSql = formatSQL(constants.DbCAnnApi, cAnnApiColumns, cAnnApiValues, insertSql)
	}
	if len(commOpValues) != 0 {
		insertSql = formatSQL(constants.DbCommOp, commOpColumns, commOpValues, insertSql)
	}
	if len(mSTXEventsValues) != 0 {
		insertSql = formatSQL(constants.DbMSTXEvents, mSTXEventsColumns, mSTXEventsValues, insertSql)
	}
	if len(stepTimeValues) != 0 {
		insertSql = formatSQL(constants.DbStepTime, stepTimeColumns, stepTimeValues, insertSql)
	}
	if len(taskValues) != 0 {
		insertSql = formatSQL(constants.DbTask, taskColumns, taskValues, insertSql)
	}
	return insertSql, nil
}

func getCAnnApiValues(cAnnApi *model.CAnnApi, cAnnApiValues []string) []string {
	if cAnnApi.StartNs == constants.DefaultNameIndex || cAnnApi.EndNs == constants.DefaultNameIndex {
		// 不记录这条数据，下一次读取时再记录
		return cAnnApiValues
	}
	cAnnApiValueStr := fmt.Sprintf("(%d, %d, %d, %d, %d, %d)",
		cAnnApi.StartNs, cAnnApi.EndNs, cAnnApi.ApiType, cAnnApi.GlobalTid, cAnnApi.ConnectionId, cAnnApi.Name,
	)
	return append(cAnnApiValues, cAnnApiValueStr)
}

func getCommOpValues(commOp *model.CommOp, commOpValues []string) []string {
	if commOp.StartNs == constants.DefaultNameIndex || commOp.EndNs == constants.DefaultNameIndex {
		// 不记录这条数据，下一次读取时再记录
		return commOpValues
	}
	commOpValueStr := fmt.Sprintf("(%d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d)",
		commOp.OpName, commOp.StartNs, commOp.EndNs, commOp.ConnectionId, commOp.GroupName, commOp.OpId, commOp.Relay,
		commOp.Retry, commOp.DataType, commOp.AlgType, commOp.Count, commOp.OpType,
	)
	return append(commOpValues, commOpValueStr)
}

func getMSTXEventsValues(mSTXEvents *model.MSTXEvents, mSTXEventsValues []string) []string {
	if mSTXEvents.StartNs == constants.DefaultNameIndex || mSTXEvents.EndNs == constants.DefaultNameIndex {
		// 不记录这条数据，下一次读取时再记录
		return mSTXEventsValues
	}
	mSTXEventsValueStr := fmt.Sprintf("(%d, %d, %d, %d, %d, %d, %d, %d, %d, %d)",
		mSTXEvents.StartNs, mSTXEvents.EndNs, mSTXEvents.EventType, mSTXEvents.RangeId, mSTXEvents.Category,
		mSTXEvents.Message, mSTXEvents.GlobalTid, mSTXEvents.EndGlobalTid, mSTXEvents.DomainId, mSTXEvents.ConnectionId,
	)
	return append(mSTXEventsValues, mSTXEventsValueStr)
}

func getStepTimeValues(stepTime *model.StepTime, stepTimeValues []string) []string {
	if stepTime.StartNs == constants.DefaultNameIndex || stepTime.EndNs == constants.DefaultNameIndex {
		// 不记录这条数据，下一次读取时再记录
		return stepTimeValues
	}
	stepTimeValueStr := fmt.Sprintf("(%d, %d, %d)",
		stepTime.Id, stepTime.StartNs, stepTime.EndNs,
	)
	if len(stepTimeValues) == 0 {
		return append(stepTimeValues, stepTimeValueStr)
	}
	if !utils.StrContains(stepTimeValues, fmt.Sprintf("(%d,", stepTime.Id)) {
		return append(stepTimeValues, stepTimeValueStr)
	}

	return stepTimeValues
}

func getTaskValues(task *model.Task, taskValues []string) []string {
	if task.StartNs == constants.DefaultNameIndex || task.EndNs == constants.DefaultNameIndex {
		// 不记录这条数据，下一次读取时再记录
		return taskValues
	}
	taskValueStr := fmt.Sprintf("(%d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d)",
		task.StartNs, task.EndNs, task.DeviceId, task.ConnectionId, task.GlobalTaskId, task.GlobalPid, task.TaskType,
		task.ContextId, task.StreamId, task.TaskId, task.ModelId,
	)
	return append(taskValues, taskValueStr)
}

func formatSQL(dbName string, columns []string, values []string, insertSql []string) []string {
	columnsStr := strings.Join(columns, ", ")
	lenValues := len(values)
	for index := 0; index < lenValues; index += constants.InsertNumber {
		endIndex := index + constants.InsertNumber
		if endIndex > lenValues {
			endIndex = lenValues
		}
		valuesStr := strings.Join((values)[index:endIndex], ", ")
		insertSqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s;", dbName, columnsStr, valuesStr)
		insertSql = append(insertSql, insertSqlStr)
	}
	return insertSql
}
