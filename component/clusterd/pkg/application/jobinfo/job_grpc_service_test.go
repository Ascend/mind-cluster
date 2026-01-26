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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/uuid"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/logs"
	"clusterd/pkg/domain/common"
	jobstorage "clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/job"
)

const (
	CCAgentClientName  = "CCAgent"
	jobSignalChanLen   = 5
	testJob1           = "test-job1"
	testJob2           = "test-job2"
	two                = 2
	oneHundred         = 100
	fifteenThousand    = 15000
	thirtyFiveThousand = 35000
	fiftyThousand      = 50000
)

func init() {
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	err := hwlog.InitRunLogger(logConfig, context.Background())
	convey.ShouldBeNil(err)
	logger, err := hwlog.NewCustomLogger(logConfig, context.Background())
	convey.ShouldBeNil(err)
	logs.JobEventLog = logger
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

		convey.Convey("It should init limiter correctly", func() {
			convey.So(server.limiter, convey.ShouldNotBeNil)
			convey.So(server.limiter.Limit(), convey.ShouldEqual, rate.Limit(constant.RequestNumPerSecondLimit))
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

		convey.Convey("When registering clients more than maxClientPerRole, it should fail", func() {
			req := &job.ClientInfo{Role: CCAgentClientName}
			mockClientMap := make(map[string]*clientState)
			for i := 0; i < constant.MaxClientPerRole; i++ {
				clientID := string(uuid.NewUUID())
				mockClientMap[clientID] = &clientState{role: CCAgentClientName}
			}
			server.clients = mockClientMap
			resp, err := server.Register(ctx, req)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(resp.Code, convey.ShouldEqual, int32(common.RateLimitedCode))
			convey.So(resp.ClientId, convey.ShouldBeEmpty)
			convey.So(len(server.clients), convey.ShouldEqual, constant.MaxClientPerRole)
		})
	})
}

// TestJobServerSubscribe tests job summary subscription
func TestJobServerSubscribe(t *testing.T) {
	convey.Convey("Given a registered client", t, func() {
		ctx := context.Background()
		server := NewJobServer(ctx)
		clientID := registerTestClient(server, ctx, constant.RoleFdAgent)
		convey.Convey("When subscribing with normal stream", func() {
			stream := &mockStream{ctx: ctx}
			req := &job.ClientInfo{ClientId: clientID}
			go func() {
				err := server.SubscribeJobSummarySignal(req, stream)
				convey.ShouldBeNil(err)
			}()
			convey.Convey("It should receive broadcast messages", func() {
				time.Sleep(time.Millisecond * oneHundred)
				signal := job.JobSummarySignal{JobId: testJob1}
				server.broadcastJobUpdate(signal)
				time.Sleep(time.Millisecond * oneHundred)
				convey.So(len(stream.msgs), convey.ShouldEqual, 1)
				convey.So(stream.msgs[0].JobId, convey.ShouldEqual, testJob1)
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
		wg.Add(two)
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
			state.safeCloseClientResources()
			convey.So(state.closed, convey.ShouldBeTrue)
		})
		convey.Convey("Reclosing should not panic", func() {
			convey.So(func() { state.safeCloseClientResources() }, convey.ShouldNotPanic)
		})
		convey.Convey("Concurrent close should not panic", func() {
			var wg sync.WaitGroup
			for i := 0; i < ten; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					state.safeCloseClientResources()
				}()
			}
			wg.Wait()
			convey.So(state.closed, convey.ShouldBeTrue)
		})
	})
}

func TestSubscribeJobSummarySignalList(t *testing.T) {
	convey.Convey("Given a JobServer and registered client", t, func() {
		ctx := context.Background()
		server := NewJobServer(ctx)
		clientID := registerTestClient(server, ctx, CCAgentClientName)
		req := &job.ClientInfo{ClientId: clientID}
		JobRankTable := constant.RankTable{
			ServerList: []constant.ServerHccl{
				{DeviceList: []constant.Device{{DeviceID: "0", RankID: "0"}, {DeviceID: "1", RankID: "1"}}},
			}}
		mockJobMap := map[string]constant.JobInfo{testJob1: {Key: testJob1, JobRankTable: JobRankTable}}
		patch := gomonkey.ApplyFunc(jobstorage.GetAllJobCache, func() map[string]constant.JobInfo {
			return mockJobMap
		})
		defer patch.Reset()
		testSubscribeJobSummarySignalListWithValidClient(ctx, server, req)
		convey.Convey("When subscribe with invalid clientId", func() {
			invalidReq := &job.ClientInfo{ClientId: "fake-client"}
			mockSignalsStream := NewMockJobSummarySignalListServer(ctx, nil)
			err := server.SubscribeJobSummarySignalList(invalidReq, mockSignalsStream)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid clientId")
		})
		convey.Convey("When rate limited", func() {
			for i := 0; i < constant.RequestNumPerSecondLimit; i++ {
				server.limiter.Allow()
			}
			mockSignalsStream := NewMockJobSummarySignalListServer(ctx, nil)
			err := server.SubscribeJobSummarySignalList(req, mockSignalsStream)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "rate limited")
		})
		convey.Convey("When stream.Send returns error", func() {
			mockSignalsStream := NewMockJobSummarySignalListServer(ctx, errors.New("send error"))
			err := server.SubscribeJobSummarySignalList(req, mockSignalsStream)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "send error")
		})
	})
}

func testSubscribeJobSummarySignalListWithValidClient(ctx context.Context, server *JobServer, req *job.ClientInfo) {
	convey.Convey("When subscribe with valid client", func() {
		mockSignalsStream := NewMockJobSummarySignalListServer(ctx, nil)
		go func() {
			server.SubscribeJobSummarySignalList(req, mockSignalsStream)
		}()
		time.Sleep(time.Millisecond * oneHundred)
		signals := mockSignalsStream.GetSentSignals()
		convey.So(len(signals), convey.ShouldBeGreaterThan, 0)
		convey.So(signals[0].JobSummarySignals[0].JobId,
			convey.ShouldEqual, testJob1)
		broadcastSignal := job.JobSummarySignal{
			JobId:    testJob2,
			HcclJson: `{"rank_id":0}` + strings.Repeat(`{"rank_id":1},`, constant.MaxNPUsPerBatch+1),
		}
		server.broadcastJobUpdate(broadcastSignal)
		time.Sleep(time.Millisecond * oneHundred)
		latestSignals := mockSignalsStream.GetSentSignals()[len(mockSignalsStream.GetSentSignals())-1]
		convey.So(latestSignals.JobSummarySignals[0].JobId, convey.ShouldEqual, testJob2)
		convey.So(latestSignals.JobSummarySignals[0].HcclJson, convey.ShouldBeEmpty)
	})
}

func TestHandleLargeNPUJob(t *testing.T) {
	convey.Convey("Given JobSummarySignal with different NPU num", t, func() {
		server := NewJobServer(context.Background())
		convey.Convey("When NPU num <= maxNPUsPerBatch", func() {
			signal := job.JobSummarySignal{
				JobId:    testJob1,
				HcclJson: `{"rank_id":0},{"rank_id":1}`,
			}
			err := server.handleSingleJobInfo(&signal)
			if err != nil {
				t.Fatalf("handleSingleJobInfo failed: %v", err)
			}
			convey.So(signal.HcclJson, convey.ShouldEqual, signal.HcclJson)
		})
		convey.Convey("When NPU num > maxNPUsPerBatch", func() {
			signal := job.JobSummarySignal{
				JobId:    testJob1,
				HcclJson: `{"rank_id":0}` + strings.Repeat(`{"rank_id":1},`, constant.MaxNPUsPerBatch+1),
			}
			err := server.handleSingleJobInfo(&signal)
			if err != nil {
				t.Fatalf("handleSingleJobInfo failed: %v", err)
			}
			convey.So(signal.HcclJson, convey.ShouldBeEmpty)
		})
	})
}

func TestManageClientContext(t *testing.T) {
	convey.Convey("Given a clientState", t, func() {
		ctx := context.Background()
		server := NewJobServer(ctx)
		clientID := registerTestClient(server, ctx, CCAgentClientName)
		server.mu.RLock()
		cltState := server.clients[clientID]
		server.mu.RUnlock()
		convey.Convey("When manage new context", func() {
			newCtx, cancel := context.WithCancel(ctx)
			server.manageClientContext(cltState, newCtx, clientID)
			convey.So(cltState.ctx, convey.ShouldNotBeNil)
			convey.So(atomic.LoadInt32(&cltState.ctxCount), convey.ShouldEqual, 1)
			anotherCtx, anotherCancel := context.WithCancel(ctx)
			server.manageClientContext(cltState, anotherCtx, clientID)
			convey.So(cltState.ctx, convey.ShouldNotBeNil)
			convey.So(atomic.LoadInt32(&cltState.ctxCount), convey.ShouldEqual, two)
			select {
			case <-newCtx.Done():
				convey.So(newCtx.Err(), convey.ShouldEqual, context.Canceled)
			default:
				hwlog.RunLog.Errorf("old context should be canceled")
			}
			cancel()
			anotherCancel()
		})
	})
}

func TestCleanupClientContext(t *testing.T) {
	convey.Convey("Given a clientState with ctxCount", t, func() {
		ctx := context.Background()
		server := NewJobServer(ctx)
		clientID := registerTestClient(server, ctx, CCAgentClientName)
		server.mu.RLock()
		cltState := server.clients[clientID]
		server.mu.RUnlock()
		convey.Convey("When ctxCount > 0 after cleanup", func() {
			atomic.StoreInt32(&cltState.ctxCount, two)
			server.cleanupClientContext(cltState, clientID)
			convey.So(atomic.LoadInt32(&cltState.ctxCount), convey.ShouldEqual, 1)
			server.mu.RLock()
			convey.So(server.clients[clientID], convey.ShouldNotBeNil)
			server.mu.RUnlock()
		})
		convey.Convey("When ctxCount <= 0 after cleanup", func() {
			atomic.StoreInt32(&cltState.ctxCount, 1)
			atomic.StoreInt32(&cltState.ctxCount, 0)
			server.cleanupClientContext(cltState, clientID)
			server.mu.RLock()
			convey.So(server.clients[clientID], convey.ShouldBeNil)
			server.mu.RUnlock()
			convey.So(cltState.closed, convey.ShouldBeTrue)
		})
	})
}

// Helper function to register a test client
func registerTestClient(server *JobServer, ctx context.Context, role string) string {
	req := &job.ClientInfo{Role: role}
	resp, err := server.Register(ctx, req)
	if err != nil {
		panic(fmt.Sprintf("register test client failed: %v", err))
	}
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

type MockJobSummarySignalListServer struct {
	grpc.ServerStream
	sentSignals []*job.JobSummarySignalList
	sendErr     error
	ctx         context.Context
	mu          sync.Mutex
}

func NewMockJobSummarySignalListServer(ctx context.Context, sendErr error) *MockJobSummarySignalListServer {
	return &MockJobSummarySignalListServer{
		sentSignals: make([]*job.JobSummarySignalList, 0),
		sendErr:     sendErr,
		ctx:         ctx,
	}
}

func (m *MockJobSummarySignalListServer) Send(signals *job.JobSummarySignalList) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sentSignals = append(m.sentSignals, signals)
	return nil
}

func (m *MockJobSummarySignalListServer) Context() context.Context {
	return m.ctx
}

func (m *MockJobSummarySignalListServer) GetSentSignals() []*job.JobSummarySignalList {
	m.mu.Lock()
	defer m.mu.Unlock()
	copySignals := make([]*job.JobSummarySignalList, len(m.sentSignals))
	copy(copySignals, m.sentSignals)
	return copySignals
}

func (m *MockJobSummarySignalListServer) GetSendCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sentSignals)
}

func fakeJobInfoWithNPU(npuNum int, jobKey string) constant.JobInfo {
	deviceList := make([]constant.Device, 0, npuNum)
	for i := 0; i < npuNum; i++ {
		deviceList = append(deviceList, constant.Device{DeviceID: strconv.Itoa(i), RankID: strconv.Itoa(i)})
	}
	return constant.JobInfo{
		Key: jobKey,
		JobRankTable: constant.RankTable{
			ServerList: []constant.ServerHccl{{DeviceList: deviceList}},
		},
		Framework: ptFramework,
		Status:    StatusJobCompleted,
		AddTime:   time.Now().Unix(),
	}
}

type ProcessJobSliceForBatchSignalsTestCase struct {
	name             string
	jobMap           map[string]constant.JobInfo
	maxNPUs          int
	wantBatchSignals int
	wantBatchJobIds  [][]string
	wantHcclEmpty    []string
}

func buildProcessJobSliceForBatchSignalsTestCases1() []ProcessJobSliceForBatchSignalsTestCase {
	tests := []ProcessJobSliceForBatchSignalsTestCase{
		{
			name:             "empty job map",
			jobMap:           map[string]constant.JobInfo{},
			maxNPUs:          constant.MaxNPUsPerBatch,
			wantBatchSignals: 0,
			wantBatchJobIds:  nil,
			wantHcclEmpty:    nil,
		},
		{
			name: "single job with NPU > max",
			jobMap: map[string]constant.JobInfo{
				fmt.Sprintf("%s-%d", jobName1, fiftyThousand): fakeJobInfoWithNPU(fiftyThousand,
					fmt.Sprintf("%s-%d", jobName1, fiftyThousand)),
			},
			maxNPUs:          constant.MaxNPUsPerBatch,
			wantBatchSignals: 1,
			wantBatchJobIds:  [][]string{{fmt.Sprintf("%s-%d", jobName1, fiftyThousand)}},
			wantHcclEmpty:    []string{fmt.Sprintf("%s-%d", jobName1, fiftyThousand)},
		},
		{
			name: "single job with NPU = max",
			jobMap: map[string]constant.JobInfo{
				fmt.Sprintf("%s-%d", jobName1, constant.MaxNPUsPerBatch): fakeJobInfoWithNPU(
					constant.MaxNPUsPerBatch, fmt.Sprintf("%s-%d", jobName1, constant.MaxNPUsPerBatch)),
			},
			maxNPUs:          constant.MaxNPUsPerBatch,
			wantBatchSignals: 1,
			wantBatchJobIds:  [][]string{{fmt.Sprintf("%s-%d", jobName1, constant.MaxNPUsPerBatch)}},
			wantHcclEmpty:    nil,
		},
	}
	return tests
}

func buildProcessJobSliceForBatchSignalsTestCases2() []ProcessJobSliceForBatchSignalsTestCase {
	tests := []ProcessJobSliceForBatchSignalsTestCase{
		{
			name: "multiple jobs with total NPU ≤ max",
			jobMap: map[string]constant.JobInfo{
				fmt.Sprintf("%s-%d", jobName1, fifteenThousand): fakeJobInfoWithNPU(fifteenThousand,
					fmt.Sprintf("%s-%d", jobName1, fifteenThousand)),
				fmt.Sprintf("%s-%d", jobName2, fifteenThousand): fakeJobInfoWithNPU(fifteenThousand,
					fmt.Sprintf("%s-%d", jobName2, fifteenThousand)),
			},
			maxNPUs:          constant.MaxNPUsPerBatch,
			wantBatchSignals: 1,
			wantBatchJobIds: [][]string{{fmt.Sprintf("%s-%d", jobName1, fifteenThousand),
				fmt.Sprintf("%s-%d", jobName2, fifteenThousand)}},
			wantHcclEmpty: nil,
		},
		{
			name: "multiple jobs with accumulated NPU > max",
			jobMap: map[string]constant.JobInfo{
				fmt.Sprintf("%s-%d", jobName1, thirtyFiveThousand): fakeJobInfoWithNPU(thirtyFiveThousand,
					fmt.Sprintf("%s-%d", jobName1, thirtyFiveThousand)),
				fmt.Sprintf("%s-%d", jobName2, fifteenThousand): fakeJobInfoWithNPU(fifteenThousand,
					fmt.Sprintf("%s-%d", jobName2, fifteenThousand)),
			},
			maxNPUs:          constant.MaxNPUsPerBatch,
			wantBatchSignals: 2,
			wantBatchJobIds: [][]string{{fmt.Sprintf("%s-%d", jobName1, thirtyFiveThousand)},
				{fmt.Sprintf("%s-%d", jobName2, fifteenThousand)}},
			wantHcclEmpty: nil,
		},
		{
			name: "mixed scenario (single over + accumulated over)",
			jobMap: map[string]constant.JobInfo{
				fmt.Sprintf("%s-%d", jobName1, fiftyThousand): fakeJobInfoWithNPU(fiftyThousand,
					fmt.Sprintf("%s-%d", jobName1, fiftyThousand)),
				fmt.Sprintf("%s-%d", jobName2, thirtyFiveThousand): fakeJobInfoWithNPU(thirtyFiveThousand,
					fmt.Sprintf("%s-%d", jobName2, thirtyFiveThousand)),
				fmt.Sprintf("%s-%d", jobName3, fifteenThousand): fakeJobInfoWithNPU(fifteenThousand,
					fmt.Sprintf("%s-%d", jobName3, fifteenThousand)),
			},
			maxNPUs:          constant.MaxNPUsPerBatch,
			wantBatchSignals: three,
			wantBatchJobIds: [][]string{
				{fmt.Sprintf("%s-%d", jobName1, fiftyThousand)},
				{fmt.Sprintf("%s-%d", jobName2, thirtyFiveThousand)},
				{fmt.Sprintf("%s-%d", jobName3, fifteenThousand)}},
			wantHcclEmpty: []string{fmt.Sprintf("%s-%d", jobName1, fiftyThousand)},
		},
	}
	return tests
}

func TestProcessJobSliceForBatchSignals(t *testing.T) {
	convey.Convey("Test processJobSliceForBatchSignals with all scenarios", t, func() {
		patchCalcNPU := gomonkey.ApplyFunc(calcJobNPUNum, func(jobInfo constant.JobInfo) int {
			parts := strings.Split(jobInfo.Key, "-")
			if len(parts) < two {
				return 0
			}
			npuNum, err := strconv.Atoi(parts[1])
			if err != nil {
				return 0
			}
			return npuNum
		})
		defer patchCalcNPU.Reset()
		tests := buildProcessJobSliceForBatchSignalsTestCases1()
		tests = append(tests, buildProcessJobSliceForBatchSignalsTestCases2()...)
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				batchSignals, batchJobIds := processJobSliceForBatchSignals(tt.jobMap, tt.maxNPUs)
				convey.So(len(batchSignals), convey.ShouldEqual, tt.wantBatchSignals)
				convey.So(len(batchJobIds), convey.ShouldEqual, tt.wantBatchSignals)
				for i, jobIds := range tt.wantBatchJobIds {
					convey.So(len(batchJobIds[i]), convey.ShouldEqual, len(jobIds))
				}
			})
		}
	})
}

type GetAllBatchJobSummarySignalTestCase struct {
	name             string
	mockJobCache     map[string]constant.JobInfo
	wantBatchSignals int
	wantBatchJobIds  int
	wantJobTotalNum  int32
}

func buildGetAllBatchJobSummarySignalTestCases() []GetAllBatchJobSummarySignalTestCase {
	return []GetAllBatchJobSummarySignalTestCase{
		{
			name:             "empty job cache",
			mockJobCache:     map[string]constant.JobInfo{},
			wantBatchSignals: 0,
			wantBatchJobIds:  0,
			wantJobTotalNum:  0,
		},
		{
			name: "single job in cache",
			mockJobCache: map[string]constant.JobInfo{
				jobName1: fakeJobInfoWithNPU(fifteenThousand, jobName1),
			},
			wantBatchSignals: 1,
			wantBatchJobIds:  1,
			wantJobTotalNum:  1,
		},
		{
			name: "multiple jobs in cache",
			mockJobCache: map[string]constant.JobInfo{
				jobName1: fakeJobInfoWithNPU(fifteenThousand, jobName1),
				jobName2: fakeJobInfoWithNPU(fiftyThousand, jobName2),
				jobName3: fakeJobInfoWithNPU(constant.MaxNPUsPerBatch, jobName3),
			},
			wantBatchSignals: three,
			wantBatchJobIds:  three,
			wantJobTotalNum:  three,
		},
	}
}

func TestGetAllBatchJobSummarySignals(t *testing.T) {
	convey.Convey("Test GetAllBatchJobSummarySignals with all scenarios", t, func() {
		mockProcess := func(jobMap map[string]constant.JobInfo, maxNPUs int) ([][]*job.JobSummarySignal, [][]string) {
			batchSignals := make([][]*job.JobSummarySignal, 0)
			batchJobIds := make([][]string, 0)
			for jobId := range jobMap {
				batchSignals = append(batchSignals, []*job.JobSummarySignal{{JobId: jobId}})
				batchJobIds = append(batchJobIds, []string{jobId})
			}
			return batchSignals, batchJobIds
		}
		patchProcess := gomonkey.ApplyFunc(processJobSliceForBatchSignals, mockProcess)
		defer patchProcess.Reset()

		tests := buildGetAllBatchJobSummarySignalTestCases()
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				patchCache := gomonkey.ApplyFuncReturn(jobstorage.GetAllJobCache, tt.mockJobCache)
				defer patchCache.Reset()
				batchSignals, batchJobIds, JobTotalNum := GetAllBatchJobSummarySignals()
				convey.So(len(batchSignals), convey.ShouldEqual, tt.wantBatchSignals)
				convey.So(len(batchJobIds), convey.ShouldEqual, tt.wantBatchJobIds)
				convey.So(JobTotalNum, convey.ShouldEqual, tt.wantJobTotalNum)
			})
		}
	})
}
