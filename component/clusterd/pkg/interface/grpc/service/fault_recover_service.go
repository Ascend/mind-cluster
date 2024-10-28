// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package service a series of service function
package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/application/resource"
	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/grpc/common"
	"clusterd/pkg/interface/grpc/pb"
	"clusterd/pkg/interface/kube"
)

var _ common.Publisher = &FaultRecoverService{}
var randomBytesLen = 16
var retryTimes = 5

var (
	stopTrainActions      = []string{"stop_train"}
	globalFaultActions    = []string{"on_global_rank"}
	changeStrategyActions = []string{"change_strategy"}
)

// EventController is recover event controller
type EventController struct {
	taskId                    string
	taskName                  string
	nameSpace                 string
	eventId                   string
	latestStrategy            []string
	signalChan                chan *pb.ProcessManageSignal
	softwareCacheFaultRankIds []string
	deviceCacheFaultRankIds   []string
	state                     common.MachineState
	mode                      common.RecoverMode
	lock                      sync.RWMutex
}

// NewEventController return pointer of EventController
func NewEventController(taskId, taskName, namespace string) *EventController {
	ctl := &EventController{}
	ctl.taskId = taskId
	ctl.eventId = "init"
	ctl.taskName, ctl.nameSpace = taskName, namespace
	ctl.latestStrategy = []string{}
	ctl.signalChan = make(chan *pb.ProcessManageSignal, 1)
	ctl.softwareCacheFaultRankIds = []string{}
	ctl.deviceCacheFaultRankIds = []string{}
	ctl.state = common.INIT
	ctl.mode = common.InitMode
	return ctl
}

func (ctl *EventController) loopCheckEventHealthy() {
	var latestState common.MachineState = ctl.state
	latestStateCheckTimestamp := time.Now().Unix()
	for latestState != common.INIT {
		time.Sleep(time.Minute)
		curState := ctl.state
		now := time.Now().Unix()
		if curState == latestState && curState != common.INIT && now-latestStateCheckTimestamp >= common.StateTimeoutSecond {
			hwlog.RunLog.Errorf("machine state time out, taskId=%s, eventId=%s, curMode=%s, curState=%s",
				ctl.taskId, ctl.eventId, common.ModeToString(ctl.mode), common.StateToString(ctl.state))
			ctl.reset(false)
			break
		}
		latestStateCheckTimestamp = now
		latestState = curState
	}
}

func updateFixResult(name, namespace, platStrategy string, success bool) {
	switch platStrategy {
	case common.ProcessArfStrategy:
		if success {
			UpdateRecoverStatus(name, namespace, common.RecoverSuccess)
		} else {
			UpdateRecoverStatus(name, namespace, common.RecoverFailed)
		}
	case common.ProcessDumpStrategy:
		if success {
			UpdateRecoverStatus(name, namespace, common.DumpSuccess)
		} else {
			UpdateRecoverStatus(name, namespace, common.DumpFailed)
		}
	case common.ProcessExitStrategy:
		UpdateRecoverStatus(name, namespace, common.ExitCompleted)
	default:
		hwlog.RunLog.Infof("name=%s is not platform job, don't need update recover status", name)
	}
}

func (ctl *EventController) resetControllerParameters() {
	ctl.softwareCacheFaultRankIds, ctl.deviceCacheFaultRankIds =
		ctl.softwareCacheFaultRankIds[:0], ctl.deviceCacheFaultRankIds[:0]
	if ctl.signalChan != nil && ctl.state != common.INIT && ctl.mode != common.InitMode {
		close(ctl.signalChan)
	}
	ctl.eventId, ctl.latestStrategy = "init", []string{}
	ctl.signalChan = make(chan *pb.ProcessManageSignal, 1)
	ctl.state, ctl.mode = common.INIT, common.InitMode
}

func (ctl *EventController) reset(success bool) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	if _, err := common.WriteResetInfoToCM(ctl.taskName, ctl.nameSpace,
		[]string{}, "clear"); err != nil {
		hwlog.RunLog.Errorf("clear reset cm failed when reset, name=%s, namespace=%s",
			ctl.taskName, ctl.nameSpace)
	}
	if len(ctl.latestStrategy) > 0 {
		platStrategy := GetPlatStrategy(ctl.taskName+common.MiddleLine+ctl.taskId,
			ctl.nameSpace)
		updateFixResult(ctl.taskName+common.MiddleLine+ctl.taskId, ctl.nameSpace, platStrategy, success)
	}
	if !success {
		ctl.mode = common.PodRescheduleMode
		ctl.state = common.StartPodReschedule
		if _, err := common.ChangeProcessSchedulingMode(ctl.taskName+common.MiddleLine+ctl.taskId,
			ctl.nameSpace, common.ProcessReschedulingPause); err != nil {
			hwlog.RunLog.Errorf("failed to change the process rescheduling label %s of pg %s",
				common.ProcessReschedulingPause, ctl.taskName)
		} else {
			hwlog.RunLog.Infof("change process rescheduling label %s success,"+
				" taskId=%s, eventId=%s", common.ProcessReschedulingPause, ctl.taskId, ctl.eventId)
		}
		go func() {
			worker, exist := kube.JobMgr.BsWorker[ctl.taskId]
			if !exist {
				hwlog.RunLog.Error(fmt.Errorf("taskId=%s not exist", ctl.taskId))
				return
			}
			for i := 1; i <= common.CheckPGRunningRetryTimes; i++ {
				time.Sleep(time.Second * common.SleepSecondBeforeCheckPGRunning)
				if worker.PGRunning() {
					break
				}
			}
			if _, err := common.ChangeProcessSchedulingMode(
				ctl.taskName+common.MiddleLine+ctl.taskId, ctl.nameSpace, common.ProcessReschedulingEnable); err != nil {
				hwlog.RunLog.Errorf("failed to change the process rescheduling label %s of pg %s",
					common.ProcessReschedulingEnable, ctl.taskName)
			} else {
				hwlog.RunLog.Infof("change process rescheduling label %s success,"+
					" taskId=%s, eventId=%s", common.ProcessReschedulingEnable, ctl.taskId, ctl.eventId)
			}
		}()
	}
	ctl.resetControllerParameters()
}

func (ctl *EventController) appendStrategy(strategy string) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	ctl.latestStrategy = append(ctl.latestStrategy, strategy)
}

func (ctl *EventController) openEvent(eventId string, openMode common.RecoverMode) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	ctl.eventId = eventId
	ctl.mode = openMode
	go ctl.loopCheckEventHealthy()
}

func (ctl *EventController) saveSoftwareFaultRankIds(ranks []string) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	mr := util.RemoveSliceDuplicateElement(append(ctl.softwareCacheFaultRankIds, ranks...))
	ctl.softwareCacheFaultRankIds = mr
}

func (ctl *EventController) saveDeviceFaultRankIds(ranks []string) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	mr := util.RemoveSliceDuplicateElement(append(ctl.deviceCacheFaultRankIds, ranks...))
	ctl.deviceCacheFaultRankIds = mr
}

func (ctl *EventController) getSoftwareFaultRankIds() []string {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	var res []string
	res = append(res, ctl.softwareCacheFaultRankIds...)
	return res
}

func (ctl *EventController) getHardFaultRankIds() []string {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	var res []string
	res = append(res, ctl.deviceCacheFaultRankIds...)
	return res
}

func (ctl *EventController) tryChangeState(mode common.RecoverMode, newState common.MachineState,
	expectPreStates []common.MachineState, force bool) *pb.Status {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	baseInfo := fmt.Sprintf("taskId=%s, oldMode=%s, newMode=%s, oldState=%s, newState=%s",
		ctl.taskId, common.ModeToString(ctl.mode), common.ModeToString(mode),
		common.StateToString(ctl.state), common.StateToString(newState))
	if force == true {
		info := fmt.Sprintf("state changed success with force, %s", baseInfo)
		ctl.state = newState
		ctl.mode = mode
		return &pb.Status{Code: pb.RespCode_OK, Info: info}
	}
	if ctl.mode == common.InitMode { // event not open
		if !common.CheckOrder(ctl.state, expectPreStates) {
			info := fmt.Sprintf("state changed rejected, %s", baseInfo)
			return &pb.Status{Code: pb.RespCode_ORDER_MIXED, Info: info}
		}
		info := fmt.Sprintf("state changed success, %s", baseInfo)
		ctl.mode = mode
		ctl.state = newState
		hwlog.RunLog.Info(info)
		return &pb.Status{Code: pb.RespCode_OK, Info: info}
	} else { // event processing
		if ctl.mode != mode {
			info := fmt.Sprintf("state changed rejected, %s", baseInfo)
			return &pb.Status{Code: pb.RespCode_MODEL_MIXED, Info: info}
		}
		if !common.CheckOrder(ctl.state, expectPreStates) {
			info := fmt.Sprintf("state changed rejected, %s", baseInfo)
			return &pb.Status{Code: pb.RespCode_ORDER_MIXED, Info: info}
		}
		info := fmt.Sprintf("state changed success, %s", baseInfo)
		ctl.state = newState
		hwlog.RunLog.Info(info)
		return &pb.Status{Code: pb.RespCode_OK, Info: info}
	}
}

// FaultRecoverService is a service for hbm fault
type FaultRecoverService struct {
	pb.UnimplementedRecoverServer
	eventCtl      map[string]*EventController
	regisMap      map[string]struct{}
	rescheduleMap map[string]chan string
	jobMgr        common.JobHealthyMgr
	lock          sync.RWMutex
}

// NewFaultRecoverService return a pointer of FaultRecoverService
func NewFaultRecoverService() *FaultRecoverService {
	svc := &FaultRecoverService{}
	svc.jobMgr = resource.NewJobSourceStatusManager(svc)
	svc.regisMap = make(map[string]struct{})
	svc.eventCtl = make(map[string]*EventController)
	svc.rescheduleMap = make(map[string]chan string)
	return svc
}

func (s *FaultRecoverService) waitRestartAllProcess() {
	time.Sleep(common.WaitProcessRestart * time.Second)
}

func (s *FaultRecoverService) handleTaskScheduleResult(controller *EventController,
	scheduleSuccess bool, strategy string) {
	switch strategy {
	case common.ProcessDumpStrategy, common.ProcessExitStrategy:
		if scheduleSuccess {
			configMap, err := common.RetryWriteResetCM(controller.taskName, controller.nameSpace,
				nil, common.RestartAllProcess)
			if err != nil {
				hwlog.RunLog.Errorf("reset configMap update to restart all process err: %v", err)
				controller.reset(false)
				return
			}
			hwlog.RunLog.Infof("write restart configMap success, %s", configMap.Data[common.ResetInfoCMDataKey])
			s.waitRestartAllProcess()
		}
		controller.reset(scheduleSuccess)
	case common.ProcessArfStrategy:
		if !scheduleSuccess {
			hwlog.RunLog.Warnf("pod schedule failed, taskName=%s, upgrade recover strategy", controller.taskName)
			s.upgradeRecoverStrategy(controller, strategy)
			return
		}
		_, err := common.RetryWriteResetCM(controller.taskName, controller.nameSpace,
			nil, "clear")
		if err != nil {
			hwlog.RunLog.Errorf("reset configMap update to restart all process err: %v", err)
			controller.reset(false)
			return
		}
	default:
		hwlog.RunLog.Errorf("unexpect case, no support strategy=%s", strategy)
	}
}

// NotifyJobSchedulerResult handle schedule listen result
func (s *FaultRecoverService) NotifyJobSchedulerResult(scheduleSuccess bool, taskId string, strategy string) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	hwlog.RunLog.Infof("get schedule result, taskId=%s, strategy=%s, success=%v", taskId, strategy, scheduleSuccess)
	if controller, exist := s.eventCtl[taskId]; exist && controller != nil {
		if strategy == common.ProcessArfStrategy {
			s.handleTaskScheduleResult(controller, scheduleSuccess, strategy)
			return
		}
		if status := controller.tryChangeState(common.ProcessFaultRecoverMode, common.GetJobScheduleResult,
			[]common.MachineState{common.StartListenSchedule}, false); status != nil && status.Code != pb.RespCode_OK {
			hwlog.RunLog.Errorf("error occur when handle schedule result, %s", status.String())
			controller.reset(false)
			return
		}
		s.handleTaskScheduleResult(controller, scheduleSuccess, strategy)
	}
}

func (s *FaultRecoverService) isRegistered(id string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if _, ok := s.regisMap[id]; ok {
		return true
	}
	return false
}

func (s *FaultRecoverService) registry(id string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.regisMap[id] = struct{}{}
}

func (s *FaultRecoverService) checkRegistered(id string, info string) (bool, string) {
	if !s.isRegistered(id) {
		returnInfo := fmt.Sprintf("not registered, %s", info)
		return false, returnInfo
	}
	return true, info
}

// Register is task register service
func (s *FaultRecoverService) Register(ctx context.Context, req *pb.ClientInfo) (*pb.Status, error) {
	info := fmt.Sprintf("role=%s, taskId=%s, addr=%s:%s", req.Role, req.TaskId, req.Ip, req.Port)
	hwlog.RunLog.Infof("service receive Register request, %s", info)
	if !s.jobMgr.JobExist(req.TaskId) {
		hwlog.RunLog.Errorf("registry failed, jobId=%s not exist", req.TaskId)
		return &pb.Status{
			Code: pb.RespCode_COMMON_ERROR,
			Info: fmt.Sprintf("jobId=%s not exist", req.TaskId),
		}, nil
	}
	if len(s.regisMap) > common.MaxServeJobs {
		hwlog.RunLog.Errorf("registed jobs > %d, jobId=%s will not be registed",
			common.MaxServeJobs, req.TaskId)
		return &pb.Status{
			Code: pb.RespCode_COMMON_ERROR,
			Info: fmt.Sprintf("registed jobs > %d, jobId=%s will not be registed",
				common.MaxServeJobs, req.TaskId),
		}, nil
	}
	s.registry(req.TaskId)
	hwlog.RunLog.Infof("return message is {Code: %d, info: %s}", pb.RespCode_OK, info)
	return &pb.Status{
		Code: pb.RespCode_OK,
		Info: info,
	}, nil
}

func (s *FaultRecoverService) onNotifyStepRetry(req *pb.NotifyStepRetryRequest) *pb.Status {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var controller *EventController
	var exist bool = false
	if controller, exist = s.eventCtl[req.TaskId]; exist && controller != nil {
		return controller.tryChangeState(common.HbmFaultStepRetryMode, common.ReceiveStepRetry,
			[]common.MachineState{common.INIT}, false)
	}
	return &pb.Status{
		Code: pb.RespCode_COMMON_ERROR,
		Info: fmt.Sprintf("unexpected case, taskId=%s, controller not exist", req.TaskId),
	}
}

// NotifyStepRetry request ClusterD whether to take step retry
func (s *FaultRecoverService) NotifyStepRetry(ctx context.Context,
	req *pb.NotifyStepRetryRequest) (*pb.Status, error) {
	baseInfo := fmt.Sprintf("taskId=%s, step=%s", req.TaskId, req.Step)
	hwlog.RunLog.Infof("service receive NotifyStepRetry request, %s", baseInfo)
	if ok, returnInfo := s.checkRegistered(req.TaskId, baseInfo); !ok {
		hwlog.RunLog.Debugf("return message is {Code: %d, info: %s}", pb.RespCode_UNREGISTERED, returnInfo)
		return &pb.Status{
			Code: pb.RespCode_UNREGISTERED,
			Info: returnInfo,
		}, nil
	}
	status := s.onNotifyStepRetry(req)
	if status != nil && status.Code != pb.RespCode_OK {
		s.taskEventFinish(req.TaskId, false)
		hwlog.RunLog.Debugf("return message is {Code: %d, info: %s}", status.Code, status.Info)
		return status, nil
	}
	s.taskEventOpen(req.TaskId, common.NewEventId(randomBytesLen), common.HbmFaultStepRetryMode)
	isHealth, faultRanks := s.jobMgr.GetJobHealthy(req.TaskId)
	var returnInfo string
	if !isHealth {
		hwlog.RunLog.Errorf("job is unhealthy, fault rank is %v", faultRanks)
		returnInfo = fmt.Sprintf("can't take step reCalc: %s, %s", baseInfo, baseInfo)
		hwlog.RunLog.Debugf("return message is {Code: %d, info: %s}", pb.RespCode_DO_NOT_STEP_RECALCULATE, returnInfo)
		s.taskEventFinish(req.TaskId, false)
		return &pb.Status{
			Code: pb.RespCode_DO_NOT_STEP_RECALCULATE,
			Info: returnInfo,
		}, nil
	}
	returnInfo = fmt.Sprintf("can take step reCalc: %s", baseInfo)
	hwlog.RunLog.Debugf("return message is {Code: %d, info: %s}", pb.RespCode_OK, returnInfo)
	return &pb.Status{
		Code: pb.RespCode_OK,
		Info: returnInfo,
	}, nil
}

func (s *FaultRecoverService) onNotifyRetryStatus(req *pb.NotifyRetryStatusRequest) *pb.Status {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var controller *EventController
	var exist bool = false
	if controller, exist = s.eventCtl[req.TaskId]; exist && controller != nil {
		return controller.tryChangeState(common.HbmFaultStepRetryMode, common.ReceiveStepRetryStatus,
			[]common.MachineState{common.ReceiveStepRetry}, false)
	}
	return &pb.Status{
		Code: pb.RespCode_COMMON_ERROR,
		Info: fmt.Sprintf("unexpected case, taskId=%s, controller not exist", req.TaskId),
	}
}

// NotifyRetryStatus notify ClusterD step retry result
func (s *FaultRecoverService) NotifyRetryStatus(ctx context.Context,
	req *pb.NotifyRetryStatusRequest) (*pb.Status, error) {
	info := fmt.Sprintf("taskId=%s, step=%s, status.code=%s, status.info=%s",
		req.TaskId, req.Status, req.Status.Code, req.Status.Info)
	hwlog.RunLog.Infof("service receive NotifyRetryStatus request, %s", info)
	if ok, returnInfo := s.checkRegistered(req.TaskId, info); !ok {
		hwlog.RunLog.Debugf("return message is {Code: %d, info: %s}", pb.RespCode_UNREGISTERED, returnInfo)
		return &pb.Status{
			Code: pb.RespCode_UNREGISTERED,
			Info: returnInfo,
		}, nil
	}
	returnInfo := fmt.Sprintf("ClusterD has receive NotifyRetryStatus request: %s", info)
	hwlog.RunLog.Infof("return message is {Code: %d, info: %s}", pb.RespCode_OK, returnInfo)
	status := s.onNotifyRetryStatus(req)
	if status != nil && status.Code != pb.RespCode_OK {
		s.taskEventFinish(req.TaskId, false)
		hwlog.RunLog.Debugf("return message is {Code: %d, info: %s}", status.Code, status.Info)
		return status, nil
	}
	s.taskEventFinish(req.TaskId, true)
	hwlog.RunLog.Infof("return message is {Code: %d, info: %s}", pb.RespCode_OK, returnInfo)
	return &pb.Status{
		Code: pb.RespCode_OK,
		Info: returnInfo,
	}, nil
}

func (s *FaultRecoverService) taskEventFinish(taskId string, success bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var controller *EventController
	var exist bool = false
	if controller, exist = s.eventCtl[taskId]; exist && controller != nil {
		if success {
			hwlog.RunLog.Infof("event finish close, taskId=%s, eventMode=%s, eventId=%s",
				controller.taskId, common.ModeToString(controller.mode), controller.eventId)
		} else {
			hwlog.RunLog.Infof("event exception close, taskId=%s, eventMode=%s, eventId=%s, curState=%s",
				controller.taskId, common.ModeToString(controller.mode), controller.eventId, common.StateToString(controller.state))
		}
		controller.reset(success)
	}
}

func (s *FaultRecoverService) taskEventOpen(taskId string, eventId string, mode common.RecoverMode) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var controller *EventController
	var exist bool = false
	if controller, exist = s.eventCtl[taskId]; exist && controller != nil {
		controller.openEvent(eventId, mode)
		hwlog.RunLog.Infof("event open, taskId=%s, eventMode=%s, eventId=%s",
			taskId, common.ModeToString(controller.mode), eventId)
	}
}

/*
process recover code
*/

// PublishSignal push signal to send chan
func (s *FaultRecoverService) PublishSignal(signal *pb.ProcessManageSignal, expectStates common.MachineStates) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	controller, exist := s.eventCtl[signal.TaskId]
	if !exist {
		hwlog.RunLog.Errorf("discard signal when publish signal, taskId=%s not registered, signalType=%s",
			signal.TaskId, signal.SignalType)
		return
	}
	if !common.CheckProcessRecoverOpen(controller.taskName+common.MiddleLine+controller.taskId, controller.nameSpace) {
		hwlog.RunLog.Debugf("job not open process recover mode, taskName=%s",
			controller.taskName+common.MiddleLine+controller.taskId)
		return
	}
	if signal.SignalType == common.StopTrainSignalType && len(signal.FaultRankIds) > 0 {
		deviceRankIds := append([]string(nil), signal.FaultRankIds...)
		controller.saveDeviceFaultRankIds(deviceRankIds)

		isPlatForm, _, err := WaitProcessContinue(controller.taskName+common.MiddleLine+controller.taskId,
			controller.nameSpace)
		if isPlatForm {
			if err != nil {
				hwlog.RunLog.Errorf("platForm err: %v, reset process", err)
				s.taskEventFinish(signal.TaskId, false)
				return
			}
			hardRanks := controller.getHardFaultRankIds()
			softRanks := controller.getSoftwareFaultRankIds()
			allFaultRanks := util.RemoveSliceDuplicateElement(append(hardRanks, softRanks...))
			err = UpdateProcessConfirmFault(controller.taskName+common.MiddleLine+controller.taskId,
				controller.nameSpace, allFaultRanks)
			if err != nil {
				hwlog.RunLog.Errorf("update platform ProcessConfirmFault err: %v, reset process", err)
				s.taskEventFinish(signal.TaskId, false)
				return
			}
		}
	}
	afterCheckOrder(signal, expectStates, controller)

}

func afterCheckOrder(signal *pb.ProcessManageSignal, expectStates common.MachineStates, controller *EventController) {
	if common.CheckOrder(controller.state, expectStates) {
		select {
		case controller.signalChan <- signal:
			hwlog.RunLog.Infof("signal in queue, taskId=%s, signalType=%s",
				signal.TaskId, signal.SignalType)
		case <-time.After(time.Second):
			hwlog.RunLog.Warnf("signal chan should not have unsent signal when publish a signal, "+
				"discard it. taskId=%s, signalType=%s",
				signal.TaskId, signal.SignalType)
		}
	} else {
		hwlog.RunLog.Errorf("discard signal when publish signal cause order mixed. "+
			"state=%s, expectStates=%s, taskId=%s, signalType=%s",
			common.StateToString(controller.state), expectStates.String(), signal.TaskId, signal.SignalType)
	}
}

func (s *FaultRecoverService) onSignalSent(signal *pb.ProcessManageSignal) *pb.Status {
	s.lock.RLock()
	defer s.lock.RUnlock()
	controller, exist := s.eventCtl[signal.TaskId]
	if !exist {
		return &pb.Status{Code: pb.RespCode_COMMON_ERROR, Info: "taskId=%s not event controller map"}
	}
	switch signal.SignalType {
	case common.StopTrainSignalType:
		return controller.tryChangeState(common.ProcessFaultRecoverMode, common.SentStopTrain,
			[]common.MachineState{common.INIT}, false)
	case common.GlobalFaultSignalType:
		return controller.tryChangeState(common.ProcessFaultRecoverMode, common.SentGlobalFault,
			[]common.MachineState{common.ReceiveStopFinish}, false)
	case common.ChangeStrategySignalType:
		n := len(controller.latestStrategy)
		if n == 0 {
			return &pb.Status{Code: pb.RespCode_COMMON_ERROR, Info: "unexpect case, latestStrategy must > 0"}
		}
		lastStrategy := controller.latestStrategy[n-1]
		if lastStrategy == common.ProcessArfStrategy {
			s.jobMgr.ListenTaskScheduleResult(signal.TaskId, lastStrategy)
			return controller.tryChangeState(common.ProcessFaultRecoverMode, common.ListenOnlineRecoverStatus,
				[]common.MachineState{common.ReceiveSupportStrategy}, false)
		}
		if lastStrategy == common.ProcessDumpStrategy {
			return controller.tryChangeState(common.ProcessFaultRecoverMode, common.ListenCheckPointSave,
				[]common.MachineState{common.ReceiveSupportStrategy, common.ListenOnlineRecoverStatus}, false)
		}
		if lastStrategy == common.ProcessExitStrategy {
			status := controller.tryChangeState(common.ProcessFaultRecoverMode, common.StartListenSchedule,
				[]common.MachineState{common.ReceiveSupportStrategy}, false)
			if status.Code == pb.RespCode_OK {
				s.jobMgr.ListenTaskScheduleResult(signal.TaskId, lastStrategy)
			}
			return status
		}
		return &pb.Status{Code: pb.RespCode_COMMON_ERROR,
			Info: fmt.Sprintf("strategy=%s, unexpect case", lastStrategy)}
	default:
		return &pb.Status{Code: pb.RespCode_COMMON_ERROR, Info: "unknown signal type"}
	}
}

func handlePlatStrategy(ctl *EventController, signal *pb.ProcessManageSignal) {
	platStrategy := GetPlatStrategy(ctl.taskName+common.MiddleLine+ctl.taskId,
		ctl.nameSpace)
	platLevel := common.StringToLevel(platStrategy)
	hwlog.RunLog.Infof("platform strategy=%s, level=%d", platStrategy, platLevel)
	if len(ctl.latestStrategy) == 1 && platStrategy != "" && platLevel >= 0 {
		curLevel := common.StringToLevel(signal.ChangeStrategy)
		for curLevel > platLevel {
			platLevel++
		}
		signal.ChangeStrategy = common.LevelToString(platLevel)
		hwlog.RunLog.Infof("final plat strategy=%s", signal.ChangeStrategy)
		if signal.ChangeStrategy == common.ProcessArfStrategy {
			WaitRankTableReady(ctl.taskName+common.MiddleLine+ctl.taskId,
				ctl.nameSpace)
		}
	}
}

func (s *FaultRecoverService) onSignalOutQueue(signal *pb.ProcessManageSignal) *pb.Status {
	s.lock.RLock()
	defer s.lock.RUnlock()
	stateMatch := false

	controller, exist := s.eventCtl[signal.TaskId]
	if !exist {
		return &pb.Status{Code: pb.RespCode_COMMON_ERROR, Info: "taskId=%s not in event controller map"}
	}
	switch signal.SignalType {
	case common.StopTrainSignalType:
		stateMatch = common.CheckOrder(controller.state, []common.MachineState{common.INIT})
	case common.GlobalFaultSignalType:
		stateMatch = common.CheckOrder(controller.state, []common.MachineState{common.ReceiveStopFinish})
		if stateMatch {
			status, done := globalFault(signal, controller)
			if done {
				return status
			}
		}
	case common.ChangeStrategySignalType:
		handlePlatStrategy(controller, signal)
		stateMatch = common.CheckOrder(controller.state,
			[]common.MachineState{common.ListenOnlineRecoverStatus, common.ReceiveSupportStrategy})
	default:
		return &pb.Status{Code: pb.RespCode_COMMON_ERROR, Info: "unknown signal type"}
	}
	if !stateMatch {
		return &pb.Status{Code: pb.RespCode_ORDER_MIXED, Info: "order mixed, signal queue out ignore"}
	}
	return &pb.Status{Code: pb.RespCode_OK, Info: "signal out queue, prepare to send"}
}

func globalFault(signal *pb.ProcessManageSignal, controller *EventController) (*pb.Status, bool) {
	hardRanks := controller.getHardFaultRankIds()
	softRanks := controller.getSoftwareFaultRankIds()
	allFaultRanks := util.RemoveSliceDuplicateElement(append(hardRanks, softRanks...))

	isPlatForm, _, err := WaitProcessContinue(controller.taskName+common.MiddleLine+controller.taskId,
		controller.nameSpace)
	if isPlatForm {
		hwlog.RunLog.Info("platFrom mode process recover")
		if err != nil {
			hwlog.RunLog.Errorf("platForm err: %v, reset process", err)
			controller.reset(false)
			return &pb.Status{Code: pb.RespCode_COMMON_ERROR,
				Info: fmt.Sprintf("platForm err: %v, reset process", err)}, true
		}
		platFaultResult, err := WaitProcessResultFault(controller.taskName+common.MiddleLine+controller.taskId,
			controller.nameSpace)
		if err != nil {
			controller.reset(false)
			return &pb.Status{Code: pb.RespCode_COMMON_ERROR,
				Info: fmt.Sprintf("WaitProcessResultFault err: %v, reset process", err)}, true
		}
		allFaultRanks = platFaultResult
	}
	if _, err := common.RetryWriteResetCM(controller.taskName, controller.nameSpace,
		allFaultRanks, "fault"); err != nil {
		controller.reset(false)
		return &pb.Status{Code: pb.RespCode_COMMON_ERROR, Info: "taskId=%s write reset info failed"}, true
	}
	signal.FaultRankIds = allFaultRanks
	hwlog.RunLog.Infof("global fault signal: %v", signal.FaultRankIds)
	return nil, false
}

func (s *FaultRecoverService) listenSendChannel(jobId string, sendChan chan *pb.ProcessManageSignal,
	stream pb.Recover_SubscribeProcessManageSignalServer) {
	for signal := range sendChan {
		signalInfo := fmt.Sprintf("taskId=%s, eventId=%s, signalType=%s, actions=%v, faultRanks=%v, changeStrategy=%s",
			signal.TaskId, signal.Uuid, signal.SignalType, signal.Actions, signal.FaultRankIds, signal.ChangeStrategy)
		hwlog.RunLog.Infof("signal out queue, %s", signalInfo)
		if status := s.onSignalOutQueue(signal); status != nil && status.Code != pb.RespCode_OK {
			hwlog.RunLog.Debugf("discard signal on signal out queue, signalInfo=%s, code=%s, info=%s",
				signalInfo, status.Code.String(), status.Info)
			continue
		}
		if signal.SignalType == common.StopTrainSignalType {
			signal.Uuid = common.NewEventId(randomBytesLen)
		}
		if err := common.SendRetry(stream, signal, retryTimes); err != nil {
			hwlog.RunLog.Errorf("signal send error: %v, client maybe offline, unregister task", err)
			s.taskEventFinish(jobId, false)
			break
		}
		if signal.SignalType == common.StopTrainSignalType {
			s.taskEventOpen(signal.TaskId, signal.Uuid, common.ProcessFaultRecoverMode)
		}
		if status := s.onSignalSent(signal); status != nil && status.Code != pb.RespCode_OK {
			hwlog.RunLog.Errorf("reset event after signal sent, taskId=%s, eventId=%s, code=%s, info=%s",
				signal.TaskId, signal.Uuid, status.Code.String(), status.Info)
			s.taskEventFinish(jobId, false)
			break
		}
		hwlog.RunLog.Infof("signal send success, %s", signalInfo)
	}
	hwlog.RunLog.Infof("listen signal break, taskId=%s", jobId)
}

// SubscribeProcessManageSignal subscribe process manage signal from ClusterD
func (s *FaultRecoverService) SubscribeProcessManageSignal(request *pb.ClientInfo,
	stream pb.Recover_SubscribeProcessManageSignalServer) error {
	hwlog.RunLog.Infof("receive Subscribe signal request, taskId=%s, rule=%s", request.TaskId, request.Role)
	requestInfo := fmt.Sprintf("taskId=%s, rule=%s", request.TaskId, request.Role)
	if ok, returnInfo := s.checkRegistered(request.TaskId, requestInfo); !ok {
		hwlog.RunLog.Errorf("return message is {Code: %d, info: %s}", pb.RespCode_UNREGISTERED, returnInfo)
		return errors.New(returnInfo)
	}
	var sendChan chan *pb.ProcessManageSignal
	s.lock.Lock()
	if controller, exist := s.eventCtl[request.TaskId]; exist && controller != nil {
		controller.reset(true)
		sendChan = controller.signalChan
	} else {
		if !s.jobMgr.JobExist(request.TaskId) {
			hwlog.RunLog.Warnf("jobId=%s not exist", request.TaskId)
			return fmt.Errorf("jobId=%s not exist", request.TaskId)
		}
		taskName, namespace := s.jobMgr.GetJobNameAndNameSpace(request.TaskId)
		controller = NewEventController(request.TaskId, taskName, namespace)
		s.eventCtl[request.TaskId] = controller
		s.regisMap[request.TaskId] = struct{}{}
		sendChan = s.eventCtl[request.TaskId].signalChan
	}
	s.lock.Unlock()
	s.listenSendChannel(request.TaskId, sendChan, stream)
	return nil
}

func (s *FaultRecoverService) onReportStopComplete(request *pb.StopCompleteRequest) *pb.Status {
	s.lock.RLock()
	defer s.lock.RUnlock()
	controller, exist := s.eventCtl[request.TaskId]
	if !exist {
		return &pb.Status{Code: pb.RespCode_COMMON_ERROR, Info: "taskId=%s not in event controller map"}
	}
	status := controller.tryChangeState(common.ProcessFaultRecoverMode, common.ReceiveStopFinish,
		[]common.MachineState{common.SentStopTrain}, false)
	if status.Code != pb.RespCode_OK {
		return status
	}
	s.jobMgr.NotifySignalSend(&common.Notifier{
		CreateTimeStamp: time.Now().Unix(),
		ProcessManageSignal: pb.ProcessManageSignal{
			Uuid:       s.getTaskEventId(request.TaskId),
			TaskId:     request.TaskId,
			SignalType: common.GlobalFaultSignalType,
			Actions:    globalFaultActions,
		},
	})
	return &pb.Status{
		Code: pb.RespCode_OK,
		Info: fmt.Sprintf("server will send global fault ranks later"),
	}
}

func (s *FaultRecoverService) getTaskEventId(taskId string) string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if controller, exist := s.eventCtl[taskId]; exist {
		return controller.eventId
	}
	return common.UnknownEventId
}

func (s *FaultRecoverService) getTaskSoftRankIds(taskId string) []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if controller, exist := s.eventCtl[taskId]; exist {
		return controller.getSoftwareFaultRankIds()
	}
	return nil
}

func (s *FaultRecoverService) getTaskHardRankIds(taskId string) []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if controller, exist := s.eventCtl[taskId]; exist {
		return controller.getSoftwareFaultRankIds()
	}
	return nil
}

// ReportStopComplete report stop process result
func (s *FaultRecoverService) ReportStopComplete(ctx context.Context,
	request *pb.StopCompleteRequest) (*pb.Status, error) {
	requestInfo := fmt.Sprintf("taskId=%s, status.code=%d, status.info=%s, faultList=%v",
		request.TaskId, request.Status.Code, request.Status.Info, request.FaultRankIds)
	hwlog.RunLog.Infof("receive ReportStopComplete, info={%s}", requestInfo)
	if ok, returnInfo := s.checkRegistered(request.TaskId, requestInfo); !ok {
		hwlog.RunLog.Debugf("return message is {Code: %d, info: %s}", pb.RespCode_UNREGISTERED, returnInfo)
		return &pb.Status{
			Code: pb.RespCode_UNREGISTERED,
			Info: returnInfo,
		}, nil
	}
	if status := s.onReportStopComplete(request); status != nil && status.Code != pb.RespCode_OK {
		hwlog.RunLog.Errorf("onReportStopComplete error, status.code=%s, status.info=%s", status.Code.String(), status.Info)
		s.taskEventFinish(request.TaskId, false)
		return &pb.Status{
			Code: status.Code,
			Info: status.Info,
		}, nil
	}
	return &pb.Status{Code: pb.RespCode_OK,
			Info: fmt.Sprintf("server receive ReportStopComplete, reqestInfo = %s", requestInfo)},
		nil
}

func (s *FaultRecoverService) onReportRecoverStrategy(request *pb.RecoverStrategyRequest,
	strategies []string) *pb.Status {
	s.lock.RLock()
	defer s.lock.RUnlock()
	controller, exist := s.eventCtl[request.TaskId]
	if !exist || controller == nil {
		return &pb.Status{Code: pb.RespCode_COMMON_ERROR,
			Info: "taskId=%s not in event controller map, upper to pod reschedule"}
	}
	status := controller.tryChangeState(common.ProcessFaultRecoverMode, common.ReceiveSupportStrategy,
		[]common.MachineState{common.SentGlobalFault}, false)
	if status.Code != pb.RespCode_OK {
		return status
	}
	strategy := chooseBestStrategy(strategies)
	controller.appendStrategy(strategy)
	hwlog.RunLog.Infof("decide strategy %s to task %s", strategy, request.TaskId)
	s.jobMgr.NotifySignalSend(&common.Notifier{
		CreateTimeStamp: time.Now().Unix(),
		ProcessManageSignal: pb.ProcessManageSignal{
			Uuid:           controller.eventId,
			TaskId:         controller.taskId,
			SignalType:     common.ChangeStrategySignalType,
			Actions:        changeStrategyActions,
			FaultRankIds:   nil,
			ChangeStrategy: strategy,
		},
	})
	return &pb.Status{Code: pb.RespCode_OK,
		Info: fmt.Sprintf("decide send stratey=%s", strategy)}
}

// ReportRecoverStrategy report supported recover strategy to ClusterD
func (s *FaultRecoverService) ReportRecoverStrategy(ctx context.Context,
	request *pb.RecoverStrategyRequest) (*pb.RecoverStrategyResponse, error) {
	requestInfo := fmt.Sprintf("taskId=%s, faultList=%v, strategyList=%v",
		request.TaskId, request.FaultRankIds, request.Strategies)
	hwlog.RunLog.Infof("receive ReportRecoverStrategy, info={%s}", requestInfo)
	if ok, returnInfo := s.checkRegistered(request.TaskId, requestInfo); !ok {
		hwlog.RunLog.Debugf("return message is {Code: %d, info: %s}", pb.RespCode_UNREGISTERED, returnInfo)
		return &pb.RecoverStrategyResponse{
			Status: &pb.Status{
				Code: pb.RespCode_UNREGISTERED,
				Info: requestInfo,
			},
			Strategy: "",
		}, nil
	}
	if status := s.onReportRecoverStrategy(request, request.Strategies); status != nil && status.Code != pb.RespCode_OK {
		hwlog.RunLog.Errorf("onReportRecoverStrategy error, status.code=%s, status.info=%s",
			status.Code.String(), status.Info)
		s.taskEventFinish(request.TaskId, false)
		return &pb.RecoverStrategyResponse{
			Status: &pb.Status{
				Code: status.Code,
				Info: status.Info,
			},
			Strategy: "",
		}, nil
	}
	return &pb.RecoverStrategyResponse{
		Status: &pb.Status{
			Code: pb.RespCode_OK,
			Info: fmt.Sprintf("server receive ReportRecoverStrategy, request info = {%s}", requestInfo),
		},
		Strategy: "",
	}, nil
}

func chooseBestStrategy(strategies []string) string {
	if len(strategies) == 0 {
		hwlog.RunLog.Error("no usage strategy, just use exit strategy")
		return common.ProcessExitStrategy
	}
	var levelList []int
	for _, str := range strategies {
		level := common.StringToLevel(str)
		levelList = append(levelList, level)
	}
	minLevel := math.MaxInt
	for _, level := range levelList {
		if level < minLevel {
			minLevel = level
		}
	}
	if minLevel == -1 {
		hwlog.RunLog.Error("strategy name not legal, just use exit strategy")
		return common.ProcessExitStrategy
	}
	return common.LevelToString(minLevel)
}

func (s *FaultRecoverService) upgradeRecoverStrategy(controller *EventController, strategy string) {
	upperStrategyLevel := common.StringToLevel(strategy) + 1
	if upperStrategyLevel > common.ExitRecoverLevel {
		upperStrategyLevel = common.ExitRecoverLevel
	}
	upperStrategy := common.LevelToString(upperStrategyLevel)
	controller.appendStrategy(upperStrategy)
	// notify change strategy
	if len(controller.latestStrategy) == common.MaxChangeStrategyTimes {
		s.jobMgr.NotifySignalSend(&common.Notifier{
			CreateTimeStamp: time.Now().Unix(),
			ProcessManageSignal: pb.ProcessManageSignal{
				Uuid:           controller.eventId,
				TaskId:         controller.taskId,
				SignalType:     common.ChangeStrategySignalType,
				Actions:        changeStrategyActions,
				ChangeStrategy: upperStrategy,
			},
		})
	}
}

func (s *FaultRecoverService) handleRecoverStatus(controller *EventController, statusCode pb.RespCode) *pb.Status {
	if len(controller.latestStrategy) == 0 {
		return &pb.Status{Code: pb.RespCode_COMMON_ERROR,
			Info: fmt.Sprintf("latest strategy length exception, "+
				"latestStrategy=%v, lenth=%d", controller.latestStrategy, len(controller.latestStrategy))}
	}
	strategy := controller.latestStrategy[len(controller.latestStrategy)-1]
	switch controller.state {
	case common.ListenOnlineRecoverStatus:
		if statusCode == pb.RespCode_OK && s.jobMgr.IsJobRunning(controller.taskId) {
			hwlog.RunLog.Infof("event finish close, taskId=%s, eventMode=%s, eventId=%s",
				controller.taskId, common.ModeToString(controller.mode), controller.eventId)
			controller.reset(true)
			return &pb.Status{Code: pb.RespCode_OK,
				Info: fmt.Sprintf("process online recover success, taskId=%s", controller.taskId)}
		}
		s.upgradeRecoverStrategy(controller, strategy)
		return &pb.Status{Code: pb.RespCode_OK,
			Info: fmt.Sprintf("process online recover fail, taskId=%s", controller.taskId)}
	case common.ListenCheckPointSave:
		status := controller.tryChangeState(common.ProcessFaultRecoverMode, common.StartListenSchedule,
			[]common.MachineState{common.ListenCheckPointSave}, false)
		if status != nil && status.Code != pb.RespCode_OK {
			return status
		}
		s.jobMgr.ListenTaskScheduleResult(controller.taskId, strategy)
		hwlog.RunLog.Infof("receive check point save result, %s", statusCode.String())
		hwlog.RunLog.Infof("event dump close, taskId=%s, eventMode=%s, eventId=%s, curState=%s",
			controller.taskId, common.ModeToString(controller.mode), controller.eventId, common.StateToString(controller.state))
		return &pb.Status{
			Code: pb.RespCode_OK,
			Info: "receive check point save, startListen schedule result",
		}
	default:
		return &pb.Status{Code: pb.RespCode_COMMON_ERROR, Info: "default case is unexpected"}
	}
}

func (s *FaultRecoverService) onReportRecoverStatus(request *pb.RecoverStatusRequest) *pb.Status {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if controller, exist := s.eventCtl[request.TaskId]; exist && controller != nil {
		return s.handleRecoverStatus(controller, request.Status.Code)
	}
	return &pb.Status{Code: pb.RespCode_COMMON_ERROR, Info: "taskId=%s not in event controller map"}
}

// ReportRecoverStatus report recover result
func (s *FaultRecoverService) ReportRecoverStatus(ctx context.Context,
	request *pb.RecoverStatusRequest) (*pb.Status, error) {
	requestInfo := fmt.Sprintf("taskId=%s, status.code=%d, status.info=%s",
		request.TaskId, request.Status.Code, request.Status.Info)
	hwlog.RunLog.Infof("receive ReportRecoverStatus, info={%s}", requestInfo)
	if ok, returnInfo := s.checkRegistered(request.TaskId, requestInfo); !ok {
		hwlog.RunLog.Debugf("return message is {Code: %d, info: %s}", pb.RespCode_UNREGISTERED, returnInfo)
		return &pb.Status{
			Code: pb.RespCode_UNREGISTERED,
			Info: returnInfo,
		}, nil
	}
	if status := s.onReportRecoverStatus(request); status != nil && status.Code != pb.RespCode_OK {
		hwlog.RunLog.Errorf("onReportRecoverStatus error, status.code=%s, status.info=%s",
			status.Code.String(), status.Info)
		s.taskEventFinish(request.TaskId, false)
		return &pb.Status{
			Code: status.Code,
			Info: status.Info,
		}, nil
	}
	return &pb.Status{
		Code: pb.RespCode_OK,
		Info: fmt.Sprintf("server receive ReportRecoverStatus, request info = {%s}", requestInfo),
	}, nil
}

func (s *FaultRecoverService) onReportProcessFault(request *pb.ProcessFaultRequest) *pb.Status {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if controller, exist := s.eventCtl[request.TaskId]; exist {
		controller.saveSoftwareFaultRankIds(
			common.GetFaultRankIdsInSameNode(request.FaultRankIds, s.jobMgr.GetJobDeviceNumPerNode(request.TaskId)))
		return &pb.Status{
			Code: pb.RespCode_OK,
			Info: "save software fault rank success",
		}
	}
	return &pb.Status{Code: pb.RespCode_COMMON_ERROR, Info: "taskId=%s not in event controller map"}
}

// ReportProcessFault report soft fault ranks to ClusterD
func (s *FaultRecoverService) ReportProcessFault(ctx context.Context,
	request *pb.ProcessFaultRequest) (*pb.Status, error) {
	requestInfo := fmt.Sprintf("taskId=%s, faultList=%v", request.TaskId, request.FaultRankIds)
	hwlog.RunLog.Infof("receive ReportProcessFault, info={%s}", requestInfo)
	if ok, returnInfo := s.checkRegistered(request.TaskId, requestInfo); !ok {
		hwlog.RunLog.Debugf("return message is {Code: %d, info: %s}", pb.RespCode_UNREGISTERED, returnInfo)
		return &pb.Status{
			Code: pb.RespCode_UNREGISTERED,
			Info: returnInfo,
		}, nil
	}
	if status := s.onReportProcessFault(request); status != nil && status.Code != pb.RespCode_OK {
		return &pb.Status{
			Code: status.Code,
			Info: status.Info,
		}, nil
	}
	s.jobMgr.NotifySignalSend(&common.Notifier{
		CreateTimeStamp: time.Now().Unix(),
		ProcessManageSignal: pb.ProcessManageSignal{
			Uuid:         "",
			TaskId:       request.TaskId,
			SignalType:   common.StopTrainSignalType,
			Actions:      stopTrainActions,
			FaultRankIds: nil,
		},
	})
	return &pb.Status{
		Code: pb.RespCode_OK,
		Info: fmt.Sprintf("server receive ReportRecoverStatus, request info = {%s}", requestInfo),
	}, nil
}

func (s *FaultRecoverService) DeleteJob(jobId string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	hwlog.RunLog.Infof("current serve jobs=%d, prepare delete jobId=%s", len(s.regisMap), jobId)
	if s.eventCtl != nil {
		delete(s.eventCtl, jobId)
	}
	if s.regisMap != nil {
		delete(s.regisMap, jobId)
	}
	hwlog.RunLog.Infof("serve jobs=%d after delete jobId=%s", len(s.regisMap), jobId)
}
