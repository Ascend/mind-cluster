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
	"strconv"
	"time"

	"github.com/google/uuid"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/framework_backend/worker/monitor/profiling"
	"taskd/framework_backend/worker/om"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

var netTool *net.NetInstance

var monitorInitCtx context.Context
var monitorInitNotify context.CancelFunc

// NetTool from worker
var NetTool *net.NetInstance

// GlobalRank of this work
var GlobalRank int

// NodeRank of this work
var NodeRank int

var netToolInitCtx context.Context
var netToolInitNotify context.CancelFunc

const (
	waitInitMsptiTimeout = 180 * time.Second
	maxRegisterTime      = 10
	maxWaitNetInitTime   = 180 * time.Second
)

func init() {
	monitorInitCtx, monitorInitNotify = context.WithCancel(context.Background())
	netToolInitCtx, netToolInitNotify = context.WithCancel(context.Background())
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
	hwlog.RunLog.Infof("worker %d noderank %d init network begin", globalRank, nodeRank)
	GlobalRank = globalRank
	NodeRank = nodeRank
	profiling.GlobalRank = globalRank
	profiling.NodeRank = nodeRank
	addr := constant.DefaultIP + constant.ProxyPort
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
	hwlog.RunLog.Infof("worker %d init network end", globalRank)
	NetTool = netTool
	profiling.NetTool = netTool
	om.NetTool = netTool
	netToolInitNotify()
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

func waitNetToolInit() bool {
	hwlog.RunLog.Info("wait NetTool init")
	select {
	case <-netToolInitCtx.Done():
		hwlog.RunLog.Info("wait NetTool inited")
		return true
	case <-time.After(maxWaitNetInitTime):
		hwlog.RunLog.Info("wait NetTool timeout")
		return false
	}
}

func registerAndLoopRecv(ctx context.Context) {
	if !waitNetToolInit() {
		hwlog.RunLog.Error("cannot RegisterAndLoopRecv for profiling, net tool init timeout")
		return
	}
	body := storage.MsgBody{
		MsgType: constant.REGISTER,
		Code:    constant.RegisterCode,
	}
	registerSucc := false
	for i := 0; i < maxRegisterTime; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		_, err := NetTool.SyncSendMessage(uuid.NewString(), "default", utils.ObjToString(body), &common.Position{
			Role:       common.MgrRole,
			ServerRank: "0",
		})
		if err != nil {
			hwlog.RunLog.Errorf("worker %d register manager err: %v", GlobalRank, err)
			continue
		}
		registerSucc = true
		break
	}
	if !registerSucc {
		hwlog.RunLog.Errorf("worker  %d register manager meet max times %d", GlobalRank, maxRegisterTime)
		return
	}
	hwlog.RunLog.Infof("worker %d register manager success, begin recv msg", GlobalRank)
	profiling.MgrProfilingCmd.Store(true)
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Errorf("worker %d exit", GlobalRank)
			return
		default:
			msg := NetTool.ReceiveMessage()
			om.ProcessMsg(GlobalRank, msg)
			profiling.ProcessMsg(GlobalRank, msg)
		}
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
	go registerAndLoopRecv(ctx)
	go profiling.ManageDomainEnableStatus(ctx)
	go profiling.ManageProfilingDiskUsage(constant.ProfilingBaseDir, ctx)
	profiling.ProfileTaskQueue = profiling.NewTaskQueue(ctx)
}
