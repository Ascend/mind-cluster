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
	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/reporter/cmreporter"
)

// PluginReporter reporter plugin interface
type PluginReporter interface {
	Report(*common.FaultDevInfo)
	Init() error
}

// ReportManager manage reporters
type ReportManager struct {
	reporters          []PluginReporter
	client             *kubeclient.ClientK8s
	faultManager       common.FaultManager
	nextFaultProcessor common.FaultProcessor
	stopChan           chan struct{}
}

// NewReporterManager  create a reporter manager
func NewReporterManager(client *kubeclient.ClientK8s) *ReportManager {
	return &ReportManager{
		client:       client,
		faultManager: common.NewFaultManager(),
		stopChan:     make(chan struct{}, 1),
	}
}

// Init register reporter plugin and initialize them
func (r *ReportManager) Init() error {
	r.reporters = append(r.reporters, cmreporter.NewConfigMapReporter(r.client))
	for _, reporter := range r.reporters {
		if err := reporter.Init(); err != nil {
			hwlog.RunLog.Errorf("init reporter failed, err is %v", err)
			return err
		}
	}
	return nil
}

// Execute reporter fault device info
func (r *ReportManager) Execute(faultDevInfo *common.FaultDevInfo) {
	r.faultManager.SetFaultDevInfo(faultDevInfo)
	for _, reporter := range r.reporters {
		go reporter.Report(faultDevInfo)
	}
}

// SetNextFaultProcessor set the next fault processor
func (r *ReportManager) SetNextFaultProcessor(processor common.FaultProcessor) {
	r.nextFaultProcessor = processor
}
