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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
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
		DeleteFunc: r.handlePodDelete,
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

// handlePodDelete handles pod delete event (both grace and force delete).
// DeleteFunc is the only reliable signal: for force-delete the update event
// (DeletionTimestamp) may be lost, and for grace-delete isValidFaultPod skips
// pods with DeletionTimestamp set. A single STS/Deployment's replicas together
// form one inference instance (a communication domain), so losing any pod
// requires rebuilding that pod's workload. Other workloads under the same
// instanceSet are unaffected. When pod-level rescheduling is enabled on the
// instanceSet, the deleted pod is rebuilt by K8s controllers instead.
func (r *Rescheduler) handlePodDelete(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}
	if !r.isValidInferPod(pod) {
		return
	}
	// pod-level rescheduling: the deleted pod is rebuilt by K8s controllers
	// (StatefulSet/Deployment), and the fault retry counter must only be
	// consumed by fault events, so skip re-entering the rescheduling flow on
	// pod delete events to avoid double-decrementing faultRetryTimesMap.
	if r.isPodLevelRescheduling(pod) {
		hwlog.RunLog.Infof("pod %s/%s deleted, pod-level rescheduling is enabled, "+
			"k8s controller will rebuild it, skip workload rebuild", pod.Namespace, pod.Name)
		return
	}
	// only trigger workload rebuild when gang scheduling is configured on the
	// instanceSet. A single STS/Deployment's replicas together form one
	// inference instance (a communication domain); without gang scheduling,
	// losing one pod does not require rebuilding the whole workload.
	if !r.isGangScheduled(pod) {
		hwlog.RunLog.Infof("pod %s/%s deleted, but gang scheduling is not configured, skip workload rebuild",
			pod.Namespace, pod.Name)
		return
	}
	hwlog.RunLog.Infof("pod %s/%s deleted, trigger its workload rebuild",
		pod.Namespace, pod.Name)
	if err := r.processFaultEvent(pod); err != nil {
		hwlog.RunLog.Errorf("failed to process delete for pod %s/%s: %v",
			pod.Namespace, pod.Name, err)
	}
}

// isGangScheduled checks whether gang scheduling is enabled on the instanceSet
// that the pod belongs to. The check follows the same convention as
// statefulset_handler/deployment_handler: instanceSet.Labels[GangScheduleLabelKey] == "true".
func (r *Rescheduler) isGangScheduled(pod *corev1.Pod) bool {
	_, instanceSetName := r.getWorkLoadNameAndInstanceSetName(pod)
	instanceSetNamespacedName := types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      instanceSetName,
	}
	var instanceSet apiv1.InstanceSet
	if err := r.client.Get(context.Background(), instanceSetNamespacedName, &instanceSet); err != nil {
		hwlog.RunLog.Errorf("failed to get instanceSet %s/%s when checking gang schedule: %v",
			instanceSetNamespacedName.Namespace, instanceSetNamespacedName.Name, err)
		return false
	}
	return instanceSet.Labels[common.GangScheduleLabelKey] == common.TrueBool
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
	// 2. read the pod-rescheduling label to decide deletion granularity
	isPodLevel := r.isPodLevelRescheduling(pod)
	if isPodLevel {
		// pod-level deletion: check FaultRetryTimes, delete the fault pod, decrement counter
		return r.processPodLevelRescheduling(pod, workLoadName, instanceSetName)
	}
	// 3. instance-level deletion: keep original logic (record fault + delete workload)
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
	// 4. trigger instanceSet reconcile
	err = r.triggerInstanceSetReconcile(ctx, &instanceSet, pod, workLoadName)
	if err != nil {
		return fmt.Errorf("failed to trigger instance set reconcile for pod %s/%s: %v, rescheduling may not work",
			pod.Namespace, pod.Name, err)
	}
	return nil
}

// processPodLevelRescheduling handles pod-level rescheduling: check FaultRetryTimes,
// delete the fault pod only, and decrement the retry counter after each deletion.
func (r *Rescheduler) processPodLevelRescheduling(pod *corev1.Pod,
	workLoadName, instanceSetName string) error {
	workLoadNamespacedName := types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      workLoadName,
	}
	currentFaultWorkLoad := faultWorkLoad{
		NamespacedName:  workLoadNamespacedName,
		instanceSetName: instanceSetName,
	}
	r.Lock()
	// retry times only apply to business faults (pod-failed); hardware faults delete
	// the fault pod directly without consuming the retry budget.
	if faultReason := pod.Annotations[common.PodStatusAnnotationKey]; strings.HasSuffix(faultReason, common.PodFailed) {
		// record initial retry times on the first fault
		if _, exists := r.faultRetryTimesMap[currentFaultWorkLoad]; !exists {
			retryTimes, err := strconv.Atoi(pod.Labels[common.FaultRetryTimesLabelKey])
			if err != nil {
				hwlog.RunLog.Errorf("pod-level rescheduling: invalid fault-retry-times label %q for workload %s/%s: %v",
					pod.Labels[common.FaultRetryTimesLabelKey], workLoadNamespacedName.Namespace, workLoadName, err)
				retryTimes = 0
			}
			r.faultRetryTimesMap[currentFaultWorkLoad] = retryTimes
		}
		// skip when retry times are exhausted
		if r.faultRetryTimesMap[currentFaultWorkLoad] <= 0 {
			r.Unlock()
			hwlog.RunLog.Infof("pod-level rescheduling: workload %s/%s retry times exhausted, skip",
				workLoadNamespacedName.Namespace, workLoadName)
			return nil
		}
		// decrement retry times before deleting the fault pod
		r.faultRetryTimesMap[currentFaultWorkLoad]--
	}
	r.Unlock()
	return r.deleteFaultPod(pod)
}

// isPodLevelRescheduling checks whether pod-level rescheduling is enabled on the
// pod via the pod-rescheduling label. The label lives on the pod template in the
// InferService, consistent with fault-scheduling and fault-retry-times which are
// also read from the pod.
func (r *Rescheduler) isPodLevelRescheduling(pod *corev1.Pod) bool {
	enabled := pod.Labels[common.PodReschedulingLabelKey] == common.PodReschedulingOn
	hwlog.RunLog.Infof("pod %s/%s pod-rescheduling label=%s, enabled=%t",
		pod.Namespace, pod.Name, pod.Labels[common.PodReschedulingLabelKey], enabled)
	return enabled
}

// deleteFaultPod deletes a single fault pod, supporting force/grace deletion modes.
// The mode is decided by the fault-scheduling label on the pod, keeping the same
// semantics as the instance-level deletePodsForExternalRescheduling:
//   - external-force / external-force-pod-failed: immediately force delete
//   - external-grace (default): graceful delete, force delete after grace timeout
func (r *Rescheduler) deleteFaultPod(pod *corev1.Pod) error {
	ctx := context.Background()
	mode := pod.Labels[common.FaultSchedulingLabelKey]
	ns, name := pod.Namespace, pod.Name
	hwlog.RunLog.Infof("pod-level rescheduling: deleting pod %s/%s with mode %s", ns, name, mode)
	switch mode {
	case common.ExternalForceReschedulingValue,
		common.ExternalForcePodFailedReschedulingValue:
		// force: delete immediately with GracePeriodSeconds(0), NotFound means already deleted
		if err := r.client.Delete(ctx, pod, client.GracePeriodSeconds(0)); err != nil &&
			!errors.IsNotFound(err) {
			return fmt.Errorf("failed to force delete fault pod %s/%s: %v", ns, name, err)
		}
		hwlog.RunLog.Infof("pod-level rescheduling: force deleted pod %s/%s", ns, name)
		return nil
	default:
		// grace: graceful delete first, force delete after timeout as fallback
		return r.gracefulDeleteFaultPod(ctx, pod)
	}
}

// gracefulDeleteFaultPod gracefully deletes the fault pod and starts an async
// force-delete fallback after the grace period expires.
func (r *Rescheduler) gracefulDeleteFaultPod(ctx context.Context, pod *corev1.Pod) error {
	ns, name := pod.Namespace, pod.Name
	waitSeconds := int64(common.DefaultTerminationGracePeriodSeconds)
	if grace := pod.Spec.TerminationGracePeriodSeconds; grace != nil && *grace > 0 {
		waitSeconds = *grace
	}
	if err := r.client.Delete(ctx, pod); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete fault pod %s/%s: %v", ns, name, err)
	}
	hwlog.RunLog.Infof("pod-level rescheduling: waiting %d seconds for graceful deletion of "+
		"pod %s/%s, will force delete after timeout", waitSeconds, ns, name)
	// wait asynchronously for the grace period, force delete the pod if it still exists
	go r.forceDeleteFaultPodAfterGrace(context.WithoutCancel(ctx), ns, name, pod.UID, waitSeconds)
	return nil
}

// forceDeleteFaultPodAfterGrace force-deletes the pod if it still exists after
// the grace period. It is the single-pod variant of forceDeletePodsAfterGrace
// (List degrades to Get for one pod).
func (r *Rescheduler) forceDeleteFaultPodAfterGrace(ctx context.Context,
	ns, name string, uid types.UID, waitSeconds int64) {
	defer func() {
		if rec := recover(); rec != nil {
			hwlog.RunLog.Errorf("panic in force-delete goroutine for pod %s/%s: %v", ns, name, rec)
		}
	}()
	time.Sleep(time.Duration(waitSeconds) * time.Second)
	var remainPod corev1.Pod
	if err := r.client.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &remainPod); err != nil {
		// pod has been gracefully deleted, no fallback needed
		if errors.IsNotFound(err) {
			return
		}
		hwlog.RunLog.Errorf("failed to get pod %s/%s on force-delete fallback: %v", ns, name, err)
		return
	}
	if remainPod.UID != uid {
		// a pod with the same name has been recreated, skip to avoid deleting the healthy pod
		hwlog.RunLog.Infof("pod %s/%s has been recreated (uid changed), skip force delete", ns, name)
		return
	}
	if err := r.client.Delete(ctx, &remainPod, client.GracePeriodSeconds(0)); err != nil &&
		!errors.IsNotFound(err) {
		hwlog.RunLog.Errorf("failed to force delete remaining pod %s/%s after grace period: %v",
			ns, name, err)
		return
	}
	hwlog.RunLog.Infof("pod-level rescheduling: force deleted remaining pod %s/%s after "+
		"grace period", ns, name)
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
	defer r.Unlock()
	// if a workload has multi faults, only process the first fault to reschedule workload
	_, exists := r.faultWorkLoadMap[currentFaultWorkLoad]
	if exists {
		hwlog.RunLog.Infof("pod %s/%s belongs to workload %s/%s which is already recorded for fault",
			pod.Namespace, pod.Name, workLoadNamespacedName.Namespace, workLoadName)
		return true
	}
	// read once; safe even if pod.Annotations is nil or key absent (returns "")
	faultReason := pod.Annotations[common.PodStatusAnnotationKey]
	r.faultWorkLoadMap[currentFaultWorkLoad] = faultReason
	if strings.HasSuffix(faultReason, common.PodFailed) {
		if _, exists := r.faultRetryTimesMap[currentFaultWorkLoad]; !exists {
			retryTimes, _ := strconv.Atoi(pod.Labels[common.FaultRetryTimesLabelKey])
			r.faultRetryTimesMap[currentFaultWorkLoad] = retryTimes
		}
	}
	hwlog.RunLog.Infof("record fault: %s for workload %s/%s",
		faultReason, pod.Namespace, workLoadName)
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

func getNamespacedNameList(workloadList []workload.WorkLoadInterface) map[types.NamespacedName]struct{} {
	namespacedNameMap := make(map[types.NamespacedName]struct{})
	for _, workload := range workloadList {
		objMeta := workload.GetWorkLoadObjMeta()
		namespacedNameMap[types.NamespacedName{Namespace: objMeta.GetNamespace(), Name: objMeta.GetName()}] = struct{}{}
	}
	return namespacedNameMap
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
	// 1. get fault workloads of current instanceSet
	workloadType := instanceSet.Spec.WorkloadTypeMeta
	gvk, err := common.WorkLoadTypeToGVK(workloadType)
	if err != nil {
		return nil, err
	}
	workloadHandler, err := r.workLoadHandlerFactory.GetWorkLoadHandler(gvk)
	if err != nil {
		return nil, err
	}
	currentFaultWorkLoadMap, err := r.getFaultWorkLoad(ctx, instanceSet, workloadHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to get fault workloads for instanceSet %s/%s: %v",
			instanceSet.Namespace, instanceSet.Name, err)
	}
	if len(currentFaultWorkLoadMap) == 0 {
		return make([]apiv1.InstanceSet, 0), nil
	}
	// 2. process priority setting
	deletedInstanceSets, err := r.processPrioritySetting(ctx, instanceSet, workloadHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to process priority setting when rescheduling: %v", err)
	}
	// 3. delete fault workloads
	err = r.deleteFaultWorkLoad(ctx, instanceSet, workloadHandler, currentFaultWorkLoadMap)
	if err != nil {
		return nil, fmt.Errorf("failed to delete fault workloads: %v", err)
	}
	deletedInstanceSets = append(deletedInstanceSets, *instanceSet)
	// 4. if rescheduling success, delete current fault workloads in faultWorkLoadMap
	r.Lock()
	defer r.Unlock()
	for currentFaultWorkLoad, _ := range currentFaultWorkLoadMap {
		delete(r.faultWorkLoadMap, currentFaultWorkLoad)
	}
	return deletedInstanceSets, nil
}

func (r *Rescheduler) processPrioritySetting(
	ctx context.Context,
	instanceSet *apiv1.InstanceSet,
	workloadHandler workload.WorkLoadHandler) ([]apiv1.InstanceSet, error) {
	priority := instanceSet.Spec.Priority
	priorityStrategy, ok2 := instanceSet.Labels[common.PrioritySchedulingStrategyLabelKey]
	if ok2 && priorityStrategy == common.SchedulingStrategyPriority && priority != nil {
		inferServiceName, ok := instanceSet.Labels[common.InferServiceNameLabelKey]
		if !ok {
			return nil, fmt.Errorf("instance set does not have infer service label: %v", instanceSet.Labels)
		}
		// delete unready workload that has lower priority than current instanceSet if it has priority setting
		deletedInstanceSets, err := r.deleteOtherWorkLoad(ctx, int(*priority), instanceSet.Namespace, inferServiceName, workloadHandler)
		if err != nil {
			return nil, fmt.Errorf("failed to delete unready work loads: %v", err)
		}
		return deletedInstanceSets, nil
	}
	return make([]apiv1.InstanceSet, 0), nil
}

func (r *Rescheduler) deleteOtherWorkLoad(
	ctx context.Context,
	priority int,
	namespace, serviceName string,
	workloadHandler workload.WorkLoadHandler) ([]apiv1.InstanceSet, error) {
	// 3.1 fetch other unready instanceSet that has lower priority than current instanceSet
	instanceSetList := &apiv1.InstanceSetList{}
	selector := labels.SelectorFromSet(labels.Set{
		common.InferServiceNameLabelKey: serviceName,
	})
	if err := r.client.List(ctx, instanceSetList,
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return nil, fmt.Errorf("failed to list InstanceSet for InferService %s/%s: %v at infer rescheduling", namespace, serviceName, err)
	}
	unreadyLowPriorityInstanceSetList := &apiv1.InstanceSetList{Items: []apiv1.InstanceSet{}}
	for _, instanceSet := range instanceSetList.Items {
		otherPriority := instanceSet.Spec.Priority
		if otherPriority == nil {
			hwlog.RunLog.Infof("instanceSet %s/%s has no priority label", instanceSet.Namespace, instanceSet.Name)
			continue
		}
		if int(*otherPriority) > priority && instanceSet.Status.ReadyReplicas < instanceSet.Status.Replicas {
			unreadyLowPriorityInstanceSetList.Items = append(unreadyLowPriorityInstanceSetList.Items, instanceSet)
		}
	}
	// 3.2 delete unready workload that has lower priority than current instanceSet
	if len(unreadyLowPriorityInstanceSetList.Items) > 0 {
		for _, instanceSet := range unreadyLowPriorityInstanceSetList.Items {
			indexer := common.InstanceIndexer{
				Namespace:      instanceSet.Namespace,
				ServiceName:    instanceSet.Labels[common.InferServiceNameLabelKey],
				InstanceSetKey: instanceSet.Labels[common.InstanceSetNameLabelKey],
			}
			selectLabels := make(map[string]string)
			selectLabels = common.AddLabelsFromIndexer(selectLabels, indexer)
			delete(selectLabels, common.InstanceIndexLabelKey)
			unReadyFilter := func(workload workload.WorkLoadInterface) bool {
				return !workload.IsWorkLoadReady()
			}
			if err := workloadHandler.DeleteWorkLoad(ctx, selectLabels, indexer.Namespace, unReadyFilter); err != nil {
				return nil, fmt.Errorf("failed to delete unready workload instance %v/%v: %v",
					instanceSet.Namespace, instanceSet.Name, err)
			}
		}
	}
	return unreadyLowPriorityInstanceSetList.Items, nil
}

func (r *Rescheduler) deleteFaultWorkLoad(
	ctx context.Context,
	instanceSet *apiv1.InstanceSet,
	workloadHandler workload.WorkLoadHandler,
	currentFaultWorkLoadMap map[faultWorkLoad]string) error {
	indexer := common.InstanceIndexer{
		Namespace:      instanceSet.Namespace,
		ServiceName:    instanceSet.Labels[common.InferServiceNameLabelKey],
		InstanceSetKey: instanceSet.Labels[common.InstanceSetNameLabelKey],
	}
	selectLabels := make(map[string]string)
	selectLabels = common.AddLabelsFromIndexer(selectLabels, indexer)
	delete(selectLabels, common.InstanceIndexLabelKey)
	faultFilter := func(workload workload.WorkLoadInterface) bool {
		objMeta := workload.GetWorkLoadObjMeta()
		currentFaultWorkLoad := faultWorkLoad{
			NamespacedName:  types.NamespacedName{Namespace: objMeta.GetNamespace(), Name: objMeta.GetName()},
			instanceSetName: instanceSet.Name,
		}
		faultReason, ok := currentFaultWorkLoadMap[currentFaultWorkLoad]
		if ok && strings.HasSuffix(faultReason, common.PodFailed) {
			r.Lock()
			defer r.Unlock()
			return r.faultRetryTimesMap[currentFaultWorkLoad] > 0
		}
		return ok
	}
	if err := workloadHandler.DeleteWorkLoad(ctx, selectLabels, instanceSet.Namespace, faultFilter); err != nil {
		return fmt.Errorf("failed to delete fault workload for instanceSet %v/%v: %v",
			instanceSet.Namespace, instanceSet.Name, err)
	}
	r.Lock()
	defer r.Unlock()
	for currentFaultWorkLoad, faultReason := range currentFaultWorkLoadMap {
		if strings.HasSuffix(faultReason, common.PodFailed) && r.faultRetryTimesMap[currentFaultWorkLoad] > 0 {
			r.faultRetryTimesMap[currentFaultWorkLoad]--
		}
	}
	return nil
}

func (r *Rescheduler) getFaultWorkLoad(
	ctx context.Context,
	instanceSet *apiv1.InstanceSet,
	workloadHandler workload.WorkLoadHandler) (map[faultWorkLoad]string, error) {
	indexer := common.InstanceIndexer{
		Namespace:      instanceSet.Namespace,
		ServiceName:    instanceSet.Labels[common.InferServiceNameLabelKey],
		InstanceSetKey: instanceSet.Labels[common.InstanceSetNameLabelKey],
	}
	selectLabels := make(map[string]string)
	selectLabels = common.AddLabelsFromIndexer(selectLabels, indexer)
	delete(selectLabels, common.InstanceIndexLabelKey)
	workloadList, err := workloadHandler.ListWorkLoad(ctx, selectLabels, instanceSet.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list all workload for instanceSet %s/%s: %v", instanceSet.Namespace, instanceSet.Name, err)
	}
	namespacedNameMap := getNamespacedNameList(workloadList)
	currentFaultWorkLoadMap := make(map[faultWorkLoad]string)
	r.Lock()
	defer r.Unlock()
	for namespacedName := range namespacedNameMap {
		currentFaultWorkLoad := faultWorkLoad{
			NamespacedName:  namespacedName,
			instanceSetName: instanceSet.Name,
		}
		faultReason, ok := r.faultWorkLoadMap[currentFaultWorkLoad]
		if ok {
			currentFaultWorkLoadMap[currentFaultWorkLoad] = faultReason
		}
	}
	return currentFaultWorkLoadMap, nil
}
