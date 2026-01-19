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

// Package stoptrain for stop train plugin

package stoptrain

import (
	"reflect"
	"sync"
	"testing"

	"ascend-common/common-utils/hwlog"
	clusterdconstant "clusterd/pkg/common/constant"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	pluginutils "taskd/framework_backend/manager/plugins/utils"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want infrastructure.ManagerPlugin
	}{
		{
			name: "get plugin object",
			want: &stopTrainingPlugin{
				HasSendMessages: make(map[string]string),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStopTrainingPluginName(t *testing.T) {
	type fields struct {
		hasToken        bool
		shot            storage.SnapShot
		signalInfo      *pluginutils.SignalInfo
		HasSendMessages map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "get plugin name",
			fields: fields{},
			want:   constant.StopTrainPluginName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stopTrainingPlugin{}
			if got := s.Name(); got != tt.want {
				t.Errorf("stopTrainingPlugin.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fieldsTestStopTrainingPluginPredicate struct {
	hasToken        bool
	shot            storage.SnapShot
	signalInfo      *pluginutils.SignalInfo
	HasSendMessages map[string]string
}
type argsTestStopTrainingPluginPredicate struct {
	shot storage.SnapShot
}

type testsTestStopTrainingPluginPredicate struct {
	name    string
	fields  fieldsTestStopTrainingPluginPredicate
	args    argsTestStopTrainingPluginPredicate
	want    infrastructure.PredicateResult
	wantErr bool
}

func TestStopTrainingPluginPredicate(t *testing.T) {
	tests := []testsTestStopTrainingPluginPredicate{
		{
			name:   "case 1: has token",
			fields: fieldsTestStopTrainingPluginPredicate{hasToken: true},
			want: infrastructure.PredicateResult{
				PluginName:      constant.StopTrainPluginName,
				CandidateStatus: constant.CandidateStatus,
				PredicateStream: map[string]string{constant.ResumeTrainingAfterFaultStream: ""}},
			wantErr: false},
		{
			name: "case 2: getSignalInfo error",
			want: infrastructure.PredicateResult{
				PluginName:      constant.StopTrainPluginName,
				CandidateStatus: constant.UnselectStatus},
			wantErr: false},
		{
			name:   "case 3: apply token",
			fields: fieldsTestStopTrainingPluginPredicate{HasSendMessages: make(map[string]string)},
			args: argsTestStopTrainingPluginPredicate{
				shot: storage.SnapShot{
					ClusterInfos: &storage.ClusterInfos{
						Clusters: map[string]*storage.ClusterInfo{
							constant.ClusterDRank: {
								Command: map[string]string{
									constant.SignalType:  clusterdconstant.StopTrainSignalType,
									constant.Timeout:     "0",
									constant.Actions:     utils.ObjToString([]string{}),
									constant.FaultRanks:  utils.ObjToString(map[int]int{}),
									constant.NodeRankIds: utils.ObjToString([]string{})}}}}}},
			want: infrastructure.PredicateResult{
				PluginName:      constant.StopTrainPluginName,
				CandidateStatus: constant.CandidateStatus,
				PredicateStream: map[string]string{constant.ResumeTrainingAfterFaultStream: ""}},
			wantErr: false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stopTrainingPlugin{
				hasToken:        tt.fields.hasToken,
				shot:            tt.fields.shot,
				HasSendMessages: tt.fields.HasSendMessages}
			got, err := s.Predicate(tt.args.shot)
			if (err != nil) != tt.wantErr {
				t.Errorf("stopTrainingPlugin.Predicate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stopTrainingPlugin.Predicate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStopTrainingPluginRelease(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "case 1: release", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stopTrainingPlugin{}
			if err := s.Release(); (err != nil) != tt.wantErr {
				t.Errorf("stopTrainingPlugin.Release() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type fieldsTestStopTrainingPluginHandle struct {
	hasToken        bool
	shot            storage.SnapShot
	signalInfo      *pluginutils.SignalInfo
	HasSendMessages map[string]string
}

type argsTestStopTrainingPluginHandle struct {
	name    string
	fields  fieldsTestStopTrainingPluginHandle
	want    infrastructure.HandleResult
	wantErr bool
}

func TestStopTrainingPluginHandle(t *testing.T) {
	tests := []argsTestStopTrainingPluginHandle{
		{name: "case 1: handle final",
			fields: fieldsTestStopTrainingPluginHandle{HasSendMessages: make(map[string]string),
				shot: storage.SnapShot{
					ClusterInfos: &storage.ClusterInfos{
						Clusters: map[string]*storage.ClusterInfo{
							constant.ClusterDRank: {
								Command: map[string]string{
									constant.SignalType:  clusterdconstant.GlobalFaultSignalType,
									constant.Timeout:     "0",
									constant.Actions:     utils.ObjToString([]string{}),
									constant.FaultRanks:  utils.ObjToString(map[int]int{}),
									constant.NodeRankIds: utils.ObjToString([]string{})}}},
						RWMutex: sync.RWMutex{}}}},
			want:    infrastructure.HandleResult{Stage: constant.HandleStageFinal},
			wantErr: false},
		{name: "case 2: handle process",
			fields: fieldsTestStopTrainingPluginHandle{HasSendMessages: make(map[string]string),
				shot: storage.SnapShot{
					ClusterInfos: &storage.ClusterInfos{
						Clusters: map[string]*storage.ClusterInfo{
							constant.ClusterDRank: {
								Command: map[string]string{}}},
						RWMutex: sync.RWMutex{}}}},
			want:    infrastructure.HandleResult{Stage: constant.HandleStageProcess},
			wantErr: false},
		{
			name:    "case 3: handle exception",
			fields:  fieldsTestStopTrainingPluginHandle{HasSendMessages: make(map[string]string), shot: storage.SnapShot{}},
			want:    infrastructure.HandleResult{Stage: constant.HandleStageException},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stopTrainingPlugin{
				hasToken:        tt.fields.hasToken,
				shot:            tt.fields.shot,
				HasSendMessages: tt.fields.HasSendMessages,
			}
			got, err := s.Handle()
			if (err != nil) != tt.wantErr {
				t.Errorf("stopTrainingPlugin.Handle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stopTrainingPlugin.Handle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStopTrainingPluginPullMsg(t *testing.T) {
	type fields struct {
		hasToken        bool
		shot            storage.SnapShot
		signalInfo      *pluginutils.SignalInfo
		HasSendMessages map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []infrastructure.Msg
		wantErr bool
	}{{name: "get all type msgs",
		fields: fields{
			signalInfo: &pluginutils.SignalInfo{
				SignalType: clusterdconstant.GlobalFaultSignalType,
				Actions:    []string{clusterdconstant.OnGlobalRankAction},
				FaultRanks: map[int]int{}},
			HasSendMessages: make(map[string]string)},
		want: []infrastructure.Msg{{Receiver: []string{constant.ControllerName},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.ProcessManageRecoverSignal,
				Extension: map[string]string{
					constant.SignalType: clusterdconstant.GlobalFaultSignalType,
					constant.Actions:    utils.ObjToString([]string{clusterdconstant.OnGlobalRankAction}),
					constant.FaultRanks: "",
					constant.Timeout:    "0",
				}}}},
		wantErr: false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stopTrainingPlugin{
				hasToken:        tt.fields.hasToken,
				shot:            tt.fields.shot,
				signalInfo:      tt.fields.signalInfo,
				HasSendMessages: tt.fields.HasSendMessages,
			}
			got, err := s.PullMsg()
			if (err != nil) != tt.wantErr {
				t.Errorf("stopTrainingPlugin.PullMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stopTrainingPlugin.PullMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}
