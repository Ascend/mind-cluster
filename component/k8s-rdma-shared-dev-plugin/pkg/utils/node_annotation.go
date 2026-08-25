/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ascend-common/common-utils/hwlog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const nodeAnnotationRetryTime = 3

// UpdateNodeAnnotation patches a node annotation using StrategicMergePatch.
func UpdateNodeAnnotation(clientset kubernetes.Interface, nodeName, key, value string) error {
	patchPayload := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{
				key: value,
			},
		},
	}
	patchByte, err := json.Marshal(patchPayload)
	if err != nil {
		return fmt.Errorf("marshal annotation patch failed: %v", err)
	}

	for i := 0; i < nodeAnnotationRetryTime; i++ {
		_, err = clientset.CoreV1().Nodes().Patch(context.TODO(), nodeName,
			types.StrategicMergePatchType, patchByte, metav1.PatchOptions{})
		if err == nil {
			hwlog.RunLog.Infof("update node annotation %s success", key)
			return nil
		}
		hwlog.RunLog.Warnf("patch node annotation failed, err: %v, retry: %d", err, i+1)
		time.Sleep(time.Second)
	}
	return err
}
