/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package multilevelscheduling for scheduling NPU job with general abstract network topology configuration.
*/
package multilevelscheduling

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
)

// MultilevelHandler represents the multilevel scheduling handler
type MultilevelHandler struct {
	base.NPUHandler
	taskLevels []util.TaskTreeLevel
}

// schedulingTreeNode contains all node data for scheduling
type schedulingTreeNode struct {
	// the depth of root node is 0
	depth    int
	node     *util.ResourceNode
	parent   *schedulingTreeNode
	children map[string]*schedulingTreeNode

	// fields for resource fragments
	hasTraversed         bool
	allocatableTaskCount int
	fragmentScore        int
	freeSubTasks         []*util.TaskNode

	// fields for resource reservation
	isReserved                    bool
	hasSufficientReservedResource bool
}

// schedulingTreeLevel contains all level data for scheduling
type schedulingTreeLevel struct {
	taskLevel     util.TaskTreeLevel
	resourceLevel util.ResourceTreeLevel
	nodes         []*schedulingTreeNode
}

// schedulingTree the scheduling state tree
type schedulingTree struct {
	root   *schedulingTreeNode
	levels []*schedulingTreeLevel
}

const (
	// MultiLevelHandlerName name of multilevel schedule handler
	MultiLevelHandlerName = "multilevel"

	nodeNpuNotMatchError      = "node usable npu not match task npu num"
	jobCheckFailedReason      = "npu num is invalid"
	blockInvalidReason        = "block config is invalid"
	scoreForNode              = 100000000
	maxNodeNpu                = 16
	sessionsForSinglePod      = 6
	defaultFragmentScoreScale = 10
)
