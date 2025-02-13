package diagcontext

import (
	"time"

	"ascend-faultdiag-online/pkg/context"
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
func (diagTicker *DiagTicker) Start(fdCtx *context.FaultDiagContext) {
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
				fdCtx.DiagCtx.DiagRecordStore.UpdateRecord(diagTicker.DiagItem, diagTicker.DiagItem.Diag(fdCtx))
			case _, ok := <-fdCtx.StopChan:
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
