// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package service a series of service function
package service

import "clusterd/pkg/interface/grpc/common"

/*
getBaseRules return machine state change rules for retry/recover/dump/exit strategy
src: origin state.
event: event. occur on origin state
dst: destination state. when even happen on src state, state will change to dst state, and take handle function
*/

func (ctl *EventController) getPreRules() []common.TransRule {
	return []common.TransRule{
		{Src: common.InitState, Event: common.FaultOccurEvent,
			Dst: common.NotifyWaitFaultFlushingState, Handler: ctl.handleNotifyWaitFaultFlushing},

		{Src: common.NotifyWaitFaultFlushingState, Event: common.NotifyFinishEvent,
			Dst: common.NotifyStopTrainState, Handler: ctl.handleNotifyStopTrain},

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

		{Src: common.WaitReportRecoverStrategyState, Event: common.ReceiveReportEvent,
			Dst: common.NotifyDecidedStrategyState, Handler: ctl.handleNotifyDecidedStrategy},
		{Src: common.WaitReportRecoverStrategyState, Event: common.ReportTimeoutEvent,
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

func (ctl *EventController) getBaseRules() []common.TransRule {
	var rules []common.TransRule
	rules = append(rules, ctl.getPreRules()...)
	return rules
}
