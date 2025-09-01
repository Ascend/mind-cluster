// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package recover a series of service function
package recover

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/recover"
	"clusterd/pkg/interface/kube"
)

func TestCanDoSwitchingNic(t *testing.T) {
	t.Run("can do switch nic ", func(t *testing.T) {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		res := ctl.canDoSwitchingNic()
		assert.Equal(t, true, res)
	})
}

func TestGetSwitchNicParam(t *testing.T) {
	t.Run("get switch nic param", func(t *testing.T) {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ranks, ops := ctl.getSwitchNicParam()
		assert.Equal(t, 0, len(ranks))
		assert.Equal(t, 0, len(ops))
	})

}

func TestIsSwitchingNic(t *testing.T) {
	t.Run("is not switching nic ", func(t *testing.T) {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		res := ctl.isSwitchingNic()
		assert.Equal(t, false, res)
	})
}

func TestSetSwitchNicParam(t *testing.T) {
	t.Run("set param success ", func(t *testing.T) {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.setSwitchNicParam([]string{"1"}, []bool{true})
		assert.Equal(t, "1", ctl.globalSwitchRankIDs[0])
		assert.Equal(t, true, ctl.globalOps[0])
	})
}

func TestSelectSendChannelSendSwitchNicChanNil(t *testing.T) {
	convey.Convey("Test chan when sendChan is nil", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &switchNicSender{}
		ctl.selectSendSwitchNicResponseChan(context.Background(), nil, stream)
	})
}

func TestSelectSendSwitchNicChanContextDone(t *testing.T) {
	convey.Convey("Test chan when context is done", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &switchNicSender{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		sendChan := make(chan *pb.SwitchNicResponse)
		ctl.selectSendSwitchNicResponseChan(ctx, sendChan, stream)
	})
}

func TestSelectSwitchNicSendChannelReceiveSignal(t *testing.T) {
	convey.Convey("Test chan when receive signal from sendChan", t, func() {
		jobID := "testJobId"
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: jobID,
			},
		}
		stream := &switchNicSender{}
		ctx := context.Background()
		sendChan := make(chan *pb.SwitchNicResponse, 1)
		signal := &pb.SwitchNicResponse{JobID: jobID}
		sendChan <- signal
		called := false
		patchSendRetry := gomonkey.ApplyFunc(common.SendWithRetry[pb.SwitchRankList, *notifySwitchNicSender],
			func(stream *notifySwitchNicSender, signal *pb.SwitchRankList, retryTimes int) error {
				called = true
				return nil
			})
		defer patchSendRetry.Reset()
		ctl.selectSendSwitchNicResponseChan(ctx, sendChan, stream)
		convey.ShouldBeTrue(called)
	})
}

func TestSelectSendSwitchNicSendChanClosed(t *testing.T) {
	convey.Convey("Test Chan when sendChan is closed", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &switchNicSender{}
		ctx := context.Background()
		sendChan := make(chan *pb.SwitchNicResponse, 1)
		close(sendChan)
		ctl.selectSendSwitchNicResponseChan(ctx, sendChan, stream)
		_, ok := <-sendChan
		convey.So(ok, convey.ShouldBeFalse)
	})
}

func TestGetCtxAndSwitchNicChan(t *testing.T) {
	convey.Convey("Testing getCtxAndSwitchNicResponseChan", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctx, ch := ctl.getCtxAndSwitchNicResponseChan()
		convey.So(ctx, convey.ShouldNotBeNil)
		convey.So(ch, convey.ShouldNotBeNil)
	})
}

func TestListenSwitchNicChannel(t *testing.T) {
	convey.Convey("Test listenSwitchNicChannel", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{JobId: "test-job-id"},
		}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendChan := make(chan *pb.SwitchNicResponse, 1)
		patches.ApplyPrivateMethod(ctl, "getCtxAndSwitchNicResponseChan",
			func() (context.Context, chan *pb.SwitchNicResponse) {
				return ctx, sendChan
			})
		patches.ApplyPrivateMethod(ctl, "reset", func() {})
		patches.ApplyPrivateMethod(ctl, "selectSendSwitchNicResponseChan", func(_ context.Context,
			_ chan *pb.SwitchNicResponse, _ pb.Recover_SubscribeSwitchNicSignalServer) {
			return
		})

		stream := &switchNicSender{}
		ctl.listenSwitchNicChannel(stream)
		convey.So(true, convey.ShouldBeTrue)

	})
}

func TestHandleWaitContinueTrainComplete(t *testing.T) {
	convey.Convey("Test handleDecideContinueTrainComplete", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {}).
			ApplyFunc(hwlog.RunLog.Warnf, func(format string, args ...interface{}) {}).
			ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})
		defer patches.Reset()

		testSwitchNicReportChanNil(ctl)
		testSwitchNicContextCanceled(ctl)
		testSwitchNicTrainError(ctl)
		testSwitchNicValidReport(ctl)
		testStressTestValidReport(ctl)
	})
}

func testSwitchNicReportChanNil(ctl *EventController) {
	convey.Convey("When reportChan is nil", func() {
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndResultChan",
			func() (context.Context, chan *pb.RecoverStatusRequest) {
				return context.Background(), nil
			})
		defer patches.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		event, respCode, err := ctl.handleDecideContinueTrainComplete()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.ServerInnerError)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId))
	})
}

func testSwitchNicContextCanceled(ctl *EventController) {
	convey.Convey("When context is canceled", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndResultChan",
			func() (context.Context, chan *pb.RecoverStatusRequest) {
				return ctx, make(chan *pb.RecoverStatusRequest, 1)
			})
		defer patches.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		event, respCode, err := ctl.handleDecideContinueTrainComplete()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.ControllerEventCancel)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testSwitchNicTrainError(ctl *EventController) {
	convey.Convey("continue train when receiving a report with train error", func() {
		reportChan := make(chan *pb.RecoverStatusRequest, 1)
		reportChan <- &pb.RecoverStatusRequest{Status: &pb.Status{Code: common.UnRecoverTrainError}}
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndResultChan",
			func() (context.Context, chan *pb.RecoverStatusRequest) {
				return context.Background(), reportChan
			})
		defer patches.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		event, respCode, err := ctl.handleDecideContinueTrainComplete()
		convey.So(event, convey.ShouldEqual, common.ContinueTrainFailEvent)
		convey.So(respCode, convey.ShouldEqual, common.ClientError)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testSwitchNicValidReport(ctl *EventController) {
	convey.Convey("When receiving a valid report, when switch nic", func() {
		reportChan := make(chan *pb.RecoverStatusRequest, 1)
		reportChan <- &pb.RecoverStatusRequest{Status: &pb.Status{Code: int32(common.OK)}}
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndResultChan",
			func() (context.Context, chan *pb.RecoverStatusRequest) {
				return context.Background(), reportChan
			})
		defer patches.Reset()
		patches2 := gomonkey.ApplyPrivateMethod(ctl, "isSwitchingNic",
			func() bool {
				return true
			})
		defer patches2.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		event, respCode, err := ctl.handleDecideContinueTrainComplete()
		convey.So(event, convey.ShouldEqual, common.SwitchNicRecvContinueEvent)
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testStressTestValidReport(ctl *EventController) {
	convey.Convey("When receiving a valid report, when stress test", func() {
		reportChan := make(chan *pb.RecoverStatusRequest, 1)
		reportChan <- &pb.RecoverStatusRequest{Status: &pb.Status{Code: int32(common.OK)}}
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndResultChan",
			func() (context.Context, chan *pb.RecoverStatusRequest) {
				return context.Background(), reportChan
			})
		defer patches.Reset()
		patches2 := gomonkey.ApplyPrivateMethod(ctl, "isStressTest",
			func() bool {
				return true
			})
		defer patches2.Reset()
		event, respCode, err := ctl.handleDecideContinueTrainComplete()
		convey.So(event, convey.ShouldEqual, common.StressTestRecvContinueEvent)
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleWaitSwitchNicFinish(t *testing.T) {
	convey.Convey("Test handleWaitSwitchNicFinish", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {}).
			ApplyFunc(hwlog.RunLog.Warnf, func(format string, args ...interface{}) {}).
			ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})
		defer patches.Reset()

		testSwitchNicFinishNil(ctl)
		testSwitchNicFinishCanceled(ctl)
		testSwitchNicFinishNotReady(ctl)
		testSwitchNicFinishFailed(ctl)
		testSwitchNicFinishValidReport(ctl)
	})
}

func testSwitchNicFinishNil(ctl *EventController) {
	convey.Convey("When reportChan is nil", func() {
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndResultChan",
			func() (context.Context, chan *pb.RecoverStatusRequest) {
				return context.Background(), nil
			})
		defer patches.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		resultChan := make(chan *pb.SwitchResult, 1)
		pat := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndSwitchNicResultChan",
			func() (context.Context, chan *pb.SwitchResult) {
				return context.Background(), resultChan
			})
		defer pat.Reset()
		defer func() {
			close(ctl.switchRankResult)
			ctl.switchRankResult = make(chan *pb.SwitchResult, 1)
		}()
		event, respCode, err := ctl.handleWaitSwitchNicFinish()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.ServerInnerError)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId))
	})
}

func testSwitchNicFinishCanceled(ctl *EventController) {
	convey.Convey("When context is canceled", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndResultChan",
			func() (context.Context, chan *pb.RecoverStatusRequest) {
				return ctx, make(chan *pb.RecoverStatusRequest, 1)
			})
		defer patches.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		resultChan := make(chan *pb.SwitchResult, 1)
		pat := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndSwitchNicResultChan",
			func() (context.Context, chan *pb.SwitchResult) {
				return ctx, resultChan
			})
		defer pat.Reset()
		defer func() {
			close(ctl.switchRankResult)
			ctl.switchRankResult = make(chan *pb.SwitchResult, 1)
		}()
		event, respCode, err := ctl.handleWaitSwitchNicFinish()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.ControllerEventCancel)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testSwitchNicFinishNotReady(ctl *EventController) {
	convey.Convey("When receiving a report with unRecover train error", func() {
		reportChan := make(chan *pb.RecoverStatusRequest, 1)
		reportChan <- &pb.RecoverStatusRequest{Status: &pb.Status{Code: common.UnRecoverableRetryError}}
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndResultChan",
			func() (context.Context, chan *pb.RecoverStatusRequest) {
				return context.Background(), reportChan
			})
		defer patches.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		resultChan := make(chan *pb.SwitchResult, 1)
		pat := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndSwitchNicResultChan",
			func() (context.Context, chan *pb.SwitchResult) {
				return context.Background(), resultChan
			})
		defer pat.Reset()
		defer func() {
			close(ctl.switchRankResult)
			ctl.switchRankResult = make(chan *pb.SwitchResult, 1)
		}()
		event, respCode, err := ctl.handleWaitSwitchNicFinish()
		convey.So(event, convey.ShouldEqual, common.WaitSwitchNicRecvFaultEvent)
		convey.So(respCode, convey.ShouldEqual, common.ClientError)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testSwitchNicFinishFailed(ctl *EventController) {
	convey.Convey("When receiving a report with switch nic error", func() {
		resultChan := make(chan *pb.SwitchResult, 1)
		resultChan <- &pb.SwitchResult{JobId: ctl.jobInfo.JobId, Result: false}
		pat := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndSwitchNicResultChan",
			func() (context.Context, chan *pb.SwitchResult) {
				return context.Background(), resultChan
			})
		defer pat.Reset()
		defer func() {
			close(ctl.switchRankResult)
			ctl.switchRankResult = make(chan *pb.SwitchResult, 1)
		}()
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndResultChan",
			func() (context.Context, chan *pb.RecoverStatusRequest) {
				return context.Background(), make(chan *pb.RecoverStatusRequest, 1)
			})
		defer patches.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		event, respCode, err := ctl.handleWaitSwitchNicFinish()
		convey.So(event, convey.ShouldEqual, common.SwitchNicFailEvent)
		convey.So(respCode, convey.ShouldEqual, common.ClientError)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testSwitchNicFinishValidReport(ctl *EventController) {
	convey.Convey("When receiving a valid report", func() {
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndResultChan",
			func() (context.Context, chan *pb.RecoverStatusRequest) {
				return context.Background(), make(chan *pb.RecoverStatusRequest, 1)
			})
		defer patches.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		resultChan := make(chan *pb.SwitchResult, 1)
		resultChan <- &pb.SwitchResult{JobId: ctl.jobInfo.JobId, Result: true}
		pat := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndSwitchNicResultChan",
			func() (context.Context, chan *pb.SwitchResult) {
				return context.Background(), resultChan
			})
		defer pat.Reset()
		defer func() {
			close(ctl.switchRankResult)
			ctl.switchRankResult = make(chan *pb.SwitchResult, 1)
		}()
		event, respCode, err := ctl.handleWaitSwitchNicFinish()
		convey.So(event, convey.ShouldEqual, common.ReceiveReportEvent)
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleSwitchNicFinish(t *testing.T) {
	convey.Convey("Test handleSwitchNicFinish", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		patches := gomonkey.ApplyPrivateMethod(ctl, "reset", func() { return })
		defer patches.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		event, respCode, err := ctl.handleSwitchNicFinish()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestNotifyContinueTrain(t *testing.T) {
	convey.Convey("Test notifyContinueTrain", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		convey.Convey("signal enqueue success, should return nil", func() {
			mockFunc := gomonkey.ApplyPrivateMethod(&EventController{}, "signalEnqueue",
				func(signal *pb.ProcessManageSignal) (string, common.RespCode, error) { return "", common.OK, nil })
			defer mockFunc.Reset()
			_, respCode, err := ctl.notifyContinueTrain()
			convey.So(err, convey.ShouldBeNil)
			convey.So(respCode == common.OK, convey.ShouldBeTrue)
		})
	})
}

func TestHandleNotifyPauseTrain(t *testing.T) {
	convey.Convey("Test notifyContinueTrain", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		convey.Convey("signal enqueue success, should return nil", func() {
			mockFunc := gomonkey.ApplyPrivateMethod(&EventController{}, "signalEnqueue",
				func(signal *pb.ProcessManageSignal) (string, common.RespCode, error) { return "", common.OK, nil })
			defer mockFunc.Reset()
			_, respCode, err := ctl.handleNotifyPauseTrain()
			convey.So(err, convey.ShouldBeNil)
			convey.So(respCode == common.OK, convey.ShouldBeTrue)
		})
	})
}

func TestHandleWaitReportPauseTrainComplete(t *testing.T) {
	convey.Convey("Test handleWaitPauseTrainComplete", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {}).
			ApplyFunc(hwlog.RunLog.Warnf, func(format string, args ...interface{}) {}).
			ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})
		defer patches.Reset()

		testSwitchNicPauseTrainNil(ctl)
		testSwitchNicPauseTrainCanceled(ctl)
		testSwitchNicPauseTrainError(ctl)
		testSwitchNicPauseTrainValidReport(ctl)
		testStressTestPauseTrainValidReport(ctl)
	})
}

func testSwitchNicPauseTrainNil(ctl *EventController) {
	convey.Convey("When reportChan is nil", func() {
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndStopCompleteChan",
			func() (context.Context, chan *pb.StopCompleteRequest) {
				return context.Background(), nil
			})
		defer patches.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		event, respCode, err := ctl.handleWaitPauseTrainComplete()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.ServerInnerError)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId))
	})
}

func testSwitchNicPauseTrainCanceled(ctl *EventController) {
	convey.Convey("When context is canceled", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndStopCompleteChan",
			func() (context.Context, chan *pb.StopCompleteRequest) {
				return ctx, make(chan *pb.StopCompleteRequest, 1)
			})
		defer patches.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		event, respCode, err := ctl.handleWaitPauseTrainComplete()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.ControllerEventCancel)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testSwitchNicPauseTrainError(ctl *EventController) {
	convey.Convey("When receiving a report with unRecover train error", func() {
		reportChan := make(chan *pb.StopCompleteRequest, 1)
		reportChan <- &pb.StopCompleteRequest{Status: &pb.Status{Code: common.UnRecoverTrainError}}
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndStopCompleteChan",
			func() (context.Context, chan *pb.StopCompleteRequest) {
				return context.Background(), reportChan
			})
		defer patches.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		event, respCode, err := ctl.handleWaitPauseTrainComplete()
		convey.So(event, convey.ShouldEqual, common.ProcessPauseFailEvent)
		convey.So(respCode, convey.ShouldEqual, common.ClientError)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testSwitchNicPauseTrainValidReport(ctl *EventController) {
	convey.Convey("When receiving a valid report, when switch nic", func() {
		reportChan := make(chan *pb.StopCompleteRequest, 1)
		reportChan <- &pb.StopCompleteRequest{Status: &pb.Status{Code: int32(common.OK)}}
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndStopCompleteChan",
			func() (context.Context, chan *pb.StopCompleteRequest) {
				return context.Background(), reportChan
			})
		defer patches.Reset()
		patches2 := gomonkey.ApplyPrivateMethod(ctl, "isSwitchingNic",
			func() bool {
				return true
			})
		defer patches2.Reset()
		defer func() {
			close(ctl.switchNicResponse)
			ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		}()
		event, respCode, err := ctl.handleWaitPauseTrainComplete()
		convey.So(event, convey.ShouldEqual, common.SwitchNicRecvPauseEvent)
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testStressTestPauseTrainValidReport(ctl *EventController) {
	convey.Convey("When receiving a valid report, when stress test", func() {
		reportChan := make(chan *pb.StopCompleteRequest, 1)
		reportChan <- &pb.StopCompleteRequest{Status: &pb.Status{Code: int32(common.OK)}}
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndStopCompleteChan",
			func() (context.Context, chan *pb.StopCompleteRequest) {
				return context.Background(), reportChan
			})
		defer patches.Reset()
		patches2 := gomonkey.ApplyPrivateMethod(ctl, "isStressTest",
			func() bool {
				return true
			})
		defer patches2.Reset()
		event, respCode, err := ctl.handleWaitPauseTrainComplete()
		convey.So(event, convey.ShouldEqual, common.StressTestRecvPauseEvent)
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestSwitchNicSignalEnqueueEnqueue(t *testing.T) {
	convey.Convey("Testing switchNicSignalEnqueue", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		signal := &pb.SwitchRankList{
			JobId:  "test-job",
			RankID: make([]string, 0),
			Op:     make([]bool, 0),
		}
		_, code, err := ctl.switchNicSignalEnqueue(signal)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetCtxAndSwitchNicNotifyChan(t *testing.T) {
	convey.Convey("Testing getCtxAndSwitchNicNotifyChan", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctx, ch := ctl.getCtxAndSwitchNicNotifyChan()
		convey.So(ctx, convey.ShouldNotBeNil)
		convey.So(ch, convey.ShouldNotBeNil)
	})
}

func TestListenSwitchNicNotifyChannel(t *testing.T) {
	convey.Convey("Test listenSwitchNicNotifyChannel", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{JobId: "test-job-id"},
		}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendChan := make(chan *pb.SwitchRankList, 1)
		patches.ApplyPrivateMethod(ctl, "getCtxAndSwitchNicNotifyChan",
			func() (context.Context, chan *pb.SwitchRankList) {
				return ctx, sendChan
			})
		patches.ApplyPrivateMethod(ctl, "reset", func() {})
		patches.ApplyPrivateMethod(ctl, "selectNotifySwitchNic", func(_ context.Context,
			_ chan *pb.SwitchRankList, _ pb.Recover_SubscribeNotifySwitchServer) bool {
			return true
		})

		stream := &notifySwitchNicSender{}
		ctl.listenSwitchNicNotifyChannel(stream)
		convey.So(true, convey.ShouldBeTrue)

	})
}

func TestSelectNotifySwitchNicNil(t *testing.T) {
	convey.Convey("Test chan when sendChan is nil", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &notifySwitchNicSender{}
		res := ctl.selectNotifySwitchNic(context.Background(), nil, stream)
		convey.ShouldBeTrue(res)
	})
}

func TestSelectNotifySwitchNicContextDone(t *testing.T) {
	convey.Convey("Test chan when context is done", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &notifySwitchNicSender{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		sendChan := make(chan *pb.SwitchRankList)
		res := ctl.selectNotifySwitchNic(ctx, sendChan, stream)
		convey.ShouldBeTrue(res)
	})
}

func fooSendWithRetry[T any, S common.StreamSender[T]](stream S, signal *T, retryTimes int) error {
	signal = nil
	return nil
}

func TestSelectNotifySwitchNicReceiveSignal(t *testing.T) {
	convey.Convey("Test chan when receive signal from sendChan", t, func() {
		jobID := "testJobId"
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: jobID,
			},
		}
		var rules []common.TransRule = ctl.getBaseRules()
		ctl.state = common.NewStateMachine(common.InitState, rules)
		stream := &notifySwitchNicSender{}
		ctx := context.Background()
		sendChan := make(chan *pb.SwitchRankList, 1)
		signal := &pb.SwitchRankList{JobId: jobID}
		sendChan <- signal
		called := false
		patchSendRetry := gomonkey.ApplyFunc(common.SendWithRetry[pb.SwitchRankList, *notifySwitchNicSender],
			func(stream *notifySwitchNicSender, signal *pb.SwitchRankList, retryTimes int) error {
				called = true
				return nil
			})
		defer patchSendRetry.Reset()
		ctl.selectNotifySwitchNic(ctx, sendChan, stream)
		convey.ShouldBeTrue(called)
	})
}

func TestSelectNotifySwitchNicClosed(t *testing.T) {
	convey.Convey("Test Chan when sendChan is closed", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &notifySwitchNicSender{}
		ctx := context.Background()
		sendChan := make(chan *pb.SwitchRankList, 1)
		close(sendChan)
		res := ctl.selectNotifySwitchNic(ctx, sendChan, stream)
		_, ok := <-sendChan
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(res, convey.ShouldBeTrue)
	})
}

func TestReplyOMResponse(t *testing.T) {
	convey.Convey("replyOMResponse, reply stress test", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.setStressTestParam(common.StressTestParam{
			"node": make(map[string][]int64),
		})
		ctl.replyOMResponse("test")
		msg := <-ctl.stressTestResponse
		convey.So(msg.Msg, convey.ShouldEqual, "test")
	})
	convey.Convey("replyOMResponse, reply switch nic", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.setSwitchNicParam([]string{"test"}, []bool{true})
		ctl.replyOMResponse("test")
		msg := <-ctl.switchNicResponse
		convey.So(msg.Msg, convey.ShouldEqual, "test")
	})
}

func TestSetStressTestParam(t *testing.T) {
	t.Run("set param success ", func(t *testing.T) {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.setStressTestParam(common.StressTestParam{
			"node": make(map[string][]int64),
		})
		assert.Equal(t, 1, len(ctl.stressTestParam))
	})
}

func TestIsStressTest(t *testing.T) {
	t.Run("is not stress test ", func(t *testing.T) {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		res := ctl.isStressTest()
		assert.Equal(t, false, res)
	})
}

func TestGetCtxAndStressTestNotifyChan(t *testing.T) {
	convey.Convey("Testing GetCtxAndStressTestNotifyChan", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctx, ch := ctl.getCtxAndStressTestNotifyChan()
		convey.So(ctx, convey.ShouldNotBeNil)
		convey.So(ch, convey.ShouldNotBeNil)
	})
}

func TestListenStressTestNotifyChannel(t *testing.T) {
	convey.Convey("Test listenStressTestNotifyChannel", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{JobId: "test-job-id"},
		}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendChan := make(chan *pb.StressTestRankParams, 1)
		patches.ApplyPrivateMethod(ctl, "getCtxAndStressTestNotifyChan",
			func() (context.Context, chan *pb.StressTestRankParams) {
				return ctx, sendChan
			})
		patches.ApplyPrivateMethod(ctl, "reset", func() {})
		patches.ApplyPrivateMethod(ctl, "selectNotifyStressTest", func(_ context.Context,
			_ chan *pb.StressTestRankParams, _ pb.Recover_SubscribeNotifyExecStressTestServer) bool {
			return true
		})

		stream := &notifyStressTestSender{}
		ctl.listenStressTestNotifyChannel(stream)
		convey.So(true, convey.ShouldBeTrue)

	})
}

func TestSelectNotifyStressTestNil(t *testing.T) {
	convey.Convey("Test chan when sendChan is nil", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &notifyStressTestSender{}
		res := ctl.selectNotifyStressTest(context.Background(), nil, stream)
		convey.ShouldBeTrue(res)
	})
}

func TestSelectNotifyStressTestContextDone(t *testing.T) {
	convey.Convey("Test chan when context is done", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &notifyStressTestSender{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		sendChan := make(chan *pb.StressTestRankParams)
		res := ctl.selectNotifyStressTest(ctx, sendChan, stream)
		convey.ShouldBeTrue(res)
	})
}

func TestSelectNotifyStressTestReceiveSignal(t *testing.T) {
	convey.Convey("Test chan when receive signal from sendChan", t, func() {
		jobID := "testJobId"
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: jobID,
			},
		}
		var rules []common.TransRule = ctl.getBaseRules()
		ctl.state = common.NewStateMachine(common.InitState, rules)
		stream := &notifyStressTestSender{}
		ctx := context.Background()
		sendChan := make(chan *pb.StressTestRankParams, 1)
		signal := &pb.StressTestRankParams{JobId: jobID}
		sendChan <- signal
		called := false
		patchSendRetry := gomonkey.ApplyFunc(common.SendWithRetry[pb.StressTestRankParams, *notifyStressTestSender],
			func(stream *notifyStressTestSender, signal *pb.StressTestRankParams, retryTimes int) error {
				called = true
				return nil
			})
		defer patchSendRetry.Reset()
		ctl.selectNotifyStressTest(ctx, sendChan, stream)
		convey.ShouldBeTrue(called)
	})
}

func TestSelectNotifyStressTestClosed(t *testing.T) {
	convey.Convey("Test Chan when sendChan is closed", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &notifyStressTestSender{}
		ctx := context.Background()
		sendChan := make(chan *pb.StressTestRankParams, 1)
		close(sendChan)
		res := ctl.selectNotifyStressTest(ctx, sendChan, stream)
		_, ok := <-sendChan
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(res, convey.ShouldBeTrue)
	})
}

func TestSetStressTestResult(t *testing.T) {
	convey.Convey("Test setStressTestResult ok", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.setStressTestResult(&pb.StressTestResult{
			JobId: jobInfo.JobId,
		})
		ret := <-ctl.stressTestResult
		convey.So(ret.JobId, convey.ShouldEqual, jobInfo.JobId)
	})
}

func TestGetCtxAndStressTestResponseChan(t *testing.T) {
	convey.Convey("Testing getCtxAndStressTestResponseChan, ok", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctx, ch := ctl.getCtxAndStressTestResponseChan()
		convey.So(ctx, convey.ShouldNotBeNil)
		convey.So(ch, convey.ShouldNotBeNil)
	})
}

func TestListenStressTestChannel(t *testing.T) {
	convey.Convey("Test listenStressTestChannel", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{JobId: "test-job-id"},
		}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendChan := make(chan *pb.StressTestResult, 1)
		patches.ApplyPrivateMethod(ctl, "getCtxAndStressTestResponseChan",
			func() (context.Context, chan *pb.StressTestResult) {
				return ctx, sendChan
			})
		patches.ApplyPrivateMethod(ctl, "reset", func() {})
		patches.ApplyPrivateMethod(ctl, "selectSendStressTestResponseChan", func(_ context.Context, _ chan *pb.StressTestResponse,
			_ pb.Recover_SubscribeStressTestResponseServer) {
			return
		})

		stream := &stressTestSender{}
		ctl.listenStressTestChannel(stream)
		convey.So(true, convey.ShouldBeTrue)

	})
}

func TestSelectSendStressTestResponseChanNil(t *testing.T) {
	convey.Convey("Test chan when sendChan is nil", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &stressTestSender{}
		ctl.selectSendStressTestResponseChan(context.Background(), nil, stream)
	})
}

func TestSelectSendStressTestResponseChanContextDone(t *testing.T) {
	convey.Convey("Test chan when context is done", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &stressTestSender{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		sendChan := make(chan *pb.StressTestResponse)
		called := false
		patchSendRetry := gomonkey.ApplyFunc(common.SendWithRetry[pb.StressTestResponse, pb.Recover_SubscribeStressTestResponseServer],
			func(stream pb.Recover_SubscribeStressTestResponseServer, signal *pb.StressTestResponse, retryTimes int) error {
				called = true
				return nil
			})
		defer patchSendRetry.Reset()
		ctl.selectSendStressTestResponseChan(ctx, sendChan, stream)
		convey.So(called, convey.ShouldBeFalse)
	})
}

func TestSelectSendStressTestResponseChanReceiveSignal(t *testing.T) {
	convey.Convey("Test chan when receive signal from sendChan", t, func() {
		jobID := "testJobId"
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: jobID,
			},
		}
		var rules []common.TransRule = ctl.getBaseRules()
		ctl.state = common.NewStateMachine(common.InitState, rules)
		stream := &stressTestSender{}
		ctx := context.Background()
		sendChan := make(chan *pb.StressTestResponse, 1)
		signal := &pb.StressTestResponse{JobID: jobID}
		sendChan <- signal
		called := false
		patchSendRetry := gomonkey.ApplyFunc(common.SendWithRetry[pb.StressTestResponse, *stressTestSender],
			func(stream *stressTestSender, signal *pb.StressTestResponse, retryTimes int) error {
				called = true
				return nil
			})
		defer patchSendRetry.Reset()
		ctl.selectSendStressTestResponseChan(ctx, sendChan, stream)
		convey.ShouldBeTrue(called)
	})
}

func TestSelectSendStressTestResponseChanClosed(t *testing.T) {
	convey.Convey("Test Chan when sendChan is closed", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &stressTestSender{}
		ctx := context.Background()
		sendChan := make(chan *pb.StressTestResponse, 1)
		close(sendChan)
		ctl.selectSendStressTestResponseChan(ctx, sendChan, stream)
		_, ok := <-sendChan
		convey.So(ok, convey.ShouldBeFalse)
	})
}

func TestGetStressTestParam(t *testing.T) {
	t.Run("get stress test param", func(t *testing.T) {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		res := ctl.getStressTestParam()
		assert.NotNil(t, res)
	})
}

func TestGetCtxAndStressTestResultChan(t *testing.T) {
	convey.Convey("Testing getCtxAndStressTestResultChan, ok", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctx, ch := ctl.getCtxAndStressTestResultChan()
		convey.So(ctx, convey.ShouldNotBeNil)
		convey.So(ch, convey.ShouldNotBeNil)
	})
}

func TestHandleStressTestFinish(t *testing.T) {
	convey.Convey("Test handleStressTestFinish, ok", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		patches := gomonkey.ApplyPrivateMethod(ctl, "reset", func() { return })
		defer patches.Reset()
		defer func() {
			close(ctl.stressTestResult)
			ctl.stressTestResult = make(chan *pb.StressTestResult, 1)
		}()
		event, respCode, err := ctl.handleStressTestFinish()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestNotifyStressTest(t *testing.T) {
	convey.Convey("Test notifyStressTest, ok", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		patches := gomonkey.ApplyPrivateMethod(ctl, "stressTestSignalEnqueue",
			func(signal *pb.StressTestRankParams) (string, common.RespCode, error) { return "", common.OK, nil })
		defer patches.Reset()
		ctl.stressTestParam = common.StressTestParam{
			"node": map[string][]int64{
				"rank": {0},
			},
		}
		event, respCode, err := ctl.notifyStressTest()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestStressTestSignalEnqueue(t *testing.T) {
	convey.Convey("Testing stressTestSignalEnqueue, ok", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		signal := &pb.StressTestRankParams{
			JobId:       "test-job",
			StressParam: map[string]*pb.StressOpList{},
		}
		_, code, err := ctl.stressTestSignalEnqueue(signal)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("Testing stressTestSignalEnqueue, ctx done", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		ctx, cancel := context.WithCancel(context.Background())
		ctl := NewEventController(jobInfo, keepAliveSeconds, ctx)
		paramChan := make(chan *pb.StressTestRankParams, 1)
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndStressTestNotifyChan", func() (context.Context, chan *pb.StressTestRankParams) {
			return ctx, paramChan
		})
		defer patches.Reset()
		signal := &pb.StressTestRankParams{
			JobId:       "test-job",
			StressParam: map[string]*pb.StressOpList{},
		}
		cancel()
		_, code, err := ctl.stressTestSignalEnqueue(signal)
		convey.So(code, convey.ShouldEqual, common.ControllerEventCancel)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleWaitStressTestFinish(t *testing.T) {
	convey.Convey("Test handleWaitStressTestFinish", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {}).
			ApplyFunc(hwlog.RunLog.Warnf, func(format string, args ...interface{}) {}).
			ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})
		defer patches.Reset()

		testStressTestFinishReportChanNil(ctl)
		testStressTestFinishResultChanNil(ctl)
		testStressTestFinishValidReport(ctl)
	})
}

func testStressTestFinishReportChanNil(ctl *EventController) {
	convey.Convey("When reportChan is nil", func() {
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndResultChan",
			func() (context.Context, chan *pb.RecoverStatusRequest) {
				return context.Background(), nil
			})
		defer patches.Reset()
		defer func() {
			close(ctl.stressTestResponse)
			ctl.stressTestResponse = make(chan *pb.StressTestResponse, 1)
		}()
		resultChan := make(chan *pb.StressTestResult, 1)
		pat := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndStressTestResultChan",
			func() (context.Context, chan *pb.StressTestResult) {
				return context.Background(), resultChan
			}).ApplyFunc(faultmanager.FilterStressTestFault, func(jobID string, nodes []string, val bool) {}).
			ApplyFunc(common.RetryWriteResetCM, func(taskName, nameSpace string, faultRankList []string, restartFaultProcess bool,
				operator string) (*v1.ConfigMap, error) {
				return &v1.ConfigMap{}, nil
			})
		defer pat.Reset()
		defer func() {
			close(ctl.stressTestResult)
			ctl.stressTestResult = make(chan *pb.StressTestResult, 1)
		}()
		ctl.stressTestParam = common.StressTestParam{
			"node": map[string][]int64{
				"rank": {0},
			},
		}
		event, respCode, _ := ctl.handleWaitStressTestFinish()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.ServerInnerError)
	})
}

func testStressTestFinishResultChanNil(ctl *EventController) {
	convey.Convey("When ResultChan is nil", func() {
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndResultChan",
			func() (context.Context, chan *pb.RecoverStatusRequest) {
				return context.Background(), make(chan *pb.RecoverStatusRequest, 1)
			})
		defer patches.Reset()
		defer func() {
			close(ctl.stressTestResponse)
			ctl.stressTestResponse = make(chan *pb.StressTestResponse, 1)
		}()
		pat := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndStressTestResultChan",
			func() (context.Context, chan *pb.StressTestResult) {
				return context.Background(), nil
			}).ApplyFunc(faultmanager.FilterStressTestFault, func(jobID string, nodes []string, val bool) {}).
			ApplyFunc(common.RetryWriteResetCM, func(taskName, nameSpace string, faultRankList []string, restartFaultProcess bool,
				operator string) (*v1.ConfigMap, error) {
				return &v1.ConfigMap{}, nil
			})
		defer pat.Reset()
		defer func() {
			close(ctl.stressTestResult)
			ctl.stressTestResult = make(chan *pb.StressTestResult, 1)
		}()
		ctl.stressTestParam = common.StressTestParam{
			"node": map[string][]int64{
				"rank": {0},
			},
		}
		event, respCode, _ := ctl.handleWaitStressTestFinish()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.ServerInnerError)
	})
}

func testStressTestFinishValidReport(ctl *EventController) {
	convey.Convey("When receiving a valid report", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndResultChan",
			func() (context.Context, chan *pb.RecoverStatusRequest) {
				return ctx, make(chan *pb.RecoverStatusRequest, 1)
			})
		defer patches.Reset()
		defer func() {
			close(ctl.stressTestResponse)
			ctl.stressTestResponse = make(chan *pb.StressTestResponse, 1)
		}()
		resultChan := make(chan *pb.StressTestResult, 1)
		pat := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndStressTestResultChan",
			func() (context.Context, chan *pb.StressTestResult) {
				return context.Background(), resultChan
			}).
			ApplyFunc(faultmanager.FilterStressTestFault, func(jobID string, nodes []string, val bool) {}).
			ApplyPrivateMethod(ctl, "waitStressTestDone", func(ctx context.Context, rch chan *pb.StressTestResult,
				ch chan *pb.RecoverStatusRequest) (string, common.RespCode, error) {
				return "", common.OK, nil
			}).ApplyFunc(common.RetryWriteResetCM, func(taskName, nameSpace string, faultRankList []string, restartFaultProcess bool,
			operator string) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, nil
		})
		defer pat.Reset()
		defer func() {
			close(ctl.stressTestResult)
			ctl.stressTestResult = make(chan *pb.StressTestResult, 1)
		}()
		ctl.stressTestParam = common.StressTestParam{
			"node": map[string][]int64{
				"rank": {0},
			},
		}
		event, respCode, err := ctl.handleWaitStressTestFinish()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestWaitStressTestDone(t *testing.T) {
	convey.Convey("Test waitStressTestDone", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {}).
			ApplyFunc(hwlog.RunLog.Warnf, func(format string, args ...interface{}) {}).
			ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})
		defer patches.Reset()

		testWaitStressTestDoneCancel(ctl)
		testWaitStressTestRecoverChan(ctl)
		testWaitStressTestResultrChanFalse(ctl)
		testWaitStressTestResultrChanTrue(ctl)
	})
}

func testWaitStressTestDoneCancel(ctl *EventController) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	resultChan := make(chan *pb.StressTestResult, 1)
	recoverChan := make(chan *pb.RecoverStatusRequest, 1)
	defer func() {
		close(resultChan)
		close(recoverChan)
	}()
	event, respCode, err := ctl.waitStressTestDone(ctx, resultChan, recoverChan)
	convey.So(event, convey.ShouldEqual, "")
	convey.So(respCode, convey.ShouldEqual, common.ControllerEventCancel)
	convey.So(err, convey.ShouldBeNil)
}

func testWaitStressTestRecoverChan(ctl *EventController) {
	resultChan := make(chan *pb.StressTestResult, 1)
	recoverChan := make(chan *pb.RecoverStatusRequest, 1)
	patches := gomonkey.ApplyPrivateMethod(ctl, "waitStressTestFinishRecvFault",
		func(ctx context.Context, rch chan *pb.StressTestResult) (string, common.RespCode, error) {
			return "", common.OK, nil
		})
	defer patches.Reset()
	defer func() {
		close(resultChan)
		close(recoverChan)
	}()
	recoverChan <- &pb.RecoverStatusRequest{
		Status: &pb.Status{
			Code: common.UnRecoverableRetryError,
		},
	}
	event, respCode, err := ctl.waitStressTestDone(context.Background(), resultChan, recoverChan)
	convey.So(event, convey.ShouldEqual, "")
	convey.So(respCode, convey.ShouldEqual, common.OK)
	convey.So(err, convey.ShouldBeNil)
}

func testWaitStressTestResultrChanFalse(ctl *EventController) {
	resultChan := make(chan *pb.StressTestResult, 1)
	recoverChan := make(chan *pb.RecoverStatusRequest, 1)
	patches := gomonkey.ApplyPrivateMethod(ctl, "parseStressTestResult",
		func(result *pb.StressTestResult) (bool, string) {
			return false, "failed"
		}).ApplyFunc(common.RetryWriteResetCM, func(taskName, nameSpace string, faultRankList []string, restartFaultProcess bool,
		operator string) (*v1.ConfigMap, error) {
		return &v1.ConfigMap{}, nil
	})
	defer patches.Reset()
	defer func() {
		close(resultChan)
		close(recoverChan)
	}()
	resultChan <- &pb.StressTestResult{}
	event, _, err := ctl.waitStressTestDone(context.Background(), resultChan, recoverChan)
	convey.So(event, convey.ShouldEqual, common.StressTestFailEvent)
	convey.So(err, convey.ShouldBeNil)
}

func testWaitStressTestResultrChanTrue(ctl *EventController) {
	resultChan := make(chan *pb.StressTestResult, 1)
	recoverChan := make(chan *pb.RecoverStatusRequest, 1)
	patches := gomonkey.ApplyPrivateMethod(ctl, "parseStressTestResult",
		func(result *pb.StressTestResult) (bool, string) {
			return true, ""
		}).ApplyFunc(common.RetryWriteResetCM, func(taskName, nameSpace string, faultRankList []string, restartFaultProcess bool,
		operator string) (*v1.ConfigMap, error) {
		return &v1.ConfigMap{}, nil
	})
	defer patches.Reset()
	defer func() {
		close(resultChan)
		close(recoverChan)
	}()
	resultChan <- &pb.StressTestResult{}
	event, _, err := ctl.waitStressTestDone(context.Background(), resultChan, recoverChan)
	convey.So(event, convey.ShouldEqual, common.ReceiveReportEvent)
	convey.So(err, convey.ShouldBeNil)
}

func TestWaitStressTestFinishRecvFault(t *testing.T) {
	convey.Convey("Test waitStressTestFinishRecvFault", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		testWaitStressTestFinishRecvFaultCancel(ctl)
		testWaitStressTestFinishRecvFault(ctl)
	})
}

func testWaitStressTestFinishRecvFaultCancel(ctl *EventController) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	resultChan := make(chan *pb.StressTestResult, 1)
	defer func() {
		close(resultChan)
	}()
	event, respCode, err := ctl.waitStressTestFinishRecvFault(ctx, resultChan)
	convey.So(event, convey.ShouldEqual, "")
	convey.So(respCode, convey.ShouldEqual, common.ControllerEventCancel)
	convey.So(err, convey.ShouldBeNil)
}

func testWaitStressTestFinishRecvFault(ctl *EventController) {
	resultChan := make(chan *pb.StressTestResult, 1)
	patches := gomonkey.ApplyPrivateMethod(ctl, "parseStressTestResult",
		func(result *pb.StressTestResult) (bool, string) {
			return false, "failed"
		})
	defer patches.Reset()
	defer func() {
		close(resultChan)
	}()
	resultChan <- &pb.StressTestResult{}
	event, _, err := ctl.waitStressTestFinishRecvFault(context.Background(), resultChan)
	convey.So(event, convey.ShouldEqual, common.StressTestFailEvent)
	convey.So(err, convey.ShouldBeNil)
}

func TestParseStressTestResult(t *testing.T) {
	jobInfo := newJobInfoWithStrategy(nil)
	serviceCtx := context.Background()
	ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
	patches := gomonkey.ApplyFunc(job.GetJobCache, func(jobKey string) (constant.JobInfo, bool) {
		return constant.JobInfo{
			JobRankTable: constant.RankTable{
				ServerList: []constant.ServerHccl{
					{
						ServerName: "node",
						DeviceList: []constant.Device{
							{
								DeviceID: "1",
								RankID:   "1",
							},
						},
					},
				},
			},
		}, true
	}).ApplyPrivateMethod(ctl, "saveCacheFault", func(faults []*pb.FaultRank) {})
	defer patches.Reset()
	convey.Convey("Test parseStressTestResult, has fault", t, func() {
		result := &pb.StressTestResult{
			StressResult: map[string]*pb.StressTestRankResult{
				"1": {
					RankResult: map[string]*pb.StressTestOpResult{
						"1": {
							Code:   "2",
							Result: "fault",
						},
					},
				},
			},
		}
		ok, _ := ctl.parseStressTestResult(result)
		convey.ShouldBeFalse(ok)
	})
	convey.Convey("Test parseStressTestResult, no fault", t, func() {
		result := &pb.StressTestResult{
			StressResult: map[string]*pb.StressTestRankResult{
				"1": {
					RankResult: map[string]*pb.StressTestOpResult{
						"1": {
							Code:   "0",
							Result: "ok",
						},
					},
				},
			},
		}
		ok, _ := ctl.parseStressTestResult(result)
		convey.ShouldBeTrue(ok)
	})
}

func TestHandleStressTestFail(t *testing.T) {
	convey.Convey("Test handleStressTestFail", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		patches := gomonkey.ApplyFunc(job.GetJobCache, func(jobKey string) (constant.JobInfo, bool) {
			return constant.JobInfo{
				JobRankTable: constant.RankTable{
					ServerList: []constant.ServerHccl{
						{
							ServerName: "node",
							DeviceList: []constant.Device{
								{
									DeviceID: "1",
									RankID:   "1",
								},
							},
						},
					},
				},
			}, true
		}).ApplyPrivateMethod(ctl, "signalEnqueue", func(signal *pb.ProcessManageSignal) (string, common.RespCode, error) {
			return "", common.OK, nil
		}).ApplyPrivateMethod(ctl, "saveCacheFault", func(faults []*pb.FaultRank) {}).
			ApplyFunc(kube.RetryPatchNodeAnnotation, func(nodeName string, retryTimes int, annotations map[string]string) error {
				return nil
			})
		defer patches.Reset()
		ctl.isolateNodes.Insert("node")
		_, respCode, err := ctl.handleStressTestFail()
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}
