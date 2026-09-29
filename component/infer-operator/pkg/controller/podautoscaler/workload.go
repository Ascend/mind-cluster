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

package podautoscaler

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "infer-operator/pkg/api/v1"
)

// InstanceSetScale adapts InstanceSet to the operations required by the KPA controller.
type InstanceSetScale struct{ client client.Client }

// NewInstanceSetScale creates an InstanceSet scaling adapter.
func NewInstanceSetScale(client client.Client) *InstanceSetScale {
	return &InstanceSetScale{client: client}
}

// GetTarget resolves the InstanceSet referenced by a PodAutoscaler.
func (scale *InstanceSetScale) GetTarget(ctx context.Context,
	pa *apiv1.PodAutoscaler) (*apiv1.InstanceSet, error) {
	target := &apiv1.InstanceSet{}
	key := types.NamespacedName{Namespace: pa.Namespace, Name: pa.Spec.ScaleTargetRef.Name}
	if err := scale.client.Get(ctx, key, target); err != nil {
		return nil, fmt.Errorf("failed to get target InstanceSet: %w", err)
	}
	return target, nil
}

// GetCurrentReplicas returns the desired replica count currently stored in the InstanceSet spec.
func (scale *InstanceSetScale) GetCurrentReplicas(target *apiv1.InstanceSet) (int32, error) {
	if target.Spec.Replicas == nil {
		return 0, fmt.Errorf("InstanceSet spec.replicas is not set")
	}
	return *target.Spec.Replicas, nil
}

// GetPodSelector parses the selector published by the InstanceSet controller.
func (scale *InstanceSetScale) GetPodSelector(target *apiv1.InstanceSet) (labels.Selector, error) {
	if target.Status.LabelSelector == "" {
		return nil, fmt.Errorf("InstanceSet status.labelSelector is not ready")
	}
	selector, err := labels.Parse(target.Status.LabelSelector)
	if err != nil {
		return nil, fmt.Errorf("invalid InstanceSet status.labelSelector: %w", err)
	}
	return selector, nil
}

// UpdateReplicas patches InstanceSet spec.replicas and retries optimistic-lock conflicts.
func (scale *InstanceSetScale) UpdateReplicas(ctx context.Context,
	pa *apiv1.PodAutoscaler, replicas int32) error {
	key := types.NamespacedName{Namespace: pa.Namespace, Name: pa.Spec.ScaleTargetRef.Name}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest := &apiv1.InstanceSet{}
		if err := scale.client.Get(ctx, key, latest); err != nil {
			return err
		}

		if latest.Spec.Replicas != nil && *latest.Spec.Replicas == replicas {
			return nil
		}

		original := latest.DeepCopy()
		latest.Spec.Replicas = &replicas
		return scale.client.Patch(ctx, latest, client.MergeFrom(original))
	})
}
