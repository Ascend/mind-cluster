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
	"errors"
	"fmt"
	"reflect"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// SetAcceleratorValue Set the acceleratorValue to distinguish between task types.
func (ab *Base910b) SetAcceleratorValue(value string) {
	ab.acceleratorValue = value
}

// GetAcceleratorValue Get the acceleratorValue to distinguish between task types.
func (ab *Base910b) GetAcceleratorValue() string {
	return ab.acceleratorValue
}

// SetArch Set the job arch to distinguish between jobs. A+X 16P,A+K 8p.
func (ab *Base910b) SetArch(value string) {
	ab.arch = value
}

// GetArch Get the job arch to distinguish between jobs. A+X 16P,A+K 8p.
func (ab *Base910b) GetArch() string {
	return ab.arch
}

// SetNpuNumInvalidMap  Set the single job not allow number. eg:A+X 16P:9,10,11,12,13,14,15
func (ab *Base910b) SetNpuNumInvalidMap(value map[int]struct{}) {
	ab.NpuNumInvalidMap = value
}

// CheckJobAllowNum check the single job require is valid. eg:A+X 16P:1,2,4,8,16;A+K 1,2,4,8.
func (ab *Base910b) CheckJobAllowNum(value int) bool {
	_, ok := ab.NpuNumInvalidMap[value]
	return !ok && value <= ab.MaxNodeNPUNum
}

func (ab *Base910b) initSelectNodeInf(npuTop []int) SelectNodeInf {
	var sNodeInf SelectNodeInf
	var leftHccsTop []int
	var rightHccsTop []int

	numHCCS := ab.MaxNodeNPUNum / util.NPUIndex2
	for _, cardID := range npuTop {
		if cardID < numHCCS {
			leftHccsTop = append(leftHccsTop, cardID)
		} else {
			rightHccsTop = append(rightHccsTop, cardID)
		}
	}

	sNodeInf.LeftNPUNum = len(leftHccsTop)
	sNodeInf.RightNPUNum = len(rightHccsTop)
	sNodeInf.AllNPUNum = sNodeInf.LeftNPUNum + sNodeInf.RightNPUNum

	if ab.NPUTaskNum > 1 {
		minLen := len(leftHccsTop)
		if minLen > len(rightHccsTop) {
			minLen = len(rightHccsTop)
		}
		sNodeInf.crossNPUNum = minLen * util.NPUIndex2
		return sNodeInf
	}
	for _, leftCardID := range leftHccsTop {
		for _, rightCardID := range rightHccsTop {
			if leftCardID+numHCCS == rightCardID {
				sNodeInf.crossNPUNum = sNodeInf.crossNPUNum + util.NPUIndex2
				break
			}
		}
	}
	return sNodeInf
}

// Judge910BNodeAndTaskNPU Judge 910BNode  wither meet npu task not.
func (ab *Base910b) Judge910BNodeAndTaskNPU(taskNPU int, nodeTop []int) error {
	dealReturnValue := func(value bool) error {
		if value {
			return nil
		}
		meetErr := fmt.Errorf("%v not meet req npu(%d)", nodeTop, taskNPU)
		klog.V(util.LogErrorLev).Infof("%s %v not meet task req:%d.", ab.GetPluginName(), nodeTop, taskNPU)
		return meetErr
	}

	sNodeInf := ab.initSelectNodeInf(nodeTop)
	if taskNPU == ab.MaxNodeNPUNum {
		return dealReturnValue(sNodeInf.AllNPUNum == ab.MaxNodeNPUNum)
	}

	if ab.CheckJobAllowNum(taskNPU) {
		return dealReturnValue((sNodeInf.LeftNPUNum >= taskNPU) || (sNodeInf.RightNPUNum >= taskNPU) ||
			(taskNPU > ab.MaxNodeNPUNum/util.NPUIndex2 && taskNPU <= sNodeInf.crossNPUNum))
	}
	return dealReturnValue(false)
}

// GetNodeBestScore Get node core
func (ab *Base910b) GetNodeBestScore(taskNPUNum int, npuTop []int) (int, error) {
	var bestScore = len(ab.AffScoreList)
	sNodeInf := ab.initSelectNodeInf(npuTop)
	if sNodeInf.AllNPUNum < 1 ||
		sNodeInf.AllNPUNum > ab.MaxNodeNPUNum {
		return 0, fmt.Errorf("node top %v is invalid for %v", npuTop, sNodeInf)
	}

	var err = fmt.Errorf("node %v is not meet task req %d", npuTop, taskNPUNum)
	if taskNPUNum == ab.MaxNodeNPUNum {
		if len(npuTop) == ab.MaxNodeNPUNum {
			return 0, nil
		}
		return 0, err
	}

	switch {
	case taskNPUNum > ab.MaxNodeNPUNum/util.NPUIndex2:
		bestScore = ab.AffScoreList[(taskNPUNum/util.NPUIndex2)-1][(sNodeInf.crossNPUNum/util.NPUIndex2)-1]
	case sNodeInf.RightNPUNum == 0:
		bestScore = ab.AffScoreList[taskNPUNum-1][sNodeInf.LeftNPUNum-1]
	case sNodeInf.LeftNPUNum == 0:
		bestScore = ab.AffScoreList[taskNPUNum-1][sNodeInf.RightNPUNum-1]
	default:
		bestScore = util.Min(ab.AffScoreList[taskNPUNum-1][sNodeInf.RightNPUNum-1],
			ab.AffScoreList[taskNPUNum-1][sNodeInf.LeftNPUNum-1])
	}
	if bestScore == len(ab.AffScoreList) {
		return 0, err
	}
	return bestScore, nil
}

// ScoreAscendNPUNodes core ascend910B node by calculate task req npu num and node npu top
func (ab *Base910b) ScoreAscendNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, sMap map[string]float64) error {
	if ab == nil || task == nil || len(nodes) == 0 || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err)
		return err
	}
	taskNPUNum, getErr := ab.GetTaskReqNPUNum(task)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum %s: %s", ab.GetPluginName(), task.Name, getErr)
		return getErr
	}
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}
		nNode, ok := ab.Nodes[node.Name]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s %s ScoreBestNPUNodes %s is not npu node",
				ab.GetPluginName(), task.Name, node.Name)
			continue
		}
		cardIds, err := ab.GetUsableTopFromNode(nNode)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", ab.GetPluginName(), err)
			continue
		}
		bestScore, err := ab.GetNodeBestScore(taskNPUNum, cardIds)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", ab.GetPluginName(), err)
			continue
		}
		healthyNPUNum, ok := nNode.Allocate[v1.ResourceName(ab.GetAnnoName())]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes node<%s> get allocate npu failed",
				ab.GetPluginName(), node.Name)
			continue
		}
		sortScore := ab.MaxNodeNPUNum - len(cardIds)
		sMap[node.Name] = float64(ab.MaxNodeNPUNum*(int(healthyNPUNum/util.NPUHexKilo)-bestScore) + sortScore)
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes task<%s> sMap<%v>", ab.GetPluginName(),
		task.Name, sMap)
	return nil
}

// Use910bAnnotation select npu for 910b task from node
func (ab *Base910b) Use910bAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if ab == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation %s.", err)
		return nil
	}
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation task<%s> node<%s> resource<%s> Annotation: %s",
		ab.GetPluginName(), task.Name, node.Name, ab.GetAnnoName(), util.SafePrint(node.Annotation))
	selectedNPU, err := ab.selectNPUFromNode(task, node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation err:%s.", ab.GetPluginName(), err)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s UseAnnotation %s select %v.", ab.GetPluginName(), task.Name, selectedNPU)

	ab.SetNPUTopologyToPodFn(task, selectedNPU)
	newNode := ab.UpdateNodeInfo(node, selectedNPU)
	return newNode
}

// Check910bNodeNPUByTask check nod npu meet 910B task req
func (ab *Base910b) Check910bNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if ab == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}

	taskNPUNum, err := ab.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum %s: %s", ab.GetPluginName(), task.Name, err)
		return err
	}

	nodeTop, err := ab.GetUsableTopFromNode(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetUsableTopFromNode %s: %s", ab.GetPluginName(), task.Name, err)
		return err
	}

	if err = ab.Judge910BNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("%s Judge910BNodeAndTaskNPU %s: %s", ab.GetPluginName(), task.Name, err)
		return fmt.Errorf("checkNodeNPUByTask %s err: %s", util.NodeNotMeetTopologyWarning, err)
	}

	return nil
}
