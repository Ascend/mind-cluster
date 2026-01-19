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

// Package elastictraining for elastic training plugin
package elastictraining

import (
	"reflect"
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
	}{{
		name: "get plugin object",
		want: &elasticTrainingPlugin{
			HasSendMessages: make(map[string]string),
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestElasticTrainingPluginName(t *testing.T) {
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
	}{{
		name:   "get plugin name",
		fields: fields{},
		want:   constant.ElasticTrainingPluginName,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &elasticTrainingPlugin{}
			if got := s.Name(); got != tt.want {
				t.Errorf("elasticTrainingPlugin.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fieldsTestElasticTrainingPluginPredicate struct {
	hasToken        bool
	shot            storage.SnapShot
	signalInfo      *pluginutils.SignalInfo
	HasSendMessages map[string]string
}

type argsTestElasticTrainingPluginPredicate struct {
	shot storage.SnapShot
}

type testsTestElasticTrainingPluginPredicate struct {
	name    string
	fields  fieldsTestElasticTrainingPluginPredicate
	args    argsTestElasticTrainingPluginPredicate
	want    infrastructure.PredicateResult
	wantErr bool
}

func getTestElasticTrainingPluginPredicateTests() []testsTestElasticTrainingPluginPredicate {
	return []testsTestElasticTrainingPluginPredicate{
		{
			name:   "case 1: has token",
			fields: fieldsTestElasticTrainingPluginPredicate{hasToken: true},
			want: infrastructure.PredicateResult{
				PluginName:      constant.ElasticTrainingPluginName,
				CandidateStatus: constant.CandidateStatus,
				PredicateStream: map[string]string{constant.ResumeTrainingAfterFaultStream: ""}},
			wantErr: false},
		{
			name:   "case 2: getSignalInfo error",
			fields: fieldsTestElasticTrainingPluginPredicate{},
			args: argsTestElasticTrainingPluginPredicate{
				shot: storage.SnapShot{ClusterInfos: &storage.ClusterInfos{
					Clusters: make(map[string]*storage.ClusterInfo),
				}},
			},
			want: infrastructure.PredicateResult{
				PluginName:      constant.ElasticTrainingPluginName,
				CandidateStatus: constant.UnselectStatus},
			wantErr: false},
		{
			name:   "case 3: apply token for scale in",
			fields: fieldsTestElasticTrainingPluginPredicate{HasSendMessages: make(map[string]string)},
			args: argsTestElasticTrainingPluginPredicate{
				shot: storage.SnapShot{
					ClusterInfos: &storage.ClusterInfos{
						Clusters: map[string]*storage.ClusterInfo{
							constant.ClusterDRank: {
								Command: map[string]string{
									constant.SignalType:     clusterdconstant.ChangeStrategySignalType,
									constant.ChangeStrategy: clusterdconstant.ScaleInStrategyName,
									constant.Timeout:        "0",
									constant.Actions:        utils.ObjToString([]string{}),
									constant.FaultRanks:     utils.ObjToString(map[int]int{}),
									constant.NodeRankIds:    utils.ObjToString([]string{})}}}}}},
			want: infrastructure.PredicateResult{
				PluginName:      constant.ElasticTrainingPluginName,
				CandidateStatus: constant.CandidateStatus,
				PredicateStream: map[string]string{constant.ResumeTrainingAfterFaultStream: ""}},
			wantErr: false},
	}
}

func getTestElasticTrainingPluginPredicateTests2() []testsTestElasticTrainingPluginPredicate {
	return []testsTestElasticTrainingPluginPredicate{
		{
			name:   "case 4: apply token for scale out",
			fields: fieldsTestElasticTrainingPluginPredicate{HasSendMessages: make(map[string]string)},
			args: argsTestElasticTrainingPluginPredicate{
				shot: storage.SnapShot{
					ClusterInfos: &storage.ClusterInfos{
						Clusters: map[string]*storage.ClusterInfo{
							constant.ClusterDRank: {
								Command: map[string]string{
									constant.SignalType:     clusterdconstant.ChangeStrategySignalType,
									constant.ChangeStrategy: clusterdconstant.ScaleOutStrategyName,
									constant.Timeout:        "0",
									constant.Actions:        utils.ObjToString([]string{}),
									constant.FaultRanks:     utils.ObjToString(map[int]int{}),
									constant.NodeRankIds:    utils.ObjToString([]string{})}}}}}},
			want: infrastructure.PredicateResult{
				PluginName:      constant.ElasticTrainingPluginName,
				CandidateStatus: constant.CandidateStatus,
				PredicateStream: map[string]string{constant.ResumeTrainingAfterFaultStream: ""}},
			wantErr: false},
		{
			name:   "case 5: invalid signal type",
			fields: fieldsTestElasticTrainingPluginPredicate{HasSendMessages: make(map[string]string)},
			args: argsTestElasticTrainingPluginPredicate{
				shot: storage.SnapShot{
					ClusterInfos: &storage.ClusterInfos{
						Clusters: map[string]*storage.ClusterInfo{
							constant.ClusterDRank: {
								Command: map[string]string{
									constant.SignalType:  "invalid_signal",
									constant.Timeout:     "0",
									constant.Actions:     utils.ObjToString([]string{}),
									constant.FaultRanks:  utils.ObjToString(map[int]int{}),
									constant.NodeRankIds: utils.ObjToString([]string{})}}}}}},
			want: infrastructure.PredicateResult{
				PluginName:      constant.ElasticTrainingPluginName,
				CandidateStatus: constant.UnselectStatus},
			wantErr: false},
	}
}

func TestElasticTrainingPluginPredicate(t *testing.T) {
	tests := getTestElasticTrainingPluginPredicateTests()
	tests = append(tests, getTestElasticTrainingPluginPredicateTests2()...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &elasticTrainingPlugin{
				hasToken:        tt.fields.hasToken,
				shot:            tt.fields.shot,
				HasSendMessages: tt.fields.HasSendMessages}
			got, err := s.Predicate(tt.args.shot)
			if (err != nil) != tt.wantErr {
				t.Errorf("elasticTrainingPlugin.Predicate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("elasticTrainingPlugin.Predicate() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fieldsTestElasticTrainingPluginHandle struct {
	hasToken        bool
	shot            storage.SnapShot
	signalInfo      *pluginutils.SignalInfo
	HasSendMessages map[string]string
}

type argsTestElasticTrainingPluginHandle struct {
	name    string
	fields  fieldsTestElasticTrainingPluginHandle
	want    infrastructure.HandleResult
	wantErr bool
}

func TestElasticTrainingPluginHandle(t *testing.T) {
	tests := []argsTestElasticTrainingPluginHandle{
		{name: "case 1: handle final - invalid signal type",
			fields: fieldsTestElasticTrainingPluginHandle{HasSendMessages: make(map[string]string),
				signalInfo: &pluginutils.SignalInfo{
					SignalType: "invalid_signal",
					Actions:    []string{},
				}},
			want:    infrastructure.HandleResult{Stage: constant.HandleStageFinal},
			wantErr: false},
		{name: "case 2: handle process - scale in",
			fields: fieldsTestElasticTrainingPluginHandle{HasSendMessages: make(map[string]string),
				signalInfo: &pluginutils.SignalInfo{
					SignalType:     clusterdconstant.ChangeStrategySignalType,
					ChangeStrategy: clusterdconstant.ScaleInStrategyName,
					Actions:        []string{},
				}},
			want:    infrastructure.HandleResult{Stage: constant.HandleStageProcess},
			wantErr: false},
		{name: "case 3: handle process - scale out",
			fields: fieldsTestElasticTrainingPluginHandle{HasSendMessages: make(map[string]string),
				signalInfo: &pluginutils.SignalInfo{
					SignalType:     clusterdconstant.ChangeStrategySignalType,
					ChangeStrategy: clusterdconstant.ScaleOutStrategyName,
					Actions:        []string{},
				}},
			want:    infrastructure.HandleResult{Stage: constant.HandleStageProcess},
			wantErr: false},
		{name: "case 4: handle exception - get signal info error",
			fields: fieldsTestElasticTrainingPluginHandle{HasSendMessages: make(map[string]string),
				shot: storage.SnapShot{ClusterInfos: &storage.ClusterInfos{
					Clusters: make(map[string]*storage.ClusterInfo)}}},
			want:    infrastructure.HandleResult{Stage: constant.HandleStageException},
			wantErr: false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &elasticTrainingPlugin{
				hasToken:        tt.fields.hasToken,
				shot:            tt.fields.shot,
				signalInfo:      tt.fields.signalInfo,
				HasSendMessages: tt.fields.HasSendMessages}
			got, err := s.Handle()
			if (err != nil) != tt.wantErr {
				t.Errorf("elasticTrainingPlugin.Handle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("elasticTrainingPlugin.Handle() = %v, want %v", got, tt.want)
			}
		})
	}
}

type TestElasticTrainingPluginPullMsgFields struct {
	hasToken        bool
	shot            storage.SnapShot
	signalInfo      *pluginutils.SignalInfo
	HasSendMessages map[string]string
}

type testElasticTrainingPluginPullMsgTests struct {
	name    string
	fields  TestElasticTrainingPluginPullMsgFields
	want    []infrastructure.Msg
	wantErr bool
}

func getTestElasticTrainingPluginPullMsgTests() []testElasticTrainingPluginPullMsgTests {
	return []testElasticTrainingPluginPullMsgTests{{
		name: "case 1: pull change strategy message",
		fields: TestElasticTrainingPluginPullMsgFields{
			signalInfo: &pluginutils.SignalInfo{
				Uuid:           "test-uuid",
				SignalType:     clusterdconstant.ChangeStrategySignalType,
				ChangeStrategy: clusterdconstant.ScaleInStrategyName,
				Actions:        []string{clusterdconstant.ChangeStrategyAction},
				Command:        make(map[string]string),
			},
			HasSendMessages: make(map[string]string),
		},
		want: []infrastructure.Msg{{
			Receiver: []string{constant.ControllerName},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.ProcessManageRecoverSignal,
				Extension: map[string]string{
					constant.SignalType:     clusterdconstant.ChangeStrategySignalType,
					constant.Actions:        utils.ObjToString([]string{clusterdconstant.ChangeStrategyAction}),
					constant.ChangeStrategy: clusterdconstant.ScaleInStrategyName,
					constant.ExtraParams:    "",
				},
			},
		}},
		wantErr: false},
	}
}

func getTestElasticTrainingPluginPullMsgTests2() []testElasticTrainingPluginPullMsgTests {
	return []testElasticTrainingPluginPullMsgTests{
		{
			name: "case 2: pull fault nodes exit message",
			fields: TestElasticTrainingPluginPullMsgFields{
				signalInfo: &pluginutils.SignalInfo{
					Uuid:        "test-uuid",
					SignalType:  clusterdconstant.FaultNodesExitSignalType,
					Actions:     []string{clusterdconstant.FaultNodesExitAction},
					NodeRankIds: []string{"1"},
					Command:     make(map[string]string),
				},
				HasSendMessages: make(map[string]string),
			},
			want: []infrastructure.Msg{{
				Receiver: []string{"Agent1"},
				Body: storage.MsgBody{
					MsgType: constant.Action,
					Code:    constant.ExitAgentCode,
					Extension: map[string]string{
						constant.SignalType: clusterdconstant.FaultNodesExitSignalType,
						constant.Actions:    utils.ObjToString([]string{clusterdconstant.FaultNodesExitAction}),
					},
				},
			}},
			wantErr: false},
		{
			name: "case 3: signal info is nil",
			fields: TestElasticTrainingPluginPullMsgFields{
				signalInfo:      nil,
				HasSendMessages: make(map[string]string),
			},
			want:    nil,
			wantErr: false},
		{
			name: "case 4: message already sent",
			fields: TestElasticTrainingPluginPullMsgFields{
				signalInfo: &pluginutils.SignalInfo{
					Uuid: "test-uuid",
				},
				HasSendMessages: map[string]string{utils.ObjToString([]string{}): ""},
			},
			want:    make([]infrastructure.Msg, 0),
			wantErr: false},
	}
}

func TestElasticTrainingPluginPullMsg(t *testing.T) {
	tests := getTestElasticTrainingPluginPullMsgTests()
	tests = append(tests, getTestElasticTrainingPluginPullMsgTests2()...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &elasticTrainingPlugin{
				hasToken:        tt.fields.hasToken,
				shot:            tt.fields.shot,
				signalInfo:      tt.fields.signalInfo,
				HasSendMessages: tt.fields.HasSendMessages,
			}
			got, err := s.PullMsg()
			if (err != nil) != tt.wantErr {
				t.Errorf("elasticTrainingPlugin.PullMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("elasticTrainingPlugin.PullMsg() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

type fieldsTestElasticTrainingPluginGetSignalInfo struct {
	hasToken        bool
	shot            storage.SnapShot
	signalInfo      *pluginutils.SignalInfo
	HasSendMessages map[string]string
}

type testsTestElasticTrainingPluginGetSignalInfo struct {
	name    string
	fields  fieldsTestElasticTrainingPluginGetSignalInfo
	wantErr bool
}

func getTestsTestElasticTrainingPluginGetSignalInfo() []testsTestElasticTrainingPluginGetSignalInfo {
	return []testsTestElasticTrainingPluginGetSignalInfo{
		{
			name: "case 1: normal get signal info",
			fields: fieldsTestElasticTrainingPluginGetSignalInfo{
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
								}}}}}},
			wantErr: false},
		{
			name: "case 2: cluster info is nil",
			fields: fieldsTestElasticTrainingPluginGetSignalInfo{
				shot: storage.SnapShot{
					ClusterInfos: &storage.ClusterInfos{
						Clusters: map[string]*storage.ClusterInfo{},
					},
				},
			},
			wantErr: true},
		{
			name: "case 3: timeout parse error",
			fields: fieldsTestElasticTrainingPluginGetSignalInfo{
				shot: storage.SnapShot{
					ClusterInfos: &storage.ClusterInfos{
						Clusters: map[string]*storage.ClusterInfo{
							constant.ClusterDRank: {
								Command: map[string]string{
									constant.SignalType: clusterdconstant.ChangeStrategySignalType,
									constant.Timeout:    "invalid",
								}}}}}},
			wantErr: true}}
}

func TestElasticTrainingPluginGetSignalInfo(t *testing.T) {
	tests := getTestsTestElasticTrainingPluginGetSignalInfo()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &elasticTrainingPlugin{
				hasToken:        tt.fields.hasToken,
				shot:            tt.fields.shot,
				signalInfo:      tt.fields.signalInfo,
				HasSendMessages: tt.fields.HasSendMessages,
			}
			if err := s.getSignalInfo(); (err != nil) != tt.wantErr {
				t.Errorf("elasticTrainingPlugin.getSignalInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
