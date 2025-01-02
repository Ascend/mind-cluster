package vnpu

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
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
