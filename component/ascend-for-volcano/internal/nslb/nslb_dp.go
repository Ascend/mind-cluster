/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package nslb is using for HuaWei Ascend pin tor affinity.
*/
package nslb

import (
	"errors"
	"strconv"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// InitDPPolicyHandler initializes the dynamic programming policy handler
func InitDPPolicyHandler(attr util.SchedulerJobAttr, env plugin.ScheduleEnv) (plugin.SchedulerPluginNeed, bool) {
	if env.Tors == nil {
		return nil, false
	}
	th := &TorHandlerDP{
		TorHandler: TorHandler{
			pluginName:   pluginName,
			globalTorEnv: env.Tors,
		},
		superPodTors: make(map[int32][]*plugin.Tor),
	}

	// Retrieve job configuration from environment
	tmpJob, ok := env.Jobs[attr.Name]
	if !ok {
		return nil, false
	}
	th.Job = &tmpJob
	th.vPodSize = th.Job.SpBlockNPUNum / util.NPUIndex16
	// Process each Tor in the environment
	for _, tor := range env.Tors.Tors {
		superPodID, valid := validateTorSuperPod(tor, env.Nodes)
		if !valid {
			continue
		}
		th.superPodTors[superPodID] = append(th.superPodTors[superPodID], tor)
	}
	return th, true
}

// validateTorSuperPod checks if all nodes in the Tor belong to the same SuperPod
func validateTorSuperPod(tor *plugin.Tor, nodes map[string]plugin.NPUNode) (int32, bool) {
	superPodIDs := collectSuperPodIDs(tor, nodes)

	// Validate SuperPod consistency
	if superPodIDs.Len() != util.NPUIndex1 {
		return -1, false
	}

	return superPodIDs.PopAny()
}

// collectSuperPodIDs collects SuperPod IDs from all nodes in the Tor
func collectSuperPodIDs(tor *plugin.Tor, nodes map[string]plugin.NPUNode) sets.Int32 {
	superPodIDs := sets.NewInt32()
	for _, tNode := range tor.Servers {
		node, ok := nodes[tNode.Name]
		if !ok || node.SuperPodID < 0 {
			continue
		}
		superPodIDs.Insert(node.SuperPodID)
	}
	return superPodIDs
}

// ScoreBestNPUNodes score node by calculate task req npu num and node npu top
func (th *TorHandlerDP) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo,
	scoreMap map[string]float64) error {
	if th == nil || task == nil || len(nodes) == 0 || scoreMap == nil || th.globalTorEnv == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes err: %s.", err.Error())
		return err
	}
	refreshScoreMap(nodes, scoreMap)
	nodeMaps := util.ChangeNodesToNodeMaps(nodes)
	return th.setTorAffinityJobNodesScore(task, nodeMaps, scoreMap)
}

// setTorAffinityJobNodesScore nslb dp rule
func (th *TorHandlerDP) setTorAffinityJobNodesScore(task *api.TaskInfo,
	nodeMaps map[string]*api.NodeInfo, scoreMap map[string]float64) error {
	if !*th.Job.JobReadyTag {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogDebugLev).Infof("ScoreBestNPUNodes err: %s.", err)
		return nil
	}
	defer func() {
		th.scoreBestNPUNodes(task, nodeMaps, scoreMap)
	}()
	if th.ServerList != nil {
		return nil
	}
	th.globalTorEnv.MarkTorListByJobV2(nodeMaps, th.Job.Name)
	for id, tors := range th.superPodTors {
		th.useSuperPodTors = append(th.useSuperPodTors, initSuperPodTors(th.globalTorEnv.TorCount, id, tors))
	}
	th.setPartialTors()
	allocate(th.useSuperPodTors, th.Job.NPUTaskNum/th.vPodSize, th.vPodSize)
	for _, spTor := range th.useSuperPodTors {
		th.setJobServerList(spTor.fullTors, spTor.usedFull)
		th.setJobServerList(spTor.usePartialTors, spTor.usedPartial)
	}
	th.setServerListAttr()
	klog.V(util.LogDebugLev).Infof("batchNodeOrderFn Get %s for NPU %+v.", task.Name, scoreMap)
	return nil
}

// UseAnnotation select npu for task from node
func (th *TorHandlerDP) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if th == nil || task == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation err: %s.", err.Error())
		return nil
	}
	task.Pod.Annotations[isSharedTor] = strconv.Itoa(freeTor)
	task.Pod.Annotations[isHealthy] = strconv.Itoa(healthyTor)
	torIp := th.globalTorEnv.GetTorIpMap()[node.Name]
	tor, getTor := th.globalTorEnv.GetTorMaps()[torIp]
	if getTor && len(th.ServerList) > 1 {
		task.Pod.Annotations[isSharedTor] = strconv.Itoa(tor.IsSharedTor)
		task.Pod.Annotations[isHealthy] = strconv.Itoa(tor.IsHealthy)
	}
	if task.Pod.Annotations[isSharedTor] == strconv.Itoa(sharedTor) {
		task.Pod.Annotations[SharedTorIp] = tor.IP
	}
	return &node
}

func (th *TorHandlerDP) setServerListAttr() {
	enableSharedTor := th.globalTorEnv.GetSharedTorNum()
	if len(th.ServerList) == util.NPUIndex1 {
		return
	}
	for _, tor := range th.ServerList {
		if tor.IsSharedTor == exclusiveTor || tor.IsSharedTor == sharedTor {
			th.globalTorEnv.GetTorMaps()[tor.IP].IsSharedTor = sharedTor
			enableSharedTor--
			continue
		}
		if tor.FreeServerCount == len(tor.Servers) {
			th.globalTorEnv.GetTorMaps()[tor.IP].IsSharedTor = exclusiveTor
			continue
		}
		if enableSharedTor > 0 {
			th.globalTorEnv.GetTorMaps()[tor.IP].IsSharedTor = sharedTor
			enableSharedTor--
			continue
		}
		th.globalTorEnv.GetTorMaps()[tor.IP].IsSharedTor = exclusiveTor
	}
}

// setPartialTors set nslb-dp is need partial tors
func (th *TorHandlerDP) setPartialTors() {
	if th.vPodSize == 0 {
		return
	}
	var vPodNum int
	for _, spt := range th.useSuperPodTors {
		vPodNum += spt.full / th.vPodSize
	}
	if vPodNum >= th.Job.NPUTaskNum/th.vPodSize {
		return
	}
	klog.V(util.LogWarningLev).Infof("job %s full tor vpod num is %v is not enough, "+
		"will use partial tor", th.Job.Name, vPodNum)
	for i := 0; i < util.NPUIndex3; i++ {
		vPodNum = 0
		for _, spt := range th.useSuperPodTors {
			spt.partial += getTorsFreeServerNum(spt.partialTors[i])
			spt.remainPart = spt.partial
			spt.usePartialTors = append(spt.usePartialTors, spt.partialTors[i]...)
			vPodNum += (spt.full + spt.partial) / th.vPodSize
		}
		if vPodNum >= th.Job.NPUTaskNum/th.vPodSize {
			return
		}
		klog.V(util.LogWarningLev).Infof("stage %d job %s tor vpod num is %v is not enough, "+
			"will use next stage tor", i, th.Job.Name, vPodNum)
	}
}

// setFillJobServerList set the fill job server list in nslb dp
func (th *TorHandlerDP) setJobServerList(Tors []*plugin.Tor, taskNum int) {
	if taskNum == 0 {
		return
	}
	var count int
	for i := 0; i < len(Tors); i++ {
		tmpTor := &plugin.Tor{
			FreeServerCount: Tors[i].FreeServerCount,
			IsSharedTor:     Tors[i].IsSharedTor,
			IP:              Tors[i].IP,
		}
		for _, k := range Tors[i].Servers {
			if k.CurrentJob != nil && *k.CurrentJob == th.Job.Name {
				count++
				tmpTor.Servers = append(tmpTor.Servers, k)
			}
			if count == taskNum {
				break
			}
		}
		th.ServerList = append(th.ServerList, tmpTor)
	}
}
