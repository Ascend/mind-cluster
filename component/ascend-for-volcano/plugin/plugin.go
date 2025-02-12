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

// NPUBuilder PluginBuilder plugin management
type NPUBuilder = func(string2 string) ISchedulerPlugin

// ISchedulerPluginBase the frame plugin need implement.
type ISchedulerPluginBase interface {
	GetPluginName() string
	SetPluginName(string)
	GetAnnoPreVal() string
	SetAnnoPreVal(string)
	GetAnnoName() string
	SetAnnoName(string)
	GetDefaultJobSchedulerConfig() map[string]string
	SetDefaultJobSchedulerConfig(map[string]string)
}

// ISchedulerPluginNeed The interface that the specific plug-in needs to implement.
type ISchedulerPluginNeed interface {
	// ValidNPUJob Valid the job part of npu scheduler policy, if not, disallowed.
	ValidNPUJob() *api.ValidateResult
	CheckNodeNPUByTask(*api.TaskInfo, NPUNode) error
	ScoreBestNPUNodes(*api.TaskInfo, []*api.NodeInfo, map[string]float64) error
	UseAnnotation(*api.TaskInfo, NPUNode) *NPUNode
	ReleaseAnnotation(*api.TaskInfo, NPUNode) *NPUNode
	PreStartAction(i interface{}, ssn *framework.Session) error
	PreStopAction(*ScheduleEnv) error
	InitMyJobPlugin(util.SchedulerJobAttr, ScheduleEnv) error
	GetReHandle() interface{}
}

// ISchedulerPlugin for volcano-npu plugin has function.
type ISchedulerPlugin interface {
	ISchedulerPluginBase
	ISchedulerPluginNeed
}

// SchedulerPlugin for all volcano-npu plugin.
type SchedulerPlugin struct {
	// the new func add name
	pluginName string
	// in k8s annotation huawei.com/Ascend310,huawei.com/Ascend910
	annoName string
	// huawei.com/
	annoPreVal string
	// config like arm x86
	defaultJobSchedulerConfig map[string]string
}

// GetPluginName get PluginName.
func (sp SchedulerPlugin) GetPluginName() string {
	return sp.pluginName
}

// SetPluginName set PluginName.
func (sp *SchedulerPlugin) SetPluginName(name string) {
	if sp == nil {
		klog.V(util.LogInfoLev).Infof("SetPluginName failed: %s.", util.ArgumentError)
		return
	}
	sp.pluginName = name
}

// GetAnnoPreVal get AnnoPreVal.
func (sp SchedulerPlugin) GetAnnoPreVal() string {
	return sp.annoPreVal
}

// SetAnnoPreVal set AnnoPreVal.
func (sp *SchedulerPlugin) SetAnnoPreVal(value string) {
	if sp == nil {
		klog.V(util.LogInfoLev).Infof("SetAnnoPreVal failed: %s.", util.ArgumentError)
		return
	}
	sp.annoPreVal = value
}

// GetAnnoName get AnnoName.
func (sp SchedulerPlugin) GetAnnoName() string {
	return sp.annoName
}

// SetAnnoName set AnnoName.
func (sp *SchedulerPlugin) SetAnnoName(annoName string) {
	if sp == nil {
		klog.V(util.LogInfoLev).Infof("SetAnnoName failed: %s.", util.ArgumentError)
		return
	}
	sp.annoName = annoName
}

// GetDefaultJobSchedulerConfig get DefaultJobSchedulerConfig.
func (sp SchedulerPlugin) GetDefaultJobSchedulerConfig() map[string]string {
	return sp.defaultJobSchedulerConfig
}

// SetDefaultJobSchedulerConfig set DefaultJobSchedulerConfig.
func (sp *SchedulerPlugin) SetDefaultJobSchedulerConfig(conf map[string]string) {
	if sp == nil {
		klog.V(util.LogInfoLev).Infof("SetDefaultJobSchedulerConfig failed: %s.", util.ArgumentError)
		return
	}

	if len(sp.defaultJobSchedulerConfig) == 0 {
		sp.defaultJobSchedulerConfig = make(map[string]string, util.MapInitNum)
	}
	sp.defaultJobSchedulerConfig = conf
}
