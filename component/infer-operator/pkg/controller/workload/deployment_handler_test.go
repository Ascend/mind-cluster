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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// TestNewDeploymentHandler tests the NewDeploymentHandler function.
func TestNewDeploymentHandler(t *testing.T) {
	convey.Convey("Test NewDeploymentHandler function", t, func() {
		convey.Convey("Should create a new DeploymentHandler", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			convey.So(handler, convey.ShouldNotBeNil)
			convey.So(handler.client, convey.ShouldEqual, fakeClient)
		})
	})
}

// TestDeploymentHandlerCheckOrCreateWorkLoad tests the CheckOrCreateWorkLoad method of DeploymentHandler.
func TestDeploymentHandlerCheckOrCreateWorkLoad(t *testing.T) {
	convey.Convey("Test DeploymentHandler CheckOrCreateWorkLoad method", t, func() {
		convey.Convey("Should successfully create Deployment and Service when Services is nil", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			instanceSet.Spec.Services = nil
			indexer := GetTestIndexer("test-service", "test-role", "0")

			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", nil)
			defer patches.Reset()
			patches2 := gomonkey.ApplyPrivateMethod(handler, "createDeployment",
				func(ctx context.Context, instanceSet *v1.InstanceSet, indexer common.InstanceIndexer) error {
					return nil
				})
			defer patches2.Reset()

			ctx := context.Background()
			err := handler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should successfully create Deployment and Service when Services is empty", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			instanceSet.Spec.Services = []v1.ServiceSpec{}
			indexer := GetTestIndexer("test-service", "test-role", "0")

			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", nil)
			defer patches.Reset()
			patches2 := gomonkey.ApplyPrivateMethod(handler, "createDeployment",
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

// TestDeploymentHandlerCheckOrCreateWorkLoad2 tests the CheckOrCreateWorkLoad method of DeploymentHandler.
func TestDeploymentHandlerCheckOrCreateWorkLoad2(t *testing.T) {
	convey.Convey("Test DeploymentHandler CheckOrCreateWorkLoad method", t, func() {
		convey.Convey("Should handle existing Service", func() {
			service := &corev1.Service{}
			indexer := GetTestIndexer("test-service", "test-role", "0")
			service.Name = common.GetServiceNameFromIndexer(indexer)
			fakeClient := NewFakeClient().WithObjects(service).Build()
			handler := NewDeploymentHandler(fakeClient)
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			instanceSet.Spec.Services = nil
			ctx := context.Background()

			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", nil)
			defer patches.Reset()
			patches2 := gomonkey.ApplyPrivateMethod(handler, "createDeployment",
				func(ctx context.Context, instanceSet *v1.InstanceSet, indexer common.InstanceIndexer) error {
					return nil
				})
			defer patches2.Reset()

			err := handler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when getting Service fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			instanceSet.Spec.Services = nil
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()

			mockErr := errors.New("failed to get service")
			patches := gomonkey.ApplyMethodReturn(handler.client, "Get", mockErr)
			defer patches.Reset()

			err := handler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(common.IsRequeueError(err), convey.ShouldBeTrue)
		})
	})
}

// TestDeploymentHandlerCheckOrCreateWorkLoad3 tests the CheckOrCreateWorkLoad method of DeploymentHandler.
func TestDeploymentHandlerCheckOrCreateWorkLoad3(t *testing.T) {
	convey.Convey("Test DeploymentHandler CheckOrCreateWorkLoad method", t, func() {
		convey.Convey("Should return error when creating Service fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			instanceSet.Spec.Services = nil
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			mockErr := errors.New("failed to create service")
			patches3 := gomonkey.ApplyMethodReturn(
				handler.client,
				"Create",
				mockErr)
			defer patches3.Reset()

			err := handler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(common.IsRequeueError(err), convey.ShouldBeTrue)
		})

		convey.Convey("Should return error when listing Deployments fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			instanceSet.Spec.Services = nil
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			patches2 := gomonkey.ApplyMethodReturn(
				handler.client,
				"Create",
				nil)
			defer patches2.Reset()

			mockErr := errors.New("failed to list deployments")
			patches3 := gomonkey.ApplyMethodReturn(
				handler.client,
				"List",
				mockErr)
			defer patches3.Reset()

			err := handler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestDeploymentHandlerCheckOrCreateWorkLoad4 tests the CheckOrCreateWorkLoad method of DeploymentHandler.
func TestDeploymentHandlerCheckOrCreateWorkLoad4(t *testing.T) {
	convey.Convey("Test DeploymentHandler CheckOrCreateWorkLoad method", t, func() {
		convey.Convey("Should skip Service creation when custom Services exist", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			instanceSet.Spec.Services = []v1.ServiceSpec{
				{Name: "custom-service"},
			}
			indexer := GetTestIndexer("test-service", "test-role", "0")

			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", nil)
			defer patches.Reset()
			patches2 := gomonkey.ApplyPrivateMethod(handler, "createDeployment",
				func(ctx context.Context, instanceSet *v1.InstanceSet, indexer common.InstanceIndexer) error {
					return nil
				})
			defer patches2.Reset()

			ctx := context.Background()
			err := handler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when creating Deployment fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			instanceSet.Spec.Services = nil
			indexer := GetTestIndexer("test-service", "test-role", "0")

			ctx := context.Background()
			patches4 := gomonkey.ApplyMethodReturn(handler.client, "Create", nil)
			defer patches4.Reset()
			mockErr := errors.New("failed to create deployment")
			patches2 := gomonkey.ApplyPrivateMethod(handler, "createDeployment",
				func(ctx context.Context, instanceSet *v1.InstanceSet, indexer common.InstanceIndexer) error {
					return mockErr
				})
			defer patches2.Reset()

			err := handler.CheckOrCreateWorkLoad(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestDeploymentHandlerCreateDeployment tests the createDeployment method of DeploymentHandler.
func TestDeploymentHandlerCreateDeployment(t *testing.T) {
	convey.Convey("Test DeploymentHandler createDeployment method", t, func() {
		convey.Convey("Should successfully create Deployment", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			indexer := GetTestIndexer("test-service", "test-role", "0")
			deploymentSpec := getTestDeploymentSpec()
			specBytes, err := json.Marshal(deploymentSpec)
			convey.So(err, convey.ShouldBeNil)
			instanceSet.Spec.InstanceSpec = runtime.RawExtension{Raw: specBytes}
			ctx := context.Background()
			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", nil)
			defer patches.Reset()
			err = handler.createDeployment(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when creating Deployment fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			indexer := GetTestIndexer("test-service", "test-role", "0")
			deploymentSpec := getTestDeploymentSpec()
			specBytes, err := json.Marshal(deploymentSpec)
			convey.So(err, convey.ShouldBeNil)
			instanceSet.Spec.InstanceSpec = runtime.RawExtension{Raw: specBytes}
			ctx := context.Background()
			mockErr := errors.New("failed to create deployment")
			patches := gomonkey.ApplyMethodReturn(handler.client, "Create", mockErr)
			defer patches.Reset()
			err = handler.createDeployment(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(common.IsRequeueError(err), convey.ShouldBeTrue)
		})

		convey.Convey("Should return error when parsing DeploymentSpec fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			indexer := GetTestIndexer("test-service", "test-role", "0")
			instanceSet.Spec.InstanceSpec = runtime.RawExtension{Raw: []byte("invalid json")}
			ctx := context.Background()
			err := handler.createDeployment(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestDeploymentHandlerCreateService tests the createService method of DeploymentHandler.
func TestDeploymentHandlerCreateService(t *testing.T) {
	convey.Convey("Test DeploymentHandler createService method", t, func() {
		convey.Convey("Should successfully create Service", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")

			ctx := context.Background()

			patches := gomonkey.ApplyMethodReturn(
				handler.client,
				"Create",
				nil)
			defer patches.Reset()

			err := handler.createService(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when creating Service fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")

			ctx := context.Background()

			mockErr := errors.New("failed to create service")
			patches := gomonkey.ApplyMethodReturn(
				handler.client,
				"Create",
				mockErr)
			defer patches.Reset()

			err := handler.createService(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(common.IsRequeueError(err), convey.ShouldBeTrue)
		})
	})
}

// TestDeploymentHandlerDeleteExtraWorkLoad tests the DeleteExtraWorkLoad method of DeploymentHandler.
func TestDeploymentHandlerDeleteExtraWorkLoad(t *testing.T) {
	convey.Convey("Test DeploymentHandler DeleteExtraWorkLoad method", t, func() {
		convey.Convey("Should successfully delete extra Deployments", func() {
			extraDeployment := CreateTestDeployment("test-service-test-role-5", "default", 1)
			extraDeployment.Labels[common.InstanceIndexLabelKey] = "5"
			fakeClient := NewFakeClient().WithObjects(extraDeployment).Build()
			handler := NewDeploymentHandler(fakeClient)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			patches2 := gomonkey.ApplyMethodReturn(handler.client, "Delete", nil)
			defer patches2.Reset()
			replicas := 3
			err := handler.DeleteExtraWorkLoad(ctx, indexer, replicas)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when listing Deployments fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			mockErr := errors.New("failed to list deployments")
			patches := gomonkey.ApplyMethodReturn(handler.client, "List", mockErr)
			defer patches.Reset()
			ctx := context.Background()
			replicas := 3
			indexer := GetTestIndexer("test-service", "test-role", "0")
			err := handler.DeleteExtraWorkLoad(ctx, indexer, replicas)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when deleting Deployment fails", func() {
			extraDeployment := CreateTestDeployment("test-service-test-role-5", "default", 1)
			extraDeployment.Labels[common.InstanceIndexLabelKey] = "5"
			fakeClient := NewFakeClient().WithObjects(extraDeployment).Build()
			handler := NewDeploymentHandler(fakeClient)
			mockErr := errors.New("failed to delete deployment")
			patches2 := gomonkey.ApplyMethodReturn(handler.client, "Delete", mockErr)
			defer patches2.Reset()
			ctx := context.Background()
			replicas := 3
			indexer := GetTestIndexer("test-service", "test-role", "0")
			err := handler.DeleteExtraWorkLoad(ctx, indexer, replicas)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestDeploymentHandlerGetWorkLoadReadyReplicas tests the GetWorkLoadReadyReplicas method of DeploymentHandler.
func TestDeploymentHandlerGetWorkLoadReadyReplicas(t *testing.T) {
	convey.Convey("Test DeploymentHandler GetWorkLoadReadyReplicas method", t, func() {
		convey.Convey("Should return correct number of ready replicas", func() {
			readyDeployment := CreateTestDeployment("test-service-test-role-0", "default", 1)
			fakeClient := NewFakeClient().WithObjects(readyDeployment).Build()
			handler := NewDeploymentHandler(fakeClient)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			readyReplicas, err := handler.GetWorkLoadReadyReplicas(ctx, indexer)
			convey.So(err, convey.ShouldBeNil)
			convey.So(readyReplicas, convey.ShouldEqual, 1)
		})

		convey.Convey("Should return 0 for not ready deployments", func() {
			notReadyDeployment := CreateTestDeployment("test-service-test-role-0", "default", 1)
			notReadyDeployment.Status.ReadyReplicas = 0
			fakeClient := NewFakeClient().WithObjects(notReadyDeployment).Build()
			handler := NewDeploymentHandler(fakeClient)

			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			readyReplicas, err := handler.GetWorkLoadReadyReplicas(ctx, indexer)
			convey.So(err, convey.ShouldBeNil)
			convey.So(readyReplicas, convey.ShouldEqual, 0)
		})

		convey.Convey("Should return error when listing Deployments fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			mockErr := errors.New("failed to list deployments")
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

// TestDeploymentHandlerListWorkLoads tests the ListWorkLoads method of DeploymentHandler.
func TestDeploymentHandlerListWorkLoads(t *testing.T) {
	convey.Convey("Test DeploymentHandler ListWorkLoads method", t, func() {
		selectLabels := map[string]string{
			common.InferServiceNameLabelKey: "test-service",
			common.InstanceSetNameLabelKey:  "test-role",
		}
		convey.Convey("Should successfully list Deployments", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			fakeClient := NewFakeClient().WithObjects(deployment).Build()
			handler := NewDeploymentHandler(fakeClient)

			ctx := context.Background()
			result, err := handler.ListWorkLoads(ctx, selectLabels)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(len(result.Items), convey.ShouldEqual, 1)
		})

		convey.Convey("Should return empty list when no Deployments found", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			ctx := context.Background()
			result, err := handler.ListWorkLoads(ctx, selectLabels)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(len(result.Items), convey.ShouldEqual, 0)
		})

		convey.Convey("Should return error when listing fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			mockErr := errors.New("failed to list deployments")
			patches := gomonkey.ApplyMethodReturn(handler.client, "List", mockErr)
			defer patches.Reset()

			ctx := context.Background()
			result, err := handler.ListWorkLoads(ctx, selectLabels)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestDeploymentHandlerValidate tests the Validate method of DeploymentHandler.
func TestDeploymentHandlerValidate(t *testing.T) {
	convey.Convey("Test DeploymentHandler Validate method", t, func() {
		convey.Convey("Should successfully validate valid DeploymentSpec", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			deploymentSpec := getTestDeploymentSpec()
			specBytes, err := json.Marshal(deploymentSpec)
			convey.So(err, convey.ShouldBeNil)
			spec := runtime.RawExtension{Raw: specBytes}

			err = handler.Validate(spec)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error for invalid DeploymentSpec", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			spec := runtime.RawExtension{Raw: []byte("invalid json")}

			err := handler.Validate(spec)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error for empty DeploymentSpec", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			spec := runtime.RawExtension{Raw: []byte{}}

			err := handler.Validate(spec)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestDeploymentHandlerGetReplicas tests the GetReplicas method of DeploymentHandler.
func TestDeploymentHandlerGetReplicas(t *testing.T) {
	convey.Convey("Test DeploymentHandler GetReplicas method", t, func() {
		convey.Convey("Should return correct replicas count", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			replicas := int32(3)
			deploymentSpec := &appsv1.DeploymentSpec{
				Replicas: &replicas,
			}
			specBytes, err := json.Marshal(deploymentSpec)
			convey.So(err, convey.ShouldBeNil)
			spec := runtime.RawExtension{Raw: specBytes}

			result, err := handler.GetReplicas(spec)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldEqual, int32(3))
		})

		convey.Convey("Should return default replicas when replicas is nil", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			deploymentSpec := &appsv1.DeploymentSpec{
				Replicas: nil,
			}
			specBytes, err := json.Marshal(deploymentSpec)
			convey.So(err, convey.ShouldBeNil)
			spec := runtime.RawExtension{Raw: specBytes}

			result, err := handler.GetReplicas(spec)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldEqual, common.DefaultReplicas)
		})

		convey.Convey("Should return error for invalid DeploymentSpec", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			spec := runtime.RawExtension{Raw: []byte("invalid json")}

			result, err := handler.GetReplicas(spec)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldEqual, common.DefaultReplicas)
		})
	})
}

// TestIsDeploymentReady tests the isDeploymentReady method of DeploymentHandler.
func TestIsDeploymentReady(t *testing.T) {
	convey.Convey("Test isDeploymentReady function", t, func() {
		convey.Convey("Should return true for ready Deployment", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)

			result := isDeploymentReady(*deployment)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("Should return false when replicas don't match", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Status.ReadyReplicas = 0

			result := isDeploymentReady(*deployment)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when generation is not latest", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Generation = 2
			deployment.Status.ObservedGeneration = 1

			result := isDeploymentReady(*deployment)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when available condition is missing", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Status.Conditions = []appsv1.DeploymentCondition{}

			result := isDeploymentReady(*deployment)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when available condition is false", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			for i := range deployment.Status.Conditions {
				if deployment.Status.Conditions[i].Type == appsv1.DeploymentAvailable {
					deployment.Status.Conditions[i].Status = corev1.ConditionFalse
				}
			}

			result := isDeploymentReady(*deployment)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

// TestGetDeploymentCondition tests the getDeploymentCondition
// method of DeploymentHandler.
func TestGetDeploymentCondition(t *testing.T) {
	convey.Convey("Test getDeploymentCondition function", t, func() {
		convey.Convey("Should return correct condition", func() {
			conditions := []appsv1.DeploymentCondition{
				{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue},
				{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue},
			}

			result := getDeploymentCondition(conditions, appsv1.DeploymentAvailable)
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(result.Type, convey.ShouldEqual, appsv1.DeploymentAvailable)
		})

		convey.Convey("Should return nil when condition not found", func() {
			conditions := []appsv1.DeploymentCondition{
				{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue},
			}

			result := getDeploymentCondition(conditions, appsv1.DeploymentProgressing)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("Should handle empty conditions list", func() {
			var conditions []appsv1.DeploymentCondition
			result := getDeploymentCondition(conditions, appsv1.DeploymentAvailable)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestParseDeploymentWithScheme tests the parseDeploymentWithScheme
// method of DeploymentHandler.
func TestParseDeploymentWithScheme(t *testing.T) {
	convey.Convey("Test DeploymentHandler parseDeploymentWithScheme method", t, func() {
		convey.Convey("Should successfully parse DeploymentSpec", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			deploymentSpec := getTestDeploymentSpec()
			specBytes, err := json.Marshal(deploymentSpec)
			convey.So(err, convey.ShouldBeNil)
			raw := runtime.RawExtension{Raw: specBytes}

			result, err := handler.parseDeploymentWithScheme(raw)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(*result.Replicas, convey.ShouldEqual, int32(1))
		})

		convey.Convey("Should return error for invalid json", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			raw := runtime.RawExtension{Raw: []byte("invalid json")}

			result, err := handler.parseDeploymentWithScheme(raw)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("Should return error for empty raw extension", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			raw := runtime.RawExtension{Raw: []byte{}}

			result, err := handler.parseDeploymentWithScheme(raw)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

func getTestDeploymentSpec() *appsv1.DeploymentSpec {
	replicas := int32(1)
	return &appsv1.DeploymentSpec{
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
	}
}
