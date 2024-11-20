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
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (reCache *DealReSchedulerCache) setFaultNodes(faultNodes []FaultNode) {
	reCache.FaultNodes = faultNodes
}

func (reCache *DealReSchedulerCache) setFaultJobs(faultJobs []FaultJob) {
	reCache.FaultJobs = faultJobs
}

func (reCache *DealReSchedulerCache) setNodeHeartbeat(nodeHeartbeat []NodeHeartbeat) {
	reCache.NodeHeartbeats = nodeHeartbeat
}

func (reCache *DealReSchedulerCache) setNodeRankOccurrenceMap(
	nodeRankOccurrenceMap map[api.JobID][]*AllocNodeRankOccurrence) {
	reCache.AllocNodeRankOccurrenceMap = nodeRankOccurrenceMap
}

func (reCache DealReSchedulerCache) getFaultNodesFromCM(buffer string) ([]FaultNode, error) {
	var faultNodes []FaultNode
	if unmarshalErr := json.Unmarshal([]byte(buffer), &faultNodes); unmarshalErr != nil {
		klog.V(util.LogInfoLev).Infof("Unmarshal FaultNodes from cache failed")
		return nil, fmt.Errorf("faultNodes convert from CM error: %s", util.SafePrint(unmarshalErr))
	}
	return faultNodes, nil
}

func (reCache DealReSchedulerCache) getFaultJobsFromCM(buffer string) ([]FaultJob, error) {
	var faultJobs []FaultJob
	if unmarshalErr := json.Unmarshal([]byte(buffer), &faultJobs); unmarshalErr != nil {
		klog.V(util.LogInfoLev).Infof("Unmarshal FaultJobs from cache failed")
		return nil, fmt.Errorf("faultJobs convert from CM failed")
	}
	return faultJobs, nil
}

func (reCache DealReSchedulerCache) getRetryTimesFromCM(buffer string) (map[api.JobID]*RemainRetryTimes, error) {
	rTimes := make(map[api.JobID]*RemainRetryTimes)
	if unmarshalErr := json.Unmarshal([]byte(buffer), &rTimes); unmarshalErr != nil {
		klog.V(util.LogDebugLev).Infof("Unmarshal remain times from cache failed")
		return nil, fmt.Errorf("remain times convert from CM error: %s", util.SafePrint(unmarshalErr))
	}
	return rTimes, nil
}

func (reCache DealReSchedulerCache) getNodeHeartbeatFromCM(buffer string) ([]NodeHeartbeat, error) {
	var nodeHBs []NodeHeartbeat
	if unmarshalErr := json.Unmarshal([]byte(buffer), &nodeHBs); unmarshalErr != nil {
		klog.V(util.LogDebugLev).Infof("Unmarshal NodeHeartbeat from cache failed")
		return nil, fmt.Errorf("faultNodes convert from CM error: %s", util.SafePrint(unmarshalErr))
	}
	return nodeHBs, nil
}

func (reCache DealReSchedulerCache) getNodeRankOccurrenceMapFromCM(
	buffer string) (map[api.JobID][]*AllocNodeRankOccurrence, error) {
	var nodeRankOccMap map[api.JobID][]*AllocNodeRankOccurrence
	if unmarshalErr := json.Unmarshal([]byte(buffer), &nodeRankOccMap); unmarshalErr != nil {
		klog.V(util.LogDebugLev).Infof("Unmarshal AllocNodeRankOccurrence from cache failed")
		return nil, fmt.Errorf("faultNodes convert from CM error: %s", util.SafePrint(unmarshalErr))
	}
	return nodeRankOccMap, nil
}

// SetFaultNodesFromCM unmarshal FaultNodes from string into struct and set the value
func (reCache *DealReSchedulerCache) SetFaultNodesFromCM() error {
	if reCache == nil {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM failed: %s, reCache is none", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	faultNodeData, ok := reCache.CMData[CmFaultNodeKind]
	if !ok {
		return fmt.Errorf("reading %s data from reScheduler configmap failed", CmFaultNodeKind)
	}
	if faultNodeData == "" {
		return nil
	}
	faultNodes, err := reCache.getFaultNodesFromCM(faultNodeData)
	if err != nil {
		return fmt.Errorf("getFaultNodesFromCM %s", util.SafePrint(err))
	}
	reCache.setFaultNodes(faultNodes)
	return nil
}

// SetFaultJobsFromCM unmarshal FaultJobs from string into struct and set the value
func (reCache *DealReSchedulerCache) SetFaultJobsFromCM(jobType string) error {
	if reCache == nil {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM failed: %s, reCache is none", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	if len(jobType) == 0 {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM failed: %s: jobType is none", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	faultJobData, ok := reCache.CMData[jobType]
	if !ok {
		return fmt.Errorf("reading %s data from reScheduler configmap failed", jobType)
	}
	if faultJobData == "" {
		return nil
	}
	faultJobs, err := reCache.getFaultJobsFromCM(faultJobData)
	if err != nil {
		return fmt.Errorf("getFaultNodesFromCM %s", util.SafePrint(err))
	}
	reCache.setFaultJobs(faultJobs)
	return nil
}

// SetNodeHeartbeatFromCM unmarshal NodeHeartbeat from string into struct and set the value
func (reCache *DealReSchedulerCache) SetNodeHeartbeatFromCM() error {
	if reCache == nil {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM failed: %s, reCache is none", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	nodeHBsData, ok := reCache.CMData[CmNodeHeartbeatKind]
	if !ok {
		return fmt.Errorf("reading %s data from reScheduler configmap failed", CmNodeHeartbeatKind)
	}
	if nodeHBsData == "" {
		return nil
	}
	nodeHBs, err := reCache.getNodeHeartbeatFromCM(nodeHBsData)
	if err != nil {
		return fmt.Errorf("getFaultNodesFromCM %s", util.SafePrint(err))
	}
	reCache.setNodeHeartbeat(nodeHBs)
	return nil
}

// SetRetryTimesFromCM unmarshal NodeHeartbeat from string into struct and set the value
func (reCache *DealReSchedulerCache) SetRetryTimesFromCM() error {
	if reCache == nil {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM failed: %s, reCache is none", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	data, ok := reCache.CMData[CmJobRemainRetryTimes]
	if !ok {
		return fmt.Errorf("reading %s data from reScheduler configmap failed", CmNodeHeartbeatKind)
	}
	if data == "" {
		return nil
	}
	remain, err := reCache.getRetryTimesFromCM(data)
	if err != nil {
		return fmt.Errorf("getFaultNodesFromCM %s", util.SafePrint(err))
	}
	reCache.JobRemainRetryTimes = remain
	return nil
}

// SetNodeRankOccurrenceMapFromCM unmarshal NodeRankOccurrenceMap from string into struct and set the value
func (reCache *DealReSchedulerCache) SetNodeRankOccurrenceMapFromCM() error {
	if reCache == nil {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM failed: %s, reCache is none", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	nodeRankOccMapData, ok := reCache.CMData[CmNodeRankTimeMapKind]
	if !ok {
		return fmt.Errorf("reading %s data from reScheduler configmap failed", CmNodeRankTimeMapKind)
	}
	if nodeRankOccMapData == "" {
		return nil
	}
	nodeRankOccMap, err := reCache.getNodeRankOccurrenceMapFromCM(nodeRankOccMapData)
	if err != nil {
		return fmt.Errorf("getFaultNodesFromCM %s", util.SafePrint(err))
	}
	reCache.setNodeRankOccurrenceMap(nodeRankOccMap)
	return nil
}

func (reCache *DealReSchedulerCache) marshalCacheDataToString(data interface{}) (string, error) {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("marshalCacheDataToString err: %s", util.SafePrint(err))
		return "", err
	}
	return string(dataBuffer), nil
}

// getRealFaultJobs only return FaultJobs whose IsFaultJob is true
func (reCache DealReSchedulerCache) getRealFaultJobs() ([]FaultJob, error) {
	var realFaultJobs []FaultJob
	for _, fJob := range reCache.FaultJobs {
		if (!fJob.IsFaultJob && !fJob.IsJobHasPreSeparateNPUKey()) || fJob.ReScheduleKey == JobOffRescheduleLabelValue {
			continue // only save real-fault and reschedule-enabled jobs
		}

		faultReason := PodFailed
		for _, faultType := range fJob.FaultTypes {
			if faultType == NodeUnhealthy || faultType == NodeCardUnhealthy {
				faultReason = faultType
				break
			}
		}
		fJob.faultReason = faultReason
		realFaultJobs = append(realFaultJobs, fJob)
	}
	if len(realFaultJobs) == 0 {
		klog.V(util.LogDebugLev).Infof("getRealFaultJobs %s.", NoFaultJobsErr)
		return nil, fmt.Errorf(NoFaultJobsErr)
	}
	return realFaultJobs, nil
}

// IsJobInRealFaultJobs only return true when Job is in FaultJobs
func (reCache DealReSchedulerCache) IsJobInRealFaultJobs(job api.JobID) bool {
	fJobs, err := reCache.getRealFaultJobs()
	if err != nil {
		return false
	}
	for _, fJob := range fJobs {
		if fJob.JobUID == job {
			return true
		}
	}
	return false
}

// GetRealFaultNodes get the nodes whose isFaultNode property takes true value
func (reCache DealReSchedulerCache) GetRealFaultNodes() []FaultNode {
	var realFaultNodes []FaultNode
	for _, fNode := range reCache.FaultNodes {
		if !fNode.IsFaultNode {
			continue
		}
		realFaultNodes = append(realFaultNodes, fNode)
	}
	return realFaultNodes
}

func (reCache *DealReSchedulerCache) writeFaultNodesToCMString() (string, error) {
	realFaultNode := reCache.GetRealFaultNodes()
	if len(realFaultNode) == 0 {
		return "", nil
	}
	nodeData, err := reCache.marshalCacheDataToString(realFaultNode)
	if err != nil {
		return "", fmt.Errorf("writeFaultNodesToCM: %s", util.SafePrint(err))
	}
	return nodeData, nil
}

func (reCache *DealReSchedulerCache) writeFaultJobsToCMString() (string, error) {
	realFaultJob, err := reCache.getRealFaultJobs()
	if err != nil {
		if err.Error() == NoFaultJobsErr {
			return "", nil
		}
		return "", fmt.Errorf("writeFaultJobsToCM: %s", util.SafePrint(err))
	}
	jobData, err := reCache.marshalCacheDataToString(getRealFaultJobForCM(realFaultJob))
	if err != nil {
		klog.V(util.LogErrorLev).Infof("WriteFaultJobsToCM: %s.", util.SafePrint(err))
		return "", fmt.Errorf("writeFaultJobsToCM: %s", util.SafePrint(err))
	}
	return jobData, nil
}

func (reCache *DealReSchedulerCache) writeNodeHeartbeatToCMString() (string, error) {
	var nodeHB NodeHeartbeat
	var nodeHBs []NodeHeartbeat
	for _, fNode := range reCache.FaultNodes {
		nodeHB = NodeHeartbeat{
			NodeName:      fNode.NodeName,
			HeartbeatTime: fNode.NewHeartbeatTime,
			UpdateTime:    fNode.UpdateHeartbeatTime,
		}
		nodeHBs = append(nodeHBs, nodeHB)
	}
	nodeHBsData, err := reCache.marshalCacheDataToString(nodeHBs)
	if err != nil {
		return "", fmt.Errorf("writeNodeHeartbeatToCM: %s", util.SafePrint(err))
	}
	return nodeHBsData, nil
}

func (reCache *DealReSchedulerCache) writeRemainTimesToCMString() (string, error) {
	if len(reCache.JobRemainRetryTimes) == 0 {
		return "", nil
	}
	nodeHBsData, err := reCache.marshalCacheDataToString(reCache.JobRemainRetryTimes)
	if err != nil {
		return "", fmt.Errorf("writeRemainTimesToCMString: %s", util.SafePrint(err))
	}
	return nodeHBsData, nil
}

func (reCache *DealReSchedulerCache) writeNodeRankOccurrenceMapToCMString() (string, error) {
	if len(reCache.AllocNodeRankOccurrenceMap) == 0 {
		return "", nil
	}
	nodeRankOccMapData, err := reCache.marshalCacheDataToString(reCache.AllocNodeRankOccurrenceMap)
	if err != nil {
		return "", fmt.Errorf("writeNodeRankOccurrenceMapToCM: %s", util.SafePrint(err))
	}
	return nodeRankOccMapData, nil
}

func (reCache *DealReSchedulerCache) writeJobRankIndexToCMString(fJob *FaultJob) (*FaultRankIdsJobCMData, string,
	error) {
	faultRankIds := &FaultRankIdsJobCMData{
		FaultRankIds: fJob.JobRankIds,
		CreatTime:    time.Now().Unix(),
	}
	faultRankIdsData, err := reCache.marshalCacheDataToString(faultRankIds)
	if err != nil {
		return nil, "", fmt.Errorf("%s writeJobRankIndexToCMString: %s", fJob.JobName, util.SafePrint(err))
	}
	return faultRankIds, faultRankIdsData, nil
}

// WriteReSchedulerCacheToEnvCache write the modifications on cache data to env to update re-scheduling configmap
func (reCache *DealReSchedulerCache) WriteReSchedulerCacheToEnvCache(env *plugin.ScheduleEnv, jobType string) error {
	if reCache == nil || env == nil {
		return errors.New(util.ArgumentError)
	}
	env.Cache.Names[RePropertyName] = CmName
	env.Cache.Namespaces[RePropertyName] = CmNameSpace
	fNodeString, err := reCache.writeFaultNodesToCMString()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("WriteReSchedulerCacheToEnvCache: %s", util.SafePrint(err))
	}
	fJobString, err := reCache.writeFaultJobsToCMString()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("WriteReSchedulerCacheToEnvCache: %s", util.SafePrint(err))
	}
	nodeHBString, err := reCache.writeNodeHeartbeatToCMString()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("WriteReSchedulerCacheToEnvCache:%s", util.SafePrint(err))
	}

	jobRemainRetryTimes, err := reCache.writeRemainTimesToCMString()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("WriteReSchedulerCacheToEnvCache: %s", util.SafePrint(err))
	}

	nodeRankOccurrenceMapString, err := reCache.writeNodeRankOccurrenceMapToCMString()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("WriteReSchedulerCacheToEnvCache: %s", util.SafePrint(err))
	}
	cmData, ok := env.Cache.Data[RePropertyName]
	if !ok {
		cmData = make(map[string]string, util.MapInitNum)
		env.Cache.Data[RePropertyName] = cmData
	}
	cmData[CmFaultNodeKind] = fNodeString
	cmData[jobType] = fJobString
	cmData[CmNodeHeartbeatKind] = nodeHBString
	cmData[CmNodeRankTimeMapKind] = nodeRankOccurrenceMapString
	cmData[CmJobRemainRetryTimes] = jobRemainRetryTimes
	if jobType != CmFaultJob910x8Kind && jobType != CmFaultJob910x4Kind && jobType != CmFaultJob910bx8Kind &&
		jobType != CmFaultJob910bx16Kind {
		return nil
	}
	if err := reCache.writeRecoveryCacheToEnv(env); err != nil {
		klog.V(util.LogDebugLev).Infof("WriteReSchedulerCacheToEnvCache: %s", util.SafePrint(err))
	}
	return nil
}

func (reCache *DealReSchedulerCache) writeRecoveryCacheToEnv(env *plugin.ScheduleEnv) error {
	realFaultJob, err := reCache.getRealFaultJobs()
	if err != nil {
		return err
	}
	for _, fJob := range realFaultJob { // configmap for recovery
		if fJob.IsFaultJob {
			jobRankIndex, jobRankIndexString, err := reCache.writeJobRankIndexToCMString(&fJob)
			if err != nil {
				klog.V(util.LogErrorLev).Infof(err.Error())
				continue
			}

			faultCm, ok := env.Cache.FaultConfigMaps[fJob.JobUID]
			if !ok {
				faultCm = &plugin.FaultRankIdData{
					Name:      JobFaultRankIDCMPre + fJob.ReferenceName,
					Namespace: fJob.JobNamespace,
				}
			}

			faultCm.Data = map[string]string{
				JobFaultRankIDCMDataKey: jobRankIndexString,
				CmCheckCode:             util.MakeDataHash(jobRankIndex),
			}
			env.Cache.FaultConfigMaps[fJob.JobUID] = faultCm
		}
	}
	return nil
}
