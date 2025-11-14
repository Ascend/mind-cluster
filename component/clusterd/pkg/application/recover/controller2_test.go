// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package recover a series of controller test function
package recover

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	apiv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/util"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	batchv1 "ascend-common/api/ascend-operator/apis/batch/v1"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/application/faultmanager/cmprocess/recoverinplace"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
	"clusterd/pkg/domain/statistics"
	"clusterd/pkg/interface/grpc/recover"
	"clusterd/pkg/interface/kube"
)

func TestHandleWaitRestartAllProcessTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(numInt2)*time.Minute)
	defer cancel()

	ctl := &EventController{
		jobInfo: common.JobBaseInfo{
			RecoverConfig: common.RecoverConfig{
				PlatFormMode: false,
			},
		},
		uuid: "testUuid",
	}

	patchGetCtx := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl), "getCtxAndScheduleResultChan",
		func(_ *EventController) (context.Context, <-chan struct{}) {
			return ctx, nil
		})
	defer patchGetCtx.Reset()

	patchTimeAfter := gomonkey.ApplyFunc(time.After, func(d time.Duration) <-chan time.Time {
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	})
	defer patchTimeAfter.Reset()

	event, code, err := ctl.handleWaitRestartAllProcess()
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if code != common.OK {
		t.Errorf("Expected response code %d, but got %d", common.OK, code)
	}
	if event != common.RestartProcessFinishEvent {
		t.Errorf("Expected event %s, but got %s", common.RestartProcessFinishEvent, event)
	}
}

func TestHandleWaitRestartAllProcessCtxDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ctl := &EventController{
		jobInfo: common.JobBaseInfo{
			RecoverConfig: common.RecoverConfig{
				PlatFormMode: false,
			},
		},
		uuid: "testUuid",
	}

	patchGetCtx := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl), "getCtxAndScheduleResultChan",
		func(_ *EventController) (context.Context, <-chan struct{}) {
			return ctx, nil
		})
	defer patchGetCtx.Reset()

	patchTimeSleep := gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {})
	defer patchTimeSleep.Reset()

	cancel()

	event, code, err := ctl.handleWaitRestartAllProcess()
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if code != common.ControllerEventCancel {
		t.Errorf("Expected response code %d, but got %d", common.ControllerEventCancel, code)
	}
	if event != "" {
		t.Errorf("Expected empty event, but got %s", event)
	}
}

func TestSelectSendChannelSendChanNil(t *testing.T) {
	convey.Convey("Test selectSendChannel when sendChan is nil", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &sender{}
		result := ctl.selectSendChannel(context.Background(), nil, stream)
		convey.So(result, convey.ShouldBeTrue)
	})
}

func TestSelectSendChannelContextDone(t *testing.T) {
	convey.Convey("Test selectSendChannel when context is done", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &sender{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		sendChan := make(chan *pb.ProcessManageSignal)
		result := ctl.selectSendChannel(ctx, sendChan, stream)
		convey.So(result, convey.ShouldBeTrue)
	})
}

func TestSelectSendChannelReceiveNonKeepAliveSignal(t *testing.T) {
	convey.Convey("Test selectSendChannel when receive non-keepalive signal from sendChan", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &sender{}
		ctx := context.Background()
		sendChan := make(chan *pb.ProcessManageSignal, 1)
		signal := &pb.ProcessManageSignal{SignalType: constant.ChangeStrategySignalType}
		sendChan <- signal

		patchSendRetry := gomonkey.ApplyFunc(common.SendRetry,
			func(sender common.SignalRetrySender, signal *pb.ProcessManageSignal, retryTimes int) error {
				return nil
			})
		defer patchSendRetry.Reset()

		patchHandleSendResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"handleSendResult",
			func(ctl *EventController, signal *pb.ProcessManageSignal, err error) {})
		defer patchHandleSendResult.Reset()

		result := ctl.selectSendChannel(ctx, sendChan, stream)
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestSelectSendChannelReceiveKeepAliveSignal(t *testing.T) {
	convey.Convey("Test selectSendChannel when receive keepalive signal from sendChan", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &sender{}
		ctx := context.Background()
		sendChan := make(chan *pb.ProcessManageSignal, 1)
		signal := &pb.ProcessManageSignal{SignalType: constant.KeepAliveSignalType}
		sendChan <- signal

		patchSendRetry := gomonkey.ApplyFunc(common.SendRetry,
			func(sender common.SignalRetrySender, signal *pb.ProcessManageSignal, retryTimes int) error {
				return nil
			})
		defer patchSendRetry.Reset()

		result := ctl.selectSendChannel(ctx, sendChan, stream)
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestSelectSendChannelSendChanClosed(t *testing.T) {
	convey.Convey("Test selectSendChannel when sendChan is closed", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		stream := &sender{}
		ctx := context.Background()
		sendChan := make(chan *pb.ProcessManageSignal)
		close(sendChan)
		result := ctl.selectSendChannel(ctx, sendChan, stream)
		convey.So(result, convey.ShouldBeTrue)
	})
}

func TestEventControllerChooseForRetryFail(t *testing.T) {
	convey.Convey("Test chooseForRetryFail", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
			agentReportStrategies: []string{},
		}

		patchRecover := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"supportRecoverStrategy", func(*EventController) bool {
				return false
			})
		defer patchRecover.Reset()
		patchDump := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"supportDumpStrategy", func(*EventController) bool {
				return false
			})
		defer patchDump.Reset()

		result := ctl.chooseForRetryFail()
		convey.So(result, convey.ShouldEqual, constant.ProcessExitStrategyName)
	})
}

func TestEventControllerChooseForRecoverFail(t *testing.T) {
	convey.Convey("Test chooseForRecoverFail", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
			agentReportStrategies: []string{},
		}

		patchDump := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"supportDumpStrategy", func(*EventController) bool {
				return false
			})
		defer patchDump.Reset()

		result := ctl.chooseForRecoverFail()
		convey.So(result, convey.ShouldEqual, constant.ProcessExitStrategyName)
	})
}

func TestEventControllerAgentSupportStrategy(t *testing.T) {
	convey.Convey("Test agentSupportStrategy", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
			agentReportStrategies: []string{constant.ProcessRetryStrategyName},
		}

		result := ctl.agentSupportStrategy(constant.ProcessRetryStrategyName)
		convey.So(result, convey.ShouldBeTrue)

		result = ctl.agentSupportStrategy("NonExistentStrategy")
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestEventControllerExtractRecoverResultNoStrategy(t *testing.T) {
	convey.Convey("Test extractRecoverResult when no strategy is decided", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchGetStrategyResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"getStrategyResult", func(*EventController) ([]string, []*pb.RecoverStatusRequest) {
				return nil, nil
			})
		defer patchGetStrategyResult.Reset()

		result, err := ctl.extractRecoverResult()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(result.Code, convey.ShouldEqual, common.ServerInnerError)
	})
}

func TestEventControllerExtractRecoverResultExitStrategy(t *testing.T) {
	convey.Convey("Test extractRecoverResult with ExitStrategy", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchGetStrategyResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"getStrategyResult", func(*EventController) ([]string, []*pb.RecoverStatusRequest) {
				return []string{constant.ProcessExitStrategyName}, nil
			})
		defer patchGetStrategyResult.Reset()

		result, err := ctl.extractRecoverResult()
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.Strategy, convey.ShouldEqual, constant.ProcessExitStrategyName)
		convey.So(result.Code, convey.ShouldEqual, common.OK)
		convey.So(result.RecoverSuccess, convey.ShouldBeTrue)
	})
}

func TestEventControllerExtractRecoverResultNormalStrategy(t *testing.T) {
	convey.Convey("Test extractRecoverResult with normal strategy", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchGetStrategyResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"getStrategyResult", func(*EventController) ([]string, []*pb.RecoverStatusRequest) {
				return []string{constant.ProcessRecoverStrategyName}, []*pb.RecoverStatusRequest{
					{
						JobId: "testJobId",
						Status: &pb.Status{
							Code: int32(common.OK),
							Info: "",
						},
						Strategy: constant.ProcessRecoverStrategyName,
					},
				}
			})
		defer patchGetStrategyResult.Reset()

		result, err := ctl.extractRecoverResult()
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.Strategy, convey.ShouldEqual, constant.ProcessRecoverStrategyName)
		convey.So(result.Code, convey.ShouldEqual, common.RespCode(common.OK))
		convey.So(result.RecoverSuccess, convey.ShouldBeTrue)
	})
}

func TestEventControllerRemoveAgentStrategy(t *testing.T) {
	convey.Convey("Test removeAgentStrategy", t, func() {
		ctl := &EventController{
			agentReportStrategies: []string{constant.ProcessRecoverStrategy,
				constant.ProcessRetryStrategyName},
		}
		ctl.removeAgentStrategy(constant.ProcessRetryStrategyName)
		convey.So(len(ctl.agentReportStrategies), convey.ShouldEqual, 1)
		convey.So(ctl.agentReportStrategies[0], convey.ShouldEqual, constant.ProcessRecoverStrategy)
	})
}

func TestEventControllerUpdateFixResultSuccess(t *testing.T) {
	convey.Convey("Test updateFixResult success", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				PgName:    "testPgName",
				Namespace: "testNamespace",
			},
		}
		result := make(map[string]interface{})
		patchRetryPatchPodGroupAnnotations := gomonkey.ApplyFunc(kube.RetryPatchPodGroupAnnotations,
			func(pgName, namespace string, retryTimes int, annotations map[string]interface{}) (*v1beta1.PodGroup, error) {
				for k, v := range annotations {
					result[k] = v
				}
				return nil, nil
			})
		defer patchRetryPatchPodGroupAnnotations.Reset()

		ctl.updateFixResult(constant.ProcessRetryStrategyName, constant.RetrySuccess)
		convey.So(len(result), convey.ShouldEqual, 1)
	})
}

func TestEventControllerUpdateFixResultFailure(t *testing.T) {
	convey.Convey("Test updateFixResult failure", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				PgName:    "testPgName",
				Namespace: "testNamespace",
			},
		}
		result := make(map[string]string)
		patchRetryPatchPodGroupAnnotations := gomonkey.ApplyFunc(kube.RetryPatchPodGroupAnnotations,
			func(pgName, namespace string, retryTimes int, annotations map[string]interface{}) (*v1beta1.PodGroup, error) {
				return nil, errors.New("patch error")
			})
		defer patchRetryPatchPodGroupAnnotations.Reset()
		ctl.updateFixResult(constant.ProcessRetryStrategyName, constant.RetrySuccess)
		convey.So(len(result), convey.ShouldEqual, 0)
	})
}

func TestEventControllerHandleCheckRecoverResultRetrySuccess(t *testing.T) {
	convey.Convey("Test handleCheckRecoverResult with RetryStrategy success", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		patchExtractRecoverResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"extractRecoverResult", func(*EventController) (common.RecoverResult, error) {
				return common.RecoverResult{
					Strategy:       constant.ProcessRetryStrategyName,
					Code:           common.OK,
					RecoverSuccess: true,
				}, nil
			})
		defer patchExtractRecoverResult.Reset()

		patchUpdateFixResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"updateFixResult", func(*EventController, string, string) {})
		defer patchUpdateFixResult.Reset()

		event, code, err := ctl.handleCheckRecoverResult()
		convey.So(err, convey.ShouldBeNil)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(event, convey.ShouldEqual, common.RecoverSuccessEvent)
	})
}

func TestEventControllerHandleCheckRecoverResultRetryFailedRecoverable(t *testing.T) {
	convey.Convey("Test handleCheckRecoverResult with RetryStrategy failed and recoverable", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchExtractRecoverResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"extractRecoverResult", func(*EventController) (common.RecoverResult, error) {
				return common.RecoverResult{
					Strategy:       constant.ProcessRetryStrategyName,
					Code:           common.RecoverableRetryError,
					RecoverSuccess: false,
				}, nil
			})
		defer patchExtractRecoverResult.Reset()

		patchUpdateFixResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"updateFixResult", func(*EventController, string, string) {})
		defer patchUpdateFixResult.Reset()

		event, code, err := ctl.handleCheckRecoverResult()
		convey.So(err, convey.ShouldBeNil)
		convey.So(code, convey.ShouldEqual, common.RecoverableRetryError)
		convey.So(event, convey.ShouldEqual, common.RecoverableRetryErrorEvent)
	})
}

func TestEventControllerHandleCheckRecoverResultRetryFailedUnrecoverable(t *testing.T) {
	convey.Convey("Test handleCheckRecoverResult with RetryStrategy failed and unrecoverable", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchExtractRecoverResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"extractRecoverResult", func(*EventController) (common.RecoverResult, error) {
				return common.RecoverResult{
					Strategy:       constant.ProcessRetryStrategyName,
					Code:           common.ClientError,
					RecoverSuccess: false,
				}, nil
			})
		defer patchExtractRecoverResult.Reset()

		patchUpdateFixResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"updateFixResult", func(*EventController, string, string) {})
		defer patchUpdateFixResult.Reset()

		patchRemoveAgentStrategyRecover := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"removeAgentStrategy", func(*EventController, string) {})
		defer patchRemoveAgentStrategyRecover.Reset()

		patchRemoveAgentStrategyDump := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"removeAgentStrategy", func(*EventController, string) {})
		defer patchRemoveAgentStrategyDump.Reset()

		event, code, err := ctl.handleCheckRecoverResult()
		convey.So(err, convey.ShouldBeNil)
		convey.So(code, convey.ShouldEqual, common.UnRecoverableRetryError)
		convey.So(event, convey.ShouldEqual, common.UnRecoverableRetryErrorEvent)
	})
}

func TestEventControllerHandleCheckRecoverResultRecoverSuccess(t *testing.T) {
	convey.Convey("Test handleCheckRecoverResult with RecoverStrategy success", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchExtractRecoverResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"extractRecoverResult", func(*EventController) (common.RecoverResult, error) {
				return common.RecoverResult{
					Strategy:       constant.ProcessRecoverStrategyName,
					Code:           common.OK,
					RecoverSuccess: true,
				}, nil
			})
		defer patchExtractRecoverResult.Reset()

		patchUpdateFixResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"updateFixResult", func(*EventController, string, string) {})
		defer patchUpdateFixResult.Reset()

		event, code, err := ctl.handleCheckRecoverResult()
		convey.So(err, convey.ShouldBeNil)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(event, convey.ShouldEqual, common.RecoverSuccessEvent)
	})
}

func TestEventControllerHandleCheckRecoverResultRecoverFailed(t *testing.T) {
	convey.Convey("Test handleCheckRecoverResult with RecoverStrategy failed", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchExtractRecoverResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"extractRecoverResult", func(*EventController) (common.RecoverResult, error) {
				return common.RecoverResult{
					Strategy:       constant.ProcessRecoverStrategyName,
					Code:           common.ClientError,
					RecoverSuccess: false,
				}, nil
			})
		defer patchExtractRecoverResult.Reset()

		patchUpdateFixResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"updateFixResult", func(*EventController, string, string) {})
		defer patchUpdateFixResult.Reset()

		event, code, err := ctl.handleCheckRecoverResult()
		convey.So(err, convey.ShouldBeNil)
		convey.So(code, convey.ShouldEqual, common.ClientError)
		convey.So(event, convey.ShouldEqual, common.RecoverFailEvent)
	})
}

func TestEventControllerHandleCheckRecoverResultDumpOrExitStrategySuccess(t *testing.T) {
	convey.Convey("Test handleCheckRecoverResult with Dump or Exit Strategy success", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchExtractRecoverResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"extractRecoverResult", func(*EventController) (common.RecoverResult, error) {
				return common.RecoverResult{
					Strategy:       constant.ProcessDumpStrategyName,
					Code:           common.OK,
					RecoverSuccess: true,
				}, nil
			})
		defer patchExtractRecoverResult.Reset()

		patchUpdateFixResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"updateFixResult", func(*EventController, string, string) {})
		defer patchUpdateFixResult.Reset()

		event, code, err := ctl.handleCheckRecoverResult()
		convey.So(err, convey.ShouldBeNil)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(event, convey.ShouldEqual, common.CheckResultFinishEvent)
	})
}

func TestEventControllerHandleCheckRecoverResultExitStrategySuccess(t *testing.T) {
	convey.Convey("Test handleCheckRecoverResult with Exit Strategy success", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchExtractRecoverResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"extractRecoverResult", func(*EventController) (common.RecoverResult, error) {
				return common.RecoverResult{
					Strategy:       constant.ProcessExitStrategyName,
					Code:           common.OK,
					RecoverSuccess: true,
				}, nil
			})
		defer patchExtractRecoverResult.Reset()

		patchUpdateFixResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"updateFixResult", func(*EventController, string, string) {})
		defer patchUpdateFixResult.Reset()

		event, code, err := ctl.handleCheckRecoverResult()
		convey.So(err, convey.ShouldBeNil)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(event, convey.ShouldEqual, common.CheckResultFinishEvent)
	})
}

func TestEventControllerHandleCheckRecoverResultDumpStrategyFailed(t *testing.T) {
	convey.Convey("Test handleCheckRecoverResult with Dump Strategy failed", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchExtractRecoverResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"extractRecoverResult", func(*EventController) (common.RecoverResult, error) {
				return common.RecoverResult{
					Strategy:       constant.ProcessDumpStrategyName,
					Code:           common.OK,
					RecoverSuccess: false,
				}, nil
			})
		defer patchExtractRecoverResult.Reset()

		patchUpdateFixResult := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"updateFixResult", func(*EventController, string, string) {})
		defer patchUpdateFixResult.Reset()

		event, code, err := ctl.handleCheckRecoverResult()
		convey.So(err, convey.ShouldBeNil)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(event, convey.ShouldEqual, common.CheckResultFinishEvent)
	})
}

func TestEventControllerHandleKillPodJobNotExist(t *testing.T) {
	convey.Convey("Test handleKillPod when job does not exist", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchGetJobIsExists := gomonkey.ApplyFunc(job.GetJobIsExists, func(jobId string) bool {
			return false
		})
		defer patchGetJobIsExists.Reset()

		_, code, err := ctl.handleKillPod()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(code, convey.ShouldEqual, common.JobNotExist)
	})
}

func TestEventControllerHandleKillPodWriteCMError(t *testing.T) {
	convey.Convey("Test handleKillPod when write CM fails", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchGetJobIsExists := gomonkey.ApplyFunc(job.GetJobIsExists, func(jobId string) bool {
			return true
		})
		defer patchGetJobIsExists.Reset()

		patchUpdateCacheFaultAndPod := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"updateCacheFaultAndPod", func(*EventController) ([]*pb.FaultRank, []string, error) {
				return nil, nil, nil
			})
		defer patchUpdateCacheFaultAndPod.Reset()

		patchRetryWriteResetCM := gomonkey.ApplyFunc(common.RetryWriteResetCM,
			func(jobName, namespace string, allFaultRanks []string,
				restartProcess bool, operation string) (*v1.ConfigMap, error) {
				return nil, errors.New("write CM error")
			})
		defer patchRetryWriteResetCM.Reset()

		_, code, err := ctl.handleKillPod()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(code, convey.ShouldEqual, common.OperateConfigMapError)
	})
}

func TestEventControllerHandleKillPodSuccess(t *testing.T) {
	convey.Convey("Test handleKillPod success", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchGetJobIsExists := gomonkey.ApplyFunc(job.GetJobIsExists, func(jobId string) bool {
			return true
		})
		defer patchGetJobIsExists.Reset()

		patchUpdateCacheFaultAndPod := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl),
			"updateCacheFaultAndPod", func(*EventController) ([]*pb.FaultRank, []string, error) {
				return nil, nil, nil
			})
		defer patchUpdateCacheFaultAndPod.Reset()

		patchRetryWriteResetCM := gomonkey.ApplyFunc(common.RetryWriteResetCM,
			func(jobName, namespace string, allFaultRanks []string,
				restartProcess bool, operation string) (*v1.ConfigMap, error) {
				return &v1.ConfigMap{Data: map[string]string{
					constant.ResetInfoCMDataKey: "test data",
				}}, nil
			})
		defer patchRetryWriteResetCM.Reset()

		event, code, err := ctl.handleKillPod()
		convey.So(err, convey.ShouldBeNil)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(event, convey.ShouldEqual, common.FinishKillPodEvent)
	})
}

func TestEventControllerHandleFaultRetryChangePauseError(t *testing.T) {
	convey.Convey("Test handleFaultRetry when change pause mode fails", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchChangeProcessRecoverEnableMode := gomonkey.ApplyFunc(common.ChangeProcessRecoverEnableMode,
			func(jobInfo common.JobBaseInfo, mode string) (*v1beta1.PodGroup, error) {
				return nil, errors.New("change pause mode error")
			})
		defer patchChangeProcessRecoverEnableMode.Reset()

		event, code, err := ctl.handleFaultRetry()
		convey.So(err, convey.ShouldBeNil)
		convey.So(code, convey.ShouldEqual, common.OperatePodGroupError)
		convey.So(event, convey.ShouldEqual, common.ChangeProcessSchedulingModePauseErrorEvent)
	})
}

func TestEventControllerHandleFaultRetryChangeEnableError(t *testing.T) {
	convey.Convey("Test handleFaultRetry when change enable mode fails", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}
		patchChangeProcessRecoverEnableMode := gomonkey.ApplyFunc(common.ChangeProcessRecoverEnableMode,
			func(jobInfo common.JobBaseInfo, mode string) (*v1beta1.PodGroup, error) {
				if mode == constant.ProcessRecoverEnable {
					return nil, errors.New("mock error")
				}
				return nil, nil
			})
		defer patchChangeProcessRecoverEnableMode.Reset()

		patchGetJobIsRunning := gomonkey.ApplyFunc(job.GetJobIsRunning, func(jobId string) bool {
			return true
		})
		defer patchGetJobIsRunning.Reset()

		event, code, err := ctl.handleFaultRetry()
		convey.So(err, convey.ShouldBeNil)
		convey.So(code, convey.ShouldEqual, common.OperatePodGroupError)
		convey.So(event, convey.ShouldEqual, common.ChangeProcessSchedulingModeEnableErrorEvent)
	})
}

func TestEventControllerHandleFaultRetrySuccess(t *testing.T) {
	convey.Convey("Test handleFaultRetry success", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "testJobId",
			},
		}

		patchChangeProcessRecoverEnableMode := gomonkey.ApplyFunc(common.ChangeProcessRecoverEnableMode,
			func(jobInfo common.JobBaseInfo, mode string) (*v1beta1.PodGroup, error) {
				return nil, nil
			})
		defer patchChangeProcessRecoverEnableMode.Reset()

		patchGetJobIsRunning := gomonkey.ApplyFunc(job.GetJobIsRunning, func(jobId string) bool {
			return true
		})
		defer patchGetJobIsRunning.Reset()

		event, code, err := ctl.handleFaultRetry()
		convey.So(err, convey.ShouldBeNil)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(event, convey.ShouldEqual, common.FinishEvent)
	})
}

func TestPGStatusEnqueue(t *testing.T) {
	ctl := &EventController{
		jobInfo: common.JobBaseInfo{JobId: "testJobId"},
		uuid:    "testUuid",
	}
	convey.Convey("Test pgStatusEnqueue method", t, func() {
		convey.Convey("Test the case where the channel is nil", func() {
			patcher := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndScheduleResultChan",
				func() (context.Context, chan bool) {
					return context.Background(), nil
				})
			defer patcher.Reset()
			ctl.pgStatusEnqueue(true)
			convey.So(func() { ctl.pgStatusEnqueue(true) }, convey.ShouldNotPanic)
		})
		convey.Convey("Test the case of normal enqueue", func() {
			ch := make(chan bool, 1)
			patcher := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndScheduleResultChan",
				func() (context.Context, chan bool) {
					return context.Background(), ch
				})
			defer patcher.Reset()
			ctl.pgStatusEnqueue(true)
			result := <-ch
			convey.So(result, convey.ShouldBeTrue)
		})
		convey.Convey("Test the case where the context is canceled", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			patcher := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndScheduleResultChan",
				func() (context.Context, chan bool) {
					return ctx, make(chan bool)
				})
			defer patcher.Reset()
			ctl.pgStatusEnqueue(true)
			convey.So(func() { ctl.pgStatusEnqueue(true) }, convey.ShouldNotPanic)
		})
		convey.Convey("Test the case of enqueue timeout", func() {
			patcher := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndScheduleResultChan",
				func() (context.Context, chan bool) {
					return context.Background(), make(chan bool)
				})
			defer patcher.Reset()
			ctl.pgStatusEnqueue(true)
			convey.So(func() { ctl.pgStatusEnqueue(true) }, convey.ShouldNotPanic)
		})
	})
}

func TestSelectEventChanEventChanNil(t *testing.T) {
	ctl := &EventController{
		jobInfo: common.JobBaseInfo{
			JobId: "testJobId",
		},
		uuid: "testUuid",
	}
	convey.Convey("Test the case where the eventChan is nil", t, func() {
		result := ctl.selectEventChan(context.Background(), nil)
		convey.So(result, convey.ShouldBeTrue)
	})
}

func TestSelectEventChanContextCanceled(t *testing.T) {
	ctl := &EventController{
		jobInfo: common.JobBaseInfo{
			JobId: "testJobId",
		},
		uuid: "testUuid",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	eventChan := make(chan string)
	convey.Convey("Test the case where the context is canceled", t, func() {
		result := ctl.selectEventChan(ctx, eventChan)
		convey.So(result, convey.ShouldBeTrue)
	})
}

func TestSelectEventChanTriggerReturnsError(t *testing.T) {
	jobInfo := newJobInfoWithStrategy(nil)
	serviceCtx := context.Background()
	ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
	eventChan := make(chan string, 1)
	eventChan <- "testEvent"
	convey.Convey("Test the case where the eventChan receives an event and trigger returns an error", t, func() {
		resetPatcher := gomonkey.ApplyPrivateMethod(ctl, "reset", func(_ *EventController) {})
		defer resetPatcher.Reset()
		result := ctl.selectEventChan(context.Background(), eventChan)
		convey.So(result, convey.ShouldBeTrue)
	})
}

func TestSelectEventChanTriggerReturnsNoErrorAndNextEventNotEmpty(t *testing.T) {
	jobInfo := newJobInfoWithStrategy(nil)
	serviceCtx := context.Background()
	ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
	eventChan := make(chan string, 1)
	eventChan <- common.FaultOccurEvent

	patches := gomonkey.ApplyFuncReturn(common.WriteResetInfoToCM, nil, nil).
		ApplyPrivateMethod(ctl, "handleNotifyWaitFaultFlushing",
			func() (string, common.RespCode, error) {
				return "", common.OK, nil
			})
	defer patches.Reset()

	convey.Convey("Test the case where the eventChan receives an event and trigger "+
		"returns no error and nextEvent is not empty",
		t, func() {
			result := ctl.selectEventChan(context.Background(), eventChan)
			convey.So(result, convey.ShouldBeFalse)
		})
}

func TestSelectEventChanEventChanClosed(t *testing.T) {
	ctl := &EventController{
		jobInfo: common.JobBaseInfo{
			JobId: "testJobId",
		},
		uuid: "testUuid",
	}
	eventChan := make(chan string)
	close(eventChan)
	convey.Convey("Test the case where the eventChan is closed", t, func() {
		result := ctl.selectEventChan(context.Background(), eventChan)
		convey.So(result, convey.ShouldBeTrue)
	})
}

func TestWaitHCCLRoutingConvergence(t *testing.T) {
	convey.Convey("Testing WaitHCCLRoutingConvergence ok", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		res := ctl.waitHCCLRoutingConvergence()
		convey.So(res, convey.ShouldEqual, true)
	})
}

func TestShouldWaitHcclRoutingConvergence(t *testing.T) {
	convey.Convey("Testing IsWaitHcclRoutingConvergence is true", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.cacheRetryFault = []*pb.FaultRank{
			{FaultType: constant.HcclFaultType},
		}
		res := ctl.shouldWaitHcclRoutingConvergence()
		convey.So(res, convey.ShouldEqual, true)
	})
}

func TestHasSameRetryFault(t *testing.T) {
	convey.Convey("Testing hasSameRetryFault is true", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.cacheRetryFault = []*pb.FaultRank{
			{FaultType: constant.HcclFaultType},
			{FaultType: constant.HcclFaultType},
		}
		res := ctl.hasSameRetryFault()
		convey.So(res, convey.ShouldEqual, true)
	})

	convey.Convey("Testing hasSameRetryFault is false", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.cacheRetryFault = []*pb.FaultRank{
			{FaultType: constant.HcclFaultType},
			{FaultType: constant.UceFaultType},
		}
		res := ctl.hasSameRetryFault()
		convey.So(res, convey.ShouldEqual, false)
	})
}

func TestNotifyHCCLRoutingTimeout(t *testing.T) {
	convey.Convey("Testing notifyHCCLRoutingTimeout ", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		signal := &pb.ProcessManageSignal{
			FaultRanks: []*pb.FaultRank{
				{FaultType: constant.HcclFaultType},
			},
		}
		res := ctl.notifyHCCLRoutingTimeout(signal)
		convey.So(res.Timeout, convey.ShouldEqual, constant.HCCLRoutingConvergenceTimeout+constant.StepRetryTimeout)
	})
}

func TestHandleRestartFaultProcess(t *testing.T) {
	convey.Convey("Testing handleRestartFaultProcess", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		mockGetNodeRankIdsByFaultRanks := gomonkey.ApplyFuncReturn(common.GetNodeRankIdsByRankIds, []string{}, nil)
		defer mockGetNodeRankIdsByFaultRanks.Reset()
		patches := gomonkey.ApplyFunc(ctl.signalEnqueue,
			func(signal *pb.ProcessManageSignal) (string, common.RespCode, error) {
				return "", common.OK, nil
			}).ApplyFuncReturn(common.WriteResetInfoToCM, &v1.ConfigMap{}, nil).
			ApplyPrivateMethod(ctl, "updateCacheFaultAndPod",
				func() ([]*pb.FaultRank, []string, error) {
					return []*pb.FaultRank{{RankId: "rank1"}}, []string{"rank1"}, nil
				}).ApplyFuncReturn(faultmanager.CallbackForReportNoRetryInfo)
		defer patches.Reset()
		convey.Convey("choose recover-in-place strategy and write json success, "+
			"should notify recover strategy", func() {
			signal := &pb.ProcessManageSignal{ChangeStrategy: constant.ProcessRecoverInPlaceStrategyName}
			_, code, _ := ctl.handleRestartFaultProcess(signal)
			convey.So(code, convey.ShouldEqual, common.OK)
			convey.So(signal.ChangeStrategy, convey.ShouldEqual, constant.ProcessRecoverStrategyName)
		})
		convey.Convey("choose retry strategy, should enqueue signal", func() {
			signal := &pb.ProcessManageSignal{ChangeStrategy: constant.ProcessRetryStrategyName}
			_, code, _ := ctl.handleRestartFaultProcess(signal)
			convey.So(code, convey.ShouldEqual, common.OK)
			convey.So(signal.ChangeStrategy, convey.ShouldEqual, constant.ProcessRetryStrategyName)
		})
		convey.Convey("choose other strategy, should kill pod", func() {
			ctl.restartFaultProcess = true
			signal := &pb.ProcessManageSignal{ChangeStrategy: constant.ProcessRecoverStrategyName}
			event, code, _ := ctl.handleRestartFaultProcess(signal)
			convey.So(event, convey.ShouldEqual, common.KillPodAfterRestartProcessEvent)
			convey.So(code, convey.ShouldEqual, common.ServerInnerError)
			convey.So(ctl.restartFaultProcess, convey.ShouldBeFalse)
		})
	})
}

func TestWaitNormalFaultRecovery(t *testing.T) {
	convey.Convey("Testing waitNormalFaultRecovery ", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		patches := gomonkey.ApplyFuncReturn(time.Sleep)
		defer patches.Reset()
		convey.Convey("job has no fault, should return nil", func() {
			err := ctl.waitNormalFaultRecovery()
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("job has fault, should return error", func() {
			patches.ApplyPrivateMethod(recoverinplace.RecoverInplaceProcessor, "JobHasFault",
				func(string) bool { return true })
			err := ctl.waitNormalFaultRecovery()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestCatchException(t *testing.T) {
	convey.Convey("Testing catchException ", t, func() {
		recovered := false

		func() {
			defer catchException()
			defer func() { recovered = true }()
			panic("test error")
		}()

		if !recovered {
			t.Error("not catch exception")
		}
	})
}

func TestSupportTargetStrategy(t *testing.T) {
	convey.Convey("MindXConfigStrategies and agentReportStrategies contain recover strategy", t, func() {
		jobInfo := newJobInfoWithStrategy([]string{
			constant.ProcessRecoverStrategyName,
			constant.ProcessDumpStrategyName})
		ctl := &EventController{
			jobInfo:               jobInfo,
			agentReportStrategies: []string{constant.ProcessRecoverStrategyName},
		}
		hasRecover := ctl.supportTargetStrategy(constant.ProcessRecoverStrategyName)
		convey.So(hasRecover, convey.ShouldBeTrue)

		ctl.platStrategy = constant.ProcessRecoverStrategyName
		hasRecover = ctl.supportTargetStrategy(constant.ProcessRetryStrategyName)
		convey.So(hasRecover, convey.ShouldBeFalse)

		ctl.platStrategy = constant.ProcessRecoverStrategyName
		hasRecover = ctl.supportTargetStrategy(constant.ProcessDumpStrategyName)
		convey.So(hasRecover, convey.ShouldBeFalse)
	})
}

func TestEventController_supportTargetStrategy(t *testing.T) {
	type fields struct {
		jobInfo               common.JobBaseInfo
		agentReportStrategies []string
	}
	type args struct {
		recoverStrategy string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "case 1: not config",
			fields: fields{
				jobInfo: common.JobBaseInfo{
					RecoverConfig: common.RecoverConfig{
						MindXConfigStrategies: []string{"recover"}}}},
			args: args{recoverStrategy: "retry"},
			want: false},
		{
			name: "case 2: not report",
			fields: fields{
				agentReportStrategies: []string{"retry"},
				jobInfo: common.JobBaseInfo{
					RecoverConfig: common.RecoverConfig{
						MindXConfigStrategies: []string{"recover"}}}},
			args: args{recoverStrategy: "recover"},
			want: false},
		{
			name: "case 3: config and report",
			fields: fields{
				agentReportStrategies: []string{"recover"},
				jobInfo: common.JobBaseInfo{
					RecoverConfig: common.RecoverConfig{
						MindXConfigStrategies: []string{"recover"}}}},
			args: args{recoverStrategy: "recover"},
			want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := &EventController{
				jobInfo:               tt.fields.jobInfo,
				agentReportStrategies: tt.fields.agentReportStrategies}
			assert.Equalf(t, tt.want, ctl.supportTargetStrategy(tt.args.recoverStrategy), "supportTargetStrategy(%v)", tt.args.recoverStrategy)
		})
	}
}

func newTestEventController(jobID string) *EventController {
	return &EventController{
		jobInfo:  common.JobBaseInfo{JobId: jobID},
		faultPod: make(map[string]string),
	}
}

func TestHandleCheckScaleStrategyRecoverResultScaleInExitIsolateSuccess(t *testing.T) {
	ctl := newTestEventController("job-123")
	result := common.RecoverResult{
		Strategy:       constant.ScaleInStrategyName,
		RecoverSuccess: false,
		Code:           common.ExitIsolateRanksCode,
		IsolateRankIds: []string{"rank-1", "rank-2"},
	}
	ctl.signalChan = make(chan *pb.ProcessManageSignal, 1)
	patches := gomonkey.ApplyFunc(common.GetNodeRankIdsByRankIds,
		func(_ string, _ []string) ([]string, error) {
			return []string{"1", "2"}, nil
		})
	defer patches.Reset()

	event, code, err := ctl.handleCheckScaleStrategyRecoverResult(result)

	assert.Empty(t, event)
	assert.Equal(t, common.OK, code)
	assert.NoError(t, err)
}

func TestHandleCheckScaleStrategyRecoverResultScaleInGetNodeRankIdsFail(t *testing.T) {
	ctl := newTestEventController("job-123")
	result := common.RecoverResult{
		Strategy:       constant.ScaleInStrategyName,
		RecoverSuccess: false,
		Code:           common.ExitIsolateRanksCode,
		IsolateRankIds: []string{"rank-1", "rank-2"},
	}

	patches := gomonkey.ApplyFunc(common.GetNodeRankIdsByRankIds,
		func(_ string, _ []string) ([]string, error) {
			return nil, fmt.Errorf("test error")
		})
	defer patches.Reset()

	event, code, err := ctl.handleCheckScaleStrategyRecoverResult(result)

	assert.Equal(t, common.NotifyFailEvent, event)
	assert.Equal(t, common.ClientError, code)
	assert.NoError(t, err)
}

func TestHandleCheckScaleStrategyRecoverResultScaleInSignalEnqueueFail(t *testing.T) {
	ctl := newTestEventController("job-123")
	result := common.RecoverResult{
		Strategy:       constant.ScaleInStrategyName,
		RecoverSuccess: false,
		Code:           common.ExitIsolateRanksCode,
		IsolateRankIds: []string{"1", "2"},
	}

	patches := gomonkey.ApplyFunc(common.GetNodeRankIdsByRankIds,
		func(_ string, _ []string) ([]string, error) {
			return []string{"1", "2"}, nil
		})
	defer patches.Reset()

	event, code, err := ctl.handleCheckScaleStrategyRecoverResult(result)

	assert.Equal(t, common.NotifyFailEvent, event)
	assert.Equal(t, common.ClientError, code)
	assert.NoError(t, err)
}

func TestHandleCheckScaleStrategyRecoverResultScaleInRecoverSuccess(t *testing.T) {
	ctl := newTestEventController("job-123")
	result := common.RecoverResult{
		Strategy:       constant.ScaleInStrategyName,
		RecoverSuccess: true,
	}

	event, code, err := ctl.handleCheckScaleStrategyRecoverResult(result)

	assert.Equal(t, common.ScaleInSuccessEvent, event)
	assert.Equal(t, common.OK, code)
	assert.NoError(t, err)
}

func TestHandleCheckScaleStrategyRecoverResultScaleOutRecoverFail(t *testing.T) {
	ctl := newTestEventController("job-123")
	result := common.RecoverResult{
		Strategy:       constant.ScaleOutStrategyName,
		RecoverSuccess: false,
	}

	event, code, err := ctl.handleCheckScaleStrategyRecoverResult(result)

	assert.Equal(t, common.RecoverFailEvent, event)
	assert.Equal(t, common.ClientError, code)
	assert.NoError(t, err)
}

func TestHandleCheckScaleStrategyRecoverResultUnknownStrategy(t *testing.T) {
	ctl := newTestEventController("job-123")
	result := common.RecoverResult{
		Strategy: "unknown-strategy",
	}

	event, code, err := ctl.handleCheckScaleStrategyRecoverResult(result)

	assert.Empty(t, event)
	assert.Equal(t, common.ServerInnerError, code)
	assert.True(t, strings.Contains(err.Error(), "not support"))
}

// TestHandleWaitReportScaleInIsolateRanksStatusResultChanNil tests when resultCh is nil.
func TestHandleWaitReportScaleInIsolateRanksStatusResultChanNil(t *testing.T) {
	ctl := &EventController{jobInfo: common.JobBaseInfo{JobId: "job-123"}}
	event, code, err := ctl.handleWaitReportScaleInIsolateRanksStatus()
	assert.Empty(t, event)
	assert.Equal(t, common.ServerInnerError, code)
	assert.Contains(t, err.Error(), "resultCh is nil")
}

// TestHandleWaitReportScaleInIsolateRanksStatusResultChanClosed tests when resultCh is closed.
func TestHandleWaitReportScaleInIsolateRanksStatusResultChanClosed(t *testing.T) {
	ctl := &EventController{
		jobInfo:           common.JobBaseInfo{JobId: "job-123"},
		controllerContext: context.Background()}
	resultCh := make(chan *pb.RecoverStatusRequest)
	close(resultCh) // Close channel before test
	ctl.reportStatusChan = resultCh
	event, code, err := ctl.handleWaitReportScaleInIsolateRanksStatus()
	assert.Empty(t, event)
	assert.Equal(t, common.OK, code)
	assert.NoError(t, err)
}

// TestHandleWaitReportScaleInIsolateRanksStatusReceiveRequest tests normal request reception.
func TestHandleWaitReportScaleInIsolateRanksStatusReceiveRequest(t *testing.T) {
	ctl := &EventController{
		jobInfo:           common.JobBaseInfo{JobId: "job-123"},
		controllerContext: context.Background(),
		state:             &common.StateMachine{},
	}
	resultCh := make(chan *pb.RecoverStatusRequest, 1)
	req := &pb.RecoverStatusRequest{Strategy: constant.ScaleInStrategyName, Status: &pb.Status{Code: 0}}
	resultCh <- req // Send test request
	ctl.reportStatusChan = resultCh
	event, code, err := ctl.handleWaitReportScaleInIsolateRanksStatus()
	assert.Equal(t, common.ReceiveReportEvent, event)
	assert.Equal(t, common.OK, code)
	assert.NoError(t, err)
}

// TestHandleWaitReportScaleInIsolateRanksStatusCtxCanceled tests when context is canceled.
func TestHandleWaitReportScaleInIsolateRanksStatusCtxCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	ctl := &EventController{
		jobInfo:           common.JobBaseInfo{JobId: "job-123"},
		controllerContext: ctx,
		reportStatusChan:  make(chan *pb.RecoverStatusRequest, 1),
	}
	event, code, err := ctl.handleWaitReportScaleInIsolateRanksStatus()
	assert.Empty(t, event)
	assert.Equal(t, common.ControllerEventCancel, code)
	assert.NoError(t, err)
}

func TestEventControllerCanChooseScaleInStrategy(t *testing.T) {
	type fields struct {
		jobInfo               common.JobBaseInfo
		agentReportStrategies []string
		faultPod              map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "return true",
			fields: fields{
				faultPod: map[string]string{},
				jobInfo: common.JobBaseInfo{
					RecoverConfig: common.RecoverConfig{
						MindXConfigStrategies: []string{constant.ElasticTrainingStrategyName},
					},
				},
				agentReportStrategies: []string{constant.ScaleInStrategyName},
			},
			want: true,
		},
		{
			name: "not config return false",
			fields: fields{
				faultPod: map[string]string{},
			},
			want: false,
		},
		{
			name: "rank 0 return false",
			fields: fields{
				faultPod: map[string]string{"0": ""},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := &EventController{
				jobInfo:               tt.fields.jobInfo,
				faultPod:              tt.fields.faultPod,
				agentReportStrategies: tt.fields.agentReportStrategies,
			}
			assert.Equalf(t, tt.want, ctl.canChooseScaleInStrategy(), "canChooseScaleInStrategy()")
		})
	}
}

// TestHandleWaitReportScaleInIsolateRanksStatusTimeout tests when report timeout occurs.
func TestHandleWaitReportScaleInIsolateRanksStatusTimeout(t *testing.T) {
	ctl := &EventController{
		jobInfo:           common.JobBaseInfo{JobId: "job-123"},
		controllerContext: context.Background(),
		reportStatusChan:  make(chan *pb.RecoverStatusRequest, 1),
		state:             &common.StateMachine{},
	}
	chanTime := make(chan time.Time, 1)
	patchTime := gomonkey.ApplyFunc(time.After,
		func(_ time.Duration) <-chan time.Time {
			return chanTime
		})
	defer patchTime.Reset()
	chanTime <- time.Now()
	event, code, err := ctl.handleWaitReportScaleInIsolateRanksStatus()
	assert.Equal(t, common.ReportTimeoutEvent, event)
	assert.Equal(t, common.WaitReportTimeout, code)
	assert.NoError(t, err)
}

func TestEventControllerHandleNotifyScaleInStrategy(t *testing.T) {
	ctl := &EventController{
		jobInfo: common.JobBaseInfo{JobId: "job-123"},
	}
	mockGetNodeRankIdsByFaultRanks := gomonkey.ApplyFuncReturn(common.GetNodeRankIdsByFaultRanks, []string{}, nil)
	defer mockGetNodeRankIdsByFaultRanks.Reset()
	event, code, err := ctl.handleNotifyScaleInStrategy()
	assert.Equal(t, "", event)
	assert.Equal(t, common.SignalQueueBusy, code)
	assert.Error(t, err)
}

func TestCheckWhetherPodVersionChangedFalse(t *testing.T) {
	ctl := &EventController{
		jobInfo: common.JobBaseInfo{JobId: "job-123"},
		faultPod: map[string]string{
			"1": "1",
		},
		prePod: map[string]string{
			"1": "1",
		},
	}
	mockGetNodeRankIdsByFaultRanks := gomonkey.ApplyFuncReturn(pod.GetPodByRankIndex, v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "job-123",
			UID:  "1",
		},
	})
	defer mockGetNodeRankIdsByFaultRanks.Reset()
	result := ctl.checkWhetherPodChanged()
	assert.Equal(t, false, result)
}

func TestCheckWhetherPodVersionChangedTrue(t *testing.T) {
	ctl := &EventController{
		jobInfo: common.JobBaseInfo{JobId: "job-123"},
		faultPod: map[string]string{
			"1": "1",
		},
		prePod: map[string]string{
			"1": "1",
		},
	}
	mockGetNodeRankIdsByFaultRanks := gomonkey.ApplyFuncReturn(pod.GetPodByRankIndex, v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "job-123",
			UID:  "2",
		},
	})
	defer mockGetNodeRankIdsByFaultRanks.Reset()
	result := ctl.checkWhetherPodChanged()
	assert.Equal(t, true, result)
}

// TestListenScheduleResultMain tests main flow of listenScheduleResult
func TestListenScheduleResultMain(t *testing.T) {
	convey.Convey("Test listenScheduleResult main flow", t, func() {
		loggerPatch := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {})
		loggerPatch.ApplyFunc(hwlog.RunLog.Warnf, func(format string, args ...interface{}) {})
		defer loggerPatch.Reset()

		testNotSupportStrategy(t)
		testSupportStrategyAndRunning(t)
		testFaultOccurAgain(t)
	})
}

func testNotSupportStrategy(t *testing.T) {
	convey.Convey("01-job does not support recover or recover-in-place strategy", func() {
		ctl := &EventController{
			jobInfo:            common.JobBaseInfo{JobId: "test-job-id"},
			scheduleResultChan: make(chan bool, 1),
		}

		mockConfigTargetStrategy := gomonkey.ApplyPrivateMethod(
			reflect.TypeOf(ctl),
			"configTargetStrategy",
			func(_ *EventController, _ string) bool { return false },
		)
		defer mockConfigTargetStrategy.Reset()

		ctl.listenScheduleResult()
		convey.So(len(ctl.scheduleResultChan), convey.ShouldEqual, 0)
	})
}

func testSupportStrategyAndRunning(t *testing.T) {
	convey.Convey("02-job supports recover strategy and pg is running", func() {
		ctl := &EventController{
			controllerContext:   context.Background(),
			restartFaultProcess: true,
			jobInfo: common.JobBaseInfo{
				JobId: "test-job-id",
				RecoverConfig: common.RecoverConfig{
					MindXConfigStrategies: []string{constant.ProcessRecoverStrategyName},
				},
			},
			scheduleResultChan: make(chan bool, 1),
		}
		patch := gomonkey.ApplyFuncReturn(podgroup.JudgeIsRunningByJobKey, true)
		defer patch.Reset()
		ctl.listenScheduleResult()
		convey.So(len(ctl.scheduleResultChan), convey.ShouldEqual, 1)
		convey.So(<-ctl.scheduleResultChan, convey.ShouldBeTrue)
	})
}

func testFaultOccurAgain(t *testing.T) {
	convey.Convey("03-job fault occurs again in scale-running not listen", func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId: "test-job-id",
				RecoverConfig: common.RecoverConfig{
					MindXConfigStrategies: []string{constant.ProcessRecoverStrategyName},
				},
			},
			restartFaultProcess: true,
			scheduleResultChan:  make(chan bool, 1),
			latestStrategy:      []string{constant.ScaleInStrategyName},
		}
		ctl.listenScheduleResult()
		convey.So(len(ctl.scheduleResultChan), convey.ShouldEqual, 0)
	})
}

func TestDealWithForceRelease(t *testing.T) {
	convey.Convey("Test dealWithForceRelease", t, func() {
		ctl := &EventController{
			jobInfo:  common.JobBaseInfo{JobId: "test-job-id"},
			faultPod: map[string]string{"test-pod": "test-node"},
		}
		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {})
		defer patches.Reset()
		patches.ApplyFunc(hwlog.RunLog.Warnf, func(format string, args ...interface{}) {})
		convey.Convey("01-should update fault info when fault info is not nil", func() {
			mockGetFaultPod := gomonkey.ApplyMethodFunc(reflect.TypeOf(ctl), "GetFaultPod", func() map[string]string {
				return map[string]string{"test-pod": "test-node"}
			})
			defer mockGetFaultPod.Reset()
			mockGetFaultInfo := gomonkey.ApplyFuncReturn(job.GetJobFaultSdIdAndNodeName, map[int]api.SuperPodFaultInfos{
				0: {JobId: "test-job-id", SdIds: []string{"sd1"}, NodeNames: []string{"node1"}}})
			defer mockGetFaultInfo.Reset()
			mockUpdateFaultInfo := gomonkey.ApplyFuncReturn(kube.CreateOrUpdateSuperPodFaultInfo)
			defer mockUpdateFaultInfo.Reset()
			mockSleep := gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {})
			defer mockSleep.Reset()
			ctl.dealWithForceRelease()
		})

		convey.Convey("02-should warn when fault info is nil", func() {
			mockGetFaultPod := gomonkey.ApplyMethodFunc(reflect.TypeOf(ctl), "GetFaultPod", func() map[string]string {
				return map[string]string{"test-pod": "test-node"}
			})
			defer mockGetFaultPod.Reset()
			mockGetFaultInfo := gomonkey.ApplyFuncReturn(job.GetJobFaultSdIdAndNodeName, nil)
			defer mockGetFaultInfo.Reset()
			ctl.dealWithForceRelease()
		})

		convey.Convey("03-should handle update fault info error", func() {
			mockGetFaultPod := gomonkey.ApplyMethodFunc(reflect.TypeOf(ctl), "GetFaultPod", func() map[string]string {
				return map[string]string{"test-pod": "test-node"}
			})
			defer mockGetFaultPod.Reset()
			mockGetFaultInfo := gomonkey.ApplyFuncReturn(job.GetJobFaultSdIdAndNodeName, map[int]api.SuperPodFaultInfos{
				0: {JobId: "test-job-id", SdIds: []string{"sd1"}, NodeNames: []string{"node1"}}})
			defer mockGetFaultInfo.Reset()
			mockUpdateFaultInfo := gomonkey.ApplyFuncReturn(
				kube.CreateOrUpdateSuperPodFaultInfo)
			defer mockUpdateFaultInfo.Reset()
			mockSleep := gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {})
			defer mockSleep.Reset()
			ctl.dealWithForceRelease()
		})
	})
}

func TestWaitScaleOutStateNotScaleInRunningState(t *testing.T) {
	convey.Convey("Test waitScaleOut when state is not ScaleInRunningState", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{JobId: "test-job-id"},
			state:   common.NewStateMachine("NotScaleInRunningState", []common.TransRule{}),
		}
		timePatch := gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {})
		defer timePatch.Reset()
		ctl.waitScaleOut()
		convey.So(true, convey.ShouldBeTrue)
	})
}

func TestWaitScaleOutJobObjectNil(t *testing.T) {
	convey.Convey("Test waitScaleOut when job object is nil", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{JobId: "test-job-id"},
			state: common.NewStateMachine(common.ScaleInRunningState, []common.TransRule{
				{Src: common.ScaleInRunningState, Event: common.FinishEvent,
					Dst: common.InitState, Handler: func() (nextEvent string, code common.RespCode, err error) {
						return "", 0, err
					}},
			}),
			events:            make(chan string, 1),
			controllerContext: context.Background(),
		}
		patches := gomonkey.ApplyFunc(statistics.GetJob, func(jobId string) metav1.Object { return nil }).
			ApplyFunc(time.Sleep, func(d time.Duration) {})
		defer patches.Reset()
		go ctl.waitScaleOut()
		event := <-ctl.events
		convey.So(event, convey.ShouldEqual, common.FinishEvent)
	})
}

func TestWaitScaleOutJobSucceeded(t *testing.T) {
	convey.Convey("Test waitScaleOut when job is succeeded", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{JobId: "test-job-id"},
			state: common.NewStateMachine(common.ScaleInRunningState, []common.TransRule{
				{Src: common.ScaleInRunningState, Event: common.FinishEvent,
					Dst: common.InitState, Handler: func() (nextEvent string, code common.RespCode, err error) {
						return "", 0, err
					}},
			}),
			events:            make(chan string, 1),
			controllerContext: context.Background(),
		}
		acJobInfo := &batchv1.AscendJob{Status: apiv1.JobStatus{Conditions: []apiv1.JobCondition{
			{Type: apiv1.JobSucceeded, Status: "true"},
		}}}
		patches := gomonkey.ApplyFunc(statistics.GetJob, func(jobId string) metav1.Object { return acJobInfo }).
			ApplyFunc(util.IsSucceeded, func(status apiv1.JobStatus) bool { return true }).
			ApplyFunc(time.Sleep, func(d time.Duration) {})
		defer patches.Reset()
		go ctl.waitScaleOut()
		event := <-ctl.events
		convey.So(event, convey.ShouldEqual, common.FinishEvent)
	})
}
