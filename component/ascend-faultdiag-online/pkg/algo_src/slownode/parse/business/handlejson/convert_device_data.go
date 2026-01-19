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
Package handlejson converts device json data
*/
package handlejson

import (
	"fmt"
	"strconv"
	"strings"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

// processDeviceData 处理Device的JSON数据
func processDeviceData(data *model.JsonData, cacheData *CacheData) error {
	if data.Domain == constants.DomainComm {
		if err := setCommOpData(cacheData, data); err != nil {
			return err
		}
		return nil
	}
	if data.Domain != constants.DomainDefault {
		return nil
	}
	if strings.Contains(strings.ToLower(data.Name), constants.StepWord) {
		if err := setStepTimeData(cacheData, data); err != nil {
			return err
		}
		return nil
	}
	if domainOther[data.Name] {
		if err := setTaskData(cacheData, data); err != nil {
			return err
		}
	}
	return nil
}

func setCommOpData(cacheData *CacheData, jsonData *model.JsonData) error {
	if cacheData.commOp == nil {
		cacheData.commOp = &model.CommOp{
			OpName:       jsonData.ParseName.IntOpName,
			StartNs:      constants.DefaultNameIndex,
			EndNs:        constants.DefaultNameIndex,
			ConnectionId: jsonData.Id,
			GroupName:    jsonData.ParseName.IntGroupName,
			OpId:         jsonData.Id,
			DataType:     jsonData.ParseName.IntDataType,
			Count:        jsonData.ParseName.IntCount,
			OpType:       jsonData.ParseName.IntOpName,
		}
	}
	if jsonData.Flag == constants.FlagMarkerStartWithDevice {
		cacheData.commOp.StartNs = jsonData.Timestamp
	}
	if jsonData.Flag == constants.FlagMarkerEndWithDevice {
		cacheData.commOp.EndNs = jsonData.Timestamp
	}
	cacheData.dbName = constants.DbCommOp
	return nil
}

func setStepTimeData(cacheData *CacheData, jsonData *model.JsonData) error {
	nameParts := strings.Fields(jsonData.Name)
	// name e.g: "step 1"
	const splitLen = 2
	if len(nameParts) < splitLen {
		return fmt.Errorf("failed to parse the 'Name' field, the 'Name' is: %s", jsonData.Name)
	}
	stepId, err := strconv.ParseInt(nameParts[1], constants.DecimalMark, constants.Base64Mark)
	if err != nil {
		return fmt.Errorf("failed to convert the name: %v", err)
	}
	if cacheData.stepTime == nil {
		cacheData.stepTime = &model.StepTime{
			Id: stepId, StartNs: constants.DefaultNameIndex, EndNs: constants.DefaultNameIndex,
		}
	}
	if jsonData.Flag == constants.FlagMarkerStartWithDevice {
		cacheData.stepTime.StartNs = jsonData.Timestamp
	}
	if jsonData.Flag == constants.FlagMarkerEndWithDevice {
		cacheData.stepTime.EndNs = jsonData.Timestamp
	}
	cacheData.dbName = constants.DbStepTime

	return nil
}

func setTaskData(cacheData *CacheData, jsonData *model.JsonData) error {
	if cacheData.task == nil {
		cacheData.task = &model.Task{
			ConnectionId: jsonData.Id, StartNs: constants.DefaultNameIndex, EndNs: constants.DefaultNameIndex,
		}
	}
	startTimeFlag := constants.FlagMarkerStartWithHost
	endTimeFlag := constants.FlagMarkerEndWithHost
	if jsonData.Name == constants.ForwardWord {
		startTimeFlag = constants.FlagMarkerStartWithDevice
		endTimeFlag = constants.FlagMarkerEndWithDevice
	}
	if jsonData.Flag == startTimeFlag {
		cacheData.task.StartNs = jsonData.Timestamp
	}
	if jsonData.Flag == endTimeFlag {
		cacheData.task.EndNs = jsonData.Timestamp
	}
	cacheData.dbName = constants.DbTask
	return nil

}
