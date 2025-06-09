package jobinfo

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/domain/common"
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

// TestJobServer_Register tests client registration
func TestJobServer_Register(t *testing.T) {
	convey.Convey("Given a JobServer", t, func() {
		ctx := context.Background()
		server := NewJobServer(ctx)

		convey.Convey("When registering with valid role", func() {
			req := &job.ClientInfo{Role: CCAgentClientName}
			resp, err := server.Register(ctx, req)

			convey.Convey("It should succeed", func() {
				convey.So(err, convey.ShouldBeNil)
				convey.So(resp.Code, convey.ShouldEqual, int32(common.SuccessCode))
				convey.So(resp.ClientId, convey.ShouldNotBeEmpty)
				convey.So(len(server.clients), convey.ShouldEqual, 1)
			})
		})

		convey.Convey("When registering with invalid role", func() {
			req := &job.ClientInfo{Role: "InvalidRole"}
			resp, err := server.Register(ctx, req)

			convey.Convey("It should fail", func() {
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(resp.Code, convey.ShouldEqual, int32(common.UnRegistry))
				convey.So(len(server.clients), convey.ShouldEqual, 0)
			})
		})
	})
}

// TestJobServer_Subscribe tests job summary subscription
func TestJobServer_Subscribe(t *testing.T) {
	convey.Convey("Given a registered client", t, func() {
		ctx := context.Background()
		server := NewJobServer(ctx)
		clientID := registerTestClient(server, ctx, CCAgentClientName)

		convey.Convey("When subscribing", func() {
			stream := &mockStream{ctx: ctx}
			req := &job.ClientInfo{ClientId: clientID}
			go func() {
				err := server.SubscribeJobSummarySignal(req, stream)
				convey.ShouldBeNil(err)
			}()

			convey.Convey("It should receive broadcast messages", func() {
				signal := job.JobSummarySignal{JobId: "test-job"}
				server.broadcastJobUpdate(signal)
				time.Sleep(time.Second)
				convey.So(len(stream.msgs), convey.ShouldEqual, 1)
				convey.So(stream.msgs[0].JobId, convey.ShouldEqual, "test-job")
			})
		})
	})
}

// TestJobServer_SubscribeBreakStream tests job summary subscription break stream
func TestJobServer_SubscribeBreakStream(t *testing.T) {
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

// TestJobServer_SubscribeFakeClient tests job summary subscription with fake client
func TestJobServer_SubscribeFakeClient(t *testing.T) {
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

// TestJobServer_Broadcast tests message broadcasting
func TestJobServer_Broadcast(t *testing.T) {
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

// TestClientState_SafeClose tests safe channel closing
func TestClientState_SafeClose(t *testing.T) {
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
