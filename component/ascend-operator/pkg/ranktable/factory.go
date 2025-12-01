/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package ranktable is using for reconcile AscendJob.
*/
package ranktable

import (
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable/common"
	"ascend-operator/pkg/ranktable/generator"
	ranktablev1 "ascend-operator/pkg/ranktable/v1"
	"ascend-operator/pkg/ranktable/v1dot2"
	"ascend-operator/pkg/utils"
)

// NewGenerator create ranktable generator
func NewGenerator(job *mindxdlv1.AscendJob) generator.RankTableGenerator {
	if job == nil {
		return ranktablev1.New(job)
	}
	if useV1dot2(job) {
		hwlog.RunLog.Info("super pod job, use ranktable v1_2")
		return v1dot2.New(job)
	}
	return ranktablev1.New(job)
}

func useV1dot2(job *mindxdlv1.AscendJob) bool {
	if policy, schedulePolicyExit := job.Annotations[common.SchedulePolicyAnnoKey]; schedulePolicyExit {
		return policy == utils.Chip2Node16Sp
	}
	if _, spBlockExit := job.Annotations[utils.AnnoKeyOfSuperPod]; spBlockExit {
		return true
	}
	if value, ok := job.Labels[api.AcceleratorTypeKey]; ok && value == api.AcceleratorTypeModule910A3SuperPod {
		return true
	}

	return false
}
