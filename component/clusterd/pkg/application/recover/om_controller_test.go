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

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/interface/grpc/recover"
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
		patchSendRetry := gomonkey.ApplyFunc(common.SwitchNicResponseSendRetry,
			func(stream pb.Recover_SubscribeSwitchNicSignalServer, signal *pb.SwitchNicResponse, retryTimes int) error {
				called = true
				return nil
			})
		defer patchSendRetry.Reset()
		ctl.selectSendSwitchNicResponseChan(ctx, sendChan, stream)
		assert.True(t, called)
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
	convey.Convey("When receiving a valid report", func() {
		reportChan := make(chan *pb.RecoverStatusRequest, 1)
		reportChan <- &pb.RecoverStatusRequest{Status: &pb.Status{Code: int32(common.OK)}}
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
		convey.So(event, convey.ShouldEqual, common.ReceiveReportEvent)
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
	convey.Convey("When receiving a valid report", func() {
		reportChan := make(chan *pb.StopCompleteRequest, 1)
		reportChan <- &pb.StopCompleteRequest{Status: &pb.Status{Code: int32(common.OK)}}
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
		convey.So(event, convey.ShouldEqual, common.ReceiveReportEvent)
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
		patchSendRetry := gomonkey.ApplyFunc(common.NotifySwitchNicSendRetry,
			func(stream pb.Recover_SubscribeNotifySwitchServer, signal *pb.SwitchRankList, retryTimes int) error {
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
