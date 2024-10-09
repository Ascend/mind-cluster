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

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/monitoring/ipmimonitor"
)

// PluginMonitor monitor plugin interface
type PluginMonitor interface {
	Monitoring()
	Init() error
	Stop()
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
	m.monitors = append(m.monitors, ipmimonitor.NewIpmiEventMonitor(m.faultManager))
	for _, monitor := range m.monitors {
		if err := monitor.Init(); err != nil {
			hwlog.RunLog.Errorf("init monitor failed, err is %v", err)
			continue
		}
		go monitor.Monitoring()
	}
	return nil
}

// Run working loop
func (m *MonitorManager) Run(ctx context.Context) {
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
		default:
			m.Execute(m.faultManager.GetFaultDevInfo())
			hwlog.RunLog.Infof("send heartbeat: %d, heartbeat interval: %d",
				m.faultManager.GetHeartbeatTime(), m.faultManager.GetHeartbeatInterval())
			time.Sleep(time.Duration(common.ParamOption.HeartbeatInterval) * time.Second)
		}
	}
}

// Stop terminate working loop
func (m *MonitorManager) Stop() {
	for _, monitor := range m.monitors {
		monitor.Stop()
	}
}

// Execute update node heartbeat and send message to next fault processor
func (m *MonitorManager) Execute(faultDevInfo *common.FaultDevInfo) {
	m.faultManager.SetHeartbeatTime(time.Now().Unix())
	m.faultManager.SetHeartbeatInterval(common.ParamOption.HeartbeatInterval)
	m.nextFaultProcessor.Execute(faultDevInfo)
}

// SetNextFaultProcessor set the next fault processor
func (m *MonitorManager) SetNextFaultProcessor(processor common.FaultProcessor) {
	m.nextFaultProcessor = processor
}
