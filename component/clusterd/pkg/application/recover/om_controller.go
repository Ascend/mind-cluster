// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package recover a series of service function
package recover

import (
	"context"
	"fmt"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/interface/grpc/recover"
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
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) handleWaitPauseTrainComplete() (string, common.RespCode, error) {
	ctx, reportChan := ctl.getCtxAndStopCompleteChan()
	if reportChan == nil {
		ctl.switchNicResponse <- &pb.SwitchNicResponse{
			Msg:   "switch nic failed, job service not ready",
			JobID: ctl.jobInfo.JobId,
		}
		hwlog.RunLog.Infof("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
		return "", common.ServerInnerError, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s",
			ctl.jobInfo.JobId, ctl.uuid)
		ctl.switchNicResponse <- &pb.SwitchNicResponse{
			Msg:   "switch nic failed, job service not ready",
			JobID: ctl.jobInfo.JobId,
		}
		return "", common.ControllerEventCancel, nil
	case req := <-reportChan:
		if req.Status.Code == common.UnRecoverTrainError {
			ctl.switchNicResponse <- &pb.SwitchNicResponse{
				Msg:   "switch nic failed, pause train failed",
				JobID: ctl.jobInfo.JobId,
			}
			return common.ProcessPauseFailEvent, common.ClientError, nil
		}
		return common.ReceiveReportEvent, common.OK, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report pause train complete timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.switchNicResponse <- &pb.SwitchNicResponse{
			Msg:   "switch nic failed, pause train timeout",
			JobID: ctl.jobInfo.JobId,
		}
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) notifySwitchNic() (string, common.RespCode, error) {
	return common.NotifySuccessEvent, common.OK, nil
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
	ctl.switchNicResponse <- &pb.SwitchNicResponse{
		Msg:   "switch nic finish",
		JobID: ctl.jobInfo.JobId,
	}
	ctl.reset(false)
	return "", common.OK, nil
}

func (ctl *EventController) handleWaitSwitchNicFinish() (string, common.RespCode, error) {
	ctx, ch := ctl.getCtxAndResultChan()
	if ch == nil {
		hwlog.RunLog.Infof("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
		ctl.switchNicResponse <- &pb.SwitchNicResponse{
			Msg:   "switch nic failed, job service not ready",
			JobID: ctl.jobInfo.JobId,
		}
		return "", common.ServerInnerError, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.switchNicResponse <- &pb.SwitchNicResponse{
			Msg:   "switch nic failed, job service not ready",
			JobID: ctl.jobInfo.JobId,
		}
		return "", common.ControllerEventCancel, nil
	case req := <-ch:
		// todo: taskD report switch nic failed
		if req.Status.Code == common.SwitchNicFail {
			ctl.switchNicResponse <- &pb.SwitchNicResponse{
				Msg:   "switch nic failed, switch nic failed",
				JobID: ctl.jobInfo.JobId,
			}
			return common.SwitchNicFailEvent, common.ClientError, nil
		}
		// mindio use report_status to report error when switching nic
		if req.Status.Code == common.UnRecoverableRetryError {
			ctl.switchNicResponse <- &pb.SwitchNicResponse{
				Msg:   "switch nic failed, report error when switching nic",
				JobID: ctl.jobInfo.JobId,
			}
			return common.WaitSwitchNicRecvFaultEvent, common.ClientError, nil
		}
		return common.ReceiveReportEvent, common.OK, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report switch nic complete timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.switchNicResponse <- &pb.SwitchNicResponse{
			Msg:   "switch nic failed, report switch nic timeout",
			JobID: ctl.jobInfo.JobId,
		}
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) handleDecideContinueTrainComplete() (string, common.RespCode, error) {
	ctx, ch := ctl.getCtxAndResultChan()
	if ch == nil {
		hwlog.RunLog.Infof("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
		ctl.switchNicResponse <- &pb.SwitchNicResponse{
			Msg:   "switch nic failed, job service not ready",
			JobID: ctl.jobInfo.JobId,
		}
		return "", common.ServerInnerError, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.switchNicResponse <- &pb.SwitchNicResponse{
			Msg:   "switch nic failed, job service not ready",
			JobID: ctl.jobInfo.JobId,
		}
		return "", common.ControllerEventCancel, nil
	case req := <-ch:
		if req.Status.Code == common.UnRecoverTrainError {
			ctl.switchNicResponse <- &pb.SwitchNicResponse{
				Msg:   "switch nic failed, continue train failed",
				JobID: ctl.jobInfo.JobId,
			}
			return common.ContinueTrainFailEvent, common.ClientError, nil
		}
		return common.ReceiveReportEvent, common.OK, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report continue train complete timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.switchNicResponse <- &pb.SwitchNicResponse{
			Msg:   "switch nic failed, continue train timeout",
			JobID: ctl.jobInfo.JobId,
		}
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) listenSwitchNicChannel(stream pb.Recover_SubscribeSwitchNicSignalServer) {
	ctx, sendChan := ctl.getCtxAndSwitchNicChan()
	hwlog.RunLog.Infof("start listen a new send channel, jobId=%s", ctl.jobInfo.JobId)
	ctl.selectSendSwitchNicChan(ctx, sendChan, stream)

}

func (ctl *EventController) getCtxAndSwitchNicChan() (context.Context, chan *pb.SwitchNicResponse) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.switchNicResponse
}

func (ctl *EventController) selectSendSwitchNicChan(ctx context.Context, sendChan chan *pb.SwitchNicResponse,
	stream pb.Recover_SubscribeSwitchNicSignalServer) {
	if sendChan == nil {
		return
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Infof("context done, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
		return
	case signal, ok := <-sendChan:
		if !ok {
			hwlog.RunLog.Infof("sendChan closed, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
			return
		}
		hwlog.RunLog.Infof("switch nic signal=%v, jobId=%s", signal, ctl.jobInfo.JobId)
		err := common.SwitchNicSendRetry(stream, signal, retryTimes)
		if err != nil {
			hwlog.RunLog.Errorf("send switch nic signal failed, err=%v, jobId=%s", err, ctl.jobInfo.JobId)
		}
		return
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("report switch nic result timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		ctl.switchNicResponse <- &pb.SwitchNicResponse{
			Msg:   "report switch nic result timeout, please check manually",
			JobID: ctl.jobInfo.JobId,
		}
		return
	}
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
