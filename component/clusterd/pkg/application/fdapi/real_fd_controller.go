//go:build fdol

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

// Package fdapi for a series of function handle ping mesh configmap create/update/delete
package fdapi

import (
	"fmt"

	"ascend-common/common-utils/hwlog"
	fdol "ascend-faultdiag-online"
	"ascend-faultdiag-online/pkg/api/v1"
	"ascend-faultdiag-online/pkg/context"
	"ascend-faultdiag-online/pkg/global/globalctx"
)

const (
	controllerUrl = "feature/netfault/controller"
	fdConfigPath  = "/usr/local/fdConfig.yaml"
	startApi      = "start"
	stopApi       = "stop"
	reloadApi     = "reload"
)

func StartFdOL() {
	hwlog.RunLog.Info("start real fd-ol")
	fdol.StartFDOnline(fdConfigPath, []string{"slowNode", "netFault"}, "cluster")
}

// StartController to start controller
func StartController() {
	requestFD(globalctx.Fdctx, startApi)
}

// StopController to stop controller
func StopController() {
	requestFD(globalctx.Fdctx, stopApi)
}

// ReloadController to reload controller
func ReloadController() {
	requestFD(globalctx.Fdctx, reloadApi)
}

func requestFD(fdCtx *context.FaultDiagContext, api string) {
	if fdCtx == nil {
		hwlog.RunLog.Errorf("fdCtx is nil")
		return
	}
	url := fmt.Sprintf("%s/%s", controllerUrl, api)
	resp, err := v1.Request(fdCtx, url, "{}")
	if err != nil {
		hwlog.RunLog.Errorf("stop controller algorithm failed: %v", err)
		return
	}
	hwlog.RunLog.Infof("the response of %s controller is %v", api, resp)
}
