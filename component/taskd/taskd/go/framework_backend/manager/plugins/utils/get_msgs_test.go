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
	"reflect"
	"testing"

	"ascend-common/common-utils/hwlog"
	clusterdconstant "clusterd/pkg/common/constant"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
}

func TestSignalInfoGetMsgs(t *testing.T) {
	type fields struct {
		SignalType     string
		Actions        []string
		FaultRanks     map[int]int
		ChangeStrategy string
		Timeout        int64
		NodeRankIds    []string
		ExtraParams    string
		Command        map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   []infrastructure.Msg
	}{
		{
			name: "test get all msgs",
			fields: fields{
				Actions: []string{clusterdconstant.StopAction,
					clusterdconstant.FaultNodesExitAction, clusterdconstant.OnGlobalRankAction,
					clusterdconstant.FaultNodesRestartAction, clusterdconstant.ChangeStrategyAction},
				FaultRanks:  map[int]int{},
				NodeRankIds: []string{"1"},
				Command:     map[string]string{},
			},
			want: getSignalInfoGetMsgsTestWant()}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SignalInfo{
				SignalType:     tt.fields.SignalType,
				Actions:        tt.fields.Actions,
				FaultRanks:     tt.fields.FaultRanks,
				ChangeStrategy: tt.fields.ChangeStrategy,
				Timeout:        tt.fields.Timeout,
				NodeRankIds:    tt.fields.NodeRankIds,
				ExtraParams:    tt.fields.ExtraParams,
				Command:        tt.fields.Command,
			}
			if got := s.GetMsgs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMsgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getSignalInfoGetMsgsTestWant() []infrastructure.Msg {
	return []infrastructure.Msg{{
		Receiver: []string{constant.ControllerName},
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    constant.ProcessManageRecoverSignal,
			Extension: map[string]string{
				constant.SignalType: "",
				constant.Actions:    utils.ObjToString([]string{clusterdconstant.StopAction}),
				constant.FaultRanks: ""}}}, {
		Receiver: []string{common.AgentRole + "1"},
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    constant.ExitAgentCode,
			Extension: map[string]string{
				constant.SignalType: "",
				constant.Actions:    utils.ObjToString([]string{clusterdconstant.FaultNodesExitAction})}}},
		{
			Receiver: []string{constant.ControllerName},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.ProcessManageRecoverSignal,
				Extension: map[string]string{
					constant.SignalType: "",
					constant.Actions:    utils.ObjToString([]string{clusterdconstant.OnGlobalRankAction}),
					constant.FaultRanks: ""}}}, {
			Receiver: []string{common.AgentRole + "1"},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.RestartWorkersCode,
				Extension: map[string]string{
					constant.SignalType: "",
					constant.Actions:    utils.ObjToString([]string{clusterdconstant.FaultNodesRestartAction}),
					constant.FaultRanks: ""}}}, {
			Receiver: []string{constant.ControllerName},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.ProcessManageRecoverSignal,
				Extension: map[string]string{
					constant.SignalType:     "",
					constant.Actions:        utils.ObjToString([]string{clusterdconstant.ChangeStrategyAction}),
					constant.ChangeStrategy: "",
					constant.ExtraParams:    ""}}}}
}
