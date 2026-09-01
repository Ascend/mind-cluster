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
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"container-manager/pkg/common"
	"container-manager/pkg/coordinator/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	grpcMsgSizeLimit       = 8 * 1024 * 1024
	reconnectInitDelay     = 1 * time.Second
	reconnectMaxDelay      = 30 * time.Second
	reconnectBackoffFactor = 2
	nodeIDMetadataKey      = "node-id"
	responseTimeout        = 10 * time.Second
	routeTimeout           = 3 * time.Second
)

// errNoActiveClient marks that a leader is not reachable yet
var errNoActiveClient = errors.New("no active client to leader")

// clientEndpoint is the ordinary-node gRPC client endpoint
type clientEndpoint struct {
	coord *Coordinator

	mu      sync.RWMutex
	entries map[string]*leaderEntry // addr -> per-leader conn/client/stream
	wg      sync.WaitGroup          // per-leader lifecycle goroutines
}

// leaderEntry holds one leader's persistent gRPC state.
type leaderEntry struct {
	addr   string
	conn   *grpc.ClientConn
	client proto.ContainerServiceClient
	stream proto.ContainerService_InitBroadcastStreamClient
	ctx    context.Context
	cancel context.CancelFunc // cancels this leader's per-stream ctx (active teardown)
}

// timeoutAckUpStream sends an acknowledgment with a timeout.
func (l *leaderEntry) sendRespWithTimeout(resp *proto.Response) error {
	done := make(chan error, 1)

	go func() {
		done <- l.stream.Send(resp)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(responseTimeout):
		return errors.New("send ack time out")
	}
}

// close tears down the leader's gRPC state.
func (l *leaderEntry) close() {
	if l == nil {
		return
	}
	if l.cancel != nil {
		l.cancel()
	}
	if l.stream != nil {
		l.stream.CloseSend()
	}
	if l.conn != nil {
		l.conn.Close()
	}
}

// newClientEndpoint builds a clientEndpoint.
func newClientEndpoint(c *Coordinator) (*clientEndpoint, error) {
	client := &clientEndpoint{
		coord:   c,
		entries: make(map[string]*leaderEntry),
	}
	return client.start(c.ctx)
}

// start starts the worker pool and connects to all non-local leaders.
func (c *clientEndpoint) start(ctx context.Context) (*clientEndpoint, error) {
	for _, addr := range common.ParamOption.LeaderAddrs {
		if c.coord.isLocalLeader(addr) {
			hwlog.RunLog.Infof("skip connect to local leader %s", addr)
			continue
		}
		c.wg.Add(1)
		go c.streamWorker(ctx, addr)
	}
	return c, nil
}

// streamWorker is the per-leader lifecycle goroutine
func (c *clientEndpoint) streamWorker(ctx context.Context, leaderAddr string) {
	defer c.wg.Done()
	delay := reconnectInitDelay
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if _, err := c.connectToLeader(ctx, leaderAddr); err != nil {
				hwlog.RunLog.Warnf("join leader %s failed: %v, retry in %v", leaderAddr, err, delay)
				backoff(ctx, &delay)
				continue
			}
			delay = reconnectInitDelay
			if err := c.handleLeaderStream(leaderAddr); err != nil {
				hwlog.RunLog.Warnf("recv from %s failed: %v", leaderAddr, err)
			}
			c.resetLeaderEntry(leaderAddr)
			hwlog.RunLog.Warnf("stream to %s closed, reconnect in %v", leaderAddr, delay)
			backoff(ctx, &delay)
		}
	}
}

// connectToLeader dial leader and init broadcast stream, construct leaderEntry
// The addr argument is a full leader address in ip:port form (from LeaderAddrs).
func (c *clientEndpoint) connectToLeader(ctx context.Context, addr string) (*leaderEntry, error) {
	hwlog.RunLog.Infof("connect to leader %s", addr)
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(grpcMsgSizeLimit),
			grpc.MaxCallSendMsgSize(grpcMsgSizeLimit),
		),
	)
	if err != nil {
		return nil, err
	}
	client := proto.NewContainerServiceClient(conn)

	streamCtx, cancel := context.WithCancel(ctx)
	stream, err := client.InitBroadcastStream(
		metadata.AppendToOutgoingContext(streamCtx, nodeIDMetadataKey, common.ParamOption.LocalNodeID))
	if err != nil {
		cancel()
		conn.Close()
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[addr] = &leaderEntry{addr: addr, conn: conn, client: client, stream: stream, ctx: streamCtx, cancel: cancel}
	hwlog.RunLog.Infof("stream to %s connected", addr)

	return c.entries[addr], nil
}

// handleLeaderStream receives CoordinateReq messages from the leader stream one at a
// time and hands each to dispatch().
func (c *clientEndpoint) handleLeaderStream(addr string) error {
	e, exist := c.getLeaderEntry(addr)
	if !exist || e == nil || e.stream == nil {
		return fmt.Errorf("no active stream to %s", addr)
	}
	for {
		req, err := e.stream.Recv()
		if err != nil {
			return err
		}
		resp := c.handleBroadcast(req)
		if resp == nil {
			resp = &proto.Response{Uuid: req.Uuid, Code: 1, Message: "nil response"}
		}
		if err := e.sendRespWithTimeout(resp); err != nil {
			hwlog.RunLog.Errorf("resp %s to %s failed: %v", req.Uuid, addr, err)
			return err
		}
	}
}

// handleBroadcast executes a broadcast CoordinateReq locally.
func (c *clientEndpoint) handleBroadcast(req *proto.CoordinateReq) *proto.Response {
	hwlog.RunLog.Infof("ordinary node receive coordinate req %s", req.String())
	if req == nil {
		hwlog.RunLog.Errorf("handle broadcast req is nil")
		return &proto.Response{Code: 1, Message: "nil request"}
	}
	if err := c.coord.executeLocal(req); err != nil {
		hwlog.RunLog.Errorf("execute local coordinate req %s failed: %v", req.Uuid, err)
		return &proto.Response{Uuid: req.Uuid, Code: 1, Message: err.Error()}
	}
	hwlog.RunLog.Infof("handle coordinate req %s success", req.Uuid)
	return &proto.Response{Uuid: req.Uuid, Code: 0}
}

func (c *clientEndpoint) getLeaderEntry(addr string) (*leaderEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[addr]
	return e, ok
}

// resetLeaderEntry tears down a leader's stream
func (c *clientEndpoint) resetLeaderEntry(addr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.entries[addr]; ok && e != nil {
		e.close()
	}
	delete(c.entries, addr)
}

func (c *clientEndpoint) callLeaderForCoord(addr string, req *proto.CoordinateReq) error {
	hwlog.RunLog.Infof("call leader %s for coordinate %s", addr, req.String())
	e, ok := c.getLeaderEntry(addr)
	if !ok || e == nil || e.client == nil {
		return fmt.Errorf("no active client to %s", addr)
	}
	ctx, cancel := context.WithTimeout(e.ctx, responseTimeout)
	defer cancel()
	resp, err := e.client.Coordinate(ctx, req)
	if err != nil {
		return err
	}
	if resp.GetCode() != 0 {
		return fmt.Errorf("leader %s rejected: %s", addr, resp.GetMessage())
	}
	return nil
}

// callLeaderForSyncData calls the SyncData RPC on the leader addr
func (c *clientEndpoint) callLeaderForSyncData(ctx context.Context, addr string, req *proto.SyncDataReq) error {
	e, ok := c.getLeaderEntry(addr)
	if !ok || e == nil || e.client == nil {
		return fmt.Errorf("%w: %s", errNoActiveClient, addr)
	}
	resp, err := e.client.SyncData(ctx, req)
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("sync data rejected by %s: %s", addr, resp.Message)
	}
	return nil
}

// close shuts the endpoint down.
func (c *clientEndpoint) close() {
	c.wg.Wait()

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, e := range c.entries {
		e.close()
	}
	c.entries = make(map[string]*leaderEntry)
}

// backoff sleeps for *delay (bounded by ctx) and doubles it for the next
// attempt. Returns false when ctx is cancelled (caller should exit).
func backoff(ctx context.Context, delay *time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(*delay):
		*delay = nextBackoff(*delay)
	}
}

// nextBackoff doubles the delay, capped at reconnectMaxDelay.
func nextBackoff(d time.Duration) time.Duration {
	d *= reconnectBackoffFactor
	if d > reconnectMaxDelay {
		d = reconnectMaxDelay
	}
	return d
}
