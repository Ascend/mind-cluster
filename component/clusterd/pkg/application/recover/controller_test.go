// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package recover a series of controller test function
package recover

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/pb"
)

const (
	keepAliveSeconds = 10
	sliceLength3     = 3
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
		convey.Convey("should dump when occur fault", func() {
			mockFunc1 := gomonkey.ApplyPrivateMethod(&EventController{}, "shouldDumpWhenOccurFault",
				func(*EventController) bool { return true })
			defer mockFunc1.Reset()
			convey.Convey("02-config strategies not contain plat strategy, should return err", func() {
				ctl.jobInfo.MindXConfigStrategies = []string{"exit"}
				result, respCode, err := ctl.handleNotifyWaitFaultFlushing()
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(respCode == common.ServerInnerError, convey.ShouldBeTrue)
				convey.So(result == "", convey.ShouldBeTrue)
			})
			convey.Convey("03-config strategies contain plat strategy, should return dump event", func() {
				ctl.jobInfo.MindXConfigStrategies = []string{"dump", "exit"}
				result, respCode, err := ctl.handleNotifyWaitFaultFlushing()
				convey.So(err, convey.ShouldBeNil)
				convey.So(respCode == common.OK, convey.ShouldBeTrue)
				convey.So(result == common.DumpForFaultEvent, convey.ShouldBeTrue)
			})
		})
		mockFunc1 := gomonkey.ApplyPrivateMethod(&EventController{}, "shouldDumpWhenOccurFault",
			func(*EventController) bool { return false })
		defer mockFunc1.Reset()
		convey.Convey("04-retry write reset cm fail, should return operate cm error", func() {
			mockFunc2 := gomonkey.ApplyFuncReturn(common.RetryWriteResetCM, nil, errors.New("mock error"))
			defer mockFunc2.Reset()
			result, respCode, err := ctl.handleNotifyWaitFaultFlushing()
			convey.So(err, convey.ShouldBeNil)
			convey.So(respCode == common.OperateConfigMapError, convey.ShouldBeTrue)
			convey.So(result == common.NotifyFinishEvent, convey.ShouldBeTrue)
		})
	})
}

func TestOnlySupportDumpStrategy(t *testing.T) {
	convey.Convey("onlySupportDumpStrategy", t, func() {
		ctl := EventController{}
		convey.Convey("01-ProcessRecoverEnable is false, should return false", func() {
			ctl.jobInfo.ProcessRecoverEnable = false
			convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeFalse)
		})
		ctl.jobInfo.ProcessRecoverEnable = true
		convey.Convey("02-cluster not support dump, should return false", func() {
			ctl.jobInfo.MindXConfigStrategies = []string{constant.ProcessRecoverStrategyName,
				constant.ProcessDumpStrategyName}
			convey.So(ctl.onlySupportDumpStrategy(), convey.ShouldBeFalse)
		})
		convey.Convey("03-cluster support exit, should return false", func() {
			ctl.jobInfo.MindXConfigStrategies = []string{constant.ProcessExitStrategyName}
			convey.So(ctl.onlySupportDumpStrategy(), convey.ShouldBeFalse)
		})
		ctl.jobInfo.MindXConfigStrategies = []string{constant.ProcessDumpStrategyName}
		convey.Convey("04-cluster support dump, not platform mode, should return true", func() {
			ctl.jobInfo.PlatFormMode = false
			convey.So(ctl.onlySupportDumpStrategy(), convey.ShouldBeTrue)
		})
		ctl.jobInfo.PlatFormMode = true
		convey.Convey("05-platform mode, platform strategy is not dump, should return false", func() {
			ctl.platStrategy = constant.ProcessExitStrategyName
			convey.So(ctl.onlySupportDumpStrategy(), convey.ShouldBeFalse)
		})
		convey.Convey("06-platform mode, platform strategy is dump, should return true", func() {
			ctl.platStrategy = constant.ProcessDumpStrategyName
			convey.So(ctl.onlySupportDumpStrategy(), convey.ShouldBeTrue)
		})
	})
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
		faultPod: map[string]string{},
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
			convey.So(ctl.cacheUceFault, convey.ShouldBeNil)
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
		convey.So(ctl.cacheUceFault, convey.ShouldHaveLength, 1)
		convey.So(ctl.cacheUceFault[0].RankId, convey.ShouldEqual, "rank2")
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
		ctl.cacheUceFault = []*pb.FaultRank{{RankId: "rank2", FaultType: constant.UceFaultType}}
		patch := gomonkey.ApplyFuncReturn(common.RetryWriteResetCM,
			&v1.ConfigMap{Data: make(map[string]string)}, nil)
		defer patch.Reset()
		ctl.reset()
		convey.So(ctl.faultFlushing, convey.ShouldBeFalse)
		convey.So(ctl.uuid, convey.ShouldBeEmpty)
		convey.So(ctl.latestStrategy, convey.ShouldHaveLength, 0)
		convey.So(ctl.faultPod, convey.ShouldHaveLength, 0)
		convey.So(ctl.cacheNormalFault, convey.ShouldHaveLength, 0)
		convey.So(ctl.cacheUceFault, convey.ShouldHaveLength, 0)
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
		gomonkey.ApplyFuncReturn(common.WriteResetInfoToCM, nil, nil)
		gomonkey.ApplyPrivateMethod(ctl, "handleNotifyWaitFaultFlushing",
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
	convey.Convey("Testing listenSendChannel", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSeconds, serviceCtx)
		gomonkey.ApplyFuncReturn(common.SendRetry, errors.New("error"))
		signal := &pb.ProcessManageSignal{}
		ctl.signalChan <- signal
		go ctl.listenSendChannel(nil)
		time.Sleep(time.Second)
		convey.ShouldEqual(len(ctl.signalChan), 0)
	})
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

func TestHandleNotifyDecidedStrategy(t *testing.T) {
	convey.Convey("Testing handleNotifyDecidedStrategy", t, func() {
		jobInfo := newJobInfoWithStrategy(nil)
		serviceCtx := context.Background()
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
			func(jobName, namespace string, faultRanks []string, operation string) (*v1.ConfigMap, error) {
				return &v1.ConfigMap{}, nil
			})
		defer patches.Reset()

		event, code, err := ctl.handleRestartAllProcess()
		convey.So(event, convey.ShouldEqual, common.NotifySuccessEvent)
		convey.So(code, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}
