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
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"taskd/toolkit_backend/net/common"
	"taskd/toolkit_backend/net/proto"
)

const (
	ackTimeout = 10 * time.Second
)

// upStreamEndpoint represents the endpoint for the upstream network.
type upStreamEndpoint struct {
	netInstance    *NetInstance
	upStreamClient proto.TaskNetClient
	stream         proto.TaskNet_InitServerDownStreamClient
	upStreamCoon   *grpc.ClientConn
	destroyed      atomic.Bool
	mu             sync.Mutex
}

// newUpStreamEndpoint creates a new upstream endpoint, joins the task network, and starts listening.
func newUpStreamEndpoint(tool *NetInstance) (*upStreamEndpoint, error) {
	ndp := &upStreamEndpoint{
		netInstance: tool,
	}
	ndp.joinTaskNetwork()
	go ndp.listenUpStreamMessage()
	return ndp, nil
}

// Register registers the endpoint with the upstream server.
func (up *upStreamEndpoint) Register(ctx context.Context, req *proto.RegisterReq) (*proto.Ack, error) {
	return up.upStreamClient.Register(ctx, req)
}

// send sends a message through the upstream client.
func (up *upStreamEndpoint) send(msg *proto.Message) (*proto.Ack, error) {
	return up.upStreamClient.TransferMessage(up.netInstance.ctx, msg)
}

// close destroys the upstream endpoint and resets the network.
func (up *upStreamEndpoint) close() {
	up.destroyed.Store(true)
	up.resetNet()
}

// resetNet closes the stream and the client connection.
func (up *upStreamEndpoint) resetNet() {
	up.netInstance.upClientInited.Store(false)
	if up.stream != nil {
		err := up.stream.CloseSend()
		up.netInstance.netlogger.Errorf("close upstream error: %v, role=%s, srvRank=%s, processRank=%s",
			err, up.netInstance.config.Pos.Role, up.netInstance.config.Pos.ServerRank, up.netInstance.config.Pos.ProcessRank)
	}
	if up.upStreamCoon != nil {
		err := up.upStreamCoon.Close()
		up.netInstance.netlogger.Errorf("close client connection error: %v, role=%s, srvRank=%s, processRank=%s",
			err, up.netInstance.config.Pos.Role, up.netInstance.config.Pos.ServerRank, up.netInstance.config.Pos.ProcessRank)
	}
}

// joinTaskNetwork attempts to join the task network and perform path joining if necessary.
func (up *upStreamEndpoint) joinTaskNetwork() {
	err := up.join()
	if err != nil {
		up.netInstance.netlogger.Errorf("join task network error: %v, role=%s, srvRank=%s, processRank=%s",
			err, up.netInstance.config.Pos.Role, up.netInstance.config.Pos.ServerRank, up.netInstance.config.Pos.ProcessRank)
	}
	for err != nil {
		if up.destroyed.Load() {
			return
		}
		time.Sleep(time.Second)
		err = up.join()
		if err != nil {
			up.netInstance.netlogger.Errorf("join task network error: %v, role=%s, srvRank=%s, processRank=%s",
				err, up.netInstance.config.Pos.Role, up.netInstance.config.Pos.ServerRank, up.netInstance.config.Pos.ProcessRank)
		}
	}
	pos := &proto.Position{
		Role:        up.netInstance.config.Pos.Role,
		ServerRank:  up.netInstance.config.Pos.ServerRank,
		ProcessRank: up.netInstance.config.Pos.ProcessRank,
	}
	if common.RoleLevel(up.netInstance.config.Pos.Role) == common.MinRoleLevel {
		up.pathJoin(pos)
	}
}

// pathJoin attempts to perform path discovery until it succeeds.
func (up *upStreamEndpoint) pathJoin(pos *proto.Position) {
	for !up.destroyed.Load() {
		_, err := up.upStreamClient.PathDiscovery(up.netInstance.ctx, &proto.PathDiscoveryReq{
			Uuid:     uuid.NewString(),
			ProxyPos: pos,
			Path:     []*proto.Position{pos},
		})
		if err == nil {
			up.netInstance.netlogger.Infof("path join success, role=%s, srvRank=%s, processRank=%s",
				up.netInstance.config.Pos.Role, up.netInstance.config.Pos.ServerRank, up.netInstance.config.Pos.ProcessRank)
			break
		}
		up.netInstance.netlogger.Errorf("path join failed, role=%s, srvRank=%s, processRank=%s, err=%v",
			up.netInstance.config.Pos.Role, up.netInstance.config.Pos.ServerRank, up.netInstance.config.Pos.ProcessRank, err)
		time.Sleep(time.Second)
	}
}

// setUpStream sets up the upstream stream.
func (up *upStreamEndpoint) setUpStream() error {
	var err error
	up.mu.Lock()
	defer up.mu.Unlock()
	metaKv := map[string]string{
		common.MetaRoleKey:        up.netInstance.config.Pos.Role,
		common.MetaServerRankKey:  up.netInstance.config.Pos.ServerRank,
		common.MetaProcessRankKey: up.netInstance.config.Pos.ProcessRank,
	}
	ctx := common.SetContextMetaData(up.netInstance.ctx, metaKv)
	up.stream, err = up.upStreamClient.InitServerDownStream(ctx)
	if err == nil {
		up.netInstance.upClientInited.Store(true)
	}
	return err
}

// timeoutAckUpStream sends an acknowledgment with a timeout.
func (up *upStreamEndpoint) timeoutAckUpStream(ack *proto.Ack) error {
	done := make(chan error, 1)

	go func() {
		done <- up.stream.Send(ack)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(ackTimeout):
		up.netInstance.netlogger.Errorf("send ack time out, msgId=%s, role=%s, srvRank=%s, processRank=%s",
			ack.Uuid, up.netInstance.config.Pos.Role, up.netInstance.config.Pos.ServerRank, up.netInstance.config.Pos.ProcessRank)
		return errors.New("send ack time out")
	}
}

// handleUpStreamData handles the data received from the upstream stream.
func (up *upStreamEndpoint) handleUpStreamData(msg *proto.Message) {
	dst := common.Position{
		Role:        msg.Header.Dst.Role,
		ServerRank:  msg.Header.Dst.ServerRank,
		ProcessRank: msg.Header.Dst.ProcessRank,
	}
	var ack *proto.Ack
	dstType := common.DstCase(&up.netInstance.config.Pos, &dst)
	switch dstType {
	case common.Dst2Self, common.Dst2LowerLevel, common.Dst2SameLevel, common.Dst2UpperLevel:
		ack, _ = up.netInstance.route(msg, dstType, common.DataFromUpper)
	default:
		ack = common.AckFrame(msg.Header.Uuid,
			common.DstTypeIllegal, &up.netInstance.config.Pos)
	}
	if msg.Header.Sync && ack != nil {
		if up.timeoutAckUpStream(ack) != nil {
			up.resetNet()
		}
	}
}

// rebuildNet resets the network and attempts to rejoin the task network and set up the stream.
func (up *upStreamEndpoint) rebuildNet() error {
	up.netInstance.netlogger.Infof("rebuild net, role=%s, srvRank=%s, processRank=%s",
		up.netInstance.config.Pos.Role, up.netInstance.config.Pos.ServerRank, up.netInstance.config.Pos.ProcessRank)
	up.resetNet()
	up.joinTaskNetwork()
	return up.setUpStream()
}

// listenUpStreamMessage listens for messages from the upstream stream.
func (up *upStreamEndpoint) listenUpStreamMessage() {
	up.netInstance.netlogger.Info("start listen upstream message")
	var err error
	err = up.setUpStream()
	for err != nil && !up.destroyed.Load() {
		err = up.rebuildNet()
		if err != nil {
			up.netInstance.netlogger.Errorf("rebuild net error: %v, role=%s, srvRank=%s, processRank=%s",
				err, up.netInstance.config.Pos.Role, up.netInstance.config.Pos.ServerRank, up.netInstance.config.Pos.ProcessRank)
		}
	}
	var msg *proto.Message
	for !up.destroyed.Load() {
		msg, err = up.stream.Recv()
		if err != nil {
			err = up.rebuildNet()
			if err != nil {
				up.netInstance.netlogger.Errorf("rebuild net error: %v, role=%s, srvRank=%s, processRank=%s",
					err, up.netInstance.config.Pos.Role, up.netInstance.config.Pos.ServerRank, up.netInstance.config.Pos.ProcessRank)
			}
			continue
		}
		if msg == nil {
			continue
		}
		up.handleUpStreamData(msg)
	}
}

// join attempts to dial the upstream server and register the endpoint.
func (up *upStreamEndpoint) join() error {
	conn, err := grpc.Dial(up.netInstance.config.UpstreamAddr,
		grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		up.netInstance.netlogger.Errorf("join task network error on dial, err=%v, role=%s, srvRank=%s, processRank=%s",
			err, up.netInstance.config.Pos.Role, up.netInstance.config.Pos.ServerRank, up.netInstance.config.Pos.ProcessRank)
		return fmt.Errorf("join task network error on dial, err=%v", err)
	}
	up.mu.Lock()
	up.upStreamCoon = conn
	up.upStreamClient = proto.NewTaskNetClient(conn)
	up.mu.Unlock()
	ack, err := up.upStreamClient.Register(context.Background(),
		common.RegisterReqFrame(&up.netInstance.config.Pos))
	if err != nil || ack == nil || ack.Code != 0 {
		if ack == nil {
			up.netInstance.netlogger.Error("register ack is nil")
			return errors.New("register ack is nil")
		}
		up.netInstance.netlogger.Errorf("join task network error on register, err=%v, code=%d, role=%s, srvRank=%s, processRank=%s",
			err, ack.Code, up.netInstance.config.Pos.Role, up.netInstance.config.Pos.ServerRank, up.netInstance.config.Pos.ProcessRank)
		return fmt.Errorf("join task network error on register, err=%v, code=%d", err, ack.Code)
	}
	return nil
}
