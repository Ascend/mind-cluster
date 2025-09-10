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
	"io"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"ascend-common/common-utils/hwlog"
	clusterd_constant "clusterd/pkg/common/constant"
	"clusterd/pkg/interface/grpc/profiling"
	"clusterd/pkg/interface/grpc/recover"
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
	// FaultRecover indicate the fault recover strategy
	FaultRecover string `json:"fault_recover"`
	// TaskDEnable indicate the taskd process enable
	TaskDEnable string `json:"taskd_enable"`
	// ClusterInfos indicate the information of cluster
	ClusterInfos []ClusterInfo `json:"cluster_infos"`
}

var managerInstance *BaseManager

// NewTaskDManager return taskd manager instance
func NewTaskDManager(config Config) *BaseManager {
	if managerInstance != nil {
		return managerInstance
	}
	managerInstance = &BaseManager{
		Config: config,
	}
	return managerInstance
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
	maxWaitTime     = 60
	waitGapTime     = 1
)

// Init base manger
func (m *BaseManager) Init() error {
	if err := utils.InitHwLogger("manager.log", context.Background()); err != nil {
		fmt.Printf("manager init hwlog failed, err: %v \n", err)
		return err
	}
	hwlog.RunLog.Infof("manager config: %v", m.Config)
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
		hwlog.RunLog.Debug("manager process loop!")
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
	hwlog.RunLog.Infof("get clusterd addr %v", addr)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		hwlog.RunLog.Errorf("init clusterd connect err: %v", err)
		m.registerClusterD(retryTime + 1)
		return
	}

	if m.TaskDEnable == "on" {
		go m.subscribeProcessManageSignal(conn)
	}

	go m.subscribeProfiling(conn, 0)
	go m.subscribeSwitchNic(conn)
	go m.subscribeStressTest(conn)
}

func (m *BaseManager) subscribeStressTest(conn *grpc.ClientConn) {
	client := pb.NewRecoverClient(conn)
	clientInfo := &pb.ClientInfo{
		JobId: m.JobId,
		Role:  roleTaskd,
	}
	for {
		exit, wTime := m.listenStressTestSignal(client, clientInfo, waitGapTime)
		if exit {
			hwlog.RunLog.Info("taskd exit, stop subscribe clusterd fault info")
			break
		}
		time.Sleep(time.Duration(wTime) * time.Second)
		if wTime > maxWaitTime {
			wTime = 1
		}
	}
}

func (m *BaseManager) listenStressTestSignal(client pb.RecoverClient, clientInfo *pb.ClientInfo, wTime int) (bool, int) {
	stream, err := client.SubscribeNotifyExecStressTest(m.svcCtx, clientInfo)
	if err != nil {
		hwlog.RunLog.Errorf("register Clusterd notify stress test fail, err: %v", err)
		return false, wTime + waitGapTime
	}
	for {
		select {
		case <-m.svcCtx.Done():
			hwlog.RunLog.Info("taskd exit, stop subscribe clusterd fault info")
			return true, 0
		case <-stream.Context().Done():
			hwlog.RunLog.Debug("server stream abnormal interruption, register again")
			return false, wTime + waitGapTime
		default:
			responseMsg, recvErr := stream.Recv()
			if recvErr == io.EOF {
				hwlog.RunLog.Info("stream EOF, register again")
				return false, waitGapTime
			}
			if recvErr != nil {
				hwlog.RunLog.Debug(recvErr)
				continue
			}
			hwlog.RunLog.Infof("receive stress test info: %v", responseMsg)
			m.enqueueStressTest(responseMsg)
		}
	}
}

func (m *BaseManager) enqueueStressTest(stressParam *pb.StressTestRankParams) {
	rankOpStr := utils.ObjToString(stressParam.StressParam)
	msg := map[string]string{
		constant.StressTestRankOPStr: rankOpStr,
		constant.StressTestJobID:     m.JobId,
	}
	message := storage.BaseMessage{
		Header: storage.MsgHeader{
			BizType: "default",
			Uuid:    uuid.New().String(),
			Src: &common.Position{
				Role:       constant.ClusterRole,
				ServerRank: constant.ClusterDRank,
			},
			Timestamp: time.Now(),
		},
		Body: storage.MsgBody{
			MsgType:   constant.Action,
			Code:      constant.StressTestCode,
			Extension: msg,
		},
	}
	err := m.MsgHd.MsgQueue.Enqueue(message)
	if err != nil {
		hwlog.RunLog.Errorf("enqueue stress test msg err %v", err)
		return
	}
	hwlog.RunLog.Infof("enqueue stress test msg %v", msg)
}

func (m *BaseManager) subscribeSwitchNic(conn *grpc.ClientConn) {
	client := pb.NewRecoverClient(conn)
	clientInfo := &pb.ClientInfo{
		JobId: m.JobId,
		Role:  roleTaskd,
	}
	for {
		exit, wTime := m.listenSwitchNicSignal(client, clientInfo, waitGapTime)
		if exit {
			hwlog.RunLog.Info("taskd exit, stop subscribe clusterd fault info")
			break
		}
		time.Sleep(time.Duration(wTime) * time.Second)
		if wTime > maxWaitTime {
			wTime = 1
		}
	}
}

func (m *BaseManager) listenSwitchNicSignal(client pb.RecoverClient, clientInfo *pb.ClientInfo, wTime int) (bool, int) {
	stream, err := client.SubscribeNotifySwitch(m.svcCtx, clientInfo)
	if err != nil {
		hwlog.RunLog.Errorf("register Clusterd notify switch fail, err: %v", err)
		return false, wTime + waitGapTime
	}
	for {
		select {
		case <-m.svcCtx.Done():
			hwlog.RunLog.Info("taskd exit, stop subscribe clusterd fault info")
			return true, 0
		case <-stream.Context().Done():
			hwlog.RunLog.Debug("server stream abnormal interruption, register again")
			return false, wTime + waitGapTime
		default:
			responseMsg, recvErr := stream.Recv()
			if recvErr == io.EOF {
				hwlog.RunLog.Info("stream EOF, register again")
				return false, waitGapTime
			}
			if recvErr != nil {
				hwlog.RunLog.Debug(recvErr)
				continue
			}
			hwlog.RunLog.Infof("receive switch nic info: %v", responseMsg)
			globalOps := responseMsg.GetOp()
			globalRanks := responseMsg.GetRankID()
			m.enqueueSwitchNic(globalRanks, globalOps)
		}
	}
}

func (m *BaseManager) enqueueSwitchNic(ranks []string, ops []bool) {
	rankStr := utils.ObjToString(ranks)
	opStr := utils.ObjToString(ops)
	msg := map[string]string{
		constant.GlobalRankKey: rankStr,
		constant.GlobalOpKey:   opStr,
		constant.SwitchJobID:   m.JobId,
	}
	message := storage.BaseMessage{
		Header: storage.MsgHeader{
			BizType: "default",
			Uuid:    uuid.New().String(),
			Src: &common.Position{
				Role:       constant.ClusterRole,
				ServerRank: constant.ClusterDRank,
			},
			Timestamp: time.Now(),
		},
		Body: storage.MsgBody{
			MsgType:   constant.Action,
			Code:      constant.SwitchNicCode,
			Extension: msg,
		},
	}
	err := m.MsgHd.MsgQueue.Enqueue(message)
	if err != nil {
		hwlog.RunLog.Errorf("enqueue switch msg err %v", err)
		return
	}
	hwlog.RunLog.Infof("enqueue switch msg %v", msg)
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

func (m *BaseManager) subscribeProcessManageSignal(conn *grpc.ClientConn) {
	m.updateFaultRecover()
	recoverClient := pb.NewRecoverClient(conn)
	clientInfo := &pb.ClientInfo{
		JobId: m.JobId,
		Role:  roleTaskd,
	}
	for {
		exit := m.startSubscribe(recoverClient, clientInfo)
		if exit {
			hwlog.RunLog.Error("taskd exit, stop subscribe clusterd fault info")
			return
		}
	}
}

func (m *BaseManager) startSubscribe(recoverClient pb.RecoverClient, clientInfo *pb.ClientInfo) bool {
	status, err := recoverClient.Init(m.svcCtx, clientInfo)
	if err != nil || status.Code != common.OK {
		hwlog.RunLog.Errorf("request Init failed, error: %v, response: %v", err, status)
		return false
	}
	status, err = recoverClient.Register(m.svcCtx, clientInfo)
	if err != nil || status.Code != common.OK {
		hwlog.RunLog.Errorf("request Register failed, error: %v, response: %v", err, status)
		return false
	}
	stream, err := recoverClient.SubscribeProcessManageSignal(m.svcCtx, clientInfo)
	if err != nil {
		hwlog.RunLog.Errorf("request SubscribeProcessManageSignal failed, error: %v", err)
		return false
	}
	for {
		select {
		case <-m.svcCtx.Done():
			hwlog.RunLog.Info("taskd exit, stop subscribe clusterd fault info")
			return true
		case <-stream.Context().Done():
			hwlog.RunLog.Info("client stream exit, stop subscribe process info and re-register")
			return false
		default:
			responseMsg, recvErr := stream.Recv()
			if recvErr == io.EOF {
				hwlog.RunLog.Info("process client stream exit, stop subscribe clusterd fault info and re-register")
				return false
			}
			if recvErr != nil {
				hwlog.RunLog.Error(recvErr)
				return false
			}
			hwlog.RunLog.Infof("receive manage signal info: %v", responseMsg)
			m.enqueueProcessManageSignal(responseMsg, constant.ClusterDRank)
		}
	}
}

func (m *BaseManager) enqueueProcessManageSignal(processManageSignal *pb.ProcessManageSignal, serverRank string) {
	action := constant.KeepAlive
	code := constant.ProcessManageKeepAliveSignal
	var params map[string]string
	if processManageSignal.SignalType != clusterd_constant.KeepAliveSignalType {
		action = constant.Action
		code = constant.ProcessManageRecoverSignal
		params = map[string]string{
			constant.Uuid:           processManageSignal.Uuid,
			constant.SignalType:     processManageSignal.SignalType,
			constant.Actions:        utils.ObjToString(processManageSignal.Actions),
			constant.FaultRanks:     utils.ObjToString(utils.GetFaultRanksMapByList(processManageSignal.FaultRanks)),
			constant.ChangeStrategy: processManageSignal.ChangeStrategy,
			constant.Timeout:        strconv.FormatInt(processManageSignal.Timeout, constant.TenBase),
			constant.NodeRankIds:    utils.ObjToString(processManageSignal.NodeRankIds),
			constant.ExtraParams:    processManageSignal.ExtraParams,
		}
	}

	message := storage.BaseMessage{
		Header: storage.MsgHeader{
			BizType: "default",
			Uuid:    uuid.New().String(),
			Src: &common.Position{
				Role:       constant.ClusterRole,
				ServerRank: serverRank,
			},
			Timestamp: time.Now(),
		},
		Body: storage.MsgBody{
			MsgType:   action,
			Code:      int32(code),
			Extension: params,
		},
	}
	err := m.MsgHd.MsgQueue.Enqueue(message)
	if err != nil {
		hwlog.RunLog.Errorf("enqueue process manage signal %v error %v", processManageSignal, err)
		return
	}
	hwlog.RunLog.Infof("enqueue process manage signal successfully, signal: %v", processManageSignal)
}
func (m *BaseManager) updateFaultRecover() {
	message := storage.BaseMessage{
		Header: storage.MsgHeader{
			BizType: "default",
			Uuid:    uuid.New().String(),
			Src: &common.Position{
				Role:       common.MgrRole,
				ServerRank: "0",
			},
			Timestamp: time.Now(),
		},
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    constant.FaultRecoverCode,
			Message: m.FaultRecover,
		},
	}
	err := m.MsgHd.MsgQueue.Enqueue(message)
	if err != nil {
		hwlog.RunLog.Infof("updateFaultRecover err, message %v error %v", message, err)
		return
	}
	hwlog.RunLog.Infof("updateFaultRecover, message %v", message)
}

// ReportControllerInfoToClusterd report controller info to clusterd
func ReportControllerInfoToClusterd(message *constant.ControllerMessage) bool {
	addr, err := utils.GetClusterdAddr()
	if err != nil {
		hwlog.RunLog.Errorf("get clusterd address err: %v", err)
		return false
	}
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		hwlog.RunLog.Errorf("init clusterd connect err: %v", err)
		return false
	}
	client := pb.NewRecoverClient(conn)
	sendTimes := 0
	for sendTimes <= constant.MaxSendTimes {
		var success bool
		switch message.Action {
		case constant.RecoverStatus:
			success = reportRecoverStatus(message, client)
		case constant.ProcessFault:
			success = reportProcessFault(message, client)
		case constant.RecoverStrategy:
			success = reportRecoverStrategy(message, client)
		case constant.StopComplete:
			success = reportStopComplete(message, client)
		default:
			hwlog.RunLog.Errorf("unknown action %v", message.Action)
			return false
		}
		if success {
			return true
		}
		sendTimes++
	}
	hwlog.RunLog.Errorf("send message to clusterd failed, max send times: %v", constant.MaxSendTimes)
	return false
}

func reportRecoverStatus(message *constant.ControllerMessage, client pb.RecoverClient) bool {
	hwlog.RunLog.Infof("send recover status to clusterd, msg: %v, fault ranks: %v", message.Msg, message.FaultRanks)
	status := &pb.Status{Code: int32(message.Code), Info: message.Msg}
	faultRanks := make([]string, 0)
	for key, _ := range message.FaultRanks {
		faultRanks = append(faultRanks, strconv.Itoa(key))
	}
	_, err := client.ReportRecoverStatus(managerInstance.svcCtx, &pb.RecoverStatusRequest{
		IsolateRankIds: faultRanks,
		JobId:          managerInstance.JobId,
		Status:         status,
		Strategy:       message.Strategy,
	})
	if err != nil {
		hwlog.RunLog.Errorf("report recover status to clusterd failed, error: %v", err)
		return false
	}
	return true
}

func reportProcessFault(message *constant.ControllerMessage, client pb.RecoverClient) bool {
	hwlog.RunLog.Infof("send process fault to clusterd, fault ranks: %v", message.FaultRanks)
	faultRanks := make([]*pb.FaultRank, 0)
	for rankId, faultType := range message.FaultRanks {
		faultRanks = append(faultRanks, &pb.FaultRank{
			RankId:    strconv.Itoa(rankId),
			FaultType: strconv.Itoa(faultType),
		})
	}
	_, err := client.ReportProcessFault(managerInstance.svcCtx, &pb.ProcessFaultRequest{
		JobId:      managerInstance.JobId,
		FaultRanks: faultRanks,
	})
	if err != nil {
		hwlog.RunLog.Errorf("report process fault to clusterd failed, error: %v", err)
		return false
	}
	return true
}

func reportRecoverStrategy(message *constant.ControllerMessage, client pb.RecoverClient) bool {
	hwlog.RunLog.Infof("send recover strategy to clusterd, strategies: %v, fault rank: %v",
		message.StrategyList, message.FaultRanks)
	faultRanks := make([]*pb.FaultRank, 0)
	for rankId, faultType := range message.FaultRanks {
		faultRanks = append(faultRanks, &pb.FaultRank{
			RankId:    strconv.Itoa(rankId),
			FaultType: strconv.Itoa(faultType),
		})
	}
	_, err := client.ReportRecoverStrategy(managerInstance.svcCtx, &pb.RecoverStrategyRequest{
		JobId:      managerInstance.JobId,
		Strategies: message.StrategyList,
		FaultRanks: faultRanks,
	})
	if err != nil {
		hwlog.RunLog.Errorf("report recover strategy to clusterd failed, error: %v", err)
		return false
	}
	return true
}

func reportStopComplete(message *constant.ControllerMessage, client pb.RecoverClient) bool {
	hwlog.RunLog.Infof("send stop complete to clusterd, msg: %v, fault ranks: %v",
		message.Msg, message.FaultRanks)
	status := &pb.Status{Code: int32(message.Code), Info: message.Msg}
	faultRanks := make([]*pb.FaultRank, 0)
	for rankId, faultType := range message.FaultRanks {
		faultRanks = append(faultRanks, &pb.FaultRank{
			RankId:    strconv.Itoa(rankId),
			FaultType: strconv.Itoa(faultType),
		})
	}
	_, err := client.ReportStopComplete(managerInstance.svcCtx, &pb.StopCompleteRequest{
		JobId:      managerInstance.JobId,
		Status:     status,
		FaultRanks: faultRanks,
	})
	if err != nil {
		hwlog.RunLog.Errorf("report stop complete to clusterd failed, error: %v", err)
		return false
	}
	return true
}
