// Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.

package pod

import (
	"sort"
	"strconv"

	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

const (
	level0 = 0
	level1 = 1
	level2 = 2
	level3 = 3
)

// getElement retrieves the LevelElement for a given level from a device's LevelList.
func getElement(curDevice constant.Device, level int, portAddrType string) (api.LevelElement, bool) {
	for _, lvl := range curDevice.LevelList {
		if lvl.Level != level {
			continue
		}
		if level == level0 || level == level1 {
			for _, v := range lvl.Info {
				return v, true
			}
			return api.LevelElement{}, false
		}
		elem, ok := lvl.Info[portAddrType]
		return elem, ok
	}
	return api.LevelElement{}, false
}

// shouldInclude determines whether the given level should be retained.
func shouldInclude(level int, portAddrType, customScaleOutType string) bool {
	if portAddrType == "" && customScaleOutType != "" {
		portAddrType = customScaleOutType
	}
	switch level {
	case level2:
		return portAddrType == constant.PortAddrTypeUBoE || portAddrType == constant.PortAddrTypeUBG
	case level3:
		return portAddrType == constant.PortAddrTypeRoCE
	default:
		return true
	}
}

// recordDevicePortAddrTypes collects all port addr types present in a device's level list.
func recordDevicePortAddrTypes(portAddrTypes map[string]struct{}, dev constant.Device) {
	for _, rankLevel := range dev.LevelList {
		for netType := range rankLevel.Info {
			if _, exist := portAddrTypes[netType]; !exist {
				hwlog.RunLog.Infof("the rank table file contains device port addr type: %v", netType)
				portAddrTypes[netType] = struct{}{}
			}
		}
	}
}

// genRankList initializes and populates rank with level0/level1 info.
func genRankList(rank *constant.Rank, device constant.Device) error {
	rank.Device = device
	localID, err := strconv.Atoi(device.DeviceID)
	if err != nil {
		hwlog.RunLog.Errorf("parse device id(%s) failed: %v", device.DeviceID, err)
		return err
	}
	rank.LocalID = localID
	rank.DeviceID = localID
	for level := level0; level <= level1; level++ {
		elem, ok := getElement(device, level, "")
		if !ok {
			hwlog.RunLog.Warnf("device %s level=%d has no valid element, skip append", device.DeviceID, level)
			continue
		}
		rank.LevelList = append(rank.LevelList, elem)
	}
	return nil
}

// ConstructRankListV2 fills rankTable.RankList/RankCount/Version for A5 jobs, in place.
// customScaleOutType is already ToUpper+TrimSpace'd (may be empty).
func ConstructRankListV2(rankTable *constant.RankTable, podsInJob map[string]v1.Pod,
	replicas int, customScaleOutType string) {
	allRanks, portAddrTypes := buildRanksV2(podsInJob, replicas)
	if len(allRanks) == 0 {
		return
	}

	supplementRanksV2(allRanks, portAddrTypes, customScaleOutType)

	sort.Slice(allRanks, func(i, j int) bool { return allRanks[i].RankID < allRanks[j].RankID })

	rankTable.RankList = make([]constant.Rank, 0, len(allRanks))
	for _, r := range allRanks {
		rankTable.RankList = append(rankTable.RankList, *r)
	}
	rankTable.RankCount = len(allRanks)
	rankTable.Version = "2.0"
}

// buildRanksV2 builds the rank list (level0/level1) per pod and collects all port addr types.
func buildRanksV2(podsInJob map[string]v1.Pod, replicas int) ([]*constant.Rank, map[string]struct{}) {
	allRanks := make([]*constant.Rank, 0)
	portAddrTypes := make(map[string]struct{})

	for _, pod := range podsInJob {
		podRank := getPodRank(pod)
		if podRank == -1 || podRank >= replicas {
			hwlog.RunLog.Warnf("illegal job information, replicas is %d, but podRank is %d", replicas, podRank)
			continue
		}
		podDev, _ := getPodDevice(pod)
		if len(podDev.Devices) == 0 {
			continue
		}
		rankFactor := len(podDev.Devices)
		for _, dev := range podDev.Devices {
			recordDevicePortAddrTypes(portAddrTypes, dev)
		}
		for index, dev := range podDev.Devices {
			var rank constant.Rank
			if err := genRankList(&rank, dev); err != nil {
				hwlog.RunLog.Errorf("gen rank list failed for device %s: %v", dev.DeviceID, err)
				continue
			}
			rank.RankID = podRank*rankFactor + index
			allRanks = append(allRanks, &rank)
		}
	}

	return allRanks, portAddrTypes
}

// supplementRanksV2 selects the network policy and appends level2/level3 info to each rank.
func supplementRanksV2(allRanks []*constant.Rank, portAddrTypes map[string]struct{}, customScaleOutType string) {
	portAddrType := ""
	if netInfo, err := constant.GetNetInfo(portAddrTypes, customScaleOutType); err != nil {
		hwlog.RunLog.Warnf("GetNetInfo failed: %v", err)
	} else {
		portAddrType = netInfo.PortAddrType
	}

	for _, rank := range allRanks {
		for _, level := range []int{level2, level3} {
			if !shouldInclude(level, portAddrType, customScaleOutType) {
				continue
			}
			elem, ok := getElement(rank.Device, level, portAddrType)
			if !ok {
				elem = api.LevelElement{
					NetLayer:      level,
					NetInstanceID: api.DefaultClusterName,
					NetType:       api.NetTypeCLOS,
					NetAttr:       api.NetAttrEmpty,
					RankAddrList:  []api.RankAddrItem{},
				}
			}
			rank.LevelList = append(rank.LevelList, elem)
		}
	}
}
