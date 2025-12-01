/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package roceping for ping by icmp in RoCE mesh net between super pods in A5
package roceping

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

const (
	// Rule for RoCE ping rule
	Rule        types.PingMeshRule = "roce-ping"
	maxFileSize                    = 64 * 1024 * 1024
	timeDelta                      = 60 * 2
)

// GeneratorImp generator impls
type GeneratorImp struct {
	local       string
	superPodId  string
	serverIndex string
	DestAddrMap map[string][]types.PingItem
}

// NewGenerator create instance of GeneratorImp
func NewGenerator(nodeName, superPodId, serverIndex string) *GeneratorImp {
	return &GeneratorImp{
		local:       nodeName,
		superPodId:  superPodId,
		serverIndex: serverIndex,
		DestAddrMap: make(map[string][]types.PingItem),
	}
}

// Generate for generate ping policy
func (g *GeneratorImp) Generate(addrs map[string]types.SuperDeviceIDs) map[string]types.DestinationAddress {
	if g == nil {
		return nil
	}

	pingListFile, err := slownet.GetRoCEPingListFilePath(g.superPodId, g.serverIndex)
	if err != nil {
		hwlog.RunLog.Errorf("get ping list file path failed, err: %v", err)
		return nil
	}
	if err = slownet.CheckIsExistAndValid(pingListFile); err != nil {
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

// GetDestAddrMap for get dst addr map info
func (g *GeneratorImp) GetDestAddrMap() map[string][]types.PingItem {
	if g == nil {
		return nil
	}
	return g.DestAddrMap
}
