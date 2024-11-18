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
Package base is using for HuaWei Ascend pin affinity schedule.
*/
package base

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// New return npu plugin
func New(name string) plugin.ISchedulerPlugin {
	return &NPUHandler{}
}

// InitMyJobPlugin set attr and env for plugin
func (tp *NPUHandler) InitMyJobPlugin(attr util.SchedulerJobAttr, env plugin.ScheduleEnv) error {
	if tp == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("InitMyJobPlugin %s.", err.Error())
		return err
	}
	tp.SetSchedulerAttr(attr)
	tp.SetSchedulerEnv(env)
	return nil
}

// ValidNPUJob check job req npu num
func (tp *NPUHandler) ValidNPUJob() *api.ValidateResult {
	if tp == nil {
		err := errors.New(util.ArgumentError)
		return &api.ValidateResult{Pass: false, Reason: err.Error(), Message: err.Error()}
	}
	klog.V(util.LogDebugLev).Infof("%s ValidNPUJob job(%s).", tp.GetPluginName(), tp.Name)

	for _, task := range tp.Tasks {
		taskNPU := task.ReqNPUNum

		klog.V(util.LogDebugLev).Infof("%s check task<%s> require npu<%d>.",
			tp.GetPluginName(), task.Name, taskNPU)

		if taskNPU < 1 || taskNPU > tp.MaxNodeNPUNum {
			err := fmt.Errorf("task<%s-%s> req npu num<%d> is invalid", tp.Name, task.Name, taskNPU)
			klog.V(util.LogErrorLev).Infof("%s ValidNPUJob err: %s", tp.GetPluginName(), err.Error())
			return &api.ValidateResult{
				Pass:    false,
				Reason:  "task req npu num is invalid",
				Message: err.Error(),
			}
		}
	}
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *NPUHandler) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s.", err.Error())
		return err
	}
	klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask task<%s> node<%s>.",
		tp.GetPluginName(), task.Name, node.Name)
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	nodeTop, err := tp.GetUsableTopFromNode(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	if err := tp.JudgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	return nil
}

// ScoreBestNPUNodes score node by calculate task req npu num and node npu top
func (tp *NPUHandler) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(scoreMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes err: %s.", err.Error())
		return err
	}
	return nil
}

// UseAnnotation select npu for task from node
func (tp *NPUHandler) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation err: %s.", err.Error())
		return nil
	}
	selectedNPU, err := tp.SelectNPUFromNode(task, node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation err: %s.", tp.GetPluginName(), err.Error())
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s UseAnnotation task<%s> select npu <%v>.",
		tp.GetPluginName(), task.Name, selectedNPU)

	tp.SetNPUTopologyToPodFn(task, selectedNPU, node)
	return tp.UpdateNodeInfo(node, selectedNPU)
}

// PreStartAction do something before schedule
// PreStartAction pre-processing actions for rescheduling
func (tp *NPUHandler) PreStartAction(_ interface{}, ssn *framework.Session) error {
	klog.V(util.LogInfoLev).Infof("Entering PreStartAction of %s...", PluginName)
	defer klog.V(util.LogInfoLev).Infof("Leaving PreStartAction of %s", PluginName)
	if tp == nil || ssn == nil || tp.FrameAttr.KubeClient == nil {
		return fmt.Errorf("%s handler not enabled or ssn is nil: %s", PluginName, util.ArgumentError)
	}
	tp.ReHandle = rescheduling.New(&tp.ScheduleEnv, rescheduling.CmFaultJob)
	if tp.ReHandle == nil {
		klog.V(util.LogErrorLev).Infof("create new fault handler failed.")
		return fmt.Errorf("%s reSchedule not enabled: %s", PluginName, util.ArgumentError)
	}
	tp.ReHandle.NewCommonReScheduler(rescheduling.CmFaultJob)
	tp.ReHandle.SynCacheFaultNodeWithSession()
	tp.ReHandle.AddFaultNodeWithSession()
	tp.ReHandle.InitFaultNodeMap()
	tp.ReHandle.SynCacheFaultJobWithSession(ssn)
	tp.ReHandle.SyncJobRemainRetryTimes(ssn)
	tp.ReHandle.SynCacheNodeRankOccMapWithSession(ssn)
	// 1. restart Fault Jobs that are recorded in cache
	if restartErr := tp.ReHandle.RestartNeedForceDeleteJobs(ssn, tp.ScheduleEnv); restartErr != nil &&
		restartErr.Error() != util.ArgumentError {
		klog.V(util.LogErrorLev).Infof("%s RestartNeedForceDeleteJobs: %s", PluginName, restartErr.Error())
	}
	// 2. get all the new 910x8 jobs in session
	runningJobs, getRunErr := tp.ReHandle.GetRunningJobs(ssn)
	if getRunErr != nil {
		klog.V(util.LogDebugLev).Infof("%s GetRunningJobs: %s", PluginName, getRunErr.Error())
	}
	// 3. get nodes of session and fault jobs
	if err := tp.ReHandle.AddFaultJobWithSession(runningJobs, tp.ScheduleEnv); err != nil {
		klog.V(util.LogErrorLev).Infof("%s AddFaultJobWithSession %s", PluginName, err)
	}
	// 4. restart the fault jobs
	if restartErr := tp.ReHandle.RestartFaultJobs(ssn, tp.ScheduleEnv); restartErr != nil {
		klog.V(util.LogErrorLev).Infof("%s RestartFaultJobs: %s", PluginName, restartErr.Error())
		return restartErr
	}
	// 5. save structure for later allocation process
	tp.ReHandle.GenerateNodeRankIndexTaskMap()
	return nil
}

// GetReHandle do something after schedule
func (tp *NPUHandler) GetReHandle() interface{} {
	return tp.ReHandle
}

// SetSchedulerAttr set scheduler attribute for plugin
func (tp *NPUHandler) SetSchedulerAttr(attr util.SchedulerJobAttr) {
	if tp == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("SetSchedulerAttr err: %s.", err.Error())
		return
	}
	tp.SchedulerJobAttr = attr
}

// SetSchedulerEnv set scheduler env for plugin
func (tp *NPUHandler) SetSchedulerEnv(env plugin.ScheduleEnv) {
	if tp == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("SetSchedulerEnv err: %s.", err.Error())
		return
	}
	tp.ScheduleEnv = env
}

// SetMaxNodeNPUNum set max npu num per node
func (tp *NPUHandler) SetMaxNodeNPUNum(num int) {
	if tp == nil || num < 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("SetMaxNodeNPUNum err: %s.", err.Error())
		return
	}
	tp.MaxNodeNPUNum = num
}

// SetMaxCardNPUNum set max npu num per card
func (tp *NPUHandler) SetMaxCardNPUNum(num int) {
	if tp == nil || num < 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("SetMaxCardNPUNum err: %s.", err.Error())
		return
	}
	tp.MaxCardNPUNum = num

}

// JudgeNodeAndTaskNPU judge node and task npu num
func (tp *NPUHandler) JudgeNodeAndTaskNPU(taskNPU int, nodeNPUTopology []int) error {
	if tp == nil {
		return errors.New(util.ArgumentError)
	}
	if taskNPU < 1 || taskNPU > tp.MaxNodeNPUNum {
		return fmt.Errorf("judgeNodeAndTaskNPU task req num<%d> is invalid", taskNPU)
	}

	if len(nodeNPUTopology) < taskNPU {
		return fmt.Errorf("judgeNodeAndTaskNPU node don't have enough resource, req<%d>, idle<%d>",
			taskNPU, len(nodeNPUTopology))
	}

	return nil
}

// SelectNPUFromNode select npu from node for task
func (tp *NPUHandler) SelectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		return nil, errors.New(util.ArgumentError)
	}
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("selectNPUFromNode err: %s", err.Error())
		return nil, err
	}

	nodeTop, err := tp.GetUsableTopFromNode(node)
	if err != nil {
		return nil, fmt.Errorf("selectNPUFromNode err: %s", err.Error())
	}
	if len(nodeTop) < taskNPUNum {
		return nil, fmt.Errorf("selectNPUFromNode node<%s> npu<%v> not meet task req num<%d>",
			node.Name, nodeTop, taskNPUNum)
	}
	return nodeTop[:taskNPUNum], nil
}

// PreStopAction post-processing actions for re-scheduling
func (tp *NPUHandler) PreStopAction(env *plugin.ScheduleEnv) error {
	klog.V(util.LogInfoLev).Infof("enter PreStopAction of %s...", PluginName)
	defer klog.V(util.LogInfoLev).Infof("leave PreStopAction of %s...", PluginName)
	if tp == nil || tp.ReHandle == nil || env == nil || tp.FrameAttr.KubeClient == nil {
		return fmt.Errorf("%s reSchedule not enabled or nil env: %s", PluginName, util.ArgumentError)
	}
	if err := tp.ReHandle.WriteReSchedulerCacheToEnvCache(env, rescheduling.CmFaultJob); err != nil {
		return err
	}
	return nil
}

// ReleaseAnnotation release annotation
func (tp *NPUHandler) ReleaseAnnotation(_ *api.TaskInfo, _ plugin.NPUNode) *plugin.NPUNode {
	return &plugin.NPUNode{}
}
