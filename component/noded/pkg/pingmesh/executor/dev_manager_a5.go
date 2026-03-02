/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package executor for execute ub pingmesh
package executor

import (
	"encoding/hex"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"nodeD/pkg/pingmesh/types"
)

func (d *DevManager) startUbPingMesh() {
	for physicID, pingItems := range d.currentPolicy.DestAddrMap {
		chip, ok := d.chips[physicID]
		if !ok {
			continue
		}
		grouped := groupPingItemsBySrcAddr(pingItems)
		ubOperateList := buildUbOperateList(grouped, d.currentPolicy.Config.TaskInterval)

		hwlog.RunLog.Infof("starting ubpingmesh, logicID: %d, physicID: %s", chip.LogicID, physicID)
		if err := d.devManager.StartHccsPingMesh(chip.LogicID, 0, common.HccspingMeshOperate{
			UBPingMeshOperateList: ubOperateList,
		}); err != nil {
			hwlog.RunLog.Errorf("start ub pingmesh failed, err: %v", err)
		}
	}
}

func (d *DevManager) restartUbPingMesh(logicID int32) {
	hwlog.RunLog.Infof("hccspingmesh task stopped, ready to restart, logicID: %d", logicID)
	physicID := findPhysicID(d.chips, logicID)
	if physicID == "" {
		hwlog.RunLog.Warnf("cannot find physicID for logicID: %d", logicID)
		return
	}

	pingItems, ok := d.currentPolicy.DestAddrMap[physicID]
	if !ok {
		hwlog.RunLog.Warnf("cannot find ping items for physicID: %s", physicID)
		return
	}

	grouped := groupPingItemsBySrcAddr(pingItems)
	ubOperateList := buildUbOperateList(grouped, d.currentPolicy.Config.TaskInterval)

	hwlog.RunLog.Infof("start pingmesh logicID: %d", logicID)
	err := d.devManager.StartHccsPingMesh(logicID, 0, common.HccspingMeshOperate{
		UBPingMeshOperateList: ubOperateList,
	})
	if err != nil {
		hwlog.RunLog.Errorf("restart ub pingmesh failed, logicID: %d, err: %v", logicID, err)
		return
	}
	hwlog.RunLog.Infof("restart ub pingmesh success, logicID: %d", logicID)
}

func findPhysicID(chips map[string]*common.ChipBaseInfo, logicID int32) string {
	for pid, chip := range chips {
		if chip.LogicID == logicID {
			return pid
		}
	}
	return ""
}

func groupPingItemsBySrcAddr(items []types.PingItem) map[string][]string {
	grouped := make(map[string][]string)
	for _, item := range items {
		grouped[item.SrcAddr] = append(grouped[item.SrcAddr], item.DstAddr)
	}
	return grouped
}

func buildUbOperateList(grouped map[string][]string, interval int) []common.UBPingMeshOperate {
	ubList := make([]common.UBPingMeshOperate, 0, len(grouped))
	for srcAddr, dstAddrs := range grouped {
		temSrcEid, err := hex.DecodeString(srcAddr)
		if err != nil {
			hwlog.RunLog.Warnf("invalid srcAddr hex string: %s, err: %v", srcAddr, err)
			continue
		}

		var srcEid common.Eid
		copy(srcEid.Raw[:], temSrcEid)

		dstEids := make([]common.Eid, len(dstAddrs))
		for i, dst := range dstAddrs {
			tempDstEid, err := hex.DecodeString(dst)
			if err != nil {
				hwlog.RunLog.Warnf("invalid dst hex string: %s, err: %v", dst, err)
				continue
			}
			copy(dstEids[i].Raw[:], tempDstEid)
		}

		ubList = append(ubList, common.UBPingMeshOperate{
			SrcEID:       srcEid,
			DstEIDList:   dstEids,
			DstNum:       len(dstEids),
			PktSize:      common.DefaultPktSize,
			PktSendNum:   common.DefaultPktSendNum,
			PktInterval:  common.DefaultPktInterval,
			Timeout:      common.DefaultTimeout,
			TaskInterval: interval,
			TaskID:       0,
		})
	}
	return ubList
}
