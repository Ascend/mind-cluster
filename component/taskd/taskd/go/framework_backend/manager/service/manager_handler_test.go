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

// Package service is to provide other service tools, i.e. clusterd
package service

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"clusterd/pkg/interface/grpc/recover"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
)

type fakeClient struct {
	pb.RecoverClient
}

func (f *fakeClient) Init(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) Register(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) SubscribeProcessManageSignal(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (pb.Recover_SubscribeProcessManageSignalClient, error) {
	return nil, nil
}

func (f *fakeClient) ReportStopComplete(ctx context.Context, in *pb.StopCompleteRequest, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) ReportRecoverStrategy(ctx context.Context, in *pb.RecoverStrategyRequest, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) ReportRecoverStatus(ctx context.Context, in *pb.RecoverStatusRequest, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) ReportProcessFault(ctx context.Context, in *pb.ProcessFaultRequest, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) SwitchNicTrack(ctx context.Context, in *pb.SwitchNics, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) SubscribeSwitchNicSignal(ctx context.Context, in *pb.SwitchNicRequest, opts ...grpc.CallOption) (pb.Recover_SubscribeSwitchNicSignalClient, error) {
	return nil, nil
}

func (f *fakeClient) SubscribeNotifySwitch(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (pb.Recover_SubscribeNotifySwitchClient, error) {
	return nil, nil
}

func (f *fakeClient) ReplySwitchNicResult(ctx context.Context, in *pb.SwitchResult, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) HealthCheck(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) StressTest(ctx context.Context, in *pb.StressTestParam, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) SubscribeStressTestResponse(ctx context.Context, in *pb.StressTestRequest, opts ...grpc.CallOption) (pb.Recover_SubscribeStressTestResponseClient, error) {
	return nil, nil
}

func (f *fakeClient) SubscribeNotifyExecStressTest(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (pb.Recover_SubscribeNotifyExecStressTestClient, error) {
	return nil, nil
}

func (f *fakeClient) ReplyStressTestResult(ctx context.Context, in *pb.StressTestResult, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func TestManagerHandler(t *testing.T) {
	t.Run("TestManagerHandler, reply to clusterd", func(t *testing.T) {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		dataPool := &storage.DataPool{
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
			},
			RWMutex: sync.RWMutex{},
		}
		msg := storage.BaseMessage{
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.ReplyToClusterDCode,
			},
		}
		processor := &MsgProcessor{}
		called := false
		var mu sync.Mutex
		patch.ApplyPrivateMethod(processor, "replyToClusterD", func(result map[string]string) {
			mu.Lock()
			defer mu.Unlock()
			called = true
			return
		})
		err := processor.managerHandler(dataPool, msg)
		time.Sleep(time.Second)
		assert.Nil(t, err)
		mu.Lock()
		defer mu.Unlock()
		assert.True(t, called)
	})
}

func TestReplyToClusterD(t *testing.T) {
	t.Run("TestReplyToClusterD, GetClusterdAddr failed", func(t *testing.T) {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		processor := &MsgProcessor{}
		called := false
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) {
			called = true
			return "", fmt.Errorf("get clusterd address err")
		})
		processor.replyToClusterD(map[string]string{})
		assert.True(t, called)
	})
}

func TestReplyToClusterDReplyMsg(t *testing.T) {
	t.Run("TestReplyToClusterD, reply to replyStressTestMsg", func(t *testing.T) {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		processor := &MsgProcessor{}
		called := false
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil).
			ApplyPrivateMethod(processor, "replyStressTestMsg", func(result string, client pb.RecoverClient) error {
				called = true
				return nil
			})
		processor.replyToClusterD(map[string]string{constant.StressTestResultStr: "success"})
		assert.True(t, called)
	})
	t.Run("TestReplyToClusterD, reply to replySwitchNicMsg", func(t *testing.T) {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		processor := &MsgProcessor{}
		called := false
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil).
			ApplyPrivateMethod(processor, "replySwitchNicMsg", func(result string, client pb.RecoverClient) error {
				called = true
				return nil
			})
		processor.replyToClusterD(map[string]string{constant.SwitchNicResultStr: "success"})
		assert.True(t, called)
	})
}

func TestReplyMsg(t *testing.T) {
	t.Run("TestReplyStressTestMsg, ok", func(t *testing.T) {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		processor := &MsgProcessor{}
		result := utils.ObjToString(&pb.StressTestResult{})
		err := processor.replyStressTestMsg(result, &fakeClient{})
		assert.Nil(t, err)
	})
	t.Run("TestReplySwitchNicMsg, ok", func(t *testing.T) {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		processor := &MsgProcessor{}
		result := utils.ObjToString(&pb.SwitchResult{})
		err := processor.replySwitchNicMsg(result, &fakeClient{})
		assert.Nil(t, err)
	})
}
