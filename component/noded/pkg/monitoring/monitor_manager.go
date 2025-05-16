/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package monitoring for monitoring the fault on the server
package monitoring

import (
	"context"
	"time"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/monitoring/ipmimonitor"
)

// PluginMonitor monitor plugin interface
type PluginMonitor interface {
	Monitoring()
	Init() error
	Stop()
	Name() string
}

// MonitorManager manage monitors
type MonitorManager struct {
	monitors           []PluginMonitor
	client             *kubeclient.ClientK8s
	faultManager       common.FaultManager
	nextFaultProcessor common.FaultProcessor
	stopChan           chan struct{}
}

// NewMonitorManager create a monitor manager
func NewMonitorManager(client *kubeclient.ClientK8s) *MonitorManager {
	return &MonitorManager{
		client:       client,
		faultManager: common.NewFaultManager(),
		stopChan:     make(chan struct{}, 1),
	}
}

// Init register monitor plugin and start them
func (m *MonitorManager) Init() error {
	const retryTime = 3
	m.monitors = append(m.monitors, ipmimonitor.NewIpmiEventMonitor(m.faultManager))
	for _, monitor := range m.monitors {
		var initSuc bool
		for i := 0; i < retryTime; i++ {
			if err := monitor.Init(); err != nil {
				hwlog.RunLog.Errorf("init monitor[%s] failed, error: %v, retry count: %d", monitor.Name(), err, i+1)
				time.Sleep(time.Second)
				continue
			}
			initSuc = true
			break
		}
		if initSuc {
			hwlog.RunLog.Infof("init monitor[%s] success", monitor.Name())
			go monitor.Monitoring()
		} else {
			hwlog.RunLog.Errorf("init monitor[%s] failed, the maximum number of retries (%d) has been reached",
				monitor.Name(), retryTime)
		}
	}
	return nil
}

// Run working loop
func (m *MonitorManager) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(common.ParamOption.ReportInterval) * time.Second)
	defer ticker.Stop()
	triggerTicker := time.NewTicker(time.Second)
	defer triggerTicker.Stop()
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Error("stop channel is closed")
				return
			}
			hwlog.RunLog.Info("receive stop signal, monitor manager shut down...")
			m.Stop()
			return
		case <-triggerTicker.C:
			m.parseTriggers()
		case <-ticker.C:
			m.Execute(m.faultManager.GetFaultDevInfo())
		}
	}
}

func (m *MonitorManager) parseTriggers() {
	select {
	case <-common.GetUpdateChan():
		hwlog.RunLog.Info("receive update trigger, processing fault report")
		m.Execute(m.faultManager.GetFaultDevInfo())
	default:
		hwlog.RunLog.Debug("No update trigger, skipping execute")
	}
}

// Stop terminate working loop
func (m *MonitorManager) Stop() {
	for _, monitor := range m.monitors {
		monitor.Stop()
	}
}

// Execute update node status and send message to next fault processor
func (m *MonitorManager) Execute(faultDevInfo *common.FaultDevInfo) {
	m.nextFaultProcessor.Execute(faultDevInfo)
}

// SetNextFaultProcessor set the next fault processor
func (m *MonitorManager) SetNextFaultProcessor(processor common.FaultProcessor) {
	m.nextFaultProcessor = processor
}
