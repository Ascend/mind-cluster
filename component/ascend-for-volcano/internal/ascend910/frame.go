/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package ascend910 is using for HuaWei Ascend pin affinity schedule.
*/
package ascend910

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910/ascend910b/card910bx2"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910/ascend910b/card910bx2infer"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910/ascend910b/module910bx16"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910/ascend910b/module910bx8"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910/ascend910b/superpod"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910/asend910old/card910x2"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910/asend910old/half910x4"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910/asend910old/module910x8"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// Name This need by frame init plugin.
func (tp *ascend910) Name() string {
	return PluginName
}

// New return npu plugin.
func New(npuName string) plugin.ISchedulerPlugin {
	var npuPlugin = &ascend910{}
	npuPlugin.SetPluginName(npuName)
	npuPlugin.SetAnnoName(util.NPU910CardName)
	npuPlugin.SetAnnoPreVal(util.NPU910CardNamePre)
	npuPlugin.SetDefaultJobSchedulerConfig(nil)

	npuPlugin.Kind = map[string]base.AscendHandler{}
	npuPlugin.Kind[card910x2.SchedulerName] = card910x2.New(card910x2.SchedulerName)
	npuPlugin.Kind[module910x8.SchedulerName] = module910x8.New(module910x8.SchedulerName)
	npuPlugin.Kind[half910x4.SchedulerName] = half910x4.New(half910x4.SchedulerName)
	npuPlugin.Kind[module910bx16.SchedulerName] = module910bx16.New(module910bx16.SchedulerName)
	npuPlugin.Kind[module910bx8.SchedulerName] = module910bx8.New(module910bx8.SchedulerName)
	npuPlugin.Kind[card910bx2.SchedulerName] = card910bx2.New(card910bx2.SchedulerName)
	npuPlugin.Kind[card910bx2infer.SchedulerName] = card910bx2infer.New(card910bx2infer.SchedulerName)
	npuPlugin.Kind[superpod.SchedulerName] = superpod.New(superpod.SchedulerName)
	return npuPlugin
}

// InitMyJobPlugin init job handle plugin
func (tp *ascend910) InitMyJobPlugin(attr util.SchedulerJobAttr, env plugin.ScheduleEnv) error {
	if tp == nil {
		err := fmt.Errorf("nil plugin %s", PluginName)
		klog.V(util.LogErrorLev).Infof("InitMyJobPlugin err: %s.", err.Error())
		return err
	}
	tp.SetSchedulerAttr(attr)
	tp.SetSchedulerEnv(env)

	var handlerName string

	_, ok := attr.Annotation[superpod.SuperPodAnnoKey]
	if ok {
		handlerName = superpod.SchedulerName
	} else {
		v, ok := attr.Selector[util.AcceleratorType]
		if !ok {
			v = util.ModuleAcceleratorType
		}
		cardNameSplit := strings.Split(attr.ReqNPUName, "-")
		cardName := cardNameSplit[0]
		if attr.ReqNPUName == util.AscendNPUCore {
			cardName = util.NPU910CardName
		}
		handlerName = cardName + v
	}

	value, ok := tp.Kind[handlerName]
	if !ok {
		err := fmt.Errorf("not support %s", handlerName)
		klog.V(util.LogErrorLev).Infof("%s InitMyJobPlugin err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	if err := value.InitMyJobPlugin(attr, env); err != nil {
		klog.V(util.LogErrorLev).Infof("%s InitMyJobPlugin err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	tp.handle = value

	return nil
}

// ValidNPUJob check job req npu num and mode
func (tp *ascend910) ValidNPUJob() *api.ValidateResult {
	if tp == nil {
		err := fmt.Errorf("nil plugin %s", PluginName)
		klog.V(util.LogErrorLev).Infof("ValidNPUJob err: %s.", err.Error())
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: err.Error(),
		}
	}
	if tp.handle != nil {
		return tp.handle.ValidNPUJob()
	}
	klog.V(util.LogDebugLev).Infof("%s ValidNPUJob handle is nil", tp.GetPluginName())
	return nil
}

// CheckNodeNPUByTask check node npu meet task request
func (tp *ascend910) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err.Error())
		return err
	}
	if tp.Type != util.JobTypeWhole && tp.Type != util.JobTypeDyCut {
		klog.V(util.LogDebugLev).Infof("%s %s CheckNodeNPUByTask is %d, skip it.", tp.GetPluginName(), task.Name,
			tp.Type)
		return nil
	}
	if tp.handle != nil {
		return tp.handle.CheckNodeNPUByTask(task, node)
	}
	klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask handle is nil", tp.GetPluginName())
	return nil
}

// ScoreBestNPUNodes score nodes which meet task req
func (tp *ascend910) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(scoreMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %v.", err.Error())
		return err
	}
	if tp.handle != nil {
		return tp.handle.ScoreBestNPUNodes(task, nodes, scoreMap)
	}
	klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes handle is nil", tp.GetPluginName())
	return nil
}

// UseAnnotation select npu for task from node
func (tp *ascend910) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation %s.", err.Error())
		return nil
	}
	if tp.handle != nil {
		return tp.handle.UseAnnotation(task, node)
	}
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation handle is nil", tp.GetPluginName())
	return nil
}

// PreStartAction pre-processing actions for rescheduling
func (tp *ascend910) PreStartAction(i interface{}, ssn *framework.Session) error {
	if tp == nil || tp.handle == nil {
		return fmt.Errorf(util.ArgumentError)
	}
	if err := tp.handle.PreStartAction(i, ssn); err != nil {
		return err
	}
	return nil
}

// PreStopAction post-processing actions for re-scheduling
func (tp *ascend910) PreStopAction(env *plugin.ScheduleEnv) error {
	if tp == nil || tp.handle == nil {
		return fmt.Errorf(util.ArgumentError)
	}
	if env == nil {
		return fmt.Errorf("env is nil: %s", util.ArgumentError)
	}
	for name, handler := range tp.Kind {
		klog.V(util.LogInfoLev).Infof("preStopAction for %s", name)
		if err := handler.PreStopAction(env); err != nil {
			if strings.Contains(err.Error(), util.ArgumentError) {
				continue
			}
			klog.V(util.LogErrorLev).Infof("preStopAction %s error: %v", name, err)
		}
		if err := handler.InitMyJobPlugin(util.SchedulerJobAttr{}, plugin.ScheduleEnv{}); err != nil {
			klog.V(util.LogErrorLev).Infof("PreStartAction init plugin failed, err: %s", err)
		}
	}
	return nil
}

// ReleaseAnnotation Release used resource.
func (tp *ascend910) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
