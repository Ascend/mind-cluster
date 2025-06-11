// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package faultrank contain job fault rank process
package faultrank

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/cmprocess/retry"
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
	podOfRank      map[string]*constant.SimplePodInfo
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

func podInfoByPodRank(podInfos map[string]*constant.SimplePodInfo) map[string]*constant.SimplePodInfo {
	result := make(map[string]*constant.SimplePodInfo)
	for _, podInfo := range podInfos {
		result[podInfo.PodRank] = podInfo
	}
	return result
}

func (m *jobPodInfoMap) getPodUidAndRankByCardRank(cardRankStr string) (string, string, error) {
	if m == nil || m.deviceNumOfPod <= 0 {
		return "", "", nil
	}
	cardRank, err := strconv.Atoi(cardRankStr)
	if err != nil {
		return "", "", fmt.Errorf("convert %s err: %v", cardRankStr, err)
	}
	podRank := cardRank / m.deviceNumOfPod
	podRankStr := strconv.Itoa(podRank)
	if info, ok := m.podOfRank[podRankStr]; ok {
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
	faultLevels []string) map[string]constant.JobFaultInfo {
	jobFaultRankInfos := processor.GetJobFaultRankInfos()
	if jobFaultRankInfos == nil {
		return nil
	}
	for jobId, jobFaultInfo := range jobFaultRankInfos {
		faultList := make([]constant.FaultRank, 0)
		for _, fault := range jobFaultInfo.FaultList {
			if !slices.Contains(faultLevels, fault.FaultLevel) {
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
	advanceDeviceInfo *constant.AdvanceDeviceFaultCm, nodeName string,
	serverList map[string]constant.ServerHccl, podInfo *jobPodInfoMap) []constant.FaultRank {
	devicesOfJobOnNode, ok := serverList[nodeName]
	if advanceDeviceInfo == nil || !ok || len(devicesOfJobOnNode.DeviceList) == 0 {
		return make([]constant.FaultRank, 0)
	}
	faultRankList := make([]constant.FaultRank, 0)
	for _, deviceInfo := range devicesOfJobOnNode.DeviceList {
		deviceName := advanceDeviceInfo.DeviceType + "-" + deviceInfo.DeviceID
		faultList := advanceDeviceInfo.FaultDeviceList[deviceName]
		faultList = processor.appendFilterFaultCodeAndLevel(podInfo.jobId, nodeName, deviceName, faultList)
		retryInManagementPlane := false
		podUid, podRankStr, err := podInfo.getPodUidAndRankByCardRank(deviceInfo.RankID)
		if err != nil {
			hwlog.RunLog.Errorf("device %s's rank id is %s, getPodUidAndRankByCardRank err: %v",
				deviceInfo.DeviceIP, deviceInfo.RankID, err)
		}
		// scan management plane fault info. management plane may filter uce fault in uceProcessor、hcclProcessor
		for _, fault := range faultList {
			restartInPlace := faultdomain.IsL2L3Fault(fault.FaultLevel) && processor.canDoRestartInPlace(podInfo.jobId)
			faultRank := constant.FaultRank{RankId: deviceInfo.RankID, PodUid: podUid, PodRank: podRankStr,
				FaultCode: fault.FaultCode, FaultLevel: fault.FaultLevel, DoStepRetry: false,
				DoRestartInPlace: restartInPlace, DeviceId: deviceInfo.DeviceID,
			}
			if faultdomain.IsUceFault(fault.FaultCode) || faultdomain.IsHcclRetryFault(fault.FaultCode) {
				retryInManagementPlane = true
				faultRank.DoStepRetry = processor.canDoStepRetry(podInfo.jobId, nodeName, deviceName)
			}
			faultRankList = append(faultRankList, faultRank)
		}
		if retryInManagementPlane {
			continue
		}
		// business plane find uce fault
		if deviceDetail, ok := processor.retryInBusinessPlane(podInfo.jobId, nodeName, deviceName); ok {
			faultRankList = append(faultRankList, constant.FaultRank{RankId: deviceInfo.RankID, PodUid: podUid,
				PodRank: podRankStr, FaultCode: faultdomain.GetRetryCodeByFaultType(deviceDetail.FaultType),
				FaultLevel:  constant.RestartBusiness,
				DoStepRetry: processor.canDoStepRetry(podInfo.jobId, nodeName, deviceName),
				DeviceId:    deviceInfo.DeviceID,
			})
		}
	}
	return faultRankList
}

func (processor *jobRankFaultInfoProcessor) appendFilterFaultCodeAndLevel(jobId, nodeName, deviceName string,
	faultList []constant.DeviceFault) []constant.DeviceFault {
	filterFault := retry.RetryProcessor.GetFilterFaultCodeAndLevel(jobId, nodeName, deviceName)
	if len(filterFault) == 0 {
		return faultList
	}
	newFaultList := make([]constant.DeviceFault, 0, len(faultList)+len(filterFault))
	for faultCode, faultLevel := range filterFault {
		if faultdomain.IsUceFault(faultCode) || faultdomain.IsHcclRetryFault(faultCode) {
			continue
		}
		found := false
		for _, fault := range faultList {
			if fault.FaultCode == faultCode {
				found = true
				break
			}
		}
		if !found {
			newFaultList = append(newFaultList, constant.DeviceFault{FaultCode: faultCode, FaultLevel: faultLevel})
		}
	}
	newFaultList = append(newFaultList, faultList...)
	hwlog.RunLog.Debugf("jobid: %s node:%s device:%s fault list:%v, new fault list: %v",
		jobId, nodeName, deviceName, faultList, newFaultList)
	return newFaultList
}

func (processor *jobRankFaultInfoProcessor) canDoStepRetry(jobId, nodeName, deviceName string) bool {
	device, found := retry.RetryProcessor.GetRetryDeviceFromJob(jobId, nodeName, deviceName)
	if !found {
		hwlog.RunLog.Debugf("job %s's uce fault is not on node %s device %s", jobId, nodeName, deviceName)
		return false
	}
	detailInfo, ok := device.FaultDetail[constant.DeviceRetryFault]
	if !ok {
		hwlog.RunLog.Debugf("job %s's uce fault is not on node %s device %s", jobId, nodeName, deviceName)
		return false
	}
	doStepRetry := faultdomain.CanDoStepRetry(&detailInfo)
	hwlog.RunLog.Debugf("device %s stepretry %v", util.ObjToString(device), doStepRetry)
	return doStepRetry
}

func (processor *jobRankFaultInfoProcessor) canDoRestartInPlace(jobId string) bool {
	return retry.RetryProcessor.CanDoRestartInPlace(jobId)
}

func (processor *jobRankFaultInfoProcessor) retryInBusinessPlane(jobId, nodeName,
	deviceName string) (constant.DeviceFaultDetail, bool) {
	retryDevice, found := retry.RetryProcessor.GetRetryDeviceFromJob(jobId, nodeName, deviceName)
	// business plane didn't find retry fault
	if !found {
		hwlog.RunLog.Debugf("business plane didn't find retry fault")
		return constant.DeviceFaultDetail{}, false
	}
	detailInfo, ok := retryDevice.FaultDetail[constant.DeviceRetryFault]
	if !ok {
		hwlog.RunLog.Debugf("business plane didn't find retry fault")
		return constant.DeviceFaultDetail{}, false
	}
	// business plane found retry fault
	result := faultdomain.ValidBusinessRecoverTime(detailInfo.RecoverTime)
	if !result {
		hwlog.RunLog.Debugf("invalid BusinessRecoverTime %v", retryDevice)
	}
	return detailInfo, result
}

// Process job fault rank info
func (processor *jobRankFaultInfoProcessor) Process(info any) any {
	allConfigmap, ok := info.(constant.AllConfigmapContent)
	if !ok {
		hwlog.RunLog.Error("convert info to AllConfigmapContent failed")
		return info
	}
	hwlog.RunLog.Debugf("allConfigmap info: %#v", util.ObjToString(allConfigmap))

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
			allConfigmap.NodeCm, allConfigmap.SwitchCm, allConfigmap.DeviceCm, jobId)
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
	switchInfos map[string]*constant.SwitchInfo, deviceCmForNodeMap map[string]*constant.AdvanceDeviceFaultCm,
	jobId string) ([]constant.FaultRank, []string, []constant.FaultDevice) {
	faultList := make([]constant.FaultRank, 0)
	faultDeviceList := make([]constant.FaultDevice, 0)
	nodeStatusList := make([]string, 0)
	info := getJobPodInfoMap(jobId)

	for nodeName, server := range serverList {
		hwlog.RunLog.Debugf("nodeName: %s, server: %#v", nodeName, server)
		switchInfo, ok := switchInfos[constant.SwitchInfoPrefix+nodeName]
		if ok {
			nodeStatusList = append(nodeStatusList, switchInfo.NodeStatus)
		}
		faultDeviceList = append(faultDeviceList, getFaultDeviceInfoBySwitchInfo(&server, switchInfo)...)
		if ok && switchInfo.NodeStatus == constant.UnHealthyState {
			hwlog.RunLog.Debugf("node %s switch is unhealthy", nodeName)
			faultCode := strings.Join(getFaultCodeBySwitchInfo(switchInfo), constant.Comma)
			faultList = append(faultList, serverHcclToFaultRank(server, info, faultCode)...)
		}
		nodeInfo, ok := nodeInfos[constant.NodeInfoPrefix+nodeName]
		if ok && nodeInfo.NodeStatus == constant.UnHealthyState {
			hwlog.RunLog.Debugf("node %s is unhealthy", nodeName)
			faultCode := strings.Join(getFaultCodeByNodeInfo(nodeInfo), constant.Comma)
			faultList = append(faultList, serverHcclToFaultRank(server, info, faultCode)...)
		}
		faultDeviceList = append(faultDeviceList, getFaultDeviceInfoByNodeInfo(&server, nodeInfo)...)
		node := kube.GetNode(nodeName)
		if node == nil || !faultdomain.IsNodeReady(node) {
			hwlog.RunLog.Debugf("node %s is not ready", nodeName)
			faultList = append(faultList, serverHcclToFaultRank(server, info, "")...)
			faultDeviceList = append(faultDeviceList, convertToFaultDevice(&server, "",
				constant.SeparateNPU, constant.EmptyDeviceId, constant.FaultTypeNode))
		}
		advanceDeviceInfo := deviceCmForNodeMap[nodeName]
		faultRankList := processor.findFaultRankForJob(advanceDeviceInfo, nodeName, serverList, info)
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
	for _, faultInfo := range switchInfo.SwitchFaultInfo.FaultInfo {
		faultList = append(faultList, convertToFaultDevice(server, faultInfo.AssembledFaultCode,
			switchInfo.SwitchFaultInfo.FaultLevel, constant.EmptyDeviceId, constant.FaultTypeSwitch))
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

func getFaultCodeBySwitchInfo(switchInfo *constant.SwitchInfo) []string {
	if switchInfo == nil {
		return nil
	}
	faultCodes := make([]string, 0)
	for _, faultInfo := range switchInfo.FaultInfo {
		faultCodes = append(faultCodes, faultInfo.AssembledFaultCode)
	}
	return faultCodes
}

func serverHcclToFaultRank(server constant.ServerHccl, podInfos *jobPodInfoMap, faultCode string) []constant.FaultRank {
	faultRanks := make([]constant.FaultRank, 0, len(server.DeviceList))
	for _, device := range server.DeviceList {
		podUid, podRank, err := podInfos.getPodUidAndRankByCardRank(device.RankID)
		if err != nil {
			hwlog.RunLog.Errorf("device %s's rank id is %s getPodUidAndRankByCardRank err: %v",
				device.DeviceIP, device.RankID, err)
		}
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
		if faultRank.FaultLevel != constant.SubHealthFault &&
			faultRank.FaultLevel != constant.NotHandleFault &&
			faultRank.FaultLevel != constant.PreSeparateNPU {
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
