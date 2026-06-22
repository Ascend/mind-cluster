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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"ascend-common/common-utils/hwlog"
	v1 "infer-operator/pkg/api/v1"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func newFakeClientBuilder() *fake.ClientBuilder {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = v1.AddToScheme(scheme)
	return fake.NewClientBuilder().WithScheme(scheme)
}

func createTestPodTemplate() *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "main",
					Env: []corev1.EnvVar{
						{
							Name:  HostSnapshotDirPathEnvKey,
							Value: "/data/host-snapshot",
						},
					},
				},
			},
		},
	}
}

func createTestInstanceSetForSnapshot(enableSnapshot bool) *v1.InstanceSet {
	instanceSet := &v1.InstanceSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-instance",
			Namespace: "default",
			Labels: map[string]string{
				InferServiceNameLabelKey: "test-service",
				InstanceSetNameLabelKey:  "prefill",
				OperatorNameKey:          TrueBool,
			},
		},
		Spec: v1.InstanceSetSpec{
			Name:     "prefill",
			Replicas: func() *int32 { r := int32(1); return &r }(),
		},
	}
	if enableSnapshot {
		instanceSet.Labels[ContainerSnapshotLabelKey] = TrueBool
	}
	return instanceSet
}

func TestAddSnapshotEnvToPodTemplate(t *testing.T) {
	convey.Convey("Test AddSnapshotEnvToPodTemplate function", t, func() {

		convey.Convey("Should not add env when snapshot is disabled", func() {
			pod := createTestPodTemplate()
			instanceSet := createTestInstanceSetForSnapshot(false)

			AddSnapshotInfoToPodTemplate(pod, instanceSet, "test")

			convey.So(len(pod.Spec.Containers[0].Env), convey.ShouldEqual, 1)
		})

		convey.Convey("Should add env when instanceSet is not prefill/decode", func() {
			pod := createTestPodTemplate()
			instanceSet := createTestInstanceSetForSnapshot(true)
			instanceSet.Labels[InstanceSetNameLabelKey] = "other-role"
			instanceSet.Spec.Name = "other-role"

			AddSnapshotInfoToPodTemplate(pod, instanceSet, "test")

			convey.So(len(pod.Spec.Containers[0].Env), convey.ShouldEqual, 4)
		})

		convey.Convey("Should not add env when host snapshot env is empty", func() {
			pod := createTestPodTemplate()
			pod.Spec.Containers[0].Env = []corev1.EnvVar{}
			instanceSet := createTestInstanceSetForSnapshot(true)

			AddSnapshotInfoToPodTemplate(pod, instanceSet, "test")

			convey.So(len(pod.Spec.Containers[0].Env), convey.ShouldEqual, 0)
		})

		convey.Convey("Should add env when snapshot exist", func() {
			pod := createTestPodTemplate()
			instanceSet := createTestInstanceSetForSnapshot(true)

			AddSnapshotInfoToPodTemplate(pod, instanceSet, "test")

			convey.So(len(pod.Spec.Containers[0].Env), convey.ShouldEqual, 4)
			envNames := make([]string, len(pod.Spec.Containers[0].Env))
			for i, env := range pod.Spec.Containers[0].Env {
				envNames[i] = env.Name
			}
			convey.So(envNames, convey.ShouldContain, GrusSnapshotRestoredFlag)
			convey.So(envNames, convey.ShouldContain, HostSnapshotPathEnvKey)
			convey.So(envNames, convey.ShouldContain, PodNameEnvKey)
		})
	})
}

func TestAddSnapshotEnvToPodTemplate2(t *testing.T) {
	convey.Convey("Test AddSnapshotEnvToPodTemplate function with existing snapshot", t, func() {
		convey.Convey("Should add env when snapshot exists and is valid", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			testDir := filepath.Join(tmpDir, "testdir")
			err = os.MkdirAll(testDir, 0755)
			convey.So(err, convey.ShouldBeNil)

			testFile := filepath.Join(testDir, "test.txt")
			err = os.WriteFile(testFile, []byte("test content"), 0644)
			convey.So(err, convey.ShouldBeNil)

			err = WriteSnapshotStatus(tmpDir, SnapshotStatusSuccess, "test message")
			convey.So(err, convey.ShouldBeNil)

			pod := &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "main",
							Env: []corev1.EnvVar{
								{
									Name:  HostSnapshotDirPathEnvKey,
									Value: "/data/host-snapshot",
								},
							},
						},
					},
				},
			}
			instanceSet := createTestInstanceSetForSnapshot(true)

			AddSnapshotInfoToPodTemplate(pod, instanceSet, "test")

			convey.So(len(pod.Spec.Containers[0].Env), convey.ShouldEqual, 4)
			convey.So(pod.Spec.Containers[0].Env[1].Name, convey.ShouldEqual, GrusSnapshotRestoredFlag)
			convey.So(pod.Spec.Containers[0].Env[2].Name, convey.ShouldEqual, HostSnapshotPathEnvKey)
			convey.So(pod.Spec.Containers[0].Env[3].Name, convey.ShouldEqual, PodNameEnvKey)
		})
	})
}

func TestAddMetadataVolume(t *testing.T) {
	convey.Convey("Test AddMetadataVolume function", t, func() {
		convey.Convey("Should add correct volume", func() {
			pod := createTestPodTemplate()

			AddMetadataVolume(pod, "test-cm", createTestInstanceSetForSnapshot(true))

			convey.So(len(pod.Spec.Volumes), convey.ShouldEqual, 1)
			convey.So(len(pod.Spec.Containers[0].VolumeMounts), convey.ShouldEqual, 1)
		})
	})
}

func TestGetHostSnapshotPathFromPodTemplate(t *testing.T) {
	convey.Convey("Test GetHostSnapshotPathFromPodTemplate function", t, func() {
		convey.Convey("Should return correct path when env exists", func() {
			pod := createTestPodTemplate()
			instanceSet := createTestInstanceSetForSnapshot(true)

			path := GetHostSnapshotPathFromPodTemplate(pod, instanceSet)
			convey.So(path, convey.ShouldEqual, "/data/host-snapshot/default/test-instance")
		})

		convey.Convey("Should return empty string when no containers", func() {
			pod := &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			}
			instanceSet := createTestInstanceSetForSnapshot(true)

			path := GetHostSnapshotPathFromPodTemplate(pod, instanceSet)
			convey.So(path, convey.ShouldEqual, "")
		})

		convey.Convey("Should return empty string when env not found", func() {
			pod := createTestPodTemplate()
			pod.Spec.Containers[0].Env = []corev1.EnvVar{}
			instanceSet := createTestInstanceSetForSnapshot(true)

			path := GetHostSnapshotPathFromPodTemplate(pod, instanceSet)
			convey.So(path, convey.ShouldEqual, "")
		})

		convey.Convey("Should return empty string when env path is empty", func() {
			pod := createTestPodTemplate()
			pod.Spec.Containers[0].Env = []corev1.EnvVar{
				{
					Name:  HostSnapshotDirPathEnvKey,
					Value: "",
				},
			}
			instanceSet := createTestInstanceSetForSnapshot(true)

			path := GetHostSnapshotPathFromPodTemplate(pod, instanceSet)
			convey.So(path, convey.ShouldEqual, "")
		})
	})
}

func TestGetSnapshotStatusFilePath(t *testing.T) {
	convey.Convey("Test GetSnapshotStatusFilePath function", t, func() {
		convey.Convey("Should return correct file path", func() {
			path := GetSnapshotStatusFilePath("/data/snapshot")
			expected := filepath.Join("/data/snapshot", SnapshotStatusFileName)
			convey.So(path, convey.ShouldEqual, expected)
		})
	})
}

func TestCalculateSnapshotSHA256(t *testing.T) {
	convey.Convey("Test CalculateSnapshotSHA256 function", t, func() {
		convey.Convey("Should return error when directory does not exist", func() {
			_, err := CalculateSnapshotSHA256("/non/existent/path")
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return empty map for empty directory", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			result, err := CalculateSnapshotSHA256(tmpDir)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("Should calculate SHA256 for directories with files", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			testDir := filepath.Join(tmpDir, "testdir")
			err = os.MkdirAll(testDir, 0755)
			convey.So(err, convey.ShouldBeNil)

			testFile := filepath.Join(testDir, "test.txt")
			err = os.WriteFile(testFile, []byte("test content"), 0644)
			convey.So(err, convey.ShouldBeNil)

			result, err := CalculateSnapshotSHA256(tmpDir)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 1)
			convey.So(result["testdir"], convey.ShouldNotBeEmpty)
		})

		convey.Convey("Should skip files in root directory", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			testFile := filepath.Join(tmpDir, "test.txt")
			err = os.WriteFile(testFile, []byte("test content"), 0644)
			convey.So(err, convey.ShouldBeNil)

			result, err := CalculateSnapshotSHA256(tmpDir)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})
	})
}

func TestCalculateSnapshotSHA256WithNestedDirs(t *testing.T) {
	convey.Convey("Test CalculateSnapshotSHA256 with nested directories", t, func() {
		convey.Convey("Should handle nested directories", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			testDir1 := filepath.Join(tmpDir, "dir1")
			err = os.MkdirAll(testDir1, 0755)
			convey.So(err, convey.ShouldBeNil)

			testDir2 := filepath.Join(tmpDir, "dir2")
			err = os.MkdirAll(testDir2, 0755)
			convey.So(err, convey.ShouldBeNil)

			testFile1 := filepath.Join(testDir1, "file1.txt")
			err = os.WriteFile(testFile1, []byte("content1"), 0644)
			convey.So(err, convey.ShouldBeNil)

			testFile2 := filepath.Join(testDir2, "file2.txt")
			err = os.WriteFile(testFile2, []byte("content2"), 0644)
			convey.So(err, convey.ShouldBeNil)

			result, err := CalculateSnapshotSHA256(tmpDir)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 2)
			convey.So(result["dir1"], convey.ShouldNotBeEmpty)
			convey.So(result["dir2"], convey.ShouldNotBeEmpty)
		})
	})
}

func TestWriteSnapshotStatus(t *testing.T) {
	convey.Convey("Test WriteSnapshotStatus function", t, func() {
		convey.Convey("Should return error when directory does not exist", func() {
			err := WriteSnapshotStatus("/non/existent/path", SnapshotStatusSuccess, "test")
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should write status file successfully", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			testDir := filepath.Join(tmpDir, "testdir")
			err = os.MkdirAll(testDir, 0755)
			convey.So(err, convey.ShouldBeNil)

			testFile := filepath.Join(testDir, "test.txt")
			err = os.WriteFile(testFile, []byte("test content"), 0644)
			convey.So(err, convey.ShouldBeNil)

			err = WriteSnapshotStatus(tmpDir, SnapshotStatusSuccess, "test message")
			convey.So(err, convey.ShouldBeNil)

			statusFile := filepath.Join(tmpDir, SnapshotStatusFileName)
			_, err = os.Stat(statusFile)
			convey.So(os.IsNotExist(err), convey.ShouldBeFalse)
		})
	})
}

func TestReadSnapshotStatus(t *testing.T) {
	convey.Convey("Test ReadSnapshotStatus function", t, func() {
		convey.Convey("Should return nil when file does not exist", func() {
			status, err := ReadSnapshotStatus("/non/existent/path")
			convey.So(err, convey.ShouldBeNil)
			convey.So(status, convey.ShouldBeNil)
		})

		convey.Convey("Should read status file successfully", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			statusContent := `{"directorySHA256":{"testdir":"abc123"},"status":"success","timestamp":"2024-01-01T00:00:00Z","message":"test"}`
			statusFile := filepath.Join(tmpDir, SnapshotStatusFileName)
			err = os.WriteFile(statusFile, []byte(statusContent), 0644)
			convey.So(err, convey.ShouldBeNil)

			status, err := ReadSnapshotStatus(tmpDir)
			convey.So(err, convey.ShouldBeNil)
			convey.So(status, convey.ShouldNotBeNil)
			convey.So(status.Status, convey.ShouldEqual, "success")
			convey.So(status.Message, convey.ShouldEqual, "test")
		})

		convey.Convey("Should return error for invalid JSON", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			statusFile := filepath.Join(tmpDir, SnapshotStatusFileName)
			err = os.WriteFile(statusFile, []byte("invalid json"), 0644)
			convey.So(err, convey.ShouldBeNil)

			_, err = ReadSnapshotStatus(tmpDir)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestIsSnapshotStatusExists(t *testing.T) {
	convey.Convey("Test IsSnapshotStatusExists function", t, func() {
		convey.Convey("Should return false when file does not exist", func() {
			exists := IsSnapshotStatusExists("/non/existent/path")
			convey.So(exists, convey.ShouldBeFalse)
		})

		convey.Convey("Should return true when file exists", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			statusFile := filepath.Join(tmpDir, SnapshotStatusFileName)
			err = os.WriteFile(statusFile, []byte("{}"), 0644)
			convey.So(err, convey.ShouldBeNil)

			exists := IsSnapshotStatusExists(tmpDir)
			convey.So(exists, convey.ShouldBeTrue)
		})
	})
}

func TestValidateSnapshotStatus(t *testing.T) {
	convey.Convey("Test ValidateSnapshotStatus function", t, func() {
		convey.Convey("Should return error when status file does not exist", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			valid, err := ValidateSnapshotStatus(tmpDir)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(valid, convey.ShouldBeFalse)
		})

		convey.Convey("Should return error when status is nil", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			valid, err := ValidateSnapshotStatus(tmpDir)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(valid, convey.ShouldBeFalse)
		})

		convey.Convey("Should return error when status is failed", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			statusContent := `{"directorySHA256":{},"status":"failed","timestamp":"2024-01-01T00:00:00Z","message":"error"}`
			statusFile := filepath.Join(tmpDir, SnapshotStatusFileName)
			err = os.WriteFile(statusFile, []byte(statusContent), 0644)
			convey.So(err, convey.ShouldBeNil)

			valid, err := ValidateSnapshotStatus(tmpDir)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(valid, convey.ShouldBeFalse)
		})

		convey.Convey("Should return error when DirectorySHA256 is empty", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			statusContent := `{"directorySHA256":{},"status":"success","timestamp":"2024-01-01T00:00:00Z","message":"test"}`
			statusFile := filepath.Join(tmpDir, SnapshotStatusFileName)
			err = os.WriteFile(statusFile, []byte(statusContent), 0644)
			convey.So(err, convey.ShouldBeNil)

			valid, err := ValidateSnapshotStatus(tmpDir)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(valid, convey.ShouldBeFalse)
		})
	})
}

func TestValidateSnapshotStatusWithValidData(t *testing.T) {
	convey.Convey("Test ValidateSnapshotStatus with valid data", t, func() {
		convey.Convey("Should return true when status is valid", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			testDir := filepath.Join(tmpDir, "testdir")
			err = os.MkdirAll(testDir, 0755)
			convey.So(err, convey.ShouldBeNil)

			testFile := filepath.Join(testDir, "test.txt")
			err = os.WriteFile(testFile, []byte("test content"), 0644)
			convey.So(err, convey.ShouldBeNil)

			err = WriteSnapshotStatus(tmpDir, SnapshotStatusSuccess, "test message")
			convey.So(err, convey.ShouldBeNil)

			valid, err := ValidateSnapshotStatus(tmpDir)
			convey.So(err, convey.ShouldBeNil)
			convey.So(valid, convey.ShouldBeTrue)
		})

		convey.Convey("Should return error when directory is missing", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			testDir := filepath.Join(tmpDir, "testdir")
			err = os.MkdirAll(testDir, 0755)
			convey.So(err, convey.ShouldBeNil)

			testFile := filepath.Join(testDir, "test.txt")
			err = os.WriteFile(testFile, []byte("test content"), 0644)
			convey.So(err, convey.ShouldBeNil)

			err = WriteSnapshotStatus(tmpDir, SnapshotStatusSuccess, "test message")
			convey.So(err, convey.ShouldBeNil)

			os.RemoveAll(testDir)

			valid, err := ValidateSnapshotStatus(tmpDir)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(valid, convey.ShouldBeFalse)
		})

		convey.Convey("Should return error when SHA256 mismatch", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			testDir := filepath.Join(tmpDir, "testdir")
			err = os.MkdirAll(testDir, 0755)
			convey.So(err, convey.ShouldBeNil)

			testFile := filepath.Join(testDir, "test.txt")
			err = os.WriteFile(testFile, []byte("test content"), 0644)
			convey.So(err, convey.ShouldBeNil)

			err = WriteSnapshotStatus(tmpDir, SnapshotStatusSuccess, "test message")
			convey.So(err, convey.ShouldBeNil)

			err = os.WriteFile(testFile, []byte("modified content"), 0644)
			convey.So(err, convey.ShouldBeNil)

			valid, err := ValidateSnapshotStatus(tmpDir)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(valid, convey.ShouldBeFalse)
		})

		convey.Convey("Should return error when unexpected directory found", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			testDir := filepath.Join(tmpDir, "testdir")
			err = os.MkdirAll(testDir, 0755)
			convey.So(err, convey.ShouldBeNil)

			testFile := filepath.Join(testDir, "test.txt")
			err = os.WriteFile(testFile, []byte("test content"), 0644)
			convey.So(err, convey.ShouldBeNil)

			err = WriteSnapshotStatus(tmpDir, SnapshotStatusSuccess, "test message")
			convey.So(err, convey.ShouldBeNil)

			extraDir := filepath.Join(tmpDir, "extradir")
			err = os.MkdirAll(extraDir, 0755)
			convey.So(err, convey.ShouldBeNil)

			extraFile := filepath.Join(extraDir, "extra.txt")
			err = os.WriteFile(extraFile, []byte("extra content"), 0644)
			convey.So(err, convey.ShouldBeNil)

			valid, err := ValidateSnapshotStatus(tmpDir)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(valid, convey.ShouldBeFalse)
		})
	})
}

func TestIsSnapshotValid(t *testing.T) {
	convey.Convey("Test IsSnapshotValid function", t, func() {
		convey.Convey("Should return false when status file does not exist", func() {
			valid := IsSnapshotValid("/non/existent/path")
			convey.So(valid, convey.ShouldBeFalse)
		})

		convey.Convey("Should return true when snapshot is valid", func() {
			tmpDir, err := os.MkdirTemp("", "snapshot-test-*")
			convey.So(err, convey.ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			testDir := filepath.Join(tmpDir, "testdir")
			err = os.MkdirAll(testDir, 0755)
			convey.So(err, convey.ShouldBeNil)

			testFile := filepath.Join(testDir, "test.txt")
			err = os.WriteFile(testFile, []byte("test content"), 0644)
			convey.So(err, convey.ShouldBeNil)

			err = WriteSnapshotStatus(tmpDir, SnapshotStatusSuccess, "test message")
			convey.So(err, convey.ShouldBeNil)

			valid := IsSnapshotValid(tmpDir)
			convey.So(valid, convey.ShouldBeTrue)
		})
	})
}

func TestSnapshotStatusStruct(t *testing.T) {
	convey.Convey("Test SnapshotStatus struct", t, func() {
		convey.Convey("Should create SnapshotStatus with correct values", func() {
			now := time.Now()
			status := SnapshotStatus{
				SHA256:          "abc123",
				DirectorySHA256: map[string]string{"dir1": "hash1"},
				Status:          SnapshotStatusSuccess,
				Timestamp:       now,
				Message:         "test message",
			}

			convey.So(status.SHA256, convey.ShouldEqual, "abc123")
			convey.So(status.DirectorySHA256["dir1"], convey.ShouldEqual, "hash1")
			convey.So(status.Status, convey.ShouldEqual, SnapshotStatusSuccess)
			convey.So(status.Timestamp, convey.ShouldEqual, now)
			convey.So(status.Message, convey.ShouldEqual, "test message")
		})
	})
}
