// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package jobv2 a series of job data collector function
package jobv2

import (
	"k8s.io/api/core/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/epranktable"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
	"clusterd/pkg/interface/kube"
)

// PodGroupCollector collector podGroup info
func PodGroupCollector(oldPGInfo, newPGInfo *v1beta1.PodGroup, operator string) {
	if oldPGInfo == nil || newPGInfo == nil {
		hwlog.RunLog.Error("oldPGInfo or newPGInfo is nil")
		return
	}
	switch operator {
	case constant.AddOperator, constant.UpdateOperator:
		podgroup.SavePodGroup(newPGInfo)
	case constant.DeleteOperator:
		kube.RecoverFaultJobInfoCm(podgroup.GetJobKeyByPG(newPGInfo))
		podgroup.DeletePodGroup(newPGInfo)
	default:
		hwlog.RunLog.Debugf("error operator: %s", operator)
		return
	}
	podGroupMessage(newPGInfo, operator)
}

// PodCollector collector pod info
func PodCollector(oldPodInfo, newPodInfo *v1.Pod, operator string) {
	if oldPodInfo == nil || newPodInfo == nil {
		hwlog.RunLog.Error("oldPodInfo or newPodInfo is nil")
		return
	}
	switch operator {
	case constant.AddOperator, constant.UpdateOperator:
		pod.SavePod(newPodInfo)
		refreshCmWhenPodRescheduleInPlace(oldPodInfo, newPodInfo)
	case constant.DeleteOperator:
		pod.DeletePod(newPodInfo)
	default:
		hwlog.RunLog.Debugf("error operator: %s", operator)
		return
	}
	podMessage(oldPodInfo, newPodInfo, operator)
}

func refreshCmWhenPodRescheduleInPlace(oldPodInfo, newPodInfo *v1.Pod) {
	if oldPodInfo.Annotations[api.RescheduleInPlaceKey] == "" &&
		newPodInfo.Annotations[api.RescheduleInPlaceKey] == api.RescheduleInPlaceValue {
		hwlog.RunLog.Infof("refresh cm when pod %s reschedule in place", newPodInfo.Name)
		go kube.RecoverFaultJobInfoCmWithSync(pod.GetJobKeyByPod(newPodInfo))
	}
}

// EpGlobalRankTableMassageCollector collector generate global rank table message
func EpGlobalRankTableMassageCollector(oldPodInfo, newPodInfo *v1.Pod, operator string) {
	if !checkPodIsControllerOrCoordinator(newPodInfo) {
		return
	}
	epranktable.InformerHandler(oldPodInfo, newPodInfo, operator)
}

// checkPodIsControllerOrCoordinator check if pod is controller or coordinator
func checkPodIsControllerOrCoordinator(obj interface{}) bool {
	changedPod, ok := obj.(*v1.Pod)
	if !ok {
		hwlog.RunLog.Errorf("Cannot convert to Pod:%v", obj)
		return false
	}
	appType, ok := changedPod.Labels[constant.MindIeAppTypeLabelKey]
	if !ok {
		return false
	}
	return appType == constant.ControllerAppType || appType == constant.CoordinatorAppType
}
