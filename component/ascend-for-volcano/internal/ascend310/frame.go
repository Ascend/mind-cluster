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
Package ascend310 is using for HuaWei A300T pin affinity schedule.
*/
package ascend310

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend310/card310x4"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend310/chip310x4"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// Name This need by frame init plugin.
func (tp *asend310) Name() string {
	return PluginName
}

// New return npu plugin.
func New(npuName string) plugin.ISchedulerPlugin {
	var npuPlugin = asend310{}
	npuPlugin.SetPluginName(npuName)
	npuPlugin.SetAnnoName(util.NPU310CardName)
	npuPlugin.SetAnnoPreVal(util.NPU310CardNamePre)
	npuPlugin.SetDefaultJobSchedulerConfig(nil)

	npuPlugin.Kind = map[string]base.AscendHandler{}
	npuPlugin.Kind[chip310x4.SchedulerName] = chip310x4.New(chip310x4.SchedulerName)
	npuPlugin.Kind[card310x4.SchedulerName] = card310x4.New(card310x4.SchedulerName)

	return &npuPlugin
}

// InitMyJobPlugin for 310 job init
func (tp *asend310) InitMyJobPlugin(attr util.SchedulerJobAttr, env plugin.ScheduleEnv) error {
	if tp == nil {
		mgs := fmt.Errorf("nil plugin %s", PluginName)
		klog.V(util.LogInfoLev).Infof("InitMyJobPlugin %v.", mgs)
		return mgs
	}
	tp.SetSchedulerAttr(attr)
	tp.SetSchedulerEnv(env)
	klog.V(util.LogDebugLev).Infof("InitMyJobPlugin attr label %#v.", attr.Label)
	v, ok := attr.Label[Accelerator310Key]
	if !ok {
		v = Chip310AcceleratorValue
	}
	value, ok := tp.Kind[attr.ReqNPUName+v]
	if !ok {
		return fmt.Errorf("not support %s", attr.ReqNPUName+v)
	}
	if err := value.InitMyJobPlugin(attr, env); err != nil {
		return err
	}

	tp.handle = value

	return nil
}

// ValidNPUJob check job req npu num
func (tp *asend310) ValidNPUJob() *api.ValidateResult {
	if tp == nil {
		err := errors.New(util.ArgumentError)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: err.Error(),
		}
	}
	if tp.handle != nil {
		tp.handle.ValidNPUJob()
	}
	klog.V(util.LogDebugLev).Infof("%s ValidNPUJob handle is nil", tp.GetPluginName())
	return nil
}

// PreStartAction pre-processing actions for rescheduling
func (tp *asend310) PreStartAction(i interface{}, _ *framework.Session) error {
	k, ok := i.(*rescheduling.ReScheduler)
	if !ok {
		return fmt.Errorf("preStartAction failed %s, interface is not ReScheduler", PluginName)
	}
	tp.reHandle = k
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *asend310) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", PluginName, err.Error())
		return err
	}
	if tp.handle != nil {
		if err := tp.handle.CheckNodeNPUByTask(task, node); err != nil {
			return err
		}
	}
	klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask handle is nil", PluginName)
	return nil
}

// ScoreBestNPUNodes score node by calculate task req npu num and node npu top
func (tp *asend310) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(scoreMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", PluginName, err.Error())
		return err
	}
	if tp.handle != nil {
		return tp.handle.ScoreBestNPUNodes(task, nodes, scoreMap)
	}
	klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes handle is nil", PluginName)
	return nil
}

// UseAnnotation select npu for task from node
func (tp *asend310) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation err: %s", PluginName, err.Error())
		return nil
	}
	if tp.handle != nil {
		return tp.handle.UseAnnotation(task, node)
	}
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation handle is nil", PluginName)
	return nil
}

// ReleaseAnnotation Release used resource.
func (tp *asend310) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
