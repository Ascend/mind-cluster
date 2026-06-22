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
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
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

func createTestInstanceSetWithSnapshot(name, namespace string, replicas int32, enableSnapshot bool) *v1.InstanceSet {
	statefulSetSpec := &appsv1.StatefulSetSpec{
		Replicas: &replicas,
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
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
		},
	}
	specBytes, _ := json.Marshal(statefulSetSpec)

	instanceSet := &v1.InstanceSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "prefill",
				common.OperatorNameKey:          common.TrueBool,
			},
		},
		Spec: v1.InstanceSetSpec{
			Name:     "prefill",
			Replicas: &replicas,
			WorkloadTypeMeta: v1.WorkloadType{
				Kind:       "StatefulSet",
				APIVersion: "apps/v1",
			},
			InstanceSpec: runtime.RawExtension{Raw: specBytes},
		},
	}
	if enableSnapshot {
		instanceSet.Labels[common.ContainerSnapshotLabelKey] = common.TrueBool
	}
	return instanceSet
}

func TestInstanceSetSnapshotReconcilerReconcile(t *testing.T) {
	convey.Convey("Test InstanceSetSnapshotReconciler Reconcile method", t, func() {
		convey.Convey("Should return empty result when InstanceSet not found", func() {
			fakeClient := getFakeClientBuilder().Build()
			reconciler := &InstanceSetSnapshotReconciler{
				Client:          fakeClient,
				SnapshotChecker: NewSnapshotChecker(fakeClient),
			}

			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "non-existent",
					Namespace: "default",
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})

		convey.Convey("Should start snapshot checker if not running", func() {
			instanceSet := createTestInstanceSetWithSnapshot("test-instance", "default", int32(1), true)
			fakeClient := getFakeClientBuilder().WithRuntimeObjects(instanceSet).Build()

			reconciler := &InstanceSetSnapshotReconciler{
				Client:          fakeClient,
				SnapshotChecker: NewSnapshotChecker(fakeClient),
			}

			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-instance",
					Namespace: "default",
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})
	})
}

func TestInstanceSetSnapshotReconcilerReconcile2(t *testing.T) {
	convey.Convey("Test InstanceSetSnapshotReconciler Reconcile method", t, func() {
		convey.Convey("Should track InstanceSet with correct labels", func() {
			instanceSet := createTestInstanceSetWithSnapshot("test-instance", "default", int32(1), true)
			fakeClient := getFakeClientBuilder().WithRuntimeObjects(instanceSet).Build()

			reconciler := &InstanceSetSnapshotReconciler{
				Client:          fakeClient,
				SnapshotChecker: NewSnapshotChecker(fakeClient),
			}

			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-instance",
					Namespace: "default",
				},
			}

			_, err := reconciler.Reconcile(ctx, req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(reconciler.SnapshotChecker.GetTrackerCount(), convey.ShouldEqual, 1)
		})

		convey.Convey("Should handle TrackInstanceSet error gracefully", func() {
			instanceSet := createTestInstanceSetWithSnapshot("test-instance", "default", int32(1), true)
			fakeClient := getFakeClientBuilder().WithRuntimeObjects(instanceSet).Build()

			reconciler := &InstanceSetSnapshotReconciler{
				Client:          fakeClient,
				SnapshotChecker: NewSnapshotChecker(fakeClient),
			}

			patches := gomonkey.ApplyMethodReturn(reconciler.SnapshotChecker, "TrackInstanceSet",
				errors.New("track error"))
			defer patches.Reset()

			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-instance",
					Namespace: "default",
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, reconcile.Result{})
		})
	})
}

func TestInstanceSetSnapshotPredicate(t *testing.T) {
	convey.Convey("Test InstanceSetSnapshot predicate", t, func() {
		convey.Convey("Should return true for create event with snapshot enabled", func() {
			instanceSet := createTestInstanceSetWithSnapshot("test-instance", "default", int32(1), true)
			createEvent := event.CreateEvent{
				Object: instanceSet,
			}

			predicate := snapshotPredicate()
			result := predicate.Create(createEvent)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("Should return false for create event with snapshot disabled", func() {
			instanceSet := createTestInstanceSetWithSnapshot("test-instance", "default", int32(1), false)
			createEvent := event.CreateEvent{
				Object: instanceSet,
			}

			predicate := snapshotPredicate()
			result := predicate.Create(createEvent)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false for create event with zero replicas", func() {
			instanceSet := createTestInstanceSetWithSnapshot("test-instance", "default", int32(0), true)
			createEvent := event.CreateEvent{
				Object: instanceSet,
			}

			predicate := snapshotPredicate()
			result := predicate.Create(createEvent)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false for create event with nil replicas", func() {
			instanceSet := createTestInstanceSetWithSnapshot("test-instance", "default", int32(1), true)
			instanceSet.Spec.Replicas = nil
			createEvent := event.CreateEvent{
				Object: instanceSet,
			}

			predicate := snapshotPredicate()
			result := predicate.Create(createEvent)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestInstanceSetSnapshotPredicate2(t *testing.T) {
	convey.Convey("Test InstanceSetSnapshot predicate", t, func() {
		convey.Convey("Should return false for create event with negative replicas", func() {
			instanceSet := createTestInstanceSetWithSnapshot("test-instance", "default", int32(-1), true)
			createEvent := event.CreateEvent{
				Object: instanceSet,
			}

			predicate := snapshotPredicate()
			result := predicate.Create(createEvent)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false for create event with wrong object type", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
			}
			createEvent := event.CreateEvent{
				Object: pod,
			}

			predicate := snapshotPredicate()
			result := predicate.Create(createEvent)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false for update event", func() {
			predicate := snapshotPredicate()
			result := predicate.Update(event.UpdateEvent{})
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false for delete event", func() {
			predicate := snapshotPredicate()
			result := predicate.Delete(event.DeleteEvent{})
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false for generic event", func() {
			predicate := snapshotPredicate()
			result := predicate.Generic(event.GenericEvent{})
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestNewInstanceSetSnapshotReconciler(t *testing.T) {
	convey.Convey("Test NewInstanceSetSnapshotReconciler function", t, func() {
		convey.Convey("Should create reconciler with correct values", func() {
			fakeClient := getFakeClientBuilder().Build()

			reconciler := &InstanceSetSnapshotReconciler{
				Client:          fakeClient,
				Scheme:          getTestScheme(),
				SnapshotChecker: NewSnapshotChecker(fakeClient),
			}

			convey.So(reconciler, convey.ShouldNotBeNil)
			convey.So(reconciler.Client, convey.ShouldNotBeNil)
			convey.So(reconciler.SnapshotChecker, convey.ShouldNotBeNil)
		})
	})
}

func getTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)
	_ = v1.AddToScheme(s)
	return s
}

func getFakeClientBuilder() *fake.ClientBuilder {
	return fake.NewClientBuilder().WithScheme(getTestScheme())
}

func TestGetStsSpecFromInstanceSpec(t *testing.T) {
	convey.Convey("Test getStsSpecFromInstanceSpec function", t, func() {
		convey.Convey("Should return error when raw extension is empty", func() {
			raw := runtime.RawExtension{Raw: []byte{}}
			spec, err := getStsSpecFromInstanceSpec(raw)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(spec, convey.ShouldBeNil)
		})

		convey.Convey("Should return error for invalid JSON", func() {
			raw := runtime.RawExtension{Raw: []byte("invalid json")}
			spec, err := getStsSpecFromInstanceSpec(raw)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(spec, convey.ShouldBeNil)
		})

		convey.Convey("Should return spec with nil replicas when not set", func() {
			statefulSetSpec := &appsv1.StatefulSetSpec{}
			specBytes, _ := json.Marshal(statefulSetSpec)
			raw := runtime.RawExtension{Raw: specBytes}

			spec, err := getStsSpecFromInstanceSpec(raw)
			convey.So(err, convey.ShouldBeNil)
			convey.So(spec, convey.ShouldNotBeNil)
			convey.So(spec.Replicas, convey.ShouldBeNil)
		})

		convey.Convey("Should return spec with correct replicas value", func() {
			replicasVal := int32(5)
			statefulSetSpec := &appsv1.StatefulSetSpec{
				Replicas: &replicasVal,
			}
			specBytes, _ := json.Marshal(statefulSetSpec)
			raw := runtime.RawExtension{Raw: specBytes}

			spec, err := getStsSpecFromInstanceSpec(raw)
			convey.So(err, convey.ShouldBeNil)
			convey.So(spec, convey.ShouldNotBeNil)
			convey.So(*spec.Replicas, convey.ShouldEqual, int32(5))
		})

		convey.Convey("Should handle spec with template", func() {
			statefulSetSpec := &appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "main"},
						},
					},
				},
			}
			specBytes, _ := json.Marshal(statefulSetSpec)
			raw := runtime.RawExtension{Raw: specBytes}

			spec, err := getStsSpecFromInstanceSpec(raw)
			convey.So(err, convey.ShouldBeNil)
			convey.So(spec, convey.ShouldNotBeNil)
			convey.So(len(spec.Template.Spec.Containers), convey.ShouldEqual, 1)
		})

		convey.Convey("Should handle zero replicas", func() {
			replicasVal := int32(0)
			statefulSetSpec := &appsv1.StatefulSetSpec{
				Replicas: &replicasVal,
			}
			specBytes, _ := json.Marshal(statefulSetSpec)
			raw := runtime.RawExtension{Raw: specBytes}

			spec, err := getStsSpecFromInstanceSpec(raw)
			convey.So(err, convey.ShouldBeNil)
			convey.So(spec, convey.ShouldNotBeNil)
			convey.So(*spec.Replicas, convey.ShouldEqual, int32(0))
		})
	})
}
