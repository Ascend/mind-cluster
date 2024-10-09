/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package test is using for HuaWei Ascend testing.
*/
package test

import (
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// InitMyJobPluginTestCase test case
type InitMyJobPluginTestCase struct {
	Name    string
	Attr    util.SchedulerJobAttr
	Env     plugin.ScheduleEnv
	WantErr error
}

// ValidNPUJobTestCase validNPUJob test case
type ValidNPUJobTestCase struct {
	WantErr *api.ValidateResult
	Name    string
	Attr    util.SchedulerJobAttr
}

// CheckNodeNPUByTaskTestCase CheckNodeNPUByTask test case
type CheckNodeNPUByTaskTestCase struct {
	Task    *api.TaskInfo
	Name    string
	Attr    util.SchedulerJobAttr
	Node    plugin.NPUNode
	WantErr error
}

// ScoreBestNPUNodesTestCase scoreBestNPUNodes test case
type ScoreBestNPUNodesTestCase struct {
	Task     *api.TaskInfo
	Nodes    []*api.NodeInfo
	ScoreMap map[string]float64
	WantSMap map[string]float64
	Name     string
	WantErr  error
	Attr     util.SchedulerJobAttr
}

// UseAnnotationTestCase useAnnotation test case
type UseAnnotationTestCase struct {
	Task     *api.TaskInfo
	WantNode *plugin.NPUNode
	Name     string
	Node     plugin.NPUNode
	PodAnno  string
	Attr     util.SchedulerJobAttr
}

// JudgeNodeAndTaskNPUTestCase JudgeNodeAndTaskNPU test case
type JudgeNodeAndTaskNPUTestCase struct {
	NodeTop []int
	Name    string
	TaskNPU int
	WantErr error
}

// SetMaxNodeNPUNumTestCase  SetMaxNodeNPUNum test case
type SetMaxNodeNPUNumTestCase struct {
	Name    string
	Num     int
	WantNum int
}

// ReleaseAnnotationTestCase releaseAnnotation test case
type ReleaseAnnotationTestCase struct {
	Task     *api.TaskInfo
	WantNode *plugin.NPUNode
	Name     string
	Node     plugin.NPUNode
	PodAnno  string
	Attr     util.SchedulerJobAttr
}
