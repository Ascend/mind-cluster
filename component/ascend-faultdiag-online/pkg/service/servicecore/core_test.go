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

// Package core provides some test cases for the packet servicecore
package servicecore

import (
	"ascend-faultdiag-online/pkg/context/contextdata"
	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/model/diagmodel"
	"ascend-faultdiag-online/pkg/service/request"
	"ascend-faultdiag-online/pkg/utils/constants"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewApi(t *testing.T) {
	subApis := []*Api{
		{Name: "child1"},
		{Name: "child2"},
	}

	parent := NewApi("parent", nil, subApis)

	if len(parent.SubApiMap) != 2 {
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

	var mockFunc = func(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext, reqCtx *request.Context, model *diagmodel.DiagModel) error {
		return nil
	}

	testCases := []struct {
		name        string
		param       *ApiFuncBuildParam
		expectFunc  ApiFunc
		expectError error
	}{
		{
			"param is nil",
			nil,
			nil,
			errors.New("invalid param: reqModel or targetFunc is nil"),
		},
		{
			"ReqModel and TargetFunc are nil",
			&ApiFuncBuildParam{ReqModel: nil, TargetFunc: nil},
			nil,
			errors.New("invalid param: reqModel or targetFunc is nil"),
		},
		{
			"TargetFunc is not a functioon",
			&ApiFuncBuildParam{ReqModel: struct{}{}, TargetFunc: struct{}{}},
			nil,
			errors.New("param targetFunc is not a function"),
		},
		{
			"TargetFunc with lack of arguments",
			&ApiFuncBuildParam{ReqModel: struct{}{}, TargetFunc: func() {}},
			nil,
			errors.New("the target function has insufficient parameters"),
		},
		{
			"ReqModel is not matched",
			&ApiFuncBuildParam{ReqModel: struct{}{}, TargetFunc: mockFunc},
			nil,
			errors.New("the type of the reqModel argument does not match"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apiFunc, err := BuildApiFunc(tc.param)
			if err != nil {
				assert.Equal(t, tc.expectError.Error(), err.Error())
				assert.Nil(t, apiFunc)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, apiFunc)
			}
		})
	}
}
