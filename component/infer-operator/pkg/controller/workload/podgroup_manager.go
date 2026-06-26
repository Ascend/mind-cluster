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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

type PodGroupManager interface {
	// GetPodGroupForInstance gets podgroup for instance.
	GetPodGroupForInstance(ctx context.Context, instanceSet *v1.InstanceSet,
		indexer common.InstanceIndexer) (*v1beta1.PodGroup, error)
	// GetOrCreatePodGroupForInstance gets or creates podgroup for instance.
	GetOrCreatePodGroupForInstance(ctx context.Context, instanceSet *v1.InstanceSet,
		indexer common.InstanceIndexer, spec v1beta1.PodGroupSpec) (bool, error)
	// DeletePodGroupForInstance deletes podgroup for instance.
	DeletePodGroupForInstance(ctx context.Context, instanceSet *v1.InstanceSet,
		indexer common.InstanceIndexer) error
}

type VolcanoPodGroupManager struct {
	client client.Client
}

// GetPodGroupForInstance gets podgroup for instance.
func (v *VolcanoPodGroupManager) GetPodGroupForInstance(
	ctx context.Context,
	instanceSet *v1.InstanceSet,
	indexer common.InstanceIndexer) (*v1beta1.PodGroup, error) {
	pgName := common.GetPGNameFromIndexer(indexer)
	podGroup := &v1beta1.PodGroup{}
	err := v.client.Get(ctx, types.NamespacedName{Name: pgName, Namespace: instanceSet.Namespace}, podGroup)
	if err != nil && !errors.IsNotFound(err) {
		hwlog.RunLog.Errorf("get podgroup<%s> error: %v", pgName, err)
		return nil, err
	}
	if errors.IsNotFound(err) {
		hwlog.RunLog.Infof("podgroup<%s> not found", pgName)
	}
	return podGroup, err
}

// DeletePodGroupForInstance deletes podgroup for instance.
func (v *VolcanoPodGroupManager) DeletePodGroupForInstance(
	ctx context.Context,
	instanceSet *v1.InstanceSet,
	indexer common.InstanceIndexer) error {
	// check if podgroup exist
	pgName := common.GetPGNameFromIndexer(indexer)
	podGroup, err := v.GetPodGroupForInstance(ctx, instanceSet, indexer)
	if errors.IsNotFound(err) {
		hwlog.RunLog.Infof("podgroup<%s> not exist, deletion skipped ", pgName)
		return nil
	}
	if err != nil {
		return err
	}

	// delete podgroup
	if err := v.client.Delete(ctx, podGroup); err != nil {
		hwlog.RunLog.Errorf("delete podgroup<%s> error: %v", pgName, err)
		return err
	}
	return nil
}

// GetOrCreatePodGroupForInstance gets or creates podgroup for instance.
func (v *VolcanoPodGroupManager) GetOrCreatePodGroupForInstance(
	ctx context.Context,
	instanceSet *v1.InstanceSet,
	indexer common.InstanceIndexer,
	spec v1beta1.PodGroupSpec) (bool, error) {
	_, err := v.GetPodGroupForInstance(ctx, instanceSet, indexer)
	if err != nil && !errors.IsNotFound(err) {
		return false, common.NewRequeueError(err.Error())
	}

	if errors.IsNotFound(err) {
		hwlog.RunLog.Infof("podgroup<%s> not exist, try to create", instanceSet.Name)
		return false, v.createPodGroupForInstance(ctx, instanceSet, indexer, spec)
	}
	return true, nil
}

func (v *VolcanoPodGroupManager) createPodGroupForInstance(
	ctx context.Context,
	instanceSet *v1.InstanceSet,
	indexer common.InstanceIndexer,
	spec v1beta1.PodGroupSpec) error {
	pgName := common.GetPGNameFromIndexer(indexer)
	labels := common.DeepCopyLabelsMap(instanceSet.Labels)
	labels = common.AddLabelsFromIndexer(labels, indexer)
	podGroup := &v1beta1.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:        pgName,
			Namespace:   instanceSet.Namespace,
			Annotations: instanceSet.Annotations,
			Labels:      labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(instanceSet, instanceSet.GroupVersionKind()),
			},
		},
		Spec: spec,
	}
	hwlog.RunLog.Infof("create podgroup<%s/%s> spec: MinMember=%d MinResources=%v",
		podGroup.Namespace, podGroup.Name, spec.MinMember, spec.MinResources)
	err := v.client.Create(ctx, podGroup)
	if err != nil {
		hwlog.RunLog.Errorf("create podgroup<%s> error: %v", instanceSet.Name, err)
	}
	return err
}

func NewVolcanoPodGroupManager(client client.Client) PodGroupManager {
	return &VolcanoPodGroupManager{
		client: client,
	}
}
