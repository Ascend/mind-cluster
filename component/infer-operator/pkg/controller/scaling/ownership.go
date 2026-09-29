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

package scaling

import (
	"context"
	"fmt"
	"sort"
	"strings"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "infer-operator/pkg/api/v1"
)

// OwnershipResolver is the shared source of truth used by HPA and KPA controllers.
// Existing HPAs win; otherwise the oldest valid PodAutoscaler wins deterministically.
type OwnershipResolver struct{ client client.Client }

// NewOwnershipResolver creates a resolver backed by live Kubernetes resources.
func NewOwnershipResolver(client client.Client) *OwnershipResolver {
	return &OwnershipResolver{client: client}
}

// CanManageWithHPA reports whether the InstanceSet may be managed by HPA.
// An already-existing HPA wins over all PodAutoscaler candidates.
func (resolver *OwnershipResolver) CanManageWithHPA(ctx context.Context, target *apiv1.InstanceSet) (bool, string, error) {
	hpa, err := resolver.hpaForTarget(ctx, target.Namespace, target.Name)
	if err != nil {
		return false, "", err
	}
	if hpa != nil {
		return true, "", nil
	}
	winner, err := resolver.kpaWinner(ctx, target.Namespace, target.Name)
	if err != nil {
		return false, "", err
	}
	if winner != nil {
		return false, fmt.Sprintf("InstanceSet is controlled by PodAutoscaler %s/%s", winner.Namespace, winner.Name), nil
	}
	return true, "", nil
}

// CanManageWithKPA reports whether pa is the deterministic KPA owner of its target.
// Without an HPA, the oldest valid PodAutoscaler wins; names break timestamp ties.
func (resolver *OwnershipResolver) CanManageWithKPA(ctx context.Context, pa *apiv1.PodAutoscaler) (bool, string, error) {
	hpa, err := resolver.hpaForTarget(ctx, pa.Namespace, pa.Spec.ScaleTargetRef.Name)
	if err != nil {
		return false, "", err
	}
	if hpa != nil {
		return false, fmt.Sprintf("InstanceSet is controlled by HorizontalPodAutoscaler %s/%s", hpa.Namespace, hpa.Name), nil
	}
	winner, err := resolver.kpaWinner(ctx, pa.Namespace, pa.Spec.ScaleTargetRef.Name)
	if err != nil {
		return false, "", err
	}
	if winner != nil && winner.Name != pa.Name {
		return false, fmt.Sprintf("InstanceSet is controlled by PodAutoscaler %s/%s", winner.Namespace, winner.Name), nil
	}
	return true, "", nil
}

func (resolver *OwnershipResolver) hpaForTarget(ctx context.Context, namespace, targetName string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	list := &autoscalingv2.HorizontalPodAutoscalerList{}
	if err := resolver.client.List(ctx, list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	for i := range list.Items {
		hpa := &list.Items[i]
		ref := hpa.Spec.ScaleTargetRef
		if ref.Name == targetName && ref.Kind == "InstanceSet" &&
			(ref.APIVersion == apiv1.GroupVersion.String() || ref.APIVersion == "") {
			return hpa, nil
		}
	}
	return nil, nil
}

func (resolver *OwnershipResolver) kpaWinner(ctx context.Context, namespace, targetName string) (*apiv1.PodAutoscaler, error) {
	list := &apiv1.PodAutoscalerList{}
	if err := resolver.client.List(ctx, list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	candidates := make([]*apiv1.PodAutoscaler, 0, len(list.Items))
	for i := range list.Items {
		pa := &list.Items[i]
		ref := pa.Spec.ScaleTargetRef
		if pa.DeletionTimestamp == nil && strings.EqualFold(pa.Spec.ScalingStrategy, apiv1.ScalingStrategyKPA) &&
			ref.Name == targetName && ref.Kind == "InstanceSet" && ref.APIVersion == apiv1.GroupVersion.String() {
			candidates = append(candidates, pa)
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]
		if left.CreationTimestamp.Equal(&right.CreationTimestamp) {
			return types.NamespacedName{Namespace: left.Namespace, Name: left.Name}.String() <
				types.NamespacedName{Namespace: right.Namespace, Name: right.Name}.String()
		}
		return left.CreationTimestamp.Before(&right.CreationTimestamp)
	})
	return candidates[0], nil
}
