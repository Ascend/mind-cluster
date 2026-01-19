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

// Package policy is used for processing superpod information
package policy

import (
	"encoding/json"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

const maxRetryTime = 180

// readConfigMap transform json to map[string]any
func readConfigMap(configMapPath string) *SuperPodInfo {
	jsonData, err := utils.LoadFile(configMapPath)
	if err != nil {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]read config map err: %v", err)
		return nil
	}
	var configMap SuperPodInfo
	if err := json.Unmarshal(jsonData, &configMap); err != nil {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]read config map err: %v", err)
		return nil
	}
	return &configMap
}

// GetCurSuperPodInfoFromMapA3 一个超节点的npu连接信息A3版本
func GetCurSuperPodInfoFromMapA3(superPodInfo *SuperPodInfo) ([]string, map[string][]string) {
	if superPodInfo == nil {
		hwlog.RunLog.Errorf("superPodInfo is nil")
		return nil, nil
	}

	/* traverse node device map to splice algorithm input */
	return parseNodeDeviceMap(superPodInfo.NodeDeviceMap)
}
