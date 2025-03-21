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
Package diagcontext some test case for the diag ticker.
*/
package diagcontext

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/context/contextdata"
)

var (
	diagItem = &DiagItem{
		Name:           "npu check",
		Interval:       10,
		Rules:          []*DiagRule{},
		CustomRules:    []*CustomRule{},
		ConditionGroup: &ConditionGroup{},
		Description:    "npu chip diagnosis",
	}
	diagTicker = NewDiagTicker(diagItem)
)

func TestNewDiagTicker(t *testing.T) {
	assert.NotNil(t, diagTicker)
}

func TestDiagTicker_Close(t *testing.T) {
	select {
	case <-diagTicker.StopChan:
		assert.True(t, false, "没有对无缓冲管道写入数据，阻塞监测管道读取数据")
	default:
		assert.True(t, true, "没有对无缓冲管道写入数据，所有case不可用，执行default分支")
	}

	diagTicker.Close()
	select {
	case _, ok := <-diagTicker.StopChan:
		assert.False(t, ok, "管道已关闭，读取操作不会阻塞，会立刻返回类型的零值")
	default:
		assert.True(t, false, "管道已关闭，读取操作不会阻塞。")
	}
}

func TestDiagTicker_Start(t *testing.T) {
	ctxData := &contextdata.CtxData{
		Environment: contextdata.NewEnvironment(),
		Framework:   &contextdata.Framework{StopChan: make(chan struct{})},
	}

	assert.False(t, diagTicker.running, "开始诊断前，诊断上下文中诊断计时器未运行")
	diagTicker.Start(ctxData, NewDiagContext())
	assert.True(t, diagTicker.running, "开始诊断后，诊断上下文中诊断计时器状态为运行")
	// todo需要等待进入协程中测试
}
