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

// Package main a main package for cgo api
package main

import (
	"C"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/worker/monitor/profiling"
)

var ctx context.Context = context.Background()

// InitTaskMonitor to init tasdD monitor, should be called by python api,
// and this python api will be called in user script
// rank: the global rank of current process, upperLimitOfDiskInMb is the upper limit of disk usage
//
//export InitTaskMonitor
func InitTaskMonitor(rank int, upperLimitOfDiskInMb int) C.int {
	profiling.SetDiskUsageUpperLimitMB(upperLimitOfDiskInMb)
	profiling.GlobalRankId = rank
	// init so should not use print to avoid impact on system calls
	err := utils.InitHwLog(ctx)
	if err != nil {
		fmt.Println(err)
		return C.int(1)
	}
	if err := profiling.InitMspti(); err != nil {
		hwlog.RunLog.Error(err)
		return C.int(1)
	}
	hwlog.RunLog.Info("successfully init mspti lib so")
	// listen to system signal
	sigChan := make(chan os.Signal, 1)
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		hwlog.RunLog.Errorf("Received signal: %v, exiting...\n", sig)
		cancel()
	}()
	return C.int(0)
}

// StartMonitorClient this function is the entrance for monitoring, is called by user through python api
//
//export StartMonitorClient
func StartMonitorClient() C.int {
	defer func() {
		if r := recover(); r != nil {
			hwlog.RunLog.Errorf("start taskd monitor panicked, all taskd monitor function is disabled: %v", r)
			fmt.Printf("[ERROR] %s start taskd monitor panicked, all taskd monitor"+
				" function is disabled: %v\n", time.Now(), r)
		}
	}()
	hwlog.RunLog.Infof("rank %d will start its client", profiling.GlobalRankId)
	if err := profiling.MsptiActivityRegisterCallbacksWrapper(); err != nil {
		return C.int(1)
	}
	go profiling.ManageSaveProfiling(ctx)
	go profiling.ManageDomainEnableStatus(ctx)
	go profiling.ManageProfilingDiskUsage(constant.ProfilingBaseDir, ctx)
	profiling.ProfilingTaskQueue = profiling.NewTaskQueue(ctx)

	return C.int(0)
}

func main() {
}
