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

// Package externalbridge for node and cluster level detection interact interface
package externalbridge

import (
	"sync"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controller"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controllerflags"
	"ascend-faultdiag-online/pkg/core/model"
	"ascend-faultdiag-online/pkg/core/model/enum"
)

const (
	successCode      = 0
	errCode          = -1
	minInputLen      = 1
	stopApi          = "stop"
	startApi         = "start"
	reloadApi        = "reload"
	registerCallBack = "registerCallBack"
)

type order struct {
	command      string
	param        any
	callbackFunc model.CallbackFunc
}

var callExecuteSyncLock sync.Mutex

func checkInputInvalid(input *model.Input) bool {
	if input.Command != enum.Start && input.Command != enum.Stop &&
		input.Command != enum.Reload && input.Command != enum.Register {
		hwlog.RunLog.Error("command not support")
		return false
	}

	if input.Command == enum.Register {
		if input.Func == nil {
			hwlog.RunLog.Error("Invalid nil register function")
			return false
		}
	}
	return true
}

// switchCommand execute API has been locked
func switchCommand(input *model.Input) (order, bool) {
	cmd := order{}
	switch input.Command {
	case enum.Register:
		if input.Func == nil {
			return order{}, false
		}
		cmd.callbackFunc = input.Func
		cmd.command = registerCallBack
	case enum.Start:
		cmd.command = startApi
		break
	case stopApi:
		cmd.command = stopApi
		break
	case reloadApi:
		cmd.command = reloadApi
		break
	default:
		hwlog.RunLog.Error("Invalid command")
		return order{}, false
	}
	return cmd, true
}

// executeOrderByOrderQueue flag value only modified by execute
func executeOrderByOrderQueue(cmd order) {
	curStatus := controllerflags.IsControllerExited.GetState()
	if cmd.command == stopApi && !curStatus {
		controllerflags.IsControllerExited.SetState(true)
		controller.Stop()
	} else if cmd.command == reloadApi || cmd.command == startApi {
		if curStatus {
			controllerflags.IsControllerExited.SetState(false)
			go controller.Start()
		} else {
			controllerflags.IsControllerExited.SetState(true)
			controller.Stop()
			controllerflags.IsControllerExited.SetState(false)
			go controller.Start()
		}
	} else if cmd.command == registerCallBack {
		go controller.RegisterDetectionCallback(cmd.callbackFunc)
	}
}

// Execute for uniform interface
func Execute(input *model.Input) int {
	callExecuteSyncLock.Lock()
	defer callExecuteSyncLock.Unlock()
	hwlog.RunLog.Infof("Commond input: %+v", input)
	if !checkInputInvalid(input) {
		return errCode
	}
	cmd, flag := switchCommand(input)
	if !flag {
		return errCode
	}
	executeOrderByOrderQueue(cmd)
	return successCode
}
