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

// Package manager is to provide other service tools, i.e. clusterd
package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"k8s.io/apimachinery/pkg/util/uuid"

	clusterd_constant "clusterd/pkg/common/constant"
	"clusterd/pkg/interface/grpc/profiling"
	"clusterd/pkg/interface/grpc/recover"
	"taskd/common/constant"
	_ "taskd/common/testtool"
	"taskd/common/utils"
	"taskd/framework_backend/manager/application"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

type fakeClient struct {
	pb.RecoverClient
}

func (f *fakeClient) Init(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (*pb.Status, error) {
	return &pb.Status{Code: common.OK}, nil
}

func (f *fakeClient) Register(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (*pb.Status, error) {
	return &pb.Status{Code: common.OK}, nil
}

func (f *fakeClient) SubscribeProcessManageSignal(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (pb.Recover_SubscribeProcessManageSignalClient, error) {
	return &fakeStream{}, nil
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

type fakeProfilingClient struct {
	profiling.TrainingDataTraceClient
}

func (f *fakeProfilingClient) SubscribeDataTraceSwitch(ctx context.Context, in *profiling.ProfilingClientInfo, opts ...grpc.CallOption) (profiling.TrainingDataTrace_SubscribeDataTraceSwitchClient, error) {
	return nil, nil
}

type fakeStream struct {
	grpc.ClientStream
}

func (s *fakeStream) Recv() (*pb.ProcessManageSignal, error) {
	return &pb.ProcessManageSignal{
		Uuid: string(uuid.NewUUID()),
	}, nil
}

func (s *fakeStream) Context() context.Context {
	return context.Background()
}

func TestReportControllerInfoToClusterd(t *testing.T) {
	convey.Convey("get clusterd addr failed", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) {
			return "", fmt.Errorf("get clusterd address err")
		})
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("init clusterd connect err", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, fmt.Errorf("grpc.Dial err"))

		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("send message to clusterd failed", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil).
			ApplyFuncReturn(reportRecoverStatus, false)
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{Action: constant.RecoverStatus})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("action is unknown action", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil)
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{Action: "action"})
		convey.So(res, convey.ShouldEqual, false)
	})
}

func TestReportControllerInfoToClusterd2(t *testing.T) {
	convey.Convey("message.Action is RecoverStatus", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil).
			ApplyFuncReturn(reportRecoverStatus, true)
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{Action: constant.RecoverStatus})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("message.Action is ProcessFault", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil).
			ApplyFuncReturn(reportProcessFault, true)
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{Action: constant.ProcessFault})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("message.Action is RecoverStrategy", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil).
			ApplyFuncReturn(reportRecoverStrategy, true)
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{Action: constant.RecoverStrategy})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("message.Action is StopComplete", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil).
			ApplyFuncReturn(reportStopComplete, true)
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{Action: constant.StopComplete})
		convey.So(res, convey.ShouldEqual, true)
	})
}

func TestReportRecoverStatus(t *testing.T) {
	message := &constant.ControllerMessage{
		Code: 0,
		Msg:  "msg",
		FaultRanks: map[int]int{
			0: 0,
		},
	}
	convey.Convey("managerInstance is nil", t, func() {
		preInst := managerInstance
		managerInstance = nil
		defer func() {
			managerInstance = preInst
		}()
		res := reportRecoverStatus(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("report recover status to clusterd ok", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		res := reportRecoverStatus(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("report recover status to clusterd failed", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyMethodReturn(&fakeClient{}, "ReportRecoverStatus", &pb.Status{}, fmt.Errorf("err"))
		res := reportRecoverStatus(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
}

func TestReportProcessFault(t *testing.T) {
	message := &constant.ControllerMessage{
		Code: 0,
		Msg:  "msg",
		FaultRanks: map[int]int{
			0: 0,
		},
	}
	convey.Convey("managerInstance is nil", t, func() {
		preInst := managerInstance
		managerInstance = nil
		defer func() {
			managerInstance = preInst
		}()
		res := reportProcessFault(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("report process fault to clusterd ok", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		res := reportProcessFault(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("report process fault to clusterd failed", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyMethodReturn(&fakeClient{}, "ReportProcessFault", &pb.Status{}, fmt.Errorf("err"))
		res := reportProcessFault(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
}

func TestReportRecoverStrategy(t *testing.T) {
	message := &constant.ControllerMessage{
		Code: 0,
		Msg:  "msg",
		FaultRanks: map[int]int{
			0: 0,
		},
	}
	convey.Convey("managerInstance is nil", t, func() {
		preInst := managerInstance
		managerInstance = nil
		defer func() {
			managerInstance = preInst
		}()
		res := reportRecoverStrategy(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("report  strategy to clusterd ok", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		res := reportRecoverStrategy(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("report strategy to clusterd failed", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyMethodReturn(&fakeClient{}, "ReportRecoverStrategy", &pb.Status{}, fmt.Errorf("err"))
		res := reportRecoverStrategy(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
}

func TestReportStopComplete(t *testing.T) {
	message := &constant.ControllerMessage{
		Code: 0,
		Msg:  "msg",
		FaultRanks: map[int]int{
			0: 0,
		},
	}
	convey.Convey("managerInstance is nil", t, func() {
		preInst := managerInstance
		managerInstance = nil
		defer func() {
			managerInstance = preInst
		}()
		res := reportStopComplete(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("report stop complete to clusterd ok", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		res := reportStopComplete(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("report stop complete to clusterd failed", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyMethodReturn(&fakeClient{}, "ReportStopComplete", &pb.Status{}, fmt.Errorf("err"))
		res := reportStopComplete(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
}

func TestBaseManager_Init_Success(t *testing.T) {
	convey.Convey("Test BaseManager Init Success", t, func() {
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()

		config := Config{
			JobId:        "test-job-id",
			NodeNums:     2,
			ProcPerNode:  4,
			PluginDir:    "/test/plugin/dir",
			FaultRecover: "test-recover",
			TaskDEnable:  "on",
		}

		manager := &BaseManager{Config: config}
		err := manager.Init()
		convey.So(err, convey.ShouldBeNil)
		convey.So(manager.svcCtx, convey.ShouldNotBeNil)
		convey.So(manager.cancelFunc, convey.ShouldNotBeNil)
		convey.So(manager.MsgHd, convey.ShouldNotBeNil)
		convey.So(manager.BusinessHandler, convey.ShouldNotBeNil)

		manager.cancelFunc()
	})
}

func TestBaseManager_Init_LoggerError(t *testing.T) {
	convey.Convey("Test BaseManager Init Logger Error", t, func() {
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, context.DeadlineExceeded)
		defer patch.Reset()

		config := Config{
			JobId:       "test-job-id",
			NodeNums:    2,
			ProcPerNode: 4,
		}

		manager := &BaseManager{Config: config}
		err := manager.Init()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldEqual, context.DeadlineExceeded)
	})
}

func TestBaseManager_Init_BusinessHandlerError(t *testing.T) {
	convey.Convey("Test BaseManager Init BusinessHandler Error", t, func() {
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()
		config := Config{
			JobId:       "test-job-id",
			NodeNums:    2,
			ProcPerNode: 4,
		}
		const workerN = 8
		manager := &BaseManager{Config: config}
		manager.MsgHd = application.NewMsgHandler(workerN)
		manager.BusinessHandler = application.NewBusinessStreamProcessor(manager.MsgHd)
		patch2 := gomonkey.ApplyMethodReturn(manager.BusinessHandler, "Init", context.Canceled)
		defer patch2.Reset()
		err := manager.Init()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldEqual, context.Canceled)
	})
}

func TestBaseManager_Init_GoroutinesStarted(t *testing.T) {
	convey.Convey("Test BaseManager Init Goroutines Started", t, func() {
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()
		const timeout = 200 * time.Millisecond
		const sleepTime = 100 * time.Millisecond
		config := Config{
			JobId:       "test-job-id",
			NodeNums:    2,
			ProcPerNode: 4,
		}
		manager := &BaseManager{Config: config}
		err := manager.Init()
		convey.So(err, convey.ShouldBeNil)
		time.Sleep(sleepTime)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		select {
		case <-ctx.Done():
			convey.So(true, convey.ShouldBeTrue)
		case <-manager.svcCtx.Done():
			convey.So(false, convey.ShouldBeTrue)
		}
		manager.cancelFunc()
	})
}

func TestBaseManager_Init_ZeroNodes(t *testing.T) {
	convey.Convey("Test BaseManager Init Zero Nodes", t, func() {
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()
		config := Config{
			JobId:       "test-job-id",
			NodeNums:    0,
			ProcPerNode: 0,
		}
		manager := &BaseManager{Config: config}
		err := manager.Init()
		convey.So(err, convey.ShouldBeNil)
		convey.So(manager.MsgHd, convey.ShouldNotBeNil)
		convey.So(manager.BusinessHandler, convey.ShouldNotBeNil)
		manager.cancelFunc()
	})
}

func TestBaseManager_Init_LargeNodes(t *testing.T) {
	convey.Convey("Test BaseManager Init Large Nodes", t, func() {
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()
		config := Config{
			JobId:       "test-job-id",
			NodeNums:    100,
			ProcPerNode: 8,
		}
		manager := &BaseManager{Config: config}
		err := manager.Init()
		convey.So(err, convey.ShouldBeNil)
		convey.So(manager.MsgHd, convey.ShouldNotBeNil)
		convey.So(manager.BusinessHandler, convey.ShouldNotBeNil)
		manager.cancelFunc()
	})
}

func TestBaseManager_Start_InitError(t *testing.T) {
	convey.Convey("Test BaseManager Start Init Error", t, func() {
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, context.DeadlineExceeded)
		defer patch.Reset()
		config := Config{JobId: "test-job-id", NodeNums: 2, ProcPerNode: 4}
		manager := &BaseManager{Config: config}
		err := manager.Start()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "manager init failed")
	})
}

func TestBaseManager_Start_ProcessError(t *testing.T) {
	convey.Convey("Test BaseManager Start Process Error", t, func() {
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()
		config := Config{JobId: "test-job-id", NodeNums: 2, ProcPerNode: 4}
		manager := &BaseManager{Config: config}
		patch1 := gomonkey.ApplyMethodReturn(manager, "Init", nil)
		defer patch1.Reset()
		patch2 := gomonkey.ApplyMethodReturn(manager, "Process", fmt.Errorf("test err"))
		defer patch2.Reset()
		err := manager.Start()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "manager process failed")
	})
}

func TestBaseManager_Start_Success(t *testing.T) {
	convey.Convey("Test BaseManager Start Success", t, func() {
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()
		config := Config{JobId: "test-job-id", NodeNums: 2, ProcPerNode: 4}
		manager := &BaseManager{Config: config}
		patch1 := gomonkey.ApplyMethodReturn(manager, "Init", nil)
		defer patch1.Reset()
		patch2 := gomonkey.ApplyMethodReturn(manager, "Process", nil)
		defer patch2.Reset()
		err := manager.Start()
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestBaseManager_Process_GetSnapShotError(t *testing.T) {
	convey.Convey("Test BaseManager Process GetSnapShot Error", t, func() {
		const workerN = 8
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()
		config := Config{JobId: "test-job-id", NodeNums: 2, ProcPerNode: 4}
		manager := &BaseManager{Config: config}
		msgHandler := application.NewMsgHandler(workerN)
		manager.MsgHd = msgHandler
		patch1 := gomonkey.ApplyMethodReturn(manager.MsgHd.DataPool, "GetSnapShot", nil, context.Canceled)
		defer patch1.Reset()
		patch2 := gomonkey.ApplyFuncReturn(getProcessInterval, int64(1))
		defer patch2.Reset()
		err := manager.Process()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "get datapool snapshot failed")
	})
}

func TestBaseManager_Process_ServiceError(t *testing.T) {
	convey.Convey("Test BaseManager Process Service Error", t, func() {
		const workerN = 8
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()
		config := Config{JobId: "test-job-id", NodeNums: 2, ProcPerNode: 4}
		manager := &BaseManager{Config: config}
		msgHandler := application.NewMsgHandler(workerN)
		manager.MsgHd = msgHandler
		patch1 := gomonkey.ApplyMethodReturn(manager.MsgHd.DataPool, "GetSnapShot", &storage.SnapShot{}, nil)
		defer patch1.Reset()
		patch2 := gomonkey.ApplyMethodReturn(manager, "Service", context.Canceled)
		defer patch2.Reset()
		err := manager.Process()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "service execute failed")
	})
}

func TestBaseManager_Process_Normal(t *testing.T) {
	convey.Convey("Test BaseManager Process Normal", t, func() {
		const timeout = 200 * time.Millisecond
		const workerN = 8
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()
		config := Config{JobId: "test-job-id", NodeNums: 2, ProcPerNode: 4}
		manager := &BaseManager{Config: config}
		msgHandler := application.NewMsgHandler(workerN)
		manager.MsgHd = msgHandler
		manager.svcCtx, manager.cancelFunc = context.WithCancel(context.Background())
		manager.BusinessHandler = application.NewBusinessStreamProcessor(manager.MsgHd)

		patch1 := gomonkey.ApplyMethodReturn(manager.MsgHd.DataPool, "GetSnapShot", &storage.SnapShot{}, nil)
		defer patch1.Reset()
		patch2 := gomonkey.ApplyMethodReturn(manager.BusinessHandler, "AllocateToken")
		defer patch2.Reset()
		patch3 := gomonkey.ApplyMethodReturn(manager.BusinessHandler, "StreamRun", nil)
		defer patch3.Reset()

		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
		defer timeoutCancel()
		done := make(chan error, 1)
		go func() {
			done <- manager.Process()
		}()
		select {
		case <-timeoutCtx.Done():
			manager.cancelFunc()
			convey.So(true, convey.ShouldBeTrue)
		case err := <-done:
			convey.So(err, convey.ShouldBeNil)
		}
	})
}

func TestBaseManager_Service_Success(t *testing.T) {
	convey.Convey("Test BaseManager Service Success", t, func() {
		const workerN = 8
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()
		config := Config{JobId: "test-job-id", NodeNums: 2, ProcPerNode: 4}
		manager := &BaseManager{Config: config}
		msgHandler := application.NewMsgHandler(workerN)
		manager.MsgHd = msgHandler
		manager.BusinessHandler = application.NewBusinessStreamProcessor(manager.MsgHd)

		patch1 := gomonkey.ApplyMethodReturn(manager.BusinessHandler, "AllocateToken")
		defer patch1.Reset()
		patch2 := gomonkey.ApplyMethodReturn(manager.BusinessHandler, "StreamRun", nil)
		defer patch2.Reset()

		err := manager.Service(&storage.SnapShot{})
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestBaseManager_Service_StreamRunError(t *testing.T) {
	convey.Convey("Test BaseManager Service StreamRun Error", t, func() {
		const workerN = 8
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()
		config := Config{JobId: "test-job-id", NodeNums: 2, ProcPerNode: 4}
		manager := &BaseManager{Config: config}
		msgHandler := application.NewMsgHandler(workerN)
		manager.MsgHd = msgHandler
		manager.BusinessHandler = application.NewBusinessStreamProcessor(manager.MsgHd)

		patch1 := gomonkey.ApplyMethodReturn(manager.BusinessHandler, "AllocateToken")
		defer patch1.Reset()
		patch2 := gomonkey.ApplyMethodReturn(manager.BusinessHandler, "StreamRun", context.Canceled)
		defer patch2.Reset()

		err := manager.Service(&storage.SnapShot{})
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "context canceled")
	})
}

func TestBaseManager_registerClusterD_MaxRetry(t *testing.T) {
	convey.Convey("Test registerClusterD max retry", t, func() {
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()
		manager := &BaseManager{}
		manager.registerClusterD(maxRegRetryTime)
		convey.So(true, convey.ShouldBeTrue)
	})
}

func TestBaseManager_registerClusterD_GetAddrError(t *testing.T) {
	convey.Convey("Test registerClusterD get addr error", t, func() {
		patch := gomonkey.ApplyFuncReturn(utils.InitHwLogger, nil)
		defer patch.Reset()
		patch1 := gomonkey.ApplyFuncReturn(utils.GetClusterdAddr, "", fmt.Errorf("address error"))
		defer patch1.Reset()
		manager := &BaseManager{}
		manager.registerClusterD(0)
		convey.So(true, convey.ShouldBeTrue)
	})
}

func fakeTaskDManager(ctx context.Context) *BaseManager {
	defaultWorkerNum := 8
	m := &BaseManager{
		Config: Config{},
		svcCtx: ctx,
	}
	m.MsgHd = application.NewMsgHandler(defaultWorkerNum)
	m.BusinessHandler = application.NewBusinessStreamProcessor(m.MsgHd)
	return m
}

func TestEnqueueProfilingSwitch(t *testing.T) {
	m := fakeTaskDManager(context.Background())
	convey.Convey("test enqueueProfilingSwitch", t, func() {
		cmd := constant.ProfilingDomainCmd{
			DefaultDomainAble: false,
			CommDomainAble:    false,
		}
		m.enqueueProfilingSwitch(cmd, "0")
		convey.So(len(m.MsgHd.MsgQueue.Queue), convey.ShouldEqual, 1)
		msg, err := m.MsgHd.MsgQueue.Dequeue()
		convey.So(err, convey.ShouldBeNil)
		convey.So(msg.Header.Src.Role, convey.ShouldEqual, constant.ClusterRole)
	})
}

func TestWatchProfilingCmdChange(t *testing.T) {
	convey.Convey("test watchProfilingCmdChange", t, func() {
		convey.Convey("context done, should not get profiling from file", func() {
			ctx, cancel := context.WithCancel(context.Background())
			m := fakeTaskDManager(ctx)
			const sleepTime = 100 * time.Millisecond
			go func() {
				time.Sleep(sleepTime)
				defer cancel()
			}()
			m.watchProfilingCmdChange()
			convey.So(len(m.MsgHd.MsgQueue.Queue), convey.ShouldEqual, 0)
		})
		convey.Convey("profiling from clusterD is true, should not get profiling from file", func() {
			m := fakeTaskDManager(context.Background())
			m.profilingFromClusterD.Store(true)
			m.watchProfilingCmdChange()
			convey.So(len(m.MsgHd.MsgQueue.Queue), convey.ShouldEqual, 0)
		})
		convey.Convey("profiling from clusterD is false, should get profiling from file", func() {
			ctx, cancel := context.WithCancel(context.Background())
			m := fakeTaskDManager(ctx)
			m.profilingFromClusterD.Store(false)
			patch := gomonkey.ApplyFuncReturn(utils.GetProfilingSwitch, constant.ProfilingSwitch{}, nil)
			defer patch.Reset()
			const sleepTime = 1100 * time.Millisecond
			go func() {
				time.Sleep(sleepTime)
				cancel()
			}()
			m.watchProfilingCmdChange()
			convey.So(len(m.MsgHd.MsgQueue.Queue), convey.ShouldNotEqual, 0)
		})
	})
}

func TestConvertProfilingMsg(t *testing.T) {
	convey.Convey("test convertProfilingMsg", t, func() {
		switchData := &profiling.ProfilingSwitch{
			CommunicationOperator: "test",
		}
		ret := convertProfilingMsg(switchData)
		convey.So(ret.CommunicationOperator, convey.ShouldEqual, "test")
	})
}

func TestSubscribeProcessManageSignal(t *testing.T) {
	convey.Convey("test subscribeProcessManageSignal", t, func() {
		convey.Convey("end subscribe", func() {
			m := fakeTaskDManager(context.Background())
			patch := gomonkey.ApplyPrivateMethod(m, "startSubscribe",
				func(m, _ pb.RecoverClient, _ *pb.ClientInfo) bool { return true })
			defer patch.Reset()
			m.subscribeProcessManageSignal(&grpc.ClientConn{})
			convey.So(len(m.MsgHd.MsgQueue.Queue), convey.ShouldEqual, 1)
			msg, err := m.MsgHd.MsgQueue.Dequeue()
			convey.So(err, convey.ShouldBeNil)
			convey.So(msg.Header.Src.Role, convey.ShouldEqual, common.MgrRole)
			convey.So(msg.Body.Code, convey.ShouldEqual, constant.FaultRecoverCode)
		})
	})
}

func TestStartSubscribe(t *testing.T) {
	convey.Convey("test startSubscribe", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		m := fakeTaskDManager(ctx)
		c := &fakeClient{}
		testRecoverClientHasError(c, m)
		testRecoverClientRecvFailed(c, m)
		testRecoverClientRecvSuccess(c, m, cancel)
	})
}

func testRecoverClientHasError(c *fakeClient, m *BaseManager) {
	convey.Convey("client init failed, should return false", func() {
		patch := gomonkey.ApplyMethodReturn(c, "Init", &pb.Status{}, errors.New("init failed"))
		defer patch.Reset()
		ret := m.startSubscribe(c, &pb.ClientInfo{})
		convey.So(ret, convey.ShouldBeFalse)
	})
	convey.Convey("client register failed,  should return false", func() {
		patch := gomonkey.ApplyMethodReturn(c, "Register",
			&pb.Status{}, errors.New("register failed"))
		defer patch.Reset()
		ret := m.startSubscribe(c, &pb.ClientInfo{})
		convey.So(ret, convey.ShouldBeFalse)
	})
	convey.Convey("client SubscribeProcessManageSignal failed,  should return false", func() {
		patch := gomonkey.ApplyMethodReturn(c, "SubscribeProcessManageSignal",
			nil, errors.New("subscribe process manage signal failed"))
		defer patch.Reset()
		ret := m.startSubscribe(c, &pb.ClientInfo{})
		convey.So(ret, convey.ShouldBeFalse)
	})
}

func testRecoverClientRecvFailed(c *fakeClient, m *BaseManager) {
	patch := gomonkey.ApplyMethodReturn(c, "SubscribeProcessManageSignal",
		&fakeStream{}, nil)
	defer patch.Reset()
	convey.Convey("receive EOF, should return false", func() {
		patch.ApplyMethodReturn(&fakeStream{}, "Recv", nil, io.EOF)
		ret := m.startSubscribe(c, &pb.ClientInfo{})
		convey.So(ret, convey.ShouldBeFalse)
	})
	convey.Convey("Recv return error, should return false", func() {
		patch.ApplyMethodReturn(&fakeStream{}, "Recv", nil, errors.New("recv error"))
		ret := m.startSubscribe(c, &pb.ClientInfo{})
		convey.So(ret, convey.ShouldBeFalse)
	})
	convey.Convey("stream context done, should return false", func() {
		ctx, cancel := context.WithCancel(context.Background())
		patch.ApplyMethodReturn(&fakeStream{}, "Context", ctx)
		cancel()
		ret := m.startSubscribe(c, &pb.ClientInfo{})
		convey.So(ret, convey.ShouldBeFalse)
	})
}

func testRecoverClientRecvSuccess(c *fakeClient, m *BaseManager, cancel context.CancelFunc) {
	patch := gomonkey.ApplyMethodReturn(c, "SubscribeProcessManageSignal",
		&fakeStream{}, nil)
	defer patch.Reset()
	convey.Convey("receive msg and context down, should return true", func() {
		step := 0
		patch.ApplyMethod(&fakeStream{}, "Recv", func() (*pb.ProcessManageSignal, error) {
			if step == 0 {
				cancel()
				step++
				return &pb.ProcessManageSignal{
					Uuid:       string(uuid.NewUUID()),
					SignalType: clusterd_constant.WaitStartAgentSignalType,
				}, nil
			}
			return nil, errors.New("receive error")
		})
		ret := m.startSubscribe(c, &pb.ClientInfo{})
		convey.So(ret, convey.ShouldBeTrue)
		convey.So(len(m.MsgHd.MsgQueue.Queue), convey.ShouldEqual, 1)
		msg, err := m.MsgHd.MsgQueue.Dequeue()
		convey.So(err, convey.ShouldBeNil)
		convey.So(msg.Header.Src.Role, convey.ShouldEqual, common.MgrRole)
		convey.So(msg.Body.Code, convey.ShouldEqual, constant.ProcessManageRecoverSignal)
	})
}

func newBaseManagerForTest() *BaseManager {
	m := &BaseManager{
		Config: Config{JobId: "test-job-id", NodeNums: 1, ProcPerNode: 1},
	}
	m.svcCtx, m.cancelFunc = context.WithCancel(context.Background())
	m.MsgHd = application.NewMsgHandler(m.NodeNums * m.ProcPerNode)
	return m
}

func TestEnqueueStressTest(t *testing.T) {
	convey.Convey("Test BaseManager enqueueStressTest", t, func() {
		m := newBaseManagerForTest()
		defer m.cancelFunc()
		convey.Convey("When called with valid data,enqueue a message with correct code and job ID", func() {
			stressParam := &pb.StressTestRankParams{
				JobId: "test-job-id",
				StressParam: map[string]*pb.StressOpList{
					"test_key": {Ops: []int64{100, 200, 300}},
				},
			}
			m.enqueueStressTest(stressParam)
			dequeuedMsg, err := m.MsgHd.MsgQueue.Dequeue()
			convey.So(err, convey.ShouldBeNil)
			convey.So(dequeuedMsg.Body.Code, convey.ShouldEqual, constant.StressTestCode)
			convey.So(dequeuedMsg.Body.Extension[constant.StressTestJobID], convey.ShouldEqual, stressParam.JobId)
		})
	})
}

func TestSubscribeSwitchNic(t *testing.T) {
	convey.Convey("Test BaseManager.subscribeSwitchNic", t, func() {
		m := newBaseManagerForTest()
		defer m.cancelFunc()
		convey.Convey("When listenSwitchNicSignal returns exit=true, it should break the loop", func() {
			mockStream := &mockRecoverClientStream{
				recvMsg: &pb.SwitchRankList{Op: []bool{true}, RankID: []string{"rank-0"}},
				ctx:     m.svcCtx,
			}
			mockClient := &mockRecoverClient{stream: mockStream}
			patches := gomonkey.ApplyFuncReturn((*BaseManager).listenSwitchNicSignal, true, 0).
				ApplyFuncReturn(pb.NewRecoverClient, mockClient)
			defer patches.Reset()
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				m.subscribeSwitchNic(nil)
			}()
			wg.Wait()
		})
	})
}

func TestListenSwitchNicSignal(t *testing.T) {
	convey.Convey("Test BaseManager.listenSwitchNicSignal", t, func() {
		m := newBaseManagerForTest()
		defer m.cancelFunc()
		convey.Convey("When it receives a message, it should enqueue it and exit", func() {
			mockStream := &mockRecoverClientStream{
				recvMsg: &pb.SwitchRankList{Op: []bool{true}, RankID: []string{"rank-0"}},
				ctx:     m.svcCtx,
			}
			mockClient := &mockRecoverClient{stream: mockStream}
			go func() {
				m.cancelFunc()
			}()
			_, _ = m.listenSwitchNicSignal(mockClient, &pb.ClientInfo{}, 1)
			dequeuedMsg, err := m.MsgHd.MsgQueue.Dequeue()
			convey.So(err, convey.ShouldBeNil)
			convey.So(dequeuedMsg.Body.Code, convey.ShouldEqual, constant.SwitchNicCode)
		})
		convey.Convey("When the gRPC stream returns an error, it should exit without enqueuing", func() {
			mockStream := &mockRecoverClientStream{
				recvErr: io.EOF,
				ctx:     m.svcCtx,
			}
			mockClient := &mockRecoverClient{stream: mockStream}
			exit, _ := m.listenSwitchNicSignal(mockClient, &pb.ClientInfo{}, 1)
			convey.So(exit, convey.ShouldBeFalse)
			_, err := m.MsgHd.MsgQueue.Dequeue()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestEnqueueSwitchNic(t *testing.T) {
	convey.Convey("Test BaseManager.enqueueSwitchNic", t, func() {
		m := newBaseManagerForTest()
		defer m.cancelFunc()
		convey.Convey("When called with valid data, it should enqueue a message with correct code", func() {
			ranks := []string{"rank-0", "rank-1"}
			ops := []bool{true, false}
			m.enqueueSwitchNic(ranks, ops)
			dequeuedMsg, err := m.MsgHd.MsgQueue.Dequeue()
			convey.So(err, convey.ShouldBeNil)
			convey.So(dequeuedMsg.Body.Code, convey.ShouldEqual, constant.SwitchNicCode)
		})
	})
}

func TestSubscribeProfiling(t *testing.T) {
	convey.Convey("Test BaseManager.subscribeProfiling with full line coverage", t, func() {
		m := newBaseManagerForTest()
		defer m.cancelFunc()
		testServer := getTestProfilingServer()
		conn, stopGrpc, err := getTestGrpcServerAndConn(t, testServer)
		convey.So(err, convey.ShouldBeNil)
		defer stopGrpc()
		convey.Convey("When retryTime reaches maxRegRetryTime, it should log and return", func() {
			m.subscribeProfiling(nil, maxRegRetryTime)
		})
		convey.Convey("When creating gRPC stream fails, it should retry", func() {
			subscribeProfilingGrpcFail(m, testServer, conn)
		})
		convey.Convey("When service context is done, it should exit gracefully", func() {
			subscribeProfilingContextDone(m, testServer, conn)
		})
		convey.Convey("When receiving a message successfully, it should enqueue it", func() {
			subscribeProfilingMsgEnqueue(m, testServer, conn)
		})
		convey.Convey("When receiving a message returns an error, it should log and continue", func() {
			subscribeProfilingMsgError(m, testServer, conn)
		})
	})
}

func getTestProfilingServer() *testProfilingServer {
	return &testProfilingServer{
		streamToClose:   make(chan struct{}, 1),
		streamToSendMsg: make(chan *profiling.DataStatusRes, 1),
	}
}

func getTestGrpcServerAndConn(t *testing.T, testServer *testProfilingServer) (
	*grpc.ClientConn, func(), error) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	profiling.RegisterTrainingDataTraceServer(s, testServer)
	go func() {
		if err := s.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("gRPC server exited with error: %v", err)
		}
	}()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(
		ctx context.Context, addr string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithInsecure())
	if err != nil {
		s.Stop()
		return nil, nil, fmt.Errorf("failed to dial bufnet: %w", err)
	}
	stopFunc := func() {
		_ = conn.Close()
		s.Stop()
		_ = lis.Close()
	}
	return conn, stopFunc, nil
}

func subscribeProfilingGrpcFail(m *BaseManager, testServer *testProfilingServer, conn *grpc.ClientConn) {
	testServer.subscribeErr = errors.New("simulated gRPC failure")
	initialGoroutineCount := runtime.NumGoroutine()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.subscribeProfiling(conn, 0)
	}()
	time.Sleep(constant.Hundred * time.Millisecond)
	m.cancelFunc()
	wg.Wait()
	finalGoroutineCount := runtime.NumGoroutine()
	const addGos = 2
	convey.So(finalGoroutineCount+addGos, convey.ShouldBeGreaterThanOrEqualTo, initialGoroutineCount)
}

func subscribeProfilingContextDone(m *BaseManager, testServer *testProfilingServer, conn *grpc.ClientConn) {
	testServer.subscribeErr = nil
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.subscribeProfiling(conn, 0)
	}()
	time.Sleep(constant.Hundred * time.Millisecond)
	m.cancelFunc()
	wg.Wait()
}

func subscribeProfilingMsgEnqueue(m *BaseManager, testServer *testProfilingServer, conn *grpc.ClientConn) {
	testServer.subscribeErr = nil
	protoMsg := &profiling.DataStatusRes{
		ProfilingSwitch: &profiling.ProfilingSwitch{
			CommunicationOperator: "test_domain",
			Step:                  "some_step",
		},
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.subscribeProfiling(conn, 0)
	}()
	time.Sleep(constant.Hundred * time.Millisecond)
	testServer.streamToSendMsg <- protoMsg
	time.Sleep(constant.Hundred * time.Millisecond)
	m.cancelFunc()
	wg.Wait()
	dequeuedMsg, _ := m.MsgHd.MsgQueue.Dequeue()
	convey.So(dequeuedMsg.Body.Message, convey.ShouldEqual, "")
}

func subscribeProfilingMsgError(m *BaseManager, testServer *testProfilingServer, conn *grpc.ClientConn) {
	testServer.subscribeErr = nil
	testServer.streamToSendErr = io.EOF
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.subscribeProfiling(conn, 0)
	}()
	time.Sleep(constant.Hundred * time.Millisecond)
	m.cancelFunc()
	wg.Wait()
	_, err := m.MsgHd.MsgQueue.Dequeue()
	convey.So(err, convey.ShouldNotBeNil)
}

type mockRecoverClientStream struct {
	grpc.ClientStream
	recvMsg *pb.SwitchRankList
	recvErr error
	ctx     context.Context
}

func (m *mockRecoverClientStream) Recv() (*pb.SwitchRankList, error) {
	select {
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	default:
	}
	if m.recvErr != nil {
		return nil, m.recvErr
	}
	msg := m.recvMsg
	m.recvMsg = nil
	return msg, nil
}

func (m *mockRecoverClientStream) Context() context.Context {
	return m.ctx
}

type mockRecoverClient struct {
	pb.RecoverClient
	stream *mockRecoverClientStream
}

func (m *mockRecoverClient) SubscribeNotifySwitch(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (pb.Recover_SubscribeNotifySwitchClient, error) {
	m.stream.ctx = ctx
	return m.stream, nil
}

type testProfilingServer struct {
	profiling.UnimplementedTrainingDataTraceServer
	subscribeErr    error
	streamToClose   chan struct{}
	streamToSendMsg chan *profiling.DataStatusRes
	streamToSendErr error
}

func (s *testProfilingServer) SubscribeDataTraceSwitch(req *profiling.ProfilingClientInfo, stream profiling.TrainingDataTrace_SubscribeDataTraceSwitchServer) error {
	if s.subscribeErr != nil {
		return s.subscribeErr
	}
	for {
		select {
		case <-s.streamToClose:
			return errors.New("stream closed by test")
		case msg := <-s.streamToSendMsg:
			if err := stream.Send(msg); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		default:
			if s.streamToSendErr != nil {
				errToReturn := s.streamToSendErr
				s.streamToSendErr = nil
				return errToReturn
			}
			time.Sleep(constant.Ten * time.Millisecond)
		}
	}
}
