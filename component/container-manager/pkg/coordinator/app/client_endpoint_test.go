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
	"errors"
	"net"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"

	"container-manager/pkg/common"
	"container-manager/pkg/coordinator/proto"
)

func TestClientEndpoint(t *testing.T) {
	convey.Convey("test sendRespWithTimeout", t, testSendRespWithTimeout)
	convey.Convey("test leaderEntry close", t, testLeaderEntryClose)
	convey.Convey("test newClientEndpoint and start", t, testNewClientEndpoint)
	convey.Convey("test connectToLeader", t, testConnectToLeader)
	convey.Convey("test handleLeaderStream", t, testHandleLeaderStream)
	convey.Convey("test handleBroadcast", t, testHandleBroadcast)
	convey.Convey("test get/reset leader entry", t, testGetResetLeaderEntry)
	convey.Convey("test callLeaderForCoord", t, testCallLeaderForCoord)
	convey.Convey("test callLeaderForSyncData", t, testCallLeaderForSyncData)
	convey.Convey("test backoff and nextBackoff", t, testBackoff)
	convey.Convey("test client close", t, testClientClose)
}

// blockingStream Send blocks until released so the timeout branch can be exercised.
type blockingStream struct {
	mockClientStream
	release chan struct{}
}

func (b *blockingStream) Send(*proto.Response) error {
	<-b.release
	return nil
}

func testSendRespWithTimeout() {
	convey.Convey("send succeeds", func() {
		stream := &mockClientStream{}
		entry := &leaderEntry{stream: stream}
		err := entry.sendRespWithTimeout(&proto.Response{Uuid: testUUID, Code: 0})
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("send times out", func() {
		stream := &blockingStream{release: make(chan struct{})}
		entry := &leaderEntry{stream: stream}
		// time.After is inlined into sendRespWithTimeout as NewTimer(d).C, so
		// patch NewTimer to return an already-fired timer instead of patching
		// time.After (which the inlined call site never hits). This works both
		// with inlining enabled and disabled.
		var p1 = gomonkey.ApplyFunc(time.NewTimer, func(_ time.Duration) *time.Timer {
			ch := make(chan time.Time, 1)
			ch <- time.Time{}
			return &time.Timer{C: ch}
		})
		defer p1.Reset()
		err := entry.sendRespWithTimeout(&proto.Response{Uuid: testUUID, Code: 0})
		convey.So(err, convey.ShouldNotBeNil)
		close(stream.release) // release the blocked Send goroutine
	})
}

func testLeaderEntryClose() {
	convey.Convey("nil entry", func() {
		var e *leaderEntry
		e.close()
	})

	convey.Convey("full teardown", func() {
		ctx, cancel := context.WithCancel(context.Background())
		stream := &mockClientStream{}
		conn := &grpc.ClientConn{}
		var p1 = gomonkey.ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil)
		defer p1.Reset()
		entry := &leaderEntry{cancel: cancel, stream: stream, conn: conn}
		entry.close()
		convey.So(stream.closed, convey.ShouldBeTrue)
		convey.So(ctx.Err(), convey.ShouldNotBeNil)
	})
}

func testNewClientEndpoint() {
	convey.Convey("no leaders", func() {
		resetParamOption()
		c := newMockCoordinator()
		client, err := newClientEndpoint(c)
		convey.So(err, convey.ShouldBeNil)
		convey.So(client, convey.ShouldNotBeNil)
		convey.So(len(client.entries), convey.ShouldEqual, 0)
	})

	convey.Convey("skips local leader and starts workers for others", func() {
		resetParamOption()
		common.ParamOption.ListenAddr = testLeaderA
		common.ParamOption.LeaderAddrs = []string{testLeaderA, testLeaderB}
		common.ParamOption.LocalNodeID = testNodeID
		c := newMockCoordinator()
		var p1 = gomonkey.ApplyPrivateMethod(&clientEndpoint{}, "streamWorker",
			func(ce *clientEndpoint, _ context.Context, _ string) { ce.wg.Done() })
		defer p1.Reset()
		client, err := newClientEndpoint(c)
		convey.So(err, convey.ShouldBeNil)
		convey.So(client, convey.ShouldNotBeNil)
		client.wg.Wait()
	})
}

func testConnectToLeader() {
	convey.Convey("new client fails", func() {
		resetParamOption()
		c := newMockCoordinator()
		client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{}}
		var p1 = gomonkey.ApplyFuncReturn(grpc.NewClient, nil, testErr)
		defer p1.Reset()
		_, err := client.connectToLeader(context.Background(), testLeaderA)
		convey.So(err, convey.ShouldResemble, testErr)
	})

	convey.Convey("init stream fails when leader is down", func() {
		resetParamOption()
		common.ParamOption.LocalNodeID = testNodeID
		c := newMockCoordinator()
		s, addr := startLiveServer(c)
		// Stop the server so the dial/RPC fails at call time with a fast
		// connection-refused error instead of a long timeout.
		s.server.Stop()
		client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{}}
		_, err := client.connectToLeader(context.Background(), addr)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("connects to live leader and registers entry", func() {
		resetParamOption()
		common.ParamOption.LocalNodeID = testNodeID
		c := newMockCoordinator()
		s, addr := startLiveServer(c)
		defer s.server.Stop()
		client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{}}
		entry, err := client.connectToLeader(context.Background(), addr)
		convey.So(err, convey.ShouldBeNil)
		convey.So(entry, convey.ShouldNotBeNil)
		convey.So(entry.stream, convey.ShouldNotBeNil)
		convey.So(client.entries[addr], convey.ShouldEqual, entry)
		client.resetLeaderEntry(addr)
	})
}

// startLiveServer starts a real in-process gRPC server on 127.0.0.1:0 and
// waits until it is ready to accept connections.
func startLiveServer(c *Coordinator) (*serverEndpoint, string) {
	s := &serverEndpoint{coord: c, streams: make(map[string]*serverStream), acks: make(map[string]chan *proto.Response)}
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	s.server = grpc.NewServer()
	proto.RegisterContainerServiceServer(s.server, s)
	go s.server.Serve(lis)
	addr := lis.Addr().String()
	for {
		conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	return s, addr
}

func testHandleLeaderStream() {
	convey.Convey("no active stream", func() {
		resetParamOption()
		c := newMockCoordinator()
		client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{}}
		err := client.handleLeaderStream(testLeaderA)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("recv returns error", func() {
		resetParamOption()
		c := newMockCoordinator()
		stream := &mockClientStream{}
		client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{
			testLeaderA: {addr: testLeaderA, stream: stream},
		}}
		err := client.handleLeaderStream(testLeaderA)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("handles broadcast and replies", func() {
		resetParamOption()
		c := newMockCoordinator()
		req := &proto.CoordinateReq{Uuid: testUUID, JobIds: []string{testJobID}, Action: common.ActionStop}
		stream := &mockClientStream{recvSeq: []*proto.CoordinateReq{req}}
		client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{
			testLeaderA: {addr: testLeaderA, stream: stream},
		}}
		err := client.handleLeaderStream(testLeaderA)
		convey.So(err, convey.ShouldNotBeNil) // stream ends after the single request
		convey.So(len(stream.sent), convey.ShouldEqual, 1)
		convey.So(stream.sent[0].GetCode(), convey.ShouldEqual, 0)
	})

	convey.Convey("send resp fails", func() {
		resetParamOption()
		c := newMockCoordinator()
		req := &proto.CoordinateReq{Uuid: testUUID}
		stream := &mockClientStream{recvSeq: []*proto.CoordinateReq{req}}
		client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{
			testLeaderA: {addr: testLeaderA, stream: stream},
		}}
		var p1 = gomonkey.ApplyPrivateMethod(&leaderEntry{}, "sendRespWithTimeout",
			func(_ *leaderEntry, _ *proto.Response) error { return testErr })
		defer p1.Reset()
		err := client.handleLeaderStream(testLeaderA)
		convey.So(err, convey.ShouldResemble, testErr)
	})
}

func testHandleBroadcast() {
	convey.Convey("nil request", func() {
		resetParamOption()
		c := newMockCoordinator()
		client := &clientEndpoint{coord: c}
		resp := client.handleBroadcast(nil)
		convey.So(resp.GetCode(), convey.ShouldEqual, 1)
	})

	convey.Convey("execute fails", func() {
		resetParamOption()
		ops := &mockOps{pauseErr: testErr}
		c := NewCoordinator(context.Background(), ops)
		client := &clientEndpoint{coord: c}
		resp := client.handleBroadcast(&proto.CoordinateReq{Uuid: testUUID, Action: common.ActionStop})
		convey.So(resp.GetCode(), convey.ShouldEqual, 1)
		convey.So(resp.GetMessage(), convey.ShouldContainSubstring, "test error")
	})

	convey.Convey("success", func() {
		resetParamOption()
		c := newMockCoordinator()
		client := &clientEndpoint{coord: c}
		resp := client.handleBroadcast(&proto.CoordinateReq{Uuid: testUUID, Action: common.ActionStop})
		convey.So(resp.GetCode(), convey.ShouldEqual, 0)
	})
}

func testGetResetLeaderEntry() {
	resetParamOption()
	c := newMockCoordinator()
	entry := &leaderEntry{addr: testLeaderA}
	client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{testLeaderA: entry}}

	got, ok := client.getLeaderEntry(testLeaderA)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(got, convey.ShouldEqual, entry)

	_, ok = client.getLeaderEntry(testLeaderB)
	convey.So(ok, convey.ShouldBeFalse)

	client.resetLeaderEntry(testLeaderA)
	_, ok = client.getLeaderEntry(testLeaderA)
	convey.So(ok, convey.ShouldBeFalse)
}

func testCallLeaderForCoord() {
	convey.Convey("no entry", func() {
		resetParamOption()
		c := newMockCoordinator()
		client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{}}
		err := client.callLeaderForCoord(testLeaderA, &proto.CoordinateReq{Uuid: testUUID})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("rejected by leader", func() {
		resetParamOption()
		c := newMockCoordinator()
		entry := &leaderEntry{
			addr:   testLeaderA,
			ctx:    context.Background(),
			client: &mockContainerClient{coordResp: &proto.Response{Code: 1, Message: "rejected"}},
		}
		client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{testLeaderA: entry}}
		err := client.callLeaderForCoord(testLeaderA, &proto.CoordinateReq{Uuid: testUUID})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("success", func() {
		resetParamOption()
		c := newMockCoordinator()
		entry := &leaderEntry{
			addr:   testLeaderA,
			ctx:    context.Background(),
			client: &mockContainerClient{coordResp: &proto.Response{Code: 0}},
		}
		client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{testLeaderA: entry}}
		err := client.callLeaderForCoord(testLeaderA, &proto.CoordinateReq{Uuid: testUUID})
		convey.So(err, convey.ShouldBeNil)
	})
}

func testCallLeaderForSyncData() {
	convey.Convey("no entry", func() {
		resetParamOption()
		c := newMockCoordinator()
		client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{}}
		err := client.callLeaderForSyncData(context.Background(), testLeaderA, &proto.SyncDataReq{})
		convey.So(errors.Is(err, errNoActiveClient), convey.ShouldBeTrue)
	})

	convey.Convey("rejected by leader", func() {
		resetParamOption()
		c := newMockCoordinator()
		entry := &leaderEntry{
			addr:   testLeaderA,
			client: &mockContainerClient{syncResp: &proto.Response{Code: 1, Message: "rejected"}},
		}
		client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{testLeaderA: entry}}
		err := client.callLeaderForSyncData(context.Background(), testLeaderA, &proto.SyncDataReq{})
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("success", func() {
		resetParamOption()
		c := newMockCoordinator()
		entry := &leaderEntry{
			addr:   testLeaderA,
			client: &mockContainerClient{syncResp: &proto.Response{Code: 0}},
		}
		client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{testLeaderA: entry}}
		err := client.callLeaderForSyncData(context.Background(), testLeaderA, &proto.SyncDataReq{})
		convey.So(err, convey.ShouldBeNil)
	})
}

func testBackoff() {
	convey.Convey("nextBackoff doubles", func() {
		convey.So(nextBackoff(time.Second), convey.ShouldEqual, 2*time.Second)
	})

	convey.Convey("nextBackoff caps at max", func() {
		convey.So(nextBackoff(reconnectMaxDelay), convey.ShouldEqual, reconnectMaxDelay)
	})

	convey.Convey("backoff with cancelled ctx returns fast", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		delay := time.Hour
		start := time.Now()
		backoff(ctx, &delay)
		convey.So(time.Since(start), convey.ShouldBeLessThan, time.Second)
	})

	convey.Convey("backoff sleeps and doubles", func() {
		delay := time.Millisecond
		start := time.Now()
		backoff(context.Background(), &delay)
		convey.So(time.Since(start), convey.ShouldBeGreaterThanOrEqualTo, time.Millisecond)
		convey.So(delay, convey.ShouldEqual, 2*time.Millisecond)
	})
}

func testClientClose() {
	resetParamOption()
	c := newMockCoordinator()
	stream := &mockClientStream{}
	entry := &leaderEntry{addr: testLeaderA, stream: stream}
	client := &clientEndpoint{coord: c, entries: map[string]*leaderEntry{testLeaderA: entry}}
	client.close()
	convey.So(len(client.entries), convey.ShouldEqual, 0)
	convey.So(stream.closed, convey.ShouldBeTrue)
}
