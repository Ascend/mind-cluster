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

// Package workflow interface Module and related function
package workflow

import (
	"context"
	"syscall"

	"ascend-common/common-utils/hwlog"
	"container-manager/pkg/common"
)

// Module module interface
type Module interface {
	Name() string
	Init() error
	Work(ctx context.Context)
	ShutDown()
}

// ModuleMgr module manager
type ModuleMgr struct {
	modules []Module
}

// NewModuleMgr new module manager
func NewModuleMgr() *ModuleMgr {
	return &ModuleMgr{}
}

// Register register module
func (mm *ModuleMgr) Register(module Module) {
	mm.modules = append(mm.modules, module)
}

// Init module init
func (mm *ModuleMgr) Init() error {
	for _, module := range mm.modules {
		if err := module.Init(); err != nil {
			return err
		}
	}
	return nil
}

// Work module work
func (mm *ModuleMgr) Work(ctx context.Context) {
	for _, module := range mm.modules {
		go module.Work(ctx)
	}
}

// ShutDown module shutdown
func (mm *ModuleMgr) ShutDown(cancel context.CancelFunc) {
	osSignChan := common.NewSignWatcher(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	if osSignChan == nil {
		hwlog.RunLog.Error("the stop signal is not initialized")
		return
	}
	select {
	case s, signEnd := <-osSignChan:
		if signEnd == false {
			hwlog.RunLog.Info("catch stop signal channel is closed")
			return
		}
		hwlog.RunLog.Infof("received signal: %s, shutting down", s.String())
		cancel()
		for _, module := range mm.modules {
			module.ShutDown()
		}
	}
}
