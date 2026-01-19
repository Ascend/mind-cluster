/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
// package net is a Go package that provides a network tool for taskd.
package net

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"taskd/toolkit_backend/grpool"
	"taskd/toolkit_backend/net/common"
	"taskd/toolkit_backend/net/proto"
)

// downstreamEntry represents an entry for the downstream connection.
type downstreamEntry struct {
	mu          sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
	stream      proto.TaskNet_InitServerDownStreamServer
	entryLogger *hwlog.CustomLogger
}

// setEntry sets the stream, context, and cancel function for the downstream entry.
func (entry *downstreamEntry) setEntry(stream proto.TaskNet_InitServerDownStreamServer,
	ctx context.Context, cancel context.CancelFunc) {
	entry.mu.Lock()
	entry.stream = stream
	entry.ctx, entry.cancel = ctx, cancel
	entry.mu.Unlock()
}

// close cancels the context of the downstream entry.
func (entry *downstreamEntry) close() {
	entry.mu.Lock()
	defer entry.mu.Unlock()
	if entry.cancel != nil {
		entry.cancel()
	}
}

// ackWrapper wraps the acknowledgment and error.
type ackWrapper struct {
	ack *proto.Ack
	err error
}

// waitAck waits for an acknowledgment from the stream.
func (entry *downstreamEntry) waitAck(curPos *common.Position, msgId string) *ackWrapper {
	ctx, cancel := context.WithTimeout(entry.ctx, common.AckTimeout)
	defer cancel()
	resultChan := make(chan *ackWrapper, 1)
	go func() {
		ack, err := entry.stream.Recv()
		if err != nil {
			entry.entryLogger.Errorf("recv ack failed, msgid=%s, role=%s, srvRank=%s, processRank=%s, err=%v",
				msgId, curPos.Role, curPos.ServerRank, curPos.ProcessRank, err)
		}
		resultChan <- &ackWrapper{
			ack: ack,
			err: err,
		}
	}()
	select {
	case res := <-resultChan:
		return res
	case <-ctx.Done():
		entry.entryLogger.Errorf("recv ack timeout, msgid=%s, role=%s, srvRank=%s, processRank=%s",
			msgId, curPos.Role, curPos.ServerRank, curPos.ProcessRank)
		return &ackWrapper{
			ack: common.AckFrame(msgId, common.NetworkAckLost, curPos),
			err: errors.New("ack time out"),
		}
	}
}

// send sends a message through the stream and waits for an acknowledgment if necessary.
func (entry *downstreamEntry) send(msg *proto.Message, curPos *common.Position) (*proto.Ack, error) {
	entry.mu.Lock()
	defer entry.mu.Unlock()
	src := &common.Position{
		Role:        msg.Header.Src.Role,
		ServerRank:  msg.Header.Src.ServerRank,
		ProcessRank: msg.Header.Src.ProcessRank,
	}
	if entry.stream == nil {
		entry.entryLogger.Errorf("stream is nil, msgid=%s, role=%s, srvRank=%s, processRank=%s",
			msg.Header.Uuid, curPos.Role, curPos.ServerRank, curPos.ProcessRank)
		return common.AckFrame(msg.Header.Uuid, common.NetStreamNotInited, src),
			errors.New("stream not inited")
	}

	err := entry.stream.Send(msg)
	if err != nil {
		entry.entryLogger.Errorf("send msg failed, msgid=%s, role=%s, srvRank=%s, processRank=%s, err=%v",
			msg.Header.Uuid, curPos.Role, curPos.ServerRank, curPos.ProcessRank, err)
		return common.AckFrame(msg.Header.Uuid, common.NetworkSendLost, curPos),
			err
	}
	if !msg.Header.Sync {
		entry.entryLogger.Infof("send msg success, msgid=%s, role=%s, srvRank=%s, processRank=%s",
			msg.Header.Uuid, curPos.Role, curPos.ServerRank, curPos.ProcessRank)
		return common.AckFrame(msg.Header.Uuid, common.OK, curPos), nil
	}
	entry.entryLogger.Infof("start wait ack, msgid=%s, role=%s, srvRank=%s, processRank=%s, need wait ack",
		msg.Header.Uuid, curPos.Role, curPos.ServerRank, curPos.ProcessRank)
	ackWrap := entry.waitAck(curPos, msg.Header.Uuid)
	return ackWrap.ack, ackWrap.err
}

// downStreamEndpoint represents the endpoint for the downstream network.
type downStreamEndpoint struct {
	netInstance *NetInstance
	server      *grpc.Server
	destroyed   atomic.Bool
	rwLock      sync.RWMutex
	entryMap    map[common.Position]*downstreamEntry
	routeTable  map[common.Position]common.Position
	proto.UnimplementedTaskNetServer
}

// newDownStreamEndpoint creates a new downstream endpoint and starts the server.
func newDownStreamEndpoint(tool *NetInstance) (*downStreamEndpoint, error) {
	edp := &downStreamEndpoint{
		netInstance: tool,
		entryMap:    make(map[common.Position]*downstreamEntry),
		routeTable:  make(map[common.Position]common.Position),
	}
	return edp.startServer()
}

// limitQPS limits the QPS of gRPC requests.
func limitQPS(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if !limiter.Allow() {
		return nil, fmt.Errorf("qps exceeded, method=%s", info.FullMethod)
	}
	return handler(ctx, req)
}

// startServer starts the gRPC server for the downstream endpoint.
func (de *downStreamEndpoint) startServer() (*downStreamEndpoint, error) {
	if err := utils.IsHostValid(common.GetHostFromAddr(de.netInstance.config.ListenAddr)); err != nil {
		return nil, err
	}
	listen, err := net.Listen("tcp", de.netInstance.config.ListenAddr)
	if err != nil {
		return nil, err
	}
	keepAlive := keepalive.ServerParameters{
		Time:    common.KeepAlivePeriod,
		Timeout: common.KeepAliveTimeout,
	}
	de.server = grpc.NewServer(grpc.MaxRecvMsgSize(common.MaxGRPCRecvMsgSize),
		grpc.MaxSendMsgSize(common.MaxGRPCSendMsgSize),
		grpc.UnaryInterceptor(limitQPS), grpc.KeepaliveParams(keepAlive))
	proto.RegisterTaskNetServer(de.server, de)
	go func() {
		err = de.server.Serve(listen)
		de.netInstance.netlogger.Errorf("downstream server serve failed, err=%v", err)
	}()

	// Wait for grpc server ready
	for len(de.server.GetServiceInfo()) <= 0 {
		time.Sleep(time.Second)
	}
	return de, nil
}

// addDownStream adds a downstream connection.
func (de *downStreamEndpoint) addDownStream(pos common.Position,
	stream proto.TaskNet_InitServerDownStreamServer) error {
	de.rwLock.RLock()
	entry, exist := de.entryMap[pos]
	if !exist {
		de.rwLock.RUnlock()
		de.netInstance.netlogger.Errorf("downstream entry not exist, role=%s, srvRank=%s, processRank=%s",
			pos.Role, pos.ServerRank, pos.ProcessRank)
		return errors.New("un registry error")
	}
	de.rwLock.RUnlock()
	entry.close()
	ctx, cancel := context.WithCancel(de.netInstance.ctx)
	entry.setEntry(stream, ctx, cancel)

	select {
	case <-ctx.Done():
		return errors.New("entry context done")
	case <-stream.Context().Done():
		return errors.New("stream context done")
	}
}

/*
   next implement grpc interface
*/

// Register handles the registration request.
func (de *downStreamEndpoint) Register(ctx context.Context, req *proto.RegisterReq) (*proto.Ack, error) {
	de.netInstance.netlogger.Infof("recv register req, role=%s, srvRank=%s, processRank=%s",
		req.Pos.Role, req.Pos.ServerRank, req.Pos.ProcessRank)
	pos := common.Position{
		Role:        req.Pos.Role,
		ServerRank:  req.Pos.ServerRank,
		ProcessRank: req.Pos.ProcessRank,
	}
	de.rwLock.Lock()
	defer de.rwLock.Unlock()
	entry, exist := de.entryMap[pos]
	if exist {
		entry.close()
		return common.AckFrame(req.Uuid, common.OK, &de.netInstance.config.Pos), nil
	}
	if len(de.entryMap) >= common.MaxRegistryNum {
		de.netInstance.netlogger.Errorf("exceed max registry num, role=%s, srvRank=%s, processRank=%s",
			pos.Role, pos.ServerRank, pos.ProcessRank)
		return common.AckFrame(req.Uuid, common.ExceedMaxRegistryNum, &de.netInstance.config.Pos),
			errors.New("exceed max registry num")
	}
	de.entryMap[pos] = &downstreamEntry{entryLogger: de.netInstance.netlogger}
	return common.AckFrame(req.Uuid, common.OK, &de.netInstance.config.Pos), nil
}

func (de *downStreamEndpoint) PathDiscovery(ctx context.Context, req *proto.PathDiscoveryReq) (*proto.Ack, error) {
	de.netInstance.netlogger.Infof("recv path discovery req, role=%s, srvRank=%s, processRank=%s",
		req.ProxyPos.Role, req.ProxyPos.ServerRank, req.ProxyPos.ProcessRank)
	proxyPos := common.Position{
		Role:        req.ProxyPos.Role,
		ServerRank:  req.ProxyPos.ServerRank,
		ProcessRank: req.ProxyPos.ProcessRank,
	}
	pathPositions := make([]common.Position, len(req.Path))
	for i, pos := range req.Path {
		pathPositions[i] = common.Position{
			Role:        pos.Role,
			ServerRank:  pos.ServerRank,
			ProcessRank: pos.ProcessRank,
		}
	}
	de.rwLock.Lock()
	for _, pos := range pathPositions {
		de.routeTable[pos] = proxyPos
	}
	de.rwLock.Unlock()
	return de.netInstance.proxyPathDiscovery(ctx, req)
}

func (de *downStreamEndpoint) TransferMessage(ctx context.Context, msg *proto.Message) (*proto.Ack, error) {
	de.netInstance.netlogger.Infof("recv transfer message, msgid=%s, role=%s, srvRank=%s, processRank=%s",
		msg.Header.Uuid, de.netInstance.config.Pos.Role, de.netInstance.config.Pos.ServerRank,
		de.netInstance.config.Pos.ProcessRank)
	dst := common.Position{
		Role:        msg.Header.Dst.Role,
		ServerRank:  msg.Header.Dst.ServerRank,
		ProcessRank: msg.Header.Dst.ProcessRank,
	}
	dstType := common.DstCase(&de.netInstance.config.Pos, &dst)
	switch dstType {
	case common.Dst2Self, common.Dst2LowerLevel, common.Dst2SameLevel, common.Dst2UpperLevel:
		return de.netInstance.route(msg, dstType, common.DataFromLower)
	default:
		de.netInstance.netlogger.Errorf("dst type illegal, msgid=%s, role=%s, srvRank=%s, processRank=%s",
			msg.Header.Uuid, de.netInstance.config.Pos.Role, de.netInstance.config.Pos.ServerRank,
			de.netInstance.config.Pos.ProcessRank)
		return common.AckFrame(msg.Header.Uuid,
			common.DstTypeIllegal, &de.netInstance.config.Pos), errors.New("dst type illegal")
	}
}

func (de *downStreamEndpoint) InitServerDownStream(stream proto.TaskNet_InitServerDownStreamServer) error {
	de.netInstance.netlogger.Infof("recv init server down stream, role=%s, srvRank=%s, processRank=%s",
		de.netInstance.config.Pos.Role, de.netInstance.config.Pos.ServerRank,
		de.netInstance.config.Pos.ProcessRank)
	ctx := stream.Context()
	pos := common.Position{
		Role:        common.GetContextMetaData(ctx, common.MetaRoleKey),
		ServerRank:  common.GetContextMetaData(ctx, common.MetaServerRankKey),
		ProcessRank: common.GetContextMetaData(ctx, common.MetaProcessRankKey),
	}
	return de.addDownStream(pos, stream)
}

func (de *downStreamEndpoint) getBroadCastNextHops(msg *proto.Message) []common.Position {
	de.rwLock.RLock()
	defer de.rwLock.RUnlock()

	var res []common.Position
	for dst, proxy := range de.routeTable {
		if dst.Role != msg.Header.Dst.Role {
			continue
		}
		if dst.ServerRank != msg.Header.Dst.ServerRank && msg.Header.Dst.ServerRank != common.BroadCastPos {
			continue
		}
		if !common.RoleHasProcessProperty(msg.Header.Dst.Role) {
			res = append(res, proxy)
			continue
		}
		if dst.ProcessRank != msg.Header.Dst.ProcessRank && msg.Header.Dst.ProcessRank != common.BroadCastPos {
			continue
		}
		res = append(res, proxy)
	}
	return res
}

func (de *downStreamEndpoint) broadCast(msg *proto.Message) error {
	var success atomic.Bool
	success.Store(true)
	nextHops := de.getBroadCastNextHops(msg)
	group := de.netInstance.grPool.Group()
	for _, nextHop := range nextHops {
		hop := nextHop
		taskFunc := func(t grpool.Task) (interface{}, error) {
			ack, err := de.doSend(msg, hop)
			if err != nil {
				de.netInstance.netlogger.Debugf("broadcast failed, msgid=%s, role=%s, srvRank=%s, processRank=%s, err=%v",
					msg.Header.Uuid, de.netInstance.config.Pos.Role, de.netInstance.config.Pos.ServerRank,
					de.netInstance.config.Pos.ProcessRank, err)
				success.Store(false)
			}
			return ack, err
		}
		group.Submit(taskFunc)
	}
	group.WaitGroup()
	if success.Load() {
		return nil
	}
	return errors.New("broadcast error")
}

func (de *downStreamEndpoint) uniCast(msg *proto.Message) (*proto.Ack, error) {
	de.rwLock.RLock()
	nextHop, exist := de.routeTable[common.Position{
		Role:        msg.Header.Dst.Role,
		ServerRank:  msg.Header.Dst.ServerRank,
		ProcessRank: msg.Header.Dst.ProcessRank,
	}]
	if !exist {
		de.netInstance.netlogger.Errorf("next hop not found in route table, msgid=%s, role=%s, srvRank=%s, processRank=%s",
			msg.Header.Uuid, de.netInstance.config.Pos.Role, de.netInstance.config.Pos.ServerRank,
			de.netInstance.config.Pos.ProcessRank)
		de.rwLock.RUnlock()
		return common.AckFrame(msg.Header.Uuid, common.NoRoute, &de.netInstance.config.Pos),
			errors.New("no route")
	}
	de.rwLock.RUnlock()
	return de.doSend(msg, nextHop)
}

func (de *downStreamEndpoint) doSend(msg *proto.Message,
	nextHop common.Position) (*proto.Ack, error) {
	de.rwLock.RLock()
	entry, exist := de.entryMap[nextHop]
	de.rwLock.RUnlock()
	if !exist {
		de.netInstance.netlogger.Errorf("next hop entry not inited, msgid=%s, role=%s, srvRank=%s, processRank=%s",
			msg.Header.Uuid, de.netInstance.config.Pos.Role, de.netInstance.config.Pos.ServerRank,
			de.netInstance.config.Pos.ProcessRank)
		return common.AckFrame(msg.Header.Uuid, common.NoRoute, &de.netInstance.config.Pos),
			errors.New("no route")
	}
	return entry.send(msg, &de.netInstance.config.Pos)
}

func (de *downStreamEndpoint) send(msg *proto.Message) (*proto.Ack, error) {
	if common.IsBroadCast(msg.Header.Dst) {
		return nil, de.broadCast(msg)
	}
	return de.uniCast(msg)
}

func (de *downStreamEndpoint) close() {
	de.netInstance.netlogger.Infof("downstream endpoint close, role=%s, srvRank=%s, processRank=%s",
		de.netInstance.config.Pos.Role, de.netInstance.config.Pos.ServerRank,
		de.netInstance.config.Pos.ProcessRank)
	de.destroyed.Store(true)
	de.rwLock.Lock()
	defer de.rwLock.Unlock()
	if de.server != nil {
		de.server.Stop()
	}
	if len(de.entryMap) == 0 {
		return
	}
	for _, entry := range de.entryMap {
		entry.close()
	}
	de.entryMap = nil
	de.routeTable = nil
}
