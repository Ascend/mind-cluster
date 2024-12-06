// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package jobv2 a series of job data collector function
package jobv2

import (
	"k8s.io/api/core/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podGroup"
)

// PodGroupCollector collector podGroup info
func PodGroupCollector(oldPGInfo, newPGInfo *v1beta1.PodGroup, operator string) {
	switch operator {
	case constant.AddOperator, constant.UpdateOperator:
		podGroup.SavePodGroup(newPGInfo)
	case constant.DeleteOperator:
		podGroup.DeletePodGroup(newPGInfo)
	default:
		hwlog.RunLog.Debugf("error operator: %s", operator)
		return
	}
	podGroupMessage(newPGInfo, operator)
}

// PodCollector collector pod info
func PodCollector(oldPodInfo, newPodInfo *v1.Pod, operator string) {
	switch operator {
	case constant.AddOperator, constant.UpdateOperator:
		pod.SavePod(newPodInfo)
	case constant.DeleteOperator:
		pod.DeletePod(newPodInfo)
	default:
		hwlog.RunLog.Debugf("error operator: %s", operator)
		return
	}
	podMessage(oldPodInfo, newPodInfo, operator)
}
