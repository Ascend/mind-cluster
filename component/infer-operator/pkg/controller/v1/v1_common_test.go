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

package v1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

// NewFakeClient returns a new fake client with the given runtime objects
func NewFakeClient(objects ...runtime.Object) *fake.ClientBuilder {
	return fake.NewClientBuilder().WithScheme(GetScheme()).WithRuntimeObjects(objects...)
}

// mockStatusWriter mocks client.StatusWriter interface
type mockStatusWriter struct {
	updateErr error
}

func (m *mockStatusWriter) Update(context.Context, client.Object, ...client.UpdateOption) error {
	return m.updateErr
}

func (m *mockStatusWriter) Patch(context.Context, client.Object, client.Patch, ...client.PatchOption) error {
	return nil
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

// CreateTestInferService creates a test InferService object.
func CreateTestInferService(name, namespace string) *v1.InferService {
	return &v1.InferService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.InferServiceSpec{
			Roles: []v1.InstanceSetSpec{
				{
					Name:     "role1",
					Replicas: func() *int32 { r := int32(1); return &r }(),
				},
			},
		},
	}
}

// CreateTestInferServiceSet creates a test InferServiceSet object.
func CreateTestInferServiceSet(name, namespace string, replicas int32) *v1.InferServiceSet {
	return &v1.InferServiceSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.InferServiceSetSpec{
			Replicas: &replicas,
			InferServiceTemplate: v1.InferServiceSpec{
				Roles: []v1.InstanceSetSpec{
					{
						Name:     "role1",
						Replicas: func() *int32 { r := int32(1); return &r }(),
					},
				},
			},
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

// GetTestIndexer returns a test InstanceIndexer object.
func GetTestIndexer(serviceName, instanceSetKey, instanceIndex string) common.InstanceIndexer {
	return common.InstanceIndexer{
		ServiceName:    serviceName,
		InstanceSetKey: instanceSetKey,
		InstanceIndex:  instanceIndex,
	}
}
