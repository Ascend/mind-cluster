/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package utils is using for generating ranktable.
*/

package utils

import (
	"os"
	"strings"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

// GenRankTableDir generate rank table dir
func GenRankTableDir(job *mindxdlv1.AscendJob) string {
	for _, replSpec := range job.Spec.ReplicaSpecs {
		for _, volume := range replSpec.Template.Spec.Volumes {
			if volume.Name != rankTableName || volume.VolumeSource.HostPath == nil {
				continue
			}
			return volume.VolumeSource.HostPath.Path
		}
	}
	hwlog.RunLog.Info("ranktable file path is not set")
	return ""
}

// PodHasAllocated check if pod has allocated device
func PodHasAllocated(pod *corev1.Pod) bool {
	if pod.GetDeletionTimestamp() != nil {
		return false
	}
	if !podUseNpu(pod) {
		return true
	}
	if _, ok := pod.Annotations[PodDeviceKey]; !ok {
		hwlog.RunLog.Debugf("Pod %s has not allocated device", pod.Name)
		return false
	}
	return true
}

func podUseNpu(pod *corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		for resName, resVal := range container.Resources.Requests {
			resValNum, ok := resVal.AsInt64()
			if !ok {
				continue
			}
			if strings.Contains(string(resName), npuPrefix) && resValNum > 0 {
				return true
			}
		}
	}
	hwlog.RunLog.Infof("pod %v not use npu", pod.Name)
	return false
}

// check exsitence and permission of ranktable directory, if not exist, try to create the directory
func CheckDirPath(dirPath string) error {
	if dirPath == "" {
		return nil
	}
	s, err := os.Stat(dirPath)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		if s.IsDir() {
			hwlog.RunLog.Info("try to mkdir")
			return os.MkdirAll(dirPath, os.ModePerm)
		}
		return err
	}
	hwlog.RunLog.Errorf("directory path exists but has no permission : %v", err)
	return err
}

const (
	// PodDeviceKey Pod annoation Key
	PodDeviceKey = "ascend.kubectl.kubernetes.io/ascend-910-configuration"
	// PodRankKey Pod annoation Key
	PodRankKey = "hccl/rankIndex"

	rankTableDir = "/user/mindx-dl/ranktable"

	// prefix of request npu name
	npuPrefix = "huawei.com/"

	// rank table volume name
	rankTableName = "ranktable"
)

// RankTableStatus is rank table status
type RankTableStatus string

const (
	// InitialRTStatus initial rank table status
	InitialRTStatus RankTableStatus = "initializing"
	// CompletedRTStatus completed rank table status
	CompletedRTStatus RankTableStatus = "completed"
)

// check configmap exsit or not
type ConfigmapCheck string

const (
	// ConfigmapExsit configmap exist
	ConfigmapExsit ConfigmapCheck = "configmapExist"
	// ConfigmapNotExist configmap not exist
	ConfigmapNotExist ConfigmapCheck = "configmapNotExist"
)
