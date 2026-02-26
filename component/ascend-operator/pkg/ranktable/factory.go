/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package ranktable is using for reconcile AscendJob.
*/
package ranktable

import (
	corev1 "k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable/common"
	"ascend-operator/pkg/ranktable/generator"
	ranktablev1 "ascend-operator/pkg/ranktable/v1"
	"ascend-operator/pkg/ranktable/v1dot2"
	"ascend-operator/pkg/ranktable/v2dot0"
	"ascend-operator/pkg/utils"
)

// NewGenerator create ranktable generator
func NewGenerator(job *mindxdlv1.AscendJob) generator.RankTableGenerator {
	if job == nil {
		return ranktablev1.New(job)
	}
	// for A5 ranktable
	if useV2dot0(job) {
		hwlog.RunLog.Info("use ranktable generator v2.0")
		return v2dot0.New(job)
	}
	if useV1dot2(job) {
		hwlog.RunLog.Info("A3 super pod job, use ranktable v1_2")
		return v1dot2.New(job)
	}
	return ranktablev1.New(job)
}

func useV2dot0(job *mindxdlv1.AscendJob) bool {
	for _, replicaSpec := range job.Spec.ReplicaSpecs {
		if containsNPURequest(replicaSpec.Template.Spec.Containers) {
			return true
		}
	}
	return false
}

func containsNPURequest(containers []corev1.Container) bool {
	for _, container := range containers {
		if _, ok := container.Resources.Requests[api.HuaweiNPU]; ok {
			return true
		}
	}
	return false
}

func useV1dot2(job *mindxdlv1.AscendJob) bool {
	if policy, schedulePolicyExit := job.Annotations[common.SchedulePolicyAnnoKey]; schedulePolicyExit {
		return policy == utils.Chip2Node16Sp || policy == utils.Chip2Node8Sp
	}
	if _, spBlockExit := job.Annotations[utils.AnnoKeyOfSuperPod]; spBlockExit {
		return true
	}
	for _, replicaSpec := range job.Spec.ReplicaSpecs {
		if replicaSpec.Template.Spec.NodeSelector == nil {
			continue
		}
		value, ok := replicaSpec.Template.Spec.NodeSelector[api.AcceleratorTypeKey]
		if ok && (value == api.AcceleratorTypeModule910A3x16SuperPod ||
			value == api.AcceleratorTypeModule910A3x8SuperPod) {
			return true
		}
	}
	return false
}
