/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package utils is using for generating ranktable.
*/
package utils

import (
	"io/fs"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	mindxdlv1 "ascend-operator/pkg/api/v1"
)

const (
	defaultDirPerm fs.FileMode = 0744
)

// GenRankTableDir generate rank table dir
func GenRankTableDir(job *mindxdlv1.AscendJob) string {
	ranktableDir := readRankTableDir(job)
	if ranktableDir == "" {
		return ranktableDir
	}
	checkedPath, err := utils.PathStringChecker(ranktableDir)
	if err != nil {
		hwlog.RunLog.Errorf("rank table directory is invalid, err: %v", err)
		return ""
	}
	if utils.IsExist(ranktableDir) {
		isSoftlink, err := utils.IsSoftlink(checkedPath)
		if err != nil {
			hwlog.RunLog.Errorf("rank table directory existed but is invalid, err: %v", err)
			return ""
		}
		if isSoftlink {
			hwlog.RunLog.Error("rank table directory existed but is softlink")
			return ""
		}
		return checkedPath
	}
	if err := os.MkdirAll(checkedPath, defaultDirPerm); err != nil {
		hwlog.RunLog.Errorf("failed to create directory, err: %v", err)
		return ""
	}
	hwlog.RunLog.Infof("create rank table directory success, set mode to %v", defaultDirPerm)
	return checkedPath
}

func readRankTableDir(job *mindxdlv1.AscendJob) string {
	for _, replSpec := range job.Spec.ReplicaSpecs {
		for _, volume := range replSpec.Template.Spec.Volumes {
			if volume.Name != rankTableName {
				continue
			}
			return getVolumePath(volume)
		}
	}
	hwlog.RunLog.Info("ranktable file path is not set")
	return ""
}

func getVolumePath(volume corev1.Volume) string {
	switch {
	case volume.VolumeSource.HostPath != nil:
		return volume.VolumeSource.HostPath.Path
	case volume.VolumeSource.NFS != nil:
		return volume.VolumeSource.NFS.Path
	default:
		hwlog.RunLog.Error("please mount rank table by nfs or hostPath")
		return ""
	}
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
			if strings.Contains(string(resName), api.ResourceNamePrefix) && resValNum > 0 {
				return true
			}
		}
	}
	hwlog.RunLog.Infof("pod %v not use npu", pod.Name)
	return false
}

const (
	// PodDeviceKey Pod annoation Key
	PodDeviceKey = "ascend.kubectl.kubernetes.io/ascend-910-configuration"
	// PodRankKey Pod annoation Key
	PodRankKey = "hccl/rankIndex"

	rankTableDir = "/user/mindx-dl/ranktable"

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
	// UnknownStatus unknown rank table status
	UnknownStatus = "unknown"
)

// ConfigmapCheck check configmap exsit or not
type ConfigmapCheck string

const (
	// ConfigmapExsit configmap exist
	ConfigmapExsit ConfigmapCheck = "configmapExist"
	// ConfigmapNotExist configmap not exist
	ConfigmapNotExist ConfigmapCheck = "configmapNotExist"
)
