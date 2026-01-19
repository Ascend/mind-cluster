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

// Package reporter for reporter fault device info
package reporter

import (
	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/processmanager"
)

// ReportManager manage reporters
type ReportManager struct {
	reporters          []common.PluginReporter
	client             *kubeclient.ClientK8s
	nextFaultProcessor common.FaultProcessor
	stopChan           chan struct{}
}

// NewReporterManager  create a reporter manager
func NewReporterManager(client *kubeclient.ClientK8s) *ReportManager {
	return &ReportManager{
		client:   client,
		stopChan: make(chan struct{}, 1),
	}
}

// Execute reporter fault device info
func (r *ReportManager) Execute(fcInfo *common.FaultAndConfigInfo, processType string) {
	reporters := processmanager.GetReporterPlugins(processType)
	for _, reporter := range reporters {
		go reporter.Report(fcInfo)
	}
}

// SetNextFaultProcessor set the next fault processor
func (r *ReportManager) SetNextFaultProcessor(processor common.FaultProcessor) {
	r.nextFaultProcessor = processor
}
