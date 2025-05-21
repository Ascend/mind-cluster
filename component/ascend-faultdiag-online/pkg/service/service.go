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
Package service 提供服务相关的功能，包括启动循环服务、处理请求等。
*/
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ascend-faultdiag-online/pkg/context"
	"ascend-faultdiag-online/pkg/model/servicemodel"
	"ascend-faultdiag-online/pkg/service/request"
)

const (
	RequestTimeOut = 5 * time.Second
)

// startLoopService 启动循环服务
func startLoopService(ctx *context.FaultDiagContext) {
	semaphore := make(chan struct{}, 1000)
	for {
		select {
		case reqCtx, ok := <-ctx.ReqQue:
			if !ok {
				ctx.Logger.Println("request context chan closed")
				return
			}
			semaphore <- struct{}{}
			go func() {
				defer func() {
					close(reqCtx.FinishChan)
					<-semaphore
				}()
				if apiFunc, err := ctx.Router.HandleApi(reqCtx.Api); err != nil {
					reqCtx.Response = servicemodel.ErrorResponse(err.Error())
				} else {
					err = apiFunc(ctx.GetCtxData(), ctx.DiagContext, reqCtx)
					if err != nil {
						reqCtx.Response = servicemodel.ErrorResponse(err.Error())
					}
				}
			}()
		case <-ctx.StopChan:
			ctx.Logger.Println("Loop service stopped")
			return
		}
	}
}

// StartFaultDiagService 开启循环服务和指标诊断服务
func StartFaultDiagService(ctx *context.FaultDiagContext) {
	ctx.IsRunning = true
	go startLoopService(ctx)
	go ctx.DiagContext.StartDiag(ctx.GetCtxData())
}

// StopFaultDiagService 停止循环服务
func StopFaultDiagService(ctx *context.FaultDiagContext) {
	close(ctx.StopChan)
	close(ctx.ReqQue)
	ctx.IsRunning = false
}

// HandleRequest 处理请求
func HandleRequest(ctx *context.FaultDiagContext, api string, reqJson string) (string, error) {
	if !ctx.IsRunning {
		return "", fmt.Errorf("service is not running")
	}
	reqCtx := request.NewRequestContext(api, reqJson)
	// 等待添加进队列
	select {
	case ctx.ReqQue <- reqCtx:
	case <-time.After(RequestTimeOut):
		return "{}", errors.New("the request queue is full")
	}
	// 阻塞等待响应完成
	select {
	case <-reqCtx.FinishChan:
		jsonStr, err := json.Marshal(reqCtx.Response)
		if err != nil {
			return "", err
		}
		return string(jsonStr), err
	}
}
