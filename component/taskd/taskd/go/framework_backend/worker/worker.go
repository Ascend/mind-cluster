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

// Package worker for taskd worker backend
package worker

import "C"
import (
	"context"
	"os"
	"strconv"
	"time"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/framework_backend/worker/monitor/profiling"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

var netTool *net.NetInstance

var monitorInitCtx context.Context
var monitorInitNotify context.CancelFunc

const (
	waitInitMsptiTimeout = 60 * time.Second
)

func init() {
	monitorInitCtx, monitorInitNotify = context.WithCancel(context.Background())
}

// InitMonitor to init taskd monitor,
func InitMonitor(ctx context.Context, globalRank int, upperLimitOfDiskInMb int) {
	profiling.SetDiskUsageUpperLimitMB(upperLimitOfDiskInMb)
	profiling.GlobalRankId = globalRank

	hwlog.RunLog.Info("begin init mspti lib so")
	if err := profiling.InitMspti(); err != nil {
		hwlog.RunLog.Errorf("init profiling err: %v", err)
		return
	}
	profiling.MgrProfilingCmd.Store(false)
	hwlog.RunLog.Info("successfully init mspti lib so")
	monitorInitNotify()
}

// InitNetwork register worker to manager
func InitNetwork(globalRank, nodeRank int) {
	profiling.GlobalRank = globalRank
	profiling.NodeRank = nodeRank
	ip := os.Getenv("POD_IP")
	if ip == "" {
		ip = "127.0.0.1"
	}
	addr := ip + constant.ProxyPort
	var err error
	netTool, err = net.InitNetwork(&common.TaskNetConfig{
		Pos: common.Position{
			Role:        common.WorkerRole,
			ServerRank:  strconv.Itoa(nodeRank),
			ProcessRank: strconv.Itoa(globalRank),
		},
		ListenAddr:   "",
		UpstreamAddr: addr,
		EnableTls:    false,
		TlsConf:      nil,
	})
	if err != nil {
		hwlog.RunLog.Errorf("worker %d init network err: %v", globalRank, err)
	}
	profiling.NetTool = netTool
	profiling.NetToolInitNotify()
}

func waitMonitorInit() bool {
	hwlog.RunLog.Info("wait monitor init")
	select {
	case <-monitorInitCtx.Done():
		hwlog.RunLog.Info("wait monitor inited")
		return true
	case <-time.After(waitInitMsptiTimeout):
		hwlog.RunLog.Info("wait monitor timeout")
		return false
	}
}

func StartMonitor(ctx context.Context) {
	if !waitMonitorInit() {
		hwlog.RunLog.Error("cannot StartMonitor, wait monitor timeout")
		return
	}
	if err := profiling.MsptiActivityRegisterCallbacksWrapper(); err != nil {
		hwlog.RunLog.Errorf("cannot StartMonitor, err: %v", err)
		return
	}
	go profiling.ManageSaveProfiling(ctx)
	go profiling.RegisterAndLoopRecv(ctx)
	go profiling.ManageDomainEnableStatus(ctx)
	go profiling.ManageProfilingDiskUsage(constant.ProfilingBaseDir, ctx)
	profiling.ProfileTaskQueue = profiling.NewTaskQueue(ctx)
}
