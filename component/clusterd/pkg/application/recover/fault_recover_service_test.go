// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package recover a series of fault recover service test function
package recover

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/podgroup"
	"clusterd/pkg/interface/grpc/recover"
)

const (
	keepAliveSecond = 10
	length2         = 2
	fakeJobID1      = "fakeJobID1"
	fakeJobID2      = "fakeJobID2"
	fakeJobID       = "fakeJobID"
)

func TestNotifyFaultInfoForJob(t *testing.T) {
	convey.Convey("Test notifyFaultInfoForJob", t, func() {
		svr := &FaultRecoverService{
			eventCtl: map[string]*EventController{
				fakeJobID1: {jobInfo: common.JobBaseInfo{
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
					{RankId: "1", DoStepRetry: false},
					{RankId: "2", DoStepRetry: true},
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
			convey.So(len(ctl.cacheUceFault) == 1, convey.ShouldBeFalse)
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
		patches := gomonkey.ApplyFunc(ctl.reset, func() {
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

		patches := gomonkey.ApplyFunc(ctl.reset, func() {
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

		patches := gomonkey.ApplyFunc(ctl.reset, func() {
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

		patches := gomonkey.ApplyFunc(faultmanager.CallbackForReportUceInfo,
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

func TestGetJobBaseInfo_Normal(t *testing.T) {
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

func TestGetJobBaseInfo_GetPGFromCacheError(t *testing.T) {
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

func TestGetJobBaseInfo_GetRecoverBaseInfoError(t *testing.T) {
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

func TestGetJobBaseInfo_ProcessRecoverEnableOff(t *testing.T) {
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

func Test_GetFaultReason(t *testing.T) {
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
		convey.So(getFaultReason(faults3), convey.ShouldEqual, uceFaultValue)
	})
}
