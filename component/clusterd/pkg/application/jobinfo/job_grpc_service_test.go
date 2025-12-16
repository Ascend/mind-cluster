/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package jobinfo is used to return job info by subscribe

package jobinfo

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/uuid"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	jobstorage "clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/job"
)

const (
	stramCount        = 2
	CCAgentClientName = "CCAgent"
	jobSignalChanLen  = 5
)

func init() {
	err := hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
	convey.ShouldBeNil(err)
}

// TestNewJobServer tests JobServer initialization
func TestNewJobServer(t *testing.T) {
	convey.Convey("Given a new JobServer", t, func() {
		ctx := context.Background()
		server := NewJobServer(ctx)

		convey.Convey("It should not be nil", func() {
			convey.So(server, convey.ShouldNotBeNil)
		})

		convey.Convey("It should have an empty clients map", func() {
			convey.So(server.clients, convey.ShouldNotBeNil)
			convey.So(len(server.clients), convey.ShouldEqual, 0)
		})
	})
}

// TestJobServerRegister tests client registration
func TestJobServerRegister(t *testing.T) {
	convey.Convey("Given a JobServer", t, func() {
		ctx := context.Background()
		server := NewJobServer(ctx)

		convey.Convey("When registering with valid role, it should succeed", func() {
			req := &job.ClientInfo{Role: CCAgentClientName}
			resp, err := server.Register(ctx, req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp.Code, convey.ShouldEqual, int32(common.SuccessCode))
			convey.So(resp.ClientId, convey.ShouldNotBeEmpty)
			convey.So(len(server.clients), convey.ShouldEqual, 1)
		})

		convey.Convey("When registering with invalid role, it should fail", func() {
			req := &job.ClientInfo{Role: "InvalidRole"}
			resp, err := server.Register(ctx, req)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(resp.Code, convey.ShouldEqual, int32(common.UnRegistry))
			convey.So(len(server.clients), convey.ShouldEqual, 0)
		})

		convey.Convey("When registering clients more than maxClientNum, it should fail", func() {
			req := &job.ClientInfo{Role: CCAgentClientName}
			mockClientMap := make(map[string]*clientState, maxClientNum)
			for i := 0; i < maxClientNum; i++ {
				mockClientMap[string(uuid.NewUUID())] = &clientState{}
			}
			server.clients = mockClientMap
			resp, err := server.Register(ctx, req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp.Code, convey.ShouldEqual, int32(common.RateLimitedCode))
			convey.So(resp.ClientId, convey.ShouldBeEmpty)
			convey.So(len(server.clients), convey.ShouldEqual, maxClientNum)
		})
	})
}

// TestJobServerSubscribe tests job summary subscription
func TestJobServerSubscribe(t *testing.T) {
	convey.Convey("Given a registered client", t, func() {
		const (
			testJob1 = "test-job1"
			testJob2 = "test-job2"
			jobNum   = 2
		)
		patch := gomonkey.ApplyFunc(jobstorage.GetAllJobCache, func() map[string]constant.JobInfo {
			return map[string]constant.JobInfo{testJob1: {Key: testJob1}}
		})
		defer patch.Reset()
		ctx := context.Background()
		server := NewJobServer(ctx)
		clientID := registerTestClient(server, ctx, constant.RoleFdAgent)

		convey.Convey("When subscribing", func() {
			stream := &mockStream{ctx: ctx}
			req := &job.ClientInfo{ClientId: clientID}
			go func() {
				err := server.SubscribeJobSummarySignal(req, stream)
				convey.ShouldBeNil(err)
			}()

			convey.Convey("It should receive broadcast messages", func() {
				time.Sleep(time.Second)
				signal := job.JobSummarySignal{JobId: testJob2}
				server.broadcastJobUpdate(signal)
				time.Sleep(time.Second)
				convey.So(len(stream.msgs), convey.ShouldEqual, jobNum)
				convey.So(stream.msgs[0].JobId, convey.ShouldEqual, testJob1)
				convey.So(stream.msgs[1].JobId, convey.ShouldEqual, testJob2)
			})
		})
	})
}

// TestJobServerSubscribeBreakStream tests job summary subscription break stream
func TestJobServerSubscribeBreakStream(t *testing.T) {
	convey.Convey("Given a registered client", t, func() {
		ctx := context.Background()
		server := NewJobServer(ctx)
		clientID := registerTestClient(server, ctx, CCAgentClientName)

		convey.Convey("When subscribing", func() {
			streamCtx, cancel := context.WithCancel(ctx)
			stream := &mockStream{ctx: streamCtx}
			req := &job.ClientInfo{ClientId: clientID}
			go func() {
				err := server.SubscribeJobSummarySignal(req, stream)
				convey.ShouldBeNil(err)
			}()

			convey.Convey("It should be closed", func() {
				cancel()
				time.Sleep(time.Second)
				convey.ShouldBeNil(server.clients[clientID])
			})
		})
	})
}

// TestJobServerSubscribeFakeClient tests job summary subscription with fake client
func TestJobServerSubscribeFakeClient(t *testing.T) {
	convey.Convey("Given a registered client", t, func() {
		ctx := context.Background()
		server := NewJobServer(ctx)

		convey.Convey("When subscribing", func() {
			stream := &mockStream{ctx: ctx}
			req := &job.ClientInfo{ClientId: "fakeClient"}
			go func() {
				err := server.SubscribeJobSummarySignal(req, stream)
				convey.ShouldNotBeNil(err)
			}()
			time.Sleep(time.Second)
		})
	})
}

// TestJobServerBroadcast tests message broadcasting
func TestJobServerBroadcast(t *testing.T) {
	convey.Convey("Given multiple clients", t, func() {
		ctx := context.Background()
		server := NewJobServer(ctx)
		// Register 2 clients
		client1 := registerTestClient(server, ctx, CCAgentClientName)
		client2 := registerTestClient(server, ctx, "DefaultUser1")
		// Create mock streams
		stream1 := &mockStream{ctx: ctx}
		stream2 := &mockStream{ctx: ctx}
		// Start subscriptions in goroutines
		var wg sync.WaitGroup
		wg.Add(stramCount)
		go func() {
			defer wg.Done()
			err := server.SubscribeJobSummarySignal(&job.ClientInfo{ClientId: client1}, stream1)
			convey.ShouldBeNil(err)
		}()
		go func() {
			defer wg.Done()
			err := server.SubscribeJobSummarySignal(&job.ClientInfo{ClientId: client2}, stream2)
			convey.ShouldBeNil(err)
		}()
		time.Sleep(time.Second)
		convey.Convey("When broadcasting a message", func() {
			signal := job.JobSummarySignal{JobId: "shared-job"}
			server.broadcastJobUpdate(signal)
			time.Sleep(time.Second)
			convey.Convey("All clients should receive it", func() {
				convey.So(len(stream1.msgs), convey.ShouldEqual, 1)
				convey.So(len(stream2.msgs), convey.ShouldEqual, 1)
				convey.So(stream1.msgs[0].JobId, convey.ShouldEqual, "shared-job")
				convey.So(stream2.msgs[0].JobId, convey.ShouldEqual, "shared-job")
			})
		})
	})
}

// TestClientStateSafeClose tests safe channel closing
func TestClientStateSafeClose(t *testing.T) {
	convey.Convey("Given a client state", t, func() {
		state := &clientState{
			clientChan: make(chan job.JobSummarySignal, jobSignalChanLen),
			closed:     false,
		}

		convey.Convey("When closing the channel", func() {
			state.safeCloseChannel()

			convey.Convey("It should be marked as closed", func() {
				convey.So(state.closed, convey.ShouldBeTrue)
			})

			convey.Convey("Reclosing should not panic", func() {
				convey.So(func() { state.safeCloseChannel() }, convey.ShouldNotPanic)
			})
		})
	})
}

// Helper function to register a test client
func registerTestClient(server *JobServer, ctx context.Context, role string) string {
	req := &job.ClientInfo{Role: role}
	resp, _ := server.Register(ctx, req)
	return resp.ClientId
}

// Mock implementation of Job_SubscribeJobSummarySignalServer
type mockStream struct {
	job.Job_SubscribeJobSummarySignalServer
	ctx       context.Context
	msgs      []job.JobSummarySignal
	sendError error
}

func (m *mockStream) Context() context.Context { return m.ctx }
func (m *mockStream) Send(msg *job.JobSummarySignal) error {
	if m.sendError != nil {
		return m.sendError
	}
	m.msgs = append(m.msgs, *msg)
	return nil
}
