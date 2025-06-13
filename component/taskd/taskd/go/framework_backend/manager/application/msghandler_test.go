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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/framework_backend/manager/service"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

const fiveHundred = 500

// TestMain test main
func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	return initLog()
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return err
	}
	return nil
}

func capturePanic(f func()) error {
	var err error
	defer func() {
		err = nil
		if recovered := recover(); recovered != nil {
			err = errors.New("panic error")
		}
	}()
	f()
	return err
}

// TestNewMsgHandler test new message handler
func TestNewMsgHandler(t *testing.T) {
	convey.Convey("TestNewMsgHandler new msg handler return struct", t, func() {
		mhd := NewMsgHandler()
		convey.So(mhd, convey.ShouldResemble, &MsgHandler{
			Sender: &service.MsgSender{
				RequestChan: mhd.Sender.RequestChan,
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
				},
				RWMutex: sync.RWMutex{},
			},
			MsgQueue: &storage.MsgQueue{},
		})
	})
}

// TestStartAndInit test msg handler start and grpc init
func TestStartAndInit(t *testing.T) {
	mhd := NewMsgHandler()
	convey.Convey("Test Start and init manager grpc success", t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), fiveHundred*time.Millisecond)
		defer cancel()
		convey.So(capturePanic(func() { mhd.Start(ctx) }), convey.ShouldBeNil)
	})
	convey.Convey("Test Start and init manager grpc error", t, func() {
		patch := gomonkey.ApplyFuncReturn(net.InitNetwork, &net.NetInstance{}, fmt.Errorf("test error"))
		defer patch.Reset()
		ctx, cancel := context.WithTimeout(context.Background(), fiveHundred*time.Millisecond)
		defer cancel()
		convey.So(capturePanic(func() { mhd.Start(ctx) }), convey.ShouldBeNil)
	})
}

// TestSendMsgUseGrpc test send grpc msg
func TestSendMsgUseGrpc(t *testing.T) {
	mhd := NewMsgHandler()
	testDst := &common.Position{Role: common.WorkerRole}
	convey.Convey("TestSendMsgUseGrpc send grpc msg success", t, func() {
		mhd.SendMsgUseGrpc("test-type", "test-body", testDst)
		req := <-mhd.Sender.RequestChan
		convey.So(req.Uuid, convey.ShouldNotBeNil)
		convey.So(req.MsgType, convey.ShouldEqual, "test-type")
		convey.So(req.MsgBody, convey.ShouldEqual, "test-body")
		convey.So(req.Dst, convey.ShouldEqual, testDst)
	})
}

// TestSendMsgToMgr test manager send msg enqueue
func TestSendMsgToMgr(t *testing.T) {
	convey.Convey("TestSendMsgToMgr manager send msg enqueue success", t, func() {
		mhd := NewMsgHandler()
		testSrc := &common.Position{Role: common.WorkerRole}
		oldLength := len(mhd.MsgQueue.Queue)
		mhd.SendMsgToMgr("test-uuid", "test-type", testSrc, storage.MsgBody{})
		convey.So(oldLength+1, convey.ShouldEqual, len(mhd.MsgQueue.Queue))
	})
	convey.Convey("TestSendMsgToMgr manager send msg enqueue fail", t, func() {
		mhd := &MsgHandler{
			Sender: &service.MsgSender{RequestChan: make(chan service.SendGrpcMsg, constant.RequestChanNum)},
			MsgQueue: &storage.MsgQueue{Queue: make([]storage.BaseMessage, constant.MaxMsgQueueLength),
				Mutex: sync.Mutex{}},
		}
		testSrc := &common.Position{Role: common.WorkerRole}
		mhd.SendMsgToMgr("test-uuid", "test-type", testSrc, storage.MsgBody{})
		convey.So(len(mhd.MsgQueue.Queue), convey.ShouldEqual, constant.MaxMsgQueueLength)
	})
}
