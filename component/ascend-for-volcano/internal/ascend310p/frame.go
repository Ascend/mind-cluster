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
Package ascend310p is using for HuaWei 310P Ascend pin affinity schedule.
*/
package ascend310p

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend310p/card310px2"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend310p/chip310px2"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// New return npu plugin.
func New(npuName string) plugin.ISchedulerPlugin {
	npuPlugin := &ascend310P{}
	npuPlugin.SetPluginName(npuName)
	npuPlugin.SetAnnoName(util.NPU310PCardName)
	npuPlugin.SetAnnoPreVal(util.NPU310PCardNamePre)
	npuPlugin.SetDefaultJobSchedulerConfig(nil)
	npuPlugin.SetMaxNodeNPUNum(maxNodeNPUNum)
	npuPlugin.InitVNPU()

	npuPlugin.Kind = map[string]base.AscendHandler{}
	npuPlugin.Kind[chip310px2.SchedulerName] = chip310px2.New(chip310px2.SchedulerName)
	npuPlugin.Kind[card310px2.SchedulerName] = card310px2.New(card310px2.SchedulerName)

	return npuPlugin
}

// InitMyJobPlugin for 300I duo job init
func (tp *ascend310P) InitMyJobPlugin(attr util.SchedulerJobAttr, env plugin.ScheduleEnv) error {
	if tp == nil {
		mgs := fmt.Errorf("nil plugin %s", PluginName)
		klog.V(util.LogInfoLev).Infof("InitMyJobPlugin %v.", mgs)
		return mgs
	}
	tp.SetSchedulerAttr(attr)
	tp.SetSchedulerEnv(env)
	klog.V(util.LogDebugLev).Infof("InitMyJobPlugin attr label %#v.", attr.Label)
	duo, ok := attr.Label[DuoKeyLabel]
	if !ok {
		klog.V(util.LogInfoLev).Info("not 300I duo")
		return nil
	}
	v, ok := attr.Label[Accelerator310Key]
	if !ok {
		v = Chip310AcceleratorValue
	}
	if duo == TrueStr {
		duo = DuoKeyLabel
	}
	value, ok := tp.Kind[attr.ReqNPUName+duo+v]
	if !ok {
		return fmt.Errorf("not support %s", attr.ReqNPUName+duo+v)
	}
	if err := value.InitMyJobPlugin(attr, env); err != nil {
		return err
	}
	tp.handle = value
	return nil
}

// PreStartAction pre-processing actions for rescheduling
func (tp *ascend310P) PreStartAction(i interface{}, ssn *framework.Session) error {
	klog.V(util.LogDebugLev).Infof("Entering PreStartAction of %s", util.NPU310PCardName)
	defer klog.V(util.LogDebugLev).Infof("Leaving PreStartAction of %s", util.NPU310PCardName)
	if tp == nil || ssn == nil || tp.FrameAttr.KubeClient == nil {
		return fmt.Errorf("%s handler not enabled or ssn is nil: %s", util.NPU310PCardName, util.ArgumentError)
	}

	reErr := tp.preStartRescheduling(i)
	vErr := tp.preStartVNPU(ssn)
	if reErr == nil && vErr == nil {
		return nil
	}

	return fmt.Errorf("%s %s", util.SafePrint(reErr), util.SafePrint(vErr))
}

// ValidNPUJob check job req npu num and mode
func (tp *ascend310P) ValidNPUJob() *api.ValidateResult {
	if tp == nil {
		err := errors.New(util.ArgumentError)
		return &api.ValidateResult{Pass: false, Reason: err.Error(), Message: err.Error()}
	}
	klog.V(util.LogDebugLev).Infof("%s ValidNPUJob job(%s).", tp.GetPluginName(), tp.Name)
	if tp.VJob == nil {
		// this is the old whole card.
		return tp.NPUHandler.ValidNPUJob()
	}
	var err error = nil
	switch tp.Type {
	case util.JobTypeWhole:
		if tp.SchedulerJobAttr.Label[DuoKeyLabel] == TrueStr {
			return tp.handle.ValidNPUJob()
		}
		if validErr := tp.NPUHandler.ValidNPUJob(); validErr != nil {
			return validErr
		}
		return tp.reHandle.ValidJobByReschedule(tp.SchedulerJobAttr)
	case util.JobTypeStCut:
		return tp.validStVNPUJob()
	case util.JobTypeDyCut:
		return tp.validDyVNPUJob()
	default:
		err = fmt.Errorf("%s no type %d", tp.Name, tp.Type)
		klog.V(util.LogDebugLev).Infof("%s ValidNPUJob %s %s.", tp.GetPluginName(), tp.Name, err)
	}

	return &api.ValidateResult{Pass: false, Reason: err.Error(), Message: err.Error()}
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *ascend310P) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask job(%s).", tp.GetPluginName(), tp.Name)
	if task == nil || len(node.Annotation) == 0 {
		return errors.New(util.ArgumentError)
	}

	var err error
	if tp.VJob == nil {
		// this is the old whole card.
		if err = tp.NPUHandler.CheckNodeNPUByTask(task, node); err != nil {
			return err
		}
		return nil
	}
	nJob, jobOK := tp.NPUHandler.Jobs[task.Job]
	if !jobOK {
		err = fmt.Errorf("%s not in jobs", task.Job)
		klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask %s.", tp.GetPluginName(), err)
		return err
	}
	tpTask, ok := nJob.NPUJob.Tasks[task.UID]
	if !ok {
		err = fmt.Errorf("%s not in tasks", task.Name)
		klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask %s.", tp.GetPluginName(), err)
		return err
	}
	err = tp.checkByJobType(task, node, tpTask, err)
	if err != nil {
		return err
	}

	if reErr := tp.reHandle.CheckNodeNPUByTask(task, node, tp.ReqNPUName); reErr != nil {
		return fmt.Errorf("rescheduling CheckNodeNPUByTask %s", reErr.Error())
	}
	return nil
}

func (tp *ascend310P) checkByJobType(task *api.TaskInfo, node plugin.NPUNode,
	tpTask util.NPUTask, err error) error {
	switch tpTask.VTask.Type {
	case util.JobTypeWhole:
		if tp.SchedulerJobAttr.Label[DuoKeyLabel] == TrueStr {
			if err := tp.handle.CheckNodeNPUByTask(task, node); err != nil {
				return err
			}
		} else {
			if err = tp.NPUHandler.CheckNodeNPUByTask(task, node); err != nil {
				return err
			}
		}
	case util.JobTypeStCut:
		if err = tp.vHandle.StaticVNPU.CheckNodeNPUByTask(task, node, util.VResource{}); err != nil {
			return err
		}
	case util.JobTypeDyCut:
		taskRes, err := tp.vHandle.GetTaskResource(task, node)
		if err != nil {
			return err
		}
		if err = tp.vHandle.CheckNodeNPUByDyTask(task, node, taskRes); err != nil {
			return err
		}
	default:
		err = fmt.Errorf("%s no type %d", tp.Name, tp.Type)
		klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask %s %s.", tp.GetPluginName(), tp.Name, err)
		return err
	}
	return nil
}

// ScoreBestNPUNodes score node by calculate task req npu num and node npu top
func (tp *ascend310P) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes job(%s).", tp.GetPluginName(), tp.Name)
	if tp.VJob == nil {
		// this is the old whole card.
		if err := tp.NPUHandler.ScoreBestNPUNodes(task, nodes, scoreMap); err != nil {
			return err
		}
		return nil
	}

	err, done := tp.scoreByJobType(task, nodes, scoreMap)
	if done {
		return err
	}

	if reErr := tp.reHandle.ScoreBestNPUNodes(task, scoreMap); reErr != nil {
		klog.V(util.LogErrorLev).Infof("%s rescheduling ScoreBestNPUNodes failed :%s.",
			tp.GetPluginName(), reErr.Error())
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes task<%s> scoreMap<%v>", tp.GetPluginName(),
		task.Name, scoreMap)
	return nil
}

func (tp *ascend310P) scoreByJobType(task *api.TaskInfo, nodes []*api.NodeInfo,
	scoreMap map[string]float64) (error, bool) {
	switch tp.Type {
	case util.JobTypeWhole:
		if tp.SchedulerJobAttr.Label[DuoKeyLabel] == TrueStr {
			if err := tp.handle.ScoreBestNPUNodes(task, nodes, scoreMap); err != nil {
				return err, true
			}
		} else {
			if err := tp.NPUHandler.ScoreBestNPUNodes(task, nodes, scoreMap); err != nil {
				return err, true
			}
		}
	case util.JobTypeStCut:
		if err := tp.vHandle.StaticVNPU.ScoreBestNPUNodes(task, nodes, scoreMap); err != nil {
			return err, true
		}
	case util.JobTypeDyCut:
		if err := tp.vHandle.DynamicVNPU.ScoreBestNPUNodes(task, nodes, scoreMap); err != nil {
			return err, true
		}
	default:
		err := fmt.Errorf("%s no type %d", tp.Name, tp.Type)
		klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes %s %s.", tp.GetPluginName(), tp.Name, err)
		return err, true
	}
	return nil, false
}

// UseAnnotation select npu for task from node
func (tp *ascend310P) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation job(%s).", tp.GetPluginName(), tp.Name)

	if tp.VJob == nil {
		// this is the old whole card.
		return tp.NPUHandler.UseAnnotation(task, node)
	}
	nJob, jobOK := tp.NPUHandler.Jobs[task.Job]
	if !jobOK {
		klog.V(util.LogDebugLev).Infof("%s UseAnnotation %s not exist in jobs.", tp.GetPluginName(), task.Job)
		return &node
	}
	tpTask, taskOk := nJob.NPUJob.Tasks[task.UID]
	if !taskOk {
		klog.V(util.LogDebugLev).Infof("%s UseAnnotation %s not npu tasks.", tp.GetPluginName(), task.Name)
		return &node
	}
	switch tpTask.VTask.Type {
	case util.JobTypeWhole:
		if tp.SchedulerJobAttr.Label[DuoKeyLabel] == TrueStr {
			return tp.handle.UseAnnotation(task, node)
		}
		return tp.NPUHandler.UseAnnotation(task, node)
	case util.JobTypeStCut:
		return tp.vHandle.StaticVNPU.UseAnnotation(task, node, util.VResource{}, tp.vHandle.VT)
	case util.JobTypeDyCut:
		taskRes, err := tp.vHandle.GetTaskResource(task, node)
		klog.V(util.LogDebugLev).Infof("task<%s> require resource<%#v>", task.Name, taskRes)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("%s UseAnnotation job(%s) get require task resource failed: %s",
				tp.GetPluginName(), tp.Name, err)
		}
		return tp.vHandle.DynamicVNPU.UseAnnotation(task, node, taskRes, tp.vHandle.VT)
	default:
		err := fmt.Errorf("%s no type %d", tp.Name, tp.Type)
		klog.V(util.LogDebugLev).Infof("%s UseAnnotation %s %s.", tp.GetPluginName(), tp.Name, err)
	}

	return &node
}

// ReleaseAnnotation release select npu for task to node
func (tp *ascend310P) ReleaseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation job(%s).", tp.GetPluginName(), tp.Name)
	if tp.VJob == nil {
		// this is the old whole card.
		return &node
	}

	var err error
	switch tp.Type {
	case util.JobTypeWhole, util.JobTypeStCut:
		return &node
	case util.JobTypeDyCut:
		return tp.vHandle.DynamicVNPU.ReleaseAnnotation(task, node)
	default:
		err = fmt.Errorf("%s no type %d", tp.Name, tp.Type)
		klog.V(util.LogDebugLev).Infof("%s ReleaseAnnotation %s %s.", tp.GetPluginName(), tp.Name, err)
	}

	return &node
}
