// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package recover a series of service function
// om is operation manager
package recover

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/interface/grpc/recover"
)

const (
	stressTestTimeout = 15 * 60 // seconds
)

func (ctl *EventController) handleNotifyPauseTrain() (string, common.RespCode, error) {
	ctl.uuid = common.NewEventId(randomLen)
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.StopTrainSignalType,
		Actions:        pauseTrainActions,
		ChangeStrategy: "",
	}
	if ctl.isStressTest() {
		signal.Timeout = stressTestTimeout
	}
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) handleWaitPauseTrainComplete() (string, common.RespCode, error) {
	ctx, reportChan := ctl.getCtxAndStopCompleteChan()
	if reportChan == nil {
		ctl.replyOMResponse("pause train failed, job service not ready")
		hwlog.RunLog.Infof("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
		return "", common.ServerInnerError, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case req := <-reportChan:
		if req.Status.Code == common.UnRecoverTrainError {
			ctl.replyOMResponse("om failed, pause train failed")
			return common.ProcessPauseFailEvent, common.ClientError, nil
		}
		if ctl.isSwitchingNic() {
			return common.SwitchNicRecvPauseEvent, common.OK, nil
		} else if ctl.isStressTest() {
			return common.StressTestRecvPauseEvent, common.OK, nil
		}
		return "", common.OK, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report pause train complete timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.replyOMResponse("pause train timeout")
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) notifySwitchNic() (string, common.RespCode, error) {
	globalRanks, globalOps := ctl.getSwitchNicParam()
	signal := &pb.SwitchRankList{
		RankID: globalRanks,
		Op:     globalOps,
		JobId:  ctl.jobInfo.JobId,
	}
	return ctl.switchNicSignalEnqueue(signal)
}

func (ctl *EventController) notifyContinueTrain() (string, common.RespCode, error) {
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.ChangeStrategySignalType,
		Actions:        changeStrategyActions,
		FaultRanks:     ctl.cacheNormalFault,
		ChangeStrategy: constant.ProcessContinueTrain,
	}
	var err error
	signal.NodeRankIds, err = common.GetNodeRankIdsByFaultRanks(ctl.jobInfo.JobId, signal.FaultRanks)
	if err != nil {
		hwlog.RunLog.Warnf("jobId=%s, GetNodeRankIdsByFaultRanks err:%v", ctl.jobInfo.JobId, err)
	}
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) handleSwitchNicFinish() (string, common.RespCode, error) {
	ctl.replyOMResponse("switch nic finish")
	ctl.reset(false)
	return "", common.OK, nil
}

func (ctl *EventController) handleWaitSwitchNicFinish() (string, common.RespCode, error) {
	hwlog.RunLog.Infof("jobId=%s, wait switch nic finish....", ctl.jobInfo.JobId)
	ctx, ch := ctl.getCtxAndResultChan()
	if ch == nil {
		hwlog.RunLog.Infof("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
		ctl.replyOMResponse("switch nic failed, job service not ready")
		return "", common.ServerInnerError, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
	}
	ctx, rch := ctl.getCtxAndSwitchNicResultChan()
	if rch == nil {
		hwlog.RunLog.Infof("jobId=%s, resultChan is nil", ctl.jobInfo.JobId)
		ctl.replyOMResponse("switch nic failed, job service not ready")
		return "", common.ServerInnerError, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case req := <-rch:
		if !req.Result {
			ctl.replyOMResponse("switch nic failed, switch nic failed")
			return common.SwitchNicFailEvent, common.ClientError, nil
		}
		return common.ReceiveReportEvent, common.OK, nil
	case req := <-ch:
		// mindio use report_status to report error when switching nic
		if req.Status.Code == common.UnRecoverableRetryError {
			ctl.replyOMResponse("switch nic failed, report error when switching nic")
			return common.WaitSwitchNicRecvFaultEvent, common.ClientError, nil
		}
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report switch nic complete timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.replyOMResponse("switch nic failed, report switch nic timeout")
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
	return "", common.OK, nil
}

func (ctl *EventController) handleDecideContinueTrainComplete() (string, common.RespCode, error) {
	ctx, ch := ctl.getCtxAndResultChan()
	if ch == nil {
		hwlog.RunLog.Infof("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
		ctl.replyOMResponse("continue train failed, job service not ready")
		return "", common.ServerInnerError, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case req := <-ch:
		if req.Status.Code == common.UnRecoverTrainError {
			ctl.replyOMResponse("switch nic failed, continue train failed")
			return common.ContinueTrainFailEvent, common.ClientError, nil
		}
		if ctl.isSwitchingNic() {
			return common.SwitchNicRecvContinueEvent, common.OK, nil
		} else if ctl.isStressTest() {
			return common.StressTestRecvContinueEvent, common.OK, nil
		}
		return "", common.OK, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report continue train complete timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.replyOMResponse("switch nic failed, continue train timeout")
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) listenSwitchNicChannel(stream pb.Recover_SubscribeSwitchNicSignalServer) {
	ctx, sendChan := ctl.getCtxAndSwitchNicResponseChan()
	hwlog.RunLog.Infof("start listen a new send channel, jobId=%s", ctl.jobInfo.JobId)
	ctl.selectSendSwitchNicResponseChan(ctx, sendChan, stream)
}

func (ctl *EventController) getCtxAndSwitchNicResponseChan() (context.Context, chan *pb.SwitchNicResponse) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.switchNicResponse
}

func (ctl *EventController) switchNicSignalEnqueue(signal *pb.SwitchRankList) (string, common.RespCode, error) {
	ctx, sendChan := ctl.getCtxAndSwitchNicNotifyChan()
	if sendChan == nil {
		hwlog.RunLog.Errorf("jobId=%s, sendChan is nil", ctl.jobInfo.JobId)
		return "", common.SignalQueueBusy, errors.New("sendChan is nil")
	}
	select {
	case sendChan <- signal:
		hwlog.RunLog.Infof("signal enqueue, jobId=%s, ranks=%v, ops=%v", signal.JobId, signal.RankID, signal.Op)
		return "", common.OK, nil
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case <-time.After(time.Second):
		info := fmt.Sprintf("add signal time-out for jobId=%s, program may running in chaos", signal.JobId)
		hwlog.RunLog.Error(info)
		return "", common.SignalQueueBusy, errors.New(info)
	}
}

func (ctl *EventController) getCtxAndSwitchNicNotifyChan() (context.Context, chan *pb.SwitchRankList) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.switchRankList
}

func (ctl *EventController) getCtxAndSwitchNicResultChan() (context.Context, chan *pb.SwitchResult) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.switchRankResult
}

func (ctl *EventController) listenSwitchNicNotifyChannel(stream pb.Recover_SubscribeNotifySwitchServer) {
	ctx, sendChan := ctl.getCtxAndSwitchNicNotifyChan()
	hwlog.RunLog.Infof("start listen a new switch nic send channel, jobId=%s", ctl.jobInfo.JobId)
	for {
		exit := ctl.selectNotifySwitchNic(ctx, sendChan, stream)
		if exit {
			break
		}
	}

}

func (ctl *EventController) selectNotifySwitchNic(ctx context.Context, sendChan chan *pb.SwitchRankList,
	stream pb.Recover_SubscribeNotifySwitchServer) bool {
	if sendChan == nil {
		return true
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Infof("context done, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
		return true
	case <-stream.Context().Done():
		hwlog.RunLog.Infof("stream context done, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
		return true
	case signal, ok := <-sendChan:
		if !ok {
			hwlog.RunLog.Infof("sendChan closed, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
			return true
		}
		err := common.SendWithRetry(stream, signal, retryTimes)
		if err != nil {
			hwlog.RunLog.Errorf("send switch nic signal failed, err=%v, jobId=%s", err, ctl.jobInfo.JobId)
			ctl.replyOMResponse("send switch nic signal failed, please check manually")
			ctl.addEvent(common.NotifyFailEvent)
			return false
		}
		hwlog.RunLog.Infof("switch nic signal=%v, jobId=%s", signal, ctl.jobInfo.JobId)
		ctl.addEvent(common.NotifySuccessEvent)
		return false
	}
}

func (ctl *EventController) selectSendSwitchNicResponseChan(ctx context.Context, sendChan chan *pb.SwitchNicResponse,
	stream pb.Recover_SubscribeSwitchNicSignalServer) {
	if sendChan == nil {
		return
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Infof("context done, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
		return
	case <-stream.Context().Done():
		hwlog.RunLog.Infof("stream context done, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
		return
	case signal, ok := <-sendChan:
		if !ok {
			hwlog.RunLog.Infof("sendChan closed, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
			return
		}
		hwlog.RunLog.Infof("switch nic signal=%v, jobId=%s", signal, ctl.jobInfo.JobId)
		err := common.SendWithRetry(stream, signal, retryTimes)
		if err != nil {
			hwlog.RunLog.Errorf("send switch nic signal failed, err=%v, jobId=%s", err, ctl.jobInfo.JobId)
		}
		return
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("report switch nic result timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.replyOMResponse("report switch nic result timeout, please check manually")
	}
}

func (ctl *EventController) setSwitchNicResult(result *pb.SwitchResult) {
	ctl.switchRankResult <- result
}

func (ctl *EventController) setSwitchNicParam(globalSwitchRankIDs []string, globalOps []bool) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	ctl.globalSwitchRankIDs = globalSwitchRankIDs
	ctl.globalOps = globalOps
}

func (ctl *EventController) getSwitchNicParam() ([]string, []bool) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.globalSwitchRankIDs, ctl.globalOps
}

func (ctl *EventController) isSwitchingNic() bool {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return len(ctl.globalSwitchRankIDs) > 0
}

func (ctl *EventController) canDoSwitchingNic() bool {
	return ctl.state.GetState() == common.InitState
}

func (ctl *EventController) replyOMResponse(msg string) {
	if ctl.isSwitchingNic() {
		ctl.switchNicResponse <- &pb.SwitchNicResponse{
			Msg:   msg,
			JobID: ctl.jobInfo.JobId,
		}
	} else if ctl.isStressTest() {
		ctl.stressTestResponse <- &pb.StressTestResponse{
			Msg:   msg,
			JobID: ctl.jobInfo.JobId,
		}
	}
}

func (ctl *EventController) setStressTestParam(param common.StressTestParam) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	ctl.stressTestParam = param
}

func (ctl *EventController) isStressTest() bool {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return len(ctl.stressTestParam) > 0
}

func (ctl *EventController) getStressTestParam() common.StressTestParam {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.stressTestParam
}

func (ctl *EventController) notifyStressTest() (string, common.RespCode, error) {
	return "", common.OK, nil
}

func (ctl *EventController) stressTestSignalEnqueue(signal *pb.StressTestRankParams) (string, common.RespCode, error) {
	return "", common.OK, nil
}

func (ctl *EventController) getCtxAndStressTestNotifyChan() (context.Context, chan *pb.StressTestRankParams) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.stressTestNotifyChan
}

func (ctl *EventController) listenStressTestNotifyChannel(stream pb.Recover_SubscribeNotifyExecStressTestServer) {
	ctx, sendChan := ctl.getCtxAndStressTestNotifyChan()
	hwlog.RunLog.Infof("start listen a new stress test send channel, jobId=%s", ctl.jobInfo.JobId)
	for {
		if ctl.selectNotifyStressTest(ctx, sendChan, stream) {
			break
		}
	}
}

func (ctl *EventController) selectNotifyStressTest(ctx context.Context, sendChan chan *pb.StressTestRankParams,
	stream pb.Recover_SubscribeNotifyExecStressTestServer) bool {
	if sendChan == nil {
		return true
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Infof("context done, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
		return true
	case <-stream.Context().Done():
		hwlog.RunLog.Infof("stream context done, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
		return true
	case signal, ok := <-sendChan:
		if !ok {
			hwlog.RunLog.Infof("sendChan closed, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
			return true
		}
		err := common.SendWithRetry(stream, signal, retryTimes)
		if err != nil {
			hwlog.RunLog.Errorf("send stress test signal failed, err=%v, jobId=%s", err, ctl.jobInfo.JobId)
			ctl.replyOMResponse("send stress test signal failed, please check manually")
			ctl.addEvent(common.NotifyFailEvent)
			return false
		}
		hwlog.RunLog.Infof("stress test signal=%v, jobId=%s", signal, ctl.jobInfo.JobId)
		ctl.addEvent(common.NotifySuccessEvent)
		return false
	}
}

func (ctl *EventController) waitStressTestFinishRecvFault(ctx context.Context,
	rch chan *pb.StressTestResult) (string, common.RespCode, error) {
	return "", common.OK, nil
}

func (ctl *EventController) handleWaitStressTestFinish() (string, common.RespCode, error) {
	return "", common.OK, nil
}

func (ctl *EventController) waitStressTestDone(ctx context.Context, rch chan *pb.StressTestResult,
	ch chan *pb.RecoverStatusRequest) (string, common.RespCode, error) {
	return "", common.OK, nil
}

func (ctl *EventController) parseStressTestResult(result *pb.StressTestResult) (bool, string) {
	return true, ""
}

func (ctl *EventController) setStressTestResult(result *pb.StressTestResult) {
	ctl.stressTestResult <- result
}

func (ctl *EventController) getCtxAndStressTestResultChan() (context.Context, chan *pb.StressTestResult) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.stressTestResult
}

func (ctl *EventController) getCtxAndStressTestResponseChan() (context.Context, chan *pb.StressTestResponse) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.stressTestResponse
}

func (ctl *EventController) handleStressTestFail() (string, common.RespCode, error) {
	return "", common.OK, nil
}

func (ctl *EventController) handleStressTestFinish() (string, common.RespCode, error) {
	ctl.reset(false)
	return "", common.OK, nil
}

func (ctl *EventController) listenStressTestChannel(stream pb.Recover_SubscribeStressTestResponseServer) {
	ctx, sendChan := ctl.getCtxAndStressTestResponseChan()
	hwlog.RunLog.Infof("start listen a new send channel, jobId=%s", ctl.jobInfo.JobId)
	ctl.selectSendStressTestResponseChan(ctx, sendChan, stream)
}

func (ctl *EventController) selectSendStressTestResponseChan(ctx context.Context, sendChan chan *pb.StressTestResponse,
	stream pb.Recover_SubscribeStressTestResponseServer) {
	if sendChan == nil {
		return
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Infof("context done, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
	case <-stream.Context().Done():
		hwlog.RunLog.Infof("stream context done, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
	case signal, ok := <-sendChan:
		if !ok {
			hwlog.RunLog.Infof("sendChan closed, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
			return
		}
		hwlog.RunLog.Infof("stress test signal=%v, jobId=%s", signal, ctl.jobInfo.JobId)
		err := common.SendWithRetry(stream, signal, retryTimes)
		if err != nil {
			hwlog.RunLog.Errorf("send stress test signal failed, err=%v, jobId=%s", err, ctl.jobInfo.JobId)
		}
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("report stress test result timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.replyOMResponse("report stress test result timeout, please check manually")
	}
}
