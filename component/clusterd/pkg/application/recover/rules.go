// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package recover a series of service function
package recover

import (
	"clusterd/pkg/domain/common"
)

func (ctl *EventController) getExtendPreRules() []common.TransRule {
	return []common.TransRule{
		{Src: common.NotifyDecidedStrategyState, Event: common.WaitHCCLRoutingConvergenceFail,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},
		{Src: common.NotifyDecidedStrategyState, Event: common.KillPodAfterRestartProcessEvent,
			Dst: common.KillPodForChooseStrategyAgainState, Handler: ctl.handleKillPod},
		{Src: common.KillPodForChooseStrategyAgainState, Event: common.FinishKillPodEvent,
			Dst: common.NotifyDecidedStrategyState, Handler: ctl.handleNotifyDecidedStrategy},
	}
}

func (ctl *EventController) geOMRules() []common.TransRule {
	return []common.TransRule{
		{Src: common.InitState, Event: common.StartSwitchNic,
			Dst: common.NotifyPauseTrainState, Handler: ctl.handleNotifyPauseTrain},

		{Src: common.NotifyPauseTrainState, Event: common.NotifySuccessEvent,
			Dst: common.WaitReportPauseCompleteState, Handler: ctl.handleWaitPauseTrainComplete},
		{Src: common.NotifyPauseTrainState, Event: common.NotifyFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},

		{Src: common.WaitReportPauseCompleteState, Event: common.ReceiveReportEvent,
			Dst: common.NotifySwitchNicState, Handler: ctl.notifySwitchNic},
		{Src: common.WaitReportPauseCompleteState, Event: common.ReportTimeoutEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},
		{Src: common.WaitReportPauseCompleteState, Event: common.ProcessPauseFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},

		{Src: common.NotifySwitchNicState, Event: common.NotifySuccessEvent,
			Dst: common.WaitSwitchNicFinishedState, Handler: ctl.handleWaitSwitchNicFinish},
		{Src: common.NotifySwitchNicState, Event: common.NotifyFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},

		{Src: common.WaitSwitchNicFinishedState, Event: common.ReceiveReportEvent,
			Dst: common.NotifyContinueTrainState, Handler: ctl.notifyContinueTrain},
		{Src: common.WaitSwitchNicFinishedState, Event: common.WaitSwitchNicRecvFaultEvent,
			Dst: common.NotifyStopTrainState, Handler: ctl.handleNotifyStopTrain},
		{Src: common.WaitSwitchNicFinishedState, Event: common.ReportTimeoutEvent,
			Dst: common.NotifyDumpState, Handler: ctl.handleNotifyDump},
		{Src: common.WaitSwitchNicFinishedState, Event: common.SwitchNicFailEvent,
			Dst: common.NotifyDumpState, Handler: ctl.handleNotifyDump},

		{Src: common.CheckRecoverResultState, Event: common.SwitchNicFailRecoverEvent,
			Dst: common.NotifyStopTrainState, Handler: ctl.handleNotifyStopTrain},

		{Src: common.NotifyContinueTrainState, Event: common.NotifyContinueSuccessEvent,
			Dst: common.WaitContinueTrainState, Handler: ctl.handleDecideContinueTrainComplete},
		{Src: common.NotifyContinueTrainState, Event: common.NotifyFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},

		{Src: common.WaitContinueTrainState, Event: common.ReceiveReportEvent,
			Dst: common.InitState, Handler: ctl.handleSwitchNicFinish},
		{Src: common.WaitContinueTrainState, Event: common.ReportTimeoutEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},
		{Src: common.WaitContinueTrainState, Event: common.ContinueTrainFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},
	}
}

func (ctl *EventController) getPreRules() []common.TransRule {
	return []common.TransRule{
		{Src: common.InitState, Event: common.FaultOccurEvent,
			Dst: common.NotifyWaitFaultFlushingState, Handler: ctl.handleNotifyWaitFaultFlushing},

		{Src: common.NotifyWaitFaultFlushingState, Event: common.NotifyFinishEvent,
			Dst: common.NotifyStopTrainState, Handler: ctl.handleNotifyStopTrain},
		{Src: common.NotifyWaitFaultFlushingState, Event: common.WaitPlatStrategyTimeoutEvent,
			Dst: common.FaultRetryState, Handler: ctl.handleFaultRetry},
		{Src: common.NotifyWaitFaultFlushingState, Event: common.DumpForFaultEvent,
			Dst: common.NotifyDumpState, Handler: ctl.handleNotifyDump},

		{Src: common.NotifyStopTrainState, Event: common.NotifySuccessEvent,
			Dst: common.WaitReportStopCompleteState, Handler: ctl.handleWaitReportStopComplete},
		{Src: common.NotifyStopTrainState, Event: common.NotifyFailEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},

		{Src: common.NotifyDumpState, Event: common.NotifySuccessEvent,
			Dst: common.WaitReportDumpStatusState, Handler: ctl.handleDecideDumpStrategy},
		{Src: common.NotifyDumpState, Event: common.NotifyFailEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},

		{Src: common.WaitReportStopCompleteState, Event: common.ReceiveReportEvent,
			Dst: common.WaitFaultFlushFinishedState, Handler: ctl.handleWaitFlushFinish},
		{Src: common.WaitReportStopCompleteState, Event: common.ReportTimeoutEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},
		{Src: common.WaitReportStopCompleteState, Event: common.ProcessNotReadyEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},

		{Src: common.WaitFaultFlushFinishedState, Event: common.FaultFlushFinishedEvent,
			Dst: common.NotifyGlobalFaultState, Handler: ctl.handleNotifyGlobalFault},

		{Src: common.NotifyGlobalFaultState, Event: common.NotifySuccessEvent,
			Dst: common.WaitReportRecoverStrategyState, Handler: ctl.handleWaitReportRecoverStrategy},
		{Src: common.NotifyGlobalFaultState, Event: common.NotifyFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},
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
		{Src: common.CheckRecoverResultState, Event: common.RecoverableRetryErrorEvent,
			Dst: common.WaitFaultFlushFinishedState, Handler: ctl.handleWaitFlushFinish},
		{Src: common.CheckRecoverResultState, Event: common.UnRecoverableRetryErrorEvent,
			Dst: common.KillPodForUnrecoverableRetryState, Handler: ctl.handleKillPod},
		{Src: common.CheckRecoverResultState, Event: common.CheckResultFinishEvent,
			Dst: common.ListenScheduleResultState, Handler: ctl.handleListenScheduleResult},

		{Src: common.KillPodForUnrecoverableRetryState, Event: common.FinishKillPodEvent,
			Dst: common.NotifyDecidedStrategyState, Handler: ctl.handleNotifyDecidedStrategy},
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
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},

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
	rules = append(rules, ctl.getExtendPreRules()...)
	rules = append(rules, ctl.getFixRules()...)
	rules = append(rules, ctl.getAfterRules()...)
	rules = append(rules, ctl.geOMRules()...)
	return rules
}
