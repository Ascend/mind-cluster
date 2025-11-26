/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package app workflow of module
package app

import (
	"context"
	"time"

	"ascend-common/common-utils/hwlog"
	"container-manager/pkg/common"
	"container-manager/pkg/reset/domain"
	"container-manager/pkg/workflow"
)

const (
	resetMgrModuleName             = "reset-manager"
	npuContinuouslyResetCountLimit = 3
	lastingToHandlePeriodInSeconds = 60
	resetMgrCycleInterval          = time.Second
	cooldownPeriod                 = time.Second * 30
	defaultWaitDeviceResetTime     = time.Second * 150
)

var (
	faultsToHandleLastingOneMinute = map[string]struct{}{common.RestartRequest: {}, common.RestartBusiness: {}}
	faultsToHandleAtOnce           = map[string]struct{}{common.FreeRestartNPU: {}, common.RestartNPU: {}}
)

// ResetMgr struct for reset manager
type ResetMgr struct {
	lastSuccessResetTime *time.Time
	resetCache           *domain.NpuInResetCache
	countCache           *domain.FailedResetCountCache
}

// NewResetMgr new reset manager
func NewResetMgr() workflow.Module {
	return &ResetMgr{
		resetCache: domain.GetNpuInResetCache(),
		countCache: domain.NewFailedResetCountCache(),
	}
}

// Name the name this module
func (r *ResetMgr) Name() string {
	return resetMgrModuleName
}

// Init do init job for module
func (r *ResetMgr) Init() error {
	hwlog.RunLog.Infof("init module <%s> success", r.Name())
	return nil
}

// Work main work flow cycle
func (r *ResetMgr) Work(ctx context.Context) {
	ticker := time.NewTicker(resetMgrCycleInterval)
	defer ticker.Stop()
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel closed")
			}
			hwlog.RunLog.Info("reset manager stop")
			return
		case <-ticker.C:
			r.processResetWork()
		}
	}
}

// ShutDown shut down module
func (r *ResetMgr) ShutDown() {
	// reset manager do nothing
}
