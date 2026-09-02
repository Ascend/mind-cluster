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

// Package domain test for container store
package domain

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"container-manager/pkg/common"
	"container-manager/pkg/coordinator/proto"
)

const (
	jobId1 = "job1"
	jobId2 = "job2"

	nodeId1       = "node-1"
	nodeId2       = "node-2"
	expiredNodeId = "expired-node"
	aliveNodeId   = "alive-node"
	missingNodeId = "missing-node"

	ctrId1 = "ctr1"
	ctrId2 = "ctr2"
	ctrId3 = "ctr3"
	ctrId5 = "ctr5"
)

func TestContainerStore(t *testing.T) {
	convey.Convey("test NewContainerStore", t, testNewContainerStore)
	convey.Convey("test groupByJob", t, testGroupByJob)
	convey.Convey("test ApplySync and GetNodesByJob", t, testApplySyncAndGetNodesByJob)
	convey.Convey("test GetNodesByJob with expired nodes", t, testGetNodesByJobExpired)
	convey.Convey("test GetContainersByJob", t, testGetContainersByJob)
	convey.Convey("test cloneContainerInfo", t, testCloneContainerInfo)
	convey.Convey("test cleanContainerByNodes", t, testCleanContainerByNodes)
	convey.Convey("test nodeSnapshot expired", t, testNodeSnapshotExpired)
}

func mockContainerInfo(jobID, ctrID, status string) *proto.ContainerInfo {
	return &proto.ContainerInfo{
		JobId:         jobID,
		ContainerId:   ctrID,
		Status:        status,
		JobReplica:    1,
		EnableRecover: true,
		PhyIds:        []int32{0, 1},
	}
}

func testNewContainerStore() {
	store := NewContainerStore()
	convey.So(store, convey.ShouldNotBeNil)
	convey.So(len(store.nodes), convey.ShouldEqual, 0)
}

func testGroupByJob() {
	containers := []*proto.ContainerInfo{
		mockContainerInfo(jobId1, ctrId1, common.StatusPaused),
		mockContainerInfo(jobId1, ctrId2, common.StatusRunning),
		mockContainerInfo(jobId2, ctrId3, common.StatusRunning),
	}
	byJob := groupByJob(containers)
	convey.So(len(byJob), convey.ShouldEqual, 2)
	convey.So(len(byJob[jobId1]), convey.ShouldEqual, 2)
	convey.So(len(byJob[jobId2]), convey.ShouldEqual, 1)
}

func testApplySyncAndGetNodesByJob() {
	store := NewContainerStore()
	req := &proto.SyncDataReq{
		Uuid:       "uuid-1",
		NodeId:     nodeId1,
		Containers: []*proto.ContainerInfo{mockContainerInfo(jobId1, ctrId1, common.StatusPaused)},
	}
	store.ApplySync(req)

	// empty job list returns nil
	convey.So(store.GetNodesByJob(nil), convey.ShouldBeNil)
	convey.So(store.GetNodesByJob([]string{}), convey.ShouldBeNil)

	// matching job
	nodes := store.GetNodesByJob([]string{jobId1})
	convey.So(nodes, convey.ShouldResemble, []string{nodeId1})

	// non-matching job
	nodes = store.GetNodesByJob([]string{"job-unknown"})
	convey.So(nodes, convey.ShouldBeEmpty)

	// snapshot overwrite
	req2 := &proto.SyncDataReq{NodeId: nodeId1, Containers: []*proto.ContainerInfo{mockContainerInfo(jobId2, ctrId5, common.StatusRunning)}}
	store.ApplySync(req2)
	nodes = store.GetNodesByJob([]string{jobId1})
	convey.So(nodes, convey.ShouldBeEmpty)
	nodes = store.GetNodesByJob([]string{jobId2})
	convey.So(nodes, convey.ShouldResemble, []string{nodeId1})
}

func testGetNodesByJobExpired() {
	store := NewContainerStore()
	now := time.Now()
	// give expired-node an old syncTime so its snapshot is stale
	var p1 = gomonkey.ApplyFuncReturn(time.Now, now.Add(-time.Hour*24))
	store.ApplySync(&proto.SyncDataReq{
		NodeId:     expiredNodeId,
		Containers: []*proto.ContainerInfo{mockContainerInfo(jobId1, ctrId1, common.StatusPaused)},
	})
	p1.Reset()

	// alive-node gets the current syncTime
	store.ApplySync(&proto.SyncDataReq{
		NodeId:     aliveNodeId,
		Containers: []*proto.ContainerInfo{mockContainerInfo(jobId1, ctrId2, common.StatusPaused)},
	})

	nodes := store.GetNodesByJob([]string{jobId1})
	convey.So(nodes, convey.ShouldResemble, []string{aliveNodeId})
	// expired node should be cleaned up
	_, exist := store.nodes[expiredNodeId]
	convey.So(exist, convey.ShouldBeFalse)
}

func testGetContainersByJob() {
	store := NewContainerStore()
	store.ApplySync(&proto.SyncDataReq{
		NodeId:     nodeId1,
		Containers: []*proto.ContainerInfo{mockContainerInfo(jobId1, ctrId1, common.StatusPaused)},
	})
	store.ApplySync(&proto.SyncDataReq{
		NodeId:     nodeId2,
		Containers: []*proto.ContainerInfo{mockContainerInfo(jobId1, ctrId2, common.StatusPaused)},
	})

	convey.So(store.GetContainersByJob(nil), convey.ShouldBeNil)

	infos := store.GetContainersByJob([]string{jobId1})
	convey.So(len(infos), convey.ShouldEqual, 2)

	// deep copy: modifying returned info must not affect the store
	for _, c := range infos {
		c.ContainerId = "mutated"
		c.PhyIds[0] = 99
	}
	infos2 := store.GetContainersByJob([]string{jobId1})
	for _, c := range infos2 {
		convey.So(c.ContainerId, convey.ShouldNotEqual, "mutated")
	}
}

func testCloneContainerInfo() {
	convey.So(cloneContainerInfo(nil), convey.ShouldBeNil)
	c := mockContainerInfo(jobId1, ctrId1, common.StatusPaused)
	clone := cloneContainerInfo(c)
	convey.So(clone.ContainerId, convey.ShouldEqual, c.ContainerId)
	convey.So(len(clone.PhyIds), convey.ShouldEqual, len(c.PhyIds))
	// mutating clone phyIds must not affect the source
	clone.PhyIds[0] = 100
	convey.So(c.PhyIds[0], convey.ShouldEqual, 0)
}

func testCleanContainerByNodes() {
	store := NewContainerStore()
	store.ApplySync(&proto.SyncDataReq{NodeId: nodeId1, Containers: []*proto.ContainerInfo{mockContainerInfo(jobId1, ctrId1, common.StatusPaused)}})
	store.ApplySync(&proto.SyncDataReq{NodeId: nodeId2, Containers: []*proto.ContainerInfo{mockContainerInfo(jobId1, ctrId2, common.StatusPaused)}})

	// empty input is a no-op
	store.cleanContainerByNodes(nil)
	convey.So(len(store.nodes), convey.ShouldEqual, 2)

	store.cleanContainerByNodes([]string{nodeId1, missingNodeId})
	_, exist1 := store.nodes[nodeId1]
	_, exist2 := store.nodes[nodeId2]
	convey.So(exist1, convey.ShouldBeFalse)
	convey.So(exist2, convey.ShouldBeTrue)
}

func testNodeSnapshotExpired() {
	ns := &nodeSnapshot{syncTime: time.Now().Unix()}
	convey.So(ns.expired(), convey.ShouldBeFalse)

	ns.syncTime = time.Now().Unix() - int64(common.ParamOption.ScheduledSyncInterval*expireFactor) - 1
	convey.So(ns.expired(), convey.ShouldBeTrue)
}
