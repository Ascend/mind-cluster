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
Package diagcontext some test case for the diag condition.
*/
package diagcontext

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/fdol/context/contextdata"
	"ascend-faultdiag-online/pkg/model/enum"
)

var (
	// 给定数据
	srcAndDestData = &contextdata.CtxData{
		Environment: contextdata.NewEnvironment(),
		Framework:   nil,
	}
	// 定义诊断条件1：匹配输入数据和给定数据相同
	conditionOne = Condition{
		Data: srcAndDestData,
		MatchingFunc: func(ctxData *contextdata.CtxData, data interface{}) bool {
			return ctxData == data
		},
	}
	// 定义诊断条件2：环境中服务器节点的芯片类型相同
	conditionTwo = Condition{
		Data: srcAndDestData,
		MatchingFunc: func(ctxData *contextdata.CtxData, data interface{}) bool {
			return ctxData.Environment.NodeStatus.ChipType == enum.Ascend910A2
		},
	}
	conditionGroup = &ConditionGroup{
		StaticConditions:  []*Condition{&conditionOne, &conditionTwo},
		DynamicConditions: []*Condition{&conditionOne, &conditionTwo},
	}
)

func TestConditionIsMatching(t *testing.T) {
	// 匹配
	assert.True(t, conditionOne.IsMatching(srcAndDestData))
	// 不匹配
	setData := &contextdata.CtxData{
		Environment: nil,
		Framework:   nil,
	}
	assert.False(t, conditionOne.IsMatching(setData))
}

func TestConditionGroupIsStaticMatching(t *testing.T) {
	assert.True(t, conditionGroup.IsStaticMatching(srcAndDestData))
	// 测试len(group.StaticConditions) == 0
	nilGroup := &ConditionGroup{}
	assert.True(t, nilGroup.IsStaticMatching(srcAndDestData))

}

func TestConditionGroupIsDynamicMatching(t *testing.T) {
	assert.True(t, conditionGroup.IsDynamicMatching(srcAndDestData))
}
