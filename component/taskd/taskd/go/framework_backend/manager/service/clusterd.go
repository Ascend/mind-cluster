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
package service

import (
	"context"
	"os"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/grpc/profiling"
)

var profilingFromClusterD = atomic.Bool{}

const (
	clusterdPort    = "8899"
	roleTaskd       = "taskd"
	ok              = 0
	maxRegRetryTime = 10
)

func registerClusterD(jobId string, ctx context.Context, retryTime time.Duration) {
	if retryTime >= maxRegRetryTime {
		hwlog.RunLog.Error("init clusterd connect meet max retry time")
		return
	}
	time.Sleep(retryTime * time.Second)
	conn, err := grpc.Dial(getClusterDAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		hwlog.RunLog.Error("init clusterd connect failed")
		registerClusterD(jobId, ctx, retryTime+1)
		return
	}

	go subscribeProfiling(jobId, ctx, conn, 0)
}

func subscribeProfiling(jobId string, ctx context.Context, conn *grpc.ClientConn, retryTime time.Duration) {
	profilingFromClusterD.Store(false)
	if retryTime >= maxRegRetryTime {
		hwlog.RunLog.Error("register Cluster profiling meet max retry time")
		return
	}
	time.Sleep(retryTime * time.Second)
	traceClient := profiling.NewTrainingDataTraceClient(conn)
	stream, err := traceClient.SubscribeDataTraceSwitch(ctx, &profiling.ProfilingClientInfo{
		JobId: jobId,
		Role:  roleTaskd,
	})
	if err != nil {
		hwlog.RunLog.Errorf("register Cluster profiling fail, err: %v", err)
		go subscribeProfiling(jobId, ctx, conn, retryTime+1)
		return
	}
	profilingFromClusterD.Store(true)
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Infof("taskd exit, stop subscribe clusterd fault info")
			return
		case <-stream.Context().Done():
			hwlog.RunLog.Infof("client stream exit, stop subscribe profiling info and re-register")
			go subscribeProfiling(jobId, ctx, conn, retryTime+1)
			return
		default:
			responseMsg, recvErr := stream.Recv()
			if recvErr != nil {
				hwlog.RunLog.Error(recvErr)
			} else {
				hwlog.RunLog.Infof("receive profiling info: %v", responseMsg)
				profilingMsg := responseMsg.GetProfilingSwitch()
				// notify framework receive profiling msg
				enqueueProfilingSwitch(profilingMsg, "Clusterd")
			}
		}
	}
}

func enqueueProfilingSwitch(profilingSwitch any, whichServer string) {
}

func getClusterDAddr() string {
	return os.Getenv("MINDX_SERVER_IP") + ":" + clusterdPort
}

func watchProfilingSwitchChange(ctx context.Context) {
	hwlog.RunLog.Info("begin watch ProfilingSwitchFilePath...")
	ticker := time.Tick(time.Second)
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("end watch ProfilingSwitchFilePath...")
			return
		case <-ticker:
			if profilingFromClusterD.Load() {
				hwlog.RunLog.Infof("manager register clusterd, donot watch profiling file.")
				return
			}
			getProfilingFromFile()
		}
	}
}

func getProfilingFromFile() {
	return
}
