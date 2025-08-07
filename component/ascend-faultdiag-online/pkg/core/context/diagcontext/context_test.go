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
Package diagcontext some test case for the diag context.
*/
package diagcontext

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/core/context/contextdata"
)

var (
	ctx       = NewDiagContext()
	diagItems = []*DiagItem{
		{
			Name:           "npu check",
			Interval:       10,
			Rules:          []*DiagRule{},
			CustomRules:    []*CustomRule{},
			ConditionGroup: &ConditionGroup{},
			Description:    "npu chip diagnosis",
		},
		{
			Name:           "env check",
			Interval:       20,
			Rules:          []*DiagRule{},
			CustomRules:    []*CustomRule{},
			ConditionGroup: &ConditionGroup{},
			Description:    "environment diagnosis",
		},
	}
	tickers = ctx.UpdateDiagItems(diagItems)
)

func TestNewDiagContext(t *testing.T) {
	assert.NotNil(t, ctx)
}

func TestUpdateDiagItems(t *testing.T) {
	assert.Equal(t, len(tickers), len(diagItems), "诊断计时器数量应与诊断项数量相同")

	for _, item := range diagItems {
		assert.Equal(t, item, ctx.DiagItemMap[item.Name], "现有列表中诊断项应与添加的诊断项相同")
		assert.Equal(t, item, ctx.tickerMap[item.Name].DiagItem, "现有列表中诊断计时器应与添加的诊断计时器相同")
	}
}

func TestStartDiag(t *testing.T) {
	for _, ticker := range ctx.tickerMap {
		assert.False(t, ticker.running, "开始诊断前，诊断上下文中诊断计时器未运行")
	}
	ctx.StartDiag(&contextdata.CtxData{Framework: &contextdata.Framework{StopChan: make(chan struct{})}})
	for _, ticker := range ctx.tickerMap {
		assert.True(t, ticker.running, "开始诊断后，诊断上下文中诊断计时器状态为运行")
	}
	var sleepTime time.Duration = 2
	time.Sleep(sleepTime * time.Second)
}

func TestCloseDiagItem(t *testing.T) {
	// 关闭一个存在的诊断项
	exitItemName := "npu check"
	for _, item := range diagItems {
		if item.Name == exitItemName {
			ctx.CloseDiagItem(exitItemName)
		}
		_, exit := ctx.tickerMap[item.Name]
		if item.Name == exitItemName {
			assert.False(t, exit, "关闭诊断项后，现有诊断计时器map中没有该诊断项")
		} else {
			assert.True(t, exit, "未关闭的诊断项任然在现有诊断计时器map中")
		}
	}

	// 关闭一个不存在的诊断项，直接返回
	itemName := "name not exit"
	_, preNotExit := ctx.tickerMap[itemName]
	assert.False(t, preNotExit)
	ctx.CloseDiagItem(itemName)
	_, sufNotExit := ctx.tickerMap[itemName]
	assert.False(t, sufNotExit)
}
