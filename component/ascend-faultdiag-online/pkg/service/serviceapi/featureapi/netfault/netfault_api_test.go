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

// Package netfault for
package netfault

import (
	"errors"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ascend-faultdiag-online/pkg/context/contextdata"
	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/context/sohandle"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/model/feature"
	"ascend-faultdiag-online/pkg/model/servicemodel"
	"ascend-faultdiag-online/pkg/service/request"
)

type MockHandler struct {
	mock.Mock
}

// ExecuteFunc 模拟 SoHandler 的 ExecuteFunc 方法
func (m *MockHandler) ExecuteFunc(input []byte, output []byte) (int, error) {
	args := m.Called(input, output)
	return args.Int(0), args.Error(1)
}

func createMockSoHandler(mockHandler *MockHandler) *sohandle.SoHandler {
	return &sohandle.SoHandler{
		SoHandle: nil,
		SoType:   "mock",
		ExecuteFunc: func(input []byte, output []byte) (int, error) {
			return mockHandler.ExecuteFunc(input, output)
		},
	}
}

func createMockCtxData(mockHandler *MockHandler) *contextdata.CtxData {
	return &contextdata.CtxData{
		Framework: &contextdata.Framework{
			Config:       nil, // 如果需要，可以提供具体的配置
			SoHandlerMap: map[string]*sohandle.SoHandler{"netfault": createMockSoHandler(mockHandler)},
			ReqQue:       make(chan *request.Context, 10),
			IsRunning:    true,
			StopChan:     make(chan struct{}),
			Logger:       log.New(os.Stdout, "TEST: ", log.LstdFlags),
		},
	}
}

// TestControllerStartFunc test for controller call successfully
func TestControllerStartFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	diagCtx := &diagcontext.DiagContext{}
	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 第一次调用 registerCallback
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 第二次调用 start

	// 调用方法
	err := ControllerStartFunc(ctxData, diagCtx, reqCtx, &feature.Status{})

	// 验证结果
	assert.NoError(t, err, "ControllerStartFunc 成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "netfault controller start successfully", reqCtx.Response.Msg, "Response 消息正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

// TestControllerStartFuncError test for controller call error
func TestControllerStartFuncError(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	// 创建其他必要的参数
	diagCtx := &diagcontext.DiagContext{}
	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}

	// 模拟 handler 的行为 registerCallback 出错
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, errors.New("mock error")).Once()

	// 调用方法
	err := ControllerStartFunc(ctxData, diagCtx, reqCtx, &feature.Status{})

	// 验证结果
	assert.Error(t, err, "ControllerStartFunc 返回错误")
	assert.Equal(t, "mock error", err.Error(), "错误消息正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

// TestControllerStopFunc test for controller stop call successfully
func TestControllerStopFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	// 创建其他必要的参数
	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}
	diagCtx := &diagcontext.DiagContext{}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 stop

	// 调用方法
	err := ControllerStopFunc(ctxData, diagCtx, reqCtx, &feature.Status{})

	// 验证结果
	assert.NoError(t, err, "ControllerStopFunc 成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "netfault controller stop successfully", reqCtx.Response.Msg, "Response 消息正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

// TestControllerReloadFunc test for reload call successfully
func TestControllerReloadFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 reload

	// 调用方法
	err := ControllerReloadFunc(ctxData, nil, reqCtx, &feature.Status{})

	// 验证结果
	assert.NoError(t, err, "ControllerReloadFunc 应该成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "netfault controller reload successfully", reqCtx.Response.Msg, "Response 消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

// TestControllerReloadFuncError test for reload call error
func TestControllerReloadFuncError(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, errors.New("mock error")).Once() // reload 出错

	// 调用方法
	err := ControllerReloadFunc(ctxData, nil, reqCtx, &feature.Status{})

	// 验证结果
	assert.Error(t, err, "ControllerReloadFunc 应该返回错误")
	assert.Equal(t, "mock error", err.Error(), "错误消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}
