/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"log"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"

	"taskd/toolkit_backend/grpool"
	"taskd/toolkit_backend/net/common"
	"taskd/toolkit_backend/net/proto"
)

const (
	int10 = 10
)

// limiter limits the QPS of the network.
var limiter = rate.NewLimiter(rate.Every(time.Second), common.GrpcQps)

// NetInstance represents the network netIns.
type NetInstance struct {
	config       *common.TaskNetConfig
	upEndpoint   *upStreamEndpoint
	downEndpoint *downStreamEndpoint
	recvBuffer   chan *common.Message
	destroyed    atomic.Bool
	ctx          context.Context
	cancel       context.CancelFunc
	grPool       grpool.GrPool
	rw           sync.RWMutex
}

// InitNetwork initializes the network netIns.
func InitNetwork(conf *common.TaskNetConfig) (*NetInstance, error) {
	err := common.CheckConfig(conf)
	if err != nil {
		return nil, err
	}
	netIns := &NetInstance{config: conf}
	netIns.destroyed.Store(false)
	netIns.recvBuffer = make(chan *common.Message, common.RoleRecvBuffer(conf.Pos.Role))
	netIns.ctx, netIns.cancel = context.WithCancel(context.Background())
	workers := common.RoleWorkerNum(conf.Pos.Role)
	if workers <= 0 {
		return nil, errors.New("worker num must be greater than 0")
	}
	netIns.grPool = grpool.NewPool(uint32(workers), netIns.ctx)
	if common.RoleLevel(conf.Pos.Role) > common.MinRoleLevel {
		log.Println("need start server")
		netIns.downEndpoint, err = newDownStreamEndpoint(netIns)
		if err != nil {
			return nil, err
		}
	}
	if common.RoleLevel(conf.Pos.Role) < common.MaxRoleLevel {
		log.Println("need start client")
		netIns.upEndpoint, err = newUpStreamEndpoint(netIns)
		if err != nil {
			return nil, err
		}
	}
	return netIns, nil
}

// SyncSendMessage sends a message synchronously.
func (nt *NetInstance) SyncSendMessage(uuid, mtype, msgBody string, dst *common.Position) (*common.Ack, error) {
	data := common.DataFrame(uuid, mtype, msgBody, &nt.config.Pos, dst)
	data.Header.Sync = true
	code, err := common.ValidateAndCorrectFrame(data)
	if err != nil {
		return &common.Ack{
			Uuid: data.Header.Uuid,
			Code: uint32(code),
			Src:  &nt.config.Pos,
		}, err
	}
	if common.IsBroadCast(data.Header.Dst) {
		data.Header.Sync = false
	}
	dstType := common.DstCase(&nt.config.Pos, dst)
	protoAck, err := nt.route(data, dstType, common.DataFromLower)
	return common.ExtractAckFrame(protoAck), err
}

// AsyncSendMessage sends a message asynchronously.
func (nt *NetInstance) AsyncSendMessage(uuid, mtype, msgBody string, dst *common.Position) error {
	data := common.DataFrame(uuid, mtype, msgBody, &nt.config.Pos, dst)
	data.Header.Sync = false
	_, err := common.ValidateAndCorrectFrame(data)
	if err != nil {
		return err
	}
	dstType := common.DstCase(&nt.config.Pos, dst)
	_, err = nt.route(data, dstType, common.DataFromLower)
	return err
}

// ReceiveMessage receives a message from the receive buffer.
func (nt *NetInstance) ReceiveMessage() *common.Message {
	select {
	case msg := <-nt.recvBuffer:
		return msg
	case <-nt.ctx.Done():
		return nil
	}
}

// Destroy destroys the network netIns.
func (nt *NetInstance) Destroy() {
	nt.destroyed.Store(true)
	nt.grPool.Close()
	nt.cancel()
	if nt.downEndpoint != nil {
		nt.downEndpoint.close()
	}
	if nt.upEndpoint != nil {
		nt.upEndpoint.close()
	}
}

// send2Buffer sends a message to the receive buffer.
func (nt *NetInstance) send2Buffer(msg *proto.Message) (*proto.Ack, error) {
	select {
	case nt.recvBuffer <- common.ExtractDataFrame(msg):
		return common.AckFrame(msg.Header.Uuid, common.OK, &nt.config.Pos), nil
	case <-time.After(time.Millisecond * int10):
		return common.AckFrame(msg.Header.Uuid, common.RecvBufBusy, &nt.config.Pos),
			errors.New("dst recv buffer busy")
	}
}

// route routes the message based on the destination type.
func (nt *NetInstance) route(msg *proto.Message, dstType string, fromType string) (*proto.Ack, error) {
	switch dstType {
	case common.Dst2Self:
		log.Println("dst is self", msg.Body)
		return nt.send2Buffer(msg)
	case common.Dst2LowerLevel:
		log.Println("dst is lower level", msg.Body)
		return nt.downEndpoint.send(msg)
	case common.Dst2SameLevel, common.Dst2UpperLevel:
		if dstType == common.Dst2SameLevel {
			log.Println("dst is same level", msg.Body)
		} else {
			log.Println("dst is upper level", msg.Body)
		}
		if fromType == common.DataFromUpper {
			log.Println("from is upper is not allowed")
			return common.AckFrame(msg.Header.Uuid, common.NoRoute, &nt.config.Pos),
				errors.New("no route")
		}
		return nt.upEndpoint.send(msg)
	default:
		return common.AckFrame(msg.Header.Uuid, common.ClientErr, &nt.config.Pos),
			errors.New("no route")
	}
}

// proxyPathDiscovery handles the path discovery request.
func (nt *NetInstance) proxyPathDiscovery(ctx context.Context, req *proto.PathDiscoveryReq) (*proto.Ack, error) {
	if common.RoleLevel(nt.config.Pos.Role) == common.MaxRoleLevel {
		return common.AckFrame(req.Uuid, common.OK, &nt.config.Pos), nil
	}
	pos := &proto.Position{
		Role:        nt.config.Pos.Role,
		ServerRank:  nt.config.Pos.ServerRank,
		ProcessRank: nt.config.Pos.ProcessRank,
	}
	req.ProxyPos = pos
	req.Path = append(req.Path, pos)
	if nt.upEndpoint == nil || nt.upEndpoint.upStreamClient == nil {
		return common.AckFrame(req.Uuid, common.OK, &nt.config.Pos), nil
	}
	return nt.upEndpoint.upStreamClient.PathDiscovery(ctx, req)
}
