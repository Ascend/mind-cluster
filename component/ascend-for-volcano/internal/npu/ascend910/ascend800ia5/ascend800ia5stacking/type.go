/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package ascend800ia5stacking defines the core types, structures and constants
// for Ascend 800i A5 stacking scenarios in the Volcano scheduler plugin.
// It includes module definitions, resource constraints, and base structures
// required for node management and NPU resource allocation.
package ascend800ia5stacking

import (
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

type module800ia5stacking struct {
	Base800ia5
	netUnhealthyKey  string
	SuperPodCache    map[int32][]plugin.NPUNode
	NPUSelectedCache map[api.JobID]map[int32][]int
	// PickedNodeCache the available or used node name set. eg. {nodeName : true}
	PickedNodeCache map[string]bool
}

const (
	nodeNPUNumber      = 8
	stackingNodeNumber = 2
	// Ascend800ia5stacking value of accelerator-type is 800i-stacking-a5-8
	Ascend800ia5stacking = "800I-Stacking-A5-8"
	// InferSchedulerName name of scheduler
	InferSchedulerName = "huawei.com/Ascend910" + Ascend800ia5stacking
)

// Base800ia5 for Ascend 800ia5 base.
type Base800ia5 struct {
	base.NPUHandler
	AffScoreList     [][]int
	NpuNumInvalidMap map[int]struct{}
	arch             string
}
