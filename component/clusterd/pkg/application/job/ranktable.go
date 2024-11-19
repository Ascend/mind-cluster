// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job function
package job

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"k8s.io/api/core/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

var _ RankTabler = &RankTable{}

// RankTabler interface to maintain properties
type RankTabler interface {
	// UnmarshalToRankTable Unmarshal json string to RankTable
	UnmarshalToRankTable(jsonString string) error
	// CachePodInfo cache pod info to RankTableV1
	CachePodInfo(pod *v1.Pod, instance Instance, rankIndex *int) error
	// RemovePodInfo Remove pod info from RankTable
	RemovePodInfo(namespace string, name string) error
	// SetStatus Set status of RankTableStatus
	SetStatus(status string)
	// GetStatus Get status of RankTableStatus
	GetStatus() string
	// GetPodNum get pod num
	GetPodNum() int
	// GetJobHealthy get job's resource status
	GetJobHealthy() (bool, []string)
	// SetJobNodeHealthy get job's nodes status
	SetJobNodeHealthy(nodeName string, health bool)
	// SetJobDeviceHealthy get job's nodes status
	SetJobDeviceHealthy(nodeName string, networkUnhealthy, unHealthy string)
	// GetJobDeviceNumPerNode get the num of devices per node
	GetJobDeviceNumPerNode() int
	// GetHccLJsonSlice get hccl json slice
	GetHccLJsonSlice() []string
	// GetFirstServerIp get the vcJob master addr
	GetFirstServerIp() string
	// GetServerList get servers in rank table
	GetServerList() []*ServerHccl
}

// SetStatus Set status of RankTableStatus
func (r *RankTableStatus) SetStatus(status string) {
	r.Status = status
}

// GetStatus Get status of RankTableStatus
func (r *RankTableStatus) GetStatus() string {
	return r.Status
}

// UnmarshalToRankTable Unmarshal json string to RankTable
func (r *RankTableStatus) UnmarshalToRankTable(jsonString string) error {
	// get string bytes with len
	if len(jsonString) > cmDataMaxMemory {
		return fmt.Errorf("rank table date size is out of memory")
	}
	err := json.Unmarshal([]byte(jsonString), &r)
	if err != nil {
		return fmt.Errorf("parse configmap data error: %#v", err)
	}
	if r.Status != ConfigmapCompleted && r.Status != ConfigmapInitializing {
		return fmt.Errorf("configmap status abnormal: %#v", err)
	}
	return nil
}

// CheckDeviceInfo validation of DeviceInfo
func CheckDeviceInfo(instance *Instance) bool {
	if parsedIP := net.ParseIP(instance.ServerID); parsedIP == nil {
		return false
	}
	if len(instance.Devices) == 0 || len(instance.Devices) > A800MaxChipNum {
		return false
	}
	for _, item := range instance.Devices {
		if value, err := strconv.Atoi(item.DeviceID); err != nil || value < 0 {
			return false
		}
		if parsedIP := net.ParseIP(item.DeviceIP); parsedIP == nil {
			return false
		}
	}
	return true
}

// CachePodInfo Cache pod info to RankTableV2
func (r *RankTable) CachePodInfo(pod *v1.Pod, instance Instance, rankIndex *int) error {
	var server ServerHccl
	if !CheckDeviceInfo(&instance) {
		return fmt.Errorf("deviceInfo failed the validation")
	}
	for _, server := range r.ServerList {
		if server.PodID == instance.PodName {
			return fmt.Errorf("pod %s/%s is already cached", pod.Namespace, pod.Name)
		}
	}

	// Build new server-level struct from device info
	server.ServerID = instance.ServerID
	server.PodID = instance.PodName
	server.ServerName = pod.Spec.NodeName
	rankFactor := len(instance.Devices)
	if rankFactor > A800MaxChipNum {
		return fmt.Errorf("get error device num(%d), device num is out of range", rankFactor)
	}
	for _, device := range instance.Devices {
		var serverDevice Device
		serverDevice.DeviceID = device.DeviceID
		serverDevice.DeviceIP = device.DeviceIP
		serverDevice.RankID = strconv.Itoa(*rankIndex*rankFactor + len(server.DeviceList))

		server.DeviceList = append(server.DeviceList, &serverDevice)
	}
	if len(server.DeviceList) < 1 {
		return fmt.Errorf("pod %s/%s failed to get the list of device", pod.Namespace, pod.Name)
	}

	r.ServerList = append(r.ServerList, &server)
	sort.Slice(r.ServerList, func(i, j int) bool {
		iRank, err := strconv.ParseInt(r.ServerList[i].DeviceList[0].RankID, Decimal, BitSize32)
		jRank, err2 := strconv.ParseInt(r.ServerList[j].DeviceList[0].RankID, Decimal, BitSize32)
		if err != nil || err2 != nil {
			return false
		}
		return iRank < jRank
	})
	r.ServerCount = strconv.Itoa(len(r.ServerList))
	*rankIndex++
	return nil
}

// RemovePodInfo Remove pod info from RankTableV2
func (r *RankTable) RemovePodInfo(namespace string, podID string) error {
	hasInfoToRemove := false
	serverList := r.ServerList
	for idx, server := range serverList {
		if server.PodID == podID {
			length := len(serverList)
			serverList[idx] = serverList[length-1]
			serverList = serverList[:length-1]
			hasInfoToRemove = true
			break
		}
	}

	if !hasInfoToRemove {
		return fmt.Errorf("no data of pod %s/%s can be removed", namespace, podID)
	}
	r.ServerList = serverList
	r.ServerCount = strconv.Itoa(len(r.ServerList))

	return nil
}

// GetPodNum get pod num
func (r *RankTable) GetPodNum() int {
	return len(r.ServerList)
}

// GetFirstServerIp get the vcJob master addr
func (r *RankTable) GetFirstServerIp() string {
	if r == nil || len(r.ServerList) == 0 {
		return ""
	}
	return r.ServerList[0].ServerID
}

// GetHccLJsonSlice get slice of HccL json
func (r *RankTable) GetHccLJsonSlice() []string {
	if len(r.ServerList) == 0 {
		return nil
	}
	serverNum := len(r.ServerList) * len(r.ServerList[0].DeviceList)
	if serverNum <= deviceNumThresholds {
		r.Total = 1
		str, err := json.Marshal(r)
		if err == nil {
			return []string{string(str)}
		}
		hwlog.RunLog.Errorf("Marshal hccl json error %v", err)
		return nil
	}
	if serverNum%deviceNumThresholds == 0 {
		r.Total = serverNum / deviceNumThresholds
	} else {
		r.Total = serverNum/deviceNumThresholds + 1
	}
	hcclJsons := make([]string, 0)

	tmpR := *r
	num := deviceNumThresholds / len(r.ServerList[0].DeviceList)
	for i := 0; i < len(r.ServerList); i += num {
		if i+num > len(r.ServerList) {
			tmpR.ServerList = r.ServerList[i:]
		} else {
			tmpR.ServerList = r.ServerList[i : i+num]
		}
		str, err := json.Marshal(tmpR)
		if err != nil {
			hwlog.RunLog.Errorf("Marshal hccl json part %v error, error is %v", i, err)
			continue
		}

		hcclJsons = append(hcclJsons, string(str))
	}
	return hcclJsons
}

// GetJobHealthy get job resource status
func (r *RankTable) GetJobHealthy() (bool, []string) {
	var faultRanks []string
	nodeFaults := r.getNodeUnhealthyRanks()
	if len(nodeFaults) != 0 {
		hwlog.RunLog.Debugf("node fault is %v", nodeFaults)
	}
	deviceFaults := r.getJobUnhealthyDeviceRanks()
	if len(deviceFaults) != 0 {
		hwlog.RunLog.Debugf("device fault is %v", deviceFaults)
	}
	faultRanks = append(faultRanks, nodeFaults...)
	faultRanks = append(faultRanks, deviceFaults...)
	if len(faultRanks) > 0 {
		return false, util.RemoveSliceDuplicateElement(faultRanks)
	}
	return true, nil
}

// GetJobDeviceNumPerNode get the num of devices per node in a job
func (r *RankTable) GetJobDeviceNumPerNode() int {
	if len(r.ServerList) == 0 {
		hwlog.RunLog.Error("failed to get device num per node: the length of server list is 0")
		return -1
	}
	return len(r.ServerList[0].DeviceList)
}

func (r *RankTable) getNodeUnhealthyRanks() []string {
	var faultRanks []string
	for _, ranks := range r.UnHealthyNode {
		faultRanks = append(faultRanks, ranks...)
	}
	return faultRanks
}

func (r *RankTable) getJobUnhealthyDeviceRanks() []string {
	var faultRanks []string
	for _, rankId := range r.UnHealthyDevice {
		faultRanks = append(faultRanks, rankId)
	}
	return faultRanks
}

// SetJobNodeHealthy set job's node healthy status
func (r *RankTable) SetJobNodeHealthy(nodeName string, health bool) {
	for _, server := range r.ServerList {
		if server.ServerName == nodeName {
			if health {
				delete(r.UnHealthyNode, nodeName)
			} else {
				r.appendNodeRank(nodeName, server)
			}
		}
	}
}

func (r *RankTable) appendNodeRank(nodeName string, server *ServerHccl) {
	var ranks = make([]string, 0, len(server.DeviceList))
	for _, device := range server.DeviceList {
		ranks = append(ranks, device.RankID)
	}
	r.UnHealthyNode[nodeName] = ranks
}

// SetJobDeviceHealthy set job's device healthy status
func (r *RankTable) SetJobDeviceHealthy(nodeName string, networkUnhealthyCards, unHealthyCards string) {
	for index, server := range r.ServerList {
		if server.ServerName == nodeName {
			r.setNodeDeviceHealthy(index, networkUnhealthyCards, unHealthyCards)
		}
	}
}

func (r *RankTable) setNodeDeviceHealthy(serverIndex int, networkUnhealthyCards, unHealthyCards string) {
	for _, dev := range r.ServerList[serverIndex].DeviceList {
		deviceKey := constant.AscendDevPrefix + dev.DeviceID
		if strings.Contains(networkUnhealthyCards, deviceKey) || strings.Contains(unHealthyCards, deviceKey) {
			r.UnHealthyDevice[dev.RankID] = dev.RankID
		} else {
			delete(r.UnHealthyDevice, dev.RankID)
		}
	}
}

// GetServerList get servers in rank table
func (r *RankTable) GetServerList() []*ServerHccl {
	return r.ServerList
}
