// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package service a series of service function
package service

import "clusterd/pkg/interface/grpc/common"

func (ctl *EventController) getPreRules() []common.TransRule {
	return []common.TransRule{
		{Src: common.InitState, Event: common.FaultOccurEvent,
			Dst: common.NotifyWaitFaultFlushingState, Handler: ctl.handleNotifyWaitFaultFlushing},

		{Src: common.NotifyWaitFaultFlushingState, Event: common.NotifyFinishEvent,
			Dst: common.NotifyStopTrainState, Handler: ctl.handleNotifyStopTrain},
		{Src: common.NotifyWaitFaultFlushingState, Event: common.WaitPlatStrategyTimeoutEvent,
			Dst: common.FaultRetryState, Handler: ctl.handleFaultRetry},

		{Src: common.NotifyStopTrainState, Event: common.NotifySuccessEvent,
			Dst: common.WaitReportStopCompleteState, Handler: ctl.handleWaitReportStopComplete},
		{Src: common.NotifyStopTrainState, Event: common.NotifyFailEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},

		{Src: common.WaitReportStopCompleteState, Event: common.ReceiveReportEvent,
			Dst: common.WaitFaultFlushFinishedState, Handler: ctl.handleWaitFlushFinish},
		{Src: common.WaitReportStopCompleteState, Event: common.ReportTimeoutEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},

		{Src: common.WaitFaultFlushFinishedState, Event: common.FaultFlushFinishedEvent,
			Dst: common.NotifyGlobalFaultState, Handler: ctl.handleNotifyGlobalFault},

		{Src: common.NotifyGlobalFaultState, Event: common.NotifySuccessEvent,
			Dst: common.WaitReportRecoverStrategyState, Handler: ctl.handleWaitReportRecoverStrategy},
		{Src: common.NotifyGlobalFaultState, Event: common.NotifyFailEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},
		{Src: common.NotifyGlobalFaultState, Event: common.WriteConfirmFaultOrWaitResultFaultTimeoutEvent,
			Dst: common.FaultRetryState, Handler: ctl.handleFaultRetry},

		{Src: common.WaitReportRecoverStrategyState, Event: common.ReceiveReportEvent,
			Dst: common.NotifyDecidedStrategyState, Handler: ctl.handleNotifyDecidedStrategy},
		{Src: common.WaitReportRecoverStrategyState, Event: common.ReportTimeoutEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},

		{Src: common.NotifyDecidedStrategyState, Event: common.WaitRankTableReadyTimeoutEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},
		{Src: common.NotifyDecidedStrategyState, Event: common.NotifyFailEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},
		{Src: common.NotifyDecidedStrategyState, Event: common.NotifyRetrySuccessEvent,
			Dst: common.WaitReportStepRetryStatusState, Handler: ctl.handleDecideRetryStrategy},
		{Src: common.NotifyDecidedStrategyState, Event: common.NotifyRecoverSuccessEvent,
			Dst: common.WaitReportProcessRecoverStatusState, Handler: ctl.handleDecideRecoverStrategy},
		{Src: common.NotifyDecidedStrategyState, Event: common.NotifyDumpSuccessEvent,
			Dst: common.WaitReportDumpStatusState, Handler: ctl.handleDecideDumpStrategy},
		{Src: common.NotifyDecidedStrategyState, Event: common.NotifyExitSuccessEvent,
			Dst: common.CheckRecoverResultState, Handler: ctl.handleDecideExitStrategy},
	}
}

func (ctl *EventController) getFixRules() []common.TransRule {
	return []common.TransRule{
		{Src: common.WaitReportStepRetryStatusState, Event: common.ReceiveReportEvent,
			Dst: common.CheckRecoverResultState, Handler: ctl.handleCheckRecoverResult},
		{Src: common.WaitReportStepRetryStatusState, Event: common.ReportTimeoutEvent,
			Dst: common.FaultRetryState, Handler: ctl.handleFaultRetry},

		{Src: common.WaitReportProcessRecoverStatusState, Event: common.ReceiveReportEvent,
			Dst: common.CheckRecoverResultState, Handler: ctl.handleCheckRecoverResult},
		{Src: common.WaitReportProcessRecoverStatusState, Event: common.ScheduleTimeoutEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},
		{Src: common.WaitReportProcessRecoverStatusState, Event: common.ReportTimeoutEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},
		{Src: common.WaitReportProcessRecoverStatusState, Event: common.ClearConfigMapFaultFailEvent,
			Dst: common.FaultRetryState, Handler: ctl.handleFaultClear},

		{Src: common.WaitReportDumpStatusState, Event: common.ReceiveReportEvent,
			Dst: common.CheckRecoverResultState, Handler: ctl.handleCheckRecoverResult},
		{Src: common.WaitReportDumpStatusState, Event: common.ReportTimeoutEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},

		{Src: common.CheckRecoverResultState, Event: common.RecoverSuccessEvent,
			Dst: common.InitState, Handler: ctl.handleFinish},
		{Src: common.CheckRecoverResultState, Event: common.RecoverFailEvent,
			Dst: common.NotifyDecidedStrategyState, Handler: ctl.handleNotifyDecidedStrategy},
		{Src: common.CheckRecoverResultState, Event: common.DeviceCleanFailEvent,
			Dst: common.WaitFaultFlushFinishedState, Handler: ctl.handleWaitFlushFinish},
		{Src: common.CheckRecoverResultState, Event: common.CheckResultFinishEvent,
			Dst: common.ListenScheduleResultState, Handler: ctl.handleListenScheduleResult},
	}
}

func (ctl *EventController) getAfterRules() []common.TransRule {
	return []common.TransRule{
		{Src: common.ListenScheduleResultState, Event: common.ScheduleTimeoutEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},
		{Src: common.ListenScheduleResultState, Event: common.ScheduleSuccessEvent,
			Dst: common.NotifyRestartAllProcessState, Handler: ctl.handleRestartAllProcess},

		{Src: common.NotifyRestartAllProcessState, Event: common.NotifySuccessEvent,
			Dst: common.WaitRestartAllProcessState, Handler: ctl.handleWaitRestartAllProcess},
		{Src: common.NotifyRestartAllProcessState, Event: common.NotifyFailEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},

		{Src: common.WaitRestartAllProcessState, Event: common.RestartProcessFinishEvent,
			Dst: common.InitState, Handler: ctl.handleFinish},

		{Src: common.FaultClearState, Event: common.ClearConfigMapFaultSuccessEvent,
			Dst: common.FaultRetryState, Handler: ctl.handleFaultRetry},
		{Src: common.FaultClearState, Event: common.ClearConfigMapFaultFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},

		{Src: common.FaultRetryState, Event: common.FinishEvent,
			Dst: common.InitState, Handler: ctl.handleFinish},
		{Src: common.FaultRetryState, Event: common.ChangeProcessSchedulingModePauseErrorEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},
		{Src: common.FaultRetryState, Event: common.ChangeProcessSchedulingModeEnableErrorEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},
		{Src: common.FaultRetryState, Event: common.ScheduleTimeoutEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},

		{Src: common.NotifyKillJobState, Event: common.FinishEvent,
			Dst: common.InitState, Handler: ctl.handleFinish},
	}
}

func (ctl *EventController) getBaseRules() []common.TransRule {
	var rules []common.TransRule
	rules = append(rules, ctl.getPreRules()...)
	rules = append(rules, ctl.getFixRules()...)
	rules = append(rules, ctl.getAfterRules()...)
	return rules
}
