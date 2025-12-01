/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package pingmesh a series of function handle ping mesh configmap create/update/delete.
*/
package pingmesh

import (
	"encoding/json"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

// handlerSuperPodRoce handle ras netfault superpod roce info
func handlerSuperPodRoce(res map[int]string) {
	// 0.get all the A5 Super Pod id list
	superPodIds := filterA5SuperPodIDFromMap(res)
	if len(superPodIds) == 0 {
		hwlog.RunLog.Info("no A5 Super Pod in this cluster, can't create super-pod-roce dir")
		return
	}
	listData := &constant.CrossSuperPodListData{
		SuperPodList: make([]int, 0),
	}
	// 1.append to the super-pod-roce list
	listData.SuperPodList = append(listData.SuperPodList, superPodIds...)

	// 2.check roce key whether exist
	ConfigPingMeshInst.RLock()
	if _, exist := ConfigPingMeshInst.configCMInfo[constant.RasRoceKey]; !exist {
		hwlog.RunLog.Warnf("the roce config is not exist")
		ConfigPingMeshInst.RUnlock()
		return
	}
	ConfigPingMeshInst.RUnlock()
	// 3.create conf from configmap
	cfg := ConfigPingMeshInst.getRasConfigBySuperPodId(constant.RasRoceKey)

	hwlog.RunLog.Infof("these super-pod list<%v> will write to file", listData.SuperPodList)
	// 4.create dir and create file or update file
	b, err := json.MarshalIndent(listData, "", jsonIndent)
	if err != nil || len(b) == 0 {
		hwlog.RunLog.Errorf("marshal bytes illegal, SuperPodID=%s, err=%v", constant.RasRoceKey, err)
		return
	}
	if errWrite := writeJsonDataByteToFile(constant.RasRoceKey, b); errWrite != nil {
		hwlog.RunLog.Errorf("write json data to file failed, err: %v", errWrite)
		return
	}
	
	// 5.create .conf file or update
	err = saveConfigToFile(constant.RasRoceKey, cfg)
	if err != nil {
		hwlog.RunLog.Errorf("save config to file failed, err=%v, superPodID=%s", err, constant.RasRoceKey)
		return
	}
}

// filterA5SuperPodIDFromMap filter a5 type superpod info
func filterA5SuperPodIDFromMap(resMap map[int]string) []int {
	res := make([]int, 0, len(resMap))
	for id, acceleratorType := range resMap {
		if acceleratorType != api.A5PodType {
			hwlog.RunLog.Infof("the superPodID<%d> is not A5 Super Pod and do not append to list", id)
			continue
		}
		res = append(res, id)
	}
	return res
}
