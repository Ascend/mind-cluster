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
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

// TestNewWorkLoadReconciler tests the NewWorkLoadReconciler function.
func TestNewWorkLoadReconciler(t *testing.T) {
	convey.Convey("Test NewWorkLoadReconciler function", t, func() {
		convey.Convey("Should create a new WorkLoadReconciler", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			convey.So(reconciler, convey.ShouldNotBeNil)
			convey.So(reconciler.client, convey.ShouldEqual, fakeClient)
			convey.So(reconciler.handlerRegisterMap, convey.ShouldNotBeNil)
			convey.So(reconciler.PodGroupManager, convey.ShouldNotBeNil)
		})
	})
}

// TestWorkLoadReconcilerRegister tests the Register method of WorkLoadReconciler.
func TestWorkLoadReconcilerRegister(t *testing.T) {
	convey.Convey("Test WorkLoadReconciler Register method", t, func() {
		convey.Convey("Should register a workload handler", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			mockHandler := &mockWorkLoadHandler{}
			reconciler.Register(gvk, mockHandler)

			handler, ok := reconciler.handlerRegisterMap[gvk.String()]
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(handler, convey.ShouldEqual, mockHandler)
		})
	})
}

// TestWorkLoadReconcilerValidate tests the Validate method of WorkLoadReconciler.
func TestWorkLoadReconcilerValidate(t *testing.T) {
	convey.Convey("Test WorkLoadReconciler Validate method", t, func() {
		convey.Convey("Should successfully validate InstanceSet", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)

			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			mockHandler := &mockWorkLoadHandler{}
			reconciler.Register(gvk, mockHandler)

			err := reconciler.Validate(instanceSet)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when validation fails", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			mockHandler := &mockWorkLoadHandler{
				validateError: errors.New("validation failed"),
			}
			reconciler.Register(gvk, mockHandler)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			err := reconciler.Validate(instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestWorkLoadReconcilerReconcile tests the Reconcile method of WorkLoadReconciler.
func TestWorkLoadReconcilerReconcile(t *testing.T) {
	convey.Convey("Test WorkLoadReconciler Reconcile method", t, func() {
		convey.Convey("Should successfully reconcile without gang scheduling", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			mockHandler := &mockWorkLoadHandler{}
			reconciler.Register(gvk, mockHandler)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err := reconciler.Reconcile(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should successfully reconcile with gang scheduling", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			instanceSet.Labels[common.GangScheduleLabelKey] = common.TrueBool
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			mockHandler := &mockWorkLoadHandler{}
			reconciler.Register(gvk, mockHandler)

			patches := gomonkey.ApplyMethodReturn(reconciler.PodGroupManager, "GetOrCreatePodGroupForInstance", true, nil)
			defer patches.Reset()

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err := reconciler.Reconcile(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestWorkLoadReconcilerReconcile2 tests the Reconcile method of WorkLoadReconciler
// when gang scheduling is enabled and creating PodGroup fails.
func TestWorkLoadReconcilerReconcile2(t *testing.T) {
	convey.Convey("Test WorkLoadReconciler Reconcile method", t, func() {
		convey.Convey("Should return error when creating PodGroup fails", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			instanceSet.Labels[common.GangScheduleLabelKey] = common.TrueBool
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			mockHandler := &mockWorkLoadHandler{}
			reconciler.Register(gvk, mockHandler)

			mockErr := errors.New("failed to create podgroup")
			patches := gomonkey.ApplyMethodReturn(reconciler.PodGroupManager,
				"GetOrCreatePodGroupForInstance", false, mockErr)
			defer patches.Reset()

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err := reconciler.Reconcile(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when CheckOrCreateWorkLoad fails", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			mockHandler := &mockWorkLoadHandler{
				checkOrCreateError: errors.New("failed to create workload"),
			}
			reconciler.Register(gvk, mockHandler)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err := reconciler.Reconcile(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestWorkLoadReconcilerInstanceReady tests the InstanceReady method of WorkLoadReconciler.
func TestWorkLoadReconcilerInstanceReady(t *testing.T) {
	convey.Convey("Test WorkLoadReconciler InstanceReady method", t, func() {
		convey.Convey("Should return ready replicas count", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			mockHandler := &mockWorkLoadHandler{
				readyReplicas: 1,
			}
			reconciler.Register(gvk, mockHandler)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			readyReplicas, err := reconciler.InstanceReady(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
			convey.So(readyReplicas, convey.ShouldEqual, 1)
		})

		convey.Convey("Should return error when getting ready replicas fails", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			mockHandler := &mockWorkLoadHandler{
				getReadyReplicasError: errors.New("failed to get ready replicas"),
			}
			reconciler.Register(gvk, mockHandler)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			_, err := reconciler.InstanceReady(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestWorkLoadReconcilerDeleteExtraInstances tests the DeleteExtraInstances method of WorkLoadReconciler.
func TestWorkLoadReconcilerDeleteExtraInstances(t *testing.T) {
	convey.Convey("Test WorkLoadReconciler DeleteExtraInstances method", t, func() {
		convey.Convey("Should successfully delete extra instances", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			replicas := int32(3)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			mockHandler := &mockWorkLoadHandler{}
			reconciler.Register(gvk, mockHandler)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err := reconciler.DeleteExtraInstances(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when deleting extra instances fails", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			mockHandler := &mockWorkLoadHandler{
				deleteExtraWorkLoadError: errors.New("failed to delete extra instances"),
			}
			reconciler.Register(gvk, mockHandler)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err := reconciler.DeleteExtraInstances(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestGetWorkLoadReconciler tests the getWorkLoadReconciler method of WorkLoadReconciler.
func TestGetWorkLoadReconciler(t *testing.T) {
	convey.Convey("Test WorkLoadReconciler getWorkLoadReconciler method", t, func() {
		convey.Convey("Should return registered handler", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			mockHandler := &mockWorkLoadHandler{}
			reconciler.Register(gvk, mockHandler)

			handler, err := reconciler.getWorkLoadReconciler(instanceSet)
			convey.So(err, convey.ShouldBeNil)
			convey.So(handler, convey.ShouldEqual, mockHandler)
		})

		convey.Convey("Should return error when handler not registered", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			handler, err := reconciler.getWorkLoadReconciler(instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(handler, convey.ShouldBeNil)
		})
	})
}

// TestNewPodGroupSpec tests the newPodGroupSpec function.
func TestNewPodGroupSpec(t *testing.T) {
	convey.Convey("Test newPodGroupSpec function", t, func() {
		convey.Convey("Should create PodGroupSpec with correct MinMember", func() {
			normalReplicas := int32(3)
			spec := newPodGroupSpec(normalReplicas)

			convey.So(spec.MinMember, convey.ShouldEqual, int32(3))
		})

		convey.Convey("Should handle zero replicas", func() {
			zeroReplicas := int32(0)
			spec := newPodGroupSpec(zeroReplicas)

			convey.So(spec.MinMember, convey.ShouldEqual, int32(0))
		})
	})
}

type mockWorkLoadHandler struct {
	validateError            error
	checkOrCreateError       error
	deleteExtraWorkLoadError error
	getReadyReplicasError    error
	readyReplicas            int
}

func (m *mockWorkLoadHandler) CheckOrCreateWorkLoad(context.Context, *v1.InstanceSet,
	common.InstanceIndexer) error {
	return m.checkOrCreateError
}

func (m *mockWorkLoadHandler) DeleteExtraWorkLoad(context.Context, common.InstanceIndexer, int) error {
	return m.deleteExtraWorkLoadError
}

func (m *mockWorkLoadHandler) GetWorkLoadReadyReplicas(context.Context, common.InstanceIndexer) (int, error) {
	return m.readyReplicas, m.getReadyReplicasError
}

func (m *mockWorkLoadHandler) Validate(runtime.RawExtension) error {
	return m.validateError
}

func (m *mockWorkLoadHandler) GetReplicas(runtime.RawExtension) (int32, error) {
	return 1, nil
}
