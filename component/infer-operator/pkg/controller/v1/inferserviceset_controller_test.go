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

// Package v1 contains API Schema definitions for the mindcluster v1 API group
package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"

	apiv1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

func TestGetInferServiceSetNotFound(t *testing.T) {
	convey.Convey("TestGetInferServiceSet not found", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceSetReconciler{client: fakeClient}

		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}

		result, err := reconciler.getInferServiceSet(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result == nil, convey.ShouldBeTrue)
	})
}

func TestGetInferServiceSetBeingDeleted(t *testing.T) {
	convey.Convey("TestGetInferServiceSet being deleted", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceSetReconciler{client: fakeClient}

		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}

		iss := &apiv1.InferServiceSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test",
				Namespace:         "default",
				DeletionTimestamp: &metav1.Time{Time: metav1.Now().Time},
			},
		}
		err := fakeClient.Create(context.Background(), iss)
		convey.So(err, convey.ShouldBeNil)

		result, err := reconciler.getInferServiceSet(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result == nil, convey.ShouldBeTrue)
	})
}

func TestGetInferServiceSetSuccess(t *testing.T) {
	convey.Convey("TestGetInferServiceSet success", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceSetReconciler{client: fakeClient}

		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}

		iss := &apiv1.InferServiceSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}
		err := fakeClient.Create(context.Background(), iss)
		convey.So(err, convey.ShouldBeNil)

		result, err := reconciler.getInferServiceSet(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldNotBeNil)
	})
}

func TestISSValidateNilReplicas(t *testing.T) {
	convey.Convey("TestValidate nil replicas", t, func() {
		reconciler := &InferServiceSetReconciler{}
		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}

		iss := &apiv1.InferServiceSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}
		err := reconciler.validate(context.Background(), iss, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(*iss.Spec.Replicas, convey.ShouldEqual, 1)
	})
}

func TestISSValidateReplicasExceedsMax(t *testing.T) {
	convey.Convey("TestValidate replicas exceeds max", t, func() {
		reconciler := &InferServiceSetReconciler{}
		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}

		replicas := int32(common.MaxInferServiceReplicas + 1)
		iss := &apiv1.InferServiceSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: apiv1.InferServiceSetSpec{
				Replicas: &replicas,
			},
		}
		err := reconciler.validate(context.Background(), iss, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestISSValidateReplicasLessThanOne(t *testing.T) {
	convey.Convey("TestValidate replicas less than 1", t, func() {
		reconciler := &InferServiceSetReconciler{}
		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}

		replicas := int32(0)
		iss := &apiv1.InferServiceSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: apiv1.InferServiceSetSpec{
				Replicas: &replicas,
			},
		}
		err := reconciler.validate(context.Background(), iss, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestISSValidateValidReplicas(t *testing.T) {
	convey.Convey("TestValidate valid replicas", t, func() {
		reconciler := &InferServiceSetReconciler{}
		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}

		replicas := int32(3)
		iss := &apiv1.InferServiceSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: apiv1.InferServiceSetSpec{
				Replicas: &replicas,
			},
		}
		err := reconciler.validate(context.Background(), iss, req)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestListInferServices(t *testing.T) {
	convey.Convey("TestListInferServices", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceSetReconciler{client: fakeClient}

		iss := &apiv1.InferServiceSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}

		convey.Convey("list error", func() {
			patch := gomonkey.ApplyMethodFunc(fakeClient, "List",
				func(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
					return errors.New("list error")
				})
			defer patch.Reset()

			_, _, err := reconciler.listInferServices(context.Background(), iss)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("success", func() {
			result, selector, err := reconciler.listInferServices(context.Background(), iss)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(selector, convey.ShouldNotBeNil)
		})
	})
}

func TestBuildInferServiceMapEmptyList(t *testing.T) {
	convey.Convey("TestBuildInferServiceMap empty list", t, func() {
		reconciler := &InferServiceSetReconciler{}
		isList := &apiv1.InferServiceList{}
		result := reconciler.buildInferServiceMap(isList)
		convey.So(result, convey.ShouldBeEmpty)
	})
}

func TestBuildInferServiceMapWithValidIndex(t *testing.T) {
	convey.Convey("TestBuildInferServiceMap with valid index", t, func() {
		reconciler := &InferServiceSetReconciler{}
		isList := &apiv1.InferServiceList{
			Items: []apiv1.InferService{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-0",
						Namespace: "default",
						Labels:    map[string]string{common.InferServiceIndexLabelKey: "0"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-1",
						Namespace: "default",
						Labels:    map[string]string{common.InferServiceIndexLabelKey: "1"},
					},
				},
			},
		}
		result := reconciler.buildInferServiceMap(isList)
		convey.So(len(result), convey.ShouldEqual, int2)
		convey.So(result[0].Name, convey.ShouldEqual, "test-0")
		convey.So(result[1].Name, convey.ShouldEqual, "test-1")
	})
}

func TestBuildInferServiceMapWithoutIndexLabel(t *testing.T) {
	convey.Convey("TestBuildInferServiceMap without index label", t, func() {
		reconciler := &InferServiceSetReconciler{}
		isList := &apiv1.InferServiceList{
			Items: []apiv1.InferService{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
				},
			},
		}
		result := reconciler.buildInferServiceMap(isList)
		convey.So(result, convey.ShouldBeEmpty)
	})
}

func TestBuildInferServiceMapInvalidIndex(t *testing.T) {
	convey.Convey("TestBuildInferServiceMap invalid index", t, func() {
		reconciler := &InferServiceSetReconciler{}
		isList := &apiv1.InferServiceList{
			Items: []apiv1.InferService{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
						Labels:    map[string]string{common.InferServiceIndexLabelKey: "invalid"},
					},
				},
			},
		}
		result := reconciler.buildInferServiceMap(isList)
		convey.So(result, convey.ShouldBeEmpty)
	})
}

func testISS() *apiv1.InferServiceSet {
	replicas := int32(2)
	return &apiv1.InferServiceSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: apiv1.InferServiceSetSpec{
			Replicas: &replicas,
			InferServiceTemplate: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{{Name: "template"}},
			},
		},
	}
}

func TestCalculateInferServiceOperationsCreateNew(t *testing.T) {
	convey.Convey("TestCalculateInferServiceOperations create new", t, func() {
		reconciler := &InferServiceSetReconciler{}
		iss := testISS()
		existedIsList := make(map[int]*apiv1.InferService)
		isToCreate, isToUpdate, isToDelete := reconciler.calculateInferServiceOperations(iss, existedIsList)
		convey.So(len(isToCreate), convey.ShouldEqual, int2)
		convey.So(len(isToUpdate), convey.ShouldEqual, 0)
		convey.So(len(isToDelete), convey.ShouldEqual, 0)
	})
}

func TestCalculateInferServiceOperationsUpdateExisting(t *testing.T) {
	convey.Convey("TestCalculateInferServiceOperations update existing", t, func() {
		reconciler := &InferServiceSetReconciler{}
		iss := testISS()
		existedIsList := map[int]*apiv1.InferService{
			0: {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-0",
					Namespace: "default",
					Labels:    map[string]string{common.InferServiceIndexLabelKey: "0"},
				},
				Spec: apiv1.InferServiceSpec{Roles: []apiv1.InstanceSetSpec{{Name: "old-template"}}},
			},
		}
		isToCreate, isToUpdate, isToDelete := reconciler.calculateInferServiceOperations(iss, existedIsList)
		convey.So(len(isToCreate), convey.ShouldEqual, 1)
		convey.So(len(isToUpdate), convey.ShouldEqual, 1)
		convey.So(len(isToDelete), convey.ShouldEqual, 0)
	})
}

func TestCalculateInferServiceOperationsDeleteExtra(t *testing.T) {
	convey.Convey("TestCalculateInferServiceOperations delete extra", t, func() {
		reconciler := &InferServiceSetReconciler{}
		iss := testISS()
		existedIsList := map[int]*apiv1.InferService{
			0: {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-0",
					Namespace: "default",
					Labels:    map[string]string{common.InferServiceIndexLabelKey: "0"},
				},
				Spec: apiv1.InferServiceSpec{Roles: []apiv1.InstanceSetSpec{{Name: "template"}}},
			},
			2: {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-2",
					Namespace: "default",
					Labels:    map[string]string{common.InferServiceIndexLabelKey: "2"},
				},
				Spec: apiv1.InferServiceSpec{Roles: []apiv1.InstanceSetSpec{{Name: "template"}}},
			},
		}
		isToCreate, isToUpdate, isToDelete := reconciler.calculateInferServiceOperations(iss, existedIsList)
		convey.So(len(isToCreate), convey.ShouldEqual, 1)
		convey.So(len(isToUpdate), convey.ShouldEqual, 0)
		convey.So(len(isToDelete), convey.ShouldEqual, 1)
	})
}

func testISSWithTemplate() *apiv1.InferServiceSet {
	return &apiv1.InferServiceSet{
		Spec: apiv1.InferServiceSetSpec{
			InferServiceTemplate: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{{Name: "template"}},
			},
		},
	}
}

func TestInferServiceUpdatedNilISS(t *testing.T) {
	convey.Convey("TestInferServiceUpdated nil iss", t, func() {
		reconciler := &InferServiceSetReconciler{}
		is := &apiv1.InferService{}
		result := reconciler.inferServiceUpdated(nil, is)
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestInferServiceUpdatedNilIS(t *testing.T) {
	convey.Convey("TestInferServiceUpdated nil is", t, func() {
		reconciler := &InferServiceSetReconciler{}
		iss := testISSWithTemplate()
		result := reconciler.inferServiceUpdated(iss, nil)
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestInferServiceUpdatedSpecDifferent(t *testing.T) {
	convey.Convey("TestInferServiceUpdated spec different", t, func() {
		reconciler := &InferServiceSetReconciler{}
		iss := testISSWithTemplate()
		is := &apiv1.InferService{
			Spec: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{{Name: "different"}},
			},
		}
		result := reconciler.inferServiceUpdated(iss, is)
		convey.So(result, convey.ShouldBeTrue)
	})
}

func TestInferServiceUpdatedSpecSame(t *testing.T) {
	convey.Convey("TestInferServiceUpdated spec same", t, func() {
		reconciler := &InferServiceSetReconciler{}
		iss := testISSWithTemplate()
		is := &apiv1.InferService{
			Spec: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{{Name: "template"}},
			},
		}
		result := reconciler.inferServiceUpdated(iss, is)
		convey.So(result, convey.ShouldBeFalse)
	})
}

func testISSForScale() *apiv1.InferServiceSet {
	return &apiv1.InferServiceSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
}

func TestScaleDownNilISS(t *testing.T) {
	convey.Convey("TestScaleDown nil iss", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceSetReconciler{client: fakeClient}

		err := reconciler.scaleDown(context.Background(), nil, nil)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestScaleDownDeleteError(t *testing.T) {
	convey.Convey("TestScaleDown delete error", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceSetReconciler{client: fakeClient}

		iss := testISSForScale()
		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-0",
				Namespace: "default",
			},
		}
		err := fakeClient.Create(context.Background(), is)
		convey.So(err, convey.ShouldBeNil)

		patch := gomonkey.ApplyMethodFunc(fakeClient, "Delete",
			func(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
				return errors.New("delete error")
			})
		defer patch.Reset()

		err = reconciler.scaleDown(context.Background(), iss, []*apiv1.InferService{is})
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestScaleDownNotFound(t *testing.T) {
	convey.Convey("TestScaleDown not found", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceSetReconciler{client: fakeClient}

		iss := testISSForScale()
		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-0",
				Namespace: "default",
			},
		}
		err := reconciler.scaleDown(context.Background(), iss, []*apiv1.InferService{is})
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestScaleDownSuccess(t *testing.T) {
	convey.Convey("TestScaleDown success", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceSetReconciler{client: fakeClient}

		iss := testISSForScale()
		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-0",
				Namespace: "default",
			},
		}
		err := fakeClient.Create(context.Background(), is)
		convey.So(err, convey.ShouldBeNil)

		err = reconciler.scaleDown(context.Background(), iss, []*apiv1.InferService{is})
		convey.So(err, convey.ShouldBeNil)
	})
}

func testReconcilerForScaleUp() (*InferServiceSetReconciler, client.Client) {
	scheme := runtime.NewScheme()
	apiv1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	return &InferServiceSetReconciler{
		client: fakeClient,
		scheme: scheme,
	}, fakeClient
}

func TestScaleUpNilISS(t *testing.T) {
	convey.Convey("TestScaleUp nil iss", t, func() {
		reconciler, _ := testReconcilerForScaleUp()
		err := reconciler.scaleUp(context.Background(), nil, nil)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestScaleUpCreateError(t *testing.T) {
	convey.Convey("TestScaleUp create error", t, func() {
		reconciler, fakeClient := testReconcilerForScaleUp()
		iss := testISSForScale()
		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-0",
				Namespace: "default",
			},
		}

		patch := gomonkey.ApplyMethodFunc(fakeClient, "Create",
			func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
				return errors.New("create error")
			})
		defer patch.Reset()

		err := reconciler.scaleUp(context.Background(), iss, []*apiv1.InferService{is})
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestScaleUpAlreadyExists(t *testing.T) {
	convey.Convey("TestScaleUp already exists", t, func() {
		reconciler, fakeClient := testReconcilerForScaleUp()
		iss := testISSForScale()
		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-0",
				Namespace: "default",
			},
		}
		err := fakeClient.Create(context.Background(), is)
		convey.So(err, convey.ShouldBeNil)

		is2 := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-0",
				Namespace: "default",
			},
		}
		err = reconciler.scaleUp(context.Background(), iss, []*apiv1.InferService{is2})
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestScaleUpSuccess(t *testing.T) {
	convey.Convey("TestScaleUp success", t, func() {
		reconciler, _ := testReconcilerForScaleUp()
		iss := testISSForScale()
		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-0",
				Namespace: "default",
			},
		}
		err := reconciler.scaleUp(context.Background(), iss, []*apiv1.InferService{is})
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestNewInferService(t *testing.T) {
	convey.Convey("TestNewInferService", t, func() {
		iss := &apiv1.InferServiceSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: apiv1.InferServiceSetSpec{
				InferServiceTemplate: apiv1.InferServiceSpec{
					Roles: []apiv1.InstanceSetSpec{{Name: "template"}},
				},
			},
		}

		result := newInferService(iss, 0)
		convey.So(result.Name, convey.ShouldEqual, "test-0")
		convey.So(result.Namespace, convey.ShouldEqual, "default")
		convey.So(result.Labels[common.InferServiceSetNameLabelKey], convey.ShouldEqual, "test")
		convey.So(result.Labels[common.InferServiceIndexLabelKey], convey.ShouldEqual, "0")
		convey.So(len(result.Spec.Roles), convey.ShouldEqual, 1)
		convey.So(result.Spec.Roles[0].Name, convey.ShouldEqual, "template")
	})
}

func TestISSCalculateStatus(t *testing.T) {
	convey.Convey("TestCalculateStatus", t, func() {
		reconciler := &InferServiceSetReconciler{}

		replicas := int32(2)
		iss := &apiv1.InferServiceSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test",
				Namespace:  "default",
				Generation: 1,
			},
			Spec: apiv1.InferServiceSetSpec{
				Replicas: &replicas,
			},
		}

		convey.Convey("calculate status", func() {
			isList := &apiv1.InferServiceList{
				Items: []apiv1.InferService{
					{
						Status: apiv1.InferServiceStatus{
							Conditions: []metav1.Condition{
								{Type: string(common.InferServiceSetReady), Status: metav1.ConditionTrue},
							},
						},
					},
				},
			}
			result := reconciler.calculateStatus(iss, isList)
			convey.So(result.Replicas, convey.ShouldEqual, int2)
			convey.So(result.ReadyReplicas, convey.ShouldEqual, 1)
			convey.So(result.ObservedGeneration, convey.ShouldEqual, 1)
		})
	})
}

func TestCalculateReadyReplicas(t *testing.T) {
	convey.Convey("TestCalculateReadyReplicas", t, func() {
		reconciler := &InferServiceSetReconciler{}

		convey.Convey("empty list", func() {
			isList := &apiv1.InferServiceList{}
			result := reconciler.calculateReadyReplicas(isList)
			convey.So(result, convey.ShouldEqual, 0)
		})

		convey.Convey("with ready replicas", func() {
			isList := &apiv1.InferServiceList{
				Items: []apiv1.InferService{
					{
						Status: apiv1.InferServiceStatus{
							Conditions: []metav1.Condition{
								{Type: string(common.InferServiceSetReady), Status: metav1.ConditionTrue},
							},
						},
					},
					{
						Status: apiv1.InferServiceStatus{
							Conditions: []metav1.Condition{
								{Type: string(common.InferServiceSetReady), Status: metav1.ConditionFalse},
							},
						},
					},
				},
			}
			result := reconciler.calculateReadyReplicas(isList)
			convey.So(result, convey.ShouldEqual, 1)
		})
	})
}

func TestCreateReadyCondition(t *testing.T) {
	convey.Convey("TestCreateReadyCondition", t, func() {
		reconciler := &InferServiceSetReconciler{}

		convey.Convey("all ready", func() {
			result := reconciler.createReadyCondition(int2, int2)
			convey.So(result.Type, convey.ShouldEqual, string(common.InferServiceSetReady))
			convey.So(result.Status, convey.ShouldEqual, metav1.ConditionTrue)
			convey.So(result.Reason, convey.ShouldEqual, "AllInferServicesReady")
		})

		convey.Convey("not ready", func() {
			result := reconciler.createReadyCondition(1, int2)
			convey.So(result.Type, convey.ShouldEqual, string(common.InferServiceSetReady))
			convey.So(result.Status, convey.ShouldEqual, metav1.ConditionFalse)
			convey.So(result.Reason, convey.ShouldEqual, "ReplicasNotReady")
		})
	})
}

func TestFilterInferServiceEventsCreateFunc(t *testing.T) {
	convey.Convey("TestFilterInferServiceEvents CreateFunc", t, func() {
		filter := filterInferServiceEvents()
		result := filter.CreateFunc(event.CreateEvent{})
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestFilterInferServiceEventsUpdateFuncWrongType(t *testing.T) {
	convey.Convey("TestFilterInferServiceEvents UpdateFunc wrong type", t, func() {
		filter := filterInferServiceEvents()
		result := filter.UpdateFunc(event.UpdateEvent{})
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestFilterInferServiceEventsUpdateFuncNoChange(t *testing.T) {
	convey.Convey("TestFilterInferServiceEvents UpdateFunc no change", t, func() {
		filter := filterInferServiceEvents()
		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
				Labels:    map[string]string{common.InferServiceSetNameLabelKey: "test"},
			},
		}
		result := filter.UpdateFunc(event.UpdateEvent{
			ObjectOld: is,
			ObjectNew: is,
		})
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestFilterInferServiceEventsUpdateFuncWithChange(t *testing.T) {
	convey.Convey("TestFilterInferServiceEvents UpdateFunc with change", t, func() {
		filter := filterInferServiceEvents()
		oldIs := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
				Labels:    map[string]string{common.InferServiceSetNameLabelKey: "test"},
			},
			Spec: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{{Name: "role1"}},
			},
		}
		newIs := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
				Labels:    map[string]string{common.InferServiceSetNameLabelKey: "test"},
			},
			Spec: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{{Name: "role2"}},
			},
		}
		result := filter.UpdateFunc(event.UpdateEvent{
			ObjectOld: oldIs,
			ObjectNew: newIs,
		})
		convey.So(result, convey.ShouldBeTrue)
	})
}

func TestFilterInferServiceEventsDeleteFuncWrongType(t *testing.T) {
	convey.Convey("TestFilterInferServiceEvents DeleteFunc wrong type", t, func() {
		filter := filterInferServiceEvents()
		result := filter.DeleteFunc(event.DeleteEvent{})
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestFilterInferServiceEventsDeleteFuncWithLabel(t *testing.T) {
	convey.Convey("TestFilterInferServiceEvents DeleteFunc with label", t, func() {
		filter := filterInferServiceEvents()
		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
				Labels:    map[string]string{common.InferServiceSetNameLabelKey: "test"},
			},
		}
		result := filter.DeleteFunc(event.DeleteEvent{Object: is})
		convey.So(result, convey.ShouldBeTrue)
	})
}

func TestFilterInferServiceEventsDeleteFuncWithoutLabel(t *testing.T) {
	convey.Convey("TestFilterInferServiceEvents DeleteFunc without label", t, func() {
		filter := filterInferServiceEvents()
		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}
		result := filter.DeleteFunc(event.DeleteEvent{Object: is})
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestFilterInferServiceEventsGenericFunc(t *testing.T) {
	convey.Convey("TestFilterInferServiceEvents GenericFunc", t, func() {
		filter := filterInferServiceEvents()
		result := filter.GenericFunc(event.GenericEvent{})
		convey.So(result, convey.ShouldBeFalse)
	})
}
