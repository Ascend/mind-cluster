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
Package module910bx16 is using for HuaWei Ascend910B A+X pin affinity schedule.
*/
package module910bx16

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// New return npu plugin
func New(name string) base.AscendHandler {
	m := &module910bx16{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetDefaultJobSchedulerConfig(nil)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.SetAcceleratorValue(util.JobKind910BValue)
	m.SetArch(util.HuaweiArchX86)
	m.InitVNPU()
	m.SetNpuNumInvalidMap(map[int]struct{}{util.NPUIndex9: {}, util.NPUIndex11: {}, util.NPUIndex13: {},
		util.NPUIndex15: {}})
	m.netUnhealthyKey = networkUnhealthyNPU
	m.AffScoreList = [][]int{
		{util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3, util.AffScore4, util.AffScore5,
			util.AffScore6, util.AffScore7},
		{util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3, util.AffScore4,
			util.AffScore5, util.AffScore6},
		{util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3,
			util.AffScore4, util.AffScore5},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2,
			util.AffScore3, util.AffScore4},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1,
			util.AffScore2, util.AffScore3},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0,
			util.AffScore1, util.AffScore2},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore0, util.AffScore1},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore0},
	}
	return m
}

// ValidNPUJob check job req npu num and mode
func (tp *module910bx16) ValidNPUJob() *api.ValidateResult {
	if tp.VJob.Type == util.JobTypeDyCut {
		return tp.ValidDyVNPUJob()
	}
	if err := tp.Valid910bNPUJob(); err != nil {
		return err
	}
	return tp.reHandle.ValidJobByReschedule(tp.SchedulerJobAttr)
}

// PreStartAction pre-processing actions for rescheduling
func (tp *module910bx16) PreStartAction(i interface{}, ssn *framework.Session) error {
	k, ok := i.(*rescheduling.ReScheduler)
	if !ok {
		return fmt.Errorf("PreStartAction failed %s, interface is not ReScheduler", SchedulerName)
	}
	tp.reHandle = k
	if vErr := tp.PreStartVNPU(ssn); vErr != nil {
		return fmt.Errorf("preStartVNPU failed %s, err is %s", SchedulerName, vErr)
	}
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *module910bx16) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err.Error())
		return err
	}
	switch tp.VJob.Type {
	case util.JobTypeDyCut:
		if err := tp.checkNodeNPUForDyCut(task, node); err != nil {
			return err
		}
	case util.JobTypeWhole:
		if err := tp.checkNodeNPUForWholeCard(task, node); err != nil {
			return err
		}
	default:
		return nil
	}

	if tp.reHandle != nil {
		if reErr := tp.reHandle.CheckNodeNPUByTask(task, node, tp.ReqNPUName); reErr != nil {
			return fmt.Errorf("rescheduling %s", reErr.Error())
		}
	}
	return nil
}

func (tp *module910bx16) checkNodeNPUForWholeCard(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err.Error())
		return err
	}
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	nodeTop, err := tp.getUsableTopFromNode(node, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	if err = tp.Judge910BNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("%s Judge910BNodeAndTaskNPU err: %s", tp.GetPluginName(), err.Error())
		return fmt.Errorf("npu topology not meet job require,network unhealthy card is [ %s ]",
			node.Annotation[tp.netUnhealthyKey])
	}
	return nil
}

func (tp *module910bx16) checkNodeNPUForDyCut(task *api.TaskInfo, node plugin.NPUNode) error {
	taskRes, err := tp.VHandle.GetTaskResource(task, node)
	if err != nil {
		return err
	}
	if !node.IsResourceWholeCard(taskRes.Aicore) {
		return tp.VHandle.CheckNodeNPUByDyTask(task, node, taskRes)
	}
	nodeTop := node.GetNodeTopForWholeCard()
	taskNPUNum := taskRes.Aicore / node.AiCorePerChip
	return tp.Judge910BNodeAndTaskNPU(taskNPUNum, nodeTop)
}

func (tp *module910bx16) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, sMap map[string]float64) error {
	if tp == nil || task == nil || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err)
		return err
	}
	if tp.VJob.Type == util.JobTypeDyCut {
		return tp.VHandle.DynamicVNPU.ScoreBestNPUNodes(task, nodes, sMap)
	}
	taskNPUNum, getErr := tp.GetTaskReqNPUNum(task)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum %s: %s", tp.GetPluginName(), task.Name, getErr)
		return getErr
	}
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}
		nNode, ok := tp.Nodes[node.Name]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s %s ScoreBestNPUNodes %s is not npu node",
				tp.GetPluginName(), task.Name, node.Name)
			continue
		}
		cardIds, err := tp.getUsableTopFromNode(nNode, tp.NPUTaskNum > 1)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
			continue
		}
		bestScore, err := tp.GetNodeBestScore(taskNPUNum, cardIds)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
			continue
		}
		healthyNPUNum, ok := nNode.Allocate[v1.ResourceName(tp.GetAnnoName())]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes node<%s> get allocate npu failed",
				tp.GetPluginName(), node.Name)
			continue
		}
		sortScore := tp.MaxNodeNPUNum - len(cardIds)
		sMap[node.Name] = float64(tp.MaxNodeNPUNum*(int(healthyNPUNum/util.NPUHexKilo)-bestScore) + sortScore)
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes task<%s> sMap<%v>", tp.GetPluginName(),
		task.Name, sMap)
	return tp.reHandle.ScoreBestNPUNodes(task, sMap)
}

// UseAnnotation select npu for task from node
func (tp *module910bx16) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation %s.", err)
		return nil
	}
	if tp.VJob.Type == util.JobTypeDyCut {
		return tp.useAnnotationForDyCut(task, node)
	}
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation task<%s> node<%s> resource<%s> Annotation: %s",
		tp.GetPluginName(), task.Name, node.Name, tp.GetAnnoName(), util.SafePrint(node.Annotation))
	selectedNPU, err := tp.selectNPUFromNode(task, node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation err:%s.", tp.GetPluginName(), err)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s UseAnnotation %s select %v.", tp.GetPluginName(), task.Name, selectedNPU)

	tp.SetNPUTopologyToPodFn(task, selectedNPU, node)
	newNode := tp.UpdateNodeInfo(node, selectedNPU)
	return newNode
}

func (tp *module910bx16) useAnnotationForDyCut(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	taskRes, err := tp.VHandle.GetTaskResource(task, node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation job(%s) get require task resource failed: %s",
			tp.GetPluginName(), tp.Name, err)
		return &node
	}
	if !node.IsResourceWholeCard(taskRes.Aicore) {
		return tp.VHandle.DynamicVNPU.UseAnnotation(task, node, taskRes, tp.VHandle.VT)
	}
	nodeTop := node.GetNodeTopForWholeCard()
	taskNPUNum := taskRes.Aicore / node.AiCorePerChip
	selectNpu, err := tp.selectNPUByTaskNPUNumAndNodeTop(taskNPUNum, nodeTop)
	if err != nil {
		return nil
	}
	allocChipID := strings.Join(changeIntSliceToString(selectNpu), ",")
	tp.VHandle.SetNPUTopologyToPodFn(task, node, taskRes, allocChipID, tp.VHandle.VT)
	return tp.VHandle.UpdateNodeInfo(node, allocChipID, taskRes)

}

func (tp *module910bx16) selectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	nodeTop, err := tp.getUsableTopFromNode(node, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	return tp.selectNPUByTaskNPUNumAndNodeTop(taskNPUNum, nodeTop)
}

func (tp *module910bx16) selectNPUByTaskNPUNumAndNodeTop(taskNPUNum int, nodeTop []int) ([]int, error) {
	if taskNPUNum == tp.MaxNodeNPUNum {
		if len(nodeTop) == tp.MaxNodeNPUNum {
			return nodeTop, nil
		}
		err := fmt.Errorf("node top<%v> can not meet task req<%d>", nodeTop, taskNPUNum)
		klog.V(util.LogErrorLev).Infof("%s selectNPUFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	priorityArray, err := tp.GetNPUAllocPriorityArray(taskNPUNum)
	if err != nil {
		klog.V(util.LogErrorLev).Info(err.Error())
		return nil, err
	}
	klog.V(util.LogInfoLev).Infof("selectNPUFromNode %s[%d] priority:%v in %v.",
		tp.GetPluginName(), taskNPUNum, priorityArray, nodeTop)

	leftHccsArray, rightHccsArray, samePlaceHccsArray := tp.GetNodeHccsArray(nodeTop, tp.NPUTaskNum > 1)
	for _, priority := range priorityArray {
		if priority == len(leftHccsArray) {
			return leftHccsArray[:taskNPUNum], nil
		}
		if priority == len(rightHccsArray) {
			return rightHccsArray[:taskNPUNum], nil
		}
		if priority == len(samePlaceHccsArray) {
			return samePlaceHccsArray[:taskNPUNum], nil
		}
	}
	err = fmt.Errorf("node top<%v> can not meet task req<%d>", len(nodeTop), taskNPUNum)
	klog.V(util.LogErrorLev).Infof("%s selectNPUFromNode err: %s", tp.GetPluginName(), err.Error())
	return nil, err
}

// ReleaseAnnotation Release used resource.
func (tp *module910bx16) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}

func (tp *module910bx16) GetNPUAllocPriorityArray(taskNPUNumber int) ([]int, error) {

	var err error
	if !tp.CheckJobAllowNum(taskNPUNumber) {
		err = fmt.Errorf("illegal request npu number: %d", taskNPUNumber)
		klog.V(util.LogErrorLev).Infof("%s %s.", tp.GetPluginName(), err)
		return nil, err
	}
	var priorityArray []int
	if taskNPUNumber == tp.MaxNodeNPUNum {
		return []int{tp.MaxNodeNPUNum}, nil
	}
	if taskNPUNumber <= tp.MaxNodeNPUNum/util.NPUIndex2 {
		for i := taskNPUNumber; i <= tp.MaxNodeNPUNum/util.NPUIndex2; i++ {
			priorityArray = append(priorityArray, i)
		}
		return priorityArray, nil
	}
	if taskNPUNumber > tp.MaxNodeNPUNum/util.NPUIndex2 {
		for i := taskNPUNumber; i <= tp.MaxNodeNPUNum; i = i + util.NPUIndex2 {
			priorityArray = append(priorityArray, i)
		}
		return priorityArray, nil
	}
	return priorityArray, nil
}

func (tp *module910bx16) GetNodeHccsArray(nodeTop []int, isMultNpuReplica bool) ([]int, []int, []int) {
	var leftHccsArray []int
	var rightHccsArray []int

	idCutNum := tp.MaxNodeNPUNum / util.NPUIndex2
	for _, v := range nodeTop {
		if v < idCutNum {
			leftHccsArray = append(leftHccsArray, v)
			continue
		}
		rightHccsArray = append(rightHccsArray, v)
	}
	crossHccsArray := getCrossHccsArray(leftHccsArray, rightHccsArray, isMultNpuReplica, idCutNum)
	return leftHccsArray, rightHccsArray, crossHccsArray
}

func getCrossHccsArray(leftHccsArray, rightHccsArray []int, isMultNpuReplica bool, idCutNum int) []int {
	var crossHccsArray []int
	if isMultNpuReplica {
		minLen := len(leftHccsArray)
		if minLen > len(rightHccsArray) {
			minLen = len(rightHccsArray)
		}
		for i := 0; i < minLen; i++ {
			crossHccsArray = append(crossHccsArray, leftHccsArray[i], rightHccsArray[i])
		}
	} else {
		for _, leftCardID := range leftHccsArray {
			for _, rightCardID := range rightHccsArray {
				if leftCardID+idCutNum == rightCardID {
					crossHccsArray = append(crossHccsArray, leftCardID, rightCardID)
					break
				}
			}
		}
	}

	// npu num must bigger than hccs's npu number, if task is cross hccs
	if len(crossHccsArray) <= idCutNum {
		crossHccsArray = []int{}
	}
	return crossHccsArray
}
