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

func (ctl *EventController) getStressTestRules() []common.TransRule {
	return []common.TransRule{
		{Src: common.InitState, Event: common.StartStressTest,
			Dst: common.NotifyPauseTrainState, Handler: ctl.handleNotifyPauseTrain},
		{Src: common.WaitReportPauseCompleteState, Event: common.StressTestRecvPauseEvent,
			Dst: common.NotifyStressTestState, Handler: ctl.notifyStressTest},

		{Src: common.NotifyStressTestState, Event: common.NotifySuccessEvent,
			Dst: common.WaitStressTestFinishedState, Handler: ctl.handleWaitStressTestFinish},
		{Src: common.NotifyStressTestState, Event: common.NotifyFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},

		{Src: common.WaitStressTestFinishedState, Event: common.ReceiveReportEvent,
			Dst: common.NotifyContinueTrainState, Handler: ctl.notifyContinueTrain},
		{Src: common.WaitStressTestFinishedState, Event: common.ReportTimeoutEvent,
			Dst: common.NotifyStopTrainState, Handler: ctl.handleNotifyStopTrain},
		{Src: common.WaitStressTestFinishedState, Event: common.StressTestFailEvent,
			Dst: common.NotifyStopTrainState, Handler: ctl.handleStressTestFail},

		{Src: common.WaitContinueTrainState, Event: common.StressTestRecvContinueEvent,
			Dst: common.InitState, Handler: ctl.handleStressTestFinish},
	}
}

func (ctl *EventController) getSwitchNicRules() []common.TransRule {
	return []common.TransRule{
		{Src: common.InitState, Event: common.StartSwitchNic,
			Dst: common.NotifyPauseTrainState, Handler: ctl.handleNotifyPauseTrain},

		{Src: common.WaitReportPauseCompleteState, Event: common.SwitchNicRecvPauseEvent,
			Dst: common.NotifySwitchNicState, Handler: ctl.notifySwitchNic},

		{Src: common.NotifySwitchNicState, Event: common.NotifySuccessEvent,
			Dst: common.WaitSwitchNicFinishedState, Handler: ctl.handleWaitSwitchNicFinish},
		{Src: common.NotifySwitchNicState, Event: common.NotifyFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},

		{Src: common.WaitSwitchNicFinishedState, Event: common.ReceiveReportEvent,
			Dst: common.NotifyContinueTrainState, Handler: ctl.notifyContinueTrain},
		{Src: common.WaitSwitchNicFinishedState, Event: common.WaitSwitchNicRecvFaultEvent,
			Dst: common.NotifyStopTrainState, Handler: ctl.handleNotifyStopTrain},
		{Src: common.WaitSwitchNicFinishedState, Event: common.ReportTimeoutEvent,
			Dst: common.NotifyStopTrainState, Handler: ctl.handleNotifyStopTrain},
		{Src: common.WaitSwitchNicFinishedState, Event: common.SwitchNicFailEvent,
			Dst: common.NotifyStopTrainState, Handler: ctl.handleNotifyStopTrain},

		{Src: common.WaitContinueTrainState, Event: common.SwitchNicRecvContinueEvent,
			Dst: common.InitState, Handler: ctl.handleSwitchNicFinish},
	}
}

func (ctl *EventController) getOMRules() []common.TransRule {
	return []common.TransRule{
		{Src: common.NotifyPauseTrainState, Event: common.NotifySuccessEvent,
			Dst: common.WaitReportPauseCompleteState, Handler: ctl.handleWaitPauseTrainComplete},
		{Src: common.NotifyPauseTrainState, Event: common.NotifyFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},

		{Src: common.WaitReportPauseCompleteState, Event: common.ReportTimeoutEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},
		{Src: common.WaitReportPauseCompleteState, Event: common.ProcessPauseFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},

		{Src: common.NotifyContinueTrainState, Event: common.NotifyContinueSuccessEvent,
			Dst: common.WaitContinueTrainState, Handler: ctl.handleDecideContinueTrainComplete},
		{Src: common.NotifyContinueTrainState, Event: common.NotifyFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},

		{Src: common.WaitContinueTrainState, Event: common.ReportTimeoutEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},
		{Src: common.WaitContinueTrainState, Event: common.ContinueTrainFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},
	}
}

func (ctl *EventController) getHotSwitchRules() []common.TransRule {
	return []common.TransRule{
		{Src: common.InitState, Event: common.BeginHotSwitchEvent,
			Dst: common.NotifyPrepareHotSwitchState, Handler: ctl.notifyPrepareHotSwitch},

		{Src: common.NotifyPrepareHotSwitchState, Event: common.NotifySuccessEvent,
			Dst: common.WaitNewPodState, Handler: ctl.notifyCreateNewPod},
		{Src: common.NotifyPrepareHotSwitchState, Event: common.NotifyFailEvent,
			Dst: common.InitState, Handler: ctl.cleanStateWhenFailed},

		{Src: common.WaitNewPodState, Event: common.NewPodRunningEvent,
			Dst: common.WaitNotifyPodRunningResultState, Handler: ctl.notifyNewPodRunningHandler},
		{Src: common.WaitNewPodState, Event: common.NewPodTimeoutEvent,
			Dst: common.WaitNotifyPodFailedResultState, Handler: ctl.notifyNewPodFailedHandler},

		{Src: common.WaitNotifyPodFailedResultState, Event: common.NotifySuccessEvent,
			Dst: common.InitState, Handler: ctl.cleanStateWhenFailed},
		{Src: common.WaitNotifyPodFailedResultState, Event: common.NotifyFailEvent,
			Dst: common.InitState, Handler: ctl.cleanStateWhenFailed},

		{Src: common.WaitNotifyPodRunningResultState, Event: common.NotifySuccessEvent,
			Dst: common.WaitReportPauseResultState, Handler: ctl.waitReportPauseTrainResult},
		{Src: common.WaitNotifyPodRunningResultState, Event: common.NotifyFailEvent,
			Dst: common.WaitReportPauseResultState, Handler: ctl.waitReportPauseTrainResult},

		{Src: common.WaitReportPauseResultState, Event: common.ReportTimeoutEvent,
			Dst: common.NotifyStopJobState, Handler: ctl.notifyStopJob},
		{Src: common.WaitReportPauseResultState, Event: common.ExitEvent,
			Dst: common.NotifyStopJobState, Handler: ctl.notifyStopJob},
		{Src: common.WaitReportPauseResultState, Event: common.MigrationEvent,
			Dst: common.WaitOldPodDeletedState, Handler: ctl.notifyDeleteOldPod},

		{Src: common.NotifyStopJobState, Event: common.NotifySuccessEvent,
			Dst: common.InitState, Handler: ctl.cleanStateWhenFailed},
		{Src: common.NotifyStopJobState, Event: common.NotifyFailEvent,
			Dst: common.InitState, Handler: ctl.cleanStateWhenFailed},

		{Src: common.WaitOldPodDeletedState, Event: common.OldPodDeletedEvent,
			Dst: common.WaitNotifyRestartTrainResultState, Handler: ctl.notifyRestartTrain},

		{Src: common.WaitNotifyRestartTrainResultState, Event: common.NotifySuccessEvent,
			Dst: common.WaitReportRestartTrainResultState, Handler: ctl.handleWaitReportRestartTrainStatus},
		{Src: common.WaitNotifyRestartTrainResultState, Event: common.NotifyFailEvent,
			Dst: common.NotifyStopJobState, Handler: ctl.notifyStopJob},

		{Src: common.WaitReportRestartTrainResultState, Event: common.RestartSuccessEvent,
			Dst: common.InitState, Handler: ctl.cleanStateWhenSuccess},
		{Src: common.WaitReportRestartTrainResultState, Event: common.RestartFaildEvent,
			Dst: common.NotifyStopJobState, Handler: ctl.notifyStopJob},
		{Src: common.WaitReportRestartTrainResultState, Event: common.ReportTimeoutEvent,
			Dst: common.NotifyStopJobState, Handler: ctl.notifyStopJob},
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
			Dst: common.NotifyDumpState, Handler: ctl.handleNotifyElagantDump},
		{Src: common.NotifyWaitFaultFlushingState, Event: common.NotifyFailEvent,
			Dst: common.NotifyKillJobState, Handler: ctl.handleKillJob},

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

func (ctl *EventController) getDPScaleStrategyRules() []common.TransRule {
	return []common.TransRule{
		{Src: common.NotifyDecidedStrategyState, Event: common.NotifyScaleInStrategySuccessEvent,
			Dst: common.WaitReportScaleInIsolateRanksState, Handler: ctl.handleWaitReportScaleInIsolateRanksStatus},

		{Src: common.WaitReportProcessRecoverStatusState, Event: common.NeedTryScaleInStrategyEvent,
			Dst: common.NotifyScaleInStrategyState, Handler: ctl.handleNotifyScaleInStrategy},

		{Src: common.NotifyScaleInStrategyState, Event: common.NotifyScaleInStrategySuccessEvent,
			Dst: common.WaitReportScaleInIsolateRanksState, Handler: ctl.handleWaitReportScaleInIsolateRanksStatus},
		{Src: common.NotifyScaleInStrategyState, Event: common.NotifyFailEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},

		{Src: common.WaitReportScaleInIsolateRanksState, Event: common.ReceiveReportEvent,
			Dst: common.CheckReportScaleInIsolateRanksState, Handler: ctl.handleCheckRecoverResult},
		{Src: common.WaitReportScaleInIsolateRanksState, Event: common.ReportTimeoutEvent,
			Dst: common.NotifyDecidedStrategyState, Handler: ctl.handleNotifyDecidedStrategy},

		{Src: common.CheckReportScaleInIsolateRanksState, Event: common.NotifyFailEvent,
			Dst: common.NotifyDecidedStrategyState, Handler: ctl.handleNotifyDecidedStrategy},
		{Src: common.CheckReportScaleInIsolateRanksState, Event: common.NotifyFaultNodesExitSuccessEvent,
			Dst: common.WaitReportScaleInStatusState, Handler: ctl.handleWaitReportScaleInStatus},
		{Src: common.CheckReportScaleInIsolateRanksState, Event: common.RecoverFailEvent,
			Dst: common.NotifyDecidedStrategyState, Handler: ctl.handleNotifyDecidedStrategy},

		{Src: common.WaitReportScaleInStatusState, Event: common.ReceiveReportEvent,
			Dst: common.CheckRecoverResultState, Handler: ctl.handleCheckRecoverResult},
		{Src: common.WaitReportScaleInStatusState, Event: common.ReportTimeoutEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},

		{Src: common.CheckRecoverResultState, Event: common.ScaleInSuccessEvent,
			Dst: common.ScaleInRunningState, Handler: ctl.handleScaleInRunningState},
		{Src: common.CheckRecoverResultState, Event: common.ScaleOutSuccessEvent,
			Dst: common.InitState, Handler: ctl.handleFinish},

		{Src: common.ScaleInRunningState, Event: common.NotifyScaleOutStrategySuccessEvent,
			Dst: common.WaitReportScaleOutStatusState, Handler: ctl.handleWaitReportScaleOutStatusState},
		{Src: common.WaitReportScaleOutStatusState, Event: common.ReceiveReportEvent,
			Dst: common.CheckRecoverResultState, Handler: ctl.handleCheckRecoverResult},
		{Src: common.WaitReportScaleOutStatusState, Event: common.ReportTimeoutEvent,
			Dst: common.FaultClearState, Handler: ctl.handleFaultClear},
		{Src: common.CheckRecoverResultState, Event: common.ScaleOutSuccessEvent,
			Dst: common.InitState, Handler: ctl.handleFinish},

		{Src: common.ScaleInRunningState, Event: common.FaultOccurEvent,
			Dst: common.NotifyStopTrainState, Handler: ctl.handleNotifyStopTrain},
		{Src: common.ScaleInRunningState, Event: common.FinishEvent,
			Dst: common.InitState, Handler: ctl.handleFinish},
		{Src: common.ScaleInRunningState, Event: common.NotifyFailEvent,
			Dst: common.ScaleInRunningState, Handler: ctl.handleScaleInRunningState},
	}
}

func (ctl *EventController) getBaseRules() []common.TransRule {
	var rules []common.TransRule
	rules = append(rules, ctl.getPreRules()...)
	rules = append(rules, ctl.getExtendPreRules()...)
	rules = append(rules, ctl.getFixRules()...)
	rules = append(rules, ctl.getAfterRules()...)
	rules = append(rules, ctl.getOMRules()...)
	rules = append(rules, ctl.getSwitchNicRules()...)
	rules = append(rules, ctl.getStressTestRules()...)
	rules = append(rules, ctl.getDPScaleStrategyRules()...)
	rules = append(rules, ctl.getHotSwitchRules()...)
	return rules
}
