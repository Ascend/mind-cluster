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
	"sort"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// Option the func for AscendHandler add attr
type Option func(AscendHandler)

// WithNpuInvalidMap build AscendHandler with NpuInvalidMap
func WithNpuInvalidMap(m map[int]struct{}) Option {
	return func(h AscendHandler) {
		h.SetNpuNumInvalidMap(m)
	}
}

// WithMaxNodeNum build AscendHandler WithMaxNodeNum
func WithMaxNodeNum(num int) Option {
	return func(h AscendHandler) {
		h.SetMaxNodeNPUNum(num)
	}
}

// WithAnnoPreVal build AscendHandler WithAnnoPreVal
func WithAnnoPreVal(annoPre string) Option {
	return func(h AscendHandler) {
		h.SetAnnoPreVal(annoPre)
	}
}

// WithAnnoName build AscendHandler WithAnnoName
func WithAnnoName(annoName string) Option {
	return func(h AscendHandler) {
		h.SetAnnoName(annoName)
	}
}

// WithNetworkFault build AscendHandler WithNetworkFault
func WithNetworkFault(enable bool) Option {
	return func(h AscendHandler) {
		h.SetIsNetworkFaultAttention(enable)
	}
}

// WithMaxCardNum build AscendHandler WithMaxCardNum
func WithMaxCardNum(num int) Option {
	return func(h AscendHandler) {
		h.SetMaxCardNPUNum(num)
	}
}

// New return npu plugin
func New(name string, opts ...Option) AscendHandler {
	m := &NPUHandler{}
	m.SetPluginName(name)
	m.SetAnnoName(name)
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// PreStartAction pre-processing actions for rescheduling
func (tp *NPUHandler) PreStartAction(ssn *framework.Session) error {
	return nil
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
	helper := util.NewTaskValidateHelper()
	for _, task := range tp.Tasks {
		if !task.IsNPUTask() || helper.HasTask(task.TaskSpecKey) {
			continue
		}
		taskNPU := task.ReqNPUNum
		if taskNPU < 1 || taskNPU > tp.MaxNodeNPUNum || !tp.IsVaildNpuNum(taskNPU) {
			klog.V(util.LogDebugLev).Infof("%s ValidNPUJob err: job<%s>-task<%s> req npu num<%d> is invalid",
				tp.GetPluginName(), tp.Name, task.Name, taskNPU)
			helper.AddInvalidResourceRequest(task.TaskSpecKey, taskNPU)
		}
	}
	return helper.TaskValidResult(fmt.Sprintf("job<%s> req npu num should small than %d, and not in %v", tp.Name,
		tp.MaxNodeNPUNum, tp.NpuNumInvalidMap))
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
		klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	nodeTop, err := tp.GetUsableTopFromNode(node, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("task %s CheckNodeNPUByTask err: %s", task.Name, err.Error())
		return err
	}
	if err := tp.JudgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogDebugLev).Infof("task %s CheckNodeNPUByTask err: %s", task.Name, err.Error())
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
	for _, node := range nodes {
		nNode, ok := tp.Nodes[node.Name]
		if !ok {
			continue
		}
		nodeTop, err := tp.GetUsableTopFromNode(nNode, tp.NPUTaskNum > 1)
		if err != nil {
			klog.V(util.LogDebugLev).Infof("task %s ScoreBestNPUNodes err: %s", task.Name, err.Error())
			continue
		}
		if len(nodeTop) > tp.MaxNodeNPUNum {
			continue
		}
		unhealthyNPUNum := tp.getUnhealthyNPU(nNode)
		healthyCardsNum := tp.MaxNodeNPUNum - len(unhealthyNPUNum)
		scoreMap[node.Name] = float64(healthyCardsNum*nodeWeight - len(nodeTop))
	}
	return nil
}

// UseAnnotation select npu for task from node
func (tp *NPUHandler) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogDebugLev).Infof("UseAnnotation err: %s.", err.Error())
		return nil
	}
	selectedNPU, err := tp.SelectNPUFromNode(task, node)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("task %s UseAnnotation err: %s.", task.Name, err.Error())
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s UseAnnotation task<%s> select npu <%v>.",
		tp.GetPluginName(), task.Name, selectedNPU)
	tp.SetNPUTopologyToPodFn(task, selectedNPU, node)
	return tp.UpdateNodeInfo(node, selectedNPU)
}

// SetIsNetworkFaultAttention set network fault attention
func (tp *NPUHandler) SetIsNetworkFaultAttention(value bool) {
	tp.IsNetworkFaultAttention = value
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

// SetNpuNumInvalidMap  Set the single job not allow number. eg: 16P:9,10,11,12,13,14,15
func (tp *NPUHandler) SetNpuNumInvalidMap(value map[int]struct{}) {
	tp.NpuNumInvalidMap = value
}

// IsVaildNpuNum check the single job require is valid. eg: 16P:1,2,4,8,16;8P 1,2,4,8.
func (tp *NPUHandler) IsVaildNpuNum(value int) bool {
	_, ok := tp.NpuNumInvalidMap[value]
	return !ok && value <= tp.MaxNodeNPUNum
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

func (tp *NPUHandler) GetMaxCardNPUNum() int {
	if tp == nil {
		return 0
	}
	return tp.MaxCardNPUNum
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
		klog.V(util.LogWarningLev).Infof("judgeNodeAndTaskNPU node don't have enough resource, req<%d>, idle<%d>", taskNPU, len(nodeNPUTopology))
		return fmt.Errorf("%s node don't have enough resource, req<%d>, idle<%d>",
			util.NPUResourceShortageError, taskNPU, len(nodeNPUTopology))
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
		klog.V(util.LogDebugLev).Infof("selectNPUFromNode err: %s", err.Error())
		return nil, err
	}

	nodeTop, err := tp.GetUsableTopFromNode(node, tp.NPUTaskNum > 1)
	if err != nil {
		return nil, fmt.Errorf("selectNPUFromNode err: %s", err.Error())
	}
	if len(nodeTop) < taskNPUNum {
		return nil, fmt.Errorf("%s selectNPUFromNode node<%s> top<%v> not meet task req<%d>",
			util.NPUResourceShortageError, node.Name, nodeTop, taskNPUNum)
	}
	return nodeTop[:taskNPUNum], nil
}

// Preemptable default preempt logic: cross-card aggregation strategy
func (tp *NPUHandler) Preemptable(preemptor *api.TaskInfo, preemptees []*api.TaskInfo,
	vcNode *plugin.NPUNode) ([]*api.TaskInfo, bool) {
	if tp == nil || preemptor == nil || vcNode == nil || len(preemptees) == 0 {
		klog.V(util.LogInfoLev).Infof("Preemptable: invalid arguments, handler nil=%v preemptor nil=%v "+
			"vcNode nil=%v preemptees=%d", tp == nil, preemptor == nil, vcNode == nil, len(preemptees))
		return nil, false
	}
	maxCardNPUNum := tp.GetMaxCardNPUNum()
	if maxCardNPUNum <= 0 {
		klog.V(util.LogInfoLev).Infof("Preemptable: maxCardNPUNum is 0")
		return nil, false
	}
	reqNPUNum, err := tp.GetTaskReqNPUNum(preemptor)
	if err != nil || reqNPUNum <= 0 {
		klog.V(util.LogInfoLev).Infof("Preemptable: invalid reqNPUNum %d, err %v", reqNPUNum, err)
		return nil, false
	}

	klog.V(util.LogInfoLev).Infof("Preemptable: task<%s> req<%d> maxCardNPUNum<%d> on node<%s>, "+
		"preemptees<%d>", preemptor.Name, reqNPUNum, maxCardNPUNum, vcNode.Name, len(preemptees))

	cardFreeCount := plugin.CalcCardFreeCount(vcNode, preemptees, maxCardNPUNum)
	if len(cardFreeCount) == 0 {
		klog.V(util.LogInfoLev).Infof("Preemptable: no free cards after preemptees removed on node<%s>",
			vcNode.Name)
		return nil, false
	}

	totalFree := 0
	for _, fc := range cardFreeCount {
		totalFree += fc
	}
	if totalFree < reqNPUNum {
		klog.V(util.LogInfoLev).Infof("Preemptable: totalFree<%d> < reqNPUNum<%d> on node<%s>, not feasible",
			totalFree, reqNPUNum, vcNode.Name)
		return nil, false
	}

	type cardInfo struct {
		id        int
		freeCount int
	}
	cards := make([]cardInfo, 0, len(cardFreeCount))
	for id, fc := range cardFreeCount {
		if fc > 0 {
			cards = append(cards, cardInfo{id, fc})
		}
	}
	sort.Slice(cards, func(i, j int) bool { return cards[i].freeCount > cards[j].freeCount })

	remaining := reqNPUNum
	feasibleCards := make(map[int]struct{})
	for _, c := range cards {
		if remaining <= 0 {
			break
		}
		feasibleCards[c.id] = struct{}{}
		remaining -= c.freeCount
	}

	klog.V(util.LogInfoLev).Infof("Preemptable: task<%s> on node<%s>, selected %d feasible cards, "+
		"remaining need<%d>", preemptor.Name, vcNode.Name, len(feasibleCards), remaining)

	filtered := plugin.FilterPreempteesByFeasibleCards(vcNode, preemptees, feasibleCards, maxCardNPUNum)
	if len(filtered) == 0 {
		klog.V(util.LogInfoLev).Infof("Preemptable: task<%s> on node<%s>, no preemptees on feasible cards",
			preemptor.Name, vcNode.Name)
		return nil, false
	}
	klog.V(util.LogInfoLev).Infof("Preemptable: task<%s> on node<%s>, filtered %d/%d preemptees, feasible",
		preemptor.Name, vcNode.Name, len(filtered), len(preemptees))
	return filtered, true
}

// Reclaimable default reclaim logic: same cross-card aggregation strategy as preempt.
func (tp *NPUHandler) Reclaimable(reclaimer *api.TaskInfo, reclaimees []*api.TaskInfo,
	vcNode *plugin.NPUNode) ([]*api.TaskInfo, bool) {
	if tp == nil || reclaimer == nil || vcNode == nil || len(reclaimees) == 0 {
		klog.V(util.LogInfoLev).Infof("Reclaimable: invalid arguments, handler nil=%v reclaimer nil=%v "+
			"vcNode nil=%v reclaimees=%d", tp == nil, reclaimer == nil, vcNode == nil, len(reclaimees))
		return nil, false
	}
	maxCardNPUNum := tp.GetMaxCardNPUNum()
	if maxCardNPUNum <= 0 {
		klog.V(util.LogInfoLev).Infof("Reclaimable: invalid maxCardNPUNum<%d>", maxCardNPUNum)
		return nil, false
	}
	reqNPUNum, err := tp.GetTaskReqNPUNum(reclaimer)
	if err != nil || reqNPUNum <= 0 {
		klog.V(util.LogInfoLev).Infof("Reclaimable: invalid reqNPUNum %d, err %v", reqNPUNum, err)
		return nil, false
	}

	klog.V(util.LogInfoLev).Infof("Reclaimable: task<%s> req<%d> maxCardNPUNum<%d> on node<%s>, "+
		"reclaimees<%d>", reclaimer.Name, reqNPUNum, maxCardNPUNum, vcNode.Name, len(reclaimees))

	cardFreeCount := plugin.CalcCardFreeCount(vcNode, reclaimees, maxCardNPUNum)
	if len(cardFreeCount) == 0 {
		klog.V(util.LogInfoLev).Infof("Reclaimable: no free cards after reclaimees removed on node<%s>",
			vcNode.Name)
		return nil, false
	}

	totalFree := 0
	for _, fc := range cardFreeCount {
		totalFree += fc
	}
	if totalFree < reqNPUNum {
		klog.V(util.LogInfoLev).Infof("Reclaimable: totalFree<%d> < reqNPUNum<%d> on node<%s>, not feasible",
			totalFree, reqNPUNum, vcNode.Name)
		return nil, false
	}

	type cardInfo struct {
		id        int
		freeCount int
	}
	cards := make([]cardInfo, 0, len(cardFreeCount))
	for id, fc := range cardFreeCount {
		if fc > 0 {
			cards = append(cards, cardInfo{id, fc})
		}
	}
	sort.Slice(cards, func(i, j int) bool { return cards[i].freeCount > cards[j].freeCount })

	remaining := reqNPUNum
	feasibleCards := make(map[int]struct{})
	for _, c := range cards {
		if remaining <= 0 {
			break
		}
		feasibleCards[c.id] = struct{}{}
		remaining -= c.freeCount
	}

	klog.V(util.LogInfoLev).Infof("Reclaimable: task<%s> on node<%s>, selected %d feasible cards, "+
		"remaining need<%d>", reclaimer.Name, vcNode.Name, len(feasibleCards), remaining)

	filtered := plugin.FilterPreempteesByFeasibleCards(vcNode, reclaimees, feasibleCards, maxCardNPUNum)
	if len(filtered) == 0 {
		klog.V(util.LogInfoLev).Infof("Reclaimable: task<%s> on node<%s>, no reclaimees on feasible cards",
			reclaimer.Name, vcNode.Name)
		return nil, false
	}
	klog.V(util.LogInfoLev).Infof("Reclaimable: task<%s> on node<%s>, filtered %d/%d reclaimees, feasible",
		reclaimer.Name, vcNode.Name, len(filtered), len(reclaimees))
	return filtered, true
}

// ReleaseAnnotation release annotation
func (tp *NPUHandler) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
