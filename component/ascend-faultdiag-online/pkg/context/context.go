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

/*
Package context is used to manage the global state and resources of the plugin.
*/
package context

import (
	"log"
	"os"

	"ascend-faultdiag-online/pkg/config"
	"ascend-faultdiag-online/pkg/context/contextdata"
	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/context/sohandle"
	"ascend-faultdiag-online/pkg/service/request"
	"ascend-faultdiag-online/pkg/service/route"
)

// FaultDiagContext represents the global context for the plugin.
type FaultDiagContext struct {
	contextdata.Framework                            // 架构信息集合
	contextdata.Environment                          // 环境信息集合
	DiagContext             *diagcontext.DiagContext // 诊断上下文
	Router                  *route.Router            // 请求路由
}

// NewFaultDiagContext creates a new instance of FaultDiagContext.
func NewFaultDiagContext(config *config.FaultDiagConfig) (*FaultDiagContext, error) {
	soHandlerMap, err := sohandle.GenerateSoHandlerMap(config.SoDir)
	if err != nil {
		return nil, err
	}
	logger := log.New(os.Stdout, "[FaultDiag Online] ", log.LstdFlags)
	fdCtx := &FaultDiagContext{
		Framework: contextdata.Framework{Config: config,
			SoHandlerMap: soHandlerMap,
			ReqQue:       make(chan *request.Context, config.QueueSize),
			StopChan:     make(chan struct{}),
			Logger:       logger,
		},
		Environment: *contextdata.NewEnvironment(),
		Router:      route.NewRouter(),
		DiagContext: diagcontext.NewDiagContext(),
	}
	fdCtx.loadDiagItems()
	return fdCtx, nil
}

// loadDiagItems 加载诊断项
func (fdCtx *FaultDiagContext) loadDiagItems() {
	if fdCtx == nil {
		return
	}
	var diagItems []*diagcontext.DiagItem
	fdCtx.DiagContext.UpdateDiagItems(diagItems)
}

// GetCtxData 返回上下文信息
func (fdCtx *FaultDiagContext) GetCtxData() *contextdata.CtxData {
	if fdCtx == nil {
		return nil
	}
	return &contextdata.CtxData{
		Environment: &fdCtx.Environment,
		Framework:   &fdCtx.Framework,
	}
}
