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
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"

	"clusterd/pkg/interface/grpc/profiling"
	"clusterd/pkg/interface/grpc/recover"
	"taskd/common/constant"
	_ "taskd/common/testtool"
	"taskd/common/utils"
	"taskd/framework_backend/manager/application"
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

type fakeProfilingClient struct {
	profiling.TrainingDataTraceClient
}

func (f *fakeProfilingClient) SubscribeDataTraceSwitch(ctx context.Context, in *profiling.ProfilingClientInfo, opts ...grpc.CallOption) (profiling.TrainingDataTrace_SubscribeDataTraceSwitchClient, error) {
	return nil, nil
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
