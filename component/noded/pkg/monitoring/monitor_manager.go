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
	"nodeD/pkg/processmanager"
)

// MonitorManager manage monitors
type MonitorManager struct {
	client             *kubeclient.ClientK8s
	nextFaultProcessor common.FaultProcessor
	stopChan           chan struct{}
}

// NewMonitorManager create a monitor manager
func NewMonitorManager(client *kubeclient.ClientK8s) *MonitorManager {
	return &MonitorManager{
		client:   client,
		stopChan: make(chan struct{}, 1),
	}
}

// Run working loop
func (m *MonitorManager) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(common.ParamOption.ReportInterval) * time.Second)
	defer ticker.Stop()
	triggerTicker := time.NewTicker(time.Second)
	defer triggerTicker.Stop()
	lastUpdateTime := time.Now()
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
			lastUpdateTime = time.Now()
			m.parseTriggers()
		case <-ticker.C:
			if time.Since(lastUpdateTime) < time.Duration(common.ParamOption.ReportInterval)*time.Second {
				continue
			}
			lastUpdateTime = time.Now()
			m.ExecuteAll()
		}
	}
}

func (m *MonitorManager) parseTriggers() {
	for {
		select {
		case processType, ok := <-common.GetTrigger():
			if !ok {
				hwlog.RunLog.Error("updateTrigger channel is closed")
				return
			}
			m.Execute(processType)
		default:
			hwlog.RunLog.Debug("No update trigger, skipping execute")
			return
		}
	}
}

// Stop terminate working loop
func (m *MonitorManager) Stop() {
	processTypes := processmanager.GetAllProcessType()
	for _, processType := range processTypes {
		pluginMonitor := processmanager.GetMonitorPlugins(processType)
		if pluginMonitor != nil {
			pluginMonitor.Stop()
		}
	}
}

// ExecuteAll periodically report all monitoring data regardless of whether the data has changed.
func (m *MonitorManager) ExecuteAll() {
	processTypes := processmanager.GetAllLoopProcessType()
	for _, processType := range processTypes {
		m.Execute(processType)
	}
}

// Execute update node status and send message to next fault processor
func (m *MonitorManager) Execute(processType string) {
	pluginMonitor := processmanager.GetMonitorPlugins(processType)
	if pluginMonitor == nil {
		hwlog.RunLog.Errorf("processType:%s don't have monitor", processType)
		return
	}
	m.nextFaultProcessor.Execute(pluginMonitor.GetMonitorData(), processType)
}

// SetNextFaultProcessor set the next fault processor
func (m *MonitorManager) SetNextFaultProcessor(processor common.FaultProcessor) {
	m.nextFaultProcessor = processor
}
