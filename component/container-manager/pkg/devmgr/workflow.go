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

// Package devmgr hwDevMgr workflow
package devmgr

import (
	"context"
	"errors"
	"fmt"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
)

const (
	waitDevResetMaxTime = 60
	maxDevNum           = 100
)

var DevMgr = &HwDevMgr{}

// NewHwDevMgr new huawei dev manager
func NewHwDevMgr() error {
	var setters = []func() error{
		DevMgr.initDmgr,
		DevMgr.initInfoRelatedDev,
		DevMgr.initInfoRelatedNode,
	}
	for _, setter := range setters {
		if err := setter(); err != nil {
			return err
		}
	}
	return nil
}

// Name module name
func (hdm *HwDevMgr) Name() string {
	return "hwDev manager"
}

// Init module init
func (hdm *HwDevMgr) Init() error {
	hwlog.RunLog.Infof("init module <%s> success", hdm.Name())
	return nil
}

// Work module work
func (hdm *HwDevMgr) Work(ctx context.Context) {
}

// ShutDown module shutdown
func (hdm *HwDevMgr) ShutDown() {
	if err := hdm.GetDmgr().ShutDown(); err != nil {
		hwlog.RunLog.Warnf("shut down hdm dev manager failed, error: %v", err)
	}
}

func (hdm *HwDevMgr) initDmgr() error {
	devMgr, err := devmanager.AutoInit("", waitDevResetMaxTime)
	if err != nil {
		hwlog.RunLog.Errorf("init devmanager failed, err: %v", err)
		return errors.New("init devmanager failed")
	}
	hdm.SetDmgr(devMgr)
	return nil
}

func (hdm *HwDevMgr) initInfoRelatedDev() error {
	devNum, logicIds, err := hdm.dmgr.GetDeviceList()
	if err != nil {
		return err
	}
	if devNum > maxDevNum {
		return fmt.Errorf("invalid device num: %d", devNum)
	}
	// init npuInfos
	hdm.npuInfos, err = hdm.setNodeNPUInfo(logicIds, devNum)
	if err != nil {
		return err
	}
	if len(hdm.npuInfos) == 0 {
		return errors.New("npu info is nil")
	}
	return nil
}

func (hdm *HwDevMgr) initInfoRelatedNode() error {
	// init devType and workMode
	devType := hdm.GetDmgr().GetDevType()
	hdm.devType = devType
	switch devType {
	case api.Ascend910A, api.Ascend910B, api.Ascend910A3:
		hdm.workMode = hdm.dmgr.GetNpuWorkMode()
	default:
	}

	if len(hdm.npuInfos) == 0 {
		// unreachable branch
		return errors.New("npu info is nil")
	}
	// init boardId
	if err := hdm.setBoardId(hdm.npuInfos[0].LogicID); err != nil {
		return err
	}
	// init usage
	if err := hdm.setDeviceUsage(hdm.npuInfos[0].PhyID); err != nil {
		return err
	}
	// init ring info, based on board id and dev usage
	return hdm.setRingInfo()
}
