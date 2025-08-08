/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package register provides a register func for registering functions
package register

import (
	"ascend-faultdiag-online/pkg/algo_src/netfault"
	"ascend-faultdiag-online/pkg/algo_src/slownode"
	"ascend-faultdiag-online/pkg/core/context"
	"ascend-faultdiag-online/pkg/core/funchandler"
	"ascend-faultdiag-online/pkg/core/model/enum"
)

// functionMap registers the function handlers in the FaultDiagContext.
func functionMap(fdCtx *context.FaultDiagContext) {
	handlerMap := map[string]*funchandler.Handler{
		enum.SlowNode: {
			FuncType:    slownode.GetType(),
			FuncVersion: slownode.GetVersion(),
			ExecuteFunc: funchandler.GenerateExecuteFunc(slownode.Execute, enum.SlowNode),
		},
		enum.NetFault: {
			FuncType:    netfault.GetType(),
			FuncVersion: netfault.GetVersion(),
			ExecuteFunc: funchandler.GenerateExecuteFunc(netfault.Execute, enum.NetFault),
		},
	}
	fdCtx.RegisterFunc(handlerMap)
}
