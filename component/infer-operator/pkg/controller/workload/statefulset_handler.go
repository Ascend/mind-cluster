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
	"encoding/json"
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

type StatefulSetHandler struct {
	client client.Client
}

func NewStatefulSetHandler(client client.Client) *StatefulSetHandler {
	return &StatefulSetHandler{
		client: client,
	}
}

// CheckOrCreateWorkLoad checks if the statefulset exists and creates it if not
func (s *StatefulSetHandler) CheckOrCreateWorkLoad(
	ctx context.Context,
	instanceSet *v1.InstanceSet,
	indexer common.InstanceIndexer) error {
	// 1. fetch service
	service := &corev1.Service{}
	serviceNamespacedName := types.NamespacedName{
		Name:      common.GetServiceNameFromIndexer(indexer),
		Namespace: instanceSet.Namespace,
	}
	err := s.client.Get(ctx, serviceNamespacedName, service)
	if err != nil && !errors.IsNotFound(err) {
		hwlog.RunLog.Errorf("Failed to get service %s/%s: %v",
			instanceSet.Namespace, instanceSet.Name, err)
		return common.NewRequeueError(err.Error())
	}
	if errors.IsNotFound(err) {
		hwlog.RunLog.Infof("service of <%v> not exist, try to create", indexer)
		// 2. create service if not exist
		if err := s.createService(ctx, instanceSet, indexer); err != nil {
			return common.NewRequeueError(err.Error())
		}
	}
	// 3. fetch workload
	selectLabels := make(map[string]string)
	selectLabels = common.AddLabelsFromIndexer(selectLabels, indexer)
	statefulsetList, err := s.ListWorkLoads(ctx, selectLabels, indexer.Namespace)
	if err != nil {
		return err
	}
	// 4. create if not exist
	if len(statefulsetList.Items) == 0 {
		hwlog.RunLog.Infof("statefulset of <%v> not exist, try to create", indexer)
		err := s.createStatefulSet(ctx, instanceSet, indexer)
		if err != nil {
			return err
		}
	}
	// 5. check extra ones
	if len(statefulsetList.Items) > 1 {
		hwlog.RunLog.Warnf("More than one StatefulSet exists in InstanceSet<%s>", instanceSet.Name)
	}
	return nil
}

func (s *StatefulSetHandler) createStatefulSet(
	ctx context.Context,
	instanceSet *v1.InstanceSet,
	indexer common.InstanceIndexer) error {
	// 1. resolve statefulset spec
	statefulsetSpec, err := s.parseStatefulSetWithScheme(instanceSet.Spec.InstanceSpec)
	if err != nil {
		return err
	}
	// 2. add labels and annotations
	statefulsetLabels := common.DeepCopyLabelsMap(instanceSet.Labels)
	for k, v := range instanceSet.Spec.WorkloadObjectMeta.Labels {
		statefulsetLabels[k] = v
	}
	statefulsetLabels = common.AddLabelsFromIndexer(statefulsetLabels, indexer)
	statefulsetAnnotations := common.DeepCopyLabelsMap(instanceSet.Annotations)
	for k, v := range instanceSet.Spec.WorkloadObjectMeta.Annotations {
		statefulsetAnnotations[k] = v
	}

	statefulsetSpec.Template.Labels = common.AddLabelsFromIndexer(statefulsetSpec.Template.Labels, indexer)
	if statefulsetSpec.Template.Annotations == nil {
		statefulsetSpec.Template.Annotations = map[string]string{}
	}
	useGangScheduling := instanceSet.Labels[common.GangScheduleLabelKey] == common.TrueBool
	if useGangScheduling {
		statefulsetSpec.Template.Annotations[common.GroupNameAnnotationKey] = common.GetPGNameFromIndexer(indexer)
	}
	statefulsetSpec.ServiceName = common.GetServiceNameFromIndexer(indexer)
	common.AddEnvToPodTemplate(&statefulsetSpec.Template, indexer)
	// 3. create statefulset template
	newStatefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        common.GetWorkLoadNameFromIndexer(indexer),
			Namespace:   instanceSet.Namespace,
			Annotations: statefulsetAnnotations,
			Labels:      statefulsetLabels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(instanceSet, instanceSet.GroupVersionKind()),
			},
		},
		Spec: *statefulsetSpec,
	}
	// 4. create statefulset
	err = s.client.Create(ctx, newStatefulSet)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to create StatefulSet<%s>: %v", newStatefulSet.Name, err)
		return common.NewRequeueError(err.Error())
	}
	return nil
}

func (s *StatefulSetHandler) createService(
	ctx context.Context,
	instanceSet *v1.InstanceSet,
	indexer common.InstanceIndexer) error {
	labels := make(map[string]string)
	labels = common.AddLabelsFromIndexer(labels, indexer)
	newService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        common.GetServiceNameFromIndexer(indexer),
			Namespace:   instanceSet.Namespace,
			Annotations: instanceSet.Annotations,
			Labels:      labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(instanceSet, instanceSet.GroupVersionKind()),
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Selector:  labels,
			Ports: []corev1.ServicePort{
				{
					Name: common.DefaultPortName,
					Port: common.DefaultPort,
				},
			},
		},
	}
	err := s.client.Create(ctx, newService)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to create Service<%s>: %v", newService.Name, err)
		return common.NewRequeueError(err.Error())
	}
	return nil
}

// DeleteExtraWorkLoad deletes statefulsets that exceed the specified index limit
func (s *StatefulSetHandler) DeleteExtraWorkLoad(
	ctx context.Context,
	indexer common.InstanceIndexer, indexLimit int) error {
	// 1. fetch workload
	selectLabels := make(map[string]string)
	selectLabels = common.AddLabelsFromIndexer(selectLabels, indexer)
	delete(selectLabels, common.InstanceIndexLabelKey)
	statefulsetList, err := s.ListWorkLoads(ctx, selectLabels, indexer.Namespace)
	if err != nil {
		return err
	}

	// 2. delete workload if its instance-index >= indexLimit
	for _, statefulset := range statefulsetList.Items {
		instanceIndexStr, ok := statefulset.Labels[common.InstanceIndexLabelKey]
		if !ok {
			continue
		}
		instanceIndex, err := strconv.Atoi(instanceIndexStr)
		if err != nil {
			hwlog.RunLog.Warnf("StatefulSet<%s> Failed to convert instance index to int: %v",
				statefulset.Name, instanceIndexStr)
			// invalid workload, skip it
			continue
		}
		if instanceIndex < indexLimit && instanceIndex >= 0 {
			// normal range, keep it
			continue
		}
		if err = s.client.Delete(ctx, &statefulset); err != nil {
			hwlog.RunLog.Errorf("Failed to delete StatefulSet<%s>: %v", statefulset.Name, err)
			return err
		}
		hwlog.RunLog.Infof("Delete Extra StatefulSet<%s>", statefulset.Name)
	}
	// 3. delete extra services
	return s.deleteExtraService(ctx, selectLabels, indexLimit)
}

// GetWorkLoadReadyReplicas returns the number of ready replicas of the statefulset
func (s *StatefulSetHandler) GetWorkLoadReadyReplicas(
	ctx context.Context,
	indexer common.InstanceIndexer) (int, error) {
	// 1. fetch workload
	readyReplicas := 0
	selectLabels := make(map[string]string)
	selectLabels = common.AddLabelsFromIndexer(selectLabels, indexer)
	delete(selectLabels, common.InstanceIndexLabelKey)
	statefulsetList, err := s.ListWorkLoads(ctx, selectLabels, indexer.Namespace)
	if err != nil {
		return readyReplicas, err
	}

	// 2. get ready num
	for _, statefulset := range statefulsetList.Items {
		if isStatefulsetReady(statefulset) {
			readyReplicas++
		}
	}
	return readyReplicas, nil
}

func (s *StatefulSetHandler) deleteExtraService(
	ctx context.Context,
	selectLabels map[string]string,
	indexLimit int) error {
	// 1. fetch services
	serviceList := &corev1.ServiceList{}
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: selectLabels,
	})
	if err != nil {
		hwlog.RunLog.Errorf("Failed to convert label selector to selector: %v", err)
		return common.NewRequeueError(err.Error())
	}
	if err = s.client.List(ctx, serviceList, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		hwlog.RunLog.Errorf("Failed to list extra services: %v", err)
		return common.NewRequeueError(err.Error())
	}
	// 2. delete extra services
	for _, service := range serviceList.Items {
		instanceIndexStr, ok := service.Labels[common.InstanceIndexLabelKey]
		if !ok {
			continue
		}
		instanceIndex, err := strconv.Atoi(instanceIndexStr)
		if err != nil {
			hwlog.RunLog.Warnf("service<%s> Failed to convert instance index to int: %v",
				service.Name, instanceIndexStr)
			continue
		}
		if instanceIndex < indexLimit && instanceIndex >= 0 {
			// normal range, keep it
			continue
		}
		err = s.client.Delete(ctx, &service)
		if err != nil {
			hwlog.RunLog.Errorf("Failed to delete Extra Service<%s>: %v", service.Name, err)
			return common.NewRequeueError(err.Error())
		}
	}
	return nil
}

// ListWorkLoads lists deployments with the specified labels in the given namespace
func (s *StatefulSetHandler) ListWorkLoads(
	ctx context.Context,
	selectLabels map[string]string,
	namespace string) (*appsv1.StatefulSetList, error) {
	statefulsetList := &appsv1.StatefulSetList{}
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: selectLabels,
	})
	if err != nil {
		hwlog.RunLog.Errorf("Failed to convert label selector to selector: %v", err)
		return statefulsetList, common.NewRequeueError(err.Error())
	}
	if err = s.client.List(ctx, statefulsetList,
		client.MatchingLabelsSelector{Selector: selector}, client.InNamespace(namespace)); err != nil {
		hwlog.RunLog.Errorf("Failed to list extra statefulsets: %v", err)
		return nil, common.NewRequeueError(err.Error())
	}
	return statefulsetList, nil
}

// Validate checks if the statefulset specification is valid
func (s *StatefulSetHandler) Validate(spec runtime.RawExtension) error {
	_, err := s.parseStatefulSetWithScheme(spec)
	if err != nil {
		return err
	}
	return nil
}

// GetReplicas returns the number of replicas specified in the statefulset specification
func (s *StatefulSetHandler) GetReplicas(spec runtime.RawExtension) (int32, error) {
	statefulsetSpec, err := s.parseStatefulSetWithScheme(spec)
	if err != nil {
		return common.DefaultReplicas, err
	}

	replicas := statefulsetSpec.Replicas
	if replicas == nil {
		return common.DefaultReplicas, nil
	}
	return *replicas, nil
}

func isStatefulsetReady(sts appsv1.StatefulSet) bool {
	// 1. get desired replicas
	desiredReplicas := int32(1)
	if sts.Spec.Replicas != nil {
		desiredReplicas = *sts.Spec.Replicas
	}
	// 2. check if status is latest
	if sts.Generation > 0 && sts.Status.ObservedGeneration < sts.Generation {
		return false
	}
	// 3. check replicas number
	if sts.Status.ReadyReplicas != desiredReplicas ||
		sts.Status.UpdatedReplicas != desiredReplicas {
		return false
	}
	// 4. check revision (rollout update)
	if sts.Status.CurrentRevision != "" && sts.Status.UpdateRevision != "" &&
		sts.Status.CurrentRevision != sts.Status.UpdateRevision {
		// A rolling update is in progress, need to confirm all replicas are updated
		// Even if the number of replicas meets the requirement, it is not considered fully ready if still rolling
		return false
	}
	return true
}

func (s *StatefulSetHandler) parseStatefulSetWithScheme(raw runtime.RawExtension) (*appsv1.StatefulSetSpec, error) {
	if len(raw.Raw) == 0 {
		return nil, fmt.Errorf("raw extension is empty")
	}

	// decode raw spec of statefulset
	var spec appsv1.StatefulSetSpec
	if err := json.Unmarshal(raw.Raw, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal RawExtension to StatefulSetSpec: %w", err)
	}
	return &spec, nil
}
