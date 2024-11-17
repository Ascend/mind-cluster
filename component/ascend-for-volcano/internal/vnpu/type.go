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
Package vnpu is using for HuaWei Ascend pin vnpu allocation.
*/
package vnpu

import (
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	// PodEventMsgNoResourceFailed dp pod segment failed msg for not enough resource.
	PodEventMsgNoResourceFailed = "NoNPUResource"
	// PodEventMsgDyCutFailed dp pod segment failed msg for DCMI failed.
	PodEventMsgDyCutFailed = "NPUSegmentFailed"
	// PodEventReasonAllocateFailed dp pod segment failed reason
	PodEventReasonAllocateFailed = "UnexpectedAdmissionError"
	// DyCutFailedError for device-plugin cut failed error.
	DyCutFailedError = "chipDyCutErr"
	// PodEventTypeAllocateFailed dp pod segment failed type
	PodEventTypeAllocateFailed = "Warning"
	podObjectType              = "Pod"
)

// VTemplate vNPU template
type VTemplate struct {
	Data map[string]util.VResource
}

// VirtualNPU vnpu struct
type VirtualNPU struct {
	StaticByConf bool
	VT           VTemplate
	StaticVNPU
	DynamicVNPU
}

// StaticVNPU Static VNPU struct.
type StaticVNPU struct {
	vnpuHandler
}

// DynamicVNPU dynamic VNPU struct.
type DynamicVNPU struct {
	vnpuHandler
	DowngradeCache map[string][]string // taskName: nodes
	// for Concurrent task. not same core request task only has one on a node in same time.
	// nodeName: templateName:taskUID
	ConCache map[string]map[string]map[api.TaskID]struct{}
}

type vnpuHandler interface {
	CheckNodeNPUByTask(*api.TaskInfo, plugin.NPUNode, util.VResource) error
	ScoreBestNPUNodes(*api.TaskInfo, []*api.NodeInfo, map[string]float64) error
	UseAnnotation(*api.TaskInfo, plugin.NPUNode, util.VResource, VTemplate) *plugin.NPUNode
}

// Action vnpu actions
type Action struct {
	template map[string]util.VResource
}
