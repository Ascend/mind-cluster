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
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/grpc/profiling"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/application"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
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
	BusinessHandler       *application.BusinessStreamProcessor
	MsgHd                 *application.MsgHandler
	svcCtx                context.Context
	cancelFunc            context.CancelFunc
	profilingFromClusterD atomic.Bool
}

const (
	roleTaskd       = "taskd"
	maxRegRetryTime = 60
)

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
	go m.registerClusterD(0)
	go m.watchProfilingCmdChange()

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

func (m *BaseManager) registerClusterD(retryTime time.Duration) {
	if retryTime >= maxRegRetryTime {
		hwlog.RunLog.Error("init clusterd connect meet max retry time")
		return
	}
	time.Sleep(retryTime * time.Second)
	addr, err := utils.GetClusterdAddr()
	if err != nil {
		hwlog.RunLog.Errorf("get clusterd address err: %v", err)
		return
	}
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		hwlog.RunLog.Errorf("init clusterd connect err: %v", err)
		m.registerClusterD(retryTime + 1)
		return
	}

	go m.subscribeProfiling(conn, 0)
}

func (m *BaseManager) subscribeProfiling(conn *grpc.ClientConn, retryTime time.Duration) {
	m.profilingFromClusterD.Store(false)
	if retryTime >= maxRegRetryTime {
		hwlog.RunLog.Error("register Cluster profiling meet max retry time")
		return
	}
	time.Sleep(retryTime * time.Second)
	traceClient := profiling.NewTrainingDataTraceClient(conn)
	stream, err := traceClient.SubscribeDataTraceSwitch(m.svcCtx, &profiling.ProfilingClientInfo{
		JobId: m.JobId,
		Role:  roleTaskd,
	})
	if err != nil {
		hwlog.RunLog.Errorf("register Cluster profiling fail, err: %v", err)
		go m.subscribeProfiling(conn, retryTime+1)
		return
	}
	m.profilingFromClusterD.Store(true)
	for {
		select {
		case <-m.svcCtx.Done():
			hwlog.RunLog.Info("taskd exit, stop subscribe clusterd fault info")
			return
		case <-stream.Context().Done():
			hwlog.RunLog.Info("client stream exit, stop subscribe profiling info and re-register")
			go m.subscribeProfiling(conn, retryTime+1)
			return
		default:
			responseMsg, recvErr := stream.Recv()
			if recvErr != nil {
				hwlog.RunLog.Error(recvErr)
			} else {
				hwlog.RunLog.Infof("receive profiling info: %v", responseMsg)
				profilingMsg := responseMsg.GetProfilingSwitch()
				// notify framework receive profiling msg
				domainSwitch := utils.PfSwitchToPfDomainSwitch(convertProfilingMsg(profilingMsg))
				m.enqueueProfilingSwitch(domainSwitch, constant.ClusterDRank)
			}
		}
	}
}

func (m *BaseManager) enqueueProfilingSwitch(cmd constant.ProfilingDomainCmd, whichServer string) {
	message := storage.BaseMessage{
		Header: storage.MsgHeader{
			BizType: "default",
			Uuid:    uuid.New().String(),
			Src: &common.Position{
				Role:       constant.ClusterRole,
				ServerRank: whichServer,
			},
			Timestamp: time.Now(),
		},
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    utils.ProfilingCmdToBizCode(cmd),
		},
	}
	err := m.MsgHd.MsgQueue.Enqueue(message)
	if err != nil {
		hwlog.RunLog.Infof("%s enqueue profiling cmd %v err %v", whichServer, cmd, err)
		return
	}
	hwlog.RunLog.Infof("%s enqueue profiling cmd %v", whichServer, cmd)
}

func (m *BaseManager) watchProfilingCmdChange() {
	hwlog.RunLog.Info("begin watch ProfilingSwitchFilePath...")
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-m.svcCtx.Done():
			hwlog.RunLog.Info("end watch ProfilingSwitchFilePath...")
			return
		case <-ticker.C:
			if m.profilingFromClusterD.Load() {
				hwlog.RunLog.Infof("manager register clusterd, donot watch profiling file.")
				return
			}
			m.getProfilingFromFile()
		}
	}
}

func (m *BaseManager) getProfilingFromFile() {
	profilingSwitch, err := utils.GetProfilingSwitch(constant.ProfilingSwitchFilePath)
	if err != nil {
		hwlog.RunLog.Errorf("GetProfilingSwitch err: %v", err)
		return
	}
	domainSwitch := utils.PfSwitchToPfDomainSwitch(profilingSwitch)
	m.enqueueProfilingSwitch(domainSwitch, constant.TaskDRank)
}

func convertProfilingMsg(profilingSwitchData *profiling.ProfilingSwitch) constant.ProfilingSwitch {
	profilingSwitch := constant.ProfilingSwitch{
		CommunicationOperator: profilingSwitchData.CommunicationOperator,
		Step:                  profilingSwitchData.Step,
		SaveCheckpoint:        profilingSwitchData.SaveCheckpoint,
		FP:                    profilingSwitchData.FP,
		DataLoader:            profilingSwitchData.DataLoader,
	}
	return profilingSwitch
}
