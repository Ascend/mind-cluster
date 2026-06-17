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
Package plugin is using for HuaWei Ascend pin affinity schedule frame.
*/
package plugin

import (
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// PolicyBuilder PolicyBuilder plugin management
type PolicyBuilder = func() SchedulerPluginNeed

// SchedulerPluginBase the frame plugin need implement.
type SchedulerPluginBase interface {
	GetPluginName() string
	SetPluginName(string)
	GetAnnoPreVal(string) string
	SetAnnoPreVal(string)
	GetAnnoName(string) string
	SetAnnoName(string)
}

// SchedulerPluginNeed The interface that the specific plug-in needs to implement.
type SchedulerPluginNeed interface {
	// ValidNPUJob Valid the job part of npu scheduler policy, if not, disallowed.
	ValidNPUJob() *api.ValidateResult
	CheckNodeNPUByTask(*api.TaskInfo, NPUNode) error
	ScoreBestNPUNodes(*api.TaskInfo, []*api.NodeInfo, map[string]float64) error
	UseAnnotation(*api.TaskInfo, NPUNode) *NPUNode
	ReleaseAnnotation(*api.TaskInfo, NPUNode) *NPUNode
	PreStartAction(ssn *framework.Session) error
	InitMyJobPlugin(util.SchedulerJobAttr, ScheduleEnv) error
	GetMaxCardNPUNum() int
	Preemptable(preemptor *api.TaskInfo, preemptees []*api.TaskInfo, vcNode *NPUNode) ([]*api.TaskInfo, bool)
}

// BackupPodAllocatedHook is an optional interface for policy handlers that need
// to update internal state when a hot-switch backup pod is allocated.
// Only multi-level scheduling implements this to refresh SuperPods with the
// backup pod's node, since the original fault pod will be deleted.
type BackupPodAllocatedHook interface {
	OnBackupPodAllocated(task *api.TaskInfo, job *SchedulerJob, nodeName string)
}

// SchedulerPlugin for volcano-npu plugin has function.
type SchedulerPlugin interface {
	SchedulerPluginBase
	SchedulerPluginNeed
}

// FaultHandler fault handler for job
type FaultHandler interface {
	Execute(*ScheduleEnv, *framework.Session) error
	CheckNodeNPUByTask(*api.TaskInfo, *NPUNode) error
	ScoreBestNPUNodes(*api.TaskInfo, map[string]float64)
	UseAnnotation(*api.TaskInfo)
	PreStopAction(*ScheduleEnv) error
	// IsNodeFault returns true if the node has any registered fault
	// (hard fault, card sub-health, or switch sub-health).
	IsNodeFault(nodeName string) bool
	// IsFaultTaskByRank returns true if the given rank was recorded as a fault
	// task in the fault job cache. Uses the fault snapshot rather than
	// real-time node health.
	IsFaultTaskByRank(jobID api.JobID, rankIndex string) bool
}

// SchedulerBaseAttr for all volcano-npu plugin.
type SchedulerBaseAttr struct {
	// the new func add name
	pluginName string
	// in k8s annotation huawei.com/Ascend310,huawei.com/Ascend910
	annoName string
	// huawei.com/
	annoPreVal string
}

// GetPluginName get PluginName.
func (sp SchedulerBaseAttr) GetPluginName() string {
	return sp.pluginName
}

// SetPluginName set PluginName.
func (sp *SchedulerBaseAttr) SetPluginName(name string) {
	if sp == nil {
		klog.V(util.LogInfoLev).Infof("SetPluginName failed: %s.", util.ArgumentError)
		return
	}
	sp.pluginName = name
}

// GetAnnoPreVal get AnnoPreVal.
func (sp SchedulerBaseAttr) GetAnnoPreVal(reqNPUName string) string {
	klog.V(util.LogDebugLev).Infof("GetAnnoPreVal reqNPUName: %s.", reqNPUName)
	if reqNPUName == util.NPUCardName {
		return util.NPUCardNamePre
	}
	if reqNPUName == util.NPU910CardName {
		return util.NPU910CardNamePre
	}
	return sp.annoPreVal
}

// SetAnnoPreVal set AnnoPreVal.
func (sp *SchedulerBaseAttr) SetAnnoPreVal(value string) {
	if sp == nil {
		klog.V(util.LogInfoLev).Infof("SetAnnoPreVal failed: %s.", util.ArgumentError)
		return
	}
	sp.annoPreVal = value
}

// GetAnnoName get AnnoName.
func (sp SchedulerBaseAttr) GetAnnoName(reqNPUName string) string {
	klog.V(util.LogDebugLev).Infof("GetAnnoName reqNPUName: %s.", reqNPUName)
	if reqNPUName == util.NPUCardName {
		return util.NPUCardName
	}
	if reqNPUName == util.NPU910CardName {
		return util.NPU910CardName
	}
	return sp.annoName
}

// SetAnnoName set AnnoName.
func (sp *SchedulerBaseAttr) SetAnnoName(annoName string) {
	if sp == nil {
		klog.V(util.LogInfoLev).Infof("SetAnnoName failed: %s.", util.ArgumentError)
		return
	}
	sp.annoName = annoName
}
