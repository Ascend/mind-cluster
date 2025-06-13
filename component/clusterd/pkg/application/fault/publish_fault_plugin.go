// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package fault a series of service function
package fault

import (
	"sort"
	"sync"

	"github.com/golang/protobuf/proto"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/config"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/fault"
)

func (s *FaultServer) checkFaultFromFaultCenter() {
	for {
		select {
		case <-s.serviceCtx.Done():
			return
		case allJobFaultInfo, ok := <-s.faultCh:
			if !ok {
				hwlog.RunLog.Info("faultCh has been closed.")
				return
			}
			hwlog.RunLog.Debugf("global fault info: %v", allJobFaultInfo)
			s.checkPublishFault(allJobFaultInfo)
		}
	}
}

func (s *FaultServer) checkPublishFault(allJobFaultInfo map[string]constant.JobFaultInfo) {
	jobFaultMap := make(map[string][]constant.FaultDevice, len(allJobFaultInfo))
	for jobId, jobFaultInfo := range allJobFaultInfo {
		jobInfo, ok := job.GetJobCache(jobId)
		if !ok {
			hwlog.RunLog.Errorf("jobId=%s jobInfo is not found", jobId)
			continue
		}
		filterList := filterFault(jobFaultInfo.FaultDevice)
		if len(filterList) == 0 {
			continue
		}
		jobFaultMap[jobId] = append(jobFaultMap[jobId], filterList...)

		if len(jobInfo.MultiInstanceJobId) > 0 {
			jobFaultMap[jobInfo.MultiInstanceJobId] = append(jobFaultMap[jobInfo.MultiInstanceJobId], filterList...)
		}
	}
	s.dealWithFaultInfoForMergedJob(jobFaultMap)
}

func (s *FaultServer) dealWithFaultInfoForMergedJob(jobFaultDeviceMap map[string][]constant.FaultDevice) {
	jobPublisherMap := make(map[string]*config.ConfigPublisher[*fault.FaultMsgSignal], s.serveJobNum())
	s.lock.Lock()
	for targetJobId, faultPublisher := range s.faultPublisher {
		if faultPublisher == nil || !faultPublisher.IsSubscribed() {
			hwlog.RunLog.Debugf("jobId=%s not registered or subscribe fault service", targetJobId)
			continue
		}
		jobPublisherMap[targetJobId] = faultPublisher
	}
	s.lock.Unlock()

	wg := &sync.WaitGroup{}
	for targetJobId, faultPublisher := range jobPublisherMap {
		faultInfo, ok := jobFaultDeviceMap[targetJobId]
		if ok {
			// fault occur and job has been subscribed,  send a fault occur msg to the client
			wg.Add(1)
			go func(targetJobId string, faultInfo []constant.FaultDevice,
				faultPublisher *config.ConfigPublisher[*fault.FaultMsgSignal]) {
				defer wg.Done()
				s.publishFaultInfoForMergedJob(targetJobId, faultInfo, faultPublisher)
			}(targetJobId, faultInfo, faultPublisher)
			continue
		}

		data := faultPublisher.GetSentData()
		if data == nil || data.SignalType == constant.SignalTypeNormal {
			continue
		}
		// fault recover and job has been subscribed,  send a fault recover msg to the client
		wg.Add(1)
		go func(jobId string, faultPublisher *config.ConfigPublisher[*fault.FaultMsgSignal]) {
			defer wg.Done()
			faultPublisher.SaveData(&fault.FaultMsgSignal{
				JobId:      jobId,
				SignalType: constant.SignalTypeNormal,
			})
		}(targetJobId, faultPublisher)
	}
	wg.Wait()
}

func (s *FaultServer) publishFaultInfoForMergedJob(targetJobId string, faultList []constant.FaultDevice,
	faultPublisher *config.ConfigPublisher[*fault.FaultMsgSignal]) {
	msg := faultDeviceToSortedFaultMsgSignal(targetJobId, faultList)
	sentData := faultPublisher.GetSentData()
	hwlog.RunLog.Debugf("jobId=%s generate fault msg=%v", targetJobId, msg)
	hwlog.RunLog.Debugf("jobId=%s sent fault msg=%v", targetJobId, sentData)
	if compareFaultMsg(msg, sentData) {
		hwlog.RunLog.Debugf("jobId=%s fault msg is equal, data=%v sentData=%v", targetJobId, msg, sentData)
		return
	}
	saved := faultPublisher.SaveData(msg)
	if !saved {
		hwlog.RunLog.Errorf("jobId=%v save fault msg failed, SignalType=%s", targetJobId, msg.SignalType)
		return
	}
	hwlog.RunLog.Infof("jobId=%v save fault msg success, SignalType=%s", targetJobId, msg.SignalType)
}

func faultDeviceToSortedFaultMsgSignal(targetJobId string, faultList []constant.FaultDevice) *fault.FaultMsgSignal {
	msg := &fault.FaultMsgSignal{}
	msg.JobId = targetJobId
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
		return &fault.FaultMsgSignal{JobId: targetJobId, SignalType: constant.SignalTypeNormal}
	}
	msg.SignalType = constant.SignalTypeNormal
	for _, nodeFaultInfo := range msg.NodeFaultInfo {
		if nodeFaultInfo.FaultLevel != constant.HealthyState {
			msg.SignalType = constant.SignalTypeFault
			break
		}
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
		NodeIP:   faultList[0].ServerId,
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
		}
	}
	info.FaultLevel = getStateByLevel(maxLevel)
	sort.Strings(info.FaultCodes)
	return info, maxLevel
}

// GetStateLevelByFaultLevel get state by level
func GetStateLevelByFaultLevel(faultLevel string) (string, int) {
	if faultLevel == "" {
		return constant.HealthyState, constant.HealthyLevel
	}
	switch faultLevel {
	case constant.NotHandleFault, constant.NotHandleFaultLevelStr:
		return constant.HealthyState, constant.HealthyLevel
	case constant.SubHealthFault, constant.PreSeparateFault, constant.PreSeparateFaultLevelStr:
		return constant.SubHealthyState, constant.SubHealthyLevel
	default:
		return constant.UnHealthyState, constant.UnHealthyLevel
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
	if this == nil {
		return other.SignalType == constant.SignalTypeNormal
	}
	if other == nil {
		return this.SignalType == constant.SignalTypeNormal
	}
	if this.SignalType == constant.SignalTypeNormal && other.SignalType == constant.SignalTypeNormal {
		return true
	}
	if this.SignalType != other.SignalType {
		return false
	}
	return proto.Equal(this, other)
}

func filterFault(faultDeviceList []constant.FaultDevice) []constant.FaultDevice {
	if len(faultDeviceList) == 0 {
		return nil
	}
	filteredList := make([]constant.FaultDevice, 0, len(faultDeviceList))
	for _, faultDevice := range faultDeviceList {
		if _, level := GetStateLevelByFaultLevel(faultDevice.FaultLevel); level == constant.HealthyLevel {
			hwlog.RunLog.Debugf("fileter fault device %v", faultDevice)
			continue
		}
		filteredList = append(filteredList, faultDevice)
	}
	return filteredList
}
