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

// Package controller
package controller

import (
	"os"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controllerflags"
	"ascend-faultdiag-online/pkg/core/model"
	"ascend-faultdiag-online/pkg/utils/fileutils"
)

/* controller interface sync lock（公共锁保证start和stop的同步） */
var controllerSyncOperatorLock sync.Mutex

var controllerExitCond *sync.Cond = sync.NewCond(&controllerSyncOperatorLock)

const (
	confFileRetryTime = 30
	configFile        = "cathelper.conf"
)

/* go routine started controller */
func startController(clusterPath string) {
	/* check invalid */
	if clusterPath == "/cluster" {
		hwlog.RunLog.Errorf("empty input : no root dir")
		controllerflags.IsControllerExited.SetState(true)
		if controllerExitCond != nil {
			controllerExitCond.Signal()
		}
		return
	}
	/* check directory exist */
	for i := 0; i <= confFileRetryTime && !controllerflags.IsControllerExited.GetState(); i++ {
		_, err := os.Stat(clusterPath)
		if err == nil {
			// 等待成功
			break
		}
		hwlog.RunLog.Errorf("waiting for file creating: %v and retry times : %d", err, i)
		if i == confFileRetryTime {
			// 等待失败
			hwlog.RunLog.Errorf("path not exist: %s", clusterPath)
			controllerflags.IsControllerExited.SetState(true)
			if controllerExitCond != nil {
				controllerExitCond.Signal()
			}
			return
		}
		time.Sleep(time.Second)
	}
	if controllerflags.IsControllerExited.GetState() {
		if controllerExitCond != nil {
			controllerExitCond.Signal()
		}
		return
	}
	startSuperPodsDetectionAsync(clusterPath)
}

/* stop detection:仅同步调用 */
func stopController() {
	controllerSyncOperatorLock.Lock()
	defer controllerSyncOperatorLock.Unlock()
	if controllerExitCond != nil {
		controllerExitCond.Wait()
	}
	hwlog.RunLog.Info("controller has been stopped")
}

/* stop first, start then */
func reloadController(clusterPath string) {
	stopController()
	startController(clusterPath)
}

var clusterLevelPath = os.Getenv("RAS_NET_ROOT_PATH") + "/cluster"

var callbackFunc model.CallbackFunc = nil

// Start controller
func Start() {
	absPath, err := fileutils.CheckPath(clusterLevelPath)
	if err != nil {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]clusterLevelPath %s invalid, err: %v", clusterLevelPath, err)
		return
	}
	startController(absPath)
}

// Reload controller
func Reload() {
	absPath, err := fileutils.CheckPath(clusterLevelPath)
	if err != nil {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]clusterLevelPath %s invalid, err: %v", clusterLevelPath, err)
		return
	}
	reloadController(absPath)
}

// Stop controller 仅同步调用
func Stop() {
	stopController()
}

// RegisterDetectionCallback register detect callback
func RegisterDetectionCallback(callback model.CallbackFunc) {
	if callback == nil {
		return
	}
	callbackFunc = callback
}
