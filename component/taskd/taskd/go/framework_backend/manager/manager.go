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

// Package manager for taskd manager backend
package manager

import (
	"context"
	"fmt"
	"time"

	"ascend-common/common-utils/hwlog"
	"taskd/common/utils"
	"taskd/framework_backend/manager/application"
	"taskd/framework_backend/manager/infrastructure/storage"
)

// ClusterInfo define the information from the cluster
type ClusterInfo struct {
	// IP indicate cluster server ip
	Ip string `json:"ip"`
	// Port indicate cluster server port
	Port string `json:"port"`
	// Name indicate cluster server service name
	Name string `json:"name"`
	// Role
	Role string `json:"role"`
}

// Config define the configuration of manager
type Config struct {
	// JobId indicate the id of the job where the manager is located
	JobId string `json:"job_id"`
	// NodeNums indicate the number of nodes where the manager is located
	NodeNums int `json:"node_nums"`
	// ProcPerNode indicate the number of business processes where the manager's job is located
	ProcPerNode int `json:"proc_per_node"`
	// PluginDir indicate the plugin dir
	PluginDir string `json:"plugin_dir"`
	// ClusterInfos indicate the information of cluster
	ClusterInfos []ClusterInfo `json:"cluster_infos"`
}

// NewTaskDManager return taskd manager instance
func NewTaskDManager(config Config) *BaseManager {
	return &BaseManager{
		Config: config,
	}
}

// BaseManager the class taskd manager backend
type BaseManager struct {
	Config
	BusinessHandler *application.BusinessStreamProcessor
	MsgHd           *application.MsgHandler
	svcCtx          context.Context
	cancelFunc      context.CancelFunc
}

// Init base manger
func (m *BaseManager) Init() error {
	if err := utils.InitHwLogger("manager.log", context.Background()); err != nil {
		fmt.Printf("manager init hwlog failed, err: %v \n", err)
		return err
	}
	m.svcCtx, m.cancelFunc = context.WithCancel(context.Background())
	m.MsgHd = application.NewMsgHandler()
	m.MsgHd.Start(m.svcCtx)

	m.BusinessHandler = application.NewBusinessStreamProcessor(m.MsgHd)
	if err := m.BusinessHandler.Init(); err != nil {
		hwlog.RunLog.Errorf("business handler init failed, err: %v", err)
		return err
	}

	hwlog.RunLog.Info("manager init success!")
	return nil
}

// Start taskd manager
func (m *BaseManager) Start() error {
	if err := m.Init(); err != nil {
		fmt.Printf("manager init failed, err: %v \n", err)
		return fmt.Errorf("manager init failed, err: %v", err)
	}
	if err := m.Process(); err != nil {
		hwlog.RunLog.Errorf("manager process failed, err: %v", err)
		return fmt.Errorf("manager process failed, err: %v", err)
	}
	return nil
}

// Process task main process
func (m *BaseManager) Process() error {
	for {
		time.Sleep(time.Second)
		snapshot, err := m.MsgHd.DataPool.GetSnapShot()
		if err != nil {
			return fmt.Errorf("get datapool snapshot failed, err: %v", err)
		}
		if err := m.Service(snapshot); err != nil {
			return fmt.Errorf("service execute failed, err: %v", err)
		}
		hwlog.RunLog.Infof("manager process loop!")
	}
}

// Service for taskd business serve
func (m *BaseManager) Service(snapshot *storage.SnapShot) error {
	m.BusinessHandler.AllocateToken(snapshot)
	if err := m.BusinessHandler.StreamRun(); err != nil {
		hwlog.RunLog.Errorf("business handler stream run failed, err: %v", err)
		return err
	}
	return nil
}
