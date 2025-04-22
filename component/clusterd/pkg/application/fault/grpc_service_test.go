package fault

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/config"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/fault"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

const sleepTime = 100 * time.Millisecond

func fakeFaultService() *FaultServer {
	return &FaultServer{
		serviceCtx:     context.Background(),
		faultPublisher: make(map[string]*config.ConfigPublisher[*fault.FaultMsgSignal]),
		lock:           sync.RWMutex{},
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
			req := &fault.ClientInfo{JobId: "testFaultJobId", Role: ""}
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
		Role:  "",
	}
	convey.Convey("controller not exist, should return error", t, func() {
		service := fakeFaultService()
		err := service.SubscribeFaultMsgSignal(req, nil)
		convey.So(err, convey.ShouldResemble, errors.New("jobId=fakeJobId1 not registered, role="))
	})
	convey.Convey("subscribe rank table service success, should return nil", t, func() {
		service := fakeFaultService()
		service.addPublisher(fakeJobID1)
		go func() {
			for {
				publisher, ok := service.getPublisher(fakeJobID1)
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

func TestAddAndGetFaultPublisher(t *testing.T) {
	convey.Convey("Test addFaultPublisher and GetFaultPublisher success", t, func() {
		service := fakeFaultService()
		service.addPublisher(fakeJobID1)

		publisher, ok := service.getPublisher(fakeJobID1)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(publisher, convey.ShouldNotBeNil)
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
