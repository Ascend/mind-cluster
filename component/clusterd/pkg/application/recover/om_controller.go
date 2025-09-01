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
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/recover"
	"clusterd/pkg/interface/kube"
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
		ChangeStrategy: constant.ProcessContinueTrain,
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
	stressTestParam := ctl.getStressTestParam()
	rankOps := make(map[string]*pb.StressOpList)
	for _, p := range stressTestParam {
		for rankID, ops := range p {
			rankOps[rankID] = &pb.StressOpList{
				Ops: ops,
			}
		}
	}
	signal := &pb.StressTestRankParams{
		StressParam: rankOps,
		JobId:       ctl.jobInfo.JobId,
	}
	return ctl.stressTestSignalEnqueue(signal)
}

func (ctl *EventController) stressTestSignalEnqueue(signal *pb.StressTestRankParams) (string, common.RespCode, error) {
	ctx, sendChan := ctl.getCtxAndStressTestNotifyChan()
	if sendChan == nil {
		hwlog.RunLog.Errorf("jobId=%s, sendChan is nil", ctl.jobInfo.JobId)
		return "", common.SignalQueueBusy, errors.New("sendChan is nil")
	}
	select {
	case sendChan <- signal:
		hwlog.RunLog.Infof("signal enqueue, jobId=%s, params=%v", signal.JobId, signal.StressParam)
		return "", common.OK, nil
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case <-time.After(time.Second):
		info := fmt.Sprintf("add signal time-out for jobId=%s, program may running in chaos", signal.JobId)
		hwlog.RunLog.Errorf("signal: %v enqueue time-out, %s", signal, info)
		return "", common.SignalQueueBusy, errors.New(info)
	}
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
	hwlog.RunLog.Warnf("recv fault, when stressing test, jobId=%s", ctl.jobInfo.JobId)
	ctl.replyOMResponse("recv fault, when stressing test")
	select {
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case req := <-rch:
		_, msg := ctl.parseStressTestResult(req)
		ctl.replyOMResponse(msg)
		hwlog.RunLog.Warnf("stress test failed, start recover..., jobId=%s", ctl.jobInfo.JobId)
		return common.StressTestFailEvent, common.ClientError, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report stress test timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.replyOMResponse("stress test failed, report stress test timeout")
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) handleWaitStressTestFinish() (string, common.RespCode, error) {
	hwlog.RunLog.Infof("jobId=%s, wait stress test finish....", ctl.jobInfo.JobId)
	nodes := make([]string, 0)
	for node, _ := range ctl.getStressTestParam() {
		nodes = append(nodes, node)
	}
	faultmanager.FilterStressTestFault(ctl.jobInfo.JobId, nodes, true)
	defer faultmanager.FilterStressTestFault(ctl.jobInfo.JobId, nodes, false)
	cm, err := common.RetryWriteResetCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace, nil, false,
		constant.NotifyFaultFlushingOperation)
	if err != nil {
		hwlog.RunLog.Errorf("notify agent faultFlushing error, err=%v", err)
	} else {
		hwlog.RunLog.Infof("write configmap FaultFlushing success, %s", cm.Data[constant.ResetInfoCMDataKey])
	}
	ctx, ch := ctl.getCtxAndResultChan()
	if ch == nil {
		hwlog.RunLog.Infof("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
		ctl.replyOMResponse("stress test failed, job service not ready")
		return "", common.ServerInnerError, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
	}
	ctx, rch := ctl.getCtxAndStressTestResultChan()
	if rch == nil {
		hwlog.RunLog.Infof("jobId=%s, resultChan is nil", ctl.jobInfo.JobId)
		ctl.replyOMResponse("stress test failed, job service not ready")
		return "", common.ServerInnerError, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
	}
	return ctl.waitStressTestDone(ctx, rch, ch)
}

func (ctl *EventController) waitStressTestDone(ctx context.Context, rch chan *pb.StressTestResult,
	ch chan *pb.RecoverStatusRequest) (string, common.RespCode, error) {
	select {
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.replyOMResponse("stress test failed, job service not ready")
		return "", common.ControllerEventCancel, nil
	case req := <-rch:
		if _, err := common.RetryWriteResetCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace, nil,
			false, constant.ClearOperation); err != nil {
			hwlog.RunLog.Errorf("notify agent faultFlushing error, err=%v", err)
		}
		ok, msg := ctl.parseStressTestResult(req)
		if !ok {
			ctl.replyOMResponse(msg)
			return common.StressTestFailEvent, common.ClientError, nil
		}
		ctl.replyOMResponse(msg)
		return common.ReceiveReportEvent, common.OK, nil
	case req := <-ch:
		if req.Status.Code == common.UnRecoverableRetryError {
			return ctl.waitStressTestFinishRecvFault(ctx, rch)
		}
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report stress test timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.replyOMResponse("stress test failed, report stress test timeout")
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
	return "", common.OK, nil
}

func (ctl *EventController) parseStressTestResult(result *pb.StressTestResult) (bool, string) {
	hwlog.RunLog.Infof("jobId=%s, StressTestResult is %v", result.JobId, result.StressResult)
	jobInfo, ok := job.GetJobCache(result.JobId)
	if !ok {
		hwlog.RunLog.Errorf("get job cache failed, jobId=%s", result.JobId)
		return false, fmt.Sprintf("get job cache failed, jobId=%s", result.JobId)
	}
	rankNodeMap := make(map[string]string) // rank -> node
	rankDevMap := make(map[string]string)  // rank -> dev
	for _, server := range jobInfo.JobRankTable.ServerList {
		for _, dev := range server.DeviceList {
			rankNodeMap[dev.RankID] = server.ServerName
			rankDevMap[dev.RankID] = dev.DeviceID
		}
	}
	hwlog.RunLog.Infof("jobId=%s, , rankNodeMap is %v", result.JobId, rankNodeMap)
	nodeRankResultMap := make(map[string]map[string]*pb.StressTestRankResult) // node -> rank -> msg
	faultRank := make([]*pb.FaultRank, 0)
	for rankID, opResult := range result.StressResult {
		nodeName := rankNodeMap[rankID]
		if _, ok := nodeRankResultMap[nodeName]; !ok {
			nodeRankResultMap[nodeName] = make(map[string]*pb.StressTestRankResult)
		}
		devID := rankDevMap[rankID]
		nodeRankResultMap[nodeName][devID] = opResult
		for _, res := range opResult.RankResult {
			if res.Code == constant.StressTestFindFault {
				ctl.isolateNodes.Insert(nodeName)
			}
			if res.Code == constant.StressTestTimeout || res.Code == constant.StressTestVolRecoverFail {
				faultRank = append(faultRank, &pb.FaultRank{RankId: rankID, FaultType: constant.NormalFaultType})
			}
		}
	}
	ctl.saveCacheFault(faultRank)
	retStr := util.ObjToString(nodeRankResultMap)
	hwlog.RunLog.Infof("jobId=%s, isolateNode:%v result:%v", result.JobId, ctl.isolateNodes, retStr)
	if len(ctl.isolateNodes) > 0 {
		return false, fmt.Sprintf("stress test find fault, isolate node:%v,result:%v", ctl.isolateNodes, retStr)
	}
	if len(faultRank) > 0 {
		return false, fmt.Sprintf("stress test timeout fault, faultRank:%v,result:%v", faultRank, retStr)
	}
	return true, fmt.Sprintf("stress test finish, result:%v", retStr)
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
	jobInfo, ok := job.GetJobCache(ctl.jobInfo.JobId)
	if !ok {
		hwlog.RunLog.Errorf("get job cache failed, jobId=%s", ctl.jobInfo.JobId)
		return "", common.ServerInnerError, fmt.Errorf("get job cache failed, jobId=%s", ctl.jobInfo.JobId)
	}
	nodeRankMap := make(map[string][]string)
	for _, server := range jobInfo.JobRankTable.ServerList {
		ranks := make([]string, 0)
		for _, dev := range server.DeviceList {
			ranks = append(ranks, dev.RankID)
		}
		nodeRankMap[server.ServerName] = ranks
	}
	faultRank := make([]*pb.FaultRank, 0)
	for node := range ctl.isolateNodes {
		for _, rank := range nodeRankMap[node] {
			faultRank = append(faultRank, &pb.FaultRank{RankId: rank, FaultType: constant.NormalFaultType})
		}
		labels := map[string]string{constant.NodeHealthyStatusKey: constant.NodeUnHealthy}
		err := kube.RetryPatchNodeAnnotation(node, constant.PatchNodeTimes, labels)
		if err != nil {
			hwlog.RunLog.Errorf("patch node:%s failed: %v", node, err)
		}
	}
	ctl.saveCacheFault(faultRank)
	ctl.uuid = common.NewEventId(randomLen)
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.StopTrainSignalType,
		Actions:        stopTrainActions,
		ChangeStrategy: "",
	}
	signal.FaultRanks = append(ctl.cacheRetryFault, ctl.cacheNormalFault...)
	return ctl.signalEnqueue(signal)
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
