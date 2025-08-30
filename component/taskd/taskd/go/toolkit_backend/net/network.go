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
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"

	"ascend-common/common-utils/hwlog"
	"taskd/toolkit_backend/grpool"
	"taskd/toolkit_backend/net/common"
	"taskd/toolkit_backend/net/proto"
)

const (
	int10 = 10
)

// limiter limits the QPS of the network.
var limiter = rate.NewLimiter(rate.Limit(common.GrpcQps), common.GrpcQps)

// NetInstance represents the network netIns.
type NetInstance struct {
	config         *common.TaskNetConfig
	upEndpoint     *upStreamEndpoint
	upClientInited atomic.Bool
	downEndpoint   *downStreamEndpoint
	recvBuffer     chan *common.Message
	destroyed      atomic.Bool
	ctx            context.Context
	cancel         context.CancelFunc
	grPool         grpool.GrPool
	rw             sync.RWMutex
	netlogger      *hwlog.CustomLogger
}

// InitNetwork initializes the network netIns.
func InitNetwork(conf *common.TaskNetConfig, logger *hwlog.CustomLogger) (*NetInstance, error) {
	err := common.CheckConfig(conf)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		return nil, errors.New("logger is nil")
	}
	netIns := &NetInstance{config: conf, netlogger: logger}
	netIns.destroyed.Store(false)
	netIns.recvBuffer = make(chan *common.Message, common.RoleRecvBuffer(conf.Pos.Role))
	netIns.ctx, netIns.cancel = context.WithCancel(context.Background())
	workers := common.RoleWorkerNum(conf.Pos.Role)
	if workers <= 0 {
		netIns.netlogger.Errorf("worker num must be greater than 0, but got %d", workers)
		return nil, errors.New("worker num must be greater than 0")
	}
	netIns.grPool = grpool.NewPool(uint32(workers), netIns.ctx)
	if common.RoleLevel(conf.Pos.Role) > common.MinRoleLevel {
		netIns.netlogger.Infof("need start server, role=%s, srvRank=%s, processRank=%s",
			conf.Pos.Role, conf.Pos.ServerRank, conf.Pos.ProcessRank)
		netIns.downEndpoint, err = newDownStreamEndpoint(netIns)
		if err != nil {
			return nil, err
		}
	}
	if common.RoleLevel(conf.Pos.Role) < common.MaxRoleLevel {
		netIns.netlogger.Infof("need start client, role=%s, srvRank=%s, processRank=%s",
			conf.Pos.Role, conf.Pos.ServerRank, conf.Pos.ProcessRank)
		netIns.upEndpoint, err = newUpStreamEndpoint(netIns)
		if err != nil {
			netIns.netlogger.Errorf("newUpStreamEndpoint failed, err=%v", err)
			return nil, err
		}
	}
	return netIns, nil
}

// SyncSendMessage sends a message synchronously.
func (nt *NetInstance) SyncSendMessage(uuid, mtype, msgBody string, dst *common.Position) (*common.Ack, error) {
	data := common.DataFrame(uuid, mtype, msgBody, &nt.config.Pos, dst)
	if data == nil {
		return &common.Ack{
			Uuid: uuid,
			Code: common.ClientErr,
			Src:  &nt.config.Pos,
		}, errors.New("nil data")
	}
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
	nt.netlogger.Debugf("SyncSendMessage error, uuid=%s, mtype=%s, msgBody=%s, dst=%v, dstType=%s, protoAck=%v, err=%v",
		uuid, mtype, msgBody, dst, dstType, protoAck, err)
	return common.ExtractAckFrame(protoAck), err
}

// AsyncSendMessage sends a message asynchronously.
func (nt *NetInstance) AsyncSendMessage(uuid, mtype, msgBody string, dst *common.Position) error {
	data := common.DataFrame(uuid, mtype, msgBody, &nt.config.Pos, dst)
	if data == nil {
		return errors.New("nil data")
	}
	data.Header.Sync = false
	_, err := common.ValidateAndCorrectFrame(data)
	if err != nil {
		return err
	}
	dstType := common.DstCase(&nt.config.Pos, dst)
	_, err = nt.route(data, dstType, common.DataFromLower)
	nt.netlogger.Debugf("AsyncSendMessage error, uuid=%s, mtype=%s, msgBody=%s, dst=%v, dstType=%s, err=%v",
		uuid, mtype, msgBody, dst, dstType, err)
	return err
}

// ReceiveMessage receives a message from the receive buffer.
func (nt *NetInstance) ReceiveMessage() *common.Message {
	select {
	case msg := <-nt.recvBuffer:
		nt.netlogger.Debugf("receive a message, msg=%v", msg)
		return msg
	case <-nt.ctx.Done():
		nt.netlogger.Infof("ReceiveMessage, ctx done, role=%s, srvRank=%s, processRank=%s",
			nt.config.Pos.Role, nt.config.Pos.ServerRank, nt.config.Pos.ProcessRank)
		return nil
	}
}

// Destroy destroys the network netIns.
func (nt *NetInstance) Destroy() {
	nt.netlogger.Infof("taskNet Destroy, role=%s, srvRank=%s, processRank=%s",
		nt.config.Pos.Role, nt.config.Pos.ServerRank, nt.config.Pos.ProcessRank)
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

// GetNetworkerLogger returns the networker logger.
func (nt *NetInstance) GetNetworkerLogger() *hwlog.CustomLogger {
	return nt.netlogger
}

// send2Buffer sends a message to the receive buffer.
func (nt *NetInstance) send2Buffer(msg *proto.Message) (*proto.Ack, error) {
	select {
	case nt.recvBuffer <- common.ExtractDataFrame(msg):
		return common.AckFrame(msg.Header.Uuid, common.OK, &nt.config.Pos), nil
	case <-time.After(time.Millisecond * int10):
		nt.netlogger.Errorf("send2Buffer failed, dst recv buffer busy, msgid=%s, role=%s, srvRank=%s, processRank=%s",
			msg.Header.Uuid, nt.config.Pos.Role, nt.config.Pos.ServerRank, nt.config.Pos.ProcessRank)
		return common.AckFrame(msg.Header.Uuid, common.RecvBufBusy, &nt.config.Pos),
			errors.New("dst recv buffer busy")
	}
}

// route routes the message based on the destination type.
func (nt *NetInstance) route(msg *proto.Message, dstType string, fromType string) (*proto.Ack, error) {
	switch dstType {
	case common.Dst2Self:
		return nt.send2Buffer(msg)
	case common.Dst2LowerLevel:
		return nt.downEndpoint.send(msg)
	case common.Dst2SameLevel, common.Dst2UpperLevel:
		if fromType == common.DataFromUpper {
			nt.netlogger.Errorf("from is upper is not allowed, msgid=%s, role=%s, srvRank=%s, processRank=%s",
				msg.Header.Uuid, nt.config.Pos.Role, nt.config.Pos.ServerRank, nt.config.Pos.ProcessRank)
			return common.AckFrame(msg.Header.Uuid, common.NoRoute, &nt.config.Pos),
				errors.New("no route")
		}
		if nt.upEndpoint == nil || nt.upClientInited.Load() == false {
			nt.netlogger.Errorf("client not inited, role=%s, srvRank=%s, processRank=%s",
				nt.config.Pos.Role, nt.config.Pos.ServerRank, nt.config.Pos.ProcessRank)
			return common.AckFrame(msg.Header.Uuid, common.ServerErr, &nt.config.Pos),
				fmt.Errorf("client not inited, role=%s, srvRank=%s, processRank=%s",
					nt.config.Pos.Role, nt.config.Pos.ServerRank, nt.config.Pos.ProcessRank)
		}
		return nt.upEndpoint.send(msg)
	default:
		nt.netlogger.Errorf("dst type illegal, msgid=%s, role=%s, srvRank=%s, processRank=%s",
			msg.Header.Uuid, nt.config.Pos.Role, nt.config.Pos.ServerRank, nt.config.Pos.ProcessRank)
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
	if nt.upEndpoint == nil || nt.upEndpoint.upStreamClient == nil || nt.upClientInited.Load() == false {
		nt.netlogger.Errorf("client not inited, role=%s, srvRank=%s, processRank=%s",
			nt.config.Pos.Role, nt.config.Pos.ServerRank, nt.config.Pos.ProcessRank)
		return common.AckFrame(req.Uuid, common.ClientErr, &nt.config.Pos),
			fmt.Errorf("client not inited, role=%s, srvRank=%s, processRank=%s",
				nt.config.Pos.Role, nt.config.Pos.ServerRank, nt.config.Pos.ProcessRank)
	}
	return nt.upEndpoint.upStreamClient.PathDiscovery(ctx, req)
}
