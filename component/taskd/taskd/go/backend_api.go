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
	"strconv"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager"
	"taskd/framework_backend/manager/application"
	"taskd/framework_backend/proxy"
	"taskd/framework_backend/worker"
	"taskd/framework_backend/worker/om"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

var ctx context.Context = context.Background()
var netLifeCtl = make(map[uintptr]*net.NetInstance)
var rw sync.RWMutex
var logLock sync.RWMutex
var managerInstance = &manager.BaseManager{}
var loggerLifeCtl = make(map[uintptr]*hwlog.CustomLogger)

// RegisterSwitchCallback register switch callback, is called by om worker
//
//export RegisterSwitchCallback
func RegisterSwitchCallback(cb uintptr) {
	om.RegisterSwitchNicCallback(cb)
}

// RegisterStressTestCallback register stress test callback, is called by worker
//
//export RegisterStressTestCallback
func RegisterStressTestCallback(cb uintptr) {
	om.RegisterStressTestCallback(cb)
}

// InitWorker to init worker, should be called by python api,
// and this python api will be called in user script
// globalRank: the global rank of current process
// nodeRank: the node rank
// upperLimitOfDiskInMb: is the upper limit of disk usage
//
//export InitWorker
func InitWorker(globalRank, nodeRank, upperLimitOfDiskInMb int) C.int {
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
	err := utils.InitHwLog(fmt.Sprintf(constant.WorkerLogPathPattern, strconv.Itoa(globalRank)), ctx)
	if err != nil {
		fmt.Println(err)
	}
	go worker.InitMonitor(ctx, globalRank, upperLimitOfDiskInMb)
	go worker.InitNetwork(globalRank, nodeRank)
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
	hwlog.RunLog.Infof("rank %d will start its client", worker.GlobalRank)
	go worker.StartMonitor(ctx)
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
func InitNetwork(configJSON *C.char, loggerPtr uintptr) uintptr {
	configStr := C.GoString(configJSON)
	var config common.TaskNetConfig
	err := json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		return 0
	}
	logLock.RLock()
	logger, ok := loggerLifeCtl[loggerPtr]
	logLock.RUnlock()
	if !ok {
		return 0
	}
	tool, err := net.InitNetwork(&config, logger)
	if err != nil {
		return 0
	}
	logger.Infof("init network success, role=%s, srvRank=%s, processRank=%s",
		config.Pos.Role, config.Pos.ServerRank, config.Pos.ProcessRank)
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
	if goMessage.Dst != nil {
		tool.GetNetworkerLogger().Infof("py sync send message, dstRole=%s, dstSrvRank=%s, DstProcessRank=%s",
			goMessage.Dst.Role, goMessage.Dst.ServerRank, goMessage.Dst.ProcessRank)
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
	if goMessage.Dst != nil {
		tool.GetNetworkerLogger().Infof("py async send message, dstRole=%s, dstSrvRank=%s, DstProcessRank=%s",
			goMessage.Dst.Role, goMessage.Dst.ServerRank, goMessage.Dst.ProcessRank)
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
		rw.RUnlock()
		return unsafe.Pointer(nil)
	}
	rw.RUnlock()
	msg := tool.ReceiveMessage()
	if msg == nil {
		return unsafe.Pointer(nil)
	}

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

//export CreateTaskdLog
func CreateTaskdLog(logName *C.char) uintptr {
	logFileName := C.GoString(logName) // convert C string to g
	hwLogConfig, err := utils.GetLoggerConfigWithFileName(logFileName)
	if err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return 0
	}
	logger, err := hwlog.NewCustomLogger(hwLogConfig, context.Background())
	if err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return 0
	}
	loggerPtr := uintptr(unsafe.Pointer(logger))
	logLock.Lock()
	loggerLifeCtl[loggerPtr] = logger
	logLock.Unlock()
	return loggerPtr
}

//export SendMessageToBackend
func SendMessageToBackend(msgJSON *C.char) C.int {
	msg := C.GoString(msgJSON)
	var goMessage constant.ControllerMessage
	err := json.Unmarshal([]byte(msg), &goMessage)
	if err != nil {
		return C.int(-1)
	}
	res := manager.ReportControllerInfoToClusterd(&goMessage)
	if res != true {
		return C.int(1)
	}
	return C.int(0)
}

//export RegisterBackendCallback
func RegisterBackendCallback(cb uintptr) {
	application.RegisterControllerCallback(cb)
}

func main() {
}
