/*
Copyright(C)2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package ascend910b is using for HuaWei Ascend 910B pin affinity schedule.
*/
package ascend910b

import (
	"fmt"
	"strings"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// Valid910bNPUJob check the 910b job req npu num and mode
func (ab *Base910b) Valid910bNPUJob() *api.ValidateResult {
	vResult := &api.ValidateResult{}
	var vErr error
	defer func() {
		if vErr != nil {
			vResult.Pass = false
			vResult.Reason = vErr.Error()
			vResult.Message = vErr.Error()
			return
		}
	}()

	// 1. check parameter.
	if ab == nil {
		vErr = fmt.Errorf("nil plugin %s", ab.GetPluginName())
		klog.V(util.LogErrorLev).Infof("ValidNPUJob err: %s.", vErr)
		return vResult
	}

	// 2.check ring-controller.atlas
	if vErr = ab.CheckJobForm(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobForm: %s.", vErr)
		return vResult
	}

	// 3.check host-arch for A+X not same as A+K
	if vErr = ab.CheckJobArch(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobArch: %s.", vErr)
		return vResult
	}

	// 4.check job train mode:distribute and single.
	if vErr = ab.checkJobTrainMode(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobTrainMode: %s.", vErr)
		return vResult
	}

	return nil
}

// CheckJobForm to check job ring-controller.atlas for future unification.
func (ab *Base910b) CheckJobForm() error {
	// for vcJob and deployment.
	lValue, ok := ab.Label[util.JobKindKey]
	if !ok {
		return fmt.Errorf("%s not has no label:%s", ab.Name, util.JobKindKey)
	}

	if lValue != ab.GetAcceleratorValue() {
		return fmt.Errorf("%s label:%s not right(%s)", ab.Name, lValue, ab.GetAcceleratorValue())
	}
	return nil
}

// CheckJobArch to check job host-arch for A+X not same as A+K.
func (ab *Base910b) CheckJobArch() error {
	// for vcJob and deployment.
	lValue, ok := ab.Selector[util.ArchSelector]
	if !ok {
		return fmt.Errorf("%s not has no selector:%s", ab.Name, util.ArchSelector)
	}

	if !strings.Contains(ab.GetArch(), lValue) {
		return fmt.Errorf("%s selector:%s not right %s", ab.Name, util.ArchSelector, lValue)
	}
	return nil
}

// checkJobTrainMode to check job train mode:distribute and single.
func (ab *Base910b) checkJobTrainMode() error {
	if ab.NPUTaskNum == 0 {
		klog.V(util.LogErrorLev).Infof("GetVTaskNumInVJob %s has no npu tasks.", ab.Name)
		return fmt.Errorf("%s no npu job", ab.Name)
	}
	klog.V(util.LogDebugLev).Infof("checkJobTrainMode job(%s) has %d tasks.", ab.Name, len(ab.Tasks))
	nTaskReqNpuNum := ab.ReqNPUNum / ab.NPUTaskNum
	if ab.CheckJobAllowNum(nTaskReqNpuNum) {
		return nil
	}
	return fmt.Errorf("%s checkJobTrainMode %s req npu is invalid", ab.GetPluginName(), ab.Name)
}

func (ab *Base910b) GetNPUAllocPriorityArray(taskNPUNumber int) ([]int, error) {
	var priorityArray []int
	var err error
	if !ab.CheckJobAllowNum(taskNPUNumber) {
		err = fmt.Errorf("illegal request npu number: %d", taskNPUNumber)
		klog.V(util.LogErrorLev).Infof("%s %s.", ab.GetPluginName(), err)
		return nil, err
	}

	for i := taskNPUNumber; i <= ab.MaxNodeNPUNum/util.NPUIndex2; i++ {
		priorityArray = append(priorityArray, i)
	}

	if ab.MaxNodeNPUNum < util.NPUIndex8 {
		priorityArray = []int{}
		for i := taskNPUNumber; i <= len(ab.AffScoreList); i++ {
			priorityArray = append(priorityArray, i)
		}
	}

	if taskNPUNumber == ab.MaxNodeNPUNum {
		priorityArray = []int{ab.MaxNodeNPUNum}
	}
	return priorityArray, nil
}

func (ab *Base910b) selectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	taskNPUNum, err := ab.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", ab.GetPluginName(), err)
		return nil, err
	}
	nodeTop, err := ab.GetUsableTopFromNode(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", ab.GetPluginName(), err)
		return nil, err
	}
	if taskNPUNum == ab.MaxNodeNPUNum {
		if len(nodeTop) == ab.MaxNodeNPUNum {
			return nodeTop, nil
		}
		err = fmt.Errorf("%s %v can not meet task req:%d", node.Name, nodeTop, taskNPUNum)
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", ab.GetPluginName(), err)
		return nil, err
	}
	priorityArray, err := ab.GetNPUAllocPriorityArray(taskNPUNum)
	if err != nil {
		klog.V(util.LogErrorLev).Info(err)
		return nil, err
	}
	klog.V(util.LogInfoLev).Infof("%s selectNPUFromNode %s[%d] priority:%v in %v.", ab.GetPluginName(),
		task.Name, taskNPUNum, priorityArray, nodeTop)

	leftHCCSArray, rightHCCSArray := ab.GetNodeHccsArray(nodeTop)
	for _, priority := range priorityArray {
		if priority == len(leftHCCSArray) && len(leftHCCSArray) != 0 {
			return leftHCCSArray[:taskNPUNum], nil
		}
		if priority == len(rightHCCSArray) && len(rightHCCSArray) != 0 {
			return rightHCCSArray[:taskNPUNum], nil
		}
	}
	err = fmt.Errorf("node<%s> top<%v> can not meet task req<%d>", node.Name, len(nodeTop), taskNPUNum)
	klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", ab.GetPluginName(), err)
	return nil, err
}
