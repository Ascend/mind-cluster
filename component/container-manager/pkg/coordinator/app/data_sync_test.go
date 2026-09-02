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
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"container-manager/pkg/common"
	"container-manager/pkg/coordinator/proto"
)

func TestDataSync(t *testing.T) {
	convey.Convey("test syncOnce", t, testSyncOnce)
	convey.Convey("test syncDataToAllLeaders", t, testSyncDataToAllLeaders)
	convey.Convey("test dataSyncLoop", t, testDataSyncLoop)
}

func testSyncOnce() {
	convey.Convey("ops nil", func() {
		c := NewCoordinator(context.Background(), nil)
		c.syncOnce("changed")
	})

	convey.Convey("no local containers", func() {
		resetParamOption()
		c := newMockCoordinator()
		c.syncOnce("changed")
	})

	convey.Convey("sync ok", func() {
		resetParamOption()
		c := newMockCoordinator()
		ops := c.ops.(*mockOps)
		ops.containers = []*proto.ContainerInfo{mockContainerInfo(testJobID, testCtrID, "running")}
		var called bool
		var p1 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "syncDataToAllLeaders",
			func(_ *Coordinator, _ *proto.SyncDataReq) { called = true })
		defer p1.Reset()
		c.syncOnce("changed")
		convey.So(called, convey.ShouldBeTrue)
	})
}

func testSyncDataToAllLeaders() {
	convey.Convey("no leaders", func() {
		resetParamOption()
		c := newMockCoordinator()
		c.syncDataToAllLeaders(&proto.SyncDataReq{})
	})

	convey.Convey("sync to local leader applies to store", func() {
		resetParamOption()
		common.ParamOption.ListenAddr = testLeaderA
		common.ParamOption.LeaderAddrs = []string{testLeaderA}
		c := newMockCoordinator()
		req := &proto.SyncDataReq{
			NodeId:     testNodeID,
			Containers: []*proto.ContainerInfo{mockContainerInfo(testJobID, testCtrID, "running")},
			SyncTime:   time.Now().Unix(),
		}
		c.syncDataToAllLeaders(req)
		nodes := c.containerStore.GetNodesByJob([]string{testJobID})
		convey.So(nodes, convey.ShouldContain, testNodeID)
	})

	convey.Convey("sync to remote leader", func() {
		resetParamOption()
		common.ParamOption.LeaderAddrs = []string{testLeaderA}
		c := newMockCoordinator()
		c.client = &clientEndpoint{entries: map[string]*leaderEntry{}}
		c.syncDataToAllLeaders(&proto.SyncDataReq{NodeId: testNodeID})
	})
}

func testDataSyncLoop() {
	convey.Convey("context cancelled returns immediately", func() {
		resetParamOption()
		common.ParamOption.EventSyncInterval = 1
		common.ParamOption.ScheduledSyncInterval = 1
		c := newMockCoordinator()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		c.dataSyncLoop(ctx)
	})

	convey.Convey("ticks trigger sync", func() {
		resetParamOption()
		common.ParamOption.EventSyncInterval = 1
		common.ParamOption.ScheduledSyncInterval = 1
		c := newMockCoordinator()
		ops := c.ops.(*mockOps)
		ops.changed = true
		var calls []string
		var p1 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "syncOnce",
			func(_ *Coordinator, trigger string) { calls = append(calls, trigger) })
		defer p1.Reset()

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			defer close(done)
			c.dataSyncLoop(ctx)
		}()
		time.Sleep(1300 * time.Millisecond)
		cancel()
		<-done
		convey.So(len(calls), convey.ShouldBeGreaterThanOrEqualTo, 1)
	})
}
