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

package crd

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apiv1 "infer-operator/pkg/api/v1"
)

func TestGetInferServiceGVK(t *testing.T) {
	convey.Convey("GetInferServiceGVK should return correct GVK for InferService", t, func() {
		expectedGVK := schema.FromAPIVersionAndKind(apiv1.GroupVersion.String(), "InferService")
		actualGVK := GetInferServiceGVK()

		convey.So(actualGVK.Group, convey.ShouldEqual, expectedGVK.Group)
		convey.So(actualGVK.Version, convey.ShouldEqual, expectedGVK.Version)
		convey.So(actualGVK.Kind, convey.ShouldEqual, expectedGVK.Kind)

		expectedAPIVersion := apiv1.GroupVersion.String()
		convey.So(actualGVK.GroupVersion().String(), convey.ShouldEqual, expectedAPIVersion)
		convey.So(actualGVK.Kind, convey.ShouldEqual, "InferService")
	})
}

func TestCRDExists(t *testing.T) {
	var scheme *runtime.Scheme
	convey.Convey("Initialize scheme", t, func() {
		scheme = runtime.NewScheme()
		err := v1.AddToScheme(scheme)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("Test CRDExists scenarios", t, func() {
		convey.Convey("CRD exists and is established", func() {
			crd := &v1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: "test-crd"},
				Status: v1.CustomResourceDefinitionStatus{
					Conditions: []v1.CustomResourceDefinitionCondition{{
						Type:   v1.Established,
						Status: v1.ConditionTrue,
					}},
				},
			}
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(crd).Build()
			err := CRDExists(context.Background(), fakeClient, "test-crd")
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("CRD exists but not established", func() {
			crd := &v1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: "test-crd"},
				Status: v1.CustomResourceDefinitionStatus{
					Conditions: []v1.CustomResourceDefinitionCondition{{
						Type:   v1.Established,
						Status: v1.ConditionFalse,
					}},
				},
			}
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(crd).Build()
			err := CRDExists(context.Background(), fakeClient, "test-crd")
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("CRD does not exist", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			err := CRDExists(context.Background(), fakeClient, "non-existent-crd")
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("CRD exists with no conditions", func() {
			crd := &v1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: "test-crd"},
				Status:     v1.CustomResourceDefinitionStatus{Conditions: []v1.CustomResourceDefinitionCondition{}},
			}
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(crd).Build()
			err := CRDExists(context.Background(), fakeClient, "test-crd")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestOwnedByGVK(t *testing.T) {
	targetGVK := schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}

	convey.Convey("Test OwnedByGVK scenarios", t, func() {
		convey.Convey("Object is owned by target GVK", func() {
			obj := &fakeObject{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "test-deployment",
					}},
				},
			}
			convey.So(OwnedByGVK(obj, targetGVK), convey.ShouldBeTrue)
		})

		convey.Convey("Object is not owned by target GVK", func() {
			obj := &fakeObject{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						APIVersion: "apps/v1",
						Kind:       "StatefulSet",
						Name:       "test-statefulset",
					}},
				},
			}
			convey.So(OwnedByGVK(obj, targetGVK), convey.ShouldBeFalse)
		})

		convey.Convey("Object has no owner references", func() {
			obj := &fakeObject{ObjectMeta: metav1.ObjectMeta{OwnerReferences: []metav1.OwnerReference{}}}
			convey.So(OwnedByGVK(obj, targetGVK), convey.ShouldBeFalse)
		})
	})
}

func TestOwnedByGVK2(t *testing.T) {
	targetGVK := schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}

	convey.Convey("Test OwnedByGVK scenarios", t, func() {
		convey.Convey("Object has multiple owner references", func() {
			obj := &fakeObject{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{APIVersion: "apps/v1", Kind: "StatefulSet", Name: "test-statefulset"},
						{APIVersion: "apps/v1", Kind: "Deployment", Name: "test-deployment"},
					},
				},
			}
			convey.So(OwnedByGVK(obj, targetGVK), convey.ShouldBeTrue)
		})

		convey.Convey("Owner reference matches group but not version", func() {
			obj := &fakeObject{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						APIVersion: "apps/v2",
						Kind:       "Deployment",
						Name:       "test-deployment",
					}},
				},
			}
			convey.So(OwnedByGVK(obj, targetGVK), convey.ShouldBeFalse)
		})

		convey.Convey("Owner reference matches version but not group", func() {
			obj := &fakeObject{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						APIVersion: "batch/v1",
						Kind:       "Deployment",
						Name:       "test-deployment",
					}},
				},
			}
			convey.So(OwnedByGVK(obj, targetGVK), convey.ShouldBeFalse)
		})
	})
}

type fakeObject struct {
	metav1.ObjectMeta
}

func (f *fakeObject) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

func (f *fakeObject) DeepCopyObject() runtime.Object {
	return &fakeObject{
		ObjectMeta: *f.ObjectMeta.DeepCopy(),
	}
}
