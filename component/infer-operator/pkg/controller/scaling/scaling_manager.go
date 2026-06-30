/*
Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package scaling

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"ascend-common/common-utils/hwlog"
	apiv1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

const (
	ScalingResourceOwnedByInstanceSet = common.LabelKeyPrefix + "scaling-owned-by"
)

// ScalingManager manages the lifecycle of scaling resources for InstanceSet objects.
type ScalingManager struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewScalingManager creates a new ScalingManager instance.
func NewScalingManager(cli client.Client, scheme *runtime.Scheme) *ScalingManager {
	return &ScalingManager{
		Client: cli,
		Scheme: scheme,
	}
}

// ReconcileScalingResource reconciles the scaling resources for the given InstanceSet.
func (m *ScalingManager) ReconcileScalingResource(
	ctx context.Context,
	instanceSet *apiv1.InstanceSet,
) (*apiv1.ScalingResourceStatus, error) {
	if instanceSet.Spec.ScalingPolicy == nil {
		return m.cleanupScalingResource(ctx, instanceSet)
	}

	switch instanceSet.Spec.ScalingPolicy.Type {
	case common.ScalingPolicyTypeHPA:
		return m.reconcileHPA(ctx, instanceSet)
	default:
		errMsg := fmt.Sprintf("unsupported scaling policy type: %s", instanceSet.Spec.ScalingPolicy.Type)
		hwlog.RunLog.Errorf("InstanceSet %s/%s: %s", instanceSet.Namespace, instanceSet.Name, errMsg)
		return &apiv1.ScalingResourceStatus{
			Type:    instanceSet.Spec.ScalingPolicy.Type,
			Ready:   false,
			Message: errMsg,
		}, nil
	}
}

func (m *ScalingManager) reconcileHPA(
	ctx context.Context,
	instanceSet *apiv1.InstanceSet,
) (*apiv1.ScalingResourceStatus, error) {
	hpaSpec := &autoscalingv2.HorizontalPodAutoscalerSpec{}
	if err := json.Unmarshal(instanceSet.Spec.ScalingPolicy.Spec.Raw, hpaSpec); err != nil {
		errMsg := fmt.Sprintf("failed to unmarshal HPA spec: %v", err)
		hwlog.RunLog.Errorf("InstanceSet %s/%s: %s", instanceSet.Namespace, instanceSet.Name, errMsg)
		return &apiv1.ScalingResourceStatus{
			Type:    common.ScalingPolicyTypeHPA,
			Ready:   false,
			Message: errMsg,
		}, nil
	}

	hpaSpec.ScaleTargetRef = autoscalingv2.CrossVersionObjectReference{
		APIVersion: instanceSet.GroupVersionKind().GroupVersion().String(),
		Kind:       instanceSet.GroupVersionKind().Kind,
		Name:       instanceSet.Name,
	}

	injectMetricSelectorLabels(hpaSpec, instanceSet)

	hpaName := buildScalingResourceName(instanceSet)
	existingHPA := &autoscalingv2.HorizontalPodAutoscaler{}
	err := m.Get(ctx, types.NamespacedName{Name: hpaName, Namespace: instanceSet.Namespace}, existingHPA)
	if err != nil && !apierrors.IsNotFound(err) {
		hwlog.RunLog.Errorf("InstanceSet %s/%s: failed to get HPA %s: %v",
			instanceSet.Namespace, instanceSet.Name, hpaName, err)
		return nil, err
	}

	if apierrors.IsNotFound(err) {
		return m.createHPA(ctx, instanceSet, hpaName, hpaSpec)
	}

	return m.updateHPA(ctx, instanceSet, existingHPA, hpaSpec)
}

func (m *ScalingManager) createHPA(
	ctx context.Context,
	instanceSet *apiv1.InstanceSet,
	hpaName string,
	hpaSpec *autoscalingv2.HorizontalPodAutoscalerSpec,
) (*apiv1.ScalingResourceStatus, error) {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:        hpaName,
			Namespace:   instanceSet.Namespace,
			Labels:      buildScalingResourceLabels(instanceSet),
			Annotations: instanceSet.Annotations,
		},
		Spec: *hpaSpec,
	}

	if err := controllerutil.SetControllerReference(instanceSet, hpa, m.Scheme); err != nil {
		hwlog.RunLog.Errorf("InstanceSet %s/%s: failed to set controller reference for HPA %s: %v",
			instanceSet.Namespace, instanceSet.Name, hpaName, err)
		return nil, err
	}

	if err := m.Create(ctx, hpa); err != nil {
		errMsg := fmt.Sprintf("failed to create HPA %s: %v", hpaName, err)
		hwlog.RunLog.Errorf("InstanceSet %s/%s: %s", instanceSet.Namespace, instanceSet.Name, errMsg)
		return &apiv1.ScalingResourceStatus{
			Type:    common.ScalingPolicyTypeHPA,
			Name:    hpaName,
			Ready:   false,
			Message: errMsg,
		}, err
	}

	hwlog.RunLog.Infof("InstanceSet %s/%s: created HPA %s successfully",
		instanceSet.Namespace, instanceSet.Name, hpaName)
	return &apiv1.ScalingResourceStatus{
		Type:    common.ScalingPolicyTypeHPA,
		Name:    hpaName,
		Ready:   true,
		Message: "HPA created successfully",
	}, nil
}

func (m *ScalingManager) updateHPA(
	ctx context.Context,
	instanceSet *apiv1.InstanceSet,
	existingHPA *autoscalingv2.HorizontalPodAutoscaler,
	desiredSpec *autoscalingv2.HorizontalPodAutoscalerSpec,
) (*apiv1.ScalingResourceStatus, error) {
	hpaName := existingHPA.Name

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestHPA := &autoscalingv2.HorizontalPodAutoscaler{}
		if err := m.Get(ctx, types.NamespacedName{Name: hpaName, Namespace: instanceSet.Namespace}, latestHPA); err != nil {
			return err
		}
		latestHPA.Spec = *desiredSpec
		return m.Update(ctx, latestHPA)
	})
	if err != nil {
		errMsg := fmt.Sprintf("failed to update HPA %s: %v", hpaName, err)
		hwlog.RunLog.Errorf("InstanceSet %s/%s: %s", instanceSet.Namespace, instanceSet.Name, errMsg)
		return &apiv1.ScalingResourceStatus{
			Type:    common.ScalingPolicyTypeHPA,
			Name:    hpaName,
			Ready:   false,
			Message: errMsg,
		}, err
	}

	hwlog.RunLog.Infof("InstanceSet %s/%s: updated HPA %s successfully",
		instanceSet.Namespace, instanceSet.Name, hpaName)
	return &apiv1.ScalingResourceStatus{
		Type:    common.ScalingPolicyTypeHPA,
		Name:    hpaName,
		Ready:   true,
		Message: "HPA updated successfully",
	}, nil
}

func (m *ScalingManager) cleanupScalingResource(
	ctx context.Context,
	instanceSet *apiv1.InstanceSet,
) (*apiv1.ScalingResourceStatus, error) {
	hpaList := &autoscalingv2.HorizontalPodAutoscalerList{}
	if err := m.List(ctx, hpaList,
		client.InNamespace(instanceSet.Namespace),
		client.MatchingLabels{ScalingResourceOwnedByInstanceSet: instanceSet.Name},
	); err != nil {
		hwlog.RunLog.Errorf("InstanceSet %s/%s: failed to list HPAs for cleanup: %v",
			instanceSet.Namespace, instanceSet.Name, err)
		return nil, err
	}

	for i := range hpaList.Items {
		hpa := &hpaList.Items[i]
		if err := m.Delete(ctx, hpa); err != nil && !apierrors.IsNotFound(err) {
			hwlog.RunLog.Errorf("InstanceSet %s/%s: failed to delete HPA %s: %v",
				instanceSet.Namespace, instanceSet.Name, hpa.Name, err)
			return nil, err
		}
		hwlog.RunLog.Infof("InstanceSet %s/%s: deleted HPA %s",
			instanceSet.Namespace, instanceSet.Name, hpa.Name)
	}

	return nil, nil
}

func buildScalingResourceName(instanceSet *apiv1.InstanceSet) string {
	return fmt.Sprintf("%s-scaler", instanceSet.Name)
}

func buildScalingResourceLabels(instanceSet *apiv1.InstanceSet) map[string]string {
	labels := make(map[string]string)
	for k, v := range instanceSet.Labels {
		labels[k] = v
	}
	labels[ScalingResourceOwnedByInstanceSet] = instanceSet.Name
	labels[common.OperatorNameKey] = ""
	return labels
}

func injectMetricSelectorLabels(hpaSpec *autoscalingv2.HorizontalPodAutoscalerSpec, instanceSet *apiv1.InstanceSet) {
	autoLabels := buildAutoInjectedLabels(instanceSet)
	if len(autoLabels) == 0 {
		return
	}

	for i := range hpaSpec.Metrics {
		if hpaSpec.Metrics[i].Type != autoscalingv2.ExternalMetricSourceType {
			continue
		}
		if hpaSpec.Metrics[i].External == nil {
			continue
		}
		external := hpaSpec.Metrics[i].External
		if external.Metric.Selector == nil {
			external.Metric.Selector = &metav1.LabelSelector{}
		}
		if external.Metric.Selector.MatchLabels == nil {
			external.Metric.Selector.MatchLabels = make(map[string]string)
		}
		for k, v := range autoLabels {
			if _, exists := external.Metric.Selector.MatchLabels[k]; !exists {
				external.Metric.Selector.MatchLabels[k] = v
			}
		}
	}
}

func buildAutoInjectedLabels(instanceSet *apiv1.InstanceSet) map[string]string {
	labels := make(map[string]string)

	if v, ok := instanceSet.Labels[common.InferServiceNameLabelKey]; ok {
		labels[common.InferServiceNameLabelKey] = v
		if idx := strings.LastIndex(v, "-"); idx >= 0 {
			labels[common.InferServiceSetNameLabelKey] = v[:idx]
		}
	}
	if v, ok := instanceSet.Labels[common.InstanceSetNameLabelKey]; ok {
		labels[common.RoleNameLabelKey] = v
	}

	return labels
}
