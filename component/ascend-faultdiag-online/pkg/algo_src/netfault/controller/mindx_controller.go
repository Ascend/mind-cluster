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

	"ascend-faultdiag-online/pkg/core/model"
)

const (
	configFile = "cathelper.conf"
)

var clusterLevelPath = os.Getenv("RAS_NET_ROOT_PATH") + "/cluster"

var callbackFunc model.CallbackFunc = nil

// Start controller
func Start() {
	startController(clusterLevelPath)
}

// Reload controller
func Reload() {
	reloadController(clusterLevelPath)
}

// Stop controller 仅同步调用
func Stop() {
	stopController()
}

// RegisterDetectionCallback register detect callback
func RegisterDetectionCallback(callback model.CallbackFunc) {
	if callbackFunc == nil {
		return
	}
	callbackFunc = callback
}
