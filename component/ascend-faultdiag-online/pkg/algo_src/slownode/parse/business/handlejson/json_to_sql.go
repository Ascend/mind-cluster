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
Package handlejson converts JSON to DB statements
*/
package handlejson

import (
	"encoding/json"
	"fmt"
	"strconv"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/business/dealdb"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/db"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

// StartParseJsonDataToSql 将jsonData转为sql
func StartParseJsonDataToSql(snpRankCtx *context.SnpRankContext, parseFileCtx *ParseFileContext,
	stopFlag chan struct{}) {
	go parseJsonData(snpRankCtx, parseFileCtx, stopFlag)
	go parseJsonDataToSql(snpRankCtx, parseFileCtx, stopFlag)
}

func parseJsonData(snpRankCtx *context.SnpRankContext, parseFileCtx *ParseFileContext, stopFlag chan struct{}) {
	for {
		select {
		case _, ok := <-stopFlag:
			if !ok {
				return
			}
		case jsonData, ok := <-snpRankCtx.JsonDataQue:
			if !ok {
				return
			}
			err := updateJsonData(jsonData, snpRankCtx.ContextData.DbCtx, parseFileCtx)
			if err != nil {
				continue
			}
			parseFileCtx.DealJsonData(jsonData)
		}
	}
}

func parseJsonDataToSql(snpRankCtx *context.SnpRankContext, parseFileCtx *ParseFileContext, stopFlag chan struct{}) {
	for {
		select {
		case _, ok := <-stopFlag:
			if !ok {
				return
			}
		case stack, ok := <-parseFileCtx.EventStackQue:
			if !ok {
				return
			}
			cacheData := &CacheData{}
			handelCacheData(stack, cacheData)
			sqlSlice, err := ProcessInsertSQL([]*CacheData{cacheData})
			if err != nil {
				return
			}
			for _, sql := range sqlSlice {
				snpRankCtx.InsertSqlQue <- sql
			}
		}
	}
}

func handelCacheData(stack *EventStack, cacheData *CacheData) {
	for _, data := range stack.JsonDataList {
		if data.SourceKind == constants.SourceKindHost {
			if err := processHostData(data, cacheData); err != nil {
				hwlog.RunLog.Warnf("[SLOWNODE PARSE]Failed to processing host data: %v", err)
			}
		} else if data.SourceKind == constants.SourceKindDevice {
			if err := processDeviceData(data, cacheData); err != nil {
				hwlog.RunLog.Warnf("[SLOWNODE PARSE]Failed to processing device data: %v", err)
			}
		}
	}
}

// updateJsonData 更新json数据
func updateJsonData(jsonData *model.JsonData, dbCtx *db.SnpDbContext, parseFileCtx *ParseFileContext) error {
	if dbCtx == nil {
		return nil
	}
	parseFileCtx.HandleIdDomain(jsonData)
	if jsonData.Name == "" || jsonData.Domain == "" {
		return nil
	}
	nameId, err := stringConvertId(dbCtx, jsonData.Name)
	if err != nil {
		return err
	}
	jsonData.ParseName.NameId = nameId

	if jsonData.Domain == constants.DomainDefault {
		return nil
	}
	var nameInstance *model.JsonName
	if err := json.Unmarshal([]byte(jsonData.Name), &nameInstance); err != nil {
		return fmt.Errorf(
			"the value of 'Name' in the JSON is: '%s', error unmarshalling JSON: %v", jsonData.Name, err)
	}
	if nameInstance == nil {
		return fmt.Errorf("invalid json name: %s", jsonData.Name)
	}

	intDataType, err := stringConvertId(dbCtx, nameInstance.DataType)
	if err != nil {
		return err
	}
	jsonData.ParseName.IntDataType = intDataType

	intGroupName, err := stringConvertId(dbCtx, nameInstance.GroupName)
	if err != nil {
		return err
	}
	jsonData.ParseName.IntGroupName = intGroupName

	intOpName, err := stringConvertId(dbCtx, nameInstance.OpName)
	if err != nil {
		return err
	}
	jsonData.ParseName.IntOpName = intOpName

	count, err := strconv.ParseInt(nameInstance.Count, constants.DecimalMark, constants.Base64Mark)
	if err != nil {
		return err
	}
	jsonData.ParseName.IntCount = count
	return nil
}

// stringConvertId 赋值stringIdsMap
func stringConvertId(dbCtx *db.SnpDbContext, value string) (int64, error) {
	queryIdView, err := dealdb.QueryStringId(dbCtx, value)
	if err != nil {
		return -1, err
	}
	if queryIdView != nil {
		return queryIdView.Id, nil
	}

	idView, err := dealdb.InsertStringIds(dbCtx, value)
	if err != nil {
		return -1, err
	}
	if idView != nil {
		return idView.Id, nil
	}
	return -1, nil
}
