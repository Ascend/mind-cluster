// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package resource a series of resource function
package resource

import (
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/grpc/common"
	"clusterd/pkg/interface/grpc/pb"
	"clusterd/pkg/interface/kube"
)

const (
	notifyChanLen  = 10000
	faultSendDelay = 10
)

// JobSourceStatusManager manage job resource healthyStatus
type JobSourceStatusManager struct {
	publisher   common.Publisher
	notifyChan  chan *common.Notifier
	vcClientK8s *kube.VcK8sClient
}

// NewJobSourceStatusManager return a pointer of JobSourceStatusManager
func NewJobSourceStatusManager(publisher common.Publisher) *JobSourceStatusManager {
	mgr := &JobSourceStatusManager{}
	mgr.publisher = publisher
	mgr.notifyChan = make(chan *common.Notifier, notifyChanLen)
	mgr.vcClientK8s = kube.GetClientVolcano()
	go mgr.checkNotifyChan()
	go mgr.checkNpuDeviceFault()
	return mgr
}

func (mgr *JobSourceStatusManager) checkNpuDeviceFault() {
	for {
		if kube.JobMgr == nil {
			hwlog.RunLog.Error("job mgr is nil, try it after 1s")
			time.Sleep(time.Second)
			continue
		}
		kube.JobMgr.RwMutex.RLock()
		for _, worker := range kube.JobMgr.BsWorker {
			taskId := worker.GetBaseInfo().Uid
			_, rankIds := mgr.GetJobHealthy(taskId)
			if len(rankIds) > 0 {
				mgr.publisher.PublishSignal(&pb.ProcessManageSignal{
					TaskId:       taskId,
					SignalType:   common.StopTrainSignalType,
					Actions:      []string{"stop_train"},
					FaultRankIds: rankIds,
				}, []common.MachineState{common.INIT})
			}
		}
		kube.JobMgr.RwMutex.RUnlock()
		time.Sleep(time.Second * constant.CheckFaultGapSecond)
	}
}

// ListenTaskScheduleResult listen task schedule result and notify publisher handle it.
func (mgr *JobSourceStatusManager) ListenTaskScheduleResult(taskId string, strategy string) {
	if kube.JobMgr == nil {
		hwlog.RunLog.Error("job mgr is nil")
		return
	}
	var worker job.PodWorker
	var exist bool
	kube.JobMgr.RwMutex.RLock()
	defer kube.JobMgr.RwMutex.RUnlock()
	if worker, exist = kube.JobMgr.BsWorker[taskId]; !exist {
		hwlog.RunLog.Errorf("taskId=%s not exist", taskId)
		return
	}
	go func() {
		pgRunning := false
		for i := 1; i <= common.CheckPGRunningRetryTimes; i++ {
			hwlog.RunLog.Debugf("checkout %d times", i)
			time.Sleep(time.Second * common.SleepSecondBeforeCheckPGRunning)
			if worker.PGRunning() {
				pgRunning = true
				break
			}
		}
		mgr.publisher.NotifyJobSchedulerResult(pgRunning, taskId, strategy)
	}()
}

// NotifySignalSend push notifier to notify chan
func (mgr *JobSourceStatusManager) NotifySignalSend(notifier *common.Notifier) {
	mgr.notifyChan <- notifier
}

func delaySleep(timePassSecond, shouldDelaySecond int64) {
	sleepSecond := shouldDelaySecond - timePassSecond
	if sleepSecond > 0 {
		time.Sleep(time.Second * time.Duration(sleepSecond))
	}
}

func (mgr *JobSourceStatusManager) checkNotifyChan() {
	for item := range mgr.notifyChan {
		notifier := item
		go func() {
			if notifier.SignalType == common.GlobalFaultSignalType {
				now := time.Now().Unix()
				timePassed := util.MaxInt(now-notifier.CreateTimeStamp, 0)
				delaySleep(timePassed, faultSendDelay)
			}
			signal := &pb.ProcessManageSignal{
				Uuid:         notifier.Uuid,
				TaskId:       notifier.TaskId,
				SignalType:   notifier.SignalType,
				Actions:      nil,
				FaultRankIds: nil,
			}
			signal.Actions = append(signal.Actions, notifier.Actions...)
			signal.ChangeStrategy = notifier.ChangeStrategy
			switch signal.SignalType {
			case common.StopTrainSignalType:
				mgr.publisher.PublishSignal(signal, []common.MachineState{common.INIT})
			case common.GlobalFaultSignalType:
				mgr.publisher.PublishSignal(signal, []common.MachineState{common.ReceiveStopFinish})
			case common.ChangeStrategySignalType:
				mgr.publisher.PublishSignal(signal,
					[]common.MachineState{common.ListenOnlineRecoverStatus, common.ReceiveSupportStrategy})
			default:
				hwlog.RunLog.Errorf("not support signalType=%s", signal.SignalType)
			}

		}()
	}
}

// GetJobHealthy return whether the job's resource health
func (mgr *JobSourceStatusManager) GetJobHealthy(jobId string) (bool, []string) {

	if kube.JobMgr == nil {
		hwlog.RunLog.Error("job mgr is nil")
		return false, nil
	}
	kube.JobMgr.RwMutex.RLock()
	defer kube.JobMgr.RwMutex.RUnlock()
	if worker, exist := kube.JobMgr.BsWorker[jobId]; !exist {
		hwlog.RunLog.Errorf("taskId=%s not exist", jobId)
		return false, nil
	} else {
		return worker.GetJobHealth()
	}
}

// GetJobNameAndNameSpace return job's name and namespace
func (mgr *JobSourceStatusManager) GetJobNameAndNameSpace(jobId string) (string, string) {
	if kube.JobMgr == nil {
		hwlog.RunLog.Error("job mgr is nil")
		return "", ""
	}
	kube.JobMgr.RwMutex.RLock()
	defer kube.JobMgr.RwMutex.RUnlock()
	worker, exist := kube.JobMgr.BsWorker[jobId]
	if !exist {
		hwlog.RunLog.Errorf("taskId=%s not exist", jobId)
		return "", ""
	}
	baseInfo := worker.GetBaseInfo()
	return baseInfo.Name, baseInfo.Namespace
}

// GetJobDeviceNumPerNode return job's device num per node
func (mgr *JobSourceStatusManager) GetJobDeviceNumPerNode(jobId string) int {
	if kube.JobMgr == nil {
		hwlog.RunLog.Error("job mgr is nil")
		return -1
	}
	kube.JobMgr.RwMutex.RLock()
	defer kube.JobMgr.RwMutex.RUnlock()
	if worker, exist := kube.JobMgr.BsWorker[jobId]; !exist {
		hwlog.RunLog.Errorf("taskId=%s not exist", jobId)
		return -1
	} else {
		return worker.GetDeviceNumPerNode()
	}
}

// IsJobRunning return whether job is running
func (mgr *JobSourceStatusManager) IsJobRunning(jobId string) bool {
	if kube.JobMgr == nil {
		hwlog.RunLog.Error("job mgr is nil")
		return false
	}
	kube.JobMgr.RwMutex.RLock()
	defer kube.JobMgr.RwMutex.RUnlock()
	worker, exist := kube.JobMgr.BsWorker[jobId]
	if !exist {
		hwlog.RunLog.Errorf("taskId=%s not exist", jobId)
		return false
	}
	return worker.PGRunning()
}

// JobExist return whether job exist
func (mgr *JobSourceStatusManager) JobExist(jobId string) bool {
	if kube.JobMgr == nil {
		hwlog.RunLog.Error("job mgr is nil")
		return false
	}
	kube.JobMgr.RwMutex.RLock()
	defer kube.JobMgr.RwMutex.RUnlock()
	_, exist := kube.JobMgr.BsWorker[jobId]
	return exist
}
