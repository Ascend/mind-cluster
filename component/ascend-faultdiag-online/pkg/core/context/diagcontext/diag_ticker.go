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
Package diagcontext provides the function to process the diag item.
*/
package diagcontext

import (
	"time"

	"ascend-faultdiag-online/pkg/core/context/contextdata"
)

// DiagTicker 诊断计时器
type DiagTicker struct {
	DiagItem *DiagItem     // 诊断项
	StopChan chan struct{} // 停止chan
	running  bool          // 运行状态
}

// NewDiagTicker 构造函数
func NewDiagTicker(diagItem *DiagItem) *DiagTicker {
	return &DiagTicker{
		DiagItem: diagItem,
		StopChan: make(chan struct{}),
		running:  false,
	}
}

// Close 关闭
func (diagTicker *DiagTicker) Close() {
	close(diagTicker.StopChan)
}

// Start 开始诊断任务
func (diagTicker *DiagTicker) Start(ctxData *contextdata.CtxData, diagCtx *DiagContext) {
	if diagTicker == nil {
		return
	}
	if diagTicker.running {
		return
	}
	diagTicker.running = true
	go diagTicker.startTicker(ctxData, diagCtx, diagTicker.DiagItem.Interval)
}

func (diagTicker *DiagTicker) startTicker(ctxData *contextdata.CtxData, diagCtx *DiagContext, interval int) {
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			diagCtx.DiagRecordStore.UpdateRecord(diagTicker.DiagItem, diagTicker.DiagItem.Diag(ctxData, diagCtx))
		case _, ok := <-ctxData.Framework.StopChan:
			if !ok {
				return
			}
		case _, ok := <-diagTicker.StopChan:
			if !ok {
				return
			}
		}
	}
}
