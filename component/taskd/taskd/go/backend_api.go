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

/*
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager"
	"taskd/framework_backend/proxy"
	"taskd/framework_backend/worker/monitor/profiling"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

var ctx context.Context = context.Background()
var netLifeCtl = make(map[uintptr]*net.NetInstance)
var rw sync.RWMutex
var managerInstance = &manager.BaseManager{}

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
		hwlog.RunLog.Errorf("Received signal: %v, exiting...", sig)
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
	profiling.ProfileTaskQueue = profiling.NewTaskQueue(ctx)

	return C.int(0)
}

// StepOut this function return whether step out, is called by user through python api
//
//export StepOut
func StepOut() C.int {
	if profiling.StepOut() {
		return C.int(1)
	}
	return C.int(0)
}

// InitTaskdManager this function is the entrance for initialize taskd manager, is called by user through python api
//
//export InitTaskdManager
func InitTaskdManager(configStr *C.char) C.int {
	var config manager.Config
	if err := json.Unmarshal([]byte(C.GoString(configStr)), &config); err != nil {
		return C.int(1)
	}
	managerInstance = manager.NewTaskDManager(config)
	return C.int(0)
}

// StartTaskdManager this function is the entrance for start taskd manager, is called by user through python api
//
//export StartTaskdManager
func StartTaskdManager() C.int {
	if err := managerInstance.Start(); err != nil {
		return C.int(1)
	}
	return C.int(0)
}

//export InitNetwork
func InitNetwork(configJSON *C.char) uintptr {
	configStr := C.GoString(configJSON)
	var config common.TaskNetConfig
	err := json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		return 0
	}
	tool, err := net.InitNetwork(&config)
	if err != nil {
		return 0
	}
	toolPtr := uintptr(unsafe.Pointer(tool))
	rw.Lock()
	netLifeCtl[toolPtr] = tool
	rw.Unlock()
	return toolPtr
}

//export SyncSendMessage
func SyncSendMessage(toolPtr uintptr, msgJSON *C.char) C.int {
	var tool *net.NetInstance
	msg := C.GoString(msgJSON)
	var goMessage common.Message
	err := json.Unmarshal([]byte(msg), &goMessage)
	if err != nil {
		return C.int(-1)
	}
	rw.RLock()
	defer rw.RUnlock()
	tool, ok := netLifeCtl[toolPtr]
	if !ok {
		return C.int(-1)
	}
	_, err = tool.SyncSendMessage(goMessage.Uuid, goMessage.BizType, goMessage.Body, goMessage.Dst)
	if err != nil {
		return C.int(-1)
	}
	return C.int(0)
}

//export AsyncSendMessage
func AsyncSendMessage(toolPtr uintptr, msgJSON *C.char) C.int {
	var tool *net.NetInstance
	msg := C.GoString(msgJSON)
	var goMessage common.Message
	err := json.Unmarshal([]byte(msg), &goMessage)
	if err != nil {
		return C.int(-1)
	}
	rw.RLock()
	defer rw.RUnlock()
	tool, ok := netLifeCtl[toolPtr]
	if !ok {
		return C.int(-1)
	}
	err = tool.AsyncSendMessage(goMessage.Uuid, goMessage.BizType, goMessage.Body, goMessage.Dst)
	if err != nil {
		return C.int(-1)
	}
	return C.int(0)
}

//export ReceiveMessageC
func ReceiveMessageC(toolPtr uintptr) unsafe.Pointer {
	rw.RLock()
	tool, ok := netLifeCtl[toolPtr]
	if !ok {
		return unsafe.Pointer(nil)
	}

	msg := tool.ReceiveMessage()
	if msg == nil {
		return unsafe.Pointer(nil)
	}
	rw.RUnlock()

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return unsafe.Pointer(nil)
	}

	msgCStr := C.CString(string(msgJSON))
	return unsafe.Pointer(msgCStr)
}

//export DestroyNetTool
func DestroyNetTool(toolPtr uintptr) {
	rw.RLock()
	defer rw.RUnlock()
	tool, ok := netLifeCtl[toolPtr]
	if !ok {
		return
	}
	tool.Destroy()
}

//export FreeCMemory
func FreeCMemory(ptr unsafe.Pointer) {
	if ptr != nil {
		C.free(ptr)
	}
}

// InitTaskdProxy to init tasdD proxy, should be called by taskd agent python api
//
//export InitTaskdProxy
func InitTaskdProxy(configJson *C.char) C.int {
	if configJson == nil {
		return C.int(1)
	}
	configStr := C.GoString(configJson)
	var config common.TaskNetConfig
	err := json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		return C.int(1)
	}

	err = proxy.InitProxy(&config)
	if err != nil {
		return C.int(1)
	}
	return C.int(0)
}

// DestroyTaskdProxy to destroy tasdD proxy, should be called by taskd agent python api
//
//export DestroyTaskdProxy
func DestroyTaskdProxy() {
	proxy.DestroyProxy()
}

//export InitTaskdLog
func InitTaskdLog(logName *C.char) C.int {
	logStr := C.GoString(logName) // convert C string to g
	if err := utils.InitHwLogger(logStr, context.Background()); err != nil {
		fmt.Printf("%s init hwlog failed, err: %v", logStr, err)
		return C.int(1)
	}
	return C.int(0)
}

func main() {
}
