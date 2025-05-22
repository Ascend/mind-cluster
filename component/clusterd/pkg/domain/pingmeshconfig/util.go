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

// Package pingmeshconfig for faultnetwork feature
package pingmeshconfig

import (
	"encoding/json"
	"fmt"

	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

// ParseFaultNetworkInfoCM parse fault network feature cm
func ParseFaultNetworkInfoCM(obj interface{}) (constant.ConfigPingMesh, error) {
	configCm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return constant.ConfigPingMesh{}, fmt.Errorf("not fault network of ras feature configmap")
	}
	configInfo := constant.ConfigPingMesh{}
	// marshal every item to config struct
	for key, config := range configCm.Data {
		pingMeshItem := constant.HccspingMeshItem{}
		if unmarshalErr := json.Unmarshal([]byte(config), &pingMeshItem); unmarshalErr != nil {
			return constant.ConfigPingMesh{}, fmt.Errorf("unmarshal failed: %v, configmap name: %s",
				unmarshalErr, configCm.Name)
		}
		configInfo[key] = &pingMeshItem
	}

	return configInfo, nil
}

// DeepCopy deep copy FaultNetwork
func DeepCopy(info constant.ConfigPingMesh) constant.ConfigPingMesh {
	if info == nil {
		return nil
	}
	data, err := json.Marshal(info)
	if err != nil {
		hwlog.RunLog.Errorf("marshal pingmesh config info failed, err is %v", err)
		return nil
	}
	newInfo := constant.ConfigPingMesh{}
	if err := json.Unmarshal(data, &newInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal pingmesh config info failed, err is %v", err)
		return nil
	}
	return newInfo
}
