/*
Copyright(C)2020-2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package vnpu is using for HuaWei Ascend pin vnpu allocation.
*/
package vnpu

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

type staticVNPUArgs struct {
	task     *api.TaskInfo
	node     plugin.NPUNode
	scoreMap map[string]float64
}

type staticVNPUTestCase struct {
	name    string
	args    staticVNPUArgs
	wantErr bool
	wantRes bool
}

func buildStaticVNPUTestCases() []staticVNPUTestCase {
	return []staticVNPUTestCase{
		{
			name:    "01 will return err when task is nil",
			args:    staticVNPUArgs{task: nil, node: plugin.NPUNode{}},
			wantErr: true,
			wantRes: true,
		},
		{
			name:    "02 will return nil when task is not nil",
			args:    staticVNPUArgs{task: &api.TaskInfo{}, node: plugin.NPUNode{}},
			wantErr: false,
			wantRes: true,
		},
	}
}

func TestCheckNodeNPUByTask(t *testing.T) {
	tp := &StaticVNPU{}
	tests := buildStaticVNPUTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tp.CheckNodeNPUByTask(tt.args.task, tt.args.node, util.VResource{})
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScoreBestNPUNodes(t *testing.T) {
	tp := &StaticVNPU{}
	tests := buildStaticVNPUTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tp.ScoreBestNPUNodes(tt.args.task, []*api.NodeInfo{}, tt.args.scoreMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScoreBestNPUNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStaticVNPUUseAnnotation(t *testing.T) {
	tp := &StaticVNPU{}
	tests := buildStaticVNPUTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tp.UseAnnotation(tt.args.task, tt.args.node, util.VResource{}, VTemplate{})
			if (res != nil) != tt.wantRes {
				t.Errorf("UseAnnotation() res = %v, wantRes %v", res, tt.wantRes)
			}
		})
	}
}
