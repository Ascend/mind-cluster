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
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apiv1 "infer-operator/pkg/api/v1"
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

func validPA() *apiv1.PodAutoscaler {
	return &apiv1.PodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: "pa", Namespace: "default"},
		Spec: apiv1.PodAutoscalerSpec{ScaleTargetRef: apiv1.ScaleTargetRef{APIVersion: apiv1.GroupVersion.String(), Kind: "InstanceSet", Name: "target"},
			MinReplicas: 1, MaxReplicas: 10, ScalingStrategy: apiv1.ScalingStrategyKPA,
			MetricsSources: []apiv1.MetricSource{{TargetName: "qps", TargetValue: "10", Port: "8000"}}}}
}
