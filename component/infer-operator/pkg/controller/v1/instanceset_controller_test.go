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

// Package v1 contains API Schema definitions for the mindcluster v1 API group
package v1

import (
	"context"
	"errors"
	"fmt"
	"infer-operator/pkg/controller/rescheduling"
	"infer-operator/pkg/controller/scaling"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
	"infer-operator/pkg/controller/schedule"
	"infer-operator/pkg/controller/workload"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// TestInstanceSetReconcilerReconcile1 tests the Reconcile method of InstanceSetReconciler.
func TestInstanceSetReconcilerReconcile1(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler Reconcile method", t, func() {
		convey.Convey("Should successfully reconcile InstanceSet", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()

			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
				ScalingManager:     scaling.NewScalingManager(fakeClient, GetScheme()),
			}

			patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "validate",
				func(_ *InstanceSetReconciler, is *v1.InstanceSet) error { return nil })
			defer patches.Reset()
			patches2 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "reconcileWorkLoads",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return nil })
			defer patches2.Reset()
			patches3 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "updateStatus",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error {
					return nil
				})
			defer patches3.Reset()
			patches4 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "doRescheduling",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return nil })
			defer patches4.Reset()
			patches5 := gomonkey.ApplyMethodReturn(reconciler.ScalingManager, "ReconcileScalingResource",
				&v1.ScalingResourceStatus{}, nil)
			defer patches5.Reset()

			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-instance",
					Namespace: "default",
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})

		convey.Convey("Should return not found when InstanceSet does not exist", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := &InstanceSetReconciler{
				Client: fakeClient,
			}

			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-instance",
					Namespace: "default",
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})
	})
}

// TestInstanceSetReconcilerReconcile2 tests the Reconcile method of InstanceSetReconciler.
func TestInstanceSetReconcilerReconcile2(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler Reconcile method", t, func() {
		convey.Convey("Should return error when getting InstanceSet fails", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := &InstanceSetReconciler{
				Client: fakeClient,
			}

			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-instance",
					Namespace: "default",
				},
			}

			mockErr := errors.New("failed to get instanceset")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Get", mockErr)
			defer patches.Reset()

			_, err := reconciler.Reconcile(ctx, req)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestUpdateStatusForScalingSuccess(t *testing.T) {
	convey.Convey("should update ScalingResourceStatus and LabelSelector", t, func() {
		instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
		fakeClient := NewFakeClient(instanceSet).Build()
		reconciler := &InstanceSetReconciler{Client: fakeClient}

		scalingStatus := &v1.ScalingResourceStatus{
			Type: "HPA", Name: "test-hpa", Ready: true, Message: "HPA is ready",
		}

		ctx := context.Background()
		err := reconciler.updateStatusForScaling(ctx, instanceSet, scalingStatus)
		convey.So(err, convey.ShouldBeNil)

		updatedInstanceSet := &v1.InstanceSet{}
		_ = fakeClient.Get(ctx, types.NamespacedName{Name: "test-instance", Namespace: "default"}, updatedInstanceSet)
		convey.So(updatedInstanceSet.Status.ScalingResourceStatus, convey.ShouldResemble, scalingStatus)
		expectedLabelSelector := fmt.Sprintf("%s=%s,%s=%s",
			common.InferServiceNameLabelKey, instanceSet.Labels[common.InferServiceNameLabelKey],
			common.InstanceSetNameLabelKey, instanceSet.Labels[common.InstanceSetNameLabelKey])
		convey.So(updatedInstanceSet.Status.LabelSelector, convey.ShouldEqual, expectedLabelSelector)
	})
}

func TestUpdateStatusForScalingSkipUnchanged(t *testing.T) {
	convey.Convey("should skip update when ScalingResourceStatus and LabelSelector unchanged", t, func() {
		instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
		scalingStatus := &v1.ScalingResourceStatus{
			Type: "HPA", Name: "test-hpa", Ready: true, Message: "HPA is ready",
		}
		instanceSet.Status.ScalingResourceStatus = scalingStatus
		instanceSet.Status.LabelSelector = fmt.Sprintf("%s=%s,%s=%s",
			common.InferServiceNameLabelKey, instanceSet.Labels[common.InferServiceNameLabelKey],
			common.InstanceSetNameLabelKey, instanceSet.Labels[common.InstanceSetNameLabelKey])

		fakeClient := NewFakeClient(instanceSet).Build()
		reconciler := &InstanceSetReconciler{Client: fakeClient}

		ctx := context.Background()
		err := reconciler.updateStatusForScaling(ctx, instanceSet, scalingStatus)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestUpdateStatusForScalingStatusChanged(t *testing.T) {
	convey.Convey("should update when ScalingResourceStatus changed", t, func() {
		instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
		instanceSet.Status.ScalingResourceStatus = &v1.ScalingResourceStatus{
			Type: "HPA", Name: "old-hpa", Ready: false, Message: "HPA not ready",
		}
		instanceSet.Status.LabelSelector = fmt.Sprintf("%s=%s,%s=%s",
			common.InferServiceNameLabelKey, instanceSet.Labels[common.InferServiceNameLabelKey],
			common.InstanceSetNameLabelKey, instanceSet.Labels[common.InstanceSetNameLabelKey])

		fakeClient := NewFakeClient(instanceSet).Build()
		reconciler := &InstanceSetReconciler{Client: fakeClient}

		newStatus := &v1.ScalingResourceStatus{
			Type: "HPA", Name: "new-hpa", Ready: true, Message: "HPA is ready",
		}

		ctx := context.Background()
		err := reconciler.updateStatusForScaling(ctx, instanceSet, newStatus)
		convey.So(err, convey.ShouldBeNil)

		updatedInstanceSet := &v1.InstanceSet{}
		_ = fakeClient.Get(ctx, types.NamespacedName{Name: "test-instance", Namespace: "default"}, updatedInstanceSet)
		convey.So(updatedInstanceSet.Status.ScalingResourceStatus, convey.ShouldResemble, newStatus)
	})
}

func TestUpdateStatusForScalingLabelSelectorChanged(t *testing.T) {
	convey.Convey("should update when LabelSelector changed", t, func() {
		instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
		instanceSet.Status.ScalingResourceStatus = &v1.ScalingResourceStatus{
			Type: "HPA", Name: "test-hpa", Ready: true,
		}
		instanceSet.Status.LabelSelector = "old-label-selector"

		fakeClient := NewFakeClient(instanceSet).Build()
		reconciler := &InstanceSetReconciler{Client: fakeClient}

		ctx := context.Background()
		err := reconciler.updateStatusForScaling(ctx, instanceSet, instanceSet.Status.ScalingResourceStatus)
		convey.So(err, convey.ShouldBeNil)

		updatedInstanceSet := &v1.InstanceSet{}
		_ = fakeClient.Get(ctx, types.NamespacedName{Name: "test-instance", Namespace: "default"}, updatedInstanceSet)
		expectedLabelSelector := fmt.Sprintf("%s=%s,%s=%s",
			common.InferServiceNameLabelKey, instanceSet.Labels[common.InferServiceNameLabelKey],
			common.InstanceSetNameLabelKey, instanceSet.Labels[common.InstanceSetNameLabelKey])
		convey.So(updatedInstanceSet.Status.LabelSelector, convey.ShouldEqual, expectedLabelSelector)
	})
}

func TestUpdateStatusForScalingGetError(t *testing.T) {
	convey.Convey("should return error when getting InstanceSet fails", t, func() {
		instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
		fakeClient := NewFakeClient(instanceSet).Build()
		reconciler := &InstanceSetReconciler{Client: fakeClient}

		patches := gomonkey.ApplyMethodReturn(reconciler.Client, "Get", errors.New("failed to get instanceset"))
		defer patches.Reset()

		scalingStatus := &v1.ScalingResourceStatus{Type: "HPA", Name: "test-hpa", Ready: true}

		ctx := context.Background()
		err := reconciler.updateStatusForScaling(ctx, instanceSet, scalingStatus)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestUpdateStatusForScalingUpdateError(t *testing.T) {
	convey.Convey("should return error when status update fails", t, func() {
		instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
		fakeClient := NewFakeClient(instanceSet).Build()
		reconciler := &InstanceSetReconciler{Client: fakeClient}

		patches := gomonkey.ApplyMethodFunc(reconciler.Client, "Status", func() client.StatusWriter {
			return &mockStatusWriter{updateErr: errors.New("failed to update status")}
		})
		defer patches.Reset()

		scalingStatus := &v1.ScalingResourceStatus{Type: "HPA", Name: "test-hpa", Ready: true}

		ctx := context.Background()
		err := reconciler.updateStatusForScaling(ctx, instanceSet, scalingStatus)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestUpdateStatusForScalingNilStatus(t *testing.T) {
	convey.Convey("should handle nil scalingStatus", t, func() {
		instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
		instanceSet.Status.ScalingResourceStatus = &v1.ScalingResourceStatus{
			Type: "HPA", Name: "old-hpa", Ready: true,
		}

		fakeClient := NewFakeClient(instanceSet).Build()
		reconciler := &InstanceSetReconciler{Client: fakeClient}

		ctx := context.Background()
		err := reconciler.updateStatusForScaling(ctx, instanceSet, nil)
		convey.So(err, convey.ShouldBeNil)

		updatedInstanceSet := &v1.InstanceSet{}
		_ = fakeClient.Get(ctx, types.NamespacedName{Name: "test-instance", Namespace: "default"}, updatedInstanceSet)
		convey.So(updatedInstanceSet.Status.ScalingResourceStatus, convey.ShouldBeNil)
	})
}

func newTestReconcilerForScaling(fakeClient client.Client) *InstanceSetReconciler {
	return &InstanceSetReconciler{
		Client:             fakeClient,
		WorkLoadReconciler: workload.NewWorkLoadReconciler(fakeClient),
		ScalingManager:     scaling.NewScalingManager(fakeClient, GetScheme()),
	}
}

func patchReconcilerBasics(r *InstanceSetReconciler) []*gomonkey.Patches {
	var allPatches []*gomonkey.Patches
	methods := map[string]interface{}{
		"validate":           func(_ *InstanceSetReconciler, is *v1.InstanceSet) error { return nil },
		"reconcileWorkLoads": func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return nil },
		"updateStatus":       func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return nil },
		"doRescheduling":     func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return nil },
	}
	for name, fn := range methods {
		p := gomonkey.ApplyPrivateMethod(reflect.TypeOf(r), name, fn)
		allPatches = append(allPatches, p)
	}
	return allPatches
}

func resetAllPatches(allPatches []*gomonkey.Patches) {
	for _, p := range allPatches {
		p.Reset()
	}
}

func TestReconcileCallsUpdateStatusForScalingOnSuccess(t *testing.T) {
	convey.Convey("should call updateStatusForScaling when scalingErr is nil", t, func() {
		instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
		fakeClient := NewFakeClient(instanceSet).Build()
		reconciler := newTestReconcilerForScaling(fakeClient)
		allPatches := patchReconcilerBasics(reconciler)
		defer resetAllPatches(allPatches)

		scalingStatus := &v1.ScalingResourceStatus{Type: "HPA", Name: "test-hpa", Ready: true}
		patches5 := gomonkey.ApplyMethodReturn(reconciler.ScalingManager, "ReconcileScalingResource", scalingStatus, nil)
		defer patches5.Reset()

		called := false
		patches6 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "updateStatusForScaling",
			func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet, status *v1.ScalingResourceStatus) error {
				called = true
				convey.So(status, convey.ShouldResemble, scalingStatus)
				return nil
			})
		defer patches6.Reset()

		ctx := context.Background()
		req := reconcile.Request{NamespacedName: types.NamespacedName{Name: "test-instance", Namespace: "default"}}
		_, err := reconciler.Reconcile(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(called, convey.ShouldBeTrue)
	})
}

func TestReconcileSkipsUpdateStatusForScalingOnError(t *testing.T) {
	convey.Convey("should not call updateStatusForScaling when scalingErr is not nil", t, func() {
		instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
		fakeClient := NewFakeClient(instanceSet).Build()
		reconciler := newTestReconcilerForScaling(fakeClient)
		reconciler.Recorder = record.NewFakeRecorder(10)
		allPatches := patchReconcilerBasics(reconciler)
		defer resetAllPatches(allPatches)

		patches5 := gomonkey.ApplyMethodReturn(reconciler.ScalingManager, "ReconcileScalingResource",
			(*v1.ScalingResourceStatus)(nil), errors.New("scaling reconcile error"))
		defer patches5.Reset()

		called := false
		patches6 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "updateStatusForScaling",
			func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet, status *v1.ScalingResourceStatus) error {
				called = true
				return nil
			})
		defer patches6.Reset()

		ctx := context.Background()
		req := reconcile.Request{NamespacedName: types.NamespacedName{Name: "test-instance", Namespace: "default"}}
		_, err := reconciler.Reconcile(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(called, convey.ShouldBeFalse)
	})
}

func TestReconcileContinuesWhenUpdateStatusForScalingFails(t *testing.T) {
	convey.Convey("should continue when updateStatusForScaling fails", t, func() {
		instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
		fakeClient := NewFakeClient(instanceSet).Build()
		reconciler := newTestReconcilerForScaling(fakeClient)
		allPatches := patchReconcilerBasics(reconciler)
		defer resetAllPatches(allPatches)

		scalingStatus := &v1.ScalingResourceStatus{Type: "HPA", Name: "test-hpa", Ready: true}
		patches5 := gomonkey.ApplyMethodReturn(reconciler.ScalingManager, "ReconcileScalingResource", scalingStatus, nil)
		defer patches5.Reset()

		patches6 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "updateStatusForScaling",
			func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet, status *v1.ScalingResourceStatus) error {
				return errors.New("update scaling status failed")
			})
		defer patches6.Reset()

		ctx := context.Background()
		req := reconcile.Request{NamespacedName: types.NamespacedName{Name: "test-instance", Namespace: "default"}}
		result, err := reconciler.Reconcile(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldResemble, reconcile.Result{})
	})
}

func TestUpdateStatusPreservesScalingFieldsFromServer(t *testing.T) {
	convey.Convey("updateStatus should preserve ScalingResourceStatus and LabelSelector from server", t, func() {
		instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
		scalingStatus := &v1.ScalingResourceStatus{
			Type: "HPA", Name: "test-hpa", Ready: true, Message: "HPA is ready",
		}
		instanceSet.Status.ScalingResourceStatus = scalingStatus
		instanceSet.Status.LabelSelector = fmt.Sprintf("%s=%s,%s=%s",
			common.InferServiceNameLabelKey, instanceSet.Labels[common.InferServiceNameLabelKey],
			common.InstanceSetNameLabelKey, instanceSet.Labels[common.InstanceSetNameLabelKey])

		fakeClient := NewFakeClient(instanceSet).Build()
		workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
		reconciler := &InstanceSetReconciler{Client: fakeClient, WorkLoadReconciler: workLoadReconciler}

		patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 1, nil)
		defer patches.Reset()

		ctx := context.Background()
		err := reconciler.updateStatus(ctx, instanceSet)
		convey.So(err, convey.ShouldBeNil)

		updatedInstanceSet := &v1.InstanceSet{}
		_ = fakeClient.Get(ctx, types.NamespacedName{Name: "test-instance", Namespace: "default"}, updatedInstanceSet)
		convey.So(updatedInstanceSet.Status.ScalingResourceStatus, convey.ShouldResemble, scalingStatus)
		convey.So(updatedInstanceSet.Status.LabelSelector, convey.ShouldEqual, instanceSet.Status.LabelSelector)
	})
}

func TestUpdateStatusPreservesNilScalingFieldsFromServer(t *testing.T) {
	convey.Convey("updateStatus should preserve nil ScalingResourceStatus from server", t, func() {
		instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
		fakeClient := NewFakeClient(instanceSet).Build()
		workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
		reconciler := &InstanceSetReconciler{Client: fakeClient, WorkLoadReconciler: workLoadReconciler}

		patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 1, nil)
		defer patches.Reset()

		ctx := context.Background()
		err := reconciler.updateStatus(ctx, instanceSet)
		convey.So(err, convey.ShouldBeNil)

		updatedInstanceSet := &v1.InstanceSet{}
		_ = fakeClient.Get(ctx, types.NamespacedName{Name: "test-instance", Namespace: "default"}, updatedInstanceSet)
		convey.So(updatedInstanceSet.Status.ScalingResourceStatus, convey.ShouldBeNil)
	})
}

// TestInstanceSetReconcilerReconcile3 tests the Reconcile method of InstanceSetReconciler.
func TestInstanceSetReconcilerReconcile3(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler Reconcile method", t, func() {
		convey.Convey("Should return error when validation fails", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()

			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			validateErr := errors.New("validation failed")
			patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "validate",
				func(_ *InstanceSetReconciler, is *v1.InstanceSet) error { return validateErr })
			defer patches.Reset()

			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-instance",
					Namespace: "default",
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})

		convey.Convey("Should return error when reconcileWorkLoads fails with non-conflict error", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()

			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			sm := scaling.NewScalingManager(fakeClient, GetScheme())
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
				ScalingManager:     sm,
			}

			patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "validate",
				func(_ *InstanceSetReconciler, is *v1.InstanceSet) error { return nil })
			defer patches.Reset()
			reconcileErr := errors.New("reconcile failed")
			patches2 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "reconcileWorkLoads",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return reconcileErr })
			defer patches2.Reset()
			patches3 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "doRescheduling",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return nil })
			defer patches3.Reset()
			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-instance",
					Namespace: "default",
				},
			}

			_, err := reconciler.Reconcile(ctx, req)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestInstanceSetReconcilerReconcile4 tests the Reconcile method of InstanceSetReconciler.
func TestInstanceSetReconcilerReconcile4(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler Reconcile method", t, func() {
		convey.Convey("Should return nil when reconcileWorkLoads fails with conflict error", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()

			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			sm := scaling.NewScalingManager(fakeClient, GetScheme())
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
				ScalingManager:     sm,
			}

			patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "validate",
				func(_ *InstanceSetReconciler, is *v1.InstanceSet) error { return nil })
			defer patches.Reset()
			conflictErr := apierrors.NewConflict(schema.GroupResource{},
				"test-instance", errors.New("conflict"))
			patches2 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "reconcileWorkLoads",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return conflictErr })
			defer patches2.Reset()
			patches3 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "doRescheduling",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return nil })
			defer patches3.Reset()

			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-instance",
					Namespace: "default",
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{Requeue: true})
		})
	})
}

// TestInstanceSetReconcilerReconcile5 tests the Reconcile method of InstanceSetReconciler.
func TestInstanceSetReconcilerReconcile5(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler Reconcile method", t, func() {
		convey.Convey("Should return error when updateStatus fails", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()

			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
				ScalingManager:     scaling.NewScalingManager(fakeClient, GetScheme()),
			}

			patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "validate",
				func(_ *InstanceSetReconciler, is *v1.InstanceSet) error { return nil })
			defer patches.Reset()
			patches2 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "reconcileWorkLoads",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return nil })
			defer patches2.Reset()
			updateStatusErr := errors.New("update status failed")
			patches3 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "updateStatus",
				func(_ *InstanceSetReconciler, ctx context.Context,
					is *v1.InstanceSet) error {
					return updateStatusErr
				})
			defer patches3.Reset()
			patches4 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "doRescheduling",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return nil })
			defer patches4.Reset()
			patches5 := gomonkey.ApplyMethodReturn(reconciler.ScalingManager, "ReconcileScalingResource",
				&v1.ScalingResourceStatus{}, nil)
			defer patches5.Reset()

			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-instance",
					Namespace: "default",
				},
			}

			_, err := reconciler.Reconcile(ctx, req)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestInstanceSetReconcilerReconcileWithRequeueError tests the Reconcile method of InstanceSetReconciler.
func TestInstanceSetReconcilerReconcileWithRequeueError(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler Reconcile method", t, func() {
		convey.Convey("Should return nil when reconcileWorkLoads fails with requeue error", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()

			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			workLoadHandlerFactory := workload.NewWorkLoadHandlerFactory()
			deploymentGVK := appsv1.SchemeGroupVersion.WithKind("Deployment")
			deploymentHandler := workload.NewDeploymentHandler(fakeClient)
			err := workLoadHandlerFactory.Register(deploymentGVK, deploymentHandler)
			convey.So(err, convey.ShouldBeNil)
			rescheduler := rescheduling.NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			rescheduler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
				rescheduler:        rescheduler,
				ScalingManager:     scaling.NewScalingManager(fakeClient, GetScheme()),
			}

			patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "validate",
				func(_ *InstanceSetReconciler, is *v1.InstanceSet) error { return nil })
			defer patches.Reset()
			requeueErr := common.NewRequeueError("requeue error")
			patches2 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "reconcileWorkLoads",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return requeueErr })
			defer patches2.Reset()
			patches3 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "updateStatus",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error {
					return nil
				})
			defer patches3.Reset()
			patches4 := gomonkey.ApplyMethodReturn(reconciler.ScalingManager, "ReconcileScalingResource",
				&v1.ScalingResourceStatus{}, nil)
			defer patches4.Reset()

			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-instance",
					Namespace: "default",
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{RequeueAfter: common.DefaultReEnqueueInterval})
		})
	})
}

// TestReconcileWorkLoadsShouldSchedule tests the reconcileWorkLoads method with different ShouldSchedule results.
func TestReconcileWorkLoadsShouldSchedule(t *testing.T) {
	convey.Convey("Test reconcileWorkLoads with ShouldSchedule", t, func() {
		instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
		instanceSet.Labels[common.PrioritySchedulingStrategyLabelKey] = common.SchedulingStrategyPriority
		fakeClient := NewFakeClient(instanceSet).Build()
		workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
		reconciler := &InstanceSetReconciler{Client: fakeClient, WorkLoadReconciler: workLoadReconciler}
		ctx := context.Background()
		patches := gomonkey.ApplyMethod(reflect.TypeOf(workLoadReconciler), "DeleteExtraInstances",
			func(_ *workload.WorkLoadReconciler, ctx context.Context, is *v1.InstanceSet, indexer common.InstanceIndexer) error {
				return nil
			})
		defer patches.Reset()
		patches2 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "reconcileServices",
			func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet, indexer common.InstanceIndexer) error {
				return nil
			})
		defer patches2.Reset()
		ifSchedule := false
		patches3 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "reconcileWorkloadInstances",
			func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet, indexer common.InstanceIndexer) error {
				ifSchedule = true
				return nil
			})
		defer patches3.Reset()

		convey.Convey("Should return nil when ShouldSchedule returns false", func() {
			patches4 := gomonkey.ApplyFuncReturn(schedule.ShouldSchedule, false, nil)
			defer patches4.Reset()
			err := reconciler.reconcileWorkLoads(ctx, instanceSet)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ifSchedule, convey.ShouldBeFalse)
		})
		convey.Convey("Should continue reconciliation when ShouldSchedule returns true", func() {
			patches4 := gomonkey.ApplyFuncReturn(schedule.ShouldSchedule, true, nil)
			defer patches4.Reset()
			err := reconciler.reconcileWorkLoads(ctx, instanceSet)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ifSchedule, convey.ShouldBeTrue)
		})
		convey.Convey("Should return error when ShouldSchedule returns error", func() {
			scheduleErr := errors.New("schedule error")
			patches4 := gomonkey.ApplyFuncReturn(schedule.ShouldSchedule, false, scheduleErr)
			defer patches4.Reset()
			err := reconciler.reconcileWorkLoads(ctx, instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestCheckOrCreateService tests the checkOrCreateService method of WorkLoadReconciler.
func TestCheckOrCreateService(t *testing.T) {
	convey.Convey("Test WorkLoadReconciler checkOrCreateService method", t, func() {
		convey.Convey("Should return nil when service already exists", func() {
			existingService := CreateTestService("test-service", "default")
			fakeClient := NewFakeClient(existingService).Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			instanceSet := CreateTestInstanceSet("test", "default", int32(1))
			serviceSpec := v1.ServiceSpec{
				Name: "test-service",
				Spec: existingService.Spec,
			}

			ctx := context.Background()
			indexer := GetTestIndexer("test-service", "test-role", "0")
			err := reconciler.checkOrCreateService(ctx, instanceSet, serviceSpec, indexer)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should create new service when not found", func() {
			fakeClient := NewFakeClient().Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			instanceSet := CreateTestInstanceSet("test", "default", int32(1))
			serviceSpec := getTestServiceSpec()

			ctx := context.Background()
			indexer := GetTestIndexer("inferservice", "test-role", "0")
			err := reconciler.checkOrCreateService(ctx, instanceSet, serviceSpec, indexer)
			convey.So(err, convey.ShouldBeNil)

			createdService := &corev1.Service{}
			serviceName := fmt.Sprintf("new-service-%s-%s", indexer.ServiceName, indexer.InstanceSetKey)
			err = fakeClient.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: "default"}, createdService)
			convey.So(err, convey.ShouldBeNil)
			convey.So(createdService.Name, convey.ShouldEqual, serviceName)
			convey.So(createdService.Namespace, convey.ShouldEqual, "default")
			convey.So(len(createdService.OwnerReferences), convey.ShouldEqual, 1)
		})
	})
}

// TestCheckOrCreateService2 tests the checkOrCreateService method of WorkLoadReconciler.
func TestCheckOrCreateService2(t *testing.T) {
	convey.Convey("Test WorkLoadReconciler checkOrCreateService method", t, func() {
		convey.Convey("Should return error when getting service fails with non-NotFound error", func() {
			fakeClient := NewFakeClient().Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}
			instanceSet := CreateTestInstanceSet("test", "default", int32(1))
			serviceSpec := v1.ServiceSpec{
				Name: "test-service",
				Spec: corev1.ServiceSpec{},
			}

			patches := gomonkey.ApplyMethodReturn(reconciler.Client, "Get",
				errors.New("network error"))
			defer patches.Reset()

			ctx := context.Background()
			indexer := GetTestIndexer("test-service", "test-role", "0")
			err := reconciler.checkOrCreateService(ctx, instanceSet, serviceSpec, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when creating service fails", func() {
			fakeClient := NewFakeClient().Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			bufferSize := 100
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
				Recorder:           record.NewFakeRecorder(bufferSize),
			}
			instanceSet := CreateTestInstanceSet("test", "default", int32(1))
			serviceSpec := getTestServiceSpec()

			patches := gomonkey.ApplyMethodReturn(reconciler.Client, "Create",
				errors.New("create failed"))
			defer patches.Reset()

			ctx := context.Background()
			indexer := GetTestIndexer("new-service", "test-role", "0")
			err := reconciler.checkOrCreateService(ctx, instanceSet, serviceSpec, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestInstanceSetReconcilerValidate tests the validate method of InstanceSetReconciler.
func TestInstanceSetReconcilerValidate(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler validate method", t, func() {
		convey.Convey("Should successfully validate InstanceSet", func() {
			fakeClient := NewFakeClient().Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}
			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "Validate", nil)
			defer patches.Reset()
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			err := reconciler.validate(instanceSet)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when gang-scheduling is enabled but not supported", func() {
			fakeClient := NewFakeClient().Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			bufferSize := 10
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
				SupportPodGroup:    false,
				Recorder:           record.NewFakeRecorder(bufferSize),
			}
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			instanceSet.Labels[common.GangScheduleLabelKey] = common.TrueBool
			err := reconciler.validate(instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when replicas is nil", func() {
			fakeClient := NewFakeClient().Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			bufferSize := 10
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
				Recorder:           record.NewFakeRecorder(bufferSize),
			}
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			instanceSet.Spec.Replicas = nil
			err := reconciler.validate(instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestInstanceSetReconcilerValidate2 tests the validate method of InstanceSetReconciler.
func TestInstanceSetReconcilerValidate2(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler validate method", t, func() {
		convey.Convey("Should return error when replicas is negative", func() {
			fakeClient := NewFakeClient().Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			bufferSize := 10
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
				Recorder:           record.NewFakeRecorder(bufferSize),
			}

			replicas := int32(-1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			err := reconciler.validate(instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when InferServiceNameLabelKey is missing", func() {
			fakeClient := NewFakeClient().Build()
			bufferSize := 10
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
				Recorder:           record.NewFakeRecorder(bufferSize),
			}
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			delete(instanceSet.Labels, common.InferServiceNameLabelKey)
			err := reconciler.validate(instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestInstanceSetReconcilerValidate3 tests the validate method of InstanceSetReconciler.
func TestInstanceSetReconcilerValidate3(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler validate method", t, func() {
		convey.Convey("Should return error when workload validation fails", func() {
			fakeClient := NewFakeClient().Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			bufferSize := 10
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
				Recorder:           record.NewFakeRecorder(bufferSize),
			}
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			mockErr := errors.New("workload validation failed")
			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "Validate", mockErr)
			defer patches.Reset()
			err := reconciler.validate(instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestReconcileWorkLoads tests the reconcileWorkLoads method of InstanceSetReconciler.
func TestReconcileWorkLoads(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler reconcileWorkLoads method", t, func() {
		convey.Convey("Should successfully reconcile workloads", func() {
			fakeClient := NewFakeClient().Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "DeleteExtraInstances", nil)
			defer patches.Reset()
			patches2 := gomonkey.ApplyMethodReturn(workLoadReconciler, "Reconcile", nil)
			defer patches2.Reset()

			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			ctx := context.Background()
			err := reconciler.reconcileWorkLoads(ctx, instanceSet)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when DeleteExtraInstances fails", func() {
			fakeClient := NewFakeClient().Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))

			deleteErr := errors.New("delete extra instances failed")
			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "DeleteExtraInstances", deleteErr)
			defer patches.Reset()

			ctx := context.Background()
			err := reconciler.reconcileWorkLoads(ctx, instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestReconcileWorkLoads2 tests the reconcileWorkLoads method of InstanceSetReconciler.
func TestReconcileWorkLoads2(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler reconcileWorkLoads method", t, func() {
		convey.Convey("Should return error when Reconcile fails", func() {
			fakeClient := NewFakeClient().Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))

			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "DeleteExtraInstances", nil)
			defer patches.Reset()

			reconcileErr := errors.New("reconcile failed")
			patches2 := gomonkey.ApplyMethodReturn(workLoadReconciler, "Reconcile", reconcileErr)
			defer patches2.Reset()

			ctx := context.Background()
			err := reconciler.reconcileWorkLoads(ctx, instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should handle zero replicas", func() {
			fakeClient := NewFakeClient().Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}
			replicas := int32(0)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)

			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "DeleteExtraInstances", nil)
			defer patches.Reset()

			ctx := context.Background()
			err := reconciler.reconcileWorkLoads(ctx, instanceSet)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestInstanceSetReconcilerUpdateStatus tests the updateStatus method of InstanceSetReconciler.
func TestInstanceSetReconcilerUpdateStatus(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler updateStatus method", t, func() {
		convey.Convey("Should successfully update status", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 1, nil)
			defer patches.Reset()

			ctx := context.Background()
			err := reconciler.updateStatus(ctx, instanceSet)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when getting InstanceSet fails", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 1, nil)
			defer patches.Reset()
			mockErr := errors.New("failed to get instanceset")
			patches2 := gomonkey.ApplyMethodReturn(reconciler.Client, "Get", mockErr)
			defer patches2.Reset()

			ctx := context.Background()
			err := reconciler.updateStatus(ctx, instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestInstanceSetReconcilerUpdateStatus2 tests the updateStatus method of InstanceSetReconciler.
func TestInstanceSetReconcilerUpdateStatus2(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler updateStatus method", t, func() {
		convey.Convey("Should return error when InstanceReady fails", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()

			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			mockErr := errors.New("failed to get ready replicas")
			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 0, mockErr)
			defer patches.Reset()

			ctx := context.Background()
			err := reconciler.updateStatus(ctx, instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when updating status fails", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()

			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 1, nil)
			defer patches.Reset()

			ctx := context.Background()

			mockErr := errors.New("failed to update status")
			patches2 := gomonkey.ApplyMethodFunc(reconciler.Client, "Status", func() client.StatusWriter {
				mockStatusWriter := &mockStatusWriter{
					updateErr: mockErr,
				}
				return mockStatusWriter
			})
			defer patches2.Reset()

			err := reconciler.updateStatus(ctx, instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestWorkLoadPredicate tests the WorkLoadPredicate function.
func TestWorkLoadPredicate(t *testing.T) {
	convey.Convey("Test WorkLoadPredicate function", t, func() {
		convey.Convey("Should skip create event", func() {
			predicate := WorkLoadPredicate()
			result := predicate.Create(event.CreateEvent{})
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should process update event", func() {
			predicate := WorkLoadPredicate()
			result := predicate.Update(event.UpdateEvent{})
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("Should process delete event", func() {
			predicate := WorkLoadPredicate()
			result := predicate.Delete(event.DeleteEvent{})
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

// TestPodGroupPredicate tests the PodGroupPredicate function.
func TestPodGroupPredicate(t *testing.T) {
	convey.Convey("Test PodGroupPredicate function", t, func() {
		convey.Convey("Should skip create event", func() {
			predicate := PodGroupPredicate()
			result := predicate.Create(event.CreateEvent{})
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should skip update event", func() {
			predicate := PodGroupPredicate()
			result := predicate.Update(event.UpdateEvent{})
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should process delete event", func() {
			predicate := PodGroupPredicate()
			result := predicate.Delete(event.DeleteEvent{})
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

// TestGetNewStatus tests the getNewStatus method of InstanceSetReconciler.
func TestGetNewStatus(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler getNewStatus method", t, func() {
		convey.Convey("Should successfully get new status when all replicas are ready", func() {
			replicas := int32(2)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			fakeClient := NewFakeClient(instanceSet).Build()

			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			readyReplicas := 2
			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", readyReplicas, nil)
			defer patches.Reset()

			ctx := context.Background()
			indexer := GetTestIndexer("test-service", "test-role", "0")
			newStatus, err := reconciler.getNewStatus(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
			convey.So(newStatus.Replicas, convey.ShouldEqual, int32(2))
			convey.So(newStatus.ReadyReplicas, convey.ShouldEqual, int32(2))
			convey.So(newStatus.ObservedGeneration, convey.ShouldEqual, instanceSet.Generation)
			convey.So(len(newStatus.Conditions), convey.ShouldEqual, 1)
			convey.So(newStatus.Conditions[0].Type, convey.ShouldEqual, string(common.InstanceSetReady))
			convey.So(newStatus.Conditions[0].Status, convey.ShouldEqual, metav1.ConditionTrue)
			convey.So(newStatus.Conditions[0].Reason, convey.ShouldEqual, "AllWorkLoadReady")
			convey.So(newStatus.ScalingResourceStatus, convey.ShouldBeNil)
		})
	})
}

// TestGetNewStatus2 tests the getNewStatus method of InstanceSetReconciler.
func TestGetNewStatus2(t *testing.T) {
	convey.Convey("Test InstanceSetReconciler getNewStatus method", t, func() {
		convey.Convey("Should successfully get new status when not all replicas are ready", func() {
			replicas := int32(2)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			fakeClient := NewFakeClient(instanceSet).Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 1, nil)
			defer patches.Reset()

			ctx := context.Background()
			indexer := GetTestIndexer("test-service", "test-role", "0")
			newStatus, err := reconciler.getNewStatus(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
			convey.So(newStatus.Replicas, convey.ShouldEqual, int32(2))
			convey.So(newStatus.ReadyReplicas, convey.ShouldEqual, int32(1))
			convey.So(newStatus.ObservedGeneration, convey.ShouldEqual, instanceSet.Generation)
			convey.So(len(newStatus.Conditions), convey.ShouldEqual, 1)
			convey.So(newStatus.Conditions[0].Type, convey.ShouldEqual, string(common.InstanceSetReady))
			convey.So(newStatus.Conditions[0].Status, convey.ShouldEqual, metav1.ConditionFalse)
			convey.So(newStatus.Conditions[0].Reason, convey.ShouldEqual, "ReplicasNotReady")
			convey.So(newStatus.ScalingResourceStatus, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when InstanceReady fails", func() {
			replicas := int32(2)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			fakeClient := NewFakeClient(instanceSet).Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			mockErr := errors.New("failed to get ready replicas")
			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 0, mockErr)
			defer patches.Reset()

			ctx := context.Background()
			indexer := GetTestIndexer("test-service", "test-role", "0")
			_, err := reconciler.getNewStatus(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestValidateServices tests the validateServices function.
func TestValidateServices(t *testing.T) {
	convey.Convey("Test validateServices function", t, func() {
		convey.Convey("Should successfully validate services with Non-NodePort type", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			instanceSet.Spec.Services = []v1.ServiceSpec{
				getTestServiceSpec(),
			}
			instanceSet.Spec.Services[0].Spec.Ports[0].NodePort = int32(30080)
			instanceSet.Spec.Services[0].Spec.Type = corev1.ServiceTypeClusterIP

			err := validateServices(instanceSet)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should successfully validate services with NodePort type and non-zero nodePort", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			instanceSet.Spec.Services = []v1.ServiceSpec{
				getTestServiceSpec(),
			}
			instanceSet.Spec.Services[0].Spec.Ports[0].NodePort = int32(30080)
			instanceSet.Spec.Services[0].Spec.Type = corev1.ServiceTypeNodePort

			err := validateServices(instanceSet)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when nodePort is zero for NodePort type", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			instanceSet.Spec.Services = []v1.ServiceSpec{
				getTestServiceSpec(),
			}
			instanceSet.Spec.Services[0].Spec.Ports[0].NodePort = int32(0)
			instanceSet.Spec.Services[0].Spec.Type = corev1.ServiceTypeNodePort

			err := validateServices(instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestHandleConflictNodePort tests the handleConflictNodePort function.
func TestHandleConflictNodePort(t *testing.T) {
	convey.Convey("Test handleConflictNodePort function", t, func() {
		convey.Convey("Should successfully add offset to nodePort", func() {
			serviceSpec := getTestServiceSpec()
			serviceSpec.Spec.Ports[0].NodePort = int32(30080)
			indexer := common.InstanceIndexer{
				ServiceName:    "test-service-1",
				InstanceSetKey: "test-role",
			}
			err := handleConflictNodePort(&serviceSpec, indexer)
			convey.So(err, convey.ShouldBeNil)
			convey.So(serviceSpec.Spec.Ports[0].NodePort, convey.ShouldEqual, int32(30081))
		})

		convey.Convey("Should return error when service index is not a number", func() {
			serviceSpec := getTestServiceSpec()
			serviceSpec.Spec.Ports[0].NodePort = int32(30080)
			indexer := common.InstanceIndexer{
				ServiceName:    "test-service-abc",
				InstanceSetKey: "test-role",
			}
			err := handleConflictNodePort(&serviceSpec, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when nodePort is zero", func() {
			serviceSpec := getTestServiceSpec()
			serviceSpec.Spec.Ports[0].NodePort = int32(0)
			indexer := common.InstanceIndexer{
				ServiceName:    "test-service-1",
				InstanceSetKey: "test-role",
			}
			err := handleConflictNodePort(&serviceSpec, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func getTestServiceSpec() v1.ServiceSpec {
	return v1.ServiceSpec{
		Name: "new-service",
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "test",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Port:     8080,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
}

func TestInstanceSetPredicate(t *testing.T) {
	convey.Convey("Given an InstanceSetReconciler", t, func() {
		ir := &InstanceSetReconciler{}

		// Mock CleanupWithInstanceSetDeletion method
		patches := gomonkey.ApplyMethod(ir.rescheduler, "CleanupWithInstanceSetDeletion",
			func(_ *rescheduling.Rescheduler, name string) {
			})
		defer patches.Reset()

		convey.Convey("When creating predicate", func() {
			pred := instanceSetPredicate(ir)

			convey.Convey("Then UpdateFunc should return true", func() {
				updateEvent := event.UpdateEvent{}
				convey.So(pred.Update(updateEvent), convey.ShouldBeTrue)
			})

			convey.Convey("Then CreateFunc should return true", func() {
				createEvent := event.CreateEvent{}
				convey.So(pred.Create(createEvent), convey.ShouldBeTrue)
			})

			convey.Convey("Then DeleteFunc should return true and call cleanup", func() {
				deleteEvent := event.DeleteEvent{
					Object: &v1.InstanceSet{
						ObjectMeta: metav1.ObjectMeta{Name: "test-instance"},
					},
				}
				convey.So(pred.Delete(deleteEvent), convey.ShouldBeTrue)
			})
		})
	})
}

func TestReconcileScalingResources(t *testing.T) {
	convey.Convey("TestReconcileScalingResources", t, func() {
		convey.Convey("ScalingManager returns error", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()
			sm := scaling.NewScalingManager(fakeClient, GetScheme())
			bufferSize := 10
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workload.NewWorkLoadReconciler(fakeClient),
				ScalingManager:     sm,
				Recorder:           record.NewFakeRecorder(bufferSize),
			}

			scalingErr := errors.New("scaling reconcile error")
			patches := gomonkey.ApplyMethodReturn(sm, "ReconcileScalingResource",
				(*v1.ScalingResourceStatus)(nil), scalingErr)
			defer patches.Reset()

			ctx := context.Background()
			_, err := reconciler.reconcileScalingResources(ctx, instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("ScalingManager returns nil status", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()
			sm := scaling.NewScalingManager(fakeClient, GetScheme())
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workload.NewWorkLoadReconciler(fakeClient),
				ScalingManager:     sm,
			}

			patches := gomonkey.ApplyMethodReturn(sm, "ReconcileScalingResource",
				(*v1.ScalingResourceStatus)(nil), nil)
			defer patches.Reset()

			ctx := context.Background()
			status, err := reconciler.reconcileScalingResources(ctx, instanceSet)
			convey.So(err, convey.ShouldBeNil)
			convey.So(status, convey.ShouldBeNil)
		})

		convey.Convey("ScalingManager returns status", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()
			sm := scaling.NewScalingManager(fakeClient, GetScheme())
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workload.NewWorkLoadReconciler(fakeClient),
				ScalingManager:     sm,
			}

			expectedStatus := &v1.ScalingResourceStatus{
				Type:    "HPA",
				Name:    "test-hpa",
				Ready:   true,
				Message: "HPA is ready",
			}
			patches := gomonkey.ApplyMethodReturn(sm, "ReconcileScalingResource",
				expectedStatus, nil)
			defer patches.Reset()

			ctx := context.Background()
			status, err := reconciler.reconcileScalingResources(ctx, instanceSet)
			convey.So(err, convey.ShouldBeNil)
			convey.So(status, convey.ShouldResemble, expectedStatus)
		})
	})
}

func TestUpdateStatusWithScaling(t *testing.T) {
	convey.Convey("TestUpdateStatusWithScaling", t, func() {
		convey.Convey("nil instanceSet", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workload.NewWorkLoadReconciler(fakeClient),
			}
			err := reconciler.updateStatus(context.Background(), nil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("status unchanged", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 0, nil)
			defer patches.Reset()

			ctx := context.Background()
			err := reconciler.updateStatus(ctx, instanceSet)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("success with nil scalingStatus", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 1, nil)
			defer patches.Reset()

			ctx := context.Background()
			err := reconciler.updateStatus(ctx, instanceSet)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("success with scalingStatus", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 1, nil)
			defer patches.Reset()

			scalingStatus := &v1.ScalingResourceStatus{
				Type:    "HPA",
				Name:    "test-hpa",
				Ready:   true,
				Message: "HPA is ready",
			}
			instanceSet.Status.ScalingResourceStatus = scalingStatus
			ctx := context.Background()
			err := reconciler.updateStatus(ctx, instanceSet)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("getNewStatus error", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			mockErr := errors.New("instance ready error")
			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 0, mockErr)
			defer patches.Reset()

			ctx := context.Background()
			err := reconciler.updateStatus(ctx, instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("get latest instanceSet error", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 1, nil)
			defer patches.Reset()
			mockErr := errors.New("get error")
			patches2 := gomonkey.ApplyMethodReturn(reconciler.Client, "Get", mockErr)
			defer patches2.Reset()

			ctx := context.Background()
			err := reconciler.updateStatus(ctx, instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("status update error", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()
			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
			}

			patches := gomonkey.ApplyMethodReturn(workLoadReconciler, "InstanceReady", 1, nil)
			defer patches.Reset()

			mockErr := errors.New("failed to update status")
			patches2 := gomonkey.ApplyMethodFunc(reconciler.Client, "Status", func() client.StatusWriter {
				return &mockStatusWriter{updateErr: mockErr}
			})
			defer patches2.Reset()

			ctx := context.Background()
			err := reconciler.updateStatus(ctx, instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestReconcileScalingResourcesError(t *testing.T) {
	convey.Convey("TestReconcile ReconcileScalingResources error path", t, func() {
		convey.Convey("Should return error when reconcileScalingResources fails", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := NewFakeClient(instanceSet).Build()

			workLoadReconciler := workload.NewWorkLoadReconciler(fakeClient)
			sm := scaling.NewScalingManager(fakeClient, GetScheme())
			reconciler := &InstanceSetReconciler{
				Client:             fakeClient,
				WorkLoadReconciler: workLoadReconciler,
				ScalingManager:     sm,
				Recorder:           record.NewFakeRecorder(10),
			}

			patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "validate",
				func(_ *InstanceSetReconciler, is *v1.InstanceSet) error { return nil })
			defer patches.Reset()
			patches2 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "reconcileWorkLoads",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return nil })
			defer patches2.Reset()
			patches3 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "doRescheduling",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return nil })
			defer patches3.Reset()
			patches4 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(reconciler), "updateStatus",
				func(_ *InstanceSetReconciler, ctx context.Context, is *v1.InstanceSet) error { return nil })
			defer patches4.Reset()

			scalingErr := errors.New("scaling reconcile error")
			patches5 := gomonkey.ApplyMethodReturn(sm, "ReconcileScalingResource",
				(*v1.ScalingResourceStatus)(nil), scalingErr)
			defer patches5.Reset()

			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-instance",
					Namespace: "default",
				},
			}

			_, err := reconciler.Reconcile(ctx, req)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}
