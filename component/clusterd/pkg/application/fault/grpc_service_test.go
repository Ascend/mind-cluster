// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package fault a series of controller test function
package fault

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/config"
	"clusterd/pkg/application/faultmanager/jobprocess/faultrank"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/logs"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/fault"
)

func init() {
	ctx := context.Background()
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, ctx)
	logs.InitGrpcEventLogger(ctx)
}

const (
	sleepTime            = 100 * time.Millisecond
	mockJobIdWithChinese = "job123你好"
	mockValidJobId       = "abc12345"
)

func fakeFaultService() *FaultServer {
	return &FaultServer{
		serviceCtx:     context.Background(),
		faultPublisher: make(map[string]map[string]*config.ConfigPublisher[*fault.FaultMsgSignal]),
		lock:           sync.RWMutex{},
		limiter:        util.NewAdvancedRateLimiter(defaultTokenRate, defaultBurst, defaultMaxQueueLen),
	}
}

func TestRegister(t *testing.T) {
	jobInfo := constant.JobInfo{MultiInstanceJobId: "testFaultJobId", AppType: "app", NameSpace: "ns"}
	job.SaveJobCache("job1", jobInfo)
	convey.Convey("Testing Register", t, func() {
		convey.Convey("register ordinary job success", func() {
			service := fakeFaultService()
			req := &fault.ClientInfo{JobId: "job1", Role: "app"}
			status, err := service.Register(nil, req)
			convey.So(status, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("register multi-instance failed, should return 402 code and error", func() {
			service := fakeFaultService()
			req := &fault.ClientInfo{JobId: "testFaultJobId", Role: "role"}
			status, err := service.Register(nil, req)
			convey.So(status.Code, convey.ShouldEqual, common.JobNotExist)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("register multi-instance job success", func() {
			service := fakeFaultService()
			req := &fault.ClientInfo{JobId: "testFaultJobId", Role: "app"}
			status, err := service.Register(nil, req)
			convey.So(status, convey.ShouldNotBeNil)
			convey.So(status.Info, convey.ShouldEqual, "register success")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSubscribeFaultMsgSignal(t *testing.T) {
	req := &fault.ClientInfo{
		JobId: fakeJobID1,
		Role:  fakeRole1,
	}
	convey.Convey("controller not exist, should return error", t, func() {
		service := fakeFaultService()
		err := service.SubscribeFaultMsgSignal(req, nil)
		convey.So(err, convey.ShouldResemble, errors.New("jobId=fakeJobId1 not registered, role=fakeRole1"))
	})
	convey.Convey("subscribe rank table service success, should return nil", t, func() {
		service := fakeFaultService()
		service.addPublisher(fakeJobID1, fakeRole1)
		go func() {
			for {
				publisher, ok := service.getPublisher(fakeJobID1, fakeRole1)
				if ok && publisher.IsSubscribed() {
					publisher.Stop()
					break
				}
				time.Sleep(sleepTime)
			}
		}()
		err := service.SubscribeFaultMsgSignal(req, &mockConfigSubscribeFaultMsgServer{})
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestIsValidJobId test isValidJobId
func TestIsValidJobId(t *testing.T) {
	tests := []struct {
		name     string
		jobId    string
		expected bool
	}{
		{
			name:     "TestIsValidJobId with valid jobId",
			jobId:    mockValidJobId,
			expected: true,
		},
		{
			name:     "TestIsValidJobId with too short jobId",
			jobId:    string(bytes.Repeat([]byte{'a'}, minJobIdLen-1)),
			expected: false,
		},
		{
			name:     "TestIsValidJobId with Too long jobId",
			jobId:    string(bytes.Repeat([]byte{'a'}, maxJobIdLen+1)),
			expected: false,
		},
		{
			name:     "TestIsValidJobId with jobId contains Chinese",
			jobId:    mockJobIdWithChinese,
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidJobId(tt.jobId)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetFaultMsgSignal test getFaultMsgSignal
func TestGetFaultMsgSignal(t *testing.T) {
	service := fakeFaultService()
	convey.Convey("TestGetFaultMsgSignal with invalid jobId", t, func() {
		resp, err := service.GetFaultMsgSignal(context.TODO(), &fault.ClientInfo{
			JobId: mockJobIdWithChinese,
			Role:  "",
		})
		convey.ShouldNotBeNil(err)
		convey.So(resp.Code, convey.ShouldEqual, int32(common.InvalidReqParam))
	})
	convey.Convey("TestGetFaultMsgSignal with valid jobId but jobId not exist", t, func() {
		resp, err := service.GetFaultMsgSignal(context.TODO(), &fault.ClientInfo{
			JobId: mockValidJobId,
			Role:  "",
		})
		convey.ShouldBeNil(err)
		convey.So(resp.Code, convey.ShouldEqual, int32(common.SuccessCode))
		convey.So(resp.FaultSignal, convey.ShouldBeNil)
	})
	convey.Convey("TestGetFaultMsgSignal with valid jobId but jobId exist", t, func() {
		patches := gomonkey.ApplyMethodReturn(faultrank.JobFaultRankProcessor, "GetJobFaultRankInfos",
			map[string]constant.JobFaultInfo{mockValidJobId: {}})
		defer patches.Reset()
		resp, err := service.GetFaultMsgSignal(context.TODO(), &fault.ClientInfo{
			JobId: mockValidJobId,
			Role:  "",
		})
		convey.ShouldBeNil(err)
		convey.So(resp.Code, convey.ShouldEqual, int32(common.SuccessCode))
		convey.So(resp.FaultSignal, convey.ShouldNotBeNil)
	})
	convey.Convey("TestGetFaultMsgSignal without jobId", t, func() {
		resp, err := service.GetFaultMsgSignal(context.TODO(), &fault.ClientInfo{
			JobId: "",
			Role:  "",
		})
		convey.ShouldBeNil(err)
		convey.So(resp.Code, convey.ShouldEqual, int32(common.SuccessCode))
	})
	convey.Convey("TestGetFaultMsgSignal with too many requests", t, func() {
		patches := gomonkey.ApplyMethodReturn(&util.AdvancedRateLimiter{}, "Allow", false)
		defer patches.Reset()
		resp, err := service.GetFaultMsgSignal(context.TODO(), &fault.ClientInfo{
			JobId: "",
			Role:  "",
		})
		convey.ShouldNotBeNil(err)
		convey.So(resp.Code, convey.ShouldEqual, int32(common.RateLimitedCode))
	})
}

func TestAddAndGetFaultPublisher(t *testing.T) {
	convey.Convey("Test addFaultPublisher and GetFaultPublisher success", t, func() {
		service := fakeFaultService()
		service.addPublisher(fakeJobID1, fakeRole1)

		publisher, ok := service.getPublisher(fakeJobID1, fakeRole1)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(publisher, convey.ShouldNotBeNil)
	})
}

// TestPreemptPublisher for test preemptPublisher
func TestPreemptPublisher(t *testing.T) {
	convey.Convey("test preemptPublisher", t, func() {
		service := fakeFaultService()
		service.addPublisher(fakeJobID1, fakeRole1)
		publisher, _ := service.getPublisher(fakeJobID1, fakeRole1)
		convey.Convey("01-publisher already exist, should preempt old publisher", func() {
			newPublisher := service.preemptPublisher(fakeJobID1, fakeRole1)
			convey.So(newPublisher, convey.ShouldNotBeNil)
			convey.So(newPublisher.GetCreateTime().After(publisher.GetCreateTime()), convey.ShouldBeTrue)
		})
	})
}

func TestGetPublisherList(t *testing.T) {
	convey.Convey("test getPublisherListByJobId and getAllPublisherList", t, func() {
		service := fakeFaultService()
		service.addPublisher(fakeJobID1, fakeRole1)
		publisherList := service.getPublisherListByJobId(fakeJobID1)
		allPublisherList := service.getAllPublisherList()
		convey.ShouldResemble(publisherList, allPublisherList)
		publisherList = service.getPublisherListByJobId(fakeJobID2)
		convey.ShouldBeTrue(len(publisherList) == 0)
	})
}

func TestDeletePublisher(t *testing.T) {
	convey.Convey("test deletePublisher", t, func() {
		service := fakeFaultService()
		service.addPublisher(fakeJobID1, fakeRole1)
		publisher, _ := service.getPublisher(fakeJobID1, fakeRole1)
		service.deletePublisher(fakeJobID2, fakeRole1, time.Now())
		convey.ShouldBeFalse(len(service.faultPublisher) == 0)
		service.deletePublisher(fakeJobID1, "role", time.Now())
		convey.ShouldBeFalse(len(service.faultPublisher) == 0)
		service.deletePublisher(fakeJobID1, fakeRole1, publisher.GetCreateTime())
		convey.ShouldBeTrue(len(service.faultPublisher) == 0)
	})
}

type mockConfigSubscribeFaultMsgServer struct {
	grpc.ServerStream
}

func (x *mockConfigSubscribeFaultMsgServer) Send(m *fault.FaultMsgSignal) error {
	return errors.New("send failed")
}

func (x *mockConfigSubscribeFaultMsgServer) Context() context.Context {
	return context.Background()
}
