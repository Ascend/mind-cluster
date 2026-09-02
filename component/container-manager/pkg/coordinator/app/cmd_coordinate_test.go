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

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"container-manager/pkg/common"
	"container-manager/pkg/coordinator/domain"
	"container-manager/pkg/coordinator/proto"
)

func TestCmdCoordinate(t *testing.T) {
	convey.Convey("test RequestStopJobs and RequestStartJobs", t, testRequestActions)
	convey.Convey("test requestJobs", t, testRequestJobs)
	convey.Convey("test callAllLeadersForCoord", t, testCallAllLeadersForCoord)
	convey.Convey("test tryLeader", t, testTryLeader)
	convey.Convey("test broadcastToOrdinary", t, testBroadcastToOrdinary)
	convey.Convey("test isStreamsAlive", t, testIsStreamsAlive)
	convey.Convey("test broadcastCoordReq", t, testBroadcastCoordReq)
	convey.Convey("test validateJobsForCoord", t, testValidateJobsForCoord)
	convey.Convey("test executeLocal", t, testExecuteLocal)
	convey.Convey("test isLocalLeader", t, testIsLocalLeader)
}

func testRequestActions() {
	resetParamOption()
	c := newMockCoordinator()
	// both wrappers forward to requestJobs; empty job ids short-circuit to nil
	convey.So(c.RequestStopJobs(nil, nil), convey.ShouldBeNil)
	convey.So(c.RequestStartJobs(nil, nil), convey.ShouldBeNil)
}

func testRequestJobs() {
	convey.Convey("empty job ids returns nil", func() {
		resetParamOption()
		c := newMockCoordinator()
		err := c.requestJobs(nil, nil, common.ActionStop)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("disabled returns errCoordNotReady", func() {
		resetParamOption()
		c := newMockCoordinator()
		err := c.requestJobs([]string{testJobID}, nil, common.ActionStop)
		convey.So(err, convey.ShouldEqual, errCoordNotReady)
	})

	convey.Convey("enabled forwards to callAllLeadersForCoord", func() {
		resetParamOption()
		common.ParamOption.LeaderAddrs = []string{testLeaderA}
		common.ParamOption.LocalNodeID = testNodeID
		c := newMockCoordinator()
		var called bool
		var p1 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "callAllLeadersForCoord",
			func(_ *Coordinator, _ *proto.CoordinateReq) error { called = true; return nil })
		defer p1.Reset()
		err := c.requestJobs([]string{testJobID}, []string{testCtrID}, common.ActionStart)
		convey.So(err, convey.ShouldBeNil)
		convey.So(called, convey.ShouldBeTrue)
	})
}

func testCallAllLeadersForCoord() {
	convey.Convey("no leader found", func() {
		resetParamOption()
		common.ParamOption.LeaderAddrs = nil
		c := newMockCoordinator()
		err := c.callAllLeadersForCoord(&proto.CoordinateReq{})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("all leaders fail", func() {
		resetParamOption()
		common.ParamOption.LeaderAddrs = []string{testLeaderA, testLeaderB}
		c := newMockCoordinator()
		var p1 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "tryLeader",
			func(_ *Coordinator, _ string, _ *proto.CoordinateReq) error { return testErr })
		defer p1.Reset()
		err := c.callAllLeadersForCoord(&proto.CoordinateReq{})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("second leader succeeds and updates idx", func() {
		resetParamOption()
		common.ParamOption.LeaderAddrs = []string{testLeaderA, testLeaderB}
		c := newMockCoordinator()
		var p1 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "tryLeader",
			func(_ *Coordinator, addr string, _ *proto.CoordinateReq) error {
				if addr == testLeaderA {
					return testErr
				}
				return nil
			})
		defer p1.Reset()
		err := c.callAllLeadersForCoord(&proto.CoordinateReq{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(c.leaderIdx.Load(), convey.ShouldEqual, 1)
	})
}

func testTryLeader() {
	convey.Convey("local leader broadcasts", func() {
		resetParamOption()
		common.ParamOption.ListenAddr = testLeaderA
		c := newMockCoordinator()
		var called bool
		var p1 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "broadcastToOrdinary",
			func(_ *Coordinator, _ *proto.CoordinateReq) error { called = true; return nil })
		defer p1.Reset()
		err := c.tryLeader(testLeaderA, &proto.CoordinateReq{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(called, convey.ShouldBeTrue)
	})

	convey.Convey("client nil returns not ready", func() {
		resetParamOption()
		c := newMockCoordinator()
		err := c.tryLeader(testLeaderA, &proto.CoordinateReq{})
		convey.So(err, convey.ShouldEqual, errCoordNotReady)
	})

	convey.Convey("forwards to client", func() {
		resetParamOption()
		c := newMockCoordinator()
		c.client = &clientEndpoint{entries: map[string]*leaderEntry{}}
		var called bool
		var p1 = gomonkey.ApplyPrivateMethod(&clientEndpoint{}, "callLeaderForCoord",
			func(_ *clientEndpoint, _ string, _ *proto.CoordinateReq) error { called = true; return nil })
		defer p1.Reset()
		err := c.tryLeader(testLeaderA, &proto.CoordinateReq{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(called, convey.ShouldBeTrue)
	})
}

func testBroadcastToOrdinary() {
	convey.Convey("server nil", func() {
		resetParamOption()
		c := newMockCoordinator()
		err := c.broadcastToOrdinary(&proto.CoordinateReq{})
		convey.So(err, convey.ShouldEqual, errCoordNotReady)
	})

	convey.Convey("no hosting nodes", func() {
		resetParamOption()
		c := newMockCoordinator()
		c.server = newMockServerEndpoint(c)
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetNodesByJob", []string{})
		defer p1.Reset()
		err := c.broadcastToOrdinary(&proto.CoordinateReq{JobIds: []string{testJobID}})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("streams not alive", func() {
		resetParamOption()
		common.ParamOption.LocalNodeID = testNodeID
		c := newMockCoordinator()
		c.server = newMockServerEndpoint(c)
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetNodesByJob",
			[]string{testNodeID, testPeerID})
		defer p1.Reset()
		err := c.broadcastToOrdinary(&proto.CoordinateReq{JobIds: []string{testJobID}})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("validate fails", func() {
		resetParamOption()
		common.ParamOption.LocalNodeID = testNodeID
		c := newMockCoordinator()
		c.server = newMockServerEndpoint(c)
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetNodesByJob",
			[]string{testNodeID})
		defer p1.Reset()
		var p2 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "validateJobsForCoord",
			func(_ *Coordinator, _ *proto.CoordinateReq) error { return testErr })
		defer p2.Reset()
		err := c.broadcastToOrdinary(&proto.CoordinateReq{JobIds: []string{testJobID}})
		convey.So(err, convey.ShouldResemble, testErr)
	})

	convey.Convey("broadcast rejected", func() {
		resetParamOption()
		common.ParamOption.LocalNodeID = testNodeID
		c := newMockCoordinator()
		c.server = newMockServerEndpoint(c)
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetNodesByJob",
			[]string{testNodeID})
		defer p1.Reset()
		var p2 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "validateJobsForCoord",
			func(_ *Coordinator, _ *proto.CoordinateReq) error { return nil })
		defer p2.Reset()
		var p3 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "broadcastCoordReq",
			func(_ *Coordinator, _ *proto.CoordinateReq, _ []string) *proto.Response {
				return &proto.Response{Uuid: testUUID, Code: 1, Message: "boom"}
			})
		defer p3.Reset()
		err := c.broadcastToOrdinary(&proto.CoordinateReq{Uuid: testUUID, JobIds: []string{testJobID}})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("success", func() {
		resetParamOption()
		common.ParamOption.LocalNodeID = testNodeID
		c := newMockCoordinator()
		c.server = newMockServerEndpoint(c)
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetNodesByJob",
			[]string{testNodeID})
		defer p1.Reset()
		var p2 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "validateJobsForCoord",
			func(_ *Coordinator, _ *proto.CoordinateReq) error { return nil })
		defer p2.Reset()
		var p3 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "broadcastCoordReq",
			func(_ *Coordinator, _ *proto.CoordinateReq, _ []string) *proto.Response {
				return &proto.Response{Uuid: testUUID, Code: 0}
			})
		defer p3.Reset()
		err := c.broadcastToOrdinary(&proto.CoordinateReq{Uuid: testUUID, JobIds: []string{testJobID}})
		convey.So(err, convey.ShouldBeNil)
	})
}

func testIsStreamsAlive() {
	convey.Convey("local node only is alive", func() {
		resetParamOption()
		common.ParamOption.LocalNodeID = testNodeID
		c := newMockCoordinator()
		c.server = newMockServerEndpoint(c)
		convey.So(c.isStreamsAlive([]string{testNodeID}), convey.ShouldBeTrue)
	})

	convey.Convey("missing stream is not alive", func() {
		resetParamOption()
		common.ParamOption.LocalNodeID = testNodeID
		c := newMockCoordinator()
		c.server = newMockServerEndpoint(c)
		convey.So(c.isStreamsAlive([]string{testPeerID}), convey.ShouldBeFalse)
	})

	convey.Convey("all streams alive", func() {
		resetParamOption()
		common.ParamOption.LocalNodeID = testNodeID
		c := newMockCoordinator()
		c.server = newMockServerEndpoint(c)
		convey.So(c.server.addStream(testPeerID, &serverStream{ctx: context.Background()}), convey.ShouldBeNil)
		convey.So(c.isStreamsAlive([]string{testPeerID, testNodeID}), convey.ShouldBeTrue)
	})
}

func testBroadcastCoordReq() {
	convey.Convey("all succeed", func() {
		resetParamOption()
		common.ParamOption.LocalNodeID = testNodeID
		c := newMockCoordinator()
		c.server = newMockServerEndpoint(c)
		var p1 = gomonkey.ApplyPrivateMethod(&serverEndpoint{}, "sendCoordReq",
			func(_ *serverEndpoint, _ *proto.CoordinateReq, _ string) error { return nil })
		defer p1.Reset()
		var p2 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "executeLocal",
			func(_ *Coordinator, _ *proto.CoordinateReq) error { return nil })
		defer p2.Reset()
		resp := c.broadcastCoordReq(&proto.CoordinateReq{Uuid: testUUID}, []string{testNodeID, testPeerID})
		convey.So(resp.GetCode(), convey.ShouldEqual, 0)
	})

	convey.Convey("partial failure", func() {
		resetParamOption()
		common.ParamOption.LocalNodeID = testNodeID
		c := newMockCoordinator()
		c.server = newMockServerEndpoint(c)
		var p1 = gomonkey.ApplyPrivateMethod(&serverEndpoint{}, "sendCoordReq",
			func(_ *serverEndpoint, _ *proto.CoordinateReq, _ string) error { return testErr })
		defer p1.Reset()
		var p2 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "executeLocal",
			func(_ *Coordinator, _ *proto.CoordinateReq) error { return nil })
		defer p2.Reset()
		resp := c.broadcastCoordReq(&proto.CoordinateReq{Uuid: testUUID}, []string{testNodeID, testPeerID})
		convey.So(resp.GetCode(), convey.ShouldEqual, 1)
	})
}

func testValidateJobsForCoord() {
	convey.Convey("no container info", func() {
		c := newMockCoordinator()
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetContainersByJob",
			[]*proto.ContainerInfo{})
		defer p1.Reset()
		err := c.validateJobsForCoord(&proto.CoordinateReq{JobIds: []string{testJobID}})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("recovery disabled", func() {
		c := newMockCoordinator()
		info := mockContainerInfo(testJobID, testCtrID, common.StatusPaused)
		info.EnableRecover = false
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetContainersByJob",
			[]*proto.ContainerInfo{info})
		defer p1.Reset()
		err := c.validateJobsForCoord(&proto.CoordinateReq{JobIds: []string{testJobID}})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("start but paused by other node", func() {
		c := newMockCoordinator()
		info := mockContainerInfo(testJobID, testCtrID, common.StatusPaused)
		info.PausedByPeer = testPeerID
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetContainersByJob",
			[]*proto.ContainerInfo{info})
		defer p1.Reset()
		err := c.validateJobsForCoord(&proto.CoordinateReq{
			JobIds: []string{testJobID}, NodeId: testNodeID, Action: common.ActionStart})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("start but not paused", func() {
		c := newMockCoordinator()
		info := mockContainerInfo(testJobID, testCtrID, "running")
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetContainersByJob",
			[]*proto.ContainerInfo{info})
		defer p1.Reset()
		err := c.validateJobsForCoord(&proto.CoordinateReq{
			JobIds: []string{testJobID}, NodeId: testNodeID, Action: common.ActionStart})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("start allowed when paused by self", func() {
		c := newMockCoordinator()
		info := mockContainerInfo(testJobID, testCtrID, common.StatusPaused)
		info.PausedByPeer = testNodeID
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetContainersByJob",
			[]*proto.ContainerInfo{info})
		defer p1.Reset()
		err := c.validateJobsForCoord(&proto.CoordinateReq{
			JobIds: []string{testJobID}, NodeId: testNodeID, Action: common.ActionStart})
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("replica mismatch", func() {
		c := newMockCoordinator()
		info := mockContainerInfo(testJobID, testCtrID, common.StatusPaused)
		info.JobReplica = 2
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetContainersByJob",
			[]*proto.ContainerInfo{info})
		defer p1.Reset()
		err := c.validateJobsForCoord(&proto.CoordinateReq{
			JobIds: []string{testJobID}, Action: common.ActionStop})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("replica not configured", func() {
		c := newMockCoordinator()
		info := mockContainerInfo(testJobID, testCtrID, common.StatusPaused)
		info.JobReplica = 0
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetContainersByJob",
			[]*proto.ContainerInfo{info})
		defer p1.Reset()
		err := c.validateJobsForCoord(&proto.CoordinateReq{
			JobIds: []string{testJobID}, Action: common.ActionStop})
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("stop success with replica match", func() {
		c := newMockCoordinator()
		info := mockContainerInfo(testJobID, testCtrID, common.StatusPaused)
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetContainersByJob",
			[]*proto.ContainerInfo{info})
		defer p1.Reset()
		err := c.validateJobsForCoord(&proto.CoordinateReq{
			JobIds: []string{testJobID}, Action: common.ActionStop})
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("start success with replica match", func() {
		c := newMockCoordinator()
		info := mockContainerInfo(testJobID, testCtrID, common.StatusPaused)
		info.PausedByPeer = testNodeID
		var p1 = gomonkey.ApplyMethodReturn(&domain.ContainerStore{}, "GetContainersByJob",
			[]*proto.ContainerInfo{info})
		defer p1.Reset()
		err := c.validateJobsForCoord(&proto.CoordinateReq{
			JobIds: []string{testJobID}, NodeId: testNodeID, Action: common.ActionStart})
		convey.So(err, convey.ShouldBeNil)
	})
}

func testExecuteLocal() {
	convey.Convey("ops nil", func() {
		c := NewCoordinator(context.Background(), nil)
		err := c.executeLocal(&proto.CoordinateReq{Action: common.ActionStop})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("stop action", func() {
		resetParamOption()
		ops := &mockOps{pauseErr: testErr}
		c := NewCoordinator(context.Background(), ops)
		err := c.executeLocal(&proto.CoordinateReq{Action: common.ActionStop})
		convey.So(err, convey.ShouldResemble, testErr)
		ops.pauseErr = nil
		err = c.executeLocal(&proto.CoordinateReq{Action: common.ActionStop})
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("start action", func() {
		resetParamOption()
		ops := &mockOps{resumeErr: testErr}
		c := NewCoordinator(context.Background(), ops)
		err := c.executeLocal(&proto.CoordinateReq{Action: common.ActionStart})
		convey.So(err, convey.ShouldResemble, testErr)
		ops.resumeErr = nil
		err = c.executeLocal(&proto.CoordinateReq{Action: common.ActionStart})
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("unknown action", func() {
		resetParamOption()
		c := newMockCoordinator()
		err := c.executeLocal(&proto.CoordinateReq{Action: "bogus"})
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testIsLocalLeader() {
	resetParamOption()
	common.ParamOption.ListenAddr = testLeaderA
	c := newMockCoordinator()
	convey.So(c.isLocalLeader(testLeaderA), convey.ShouldBeTrue)
	convey.So(c.isLocalLeader(testLeaderB), convey.ShouldBeFalse)
}
