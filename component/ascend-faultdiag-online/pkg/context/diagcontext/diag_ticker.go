package diagcontext

import (
	"time"

	"ascend-faultdiag-online/pkg/context/contextdata"
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
	if diagTicker.running {
		return
	}
	diagTicker.running = true
	interval := diagTicker.DiagItem.Interval
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				diagCtx.DiagRecordStore.UpdateRecord(diagTicker.DiagItem, diagTicker.DiagItem.Diag(ctxData, diagCtx))
			case _, ok := <-ctxData.Framework.StopChan:
				if !ok {
					break
				}
			case _, ok := <-diagTicker.StopChan:
				if !ok {
					break
				}
			}
		}
	}()
}
