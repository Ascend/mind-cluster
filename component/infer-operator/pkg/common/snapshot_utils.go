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

// Package common for snapshot functions
package common

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	v1 "infer-operator/pkg/api/v1"
)

// SnapshotStatus represents the status of a snapshot operation
type SnapshotStatus struct {
	// SHA256 is the SHA256 hash of the snapshot
	SHA256 string `json:"sha256,omitempty"`
	// DirectorySHA256 maps directory names to their SHA256 hashes
	DirectorySHA256 map[string]string `json:"directorySHA256"`
	// Status is the current snapshot status
	Status string `json:"status"`
	// Timestamp is when the snapshot status was recorded
	Timestamp time.Time `json:"timestamp"`
	// Message is an optional status message
	Message string `json:"message,omitempty"`
}

// SnapshotMetaData contains metadata about a snapshot
type SnapshotMetaData struct {
	// InstanceName is the instance/job name
	InstanceName string `json:"job_name,omitempty"`
	// Namespace is the instance namespace
	Namespace string `json:"namespace,omitempty"`
}

// AddSnapshotInfoToPodTemplate adds snapshot info to pod template
func AddSnapshotInfoToPodTemplate(pod *corev1.PodTemplateSpec, instanceSet *v1.InstanceSet, cmName string) {
	if !IsContainerSnapshotOn(instanceSet) {
		return
	}

	hostSnapshotPath := GetHostSnapshotPathFromPodTemplate(pod, instanceSet)
	if hostSnapshotPath == "" {
		return
	}

	AddSnapshotEnv(pod, hostSnapshotPath, cmName)
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[ContainerSnapshotLabelKey] = TrueBool

}

// AddMetadataVolume adds metadata volume to pod template
func AddMetadataVolume(pod *corev1.PodTemplateSpec, cmName string, instanceSet *v1.InstanceSet) {
	if !IsContainerSnapshotOn(instanceSet) {
		return
	}
	metadataVolume := corev1.Volume{
		Name: "metadata",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cmName,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  "snapshot_metadata.json",
						Path: "snapshot_metadata.json",
					},
				},
			},
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, metadataVolume)
	metadataVolumeMounts := corev1.VolumeMount{
		Name:      "metadata",
		MountPath: "/snapshot/configmap",
		ReadOnly:  true,
	}
	if len(pod.Spec.Containers) <= 0 {
		hwlog.RunLog.Errorf("Pod has no containers, cannot add volume mounts")
		return
	}
	pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, metadataVolumeMounts)
}

// AddSnapshotEnv adds snapshot env vars to pod containers
func AddSnapshotEnv(pod *corev1.PodTemplateSpec, hostSnapshotPath, cmName string) {
	for index := range pod.Spec.Containers {
		pod.Spec.Containers[index].Env = append(pod.Spec.Containers[index].Env,
			corev1.EnvVar{
				Name: GrusSnapshotRestoredFlag,
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cmName,
						},
						Key: "GrusSnapshotRestoredFlag",
					},
				},
			})
		pod.Spec.Containers[index].Env = append(pod.Spec.Containers[index].Env,
			corev1.EnvVar{
				Name:  HostSnapshotPathEnvKey,
				Value: hostSnapshotPath,
			})
		pod.Spec.Containers[index].Env = append(pod.Spec.Containers[index].Env,
			corev1.EnvVar{
				Name: PodNameEnvKey,
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			})
	}
}

// GetHostSnapshotPathFromPodTemplate returns the host snapshot path for a pod
func GetHostSnapshotPathFromPodTemplate(pod *corev1.PodTemplateSpec, instanceSet *v1.InstanceSet) string {
	snapshotPath := ""
	if len(pod.Spec.Containers) <= 0 {
		hwlog.RunLog.Errorf("Pod has no containers, cannot get host snapshot path")
		return ""
	}
	for _, env := range pod.Spec.Containers[0].Env {
		if env.Name == HostSnapshotDirPathEnvKey && env.Value != "" {
			snapshotPath = env.Value
			break
		}
	}
	if snapshotPath == "" {
		hwlog.RunLog.Errorf("Host snapshot path env '%s' not found in pod %s/%s, "+
			"cannot get snapshot path", HostSnapshotDirPathEnvKey, pod.Namespace, pod.Name)
		return ""
	}

	snapshotPath = filepath.Join(snapshotPath, instanceSet.Namespace, instanceSet.Name)
	return snapshotPath
}

// GetSnapshotStatusFilePath returns the path to the snapshot status file
func GetSnapshotStatusFilePath(snapshotPath string) string {
	return filepath.Join(snapshotPath, SnapshotStatusFileName)
}

// CalculateSnapshotSHA256 calculates SHA256 hashes for snapshot directories
func CalculateSnapshotSHA256(snapshotPath string) (map[string]string, error) {
	result := make(map[string]string)
	entries, err := os.ReadDir(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot directory %s: %v", snapshotPath, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		if _, exists := result[dirName]; exists {
			hwlog.RunLog.Debugf("Skip SHA256 calculation for directory %s, already cached", dirName)
			continue
		}

		dirPath := filepath.Join(snapshotPath, dirName)
		hash := sha256.New()
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %v", dirPath, err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			filePath := filepath.Join(dirPath, entry.Name())
			file, err := os.Open(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to open file %s: %v", filePath, err)
			}
			if _, err := io.Copy(hash, file); err != nil {
				file.Close()
				return nil, fmt.Errorf("failed to hash file %s: %v", filePath, err)
			}
			file.Close()
		}

		if err != nil {
			return nil, fmt.Errorf("failed to calculate SHA256 for directory %s: %v", dirPath, err)
		}
		result[dirName] = hex.EncodeToString(hash.Sum(nil))
	}

	return result, nil
}

// WriteSnapshotStatus writes snapshot status to file
func WriteSnapshotStatus(snapshotPath string, status string, message string) error {
	statusFilePath := GetSnapshotStatusFilePath(snapshotPath)
	directorySHA256, err := CalculateSnapshotSHA256(snapshotPath)
	if err != nil {
		msg := fmt.Errorf("Failed to calculate SHA256 for snapshot %s: %v", snapshotPath, err)
		hwlog.RunLog.Error(msg)
		return msg
	}

	snapshotStatus := SnapshotStatus{
		DirectorySHA256: directorySHA256,
		Status:          status,
		Timestamp:       time.Now(),
		Message:         message,
	}

	statusBytes, err := json.MarshalIndent(snapshotStatus, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot status: %v", err)
	}

	if err := os.WriteFile(statusFilePath, statusBytes, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot status file: %v", err)
	}

	hwlog.RunLog.Infof("Wrote snapshot status to %s: status=%s, directorySHA256=%v", statusFilePath, status, directorySHA256)
	return nil
}

// ReadSnapshotStatus reads snapshot status from file
func ReadSnapshotStatus(snapshotPath string) (*SnapshotStatus, error) {
	statusFilePath := GetSnapshotStatusFilePath(snapshotPath)

	data, err := os.ReadFile(statusFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read snapshot status file: %v", err)
	}

	var status SnapshotStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot status: %v", err)
	}

	return &status, nil
}

// IsSnapshotStatusExists checks if snapshot status file exists
func IsSnapshotStatusExists(snapshotPath string) bool {
	statusFilePath := GetSnapshotStatusFilePath(snapshotPath)
	_, err := os.Stat(statusFilePath)
	return !os.IsNotExist(err)
}

// ValidateSnapshotStatus validates the snapshot status and SHA256 hashes
func ValidateSnapshotStatus(snapshotPath string) (bool, error) {
	status, err := ReadSnapshotStatus(snapshotPath)
	if err != nil {
		return false, err
	}

	if status == nil {
		return false, fmt.Errorf("snapshot status is nil")
	}

	if status.Status != SnapshotStatusSuccess {
		return false, fmt.Errorf("snapshot status is %s: %s", status.Status, status.Message)
	}

	if len(status.DirectorySHA256) == 0 {
		return false, fmt.Errorf("snapshot status has empty DirectorySHA256")
	}

	currentDirectorySHA256, err := CalculateSnapshotSHA256(snapshotPath)
	if err != nil {
		hwlog.RunLog.Warnf("Failed to calculate current SHA256: %v", err)
		return false, err
	}

	for dirName, expectedHash := range status.DirectorySHA256 {
		currentHash, exists := currentDirectorySHA256[dirName]
		if !exists {
			return false, fmt.Errorf("directory %s not found in current snapshot", dirName)
		}
		if currentHash != expectedHash {
			return false, fmt.Errorf("SHA256 mismatch for directory %s: expected %s, got %s", dirName, expectedHash, currentHash)
		}
	}

	for dirName := range currentDirectorySHA256 {
		if _, exists := status.DirectorySHA256[dirName]; !exists {
			return false, fmt.Errorf("unexpected directory %s found in snapshot", dirName)
		}
	}

	return true, nil
}

// IsSnapshotValid checks if a snapshot is valid
func IsSnapshotValid(snapshotPath string) bool {
	valid, err := ValidateSnapshotStatus(snapshotPath)
	if err != nil {
		hwlog.RunLog.Warnf("Snapshot validation failed for %s: %v", snapshotPath, err)
		return false
	}
	return valid
}
