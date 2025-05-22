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

// Package pingmesh for a series of function handle ping mesh configmap create/update/delete
package pingmesh

import (
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

const controllerUrl = "feature/netfault/controller"

// StartController to start controller
func StartController() {
	requestFD(constant.StartApi)
}

// StopController to stop controller
func StopController() {
	requestFD(constant.StopApi)
}

// ReloadController to reload controller
func ReloadController() {
	requestFD(constant.ReloadApi)
}

func requestFD(api string) {
	hwlog.RunLog.Infof("unsupported")
}
