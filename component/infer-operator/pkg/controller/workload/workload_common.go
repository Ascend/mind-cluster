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

package workload

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/common"
)

// deletePodsForExternalRescheduling deletes pods belonging to the workload
// when the workload is configured with external-force rescheduling mode.
// It uses the workload's own identifying labels (inferServiceName, instanceSetName, instanceSetIndex)
// to accurately locate pods owned by this specific workload.
func deletePodsForExternalRescheduling(ctx context.Context, cli client.Client,
	workload WorkLoadInterface) error {
	meta := workload.GetWorkLoadObjMeta()
	if meta.Labels[common.FaultSchedulingLabelKey] != common.ExternalForceReschedulingValue {
		return nil
	}
	podLabels := client.MatchingLabels{
		common.InferServiceNameLabelKey: meta.Labels[common.InferServiceNameLabelKey],
		common.InstanceSetNameLabelKey:  meta.Labels[common.InstanceSetNameLabelKey],
		common.InstanceIndexLabelKey:    meta.Labels[common.InstanceIndexLabelKey],
	}
	podList := &corev1.PodList{}
	if err := cli.List(ctx, podList, client.InNamespace(meta.Namespace), podLabels); err != nil {
		return fmt.Errorf("failed to list pods for work load %s/%s: %w",
			meta.Namespace, meta.Name, err)
	}
	deleteOpts := []client.DeleteOption{
		client.GracePeriodSeconds(0),
	}
	for _, pod := range podList.Items {
		if err := cli.Delete(ctx, &pod, deleteOpts...); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			hwlog.RunLog.Errorf("failed to force delete pod %s/%s: %v", pod.Namespace, pod.Name, err)
		}
	}
	return nil
}
