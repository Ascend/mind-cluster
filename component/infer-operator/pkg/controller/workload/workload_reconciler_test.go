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
	"sort"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			convey.So(reconciler.workLoadHandlerFactory, convey.ShouldNotBeNil)
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
			workLoadHandlerFactory := NewWorkLoadHandlerFactory()
			err := workLoadHandlerFactory.Register(gvk, mockHandler)
			convey.So(err, convey.ShouldBeNil)
			reconciler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)

			handler, err := reconciler.workLoadHandlerFactory.GetWorkLoadHandler(gvk)
			convey.So(err, convey.ShouldBeNil)
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
			workLoadHandlerFactory := NewWorkLoadHandlerFactory()
			err := workLoadHandlerFactory.Register(gvk, mockHandler)
			convey.So(err, convey.ShouldBeNil)
			reconciler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)

			err = reconciler.Validate(instanceSet)
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
			workLoadHandlerFactory := NewWorkLoadHandlerFactory()
			err := workLoadHandlerFactory.Register(gvk, mockHandler)
			convey.So(err, convey.ShouldBeNil)
			reconciler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			err = reconciler.Validate(instanceSet)
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
			workLoadHandlerFactory := NewWorkLoadHandlerFactory()
			err := workLoadHandlerFactory.Register(gvk, mockHandler)
			convey.So(err, convey.ShouldBeNil)
			reconciler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err = reconciler.Reconcile(ctx, instanceSet, indexer)
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
			workLoadHandlerFactory := NewWorkLoadHandlerFactory()
			err := workLoadHandlerFactory.Register(gvk, mockHandler)
			convey.So(err, convey.ShouldBeNil)
			reconciler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)

			patches := gomonkey.ApplyMethodReturn(reconciler.PodGroupManager, "GetOrCreatePodGroupForInstance", true, nil)
			defer patches.Reset()

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err = reconciler.Reconcile(ctx, instanceSet, indexer)
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
			workLoadHandlerFactory := NewWorkLoadHandlerFactory()
			err := workLoadHandlerFactory.Register(gvk, mockHandler)
			convey.So(err, convey.ShouldBeNil)
			reconciler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)

			mockErr := errors.New("failed to create podgroup")
			patches := gomonkey.ApplyMethodReturn(reconciler.PodGroupManager,
				"GetOrCreatePodGroupForInstance", false, mockErr)
			defer patches.Reset()

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err = reconciler.Reconcile(ctx, instanceSet, indexer)
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
			workLoadHandlerFactory := NewWorkLoadHandlerFactory()
			err := workLoadHandlerFactory.Register(gvk, mockHandler)
			convey.So(err, convey.ShouldBeNil)
			reconciler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err = reconciler.Reconcile(ctx, instanceSet, indexer)
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
			workLoadHandlerFactory := NewWorkLoadHandlerFactory()
			err := workLoadHandlerFactory.Register(gvk, mockHandler)
			convey.So(err, convey.ShouldBeNil)
			reconciler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)

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
			workLoadHandlerFactory := NewWorkLoadHandlerFactory()
			err := workLoadHandlerFactory.Register(gvk, mockHandler)
			convey.So(err, convey.ShouldBeNil)
			reconciler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			_, err = reconciler.InstanceReady(ctx, instanceSet, indexer)
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
			workLoadHandlerFactory := NewWorkLoadHandlerFactory()
			err := workLoadHandlerFactory.Register(gvk, mockHandler)
			convey.So(err, convey.ShouldBeNil)
			reconciler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err = reconciler.DeleteExtraInstances(ctx, instanceSet, indexer)
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
			workLoadHandlerFactory := NewWorkLoadHandlerFactory()
			err := workLoadHandlerFactory.Register(gvk, mockHandler)
			convey.So(err, convey.ShouldBeNil)
			reconciler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err = reconciler.DeleteExtraInstances(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestWorkLoadReconcilerGetInstanceIndices tests the GetInstanceIndices method of WorkLoadReconciler.
func TestWorkLoadReconcilerGetInstanceIndices(t *testing.T) {
	convey.Convey("Test WorkLoadReconciler GetInstanceIndices method", t, func() {
		convey.Convey("Should include terminating instances and skip missing or invalid index labels", func() {
			const ns = "default"
			instanceSet := CreateTestInstanceSet("test-instance", ns, int32(4))

			deploy0 := CreateTestDeployment("test-service-test-role-0", ns, 1)
			deploy0.Labels[common.InstanceIndexLabelKey] = "0"
			deploy2 := CreateTestDeployment("test-service-test-role-2", ns, 1)
			deploy2.Labels[common.InstanceIndexLabelKey] = "2"
			// Terminating instance: deletion timestamp set, but still listed and its index counted.
			terminating := CreateTestDeployment("test-service-test-role-3", ns, 1)
			terminating.Labels[common.InstanceIndexLabelKey] = "3"
			deletionTime := metav1.Now()
			terminating.DeletionTimestamp = &deletionTime
			terminating.Finalizers = []string{"test/finalizer"}
			// Skipped instances: missing or illegal index label.
			noIndex := CreateTestDeployment("test-service-test-role-noidx", ns, 1)
			delete(noIndex.Labels, common.InstanceIndexLabelKey)
			invalidIndex := CreateTestDeployment("test-service-test-role-bad", ns, 1)
			invalidIndex.Labels[common.InstanceIndexLabelKey] = "abc"

			fakeClient := NewFakeClient(deploy0, deploy2, terminating, noIndex, invalidIndex).Build()
			reconciler := NewWorkLoadReconciler(fakeClient)
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			workLoadHandlerFactory := NewWorkLoadHandlerFactory()
			err := workLoadHandlerFactory.Register(gvk, NewDeploymentHandler(fakeClient))
			convey.So(err, convey.ShouldBeNil)
			reconciler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			indices, err := reconciler.GetInstanceIndices(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
			sort.Ints(indices)
			convey.So(indices, convey.ShouldResemble, []int{0, 2, 3})
		})

		convey.Convey("Should return error when the workload handler is not registered", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			indices, err := reconciler.GetInstanceIndices(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(indices, convey.ShouldBeNil)
		})
	})
}

// TestGetWorkLoadReconciler tests the GetWorkLoadReconciler method of WorkLoadReconciler.
func TestGetWorkLoadReconciler(t *testing.T) {
	convey.Convey("Test WorkLoadReconciler GetWorkLoadReconciler method", t, func() {
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
			workLoadHandlerFactory := NewWorkLoadHandlerFactory()
			err := workLoadHandlerFactory.Register(gvk, mockHandler)
			convey.So(err, convey.ShouldBeNil)
			reconciler.SetWorkLoadHandlerFactory(workLoadHandlerFactory)
			handler, err := reconciler.GetWorkLoadReconciler(instanceSet)
			convey.So(err, convey.ShouldBeNil)
			convey.So(handler, convey.ShouldEqual, mockHandler)
		})

		convey.Convey("Should return error when handler not registered", func() {
			fakeClient := NewFakeClient().Build()
			reconciler := NewWorkLoadReconciler(fakeClient)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test", "default", replicas)
			handler, err := reconciler.GetWorkLoadReconciler(instanceSet)
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
			spec := newPodGroupSpec(normalReplicas, nil)

			convey.So(spec.MinMember, convey.ShouldEqual, int32(3))
			convey.So(spec.MinResources, convey.ShouldBeNil)
		})

		convey.Convey("Should handle zero replicas", func() {
			zeroReplicas := int32(0)
			spec := newPodGroupSpec(zeroReplicas, nil)

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

func (m *mockWorkLoadHandler) DeleteWorkLoad(
	ctx context.Context,
	selectLabels map[string]string,
	namespace string,
	filters ...WorkLoadFilter) error {
	return nil
}

func (m *mockWorkLoadHandler) ListWorkLoad(
	ctx context.Context,
	selectLabels map[string]string,
	namespace string,
	filters ...WorkLoadFilter) ([]WorkLoadInterface, error) {
	return nil, nil
}

func (m *mockWorkLoadHandler) UpdateWorkLoad(
	ctx context.Context,
	selectLabels map[string]string,
	namespace string,
	updater WorkloadUpdater,
	filters ...WorkLoadFilter) error {
	return nil
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

func (m *mockWorkLoadHandler) GetMinResources(runtime.RawExtension) (*corev1.ResourceList, error) {
	return nil, nil
}
