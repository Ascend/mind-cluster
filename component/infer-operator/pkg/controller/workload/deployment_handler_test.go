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
	"slices"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

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

		convey.Convey("Should add inferserviceid to pod template labels", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			instanceSet.Labels[common.InferServiceIDLabelKey] = "test-uid"
			indexer := GetTestIndexer("test-service", "test-role", "0")
			deploymentSpec := getTestDeploymentSpec()
			specBytes, err := json.Marshal(deploymentSpec)
			convey.So(err, convey.ShouldBeNil)
			instanceSet.Spec.InstanceSpec = runtime.RawExtension{Raw: specBytes}
			ctx := context.Background()
			err = handler.createDeployment(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
			created := &appsv1.Deployment{}
			err = fakeClient.Get(ctx, types.NamespacedName{
				Name: common.GetWorkLoadNameFromIndexer(indexer), Namespace: "default"}, created)
			convey.So(err, convey.ShouldBeNil)
			convey.So(created.Spec.Template.Labels[common.InferServiceIDLabelKey], convey.ShouldEqual, "test-uid")
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
	const ns = "default"
	convey.Convey("Test DeploymentHandler DeleteExtraWorkLoad method", t, func() {
		convey.Convey("Should delete illegal index first, then unready, then ready by descending index, and keep invalid", func() {
			ready0 := CreateTestDeployment("test-service-test-role-0", ns, 1)
			ready0.Labels[common.InstanceIndexLabelKey] = "0"
			unready1 := CreateTestDeployment("test-service-test-role-1", ns, 1)
			unready1.Labels[common.InstanceIndexLabelKey] = "1"
			unready1.Status.ReadyReplicas = 0
			ready2 := CreateTestDeployment("test-service-test-role-2", ns, 1)
			ready2.Labels[common.InstanceIndexLabelKey] = "2"
			ready3 := CreateTestDeployment("test-service-test-role-3", ns, 1)
			ready3.Labels[common.InstanceIndexLabelKey] = "3"
			purgeNeg := CreateTestDeployment("test-service-test-role-neg", ns, 1)
			purgeNeg.Labels[common.InstanceIndexLabelKey] = "-1"
			noIndex := CreateTestDeployment("test-service-test-role-noidx", ns, 1)
			delete(noIndex.Labels, common.InstanceIndexLabelKey)
			invalidIndex := CreateTestDeployment("test-service-test-role-bad", ns, 1)
			invalidIndex.Labels[common.InstanceIndexLabelKey] = "abc"

			svc0 := CreateTestServiceWithIndex("test-service-test-role-0", ns, "0")
			svc1 := CreateTestServiceWithIndex("test-service-test-role-1", ns, "1")
			svc2 := CreateTestServiceWithIndex("test-service-test-role-2", ns, "2")
			svc3 := CreateTestServiceWithIndex("test-service-test-role-3", ns, "3")
			svcNeg := CreateTestServiceWithIndex("test-service-test-role-neg", ns, "-1")
			svcNoIdx := CreateTestServiceWithoutIndex("test-service-test-role-noidx", ns)

			fc := NewFakeClient(ready0, unready1, ready2, ready3, purgeNeg, noIndex, invalidIndex,
				svc0, svc1, svc2, svc3, svcNeg, svcNoIdx).Build()
			handler := NewDeploymentHandler(fc)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()

			existedDeploy := func(name string) bool {
				return fc.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, &appsv1.Deployment{}) == nil
			}
			existedSvc := func(name string) bool {
				return fc.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, &corev1.Service{}) == nil
			}

			// 7 listed workloads with indexLimit 5 => shrinkCount 2:
			// delete index -1 (highest priority), then the unready index 1; the quota is then exhausted.
			err := handler.DeleteExtraWorkLoad(ctx, indexer, 5)
			convey.So(err, convey.ShouldBeNil)

			// deleted indices in order: -1 (illegal), 1 (unready)
			convey.So(existedDeploy("test-service-test-role-1"), convey.ShouldBeFalse)
			convey.So(existedDeploy("test-service-test-role-neg"), convey.ShouldBeFalse)
			// kept indices: 0, 2, 3 (ready ones are not reached) and workloads with missing / invalid label
			convey.So(existedDeploy("test-service-test-role-0"), convey.ShouldBeTrue)
			convey.So(existedDeploy("test-service-test-role-2"), convey.ShouldBeTrue)
			convey.So(existedDeploy("test-service-test-role-3"), convey.ShouldBeTrue)
			convey.So(existedDeploy("test-service-test-role-noidx"), convey.ShouldBeTrue)
			convey.So(existedDeploy("test-service-test-role-bad"), convey.ShouldBeTrue)

			// services of the deleted index set are removed together with their workloads
			convey.So(existedSvc("test-service-test-role-1"), convey.ShouldBeFalse)
			convey.So(existedSvc("test-service-test-role-neg"), convey.ShouldBeFalse)
			convey.So(existedSvc("test-service-test-role-0"), convey.ShouldBeTrue)
			convey.So(existedSvc("test-service-test-role-2"), convey.ShouldBeTrue)
			convey.So(existedSvc("test-service-test-role-3"), convey.ShouldBeTrue)
			convey.So(existedSvc("test-service-test-role-noidx"), convey.ShouldBeTrue)
		})

		convey.Convey("Should return error when listing Deployments fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			mockErr := errors.New("failed to list deployments")
			patches := gomonkey.ApplyMethodReturn(handler.client, "List", mockErr)
			defer patches.Reset()
			ctx := context.Background()
			indexer := GetTestIndexer("test-service", "test-role", "0")
			err := handler.DeleteExtraWorkLoad(ctx, indexer, 5)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when deleting Deployment fails", func() {
			readyDeployment := CreateTestDeployment("test-service-test-role-0", ns, 1)
			readyDeployment.Labels[common.InstanceIndexLabelKey] = "0"
			fakeClient := NewFakeClient().WithObjects(readyDeployment).Build()
			handler := NewDeploymentHandler(fakeClient)
			mockErr := errors.New("failed to delete deployment")
			patches2 := gomonkey.ApplyMethodReturn(handler.client, "Delete", mockErr)
			defer patches2.Reset()
			ctx := context.Background()
			indexer := GetTestIndexer("test-service", "test-role", "0")
			// indexLimit 0 with one workload => shrinkCount 1, index 0 is selected and its deletion fails
			err := handler.DeleteExtraWorkLoad(ctx, indexer, 0)
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
			result, err := handler.ListWorkLoads(ctx, selectLabels, "default")
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(len(result.Items), convey.ShouldEqual, 1)
		})

		convey.Convey("Should return empty list when no Deployments found", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			ctx := context.Background()
			result, err := handler.ListWorkLoads(ctx, selectLabels, "default")
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
			result, err := handler.ListWorkLoads(ctx, selectLabels, "default")
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

func createTestWorkLoadFilter() WorkLoadFilter {
	return func(workload WorkLoadInterface) bool {
		return workload.GetWorkLoadReplicas() > 0
	}
}

func createTestWorkloadUpdater() WorkloadUpdater {
	return func(workload WorkLoadInterface) {
		workload.SetWorkLoadObjMeta(metav1.ObjectMeta{
			Labels: map[string]string{"updated": "true"},
		})
	}
}

func TestDeploymentHandlerListWorkLoad(t *testing.T) {
	convey.Convey("Test DeploymentHandler ListWorkLoad method", t, func() {
		selectLabels := map[string]string{
			common.InferServiceNameLabelKey: "test-service",
			common.InstanceSetNameLabelKey:  "test-role",
		}

		convey.Convey("Should list workloads successfully without filters", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Spec.Replicas = func() *int32 { r := int32(2); return &r }()
			fakeClient := NewFakeClient().WithObjects(deployment).Build()
			handler := NewDeploymentHandler(fakeClient)

			result, err := handler.ListWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 1)
			convey.So(result[0].GetWorkLoadReplicas(), convey.ShouldEqual, int32(2))
		})

		convey.Convey("Should filter workloads with filter function", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Spec.Replicas = func() *int32 { r := int32(0); return &r }()
			fakeClient := NewFakeClient().WithObjects(deployment).Build()
			handler := NewDeploymentHandler(fakeClient)
			filter := createTestWorkLoadFilter()

			result, err := handler.ListWorkLoad(context.Background(), selectLabels, "default", filter)

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("Should return empty list when no deployments match", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)

			result, err := handler.ListWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("Should return error when selector creation fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			invalidLabels := map[string]string{"invalid/label": "value"}
			patch := gomonkey.ApplyFuncReturn(metav1.LabelSelectorAsSelector, nil, fmt.Errorf("create selector failed"))
			defer patch.Reset()
			result, err := handler.ListWorkLoad(context.Background(), invalidLabels, "default")

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when client List fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			mockErr := errors.New("failed to list deployments")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "List", mockErr)
			defer patches.Reset()

			result, err := handler.ListWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

func TestDeploymentHandlerDeleteWorkLoad(t *testing.T) {
	convey.Convey("Test DeploymentHandler DeleteWorkLoad method", t, func() {
		selectLabels := map[string]string{
			common.InferServiceNameLabelKey: "test-service",
			common.InstanceSetNameLabelKey:  "test-role",
		}

		convey.Convey("Should delete workloads successfully without filters", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			fakeClient := NewFakeClient().WithObjects(deployment).Build()
			handler := NewDeploymentHandler(fakeClient)
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Delete", nil)
			defer patches.Reset()

			err := handler.DeleteWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should skip workloads that don't match filter", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 0)
			fakeClient := NewFakeClient().WithObjects(deployment).Build()
			handler := NewDeploymentHandler(fakeClient)
			filter := createTestWorkLoadFilter()

			err := handler.DeleteWorkLoad(context.Background(), selectLabels, "default", filter)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when ListWorkLoads fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			mockErr := errors.New("failed to list deployments")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "List", mockErr)
			defer patches.Reset()

			err := handler.DeleteWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when client Delete fails", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			fakeClient := NewFakeClient().WithObjects(deployment).Build()
			handler := NewDeploymentHandler(fakeClient)
			mockErr := errors.New("failed to delete deployment")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Delete", mockErr)
			defer patches.Reset()

			err := handler.DeleteWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should delete pods when external-force mode with matching pods", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Labels[common.FaultSchedulingLabelKey] = common.ExternalForceReschedulingValue
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels: map[string]string{
						common.InferServiceNameLabelKey: "test-service",
						common.InstanceSetNameLabelKey:  "test-role",
						common.InstanceIndexLabelKey:    "0",
					},
				},
			}
			fakeClient := NewFakeClient().WithObjects(deployment, pod).Build()
			handler := NewDeploymentHandler(fakeClient)
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Delete", nil)
			defer patches.Reset()

			err := handler.DeleteWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should succeed when external-force mode but no pods exist", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Labels[common.FaultSchedulingLabelKey] = common.ExternalForceReschedulingValue
			fakeClient := NewFakeClient().WithObjects(deployment).Build()
			handler := NewDeploymentHandler(fakeClient)
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Delete", nil)
			defer patches.Reset()

			err := handler.DeleteWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should log error and continue when pod deletion fails in external-force mode", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Labels[common.FaultSchedulingLabelKey] = common.ExternalForceReschedulingValue
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels: map[string]string{
						common.InferServiceNameLabelKey: "test-service",
						common.InstanceSetNameLabelKey:  "test-role",
						common.InstanceIndexLabelKey:    "0",
					},
				},
			}
			fakeClient := NewFakeClient().WithObjects(deployment, pod).Build()
			handler := NewDeploymentHandler(fakeClient)
			mockErr := errors.New("failed to delete pod")
			patches := gomonkey.ApplyMethodSeq(fakeClient, "Delete", []gomonkey.OutputCell{
				{Values: gomonkey.Params{nil}, Times: 1},
				{Values: gomonkey.Params{mockErr}, Times: 1},
			})
			defer patches.Reset()

			err := handler.DeleteWorkLoad(context.Background(), selectLabels, "default")

			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDeploymentHandlerUpdateWorkLoad(t *testing.T) {
	convey.Convey("Test DeploymentHandler UpdateWorkLoad method", t, func() {
		selectLabels := map[string]string{
			common.InferServiceNameLabelKey: "test-service",
			common.InstanceSetNameLabelKey:  "test-role",
		}

		convey.Convey("Should update workloads successfully without filters", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			fakeClient := NewFakeClient().WithObjects(deployment).Build()
			handler := NewDeploymentHandler(fakeClient)
			updater := createTestWorkloadUpdater()
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Update", nil)
			defer patches.Reset()

			err := handler.UpdateWorkLoad(context.Background(), selectLabels, "default", updater)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should skip workloads that don't match filter", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 0)
			fakeClient := NewFakeClient().WithObjects(deployment).Build()
			handler := NewDeploymentHandler(fakeClient)
			filter := createTestWorkLoadFilter()
			updater := createTestWorkloadUpdater()

			err := handler.UpdateWorkLoad(context.Background(), selectLabels, "default", updater, filter)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when ListWorkLoads fails", func() {
			fakeClient := NewFakeClient().Build()
			handler := NewDeploymentHandler(fakeClient)
			mockErr := errors.New("failed to list deployments")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "List", mockErr)
			defer patches.Reset()
			updater := createTestWorkloadUpdater()

			err := handler.UpdateWorkLoad(context.Background(), selectLabels, "default", updater)

			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when client Update fails", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			fakeClient := NewFakeClient().WithObjects(deployment).Build()
			handler := NewDeploymentHandler(fakeClient)
			mockErr := errors.New("failed to update deployment")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Update", mockErr)
			defer patches.Reset()
			updater := createTestWorkloadUpdater()

			err := handler.UpdateWorkLoad(context.Background(), selectLabels, "default", updater)

			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func createTestDeploymentWorkLoad() *DeploymentWorkLoad {
	deployment := CreateTestDeployment("test-deployment", "default", 1)
	deployment.Status.ReadyReplicas = 1
	deployment.Status.AvailableReplicas = 1
	deployment.Status.UpdatedReplicas = 1
	deployment.Status.ObservedGeneration = 1
	deployment.Status.Conditions = []appsv1.DeploymentCondition{
		{
			Type:   appsv1.DeploymentAvailable,
			Status: corev1.ConditionTrue,
		},
		{
			Type:   appsv1.DeploymentProgressing,
			Status: corev1.ConditionTrue,
		},
	}
	return &DeploymentWorkLoad{Deployment: deployment}
}

func TestDeploymentWorkLoadSetWorkLoadObjMeta(t *testing.T) {
	convey.Convey("Test DeploymentWorkLoad SetWorkLoadObjMeta method", t, func() {
		convey.Convey("Should set object meta successfully", func() {
			workload := &DeploymentWorkLoad{
				Deployment: &appsv1.Deployment{},
			}
			objectMeta := metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
				Labels:    map[string]string{"key": "value"},
			}

			workload.SetWorkLoadObjMeta(objectMeta)

			convey.So(workload.Name, convey.ShouldEqual, "test-deployment")
			convey.So(workload.Namespace, convey.ShouldEqual, "default")
			convey.So(workload.Labels["key"], convey.ShouldEqual, "value")
		})

		convey.Convey("Should handle nil workload", func() {
			var workload *DeploymentWorkLoad
			objectMeta := metav1.ObjectMeta{Name: "test"}

			workload.SetWorkLoadObjMeta(objectMeta)

			convey.So(workload, convey.ShouldBeNil)
		})
	})
}

func TestDeploymentWorkLoadGetWorkLoadObjMeta(t *testing.T) {
	convey.Convey("Test DeploymentWorkLoad GetWorkLoadObjMeta method", t, func() {
		convey.Convey("Should get object meta successfully", func() {
			workload := &DeploymentWorkLoad{
				Deployment: &appsv1.Deployment{},
			}
			expectedMeta := metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			}
			workload.SetWorkLoadObjMeta(expectedMeta)

			result := workload.GetWorkLoadObjMeta()

			convey.So(result.Name, convey.ShouldEqual, "test-deployment")
			convey.So(result.Namespace, convey.ShouldEqual, "default")
		})

		convey.Convey("Should return empty meta for nil workload", func() {
			var workload *DeploymentWorkLoad

			result := workload.GetWorkLoadObjMeta()

			convey.So(result, convey.ShouldResemble, metav1.ObjectMeta{})
		})
	})
}

func TestDeploymentWorkLoadIsWorkloadReady(t *testing.T) {
	convey.Convey("Test DeploymentWorkLoad IsWorkLoadReady method", t, func() {
		convey.Convey("Should return true for ready workload", func() {
			workload := createTestDeploymentWorkLoad()

			result := workload.IsWorkLoadReady()

			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("Should return false when replicas don't match", func() {
			workload := createTestDeploymentWorkLoad()
			workload.Status.ReadyReplicas = 0

			result := workload.IsWorkLoadReady()

			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when generation is not latest", func() {
			workload := createTestDeploymentWorkLoad()
			workload.Generation = 2
			workload.Status.ObservedGeneration = 1

			result := workload.IsWorkLoadReady()

			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when available condition is missing", func() {
			workload := createTestDeploymentWorkLoad()
			workload.Status.Conditions = []appsv1.DeploymentCondition{}

			result := workload.IsWorkLoadReady()

			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when available condition is false", func() {
			workload := createTestDeploymentWorkLoad()
			for i := range workload.Status.Conditions {
				if workload.Status.Conditions[i].Type == appsv1.DeploymentAvailable {
					workload.Status.Conditions[i].Status = corev1.ConditionFalse
				}
			}

			result := workload.IsWorkLoadReady()

			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when progressing condition is false", func() {
			workload := createTestDeploymentWorkLoad()
			for i := range workload.Status.Conditions {
				if workload.Status.Conditions[i].Type == appsv1.DeploymentProgressing {
					workload.Status.Conditions[i].Status = corev1.ConditionFalse
				}
			}

			result := workload.IsWorkLoadReady()

			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false for nil workload", func() {
			var workload *DeploymentWorkLoad

			result := workload.IsWorkLoadReady()

			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestDeploymentWorkLoadGetWorkLoadReplicas(t *testing.T) {
	convey.Convey("Test DeploymentWorkLoad GetWorkLoadReplicas method", t, func() {
		convey.Convey("Should return correct replicas", func() {
			workload := createTestDeploymentWorkLoad()
			replicas := int32(5)
			workload.Spec.Replicas = &replicas

			result := workload.GetWorkLoadReplicas()

			convey.So(result, convey.ShouldEqual, int32(5))
		})

		convey.Convey("Should return default replicas when replicas is nil", func() {
			workload := createTestDeploymentWorkLoad()
			workload.Spec.Replicas = nil

			result := workload.GetWorkLoadReplicas()

			convey.So(result, convey.ShouldEqual, common.DefaultReplicas)
		})

		convey.Convey("Should return default replicas for nil workload", func() {
			var workload *DeploymentWorkLoad

			result := workload.GetWorkLoadReplicas()

			convey.So(result, convey.ShouldEqual, common.DefaultReplicas)
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

// TestDeploymentHandlerDeleteServicesNotInIndices tests the deleteServicesNotInIndices method of DeploymentHandler.
func TestDeploymentHandlerDeleteServicesNotInIndices(t *testing.T) {
	const (
		ns         = "default"
		svcName    = "test-service"
		instSetKey = "test-role"
		extraIdx   = "5"
		negIdx     = "-1"
		normalIdx  = "0"
		invalidIdx = "invalid"
	)
	selectLabels := map[string]string{
		common.InferServiceNameLabelKey: svcName,
		common.InstanceSetNameLabelKey:  instSetKey,
	}
	keepIndices := []int{0}
	testCases := []struct {
		name       string
		services   []*corev1.Service
		shouldKeep []string
		shouldGone []string
	}{
		{name: "should return nil when no services exist"},
		{name: "should delete service when index is not in the keep set",
			services:   []*corev1.Service{CreateTestServiceWithIndex("svc-extra", ns, extraIdx)},
			shouldGone: []string{"svc-extra"}},
		{name: "should delete service when index is negative and not in the keep set",
			services:   []*corev1.Service{CreateTestServiceWithIndex("svc-neg", ns, negIdx)},
			shouldGone: []string{"svc-neg"}},
		{name: "should keep service when index is in the keep set",
			services:   []*corev1.Service{CreateTestServiceWithIndex("svc-normal", ns, normalIdx)},
			shouldKeep: []string{"svc-normal"}},
		{name: "should skip service when index label is missing",
			services:   []*corev1.Service{CreateTestServiceWithoutIndex("svc-noidx", ns)},
			shouldKeep: []string{"svc-noidx"}},
		{name: "should skip service when index is invalid integer",
			services:   []*corev1.Service{CreateTestServiceWithIndex("svc-invalid", ns, invalidIdx)},
			shouldKeep: []string{"svc-invalid"}},
	}
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			objs := make([]runtime.Object, len(tc.services))
			for i, svc := range tc.services {
				objs[i] = svc
			}
			fc := NewFakeClient(objs...).Build()
			handler := NewDeploymentHandler(fc)
			err := handler.deleteServicesNotInIndices(context.Background(), selectLabels, ns, keepIndices)
			convey.So(err, convey.ShouldBeNil)

			remaining := &corev1.ServiceList{}
			getErr := fc.List(context.Background(), remaining)
			convey.So(getErr, convey.ShouldBeNil)
			remainNames := make(map[string]bool, len(remaining.Items))
			for i := range remaining.Items {
				remainNames[remaining.Items[i].Name] = true
			}
			for _, name := range tc.shouldKeep {
				convey.So(remainNames[name], convey.ShouldBeTrue)
			}
			for _, name := range tc.shouldGone {
				convey.So(remainNames[name], convey.ShouldBeFalse)
			}
		})
	}
}

// TestDeploymentHandlerDeleteServicesNotInIndicesError tests deleteServicesNotInIndices error cases.
func TestDeploymentHandlerDeleteServicesNotInIndicesError(t *testing.T) {
	const (
		ns         = "default"
		svcName    = "test-service"
		instSetKey = "test-role"
		extraIdx   = "5"
	)
	selectLabels := map[string]string{
		common.InferServiceNameLabelKey: svcName,
		common.InstanceSetNameLabelKey:  instSetKey,
	}
	keepIndices := []int{0}
	testCases := []struct {
		name      string
		services  []*corev1.Service
		listErr   error
		deleteErr error
	}{
		{name: "should return error when listing services fails",
			listErr: errors.New("list failed")},
		{name: "should return error when deleting service fails",
			services:  []*corev1.Service{CreateTestServiceWithIndex("svc-extra", ns, extraIdx)},
			deleteErr: errors.New("delete failed")},
	}
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			objs := make([]runtime.Object, len(tc.services))
			for i, svc := range tc.services {
				objs[i] = svc
			}
			fc := NewFakeClient(objs...).Build()
			handler := NewDeploymentHandler(fc)
			if tc.listErr != nil {
				patches := gomonkey.ApplyMethodReturn(fc, "List", tc.listErr)
				defer patches.Reset()
			}
			if tc.deleteErr != nil {
				patches := gomonkey.ApplyMethodReturn(fc, "Delete", tc.deleteErr)
				defer patches.Reset()
			}
			err := handler.deleteServicesNotInIndices(context.Background(), selectLabels, ns, keepIndices)
			convey.So(err, convey.ShouldNotBeNil)
		})
	}
}

// legacyDeletedIndices mirrors the pre-refactor deletion set: every instance whose index is
// negative or not below indexLimit is deleted. The result is returned in descending order,
// matching selectIndicesToDelete, so the two can be compared element by element.
func legacyDeletedIndices(indices []int, indexLimit int) []int {
	deleted := make([]int, 0, len(indices))
	for _, index := range indices {
		if index < 0 || index >= indexLimit {
			deleted = append(deleted, index)
		}
	}
	slices.SortFunc(deleted, func(a, b int) int { return b - a })
	return deleted
}

// indicesNotInSet returns the elements of all that are absent from a set, in descending order.
func indicesNotInSet(all []int, present map[int]struct{}) []int {
	missing := make([]int, 0, len(all))
	for _, index := range all {
		if _, ok := present[index]; !ok {
			missing = append(missing, index)
		}
	}
	slices.SortFunc(missing, func(a, b int) int { return b - a })
	return missing
}

// TestDFxC01AllReadyDeleteSetMatchesLegacy reproduces design §9.2.3 TC-DFx-C01: with a
// contiguous all-ready instance set the refactored selection must be bit-for-bit identical
// to the legacy rule (delete index<0 or index>=indexLimit), both on the pure function and
// end to end through DeleteExtraWorkLoad on Deployments.
func TestDFxC01AllReadyDeleteSetMatchesLegacy(t *testing.T) {
	const (
		ns         = "default"
		indexLimit = 5
	)
	allIndices := []int{0, 1, 2, 3, 4, 5, 6, 7}
	wantDeleted := legacyDeletedIndices(allIndices, indexLimit) // 8->5 => {7,6,5}

	// 1. pure function: contiguous all-ready workloads degrade to the highest indices.
	workloads := make([]WorkLoadInterface, 0, len(allIndices))
	for _, index := range allIndices {
		workloads = append(workloads, readyWorkLoad(fmt.Sprintf("%d", index)))
	}
	gotIndices := selectIndicesToDelete(workloads, indexLimit)
	if !slices.Equal(gotIndices, wantDeleted) {
		t.Fatalf("selectIndicesToDelete() = %v, legacy = %v", gotIndices, wantDeleted)
	}

	// 2. handler level: the deleted Deployment indices match the legacy expectation.
	objs := make([]runtime.Object, 0, len(allIndices)*2)
	for _, index := range allIndices {
		name := fmt.Sprintf("test-service-test-role-%d", index)
		deploy := CreateTestDeployment(name, ns, 1)
		deploy.Labels[common.InstanceIndexLabelKey] = fmt.Sprintf("%d", index)
		objs = append(objs, deploy, CreateTestServiceWithIndex(name, ns, fmt.Sprintf("%d", index)))
	}
	fc := NewFakeClient(objs...).Build()
	handler := NewDeploymentHandler(fc)
	indexer := GetTestIndexer("test-service", "test-role", "0")
	if err := handler.DeleteExtraWorkLoad(context.Background(), indexer, indexLimit); err != nil {
		t.Fatal(err)
	}

	deletedIndices := make([]int, 0, len(allIndices))
	for _, index := range allIndices {
		name := fmt.Sprintf("test-service-test-role-%d", index)
		if err := fc.Get(context.Background(),
			types.NamespacedName{Name: name, Namespace: ns}, &appsv1.Deployment{}); err != nil {
			deletedIndices = append(deletedIndices, index)
		}
	}
	slices.SortFunc(deletedIndices, func(a, b int) int { return b - a })
	if !slices.Equal(deletedIndices, wantDeleted) {
		t.Fatalf("deleted Deployment indices = %v, legacy = %v", deletedIndices, wantDeleted)
	}
	if len(deletedIndices) == 0 {
		t.Fatal("fixture deleted no Deployment, the assertion would be vacuous")
	}
}

// TestDFxC02ServiceAndWorkloadDeleteSetsAgree reproduces design §9.2.3 TC-DFx-C02: the
// Service cleanup must keep exactly the same instance-index set as the workload deletion.
func TestDFxC02ServiceAndWorkloadDeleteSetsAgree(t *testing.T) {
	const (
		ns         = "default"
		indexLimit = 5
	)
	allIndices := []int{0, 1, 2, 3, 4, 5, 6, 7}

	objs := make([]runtime.Object, 0, len(allIndices)*2)
	for _, index := range allIndices {
		label := fmt.Sprintf("%d", index)
		name := fmt.Sprintf("test-service-test-role-%d", index)
		deploy := CreateTestDeployment(name, ns, 1)
		deploy.Labels[common.InstanceIndexLabelKey] = label
		objs = append(objs, deploy, CreateTestServiceWithIndex(name, ns, label))
	}
	fc := NewFakeClient(objs...).Build()
	handler := NewDeploymentHandler(fc)
	indexer := GetTestIndexer("test-service", "test-role", "0")
	if err := handler.DeleteExtraWorkLoad(context.Background(), indexer, indexLimit); err != nil {
		t.Fatal(err)
	}

	remainingDeploy := make(map[int]struct{})
	deployList := &appsv1.DeploymentList{}
	if err := fc.List(context.Background(), deployList); err != nil {
		t.Fatal(err)
	}
	for i := range deployList.Items {
		if index, ok := parseInstanceIndex(deployList.Items[i].ObjectMeta); ok {
			remainingDeploy[index] = struct{}{}
		}
	}

	remainingService := make(map[int]struct{})
	serviceList := &corev1.ServiceList{}
	if err := fc.List(context.Background(), serviceList); err != nil {
		t.Fatal(err)
	}
	for i := range serviceList.Items {
		if index, ok := parseInstanceIndex(serviceList.Items[i].ObjectMeta); ok {
			remainingService[index] = struct{}{}
		}
	}

	deletedDeploy := indicesNotInSet(allIndices, remainingDeploy)
	deletedService := indicesNotInSet(allIndices, remainingService)
	if !slices.Equal(deletedDeploy, deletedService) {
		t.Fatalf("deleted workload indices = %v, deleted Service indices = %v", deletedDeploy, deletedService)
	}
	if len(deletedDeploy) != len(allIndices)-indexLimit {
		t.Fatalf("deleted %d instances, want %d", len(deletedDeploy), len(allIndices)-indexLimit)
	}
}

// TestDeploymentHandlerDeleteServicesCleansOrphanService covers the review finding: when the
// workload of a deleted instance is already gone, its leftover Service must still be cleaned up
// on a later round, and the Service of a surviving instance above indexLimit must be kept.
func TestDeploymentHandlerDeleteServicesCleansOrphanService(t *testing.T) {
	const (
		ns         = "default"
		indexLimit = 4
	)
	// The workloads now number exactly indexLimit, so this round selects nothing to delete: index 2
	// was deleted as unready in a previous round and only its Service is left behind. index 4
	// survives even though 4 >= indexLimit, so its Service must not be deleted by the boundary rule.
	remaining := []int{0, 1, 3, 4}
	objs := make([]runtime.Object, 0, len(remaining)*2+1)
	for _, index := range remaining {
		label := fmt.Sprintf("%d", index)
		name := fmt.Sprintf("test-service-test-role-%d", index)
		deploy := CreateTestDeployment(name, ns, 1)
		deploy.Labels[common.InstanceIndexLabelKey] = label
		objs = append(objs, deploy, CreateTestServiceWithIndex(name, ns, label))
	}
	// orphan Service: its workload was already deleted in a previous round.
	objs = append(objs, CreateTestServiceWithIndex("orphan-svc", ns, "2"))

	fc := NewFakeClient(objs...).Build()
	handler := NewDeploymentHandler(fc)
	indexer := GetTestIndexer("test-service", "test-role", "0")
	if err := handler.DeleteExtraWorkLoad(context.Background(), indexer, indexLimit); err != nil {
		t.Fatal(err)
	}

	if err := fc.Get(context.Background(),
		types.NamespacedName{Name: "orphan-svc", Namespace: ns}, &corev1.Service{}); err == nil {
		t.Fatal("orphan Service was not cleaned up")
	}
	for _, index := range remaining {
		name := fmt.Sprintf("test-service-test-role-%d", index)
		if err := fc.Get(context.Background(),
			types.NamespacedName{Name: name, Namespace: ns}, &corev1.Service{}); err != nil {
			t.Fatalf("Service of surviving instance %d was deleted: %v", index, err)
		}
	}
}
