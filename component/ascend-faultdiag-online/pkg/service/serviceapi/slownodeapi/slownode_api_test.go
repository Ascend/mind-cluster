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

// Package slownodeapi is a DT collection for func in slownode_api
package slownodeapi

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/config"
	"ascend-faultdiag-online/pkg/core/context/contextdata"
	"ascend-faultdiag-online/pkg/core/context/diagcontext"
	"ascend-faultdiag-online/pkg/core/funchandler"
	"ascend-faultdiag-online/pkg/core/model"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/cluster"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/node"
)

const logLineLength = 256

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout:  true,
		MaxLineLength: logLineLength,
	}
	err := hwlog.InitRunLogger(&config, context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}

// MockHandler 模拟 SoHandler 的行为
type MockHandler struct {
	mock.Mock
}

// ExecuteFunc 模拟 SoHandler 的 ExecuteFunc 方法
func (m *MockHandler) ExecuteFunc(input model.Input) (int, error) {
	args := m.Called(input)
	return args.Int(0), args.Error(1)
}

func createMockHandlerFunc(mockHandler *MockHandler) *funchandler.Handler {
	return &funchandler.Handler{
		ExecuteFunc: func(input model.Input) (int, error) {
			return mockHandler.ExecuteFunc(input)
		},
	}
}

func createMockCtxData(mockHandler *MockHandler) *contextdata.CtxData {
	return &contextdata.CtxData{
		Framework: &contextdata.Framework{
			Config:      nil, // 如果需要，可以提供具体的配置
			FuncHandler: map[string]*funchandler.Handler{enum.SlowNode: createMockHandlerFunc(mockHandler)},
			ReqQue:      make(chan *model.RequestContext, 10),
			IsRunning:   true,
			StopChan:    make(chan struct{}),
			Logger:      log.New(os.Stdout, "TEST: ", log.LstdFlags),
		},
	}
}

func TestClusterStartFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	diagCtx := &diagcontext.DiagContext{}
	reqCtx := &model.RequestContext{
		Response: &model.ResponseBody{},
	}
	inputModel := &slownode.ReqInput{EventType: enum.DataParse}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 第一次调用 registerCallback
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 第二次调用 start

	// 调用方法
	err := ClusterStartFunc(ctxData, diagCtx, reqCtx, inputModel)

	// 验证结果
	assert.NoError(t, err, "ClusterStartFunc 成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "start dataParse successfully", reqCtx.Response.Msg, "Response 消息正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestClusterStartFuncError(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	// 创建其他必要的参数
	diagCtx := &diagcontext.DiagContext{}
	reqCtx := &model.RequestContext{
		Response: &model.ResponseBody{},
	}
	inputModel := &slownode.ReqInput{}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, errors.New("mock error")).Once() // registerCallback 出错

	// 调用方法
	err := ClusterStartFunc(ctxData, diagCtx, reqCtx, inputModel)

	// 验证结果
	assert.Error(t, err, "ClusterStartFunc 返回错误")
	assert.Equal(t, "mock error", err.Error(), "错误消息正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestClusterStopFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	// 创建其他必要的参数
	reqCtx := &model.RequestContext{
		Response: &model.ResponseBody{},
	}
	diagCtx := &diagcontext.DiagContext{}
	inputModel := &slownode.ReqInput{EventType: enum.SlowNodeAlgo}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 stop

	// 调用方法
	err := ClusterStopFunc(ctxData, diagCtx, reqCtx, inputModel)

	// 验证结果
	assert.NoError(t, err, "ClusterStopFunc 成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "stop slowNodeAlgo successfully", reqCtx.Response.Msg, "Response 消息正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestClusterReloadFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &model.RequestContext{
		Response: &model.ResponseBody{},
	}
	inputModel := &slownode.ReqInput{EventType: enum.SlowNodeAlgo}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 reload

	// 调用方法
	err := ClusterReloadFunc(ctxData, &diagcontext.DiagContext{}, reqCtx, inputModel)

	// 验证结果
	assert.NoError(t, err, "ClusterReloadFunc 应该成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "reload slowNodeAlgo successfully", reqCtx.Response.Msg, "Response 消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestClusterReloadFuncError(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &model.RequestContext{
		Response: &model.ResponseBody{},
	}
	inputModel := &slownode.ReqInput{}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, errors.New("mock error")).Once() // reload 出错

	// 调用方法
	err := ClusterReloadFunc(ctxData, &diagcontext.DiagContext{}, reqCtx, inputModel)

	// 验证结果
	assert.Error(t, err, "ClusterReloadFunc 应该返回错误")
	assert.Equal(t, "mock error", err.Error(), "错误消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestNodeStartFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &model.RequestContext{
		Response: &model.ResponseBody{},
	}
	inputModel := &slownode.ReqInput{EventType: enum.DataParse}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 registerCallBack
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 start

	// 调用方法
	err := NodeStartFunc(ctxData, &diagcontext.DiagContext{}, reqCtx, inputModel)

	// 验证结果
	assert.NoError(t, err, "NodeStartFunc 应该成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "start dataParse successfully", reqCtx.Response.Msg, "Response 消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestNodeStartFuncError(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &model.RequestContext{
		Response: &model.ResponseBody{},
	}
	inputModel := &slownode.ReqInput{}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, errors.New("mock error")).Once() // start 出错

	// 调用方法
	err := NodeStartFunc(ctxData, &diagcontext.DiagContext{}, reqCtx, inputModel)

	// 验证结果
	assert.Error(t, err, "NodeStartFunc 应该返回错误")
	assert.Equal(t, "mock error", err.Error(), "错误消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestNodeStopFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &model.RequestContext{
		Response: &model.ResponseBody{},
	}
	inputModel := &slownode.ReqInput{EventType: enum.DataParse}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 stop

	// 调用方法
	err := NodeStopFunc(ctxData, nil, reqCtx, inputModel)

	// 验证结果
	assert.NoError(t, err, "NodeStopFunc 应该成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "stop dataParse successfully", reqCtx.Response.Msg, "Response 消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestNodeReloadFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &model.RequestContext{
		Response: &model.ResponseBody{},
	}
	inputModel := &slownode.ReqInput{EventType: enum.DataParse}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 reload

	// 调用方法
	err := NodeReloadFunc(ctxData, &diagcontext.DiagContext{}, reqCtx, inputModel)

	// 验证结果
	assert.NoError(t, err, "NodeReloadFunc 应该成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "reload dataParse successfully", reqCtx.Response.Msg, "Response 消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestAlgoResultCallBack(t *testing.T) {
	convey.Convey("test algoResultCallBack", t, func() {
		patch := gomonkey.ApplyFunc(node.AlgoCallbackProcessor, func(string) {
			fmt.Printf("Mocked node AlgoCallbackProcessor called\n")
		})
		patch.ApplyFunc(cluster.AlgoCallbackProcessor, func(string) {
			fmt.Printf("Mocked cluster AlgoCallbackProcessor called\n")
		})
		defer patch.Reset()

		// ctx is nil
		dContext = nil
		contextData = nil
		algoResultCallBack("")
		// partial contextData is nil
		contextData = &contextdata.CtxData{
			Framework: &contextdata.Framework{
				Config: &config.FaultDiagConfig{
					Mode: enum.Cluster,
				},
			},
		}
		algoResultCallBack("")
		// normal case -> cluster
		dContext = &diagcontext.DiagContext{}
		algoResultCallBack("")
		// normal case -> node
		contextData.Framework.Config.Mode = enum.Node
		algoResultCallBack("")
	})
}

func TestDataParseResultCallback(t *testing.T) {
	convey.Convey("test dataParseResultCallback", t, func() {
		patch := gomonkey.ApplyFunc(node.DataParseCallbackProcessor, func(string) {
			fmt.Printf("Mocked DataParseCallbackProcessor called\n")
		})
		defer patch.Reset()

		// ctx is nil
		dContext = nil
		contextData = nil
		dataParseResultCallback("")
		// partial contextData is nil
		contextData = &contextdata.CtxData{
			Framework: &contextdata.Framework{
				Config: &config.FaultDiagConfig{
					Mode: enum.Cluster,
				},
			},
		}
		dataParseResultCallback("")
		// normal case
		dContext = &diagcontext.DiagContext{}
		dataParseResultCallback("")
	})
}

func TestMergeParalleGroupInfoResultCallback(t *testing.T) {
	convey.Convey("test mergeParalleGroupInfoResultCallback", t, func() {
		patch := gomonkey.ApplyFunc(cluster.ParallelGroupInfoCallbackProcessor, func(string) {
			fmt.Printf("Mocked ParallelGroupInfoCallbackProcessor called\n")
		})
		defer patch.Reset()

		// ctx is nil
		dContext = nil
		contextData = nil
		mergeParalleGroupInfoResultCallback("")
		// partial contextData is nil
		contextData = &contextdata.CtxData{
			Framework: &contextdata.Framework{
				Config: &config.FaultDiagConfig{
					Mode: enum.Cluster,
				},
			},
		}
		mergeParalleGroupInfoResultCallback("")
		// normal case
		dContext = &diagcontext.DiagContext{}
		mergeParalleGroupInfoResultCallback("")
	})
}

func TestCreateInput(t *testing.T) {
	convey.Convey("test createInput", t, func() {
		convey.Convey("test inpuModel is nil", func() {
			input := createInput(enum.Start, nil, enum.Cluster, nil)
			convey.So(input.Command, convey.ShouldEqual, enum.Start)
			convey.So(input.EventType, convey.ShouldEqual, enum.SlowNodeAlgo)
			convey.So(input.Model, convey.ShouldBeNil)
		})
		convey.Convey("test model is slowNodeAlgo", func() {
			reqInput := &slownode.ReqInput{
				EventType: enum.SlowNodeAlgo,
				AlgoInput: slownode.AlgoInput{FilePath: "test"},
			}
			input := createInput(enum.Start, nil, enum.Cluster, reqInput)
			convey.So(input.Command, convey.ShouldEqual, enum.Start)
			convey.So(input.EventType, convey.ShouldEqual, enum.SlowNodeAlgo)
			convey.So(input.Model, convey.ShouldNotBeNil)
		})
		convey.Convey("test model is dataParse", func() {
			reqInput := &slownode.ReqInput{
				EventType:      enum.DataParse,
				DataParseInput: slownode.DataParseInput{Traffic: 1},
			}
			input := createInput(enum.Start, nil, enum.Cluster, reqInput)
			convey.So(input.Command, convey.ShouldEqual, enum.Start)
			convey.So(input.EventType, convey.ShouldEqual, enum.DataParse)
			convey.So(input.Model, convey.ShouldNotBeNil)
		})
	})
}
