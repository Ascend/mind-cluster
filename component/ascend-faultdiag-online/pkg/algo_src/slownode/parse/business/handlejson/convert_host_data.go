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
Package handlejson converts host json data
*/
package handlejson

import (
	"strings"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

var domainOther = map[string]bool{
	constants.CKPTWord:       true,
	constants.ForwardWord:    true,
	constants.DataLoaderWord: true,
}

// processHostData 处理Host的JSON数据
func processHostData(data *model.JsonData, cacheData *CacheData) error {
	if data.Domain == constants.DomainComm {
		if err := setCAnnApiData(cacheData, data); err != nil {
			return err
		}
		return nil
	}
	if data.Domain != constants.DomainDefault {
		return nil
	}
	if domainOther[data.Name] || strings.Contains(strings.ToLower(data.Name), constants.StepWord) {
		if err := setMSTXEventsData(cacheData, data); err != nil {
			return err
		}
	}
	return nil
}

func setCAnnApiData(cacheData *CacheData, jsonData *model.JsonData) error {
	if cacheData.cAnnApi == nil {
		cacheData.cAnnApi = &model.CAnnApi{
			ConnectionId: jsonData.Id, Name: jsonData.ParseName.NameId, StartNs: constants.DefaultNameIndex,
			EndNs: constants.DefaultNameIndex,
		}
	}
	if jsonData.Flag == constants.FlagMarkerStartWithDevice {
		cacheData.cAnnApi.StartNs = jsonData.Timestamp
	}
	if jsonData.Flag == constants.FlagMarkerEndWithDevice {
		cacheData.cAnnApi.EndNs = jsonData.Timestamp
	}
	cacheData.dbName = constants.DbCAnnApi
	return nil
}

func setMSTXEventsData(cacheData *CacheData, jsonData *model.JsonData) error {
	if cacheData.mSTXEvents == nil {
		cacheData.mSTXEvents = &model.MSTXEvents{
			ConnectionId: jsonData.Id, Message: jsonData.ParseName.NameId, StartNs: constants.DefaultNameIndex,
			EndNs: constants.DefaultNameIndex,
		}
	}
	startTimeFlag := constants.FlagMarkerStartWithHost
	endTimeFlag := constants.FlagMarkerEndWithHost
	if jsonData.Name == constants.ForwardWord {
		startTimeFlag = constants.FlagMarkerStartWithDevice
		endTimeFlag = constants.FlagMarkerEndWithDevice
	}

	if jsonData.Flag == startTimeFlag {
		cacheData.mSTXEvents.StartNs = jsonData.Timestamp
	}
	if jsonData.Flag == endTimeFlag {
		cacheData.mSTXEvents.EndNs = jsonData.Timestamp
	}
	cacheData.dbName = constants.DbMSTXEvents
	return nil
}
