// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package podgroup a series of pg util function
package podgroup

import (
	"strings"

	"k8s.io/utils/strings/slices"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/pod"
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
		if owner.Controller != nil && *owner.Controller {
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
		if owner.Controller != nil && *owner.Controller {
			return owner.Kind
		}
	}
	return ""
}

// GetModelFramework get model framework
func GetModelFramework(info *v1beta1.PodGroup) string {
	if info == nil {
		return ""
	}
	modelFramework := info.GetLabels()
	framework, ok := modelFramework[frameWorkKey]
	if ok {
		return framework
	}
	hwlog.RunLog.Debug("get framework from podGroup failed")
	return ""
}

// JudgeRetryByJobKey judge uce label by jobKey
func JudgeRetryByJobKey(jobKey string) bool {
	return judgeTargetRecoverStrategyByJobKey(jobKey, constant.ProcessRetryStrategyName)
}

func judgeTargetRecoverStrategyByJobKey(jobKey string, strategy string) bool {
	podGroup := GetPodGroup(jobKey)
	flag, exit := podGroup.Labels[constant.ProcessRecoverEnableLabel]
	if !exit || flag != constant.ProcessRecoverEnable {
		hwlog.RunLog.Debugf("label isn't contains %s", constant.ProcessRecoverEnableLabel)
		return false
	}
	mindXConfig, exit := podGroup.Annotations[constant.RecoverStrategies]
	if !exit {
		hwlog.RunLog.Debugf("annotation isn't contains %s", constant.RecoverStrategies)
		return false
	}
	mindXConfig = strings.Replace(mindXConfig, " ", "", -1)
	strategyList := strings.Split(mindXConfig, ",")
	if slices.Contains(strategyList, strategy) {
		return true
	}
	hwlog.RunLog.Debugf("strategyList isn't contains %s", strategy)
	return false
}

// JudgeRestartProcessByJobKey judge recover restart-process label by jobKey
func JudgeRestartProcessByJobKey(jobKey string) bool {
	return judgeTargetRecoverStrategyByJobKey(jobKey, constant.ProcessRecoverInPlaceStrategyName)
}

// JudgeIsRunningByJobKey judge podGroup is running  by jobKey
func JudgeIsRunningByJobKey(jobKey string) bool {
	podGroup := GetPodGroup(jobKey)
	return podGroup.Status.Phase == v1beta1.PodGroupRunning
}

// GetPGFromCacheOrPod return job's name, podGroup name and namespace
func GetPGFromCacheOrPod(jobKey string) (jobName, pgName, namespace string) {
	pg := GetPodGroup(jobKey)
	if pg.Name == "" {
		return pod.GetPGByPod(jobKey)
	}
	return GetJobNameByPG(&pg), pg.GetName(), pg.Namespace
}

// GetResourceType get resource type
func GetResourceType(info *v1beta1.PodGroup) string {
	if info == nil {
		return ""
	}
	for key, _ := range info.Spec.MinResources.DeepCopy() {
		if strings.Contains(string(key), constant.Ascend910) {
			return constant.Ascend910
		}
		if strings.Contains(string(key), constant.Ascend310) {
			return constant.Ascend310
		}
		if strings.Contains(string(key), constant.Ascend310P) {
			return constant.Ascend310P
		}
	}
	hwlog.RunLog.Warnf("GetResourceType failed for pg %s", info.GetName())
	return constant.UnknownResourceType
}
