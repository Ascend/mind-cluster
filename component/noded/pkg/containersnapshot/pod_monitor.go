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

// Package snapshot for the pod monitor
package snapshot

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
)

var semp = newSemaphore(common.MaxCheckpointRequest)

// PodMonitor manage pod monitoring information
type PodMonitor struct {
	client   *kubeclient.ClientK8s
	stopChan chan struct{}
}

// NewPodMonitor create a pod monitor
func NewPodMonitor(client *kubeclient.ClientK8s) *PodMonitor {
	return &PodMonitor{
		client:   client,
		stopChan: make(chan struct{}, 1),
	}
}

// Monitoring start working loop
func (p *PodMonitor) Monitoring() {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(p.client.ClientSet, 0,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = labels.SelectorFromSet(labels.Set{common.InferLabel: "true"}).String()
			options.FieldSelector = "spec.nodeName=" + p.client.NodeName
		}))
	podInformer := informerFactory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    p.AddPod,
			UpdateFunc: p.UpdatePod,
		})
	informerFactory.Start(wait.NeverStop)
}

// Stop terminate working loop
func (p *PodMonitor) Stop() {
	p.stopChan <- struct{}{}
}

// AddPod handle pod add event
func (p *PodMonitor) AddPod(obj interface{}) {
	if p == nil {
		hwlog.RunLog.Error("pod monitor is nil when add pod")
		return
	}
	pod, ok := obj.(*v1.Pod)
	if !ok {
		hwlog.RunLog.Error("failed convert pod when add pod")
		return
	}
	podKey := pod.Namespace + "/" + pod.Name
	hwlog.RunLog.Infof("start to handle pod: %s", podKey)

	p.containerSnapshot(pod, podKey)
}

// UpdatePod handle pod update event
func (p *PodMonitor) UpdatePod(_, new interface{}) {
	if p == nil {
		hwlog.RunLog.Error("pod monitor is nil when update pod")
		return
	}
	pod, ok := new.(*v1.Pod)
	if !ok {
		hwlog.RunLog.Error("failed convert pod when update pod")
		return
	}
	podKey := pod.Namespace + "/" + pod.Name
	hwlog.RunLog.Infof("pod updated: %s", podKey)

	p.containerSnapshot(pod, podKey)
}

func (p *PodMonitor) containerSnapshot(pod *v1.Pod, podKey string) {
	snapshotMode, _ := pod.Annotations[common.SnapshotMode]
	if snapshotMode != common.SnapshotSaveMode {
		hwlog.RunLog.Infof("container snapshot mode is not checkpoint: %s", snapshotMode)
		return
	}
	annotation, exist := pod.Annotations[common.SnapshotFinishFlag]
	if exist {
		hwlog.RunLog.Infof("container snapshot has been processed: %s", annotation)
		return
	}
	podReady := v1.ConditionUnknown
	for _, cond := range pod.Status.Conditions {
		if cond.Type == v1.PodReady {
			podReady = cond.Status
			break
		}
	}
	if podReady != v1.ConditionTrue {
		hwlog.RunLog.Infof("pod: %s is not ready", podKey)
		return
	}
	containers := pod.Spec.Containers
	if len(containers) != 1 {
		hwlog.RunLog.Infof("the container number of pod %s is not 1", podKey)
		return
	}
	snapshot := false
	var snapshotPath string
	for _, env := range containers[0].Env {
		if env.Name == common.SnapshotPath && env.Value != "" {
			snapshot = true
			snapshotPath = env.Value
		}
	}
	if !snapshot {
		hwlog.RunLog.Info("the container has no snapshot save path env")
		return
	}
	var containerId string
	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerId = getContainerID(containerStatus.ContainerID)
	}
	podIndex, err := utils.GetLastNumberFromString(pod.Name)
	if err != nil {
		hwlog.RunLog.Errorf("get pod index from pod name failed: %v", err)
		return
	}
	p.checkpoint(pod, podKey, containerId, filepath.Join(snapshotPath, podIndex))
}

func getContainerID(conID string) string {
	parts := strings.SplitN(conID, "://", 2)
	if len(parts) != 2 {
		return conID
	}
	return parts[1]
}

func (p *PodMonitor) checkpoint(pod *v1.Pod, podKey string, containerId string, snapshotPath string) {
	var value string
	for {
		err := semp.acquire(containerId)
		if err == nil {
			break
		}
		hwlog.RunLog.Errorf("acquire token failed: %v, wait and try again", err)
		time.Sleep(time.Minute)
	}
	defer semp.release(containerId)
	err := doCheckpoint(containerId, snapshotPath, -1)
	if err != nil {
		hwlog.RunLog.Errorf("checkpoint %s failed: %v", containerId, err)
		value = common.Failed
	} else {
		hwlog.RunLog.Infof("checkpoint %s finished successfully", containerId)
		value = common.Finished
	}
	err = p.client.UpdatePodAnnotation(common.SnapshotFinishFlag, value, pod)
	if err != nil {
		hwlog.RunLog.Errorf("update pod %s annotation failed: %v", podKey, err)
		return
	}
	hwlog.RunLog.Infof("update pod %s annotation success", podKey)
	return
}

func doCheckpoint(conID, localDir string, timeout int) error {
	if _, err := os.Stat(localDir); err == nil {
		hwlog.RunLog.Errorf("location: %s already exists", localDir)
		if files, _ := os.ReadDir(localDir); len(files) != 0 {
			hwlog.RunLog.Errorf("location: %s snapshot already exists, checkpoint already finished, skip", localDir)
			return nil
		}
		return fmt.Errorf("location: %s already exists", localDir)
	}

	lp, err := exec.LookPath(common.AscendDockerRuntimePath)
	if err != nil {
		return fmt.Errorf("ascend-docker-runtime not found: %v", err)
	}

	args := []string{"checkpoint", "--image-path", localDir, conID}

	exitCode, runErr := RunCmd(lp, args, os.Environ(), timeout)
	if runErr == nil {
		if exitCode == 0 {
			return nil
		}
		msg := fmt.Sprintf("checkpoint failed with exit code %d", exitCode)
		hwlog.RunLog.Errorf("checkpoint %s %s", conID, msg)
		return errors.New(msg)
	}
	if timeout <= 0 {
		return runErr
	}

	if !strings.Contains(runErr.Error(), "timeout") {
		return runErr
	}

	resumeErr := tryResumeContainer(conID, localDir)

	if errors.Is(runErr, ErrCommandTimeout) {
		hwlog.RunLog.Errorf("checkpoint %s timeout, return err: %v, %v", conID, resumeErr, runErr)
		if resumeErr == nil {
			return errors.New("checkpoint timeout")
		}
		return errors.New("checkpoint timeout: failed to resume container after process termination")
	}
	hwlog.RunLog.Errorf("checkpoint %s timeout, return err: %v, %v", conID, resumeErr, runErr)
	if resumeErr == nil {
		return errors.New("checkpoint timeout and kill defeated: container resume succeed")
	}
	return errors.New("checkpoint timeout and kill defeated: resume container failed")
}

func tryResumeContainer(conID string, locDir string) error {
	lp, err := exec.LookPath(common.AscendDockerRuntimePath)
	if err != nil {
		return fmt.Errorf("ascend-docker-runtime not found for resume: %v", err)
	}
	cmd := exec.Command(lp, "resume", conID)
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	output, err := cmd.CombinedOutput()
	if err == nil {
		hwlog.RunLog.Infof("successfully resumed container %s", conID)
		return nil
	}

	// if error is "container not paused", treated as successfully
	if strings.Contains(string(output), "container not paused") {
		hwlog.RunLog.Infof("container %s is not paused, skip resume (treated as success)", conID)
		return nil
	}

	hwlog.RunLog.Errorf("failed to resume container %s: %v, output: %s", conID, err, string(output))
	return fmt.Errorf("runc resume failed: %v", err)
}
