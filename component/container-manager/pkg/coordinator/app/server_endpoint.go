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
	"fmt"
	"net"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"container-manager/pkg/common"
	"container-manager/pkg/coordinator/proto"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// Server-side gRPC constants.
const (
	keepAlivePeriod      = 5 * time.Second
	keepAliveTimeout     = time.Second
	grpcQps              = 10000
	maxContainersPerSync = 128
)

// Validation errors for gRPC requests.
var (
	errEmptyNodeID       = errors.New("node-id is required")
	errTooManyContainers = errors.New("too many containers in sync request")
	errInvalidAction     = errors.New("action must be stop or start")
	errNoStream          = errors.New("no active broadcast stream for node")
)

// limiter limits the QPS of gRPC requests.
var limiter = rate.NewLimiter(rate.Limit(grpcQps), grpcQps)

// serverStream is the per-node stream entry
type serverStream struct {
	stream proto.ContainerService_InitBroadcastStreamServer
	sendMu sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

// send pushes req through the stream under the per-stream send lock.
func (ss *serverStream) send(req *proto.CoordinateReq) error {
	ss.sendMu.Lock()
	defer ss.sendMu.Unlock()
	if ss.stream == nil {
		return errNoStream
	}
	return ss.stream.Send(req)
}

// send pushes req through the stream under the per-stream send lock.
func (ss *serverStream) close() {
	ss.sendMu.Lock()
	defer ss.sendMu.Unlock()
	if ss.cancel != nil {
		ss.cancel()
	}
}

// serverEndpoint is the leader-side gRPC endpoint
type serverEndpoint struct {
	coord   *Coordinator // back-reference to the facade (like taskd's netInstance)
	server  *grpc.Server
	rwLock  sync.RWMutex
	streams map[string]*serverStream        // nodeID -> per-node stream entry
	acks    map[string]chan *proto.Response // uuid -> ACK collector
	proto.UnimplementedContainerServiceServer
}

// newServerEndpoint creates the leader-side endpoint.
func newServerEndpoint(c *Coordinator) (*serverEndpoint, error) {
	s := &serverEndpoint{
		coord:   c,
		streams: make(map[string]*serverStream),
		acks:    make(map[string]chan *proto.Response),
	}
	return s.startServer()
}

// startServer listens and serves gRPC
func (s *serverEndpoint) startServer() (*serverEndpoint, error) {
	hwlog.RunLog.Infof("server listen on :%s", common.ParamOption.ListenAddr)
	listen, err := net.Listen("tcp", common.ParamOption.ListenAddr)
	if err != nil {
		return s, err
	}
	keepAlive := keepalive.ServerParameters{
		Time:    keepAlivePeriod,
		Timeout: keepAliveTimeout,
	}
	s.server = grpc.NewServer(
		grpc.MaxRecvMsgSize(grpcMsgSizeLimit),
		grpc.MaxSendMsgSize(grpcMsgSizeLimit),
		grpc.UnaryInterceptor(limitQPS), grpc.KeepaliveParams(keepAlive),
	)
	proto.RegisterContainerServiceServer(s.server, s)

	go func() {
		err := s.server.Serve(listen)
		hwlog.RunLog.Errorf("server stopped: %v", err)
	}()
	for len(s.server.GetServiceInfo()) <= 0 {
		time.Sleep(time.Second)
	}
	return s, nil
}

// close stops the gRPC server and drops the stream/ACK registries.
func (s *serverEndpoint) close() {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	if s.server != nil {
		s.server.Stop()
	}
	for _, ss := range s.streams {
		ss.close()
	}
	s.streams = nil
	s.acks = nil
}

// addStream records an inbound broadcast stream from nodeID.
func (s *serverEndpoint) addStream(nodeID string, ss *serverStream) error {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	if _, exists := s.streams[nodeID]; !exists && len(s.streams) >= common.MaxNodeNum {
		return fmt.Errorf("node count exceeds limit %d, node %s rejected", common.MaxNodeNum, nodeID)
	}
	s.streams[nodeID] = ss
	hwlog.RunLog.Infof("stream registered from node %s", nodeID)
	return nil
}

// removeStream removes nodeID's stream, but only if it is the same entry that
// the caller registered.
func (s *serverEndpoint) removeStream(nodeID string, ss *serverStream) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	if cur, ok := s.streams[nodeID]; ok && cur == ss {
		delete(s.streams, nodeID)
		hwlog.RunLog.Infof("broadcast stream unregistered for node %s", nodeID)
	}
}

// getStream returns nodeID's stream entry, if any. The caller uses it to send
// (ss.send) — the node is treated as failed when no stream is active.
func (s *serverEndpoint) getStream(nodeID string) (*serverStream, bool) {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()
	ss, ok := s.streams[nodeID]
	return ss, ok
}

// registerUUID records a collector that broadcastToOne waits on.
func (s *serverEndpoint) registerUUID(uuidKey string, ch chan *proto.Response) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	s.acks[uuidKey] = ch
}

// unregisterUUID removes a collector.
func (s *serverEndpoint) unregisterUUID(uuidKey string) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	delete(s.acks, uuidKey)
}

// routeResp delivers a Response from nodeID back to the broadcaster
func (s *serverEndpoint) routeResp(nodeID string, resp *proto.Response) {
	s.rwLock.Lock()
	ch, ok := s.acks[resp.Uuid+"|"+nodeID]
	s.rwLock.Unlock()
	if !ok {
		return
	}
	select {
	case ch <- resp:
	case <-time.After(routeTimeout):
		hwlog.RunLog.Errorf("timeout for ack %s from %s", resp.Uuid, nodeID)
	}
}

// SyncData applies an ordinary node's container snapshot to the store.
func (s *serverEndpoint) SyncData(ctx context.Context, req *proto.SyncDataReq) (*proto.Response, error) {
	hwlog.RunLog.Infof("receive sync data req %s from node %s: %d containers", req.Uuid, req.NodeId, len(req.Containers))
	if err := validateSyncDataReq(req); err != nil {
		hwlog.RunLog.Errorf("validate sync data req %s from node %s failed: %v", req.Uuid, req.NodeId, err)
		return &proto.Response{Uuid: req.Uuid, Code: 1, Message: err.Error()}, nil
	}
	s.coord.containerStore.ApplySync(req)
	return &proto.Response{Uuid: req.Uuid, Code: 0}, nil
}

// Coordinate handles a stop/start request from an ordinary node.
func (s *serverEndpoint) Coordinate(ctx context.Context, req *proto.CoordinateReq) (*proto.Response, error) {
	hwlog.RunLog.Infof("receive coordinate req %s", req.String())
	if err := validateCoordinateReq(req); err != nil {
		hwlog.RunLog.Errorf("validate coordinate req %s failed: %v", req.String(), err)
		return &proto.Response{Uuid: req.Uuid, Code: 1, Message: err.Error()}, nil
	}
	if err := s.coord.broadcastToOrdinary(req); err != nil {
		hwlog.RunLog.Errorf("broadcast req %s to ordinary node failed: %v", req.String(), err)
		return &proto.Response{Uuid: req.Uuid, Code: 1, Message: err.Error()}, nil
	}
	return &proto.Response{Uuid: req.Uuid, Code: 0}, nil
}

// InitBroadcastStream handles the bidirectional stream from an ordinary node.
func (s *serverEndpoint) InitBroadcastStream(stream proto.ContainerService_InitBroadcastStreamServer) error {
	nodeID, err := extractNodeID(stream)
	if err != nil {
		hwlog.RunLog.Errorf("reject stream: %v", err)
		return err
	}
	oldStream, ok := s.getStream(nodeID)
	if ok && oldStream != nil {
		oldStream.close()
	}
	ctx, cancelFunc := context.WithCancel(s.coord.ctx)
	curStream := &serverStream{stream: stream, ctx: ctx, cancel: cancelFunc}
	if err := s.addStream(nodeID, curStream); err != nil {
		curStream.close()
		hwlog.RunLog.Errorf("reject stream for node %s: %v", nodeID, err)
		return err
	}
	defer s.removeStream(nodeID, curStream)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-stream.Context().Done():
			return stream.Context().Err()
		default:
			resp, err := stream.Recv()
			if err != nil {
				hwlog.RunLog.Infof("stream for node %s closed: %v", nodeID, err)
				return err
			}
			s.routeResp(nodeID, resp)
		}
	}
}

// sendCoordReq sends req to a single node and waits for its Response.
func (s *serverEndpoint) sendCoordReq(req *proto.CoordinateReq, nodeID string) error {
	ss, ok := s.getStream(nodeID)
	if !ok {
		return fmt.Errorf("node %s has no active stream", nodeID)
	}

	ackCh := make(chan *proto.Response, 1)
	uuidKey := req.Uuid + "|" + nodeID
	s.registerUUID(uuidKey, ackCh)
	defer s.unregisterUUID(uuidKey)

	if err := ss.send(req); err != nil {
		return fmt.Errorf("send to %s failed: %v", nodeID, err)
	}
	resp, err := ss.waitResp(ackCh)
	if err != nil || (resp != nil && resp.GetCode() != 0) {
		return fmt.Errorf("node %s returned code %d: %s", nodeID, resp.GetCode(), resp.GetMessage())
	}
	return nil
}

// waitResp waits for a node's Response on ackCh
func (ss *serverStream) waitResp(ackCh chan *proto.Response) (*proto.Response, error) {
	ctx, cancel := context.WithTimeout(ss.ctx, responseTimeout)
	defer cancel()
	select {
	case resp := <-ackCh:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func validateSyncDataReq(req *proto.SyncDataReq) error {
	if req.NodeId == "" {
		return errEmptyNodeID
	}
	if len(req.Containers) > maxContainersPerSync {
		return errTooManyContainers
	}
	return nil
}

func validateCoordinateReq(req *proto.CoordinateReq) error {
	if req.NodeId == "" {
		return errEmptyNodeID
	}
	if req.Action != common.ActionStop && req.Action != common.ActionStart {
		return errInvalidAction
	}
	return nil
}

// limitQPS limits the QPS of gRPC requests.
func limitQPS(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if !limiter.Allow() {
		return nil, fmt.Errorf("qps exceeded, method=%s", info.FullMethod)
	}
	return handler(ctx, req)
}

func extractNodeID(stream proto.ContainerService_InitBroadcastStreamServer) (string, error) {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return "", errEmptyNodeID
	}
	vals := md.Get(nodeIDMetadataKey)
	if len(vals) == 0 || vals[0] == "" {
		return "", errEmptyNodeID
	}
	return vals[0], nil
}
