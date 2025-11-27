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

// Package app container controller workflow
package app

import (
	"context"
	"errors"
	"time"

	"ascend-common/common-utils/hwlog"
	"container-manager/pkg/common"
	"container-manager/pkg/container/domain"
	"container-manager/pkg/devmgr"
	domain2 "container-manager/pkg/fault/domain"
)

const workDuration = 2 * time.Second

// CtrCtl container controller
type CtrCtl struct {
	client     ContainerClient
	ctrInfoMap *domain.CtrCache // key: ctr id (used dev ctr); value: ctr info
	devInfoMap *domain.DevCache // key: dev phy id; value: ctr id
}

// NewCtrCtl new container controller
func NewCtrCtl() (*CtrCtl, error) {
	switch common.ParamOption.RuntimeType {
	case common.DockerType:
		dClient := NewDockerClient()
		if err := dClient.init(); err != nil {
			hwlog.RunLog.Errorf("connect to container runtime failed, error: %v", err)
			return nil, errors.New("connect to container runtime failed")
		}
		return &CtrCtl{
			client:     dClient,
			ctrInfoMap: domain.GetCtrInfo(),
			devInfoMap: domain.NewDevCache(devmgr.DevMgr.GetPhyIds()),
		}, nil
	case common.ContainerDType:
		cClient := NewContainerdClient()
		if err := cClient.init(); err != nil {
			hwlog.RunLog.Errorf("connect to container runtime failed, error: %v", err)
			return nil, errors.New("connect to container runtime failed")
		}
		return &CtrCtl{
			client:     cClient,
			ctrInfoMap: domain.GetCtrInfo(),
			devInfoMap: domain.NewDevCache(devmgr.DevMgr.GetPhyIds()),
		}, nil
	default:
		return nil, errors.New("unknown runtime mode")
	}
}

// Name module name
func (cm *CtrCtl) Name() string {
	return "container controller"
}

// Init module init
func (cm *CtrCtl) Init() error {
	hwlog.RunLog.Infof("init module <%s> success", cm.Name())
	return nil
}

// Work module work
func (cm *CtrCtl) Work(ctx context.Context) {
	ticker := time.NewTicker(workDuration)
	defer ticker.Stop()
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel closed")
			}
			hwlog.RunLog.Info("container controller stop")
			return
		case <-ticker.C:
			cm.initAndControl()
		case _, ok := <-domain2.SharedFaultCache.UpdateChan:
			if !ok {
				hwlog.RunLog.Info("catch update signal channel closed")
				return
			}
			faultCache := domain2.SharedFaultCache.GetAndClean()
			cm.devInfoMap.UpdateDevStatus(faultCache)
			cm.initAndControl()
		}
	}
}

// ShutDown module shutdown
func (cm *CtrCtl) ShutDown() {
	if err := cm.client.close(); err != nil {
		hwlog.RunLog.Errorf("close containerd client failed, error: %v", err)
	}
}
