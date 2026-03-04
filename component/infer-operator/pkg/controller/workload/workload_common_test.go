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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
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
