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
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"

	ascendapi "ascend-common/api"
	"ascend-common/common-utils/hwlog"
	apiv1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

const (
	int2  = 2
	int10 = 10
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestValidateRolesNilInferService(t *testing.T) {
	convey.Convey("TestValidateRoles nil InferService", t, func() {
		reconciler := &InferServiceReconciler{}
		err := reconciler.validateRoles(nil)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "cannot be nil")
	})
}

func TestValidateRolesExceedMaxCount(t *testing.T) {
	convey.Convey("TestValidateRoles roles exceed max count", t, func() {
		reconciler := &InferServiceReconciler{}
		roles := make([]apiv1.InstanceSetSpec, common.MaxRoleTypeCount+1)
		for i := range roles {
			roles[i] = apiv1.InstanceSetSpec{Name: fmt.Sprintf("role%d", i)}
		}
		is := &apiv1.InferService{
			Spec: apiv1.InferServiceSpec{Roles: roles},
		}
		err := reconciler.validateRoles(is)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "exceeds maximum allowed")
	})
}

func TestValidateRolesEmptyRoleName(t *testing.T) {
	convey.Convey("TestValidateRoles empty role name", t, func() {
		reconciler := &InferServiceReconciler{}
		is := &apiv1.InferService{
			Spec: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{{Name: ""}},
			},
		}
		err := reconciler.validateRoles(is)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "empty name")
	})
}

func TestValidateRolesDuplicateRoleName(t *testing.T) {
	convey.Convey("TestValidateRoles duplicate role name", t, func() {
		reconciler := &InferServiceReconciler{}
		is := &apiv1.InferService{
			Spec: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{
					{Name: "role1"},
					{Name: "role1"},
				},
			},
		}
		err := reconciler.validateRoles(is)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "duplicate")
	})
}

func TestValidateRolesInvalidRoleName(t *testing.T) {
	convey.Convey("TestValidateRoles invalid role name", t, func() {
		reconciler := &InferServiceReconciler{}
		is := &apiv1.InferService{
			Spec: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{{Name: "Invalid_Name"}},
			},
		}
		err := reconciler.validateRoles(is)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")
	})
}

func TestValidateRolesValidRoles(t *testing.T) {
	convey.Convey("TestValidateRoles valid roles", t, func() {
		reconciler := &InferServiceReconciler{}
		is := &apiv1.InferService{
			Spec: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{
					{Name: "role1"},
					{Name: "role2"},
				},
			},
		}
		err := reconciler.validateRoles(is)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestValidate(t *testing.T) {
	convey.Convey("TestValidate", t, func() {
		reconciler := &InferServiceReconciler{}
		is := &apiv1.InferService{
			Spec: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{{Name: "role1"}},
			},
		}
		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}

		convey.Convey("validation success", func() {
			err := reconciler.validate(context.Background(), is, req)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("validation failed", func() {
			is.Spec.Roles = []apiv1.InstanceSetSpec{{Name: ""}}
			err := reconciler.validate(context.Background(), is, req)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestListInstanceSets(t *testing.T) {
	convey.Convey("TestListInstanceSets", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{client: fakeClient}

		is := &apiv1.InferService{
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

			_, _, err := reconciler.listInstanceSets(context.Background(), is)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("success", func() {
			ist := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-role1",
					Namespace: "default",
					Labels: map[string]string{
						common.InferServiceNameLabelKey: "test",
					},
				},
			}
			err := fakeClient.Create(context.Background(), ist)
			convey.So(err, convey.ShouldBeNil)

			list, _, err := reconciler.listInstanceSets(context.Background(), is)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(list.Items), convey.ShouldEqual, 1)
		})
	})
}

func TestCalculateInstanceSetOperations(t *testing.T) {
	convey.Convey("TestCalculateInstanceSetOperations", t, func() {
		reconciler := &InferServiceReconciler{}

		is := &apiv1.InferService{
			Spec: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{
					{Name: "role1"},
					{Name: "role2"},
				},
			},
		}

		convey.Convey("create new", func() {
			existedMap := make(map[string]*apiv1.InstanceSet)
			toCreate, toUpdate, toDelete := reconciler.calculateInstanceSetOperations(is, existedMap)
			convey.So(len(toCreate), convey.ShouldEqual, int2)
			convey.So(len(toUpdate), convey.ShouldEqual, 0)
			convey.So(len(toDelete), convey.ShouldEqual, 0)
		})

		convey.Convey("update existing", func() {
			existedMap := map[string]*apiv1.InstanceSet{
				"role1": {
					ObjectMeta: metav1.ObjectMeta{Name: "test-role1"},
					Spec:       apiv1.InstanceSetSpec{Name: "role1"},
				},
			}
			toCreate, toUpdate, toDelete := reconciler.calculateInstanceSetOperations(is, existedMap)
			convey.So(len(toCreate), convey.ShouldEqual, 1)
			convey.So(len(toUpdate), convey.ShouldEqual, 0)
			convey.So(len(toDelete), convey.ShouldEqual, 0)
		})

		convey.Convey("delete obsolete", func() {
			existedMap := map[string]*apiv1.InstanceSet{
				"role3": {
					ObjectMeta: metav1.ObjectMeta{Name: "test-role3"},
					Spec:       apiv1.InstanceSetSpec{Name: "role3"},
				},
			}
			toCreate, toUpdate, toDelete := reconciler.calculateInstanceSetOperations(is, existedMap)
			convey.So(len(toCreate), convey.ShouldEqual, int2)
			convey.So(len(toUpdate), convey.ShouldEqual, 0)
			convey.So(len(toDelete), convey.ShouldEqual, 1)
		})
	})
}

func TestNewInstanceSet(t *testing.T) {
	convey.Convey("TestNewInstanceSet", t, func() {
		reconciler := &InferServiceReconciler{}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}
		role := apiv1.InstanceSetSpec{
			Name: "role1",
			WorkloadObjectMeta: apiv1.ObjectMeta{
				Labels:      map[string]string{"key1": "value1"},
				Annotations: map[string]string{"key2": "value2"},
			},
		}

		result := reconciler.newInstanceSet(is, role)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.Name, convey.ShouldEqual, "test-role1")
		convey.So(result.Namespace, convey.ShouldEqual, "default")
		convey.So(result.Labels[common.InferServiceNameLabelKey], convey.ShouldEqual, "test")
		convey.So(result.Labels[common.InstanceSetNameLabelKey], convey.ShouldEqual, "role1")
		convey.So(result.Labels["key1"], convey.ShouldEqual, "value1")
		convey.So(result.Annotations["key2"], convey.ShouldEqual, "value2")
	})
}

// targetSuperPodJobCase is a case for TestIsTargetSuperPodJob.
type targetSuperPodJobCase struct {
	name        string
	annotations map[string]string
	expected    bool
}

// getTargetSuperPodJobCases returns cases covering nil/empty/missing-key, known
// target policies, custom policies with the super-pod suffix, and non-target values.
func getTargetSuperPodJobCases() []targetSuperPodJobCase {
	return []targetSuperPodJobCase{
		{name: "nil annotations", annotations: nil, expected: false},
		{name: "empty annotations", annotations: map[string]string{}, expected: false},
		{name: "no schedule_policy key",
			annotations: map[string]string{"other": "value"}, expected: false},
		{name: "policy chip8-node8-sp",
			annotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: ascendapi.Chip8Node8Sp},
			expected:    true},
		{name: "policy chip8-node8-ra64-sp",
			annotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: ascendapi.Chip8Node8Ra64Sp},
			expected:    true},
		{name: "policy chip2-node16-sp",
			annotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: ascendapi.Chip2Node16Sp},
			expected:    true},
		{name: "policy chip2-node8-sp",
			annotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: ascendapi.Chip2Node8Sp},
			expected:    true},
		{name: "custom policy with -sp suffix",
			annotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: "chip9-node8-sp"},
			expected:    true},
		{name: "policy exactly -sp",
			annotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: "-sp"}, expected: false},
		{name: "policy sp without dash",
			annotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: "sp"}, expected: false},
		{name: "non-target policy chip2-node8",
			annotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: "chip2-node8"}, expected: false},
	}
}

func TestIsTargetSuperPodJob(t *testing.T) {
	convey.Convey("TestIsTargetSuperPodJob", t, func() {
		cases := getTargetSuperPodJobCases()
		for i := range cases {
			tc := cases[i]
			convey.Convey(tc.name, func() {
				convey.So(isTargetSuperPodJob(tc.annotations), convey.ShouldEqual, tc.expected)
			})
		}
	})
}

// inferServiceIDInjectionCase is a case for TestNewInstanceSetInferServiceIDInjection.
type inferServiceIDInjectionCase struct {
	name                   string
	isName                 string
	uid                    types.UID
	roleLabels             map[string]string
	annotations            map[string]string
	podTemplateLabels      map[string]string
	podTemplateAnnotations map[string]string
	expectedID             string // empty means inferserviceid must not be injected
}

// getInferServiceIDInjectionCases returns cases covering uid injection for target
// policies, no-injection for non-target/missing annotation, and user-provided value
// precedence.
func getInferServiceIDInjectionCases() []inferServiceIDInjectionCase {
	return []inferServiceIDInjectionCase{
		{name: "chip8-node8-sp injects inferservice uid", isName: "svc-0",
			uid:         "11111111-2222-3333-4444-555555555555",
			annotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: ascendapi.Chip8Node8Sp},
			expectedID:  "11111111-2222-3333-4444-555555555555"},
		{name: "non-target policy does not inject", isName: "svc-0",
			uid:         "11111111-2222-3333-4444-555555555555",
			annotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: "chip2-node8"},
			expectedID:  ""},
		{name: "no annotation does not inject", isName: "svc-0",
			uid:         "11111111-2222-3333-4444-555555555555",
			annotations: nil, expectedID: ""},
		{name: "user-provided inferserviceid takes precedence", isName: "svc-0",
			uid:         "11111111-2222-3333-4444-555555555555",
			roleLabels:  map[string]string{common.InferServiceIDLabelKey: "custom-xxx"},
			annotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: ascendapi.Chip8Node8Sp},
			expectedID:  "custom-xxx"},
		{name: "schedule policy in pod template injects inferservice uid", isName: "svc-2",
			uid:                    "22222222-3333-4444-5555-666666666666",
			podTemplateAnnotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: ascendapi.Chip2Node16Sp},
			expectedID:             "22222222-3333-4444-5555-666666666666"},
		{name: "workload meta policy wins over pod template policy", isName: "svc-3",
			uid:                    "33333333-4444-5555-6666-777777777777",
			annotations:            map[string]string{ascendapi.SchedulePolicyAnnoKey: ascendapi.Chip2Node8Sp},
			podTemplateAnnotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: ascendapi.Chip8Node8Sp},
			expectedID:             "33333333-4444-5555-6666-777777777777"},
		{name: "non-target policy in pod template does not inject", isName: "svc-4",
			uid:                    "44444444-5555-6666-7777-888888888888",
			podTemplateAnnotations: map[string]string{ascendapi.SchedulePolicyAnnoKey: "chip2-node8"},
			expectedID:             ""},
		{name: "empty workload spec does not inject", isName: "svc-5",
			uid:        "55555555-6666-7777-8888-999999999999",
			expectedID: ""},
		{name: "pod template inferserviceid label skips auto injection", isName: "svc-6",
			uid:               "66666666-7777-8888-9999-aaaaaaaaaaaa",
			annotations:       map[string]string{ascendapi.SchedulePolicyAnnoKey: ascendapi.Chip8Node8Sp},
			podTemplateLabels: map[string]string{common.InferServiceIDLabelKey: "custom-pod"},
			expectedID:        ""},
	}
}

// marshalPodTemplateMeta marshals pod template labels/annotations into a
// workload spec RawExtension with the same shape as roles[].spec.
func marshalPodTemplateMeta(labels, annotations map[string]string) runtime.RawExtension {
	workloadSpec := struct {
		Template struct {
			Metadata metav1.ObjectMeta `json:"metadata"`
		} `json:"template"`
	}{}
	workloadSpec.Template.Metadata.Labels = labels
	workloadSpec.Template.Metadata.Annotations = annotations
	raw, _ := json.Marshal(&workloadSpec)
	return runtime.RawExtension{Raw: raw}
}

func TestNewInstanceSetInferServiceIDInjection(t *testing.T) {
	convey.Convey("TestNewInstanceSetInferServiceIDInjection", t, func() {
		reconciler := &InferServiceReconciler{}
		cases := getInferServiceIDInjectionCases()
		for i := range cases {
			tc := cases[i]
			convey.Convey(tc.name, func() {
				is := &apiv1.InferService{
					ObjectMeta: metav1.ObjectMeta{Name: tc.isName, Namespace: "default", UID: tc.uid},
				}
				role := apiv1.InstanceSetSpec{
					Name: "role1",
					WorkloadObjectMeta: apiv1.ObjectMeta{
						Labels:      tc.roleLabels,
						Annotations: tc.annotations,
					},
					InstanceSpec: marshalPodTemplateMeta(tc.podTemplateLabels, tc.podTemplateAnnotations),
				}
				result := reconciler.newInstanceSet(is, role)
				convey.So(result, convey.ShouldNotBeNil)
				if tc.expectedID == "" {
					_, exist := result.Labels[common.InferServiceIDLabelKey]
					convey.So(exist, convey.ShouldBeFalse)
				} else {
					convey.So(result.Labels[common.InferServiceIDLabelKey],
						convey.ShouldEqual, tc.expectedID)
				}
			})
		}
	})
}

func TestDeleteInstanceSets(t *testing.T) {
	convey.Convey("TestDeleteInstanceSets", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{client: fakeClient}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}

		convey.Convey("nil InferService", func() {
			err := reconciler.deleteInstanceSets(context.Background(), nil, nil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("delete error", func() {
			patch := gomonkey.ApplyMethodFunc(fakeClient, "Delete",
				func(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
					return errors.New("delete error")
				})
			defer patch.Reset()

			ist := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			err := reconciler.deleteInstanceSets(context.Background(), is, []*apiv1.InstanceSet{ist})
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("not found", func() {
			ist := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			err := reconciler.deleteInstanceSets(context.Background(), is, []*apiv1.InstanceSet{ist})
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("success", func() {
			ist := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			err := fakeClient.Create(context.Background(), ist)
			convey.So(err, convey.ShouldBeNil)

			err = reconciler.deleteInstanceSets(context.Background(), is, []*apiv1.InstanceSet{ist})
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUpdateExistInstanceSets(t *testing.T) {
	convey.Convey("TestUpdateExistInstanceSets", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{client: fakeClient}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}

		convey.Convey("nil InferService", func() {
			err := reconciler.updateExistInstanceSets(context.Background(), nil, nil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("get error", func() {
			patch := gomonkey.ApplyMethodFunc(fakeClient, "Get",
				func(_ context.Context, _ types.NamespacedName, _ client.Object, _ ...client.GetOption) error {
					return errors.New("get error")
				})
			defer patch.Reset()

			ist := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			err := reconciler.updateExistInstanceSets(context.Background(), is, []*apiv1.InstanceSet{ist})
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("update error", func() {
			patch := gomonkey.ApplyMethodFunc(fakeClient, "Update",
				func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
					return errors.New("update error")
				})
			defer patch.Reset()

			ist := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			err := reconciler.updateExistInstanceSets(context.Background(), is, []*apiv1.InstanceSet{ist})
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("success without HPA", func() {
			ist := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec:       apiv1.InstanceSetSpec{Name: "role1"},
			}
			err := fakeClient.Create(context.Background(), ist)
			convey.So(err, convey.ShouldBeNil)

			err = reconciler.updateExistInstanceSets(context.Background(), is, []*apiv1.InstanceSet{ist})
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("HPA managed preserves replicas", func() {
			existingReplicas := int32(5)
			ist := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{Name: "hpa-test", Namespace: "default"},
				Spec: apiv1.InstanceSetSpec{
					Name:     "role1",
					Replicas: &existingReplicas,
					ScalingPolicy: &apiv1.ScalingPolicy{
						Type: common.ScalingPolicyTypeHPA,
					},
				},
			}
			err := fakeClient.Create(context.Background(), ist)
			convey.So(err, convey.ShouldBeNil)

			desiredReplicas := int32(3)
			desiredIst := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{Name: "hpa-test", Namespace: "default"},
				Spec: apiv1.InstanceSetSpec{
					Name:     "role1",
					Replicas: &desiredReplicas,
				},
			}
			err = reconciler.updateExistInstanceSets(context.Background(), is, []*apiv1.InstanceSet{desiredIst})
			convey.So(err, convey.ShouldBeNil)

			updatedIst := &apiv1.InstanceSet{}
			err = fakeClient.Get(context.Background(), types.NamespacedName{
				Name: "hpa-test", Namespace: "default",
			}, updatedIst)
			convey.So(err, convey.ShouldBeNil)
			convey.So(*updatedIst.Spec.Replicas, convey.ShouldEqual, 5)
		})

		convey.Convey("non-HPA managed overwrites replicas", func() {
			existingReplicas := int32(5)
			ist := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{Name: "non-hpa-test", Namespace: "default"},
				Spec: apiv1.InstanceSetSpec{
					Name:     "role1",
					Replicas: &existingReplicas,
				},
			}
			err := fakeClient.Create(context.Background(), ist)
			convey.So(err, convey.ShouldBeNil)

			desiredReplicas := int32(3)
			desiredIst := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{Name: "non-hpa-test", Namespace: "default"},
				Spec: apiv1.InstanceSetSpec{
					Name:     "role1",
					Replicas: &desiredReplicas,
				},
			}
			err = reconciler.updateExistInstanceSets(context.Background(), is, []*apiv1.InstanceSet{desiredIst})
			convey.So(err, convey.ShouldBeNil)

			updatedIst := &apiv1.InstanceSet{}
			err = fakeClient.Get(context.Background(), types.NamespacedName{
				Name: "non-hpa-test", Namespace: "default",
			}, updatedIst)
			convey.So(err, convey.ShouldBeNil)
			convey.So(*updatedIst.Spec.Replicas, convey.ShouldEqual, 3)
		})
	})
}

func TestCreateInstanceSetsNilInferService(t *testing.T) {
	convey.Convey("TestCreateInstanceSets nil InferService", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client: fakeClient,
			scheme: scheme,
		}

		err := reconciler.createInstanceSets(context.Background(), nil, nil)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestCreateInstanceSetsControllerReferenceError(t *testing.T) {
	convey.Convey("TestCreateInstanceSets set controller reference error", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client: fakeClient,
			scheme: scheme,
		}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}

		patch := gomonkey.ApplyFuncReturn(controllerutil.SetControllerReference, errors.New("reference error"))
		defer patch.Reset()

		ist := &apiv1.InstanceSet{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		}
		err := reconciler.createInstanceSets(context.Background(), is, []*apiv1.InstanceSet{ist})
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestCreateInstanceSetsCreateError(t *testing.T) {
	convey.Convey("TestCreateInstanceSets create error", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client: fakeClient,
			scheme: scheme,
		}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}

		patch := gomonkey.ApplyMethodFunc(fakeClient, "Create",
			func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
				return errors.New("create error")
			})
		defer patch.Reset()

		ist := &apiv1.InstanceSet{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		}
		err := reconciler.createInstanceSets(context.Background(), is, []*apiv1.InstanceSet{ist})
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestCreateInstanceSetsAlreadyExists(t *testing.T) {
	convey.Convey("TestCreateInstanceSets already exists", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client: fakeClient,
			scheme: scheme,
		}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}

		ist := &apiv1.InstanceSet{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		}
		err := fakeClient.Create(context.Background(), ist)
		convey.So(err, convey.ShouldBeNil)

		ist2 := &apiv1.InstanceSet{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		}
		err = reconciler.createInstanceSets(context.Background(), is, []*apiv1.InstanceSet{ist2})
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestCreateInstanceSetsSuccess(t *testing.T) {
	convey.Convey("TestCreateInstanceSets success", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client: fakeClient,
			scheme: scheme,
		}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}

		ist := &apiv1.InstanceSet{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		}
		err := reconciler.createInstanceSets(context.Background(), is, []*apiv1.InstanceSet{ist})
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestManageInstanceSetsDeleteError(t *testing.T) {
	convey.Convey("TestManageInstanceSets delete error", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client: fakeClient,
			scheme: scheme,
		}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}

		patch := gomonkey.ApplyPrivateMethod(reconciler, "deleteInstanceSets",
			func(_ context.Context, _ *apiv1.InferService, _ []*apiv1.InstanceSet) error {
				return errors.New("delete error")
			})
		defer patch.Reset()

		err := reconciler.manageInstanceSets(context.Background(), is, nil, nil, nil)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestManageInstanceSetsUpdateError(t *testing.T) {
	convey.Convey("TestManageInstanceSets update error", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client: fakeClient,
			scheme: scheme,
		}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}

		patch := gomonkey.ApplyPrivateMethod(reconciler, "updateExistInstanceSets",
			func(_ context.Context, _ *apiv1.InferService, _ []*apiv1.InstanceSet) error {
				return errors.New("update error")
			})
		defer patch.Reset()

		err := reconciler.manageInstanceSets(context.Background(), is, nil, nil, nil)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestManageInstanceSetsCreateError(t *testing.T) {
	convey.Convey("TestManageInstanceSets create error", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client: fakeClient,
			scheme: scheme,
		}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}

		patch := gomonkey.ApplyPrivateMethod(reconciler, "createInstanceSets",
			func(_ context.Context, _ *apiv1.InferService, _ []*apiv1.InstanceSet) error {
				return errors.New("create error")
			})
		defer patch.Reset()

		err := reconciler.manageInstanceSets(context.Background(), is, nil, nil, nil)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestManageInstanceSetsSuccess(t *testing.T) {
	convey.Convey("TestManageInstanceSets success", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client: fakeClient,
			scheme: scheme,
		}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}

		err := reconciler.manageInstanceSets(context.Background(), is, nil, nil, nil)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestCalculateStatus(t *testing.T) {
	convey.Convey("TestCalculateStatus", t, func() {
		reconciler := &InferServiceReconciler{}

		is := &apiv1.InferService{
			Spec: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{
					{Name: "role1"},
					{Name: "role2"},
				},
			},
		}

		convey.Convey("calculate status", func() {
			instanceSetList := &apiv1.InstanceSetList{
				Items: []apiv1.InstanceSet{
					{
						Status: apiv1.InstanceSetStatus{
							Conditions: []metav1.Condition{
								{Type: string(common.InstanceSetReady), Status: metav1.ConditionTrue},
							},
						},
					},
				},
			}
			result := reconciler.calculateStatus(is, instanceSetList)
			convey.So(result.Replicas, convey.ShouldEqual, int2)
			convey.So(result.ReadyReplicas, convey.ShouldEqual, 1)
			convey.So(result.ObservedGeneration, convey.ShouldEqual, 0)
		})
	})
}

func TestUpdateStatus(t *testing.T) {
	convey.Convey("TestUpdateStatus", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client: fakeClient,
			scheme: scheme,
		}

		convey.Convey("nil InferService", func() {
			err := reconciler.updateStatus(context.Background(), nil, nil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("status unchanged", func() {
			is := &apiv1.InferService{
				Status: apiv1.InferServiceStatus{
					Replicas:      1,
					ReadyReplicas: 1,
				},
			}
			instanceSetList := &apiv1.InstanceSetList{}
			patch := gomonkey.ApplyPrivateMethod(reconciler, "updateStatusWithRetry",
				func(_ context.Context, _ *apiv1.InferService, _ apiv1.InferServiceStatus) error {
					return nil
				})
			defer patch.Reset()
			err := reconciler.updateStatus(context.Background(), is, instanceSetList)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("status changed", func() {
			is := &apiv1.InferService{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Status:     apiv1.InferServiceStatus{},
			}
			instanceSetList := &apiv1.InstanceSetList{}
			patch := gomonkey.ApplyPrivateMethod(reconciler, "updateStatusWithRetry",
				func(_ context.Context, _ *apiv1.InferService, _ apiv1.InferServiceStatus) error {
					return nil
				})
			defer patch.Reset()

			err := reconciler.updateStatus(context.Background(), is, instanceSetList)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestReconcileGetError(t *testing.T) {
	convey.Convey("TestReconcile get error", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client:   fakeClient,
			scheme:   scheme,
			recorder: record.NewFakeRecorder(int10),
		}

		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}

		patch := gomonkey.ApplyPrivateMethod(reconciler, "getInferService",
			func(_ context.Context, _ ctrl.Request) (*apiv1.InferService, error) {
				return nil, errors.New("get error")
			})
		defer patch.Reset()

		_, err := reconciler.Reconcile(context.Background(), req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestReconcileNilInferService(t *testing.T) {
	convey.Convey("TestReconcile nil InferService", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client:   fakeClient,
			scheme:   scheme,
			recorder: record.NewFakeRecorder(int10),
		}

		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}

		patch := gomonkey.ApplyPrivateMethod(reconciler, "getInferService",
			func(_ context.Context, _ ctrl.Request) (*apiv1.InferService, error) {
				return nil, nil
			})
		defer patch.Reset()

		result, err := reconciler.Reconcile(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.RequeueAfter, convey.ShouldEqual, 0)
	})
}

func TestReconcileValidateError(t *testing.T) {
	convey.Convey("TestReconcile validate error", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client:   fakeClient,
			scheme:   scheme,
			recorder: record.NewFakeRecorder(int10),
		}

		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}
		patch := gomonkey.ApplyPrivateMethod(reconciler, "getInferService",
			func(_ context.Context, _ ctrl.Request) (*apiv1.InferService, error) {
				return is, nil
			})
		defer patch.Reset()

		patch2 := gomonkey.ApplyPrivateMethod(reconciler, "validate",
			func(_ context.Context, _ *apiv1.InferService, _ ctrl.Request) error {
				return errors.New("validate error")
			})
		defer patch2.Reset()

		_, err := reconciler.Reconcile(context.Background(), req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestReconcileListError(t *testing.T) {
	convey.Convey("TestReconcile list error", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client:   fakeClient,
			scheme:   scheme,
			recorder: record.NewFakeRecorder(int10),
		}

		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{{Name: "role1"}},
			},
		}
		patch := gomonkey.ApplyPrivateMethod(reconciler, "getInferService",
			func(_ context.Context, _ ctrl.Request) (*apiv1.InferService, error) {
				return is, nil
			})
		defer patch.Reset()

		patch2 := gomonkey.ApplyPrivateMethod(reconciler, "listInstanceSets",
			func(_ context.Context, _ *apiv1.InferService) (*apiv1.InstanceSetList, interface{}, error) {
				return nil, nil, errors.New("list error")
			})
		defer patch2.Reset()

		_, err := reconciler.Reconcile(context.Background(), req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestReconcileSuccess(t *testing.T) {
	convey.Convey("TestReconcile success", t, func() {
		scheme := runtime.NewScheme()
		apiv1.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler := &InferServiceReconciler{
			client:   fakeClient,
			scheme:   scheme,
			recorder: record.NewFakeRecorder(int10),
		}

		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}

		is := &apiv1.InferService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: apiv1.InferServiceSpec{
				Roles: []apiv1.InstanceSetSpec{{Name: "role1"}},
			},
		}
		patch := gomonkey.ApplyPrivateMethod(reconciler, "getInferService",
			func(_ context.Context, _ ctrl.Request) (*apiv1.InferService, error) {
				return is, nil
			})
		defer patch.Reset()

		patch2 := gomonkey.ApplyPrivateMethod(reconciler, "listInstanceSets",
			func(_ context.Context, _ *apiv1.InferService) (*apiv1.InstanceSetList, interface{}, error) {
				return &apiv1.InstanceSetList{}, nil, nil
			})
		defer patch2.Reset()

		patch3 := gomonkey.ApplyPrivateMethod(reconciler, "manageInstanceSets",
			func(_ context.Context, _ *apiv1.InferService, _ []*apiv1.InstanceSet, _ []*apiv1.InstanceSet, _ []*apiv1.InstanceSet) error {
				return nil
			})
		defer patch3.Reset()

		patch4 := gomonkey.ApplyPrivateMethod(reconciler, "updateInferServiceStatus",
			func(_ context.Context, _ *apiv1.InferService, _ interface{}) error {
				return nil
			})
		defer patch4.Reset()

		result, err := reconciler.Reconcile(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.RequeueAfter, convey.ShouldEqual, 0)
	})
}

func TestFilterInstanceSetEventsCreateFunc(t *testing.T) {
	convey.Convey("TestFilterInstanceSetEvents CreateFunc", t, func() {
		filter := filterInstanceSetEvents()
		result := filter.CreateFunc(event.CreateEvent{})
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestFilterInstanceSetEventsUpdateFunc(t *testing.T) {
	convey.Convey("TestFilterInstanceSetEvents UpdateFunc", t, func() {
		filter := filterInstanceSetEvents()

		convey.Convey("wrong type", func() {
			result := filter.UpdateFunc(event.UpdateEvent{})
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("no change", func() {
			ist := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Labels:    map[string]string{common.InferServiceNameLabelKey: "test"},
				},
			}
			result := filter.UpdateFunc(event.UpdateEvent{
				ObjectOld: ist,
				ObjectNew: ist,
			})
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("with change", func() {
			oldIst := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Labels:    map[string]string{common.InferServiceNameLabelKey: "test"},
				},
				Spec: apiv1.InstanceSetSpec{Name: "role1"},
			}
			newIst := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Labels:    map[string]string{common.InferServiceNameLabelKey: "test"},
				},
				Spec: apiv1.InstanceSetSpec{Name: "role2"},
			}
			result := filter.UpdateFunc(event.UpdateEvent{
				ObjectOld: oldIst,
				ObjectNew: newIst,
			})
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

func TestFilterInstanceSetEventsDeleteFunc(t *testing.T) {
	convey.Convey("TestFilterInstanceSetEvents DeleteFunc", t, func() {
		filter := filterInstanceSetEvents()

		convey.Convey("wrong type", func() {
			result := filter.DeleteFunc(event.DeleteEvent{})
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("with label", func() {
			ist := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Labels:    map[string]string{common.InferServiceNameLabelKey: "test"},
				},
			}
			result := filter.DeleteFunc(event.DeleteEvent{Object: ist})
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("without label", func() {
			ist := &apiv1.InstanceSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			}
			result := filter.DeleteFunc(event.DeleteEvent{Object: ist})
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestFilterInstanceSetEventsGenericFunc(t *testing.T) {
	convey.Convey("TestFilterInstanceSetEvents GenericFunc", t, func() {
		filter := filterInstanceSetEvents()
		result := filter.GenericFunc(event.GenericEvent{})
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestIsManagedByHPA(t *testing.T) {
	convey.Convey("TestIsManagedByHPA", t, func() {
		reconciler := &InferServiceReconciler{}

		convey.Convey("nil ScalingPolicy", func() {
			ist := &apiv1.InstanceSet{
				Spec: apiv1.InstanceSetSpec{
					ScalingPolicy: nil,
				},
			}
			convey.So(reconciler.isManagedByScalingController(ist), convey.ShouldBeFalse)
		})

		convey.Convey("HPA type", func() {
			ist := &apiv1.InstanceSet{
				Spec: apiv1.InstanceSetSpec{
					ScalingPolicy: &apiv1.ScalingPolicy{
						Type: common.ScalingPolicyTypeHPA,
					},
				},
			}
			convey.So(reconciler.isManagedByScalingController(ist), convey.ShouldBeTrue)
		})

		convey.Convey("non-HPA type", func() {
			ist := &apiv1.InstanceSet{
				Spec: apiv1.InstanceSetSpec{
					ScalingPolicy: &apiv1.ScalingPolicy{
						Type: "Other",
					},
				},
			}
			convey.So(reconciler.isManagedByScalingController(ist), convey.ShouldBeFalse)
		})

		convey.Convey("empty type", func() {
			ist := &apiv1.InstanceSet{
				Spec: apiv1.InstanceSetSpec{
					ScalingPolicy: &apiv1.ScalingPolicy{
						Type: "",
					},
				},
			}
			convey.So(reconciler.isManagedByScalingController(ist), convey.ShouldBeFalse)
		})
	})
}

func TestSpecChangedExcludingReplicas(t *testing.T) {
	convey.Convey("TestSpecChangedExcludingReplicas", t, func() {
		reconciler := &InferServiceReconciler{}

		convey.Convey("identical spec", func() {
			replicas := int32(3)
			current := apiv1.InstanceSetSpec{
				Name:     "role1",
				Replicas: &replicas,
			}
			desired := apiv1.InstanceSetSpec{
				Name:     "role1",
				Replicas: &replicas,
			}
			convey.So(reconciler.specChangedExcludingReplicas(current, desired), convey.ShouldBeFalse)
		})

		convey.Convey("only replicas differ", func() {
			currentReplicas := int32(3)
			desiredReplicas := int32(5)
			current := apiv1.InstanceSetSpec{
				Name:     "role1",
				Replicas: &currentReplicas,
			}
			desired := apiv1.InstanceSetSpec{
				Name:     "role1",
				Replicas: &desiredReplicas,
			}
			convey.So(reconciler.specChangedExcludingReplicas(current, desired), convey.ShouldBeFalse)
		})

		convey.Convey("name differs", func() {
			replicas := int32(3)
			current := apiv1.InstanceSetSpec{
				Name:     "role1",
				Replicas: &replicas,
			}
			desired := apiv1.InstanceSetSpec{
				Name:     "role2",
				Replicas: &replicas,
			}
			convey.So(reconciler.specChangedExcludingReplicas(current, desired), convey.ShouldBeTrue)
		})

		convey.Convey("name and replicas differ", func() {
			currentReplicas := int32(3)
			desiredReplicas := int32(5)
			current := apiv1.InstanceSetSpec{
				Name:     "role1",
				Replicas: &currentReplicas,
			}
			desired := apiv1.InstanceSetSpec{
				Name:     "role2",
				Replicas: &desiredReplicas,
			}
			convey.So(reconciler.specChangedExcludingReplicas(current, desired), convey.ShouldBeTrue)
		})

		convey.Convey("scaling policy differs", func() {
			replicas := int32(3)
			current := apiv1.InstanceSetSpec{
				Name:     "role1",
				Replicas: &replicas,
			}
			desired := apiv1.InstanceSetSpec{
				Name:     "role1",
				Replicas: &replicas,
				ScalingPolicy: &apiv1.ScalingPolicy{
					Type: common.ScalingPolicyTypeHPA,
				},
			}
			convey.So(reconciler.specChangedExcludingReplicas(current, desired), convey.ShouldBeTrue)
		})
	})
}

func TestInstanceSetUpdated(t *testing.T) {
	convey.Convey("TestInstanceSetUpdated", t, func() {
		reconciler := &InferServiceReconciler{}

		convey.Convey("nil instanceSet", func() {
			role := apiv1.InstanceSetSpec{Name: "role1"}
			convey.So(reconciler.instanceSetUpdated(nil, role), convey.ShouldBeFalse)
		})

		convey.Convey("spec unchanged", func() {
			replicas := int32(3)
			ist := &apiv1.InstanceSet{
				Spec: apiv1.InstanceSetSpec{
					Name:     "role1",
					Replicas: &replicas,
				},
			}
			role := apiv1.InstanceSetSpec{
				Name:     "role1",
				Replicas: &replicas,
			}
			convey.So(reconciler.instanceSetUpdated(ist, role), convey.ShouldBeFalse)
		})

		convey.Convey("spec changed", func() {
			replicas := int32(3)
			ist := &apiv1.InstanceSet{
				Spec: apiv1.InstanceSetSpec{
					Name:     "role1",
					Replicas: &replicas,
				},
			}
			role := apiv1.InstanceSetSpec{
				Name:     "role2",
				Replicas: &replicas,
			}
			convey.So(reconciler.instanceSetUpdated(ist, role), convey.ShouldBeTrue)
		})

		convey.Convey("HPA managed only replicas changed", func() {
			currentReplicas := int32(5)
			desiredReplicas := int32(3)
			ist := &apiv1.InstanceSet{
				Spec: apiv1.InstanceSetSpec{
					Name:     "role1",
					Replicas: &currentReplicas,
					ScalingPolicy: &apiv1.ScalingPolicy{
						Type: common.ScalingPolicyTypeHPA,
					},
				},
			}
			role := apiv1.InstanceSetSpec{
				Name:     "role1",
				Replicas: &desiredReplicas,
				ScalingPolicy: &apiv1.ScalingPolicy{
					Type: common.ScalingPolicyTypeHPA,
				},
			}
			convey.So(reconciler.instanceSetUpdated(ist, role), convey.ShouldBeFalse)
		})

		convey.Convey("HPA managed name and replicas changed", func() {
			currentReplicas := int32(5)
			desiredReplicas := int32(3)
			ist := &apiv1.InstanceSet{
				Spec: apiv1.InstanceSetSpec{
					Name:     "role1",
					Replicas: &currentReplicas,
					ScalingPolicy: &apiv1.ScalingPolicy{
						Type: common.ScalingPolicyTypeHPA,
					},
				},
			}
			role := apiv1.InstanceSetSpec{
				Name:     "role2",
				Replicas: &desiredReplicas,
			}
			convey.So(reconciler.instanceSetUpdated(ist, role), convey.ShouldBeTrue)
		})

		convey.Convey("non-HPA managed replicas changed", func() {
			currentReplicas := int32(5)
			desiredReplicas := int32(3)
			ist := &apiv1.InstanceSet{
				Spec: apiv1.InstanceSetSpec{
					Name:     "role1",
					Replicas: &currentReplicas,
				},
			}
			role := apiv1.InstanceSetSpec{
				Name:     "role1",
				Replicas: &desiredReplicas,
			}
			convey.So(reconciler.instanceSetUpdated(ist, role), convey.ShouldBeTrue)
		})
	})
}
