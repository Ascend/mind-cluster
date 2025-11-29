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

// Package app fault manager workflow
package app

import (
	"context"
	"fmt"

	"ascend-common/common-utils/hwlog"
	"container-manager/pkg/common"
	"container-manager/pkg/devmgr"
	"container-manager/pkg/fault/domain"
)

// FaultMgr fault manager
type FaultMgr struct {
	faultInfo *domain.FaultCache
}

// NewFaultMgr new fault manager
func NewFaultMgr() *FaultMgr {
	return &FaultMgr{
		faultInfo: domain.GetFaultCache(),
	}
}

// Name module name
func (fm *FaultMgr) Name() string {
	return "fault manager"
}

// Init module init
func (fm *FaultMgr) Init() error {
	if err := loadFaultCodeFromFile(); err != nil {
		hwlog.RunLog.Errorf("load fault code from file %s failed, error: %v", common.ParamOption.FaultCfgPath, err)
		return fmt.Errorf("load fault code from file %s failed", common.ParamOption.FaultCfgPath)
	}
	hwlog.RunLog.Infof("init module <%s> success", fm.Name())
	return nil
}

// Work module work
func (fm *FaultMgr) Work(ctx context.Context) {
	// to prevent cache information loss, actively query all fault information after startup
	fm.getAllFaultInfo()
	if err := devmgr.DevMgr.SubscribeFaultEvent(saveDevFaultInfo); err != nil {
		hwlog.RunLog.Errorf("subscribe fault event failed, error: %v", err)
		return
	}
	go fm.ProcessDCMIFault(ctx)
	go fm.checkMoreThanFiveMinFaults(ctx)
}

// ShutDown module shutdown
func (fm *FaultMgr) ShutDown() {
}
