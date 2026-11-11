/*
Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
*/

package podautoscaler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"ascend-common/common-utils/hwlog"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apiv1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/autoscaling/algorithm"
	"infer-operator/pkg/autoscaling/autoscaler"
	autoscalingcontext "infer-operator/pkg/autoscaling/context"
	"infer-operator/pkg/autoscaling/metrics"
	autoscalingtypes "infer-operator/pkg/autoscaling/types"
	"infer-operator/pkg/controller/scaling"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestValidateOnlyAcceptsKPAInstanceSet(t *testing.T) {
	valid := validPA()
	if err := validatePodAutoscaler(valid); err != nil {
		t.Fatalf("valid object rejected: %v", err)
	}
	invalidStrategy := valid.DeepCopy()
	invalidStrategy.Spec.ScalingStrategy = "HPA"
	if err := validatePodAutoscaler(invalidStrategy); err == nil {
		t.Fatal("HPA strategy was accepted")
	}
	invalidTarget := valid.DeepCopy()
	invalidTarget.Spec.ScaleTargetRef.Kind = "Deployment"
	if err := validatePodAutoscaler(invalidTarget); err == nil {
		t.Fatal("Deployment target was accepted")
	}
}

func TestScaleDownStabilizationKeepsHighestRecommendation(t *testing.T) {
	store := newRuntimeStateStore()
	state := &runtimeState{evaluation: autoscalingcontext.NewEvaluationContext(), recommendations: []recommendation{
		{at: time.Now().Add(-time.Second), replicas: 4},
	}}
	if got := store.stabilize(state, 5, 2, time.Now()); got != 4 {
		t.Fatalf("got %d, want 4", got)
	}
}

func TestZeroCooldownDisablesStabilization(t *testing.T) {
	store := newRuntimeStateStore()
	state := &runtimeState{evaluation: autoscalingcontext.NewEvaluationContext(), recommendations: []recommendation{
		{at: time.Now().Add(-time.Minute), replicas: 2},
	}}
	state.evaluation.Policy.ScaleUpCooldownWindow = 0

	if got := store.stabilize(state, 2, 3, time.Now()); got != 3 {
		t.Fatalf("got %d, want 3", got)
	}
	if len(state.recommendations) != 0 {
		t.Fatalf("zero cooldown retained %d recommendations", len(state.recommendations))
	}
}

type fixedAutoScaler struct {
	desired int32
	cleaned bool
}

func (fixed *fixedAutoScaler) Recommend(context.Context,
	autoscaler.ReplicaComputeRequest) (*autoscaler.ReplicaComputeResult, error) {
	return &autoscaler.ReplicaComputeResult{DesiredReplicas: fixed.desired, Reason: "test recommendation",
		MetricValues: map[string]float64{"qps": 15}, Algorithm: "KPA", Valid: true}, nil
}

func (fixed *fixedAutoScaler) Forget(string, string) { fixed.cleaned = true }

// capturingAutoScaler records the ReplicaComputeRequest passed by the reconciler so
// tests can assert the evaluation base, while still returning a fixed recommendation.
type capturingAutoScaler struct {
	desired int32
	called  bool
	last    autoscaler.ReplicaComputeRequest
}

func (capture *capturingAutoScaler) Recommend(_ context.Context,
	request autoscaler.ReplicaComputeRequest) (*autoscaler.ReplicaComputeResult, error) {
	capture.called = true
	capture.last = request
	return &autoscaler.ReplicaComputeResult{DesiredReplicas: capture.desired, Reason: "test recommendation",
		MetricValues: map[string]float64{"qps": 15}, Algorithm: "KPA", Valid: true}, nil
}

func (capture *capturingAutoScaler) Forget(string, string) {}

type fixedCollector struct {
	value    float64
	failures map[string]error
}

func (collector *fixedCollector) Observe(_ context.Context,
	query metrics.PodMetricQuery) (autoscalingtypes.MetricObservation, error) {
	if err := collector.failures[query.Source.TargetName]; err != nil {
		return autoscalingtypes.MetricObservation{}, err
	}
	return autoscalingtypes.MetricObservation{
		Name: query.Source.TargetName, Values: []float64{collector.value}, Timestamp: query.Timestamp}, nil
}

func TestCollectObservationsToleratesOneUnavailableSource(t *testing.T) {
	pa := validPA()
	pa.Spec.MetricsSources = append(pa.Spec.MetricsSources,
		apiv1.MetricSource{TargetName: "waiting", TargetValue: "10", Port: "8000"})
	reconciler := &Reconciler{collector: &fixedCollector{value: 7,
		failures: map[string]error{"qps": errors.New("unavailable")}}}

	observations, err := reconciler.collectObservations(context.Background(), pa, nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(observations) != 1 || observations[0].Name != "waiting" || observations[0].Values[0] != 7 {
		t.Fatalf("unexpected observations: %+v", observations)
	}

	reconciler.collector = &fixedCollector{failures: map[string]error{
		"qps": errors.New("unavailable"), "waiting": errors.New("unavailable")}}
	if _, err := reconciler.collectObservations(context.Background(), pa, nil, time.Now()); err == nil {
		t.Fatal("all unavailable metric sources were accepted")
	}
}

func TestReconcileScalesInstanceSetAndUpdatesStatus(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := apiv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := autoscalingv2.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	replicas := int32(2)
	pa := validPA()
	target := &apiv1.InstanceSet{ObjectMeta: metav1.ObjectMeta{Name: "target", Namespace: "default"},
		Spec: apiv1.InstanceSetSpec{Replicas: &replicas}, Status: apiv1.InstanceSetStatus{LabelSelector: "app=demo"}}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "default", Labels: map[string]string{"app": "demo"}},
		Status: corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.1",
			Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}}}
	client := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(pa).WithObjects(pa, target, pod).Build()
	reconciler := &Reconciler{Client: client, Scheme: scheme, Recorder: record.NewFakeRecorder(10),
		collector: &fixedCollector{value: 15}, engine: &fixedAutoScaler{desired: 3}, scale: NewInstanceSetScale(client),
		resolver: scaling.NewOwnershipResolver(client), stateStore: newRuntimeStateStore()}
	if _, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{
		Namespace: "default", Name: "pa"}}); err != nil {
		t.Fatal(err)
	}
	updatedTarget := &apiv1.InstanceSet{}
	if err := client.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "target"}, updatedTarget); err != nil {
		t.Fatal(err)
	}
	if updatedTarget.Spec.Replicas == nil || *updatedTarget.Spec.Replicas != 3 {
		t.Fatalf("target replicas were not updated: %+v", updatedTarget.Spec.Replicas)
	}
	updatedPA := &apiv1.PodAutoscaler{}
	if err := client.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "pa"}, updatedPA); err != nil {
		t.Fatal(err)
	}
	if updatedPA.Status.Replicas.Desired != 3 || updatedPA.Status.Replicas.Ready != 1 {
		t.Fatalf("unexpected PodAutoscaler status: %+v", updatedPA.Status)
	}
	if updatedPA.Status.LastScale == nil || updatedPA.Status.LastScale.PreviousReplicas != 2 ||
		updatedPA.Status.LastScale.NewReplicas != 3 {
		t.Fatalf("last applied scale was not recorded: %+v", updatedPA.Status.LastScale)
	}
}

func TestReconcileReportsHPAConflict(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = apiv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = autoscalingv2.AddToScheme(scheme)
	pa := validPA()
	hpa := &autoscalingv2.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: "target-scaler", Namespace: "default"},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
			APIVersion: apiv1.GroupVersion.String(), Kind: "InstanceSet", Name: "target"}}}
	client := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(pa).WithObjects(pa, hpa).Build()
	fixed := &fixedAutoScaler{}
	reconciler := &Reconciler{Client: client, Scheme: scheme, Recorder: record.NewFakeRecorder(10), engine: fixed,
		scale: NewInstanceSetScale(client), resolver: scaling.NewOwnershipResolver(client), stateStore: newRuntimeStateStore()}
	if _, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: "pa"}}); err != nil {
		t.Fatal(err)
	}
	updated := &apiv1.PodAutoscaler{}
	_ = client.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "pa"}, updated)
	condition := meta.FindStatusCondition(updated.Status.Conditions, string(autoscalingtypes.ConditionConflict))
	if condition == nil || condition.Status != metav1.ConditionTrue {
		t.Fatalf("conflict was not reported: %+v", updated.Status)
	}
	reconciler.cleanupRuntimeState(types.NamespacedName{Namespace: "default", Name: "pa"})
	if !fixed.cleaned {
		t.Fatal("autoscaler state was not cleaned")
	}
}

// newScaleReconciler builds a reconciler backed by a fake client with one InstanceSet
// target (spec.replicas = replicas) and the supplied Pods labelled app=demo.
func newScaleReconciler(t *testing.T, replicas int32, pods []*corev1.Pod,
	collector metrics.ObservationCollector, engine autoscaler.RecommendationEngine) *Reconciler {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := apiv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := autoscalingv2.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	pa := validPA()
	target := &apiv1.InstanceSet{ObjectMeta: metav1.ObjectMeta{Name: "target", Namespace: "default"},
		Spec: apiv1.InstanceSetSpec{Replicas: &replicas}, Status: apiv1.InstanceSetStatus{LabelSelector: "app=demo"}}
	objects := []crclient.Object{pa, target}
	for _, pod := range pods {
		objects = append(objects, pod)
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(pa).WithObjects(objects...).Build()
	return &Reconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(10),
		collector: collector, engine: engine, scale: NewInstanceSetScale(fakeClient),
		resolver: scaling.NewOwnershipResolver(fakeClient), stateStore: newRuntimeStateStore()}
}

// readyPodsFixture builds count Running and Ready Pods labelled app=demo.
func readyPodsFixture(count int) []*corev1.Pod {
	pods := make([]*corev1.Pod, 0, count)
	for i := 0; i < count; i++ {
		pods = append(pods, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("pod-%d", i), Namespace: "default",
				Labels: map[string]string{"app": "demo"}},
			Status: corev1.PodStatus{Phase: corev1.PodRunning, PodIP: fmt.Sprintf("10.0.0.%d", i+1),
				Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}}})
	}
	return pods
}

func reconcileRequest() ctrl.Request {
	return ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: "pa"}}
}

// UT-K1: when every Pod is ready, the evaluation base equals spec.replicas, so the
// behaviour is identical to the pre-M7 code path.
func TestReconcileUsesReadyReplicasAsEvaluationBaseWhenAllReady(t *testing.T) {
	const replicas = int32(4)
	engine := &capturingAutoScaler{desired: replicas}
	reconciler := newScaleReconciler(t, replicas, readyPodsFixture(int(replicas)),
		&fixedCollector{value: 15}, engine)

	if _, err := reconciler.Reconcile(context.Background(), reconcileRequest()); err != nil {
		t.Fatal(err)
	}
	if !engine.called {
		t.Fatal("engine was not invoked")
	}
	if engine.last.CurrentReplicas != replicas {
		t.Fatalf("engine evaluation base=%d, want %d", engine.last.CurrentReplicas, replicas)
	}
}

// UT-K2: with only part of the Pods ready the evaluation base must be the ready count
// (not spec.replicas), which is the core M7 semantic and the one-step convergence fix.
func TestReconcileUsesReadyReplicasAsEvaluationBaseWhenPartiallyReady(t *testing.T) {
	const specReplicas = int32(5)
	engine := &capturingAutoScaler{desired: specReplicas}
	reconciler := newScaleReconciler(t, specReplicas, readyPodsFixture(2),
		&fixedCollector{value: 15}, engine)

	if _, err := reconciler.Reconcile(context.Background(), reconcileRequest()); err != nil {
		t.Fatal(err)
	}
	if !engine.called {
		t.Fatal("engine was not invoked")
	}
	if engine.last.CurrentReplicas != 2 {
		t.Fatalf("engine evaluation base=%d, want 2 (ready Pods)", engine.last.CurrentReplicas)
	}
}

// UT-K3: with no ready Pods the real collector fails ("no ready Pods available"), so the
// reconciler must take the MetricsUnavailable branch and never scale down.
func TestReconcileDoesNotScaleDownWhenNoReadyPods(t *testing.T) {
	const specReplicas = int32(3)
	engine := &capturingAutoScaler{desired: 1}
	collector := &fixedCollector{failures: map[string]error{
		"qps": errors.New("no ready Pods available for metric \"qps\"")}}
	reconciler := newScaleReconciler(t, specReplicas, nil, collector, engine)

	if _, err := reconciler.Reconcile(context.Background(), reconcileRequest()); err != nil {
		t.Fatal(err)
	}
	if engine.called {
		t.Fatal("engine was invoked even though no metrics were available")
	}

	updatedTarget := &apiv1.InstanceSet{}
	if err := reconciler.Get(context.Background(),
		types.NamespacedName{Namespace: "default", Name: "target"}, updatedTarget); err != nil {
		t.Fatal(err)
	}
	if updatedTarget.Spec.Replicas == nil || *updatedTarget.Spec.Replicas != specReplicas {
		t.Fatalf("target replicas changed without ready Pods: %+v", updatedTarget.Spec.Replicas)
	}

	updatedPA := &apiv1.PodAutoscaler{}
	if err := reconciler.Get(context.Background(),
		types.NamespacedName{Namespace: "default", Name: "pa"}, updatedPA); err != nil {
		t.Fatal(err)
	}
	condition := meta.FindStatusCondition(updatedPA.Status.Conditions, string(autoscalingtypes.ConditionMetrics))
	if condition == nil || condition.Status != metav1.ConditionFalse || condition.Reason != "MetricsUnavailable" {
		t.Fatalf("MetricsUnavailable condition was not recorded: %+v", updatedPA.Status.Conditions)
	}
}

func validPA() *apiv1.PodAutoscaler {
	return &apiv1.PodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: "pa", Namespace: "default"},
		Spec: apiv1.PodAutoscalerSpec{ScaleTargetRef: apiv1.ScaleTargetRef{APIVersion: apiv1.GroupVersion.String(), Kind: "InstanceSet", Name: "target"},
			MinReplicas: 1, MaxReplicas: 10, ScalingStrategy: apiv1.ScalingStrategyKPA,
			MetricsSources: []apiv1.MetricSource{{TargetName: "qps", TargetValue: "10", Port: "8000"}}}}
}

// TestDFxC05KPAAllReadyOutputIdentical reproduces design §9.2.3 TC-DFx-C05: when every Pod is
// ready the reconciler must feed the engine the legacy evaluation base (spec.replicas), and the
// real KPA algorithm must produce the same desired replicas as that legacy base, so the all-ready
// output is bit-for-bit unchanged.
func TestDFxC05KPAAllReadyOutputIdentical(t *testing.T) {
	const specReplicas = int32(4)
	pods := readyPodsFixture(int(specReplicas))

	// 1. reconciler level: the engine receives the legacy base while every Pod is ready.
	engine := &capturingAutoScaler{desired: 5}
	reconciler := newScaleReconciler(t, specReplicas, pods, &fixedCollector{value: 15}, engine)
	if _, err := reconciler.Reconcile(context.Background(), reconcileRequest()); err != nil {
		t.Fatal(err)
	}
	if !engine.called {
		t.Fatal("engine was not invoked")
	}
	if engine.last.CurrentReplicas != specReplicas {
		t.Fatalf("engine evaluation base=%d, want spec.replicas=%d", engine.last.CurrentReplicas, specReplicas)
	}

	// 2. algorithm level: the ready count equals spec.replicas, so the new and legacy bases
	// are equal and the real KPA result must therefore be identical.
	podValues := make([]corev1.Pod, len(pods))
	for i, pod := range pods {
		podValues[i] = *pod
	}
	readyCount := readyPods(podValues)
	if readyCount != specReplicas {
		t.Fatalf("all-ready fixture readyPods=%d, want %d", readyCount, specReplicas)
	}

	computeDesired := func(base int32) int32 {
		t.Helper()
		evaluation := autoscalingcontext.NewEvaluationContext()
		if err := evaluation.Refresh(validPA()); err != nil {
			t.Fatal(err)
		}
		recommendation, err := (&algorithm.KPAAlgorithm{}).ComputeRecommendation(context.Background(),
			algorithm.ScalingRequest{CurrentReplicas: base, Policy: &evaluation.Policy,
				RuntimeState: evaluation.StateFor("qps"),
				AggregatedMetrics: &autoscalingtypes.AggregatedMetrics{
					MetricKey:    autoscalingtypes.MetricKey{MetricName: "qps"},
					CurrentValue: 15, StableValue: 15, PanicValue: 15}})
		if err != nil {
			t.Fatal(err)
		}
		return recommendation.DesiredReplicas
	}

	legacyDesired := computeDesired(specReplicas)
	allReadyDesired := computeDesired(readyCount)
	if legacyDesired != allReadyDesired {
		t.Fatalf("legacy base desired=%d, all-ready base desired=%d", legacyDesired, allReadyDesired)
	}
	// observed 15 vs target 10 with base 4 gives ceil(4*15/10)=6, outside the tolerance band
	// (9..11) and within the rate limits (2..8), so the expected all-ready output is 6.
	if allReadyDesired != 6 {
		t.Fatalf("all-ready KPA desired=%d, want 6", allReadyDesired)
	}
}
