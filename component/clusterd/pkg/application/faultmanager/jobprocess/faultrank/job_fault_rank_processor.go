// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package faultrank contain job fault rank process
package faultrank

import (
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
	nodeDeviceInfoMap map[string]constant.AdvanceDeviceFaultCm,
	nodeName string, serverList map[string]constant.ServerHccl, jobId string) []constant.FaultRank {
	advanceDeviceInfo := nodeDeviceInfoMap[nodeName]
	devicesOfJobOnNode, ok := serverList[nodeName]
	faultRankList := make([]constant.FaultRank, 0)
	if !ok || len(devicesOfJobOnNode.DeviceList) == 0 {
		return faultRankList
	}
	for _, deviceInfo := range devicesOfJobOnNode.DeviceList {
		deviceName := advanceDeviceInfo.ServerType + "-" + deviceInfo.DeviceID
		faultList := advanceDeviceInfo.FaultDeviceList[deviceName]
		podRank, podUid := pod.GetPodRankAndPodUid(jobId, deviceInfo.RankID)
		uceInManagementPlane := false
		// scan management plane fault info. management plane may filter uce fault in uceProcessor
		for _, fault := range faultList {
			faultRank := constant.FaultRank{
				RankId:      deviceInfo.RankID,
				PodUid:      podUid,
				PodRank:     podRank,
				FaultCode:   fault.FaultCode,
				FaultLevel:  fault.FaultLevel,
				DoStepRetry: false,
				DeviceId:    deviceInfo.DeviceID,
			}
			if strings.Contains(fault.FaultCode, constant.UceFaultCode) {
				// management plane find uce fault
				uceInManagementPlane = true
				faultRank.DoStepRetry = processor.canDoStepRetry(jobId, nodeName, deviceName)
			}
			faultRankList = append(faultRankList, faultRank)
		}
		if uceInManagementPlane {
			continue
		}
		// business plane find uce fault
		if processor.uceInBusinessPlane(jobId, nodeName, deviceName) {
			faultRankList = append(faultRankList, constant.FaultRank{
				RankId:      deviceInfo.RankID,
				PodUid:      podUid,
				PodRank:     podRank,
				FaultCode:   constant.UceFaultCode,
				FaultLevel:  constant.RestartBusiness,
				DoStepRetry: processor.canDoStepRetry(jobId, nodeName, deviceName),
				DeviceId:    deviceInfo.DeviceID,
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
		faultList, nodeStatusList, faultDeviceList := processor.findNodeDeviceAndSwitchFault(serverList,
			nodeInfos, switchInfos, deviceCmForNodeMap, jobId)
		jobFaultInfo.FaultList = faultList
		if len(jobFaultInfo.FaultList) > 0 {
			hwlog.RunLog.Debugf("jobFaultInfo: %#v", jobFaultInfo)
		}
		podStrategiesMap := relationfault.RelationProcessor.GetPodStrategiesMapsByJobId(jobId)
		hwlog.RunLog.Debugf("jobId=%s, faultRank=%v, nodeStatus=%v, podStrategiesMap=%v",
			jobId, faultList, nodeStatusList, podStrategiesMap)
		jobFaultInfo.HealthyState = getHealthState(faultList, nodeStatusList, podStrategiesMap)
		jobFaultInfo.FaultDevice = faultDeviceList
		jobFaultInfos[jobId] = jobFaultInfo
	}
	processor.setJobFaultRankInfos(jobFaultInfos)
	return nil
}

func (processor *jobRankFaultInfoProcessor) findNodeDeviceAndSwitchFault(
	serverList map[string]constant.ServerHccl, nodeInfos map[string]*constant.NodeInfo,
	switchInfos map[string]*constant.SwitchInfo, deviceCmForNodeMap map[string]constant.AdvanceDeviceFaultCm,
	jobId string) ([]constant.FaultRank, []string, []constant.FaultDevice) {
	faultList := make([]constant.FaultRank, 0)
	faultDeviceList := make([]constant.FaultDevice, 0)
	nodeStatusList := make([]string, 0)
	for nodeName, server := range serverList {
		hwlog.RunLog.Debugf("nodeName: %s, server: %#v", nodeName, server)
		switchInfo, ok := switchInfos[constant.SwitchInfoPrefix+nodeName]
		if ok {
			nodeStatusList = append(nodeStatusList, switchInfo.NodeStatus)
		}
		faultDeviceList = append(faultDeviceList, getFaultDeviceInfoBySwitchInfo(&server, switchInfo)...)
		if ok && switchInfo.NodeStatus == constant.UnHealthyState {
			hwlog.RunLog.Debugf("node %s switch is unhealthy", nodeName)
			faultCode := strings.Join(switchInfo.FaultCode, constant.Comma)
			faultList = append(faultList, serverHcclToFaultRank(server, jobId, faultCode)...)
		}
		nodeInfo, ok := nodeInfos[constant.NodeInfoPrefix+nodeName]
		if ok && nodeInfo.NodeStatus == constant.UnHealthyState {
			hwlog.RunLog.Debugf("node %s is unhealthy", nodeName)
			faultCode := strings.Join(getFaultCodeByNodeInfo(nodeInfo), constant.Comma)
			faultList = append(faultList, serverHcclToFaultRank(server, jobId, faultCode)...)
		}
		faultDeviceList = append(faultDeviceList, getFaultDeviceInfoByNodeInfo(&server, nodeInfo)...)
		node := kube.GetNode(nodeName)
		if node == nil || !faultdomain.IsNodeReady(node) {
			hwlog.RunLog.Debugf("node %s is not ready", nodeName)
			faultList = append(faultList, serverHcclToFaultRank(server, jobId, "")...)
			faultDeviceList = append(faultDeviceList, convertToFaultDevice(&server, "",
				constant.SeparateNPU, constant.EmptyDeviceId, constant.FaultTypeNode))
		}
		faultRankList := processor.findFaultRankForJob(deviceCmForNodeMap, nodeName, serverList, jobId)
		faultList = append(faultList, faultRankList...)
		faultDeviceList = append(faultDeviceList, getFautDeviceInfoByFaultRank(&server, faultRankList)...)
	}
	return faultList, nodeStatusList, faultDeviceList
}

func getFautDeviceInfoByFaultRank(server *constant.ServerHccl,
	faultRankList []constant.FaultRank) []constant.FaultDevice {
	if len(faultRankList) == 0 {
		return nil
	}
	faultList := make([]constant.FaultDevice, 0)
	for _, faultRank := range faultRankList {
		faultList = append(faultList, convertToFaultDevice(server, faultRank.FaultCode, faultRank.FaultLevel,
			faultRank.DeviceId, constant.FaultTypeNPU))
	}
	return faultList
}

func getFaultDeviceInfoByNodeInfo(server *constant.ServerHccl, nodeInfo *constant.NodeInfo) []constant.FaultDevice {
	if nodeInfo == nil {
		return nil
	}
	faultList := make([]constant.FaultDevice, 0)
	for _, faultDev := range nodeInfo.FaultDevList {
		for _, faultCode := range faultDev.FaultCode {
			deviceId := strconv.FormatInt(faultDev.DeviceId, constant.FormatBase)
			faultList = append(faultList, convertToFaultDevice(server, faultCode, faultDev.FaultLevel,
				deviceId, faultDev.DeviceType))
		}
	}
	return faultList
}

func getFaultDeviceInfoBySwitchInfo(server *constant.ServerHccl,
	switchInfo *constant.SwitchInfo) []constant.FaultDevice {
	if switchInfo == nil {
		return nil
	}
	faultList := make([]constant.FaultDevice, 0)
	for _, faultCode := range switchInfo.SwitchFaultInfo.FaultCode {
		faultList = append(faultList, convertToFaultDevice(server, faultCode, switchInfo.SwitchFaultInfo.FaultLevel,
			constant.EmptyDeviceId, constant.FaultTypeSwitch))
	}
	return faultList
}

func convertToFaultDevice(server *constant.ServerHccl, faultCode,
	faultLevel, deviceId, deviceType string) constant.FaultDevice {
	return constant.FaultDevice{
		ServerName: server.ServerName,
		ServerSN:   server.ServerSN,
		ServerId:   server.ServerID,
		DeviceId:   deviceId,
		FaultCode:  faultCode,
		FaultLevel: faultLevel,
		DeviceType: deviceType,
	}
}

func getFaultCodeByNodeInfo(nodeInfo *constant.NodeInfo) []string {
	if nodeInfo == nil {
		return nil
	}
	faultCodes := make([]string, 0)
	for _, faultDev := range nodeInfo.FaultDevList {
		faultCodes = append(faultCodes, faultDev.FaultCode...)
	}
	return util.RemoveSliceDuplicateElement(faultCodes)
}

func serverHcclToFaultRank(server constant.ServerHccl, jobId, faultCode string) []constant.FaultRank {
	faultRanks := make([]constant.FaultRank, 0, len(server.DeviceList))
	for _, device := range server.DeviceList {
		podRank, podUid := pod.GetPodRankAndPodUid(jobId, device.RankID)
		faultRanks = append(faultRanks, constant.FaultRank{
			RankId:      device.RankID,
			PodUid:      podUid,
			PodRank:     podRank,
			FaultCode:   faultCode,
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
