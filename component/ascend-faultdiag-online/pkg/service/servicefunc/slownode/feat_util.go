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

// Package slownode a series of feature function
package slownode

import (
	"encoding/json"
	"fmt"

	core "k8s.io/api/core/v1"
)

// ParseCMResult is a general func parsing source to result
func ParseCMResult(source any, cmKey string, result any) error {
	cm, ok := source.(*core.ConfigMap)
	if !ok {
		return fmt.Errorf("[FD-OL SLOWNODE]source %s is not a feature configmap", source)
	}
	data, ok := cm.Data[cmKey]
	if !ok {
		return fmt.Errorf("[FD-OL SLOWNODE]configmap %s has no key %s", cm.Name, cmKey)
	}

	if err := json.Unmarshal([]byte(data), result); err != nil {
		return fmt.Errorf("[FD-OL SLOWNODE]json unmarshal failed: %v, configmap name: %s and source data: %s",
			err, cm.Name, data)
	}
	return nil
}
