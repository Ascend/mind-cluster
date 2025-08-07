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

// Package api provides some test cases for the packet servicecore
package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/core/context/contextdata"
	"ascend-faultdiag-online/pkg/core/context/diagcontext"
	"ascend-faultdiag-online/pkg/core/model"
	"ascend-faultdiag-online/pkg/core/model/diagmodel"
	"ascend-faultdiag-online/pkg/utils/constants"
)

func TestNewApi(t *testing.T) {
	subApis := []*Api{
		{Name: "child1"},
		{Name: "child2"},
	}

	parent := NewApi("parent", nil, subApis)

	const ExpectedSubApiCount = 2

	if len(parent.SubApiMap) != ExpectedSubApiCount {
		assert.FailNow(t, "Expected 2 sub APIs, got %d", len(parent.SubApiMap))
	}

	for _, child := range subApis {
		assert.Equal(t, parent, child.ParentApi)
	}
}

func TestGetFullApiStr(t *testing.T) {
	root := NewApi("root", nil, []*Api{
		NewApi("v1", nil, []*Api{
			NewApi("users", nil, nil),
		}),
	})

	testCases := []struct {
		name     string
		api      *Api
		expected string
	}{
		{
			name:     "single level",
			api:      NewApi("test", nil, nil),
			expected: "test",
		},
		{
			name:     "three levels",
			api:      root.SubApiMap["v1"].SubApiMap["users"],
			expected: strings.Join([]string{"root", "v1", "users"}, constants.ApiSeparator),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.api.GetFullApiStr()
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestBuildApiFunc(t *testing.T) {
	var mockFunc = func(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
		reqCtx *model.RequestContext, model *diagmodel.DiagModel) error {
		return nil
	}
	// param is nil
	apiFunc, err := BuildApiFunc(nil)
	assert.Equal(t, "invalid param: reqModel or targetFunc is nil", err.Error())
	assert.Nil(t, apiFunc)

	// ReqModel and TargetFunc are nil
	var param = &ApiFuncBuildParam{}
	apiFunc, err = BuildApiFunc(param)
	assert.Equal(t, "invalid param: reqModel or targetFunc is nil", err.Error())
	assert.Nil(t, apiFunc)

	// TargetFunc is not a functioon
	param = &ApiFuncBuildParam{ReqModel: struct{}{}, TargetFunc: struct{}{}}
	apiFunc, err = BuildApiFunc(param)
	assert.Equal(t, "param targetFunc is not a function", err.Error())
	assert.Nil(t, apiFunc)

	// TargetFunc with lack of arguments
	param = &ApiFuncBuildParam{ReqModel: struct{}{}, TargetFunc: func() {}}
	apiFunc, err = BuildApiFunc(param)
	assert.Equal(t, "the target function has insufficient parameters", err.Error())
	assert.Nil(t, apiFunc)

	// ReqModel is not matched
	param = &ApiFuncBuildParam{ReqModel: struct{}{}, TargetFunc: mockFunc}
	apiFunc, err = BuildApiFunc(param)
	assert.Equal(t, "the type of the reqModel argument does not match", err.Error())
	assert.Nil(t, apiFunc)
}
