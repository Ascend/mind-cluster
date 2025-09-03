// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package recover a series of service function
package recover

import (
	"fmt"
	"time"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/interface/grpc/recover"
	"clusterd/pkg/interface/kube"
	"k8s.io/api/core/v1"
)

func (ctl *EventController) notifyPrepareHotSwitch() (string, common.RespCode, error) {
	hwlog.RunLog.Infof("hotswitch flow, jobId: %v, enter into notifyPrepareHotSwitch", ctl.jobInfo.JobId)
	// mark all ranks in same pod/node as fault in hotswitch scenario
	allFaults, _ := ctl.normalFaultAssociateSameNodeRank()
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.HotSwitchSignalType,
		Actions:        hotSwitchActions,
		FaultRanks:     allFaults,
		ChangeStrategy: "",
	}
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) notifyCreateNewPod() (string, common.RespCode, error) {
	hwlog.RunLog.Infof("hotswitch flow, jobId: %v, enter into notifyCreateNewPod", ctl.jobInfo.JobId)
	for podRank, podId := range ctl.GetFaultPod() {
		hwlog.RunLog.Infof("hotswitch: notifyCreateNewPod,podRank=%s, podId=%s", podRank, podId)
		pod, exist := pod.GetPodByPodId(podId)
		if !exist {
			continue
		}
		err := kube.RetryPatchPodAnnotations(pod.Name, pod.Namespace, api.DefaultRetryTimes,
			map[string]string{api.InHotSwitchFlowKey: api.InHotSwitchFlowValue, api.NeedOperatorOpeKey: api.OpeTypeCreate})
		if err != nil {
			hwlog.RunLog.Infof("hotSwitch: notifyCreateNewPod faild,pod :%v, err: %v", pod.Name, err)
			continue
		}

		hwlog.RunLog.Infof("hotSwitch: notifyCreateNewPod success,pod :%v", pod.Name)
		// only handle one fault pod
		ctl.currentHotSwitchFaultPodId = string(pod.UID)
		break
	}
	go monitorNewPodStatus(ctl)

	return "", common.OK, nil
}

func monitorNewPodStatus(ctl *EventController) {
	ctx, ch := ctl.getCtxAndNewStatusMonitorChan()
	if ch == nil {
		hwlog.RunLog.Infof("jobId=%s, newPodStatusMonitorChan is nil", ctl.jobInfo.JobId)
		return
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
	case req, ok := <-ch:
		if !ok {
			hwlog.RunLog.Warnf("jobId=%s, uuid=%s, newPodStatusMonitorChan is closed", ctl.jobInfo.JobId, ctl.uuid)
			return
		}
		if req == v1.PodRunning {
			hwlog.RunLog.Infof("jobId=%s, uuid=%s, new pod running success", ctl.jobInfo.JobId, ctl.uuid)
		}
		ctl.addEvent(common.NewPodRunningEvent)
	case <-time.After(time.Minute):
		hwlog.RunLog.Errorf("hotswitch: wait new pod running timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.addEvent(common.NewPodTimeoutEvent)
	}
}

func (ctl *EventController) notifyNewPodFailedHandler() (string, common.RespCode, error) {
	hwlog.RunLog.Infof("hotswitch flow, jobId: %v, enter into notifyNewPodFailedHandler", ctl.jobInfo.JobId)
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.HotSwitchSignalType,
		Actions:        stopHotSwitchActions,
		ChangeStrategy: "",
	}
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) notifyNewPodRunningHandler() (string, common.RespCode, error) {
	hwlog.RunLog.Infof("hotswitch flow, jobId: %v, enter into notifyNewPodRunningHandler", ctl.jobInfo.JobId)
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.HotSwitchSignalType,
		Actions:        newPodRunningActions,
		ChangeStrategy: "",
	}
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) notifyDeleteOldPod() (string, common.RespCode, error) {
	hwlog.RunLog.Infof("hotswitch flow, jobId: %v, enter into notifyDeleteOldPod,podId:%v",
		ctl.jobInfo.JobId, ctl.currentHotSwitchFaultPodId)
	pod, exist := pod.GetPodByPodId(ctl.currentHotSwitchFaultPodId)
	if !exist {
		return "", common.OK, nil
	}
	kube.RetryPatchPodAnnotations(pod.Name, pod.Namespace, api.DefaultRetryTimes,
		map[string]string{api.NeedVolcanoOpeKey: api.OpeTypeDelete})
	kube.RetryPatchPodLabels(pod.Name, pod.Namespace, api.DefaultRetryTimes,
		map[string]string{constant.TaskFaultKey: constant.SubHealthFaultStrategy})

	hwlog.RunLog.Infof("hotswitch flow, notifyDeleteOldPod success,pod :%v", pod.Name)
	return "", common.OK, nil
}

func (ctl *EventController) notifyStopJob() (string, common.RespCode, error) {
	hwlog.RunLog.Infof("hotswitch flow, jobId: %v, enter into notifyStopJob", ctl.jobInfo.JobId)
	pod := pod.GetPodByRankIndex(ctl.jobInfo.JobId, "0")
	if pod.Name == "" {
		hwlog.RunLog.Errorf("hotSwitch: can not find rank0 pod,jobId:%v,jobName:%v", ctl.jobInfo.JobId, ctl.jobInfo.JobName)
		return common.NotifySuccessEvent, common.OK, nil
	}
	kube.RetryPatchPodAnnotations(pod.Name, pod.Namespace, api.DefaultRetryTimes,
		map[string]string{api.NeedVolcanoOpeKey: api.OpeTypeDelete})
	kube.RetryPatchPodLabels(pod.Name, pod.Namespace, api.DefaultRetryTimes,
		map[string]string{constant.TaskFaultKey: constant.SubHealthFaultStrategy})
	return common.NotifySuccessEvent, common.OK, nil
}

func (ctl *EventController) notifyRestartTrain() (string, common.RespCode, error) {
	hwlog.RunLog.Infof("hotswitch flow, jobId: %v, enter into notifyRestartTrain", ctl.jobInfo.JobId)
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.ChangeStrategySignalType,
		Actions:        changeStrategyActions,
		ChangeStrategy: constant.ProcessMigration,
	}
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) waitReportPauseTrainResult() (string, common.RespCode, error) {
	hwlog.RunLog.Infof("hotswitch flow, jobId: %v, enter into waitReportPauseTrainResult", ctl.jobInfo.JobId)

	ctx, strategyChan := ctl.getCtxAndReportRecoverStrategyChan()
	if strategyChan == nil {
		hwlog.RunLog.Errorf("jobId=%s, strategyChan is nil", ctl.jobInfo.JobId)
		return "", common.ServerInnerError, fmt.Errorf("jobId=%s, strategyChan is nil", ctl.jobInfo.JobId)
	}
	select {
	case req, ok := <-strategyChan:
		if !ok {
			hwlog.RunLog.Warnf("strategyChan closed, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
			return "", common.OK, nil
		}
		strategies := req.Strategies
		hwlog.RunLog.Infof("hotSwitch: recieve mindio report train pause result, strategies:%v", strategies)
		if len(strategies) == 0 {
			hwlog.RunLog.Errorf("recevice hotSwitch strategies is err,exported one, but [%v]", len(strategies))
			ctl.addEvent(common.ExitEvent)
			return "", common.OK, nil
		}
		if strategies[0] == constant.ProcessMigration {
			ctl.addEvent(common.MigrationEvent)
		} else {
			ctl.addEvent(common.ExitEvent)
		}
		return "", common.OK, nil
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case <-time.After(reportTimeoutMinutes * time.Minute):
		hwlog.RunLog.Errorf("wait report recover strategy timeout, jobId=%s", ctl.jobInfo.JobId)
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) handleWaitReportRestartTrainStatus() (string, common.RespCode, error) {
	hwlog.RunLog.Infof("hotswitch flow, jobId: %v, enter into handleWaitReportRestartTrainStatus", ctl.jobInfo.JobId)

	ctx, ch := ctl.getCtxAndResultChan()
	if ch == nil {
		hwlog.RunLog.Infof("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
		return common.RestartFaildEvent, common.OK, nil
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case req := <-ch:
		hwlog.RunLog.Infof("hotswitch: recieve mindio report restart train result,req.Status:%#v", req.Status)
		if req.Status.Code == common.UnRecoverableRetryError {
			return common.RestartFaildEvent, common.OK, nil
		} else if req.Status.Code == int32(common.OK) {
			return common.RestartSuccessEvent, common.OK, nil
		} else {
			hwlog.RunLog.Warnf("responce code from mindio is not expected,code: %v", req.Status.Code)
		}
	case <-time.After(reportTimeoutMinutes * time.Minute):
		hwlog.RunLog.Errorf("wait report restart train complete timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
	return "", common.OK, nil
}

func (ctl *EventController) cleanStateWhenFailed() (string, common.RespCode, error) {
	hwlog.RunLog.Infof("hotswitch flow, jobId: %v, enter into cleanStateWhenFailed", ctl.jobInfo.JobId)
	// change subhealthy strategy to ignore when hotswitch failed
	hwlog.RunLog.Infof("jobId: %v,change subhealthy strategy to ignore when hotswitch failed", ctl.jobInfo.JobId)
	defer func() {
		ctl.jobInfo.HotSwitch = false
		ctl.jobInfo.SubHealthyStrategy = constant.SubHealthyIngore
		ctl.reset(false)
	}()
	// fault pod, get lastest pod from informer
	faultPod, exist := pod.GetPodByPodId(ctl.currentHotSwitchFaultPodId)
	if !exist {
		return "", common.OK, nil
	}
	_, ok := faultPod.Annotations[api.InHotSwitchFlowKey]
	if !ok {
		return "", common.OK, nil
	}
	err := kube.DeletePodAnnotation(faultPod.Namespace, faultPod.Name,
		[]string{api.InHotSwitchFlowKey, api.BackupNewPodNameKey, api.NeedOperatorOpeKey})
	if err != nil {
		hwlog.RunLog.Errorf("hotSwitch: delete faultPod %s annotation inHotSwitchFlow、"+
			"backupNewPodName、needOperatorOpe failed,err: %v", faultPod.Name, err)
		return "", common.ServerInnerError, err
	}
	// new pod
	newPodName := faultPod.Annotations[api.BackupNewPodNameKey]
	newPod, exists := pod.GetPodByJobIdAndPodName(ctl.jobInfo.JobId, newPodName)
	if !exists {
		hwlog.RunLog.Errorf("hotSwitch: newPod %s not exists in informer", faultPod.Name)
		return "", common.ServerInnerError, err
	}
	err = kube.RetryPatchPodAnnotations(newPod.Name, newPod.Namespace, api.DefaultRetryTimes,
		map[string]string{api.NeedOperatorOpeKey: api.OpeTypeDelete})
	if err != nil {
		hwlog.RunLog.Errorf("hotSwitch: patch newPod %s annotation needOperatorOpe[delete] failed,err: %v",
			newPod.Name, err)
		return "", common.ServerInnerError, err
	}
	return "", common.OK, nil
}

func (ctl *EventController) cleanStateWhenSuccess() (string, common.RespCode, error) {
	hwlog.RunLog.Infof("hotswitch flow, jobId: %v, enter into cleanStateWhenSuccess", ctl.jobInfo.JobId)
	defer func() {
		ctl.reset(false)
	}()

	// new pod, get lastest pod from informer
	pod, exist := pod.GetPodByPodId(ctl.currentHotSwitchBackupPodId)
	if !exist {
		return "", common.OK, nil
	}
	err := kube.DeletePodAnnotation(pod.Namespace, pod.Name, []string{api.InHotSwitchFlowKey})
	if err != nil {
		return "", common.ServerInnerError, err
	}
	return "", common.OK, nil
}
