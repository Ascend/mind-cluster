/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package snapshot for the pod monitor test
package snapshot

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// TestNewPodMonitor tests the NewPodMonitor function
func TestNewPodMonitor(t *testing.T) {
	client := &kubeclient.ClientK8s{}
	monitor := NewPodMonitor(client)
	assert.NotNil(t, monitor)
	assert.Equal(t, client, monitor.client)
	assert.NotNil(t, monitor.stopChan)
}

// TestPodMonitor_Stop tests the PodMonitor.Stop method
func TestPodMonitor_Stop(t *testing.T) {
	monitor := &PodMonitor{
		stopChan: make(chan struct{}, 1),
	}
	monitor.Stop()
	select {
	case <-monitor.stopChan:
	default:
		t.Error("Stop method did not send signal to stopChan")
	}
}

// TestPodMonitor_AddPod tests the PodMonitor.AddPod method
func TestPodMonitor_AddPod(t *testing.T) {
	monitor := &PodMonitor{}
	// Test nil PodMonitor
	monitor.AddPod(nil)
	// Test non-Pod object
	monitor.AddPod("not a pod")
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-0",
			Namespace: "default",
			Annotations: map[string]string{
				common.SnapshotFinishFlag: common.Finished,
			},
		},
	}
	// Patch the containerSnapshot method
	patches := gomonkey.ApplyPrivateMethod(&PodMonitor{}, "containerSnapshot", func(pod *v1.Pod, podKey string) {
		return
	})
	defer patches.Reset()
	monitor.AddPod(pod)
}

// TestPodMonitor_UpdatePod tests the PodMonitor.UpdatePod method
func TestPodMonitor_UpdatePod(t *testing.T) {
	monitor := &PodMonitor{}
	// Test nil PodMonitor
	monitor.UpdatePod(nil, nil)
	// Test non-Pod object
	monitor.UpdatePod(nil, "not a pod")
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-0",
			Namespace: "default",
		},
	}
	// Patch the containerSnapshot method
	patches := gomonkey.ApplyPrivateMethod(&PodMonitor{}, "containerSnapshot", func(pod *v1.Pod, podKey string) {
		return
	})
	defer patches.Reset()
	monitor.UpdatePod(nil, pod)
}

// TestPodMonitor_containerSnapshot tests the PodMonitor.containerSnapshot method
func TestPodMonitor_containerSnapshot(t *testing.T) {
	monitor := &PodMonitor{}
	// Test pod that has already been processed
	podWithAnnotation := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-0",
			Namespace: "default",
			Annotations: map[string]string{
				common.SnapshotFinishFlag: common.Finished,
				common.SnapshotMode:       "save",
			},
		},
	}
	monitor.containerSnapshot(podWithAnnotation, "default/test-pod-0")
	// Test pod that is not ready
	podNotReady := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-0",
			Namespace: "default",
			Annotations: map[string]string{
				common.SnapshotMode: "save",
			},
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionFalse,
				},
			},
		},
	}
	monitor.containerSnapshot(podNotReady, "default/test-pod-0")
	// Test pod with multiple containers
	podWithMultipleContainers := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-0",
			Namespace: "default",
			Annotations: map[string]string{
				common.SnapshotMode: "save",
			},
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: "container1"},
				{Name: "container2"},
			},
		},
	}
	monitor.containerSnapshot(podWithMultipleContainers, "default/test-pod-0")
	// Test pod without snapshot path
	podWithoutSnapshotPath := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-0",
			Namespace: "default",
			Annotations: map[string]string{
				common.SnapshotMode: "save",
			},
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "container1",
					Env:  []v1.EnvVar{},
				},
			},
		},
	}
	monitor.containerSnapshot(podWithoutSnapshotPath, "default/test-pod-0")
	// Test normal pod
	normalPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-0",
			Namespace: "default",
			Annotations: map[string]string{
				common.SnapshotMode: "save",
			},
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
			ContainerStatuses: []v1.ContainerStatus{
				{
					ContainerID: "container-id",
				},
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "container1",
					Env: []v1.EnvVar{
						{
							Name:  common.SnapshotPath,
							Value: "/snapshot",
						},
					},
				},
			},
		},
	}
	patches := gomonkey.NewPatches()
	patches.ApplyFunc(utils.GetLastNumberFromString, func(s string) (string, error) {
		return "0", nil
	})
	patches.ApplyPrivateMethod(&PodMonitor{}, "checkpoint", func(pod *v1.Pod, podKey string, containerId string, snapshotPath string) {
		return
	})
	defer patches.Reset()
	monitor.containerSnapshot(normalPod, "default/test-pod-0")
}

// TestPodMonitor_checkpoint tests the PodMonitor.checkpoint method
func TestPodMonitor_checkpoint(t *testing.T) {
	client := &kubeclient.ClientK8s{}
	monitor := &PodMonitor{
		client: client,
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-0",
			Namespace: "default",
		},
	}
	// Patch doCheckpoint and UpdatePodAnnotation methods
	patches := gomonkey.ApplyFunc(doCheckpoint, func(string, string, int) error {
		return nil
	})
	patches.ApplyMethod(&kubeclient.ClientK8s{}, "UpdatePodAnnotation", func(*kubeclient.ClientK8s, string, string, *v1.Pod) error {
		return nil
	})
	defer patches.Reset()
	// Test normal case
	monitor.checkpoint(pod, "default/test-pod-0", "container-id", "/snapshot/0")
	// Test doCheckpoint failure case
	patches = gomonkey.ApplyFunc(doCheckpoint, func(string, string, int) error {
		return errors.New("checkpoint failed")
	})
	defer patches.Reset()
	monitor.checkpoint(pod, "default/test-pod-0", "container-id", "/snapshot/0")
	// Test UpdatePodAnnotation failure case
	patches = gomonkey.ApplyFunc(doCheckpoint, func(string, string, int) error {
		return nil
	})
	patches.ApplyMethod(&kubeclient.ClientK8s{}, "UpdatePodAnnotation", func(*kubeclient.ClientK8s, string, string, *v1.Pod) error {
		return errors.New("update annotation failed")
	})
	defer patches.Reset()
	monitor.checkpoint(pod, "default/test-pod-0", "container-id", "/snapshot/0")
}

// TestDoCheckpoint tests the doCheckpoint function
func TestDoCheckpoint(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	// Test case where directory already exists
	tempDir, err := os.MkdirTemp("", "test-snapshot")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)
	err = doCheckpoint("container-id", tempDir, -1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	// Test case where ascend-docker-runtime does not exist
	patches.ApplyFunc(exec.LookPath, func(string) (string, error) {
		return "", errors.New("not found")
	})
	err = doCheckpoint("container-id", filepath.Join(tempDir, "new"), -1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ascend-docker-runtime not found")
	// Test case where RunCmd fails
	patches = gomonkey.ApplyFunc(exec.LookPath, func(string) (string, error) {
		return "/bin/echo", nil
	})
	patches.ApplyFunc(RunCmd, func(string, []string, []string, int) (int, error) {
		return 1, errors.New("run cmd failed")
	})
	defer patches.Reset()
	err = doCheckpoint("container-id", filepath.Join(tempDir, "new"), -1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "run cmd failed")
	// Test case where RunCmd succeeds but exit code is non-zero
	patches = gomonkey.ApplyFunc(RunCmd, func(string, []string, []string, int) (int, error) {
		return 1, nil
	})
	defer patches.Reset()
	err = doCheckpoint("container-id", filepath.Join(tempDir, "new"), -1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checkpoint failed with exit code")
	// Test case where RunCmd succeeds and exit code is zero
	patches = gomonkey.ApplyFunc(RunCmd, func(string, []string, []string, int) (int, error) {
		return 0, nil
	})
	defer patches.Reset()
	err = doCheckpoint("container-id", filepath.Join(tempDir, "new"), -1)
	assert.NoError(t, err)
	// Test timeout case
	patches = gomonkey.ApplyFunc(RunCmd, func(string, []string, []string, int) (int, error) {
		return 0, ErrCommandTimeout
	})
	patches.ApplyFunc(tryResumeContainer, func(string, string) error {
		return nil
	})
	defer patches.Reset()
	err = doCheckpoint("container-id", filepath.Join(tempDir, "new"), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checkpoint timeout")
	// Test timeout and resume failure case
	patches = gomonkey.ApplyFunc(tryResumeContainer, func(string, string) error {
		return errors.New("resume failed")
	})
	defer patches.Reset()
	err = doCheckpoint("container-id", filepath.Join(tempDir, "new"), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resume container after process termination")
}

// TestTryResumeContainer tests the tryResumeContainer function
func TestTryResumeContainer(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	// Test case where ascend-docker-runtime does not exist
	patches.ApplyFunc(exec.LookPath, func(string) (string, error) {
		return "", errors.New("not found")
	})
	err := tryResumeContainer("container-id", "/snapshot")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ascend-docker-runtime not found for resume")
	// Test case where resume succeeds
	patches = gomonkey.ApplyFunc(exec.LookPath, func(string) (string, error) {
		return "/bin/echo", nil
	})
	patches.ApplyFunc((*exec.Cmd).CombinedOutput, func(*exec.Cmd) ([]byte, error) {
		return []byte("success"), nil
	})
	defer patches.Reset()
	err = tryResumeContainer("container-id", "/snapshot")
	assert.NoError(t, err)
	// Test case where container is not paused
	patches = gomonkey.ApplyFunc((*exec.Cmd).CombinedOutput, func(*exec.Cmd) ([]byte, error) {
		return []byte("container not paused"), errors.New("error")
	})
	defer patches.Reset()
	err = tryResumeContainer("container-id", "/snapshot")
	assert.NoError(t, err)
	// Test case where resume fails
	patches = gomonkey.ApplyFunc((*exec.Cmd).CombinedOutput, func(*exec.Cmd) ([]byte, error) {
		return []byte("failed"), errors.New("error")
	})
	defer patches.Reset()
	err = tryResumeContainer("container-id", "/snapshot")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "runc resume failed")
}
