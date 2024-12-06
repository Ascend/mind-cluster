// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package podGroup a series of pg util function
package podGroup

import (
	"ascend-common/common-utils/hwlog"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
)

// GetJobKeyByPG get job unique key by podGroup
func GetJobKeyByPG(info *v1beta1.PodGroup) string {
	key, _ := GetJobKeyAndNameByPG(info)
	return key
}

// GetJobNameByPG get job name by podGroup
func GetJobNameByPG(info *v1beta1.PodGroup) string {
	_, name := GetJobKeyAndNameByPG(info)
	return name
}

// GetJobKeyAndNameByPG get job unique key and name by podGroup
func GetJobKeyAndNameByPG(info *v1beta1.PodGroup) (key, name string) {
	if info == nil {
		hwlog.RunLog.Error("get unique key failed, podGroup is nil")
		return "", ""
	}
	for _, owner := range info.GetOwnerReferences() {
		if *owner.Controller {
			return string(owner.UID), owner.Name
		}
	}
	hwlog.RunLog.Error("get unique key failed, podGroup don't have controller")
	return "", ""
}

// GetJobTypeByPG get job type by podGroup
func GetJobTypeByPG(podGroup *v1beta1.PodGroup) string {
	if podGroup == nil || len(podGroup.OwnerReferences) == 0 {
		return ""
	}
	for _, owner := range podGroup.GetOwnerReferences() {
		if *owner.Controller {
			return owner.Kind
		}
	}
	return ""
}

// GetModelFramework get model framework
func GetModelFramework(info *v1beta1.PodGroup) string {
	modelFramework := info.GetLabels()
	framework, ok := modelFramework[frameWorkKey]
	if ok {
		return framework
	}
	hwlog.RunLog.Debug("get framework from podGroup failed")
	return ""
}

// JudgeUceByJobKey judge uce label by jobKey
func JudgeUceByJobKey(jobKey string) bool {
	podGroup := GetPodGroup(jobKey)
	if len(podGroup.Labels) == 0 {
		return false
	}
	if flag, exit := podGroup.Labels[stepRetryKey]; exit && flag == onStepRetry {
		return true
	}
	return false
}

// JudgeIsRunningByJobKey judge podGroup is running  by jobKey
func JudgeIsRunningByJobKey(jobKey string) bool {
	podGroup := GetPodGroup(jobKey)
	return podGroup.Status.Phase == v1beta1.PodGroupRunning
}
