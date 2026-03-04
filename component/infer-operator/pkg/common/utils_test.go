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

package common

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/api/v1"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// TestDeepCopyLabelsMap tests the DeepCopyLabelsMap function.
func TestDeepCopyLabelsMap(t *testing.T) {
	convey.Convey("Test DeepCopyLabelsMap function", t, func() {
		convey.Convey("Should create a deep copy of labels map", func() {
			original := map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			copied := DeepCopyLabelsMap(original)

			convey.So(copied, convey.ShouldNotBeNil)
			convey.So(len(copied), convey.ShouldEqual, len(original))
			convey.So(copied["key1"], convey.ShouldEqual, "value1")
			convey.So(copied["key2"], convey.ShouldEqual, "value2")
		})

		convey.Convey("Should create independent copy", func() {
			original := map[string]string{
				"key1": "value1",
			}
			copied := DeepCopyLabelsMap(original)

			copied["key1"] = "modified"
			copied["key2"] = "new"

			convey.So(original["key1"], convey.ShouldEqual, "value1")
			originLen := 1
			convey.So(len(original), convey.ShouldEqual, originLen)
			copiedLen := 2
			convey.So(len(copied), convey.ShouldEqual, copiedLen)
		})

		convey.Convey("Should handle empty map", func() {
			original := map[string]string{}
			copied := DeepCopyLabelsMap(original)

			convey.So(copied, convey.ShouldNotBeNil)
			convey.So(len(copied), convey.ShouldEqual, 0)
		})

		convey.Convey("Should handle nil map", func() {
			var original map[string]string = nil
			copied := DeepCopyLabelsMap(original)

			convey.So(copied, convey.ShouldNotBeNil)
			convey.So(len(copied), convey.ShouldEqual, 0)
		})
	})
}

// TestWorkLoadTypeToGVK tests the WorkLoadTypeToGVK function.
func TestWorkLoadTypeToGVK(t *testing.T) {
	convey.Convey("Test WorkLoadTypeToGVK function", t, func() {
		convey.Convey("Should handle full APIVersion with group", func() {
			workloadType := v1.WorkloadType{
				Kind:       "CustomResource",
				APIVersion: "custom.example.com/v1alpha1",
			}

			gvk, err := WorkLoadTypeToGVK(workloadType)

			convey.So(err, convey.ShouldBeNil)
			convey.So(gvk.Kind, convey.ShouldEqual, "CustomResource")
			convey.So(gvk.Group, convey.ShouldEqual, "custom.example.com")
			convey.So(gvk.Version, convey.ShouldEqual, "v1alpha1")
		})

		convey.Convey("Should return error when ParseGroupVersion fails", func() {
			patches := gomonkey.ApplyFunc(schema.ParseGroupVersion, func(version string) (schema.GroupVersion, error) {
				return schema.GroupVersion{}, fmt.Errorf("mock error")
			})
			defer patches.Reset()

			workloadType := v1.WorkloadType{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			}
			_, err := WorkLoadTypeToGVK(workloadType)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestAddLabelsFromIndexer tests the AddLabelsFromIndexer function.
func TestAddLabelsFromIndexer(t *testing.T) {
	convey.Convey("Test AddLabelsFromIndexer function", t, func() {
		convey.Convey("Should add labels from indexer to labels map", func() {
			labels := make(map[string]string)
			indexer := InstanceIndexer{
				ServiceName:    "test-service",
				InstanceSetKey: "test-role",
				InstanceIndex:  "0",
			}
			labels = AddLabelsFromIndexer(labels, indexer)
			value, exists := labels[InferServiceNameLabelKey]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "test-service")
			value, exists = labels[InstanceSetNameLabelKey]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "test-role")
			value, exists = labels[InstanceIndexLabelKey]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "0")
			value, exists = labels[OperatorNameKey]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, TrueBool)
		})

		convey.Convey("Should overwrite existing indexer labels", func() {
			labels := map[string]string{
				InferServiceNameLabelKey: "old-service",
				InstanceSetNameLabelKey:  "old-role",
				InstanceIndexLabelKey:    "1",
			}
			indexer := InstanceIndexer{
				ServiceName:    "new-service",
				InstanceSetKey: "new-role",
				InstanceIndex:  "2",
			}
			labels = AddLabelsFromIndexer(labels, indexer)
			value, exists := labels[InferServiceNameLabelKey]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "new-service")
			value, exists = labels[InstanceSetNameLabelKey]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "new-role")
			value, exists = labels[InstanceIndexLabelKey]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "2")
		})
	})
}

// TestAddLabelsFromIndexer2 tests the AddLabelsFromIndexer function.
func TestAddLabelsFromIndexer2(t *testing.T) {
	convey.Convey("Test AddLabelsFromIndexer function", t, func() {
		convey.Convey("Should preserve existing labels", func() {
			labels := map[string]string{
				"existing-key": "existing-value",
			}
			indexer := InstanceIndexer{
				ServiceName:    "test-service",
				InstanceSetKey: "test-role",
				InstanceIndex:  "0",
			}
			labels = AddLabelsFromIndexer(labels, indexer)
			value, exists := labels["existing-key"]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "existing-value")
			value, exists = labels[InferServiceNameLabelKey]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "test-service")
		})

		convey.Convey("Should handle empty indexer values", func() {
			labels := make(map[string]string)
			indexer := InstanceIndexer{
				ServiceName:    "",
				InstanceSetKey: "",
				InstanceIndex:  "",
			}
			labels = AddLabelsFromIndexer(labels, indexer)
			_, exists := labels[InferServiceNameLabelKey]
			convey.So(exists, convey.ShouldBeTrue)
			_, exists = labels[InstanceSetNameLabelKey]
			convey.So(exists, convey.ShouldBeTrue)
			_, exists = labels[InstanceIndexLabelKey]
			convey.So(exists, convey.ShouldBeTrue)
			_, exists = labels[OperatorNameKey]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(labels[OperatorNameKey], convey.ShouldEqual, TrueBool)
		})
	})
}

// TestAddLabelsFromIndexer3 tests the AddLabelsFromIndexer function.
func TestAddLabelsFromIndexer3(t *testing.T) {
	convey.Convey("Test AddLabelsFromIndexer function", t, func() {
		convey.Convey("Should handle nil labels map", func() {
			var labels map[string]string = nil
			indexer := InstanceIndexer{
				ServiceName:    "test-service",
				InstanceSetKey: "test-role",
				InstanceIndex:  "0",
			}
			labels = AddLabelsFromIndexer(labels, indexer)
			convey.So(labels, convey.ShouldNotBeNil)
			value, exists := labels[InferServiceNameLabelKey]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "test-service")
		})
	})
}

// TestGetInstanceSetNameFromLabels tests the GetInstanceSetNameFromLabels function.
func TestGetInstanceSetNameFromLabels(t *testing.T) {
	convey.Convey("Test GetInstanceSetNameFromLabels function", t, func() {
		convey.Convey("Should generate InstanceSet name from labels", func() {
			labels := map[string]string{
				InferServiceNameLabelKey: "test-service",
				InstanceSetNameLabelKey:  "test-role",
			}

			name := GetInstanceSetNameFromLabels(labels)

			convey.So(name, convey.ShouldEqual, "test-service-test-role")
		})
	})
}

// TestGetWorkLoadNameFromIndexer tests the GetWorkLoadNameFromIndexer function.
func TestGetWorkLoadNameFromIndexer(t *testing.T) {
	convey.Convey("Test GetWorkLoadNameFromIndexer function", t, func() {
		convey.Convey("Should generate workload name from indexer", func() {
			indexer := InstanceIndexer{
				ServiceName:    "test-service",
				InstanceSetKey: "test-role",
				InstanceIndex:  "0",
			}

			name := GetWorkLoadNameFromIndexer(indexer)

			convey.So(name, convey.ShouldEqual, "test-service-test-role-0")
		})
	})
}

// TestGetServiceNameFromIndexer tests the GetServiceNameFromIndexer function.
func TestGetServiceNameFromIndexer(t *testing.T) {
	convey.Convey("Test GetServiceNameFromIndexer function", t, func() {
		convey.Convey("Should generate service name from indexer", func() {
			indexer := InstanceIndexer{
				ServiceName:    "test-service",
				InstanceSetKey: "test-role",
				InstanceIndex:  "0",
			}

			name := GetServiceNameFromIndexer(indexer)
			convey.So(name, convey.ShouldEqual, "service-test-service-test-role-0")
		})
	})
}

// TestGetPGNameFromIndexer tests the GetPGNameFromIndexer function.
func TestGetPGNameFromIndexer(t *testing.T) {
	convey.Convey("Test GetPGNameFromIndexer function", t, func() {
		convey.Convey("Should generate PodGroup name from indexer", func() {
			indexer := InstanceIndexer{
				ServiceName:    "test-service",
				InstanceSetKey: "test-role",
				InstanceIndex:  "0",
			}

			name := GetPGNameFromIndexer(indexer)

			convey.So(name, convey.ShouldEqual, "pg-test-service-test-role-0")
		})
	})
}

// TestAddEnvToPodTemplate tests the AddEnvToPodTemplate function.
func TestAddEnvToPodTemplate(t *testing.T) {
	convey.Convey("Test AddEnvToPodTemplate function", t, func() {
		convey.Convey("Should add environment variables to pod template", func() {
			podTemplate := &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Env: []corev1.EnvVar{
								{Name: "EXISTING_ENV", Value: "existing_value"},
							},
						},
					},
				},
			}
			indexer := InstanceIndexer{
				ServiceName:    "test-service-0",
				InstanceSetKey: "test-role",
				InstanceIndex:  "0",
			}

			AddEnvToPodTemplate(podTemplate, indexer)

			convey.So(len(podTemplate.Spec.Containers[0].Env), convey.ShouldEqual, 5)
			convey.So(podTemplate.Spec.Containers[0].Env[0].Name, convey.ShouldEqual, "EXISTING_ENV")
			convey.So(podTemplate.Spec.Containers[0].Env[1].Name, convey.ShouldEqual, InstanceIndexEnvKey)
			convey.So(podTemplate.Spec.Containers[0].Env[1].Value, convey.ShouldEqual, "0")
			convey.So(podTemplate.Spec.Containers[0].Env[2].Name, convey.ShouldEqual, InstanceRoleEnvKey)
			convey.So(podTemplate.Spec.Containers[0].Env[2].Value, convey.ShouldEqual, "test-role")
			convey.So(podTemplate.Spec.Containers[0].Env[3].Name, convey.ShouldEqual, InferServiceIndexEnvKey)
			convey.So(podTemplate.Spec.Containers[0].Env[3].Value, convey.ShouldEqual, "0")
			convey.So(podTemplate.Spec.Containers[0].Env[4].Name, convey.ShouldEqual, InferServiceNameEnvKey)
			convey.So(podTemplate.Spec.Containers[0].Env[4].Value, convey.ShouldEqual, "test-service")
		})
	})
}

// TestAddEnvToPodTemplate2 tests the AddEnvToPodTemplate function.
func TestAddEnvToPodTemplate2(t *testing.T) {
	convey.Convey("Test AddEnvToPodTemplate function", t, func() {
		convey.Convey("Should handle multiple containers", func() {
			podTemplate := &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "container1"},
						{Name: "container2"},
					},
				},
			}
			indexer := InstanceIndexer{
				ServiceName:    "test-service",
				InstanceSetKey: "test-role",
				InstanceIndex:  "1",
			}

			AddEnvToPodTemplate(podTemplate, indexer)

			for _, container := range podTemplate.Spec.Containers {
				envNum := 4
				convey.So(len(container.Env), convey.ShouldEqual, envNum)
				convey.So(container.Env[0].Name, convey.ShouldEqual, InstanceIndexEnvKey)
				convey.So(container.Env[0].Value, convey.ShouldEqual, "1")
				convey.So(container.Env[1].Name, convey.ShouldEqual, InstanceRoleEnvKey)
				convey.So(container.Env[1].Value, convey.ShouldEqual, "test-role")
			}
		})

		convey.Convey("Should handle empty containers list", func() {
			podTemplate := &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			}
			indexer := InstanceIndexer{
				ServiceName:    "test-service",
				InstanceSetKey: "test-role",
				InstanceIndex:  "0",
			}

			AddEnvToPodTemplate(podTemplate, indexer)

			convey.So(len(podTemplate.Spec.Containers), convey.ShouldEqual, 0)
		})
	})
}

// TestAddEnvToPodTemplate3 tests the AddEnvToPodTemplate function.
func TestAddEnvToPodTemplate3(t *testing.T) {
	convey.Convey("Test AddEnvToPodTemplate function", t, func() {
		convey.Convey("Should handle nil pod template", func() {
			var podTemplate *corev1.PodTemplateSpec = nil
			indexer := InstanceIndexer{
				ServiceName:    "test-service",
				InstanceSetKey: "test-role",
				InstanceIndex:  "0",
			}

			defer func() {
				r := recover()
				convey.So(r, convey.ShouldNotBeNil)
			}()

			AddEnvToPodTemplate(podTemplate, indexer)
		})
	})
}

// TestIsRequeueError tests the IsRequeueError function.
func TestIsRequeueError(t *testing.T) {
	convey.Convey("Test IsRequeueError function", t, func() {
		convey.Convey("Should return true for RequeueError", func() {
			err := NewRequeueError("test requeue error")
			result := IsRequeueError(err)

			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("Should return false for nil error", func() {
			result := IsRequeueError(nil)

			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false for standard error", func() {
			err := errors.New("standard error")
			result := IsRequeueError(err)

			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return true for pointer to RequeueError", func() {
			err := &RequeueError{Message: "test message"}
			result := IsRequeueError(err)

			convey.So(result, convey.ShouldBeTrue)
		})
	})
}
