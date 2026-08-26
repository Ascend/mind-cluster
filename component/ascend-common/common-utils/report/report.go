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

// Package report for report version info
package report

import (
	"context"
	"encoding/json"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/version"
)

// ReportVersionToConfigMap report version info to configmap through controller-runtime
func ReportVersionToConfigMap(clt client.Client, ctx context.Context, info version.Info, componentName string) {
	cmData := map[string]string{componentName: version.ToJSON(info)}
	newCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      api.VersionName,
			Namespace: api.DLNamespace,
		},
		Data: cmData,
	}

	backoff := 1 * time.Second
	for attempt := 0; attempt < version.VersionReportMaxAttempts; attempt++ {
		err := clt.Create(ctx, newCM)
		if err == nil {
			return
		}
		if !errors.IsAlreadyExists(err) {
			hwlog.RunLog.Errorf("create cm failed, err is %v", err)
			break
		}

		patchData := map[string]any{
			"data": cmData,
		}
		patchBytes, err := json.Marshal(patchData)
		if err != nil {
			hwlog.RunLog.Errorf("marshal patch data failed, err is %v", err)
			return
		}
		cm := &corev1.ConfigMap{}
		cm.Name = api.VersionName
		cm.Namespace = api.DLNamespace
		err = clt.Patch(ctx, cm, client.RawPatch(types.StrategicMergePatchType, patchBytes))
		if err == nil {
			return
		}
		hwlog.RunLog.Errorf("patch cm data failed, err is %v", err)
		time.Sleep(backoff)
		backoff *= 2
	}
	hwlog.RunLog.Errorf("failed to report %s version to configmap after 3 attempts", componentName)
}
