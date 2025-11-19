// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package fault a series of service function
package fault

import (
	"sort"
	"sync"

	"github.com/golang/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/uuid"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/config"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/l2fault"
	"clusterd/pkg/interface/grpc/fault"
)

func (s *FaultServer) checkFaultFromFaultCenter() {
	for {
		select {
		case <-s.serviceCtx.Done():
			return
		case allJobFaultInfo, ok := <-s.faultCh:
			if !ok {
				hwlog.RunLog.Info("faultCh has been closed")
				return
			}
			hwlog.RunLog.Debugf("global fault info: %v", allJobFaultInfo)
			s.checkPublishFault(allJobFaultInfo)
		}
	}
}

func (s *FaultServer) checkPublishFault(allJobFaultInfo map[string]constant.JobFaultInfo) {
	jobFaultMap := make(map[string][]constant.FaultDevice, len(allJobFaultInfo))
	deletedJobFaultDeviceMap := l2fault.L2FaultCache.GetDeletedJobFaultDeviceMap()
	for jobId, jobFaultInfo := range allJobFaultInfo {
		jobInfo, ok := job.GetJobCache(jobId)
		if !ok {
			hwlog.RunLog.Errorf("jobId=%s jobInfo is not found", jobId)
			continue
		}
		currentFaults := jobFaultInfo.FaultDevice
		deletedFaults, hasDeleted := deletedJobFaultDeviceMap[jobId]
		if len(currentFaults) == 0 && (!hasDeleted || len(deletedFaults) == 0) {
			continue
		}
		jobFaultMap[jobId] = append(jobFaultMap[jobId], currentFaults...)
		if hasDeleted && len(deletedFaults) > 0 {
			jobFaultMap[jobId] = append(jobFaultMap[jobId], deletedFaults...)
			hwlog.RunLog.Debugf("backfilled deleted L2 faults to job %s: %v", jobId, deletedFaults)
		}
		if multiJobId := jobInfo.MultiInstanceJobId; len(multiJobId) > 0 {
			jobFaultMap[multiJobId] = append(jobFaultMap[multiJobId], currentFaults...)
			if hasDeleted && len(deletedFaults) > 0 {
				jobFaultMap[multiJobId] = append(jobFaultMap[multiJobId], deletedFaults...)
				hwlog.RunLog.Debugf("backfilled deleted L2 faults to multi-instance job %s: %v",
					multiJobId, deletedFaults)
			}
		}
	}
	s.dealWithFaultInfoForMergedJob(jobFaultMap)
	s.dealWithFaultInfoForClusterJob(jobFaultMap)
}

func (s *FaultServer) dealWithFaultInfoForClusterJob(jobFaultDeviceMap map[string][]constant.FaultDevice) {
	publisherList := s.getPublisherListByJobId(constant.DefaultJobId)
	for _, publisher := range publisherList {
		s.dealWithFaultInfoForEachClusterJob(jobFaultDeviceMap, publisher)
	}
}

func (s *FaultServer) dealWithFaultInfoForEachClusterJob(jobFaultDeviceMap map[string][]constant.FaultDevice,
	clusterPublisher *config.ConfigPublisher[*fault.FaultMsgSignal]) {
	if clusterPublisher == nil || !clusterPublisher.IsSubscribed() {
		return
	}
	// job fault occur, send a fault occur msg to the client
	for jobId, faultInfo := range jobFaultDeviceMap {
		if _, ok := job.GetJobCache(jobId); !ok {
			continue
		}
		s.publishFaultInfoForMergedJob(constant.DefaultJobId, jobId, faultInfo, clusterPublisher)
	}
	deletedJob := make([]string, 0)
	// job fault recover, send a fault recover msg to the client
	for _, jobId := range clusterPublisher.GetAllSentJobIdList() {
		// filter job that already send fault msg to the client
		if _, ok := jobFaultDeviceMap[jobId]; ok {
			continue
		}
		// filter job that already be deleted
		if _, ok := job.GetJobCache(jobId); !ok {
			deletedJob = append(deletedJob, jobId)
			continue
		}
		s.publishFaultInfoForMergedJob(constant.DefaultJobId, jobId, nil, clusterPublisher)
	}
	if len(deletedJob) > 0 {
		clusterPublisher.ClearDeletedJobIdList(deletedJob)
		hwlog.RunLog.Infof("cluster fault publisher delete job key: %v", deletedJob)
	}
}

func (s *FaultServer) dealWithFaultInfoForMergedJob(jobFaultDeviceMap map[string][]constant.FaultDevice) {
	jobPublisherList := s.getAllPublisherList()
	wg := &sync.WaitGroup{}
	for _, faultPublisher := range jobPublisherList {
		targetJobId := faultPublisher.GetJobId()
		if faultPublisher == nil || !faultPublisher.IsSubscribed() {
			hwlog.RunLog.Debugf("jobId=%s not registered or subscribe fault service", targetJobId)
			continue
		}
		faultInfo, ok := jobFaultDeviceMap[targetJobId]
		if ok {
			// fault occur and job has been subscribed,  send a fault occur msg to the client
			wg.Add(1)
			go func(jobId string, faultInfo []constant.FaultDevice,
				faultPublisher *config.ConfigPublisher[*fault.FaultMsgSignal]) {
				defer wg.Done()
				s.publishFaultInfoForMergedJob(jobId, jobId, faultInfo, faultPublisher)
			}(targetJobId, faultInfo, faultPublisher)
			continue
		}

		data := faultPublisher.GetSentData(targetJobId)
		if data == nil || data.SignalType == constant.SignalTypeNormal {
			continue
		}
		// fault recover and job has been subscribed,  send a fault recover msg to the client
		wg.Add(1)
		go func(jobId string, faultPublisher *config.ConfigPublisher[*fault.FaultMsgSignal]) {
			defer wg.Done()
			s.publishFaultInfoForMergedJob(jobId, jobId, nil, faultPublisher)
		}(targetJobId, faultPublisher)
	}
	wg.Wait()
}

func (s *FaultServer) publishFaultInfoForMergedJob(pubJobId, faultJobId string, faultList []constant.FaultDevice,
	faultPublisher *config.ConfigPublisher[*fault.FaultMsgSignal]) {
	msg := faultDeviceToSortedFaultMsgSignal(faultJobId, faultList)
	sentData := faultPublisher.GetSentData(faultJobId)
	hwlog.RunLog.Debugf("jobId=%s generate fault msg=%v", pubJobId, msg)
	hwlog.RunLog.Debugf("jobId=%s sent fault msg=%v", pubJobId, sentData)
	if compareFaultMsg(msg, sentData) {
		hwlog.RunLog.Debugf("jobId=%s fault msg is equal, data=%v sentData=%v", pubJobId, msg, sentData)
		return
	}
	saved := faultPublisher.SaveData(faultJobId, msg)
	if !saved {
		hwlog.RunLog.Errorf("jobId=%v save fault msg failed, SignalType=%s", pubJobId, msg.SignalType)
		return
	}
	hwlog.RunLog.Infof("jobId=%v save fault msg success, SignalType=%s", pubJobId, msg.SignalType)
}

func faultDeviceToSortedFaultMsgSignal(targetJobId string, faultList []constant.FaultDevice) *fault.FaultMsgSignal {
	msg := &fault.FaultMsgSignal{Uuid: string(uuid.NewUUID())}
	msg.JobId = targetJobId
	msg.SignalType = constant.SignalTypeFault
	if len(faultList) == 0 {
		msg.SignalType = constant.SignalTypeNormal
		return msg
	}
	nodeFaultMap := make(map[string][]constant.FaultDevice, len(faultList))
	for _, faultInfo := range faultList {
		nodeFaultMap[faultInfo.ServerId] = append(nodeFaultMap[faultInfo.ServerId], faultInfo)
	}
	for _, nodeFaultList := range nodeFaultMap {
		nodeInfo := getNodeFaultInfo(nodeFaultList)
		if nodeInfo != nil {
			msg.NodeFaultInfo = append(msg.NodeFaultInfo, nodeInfo)
		}
	}
	if len(msg.NodeFaultInfo) == 0 {
		msg.SignalType = constant.SignalTypeNormal
		return msg
	}
	sort.Slice(msg.NodeFaultInfo, func(i, j int) bool {
		return msg.NodeFaultInfo[i].NodeIP < msg.NodeFaultInfo[j].NodeIP
	})
	return msg
}

// getNodeFaultInfo get node all fault info
func getNodeFaultInfo(faultList []constant.FaultDevice) *fault.NodeFaultInfo {
	if len(faultList) == 0 {
		return nil
	}
	deviceFaultMap := make(map[string][]constant.FaultDevice, len(faultList))
	for _, faultInfo := range faultList {
		key := faultInfo.DeviceId + constant.Comma + faultInfo.DeviceType
		deviceFaultMap[key] = append(deviceFaultMap[key], faultInfo)
	}
	info := &fault.NodeFaultInfo{
		NodeName: faultList[0].ServerName,
		NodeIP:   faultList[0].HostIp,
		NodeSN:   faultList[0].ServerSN,
	}
	maxLevel := constant.HealthyLevel
	for _, deviceFaultList := range deviceFaultMap {
		deviceInfo, level := getFaultDeviceInfo(deviceFaultList)
		if deviceInfo != nil {
			info.FaultDevice = append(info.FaultDevice, deviceInfo)
		}
		if level > maxLevel {
			maxLevel = level
		}
	}
	info.FaultLevel = getStateByLevel(maxLevel)
	sort.Slice(info.FaultDevice, func(i, j int) bool {
		if info.FaultDevice[i].DeviceId == info.FaultDevice[j].DeviceId {
			return info.FaultDevice[i].DeviceType < info.FaultDevice[j].DeviceType
		}
		return info.FaultDevice[i].DeviceId < info.FaultDevice[j].DeviceId
	})
	return info
}

// getFaultDeviceInfo get device all fault info
func getFaultDeviceInfo(faultList []constant.FaultDevice) (*fault.DeviceFaultInfo, int) {
	if len(faultList) == 0 {
		return nil, constant.HealthyLevel
	}
	info := &fault.DeviceFaultInfo{
		DeviceId:   faultList[0].DeviceId,
		DeviceType: faultList[0].DeviceType,
	}
	maxLevel := constant.HealthyLevel
	for _, faultInfo := range faultList {
		_, level := GetStateLevelByFaultLevel(faultInfo.FaultLevel)
		if level > maxLevel {
			maxLevel = level
		}
		if faultInfo.FaultCode != "" {
			info.FaultCodes = append(info.FaultCodes, faultInfo.FaultCode)
			info.FaultLevels = append(info.FaultLevels, faultInfo.FaultLevel)
			if faultInfo.SwitchChipId != "" && faultInfo.SwitchPortId != "" && faultInfo.SwitchFaultTime != "" {
				info.SwitchFaultInfos = append(info.SwitchFaultInfos, &fault.SwitchFaultInfo{
					FaultCode:    faultInfo.FaultCode,
					SwitchChipId: faultInfo.SwitchChipId,
					SwitchPortId: faultInfo.SwitchPortId,
					FaultTime:    faultInfo.SwitchFaultTime,
					FaultLevel:   faultInfo.FaultLevel,
				})
			}
		}
	}
	info.FaultLevel = getStateByLevel(maxLevel)
	if len(info.FaultCodes) > 1 {
		info.FaultCodes = util.RemoveDuplicates(info.FaultCodes)
		sort.Strings(info.FaultCodes)
	}
	return info, maxLevel
}

// GetStateLevelByFaultLevel get state by level
func GetStateLevelByFaultLevel(faultLevel string) (string, int) {
	switch faultLevel {
	case constant.NotHandleFault, constant.NotHandleFaultLevelStr:
		return constant.HealthyState, constant.HealthyLevel
	case constant.SubHealthFault, constant.PreSeparateFault, constant.PreSeparateFaultLevelStr,
		constant.SubHealthFaultStrategy, constant.PreSeparateNPU, constant.FreeRestartNPU:
		return constant.SubHealthyState, constant.SubHealthyLevel
	case constant.RestartRequest, constant.RestartBusiness, constant.RestartNPU, constant.SeparateNPU,
		constant.ManuallySeparateNPU, constant.SeparateFault, constant.SeparateFaultStrategy:
		return constant.UnHealthyState, constant.UnHealthyLevel
	default:
		return constant.HealthyState, constant.HealthyLevel
	}
}

func getStateByLevel(stateLeve int) string {
	switch stateLeve {
	case constant.SubHealthyLevel:
		return constant.SubHealthyState
	case constant.UnHealthyLevel:
		return constant.UnHealthyState
	default:
		return constant.HealthyState
	}
}

func compareFaultMsg(this, other *fault.FaultMsgSignal) bool {
	if this == nil && other == nil {
		return true
	}
	if this == nil || other == nil {
		return false
	}
	if this.SignalType != other.SignalType || this.JobId != other.JobId ||
		len(this.NodeFaultInfo) != len(other.NodeFaultInfo) {
		return false
	}
	for i, faultInfo := range this.NodeFaultInfo {
		if !proto.Equal(faultInfo, other.NodeFaultInfo[i]) {
			return false
		}
	}
	return true
}
