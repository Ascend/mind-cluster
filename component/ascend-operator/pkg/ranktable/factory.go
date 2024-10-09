/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package ranktable

import (
	"huawei.com/npu-exporter/v5/common-utils/hwlog"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable/generator"
	ranktablev1 "ascend-operator/pkg/ranktable/v1"
	"ascend-operator/pkg/ranktable/v1dot2"
)

// NewGenerator create ranktable generator
func NewGenerator(job *mindxdlv1.AscendJob) generator.RankTableGenerator {
	if _, ok := job.Annotations[v1dot2.AnnoKeyOfSuperPod]; ok {
		hwlog.RunLog.Info("sp-block is exist, use ranktable v1_2")
		return v1dot2.New(job)
	}
	return ranktablev1.New(job)
}
