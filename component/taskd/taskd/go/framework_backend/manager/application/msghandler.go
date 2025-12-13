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

// Package application for taskd manager application
package application

import (
	"context"
	"errors"
	"os"
	"strconv"
	"sync"

	"github.com/google/uuid"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	clusterdconstant "clusterd/pkg/common/constant"
	"taskd/common/constant"
	taskdutils "taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/framework_backend/manager/service"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

// MsgHandlerInterface define MsgHandler interface
type MsgHandlerInterface interface {
	GetDataPool() *storage.DataPool
	SendMsgUseGrpc(msgType string, msgBody string, dst *common.Position)
	SendMsgToMgr(uuid string, bizType string, src *common.Position, msgBody storage.MsgBody)
}

// MsgHandler receive, send, process and store message info
type MsgHandler struct {
	Sender    *service.MsgSender
	Receiver  *service.MsgReceiver
	Processor *service.MsgProcessor
	DataPool  *storage.DataPool
	MsgQueue  *storage.MsgQueue
}

// NewMsgHandler new message handler
func NewMsgHandler(workerNum int) *MsgHandler {
	return &MsgHandler{
		Sender: &service.MsgSender{
			RequestChan: make(chan service.SendGrpcMsg, constant.RequestChanNum),
		},
		Receiver:  &service.MsgReceiver{},
		Processor: &service.MsgProcessor{},
		DataPool: &storage.DataPool{
			Snapshot: &storage.SnapShot{
				AgentInfos: &storage.AgentInfos{
					Agents:    map[string]*storage.AgentInfo{},
					AllStatus: map[string]string{},
					RWMutex:   sync.RWMutex{},
				},
				WorkerInfos: &storage.WorkerInfos{
					Workers:   map[string]*storage.WorkerInfo{},
					AllStatus: map[string]string{},
					RWMutex:   sync.RWMutex{},
				},
				ClusterInfos: &storage.ClusterInfos{
					Clusters:  map[string]*storage.ClusterInfo{},
					AllStatus: map[string]string{},
					RWMutex:   sync.RWMutex{},
				},
				MgrInfos: &storage.MgrInfo{
					Status:  map[string]string{},
					RWMutex: sync.RWMutex{},
				},
				WorkerNum: workerNum,
			},
			RWMutex: sync.RWMutex{},
		},
		MsgQueue: &storage.MsgQueue{},
	}
}

// Start the message handler
func (mhd *MsgHandler) Start(ctx context.Context) {
	err, tool := mhd.initManagerGrpc()
	if err != nil {
		hwlog.RunLog.Errorf("init manager grpc err:%v", err)
		return
	}
	go mhd.receiver(tool, ctx)
	go mhd.sender(tool, ctx)
	go mhd.process(ctx)
}

func (mhd *MsgHandler) initManagerGrpc() (error, *net.NetInstance) {
	ip := os.Getenv("POD_IP")
	if ip == "" {
		ip = constant.DefaultIP
	}
	if err := utils.IsHostValid(ip); err != nil {
		return err, nil
	}
	proxyIp := os.Getenv(constant.LocalProxyEnableEnv)
	if proxyIp == constant.LocalProxyEnableOn {
		hwlog.RunLog.Infof("taskd mgr use local proxy")
		ip = constant.LocalProxyIP
	}
	customLogger := hwlog.SetCustomLogger(hwlog.RunLog)
	if customLogger == nil {
		return errors.New("manager SetCustomLogger failed"), nil
	}
	tool, err := net.InitNetwork(&common.TaskNetConfig{
		Pos: common.Position{
			Role:        common.MgrRole,
			ServerRank:  "0",
			ProcessRank: "-1",
		},
		ListenAddr: ip + constant.MgrPort,
	}, customLogger)
	if err != nil {
		return err, nil
	}
	return nil, tool
}

func (mhd *MsgHandler) sender(tool *net.NetInstance, ctx context.Context) {
	for {
		select {
		case req, ok := <-mhd.Sender.RequestChan:
			if ok {
				go func() {
					mhd.Sender.SendMessage(tool, req)
				}()
			}
		case <-ctx.Done():
			hwlog.RunLog.Info("exiting sender success")
			return
		}
	}
}

func (mhd *MsgHandler) process(ctx context.Context) {
	for i := 0; i < constant.RequestChanNum; i++ {
		go mhd.processOne(ctx)
	}
}

func (mhd *MsgHandler) processOne(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := mhd.MsgQueue.Dequeue()
			if err != nil {
				continue
			}
			hwlog.RunLog.Debugf("dequeue msg: %v", msg)
			err = mhd.Processor.MsgProcessor(mhd.DataPool, msg)
			if err != nil {
				hwlog.RunLog.Error(err)
			}
			if msg.Header.Src.Role == common.AgentRole && msg.Body.Code == constant.RestartTimeCode {
				mhd.responseAgentRestartTimes(msg)
			}
		}
	}
}

func (mhd *MsgHandler) responseAgentRestartTimes(msg storage.BaseMessage) {
	mgrInfo, err := mhd.DataPool.GetMgr()
	if mgrInfo == nil {
		hwlog.RunLog.Error("responseAgentRestartTimes: failed to get manager info, mgrInfo is nil")
		return
	}
	if err != nil {
		hwlog.RunLog.Errorf("responseAgentRestartTimes: failed to get manager info, err: %v", err)
		return
	}
	if mhd.shouldStartAgent(msg, mgrInfo) {
		return
	}
	mgrRestartTimes := 0
	if restartTimeStr, exists := mgrInfo.Status[constant.ReportRestartTime]; exists && restartTimeStr != "" {
		var parseErr error
		mgrRestartTimes, parseErr = strconv.Atoi(restartTimeStr)
		if parseErr != nil {
			hwlog.RunLog.Errorf("responseAgentRestartTimes: failed to parse manager restart times '%s', err: %v", restartTimeStr, parseErr)
		}
	} else {
		hwlog.RunLog.Info("responseAgentRestartTimes: manager restart time not found or empty, using default value 0")
	}
	agentRestartTimes := 0
	if msg.Body.Message != "" {
		hwlog.RunLog.Infof("responseAgentRestartTimes: received agent restart times message '%s'", msg.Body.Message)
		var parseErr error
		agentRestartTimes, parseErr = strconv.Atoi(msg.Body.Message)
		if parseErr != nil {
			hwlog.RunLog.Errorf("responseAgentRestartTimes: failed to parse agent restart times '%s', err: %v", msg.Body.Message, parseErr)
		}
	}
	restartTimes := mgrRestartTimes
	if mgrRestartTimes == 0 {
		restartTimes = agentRestartTimes
		hwlog.RunLog.Debugf("responseAgentRestartTimes: using agent restart times %d as manager restart times is 0", agentRestartTimes)
	}
	msgBody := storage.MsgBody{
		MsgType: constant.Action,
		Code:    constant.StartAgentCode,
		Message: strconv.Itoa(restartTimes),
	}
	hwlog.RunLog.Infof("responseAgentRestartTimes: sending response with restart times %d", restartTimes)
	mhd.SendMsgUseGrpc(msg.Header.BizType, taskdutils.ObjToString(msgBody), msg.Header.Src)
}

func (mhd *MsgHandler) shouldStartAgent(msg storage.BaseMessage, mgrInfo *storage.MgrInfo) bool {
	if mgrInfo.Status[constant.SignalType] == clusterdconstant.WaitStartAgentSignalType {
		hwlog.RunLog.Infof("recv: wait start agent signal: %v", mgrInfo.Status[constant.SignalType])
		// wait start agent signal, enqueue msg
		err := mhd.MsgQueue.Enqueue(msg)
		if err != nil {
			hwlog.RunLog.Errorf("enqueue msg failed: %v", err)
		}
		return true
	}
	if mgrInfo.Status[constant.SignalType] == clusterdconstant.ContinueStartAgentSignalType {
		hwlog.RunLog.Infof("recv: continue start agent signal: %v", mgrInfo.Status[constant.SignalType])
	}
	return false
}

func (mhd *MsgHandler) receiver(tool *net.NetInstance, ctx context.Context) {
	for i := 0; i < constant.RequestChanNum; i++ {
		go mhd.receiveGoroutine(tool, ctx)
	}
}

func (mhd *MsgHandler) receiveGoroutine(tool *net.NetInstance, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Debug("Mgr ReceiveMsg exit")
			break
		default:
			msg, msgBody, err := mhd.Receiver.ReceiveMsg(mhd.MsgQueue, tool)
			if err != nil {
				hwlog.RunLog.Errorf("receive msg enqueue failed: %v", err)
				mhd.SendMsgUseGrpc(msg.BizType, taskdutils.ObjToString(msgBody), msg.Src)
			}
		}
	}
}

// SendMsgUseGrpc send message use grpc server
func (mhd *MsgHandler) SendMsgUseGrpc(msgType string, msgBody string, dst *common.Position) {
	mhd.Sender.RequestChan <- service.SendGrpcMsg{
		Uuid:    uuid.New().String(),
		MsgType: msgType,
		MsgBody: msgBody,
		Dst:     dst,
	}
}

// SendMsgToMgr send message into manager message queue
func (mhd *MsgHandler) SendMsgToMgr(uuid string, bizType string, src *common.Position, msgBody storage.MsgBody) {
	data := mhd.MsgQueue.NewMsg(uuid, bizType, src, msgBody)
	err := mhd.MsgQueue.Enqueue(data)
	if err != nil {
		hwlog.RunLog.Errorf("enqueue failed: %v", err)
		mhd.SendMsgUseGrpc(bizType, err.Error(), src)
	}
}

// GetDataPool return data pool
func (mhd *MsgHandler) GetDataPool() *storage.DataPool {
	return mhd.DataPool
}
