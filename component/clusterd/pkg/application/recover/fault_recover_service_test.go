// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package recover a series of fault recover service test function
package recover

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
	"clusterd/pkg/interface/grpc/recover"
)

const (
	keepAliveSecond = 10
	length2         = 2
	fakeJobID1      = "fakeJobID1"
	fakeJobID2      = "fakeJobID2"
	fakeJobID       = "fakeJobID"
	testJobID1      = "testJobID1"
	testServerName1 = "server1"
	testServerName2 = "server2"
	testServerName3 = "server3"
	testPodUid1     = "podUid1"
	testPodUid2     = "podUid2"
	testPodUid3     = "podUid3"
	testNodeName1   = "node1"
	testNodeName2   = "node2"
	testRankId1     = "rank1"
	testRankId2     = "rank2"
	testRankId3     = "rank3"
	emptyString     = ""
)

func TestNotifyFaultInfoForJob(t *testing.T) {
	convey.Convey("Test notifyFaultInfoForJob", t, func() {
		svr := &FaultRecoverService{
			eventCtl: map[string]*EventController{
				fakeJobID1: {
					faultPod: map[string]string{},
					jobInfo: common.JobBaseInfo{
						RecoverConfig: common.RecoverConfig{GraceExit: false}},
				},
			},
		}
		convey.Convey("01-controller not exist, should not add event", func() {
			mockJob := fakeJobID2
			info := constant.JobFaultInfo{JobId: mockJob}
			svr.notifyFaultInfoForJob(info)
			convey.So(svr.eventCtl[mockJob], convey.ShouldBeNil)
		})
		convey.Convey("02-subHealthy fault and not graceExit, should not add event", func() {
			mockJob := fakeJobID1
			info := constant.JobFaultInfo{
				JobId:        mockJob,
				HealthyState: constant.SubHealthyState,
			}
			svr.notifyFaultInfoForJob(info)
			convey.So(svr.eventCtl[mockJob].healthState, convey.ShouldBeEmpty)
		})
		convey.Convey("03-unHealthy fault, should add event", func() {
			mockJob := fakeJobID1
			info := constant.JobFaultInfo{
				JobId:        mockJob,
				HealthyState: constant.UnHealthyState,
				FaultList: []constant.FaultRank{
					{RankId: "1", DoStepRetry: false, PodRank: "1", PodUid: "111"},
					{RankId: "2", DoStepRetry: true, PodRank: "2", PodUid: "222"},
				},
			}
			mockFunc := gomonkey.ApplyPrivateMethod(&EventController{}, "addEvent",
				func(*EventController, string) {})
			defer mockFunc.Reset()
			svr.notifyFaultInfoForJob(info)
			ctl := svr.eventCtl[mockJob]
			convey.So(ctl, convey.ShouldNotBeNil)
			convey.So(ctl.healthState == constant.UnHealthyState, convey.ShouldBeTrue)
			convey.So(len(ctl.cacheNormalFault) == 1, convey.ShouldBeFalse)
			convey.So(len(ctl.cacheRetryFault) == 1, convey.ShouldBeFalse)
		})
	})
}

func fakeService() *FaultRecoverService {
	ctx := context.Background()
	return NewFaultRecoverService(keepAliveSecond, ctx)
}

func fakeCommonBaseInfo() common.JobBaseInfo {
	return common.JobBaseInfo{
		JobId:     "fakeJobId",
		JobName:   "fakeJobName",
		PgName:    "fakePgName",
		Namespace: "fakeNamespace",
		RecoverConfig: common.RecoverConfig{
			ProcessRecoverEnable:  true,
			MindXConfigStrategies: []string{constant.ProcessExitStrategyName},
			PlatFormMode:          false,
		},
	}
}

func fakeClientInfo() *pb.ClientInfo {
	return &pb.ClientInfo{
		JobId: fakeJobID,
		Role:  "fakeRole",
	}
}

func TestInit(t *testing.T) {
	convey.Convey("Test Init", t, func() {
		ctx := context.Background()
		convey.Convey("case job init before", func() {
			s := fakeService()
			info := fakeClientInfo()
			s.initJob[info.JobId] = fakeCommonBaseInfo()
			res, err := s.Init(ctx, info)
			convey.So(err, convey.ShouldBeNil)
			convey.So(res.Code, convey.ShouldEqual, int32(common.OK))
		})
		convey.Convey("case job not init before", func() {
			s := fakeService()
			info := fakeClientInfo()
			path1 := gomonkey.ApplyFuncReturn(common.ChangeProcessRecoverEnableMode, nil, nil)
			defer path1.Reset()
			path2 := gomonkey.ApplyFuncReturn(getJobBaseInfo, fakeCommonBaseInfo(), common.OK, nil)
			defer path2.Reset()
			res, err := s.Init(ctx, info)
			convey.So(err, convey.ShouldBeNil)
			convey.So(res.Code, convey.ShouldEqual, int32(common.OK))
			convey.So(len(s.initJob), convey.ShouldEqual, 1)
		})
	})
}

func TestRegister(t *testing.T) {
	convey.Convey("Test Register", t, func() {
		ctx := context.Background()
		convey.Convey("case job registered before", func() {
			s := fakeService()
			info := fakeClientInfo()
			s.eventCtl[info.JobId] = &EventController{}
			res, err := s.Register(ctx, info)
			convey.So(err, convey.ShouldBeNil)
			convey.So(res.Code, convey.ShouldEqual, int32(common.OK))
		})

		convey.Convey("case job not registered", func() {
			s := fakeService()
			info := fakeClientInfo()
			path1 := gomonkey.ApplyPrivateMethod(s, "preRegistry", func(req *pb.ClientInfo) (common.RespCode, error) {
				return common.OK, nil
			})
			defer path1.Reset()
			path2 := gomonkey.ApplyFuncReturn(common.ChangeProcessRecoverEnableMode, nil, nil)
			defer path2.Reset()
			convey.Convey("case has init before", func() {
				s.initJob[info.JobId] = fakeCommonBaseInfo()
				res, err := s.Register(ctx, info)
				convey.So(err, convey.ShouldBeNil)
				convey.So(res.Code, convey.ShouldEqual, int32(common.OK))
			})
			convey.Convey("case not init before", func() {
				delete(s.initJob, info.JobId)
				res, err := s.Register(ctx, info)
				convey.So(err, convey.ShouldBeNil)
				convey.So(res.Code, convey.ShouldEqual, int32(common.UnInit))
			})
		})
	})
}

func TestReportProcessFault(t *testing.T) {
	convey.Convey("Test ReportProcessFault", t, func() {
		ctx := context.Background()
		convey.Convey("case job not registered", func() {
			s := fakeService()
			info := fakeClientInfo()
			res, err := s.ReportProcessFault(ctx, &pb.ProcessFaultRequest{
				JobId:      info.JobId,
				FaultRanks: nil,
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(res.Code, convey.ShouldEqual, int32(common.UnRegistry))
		})
		convey.Convey("case job registered", func() {
			s := fakeService()
			info := fakeClientInfo()
			ctl := NewEventController(fakeCommonBaseInfo(), keepAliveSecond, ctx)
			s.eventCtl[info.JobId] = ctl
			patch1 := gomonkey.ApplyFuncReturn(common.LabelFaultPod, nil, nil)
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(giveSoftFault2FaultCenter, func(jobId string, faults []*pb.FaultRank) {
				return
			})
			defer patch2.Reset()
			convey.Convey("case uce fault", func() {
				res, err := s.ReportProcessFault(ctx, &pb.ProcessFaultRequest{
					JobId:      info.JobId,
					FaultRanks: []*pb.FaultRank{{RankId: "8", FaultType: "0"}},
				})
				convey.So(err, convey.ShouldBeNil)
				convey.So(res.Code, convey.ShouldEqual, int32(common.OK))
				convey.So(len(ctl.events), convey.ShouldEqual, 0)
			})
			convey.Convey("case normal fault", func() {
				res, err := s.ReportProcessFault(ctx, &pb.ProcessFaultRequest{
					JobId:      info.JobId,
					FaultRanks: []*pb.FaultRank{{RankId: "8", FaultType: "1"}},
				})
				convey.So(err, convey.ShouldBeNil)
				convey.So(res.Code, convey.ShouldEqual, int32(common.OK))
				convey.So(len(ctl.events), convey.ShouldEqual, 1)
			})
		})
	})
}

func TestReportRecoverStatus(t *testing.T) {
	convey.Convey("Test ReportRecoverStatus", t, func() {
		info := fakeClientInfo()
		ctx := context.Background()
		convey.Convey("case job not registered", func() {
			s := fakeService()
			res, err := s.ReportRecoverStatus(ctx, &pb.RecoverStatusRequest{
				JobId: info.JobId,
				Status: &pb.Status{
					Code: int32(common.OK),
					Info: "",
				},
				Strategy: "",
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(res.Code, convey.ShouldEqual, int32(common.UnRegistry))
		})
		convey.Convey("case job registered", func() {
			s := fakeService()
			ctl := &EventController{reportStatusChan: make(chan *pb.RecoverStatusRequest, 1)}
			s.eventCtl[info.JobId] = ctl
			res, err := s.ReportRecoverStatus(ctx, &pb.RecoverStatusRequest{
				JobId: info.JobId,
				Status: &pb.Status{
					Code: int32(common.OK),
					Info: "",
				},
				Strategy: "",
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(res.Code, convey.ShouldEqual, int32(common.OK))
			convey.So(len(ctl.reportStatusChan), convey.ShouldEqual, 1)
		})
	})
}

func TestReportRecoverStrategy(t *testing.T) {
	convey.Convey("Test ReportRecoverStrategy", t, func() {
		info := fakeClientInfo()
		ctx := context.Background()
		convey.Convey("case job not registered", func() {
			s := fakeService()
			res, err := s.ReportRecoverStrategy(ctx, &pb.RecoverStrategyRequest{
				JobId:      info.JobId,
				FaultRanks: nil,
				Strategies: nil,
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(res.Code, convey.ShouldEqual, int32(common.UnRegistry))
		})
		convey.Convey("case job registered", func() {
			s := fakeService()
			ctl := &EventController{reportRecoverStrategyChan: make(chan *pb.RecoverStrategyRequest, 1)}
			s.eventCtl[info.JobId] = ctl
			res, err := s.ReportRecoverStrategy(ctx, &pb.RecoverStrategyRequest{
				JobId:      info.JobId,
				FaultRanks: nil,
				Strategies: nil,
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(res.Code, convey.ShouldEqual, int32(common.OK))
			convey.So(len(ctl.reportRecoverStrategyChan), convey.ShouldEqual, 1)
		})
	})
}

func TestReportStopComplete(t *testing.T) {
	convey.Convey("Test ReportStopComplete", t, func() {
		info := fakeClientInfo()
		ctx := context.Background()
		convey.Convey("case job not registered", func() {
			s := fakeService()
			res, err := s.ReportStopComplete(ctx, &pb.StopCompleteRequest{
				JobId: info.JobId,
				Status: &pb.Status{
					Code: int32(common.OK),
					Info: "",
				},
				FaultRanks: []*pb.FaultRank{
					&pb.FaultRank{
						RankId:    "8",
						FaultType: "1",
					},
				},
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(res.Code, convey.ShouldEqual, int32(common.UnRegistry))
		})
		convey.Convey("case job registered", func() {
			s := fakeService()
			ctl := &EventController{reportStopCompleteChan: make(chan *pb.StopCompleteRequest, 1)}
			s.eventCtl[info.JobId] = ctl
			res, err := s.ReportStopComplete(ctx, &pb.StopCompleteRequest{
				JobId: info.JobId,
				Status: &pb.Status{
					Code: int32(common.OK),
					Info: "",
				},
				FaultRanks: []*pb.FaultRank{
					&pb.FaultRank{
						RankId:    "8",
						FaultType: "1",
					},
				},
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(res.Code, convey.ShouldEqual, int32(common.OK))
			convey.So(len(ctl.reportStopCompleteChan), convey.ShouldEqual, 1)
		})
	})
}

func TestSubscribeProcessManageSignal(t *testing.T) {
	convey.Convey("Test SubscribeProcessManageSignal", t, func() {
		info := fakeClientInfo()
		convey.Convey("case job not registered", func() {
			s := fakeService()
			err := s.SubscribeProcessManageSignal(info, nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("case job registered", func() {
			s := fakeService()
			patch := gomonkey.ApplyPrivateMethod(&EventController{}, "listenSendChannel",
				func(stream pb.Recover_SubscribeProcessManageSignalServer) {
					return
				})
			defer patch.Reset()
			s.eventCtl[info.JobId] = &EventController{}
			err := s.SubscribeProcessManageSignal(info, nil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestNewFaultRecoverService(t *testing.T) {
	convey.Convey("Test NewFaultRecoverService", t, func() {
		s := fakeService()
		convey.So(len(s.eventCtl), convey.ShouldEqual, 0)
		convey.So(len(s.initJob), convey.ShouldEqual, 0)
		convey.ShouldImplement(s, (*pb.RecoverServer)(nil))
	})
}

func TestDeleteJob(t *testing.T) {
	convey.Convey("Testing DeleteJob", t, func() {
		testDeleteJobCase1()
		testDeleteJobCase2()
		testDeleteJobCase3()
		testDeleteJobCase4()
		testDeleteJobCase5()
		testDeleteJobCase6()
	})
}

func testDeleteJobCase1() {
	convey.Convey("eventCtl is nil", func() {
		service := fakeService()
		service.eventCtl = nil
		service.DeleteJob(fakeJobID1)
		convey.So(service.eventCtl, convey.ShouldHaveLength, 0)
	})
}

func testDeleteJobCase2() {
	convey.Convey("jobId does not exist in eventCtl", func() {
		service := fakeService()
		service.eventCtl[fakeJobID1] = nil
		service.DeleteJob("non-existent-job")
		convey.So(service.eventCtl, convey.ShouldHaveLength, 1)
	})
}

func testDeleteJobCase3() {
	convey.Convey("controller is nil", func() {
		service := fakeService()
		service.eventCtl[fakeJobID1] = nil
		service.DeleteJob(fakeJobID1)
		convey.So(service.eventCtl, convey.ShouldHaveLength, 1)
	})
}

func testDeleteJobCase4() {
	convey.Convey("normal delete", func() {
		jobInfo := fakeCommonBaseInfo()
		jobInfo.PlatFormMode = true
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSecond, serviceCtx)

		service := fakeService()
		service.eventCtl[fakeJobID1] = ctl
		service.initJob[fakeJobID1] = jobInfo
		patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl), "reset", func(*EventController, bool) {
			return
		})
		defer patches.Reset()

		service.DeleteJob(fakeJobID1)
		convey.So(service.eventCtl, convey.ShouldNotContainKey, fakeJobID1)
		convey.So(service.initJob, convey.ShouldNotContainKey, fakeJobID1)
	})
}

func testDeleteJobCase5() {
	convey.Convey("jobId exists in eventCtl, but controller is not nil", func() {
		jobInfo := fakeCommonBaseInfo()
		jobInfo.PlatFormMode = true
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSecond, serviceCtx)

		service := fakeService()
		service.eventCtl[jobInfo.JobId] = ctl
		service.initJob[jobInfo.JobId] = jobInfo

		patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl), "reset", func(*EventController, bool) {
			return
		})
		defer patches.Reset()

		service.DeleteJob(jobInfo.JobId)
		convey.So(service.eventCtl, convey.ShouldNotContainKey, jobInfo.JobId)
		convey.So(service.initJob, convey.ShouldNotContainKey, jobInfo.JobId)
	})
}

func testDeleteJobCase6() {
	convey.Convey("jobId exists in eventCtl, but initJob does not contain jobId", func() {
		jobInfo := fakeCommonBaseInfo()
		jobInfo.PlatFormMode = true
		serviceCtx := context.Background()
		ctl := NewEventController(jobInfo, keepAliveSecond, serviceCtx)

		service := fakeService()
		service.eventCtl[jobInfo.JobId] = ctl

		patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ctl), "reset", func(*EventController, bool) {
			return
		})
		defer patches.Reset()

		service.DeleteJob(jobInfo.JobId)
		convey.So(service.eventCtl, convey.ShouldNotContainKey, jobInfo.JobId)
		convey.So(service.initJob, convey.ShouldNotContainKey, jobInfo.JobId)
	})
}

func TestGiveSoftFault2FaultCenter(t *testing.T) {
	convey.Convey("Testing giveSoftFault2FaultCenter", t, func() {
		jobId := "fakeJobId"
		faults := []*pb.FaultRank{
			{RankId: "rank1"},
			{RankId: "rank2"},
		}

		patches := gomonkey.ApplyFunc(faultmanager.CallbackForReportRetryInfo,
			func(infos []constant.ReportRecoverInfo) {
				for i := 0; i < len(infos) && i < len(faults); i++ {
					convey.So(infos[i].JobId, convey.ShouldEqual, jobId)
					convey.So(infos[i].Rank, convey.ShouldEqual, faults[i].RankId)
				}
			})
		defer patches.Reset()

		giveSoftFault2FaultCenter(jobId, faults)
	})
}

func TestDealWithJobFaultInfo(t *testing.T) {
	convey.Convey("Testing dealWithJobFaultInfo", t, func() {
		jobFaultInfoList := []constant.JobFaultInfo{
			{JobId: fakeJobID1, FaultList: []constant.FaultRank{}},
			{JobId: fakeJobID1, FaultList: []constant.FaultRank{}},
		}

		service := fakeService()

		patches := gomonkey.ApplyFunc(service.notifyFaultInfoForJob, func(jobFaultInfo constant.JobFaultInfo) {
			convey.So(jobFaultInfo.JobId, convey.ShouldEqual, jobFaultInfoList[0].JobId)
		})
		defer patches.Reset()

		service.dealWithJobFaultInfo(jobFaultInfoList)
	})
}

func TestCheckFault(t *testing.T) {
	convey.Convey("Testing checkFault", t, func() {
		service := fakeService()
		patches := gomonkey.ApplyFunc(service.dealWithJobFaultInfo, func(jobFaultInfoList []constant.JobFaultInfo) {
			convey.So(jobFaultInfoList, convey.ShouldHaveLength, 1)
		})
		defer patches.Reset()
		info := map[string]constant.JobFaultInfo{
			fakeJobID1: {JobId: fakeJobID1, FaultList: []constant.FaultRank{{}}},
			fakeJobID2: {JobId: fakeJobID2, FaultList: []constant.FaultRank{{}}},
		}
		service.checkFault(info)

		faultmanager.GlobalFaultProcessCenter = nil
		service.checkFault(info)
	})
}

func TestServeJobNum(t *testing.T) {
	convey.Convey("Testing serveJobNum", t, func() {
		service := fakeService()
		service.eventCtl[fakeJobID1] = &EventController{}
		service.eventCtl[fakeJobID2] = &EventController{}

		num := service.serveJobNum()
		convey.So(num, convey.ShouldEqual, length2)
	})
}

func TestPreRegistry(t *testing.T) {
	convey.Convey("Testing preRegistry", t, func() {
		service := fakeService()
		req := &pb.ClientInfo{JobId: "non-existent-job"}
		code, err := service.preRegistry(req)
		convey.So(code, convey.ShouldEqual, common.JobNotExist)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestGetJobBaseInfoNormal(t *testing.T) {
	patch1 := gomonkey.ApplyFunc(podgroup.GetPGFromCacheOrPod,
		func(jobId string) (string, string, string) {
			return "testJobName", "testPgName", "testNamespace"
		})
	defer patch1.Reset()

	patch2 := gomonkey.ApplyFunc(common.GetRecoverBaseInfo,
		func(pgName, namespace string) (common.RecoverConfig, common.RespCode, error) {
			return common.RecoverConfig{
				ProcessRecoverEnable: true,
			}, common.OK, nil
		})
	defer patch2.Reset()

	jobId := "testJobId"
	info, code, err := getJobBaseInfo(jobId)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if code != common.OK {
		t.Errorf("Expected response code %d, but got %d", common.OK, code)
	}
	if info.JobId != jobId {
		t.Errorf("Expected jobId %s, but got %s", jobId, info.JobId)
	}
}

func TestGetJobBaseInfoGetPGFromCacheError(t *testing.T) {
	patch := gomonkey.ApplyFunc(podgroup.GetPGFromCacheOrPod,
		func(jobId string) (string, string, string) {
			return "", "", ""
		})
	defer patch.Reset()

	jobId := "testJobId"
	_, code, err := getJobBaseInfo(jobId)
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}
	if code != common.OperatePodGroupError {
		t.Errorf("Expected response code %d, but got %d", common.OperatePodGroupError, code)
	}
}

func TestGetJobBaseInfoGetRecoverBaseInfoError(t *testing.T) {
	patch1 := gomonkey.ApplyFunc(podgroup.GetPGFromCacheOrPod,
		func(jobId string) (string, string, string) {
			return "testJobName", "testPgName", "testNamespace"
		})
	defer patch1.Reset()

	patch2 := gomonkey.ApplyFunc(common.GetRecoverBaseInfo,
		func(pgName, namespace string) (common.RecoverConfig, common.RespCode, error) {
			return common.RecoverConfig{}, common.OperatePodGroupError, errors.New("test error")
		})
	defer patch2.Reset()

	jobId := "testJobId"
	_, code, err := getJobBaseInfo(jobId)
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}
	if code != common.OperatePodGroupError {
		t.Errorf("Expected response code %d, but got %d", common.OperatePodGroupError, code)
	}
}

func TestGetJobBaseInfoProcessRecoverEnableOff(t *testing.T) {
	patch1 := gomonkey.ApplyFunc(podgroup.GetPGFromCacheOrPod,
		func(jobId string) (string, string, string) {
			return "testJobName", "testPgName", "testNamespace"
		})
	defer patch1.Reset()

	patch2 := gomonkey.ApplyFunc(common.GetRecoverBaseInfo,
		func(pgName, namespace string) (common.RecoverConfig, common.RespCode, error) {
			return common.RecoverConfig{
				ProcessRecoverEnable: false,
			}, common.OK, nil
		})
	defer patch2.Reset()

	jobId := "testJobId"
	_, code, err := getJobBaseInfo(jobId)
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}
	if code != common.ProcessRecoverEnableOff {
		t.Errorf("Expected response code %d, but got %d", common.ProcessRecoverEnableOff, code)
	}
}

func TestGetFaultReason(t *testing.T) {
	convey.Convey("Testing getFaultReason", t, func() {
		faults1 := []*pb.FaultRank{&pb.FaultRank{
			RankId:    "0",
			FaultType: "1",
		}}
		faults2 := []*pb.FaultRank{&pb.FaultRank{
			RankId:    "0",
			FaultType: "0",
		}, &pb.FaultRank{
			RankId:    "1",
			FaultType: "1",
		}}
		faults3 := []*pb.FaultRank{&pb.FaultRank{
			RankId:    "0",
			FaultType: "0",
		}}
		convey.So(getFaultReason(faults1), convey.ShouldEqual, normalFaultValue)
		convey.So(getFaultReason(faults2), convey.ShouldEqual, normalFaultValue)
		convey.So(getFaultReason(faults3), convey.ShouldEqual, retryFaultValue)
	})
}

func TestFaultRecoverServiceHealthCheck(t *testing.T) {
	convey.Convey("Test FaultRecoverService HealthCheck", t, func() {
		s := fakeService()
		info := fakeClientInfo()
		ctx := context.Background()

		convey.Convey("case: receive healthcheck", func() {
			ctl := NewEventController(fakeCommonBaseInfo(), keepAliveSecond, ctx)
			s.eventCtl[info.JobId] = ctl

			resp, err := s.HealthCheck(ctx, info)

			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(resp.Code, convey.ShouldEqual, int32(common.OK))
		})
	})
}

func TestCatchAndSetExceptionInfo_Panic(t *testing.T) {
	ctl := &EventController{}
	var code int32 = int32(common.OK)
	var info string = "original info"

	func() {
		defer catchAndSetExceptionInfo(&code, &info, ctl)
		panic("channel closed")
	}()

	if code != int32(common.ServerInnerError) {
		t.Errorf("expect code is %d，actual is %d", common.ServerInnerError, code)
	}
}

type testSubHealthyCase struct {
	name                    string
	controller              *EventController
	faultInfo               constant.JobFaultInfo
	mockFramework           string
	mockOnlySupportDump     bool
	state                   string
	expectedResult          bool
	expectHotSwitchDisabled bool
}

func buildTestCases1() []testSubHealthyCase {
	return []testSubHealthyCase{
		{name: "not sub healthy state",
			controller:              &EventController{jobInfo: common.JobBaseInfo{}},
			faultInfo:               constant.JobFaultInfo{HealthyState: constant.UnHealthyState},
			expectedResult:          false,
			expectHotSwitchDisabled: false,
		}, {name: "hotswitch with non-pytorch framework",
			controller: &EventController{
				jobInfo: common.JobBaseInfo{RecoverConfig: common.RecoverConfig{HotSwitch: true}, Framework: "tensorflow"},
			},
			faultInfo:               constant.JobFaultInfo{HealthyState: constant.SubHealthyState},
			expectedResult:          true,
			expectHotSwitchDisabled: true,
		}, {
			name: "hotswitch with only master fault",
			controller: &EventController{
				jobInfo: common.JobBaseInfo{
					RecoverConfig: common.RecoverConfig{HotSwitch: true}, Framework: constant.PtFramework,
				},
			},
			faultInfo: constant.JobFaultInfo{HealthyState: constant.SubHealthyState,
				FaultList: []constant.FaultRank{
					{RankId: "0", PodUid: "0", PodRank: "0"}}},
			expectedResult:          true,
			expectHotSwitchDisabled: false,
		}, {
			name: "hotswitch with pytorch framework and normal fault pods count",
			controller: &EventController{
				jobInfo: common.JobBaseInfo{
					RecoverConfig: common.RecoverConfig{HotSwitch: true}, Framework: constant.PtFramework,
				},
			},
			faultInfo: constant.JobFaultInfo{HealthyState: constant.SubHealthyState,
				FaultList: []constant.FaultRank{{PodUid: "0", PodRank: "0"}, {PodUid: "1", PodRank: "1"}}},
			expectedResult:          false,
			expectHotSwitchDisabled: false,
		}, {
			name: "hotswitch with pytorch framework and normal fault pods count",
			controller: &EventController{
				jobInfo: common.JobBaseInfo{
					RecoverConfig: common.RecoverConfig{HotSwitch: true}, Framework: constant.PtFramework,
				},
			},
			faultInfo: constant.JobFaultInfo{HealthyState: constant.SubHealthyState,
				FaultList: []constant.FaultRank{{PodUid: "0", PodRank: "0"}, {PodUid: "1", PodRank: "1"}}},
			expectedResult:          false,
			expectHotSwitchDisabled: false,
		},
	}
}

func buildTestCases2() []testSubHealthyCase {
	return []testSubHealthyCase{
		{name: "sub healthy without hotswitch and graceExit false",
			controller: &EventController{
				jobInfo: common.JobBaseInfo{
					RecoverConfig: common.RecoverConfig{HotSwitch: false, GraceExit: false}, Framework: constant.PtFramework,
				},
			},
			faultInfo:               constant.JobFaultInfo{HealthyState: constant.SubHealthyState},
			expectedResult:          true,
			expectHotSwitchDisabled: false,
		}, {name: "sub healthy without hotswitch, graceExit true and not only dump strategy",
			controller: &EventController{
				jobInfo: common.JobBaseInfo{
					RecoverConfig: common.RecoverConfig{HotSwitch: false, GraceExit: true}, Framework: constant.PtFramework,
				},
			},
			faultInfo:               constant.JobFaultInfo{HealthyState: constant.SubHealthyState},
			mockOnlySupportDump:     false,
			expectedResult:          true,
			expectHotSwitchDisabled: false,
		}, {name: "sub healthy without hotswitch, graceExit true and only dump strategy",
			controller: &EventController{
				jobInfo: common.JobBaseInfo{
					RecoverConfig: common.RecoverConfig{HotSwitch: false, GraceExit: true}, Framework: constant.PtFramework,
				},
			},
			faultInfo:               constant.JobFaultInfo{HealthyState: constant.SubHealthyState},
			mockOnlySupportDump:     true,
			expectedResult:          false,
			expectHotSwitchDisabled: false,
		}, {
			name: "hotswitch with too much fault pod ",
			controller: &EventController{
				jobInfo: common.JobBaseInfo{
					RecoverConfig: common.RecoverConfig{HotSwitch: true}, Framework: constant.PtFramework,
				},
			},
			faultInfo: constant.JobFaultInfo{HealthyState: constant.SubHealthyState,
				FaultList: []constant.FaultRank{{PodUid: "1", PodRank: "1"}, {PodUid: "2", PodRank: "2"},
					{PodUid: "3", PodRank: "3"}, {PodUid: "4", PodRank: "4"}, {PodUid: "5", PodRank: "5"},
					{PodUid: "6", PodRank: "6"}, {PodUid: "7", PodRank: "7"}, {PodUid: "8", PodRank: "8"},
					{PodUid: "9", PodRank: "9"}, {PodUid: "10", PodRank: "10"}, {PodUid: "11", PodRank: "11"}}},
			expectedResult:          true,
			expectHotSwitchDisabled: true,
		},
	}
}
func buildTestCases3() []testSubHealthyCase {
	return []testSubHealthyCase{
		{name: "skip when state machine is in annother state",
			controller: &EventController{
				jobInfo: common.JobBaseInfo{
					Framework: constant.PtFramework,
				},
			},
			faultInfo:               constant.JobFaultInfo{HealthyState: constant.SubHealthyState},
			state:                   common.FaultOccurEvent,
			expectedResult:          true,
			expectHotSwitchDisabled: false,
		},
		{
			name: "should return false when hotswitch with mindspore framework and normal fault pods count",
			controller: &EventController{
				jobInfo: common.JobBaseInfo{
					RecoverConfig: common.RecoverConfig{HotSwitch: true}, Framework: constant.MsFramework,
				},
			},
			faultInfo: constant.JobFaultInfo{HealthyState: constant.SubHealthyState,
				FaultList: []constant.FaultRank{{PodUid: "0", PodRank: "0"}, {PodUid: "1", PodRank: "1"}}},
			expectedResult:          false,
			expectHotSwitchDisabled: false,
		},
	}
}

func TestSkipHandleSubHealthyFaults(t *testing.T) {
	var tests []testSubHealthyCase
	tests = append(tests, buildTestCases1()...)
	tests = append(tests, buildTestCases2()...)
	tests = append(tests, buildTestCases3()...)

	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			patch := gomonkey.NewPatches()
			defer patch.Reset()
			patch.ApplyPrivateMethod(tt.controller, "onlySupportDumpStrategy", func() bool {
				return tt.mockOnlySupportDump
			})
			tt.controller.state = common.NewStateMachine(common.InitState, nil)
			if tt.state != "" {
				tt.controller.state = common.NewStateMachine(tt.state, nil)
			}
			result := (&FaultRecoverService{}).skipHandleSubHealthyFaults(tt.controller, &tt.faultInfo)
			convey.So(result, convey.ShouldEqual, tt.expectedResult)
			if tt.expectHotSwitchDisabled {
				convey.So(tt.controller.jobInfo.HotSwitch, convey.ShouldBeFalse)
				convey.So(tt.controller.jobInfo.SubHealthyStrategy, convey.ShouldEqual, constant.SubHealthyIngore)
			}
		})
	}
}

func TestGetGrpcFormatFaults(t *testing.T) {
	convey.Convey("TestGetGrpcFormatFaults - Normal Cases", t, func() {
		svr := &FaultRecoverService{}
		ctl := &EventController{
			faultPod:       map[string]string{"0": ""},
			jobInfo:        common.JobBaseInfo{JobId: "test-job"},
			latestStrategy: []string{constant.ScaleInStrategyName},
		}
		convey.Convey("01-Normal case", func() {
			faultInfo := constant.JobFaultInfo{
				HealthyState: constant.UnHealthyState,
				FaultList: []constant.FaultRank{
					{PodUid: "", PodRank: "0", RankId: "rank1"},
					{PodUid: "pod3", PodRank: "", RankId: "rank3"},
					{PodUid: "pod1", PodRank: "1", RankId: "rank1", FaultLevel: constant.SubHealthFault},
					{PodUid: "pod0", PodRank: "0", RankId: "rank0"},
					{PodUid: "pod2", PodRank: "2", RankId: "rank2",
						FaultCode: constant.UceFaultCode, DoStepRetry: true},
					{PodUid: "pod4", PodRank: "4", RankId: "rank4",
						FaultCode: constant.HcclRetryFaultCode, DoStepRetry: true},
					{PodUid: "pod5", PodRank: "5", RankId: "rank5"},
				},
			}

			result := svr.getGrpcFormatFaults(faultInfo, ctl)
			expectedResultLength := 3
			convey.So(len(result), convey.ShouldEqual, expectedResultLength)
			convey.So(result[0].RankId, convey.ShouldEqual, "rank2")
			convey.So(result[0].FaultType, convey.ShouldEqual, constant.UceFaultType)
			convey.So(result[1].RankId, convey.ShouldEqual, "rank4")
			convey.So(result[1].FaultType, convey.ShouldEqual, constant.HcclFaultType)
			convey.So(result[2].RankId, convey.ShouldEqual, "rank5")
			convey.So(result[2].FaultType, convey.ShouldEqual, constant.NormalFaultType)
		})
	})
}

type getJobServerInfosTestCase struct {
	name                string
	jobId               string
	mockPods            map[string]*constant.SimplePodInfo
	expectedServerMap   map[string]bool
	expectedPodToServer map[string]string
}

func buildGetJobServerInfosTestCases() []getJobServerInfosTestCase {
	return []getJobServerInfosTestCase{
		{
			name:                "should return empty maps when no pods exist",
			jobId:               testJobID1,
			mockPods:            map[string]*constant.SimplePodInfo{},
			expectedServerMap:   map[string]bool{},
			expectedPodToServer: map[string]string{},
		},
		{
			name:  "should return correct maps when pods have node names",
			jobId: testJobID1,
			mockPods: map[string]*constant.SimplePodInfo{
				testPodUid1: {PodUid: testPodUid1, NodeName: testNodeName1},
				testPodUid2: {PodUid: testPodUid2, NodeName: testNodeName2},
			},
			expectedServerMap: map[string]bool{
				testNodeName1: true,
				testNodeName2: true,
			},
			expectedPodToServer: map[string]string{
				testPodUid1: testNodeName1,
				testPodUid2: testNodeName2,
			},
		},
		{
			name:  "should skip pods without node names",
			jobId: testJobID1,
			mockPods: map[string]*constant.SimplePodInfo{
				testPodUid1: {PodUid: testPodUid1, NodeName: testNodeName1},
				testPodUid2: {PodUid: testPodUid2, NodeName: emptyString},
				testPodUid3: {PodUid: testPodUid3, NodeName: testNodeName2},
			},
			expectedServerMap: map[string]bool{
				testNodeName1: true,
				testNodeName2: true,
			},
			expectedPodToServer: map[string]string{
				testPodUid1: testNodeName1,
				testPodUid3: testNodeName2,
			},
		},
	}
}

func TestGetJobServerInfos(t *testing.T) {
	testCases := buildGetJobServerInfosTestCases()
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			patch := gomonkey.ApplyFuncReturn(pod.GetSimplePodByJobId, tc.mockPods)
			defer patch.Reset()
			serverMap, podToServerMap := getJobServerInfos(tc.jobId)
			convey.So(serverMap, convey.ShouldResemble, tc.expectedServerMap)
			convey.So(podToServerMap, convey.ShouldResemble, tc.expectedPodToServer)
		})
	}
}

type preHandleFaultInfoTestCase struct {
	name                   string
	jobId                  string
	faultInfo              constant.JobFaultInfo
	mockServerMap          map[string]bool
	mockPodToServerMap     map[string]string
	expectedFaultDeviceLen int
	expectedFaultListLen   int
}

func buildPreHandleFaultInfoTestCases1() []preHandleFaultInfoTestCase {
	return []preHandleFaultInfoTestCase{
		{
			name:  "should return early when fault list is empty",
			jobId: testJobID1,
			faultInfo: constant.JobFaultInfo{
				FaultList: []constant.FaultRank{},
			},
			expectedFaultDeviceLen: 0,
			expectedFaultListLen:   0,
		},
		{
			name:  "should filter fault device not in current server list",
			jobId: testJobID1,
			faultInfo: constant.JobFaultInfo{
				FaultList: []constant.FaultRank{
					{PodUid: testPodUid1, RankId: testRankId1},
				},
				FaultDevice: []constant.FaultDevice{
					{ServerName: testServerName1},
					{ServerName: testServerName3},
				},
			},
			mockServerMap: map[string]bool{
				testServerName1: true,
				testServerName2: true,
			},
			mockPodToServerMap: map[string]string{
				testPodUid1: testServerName1,
			},
			expectedFaultDeviceLen: 1,
			expectedFaultListLen:   1,
		},
	}
}

func buildPreHandleFaultInfoTestCases2() []preHandleFaultInfoTestCase {
	const num1 = 1
	const num2 = 2
	return []preHandleFaultInfoTestCase{
		{
			name:  "should filter fault list when pod not running on fault server",
			jobId: testJobID1,
			faultInfo: constant.JobFaultInfo{
				FaultList: []constant.FaultRank{
					{PodUid: testPodUid1, RankId: testRankId1},
					{PodUid: testPodUid2, RankId: testRankId2},
					{PodUid: testPodUid3, RankId: testRankId3},
				},
				FaultDevice: []constant.FaultDevice{{ServerName: testServerName1}},
			},
			mockServerMap: map[string]bool{testServerName1: true, testServerName2: true},
			mockPodToServerMap: map[string]string{
				testPodUid1: testServerName1,
				testPodUid2: testServerName2,
				testPodUid3: testServerName1,
			},
			expectedFaultDeviceLen: num1,
			expectedFaultListLen:   num2},
		{name: "should skip pending pods without server mapping",
			jobId: testJobID1,
			faultInfo: constant.JobFaultInfo{
				FaultList: []constant.FaultRank{
					{PodUid: testPodUid1, RankId: testRankId1},
					{PodUid: testPodUid2, RankId: testRankId2},
				},
				FaultDevice: []constant.FaultDevice{{ServerName: testServerName1}},
			},
			mockServerMap:          map[string]bool{testServerName1: true},
			mockPodToServerMap:     map[string]string{testPodUid1: testServerName1},
			expectedFaultDeviceLen: num1,
			expectedFaultListLen:   num1},
	}
}

func TestPreHandleFaultInfo(t *testing.T) {
	var testCases []preHandleFaultInfoTestCase
	testCases = append(testCases, buildPreHandleFaultInfoTestCases1()...)
	testCases = append(testCases, buildPreHandleFaultInfoTestCases2()...)
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			service := &FaultRecoverService{}
			faultInfo := tc.faultInfo
			patch := gomonkey.ApplyFuncReturn(getJobServerInfos, tc.mockServerMap, tc.mockPodToServerMap)
			defer patch.Reset()
			service.preHandleFaultInfo(tc.jobId, &faultInfo)
			convey.So(len(faultInfo.FaultDevice), convey.ShouldEqual, tc.expectedFaultDeviceLen)
			convey.So(len(faultInfo.FaultList), convey.ShouldEqual, tc.expectedFaultListLen)
		})
	}
}
