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
	"fmt"
	"sort"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// CheckNodeNPUByTask check nod npu meet task req
func (th *TorHandlerV1) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if th == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s.", err.Error())
		return err
	}
	klog.V(util.LogDebugLev).Infof("%s NodePredicate %s select successes.", th.GetPluginName(), node.Name)
	if !(th.Job.SchedulingTaskNum < len(th.Job.Tasks) && th.Job.IsTorAffinityJob()) {
		return nil
	}
	return th.CheckTorJobSinglePodDeleteV1(task, node)
}

// ScoreBestNPUNodes score node by calculate task req npu num and node npu top
func (th *TorHandlerV1) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo,
	scoreMap map[string]float64) error {
	if th == nil || task == nil || len(nodes) == 0 || len(scoreMap) == 0 || th.globalTorEnv == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes err: %s.", err.Error())
		return err
	}
	nodeMaps := util.ChangeNodesToNodeMaps(nodes)
	return th.SetTorAffinityJobNodesScoreV1(task, nodeMaps, scoreMap)
}

// SetTorAffinityJobNodesScoreV1 nslb 1.0 rule
func (th *TorHandlerV1) SetTorAffinityJobNodesScoreV1(task *api.TaskInfo,
	nodeMaps map[string]*api.NodeInfo, scoreMap map[string]float64) error {
	if len(nodeMaps) == 0 || len(scoreMap) == 0 || !*th.Job.JobReadyTag {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogDebugLev).Infof("ScoreBestNPUNodes %s.", err)
		return nil
	}

	result := th.setTorHandlerServerListV1(nodeMaps)
	if result != nil {
		klog.V(util.LogErrorLev).Infof("check job %s tor affinity failed: %s", th.Job.Name, result)
		switch th.Job.Label[TorAffinityKey] {
		case LargeModelTag:
			*th.Job.JobReadyTag = false
		case NormalSchema:
			th.SetNormalJobServerList(th.Job.SchedulingTaskNum)
		default:
			return nil
		}
	}
	if errGet := th.scoreBestNPUNodes(task, nodeMaps, scoreMap); errGet != nil {
		// get suitable node failed
		klog.V(util.LogDebugLev).Infof("batchNodeOrderFn task[%s] is failed[%s].", task.Name, util.SafePrint(errGet))
	}
	klog.V(util.LogDebugLev).Infof("batchNodeOrderFn set %s for NPU %+v.", task.Name, scoreMap)
	return result
}

func (th *TorHandlerV1) setTorHandlerServerListV1(nodeMaps map[string]*api.NodeInfo) error {
	if th == nil || th.globalTorEnv == nil || len(nodeMaps) == 0 {
		err := errors.New(util.ArgumentError)
		return fmt.Errorf("initTorHandlerV1 err: %s", err.Error())
	}
	if len(th.ServerList) != 0 {
		klog.V(util.LogDebugLev).Infof("InitTorHandlerV1 len(serverList):%d", len(th.ServerList))
		return nil
	}
	schedulingTaskNum := th.Job.SchedulingTaskNum
	th.globalTorEnv.MarkTorListByJobV1(nodeMaps, th.Job.Name, schedulingTaskNum)
	fullTorNum := th.globalTorEnv.GetFullTorNumFromTorInfo(th.Job.Name)
	sort.Slice(th.globalTorEnv.Tors, func(i, j int) bool {
		return th.globalTorEnv.Tors[i].FreeServerCount > th.globalTorEnv.Tors[j].FreeServerCount
	})
	netSliceNum := th.globalTorEnv.TorCount
	if schedulingTaskNum < netSliceNum {
		if err := th.SetFillJobServerList(th.globalTorEnv.Tors, schedulingTaskNum); err == nil ||
			isFillJob(th.Job.Label, th.Job.NPUTaskNum) {
			return err
		}
	}
	taskRow, taskColumn := getTaskRowAndTaskColumn(schedulingTaskNum, netSliceNum)
	if taskRow == -1 {
		return fmt.Errorf("taskRow and taskColumn is illegal")
	}
	if taskRow+1 <= fullTorNum {
		th.SetJobServerCacheTosHandler(th.globalTorEnv.Tors, taskRow, taskColumn)
		th.MarkMulJobServerList()
		return nil
	}
	logicTor, fullTorNum := th.globalTorEnv.GetLogicTorsAndFullTorNum(th.Job.Name, taskColumn,
		taskRow, netSliceNum)
	if logicTor == nil {
		return fmt.Errorf("logicTor is illegal")
	}
	if taskRow < 1 && taskColumn != netSliceNum-1 {
		err := th.SetFillJobServerList(logicTor, schedulingTaskNum)
		th.MarkMulJobServerList()
		return err
	}

	th.SetJobServerCacheTosHandler(logicTor, taskRow, taskColumn)
	th.MarkMulJobServerList()
	return nil
}

// SetJobServerCacheTosHandler set job server list and update the job in sHandler
func (th *TorHandlerV1) SetJobServerCacheTosHandler(tors []*plugin.Tor, taskRow, taskColumn int) {
	if th == nil || len(tors) == 0 {
		klog.V(util.LogDebugLev).Infof("SetJobServerCacheTosHandler failed:%s", util.ArgumentError)
		return
	}
	if taskRow >= len(tors) {
		klog.V(util.LogDebugLev).Infof("invalid taskRow: %d, pyTor length: %d", taskRow, len(tors))
		return
	}
	tmpTors := copyTorList(tors[:taskRow])
	tmpTor := &plugin.Tor{}
	tmpTor.Servers = append(tmpTor.Servers, tors[taskRow].Servers[:taskColumn+1]...)
	tmpTors = append(tmpTors, tmpTor)
	th.ServerList = tmpTors
}

// MarkMulJobServerList mark the job if the server job used is over 1 tor
func (th *TorHandlerV1) MarkMulJobServerList() {
	if th.ServerList == nil {
		return
	}
	for _, tor := range th.ServerList {
		if tor.Servers == nil {
			continue
		}
		for _, server := range tor.Servers {
			server.IsUsedByMulJob = true
		}
	}
}

// SetNormalJobServerList set the server list of normal job in nslb 1.0
func (th *TorHandlerV1) SetNormalJobServerList(schedulingTaskNum int) {
	if th == nil {
		klog.V(util.LogDebugLev).Infof("SetNormalJobServerList failed:%s", util.ArgumentError)
		return
	}
	th.ServerList = []*plugin.Tor{}
	var count int
	for _, tor := range th.globalTorEnv.Tors {
		tmpTor := &plugin.Tor{IP: tor.IP, Id: tor.Id}
		for _, server := range tor.Servers {
			if server.CurrentJob != nil && *server.CurrentJob == th.Job.Name {
				tmpTor.Servers = append(tmpTor.Servers, server)
				count++
			}
			if count != schedulingTaskNum {
				continue
			}
			th.ServerList = append(th.ServerList, tmpTor)
			if len(th.ServerList) > 1 {
				th.MarkMulJobServerList()
			}
			return
		}
		th.ServerList = append(th.ServerList, tmpTor)
	}
	*th.Job.JobReadyTag = false
}

// CheckTorJobSinglePodDeleteV1 valid node.
func (th *TorHandlerV1) CheckTorJobSinglePodDeleteV1(taskInfo *api.TaskInfo, vcNode plugin.NPUNode) error {
	if isFillJob(th.Job.Label, th.Job.NPUTaskNum) {
		return fmt.Errorf("check node err by: large model job can not over tor")
	}
	nodeName, ok := th.Job.Annotation[taskInfo.Name]
	if !ok {
		klog.V(util.LogWarningLev).Infof("Cannot get task used fault node name")
		return nil
	}
	faultServer, isTorNode := th.globalTorEnv.GetServerMaps()[nodeName]
	if !isTorNode {
		return fmt.Errorf("cannot get task used fault node name")
	}

	server, isTorNode := th.globalTorEnv.GetServerMaps()[vcNode.Name]
	if !isTorNode {
		return fmt.Errorf("node is not in tor node list by not get server")
	}

	torIp, getTorIp := th.globalTorEnv.GetTorIpMap()[vcNode.Name]
	if !getTorIp {
		return fmt.Errorf("node is not in tor node list by not get tor ip")
	}

	tor, isTor := th.globalTorEnv.GetTorMaps()[torIp]
	if !isTor {
		return fmt.Errorf("node is not in tor node list by not get tor")
	}

	if faultServer.SliceId != server.SliceId || tor.HasAcrossJob(false, th.Job.Name) {
		return fmt.Errorf("node sliceId is not meet task require")
	}
	return nil
}

func getTaskRowAndTaskColumn(nTaskNum int, netSliceNum int) (int, int) {
	if netSliceNum == 0 {
		return -1, -1
	}
	taskRow := nTaskNum / netSliceNum
	if nTaskNum%netSliceNum == 0 {
		taskRow = nTaskNum/netSliceNum - 1
	}
	taskColumn := (nTaskNum%netSliceNum + netSliceNum - 1) % netSliceNum
	return taskRow, taskColumn
}
