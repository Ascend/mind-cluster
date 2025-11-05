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
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/api"
	"ascend-faultdiag-online/pkg/core/config"
	"ascend-faultdiag-online/pkg/core/context/contextdata"
	"ascend-faultdiag-online/pkg/core/context/diagcontext"
	"ascend-faultdiag-online/pkg/core/funchandler"
	"ascend-faultdiag-online/pkg/core/model"
	"ascend-faultdiag-online/pkg/core/route"
)

const (
	// RequestTimeOut is a constant of request timeout
	RequestTimeOut = 5 * time.Second
)

var (
	// FdCtx is a global variabl of fault diag context
	FdCtx     *FaultDiagContext
	fdCtxOnce sync.Once
)

// FaultDiagContext represents the global context for the plugin.
type FaultDiagContext struct {
	contextdata.Framework                            // 架构信息集合
	contextdata.Environment                          // 环境信息集合
	DiagContext             *diagcontext.DiagContext // 诊断上下文
	Router                  *route.Router            // 请求路由
}

// NewFaultDiagContext creates a new instance of FaultDiagContext.
func NewFaultDiagContext(configPath string) (*FaultDiagContext, error) {
	var err error
	var cg *config.FaultDiagConfig
	fdCtxOnce.Do(func() {
		cg, err = config.LoadConfig(configPath)
		if err != nil {
			return
		}
		logger := log.New(os.Stdout, "[FaultDiag Online] ", log.LstdFlags)
		FdCtx = &FaultDiagContext{
			Framework: contextdata.Framework{Config: cg,
				ReqQue:   make(chan *model.RequestContext, cg.QueueSize),
				StopChan: make(chan struct{}),
				Logger:   logger,
			},
			Environment: *contextdata.NewEnvironment(),
			Router:      route.NewRouter(),
			DiagContext: diagcontext.NewDiagContext(),
		}
		FdCtx.loadDiagItems()
	})
	return FdCtx, err
}

// RegisterFunc registers the execute function in the FaultDiagContext.
func (fdCtx *FaultDiagContext) RegisterFunc(funcHandlerMap map[string]*funchandler.Handler) {
	if fdCtx == nil {
		return
	}
	// register all the api in the fd context
	fdCtx.FuncHandler = funcHandlerMap
}

// loadDiagItems 加载诊断项
func (fdCtx *FaultDiagContext) loadDiagItems() {
	if fdCtx == nil || fdCtx.DiagContext == nil {
		return
	}
	var diagItems []*diagcontext.DiagItem
	fdCtx.DiagContext.UpdateDiagItems(diagItems)
}

// GetCtxData 返回上下文信息
func (fdCtx *FaultDiagContext) getCtxData() *contextdata.CtxData {
	if fdCtx == nil {
		return nil
	}
	return &contextdata.CtxData{
		Environment: &fdCtx.Environment,
		Framework:   &fdCtx.Framework,
	}
}

func (fdCtx *FaultDiagContext) executeFunc(reqCtx *model.RequestContext) {
	if fdCtx == nil || fdCtx.Router == nil || reqCtx == nil {
		return
	}
	done := make(chan struct{}, 1)
	go func() {
		if apiFunc, err := fdCtx.Router.HandleApi(reqCtx.Api); err != nil {
			reqCtx.Response = model.ErrorResponse(err.Error())
		} else {
			err = apiFunc(fdCtx.getCtxData(), fdCtx.DiagContext, reqCtx)
			if err != nil {
				reqCtx.Response = model.ErrorResponse(err.Error())
			}
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(RequestTimeOut):
		reqCtx.Response = model.ErrorResponse("the request is timeout")
	}
}

// startLoopService 启动循环服务
func (fdCtx *FaultDiagContext) startLoopService() {
	if fdCtx == nil || fdCtx.Config == nil {
		hwlog.RunLog.Error("[FDOL]invalid nil fdCtx or fdCtx.Config")
		return
	}
	semaphore := make(chan struct{}, fdCtx.Config.QueueSize)
	for {
		select {
		case reqCtx, ok := <-fdCtx.ReqQue:
			if !ok {
				fdCtx.Logger.Println("request context chan closed")
				return
			}
			semaphore <- struct{}{}
			go func() {
				fdCtx.executeFunc(reqCtx)
				close(reqCtx.FinishChan)
				<-semaphore
			}()
		case <-fdCtx.StopChan:
			fdCtx.Logger.Println("Loop service stopped")
			return
		}
	}
}

// HandleRequest 处理请求
func (fdCtx *FaultDiagContext) handleRequest(api string, reqJson string) (string, error) {
	if fdCtx == nil {
		return "", errors.New("fdCtx is nil")
	}
	if !fdCtx.IsRunning {
		return "", errors.New("service is not running")
	}
	reqCtx := model.NewRequestContext(api, reqJson)

	select {
	// add the req ctx to req queue
	case fdCtx.ReqQue <- reqCtx:
	case <-time.After(RequestTimeOut):
		return "{}", errors.New("the request queue is full")
	}

	// wait util req finished
	<-reqCtx.FinishChan
	jsonStr, err := json.Marshal(reqCtx.Response)
	if err != nil {
		return "", err
	}
	return string(jsonStr), err
}

// RegisterRouter register all the api in the fd context
func (fdCtx *FaultDiagContext) RegisterRouter(featureApi *api.Api) {
	if fdCtx == nil {
		return
	}
	fdCtx.Router.RootApi = featureApi
}

// StartService 开启循环服务和指标诊断服务
func (fdCtx *FaultDiagContext) StartService() {
	if fdCtx == nil || fdCtx.IsRunning {
		return
	}
	fdCtx.IsRunning = true
	go fdCtx.startLoopService()
	go fdCtx.DiagContext.StartDiag(fdCtx.getCtxData())
}

// StopService 停止循环服务
func (fdCtx *FaultDiagContext) StopService() {
	if fdCtx == nil {
		return
	}
	close(fdCtx.StopChan)
	close(fdCtx.ReqQue)
	fdCtx.IsRunning = false
}

// Request 发起请求
func (fdCtx *FaultDiagContext) Request(api string, reqJson string) (string, error) {
	if fdCtx == nil {
		return "", errors.New("fdCtx is nil")
	}
	resp, err := fdCtx.handleRequest(api, reqJson)
	if err != nil {
		return "", err
	}
	return resp, nil
}
