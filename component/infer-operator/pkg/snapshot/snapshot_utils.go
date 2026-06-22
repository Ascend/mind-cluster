/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package snapshot

import (
	"path/filepath"

	corev1 "k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/common"
)

// GetHostSnapshotPath returns the host snapshot path for the pod.
func GetHostSnapshotPath(pod *corev1.Pod) string {
	snapshotPath := ""
	if len(pod.Spec.Containers) <= 0 {
		hwlog.RunLog.Errorf("Pod has no containers, cannot get host snapshot path")
		return ""
	}
	for _, env := range pod.Spec.Containers[0].Env {
		if env.Name == common.HostSnapshotDirPathEnvKey && env.Value != "" {
			snapshotPath = env.Value
			break
		}
	}
	if snapshotPath == "" {
		hwlog.RunLog.Errorf("Host snapshot path env '%s' not found in pod %s/%s, "+
			"cannot get snapshot path", common.HostSnapshotDirPathEnvKey, pod.Namespace, pod.Name)
		return ""
	}

	instancesetName := common.GetInstanceSetNameFromLabels(pod.Labels)
	snapshotPath = filepath.Join(snapshotPath, pod.Namespace, instancesetName)
	return snapshotPath
}
