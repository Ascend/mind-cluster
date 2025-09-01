// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package recover a series of controller test function
package recover

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"k8s.io/api/core/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/application/faultmanager/cmprocess/retry"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/interface/grpc/recover"
	"clusterd/pkg/interface/kube"
)

const (
	keepAliveSeconds = 10
	sliceLength3     = 3
	numInt2          = 2
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		return
	}
}

type stressTestSender struct {
	mockStream
}

func (s *stressTestSender) Send(signal *pb.StressTestResponse) error {
	return nil
}

type notifyStressTestSender struct {
	mockStream
}

func (s *notifyStressTestSender) Send(signal *pb.StressTestRankParams) error {
	return nil
}

type sender struct {
	mockStream
}

type switchNicSender struct {
	mockStream
}

type notifySwitchNicSender struct {
	mockStream
}

func (s *notifySwitchNicSender) Send(signal *pb.SwitchRankList) error {
	return nil
}

func (s *switchNicSender) Send(signal *pb.SwitchNicResponse) error {
	return nil
}

func (s *sender) Send(signal *pb.ProcessManageSignal) error {
	return nil
}

type mockStream struct {
}

func (ms *mockStream) Context() context.Context {
	return context.Background()
}

func (ms *mockStream) SendMsg(m interface{}) error {
	return nil
}

func (ms *mockStream) RecvMsg(m interface{}) error {
	return nil
}

func (ms *mockStream) SetHeader(md metadata.MD) error {
	return nil
}

func (ms *mockStream) SendHeader(md metadata.MD) error {
	return nil
}

func (ms *mockStream) SetTrailer(md metadata.MD) {
}

func TestHandleNotifyDump(t *testing.T) {
	convey.Convey("Test handleNotifyDump", t, func() {
		ctl := &EventController{
			uuid: "test-uuid",
			jobInfo: common.JobBaseInfo{
				JobId:     "test-job-id",
				JobName:   "test-job-name",
				Namespace: "test-namespace",
			},
		}
		convey.Convey("01-update cache fault and pod fail, should return err", func() {
			mockFunc := gomonkey.ApplyPrivateMethod(&EventController{}, "updateCacheFaultAndPod",
				func() ([]*pb.FaultRank, []string, error) { return nil, nil, errors.New("mock error") })
			defer mockFunc.Reset()

			result, respCode, err := ctl.handleNotifyDump()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(respCode == common.ServerInnerError, convey.ShouldBeTrue)
			convey.So(result == "", convey.ShouldBeTrue)
		})

		mockFunc := gomonkey.ApplyPrivateMethod(&EventController{}, "updateCacheFaultAndPod",
			func() ([]*pb.FaultRank, []string, error) { return nil, []string{"1", "2"}, nil })
		defer mockFunc.Reset()
		convey.Convey("02-write reset info fail, should return err", func() {
			mockFunc1 := gomonkey.ApplyFuncReturn(common.WriteResetInfoToCM, nil, errors.New("mock error1"))
			defer mockFunc1.Reset()
			result, respCode, err := ctl.handleNotifyDump()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(respCode == common.OperateConfigMapError, convey.ShouldBeTrue)
			convey.So(result == "", convey.ShouldBeTrue)
		})
		mockFunc1 := gomonkey.ApplyFuncReturn(common.WriteResetInfoToCM,
			&v1.ConfigMap{Data: map[string]string{"key": "value"}}, nil)
		defer mockFunc1.Reset()
		convey.Convey("03-signal enqueue success, should return nil", func() {
			mockFunc2 := gomonkey.ApplyPrivateMethod(&EventController{}, "signalEnqueue",
				func(signal *pb.ProcessManageSignal) (string, common.RespCode, error) { return "", common.OK, nil })
			defer mockFunc2.Reset()
			_, respCode, err := ctl.handleNotifyDump()
			convey.So(err, convey.ShouldBeNil)
			convey.So(respCode == common.OK, convey.ShouldBeTrue)
		})
	})
}

func TestHandleNotifyWaitFaultFlushing(t *testing.T) {
	convey.Convey("Test handleNotifyWaitFaultFlushing", t, func() {
		ctl := &EventController{uuid: "test-uuid", jobInfo: common.JobBaseInfo{
			JobId:         "test-job-id",
			JobName:       "test-job-name",
			Namespace:     "test-namespace",
			RecoverConfig: common.RecoverConfig{PlatFormMode: true}}}
		convey.Convey("01-wait plat form strategy fail, should return timeout event", func() {
			mockFunc := gomonkey.ApplyFuncReturn(WaitPlatFormStrategyReady, "", errors.New("mock error"))
			defer mockFunc.Reset()
			result, respCode, err := ctl.handleNotifyWaitFaultFlushing()
			convey.So(err, convey.ShouldBeNil)
			convey.So(respCode == common.WaitPlatStrategyTimeout, convey.ShouldBeTrue)
			convey.So(result == common.WaitPlatStrategyTimeoutEvent, convey.ShouldBeTrue)
		})
		strategy := "dump"
		mockFunc := gomonkey.ApplyFuncReturn(WaitPlatFormStrategyReady, strategy, nil)
		defer mockFunc.Reset()
		convey.Convey("02-should dump when occur fault", func() {
			mockFunc1 := gomonkey.ApplyPrivateMethod(&EventController{}, "shouldDumpWhenOccurFault",
				func(*EventController) bool { return true })
			defer mockFunc1.Reset()
			result, respCode, err := ctl.handleNotifyWaitFaultFlushing()
			convey.So(err, convey.ShouldBeNil)
			convey.So(respCode == common.OK, convey.ShouldBeTrue)
			convey.So(result == common.DumpForFaultEvent, convey.ShouldBeTrue)
		})
		convey.Convey("04-retry write reset cm fail, should return operate cm error", func() {
			mockFunc1 := gomonkey.ApplyPrivateMethod(&EventController{}, "shouldDumpWhenOccurFault",
				func(*EventController) bool { return false })
			defer mockFunc1.Reset()
			mockFunc2 := gomonkey.ApplyFuncReturn(common.RetryWriteResetCM, nil, errors.New("mock error"))
			defer mockFunc2.Reset()
			result, respCode, err := ctl.handleNotifyWaitFaultFlushing()
			convey.So(err, convey.ShouldBeNil)
			convey.So(respCode == common.OperateConfigMapError, convey.ShouldBeTrue)
			convey.So(result == common.NotifyFailEvent, convey.ShouldBeTrue)
		})
	})
}

type getOnlySupportDumpStrategyTestCase struct {
	name    string
	ctl     *EventController
	wantRet bool
}

func buildGetOnlySupportDumpStrategyTestCases() []getOnlySupportDumpStrategyTestCase {
	firstDump := []string{constant.ProcessDumpStrategyName, constant.ProcessExitStrategyName}
	notFirstDump := []string{constant.ProcessRecoverStrategyName, constant.ProcessDumpStrategyName,
		constant.ProcessExitStrategyName}
	notSupportDump := []string{constant.ProcessRecoverStrategyName, constant.ProcessExitStrategyName}
	normalModeCase := getOnlySupportDumpStrategyNotPlatModeTestCase(firstDump, notFirstDump, notSupportDump)
	platModeCase := getOnlySupportDumpStrategyPlatModeTestCase(firstDump, notFirstDump, notSupportDump)
	return append(normalModeCase, platModeCase...)
}

func getOnlySupportDumpStrategyNotPlatModeTestCase(firstDump, notFirstDump,
	notSupportDump []string) []getOnlySupportDumpStrategyTestCase {
	return []getOnlySupportDumpStrategyTestCase{
		{
			name: "01-ProcessRecoverEnable is false, should return false",
			ctl: &EventController{
				jobInfo: common.JobBaseInfo{RecoverConfig: common.RecoverConfig{
					ProcessRecoverEnable: false, PlatFormMode: false, MindXConfigStrategies: firstDump}},
			},
			wantRet: false,
		},
		{
			name: "02-not platform mode, cluster not first support dump, should return false",
			ctl: &EventController{
				jobInfo: common.JobBaseInfo{RecoverConfig: common.RecoverConfig{
					ProcessRecoverEnable: true, PlatFormMode: false, MindXConfigStrategies: notFirstDump}},
			},
			wantRet: false,
		},
		{
			name: "03-not platform mode, cluster support dump, should return true",
			ctl: &EventController{
				jobInfo: common.JobBaseInfo{RecoverConfig: common.RecoverConfig{
					ProcessRecoverEnable: true, PlatFormMode: false, MindXConfigStrategies: firstDump}},
			},
			wantRet: true,
		},
	}
}

func getOnlySupportDumpStrategyPlatModeTestCase(firstDump, notFirstDump,
	notSupportDump []string) []getOnlySupportDumpStrategyTestCase {
	return []getOnlySupportDumpStrategyTestCase{
		{
			name: "04-platform mode, platform strategy is not dump, should return false",
			ctl: &EventController{
				jobInfo: common.JobBaseInfo{RecoverConfig: common.RecoverConfig{
					ProcessRecoverEnable: true, PlatFormMode: true, MindXConfigStrategies: firstDump}},
				platStrategy: constant.ProcessExitStrategyName,
			},
			wantRet: false,
		},
		{
			name: "05-platform mode, platform strategy is dump, cluster not contain dump, should return false",
			ctl: &EventController{
				jobInfo: common.JobBaseInfo{RecoverConfig: common.RecoverConfig{
					ProcessRecoverEnable: true, PlatFormMode: true, MindXConfigStrategies: notSupportDump}},
				platStrategy: constant.ProcessDumpStrategyName,
			},
			wantRet: false,
		},
		{
			name: "06-platform mode, platform strategy is dump, cluster contain dump, should return true",
			ctl: &EventController{
				jobInfo: common.JobBaseInfo{RecoverConfig: common.RecoverConfig{
					ProcessRecoverEnable: true, PlatFormMode: true, MindXConfigStrategies: notFirstDump}},
				platStrategy: constant.ProcessDumpStrategyName,
			},
			wantRet: true,
		},
	}
}

func TestOnlySupportDumpStrategy(t *testing.T) {
	tests := buildGetOnlySupportDumpStrategyTestCases()
	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			shouldDump := tt.ctl.onlySupportDumpStrategy()
			convey.So(shouldDump, convey.ShouldEqual, tt.wantRet)
		})
	}
}

func TestShouldDumpWhenOccurFault(t *testing.T) {
	ctl := EventController{}
	convey.Convey("shouldDumpWhenOccurFault", t, func() {
		convey.Convey("01-ProcessRecoverEnable is false, should return false", func() {
			ctl.jobInfo.ProcessRecoverEnable = false
			convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeFalse)
		})
		ctl.jobInfo.ProcessRecoverEnable = true
		convey.Convey("02-config not support dump, should return false", func() {
			ctl.jobInfo.MindXConfigStrategies = []string{constant.ProcessRecoverStrategyName,
				constant.ProcessDumpStrategyName}
			convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeFalse)
		})
		ctl.jobInfo.MindXConfigStrategies = []string{constant.ProcessDumpStrategyName}
		ctl.jobInfo.PlatFormMode = false
		convey.Convey("03-healthState is healthy, should return false", func() {
			ctl.healthState = constant.HealthyState
			convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeFalse)
		})
		convey.Convey("05-healthState is not healthy", func() {
			convey.Convey("healthState is unhealthy should return true", func() {
				ctl.healthState = constant.UnHealthyState
				convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeTrue)
			})
			ctl.healthState = constant.SubHealthyState
			convey.Convey("06-healthState is subHealthy and graceExit is false, should return false", func() {
				ctl.jobInfo.GraceExit = false
				convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeFalse)
			})
			convey.Convey("07-healthState is subHealthy and graceExit is true, should return true", func() {
				ctl.jobInfo.GraceExit = true
				convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeTrue)
			})
		})
	})
}

func TestUpdateCacheFaultAndPod(t *testing.T) {
	ctl := EventController{
		faultPod:        map[string]string{},
		faultPodVersion: make(map[string]string),
	}
	convey.Convey("updateCacheFaultAndPod", t, func() {
		faultRank := []*pb.FaultRank{{RankId: "1", FaultType: constant.NormalFaultType}}
		ranks := []string{"1"}
		mockFunc := gomonkey.ApplyPrivateMethod(&EventController{}, "normalFaultAssociateSameNodeRank",
			func(*EventController) ([]*pb.FaultRank, []string) { return faultRank, ranks })
		defer mockFunc.Reset()

		convey.Convey("01-get pod map failed, should return error", func() {
			mockFunc1 := gomonkey.ApplyFuncReturn(common.GetPodMap, nil, errors.New("mock error"))
			defer mockFunc1.Reset()
			_, _, err := ctl.updateCacheFaultAndPod()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-update cache fault and pod success, should not return error", func() {
			podMap := map[string]string{"1": "uuid1"}
			mockFunc1 := gomonkey.ApplyFuncReturn(common.GetPodMap, podMap, nil)
			defer mockFunc1.Reset()
			retFaultrank, retRank, err := ctl.updateCacheFaultAndPod()
			convey.So(err, convey.ShouldBeNil)
			convey.So(retFaultrank, convey.ShouldResemble, faultRank)
			convey.So(retRank, convey.ShouldResemble, ranks)
			convey.So(ctl.faultPod, convey.ShouldResemble, podMap)
			convey.So(ctl.cacheRetryFault, convey.ShouldBeNil)
			convey.So(ctl.cacheNormalFault, convey.ShouldResemble, faultRank)
		})
	})
}

func newJobInfoWithStrategy(strategies []string) common.JobBaseInfo {
	jobInfo := common.JobBaseInfo{
		JobId:     "test-job",
		JobName:   "test-job-name",
		Namespace: "test-namespace",
		PgName:    "test-pg-name",
	}
	for _, strategy := range strategies {
		jobInfo.MindXConfigStrategies = append(jobInfo.MindXConfigStrategies, strategy)
	}
	return jobInfo
}

func TestNewEventController(t *testing.T) {
	convey.Convey("Testing NewEventController", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		convey.So(ctl, convey.ShouldNotBeNil)
		convey.So(ctl.jobInfo.JobId, convey.ShouldEqual, jobInfo.JobId)
		convey.So(ctl.keepAliveSecond, convey.ShouldEqual, keepAliveSeconds)
	})
}

func TestGetFaultPod(t *testing.T) {
	convey.Convey("Testing GetFaultPod", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.faultPod = map[string]string{
			"rank1": "pod1",
			"rank2": "pod2",
		}
		faultPod := ctl.GetFaultPod()
		convey.So(faultPod, convey.ShouldHaveLength, sliceLength3-1)
		convey.So(faultPod["rank1"], convey.ShouldEqual, "pod1")
		convey.So(faultPod["rank2"], convey.ShouldEqual, "pod2")
	})
}

func TestMergeFaultPod(t *testing.T) {
	convey.Convey("Testing mergeFaultPod", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.faultPod = map[string]string{
			"rank1": "pod1",
		}
		newFaultPod := map[string]string{
			"rank2": "pod2",
			"rank3": "pod3",
		}
		ctl.mergeFaultPod(newFaultPod)
		convey.So(ctl.faultPod, convey.ShouldHaveLength, sliceLength3)
		convey.So(ctl.faultPod["rank1"], convey.ShouldEqual, "pod1")
		convey.So(ctl.faultPod["rank2"], convey.ShouldEqual, "pod2")
		convey.So(ctl.faultPod["rank3"], convey.ShouldEqual, "pod3")
	})
}

func TestSaveCacheFault(t *testing.T) {
	convey.Convey("Testing saveCacheFault", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		faults := []*pb.FaultRank{
			{RankId: "rank1", FaultType: constant.NormalFaultType},
			{RankId: "rank2", FaultType: constant.UceFaultType},
		}
		ctl.saveCacheFault(faults)
		convey.So(ctl.cacheNormalFault, convey.ShouldHaveLength, 1)
		convey.So(ctl.cacheNormalFault[0].RankId, convey.ShouldEqual, "rank1")
		convey.So(ctl.cacheRetryFault, convey.ShouldHaveLength, 1)
		convey.So(ctl.cacheRetryFault[0].RankId, convey.ShouldEqual, "rank2")
	})
}

func TestReset(t *testing.T) {
	convey.Convey("Testing reset", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.faultFlushing = true
		ctl.uuid = "test-uuid"
		ctl.latestStrategy = []string{"strategy1"}
		ctl.faultPod = map[string]string{"rank1": "pod1"}
		ctl.cacheNormalFault = []*pb.FaultRank{{RankId: "rank1", FaultType: constant.NormalFaultType}}
		ctl.cacheRetryFault = []*pb.FaultRank{{RankId: "rank2", FaultType: constant.UceFaultType}}
		patch := gomonkey.ApplyFuncReturn(common.RetryWriteResetCM,
			&v1.ConfigMap{Data: make(map[string]string)}, nil)
		defer patch.Reset()
		ctl.reset(false)
		convey.So(ctl.faultFlushing, convey.ShouldBeFalse)
		convey.So(ctl.uuid, convey.ShouldBeEmpty)
		convey.So(ctl.latestStrategy, convey.ShouldHaveLength, 0)
		convey.So(ctl.faultPod, convey.ShouldHaveLength, 0)
		convey.So(ctl.cacheNormalFault, convey.ShouldHaveLength, 0)
		convey.So(ctl.cacheRetryFault, convey.ShouldHaveLength, 0)
	})
}

func TestCleanControllerMap(t *testing.T) {
	convey.Convey("Testing cleanControllerMapAndSet", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.cleanControllerMapAndSet()
		convey.So(ctl.stressTestParam, convey.ShouldHaveLength, 0)
		convey.So(ctl.isolateNodes, convey.ShouldHaveLength, 0)
	})
}

func TestInitControllerChan(t *testing.T) {
	convey.Convey("Testing initControllerChan", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.initControllerChan()
		convey.So(ctl.events, convey.ShouldNotBeNil)
		convey.So(ctl.signalChan, convey.ShouldNotBeNil)
		convey.So(ctl.reportStopCompleteChan, convey.ShouldNotBeNil)
		convey.So(ctl.reportRecoverStrategyChan, convey.ShouldNotBeNil)
		convey.So(ctl.reportStatusChan, convey.ShouldNotBeNil)
		convey.So(ctl.scheduleResultChan, convey.ShouldNotBeNil)
		convey.So(ctl.switchNicResponse, convey.ShouldNotBeNil)
		convey.So(ctl.switchRankList, convey.ShouldNotBeNil)
		convey.So(ctl.switchRankResult, convey.ShouldNotBeNil)
	})
}

func TestCleanControllerSlice(t *testing.T) {
	convey.Convey("Testing cleanControllerSlice", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.cleanControllerSlice()
		convey.So(ctl.cacheRetryFault, convey.ShouldHaveLength, 0)
		convey.So(ctl.cacheNormalFault, convey.ShouldHaveLength, 0)
		convey.So(ctl.latestRecoverResult, convey.ShouldHaveLength, 0)
		convey.So(ctl.agentReportStrategies, convey.ShouldHaveLength, 0)
		convey.So(ctl.globalSwitchRankIDs, convey.ShouldHaveLength, 0)
		convey.So(ctl.globalOps, convey.ShouldHaveLength, 0)
	})
}

func TestSelectKeepAlive(t *testing.T) {
	convey.Convey("Testing selectKeepAlive", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		signalChan := make(chan *pb.ProcessManageSignal, 1)
		exit := ctl.selectKeepAlive(ctl.controllerContext, signalChan)
		convey.So(exit, convey.ShouldBeFalse)
	})
}

func TestKeepAlive(t *testing.T) {
	convey.Convey("Testing keepAlive", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		signalChan := make(chan *pb.ProcessManageSignal, 1)
		go ctl.keepAlive()
		time.Sleep(time.Second)
		select {
		case <-signalChan:
			convey.So(true, convey.ShouldBeTrue)
		default:
			convey.So(false, convey.ShouldBeFalse)
		}
	})
}

func TestAnnotationWithRetryStrategy(t *testing.T) {
	convey.Convey("Testing annotationWithRetryStrategy", t, func() {
		jobInfo := newJobInfoWithStrategy([]string{constant.ProcessRetryStrategyName})
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		hasRetry := ctl.annotationWithRetryStrategy()
		convey.So(hasRetry, convey.ShouldBeTrue)
	})
}

func TestSupportRetryStrategy(t *testing.T) {
	convey.Convey("Testing supportRetryStrategy", t, func() {
		jobInfo := newJobInfoWithStrategy([]string{constant.ProcessRetryStrategyName})
		jobInfo.PlatFormMode = true
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		hasRetry := ctl.supportRetryStrategy()
		convey.So(hasRetry, convey.ShouldBeFalse)

		ctl.jobInfo.PlatFormMode = false
		hasRetry = ctl.supportRetryStrategy()
		convey.So(hasRetry, convey.ShouldBeTrue)

		ctl.jobInfo.PlatFormMode = true
		hasRetry = ctl.supportRetryStrategy()
		convey.So(hasRetry, convey.ShouldBeFalse)

		ctl.jobInfo.PlatFormMode = true
		ctl.platStrategy = constant.ProcessRetryStrategyName
		hasRetry = ctl.supportRetryStrategy()
		convey.So(hasRetry, convey.ShouldBeTrue)

		ctl.jobInfo = newJobInfoWithStrategy(nil)
		hasRetry = ctl.supportRetryStrategy()
		convey.So(hasRetry, convey.ShouldBeFalse)
	})
}

// Test case 1: MindXConfigStrategies contains recover strategy and platform mode is enabled
func testSupportRecoverStrategyCase1() {
	convey.Convey("MindXConfigStrategies contains recover strategy "+
		"and platform mode is enabled", func() {
		jobInfo := newJobInfoWithStrategy([]string{constant.ProcessRecoverStrategyName})
		jobInfo.PlatFormMode = true
		ctl := &EventController{
			jobInfo:               jobInfo,
			agentReportStrategies: []string{constant.ProcessRecoverStrategyName},
		}
		hasRecover := ctl.supportRecoverStrategy()
		convey.So(hasRecover, convey.ShouldBeFalse) // Expected result: false

		ctl.platStrategy = constant.ProcessRecoverStrategyName
		hasRecover = ctl.supportRecoverStrategy()
		convey.So(hasRecover, convey.ShouldBeTrue) // Expected result: true
	})
}

// Test case 2: MindXConfigStrategies contains recover strategy but platform mode is disabled
func testSupportRecoverStrategyCase2() {
	convey.Convey("MindXConfigStrategies contains recover strategy "+
		"but platform mode is disabled", func() {
		jobInfo := newJobInfoWithStrategy([]string{constant.ProcessRecoverStrategyName})
		jobInfo.PlatFormMode = false
		ctl := &EventController{
			jobInfo:               jobInfo,
			agentReportStrategies: []string{constant.ProcessRecoverStrategyName},
		}
		hasRecover := ctl.supportRecoverStrategy()
		convey.So(hasRecover, convey.ShouldBeTrue) // Expected result: true
	})
}

// Test case 3: MindXConfigStrategies does not contain recover strategy but platform mode is enabled
func testSupportRecoverStrategyCase3() {
	convey.Convey("MindXConfigStrategies does not contain recover strategy "+
		"but platform mode is enabled", func() {
		jobInfo := newJobInfoWithStrategy(nil)
		jobInfo.PlatFormMode = true
		ctl := &EventController{
			jobInfo:               jobInfo,
			agentReportStrategies: []string{constant.ProcessRecoverStrategyName},
		}
		hasRecover := ctl.supportRecoverStrategy()
		convey.So(hasRecover, convey.ShouldBeFalse) // Expected result: false
	})
}

// Test case 4: MindXConfigStrategies does not contain recover strategy and platform mode is disabled
func testSupportRecoverStrategyCase4() {
	convey.Convey("MindXConfigStrategies does not contain recover strategy "+
		"and platform mode is disabled", func() {
		jobInfo := newJobInfoWithStrategy(nil)
		jobInfo.PlatFormMode = false
		ctl := &EventController{
			jobInfo:               jobInfo,
			agentReportStrategies: []string{},
		}
		hasRecover := ctl.supportRecoverStrategy()
		convey.So(hasRecover, convey.ShouldBeFalse) // Expected result: false
	})
}

// Test case 5: MindXConfigStrategies contains recover strategy
// but agentReportStrategies does not
func testSupportRecoverStrategyCase5() {
	convey.Convey("MindXConfigStrategies contains recover strategy "+
		"but agentReportStrategies does not", func() {
		jobInfo := newJobInfoWithStrategy([]string{constant.ProcessRecoverStrategyName})
		jobInfo.PlatFormMode = true
		ctl := &EventController{
			jobInfo:               jobInfo,
			agentReportStrategies: []string{},
		}
		hasRecover := ctl.supportRecoverStrategy()
		convey.So(hasRecover, convey.ShouldBeFalse) // Expected result: false
	})
}

// Test case 6: MindXConfigStrategies contains recover strategy, agentReportStrategies contains,
// but platform mode is disabled
func testSupportRecoverStrategyCase6() {
	convey.Convey("MindXConfigStrategies contains recover strategy, "+
		"agentReportStrategies contains, but platform mode is disabled", func() {
		jobInfo := newJobInfoWithStrategy([]string{constant.ProcessRecoverStrategyName})
		jobInfo.PlatFormMode = false
		ctl := &EventController{
			jobInfo:               jobInfo,
			agentReportStrategies: []string{constant.ProcessRecoverStrategyName},
		}
		hasRecover := ctl.supportRecoverStrategy()
		convey.So(hasRecover, convey.ShouldBeTrue) // Expected result: true
	})
}

// Test supportRecoverStrategy method
func TestSupportRecoverStrategy(t *testing.T) {
	convey.Convey("Testing supportRecoverStrategy", t, func() {
		testSupportRecoverStrategyCase1()
		testSupportRecoverStrategyCase2()
		testSupportRecoverStrategyCase3()
		testSupportRecoverStrategyCase4()
		testSupportRecoverStrategyCase5()
		testSupportRecoverStrategyCase6()
	})
}

func TestSupportDumpStrategy(t *testing.T) {
	convey.Convey("Testing supportDumpStrategy", t, func() {
		// Test Case 1: Platform mode is true, jobInfo has the strategy, and agent reports the strategy.
		// Expected Result: The method should return true as all conditions for dump support are met.
		convey.Convey("Test Case 1: Platform mode is true, "+
			"jobInfo has the strategy, and agent reports the strategy", func() {
			testSupportDumpStrategyCase1(t)
		})

		// Test Case 2: Platform mode is false, jobInfo has the strategy, and agent reports the strategy.
		// Expected Result: The method should return true because the agent supports the strategy.
		convey.Convey("Test Case 2: Platform mode is false, "+
			"jobInfo has the strategy, and agent reports the strategy", func() {
			testSupportDumpStrategyCase2(t)
		})

		// Test Case 3: Platform mode is true, jobInfo has the strategy, but agent does not report the strategy.
		// Expected Result: The method should return false as the agent does not support the strategy.
		convey.Convey("Test Case 3: Platform mode is true, "+
			"jobInfo has the strategy, but agent does not report the strategy", func() {
			testSupportDumpStrategyCase3(t)
		})

		// Test Case 4: Platform mode is true, jobInfo does not have the strategy, but agent reports the strategy.
		// Expected Result: The method should return false as jobInfo does not have the required strategy.
		convey.Convey("Test Case 4: Platform mode is true, "+
			"jobInfo does not have the strategy, but agent reports the strategy", func() {
			testSupportDumpStrategyCase4(t)
		})
	})
}

// testSupportDumpStrategyCase1 tests the scenario where platform mode is true,
// jobInfo has the strategy, and agent reports the strategy.
// Expected Result: The method should return true.
func testSupportDumpStrategyCase1(t *testing.T) {
	jobInfo := newJobInfoWithStrategy([]string{constant.ProcessDumpStrategyName})
	jobInfo.PlatFormMode = true
	ctl := &EventController{
		jobInfo:               jobInfo,
		agentReportStrategies: []string{constant.ProcessDumpStrategyName},
	}
	ctl.platStrategy = constant.ProcessDumpStrategyName
	hasDump := ctl.supportDumpStrategy()
	convey.So(hasDump, convey.ShouldBeTrue)
}

// testSupportDumpStrategyCase2 tests the scenario where platform mode is false,
// jobInfo has the strategy, and agent reports the strategy.
// Expected Result: The method should return true.
func testSupportDumpStrategyCase2(t *testing.T) {
	jobInfo := newJobInfoWithStrategy([]string{constant.ProcessDumpStrategyName})
	jobInfo.PlatFormMode = false
	ctl := &EventController{
		jobInfo:               jobInfo,
		agentReportStrategies: []string{constant.ProcessDumpStrategyName},
	}
	hasDump := ctl.supportDumpStrategy()
	convey.So(hasDump, convey.ShouldBeTrue)
}

// testSupportDumpStrategyCase3 tests the scenario where platform mode is true,
// jobInfo has the strategy, but agent does not report the strategy.
// Expected Result: The method should return false.
func testSupportDumpStrategyCase3(t *testing.T) {
	jobInfo := newJobInfoWithStrategy([]string{constant.ProcessDumpStrategyName})
	jobInfo.PlatFormMode = true
	ctl := &EventController{
		jobInfo:               jobInfo,
		agentReportStrategies: []string{},
	}
	hasDump := ctl.supportDumpStrategy()
	convey.So(hasDump, convey.ShouldBeFalse)
}

// testSupportDumpStrategyCase4 tests the scenario where platform mode is true,
// jobInfo does not have the strategy, but agent reports the strategy.
// Expected Result: The method should return false.
func testSupportDumpStrategyCase4(t *testing.T) {
	jobInfo := newJobInfoWithStrategy([]string{})
	jobInfo.PlatFormMode = true
	ctl := &EventController{
		jobInfo:               jobInfo,
		agentReportStrategies: []string{constant.ProcessDumpStrategyName},
	}
	hasDump := ctl.supportDumpStrategy()
	convey.So(hasDump, convey.ShouldBeFalse)
}

func TestAddEvent(t *testing.T) {
	convey.Convey("Testing addEvent", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		convey.Convey("case order mix", func() {
			ctl.addEvent("test-event")
			convey.So(len(ctl.events), convey.ShouldEqual, 0)
		})
		convey.Convey("case normal", func() {
			ctl.addEvent(common.FaultOccurEvent)
			convey.So(len(ctl.events), convey.ShouldEqual, 1)
		})
	})
}

func TestGetCtxAndEventChan(t *testing.T) {
	convey.Convey("Testing getCtxAndEventChan", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctx, ch := ctl.getCtxAndEventChan()
		convey.So(ctx, convey.ShouldNotBeNil)
		convey.So(ch, convey.ShouldNotBeNil)
	})
}

func TestListenEvent(t *testing.T) {
	convey.Convey("Testing listenEvent", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctl.addEvent(common.FaultOccurEvent)
		patches := gomonkey.ApplyFuncReturn(common.WriteResetInfoToCM, nil, nil)
		defer patches.Reset()
		patches.ApplyPrivateMethod(ctl, "handleNotifyWaitFaultFlushing",
			func() (string, common.RespCode, error) {
				return "", common.OK, nil
			})
		go ctl.listenEvent()
		time.Sleep(time.Second)
		convey.ShouldEqual(ctl.state.GetState(), common.NotifyWaitFaultFlushingState)
	})
}

func TestGetCtxAndSignalChan(t *testing.T) {
	convey.Convey("Testing getCtxAndSignalChan", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		ctx, ch := ctl.getCtxAndSignalChan()
		convey.So(ctx, convey.ShouldNotBeNil)
		convey.So(ch, convey.ShouldNotBeNil)
	})
}

func TestListenSendChannel(t *testing.T) {
	convey.Convey("Test listenSendChannel", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{JobId: "test-job-id"},
		}

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {})
		defer patches.Reset()

		convey.Convey("When exit on first iteration", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			sendChan := make(chan *pb.ProcessManageSignal)

			patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndSignalChan",
				func() (context.Context, chan *pb.ProcessManageSignal) {
					return ctx, sendChan
				})
			patches.ApplyPrivateMethod(ctl, "reset", func() {})
			defer patches.Reset()

			patches.ApplyPrivateMethod(ctl, "selectSendChannel", func(_ context.Context,
				_ chan *pb.ProcessManageSignal, _ pb.Recover_SubscribeProcessManageSignalServer) bool {
				return true

			})

			stream := &mockSubscribeProcessManageSignalServer{}

			ctl.listenSendChannel(stream)
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

type mockSubscribeProcessManageSignalServer struct {
	pb.Recover_SubscribeProcessManageSignalServer
}

func TestSignalEnqueue(t *testing.T) {
	convey.Convey("Testing signalEnqueue", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		signal := &pb.ProcessManageSignal{
			Uuid:       "test-uuid",
			JobId:      "test-job",
			SignalType: constant.KeepAliveSignalType,
		}
		_, code, err := ctl.signalEnqueue(signal)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestTrigger(t *testing.T) {
	convey.Convey("Testing trigger", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		nextEvent, code, err := ctl.trigger("test-event")
		convey.So(nextEvent, convey.ShouldBeEmpty)
		convey.So(code, convey.ShouldEqual, common.OrderMix)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestHandleFinish(t *testing.T) {
	convey.Convey("Testing handleFinish", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl), "reset", func(*EventController, bool) {
			return
		})
		defer patches.Reset()
		_, code, err := ctl.handleFinish()
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleFaultClear(t *testing.T) {
	convey.Convey("Testing handleFaultClear", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		patch := gomonkey.ApplyFuncReturn(common.RetryWriteResetCM,
			&v1.ConfigMap{Data: make(map[string]string)}, nil)
		defer patch.Reset()
		_, code, err := ctl.handleFaultClear()
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleNotifyStopTrain(t *testing.T) {
	convey.Convey("Testing handleNotifyStopTrain", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		_, code, err := ctl.handleNotifyStopTrain()
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleNotifyGlobalFault(t *testing.T) {
	convey.Convey("Testing handleNotifyGlobalFault", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		_, code, err := ctl.handleNotifyGlobalFault()
		convey.So(code, convey.ShouldEqual, common.JobNotExist)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestHandleNotifyGlobalFaultDefer(t *testing.T) {
	convey.Convey("Test handleNotifyGlobalFault defer logic", t, func() {
		ctl := &EventController{
			jobInfo:  common.JobBaseInfo{JobId: "test-job"},
			faultPod: map[string]string{"test-pod": "test"},
		}

		mockJobExists := gomonkey.ApplyFuncReturn(job.GetJobIsExists, true)
		defer mockJobExists.Reset()

		mockGetPod := gomonkey.ApplyFuncReturn(pod.GetPodByRankIndex, v1.Pod{})
		defer mockGetPod.Reset()

		mockGetFaultInfo := gomonkey.ApplyFuncReturn(job.GetJobFaultSdIdAndNodeName, nil)
		defer mockGetFaultInfo.Reset()

		mockUpdateFaultInfo := gomonkey.ApplyFuncReturn(kube.CreateOrUpdateSuperPodFaultInfo)
		defer mockUpdateFaultInfo.Reset()

		convey.Convey("01-should update fault info when no retry faults", func() {
			_, _, _ = ctl.handleNotifyGlobalFault()
		})
	})
}

func TestHandleNotifyDecidedStrategy(t *testing.T) {
	convey.Convey("Testing handleNotifyDecidedStrategy", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		mockGetNodeRankIdsByFaultRanks := gomonkey.ApplyFuncReturn(common.GetNodeRankIdsByFaultRanks, []string{}, nil)
		defer mockGetNodeRankIdsByFaultRanks.Reset()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		_, code, err := ctl.handleNotifyDecidedStrategy()
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleCheckRecoverResult(t *testing.T) {
	convey.Convey("Testing handleCheckRecoverResult", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		_, code, err := ctl.handleCheckRecoverResult()
		convey.So(code, convey.ShouldEqual, common.ServerInnerError)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestHandleKillJob(t *testing.T) {
	convey.Convey("Testing handleKillJob", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		patches := gomonkey.ApplyFunc(ctl.signalEnqueue,
			func(signal *pb.ProcessManageSignal) (string, common.RespCode, error) {
				return "", common.OK, nil
			})
		defer patches.Reset()

		_, code, err := ctl.handleKillJob()
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleWaitReportRecoverStrategy(t *testing.T) {
	convey.Convey("Testing handleWaitReportRecoverStrategy", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		reportChan := make(chan *pb.RecoverStrategyRequest, 1)
		ctl.reportRecoverStrategyChan = reportChan

		go func() {
			reportChan <- &pb.RecoverStrategyRequest{
				Strategies: []string{constant.ProcessRetryStrategyName},
			}
		}()

		event, code, err := ctl.handleWaitReportRecoverStrategy()
		convey.So(event, convey.ShouldEqual, common.ReceiveReportEvent)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleDecideRetryStrategy(t *testing.T) {
	convey.Convey("Testing handleDecideRetryStrategy", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		reportChan := make(chan *pb.RecoverStatusRequest, 1)
		ctl.reportStatusChan = reportChan

		go func() {
			reportChan <- &pb.RecoverStatusRequest{
				Status: &pb.Status{
					Code: int32(common.OK),
				},
				Strategy: constant.ProcessRetryStrategyName,
			}
		}()

		event, code, err := ctl.handleDecideRetryStrategy()
		convey.So(event, convey.ShouldEqual, common.ReceiveReportEvent)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleDecideRecoverStrategy(t *testing.T) {
	convey.Convey("Testing handleDecideRecoverStrategy", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		reportChan := make(chan *pb.RecoverStatusRequest, 1)
		ctl.reportStatusChan = reportChan

		go func() {
			reportChan <- &pb.RecoverStatusRequest{
				Status: &pb.Status{
					Code: int32(common.OK),
				},
				Strategy: constant.ProcessRecoverStrategyName,
			}
		}()

		event, code, err := ctl.handleDecideRecoverStrategy()
		convey.So(event, convey.ShouldEqual, common.ReceiveReportEvent)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleDecideDumpStrategy(t *testing.T) {
	convey.Convey("Testing handleDecideDumpStrategy", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		reportChan := make(chan *pb.RecoverStatusRequest, 1)
		ctl.reportStatusChan = reportChan

		go func() {
			reportChan <- &pb.RecoverStatusRequest{
				Status: &pb.Status{
					Code: int32(common.OK),
				},
				Strategy: constant.ProcessDumpStrategyName,
			}
		}()

		event, code, err := ctl.handleDecideDumpStrategy()
		convey.So(event, convey.ShouldEqual, common.ReceiveReportEvent)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleDecideExitStrategy(t *testing.T) {
	convey.Convey("Testing handleDecideExitStrategy", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		event, code, err := ctl.handleDecideExitStrategy()
		convey.So(event, convey.ShouldEqual, common.CheckResultFinishEvent)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleListenScheduleResult(t *testing.T) {
	convey.Convey("Testing handleListenScheduleResult", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		patches := gomonkey.ApplyFunc(job.GetJobIsRunning, func(jobId string) bool {
			return true
		})
		defer patches.Reset()

		event, code, err := ctl.handleListenScheduleResult()
		convey.So(event, convey.ShouldEqual, common.ScheduleSuccessEvent)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleRestartAllProcess(t *testing.T) {
	convey.Convey("Testing handleRestartAllProcess", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)

		patches := gomonkey.ApplyFunc(common.RetryWriteResetCM,
			func(jobName, namespace string, faultRanks []string,
				restartProcess bool, operation string) (*v1.ConfigMap, error) {
				return &v1.ConfigMap{}, nil
			})
		defer patches.Reset()

		event, code, err := ctl.handleRestartAllProcess()
		convey.So(event, convey.ShouldEqual, common.NotifySuccessEvent)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetCtxAndStopCompleteChan(t *testing.T) {
	convey.Convey("Test getCtxAndStopCompleteChan", t, func() {
		ctl := &EventController{
			lock:                   sync.RWMutex{},
			controllerContext:      context.Background(),
			reportStopCompleteChan: make(chan *pb.StopCompleteRequest),
		}

		patches := gomonkey.ApplyMethodFunc(&ctl.lock, "RLock", func() {
		})
		defer patches.Reset()

		patches.ApplyMethodFunc(&ctl.lock, "RUnlock", func() {
		})

		convey.Convey("Should return correct context and channel", func() {
			ctx, stopCompleteChan := ctl.getCtxAndStopCompleteChan()
			convey.So(ctx, convey.ShouldResemble, ctl.controllerContext)

			convey.So(stopCompleteChan, convey.ShouldResemble, ctl.reportStopCompleteChan)
		})
	})
}

func TestHandleWaitReportStopComplete(t *testing.T) {
	convey.Convey("Test handleWaitReportStopComplete", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{JobId: "test-job-id"},
			uuid:    "test-uuid",
		}

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {}).
			ApplyFunc(hwlog.RunLog.Warnf, func(format string, args ...interface{}) {}).
			ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})
		defer patches.Reset()

		testReportChanNil(ctl)
		testContextCanceled(ctl)
		testProcessNotReady(ctl)
		testValidReport(ctl)
	})
}

func testReportChanNil(ctl *EventController) {
	convey.Convey("When reportChan is nil", func() {
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndStopCompleteChan",
			func() (context.Context, chan *pb.StopCompleteRequest) {
				return context.Background(), nil
			})
		defer patches.Reset()

		event, respCode, err := ctl.handleWaitReportStopComplete()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId))
	})
}

func testContextCanceled(ctl *EventController) {
	convey.Convey("When context is canceled", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndStopCompleteChan",
			func() (context.Context, chan *pb.StopCompleteRequest) {
				return ctx, make(chan *pb.StopCompleteRequest)
			})
		defer patches.Reset()

		event, respCode, err := ctl.handleWaitReportStopComplete()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.ControllerEventCancel)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testProcessNotReady(ctl *EventController) {
	convey.Convey("When receiving a report with ProcessNotReady status", func() {
		reportChan := make(chan *pb.StopCompleteRequest, 1)
		reportChan <- &pb.StopCompleteRequest{Status: &pb.Status{Code: common.ProcessNotReady}}
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndStopCompleteChan",
			func() (context.Context, chan *pb.StopCompleteRequest) {
				return context.Background(), reportChan
			})
		defer patches.Reset()

		event, respCode, err := ctl.handleWaitReportStopComplete()
		convey.So(event, convey.ShouldEqual, common.ProcessNotReadyEvent)
		convey.So(respCode, convey.ShouldEqual, common.ClientError)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testValidReport(ctl *EventController) {
	convey.Convey("When receiving a valid report", func() {
		reportChan := make(chan *pb.StopCompleteRequest, 1)
		reportChan <- &pb.StopCompleteRequest{Status: &pb.Status{Code: int32(common.OK)}}
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndStopCompleteChan",
			func() (context.Context, chan *pb.StopCompleteRequest) {
				return context.Background(), reportChan
			})
		defer patches.Reset()

		event, respCode, err := ctl.handleWaitReportStopComplete()
		convey.So(event, convey.ShouldEqual, common.ReceiveReportEvent)
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleWaitFlushFinish(t *testing.T) {
	convey.Convey("Test handleWaitFlushFinish", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{JobId: "test-job-id"},
		}

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {})
		defer patches.Reset()

		testUceFaultsOnly(ctl)
		testNormalFaultsExist(ctl)
		testContextDone(ctl)
		testTimeout(ctl)
	})
}

func testUceFaultsOnly(ctl *EventController) {
	convey.Convey("When only UCE faults exist and retry strategy is enabled", func() {
		patches := gomonkey.ApplyPrivateMethod(ctl, "takeRetryFault2NormalFault", func() ([]string, []string) {
			return []string{"uce-fault"}, []string{}
		})
		defer patches.Reset()

		patches.ApplyPrivateMethod(ctl, "annotationWithRetryStrategy", func() bool {
			return true
		})

		patches.ApplyPrivateMethod(ctl, "getCtxAndEventChan", func() (context.Context, chan interface{}) {
			return context.Background(), make(chan interface{})
		})

		event, respCode, err := ctl.handleWaitFlushFinish()
		convey.So(event, convey.ShouldEqual, common.FaultFlushFinishedEvent)
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testNormalFaultsExist(ctl *EventController) {
	convey.Convey("When normal faults exist", func() {
		patches := gomonkey.ApplyPrivateMethod(ctl, "takeRetryFault2NormalFault", func() ([]string, []string) {
			return []string{"uce-fault"}, []string{"normal-fault"}
		}).ApplyPrivateMethod(ctl, "annotationWithRetryStrategy", func() bool {
			return true
		}).ApplyPrivateMethod(ctl, "getCtxAndEventChan", func() (context.Context, chan interface{}) {
			return context.Background(), make(chan interface{})
		})
		defer patches.Reset()

		event, respCode, err := ctl.handleWaitFlushFinish()
		convey.So(event, convey.ShouldEqual, common.FaultFlushFinishedEvent)
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testContextDone(ctl *EventController) {
	convey.Convey("When context is done", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndEventChan",
			func() (context.Context, chan interface{}) {
				return ctx, make(chan interface{})
			})
		defer patches.Reset()

		patches.ApplyPrivateMethod(ctl, "takeRetryFault2NormalFault", func() ([]string, []string) {
			return []string{}, []string{}
		})

		event, respCode, err := ctl.handleWaitFlushFinish()
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testTimeout(ctl *EventController) {
	convey.Convey("When timeout occurs", func() {
		timeoutChan := make(chan time.Time, 1)
		timeoutChan <- time.Now()
		patches := gomonkey.ApplyFunc(time.After, func(d time.Duration) <-chan time.Time {
			return timeoutChan
		})
		defer patches.Reset()

		patches.ApplyPrivateMethod(ctl, "getCtxAndEventChan", func() (context.Context, chan interface{}) {
			return context.Background(), make(chan interface{})
		})

		patches.ApplyPrivateMethod(ctl, "takeRetryFault2NormalFault", func() ([]string, []string) {
			return []string{}, []string{}
		})

		event, respCode, err := ctl.handleWaitFlushFinish()
		convey.So(event, convey.ShouldEqual, common.FaultFlushFinishedEvent)
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestNormalFaultAssociateSameNodeRank(t *testing.T) {
	convey.Convey("Test normalFaultAssociateSameNodeRank", t, func() {
		ctl := &EventController{
			cacheNormalFault: []*pb.FaultRank{
				{RankId: "rank1"},
				{RankId: "rank2"},
			},
			jobInfo: common.JobBaseInfo{JobId: "test-job-id"},
		}

		testNoDuplicateRanks(ctl)
		testWithDuplicateRanks(ctl)
	})
}

func testNoDuplicateRanks(ctl *EventController) {
	convey.Convey("When no duplicate ranks exist", func() {
		patches := gomonkey.ApplyFunc(common.GetFaultRankIdsInSameNode, func(rankIds []string, deviceNum int) []string {
			return []string{"rank1", "rank2"}
		})
		defer patches.Reset()

		res, rankIds := ctl.normalFaultAssociateSameNodeRank()
		convey.So(res, convey.ShouldResemble, []*pb.FaultRank{
			{RankId: "rank1", FaultType: constant.NormalFaultType},
			{RankId: "rank2", FaultType: constant.NormalFaultType},
		})
		convey.So(rankIds, convey.ShouldResemble, []string{"rank1", "rank2"})
	})
}

func testWithDuplicateRanks(ctl *EventController) {
	convey.Convey("When duplicate ranks exist", func() {
		patches := gomonkey.ApplyFunc(common.GetFaultRankIdsInSameNode, func(rankIds []string, deviceNum int) []string {
			return []string{"rank1", "rank2", "rank1"}
		})
		defer patches.Reset()

		res, rankIds := ctl.normalFaultAssociateSameNodeRank()
		convey.So(res, convey.ShouldResemble, []*pb.FaultRank{
			{RankId: "rank1", FaultType: constant.NormalFaultType},
			{RankId: "rank2", FaultType: constant.NormalFaultType},
		})
		convey.So(rankIds, convey.ShouldResemble, []string{"rank1", "rank2"})
	})
}

func TestWriteConfirmFaultAndWaitPlatResultFault(t *testing.T) {
	convey.Convey("Test writeConfirmFaultAndWaitPlatResultFault", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId:     "test-job-id",
				PgName:    "test-pg",
				Namespace: "test-namespace",
			},
		}

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {})
		defer patches.Reset()
		patches.ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})

		testUpdateProcessConfirmFaultError(ctl)
		testWaitProcessResultFaultError(ctl)
		testPlatformStrategyError(ctl)
		testSuccess(ctl)
	})
}

func testUpdateProcessConfirmFaultError(ctl *EventController) {
	convey.Convey("When UpdateProcessConfirmFault returns an error", func() {
		patches := gomonkey.ApplyFunc(common.RemoveSliceDuplicateFaults, func(faults []*pb.FaultRank) []*pb.FaultRank {
			return faults
		})
		defer patches.Reset()

		patches.ApplyFunc(UpdateProcessConfirmFault, func(pgName, namespace string, faults []*pb.FaultRank) error {
			return errors.New("update error")
		})

		_, err := ctl.writeConfirmFaultAndWaitPlatResultFault([]*pb.FaultRank{{RankId: "rank1"}})
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "update process confirm fault err")
	})
}

func testWaitProcessResultFaultError(ctl *EventController) {
	convey.Convey("When WaitProcessResultFault returns an error", func() {
		patches := gomonkey.ApplyFunc(common.RemoveSliceDuplicateFaults, func(faults []*pb.FaultRank) []*pb.FaultRank {
			return faults
		})
		defer patches.Reset()

		patches.ApplyFunc(UpdateProcessConfirmFault, func(pgName, namespace string, faults []*pb.FaultRank) error {
			return nil
		})

		patches.ApplyFunc(WaitProcessResultFault, func(pgName, namespace string) ([]*pb.FaultRank, error) {
			return nil, errors.New("wait error")
		})

		_, err := ctl.writeConfirmFaultAndWaitPlatResultFault([]*pb.FaultRank{{RankId: "rank1"}})
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "wait process result fault err")
	})
}

func testPlatformStrategyError(ctl *EventController) {
	convey.Convey("When platformStrategy returns an error", func() {
		patches := gomonkey.ApplyFunc(common.RemoveSliceDuplicateFaults, func(faults []*pb.FaultRank) []*pb.FaultRank {
			return faults
		})
		defer patches.Reset()

		patches.ApplyFunc(UpdateProcessConfirmFault, func(pgName, namespace string, faults []*pb.FaultRank) error {
			return nil
		})

		patches.ApplyFunc(WaitProcessResultFault, func(pgName, namespace string) ([]*pb.FaultRank, error) {
			return []*pb.FaultRank{{RankId: "rank2"}}, nil
		})

		patches.ApplyFunc(platFormStrategy, func(pgName, namespace string, flag bool) (string, error) {
			return "", errors.New("strategy error")
		})

		_, err := ctl.writeConfirmFaultAndWaitPlatResultFault([]*pb.FaultRank{{RankId: "rank1"}})
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "confirm plat strategy err")
	})
}

func testSuccess(ctl *EventController) {
	convey.Convey("When all operations succeed", func() {
		patches := gomonkey.ApplyFunc(common.RemoveSliceDuplicateFaults, func(faults []*pb.FaultRank) []*pb.FaultRank {
			return faults
		})
		defer patches.Reset()

		patches.ApplyFunc(UpdateProcessConfirmFault, func(pgName, namespace string, faults []*pb.FaultRank) error {
			return nil
		})

		patches.ApplyFunc(WaitProcessResultFault, func(pgName, namespace string) ([]*pb.FaultRank, error) {
			return []*pb.FaultRank{{RankId: "rank2"}}, nil
		})

		patches.ApplyFunc(platFormStrategy, func(pgName, namespace string, flag bool) (string, error) {
			return "strategy", nil
		})

		result, err := ctl.writeConfirmFaultAndWaitPlatResultFault([]*pb.FaultRank{{RankId: "rank1"}})
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(result), convey.ShouldResemble, numInt2)
	})
}

func TestTakeUceFault2NormalFault(t *testing.T) {
	convey.Convey("Test takeRetryFault2NormalFault", t, func() {
		ctl := &EventController{
			lock: sync.RWMutex{},
			cacheRetryFault: []*pb.FaultRank{
				{RankId: "rank1", FaultType: constant.UceFaultType},
			},
			cacheNormalFault: []*pb.FaultRank{
				{RankId: "rank2", FaultType: constant.NormalFaultType},
			},
			latestRecoverResult: []*pb.RecoverStatusRequest{
				{Strategy: constant.ProcessRetryStrategyName},
			},
		}

		testRetryStrategyEnabled(ctl)
		testRetryStrategyDisabled(ctl)
		testNoRetryStrategySupport(ctl)
	})
}

func testRetryStrategyEnabled(ctl *EventController) {
	convey.Convey("When retry strategy is enabled", func() {
		patches := gomonkey.ApplyPrivateMethod(ctl, "supportRetryStrategy", func() bool {
			return true
		})
		defer patches.Reset()

		uceFaults, normalFaults := ctl.takeRetryFault2NormalFault()
		convey.So(uceFaults, convey.ShouldBeEmpty)
		convey.So(normalFaults, convey.ShouldResemble, []*pb.FaultRank{
			{RankId: "rank2", FaultType: constant.NormalFaultType},
			{RankId: "rank1", FaultType: constant.UceFaultType},
		})
	})
}

func testRetryStrategyDisabled(ctl *EventController) {
	convey.Convey("When retry strategy is disabled", func() {
		ctl.latestRecoverResult = []*pb.RecoverStatusRequest{} // 清空 recoverResult
		patches := gomonkey.ApplyPrivateMethod(ctl, "supportRetryStrategy", func() bool {
			return true
		})
		defer patches.Reset()

		uceFaults, normalFaults := ctl.takeRetryFault2NormalFault()
		convey.So(uceFaults, convey.ShouldBeEmpty)
		convey.So(normalFaults, convey.ShouldResemble, []*pb.FaultRank{
			{RankId: "rank2", FaultType: constant.NormalFaultType},
			{RankId: "rank1", FaultType: constant.UceFaultType},
		})
	})
}

func testNoRetryStrategySupport(ctl *EventController) {
	convey.Convey("When retry strategy is not supported", func() {
		patches := gomonkey.ApplyPrivateMethod(ctl, "supportRetryStrategy", func() bool {
			return false
		})
		defer patches.Reset()

		uceFaults, normalFaults := ctl.takeRetryFault2NormalFault()
		convey.So(uceFaults, convey.ShouldBeEmpty)
		convey.So(normalFaults, convey.ShouldResemble, []*pb.FaultRank{
			{RankId: "rank2", FaultType: constant.NormalFaultType},
			{RankId: "rank1", FaultType: constant.UceFaultType},
		})
	})
}

func TestNotifyFaultForUceFaultCase(t *testing.T) {
	convey.Convey("Test notifyFaultForRetryFaultCase", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId:         "test-job-id",
				JobName:       "test-job",
				Namespace:     "test-namespace",
				RecoverConfig: common.RecoverConfig{PlatFormMode: true},
			},
			faultPod:        make(map[string]string),
			faultPodVersion: make(map[string]string),
			uuid:            "test-uuid",
		}

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {})
		defer patches.Reset()
		patches.ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})

		testPlatformModeWriteConfirmFaultError(ctl)
		testPlatformModeNonUceFault(ctl)
		testPlatformModeUceFault(ctl)
		testNonPlatformMode(ctl)
	})
}

func testPlatformModeWriteConfirmFaultError(ctl *EventController) {
	convey.Convey("When platform mode and writeConfirmFaultAndWaitPlatResultFault returns error", func() {
		patches := gomonkey.ApplyPrivateMethod(ctl, "writeConfirmFaultAndWaitPlatResultFault",
			func(faults []*pb.FaultRank) ([]*pb.FaultRank, error) {
				return nil, errors.New("write confirm fault error")
			})
		defer patches.Reset()

		event, respCode, err := ctl.notifyFaultForRetryFaultCase(
			[]*pb.FaultRank{{RankId: "rank1"}}, []*pb.FaultRank{{RankId: "rank2"}})
		convey.So(event, convey.ShouldEqual, common.WriteConfirmFaultOrWaitResultFaultTimeoutEvent)
		convey.So(respCode, convey.ShouldEqual, common.WriteConfirmFaultOrWaitPlatResultFault)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testPlatformModeNonUceFault(ctl *EventController) {
	convey.Convey("When platform mode and non-UCE fault", func() {
		mockGetNodeRankIdsByFaultRanks := gomonkey.ApplyFuncReturn(common.GetNodeRankIdsByRankIds, []string{}, nil)
		defer mockGetNodeRankIdsByFaultRanks.Reset()
		patches := gomonkey.ApplyPrivateMethod(ctl, "writeConfirmFaultAndWaitPlatResultFault",
			func(faults []*pb.FaultRank) ([]*pb.FaultRank, error) {
				return []*pb.FaultRank{{RankId: "rank1"}}, nil
			})
		defer patches.Reset()

		patches.ApplyFunc(common.IsRetryFault, func(faults []*pb.FaultRank) bool {
			return false
		})

		patches.ApplyPrivateMethod(ctl, "normalFaultAssociateSameNodeRank",
			func() ([]*pb.FaultRank, []string) {
				return []*pb.FaultRank{{RankId: "rank2"}}, []string{"rank2"}
			})

		patches.ApplyFunc(common.GetPodMap, func(jobId string, ranks []string) (map[string]string, error) {
			return map[string]string{"rank2": "pod2"}, nil
		})

		patches.ApplyPrivateMethod(ctl, "getCtxAndSignalChan",
			func() (context.Context, chan *pb.ProcessManageSignal) {
				return context.Background(), make(chan *pb.ProcessManageSignal, 1)
			})

		patches.ApplyFunc(common.RetryWriteResetCM,
			func(jobName, namespace string, ranks []string,
				restartProcess bool, operation string) (*v1.ConfigMap, error) {
				return &v1.ConfigMap{Data: map[string]string{constant.ResetInfoCMDataKey: "test-data"}}, nil
			})

		event, respCode, err := ctl.notifyFaultForRetryFaultCase(
			[]*pb.FaultRank{{RankId: "rank1"}}, []*pb.FaultRank{{RankId: "rank2"}})
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testPlatformModeUceFault(ctl *EventController) {
	convey.Convey("When platform mode and UCE fault", func() {
		patches := gomonkey.ApplyPrivateMethod(ctl, "writeConfirmFaultAndWaitPlatResultFault",
			func(faults []*pb.FaultRank) ([]*pb.FaultRank, error) {
				return []*pb.FaultRank{{RankId: "rank1"}}, nil
			})
		defer patches.Reset()

		patches.ApplyFunc(common.IsRetryFault, func(faults []*pb.FaultRank) bool {
			return true
		})

		patches.ApplyPrivateMethod(ctl, "getCtxAndSignalChan",
			func() (context.Context, chan *pb.ProcessManageSignal) {
				return context.Background(), make(chan *pb.ProcessManageSignal, 1)
			})

		event, respCode, err := ctl.notifyFaultForRetryFaultCase(
			[]*pb.FaultRank{{RankId: "rank1"}}, []*pb.FaultRank{{RankId: "rank2"}})
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testNonPlatformMode(ctl *EventController) {
	convey.Convey("When non-platform mode", func() {
		ctl.jobInfo.PlatFormMode = false
		defer func() { ctl.jobInfo.PlatFormMode = true }()

		patches := gomonkey.ApplyPrivateMethod(ctl, "getCtxAndSignalChan",
			func() (context.Context, chan *pb.ProcessManageSignal) {
				return context.Background(), make(chan *pb.ProcessManageSignal, 1)
			})
		defer patches.Reset()

		event, respCode, err := ctl.notifyFaultForRetryFaultCase(
			[]*pb.FaultRank{{RankId: "rank1"}}, []*pb.FaultRank{{RankId: "rank2"}})
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleSendResult(t *testing.T) {
	convey.Convey("Test handleSendResult", t, func() {
		ctl := &EventController{}

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})
		defer patches.Reset()

		testKillMasterSignal(ctl)
		testErrorCase(ctl)
		testErrorSwitchNicCase(ctl)
		testNotifySuccessEvent(ctl)
		testChangeStrategyRetry(ctl)
		testChangeStrategyRecover(ctl)
		testChangeStrategyDump(ctl)
		testChangeStrategyExit(ctl)
		testUnsupportedStrategy(ctl)
		testChangeStrategyContinue(ctl)
	})
}

func testKillMasterSignal(ctl *EventController) {
	convey.Convey("When signal type is KillMasterSignalType", func() {
		signal := &pb.ProcessManageSignal{SignalType: constant.KillMasterSignalType}
		addedEvent := ""
		patches := gomonkey.ApplyPrivateMethod(ctl, "addEvent", func(ctl *EventController, event string) {
			addedEvent = event
		})
		defer patches.Reset()
		ctl.handleSendResult(signal, nil)
		convey.So(addedEvent, convey.ShouldEqual, common.FinishEvent)
	})
}

func testErrorCase(ctl *EventController) {
	convey.Convey("When error is not nil", func() {
		addedEvent := ""
		patches := gomonkey.ApplyPrivateMethod(ctl, "addEvent", func(ctl *EventController, event string) {
			addedEvent = event
		})
		defer patches.Reset()
		signal := &pb.ProcessManageSignal{}
		ctl.handleSendResult(signal, errors.New("test error"))
		convey.So(addedEvent, convey.ShouldEqual, common.NotifyFailEvent)
	})
}

func testErrorSwitchNicCase(ctl *EventController) {
	convey.Convey("When error is not nil, and switching nic", func() {
		addedEvent := ""
		patches := gomonkey.ApplyPrivateMethod(ctl, "addEvent", func(ctl *EventController, event string) {
			addedEvent = event
		})
		patches2 := gomonkey.ApplyPrivateMethod(ctl, "isSwitchingNic", func() bool {
			return true
		})
		defer patches.Reset()
		defer patches2.Reset()
		signal := &pb.ProcessManageSignal{}
		ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, 1)
		ctl.handleSendResult(signal, errors.New("test error"))
		res := <-ctl.switchNicResponse
		convey.So(res.Msg, convey.ShouldEqual, "om failed, send signal failed")
		convey.So(addedEvent, convey.ShouldEqual, common.NotifyFailEvent)
	})
}

func testNotifySuccessEvent(ctl *EventController) {
	convey.Convey("When signal type is not ChangeStrategySignalType", func() {
		signal := &pb.ProcessManageSignal{SignalType: constant.GlobalFaultSignalType}
		addedEvent := ""
		patches := gomonkey.ApplyPrivateMethod(ctl, "addEvent", func(ctl *EventController, event string) {
			addedEvent = event
		})
		defer patches.Reset()
		ctl.handleSendResult(signal, nil)
		convey.So(addedEvent, convey.ShouldEqual, common.NotifySuccessEvent)
	})
}

func testChangeStrategyRetry(ctl *EventController) {
	convey.Convey("When change strategy is ProcessRetryStrategyName", func() {
		signal := &pb.ProcessManageSignal{
			SignalType:     constant.ChangeStrategySignalType,
			ChangeStrategy: constant.ProcessRetryStrategyName,
		}
		addedEvent := ""
		patches := gomonkey.ApplyPrivateMethod(ctl, "addEvent", func(ctl *EventController, event string) {
			addedEvent = event
		})
		defer patches.Reset()
		ctl.handleSendResult(signal, nil)
		convey.So(addedEvent, convey.ShouldEqual, common.NotifyRetrySuccessEvent)
	})
}

func testChangeStrategyRecover(ctl *EventController) {
	convey.Convey("When change strategy is ProcessRecoverStrategyName", func() {
		signal := &pb.ProcessManageSignal{
			SignalType:     constant.ChangeStrategySignalType,
			ChangeStrategy: constant.ProcessRecoverStrategyName,
		}
		addedEvent := ""
		patches := gomonkey.ApplyPrivateMethod(ctl, "addEvent", func(ctl *EventController, event string) {
			addedEvent = event
		})
		defer patches.Reset()
		ctl.handleSendResult(signal, nil)
		convey.So(addedEvent, convey.ShouldEqual, common.NotifyRecoverSuccessEvent)
	})
}

func testChangeStrategyDump(ctl *EventController) {
	convey.Convey("When change strategy is ProcessDumpStrategyName", func() {
		signal := &pb.ProcessManageSignal{
			SignalType:     constant.ChangeStrategySignalType,
			ChangeStrategy: constant.ProcessDumpStrategyName,
		}
		addedEvent := ""
		patches := gomonkey.ApplyPrivateMethod(ctl, "addEvent", func(ctl *EventController, event string) {
			addedEvent = event
		})
		defer patches.Reset()
		ctl.handleSendResult(signal, nil)
		convey.So(addedEvent, convey.ShouldEqual, common.NotifyDumpSuccessEvent)
	})
}

func testChangeStrategyExit(ctl *EventController) {
	convey.Convey("When change strategy is ProcessExitStrategyName", func() {
		signal := &pb.ProcessManageSignal{
			SignalType:     constant.ChangeStrategySignalType,
			ChangeStrategy: constant.ProcessExitStrategyName,
		}
		addedEvent := ""
		patches := gomonkey.ApplyPrivateMethod(ctl, "addEvent", func(ctl *EventController, event string) {
			addedEvent = event
		})
		defer patches.Reset()
		ctl.handleSendResult(signal, nil)
		convey.So(addedEvent, convey.ShouldEqual, common.NotifyExitSuccessEvent)
	})
}

func testChangeStrategyContinue(ctl *EventController) {
	convey.Convey("When change strategy is ProcessContinueTrain", func() {
		signal := &pb.ProcessManageSignal{
			SignalType:     constant.ChangeStrategySignalType,
			ChangeStrategy: constant.ProcessContinueTrain,
		}
		addedEvent := ""
		patches := gomonkey.ApplyPrivateMethod(ctl, "addEvent", func(ctl *EventController, event string) {
			addedEvent = event
		})
		defer patches.Reset()
		ctl.handleSendResult(signal, nil)
		convey.So(addedEvent, convey.ShouldEqual, common.NotifyContinueSuccessEvent)
	})
}

func testUnsupportedStrategy(ctl *EventController) {
	convey.Convey("When change strategy is unsupported", func() {
		signal := &pb.ProcessManageSignal{
			SignalType:     constant.ChangeStrategySignalType,
			ChangeStrategy: "unsupported-strategy",
			JobId:          "test-job-id",
		}
		addedEvent := ""
		patches := gomonkey.ApplyPrivateMethod(ctl, "addEvent", func(ctl *EventController, event string) {
			addedEvent = event
		})
		defer patches.Reset()
		ctl.handleSendResult(signal, nil)
		convey.So(addedEvent, convey.ShouldEqual, "")
	})
}

func TestNotifyFaultForNormalFaultCase(t *testing.T) {
	convey.Convey("Test notifyFaultForNormalFaultCase", t, func() {
		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId:     "test-job-id",
				JobName:   "test-job",
				Namespace: "test-namespace",
			},
			uuid: "test-uuid",
		}

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {})
		defer patches.Reset()
		patches.ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})
		patches.ApplyPrivateMethod(ctl, "getCtxAndSignalChan",
			func() (context.Context, chan *pb.ProcessManageSignal) {
				return context.Background(), make(chan *pb.ProcessManageSignal, 1)
			})

		testPlatformModeWriteConfirmNormalFaultError(ctl)
		testPlatformModeSuccess(ctl)
		testNonPlatformModeSuccess(ctl)
	})
}

func testPlatformModeWriteConfirmNormalFaultError(ctl *EventController) {
	convey.Convey("When platform mode and writeConfirmFaultAndWaitPlatResultFault returns error", func() {
		ctl.jobInfo.PlatFormMode = true
		patches := gomonkey.ApplyPrivateMethod(ctl, "writeConfirmFaultAndWaitPlatResultFault",
			func(faults []*pb.FaultRank) ([]*pb.FaultRank, error) {
				return nil, errors.New("write confirm fault error")
			})
		defer patches.Reset()

		event, respCode, err := ctl.notifyFaultForNormalFaultCase(
			[]*pb.FaultRank{{RankId: "rank1"}}, []*pb.FaultRank{{RankId: "rank2"}})
		convey.So(event, convey.ShouldEqual, common.WriteConfirmFaultOrWaitResultFaultTimeoutEvent)
		convey.So(respCode, convey.ShouldEqual, common.WriteConfirmFaultOrWaitPlatResultFault)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testPlatformModeSuccess(ctl *EventController) {
	convey.Convey("When platform mode and all operations succeed", func() {
		mockGetNodeRankIdsByFaultRanks := gomonkey.ApplyFuncReturn(common.GetNodeRankIdsByRankIds, []string{}, nil)
		defer mockGetNodeRankIdsByFaultRanks.Reset()
		ctl.jobInfo.PlatFormMode = true
		patches := gomonkey.ApplyPrivateMethod(ctl, "writeConfirmFaultAndWaitPlatResultFault",
			func(faults []*pb.FaultRank) ([]*pb.FaultRank, error) {
				return []*pb.FaultRank{{RankId: "rank1"}}, nil
			})
		defer patches.Reset()

		patches.ApplyPrivateMethod(ctl, "updateCacheFaultAndPod",
			func() ([]*pb.FaultRank, []string, error) {
				return []*pb.FaultRank{{RankId: "rank1"}}, []string{"rank1"}, nil
			})

		patches.ApplyFunc(common.WriteResetInfoToCM,
			func(jobName, namespace string, ranks []string,
				restartProcess bool, operation string) (*v1.ConfigMap, error) {
				return &v1.ConfigMap{Data: map[string]string{constant.ResetInfoCMDataKey: "test-data"}}, nil
			})

		event, respCode, err := ctl.notifyFaultForNormalFaultCase(
			[]*pb.FaultRank{{RankId: "rank1"}}, []*pb.FaultRank{{RankId: "rank2"}})
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testNonPlatformModeSuccess(ctl *EventController) {
	convey.Convey("When non-platform mode and all operations succeed", func() {
		mockGetNodeRankIdsByFaultRanks := gomonkey.ApplyFuncReturn(common.GetNodeRankIdsByRankIds, []string{}, nil)
		defer mockGetNodeRankIdsByFaultRanks.Reset()
		ctl.jobInfo.PlatFormMode = false
		patches := gomonkey.ApplyPrivateMethod(ctl, "updateCacheFaultAndPod",
			func() ([]*pb.FaultRank, []string, error) {
				return []*pb.FaultRank{{RankId: "rank1"}}, []string{"rank1"}, nil
			})
		defer patches.Reset()

		patches.ApplyFunc(common.WriteResetInfoToCM,
			func(jobName, namespace string, ranks []string,
				restartProcess bool, operation string) (*v1.ConfigMap, error) {
				return &v1.ConfigMap{Data: map[string]string{constant.ResetInfoCMDataKey: "test-data"}}, nil
			})

		event, respCode, err := ctl.notifyFaultForNormalFaultCase(
			[]*pb.FaultRank{{RankId: "rank1"}}, []*pb.FaultRank{{RankId: "rank2"}})
		convey.So(event, convey.ShouldEqual, "")
		convey.So(respCode, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestHandleWaitRestartAllProcessPlatFormModeTrue(t *testing.T) {
	ctl := &EventController{
		jobInfo: common.JobBaseInfo{
			RecoverConfig: common.RecoverConfig{
				PlatFormMode: true,
			},
		},
		uuid: "testUuid",
	}

	patchGetCtx := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl), "getCtxAndScheduleResultChan",
		func(_ *EventController) (context.Context, <-chan struct{}) {
			return context.Background(), nil
		})
	defer patchGetCtx.Reset()

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
		result := make(map[string]string)
		patchRetryPatchPodGroupAnnotations := gomonkey.ApplyFunc(kube.RetryPatchPodGroupAnnotations,
			func(pgName, namespace string, retryTimes int, annotations map[string]string) (*v1beta1.PodGroup, error) {
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
			func(pgName, namespace string, retryTimes int, annotations map[string]string) (*v1beta1.PodGroup, error) {
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
			patches.ApplyPrivateMethod(retry.RetryProcessor, "JobHasFault",
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
