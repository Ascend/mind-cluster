/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"taskd/framework_backend/manager/infrastructure/storage"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
}

type argsTestParseSignalInfo struct {
	clusterInfo *storage.ClusterInfo
}

type testCaseTestParseSignalInfo struct {
	name    string
	args    argsTestParseSignalInfo
	want    *SignalInfo
	wantErr bool
}

var testsTestParseSignalInfo = []testCaseTestParseSignalInfo{
	{
		name: "case 1: normal signal info with all fields",
		args: argsTestParseSignalInfo{
			clusterInfo: &storage.ClusterInfo{
				Command: map[string]string{
					constant.SignalType:     clusterdconstant.ChangeStrategySignalType,
					constant.ChangeStrategy: clusterdconstant.ScaleInStrategyName,
					constant.ExtraParams:    "extra_params",
					constant.Uuid:           "test-uuid",
					constant.Timeout:        "60",
					constant.Actions:        utils.ObjToString([]string{"action1", "action2"}),
					constant.FaultRanks:     utils.ObjToString(map[int]int{1: 1, 2: 2}),
					constant.NodeRankIds:    utils.ObjToString([]string{"1", "2"}),
				},
			},
		},
		want: &SignalInfo{
			SignalType:     clusterdconstant.ChangeStrategySignalType,
			ChangeStrategy: clusterdconstant.ScaleInStrategyName,
			ExtraParams:    "extra_params",
			Uuid:           "test-uuid",
			Timeout:        60,
			Actions:        []string{"action1", "action2"},
			FaultRanks:     map[int]int{1: 1, 2: 2},
			NodeRankIds:    []string{"1", "2"},
			Command: map[string]string{
				constant.SignalType:     clusterdconstant.ChangeStrategySignalType,
				constant.ChangeStrategy: clusterdconstant.ScaleInStrategyName,
				constant.ExtraParams:    "extra_params",
				constant.Uuid:           "test-uuid",
				constant.Timeout:        "60",
				constant.Actions:        utils.ObjToString([]string{"action1", "action2"}),
				constant.FaultRanks:     utils.ObjToString(map[int]int{1: 1, 2: 2}),
				constant.NodeRankIds:    utils.ObjToString([]string{"1", "2"}),
			},
		},
		wantErr: false,
	},
	{
		name: "case 2: cluster info is nil",
		args: argsTestParseSignalInfo{
			clusterInfo: nil,
		},
		want:    nil,
		wantErr: true,
	},
}

func TestParseSignalInfo(t *testing.T) {
	for _, tt := range testsTestParseSignalInfo {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSignalInfo(tt.args.clusterInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSignalInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got.SignalType, tt.want.SignalType) {
					t.Errorf("ParseSignalInfo() SignalType = %v, want %v", got.SignalType, tt.want.SignalType)
				}
				if !reflect.DeepEqual(got.ChangeStrategy, tt.want.ChangeStrategy) {
					t.Errorf("ParseSignalInfo() ChangeStrategy = %v, want %v", got.ChangeStrategy, tt.want.ChangeStrategy)
				}
				if !reflect.DeepEqual(got.ExtraParams, tt.want.ExtraParams) {
					t.Errorf("ParseSignalInfo() ExtraParams = %v, want %v", got.ExtraParams, tt.want.ExtraParams)
				}
				if !reflect.DeepEqual(got.Uuid, tt.want.Uuid) {
					t.Errorf("ParseSignalInfo() Uuid = %v, want %v", got.Uuid, tt.want.Uuid)
				}
				if !reflect.DeepEqual(got.Timeout, tt.want.Timeout) {
					t.Errorf("ParseSignalInfo() Timeout = %v, want %v", got.Timeout, tt.want.Timeout)
				}
				if !reflect.DeepEqual(got.Actions, tt.want.Actions) {
					t.Errorf("ParseSignalInfo() Actions = %v, want %v", got.Actions, tt.want.Actions)
				}
				if !reflect.DeepEqual(got.FaultRanks, tt.want.FaultRanks) {
					t.Errorf("ParseSignalInfo() FaultRanks = %v, want %v", got.FaultRanks, tt.want.FaultRanks)
				}
				if !reflect.DeepEqual(got.NodeRankIds, tt.want.NodeRankIds) {
					t.Errorf("ParseSignalInfo() NodeRankIds = %v, want %v", got.NodeRankIds, tt.want.NodeRankIds)
				}
				if !reflect.DeepEqual(got.Command, tt.want.Command) {
					t.Errorf("ParseSignalInfo() Command = %v, want %v", got.Command, tt.want.Command)
				}
			}
		})
	}
}

type argsTestParseSignalFields struct {
	signalInfo *SignalInfo
	command    map[string]string
}

type testCaseTestParseSignalFields struct {
	name    string
	args    argsTestParseSignalFields
	wantErr bool
}

var testsTestParseSignalFields = []testCaseTestParseSignalFields{
	{
		name: "case 1: normal signal fields",
		args: argsTestParseSignalFields{
			signalInfo: &SignalInfo{},
			command: map[string]string{
				constant.Timeout:     "60",
				constant.Actions:     utils.ObjToString([]string{"action1"}),
				constant.FaultRanks:  utils.ObjToString(map[int]int{1: 1}),
				constant.NodeRankIds: utils.ObjToString([]string{"1"}),
			},
		},
		wantErr: false,
	},
	{
		name: "case 2: timeout parse error",
		args: argsTestParseSignalFields{
			signalInfo: &SignalInfo{},
			command: map[string]string{
				constant.Timeout: "invalid",
			},
		},
		wantErr: true,
	},
	{
		name: "case 3: actions parse error",
		args: argsTestParseSignalFields{
			signalInfo: &SignalInfo{},
			command: map[string]string{
				constant.Timeout: "60",
				constant.Actions: "invalid",
			},
		},
		wantErr: true,
	},
	{
		name: "case 4: fault ranks parse error",
		args: argsTestParseSignalFields{
			signalInfo: &SignalInfo{},
			command: map[string]string{
				constant.Timeout:     "60",
				constant.Actions:     utils.ObjToString([]string{}),
				constant.FaultRanks:  "invalid",
				constant.NodeRankIds: utils.ObjToString([]string{}),
			},
		},
		wantErr: true,
	},
}

func TestParseSignalFields(t *testing.T) {
	for _, tt := range testsTestParseSignalFields {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseSignalFields(tt.args.signalInfo, tt.args.command); (err != nil) != tt.wantErr {
				t.Errorf("parseSignalFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
