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

package snapshot

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"ascend-common/common-utils/hwlog"
	v1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func newPodSnapshotTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = v1.AddToScheme(scheme)
	return scheme
}

func newPodSnapshotFakeClientBuilder() *fake.ClientBuilder {
	return fake.NewClientBuilder().WithScheme(newPodSnapshotTestScheme())
}

func createTestPodForSnapshotReadiness(name, namespace string, labels, annotations map[string]string, readinessGates []corev1.PodReadinessGate) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			ReadinessGates: readinessGates,
		},
	}
}

func createTestPodWithHostSnapshotVolume(name, namespace string, labels, annotations map[string]string, readinessGates []corev1.PodReadinessGate) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			ReadinessGates: readinessGates,
			Containers: []corev1.Container{
				{
					Name: "main",
					Env: []corev1.EnvVar{
						{
							Name:  common.HostSnapshotDirPathEnvKey,
							Value: "/data/host-snapshot",
						},
					},
				},
			},
		},
	}
}

func TestPodSnapshotReadinessReconcile(t *testing.T) {
	convey.Convey("Test PodSnapshotReadinessReconciler Reconcile", t, func() {
		convey.Convey("Should return not found when pod does not exist", func() {
			fakeClient := newPodSnapshotFakeClientBuilder().Build()
			reconciler := &PodSnapshotReconciler{Client: fakeClient}

			req := reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: "default", Name: "non-existent-pod",
			}}

			result, err := reconciler.Reconcile(context.Background(), req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})

		convey.Convey("Should not update when pod has no snapshot readiness gate", func() {
			pod := createTestPodWithHostSnapshotVolume("test-pod", "default", map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-instance",
			}, nil, nil)

			fakeClient := newPodSnapshotFakeClientBuilder().WithObjects(pod).Build()
			reconciler := &PodSnapshotReconciler{Client: fakeClient}

			req := reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: "default", Name: "test-pod",
			}}

			result, err := reconciler.Reconcile(context.Background(), req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})
	})
}

func TestSetPodSnapshotModeAnnotation(t *testing.T) {
	convey.Convey("Test setPodSnapshotModeAnnotation", t, func() {
		convey.Convey("Should return error when host snapshot path is empty", func() {
			pod := createTestPodForSnapshotReadiness("test-pod", "default", nil, nil, nil)
			fakeClient := newPodSnapshotFakeClientBuilder().WithObjects(pod).Build()
			reconciler := &PodSnapshotReconciler{Client: fakeClient}

			err := reconciler.setPodSnapshotModeAnnotation(context.Background(), pod)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should set save mode when snapshot status not exists", func() {
			pod := createTestPodWithHostSnapshotVolume("test-pod", "default", map[string]string{
				common.InstanceIndexLabelKey: "0",
			}, nil, nil)
			fakeClient := newPodSnapshotFakeClientBuilder().WithObjects(pod).Build()
			reconciler := &PodSnapshotReconciler{Client: fakeClient}

			err := reconciler.setPodSnapshotModeAnnotation(context.Background(), pod)
			convey.So(err, convey.ShouldBeNil)
			convey.So(pod.Annotations[common.SnapshotModeAnnotationKey], convey.ShouldEqual, common.SnapshotSaveMode)
		})

		convey.Convey("Should not update when mode already set", func() {
			pod := createTestPodWithHostSnapshotVolume("test-pod", "default", map[string]string{
				common.InstanceIndexLabelKey: "0",
			}, map[string]string{common.SnapshotModeAnnotationKey: common.SnapshotSaveMode}, nil)
			fakeClient := newPodSnapshotFakeClientBuilder().WithObjects(pod).Build()
			reconciler := &PodSnapshotReconciler{Client: fakeClient}

			err := reconciler.setPodSnapshotModeAnnotation(context.Background(), pod)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should skip when instance index is not 0", func() {
			pod := createTestPodWithHostSnapshotVolume("test-pod", "default", map[string]string{
				common.InstanceIndexLabelKey: "1",
			}, nil, nil)
			fakeClient := newPodSnapshotFakeClientBuilder().WithObjects(pod).Build()
			reconciler := &PodSnapshotReconciler{Client: fakeClient}

			err := reconciler.setPodSnapshotModeAnnotation(context.Background(), pod)
			convey.So(err, convey.ShouldBeNil)
			convey.So(pod.Annotations[common.SnapshotModeAnnotationKey], convey.ShouldEqual, "")
		})
	})
}

func TestPodSnapshotReadinessReconcileAnnotation(t *testing.T) {
	convey.Convey("Test PodSnapshotReadinessReconciler with annotation", t, func() {
		convey.Convey("Should update when annotation finish flag is true", func() {
			pod := createTestPodWithHostSnapshotVolume("test-pod", "default", map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-instance",
			}, map[string]string{
				common.HostSnapshotFlagAnnotationKey: common.TrueBool,
			}, []corev1.PodReadinessGate{
				{ConditionType: common.PodSnapshotReadyConditionType},
			})

			fakeClient := newPodSnapshotFakeClientBuilder().WithObjects(pod).Build()
			reconciler := &PodSnapshotReconciler{Client: fakeClient}

			req := reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: "default", Name: "test-pod",
			}}

			result, err := reconciler.Reconcile(context.Background(), req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})
	})
}

func TestPodSnapshotReadinessReconcileLoadMode(t *testing.T) {
	convey.Convey("Test PodSnapshotReadinessReconciler with load mode", t, func() {
		convey.Convey("Should update when ConfigMap is in load mode", func() {
			pod := createTestPodWithHostSnapshotVolume("test-pod", "default", map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-instance",
			}, nil, []corev1.PodReadinessGate{
				{ConditionType: common.PodSnapshotReadyConditionType},
			})

			fakeClient := newPodSnapshotFakeClientBuilder().WithObjects(pod).Build()
			reconciler := &PodSnapshotReconciler{Client: fakeClient}

			req := reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: "default", Name: "test-pod",
			}}

			result, err := reconciler.Reconcile(context.Background(), req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})
	})
}

func TestPodSnapshotReadinessReconcileSaveMode(t *testing.T) {
	convey.Convey("Test PodSnapshotReadinessReconciler with save mode", t, func() {
		convey.Convey("Should update when save mode and instance index is not 0", func() {
			pod := createTestPodWithHostSnapshotVolume("test-pod", "default", map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-instance",
				common.InstanceIndexLabelKey:    "1",
			}, nil, []corev1.PodReadinessGate{
				{ConditionType: common.PodSnapshotReadyConditionType},
			})

			fakeClient := newPodSnapshotFakeClientBuilder().WithObjects(pod).Build()
			reconciler := &PodSnapshotReconciler{Client: fakeClient}

			req := reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: "default", Name: "test-pod",
			}}

			result, err := reconciler.Reconcile(context.Background(), req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})

		convey.Convey("Should not update when save mode and instance index is 0", func() {
			pod := createTestPodWithHostSnapshotVolume("test-pod", "default", map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-instance",
				common.InstanceIndexLabelKey:    "0",
			}, nil, []corev1.PodReadinessGate{
				{ConditionType: common.PodSnapshotReadyConditionType},
			})

			fakeClient := newPodSnapshotFakeClientBuilder().WithObjects(pod).Build()
			reconciler := &PodSnapshotReconciler{Client: fakeClient}

			req := reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: "default", Name: "test-pod",
			}}

			result, err := reconciler.Reconcile(context.Background(), req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})
	})
}

func TestPodSnapshotReadinessReconcileNoUpdate(t *testing.T) {
	convey.Convey("Test PodSnapshotReadinessReconciler no update scenarios", t, func() {
		convey.Convey("Should not update when ConfigMap is empty", func() {
			pod := createTestPodWithHostSnapshotVolume("test-pod", "default", map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-instance",
			}, nil, []corev1.PodReadinessGate{
				{ConditionType: common.PodSnapshotReadyConditionType},
			})

			fakeClient := newPodSnapshotFakeClientBuilder().WithObjects(pod).Build()
			reconciler := &PodSnapshotReconciler{Client: fakeClient}

			req := reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: "default", Name: "test-pod",
			}}

			result, err := reconciler.Reconcile(context.Background(), req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})

		convey.Convey("Should not update when condition is already true", func() {
			pod := createTestPodWithHostSnapshotVolume("test-pod", "default", map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-instance",
			}, map[string]string{
				common.HostSnapshotFlagAnnotationKey: common.TrueBool,
			}, []corev1.PodReadinessGate{
				{ConditionType: common.PodSnapshotReadyConditionType},
			})
			pod.Status.Conditions = []corev1.PodCondition{
				{Type: common.PodSnapshotReadyConditionType, Status: corev1.ConditionTrue},
			}

			fakeClient := newPodSnapshotFakeClientBuilder().WithObjects(pod).Build()
			reconciler := &PodSnapshotReconciler{Client: fakeClient}

			req := reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: "default", Name: "test-pod",
			}}

			result, err := reconciler.Reconcile(context.Background(), req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})
	})
}

func TestShouldProcessPod(t *testing.T) {
	convey.Convey("Test shouldProcessPod", t, func() {
		convey.Convey("Should return true when all conditions are met", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						common.OperatorNameKey:           common.TrueBool,
						common.InferServiceNameLabelKey:  "test-service",
						common.InstanceSetNameLabelKey:   "test-instance",
						common.InstanceIndexLabelKey:     "0",
						common.ContainerSnapshotLabelKey: common.TrueBool,
					},
				},
				Spec: corev1.PodSpec{
					ReadinessGates: []corev1.PodReadinessGate{
						{ConditionType: common.PodSnapshotReadyConditionType},
					},
				},
			}
			convey.So(shouldProcessPod(pod), convey.ShouldBeTrue)
		})

		convey.Convey("Should return false when operator name is missing", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						common.InferServiceNameLabelKey:  "test-service",
						common.InstanceSetNameLabelKey:   "test-instance",
						common.InstanceIndexLabelKey:     "0",
						common.ContainerSnapshotLabelKey: common.TrueBool,
					},
				},
				Spec: corev1.PodSpec{
					ReadinessGates: []corev1.PodReadinessGate{
						{ConditionType: common.PodSnapshotReadyConditionType},
					},
				},
			}
			convey.So(shouldProcessPod(pod), convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when infer service name is missing", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						common.OperatorNameKey:           common.TrueBool,
						common.InstanceSetNameLabelKey:   "test-instance",
						common.InstanceIndexLabelKey:     "0",
						common.ContainerSnapshotLabelKey: common.TrueBool,
					},
				},
				Spec: corev1.PodSpec{
					ReadinessGates: []corev1.PodReadinessGate{
						{ConditionType: common.PodSnapshotReadyConditionType},
					},
				},
			}
			convey.So(shouldProcessPod(pod), convey.ShouldBeFalse)
		})
	})
}

func TestShouldProcessPod2(t *testing.T) {
	convey.Convey("Test shouldProcessPod", t, func() {
		convey.Convey("Should return false when instance set name is missing", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						common.OperatorNameKey:           common.TrueBool,
						common.InferServiceNameLabelKey:  "test-service",
						common.InstanceIndexLabelKey:     "0",
						common.ContainerSnapshotLabelKey: common.TrueBool,
					},
				},
				Spec: corev1.PodSpec{
					ReadinessGates: []corev1.PodReadinessGate{
						{ConditionType: common.PodSnapshotReadyConditionType},
					},
				},
			}
			convey.So(shouldProcessPod(pod), convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when instance index is missing", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						common.OperatorNameKey:           common.TrueBool,
						common.InferServiceNameLabelKey:  "test-service",
						common.InstanceSetNameLabelKey:   "test-instance",
						common.ContainerSnapshotLabelKey: common.TrueBool,
					},
				},
				Spec: corev1.PodSpec{
					ReadinessGates: []corev1.PodReadinessGate{
						{ConditionType: common.PodSnapshotReadyConditionType},
					},
				},
			}
			convey.So(shouldProcessPod(pod), convey.ShouldBeFalse)
		})

		convey.Convey("Should return true even without snapshot readiness gate", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						common.OperatorNameKey:           common.TrueBool,
						common.InferServiceNameLabelKey:  "test-service",
						common.InstanceSetNameLabelKey:   "test-instance",
						common.InstanceIndexLabelKey:     "0",
						common.ContainerSnapshotLabelKey: common.TrueBool,
					},
				},
				Spec: corev1.PodSpec{},
			}
			convey.So(shouldProcessPod(pod), convey.ShouldBeTrue)
		})

		convey.Convey("Should return false when operator name is not true", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						common.OperatorNameKey:           "false",
						common.InferServiceNameLabelKey:  "test-service",
						common.InstanceSetNameLabelKey:   "test-instance",
						common.InstanceIndexLabelKey:     "0",
						common.ContainerSnapshotLabelKey: common.TrueBool,
					},
				},
				Spec: corev1.PodSpec{
					ReadinessGates: []corev1.PodReadinessGate{
						{ConditionType: common.PodSnapshotReadyConditionType},
					},
				},
			}
			convey.So(shouldProcessPod(pod), convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when container snapshot label is missing", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						common.OperatorNameKey:          common.TrueBool,
						common.InferServiceNameLabelKey: "test-service",
						common.InstanceSetNameLabelKey:  "test-instance",
						common.InstanceIndexLabelKey:    "0",
					},
				},
				Spec: corev1.PodSpec{},
			}
			convey.So(shouldProcessPod(pod), convey.ShouldBeFalse)
		})
	})
}
