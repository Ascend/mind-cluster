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

// Package app test for the coordinator application layer
package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"ascend-common/common-utils/hwlog"
	"container-manager/pkg/coordinator/proto"
)

var (
	testErr = errors.New("test error")
)

const (
	testNodeID  = "node-1"
	testPeerID  = "node-2"
	testListen  = "127.0.0.1:0"
	testJobID   = "job-1"
	testCtrID   = "ctr-1"
	testUUID    = "uuid-1"
	testLeaderA = "10.0.0.1:30000"
	testLeaderB = "10.0.0.2:30000"

	testMissingNodeID = "missing"
	testAckKey        = "key"
	testNodeFullID    = "node-full"
	testBogusAction   = "bogus"
)

func TestMain(m *testing.M) {
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

// mockOps is a configurable implementation of coordinator.ContainerOps.
type mockOps struct {
	containers []*proto.ContainerInfo
	changed    bool
	pauseErr   error
	resumeErr  error
}

func (m *mockOps) GetLocalContainers() []*proto.ContainerInfo { return m.containers }
func (m *mockOps) HasDataChanged() bool                       { return m.changed }
func (m *mockOps) PauseJobContainers(jobIDs, ctrIds []string, peerNodeID string) error {
	return m.pauseErr
}
func (m *mockOps) ResumeJobContainers(jobIDs, ctrIds []string, peerNodeID string) error {
	return m.resumeErr
}

func newMockCoordinator() *Coordinator {
	return NewCoordinator(context.Background(), &mockOps{})
}

// mockServerStream implements proto.ContainerService_InitBroadcastStreamServer.
type mockServerStream struct {
	ctx     context.Context
	mu      sync.Mutex
	recvSeq []*proto.Response
	recvIdx int
	sent    []*proto.CoordinateReq
	sendErr error
}

func (m *mockServerStream) Send(req *proto.CoordinateReq) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, req)
	return m.sendErr
}

func (m *mockServerStream) Recv() (*proto.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.recvIdx >= len(m.recvSeq) {
		return nil, errors.New("stream closed")
	}
	resp := m.recvSeq[m.recvIdx]
	m.recvIdx++
	return resp, nil
}

func (m *mockServerStream) SetHeader(metadata.MD) error  { return nil }
func (m *mockServerStream) SendHeader(metadata.MD) error { return nil }
func (m *mockServerStream) SetTrailer(metadata.MD)       {}
func (m *mockServerStream) Context() context.Context     { return m.ctx }
func (m *mockServerStream) SendMsg(interface{}) error    { return nil }
func (m *mockServerStream) RecvMsg(interface{}) error    { return nil }

// mockClientStream implements proto.ContainerService_InitBroadcastStreamClient.
type mockClientStream struct {
	mu      sync.Mutex
	recvSeq []*proto.CoordinateReq
	recvIdx int
	sent    []*proto.Response
	sendErr error
	closed  bool
}

func (m *mockClientStream) Send(resp *proto.Response) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, resp)
	return m.sendErr
}

func (m *mockClientStream) Recv() (*proto.CoordinateReq, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.recvIdx >= len(m.recvSeq) {
		return nil, errors.New("stream closed")
	}
	req := m.recvSeq[m.recvIdx]
	m.recvIdx++
	return req, nil
}

func (m *mockClientStream) CloseSend() error { m.closed = true; return nil }
func (m *mockClientStream) Header() (metadata.MD, error) {
	return nil, nil
}
func (m *mockClientStream) Trailer() metadata.MD { return nil }
func (m *mockClientStream) Context() context.Context {
	return context.Background()
}
func (m *mockClientStream) SendMsg(interface{}) error { return nil }
func (m *mockClientStream) RecvMsg(interface{}) error { return nil }

// mockContainerClient implements proto.ContainerServiceClient.
type mockContainerClient struct {
	initStream proto.ContainerService_InitBroadcastStreamClient
	initErr    error
	coordResp  *proto.Response
	coordErr   error
	syncResp   *proto.Response
	syncErr    error
}

func (m *mockContainerClient) SyncData(ctx context.Context, in *proto.SyncDataReq,
	opts ...grpc.CallOption) (*proto.Response, error) {
	if m.syncErr != nil {
		return nil, m.syncErr
	}
	return m.syncResp, nil
}

func (m *mockContainerClient) Coordinate(ctx context.Context, in *proto.CoordinateReq,
	opts ...grpc.CallOption) (*proto.Response, error) {
	if m.coordErr != nil {
		return nil, m.coordErr
	}
	return m.coordResp, nil
}

func (m *mockContainerClient) InitBroadcastStream(ctx context.Context,
	opts ...grpc.CallOption) (proto.ContainerService_InitBroadcastStreamClient, error) {
	if m.initErr != nil {
		return nil, m.initErr
	}
	return m.initStream, nil
}

// newMockServerEndpoint builds a serverEndpoint without starting a real gRPC server.
func newMockServerEndpoint(c *Coordinator) *serverEndpoint {
	return &serverEndpoint{
		coord:   c,
		streams: make(map[string]*serverStream),
		acks:    make(map[string]chan *proto.Response),
	}
}

func mockContainerInfo(jobID, ctrID, status string) *proto.ContainerInfo {
	return &proto.ContainerInfo{
		JobId:         jobID,
		ContainerId:   ctrID,
		Status:        status,
		JobReplica:    1,
		EnableRecover: true,
		PhyIds:        []int32{0},
	}
}
