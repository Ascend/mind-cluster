/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package superpod for score nodes
package superpod

import (
	"errors"
	"sort"
	"strconv"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func (tp *module910a5SuperPod) selectNodeForStandaloneJob(nodes []*api.NodeInfo) []*api.NodeInfo {
	if len(nodes) < 1 {
		klog.V(util.LogWarningLev).Infof("%s there is not enough nodes for a standalone job", tp.GetPluginName())
		return nodes
	}
	klog.V(util.LogInfoLev).Infof("%s select one node for a standalone job", tp.GetPluginName())
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Idle.ScalarResources[v1.ResourceName(tp.ReqNPUName)] < nodes[j].Idle.
			ScalarResources[v1.ResourceName(tp.ReqNPUName)]
	})
	leastResourceNode := nodes[0]
	suitableNodes := make([]*api.NodeInfo, 0)
	for _, node := range nodes {
		if node.Idle.ScalarResources[v1.ResourceName(tp.ReqNPUName)] > leastResourceNode.Idle.ScalarResources[v1.
			ResourceName(tp.ReqNPUName)] {
			break
		}
		suitableNodes = append(suitableNodes, node)
	}

	return suitableNodes
}

func (tp *module910a5SuperPod) scoreNodeForReadyJob(task *api.TaskInfo,
	job *plugin.SchedulerJob, sMap map[string]float64) {
	var rank int
	var err error
	if rankIndex, ok := task.Pod.Annotations[plugin.PodRankIndexKey]; ok {
		rank, err = strconv.Atoi(rankIndex)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s %s scoreNodeForReadyJob failed %s: rankIndex is not a int type",
				tp.GetPluginName(), task.Name, task.Name)
			return
		}
	} else {
		klog.V(util.LogWarningLev).Infof("%s %s scoreNodeForReadyJob %s: rankIndex is not exist",
			tp.GetPluginName(), task.Name, task.Name)
		nTask, ok := job.Tasks[task.UID]
		if !ok {
			klog.V(util.LogErrorLev).Infof("%s scoreNodeForReadyJob %s: task is not exist", tp.GetPluginName(),
				task.Name)
			return
		}
		rank = nTask.Index
	}
	if tp.spBlock == 0 {
		klog.V(util.LogErrorLev).Info("get a zero value of spBlock")
		return
	}
	superPodRank := rank / tp.spBlock
	localRank := rank % tp.spBlock

	superPodRankIndex := strconv.Itoa(superPodRank)
	if superPodRankIndex == "" {
		return
	}
	if localRank >= len(job.SuperPods[superPodRankIndex]) {
		klog.V(util.LogErrorLev).Infof("superPodRank: %d, localRank: %d out of rank", superPodRank, localRank)
		return
	}
	node := job.SuperPods[superPodRankIndex][localRank]
	if _, find := sMap[node.Name]; find {
		sMap[node.Name] += scoreForNode
		return
	}
	klog.V(util.LogWarningLev).Infof("scoreNodeForReadyJob failed, the selected node %v is not in the sMap, "+
		"job superpods: %v, sMap: %v", node.Name, job.SuperPods, sMap)
}

func (tp *module910a5SuperPod) scoreNodeBatchForReadyJob(task *api.TaskInfo, job *plugin.SchedulerJob,
	sMap map[string]float64) {
	if task == nil || job == nil || len(sMap) == 0 || tp.spBlock == 0 {
		klog.V(util.LogErrorLev).Infof("scoreNodeBatchForReadyJob %s", errors.New(util.ArgumentError))
		return
	}
	rankIdMap := tp.obtainBatchScoreRank(task, job)
	if len(rankIdMap) == 0 {
		klog.V(util.LogErrorLev).Infof("%s scoreNodeBatchForReadyJob %s: rankIdMap empty",
			tp.GetPluginName(), task.Name)
		*job.JobReadyTag = false
		return
	}
	for rankId := range rankIdMap {
		superPodRank := rankId / tp.spBlock
		localRank := rankId % tp.spBlock
		klog.V(util.LogInfoLev).Infof("superPodRank: %d, localRank: %d", superPodRank, localRank)
		superPodRankIndex := strconv.Itoa(superPodRank)
		if localRank >= len(job.SuperPods[superPodRankIndex]) {
			klog.V(util.LogErrorLev).Infof("superPodRank: %d, localRank: %d out of rank", superPodRank, localRank)
			*job.JobReadyTag = false
			break
		}
		spn := job.SuperPods[superPodRankIndex][localRank]
		if _, ok := sMap[spn.Name]; !ok {
			klog.V(util.LogErrorLev).Infof("%s scoreNodeBatchForReadyJob %s: node<%s> not in sMap, select fail",
				tp.GetPluginName(), task.Name, spn.Name)
			*job.JobReadyTag = false
			break
		}
		klog.V(util.LogInfoLev).Infof("%s scoreNodeBatchForReadyJob %s: node<%s/%s> is exist in "+
			"SuperPodID: %d, select success", tp.GetPluginName(), task.Name, spn.Name, superPodRankIndex,
			spn.SuperPodID)
		sMap[spn.Name] = float64(scoreForNode - rankId)
	}
}

func (tp *module910a5SuperPod) obtainBatchScoreRank(task *api.TaskInfo, job *plugin.SchedulerJob) map[int]struct{} {
	if task == nil || job == nil {
		klog.V(util.LogErrorLev).Infof("obtainBatchScoreRank %s", errors.New(util.ArgumentError))
		return nil
	}
	spec, ok := task.Pod.Annotations[taskSpec]
	if !ok {
		klog.V(util.LogErrorLev).Infof("obtainBatchScoreRank %s: (%s/%s) obtain annotation %s failed, skip",
			tp.GetPluginName(), task.Namespace, task.Name, taskSpec)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("obtainOriginalRankIdMap job (%s/%s), len(job.Tasks) %d",
		job.NameSpace, job.Name, len(job.Tasks))
	m := make(map[int]struct{}, len(job.Tasks))
	for _, task := range job.Tasks {
		if !task.IsNPUTask() || task.Annotation[taskSpec] != spec {
			continue
		}
		if task.PodStatus != v1.PodPending {
			continue
		}
		rankIndex, ok := task.Annotation[plugin.PodRankIndexKey]
		if !ok {
			klog.V(util.LogWarningLev).Infof("obtainBatchScoreRank (%s/%s): rankIndex is not exist",
				task.NameSpace, task.Name)
			continue
		}
		rank, err := strconv.Atoi(rankIndex)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("obtainBatchScoreRank (%s/%s): rankIndex is not int",
				task.NameSpace, task.Name)
			continue
		}
		m[rank] = struct{}{}
	}
	klog.V(util.LogInfoLev).Infof("obtainBatchScoreRank job (%s/%s), len(rankMap) %d",
		job.NameSpace, job.Name, len(m))
	return m
}

func obtainOriginalRankIdMap(job *plugin.SchedulerJob) map[int]util.NPUTask {
	if job == nil {
		klog.V(util.LogErrorLev).Infof("obtainOriginalRankIdMap %s", errors.New(util.ArgumentError))
		return nil
	}
	klog.V(util.LogInfoLev).Infof("obtainOriginalRankIdMap job (%s/%s), len(job.Tasks) %d",
		job.NameSpace, job.Name, len(job.Tasks))
	m := make(map[int]util.NPUTask, len(job.Tasks))
	for _, task := range job.Tasks {
		if !task.IsNPUTask() {
			continue
		}
		if task.PodStatus != v1.PodPending {
			continue
		}
		rankIndex, ok := task.Annotation[plugin.PodRankIndexKey]
		if !ok {
			klog.V(util.LogWarningLev).Infof("obtainOriginalRankIdMap (%s/%s): rankIndex is not exist",
				task.NameSpace, task.Name)
			m[task.Index] = task
			continue
		}
		rank, err := strconv.Atoi(rankIndex)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("obtainOriginalRankIdMap (%s/%s): rankIndex is not int",
				task.NameSpace, task.Name)
			continue
		}
		m[rank] = task
	}
	klog.V(util.LogInfoLev).Infof("obtainOriginalRankIdMap job (%s/%s), len(rankMap) %d",
		job.NameSpace, job.Name, len(m))
	return m
}
