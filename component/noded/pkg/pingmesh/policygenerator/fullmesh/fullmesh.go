// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package fullmesh is one of policy generator for pingmesh
package fullmesh

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"ascend-common/api/slownet"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager/common"
	"nodeD/pkg/pingmesh/types"
)

// GeneratorImp generator impls
type GeneratorImp struct {
	local       string
	superPodId  string
	serverIndex string
	DestAddrMap map[string][]types.PingItem
}

// Rule is the rule for generating pingmesh destination addresses
const (
	Rule        types.PingMeshRule = "full-mesh"
	maxFileSize                    = 64 * 1024 * 1024
)

// New create a new generator
func New(node string, superPodId string, serverIndex string) *GeneratorImp {
	return &GeneratorImp{
		local:       node,
		superPodId:  superPodId,
		serverIndex: serverIndex,
		DestAddrMap: make(map[string][]types.PingItem),
	}
}

// Generate generate pingmesh dest addresses
func (g *GeneratorImp) Generate(addrs map[string]types.SuperDeviceIDs) map[string]types.DestinationAddress {
	if g == nil {
		return nil
	}
	_, ok := addrs[g.local]
	if !ok {
		hwlog.RunLog.Errorf("local node %s not found in addrs", g.local)
		return nil
	}
	pingListFile, err := slownet.GetPingListFilePath(g.superPodId, g.serverIndex)
	if err != nil {
		hwlog.RunLog.Errorf("get ping list file path failed, err: %v", err)
		return nil
	}
	data, err := utils.ReadLimitBytes(pingListFile, maxFileSize)
	if err != nil {
		hwlog.RunLog.Errorf("read data from ping list file %s failed, err: %v", pingListFile, err)
		return nil
	}
	pingListInfos := make([]types.PingListInfo, 0)
	if err = json.Unmarshal(data, &pingListInfos); err != nil {
		hwlog.RunLog.Errorf("unmarshal data from ping list file %s failed, err: %v", pingListFile, err)
		return nil
	}
	if len(pingListInfos) != 1 {
		hwlog.RunLog.Error("ping list info length is not equal to 1")
		return nil
	}
	destAddresses := make(map[string]types.DestinationAddress)
	destAddrMap := make(map[string][]types.PingItem)
	for _, pingItem := range pingListInfos[0].PingList {
		item := pingItem
		srcCardPhyId := strconv.Itoa(item.SrcCardPhyId)
		destAddrMap[srcCardPhyId] = append(destAddrMap[srcCardPhyId], item)
	}
	for srcCardPhyId, destAddrs := range destAddrMap {
		pingDestAddr := make([]string, 0, len(destAddrs))
		for _, item := range destAddrs {
			pingDestAddr = append(pingDestAddr, item.DstAddr)
		}
		sort.Strings(pingDestAddr)
		hwlog.RunLog.Debugf("internal dest for card(%s) is %v", srcCardPhyId, pingDestAddr)
		if destAddresses[srcCardPhyId] == nil {
			destAddresses[srcCardPhyId] = make(types.DestinationAddress)
		}
		destAddresses[srcCardPhyId][common.InternalPingMeshTaskID] = strings.Join(pingDestAddr, ",")
	}

	g.DestAddrMap = destAddrMap
	return destAddresses
}

// GetDestAddrMap get destAddrMap info
func (g *GeneratorImp) GetDestAddrMap() map[string][]types.PingItem {
	if g == nil {
		return nil
	}
	return g.DestAddrMap
}
