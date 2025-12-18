/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package grpc is a DT collection for func in grpc
package grpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model"
	"ascend-faultdiag-online/pkg/utils"
	"ascend-faultdiag-online/pkg/utils/grpc/job"
	"ascend-faultdiag-online/pkg/utils/grpc/profiling"
	"ascend-faultdiag-online/pkg/utils/grpc/pubfault"
)

var (
	connectFailed = false
	testJobId     = "testJobId"
	testJobName   = "testJobName"
	testNamespace = "testNamespace"
	failedCount   = 2
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&config, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func TestMain(m *testing.M) {
	// mock grpc.NewClient
	mockNewClient := gomonkey.ApplyFunc(grpc.Dial, func(string, ...grpc.DialOption) (*grpc.ClientConn, error) {
		if connectFailed {
			return nil, fmt.Errorf("failed to connect to grpc server")
		}
		return &grpc.ClientConn{}, nil
	})

	defer mockNewClient.Reset()

	// Run the tests
	code := m.Run()

	// Exit with the test result code
	fmt.Printf("exit_code = %v\n", code)
}

func TestConnect(t *testing.T) {
	var validIp = "127.0.0.1"
	var invalidIp = "0000"
	convey.Convey("test connect", t, func() {
		// c.conn is not nil
		c := &Client{}
		c.conn = &grpc.ClientConn{}
		err := c.connect(validIp)
		convey.So(err, convey.ShouldBeNil)
		c.conn = nil
		// invalid ip
		err = c.connect(invalidIp)
		convey.So(err.Error(), convey.ShouldContainSubstring, "invalid host")
		// dial faild
		connectFailed = true
		err = c.connect(validIp)
		convey.So(err.Error(), convey.ShouldContainSubstring, "failed to connect to grpc server")
		// success
		connectFailed = false
		err = c.connect(validIp)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestClose(t *testing.T) {
	c := &Client{}
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&grpc.ClientConn{}), "Close", func(*grpc.ClientConn) error {
		c.conn = nil
		return nil
	})
	defer patch.Reset()
	convey.Convey("test Close", t, func() {
		convey.So(c.conn, convey.ShouldBeNil)
		// conn is nil
		c.Close()
		convey.So(c.conn, convey.ShouldBeNil)
		c.conn = &grpc.ClientConn{}
		c.Close()
		convey.So(c.conn, convey.ShouldBeNil)
	})
}

func TestGrpc(t *testing.T) {
	// test connect
	connectFailed = true
	_, err := GetClient()
	assert.NotNil(t, err)
	connectFailed = false

	// test multiple connect
	client1, err := GetClient()
	assert.NotNil(t, err)
	client2, err := GetClient()
	assert.NotNil(t, err)
	assert.Equal(t, client1, client2)
}

func TestProfiling(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.GetClusterIp, func() string {
		return "127.0.0.1"
	})
	defer patches.Reset()
	patches.ApplyFunc(GetClient, func() (*Client, error) {
		return &Client{}, nil
	})
	c, err := GetClient()
	assert.Nil(t, err)

	// mock profilingSwitch
	patches.ApplyPrivateMethod(reflect.TypeOf(c), "profilingSwitch",
		func(*Client, *profiling.DataTypeReq) (*profiling.DataTypeRes, error) {
			return &profiling.DataTypeRes{
				Message: "success",
				Code:    0,
			}, nil
		})
	// test start all profiling
	err = c.StartAllProfiling("job1", "ns1")
	assert.Nil(t, err)

	// test start heavy profiling
	err = c.StartHeavyProfiling("job1", "ns1")
	assert.Nil(t, err)

	// test stop heavy profiling
	err = c.StopHeavyProfiling("job1", "ns1")
	assert.Nil(t, err)
}

func TestProfilingSwitch(t *testing.T) {
	var c = &Client{conn: &grpc.ClientConn{}}
	c.tc = profiling.NewTrainingDataTraceClient(c.conn)
	patch := gomonkey.ApplyMethodFunc(
		reflect.TypeOf(c.tc),
		"ModifyTrainingDataTraceSwitch",
		func(context.Context, *profiling.DataTypeReq, ...grpc.CallOption) (*profiling.DataTypeRes, error) {
			return &profiling.DataTypeRes{}, nil
		},
	)
	defer patch.Reset()
	convey.Convey("test profilingSwitch", t, func() {
		_, err := c.profilingSwitch(&profiling.DataTypeReq{})
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestReportFault(t *testing.T) {
	var c = &Client{conn: &grpc.ClientConn{}}
	c.pf = pubfault.NewPubFaultClient(c.conn)
	patch := gomonkey.ApplyMethodFunc(
		reflect.TypeOf(c.pf),
		"SendPublicFault",
		func(context.Context, *pubfault.PublicFaultRequest, ...grpc.CallOption) (*pubfault.RespStatus, error) {
			return &pubfault.RespStatus{}, nil
		},
	)
	defer patch.Reset()
	convey.Convey("test ReportFault", t, func() {
		err := c.ReportFault([]*pubfault.Fault{})
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestRegisterJobSummary(t *testing.T) {
	c := &Client{
		conn: &grpc.ClientConn{},
		jobSummaryWatcher: jobSummaryWatcher{
			callbacks: utils.NewStorage[callback](),
		},
	}
	c.jc = job.NewJobClient(c.conn)
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	var callCount = 0
	patches.ApplyMethodFunc(
		reflect.TypeOf(c.jc),
		"Register",
		func(ctx context.Context, info *job.ClientInfo, opts ...grpc.CallOption) (
			*job.Status, error) {
			callCount++
			return &job.Status{ClientId: "test-client-id"}, nil
		},
	)
	patches.ApplyMethodFunc(
		reflect.TypeOf(c.jc),
		"SubscribeJobSummarySignal",
		func(ctx context.Context, info *job.ClientInfo, opts ...grpc.CallOption) (
			job.Job_SubscribeJobSummarySignalClient, error) {
			return nil, nil
		},
	)
	patches.ApplyPrivateMethod(reflect.TypeOf(c), "processJobSummary", func(
		*Client, job.Job_SubscribeJobSummarySignalClient) {
	})
	patches.ApplyPrivateMethod(reflect.TypeOf(c), "supervisor", func(*Client) {})
	convey.Convey("test registerJobSummary", t, func() {
		// register the first time
		convey.So(c.isRegisterd, convey.ShouldBeFalse)
		c.registerJobSummary()
		convey.So(c.isRegisterd, convey.ShouldBeTrue)
		convey.So(callCount, convey.ShouldEqual, 1)

		// parallel call register
		c.reset()
		callCount = 0
		var count = 100
		for i := 0; i < count; i++ {
			c.registerJobSummary()
		}
		convey.So(callCount, convey.ShouldEqual, 1)
		// wait until all goroutines finished
		time.Sleep(time.Millisecond)
	})
}

type mockJobSummaryStream struct {
	job.Job_SubscribeJobSummarySignalClient
	closeSendCalled int
	receiveCount    int
}

func (m *mockJobSummaryStream) CloseSend() error {
	m.closeSendCalled++
	if m.closeSendCalled < failedCount {
		return errors.New("mock close send failed")
	}
	return nil
}

func (m *mockJobSummaryStream) Recv() (*job.JobSummarySignal, error) {
	if m.receiveCount == 1 {
		return nil, fmt.Errorf("mock recv error")
	}
	m.receiveCount++
	return &job.JobSummarySignal{
		JobId:     testJobId,
		JobName:   testJobName,
		Namespace: testNamespace,
		HcclJson:  "{}",
		JobStatus: enum.IsRunning,
	}, nil
}

func TestProcessJobSummary(t *testing.T) {
	c := &Client{
		jobSummaryWatcher: jobSummaryWatcher{
			callbacks: utils.NewStorage[callback](),
		},
	}
	stream := &mockJobSummaryStream{}
	convey.Convey("test processJobSummary", t, func() {
		c.closeSignal = make(chan struct{})
		// callbacks is empty, end loop
		var received bool
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			<-c.closeSignal
			received = true
			defer wg.Done()
		}()
		c.processJobSummary(stream)
		wg.Wait()
		convey.So(received, convey.ShouldBeTrue)
		convey.So(stream.closeSendCalled, convey.ShouldEqual, failedCount)

		var f = func(job *model.JobSummary) {}
		c.callbacks.Store("testRegisterId", callback{
			registerId: "testRegisterId",
			jobName:    testJobName,
			namespace:  testNamespace,
			f:          f,
		})
		// mock recived data and then failed
		received = false
		c.disconnectedSignal = make(chan struct{})
		wg.Add(1)
		go func() {
			<-c.disconnectedSignal
			received = true
			defer wg.Done()
		}()
		c.processJobSummary(stream)
		wg.Wait()
		convey.So(received, convey.ShouldBeTrue)
		time.Sleep(time.Millisecond)
		// got value
		job, ok := storage.Load(fmt.Sprintf("%s/%s", testNamespace, testJobName))
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(job.JobId, convey.ShouldEqual, testJobId)
		convey.So(job.JobName, convey.ShouldEqual, testJobName)
		convey.So(job.Namespace, convey.ShouldEqual, testNamespace)
	})
}

func TestSupervisor(t *testing.T) {
	c := &Client{}
	patch := gomonkey.ApplyPrivateMethod(
		c,
		"registerJobSummary",
		func(*Client) error {
			go c.supervisor()
			c.isRegisterd = true
			return nil
		},
	)
	defer patch.Reset()
	convey.Convey("test supervisor", t, func() {
		c.disconnectedSignal = make(chan struct{})
		c.closeSignal = make(chan struct{})
		go c.supervisor()
		// mock disconnect, client whill start supervisor again
		c.disconnectedSignal <- struct{}{}
		time.Sleep(time.Millisecond)
		convey.So(c.isRegisterd, convey.ShouldBeTrue)
		// mock close
		c.closeSignal <- struct{}{}
		time.Sleep(time.Millisecond)
		convey.So(c.isRegisterd, convey.ShouldBeFalse)
	})
}

func TestSubAndUnSubJobSummary(t *testing.T) {
	var c = &Client{
		jobSummaryWatcher: jobSummaryWatcher{
			callbacks: utils.NewStorage[callback](),
		},
	}
	convey.Convey("test UnsubscribeJobSummary", t, func() {
		testSubNoPanic(c)
		testSubAndUnsub(c)
		testSubFailed(c)
	})
}

func testSubNoPanic(c *Client) {
	// callbacks is empty, unsub no panic
	c.UnsubscribeJobSummary("ramdonId")
	convey.So(func() {
		c.UnsubscribeJobSummary("ramdonId")
	}, convey.ShouldNotPanic)
}

func testSubAndUnsub(c *Client) {
	// sub then unsub
	patch := gomonkey.ApplyPrivateMethod(
		c,
		"registerJobSummary",
		func(*Client) error {
			return nil
		},
	)
	defer patch.Reset()
	// register 10 func
	var funcCount = 10
	var ids = []string{}
	for i := 0; i < funcCount; i++ {
		registerId, err := c.SubscribeJobSummary("", "", func(job *model.JobSummary) {})
		convey.So(err, convey.ShouldBeNil)
		ids = append(ids, registerId)
	}
	convey.So(len(ids), convey.ShouldEqual, funcCount)
	convey.So(c.callbacks.Len(), convey.ShouldEqual, funcCount)
	// unregister all
	for _, id := range ids {
		c.UnsubscribeJobSummary(id)
	}
	convey.So(c.callbacks.Len(), convey.ShouldEqual, 0)

	// test data in storage
	storage.Store("", &model.JobSummary{JobId: testJobId})
	var f = func(job *model.JobSummary) {
		convey.So(job.JobId, convey.ShouldEqual, testJobId)
	}
	registerId, err := c.SubscribeJobSummary(testJobName, testNamespace, f)
	convey.So(err, convey.ShouldBeNil)
	convey.So(registerId, convey.ShouldNotBeEmpty)
	c.UnsubscribeJobSummary(registerId)
}

func testSubFailed(c *Client) {
	// register failed
	patch := gomonkey.ApplyPrivateMethod(
		c,
		"registerJobSummary",
		func(*Client) error {
			return errors.New("mock registerJobSummary failed")
		},
	)
	defer patch.Reset()
	var funcCount = 10
	for i := 0; i < funcCount; i++ {
		registerId, err := c.SubscribeJobSummary(testJobName, testNamespace, func(job *model.JobSummary) {})
		convey.So(err.Error(), convey.ShouldEqual, "mock registerJobSummary failed")
		convey.So(registerId, convey.ShouldBeEmpty)
	}
	convey.So(c.callbacks.Len(), convey.ShouldEqual, 0)

}
