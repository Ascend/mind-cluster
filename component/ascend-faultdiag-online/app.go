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

// Package ascendfaultdiagonline is a app collections
package ascendfaultdiagonline

import (
	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/context"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/register"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode"
	"ascend-faultdiag-online/pkg/utils"
)

var appFunc = map[string]func(enum.DeployMode){
	enum.SlowNode: slownode.StartSlowNode,
	enum.NetFault: nil,
}

// StartFDOnline is the start func for fd online
func StartFDOnline(fdConfigPath string, apps []string, target enum.DeployMode) {
	hwlog.RunLog.Infof("[FD-OL]received start FD-OL request, fdConfigPath is: %s, enabled apps: %s, and target is: %s",
		fdConfigPath, apps, target)
	fdCtx, err := context.NewFaultDiagContext(fdConfigPath)
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL]created fd context failed: %v", err)
		return
	}
	register.Setup(fdCtx)
	fdCtx.StartService()
	utils.WriteStartInfo()
	for _, app := range apps {
		f := appFunc[app]
		if f == nil {
			hwlog.RunLog.Errorf("[FD-OL]feature func is not exist by app name: %s", app)
			continue
		}
		go f(target)
	}
}
