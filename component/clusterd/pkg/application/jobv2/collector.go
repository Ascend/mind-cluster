// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package jobv2 a series of job data collector function
package jobv2

import (
	"time"

	"k8s.io/api/core/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/logs"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/epranktable"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
	"clusterd/pkg/domain/statistics"
	"clusterd/pkg/interface/kube"
)

// PodGroupCollector collector podGroup info
func PodGroupCollector(oldPGInfo, newPGInfo *v1beta1.PodGroup, operator string) {
	if newPGInfo == nil {
		hwlog.RunLog.Error("newPGInfo is nil")
		return
	}
	switch operator {
	case constant.AddOperator, constant.UpdateOperator:
		podgroup.SavePodGroup(newPGInfo)
	case constant.DeleteOperator:
		kube.RecoverFaultJobInfoCm(podgroup.GetJobKeyByPG(newPGInfo), "")
		podgroup.DeletePodGroup(newPGInfo)
	default:
		hwlog.RunLog.Debugf("error operator: %s", operator)
		return
	}
	podGroupMessage(newPGInfo, operator)
}

// PodCollector collector pod info
func PodCollector(oldPodInfo, newPodInfo *v1.Pod, operator string) {
	if newPodInfo == nil {
		hwlog.RunLog.Error("newPodInfo is nil")
		return
	}
	switch operator {
	case constant.AddOperator:
		pod.SavePod(newPodInfo)
		refreshCmWhenPodRescheduleInPlace(oldPodInfo, newPodInfo)
		recordPodErrorOnFailure(oldPodInfo, newPodInfo, operator)
	case constant.UpdateOperator:
		pod.UpdatePod(oldPodInfo, newPodInfo)
		refreshCmWhenPodRescheduleInPlace(oldPodInfo, newPodInfo)
		recordPodErrorOnFailure(oldPodInfo, newPodInfo, operator)
	case constant.DeleteOperator:
		pod.DeletePod(newPodInfo)
		recordPodErrorOnFailure(oldPodInfo, newPodInfo, operator)
	default:
		hwlog.RunLog.Debugf("error operator: %s", operator)
		return
	}
	podMessage(oldPodInfo, newPodInfo, operator)
}

func refreshCmWhenPodRescheduleInPlace(oldPodInfo, newPodInfo *v1.Pod) {
	if oldPodInfo == nil || newPodInfo == nil {
		return
	}
	if oldPodInfo.Annotations[api.RescheduleInPlaceKey] == "" &&
		newPodInfo.Annotations[api.RescheduleInPlaceKey] == api.RescheduleInPlaceValue {
		hwlog.RunLog.Infof("refresh cm when pod %s reschedule in place", newPodInfo.Name)
		kube.RecoverFaultJobInfoCmWithSync(pod.GetJobKeyByPod(newPodInfo), newPodInfo.Spec.NodeName)
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

func recordPodErrorOnFailure(oldPodInfo, newPodInfo *v1.Pod, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		if newPodInfo.Status.Phase != v1.PodFailed {
			return
		}
		if oldPodInfo != nil && oldPodInfo.Status.Phase == v1.PodFailed {
			return
		}
	}
	if newPodInfo == nil || newPodInfo.Spec.NodeName == "" {
		return
	}
	jobKey := pod.GetJobKeyByPod(newPodInfo)
	if jobKey == "" {
		return
	}
	nowTime := time.Now().Unix()
	statistics.JobStcMgrInst.UpdateJobStatistic(jobKey, func(jobStc *constant.JobStatisticV2) {
		jobStc.PodErrorTimestamp = appendPodErrorTimestamp(jobKey,
			jobStc.PodErrorTimestamp, nowTime, newPodInfo.Spec.NodeName, newPodInfo.Name)
		logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(jobStc))
	})
}

func appendPodErrorTimestamp(k8sJobID string, infos []constant.PodErrorInfo,
	ts int64, nodeName, podName string) []constant.PodErrorInfo {
	if len(infos) >= constant.MaxTimestampRecords {
		hwlog.RunLog.Warnf("job %s PodErrorTimestamp slice length is over %v", k8sJobID, constant.MaxTimestampRecords)
		infos = infos[1:]
	}
	return append(infos, constant.PodErrorInfo{
		Timestamp: ts, NodeName: nodeName, PodName: podName,
	})
}
