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

// Package kubeclient a series of k8s function
package kubeclient

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"ascend-common/common-utils/hwlog"
)

// ParseFaultNetworkInfoCM parse fault network feature cm
func ParseFaultNetworkInfoCM(obj interface{}) (ConfigPingMesh, error) {
	configCm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return ConfigPingMesh{}, fmt.Errorf("not fault network of ras feature configmap")
	}
	configInfo := ConfigPingMesh{}
	// marshal every item to config struct
	for key, config := range configCm.Data {
		pingMeshItem := HccspingMeshItem{}
		if unmarshalErr := json.Unmarshal([]byte(config), &pingMeshItem); unmarshalErr != nil {
			return ConfigPingMesh{}, fmt.Errorf("unmarshal failed: %v, configmap name: %s",
				unmarshalErr, configCm.Name)
		}
		configInfo[key] = &pingMeshItem
	}

	return configInfo, nil
}

// IsConfigMapExists judge the configMap with specific name whether exist
func IsConfigMapExists(client kubernetes.Interface, namespace, name string) (*v1.ConfigMap, bool) {
	configMap, err := client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil || configMap == nil {
		hwlog.RunLog.Infof("the configmap<name: %s , namespace: %s> is not exist and err is %v",
			name, namespace, err)
		return nil, false
	}

	return configMap, true
}
