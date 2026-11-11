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
	"infer-operator/pkg/common/utils"
)

// DeploymentWorkLoad implements WorkLoad interface for the deployment
type DeploymentHandler struct {
	client client.Client
}

func NewDeploymentHandler(client client.Client) *DeploymentHandler {
	return &DeploymentHandler{
		client: client,
	}
}

type DeploymentWorkLoad struct {
	*appsv1.Deployment
}

// SetWorkLoadObjMeta set the object meta of the deployment
func (d *DeploymentWorkLoad) SetWorkLoadObjMeta(objectMeta metav1.ObjectMeta) {
	if d == nil {
		return
	}
	d.ObjectMeta = objectMeta
}

// GetWorkLoadObjMeta get the object meta of the workload
func (d *DeploymentWorkLoad) GetWorkLoadObjMeta() metav1.ObjectMeta {
	if d == nil {
		return metav1.ObjectMeta{}
	}
	return d.ObjectMeta
}

// IsWorkLoadReady returns true if the workload is ready
func (d *DeploymentWorkLoad) IsWorkLoadReady() bool {
	if d == nil {
		return false
	}
	// 1. get desired replicas
	desiredReplicas := common.DefaultReplicas
	if d.Spec.Replicas != nil {
		desiredReplicas = *d.Spec.Replicas
	}
	// 2. check if status is latest
	if d.Generation > 0 && d.Status.ObservedGeneration < d.Generation {
		hwlog.RunLog.Warnf("Deployment %s/%s is not latest", d.Namespace, d.Name)
		return false
	}
	// 3. check replicas number
	if d.Status.ReadyReplicas != desiredReplicas ||
		d.Status.AvailableReplicas != desiredReplicas ||
		d.Status.UpdatedReplicas != desiredReplicas {
		return false
	}
	// 4. check conditions
	available := getDeploymentCondition(d.Status.Conditions, appsv1.DeploymentAvailable)
	progressing := getDeploymentCondition(d.Status.Conditions, appsv1.DeploymentProgressing)
	if available == nil || available.Status != corev1.ConditionTrue {
		hwlog.RunLog.Warnf("Deployment %s/%s is not available, Condition<%s> is not true",
			d.Namespace, d.Name, appsv1.DeploymentAvailable)
		return false
	}
	if progressing == nil || progressing.Status != corev1.ConditionTrue {
		hwlog.RunLog.Warnf("Deployment %s/%s is not progressing, Condition<%s> is not true",
			d.Namespace, d.Name, appsv1.DeploymentAvailable)
		return false
	}
	return true
}

// GetWorkLoadReplicas returns the number of replicas of the deployment
func (d *DeploymentWorkLoad) GetWorkLoadReplicas() int32 {
	if d == nil {
		return common.DefaultReplicas
	}
	replicas := d.Spec.Replicas
	if replicas == nil {
		return common.DefaultReplicas
	}
	return *replicas
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
	faultScheduling, ok := deploymentSpec.Template.Labels[common.FaultSchedulingLabelKey]
	if ok {
		deployLabels[common.FaultSchedulingLabelKey] = faultScheduling
	}
	deployAnnotations := common.DeepCopyLabelsMap(instanceSet.Annotations)
	for k, v := range instanceSet.Spec.WorkloadObjectMeta.Annotations {
		deployAnnotations[k] = v
	}

	deploymentSpec.Template.Labels = common.AddLabelsFromIndexer(deploymentSpec.Template.Labels, indexer)
	common.AddInferServiceIDToPodTemplate(instanceSet.Labels, &deploymentSpec.Template)
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
	if newDeployment.Spec.Replicas == nil {
		return fmt.Errorf("replicas is nil")
	}
	replicas := *newDeployment.Spec.Replicas
	hwlog.RunLog.Infof("create deployment<%s/%s> replicas=%d",
		newDeployment.Namespace, newDeployment.Name, replicas)
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

// DeleteExtraWorkLoad deletes the deployments that exceed the specified index limit.
// Instances with a negative index are deleted first, then unready instances, then ready
// instances, each group in descending index order, until the shrink count is reached.
// Instances with a missing or illegal index label are kept.
func (d *DeploymentHandler) DeleteExtraWorkLoad(
	ctx context.Context,
	indexer common.InstanceIndexer, indexLimit int) error {
	// 1. fetch workloads
	selectLabels := make(map[string]string)
	selectLabels = common.AddLabelsFromIndexer(selectLabels, indexer)
	delete(selectLabels, common.InstanceIndexLabelKey)
	hwlog.RunLog.Infof("try to delete extra instances, labels: %v", selectLabels)
	deployList, err := d.ListWorkLoads(ctx, selectLabels, indexer.Namespace)
	if err != nil {
		return err
	}

	// 2. select the indices to delete
	workLoads := make([]WorkLoadInterface, 0, len(deployList.Items))
	indexedWorkLoads := make(map[int][]*DeploymentWorkLoad, len(deployList.Items))
	for i := range deployList.Items {
		deploy := &DeploymentWorkLoad{&deployList.Items[i]}
		workLoads = append(workLoads, deploy)
		if index, ok := parseInstanceIndex(deploy.GetWorkLoadObjMeta()); ok {
			indexedWorkLoads[index] = append(indexedWorkLoads[index], deploy)
		}
	}
	deletedIndices := selectIndicesToDelete(workLoads, indexLimit)

	// 3. delete the selected deployments in descending index order
	for _, index := range deletedIndices {
		for _, deploy := range indexedWorkLoads[index] {
			if index >= 0 && !deploy.IsWorkLoadReady() {
				hwlog.RunLog.Warnf("delete unready instance index<%d> of Deployment<%s> during scale-down, "+
					"desiredReplicas=%d", index, deploy.Name, deploy.GetWorkLoadReplicas())
			}
			if err = d.client.Delete(ctx, deploy.Deployment); err != nil {
				hwlog.RunLog.Errorf("Failed to delete Deployment<%s>: %v", deploy.Name, err)
				return err
			}
			hwlog.RunLog.Infof("Delete Extra Deployment<%s>", deploy.Name)
		}
		indexedWorkLoads[index] = nil
	}

	// 4. delete the services of the instances that no longer exist. The keep set is derived from
	// the remaining workloads rather than from the current deletion set, so a service whose
	// workload has already been deleted is retried on the next round, and orphan services without
	// any workload are cleaned up as well.
	keepIndices := make([]int, 0, len(indexedWorkLoads))
	for index, workloads := range indexedWorkLoads {
		if len(workloads) > 0 {
			keepIndices = append(keepIndices, index)
		}
	}
	return d.deleteServicesNotInIndices(ctx, selectLabels, indexer.Namespace, keepIndices)
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

// deleteServicesNotInIndices deletes the services whose instance index is not in keepIndices.
// The keep set comes from the workloads that survive this round, which makes the cleanup
// retryable: a service whose workload was already deleted is still selected, and services
// without any workload (orphans) are removed as well.
func (d *DeploymentHandler) deleteServicesNotInIndices(
	ctx context.Context,
	selectLabels map[string]string,
	namespace string,
	keepIndices []int) error {
	// 1. fetch services
	serviceList := &corev1.ServiceList{}
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: selectLabels,
	})
	if err != nil {
		hwlog.RunLog.Errorf("Failed to create ServiceList<%s>: %v", selectLabels, err)
		return common.NewRequeueError(err.Error())
	}
	if err = d.client.List(ctx, serviceList,
		client.MatchingLabelsSelector{Selector: selector}, client.InNamespace(namespace)); err != nil {
		hwlog.RunLog.Errorf("Failed to list ServiceList<%s>: %v", selectLabels, err)
		return common.NewRequeueError(err.Error())
	}
	// 2. delete the services whose instance index is not in the keep set
	keepIndexSet := make(map[int]struct{}, len(keepIndices))
	for _, index := range keepIndices {
		keepIndexSet[index] = struct{}{}
	}
	for _, service := range serviceList.Items {
		index, ok := parseInstanceIndex(service.ObjectMeta)
		if !ok {
			continue
		}
		if _, keep := keepIndexSet[index]; keep {
			continue
		}
		if err = d.client.Delete(ctx, &service); err != nil {
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

// GetMinResources calculates the minimal resources required by the deployment
// based on its pod template and replicas.
func (d *DeploymentHandler) GetMinResources(spec runtime.RawExtension) (*corev1.ResourceList, error) {
	deploymentSpec, err := d.parseDeploymentWithScheme(spec)
	if err != nil {
		return nil, err
	}
	if deploymentSpec.Replicas == nil {
		return nil, fmt.Errorf("replicas is nil")
	}
	replicas := *deploymentSpec.Replicas
	return utils.CalcMinResources(replicas, deploymentSpec.Template.Spec), nil
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

// ListWorkLoad list workloads via selector
func (d *DeploymentHandler) ListWorkLoad(
	ctx context.Context,
	selectLabels map[string]string,
	namespace string,
	filters ...WorkLoadFilter) ([]WorkLoadInterface, error) {
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
	deploymentWorkLoadList := make([]WorkLoadInterface, 0, len(deployList.Items))
	for _, deploy := range deployList.Items {
		deployCopy := deploy
		ok := true
		deploymentWorkLoad := &DeploymentWorkLoad{Deployment: &deployCopy}
		for _, filter := range filters {
			ok = ok && filter(deploymentWorkLoad)
			if !ok {
				break
			}
		}
		if ok {
			deploymentWorkLoadList = append(deploymentWorkLoadList, deploymentWorkLoad)
		}
	}
	return deploymentWorkLoadList, nil
}

// DeleteWorkLoad fetches workloads via selector and deletes those filtered by filters, with the timeout for graceful deletion
func (d *DeploymentHandler) DeleteWorkLoad(
	ctx context.Context,
	selectLabels map[string]string,
	namespace string,
	filters ...WorkLoadFilter) error {
	deployList, err := d.ListWorkLoads(ctx, selectLabels, namespace)
	if err != nil {
		return fmt.Errorf("failed to list deployment work loads: %w", err)
	}
	var workloadList []*DeploymentWorkLoad
	for _, deploy := range deployList.Items {
		ok := true
		deployCopy := deploy
		workload := &DeploymentWorkLoad{Deployment: &deployCopy}
		for _, filter := range filters {
			ok = ok && filter(workload)
			if !ok {
				break
			}
		}
		if ok {
			workloadList = append(workloadList, workload)
		}
	}
	for _, workload := range workloadList {
		if err := d.client.Delete(ctx, workload.Deployment); err != nil {
			return fmt.Errorf("failed to delete deployment work load %s/%s: %w", workload.Namespace, workload.Name, err)
		}
		if err := deletePodsForExternalRescheduling(ctx, d.client, workload); err != nil {
			return err
		}
	}
	return nil
}

// UpdateWorkLoad updates workloads match selector and filters with updater function
func (d *DeploymentHandler) UpdateWorkLoad(
	ctx context.Context,
	selectLabels map[string]string,
	namespace string,
	updater WorkloadUpdater,
	filters ...WorkLoadFilter) error {
	deployList, err := d.ListWorkLoads(ctx, selectLabels, namespace)
	if err != nil {
		return fmt.Errorf("failed to list deployment work loads: %w", err)
	}
	var workloadList []*DeploymentWorkLoad
	for _, deploy := range deployList.Items {
		ok := true
		deployCopy := deploy
		workload := &DeploymentWorkLoad{Deployment: &deployCopy}
		for _, filter := range filters {
			ok = ok && filter(workload)
			if !ok {
				break
			}
		}
		if ok {
			workloadList = append(workloadList, workload)
		}
	}
	for _, workload := range workloadList {
		updater(workload)
		if err := d.client.Update(ctx, workload.Deployment); err != nil {
			return fmt.Errorf("failed to update deployment work load %s/%s: %w", workload.Namespace, workload.Name, err)
		}
	}
	return nil
}
