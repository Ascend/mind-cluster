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

// Package constant a package for constant
package constant

import (
	"ascend-common/common-utils/hwlog"
)

// NewProfilingExecRes construct ProfilingExecRes from string
func NewProfilingExecRes(status string) ProfilingExecRes {
	switch status {
	case On, Off, Unknown, Exp:
		return ProfilingExecRes{status: status}
	default:
		hwlog.RunLog.Errorf("invalid status: %s", status)
		return ProfilingExecRes{status: Unknown}
	}
}

// NewWorkerProfilingState construct ProfilingWorkerState from string
func NewWorkerProfilingState(state string) ProfilingWorkerState {
	switch state {
	case Opened, Closed, WaitOpen, Exception, WaitClose:
		return ProfilingWorkerState{state: state}
	default:
		hwlog.RunLog.Errorf("invalid state: %s", state)
		return ProfilingWorkerState{state: Invalid}
	}
}
