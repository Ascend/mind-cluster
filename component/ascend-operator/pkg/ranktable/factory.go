/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package ranktable is using for reconcile AscendJob.
*/
package ranktable

import (
	"ascend-common/common-utils/hwlog"
	mindxdlv1 "ascend-operator/pkg/api/v1"
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
	if _, ok := job.Annotations[utils.AnnoKeyOfSuperPod]; ok {
		hwlog.RunLog.Info("sp-block is exist, use ranktable v1_2")
		return v1dot2.New(job)
	}
	return ranktablev1.New(job)
}
