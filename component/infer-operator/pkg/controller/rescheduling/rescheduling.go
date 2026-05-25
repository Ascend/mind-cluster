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

package rescheduling

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"ascend-common/common-utils/hwlog"
	apiv1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
	"infer-operator/pkg/controller/workload"
)

// Rescheduler manages infer operator rescheduling
type Rescheduler struct {
	client                 client.Client
	workLoadHandlerFactory *workload.WorkLoadHandlerFactory
	cleanupInterval        time.Duration
	faultWorkLoadRecord
}

// faultWorkLoadRecord records workloads that have fault pod and retry times
type faultWorkLoadRecord struct {
	sync.Mutex
	faultWorkLoadMap   map[faultWorkLoad]string
	faultRetryTimesMap map[faultWorkLoad]int
}

type faultWorkLoad struct {
	// workload namespaced name
	types.NamespacedName
	// instanceSet name
	instanceSetName string
}

func NewRescheduler(client client.Client, cleanupInterval time.Duration) *Rescheduler {
	return &Rescheduler{
		client:          client,
		cleanupInterval: cleanupInterval,
		faultWorkLoadRecord: faultWorkLoadRecord{
			faultWorkLoadMap:   make(map[faultWorkLoad]string),
			faultRetryTimesMap: make(map[faultWorkLoad]int),
			Mutex:              sync.Mutex{},
		},
	}
}

func (r *Rescheduler) SetWorkLoadHandlerFactory(factory *workload.WorkLoadHandlerFactory) {
	r.workLoadHandlerFactory = factory
}

func (r *Rescheduler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	podInformer, err := mgr.GetCache().GetInformer(context.Background(), &corev1.Pod{})
	if err != nil {
		return err
	}
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: r.handlePodUpdate,
	})
	return nil
}

func (r *Rescheduler) CleanupWithInstanceSetDeletion(instanceSetName string) {
	r.Lock()
	defer r.Unlock()
	hwlog.RunLog.Infof("Performing cleanup fault retry times map with instanceSet deletion")
	for currentFaultWorkLoad, _ := range r.faultRetryTimesMap {
		if currentFaultWorkLoad.instanceSetName == instanceSetName {
			delete(r.faultRetryTimesMap, currentFaultWorkLoad)
		}
	}
	for currentFaultWorkLoad, _ := range r.faultWorkLoadMap {
		if currentFaultWorkLoad.instanceSetName == instanceSetName {
			delete(r.faultWorkLoadMap, currentFaultWorkLoad)
		}
	}
}

func (r *Rescheduler) handlePodUpdate(oldObj, newObj interface{}) {
	pod, ok := newObj.(*corev1.Pod)
	if !ok {
		return
	}
	if !r.isValidFaultPod(pod) {
		return
	}
	hwlog.RunLog.Debugf("pod %s/%s is a valid fault pod, start to record fault", pod.Namespace, pod.Name)
	err := r.processFaultEvent(pod)
	if err != nil {
		hwlog.RunLog.Errorf("failed to record fault for pod %s/%s: %v", pod.Namespace, pod.Name, err)
	}
}

func (r *Rescheduler) isValidFaultPod(pod *corev1.Pod) bool {
	if !r.isValidInferPod(pod) {
		return false
	}
	// pod status must be unhealthy
	podStatus, exists := pod.Annotations[common.PodStatusAnnotationKey]
	if !exists || !strings.HasPrefix(podStatus, common.CommonUnhealthyStatus) {
		hwlog.RunLog.Infof("pod %s/%s has no unhealthy status, skip it", pod.Namespace, pod.Name)
		return false
	}
	// business fault must have retryTimes setting
	if strings.HasSuffix(podStatus, common.PodFailed) {
		retryTimeStr, exists := pod.Labels[common.FaultRetryTimesLabelKey]
		if !exists {
			hwlog.RunLog.Infof("pod %s/%s has business fault but no faultRetryTimes label", pod.Namespace, pod.Name)
			return false
		}
		retryTimes, err := strconv.Atoi(retryTimeStr)
		if err != nil || retryTimes < 0 {
			hwlog.RunLog.Errorf("pod %s/%s has business fault but retryTimes setting is invalid", pod.Namespace, pod.Name)
			return false
		}
	}
	// pod is being deleted
	if !(pod.DeletionTimestamp == nil || pod.DeletionTimestamp.IsZero()) {
		hwlog.RunLog.Infof("pod %s/%s is being deleted, skip it", pod.Namespace, pod.Name)
		return false
	}
	return true
}

func (r *Rescheduler) isValidInferPod(pod *corev1.Pod) bool {
	if pod.Labels == nil {
		hwlog.RunLog.Infof("pod %s/%s has no labels, skip it", pod.Namespace, pod.Name)
		return false
	}
	isInfer, exists := pod.Labels[common.OperatorNameKey]
	if !exists || isInfer != common.TrueBool {
		hwlog.RunLog.Infof("pod %s/%s is not a infer operator pod, skip it", pod.Namespace, pod.Name)
		return false
	}
	inferServiceName, exists := pod.Labels[common.InferServiceNameLabelKey]
	if !exists || inferServiceName == "" {
		hwlog.RunLog.Infof("pod %s/%s has no inferServiceName label, skip it", pod.Namespace, pod.Name)
		return false
	}
	instanceSetName, exists := pod.Labels[common.InstanceSetNameLabelKey]
	if !exists || instanceSetName == "" {
		hwlog.RunLog.Infof("pod %s/%s has no instanceSetName label, skip it", pod.Namespace, pod.Name)
		return false
	}
	instanceSetIndex, exists := pod.Labels[common.InstanceIndexLabelKey]
	if !exists || instanceSetIndex == "" {
		hwlog.RunLog.Infof("pod %s/%s has no instanceSetIndex label, skip it", pod.Namespace, pod.Name)
		return false
	}
	return true
}

func (r *Rescheduler) processFaultEvent(pod *corev1.Pod) error {
	// 1. get workload name and instance set name from pod
	workLoadName, instanceSetName := r.getWorkLoadNameAndInstanceSetName(pod)
	// 2. record fault for workload
	done := r.recordWorkLoadFault(pod, workLoadName, instanceSetName)
	if done {
		return nil
	}
	ctx := context.Background()
	var instanceSet apiv1.InstanceSet
	instanceSetNamespacedName := types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      instanceSetName,
	}
	err := r.client.Get(ctx, instanceSetNamespacedName, &instanceSet)
	if err != nil {
		return fmt.Errorf("failed to get instance set %s/%s: %v, rescheduling may not work",
			instanceSetNamespacedName.Namespace, instanceSetNamespacedName.Name, err)
	}
	// 3. trigger instanceSet reconcile
	err = r.triggerInstanceSetReconcile(ctx, &instanceSet, pod, workLoadName)
	if err != nil {
		return fmt.Errorf("failed to trigger instance set reconcile for pod %s/%s: %v, rescheduling may not work",
			pod.Namespace, pod.Name, err)
	}
	return nil
}

func (r *Rescheduler) recordWorkLoadFault(pod *corev1.Pod, workLoadName string, instanceSetName string) bool {
	workLoadNamespacedName := types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      workLoadName,
	}
	currentFaultWorkLoad := faultWorkLoad{
		NamespacedName:  workLoadNamespacedName,
		instanceSetName: instanceSetName,
	}
	r.Lock()
	// if a workload has multi faults, only process the first fault to reschedule workload
	_, exists := r.faultWorkLoadMap[currentFaultWorkLoad]
	if exists {
		hwlog.RunLog.Infof("pod %s/%s belongs to workload %s/%s which is already recorded for fault",
			pod.Namespace, pod.Name, workLoadNamespacedName.Namespace, workLoadName)
		return true
	}
	r.faultWorkLoadMap[currentFaultWorkLoad] = pod.Annotations[common.PodStatusAnnotationKey]
	if strings.HasSuffix(pod.Annotations[common.PodStatusAnnotationKey], common.PodFailed) {
		if _, exists := r.faultRetryTimesMap[currentFaultWorkLoad]; !exists {
			retryTimes, _ := strconv.Atoi(pod.Labels[common.FaultRetryTimesLabelKey])
			r.faultRetryTimesMap[currentFaultWorkLoad] = retryTimes
		}
	}
	r.Unlock()
	hwlog.RunLog.Infof("record fault: %s for workload %s/%s",
		pod.Annotations[common.PodStatusAnnotationKey], pod.Namespace, workLoadName)
	return false
}

func (r *Rescheduler) getWorkLoadNameAndInstanceSetName(pod *corev1.Pod) (string, string) {
	inferServiceName := pod.Labels[common.InferServiceNameLabelKey]
	instanceSetName := pod.Labels[common.InstanceSetNameLabelKey]
	instanceSetIndex := pod.Labels[common.InstanceIndexLabelKey]
	workLoadName := fmt.Sprintf("%s-%s-%s", inferServiceName, instanceSetName, instanceSetIndex)
	instanceSetName = fmt.Sprintf("%s-%s", inferServiceName, instanceSetName)
	return workLoadName, instanceSetName
}

// triggerInstanceSetReconcile trigger instanceSet reconcile by modifying instanceSet annotation
func (r *Rescheduler) triggerInstanceSetReconcile(
	ctx context.Context,
	instanceSet *apiv1.InstanceSet,
	pod *corev1.Pod,
	workloadName string) error {
	workloadGVK, err := common.WorkLoadTypeToGVK(instanceSet.Spec.WorkloadTypeMeta)
	if err != nil {
		return err
	}
	workloadHandler, err := r.workLoadHandlerFactory.GetWorkLoadHandler(workloadGVK)
	if err != nil {
		return fmt.Errorf("failed to get workLoadHandler for %s/%s", workloadGVK.Group, workloadGVK.Version)
	}
	updater := func(workLoad workload.WorkLoadInterface) {
		objMeta := workLoad.GetWorkLoadObjMeta()
		if objMeta.Annotations == nil {
			objMeta.Annotations = make(map[string]string)
		}
		objMeta.Annotations[common.DeletingTriggerAnnotationKey] = common.TrueBool
		workLoad.SetWorkLoadObjMeta(objMeta)
	}
	indexer := common.InstanceIndexer{
		Namespace:      pod.Namespace,
		ServiceName:    pod.Labels[common.InferServiceNameLabelKey],
		InstanceSetKey: pod.Labels[common.InstanceSetNameLabelKey],
		InstanceIndex:  pod.Labels[common.InstanceIndexLabelKey],
	}
	selectLabels := make(map[string]string)
	selectLabels = common.AddLabelsFromIndexer(selectLabels, indexer)
	if err := workloadHandler.UpdateWorkLoad(ctx, selectLabels, pod.Namespace, updater); err != nil {
		return fmt.Errorf("failed to update workload %s/%s: %v", pod.Namespace, workloadName, err)
	}
	return nil
}

func (r *Rescheduler) DoRescheduling(
	ctx context.Context,
	instanceSet *apiv1.InstanceSet) ([]apiv1.InstanceSet, error) {
	return nil, nil
}
