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
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/common"
)

// deletePodsForExternalRescheduling deletes pods belonging to the workload
// when the workload is configured with external rescheduling mode.
// It uses the workload's own identifying labels (inferServiceName, instanceSetName, instanceSetIndex)
// to accurately locate pods owned by this specific workload.
//
// - external-force: immediately force-deletes pods with GracePeriodSeconds(0).
// - external-grace: reads the Pod's own TerminationGracePeriodSeconds, starts a timer,
//   and force-deletes any remaining pods after the grace period expires.
func deletePodsForExternalRescheduling(ctx context.Context, cli client.Client,
	workload WorkLoadInterface) error {
	meta := workload.GetWorkLoadObjMeta()
	mode := meta.Labels[common.FaultSchedulingLabelKey]

	if mode != common.ExternalForceReschedulingValue &&
		mode != common.ExternalGraceReschedulingValue {
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

	if len(podList.Items) == 0 {
		return nil
	}

	switch mode {
	case common.ExternalForceReschedulingValue:
		forceDeletePodList(ctx, cli, podList.Items)

	default:
		waitSeconds := int64(common.DefaultTerminationGracePeriodSeconds)
		if grace := podList.Items[0].Spec.TerminationGracePeriodSeconds; grace != nil && *grace > 0 {
			waitSeconds = *grace
		}

		ns := meta.Namespace
		name := meta.Name
		hwlog.RunLog.Infof("waiting %d seconds for graceful deletion of %d pod(s) in work load %s/%s, "+
			"will force delete remaining pods after timeout",
			waitSeconds, len(podList.Items), ns, name)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					hwlog.RunLog.Errorf("panic in force-delete goroutine for work load %s/%s: %v",
						ns, name, r)
				}
			}()
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			forceDeletePodsAfterGrace(context.WithoutCancel(ctx), cli, ns, name, podLabels)
		}()
	}

	return nil
}

// forceDeletePodsAfterGrace lists pods by the given labels and force-deletes
// any remaining pods with GracePeriodSeconds(0). It is intended to be called
// after a grace period has elapsed for external-grace rescheduling mode.
func forceDeletePodsAfterGrace(ctx context.Context, cli client.Client,
	ns, workloadName string, podLabels client.MatchingLabels) {
	podList := &corev1.PodList{}
	if err := cli.List(ctx, podList, client.InNamespace(ns), podLabels); err != nil {
		hwlog.RunLog.Errorf("failed to list pods after grace period for work load %s/%s: %v",
			ns, workloadName, err)
		return
	}

	if len(podList.Items) == 0 {
		return
	}

	hwlog.RunLog.Infof("force deleting %d remaining pod(s) after grace period for work load %s/%s",
		len(podList.Items), ns, workloadName)

	forceDeletePodList(ctx, cli, podList.Items)
}

// forceDeletePodList force-deletes all given pods with GracePeriodSeconds(0).
// NotFound errors are skipped, other errors are logged and the loop continues.
func forceDeletePodList(ctx context.Context, cli client.Client, pods []corev1.Pod) {
	for _, pod := range pods {
		if err := cli.Delete(ctx, &pod, client.GracePeriodSeconds(0)); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			hwlog.RunLog.Errorf("failed to force delete pod %s/%s: %v", pod.Namespace, pod.Name, err)
		}
	}
}
