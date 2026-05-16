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
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

type DeploymentHandler struct {
	client client.Client
}

func NewDeploymentHandler(client client.Client) *DeploymentHandler {
	return &DeploymentHandler{
		client: client,
	}
}

// CheckOrCreateWorkLoad checks if the deployment exists and creates it if not
func (d *DeploymentHandler) CheckOrCreateWorkLoad(ctx context.Context,
	instanceSet *v1.InstanceSet,
	indexer common.InstanceIndexer) error {
	// 1. reconcile service if no custom services
	if instanceSet.Spec.Services == nil || len(instanceSet.Spec.Services) == 0 {
		if err := d.checkOrCreateService(ctx, instanceSet, indexer); err != nil {
			return err
		}
	}

	// 2. fetch workload
	selectLabels := make(map[string]string)
	selectLabels = common.AddLabelsFromIndexer(selectLabels, indexer)
	deployList, err := d.ListWorkLoads(ctx, selectLabels, indexer.Namespace)
	if err != nil {
		return err
	}
	// 3. create if not exist
	if len(deployList.Items) == 0 {
		hwlog.RunLog.Infof("deployment of <%v> not exist, try to create", indexer)
		err := d.createDeployment(ctx, instanceSet, indexer)
		if err != nil {
			return err
		}
	}
	// 4. check extra ones
	if len(deployList.Items) > 1 {
		hwlog.RunLog.Warnf("More than one Deployment exists in InstanceSet<%s>", instanceSet.Name)
	}
	return nil
}

func (d *DeploymentHandler) checkOrCreateService(
	ctx context.Context,
	instanceSet *v1.InstanceSet,
	indexer common.InstanceIndexer) error {
	// 1. fetch service
	service := &corev1.Service{}
	serviceNamespacedName := types.NamespacedName{
		Name:      common.GetServiceNameFromIndexer(indexer),
		Namespace: instanceSet.Namespace,
	}
	err := d.client.Get(ctx, serviceNamespacedName, service)
	if err != nil && !errors.IsNotFound(err) {
		hwlog.RunLog.Errorf("Failed to get service %s/%s: %v",
			instanceSet.Namespace, instanceSet.Name, err)
		return common.NewRequeueError(err.Error())
	}
	if errors.IsNotFound(err) {
		hwlog.RunLog.Infof("service of <%v> not exist, try to create", indexer)
		// 2. create service if not exist
		if err := d.createService(ctx, instanceSet, indexer); err != nil {
			return common.NewRequeueError(err.Error())
		}
	}
	return nil
}

func (d *DeploymentHandler) createDeployment(
	ctx context.Context,
	instanceSet *v1.InstanceSet,
	indexer common.InstanceIndexer) error {
	// 1. resolve deployment spec
	deploymentSpec, err := d.parseDeploymentWithScheme(instanceSet.Spec.InstanceSpec)
	if err != nil {
		return err
	}

	// 2. add labels and annotations
	deployLabels := common.DeepCopyLabelsMap(instanceSet.Labels)
	for k, v := range instanceSet.Spec.WorkloadObjectMeta.Labels {
		deployLabels[k] = v
	}
	deployLabels = common.AddLabelsFromIndexer(deployLabels, indexer)
	deployAnnotations := common.DeepCopyLabelsMap(instanceSet.Annotations)
	for k, v := range instanceSet.Spec.WorkloadObjectMeta.Annotations {
		deployAnnotations[k] = v
	}

	deploymentSpec.Template.Labels = common.AddLabelsFromIndexer(deploymentSpec.Template.Labels, indexer)
	if deploymentSpec.Template.Annotations == nil {
		deploymentSpec.Template.Annotations = map[string]string{}
	}
	useGangScheduling := instanceSet.Labels[common.GangScheduleLabelKey] == common.TrueBool
	if useGangScheduling {
		deploymentSpec.Template.Annotations[common.GroupNameAnnotationKey] = common.GetPGNameFromIndexer(indexer)
	}
	common.AddEnvToPodTemplate(&deploymentSpec.Template, indexer)

	// 3. create deployment template
	newDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        common.GetWorkLoadNameFromIndexer(indexer),
			Namespace:   instanceSet.Namespace,
			Annotations: deployAnnotations,
			Labels:      deployLabels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(instanceSet, instanceSet.GroupVersionKind()),
			},
		},
		Spec: *deploymentSpec,
	}

	// 4. create deployment
	err = d.client.Create(ctx, newDeployment)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to create Deployment<%s>: %v", newDeployment.Name, err)
		return common.NewRequeueError(err.Error())
	}
	return nil
}

func (d *DeploymentHandler) createService(
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
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       common.DefaultPortName,
					Port:       common.DefaultPort,
					TargetPort: intstr.FromInt(common.DefaultPort),
				},
			},
		},
	}
	err := d.client.Create(ctx, newService)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to create Service<%s>: %v", newService.Name, err)
		return common.NewRequeueError(err.Error())
	}
	return nil
}

// DeleteExtraWorkLoad deletes deployments that exceed the specified index limit
func (d *DeploymentHandler) DeleteExtraWorkLoad(
	ctx context.Context,
	indexer common.InstanceIndexer, indexLimit int) error {
	// 1. fetch workload
	selectLabels := make(map[string]string)
	selectLabels = common.AddLabelsFromIndexer(selectLabels, indexer)
	delete(selectLabels, common.InstanceIndexLabelKey)
	hwlog.RunLog.Infof("try to delete extra instances, labels: %v", selectLabels)
	deployList, err := d.ListWorkLoads(ctx, selectLabels, indexer.Namespace)
	if err != nil {
		return err
	}

	// 2. delete workload if its instance-index >= indexLimit
	for _, deploy := range deployList.Items {
		instanceIndexStr, ok := deploy.Labels[common.InstanceIndexLabelKey]
		if !ok {
			continue
		}
		instanceIndex, err := strconv.Atoi(instanceIndexStr)
		if err != nil {
			hwlog.RunLog.Warnf("Deployment<%s> Failed to convert instance index to int: %v",
				deploy.Name, instanceIndexStr)
			// invalid workload, skip it
			continue
		}
		if instanceIndex < indexLimit && instanceIndex >= 0 {
			// normal range, keep it
			continue
		}
		if err = d.client.Delete(ctx, &deploy); err != nil {
			hwlog.RunLog.Errorf("Failed to delete Deployment<%s>: %v", deploy.Name, err)
			return err
		}
		hwlog.RunLog.Infof("Delete Extra Deployment<%s>", deploy.Name)
	}
	return d.deleteExtraService(ctx, selectLabels, indexLimit)
}

// GetWorkLoadReadyReplicas returns the number of ready replicas of the deployment
func (d *DeploymentHandler) GetWorkLoadReadyReplicas(
	ctx context.Context,
	indexer common.InstanceIndexer) (int, error) {
	// 1. fetch workload
	readyReplicas := 0
	selectLabels := make(map[string]string)
	selectLabels = common.AddLabelsFromIndexer(selectLabels, indexer)
	delete(selectLabels, common.InstanceIndexLabelKey)
	deployList, err := d.ListWorkLoads(ctx, selectLabels, indexer.Namespace)
	if err != nil {
		return readyReplicas, err
	}

	// 2. get ready num
	for _, deployment := range deployList.Items {
		if isDeploymentReady(deployment) {
			readyReplicas++
		}
	}
	return readyReplicas, nil
}

func (d *DeploymentHandler) deleteExtraService(
	ctx context.Context,
	selectLabels map[string]string,
	indexLimit int) error {
	// 1. fetch services
	serviceList := &corev1.ServiceList{}
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: selectLabels,
	})
	if err != nil {
		hwlog.RunLog.Errorf("Failed to create ServiceList<%s>: %v", selectLabels, err)
		return common.NewRequeueError(err.Error())
	}
	if err = d.client.List(ctx, serviceList, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		hwlog.RunLog.Errorf("Failed to list ServiceList<%s>: %v", selectLabels, err)
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
			// invalid workload, skip it
			continue
		}
		if instanceIndex < indexLimit && instanceIndex >= 0 {
			// normal range, keep it
			continue
		}
		err = d.client.Delete(ctx, &service)
		if err != nil {
			hwlog.RunLog.Errorf("Failed to delete Extra Service<%s>: %v", service.Name, err)
			return common.NewRequeueError(err.Error())
		}
	}
	return nil
}

// ListWorkLoads lists deployments with the specified labels in the given namespace
func (d *DeploymentHandler) ListWorkLoads(
	ctx context.Context,
	selectLabels map[string]string,
	namespace string) (*appsv1.DeploymentList, error) {
	deployList := &appsv1.DeploymentList{}
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: selectLabels,
	})
	if err != nil {
		hwlog.RunLog.Errorf("Failed to create selector: %v", err)
		return nil, common.NewRequeueError(err.Error())
	}
	if err = d.client.List(ctx, deployList,
		client.MatchingLabelsSelector{Selector: selector}, client.InNamespace(namespace)); err != nil {
		hwlog.RunLog.Errorf("Failed to list Deployments: %v", err)
		return nil, common.NewRequeueError(err.Error())
	}
	return deployList, nil
}

// Validate checks if the deployment specification is valid
func (d *DeploymentHandler) Validate(spec runtime.RawExtension) error {
	_, err := d.parseDeploymentWithScheme(spec)
	if err != nil {
		return err
	}
	return nil
}

// GetReplicas returns the number of replicas specified in the deployment specification
func (d *DeploymentHandler) GetReplicas(spec runtime.RawExtension) (int32, error) {
	deploymentSpec, err := d.parseDeploymentWithScheme(spec)
	if err != nil {
		return common.DefaultReplicas, err
	}

	replicas := deploymentSpec.Replicas
	if replicas == nil {
		return common.DefaultReplicas, nil
	}
	return *replicas, nil
}

func isDeploymentReady(deployment appsv1.Deployment) bool {
	// 1. get desired replicas
	desiredReplicas := int32(1)
	if deployment.Spec.Replicas != nil {
		desiredReplicas = *deployment.Spec.Replicas
	}
	// 2. check if status is latest
	if deployment.Generation > 0 && deployment.Status.ObservedGeneration < deployment.Generation {
		hwlog.RunLog.Warnf("Deployment %s/%s is not latest", deployment.Namespace, deployment.Name)
		return false
	}
	// 3. check replicas number
	if deployment.Status.ReadyReplicas != desiredReplicas ||
		deployment.Status.AvailableReplicas != desiredReplicas ||
		deployment.Status.UpdatedReplicas != desiredReplicas {
		return false
	}
	// 4. check conditions
	available := getDeploymentCondition(deployment.Status.Conditions, appsv1.DeploymentAvailable)
	progressing := getDeploymentCondition(deployment.Status.Conditions, appsv1.DeploymentProgressing)
	if available == nil || available.Status != corev1.ConditionTrue {
		hwlog.RunLog.Warnf("Deployment %s/%s is not available, Condition<%s> is not true",
			deployment.Namespace, deployment.Name, appsv1.DeploymentAvailable)
		return false
	}
	if progressing == nil || progressing.Status != corev1.ConditionTrue {
		hwlog.RunLog.Warnf("Deployment %s/%s is not progressing, Condition<%s> is not true",
			deployment.Namespace, deployment.Name, appsv1.DeploymentAvailable)
		return false
	}
	return true
}

func getDeploymentCondition(
	conditions []appsv1.DeploymentCondition,
	condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}

func (d *DeploymentHandler) parseDeploymentWithScheme(raw runtime.RawExtension) (*appsv1.DeploymentSpec, error) {
	if len(raw.Raw) == 0 {
		return nil, fmt.Errorf("raw extension is empty")
	}

	// decode raw spec of deployment
	var spec appsv1.DeploymentSpec
	if err := json.Unmarshal(raw.Raw, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal RawExtension to DeploymentSpec: %w", err)
	}
	return &spec, nil
}
