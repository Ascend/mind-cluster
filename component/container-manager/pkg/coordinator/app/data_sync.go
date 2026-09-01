/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package app

import (
	"context"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
	"container-manager/pkg/coordinator/proto"
)

// dataSyncLoop pushes the local container snapshot to the leader(s)
func (c *Coordinator) dataSyncLoop(ctx context.Context) {
	checkTick := time.NewTicker(time.Duration(common.ParamOption.EventSyncInterval) * time.Second)
	defer checkTick.Stop()

	fullTick := time.NewTicker(time.Duration(common.ParamOption.ScheduledSyncInterval) * time.Second)
	defer fullTick.Stop()
	hwlog.RunLog.Infof("dataSync started: change-check=%ds full-sync=%ds",
		common.ParamOption.EventSyncInterval, common.ParamOption.ScheduledSyncInterval)

	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Infof("dataSync stopped")
			return
		case <-checkTick.C:
			if c.ops.HasDataChanged() {
				c.syncOnce("changed")
			}
		case <-fullTick.C:
			// Hourly safety net / repair sync regardless of changes.
			c.syncOnce("scheduled")
		}
	}
}

// syncOnce builds a SyncDataReq from the local container snapshot and sends it
// to all leaders.
func (c *Coordinator) syncOnce(trigger string) {
	if c.ops == nil {
		hwlog.RunLog.Debugf("container ops not injected")
		return
	}
	containers := c.ops.GetLocalContainers()
	if len(containers) == 0 {
		hwlog.RunLog.Debugf("no local containers to sync")
		return
	}
	req := &proto.SyncDataReq{
		Uuid:       utils.NewUUID(),
		NodeId:     common.ParamOption.LocalNodeID,
		Containers: containers,
		SyncTime:   time.Now().Unix(),
	}
	c.syncDataToAllLeaders(req)
	hwlog.RunLog.Infof("syncData (%s, %d containers) ok", trigger, len(containers))
}

// syncDataToAllLeaders sends a SyncDataReq to ALL leaders concurrently.
func (c *Coordinator) syncDataToAllLeaders(req *proto.SyncDataReq) {
	leaders := common.ParamOption.LeaderAddrs
	if len(leaders) == 0 {
		return
	}
	var wg sync.WaitGroup
	wg.Add(len(leaders))
	for _, addr := range leaders {
		go func(leaderAddr string) {
			defer wg.Done()
			if c.isLocalLeader(leaderAddr) {
				c.containerStore.ApplySync(req)
			} else {
				err := c.client.callLeaderForSyncData(c.ctx, leaderAddr, req)
				if err != nil {
					hwlog.RunLog.Warnf("sync data to leader %s failed: %v", leaderAddr, err)
				}
			}
		}(addr)
	}
	wg.Wait()
}
