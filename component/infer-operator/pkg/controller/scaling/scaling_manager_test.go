/*
Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package scaling

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"ascend-common/common-utils/hwlog"
	apiv1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func buildTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	apiv1.AddToScheme(scheme)
	autoscalingv2.AddToScheme(scheme)
	return scheme
}

func buildTestInstanceSet(name, namespace string, scalingPolicy *apiv1.ScalingPolicy) *apiv1.InstanceSet {
	is := &apiv1.InstanceSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "mindcluster.huawei.com/v1",
			Kind:       "InstanceSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: apiv1.InstanceSetSpec{
			ScalingPolicy: scalingPolicy,
		},
	}
	return is
}

func buildTestHPASpec(minReplicas, maxReplicas int32) *autoscalingv2.HorizontalPodAutoscalerSpec {
	min := minReplicas
	return &autoscalingv2.HorizontalPodAutoscalerSpec{
		MinReplicas: &min,
		MaxReplicas: maxReplicas,
		Metrics: []autoscalingv2.MetricSpec{
			{
				Type: autoscalingv2.ResourceMetricSourceType,
				Resource: &autoscalingv2.ResourceMetricSource{
					Name: "cpu",
					Target: autoscalingv2.MetricTarget{
						Type:               autoscalingv2.UtilizationMetricType,
						AverageUtilization: ptrTo[int32](80),
					},
				},
			},
		},
	}
}

func ptrTo[T any](v T) *T {
	return &v
}

func TestNewScalingManager(t *testing.T) {
	convey.Convey("TestNewScalingManager", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)
		convey.So(mgr, convey.ShouldNotBeNil)
		convey.So(mgr.Client, convey.ShouldEqual, fakeClient)
		convey.So(mgr.Scheme, convey.ShouldEqual, scheme)
	})
}

func TestReconcileScalingResourceNilPolicy(t *testing.T) {
	convey.Convey("TestReconcileScalingResource nil scaling policy", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		instanceSet := buildTestInstanceSet("test", "default", nil)
		status, err := mgr.ReconcileScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(status, convey.ShouldBeNil)
	})
}

func TestReconcileScalingResourceUnsupportedType(t *testing.T) {
	convey.Convey("TestReconcileScalingResource unsupported type", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		instanceSet := buildTestInstanceSet("test", "default", &apiv1.ScalingPolicy{
			Type: "Unsupported",
		})
		status, err := mgr.ReconcileScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(status, convey.ShouldNotBeNil)
		convey.So(status.Ready, convey.ShouldBeFalse)
		convey.So(status.Type, convey.ShouldEqual, "Unsupported")
		convey.So(status.Message, convey.ShouldContainSubstring, "unsupported")
	})
}

func TestReconcileHPAUnmarshalError(t *testing.T) {
	convey.Convey("TestReconcileHPA unmarshal error", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		instanceSet := buildTestInstanceSet("test", "default", &apiv1.ScalingPolicy{
			Type: common.ScalingPolicyTypeHPA,
			Spec: runtime.RawExtension{Raw: []byte("invalid-json")},
		})
		status, err := mgr.ReconcileScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(status, convey.ShouldNotBeNil)
		convey.So(status.Ready, convey.ShouldBeFalse)
		convey.So(status.Type, convey.ShouldEqual, common.ScalingPolicyTypeHPA)
		convey.So(status.Message, convey.ShouldContainSubstring, "failed to unmarshal")
	})
}

func TestReconcileHPAGetError(t *testing.T) {
	convey.Convey("TestReconcileHPA get error", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		hpaSpec := buildTestHPASpec(1, 10)
		raw, _ := json.Marshal(hpaSpec)
		instanceSet := buildTestInstanceSet("test", "default", &apiv1.ScalingPolicy{
			Type: common.ScalingPolicyTypeHPA,
			Spec: runtime.RawExtension{Raw: raw},
		})

		patch := gomonkey.ApplyMethodFunc(fakeClient, "Get",
			func(_ context.Context, _ types.NamespacedName, _ client.Object) error {
				return errors.New("get error")
			})
		defer patch.Reset()

		status, err := mgr.ReconcileScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(status, convey.ShouldBeNil)
	})
}

func TestReconcileHPACreateSuccess(t *testing.T) {
	convey.Convey("TestReconcileHPA create success", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		hpaSpec := buildTestHPASpec(1, 10)
		raw, _ := json.Marshal(hpaSpec)
		instanceSet := buildTestInstanceSet("test", "default", &apiv1.ScalingPolicy{
			Type: common.ScalingPolicyTypeHPA,
			Spec: runtime.RawExtension{Raw: raw},
		})

		status, err := mgr.ReconcileScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(status, convey.ShouldNotBeNil)
		convey.So(status.Ready, convey.ShouldBeTrue)
		convey.So(status.Type, convey.ShouldEqual, common.ScalingPolicyTypeHPA)
		convey.So(status.Name, convey.ShouldEqual, "test-scaler")
		convey.So(status.Message, convey.ShouldEqual, "HPA created successfully")

		createdHPA := &autoscalingv2.HorizontalPodAutoscaler{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{
			Name: "test-scaler", Namespace: "default",
		}, createdHPA)
		convey.So(err, convey.ShouldBeNil)
		convey.So(createdHPA.Spec.MaxReplicas, convey.ShouldEqual, 10)
		convey.So(createdHPA.Labels[ScalingResourceOwnedByInstanceSet], convey.ShouldEqual, "test")
	})
}

func TestReconcileHPACreateControllerRefError(t *testing.T) {
	convey.Convey("TestReconcileHPA create controller reference error", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		hpaSpec := buildTestHPASpec(1, 10)
		raw, _ := json.Marshal(hpaSpec)
		instanceSet := buildTestInstanceSet("test", "default", &apiv1.ScalingPolicy{
			Type: common.ScalingPolicyTypeHPA,
			Spec: runtime.RawExtension{Raw: raw},
		})

		patch := gomonkey.ApplyFuncReturn(controllerutil.SetControllerReference, errors.New("reference error"))
		defer patch.Reset()

		status, err := mgr.ReconcileScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "reference error")
		convey.So(status, convey.ShouldBeNil)
	})
}

func TestReconcileHPACreateError(t *testing.T) {
	convey.Convey("TestReconcileHPA create error", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		hpaSpec := buildTestHPASpec(1, 10)
		raw, _ := json.Marshal(hpaSpec)
		instanceSet := buildTestInstanceSet("test", "default", &apiv1.ScalingPolicy{
			Type: common.ScalingPolicyTypeHPA,
			Spec: runtime.RawExtension{Raw: raw},
		})

		patch := gomonkey.ApplyMethodFunc(fakeClient, "Create",
			func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
				return errors.New("create error")
			})
		defer patch.Reset()

		status, err := mgr.ReconcileScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(status, convey.ShouldNotBeNil)
		convey.So(status.Ready, convey.ShouldBeFalse)
		convey.So(status.Message, convey.ShouldContainSubstring, "failed to create HPA")
	})
}

func TestReconcileHPAUpdateNoChange(t *testing.T) {
	convey.Convey("TestReconcileHPA update no change", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		hpaSpec := buildTestHPASpec(1, 10)
		raw, _ := json.Marshal(hpaSpec)
		instanceSet := buildTestInstanceSet("test", "default", &apiv1.ScalingPolicy{
			Type: common.ScalingPolicyTypeHPA,
			Spec: runtime.RawExtension{Raw: raw},
		})

		existingHPA := &autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-scaler",
				Namespace: "default",
				Labels:    buildScalingResourceLabels(instanceSet),
			},
			Spec: *hpaSpec,
		}
		err := fakeClient.Create(context.Background(), existingHPA)
		convey.So(err, convey.ShouldBeNil)

		status, err := mgr.ReconcileScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(status, convey.ShouldNotBeNil)
		convey.So(status.Ready, convey.ShouldBeTrue)
		convey.So(status.Message, convey.ShouldEqual, "HPA updated successfully")
	})
}

func TestReconcileHPAUpdateWithChange(t *testing.T) {
	convey.Convey("TestReconcileHPA update with change", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		oldSpec := buildTestHPASpec(1, 5)
		existingHPA := &autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-scaler",
				Namespace: "default",
			},
			Spec: *oldSpec,
		}
		err := fakeClient.Create(context.Background(), existingHPA)
		convey.So(err, convey.ShouldBeNil)

		newSpec := buildTestHPASpec(2, 10)
		raw, _ := json.Marshal(newSpec)
		instanceSet := buildTestInstanceSet("test", "default", &apiv1.ScalingPolicy{
			Type: common.ScalingPolicyTypeHPA,
			Spec: runtime.RawExtension{Raw: raw},
		})

		status, err := mgr.ReconcileScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(status, convey.ShouldNotBeNil)
		convey.So(status.Ready, convey.ShouldBeTrue)
		convey.So(status.Message, convey.ShouldEqual, "HPA updated successfully")
	})
}

func TestReconcileHPAUpdateError(t *testing.T) {
	convey.Convey("TestReconcileHPA update error", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		oldSpec := buildTestHPASpec(1, 5)
		existingHPA := &autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-scaler",
				Namespace: "default",
			},
			Spec: *oldSpec,
		}
		err := fakeClient.Create(context.Background(), existingHPA)
		convey.So(err, convey.ShouldBeNil)

		newSpec := buildTestHPASpec(2, 10)
		raw, _ := json.Marshal(newSpec)
		instanceSet := buildTestInstanceSet("test", "default", &apiv1.ScalingPolicy{
			Type: common.ScalingPolicyTypeHPA,
			Spec: runtime.RawExtension{Raw: raw},
		})

		patch := gomonkey.ApplyMethodFunc(fakeClient, "Update",
			func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
				return errors.New("update error")
			})
		defer patch.Reset()

		status, err := mgr.ReconcileScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(status, convey.ShouldNotBeNil)
		convey.So(status.Ready, convey.ShouldBeFalse)
		convey.So(status.Message, convey.ShouldContainSubstring, "failed to update HPA")
	})
}

func TestCleanupScalingResourceListError(t *testing.T) {
	convey.Convey("TestCleanupScalingResource list error", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		instanceSet := buildTestInstanceSet("test", "default", nil)

		patch := gomonkey.ApplyMethodFunc(fakeClient, "List",
			func(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
				return errors.New("list error")
			})
		defer patch.Reset()

		status, err := mgr.cleanupScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(status, convey.ShouldBeNil)
	})
}

func TestCleanupScalingResourceDeleteError(t *testing.T) {
	convey.Convey("TestCleanupScalingResource delete error", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		instanceSet := buildTestInstanceSet("test", "default", nil)

		hpa := &autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-scaler",
				Namespace: "default",
				Labels: map[string]string{
					ScalingResourceOwnedByInstanceSet: "test",
				},
			},
		}
		err := fakeClient.Create(context.Background(), hpa)
		convey.So(err, convey.ShouldBeNil)

		patch := gomonkey.ApplyMethodFunc(fakeClient, "Delete",
			func(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
				return errors.New("delete error")
			})
		defer patch.Reset()

		status, err := mgr.cleanupScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(status, convey.ShouldBeNil)
	})
}

func TestCleanupScalingResourceSuccess(t *testing.T) {
	convey.Convey("TestCleanupScalingResource success", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		instanceSet := buildTestInstanceSet("test", "default", nil)

		hpa := &autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-scaler",
				Namespace: "default",
				Labels: map[string]string{
					ScalingResourceOwnedByInstanceSet: "test",
				},
			},
		}
		err := fakeClient.Create(context.Background(), hpa)
		convey.So(err, convey.ShouldBeNil)

		status, err := mgr.cleanupScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(status, convey.ShouldBeNil)

		existingHPA := &autoscalingv2.HorizontalPodAutoscaler{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{
			Name: "test-scaler", Namespace: "default",
		}, existingHPA)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestCleanupScalingResourceNoHPA(t *testing.T) {
	convey.Convey("TestCleanupScalingResource no HPA to clean", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		instanceSet := buildTestInstanceSet("test", "default", nil)

		status, err := mgr.cleanupScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(status, convey.ShouldBeNil)
	})
}

func TestBuildScalingResourceName(t *testing.T) {
	convey.Convey("TestBuildScalingResourceName", t, func() {
		instanceSet := buildTestInstanceSet("myapp", "default", nil)
		name := buildScalingResourceName(instanceSet)
		convey.So(name, convey.ShouldEqual, "myapp-scaler")
	})
}

func TestBuildScalingResourceLabels(t *testing.T) {
	convey.Convey("TestBuildScalingResourceLabels", t, func() {
		convey.Convey("with instance labels", func() {
			instanceSet := buildTestInstanceSet("test", "default", nil)
			instanceSet.Labels = map[string]string{
				"app": "myapp",
			}
			labels := buildScalingResourceLabels(instanceSet)
			convey.So(labels[ScalingResourceOwnedByInstanceSet], convey.ShouldEqual, "test")
			convey.So(labels[common.OperatorNameKey], convey.ShouldEqual, "")
			convey.So(labels["app"], convey.ShouldEqual, "myapp")
		})

		convey.Convey("without instance labels", func() {
			instanceSet := buildTestInstanceSet("test", "default", nil)
			labels := buildScalingResourceLabels(instanceSet)
			convey.So(labels[ScalingResourceOwnedByInstanceSet], convey.ShouldEqual, "test")
			convey.So(labels[common.OperatorNameKey], convey.ShouldEqual, "")
		})
	})
}

func TestUpdateHPAMinReplicasChanged(t *testing.T) {
	convey.Convey("TestUpdateHPA minReplicas changed", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		oldMin := int32(1)
		existingHPA := &autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-scaler",
				Namespace: "default",
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				MinReplicas: &oldMin,
				MaxReplicas: 10,
			},
		}
		err := fakeClient.Create(context.Background(), existingHPA)
		convey.So(err, convey.ShouldBeNil)

		newMin := int32(2)
		desiredSpec := &autoscalingv2.HorizontalPodAutoscalerSpec{
			MinReplicas: &newMin,
			MaxReplicas: 10,
		}

		instanceSet := buildTestInstanceSet("test", "default", nil)
		status, err := mgr.updateHPA(context.Background(), instanceSet, existingHPA, desiredSpec)
		convey.So(err, convey.ShouldBeNil)
		convey.So(status.Ready, convey.ShouldBeTrue)
		convey.So(status.Message, convey.ShouldEqual, "HPA updated successfully")
	})
}

func TestInjectMetricSelectorLabels(t *testing.T) {
	convey.Convey("TestInjectMetricSelectorLabels", t, func() {
		convey.Convey("no auto labels", func() {
			hpaSpec := &autoscalingv2.HorizontalPodAutoscalerSpec{
				Metrics: []autoscalingv2.MetricSpec{
					{
						Type: autoscalingv2.ExternalMetricSourceType,
						External: &autoscalingv2.ExternalMetricSource{
							Metric: autoscalingv2.MetricIdentifier{
								Name: "qps",
							},
						},
					},
				},
			}
			instanceSet := buildTestInstanceSet("test", "default", nil)
			injectMetricSelectorLabels(hpaSpec, instanceSet)
			convey.So(hpaSpec.Metrics[0].External.Metric.Selector, convey.ShouldBeNil)
		})

		convey.Convey("inject labels into external metric", func() {
			hpaSpec := &autoscalingv2.HorizontalPodAutoscalerSpec{
				Metrics: []autoscalingv2.MetricSpec{
					{
						Type: autoscalingv2.ExternalMetricSourceType,
						External: &autoscalingv2.ExternalMetricSource{
							Metric: autoscalingv2.MetricIdentifier{
								Name: "qps",
							},
						},
					},
				},
			}
			instanceSet := buildTestInstanceSet("test", "default", nil)
			instanceSet.Labels = map[string]string{
				common.InferServiceNameLabelKey: "myservice-0",
				common.InstanceSetNameLabelKey:  "role1",
			}
			injectMetricSelectorLabels(hpaSpec, instanceSet)
			convey.So(hpaSpec.Metrics[0].External.Metric.Selector, convey.ShouldNotBeNil)
			convey.So(hpaSpec.Metrics[0].External.Metric.Selector.MatchLabels[common.InferServiceNameLabelKey], convey.ShouldEqual, "myservice-0")
			convey.So(hpaSpec.Metrics[0].External.Metric.Selector.MatchLabels[common.RoleNameLabelKey], convey.ShouldEqual, "role1")
			convey.So(hpaSpec.Metrics[0].External.Metric.Selector.MatchLabels[common.InferServiceSetNameLabelKey], convey.ShouldEqual, "myservice")
		})

		convey.Convey("skip non-external metric", func() {
			hpaSpec := &autoscalingv2.HorizontalPodAutoscalerSpec{
				Metrics: []autoscalingv2.MetricSpec{
					{
						Type: autoscalingv2.ResourceMetricSourceType,
						Resource: &autoscalingv2.ResourceMetricSource{
							Name: "cpu",
						},
					},
				},
			}
			instanceSet := buildTestInstanceSet("test", "default", nil)
			instanceSet.Labels = map[string]string{
				common.InferServiceNameLabelKey: "myservice-0",
			}
			injectMetricSelectorLabels(hpaSpec, instanceSet)
			convey.So(hpaSpec.Metrics[0].Resource, convey.ShouldNotBeNil)
		})

		convey.Convey("external metric with nil external", func() {
			hpaSpec := &autoscalingv2.HorizontalPodAutoscalerSpec{
				Metrics: []autoscalingv2.MetricSpec{
					{
						Type:     autoscalingv2.ExternalMetricSourceType,
						External: nil,
					},
				},
			}
			instanceSet := buildTestInstanceSet("test", "default", nil)
			instanceSet.Labels = map[string]string{
				common.InferServiceNameLabelKey: "myservice-0",
			}
			injectMetricSelectorLabels(hpaSpec, instanceSet)
			convey.So(hpaSpec.Metrics[0].External, convey.ShouldBeNil)
		})

		convey.Convey("preserve existing selector labels", func() {
			hpaSpec := &autoscalingv2.HorizontalPodAutoscalerSpec{
				Metrics: []autoscalingv2.MetricSpec{
					{
						Type: autoscalingv2.ExternalMetricSourceType,
						External: &autoscalingv2.ExternalMetricSource{
							Metric: autoscalingv2.MetricIdentifier{
								Name: "qps",
								Selector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										common.InferServiceNameLabelKey: "existing-value",
									},
								},
							},
						},
					},
				},
			}
			instanceSet := buildTestInstanceSet("test", "default", nil)
			instanceSet.Labels = map[string]string{
				common.InferServiceNameLabelKey: "new-value",
				common.InstanceSetNameLabelKey:  "role1",
			}
			injectMetricSelectorLabels(hpaSpec, instanceSet)
			convey.So(hpaSpec.Metrics[0].External.Metric.Selector.MatchLabels[common.InferServiceNameLabelKey], convey.ShouldEqual, "existing-value")
			convey.So(hpaSpec.Metrics[0].External.Metric.Selector.MatchLabels[common.RoleNameLabelKey], convey.ShouldEqual, "role1")
		})
	})
}

func TestBuildAutoInjectedLabels(t *testing.T) {
	convey.Convey("TestBuildAutoInjectedLabels", t, func() {
		convey.Convey("no relevant labels", func() {
			instanceSet := buildTestInstanceSet("test", "default", nil)
			labels := buildAutoInjectedLabels(instanceSet)
			convey.So(len(labels), convey.ShouldEqual, 0)
		})

		convey.Convey("with InferServiceNameLabelKey", func() {
			instanceSet := buildTestInstanceSet("test", "default", nil)
			instanceSet.Labels = map[string]string{
				common.InferServiceNameLabelKey: "myservice-0",
			}
			labels := buildAutoInjectedLabels(instanceSet)
			convey.So(labels[common.InferServiceNameLabelKey], convey.ShouldEqual, "myservice-0")
			convey.So(labels[common.InferServiceSetNameLabelKey], convey.ShouldEqual, "myservice")
		})

		convey.Convey("with InferServiceNameLabelKey insufficient split", func() {
			instanceSet := buildTestInstanceSet("test", "default", nil)
			instanceSet.Labels = map[string]string{
				common.InferServiceNameLabelKey: "single",
			}
			labels := buildAutoInjectedLabels(instanceSet)
			convey.So(labels[common.InferServiceNameLabelKey], convey.ShouldEqual, "single")
			_, exists := labels[common.InferServiceSetNameLabelKey]
			convey.So(exists, convey.ShouldBeFalse)
		})

		convey.Convey("with InstanceSetNameLabelKey", func() {
			instanceSet := buildTestInstanceSet("test", "default", nil)
			instanceSet.Labels = map[string]string{
				common.InstanceSetNameLabelKey: "role1",
			}
			labels := buildAutoInjectedLabels(instanceSet)
			convey.So(labels[common.RoleNameLabelKey], convey.ShouldEqual, "role1")
		})

		convey.Convey("with both labels", func() {
			instanceSet := buildTestInstanceSet("test", "default", nil)
			instanceSet.Labels = map[string]string{
				common.InferServiceNameLabelKey: "myservice-0",
				common.InstanceSetNameLabelKey:  "role1",
			}
			labels := buildAutoInjectedLabels(instanceSet)
			convey.So(labels[common.InferServiceNameLabelKey], convey.ShouldEqual, "myservice-0")
			convey.So(labels[common.InferServiceSetNameLabelKey], convey.ShouldEqual, "myservice")
			convey.So(labels[common.RoleNameLabelKey], convey.ShouldEqual, "role1")
		})
	})
}

func TestCreateHPAWithAnnotations(t *testing.T) {
	convey.Convey("TestCreateHPA with annotations", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		instanceSet := buildTestInstanceSet("test", "default", nil)
		instanceSet.Annotations = map[string]string{
			"annotation-key": "annotation-value",
		}

		hpaSpec := buildTestHPASpec(1, 5)
		status, err := mgr.createHPA(context.Background(), instanceSet, "test-scaler", hpaSpec)
		convey.So(err, convey.ShouldBeNil)
		convey.So(status.Ready, convey.ShouldBeTrue)

		createdHPA := &autoscalingv2.HorizontalPodAutoscaler{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{
			Name: "test-scaler", Namespace: "default",
		}, createdHPA)
		convey.So(err, convey.ShouldBeNil)
		convey.So(createdHPA.Annotations["annotation-key"], convey.ShouldEqual, "annotation-value")
	})
}

func TestReconcileHPAWithExternalMetricAndLabels(t *testing.T) {
	convey.Convey("TestReconcileHPA with external metric and labels injection", t, func() {
		scheme := buildTestScheme()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mgr := NewScalingManager(fakeClient, scheme)

		min := int32(1)
		hpaSpec := &autoscalingv2.HorizontalPodAutoscalerSpec{
			MinReplicas: &min,
			MaxReplicas: 10,
			Metrics: []autoscalingv2.MetricSpec{
				{
					Type: autoscalingv2.ExternalMetricSourceType,
					External: &autoscalingv2.ExternalMetricSource{
						Metric: autoscalingv2.MetricIdentifier{
							Name: "qps",
						},
					},
				},
			},
		}
		raw, _ := json.Marshal(hpaSpec)

		instanceSet := buildTestInstanceSet("test", "default", &apiv1.ScalingPolicy{
			Type: common.ScalingPolicyTypeHPA,
			Spec: runtime.RawExtension{Raw: raw},
		})
		instanceSet.Labels = map[string]string{
			common.InferServiceNameLabelKey: "myservice-0",
			common.InstanceSetNameLabelKey:  "role1",
		}

		status, err := mgr.ReconcileScalingResource(context.Background(), instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(status.Ready, convey.ShouldBeTrue)

		createdHPA := &autoscalingv2.HorizontalPodAutoscaler{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{
			Name: "test-scaler", Namespace: "default",
		}, createdHPA)
		convey.So(err, convey.ShouldBeNil)
		convey.So(createdHPA.Spec.Metrics[0].External.Metric.Selector, convey.ShouldNotBeNil)
		convey.So(createdHPA.Spec.Metrics[0].External.Metric.Selector.MatchLabels[common.InferServiceNameLabelKey], convey.ShouldEqual, "myservice-0")
		convey.So(createdHPA.Spec.Metrics[0].External.Metric.Selector.MatchLabels[common.RoleNameLabelKey], convey.ShouldEqual, "role1")
	})
}
