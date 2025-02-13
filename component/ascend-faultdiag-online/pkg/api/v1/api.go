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
Package v1 provides API endpoints for managing fault diagnosis operations.
*/
package v1

import (
	"ascend-faultdiag-online/pkg/config"
	"ascend-faultdiag-online/pkg/context"
	"ascend-faultdiag-online/pkg/service"
)

// CreateFdCtx 创建故障诊断上下文
func CreateFdCtx(configPath string) (*context.FaultDiagContext, error) {
	fdConfig, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	ctx, err := context.NewFaultDiagContext(fdConfig)
	if err != nil {
		return nil, err
	}
	return ctx, err
}

// StartService 启动服务
func StartService(ctx *context.FaultDiagContext) {
	service.StartFaultDiagService(ctx)
}

// StopService 停止服务
func StopService(ctx *context.FaultDiagContext) {
	service.StopFaultDiagService(ctx)
}

// Request 发起请求
func Request(ctx *context.FaultDiagContext, api string, reqJson string) (string, error) {
	resp, err := service.HandleRequest(ctx, api, reqJson)
	if err != nil {
		return "", err
	}
	return resp, nil
}
