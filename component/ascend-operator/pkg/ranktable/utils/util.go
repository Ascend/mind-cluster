/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package utils is using for generating ranktable.
*/

package utils

import (
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

// GenRankTableDir generate rank table dir
func GenRankTableDir(job *mindxdlv1.AscendJob) string {
	return rankTableDir + "/" + job.Namespace + "." + job.Name
}

// PodHasAllocated check if pod has allocated device
func PodHasAllocated(pod *corev1.Pod) bool {
	if pod.GetDeletionTimestamp() != nil {
		return false
	}
	if _, ok := pod.Annotations[PodDeviceKey]; !ok {
		hwlog.RunLog.Debugf("Pod %s has not allocated device", pod.Name)
		return false
	}
	return true
}

const (
	// PodDeviceKey Pod annoation Key
	PodDeviceKey = "ascend.kubectl.kubernetes.io/ascend-910-configuration"
	// PodRankKey Pod annoation Key
	PodRankKey = "hccl/rankIndex"

	rankTableDir = "/user/mindx-dl/ranktable"
)

// RankTableStatus is rank table status
type RankTableStatus string

const (
	// InitialRTStatus initial rank table status
	InitialRTStatus RankTableStatus = "initializing"
	// CompletedRTStatus completed rank table status
	CompletedRTStatus RankTableStatus = "completed"
)
