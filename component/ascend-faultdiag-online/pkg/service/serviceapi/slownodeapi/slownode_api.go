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
Package slownodeapi provides API
*/
package slownodeapi

import (
	"fmt"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/api"
	"ascend-faultdiag-online/pkg/core/context/contextdata"
	"ascend-faultdiag-online/pkg/core/context/diagcontext"
	"ascend-faultdiag-online/pkg/core/model"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/cluster"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/node"
)

var dContext *diagcontext.DiagContext = nil
var contextData *contextdata.CtxData = nil

// ClusterStart 开启Cluster
func ClusterStart() *api.Api {
	return api.BuildApi(string(enum.Start), &slownode.ReqInput{}, ClusterStartFunc, nil)
}

// ClusterStop 关闭Cluster
func ClusterStop() *api.Api {
	return api.BuildApi(string(enum.Stop), &slownode.ReqInput{}, ClusterStopFunc, nil)
}

// ClusterReload reload Cluster
func ClusterReload() *api.Api {
	return api.BuildApi(string(enum.Reload), &slownode.ReqInput{}, ClusterReloadFunc, nil)
}

// NodeStart 开启Node
func NodeStart() *api.Api {
	return api.BuildApi(string(enum.Start), &slownode.ReqInput{}, NodeStartFunc, nil)
}

// NodeStop 关闭Node
func NodeStop() *api.Api {
	return api.BuildApi(string(enum.Stop), &slownode.ReqInput{}, NodeStopFunc, nil)
}

// NodeReload reload Node
func NodeReload() *api.Api {
	return api.BuildApi(string(enum.Reload), &slownode.ReqInput{}, NodeReloadFunc, nil)
}

func startFunc(
	ctxData *contextdata.CtxData,
	diagCtx *diagcontext.DiagContext,
	reqCtx *model.RequestContext,
	inputModel *slownode.ReqInput,
	target enum.DeployMode,
) error {
	dContext = diagCtx
	contextData = ctxData
	handler := ctxData.Framework.FuncHandler[enum.SlowNode]
	if handler == nil {
		return fmt.Errorf("start failed: no handler for %s", enum.SlowNode)
	}

	eventType := inputModel.EventType
	callbackFunc := algoResultCallBack

	if eventType == enum.DataParse {
		if target == enum.Cluster {
			callbackFunc = mergeParalleGroupInfoResultCallback
		} else {
			callbackFunc = dataParseResultCallback
		}
	}

	input := createInput(enum.Register, callbackFunc, target, inputModel)
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]register %s req data: %+v", inputModel.EventType, input)

	_, err := handler.ExecuteFunc(input)
	if err != nil {
		return err
	}
	input = createInput(enum.Start, nil, target, inputModel)
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]start %s req data: %+v", inputModel.EventType, input)
	_, err = handler.ExecuteFunc(input)
	if err != nil {
		return err
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = fmt.Sprintf("start %s successfully", inputModel.EventType)
	return nil
}

func stopFunc(
	ctxData *contextdata.CtxData,
	reqCtx *model.RequestContext,
	inputModel *slownode.ReqInput,
	target enum.DeployMode,
) error {
	handler := ctxData.Framework.FuncHandler[enum.SlowNode]
	if handler == nil {
		return fmt.Errorf("start failed: no handler for %s", enum.SlowNode)
	}

	input := createInput(enum.Stop, nil, target, inputModel)
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]stop %s req data: %+v", inputModel.EventType, input)
	_, err := handler.ExecuteFunc(input)
	if err != nil {
		return err
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = fmt.Sprintf("stop %s successfully", inputModel.EventType)
	return nil
}

func reloadFunc(
	ctxData *contextdata.CtxData,
	diagCtx *diagcontext.DiagContext,
	reqCtx *model.RequestContext,
	inputModel *slownode.ReqInput,
	target enum.DeployMode,
) error {
	dContext = diagCtx
	contextData = ctxData
	handler := ctxData.Framework.FuncHandler[enum.SlowNode]
	if handler == nil {
		return fmt.Errorf("start failed: no handler for %s", enum.SlowNode)
	}
	input := createInput(enum.Reload, nil, target, inputModel)

	hwlog.RunLog.Infof("[FD-OL SLOWNODE]reload %s req data: %+v", inputModel.EventType, input)
	_, err := handler.ExecuteFunc(input)
	if err != nil {
		return err
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = fmt.Sprintf("reload %s successfully", inputModel.EventType)
	return nil
}

// ClusterStartFunc /start api对应方法，启动算法
func ClusterStartFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *model.RequestContext, inputModel *slownode.ReqInput) error {
	return startFunc(ctxData, diagCtx, reqCtx, inputModel, enum.Cluster)
}

// ClusterStopFunc /stop api对应方法，停止算法
func ClusterStopFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *model.RequestContext, inputModel *slownode.ReqInput) error {
	return stopFunc(ctxData, reqCtx, inputModel, enum.Cluster)
}

// ClusterReloadFunc /reload api对应方法，reload算法
func ClusterReloadFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *model.RequestContext, inputModel *slownode.ReqInput) error {
	return reloadFunc(ctxData, diagCtx, reqCtx, inputModel, enum.Cluster)
}

// NodeStartFunc /start api对应方法，启动算法
func NodeStartFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *model.RequestContext, inputModel *slownode.ReqInput) error {
	return startFunc(ctxData, diagCtx, reqCtx, inputModel, enum.Node)
}

// NodeStopFunc /stop api对应方法，停止算法
func NodeStopFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *model.RequestContext, inputModel *slownode.ReqInput) error {
	return stopFunc(ctxData, reqCtx, inputModel, enum.Node)
}

// NodeReloadFunc /reload api对应方法，reload算法
func NodeReloadFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *model.RequestContext, inputModel *slownode.ReqInput) error {
	return reloadFunc(ctxData, diagCtx, reqCtx, inputModel, enum.Node)
}

// algoResultCallBack process the callback data of algo result
func algoResultCallBack(message string) {
	if dContext == nil || contextData == nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]process slow node algo callback failed: context is nil")
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]%v: got the slow node algo callback data: %v",
		contextData.Framework.Config.Mode, message)

	if contextData.Framework.Config.Mode == enum.Cluster {
		cluster.AlgoCallbackProcessor(message)
	} else {
		node.AlgoCallbackProcessor(message)
	}
}

// dataParseResultCallback the callback func for slow node data parse.
func dataParseResultCallback(message string) {
	if dContext == nil || contextData == nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]process data parse result callback data failed: context is nil")
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]%v: got the data parse callback data: %s",
		contextData.Framework.Config.Mode, message)
	// only fd-on running in the node will trigger this callback
	// we don't need to determine the target is cluster or node
	node.DataParseCallbackProcessor(message)
}

// mergeParalleGroupInfoResultCallback the callback func for slow node merge parallel group info.
func mergeParalleGroupInfoResultCallback(message string) {
	if dContext == nil || contextData == nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]process merge parallel group info callback data failed: context is nil")
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]%v: got the merge parallel group info callback data: %s",
		contextData.Framework.Config.Mode, message)
	cluster.ParallelGroupInfoCallbackProcessor(message)
}

// createInput create a model.Input as input data of a param of execute
func createInput(
	command enum.Command,
	f model.CallbackFunc,
	target enum.DeployMode,
	inputModel *slownode.ReqInput,
) model.Input {
	input := model.Input{
		Command:   command,
		Target:    target,
		Func:      f,
		EventType: enum.SlowNodeAlgo,
	}
	if inputModel == nil {
		return input
	}
	input.EventType = inputModel.EventType
	input.Model = inputModel.AlgoInput
	if input.EventType == enum.DataParse {
		input.Model = inputModel.DataParseInput
	}
	return input
}
