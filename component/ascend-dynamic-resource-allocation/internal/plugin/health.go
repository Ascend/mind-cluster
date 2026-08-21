/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

import (
	"context"

	"ascend-common/common-utils/healthz"
	"ascend-common/common-utils/hwlog"
)

var draHealthCheckers = []healthz.HealthChecker{
	draHealthChecker{},
}

type DraHealthManager struct {
	DraHealthzConfig *healthz.Config
}

// NewDraHealthChecker creates a DRA health manager wrapping the healthz config.
func NewDraHealthChecker(draHealthzConfig *healthz.Config) *DraHealthManager {
	return &DraHealthManager{
		DraHealthzConfig: draHealthzConfig,
	}
}

// RegisterHealthChecker registers the DRA health checkers with the healthz package.
func (adp *DraHealthManager) RegisterHealthChecker() {
	for _, hc := range draHealthCheckers {
		healthz.RegisterHealthChecker(hc)
	}
}

// StartHealthyCheck starts the healthz HTTP server and blocks until ctx is
// cancelled or the server fails to serve.
func (adp *DraHealthManager) StartHealthyCheck(ctx context.Context) error {
	if err := adp.DraHealthzConfig.Serve(ctx); err != nil {
		return err
	}
	hwlog.RunLog.Infof("healthz server started, addr=%s", adp.DraHealthzConfig.HealthzAddress)
	return nil
}

type draHealthChecker struct{}

// Check reports nil as long as the DRA plugin is reachable.
func (d draHealthChecker) Check(ctx context.Context) error {
	return nil
}
