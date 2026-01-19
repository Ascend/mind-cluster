/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package utils for common func
package utils

import (
	"strconv"

	clusterdconstant "clusterd/pkg/common/constant"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

// SignalInfo signal info define
type SignalInfo struct {
	SignalType     string
	Actions        []string
	FaultRanks     map[int]int
	ChangeStrategy string
	Timeout        int64
	NodeRankIds    []string
	ExtraParams    string
	Command        map[string]string
	Uuid           string
}

// GetMsgs returns msgs by actions
func (s *SignalInfo) GetMsgs() []infrastructure.Msg {
	msgs := make([]infrastructure.Msg, 0)
	for _, action := range s.Actions {
		if action == clusterdconstant.StopAction {
			msgs = append(msgs, s.getStopTrainActionMsgs()...)
		} else if action == clusterdconstant.PauseTrainAction {
			msgs = append(msgs, s.getPauseTrainActionMsgs()...)
		} else if action == clusterdconstant.FaultNodesExitAction {
			msgs = append(msgs, s.getFaultNodesExitActionMsgs()...)
		} else if action == clusterdconstant.OnGlobalRankAction {
			msgs = append(msgs, s.getOnGlobalRankActionMsgs()...)
		} else if action == clusterdconstant.FaultNodesRestartAction {
			msgs = append(msgs, s.getFaultNodesRestartActionMsgs()...)
		} else if action == clusterdconstant.ChangeStrategyAction {
			msgs = append(msgs, s.getChangeStrategyActionMsgs()...)
		}
	}
	return msgs
}

func (s *SignalInfo) getPauseTrainActionMsgs() []infrastructure.Msg {
	return []infrastructure.Msg{
		{
			Receiver: []string{constant.ControllerName},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.ProcessManageRecoverSignal,
				Extension: map[string]string{
					constant.SignalType: s.SignalType,
					constant.Actions:    utils.ObjToString([]string{clusterdconstant.PauseTrainAction}),
					constant.FaultRanks: s.Command[constant.FaultRanks],
					constant.Timeout:    strconv.FormatInt(s.Timeout, constant.Dec),
				},
			},
		},
	}
}

func (s *SignalInfo) getStopTrainActionMsgs() []infrastructure.Msg {
	return []infrastructure.Msg{
		{
			Receiver: []string{constant.ControllerName},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.ProcessManageRecoverSignal,
				Extension: map[string]string{
					constant.SignalType: s.SignalType,
					constant.Actions:    utils.ObjToString([]string{clusterdconstant.StopAction}),
					constant.FaultRanks: s.Command[constant.FaultRanks],
					constant.Timeout:    strconv.FormatInt(s.Timeout, constant.Dec),
				},
			},
		},
	}
}

func (s *SignalInfo) getFaultNodesExitActionMsgs() []infrastructure.Msg {
	msgs := make([]infrastructure.Msg, 0)
	for _, nodeRankId := range s.NodeRankIds {
		msgs = append(msgs, infrastructure.Msg{
			Receiver: []string{common.AgentRole + nodeRankId},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.ExitAgentCode,
				Extension: map[string]string{
					constant.SignalType: s.SignalType,
					constant.Actions:    utils.ObjToString([]string{clusterdconstant.FaultNodesExitAction}),
				},
			},
		})
	}
	return msgs
}

func (s *SignalInfo) getFaultNodesRestartActionMsgs() []infrastructure.Msg {
	msgs := make([]infrastructure.Msg, 0)
	for _, nodeRankId := range s.NodeRankIds {
		msgs = append(msgs, infrastructure.Msg{
			Receiver: []string{common.AgentRole + nodeRankId},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.RestartWorkersCode,
				Extension: map[string]string{
					constant.SignalType: s.SignalType,
					constant.Actions:    utils.ObjToString([]string{clusterdconstant.FaultNodesRestartAction}),
					constant.FaultRanks: s.Command[constant.FaultRanks],
				},
			},
		})
	}
	return msgs
}

func (s *SignalInfo) getOnGlobalRankActionMsgs() []infrastructure.Msg {
	return []infrastructure.Msg{
		{
			Receiver: []string{constant.ControllerName},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.ProcessManageRecoverSignal,
				Extension: map[string]string{
					constant.SignalType: s.SignalType,
					constant.Actions:    utils.ObjToString([]string{clusterdconstant.OnGlobalRankAction}),
					constant.FaultRanks: s.Command[constant.FaultRanks],
					constant.Timeout:    strconv.FormatInt(s.Timeout, constant.Dec),
				},
			},
		},
	}
}

func (s *SignalInfo) getChangeStrategyActionMsgs() []infrastructure.Msg {
	return []infrastructure.Msg{
		{
			Receiver: []string{constant.ControllerName},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.ProcessManageRecoverSignal,
				Extension: map[string]string{
					constant.SignalType:     s.SignalType,
					constant.Actions:        utils.ObjToString([]string{clusterdconstant.ChangeStrategyAction}),
					constant.ChangeStrategy: s.ChangeStrategy,
					constant.ExtraParams:    s.ExtraParams,
				},
			},
		},
	}
}
