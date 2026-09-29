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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiMeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"ascend-common/common-utils/hwlog"

	apiv1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/autoscaling/autoscaler"
	"infer-operator/pkg/autoscaling/metrics"
	autoscalingtypes "infer-operator/pkg/autoscaling/types"
	"infer-operator/pkg/controller/scaling"
)

// Reconciler reconciles PodAutoscaler resources against InstanceSet targets.
type Reconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	collector  metrics.ObservationCollector
	engine     autoscaler.RecommendationEngine
	scale      *InstanceSetScale
	resolver   *scaling.OwnershipResolver
	stateStore *runtimeStateStore
}

// NewReconciler creates a PodAutoscaler reconciler and its in-memory KPA runtime state.
func NewReconciler(manager ctrl.Manager) *Reconciler {
	return &Reconciler{Client: manager.GetClient(), Scheme: manager.GetScheme(),
		Recorder:  manager.GetEventRecorderFor("podautoscaler-controller"),
		collector: metrics.NewDefaultCollector(), engine: autoscaler.NewEngine(),
		scale: NewInstanceSetScale(manager.GetClient()), resolver: scaling.NewOwnershipResolver(manager.GetClient()),
		stateStore: newRuntimeStateStore()}
}

// +kubebuilder:rbac:groups=mindcluster.huawei.com,resources=podautoscalers,verbs=get;list;watch
// +kubebuilder:rbac:groups=mindcluster.huawei.com,resources=podautoscalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mindcluster.huawei.com,resources=instancesets,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile validates ownership, collects Pod metrics, computes and stabilizes
// a KPA recommendation, updates InstanceSet replicas, and publishes status.
func (reconciler *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	pa := &apiv1.PodAutoscaler{}
	if err := reconciler.Get(ctx, request.NamespacedName, pa); err != nil {
		if apierrors.IsNotFound(err) {
			reconciler.cleanupRuntimeState(request.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if validationErr := validatePodAutoscaler(pa); validationErr != nil {
		err := reconciler.updatePodAutoscalerStatus(ctx, request.NamespacedName, func(status *apiv1.PodAutoscalerStatus) {
			status.ObservedGeneration = pa.Generation
			setCondition(status, autoscalingtypes.ConditionValidSpec, metav1.ConditionFalse, "InvalidSpec", validationErr.Error())
			setCondition(status, autoscalingtypes.ConditionReady, metav1.ConditionFalse, "InvalidSpec", validationErr.Error())
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	allowed, message, err := reconciler.resolver.CanManageWithKPA(ctx, pa)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !allowed {
		err := reconciler.updatePodAutoscalerStatus(ctx, request.NamespacedName, func(status *apiv1.PodAutoscalerStatus) {
			status.ObservedGeneration = pa.Generation
			setCondition(status, autoscalingtypes.ConditionConflict, metav1.ConditionTrue, "TargetAlreadyControlled", message)
			setCondition(status, autoscalingtypes.ConditionReady, metav1.ConditionFalse, "TargetAlreadyControlled", message)
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		reconciler.Recorder.Event(pa, corev1.EventTypeWarning, "AutoscalingConflict", message)
		return ctrl.Result{RequeueAfter: autoscalingtypes.DefaultReconcileInterval}, nil
	}

	workload, err := reconciler.getTargetWorkloadState(ctx, pa)
	if err != nil {
		return reconciler.fail(ctx, pa, "TargetUnavailable", err)
	}

	state, err := reconciler.stateStore.getOrCreate(pa)
	if err != nil {
		return reconciler.fail(ctx, pa, "InvalidSpec", err)
	}
	reconciler.logConfigurationIfChanged(pa, state)

	now := time.Now()
	observations, err := reconciler.collectObservations(ctx, pa, workload.pods, now)
	if err != nil {
		return reconciler.fail(ctx, pa, "MetricsUnavailable", err)
	}
	computeResult, err := reconciler.engine.Recommend(ctx, autoscaler.ReplicaComputeRequest{
		PodAutoscaler: pa, Evaluation: state.evaluation, CurrentReplicas: workload.currentReplicas,
		Observations: observations, Timestamp: now})
	if err != nil {
		return reconciler.fail(ctx, pa, "MetricsUnavailable", err)
	}

	desired := reconciler.stateStore.stabilize(state, workload.currentReplicas, computeResult.DesiredReplicas, now)
	for _, decision := range computeResult.MetricDecisions {
		hwlog.RunLog.Infof("KPA metric %s/%s metric=%s currentMetric=%.4g stableMetric=%.4g panicMetric=%.4g target=%.4g mode=%s",
			pa.Namespace, pa.Name, decision.Name, decision.CurrentValue, decision.StableValue,
			decision.PanicValue, decision.TargetValue, decision.Mode)
	}
	hwlog.RunLog.Infof("KPA decision %s/%s target=%s readyPods=%d selectedPods=%d currentReplicas=%d selectedMetric=%s recommendedReplicas=%d cooldownStabilizedReplicas=%d reason=%s",
		pa.Namespace, pa.Name, workload.target.Name, readyPods(workload.pods), len(workload.pods), workload.currentReplicas,
		computeResult.SelectedMetric, computeResult.DesiredReplicas, desired, computeResult.Reason)
	if desired != workload.currentReplicas {
		if err := reconciler.scale.UpdateReplicas(ctx, pa, desired); err != nil {
			return reconciler.fail(ctx, pa, "ScaleFailed", err)
		}
		hwlog.RunLog.Infof("PodAutoscaler %s/%s scaled InstanceSet %s from %d to %d replicas: metrics=%v algorithm=%s reason=%s",
			pa.Namespace, pa.Name, workload.target.Name, workload.currentReplicas, desired, computeResult.MetricValues,
			computeResult.Algorithm, computeResult.Reason)
		reconciler.Recorder.Eventf(pa, corev1.EventTypeNormal, "SuccessfulScale",
			"Scaled InstanceSet %s from %d to %d replicas", workload.target.Name, workload.currentReplicas, desired)
	}

	if err := reconciler.updatePodAutoscalerStatus(ctx, request.NamespacedName, func(status *apiv1.PodAutoscalerStatus) {
		status.ObservedGeneration = pa.Generation
		status.Replicas = apiv1.PodAutoscalerReplicaStatus{
			Current: workload.currentReplicas, Ready: readyPods(workload.pods), Desired: desired}
		status.MetricValues = computeResult.MetricValues
		if desired != workload.currentReplicas {
			status.LastScale = &apiv1.PodAutoscalerScaleStatus{Time: metav1.NewTime(now),
				PreviousReplicas: workload.currentReplicas, NewReplicas: desired, Reason: computeResult.Reason}
		}
		setCondition(status, autoscalingtypes.ConditionValidSpec, metav1.ConditionTrue, "Valid", "PodAutoscaler specification is valid")
		setCondition(status, autoscalingtypes.ConditionConflict, metav1.ConditionFalse, "NoConflict", "InstanceSet has one autoscaling controller")
		setCondition(status, autoscalingtypes.ConditionMetrics, metav1.ConditionTrue, "MetricsCollected", "Pod metrics were collected")
		setCondition(status, autoscalingtypes.ConditionScalingActive, metav1.ConditionTrue, "RecommendationComputed", computeResult.Reason)
		setCondition(status, autoscalingtypes.ConditionReady, metav1.ConditionTrue, "Ready", "PodAutoscaler is evaluating metrics")
	}); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: autoscalingtypes.DefaultReconcileInterval}, nil
}

type targetWorkloadState struct {
	target          *apiv1.InstanceSet
	currentReplicas int32
	pods            []corev1.Pod
}

func (reconciler *Reconciler) getTargetWorkloadState(ctx context.Context,
	pa *apiv1.PodAutoscaler) (*targetWorkloadState, error) {
	target, err := reconciler.scale.GetTarget(ctx, pa)
	if err != nil {
		return nil, err
	}
	current, err := reconciler.scale.GetCurrentReplicas(target)
	if err != nil {
		return nil, fmt.Errorf("failed to get current target replicas: %w", err)
	}
	selector, err := reconciler.scale.GetPodSelector(target)
	if err != nil {
		return nil, fmt.Errorf("failed to get target Pod selector: %w", err)
	}
	pods := &corev1.PodList{}
	if err := reconciler.List(ctx, pods, client.InNamespace(pa.Namespace),
		client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return nil, fmt.Errorf("failed to list target Pods: %w", err)
	}
	return &targetWorkloadState{target: target, currentReplicas: current, pods: pods.Items}, nil
}

func (reconciler *Reconciler) collectObservations(ctx context.Context, pa *apiv1.PodAutoscaler,
	pods []corev1.Pod, timestamp time.Time) ([]autoscalingtypes.MetricObservation, error) {
	if reconciler.collector == nil {
		return nil, fmt.Errorf("metric collector is not configured")
	}
	observations := make([]autoscalingtypes.MetricObservation, 0, len(pa.Spec.MetricsSources))
	failures := make([]error, 0)
	for _, source := range pa.Spec.MetricsSources {
		observation, err := reconciler.collector.Observe(ctx, metrics.PodMetricQuery{
			Namespace: pa.Namespace, TargetName: pa.Spec.ScaleTargetRef.Name,
			Source: source, Pods: pods, Timestamp: timestamp})
		if err != nil {
			failures = append(failures, fmt.Errorf("metric %q: %w", source.TargetName, err))
			continue
		}
		observations = append(observations, observation)
	}
	if len(observations) == 0 {
		return nil, fmt.Errorf("all metric sources failed: %w", errors.Join(failures...))
	}
	if len(failures) > 0 {
		hwlog.RunLog.Warnf("PodAutoscaler %s/%s ignored unavailable metric sources: %v",
			pa.Namespace, pa.Name, errors.Join(failures...))
	}
	return observations, nil
}

func (reconciler *Reconciler) logConfigurationIfChanged(pa *apiv1.PodAutoscaler, state *runtimeState) {
	stableWindow, panicWindow := autoscaler.MetricWindows(pa)
	policy := state.evaluation.Policy
	configLine := fmt.Sprintf("KPA config %s/%s target=%s min=%d max=%d stableWindow=%s panicWindow=%s panicThreshold=%.4g",
		pa.Namespace, pa.Name, pa.Spec.ScaleTargetRef.Name,
		policy.MinReplicas, policy.MaxReplicas, stableWindow, panicWindow, policy.PanicThreshold)
	policyLine := fmt.Sprintf("KPA policy %s/%s scaleUpRate=%.4g scaleDownRate=%.4g upTolerance=%.4g downTolerance=%.4g scaleUpCooldown=%s scaleDownCooldown=%s",
		pa.Namespace, pa.Name, policy.MaxScaleUpRate, policy.MaxScaleDownRate,
		policy.ScaleUpTolerance, policy.ScaleDownTolerance,
		policy.ScaleUpCooldownWindow, policy.ScaleDownCooldownWindow)
	metricLines := make([]string, 0, len(pa.Spec.MetricsSources))
	for _, source := range pa.Spec.MetricsSources {
		metricLines = append(metricLines, fmt.Sprintf("KPA target %s/%s metric=%s targetValue=%s",
			pa.Namespace, pa.Name, source.TargetName, source.TargetValue))
	}
	signature := configLine + "\x00" + policyLine + "\x00" + strings.Join(metricLines, "\x00")

	if !reconciler.stateStore.updateConfigSignature(state, signature) {
		return
	}

	hwlog.RunLog.Info(configLine)
	hwlog.RunLog.Info(policyLine)
	for _, metricLine := range metricLines {
		hwlog.RunLog.Info(metricLine)
	}
}

func (reconciler *Reconciler) fail(ctx context.Context, pa *apiv1.PodAutoscaler, reason string, cause error) (ctrl.Result, error) {
	hwlog.RunLog.Warnf("PodAutoscaler %s/%s: %v", pa.Namespace, pa.Name, cause)
	_ = reconciler.updatePodAutoscalerStatus(ctx, types.NamespacedName{Namespace: pa.Namespace, Name: pa.Name}, func(status *apiv1.PodAutoscalerStatus) {
		status.ObservedGeneration = pa.Generation
		setCondition(status, autoscalingtypes.ConditionMetrics, metav1.ConditionFalse, reason, cause.Error())
		setCondition(status, autoscalingtypes.ConditionReady, metav1.ConditionFalse, reason, cause.Error())
	})
	return ctrl.Result{RequeueAfter: autoscalingtypes.DefaultReconcileInterval}, nil
}

func validatePodAutoscaler(pa *apiv1.PodAutoscaler) error {
	ref := pa.Spec.ScaleTargetRef
	if ref.APIVersion != apiv1.GroupVersion.String() || ref.Kind != "InstanceSet" || ref.Name == "" {
		return fmt.Errorf("scaleTargetRef must identify a %s InstanceSet", apiv1.GroupVersion.String())
	}
	if pa.Spec.ScalingStrategy != "" && !strings.EqualFold(pa.Spec.ScalingStrategy, apiv1.ScalingStrategyKPA) {
		return fmt.Errorf("unsupported scalingStrategy %q: only %q is supported",
			pa.Spec.ScalingStrategy, apiv1.ScalingStrategyKPA)
	}
	if pa.Spec.MaxReplicas < 1 || pa.Spec.MinReplicas < 1 || pa.Spec.MinReplicas > pa.Spec.MaxReplicas {
		return fmt.Errorf("replica bounds are invalid")
	}
	if len(pa.Spec.MetricsSources) == 0 {
		return fmt.Errorf("at least one metricsSource is required")
	}
	for i, source := range pa.Spec.MetricsSources {
		value, err := strconv.ParseFloat(source.TargetValue, 64)
		if source.TargetName == "" || source.Port == "" || err != nil || value <= 0 {
			return fmt.Errorf("metricsSources[%d] requires a name, port and positive value", i)
		}
	}
	return nil
}

func (reconciler *Reconciler) cleanupRuntimeState(key types.NamespacedName) {
	reconciler.stateStore.delete(key)
	reconciler.engine.Forget(key.Namespace, key.Name)
}

func (reconciler *Reconciler) updatePodAutoscalerStatus(ctx context.Context, key types.NamespacedName,
	mutate func(*apiv1.PodAutoscalerStatus)) error {
	latest := &apiv1.PodAutoscaler{}
	if err := reconciler.Get(ctx, key, latest); err != nil {
		return client.IgnoreNotFound(err)
	}
	mutate(&latest.Status)
	return reconciler.Status().Update(ctx, latest)
}

func setCondition(status *apiv1.PodAutoscalerStatus, conditionType autoscalingtypes.ConditionType,
	conditionStatus metav1.ConditionStatus,
	reason, message string) {
	apiMeta.SetStatusCondition(&status.Conditions, metav1.Condition{Type: string(conditionType), Status: conditionStatus,
		Reason: reason, Message: message, ObservedGeneration: status.ObservedGeneration, LastTransitionTime: metav1.Now()})
}

func readyPods(pods []corev1.Pod) int32 {
	var count int32
	for i := range pods {
		if pods[i].DeletionTimestamp != nil || pods[i].Status.Phase != corev1.PodRunning {
			continue
		}
		for _, condition := range pods[i].Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				count++
				break
			}
		}
	}
	return count
}

// SetupWithManager registers PodAutoscaler changes. Target state and ownership
// are evaluated periodically, keeping this controller decoupled from workload events.
func (reconciler *Reconciler) SetupWithManager(manager ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(manager).
		For(&apiv1.PodAutoscaler{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named("podautoscaler-controller").
		Complete(reconciler)
}
