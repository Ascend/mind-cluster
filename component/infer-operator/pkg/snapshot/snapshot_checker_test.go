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
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

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

var localScheme = runtime.NewScheme()

func init() {
	_ = scheme.AddToScheme(localScheme)
	_ = v1.AddToScheme(localScheme)
}

func getCheckerTestScheme() *runtime.Scheme {
	return localScheme
}

func newFakeClientBuilder(objects ...runtime.Object) *fake.ClientBuilder {
	return fake.NewClientBuilder().WithScheme(getCheckerTestScheme()).WithRuntimeObjects(objects...)
}

func createTestPod(name, namespace string, annotations map[string]string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
			Labels: map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "main",
					Env: []corev1.EnvVar{
						{
							Name:  common.HostSnapshotDirPathEnvKey,
							Value: "/data/snapshot/host",
						},
					},
				},
			},
		},
	}
}

func createTestInstanceSet(name, namespace string, replicas int32) *v1.InstanceSet {
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
		},
	}
}

func TestNewSnapshotChecker(t *testing.T) {
	convey.Convey("Test NewSnapshotChecker function", t, func() {
		convey.Convey("Should create SnapshotChecker with correct initial values", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)

			convey.So(checker, convey.ShouldNotBeNil)
			convey.So(checker.Client, convey.ShouldEqual, fakeClient)
			convey.So(checker.instanceTrackers, convey.ShouldNotBeNil)
			convey.So(len(checker.instanceTrackers), convey.ShouldEqual, 0)
			convey.So(checker.running, convey.ShouldBeFalse)
		})
	})
}

func TestSnapshotCheckerStart(t *testing.T) {
	convey.Convey("Test SnapshotChecker Start method", t, func() {
		convey.Convey("Should set context correctly", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)

			ctx := context.Background()
			checker.Start(ctx)

			convey.So(checker.ctx, convey.ShouldNotBeNil)
		})
	})
}

func TestSnapshotCheckerStop(t *testing.T) {
	convey.Convey("Test SnapshotChecker Stop method", t, func() {
		convey.Convey("Should stop running checker and clear trackers", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			checker.Start(context.Background())

			instanceSet := createTestInstanceSet("test-instance", "default", int32(1))
			_ = checker.TrackInstanceSet(instanceSet, map[string]string{"app": "test"}, int32(1))

			checker.Stop()

			convey.So(checker.running, convey.ShouldBeFalse)
			convey.So(len(checker.instanceTrackers), convey.ShouldEqual, 0)
		})

		convey.Convey("Should handle stop when not running", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)

			checker.Stop()

			convey.So(checker.running, convey.ShouldBeFalse)
		})
	})
}

func TestSnapshotCheckerTrackInstanceSet(t *testing.T) {
	convey.Convey("Test SnapshotChecker TrackInstanceSet method", t, func() {
		convey.Convey("Should return error when instanceSet is nil", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			err := checker.TrackInstanceSet(nil,
				map[string]string{"app": "test"}, int32(1))
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "instanceSet is nil")
		})

		convey.Convey("Should return error when selectLabels is empty", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			instanceSet := createTestInstanceSet("test-instance", "default", int32(1))
			err := checker.TrackInstanceSet(instanceSet, map[string]string{}, int32(1))
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "selectLabels is empty")
		})

		convey.Convey("Should successfully track InstanceSet", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			checker.Start(context.Background())
			instanceSet := createTestInstanceSet("test-instance", "default", int32(1))
			selectLabels := map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
			}
			err := checker.TrackInstanceSet(instanceSet, selectLabels, int32(1))
			convey.So(err, convey.ShouldBeNil)
			convey.So(checker.GetTrackerCount(), convey.ShouldEqual, 1)
		})

		convey.Convey("Should handle duplicate tracking", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			checker.Start(context.Background())
			instanceSet := createTestInstanceSet("test-instance", "default", int32(1))
			selectLabels := map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
			}
			_ = checker.TrackInstanceSet(instanceSet, selectLabels, int32(1))
			err := checker.TrackInstanceSet(instanceSet, selectLabels, int32(1))
			convey.So(err, convey.ShouldBeNil)
			convey.So(checker.GetTrackerCount(), convey.ShouldEqual, 1)
		})
	})
}

func TestSnapshotCheckerGetInstanceSetKey(t *testing.T) {
	convey.Convey("Test SnapshotChecker getInstanceSetKey method", t, func() {
		convey.Convey("Should return correct key format", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)

			key := checker.getInstanceSetKey("default", "test-instance")
			convey.So(key, convey.ShouldEqual, "default/test-instance")
		})
	})
}

func TestSnapshotCheckerCheckPodSnapshotStatus(t *testing.T) {
	convey.Convey("Test SnapshotChecker checkPodSnapshotStatus method", t, func() {
		convey.Convey("Should return false when pod has no annotations", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			pod := createTestPod("test-pod", "default", nil)
			pod.Annotations = nil

			finished := checker.checkPodSnapshotStatus(pod)
			convey.So(finished, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when annotation key not exists", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			pod := createTestPod("test-pod", "default", map[string]string{})

			finished := checker.checkPodSnapshotStatus(pod)
			convey.So(finished, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when annotation value is not true", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			pod := createTestPod("test-pod", "default", map[string]string{
				common.HostSnapshotFlagAnnotationKey: common.FalseBool,
			})

			finished := checker.checkPodSnapshotStatus(pod)
			convey.So(finished, convey.ShouldBeFalse)
		})

		convey.Convey("Should return true when annotation value is true", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			pod := createTestPod("test-pod", "default", map[string]string{
				common.HostSnapshotFlagAnnotationKey: common.TrueBool,
			})

			finished := checker.checkPodSnapshotStatus(pod)
			convey.So(finished, convey.ShouldBeTrue)
		})
	})
}

func TestSnapshotCheckerCleanupSnapshotPath(t *testing.T) {
	convey.Convey("Test SnapshotChecker cleanupSnapshotPath method", t, func() {
		convey.Convey("Should return error when snapshot path is empty", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)

			err := checker.cleanupSnapshotPath("")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "snapshot path is empty")
		})

		convey.Convey("Should return nil when path does not exist", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)

			err := checker.cleanupSnapshotPath("/non/existent/path")
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should clean up directory but preserve status file", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)

			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			testDir := filepath.Join(tmpDir, "testdir")
			err = os.MkdirAll(testDir, 0755)
			convey.So(err, convey.ShouldBeNil)

			testFile := filepath.Join(tmpDir, "test.txt")
			err = os.WriteFile(testFile, []byte("test content"), 0644)
			convey.So(err, convey.ShouldBeNil)

			statusFile := filepath.Join(tmpDir, common.SnapshotStatusFileName)
			err = os.WriteFile(statusFile, []byte(`{"status":"success"}`), 0644)
			convey.So(err, convey.ShouldBeNil)

			err = checker.cleanupSnapshotPath(tmpDir)
			convey.So(err, convey.ShouldBeNil)

			_, err = os.Stat(tmpDir)
			convey.So(os.IsNotExist(err), convey.ShouldBeFalse)

			_, err = os.Stat(testDir)
			convey.So(os.IsNotExist(err), convey.ShouldBeTrue)

			_, err = os.Stat(testFile)
			convey.So(os.IsNotExist(err), convey.ShouldBeTrue)

			_, err = os.Stat(statusFile)
			convey.So(os.IsNotExist(err), convey.ShouldBeFalse)
		})

		convey.Convey("Should handle empty directory", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)

			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			err = checker.cleanupSnapshotPath(tmpDir)
			convey.So(err, convey.ShouldBeNil)

			_, err = os.Stat(tmpDir)
			convey.So(os.IsNotExist(err), convey.ShouldBeFalse)
		})

		convey.Convey("Should preserve only status file when it exists", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)

			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			statusFile := filepath.Join(tmpDir, common.SnapshotStatusFileName)
			err = os.WriteFile(statusFile, []byte(`{"status":"success"}`), 0644)
			convey.So(err, convey.ShouldBeNil)

			err = checker.cleanupSnapshotPath(tmpDir)
			convey.So(err, convey.ShouldBeNil)

			_, err = os.Stat(statusFile)
			convey.So(os.IsNotExist(err), convey.ShouldBeFalse)
		})
	})
}

func TestSnapshotCheckerRemoveTracker(t *testing.T) {
	convey.Convey("Test SnapshotChecker removeTracker method", t, func() {
		convey.Convey("Should successfully remove tracker", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			checker.Start(context.Background())

			instanceSet := createTestInstanceSet("test-instance", "default", int32(1))
			selectLabels := map[string]string{"app": "test"}
			_ = checker.TrackInstanceSet(instanceSet, selectLabels, int32(1))

			convey.So(checker.GetTrackerCount(), convey.ShouldEqual, 1)

			checker.removeTracker("default/test-instance")
			convey.So(checker.GetTrackerCount(), convey.ShouldEqual, 0)
		})
	})
}

func TestSnapshotCheckerGetTrackerCount(t *testing.T) {
	convey.Convey("Test SnapshotChecker GetTrackerCount method", t, func() {
		convey.Convey("Should return correct count", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			checker.Start(context.Background())

			convey.So(checker.GetTrackerCount(), convey.ShouldEqual, 0)

			instanceSet1 := createTestInstanceSet("instance1", "default", int32(1))
			instanceSet2 := createTestInstanceSet("instance2", "default", int32(1))
			_ = checker.TrackInstanceSet(instanceSet1,
				map[string]string{"app": "test1"}, int32(1))
			_ = checker.TrackInstanceSet(instanceSet2,
				map[string]string{"app": "test2"}, int32(1))

			convey.So(checker.GetTrackerCount(), convey.ShouldEqual, 2)
		})
	})
}

func TestSnapshotCheckerIsRunning(t *testing.T) {
	convey.Convey("Test SnapshotChecker IsRunning method", t, func() {
		convey.Convey("Should return false initially", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)

			convey.So(checker.IsRunning(), convey.ShouldBeFalse)
		})

		convey.Convey("Should return true after tracking starts", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			checker.Start(context.Background())

			instanceSet := createTestInstanceSet("test-instance", "default", int32(1))
			_ = checker.TrackInstanceSet(instanceSet, map[string]string{"app": "test"}, int32(1))

			convey.So(checker.IsRunning(), convey.ShouldBeTrue)
		})

		convey.Convey("Should return false after stop", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			checker.Start(context.Background())

			instanceSet := createTestInstanceSet("test-instance", "default", int32(1))
			_ = checker.TrackInstanceSet(instanceSet, map[string]string{"app": "test"}, int32(1))
			checker.Stop()

			convey.So(checker.IsRunning(), convey.ShouldBeFalse)
		})
	})
}

func TestSnapshotCheckerUpdatePodReadinessGate(t *testing.T) {
	convey.Convey("Test SnapshotChecker updatePodReadinessGate method", t, func() {
		convey.Convey("Should successfully update pod readiness gate", func() {
			pod := createTestPod("test-pod", "default", nil)
			fakeClient := newFakeClientBuilder(pod).Build()
			checker := NewSnapshotChecker(fakeClient)

			ctx := context.Background()
			err := checker.updatePodReadinessGate(ctx, pod, true)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when patch fails", func() {
			pod := createTestPod("test-pod", "default", nil)
			fakeClient := newFakeClientBuilder(pod).Build()
			checker := NewSnapshotChecker(fakeClient)

			patches := gomonkey.ApplyMethodReturn(fakeClient, "Status", &mockStatusWriter{
				patchErr: errors.New("patch error"),
			})
			defer patches.Reset()

			ctx := context.Background()
			err := checker.updatePodReadinessGate(ctx, pod, true)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestSetPodsActiveLabel(t *testing.T) {
	convey.Convey("Test SnapshotChecker setPodsActiveLabel method", t, func() {
		convey.Convey("Should return nil when pods active label exists", func() {
			pod := createTestPod("test-pod", "default", nil)
			fakeClient := newFakeClientBuilder(pod).Build()
			checker := NewSnapshotChecker(fakeClient)

			ctx := context.Background()

			err := checker.setPodsActiveLabel(ctx, []corev1.Pod{*pod})
			pod.Labels[common.ActiveLabelKey] = common.TrueBool
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should return nil when patch succeed", func() {
			pod := createTestPod("test-pod", "default", nil)
			fakeClient := newFakeClientBuilder(pod).Build()
			checker := NewSnapshotChecker(fakeClient)

			patches := gomonkey.ApplyMethodReturn(fakeClient, "Patch", nil)
			defer patches.Reset()

			ctx := context.Background()
			err := checker.setPodsActiveLabel(ctx, []corev1.Pod{*pod})
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSnapshotCheckerCheckInstanceSetSnapshot(t *testing.T) {
	convey.Convey("Test SnapshotChecker checkInstanceSetSnapshot method", t, func() {
		convey.Convey("Should handle non-existent tracker", func() {
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)

			checker.checkInstanceSetSnapshot("non/existent")
			convey.So(checker.GetTrackerCount(), convey.ShouldEqual, 0)
		})

		convey.Convey("Should handle empty pod list", func() {
			instanceSet := createTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := newFakeClientBuilder(instanceSet).Build()
			checker := NewSnapshotChecker(fakeClient)
			checker.Start(context.Background())

			selectLabels := map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
			}
			_ = checker.TrackInstanceSet(instanceSet, selectLabels, int32(1))

			checker.checkInstanceSetSnapshot("default/test-instance")
			convey.So(checker.GetTrackerCount(), convey.ShouldEqual, 1)
		})

		convey.Convey("Should handle list pods error", func() {
			instanceSet := createTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := newFakeClientBuilder(instanceSet).Build()
			checker := NewSnapshotChecker(fakeClient)
			checker.Start(context.Background())

			selectLabels := map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
			}
			_ = checker.TrackInstanceSet(instanceSet, selectLabels, int32(1))

			patches := gomonkey.ApplyMethodReturn(fakeClient, "List", errors.New("list error"))
			defer patches.Reset()

			checker.checkInstanceSetSnapshot("default/test-instance")
		})
	})
}

func TestSetAndCleanSnapshot(t *testing.T) {
	convey.Convey("Test SnapshotChecker setAndCleanSnapshot method", t, func() {
		convey.Convey("Should remove tracker when finished", func() {
			instanceSet := createTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := newFakeClientBuilder().Build()
			checker := NewSnapshotChecker(fakeClient)
			checker.Start(context.Background())
			selectLabels := map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
			}
			_ = checker.TrackInstanceSet(instanceSet, selectLabels, int32(1))
			snapshotPods := []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-ns",
					},
				},
			}
			tracker, _ := checker.instanceTrackers["default/test-instance"]
			patches := gomonkey.ApplyFuncReturn(GetHostSnapshotPath, "test")
			defer patches.Reset()
			checker.setAndCleanSnapshot("default/test-instance", true, snapshotPods, tracker,
				context.TODO(), &corev1.PodList{})
			convey.So(checker.GetTrackerCount(), convey.ShouldEqual, 0)
		})

		convey.Convey("Should remove tracker when timeout", func() {
			instanceSet := createTestInstanceSet("test-instance", "default", int32(1))
			fakeClient := newFakeClientBuilder(instanceSet).Build()
			checker := NewSnapshotChecker(fakeClient)
			checker.Start(context.Background())

			selectLabels := map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
			}
			_ = checker.TrackInstanceSet(instanceSet, selectLabels, int32(1))
			tracker, _ := checker.instanceTrackers["default/test-instance"]
			tracker.StartTime = time.Now().Add(-60 * time.Minute)
			podList := &corev1.PodList{
				Items: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test-ns",
						},
					},
				},
			}
			patches := gomonkey.ApplyFuncReturn(GetHostSnapshotPath, "test")
			defer patches.Reset()
			checker.setAndCleanSnapshot("default/test-instance", false, []corev1.Pod{}, tracker,
				context.TODO(), podList)
			convey.So(checker.GetTrackerCount(), convey.ShouldEqual, 0)
		})
	})
}

func TestSnapshotCheckerCheckAllInstanceSets(t *testing.T) {
	convey.Convey("Test SnapshotChecker checkAllInstanceSets method", t, func() {
		convey.Convey("Should check all tracked InstanceSets", func() {
			instanceSet1 := createTestInstanceSet("instance1", "default", int32(1))
			instanceSet2 := createTestInstanceSet("instance2", "ns1", int32(1))
			fakeClient := newFakeClientBuilder(instanceSet1, instanceSet2).Build()
			checker := NewSnapshotChecker(fakeClient)
			checker.Start(context.Background())

			_ = checker.TrackInstanceSet(instanceSet1,
				map[string]string{"app": "test1"}, int32(1))
			_ = checker.TrackInstanceSet(instanceSet2,
				map[string]string{"app": "test2"}, int32(1))

			checker.checkAllInstanceSets()
			convey.So(checker.GetTrackerCount(), convey.ShouldEqual, 2)
		})
	})
}

type mockStatusWriter struct {
	patchErr error
}

func (m *mockStatusWriter) Update(context.Context, client.Object, ...client.UpdateOption) error {
	return nil
}

func (m *mockStatusWriter) Patch(context.Context, client.Object, client.Patch, ...client.PatchOption) error {
	return m.patchErr
}

func TestGetHostSnapshotPath(t *testing.T) {
	convey.Convey("Test GetHostSnapshotPath function", t, func() {
		convey.Convey("Should return correct path when env exists", func() {
			pod := createTestPod("test-pod", "default", nil)

			path := GetHostSnapshotPath(pod)
			convey.So(path, convey.ShouldEqual,
				"/data/snapshot/host/default/test-service-test-role")
		})

		convey.Convey("Should return empty string when env not found", func() {
			pod := createTestPod("test-pod", "default", nil)
			pod.Spec.Containers[0].Env = []corev1.EnvVar{}

			path := GetHostSnapshotPath(pod)
			convey.So(path, convey.ShouldEqual, "")
		})

		convey.Convey("Should return empty string when HostPath is nil", func() {
			pod := createTestPod("test-pod", "default", nil)
			pod.Spec.Containers[0].Env = []corev1.EnvVar{
				{
					Name:  common.HostSnapshotDirPathEnvKey,
					Value: "",
				},
			}

			path := GetHostSnapshotPath(pod)
			convey.So(path, convey.ShouldEqual, "")
		})
	})
}

func TestInstanceSetTracker(t *testing.T) {
	convey.Convey("Test InstanceSetTracker struct", t, func() {
		convey.Convey("Should create tracker with correct values", func() {
			now := time.Now()
			tracker := &InstanceSetTracker{
				InstanceSetName: "test-instance",
				Namespace:       "default",
				SelectLabels:    map[string]string{"app": "test"},
				StartTime:       now,
				Replicas:        int32(3),
			}

			convey.So(tracker.InstanceSetName, convey.ShouldEqual, "test-instance")
			convey.So(tracker.Namespace, convey.ShouldEqual, "default")
			convey.So(tracker.SelectLabels["app"], convey.ShouldEqual, "test")
			convey.So(tracker.StartTime, convey.ShouldEqual, now)
			convey.So(tracker.Replicas, convey.ShouldEqual, int32(3))
		})
	})
}

func TestSnapshotCheckerWithRealSnapshotPath(t *testing.T) {
	convey.Convey("Test SnapshotChecker with real snapshot path", t, func() {
		convey.Convey("Should handle snapshot status file operations", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-status-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			statusFile := filepath.Join(tmpDir, common.SnapshotStatusFileName)
			statusContent := `{"status":"success","timestamp":"2024-01-01T00:00:00Z","message":"test"}`
			err = os.WriteFile(statusFile, []byte(statusContent), 0644)
			convey.So(err, convey.ShouldBeNil)

			exists := common.IsSnapshotStatusExists(tmpDir)
			convey.So(exists, convey.ShouldBeTrue)
		})
	})
}
