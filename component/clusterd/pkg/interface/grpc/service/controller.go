// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package service a series of service function
package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/application/job"
	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/grpc/common"
	"clusterd/pkg/interface/grpc/pb"
	"clusterd/pkg/interface/kube"
)

var (
	retryTimes            = 3
	randomLen             = 16
	reportTimeoutMinutes  = 15
	faultFlushSeconds     = 10
	eventChanLength       = 10
	stopTrainActions      = []string{"stop_train"}
	globalFaultActions    = []string{"on_global_rank"}
	changeStrategyActions = []string{"change_strategy"}
)

// EventController is recover event controller
type EventController struct {
	jobInfo                   common.JobBaseInfo
	faultFlushing             bool
	keepAliveSecond           int
	uuid                      string
	events                    chan string
	latestStrategy            []string
	latestRecoverResult       []*pb.RecoverStatusRequest
	agentReportStrategies     []string
	platStrategies            []string
	signalChan                chan *pb.ProcessManageSignal
	cacheNormalFault          []*pb.FaultRank
	cacheUceFault             []*pb.FaultRank
	controllerContext         context.Context
	ctxCancelFunc             context.CancelFunc
	serviceContext            context.Context
	state                     *common.StateMachine
	reportStopCompleteChan    chan *pb.StopCompleteRequest
	reportRecoverStrategyChan chan *pb.RecoverStrategyRequest
	reportStatusChan          chan *pb.RecoverStatusRequest
	scheduleResultChan        chan bool
	scheduleResult            bool
	haveGetScheduleResult     bool
	lock                      sync.RWMutex
}

// NewEventController return pointer of EventController
func NewEventController(jobInfo common.JobBaseInfo, keepAlive int, serviceCtx context.Context) *EventController {
	ctl := &EventController{
		jobInfo:                   jobInfo,
		faultFlushing:             false,
		scheduleResult:            false,
		haveGetScheduleResult:     false,
		keepAliveSecond:           keepAlive,
		uuid:                      "",
		events:                    make(chan string, eventChanLength),
		latestStrategy:            []string{},
		latestRecoverResult:       []*pb.RecoverStatusRequest{},
		agentReportStrategies:     []string{},
		platStrategies:            []string{},
		signalChan:                make(chan *pb.ProcessManageSignal, 1),
		reportStopCompleteChan:    make(chan *pb.StopCompleteRequest, 1),
		reportRecoverStrategyChan: make(chan *pb.RecoverStrategyRequest, 1),
		reportStatusChan:          make(chan *pb.RecoverStatusRequest, 1),
		scheduleResultChan:        make(chan bool, 1),
		cacheNormalFault:          []*pb.FaultRank{},
		cacheUceFault:             []*pb.FaultRank{},
		serviceContext:            serviceCtx,
		lock:                      sync.RWMutex{},
	}
	var rules []common.TransRule = ctl.getBaseRules()
	ctl.state = common.NewStateMachine(common.InitState, rules)
	ctl.controllerContext, ctl.ctxCancelFunc = context.WithCancel(ctl.serviceContext)
	return ctl
}

func (ctl *EventController) saveCacheFault(faults []*pb.FaultRank) {
	mergedFaults := common.RemoveSliceDuplicateFaults(faults)
	for _, fault := range mergedFaults {
		if fault.FaultType == "1" {
			ctl.cacheNormalFault = append(ctl.cacheNormalFault, fault)
		} else {
			ctl.cacheUceFault = append(ctl.cacheUceFault, fault)
		}
	}
}

func (ctl *EventController) reset(useLock bool) {
	if useLock {
		ctl.lock.Lock()
		defer ctl.lock.Unlock()
	}
	if ctl.ctxCancelFunc != nil {
		ctl.ctxCancelFunc()
	}
	ctl.faultFlushing = false
	ctl.uuid = ""
	ctl.latestStrategy = ctl.latestStrategy[:0]
	close(ctl.events)
	close(ctl.signalChan)
	close(ctl.reportStopCompleteChan)
	close(ctl.reportRecoverStrategyChan)
	close(ctl.reportStatusChan)
	close(ctl.scheduleResultChan)
	ctl.events = make(chan string, eventChanLength)
	ctl.signalChan = make(chan *pb.ProcessManageSignal, 1)
	ctl.reportStopCompleteChan = make(chan *pb.StopCompleteRequest, 1)
	ctl.reportRecoverStrategyChan = make(chan *pb.RecoverStrategyRequest, 1)
	ctl.reportStatusChan = make(chan *pb.RecoverStatusRequest, 1)
	ctl.scheduleResultChan = make(chan bool, 1)
	ctl.scheduleResult = false
	ctl.haveGetScheduleResult = false
	ctl.cacheUceFault = ctl.cacheUceFault[:0]
	ctl.cacheNormalFault = ctl.cacheNormalFault[:0]
	ctl.latestRecoverResult = ctl.latestRecoverResult[:0]
	ctl.agentReportStrategies = ctl.agentReportStrategies[:0]
	ctl.platStrategies = ctl.platStrategies[:0]
	ctl.state.Reset()
	ctl.controllerContext, ctl.ctxCancelFunc = context.WithCancel(ctl.serviceContext)
	go ctl.listenEvent()
	go ctl.keepAlive()
}

func (ctl *EventController) selectKeepAlive(ctx context.Context, sendChan chan *pb.ProcessManageSignal) bool {
	if sendChan == nil {
		return true
	}
	select {
	case <-ctx.Done():
		return true
	case <-time.After(time.Duration(ctl.keepAliveSecond) * time.Second):
		signal := &pb.ProcessManageSignal{
			Uuid:       ctl.uuid,
			JobId:      ctl.jobInfo.JobId,
			SignalType: common.KeepAliveSignalType,
		}
		select {
		case sendChan <- signal:
			return false
		case <-ctx.Done():
			return true
		}
	}
}

func (ctl *EventController) keepAlive() {
	hwlog.RunLog.Infof("listen new keep-alive, jobId=%s", ctl.jobInfo.JobId)
	ctx, sendChan := ctl.getPreControllerAndSignalChan()
	exit := false
	for {
		exit = ctl.selectKeepAlive(ctx, sendChan)
		if exit {
			break
		}
	}
}

func (ctl *EventController) supportRetryStrategy() bool {
	mindXConfiged := false
	for _, strategy := range ctl.jobInfo.MindXConfigStrategies {
		if strategy == common.ProcessRetryStrategyName {
			mindXConfiged = true
			break
		}
	}
	if !mindXConfiged || !ctl.jobInfo.PlatFormMode {
		return mindXConfiged
	}
	for _, strategy := range ctl.platStrategies {
		if strategy == common.ProcessRetryStrategyName {
			return true
		}
	}
	return false
}

func (ctl *EventController) supportRecoverStrategy() bool {
	mindXConfiged := false
	for _, strategy := range ctl.jobInfo.MindXConfigStrategies {
		if strategy == common.ProcessRecoverStrategyName {
			mindXConfiged = true
			break
		}
	}
	if !mindXConfiged {
		return false
	}
	agentSupport := false
	for _, strategy := range ctl.agentReportStrategies {
		if strategy == common.ProcessRecoverStrategyName {
			agentSupport = true
			break
		}
	}
	if !agentSupport || !ctl.jobInfo.PlatFormMode {
		return agentSupport
	}
	for _, strategy := range ctl.platStrategies {
		if strategy == common.ProcessRecoverStrategyName {
			return true
		}
	}
	return false
}

func (ctl *EventController) supportDumpStrategy() bool {
	mindXConfiged := false
	for _, strategy := range ctl.jobInfo.MindXConfigStrategies {
		if strategy == common.ProcessDumpStrategyName {
			mindXConfiged = true
			break
		}
	}
	if !mindXConfiged {
		return false
	}
	agentSupport := false
	for _, strategy := range ctl.agentReportStrategies {
		if strategy == common.ProcessDumpStrategyName {
			agentSupport = true
			break
		}
	}
	if !ctl.jobInfo.PlatFormMode || !agentSupport {
		return agentSupport
	}
	for _, strategy := range ctl.platStrategies {
		if strategy == common.ProcessRecoverStrategyName {
			return true
		}
	}
	return false
}

func (ctl *EventController) addEvent(event string) {
	if !ctl.state.RuleCheck(ctl.state.GetState(), event) {
		hwlog.RunLog.Errorf("event add fail, order mix, jobId=%s, uuid=%s, event=%s",
			ctl.jobInfo.JobId, ctl.uuid, event)
		return
	}
	select {
	case <-ctl.controllerContext.Done():
		hwlog.RunLog.Warnf("event add fail, controller context canceled, jobId=%s, uuid=%s, event=%s",
			ctl.jobInfo.JobId, ctl.uuid, event)
	case ctl.events <- event:
		hwlog.RunLog.Infof("jobId=%s uuid=%s state is %s, event=%s enqueue success",
			ctl.jobInfo.JobId, ctl.uuid, ctl.state.GetState(), event)
	case <-time.After(time.Second):
		ctl.reset(false)
	}
}

func (ctl *EventController) getPreControllerAndEventChan() (context.Context, chan string) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.events
}

func (ctl *EventController) selectEventChan(ctx context.Context, eventChan chan string) bool {
	if eventChan == nil {
		return true
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Infof("controller context done, break listen event, jobId=%s", ctl.jobInfo.JobId)
		return true
	case event, ok := <-eventChan:
		if ok {
			nextEvent, code, err := ctl.trigger(event)
			hwlog.RunLog.Infof("jobId=%s's action path = {%s}", ctl.jobInfo.JobId, ctl.state.GetPathGraph())
			if err != nil {
				hwlog.RunLog.Errorf("jobId=%s trigger error, code=%d, err=%v", ctl.jobInfo.JobId, code, err)
				ctl.reset(true)
				return true
			}
			if nextEvent != "" {
				ctl.addEvent(nextEvent)
			}
			return false
		} else {
			hwlog.RunLog.Infof("event channel closed, break listen event, jobId=%s", ctl.jobInfo.JobId)
			return true
		}
	}
}

func (ctl *EventController) listenEvent() {
	hwlog.RunLog.Infof("start listen a new event, jobId=%s", ctl.jobInfo.JobId)
	ctx, eventChan := ctl.getPreControllerAndEventChan()
	exit := false
	for {
		exit = ctl.selectEventChan(ctx, eventChan)
		if exit {
			break
		}
	}
}

func (ctl *EventController) getPreControllerAndSignalChan() (context.Context, chan *pb.ProcessManageSignal) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.signalChan
}

func (ctl *EventController) handleSendResult(signal *pb.ProcessManageSignal, err error) {
	if signal.SignalType == common.KillMasterSignalType {
		ctl.addEvent(common.FinishEvent)
		return
	}
	if err != nil {
		ctl.addEvent(common.NotifyFailEvent)
		return
	}
	if signal.SignalType != common.ChangeStrategySignalType {
		ctl.addEvent(common.NotifySuccessEvent)
		return
	}
	if signal.ChangeStrategy == common.ProcessRetryStrategyName {
		ctl.addEvent(common.NotifyRetrySuccessEvent)
	} else if signal.ChangeStrategy == common.ProcessRecoverStrategyName {
		ctl.addEvent(common.NotifyRecoverSuccessEvent)
	} else if signal.ChangeStrategy == common.ProcessDumpStrategyName {
		ctl.addEvent(common.NotifyDumpSuccessEvent)
	} else if signal.ChangeStrategy == common.ProcessExitStrategyName {
		ctl.addEvent(common.NotifyExitSuccessEvent)
	} else {
		hwlog.RunLog.Errorf("unsupported strategy=%s, jobId=%s",
			signal.ChangeStrategy, signal.JobId)
	}
}

func (ctl *EventController) selectSendChannel(ctx context.Context, sendChan chan *pb.ProcessManageSignal,
	stream pb.Recover_SubscribeProcessManageSignalServer) bool {
	if sendChan == nil {
		return true
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Infof("context done, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
		return true
	case signal, ok := <-sendChan:
		if ok {
			err := common.SendRetry(stream, signal, retryTimes)
			if signal.SignalType != common.KeepAliveSignalType {
				ctl.handleSendResult(signal, err)
			}
			return false
		} else {
			hwlog.RunLog.Infof("sendChan closed, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
			return true
		}
	}
}

func (ctl *EventController) listenSendChannel(stream pb.Recover_SubscribeProcessManageSignalServer) {
	ctl.reset(true)
	ctx, sendChan := ctl.getPreControllerAndSignalChan()
	hwlog.RunLog.Infof("start listen a new send channel, jobId=%s", ctl.jobInfo.JobId)
	exit := false
	for {
		exit = ctl.selectSendChannel(ctx, sendChan, stream)
		if exit {
			break
		}
	}
}

func (ctl *EventController) signalEnqueue(signal *pb.ProcessManageSignal) (string, common.RespCode, error) {
	select {
	case ctl.signalChan <- signal:
		hwlog.RunLog.Infof("signal enqueue, jobId=%s, uuid=%s, signalType=%s, strategy=%s, faults=%s",
			signal.JobId, signal.Uuid, signal.SignalType, signal.ChangeStrategy, common.Faults2String(signal.FaultRankIds))
		return "", common.OK, nil
	case <-ctl.controllerContext.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case <-time.After(time.Second):
		info := fmt.Sprintf("add signal time-out for jobId=%s, "+
			"program may running in chaos", signal.JobId)
		hwlog.RunLog.Error(info)
		return "", common.SignalQueueBusy, errors.New(info)
	}
}

func (ctl *EventController) trigger(event string) (string, common.RespCode, error) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	return ctl.state.Trigger(event)
}

func (ctl *EventController) handleFinish() (string, common.RespCode, error) {
	ctl.reset(false)
	return "", common.OK, nil
}

func (ctl *EventController) handleNotifyWaitFaultFlushing() (string, common.RespCode, error) {
	if ctl.jobInfo.PlatFormMode {
		strategies, err := WaitPlatFormStrategyReady(ctl.jobInfo.JobName, ctl.jobInfo.Namespace)
		if err != nil {
			return common.WaitPlatStrategyTimeoutEvent, common.WaitPlatStrategyTimeout, nil
		}
		ctl.platStrategies = strategies
	}
	cm, err := common.WriteResetInfoToCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace, nil,
		common.NotifyFaultFlushingOperation)
	if err != nil {
		hwlog.RunLog.Errorf("notify agent faultFlushing error, err=%v", err)
		return common.NotifyFinishEvent, common.OperateConfigMapError, nil
	}
	hwlog.RunLog.Infof("write configmap FaultFlushing success, %s", cm.Data[common.ResetInfoCMDataKey])
	return common.NotifyFinishEvent, common.OK, nil
}

func (ctl *EventController) handleFaultClear() (string, common.RespCode, error) {
	cm, err := common.WriteResetInfoToCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace, nil, common.ClearOperation)
	if err != nil {
		hwlog.RunLog.Errorf("clear reset configmap error, err=%v", err)
		return common.ClearConfigMapFaultFailEvent, common.OperateConfigMapError, nil
	}
	hwlog.RunLog.Infof("clear reset configmap success, %s", cm.Data[common.ResetInfoCMDataKey])
	return common.ClearConfigMapFaultSuccessEvent, common.OK, nil
}

func (ctl *EventController) handleNotifyStopTrain() (string, common.RespCode, error) {
	ctl.uuid = common.NewEventId(randomLen)
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     common.StopTrainSignalType,
		Actions:        stopTrainActions,
		ChangeStrategy: "",
	}
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) handleWaitReportStopComplete() (string, common.RespCode, error) {
	select {
	case <-ctl.controllerContext.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s",
			ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case _ = <-ctl.reportStopCompleteChan:
		hwlog.RunLog.Infof("jobId=%s, outside triger event receiveReport", ctl.jobInfo.JobId)
		return common.ReceiveReportEvent, common.OK, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report stop complete timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) handleWaitFlushFinish() (string, common.RespCode, error) {
	go func() {
		select {
		case <-time.After(time.Duration(faultFlushSeconds) * time.Second):
			ctl.addEvent(common.FaultFlushFinishedEvent)
		case <-ctl.controllerContext.Done():
		}
	}()
	return "", common.OK, nil
}

func (ctl *EventController) normalFaultAssociateSameNodeRank(worker job.PodWorker) ([]*pb.FaultRank, []string) {
	var faultRankIds []string
	for _, fault := range ctl.cacheNormalFault {
		faultRankIds = append(faultRankIds, fault.RankId)
	}
	allFaultRankIds := common.GetFaultRankIdsInSameNode(faultRankIds, worker.GetDeviceNumPerNode())
	removeSameRankIds := util.RemoveSliceDuplicateElement(allFaultRankIds)
	var res []*pb.FaultRank
	for _, rank := range removeSameRankIds {
		res = append(res, &pb.FaultRank{
			RankId:    rank,
			FaultType: "1",
		})
	}
	return res, removeSameRankIds
}

func (ctl *EventController) writeConfirmFaultAndWaitPlatResultFault(faults []*pb.FaultRank) ([]*pb.FaultRank, error) {
	allFaultRanks := common.RemoveSliceDuplicateFaults(faults)
	err := UpdateProcessConfirmFault(ctl.jobInfo.JobName, ctl.jobInfo.Namespace, allFaultRanks)
	if err != nil {
		hwlog.RunLog.Errorf("update process confirm fault err: %v, jobId=%s", err, ctl.jobInfo.JobId)
		return nil, fmt.Errorf("update process confirm fault err: %v, jobId=%s", err, ctl.jobInfo.JobId)
	}
	platFaultResult, err := WaitProcessResultFault(ctl.jobInfo.PgName, ctl.jobInfo.Namespace)
	if err != nil {
		hwlog.RunLog.Errorf("wait process result fault err: %v, jobId=%s", err, ctl.jobInfo.JobId)
		return nil, fmt.Errorf("wait process result fault err: %v, jobId=%s", err, ctl.jobInfo.JobId)
	}
	allFaultRanks = common.RemoveSliceDuplicateFaults(append(allFaultRanks, platFaultResult...))
	return allFaultRanks, nil
}

func isUceFault(faults []*pb.FaultRank) bool {
	for _, fault := range faults {
		if fault.FaultType == "1" {
			return false
		}
	}
	return true
}

func (ctl *EventController) takeUceFault2NormalFault() {
	n := len(ctl.latestRecoverResult)
	if n > 0 && ctl.latestRecoverResult[n-1].Strategy == common.ProcessRetryStrategyName {
		ctl.cacheNormalFault = append(ctl.cacheNormalFault, ctl.cacheUceFault...)
		ctl.cacheUceFault = ctl.cacheUceFault[:0]
	}
	if len(ctl.cacheNormalFault) > 0 {
		ctl.cacheNormalFault = append(ctl.cacheNormalFault, ctl.cacheUceFault...)
		ctl.cacheUceFault = ctl.cacheUceFault[:0]
	}
	if !ctl.supportRetryStrategy() {
		ctl.cacheNormalFault = append(ctl.cacheNormalFault, ctl.cacheUceFault...)
		ctl.cacheUceFault = ctl.cacheUceFault[:0]
	}
}

func (ctl *EventController) notifyFaultForUceFaultCase(worker job.PodWorker) (string, common.RespCode, error) {
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     common.GlobalFaultSignalType,
		Actions:        globalFaultActions,
		ChangeStrategy: "",
	}
	if ctl.jobInfo.PlatFormMode {
		allFaults, err := ctl.writeConfirmFaultAndWaitPlatResultFault(ctl.cacheUceFault)
		if err != nil {
			hwlog.RunLog.Errorf("write confirm fault or wait plat result fault timeout, err=%v", err)
			return common.WriteConfirmFaultOrWaitResultFaultTimeoutEvent,
				common.WriteConfirmFaultOrWaitPlatResultFault, nil
		}
		if !isUceFault(allFaults) {
			ctl.cacheUceFault = ctl.cacheUceFault[:0]
			allFaults, allFaultRanks := ctl.normalFaultAssociateSameNodeRank(worker)
			ctl.cacheNormalFault = allFaults
			// label fault pod
			err := common.LabelSoftWareFaultPod(ctl.jobInfo.JobId, allFaultRanks)
			if err != nil {
				hwlog.RunLog.Errorf("label pod fault err: %v, jobId=%s", err, ctl.jobInfo.JobId)
			}
			cm, err := common.WriteResetInfoToCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace,
				allFaultRanks, common.NotifyFaultListOperation)
			if err != nil {
				hwlog.RunLog.Errorf("notify agent faultList error, err=%v", err)
				return common.NotifyFailEvent, common.OperateConfigMapError, nil
			}
			signal.FaultRankIds = ctl.cacheNormalFault
			hwlog.RunLog.Infof("write configmap faultList success, %s", cm.Data[common.ResetInfoCMDataKey])
		} else {
			signal.FaultRankIds = ctl.cacheUceFault
		}
	} else {
		signal.FaultRankIds = ctl.cacheUceFault
	}

	signal.FaultRankIds = append(signal.FaultRankIds, ctl.cacheUceFault...)
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) notifyFaultForNormalFaultCase(worker job.PodWorker) (string, common.RespCode, error) {
	if ctl.jobInfo.PlatFormMode {
		allFaults, err := ctl.writeConfirmFaultAndWaitPlatResultFault(ctl.cacheNormalFault)
		if err != nil {
			hwlog.RunLog.Errorf("write confirm fault or wait plat result fault timeout, err=%v", err)
			return common.WriteConfirmFaultOrWaitResultFaultTimeoutEvent,
				common.WriteConfirmFaultOrWaitPlatResultFault, nil
		}
		ctl.cacheNormalFault = allFaults
	}
	allFaults, allFaultRanks := ctl.normalFaultAssociateSameNodeRank(worker)
	ctl.cacheNormalFault = allFaults
	// label fault pod
	err := common.LabelSoftWareFaultPod(ctl.jobInfo.JobId, allFaultRanks)
	if err != nil {
		hwlog.RunLog.Errorf("label pod fault err: %v, jobId=%s", err, ctl.jobInfo.JobId)
	}
	cm, err := common.WriteResetInfoToCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace,
		allFaultRanks, common.NotifyFaultListOperation)
	if err != nil {
		hwlog.RunLog.Errorf("notify agent faultList error, err=%v", err)
		return common.NotifyFailEvent, common.OperateConfigMapError, nil
	}
	hwlog.RunLog.Infof("write configmap faultList success, %s", cm.Data[common.ResetInfoCMDataKey])
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     common.GlobalFaultSignalType,
		Actions:        globalFaultActions,
		ChangeStrategy: "",
	}
	signal.FaultRankIds = append(signal.FaultRankIds, ctl.cacheNormalFault...)
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) handleNotifyGlobalFault() (string, common.RespCode, error) {
	ctl.takeUceFault2NormalFault()
	worker := kube.JobMgr.GetBsWorker(ctl.jobInfo.JobId)
	if worker == nil {
		return "", common.JobNotExist, fmt.Errorf("jobId=%s not exist", ctl.jobInfo.JobId)
	}
	// if len(ctl.cacheUceFault) still bigger than 0 after takeUceFault2NormalFault
	// that means job support retry strategy, and it's first time choose strategy case only have uce fault
	if len(ctl.cacheUceFault) > 0 {
		return ctl.notifyFaultForUceFaultCase(worker)
	}
	return ctl.notifyFaultForNormalFaultCase(worker)
}

func (ctl *EventController) firstChooseStrategy() string {
	hwlog.RunLog.Infof("first choose strategy, jobId=%s", ctl.jobInfo.JobId)
	if ctl.supportRetryStrategy() && len(ctl.cacheNormalFault) <= 0 {
		return common.ProcessRetryStrategyName
	}
	if ctl.supportRecoverStrategy() {
		return common.ProcessRecoverStrategyName
	}
	if ctl.supportDumpStrategy() {
		return common.ProcessDumpStrategyName
	}
	return common.ProcessExitStrategyName
}

func (ctl *EventController) chooseForRetryFail() string {
	if ctl.supportRecoverStrategy() {
		return common.ProcessRecoverStrategyName
	}
	if ctl.supportDumpStrategy() {
		return common.ProcessDumpStrategyName
	}
	return common.ProcessExitStrategyName
}

func (ctl *EventController) chooseForRecoverFail() string {
	if ctl.supportDumpStrategy() {
		return common.ProcessDumpStrategyName
	}
	return common.ProcessExitStrategyName
}

func (ctl *EventController) chooseStrategy() string {
	n := len(ctl.latestRecoverResult)
	if n == 0 {
		return ctl.firstChooseStrategy()
	}
	res := ctl.latestRecoverResult[n-1]
	if res.Strategy == common.ProcessRetryStrategyName {
		return ctl.chooseForRetryFail()
	} else if res.Strategy == common.ProcessRecoverStrategy {
		return ctl.chooseForRecoverFail()
	}
	return common.ProcessExitStrategyName
}

func (ctl *EventController) handleNotifyDecidedStrategy() (string, common.RespCode, error) {
	signal := &pb.ProcessManageSignal{
		Uuid:       ctl.uuid,
		JobId:      ctl.jobInfo.JobId,
		SignalType: common.ChangeStrategySignalType,
		Actions:    changeStrategyActions,
	}
	signal.ChangeStrategy = ctl.chooseStrategy()
	if ctl.jobInfo.PlatFormMode && signal.ChangeStrategy == common.ProcessRecoverStrategyName {
		err := WaitRankTableReady(ctl.jobInfo.PgName, ctl.jobInfo.Namespace)
		if err != nil {
			return common.WaitRankTableReadyTimeoutEvent, common.ServerInnerError, nil
		}
	}
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) handleCheckRecoverResult() (string, common.RespCode, error) {
	n := len(ctl.latestRecoverResult)
	if n == 0 {
		return "", common.ServerInnerError,
			fmt.Errorf("unexpected case, jobId=%s, have not decide strategy", ctl.jobInfo.JobId)
	}
	strategy := ctl.latestRecoverResult[n-1].Strategy
	code := ctl.latestRecoverResult[n-1].Status.Code
	recoverSuccess := ctl.latestRecoverResult[n-1].Status.Code == int32(common.OK)
	switch strategy {
	case common.ProcessRetryStrategyName, common.ProcessRecoverStrategyName:
		if recoverSuccess {
			return common.RecoverSuccessEvent, common.OK, nil
		}
		if strategy == common.ProcessRetryStrategyName && code == int32(common.CleanDeviceError) {
			return common.DeviceCleanFailEvent, common.CleanDeviceError, nil
		}
		return common.RecoverFailEvent, common.ClientError, nil
	case common.ProcessDumpStrategyName, common.ProcessExitStrategyName:
		return common.CheckResultFinishEvent, common.OK, nil
	default:
		return "", common.ServerInnerError, fmt.Errorf("unexpected case, strategy=%s "+
			"not support, jobId=%s", strategy, ctl.jobInfo.JobId)
	}
}

func (ctl *EventController) handleFaultRetry() (string, common.RespCode, error) {
	if _, err := common.ChangeProcessSchedulingMode(ctl.jobInfo.PgName, ctl.jobInfo.Namespace,
		common.ProcessReschedulingPause); err != nil {
		hwlog.RunLog.Errorf("failed to change the process rescheduling label pause %s of pg %s, "+
			"prepare notify agent kill master through grpc channel",
			common.ProcessReschedulingPause, ctl.jobInfo.PgName)
		return common.ChangeProcessSchedulingModePauseErrorEvent, common.OperatePodGroupError, nil
	}
	hwlog.RunLog.Infof("change process rescheduling label %s success,"+
		" pgName=%s, uuid=%s", common.ProcessReschedulingPause, ctl.jobInfo.PgName, ctl.uuid)

	scheduleSuccess := false
	for i := 1; i <= common.CheckPGRunningRetryTimes/2; i++ {
		time.Sleep(time.Second * common.SleepSecondBeforeCheckPGRunning)
		if kube.JobMgr.JobRunning(ctl.jobInfo.JobId) {
			scheduleSuccess = true
			break
		}
	}

	if !scheduleSuccess {
		hwlog.RunLog.Errorf("jobId=%s schedule timeout, "+
			"prepare notify agent kill master through grpc channel", ctl.jobInfo.JobId)
		return common.ScheduleTimeoutEvent, common.ScheduleTimeout, nil
	}

	if _, err := common.ChangeProcessSchedulingMode(
		ctl.jobInfo.PgName, ctl.jobInfo.Namespace, common.ProcessReschedulingEnable); err != nil {
		hwlog.RunLog.Errorf("failed to change the process rescheduling label on %s of pg %s, "+
			"prepare notify agent kill master through grpc channel",
			common.ProcessReschedulingEnable, ctl.jobInfo.PgName)
		return common.ChangeProcessSchedulingModeEnableErrorEvent, common.OperatePodGroupError, nil
	}
	hwlog.RunLog.Infof("change process rescheduling label %s success,"+
		" jobId=%s, uuid=%s", common.ProcessReschedulingEnable, ctl.jobInfo.JobId, ctl.uuid)
	return common.FinishEvent, common.OK, nil
}

func (ctl *EventController) handleKillJob() (string, common.RespCode, error) {
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     common.KillMasterSignalType,
		Actions:        nil,
		FaultRankIds:   nil,
		ChangeStrategy: "",
	}
	select {
	case ctl.signalChan <- signal:
		hwlog.RunLog.Infof("kill master signal enqueue, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.OK, nil
	case <-ctl.controllerContext.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case <-time.After(time.Minute):
		hwlog.RunLog.Warnf("add signal=%s time-out, program may running in chaos", ctl.jobInfo.JobId)
		return common.FinishEvent, common.OK, nil
	}
}

func (ctl *EventController) handleWaitReportRecoverStrategy() (string, common.RespCode, error) {
	go func() {
		ctl.listenScheduleResult()
	}()
	select {
	case req := <-ctl.reportRecoverStrategyChan:
		ctl.agentReportStrategies = req.Strategies
		return common.ReceiveReportEvent, common.OK, nil
	case <-ctl.controllerContext.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report recover strategy timeout, jobId=%s", ctl.jobInfo.JobId)
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) listenScheduleResult() {
	worker := kube.JobMgr.GetBsWorker(ctl.jobInfo.JobId)
	if worker == nil {
		hwlog.RunLog.Warnf("jobId=%s not exist", ctl.jobInfo.JobId)
		return
	}
	pgRunning := false
	for i := 1; i <= common.CheckPGRunningRetryTimes; i++ {
		time.Sleep(time.Second * common.SleepSecondBeforeCheckPGRunning)
		hwlog.RunLog.Infof("check pg running %d times", i)
		if worker.PGRunning() {
			pgRunning = true
			break
		}
	}
	select {
	case ctl.scheduleResultChan <- pgRunning:
		hwlog.RunLog.Infof("schedule result enqueue success, jobId=%s, value=%s",
			ctl.jobInfo.JobId, strconv.FormatBool(pgRunning))
	case <-ctl.controllerContext.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
	case <-time.After(time.Second):
		hwlog.RunLog.Errorf("schedule result enqueue timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
	}
}

func (ctl *EventController) handleDecideRetryStrategy() (string, common.RespCode, error) {
	ctl.latestStrategy = append(ctl.latestStrategy, common.ProcessRetryStrategyName)
	select {
	case req := <-ctl.reportStatusChan:
		hwlog.RunLog.Infof("cur state is wait retry strategy status, report strategy=%s result", req.Strategy)
		ctl.latestRecoverResult = append(ctl.latestRecoverResult, req)
		return common.ReceiveReportEvent, common.OK, nil
	case <-ctl.controllerContext.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report recover status timeout, jobId=%s", ctl.jobInfo.JobId)
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) handleDecideRecoverStrategy() (string, common.RespCode, error) {
	ctl.latestStrategy = append(ctl.latestStrategy, common.ProcessRecoverStrategy)
	timer := time.NewTimer(time.Duration(reportTimeoutMinutes) * time.Minute)
	defer timer.Stop()
	for {
		select {
		case scheduleSuccess := <-ctl.scheduleResultChan:
			ctl.haveGetScheduleResult = true
			ctl.scheduleResult = scheduleSuccess
			if !scheduleSuccess {
				return common.ScheduleTimeoutEvent, common.ScheduleTimeout, nil
			}
			_, err := common.RetryWriteResetCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace, nil, common.ClearOperation)
			if err != nil {
				hwlog.RunLog.Errorf("clear reset configMap error, err=%v, jobId=%s, uuid=%s", err, ctl.jobInfo.JobId, ctl.uuid)
				return common.ClearConfigMapFaultFailEvent, common.OperateConfigMapError, nil
			}
		case req := <-ctl.reportStatusChan:
			hwlog.RunLog.Infof("cur state is %s, strategy=%s, code=%d", ctl.state.GetState(), req.Strategy, req.Status.Code)
			ctl.latestRecoverResult = append(ctl.latestRecoverResult, req)
			return common.ReceiveReportEvent, common.OK, nil
		case <-ctl.controllerContext.Done():
			hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
			return "", common.ControllerEventCancel, nil
		case <-timer.C:
			hwlog.RunLog.Errorf("wait report recover status timeout, jobId=%s", ctl.jobInfo.JobId)
			return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
		}
	}
}

func (ctl *EventController) handleDecideDumpStrategy() (string, common.RespCode, error) {
	ctl.latestStrategy = append(ctl.latestStrategy, common.ProcessDumpStrategyName)
	select {
	case req := <-ctl.reportStatusChan:
		hwlog.RunLog.Infof("cur state is %s, strategy=%s, code=%d", ctl.state.GetState(), req.Strategy, req.Status.Code)
		ctl.latestRecoverResult = append(ctl.latestRecoverResult, req)
		return common.ReceiveReportEvent, common.OK, nil
	case <-ctl.controllerContext.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("%s timeout, jobId=%s", ctl.state.GetState(), ctl.jobInfo.JobId)
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) handleDecideExitStrategy() (string, common.RespCode, error) {
	ctl.latestStrategy = append(ctl.latestStrategy, common.ProcessExitStrategyName)
	return common.CheckResultFinishEvent, common.OK, nil
}

func (ctl *EventController) handleListenScheduleResult() (string, common.RespCode, error) {
	if ctl.haveGetScheduleResult {
		if ctl.scheduleResult {
			return common.ScheduleSuccessEvent, common.OK, nil
		}
		return common.ScheduleTimeoutEvent, common.ScheduleTimeout, fmt.Errorf("jobId=%s schedule timeout", ctl.jobInfo.JobId)
	}
	scheduleSuccess := false
	for i := 1; i <= common.CheckPGRunningRetryTimes; i++ {
		time.Sleep(time.Second * common.SleepSecondBeforeCheckPGRunning)
		if kube.JobMgr.JobRunning(ctl.jobInfo.JobId) {
			scheduleSuccess = true
			break
		}
	}
	if scheduleSuccess {
		return common.ScheduleSuccessEvent, common.OK, nil
	}
	return common.ScheduleTimeoutEvent, common.ScheduleTimeout, fmt.Errorf("jobId=%s schedule timeout", ctl.jobInfo.JobId)
}

func (ctl *EventController) handleRestartAllProcess() (string, common.RespCode, error) {
	_, err := common.RetryWriteResetCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace, nil, common.RestartAllProcessOperation)
	if err != nil {
		hwlog.RunLog.Errorf("clear reset configMap error, err=%v, jobId=%s, uuid=%s",
			err, ctl.jobInfo.JobId, ctl.uuid)
		return common.NotifyFailEvent, common.OperateConfigMapError, nil
	}
	return common.NotifySuccessEvent, common.OK, nil
}

func (ctl *EventController) handleWaitRestartAllProcess() (string, common.RespCode, error) {
	go func() {
		select {
		case <-time.After(time.Minute):
			ctl.addEvent(common.RestartProcessFinishEvent)
		case <-ctl.controllerContext.Done():
			hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s",
				ctl.jobInfo.JobId, ctl.uuid)
		}
	}()
	return "", common.OK, nil
}
