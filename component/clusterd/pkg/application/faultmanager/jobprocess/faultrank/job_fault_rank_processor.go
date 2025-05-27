// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package faultrank contain job fault rank process
package faultrank

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/cmprocess/uce"
	"clusterd/pkg/application/faultmanager/jobprocess/relationfault"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/interface/kube"
)

// JobFaultRankProcessor process job fault rank
var JobFaultRankProcessor *jobRankFaultInfoProcessor

func init() {
	JobFaultRankProcessor = &jobRankFaultInfoProcessor{
		jobFaultInfoMap: make(map[string]constant.JobFaultInfo),
		mutex:           sync.RWMutex{},
	}
}

type jobRankFaultInfoProcessor struct {
	jobFaultInfoMap map[string]constant.JobFaultInfo
	mutex           sync.RWMutex
}

type jobPodInfoMap struct {
	podOfRank      map[int]*constant.SimplePodInfo
	deviceNumOfPod int
	jobId          string
}

func getJobPodInfoMap(jobId string) *jobPodInfoMap {
	return &jobPodInfoMap{
		podOfRank:      podInfoByPodRank(pod.GetSimplePodByJobId(jobId)),
		deviceNumOfPod: pod.GetPodDeviceNumByJobId(jobId),
		jobId:          jobId,
	}
}

func (m *jobPodInfoMap) getPodUidAndRankByCardRank(cardRankStr string) (string, string, error) {
	if m == nil || m.deviceNumOfPod <= 0 {
		return "", "", fmt.Errorf("invalid jobPodInfoMap")
	}
	cardRank, err := strconv.Atoi(cardRankStr)
	if err != nil {
		return "", "", fmt.Errorf("convert %s err: %v", cardRankStr, err)
	}
	podRank := cardRank / m.deviceNumOfPod
	podRankStr := strconv.Itoa(podRank)
	if info, ok := m.podOfRank[podRank]; ok {
		return info.PodUid, podRankStr, nil
	}
	return "", "", fmt.Errorf("cardRank %s has no podRank", cardRankStr)
}

// GetJobFaultRankInfos get job fault rank information
func (processor *jobRankFaultInfoProcessor) GetJobFaultRankInfos() map[string]constant.JobFaultInfo {
	processor.mutex.RLock()
	defer processor.mutex.RUnlock()
	result := new(map[string]constant.JobFaultInfo)
	err := util.DeepCopy(result, processor.jobFaultInfoMap)
	if err != nil {
		hwlog.RunLog.Errorf("get job fault rank failed, err: %v", err)
		return nil
	}
	hwlog.RunLog.Debugf("get job fault rank: %v", util.ObjToString(*result))
	return *result
}

// GetJobFaultRankInfosFilterLevel query jobs fault rank info, and filter fault below `faultLevel`
func (processor *jobRankFaultInfoProcessor) GetJobFaultRankInfosFilterLevel(
	faultLevel string) map[string]constant.JobFaultInfo {
	jobFaultRankInfos := processor.GetJobFaultRankInfos()
	if jobFaultRankInfos == nil {
		return nil
	}
	for jobId, jobFaultInfo := range jobFaultRankInfos {
		faultList := make([]constant.FaultRank, 0)
		for _, fault := range jobFaultInfo.FaultList {
			if fault.FaultLevel != faultLevel {
				faultList = append(faultList, fault)
			}
		}
		jobFaultInfo.FaultList = faultList
		jobFaultRankInfos[jobId] = jobFaultInfo
	}
	return jobFaultRankInfos
}

func (processor *jobRankFaultInfoProcessor) setJobFaultRankInfos(faultInfos map[string]constant.JobFaultInfo) {
	processor.mutex.Lock()
	defer processor.mutex.Unlock()
	processor.jobFaultInfoMap = faultInfos
}

func (processor *jobRankFaultInfoProcessor) findFaultRankForJob(
	nodeDeviceInfoMap map[string]constant.AdvanceDeviceFaultCm, nodeName string,
	serverList map[string]constant.ServerHccl, jobId string, podInfo *jobPodInfoMap) []constant.FaultRank {
	advanceDeviceInfo := nodeDeviceInfoMap[nodeName]
	devicesOfJobOnNode, ok := serverList[nodeName]
	faultRankList := make([]constant.FaultRank, 0)
	if !ok || len(devicesOfJobOnNode.DeviceList) == 0 {
		return faultRankList
	}
	for _, deviceInfo := range devicesOfJobOnNode.DeviceList {
		deviceName := advanceDeviceInfo.ServerType + "-" + deviceInfo.DeviceID
		faultList := advanceDeviceInfo.FaultDeviceList[deviceName]
		uceInManagementPlane := false
		for _, fault := range faultList {
			if podInfo == nil {
				podInfo = getJobPodInfoMap(jobId)
			}
			podUid, podRankStr, err := podInfo.getPodUidAndRankByCardRank(deviceInfo.RankID)
			if err != nil {
				hwlog.RunLog.Errorf("device %s's rank id is %s, getPodUidAndRankByCardRank err: %v",
					deviceInfo.DeviceIP, deviceInfo.RankID, err)
			}
			faultRank := constant.FaultRank{RankId: deviceInfo.RankID, PodUid: podUid, PodRank: podRankStr,
				FaultCode: fault.FaultCode, FaultLevel: fault.FaultLevel, DoStepRetry: false,
			}
			if strings.Contains(fault.FaultCode, constant.UceFaultCode) {
				uceInManagementPlane = true
				faultRank.DoStepRetry = processor.canDoStepRetry(jobId, nodeName, deviceName)
			}
			faultRankList = append(faultRankList, faultRank)
		}
		if uceInManagementPlane {
			continue
		}
		if processor.uceInBusinessPlane(jobId, nodeName, deviceName) {
			if podInfo == nil {
				podInfo = getJobPodInfoMap(jobId)
			}
			podUid, podRankStr, err := podInfo.getPodUidAndRankByCardRank(deviceInfo.RankID)
			if err != nil {
				hwlog.RunLog.Errorf("device %s's rank id is %s, getPodUidAndRankByCardRank err: %v",
					deviceInfo.DeviceIP, deviceInfo.RankID, err)
			}
			faultRankList = append(faultRankList, constant.FaultRank{RankId: deviceInfo.RankID, PodUid: podUid,
				PodRank: podRankStr, FaultCode: constant.UceFaultCode, FaultLevel: constant.RestartBusiness,
				DoStepRetry: processor.canDoStepRetry(jobId, nodeName, deviceName),
			})
		}
	}
	return faultRankList
}

func (processor *jobRankFaultInfoProcessor) canDoStepRetry(jobId, nodeName, deviceName string) bool {
	uceDevice, found := uce.UceProcessor.GetUceDeviceFromJob(jobId, nodeName, deviceName)
	if !found {
		hwlog.RunLog.Debugf("job %s's uce fault is not on node %s device %s", jobId, nodeName, deviceName)
		return false
	}
	doStepRetry := faultdomain.CanDoStepRetry(&uceDevice)
	hwlog.RunLog.Debugf("uceDevice %s stepretry %v", util.ObjToString(uceDevice), doStepRetry)
	return doStepRetry
}

func (processor *jobRankFaultInfoProcessor) uceInBusinessPlane(jobId, nodeName, deviceName string) bool {
	uceDevice, found := uce.UceProcessor.GetUceDeviceFromJob(jobId, nodeName, deviceName)
	// business plane didn't find uce fault
	if !found {
		hwlog.RunLog.Debugf("business plane didn't find uce fault")
		return false
	}
	// business plane found uce fault
	result := faultdomain.ValidBusinessRecoverTime(uceDevice.RecoverTime)
	if !result {
		hwlog.RunLog.Debugf("invalid BusinessRecoverTime %v", uceDevice)
	}
	return result
}

// Process job fault rank info
func (processor *jobRankFaultInfoProcessor) Process(info any) any {
	allConfigmap, ok := info.(constant.AllConfigmapContent)
	if !ok {
		hwlog.RunLog.Error("convert info to AllConfigmapContent failed")
		return info
	}
	deviceInfos := allConfigmap.DeviceCm
	deviceCmForNodeMap := faultdomain.GetAdvanceDeviceCmForNodeMap(deviceInfos)
	hwlog.RunLog.Debugf("deviceInfos: %#v", deviceInfos)
	nodeInfos := allConfigmap.NodeCm
	hwlog.RunLog.Debugf("nodeInfos: %#v", nodeInfos)
	switchInfos := allConfigmap.SwitchCm
	hwlog.RunLog.Debugf("switchInfos: %#v", switchInfos)

	jobFaultInfos := make(map[string]constant.JobFaultInfo)
	jobServerInfoMap := job.GetJobServerInfoMap()
	for jobId, serverList := range jobServerInfoMap.InfoMap {
		jobFaultInfo := constant.JobFaultInfo{
			JobId:        jobId,
			FaultList:    make([]constant.FaultRank, 0),
			HealthyState: constant.HealthyState,
		}
		hwlog.RunLog.Debugf("serverList: %d", len(serverList))
		faultList, nodeStatusList := processor.findNodeDeviceAndSwitchFault(serverList,
			nodeInfos, switchInfos, deviceCmForNodeMap, jobId)
		jobFaultInfo.FaultList = faultList
		if len(jobFaultInfo.FaultList) > 0 {
			hwlog.RunLog.Debugf("jobFaultInfo: %#v", jobFaultInfo)
		}
		podStrategiesMap := relationfault.RelationProcessor.GetPodStrategiesMapsByJobId(jobId)
		hwlog.RunLog.Debugf("jobId=%s, faultRank=%v, nodeStatus=%v, podStrategiesMap=%v",
			jobId, faultList, nodeStatusList, podStrategiesMap)
		jobFaultInfo.HealthyState = getHealthState(faultList, nodeStatusList, podStrategiesMap)
		jobFaultInfos[jobId] = jobFaultInfo
	}
	processor.setJobFaultRankInfos(jobFaultInfos)
	return nil
}

func (processor *jobRankFaultInfoProcessor) findNodeDeviceAndSwitchFault(
	serverList map[string]constant.ServerHccl, nodeInfos map[string]*constant.NodeInfo,
	switchInfos map[string]*constant.SwitchInfo, deviceCmForNodeMap map[string]constant.AdvanceDeviceFaultCm,
	jobId string) ([]constant.FaultRank, []string) {
	faultList := make([]constant.FaultRank, 0)
	nodeStatusList := make([]string, 0)
	var info *jobPodInfoMap = nil
	var initJobPodInfoMap = sync.Once{}

	for nodeName, server := range serverList {
		hwlog.RunLog.Debugf("nodeName: %s, server: %#v", nodeName, server)
		switchInfo, ok := switchInfos[constant.SwitchInfoPrefix+nodeName]
		if ok {
			nodeStatusList = append(nodeStatusList, switchInfo.NodeStatus)
		}
		var hcclToFaultRank []constant.FaultRank = nil
		var initHcclToFaultRank = sync.Once{}
		if ok && switchInfo.NodeStatus == constant.UnHealthyState {
			hwlog.RunLog.Debugf("node %s switch is unhealthy", nodeName)
			initJobPodInfoMap.Do(func() { info = getJobPodInfoMap(jobId) })
			initHcclToFaultRank.Do(func() { hcclToFaultRank = serverHcclToFaultRank(server, info) })
			faultList = append(faultList, hcclToFaultRank...)
			continue
		}
		nodeInfo, ok := nodeInfos[constant.NodeInfoPrefix+nodeName]
		if ok && nodeInfo.NodeStatus == constant.UnHealthyState {
			hwlog.RunLog.Debugf("node %s is unhealthy", nodeName)
			initJobPodInfoMap.Do(func() { info = getJobPodInfoMap(jobId) })
			initHcclToFaultRank.Do(func() { hcclToFaultRank = serverHcclToFaultRank(server, info) })
			faultList = append(faultList, hcclToFaultRank...)
			continue
		}
		node := kube.GetNode(nodeName)
		if node == nil || !faultdomain.IsNodeReady(node) {
			hwlog.RunLog.Debugf("node %s is not ready", nodeName)
			initJobPodInfoMap.Do(func() { info = getJobPodInfoMap(jobId) })
			initHcclToFaultRank.Do(func() { hcclToFaultRank = serverHcclToFaultRank(server, info) })
			faultList = append(faultList, hcclToFaultRank...)
			continue
		}
		faultRankList := processor.findFaultRankForJob(deviceCmForNodeMap, nodeName, serverList, jobId, info)
		faultList = append(faultList, faultRankList...)
	}
	return faultList, nodeStatusList
}

func podInfoByPodRank(podInfos map[string]*constant.SimplePodInfo) map[int]*constant.SimplePodInfo {
	result := make(map[int]*constant.SimplePodInfo)
	for _, podInfo := range podInfos {
		result[podInfo.PodRank] = podInfo
	}
	return result
}

func serverHcclToFaultRank(server constant.ServerHccl, podInfos *jobPodInfoMap) []constant.FaultRank {
	faultRanks := make([]constant.FaultRank, 0, len(server.DeviceList))
	for _, device := range server.DeviceList {
		podUid, podRank, err := podInfos.getPodUidAndRankByCardRank(device.RankID)
		if err != nil {
			hwlog.RunLog.Errorf("device %s's rank id is %s getPodUidAndRankByCardRank err: %v",
				device.DeviceIP, device.RankID, err)
			continue
		}
		faultRanks = append(faultRanks, constant.FaultRank{
			RankId:      device.RankID,
			PodUid:      podUid,
			PodRank:     podRank,
			FaultCode:   "",
			FaultLevel:  constant.SeparateNPU,
			DoStepRetry: false,
		})
	}
	return faultRanks
}

func getHealthState(faultList []constant.FaultRank, nodeStatusList []string,
	podStrategiesMap map[string]string) string {
	hasSubHealthFault := false
	for _, faultRank := range faultList {
		if faultRank.FaultLevel != constant.SubHealthFault && faultRank.FaultLevel != constant.NotHandleFault {
			return constant.UnHealthyState
		}
		if faultRank.FaultLevel == constant.SubHealthFault {
			hasSubHealthFault = true
		}
	}
	for _, status := range nodeStatusList {
		if status == constant.UnHealthyState {
			return constant.UnHealthyState
		}
		if status == constant.SubHealthyState {
			hasSubHealthFault = true
		}
	}
	for _, strategy := range podStrategiesMap {
		if strategy == constant.SeparateFaultStrategy {
			return constant.UnHealthyState
		}
		if strategy == constant.SubHealthFaultStrategy {
			hasSubHealthFault = true
		}
	}
	if hasSubHealthFault {
		return constant.SubHealthyState
	}
	return constant.HealthyState
}
