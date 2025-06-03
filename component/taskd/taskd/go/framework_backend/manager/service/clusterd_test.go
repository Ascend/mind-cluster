package service

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

// Package service is to provide other service tools, i.e. clusterd
import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/grpc/profiling"
)

const (
	jobId              = "1234"
	watchCheckInterval = 2 * time.Second
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	return initLog()
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return err
	}
	return nil
}

func TestGetClusterDAddr(t *testing.T) {
	convey.Convey("TestGetClusterDAddr", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(os.Getenv, func(key string) string {
			return ""
		})
		addr := getClusterDAddr()
		convey.So(addr, convey.ShouldEqual, ":"+clusterdPort)
	})
}

func TestRegisterClusterD(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(grpc.Dial, func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return nil, nil
	})

	callSubscribeProfiling := false
	patches.ApplyFunc(subscribeProfiling,
		func(jobId string, ctx context.Context, conn *grpc.ClientConn, retryTime time.Duration) {
			callSubscribeProfiling = true
		})
	time.Sleep(time.Second)
	convey.ShouldBeTrue(callSubscribeProfiling)
}

type mockProfilingStream struct {
	recv  *profiling.DataStatusRes
	ctx   context.Context
	mutex sync.Mutex
}

func (s *mockProfilingStream) Recv() (*profiling.DataStatusRes, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.recv = &profiling.DataStatusRes{Code: 1}
	time.Sleep(time.Second)
	return s.recv, nil
}
func (s *mockProfilingStream) getRecvCode() int32 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.recv.Code
}

func (s *mockProfilingStream) Header() (metadata.MD, error) { return nil, nil }

func (s *mockProfilingStream) Trailer() metadata.MD { return nil }

func (s *mockProfilingStream) CloseSend() error { return nil }

func (s *mockProfilingStream) Context() context.Context { return s.ctx }

func (s *mockProfilingStream) SendMsg(m interface{}) error { return nil }

func (s *mockProfilingStream) RecvMsg(m interface{}) error { return nil }

func TestSubscribeProfiling(t *testing.T) {
	convey.Convey("TestSubscribeProfiling", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		mockClient := profiling.NewTrainingDataTraceClient(nil)
		patches.ApplyFunc(profiling.NewTrainingDataTraceClient,
			func(cc grpc.ClientConnInterface) profiling.TrainingDataTraceClient {
				return mockClient
			})

		ctx, cancelFunc := context.WithCancel(context.Background())
		stream := &mockProfilingStream{ctx: ctx}
		patches.ApplyMethod(mockClient, "SubscribeDataTraceSwitch", func(profiling.TrainingDataTraceClient,
			context.Context, *profiling.ProfilingClientInfo, ...grpc.CallOption) (
			profiling.TrainingDataTrace_SubscribeDataTraceSwitchClient, error) {
			return stream, nil
		})
		go func() {
			time.Sleep(time.Millisecond)
			cancelFunc()
		}()
		subscribeProfiling(jobId, ctx, nil, 0)
		convey.So(stream.getRecvCode(), convey.ShouldEqual, 1)
	})
}

func TestWatchProfilingSwitchChange(t *testing.T) {
	convey.Convey("TestWatchProfilingSwitchChange", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		called := atomic.Bool{}
		patches.ApplyFunc(getProfilingFromFile, func() {
			called.Store(true)
		})

		ctx, cancelFunc := context.WithCancel(context.Background())
		profilingFromClusterD.Store(false)
		go func() {
			cancelFunc()
		}()
		watchProfilingSwitchChange(ctx)
		convey.ShouldBeTrue(called.Load())
	})
}
