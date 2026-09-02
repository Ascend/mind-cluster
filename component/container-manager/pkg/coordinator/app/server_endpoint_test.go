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
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"container-manager/pkg/common"
	"container-manager/pkg/coordinator/proto"
)

func TestServerEndpoint(t *testing.T) {
	convey.Convey("test newServerEndpoint and startServer", t, testNewServerEndpoint)
	convey.Convey("test close", t, testServerClose)
	convey.Convey("test addStream", t, testAddStream)
	convey.Convey("test removeStream", t, testRemoveStream)
	convey.Convey("test getStream", t, testGetStream)
	convey.Convey("test registerUUID/unregisterUUID", t, testRegisterUUID)
	convey.Convey("test routeResp", t, testRouteResp)
	convey.Convey("test SyncData handler", t, testSyncDataHandler)
	convey.Convey("test Coordinate handler", t, testCoordinateHandler)
	convey.Convey("test InitBroadcastStream", t, testInitBroadcastStream)
	convey.Convey("test sendCoordReq", t, testSendCoordReq)
	convey.Convey("test waitResp", t, testWaitResp)
	convey.Convey("test validators", t, testValidators)
	convey.Convey("test limitQPS", t, testLimitQPS)
	convey.Convey("test extractNodeID", t, testExtractNodeID)
}

func testNewServerEndpoint() {
	resetParamOption()
	common.ParamOption.ListenAddr = testListen
	c := newMockCoordinator()
	s, err := newServerEndpoint(c)
	defer func() {
		if s != nil {
			s.close()
		}
	}()
	convey.So(err, convey.ShouldBeNil)
	// Use plain nil comparisons instead of ShouldNotBeNil: the latter deep-prints
	// the value, reflecting into the live grpc.Server (atomic fields / service
	// maps) that Serve is concurrently mutating, which trips the race detector.
	convey.So(s == nil, convey.ShouldBeFalse)
	if s != nil {
		convey.So(s.server == nil, convey.ShouldBeFalse)
	}
}

func testServerClose() {
	c := newMockCoordinator()
	s := newMockServerEndpoint(c)
	ss := &serverStream{ctx: context.Background(), cancel: func() {}}
	s.streams["n1"] = ss
	s.acks["k"] = make(chan *proto.Response, 1)
	s.close()
	convey.So(s.streams, convey.ShouldBeNil)
	convey.So(s.acks, convey.ShouldBeNil)
}

func testAddStream() {
	c := newMockCoordinator()
	s := newMockServerEndpoint(c)
	ss := &serverStream{ctx: context.Background()}
	err := s.addStream("n1", ss)
	convey.So(err, convey.ShouldBeNil)
	got, ok := s.streams["n1"]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(got, convey.ShouldEqual, ss)

	// reaching node limit rejects the new node
	for i := 0; i < common.MaxNodeNum-1; i++ {
		convey.So(s.addStream(fmt.Sprintf("fill-%d", i), &serverStream{ctx: context.Background()}), convey.ShouldBeNil)
	}
	err = s.addStream("overflow", &serverStream{ctx: context.Background()})
	convey.So(err, convey.ShouldNotBeNil)
}

func testRemoveStream() {
	c := newMockCoordinator()
	s := newMockServerEndpoint(c)
	ss1 := &serverStream{ctx: context.Background()}
	ss2 := &serverStream{ctx: context.Background()}
	_ = s.addStream("n1", ss1)

	// different entry: no removal
	s.removeStream("n1", ss2)
	_, ok := s.streams["n1"]
	convey.So(ok, convey.ShouldBeTrue)

	// matching entry: removed
	s.removeStream("n1", ss1)
	_, ok = s.streams["n1"]
	convey.So(ok, convey.ShouldBeFalse)

	// unknown node: no-op
	s.removeStream(testMissingNodeID, ss1)
}

func testGetStream() {
	c := newMockCoordinator()
	s := newMockServerEndpoint(c)
	ss := &serverStream{ctx: context.Background()}
	_ = s.addStream("n1", ss)
	got, ok := s.getStream("n1")
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(got, convey.ShouldEqual, ss)

	_, ok = s.getStream(testMissingNodeID)
	convey.So(ok, convey.ShouldBeFalse)
}

func testRegisterUUID() {
	c := newMockCoordinator()
	s := newMockServerEndpoint(c)
	ch := make(chan *proto.Response, 1)
	s.registerUUID(testAckKey, ch)
	got, ok := s.acks[testAckKey]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(got, convey.ShouldEqual, ch)

	s.unregisterUUID(testAckKey)
	_, ok = s.acks[testAckKey]
	convey.So(ok, convey.ShouldBeFalse)
}

func testRouteResp() {
	c := newMockCoordinator()
	s := newMockServerEndpoint(c)
	ch := make(chan *proto.Response, 1)
	s.registerUUID(testUUID+"|"+testNodeID, ch)
	resp := &proto.Response{Uuid: testUUID, Code: 0}
	s.routeResp(testNodeID, resp)
	select {
	case got := <-ch:
		convey.So(got, convey.ShouldEqual, resp)
	case <-time.After(time.Second):
		convey.So("timeout", convey.ShouldEqual, "no resp")
	}

	// no registered ack: silently ignored
	s.routeResp(testNodeID, &proto.Response{Uuid: "other", Code: 0})

	// ack channel full: hits the timeout branch (time.After mocked to fire)
	var p1 = gomonkey.ApplyFunc(time.After, func(d time.Duration) <-chan time.Time {
		ch := make(chan time.Time)
		close(ch)
		return ch
	})
	defer p1.Reset()
	fullCh := make(chan *proto.Response) // unbuffered, no receiver
	s.registerUUID(testUUID+"|node-full", fullCh)
	s.routeResp(testNodeFullID, &proto.Response{Uuid: testUUID, Code: 0})
}

func testSyncDataHandler() {
	c := newMockCoordinator()
	s := newMockServerEndpoint(c)

	convey.Convey("valid sync request", func() {
		req := &proto.SyncDataReq{Uuid: testUUID, NodeId: testNodeID,
			Containers: []*proto.ContainerInfo{mockContainerInfo(testJobID, testCtrID, common.StatusPaused)}}
		resp, err := s.SyncData(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp.Code, convey.ShouldEqual, 0)
		nodes := c.containerStore.GetNodesByJob([]string{testJobID})
		convey.So(nodes, convey.ShouldResemble, []string{testNodeID})
	})

	convey.Convey("missing node id", func() {
		req := &proto.SyncDataReq{Uuid: testUUID, NodeId: ""}
		resp, err := s.SyncData(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp.Code, convey.ShouldEqual, 1)
	})

	convey.Convey("too many containers", func() {
		var containers []*proto.ContainerInfo
		for i := 0; i < maxContainersPerSync+1; i++ {
			containers = append(containers, mockContainerInfo(testJobID, fmt.Sprintf("c%d", i), common.StatusPaused))
		}
		req := &proto.SyncDataReq{Uuid: testUUID, NodeId: testNodeID, Containers: containers}
		resp, err := s.SyncData(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp.Code, convey.ShouldEqual, 1)
	})
}

func testCoordinateHandler() {
	c := newMockCoordinator()
	s := newMockServerEndpoint(c)

	convey.Convey("valid coordinate request", func() {
		req := &proto.CoordinateReq{Uuid: testUUID, NodeId: testNodeID, JobIds: []string{testJobID}, Action: common.ActionStop}
		var p1 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "broadcastToOrdinary",
			func(_ *Coordinator, _ *proto.CoordinateReq) error { return nil })
		defer p1.Reset()
		resp, err := s.Coordinate(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp.Code, convey.ShouldEqual, 0)
	})

	convey.Convey("broadcast fails", func() {
		req := &proto.CoordinateReq{Uuid: testUUID, NodeId: testNodeID, JobIds: []string{testJobID}, Action: common.ActionStop}
		var p1 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "broadcastToOrdinary",
			func(_ *Coordinator, _ *proto.CoordinateReq) error { return testErr })
		defer p1.Reset()
		resp, err := s.Coordinate(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp.Code, convey.ShouldEqual, 1)
	})

	convey.Convey("invalid action", func() {
		req := &proto.CoordinateReq{Uuid: testUUID, NodeId: testNodeID, Action: testBogusAction}
		resp, err := s.Coordinate(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp.Code, convey.ShouldEqual, 1)
	})

	convey.Convey("missing node id", func() {
		req := &proto.CoordinateReq{Uuid: testUUID, NodeId: "", Action: common.ActionStop}
		resp, err := s.Coordinate(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp.Code, convey.ShouldEqual, 1)
	})
}

func testInitBroadcastStream() {
	c := newMockCoordinator()
	s := newMockServerEndpoint(c)

	convey.Convey("missing node id metadata", func() {
		stream := &mockServerStream{ctx: context.Background()}
		err := s.InitBroadcastStream(stream)
		convey.So(err, convey.ShouldResemble, errEmptyNodeID)
	})

	convey.Convey("normal stream lifecycle", func() {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(nodeIDMetadataKey, testNodeID))
		stream := &mockServerStream{ctx: ctx,
			recvSeq: []*proto.Response{{Uuid: testUUID, Code: 0}}}
		err := s.InitBroadcastStream(stream)
		convey.So(err, convey.ShouldNotBeNil) // Recv eventually returns closed error
		// stream must be unregistered afterwards
		_, ok := s.getStream(testNodeID)
		convey.So(ok, convey.ShouldBeFalse)
	})

	convey.Convey("replaces an existing stream", func() {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(nodeIDMetadataKey, testNodeID))
		oldCtx, cancel := context.WithCancel(context.Background())
		oldStream := &mockServerStream{ctx: ctx}
		s.streams[testNodeID] = &serverStream{stream: oldStream, ctx: oldCtx, cancel: cancel}
		stream := &mockServerStream{ctx: ctx, recvSeq: []*proto.Response{{Uuid: testUUID, Code: 0}}}
		_ = s.InitBroadcastStream(stream)
		convey.So(oldCtx.Err(), convey.ShouldNotBeNil) // old stream cancelled
	})

	convey.Convey("rejected when node limit reached", func() {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(nodeIDMetadataKey, testNodeID))
		for i := 0; i < common.MaxNodeNum; i++ {
			s.streams[fmt.Sprintf("n%d", i)] = &serverStream{ctx: context.Background()}
		}
		stream := &mockServerStream{ctx: ctx, recvSeq: []*proto.Response{{Uuid: testUUID, Code: 0}}}
		err := s.InitBroadcastStream(stream)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testSendCoordReq() {
	c := newMockCoordinator()
	s := newMockServerEndpoint(c)

	convey.Convey("no active stream", func() {
		req := &proto.CoordinateReq{Uuid: testUUID, NodeId: testNodeID}
		err := s.sendCoordReq(req, testMissingNodeID)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("send fails", func() {
		ss := &serverStream{stream: &mockServerStream{sendErr: testErr}, ctx: context.Background()}
		s.streams[testNodeID] = ss
		req := &proto.CoordinateReq{Uuid: testUUID, NodeId: testNodeID}
		err := s.sendCoordReq(req, testNodeID)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("success", func() {
		ss := &serverStream{stream: &mockServerStream{}, ctx: context.Background()}
		s.streams[testNodeID] = ss
		req := &proto.CoordinateReq{Uuid: testUUID, NodeId: testNodeID}
		go func() {
			time.Sleep(10 * time.Millisecond)
			s.routeResp(testNodeID, &proto.Response{Uuid: testUUID, Code: 0})
		}()
		err := s.sendCoordReq(req, testNodeID)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("leader returns non-zero code", func() {
		ss := &serverStream{stream: &mockServerStream{}, ctx: context.Background()}
		s.streams[testNodeID] = ss
		req := &proto.CoordinateReq{Uuid: testUUID, NodeId: testNodeID}
		go func() {
			time.Sleep(10 * time.Millisecond)
			s.routeResp(testNodeID, &proto.Response{Uuid: testUUID, Code: 1, Message: "rejected"})
		}()
		err := s.sendCoordReq(req, testNodeID)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("wait resp times out", func() {
		// pre-cancelled context makes waitResp return immediately
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		ss := &serverStream{stream: &mockServerStream{}, ctx: cancelledCtx}
		s.streams[testNodeID] = ss
		req := &proto.CoordinateReq{Uuid: testUUID, NodeId: testNodeID}
		err := s.sendCoordReq(req, testNodeID)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testWaitResp() {
	ss := &serverStream{ctx: context.Background()}
	ackCh := make(chan *proto.Response, 1)
	ackCh <- &proto.Response{Uuid: testUUID, Code: 0}
	resp, err := ss.waitResp(ackCh)
	convey.So(err, convey.ShouldBeNil)
	convey.So(resp.Code, convey.ShouldEqual, 0)

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	ss2 := &serverStream{ctx: cancelledCtx}
	_, err = ss2.waitResp(make(chan *proto.Response, 1))
	convey.So(err, convey.ShouldNotBeNil)
}

func testValidators() {
	convey.So(validateSyncDataReq(&proto.SyncDataReq{NodeId: ""}), convey.ShouldResemble, errEmptyNodeID)
	var many []*proto.ContainerInfo
	for i := 0; i < maxContainersPerSync+1; i++ {
		many = append(many, &proto.ContainerInfo{})
	}
	convey.So(validateSyncDataReq(&proto.SyncDataReq{NodeId: testNodeID, Containers: many}), convey.ShouldResemble, errTooManyContainers)
	convey.So(validateSyncDataReq(&proto.SyncDataReq{NodeId: testNodeID}), convey.ShouldBeNil)

	convey.So(validateCoordinateReq(&proto.CoordinateReq{NodeId: ""}), convey.ShouldResemble, errEmptyNodeID)
	convey.So(validateCoordinateReq(&proto.CoordinateReq{NodeId: testNodeID, Action: testBogusAction}), convey.ShouldResemble, errInvalidAction)
	convey.So(validateCoordinateReq(&proto.CoordinateReq{NodeId: testNodeID, Action: common.ActionStop}), convey.ShouldBeNil)
	convey.So(validateCoordinateReq(&proto.CoordinateReq{NodeId: testNodeID, Action: common.ActionStart}), convey.ShouldBeNil)
}

func testLimitQPS() {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) { return "ok", nil }
	info := &grpc.UnaryServerInfo{FullMethod: "/ContainerService/SyncData"}

	// normal path
	out, err := limitQPS(context.Background(), nil, info, handler)
	convey.So(err, convey.ShouldBeNil)
	convey.So(out, convey.ShouldEqual, "ok")

	// QPS exceeded
	var p1 = gomonkey.ApplyMethodReturn(limiter, "Allow", false)
	defer p1.Reset()
	_, err = limitQPS(context.Background(), nil, info, handler)
	convey.So(err, convey.ShouldNotBeNil)
}

func testExtractNodeID() {
	convey.Convey("metadata present", func() {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(nodeIDMetadataKey, testNodeID))
		stream := &mockServerStream{ctx: ctx}
		nodeID, err := extractNodeID(stream)
		convey.So(err, convey.ShouldBeNil)
		convey.So(nodeID, convey.ShouldEqual, testNodeID)
	})

	convey.Convey("no metadata", func() {
		stream := &mockServerStream{ctx: context.Background()}
		_, err := extractNodeID(stream)
		convey.So(err, convey.ShouldResemble, errEmptyNodeID)
	})

	convey.Convey("empty node id", func() {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(nodeIDMetadataKey, ""))
		stream := &mockServerStream{ctx: ctx}
		_, err := extractNodeID(stream)
		convey.So(err, convey.ShouldResemble, errEmptyNodeID)
	})
}
