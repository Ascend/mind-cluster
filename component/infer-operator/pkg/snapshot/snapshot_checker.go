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
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"ascend-common/common-utils/hwlog"
	v1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

// InstanceSetTracker tracks the snapshot status of an instance set
type InstanceSetTracker struct {
	// InstanceSetName is the name of the instance set
	InstanceSetName string
	// Namespace is the namespace where the instance set resides
	Namespace string
	// SelectLabels are the labels used to select pods of the instance set
	SelectLabels map[string]string
	// StartTime is the time when snapshot tracking started
	StartTime time.Time
	// Replicas is the number of pod replicas in the workload
	Replicas int32
}

// SnapshotChecker checks and monitors the snapshot completion status of instance sets
type SnapshotChecker struct {
	client.Client
	mu               sync.RWMutex
	instanceTrackers map[string]*InstanceSetTracker
	stopCh           chan struct{}
	running          bool
	ctx              context.Context
	snapshotTimeout  time.Duration
}

// NewSnapshotChecker creates a new SnapshotChecker instance
func NewSnapshotChecker(k8sClient client.Client, timeout int) *SnapshotChecker {
	timeoutDuration := common.SnapshotTimeout
	if timeoutConfig, err := time.ParseDuration(fmt.Sprintf("%dm", min(max(timeout, 1), 600))); err == nil {
		timeoutDuration = timeoutConfig
	}
	return &SnapshotChecker{
		Client:           k8sClient,
		instanceTrackers: make(map[string]*InstanceSetTracker),
		stopCh:           make(chan struct{}),
		snapshotTimeout:  timeoutDuration,
	}
}

// Start initializes the SnapshotChecker in lazy start mode
func (sc *SnapshotChecker) Start(ctx context.Context) {
	sc.ctx = ctx
	hwlog.RunLog.Info("Snapshot checker initialized (lazy start mode)")
}

// Stop stops the SnapshotChecker and cleans up resources
func (sc *SnapshotChecker) Stop() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.running {
		close(sc.stopCh)
		sc.running = false
	}
	sc.instanceTrackers = make(map[string]*InstanceSetTracker)
	hwlog.RunLog.Info("Snapshot checker stopped")
}

// TrackInstanceSet starts tracking the specified InstanceSet
// instanceSet is the instance set to track
// selectLabels are the labels used to select pods
// podReplicas is the number of pod replicas
// Returns error if instanceSet is nil or selectLabels is empty
func (sc *SnapshotChecker) TrackInstanceSet(
	instanceSet *v1.InstanceSet,
	selectLabels map[string]string,
	podReplicas int32,
) error {
	if instanceSet == nil {
		return fmt.Errorf("instanceSet is nil")
	}

	if len(selectLabels) == 0 {
		return fmt.Errorf("selectLabels is empty")
	}

	trackerKey := sc.getInstanceSetKey(instanceSet.Namespace, instanceSet.Name)

	sc.mu.Lock()
	defer sc.mu.Unlock()

	if _, exists := sc.instanceTrackers[trackerKey]; exists {
		hwlog.RunLog.Debugf("InstanceSet %s is already being tracked", trackerKey)
		return nil
	}

	tracker := &InstanceSetTracker{
		InstanceSetName: instanceSet.Name,
		Namespace:       instanceSet.Namespace,
		SelectLabels:    selectLabels,
		StartTime:       time.Now(),
		Replicas:        podReplicas,
	}

	sc.instanceTrackers[trackerKey] = tracker
	hwlog.RunLog.Infof("Started tracking InstanceSet %s with labels %v for snapshot completion",
		trackerKey, selectLabels)

	if !sc.running {
		sc.running = true
		go sc.monitorLoop()
		hwlog.RunLog.Info("Snapshot monitor loop started (lazy initialization)")
	}

	return nil
}

func (sc *SnapshotChecker) getInstanceSetKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func (sc *SnapshotChecker) monitorLoop() {
	ticker := time.NewTicker(common.SnapshotCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sc.stopCh:
			hwlog.RunLog.Info("Snapshot monitor loop stopped")
			return
		case <-sc.ctx.Done():
			hwlog.RunLog.Info("Snapshot monitor loop stopped (context cancelled)")
			return
		case <-ticker.C:
			sc.checkAllInstanceSets()
		}
	}
}

func (sc *SnapshotChecker) checkAllInstanceSets() {
	sc.mu.RLock()
	trackerKeys := make([]string, 0, len(sc.instanceTrackers))
	for key := range sc.instanceTrackers {
		trackerKeys = append(trackerKeys, key)
	}
	sc.mu.RUnlock()

	for _, key := range trackerKeys {
		sc.checkInstanceSetSnapshot(key)
	}
}

func (sc *SnapshotChecker) checkInstanceSetSnapshot(trackerKey string) {
	sc.mu.RLock()
	tracker, exists := sc.instanceTrackers[trackerKey]
	if !exists {
		sc.mu.RUnlock()
		return
	}
	sc.mu.RUnlock()

	ctx := context.Background()

	podList := &corev1.PodList{}
	if err := sc.List(ctx, podList,
		client.InNamespace(tracker.Namespace),
		client.MatchingLabels(tracker.SelectLabels)); err != nil {
		hwlog.RunLog.Errorf("Failed to list pods for InstanceSet %s: %v", trackerKey, err)
		return
	}

	if len(podList.Items) == 0 {
		hwlog.RunLog.Debugf("No pods found yet for InstanceSet %s, waiting...", trackerKey)
		return
	}

	allFinished := true
	snapshotPods := []corev1.Pod{}
	for i := range podList.Items {
		pod := &podList.Items[i]

		finished := true
		if pod.Labels[common.InstanceIndexLabelKey] == "0" {
			// only check pods which need to save snapshot(the first P/D instance)
			finished = sc.checkPodSnapshotStatus(pod)
			snapshotPods = append(snapshotPods, *pod)
		}
		if finished {
			hwlog.RunLog.Debugf("Pod %s/%s snapshot finished", pod.Namespace, pod.Name)
		} else {
			allFinished = false
		}
	}

	sc.setAndCleanSnapshot(trackerKey, allFinished, snapshotPods, tracker, ctx, podList)
}

func (sc *SnapshotChecker) setAndCleanSnapshot(trackerKey string, allFinished bool, snapshotPods []corev1.Pod,
	tracker *InstanceSetTracker, ctx context.Context, podList *corev1.PodList) {
	if allFinished && len(snapshotPods) == int(tracker.Replicas) {
		hwlog.RunLog.Infof("All %d pods finished snapshot for InstanceSet %s", len(snapshotPods), trackerKey)
		if len(snapshotPods) == 0 {
			hwlog.RunLog.Warnf("pod replicas of InstanceSet %s is zero", trackerKey)
			return
		}
		hostSnapshotPath := GetHostSnapshotPath(&snapshotPods[0])
		if hostSnapshotPath != "" {
			if err := common.WriteSnapshotStatus(hostSnapshotPath, common.SnapshotStatusSuccess,
				"snapshot completed successfully"); err != nil {
				hwlog.RunLog.Errorf("Failed to write snapshot status for %s: %v", hostSnapshotPath, err)
			}
		}

		if err := sc.setPodsActiveLabel(ctx, snapshotPods); err != nil {
			hwlog.RunLog.Errorf("Failed to set active label for pods: %v", err)
		}

		sc.updateSnapshotCMCheckpoint(ctx, &snapshotPods[0])

		sc.removeTracker(trackerKey)
		return
	}

	// clean up snapshot paths if timeout
	if time.Since(tracker.StartTime) >= sc.snapshotTimeout {
		hwlog.RunLog.Warnf("Snapshot timeout for InstanceSet %s, cleaning up snapshot paths", trackerKey)
		for _, pod := range podList.Items {
			snapshotPath := GetHostSnapshotPath(&pod)
			if snapshotPath != "" {
				if err := common.WriteSnapshotStatus(snapshotPath, common.SnapshotStatusFailed,
					"snapshot timeout"); err != nil {
					hwlog.RunLog.Errorf("Failed to write snapshot status for %s: %v", snapshotPath, err)
				}
				if err := sc.cleanupSnapshotPath(snapshotPath); err != nil {
					hwlog.RunLog.Errorf("Failed to cleanup snapshot path %s: %v", snapshotPath, err)
				}
			}
		}
		sc.removeTracker(trackerKey)
	}
}

func (sc *SnapshotChecker) updateSnapshotCMCheckpoint(ctx context.Context, pod *corev1.Pod) {
	instanceSetName := fmt.Sprintf("%s-%s",
		common.GetInstanceSetNameFromLabels(pod.Labels), pod.Labels[common.InstanceIndexLabelKey])
	cmName := common.SnapshotMetadataPrefix + instanceSetName

	cm := &corev1.ConfigMap{}
	var err error

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Millisecond * 1000 * time.Duration(attempt))
		}

		err = sc.Get(ctx, client.ObjectKey{Namespace: pod.Namespace, Name: cmName}, cm)
		if err == nil {
			break
		}
	}

	if err != nil {
		if apierrors.IsNotFound(err) {
			hwlog.RunLog.Errorf("ConfigMap %s/%s not found after retries, it may not be created yet or cache not synced",
				pod.Namespace, cmName)
		} else {
			hwlog.RunLog.Errorf("Failed to get configmap %s/%s after retries: %v", pod.Namespace, cmName, err)
		}
		return
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}

	data, exists := cm.Data[common.SnapshotMetadataJson]
	if data == "" || !exists {
		hwlog.RunLog.Error("configmap content error")
		return
	}

	var checkpointData common.SnapshotMetaData
	if err := json.Unmarshal([]byte(data), &checkpointData); err != nil {
		hwlog.RunLog.Errorf("Failed to unmarshal existing snapshot metadata: %v, will create new one", err)
		return
	}

	checkpointData.Checkpoint = common.SnapshotFinished

	checkpointBytes, err := json.Marshal(checkpointData)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to marshal checkpoint data: %v", err)
		return
	}

	cm.Data[common.SnapshotMetadataJson] = string(checkpointBytes)

	if err := sc.Update(ctx, cm); err != nil {
		hwlog.RunLog.Errorf("Failed to update configmap %s/%s: %v", pod.Namespace, cmName, err)
		return
	}

	hwlog.RunLog.Infof("Updated snapshot configmap %s/%s, set checkpoint to done", pod.Namespace, cmName)
}

func (sc *SnapshotChecker) checkPodSnapshotStatus(pod *corev1.Pod) bool {
	if pod.Annotations == nil {
		return false
	}

	annoValue, exists := pod.Annotations[common.HostSnapshotFlagAnnotationKey]
	if !exists {
		return false
	}
	return annoValue == common.TrueBool
}

func (sc *SnapshotChecker) cleanupSnapshotPath(snapshotPath string) error {
	if snapshotPath == "" {
		return fmt.Errorf("snapshot path is empty")
	}

	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		hwlog.RunLog.Infof("Snapshot path %s does not exist, skip cleanup", snapshotPath)
		return nil
	}

	entries, err := os.ReadDir(snapshotPath)
	if err != nil {
		return fmt.Errorf("failed to read snapshot path %s: %v", snapshotPath, err)
	}

	for _, entry := range entries {
		if entry.Name() == common.SnapshotStatusFileName {
			continue
		}

		entryPath := filepath.Join(snapshotPath, entry.Name())
		if err := os.RemoveAll(entryPath); err != nil {
			hwlog.RunLog.Warnf("Failed to remove %s: %v", entryPath, err)
			continue
		}
		hwlog.RunLog.Debugf("Removed %s from snapshot path", entryPath)
	}

	hwlog.RunLog.Infof("Successfully cleaned up snapshot path: %s (preserved status file)", snapshotPath)
	return nil
}

func (sc *SnapshotChecker) removeTracker(trackerKey string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	delete(sc.instanceTrackers, trackerKey)
	hwlog.RunLog.Infof("Removed tracker for InstanceSet %s", trackerKey)
}

// GetTrackerCount returns the number of tracked instance sets
// Returns the count of currently tracked instance sets
func (sc *SnapshotChecker) GetTrackerCount() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return len(sc.instanceTrackers)
}

// IsRunning checks if the SnapshotChecker is running
// Returns true if the checker is currently running
func (sc *SnapshotChecker) IsRunning() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.running
}

func (sc *SnapshotChecker) updatePodReadinessGate(ctx context.Context, pod *corev1.Pod, ready bool) error {
	updatedPod := pod.DeepCopy()

	conditionStatus := corev1.ConditionFalse
	if ready {
		conditionStatus = corev1.ConditionTrue
	}

	condition := corev1.PodCondition{
		Type:               common.PodSnapshotReadyConditionType,
		Status:             conditionStatus,
		LastProbeTime:      metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             "SnapshotCompleted",
		Message:            "Container snapshot has been completed successfully",
	}

	conditionExists := false
	for i, c := range updatedPod.Status.Conditions {
		if c.Type == common.PodSnapshotReadyConditionType {
			updatedPod.Status.Conditions[i] = condition
			conditionExists = true
			break
		}
	}

	if !conditionExists {
		updatedPod.Status.Conditions = append(updatedPod.Status.Conditions, condition)
	}

	if err := sc.Status().Patch(ctx, updatedPod, client.MergeFrom(pod)); err != nil {
		return fmt.Errorf("failed to patch pod status: %v", err)
	}

	hwlog.RunLog.Infof("Updated readiness gate for pod %s/%s to %v", pod.Namespace, pod.Name, ready)
	return nil
}

func (sc *SnapshotChecker) setPodsActiveLabel(ctx context.Context, pods []corev1.Pod) error {
	for i := range pods {
		pod := &pods[i]
		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}
		if pod.Labels[common.ActiveLabelKey] == common.TrueBool {
			continue
		}

		updatedPod := pod.DeepCopy()
		updatedPod.Labels[common.ActiveLabelKey] = common.TrueBool

		if err := sc.Patch(ctx, updatedPod, client.MergeFrom(pod)); err != nil {
			hwlog.RunLog.Errorf("Failed to patch pod %s/%s with active label: %v",
				pod.Namespace, pod.Name, err)
			continue
		}
		hwlog.RunLog.Infof("Set active label for pod %s/%s", pod.Namespace, pod.Name)
	}
	return nil
}
