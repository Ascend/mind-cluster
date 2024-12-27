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

	"k8s.io/client-go/kubernetes"
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

func (reCache DealReSchedulerCache) getRecentReschedulingRecordsFromCm(buffer string) (
	map[api.JobID]*RescheduleReason, error) {
	rescheduleRecords := make(map[api.JobID]*RescheduleReason)
	if unmarshalErr := json.Unmarshal([]byte(buffer), &rescheduleRecords); unmarshalErr != nil {
		klog.V(util.LogDebugLev).Info("Unmarshal reschedule records from cache failed")
		return nil, fmt.Errorf("reschedule records convert from CM error: %s", util.SafePrint(unmarshalErr))
	}
	return rescheduleRecords, nil
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

// SetRetryTimesFromCM unmarshal RetryTimes from string into struct and set the value
func (reCache *DealReSchedulerCache) SetRetryTimesFromCM() error {
	if reCache == nil {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM failed: %s, reCache is none", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	data, ok := reCache.CMData[CmJobRemainRetryTimes]
	if !ok {
		return fmt.Errorf("reading %s data from reScheduler configmap failed", CmJobRemainRetryTimes)
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

// SetJobRecentRescheduleRecords get already recorded rescheduling records from cm, and cache it
func (reCache *DealReSchedulerCache) SetJobRecentRescheduleRecords(firstStartup *bool,
	client kubernetes.Interface) error {
	if reCache == nil || reCache.CMData == nil {
		klog.V(util.LogErrorLev).Infof("SetJobRecentRescheduleRecords failed: %s", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	if firstStartup != nil && *firstStartup {
		cm, err := util.GetConfigMap(client, RescheduleReasonCmNamespace, RescheduleReasonCmName)
		if err != nil {
			return fmt.Errorf("failed to get reschedule reason configmap, err: %s", err.Error())
		}
		reCache.CMData[ReschedulingReasonKey] = cm.Data[CmJobRescheduleReasonsKey]
	}

	data, ok := reCache.CMData[ReschedulingReasonKey]
	if !ok {
		// if not initialise now, will give this key an empty content
		return fmt.Errorf("reading %s data from reScheduler configmap failed", CmJobRescheduleReasonsKey)
	}
	if data == "" {
		return nil
	}
	recordedRecords, err := reCache.getRecentReschedulingRecordsFromCm(data)
	if err != nil {
		return fmt.Errorf("getRecentReschedulingRecordsFromCm %s", util.SafePrint(err))
	}
	reCache.JobRecentRescheduleRecords = recordedRecords
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

func (reCache *DealReSchedulerCache) setRealFaultJobs(faultJobs []FaultJob) {
	reCache.RealFaultJobs = faultJobs
}

// getRealFaultJobs only return FaultJobs whose IsFaultJob is true
func (reCache DealReSchedulerCache) getRealFaultJobs() ([]FaultJob, error) {
	realFaultJobs := make([]FaultJob, 0)
	for _, fJob := range reCache.FaultJobs {
		if (!fJob.IsFaultJob && !fJob.IsNormalJobNeedRestart()) || fJob.ReScheduleKey == JobOffRescheduleLabelValue {
			continue // only save real-fault and reschedule-enabled jobs
		}

		faultReason := PodFailed
		for _, faultType := range fJob.FaultTypes {
			if faultType == NodeUnhealthy || faultType == NodeCardUnhealthy || faultType == SubHealthFault {
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

func (reCache *DealReSchedulerCache) writeFaultNodesToCMString() (string, string, error) {
	realFaultNode := reCache.GetRealFaultNodes()
	if len(realFaultNode) == 0 {
		return "", "", nil
	}
	nodeData, err := reCache.marshalCacheDataToString(realFaultNode)
	if err != nil {
		return "", "", fmt.Errorf("writeFaultNodesToCM: %s", util.SafePrint(err))
	}
	nodeDataToCm, marshalErr := reCache.marshalCacheDataToString(getFaultNodeToCm(realFaultNode))
	if marshalErr != nil {
		return "", "", fmt.Errorf("writeFaultNodesToCM: nodeDataToCm failed %s", util.SafePrint(marshalErr))
	}
	return nodeData, nodeDataToCm, nil
}

func getFaultNodeToCm(realFaultNode []FaultNode) []FaultNodeInfoToCm {
	faultNodeToCm := make([]FaultNodeInfoToCm, len(realFaultNode))
	for i, fNode := range realFaultNode {
		faultNodeToCm[i] = initFaultNodeToCmByFaultNode(fNode)
	}
	return faultNodeToCm
}

func initFaultNodeToCmByFaultNode(fNode FaultNode) FaultNodeInfoToCm {
	return FaultNodeInfoToCm{
		FaultDeviceList:     fNode.FaultDeviceList,
		NodeName:            fNode.NodeName,
		UnhealthyNPU:        fNode.UnhealthyNPU,
		NetworkUnhealthyNPU: fNode.NetworkUnhealthyNPU,
		NodeDEnable:         fNode.NodeDEnable,
		NodeHealthState:     fNode.NodeHealthState,
		UpdateTime:          fNode.UpdateTime,
	}
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

func (reCache *DealReSchedulerCache) writeRescheduleReasonsToCMString() (string, error) {
	if len(reCache.JobRecentRescheduleRecords) == 0 {
		return "", nil
	}
	rescheduleReasonStr, err := reCache.marshalCacheDataToString(reCache.JobRecentRescheduleRecords)
	if err != nil {
		return "", fmt.Errorf("writeRescheduleReasonsToCMString: %s", util.SafePrint(err))
	}
	if len(rescheduleReasonStr) > MaxKbOfRescheduleRecords {
		klog.V(util.LogWarningLev).Infof("reason of reschedule into configmap is more than %d Kb, "+
			"will reduce it", MaxKbOfRescheduleRecords)
	}
	// only keep every job newest server record, each time will cut the oldest record of each job
	// to make sure the returned reschedule reason str len is under MaxKbOfRescheduleRecords Kb,
	// each time will reduce the length by 1/10
	// by the way, to avoid dead loop, there is a loop limit
	for i := 0; len(rescheduleReasonStr) > MaxKbOfRescheduleRecords && i < MaxRescheduleRecordsNum; i++ {
		for jobId, reason := range reCache.JobRecentRescheduleRecords {
			// must keep the newest rescheduling record
			if len(reason.RescheduleRecords) <= 1 {
				continue
			}
			lastRecord := reason.RescheduleRecords[len(reason.RescheduleRecords)-1]
			reason.RescheduleRecords = reason.RescheduleRecords[:len(reason.RescheduleRecords)-1]
			klog.V(util.LogWarningLev).Infof("cut job %v reschedule reason of timestamp %d from cm, "+
				"to reduce records length", jobId, lastRecord.RescheduleTimeStamp)
		}
		// to avoid frequently marshal a 950 Kb json, time-consuming
		rescheduleReasonStr, err = reCache.marshalCacheDataToString(reCache.JobRecentRescheduleRecords)
		if err != nil {
			return "", fmt.Errorf("writeRescheduleReasonsToCMString: %s", util.SafePrint(err))
		}
		if len(rescheduleReasonStr) <= MaxKbOfRescheduleRecords {
			break
		}
	}
	return rescheduleReasonStr, nil
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

// WriteReSchedulerCacheToEnvCache write the modifications on cache data to env to update re-scheduling configmap
func (reCache *DealReSchedulerCache) WriteReSchedulerCacheToEnvCache(env *plugin.ScheduleEnv, jobType string) error {
	if reCache == nil || env == nil {
		return errors.New(util.ArgumentError)
	}
	env.Cache.Names[RePropertyName] = CmName
	env.Cache.Namespaces[RePropertyName] = CmNameSpace
	fNodeString, fNodeToCMString, err := reCache.writeFaultNodesToCMString()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("WriteReSchedulerCacheToEnvCache: %s", util.SafePrint(err))
	}
	fJobString, err := reCache.writeFaultJobsToCMString()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("WriteReSchedulerCacheToEnvCache: %s", util.SafePrint(err))
	}

	jobRemainRetryTimes, err := reCache.writeRemainTimesToCMString()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("WriteReSchedulerCacheToEnvCache: %s", util.SafePrint(err))
	}

	nodeRankOccurrenceMapString, err := reCache.writeNodeRankOccurrenceMapToCMString()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("WriteReSchedulerCacheToEnvCache: %s", util.SafePrint(err))
	}
	// update the reschedule reason cache
	jobRescheduleReasons, err := reCache.setRescheduleReasonToCache(env)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("setRescheduleReasonToCache: %s", util.SafePrint(err))
	}

	cmData, ok := env.Cache.Data[RePropertyName]
	if !ok {
		cmData = make(map[string]string, util.MapInitNum)
		env.Cache.Data[RePropertyName] = cmData
	}

	cmData[CmJobRemainRetryTimes] = jobRemainRetryTimes
	cmDataForCache := util.DeepCopyCmData(cmData)
	cmData[CmFaultNodeKind] = fNodeToCMString
	cmDataForCache[jobType] = fJobString
	cmDataForCache[CmFaultNodeKind] = fNodeString
	cmDataForCache[CmNodeRankTimeMapKind] = nodeRankOccurrenceMapString
	cmDataForCache[ReschedulingReasonKey] = jobRescheduleReasons
	reSchedulerConfigmap.updateReSchedulerCMCache(cmDataForCache)
	return nil
}

func (reCache *DealReSchedulerCache) setRescheduleReasonToCache(env *plugin.ScheduleEnv) (string, error) {
	env.Cache.Names[ReschedulingReasonKey] = RescheduleReasonCmName
	env.Cache.Namespaces[ReschedulingReasonKey] = RescheduleReasonCmNamespace
	jobRescheduleReasons, err := reCache.writeRescheduleReasonsToCMString()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("writeRescheduleReasonsToCMString: %s", util.SafePrint(err))
		return "", fmt.Errorf("writeRescheduleReasonsToCMString: %s", util.SafePrint(err))
	}
	reasonCmData, ok := env.Cache.Data[ReschedulingReasonKey]
	if !ok {
		reasonCmData = make(map[string]string, util.MapInitNum)
		env.Cache.Data[ReschedulingReasonKey] = reasonCmData
	}
	reasonCmData[CmJobRescheduleReasonsKey] = jobRescheduleReasons
	return jobRescheduleReasons, nil
}
