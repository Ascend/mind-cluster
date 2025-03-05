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

// Package core provides the definition and the model of API
package servicecore

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"ascend-faultdiag-online/pkg/context/contextdata"
	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/service/request"
	"ascend-faultdiag-online/pkg/utils/constants"
)

type ApiFunc func(fdCtxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext, reqCtx *request.Context) error

// Api 接口
type Api struct {
	Name      string          // 接口名(可以是中间片段)
	ApiFunc   ApiFunc         // 接口函数
	SubApiMap map[string]*Api // 获取下一级api
	ParentApi *Api            // 父节点
}

// GetFullApiStr 获取全量Api
func (api *Api) GetFullApiStr() string {
	curApi := api
	apiParts := make([]string, 0)
	for curApi != nil {
		apiParts = append(apiParts, curApi.Name)
		curApi = curApi.ParentApi
	}
	sort.Slice(apiParts, func(i, j int) bool {
		return i > j
	})
	return strings.Join(apiParts, constants.ApiSeparator)
}

// NewApi 创建实例
func NewApi(name string, apiFunc ApiFunc, subApis []*Api) *Api {
	subApiMap := make(map[string]*Api)
	newApi := &Api{Name: name, ApiFunc: apiFunc, SubApiMap: subApiMap}
	for _, api := range subApis {
		if api == nil {
			continue
		}
		api.ParentApi = newApi
		subApiMap[api.Name] = api
	}
	return newApi
}

// BuildApi 构建api函数
func BuildApi(api string, reqModel interface{}, targetFunc interface{}, subApis []*Api) *Api {
	param := &ApiFuncBuildParam{
		ReqModel:   reqModel,
		TargetFunc: targetFunc,
	}
	apiFunc, err := BuildApiFunc(param)
	if err != nil {
		return nil
	}
	return NewApi(api, apiFunc, subApis)
}

const (
	targetFuncParamSize = 4
	modelArgIdx         = 3
)

// ApiFuncBuildParam 构建参数
type ApiFuncBuildParam struct {
	ReqModel   interface{} // 请求模型
	TargetFunc interface{} // 目标函数, 首个参数必须为ctxData, 第二个参数为 diagCtx，第二个参数为reqCtx，第三个参数为ReqModel, 返回 error
}

// BuildApiFunc 通过反射创建Api处理函数，避免重复执行反序列化
func BuildApiFunc(param *ApiFuncBuildParam) (ApiFunc, error) {
	if param == nil || param.ReqModel == nil || param.TargetFunc == nil {
		return nil, errors.New("invalid param: reqModel or targetFunc is nil")
	}
	funcValue := reflect.ValueOf(param.TargetFunc)
	funcType := funcValue.Type()
	if funcType.Kind() != reflect.Func {
		return nil, errors.New("param targetFunc is not a function")
	}
	if funcType.NumIn() != targetFuncParamSize {
		return nil, errors.New("the target function has insufficient parameters")
	}
	modelArgValue := reflect.ValueOf(param.ReqModel)
	if funcType.In(modelArgIdx) != modelArgValue.Type() {
		return nil, errors.New("the type of the reqModel argument does not match")
	}
	args := make([]reflect.Value, 0, targetFuncParamSize)
	return func(fdCtxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext, reqCtx *request.Context) error {
		err := json.Unmarshal([]byte(reqCtx.ReqJson), param.ReqModel)
		if err != nil {
			return err
		}
		args = append(args, reflect.ValueOf(fdCtxData))
		args = append(args, reflect.ValueOf(reqCtx))
		args = append(args, reflect.ValueOf(diagCtx))
		args = append(args, modelArgValue)
		results := funcValue.Call(args)
		if len(results) == 0 {
			return errors.New("the api function must have return value")
		}
		if results[0].Interface() == nil {
			return nil
		}
		return fmt.Errorf("%v", results[0].Interface())
	}, nil
}
