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
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

// TestNewStatefulSetHandler tests the NewStatefulSetHandler function.
func TestNewStatefulSetHandler(t *testing.T) {
	convey.Convey("Test NewStatefulSetHandler function", t, func() {
		convey.Convey("Should create a new StatefulSetHandler", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			convey.So(handler, convey.ShouldNotBeNil)
			convey.So(handler.client, convey.ShouldEqual, fakeClient)
		})
	})
}

// TestStatefulSetHandlerCheckOrCreateWorkLoad tests the CheckOrCreateWorkLoad method of StatefulSetHandler.
func TestStatefulSetHandlerCheckOrCreateWorkLoad(t *testing.T) {
	convey.Convey("Test StatefulSetHandler CheckOrCreateWorkLoad method", t, func() {
		convey.Convey("Should successfully create StatefulSet and Service", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", nil)
			defer patches.Reset()
			patches2 := gomonkey.ApplyPrivateMethod(handler, "createStatefulSet",
				func(ctx context.Context, instanceSet *v1.InstanceSet, indexer common.InstanceIndexer) error {
					return nil
				})
			defer patches2.Reset()

			ctx := context.Background()
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			indexer := GetTestIndexer("test-service", "test-role", "0")
			err := handler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should handle existing Service", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			indexer := GetTestIndexer("test-service", "test-role", "0")
			service := CreateTestService("service-test-service-test-role-0", "default")
			fakeClient := NewFakeClient().WithObjects(service).Build()
			handler := NewStatefulSetHandler(fakeClient)

			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", nil)
			defer patches.Reset()
			patches2 := gomonkey.ApplyPrivateMethod(handler, "createStatefulSet",
				func(ctx context.Context, instanceSet *v1.InstanceSet, indexer common.InstanceIndexer) error {
					return nil
				})
			defer patches2.Reset()

			ctx := context.Background()
			err := handler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestStatefulSetHandlerCheckOrCreateWorkLoad2 tests the CheckOrCreateWorkLoad method of StatefulSetHandler.
func TestStatefulSetHandlerCheckOrCreateWorkLoad2(t *testing.T) {
	convey.Convey("Test StatefulSetHandler CheckOrCreateWorkLoad method", t, func() {
		convey.Convey("Should return error when getting Service fails", func() {
			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")

			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			mockErr := errors.New("failed to get service")
			patches := gomonkey.ApplyMethodReturn(handler.client, "Get", mockErr)
			defer patches.Reset()

			ctx := context.Background()
			err := handler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(common.IsRequeueError(err), convey.ShouldBeTrue)
		})

		convey.Convey("Should return error when creating Service fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			ctx := context.Background()
			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			mockErr := errors.New("failed to create service")
			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", mockErr)
			defer patches.Reset()

			err := handler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(common.IsRequeueError(err), convey.ShouldBeTrue)
		})
	})
}

// TestStatefulSetHandlerCheckOrCreateWorkLoad3 tests the CheckOrCreateWorkLoad method of StatefulSetHandler.
func TestStatefulSetHandlerCheckOrCreateWorkLoad3(t *testing.T) {
	convey.Convey("Test StatefulSetHandler CheckOrCreateWorkLoad method", t, func() {
		convey.Convey("Should return error when listing StatefulSets fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			mockErr := errors.New("failed to list statefulsets")
			patches := gomonkey.ApplyMethodReturn(handler.client, "List", mockErr)
			defer patches.Reset()

			ctx := context.Background()
			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			err := handler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when creating StatefulSet fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			mockErr := errors.New("failed to create statefulset")
			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", mockErr)
			defer patches.Reset()
			patches2 := gomonkey.ApplyPrivateMethod(
				handler,
				"createStatefulSet",
				func(ctx context.Context, instanceSet *v1.InstanceSet, indexer common.InstanceIndexer) error {
					return mockErr
				})
			defer patches2.Reset()

			ctx := context.Background()
			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			err := handler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(common.IsRequeueError(err), convey.ShouldBeTrue)
		})
	})
}

// TestCreateStatefulSet tests the createStatefulSet method of StatefulSetHandler.
func TestCreateStatefulSet(t *testing.T) {
	convey.Convey("Test StatefulSetHandler createStatefulSet method", t, func() {
		convey.Convey("Should successfully create StatefulSet", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			indexer := GetTestIndexer("test-service", "test-role", "0")
			statefulSetSpec := getTestStatefulSetSpec()
			specBytes, err := json.Marshal(statefulSetSpec)
			convey.So(err, convey.ShouldBeNil)
			instanceSet.Spec.InstanceSpec = runtime.RawExtension{Raw: specBytes}
			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", nil)
			defer patches.Reset()
			ctx := context.Background()
			err = handler.createStatefulSet(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when creating StatefulSet fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			indexer := GetTestIndexer("test-service", "test-role", "0")
			statefulSetSpec := getTestStatefulSetSpec()
			specBytes, err := json.Marshal(statefulSetSpec)
			convey.So(err, convey.ShouldBeNil)
			instanceSet.Spec.InstanceSpec = runtime.RawExtension{Raw: specBytes}
			ctx := context.Background()
			mockErr := errors.New("failed to create statefulset")
			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", mockErr)
			defer patches.Reset()
			err = handler.createStatefulSet(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(common.IsRequeueError(err), convey.ShouldBeTrue)
		})

		convey.Convey("Should return error when parsing StatefulSetSpec fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			indexer := GetTestIndexer("test-service", "test-role", "0")
			instanceSet.Spec.InstanceSpec = runtime.RawExtension{Raw: []byte("invalid json")}
			ctx := context.Background()
			err := handler.createStatefulSet(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestStatefulSetHandlerCreateService tests the createService method of StatefulSetHandler.
func TestStatefulSetHandlerCreateService(t *testing.T) {
	convey.Convey("Test StatefulSetHandler createService method", t, func() {
		convey.Convey("Should successfully create Service", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", nil)
			defer patches.Reset()

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err := handler.createService(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when creating Service fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			mockErr := errors.New("failed to create service")
			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", mockErr)
			defer patches.Reset()

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err := handler.createService(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(common.IsRequeueError(err), convey.ShouldBeTrue)
		})
	})
}

// TestStatefulSetHandlerDeleteExtraWorkLoad tests the DeleteExtraWorkLoad method of StatefulSetHandler.
func TestStatefulSetHandlerDeleteExtraWorkLoad(t *testing.T) {
	convey.Convey("Test StatefulSetHandler DeleteExtraWorkLoad method", t, func() {
		convey.Convey("Should successfully delete extra StatefulSets", func() {
			extraStatefulSet := CreateTestStatefulSet("test-service-test-role-5", "default", 1)
			extraStatefulSet.Labels[common.InstanceIndexLabelKey] = "5"
			fakeClient := NewFakeClient().WithObjects(extraStatefulSet).Build()
			handler := NewStatefulSetHandler(fakeClient)

			patches := gomonkey.ApplyMethodReturn(
				handler.client,
				"Delete",
				nil)
			defer patches.Reset()

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			replicas := 3
			err := handler.DeleteExtraWorkLoad(ctx, indexer, replicas)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when listing StatefulSets fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			mockErr := errors.New("failed to list statefulsets")
			patches := gomonkey.ApplyMethodReturn(
				handler.client,
				"List",
				mockErr)
			defer patches.Reset()

			ctx := context.Background()
			indexer := GetTestIndexer("test-service", "test-role", "0")
			replicas := 3
			err := handler.DeleteExtraWorkLoad(ctx, indexer, replicas)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestStatefulSetHandlerDeleteExtraWorkLoad2 tests the DeleteExtraWorkLoad method of StatefulSetHandler.
func TestStatefulSetHandlerDeleteExtraWorkLoad2(t *testing.T) {
	convey.Convey("Test StatefulSetHandler DeleteExtraWorkLoad method", t, func() {
		convey.Convey("Should return error when deleting StatefulSet fails", func() {
			extraStatefulSet := CreateTestStatefulSet("test-service-test-role-5", "default", 1)
			extraStatefulSet.Labels[common.InstanceIndexLabelKey] = "5"
			fakeClient := NewFakeClient().WithObjects(extraStatefulSet).Build()
			handler := NewStatefulSetHandler(fakeClient)

			mockErr := errors.New("failed to delete statefulset")
			patches := gomonkey.ApplyMethodReturn(
				handler.client,
				"Delete",
				mockErr)
			defer patches.Reset()

			ctx := context.Background()
			indexer := GetTestIndexer("test-service", "test-role", "0")
			replicas := 3
			err := handler.DeleteExtraWorkLoad(ctx, indexer, replicas)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should skip StatefulSets without instance index label", func() {
			statefulSet := CreateTestStatefulSet("test-statefulset", "default", 1)
			delete(statefulSet.Labels, common.InstanceIndexLabelKey)
			fakeClient := NewFakeClient().WithObjects(statefulSet).Build()
			handler := NewStatefulSetHandler(fakeClient)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()

			replicas := 3
			err := handler.DeleteExtraWorkLoad(ctx, indexer, replicas)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestStatefulSetHandlerGetWorkLoadReadyReplicas tests the GetWorkLoadReadyReplicas method of StatefulSetHandler.
func TestStatefulSetHandlerGetWorkLoadReadyReplicas(t *testing.T) {
	convey.Convey("Test StatefulSetHandler GetWorkLoadReadyReplicas method", t, func() {
		convey.Convey("Should return correct number of ready replicas", func() {
			readyStatefulSet := CreateTestStatefulSet("test-service-test-role-0", "default", 1)
			fakeClient := NewFakeClient().WithObjects(readyStatefulSet).Build()
			handler := NewStatefulSetHandler(fakeClient)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			readyReplicas, err := handler.GetWorkLoadReadyReplicas(ctx, indexer)
			convey.So(err, convey.ShouldBeNil)
			convey.So(readyReplicas, convey.ShouldEqual, 1)
		})

		convey.Convey("Should return 0 for not ready StatefulSets", func() {
			notReadyStatefulSet := CreateTestStatefulSet("test-service-test-role-0",
				"default", 1)
			notReadyStatefulSet.Status.ReadyReplicas = 0
			fakeClient := NewFakeClient().WithObjects(notReadyStatefulSet).Build()
			handler := NewStatefulSetHandler(fakeClient)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			readyReplicas, err := handler.GetWorkLoadReadyReplicas(ctx, indexer)
			convey.So(err, convey.ShouldBeNil)
			convey.So(readyReplicas, convey.ShouldEqual, 0)
		})

		convey.Convey("Should return error when listing StatefulSets fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			mockErr := errors.New("failed to list statefulsets")
			patches := gomonkey.ApplyMethodReturn(
				handler.client,
				"List",
				mockErr)
			defer patches.Reset()

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			readyReplicas, err := handler.GetWorkLoadReadyReplicas(ctx, indexer)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(readyReplicas, convey.ShouldEqual, 0)
		})
	})
}

// TestStatefulSetHandlerListWorkLoads tests the ListWorkLoads method of StatefulSetHandler.
func TestStatefulSetHandlerListWorkLoads(t *testing.T) {
	convey.Convey("Test StatefulSetHandler ListWorkLoads method", t, func() {
		convey.Convey("Should successfully list StatefulSets", func() {
			statefulSet := CreateTestStatefulSet("test-statefulset", "default", 1)
			fakeClient := NewFakeClient().WithObjects(statefulSet).Build()
			handler := NewStatefulSetHandler(fakeClient)
			selectLabels := map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
			}
			ctx := context.Background()
			result, err := handler.ListWorkLoads(ctx, selectLabels, "default")
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(len(result.Items), convey.ShouldEqual, 1)
		})

		convey.Convey("Should return empty list when no StatefulSets found", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			selectLabels := map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
			}
			ctx := context.Background()
			result, err := handler.ListWorkLoads(ctx, selectLabels, "default")
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(len(result.Items), convey.ShouldEqual, 0)
		})

		convey.Convey("Should return error when listing fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			selectLabels := map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
			}
			ctx := context.Background()

			mockErr := errors.New("failed to list statefulsets")
			patches := gomonkey.ApplyMethodReturn(handler.client, "List", mockErr)
			defer patches.Reset()

			result, err := handler.ListWorkLoads(ctx, selectLabels, "default")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestStatefulSetHandlerValidate tests the Validate method of StatefulSetHandler.
func TestStatefulSetHandlerValidate(t *testing.T) {
	convey.Convey("Test StatefulSetHandler Validate method", t, func() {
		convey.Convey("Should successfully validate valid StatefulSetSpec", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			statefulSetSpec := getTestStatefulSetSpec()
			specBytes, err := json.Marshal(statefulSetSpec)
			convey.So(err, convey.ShouldBeNil)
			spec := runtime.RawExtension{Raw: specBytes}

			err = handler.Validate(spec)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error for invalid StatefulSetSpec", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			spec := runtime.RawExtension{Raw: []byte("invalid json")}

			err := handler.Validate(spec)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error for empty StatefulSetSpec", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			spec := runtime.RawExtension{Raw: []byte{}}

			err := handler.Validate(spec)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestStatefulSetHandlerGetReplicas tests the GetReplicas method of StatefulSetHandler.
func TestStatefulSetHandlerGetReplicas(t *testing.T) {
	convey.Convey("Test StatefulSetHandler GetReplicas method", t, func() {
		convey.Convey("Should return correct replicas count", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			replicas := int32(3)
			statefulSetSpec := &appsv1.StatefulSetSpec{
				Replicas: &replicas,
			}
			specBytes, err := json.Marshal(statefulSetSpec)
			convey.So(err, convey.ShouldBeNil)
			spec := runtime.RawExtension{Raw: specBytes}

			result, err := handler.GetReplicas(spec)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldEqual, int32(3))
		})

		convey.Convey("Should return default replicas when replicas is nil", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			statefulSetSpec := &appsv1.StatefulSetSpec{
				Replicas: nil,
			}
			specBytes, err := json.Marshal(statefulSetSpec)
			convey.So(err, convey.ShouldBeNil)
			spec := runtime.RawExtension{Raw: specBytes}

			result, err := handler.GetReplicas(spec)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldEqual, common.DefaultReplicas)
		})

		convey.Convey("Should return error for invalid StatefulSetSpec", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			spec := runtime.RawExtension{Raw: []byte("invalid json")}

			result, err := handler.GetReplicas(spec)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldEqual, common.DefaultReplicas)
		})
	})
}

// TestIsStatefulsetReady tests the isStatefulsetReady function.
func TestIsStatefulsetReady(t *testing.T) {
	convey.Convey("Test isStatefulsetReady function", t, func() {
		convey.Convey("Should return true for ready StatefulSet", func() {
			statefulSet := CreateTestStatefulSet("test-statefulset", "default", 1)

			result := isStatefulsetReady(*statefulSet)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("Should return false when ready replicas don't match", func() {
			statefulSet := CreateTestStatefulSet("test-statefulset", "default", 1)
			statefulSet.Status.ReadyReplicas = 0

			result := isStatefulsetReady(*statefulSet)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when generation is not latest", func() {
			statefulSet := CreateTestStatefulSet("test-statefulset", "default", 1)
			statefulSet.Generation = 2
			statefulSet.Status.ObservedGeneration = 1

			result := isStatefulsetReady(*statefulSet)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when update is in progress", func() {
			statefulSet := CreateTestStatefulSet("test-statefulset", "default", 1)
			statefulSet.Status.CurrentRevision = "v1"
			statefulSet.Status.UpdateRevision = "v2"

			result := isStatefulsetReady(*statefulSet)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return true when revisions match", func() {
			statefulSet := CreateTestStatefulSet("test-statefulset", "default", 1)
			statefulSet.Status.CurrentRevision = "v1"
			statefulSet.Status.UpdateRevision = "v1"

			result := isStatefulsetReady(*statefulSet)
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

// TestParseStatefulSetWithScheme tests the parseStatefulSetWithScheme
// method of StatefulSetHandler.
func TestParseStatefulSetWithScheme(t *testing.T) {
	convey.Convey("Test StatefulSetHandler parseStatefulSetWithScheme method", t, func() {
		convey.Convey("Should successfully parse StatefulSetSpec", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			statefulSetSpec := getTestStatefulSetSpec()
			specBytes, err := json.Marshal(statefulSetSpec)
			convey.So(err, convey.ShouldBeNil)
			raw := runtime.RawExtension{Raw: specBytes}

			result, err := handler.parseStatefulSetWithScheme(raw)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(*result.Replicas, convey.ShouldEqual, int32(1))
		})

		convey.Convey("Should return error for invalid JSON", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			raw := runtime.RawExtension{Raw: []byte("invalid json")}

			result, err := handler.parseStatefulSetWithScheme(raw)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("Should return error for empty raw extension", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			raw := runtime.RawExtension{Raw: []byte{}}

			result, err := handler.parseStatefulSetWithScheme(raw)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

func createTestStatefulSetWorkLoadFilter() WorkLoadFilter {
	return func(workload WorkLoadInterface) bool {
		return workload.GetWorkLoadReplicas() > 0
	}
}

func createTestStatefulSetWorkloadUpdater() WorkloadUpdater {
	return func(workload WorkLoadInterface) {
		workload.SetWorkLoadObjMeta(metav1.ObjectMeta{
			Labels: map[string]string{"updated": "true"},
		})
	}
}

func createTestStatefulSetWorkLoad() *StatefulSetWorkLoad {
	statefulset := createTestStatefulSet("test-statefulset", "default", 1)
	statefulset.Status.ReadyReplicas = 1
	statefulset.Status.UpdatedReplicas = 1
	statefulset.Status.ObservedGeneration = 1
	return &StatefulSetWorkLoad{StatefulSet: statefulset}
}

func TestStatefulSetWorkLoadSetWorkLoadObjMeta(t *testing.T) {
	convey.Convey("Test StatefulSetWorkLoad SetWorkLoadObjMeta method", t, func() {
		convey.Convey("Should set object meta successfully", func() {
			workload := &StatefulSetWorkLoad{
				StatefulSet: &appsv1.StatefulSet{},
			}
			objectMeta := metav1.ObjectMeta{
				Name:      "test-statefulset",
				Namespace: "default",
				Labels:    map[string]string{"key": "value"},
			}

			workload.SetWorkLoadObjMeta(objectMeta)

			convey.So(workload.Name, convey.ShouldEqual, "test-statefulset")
			convey.So(workload.Namespace, convey.ShouldEqual, "default")
			convey.So(workload.Labels["key"], convey.ShouldEqual, "value")
		})

		convey.Convey("Should handle nil workload", func() {
			var workload *StatefulSetWorkLoad
			objectMeta := metav1.ObjectMeta{Name: "test"}

			workload.SetWorkLoadObjMeta(objectMeta)

			convey.So(workload, convey.ShouldBeNil)
		})
	})
}

func TestStatefulSetWorkLoadGetWorkLoadObjMeta(t *testing.T) {
	convey.Convey("Test StatefulSetWorkLoad GetWorkLoadObjMeta method", t, func() {
		convey.Convey("Should get object meta successfully", func() {
			workload := &StatefulSetWorkLoad{
				StatefulSet: &appsv1.StatefulSet{},
			}
			expectedMeta := metav1.ObjectMeta{
				Name:      "test-statefulset",
				Namespace: "default",
			}
			workload.SetWorkLoadObjMeta(expectedMeta)

			result := workload.GetWorkLoadObjMeta()

			convey.So(result.Name, convey.ShouldEqual, "test-statefulset")
			convey.So(result.Namespace, convey.ShouldEqual, "default")
		})

		convey.Convey("Should return empty meta for nil workload", func() {
			var workload *StatefulSetWorkLoad

			result := workload.GetWorkLoadObjMeta()

			convey.So(result, convey.ShouldResemble, metav1.ObjectMeta{})
		})
	})
}

func TestStatefulSetWorkLoadIsWorkloadReady(t *testing.T) {
	convey.Convey("Test StatefulSetWorkLoad IsWorkLoadReady method", t, func() {
		convey.Convey("Should return true for ready workload", func() {
			workload := createTestStatefulSetWorkLoad()

			result := workload.IsWorkLoadReady()

			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("Should return false when replicas don't match", func() {
			workload := createTestStatefulSetWorkLoad()
			workload.Status.ReadyReplicas = 0

			result := workload.IsWorkLoadReady()

			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when generation is not latest", func() {
			workload := createTestStatefulSetWorkLoad()
			workload.Generation = 2
			workload.Status.ObservedGeneration = 1

			result := workload.IsWorkLoadReady()

			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when rolling update in progress", func() {
			workload := createTestStatefulSetWorkLoad()
			workload.Status.CurrentRevision = "rev-1"
			workload.Status.UpdateRevision = "rev-2"

			result := workload.IsWorkLoadReady()

			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false for nil workload", func() {
			var workload *StatefulSetWorkLoad

			result := workload.IsWorkLoadReady()

			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestStatefulSetWorkLoadGetWorkLoadReplicas(t *testing.T) {
	convey.Convey("Test StatefulSetWorkLoad GetWorkLoadReplicas method", t, func() {
		convey.Convey("Should return correct replicas", func() {
			workload := createTestStatefulSetWorkLoad()
			replicas := int32(5)
			workload.Spec.Replicas = &replicas

			result := workload.GetWorkLoadReplicas()

			convey.So(result, convey.ShouldEqual, int32(5))
		})

		convey.Convey("Should return default replicas when replicas is nil", func() {
			workload := createTestStatefulSetWorkLoad()
			workload.Spec.Replicas = nil

			result := workload.GetWorkLoadReplicas()

			convey.So(result, convey.ShouldEqual, common.DefaultReplicas)
		})

		convey.Convey("Should return default replicas for nil workload", func() {
			var workload *StatefulSetWorkLoad

			result := workload.GetWorkLoadReplicas()

			convey.So(result, convey.ShouldEqual, common.DefaultReplicas)
		})
	})
}

func TestStatefulSetHandlerListWorkLoad(t *testing.T) {
	convey.Convey("Test StatefulSetHandler ListWorkLoad method", t, func() {
		selectLabels := map[string]string{
			common.InferServiceNameLabelKey: "test-service",
			common.InstanceSetNameLabelKey:  "test-role",
		}

		convey.Convey("Should list workloads successfully without filters", func() {
			statefulset := createTestStatefulSet("test-statefulset", "default", 1)
			statefulset.Spec.Replicas = func() *int32 { r := int32(2); return &r }()
			fakeClient := NewFakeClient().WithObjects(statefulset).Build()
			handler := NewStatefulSetHandler(fakeClient)

			result, err := handler.ListWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 1)
			convey.So(result[0].GetWorkLoadReplicas(), convey.ShouldEqual, int32(2))
		})

		convey.Convey("Should filter workloads with filter function", func() {
			statefulset := createTestStatefulSet("test-statefulset", "default", 1)
			statefulset.Spec.Replicas = func() *int32 { r := int32(0); return &r }()
			fakeClient := NewFakeClient().WithObjects(statefulset).Build()
			handler := NewStatefulSetHandler(fakeClient)
			filter := createTestStatefulSetWorkLoadFilter()

			result, err := handler.ListWorkLoad(context.Background(), selectLabels, "default", filter)

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("Should return empty list when no statefulsets match", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)

			result, err := handler.ListWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("Should return error when selector creation fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			invalidLabels := map[string]string{"invalid/label": "value"}
			patch := gomonkey.ApplyFuncReturn(metav1.LabelSelectorAsSelector, nil, fmt.Errorf("create selector failed"))
			defer patch.Reset()
			result, err := handler.ListWorkLoad(context.Background(), invalidLabels, "default")

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when client List fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			mockErr := errors.New("failed to list statefulsets")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "List", mockErr)
			defer patches.Reset()

			result, err := handler.ListWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

func TestStatefulSetHandlerDeleteWorkLoad(t *testing.T) {
	convey.Convey("Test StatefulSetHandler DeleteWorkLoad method", t, func() {
		selectLabels := map[string]string{
			common.InferServiceNameLabelKey: "test-service",
			common.InstanceSetNameLabelKey:  "test-role",
		}

		convey.Convey("Should delete workloads successfully without filters", func() {
			statefulset := createTestStatefulSet("test-statefulset", "default", 1)
			fakeClient := NewFakeClient().WithObjects(statefulset).Build()
			handler := NewStatefulSetHandler(fakeClient)
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Delete", nil)
			defer patches.Reset()

			err := handler.DeleteWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should skip workloads that don't match filter", func() {
			statefulset := createTestStatefulSet("test-statefulset", "default", 0)
			fakeClient := NewFakeClient().WithObjects(statefulset).Build()
			handler := NewStatefulSetHandler(fakeClient)
			filter := createTestStatefulSetWorkLoadFilter()

			err := handler.DeleteWorkLoad(context.Background(), selectLabels, "default", filter)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when ListWorkLoads fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			mockErr := errors.New("failed to list statefulsets")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "List", mockErr)
			defer patches.Reset()

			err := handler.DeleteWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when client Delete fails", func() {
			statefulset := createTestStatefulSet("test-statefulset", "default", 1)
			fakeClient := NewFakeClient().WithObjects(statefulset).Build()
			handler := NewStatefulSetHandler(fakeClient)
			mockErr := errors.New("failed to delete statefulset")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Delete", mockErr)
			defer patches.Reset()

			err := handler.DeleteWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestStatefulSetHandlerUpdateWorkLoad(t *testing.T) {
	convey.Convey("Test StatefulSetHandler UpdateWorkLoad method", t, func() {
		selectLabels := map[string]string{
			common.InferServiceNameLabelKey: "test-service",
			common.InstanceSetNameLabelKey:  "test-role",
		}

		convey.Convey("Should update workloads successfully without filters", func() {
			statefulset := createTestStatefulSet("test-statefulset", "default", 1)
			fakeClient := NewFakeClient().WithObjects(statefulset).Build()
			handler := NewStatefulSetHandler(fakeClient)
			updater := createTestStatefulSetWorkloadUpdater()
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Update", nil)
			defer patches.Reset()

			err := handler.UpdateWorkLoad(context.Background(), selectLabels, "default", updater)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should skip workloads that don't match filter", func() {
			statefulset := createTestStatefulSet("test-statefulset", "default", 0)
			fakeClient := NewFakeClient().WithObjects(statefulset).Build()
			handler := NewStatefulSetHandler(fakeClient)
			filter := createTestStatefulSetWorkLoadFilter()
			updater := createTestStatefulSetWorkloadUpdater()

			err := handler.UpdateWorkLoad(context.Background(), selectLabels, "default", updater, filter)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when ListWorkLoads fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewStatefulSetHandler(fakeClient)
			mockErr := errors.New("failed to list statefulsets")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "List", mockErr)
			defer patches.Reset()
			updater := createTestStatefulSetWorkloadUpdater()

			err := handler.UpdateWorkLoad(context.Background(), selectLabels, "default", updater)

			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when client Update fails", func() {
			statefulset := createTestStatefulSet("test-statefulset", "default", 1)
			fakeClient := NewFakeClient().WithObjects(statefulset).Build()
			handler := NewStatefulSetHandler(fakeClient)
			mockErr := errors.New("failed to update statefulset")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Update", mockErr)
			defer patches.Reset()
			updater := createTestStatefulSetWorkloadUpdater()

			err := handler.UpdateWorkLoad(context.Background(), selectLabels, "default", updater)

			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func createTestStatefulSet(name, namespace string, replicas int32) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "test-container", Image: "test-image"},
					},
				},
			},
			ServiceName: "test-service",
		},
	}
}

func getTestStatefulSetSpec() *appsv1.StatefulSetSpec {
	replicas := int32(1)
	return &appsv1.StatefulSetSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"app": "test"},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app": "test"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "test-container", Image: "test-image"},
				},
			},
		},
		ServiceName: "test-service",
	}
}
