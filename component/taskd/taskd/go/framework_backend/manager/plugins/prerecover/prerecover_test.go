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

// Package prerecover for preparation before recovery,such as stoptrain, globalfault, etc

package prerecover

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
			want: &preRecoverPlugin{
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

func TestPreRecoverPluginName(t *testing.T) {
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
			s := &preRecoverPlugin{}
			if got := s.Name(); got != tt.want {
				t.Errorf("preOperationPlugin.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fieldsTestPreRecoverPluginPredicate struct {
	hasToken        bool
	shot            storage.SnapShot
	signalInfo      *pluginutils.SignalInfo
	HasSendMessages map[string]string
	lastUuid        string
}
type argsTestPreRecoverPluginPredicate struct {
	shot storage.SnapShot
}

type testsTestPreRecoverPluginPredicate struct {
	name    string
	fields  fieldsTestPreRecoverPluginPredicate
	args    argsTestPreRecoverPluginPredicate
	want    infrastructure.PredicateResult
	wantErr bool
}

func TestPreRecoverPluginPredicate(t *testing.T) {
	tests := getArgsTestPreRecoverPluginPredicateTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &preRecoverPlugin{
				hasToken:        tt.fields.hasToken,
				shot:            tt.fields.shot,
				HasSendMessages: tt.fields.HasSendMessages,
				lastUuid:        tt.fields.lastUuid,
			}
			got, err := s.Predicate(tt.args.shot)
			if (err != nil) != tt.wantErr {
				t.Errorf("preOperationPlugin.Predicate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("preOperationPlugin.Predicate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getArgsTestPreRecoverPluginPredicateTestCases() []testsTestPreRecoverPluginPredicate {
	candidateResult := infrastructure.PredicateResult{
		PluginName:      constant.StopTrainPluginName,
		CandidateStatus: constant.CandidateStatus,
		PredicateStream: map[string]string{constant.ResumeTrainingAfterFaultStream: ""},
	}
	return []testsTestPreRecoverPluginPredicate{
		{
			name:    "case 1: has token",
			fields:  fieldsTestPreRecoverPluginPredicate{hasToken: true},
			want:    candidateResult,
			wantErr: false},
		{
			name: "case 2: getSignalInfo error",
			want: infrastructure.PredicateResult{
				PluginName:      constant.StopTrainPluginName,
				CandidateStatus: constant.UnselectStatus},
			wantErr: false},
		{
			name:    "case 3: apply token for stop_train signal",
			fields:  fieldsTestPreRecoverPluginPredicate{HasSendMessages: make(map[string]string)},
			args:    getArgsTestPreRecoverPluginPredicate(clusterdconstant.StopTrainSignalType),
			want:    candidateResult,
			wantErr: false},
		{
			name: "case 4: apply token for global_fault signal",
			fields: fieldsTestPreRecoverPluginPredicate{HasSendMessages: make(map[string]string),
				lastUuid: "randomUuid"},
			args:    getArgsTestPreRecoverPluginPredicate(clusterdconstant.GlobalFaultSignalType),
			want:    candidateResult,
			wantErr: false},
		{
			name: "case 5: apply token for pre_exit_process signal",
			fields: fieldsTestPreRecoverPluginPredicate{HasSendMessages: make(map[string]string),
				lastUuid: "randomUuid"},
			args:    getArgsTestPreRecoverPluginPredicate(clusterdconstant.PreExitProcessSignalType),
			want:    candidateResult,
			wantErr: false}}
}

func getArgsTestPreRecoverPluginPredicate(signalType string) argsTestPreRecoverPluginPredicate {
	return argsTestPreRecoverPluginPredicate{
		shot: storage.SnapShot{
			ClusterInfos: &storage.ClusterInfos{
				Clusters: map[string]*storage.ClusterInfo{
					constant.ClusterDRank: {
						Command: map[string]string{
							constant.SignalType:  signalType,
							constant.Timeout:     "0",
							constant.Actions:     utils.ObjToString([]string{}),
							constant.FaultRanks:  utils.ObjToString(map[int]int{}),
							constant.NodeRankIds: utils.ObjToString([]string{})}}}}}}
}

func TestPreRecoverPluginRelease(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "case 1: release", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &preRecoverPlugin{}
			if err := s.Release(); (err != nil) != tt.wantErr {
				t.Errorf("preOperationPlugin.Release() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type fieldsTestPreRecoverPluginHandle struct {
	hasToken        bool
	shot            storage.SnapShot
	signalInfo      *pluginutils.SignalInfo
	HasSendMessages map[string]string
}

type argsTestPreRecoverPluginHandle struct {
	name    string
	fields  fieldsTestPreRecoverPluginHandle
	want    infrastructure.HandleResult
	wantErr bool
}

func TestPreRecoverPluginHandle(t *testing.T) {
	tests := []argsTestPreRecoverPluginHandle{
		{name: "case 1: handle final",
			fields: fieldsTestPreRecoverPluginHandle{HasSendMessages: make(map[string]string),
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
			fields: fieldsTestPreRecoverPluginHandle{HasSendMessages: make(map[string]string),
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
			fields:  fieldsTestPreRecoverPluginHandle{HasSendMessages: make(map[string]string), shot: storage.SnapShot{}},
			want:    infrastructure.HandleResult{Stage: constant.HandleStageException},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &preRecoverPlugin{
				hasToken:        tt.fields.hasToken,
				shot:            tt.fields.shot,
				HasSendMessages: tt.fields.HasSendMessages,
			}
			got, err := s.Handle()
			if (err != nil) != tt.wantErr {
				t.Errorf("preOperationPlugin.Handle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("preOperationPlugin.Handle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPreRecoverPluginPullMsg(t *testing.T) {
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
			s := &preRecoverPlugin{
				hasToken:        tt.fields.hasToken,
				shot:            tt.fields.shot,
				signalInfo:      tt.fields.signalInfo,
				HasSendMessages: tt.fields.HasSendMessages,
			}
			got, err := s.PullMsg()
			if (err != nil) != tt.wantErr {
				t.Errorf("preOperationPlugin.PullMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("preOperationPlugin.PullMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fieldsTestPreRecoverPluginGetSignalInfo struct {
	hasToken        bool
	shot            storage.SnapShot
	signalInfo      *pluginutils.SignalInfo
	HasSendMessages map[string]string
	lastUuid        string
}

var testsTestPreRecoverPluginGetSignalInfoCases = []struct {
	name    string
	fields  fieldsTestPreRecoverPluginGetSignalInfo
	wantErr bool
}{
	{
		name: "case 1: normal get signal info",
		fields: fieldsTestPreRecoverPluginGetSignalInfo{
			shot: storage.SnapShot{
				ClusterInfos: &storage.ClusterInfos{
					Clusters: map[string]*storage.ClusterInfo{
						constant.ClusterDRank: {
							Command: map[string]string{
								constant.SignalType:     clusterdconstant.ChangeStrategySignalType,
								constant.ChangeStrategy: clusterdconstant.ScaleInStrategyName,
								constant.Timeout:        "60",
								constant.Actions:        utils.ObjToString([]string{"action1"}),
								constant.FaultRanks:     utils.ObjToString(map[int]int{1: 1}),
								constant.NodeRankIds:    utils.ObjToString([]string{"1"}),
								constant.ExtraParams:    "extra",
								constant.Uuid:           "test-uuid",
							}}}}}},
		wantErr: false},
	{
		name: "case 2: cluster info is nil",
		fields: fieldsTestPreRecoverPluginGetSignalInfo{
			shot: storage.SnapShot{
				ClusterInfos: &storage.ClusterInfos{
					Clusters: map[string]*storage.ClusterInfo{},
				},
			},
		},
		wantErr: true},
	{
		name: "case 4: empty signal type",
		fields: fieldsTestPreRecoverPluginGetSignalInfo{
			shot: storage.SnapShot{
				ClusterInfos: &storage.ClusterInfos{
					Clusters: map[string]*storage.ClusterInfo{
						constant.ClusterDRank: {
							Command: map[string]string{
								constant.Uuid: "test-uuid",
							}}}}}},
		wantErr: false},
}

func TestPreRecoverPluginGetSignalInfo(t *testing.T) {
	for _, tt := range testsTestPreRecoverPluginGetSignalInfoCases {
		t.Run(tt.name, func(t *testing.T) {
			s := &preRecoverPlugin{
				hasToken:        tt.fields.hasToken,
				shot:            tt.fields.shot,
				signalInfo:      tt.fields.signalInfo,
				HasSendMessages: tt.fields.HasSendMessages,
				lastUuid:        tt.fields.lastUuid,
			}
			if err := s.getSignalInfo(); (err != nil) != tt.wantErr {
				t.Errorf("getSignalInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
