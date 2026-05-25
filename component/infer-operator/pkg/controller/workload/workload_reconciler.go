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

	"sigs.k8s.io/controller-runtime/pkg/client"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

type WorkLoadReconciler struct {
	workLoadHandlerFactory *WorkLoadHandlerFactory
	PodGroupManager
	client client.Client
}

// NewWorkLoadReconciler creates a new WorkLoadReconciler.
func NewWorkLoadReconciler(client client.Client) *WorkLoadReconciler {
	return &WorkLoadReconciler{
		workLoadHandlerFactory: NewWorkLoadHandlerFactory(),
		client:                 client,
		PodGroupManager:        NewVolcanoPodGroupManager(client),
	}
}

func (r *WorkLoadReconciler) SetWorkLoadHandlerFactory(factory *WorkLoadHandlerFactory) {
	r.workLoadHandlerFactory = factory
}

// Validate validates the InstanceSet.
func (r *WorkLoadReconciler) Validate(instanceSet *v1.InstanceSet) error {
	// 1. check WorkLoadHandler
	workloadHandler, err := r.GetWorkLoadReconciler(instanceSet)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to get workload handler: %v", err)
		return err
	}

	// 2. validate workload
	return workloadHandler.Validate(instanceSet.Spec.InstanceSpec)
}

// Reconcile reconciles the InstanceSet.
func (r *WorkLoadReconciler) Reconcile(
	ctx context.Context,
	instanceSet *v1.InstanceSet,
	indexer common.InstanceIndexer) error {
	// 1. check WorkLoadHandler
	workloadHandler, err := r.GetWorkLoadReconciler(instanceSet)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to get workload handler: %v", err)
		return err
	}

	// 2. check or create podgroup
	useGangScheduling := instanceSet.Labels[common.GangScheduleLabelKey] == common.TrueBool
	if useGangScheduling {
		workloadReplicas, err := workloadHandler.GetReplicas(instanceSet.Spec.InstanceSpec)
		if err != nil {
			hwlog.RunLog.Errorf("Failed to get workload replicas: %v", err)
			return err
		}
		podGroupSpec := newPodGroupSpec(workloadReplicas)
		_, err = r.PodGroupManager.GetOrCreatePodGroupForInstance(ctx, instanceSet, indexer, podGroupSpec)
		if err != nil {
			hwlog.RunLog.Errorf("Failed to create pod group: %v", err)
			return err
		}
	}

	// 3. reconcile workload
	err = workloadHandler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
	if err != nil {
		return err
	}
	return nil
}

// InstanceReady checks if instance is ready.
func (r *WorkLoadReconciler) InstanceReady(
	ctx context.Context,
	instanceSet *v1.InstanceSet,
	indexer common.InstanceIndexer) (int, error) {
	// 1. check WorkLoadHandler
	readyReplicas := 0
	workloadHandler, err := r.GetWorkLoadReconciler(instanceSet)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to get workload handler: %v", err)
		return readyReplicas, err
	}

	// 2. check if workloads is ready
	return workloadHandler.GetWorkLoadReadyReplicas(ctx, indexer)
}

// DeleteExtraInstances deletes extra instances.
func (r *WorkLoadReconciler) DeleteExtraInstances(
	ctx context.Context,
	instanceSet *v1.InstanceSet,
	indexer common.InstanceIndexer) error {
	// 1. check WorkLoadHandler
	workloadHandler, err := r.GetWorkLoadReconciler(instanceSet)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to get workload handler: %v", err)
		return err
	}

	// 2. delete extra workloads
	replicaNum := int(*instanceSet.Spec.Replicas)
	hwlog.RunLog.Infof("try to delete extra instances, replicaNum: %d", replicaNum)
	return workloadHandler.DeleteExtraWorkLoad(ctx, indexer, replicaNum)
}

// GetWorkLoadReconciler returns the WorkLoadHandler for the given InstanceSet
func (r *WorkLoadReconciler) GetWorkLoadReconciler(instanceSet *v1.InstanceSet) (WorkLoadHandler, error) {
	workloadType := instanceSet.Spec.WorkloadTypeMeta
	gvk, err := common.WorkLoadTypeToGVK(workloadType)
	if err != nil {
		return nil, err
	}
	workloadHandler, err := r.workLoadHandlerFactory.GetWorkLoadHandler(gvk)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}
	return workloadHandler, nil
}

func newPodGroupSpec(workloadReplicas int32) v1beta1.PodGroupSpec {
	return v1beta1.PodGroupSpec{
		MinMember: workloadReplicas,
	}
}
