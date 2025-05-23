//go:build !fdol

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

// Package fdapi for a series of function of fd controller
package fdapi

import (
	"ascend-common/common-utils/hwlog"
)

// StartFdOL start fd-ol
func StartFdOL() {
	hwlog.RunLog.Warn("start fd-ol not support, please build with fd-ol tags")
}

// StartController to start controller
func StartController() {
	hwlog.RunLog.Warn("start controller not support, please build with fd-ol tags")
}

// StopController to stop controller
func StopController() {
	hwlog.RunLog.Warn("stop controller not support, please build with fd-ol tags")
}

// ReloadController to reload controller
func ReloadController() {
	hwlog.RunLog.Warn("reload controller not support, please build with fd-ol tags")
}
