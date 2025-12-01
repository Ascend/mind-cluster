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

		hwlog.RunLog.Infof("starting ubpingmesh, cardID: %d, deviceID: %d, physicID: %s",
			chip.CardID, chip.DeviceID, physicID)
		if err := d.devManager.DcStartHccsPingMesh(chip.CardID, chip.DeviceID, 0, common.HccspingMeshOperate{
			UBPingMeshOperateList: ubOperateList,
		}); err != nil {
			hwlog.RunLog.Errorf("start ub pingmesh failed, err: %v", err)
		}
	}
}

func (d *DevManager) restartUbPingMesh(cardID, deviceID int32) {
	hwlog.RunLog.Infof("hccspingmesh task stopped, ready to restart, cardID: %d, "+"deviceID: %d", cardID, deviceID)
	physicID := findPhysicID(d.chips, cardID, deviceID)
	if physicID == "" {
		hwlog.RunLog.Warnf("cannot find physicID for cardID: %d, deviceID: %d", cardID, deviceID)
		return
	}

	pingItems, ok := d.currentPolicy.DestAddrMap[physicID]
	if !ok {
		hwlog.RunLog.Warnf("cannot find ping items for physicID: %s", physicID)
		return
	}

	grouped := groupPingItemsBySrcAddr(pingItems)
	ubOperateList := buildUbOperateList(grouped, d.currentPolicy.Config.TaskInterval)

	hwlog.RunLog.Infof("start pingmesh cardID: %d, deviceID: %d", cardID, deviceID)
	err := d.devManager.DcStartHccsPingMesh(cardID, deviceID, 0, common.HccspingMeshOperate{
		UBPingMeshOperateList: ubOperateList,
	})
	if err != nil {
		hwlog.RunLog.Errorf("restart ub pingmesh failed, cardID: %d, deviceID: %d, err: %v", cardID, deviceID, err)
		return
	}
	hwlog.RunLog.Infof("restart ub pingmesh success, cardID: %d, deviceID: %d", cardID, deviceID)
}

func findPhysicID(chips map[string]*common.ChipBaseInfo, cardID, deviceID int32) string {
	for pid, chip := range chips {
		if chip.CardID == cardID && chip.DeviceID == deviceID {
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
