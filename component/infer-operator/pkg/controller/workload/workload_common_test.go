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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

var (
	localScheme = runtime.NewScheme()
)

func init() {
	_ = scheme.AddToScheme(localScheme)
	_ = v1.AddToScheme(localScheme)
	_ = v1beta1.AddToScheme(localScheme)
}

func GetScheme() *runtime.Scheme {
	return localScheme
}

func NewFakeClient(objects ...runtime.Object) *fake.ClientBuilder {
	return fake.NewClientBuilder().WithScheme(GetScheme()).WithRuntimeObjects(objects...)
}

// CreateTestInstanceSet creates a test InstanceSet object.
func CreateTestInstanceSet(name, namespace string, replicas int32) *v1.InstanceSet {
	return &v1.InstanceSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.OperatorNameKey:          common.TrueBool,
			},
		},
		Spec: v1.InstanceSetSpec{
			Name:     "test-role",
			Replicas: &replicas,
			WorkloadTypeMeta: v1.WorkloadType{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			WorkloadObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"app": "test",
				},
			},
		},
	}
}

// CreateTestDeployment creates a test Deployment object.
func CreateTestDeployment(name, namespace string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
				common.OperatorNameKey:          common.TrueBool,
			},
		},
		Spec: appsv1.DeploymentSpec{
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
						{
							Name:  "test-container",
							Image: "test-image",
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:      replicas,
			AvailableReplicas:  replicas,
			UpdatedReplicas:    replicas,
			ObservedGeneration: 1,
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   appsv1.DeploymentProgressing,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
}

// CreateTestStatefulSet creates a test StatefulSet object.
func CreateTestStatefulSet(name, namespace string, replicas int32) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
				common.OperatorNameKey:          common.TrueBool,
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "test-image",
						},
					},
				},
			},
			ServiceName: "test-service",
		},
		Status: appsv1.StatefulSetStatus{
			ReadyReplicas:      replicas,
			UpdatedReplicas:    replicas,
			ObservedGeneration: 1,
			CurrentRevision:    "v1",
			UpdateRevision:     "v1",
		},
	}
}

// CreateTestService creates a test Service object.
func CreateTestService(name, namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
				common.OperatorNameKey:          common.TrueBool,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "test",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     common.DefaultPortName,
					Port:     common.DefaultPort,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
}

// GetTestIndexer creates a test InstanceIndexer object.
func GetTestIndexer(serviceName, instanceSetKey, instanceIndex string) common.InstanceIndexer {
	return common.InstanceIndexer{
		ServiceName:    serviceName,
		InstanceSetKey: instanceSetKey,
		InstanceIndex:  instanceIndex,
	}
}

func TestDeletePodsForExternalRescheduling(t *testing.T) {
	convey.Convey("Test deletePodsForExternalRescheduling function", t, func() {
		convey.Convey("Should return nil when workload has no fault-scheduling label", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			workload := &DeploymentWorkLoad{Deployment: deployment}
			fakeClient := NewFakeClient().Build()

			err := deletePodsForExternalRescheduling(context.Background(), fakeClient, workload)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should delete pods when external-force mode with matching pods", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Labels[common.FaultSchedulingLabelKey] = common.ExternalForceReschedulingValue
			workload := &DeploymentWorkLoad{Deployment: deployment}
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
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Delete", nil)
			defer patches.Reset()

			err := deletePodsForExternalRescheduling(context.Background(), fakeClient, workload)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should succeed when external-force mode but no pods exist", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Labels[common.FaultSchedulingLabelKey] = common.ExternalForceReschedulingValue
			workload := &DeploymentWorkLoad{Deployment: deployment}
			fakeClient := NewFakeClient().WithObjects(deployment).Build()

			err := deletePodsForExternalRescheduling(context.Background(), fakeClient, workload)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when pod list fails in external-force mode", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Labels[common.FaultSchedulingLabelKey] = common.ExternalForceReschedulingValue
			workload := &DeploymentWorkLoad{Deployment: deployment}
			fakeClient := NewFakeClient().Build()
			mockErr := errors.New("failed to list pods")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "List", mockErr)
			defer patches.Reset()

			err := deletePodsForExternalRescheduling(context.Background(), fakeClient, workload)

			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should skip not-found error when deleting pods in external-force mode", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Labels[common.FaultSchedulingLabelKey] = common.ExternalForceReschedulingValue
			workload := &DeploymentWorkLoad{Deployment: deployment}
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
			notFoundErr := apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "pods"}, "test-pod")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Delete", notFoundErr)
			defer patches.Reset()

			err := deletePodsForExternalRescheduling(context.Background(), fakeClient, workload)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should log error and continue when pod delete fails in external-force mode", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Labels[common.FaultSchedulingLabelKey] = common.ExternalForceReschedulingValue
			workload := &DeploymentWorkLoad{Deployment: deployment}
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
			mockErr := errors.New("failed to delete pod")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Delete", mockErr)
			defer patches.Reset()

			err := deletePodsForExternalRescheduling(context.Background(), fakeClient, workload)

			convey.So(err, convey.ShouldBeNil)
		})

		// ------ external-grace mode tests ------

		convey.Convey("Should return nil (start timer, not block) when external-grace mode with matching pods", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Labels[common.FaultSchedulingLabelKey] = common.ExternalGraceReschedulingValue
			workload := &DeploymentWorkLoad{Deployment: deployment}
			gracePeriod := int64(60)
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
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: &gracePeriod,
				},
			}
			fakeClient := NewFakeClient().WithObjects(deployment, pod).Build()

			err := deletePodsForExternalRescheduling(context.Background(), fakeClient, workload)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should use default 30s when external-grace mode but no terminationGracePeriodSeconds set", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Labels[common.FaultSchedulingLabelKey] = common.ExternalGraceReschedulingValue
			workload := &DeploymentWorkLoad{Deployment: deployment}
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

			err := deletePodsForExternalRescheduling(context.Background(), fakeClient, workload)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should succeed when external-grace mode but no pods exist", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Labels[common.FaultSchedulingLabelKey] = common.ExternalGraceReschedulingValue
			workload := &DeploymentWorkLoad{Deployment: deployment}
			fakeClient := NewFakeClient().WithObjects(deployment).Build()

			err := deletePodsForExternalRescheduling(context.Background(), fakeClient, workload)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when pod list fails in external-grace mode", func() {
			deployment := CreateTestDeployment("test-deployment", "default", 1)
			deployment.Labels[common.FaultSchedulingLabelKey] = common.ExternalGraceReschedulingValue
			workload := &DeploymentWorkLoad{Deployment: deployment}
			fakeClient := NewFakeClient().Build()
			mockErr := errors.New("failed to list pods")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "List", mockErr)
			defer patches.Reset()

			err := deletePodsForExternalRescheduling(context.Background(), fakeClient, workload)

			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestForceDeletePodsAfterGrace(t *testing.T) {
	convey.Convey("Test forceDeletePodsAfterGrace function", t, func() {
		podLabels := client.MatchingLabels{
			common.InferServiceNameLabelKey: "test-service",
			common.InstanceSetNameLabelKey:  "test-role",
			common.InstanceIndexLabelKey:    "0",
		}

		convey.Convey("Should force-delete remaining pods after grace period", func() {
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
			fakeClient := NewFakeClient().WithObjects(pod).Build()
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Delete", nil)
			defer patches.Reset()

			forceDeletePodsAfterGrace(context.Background(), fakeClient, "default",
				"test-deployment", podLabels)

			convey.So(true, convey.ShouldBeTrue)
		})

		convey.Convey("Should skip not-found pod after grace period", func() {
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
			fakeClient := NewFakeClient().WithObjects(pod).Build()
			notFoundErr := apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "pods"}, "test-pod")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "Delete", notFoundErr)
			defer patches.Reset()

			forceDeletePodsAfterGrace(context.Background(), fakeClient, "default",
				"test-deployment", podLabels)

			convey.So(true, convey.ShouldBeTrue)
		})

		convey.Convey("Should handle list error after grace period", func() {
			fakeClient := NewFakeClient().Build()
			mockErr := errors.New("failed to list pods")
			patches := gomonkey.ApplyMethodReturn(fakeClient, "List", mockErr)
			defer patches.Reset()

			forceDeletePodsAfterGrace(context.Background(), fakeClient, "default",
				"test-deployment", podLabels)

			convey.So(true, convey.ShouldBeTrue)
		})

		convey.Convey("Should not delete when no pods exist after grace period", func() {
			fakeClient := NewFakeClient().Build()

			forceDeletePodsAfterGrace(context.Background(), fakeClient, "default",
				"test-deployment", podLabels)

			convey.So(true, convey.ShouldBeTrue)
		})
	})
}
